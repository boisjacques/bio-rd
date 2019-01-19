package server

import (
	"fmt"
	"sync"

	"github.com/bio-routing/bio-rd/config"
	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/device"
	"github.com/bio-routing/bio-rd/protocols/isis/packet"
	"github.com/bio-routing/bio-rd/protocols/isis/types"
	"github.com/pkg/errors"

	log "github.com/sirupsen/logrus"
)

type devInterface interface {
	processIngressPacket() error
	processP2PHello(*packet.P2PHello, types.MACAddress) error
	processLSPDU(*packet.LSPDU, types.MACAddress) error
	processCSNP(*packet.CSNP, types.MACAddress) error
	processPSNP(*packet.PSNP, types.MACAddress) error
}

type dev struct {
	self           devInterface
	name           string
	srv            *Server
	sys            sys
	up             bool
	passive        bool
	p2p            bool
	level2         *level
	phy            *device.Device
	phyMu          sync.RWMutex
	done           chan struct{}
	wg             sync.WaitGroup
	helloMethod    func()
	receiverMethod func()
}

type level struct {
	HelloInterval   uint16
	HoldTime        uint16
	Metric          uint32
	neighborManager neighborManagerInterface
}

func newDev(srv *Server, ifcfg *config.ISISInterfaceConfig) *dev {
	d := &dev{
		name:    ifcfg.Name,
		srv:     srv,
		passive: ifcfg.Passive,
		p2p:     ifcfg.P2P,
		done:    make(chan struct{}),
	}
	d.self = d

	d.helloMethod = d.helloRoutine
	d.receiverMethod = d.receiverRoutine

	if ifcfg.ISISLevel2Config != nil {
		d.level2 = &level{}
		d.level2.HelloInterval = ifcfg.ISISLevel2Config.HelloInterval
		d.level2.HoldTime = ifcfg.ISISLevel2Config.HoldTime
		d.level2.Metric = ifcfg.ISISLevel2Config.Metric
	}

	return d
}

// DeviceUpdate receives interface status information and manages ISIS interface state
func (d *dev) DeviceUpdate(phy *device.Device) {
	d.phyMu.Lock()
	defer d.phyMu.Unlock()

	d.phy = phy
	if d.phy.OperState == device.IfOperUp {
		err := d.enable()
		if err != nil {
			log.Errorf("Unable to enable ISIS on %q: %v", d.name, err)
		}
		return
	}

	err := d.disable()
	if err != nil {
		log.Errorf("Unable to disable ISIS on %q: %v", d.name, err)
		return
	}
}

func (d *dev) enable() error {
	err := d.sys.openPacketSocket()
	if err != nil {
		return fmt.Errorf("Failed to open packet socket: %v", err)
	}

	err = d.sys.mcastJoin(packet.AllP2PISS)
	if err != nil {
		return fmt.Errorf("Failed to join multicast group: %v", err)
	}

	d.done = make(chan struct{})

	d.wg.Add(1)
	go d.receiverMethod()

	d.wg.Add(1)
	go d.helloMethod()

	d.srv.log.Infof("ISIS: Interface %q is now up", d.name)
	d.up = true
	return nil
}

func (d *dev) disable() error {
	close(d.done)

	err := d.sys.closePacketSocket()
	if err != nil {
		return errors.Wrap(err, "Unable to close socket")
	}

	d.wg.Wait()
	d.up = false
	return nil
}

func (d *dev) validateNeighborAddresses(addrs []uint32) []uint32 {
	res := make([]uint32, 0, len(addrs))
	d.phyMu.RLock()
	defer d.phyMu.RUnlock()

	for i := range addrs {
		for j := range d.phy.Addrs {
			if d.phy.Addrs[j].Contains(bnet.NewPfx(bnet.IPv4(addrs[i]), 32)) {
				res = append(res, addrs[i])
			}
		}
	}

	return res
}

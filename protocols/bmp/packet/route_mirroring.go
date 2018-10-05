package packet

import (
	"bytes"
	"fmt"
)

// RouteMirroringMsg represents a route mirroring message
type RouteMirroringMsg struct {
	CommonHeader  *CommonHeader
	PerPeerHeader *PerPeerHeader
	TLVs          []*InformationTLV
}

func decodeRouteMirroringMsg(buf *bytes.Buffer, ch *CommonHeader) (*RouteMirroringMsg, error) {
	rm := &RouteMirroringMsg{
		CommonHeader: ch,
	}

	pph, err := decodePerPeerHeader(buf)
	if err != nil {
		return nil, fmt.Errorf("Unable to decode per peer header: %v", err)
	}

	rm.PerPeerHeader = pph

	toRead := buf.Len()
	read := 0

	for read < toRead {
		tlv, err := decodeInformationTLV(buf)
		if err != nil {
			return nil, fmt.Errorf("Unable to decode TLV: %v", err)
		}

		rm.TLVs = append(rm.TLVs, tlv)
		read += int(tlv.InformationLength) + MinInformationTLVLen
	}

	return rm, nil
}

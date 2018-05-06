package routingtable

// ClientManager manages clients of routing tables (observer pattern)
type ClientManager struct {
	clients map[RouteTableClient]struct{} // Ensures a client registers at most once
	master  RouteTableClient
}

// NewClientManager creates and initializes a new client manager
func NewClientManager(master RouteTableClient) ClientManager {
	return ClientManager{
		clients: make(map[RouteTableClient]struct{}, 0),
		master:  master,
	}
}

// Register registers a client for updates
func (c *ClientManager) Register(client RouteTableClient) {
	if c.clients == nil {
		c.clients = make(map[RouteTableClient]struct{}, 0)
	}
	c.clients[client] = struct{}{}
	c.master.UpdateNewClient(client)
}

// Unregister unregisters a client
func (c *ClientManager) Unregister(client RouteTableClient) {
	if _, ok := c.clients[client]; !ok {
		return
	}
	delete(c.clients, client)
}

// Clients returns a list of registered clients
func (c *ClientManager) Clients() []RouteTableClient {
	ret := make([]RouteTableClient, 0)
	for rtc := range c.clients {
		ret = append(ret, rtc)
	}

	return ret
}
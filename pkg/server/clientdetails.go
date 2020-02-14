package server

import "net"

type ClientResponse struct {
	Addr string
	PacketSize int
}

// used to track client connections for return traffic
type ClientDetails struct {
	ClientConnMap map[string]net.Conn
}

func NewClientDetails() ClientDetails {
	c := ClientDetails{}
  c.ClientConnMap = make(map[string]net.Conn)

	return c
}

func (c *ClientDetails) AddClient( clientAddr string, conn net.Conn ) error {

	// lock it!
	c.ClientConnMap[clientAddr] = conn
	return nil
}


func (c *ClientDetails) DoesClientExists( clientAddr string) bool {

	_,ok := c.ClientConnMap[clientAddr]
	return ok
}








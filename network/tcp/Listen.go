package tcp

import (
	"goworker/network"
	"net"
)

type Listen struct {
	id         uint32
	url        *network.Url
	event      network.Event
	protocol   network.Protocol
	conn       net.Conn
	newConnect func(listen network.ListenTcp, conn net.Conn) network.Connect
}

func (c *Listen) SetUrl(address *network.Url) {
	c.url = address
}

func (c *Listen) Url() *network.Url {
	return c.url
}

func (c *Listen) SetEvent(event network.Event) {
	c.event = event
}

func (c *Listen) Event() network.Event {
	return c.event
}

func (c *Listen) SetProtocol(protocol network.Protocol) {
	c.protocol = protocol
}

func (c *Listen) Protocol() network.Protocol {
	return c.protocol
}

func (c *Listen) Close() {
	_ = c.conn.Close()
}

func (c *Listen) SetNewConnect(new func(listen network.ListenTcp, conn net.Conn) network.Connect) {
	c.newConnect = new
}

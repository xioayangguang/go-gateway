package tcp

import (
	"goworker/network"
	"net"
	"regexp"
)

type ConnectTcp struct {
	ip     uint32
	id     uint32
	uid    string
	url    *network.Url
	conn   net.Conn
	Listen network.ListenTcp
	header network.Header
}

func (c *ConnectTcp) GetIp() uint32 {
	if c.ip != 0 {
		return c.ip
	}
	ipStr := c.conn.RemoteAddr().String()
	r := `^(\d{1,3})\.(\d{1,3})\.(\d{1,3})\.(\d{1,3})`
	reg, err := regexp.Compile(r)
	if err != nil {
		return 0
	}
	ips := reg.FindStringSubmatch(ipStr)
	if ips == nil {
		return 0
	}

	c.ip = network.Ip2long(ips[0])
	return c.ip
}

func (c *ConnectTcp) GetPort() uint16 {
	return c.url.Port
}

func (c *ConnectTcp) GetCon() net.Conn {
	return c.conn
}

func (c *ConnectTcp) Close() {
	_ = c.conn.Close()
}

func (c *ConnectTcp) Id() uint32 {
	return c.id
}

func (c *ConnectTcp) SetUid(uid string) {
	c.uid = uid
}

func (c *ConnectTcp) Uid() string {
	return c.uid
}

func (c *ConnectTcp) Url() *network.Url {
	return c.url
}

func (c *ConnectTcp) SendByte(msg []byte) bool {
	err := c.Listen.Protocol().Write(c.conn, msg)
	if err != nil {
		return false
	}
	return true
}

func (c *ConnectTcp) SendString(msg string) bool {
	err := c.Listen.Protocol().Write(c.conn, []byte(msg))
	if err != nil {
		return false
	}
	return true
}

func (c *ConnectTcp) SetHeader(header network.Header) {
	c.header = header
}
func (c *ConnectTcp) Header() network.Header {
	return c.header
}

func (c *ConnectTcp) SetExtData(bytes []byte) {
}

func (c *ConnectTcp) ExtData() interface{} {
	return nil
}

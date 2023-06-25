package tcp

import (
	"encoding/binary"
	"goworker/network"
	"net"
	"regexp"
)

type ConnectTcp struct {
	ip uint32
	id uint32
	//clientId string
	userId   uint64
	groupsId map[uint64]struct{}
	url      *network.Url
	conn     net.Conn
	Listen   network.ListenTcp
	header   network.Header
	extData  []byte
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

//func (c *ConnectTcp) SetClientId(uid string) {
//	c.clientId = uid
//}
//
//func (c *ConnectTcp) ClientId() string {
//	return c.clientId
//}

func (c *ConnectTcp) Url() *network.Url {
	return c.url
}

func (c *ConnectTcp) SendByte(msg []byte) error {
	var buf32 = make([]byte, 4)
	binary.BigEndian.PutUint32(buf32, uint32(len(msg)))
	err := c.Listen.Protocol().Write(c.conn, append(buf32, msg...))
	if err != nil {
		return err
	}
	return nil
}

func (c *ConnectTcp) SendString(msg string) error {
	var buf32 = make([]byte, 4)
	binary.BigEndian.PutUint32(buf32, uint32(len(msg)))
	err := c.Listen.Protocol().Write(c.conn, append(buf32, []byte(msg)...))
	if err != nil {
		return err
	}
	return nil
}

func (c *ConnectTcp) SetHeader(header network.Header) {
	c.header = header
}
func (c *ConnectTcp) Header() network.Header {
	return c.header
}

func (c *ConnectTcp) SetExtData(bytes []byte) {
	c.extData = bytes
}

func (c *ConnectTcp) ExtData() []byte {
	return c.extData
}

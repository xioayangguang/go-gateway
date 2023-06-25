package tcp

import (
	"errors"
	"goworker/network"
	"net"
)

var id uint32

type Connection struct {
	ConnectTcp
}

func (c *Connection) Send(msg interface{}) error {
	switch msg.(type) {
	case []byte:
		return c.SendByte(msg.([]byte))
	case string:
		return c.SendString(msg.(string))
	}
	return errors.New("不支持的类型")
}

func (c *Connection) SetUserId(uid uint64) {
	c.userId = uid
}

func (c *Connection) UserId() uint64 {
	return c.userId
}

func (c *Connection) SetGroupId(groupsId uint64) {
	c.groupsId[groupsId] = struct{}{}
}

func (c *Connection) DeleteGroupId(groupsId uint64) {
	delete(c.groupsId, groupsId)
}

func (c *Connection) GroupsId() map[uint64]struct{} {
	return c.groupsId
}

func NewConnect(listen network.ListenTcp, conn net.Conn) network.Connect {
	id = id + 1
	url := network.NewUrl(listen.Url().Scheme + "://" + listen.Url().Host)
	return &Connection{
		ConnectTcp: ConnectTcp{
			id:     id,
			userId: 0,
			url:    url,
			conn:   conn,
			Listen: listen,
		},
	}
}

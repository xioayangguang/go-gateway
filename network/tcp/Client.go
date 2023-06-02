package tcp

import (
	"fmt"
	error2 "goworker/network/tcp_error"
	"log"
	"net"
)

type Client struct {
	Listen
}

func (c *Client) ListenAndServe() {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", c.url.Host)
	fmt.Println(c.url.Host)
	c.conn, err = net.DialTCP("tcp", nil, tcpAddr)
	fmt.Println(tcpAddr)
	if err != nil {
		go c.event.OnError(c, &error2.ListenError{c.url})
		log.Printf("tcp client 启动失败, err : %v\n", err.Error())
		return
	}
	c.id += 1
	go c.event.OnStart(c)

	if c.newConnect == nil {
		c.newConnect = NewConnect
	}
	Connect := c.newConnect(c, c.conn)
	c.protocol.Init()
	header, err := c.protocol.OnConnect(c.conn)
	if err != nil {
		_ = c.conn.Close()
		go c.event.OnError(c, &error2.ListenError{c.url})
		log.Printf("%v\n", err.Error())
		return
	}
	Connect.SetHeader(header)
	go c.event.OnConnect(Connect)

	for {
		msg, err := c.protocol.Read(c.conn)
		if err != nil {
			c.event.OnClose(Connect)
			return
		}
		c.event.OnMessage(Connect, msg)
	}
}

package tcp

import (
	"fmt"
	error2 "goworker/network/tcp_error"
	"log"
	"net"
)

type Server struct {
	Listen
}

/*
启动监听
*/
func (this *Server) ListenAndServe() {
	listener, err := net.Listen("tcp", this.Url().Host)
	if err != nil {
		go this.event.OnError(this, &error2.ListenError{this.Url()})
		log.Fatal("Error starting TCP server.", this.Url().Host, err)
		return
	}
	if this.newConnect == nil {
		this.newConnect = NewConnect
	}
	defer this.Close()
	this.protocol.Init()
	this.event.OnStart(this)
	for {
		con, _ := listener.Accept()
		this.id += 1
		go this.newConnection(con)
	}
}

/*
新的连接
*/
func (this *Server) newConnection(conn net.Conn) {
	Connect := this.newConnect(this, conn)
	header, err := this.protocol.OnConnect(conn)
	fmt.Println(header)
	if err != nil {
		_ = this.conn.Close()
		go this.event.OnError(this, &error2.ListenError{this.url})
		log.Printf("%v\n", err.Error())
		return
	}
	Connect.SetHeader(header)
	defer this.event.OnClose(Connect)
	go this.event.OnConnect(Connect)

	for {
		message, err := this.protocol.Read(conn)
		if err != nil {
			Connect.Close()
			break
		}
		go this.event.OnMessage(Connect, message)
	}
}

package tcp

import (
	"goworker/network"
	"goworker/network/protocol"
)

func NewClient(address string) network.ListenTcp {
	url := network.NewUrl(address)
	client := Client{}
	switch url.Scheme {
	case "ws":
		client.SetProtocol(&protocol.WsProtocol{})
	case "text":
		client.SetProtocol(&protocol.TextProtocol{})
	case "pack":
		client.SetProtocol(&protocol.PackageLenProtocol{})
	default:
		panic("ws or text")
	}
	client.SetUrl(url)
	return &client
}

func NewServer(address string) network.ListenTcp {
	url := network.NewUrl(address)
	server := Server{}
	switch url.Scheme {
	case "ws":
		server.SetProtocol(&protocol.WebsocketProtocol{})
	case "text":
		server.SetProtocol(&protocol.TextProtocol{})
	case "pack":
		server.SetProtocol(&protocol.PackageLenProtocol{})
	default:
		panic("ws、text、pack")
	}
	server.SetUrl(url)
	return &server
}

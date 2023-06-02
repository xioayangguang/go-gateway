package worker

import (
	"goworker/lib"
	"goworker/network"
	"log"
)

type GatewayHandle struct {
}

func (g *GatewayHandle) OnConnect(message lib.GatewayMessage) {
	clientId := network.Bin2hex(message.LocalIp, message.LocalPort, message.ConnectionId)
	lib.BussinessEvent.OnConnect(clientId)
}

func (*GatewayHandle) OnMessage(message lib.GatewayMessage) {
	clientId := network.Bin2hex(message.LocalIp, message.LocalPort, message.ConnectionId)
	lib.BussinessEvent.OnMessage(clientId, message.Body)
}

func (*GatewayHandle) OnClose(message lib.GatewayMessage) {
	clientId := network.Bin2hex(message.LocalIp, message.LocalPort, message.ConnectionId)
	lib.BussinessEvent.OnClose(clientId)
}

func (*GatewayHandle) onWebsocketConnect(message lib.GatewayMessage) {
	clientId := network.Bin2hex(message.LocalIp, message.LocalPort, message.ConnectionId)
	lib.BussinessEvent.OnWebSocketConnect(clientId, message.Body)
}

func (*GatewayHandle) OnTodo(message lib.GatewayMessage) {
	log.Println("todo ", message)
}

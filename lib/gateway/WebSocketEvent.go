package gateway

import (
	"goworker/lib"
	"goworker/network"
	"log"
	"unicode/utf8"
)

/*
客户端信息
*/
type GatewayHeader struct {
	// 内部通讯地址 , 对应本机地址
	LocalIp      uint32
	LocalPort    uint16
	ClientIp     uint32
	ClientPort   uint16
	GatewayPort  uint16
	ConnectionId uint32
	flag         uint8
}

func NewWebSocketEvent() network.Event {
	return &WebSocketEvent{}
}

type WebSocketEvent struct {
	// 内部通讯地址
	WorkerServerIp   string
	WorkerServerPort uint16
}

func (ws *WebSocketEvent) OnError(listen network.ListenTcp, err error) {

}

func (ws *WebSocketEvent) OnStart(listen network.ListenTcp) {
	ws.WorkerServerIp = lib.Config.Worker.Ip
	ws.WorkerServerPort = lib.Config.Worker.Port
	log.Println("ws server listening at: ", listen.Url().Origin)
}

//func (ws *WebSocketEvent) GetClientId(client network.Connect) string {
//	return network.Bin2hex(network.Ip2long(ws.WorkerServerIp), ws.WorkerServerPort, client.Id())
//}

func (ws *WebSocketEvent) OnConnect(client network.Connect) {
	//client.SetClientId(ws.GetClientId(client))
	// 添加连接池
	CG.AddedClient(client)
	ws.SendToWorker(client, lib.CMD_ON_CONNECT, []byte(""), 1, "")
}

// 收到websocket信息
func (ws *WebSocketEvent) OnMessage(c network.Connect, message []byte) {
	extData := c.ExtData()
	if extData == nil {
		ws.SendToWorker(c, lib.CMD_ON_MESSAGE, message, 1, "")
	} else {
		ws.SendToWorker(c, lib.CMD_ON_MESSAGE, message, 1, string(extData))
	}
}

func (ws *WebSocketEvent) OnClose(c network.Connect) {
	ws.SendToWorker(c, lib.CMD_ON_CLOSE, []byte(""), 1, "")
	CG.DeleteClient(c.Id())
}

// SendToWorker 发送信息的worker
func (ws *WebSocketEvent) SendToWorker(client network.Connect, cmd uint8, body []byte, flag uint8, ExtData string) {
	msg := lib.GatewayMessage{
		PackageLen:   28 + uint32(len(body)),
		Cmd:          cmd,
		LocalIp:      network.Ip2long(ws.WorkerServerIp),
		LocalPort:    ws.WorkerServerPort,
		ClientIp:     client.GetIp(),
		ClientPort:   client.GetPort(),
		ConnectionId: client.Id(),
		Flag:         flag,
		GatewayPort:  lib.Config.Gateway.Port,
		ExtLen:       uint32(utf8.RuneCountInString(ExtData)),
		ExtData:      ExtData,
		Body:         body,
	}
	worker, err := CG.GetWorker(client)
	if err != nil { // worker 找不到 获取连接
		log.Println("主动断开客户端连接 err:", err)
		client.Close()
		CG.DeleteClient(client.Id())
		return
	}
	log.Println("发信息给worker", string(msg.Body))
	err = worker.SendByte(lib.GatewayMessageToByte(msg))
	if err != nil {
		log.Println("发信息给worker", err)
	}
}

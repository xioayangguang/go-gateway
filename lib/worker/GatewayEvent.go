package worker

import (
	"encoding/json"
	"goworker/lib"
	"goworker/network"
	"log"
	"strconv"
)

type WorkerKey struct {
	Worker string `json:"worker_key"`
	Secret string `json:"secret_key"`
}

// 接受 lib 发来的数据
type GatewayEvent struct {
	registerEvent    *RegisterEvent
	gatewayAddresses *network.Url

	HandleFunc map[uint8]func(message lib.GatewayMessage)
}

func NewGatewayEvent(r *RegisterEvent, address string) network.Event {
	return &GatewayEvent{
		registerEvent:    r,
		gatewayAddresses: network.NewUrl(address),
	}
}

func (g *GatewayEvent) OnStart(listen network.ListenTcp) {
	if g.HandleFunc == nil {
		handle := GatewayHandle{}
		g.HandleFunc = map[uint8]func(message lib.GatewayMessage){}
		g.HandleFunc[lib.CMD_ON_CONNECT] = handle.OnConnect
		g.HandleFunc[lib.CMD_ON_MESSAGE] = handle.OnMessage
		g.HandleFunc[lib.CMD_ON_CLOSE] = handle.OnClose
		g.HandleFunc[lib.CMD_ON_WEBSOCKET_CONNECT] = handle.onWebsocketConnect
	}
}

// OnConnect gateway 连接
func (g *GatewayEvent) OnConnect(gateway network.Connect) {
	g.registerEvent.UpdateGatewayConnections(g.GetGatewayAddress(), gateway)
	msg := WorkerKey{
		Worker: "BussinessWorker:" + strconv.Itoa(int(gateway.Id())),
		Secret: lib.Config.SecretKey,
	}
	g.SendToGateway(gateway, lib.CMD_WORKER_CONNECT, msg)
}

// 接受 gateway 转发来的client数据
func (g *GatewayEvent) OnMessage(c network.Connect, message []byte) {
	msg := lib.ToGatewayMessage(message)
	if handle, ok := g.HandleFunc[msg.Cmd]; ok {
		handle(msg)
	} else {
		log.Println("GatewayEvent 不认识的命令", msg.Cmd, message)
	}
}

func (g *GatewayEvent) OnClose(connect network.Connect) {
	g.registerEvent.UpdateGatewayConnections(g.GetGatewayAddress(), nil)
}

func (g *GatewayEvent) OnError(listen network.ListenTcp, err error) {
	g.registerEvent.UpdateGatewayConnections(g.GetGatewayAddress(), nil)
}

func (g *GatewayEvent) GetGatewayAddress() string {
	return g.gatewayAddresses.Ip + ":" + strconv.Itoa(int(g.gatewayAddresses.Port))
}

// 发送数据到对应的客户端
func (g *GatewayEvent) SendToClient(uid string, cmd uint8, msg interface{}) {

}

// 发送数据到对应的 gateway 进程
func (g *GatewayEvent) SendToGateway(gatewayConnect network.Connect, cmd uint8, msg interface{}) {
	var body []byte
	switch msg.(type) {
	case []byte:
		body = msg.([]byte)
	case string:
		body = []byte(msg.(string))
	default:
		// 未知类型，直接转json
		body, _ = json.Marshal(msg)
	}

	gm := lib.GatewayMessage{
		PackageLen:   28 + uint32(len(body)),
		Cmd:          cmd,
		LocalIp:      0,
		LocalPort:    0,
		ClientIp:     gatewayConnect.GetIp(),
		ClientPort:   gatewayConnect.GetPort(),
		ConnectionId: gatewayConnect.Id(),
		Flag:         1,
		GatewayPort:  g.gatewayAddresses.Port,
		ExtLen:       0,
		ExtData:      "",
		Body:         body,
	}
	gatewayConnect.Send(lib.GatewayMessageToByte(gm))
}

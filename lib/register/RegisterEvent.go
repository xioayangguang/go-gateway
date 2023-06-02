package register

import (
	"encoding/json"
	"fmt"
	"goworker/lib"
	"goworker/network"
	"log"
)

type RegisterMessage struct {
	Event     string `json:"event"`
	Address   string `json:"address"`
	SecretKey string `json:"secret_key"`
}

type RegisterEvent struct {
	// ID 地址保存
	gatewayConnections map[uint32]string
	workerConnections  map[uint32]network.Connect
}

func NewRegisterEvent() network.Event {
	return &RegisterEvent{
		gatewayConnections: make(map[uint32]string),
		workerConnections:  make(map[uint32]network.Connect),
	}
}
func (self *RegisterEvent) OnStart(listen network.ListenTcp) {
	log.Println("register server listening at: ", listen.Url().Origin)
}

func (*RegisterEvent) OnConnect(connect network.Connect) {

}

func (self *RegisterEvent) OnMessage(connect network.Connect, message []byte) {
	var data RegisterMessage
	err := json.Unmarshal(message, &data)
	if err != nil {
		fmt.Println(err)
		connect.Close()
		return
	}
	if lib.Config.SecretKey != "" {
		if data.SecretKey != lib.Config.SecretKey {
			fmt.Println("秘要不对")
			connect.Close()
			return
		}
	}
	switch data.Event {
	case "gateway_connect":
		self.gatewayConnect(connect, data)
	case "worker_connect":
		self.workerConnect(connect, data)
	case "ping":
		return
	default:
		fmt.Println("不认识的事件定义")
		connect.Close()
	}
}

func (self *RegisterEvent) OnClose(connect network.Connect) {
	_, hasG := self.gatewayConnections[connect.Id()]
	if hasG == true {
		delete(self.gatewayConnections, connect.Id())
		self.broadcastAddresses(0)
	}

	_, hasW := self.workerConnections[connect.Id()]
	if hasW == true {
		delete(self.workerConnections, connect.Id())
	}
}

func (*RegisterEvent) OnError(listen network.ListenTcp, err error) {
	log.Println("注册中心启动失败", err)
}

// gateway 链接
func (self *RegisterEvent) gatewayConnect(c network.Connect, msg RegisterMessage) {
	if msg.Address == "" {
		println("address not found")
		c.Close()
		return
	}
	// 推入列表
	self.gatewayConnections[c.Id()] = msg.Address
	self.broadcastAddresses(0)
}

// worker 链接
func (self *RegisterEvent) workerConnect(c network.Connect, msg RegisterMessage) {
	// 推入列表
	self.workerConnections[c.Id()] = c
	self.broadcastAddresses(0)
}

/*
向 BusinessWorker 广播 gateway 内部通讯地址
0 全部发生
*/
func (self *RegisterEvent) broadcastAddresses(id uint32) {
	type ConList struct {
		Event     string   `json:"event"`
		Addresses []string `json:"addresses"`
	}
	data := ConList{Event: "broadcast_addresses"}

	for _, address := range self.gatewayConnections {
		data.Addresses = append(data.Addresses, address)
	}

	jsonByte, _ := json.Marshal(data)

	if id != 0 {
		worker := self.workerConnections[id]
		worker.SendByte(jsonByte)
		return
	}

	for _, worker := range self.workerConnections {
		worker.SendByte(jsonByte)
	}
}

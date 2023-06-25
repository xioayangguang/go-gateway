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
func (r *RegisterEvent) OnStart(listen network.ListenTcp) {
	log.Println("register server listening at: ", listen.Url().Origin)
}

func (r *RegisterEvent) OnConnect(connect network.Connect) {
	log.Println("connect: ", connect.GetIp(), connect.GetPort())
}

func (r *RegisterEvent) OnMessage(connect network.Connect, message []byte) {
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
	log.Printf("Register OnMessage:%s:%v Data:%s ", network.Long2Ip(connect.GetIp()), connect.GetPort(), string(message))
	switch data.Event {
	case "gateway_connect":
		r.gatewayConnect(connect, data)
	case "worker_connect":
		r.workerConnect(connect, data)
	case "ping":
		return
	default:
		fmt.Println("不认识的事件定义")
		connect.Close()
	}
}

func (r *RegisterEvent) OnClose(connect network.Connect) {
	_, hasG := r.gatewayConnections[connect.Id()]
	if hasG == true {
		delete(r.gatewayConnections, connect.Id())
		r.broadcastAddresses(0)
	}
	_, hasW := r.workerConnections[connect.Id()]
	if hasW == true {
		delete(r.workerConnections, connect.Id())
	}
}

func (r *RegisterEvent) OnError(listen network.ListenTcp, err error) {
	log.Println("注册中心启动失败", err)
}

// gateway 链接
func (r *RegisterEvent) gatewayConnect(c network.Connect, msg RegisterMessage) {
	if msg.Address == "" {
		println("address not found")
		c.Close()
		return
	}
	// 推入列表
	r.gatewayConnections[c.Id()] = msg.Address
	r.broadcastAddresses(0)
}

// worker 链接
func (r *RegisterEvent) workerConnect(c network.Connect, msg RegisterMessage) {
	// 推入列表
	r.workerConnections[c.Id()] = c
	r.broadcastAddresses(0)
}

/*
向 BusinessWorker 广播 gateway 内部通讯地址
0 全部发生
*/
func (r *RegisterEvent) broadcastAddresses(id uint32) {
	type ConList struct {
		Event     string   `json:"event"`
		Addresses []string `json:"addresses"`
	}
	data := ConList{Event: "broadcast_addresses"}
	for _, address := range r.gatewayConnections {
		data.Addresses = append(data.Addresses, address)
	}
	jsonByte, _ := json.Marshal(data)
	if id != 0 {
		worker := r.workerConnections[id]
		_ = worker.SendByte(jsonByte)
		return
	}
	for _, worker := range r.workerConnections {
		_ = worker.SendByte(jsonByte)
	}
}

package worker

import (
	"encoding/json"
	"fmt"
	"goworker/lib"
	"goworker/network"
	"goworker/network/tcp"
	"goworker/network/tcpclient"
	"log"
	"time"
)

type WorkerConnect struct {
	Event     string `json:"event"`
	SecretKey string `json:"secret_key"`
}
type BroadcastAddresses struct {
	Event     string   `json:"event"`
	Addresses []string `json:"addresses"`
}

type RegisterEvent struct {
	listen network.ListenTcp
}

/*
连接注册中心
*/
func NewRegisterEvent() network.Event {
	return &RegisterEvent{}
}

func (r *RegisterEvent) OnStart(listen network.ListenTcp) {
	log.Println("connect the register to: ", listen.Url().Origin)
	r.listen = listen
}

func (*RegisterEvent) OnConnect(c network.Connect) {
	conData := WorkerConnect{
		Event:     "worker_connect",
		SecretKey: lib.Config.SecretKey,
	}
	byteStr, _ := json.Marshal(conData)
	go c.SendByte(byteStr)
}

func (r *RegisterEvent) OnMessage(c network.Connect, message []byte) {
	strMsg := string(message)
	fmt.Println(strMsg)
	msgBA := BroadcastAddresses{}
	log.Printf("worker  Register OnMessage:%s:%v Data:%s ", network.Long2Ip(c.GetIp()), c.GetPort(), string(message))
	err := json.Unmarshal([]byte(strMsg), &msgBA)
	if err != nil {
		return
	}
	switch msgBA.Event {
	case "broadcast_addresses":
		for _, addr := range msgBA.Addresses {
			if _, ok := lib.GatewayList.GatewayCons[addr]; !ok {
				lib.GatewayList.GatewayCons[addr] = nil
			}
		}
		r.checkGatewayConnections()
	default:
		log.Println("不认识的事件", msgBA)
	}
}

func (*RegisterEvent) OnClose(c network.Connect) {

}

func (*RegisterEvent) OnError(listen network.ListenTcp, err error) {

}

func (r *RegisterEvent) checkGatewayConnections() {
	for addr, con := range lib.GatewayList.GatewayCons {
		if con == nil {
			address := "pack://" + addr
			worker := tcp.NewClient(address)
			worker.SetEvent(NewGatewayEvent(r, address))
			go worker.ListenAndServe()
		}
	}
}

// UpdateGatewayConnections 连接成功 or 失败
func (r *RegisterEvent) UpdateGatewayConnections(addr string, con network.Connect) {
	if con != nil {
		lib.GatewayList.GatewayCons[addr] = con
		client, err := tcpclient.NewClient(addr, 1, 10, 100*time.Millisecond, 10*time.Millisecond, time.Second)
		if err != nil {
			log.Fatal(err)
		}
		lib.GatewayList.ConnectionPool[addr] = client
	} else {
		delete(lib.GatewayList.GatewayCons, addr)
	}
}

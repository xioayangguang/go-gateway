package gateway

import (
	"encoding/json"
	"goworker/lib"
	"goworker/network"
	"log"
	"time"
)

type RegisterMessage struct {
	Event     string `json:"event"`
	Address   string `json:"address"`
	SecretKey string `json:"secret_key"`
}

type RegisterEvent struct {
	retry  int16
	listen network.ListenTcp
}

// 连接注册中心
func NewRegisterEvent() network.Event {
	return &RegisterEvent{}
}

// @error
func (r *RegisterEvent) OnError(listen network.ListenTcp, err error) {
	r.retry++
	log.Println("注册中心连接失败，2秒后重试", r.retry)
	ticker := time.NewTicker(time.Second * 2)
	select {
	case <-ticker.C:
		listen.ListenAndServe()
		break
	}
}

func (r *RegisterEvent) OnStart(listen network.ListenTcp) {
	log.Println("connect the register to: ", listen.Url().Origin)
	r.listen = listen
}

func (*RegisterEvent) OnConnect(c network.Connect) {
	// 发送内部通讯地址到注册中心
	conData := RegisterMessage{
		Event:     "gateway_connect",
		Address:   lib.Config.Worker.Host,
		SecretKey: lib.Config.SecretKey,
	}
	byteStr, _ := json.Marshal(conData)
	go c.SendByte(byteStr)
}

func (*RegisterEvent) OnMessage(c network.Connect, message []byte) {
	log.Println("gateway 收到注册中心的信息 ", string(message))
}

func (r *RegisterEvent) OnClose(c network.Connect) {
	// 注册中心 关闭,定时检查
	log.Print("注册中心断开连接，2秒后重连 ", r.retry)
	ticker := time.NewTicker(time.Second * 2)
	select {
	case <-ticker.C:
		r.listen.ListenAndServe()
	}
}

package main

import (
	"fmt"
	"goworker/apps"
	"goworker/lib"
	"goworker/lib/gateway"
	"goworker/lib/register"
	"goworker/lib/worker"
	"goworker/network/tcp"
	"goworker/utils"
	"os"
	"time"
)

func main() {
	//logFile, _ := os.OpenFile("register.log", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	//log.SetOutput(logFile)
	argsLen := len(os.Args)
	fmt.Println(argsLen)
	if argsLen < 2 {
		panic("长度错误")
	} else {
		switch os.Args[1] {
		case "registerStart":
			registerStart()
		case "registerStop":
			registerStop()
		case "workerStart":
			workerStart()
		case "gatewayStart":
			gatewayStart()
		case "gatewayStop":
			gatewayStop()
		default:
			fmt.Println("不存在的命令:" + os.Args[1])
		}
	}
}

func registerStart() {
	utils.SavePidToFile("register")
	utils.ListenStopSignal(func(sig os.Signal) {
		utils.DeleteSavePidToFile("register")
		os.Exit(0)
	})
	lib.Config.SetInput()
	server := tcp.NewServer(lib.Config.Register.Origin)
	server.SetEvent(register.NewRegisterEvent())
	server.ListenAndServe()
}

func registerStop() {
	err := utils.StopSignal("register")
	if err != nil {
		fmt.Println("停止失败")
	}
	fmt.Println("停止成功")
}

func workerStart() {
	lib.Config.SetInput()
	lib.BussinessEvent = &apps.MainEvent{}
	// 连接到注册中心
	r := tcp.NewClient(lib.Config.Register.Origin)
	r.SetEvent(worker.NewRegisterEvent())
	r.ListenAndServe()
	// 断线重连
	for {
		ticker := time.NewTicker(time.Second * 2)
		select {
		case <-ticker.C:
			r.ListenAndServe()
		}
	}
}

func gatewayStart() {
	lib.Config.SetInput()
	// 启动一个内部通讯tcp server
	w := tcp.NewServer(lib.Config.Worker.Origin)
	w.SetEvent(gateway.NewWorkerEvent())
	go w.ListenAndServe()
	// 连接到注册中心
	r := tcp.NewClient(lib.Config.Register.Origin)
	r.SetEvent(gateway.NewRegisterEvent())
	go r.ListenAndServe()
	// 启动对客户端的websocket连接
	// network.WebsocketMessageType = network.BinaryMessage
	server := tcp.NewServer(lib.Config.Gateway.Origin)
	server.SetEvent(gateway.NewWebSocketEvent())
	server.ListenAndServe()
}

func gatewayStop() {
	err := utils.StopSignal("gateway")
	if err != nil {
		fmt.Println("停止失败")
	}
	fmt.Println("停止成功")
}

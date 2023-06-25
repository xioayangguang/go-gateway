package network

import (
	"net"
)

// tcp 服务端 or 客户端监听接口
type ListenTcp interface {
	SetUrl(address *Url)           // 设置监听地址
	Url() *Url                     // 地址
	SetEvent(event Event)          // 设置信息事件
	Event() Event                  // 信息事件
	SetProtocol(protocol Protocol) // 设置解析协议
	Protocol() Protocol            // 协议对象
	Close()                        // 主动关闭
	ListenAndServe()               // 启动监听，阻塞
	SetNewConnect(func(listen ListenTcp, conn net.Conn) Connect)
}

// 连接实例
type Connect interface {
	SetHeader(header Header)
	Header() Header
	GetCon() net.Conn
	Close()
	Id() uint32
	SetUserId(uid uint64)
	UserId() uint64
	GroupsId() map[uint64]struct{}
	SetGroupId(id uint64)
	DeleteGroupId(id uint64)
	Send(msg interface{}) error
	SendByte(msg []byte) error
	SendString(msg string) error
	GetIp() uint32
	GetPort() uint16
	SetExtData(bytes []byte)
	ExtData() []byte
}

type Event interface {
	OnStart(listen ListenTcp)
	// 新链接
	OnConnect(connect Connect)
	// 新信息
	OnMessage(connect Connect, message []byte)
	// 链接关闭
	OnClose(connect Connect)
	// 发送错误
	OnError(listen ListenTcp, err error)
}

type Protocol interface {
	// 初始化
	Init()
	// 第一次连接，通常获取头信息
	OnConnect(conn net.Conn) (Header, error)
	// 读入处理
	Read(conn net.Conn) ([]byte, error)
	// 发送处理
	Write(conn net.Conn, msg []byte) error
}

type Header interface {
	Has(key string) bool
	Get(key string) string
	Set(data string)
}

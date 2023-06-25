package lib

import (
	"encoding/binary"
	"log"
	"reflect"
)

// 发给worker，gateway有一个新的连接
const CMD_ON_CONNECT = 1

// 发给worker的，客户端有消息
const CMD_ON_MESSAGE = 3

// 发给worker上的关闭链接事件
const CMD_ON_CLOSE = 4

// 发给gateway的向单个用户发送数据
const CMD_SEND_TO_ONE = 5

// 发给gateway的向所有用户发送数据
const CMD_SEND_TO_ALL = 6

// 发给gateway的踢出用户
// 1、如果有待发消息，将在发送完后立即销毁用户连接
// 2、如果无待发消息，将立即销毁用户连接
const CMD_KICK = 7

// 发给gateway的立即销毁用户连接
const CMD_DESTROY = 8

// 发给gateway，通知用户session更新
const CMD_UPDATE_SESSION = 9

// 获取所有在线client_id的session，client_id为 key
const CMD_GET_ALL_CLIENT_SESSIONS = 10

// 判断是否在线
const CMD_IS_ONLINE = 11

// client_id绑定到uid
const CMD_BIND_UID = 12

// 解绑
const CMD_UNBIND_UID = 13

// 向uid发送数据
const CMD_SEND_TO_UID = 14

// 根据uid获取绑定的clientid
const CMD_GET_CLIENT_ID_BY_UID = 15

// 加入组
const CMD_JOIN_GROUP = 20

// 离开组
const CMD_LEAVE_GROUP = 21

// 向组成员发消息
const CMD_SEND_TO_GROUP = 22

// 获取组成员所有在线client_id的session
const CMD_GET_CLIENT_SESSIONS_BY_GROUP = 23

// 获取组在线连接数
const CMD_GET_CLIENT_COUNT_BY_GROUP = 24

// 按照条件查找
const CMD_SELECT = 25

// 获取在线的群组ID
const CMD_GET_GROUP_ID_LIST = 26

// 取消分组
const CMD_UNGROUP = 27

// worker连接gateway事件
const CMD_WORKER_CONNECT = 200

// 心跳
const CMD_PING = 201

// GatewayClient连接gateway事件
const CMD_GATEWAY_CLIENT_CONNECT = 202

// 根据client_id获取session
const CMD_GET_SESSION_BY_CLIENT_ID = 203

// 发给gateway，覆盖session
const CMD_SET_SESSION = 204

// 当websocket握手时触发，只有websocket协议支持此命令字
const CMD_ON_WEBSOCKET_CONNECT = 205

// 包体是标量
const FLAG_BODY_IS_SCALAR = 0x01

// 通知gateway在send时不调用协议encode方法，在广播组播时提升性能
const FLAG_NOT_CALL_ENCODE = 0x02

// "Npack_len/Ccmd/Nlocal_ip/nlocal_port/Nclient_ip/nclient_port/Nconnection_id/Cflag/ngateway_port/Next_len"
type GatewayMessage struct {
	PackageLen   uint32
	Cmd          uint8
	LocalIp      uint32
	LocalPort    uint16
	ClientIp     uint32
	ClientPort   uint16
	ConnectionId uint32
	Flag         uint8
	GatewayPort  uint16
	ExtLen       uint32
	ExtData      string
	Body         []byte
}

func ToGatewayMessage(data []byte) GatewayMessage {
	Message := GatewayMessage{
		PackageLen:   uint32(len(data)),
		Cmd:          data[0],
		LocalIp:      uint32(binary.BigEndian.Uint32(data[1:5])),
		LocalPort:    uint16(binary.BigEndian.Uint16(data[5:7])),
		ClientIp:     uint32(binary.BigEndian.Uint32(data[7:11])),
		ClientPort:   uint16(binary.BigEndian.Uint16(data[11:13])),
		ConnectionId: uint32(binary.BigEndian.Uint32(data[13:17])),
		Flag:         data[17],
		GatewayPort:  uint16(binary.BigEndian.Uint16(data[18:20])),
		ExtLen:       uint32(binary.BigEndian.Uint32(data[20:24])),
	}
	Message.ExtData = string(data[24 : 24+Message.ExtLen])
	Message.Body = data[(24 + Message.ExtLen):(Message.PackageLen)]
	Message.PackageLen += 4
	return Message
}

func GatewayMessageToByte(msg GatewayMessage) []byte {
	var msgByte []byte
	value := reflect.ValueOf(msg)
	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		switch field.Kind() {
		case reflect.String:
			var bufStr []byte
			bufStr = []byte(field.String())
			msgByte = append(msgByte, bufStr...)
		case reflect.Uint8:
			msgByte = append(msgByte, uint8(field.Uint()))
		case reflect.Uint16:
			var buf16 = make([]byte, 2)
			binary.BigEndian.PutUint16(buf16, uint16(field.Uint()))
			msgByte = append(msgByte, buf16...)
		case reflect.Uint32:
			var buf32 = make([]byte, 4)
			binary.BigEndian.PutUint32(buf32, uint32(field.Uint()))
			msgByte = append(msgByte, buf32...)
		case reflect.Slice:
			msgByte = append(msgByte, field.Bytes()...)
		default:
			log.Println("GatewayProtocol 不知道的类型", field.Type(), msg)
		}
	}
	return msgByte[4:]
}

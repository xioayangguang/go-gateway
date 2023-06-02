package lib

import (
	"goworker/network"
	"strconv"
)

var GatewayList = &gatewayList{
	GatewayCons: map[string]network.Connect{},
}

var emptyGatewayMessage = GatewayMessage{
	PackageLen:   0,
	Cmd:          0,
	LocalIp:      0,
	LocalPort:    0,
	ClientIp:     0,
	ClientPort:   0,
	ConnectionId: 0,
	Flag:         1,
	GatewayPort:  0,
	ExtLen:       0,
	ExtData:      "",
	Body:         nil,
}

type gatewayList struct {
	// [ip:port]con
	GatewayCons map[string]network.Connect
}

func (self *gatewayList) SendToGateway(address string, msg GatewayMessage) {
	con, ok := self.GatewayCons[address]
	if ok {
		msgByte := GatewayMessageToByte(msg)
		con.SendByte(msgByte)
	}
}
func (self *gatewayList) SendToAllGateway(msg GatewayMessage) {
	msgByte := GatewayMessageToByte(msg)
	for _, con := range self.GatewayCons {
		con.SendByte(msgByte)
	}
}

type gateway struct{}

func SendToAll(body []byte) {
	msg := emptyGatewayMessage
	msg.PackageLen = 28 + uint32(len(body))
	msg.Cmd = CMD_SEND_TO_ALL
	msg.Body = body
	GatewayList.SendToAllGateway(msg)
}

// 向客户端client_id发送数据
// 发给用户网关
func SendToClient(clientId string, body []byte) {
	ip, port, id := network.DecodeBin2hex(clientId)
	msg := emptyGatewayMessage
	msg.PackageLen = 28 + uint32(len(body))
	msg.Cmd = CMD_SEND_TO_ONE
	msg.ConnectionId = id
	msg.Body = body

	address := network.Long2Ip(ip) + ":" + strconv.Itoa(int(port))
	GatewayList.SendToGateway(address, msg)
}

// 断开与client_id对应的客户端的连接
func CloseClient(clientId string) {

}

// 判断$client_id是否还在线
func IsOnline(clientId string) bool {
	return false
}

//todo  实现下面这些方法
//bindUid
//unbindUid
//isUidOnline
//getClientIdByUid
//getUidByClientId
//sendToUid
//joinGroup
//leaveGroup
//ungroup
//sendToGroup
//getClientIdCountByGroup
//getClientSessionsByGroup
//getAllClientIdCount
//getAllClientSessions
//setSession
//updateSession
//getSession
//getClientIdListByGroup
//getAllClientIdList
//getUidListByGroup
//getUidCountByGroup
//getAllUidList
//getAllUidCount
//getAllGroupIdList

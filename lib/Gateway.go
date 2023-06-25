package lib

import (
	"encoding/json"
	"errors"
	"goworker/network"
	"goworker/network/tcpclient"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

var GatewayList = &gatewayList{
	GatewayCons: map[string]network.Connect{},
	//worker进程主动推送进程
	ConnectionPool: map[string]tcpclient.Client{},
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
	Body:         []byte(""),
}

type gatewayList struct {
	GatewayCons    map[string]network.Connect
	ConnectionPool map[string]tcpclient.Client
}

func (g *gatewayList) SendToGateway(address string, msg GatewayMessage) error {
	con, ok := g.GatewayCons[address]
	msg.PackageLen = 28 + uint32(len(msg.Body))
	if ok {
		msgByte := GatewayMessageToByte(msg)
		return con.SendByte(msgByte)
	}
	return errors.New("无效的链接")
}

func (g *gatewayList) SendToGatewayByClientId(clientId string, msg GatewayMessage) error {
	ip, port, id := network.DecodeBin2hex(clientId)
	msg.ConnectionId = id
	address := network.Long2Ip(ip) + ":" + strconv.Itoa(int(port))
	return GatewayList.SendToGateway(address, msg)
}

func (g *gatewayList) SendAndRecvByClientId(clientId string, msg GatewayMessage) ([]byte, error) {
	ip, port, id := network.DecodeBin2hex(clientId)
	msg.ConnectionId = id
	address := network.Long2Ip(ip) + ":" + strconv.Itoa(int(port))
	return GatewayList.SendAndRecv(address, msg)
}

// SendAndRecv 发送数据并返回 ,这里不能复用上面的链接
func (g *gatewayList) SendAndRecv(address string, msg GatewayMessage) ([]byte, error) {
	msg.PackageLen = 28 + uint32(len(msg.Body))
	msgByte := GatewayMessageToByte(msg)
	if client, ok := g.ConnectionPool[address]; ok {
		return client.Send(msgByte)
	}
	return nil, errors.New("链接不存在")
}

// SendToAllGatewayAndRecv  批量发送请求
func (g *gatewayList) SendToAllGatewayAndRecv(msg GatewayMessage, call func(addr string, data []byte) error) (err error) {
	var wg sync.WaitGroup
	var reMap map[string][]byte
	for addr, _ := range GatewayList.GatewayCons {
		wg.Add(1)
		addr := addr
		go func() {
			var reData []byte
			reData, err = GatewayList.SendAndRecv(addr, msg)
			reMap[addr] = reData
			defer wg.Done()
		}()
	}
	wg.Wait()
	if err != nil {
		return err
	}
	for addr, data := range reMap {
		err := call(addr, data)
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *gatewayList) SendToAllGateway(msg GatewayMessage) error {
	msg.PackageLen = 28 + uint32(len(msg.Body))
	msgByte := GatewayMessageToByte(msg)
	for _, con := range g.GatewayCons {
		err := con.SendByte(msgByte)
		if err != nil {
			return err
		}
	}
	return nil
}

func ToClientId(addr, cid string) (string, error) {
	contInt, err := strconv.Atoi(cid)
	if err != nil {
		return "", err
	}
	addrSplit := strings.Split(addr, ":")
	if len(addrSplit) != 2 {
		return "", errors.New("数据解析错误")
	}
	ipInt, err := strconv.Atoi(addrSplit[0])
	if err != nil {
		return "", err
	}
	portInt, err := strconv.Atoi(addrSplit[1])
	if err != nil {
		return "", err
	}
	return network.Bin2hex(uint32(ipInt), uint16(portInt), uint32(contInt)), nil
}

// SendToAll 发送消息给所有网关（用户）
func SendToAll(body []byte) error {
	msg := emptyGatewayMessage
	msg.Cmd = CMD_SEND_TO_ALL
	msg.Body = body
	return GatewayList.SendToAllGateway(msg)
}

// SendToClient 向客户端client_id发送数据
func SendToClient(clientId string, body []byte) error {
	msg := emptyGatewayMessage
	msg.Cmd = CMD_SEND_TO_ONE
	msg.Body = body
	return GatewayList.SendToGatewayByClientId(clientId, msg)
}

// CloseClient 断开与client_id对应的客户端的连接
func CloseClient(clientId string) error {
	msg := emptyGatewayMessage
	msg.Cmd = CMD_KICK
	return GatewayList.SendToGatewayByClientId(clientId, msg)
}

// IsOnline 判断设备是否有设备在线
func IsOnline(clientId string) (bool, error) {
	msg := emptyGatewayMessage
	msg.Cmd = CMD_IS_ONLINE
	recv, err := GatewayList.SendAndRecvByClientId(clientId, msg)
	if err != nil {
		return false, err
	}
	var isOnlineMap map[string]bool
	err = json.Unmarshal(recv, &isOnlineMap)
	if err != nil {
		return false, err
	}
	return isOnlineMap["isOnline"], nil
}

func BindUid(clientId, uid string) error {
	msg := emptyGatewayMessage
	msg.Cmd = CMD_BIND_UID
	msg.ExtData = uid
	return GatewayList.SendToGatewayByClientId(clientId, msg)
}

func UnBindUid(clientId, uid string) error {
	msg := emptyGatewayMessage
	msg.Cmd = CMD_UNBIND_UID
	msg.ExtData = uid
	return GatewayList.SendToGatewayByClientId(clientId, msg)
}

// IsUidOnline 判断用户是否有设备在线
func IsUidOnline(uid string) (bool, error) {
	uids, err := GetClientIdByUid(uid)
	if err != nil {
		return false, err
	}
	if len(uids) > 0 {
		return true, nil
	}
	return false, nil
}

func GetClientIdByUid(uid string) ([]string, error) {
	msg := emptyGatewayMessage
	msg.Cmd = CMD_GET_CLIENT_ID_BY_UID
	msg.ExtData = uid
	var uids []string
	err := GatewayList.SendToAllGatewayAndRecv(msg, func(addr string, data []byte) error {
		clientId, err := ToClientId(addr, string(data))
		if err != nil {
			return err
		}
		uids = append(uids, clientId)
		return nil
	})
	return uids, err
}

func SendToUid(uid string, body []byte) error {
	msg := emptyGatewayMessage
	msg.Cmd = CMD_SEND_TO_UID
	msg.Body = body
	msg.ExtData = uid
	return GatewayList.SendToAllGateway(msg)
}

func JoinGroup(clientId, group string) error {
	msg := emptyGatewayMessage
	msg.Cmd = CMD_JOIN_GROUP
	msg.ExtData = group
	return GatewayList.SendToGatewayByClientId(clientId, msg)
}

func LeaveGroup(clientId, group string) error {
	msg := emptyGatewayMessage
	msg.Cmd = CMD_LEAVE_GROUP
	msg.ExtData = group
	return GatewayList.SendToGatewayByClientId(clientId, msg)
}

func Ungroup(group string) error {
	msg := emptyGatewayMessage
	msg.Cmd = CMD_UNGROUP
	msg.ExtData = group
	return GatewayList.SendToAllGateway(msg)
}

func SendToGroup(group string, body []byte) error {
	msg := emptyGatewayMessage
	msg.Cmd = CMD_SEND_TO_GROUP
	msg.Body = body
	msg.ExtData = group
	return GatewayList.SendToAllGateway(msg)
}

func GetClientCountByGroup(group string) (int64, error) {
	msg := emptyGatewayMessage
	msg.Cmd = CMD_GET_CLIENT_COUNT_BY_GROUP
	msg.ExtData = group
	var x int64
	err := GatewayList.SendToAllGatewayAndRecv(msg, func(addr string, data []byte) error {
		count, err := strconv.Atoi(string(data))
		if err != nil {
			return err
		}
		atomic.AddInt64(&x, int64(count))
		return nil
	})
	return x, err
}

func DestroyAddress(clientId string) error {
	msg := emptyGatewayMessage
	msg.Cmd = CMD_DESTROY
	return GatewayList.SendToGatewayByClientId(clientId, msg)
}

func SetSession(clientId, session string) error {
	msg := emptyGatewayMessage
	msg.Cmd = CMD_SET_SESSION
	msg.ExtData = session
	return GatewayList.SendToGatewayByClientId(clientId, msg)
}

func GetSession(clientId, session string) ([]byte, error) {
	msg := emptyGatewayMessage
	msg.Cmd = CMD_GET_SESSION_BY_CLIENT_ID
	msg.ExtData = session
	return GatewayList.SendAndRecvByClientId(clientId, msg)
}

func UpdateSession(clientId, session string) error {
	msg := emptyGatewayMessage
	msg.Cmd = CMD_UPDATE_SESSION
	msg.ExtData = session
	return GatewayList.SendToGatewayByClientId(clientId, msg)
}

func GetAllClientSessions() (session map[string]map[string]string, err error) {
	return GetClientSessionsByGroup("")
}

func GetClientSessionsByGroup(groupId string) (session map[string]map[string]string, err error) {
	msg := emptyGatewayMessage
	if groupId == "" {
		msg.Cmd = CMD_GET_ALL_CLIENT_SESSIONS
	} else {
		msg.Cmd = CMD_GET_CLIENT_SESSIONS_BY_GROUP
		msg.ExtData = groupId
	}
	var tempSession map[string]map[string]string
	err = GatewayList.SendToAllGatewayAndRecv(msg, func(addr string, data []byte) error {
		err := json.Unmarshal(data, &tempSession)
		if err != nil {
			return err
		}
		for conId, tempData := range tempSession {
			data := tempData
			clientId, err := ToClientId(addr, conId)
			if err != nil {
				return err
			}
			session[clientId] = data
		}
		return nil
	})
	return session, err
}

func GetGroupIdList() (groupsId []string, err error) {
	msg := emptyGatewayMessage
	msg.Cmd = CMD_GET_GROUP_ID_LIST
	err = GatewayList.SendToAllGatewayAndRecv(msg, func(addr string, data []byte) error {
		var temp []string
		err := json.Unmarshal(data, &temp)
		if err != nil {
			return err
		}
		groupsId = append(groupsId, temp...)
		return nil
	})
	return groupsId, err
}

//CMD_GET_ALL_CLIENT_SESSIONS

//CMD_GATEWAY_CLIENT_CONNECT
//CMD_ON_WEBSOCKET_CONNECT
//CMD_ON_WEBSOCKET_CONNECT
//FLAG_BODY_IS_SCALAR
//// 通知gateway在send时不调用协议encode方法，在广播组播时提升性能
//FLAG_NOT_CALL_ENCODE

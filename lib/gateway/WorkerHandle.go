package gateway

import (
	"encoding/json"
	"goworker/network"
	//Serialize "goworker/gophp/serialize"
	"goworker/lib"
	"goworker/lib/worker"
	"log"
	"strconv"
)

func NewWorkerHandle() *WorkerHandle {
	return &WorkerHandle{
		groupConnections: map[int]map[uint32]network.Connect{},
	}
}

// 处理worker发来的信息命令
type WorkerHandle struct {
	// groupId=>
	groupConnections map[int]map[uint32]network.Connect
}

// 单个用户信息
func (*WorkerHandle) OnSendToOne(c network.Connect, message lib.GatewayMessage) {
	client, err := Router.GetClient(message.ConnectionId)
	if err != nil {
		Router.DeleteClient(message.ConnectionId)
		return
	}
	client.Send(message.Body)
}

// 提出用户
func (*WorkerHandle) OnKick(c network.Connect, message lib.GatewayMessage) {
	// todo
}

// 立即摧毁用户连接
func (*WorkerHandle) OnDestroy(c network.Connect, message lib.GatewayMessage) {
	// todo
}

// 发给gateway的向所有用户发送数据
func (*WorkerHandle) OnSendToAll(c network.Connect, message lib.GatewayMessage) {
	for _, client := range Router.Clients {
		client.Send(message.Body)
	}
}

func (*WorkerHandle) OnWorkerConnect(c network.Connect, message lib.GatewayMessage) {
	WorkerKey := &worker.WorkerKey{}
	err := json.Unmarshal(message.Body, WorkerKey)
	if err != nil || lib.Config.SecretKey != WorkerKey.Secret {
		c.Close()
		return
	}
	Router.AddedWorker(c)
}

func (*WorkerHandle) OnGatewayClientConnect(c network.Connect, message lib.GatewayMessage) {
	WorkerKey := &worker.WorkerKey{}
	err := json.Unmarshal(message.Body, WorkerKey)
	if err != nil || lib.Config.SecretKey != WorkerKey.Secret {
		c.Close()
		return
	}
	Router.AddedWorker(c)
}

// 获取组成员session
func (self *WorkerHandle) OnClientSessionsByGroup(c network.Connect, message lib.GatewayMessage) {
	groupId, _ := strconv.Atoi(message.ExtData)
	if list, ok := self.groupConnections[groupId]; ok {
		clientArray := make(map[uint32][]byte)
		for clientId, client := range list {
			clientArray[clientId] = client.ExtData().([]byte)
		}
		//buffer, _ := Serialize.Marshal(clientArray)
		buffer := "a:0:{}"
		c.SendByte([]byte(buffer))
	} else {
		buffer := "a:0:{}"
		c.SendString(buffer)
	}
}

// 向组成员发消息
func (self *WorkerHandle) OnSendToGroup(c network.Connect, message lib.GatewayMessage) {
	groupList := struct {
		Group   []int `json:"group"`
		Exclude []int `json:"exclude"`
	}{}
	err := json.Unmarshal([]byte(message.ExtData), &groupList)
	if err != nil {
		log.Println("向组成员发消息，格式错误，检查数据类型")
		return
	}

	for _, groupId := range groupList.Group {
		if list, ok := self.groupConnections[groupId]; ok {
			for _, clientConnect := range list {
				clientConnect.Send(message.Body)
			}
		}
	}
}

// 根据client_id获取session
// const CMD_GET_SESSION_BY_CLIENT_ID = 203
func (*WorkerHandle) OnGetSessionByClientId(c network.Connect, message lib.GatewayMessage) {
	client, err := Router.GetClient(message.ConnectionId)
	if err != nil {
		return
	}
	session := client.ExtData()
	if session == nil {
		c.SendString("N;")
	} else {
		c.SendByte(session.([]byte))
	}
}

// 重新赋值 session
func (*WorkerHandle) OnSetSession(c network.Connect, message lib.GatewayMessage) {
	client, err := Router.GetClient(message.ConnectionId)
	if err != nil {
		return
	}
	client.SetExtData([]byte(message.ExtData))
}

// 加入组
func (self *WorkerHandle) OnJoinGroup(c network.Connect, message lib.GatewayMessage) {
	groupId, _ := strconv.Atoi(message.ExtData)
	if groupId <= 0 {
		log.Println("join(group) group empty, group=" + message.ExtData)
		return
	}
	clientConnection, err := Router.GetClient(message.ConnectionId)
	if err != nil {
		return
	}
	if _, ok := self.groupConnections[groupId]; !ok {
		self.groupConnections[groupId] = map[uint32]network.Connect{}
	}
	self.groupConnections[groupId][message.ConnectionId] = clientConnection
}

// 按照条件查找
func (*WorkerHandle) OnSelect(c network.Connect, message lib.GatewayMessage) {
	log.Println("按照条件查找")
}

// 按照条件查找
func (*WorkerHandle) OnTodo(c network.Connect, message lib.GatewayMessage) {
	// todo
	log.Println("todo cmd= ", message.Cmd)
}

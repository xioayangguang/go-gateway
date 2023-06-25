package gateway

import (
	"encoding/json"
	"fmt"
	"goworker/lib"
	"goworker/lib/worker"
	"goworker/network"
	"log"
	"strconv"
)

func NewWorkerHandle() *WorkerHandle {
	return &WorkerHandle{}
}

// WorkerHandle 处理worker发来的信息命令
type WorkerHandle struct {
}

// OnSendToOne 单个用户信息
func (w *WorkerHandle) OnSendToOne(c network.Connect, message lib.GatewayMessage) {
	client, err := CG.GetClient(message.ConnectionId)
	if err != nil {
		CG.DeleteClient(message.ConnectionId)
		return
	}
	err = client.Send(message.Body)
	if err != nil {
		log.Println("数据发送异常", err)
		CG.DeleteClient(client.Id())
		return
	}
}

// OnKick 提出用户
func (w *WorkerHandle) OnKick(c network.Connect, message lib.GatewayMessage) {
	client, _ := CG.GetClient(message.ConnectionId)
	err := client.Send(message.Body)
	if err != nil {
		log.Println("数据发送异常", err)
	}
	CG.DeleteClient(message.ConnectionId)
}

// OnDestroy 立即摧毁用户连接
func (w *WorkerHandle) OnDestroy(c network.Connect, message lib.GatewayMessage) {
	CG.DeleteClient(message.ConnectionId)
}

// 发给gateway的向所有用户发送数据
func (w *WorkerHandle) OnSendToAll(c network.Connect, message lib.GatewayMessage) {
	for _, client := range CG.Clients {
		err := client.Send(message.Body)
		if err != nil {
			log.Println("数据发送异常", err)
			CG.DeleteClient(client.Id())
			return
		}
	}
}

func (w *WorkerHandle) OnWorkerConnect(c network.Connect, message lib.GatewayMessage) {
	WorkerKey := &worker.WorkerKey{}
	if lib.Config.SecretKey != WorkerKey.Secret {
		log.Println("链接秘钥错误")
		c.Close()
		return
	}
	err := json.Unmarshal(message.Body, WorkerKey)
	if err != nil {
		log.Println("数据解析异常", err)
		c.Close()
		return
	}
	CG.AddedWorker(c)
}

func (w *WorkerHandle) OnGatewayClientConnect(c network.Connect, message lib.GatewayMessage) {
	WorkerKey := &worker.WorkerKey{}
	if lib.Config.SecretKey != WorkerKey.Secret {
		log.Println("链接秘钥错误")
		c.Close()
		return
	}
	err := json.Unmarshal(message.Body, WorkerKey)
	if err != nil {
		log.Println("数据解析异常", err)
		c.Close()
		return
	}
	CG.AddedWorker(c)
}

func (w *WorkerHandle) OnGetAllClientSessions(c network.Connect, message lib.GatewayMessage) {
	allClientSessions := make(map[uint32][]byte)
	for _, client := range CG.Clients {
		allClientSessions[client.Id()] = client.ExtData()
	}
	allClientSessionsByte, err := json.Marshal(allClientSessions)
	if err != nil {
		return
	}
	err = c.SendByte(allClientSessionsByte)
	if err != nil {
		log.Println("发送数据异常", err)
	}
}

// OnClientSessionsByGroup 获取组成员session
func (w *WorkerHandle) OnClientSessionsByGroup(c network.Connect, message lib.GatewayMessage) {
	intNum, _ := strconv.Atoi(message.ExtData)
	groupId := uint64(intNum)
	if list, ok := CG.GroupConnections[groupId]; ok {
		clientArray := make(map[uint32][]byte)
		for clientId, client := range list {
			clientArray[clientId] = client.ExtData()
		}
		buffer, err := json.Marshal(clientArray)
		if err != nil {
			log.Println("解析数据异常")
			return
		}
		err = c.SendByte(buffer)
		if err != nil {
			c.Close()
			log.Println("发送数据异常", err)
		}
	} else {
		err := c.SendString("[]")
		if err != nil {
			c.Close()
			log.Println("发送数据异常", err)
		}
	}
}

// OnSendToGroup 向组成员发消息
func (w *WorkerHandle) OnSendToGroup(c network.Connect, message lib.GatewayMessage) {
	groupList := struct {
		Group   []uint64 `json:"group"`
		Exclude []uint64 `json:"exclude"`
	}{}
	err := json.Unmarshal([]byte(message.ExtData), &groupList)
	if err != nil {
		log.Println("向组成员发消息，格式错误，检查数据类型")
		return
	}
	for _, groupId := range groupList.Group {
		if list, ok := CG.GroupConnections[groupId]; ok {
			for _, clientConnect := range list {
				err := clientConnect.Send(message.Body)
				if err != nil {
					log.Println("发送数据异常", err)
				}
			}
		}
	}
}

// OnGetSessionByClientId 根据client_id获取session
func (w *WorkerHandle) OnGetSessionByClientId(c network.Connect, message lib.GatewayMessage) {
	client, err := CG.GetClient(message.ConnectionId)
	if err != nil {
		CG.DeleteClient(message.ConnectionId)
		return
	}
	session := client.ExtData()
	if session == nil {
		err := c.SendString("Nil")
		if err != nil {
			log.Println("发送数据异常", err)
		}
	} else {
		err := c.SendByte(session)
		if err != nil {
			log.Println("发送数据异常", err)
		}
	}
}

// OnSetSession 重新赋值 session
func (w *WorkerHandle) OnSetSession(c network.Connect, message lib.GatewayMessage) {
	client, err := CG.GetClient(message.ConnectionId)
	if err != nil {
		CG.DeleteClient(message.ConnectionId)
		return
	}
	client.SetExtData([]byte(message.ExtData))
}

// OnJoinGroup 加入组
func (w *WorkerHandle) OnJoinGroup(c network.Connect, message lib.GatewayMessage) {
	intNum, _ := strconv.Atoi(message.ExtData)
	groupId := uint64(intNum)
	if groupId <= 0 {
		log.Println("群组id错误" + message.ExtData)
		return
	}
	client, err := CG.GetClient(message.ConnectionId)
	if err != nil {
		log.Println("用户不存在", err)
		return
	}
	if _, ok := CG.GroupConnections[groupId]; !ok {
		CG.GroupConnections[groupId] = map[uint32]network.Connect{}
	}
	client.SetGroupId(groupId)
	CG.GroupConnections[groupId][message.ConnectionId] = client
}

// OnSelect 按照条件查找
func (w *WorkerHandle) OnSelect(c network.Connect, message lib.GatewayMessage) {
	//todo
	log.Println("按照条件查找")
}

func (w *WorkerHandle) OnGetGroupIdList(c network.Connect, message lib.GatewayMessage) {
	keys := make([]uint64, 0, len(CG.GroupConnections))
	for key, _ := range CG.GroupConnections {
		keys = append(keys, key)
	}
	jsonByte, err := json.Marshal(keys)
	if err != nil {
		return
	}
	err = c.SendByte(jsonByte)
	if err != nil {
		log.Println("发送数据异常", err)
	}
}

func (w *WorkerHandle) OnUpdateSession(c network.Connect, message lib.GatewayMessage) {
	client, err := CG.GetClient(message.ConnectionId)
	if err != nil {
		CG.DeleteClient(message.ConnectionId)
		return
	}
	var oldData map[string]string
	err = json.Unmarshal(client.ExtData(), &oldData)
	if err != nil {
		log.Println("数据解析异常", err)
		return
	}
	var extData map[string]string
	err = json.Unmarshal([]byte(message.ExtData), &extData)
	if err != nil {
		log.Println("数据解析异常", err)
		return
	}
	for k, v := range extData {
		oldData[k] = v
	}
	bytes, err := json.Marshal(oldData)
	if err != nil {
		return
	}
	client.SetExtData(bytes)
}

func (w *WorkerHandle) OnIsOnline(c network.Connect, message lib.GatewayMessage) {
	online := `{"false"}`
	_, err := CG.GetClient(message.ConnectionId)
	if err == nil {
		online = `{"true"}`
	}
	err = c.SendString(online)
	if err != nil {
		log.Println("发送数据异常", err)
	}
}

func (w *WorkerHandle) OnUnBindUid(c network.Connect, message lib.GatewayMessage) {
	client, err := CG.GetClient(message.ConnectionId)
	if err != nil {
		log.Println(err)
		return
	}
	if client.UserId() != 0 {
		delete(CG.UserIdConnections[client.UserId()], client.Id())
		if len(CG.UserIdConnections[client.UserId()]) == 0 {
			delete(CG.UserIdConnections, client.UserId())
		}
		client.SetUserId(0)
	}
}

func (w *WorkerHandle) OnBindUid(c network.Connect, message lib.GatewayMessage) {
	intNum, err := strconv.Atoi(message.ExtData)
	if err != nil {
		log.Println("数据转换异常", err)
		return
	}
	uId := uint64(intNum)
	client, err := CG.GetClient(message.ConnectionId)
	if err != nil {
		log.Println(err)
		return
	}
	if client.UserId() != 0 {
		delete(CG.UserIdConnections[client.UserId()], client.Id())
		if len(CG.UserIdConnections[client.UserId()]) == 0 {
			delete(CG.UserIdConnections, client.UserId())
		}
	}
	client.SetUserId(uId)
	CG.UserIdConnections[client.UserId()][client.Id()] = client
}

func (w *WorkerHandle) OnSendToUid(c network.Connect, message lib.GatewayMessage) {
	intNum, err := strconv.Atoi(message.ExtData)
	if err != nil {
		log.Println("数据转换异常", err)
		return
	}
	uId := uint64(intNum)
	for _, client := range CG.UserIdConnections[uId] {
		err = client.SendByte(message.Body)
		if err != nil {
			log.Println("发送数据异常", err)
		}
	}
}

func (w *WorkerHandle) OnLeaveGroup(c network.Connect, message lib.GatewayMessage) {
	intNum, err := strconv.Atoi(message.ExtData)
	if err != nil {
		log.Println("数据转换异常", err)
		return
	}
	groupId := uint64(intNum)
	client, err := CG.GetClient(message.ConnectionId)
	if err != nil {
		CG.DeleteClient(message.ConnectionId)
		return
	}
	delete(CG.GroupConnections[groupId], client.Id())
	client.DeleteGroupId(groupId)
	if len(CG.GroupConnections[groupId]) == 0 {
		delete(CG.GroupConnections, groupId)
	}
}

func (w *WorkerHandle) OnUnGroup(c network.Connect, message lib.GatewayMessage) {
	intNum, err := strconv.Atoi(message.ExtData)
	if err != nil {
		log.Println("数据转换异常", err)
		return
	}
	groupId := uint64(intNum)
	for _, client := range CG.GroupConnections[groupId] {
		client.DeleteGroupId(groupId)
	}
	delete(CG.GroupConnections, groupId)
}

// OnGetClientCountByGroup 获取群的成员数量 或者在线用户数量
func (w *WorkerHandle) OnGetClientCountByGroup(c network.Connect, message lib.GatewayMessage) {
	var count int
	if message.ExtData == "" {
		count = len(CG.Clients)
	} else {
		intNum, err := strconv.Atoi(message.ExtData)
		if err != nil {
			log.Println("数据转换异常", err)
			return
		}
		groupId := uint64(intNum)
		count = len(CG.GroupConnections[groupId])
	}
	err := c.SendString(fmt.Sprintf("{\"count\":$v}", count))
	if err != nil {
		log.Println("发送数据异常", err)
	}
}

func (w *WorkerHandle) OnGetClientIdByUid(c network.Connect, message lib.GatewayMessage) {
	intNum, err := strconv.Atoi(message.ExtData)
	if err != nil {
		log.Println("数据转换异常", err)
		return
	}
	uId := uint64(intNum)
	keys := make([]uint32, 0, len(CG.UserIdConnections[uId]))
	for key, _ := range CG.UserIdConnections[uId] {
		keys = append(keys, key)
	}
	jsonByte, err := json.Marshal(keys)
	if err != nil {
		log.Println("数据转换异常", err)
		return
	}
	err = c.SendByte(jsonByte)
	if err != nil {
		log.Println("发送数据异常", err)
	}
}

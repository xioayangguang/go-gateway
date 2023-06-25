package gateway

import (
	"errors"
	"goworker/network"
	"sync"
)

var CG = &ConnectionGroup{
	lock:       &sync.RWMutex{},
	workers:    map[uint32]network.Connect{},
	Clients:    map[uint32]network.Connect{},
	clientList: map[uint32]network.Connect{},
	//GroupId/ConnectId
	GroupConnections: map[uint64]map[uint32]network.Connect{},
	//UserId/ConnectId
	UserIdConnections: map[uint64]map[uint32]network.Connect{},
}

type ConnectionGroup struct {
	lock *sync.RWMutex
	// worker ConnectionId 映射 worker
	workers map[uint32]network.Connect
	// client ConnectionId 映射 client
	Clients map[uint32]network.Connect
	// client ConnectionId 映射 worker ，记录已经通讯过的通道 client => worker
	clientList map[uint32]network.Connect
	// groupId/connection_id  群组集合
	GroupConnections map[uint64]map[uint32]network.Connect
	//userid/connection_id  用户设备集合
	UserIdConnections map[uint64]map[uint32]network.Connect
}

func (cg *ConnectionGroup) GetWorker(c network.Connect) (network.Connect, error) {
	cg.lock.Lock()
	defer cg.lock.Unlock()
	cid := c.Id()
	if worker, ok := cg.clientList[cid]; ok {
		// 已经通信过的通道
		return worker, nil
	} else {
		// 随机分配一个worker
		for _, worker := range cg.workers {
			cg.clientList[cid] = worker
			return worker, nil
		}
	}
	// 不存在
	return nil, errors.New("找不到worker")
}

func (cg *ConnectionGroup) AddedWorker(worker network.Connect) {
	cg.lock.Lock()
	defer cg.lock.Unlock()
	cg.workers[worker.Id()] = worker
}

// DeleteWorker worker 断开
func (cg *ConnectionGroup) DeleteWorker(worker network.Connect) {
	cg.lock.Lock()
	defer cg.lock.Unlock()
	cid := worker.Id()
	delete(cg.workers, cid)
	for clientId, worker := range cg.clientList {
		if cid == worker.Id() {
			delete(cg.clientList, clientId)
		}
	}
}

// AddedClient 新增客户端，并且建立路由映射  ,用户浏览器 或者手机句柄
func (cg *ConnectionGroup) AddedClient(c network.Connect) {
	cg.lock.Lock()
	defer cg.lock.Unlock()
	ConnectionId := c.Id()
	if _, ok := cg.Clients[ConnectionId]; ok {
		cg.DeleteClient(ConnectionId)
	}
	cg.Clients[ConnectionId] = c
}

func (cg *ConnectionGroup) GetClient(ConnectionId uint32) (network.Connect, error) {
	cg.lock.Lock()
	defer cg.lock.Unlock()
	c, ok := cg.Clients[ConnectionId]
	if ok {
		return c, nil
	}
	return nil, errors.New("客户端不存在")
}

// DeleteClient 删除 客户端
func (cg *ConnectionGroup) DeleteClient(ConnectionId uint32) {
	cg.lock.Lock()
	defer cg.lock.Unlock()
	delete(cg.clientList, ConnectionId)
	client, err := cg.GetClient(ConnectionId)
	if err == nil {
		userId := client.UserId()
		if userId != 0 {
			delete(CG.UserIdConnections[userId], ConnectionId)
		}
		for k, _ := range client.GroupsId() {
			delete(CG.GroupConnections[k], ConnectionId)
		}
	}
	client.Close()
	delete(cg.Clients, ConnectionId)
}

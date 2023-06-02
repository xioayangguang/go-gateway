package gateway

import (
	"errors"
	"goworker/network"
	"sync"
)

type WorkerRouter struct {
	lock *sync.RWMutex
	// worker ConnectionId 映射 worker
	workers map[uint32]network.Connect
	// client ConnectionId 映射 client
	Clients map[uint32]network.Connect
	// client ConnectionId 映射 worker ，记录已经通讯过的通道 client =》 worker
	clientList map[uint32]network.Connect
}

var Router = &WorkerRouter{
	lock:       &sync.RWMutex{},
	workers:    map[uint32]network.Connect{},
	Clients:    map[uint32]network.Connect{},
	clientList: map[uint32]network.Connect{},
}

func (w *WorkerRouter) GetWorker(c network.Connect) (network.Connect, error) {
	w.lock.Lock()
	defer w.lock.Unlock()
	cid := c.Id()
	if worker, ok := w.clientList[cid]; ok {
		// 已经通信过的通道
		return worker, nil
	} else {
		// 随机分配一个worker
		for _, worker := range w.workers {
			w.clientList[cid] = worker
			return worker, nil
		}
	}
	// 不存在
	return nil, errors.New("找不到worker")
}

func (w *WorkerRouter) AddedWorker(worker network.Connect) {
	w.lock.Lock()
	defer w.lock.Unlock()
	w.workers[worker.Id()] = worker
}

/*
worker 断开
*/
func (w *WorkerRouter) DeleteWorker(worker network.Connect) {
	w.lock.Lock()
	defer w.lock.Unlock()
	cid := worker.Id()
	delete(w.workers, cid)
	for clientId, worker := range w.clientList {
		if cid == worker.Id() {
			delete(w.clientList, clientId)
		}
	}
}

/*
新增客户端，并且建立路由映射
*/
func (w *WorkerRouter) AddedClient(c network.Connect) {
	w.lock.Lock()
	defer w.lock.Unlock()
	ConnectionId := c.Id()
	if _, ok := w.Clients[ConnectionId]; ok {
		w.DeleteClient(ConnectionId)
	}

	w.Clients[ConnectionId] = c
}

func (w *WorkerRouter) GetClient(ConnectionId uint32) (network.Connect, error) {
	w.lock.Lock()
	defer w.lock.Unlock()
	c, ok := w.Clients[ConnectionId]
	if ok {
		return c, nil
	}
	return nil, errors.New("客户端不存在")
}

/*
删除 客户端
*/
func (w *WorkerRouter) DeleteClient(ConnectionId uint32) {
	w.lock.Lock()
	defer w.lock.Unlock()
	delete(w.clientList, ConnectionId)
	delete(w.Clients, ConnectionId)
}

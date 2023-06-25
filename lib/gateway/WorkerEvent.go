package gateway

import (
	"goworker/lib"
	"goworker/network"
	"log"
)

func NewWorkerEvent() network.Event {
	return &WorkerEvent{}
}

/*
接受worker发上来的数据
*/
type WorkerEvent struct {
	HandleFunc map[uint8]func(c network.Connect, message lib.GatewayMessage)
}

func (*WorkerEvent) OnError(listen network.ListenTcp, err error) {

}

func (w *WorkerEvent) OnStart(listen network.ListenTcp) {
	log.Println("worker server listening at: ", listen.Url().Origin)
	if w.HandleFunc == nil {
		w.HandleFunc = map[uint8]func(c network.Connect, message lib.GatewayMessage){}
		WorkerHandle := NewWorkerHandle()
		w.HandleFunc[lib.CMD_WORKER_CONNECT] = WorkerHandle.OnWorkerConnect
		w.HandleFunc[lib.CMD_GATEWAY_CLIENT_CONNECT] = WorkerHandle.OnGatewayClientConnect
		w.HandleFunc[lib.CMD_SEND_TO_ONE] = WorkerHandle.OnSendToOne
		w.HandleFunc[lib.CMD_KICK] = WorkerHandle.OnKick
		w.HandleFunc[lib.CMD_DESTROY] = WorkerHandle.OnDestroy
		w.HandleFunc[lib.CMD_SEND_TO_ALL] = WorkerHandle.OnSendToAll
		w.HandleFunc[lib.CMD_SELECT] = WorkerHandle.OnSelect
		w.HandleFunc[lib.CMD_GET_GROUP_ID_LIST] = WorkerHandle.OnGetGroupIdList
		w.HandleFunc[lib.CMD_SET_SESSION] = WorkerHandle.OnSetSession
		w.HandleFunc[lib.CMD_UPDATE_SESSION] = WorkerHandle.OnUpdateSession
		w.HandleFunc[lib.CMD_GET_SESSION_BY_CLIENT_ID] = WorkerHandle.OnGetSessionByClientId
		w.HandleFunc[lib.CMD_IS_ONLINE] = WorkerHandle.OnIsOnline
		w.HandleFunc[lib.CMD_BIND_UID] = WorkerHandle.OnBindUid
		w.HandleFunc[lib.CMD_UNBIND_UID] = WorkerHandle.OnUnBindUid
		w.HandleFunc[lib.CMD_SEND_TO_UID] = WorkerHandle.OnSendToUid
		w.HandleFunc[lib.CMD_JOIN_GROUP] = WorkerHandle.OnJoinGroup
		w.HandleFunc[lib.CMD_LEAVE_GROUP] = WorkerHandle.OnLeaveGroup
		w.HandleFunc[lib.CMD_UNGROUP] = WorkerHandle.OnUnGroup
		w.HandleFunc[lib.CMD_SEND_TO_GROUP] = WorkerHandle.OnSendToGroup
		w.HandleFunc[lib.CMD_GET_CLIENT_SESSIONS_BY_GROUP] = WorkerHandle.OnClientSessionsByGroup
		w.HandleFunc[lib.CMD_GET_ALL_CLIENT_SESSIONS] = WorkerHandle.OnGetAllClientSessions
		w.HandleFunc[lib.CMD_GET_CLIENT_COUNT_BY_GROUP] = WorkerHandle.OnGetClientCountByGroup
		w.HandleFunc[lib.CMD_GET_CLIENT_ID_BY_UID] = WorkerHandle.OnGetClientIdByUid
	}
}

func (*WorkerEvent) OnConnect(c network.Connect) {

}

func (w *WorkerEvent) OnMessage(c network.Connect, message []byte) {
	msg := lib.ToGatewayMessage(message)
	log.Println("收到worker的数据：", string(msg.Body))
	if handle, ok := w.HandleFunc[msg.Cmd]; ok {
		handle(c, msg)
	} else {
		log.Println("不认识的命令", msg.Cmd, c.GetCon().RemoteAddr().String())
	}
}

func (*WorkerEvent) OnClose(c network.Connect) {
	c.Close()
	CG.DeleteWorker(c)
}

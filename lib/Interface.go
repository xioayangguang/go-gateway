package lib

type LogicEvent interface {
	OnStart()
	// 新链接
	OnConnect(clientId string)
	// 当客户端连接上gateway完成websocket握手时触发的回调函数。
	OnWebSocketConnect(clientId string, header []byte)
	// 新信息
	OnMessage(clientId string, body []byte)
	// 链接关闭
	OnClose(clientId string)
}

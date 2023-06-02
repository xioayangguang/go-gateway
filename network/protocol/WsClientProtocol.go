package protocol

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"goworker/network"
	"goworker/network/protocol/lib"
	"math/rand"
	"net"
)

// 客户端端websocket协议
type WsProtocol struct {
	// 本地缓冲区
	cacheByte []byte
	// 缓冲长度
	cacheCount int
}

func (w *WsProtocol) Init() {

}

func (w *WsProtocol) OnConnect(conn net.Conn) (network.Header, error) {
	w.cacheByte = make([]byte, 0)

	// 发送请求头
	strHeader := "GET / HTTP/1.1\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nHost: "
	strHeader += conn.RemoteAddr().String() + "\r\n"
	strHeader += "Origin: http://" + conn.RemoteAddr().String() + "\r\n"
	strHeader += "Sec-WebSocket-Key:" + base64.StdEncoding.EncodeToString(randSeq(16)) + "\r\n"
	strHeader += "Sec-WebSocket-Version: 13\r\n\r\n"
	conn.Write([]byte(strHeader))

	// 获取协议头
	byteHeader, err := w.getHeader(conn)
	header := lib.MessageHeader{}
	header.Set(string(byteHeader))
	return &header, err
}

func (w *WsProtocol) Read(conn net.Conn) ([]byte, error) {
	// 第一个字节：FIN + RSV1-3 + OPCODE
	opcodeByte, err := w.readConnOrCache(conn, 1)
	if err != nil {
		return nil, err
	}
	FIN := opcodeByte[0] >> 7
	// RSV1 := opcodeByte[0] >> 6 & 1 // 自定义协议
	// RSV2 := opcodeByte[0] >> 5 & 1 // 自定义协议
	// RSV3 := opcodeByte[0] >> 4 & 1 // 自定义协议
	// OPCODE := opcodeByte[0] & 15
	//
	// log.Println(RSV1, RSV2, RSV3, OPCODE)

	// 第二个字节，Mask + Payload length
	payloadLenByte, err := w.readConnOrCache(conn, 1)
	payloadLen := int(payloadLenByte[0] & 0x7F) // 有效负载
	// mask := payloadLenByte[0] >> 7

	switch payloadLen {
	case 126: // 两个字节表示的是一个16进制无符号数，这个数用来表示传输数据的长度
		temLen, _ := w.readConnOrCache(conn, 2)
		payloadLen = int(binary.BigEndian.Uint16(temLen))
	case 127: // 8个字节表示的一个64位无符合数，这个数用来表示传输数据的长度
		temLen, _ := w.readConnOrCache(conn, 8)
		payloadLen = int(binary.BigEndian.Uint64(temLen))
	}

	msg := make([]byte, payloadLen)
	msg, err = w.readConnOrCache(conn, payloadLen)
	if err != nil {
		return nil, err
	}

	if FIN == 1 {
		// 最后的消息片断
		return msg, err
	}

	nextMsg, err := w.Read(conn)
	msg = append(msg, nextMsg...)
	return msg, err
}

func (w *WsProtocol) Write(conn net.Conn, msg []byte) error {
	length := len(msg)
	sendByte := make([]byte, 0)
	sendByte = append(sendByte, []byte{WebsocketMessageType}...)

	var payLenByte byte
	switch {
	case length <= 125:
		payLenByte = byte(0x80) | byte(length)
		sendByte = append(sendByte, []byte{payLenByte}...)
	case length <= 65536:
		payLenByte = byte(0x80) | byte(0x7e)
		sendByte = append(sendByte, []byte{payLenByte}...)
		// 随后的两个字节表示的是一个16进制无符号数，用来表示传输数据的长度
		payLenByte2 := make([]byte, 2)
		binary.BigEndian.PutUint16(payLenByte2, uint16(length))
		sendByte = append(sendByte, payLenByte2...)
	default:
		payLenByte = byte(0x80) | byte(0x7f)
		sendByte = append(sendByte, []byte{payLenByte}...)
		// 随后的是8个字节表示的一个64位无符合数，这个数用来表示传输数据的长度
		payLenByte8 := make([]byte, 8)
		binary.BigEndian.PutUint64(payLenByte8, uint64(length))
		sendByte = append(sendByte, payLenByte8...)
	}
	n := rand.Uint32()
	MaskingKey := [4]byte{byte(n), byte(n >> 8), byte(n >> 16), byte(n >> 24)}
	sendByte = append(sendByte, MaskingKey[:]...)

	for i := 0; i < length; i++ {
		msg[i] ^= MaskingKey[i%4]
	}
	sendByte = append(sendByte, msg...)
	_, _ = conn.Write(sendByte)
	return nil
}

// 读取指定长度数据
func (w *WsProtocol) readConnOrCache(conn net.Conn, count int) ([]byte, error) {
	if w.cacheCount > 0 {
		// 拥有缓冲数据
		if count <= w.cacheCount {
			// 缓冲数据比需要的还要大，直接拿取
			msg := w.cacheByte[:count]
			w.cacheCount = w.cacheCount - count
			w.cacheByte = w.cacheByte[count:]
			return msg, nil
		} else {
			// 缓冲数据不足，剩余需要的位数，多读取一点，可以优化速度
			data := make([]byte, count+512)
			cacheCount, err := conn.Read(data)
			if err != nil {
				return nil, errors.New("读取数据失败")
			}
			w.cacheCount = w.cacheCount + cacheCount
			w.cacheByte = append(w.cacheByte, data[:cacheCount]...)
			return w.readConnOrCache(conn, count)
		}
	} else {
		// 缓冲是空的
		data := make([]byte, 1024)
		cacheCount, err := conn.Read(data)
		if err != nil {
			return nil, errors.New("读取数据失败")
		}
		w.cacheCount = cacheCount
		w.cacheByte = append(w.cacheByte, data[:cacheCount]...)
		return w.readConnOrCache(conn, count)
	}
}

func (w *WsProtocol) getHeader(conn net.Conn) ([]byte, error) {
	data := make([]byte, 1024)
	count, err := conn.Read(data)
	if err != nil {
		return nil, errors.New("获取协议头信息错误")
	}

	sep := []byte("\r\n\r\n")
	index := bytes.Index(data, sep)
	if index > 0 {
		// 一般都是比较小数据的头部
		if count != index+4 {
			w.cacheByte = append(w.cacheByte, data[index+4:count]...)
			w.cacheCount = count - index - 4
		}
		return data[:index], nil
	} else {
		return nil, errors.New("协议头异常，不存在分隔符也不应超过1024位")
	}
}

package protocol

import (
	"bytes"
	"goworker/network"
	"net"
)

type TextProtocol struct {
	delim byte
}

func (t *TextProtocol) Init() {
	t.delim = '\n'
}

func (t *TextProtocol) OnConnect(conn net.Conn) (network.Header, error) {
	return nil, nil
}

func (t *TextProtocol) Read(conn net.Conn) ([]byte, error) {
	var message = make([]byte, 1024)
	var start = 0
	var count = 0
	for {
		w, err := conn.Read(message[start:])
		if err != nil {
			return nil, err
		}
		index := bytes.IndexByte(message[start:start+w], t.delim)
		if index >= 0 {
			count = start + index
			break
		}
		start = start + w
	}

	return message[:count], nil
}

func (t *TextProtocol) Write(conn net.Conn, msg []byte) error {
	msg = append(msg, t.delim)
	_, _ = conn.Write(msg)
	return nil
}

package protocol

import (
	"encoding/binary"
	"goworker/network"
	"net"
)

var PackageLen uint8 = 4

// 定长协议，长度头 + 报文
type PackageLenProtocol struct {
	PackageLen uint8
}

func (self *PackageLenProtocol) Init() {
	if self.PackageLen == 0 {
		self.PackageLen = PackageLen
	}
}

func (*PackageLenProtocol) OnConnect(conn net.Conn) (network.Header, error) {
	return nil, nil
}

func (self *PackageLenProtocol) Read(conn net.Conn) ([]byte, error) {
	var buf = make([]byte, self.PackageLen)
	_, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}
	PackageLen := uint32(binary.BigEndian.Uint32(buf))
	var data = make([]byte, PackageLen-4)
	_, err = conn.Read(data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (*PackageLenProtocol) Write(conn net.Conn, msg []byte) error {
	var buf32 = make([]byte, 4)
	packLen := len(msg) + 4
	binary.BigEndian.PutUint32(buf32, uint32(packLen))
	msg = append(buf32, msg...)
	_, err := conn.Write(msg)
	return err
}

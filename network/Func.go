package network

import (
	"encoding/binary"
	"encoding/hex"
	"net"
	"regexp"
	"strconv"
)

// 构建分布式唯一id
func Bin2hex(ip uint32, port uint16, id uint32) string {
	var msgByte []byte
	var buf32 = make([]byte, 4)
	var bug16 = make([]byte, 2)
	binary.BigEndian.PutUint32(buf32, ip)
	msgByte = append(msgByte, buf32...)
	binary.BigEndian.PutUint16(bug16, port)
	msgByte = append(msgByte, bug16...)
	binary.BigEndian.PutUint32(buf32, id)
	msgByte = append(msgByte, buf32...)
	return hex.EncodeToString(msgByte)
}

// 构建分布式唯一id转换回去
func DecodeBin2hex(clientId string) (uint32, uint16, uint32) {
	msgByte, _ := hex.DecodeString(clientId)
	ip4 := msgByte[0:4]
	port2 := msgByte[4:6]
	id4 := msgByte[6:]

	ip := binary.BigEndian.Uint32(ip4)
	port := binary.BigEndian.Uint16(port2)
	id := binary.BigEndian.Uint32(id4)
	return ip, port, id
}

func Ip2long(ipstr string) (ip uint32) {
	r := `^(\d{1,3})\.(\d{1,3})\.(\d{1,3})\.(\d{1,3})`
	reg, err := regexp.Compile(r)
	if err != nil {
		return
	}
	ips := reg.FindStringSubmatch(ipstr)
	if ips == nil {
		return
	}

	ip1, _ := strconv.Atoi(ips[1])
	ip2, _ := strconv.Atoi(ips[2])
	ip3, _ := strconv.Atoi(ips[3])
	ip4, _ := strconv.Atoi(ips[4])

	if ip1 > 255 || ip2 > 255 || ip3 > 255 || ip4 > 255 {
		return
	}

	ip += uint32(ip1 * 0x1000000)
	ip += uint32(ip2 * 0x10000)
	ip += uint32(ip3 * 0x100)
	ip += uint32(ip4)

	return ip
}

func Long2Ip(ip uint32) string {
	a := byte((ip >> 24) & 0xFF)
	b := byte((ip >> 16) & 0xFF)
	c := byte((ip >> 8) & 0xFF)
	d := byte(ip & 0xFF)
	return net.IPv4(a, b, c, d).String()
}

package network

import (
	"net/url"
	"strconv"
	"strings"
)

type Url struct {
	Origin string // ws://host:port
	Scheme string
	Host   string // host:port
	Path   string // test.go
	Ip     string
	Port   uint16
}

func NewUrl(addr string) *Url {
	parse, err := url.Parse(addr)
	if err != nil {
		panic("地址格式错误")
	}
	arr := strings.Split(parse.Host, ":")
	port, _ := strconv.Atoi(parse.Port())
	return &Url{
		Origin: parse.Scheme + "://" + parse.Host,
		Scheme: parse.Scheme,
		Host:   parse.Host,
		Path:   parse.Path,
		Ip:     arr[0],
		Port:   uint16(port),
	}
}

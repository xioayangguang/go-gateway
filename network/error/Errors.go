package error

import (
	"fmt"
	"goworker/network"
)

// 连接失败
type ListenError struct {
	Address *network.Url
}

func (e *ListenError) Error() string {
	return fmt.Sprintf("连接失败 :%s", e.Address.Host)
}

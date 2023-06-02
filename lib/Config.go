package lib

import (
	"goworker/network"
	c "goworker/utils"
)

var BussinessEvent LogicEvent

var Config = config{}

type config struct {
	Gateway   *network.Url // 网关对外暴露的
	Worker    *network.Url // 网关对内业务处理（内部通讯）地址
	Register  *network.Url // 注册中心
	SecretKey string
}

func (self *config) SetInput() {

	if c.GlobalConfig.Register != "" {
		self.Register = network.NewUrl("text://" + c.GlobalConfig.Register)
	}

	if c.GlobalConfig.Gateway != "" {
		self.Gateway = network.NewUrl("ws://" + c.GlobalConfig.Gateway)
	}

	if c.GlobalConfig.GatewayInternal != "" {
		self.Worker = network.NewUrl("pack://" + c.GlobalConfig.GatewayInternal)
	}

	self.SecretKey = c.GlobalConfig.Secret
}

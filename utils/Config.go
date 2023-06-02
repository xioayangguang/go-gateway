package utils

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

type Config struct {
	Debug           bool   `mapstructure:"debug" json:"debug" yaml:"debug"`
	Gateway         string `mapstructure:"gateway" json:"gateway" yaml:"gateway"`
	GatewayInternal string `mapstructure:"gatewayInternal" json:"gatewayInternal" yaml:"gatewayInternal"`
	Register        string `mapstructure:"register" json:"register" yaml:"register"`
	Secret          string `mapstructure:"secret" json:"secret" yaml:"secret"`
}

var GlobalConfig Config

func init() {
	v := viper.New()
	v.SetConfigFile("config.yaml")
	err := v.ReadInConfig()
	if err != nil {
		panic(any(fmt.Errorf("Fatal error config file: %s \n", err)))
	}
	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("config file changed:", e.Name)
		if err := v.Unmarshal(&GlobalConfig); err != nil {
			fmt.Println(err)
		}
	})
	if err := v.Unmarshal(&GlobalConfig); err != nil {
		fmt.Println(err)
	}
}

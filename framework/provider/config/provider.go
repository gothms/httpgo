package config

import (
	"github.com/gothms/httpgo/framework"
	"github.com/gothms/httpgo/framework/contract"
	"path/filepath"
)

type HttpgoConfigProvider struct{}

var _ framework.ServiceProvider = (*HttpgoConfigProvider)(nil)

func (h HttpgoConfigProvider) Register(c framework.Container) framework.NewInstance {
	return NewHttpgoConfig
}

func (h HttpgoConfigProvider) Boot(c framework.Container) error {
	return nil
}

func (h HttpgoConfigProvider) IsDefer() bool {
	return false
}

func (h HttpgoConfigProvider) Params(c framework.Container) []interface{} {
	appService := c.MustMake(contract.AppKey).(contract.App)
	envService := c.MustMake(contract.EnvKey).(contract.Env)
	env := envService.AppEnv()
	// 配置文件夹地址
	configFolder := appService.ConfigFolder()
	envFolder := filepath.Join(configFolder, env)
	return []interface{}{c, envFolder, envService.All()}
}

func (h HttpgoConfigProvider) Name() string {
	return contract.ConfigKey
}

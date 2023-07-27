package env

import (
	"github.com/gothms/httpgo/framework"
	contract "github.com/gothms/httpgo/framework/contract"
)

type HttpgoEnvProvider struct {
	Folder string
}

const (
	BasePath = "/Users/Documents/workspace/hade/"
)

var _ framework.ServiceProvider = (*HttpgoEnvProvider)(nil)

func (h *HttpgoEnvProvider) Register(c framework.Container) framework.NewInstance {
	return NewHttpgoEnv
}

func (h *HttpgoEnvProvider) Boot(c framework.Container) error {
	app := c.MustMake(contract.AppKey).(contract.App)
	h.Folder = app.BaseFolder()
	//h.Folder = test.BasePath
	return nil
}

func (h *HttpgoEnvProvider) IsDefer() bool {
	return false
}

func (h *HttpgoEnvProvider) Params(c framework.Container) []interface{} {
	return []interface{}{h.Folder}
}

func (h *HttpgoEnvProvider) Name() string {
	return contract.EnvKey
}

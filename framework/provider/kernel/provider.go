package kernel

import (
	"github.com/gothms/httpgo/framework"
	"github.com/gothms/httpgo/framework/contract"
	"github.com/gothms/httpgo/framework/gin"
)

type HttpgoKernelProvider struct {
	HttpEngine *gin.Engine
}

var _ framework.ServiceProvider = (*HttpgoKernelProvider)(nil)

func (kh *HttpgoKernelProvider) Register(c framework.Container) framework.NewInstance {
	return NewHttpgoKernelService
}

func (kh *HttpgoKernelProvider) Boot(c framework.Container) error {
	// 这里调用都会死锁
	//app := c.MustMake(contract.AppKey).(contract.App)
	//fmt.Println(app)
	if kh.HttpEngine == nil {
		kh.HttpEngine = gin.Default()
	}
	kh.HttpEngine.SetContainer(c)
	return nil
}

func (kh *HttpgoKernelProvider) IsDefer() bool {
	return false
}

func (kh *HttpgoKernelProvider) Params(c framework.Container) []interface{} {
	return []interface{}{kh.HttpEngine}
}

func (kh *HttpgoKernelProvider) Name() string {
	return contract.KernelKey
}

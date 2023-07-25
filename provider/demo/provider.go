package demo

import (
	"fmt"
	"github.com/gothms/httpgo/framework"
)

// 服务提供方
type DemoServiceProvider struct{}

func (d DemoServiceProvider) Register(container framework.Container) framework.NewInstance {
	return NewDemoService
}

// Boot 方法我们这里我们什么逻辑都不执行, 只打印一行日志信息
func (d DemoServiceProvider) Boot(container framework.Container) error {
	fmt.Println("demo service boot")
	return nil
}

// IsDefer 方法表示是否延迟实例化，我们这里设置为true，将这个服务的实例化延迟到第一次make的时候
func (d DemoServiceProvider) IsDefer() bool {
	//return false
	return true
}

// Params 方法表示实例化的参数。我们这里只实例化一个参数：container，表示我们在NewDemoService这个函数中，只有一个参数，container
func (d DemoServiceProvider) Params(container framework.Container) []interface{} {
	//fmt.Printf("params %T\n", container)
	return []interface{}{container}
}

func (d DemoServiceProvider) Name() string {
	return Key
}

var _ framework.ServiceProvider = (*DemoServiceProvider)(nil)

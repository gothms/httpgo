package demo

import (
	"github.com/gothms/httpgo/framework"
)

// 具体的接口实例
type DemoService struct {
	// 实现接口
	Service

	// 参数
	c framework.Container
}

// 初始化实例的方法
func NewDemoService(params ...interface{}) (interface{}, error) {
	//*(*string)(unsafe.Pointer(&b))

	//for i := 0; i < len(params); i++ {
	//	fmt.Println("NewDemoService:", i, params[i])
	//	fmt.Printf("%T\n", params[i])
	//	v := *(*framework.HttpgoContainer)(unsafe.Pointer(&params[i]))
	//	fmt.Println(v)
	//}

	// 这里需要将参数展开
	c := params[0].(framework.Container)
	//fmt.Println("new demo service,", c)
	// 返回实例
	return &DemoService{c: c}, nil
	//return &DemoService{c: a}, nil
}

// 实现接口
func (s *DemoService) GetFoo() Foo {
	return Foo{
		Name: "i am foo",
	}
}

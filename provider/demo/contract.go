package demo

// Demo服务的key
const Key = "httpgo:demo"

// Demo服务的接口
type Service interface {
	GetFoo() Foo
}

// Demo服务接口定义的一个数据结构
type Foo struct {
	Name string
}

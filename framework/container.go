package framework

import (
	"errors"
	"fmt"
	"sync"
)

// Container 是一个服务容器，提供绑定服务和获取服务的功能
type Container interface {
	// Bind 绑定一个服务提供者，如果关键字凭证已经存在，会进行替换操作，返回error
	Bind(provider ServiceProvider) error
	// IsBind 关键字凭证是否已经绑定服务提供者
	IsBind(key string) bool

	// Make 根据关键字凭证获取一个服务
	Make(key string) (interface{}, error)
	// MustMake 根据关键字凭证获取一个服务，如果这个关键字凭证未绑定服务提供者，那么会panic。
	// 所以在使用这个接口的时候请保证服务容器已经为这个关键字凭证绑定了服务提供者。
	MustMake(key string) interface{}
	// MakeNew 根据关键字凭证获取一个服务，只是这个服务并不是单例模式的
	// 它是根据服务提供者注册的启动函数和传递的params参数实例化出来的
	// 这个函数在需要为不同参数启动不同实例的时候非常有用
	MakeNew(key string, params []interface{}) (interface{}, error)
}

var _ Container = (*HttpgoContainer)(nil)

// HttpgoContainer 是服务容器的具体实现
type HttpgoContainer struct {
	Container // 实现接口的一种写法
	// providers 存储注册的服务提供者，key为字符串凭证
	providers map[string]ServiceProvider
	// instance 存储具体的实例，key为字符串凭证
	instances map[string]interface{}
	// lock 用于锁住对容器的变更操作
	lock sync.RWMutex
}

func NewHttpgoContainer() *HttpgoContainer {
	return &HttpgoContainer{
		providers: map[string]ServiceProvider{},
		instances: map[string]interface{}{},
		lock:      sync.RWMutex{},
	}
}

// PrintProviders 输出服务容器中注册的关键字
func (h *HttpgoContainer) PrintProviders() []string {
	ret := make([]string, len(h.providers))
	i := 0
	for _, provider := range h.providers {
		//name := provider.Name()
		//line := fmt.Sprint(name)
		//ret[i] = line
		//i++
		ret[i] = fmt.Sprintf("%T", provider)
		i++
	}
	return ret
}

// Bind 将服务容器和关键字做了绑定
func (h *HttpgoContainer) Bind(provider ServiceProvider) error {
	h.lock.Lock()
	defer h.lock.Unlock()
	key := provider.Name()
	h.providers[key] = provider
	for i, sp := range h.providers {
		fmt.Println("key:", key)
		fmt.Printf("%s, %T\n", i, sp)
	}

	// if provider is not defer
	if provider.IsDefer() == false {
		if err := provider.Boot(h); err != nil {
			return err
		}
		// 实例化方法
		//fmt.Println(*h)
		//fmt.Printf("h::%T\n", h)
		params := provider.Params(h)
		method := provider.Register(h)
		instance, err := method(params...)
		if err != nil {
			return errors.New(err.Error())
		}
		h.instances[key] = instance
	}
	return nil
}

func (h *HttpgoContainer) IsBind(key string) bool {
	return h.findServiceProvider(key) != nil
}
func (h *HttpgoContainer) findServiceProvider(key string) ServiceProvider {
	h.lock.RLock()
	defer h.lock.RUnlock()
	if sp, ok := h.providers[key]; ok {
		return sp
	}
	return nil
}
func (h *HttpgoContainer) Make(key string) (interface{}, error) {
	return h.make(key, nil, false)
}

func (h *HttpgoContainer) MustMake(key string) interface{} {
	serv, err := h.make(key, nil, false)
	if err != nil {
		panic(err)
	}
	return serv
}

func (h *HttpgoContainer) MakeNew(key string, params []interface{}) (interface{}, error) {
	return h.make(key, params, true)
}

/*
false：
{<nil> map[httpgo:demo:0x13eb658] map[] {{1 0} 0 0 {{} -1073741824} {{} 0}}}
h::*framework.HttpgoContainer
true：
{<nil> map[httpgo:demo:0x10db658] map[] {{0 0} 0 0 {{} 1} {{} 0}}}
h::*framework.HttpgoContainer
*/
func (h *HttpgoContainer) newInstance(sp ServiceProvider, params []interface{}) (interface{}, error) {
	if err := sp.Boot(h); err != nil {
		return nil, err
	}
	if params == nil {
		//fmt.Println(*h)
		//fmt.Printf("h::%T\n", h)
		params = sp.Params(h)
	}
	method := sp.Register(h)
	//fmt.Println("MakeNew:", method, params)
	instance, err := method(params...)
	//fmt.Println("MakeNew:", instance, method, params)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	return instance, nil
}

// 实例化一个服务
func (h *HttpgoContainer) make(key string, params []interface{}, forceNew bool) (interface{}, error) {
	h.lock.RLock()
	defer h.lock.RUnlock()
	// 查询是否已经注册了这个服务提供者，如果没有，则返回error
	sp := h.findServiceProvider(key)
	if sp == nil {
		return nil, errors.New("contract " + key + " not register yet")
	}
	if forceNew {
		return h.newInstance(sp, params)
	}
	// 不需要强制重新实例化，如果容器中已经实例化了，那么就直接使用容器中的实例
	if ins, ok := h.instances[key]; ok {
		return ins, nil
	}
	// 容器中还未实例化，则进行一次实例化
	inst, err := h.newInstance(sp, nil)
	//fmt.Println("instance:", inst)
	if err != nil {
		return nil, err
	}

	h.instances[key] = inst
	return inst, nil
}

package main

/*
framework.container，Lock 和 Unlock 需要对应解锁	// TODO
	h.lock.Lock()
	defer h.lock.Unlock()
	h.lock.RLock()
	defer h.lock.RUnlock()
interface conversion: []interface {} is not framework.Container: missing method Bind
	bug：
		params[i] != nil，但是 c := params[0].(framework.Container); c == nil
	原因：
		instance, err := method(params...)
		instance, err := method(params)：传的是切片，所以类型转换时与 params[i] 不匹配
	扩展：https://zhuanlan.zhihu.com/p/128711092
		https://stackoverflow.com/questions/69655263/golang-interface-conversion-error-missing-method#:~:text=You%20can%27t%20have%20both%20a%20field%20and%20a,Struct.Interface%20%28%29%2C%20only%20Struct.Interface.Interface%20%28%29.%20Rename%20your%20interface.
syscall.Kill()
	bug
		windows 下 goland 调用不了
	原因
		类似函数调用需要 linux
		示例：//go:build linux && amd64
	解决
		go 文件第一行加上：// +build !windows
	新问题
		windows 和 !windows 同包下的文件之间互相调用不了了
deadlock
	出错位置
		framework.provider.env.provider.go
			func (h *HttpgoEnvProvider) Boot(c framework.Container) error {
				app := c.MustMake(contract.AppKey).(contract.App)
				h.Folder = app.BaseFolder()
				return nil
			}
	报错
		fatal error: all goroutines are asleep - deadlock!
	原因
		container.Bind(&env.HttpgoEnvProvider{}) 时，Bind 方法：
			h.lock.Lock()
			defer h.lock.Unlock()
		而创建 NewHttpgoEnv 实例前，会调用 Provider 的 Boot 方法
		HttpgoEnvProvider 的 Boot 方法又去 MustMake（需要加锁），此时 Bind 的锁还没释放
interface conversion... missing method
	出错位置
		framework.command.env.go
			envServic := container.MustMake(contract.EnvKey).(contract.Env)
	报错
		panic: interface conversion: env.HttpgoEnv is not contract.Env: missing method All
	原因：结构体强转为 interface
		*env.HttpgoEnv 才可以转为 contract.Env
命令行
	linux
		DB_PASSWORD=123 ./httpgo app start
	windows
		？
*/

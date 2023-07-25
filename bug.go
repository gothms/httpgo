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
*/

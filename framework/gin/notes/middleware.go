package notes

/*
middleware
	其实中间件也就是装饰器或闭包，实质上就是一种返回类型为 HandlerFunc 的函数
	通俗地讲，就是一种返回函数的函数，目的就是为了在外层函数中，对内层函数进行装饰或处理
	然后再将被装饰或处理后的内层函数返回。

Logger 中间件
	使用 gin.Default() 实例一个 gin.Engine 默认添加的 Logger() 中间件的处理流程

	由于 HandlerFunc 函数只能接受一个 gin.Context 参数
	因此，在上面源代码中的 LoggerWithConfig(conf LoggerConfig) 函数中
	使用 LoggerConfig 配置，对 HandlerFunc 进行装饰，并返回。

	同样地，在返回的 HandlerFunc 匿名函数中，首先是记录进入该中间件时的一些信息，包括时间
	然后再调用 context.Next() 方法，挂起当前的处理程序，递归去调用后续的中间件，当后续所有中间件和处理函数执行完毕时，再回到此处
	如果要记录该 path 的日志，则再获取一次当前的时间，与开始记录的时间进行计算，即可得出本次请求处理的耗时
	再保存其它信息，包括请求 IP 和响应的相关信息等，最后将该请求的日志进行打印处理
Recovery 中间件
	RecoveryWithWriter(out io.Writer) 函数仅为了对最终返回的中间件 HandlerFunc 函数进行装饰

	在该中间件中，可分为两个逻辑块，一个是 defer，一个是 Next()
	Next() 与 Logger() 中间件中的 Next() 作用类似
	defer 中使用 recover() 来捕获在后续中间件中 panic 的错误信息，并对该错误信息进行处理。

	在该中间件中，首先是判断当前连接是否已中断，然后是进行相关的日志处理，最后，如果连接已中断，则直接设置错误信息，并终止该上下文
	否则，终止该上下文并返回 500 错误响应。
Auth 中间件

*/

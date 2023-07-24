package notes

/*
URL 参数解析
	两种方式
	1.路径中的参数解析
		结合了 Gin 中的路由匹配机制
		ctx.Param->engine.getValue->Param{}
	2.查询字符串的参数解析
		与 Go 自带函数库 net/url 库的区别就是，Gin 将解析后的参数保存在了上下文中
		优点是，对于获取多个参数时，则无需对查询字符串进行重复解析，使获取多个参数时的效率提高了不少
			这也是 Gin 为何效率如此之快的原因之一
		ctx.DefaultQuery->ctx.Request.URL.Query()->func ParseQuery->queryCache
URL form 表单解析
	PostForm/DefaultPostForm/PostFormMap->
	ctx.GetPostForm->func ParseMultipartForm->ctx.formCache
JSON 参数解析
	四种方式
	BindJSON, Bind, ShouldBindJSON, ShouldBind
	Must Bind：BindJSON, Bind，最终也是调用 Should Bind
		对请求进行解析时，若出现错误，会通过 c.AbortWithError(400, err).SetType(ErrorTypeBind) 终止请求
		这会把响应码设置为 400，Content-Type 设置为 text/plain; charset=utf-8，在此之后，若尝试重新设置响应码，则会出现警告
		如将响应码设置为 200：[GIN-debug] [WARNING] Headers were already written. Wanted to override status code 400 with 200
	Should Bind：ShouldBindJSON, ShouldBind
		对请求进行解析时，若出现错误，只会将错误返回，而不会主动进行响应
		所以，在使过程中，如果对产生解析错误的行为有更好的控制，最好使用 Should Bind 一类，自行对错误行为进行处理

	ctx.BindJSON
	ctx.ShouldBindJSON
JSON 解析
	参见 notes\json.go
*/

package notes

/*
IRoutes 接口
	核心接口 IRoutes：提供的是注册路由的抽象
	它的实现 Engine 类似于 Beego 的 ControllerRegister

	Use 方法提供了用户接入自定义逻辑的能力，一般情况下也被看做是插件机制
	AOP 方案

	还额外提供了静态文件的接口

	Gin 没有 Controller 的抽象
	MVC 应该是用户组织 Web 项目的模式，而不是我们中间件设计者要考虑的
Engine 实现
	Engine 类似 Beego 中的 HttpServer 和 ControllerRegister 的合体
		实现了路由树功能，提供了注册和匹配路由的功能
		它本身可以作为一个 Handler 传递到 http 包，用于启动服务器
	Engine 的路由树功能本质上是依赖于 methodTree 的
methodTrees 和 methodTree
	methodTree 才是真实的路由树
	Gin 定义了 methodTrees，它实际上代表的是森林，即每一个HTTP方法都对应到一棵树

	HandlerFunc 定义了核心抽象——处理逻辑
		在默认情况下，它代表了注册路由的业务代码
	HandlersChain 则是构造了责任链模式
													业务逻辑
	HandlerFunc1->HandlerFunc2->HandlerFunc3-->HandlerFunc4
		最后一个则是封装了业务逻辑的 HandlerFunc
Context 抽象
	Context 也是代表了执行的上下文，提供了丰富的API：
		处理请求的API，代表的是以 Get 和 Bind 为前缀的方法
		处理响应的API，例如返回 JSON 或 XML 响应的方法
		渲染页面，如 HTML 方法
	-----req------>			---req--->	|
					Context				| 业务代码
	-----resp------>		---resp--->	|

HandlerFunc		Context		Engine
	|				|			|
		methodTree & node

路由树的实现
	methodTrees：路由树也是按照 HTTP 方式组织的，例如 GET 会有一颗路由树
	methodTree：定义了单棵树
		树在 Gin 里采用的是 children 的定义方式，即树由节点构成
	node：代表树上的一个节点，里面维持住了 children，即子节点
		同时有 nodeType 和 wildChild 来标记一些特殊节点

Engine:
	用来初始化一个gin对象实例，在该对象实例中主要包含了一些框架的基础功能
	比如日志，中间件设置，路由控制(组)，以及handlercontext等相关方法.源码文件


Router:
	用来定义各种路由规则和条件，并通过HTTP服务将具体的路由注册到一个由context实现的handler中


Context:
	Context是框架中非常重要的一点，它允许我们在中间件间共享变量，管理整个流程，验证请求的json以及提供一个json的响应体
	通常情况下我们的业务逻辑处理也是在整个Context引用对象中进行实现的.


Bind:
	在Context中我们已经可以获取到请求的详细信息，比如HTTP请求头和请求体
	但是我们需要根据不同的HTTP协议参数来获取相应的格式化数据来处理底层的业务逻辑，就需要使用Bind相关的结构方法来解析context中的HTTP数据
*/

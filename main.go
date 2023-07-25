package main

import (
	"github.com/gothms/httpgo/app/console"
	httpgo "github.com/gothms/httpgo/app/http"
	"github.com/gothms/httpgo/framework"
	"github.com/gothms/httpgo/framework/provider/app"
	"github.com/gothms/httpgo/framework/provider/kernel"
)

func main() {
	// 初始化服务容器
	container := framework.NewHttpgoContainer()
	// 绑定App服务提供者
	container.Bind(&app.HttpgoAppProvider{})
	// 后续初始化需要绑定的服务提供者...

	// 将HTTP引擎初始化,并且作为服务提供者绑定到服务容器中
	//if engine, err := httpgo.NewHttpEngine(); err == nil {
	//	container.Bind(&kernel.HttpgoKernelProvider{HttpEngine: engine})
	//}
	if engine, err := httpgo.NewHttpEngine(container); err == nil {
		container.Bind(&kernel.HttpgoKernelProvider{HttpEngine: engine})
	}

	// 运行root命令
	console.RunCommand(container)

	// 引入 Command 之前
	//// 创建 engine 结构
	//core := gin.New()
	//// 绑定具体的服务
	//core.Bind(&app.HttpgoAppProvider{})
	//core.Bind(&demo.DemoServiceProvider{})
	////container := core.GetContainer().(*framework.HttpgoContainer)
	////providers := container.PrintProviders()
	////for key, provider := range providers {
	////	fmt.Println("main:", key, provider)
	////}
	//
	//core.Use(gin.Recovery())
	////core.Use(middleware.Timeout(time.Millisecond * 100))
	//core.Use(middleware.Cost())
	//
	//// 注册路由
	////registerRouter(core)
	//httpgo.Routes(core)
	//
	//server := http.Server{
	//	Handler: core,
	//	Addr:    ":8080",
	//}
	//// 启动服务的goroutine
	//go func() { server.ListenAndServe() }()
	//
	//// 当前 goroutine 等待信号量
	//quit := make(chan os.Signal)
	//// 监控信号：SIGINT, SIGTERM, SIGQUIT
	//signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	//// 会阻塞当前goroutine等待信号
	//r := <-quit
	//fmt.Println(r)
	//
	////调用Server.Shutdown graceful结束
	//timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	//defer cancel()
	////defer func() {
	////	fmt.Println("cancel")
	////	cancel()
	////}()
	//if err := server.Shutdown(timeoutCtx); err != nil {
	//	log.Fatal("Server shutdown:", err)
	//}
}

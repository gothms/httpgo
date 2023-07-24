package main

import (
	"context"
	"fmt"
	"github.com/gothms/httpgo/framework/gin"
	"github.com/gothms/httpgo/framework/middleware"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	core := gin.New()
	core.Use(gin.Recovery())
	//core.Use(middleware.Timeout(time.Millisecond * 100))
	core.Use(middleware.Cost())
	registerRouter(core)
	server := http.Server{
		Handler: core,
		Addr:    ":8080",
	}
	go func() { server.ListenAndServe() }()

	// 当前 goroutine 等待信号量
	quit := make(chan os.Signal)
	// 监控信号：SIGINT, SIGTERM, SIGQUIT
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	// 会阻塞当前goroutine等待信号
	r := <-quit
	fmt.Println(r)

	//调用Server.Shutdown graceful结束
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	//defer func() {
	//	fmt.Println("cancel")
	//	cancel()
	//}()
	if err := server.Shutdown(timeoutCtx); err != nil {
		log.Fatal("Server shutdown:", err)
	}
}

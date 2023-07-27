package test

import "testing"

//	func TestGin(t *testing.T) {
//		r := gin.Default()
//		r.GET("/ping", func(c *gin.Context) {
//			c.JSON(200, gin.H{
//				"message": "pong",
//			})
//		})
//		r.Run() // 监听并在 0.0.0.0:8080 上启动服务
//	}
//
//	func TestHTTP(t *testing.T) {
//		// 禁用控制台颜色
//		// gin.DisableConsoleColor()
//
//		// 使用默认中间件（logger 和 recovery 中间件）创建 gin 路由
//		router := gin.Default()
//
//		//router.GET("/someGet", getting)
//		//router.POST("/somePost", posting)
//		//router.PUT("/somePut", putting)
//		//router.DELETE("/someDelete", deleting)
//		//router.PATCH("/somePatch", patching)
//		//router.HEAD("/someHead", head)
//		//router.OPTIONS("/someOptions", options)
//
//		// 默认在 8080 端口启动服务，除非定义了一个 PORT 的环境变量。
//		router.Run()
//		// router.Run(":3000") hardcode 端口号
//	}
func TestInterface(t *testing.T) {
	d1 := D{}
	d2 := &D{}
	t.Logf("%T\n", d1)
	t.Logf("%T\n", d2)
}

type D struct{}

package main

import (
	"github.com/gothms/httpgo/framework/gin"
	"github.com/gothms/httpgo/framework/middleware"
)

func registerRouter(engine *gin.Engine) {
	engine.GET("/user/login", middleware.Test3(), UserLoginController)
	// 批量通用前缀
	subjectApi := engine.Group("/subject")
	{
		subjectApi.Use(middleware.Test3())
		// 动态路由
		subjectApi.DELETE("/:id", SubjectDelController)
		subjectApi.PUT("/:id", SubjectUpdateController)
		subjectApi.GET("/:id", middleware.Test2(), SubjectGetController)
		subjectApi.GET("/list/all", SubjectListController)
		subjectInnerApi := subjectApi.Group("/info")
		{
			subjectInnerApi.GET("/name", SubjectNameController)
		}
	}
	//core.Get("/foo", FooControllerHandler)

	//subjectApi.Delete("/:id" , SubjectDelController)
	//subjectApi.Put("/:id", SubjectUpdateController)
	//subjectApi.Get("/:id", SubjectGetController)
	//subjectApi.Get("/list/all", SubjectListController)
	//subjectInnerApi := subjectApi.Group("/info")
	//subjectInnerApi.Get("/name", SubjectNameController)
}

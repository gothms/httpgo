package httpgo

import (
	"github.com/gothms/httpgo/app/http/module/demo"
	"github.com/gothms/httpgo/framework/gin"
)

func Routes(r *gin.Engine) {
	r.Static("/dist", "./dist/")
	demo.Register(r)
}

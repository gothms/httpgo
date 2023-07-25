package kernel

import (
	"github.com/gothms/httpgo/framework/gin"
	"net/http"
)

// 引擎服务
type HttpgoKernelService struct {
	engine *gin.Engine
}

// 初始化 web 引擎服务实例
func NewHttpgoKernelService(params ...interface{}) (interface{}, error) {
	httpEngine := params[0].(*gin.Engine)
	return &HttpgoKernelService{engine: httpEngine}, nil
}

// 返回 web 引擎
func (hk *HttpgoKernelService) HttpEngine() http.Handler {
	return hk.engine
}

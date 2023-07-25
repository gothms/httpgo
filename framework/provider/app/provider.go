package app

import (
	"github.com/gothms/httpgo/framework"
	"github.com/gothms/httpgo/framework/contract"
)

// HttpgoAppProvider 提供App的具体实现方法
type HttpgoAppProvider struct {
	BaseFolder string
}

// Register 注册HadeApp方法
func (h *HttpgoAppProvider) Register(container framework.Container) framework.NewInstance {
	return NewHttpgoApp
}

// Boot 启动调用
func (h *HttpgoAppProvider) Boot(container framework.Container) error {
	return nil
}

// IsDefer 是否延迟初始化
func (h *HttpgoAppProvider) IsDefer() bool {
	return false
}

// Params 获取初始化参数
func (h *HttpgoAppProvider) Params(container framework.Container) []interface{} {
	return []interface{}{container, h.BaseFolder}
}

// Name 获取字符串凭证
func (h *HttpgoAppProvider) Name() string {
	return contract.AppKey
}

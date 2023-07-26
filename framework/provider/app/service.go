package app

import (
	"errors"
	"flag"
	"github.com/google/uuid"
	"github.com/gothms/httpgo/framework"
	"github.com/gothms/httpgo/framework/contract"
	"github.com/gothms/httpgo/framework/util"
	"path/filepath"
)

// HttpgoApp 代表 httpgo 框架的 App 实现
type HttpgoApp struct {
	container  framework.Container // 服务容器
	baseFolder string              // 基础路径
	appId      string              // 表示当前这个app的唯一id, 可以用于分布式锁等
}

var _ contract.App = (*HttpgoApp)(nil)

func (h HttpgoApp) AppID() string {
	return h.appId
}

// Version 实现版本
func (h HttpgoApp) Version() string {
	return "0.0.1"
}

// BaseFolder 表示基础目录，可以代表开发场景的目录，也可以代表运行时候的目录
func (h HttpgoApp) BaseFolder() string {
	if h.baseFolder != "" {
		return h.baseFolder
	}
	// 如果没有设置，则使用参数
	//var baseFolder string
	//flag.StringVar(&baseFolder, "base_folder", "", "base_folder参数, 默认为当前路径")
	//flag.Parse()
	//if baseFolder != "" {
	//	return baseFolder
	//}

	// 如果参数也没有，使用默认的当前路径
	return util.GetExecDirectory()
}

// ConfigFolder  表示配置文件地址
func (h HttpgoApp) ConfigFolder() string {
	return filepath.Join(h.BaseFolder(), "config")
}

// LogFolder 表示日志存放地址
func (h HttpgoApp) LogFolder() string {
	return filepath.Join(h.StorageFolder(), "log")
}

func (h HttpgoApp) HttpFolder() string {
	return filepath.Join(h.BaseFolder(), "http")
}

func (h HttpgoApp) ConsoleFolder() string {
	return filepath.Join(h.BaseFolder(), "console")
}

func (h HttpgoApp) StorageFolder() string {
	return filepath.Join(h.BaseFolder(), "storage")
}

// ProviderFolder 定义业务自己的服务提供者地址
func (h HttpgoApp) ProviderFolder() string {
	return filepath.Join(h.BaseFolder(), "provider")
}

// MiddlewareFolder 定义业务自己定义的中间件
func (h HttpgoApp) MiddlewareFolder() string {
	return filepath.Join(h.HttpFolder(), "middleware")
}

// CommandFolder 定义业务定义的命令
func (h HttpgoApp) CommandFolder() string {
	return filepath.Join(h.ConsoleFolder(), "command")
}

// RuntimeFolder 定义业务的运行中间态信息
func (h HttpgoApp) RuntimeFolder() string {
	return filepath.Join(h.StorageFolder(), "runtime")
}

// TestFolder 定义测试需要的信息
func (h HttpgoApp) TestFolder() string {
	return filepath.Join(h.BaseFolder(), "test")
}

// NewHttpgoApp 初始化 HttpgoApp
func NewHttpgoApp(params ...interface{}) (interface{}, error) {
	if len(params) != 2 {
		return nil, errors.New("params error")
	}
	// 有两个参数，一个是容器，一个是 baseFolder
	container := params[0].(framework.Container)
	baseFolder := params[1].(string)

	// 如果没有设置，则使用参数
	if baseFolder == "" {
		flag.StringVar(&baseFolder, "base_folder", "", "base_folder参数, 默认为当前路径")
		flag.Parse()
	}
	appId := uuid.New().String()
	return &HttpgoApp{container: container, baseFolder: baseFolder, appId: appId}, nil
}

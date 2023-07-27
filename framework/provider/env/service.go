package env

import (
	"bufio"
	"bytes"
	"errors"
	"github.com/gothms/httpgo/framework/contract"
	"io"
	"os"
	"path"
	"strings"
)

// HttpgoEnv 是 Env 的具体实现
type HttpgoEnv struct {
	folder string            // 代表.env所在的目录
	envs   map[string]string // 保存所有的环境变量
}

var _ contract.Env = (*HttpgoEnv)(nil)

// NewHttpgoEnv 参数为 .env 文件所在的目录
// example: NewHttpgoEnv("/envfolder/") 会读取文件: /envfolder/.env
// .env 的文件格式 FOO_ENV=BAR
func NewHttpgoEnv(params ...interface{}) (interface{}, error) {
	if len(params) != 1 {
		return nil, errors.New("NewHttpgoEnv param error")
	}
	// 读取 folder 文件
	folder, _ := params[0].(string)
	// 实例化
	hgEnv := HttpgoEnv{
		folder: folder,
		// 实例化环境变量，APP_ENV默认设置为开发环境
		envs: map[string]string{"APP_ENV": contract.EnvDevelopment},
	}
	// 解析 folder/.env 文件
	file := path.Join(folder, ".env")
	// 读取.env文件, 不管任意失败，都不影响后续
	// 打开文件.env
	f, err := os.Open(file)
	if err == nil {
		defer f.Close()
		// 读取文件
		buf := bufio.NewReader(f)
		for {
			// 按照行进行读取
			line, _, c := buf.ReadLine()
			if c == io.EOF {
				break
			}
			// 按照等号解析
			index := bytes.IndexRune(line, '=')
			if index < 0 { // 如果不符合规范，则过滤
				continue
			}
			hgEnv.envs[string(line[:index])] = string(line[index+1:])
		}
	}
	// 获取当前程序的环境变量，并且覆盖.env文件下的变量
	for _, env := range os.Environ() {
		//fmt.Println(env)
		index := strings.IndexRune(env, '=')
		if index < 0 {
			continue
		}
		//hgEnv.envs["APP_ENV"] = contract.EnvTesting
		hgEnv.envs[env[:index]] = env[index+1:]
	}
	return &hgEnv, nil
}
func (h *HttpgoEnv) AppEnv() string {
	return h.Get("APP_ENV")
}

func (h *HttpgoEnv) IsExist(key string) bool {
	_, ok := h.envs[key]
	return ok
}

func (h *HttpgoEnv) Get(key string) string {
	if val, ok := h.envs[key]; ok {
		return val
	}
	return ""
}

func (h *HttpgoEnv) All() map[string]string {
	return h.envs
}

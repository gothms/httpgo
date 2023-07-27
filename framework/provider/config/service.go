package config

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/gothms/httpgo/framework"
	"github.com/gothms/httpgo/framework/contract"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cast"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type HttpgoConfig struct {
	c        framework.Container    // 容器
	folder   string                 // 配置本地配置文件所在的文件夹
	keyDelim string                 // 路径的分隔符，默认为点
	lock     sync.RWMutex           // 配置文件读写锁
	envMaps  map[string]string      // 所有的环境变量
	confMaps map[string]interface{} // 每个配置解析后的结构，key为文件名
	confRaws map[string][]byte      // 每个配置的原始文件信息
}

var _ contract.Config = (*HttpgoConfig)(nil)

// NewHttpgoConfig 初始化 Config
func NewHttpgoConfig(params ...interface{}) (interface{}, error) {
	container := params[0].(framework.Container)
	envFolder := params[1].(string)
	envMaps := params[2].(map[string]string)
	// 检查文件夹是否存在
	if _, err := os.Stat(envFolder); os.IsNotExist(err) {
		return nil, errors.New("folder " + envFolder + " not exist: " + err.Error())
	}
	httpgoConf := &HttpgoConfig{
		c:        container,
		folder:   envFolder,
		envMaps:  envMaps,
		confMaps: map[string]interface{}{},
		confRaws: map[string][]byte{},
		keyDelim: ".",
		lock:     sync.RWMutex{},
	}
	// 读取每个文件
	files, err := ioutil.ReadDir(envFolder)
	if err != nil {
		//return nil, errors.WithStack(err)	// linux 方法
		return nil, errors.New(err.Error())
	}
	for _, file := range files {
		fileName := file.Name()
		err := httpgoConf.loadConfigFile(envFolder, fileName)
		if err != nil {
			log.Println(err)
			continue
		}
	}

	// 监控文件夹文件
	watch, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	err = watch.Add(envFolder)
	if err != nil {
		return nil, err
	}
	go func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println(err)
			}
		}()

		for {
			select {
			case ev := <-watch.Events:
				{
					//判断事件发生的类型
					// Create 创建
					// Write 写入
					// Remove 删除
					path, _ := filepath.Abs(ev.Name)
					index := strings.LastIndex(path, string(os.PathSeparator))
					folder := path[:index]
					fileName := path[index+1:]

					if ev.Op&fsnotify.Create == fsnotify.Create {
						log.Println("创建文件 : ", ev.Name)
						httpgoConf.loadConfigFile(folder, fileName)
					}
					if ev.Op&fsnotify.Write == fsnotify.Write {
						log.Println("写入文件 : ", ev.Name)
						httpgoConf.loadConfigFile(folder, fileName)
					}
					if ev.Op&fsnotify.Remove == fsnotify.Remove {
						log.Println("删除文件 : ", ev.Name)
						httpgoConf.removeConfigFile(folder, fileName)
					}
				}
			case err := <-watch.Errors:
				{
					log.Println("error : ", err)
					return
				}
			}
		}
	}()
	return httpgoConf, nil
}

// 删除文件的操作
func (conf *HttpgoConfig) removeConfigFile(folder string, file string) error {
	strs := strings.Split(file, ".")
	// 只有yaml或者yml后缀才执行
	if len(strs) == 2 && (strs[1] == "yaml" || strs[1] == "yml") {
		conf.lock.Lock()
		// 删除内存中对应的key
		delete(conf.confRaws, strs[0])
		delete(conf.confMaps, strs[0])
		conf.lock.Unlock()
	}
	return nil
}

// 读取某个配置文件
func (conf *HttpgoConfig) loadConfigFile(folder string, file string) error {
	// 判断文件是否是 yaml 和 yml 文件
	ss := strings.Split(file, ".")
	if len(ss) == 2 && (ss[1] == "yaml" || ss[1] == "yml") {
		conf.lock.Lock()
		defer conf.lock.Unlock()
		// 读取文件内容
		buf, err := ioutil.ReadFile(filepath.Join(folder, file))
		if err != nil {
			return err
		}
		// 直接对文本做环境变量替换
		buf = replace(buf, conf.envMaps)
		// 解析对应文件
		cm := map[string]interface{}{}
		if err = yaml.Unmarshal(buf, &cm); err != nil {
			return err
		}
		conf.confMaps[ss[0]] = cm
		conf.confRaws[ss[0]] = buf
		// 读取 app.path 中的信息，更新 app 对应的 folder
		if ss[0] == "app" && conf.c.IsBind(contract.AppKey) {
			if p, ok := cm["path"]; ok {
				appService := conf.c.MustMake(contract.AppKey).(contract.App)
				appService.LoadAppConfig(cast.ToStringMapString(p))
			}
		}
	}
	return nil
}

// replace 表示使用环境变量 envs 替换 content 中的 env(xxx) 的环境变量
func replace(content []byte, envs map[string]string) []byte {
	if envs == nil || len(envs) == 0 {
		return content
	}
	// 直接使用ReplaceAll替换。这个性能可能不是最优，但是配置文件加载，频率是比较低的，可以接受
	for key, val := range envs {
		// TODO 这种写法非常低效
		reKey := "env(" + key + ")"
		content = bytes.ReplaceAll(content, []byte(reKey), []byte(val))
	}
	return content
}

// 查找某个路径的配置项
func searchMap(source map[string]interface{}, path []string) interface{} {
	if len(path) == 0 {
		return source
	}
	next, ok := source[path[0]]
	if ok {
		if len(path) == 1 { // 判断这个路径是否为1
			return next
		}
		switch next.(type) { // 判断下一个路径的类型
		case map[interface{}]interface{}: // 如果是interface的map，使用cast进行下value转换
			return searchMap(cast.ToStringMap(next), path[1:])
		case map[string]interface{}: // 如果是map[string]，直接循环调用
			return searchMap(next.(map[string]interface{}), path[1:])
		default:
			return nil // 否则的话，返回nil
		}
	}
	return nil
}

// 通过path来获取某个配置项
func (conf *HttpgoConfig) find(key string) interface{} {
	conf.lock.RLock()
	defer conf.lock.RUnlock()
	return searchMap(conf.confMaps, strings.Split(key, conf.keyDelim))
}
func (conf *HttpgoConfig) IsExist(key string) bool {
	return conf.find(key) != nil
}

func (conf *HttpgoConfig) Get(key string) interface{} {
	return conf.find(key)
}

func (conf *HttpgoConfig) GetBool(key string) bool {
	return cast.ToBool(conf.find(key))
}

func (conf *HttpgoConfig) GetInt(key string) int {
	return cast.ToInt(conf.find(key))
}

func (conf *HttpgoConfig) GetFloat64(key string) float64 {
	return cast.ToFloat64(conf.find(key))
}

func (conf *HttpgoConfig) GetTime(key string) time.Time {
	return cast.ToTime(conf.find(key))
}

func (conf *HttpgoConfig) GetString(key string) string {
	return cast.ToString(conf.find(key))
}

func (conf *HttpgoConfig) GetIntSlice(key string) []int {
	return cast.ToIntSlice(conf.find(key))
}

func (conf *HttpgoConfig) GetStringSlice(key string) []string {
	return cast.ToStringSlice(conf.find(key))
}

func (conf *HttpgoConfig) GetStringMap(key string) map[string]interface{} {
	return cast.ToStringMap(conf.find(key))
}

func (conf *HttpgoConfig) GetStringMapString(key string) map[string]string {
	return cast.ToStringMapString(conf.find(key))
}

func (conf *HttpgoConfig) GetStringMapStringSlice(key string) map[string][]string {
	return cast.ToStringMapStringSlice(conf.find(key))
}

func (conf *HttpgoConfig) Load(key string, val interface{}) error {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "yaml",
		Result:  val,
	})
	if err != nil {
		return err
	}
	return decoder.Decode(conf.find(key))
}

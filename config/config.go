package config

import (
	"fmt"
	"github.com/hashicorp/consul/api"
	"github.com/murang/potato/util"
	"reflect"
	"sync"
)

// IConfig 配置接口 管理配置数据的对象需要实现这个接口
type IConfig interface {
	Name() string  // 配置名称
	Path() string  // 配置文件路径 如果是consul的kv 则为kv前缀
	ValuePtr() any // 返回用于解析配置结构体的指针
	OnLoad()       // 配置文件加载后的回调 用于处理一些需要配置文件加载后的逻辑
}

// FocusConsulConfig 关注consul配置
// 使用consul管理配置的时候 可以使用这个方法注册需要关注的配置
// ⚠️ 需要在SetConsul之前调用这个方法 让watchConfigUpdate知道哪些是需要解析的配置文件
func FocusConsulConfig(config IConfig) {
	if consulClient != nil {
		panic("SetConsul must be called after FocusConsulConfig")
		return
	}
	if config == nil {
		panic("config is nil")
		return
	}
	group, ok := groups[config.Name()]
	if ok {
		panic("config name already exists")
		return
	}
	group = &Group{
		Name:       config.Name(),
		Path:       config.Path(),
		ConfigType: reflect.TypeOf(config),
		ConfigMap:  &sync.Map{},
	}
	groups[config.Name()] = group
}

// SetConsul 设置consul地址
func SetConsul(addr string) {
	cli, err := api.NewClient(&api.Config{
		Address: addr,
	})
	if err != nil {
		panic(fmt.Sprintf("Config SetConsul NewClient err: %s", err))
	}
	consulClient = cli
	util.GoSafe(watchConfigUpdate)
}

// OnConsulConfigChange 注册consul配置变更回调
func OnConsulConfigChange(f func(IConfig)) {
	onConfigChange = append(onConfigChange, f)
}

// LoadConfig 加载本地配置 如果有tag则加载tag配置
func LoadConfig(config IConfig, tag ...string) {
	if config == nil {
		panic("config is nil")
		return
	}
	group, ok := groups[config.Name()]
	if ok {
		panic("config name already exists")
		return
	}
	name := config.Name()
	path := config.Path()
	configType := reflect.TypeOf(config)
	group = &Group{
		Name:       name,
		Path:       path,
		ConfigType: configType,
		ConfigMap:  &sync.Map{},
	}
	groups[name] = group

	// 加载默认配置
	if LoadConfigFromFile(name, path, config) {
		group.ConfigMap.Store(name, config)
	}
	// 加载tag配置
	if len(tag) > 0 {
		for _, t := range tag {
			tagConfig := reflect.New(configType.Elem()).Interface().(IConfig)
			if LoadConfigFromFile(fmt.Sprintf("%s_%s", name, t), path, tagConfig) {
				group.ConfigMap.Store(fmt.Sprintf("%s_%s", name, t), tagConfig)
			}
		}
	}
}

// GetConfig 获取配置
func GetConfig[T IConfig]() T {
	var cfg T
	name := cfg.Name()
	group, ok := groups[name]
	if !ok {
		var zero T
		return zero
	}

	config, ok := group.ConfigMap.Load(name)
	if !ok {
		var zero T
		return zero
	}

	return config.(T)
}

// GetConfigWithTag 获取tag配置 fallback为true的话找不到tag则返回默认配置
func GetConfigWithTag[T IConfig](tag string, fallback bool) T {
	var cfg T
	name := cfg.Name()
	group, ok := groups[name]
	if !ok {
		var zero T
		return zero
	}

	config, ok := group.ConfigMap.Load(fmt.Sprintf("%s_%s", name, tag))
	if !ok {
		if fallback {
			return GetConfig[T]()
		}
		var zero T
		return zero
	}

	return config.(T)
}

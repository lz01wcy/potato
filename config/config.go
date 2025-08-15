package config

import (
	"fmt"
	"github.com/hashicorp/consul/api"
	"github.com/murang/potato/util"
	"reflect"
	"sync"
)

// FocusConsulConfig 关注consul配置
// 使用consul管理配置的时候 可以使用这个方法注册需要关注的配置
func FocusConsulConfig(name, path string, config IConfig) {
	if name == "" || config == nil {
		panic("config name or config is nil")
		return
	}
	group, ok := groups[name]
	if ok {
		panic("config name already exists")
		return
	}
	group = &Group{
		Name:       name,
		Path:       path,
		ConfigType: reflect.TypeOf(config),
		ConfigMap:  &sync.Map{},
	}
	groups[name] = group
}

// IConfig 配置接口 管理配置数据的对象需要实现这个接口
type IConfig interface {
	ValuePtr() any // 返回用于解析配置结构体的指针
	OnLoad()       // 配置文件加载后的回调
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
func OnConsulConfigChange(f func(string, string, IConfig)) {
	onConfigChange = append(onConfigChange, f)
}

// LoadConfig 加载本地配置 如果有tag则加载tag配置
func LoadConfig(name, path string, config IConfig, tag ...string) {
	if name == "" || config == nil {
		panic("config name or config is nil")
		return
	}
	group, ok := groups[name]
	if ok {
		panic("config name already exists")
		return
	}
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
			tagConfig := reflect.New(configType).Interface().(IConfig)
			if LoadConfigFromFile(fmt.Sprintf("%s_%s", name, t), path, tagConfig) {
				group.ConfigMap.Store(fmt.Sprintf("%s_%s", name, t), tagConfig)
			}
		}
	}
}

// GetConfig 获取配置
func GetConfig(name string) IConfig {
	group, ok := groups[name]
	if !ok {
		return nil
	}

	config, ok := group.ConfigMap.Load(name)
	if !ok {
		return nil
	}

	return config.(IConfig)
}

// GetConfigWithTag 获取tag配置 fallback为true的话找不到tag则返回默认配置
func GetConfigWithTag(name, tag string, fallback bool) IConfig {
	group, ok := groups[name]
	if !ok {
		return nil
	}

	config, ok := group.ConfigMap.Load(fmt.Sprintf("%s_%s", name, tag))
	if !ok {
		if fallback {
			return GetConfig(name)
		}
		return nil
	}

	return config.(IConfig)
}

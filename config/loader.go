package config

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/consul/api"
	"github.com/murang/potato/log"
	"os"
)

func LoadConfigFromFile(name, path string, cfg IConfig) bool {
	filePath := fmt.Sprintf("%s/%s.json", path, name)
	f, err := os.ReadFile(filePath)
	if err != nil {
		log.Sugar.Errorf("LoadConfigFile failed with path: %s err: %s", filePath, err.Error())
		return false
	}
	ptr := cfg.ValuePtr()
	err = json.Unmarshal(f, ptr)
	if err != nil {
		log.Sugar.Errorf("LoadConfigFile Unmarshal failed : %s", err.Error())
		return false
	}
	cfg.OnLoad()
	return true
}

func LoadConfigFromBytes(bytes []byte, cfg IConfig) bool {
	ptr := cfg.ValuePtr()
	err := json.Unmarshal(bytes, ptr)
	if err != nil {
		log.Sugar.Errorf("LoadConfigFromBytes Unmarshal failed : %s", err.Error())
		return false
	}
	cfg.OnLoad()
	return true
}

func LoadConfigFromConsul(client *api.Client, name, path string, cfg IConfig) bool {
	consulKey := fmt.Sprintf("%s/%s.json", path, name)
	kv := client.KV()
	pair, _, err := kv.Get(consulKey, nil)
	if err != nil {
		return false
	}
	if pair == nil {
		log.Sugar.Error("key not exist")
		return false
	}

	ptr := cfg.ValuePtr()
	err = json.Unmarshal(pair.Value, ptr)
	if err != nil {
		log.Sugar.Errorf("LoadConfigFromConsul Unmarshal failed : %s", err.Error())
		return false
	}
	cfg.OnLoad()
	return true
}

package config

import (
	"reflect"
	"sync"
)

var groups = make(map[string]*Group)

type Group struct {
	Name       string
	Path       string
	ConfigType reflect.Type
	ConfigMap  *sync.Map
}

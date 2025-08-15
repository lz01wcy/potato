package pool

import (
	"reflect"
	"sync"
)

var poolMap = sync.Map{}

func getPool(t reflect.Type) *sync.Pool {
	pool, ok := poolMap.Load(t)
	if !ok {
		pool = &sync.Pool{
			New: func() interface{} {
				if t.Kind() == reflect.Ptr {
					return reflect.New(t.Elem()).Interface()
				} else {
					return reflect.New(t).Interface()
				}
			},
		}
		poolMap.Store(t, pool)
	}
	return pool.(*sync.Pool)
}

func Get(reflectType reflect.Type) interface{} {
	return getPool(reflectType).Get()
}

func Put(t reflect.Type, obj interface{}) {
	getPool(t).Put(obj)
}

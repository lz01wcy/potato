package potato

import (
	"github.com/murang/potato/app"
	"sync"
)

var (
	instance *app.Application
	once     sync.Once
)

func Instance() *app.Application {
	once.Do(func() {
		instance = app.NewApplication()
	})
	return instance
}

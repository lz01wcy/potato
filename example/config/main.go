package main

import (
	"github.com/murang/potato"
	"github.com/murang/potato/log"
)

func main() {
	potato.RegisterModule(&WorkModule{})

	potato.Start(func() bool { // 初始化app 入参为启动函数 在初始化所有组件后执行
		log.Logger.Info("all module started, server start")
		return true
	})
	potato.Run()        // 开始update 所有组件开始tick 主线程阻塞
	potato.End(func() { // 主线程开始退出 所有组件销毁后执行入参函数
		log.Logger.Info("all module stopped, server stop")
	})
}

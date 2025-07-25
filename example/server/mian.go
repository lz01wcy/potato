package main

import (
	"github.com/murang/potato/app"
	_ "github.com/murang/potato/example/nicepb/nice"
	"github.com/murang/potato/net"
)

const (
	NiceModuleId = 1
)

func main() {
	a := app.Instance()

	// 添加模块 模块需要实现IModule 可以把组件理解为unity的组件 有自己的生命周期
	// 一个app可以注册多个模块 每个模块有自己的帧率 帧率设置为0的话就是不tick
	a.RegisterModule(NiceModuleId, &NiceModule{})

	// 网络管理器 可传入配置
	netManager := net.NewManager(net.WithCodec(&net.PbCodec{})) // 使用protobuf
	// 设置消息处理器 消息处理器需要实现IMsgHandler
	netManager.SetMsgHandler(&MsgHandler{})
	// 网络监听器 支持tcp/kcp/ws 可支持同时接收多个监听器消息
	ln, err := net.NewListener("tcp", ":10086")
	if err != nil {
		panic(err)
	}
	netManager.AddListener(ln)
	netManager.Start() // 启动网络管理器 收到消息后会回调消息处理器的对应方法

	a.Init(nil)    // 初始化app 入参为启动函数 在初始化所有组件后执行
	a.StartRun()   // 启动app 所有组件开始tick 主线程阻塞
	a.Destroy(nil) // 主线程开始退出 所有组件销毁后执行入参函数
}

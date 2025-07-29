package main

import (
	"github.com/murang/potato/app"
	"github.com/murang/potato/log"
	"github.com/murang/potato/net"
	"github.com/murang/potato/rpc"
)

const (
	NiceModuleId = iota
)

func main() {
	a := app.Instance()

	// 添加模块 模块需要实现IModule 可以把组件理解为unity的组件 有自己的生命周期
	// 一个app可以注册多个模块 每个模块有自己的帧率 帧率设置为0的话就是不tick
	a.RegisterModule(NiceModuleId, &NiceModule{})

	// 网络设置
	a.SetNetConfig(&net.Config{
		SessionStartId: 0,
		ConnectLimit:   1000,
		Timeout:        30,
		Codec:          &net.PbCodec{},
		MsgHandler:     &MyMsgHandler{},
	})
	// 网络监听器 支持tcp/kcp/ws
	ln, err := net.NewListener("tcp", ":10086")
	if err != nil {
		panic(err)
	}
	// 添加网络监听器 可支持同时接收多个监听器消息
	a.GetNetManager().AddListener(ln)

	// rpc设置
	a.SetRpcConfig(&rpc.Config{
		ClusterName: "nice",
		Consul:      "0.0.0.0:8500",
		ServiceKind: nil, // 当前节点没有service 就不用设置
	})

	a.Start(func() bool { // 初始化app 入参为启动函数 在初始化所有组件后执行
		log.Logger.Info("all module started, server start")
		return true
	})
	a.StartUpdate() // 开始update 所有组件开始tick 主线程阻塞
	a.End(func() { // 主线程开始退出 所有组件销毁后执行入参函数
		log.Logger.Info("all module stopped, server stop")
	})
}

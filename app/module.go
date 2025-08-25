package app

type IModule interface {
	Name() string                          // 模块名称
	FPS() uint                             // 模块帧率 0的话就是不tick
	OnStart()                              // 模块启动
	OnUpdate()                             // 模块tick
	OnDestroy()                            // 模块销毁
	OnMsg(msg interface{})                 // 模块收到消息 不用返回结果
	OnRequest(msg interface{}) interface{} // 模块收到请求 需要返回结果
}

package app

type ModuleID any

type IModule interface {
	FPS() uint                             // 模块帧率 0的话就是不tick
	Start()                                // 模块启动
	Update()                               // 模块tick
	OnDestroy()                            // 模块销毁
	OnMsg(msg interface{})                 // 模块收到消息 不用返回结果
	OnRequest(msg interface{}) interface{} // 模块收到请求 需要返回结果
}

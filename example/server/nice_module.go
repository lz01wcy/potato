package main

import "github.com/murang/potato/log"

type NiceModule struct {
}

func (n NiceModule) FPS() uint {
	return 1
}

func (n NiceModule) OnStart() {
	log.Sugar.Info("start")
}

func (n NiceModule) OnUpdate() {
}

func (n NiceModule) OnDestroy() {
	log.Sugar.Info("destroy")
}

func (n NiceModule) OnMsg(msg interface{}) {
	log.Sugar.Infof("msg: %v", msg)
}

func (n NiceModule) OnRequest(msg interface{}) interface{} {
	log.Sugar.Infof("request: %v", msg)
	return "Nice ~ " + msg.(string)
}

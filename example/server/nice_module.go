package main

import (
	"fmt"
	"github.com/murang/potato/app"
	"github.com/murang/potato/example/nicepb/nice"
	"github.com/murang/potato/log"
)

type NiceModule struct {
}

func (n *NiceModule) FPS() uint {
	return 1
}

func (n *NiceModule) OnStart() {
	log.Sugar.Info("start")
}

func (n *NiceModule) OnUpdate() {
}

func (n *NiceModule) OnDestroy() {
	log.Sugar.Info("destroy")
}

func (n *NiceModule) OnMsg(msg interface{}) {
	log.Sugar.Infof("msg: %v", msg)
}

func (n *NiceModule) OnRequest(msg interface{}) interface{} {
	log.Sugar.Infof("request: %v", msg)
	grain := nice.GetCalculatorGrainClient(app.Instance().GetCluster(), "NiceIdentity")
	sum, err := grain.Sum(&nice.Input{A: 6, B: 6})
	if err != nil {
		log.Sugar.Errorf("sum error: %v", err)
		return "sum error: " + err.Error()
	}
	return fmt.Sprintf("Nice ~  %s cal sum : %d", msg, sum.Result)
}

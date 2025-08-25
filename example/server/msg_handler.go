package main

import (
	"github.com/murang/potato/log"
	"github.com/murang/potato/net"
	"google.golang.org/protobuf/proto"
	"reflect"
)

type Agent struct {
	session *net.Session
}

type MyMsgHandler struct {
}

func (m MyMsgHandler) OnSessionOpen(session *net.Session) {
	log.Sugar.Info("handler got open:", session.ID())
}

func (m MyMsgHandler) OnSessionClose(session *net.Session) {
	log.Sugar.Info("handler got close:", session.ID())
}

func (m MyMsgHandler) OnMsg(session *net.Session, msg any) {
	log.Sugar.Infof("handler got msg: %v", msg)
	handler, ok := msgDispatcher[reflect.TypeOf(msg)]
	if !ok {
		log.Sugar.Errorf("handler got unknown msg: %v", msg)
		return
	}
	// 可以自定义和管理agent
	agent := &Agent{
		session: session,
	}
	handler(agent, msg.(proto.Message))
}

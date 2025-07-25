package main

import (
	"github.com/murang/potato/example/nicepb/nice"
	"github.com/murang/potato/log"
	"github.com/murang/potato/net"
	"github.com/murang/potato/pb"
	"google.golang.org/protobuf/proto"
	"reflect"
)

type Agent struct {
	session *net.Session
}

type MsgHandler struct {
}

func (m MsgHandler) OnSessionOpen(session *net.Session) {
	log.Sugar.Info("handler got open:", session.ID())
}

func (m MsgHandler) OnSessionClose(session *net.Session) {
	log.Sugar.Info("handler got close:", session.ID())
}

func (m MsgHandler) OnMsg(session *net.Session, msg any) {
	log.Sugar.Infof("handler got msg: %v", msg)
	handler, ok := msgDispatcher[nice.MsgId(pb.GetIdByType(reflect.TypeOf(msg).Elem()))]
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

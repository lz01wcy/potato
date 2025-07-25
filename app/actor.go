package app

import (
	"github.com/asynkron/protoactor-go/actor"
)

type ModuleUpdate struct{}
type ModuleOnMsg struct {
	Msg interface{}
}
type ModuleOnRequest struct {
	Request interface{}
}

type moduleActor struct {
	module IModule
}

func (m *moduleActor) Receive(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *actor.Started:
		m.module.OnStart()
	case *ModuleUpdate:
		m.module.OnUpdate()
	case *actor.Stopping:
		m.module.OnDestroy()
	case *ModuleOnMsg:
		m.module.OnMsg(msg.Msg)
	case *ModuleOnRequest:
		ctx.Respond(m.module.OnRequest(msg.Request))
	}
}

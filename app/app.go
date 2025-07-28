package app

import (
	"errors"
	"github.com/asynkron/protoactor-go/actor"
	"github.com/asynkron/protoactor-go/cluster"
	"github.com/asynkron/protoactor-go/scheduler"
	"github.com/murang/potato/log"
	"github.com/murang/potato/net"
	"github.com/murang/potato/rpc"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"sync"
	"syscall"
	"time"
)

var (
	instance               *Application
	once                   sync.Once
	errModuleNotRegistered = errors.New("module has not been registered")
)

type Application struct {
	exit        bool
	id2mod      map[ModuleID]IModule // ModuleID -> IModule
	id2pid      sync.Map             // ModuleID -> actor PID
	cancels     []scheduler.CancelFunc
	actorSystem *actor.ActorSystem
	netManager  *net.Manager
	cluster     *cluster.Cluster
	rpcManager  *rpc.Manager
}

func Instance() *Application {
	once.Do(func() {
		instance = NewApplication()
	})
	return instance
}

func NewApplication() *Application {
	a := &Application{
		actorSystem: actor.NewActorSystem(actor.WithLoggerFactory(log.ColoredConsoleLogging)),
		id2mod:      map[ModuleID]IModule{},
		id2pid:      sync.Map{},
	}
	return a
}

func (a *Application) GetActorSystem() *actor.ActorSystem {
	return a.actorSystem
}
func (a *Application) SetCluster(cls *cluster.Cluster) {
	a.cluster = cls
}
func (a *Application) GetCluster() *cluster.Cluster {
	return a.cluster
}
func (a *Application) GetNetManager() *net.Manager {
	return a.netManager
}
func (a *Application) GetRpcManager() *rpc.Manager {
	return a.rpcManager
}

func (a *Application) SetNetConfig(config *net.Config) {
	a.netManager = net.NewManagerWithConfig(config)
}

func (a *Application) SetRpcConfig(config *rpc.Config) {
	a.rpcManager = rpc.NewManagerWithConfig(config)
}

func (a *Application) RegisterModule(modId ModuleID, mod IModule) {
	if _, ok := a.id2mod[modId]; ok {
		panic("RegisterModule err, repeated module id: " + reflect.TypeOf(modId).Name())
	}
	a.id2mod[modId] = mod
	log.Logger.Info("module register : " + reflect.TypeOf(mod).Name())
}

func (a *Application) SendToModule(modId ModuleID, msg interface{}) {
	if pid, ok := a.id2pid.Load(modId); ok {
		a.actorSystem.Root.Send(pid.(*actor.PID), &ModuleOnMsg{Msg: msg})
	} else {
		log.Sugar.Warnf("module %s has not been registered", reflect.TypeOf(modId).Name())
	}
}

func (a *Application) RequestToModule(modId ModuleID, msg interface{}) (interface{}, error) {
	if pid, ok := a.id2pid.Load(modId); ok {
		return a.actorSystem.Root.RequestFuture(pid.(*actor.PID), &ModuleOnRequest{Request: msg}, time.Second).Result()
	} else {
		log.Sugar.Warnf("module %s has not been registered", reflect.TypeOf(modId).Name())
		return nil, errModuleNotRegistered
	}
}

func (a *Application) Start(f func() bool) bool {
	// catch signal
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGILL, syscall.SIGTRAP, syscall.SIGABRT)
		log.Sugar.Infof("caught signal: %v", <-c)
		a.exit = true
		time.Sleep(1 * time.Minute)
		var buf [65536]byte
		n := runtime.Stack(buf[:], true)
		log.Sugar.Errorf("server not stopped in 1 minute, all stack is:\n%s", string(buf[:n]))
		time.Sleep(2 * time.Second)
		os.Exit(1)
	}()

	// 网络
	if a.netManager != nil {
		a.netManager.Start()
	}
	// rpc
	if a.rpcManager != nil {
		a.cluster = a.rpcManager.Start(a.actorSystem)
	}

	for mid, mod := range a.id2mod {
		props := actor.PropsFromProducer(func() actor.Actor {
			return &moduleActor{module: mod.(IModule)}
		})
		pid := a.actorSystem.Root.Spawn(props)
		a.id2pid.Store(mid, pid)
		log.Logger.Info("static module init: " + reflect.TypeOf(mod).String())
	}

	if f != nil {
		ret := f()
		if !ret {
			_ = log.Logger.Sync()
		}
		return ret
	}

	return true
}

func (a *Application) StartUpdate() {
	sch := scheduler.NewTimerScheduler(a.actorSystem.Root)
	for mid, mod := range a.id2mod {
		if mod.FPS() > 0 {
			interval := time.Duration(1000/mod.FPS()) * time.Millisecond
			pid, _ := a.id2pid.Load(mid)
			a.cancels = append(a.cancels, sch.SendRepeatedly(interval, interval, pid.(*actor.PID), &ModuleUpdate{}))
		}
	}
	for {
		if a.exit {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (a *Application) End(f func()) {
	// 网络
	if a.netManager != nil {
		a.netManager.OnDestroy()
	}
	// rpc
	if a.rpcManager != nil {
		a.rpcManager.OnDestroy()
	}

	for _, cancel := range a.cancels {
		cancel()
	}
	a.id2pid.Range(func(key, value interface{}) bool {
		a.actorSystem.Root.Stop(value.(*actor.PID))
		return true
	})
	if f != nil {
		f()
	}
	_ = log.Logger.Sync()
	time.Sleep(1 * time.Second)
}

func (a *Application) Exit() {
	a.exit = true
}

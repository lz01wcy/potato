package net

import (
	"github.com/murang/potato/log"
	"net"
	"sync"
	"sync/atomic"
)

type ManagerConfig struct {
	SessionStartId uint64
	ConnectLimit   int32
	Timeout        int32
	Codec          ICodec
}

func defaultManagerConfig() *ManagerConfig {
	return &ManagerConfig{
		ConnectLimit: 50000,
		Timeout:      30,
		Codec:        &JsonCodec{},
	}
}

type ManagerConfigOption func(config *ManagerConfig)

func ManagerConfigure(options ...ManagerConfigOption) *ManagerConfig {
	config := defaultManagerConfig()
	for _, option := range options {
		option(config)
	}
	return config
}

func WithCodec(codec ICodec) ManagerConfigOption {
	return func(config *ManagerConfig) {
		config.Codec = codec
	}
}

func WithConnectLimit(limit int32) ManagerConfigOption {
	return func(config *ManagerConfig) {
		config.ConnectLimit = limit
	}
}

func WithTimeout(timeout int32) ManagerConfigOption {
	return func(config *ManagerConfig) {
		config.Timeout = timeout
	}
}

type Manager struct {
	idGen            uint64
	sessionMap       sync.Map
	sessionCount     int32
	listeners        []IListener
	codec            ICodec
	connectLimit     int32
	timeout          int32
	sessionEventChan chan *SessionEvent
	msgHandler       IMsgHandler
}

func NewManager(options ...ManagerConfigOption) *Manager {
	config := ManagerConfigure(options...)
	return NewManagerWithConfig(config)
}

func NewManagerWithConfig(config *ManagerConfig) *Manager {
	m := &Manager{
		sessionMap:       sync.Map{},
		listeners:        make([]IListener, 0),
		sessionEventChan: make(chan *SessionEvent, 1024),
	}
	m.idGen = config.SessionStartId
	m.codec = config.Codec
	m.connectLimit = config.ConnectLimit
	m.timeout = config.Timeout
	return m
}

func (sm *Manager) OnNewConnection(conn net.Conn) {
	if sm.connectLimit > 0 {
		if atomic.LoadInt32(&sm.sessionCount) >= sm.connectLimit {
			log.Sugar.Warnf("connect limit: %d", sm.connectLimit)
			_ = conn.Close()
			return
		}
	}
	sess := sm.NewSession(conn)
	sess.Start()
}

func (sm *Manager) AddListener(ln IListener) {
	ln.OnNewConnection(sm.OnNewConnection)
	sm.listeners = append(sm.listeners, ln)
}

func (sm *Manager) SetMsgHandler(handler IMsgHandler) {
	sm.msgHandler = handler
}

func (sm *Manager) NewSession(conn net.Conn) *Session {
	atomic.AddUint64(&sm.idGen, 1)
	s := &Session{
		manager:     sm,
		id:          atomic.LoadUint64(&sm.idGen),
		conn:        conn,
		connGuard:   sync.RWMutex{},
		exitSync:    sync.WaitGroup{},
		sendChan:    make(chan any, 32),
		sendRawChan: make(chan []byte, 32),
	}
	return s
}

func (sm *Manager) Start() {
	for _, ln := range sm.listeners {
		ln.Start()
	}
	go func() {
		for {
			select {
			case ses := <-sm.sessionEventChan:
				switch ses.Type {
				case SessionOpen:
					sm.sessionMap.Store(ses.Session.ID(), ses.Session)
					atomic.AddInt32(&sm.sessionCount, 1)
					log.Sugar.Infof("session open: %d", ses.Session.ID())
					if sm.msgHandler != nil {
						sm.msgHandler.OnSessionOpen(ses.Session)
					}
				case SessionClose:
					sm.sessionMap.Delete(ses.Session.ID())
					atomic.AddInt32(&sm.sessionCount, -1)
					log.Sugar.Infof("session close: %d", ses.Session.ID())
					if sm.msgHandler != nil {
						sm.msgHandler.OnSessionClose(ses.Session)
					}
				case SessionMsg:
					if sm.msgHandler != nil {
						sm.msgHandler.OnMsg(ses.Session, ses.Msg)
					}
				}
			}
		}
	}()
}

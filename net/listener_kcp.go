package net

import (
	"github.com/murang/potato/log"
	"github.com/xtaci/kcp-go"
	"net"
	"time"
)

// server
type kcpListener struct {
	addr            string
	listener        net.Listener
	exit            bool
	onNewConnection func(net.Conn)
}

func newKcpListener(addr string) (*kcpListener, error) {
	l, err := kcp.Listen(addr)
	if err != nil {
		log.Sugar.Errorf("listen error on %s, because: %v", addr, err)
		return nil, err
	}
	log.Sugar.Infof("kcp listen on %s", addr)
	s := &kcpListener{
		addr:     addr,
		listener: l,
	}
	return s, nil
}

func (s *kcpListener) Start() {
	go s.accept()
	log.Sugar.Infof("kcp listener start: %+v", s.addr)
}

func (s *kcpListener) Stop() {
	s.exit = true
	err := s.listener.Close()
	if err != nil {
		log.Sugar.Errorf("close kcp listener error: %v", err)
		return
	}
}

func (s *kcpListener) OnNewConnection(f func(net.Conn)) {
	s.onNewConnection = f
}

func (s *kcpListener) accept() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
				time.Sleep(time.Millisecond)
				continue
			}
			if s.exit {
				break
			}
			// 调试状态时, 才打出accept的具体错误
			log.Sugar.Errorf("kcp.accept failed: %v", err.Error())
			break
		} else {
			if s.exit {
				break
			}
			if s.onNewConnection == nil {
				_ = conn.Close()
				continue
			}
			go s.onNewConnection(conn)
		}
	}
}

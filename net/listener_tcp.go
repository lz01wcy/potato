package net

import (
	"github.com/murang/potato/log"
	"net"
	"time"
)

// server
type tcpListener struct {
	addr            string
	listener        net.Listener
	exit            bool
	onNewConnection func(net.Conn)
}

func newTcpListener(addr string) (*tcpListener, error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Sugar.Errorf("listen error on %s, because: %v", addr, err)
		return nil, err
	}
	log.Sugar.Infof("tcp listen on %s", addr)
	s := &tcpListener{
		addr:     addr,
		listener: l,
	}
	return s, nil
}

func (s *tcpListener) Start() {
	go s.accept()
	log.Sugar.Infof("tcp listener start: %+v", s.addr)
}

func (s *tcpListener) Stop() {
	s.exit = true
	err := s.listener.Close()
	if err != nil {
		log.Sugar.Errorf("close tcp listener error: %v", err)
		return
	}
}

func (s *tcpListener) OnNewConnection(f func(net.Conn)) {
	s.onNewConnection = f
}

func (s *tcpListener) accept() {
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
			log.Sugar.Errorf("tcp.accept failed: %v", err.Error())
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

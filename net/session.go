package net

import (
	"encoding/binary"
	"errors"
	"github.com/murang/potato/log"
	"io"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type SessionEventType int32

const (
	SessionOpen SessionEventType = iota
	SessionClose
	SessionMsg
)

type Session struct {
	manager   *Manager
	id        uint64
	conn      net.Conn
	connGuard sync.RWMutex
	exitSync  sync.WaitGroup
	sendChan  chan []byte
	state     int64 //正常情况是0 主动关闭是1 出错关闭是2
}

type SessionEvent struct {
	Session *Session
	Type    SessionEventType
	Msg     interface{}
}

func (s *Session) setConn(conn net.Conn) {
	s.connGuard.Lock()
	s.conn = conn
	s.connGuard.Unlock()
}

func (s *Session) Conn() net.Conn {
	s.connGuard.RLock()
	defer s.connGuard.RUnlock()
	return s.conn
}

func (s *Session) ID() uint64 {
	return s.id
}

func (s *Session) Raw() interface{} {
	return s.Conn()
}

func (s *Session) Close() {
	state := atomic.LoadInt64(&s.state)
	if state != 0 {
		return
	}
	atomic.StoreInt64(&s.state, 2)
	conn := s.Conn()
	if conn != nil {
		conn.Close()
		conn.SetDeadline(time.Now())
		conn = nil
	}
}

func (s *Session) Send(msg interface{}) {

	// 只能通过Close关闭连接
	if msg == nil {
		return
	}

	// 已经关闭，不再发送
	if s.IsClosed() {
		return
	}

	data, err := s.manager.codec.Encode(msg)
	if err != nil {
		log.Sugar.Errorf("session encode err, sesid: %d, err: %s", s.ID(), err)
		return
	}

	s.sendChan <- data
}

func (s *Session) SendRaw(data []byte) (err error) {
	pkt := make([]byte, lenSize+len(data))

	// Length
	binary.LittleEndian.PutUint32(pkt, uint32(len(data)))

	// Value
	copy(pkt[lenSize:], data)

	writer, ok := s.Raw().(io.Writer)
	if !ok || writer == nil {
		return nil
	}
	if _, err = writer.Write(pkt); err != nil {
		return
	}
	err = s.updateDeadline()
	return
}

func (s *Session) IsClosed() bool {
	return atomic.LoadInt64(&s.state) != 0
}

func (s *Session) Start() {

	atomic.StoreInt64(&s.state, 0)

	// 需要接收和发送线程同时完成时才算真正的完成
	s.exitSync.Add(2)
	go func() {
		// 等待2个任务结束
		s.exitSync.Wait()
		s.Close()
		s.manager.sessionEventChan <- &SessionEvent{
			Session: s,
			Type:    SessionClose,
		}
	}()

	// 先处理这个 防止event handler不在同一个进程里面 出现的并发问题
	s.manager.sessionEventChan <- &SessionEvent{
		Session: s,
		Type:    SessionOpen,
	}

	// 启动并发接收goroutine
	go s.readLoop()

	// 启动并发发送goroutine
	go s.writeLoop()
}

// 接收循环
func (s *Session) readLoop() {

	for !s.IsClosed() {

		var msgBytes []byte
		var err error

		msgBytes, err = s.readMessageBytes()

		if err != nil {
			var ip string
			if s.conn != nil {
				addr := s.conn.RemoteAddr()
				if addr != nil {
					ip = addr.String()
				}
			}
			if atomic.LoadInt64(&s.state) != 1 || (err.Error() != io.ErrClosedPipe.Error() && !strings.Contains(err.Error(), "use of closed network connection")) {
				log.Sugar.Warnf("session read err, sesid: %d, err: %s ip: %s", s.ID(), err, ip)
			}
			s.sendChan <- nil //给写队列传空 用于关闭写队列
			break
		}

		ev := &SessionEvent{
			Session: s,
			Type:    SessionMsg,
		}
		ev.Msg, err = s.manager.codec.Decode(msgBytes)
		if err != nil {
			log.Sugar.Errorf("decode msg error, sesid: %d, err: %s", s.ID(), err)
			s.sendChan <- nil //给写队列传空 用于关闭写队列
			break
		}
		s.manager.sessionEventChan <- ev
	}

	// 通知完成
	s.exitSync.Done()
}

func (s *Session) readMessageBytes() (msg []byte, err error) {
	if s.manager.timeout != 0 {
		if err = s.conn.SetReadDeadline(time.Now().Add(time.Duration(s.manager.timeout) * time.Second)); err != nil {
			return
		}
	}

	reader, ok := s.Raw().(io.Reader)

	// 转换错误，或者连接已经关闭时退出
	if !ok || reader == nil {
		return nil, errors.New("reader cast error")
	}

	msg, err = ReadPacket(reader)

	if err != nil {
		return
	}

	return
}

// 发送循环
func (s *Session) writeLoop() {
	for !s.IsClosed() {
		msg, ok := <-s.sendChan
		if !ok {
			break
		}
		if msg == nil { //在读loop的时候出错 这边需要break关闭
			break
		}
		if err := s.sendMessage(msg); err != nil {
			if atomic.LoadInt64(&s.state) != 1 || (err.Error() != io.ErrClosedPipe.Error() && !strings.Contains(err.Error(), "use of closed network connection")) {
				log.Sugar.Warnf("session sendLoop sendMessage err: sesid: %d, err: %s", s.ID(), err.Error())
			}
			break
		}
	}

	// 完整关闭
	conn := s.Conn()
	if conn != nil {
		_ = conn.Close()
	}

	// 通知完成
	s.exitSync.Done()
}

func (s *Session) sendMessage(msg interface{}) (err error) {
	if s.manager.timeout != 0 {
		if err = s.conn.SetWriteDeadline(time.Now().Add(time.Duration(s.manager.timeout) * time.Second)); err != nil {
			return
		}
	}

	writer, ok := s.Raw().(io.Writer)

	// 转换错误，或者连接已经关闭时退出
	if !ok || writer == nil {
		return nil
	}

	pkg, err := s.manager.codec.Encode(msg)

	err = WritePacket(writer, pkg)
	if err != nil {
		return
	}
	return
}

func (s *Session) updateDeadline() (err error) {
	if s.manager.timeout == 0 {
		err = s.Conn().SetDeadline(time.Now().Add(time.Second * 30))
	} else {
		err = s.Conn().SetDeadline(time.Now().Add(time.Second * time.Duration(s.manager.timeout)))
	}
	if err != nil {
		log.Logger.Error("session flush deadline err")
	}
	return
}

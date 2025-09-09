package net

type IMsgHandler interface {
	IsMsgInRoutine() bool // 如果设置消息在携程中处理 消息将不会经过channel 而是直接由handler处理 需要注意并发
	OnSessionOpen(session *Session)
	OnSessionClose(session *Session)
	OnMsg(session *Session, msg any)
}

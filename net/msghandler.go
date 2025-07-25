package net

type IMsgHandler interface {
	OnSessionOpen(session *Session)
	OnSessionClose(session *Session)
	OnMsg(session *Session, msg any)
}

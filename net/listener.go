package net

import (
	"errors"
	"net"
)

type IListener interface {
	Start()
	Stop()
	OnNewConnection(func(net.Conn))
}

func NewListener(network, addr string) (IListener, error) {
	switch network {
	case "tcp":
		return newTcpListener(addr)
	case "kcp":
		return newKcpListener(addr)
	case "ws":
		return newWsListener(addr)
	}
	return nil, errors.New("not support network")
}

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
	}
	return nil, errors.New("not support network")
}

package tcp

import (
	"context"
	"go-redis/interface/tcp"
	"go-redis/lib/logger"
	"net"
)

type Config struct {
	Address string
}

func ListenAndServeWithSignal(cfg *Config, handler tcp.Handler) error {
	listener, err := net.Listen("tcp", cfg.Address)
	closeChan := make(chan struct{})
	if err != nil {
		return err
	}
	logger.Info("start listen")
	ListenAndServe(listener, handler, closeChan)
	return nil
}

func ListenAndServe(listener net.Listener, handler tcp.Handler, closeChan <-chan struct{}) {
	ctx := context.Background()
	for {
		conn, err := listener.Accept()
		if err != nil {
			break
		}
		logger.Info("accepted link")
		go func() {
			handler.Handle(ctx, conn)
		}()
	}
}

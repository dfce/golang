package bootstrap

import (
	"fmt"
	"mini-srs/internal/config"
	"mini-srs/internal/server"
	"net"
)

type App struct {
	cfg *config.Config
}

func NewApp(cfg *config.Config) *App {
	return &App{cfg: cfg}
}

func (a *App) Start() error {
	addr := fmt.Sprintf("%s:%d", a.cfg.ListenAddr, a.cfg.RtmpPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer listener.Close()
	fmt.Printf("[Mini-SRS] RTMP 转发引擎已开启，监听端口%s\n", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}

		session := server.NewSession(conn)
		go session.ServeLoop()
	}
}

package internal

import (
	"github.com/tryoo0607/coldbrew-scheduler/internal/config"
	"github.com/tryoo0607/coldbrew-scheduler/internal/server"
)

type cfg = config.DefaultConfig

func Run(cfg cfg) {

	server.Run(cfg.Server.Port)
}

package internal

import (
	"github.com/tryoo0607/coldbrew-scheduler/internal/config"
	"github.com/tryoo0607/coldbrew-scheduler/internal/pkg/log"
	"github.com/tryoo0607/coldbrew-scheduler/internal/pkg/server"
)

type cfg = config.DefaultConfig

func Run(cfg cfg, logger log.Logger) {

	server.Run(cfg.Server.Port)
}

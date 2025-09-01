package main

import (
	"fmt"

	project "github.com/tryoo0607/coldbrew-scheduler/cmd/internal"
	"github.com/tryoo0607/coldbrew-scheduler/internal/config"
	"github.com/tryoo0607/coldbrew-scheduler/internal/pkg/log"
)

// TODO. [TR-YOO] 초기 구성잡기
func main() {
	fmt.Println("init")

	// Load Config
	cfg, err := config.Load()

	if err != nil {

	}

	// Set Logger
	logger := log.NewZapLogger()
	defer logger.Sync()

	// Run
	project.Run(cfg, logger)
}

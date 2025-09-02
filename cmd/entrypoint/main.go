package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/tryoo0607/coldbrew-scheduler/internal/app/project"
	"github.com/tryoo0607/coldbrew-scheduler/internal/config"
	"github.com/tryoo0607/coldbrew-scheduler/internal/pkg/log"
)

func main() {
	fmt.Println("init")

	// Load Config
	cfg, err := config.Load()

	// TODO. [TR-YOO] config 사용하게되면 지우기
	_ = cfg

	if err != nil {

	}

	// Set Logger
	logger := log.NewZapLogger()
	defer logger.Sync()

	// Run
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	project.Run(ctx, project.ProjectOptions{
		Kubeconfig: "",
		UseFake:    true,
		InCluster:  false,
	})
}

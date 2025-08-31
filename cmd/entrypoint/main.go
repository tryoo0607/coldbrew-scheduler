package main

import (
	"fmt"

	"github.com/tryoo0607/coldbrew-scheduler/internal/config"
	project "github.com/tryoo0607/coldbrew-scheduler/cmd/internal"
)

// TODO. [TR-YOO] 초기 구성잡기
func main() {
	fmt.Println("init")
	cfg, err := config.Load()

	if err != nil {

	}

	project.Run(cfg)
}

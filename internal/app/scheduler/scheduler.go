package scheduler

import (
	"context"

	"github.com/tryoo0607/coldbrew-scheduler/internal/pkg/clientgo"
	"github.com/tryoo0607/coldbrew-scheduler/internal/pkg/clientgo/api"
)

func Run(ctx context.Context, client clientgo.Client, find api.FinderFunc) error {

	controller, err := client.NewPodController(ctx, find)

    if err != nil {
        return err
    }

	stop := make(chan struct{})

	// 상위의 context에서 종료 신호를 보내면 종료하도록
	go func() {
		<-ctx.Done()
		close(stop)
	}()

	// 위의 close 되기 전까지 Block됨
	controller.Run(stop)

	return nil
}

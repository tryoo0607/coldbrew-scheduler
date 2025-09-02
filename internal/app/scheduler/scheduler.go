package scheduler

import (
	"context"

	"github.com/tryoo0607/coldbrew-scheduler/internal/pkg/clientgo/informer"
	"k8s.io/client-go/kubernetes"
)

func Run(ctx context.Context, clientset kubernetes.Interface, find informer.FinderFunc) error {

	controller := informer.NewPodInformer(ctx, clientset, find)

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

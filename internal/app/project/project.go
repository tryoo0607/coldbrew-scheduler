package project

import (
	"context"
	"fmt"
	"net/http"

	"github.com/tryoo0607/coldbrew-scheduler/internal/app/finder"
	"github.com/tryoo0607/coldbrew-scheduler/internal/app/scheduler"
	clientk8s "github.com/tryoo0607/coldbrew-scheduler/internal/pkg/clientgo/k8s"
	"golang.org/x/sync/errgroup"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type ProjectOptions struct {
	Kubeconfig string
	InCluster  bool
	UseFake    bool
}

func Run(ctx context.Context, options ProjectOptions) error {

	clientset, err := buildClientset(options)
	if err != nil {
		return fmt.Errorf("k8s: %w", err)
	}

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {

		return scheduler.Run(ctx, clientset, finder.FindBestNode)
	})

	if err := g.Wait(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

func buildClientset(opt ProjectOptions) (kubernetes.Interface, error) {
	switch {
	case opt.UseFake:
		return clientk8s.NewFakeClient(), nil
	case opt.InCluster:
		cfg, err := rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
		return kubernetes.NewForConfig(cfg)
	default:
		cfg, err := clientcmd.BuildConfigFromFlags("", clientk8s.ResolveKubeconfigPath(opt.Kubeconfig))
		if err != nil {
			return nil, err
		}
		return kubernetes.NewForConfig(cfg)
	}
}

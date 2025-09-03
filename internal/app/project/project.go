package project

import (
	"context"
	"fmt"
	"net/http"

	"github.com/tryoo0607/coldbrew-scheduler/internal/app/finder"
	"github.com/tryoo0607/coldbrew-scheduler/internal/app/scheduler"
	clientk8s "github.com/tryoo0607/coldbrew-scheduler/internal/pkg/clientgo/k8s"
	"golang.org/x/sync/errgroup"
)

type ProjectOptions struct {
	Kubeconfig string
	InCluster  bool
	UseFake    bool
}

func Run(ctx context.Context, options ProjectOptions) error {

	convertedOpts := toOptions(options)

	clientset, err := clientk8s.BuildClientset(convertedOpts)
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

func toOptions(opts ProjectOptions) clientk8s.Options {

	return clientk8s.Options{
		Kubeconfig: opts.Kubeconfig,
		UseFake:    opts.UseFake,
		InCluster:  opts.InCluster,
	}
}

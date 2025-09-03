package project

import (
	"context"
	"fmt"
	"net/http"

	"github.com/tryoo0607/coldbrew-scheduler/internal/app/finder"
	"github.com/tryoo0607/coldbrew-scheduler/internal/app/scheduler"
	"github.com/tryoo0607/coldbrew-scheduler/internal/pkg/clientgo"
	"golang.org/x/sync/errgroup"
)

type ProjectOptions struct {
	Kubeconfig string
	InCluster  bool
	UseFake    bool
}

func Run(ctx context.Context, options ProjectOptions) error {

	convertedOpts := toOptions(options)

	client, err := clientgo.New(convertedOpts)
	if err != nil {
		return fmt.Errorf("k8s: %w", err)
	}

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {

		return scheduler.Run(ctx, client, finder.FindBestNode)
	})

	if err := g.Wait(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

func toOptions(opts ProjectOptions) clientgo.Options {

	return clientgo.Options{
		Kubeconfig: opts.Kubeconfig,
		UseFake:    opts.UseFake,
		InCluster:  opts.InCluster,
	}
}

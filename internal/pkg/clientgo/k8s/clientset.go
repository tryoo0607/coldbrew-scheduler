package k8s

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Options struct {
	Kubeconfig string
	InCluster  bool
	UseFake    bool
}

func BuildClientset(opt Options) (kubernetes.Interface, error) {

	switch {
	case opt.UseFake:
		return newFakeClientset(), nil
	case opt.InCluster:
		return newInClusterClientset(opt)
	default:
		path := ResolveKubeconfigPath(opt.Kubeconfig)
		return newKubeconfigClientset(opt, path)
	}
}

func newFakeClientset() kubernetes.Interface {

	return fake.NewClientset()
}

func newInClusterClientset(opt Options) (kubernetes.Interface, error) {

	cfg, err := rest.InClusterConfig()

	if err != nil {

		return nil, err
	}

	return newForConfig(cfg, opt)
}

func newKubeconfigClientset(opt Options, path string) (kubernetes.Interface, error) {

	cfg, err := clientcmd.BuildConfigFromFlags("", path)

	if err != nil {

		return nil, fmt.Errorf("kubeconfig %q: %w", path, err)
	}

	return newForConfig(cfg, opt)
}

func newForConfig(cfg *rest.Config, opt Options) (kubernetes.Interface, error) {

	_ = opt

	return kubernetes.NewForConfig(cfg)
}

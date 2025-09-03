package k8s

import (
	"k8s.io/client-go/kubernetes"
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
		return NewFakeClient(), nil
	case opt.InCluster:
		cfg, err := rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
		return kubernetes.NewForConfig(cfg)
	default:
		cfg, err := clientcmd.BuildConfigFromFlags("", ResolveKubeconfigPath(opt.Kubeconfig))
		if err != nil {
			return nil, err
		}
		return kubernetes.NewForConfig(cfg)
	}
}

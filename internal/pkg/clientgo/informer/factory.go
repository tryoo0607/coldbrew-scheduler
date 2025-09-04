package informer

import (
	"time"

	"github.com/tryoo0607/coldbrew-scheduler/internal/pkg/clientgo/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
)

func NewInformerFactory(clientset kubernetes.Interface) informers.SharedInformerFactory {
	return informers.NewSharedInformerFactoryWithOptions(
		clientset,
		time.Second*0, // no resync
		informers.WithTweakListOptions(func(opts *metav1.ListOptions) {
			opts.FieldSelector = fields.AndSelectors(
				fields.OneTermEqualSelector(api.SpecSchedulerName, api.ColdBrewScheduler),
				fields.OneTermEqualSelector(api.SpecNodeName, ""),
			).String()
		}),
	)
}

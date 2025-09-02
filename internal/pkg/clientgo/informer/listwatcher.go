package informer

import (
	"context"

	"github.com/tryoo0607/coldbrew-scheduler/internal/pkg/clientgo/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
)

func newListWatcher(clientset kubernetes.Interface) cache.ListerWatcher {

	selector := fields.AndSelectors(
		fields.OneTermEqualSelector(api.SpecSchedulerName, api.ColdBrewScheduler),
		fields.OneTermEqualSelector(api.SpecNodeName, ""),
	)

	if IsFakeClient(clientset) {

		return newFakeListWatcher(clientset, selector)
	}

	// 특정 대상에 대한 List를 watch하는데 사용
	listWatcher := cache.NewListWatchFromClient(
		clientset.CoreV1().RESTClient(),
		// 대상 리소스
		api.ResourcePods,
		// 대상 Namespace
		metav1.NamespaceAll,
		// 필드 셀렉터 => 해당 필드와 값을 비교해여 Equal인 것들만 필터링
		selector,
	)

	return listWatcher
}

func IsFakeClient(clientset kubernetes.Interface) bool {
	_, ok := clientset.(*fake.Clientset)
	return ok
}

// Fake Client에는 RestClient()가 없기 때문에 NewListWatchFromClient()를 사용할 수 없음
// 따라서 ListWatch() 사용
func newFakeListWatcher(cs kubernetes.Interface, selector fields.Selector) cache.ListerWatcher {
	return &cache.ListWatch{
		ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
			opts.FieldSelector = selector.String()
			return cs.CoreV1().Pods(metav1.NamespaceAll).List(context.TODO(), opts)
		},
		WatchFunc: func(opts metav1.ListOptions) (watch.Interface, error) {
			opts.FieldSelector = selector.String()
			return cs.CoreV1().Pods(metav1.NamespaceAll).Watch(context.TODO(), opts)
		},
	}
}

package informer

import (
	"github.com/tryoo0607/coldbrew-scheduler/internal/pkg/clientgo/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

func newListWatcher(clientset kubernetes.Interface, resource string) cache.ListerWatcher {

	selector := fields.AndSelectors(
		fields.OneTermEqualSelector(api.SpecSchedulerName, api.ColdBrewScheduler),
		fields.OneTermEqualSelector(api.SpecNodeName, ""),
	)

	// 특정 대상에 대한 List를 watch하는데 사용
	listWatcher := cache.NewListWatchFromClient(
		clientset.CoreV1().RESTClient(),
		// 대상 리소스
		resource,
		// 대상 Namespace
		metav1.NamespaceAll,
		// 필드 셀렉터 => 해당 필드와 값을 비교해여 Equal인 것들만 필터링
		selector,
	)

	return listWatcher
}

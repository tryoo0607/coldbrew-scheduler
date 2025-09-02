package binder

import (
	"context"
	"errors"

	"github.com/tryoo0607/coldbrew-scheduler/internal/pkg/clientgo/api"
	clientk8s "github.com/tryoo0607/coldbrew-scheduler/internal/pkg/clientgo/k8s"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type BindOptions struct {
	ClientSet kubernetes.Interface
	Ctx       context.Context
	Pod       *corev1.Pod
	NodeName  string
}

func BindPodToNode(opt BindOptions) error {
	if opt.Ctx == nil {
		opt.Ctx = context.Background()
	}

	if opt.ClientSet == nil || opt.Pod == nil || opt.NodeName == "" {

		return errors.New("invalid binder options")
	}

	// 1) 서브리소스 먼저 시도 (실클러스터 권장)
	if err := bindViaSubresource(opt); err == nil {

		// fakeClient에서는 Pods().Bind()가 성공해도 아무것도 Update하지 않음
		// 때문에 아래 로직 실행하도록 로직 추가
		if clientk8s.IsFakeClient(opt.ClientSet) {
			return bindBySpecNodeName(opt)
		}

		return nil
	}

	// 2) 폴백: spec.nodeName 업데이트 (fake에서도 동작)
	return bindBySpecNodeName(opt)
}

func bindViaSubresource(options BindOptions) error {
	pod := options.Pod

	binding := &corev1.Binding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pod.Name,
			Namespace: pod.Namespace,
			UID:       pod.UID,
		},
		Target: corev1.ObjectReference{
			Kind:       api.ResourceNode,
			APIVersion: api.V1,
			Name:       options.NodeName,
		},
	}

	return options.ClientSet.CoreV1().Pods(pod.Namespace).Bind(options.Ctx, binding, metav1.CreateOptions{})
}

func bindBySpecNodeName(options BindOptions) error {

	clientset := options.ClientSet
	nodeName := options.NodeName
	pod := options.Pod

	pod.Spec.NodeName = nodeName

	_, err := clientset.CoreV1().Pods(pod.Namespace).Update(context.Background(), pod, metav1.UpdateOptions{})

	return err
}

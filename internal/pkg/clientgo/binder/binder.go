package binder

import (
	"context"
	"errors"

	"github.com/tryoo0607/coldbrew-scheduler/internal/pkg/clientgo/api"
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

	// 1) 서브리소스 먼저 시도
	if err := bindViaSubresource(opt); err == nil {
		// fake에선 Bind가 no-op일 수 있으므로 실제로 묶였는지 확인
		if ok, _ := isBound(opt); ok {
			return nil
		}
		// 실클러스터/권한 문제 없이도 no-op일 수 있으니 폴백 계속 진행
	}

	// 2) 폴백: spec.nodeName 업데이트 (fake에서도 동작)
	return bindBySpecNodeName(opt)
}

func bindViaSubresource(opt BindOptions) error {
	pod := opt.Pod
	binding := &corev1.Binding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pod.Name,
			Namespace: pod.Namespace,
			UID:       pod.UID,
		},
		Target: corev1.ObjectReference{
			Kind:       api.ResourceNode,                   // "Node"
			APIVersion: corev1.SchemeGroupVersion.String(), // "v1"
			Name:       opt.NodeName,
		},
	}
	return opt.ClientSet.CoreV1().Pods(pod.Namespace).
		Bind(opt.Ctx, binding, metav1.CreateOptions{})
}

func bindBySpecNodeName(opt BindOptions) error {
	pod := opt.Pod
	pod.Spec.NodeName = opt.NodeName

	_, err := opt.ClientSet.CoreV1().Pods(pod.Namespace).
		Update(opt.Ctx, pod, metav1.UpdateOptions{})
	return err
}

func isBound(opt BindOptions) (bool, error) {
	p, err := opt.ClientSet.CoreV1().Pods(opt.Pod.Namespace).
		Get(opt.Ctx, opt.Pod.Name, metav1.GetOptions{})
	if err != nil {
		return false, err
	}
	return p.Spec.NodeName == opt.NodeName, nil
}

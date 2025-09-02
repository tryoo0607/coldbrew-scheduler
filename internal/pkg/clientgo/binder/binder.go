package binder

import (
	"context"
	"errors"

	"github.com/tryoo0607/coldbrew-scheduler/internal/pkg/clientgo/api"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type BindOptions struct {
	ClientSet kubernetes.Interface
	Ctx       context.Context
	Pod       *corev1.Pod
	NodeName  string
}

func BindPodToNode(options BindOptions) error {
	if options.Ctx == nil {
		options.Ctx = context.Background()
	}
	if options.ClientSet == nil || options.Pod == nil || options.NodeName == "" {
		return errors.New("invalid binder options")
	}

	if err := bindViaSubresource(options); err != nil {

		if apierrors.IsForbidden(err) || apierrors.IsMethodNotSupported(err) {
			return bindBySpecNodeName(options)
		}
		return err
	}
	return nil
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
			Kind:       api.ResourcePods,
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

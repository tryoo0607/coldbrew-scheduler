package k8s

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func ListNode(ctx context.Context, clientset kubernetes.Clientset) (*corev1.NodeList, error) {

	return clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
}

package k8s

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateNamespace(ctx context.Context, client *kubernetes.Clientset, namespace string) error {

	_, err := client.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "Unauthorized") {
			return fmt.Errorf("kubernetes auth expired: run 'gcloud auth login'")
		}
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("failed to get namespace: %w", err)
		}
	} else {
		fmt.Printf("Namespace %s already exists\n", namespace)
		return nil
	}

	namespaceObj := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
			Labels: map[string]string{
				"managed-by": "mini-porter",
			},
		},
	}

	_, err = client.CoreV1().Namespaces().Create(ctx, namespaceObj, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			fmt.Printf("Namespace %s already exists\n", namespace)
			return nil
		}
		return fmt.Errorf("failed to create namespace: %w", err)
	}

	fmt.Printf("Namespace %s created successfully\n", namespace)
	return nil
}

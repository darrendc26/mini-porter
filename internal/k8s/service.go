package k8s

import (
	"context"
	"fmt"

	"github.com/darrendc26/mini-porter/internal/config"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

func CreateService(client *kubernetes.Clientset, cfg *config.Config, serviceInfo ServiceInfo) error {
	namespace := cfg.Name
	servicesClient := client.CoreV1().Services(namespace)

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceInfo.Name,
			Labels: map[string]string{
				"app":        serviceInfo.Name,
				"managed-by": "mini-porter",
			},
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Selector: map[string]string{
				"app": serviceInfo.Name,
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Protocol:   corev1.ProtocolTCP,
					Port:       int32(serviceInfo.Port),
					TargetPort: intstr.FromInt(serviceInfo.Port),
				},
			},
		},
	}

	_, err := servicesClient.Create(context.TODO(), service, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			fmt.Println("[5/6] Service Exists, updating...")

			existing, err := servicesClient.Get(context.TODO(), serviceInfo.Name, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("failed to get existing service: %w", err)
			}

			service.ResourceVersion = existing.ResourceVersion
			service.Spec.ClusterIP = existing.Spec.ClusterIP

			_, err = servicesClient.Update(context.TODO(), service, metav1.UpdateOptions{})
			if err != nil {
				return fmt.Errorf("failed to update service: %w", err)
			}
		} else {
			return err
		}
	} else {
		fmt.Println("[5/6] Service Created")
	}

	return nil
}

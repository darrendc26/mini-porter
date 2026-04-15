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

func CreateService(client *kubernetes.Clientset, cfg *config.Config, serviceInfo ServiceInfo) (string, error) {
	fmt.Println("[4/5] Creating service...")

	servicesClient := client.CoreV1().Services("default")

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceInfo.Name,
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeNodePort,
			Selector: map[string]string{
				"app": serviceInfo.Name,
			},
			Ports: []corev1.ServicePort{
				{
					Protocol: corev1.ProtocolTCP,
					// TODO: Handle multiple services and ports
					Port:       int32(serviceInfo.Port),
					TargetPort: intstr.FromInt(serviceInfo.Port),
				},
			},
		},
	}

	svc, err := servicesClient.Create(context.TODO(), service, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			fmt.Println("Service exists, updating...")

			existing, err := servicesClient.Get(context.TODO(), serviceInfo.Name, metav1.GetOptions{})
			if err != nil {
				return "", fmt.Errorf("failed to get existing service: %w", err)
			}

			service.ResourceVersion = existing.ResourceVersion
			service.Spec.ClusterIP = existing.Spec.ClusterIP

			svc, err = servicesClient.Update(context.TODO(), service, metav1.UpdateOptions{})
			if err != nil {
				return "", fmt.Errorf("failed to update service: %w", err)
			}
		} else {
			return "", fmt.Errorf("failed to create service: %w", err)
		}
	} else {
		fmt.Println("Service created successfully")
	}

	svc, err = servicesClient.Get(context.TODO(), serviceInfo.Name, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to fetch service: %w", err)
	}

	if len(svc.Spec.Ports) == 0 {
		return "", fmt.Errorf("service has no ports")
	}

	nodePort := svc.Spec.Ports[0].NodePort

	nodes, err := client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get nodes: %w", err)
	}

	if len(nodes.Items) == 0 {
		return "", fmt.Errorf("no nodes found")
	}

	nodeIP := ""
	for _, addr := range nodes.Items[0].Status.Addresses {
		if addr.Type == corev1.NodeInternalIP {
			nodeIP = addr.Address
			break
		}
	}

	if nodeIP == "" {
		return "", fmt.Errorf("could not determine node IP")
	}

	url := fmt.Sprintf("%s is live: http://%s:%d\n", serviceInfo.Name, nodeIP, nodePort)

	return url, nil
}

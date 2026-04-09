package k8s

import (
	"context"
	"fmt"
	"mini-porter/internal/config"
	"os/exec"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

func CreateService(client *kubernetes.Clientset, cfg *config.Config) error {
	fmt.Println("[4/4] Creating service...")
	servicesClient := client.CoreV1().Services("default")
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: cfg.Name,
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeNodePort,
			Selector: map[string]string{
				"app": cfg.Name,
			},
			Ports: []corev1.ServicePort{
				{
					Port:       int32(cfg.Port),
					TargetPort: intstr.FromInt(cfg.Port),
					NodePort:   0,
				},
			},
		},
	}

	svc, err := servicesClient.Create(context.TODO(), service, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create service: %v", err)
	}

	nodeport := svc.Spec.Ports[0].NodePort

	node_ip := exec.Command("minikube", "ip")
	output, err := node_ip.Output()
	if err != nil {
		return fmt.Errorf("failed to get node IP: %v", err)
	}
	ip := strings.TrimSpace(string(output))

	fmt.Println("Service created successfully")
	fmt.Printf("App running at http://%s:%d\n", ip, nodeport)

	return nil
}

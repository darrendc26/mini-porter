package k8s

import (
	"context"
	"fmt"

	"github.com/darrendc26/mini-porter/internal/config"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ServiceInfo struct {
	Name  string
	Port  int
	Image string
}

func CreateDeployment(client *kubernetes.Clientset, cfg *config.Config, svc ServiceInfo) error {
	deploymentsClient := client.AppsV1().Deployments("default")

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: svc.Name,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(int32(cfg.Replicas)),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": svc.Name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": svc.Name},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  svc.Name,
							Image: svc.Image,
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: int32(svc.Port),
								},
							},
							ImagePullPolicy: corev1.PullAlways,
						},
					},
				},
			},
		},
	}

	_, err := deploymentsClient.Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		if errors.IsAlreadyExists(err) {
			fmt.Println("[4/6] Deployment 				Exists, updating...")

			existing, err := deploymentsClient.Get(context.TODO(), svc.Name, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("failed to get existing deployment: %w", err)
			}

			existing.Spec.Replicas = int32Ptr(int32(cfg.Replicas))
			existing.Spec.Template.Spec.Containers[0].Image = svc.Image

			_, err = deploymentsClient.Update(context.TODO(), existing, metav1.UpdateOptions{})
			if err != nil {
				return fmt.Errorf("failed to update deployment: %w", err)
			}

			return nil
		}
		return err
	}

	fmt.Println("[4/6] Deployment 				Completed")
	return nil
}

func int32Ptr(i int32) *int32 { return &i }

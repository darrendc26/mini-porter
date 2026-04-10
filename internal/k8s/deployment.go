package k8s

import (
	"context"
	"fmt"
	"mini-porter/internal/config"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateDeployment(client *kubernetes.Clientset, cfg *config.Config) error {
	fmt.Println("[3/4] Creating deployment and service...")
	deploymentsClient := client.AppsV1().Deployments("default")

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: cfg.Name,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(int32(cfg.Replicas)),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": cfg.Name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": cfg.Name},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  cfg.Name,
							Image: cfg.Image,
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: int32(cfg.Port),
								},
							},
							ImagePullPolicy: corev1.PullIfNotPresent,
						},
					},
				},
			},
		},
	}

	_, err := deploymentsClient.Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		if errors.IsAlreadyExists(err) {
			fmt.Println("Deployment exists, updating...")

			_, err = deploymentsClient.Update(context.TODO(), deployment, metav1.UpdateOptions{})
			if err != nil {
				return err
			}

			return nil
		}
		return err
	}

	fmt.Println("Deployment created successfully")
	return nil
}

func int32Ptr(i int32) *int32 { return &i }

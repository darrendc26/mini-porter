package k8s

import (
	"context"
	"fmt"

	"github.com/darrendc26/mini-porter/internal/config"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

type ServiceInfo struct {
	Name  string
	Port  int
	Image string
}

func CreateDeployment(client *kubernetes.Clientset, cfg *config.Config, svc ServiceInfo) error {
	namespace := cfg.Name

	deploymentsClient := client.AppsV1().Deployments(namespace)

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
					Labels: map[string]string{
						"app":        svc.Name,
						"managed-by": "mini-porter",
					},
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
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("100m"),
									corev1.ResourceMemory: resource.MustParse("128Mi"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("500m"),
									corev1.ResourceMemory: resource.MustParse("512Mi"),
								},
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/",
										Port: intstr.FromInt(int(svc.Port)),
									},
								},
								InitialDelaySeconds: 5,
								PeriodSeconds:       10,
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/",
										Port: intstr.FromInt(int(svc.Port)),
									},
								},
								InitialDelaySeconds: 10,
								PeriodSeconds:       10,
							},
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

			existing.Spec = deployment.Spec

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

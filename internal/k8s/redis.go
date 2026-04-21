package k8s

import (
	"context"
	"fmt"

	"github.com/darrendc26/mini-porter/internal/config"
	appsV1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

func CreateRedis(ctx context.Context, client *kubernetes.Clientset, dep config.Dependency) error {
	err := createRedisPVC(ctx, client, &dep)
	if err != nil {
		return fmt.Errorf("Error creating redis pvc: %v", err)
	}

	err = CreateRedisDeployment(ctx, client, &dep)
	if err != nil {
		return fmt.Errorf("Error creating redis deployment: %v", err)
	}
	err = createRedisService(ctx, client, &dep)
	if err != nil {
		return fmt.Errorf("Error creating redis service: %v", err)
	}

	err = wait(ctx, client, &dep)
	if err != nil {
		return fmt.Errorf("Error waiting for redis deployment: %v", err)
	}
	return nil
}

func CreateRedisDeployment(ctx context.Context, client *kubernetes.Clientset, dep *config.Dependency) error {
	deploymentClient := client.AppsV1().Deployments("default")
	envVars := buildEnvVars(dep.Env)

	deployment := &appsV1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: dep.Name,
			Labels: map[string]string{
				"app":     dep.Name,
				"service": dep.Type,
			},
		},
		Spec: appsV1.DeploymentSpec{
			Replicas: int32Ptr(int32(1)),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":     dep.Name,
					"service": dep.Type,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":     dep.Name,
						"service": dep.Type,
					},
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: "redis-data",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: dep.Name + "-pvc",
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:  dep.Name,
							Image: "redis:latest",
							Ports: []corev1.ContainerPort{
								{ContainerPort: int32(dep.Port)},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "redis-data",
									MountPath: "/data",
								},
							},
							Env: envVars,
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									Exec: &corev1.ExecAction{
										Command: []string{
											"redis-cli",
											"-a", dep.Env["REDIS_PASSWORD"],
											"ping",
										},
									},
								},
								InitialDelaySeconds: 5,
								PeriodSeconds:       5,
							},
						},
					},
				},
			},
		},
	}

	_, err := deploymentClient.Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		existing, err := deploymentClient.Get(ctx, dep.Name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("Failed to get existing deployment: %w", err)
		}

		existing.Spec.Replicas = int32Ptr(int32(1))
		existing.Spec.Template.Spec.Containers[0].Image = "redis:latest"
		existing.Spec.Template.Spec.Containers[0].Env = envVars

		_, err = deploymentClient.Update(ctx, existing, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("Failed to update deployment: %w", err)
		}
	}

	return nil
}

func createRedisService(ctx context.Context, client *kubernetes.Clientset, dep *config.Dependency) error {
	serviceClient := client.CoreV1().Services("default")

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: dep.Name,
			Labels: map[string]string{
				"app":     dep.Name,
				"service": dep.Type,
			},
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Name:       "redis",
					Port:       int32(dep.Port),
					TargetPort: intstr.FromInt(dep.Port),
				},
			},
			Selector: map[string]string{
				"app":     dep.Name,
				"service": dep.Type,
			},
		},
	}

	_, err := serviceClient.Create(ctx, service, metav1.CreateOptions{})
	if err != nil {
		existing, err := serviceClient.Get(ctx, dep.Name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("Failed to get existing service: %w", err)
		}

		existing.Spec.Ports[0].Port = int32(dep.Port)

		_, err = serviceClient.Update(ctx, existing, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("Failed to update service: %w", err)
		}
	}

	return nil
}

func createRedisPVC(ctx context.Context, client *kubernetes.Clientset, dep *config.Dependency) error {
	pvcClient := client.CoreV1().PersistentVolumeClaims("default")

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name: dep.Name + "-pvc",
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("1Gi"),
				},
			},
		},
	}

	_, err := pvcClient.Create(ctx, pvc, metav1.CreateOptions{})
	if err != nil {
		existing, err := pvcClient.Get(ctx, dep.Name+"-pvc", metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("Failed to get existing pvc: %w", err)
		}

		existing.Spec.Resources.Requests[corev1.ResourceStorage] = resource.MustParse("1Gi")

		_, err = pvcClient.Update(ctx, existing, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("Failed to update pvc: %w", err)
		}
	}

	return nil
}

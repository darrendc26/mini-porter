package k8s

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/darrendc26/mini-porter/internal/config"
	"github.com/joho/godotenv"
	appsV1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

func CreatePostgres(ctx context.Context, client *kubernetes.Clientset, dependency config.Dependency) error {
	err := createPostgresPVC(ctx, client, &dependency)
	if err != nil {
		return fmt.Errorf("Error creating postgres pvc: %v", err)
	}

	err = createPostgresDeployment(ctx, client, &dependency)
	if err != nil {
		return fmt.Errorf("Error creating postgres deployment: %v", err)
	}
	err = createPostgresService(ctx, client, &dependency)
	if err != nil {
		return fmt.Errorf("Error creating postgres service: %v", err)
	}

	err = wait(ctx, client, &dependency)
	if err != nil {
		return fmt.Errorf("Error waiting for postgres deployment: %v", err)
	}

	return nil
}

func createPostgresDeployment(ctx context.Context, client *kubernetes.Clientset, dep *config.Dependency) error {
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
							Name: "postgres-data",
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
							Image: "postgres:latest",
							Ports: []corev1.ContainerPort{
								{ContainerPort: 5432},
							},

							Env: envVars,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "postgres-data",
									MountPath: "/var/lib/postgresql/data",
								},
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									Exec: &corev1.ExecAction{
										Command: []string{
											"pg_isready",
											"-U", dep.Env["POSTGRES_USER"],
										},
									},
								},
								InitialDelaySeconds: 5,
								PeriodSeconds:       5,
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									Exec: &corev1.ExecAction{
										Command: []string{
											"pg_isready",
											"-U", dep.Env["POSTGRES_USER"],
										},
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

	_, err := deploymentClient.Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		existing, err := deploymentClient.Get(ctx, dep.Name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("Failed to get existing deployment: %w", err)
		}

		existing.Spec.Replicas = int32Ptr(int32(1))
		existing.Spec.Template.Spec.Containers[0].Image = "postgres:latest"
		existing.Spec.Template.Spec.Containers[0].Env = envVars

		_, err = deploymentClient.Update(ctx, existing, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("Failed to update deployment: %w", err)
		}
	}

	return nil
}

func createPostgresService(ctx context.Context, client *kubernetes.Clientset, dep *config.Dependency) error {
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
					Name:       "postgres",
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

func createPostgresPVC(ctx context.Context, client *kubernetes.Clientset, dep *config.Dependency) error {
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

func buildEnvVars(env map[string]string) []corev1.EnvVar {
	var result []corev1.EnvVar

	err := godotenv.Load()
	if err != nil {
		fmt.Println("No .env file found.")
	}

	for key, value := range env {
		if strings.HasPrefix(value, "env.") {
			envKey := strings.TrimPrefix(value, "env.")
			resolved := os.Getenv(envKey)

			if resolved == "" {
				fmt.Printf(" Missing env: %s\n", envKey)
				os.Exit(1)
			}

			value = resolved
		}

		result = append(result, corev1.EnvVar{
			Name:  key,
			Value: value,
		})
	}

	return result
}

func wait(ctx context.Context, client *kubernetes.Clientset, dep *config.Dependency) error {
	deployment, _ := client.AppsV1().Deployments("default").Get(ctx, dep.Name, metav1.GetOptions{})

	ready := deployment.Status.ReadyReplicas
	total := *deployment.Spec.Replicas

	fmt.Printf("Pods Ready: %d/%d\n", ready, total)

	if ready == total && total > 0 {
		return nil
	}
	time.Sleep(12 * time.Second)
	return wait(ctx, client, dep)
}

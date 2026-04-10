package k8s

import (
	"context"
	"fmt"
	"time"

	"github.com/darrendc26/mini-porter/internal/config"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func WaitForDeployment(client *kubernetes.Clientset, cfg *config.Config) error {
	fmt.Println("Waiting for deployment to be ready...")

	for {
		deploymentsClient := client.AppsV1().Deployments("default")
		deployment, err := deploymentsClient.Get(context.TODO(), cfg.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		ready := deployment.Status.ReadyReplicas
		total := *deployment.Spec.Replicas

		fmt.Printf("Pods Ready: %d/%d\n", ready, total)

		if ready == total && total > 0 {
			break
		}

		time.Sleep(12 * time.Second)
	}

	return nil
}

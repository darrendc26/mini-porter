package cmd

import (
	"context"
	"fmt"

	"github.com/darrendc26/mini-porter/internal/config"
	"github.com/darrendc26/mini-porter/internal/k8s"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check the status of a deployment",

	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		cfg, err := config.LoadConfig("mini-porter.yaml")
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}

		client, err := k8s.GetClient()
		if err != nil {
			fmt.Printf("Error getting k8s client: %v\n", err)
			return
		}

		for _, service := range cfg.Services {
			namespace := cfg.Name
			deploymentsClient := client.AppsV1().Deployments(namespace)
			deployment, err := deploymentsClient.Get(ctx, service.Name, metav1.GetOptions{})
			if err != nil {
				fmt.Printf("Error getting deployment: %v\n", err)
				return
			}
			fmt.Printf("App: %s\n", service.Name)
			fmt.Printf("Replicas: %d/%d\n",
				deployment.Status.ReadyReplicas, deployment.Status.Replicas)

			pods, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
				LabelSelector: fmt.Sprintf("app=%s", service.Name),
			})
			if err != nil {
				fmt.Printf("Error getting pods: %v\n", err)
				return
			}
			fmt.Printf("Pods: \n")
			for _, pod := range pods.Items {
				fmt.Printf("  %s\n", pod.Name)
			}
		}

		for _, dep := range cfg.Dependencies {
			namespace := cfg.Name
			deploymentsClient := client.AppsV1().Deployments(namespace)
			deployment, err := deploymentsClient.Get(context.TODO(), dep.Name, metav1.GetOptions{})
			if err != nil {
				fmt.Printf("Error getting deployment: %v\n", err)
				return
			}
			fmt.Printf("App: %s\n", dep.Name)
			fmt.Printf("Replicas: %d/%d\n",
				deployment.Status.ReadyReplicas, deployment.Status.Replicas)

			pods, err := client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
				LabelSelector: fmt.Sprintf("app=%s", dep.Name),
			})
			if err != nil {
				fmt.Printf("Error getting pods: %v\n", err)
				return
			}
			fmt.Printf("Pods: \n")
			for _, pod := range pods.Items {
				fmt.Printf("  %s\n", pod.Name)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

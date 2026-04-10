package cmd

import (
	"context"
	"fmt"

	"github.com/darrendc26/mini-porter/internal/config"
	"github.com/darrendc26/mini-porter/internal/k8s"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a deployment",

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

		fmt.Println("Deleting deployment...")
		deploymentsClient := client.AppsV1().Deployments("default")
		err = deploymentsClient.Delete(ctx, cfg.Name, metav1.DeleteOptions{})
		if err != nil {
			fmt.Printf("Error deleting deployment: %v\n", err)
			return
		}

		fmt.Println("Deleting service...")
		servicesClient := client.CoreV1().Services("default")
		err = servicesClient.Delete(ctx, cfg.Name, metav1.DeleteOptions{})
		if err != nil {
			fmt.Printf("Error deleting service: %v\n", err)
			return
		}
		fmt.Println("Deployment and service deleted successfully!")

		if err := k8s.DeleteIngress(client, cfg); err != nil {
			fmt.Printf("Error deleting ingress: %v\n", err)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}

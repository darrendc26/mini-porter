package cmd

import (
	"context"
	"fmt"

	"github.com/darrendc26/mini-porter/internal/config"
	"github.com/darrendc26/mini-porter/internal/k8s"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var serviceName string

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete app or specific service",

	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		cfg, err := config.LoadConfig("mini-porter.yaml")
		namespace := cfg.Name

		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}

		client, err := k8s.GetClient()
		if err != nil {
			fmt.Printf("Error getting k8s client: %v\n", err)
			return
		}

		//  Label selector
		var labelSelector string
		if serviceName != "" {
			fmt.Printf("Deleting service: %s\n", serviceName)
			labelSelector = fmt.Sprintf("app=%s,service=%s", cfg.Name, serviceName)
		} else {
			fmt.Println("Deleting entire app...")
			labelSelector = fmt.Sprintf("app=%s", cfg.Name)
		}

		//  Delete Deployments
		fmt.Println("Deleting deployments...")
		err = client.AppsV1().Deployments(namespace).DeleteCollection(
			ctx,
			metav1.DeleteOptions{},
			metav1.ListOptions{LabelSelector: labelSelector},
		)
		if err != nil {
			fmt.Printf("Error deleting deployments: %v\n", err)
			return
		}

		//  Delete Services
		fmt.Println("Deleting services...")
		servicesClient := client.CoreV1().Services(namespace)

		svcList, err := servicesClient.List(ctx, metav1.ListOptions{
			LabelSelector: labelSelector,
		})
		if err != nil {
			fmt.Printf("Error listing services: %v\n", err)
			return
		}

		for _, svc := range svcList.Items {
			fmt.Printf("Deleting service: %s\n", svc.Name)

			err := servicesClient.Delete(ctx, svc.Name, metav1.DeleteOptions{})
			if err != nil {
				fmt.Printf("Error deleting service %s: %v\n", svc.Name, err)
			}
		}

		//  Delete Pods (cleanup)
		fmt.Println("Deleting pods...")
		err = client.CoreV1().Pods(namespace).DeleteCollection(
			ctx,
			metav1.DeleteOptions{},
			metav1.ListOptions{LabelSelector: labelSelector},
		)
		if err != nil {
			fmt.Printf("Error deleting pods: %v\n", err)
			return
		}

		//  Handle ingress
		if serviceName != "" {
			err = k8s.DeleteIngressRule(client, cfg, serviceName)
		} else {
			err = k8s.DeleteIngress(client, cfg)
		}

		if err != nil {
			fmt.Printf("Error handling ingress: %v\n", err)
			return
		}

		fmt.Println("Delete completed successfully!")
	},
}

func init() {
	deleteCmd.Flags().StringVar(&serviceName, "service", "", "Delete specific service")
	rootCmd.AddCommand(deleteCmd)
}

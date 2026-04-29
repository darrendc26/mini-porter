package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/darrendc26/mini-porter/internal/k8s"
	"github.com/spf13/cobra"
)

var projectID string
var region string
var clusterName string
var env string

// var credPath string

var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Manage clusters",
}

var clusterCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a Kubernetes cluster on GCP",
	Long:  "Create a GKE cluster using stored or provided credentials",
	Run: func(cmd *cobra.Command, args []string) {
		switch env {
		case "gcp":
			err := createGCPCluster(projectID, region, clusterName)
			if err != nil {
				fmt.Println("Failed to create cluster:", err)
				return
			}
			fmt.Println("Cluster created successfully")
		case "local":
			err := createLocalCluster()
			if err != nil {
				fmt.Println("Failed to create cluster:", err)
				return
			}
			fmt.Println("Cluster created successfully")
		default:
			fmt.Println("Invalid environment:", env)
		}

	},
}

func init() {
	clusterCmd.AddCommand(clusterCreateCmd)
	rootCmd.AddCommand(clusterCmd)
	clusterCreateCmd.Flags().StringVarP(&env, "env", "e", "", "Environment")
	clusterCreateCmd.Flags().StringVarP(&projectID, "project-id", "p", "", "Project ID")
	clusterCreateCmd.Flags().StringVarP(&region, "region", "r", "", "Region")
	clusterCreateCmd.Flags().StringVarP(&clusterName, "name", "n", "", "Cluster name")
	clusterCreateCmd.Flags().StringVarP(&credPath, "path", "P", "", "Path to the credentials JSON file")
}

func createGCPCluster(projectID, region, clusterName string) error {
	if projectID == "" {
		return fmt.Errorf("project-id is required")
	}
	if region == "" {
		return fmt.Errorf("region is required")
	}
	if clusterName == "" {
		return fmt.Errorf("cluster name is required")
	}

	if credPath == "" {
		credPath, err := getCredentialsPath()
		if err != nil {
			return fmt.Errorf("failed to get credentials path: %w", err)
		}
		if credPath == "" {
			return fmt.Errorf("no credentials found. Run `miniporter login`")
		}
	}

	fmt.Println("Creating cluster...")
	fmt.Printf("Project: %s | Region: %s | Name: %s\n", projectID, region, clusterName)

	err := k8s.CreateGKECluster(
		context.Background(),
		credPath,
		projectID,
		region,
		clusterName,
	)
	if err != nil {
		return fmt.Errorf("failed to create cluster: %w", err)
	}

	return nil
}

func createLocalCluster() error {
	cmd := exec.Command("minikube", "start")

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Println("Error starting minikube:", err)
	}

	return nil
}

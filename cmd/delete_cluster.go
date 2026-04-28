package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"
	container "google.golang.org/api/container/v1"
	"google.golang.org/api/option"
	"k8s.io/client-go/tools/clientcmd"
)

// go run main.go cluster delete -n minikube

var location string
var deleteClusterCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a cluster",
	Run: func(cmd *cobra.Command, args []string) {
		if clusterName == "" {
			fmt.Println("cluster name is required")
			return
		}

		// 🔥 Local cluster case
		if clusterName == "minikube" {
			err := deleteLocalCluster()
			if err != nil {
				fmt.Println("❌ Failed to delete local cluster:", err)
				return
			}
			fmt.Println("✅ Minikube deleted successfully")
			return
		}

		// 🔥 Cloud cluster case
		if projectID == "" || location == "" {
			fmt.Println("project-id and location are required for cloud clusters")
			return
		}

		err := deleteGKECluster()
		if err != nil {
			fmt.Println("❌ Failed to delete GKE cluster:", err)
			return
		}
	},
}

func init() {
	deleteClusterCmd.Flags().StringVarP(&location, "location", "l", "", "Region")
	deleteClusterCmd.Flags().StringVarP(&clusterName, "name", "n", "", "Cluster name")
	deleteClusterCmd.Flags().StringVarP(&projectID, "project-id", "p", "", "Project ID")
	clusterCmd.AddCommand(deleteClusterCmd)
}

func deleteLocalCluster() error {
	cmd := exec.Command("minikube", "delete")

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Println("Deleting local cluster (minikube)...")

	return cmd.Run()
}

func deleteGKECluster() error {
	ctx := context.Background()
	credPath, err := getCredentials()
	if err != nil {
		fmt.Println(err)
		return err
	}

	svc, err := container.NewService(ctx,
		option.WithCredentialsFile(credPath.Credentials),
	)
	if err != nil {
		fmt.Println(err)
		return err
	}

	name := fmt.Sprintf(
		"projects/%s/locations/%s/clusters/%s",
		projectID,
		location,
		clusterName,
	)

	op, err := svc.Projects.Locations.Clusters.Delete(name).Do()
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("Deleting cluster...")

	opName := op.Name
	if !strings.Contains(opName, "projects/") {
		opName = fmt.Sprintf(
			"projects/%s/locations/%s/operations/%s",
			projectID,
			region,
			op.Name,
		)
	}

	for {
		opStatus, err := svc.Projects.Locations.Operations.Get(opName).Do()
		if err != nil {
			fmt.Println(err)
			return err
		}

		if opStatus.Status == "DONE" {
			if opStatus.Error != nil {
				fmt.Println("Delete failed:", opStatus.Error)
				return err
			}
			break
		}

		fmt.Println("⏳ Deleting...")
		time.Sleep(5 * time.Second)
	}

	kubeconfigPath := os.Getenv("HOME") + "/.kube/config"

	config, err := clientcmd.LoadFromFile(kubeconfigPath)
	if err == nil {
		delete(config.Clusters, clusterName)
		delete(config.Contexts, clusterName)
		delete(config.AuthInfos, clusterName)

		if config.CurrentContext == clusterName {
			config.CurrentContext = ""
		}

		clientcmd.WriteToFile(*config, kubeconfigPath)
		fmt.Println("🧹 Removed kubeconfig context:", clusterName)
	}

	fmt.Println("✅ Cluster deleted successfully")
	return nil
}

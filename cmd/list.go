package cmd

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"time"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	api "k8s.io/client-go/tools/clientcmd/api"
)

type ClusterType string

const (
	LocalCluster   ClusterType = "local"
	GCPCluster     ClusterType = "gcp"
	CloudCluster   ClusterType = "cloud"
	UnknownCluster ClusterType = "unknown"
)

type ClusterInfo struct {
	Name   string
	Server string
	Type   ClusterType
}

func DetectClusterType(ctx *api.Context, config *api.Config) ClusterType {
	cluster := config.Clusters[ctx.Cluster]
	if cluster == nil {
		return UnknownCluster
	}

	server := cluster.Server

	u, err := url.Parse(server)
	if err != nil {
		return UnknownCluster
	}

	host := u.Hostname()
	ip := net.ParseIP(host)

	if ip != nil && (ip.IsPrivate() || ip.IsLoopback()) {
		return LocalCluster
	}

	auth := config.AuthInfos[ctx.AuthInfo]
	if auth != nil && auth.AuthProvider != nil && auth.AuthProvider.Name == "gcp" {
		return GCPCluster
	}

	if ip != nil {
		return CloudCluster
	}

	return UnknownCluster
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List clusters",
	RunE: func(cmd *cobra.Command, args []string) error {
		kubeconfigPath := os.Getenv("HOME") + "/.kube/config"

		config, err := clientcmd.LoadFromFile(kubeconfigPath)
		if err != nil {
			return err
		}

		if len(config.Contexts) == 0 {
			fmt.Println("No clusters found")
			return nil
		}

		fmt.Println("Available clusters:")

		for name, ctx := range config.Contexts {
			cluster := config.Clusters[ctx.Cluster]
			if cluster == nil {
				continue
			}

			clusterType := DetectClusterType(ctx, config)

			// if !IsClusterAlive(kubeconfigPath, name) {
			// 	continue
			// }

			current := " "
			if name == config.CurrentContext {
				current = "*"
			}

			fmt.Printf("%s %s -> %s\n", current, name, cluster.Server)
			fmt.Printf("   type: %s\n", clusterType)
		}

		return nil
	},
}

func IsClusterAlive(kubeconfigPath, contextName string) bool {
	overrides := &clientcmd.ConfigOverrides{
		CurrentContext: contextName,
	}

	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		overrides,
	).ClientConfig()

	if err != nil {
		return false
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err = clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	return err == nil
}

func init() {
	clusterCmd.AddCommand(listCmd)
}

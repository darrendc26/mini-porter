package k8s

import (
	"fmt"
	"net"
	"net/url"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	clusterType "github.com/darrendc26/mini-porter/internal/config"
)

func GetClient() (*kubernetes.Clientset, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}

	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		configOverrides,
	)

	restConfig, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(restConfig)
}

func GetCurrentClusterType() (clusterType.ClusterType, error) {
	kubeconfigPath := os.Getenv("HOME") + "/.kube/config"

	config, err := clientcmd.LoadFromFile(kubeconfigPath)
	if err != nil {
		return clusterType.UnknownCluster, err
	}

	ctx := config.Contexts[config.CurrentContext]
	if ctx == nil {
		return clusterType.UnknownCluster, fmt.Errorf("no context found")
	}

	cluster := config.Clusters[ctx.Cluster]
	if cluster == nil {
		return clusterType.UnknownCluster, fmt.Errorf("no cluster found")
	}

	// 🔍 check server
	u, err := url.Parse(cluster.Server)
	if err != nil {
		return clusterType.UnknownCluster, err
	}

	host := u.Hostname()
	ip := net.ParseIP(host)

	// ✅ LOCAL
	if ip != nil && (ip.IsPrivate() || ip.IsLoopback()) {
		fmt.Println("Current context: LOCAL")
		return clusterType.LocalCluster, nil
	}

	// ✅ GCP
	auth := config.AuthInfos[ctx.AuthInfo]
	if auth != nil && auth.AuthProvider != nil && auth.AuthProvider.Name == "gcp" {
		fmt.Println("Current context: GCP")
		return clusterType.GCPCluster, nil
	}

	// 🌐 fallback
	fmt.Println("Current context: CLOUD")
	return clusterType.CloudCluster, nil
}

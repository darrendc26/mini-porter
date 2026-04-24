package k8s

import (
	"fmt"
	"strings"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
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

func GetCurrentContext() (string, error) {
	cfg, _ := clientcmd.NewDefaultClientConfigLoadingRules().Load()
	ctx := cfg.CurrentContext

	switch {
	case strings.Contains(ctx, "minikube"):
		fmt.Println("Current context : minikube")
		return "minikube", nil
	case strings.Contains(ctx, "gke"):
		fmt.Println("Current context : GCP")
		return "gke", nil
	case strings.Contains(ctx, "eks"):
		fmt.Println("Current context : AWS")
		return "eks", nil
	case strings.Contains(ctx, "do-"):
		fmt.Println("Current context : DigitalOcean")
		return "do", nil
	default:
		fmt.Println("Environment unknown")
	}
	return "", nil
}

package k8s

import (
	"fmt"
	"os"

	"k8s.io/client-go/tools/clientcmd"
)

func UseCluster(contextName string) error {
	kubeconfigPath := os.Getenv("HOME") + "/.kube/config"

	existingConfig, err := clientcmd.LoadFromFile(kubeconfigPath)
	if err != nil {
		return fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	if contextName == "" {
		contextName = existingConfig.CurrentContext
	}

	if _, exists := existingConfig.Contexts[contextName]; !exists {
		fmt.Println("Context not found:", contextName)
		fmt.Println("Available contexts:")

		for name := range existingConfig.Contexts {
			if name == existingConfig.CurrentContext {
				fmt.Println("*", name)
			} else {
				fmt.Println(" ", name)
			}
		}

		return fmt.Errorf("context %s does not exist", contextName)
	}

	existingConfig.CurrentContext = contextName

	return clientcmd.WriteToFile(*existingConfig, kubeconfigPath)
}

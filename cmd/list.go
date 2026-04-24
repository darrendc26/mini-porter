package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List apps",
	RunE: func(cmd *cobra.Command, args []string) error {
		kubeconfigPath := os.Getenv("HOME") + "/.kube/config"

		config, err := clientcmd.LoadFromFile(kubeconfigPath)
		if err != nil {
			return err
		}

		fmt.Println("Available clusters:")

		for name := range config.Contexts {
			if name == config.CurrentContext {
				fmt.Println("*", name)
			} else {
				fmt.Println(" ", name)
			}
		}

		return nil
	},
}

func init() {
	clusterCmd.AddCommand(listCmd)
}

package cmd

import (
	"fmt"

	"github.com/darrendc26/mini-porter/internal/k8s"
	"github.com/spf13/cobra"
)

var contextCmd = &cobra.Command{
	Use:   "use [context]",
	Short: "Switching K8s context",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			fmt.Println("Please provide a context name")
			return nil
		}
		err := k8s.UseCluster(args[0])
		if err != nil {
			fmt.Printf("Error switching context: %v\n", err)
			return nil
		}
		fmt.Printf("Successfully switched context to %s\n", args[0])
		return nil
	},
}

func init() {
	clusterCmd.AddCommand(contextCmd)
}

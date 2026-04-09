package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy an application",
	Long:  "Deploy an application to a Kubernetes cluster",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting deployment...")
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)
}

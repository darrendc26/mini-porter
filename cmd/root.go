package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "mini-porter",
	Short: "A PaaS CLI for Kubernetes",
	Long:  "A Paas CLI for Kubernetes that simplifies application deployment and management.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

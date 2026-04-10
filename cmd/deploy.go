package cmd

import (
	"context"
	"fmt"

	"github.com/darrendc26/mini-porter/internal/config"
	"github.com/darrendc26/mini-porter/internal/deploy"

	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy an application",
	Long:  "Deploy an application to a Kubernetes cluster",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		cfg, err := config.LoadConfig("mini-porter.yaml")
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}
		fmt.Printf("Deploying %s to %s:%d\n", cfg.Name, cfg.Image, cfg.Port)
		fmt.Println("Starting deployment...")

		if err := deploy.Deploy(ctx, cfg); err != nil {
			fmt.Printf("Error building image: %v\n", err)
			return
		}

	},
}

func init() {
	rootCmd.AddCommand(deployCmd)
}

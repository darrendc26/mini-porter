package cmd

import (
	"context"
	"fmt"
	"mini-porter/internal/config"
	"mini-porter/internal/docker"

	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy an application",
	Long:  "Deploy an application to a Kubernetes cluster",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		cfg, err := config.LoadConfig("porter.yaml")
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}
		fmt.Printf("Deploying %s to %s:%d\n", cfg.Name, cfg.Image, cfg.Port)
		fmt.Println("Starting deployment...")

		// Build Docker image
		if err := docker.BuildDockerImage(ctx, cfg.Image); err != nil {
			fmt.Printf("Error building image: %v\n", err)
			return
		}

		// Push Docker image
		if err := docker.PushDockerImage(ctx, cfg.Image); err != nil {
			fmt.Printf("Error pushing image: %v\n", err)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)
}

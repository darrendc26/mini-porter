package k8s

import (
	"context"
	"fmt"

	"github.com/darrendc26/mini-porter/internal/config"
	"k8s.io/client-go/kubernetes"
)

func CreateDependencies(ctx context.Context, client *kubernetes.Clientset, cfg *config.Config) error {
	for _, dep := range cfg.Dependencies {
		switch dep.Type {
		case "postgres":
			fmt.Println("Postgres  					Creating...")
			err := CreatePostgres(ctx, client, dep)
			if err != nil {
				return fmt.Errorf("Error creating postgres deployment: %v", err)
			}
		case "redis":
			fmt.Println("Redis  					Creating...")
			err := CreateRedis(ctx, client, dep)
			if err != nil {
				return fmt.Errorf("Error creating redis deployment: %v", err)
			}
		}
	}
	return nil
}

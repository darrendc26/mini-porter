package deploy

import (
	"context"
	"fmt"
	"mini-porter/internal/config"
	"mini-porter/internal/docker"
	"mini-porter/internal/k8s"
)

func Deploy(ctx context.Context, cfg *config.Config) error {
	if err := docker.BuildDockerImage(ctx, cfg.Image); err != nil {
		return fmt.Errorf("error building image: %v", err)
	}

	if err := docker.PushDockerImage(ctx, cfg.Image); err != nil {
		return err
	}

	client, err := k8s.GetClient()
	if err != nil {
		return fmt.Errorf("error getting k8s client: %v", err)
	}

	if err := k8s.CreateDeployment(client, cfg); err != nil {
		return fmt.Errorf("error creating deployment: %v", err)
	}

	if err := k8s.WaitForDeployment(client, cfg); err != nil {
		return fmt.Errorf("error waiting for deployment: %v", err)
	}

	if err := k8s.CreateService(client, cfg); err != nil {
		return fmt.Errorf("error creating service: %v", err)
	}

	if err := k8s.CreateIngress(client, cfg); err != nil {
		return fmt.Errorf("error creating ingress: %v", err)
	}

	fmt.Println("Deployment completed successfully!")
	fmt.Println("Run command:\n mini-porter host add ")
	return nil
}

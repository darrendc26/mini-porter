package deploy

import (
	"context"
	"fmt"
	"os"

	"github.com/darrendc26/mini-porter/internal/config"
	"github.com/darrendc26/mini-porter/internal/detector"
	"github.com/darrendc26/mini-porter/internal/docker"
)

func Deploy(ctx context.Context, cfg *config.Config) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	// var path []string

	projects, err := detector.DetectProjects(wd)

	pathSet := make(map[string]int)
	for _, service := range cfg.Services {
		pathSet[service.Path] = service.Port
	}

	for _, project := range projects {
		if port, exists := pathSet[project.Path]; exists {
			fmt.Printf("Detected project: %s (%s)\n", project.Path, project.Type)
			fmt.Printf("Port:%d\n", port)
			if err := docker.CreateDockerfile(ctx, project, port); err != nil {
				return fmt.Errorf("error generating Dockerfile: %v", err)
			}
		}
	}
	// for _, project := range projects {
	// 	if strings.Contains(project.Path, path) {
	// 		fmt.Printf("YES")
	// 	}
	// 	fmt.Printf("Detected project: %s (%s)\n", project.Path, project.Type)
	// }

	if err != nil {
		return fmt.Errorf("error detecting projects: %v", err)
	}

	// if len(projects) != 0 {
	// 	for _, project := range projects {
	// 		if err := docker.CreateDockerfile(ctx, project, cfg); err != nil {
	// 			return fmt.Errorf("error generating Dockerfile: %v", err)
	// 		}
	// 	}
	// }
	// if err := docker.BuildDockerImage(ctx, cfg.Image); err != nil {
	// 	return fmt.Errorf("error building image: %v", err)
	// }

	// if err := docker.PushDockerImage(ctx, cfg.Image); err != nil {
	// 	return err
	// }

	// client, err := k8s.GetClient()
	// if err != nil {
	// 	return fmt.Errorf("error getting k8s client: %v", err)
	// }

	// if err := k8s.CreateDeployment(client, cfg); err != nil {
	// 	return fmt.Errorf("error creating deployment: %v", err)
	// }

	// if err := k8s.WaitForDeployment(client, cfg); err != nil {
	// 	return fmt.Errorf("error waiting for deployment: %v", err)
	// }

	// if err := k8s.CreateService(client, cfg); err != nil {
	// 	return fmt.Errorf("error creating service: %v", err)
	// }

	// if err := k8s.CreateIngress(client, cfg); err != nil {
	// 	return fmt.Errorf("error creating ingress: %v", err)
	// }

	fmt.Println("Deployment completed successfully!")
	fmt.Println("Run command:\n mini-porter host add ")
	return nil
}

package deploy

import (
	"context"
	"fmt"
	"time"

	"github.com/darrendc26/mini-porter/internal/config"
	"github.com/darrendc26/mini-porter/internal/k8s"
)

func Deploy(ctx context.Context, cfg *config.Config) error {
	// var urlList []string
	// var deployedServices []k8s.ServiceInfo
	// wd, err := os.Getwd()
	// if err != nil {
	// 	return err
	// }
	// projects, err := detector.DetectProjects(wd)

	client, err := k8s.GetClient()
	if err != nil {
		return fmt.Errorf("error getting k8s client: %v", err)
	}

	type ServiceInfo struct {
		Name  string
		Port  int
		Image string
	}

	pathSet := make(map[string]ServiceInfo)
	for _, service := range cfg.Services {
		pathSet[service.Path] = ServiceInfo{
			Name:  service.Name,
			Port:  service.Port,
			Image: fmt.Sprintf("%s:%s-v%d", cfg.Image, service.Name, time.Now().Unix()),
		}
	}

	// for _, project := range projects {
	// 	if svc, exists := pathSet[project.Path]; exists {
	// 		fmt.Printf("Processing project: %s %d\n", svc.Name, svc.Port)
	// 		if err := docker.CreateDockerfile(ctx, project, svc.Port); err != nil {
	// 			return fmt.Errorf("error generating Dockerfile: %v", err)
	// 		}

	// 		if err := docker.BuildDockerImage(ctx, svc.Image, project.Path); err != nil {
	// 			return fmt.Errorf("error building image: %v", err)
	// 		}

	// 		if err := docker.PushDockerImage(ctx, svc.Image); err != nil {
	// 			return fmt.Errorf("error pushing image: %v", err)
	// 		}

	// 		if err := k8s.CreateDeployment(client, cfg, k8s.ServiceInfo(svc)); err != nil {
	// 			return fmt.Errorf("error creating deployment: %v", err)
	// 		}

	// 		if err := k8s.WaitForDeployment(client, svc.Name); err != nil {
	// 			return fmt.Errorf("error waiting for deployment: %v", err)
	// 		}

	// 		if url, err := k8s.CreateService(client, cfg, k8s.ServiceInfo(svc)); err != nil {
	// 			return fmt.Errorf("error creating service: %v", err)
	// 		} else {
	// 			urlList = append(urlList, url)
	// 		}

	// 		deployedServices = append(deployedServices, k8s.ServiceInfo(svc))
	// 	}
	// 	fmt.Println("-----------------------------------")
	// 	fmt.Println(" ")
	// }

	// if err != nil {
	// 	return fmt.Errorf("error detecting projects: %v", err)
	// }

	// if err := k8s.CreateIngress(client, cfg, deployedServices); err != nil {
	// 	return fmt.Errorf("error creating ingress: %v", err)
	// }

	if err := k8s.CreateDependencies(ctx, client, cfg); err != nil {
		return fmt.Errorf("error creating dependencies: %v", err)
	}

	fmt.Println("Deployment completed successfully!")
	// for _, url := range urlList {
	// 	fmt.Println(url)
	// }
	fmt.Println("Run command:\n mini-porter host add ")
	return nil
}

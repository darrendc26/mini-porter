package k8s

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/container/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"

	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func CreateGKECluster(ctx context.Context, credPath, projectID, region, name string) error {
	ctx, cancel := context.WithTimeout(ctx, 45*time.Minute) // ← increase timeout
	defer cancel()

	svc, err := container.NewService(ctx,
		option.WithCredentialsFile(credPath),
	)
	if err != nil {
		return err
	}

	zones := []string{
		region + "-a",
		region + "-b",
		region + "-c",
	}

	for _, zone := range zones {
		fmt.Println("Trying zone:", zone)

		parent := fmt.Sprintf("projects/%s/locations/%s", projectID, zone)

		cluster := &container.Cluster{
			Name:             name,
			InitialNodeCount: 2,
			NodeConfig: &container.NodeConfig{
				MachineType: "e2-small",
			},
		}

		op, err := svc.Projects.Locations.Clusters.Create(parent, &container.CreateClusterRequest{
			Cluster: cluster,
		}).Do()
		if err != nil {
			if gErr, ok := err.(*googleapi.Error); ok && gErr.Code == 409 {
				fmt.Println("Cluster already exists, skipping creation...")
				return setupKubeconfig(ctx, svc, credPath, projectID, zone, name)
			}
			if isStockoutError(err) {
				fmt.Println("Stockout in", zone, "- trying next zone...")
				continue
			}
			return fmt.Errorf("failed to create cluster: %w", err)
		}

		fmt.Println(" Creating cluster (10-20 min)...")

		opName := op.Name
		if !strings.Contains(opName, "projects/") {
			opName = fmt.Sprintf("projects/%s/locations/%s/operations/%s",
				projectID, zone, op.Name)
		}

		zoneStockout := false

		for {
			select {
			case <-ctx.Done():
				return fmt.Errorf("cluster creation timed out after waiting")
			default:
			}

			opStatus, err := svc.Projects.Locations.Operations.Get(opName).Do()
			if err != nil {
				return fmt.Errorf("failed to get operation status: %w", err)
			}

			if opStatus.Status == "DONE" {
				if opStatus.Error != nil {
					if isStockoutError(fmt.Errorf("%v", opStatus.Error)) {
						fmt.Println("Stockout in", zone, "- retrying next zone...")
						zoneStockout = true
						break
					}
					return fmt.Errorf("cluster creation failed: %v", opStatus.Error)
				}
				fmt.Println("Cluster created in zone:", zone)
				return setupKubeconfig(ctx, svc, credPath, projectID, zone, name)
			}

			fmt.Println("Still provisioning")
			time.Sleep(15 * time.Second)
		}

		if zoneStockout {
			continue
		}
	}

	return fmt.Errorf("all zones exhausted in region %s. Try another region like us-central1", region)
}

func setupKubeconfig(ctx context.Context, svc *container.Service, credPath, projectID, zone, name string) error {
	parent := fmt.Sprintf("projects/%s/locations/%s", projectID, zone)

	clusterResp, err := svc.Projects.Locations.Clusters.Get(
		fmt.Sprintf("%s/clusters/%s", parent, name),
	).Do()
	if err != nil {
		return err
	}

	credData, err := os.ReadFile(credPath)
	if err != nil {
		return err
	}

	creds, err := google.CredentialsFromJSON(ctx, credData, container.CloudPlatformScope)
	if err != nil {
		return err
	}

	token, err := creds.TokenSource.Token()
	if err != nil {
		return err
	}

	caCert, err := base64.StdEncoding.DecodeString(clusterResp.MasterAuth.ClusterCaCertificate)
	if err != nil {
		return err
	}

	kubeConfig := clientcmdapi.Config{
		Clusters: map[string]*clientcmdapi.Cluster{
			name: {
				Server:                   "https://" + clusterResp.Endpoint,
				CertificateAuthorityData: caCert,
			},
		},
		AuthInfos: map[string]*clientcmdapi.AuthInfo{
			name: {
				Token: token.AccessToken,
			},
		},
		Contexts: map[string]*clientcmdapi.Context{
			name: {
				Cluster:  name,
				AuthInfo: name,
			},
		},
		CurrentContext: name,
	}

	err = MergeKubeConfig(kubeConfig, name)
	if err != nil {
		return err
	}

	fmt.Println("Kubeconfig merged")
	fmt.Println("Current context:", name)

	return nil
}

func isStockoutError(err error) bool {
	return strings.Contains(err.Error(), "GCE_STOCKOUT")
}

func MergeKubeConfig(newConfig clientcmdapi.Config, contextName string) error {
	kubeconfigPath := os.Getenv("HOME") + "/.kube/config"

	// Load existing config
	existingConfig, err := clientcmd.LoadFromFile(kubeconfigPath)
	if err != nil {
		// If no config exists, just write new one
		return clientcmd.WriteToFile(newConfig, kubeconfigPath)
	}

	// Merge clusters
	for k, v := range newConfig.Clusters {
		existingConfig.Clusters[k] = v
	}

	// Merge users
	for k, v := range newConfig.AuthInfos {
		existingConfig.AuthInfos[k] = v
	}

	// Merge contexts
	for k, v := range newConfig.Contexts {
		existingConfig.Contexts[k] = v
	}

	// Set current context
	existingConfig.CurrentContext = contextName

	return clientcmd.WriteToFile(*existingConfig, kubeconfigPath)
}

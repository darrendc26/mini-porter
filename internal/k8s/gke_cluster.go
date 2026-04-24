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
	svc, err := container.NewService(ctx,
		option.WithCredentialsFile(credPath),
	)
	if err != nil {
		return err
	}

	parent := fmt.Sprintf("projects/%s/locations/%s", projectID, region)

	cluster := &container.Cluster{
		Name:             name,
		InitialNodeCount: 1,
		Locations: []string{
			region + "-a",
			region + "-b",
			region + "-c",
		},
	}

	req := &container.CreateClusterRequest{
		Cluster: cluster,
	}

	op, err := svc.Projects.Locations.Clusters.Create(parent, req).Do()

	if err != nil {
		if gErr, ok := err.(*googleapi.Error); ok && gErr.Code == 409 {
			fmt.Println("⚠️ Cluster already exists, skipping creation...")
		} else {
			return err
		}
	} else {
		fmt.Println("⏳ Creating cluster (this may take 2–5 minutes)...")

		opName := op.Name
		if !strings.Contains(opName, "projects/") {
			opName = fmt.Sprintf("projects/%s/locations/%s/operations/%s",
				projectID, region, op.Name)
		}

		fmt.Printf("Check status at https://console.cloud.google.com/kubernetes/clusters/details/%s/%s/overview?project=%s",
			region, name, projectID)
		for {
			opStatus, err := svc.Projects.Locations.Operations.Get(opName).Do()
			if err != nil {
				return fmt.Errorf("failed to get operation status: %w", err)
			}

			if opStatus.Status == "DONE" {
				if opStatus.Error != nil {
					return fmt.Errorf("cluster creation failed: %v", opStatus.Error)
				}
				fmt.Println("Cluster creation completed")
				break
			}
			if opStatus.Status == "RUNNING" {
				time.Sleep(10 * time.Second)
			}
		}
	}

	// 🔍 Fetch cluster details (always)
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

	fmt.Println("✅ Kubeconfig merged successfully")
	fmt.Println("🔄 Switched context to:", name)

	time.Sleep(20 * time.Second)

	return nil
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

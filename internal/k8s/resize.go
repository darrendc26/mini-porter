package k8s

import (
	"context"
	"fmt"
	"strings"
	"time"

	"google.golang.org/api/container/v1"
	"google.golang.org/api/option"
)

type Credentials struct {
	Provider    string `json:"provider"`
	Credentials string `json:"credentials"`
}

func ResizeNodePool(ctx context.Context, credPath Credentials, projectID string, zone string, clusterName string, size int64) error {
	credPath = getCredentials()
	svc, err := container.NewService(ctx,
		option.WithCredentialsFile(credPath.Credentials),
	)
	if err != nil {
		return err
	}

	req := &container.SetNodePoolSizeRequest{
		NodeCount: size,
	}

	nodePool := "default-pool"

	op, err := svc.Projects.Locations.Clusters.NodePools.SetSize(
		fmt.Sprintf("projects/%s/locations/%s/clusters/%s/nodePools/%s",
			projectID, zone, clusterName, nodePool),
		req,
	).Do()
	if err != nil {
		return err
	}

	fmt.Println("Resizing node pool...")

	// wait for operation
	for {
		opName := op.Name
		if !strings.Contains(opName, "projects/") {
			opName = fmt.Sprintf(
				"projects/%s/locations/%s/operations/%s",
				projectID,
				zone,
				op.Name,
			)
		}

		opStatus, err := svc.Projects.Locations.Operations.Get(opName).Do()
		if err != nil {
			return err
		}

		if opStatus.Status == "DONE" {
			if opStatus.Error != nil {
				return fmt.Errorf("resize failed: %v", opStatus.Error)
			}
			fmt.Println("Resize complete")
			return nil
		}

		time.Sleep(5 * time.Second)
	}
}

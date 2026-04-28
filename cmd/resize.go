package cmd

import (
	"context"
	"fmt"

	"github.com/darrendc26/mini-porter/internal/k8s"
	"github.com/spf13/cobra"
)

var size int64

var resizeCmd = &cobra.Command{
	Use:   "resize",
	Short: "Resize a node pool",
	Long:  "Resize a node pool to a new size",
	RunE: func(cmd *cobra.Command, args []string) error {
		if projectID == "" {
			return fmt.Errorf("project-id is required")
		}
		if region == "" {
			return fmt.Errorf("region is required")
		}
		if clusterName == "" {
			return fmt.Errorf("cluster name is required")
		}
		// if nodePool == "" {
		// 	return fmt.Errorf("node pool is required")
		// }
		if size == 0 {
			return fmt.Errorf("size is required")
		}

		if credPath == "" {
			credPath, err := getCredentialsPath()
			if err != nil {
				return fmt.Errorf("failed to get credentials path: %w", err)
			}
			if credPath == "" {
				return fmt.Errorf("no credentials found. Run `miniporter login`")
			}
		}

		ctx := context.Background()
		err := k8s.ResizeNodePool(ctx, credPath, projectID, region, clusterName, size)
		if err != nil {
			return fmt.Errorf("failed to resize node pool: %w", err)
		}
		fmt.Println("Successfully resized node pool")
		return nil
	},
}

func init() {
	clusterCmd.AddCommand(resizeCmd)
	resizeCmd.Flags().StringVarP(&projectID, "project-id", "p", "", "Project ID")
	resizeCmd.Flags().StringVarP(&region, "region", "r", "", "Region")
	resizeCmd.Flags().StringVarP(&clusterName, "name", "n", "", "Cluster name")
	resizeCmd.Flags().Int64VarP(&size, "size", "s", 0, "Node pool size")
	resizeCmd.Flags().StringVarP(&credPath, "path", "P", "", "Path to the credentials JSON file")
}

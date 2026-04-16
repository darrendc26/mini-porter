package docker

import (
	"context"
	"fmt"

	"os/exec"
)

func PushDockerImage(ctx context.Context, image string) error {
	fmt.Println("[3/6] Pushing image...")
	cmd := exec.CommandContext(ctx, "docker", "push", image)
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Failed to push image: %w", err)
	}
	fmt.Println("[3/6] Push image				Completed")
	return nil
}

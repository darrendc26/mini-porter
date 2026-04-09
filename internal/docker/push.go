package docker

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

func PushDockerImage(ctx context.Context, image string) error {
	fmt.Println("[2/4] Pushing image...")
	cmd := exec.CommandContext(ctx, "docker", "push", image)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Failed to push image: %w", err)
	}
	return nil
}

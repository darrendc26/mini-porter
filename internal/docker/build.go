package docker

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

func BuildDockerImage(ctx context.Context, image string) error {
	fmt.Println("[1/4] Building image...")
	cmd := exec.CommandContext(ctx, "docker", "build", "-t", image, ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Failed to build image: %w", err)
	}
	return nil
}

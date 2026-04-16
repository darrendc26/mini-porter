package docker

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
)

func BuildDockerImage(ctx context.Context, image string, path string) error {
	fmt.Println("[2/6] Building image...")

	cmd := exec.CommandContext(ctx,
		"docker", "build",
		"-t", image,
		"-f", filepath.Join(path, "Dockerfile"),
		path,
	)
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Failed to build image: %w", err)
	}
	fmt.Println("[2/6] Build image				Completed")
	return nil
}

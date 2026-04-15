package docker

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func BuildDockerImage(ctx context.Context, image string, path string) error {
	fmt.Println("[1/5] Building image...")

	cmd := exec.CommandContext(ctx,
		"docker", "build",
		"-t", image,
		"-f", filepath.Join(path, "Dockerfile"),
		path,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Failed to build image: %w", err)
	}
	return nil
}

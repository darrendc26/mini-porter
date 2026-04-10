package k8s

import (
	"os/exec"
	"strings"
)

func GetMinikubeIP() (string, error) {
	cmd := exec.Command("minikube", "ip")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

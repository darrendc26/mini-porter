package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"mini-porter/internal/config"
	"mini-porter/internal/k8s"

	"github.com/spf13/cobra"
)

var hostCmd = &cobra.Command{
	Use:   "host",
	Short: "Manage host entries",
}

var hostAddcmd = &cobra.Command{
	Use:   "add",
	Short: "Add domain to /etc/hosts",

	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig("mini-porter.yaml")
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}

		ip, err := k8s.GetMinikubeIP()
		if err != nil {
			fmt.Println("Error getting IP:", err)
			return
		}

		entry := fmt.Sprintf("%s %s.miniporter", ip, cfg.Name)

		command := fmt.Sprintf("echo \"%s\" >> /etc/hosts", entry)

		fmt.Println("Adding entry (requires sudo)...")

		c := exec.Command("sudo", "sh", "-c", command)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			fmt.Println("Error adding entry:", err)
			return
		}

		fmt.Println("Entry added successfully!")
	},
}

func init() {
	hostCmd.AddCommand(hostAddcmd)
	rootCmd.AddCommand(hostCmd)
}

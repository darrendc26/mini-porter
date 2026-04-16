package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new project",

	Run: func(cmd *cobra.Command, args []string) {
		_, err := os.Stat("mini-porter.yaml")
		if err == nil {
			fmt.Println("mini-porter.yaml already exists. Run 'mini-porter deploy' to deploy your app.")
			return
		}

		content := `name: my-app # Replace with your app name
image: <your-docker-name>/<your-app-name> # Replace with your app image
replicas: 1 # Replace with number of replicas needed
services:
  - name: backend # Replace with your service name (eg. backend, frontend)
    path: # Add your service path, not relative to mini-porter.yaml
    port: 8080 # Add your service port
`

		err = os.WriteFile("mini-porter.yaml", []byte(content), 0644)
		if err != nil {
			fmt.Printf("Error writing mini-porter.yaml: %v\n", err)
			return
		}

		fmt.Println("mini-porter.yaml created successfully!")
		fmt.Println("Edit the file with your app details, then run 'mini-porter deploy' to deploy your app.")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

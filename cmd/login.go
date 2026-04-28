package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	container "google.golang.org/api/container/v1"
	"google.golang.org/api/option"

	"github.com/spf13/cobra"
)

var credPath string
var provider string

type Credentials struct {
	Provider    string `json:"provider"`
	Credentials string `json:"credentials"`
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to a cloud provider",
	Long:  "Login to a cloud provider to deploy applications",
	Run: func(cmd *cobra.Command, args []string) {
		if credPath == "" {
			fmt.Println("Please provide --path to credentials JSON")
			return
		}

		ctx := context.Background()
		client, err := container.NewService(ctx,
			option.WithAuthCredentialsFile("service-account", credPath),
		)
		if err != nil {
			fmt.Printf("Error creating GCP client: %v\n", err)
			return
		}
		err = saveCredentials(Credentials{
			Provider:    provider,
			Credentials: credPath,
		})

		if err != nil {
			fmt.Printf("Error saving credentials: %v\n", err)
			return
		}
		fmt.Printf("Successfully authenticated with GCP. Client: %v\n", client)
		fmt.Println("Successfully logged in to GCP")
	},
}

func init() {
	loginCmd.Flags().StringVarP(&credPath, "path", "p", "", "Path to the credentials JSON file")
	loginCmd.Flags().StringVarP(&provider, "provider", "P", "", "Project ID")
	rootCmd.AddCommand(loginCmd)
}

func saveCredentials(creds Credentials) error {
	path, err := getCredentialsPath()
	if err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(creds)
	if err != nil {
		return fmt.Errorf("error encoding JSON: %v", err)
	}
	fmt.Printf("Credentials saved to %s\n", path)
	return nil
}

// func loadCredentials() (Credentials, error) {
// 	path, err := getCredentialsPath()
// 	if err != nil {
// 		return Credentials{}, err
// 	}
// 	file, err := os.Open(path)
// 	if err != nil {
// 		return Credentials{}, fmt.Errorf("error opening file: %v", err)
// 	}
// 	defer file.Close()
// 	decoder := json.NewDecoder(file)
// 	var creds Credentials
// 	err = decoder.Decode(&creds)
// 	if err != nil {
// 		return Credentials{}, fmt.Errorf("error decoding JSON: %v", err)
// 	}
// 	return creds, nil
// }

func getCredentialsPath() (string, error) {
	path, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error getting home directory: %v", err)
	}
	path = filepath.Join(path, ".mini-porter")
	err = os.MkdirAll(path, 0755)
	if err != nil {
		return "", fmt.Errorf("error creating directory: %v", err)
	}
	path = filepath.Join(path, "credentials.json")
	return path, nil
}

func getCredentials() (Credentials, error) {
	path, err := getCredentialsPath()
	if err != nil {
		return Credentials{}, err
	}
	file, err := os.Open(path)
	if err != nil {
		return Credentials{}, fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	var creds Credentials
	err = decoder.Decode(&creds)
	if err != nil {
		return Credentials{}, fmt.Errorf("error decoding JSON: %v", err)
	}
	return creds, nil
}

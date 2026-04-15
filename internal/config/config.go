package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type Service struct {
	Name string `yaml:"name"`
	Path string `yaml:"path"`
	Port int    `yaml:"port"`
	// Replicas int    `yaml:"replicas"`
}

type Config struct {
	Name     string    `yaml:"name"`
	Image    string    `yaml:"image"`
	Replicas int       `yaml:"replicas"`
	Services []Service `yaml:"services"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Failed to read: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("Failed to unmarshal: %w", err)
	}

	return &config, nil
}

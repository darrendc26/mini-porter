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
}

type Dependency struct {
	Name string            `yaml:"name"`
	Type string            `yaml:"type"`
	Port int               `yaml:"port"`
	Env  map[string]string `yaml:"env"`
}

type Config struct {
	Name         string       `yaml:"name"`
	Image        string       `yaml:"image"`
	Replicas     int          `yaml:"replicas"`
	Services     []Service    `yaml:"services"`
	Dependencies []Dependency `yaml:"dependencies"`
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

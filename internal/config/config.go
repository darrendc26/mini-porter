package config

import (
	"fmt"
	"os"
	"path/filepath"

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
	Domain       string       `yaml:"domain,omitempty"`
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

	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve config path: %w", err)
	}

	baseDir := filepath.Dir(absPath)

	for i := range config.Services {
		if !filepath.IsAbs(config.Services[i].Path) {
			config.Services[i].Path = filepath.Clean(
				filepath.Join(baseDir, config.Services[i].Path),
			)
		}
	}

	return &config, nil
}

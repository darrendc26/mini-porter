package detector

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Project struct {
	Path string
	Type string
}

var ignoreDirs = []string{
	".git",
	"node_modules",
	"example",
	// "examples",
	"test",
	"tests",
	"vendor",
}

func DetectProjects(root string) ([]Project, error) {
	var projects []Project

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if shouldIgnore(path) {
			return nil
		}

		switch d.Name() {
		case "package.json":
			projects = append(projects, Project{Path: filepath.Dir(path), Type: "nodejs"})
		case "go.mod":
			projects = append(projects, Project{Path: filepath.Dir(path), Type: "golang"})
		case "pom.xml":
			projects = append(projects, Project{Path: filepath.Dir(path), Type: "java"})
		case "requirements.txt":
			projects = append(projects, Project{Path: filepath.Dir(path), Type: "python"})
		case "Cargo.toml":
			projects = append(projects, Project{Path: filepath.Dir(path), Type: "rust"})
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	if len(projects) == 0 {
		return nil, fmt.Errorf("no projects detected")
	}

	return projects, nil
}

func shouldIgnore(path string) bool {
	for _, dir := range ignoreDirs {
		if strings.Contains(path, "/"+dir+"/") || strings.HasSuffix(path, "/"+dir) {
			return true
		}
	}

	return false
}

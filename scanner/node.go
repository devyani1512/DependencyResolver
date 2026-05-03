package scanner

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type PackageJSON struct {
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

func ParseNodeDependencies(projectPath string) ([]Dependency, error) {
	pkgFile := filepath.Join(projectPath, "package.json")
	data, err := os.ReadFile(pkgFile)
	if err != nil {
		return nil, err
	}

	var pkg PackageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, err
	}

	var deps []Dependency
	for name, version := range pkg.Dependencies {
		deps = append(deps, Dependency{Name: name, Version: version, Source: "package.json (dependencies)"})
	}
	for name, version := range pkg.DevDependencies {
		deps = append(deps, Dependency{Name: name, Version: version, Source: "package.json (devDependencies)"})
	}
	return deps, nil
}

func DetectNodeServices(projectPath string) []ServiceRequirement {
	var services []ServiceRequirement
	seen := map[string]bool{}

	pkgFile := filepath.Join(projectPath, "package.json")
	data, _ := os.ReadFile(pkgFile)
	var pkg PackageJSON
	json.Unmarshal(data, &pkg)

	allDeps := make(map[string]string)
	for k, v := range pkg.Dependencies {
		allDeps[k] = v
	}
	for k, v := range pkg.DevDependencies {
		allDeps[k] = v
	}

	for dep := range allDeps {
		if !seen["postgres"] && (strings.Contains(dep, "pg") || strings.Contains(dep, "postgres")) {
			seen["postgres"] = true
			services = append(services, ServiceRequirement{
				Name:    "postgres",
				Version: "14",
				Reason:  "Detected PostgreSQL package",
			})
		}
		if !seen["redis"] && strings.Contains(dep, "redis") {
			seen["redis"] = true
			services = append(services, ServiceRequirement{
				Name:    "redis",
				Version: "latest",
				Reason:  "Detected Redis package",
			})
		}
	}
	return services
}

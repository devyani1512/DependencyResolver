package scanner

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// there are 2 types of dependencies
// dependencies and devdependencies
type PackageJSON struct {
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

// this function extracts dependenceies and converts them in a unified form
func ParseNodeDependencies(projectPath string) ([]Dependency, error) {
	pkgFile := filepath.Join(projectPath, "package.json")
	data, err := os.ReadFile(pkgFile)
	if err != nil {
		return nil, err
	}

	var pkg PackageJSON
	//unmarshal converts raw json to go struct
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, err
	}

	var deps []Dependency

	//parse dependencies
	//converts dependencies to internal format
	for name, version := range pkg.Dependencies {
		deps = append(deps, Dependency{
			Name:    name,
			Version: version,
			Source:  "package.json (dependencies)",
		})
	}
	for name, version := range pkg.DevDependencies {
		deps = append(deps, Dependency{
			Name:    name,
			Version: version,
			Source:  "package.json (devDependencies)",
		})
	}
	return deps, nil
}

// detect node services detects required services from node.js project
func DetectNodeServices(projectPath string) []ServiceRequirement {
	services := []ServiceRequirement{}

	//read package.json to detect database dependencies
	pkgFile := filepath.Join(projectPath, "package.json")
	data, _ := os.ReadFile(pkgFile)

	var pkg PackageJSON
	json.Unmarshal(data, &pkg)

	//check dependencies for database packages
	allDeps := make(map[string]string)
	for k, v := range pkg.Dependencies {
		allDeps[k] = v
	}
	for k, v := range pkg.DevDependencies {
		allDeps[k] = v
	}

	//detect postgresql
	for dep := range allDeps {
		if strings.Contains(dep, "pg") || strings.Contains(dep, "postgres") {
			services = append(services, ServiceRequirement{
				Name:    "postgres",
				Version: "14",
				Reason:  "Detected Postgresql package",
			})
			break
		}
	}
	for dep := range allDeps {
		if strings.Contains(dep, "redis") {
			services = append(services, ServiceRequirement{
				Name:    "redis",
				Version: "latest",
				Reason:  "Detected Redis package",
			})
			break
		}
	}
	return services
}

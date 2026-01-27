package scanner

import (
	"fmt"
	"os"
	"path/filepath"
)

// declaring functions
// scanproject - detects project type and dependencies
func ScanProject(path string) (*ProjectInfo, error) {
	info := &ProjectInfo{
		Path:         path,
		Type:         "unknown",
		Dependencies: []Dependency{},
		Services:     []ServiceRequirement{},
	}

	//check for python project
	if fileExists(filepath.Join(path, "requirements.txt")) {
		info.Type = "python"
		deps, err := ParsePythonDependencies(path)
		if err != nil {
			return nil, fmt.Errorf("failted to parse python dependencies: %w", err)
		}
		info.Dependencies = deps
		info.Services = DetectPythonServices(path)
		return info, nil
	}

	//similarly check for node project
	if fileExists(filepath.Join(path, "package.json")) {
		info.Type = "node"
		deps, err := ParseNodeDependencies(path)
		if err != nil {
			return nil, fmt.Errorf("failed to parse node dependencies : %w", err)
		}
		info.Dependencies = deps
		info.Services = DetectNodeServices(path)
		return info, nil
	}
	return nil, fmt.Errorf("unknown project type - no requirements.txt or package.json found")
}
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

package scanner

import (
	"fmt"
	"os"
	"path/filepath"
)

// ScanProject detects project type from an existing manifest.
// Call AutoScanAndGenerate instead — it creates the manifest if missing.
func ScanProject(path string) (*ProjectInfo, error) {
	info := &ProjectInfo{
		Path:         path,
		Type:         "unknown",
		Dependencies: []Dependency{},
		Services:     []ServiceRequirement{},
	}

	if fileExists(filepath.Join(path, "requirements.txt")) {
		info.Type = "python"
		deps, err := ParsePythonDependencies(path)
		if err != nil {
			return nil, fmt.Errorf("failed to parse python dependencies: %w", err)
		}
		info.Dependencies = deps
		info.Services = DetectPythonServices(path)
		return info, nil
	}

	if fileExists(filepath.Join(path, "package.json")) {
		info.Type = "node"
		deps, err := ParseNodeDependencies(path)
		if err != nil {
			return nil, fmt.Errorf("failed to parse node dependencies: %w", err)
		}
		info.Dependencies = deps
		info.Services = DetectNodeServices(path)
		return info, nil
	}

	return nil, fmt.Errorf("no requirements.txt or package.json found in %s", path)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

package scanner

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// ParsePythonDependencies reads requirements.txt and extracts dependencies
func ParsePythonDependencies(projectPath string) ([]Dependency, error) {
	reqFile := filepath.Join(projectPath, "requirements.txt")
	file, err := os.Open(reqFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var deps []Dependency
	sc := bufio.NewScanner(file)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		dep := parsePythonLine(line)
		deps = append(deps, dep)
	}
	return deps, sc.Err()
}

func parsePythonLine(line string) Dependency {
	for _, op := range []string{"==", ">=", "<=", "~=", ">", "<"} {
		if strings.Contains(line, op) {
			parts := strings.SplitN(line, op, 2)
			return Dependency{
				Name:    strings.TrimSpace(parts[0]),
				Version: op + strings.TrimSpace(parts[1]),
				Source:  "requirements.txt",
			}
		}
	}
	return Dependency{
		Name:    strings.TrimSpace(line),
		Version: "*",
		Source:  "requirements.txt",
	}
}

// DetectPythonServices scans .py files for database/service imports
func DetectPythonServices(projectPath string) []ServiceRequirement {
	var services []ServiceRequirement
	seen := map[string]bool{}

	// walk all .py files recursively
	filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".py") {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		text := string(content)

		if !seen["postgres"] && (strings.Contains(text, "psycopg2") || strings.Contains(text, "postgresql") || strings.Contains(text, "postgres")) {
			seen["postgres"] = true
			services = append(services, ServiceRequirement{
				Name:    "postgres",
				Version: "14",
				Reason:  "Detected psycopg2/postgres import",
			})
		}
		if !seen["redis"] && (strings.Contains(text, "import redis") || strings.Contains(text, "from redis")) {
			seen["redis"] = true
			services = append(services, ServiceRequirement{
				Name:    "redis",
				Version: "latest",
				Reason:  "Detected redis import",
			})
		}
		return nil
	})
	return services
}

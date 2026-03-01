package scanner

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// ParsePythonDependencies reads requirements.txt
func ParsePythonDependencies(projectPath string) ([]Dependency, error) {
	reqFile := filepath.Join(projectPath, "requirements.txt")
	file, err := os.Open(reqFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var deps []Dependency
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		dep := parsePythonLine(line)
		deps = append(deps, dep)
	}

	return deps, scanner.Err()
}

func parsePythonLine(line string) Dependency {
	for _, op := range []string{"==", ">=", "<=", "~=", ">", "<"} {
		if strings.Contains(line, op) {
			parts := strings.Split(line, op)
			return Dependency{
				Name:    strings.TrimSpace(parts[0]),
				Version: op + strings.TrimSpace(parts[1]),
				Source:  "requirements.txt",
			}
		}
	}

	return Dependency{
		Name:    strings.TrimSpace(line),
		Version: "",
		Source:  "requirements.txt",
	}
}

// DetectPythonServices detects required services from Python project
func DetectPythonServices(projectPath string) []ServiceRequirement {
	services := []ServiceRequirement{}
	foundPostgres := false
	foundRedis := false

	// Check requirements.txt
	reqFile := filepath.Join(projectPath, "requirements.txt")
	if content, err := os.ReadFile(reqFile); err == nil {
		text := strings.ToLower(string(content))

		if strings.Contains(text, "psycopg") || strings.Contains(text, "postgresql") {
			services = append(services, ServiceRequirement{
				Name:    "postgres",
				Version: "14",
				Reason:  "Found PostgreSQL in requirements.txt",
			})
			foundPostgres = true
		}

		if strings.Contains(text, "redis") {
			services = append(services, ServiceRequirement{
				Name:    "redis",
				Version: "latest",
				Reason:  "Found Redis in requirements.txt",
			})
			foundRedis = true
		}
	}

	// Check Python files
	files, _ := filepath.Glob(filepath.Join(projectPath, "*.py"))
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		text := strings.ToLower(string(content))

		if !foundPostgres && (strings.Contains(text, "psycopg") ||
			strings.Contains(text, "postgresql") ||
			strings.Contains(text, "postgres")) {
			services = append(services, ServiceRequirement{
				Name:    "postgres",
				Version: "14",
				Reason:  "Detected PostgreSQL in code",
			})
			foundPostgres = true
		}

		if !foundRedis && strings.Contains(text, "redis") {
			services = append(services, ServiceRequirement{
				Name:    "redis",
				Version: "latest",
				Reason:  "Detected Redis in code",
			})
			foundRedis = true
		}
	}

	return services
}

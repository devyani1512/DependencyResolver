package docker

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/depsolver/resolver/internal/services"
)

// GenerateComposeFile writes a docker-compose.yml for the given service configs
func GenerateComposeFile(configs []services.ServiceConfig, projectPath string) error {
	var sb strings.Builder
	sb.WriteString("version: '3.8'\n\nservices:\n")

	hasVolumes := false
	var volumeNames []string

	for _, svc := range configs {
		sb.WriteString(fmt.Sprintf("  %s:\n", svc.Name))
		sb.WriteString(fmt.Sprintf("    image: %s\n", svc.Image))

		if len(svc.Ports) > 0 {
			sb.WriteString("    ports:\n")
			for _, p := range svc.Ports {
				sb.WriteString(fmt.Sprintf("      - \"%s\"\n", p))
			}
		}

		if len(svc.Environment) > 0 {
			sb.WriteString("    environment:\n")
			for k, v := range svc.Environment {
				sb.WriteString(fmt.Sprintf("      %s: %s\n", k, v))
			}
		}

		if len(svc.Volumes) > 0 {
			sb.WriteString("    volumes:\n")
			for _, vol := range svc.Volumes {
				sb.WriteString(fmt.Sprintf("      - %s\n", vol))
				// collect named volumes (format: "name:/path")
				parts := strings.SplitN(vol, ":", 2)
				if len(parts) == 2 && !strings.HasPrefix(parts[0], ".") && !strings.HasPrefix(parts[0], "/") {
					hasVolumes = true
					volumeNames = append(volumeNames, parts[0])
				}
			}
		}
	}

	if hasVolumes {
		sb.WriteString("\nvolumes:\n")
		for _, v := range volumeNames {
			sb.WriteString(fmt.Sprintf("  %s:\n", v))
		}
	}

	composePath := filepath.Join(projectPath, "docker-compose.yml")
	if err := os.WriteFile(composePath, []byte(sb.String()), 0644); err != nil {
		return fmt.Errorf("writing docker-compose.yml: %w", err)
	}
	fmt.Printf("  ✓ Written %s\n", composePath)
	return nil
}

// StartServices runs docker compose up -d in the project directory
func StartServices(projectPath string) error {
	// Try "docker compose" (v2) first, then "docker-compose" (v1)
	cmds := [][]string{
		{"docker", "compose", "up", "-d"},
		{"docker-compose", "up", "-d"},
	}

	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = projectPath
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			if args[0] == "docker-compose" {
				return fmt.Errorf("docker compose up failed: %w\n  Is Docker running? Start Docker Desktop and retry.", err)
			}
			continue // try docker-compose v1
		}
		return nil
	}
	return fmt.Errorf("docker not found — install Docker Desktop and ensure it is running")
}

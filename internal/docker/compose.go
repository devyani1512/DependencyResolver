package docker

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/devyani1512/DependencyResolver/internal/services"
)

// this function creates docker-compose.yml
func GenerateComposeFile(serviceConfigs []services.ServiceConfig, outputPath string) error {
	yaml := "version: '3.8'\n\nservices:\n"
	for _, svc := range serviceConfigs {
		yaml += fmt.Sprintf("  %s:\n", svc.Name)
		yaml += fmt.Sprintf("    image: %s\n", svc.Image)
		yaml += fmt.Sprintf("    ports:\n")
		yaml += fmt.Sprintf("      - \"%d:%d\"\n", svc.Port, svc.Port)

		if len(svc.Environment) > 0 {
			yaml += "    environment:\n"
			for key, val := range svc.Environment {
				yaml += fmt.Sprintf("      %s: %s\n", key, val)
			}
		}

		yaml += "\n"
	}

	composeFile := filepath.Join(outputPath, "docker-compose.yml")
	return os.WriteFile(composeFile, []byte(yaml), 0644)
}

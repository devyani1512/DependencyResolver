package bootstrap

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/devyani1512/DependencyResolver/internal/conflict"
	"github.com/devyani1512/DependencyResolver/internal/docker"
	"github.com/devyani1512/DependencyResolver/internal/graph"
	"github.com/devyani1512/DependencyResolver/internal/scanner"
	"github.com/devyani1512/DependencyResolver/internal/services"
)

// NeedsBootstrap checks if bootstrap is needed
func NeedsBootstrap(info *scanner.ProjectInfo) bool {
	switch info.Type {
	case "python":
		venvPath := filepath.Join(info.Path, "venv")
		if _, err := os.Stat(venvPath); os.IsNotExist(err) {
			return true
		}

		// Check if requirements.txt changed since last install
		reqFile := filepath.Join(info.Path, "requirements.txt")
		snapshotFile := filepath.Join(venvPath, ".requirements.snapshot")

		reqContent, err1 := os.ReadFile(reqFile)
		snapshotContent, err2 := os.ReadFile(snapshotFile)

		if err1 != nil || err2 != nil {
			return true // Can't read, re-bootstrap to be safe
		}

		if string(reqContent) != string(snapshotContent) {
			fmt.Println("  Requirements changed, need to reinstall...")
			return true
		}

		return false

	case "node":
		nodeModulesPath := filepath.Join(info.Path, "node_modules")
		if _, err := os.Stat(nodeModulesPath); os.IsNotExist(err) {
			return true
		}
		return false
	}

	return true
}

// Bootstrap sets up development environment
func Bootstrap(info *scanner.ProjectInfo) error {
	// Step 1: Build dependency graph
	g := graph.BuildGraph(info)

	// Step 2: Detect conflicts
	conflicts := conflict.DetectConflicts(g)
	if len(conflicts) > 0 {
		return fmt.Errorf("detected %d conflicts", len(conflicts))
	}

	// Step 3: Install dependencies
	if len(info.Dependencies) > 0 {
		fmt.Println("\n Setting up application environment...")
		if err := installDependencies(info); err != nil {
			return fmt.Errorf("failed to install dependencies: %w", err)
		}
	}

	// Step 4: Get required services
	serviceNodes := g.GetServiceNodes()
	if len(serviceNodes) == 0 {
		fmt.Println("\n  No services required")
		return nil
	}

	fmt.Printf("\n Detected %d service(s) needed:\n", len(serviceNodes))
	for _, node := range serviceNodes {
		fmt.Printf("   • %s:%s\n", node.Name, node.Version)
	}

	// Step 5: Generate service configurations
	var serviceConfigs []services.ServiceConfig
	for _, node := range serviceNodes {
		switch node.Name {
		case "postgres":
			serviceConfigs = append(serviceConfigs, services.GetPostgresConfig())
		case "redis":
			serviceConfigs = append(serviceConfigs, services.GetRedisConfig())
		}
	}

	// Step 6: Generate docker-compose.yml
	fmt.Println("\n Generating docker-compose.yml...")
	if err := docker.GenerateComposeFile(serviceConfigs, info.Path); err != nil {
		return fmt.Errorf("failed to generate docker-compose: %w", err)
	}
	fmt.Println("    Created docker-compose.yml")

	// Step 7: Start services
	fmt.Println("\n Starting Docker services...")
	if err := docker.StartServices(info.Path); err != nil {
		return fmt.Errorf("failed to start services: %w", err)
	}
	fmt.Println("    Services started")

	return nil
}

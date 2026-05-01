package bootstrap

import (
	"fmt"

	"github.com/depsolver/resolver/internal/conflict"
	"github.com/depsolver/resolver/internal/docker"
	"github.com/depsolver/resolver/internal/graph"
	"github.com/depsolver/resolver/internal/scanner"
	"github.com/depsolver/resolver/internal/services"
)

// bootstrap sets up development enviroment
func Bootstrap(info *scanner.ProjectInfo) error {
	//1. build dependency graph
	g := graph.BuildGraph(info)

	//2. detect conflicts
	conflicts := conflict.DetectConflicts(g)
	if len(conflicts) > 0 {
		return fmt.Errorf("detected %d conflicts", len(conflicts))
	}
	//3.get required services
	serviceNode := g.GetServiceNodes()
	if len(serviceNode) == 0 {
		fmt.Println("No services required")
		return nil
	}
	//4. generate service configurations
	var serviceConfigs []services.ServiceConfig
	for _, node := range serviceNode {
		switch node.Name {
		case "postgres":
			serviceConfigs = append(serviceConfigs, services.GetPostgresConfig())
		case "redis":
			serviceConfigs = append(serviceConfigs, services.GetRedisConfig())
		}
	}
	//5. generate docker compose.yml
	fmt.Println("Generate docker-compose.yml")
	if err := docker.GenerateComposeFile(serviceConfigs, info.Path); err != nil {
		return fmt.Errorf("failed to generate docker-compose: %w", err)
	}
	//6.start services
	fmt.Println("starting docker services")
	if err := docker.StartServices(info.Path); err != nil {
		return fmt.Errorf("failed to start services: %w", err)
	}
	return nil
}

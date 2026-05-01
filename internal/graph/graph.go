package graph

import "github.com/depsolver/resolver/internal/scanner"

type DependencyGraph struct {
	Nodes map[string]*Node
	Edges map[string][]string
}

// node represents a package in dependency graph
// type - package or service
type Node struct {
	Name    string
	Version string
	Type    string
}

// nodes - all entities- packages and services
// edges - the relation between them
func BuildGraph(info *scanner.ProjectInfo) *DependencyGraph {
	graph := &DependencyGraph{
		Nodes: make(map[string]*Node),
		Edges: make(map[string][]string),
	}

	//add package nodes
	for _, dep := range info.Dependencies {
		graph.Nodes[dep.Name] = &Node{
			Name:    dep.Name,
			Version: dep.Version,
			Type:    "package",
		}
	}
	//add service nodes
	for _, svc := range info.Services {
		graph.Nodes[svc.Name] = &Node{
			Name:    svc.Name,
			Version: svc.Version,
			Type:    "service",
		}
	}
	return graph
}

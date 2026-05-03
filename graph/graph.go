package graph

import "github.com/depsolver/resolver/internal/scanner"

// Node is a node in the simple bootstrap dependency graph
type Node struct {
	Name    string
	Version string
	IsService bool
}

// DependencyGraph is the simple graph used by bootstrap/conflict
type DependencyGraph struct {
	Nodes []*Node
}

// BuildGraph constructs a DependencyGraph from scanner.ProjectInfo
func BuildGraph(info *scanner.ProjectInfo) *DependencyGraph {
	g := &DependencyGraph{}
	for _, dep := range info.Dependencies {
		g.Nodes = append(g.Nodes, &Node{
			Name:    dep.Name,
			Version: dep.Version,
		})
	}
	for _, svc := range info.Services {
		g.Nodes = append(g.Nodes, &Node{
			Name:      svc.Name,
			Version:   svc.Version,
			IsService: true,
		})
	}
	return g
}

// GetServiceNodes returns only service nodes (postgres, redis, etc.)
func (g *DependencyGraph) GetServiceNodes() []*Node {
	var out []*Node
	for _, n := range g.Nodes {
		if n.IsService {
			out = append(out, n)
		}
	}
	return out
}

package app

import (
	"fmt"
	"strings"
)

// DepNode is a node in the dependency graph
type DepNode struct {
	Package      string
	Version      Version
	Dependencies []Dependency
	Dependents   []string // packages that depend on this one
	Depth        int
}

// DepGraph is the full resolved dependency graph
type DepGraph struct {
	Nodes    map[string]*DepNode
	RootDeps []string
}

func NewDepGraph() *DepGraph {
	return &DepGraph{
		Nodes: make(map[string]*DepNode),
	}
}

func (g *DepGraph) AddNode(pkg string, version Version, deps []Dependency, depth int) {
	if _, exists := g.Nodes[strings.ToLower(pkg)]; !exists {
		g.Nodes[strings.ToLower(pkg)] = &DepNode{
			Package:      pkg,
			Version:      version,
			Dependencies: deps,
			Depth:        depth,
		}
	}
}

func (g *DepGraph) AddDependency(from, to string) {
	fromKey := strings.ToLower(from)
	toKey := strings.ToLower(to)
	if node, ok := g.Nodes[toKey]; ok {
		for _, dep := range node.Dependents {
			if dep == fromKey {
				return
			}
		}
		node.Dependents = append(node.Dependents, fromKey)
	}
}

// Print renders the dependency tree visually
func (g *DepGraph) Print() {
	fmt.Println("\nDependency Graph:")
	fmt.Println("─────────────────────────────────────")
	for _, rootPkg := range g.RootDeps {
		g.printNode(rootPkg, "", true, make(map[string]bool))
	}
	fmt.Println("─────────────────────────────────────")
}

func (g *DepGraph) printNode(pkg, prefix string, isLast bool, visited map[string]bool) {
	key := strings.ToLower(pkg)
	node, ok := g.Nodes[key]

	connector := "├── "
	childPrefix := prefix + "│   "
	if isLast {
		connector = "└── "
		childPrefix = prefix + "    "
	}

	if !ok {
		fmt.Printf("%s%s%s (unknown)\n", prefix, connector, pkg)
		return
	}

	marker := ""
	if visited[key] {
		marker = " (↑ see above)"
	}
	fmt.Printf("%s%s%s==%s%s\n", prefix, connector, node.Package, node.Version, marker)

	if visited[key] {
		return
	}
	visited[key] = true

	for i, dep := range node.Dependencies {
		isLastDep := i == len(node.Dependencies)-1
		g.printNode(dep.Package, childPrefix, isLastDep, visited)
	}
}

// AllPackages returns all packages in topological order
func (g *DepGraph) AllPackages() []string {
	visited := make(map[string]bool)
	var order []string

	var visit func(pkg string)
	visit = func(pkg string) {
		key := strings.ToLower(pkg)
		if visited[key] {
			return
		}
		visited[key] = true
		if node, ok := g.Nodes[key]; ok {
			for _, dep := range node.Dependencies {
				visit(dep.Package)
			}
		}
		order = append(order, pkg)
	}

	for _, root := range g.RootDeps {
		visit(root)
	}
	return order
}

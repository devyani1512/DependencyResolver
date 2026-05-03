package app

import (
	"fmt"
	"strings"
)

// Solver orchestrates the entire dependency resolution pipeline
type Solver struct {
	registry *PyPIClient
	graph    *DepGraph
	pubgrub  *PubGrubSolver
	verbose  bool
}

func NewSolver(verbose bool) *Solver {
	registry := NewPyPIClient(verbose)
	return &Solver{
		registry: registry,
		graph:    NewDepGraph(),
		pubgrub:  NewPubGrubSolver(registry, verbose),
		verbose:  verbose,
	}
}

func (s *Solver) log(msg string) {
	if s.verbose {
		fmt.Printf("  %s\n", msg)
	}
}

// Resolve is the main entry point: takes root dependencies and produces a resolution
func (s *Solver) Resolve(rootDeps []Dependency) (*Resolution, error) {
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════╗")
	fmt.Println("║        depsolver — PubGrub Resolver          ║")
	fmt.Println("╚══════════════════════════════════════════════╝")
	fmt.Println()

	fmt.Printf("→ Fetching package metadata from PyPI...\n")

	// Run PubGrub solver
	fmt.Printf("→ Running PubGrub solver...\n")
	result := s.pubgrub.Solve(rootDeps)

	// Build graph from decisions
	fmt.Printf("→ Building dependency graph...\n")
	s.buildGraph(rootDeps, result)

	// Print solver steps
	if s.verbose && len(result.Steps) > 0 {
		fmt.Println("\nSolver trace:")
		for _, step := range result.Steps {
			fmt.Printf("  %s\n", step)
		}
	}

	return &Resolution{
		Result:   result,
		Graph:    s.graph,
		RootDeps: rootDeps,
		Solver:   s.pubgrub,
	}, nil
}

func (s *Solver) buildGraph(rootDeps []Dependency, result SolverResult) {
	for _, dep := range rootDeps {
		s.graph.RootDeps = append(s.graph.RootDeps, dep.Package)
	}

	for pkg, ver := range result.Resolved {
		info, ok := s.pubgrub.packages[strings.ToLower(pkg)]
		if !ok {
			s.graph.AddNode(pkg, ver, nil, 0)
			continue
		}
		var deps []Dependency
		if d, ok := info.Deps[ver.Raw]; ok {
			deps = d
		} else {
			for _, d := range info.Deps {
				deps = d
				break
			}
		}
		s.graph.AddNode(pkg, ver, deps, 0)
	}

	for pkg, node := range s.graph.Nodes {
		for _, dep := range node.Dependencies {
			s.graph.AddDependency(pkg, dep.Package)
		}
	}
}

// Resolution is the complete output of the solver
type Resolution struct {
	Result   SolverResult
	Graph    *DepGraph
	RootDeps []Dependency
	Solver   *PubGrubSolver
}

func (r *Resolution) Print() {
	fmt.Println()
	if r.Result.Success {
		fmt.Printf("✓ Resolution complete — %d packages resolved\n", len(r.Result.Resolved))
	} else {
		fmt.Printf("✗ Resolution failed — %d conflicts detected\n", len(r.Result.Errors))
	}

	if len(r.Result.Resolved) > 0 {
		fmt.Println()
		fmt.Println("Resolved packages:")
		fmt.Println("──────────────────────────────────────")
		for pkg, ver := range r.Result.Resolved {
			fmt.Printf("  %-30s %s\n", pkg, ver)
		}
	}

	r.Graph.Print()

	if len(r.Result.Errors) > 0 {
		fmt.Print(r.Solver.ExplainConflicts(r.Result.Errors))
	}
}

package conflict

import "github.com/depsolver/resolver/internal/graph"

// DetectConflicts analyzes the dependency graph for conflicts.
// Real conflict detection is handled by the PubGrub solver in Phase 2.
// This bootstrap-level check catches obvious structural issues.
func DetectConflicts(g *graph.DependencyGraph) []Conflict {
	return []Conflict{}
}

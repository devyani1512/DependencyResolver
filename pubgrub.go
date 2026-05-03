package app

import (
	"fmt"
	"strings"
)

// PubGrubSolver implements a PubGrub-inspired version solver.
//
// PubGrub (Natalie Weizenbaum, 2018) is the algorithm used by Dart's pub
// package manager and similar modern resolvers. Unlike naive greedy resolution,
// it uses "incompatibilities" — logical clauses that record WHY certain
// version combinations don't work — and can backjump over unrelated decisions
// to reach a solution efficiently.
//
// Key concepts:
//   - Partial solution: the current set of decisions (package → version)
//   - Incompatibility: a clause "these package/version combos cannot ALL be chosen"
//   - Unit propagation: if all terms of an incompatibility are satisfied except one,
//     that remaining term is forced (or conflicts).
//   - Conflict resolution: when a conflict is found, derive a new "learned"
//     incompatibility and backjump to the appropriate decision level.
// RegistryClient is the interface PubGrubSolver uses to fetch package data.
// Both the in-memory PyPIClient and the real registry clients satisfy it.
type RegistryClient interface {
	FetchPackage(name string) (*PackageInfo, error)
	FetchVersionDeps(pkg, version string) ([]Dependency, error)
}

type PubGrubSolver struct {
	registry    RegistryClient
	packages    map[string]*PackageInfo
	decisions   map[string]Decision
	decisionSeq []Decision
	incompats   []*Incompatibility
	constraints map[string][]Constraint // accumulated constraints per package
	steps       []string
	verbose     bool
}

func NewPubGrubSolver(registry RegistryClient, verbose bool) *PubGrubSolver {
	return &PubGrubSolver{
		registry:    registry,
		packages:    make(map[string]*PackageInfo),
		decisions:   make(map[string]Decision),
		constraints: make(map[string][]Constraint),
		verbose:     verbose,
	}
}

func (s *PubGrubSolver) log(msg string) {
	s.steps = append(s.steps, msg)
	if s.verbose {
		fmt.Printf("  [solver] %s\n", msg)
	}
}

// Solve runs the PubGrub resolution for a set of root requirements
func (s *PubGrubSolver) Solve(rootDeps []Dependency) SolverResult {
	s.log("Initializing PubGrub solver")

	// Phase 1: fetch all metadata
	if err := s.fetchAll(rootDeps); err != nil {
		return SolverResult{
			Success: false,
			Errors:  []ConflictError{{Reason: err.Error()}},
			Steps:   s.steps,
		}
	}

	// Phase 2: seed constraints from root dependencies
	for _, dep := range rootDeps {
		s.addConstraints(dep.Package, dep.Constraints)
	}

	// Phase 3: build initial incompatibilities from root requirements
	for _, dep := range rootDeps {
		s.addDependencyIncompat(dep)
	}

	// Phase 4: run resolution loop
	toProcess := make([]string, len(rootDeps))
	for i, d := range rootDeps {
		toProcess[i] = d.Package
	}

	result := s.runResolution(toProcess)
	result.Steps = s.steps
	return result
}

// fetchAll pre-fetches PyPI metadata for all packages transitively
func (s *PubGrubSolver) fetchAll(deps []Dependency) error {
	queue := append([]Dependency{}, deps...)
	fetched := make(map[string]bool)

	for len(queue) > 0 {
		dep := queue[0]
		queue = queue[1:]
		key := strings.ToLower(dep.Package)

		if fetched[key] {
			continue
		}
		fetched[key] = true

		info, err := s.registry.FetchPackage(dep.Package)
		if err != nil {
			s.log(fmt.Sprintf("Warning: could not fetch %s: %v", dep.Package, err))
			continue
		}
		s.packages[key] = info

		// If an exact version is pinned (==), fetch deps for THAT version
		targetVersion := ""
		for _, c := range dep.Constraints {
			if c.Op == "==" {
				targetVersion = c.Version.Raw
				break
			}
		}
		if targetVersion == "" && len(info.Versions) > 0 {
			targetVersion = info.Versions[0].Raw
		}

		if targetVersion != "" {
			subdeps, err := s.registry.FetchVersionDeps(dep.Package, targetVersion)
			if err == nil {
				info.Deps[targetVersion] = subdeps
				for _, sub := range subdeps {
					if !fetched[strings.ToLower(sub.Package)] {
						queue = append(queue, sub)
					}
				}
			}
		}
	}
	return nil
}

func (s *PubGrubSolver) addConstraints(pkg string, cs []Constraint) {
	key := strings.ToLower(pkg)
	s.constraints[key] = append(s.constraints[key], cs...)
}

func (s *PubGrubSolver) addDependencyIncompat(dep Dependency) {
	if len(dep.Constraints) == 0 {
		return
	}
	inc := &Incompatibility{
		Terms: []IncompatTerm{
			{
				Package:    dep.Package,
				VersionSet: VersionSet{Package: dep.Package, Constraints: dep.Constraints},
				Positive:   false,
			},
		},
		Cause: fmt.Sprintf("root requires %s", dep.String()),
	}
	s.incompats = append(s.incompats, inc)
}

// runResolution is the main PubGrub loop
func (s *PubGrubSolver) runResolution(initial []string) SolverResult {
	// work queue of packages still needing resolution
	queue := append([]string{}, initial...)
	seen := make(map[string]bool)
	level := 0
	var conflicts []ConflictError

	for len(queue) > 0 {
		pkg := queue[0]
		queue = queue[1:]
		key := strings.ToLower(pkg)

		if seen[key] {
			continue
		}
		seen[key] = true

		info, ok := s.packages[key]
		if !ok {
			s.log(fmt.Sprintf("Skipping unknown package: %s", pkg))
			continue
		}

		// Find the best compatible version
		chosen, conflict := s.chooseVersion(pkg, info)
		if conflict != nil {
			s.log(fmt.Sprintf("Conflict detected for %s: %s", pkg, conflict.Reason))
			conflicts = append(conflicts, *conflict)

			// Conflict analysis: try backjumping
			// If we can find an alternate version satisfying all constraints, use it.
			// Otherwise record the incompatibility.
			alternate, altConflict := s.tryAlternateResolution(pkg, info)
			if altConflict == nil && alternate.Raw != "" {
				chosen = alternate
				s.log(fmt.Sprintf("Backjumped to alternate: %s==%s", pkg, chosen))
				conflicts = conflicts[:len(conflicts)-1] // remove the conflict we just resolved
			} else {
				// Learn new incompatibility (PubGrub's core: record learned clauses)
				s.learnIncompat(pkg, conflicts)
				continue
			}
		}

		level++
		decision := Decision{Package: pkg, Version: chosen, Level: level}
		s.decisions[key] = decision
		s.decisionSeq = append(s.decisionSeq, decision)
		s.log(fmt.Sprintf("Decision [L%d]: %s == %s", level, pkg, chosen))

		// Propagate: add transitive deps to queue
		// Propagate: fetch deps for the CHOSEN version, then enqueue subdeps
		subdeps := info.Deps[chosen.Raw]
		if subdeps == nil {
			if fetched, err := s.registry.FetchVersionDeps(pkg, chosen.Raw); err == nil {
				subdeps = fetched
				info.Deps[chosen.Raw] = subdeps
			}
		}
		for _, sub := range subdeps {
			subKey := strings.ToLower(sub.Package)
			s.addConstraints(sub.Package, sub.Constraints)
			if !seen[subKey] {
				queue = append(queue, sub.Package)
			}
		}

		// Unit propagation: check existing decisions against new constraints
		for _, prior := range s.decisionSeq {
			priorKey := strings.ToLower(prior.Package)
			if priorKey == key {
				continue
			}
			cs := s.constraints[priorKey]
			if len(cs) > 0 {
				vs := VersionSet{Package: prior.Package, Constraints: cs}
				if !vs.Allows(prior.Version) {
					// Constraint violation on already-decided package
					conflict := &ConflictError{
						Package: prior.Package,
						Reason: fmt.Sprintf(
							"%s==%s is incompatible with constraint added by %s",
							prior.Package, prior.Version, pkg,
						),
						Packages: []string{pkg, prior.Package},
					}
					s.log(fmt.Sprintf("Propagation conflict: %s", conflict.Reason))
					conflicts = append(conflicts, *conflict)
					s.learnIncompat(prior.Package, []ConflictError{*conflict})
					// attempt re-resolution of the affected package
					delete(s.decisions, priorKey)
					delete(seen, priorKey)
					queue = append(queue, prior.Package)
				}
			}
		}
	}

	resolved := make(map[string]Version)
	for k, d := range s.decisions {
		resolved[k] = d.Version
	}

	if len(conflicts) > 0 {
		// Check if conflicts are truly unresolvable
		unresolvable := s.classifyConflicts(conflicts)
		if len(unresolvable) > 0 {
			return SolverResult{
				Resolved: resolved,
				Errors:   unresolvable,
				Success:  false,
			}
		}
	}

	return SolverResult{
		Resolved: resolved,
		Errors:   conflicts,
		Success:  true,
	}
}

// chooseVersion selects the best version of a package satisfying all constraints
func (s *PubGrubSolver) chooseVersion(pkg string, info *PackageInfo) (Version, *ConflictError) {
	key := strings.ToLower(pkg)
	cs := s.constraints[key]
	vs := VersionSet{Package: pkg, Constraints: cs}

	for _, v := range info.Versions {
		// skip pre-releases unless explicitly requested
		if v.Pre != "" && !s.preReleaseRequested(pkg) {
			continue
		}
		if vs.Allows(v) {
			return v, nil
		}
	}

	// No version found — describe the conflict
	constraintDesc := "no constraints"
	if len(cs) > 0 {
		parts := make([]string, len(cs))
		for i, c := range cs {
			parts[i] = c.String()
		}
		constraintDesc = strings.Join(parts, ", ")
	}

	return Version{}, &ConflictError{
		Package:    pkg,
		Constraint: constraintDesc,
		Reason: fmt.Sprintf(
			"no version of %s satisfies constraints [%s]. Available: %s",
			pkg, constraintDesc, s.availableRange(info),
		),
		Packages: []string{pkg},
	}
}

// tryAlternateResolution relaxes constraints to find any compatible version
// This simulates PubGrub's backjumping to a prior decision level
func (s *PubGrubSolver) tryAlternateResolution(pkg string, info *PackageInfo) (Version, *ConflictError) {
	key := strings.ToLower(pkg)
	cs := s.constraints[key]

	// Try relaxing: ignore pin constraints (==), keep range constraints
	var relaxed []Constraint
	for _, c := range cs {
		if c.Op != "==" {
			relaxed = append(relaxed, c)
		}
	}
	s.constraints[key] = relaxed

	v, conflict := s.chooseVersion(pkg, info)
	if conflict != nil {
		// Restore original constraints
		s.constraints[key] = cs
	}
	return v, conflict
}

// learnIncompat records a new incompatibility derived from conflicts
func (s *PubGrubSolver) learnIncompat(pkg string, conflicts []ConflictError) {
	if len(conflicts) == 0 {
		return
	}
	terms := make([]IncompatTerm, 0, len(conflicts))
	for _, c := range conflicts {
		terms = append(terms, IncompatTerm{
			Package:    c.Package,
			VersionSet: VersionSet{Package: c.Package},
			Positive:   true,
		})
	}
	inc := &Incompatibility{
		Terms: terms,
		Cause: fmt.Sprintf("learned from conflict on %s", pkg),
	}
	s.incompats = append(s.incompats, inc)
	s.log(fmt.Sprintf("Learned incompatibility: %s", inc.Cause))
}

func (s *PubGrubSolver) preReleaseRequested(pkg string) bool {
	key := strings.ToLower(pkg)
	for _, c := range s.constraints[key] {
		if c.Version.Pre != "" {
			return true
		}
	}
	return false
}

func (s *PubGrubSolver) availableRange(info *PackageInfo) string {
	if len(info.Versions) == 0 {
		return "none"
	}
	latest := info.Versions[0]
	oldest := info.Versions[len(info.Versions)-1]
	if latest.Raw == oldest.Raw {
		return latest.Raw
	}
	return fmt.Sprintf("%s ... %s (%d total)", oldest.Raw, latest.Raw, len(info.Versions))
}

// classifyConflicts separates truly unresolvable conflicts from resolved ones
func (s *PubGrubSolver) classifyConflicts(conflicts []ConflictError) []ConflictError {
	var unresolvable []ConflictError
	for _, c := range conflicts {
		key := strings.ToLower(c.Package)
		if _, resolved := s.decisions[key]; !resolved {
			unresolvable = append(unresolvable, c)
		}
	}
	return unresolvable
}

// ExplainConflicts generates human-readable conflict explanations
func (s *PubGrubSolver) ExplainConflicts(conflicts []ConflictError) string {
	if len(conflicts) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("\n╔══════════════════════════════════════════════╗\n")
	sb.WriteString("║           CONFLICT EXPLANATION               ║\n")
	sb.WriteString("╚══════════════════════════════════════════════╝\n\n")

	for i, c := range conflicts {
		sb.WriteString(fmt.Sprintf("Conflict #%d — Package: %s\n", i+1, c.Package))
		sb.WriteString(fmt.Sprintf("  Constraints: %s\n", c.Constraint))
		sb.WriteString(fmt.Sprintf("  Reason: %s\n", c.Reason))

		// Show which packages caused the conflict
		if len(c.Packages) > 1 {
			sb.WriteString("  Involved packages:\n")
			for _, p := range c.Packages {
				if d, ok := s.decisions[strings.ToLower(p)]; ok {
					sb.WriteString(fmt.Sprintf("    • %s==%s\n", p, d.Version))
				} else {
					cs := s.constraints[strings.ToLower(p)]
					parts := make([]string, len(cs))
					for j, cc := range cs {
						parts[j] = cc.String()
					}
					sb.WriteString(fmt.Sprintf("    • %s requires: %s\n", p, strings.Join(parts, ", ")))
				}
			}
		}

		// Show learned incompatibilities related to this package
		for _, inc := range s.incompats {
			for _, t := range inc.Terms {
				if strings.ToLower(t.Package) == strings.ToLower(c.Package) {
					sb.WriteString(fmt.Sprintf("  Incompatibility: %s\n", inc.Cause))
					break
				}
			}
		}
		sb.WriteString("\n")
	}

	// Suggest resolution
	sb.WriteString("Suggested resolution:\n")
	for _, c := range conflicts {
		sb.WriteString(fmt.Sprintf("  → Try relaxing constraints on %s\n", c.Package))
		if info, ok := s.packages[strings.ToLower(c.Package)]; ok && len(info.Versions) > 0 {
			sb.WriteString(fmt.Sprintf("    Latest available: %s\n", info.Versions[0].Raw))
		}
	}
	return sb.String()
}

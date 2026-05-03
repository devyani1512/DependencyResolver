package resolver

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/depsolver/resolver/app"
	"github.com/depsolver/resolver/internal/scanner"
)

// RunResult is returned by Run
type RunResult struct {
	Success   bool
	Resolved  map[string]string // package → pinned version string
	Conflicts []string
}

// Run is the Phase 2 entry point.
// It picks the right registry (PyPI or npm) based on project type,
// runs the PubGrub solver, then rewrites the manifest with pinned versions.
func Run(info *scanner.ProjectInfo, verbose bool) RunResult {
	fmt.Println()
	fmt.Println("──────────────────────────────────────────────────────")

	// ── Pick solver based on language ────────────────────────
	var solver *app.Solver
	switch info.Type {
	case "python":
		fmt.Println("[Phase 2] Resolving Python versions via PubGrub + PyPI...")
		solver = app.NewSolver(verbose)
	case "node":
		fmt.Println("[Phase 2] Resolving Node versions via PubGrub + npm...")
		solver = app.NewNodeSolver(verbose)
	default:
		fmt.Fprintf(os.Stderr, "  Unknown project type %q — skipping resolution\n", info.Type)
		return RunResult{Success: true, Resolved: map[string]string{}}
	}

	// Convert scanner.Dependency → app.Dependency
	rootDeps := convertDeps(info.Dependencies)
	if len(rootDeps) == 0 {
		fmt.Println("  No dependencies to resolve.")
		return RunResult{Success: true, Resolved: map[string]string{}}
	}

	// Run the solver
	resolution, err := solver.Resolve(rootDeps)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  Solver error: %v\n", err)
		return RunResult{Success: false}
	}

	resolution.Print()

	if !resolution.Result.Success {
		var conflicts []string
		for _, e := range resolution.Result.Errors {
			conflicts = append(conflicts, e.Reason)
		}
		return RunResult{Success: false, Conflicts: conflicts}
	}

	// Build pinned map
	pinned := make(map[string]string, len(resolution.Result.Resolved))
	for pkg, ver := range resolution.Result.Resolved {
		pinned[pkg] = ver.Raw
	}

	// Rewrite manifest with pinned versions
	if err := rewriteManifest(info, pinned); err != nil {
		fmt.Fprintf(os.Stderr, "  Warning: could not rewrite manifest: %v\n", err)
	}

	// Write lockfile
	lockPath := filepath.Join(info.Path, "depsolver.lock")
	if err := app.GenerateLockfile(resolution.Result, lockPath); err != nil {
		fmt.Fprintf(os.Stderr, "  Warning: could not write lockfile: %v\n", err)
	} else {
		fmt.Printf("  ✓ Lockfile written: %s\n", lockPath)
	}

	return RunResult{Success: true, Resolved: pinned}
}

// convertDeps converts scanner.Dependency (plain strings) → app.Dependency (parsed semver)
func convertDeps(scannerDeps []scanner.Dependency) []app.Dependency {
	var out []app.Dependency
	for _, d := range scannerDeps {
		out = append(out, app.Dependency{
			Package:     d.Name,
			Constraints: app.ParseConstraintStr(d.Version),
		})
	}
	return out
}

// rewriteManifest overwrites requirements.txt or package.json with pinned versions
func rewriteManifest(info *scanner.ProjectInfo, pinned map[string]string) error {
	switch info.Type {
	case "python":
		return rewriteRequirements(info.Path, pinned)
	case "node":
		return rewritePackageJSON(info.Path, pinned)
	}
	return nil
}

func rewriteRequirements(projectPath string, pinned map[string]string) error {
	var sb strings.Builder
	sb.WriteString("# Resolved by depsolver Phase 2 (PubGrub)\n")
	for pkg, ver := range pinned {
		if ver == "" || ver == "*" {
			sb.WriteString(fmt.Sprintf("%s\n", pkg))
		} else {
			sb.WriteString(fmt.Sprintf("%s==%s\n", pkg, ver))
		}
	}
	reqPath := filepath.Join(projectPath, "requirements.txt")
	if err := os.WriteFile(reqPath, []byte(sb.String()), 0644); err != nil {
		return err
	}
	fmt.Printf("  ✓ Rewrote %s with %d pinned packages\n", reqPath, len(pinned))
	return nil
}

func rewritePackageJSON(projectPath string, pinned map[string]string) error {
	pkgPath := filepath.Join(projectPath, "package.json")

	// preserve existing fields (name, version, scripts etc.)
	existing := map[string]interface{}{}
	if data, err := os.ReadFile(pkgPath); err == nil {
		json.Unmarshal(data, &existing)
	}

	deps := make(map[string]string, len(pinned))
	for pkg, ver := range pinned {
		if ver == "" || ver == "*" {
			deps[pkg] = "*"
		} else {
			deps[pkg] = ver
		}
	}
	existing["dependencies"] = deps

	data, err := json.MarshalIndent(existing, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(pkgPath, data, 0644); err != nil {
		return err
	}
	fmt.Printf("  ✓ Rewrote %s with %d pinned packages\n", pkgPath, len(pinned))
	return nil
}

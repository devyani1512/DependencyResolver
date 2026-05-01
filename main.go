package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/depsolver/resolver/app"
	"github.com/depsolver/resolver/internal/bootstrap"
	"github.com/depsolver/resolver/internal/scanner"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	// ── Resolver commands ────────────────────────────────────────────────────
	case "resolve":
		if len(os.Args) < 3 {
			fmt.Println("Usage: depsolver resolve <requirements.txt>")
			os.Exit(1)
		}
		runResolve(os.Args[2], false)

	case "resolve-verbose":
		if len(os.Args) < 3 {
			fmt.Println("Usage: depsolver resolve-verbose <requirements.txt>")
			os.Exit(1)
		}
		runResolve(os.Args[2], true)

	case "demo":
		runDemo(0)

	case "demo1":
		runDemo(1)

	case "demo2":
		runDemo(2)

	case "demo3":
		runDemo(3)

	// ── Bootstrap commands ───────────────────────────────────────────────────
	case "bootstrap":
		if len(os.Args) < 3 {
			fmt.Println("Usage: depsolver bootstrap <project-path>")
			os.Exit(1)
		}
		runBootstrap(os.Args[2])

	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

// ── Resolver ─────────────────────────────────────────────────────────────────

func runResolve(filename string, verbose bool) {
	deps, err := parseRequirementsFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", filename, err)
		os.Exit(1)
	}

	if len(deps) == 0 {
		fmt.Println("No dependencies found in requirements file.")
		os.Exit(0)
	}

	fmt.Printf("Resolving %d root dependencies from %s\n", len(deps), filename)

	solver := app.NewSolver(verbose)
	resolution, err := solver.Resolve(deps)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Resolution error: %v\n", err)
		os.Exit(1)
	}

	resolution.Print()

	if resolution.Result.Success {
		app.PrintLockfile(resolution.Result)
		if err := app.GenerateLockfile(resolution.Result, "depsolver.lock"); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not write lockfile: %v\n", err)
		} else {
			fmt.Println("\n✓ Lockfile written to depsolver.lock")
		}
	}
}

func parseRequirementsFile(path string) ([]app.Dependency, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var deps []app.Dependency
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		pkg, constraintStr := app.ParseRequirementsLine(line)
		if pkg == "" {
			continue
		}
		deps = append(deps, app.Dependency{
			Package:     pkg,
			Constraints: app.ParseConstraintStr(constraintStr),
		})
	}
	return deps, sc.Err()
}

func runDemo(n int) {
	scenarios := []struct {
		name  string
		desc  string
		lines []string
	}{
		{
			name:  "Demo 1 — Simple Dependencies",
			desc:  "Resolves flask and numpy with all transitive dependencies",
			lines: []string{"flask", "numpy"},
		},
		{
			name:  "Demo 2 — Transitive Dependencies",
			desc:  "Resolves flask and psycopg2-binary, expanding all subdependencies",
			lines: []string{"flask", "psycopg2-binary"},
		},
		{
			name:  "Demo 3 — Conflict Detection",
			desc:  "Detects conflict between requests==2.25 and urllib3>=2",
			lines: []string{"requests==2.25.0", "urllib3>=2"},
		},
	}

	if n == 0 {
		for i, sc := range scenarios {
			fmt.Printf("\n══════════════════════════════════════════════\n")
			fmt.Printf("  %s\n", sc.name)
			fmt.Printf("  %s\n", sc.desc)
			fmt.Printf("══════════════════════════════════════════════\n")
			runDemoScenario(i+1, sc.lines)
		}
		return
	}

	if n < 1 || n > len(scenarios) {
		fmt.Printf("Demo %d not found. Available: 1, 2, 3\n", n)
		return
	}
	sc := scenarios[n-1]
	fmt.Printf("\n══════════════════════════════════════════════\n")
	fmt.Printf("  %s\n", sc.name)
	fmt.Printf("  %s\n", sc.desc)
	fmt.Printf("══════════════════════════════════════════════\n")
	runDemoScenario(n, sc.lines)
}

func runDemoScenario(n int, lines []string) {
	var deps []app.Dependency
	for _, line := range lines {
		pkg, constraintStr := app.ParseRequirementsLine(line)
		if pkg == "" {
			continue
		}
		deps = append(deps, app.Dependency{
			Package:     pkg,
			Constraints: app.ParseConstraintStr(constraintStr),
		})
	}

	solver := app.NewSolver(true)
	resolution, err := solver.Resolve(deps)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}
	resolution.Print()

	if resolution.Result.Success {
		app.PrintLockfile(resolution.Result)
		lockPath := fmt.Sprintf("demo%d.lock", n)
		if err := app.GenerateLockfile(resolution.Result, lockPath); err == nil {
			fmt.Printf("\n✓ Lockfile written to %s\n", lockPath)
		}
	}
}

// ── Bootstrap ────────────────────────────────────────────────────────────────

func runBootstrap(projectPath string) {
	fmt.Println("Dependency Resolver - Development Environment Bootstrap Tool")
	fmt.Printf("Scanning project: %s\n", projectPath)

	projectInfo, err := scanner.ScanProject(projectPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning the project: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Detected: %s project\n", projectInfo.Type)
	fmt.Printf("Found %d dependencies\n", len(projectInfo.Dependencies))

	fmt.Println("\nBootstrapping development environment...")
	if err := bootstrap.Bootstrap(projectInfo); err != nil {
		fmt.Fprintf(os.Stderr, "Bootstrap failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nEnvironment ready! You can start development.")
	fmt.Println("\nTo stop the environment:")
	fmt.Println("  docker-compose down")
}

// ── Help ─────────────────────────────────────────────────────────────────────

func printUsage() {
	fmt.Println(`
depsolver — PubGrub-style Python dependency resolver + dev environment bootstrap

Resolver commands:
  resolve <requirements.txt>          Resolve dependencies
  resolve-verbose <requirements.txt>  Resolve with solver trace
  demo1                               Demo: simple deps
  demo2                               Demo: transitive deps
  demo3                               Demo: conflict detection
  demo                                Run all demos

Bootstrap commands:
  bootstrap <project-path>            Scan project and bootstrap dev environment

Examples:
  go run . resolve requirements.txt
  go run . resolve-verbose requirements.txt
  go run . demo3
  go run . bootstrap ./examples/python-app
`)
}

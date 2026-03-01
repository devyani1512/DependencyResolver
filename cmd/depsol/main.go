package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/devyani1512/DependencyResolver/internal/bootstrap"
	"github.com/devyani1512/DependencyResolver/internal/scanner"
)

func main() {

	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "run":
			runCommand()
			return
		case "watch":
			watchCommand()
			return
		case "help", "-h", "--help":
			printHelp()
			return
		}
	}

	bootstrapCommand()
}

func printHelp() {
	fmt.Println(` DependencyResolver - Smart Development Environment Tool

USAGE:
  depsol <project-path>           Bootstrap environment
  depsol run <project-path>       Run application (auto-bootstrap if needed)
  depsol watch <project-path>     Watch mode with auto-reload
  depsol help                     Show this help

EXAMPLES:
  depsol ./my-python-app          Create venv, install deps, start Docker
  depsol run ./my-python-app      Auto-detect imports and run
  depsol watch ./my-python-app    Watch for changes and auto-reload

FEATURES:
  ✓ Auto-detects Python/Node.js projects
  ✓ Scans code for imports (no manual requirements.txt!)
  ✓ Creates virtual environments automatically
  ✓ Detects and starts required services (PostgreSQL, Redis, etc.)
  ✓ No manual activation needed - just run!`)
}

func bootstrapCommand() {
	fmt.Println(" DependencyResolver - Development Environment Bootstrap Tool")
	fmt.Println("")

	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	projectPath := os.Args[1]

	// Step 1: Scan project
	fmt.Printf("\n Scanning project: %s\n", projectPath)
	projectInfo, err := scanner.ScanProject(projectPath)
	if err != nil {
		fmt.Printf(" Error scanning project: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf(" Detected: %s project\n", projectInfo.Type)
	fmt.Printf(" Found %d dependencies\n", len(projectInfo.Dependencies))

	if len(projectInfo.Dependencies) > 0 {
		fmt.Println("\n Dependencies:")
		for _, dep := range projectInfo.Dependencies {
			fmt.Printf("   • %s %s (from %s)\n", dep.Name, dep.Version, dep.Source)
		}
	}

	// Step 2: Bootstrap environment
	fmt.Println("\n Bootstrapping development environment...")
	if err := bootstrap.Bootstrap(projectInfo); err != nil {
		fmt.Printf(" Bootstrap failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n Environment ready! You can start developing.")
	fmt.Println("\n Next steps:")
	fmt.Printf("   depsol run %s       Run your app\n", projectPath)
	fmt.Println("   docker-compose down        Stop services")
}

func runCommand() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: depsol run <project-path>")
		fmt.Println("Example: depsol run ./examples/python-app")
		os.Exit(1)
	}

	projectPath := os.Args[2]
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		fmt.Printf(" Invalid path: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(" Smart Runner - Auto-Bootstrap & Run")
	fmt.Println("")

	// Step 1: Scan project
	fmt.Printf("\n Scanning: %s\n", absPath)
	projectInfo, err := scanner.ScanProject(absPath)
	if err != nil {
		fmt.Printf(" Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf(" Project type: %s\n", projectInfo.Type)

	// Step 2: Check if bootstrap needed
	needsBootstrap := bootstrap.NeedsBootstrap(projectInfo)

	if needsBootstrap {
		fmt.Println("\n  Bootstrapping environment...")
		if err := bootstrap.Bootstrap(projectInfo); err != nil {
			fmt.Printf(" Bootstrap failed: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Println("\n Environment up to date")
	}

	// Step 3: Run the app
	fmt.Println("\n Starting your application...")
	fmt.Println("")

	var cmd *exec.Cmd

	switch projectInfo.Type {
	case "python":
		venvPython := filepath.Join(absPath, "venv", "bin", "python")
		appFile := filepath.Join(absPath, "app.py")
		cmd = exec.Command(venvPython, appFile)

	case "node":
		cmd = exec.Command("npm", "start")

	default:
		fmt.Println(" Unsupported project type")
		os.Exit(1)
	}

	cmd.Dir = absPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		fmt.Printf("\n App exited with error: %v\n", err)
		os.Exit(1)
	}
}

func watchCommand() {
	fmt.Println(" Watch mode to be integrated later!")
	fmt.Println("For now, use: depsol run <project-path>")
}

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/devyani1512/DependencyResolver/internal/bootstrap"
	"github.com/devyani1512/DependencyResolver/internal/docker"
	"github.com/devyani1512/DependencyResolver/internal/graph"
	"github.com/devyani1512/DependencyResolver/internal/scanner"
	"github.com/devyani1512/DependencyResolver/internal/services"
)

func main() {

	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "run":
			runCommand()
			return
		case "dockerize": // NEW COMMAND
			dockerizeCommand()
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

func dockerizeCommand() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: depsol dockerize <project-path>")
		fmt.Println("Example: depsol dockerize ./examples/python-app")
		os.Exit(1)
	}

	projectPath := os.Args[2]
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		fmt.Printf(" Invalid path: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(" Dockerizing Application")
	fmt.Println("")

	// Scan project
	fmt.Printf("\n Scanning: %s\n", absPath)
	projectInfo, err := scanner.ScanProject(absPath)
	if err != nil {
		fmt.Printf(" Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf(" Project type: %s\n", projectInfo.Type)

	// Generate Dockerfile
	fmt.Println("\n Generating Dockerfile...")
	if err := docker.GenerateAppDockerfile(projectInfo); err != nil {
		fmt.Printf(" Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("   Created Dockerfile")

	// Get services
	g := graph.BuildGraph(projectInfo)
	serviceNodes := g.GetServiceNodes()

	var serviceConfigs []services.ServiceConfig
	for _, node := range serviceNodes {
		switch node.Name {
		case "postgres":
			serviceConfigs = append(serviceConfigs, services.GetPostgresConfig())
		case "redis":
			serviceConfigs = append(serviceConfigs, services.GetRedisConfig())
		}
	}

	// Generate full docker-compose.yml
	fmt.Println(" Generating docker-compose.yml...")
	if err := docker.GenerateFullDockerCompose(projectInfo, serviceConfigs); err != nil {
		fmt.Printf(" Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("    Created docker-compose.yml")

	// Generate .dockerignore
	dockerignore := `venv/
node_modules/
__pycache__/
*.pyc
.git/
.env
*.log
.DS_Store
`
	dockerignorePath := filepath.Join(absPath, ".dockerignore")
	if err := os.WriteFile(dockerignorePath, []byte(dockerignore), 0644); err != nil {
		fmt.Printf("  Warning: Could not create .dockerignore: %v\n", err)
	} else {
		fmt.Println("    Created .dockerignore")
	}

	fmt.Println("\n Dockerization complete!")
	fmt.Println("\n To build and run:")
	fmt.Printf("   cd %s\n", absPath)
	fmt.Println("   docker-compose up --build")
	fmt.Println("\n To stop:")
	fmt.Println("   docker-compose down")
	fmt.Println("\n Now your app can run on any machine with just Docker!")
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
func printHelp() {
	fmt.Println(` DependencyResolver - Smart Development Environment Tool

USAGE:
  depsol <project-path>           Bootstrap environment
  depsol run <project-path>       Run application (auto-bootstrap if needed)
  depsol dockerize <project-path> Generate Dockerfile + docker-compose.yml
  depsol help                     Show this help

EXAMPLES:
  depsol ./my-python-app          Create venv, install deps, start Docker
  depsol run ./my-python-app      Auto-detect imports and run
  depsol dockerize ./my-app       Make it Docker-ready

FEATURES:
  ✓ Auto-detects Python/Node.js projects
  ✓ Scans code for imports (no manual requirements.txt!)
  ✓ Creates virtual environments automatically
  ✓ Detects and starts required services (PostgreSQL, Redis, etc.)
  ✓ No manual activation needed - just run!
  ✓ Dockerize your entire stack with one command`)
}

package main

import (
	"fmt"
	"os"

	"github.com/devyani1512/DependencyResolver/internal/bootstrap"
	"github.com/devyani1512/DependencyResolver/internal/scanner"
)

func main() {
	fmt.Println(" Dependency Resolver - Development Environment Bootstrap Tool ")

	if len(os.Args) < 2 {
		fmt.Println("usage Dependency Resolver ")
		fmt.Println("Example: Dependency Resolver ./examples/python-app")
		os.Exit(1)
	}

	projectPath := os.Args[1]
	//scanning the project
	fmt.Printf("/Scanning project: %s\n", projectPath)
	projectInfo, err := scanner.ScanProject(projectPath)
	if err != nil {
		fmt.Printf("Error scanning the project: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Detected: %s project\n", projectInfo.Type)
	fmt.Printf("Found %d Dependencies \n", len(projectInfo.Dependencies))

	//bootstrap the enviroment
	fmt.Println("\n Bootstrapping development enviroment")
	if err := bootstrap.Bootstrap(projectInfo); err != nil {
		fmt.Printf("Bootstrap failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("\n Enviroment ready! You can start development")
	fmt.Println("\n To stop the enviroment:")
	fmt.Println("docker-compose down")
}

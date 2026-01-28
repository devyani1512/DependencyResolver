package bootstrap

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/devyani1512/DependencyResolver/internal/scanner"
)

func installDependencies(info *scanner.ProjectInfo) error {
	switch info.Type {
	case "python":
		return setupPythonEnvironment(info.Path)
	case "node":
		return setupNodeEnvironment(info.Path)
	default:
		return fmt.Errorf("unsupported project type: %s", info.Type)
	}
}

func setupPythonEnvironment(projectPath string) error {
	// CONVERT TO ABSOLUTE PATH FIRST
	absProjectPath, err := filepath.Abs(projectPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	venvPath := filepath.Join(absProjectPath, "venv")

	// Step 1: Check if Python 3 is available
	pythonCmd := findPythonCommand()
	if pythonCmd == "" {
		return fmt.Errorf("python3 or python not found - please install Python 3")
	}
	fmt.Printf("  → Using Python: %s\n", pythonCmd)

	// Step 2: Create virtual environment
	if _, err := os.Stat(venvPath); os.IsNotExist(err) {
		fmt.Println("  → Creating virtual environment...")
		cmd := exec.Command(pythonCmd, "-m", "venv", venvPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to create venv: %w", err)
		}
		fmt.Println("   Virtual environment created")
	} else {
		fmt.Println("    Virtual environment already exists")
	}

	// Step 3: Get pip path (now absolute)
	pipPath := getPipPath(venvPath)

	// Verify pip exists
	if _, err := os.Stat(pipPath); os.IsNotExist(err) {
		return fmt.Errorf("pip not found at %s - venv may be corrupted, try: rm -rf %s", pipPath, venvPath)
	}
	fmt.Printf("  → Found pip at: %s\n", pipPath)

	// Step 4: Upgrade pip
	fmt.Println("  → Upgrading pip...")
	cmd := exec.Command(pipPath, "install", "--upgrade", "pip")
	cmd.Dir = absProjectPath // Use absolute path here too
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Println("    Warning: Could not upgrade pip")
	}

	// Step 5: Install dependencies
	fmt.Println("  → Installing Python dependencies...")
	cmd = exec.Command(pipPath, "install", "-r", "requirements.txt")
	cmd.Dir = absProjectPath // Use absolute path here too
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pip install failed: %w", err)
	}

	fmt.Println("   Python dependencies installed")
	printActivationInstructions(venvPath)

	return nil
}

// Helper: Find Python command (python3 or python)
func findPythonCommand() string {
	// Try python3 first (preferred)
	if _, err := exec.LookPath("python3"); err == nil {
		return "python3"
	}

	// Fall back to python
	if _, err := exec.LookPath("python"); err == nil {
		// Verify it's Python 3
		cmd := exec.Command("python", "--version")
		output, err := cmd.Output()
		if err == nil && strings.Contains(string(output), "Python 3") {
			return "python"
		}
	}

	return ""
}

// Helper: Get correct pip path for platform
func getPipPath(venvPath string) string {
	if runtime.GOOS == "windows" {
		return filepath.Join(venvPath, "Scripts", "pip.exe")
	}
	return filepath.Join(venvPath, "bin", "pip")
}

// Helper: Print activation instructions
func printActivationInstructions(venvPath string) {
	fmt.Println("\n   To activate virtual environment:")
	if runtime.GOOS == "windows" {
		fmt.Printf("     %s\\Scripts\\activate\n", venvPath)
	} else {
		fmt.Printf("     source %s/bin/activate\n", venvPath)
	}
}

func setupNodeEnvironment(projectPath string) error {
	// Check if node_modules exists
	nodeModulesPath := filepath.Join(projectPath, "node_modules")
	if _, err := os.Stat(nodeModulesPath); err == nil {
		fmt.Println("   node_modules already exists")
	}

	// Run npm install
	fmt.Println("  → Installing Node.js dependencies...")
	cmd := exec.Command("npm", "install")
	cmd.Dir = projectPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("npm install failed: %w", err)
	}

	fmt.Println("   Node dependencies installed")
	return nil
}

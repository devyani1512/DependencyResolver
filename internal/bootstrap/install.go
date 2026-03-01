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
	absProjectPath, err := filepath.Abs(projectPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	venvPath := filepath.Join(absProjectPath, "venv")

	pythonCmd := findPythonCommand()
	if pythonCmd == "" {
		return fmt.Errorf("python3 or python not found - please install Python 3")
	}
	fmt.Printf("  → Using Python: %s\n", pythonCmd)

	// Create venv if needed
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
		fmt.Println("    Virtual environment exists")
	}

	venvPython := getVenvPython(venvPath)
	if _, err := os.Stat(venvPython); os.IsNotExist(err) {
		return fmt.Errorf("venv python not found at %s", venvPython)
	}

	// Upgrade pip
	fmt.Println("  → Upgrading pip...")
	cmd := exec.Command(venvPython, "-m", "pip", "install", "--upgrade", "pip", "--quiet")
	cmd.Dir = absProjectPath
	cmd.Run()

	// Install dependencies
	fmt.Println("  → Installing Python dependencies...")
	cmd = exec.Command(venvPython, "-m", "pip", "install", "-r", "requirements.txt")
	cmd.Dir = absProjectPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pip install failed: %w", err)
	}

	fmt.Println("   Python dependencies installed")

	// Save snapshot (NEW!)
	saveRequirementsSnapshot(absProjectPath, venvPath)

	return nil
}
func saveRequirementsSnapshot(projectPath, venvPath string) error {
	reqFile := filepath.Join(projectPath, "requirements.txt")
	snapshotFile := filepath.Join(venvPath, ".requirements.snapshot")

	data, err := os.ReadFile(reqFile)
	if err != nil {
		return err
	}

	return os.WriteFile(snapshotFile, data, 0644)
}

func setupNodeEnvironment(projectPath string) error {
	absProjectPath, err := filepath.Abs(projectPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	fmt.Println("  → Installing Node.js dependencies...")
	cmd := exec.Command("npm", "install")
	cmd.Dir = absProjectPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("npm install failed: %w", err)
	}

	fmt.Println("   Node dependencies installed")
	return nil
}

func findPythonCommand() string {
	// Try python3.12 first (best compatibility)
	if _, err := exec.LookPath("python3.12"); err == nil {
		return "python3.12"
	}
	// Try python3.11
	if _, err := exec.LookPath("python3.11"); err == nil {
		return "python3.11"
	}
	// Try python3
	if _, err := exec.LookPath("python3"); err == nil {
		return "python3"
	}
	// Try python
	if _, err := exec.LookPath("python"); err == nil {
		cmd := exec.Command("python", "--version")
		output, _ := cmd.Output()
		if strings.Contains(string(output), "Python 3") {
			return "python"
		}
	}
	return ""
}

func getVenvPython(venvPath string) string {
	if runtime.GOOS == "windows" {
		return filepath.Join(venvPath, "Scripts", "python.exe")
	}
	return filepath.Join(venvPath, "bin", "python")
}

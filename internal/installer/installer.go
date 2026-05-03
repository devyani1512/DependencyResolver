package installer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Install runs the appropriate package installer for the project type.
// Python → pip install -r requirements.txt
// Node   → npm install
func Install(projectType, projectPath string) error {
	switch projectType {
	case "python":
		return installPython(projectPath)
	case "node":
		return installNode(projectPath)
	default:
		return fmt.Errorf("unknown project type: %s", projectType)
	}
}

func installPython(projectPath string) error {
	reqFile := filepath.Join(projectPath, "requirements.txt")
	if _, err := os.Stat(reqFile); err != nil {
		return fmt.Errorf("requirements.txt not found at %s", reqFile)
	}

	fmt.Println("\n[Phase 3] Installing Python packages...")
	fmt.Println("  Running: pip install -r requirements.txt")

	// try pip3 first, fall back to pip
	pipCmd := "pip3"
	if _, err := exec.LookPath("pip3"); err != nil {
		pipCmd = "pip"
	}

	cmd := exec.Command(pipCmd, "install", "-r", reqFile)
	cmd.Dir = projectPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pip install failed: %w\n  Make sure Python and pip are installed.", err)
	}

	fmt.Println("  ✓ Python packages installed")
	fmt.Println("\n  To run your app:")
	fmt.Println("    python app.py")
	fmt.Println("    OR: flask run")
	fmt.Println("    Then open: http://localhost:5000")
	return nil
}

func installNode(projectPath string) error {
	pkgFile := filepath.Join(projectPath, "package.json")
	if _, err := os.Stat(pkgFile); err != nil {
		return fmt.Errorf("package.json not found at %s", pkgFile)
	}

	fmt.Println("\n[Phase 3] Installing Node packages...")

	// prefer npm, fall back to yarn
	npmCmd := "npm"
	npmArgs := []string{"install"}
	if _, err := exec.LookPath("npm"); err != nil {
		if _, err := exec.LookPath("yarn"); err != nil {
			return fmt.Errorf("neither npm nor yarn found — install Node.js from https://nodejs.org")
		}
		npmCmd = "yarn"
		npmArgs = []string{}
	}

	fmt.Printf("  Running: %s %v\n", npmCmd, npmArgs)
	cmd := exec.Command(npmCmd, npmArgs...)
	cmd.Dir = projectPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s install failed: %w\n  Make sure Node.js is installed: https://nodejs.org", npmCmd, err)
	}

	fmt.Println("  ✓ Node packages installed")
	fmt.Println("\n  To run your app:")
	fmt.Println("    node index.js   (or whatever your entry file is)")
	fmt.Println("    Then open: http://localhost:3000")
	return nil
}

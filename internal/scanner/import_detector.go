package scanner

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// DetectPythonImports scans Python files and extracts all imports
func DetectPythonImports(projectPath string) ([]string, error) {
	importSet := make(map[string]bool)

	err := filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip venv and hidden directories
		if info.IsDir() {
			name := info.Name()
			if name == "venv" || name == "__pycache__" || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
		}

		// Process .py files
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".py") {
			imports, err := extractPythonImportsFromFile(path)
			if err != nil {
				return err
			}

			for _, imp := range imports {
				importSet[imp] = true
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Convert set to slice
	var imports []string
	for imp := range importSet {
		imports = append(imports, imp)
	}

	return imports, nil
}

func extractPythonImportsFromFile(filepath string) ([]string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var imports []string
	scanner := bufio.NewScanner(file)

	// Regex patterns
	importPattern := regexp.MustCompile(`^import\s+([a-zA-Z0-9_]+)`)
	fromPattern := regexp.MustCompile(`^from\s+([a-zA-Z0-9_]+)\s+import`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Match "import numpy"
		if matches := importPattern.FindStringSubmatch(line); matches != nil {
			imports = append(imports, matches[1])
		}

		// Match "from flask import Flask"
		if matches := fromPattern.FindStringSubmatch(line); matches != nil {
			imports = append(imports, matches[1])
		}
	}

	return imports, scanner.Err()
}

// DetectNodeImports scans JavaScript files for require/import
func DetectNodeImports(projectPath string) ([]string, error) {
	importSet := make(map[string]bool)

	err := filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip node_modules
		if info.IsDir() && info.Name() == "node_modules" {
			return filepath.SkipDir
		}

		// Process .js files
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".js") {
			imports, err := extractNodeImportsFromFile(path)
			if err != nil {
				return err
			}

			for _, imp := range imports {
				importSet[imp] = true
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	var imports []string
	for imp := range importSet {
		imports = append(imports, imp)
	}

	return imports, nil
}

func extractNodeImportsFromFile(filepath string) ([]string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var imports []string
	scanner := bufio.NewScanner(file)

	// Patterns: require('express'), import express from 'express'
	requirePattern := regexp.MustCompile(`require\(['"]([^'"]+)['"]\)`)
	importPattern := regexp.MustCompile(`import\s+.*\s+from\s+['"]([^'"]+)['"]`)

	for scanner.Scan() {
		line := scanner.Text()

		if matches := requirePattern.FindStringSubmatch(line); matches != nil {
			// Skip relative imports
			if !strings.HasPrefix(matches[1], ".") && !strings.HasPrefix(matches[1], "/") {
				imports = append(imports, matches[1])
			}
		}

		if matches := importPattern.FindStringSubmatch(line); matches != nil {
			if !strings.HasPrefix(matches[1], ".") && !strings.HasPrefix(matches[1], "/") {
				imports = append(imports, matches[1])
			}
		}
	}

	return imports, scanner.Err()
}

// MapPythonImportToPackage maps import names to PyPI package names
func MapPythonImportToPackage(importName string) string {
	mappings := map[string]string{
		"cv2":      "opencv-python",
		"PIL":      "Pillow",
		"sklearn":  "scikit-learn",
		"yaml":     "pyyaml",
		"OpenSSL":  "pyopenssl",
		"dotenv":   "python-dotenv",
		"jwt":      "pyjwt",
		"bs4":      "beautifulsoup4",
		"psycopg2": "psycopg2-binary",
	}

	if pkg, exists := mappings[importName]; exists {
		return pkg
	}

	return importName
}

// FilterPythonStdLib removes Python standard library imports
func FilterPythonStdLib(imports []string) []string {
	stdLib := map[string]bool{
		"os": true, "sys": true, "re": true, "json": true, "time": true,
		"datetime": true, "math": true, "random": true, "collections": true,
		"itertools": true, "functools": true, "subprocess": true, "pathlib": true,
		"typing": true, "abc": true, "io": true, "pickle": true, "csv": true,
		"urllib": true, "http": true, "email": true, "html": true, "xml": true,
		"logging": true, "unittest": true, "argparse": true, "configparser": true,
		"socket": true, "threading": true, "multiprocessing": true, "queue": true,
		"tempfile": true, "shutil": true, "glob": true, "gzip": true, "zipfile": true,
	}

	var filtered []string
	for _, imp := range imports {
		if !stdLib[imp] {
			filtered = append(filtered, imp)
		}
	}

	return filtered
}

// FilterNodeBuiltins removes Node.js built-in modules
func FilterNodeBuiltins(imports []string) []string {
	builtins := map[string]bool{
		"fs": true, "path": true, "http": true, "https": true, "url": true,
		"crypto": true, "os": true, "util": true, "events": true, "stream": true,
		"buffer": true, "child_process": true, "cluster": true, "net": true,
	}

	var filtered []string
	for _, imp := range imports {
		if !builtins[imp] {
			filtered = append(filtered, imp)
		}
	}

	return filtered
}

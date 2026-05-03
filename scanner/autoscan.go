package scanner

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// AutoScanAndGenerate is the Phase 1 entry point.
// It checks for an existing manifest; if none found, it scans source files,
// generates requirements.txt or package.json, then parses the result.
func AutoScanAndGenerate(projectPath string) (*ProjectInfo, error) {
	// 1. Try existing manifests first
	if fileExists(filepath.Join(projectPath, "requirements.txt")) {
		fmt.Println("  Found existing requirements.txt — parsing...")
		return ScanProject(projectPath)
	}
	if fileExists(filepath.Join(projectPath, "package.json")) {
		fmt.Println("  Found existing package.json — parsing...")
		return ScanProject(projectPath)
	}

	// 2. No manifest — detect language by source files
	lang := detectLanguage(projectPath)
	switch lang {
	case "python":
		fmt.Println("  No requirements.txt found — scanning .py imports...")
		return generatePythonManifest(projectPath)
	case "node":
		fmt.Println("  No package.json found — scanning .js/.ts imports...")
		return generateNodeManifest(projectPath)
	default:
		return nil, fmt.Errorf(
			"no manifest found and no .py/.js/.ts source files detected in %s\n"+
				"  Create a requirements.txt (Python) or package.json (Node) and re-run.",
			projectPath,
		)
	}
}

// detectLanguage returns "python", "node", or "unknown"
func detectLanguage(projectPath string) string {
	pyCount := 0
	jsCount := 0
	filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		switch filepath.Ext(path) {
		case ".py":
			pyCount++
		case ".js", ".ts", ".jsx", ".tsx":
			jsCount++
		}
		return nil
	})
	if pyCount >= jsCount && pyCount > 0 {
		return "python"
	}
	if jsCount > 0 {
		return "node"
	}
	return "unknown"
}

// ─── Python ───────────────────────────────────────────────────────────────────

// importToPackage maps Python import names → PyPI package names
var importToPackage = map[string]string{
	"cv2":           "opencv-python",
	"PIL":           "Pillow",
	"sklearn":       "scikit-learn",
	"bs4":           "beautifulsoup4",
	"yaml":          "PyYAML",
	"dotenv":        "python-dotenv",
	"psycopg2":      "psycopg2-binary",
	"dateutil":      "python-dateutil",
	"Crypto":        "pycryptodome",
	"jwt":           "PyJWT",
	"attr":          "attrs",
	"pkg_resources": "setuptools",
	"serial":        "pyserial",
	"usb":           "pyusb",
	"Image":         "Pillow",
	"requests":      "requests",
	"flask":         "flask",
	"django":        "django",
	"fastapi":       "fastapi",
	"sqlalchemy":    "sqlalchemy",
	"celery":        "celery",
	"redis":         "redis",
	"boto3":         "boto3",
	"numpy":         "numpy",
	"pandas":        "pandas",
	"matplotlib":    "matplotlib",
	"scipy":         "scipy",
	"torch":         "torch",
	"tensorflow":    "tensorflow",
	"pydantic":      "pydantic",
	"uvicorn":       "uvicorn",
	"aiohttp":       "aiohttp",
	"httpx":         "httpx",
	"pytest":        "pytest",
	"click":         "click",
	"typer":         "typer",
	"rich":          "rich",
	"loguru":        "loguru",
	"paramiko":      "paramiko",
	"cryptography":  "cryptography",
	"lxml":          "lxml",
	"pymongo":       "pymongo",
	"motor":         "motor",
}

// stdlibModules lists common Python standard library modules to skip
var stdlibModules = map[string]bool{
	"os": true, "sys": true, "re": true, "json": true, "math": true,
	"time": true, "datetime": true, "collections": true, "itertools": true,
	"functools": true, "pathlib": true, "typing": true, "io": true,
	"abc": true, "copy": true, "enum": true, "dataclasses": true,
	"logging": true, "threading": true, "subprocess": true, "socket": true,
	"urllib": true, "http": true, "email": true, "html": true, "xml": true,
	"csv": true, "hashlib": true, "hmac": true, "base64": true, "struct": true,
	"string": true, "textwrap": true, "pprint": true, "inspect": true,
	"traceback": true, "warnings": true, "contextlib": true, "weakref": true,
	"gc": true, "platform": true, "shutil": true, "tempfile": true,
	"glob": true, "fnmatch": true, "stat": true, "queue": true, "heapq": true,
	"bisect": true, "array": true, "random": true, "statistics": true,
	"decimal": true, "fractions": true, "numbers": true, "operator": true,
	"__future__": true, "builtins": true, "types": true, "unittest": true,
	"argparse": true, "configparser": true, "pickle": true, "shelve": true,
	"sqlite3": true, "ast": true, "dis": true, "token": true, "tokenize": true,
	"concurrent": true, "asyncio": true, "multiprocessing": true,
}

var (
	pyImportRe     = regexp.MustCompile(`^import\s+([\w]+)`)
	pyFromImportRe = regexp.MustCompile(`^from\s+([\w]+)`)
)

func generatePythonManifest(projectPath string) (*ProjectInfo, error) {
	pkgSet := map[string]bool{}

	filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".py") {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer f.Close()
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			var mod string
			if m := pyImportRe.FindStringSubmatch(line); m != nil {
				mod = m[1]
			} else if m := pyFromImportRe.FindStringSubmatch(line); m != nil {
				mod = m[1]
			}
			if mod == "" || stdlibModules[mod] {
				continue
			}
			pkg := mod
			if mapped, ok := importToPackage[mod]; ok {
				pkg = mapped
			}
			pkgSet[pkg] = true
		}
		return nil
	})

	if len(pkgSet) == 0 {
		return nil, fmt.Errorf("no third-party imports found in .py files under %s", projectPath)
	}

	// Write requirements.txt
	reqPath := filepath.Join(projectPath, "requirements.txt")
	f, err := os.Create(reqPath)
	if err != nil {
		return nil, fmt.Errorf("cannot create requirements.txt: %w", err)
	}
	fmt.Fprintf(f, "# Auto-generated by depsolver Phase 1\n")
	for pkg := range pkgSet {
		fmt.Fprintf(f, "%s\n", pkg)
	}
	f.Close()
	fmt.Printf("  ✓ Generated %s (%d packages)\n", reqPath, len(pkgSet))

	// Now parse it normally
	return ScanProject(projectPath)
}

// ─── Node ─────────────────────────────────────────────────────────────────────

// nodeBuiltins are Node.js built-in modules to skip
var nodeBuiltins = map[string]bool{
	"fs": true, "path": true, "os": true, "http": true, "https": true,
	"url": true, "util": true, "events": true, "stream": true, "buffer": true,
	"crypto": true, "child_process": true, "cluster": true, "dns": true,
	"net": true, "tls": true, "readline": true, "repl": true, "vm": true,
	"assert": true, "zlib": true, "querystring": true, "string_decoder": true,
	"timers": true, "tty": true, "dgram": true, "v8": true, "worker_threads": true,
	"perf_hooks": true, "process": true, "module": true, "console": true,
}

var (
	jsRequireRe = regexp.MustCompile(`require\(['"]([^'"./][^'"]*)['"]\)`)
	jsImportRe  = regexp.MustCompile(`^import\s+.*\s+from\s+['"]([^'"./][^'"]*)['"]`)
	jsImportRe2 = regexp.MustCompile(`^import\s+['"]([^'"./][^'"]*)['"]`)
)

func generateNodeManifest(projectPath string) (*ProjectInfo, error) {
	pkgSet := map[string]bool{}
	exts := map[string]bool{".js": true, ".ts": true, ".jsx": true, ".tsx": true, ".mjs": true}

	filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		// skip node_modules
		if strings.Contains(path, "node_modules") {
			return nil
		}
		if !exts[filepath.Ext(path)] {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer f.Close()
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			var mod string
			if m := jsRequireRe.FindStringSubmatch(line); m != nil {
				mod = m[1]
			} else if m := jsImportRe.FindStringSubmatch(line); m != nil {
				mod = m[1]
			} else if m := jsImportRe2.FindStringSubmatch(line); m != nil {
				mod = m[1]
			}
			if mod == "" {
				continue
			}
			// strip scoped package subpath: @org/pkg/sub → @org/pkg
			parts := strings.SplitN(mod, "/", 3)
			if strings.HasPrefix(mod, "@") && len(parts) >= 2 {
				mod = parts[0] + "/" + parts[1]
			} else {
				mod = parts[0]
			}
			if nodeBuiltins[mod] {
				continue
			}
			pkgSet[mod] = true
		}
		return nil
	})

	if len(pkgSet) == 0 {
		return nil, fmt.Errorf("no third-party imports found in .js/.ts files under %s", projectPath)
	}

	// Write package.json
	deps := make(map[string]string, len(pkgSet))
	for pkg := range pkgSet {
		deps[pkg] = "*"
	}
	pkgJSON := map[string]interface{}{
		"name":         "auto-generated",
		"version":      "1.0.0",
		"dependencies": deps,
	}
	data, _ := json.MarshalIndent(pkgJSON, "", "  ")
	pkgPath := filepath.Join(projectPath, "package.json")
	if err := os.WriteFile(pkgPath, data, 0644); err != nil {
		return nil, fmt.Errorf("cannot write package.json: %w", err)
	}
	fmt.Printf("  ✓ Generated %s (%d packages)\n", pkgPath, len(pkgSet))

	return ScanProject(projectPath)
}

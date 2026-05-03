package main

import (
	"fmt"
	"os"

	"github.com/depsolver/resolver/internal/bootstrap"
	"github.com/depsolver/resolver/internal/installer"
	"github.com/depsolver/resolver/internal/resolver"
	"github.com/depsolver/resolver/internal/scanner"
)

func main() {
	projectPath := "."
	if len(os.Args) > 1 {
		projectPath = os.Args[1]
	}
	abs, err := os.Getwd()
	if err == nil && projectPath == "." {
		projectPath = abs
	}

	verbose := false
	skipInstall := false
	for _, arg := range os.Args[2:] {
		if arg == "-v" || arg == "--verbose" {
			verbose = true
		}
		if arg == "--no-install" {
			skipInstall = true
		}
	}

	fmt.Printf("depsolver — project: %s\n", projectPath)
	fmt.Println("─────────────────────────────────────────────────────")

	// ── Phase 1: auto-scan → generate manifest → parse deps ──
	fmt.Println("\n[Phase 1] Detecting project & generating manifest...")
	info, err := scanner.AutoScanAndGenerate(projectPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Scan error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("  Language     : %s\n", info.Type)
	fmt.Printf("  Dependencies : %d\n", len(info.Dependencies))
	fmt.Printf("  Services     : %d\n", len(info.Services))

	if len(info.Dependencies) > 0 {
		fmt.Println("\n  Packages found in manifest:")
		for _, d := range info.Dependencies {
			fmt.Printf("    %-35s %s\n", d.Name, d.Version)
		}
	}
	if len(info.Services) > 0 {
		fmt.Println("\n  External services required:")
		for _, s := range info.Services {
			fmt.Printf("    %-12s v%s  (%s)\n", s.Name, s.Version, s.Reason)
		}
	}

	// ── Phase 2: resolve versions via PubGrub + rewrite manifest ─
	// resolver.Run picks PyPI for python, npm for node automatically
	result := resolver.Run(info, verbose)
	if !result.Success {
		fmt.Fprintf(os.Stderr, "\n✗ Version resolution failed — see conflicts above.\n")
		os.Exit(1)
	}

	// ── Phase 2b: docker-compose for services (postgres, redis etc.) ─
	fmt.Println("\n[Phase 2b] Bootstrapping Docker services...")
	if err := bootstrap.Bootstrap(info); err != nil {
		// non-fatal: Docker may not be running, warn and continue
		fmt.Fprintf(os.Stderr, "  ⚠  Docker bootstrap warning: %v\n", err)
		fmt.Fprintf(os.Stderr, "  Start Docker Desktop and re-run to spin up services.\n")
	}

	// ── Phase 3: install packages automatically ───────────────
	if !skipInstall {
		if err := installer.Install(info.Type, projectPath); err != nil {
			fmt.Fprintf(os.Stderr, "\n⚠  Install warning: %v\n", err)
			fmt.Fprintf(os.Stderr, "   Run manually:\n")
			switch info.Type {
			case "python":
				fmt.Fprintf(os.Stderr, "     pip install -r %s/requirements.txt\n", projectPath)
			case "node":
				fmt.Fprintf(os.Stderr, "     cd %s && npm install\n", projectPath)
			}
		}
	} else {
		fmt.Println("\n[Phase 3] Skipped (--no-install flag set)")
		fmt.Printf("  Run manually: ")
		switch info.Type {
		case "python":
			fmt.Printf("pip install -r %s/requirements.txt\n", projectPath)
		case "node":
			fmt.Printf("cd %s && npm install\n", projectPath)
		}
	}

	fmt.Println("\n✔  All done.")
}

package app

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
)

const pypiBaseURL = "https://pypi.org/pypi"

// PyPIClient fetches package metadata from PyPI
type PyPIClient struct {
	client  *http.Client
	cache   map[string]*PackageInfo
	verbose bool
}

func NewPyPIClient(verbose bool) *PyPIClient {
	return &PyPIClient{
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
		cache:   make(map[string]*PackageInfo),
		verbose: verbose,
	}
}

// PyPI JSON API response structures
type pypiResponse struct {
	Info     pypiInfo                 `json:"info"`
	Releases map[string][]pypiRelease `json:"releases"`
}

type pypiInfo struct {
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	RequiresDist []string `json:"requires_dist"`
}

type pypiRelease struct {
	PackageType string `json:"packagetype"`
	UploadTime  string `json:"upload_time"`
}

// FetchPackage retrieves metadata for a package from PyPI
func (c *PyPIClient) FetchPackage(name string) (*PackageInfo, error) {
	if info, ok := c.cache[strings.ToLower(name)]; ok {
		return info, nil
	}

	url := fmt.Sprintf("%s/%s/json", pypiBaseURL, name)
	if c.verbose {
		fmt.Printf("  → Fetching %s from PyPI\n", name)
	}

	resp, err := c.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("HTTP error fetching %s: %w", name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("package %q not found on PyPI", name)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("PyPI returned HTTP %d for %s", resp.StatusCode, name)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response for %s: %w", name, err)
	}

	var pypi pypiResponse
	if err := json.Unmarshal(body, &pypi); err != nil {
		return nil, fmt.Errorf("parsing JSON for %s: %w", name, err)
	}

	info := &PackageInfo{
		Name: pypi.Info.Name,
		Deps: make(map[string][]Dependency),
	}

	// collect and sort versions
	for vStr := range pypi.Releases {
		if v, ok := parseVersion(vStr); ok {
			info.Versions = append(info.Versions, v)
		}
	}
	sort.Slice(info.Versions, func(i, j int) bool {
		return info.Versions[i].Compare(info.Versions[j]) > 0 // descending
	})

	// parse top-level deps (latest version)
	deps := parseDeps(pypi.Info.RequiresDist)
	info.Deps[pypi.Info.Version] = deps

	c.cache[strings.ToLower(name)] = info
	return info, nil
}

// FetchVersionDeps fetches deps for a specific version of a package
func (c *PyPIClient) FetchVersionDeps(name, version string) ([]Dependency, error) {
	url := fmt.Sprintf("%s/%s/%s/json", pypiBaseURL, name, version)

	resp, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		// fall back to top-level info
		info, err := c.FetchPackage(name)
		if err != nil {
			return nil, err
		}
		for _, deps := range info.Deps {
			return deps, nil
		}
		return nil, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var pypi pypiResponse
	if err := json.Unmarshal(body, &pypi); err != nil {
		return nil, err
	}

	return parseDeps(pypi.Info.RequiresDist), nil
}

// parseDeps converts requires_dist strings into Dependency structs
func parseDeps(requiresDist []string) []Dependency {
	var deps []Dependency
	for _, req := range requiresDist {
		// skip extras/conditional deps for now
		if strings.Contains(req, "extra ==") || strings.Contains(req, "extra==") {
			continue
		}
		// strip environment markers
		if idx := strings.Index(req, ";"); idx != -1 {
			req = strings.TrimSpace(req[:idx])
		}
		// Handle old-style parenthesized constraints: "urllib3 (>=1.21,<1.27)"
		if idx := strings.Index(req, "("); idx != -1 {
			pkgName := strings.TrimSpace(req[:idx])
			rest := req[idx+1:]
			if end := strings.Index(rest, ")"); end != -1 {
				rest = rest[:end]
			}
			pkgName = NormalizePackageName(pkgName)
			if pkgName != "" {
				dep := Dependency{
					Package:     pkgName,
					Constraints: ParseConstraintStr(strings.TrimSpace(rest)),
				}
				deps = append(deps, dep)
			}
			continue
		}
		// strip extras
		if idx := strings.Index(req, "["); idx != -1 {
			req = req[:idx] + req[strings.Index(req, "]")+1:]
			req = strings.TrimSpace(req)
		}
		if req == "" {
			continue
		}

		dep, ok := parseDependencySpec(req)
		if ok {
			deps = append(deps, dep)
		}
	}
	return deps
}

// parseDependencySpec parses a PEP 508 dependency string like "werkzeug>=2.0"
func parseDependencySpec(spec string) (Dependency, bool) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return Dependency{}, false
	}

	// find where constraints start
	constraintStart := -1
	ops := []string{"==", ">=", "<=", "!=", "~=", ">", "<"}
	for _, op := range ops {
		if idx := strings.Index(spec, op); idx != -1 {
			if constraintStart == -1 || idx < constraintStart {
				constraintStart = idx
			}
		}
	}

	var pkgName string
	var constraintStr string
	if constraintStart == -1 {
		pkgName = strings.TrimSpace(spec)
	} else {
		pkgName = strings.TrimSpace(spec[:constraintStart])
		constraintStr = strings.TrimSpace(spec[constraintStart:])
	}

	pkgName = NormalizePackageName(pkgName)

	dep := Dependency{
		Package:     pkgName,
		Constraints: ParseConstraintStr(constraintStr),
	}
	return dep, true
}

// parseConstraintStr parses ">=2.0,<3.0" into []Constraint
func ParseConstraintStr(s string) []Constraint {
	if s == "" {
		return nil
	}
	var constraints []Constraint
	parts := strings.Split(s, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		ops := []string{"==", ">=", "<=", "!=", "~=", ">", "<"}
		for _, op := range ops {
			if strings.HasPrefix(part, op) {
				vStr := strings.TrimSpace(part[len(op):])
				if v, ok := parseVersion(vStr); ok {
					constraints = append(constraints, Constraint{Op: op, Version: v})
				}
				break
			}
		}
	}
	return constraints
}

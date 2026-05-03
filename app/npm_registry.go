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

const npmBaseURL = "https://registry.npmjs.org"

// NpmClient fetches package metadata from the npm registry.
// It implements RegistryClient so it can plug directly into PubGrubSolver.
type NpmClient struct {
	client  *http.Client
	cache   map[string]*PackageInfo
	verbose bool
}

func NewNpmClient(verbose bool) *NpmClient {
	return &NpmClient{
		client:  &http.Client{Timeout: 15 * time.Second},
		cache:   make(map[string]*PackageInfo),
		verbose: verbose,
	}
}

// npm registry JSON structures
type npmPackageResponse struct {
	Name     string                    `json:"name"`
	Versions map[string]npmVersionInfo `json:"versions"`
}

type npmVersionInfo struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Dependencies map[string]string `json:"dependencies"`
}

// FetchPackage retrieves all versions + latest deps from npm
func (c *NpmClient) FetchPackage(name string) (*PackageInfo, error) {
	key := strings.ToLower(name)
	if info, ok := c.cache[key]; ok {
		return info, nil
	}

	url := fmt.Sprintf("%s/%s", npmBaseURL, name)
	if c.verbose {
		fmt.Printf("  → Fetching %s from npm\n", name)
	}

	resp, err := c.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("HTTP error fetching %s: %w", name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("package %q not found on npm", name)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("npm returned HTTP %d for %s", resp.StatusCode, name)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading npm response for %s: %w", name, err)
	}

	var npmResp npmPackageResponse
	if err := json.Unmarshal(body, &npmResp); err != nil {
		return nil, fmt.Errorf("parsing npm JSON for %s: %w", name, err)
	}

	info := &PackageInfo{
		Name: npmResp.Name,
		Deps: make(map[string][]Dependency),
	}

	// parse all versions
	for vStr, vInfo := range npmResp.Versions {
		if v, ok := parseVersion(vStr); ok {
			info.Versions = append(info.Versions, v)
			// store deps for every version
			info.Deps[vStr] = parseNpmDeps(vInfo.Dependencies)
		}
	}

	// sort versions descending (latest first)
	sort.Slice(info.Versions, func(i, j int) bool {
		return info.Versions[i].Compare(info.Versions[j]) > 0
	})

	c.cache[key] = info
	return info, nil
}

// FetchVersionDeps returns deps for a specific npm package version
func (c *NpmClient) FetchVersionDeps(name, version string) ([]Dependency, error) {
	info, err := c.FetchPackage(name)
	if err != nil {
		return nil, err
	}
	if deps, ok := info.Deps[version]; ok {
		return deps, nil
	}
	// fallback: return first available
	for _, deps := range info.Deps {
		return deps, nil
	}
	return nil, nil
}

// parseNpmDeps converts npm's dependencies map to []Dependency
func parseNpmDeps(deps map[string]string) []Dependency {
	var out []Dependency
	for name, versionRange := range deps {
		out = append(out, Dependency{
			Package:     name,
			Constraints: parseNpmVersionRange(versionRange),
		})
	}
	return out
}

// parseNpmVersionRange converts npm version range strings to []Constraint
// Handles: "^1.2.3", "~1.2.3", ">=1.0.0 <2.0.0", "*", "1.2.3"
func parseNpmVersionRange(s string) []Constraint {
	s = strings.TrimSpace(s)
	if s == "" || s == "*" || s == "latest" {
		return nil
	}

	// handle ">=x <y" style ranges (split on space)
	if strings.Contains(s, " ") {
		parts := strings.Fields(s)
		var cs []Constraint
		for _, p := range parts {
			if sub := ParseConstraintStr(p); len(sub) > 0 {
				cs = append(cs, sub...)
			}
		}
		return cs
	}

	// handle caret: ^1.2.3 → >=1.2.3 <2.0.0
	if strings.HasPrefix(s, "^") {
		v := ParseVersion(strings.TrimPrefix(s, "^"))
		return []Constraint{
			{Op: ">=", Version: v},
			{Op: "<", Version: Version{Raw: fmt.Sprintf("%d.0.0", v.Major+1), Major: v.Major + 1}},
		}
	}

	// handle tilde: ~1.2.3 → >=1.2.3 <1.3.0
	if strings.HasPrefix(s, "~") {
		v := ParseVersion(strings.TrimPrefix(s, "~"))
		return []Constraint{
			{Op: ">=", Version: v},
			{Op: "<", Version: Version{Raw: fmt.Sprintf("%d.%d.0", v.Major, v.Minor+1), Major: v.Major, Minor: v.Minor + 1}},
		}
	}

	return ParseConstraintStr(s)
}

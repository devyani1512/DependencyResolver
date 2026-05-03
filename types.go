package app

import (
	"fmt"
	"strconv"
	"strings"
)

// ─── Version ─────────────────────────────────────────────────────────────────

// Version holds a parsed semantic version.
type Version struct {
	Raw   string
	Major int
	Minor int
	Patch int
	Pre   string // pre-release tag, e.g. "alpha.1"
}

func (v Version) String() string { return v.Raw }

// ParseVersion parses a version string like "1.2.3" or "2.0.0a1".
func ParseVersion(s string) Version {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "v")
	s = strings.TrimPrefix(s, "==")
	s = strings.TrimPrefix(s, "^")
	s = strings.TrimPrefix(s, "~")

	v := Version{Raw: s}
	for _, sep := range []string{"-", "a", "b", "rc"} {
		if idx := strings.Index(s, sep); idx > 0 && sep != "-" {
			v.Pre = s[idx:]
			s = s[:idx]
			break
		} else if sep == "-" && idx > 0 {
			v.Pre = s[idx+1:]
			s = s[:idx]
			break
		}
	}
	parts := strings.SplitN(s, ".", 3)
	if len(parts) > 0 {
		v.Major, _ = strconv.Atoi(parts[0])
	}
	if len(parts) > 1 {
		v.Minor, _ = strconv.Atoi(parts[1])
	}
	if len(parts) > 2 {
		v.Patch, _ = strconv.Atoi(parts[2])
	}
	return v
}

// Compare returns -1, 0, or 1.
func (v Version) Compare(other Version) int {
	if v.Major != other.Major {
		return cmp(v.Major, other.Major)
	}
	if v.Minor != other.Minor {
		return cmp(v.Minor, other.Minor)
	}
	if v.Patch != other.Patch {
		return cmp(v.Patch, other.Patch)
	}
	return 0
}

func cmp(a, b int) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

// ─── Constraint ───────────────────────────────────────────────────────────────

// Constraint is a single version requirement, e.g. ">=1.0.0".
type Constraint struct {
	Op      string // "==", ">=", "<=", ">", "<", "~=", "^"
	Version Version
}

func (c Constraint) String() string { return c.Op + c.Version.Raw }

// Allows reports whether version v satisfies this constraint.
func (c Constraint) Allows(v Version) bool {
	cmpResult := v.Compare(c.Version)
	switch c.Op {
	case "==":
		return cmpResult == 0
	case "!=":
		return cmpResult != 0
	case ">=":
		return cmpResult >= 0
	case ">":
		return cmpResult > 0
	case "<=":
		return cmpResult <= 0
	case "<":
		return cmpResult < 0
	case "~=":
		return cmpResult >= 0 && v.Major == c.Version.Major
	case "^":
		return cmpResult >= 0 && v.Major == c.Version.Major
	case "*", "":
		return true
	}
	return true
}

// ParseConstraint parses a string like ">=1.2.3".
func ParseConstraint(s string) Constraint {
	s = strings.TrimSpace(s)
	for _, op := range []string{"~=", "==", "!=", ">=", "<=", ">", "<", "^"} {
		if strings.HasPrefix(s, op) {
			return Constraint{Op: op, Version: ParseVersion(strings.TrimPrefix(s, op))}
		}
	}
	if s == "*" || s == "" {
		return Constraint{Op: "*"}
	}
	return Constraint{Op: "==", Version: ParseVersion(s)}
}

// ─── VersionSet ───────────────────────────────────────────────────────────────

// VersionSet is the set of versions allowed by a list of constraints.
type VersionSet struct {
	Package     string
	Constraints []Constraint
}

// Allows reports whether all constraints allow v.
func (vs VersionSet) Allows(v Version) bool {
	for _, c := range vs.Constraints {
		if !c.Allows(v) {
			return false
		}
	}
	return true
}

// ─── Dependency ───────────────────────────────────────────────────────────────

// Dependency is a package + the version constraints required.
type Dependency struct {
	Package     string
	Constraints []Constraint
}

func (d Dependency) String() string {
	parts := make([]string, len(d.Constraints))
	for i, c := range d.Constraints {
		parts[i] = c.String()
	}
	return fmt.Sprintf("%s %s", d.Package, strings.Join(parts, ","))
}

// ─── Decision ─────────────────────────────────────────────────────────────────

// Decision records that a specific version was chosen at a given decision level.
type Decision struct {
	Package string
	Version Version
	Level   int
}

// ─── Incompatibility ─────────────────────────────────────────────────────────

// Incompatibility is a PubGrub logical clause.
type Incompatibility struct {
	Terms []IncompatTerm
	Cause string
}

// IncompatTerm is one literal in an incompatibility clause.
type IncompatTerm struct {
	Package    string
	VersionSet VersionSet
	Positive   bool
}

// ─── ConflictError ───────────────────────────────────────────────────────────

// ConflictError describes an unresolvable version conflict.
type ConflictError struct {
	Package    string
	Constraint string
	Reason     string
	Packages   []string
}

func (e ConflictError) Error() string { return e.Reason }

// ─── PackageInfo ──────────────────────────────────────────────────────────────

// PackageInfo holds metadata fetched from a registry (PyPI / npm).
type PackageInfo struct {
	Name     string
	Versions []Version
	Deps     map[string][]Dependency // version string → its dependencies
}

// ─── SolverResult ─────────────────────────────────────────────────────────────

// SolverResult is returned by the PubGrub solver.
type SolverResult struct {
	Resolved map[string]Version
	Errors   []ConflictError
	Steps    []string
	Success  bool
}

package app

import "fmt"

// Version represents a parsed semantic version
type Version struct {
	Major int
	Minor int
	Patch int
	Pre   string
	Raw   string
}

func (v Version) String() string { return v.Raw }

func (v Version) Compare(other Version) int {
	if v.Major != other.Major {
		return cmpInt(v.Major, other.Major)
	}
	if v.Minor != other.Minor {
		return cmpInt(v.Minor, other.Minor)
	}
	if v.Patch != other.Patch {
		return cmpInt(v.Patch, other.Patch)
	}
	if v.Pre == "" && other.Pre != "" {
		return 1
	}
	if v.Pre != "" && other.Pre == "" {
		return -1
	}
	if v.Pre < other.Pre {
		return -1
	}
	if v.Pre > other.Pre {
		return 1
	}
	return 0
}

func cmpInt(a, b int) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

type Constraint struct {
	Op      string
	Version Version
}

func (c Constraint) String() string { return fmt.Sprintf("%s%s", c.Op, c.Version.Raw) }

func (c Constraint) Satisfies(v Version) bool {
	cmp := v.Compare(c.Version)
	switch c.Op {
	case "==":
		return cmp == 0
	case "!=":
		return cmp != 0
	case ">=":
		return cmp >= 0
	case "<=":
		return cmp <= 0
	case ">":
		return cmp > 0
	case "<":
		return cmp < 0
	case "~=":
		return cmp >= 0
	}
	return false
}

type VersionSet struct {
	Package     string
	Constraints []Constraint
}

func (vs VersionSet) Allows(v Version) bool {
	for _, c := range vs.Constraints {
		if !c.Satisfies(v) {
			return false
		}
	}
	return true
}

func (vs VersionSet) String() string {
	if len(vs.Constraints) == 0 {
		return vs.Package + " (any)"
	}
	s := vs.Package
	for i, c := range vs.Constraints {
		if i == 0 {
			s += c.String()
		} else {
			s += "," + c.String()
		}
	}
	return s
}

type PackageInfo struct {
	Name     string
	Versions []Version
	Deps     map[string][]Dependency
}

type Dependency struct {
	Package     string
	Constraints []Constraint
	Extra       string
}

func (d Dependency) String() string {
	s := d.Package
	for i, c := range d.Constraints {
		if i == 0 {
			s += c.String()
		} else {
			s += "," + c.String()
		}
	}
	return s
}

type Decision struct {
	Package string
	Version Version
	Level   int
}

type Incompatibility struct {
	Terms []IncompatTerm
	Cause string
}

func (inc *Incompatibility) String() string {
	if inc.Cause != "" {
		return inc.Cause
	}
	s := "incompatibility: "
	for i, t := range inc.Terms {
		if i > 0 {
			s += " AND "
		}
		s += t.String()
	}
	return s
}

type IncompatTerm struct {
	Package    string
	VersionSet VersionSet
	Positive   bool
}

func (t IncompatTerm) String() string {
	if t.Positive {
		return t.Package + " " + t.VersionSet.String()
	}
	return "NOT " + t.Package + " " + t.VersionSet.String()
}

type SolverResult struct {
	Resolved map[string]Version
	Errors   []ConflictError
	Success  bool
	Steps    []string
}

type ConflictError struct {
	Package    string
	Constraint string
	Reason     string
	Packages   []string
}

func (e ConflictError) Error() string {
	return fmt.Sprintf("Conflict on %s: %s", e.Package, e.Reason)
}

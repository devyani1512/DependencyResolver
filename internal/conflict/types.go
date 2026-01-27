package conflict

// Conflict represents a dependency conflict
type Conflict struct {
	Type        string   // "missing", "version_mismatch", "circular"
	Description string   // Human-readable description
	Packages    []string // Affected packages
	Suggestion  string   // How to fix
}

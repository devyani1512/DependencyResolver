package conflict

// Conflict represents a dependency conflict
type Conflict struct {
	Type        string   // "missing", "version_mismatch", "circular"
	Description string
	Packages    []string
	Suggestion  string
}

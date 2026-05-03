package scanner

type ProjectInfo struct {
	Type         string
	Path         string
	Dependencies []Dependency
	Services     []ServiceRequirement
}

type Dependency struct {
	Name    string
	Version string
	Source  string // requirements.txt or package.json
}

type ServiceRequirement struct {
	Name    string // e.g. postgres, redis
	Version string
	Reason  string
}

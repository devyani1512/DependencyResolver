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
	Source  string //source of the file it came from - requiremnt.txt or package.json
}

type ServiceRequirement struct {
	Name    string // eg postgres or redis
	Version string
	Reason  string // the reason it is needed - like detected from imports
}

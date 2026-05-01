package scanner

//bufio reads file line by line
import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// reads requirements.txt and extracts dependencies
func ParsePythonDependencies(projectPath string) ([]Dependency, error) {
	reqFile := filepath.Join(projectPath, "requirements.txt")
	file, err := os.Open(reqFile)
	if err != nil {
		return nil, err
	}
	defer file.Close() //file closing
	var deps []Dependency
	scanner := bufio.NewScanner(file) //creates buffered scanner - reads file line by line

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		//to skip empth lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		//parse dependency
		dep := parsePythonLine(line)
		deps = append(deps, dep)
	}
	return deps, scanner.Err()
}

// to parse single requirement.txt line
func parsePythonLine(line string) Dependency {
	//to handle fomats like flask == 2.0.0
	//op gets operator value
	for _, op := range []string{"==", ">=", "<=", "~=", ">", "<"} {
		if strings.Contains(line, op) {
			parts := strings.Split(line, op) //seperating package name from version
			return Dependency{
				Name:    strings.TrimSpace(parts[0]),
				Version: op + strings.TrimSpace(parts[1]),
				Source:  "requirements.txt",
			}
		}
	}
	//no version specified - if there was no version given
	return Dependency{
		Name:    strings.TrimSpace(line),
		Version: "*",
		Source:  "requirements.txt",
	}
}

// mark later to add recursive scanning
// detect pythonservices and detects required servcies form python code
func DetectPythonServices(projectPath string) []ServiceRequirement {
	services := []ServiceRequirement{}
	//read python files to detect database imports
	//this is a simplified version
	//later we will recursively scan .py files

	//check for common patterns
	files, _ := filepath.Glob(filepath.Join(projectPath, "*.py"))
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		text := string(content)
		//detect postgreSQL
		//static analysis
		if strings.Contains(text, "psycopg2") || strings.Contains(text, "postgresql") {
			services = append(services, ServiceRequirement{
				Name:    "postgres",
				Version: "14",
				Reason:  "Detected psycopg2 import",
			})
		}
		//detect redis
		if strings.Contains(text, "import redis") || strings.Contains(text, "Redis") {
			services = append(services, ServiceRequirement{
				Name:    "redis",
				Version: "latest",
				Reason:  "Detected redis import",
			})
		}
	}
	return services
}

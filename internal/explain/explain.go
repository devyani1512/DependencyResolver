package explain

import "fmt"

// ExplainFailure provides human-readable error explanation
func ExplainFailure(err error) string {
	return fmt.Sprintf(`
 Bootstrap Failed

Error: %v

Possible causes:
1. Docker is not running
2. Invalid project structure
3. Missing dependency files

Suggested fixes:
- Start Docker Desktop
- Verify requirements.txt or package.json exists
- Check file permissions
`, err)
}

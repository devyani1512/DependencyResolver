package app

import "strings"

// importToPackage maps common Python import names to their PyPI package names
var importToPackage = map[string]string{
	"cv2":           "opencv-python",
	"PIL":           "Pillow",
	"sklearn":       "scikit-learn",
	"bs4":           "beautifulsoup4",
	"yaml":          "PyYAML",
	"dotenv":        "python-dotenv",
	"psycopg2":      "psycopg2-binary",
	"dateutil":      "python-dateutil",
	"Crypto":        "pycryptodome",
	"jwt":           "PyJWT",
	"attr":          "attrs",
	"pkg_resources": "setuptools",
	"google.cloud":  "google-cloud",
	"gi":            "PyGObject",
	"wx":            "wxPython",
	"usb":           "pyusb",
	"serial":        "pyserial",
	"gtk":           "PyGTK",
	"Image":         "Pillow",
}

// NormalizePackageName converts an import name to its PyPI package name
// and normalizes the casing/separators per PEP 503
func NormalizePackageName(name string) string {
	// strip version constraints to extract just the name
	for _, op := range []string{"==", ">=", "<=", "!=", "~=", ">", "<"} {
		if idx := strings.Index(name, op); idx != -1 {
			name = name[:idx]
		}
	}
	name = strings.TrimSpace(name)

	// check alias map
	if mapped, ok := importToPackage[name]; ok {
		return mapped
	}

	// PEP 503: normalize dashes/underscores/dots to dashes
	name = strings.ReplaceAll(name, "_", "-")
	name = strings.ReplaceAll(name, ".", "-")

	return name
}

// ParseRequirementsLine parses a single line from a requirements.txt file
// Returns package name and raw constraint string
func ParseRequirementsLine(line string) (string, string) {
	line = strings.TrimSpace(line)

	// skip comments and blanks
	if line == "" || strings.HasPrefix(line, "#") {
		return "", ""
	}

	// strip inline comments
	if idx := strings.Index(line, " #"); idx != -1 {
		line = strings.TrimSpace(line[:idx])
	}

	// split on first constraint operator
	ops := []string{"==", ">=", "<=", "!=", "~=", ">", "<"}
	for _, op := range ops {
		if idx := strings.Index(line, op); idx != -1 {
			pkg := strings.TrimSpace(line[:idx])
			constraint := strings.TrimSpace(line[idx:])
			return NormalizePackageName(pkg), constraint
		}
	}

	// no constraint - just a package name
	return NormalizePackageName(line), ""
}

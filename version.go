package app

import (
	"strconv"
	"strings"
)

// parseVersion parses a version string like "3.0.1", "2.4", "1.0rc2"
func parseVersion(s string) (Version, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Version{}, false
	}

	// skip complex pre/dev versions with letters mixed in epoch-style
	// but handle common ones: 1.0rc1, 1.0a1, 1.0b2, 1.0.dev1
	original := s

	// strip epoch (e.g. "1!2.3" → "2.3")
	if idx := strings.Index(s, "!"); idx != -1 {
		s = s[idx+1:]
	}

	// strip local version (+local)
	if idx := strings.Index(s, "+"); idx != -1 {
		s = s[:idx]
	}

	// extract pre-release suffix
	pre := ""
	for _, tag := range []string{".dev", "dev", "rc", "a", "b", ".post", "post"} {
		if idx := strings.Index(s, tag); idx != -1 {
			pre = s[idx:]
			s = s[:idx]
			break
		}
	}

	parts := strings.Split(s, ".")
	if len(parts) == 0 {
		return Version{}, false
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return Version{}, false
	}
	minor := 0
	patch := 0
	if len(parts) > 1 {
		minor, _ = strconv.Atoi(parts[1])
	}
	if len(parts) > 2 {
		// handle things like "1.0.0a1" — strip trailing non-numeric
		p := parts[2]
		p = strings.TrimRight(p, "abcdefghijklmnopqrstuvwxyz")
		patch, _ = strconv.Atoi(p)
	}

	return Version{
		Major: major,
		Minor: minor,
		Patch: patch,
		Pre:   pre,
		Raw:   original,
	}, true
}

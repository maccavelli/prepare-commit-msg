// Package selfupdate provides cross-platform self-update functionality from GitHub Releases.
package selfupdate

import (
	"fmt"
	"strconv"
	"strings"
)

// Version represents a parsed Semantic Version (SemVer 2.0.0).
type Version struct {
	Major      int
	Minor      int
	Patch      int
	Prerelease string
	Raw        string
}

// CleanVersion normalizes a version string by trimming whitespace and leading 'v' / 'V'.
func CleanVersion(v string) string {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(v, "v")
	v = strings.TrimPrefix(v, "V")
	return v
}

// ParseSemver parses a version string into a structured Version.
func ParseSemver(v string) (Version, error) {
	raw := strings.TrimSpace(v)
	cleaned := CleanVersion(raw)
	if cleaned == "" {
		return Version{Raw: raw}, fmt.Errorf("empty version string")
	}

	// Separate build metadata if present (ignored in semver precedence).
	if idx := strings.IndexByte(cleaned, '+'); idx != -1 {
		cleaned = cleaned[:idx]
	}

	// Separate prerelease if present.
	var prerelease string
	if idx := strings.IndexByte(cleaned, '-'); idx != -1 {
		prerelease = cleaned[idx+1:]
		cleaned = cleaned[:idx]
	}

	parts := strings.Split(cleaned, ".")
	if len(parts) < 1 || len(parts) > 3 {
		return Version{Raw: raw}, fmt.Errorf("invalid semver format %q", raw)
	}

	var major, minor, patch int
	var err error

	major, err = strconv.Atoi(parts[0])
	if err != nil || major < 0 {
		return Version{Raw: raw}, fmt.Errorf("invalid major version in %q: %w", raw, err)
	}

	if len(parts) >= 2 {
		minor, err = strconv.Atoi(parts[1])
		if err != nil || minor < 0 {
			return Version{Raw: raw}, fmt.Errorf("invalid minor version in %q: %w", raw, err)
		}
	}

	if len(parts) >= 3 {
		patch, err = strconv.Atoi(parts[2])
		if err != nil || patch < 0 {
			return Version{Raw: raw}, fmt.Errorf("invalid patch version in %q: %w", raw, err)
		}
	}

	return Version{
		Major:      major,
		Minor:      minor,
		Patch:      patch,
		Prerelease: prerelease,
		Raw:        raw,
	}, nil
}

// String returns the canonical normalized string representation (e.g. "4.3.2" or "4.3.2-rc.1").
func (v Version) String() string {
	base := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.Prerelease != "" {
		base += "-" + v.Prerelease
	}
	return base
}

// Compare compares two version strings.
// Returns:
//   - -1 if v1 < v2
//   - 0 if v1 == v2
//   - 1 if v1 > v2
func Compare(v1, v2 string) int {
	ver1, err1 := ParseSemver(v1)
	ver2, err2 := ParseSemver(v2)

	// Fallback handling if unparseable (e.g., "dev", "dirty", "unknown")
	if err1 != nil || err2 != nil {
		c1, c2 := CleanVersion(v1), CleanVersion(v2)
		if c1 == c2 {
			return 0
		}
		if err1 == nil && err2 != nil {
			// Valid version is considered newer than unversioned/dev
			return 1
		}
		if err1 != nil && err2 == nil {
			return -1
		}
		return strings.Compare(c1, c2)
	}

	if ver1.Major != ver2.Major {
		if ver1.Major > ver2.Major {
			return 1
		}
		return -1
	}

	if ver1.Minor != ver2.Minor {
		if ver1.Minor > ver2.Minor {
			return 1
		}
		return -1
	}

	if ver1.Patch != ver2.Patch {
		if ver1.Patch > ver2.Patch {
			return 1
		}
		return -1
	}

	// SemVer 2.0.0: When major, minor, and patch are equal, a normal version
	// has higher precedence than a pre-release version.
	if ver1.Prerelease == "" && ver2.Prerelease != "" {
		return 1
	}
	if ver1.Prerelease != "" && ver2.Prerelease == "" {
		return -1
	}
	if ver1.Prerelease != "" && ver2.Prerelease != "" {
		return comparePrereleases(ver1.Prerelease, ver2.Prerelease)
	}

	return 0
}

// comparePrereleases compares two dot-separated prerelease identifier strings.
func comparePrereleases(pre1, pre2 string) int {
	parts1 := strings.Split(pre1, ".")
	parts2 := strings.Split(pre2, ".")

	minLen := len(parts1)
	if len(parts2) < minLen {
		minLen = len(parts2)
	}

	for i := 0; i < minLen; i++ {
		p1, p2 := parts1[i], parts2[i]
		num1, err1 := strconv.Atoi(p1)
		num2, err2 := strconv.Atoi(p2)

		if err1 == nil && err2 == nil {
			if num1 != num2 {
				if num1 > num2 {
					return 1
				}
				return -1
			}
		} else if err1 == nil && err2 != nil {
			// Numeric identifiers always have lower precedence than non-numeric
			return -1
		} else if err1 != nil && err2 == nil {
			return 1
		} else {
			if cmp := strings.Compare(p1, p2); cmp != 0 {
				return cmp
			}
		}
	}

	if len(parts1) > len(parts2) {
		return 1
	}
	if len(parts1) < len(parts2) {
		return -1
	}
	return 0
}

// IsNewer returns true if candidate is strictly newer than current.
func IsNewer(candidate, current string) bool {
	return Compare(candidate, current) > 0
}

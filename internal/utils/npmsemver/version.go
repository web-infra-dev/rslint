package npmsemver

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

// Repository-authored npm version syntax, with no user-controlled patterns.
var fullVersion = regexp.MustCompile(`^v?(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?(?:\+[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?$`)
var coercibleVersion = regexp.MustCompile(`(?:^|[^0-9])([0-9]{1,16})(?:\.([0-9]{1,16}))?(?:\.([0-9]{1,16}))?(?:$|[^0-9])`)

func parseVersion(text string) (versionComparator, bool) {
	if ecmascript.StringCodeUnitCount(text) > 256 {
		return versionComparator{}, false
	}
	match := fullVersion.FindStringSubmatch(ecmascript.StringTrim(text))
	if match == nil || match[4] != "" && !versionPrerelease.MatchString(match[4]) {
		return versionComparator{}, false
	}
	var result versionComparator
	for i := range 3 {
		component, err := strconv.ParseUint(match[i+1], 10, 64)
		if err != nil || component > maxRangeComponent {
			return versionComparator{}, false
		}
		result.version[i] = component
	}
	if match[4] != "" {
		result.prerelease = strings.Split(match[4], ".")
	}
	return result, true
}

// CoerceVersion preserves a valid npm version (including prereleases), or
// extracts the first numeric version as npm's default semver.coerce does.
func CoerceVersion(text string) (string, bool) {
	if v, ok := parseVersion(text); ok {
		return v.string(), true
	}
	match := coercibleVersion.FindStringSubmatch(text)
	if match == nil {
		return "", false
	}
	for i := 2; i <= 3; i++ {
		if match[i] == "" {
			match[i] = "0"
		}
	}
	v, ok := parseVersion(strings.Join(match[1:4], "."))
	if !ok {
		return "", false
	}
	return v.string(), true
}

// SatisfiesComparison tests a > or >= npm comparison, including prereleases.
// Partial-version bounds follow npm, e.g. >0.9 admits 1.0.0-beta, while
// >1.0.0 requires a version later than the stable 1.0.0 release.
func SatisfiesComparison(version, comparison string) bool {
	if !strings.HasPrefix(comparison, ">") || strings.ContainsAny(comparison, "|~^<") {
		return false
	}
	v, ok := parseVersion(version)
	if !ok {
		return false
	}
	r, ok := parse(comparison, true)
	return ok && r.contains(v, true)
}

func (r Range) contains(v versionComparator, includePrerelease bool) bool {
	for _, alternative := range r.alternatives {
		matches := true
		for _, c := range alternative {
			order := v.compare(c)
			switch c.operator {
			case "", "=":
				matches = order == 0
			case ">":
				matches = order > 0
			case ">=":
				matches = order >= 0
			case "<":
				matches = order < 0
			case "<=":
				matches = order <= 0
			}
			if !matches {
				break
			}
		}
		if matches && (includePrerelease || len(v.prerelease) == 0 || hasPrereleaseTuple(alternative, &v)) {
			return true
		}
	}
	return false
}

// MinVersion returns npm's minimum satisfying version, or false for an empty
// range. It uses the same normalized comparators as the other range operations.
func (r Range) MinVersion() (string, bool) {
	for _, text := range []string{"0.0.0", "0.0.0-0"} {
		v, _ := parseVersion(text)
		if r.contains(v, false) {
			return text, true
		}
	}
	var minimum *versionComparator
	for _, alternative := range r.alternatives {
		var lower *versionComparator
		for _, c := range alternative {
			if c.operator == "<" || c.operator == "<=" {
				continue
			}
			if c.operator == ">" {
				if len(c.prerelease) == 0 {
					c.version[2]++
				} else {
					c.prerelease = append(append([]string{}, c.prerelease...), "0")
				}
			}
			if lower == nil || c.compare(*lower) > 0 {
				lower = &c
			}
		}
		if lower != nil && (minimum == nil || lower.compare(*minimum) < 0) {
			minimum = lower
		}
	}
	if minimum == nil || !r.contains(*minimum, false) {
		return "", false
	}
	return minimum.string(), true
}

func (v versionComparator) string() string {
	text := fmt.Sprintf("%d.%d.%d", v.version[0], v.version[1], v.version[2])
	if len(v.prerelease) > 0 {
		text += "-" + strings.Join(v.prerelease, ".")
	}
	return text
}

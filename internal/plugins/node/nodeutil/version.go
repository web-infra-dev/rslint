package nodeutil

import (
	"regexp"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/semver"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

// NodeVersion contains normalized comparator alternatives from tsgo's semver parser.
// Node rules use it to decide whether every supported runtime provides an API.
type NodeVersion struct{ alternatives [][]versionComparator }
type versionComparator struct {
	operator   string
	version    semver.Version
	prerelease []string
}

// The patterns are repository-authored npm range syntax, not user regexps.
var versionPrefix = regexp.MustCompile(`(^|[\s|<>=~^])v([0-9xX*])`)
var versionOperatorSpace = regexp.MustCompile(`([<>=~^])\s+`)
var versionWildcardTail = regexp.MustCompile(`(?:^|\.)[xX*]\.[0-9]`)
var versionPrerelease = regexp.MustCompile(`^(?:0|[1-9][0-9]*|[0-9]*[A-Za-z-][0-9A-Za-z-]*)(?:\.(?:0|[1-9][0-9]*|[0-9]*[A-Za-z-][0-9A-Za-z-]*))*$`)

func parseNodeVersion(text string) (NodeVersion, bool) {
	text = strings.Map(func(r rune) rune {
		if ecmascript.IsWhiteSpaceOrLineTerminator(r) {
			return ' '
		}
		return r
	}, ecmascript.StringTrim(text))
	text = versionPrefix.ReplaceAllString(text, "${1}${2}")
	text = versionOperatorSpace.ReplaceAllString(text, "$1")
	text = strings.ReplaceAll(text, "~>", "~")
	// npm treats an empty union arm as *, while tsgo drops it.
	arms := strings.Split(text, "||")
	for i, arm := range arms {
		if ecmascript.StringTrim(arm) == "" {
			arms[i] = "*"
			continue
		}
		tokens := strings.Fields(arm)
		hyphenRange := len(tokens) == 3 && tokens[1] == "-"
		for _, token := range tokens {
			if token == "-" {
				continue
			}
			version := strings.TrimLeft(token, "<>=~^")
			version, _, _ = strings.Cut(version, "+")
			base, prerelease, hasPrerelease := strings.Cut(version, "-")
			// npm expands caret, tilde and hyphen ranges before checking
			// wildcard order, so ^5.x.1 is valid even though 5.x.1 is not.
			if !hyphenRange && !strings.ContainsAny(token[:1], "~^") && versionWildcardTail.MatchString(base) {
				return NodeVersion{}, false
			}
			if hasPrerelease && !versionPrerelease.MatchString(prerelease) {
				return NodeVersion{}, false
			}
		}
	}
	// Keep range expansion in the compiler; only npm's stricter wildcard order
	// and its broader prerelease identifier grammar need adaptation here.
	parsed, ok := semver.TryParseVersionRange(strings.Join(arms, "||"))
	if !ok {
		return NodeVersion{}, false
	}
	var result NodeVersion
	for i, alternative := range strings.Split(parsed.String(), " || ") {
		var comparators []versionComparator
		for _, token := range strings.Fields(alternative) {
			if token == "*" {
				continue
			}
			version := strings.TrimLeft(token, "<>=")
			operator := strings.TrimSuffix(token, version)
			// tsgo includes prereleases in wildcard lower bounds. npm starts
			// those at the stable version, unless -0 was explicitly requested.
			if operator == ">=" && !strings.Contains(arms[i], version) {
				version = strings.TrimSuffix(version, "-0")
			}
			version, build, hasBuild := strings.Cut(version, "+")
			base, prerelease, hasPrerelease := strings.Cut(version, "-")
			if hasBuild {
				base += "+" + build
			}
			value, err := semver.TryParseVersion(base)
			if err != nil {
				return NodeVersion{}, false
			}
			comparator := versionComparator{operator: operator, version: value}
			if hasPrerelease {
				// The range parser accepts npm prereleases such as "1beta";
				// TryParseVersion rejects them. Retain them separately and reuse
				// the compiler's identifier comparison without parsing them again.
				comparator.prerelease = strings.Split(prerelease, ".")
			}
			comparators = append(comparators, comparator)
		}
		result.alternatives = append(result.alternatives, comparators)
	}
	return result, true
}

// ConfiguredNodeVersion follows options, settings.n, settings.node, package
// engines.node, devEngines.runtime, then the upstream >=16.0.0 fallback.
func ConfiguredNodeVersion(ctx rule.RuleContext, options map[string]any) NodeVersion {
	for _, raw := range settingValues("version", options, ctx.Settings) {
		if raw == nil || raw == "" || raw == false || raw == float64(0) {
			continue
		}
		if version, ok := parseNodeVersion(settingString(raw)); ok {
			return version
		}
	}
	if p := ctx.Program(); p != nil {
		if pkg := FindPackage(p, ctx.SourceFile.FileName()); pkg != nil {
			engines, _ := pkg.data["engines"].(map[string]any)
			if text, ok := engines["node"].(string); ok {
				if version, valid := parseNodeVersion(text); valid {
					return version
				}
			}
			devEngines, _ := pkg.data["devEngines"].(map[string]any)
			entries, ok := devEngines["runtime"].([]any)
			if !ok {
				entries = []any{devEngines["runtime"]}
			}
			for _, raw := range entries {
				entry, _ := raw.(map[string]any)
				if entry["name"] != "node" {
					continue
				}
				if text, ok := entry["version"].(string); ok {
					if version, valid := parseNodeVersion(text); valid {
						return version
					}
					break
				}
			}
		}
	}
	version, _ := parseNodeVersion(">=16.0.0")
	return version
}

// Supports reports whether the configured range has no intersection with
// versions below since, matching the replacement selection in eslint-plugin-n.
func (version NodeVersion) Supports(since string) bool {
	boundary := semver.MustParse(since)
	for _, alternative := range version.alternatives {
		constraints := append(append([]versionComparator{}, alternative...), versionComparator{operator: "<", version: boundary})
		intersects := true
		for i, left := range constraints {
			for _, right := range constraints[i+1:] {
				if !comparatorsIntersect(left, right) {
					intersects = false
					break
				}
			}
			if !intersects {
				break
			}
		}
		if intersects {
			return false
		}
	}
	return true
}

func comparatorsIntersect(left, right versionComparator) bool {
	for _, comparator := range []versionComparator{left, right} {
		if comparator.operator == "<" && strings.HasPrefix(comparator.version.String(), "0.0.0") {
			return false
		}
	}
	if right.operator == "=" {
		left, right = right, left
	}
	cmp := left.version.Compare(&right.version)
	if cmp == 0 {
		cmp = semver.ComparePreReleaseIdentifiers(left.prerelease, right.prerelease)
	}
	if left.operator == "=" {
		if right.operator == "=" {
			return cmp == 0
		}
		// A singleton prerelease does not satisfy a stable npm comparator.
		if len(left.prerelease) > 0 && len(right.prerelease) == 0 {
			return false
		}
		switch right.operator {
		case "<":
			return cmp < 0
		case "<=":
			return cmp <= 0
		case ">":
			return cmp > 0
		case ">=":
			return cmp >= 0
		}
	}
	leftLower, rightLower := strings.HasPrefix(left.operator, ">"), strings.HasPrefix(right.operator, ">")
	if leftLower == rightLower {
		return true
	}
	if !leftLower {
		cmp = -cmp
	}
	return cmp < 0 || cmp == 0 && strings.HasSuffix(left.operator, "=") && strings.HasSuffix(right.operator, "=")
}

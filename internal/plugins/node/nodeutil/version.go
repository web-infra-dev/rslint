package nodeutil

import (
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/semver"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

// NodeVersion contains normalized comparator alternatives from tsgo's semver parser.
// Node rules use it to decide whether every supported runtime provides an API.
type NodeVersion struct {
	alternatives [][]versionComparator
	uncertain    bool
}
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
	// npm's range grammar is ASCII after JS whitespace normalization. Do not
	// let Go's broader whitespace handling accept characters such as U+0085.
	for _, character := range text {
		if character > 127 {
			return NodeVersion{}, false
		}
	}
	text = versionPrefix.ReplaceAllString(text, "${1}${2}")
	text = versionOperatorSpace.ReplaceAllString(text, "$1")
	text = strings.ReplaceAll(text, "~>", "~")
	// npm treats an empty union arm as *, while tsgo drops it.
	arms := strings.Split(text, "||")
	uncertain := false
	for i, arm := range arms {
		if ecmascript.StringTrim(arm) == "" {
			arms[i] = "*"
			continue
		}
		tokens := strings.Fields(arm)
		hyphenRange := len(tokens) == 3 && tokens[1] == "-"
		armChanged := false
		for j, token := range tokens {
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
			// Keep valid extreme ranges in their configuration precedence slot.
			// Also guard MaxUint32 itself: range expansion can increment it.
			// Substitute only for grammar validation, never for comparisons.
			if len(base) < 10 {
				continue
			}
			parts := strings.Split(base, ".")
			changed := false
			for k, part := range parts {
				if len(part) < 10 || part[0] == '0' {
					continue
				}
				value, err := strconv.ParseUint(part, 10, 64)
				if err == nil && value >= math.MaxUint32 {
					if value > 1<<53-1 {
						return NodeVersion{}, false
					}
					parts[k] = "1"
					changed = true
					uncertain = true
				}
			}
			if changed {
				prefix := len(token) - len(strings.TrimLeft(token, "<>=~^"))
				tokens[j] = token[:prefix] + strings.Join(parts, ".") + token[prefix+len(base):]
				armChanged = true
			}
		}
		if armChanged {
			arms[i] = strings.Join(tokens, " ")
		}
	}
	// Keep range expansion and grammar validation in the compiler, including
	// validation of ranges whose numeric bounds cannot be compared safely.
	parsed, ok := semver.TryParseVersionRange(strings.Join(arms, "||"))
	if !ok {
		return NodeVersion{}, false
	}
	if uncertain {
		return NodeVersion{uncertain: true}, true
	}
	var result NodeVersion
	var emptyAlternative []versionComparator
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
		// npm drops its canonical null-set arm from unions unless it is
		// the only remaining arm. Other contradictory arms stay intact.
		if alternative == "<0.0.0-0" {
			emptyAlternative = comparators
		} else {
			result.alternatives = append(result.alternatives, comparators)
		}
	}
	if len(result.alternatives) == 0 {
		result.alternatives = [][]versionComparator{emptyAlternative}
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
	if version.uncertain {
		return false
	}
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

// IsSubsetOf checks that every configured alternative fits a supported npm
// range. Unlike Supports, it handles gaps between releases that gained a
// feature, and excludes prereleases absent from the supported range.
func (version NodeVersion) IsSubsetOf(supported string) bool {
	if version.uncertain {
		return false
	}
	domain, ok := parseNodeVersion(supported)
	if !ok || domain.uncertain {
		return false
	}
	for _, sub := range version.alternatives {
		lower, upper, empty := versionBounds(sub)
		if empty {
			// An empty arm contributes no versions, regardless of union order.
			continue
		}
		contained := false
		for _, dom := range domain.alternatives {
			domLower, domUpper, domEmpty := versionBounds(dom)
			if domEmpty || !versionBoundContains(domLower, lower, true) || !versionBoundContains(domUpper, upper, false) {
				continue
			}
			if lower.admitsPrerelease(true) && !hasPrereleaseTuple(dom, lower) ||
				upper.admitsPrerelease(false) && !hasPrereleaseTuple(dom, upper) {
				continue
			}
			contained = true
			break
		}
		if !contained {
			return false
		}
	}
	return true
}

func versionBounds(comparators []versionComparator) (lower, upper *versionComparator, empty bool) {
	if len(comparators) == 0 {
		// npm's default wildcard admits stable versions starting at 0.0.0.
		return &versionComparator{operator: ">="}, nil, false
	}
	for i := range comparators {
		c := &comparators[i]
		if c.operator == "=" {
			for _, other := range comparators {
				if !comparatorsIntersect(*c, other) {
					return nil, nil, true
				}
			}
		}
		if c.operator == "=" || strings.HasPrefix(c.operator, ">") {
			if lower == nil || !versionBoundContains(c, lower, true) {
				lower = c
			}
		}
		if c.operator == "=" || strings.HasPrefix(c.operator, "<") {
			if upper == nil || !versionBoundContains(c, upper, false) {
				upper = c
			}
		}
	}
	if lower != nil && upper != nil {
		order := lower.compare(*upper)
		empty = order > 0 || order == 0 && (lower.operator == ">" || upper.operator == "<")
	}
	return
}

func versionBoundContains(outer, inner *versionComparator, lower bool) bool {
	if outer == nil {
		return true
	}
	if inner == nil {
		return false
	}
	order := inner.compare(*outer)
	if !lower {
		order = -order
	}
	return order > 0 || order == 0 && (outer.operator != ">" && outer.operator != "<" || inner.operator == outer.operator)
}

func (bound *versionComparator) admitsPrerelease(lower bool) bool {
	if bound == nil {
		return false
	}
	return len(bound.prerelease) > 0 && (lower || bound.operator != "<" || len(bound.prerelease) != 1 || bound.prerelease[0] != "0")
}

func hasPrereleaseTuple(comparators []versionComparator, bound *versionComparator) bool {
	for _, c := range comparators {
		if c.version.Compare(&bound.version) == 0 && len(c.prerelease) > 0 {
			return true
		}
	}
	return false
}

func (left versionComparator) compare(right versionComparator) int {
	if cmp := left.version.Compare(&right.version); cmp != 0 {
		return cmp
	}
	return semver.ComparePreReleaseIdentifiers(left.prerelease, right.prerelease)
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
	cmp := left.compare(right)
	if left.operator == "=" {
		if right.operator == "=" {
			return cmp == 0
		}
		// A singleton prerelease needs a prerelease comparator on the same
		// major/minor/patch tuple, even when the numeric bounds contain it.
		if len(left.prerelease) > 0 && (len(right.prerelease) == 0 || left.version.Compare(&right.version) != 0) {
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

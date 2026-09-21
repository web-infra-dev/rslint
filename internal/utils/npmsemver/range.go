package npmsemver

import (
	"math"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/semver"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

// Range is an immutable npm-compatible version range, normalized by tsgo.
// Its zero value is an empty range. Parse must succeed before using a range.
type Range struct {
	alternatives [][]versionComparator
	raw          string
}

// Raw returns npm's whitespace-normalized input, before comparator expansion.
func (version Range) Raw() string { return version.raw }

type versionComparator struct {
	operator   string
	version    [3]uint64
	prerelease []string
}

const maxRangeComponent = 1<<53 - 1

// The patterns are repository-authored npm range syntax, not user regexps.
var versionPrefix = regexp.MustCompile(`(^|[\s|<>=~^])v([0-9xX*])`)
var versionOperatorSpace = regexp.MustCompile(`([<>=~^])\s+`)
var versionWildcardTail = regexp.MustCompile(`(?:^|\.)[xX*]\.[0-9]`)
var versionPrerelease = regexp.MustCompile(`^(?:0|[1-9][0-9]*|[0-9]*[A-Za-z-][0-9A-Za-z-]*)(?:\.(?:0|[1-9][0-9]*|[0-9]*[A-Za-z-][0-9A-Za-z-]*))*$`)

// Parse reuses tsgo range expansion with npm whitespace, integer and prerelease semantics.
func Parse(text string) (Range, bool) {
	text = strings.Map(func(r rune) rune {
		if ecmascript.IsWhiteSpaceOrLineTerminator(r) {
			return ' '
		}
		return r
	}, ecmascript.StringTrim(text))
	raw := strings.Join(strings.Fields(text), " ")
	// npm's range grammar is ASCII after JS whitespace normalization. Do not
	// let Go's broader whitespace handling accept characters such as U+0085.
	for _, character := range text {
		if character > 127 {
			return Range{}, false
		}
	}
	text = versionPrefix.ReplaceAllString(text, "${1}${2}")
	text = versionOperatorSpace.ReplaceAllString(text, "$1")
	text = strings.ReplaceAll(text, "~>", "~")
	// npm treats an empty union arm as *, while tsgo drops it.
	arms := strings.Split(text, "||")
	var largeComponents map[uint64]uint64
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
				return Range{}, false
			}
			if hasPrerelease && !versionPrerelease.MatchString(prerelease) {
				return Range{}, false
			}
			// tsgo expands ranges using uint32 components. Encode large values
			// in the upper half of that space, reserving their successor too:
			// expansion only copies, zeroes or increments a component. Restore
			// the original values before comparing any bounds. Encoding every
			// value above MaxInt32 prevents collisions with ordinary components
			// and their successors; zero/nonzero caret semantics stay intact.
			if len(base) < 10 {
				continue
			}
			parts := strings.Split(base, ".")
			changed := false
			for k, part := range parts {
				if part == "x" || part == "X" || part == "*" {
					break // The compiler ignores all following numeric components.
				}
				if len(part) < 10 || part[0] == '0' {
					continue
				}
				value, err := strconv.ParseUint(part, 10, 64)
				if err == nil && value > math.MaxInt32 {
					if value > maxRangeComponent {
						return Range{}, false
					}
					encoded := uint64(math.MaxInt32) + 2 + uint64(len(largeComponents))
					if encoded >= math.MaxUint32 {
						return Range{}, false
					}
					if largeComponents == nil {
						largeComponents = make(map[uint64]uint64)
					}
					largeComponents[encoded] = value
					largeComponents[encoded+1] = value + 1
					parts[k] = strconv.FormatUint(encoded, 10)
					changed = true
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
	result := Range{raw: raw}
	var emptyAlternative []versionComparator
	for _, arm := range arms {
		terms := strings.Fields(arm)
		hyphenRange := len(terms) == 3 && terms[1] == "-"
		if hyphenRange {
			terms = []string{arm}
		}
		var comparators []versionComparator
		for _, term := range terms {
			// Keep grammar validation and range expansion in the compiler.
			// Expand terms separately so generated bounds retain their origin:
			// npm's ^, ~ and partial hyphen upper bounds end in -0, unlike
			// an explicit <1.2.3, which can admit that tuple's prereleases.
			parsed, ok := semver.TryParseVersionRange(term)
			if !ok {
				return Range{}, false
			}
			for _, token := range strings.Fields(parsed.String()) {
				if token == "*" {
					continue
				}
				version := strings.TrimLeft(token, "<>=")
				operator := strings.TrimSuffix(token, version)
				// tsgo includes prereleases in wildcard lower bounds. npm starts
				// those at the stable version, unless -0 was explicitly requested.
				if operator == ">=" && !strings.Contains(term, version) {
					version = strings.TrimSuffix(version, "-0")
				}
				if operator == "<" && (hyphenRange || strings.HasPrefix(term, "^") || strings.HasPrefix(term, "~")) {
					version += "-0"
				}
				comparator, ok := parseVersionComparator(version, largeComponents)
				if !ok {
					return Range{}, false
				}
				comparator.operator = operator
				comparators = append(comparators, comparator)
			}
		}
		// npm drops its canonical null-set arm from unions unless it is
		// the only remaining arm. Other contradictory arms stay intact.
		if len(comparators) == 1 && comparators[0].operator == "<" && comparators[0].version == [3]uint64{} && slices.Equal(comparators[0].prerelease, []string{"0"}) {
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

func parseVersionComparator(text string, largeComponents map[uint64]uint64) (versionComparator, bool) {
	text, build, hasBuild := strings.Cut(text, "+")
	base, prerelease, hasPrerelease := strings.Cut(text, "-")
	if hasBuild {
		base += "+" + build
	}
	if _, err := semver.TryParseVersion(base); err != nil {
		return versionComparator{}, false
	}
	// Reuse compiler validation, retaining npm's full integers for comparisons.
	// Missing minor/patch components stay zero, as in the compiler.
	base, _, _ = strings.Cut(base, "+")
	var result versionComparator
	i := 0
	for part := range strings.SplitSeq(base, ".") {
		value, _ := strconv.ParseUint(part, 10, 64)
		if original, ok := largeComponents[value]; ok {
			value = original
		}
		if value > maxRangeComponent {
			return versionComparator{}, false
		}
		result.version[i] = value
		i++
	}
	if hasPrerelease {
		// TryParseVersion rejects npm prereleases such as "1beta". Retain
		// them separately and reuse the compiler's identifier comparison.
		result.prerelease = strings.Split(prerelease, ".")
	}
	return result, true
}

// IsAtLeast reports whether the configured range has no intersection with
// versions below since, matching the replacement selection in eslint-plugin-n.
func (version Range) IsAtLeast(since string) bool {
	boundary, ok := parseVersionComparator(since, nil)
	if !ok {
		return false
	}
	boundary.operator = "<"
	for _, alternative := range version.alternatives {
		constraints := append(append([]versionComparator{}, alternative...), boundary)
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
// range. Unlike IsAtLeast, it handles gaps between releases that gained a
// feature, and excludes prereleases absent from the supported range.
func (version Range) IsSubsetOf(supported string) bool {
	domain, ok := Parse(supported)
	if !ok {
		return false
	}
	return version.IsSubsetOfRange(domain)
}

// IsSubsetOfRange compares already parsed ranges using the same npm subset
// semantics as IsSubsetOf. Callers can reuse immutable support ranges.
func (version Range) IsSubsetOfRange(domain Range) bool {
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
		if c.version == bound.version && len(c.prerelease) > 0 {
			return true
		}
	}
	return false
}

func (left versionComparator) compare(right versionComparator) int {
	if cmp := slices.Compare(left.version[:], right.version[:]); cmp != 0 {
		return cmp
	}
	return semver.ComparePreReleaseIdentifiers(left.prerelease, right.prerelease)
}

func comparatorsIntersect(left, right versionComparator) bool {
	for _, comparator := range []versionComparator{left, right} {
		if comparator.operator == "<" && comparator.version == [3]uint64{} {
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
		if len(left.prerelease) > 0 && (len(right.prerelease) == 0 || left.version != right.version) {
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

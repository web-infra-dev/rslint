package no_deprecated_api

import (
	"strconv"
	"strings"
)

// This file implements the minimal subset of npm `semver` Range semantics that
// `n/no-deprecated-api` relies on: parsing a range string into OR-ed intervals
// and testing whether two ranges intersect.
//
// Upstream uses `new semver.Range(x)` + `range.intersects(other)` in two places:
//  1. getConfiguredNodeVersion -> the user's `version` option / default ">=16.0.0".
//  2. toReplaceMessage -> `version.intersects(getSemverRange("<" + supported))`
//     to decide whether a replacement API is already available on the target
//     Node version.
//
// Only `< <= > >= =`, bare/partial versions, `^`, `~`, hyphen ranges, and `||`
// are supported — enough for the `version` option, `engines.node`-style ranges,
// and the internal `<supported` comparisons. Prerelease/build metadata is
// ignored (Node release lines don't use it here).

type semverVersion struct {
	major, minor, patch int
}

func cmpVersion(a, b semverVersion) int {
	switch {
	case a.major != b.major:
		if a.major < b.major {
			return -1
		}
		return 1
	case a.minor != b.minor:
		if a.minor < b.minor {
			return -1
		}
		return 1
	case a.patch != b.patch:
		if a.patch < b.patch {
			return -1
		}
		return 1
	default:
		return 0
	}
}

// semverInterval is a single comparator set reduced to a [low, high] interval.
type semverInterval struct {
	hasLow   bool
	low      semverVersion
	lowIncl  bool
	hasHigh  bool
	high     semverVersion
	highIncl bool
	empty    bool
}

// semverRange is the OR of one or more intervals.
type semverRange struct {
	sets []semverInterval
}

func (iv *semverInterval) applyLow(v semverVersion, incl bool) {
	if !iv.hasLow {
		iv.hasLow, iv.low, iv.lowIncl = true, v, incl
		return
	}
	c := cmpVersion(v, iv.low)
	if c > 0 || (c == 0 && !incl && iv.lowIncl) {
		iv.low, iv.lowIncl = v, incl
	}
}

func (iv *semverInterval) applyHigh(v semverVersion, incl bool) {
	if !iv.hasHigh {
		iv.hasHigh, iv.high, iv.highIncl = true, v, incl
		return
	}
	c := cmpVersion(v, iv.high)
	if c < 0 || (c == 0 && !incl && iv.highIncl) {
		iv.high, iv.highIncl = v, incl
	}
}

// parsePartialVersion parses "1.2.3" / "1.2" / "1" / "1.x" / "*" into a version
// and the count of explicit numeric segments (0 means "any"). ok=false on a
// genuinely malformed segment.
func parsePartialVersion(s string) (v semverVersion, segs int, ok bool) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "v")
	if i := strings.IndexAny(s, "-+"); i >= 0 {
		s = s[:i]
	}
	if s == "" || s == "*" || s == "x" || s == "X" {
		return semverVersion{}, 0, true
	}
	parts := strings.Split(s, ".")
	for i, p := range parts {
		if i > 2 {
			break
		}
		if p == "x" || p == "X" || p == "*" {
			break
		}
		n, err := strconv.Atoi(p)
		if err != nil {
			return semverVersion{}, 0, false
		}
		switch i {
		case 0:
			v.major = n
		case 1:
			v.minor = n
		case 2:
			v.patch = n
		}
		segs++
	}
	return v, segs, true
}

// applyComparatorToken applies one comparator token (e.g. ">=16.0.0", "^1.2",
// "6.0.0", "~3") to the interval being built. Returns false if the token is
// malformed.
func applyComparatorToken(iv *semverInterval, tok string) bool {
	tok = strings.TrimSpace(tok)
	if tok == "" {
		return true
	}

	op := ""
	rest := tok
	switch {
	case strings.HasPrefix(rest, ">="):
		op, rest = ">=", rest[2:]
	case strings.HasPrefix(rest, "<="):
		op, rest = "<=", rest[2:]
	case strings.HasPrefix(rest, ">"):
		op, rest = ">", rest[1:]
	case strings.HasPrefix(rest, "<"):
		op, rest = "<", rest[1:]
	case strings.HasPrefix(rest, "="):
		op, rest = "=", rest[1:]
	case strings.HasPrefix(rest, "^"):
		op, rest = "^", rest[1:]
	case strings.HasPrefix(rest, "~"):
		op, rest = "~", rest[1:]
	}

	v, segs, ok := parsePartialVersion(rest)
	if !ok {
		return false
	}

	switch op {
	case ">=":
		iv.applyLow(v, true)
	case ">":
		iv.applyLow(v, false)
	case "<=":
		iv.applyHigh(v, true)
	case "<":
		iv.applyHigh(v, false)
	case "^":
		iv.applyLow(v, true)
		iv.applyHigh(caretUpper(v), false)
	case "~":
		iv.applyLow(v, true)
		iv.applyHigh(tildeUpper(v, segs), false)
	default: // "=" or bare/partial version
		if segs == 0 {
			// "*" / "x" → any version
			iv.applyLow(semverVersion{}, true)
		} else if segs >= 3 {
			iv.applyLow(v, true)
			iv.applyHigh(v, true)
		} else if segs == 2 {
			iv.applyLow(semverVersion{v.major, v.minor, 0}, true)
			iv.applyHigh(semverVersion{v.major, v.minor + 1, 0}, false)
		} else { // segs == 1
			iv.applyLow(semverVersion{v.major, 0, 0}, true)
			iv.applyHigh(semverVersion{v.major + 1, 0, 0}, false)
		}
	}
	return true
}

// caretUpper returns the exclusive upper bound of a `^` range.
func caretUpper(v semverVersion) semverVersion {
	switch {
	case v.major > 0:
		return semverVersion{v.major + 1, 0, 0}
	case v.minor > 0:
		return semverVersion{0, v.minor + 1, 0}
	default:
		return semverVersion{0, 0, v.patch + 1}
	}
}

// tildeUpper returns the exclusive upper bound of a `~` range.
func tildeUpper(v semverVersion, segs int) semverVersion {
	if segs <= 1 {
		return semverVersion{v.major + 1, 0, 0}
	}
	return semverVersion{v.major, v.minor + 1, 0}
}

// parseSemverRange mirrors upstream's getSemverRange: returns (range, true) for
// a valid range string, or (zero, false) — equivalent to `null` — otherwise.
func parseSemverRange(s string) (semverRange, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		// new semver.Range("") matches ">=0.0.0" (everything).
		return semverRange{sets: []semverInterval{{hasLow: true, low: semverVersion{}, lowIncl: true}}}, true
	}

	var r semverRange
	for _, setStr := range strings.Split(s, "||") {
		setStr = strings.TrimSpace(setStr)
		var iv semverInterval

		// Hyphen range: "A - B" → >=A <=B.
		if hi := strings.Index(setStr, " - "); hi >= 0 {
			lo := strings.TrimSpace(setStr[:hi])
			up := strings.TrimSpace(setStr[hi+3:])
			loV, _, ok1 := parsePartialVersion(lo)
			upV, upSegs, ok2 := parsePartialVersion(up)
			if !ok1 || !ok2 {
				return semverRange{}, false
			}
			iv.applyLow(loV, true)
			// Partial upper bound in a hyphen range rounds up (npm semver).
			if upSegs >= 3 {
				iv.applyHigh(upV, true)
			} else if upSegs == 2 {
				iv.applyHigh(semverVersion{upV.major, upV.minor + 1, 0}, false)
			} else {
				iv.applyHigh(semverVersion{upV.major + 1, 0, 0}, false)
			}
			r.sets = append(r.sets, iv)
			continue
		}

		for _, tok := range strings.Fields(setStr) {
			if !applyComparatorToken(&iv, tok) {
				return semverRange{}, false
			}
		}
		// Detect contradictory sets (e.g. ">=5 <3").
		if iv.hasLow && iv.hasHigh {
			c := cmpVersion(iv.low, iv.high)
			if c > 0 || (c == 0 && (!iv.lowIncl || !iv.highIncl)) {
				iv.empty = true
			}
		}
		r.sets = append(r.sets, iv)
	}
	return r, true
}

// intervalsOverlap reports whether two intervals share at least one version.
func intervalsOverlap(a, b semverInterval) bool {
	if a.empty || b.empty {
		return false
	}
	if a.hasLow && b.hasHigh {
		c := cmpVersion(a.low, b.high)
		if c > 0 || (c == 0 && (!a.lowIncl || !b.highIncl)) {
			return false
		}
	}
	if b.hasLow && a.hasHigh {
		c := cmpVersion(b.low, a.high)
		if c > 0 || (c == 0 && (!b.lowIncl || !a.highIncl)) {
			return false
		}
	}
	return true
}

// intersects reports whether this range and other share at least one version,
// mirroring npm semver's `Range#intersects`.
func (r semverRange) intersects(other semverRange) bool {
	for _, a := range r.sets {
		for _, b := range other.sets {
			if intervalsOverlap(a, b) {
				return true
			}
		}
	}
	return false
}

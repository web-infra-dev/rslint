// This file exists only because the case tables this package builds on and
// the ones Node reads are different editions of Unicode, and this package has
// to answer the way Node does.
//
// Go 1.26's unicode tables are derived from Unicode 15.0, and the caser behind
// [StringToUppercase] from Unicode 15.1. Node 26 carries ICU 78, which is
// Unicode 17.0. Across the editions in between, eight characters that had no
// case mapping at all were given one, and two bicameral scripts arrived whole:
// Garay in Unicode 16.0 and Beria Erfe in Unicode 17.0. Neither table knows any
// of them, so two characters JavaScript compares equal under a regexp's `iu`
// flags compare unequal here — the whole question [Canonicalize] is asked — and
// [StringToUppercase] would hand back a string JavaScript would have changed.
//
// The delta is written out rather than generated: the pairs and the runs state
// their target in the source, where a reader can check them, and building the
// linter then needs neither a Node installation nor a table generator.
//
// What the delta cannot reach is the one place the caser reads a character's
// surroundings rather than its mapping: a capital sigma is only final when a
// cased character comes before it, and to the older edition none of these
// characters is cased yet. So `Α\U00010D70Σ` lowercases with a plain sigma
// here and with a final one in Node. Every other answer agrees, character for
// character, across the whole of Unicode.
//
// TODO(go1.27): delete this file and unicode17_test.go, then drop the calls
// into them — the lookup in [Canonicalize], the two in [unicode17Remap]'s
// callers, and the seeding loop in caseTables. golang.org/x/text v0.40.0
// already ships Unicode 17.0.0 tables behind a `//go:build go1.27` constraint,
// so the standard library is expected to reach 17.0.0 in that release.
// TestUnicode17DeltaStillNeeded fails once the toolchain gets there, so this
// does not rely on anyone remembering.

package ecmascript

// unicode17ToUpper returns the simple uppercase mapping Unicode 16 and 17 gave
// r, reporting false when Go's own tables already hold the answer.
func unicode17ToUpper(r rune) (rune, bool) {
	for _, pair := range unicode17CasePairs {
		if r == pair[0] {
			return pair[1], true
		}
	}
	for _, run := range unicode17CaseRuns {
		if r >= run.lower && r <= run.lastLower {
			return r + run.toUpper, true
		}
	}
	return r, false
}

// unicode17ToLower is [unicode17ToUpper] the other way round.
func unicode17ToLower(r rune) (rune, bool) {
	for _, pair := range unicode17CasePairs {
		if r == pair[1] {
			return pair[0], true
		}
	}
	for _, run := range unicode17CaseRuns {
		if lower := r - run.toUpper; lower >= run.lower && lower <= run.lastLower {
			return lower, true
		}
	}
	return r, false
}

// unicode17Fold returns the other member of the pair r folds together with
// under Unicode 16 and 17, reporting false when Go's own tables already hold
// the answer. Each pair is a lowercase character and its uppercase, which
// simple case folding puts in one orbit.
func unicode17Fold(r rune) (rune, bool) {
	if upper, ok := unicode17ToUpper(r); ok {
		return upper, true
	}
	return unicode17ToLower(r)
}

// unicode17CaseAdditions returns every character the delta names, lowercase and
// uppercase alike. caseTables walks unicode.CaseRanges to find the characters
// that canonicalize onto one another, and cannot reach these: Go holds no case
// mapping for them at all, so no case range names them.
func unicode17CaseAdditions() []rune {
	var runes []rune
	for _, pair := range unicode17CasePairs {
		runes = append(runes, pair[0], pair[1])
	}
	for _, run := range unicode17CaseRuns {
		for lower := run.lower; lower <= run.lastLower; lower++ {
			runes = append(runes, lower, lower+run.toUpper)
		}
	}
	return runes
}

// unicode17CasePairs maps a character to the uppercase Unicode 16 or 17 gave
// it, in the order {lower, upper}.
var unicode17CasePairs = [...][2]rune{
	{0x019B, 0xA7DC},
	{0x0264, 0xA7CB},
	{0x1C8A, 0x1C89},
	{0xA7CD, 0xA7CC},
	{0xA7CF, 0xA7CE},
	{0xA7D3, 0xA7D2},
	{0xA7D5, 0xA7D4},
	{0xA7DB, 0xA7DA},
}

// unicode17CaseRuns says the same for the two scripts that arrived whole. Each
// is a run of lowercase characters whose uppercase sits a fixed distance away,
// so the run states that distance once rather than repeating forty-seven pairs.
var unicode17CaseRuns = [...]struct {
	lower, lastLower, toUpper rune
}{
	{0x10D70, 0x10D85, -0x20}, // Garay
	{0x16EBB, 0x16ED3, -0x1B}, // Beria Erfe
}

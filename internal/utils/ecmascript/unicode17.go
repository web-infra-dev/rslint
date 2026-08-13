// This file exists only because Go and Node read different editions of
// Unicode, and [Canonicalize] has to answer the way Node does.
//
// Go 1.26's unicode tables are derived from Unicode 15.0. Node 26 carries
// ICU 78, which is Unicode 17.0. Across the three editions in between,
// eight characters that Unicode 15 gave no simple uppercase mapping at all
// were given one. unicode.ToUpper leaves each of those where it is, so two
// characters JavaScript compares equal under a regexp's `i` flag compare
// unequal here — the whole question [Canonicalize] is asked.
//
// The delta is written out rather than generated: eight pairs state the
// target in the source, where a reader can check them, and building the
// linter then needs neither a Node installation nor a table generator.
//
// TODO(go1.27): delete this file and unicode17_test.go, then drop the two
// calls into them — the lookup in [Canonicalize] and the seeding loop in
// caseTables. golang.org/x/text v0.40.0 already ships Unicode 17.0.0 tables
// behind a `//go:build go1.27` constraint, so the standard library is
// expected to reach 17.0.0 in that release, at which point unicode.ToUpper
// covers these eight on its own. TestUnicode17DeltaStillNeeded fails once
// the toolchain gets there, so this does not rely on anyone remembering.

package ecmascript

// unicode17ToUpper returns the simple uppercase mapping Unicode 16 and 17 gave
// r, reporting false when Go's own tables already hold the answer.
func unicode17ToUpper(r rune) (rune, bool) {
	for _, pair := range unicode17CasePairs {
		if r == pair[0] {
			return pair[1], true
		}
	}
	return r, false
}

// unicode17CaseAdditions returns every character the delta names, lowercase and
// uppercase alike. caseTables walks unicode.CaseRanges to find the characters
// that canonicalize onto one another, and cannot reach these: Go holds no case
// mapping for them at all, so no case range names them.
func unicode17CaseAdditions() []rune {
	runes := make([]rune, 0, 2*len(unicode17CasePairs))
	for _, pair := range unicode17CasePairs {
		runes = append(runes, pair[0], pair[1])
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

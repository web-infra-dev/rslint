// Package unicode17 carries the character data Unicode 16.0 and 17.0 added on
// top of the edition the Go toolchain ships, so that a port of a JavaScript
// operation answers the way Node does rather than the way Go's tables do.
//
// Go 1.26's unicode tables are derived from Unicode 15.0, as are the tables
// behind golang.org/x/text's caser. Node 26 carries ICU 78, which is Unicode
// 17.0. Across the editions in between, ten characters that had no case mapping
// at all were given one, two bicameral scripts arrived whole, the two
// properties that decide where a Greek capital sigma lowercases to a final sigma
// — Cased and Case_Ignorable — took on a hundred and ninety more characters and
// lost two, and the general categories a rule asks a character about grew by
// nine thousand five hundred letters and ninety-three marks.
//
// Without the data here, two characters JavaScript compares equal under a
// regexp's `iu` flags compare unequal, a string comes back from a case mapping
// JavaScript would have changed, a final sigma lands on the wrong one of ς and
// σ, and a rule reading `\p{Lu}` or `\p{M}` off a character sees a letter from
// a new script as no letter at all.
//
// The data is written out rather than generated: every entry states its target
// in the source, where a reader can check it against the Unicode Character
// Database, and building the linter then needs neither a Node installation nor
// a table generator.
//
// TODO(go1.27): delete this package. The compiler names every caller once it is
// gone: a case lookup falls through to the standard library when its second
// return value is false, and a category question here stands in for the
// unicode function of the same name. golang.org/x/text v0.40.0 already ships Unicode 17.0.0
// tables behind a `//go:build go1.27` constraint, so the standard library is
// expected to reach 17.0.0 in that release. TestDeltaStillNeeded fails once the
// toolchain gets there, so this does not rely on anyone remembering.
package unicode17

import "unicode"

// ToUpper returns the simple uppercase mapping Unicode 16 or 17 gave r,
// reporting false when Go's own tables already hold the answer.
func ToUpper(r rune) (rune, bool) {
	for _, pair := range casePairs {
		if r == pair[0] {
			return pair[1], true
		}
	}
	for _, run := range caseRuns {
		if r >= run.lower && r <= run.lastLower {
			return r + run.toUpper, true
		}
	}
	return r, false
}

// ToLower is [ToUpper] the other way round.
func ToLower(r rune) (rune, bool) {
	for _, pair := range casePairs {
		if r == pair[1] {
			return pair[0], true
		}
	}
	for _, run := range caseRuns {
		if lower := r - run.toUpper; lower >= run.lower && lower <= run.lastLower {
			return lower, true
		}
	}
	return r, false
}

// Fold returns the other member of the pair r folds together with under
// Unicode 16 and 17, reporting false when Go's own tables already hold the
// answer. Each pair is a lowercase character and its uppercase, which simple
// case folding puts in one orbit.
func Fold(r rune) (rune, bool) {
	if upper, ok := ToUpper(r); ok {
		return upper, true
	}
	return ToLower(r)
}

// CaseAdditions returns every character the mapping data names, lowercase and
// uppercase alike. A caller walking unicode.CaseRanges to find the characters
// that map onto one another cannot reach these: Go holds no case mapping for
// them at all, so no case range names them.
func CaseAdditions() []rune {
	var runes []rune
	for _, pair := range casePairs {
		runes = append(runes, pair[0], pair[1])
	}
	for _, run := range caseRuns {
		for lower := run.lower; lower <= run.lastLower; lower++ {
			runes = append(runes, lower, lower+run.toUpper)
		}
	}
	return runes
}

// Cased returns Unicode 17's answer for the Cased property of r, reporting
// false when the editions agree and the caller's own derivation stands.
func Cased(r rune) (bool, bool) {
	switch {
	case unicode.Is(casedAdded, r):
		return true, true
	case unicode.Is(casedRemoved, r):
		return false, true
	}
	return false, false
}

// CaseIgnorable says for Case_Ignorable what [Cased] says for Cased.
func CaseIgnorable(r rune) (bool, bool) {
	switch {
	case unicode.Is(caseIgnorableAdded, r):
		return true, true
	case unicode.Is(caseIgnorableRemoved, r):
		return false, true
	}
	return false, false
}

// IsUpper reports whether r is an uppercase letter — the general category Lu —
// as Unicode 17 has it. It is unicode.IsUpper with the letters the two editions
// added, so a caller reads one answer rather than a table and a delta.
func IsUpper(r rune) bool {
	return unicode.IsUpper(r) || unicode.Is(upperAdded, r)
}

// IsLower says for the category Ll what [IsUpper] says for Lu. One character
// went the other way, so this takes one away as well as adding.
func IsLower(r rune) bool {
	if unicode.Is(lowerRemoved, r) {
		return false
	}
	return unicode.IsLower(r) || unicode.Is(lowerAdded, r)
}

// IsLetter says for the category L what [IsUpper] says for Lu. Most of what it
// adds is the scripts the two editions brought in whole.
func IsLetter(r rune) bool {
	return unicode.IsLetter(r) || unicode.Is(letterAdded, r)
}

// IsMark says for the category M — the combining marks — what [IsUpper] says
// for Lu.
func IsMark(r rune) bool {
	return unicode.Is(unicode.M, r) || unicode.Is(markAdded, r)
}

// casePairs maps a character to the uppercase Unicode 16 or 17 gave it, in the
// order {lower, upper}.
var casePairs = [...][2]rune{
	{0x019B, 0xA7DC},
	{0x0264, 0xA7CB},
	{0x1C8A, 0x1C89},
	{0xA7CD, 0xA7CC},
	{0xA7CF, 0xA7CE},
	{0xA7D3, 0xA7D2},
	{0xA7D5, 0xA7D4},
	{0xA7DB, 0xA7DA},
}

// caseRuns says the same for the two scripts that arrived whole. Each is a run
// of lowercase characters whose uppercase sits a fixed distance away, so the
// run states that distance once rather than repeating forty-seven pairs.
var caseRuns = [...]struct {
	lower, lastLower, toUpper rune
}{
	{0x10D70, 0x10D85, -0x20}, // Garay
	{0x16EBB, 0x16ED3, -0x1B}, // Beria Erfe
}

// casedAdded holds the characters Unicode 16 and 17 made cased: the letters of
// the two new bicameral scripts, and the capitals the editions gave to letters
// that had none. A cased character before a capital sigma is half of what makes
// that sigma final.
var casedAdded = &unicode.RangeTable{
	R16: []unicode.Range16{
		{0x1C89, 0x1C8A, 1},
		{0xA7CB, 0xA7CF, 1},
		{0xA7D2, 0xA7D2, 1},
		{0xA7D4, 0xA7D4, 1},
		{0xA7DA, 0xA7DC, 1},
		{0xA7F1, 0xA7F1, 1},
	},
	R32: []unicode.Range32{
		{0x10D50, 0x10D65, 1},
		{0x10D70, 0x10D85, 1},
		{0x16EA0, 0x16EB8, 1},
		{0x16EBB, 0x16ED3, 1},
	},
}

// casedRemoved holds the one character that stopped being cased: Unicode 16
// reclassified U+0295 from a lowercase letter to a caseless one.
var casedRemoved = &unicode.RangeTable{
	R16: []unicode.Range16{
		{0x0295, 0x0295, 1},
	},
}

// caseIgnorableAdded holds the characters Unicode 16 and 17 made
// case-ignorable — combining marks, modifier letters and the like, most of them
// from the scripts the two editions added. A capital sigma looks past these
// when asking what comes before and after it.
var caseIgnorableAdded = &unicode.RangeTable{
	R16: []unicode.Range16{
		{0x0897, 0x0897, 1},
		{0x1ACF, 0x1ADD, 1},
		{0x1AE0, 0x1AEB, 1},
		{0xA7F1, 0xA7F1, 1},
	},
	R32: []unicode.Range32{
		{0x10D4E, 0x10D4E, 1},
		{0x10D69, 0x10D6D, 1},
		{0x10D6F, 0x10D6F, 1},
		{0x10EC5, 0x10EC5, 1},
		{0x10EFA, 0x10EFC, 1},
		{0x113BB, 0x113C0, 1},
		{0x113CE, 0x113CE, 1},
		{0x113D0, 0x113D0, 1},
		{0x113D2, 0x113D2, 1},
		{0x113E1, 0x113E2, 1},
		{0x11B60, 0x11B60, 1},
		{0x11B62, 0x11B64, 1},
		{0x11B66, 0x11B66, 1},
		{0x11DD9, 0x11DD9, 1},
		{0x11F5A, 0x11F5A, 1},
		{0x1611E, 0x16129, 1},
		{0x1612D, 0x1612F, 1},
		{0x16D40, 0x16D42, 1},
		{0x16D6B, 0x16D6C, 1},
		{0x16FF2, 0x16FF3, 1},
		{0x1E5EE, 0x1E5EF, 1},
		{0x1E6E3, 0x1E6E3, 1},
		{0x1E6E6, 0x1E6E6, 1},
		{0x1E6EE, 0x1E6EF, 1},
		{0x1E6F5, 0x1E6F5, 1},
		{0x1E6FF, 0x1E6FF, 1},
	},
}

// caseIgnorableRemoved holds the one character that stopped being
// case-ignorable: Unicode 16 reclassified U+1171E from a nonspacing mark to a
// spacing one.
var caseIgnorableRemoved = &unicode.RangeTable{
	R32: []unicode.Range32{
		{0x1171E, 0x1171E, 1},
	},
}

// upperAdded holds the characters Unicode 16 and 17 made uppercase letters:
// the capitals of the two new bicameral scripts, and the ones the editions gave
// to letters that had none.
var upperAdded = &unicode.RangeTable{
	R16: []unicode.Range16{
		{0x1C89, 0x1C89, 1},
		{0xA7CB, 0xA7CC, 1},
		{0xA7CE, 0xA7CE, 1},
		{0xA7D2, 0xA7D2, 1},
		{0xA7D4, 0xA7D4, 1},
		{0xA7DA, 0xA7DA, 1},
		{0xA7DC, 0xA7DC, 1},
	},
	R32: []unicode.Range32{
		{0x10D50, 0x10D65, 1},
		{0x16EA0, 0x16EB8, 1},
	},
}

// lowerAdded holds the characters those same editions made lowercase letters,
// which is the other half of each of the pairs and scripts above.
var lowerAdded = &unicode.RangeTable{
	R16: []unicode.Range16{
		{0x1C8A, 0x1C8A, 1},
		{0xA7CD, 0xA7CD, 1},
		{0xA7CF, 0xA7CF, 1},
		{0xA7DB, 0xA7DB, 1},
	},
	R32: []unicode.Range32{
		{0x10D70, 0x10D85, 1},
		{0x16EBB, 0x16ED3, 1},
	},
}

// lowerRemoved holds the one character that stopped being a lowercase letter:
// Unicode 16 reclassified U+0295 as a letter with no case at all.
var lowerRemoved = &unicode.RangeTable{
	R16: []unicode.Range16{
		{0x0295, 0x0295, 1},
	},
}

// letterAdded holds every character Unicode 16 and 17 made a letter. Nearly all
// of it is the scripts the two editions added whole, along with the ideographs
// each extended the Han blocks by.
var letterAdded = &unicode.RangeTable{
	R16: []unicode.Range16{
		{0x088F, 0x088F, 1},
		{0x0C5C, 0x0C5C, 1},
		{0x0CDC, 0x0CDC, 1},
		{0x1C89, 0x1C8A, 1},
		{0xA7CB, 0xA7CF, 1},
		{0xA7D2, 0xA7D2, 1},
		{0xA7D4, 0xA7D4, 1},
		{0xA7DA, 0xA7DC, 1},
		{0xA7F1, 0xA7F1, 1},
	},
	R32: []unicode.Range32{
		{0x105C0, 0x105F3, 1},
		{0x10940, 0x10959, 1},
		{0x10D4A, 0x10D65, 1},
		{0x10D6F, 0x10D85, 1},
		{0x10EC2, 0x10EC7, 1},
		{0x11380, 0x11389, 1},
		{0x1138B, 0x1138B, 1},
		{0x1138E, 0x1138E, 1},
		{0x11390, 0x113B5, 1},
		{0x113B7, 0x113B7, 1},
		{0x113D1, 0x113D1, 1},
		{0x113D3, 0x113D3, 1},
		{0x11BC0, 0x11BE0, 1},
		{0x11DB0, 0x11DDB, 1},
		{0x13460, 0x143FA, 1},
		{0x16100, 0x1611D, 1},
		{0x16D40, 0x16D6C, 1},
		{0x16EA0, 0x16EB8, 1},
		{0x16EBB, 0x16ED3, 1},
		{0x16FF2, 0x16FF3, 1},
		{0x187F8, 0x187FF, 1},
		{0x18CFF, 0x18CFF, 1},
		{0x18D09, 0x18D1E, 1},
		{0x18D80, 0x18DF2, 1},
		{0x1E5D0, 0x1E5ED, 1},
		{0x1E5F0, 0x1E5F0, 1},
		{0x1E6C0, 0x1E6DE, 1},
		{0x1E6E0, 0x1E6E2, 1},
		{0x1E6E4, 0x1E6E5, 1},
		{0x1E6E7, 0x1E6ED, 1},
		{0x1E6F0, 0x1E6F4, 1},
		{0x1E6FE, 0x1E6FF, 1},
		{0x2B73A, 0x2B73F, 1},
		{0x2CEA2, 0x2CEAD, 1},
		{0x2EBF0, 0x2EE5D, 1},
		{0x323B0, 0x33479, 1},
	},
}

// markAdded holds the combining marks the two editions added, most of them
// belonging to the scripts letterAdded names.
var markAdded = &unicode.RangeTable{
	R16: []unicode.Range16{
		{0x0897, 0x0897, 1},
		{0x1ACF, 0x1ADD, 1},
		{0x1AE0, 0x1AEB, 1},
	},
	R32: []unicode.Range32{
		{0x10D69, 0x10D6D, 1},
		{0x10EFA, 0x10EFC, 1},
		{0x113B8, 0x113C0, 1},
		{0x113C2, 0x113C2, 1},
		{0x113C5, 0x113C5, 1},
		{0x113C7, 0x113CA, 1},
		{0x113CC, 0x113D0, 1},
		{0x113D2, 0x113D2, 1},
		{0x113E1, 0x113E2, 1},
		{0x11B60, 0x11B67, 1},
		{0x11F5A, 0x11F5A, 1},
		{0x1611E, 0x1612F, 1},
		{0x1E5EE, 0x1E5EF, 1},
		{0x1E6E3, 0x1E6E3, 1},
		{0x1E6E6, 0x1E6E6, 1},
		{0x1E6EE, 0x1E6EF, 1},
		{0x1E6F5, 0x1E6F5, 1},
	},
}

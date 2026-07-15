package minimatch

import "unicode"

// Minimatch renders property-based POSIX classes as JavaScript Unicode
// property escapes. rslint's supported Node/CI baseline uses Unicode 16,
// while Go 1.26 ships Unicode 15 tables. Keep the small Unicode 16 delta here
// so config discovery does not vary between the Go and Node halves.
const searchUnicodeBaseVersion = "15.0.0"
const searchMinimatchUnicodeVersion = "16.0.0"

func isSearchUnicode16Letter(character rune) bool {
	return unicode.Is(unicode.L, character) || unicode.Is(unicode.Nl, character) ||
		searchRuneInRanges(character, searchUnicode16LetterAdditions)
}

func isSearchUnicode16DecimalDigit(character rune) bool {
	return unicode.Is(unicode.Nd, character) || searchRuneInRanges(character, searchUnicode16DecimalDigitAdditions)
}

func isSearchUnicode16Lower(character rune) bool {
	return unicode.Is(unicode.Ll, character) || searchRuneInRanges(character, searchUnicode16LowerAdditions)
}

func isSearchUnicode16Upper(character rune) bool {
	return unicode.Is(unicode.Lu, character) || searchRuneInRanges(character, searchUnicode16UpperAdditions)
}

func isSearchUnicode16Punctuation(character rune) bool {
	return unicode.Is(unicode.P, character) || searchRuneInRanges(character, searchUnicode16PunctuationAdditions)
}

func isSearchUnicode16Other(character rune) bool {
	return unicode.Is(unicode.C, character) && !searchRuneInRanges(character, searchUnicode16NewAssignments)
}

func searchRuneInRanges(character rune, ranges []searchRuneRange) bool {
	low := 0
	high := len(ranges) - 1
	for low <= high {
		middle := low + (high-low)/2
		candidate := ranges[middle]
		switch {
		case character < candidate.low:
			high = middle - 1
		case character > candidate.high:
			low = middle + 1
		default:
			return true
		}
	}
	return false
}

var searchUnicode16LetterAdditions = []searchRuneRange{
	{low: 0x1c89, high: 0x1c8a},
	{low: 0xa7cb, high: 0xa7cd},
	{low: 0xa7da, high: 0xa7dc},
	{low: 0x105c0, high: 0x105f3},
	{low: 0x10d4a, high: 0x10d65},
	{low: 0x10d6f, high: 0x10d85},
	{low: 0x10ec2, high: 0x10ec4},
	{low: 0x11380, high: 0x11389},
	{low: 0x1138b, high: 0x1138b},
	{low: 0x1138e, high: 0x1138e},
	{low: 0x11390, high: 0x113b5},
	{low: 0x113b7, high: 0x113b7},
	{low: 0x113d1, high: 0x113d1},
	{low: 0x113d3, high: 0x113d3},
	{low: 0x11bc0, high: 0x11be0},
	{low: 0x13460, high: 0x143fa},
	{low: 0x16100, high: 0x1611d},
	{low: 0x16d40, high: 0x16d6c},
	{low: 0x18cff, high: 0x18cff},
	{low: 0x1e5d0, high: 0x1e5ed},
	{low: 0x1e5f0, high: 0x1e5f0},
	{low: 0x2ebf0, high: 0x2ee5d},
}

var searchUnicode16DecimalDigitAdditions = []searchRuneRange{
	{low: 0x10d40, high: 0x10d49},
	{low: 0x116d0, high: 0x116e3},
	{low: 0x11bf0, high: 0x11bf9},
	{low: 0x16130, high: 0x16139},
	{low: 0x16d70, high: 0x16d79},
	{low: 0x1ccf0, high: 0x1ccf9},
	{low: 0x1e5f1, high: 0x1e5fa},
}

var searchUnicode16LowerAdditions = []searchRuneRange{
	{low: 0x1c8a, high: 0x1c8a},
	{low: 0xa7cd, high: 0xa7cd},
	{low: 0xa7db, high: 0xa7db},
	{low: 0x10d70, high: 0x10d85},
}

var searchUnicode16UpperAdditions = []searchRuneRange{
	{low: 0x1c89, high: 0x1c89},
	{low: 0xa7cb, high: 0xa7cc},
	{low: 0xa7da, high: 0xa7da},
	{low: 0xa7dc, high: 0xa7dc},
	{low: 0x10d50, high: 0x10d65},
}

var searchUnicode16PunctuationAdditions = []searchRuneRange{
	{low: 0x1b4e, high: 0x1b4f},
	{low: 0x1b7f, high: 0x1b7f},
	{low: 0x10d6e, high: 0x10d6e},
	{low: 0x113d4, high: 0x113d5},
	{low: 0x113d7, high: 0x113d8},
	{low: 0x11be1, high: 0x11be1},
	{low: 0x16d6d, high: 0x16d6f},
	{low: 0x1e5ff, high: 0x1e5ff},
}

// Every code point assigned between Unicode 15 and 16. Go 1.26 reports these
// as General_Category=Other/Unassigned; JavaScript Unicode 16 does not.
var searchUnicode16NewAssignments = []searchRuneRange{
	{low: 0x0897, high: 0x0897},
	{low: 0x1b4e, high: 0x1b4f},
	{low: 0x1b7f, high: 0x1b7f},
	{low: 0x1c89, high: 0x1c8a},
	{low: 0x2427, high: 0x2429},
	{low: 0x2ffc, high: 0x2fff},
	{low: 0x31e4, high: 0x31e5},
	{low: 0x31ef, high: 0x31ef},
	{low: 0xa7cb, high: 0xa7cd},
	{low: 0xa7da, high: 0xa7dc},
	{low: 0x105c0, high: 0x105f3},
	{low: 0x10d40, high: 0x10d65},
	{low: 0x10d69, high: 0x10d85},
	{low: 0x10d8e, high: 0x10d8f},
	{low: 0x10ec2, high: 0x10ec4},
	{low: 0x10efc, high: 0x10efc},
	{low: 0x11380, high: 0x11389},
	{low: 0x1138b, high: 0x1138b},
	{low: 0x1138e, high: 0x1138e},
	{low: 0x11390, high: 0x113b5},
	{low: 0x113b7, high: 0x113c0},
	{low: 0x113c2, high: 0x113c2},
	{low: 0x113c5, high: 0x113c5},
	{low: 0x113c7, high: 0x113ca},
	{low: 0x113cc, high: 0x113d5},
	{low: 0x113d7, high: 0x113d8},
	{low: 0x113e1, high: 0x113e2},
	{low: 0x116d0, high: 0x116e3},
	{low: 0x11bc0, high: 0x11be1},
	{low: 0x11bf0, high: 0x11bf9},
	{low: 0x11f5a, high: 0x11f5a},
	{low: 0x13460, high: 0x143fa},
	{low: 0x16100, high: 0x16139},
	{low: 0x16d40, high: 0x16d79},
	{low: 0x18cff, high: 0x18cff},
	{low: 0x1cc00, high: 0x1ccf9},
	{low: 0x1cd00, high: 0x1ceb3},
	{low: 0x1e5d0, high: 0x1e5fa},
	{low: 0x1e5ff, high: 0x1e5ff},
	{low: 0x1f8b2, high: 0x1f8bb},
	{low: 0x1f8c0, high: 0x1f8c1},
	{low: 0x1fa89, high: 0x1fa89},
	{low: 0x1fa8f, high: 0x1fa8f},
	{low: 0x1fabe, high: 0x1fabe},
	{low: 0x1fac6, high: 0x1fac6},
	{low: 0x1fadc, high: 0x1fadc},
	{low: 0x1fadf, high: 0x1fadf},
	{low: 0x1fae9, high: 0x1fae9},
	{low: 0x1fbcb, high: 0x1fbef},
	{low: 0x2ebf0, high: 0x2ee5d},
}

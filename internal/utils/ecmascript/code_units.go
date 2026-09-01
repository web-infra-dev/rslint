package ecmascript

import (
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

// StringCodeUnits reads s as the UTF-16 code units JavaScript indexes a string
// by: a character outside the basic plane is two of them, and a lone surrogate
// is the single unit it stands for rather than the replacement characters a
// UTF-8 reader answers with.
func StringCodeUnits(s string) []uint16 {
	units := make([]uint16, 0, len(s))
	for i := 0; i < len(s); {
		r, size := decodeStringRune(s[i:])
		if r > 0xFFFF {
			high, low := utf16.EncodeRune(r)
			units = append(units, uint16(high), uint16(low))
		} else {
			units = append(units, uint16(r))
		}
		i += size
	}
	return units
}

// StringFromCodeUnits writes the UTF-16 code units JavaScript stores as a
// string. Adjacent surrogate pairs use their ordinary UTF-8 spelling; an
// unpaired surrogate uses the WTF-8 spelling the compiler uses to preserve it.
func StringFromCodeUnits(units []uint16) string {
	var result strings.Builder
	result.Grow(len(units))
	for i := 0; i < len(units); i++ {
		unit := units[i]
		if unit >= 0xD800 && unit <= 0xDBFF && i+1 < len(units) {
			next := units[i+1]
			if next >= 0xDC00 && next <= 0xDFFF {
				result.WriteRune(utf16.DecodeRune(rune(unit), rune(next)))
				i++
				continue
			}
		}
		if unit >= 0xD800 && unit <= 0xDFFF {
			result.WriteByte(byte(0xE0 | unit>>12))
			result.WriteByte(byte(0x80 | unit>>6&0x3F))
			result.WriteByte(byte(0x80 | unit&0x3F))
			continue
		}
		result.WriteRune(rune(unit))
	}
	return result.String()
}

// CombineSurrogatePairs canonicalizes a JavaScript string assembled from
// separately scanned pieces. A high surrogate at the end of one piece and a
// low surrogate at the start of the next become one ordinary supplementary
// code point, while every unpaired surrogate keeps its compiler WTF-8 bytes.
func CombineSurrogatePairs(s string) string {
	if strings.IndexByte(s, 0xED) < 0 {
		return s
	}
	var result strings.Builder
	result.Grow(len(s))
	for i := 0; i < len(s); {
		r, size := decodeStringRune(s[i:])
		if r >= 0xD800 && r <= 0xDBFF {
			low, lowSize := decodeStringRune(s[i+size:])
			if low >= 0xDC00 && low <= 0xDFFF {
				result.WriteRune(utf16.DecodeRune(r, low))
				i += size + lowSize
				continue
			}
		}
		result.WriteString(s[i : i+size])
		i += size
	}
	return result.String()
}

// StringCodeUnitCount reports s's length in the UTF-16 code units counted by
// JavaScript String#length. Unlike core.UTF16Len, it also preserves lone
// surrogates in the WTF-8 form used by compiler string values.
func StringCodeUnitCount(s string) int {
	count := 0
	for i := 0; i < len(s); {
		r, size := decodeStringRune(s[i:])
		if r > 0xFFFF {
			count += 2
		} else {
			count++
		}
		i += size
	}
	return count
}

// CompareStrings ports the comparison JavaScript's relational operators make of
// two strings: code unit by code unit, with the shorter of the two first when
// one is a prefix of the other. It returns -1 if a sorts before b, 1 if it
// sorts after, and 0 if the two are the same string.
//
// The order is not the one Go's own string comparison gives. A character
// outside the basic plane is a pair of surrogates, which rank below U+E000
// through U+FFFF although the character itself ranks above them.
//
// https://tc39.es/ecma262/2024/multipage/abstract-operations.html#sec-islessthan
func CompareStrings(a, b string) int {
	if a == b {
		return 0
	}
	// Every other character is one code unit whose bytes already rank the way
	// it does, so the pair is answered without reading either string apart.
	if !hasAstralCharacter(a) && !hasAstralCharacter(b) {
		return strings.Compare(a, b)
	}
	left, right := StringCodeUnits(a), StringCodeUnits(b)
	for i := 0; i < len(left) && i < len(right); i++ {
		if left[i] != right[i] {
			if left[i] < right[i] {
				return -1
			}
			return 1
		}
	}
	switch {
	case len(left) < len(right):
		return -1
	case len(left) > len(right):
		return 1
	default:
		return 0
	}
}

// hasAstralCharacter reports whether s carries a character outside the basic
// plane, which UTF-8 writes with a lead byte of 0xF0 or above.
func hasAstralCharacter(s string) bool {
	for i := range len(s) {
		if s[i] >= 0xF0 {
			return true
		}
	}
	return false
}

// decodeStringRune reads the first character of s the way the compiler wrote
// it: UTF-8, except for a lone surrogate, which has no UTF-8 encoding of its
// own and is carried as the three bytes UTF-8 would give it if it had one. The
// standard decoder refuses those bytes and answers a replacement character for
// each, losing the value the string holds, so they are read here instead.
func decodeStringRune(s string) (rune, int) {
	if len(s) >= 3 && s[0] == 0xED && s[1] >= 0xA0 && s[1] <= 0xBF && s[2] >= 0x80 && s[2] <= 0xBF {
		return 0xD000 | rune(s[1]&0x3F)<<6 | rune(s[2]&0x3F), 3
	}
	return utf8.DecodeRuneInString(s)
}

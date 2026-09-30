package unicornutil

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

// SplitCaseWords reproduces change-case@5.4's `split()`. The exact upstream regex
// pipeline is:
//
//  1. /([\p{Ll}\d])(\p{Lu})/gu       → insert delimiter before an uppercase
//     that follows a lowercase or digit.
//  2. /(\p{Lu})([\p{Lu}][\p{Ll}])/gu → insert delimiter between two
//     uppercases when the second is the
//     start of a TitleCase word
//     (e.g. `XMLHttp` → `XML Http`).
//  3. /[^\p{L}\d]+/giu               → collapse non-alphanumeric runs into
//     the same delimiter.
//
// Then trim leading/trailing delimiters and split. The three replacements only
// establish word boundaries, so the implementation below recognizes the same
// boundaries in one pass and returns slices of the original string. This
// avoids building three intermediate rune buffers for every case candidate.
func SplitCaseWords(s string) []string {
	var words []string
	wordStart := -1
	var previous rune
	for pos, current := range s {
		if !unicode.IsLetter(current) && !isASCIIDigit(current) {
			if wordStart >= 0 {
				words = append(words, s[wordStart:pos])
				wordStart = -1
			}
			continue
		}

		if wordStart < 0 {
			wordStart = pos
		} else {
			boundary := (unicode.IsLower(previous) || isASCIIDigit(previous)) && unicode.IsUpper(current)
			if !boundary && unicode.IsUpper(previous) && unicode.IsUpper(current) {
				_, size := utf8.DecodeRuneInString(s[pos:])
				next, _ := utf8.DecodeRuneInString(s[pos+size:])
				boundary = unicode.IsLower(next)
			}
			if boundary {
				words = append(words, s[wordStart:pos])
				wordStart = pos
			}
		}
		previous = current
	}
	if wordStart >= 0 {
		words = append(words, s[wordStart:])
	}
	return words
}

func isASCIIDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

// pascalLikeTransform is change-case@5.4's `pascalCaseTransformFactory`:
// when a non-first word starts with a digit, prepend `_` to keep the join
// readable and round-trippable; otherwise capitalize the first letter.
func pascalLikeTransform(word string, index int) string {
	if word == "" {
		return ""
	}
	char0, size := utf8.DecodeRuneInString(word)
	first := word[:size]
	rest := ecmascript.StringToLowerCase(word[size:])
	if index > 0 && isASCIIDigit(char0) {
		return "_" + first + rest
	}
	// change-case indexes the first UTF-16 code unit, so an astral initial
	// keeps its original case rather than capitalizing the whole code point.
	if char0 > 0xFFFF {
		return first + rest
	}
	return ecmascript.StringToUpperCase(first) + rest
}

// CamelCase applies change-case's default camelCase conversion.
func CamelCase(s string) string {
	words := SplitCaseWords(s)
	if len(words) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.Grow(len(s) + len(words))
	sb.WriteString(ecmascript.StringToLowerCase(words[0]))
	for i := 1; i < len(words); i++ {
		sb.WriteString(pascalLikeTransform(words[i], i))
	}
	return sb.String()
}

// PascalCase applies change-case's default pascalCase conversion.
func PascalCase(s string) string {
	words := SplitCaseWords(s)
	if len(words) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.Grow(len(s) + len(words))
	for i, w := range words {
		sb.WriteString(pascalLikeTransform(w, i))
	}
	return sb.String()
}

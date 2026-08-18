package ecmascript

import (
	"strings"

	tsstringutil "github.com/microsoft/typescript-go/shim/stringutil"
)

// StringToLowerCase maps a string the way JavaScript's
// String.prototype.toLowerCase does, including contextual and expanding
// Unicode mappings and lone UTF-16 surrogates.
func StringToLowerCase(value string) string {
	return tsstringutil.ToLowerJS(value)
}

// StringCompare compares strings in the order used by JavaScript's abstract
// relational comparison: lexicographically by UTF-16 code units. It also
// recognizes the WTF-8 sentinel tsgo uses to preserve lone surrogates. tsgo's
// CompareStringsCaseSensitive delegates to Go's UTF-8 byte ordering, so the
// UTF-16 walk itself cannot be delegated; rune decoding and surrogate
// conversion still reuse tsgo's string helpers below.
func StringCompare(a, b string) int {
	if !tsstringutil.ContainsNonASCII(a) && !tsstringutil.ContainsNonASCII(b) {
		return strings.Compare(a, b)
	}
	aUnits := stringUTF16Units(a)
	bUnits := stringUTF16Units(b)
	limit := min(len(aUnits), len(bUnits))
	for index := range limit {
		if aUnits[index] < bUnits[index] {
			return -1
		}
		if aUnits[index] > bUnits[index] {
			return 1
		}
	}
	if len(aUnits) < len(bUnits) {
		return -1
	}
	if len(aUnits) > len(bUnits) {
		return 1
	}
	return 0
}

func stringUTF16Units(value string) []uint16 {
	result := make([]uint16, 0, len(value))
	for offset := 0; offset < len(value); {
		char, size := tsstringutil.DecodeJSStringRune(value[offset:])
		offset += size
		if char <= 0xFFFF {
			result = append(result, uint16(char))
			continue
		}
		high, low := tsstringutil.CodePointToSurrogatePair(char)
		result = append(result, uint16(high), uint16(low))
	}
	return result
}

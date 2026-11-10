package unicornutil

import (
	"strconv"
	"strings"

	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

// QuoteString matches Unicorn's quote-js-string spelling, including control
// characters and lone surrogates. The delimiter is a single or double quote.
func QuoteString(value string, quote byte) string {
	var result strings.Builder
	result.WriteByte(quote)
	for len(value) > 0 {
		character, size := ecmascript.DecodeStringRune(value)
		value = value[size:]
		switch character {
		case '\\', rune(quote):
			result.WriteByte('\\')
			result.WriteRune(character)
		case '\n':
			result.WriteString(`\n`)
		case '\r':
			result.WriteString(`\r`)
		case '\t':
			result.WriteString(`\t`)
		case '\b':
			result.WriteString(`\b`)
		case '\f':
			result.WriteString(`\f`)
		case '\v':
			result.WriteString(`\v`)
		default:
			if character < 0x20 || character == 0x7f || character == 0x2028 || character == 0x2029 || (character >= 0xd800 && character <= 0xdfff) {
				result.WriteString(`\u{` + strconv.FormatUint(uint64(character), 16) + `}`)
			} else {
				result.WriteRune(character)
			}
		}
	}
	result.WriteByte(quote)
	return result.String()
}

package ecmascript

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

// JSONString preserves JavaScript's lone UTF-16 surrogates when encoding a
// compiler string value as JSON. Ordinary escaping stays with encoding/json.
type JSONString string

func (value JSONString) MarshalJSON() ([]byte, error) {
	text := string(value)
	if utf8.ValidString(text) {
		return json.Marshal(text)
	}
	encoded := make([]byte, 0, len(text)+2)
	encoded = append(encoded, '"')
	start := 0
	for i := 0; i < len(text); {
		character, size := DecodeStringRune(text[i:])
		if character >= 0xd800 && character <= 0xdfff {
			part, err := json.Marshal(text[start:i])
			if err != nil {
				return nil, err
			}
			encoded = append(encoded, part[1:len(part)-1]...)
			encoded = fmt.Appendf(encoded, `\u%04x`, character)
			start = i + size
		}
		i += size
	}
	part, err := json.Marshal(text[start:])
	if err != nil {
		return nil, err
	}
	encoded = append(encoded, part[1:len(part)-1]...)
	return append(encoded, '"'), nil
}

// JSONValueToString applies JavaScript's String conversion to decoded JSON
// values, including Array#join's treatment of null elements.
func JSONValueToString(value any) string {
	switch value := value.(type) {
	case nil:
		return "null"
	case string:
		return value
	case bool:
		return strconv.FormatBool(value)
	case float64:
		return NumberToString(value)
	case []any:
		parts := make([]string, len(value))
		for i, item := range value {
			if item != nil {
				parts[i] = JSONValueToString(item)
			}
		}
		return strings.Join(parts, ",")
	default:
		return "[object Object]"
	}
}

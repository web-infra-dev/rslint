package ecmascript

import (
	"strconv"
	"strings"
)

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

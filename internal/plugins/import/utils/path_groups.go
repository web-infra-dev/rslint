package utils

import "github.com/web-infra-dev/rslint/internal/utils/minimatch3"

// NewPathGroupMatcher shares import/order and import/extensions' minimatch
// options. An omitted options object disables comments; an explicit {} does not.
func NewPathGroupMatcher(pattern string, raw any) *minimatch3.Matcher {
	options := minimatch3.Options{NoComment: true}
	if values, ok := raw.(map[string]any); ok {
		// cspell:ignore nonegate nocomment noext nobrace nonull
		options.NoNegate = pathGroupOption(values["nonegate"])
		options.NoComment = pathGroupOption(values["nocomment"])
		options.NoCase = pathGroupOption(values["nocase"])
		options.MatchBase = pathGroupOption(values["matchBase"])
		options.NoGlobStar = pathGroupOption(values["noglobstar"])
		options.NoExt = pathGroupOption(values["noext"])
		options.NoBrace = pathGroupOption(values["nobrace"])
		options.Dot = pathGroupOption(values["dot"])
		options.Partial = pathGroupOption(values["partial"])
		options.FlipNegate = pathGroupOption(values["flipNegate"])
		options.NoNull = pathGroupOption(values["nonull"])
	}
	return minimatch3.New(pattern, options)
}

// patternOptions is an unrestricted JSON object upstream; minimatch coerces
// option values using JavaScript truthiness rather than requiring booleans.
func pathGroupOption(value any) bool {
	switch value := value.(type) {
	case nil:
		return false
	case bool:
		return value
	case string:
		return value != ""
	case float64:
		return value != 0
	case int:
		return value != 0
	default:
		return true
	}
}

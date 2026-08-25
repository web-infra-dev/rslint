package test_framework

import (
	"slices"
	"strings"

	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
)

// AssertionFunctionOptions contains the options shared by rules that recognize
// assertion calls and additional test block functions.
type AssertionFunctionOptions struct {
	AssertFunctionNames          []string
	AdditionalTestBlockFunctions []string
}

// ParseAssertionFunctionOptions parses the option shape shared by assertion
// rules. defaultAssertNames is used only when assertFunctionNames is absent.
func ParseAssertionFunctionOptions(options []any, defaultAssertNames []string) AssertionFunctionOptions {
	parsed := AssertionFunctionOptions{
		AssertFunctionNames:          slices.Clone(defaultAssertNames),
		AdditionalTestBlockFunctions: []string{},
	}

	if len(options) == 0 {
		return parsed
	}
	optsMap, _ := options[0].(map[string]interface{})
	if raw, ok := optsMap["assertFunctionNames"].([]interface{}); ok {
		parsed.AssertFunctionNames = stringList(raw)
	}
	if raw, ok := optsMap["additionalTestBlockFunctions"].([]interface{}); ok {
		parsed.AdditionalTestBlockFunctions = stringList(raw)
	}

	return parsed
}

func stringList(raw []interface{}) []string {
	out := make([]string, 0, len(raw))
	for _, value := range raw {
		if text, ok := value.(string); ok {
			out = append(out, text)
		}
	}
	return out
}

// CompileAssertPatterns compiles eslint-plugin-jest's star-pattern syntax. A
// single star matches within a dotted segment, while a double-star segment can
// span dots.
func CompileAssertPatterns(patterns []string) []*esregexp.RegExp {
	compiled := make([]*esregexp.RegExp, 0, len(patterns))
	for _, pattern := range patterns {
		compiled = append(compiled, compileAssertPattern(pattern))
	}
	return compiled
}

func compileAssertPattern(pattern string) *esregexp.RegExp {
	segments := strings.Split(pattern, ".")
	parts := make([]string, 0, len(segments))
	for _, segment := range segments {
		if segment == "**" {
			parts = append(parts, `[a-zA-Z0-9.]*`)
		} else {
			parts = append(parts, strings.ReplaceAll(segment, "*", `[a-zA-Z0-9]*`))
		}
	}

	// Other regular-expression characters are intentionally preserved because
	// upstream passes configured patterns directly to RegExp.
	re, err := esregexp.Compile(`^(?:`+strings.Join(parts, `\.`)+`)(?:\.|$)`, "iu")
	if err != nil {
		return nil
	}
	return re
}

// MatchesAssertName reports whether name matches one of the compiled assertion
// function patterns.
func MatchesAssertName(name string, compiled []*esregexp.RegExp) bool {
	if name == "" {
		return false
	}
	for _, re := range compiled {
		if re != nil && re.TestOrTimeout(name) {
			return true
		}
	}
	return false
}

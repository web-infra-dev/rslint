package ecmascript

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/diagnostics"
	"github.com/microsoft/typescript-go/shim/scanner"
)

// IsValidRegexLiteral reports whether literal is a complete ECMAScript RegExp
// literal, including leading/trailing slashes and flags, under tsgo's latest
// regular expression grammar.
func IsValidRegexLiteral(literal string) bool {
	return IsValidRegexLiteralForVersion(literal, 0)
}

// IsValidRegexLiteralForVersion reports whether literal is valid in the given
// ECMAScript edition. A zero version selects the latest grammar.
func IsValidRegexLiteralForVersion(literal string, ecmaVersion int) bool {
	flags := literalFlags(literal)
	if !regexFeaturesAvailable(flags, literal, ecmaVersion) {
		return false
	}
	s := scanner.NewScanner()
	hasError := false
	s.SetScriptTarget(scriptTargetForECMAVersion(ecmaVersion))
	s.SetOnError(func(message *diagnostics.Message, _ int, _ int, _ ...any) {
		if !isIgnorableRegexLiteralDiagnostic(message, flags) {
			hasError = true
		}
	})
	s.SetText(literal)
	s.ResetTokenState(0)
	if s.Scan() != ast.KindSlashToken {
		return false
	}
	if s.ReScanSlashToken(true) != ast.KindRegularExpressionLiteral {
		return false
	}
	return !hasError && s.TokenFlags()&ast.TokenFlagsUnterminated == 0 && s.TokenEnd() == len(literal)
}

func scriptTargetForECMAVersion(ecmaVersion int) core.ScriptTarget {
	switch {
	case ecmaVersion == 0:
		return core.ScriptTargetLatest
	case ecmaVersion < 2015:
		return core.ScriptTargetES5
	case ecmaVersion > 2025:
		return core.ScriptTargetESNext
	default:
		return core.ScriptTarget(ecmaVersion - 2013)
	}
}

func regexFeaturesAvailable(flags string, literal string, ecmaVersion int) bool {
	if ecmaVersion == 0 {
		return true
	}
	if ecmaVersion < 2015 && strings.ContainsAny(flags, "uy") ||
		ecmaVersion < 2018 && strings.Contains(flags, "s") ||
		ecmaVersion < 2022 && strings.Contains(flags, "d") ||
		ecmaVersion < 2024 && strings.Contains(flags, "v") {
		return false
	}
	return ecmaVersion >= 2025 || !hasRegExpPatternModifiers(literal)
}

func hasRegExpPatternModifiers(literal string) bool {
	if len(literal) < 2 || literal[0] != '/' {
		return false
	}
	lastSlash := strings.LastIndex(literal[1:], "/")
	if lastSlash < 0 {
		return false
	}
	pattern := literal[1 : lastSlash+1]
	inClass := false
	escaped := false
	for i := range len(pattern) {
		switch {
		case escaped:
			escaped = false
		case pattern[i] == '\\':
			escaped = true
		case pattern[i] == '[':
			inClass = true
		case pattern[i] == ']':
			inClass = false
		case !inClass && pattern[i] == '(' && i+2 < len(pattern) && pattern[i+1] == '?':
			j := i + 2
			hasModifier := false
			for j < len(pattern) && strings.ContainsRune("ims", rune(pattern[j])) {
				hasModifier = true
				j++
			}
			if j < len(pattern) && pattern[j] == '-' {
				j++
				for j < len(pattern) && strings.ContainsRune("ims", rune(pattern[j])) {
					hasModifier = true
					j++
				}
			}
			if hasModifier && j < len(pattern) && pattern[j] == ':' {
				return true
			}
		}
	}
	return false
}

// CanonicalizeRegExpLiteral returns the RegExp.prototype.toString spelling of
// a regular expression literal. A literal already carries its source pattern;
// only its flags need to be reordered to JavaScript's canonical order.
func CanonicalizeRegExpLiteral(literal string) string {
	if len(literal) < 2 || literal[0] != '/' {
		return literal
	}
	lastSlash := strings.LastIndex(literal[1:], "/")
	if lastSlash < 0 {
		return literal
	}
	flags := literal[lastSlash+2:]
	var canonical strings.Builder
	canonical.Grow(len(flags))
	for _, flag := range "dgimsuvy" {
		if strings.ContainsRune(flags, flag) {
			canonical.WriteRune(flag)
		}
	}
	if canonical.Len() != len(flags) {
		return literal
	}
	return literal[:lastSlash+2] + canonical.String()
}

func literalFlags(literal string) string {
	if len(literal) < 2 || literal[0] != '/' {
		return ""
	}
	lastSlash := strings.LastIndex(literal[1:], "/")
	if lastSlash < 0 {
		return ""
	}
	return literal[lastSlash+2:]
}

func isIgnorableRegexLiteralDiagnostic(message *diagnostics.Message, flags string) bool {
	if strings.ContainsAny(flags, "uv") {
		return false
	}
	return message == diagnostics.This_backreference_refers_to_a_group_that_does_not_exist_There_are_only_0_capturing_groups_in_this_regular_expression ||
		message == diagnostics.This_backreference_refers_to_a_group_that_does_not_exist_There_are_no_capturing_groups_in_this_regular_expression
}

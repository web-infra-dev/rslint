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
	result := scanRegexLiteral(literal, false, ecmaVersion)
	return result.complete && !result.hasError
}

// RegexLiteralCharacterEventCutoff returns the byte position where a syntax
// error stops regexpp-style character events. Semantic errors resolved after
// the walk do not create a cutoff. Unterminated is true when the lexer could
// not find the literal's closing slash.
func RegexLiteralCharacterEventCutoff(literal string) (position int, unterminated bool, ok bool) {
	result := scanRegexLiteral(literal, true, 0)
	if !result.hasError {
		return 0, false, false
	}
	return result.errorPosition, result.unterminated, true
}

type regexLiteralScanResult struct {
	errorPosition int
	hasError      bool
	unterminated  bool
	complete      bool
}

func scanRegexLiteral(literal string, ignoreDeferredEventErrors bool, ecmaVersion int) regexLiteralScanResult {
	flags := literalFlags(literal)
	s := scanner.NewScanner()
	result := regexLiteralScanResult{}
	ignoredNamedReferenceStart := -1
	s.SetScriptTarget(scriptTargetForECMAVersion(ecmaVersion))
	s.SetOnError(func(message *diagnostics.Message, start int, length int, _ ...any) {
		if ignoreDeferredEventErrors {
			if message == diagnostics.There_is_no_capturing_group_named_0_in_this_regular_expression {
				ignoredNamedReferenceStart = start
				return
			}
			if message == diagnostics.Did_you_mean_0 && start == ignoredNamedReferenceStart {
				return
			}
		}
		if !result.hasError &&
			!isIgnorableRegexLiteralDiagnostic(message, flags) &&
			!isIgnorableSingleVAmpersandDiagnostic(message, literal, flags, start) &&
			(!ignoreDeferredEventErrors || !isIgnorableRegexEventDiagnostic(message, flags)) {
			result.errorPosition = start
			if ignoreDeferredEventErrors && message == diagnostics.Range_out_of_order_in_character_class {
				// regexpp emits onCharacter for both range endpoints before it
				// rejects their ordering. tsgo highlights the whole range, so its
				// end is the equivalent callback cutoff.
				result.errorPosition += length
			}
			result.hasError = true
			result.unterminated = message == diagnostics.Unterminated_regular_expression_literal
		}
	})
	s.SetText(literal)
	s.ResetTokenState(0)
	firstToken := s.Scan()
	if firstToken != ast.KindSlashToken && firstToken != ast.KindSlashEqualsToken {
		result.hasError = true
		return result
	}
	if s.ReScanSlashToken(true) != ast.KindRegularExpressionLiteral {
		result.hasError = true
		return result
	}
	result.complete = s.TokenFlags()&ast.TokenFlagsUnterminated == 0 && s.TokenEnd() == len(literal)
	if !result.complete && !result.hasError {
		result.errorPosition = s.TokenEnd()
		result.hasError = true
		result.unterminated = s.TokenFlags()&ast.TokenFlagsUnterminated != 0
	}
	return result
}

func isIgnorableRegexEventDiagnostic(message *diagnostics.Message, flags string) bool {
	if !strings.ContainsAny(flags, "uv") {
		switch message {
		case diagnostics.Hexadecimal_digit_expected,
			diagnostics.An_extended_Unicode_escape_value_must_be_between_0x0_and_0x10FFFF_inclusive,
			diagnostics.Octal_escape_sequences_are_not_allowed_Use_the_syntax_0,
			diagnostics.Unicode_escape_sequences_are_only_available_when_the_Unicode_u_flag_or_the_Unicode_Sets_v_flag_is_set,
			diagnostics.Expected_a_Unicode_property_name,
			diagnostics.Expected_a_Unicode_property_name_or_value,
			diagnostics.Expected_a_Unicode_property_value,
			diagnostics.Unknown_Unicode_property_name,
			diagnostics.Unknown_Unicode_property_name_or_value,
			diagnostics.Unknown_Unicode_property_value,
			diagnostics.Unicode_property_value_expressions_are_only_available_when_the_Unicode_u_flag_or_the_Unicode_Sets_v_flag_is_set,
			diagnostics.Any_Unicode_property_that_would_possibly_match_more_than_a_single_character_is_only_available_when_the_Unicode_Sets_v_flag_is_set,
			diagnostics.Anything_that_would_possibly_match_more_than_a_single_character_is_invalid_inside_a_negated_character_class,
			diagnostics.Did_you_mean_0:
			// Annex B treats malformed fixed escapes and Unicode property
			// spellings as identity escapes when neither u nor v is present.
			// The bundled tsgo parser still reports its Unicode-mode diagnostics,
			// so they do not mark the point where regexpp stops character events.
			return true
		}
	}
	return false
}

func isIgnorableSingleVAmpersandDiagnostic(message *diagnostics.Message, literal string, flags string, start int) bool {
	if message != diagnostics.Unexpected_0_Did_you_mean_to_escape_it_with_backslash ||
		!strings.ContainsRune(flags, 'v') || start < 0 || start >= len(literal) || literal[start] != '&' {
		return false
	}
	// A single ampersand is a v-mode class character. Only two adjacent,
	// unescaped ampersands form an intersection; `\&&` is an escaped character
	// followed by a single one. The bundled tsgo scanner diagnoses that second
	// form even though regexpp emits both characters.
	previousAmpersand := start > 0 && literal[start-1] == '&' && !regexLiteralCharacterIsEscaped(literal, start-1)
	nextAmpersand := start+1 < len(literal) && literal[start+1] == '&' && !regexLiteralCharacterIsEscaped(literal, start+1)
	return !previousAmpersand && !nextAmpersand
}

func regexLiteralCharacterIsEscaped(literal string, position int) bool {
	backslashes := 0
	for position--; position >= 0 && literal[position] == '\\'; position-- {
		backslashes++
	}
	return backslashes%2 != 0
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

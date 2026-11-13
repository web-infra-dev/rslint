// cspell:ignore subpattern
package regexp

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/diagnostics"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
)

// Validate checks ECMAScript pattern syntax independently of the matching
// backend. In particular, Unicode sets and duplicate names in alternatives do
// not require a matching engine to implement those features.
// flags must be empty, "u", or "v"; constructor flag validation is the caller's job.
func Validate(pattern, flags string) error {
	if flags != "" && flags != "u" && flags != "v" {
		return fmt.Errorf("%w: %s", ErrUnsupportedFlag, flags)
	}
	// ReScanSlashToken takes a literal, whereas a constructor accepts raw slashes
	// and line terminators. Escape those characters without changing the grammar.
	// Close unfinished classes so its literal-boundary scan reaches the parser;
	// their missing terminators are reported after earlier syntax errors.
	var body bytes.Buffer
	escaped, inClass, inName := false, false, false
	captures, named := captureCounts(pattern, flags)
	classDepth := 0
	fallback, earlyReason := "", ""
	earlyPos := len(pattern)*10 + 100
	setEarly := func(reason string) {
		if earlyReason == "" {
			earlyPos = body.Len() + 1
			earlyReason = reason
		}
	}
	skipUntil := 0
	for index, ch := range pattern {
		if index < skipUntil {
			continue
		}
		if escaped {
			if flags == "" && !inName {
				tail := pattern[index+1:]
				identity := (ch == 'p' || ch == 'P') || (ch == 'k' && !named) || (ch == 'u' && !hasHex(tail, 4)) || (ch == 'x' && !hasHex(tail, 2))
				if identity {
					body.Truncate(body.Len() - 1)
				}
				if ch >= '0' && ch <= '9' {
					end := index + 1
					for end < len(pattern) && pattern[end] >= '0' && pattern[end] <= '9' {
						end++
					}
					value, err := strconv.ParseUint(pattern[index:end], 10, 64)
					if inClass || ch == '0' || err != nil || value > uint64(captures) {
						body.Truncate(body.Len() - 1)
						if ch <= '7' {
							n := int(ch - '0')
							end = index + 1
							limit := 2
							if ch <= '3' {
								limit = 3
							}
							for end < len(pattern) && end-index < limit && pattern[end] >= '0' && pattern[end] <= '7' {
								n = n*8 + int(pattern[end]-'0')
								end++
							}
							fmt.Fprintf(&body, `\x%02x`, n)
							skipUntil = end
							escaped = false
							continue
						}
					}
				}
			}
			switch ch {
			case '\n', '\r', '\u2028', '\u2029':
				if flags != "" {
					setEarly("Invalid escape")
				}
				fmt.Fprintf(&body, `u%04x`, ch)
			default:
				body.WriteRune(ch)
			}
			escaped = false
			continue
		}
		switch ch {
		case '\\':
			escaped = true
			body.WriteRune(ch)
		case '<':
			if !inClass && (strings.HasSuffix(pattern[:index], "(?") || strings.HasSuffix(pattern[:index], `\k`)) && index+1 < len(pattern) && pattern[index+1] != '=' && pattern[index+1] != '!' {
				inName = true
			}
			body.WriteRune(ch)
		case '>':
			inName = false
			body.WriteRune(ch)
		case '[':
			classDepth++
			inClass = true
			body.WriteRune(ch)
		case ']':
			if classDepth > 0 {
				classDepth--
			}
			inClass = flags == "v" && classDepth > 0
			body.WriteRune(ch)
		case '/':
			if inClass && flags == "v" {
				setEarly("Invalid character in character class")
			}
			body.WriteString(`\/`)
		case '\n':
			body.WriteString(`\n`)
		case '\r':
			body.WriteString(`\r`)
		default:
			body.WriteRune(ch)
		}
	}
	text := body.String()
	if escaped {
		text = text[:len(text)-1]
		fallback = `\ at end of pattern`
		if flags != "" {
			fallback = "Invalid escape"
		}
		if inClass && (flags == "" || flags == "v") {
			fallback = "Invalid character in character class"
		}
	}
	if inClass {
		text += "]"
		if fallback == "" {
			fallback = "Unterminated character class"
			if flags == "v" && (strings.HasSuffix(pattern, "[") || strings.HasSuffix(pattern, "[^")) {
				fallback = "Invalid character in character class"
			}
		}
	}
	// A leading * would otherwise make Scan read a block comment.
	text = "(?:)|" + text
	if earlyReason != "" {
		earlyPos += 5
	}
	literal := "/" + text + "/" + flags
	s := scanner.NewScanner()
	s.SetScriptTarget(core.ScriptTargetLatest)
	s.SetText(literal)
	reason := earlyReason
	bestPos := earlyPos
	if reason == "" && fallback != "" && (inClass || flags != "") {
		reason = fallback
		bestPos = body.Len() + 6
		if escaped {
			bestPos--
		}
	}
	if reason == "" {
		bestPos = len(literal) + 1
	}
	namedReason := ""
	s.SetOnError(func(msg *diagnostics.Message, start, length int, args ...any) {
		if msg == diagnostics.There_is_no_capturing_group_named_0_in_this_regular_expression {
			namedReason = "Invalid named capture referenced"
			return
		}
		if start >= bestPos {
			return
		}
		reason = syntaxReason(msg, literal, start, length, args, flags)
		bestPos = start
	})
	s.Scan()
	s.ReScanSlashToken(true)
	if reason == "" {
		reason = fallback
	}
	if reason == "" {
		reason = namedReason
	}
	if reason != "" {
		return syntaxError(fmt.Sprintf("Invalid regular expression: /%s/%s: %s", pattern, flags, reason))
	}
	return nil
}

func syntaxReason(msg *diagnostics.Message, text string, start, length int, args []any, flags string) string {
	switch msg {
	case diagnostics.Duplicate_regular_expression_flag:
		return fmt.Sprintf("Duplicated flag '%s'", text[start:start+length])
	case diagnostics.Unknown_regular_expression_flag, diagnostics.This_regular_expression_flag_cannot_be_toggled_within_a_subpattern:
		return "Invalid group"
	case diagnostics.Subpattern_flags_must_be_present_when_there_is_a_minus_sign:
		return "Invalid empty flags"
	case diagnostics.Named_capturing_groups_with_the_same_name_must_be_mutually_exclusive_to_each_other:
		return "Duplicate capture group name"
	case diagnostics.Range_out_of_order_in_character_class:
		return "Range out of order in character class"
	case diagnostics.There_is_nothing_available_for_repetition:
		return "Nothing to repeat"
	case diagnostics.Unexpected_0_Did_you_mean_to_escape_it_with_backslash:
		if len(args) > 0 {
			if args[0] == ")" {
				return "Unmatched ')'"
			}
			if args[0] == "{" {
				if start > 0 && !strings.ContainsRune("|(^$", rune(text[start-1])) {
					return "Incomplete quantifier"
				}
				return "Lone quantifier brackets"
			}
			if args[0] == "}" || args[0] == "]" {
				return "Lone quantifier brackets"
			}
		}
		if flags == "v" {
			return "Invalid character in character class"
		}
		return "Lone quantifier brackets"
	case diagnostics.X_0_expected:
		if len(args) > 0 {
			switch args[0] {
			case "]":
				return "Unterminated character class"
			case ")":
				return "Unterminated group"
			case ":":
				return "Invalid group"
			case ">":
				return "Invalid capture group name"
			case "}":
				return "Incomplete quantifier"
			}
		}
	case diagnostics.Numbers_out_of_order_in_quantifier:
		return "numbers out of order in {} quantifier"
	case diagnostics.Incomplete_quantifier_Digit_expected:
		return "Incomplete quantifier"
	case diagnostics.Expected_a_capturing_group_name:
		return "Invalid capture group name"
	case diagnostics.A_character_class_range_must_not_be_bounded_by_another_character_class:
		if flags == "v" {
			return "Invalid character in character class"
		}
		return "Invalid character class"
	case diagnostics.There_is_no_capturing_group_named_0_in_this_regular_expression:
		return "Invalid named capture referenced"
	case diagnostics.This_backreference_refers_to_a_group_that_does_not_exist_There_are_only_0_capturing_groups_in_this_regular_expression,
		diagnostics.This_backreference_refers_to_a_group_that_does_not_exist_There_are_no_capturing_groups_in_this_regular_expression:
		return "Invalid escape"
	case diagnostics.Expected_a_Unicode_property_name, diagnostics.Unknown_Unicode_property_name,
		diagnostics.Expected_a_Unicode_property_value, diagnostics.Unknown_Unicode_property_value,
		diagnostics.Expected_a_Unicode_property_name_or_value, diagnostics.Unknown_Unicode_property_name_or_value:
		return "Invalid property name"
	case diagnostics.Expected_a_class_set_operand:
		return "Invalid character in character class"
	case diagnostics.A_character_class_must_not_contain_a_reserved_double_punctuator_Did_you_mean_to_escape_it_with_backslash:
		return "Invalid character in character class"
	case diagnostics.Operators_must_not_be_mixed_within_a_character_class_Wrap_it_in_a_nested_class_instead:
		return "Invalid character in character class"
	case diagnostics.Anything_that_would_possibly_match_more_than_a_single_character_is_invalid_inside_a_negated_character_class:
		if start > 2 && text[start-3:start] == `\P{` {
			return "Invalid property name"
		}
		return "Negated character class may contain strings"
	case diagnostics.X_k_must_be_followed_by_a_capturing_group_name_enclosed_in_angle_brackets:
		return "Invalid named reference"
	case diagnostics.Any_Unicode_property_that_would_possibly_match_more_than_a_single_character_is_only_available_when_the_Unicode_Sets_v_flag_is_set,
		diagnostics.X_0_must_be_followed_by_a_Unicode_property_value_expression_enclosed_in_braces:
		return "Invalid property name"
	case diagnostics.Hexadecimal_digit_expected, diagnostics.An_extended_Unicode_escape_value_must_be_between_0x0_and_0x10FFFF_inclusive:
		if strings.LastIndex(text[:start], `\u`) > strings.LastIndex(text[:start], `\x`) {
			return "Invalid unicode escape"
		}
		return "Invalid escape"
	case diagnostics.Unterminated_regular_expression_literal:
		return "Unterminated character class"
	}
	return "Invalid escape"
}

func hasHex(text string, count int) bool {
	if len(text) < count {
		return false
	}
	for _, ch := range text[:count] {
		switch {
		case ch >= '0' && ch <= '9', ch >= 'a' && ch <= 'f', ch >= 'A' && ch <= 'F':
		default:
			return false
		}
	}
	return true
}

func captureCounts(pattern, flags string) (count int, named bool) {
	escaped, inClass := false, false
	depth := 0
	for i, ch := range pattern {
		if escaped {
			escaped = false
			continue
		}
		if ch == '\\' {
			escaped = true
			continue
		}
		if ch == '[' {
			depth++
			inClass = true
		}
		if ch == ']' {
			if depth > 0 {
				depth--
			}
			inClass = flags == "v" && depth > 0
		}
		if ch == '(' && !inClass {
			tail := pattern[i+1:]
			if !strings.HasPrefix(tail, "?") {
				count++
			} else if strings.HasPrefix(tail, "?<") && len(tail) > 2 && tail[2] != '=' && tail[2] != '!' {
				count++
				named = true
			}
		}
	}
	return count, named
}

// Preserve regexpp's diagnostic capitalization at the formatting boundary.
type syntaxError string

func (e syntaxError) Error() string { return string(e) }

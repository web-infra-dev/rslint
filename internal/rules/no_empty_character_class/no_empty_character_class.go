package no_empty_character_class

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/no-empty-character-class
var NoEmptyCharacterClassRule = rule.Rule{
	Name: "no-empty-character-class",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindRegularExpressionLiteral: func(node *ast.Node) {
				text := node.Text()
				pattern, flags := utils.ExtractRegexPatternAndFlags(text)
				unicodeSets := strings.ContainsRune(flags, 'v')
				if hasEmptyCharacterClass(pattern, unicodeSets) {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "unexpected",
						Description: "Empty class.",
					})
				}
			},
		}
	},
}

// hasEmptyCharacterClass scans the regex pattern for empty character classes [].
// When unicodeSets is true (v flag), nested character classes are supported,
// so [ inside a character class opens a nested class that can itself be empty.
func hasEmptyCharacterClass(pattern string, unicodeSets bool) bool {
	if unicodeSets {
		return hasEmptyCharacterClassV(pattern)
	}
	return hasEmptyCharacterClassLegacy(pattern)
}

// hasEmptyCharacterClassLegacy handles non-v-flag regexes with a flat state machine.
// Without the v flag, [ inside a character class is a literal — no nesting.
//
// States:
//
//	outside          — not inside any character class
//	insideStart      — just saw [, next char decides if class is empty/negated
//	insideAfterCaret — saw [^, next ] means [^] (allowed), otherwise non-empty
//	inside           — inside a non-empty class body, waiting for ]
func hasEmptyCharacterClassLegacy(pattern string) bool {
	const (
		outside          = 0
		insideStart      = 1
		insideAfterCaret = 2
		inside           = 3
	)

	state := outside
	i := 0
	for i < len(pattern) {
		ch := pattern[i]
		switch state {
		case outside:
			if ch == '\\' {
				i += 2
				continue
			}
			if ch == '[' {
				state = insideStart
			}
		case insideStart:
			if ch == ']' {
				return true
			}
			if ch == '^' {
				state = insideAfterCaret
			} else {
				state = inside
			}
		case insideAfterCaret:
			if ch == ']' {
				state = outside
			} else {
				state = inside
			}
		case inside:
			if ch == '\\' {
				i += 2
				continue
			}
			if ch == ']' {
				state = outside
			}
		}
		i++
	}
	return false
}

// hasEmptyCharacterClassV handles v-flag (ES2024 unicodeSets) regexes where
// character classes can nest arbitrarily (e.g. [[a]--[b&&[c]]]).
// Uses recursive descent: each [ opens a new level, each ] closes one.
func hasEmptyCharacterClassV(pattern string) bool {
	_, found := scanOutsideV(pattern, 0)
	return found
}

// scanOutsideV scans from position i outside any character class.
// Returns the final position and whether an empty class was found.
func scanOutsideV(pattern string, i int) (int, bool) {
	for i < len(pattern) {
		ch := pattern[i]
		if ch == '\\' {
			i += 2
			continue
		}
		if ch == '[' {
			var found bool
			i, found = scanClassContentsV(pattern, i+1)
			if found {
				return i, true
			}
			continue
		}
		i++
	}
	return i, false
}

// scanClassContentsV scans the contents of a character class starting right after '['.
// It handles the opening '^' if present, detects empty classes ([] or [^]),
// and recurses into nested classes.
func scanClassContentsV(pattern string, i int) (int, bool) {
	if i >= len(pattern) {
		return i, false
	}

	negate := false
	if pattern[i] == '^' {
		negate = true
		i++
	}

	if i < len(pattern) && pattern[i] == ']' {
		if negate {
			// [^] is allowed — matches any character
			return i + 1, false
		}
		// [] is empty
		return i + 1, true
	}

	// Scan class body until ']'
	for i < len(pattern) {
		ch := pattern[i]
		if ch == '\\' {
			i += 2
			continue
		}
		if ch == ']' {
			return i + 1, false
		}
		if ch == '[' {
			// Nested character class
			var found bool
			i, found = scanClassContentsV(pattern, i+1)
			if found {
				// Bubble up: found an empty nested class. Still need to consume
				// the rest of the outer class to return proper position.
				return consumeRestOfClassV(pattern, i), true
			}
			continue
		}
		i++
	}
	return i, false
}

// consumeRestOfClassV consumes the rest of a character class body after an empty
// nested class was found. We still need to find the matching ']' to return a
// correct position, but we know the result is true.
func consumeRestOfClassV(pattern string, i int) int {
	depth := 1
	for i < len(pattern) && depth > 0 {
		ch := pattern[i]
		switch ch {
		case '\\':
			i += 2
			continue
		case '[':
			depth++
		case ']':
			depth--
		}
		i++
	}
	return i
}

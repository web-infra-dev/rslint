package no_empty_character_class

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// https://eslint.org/docs/latest/rules/no-empty-character-class
var NoEmptyCharacterClassRule = rule.Rule{
	Name: "no-empty-character-class",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindRegularExpressionLiteral: func(node *ast.Node) {
				text := node.Text()
				pattern := extractPattern(text)
				if hasEmptyCharacterClass(pattern) {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "unexpected",
						Description: "Empty class.",
					})
				}
			},
		}
	},
}

// extractPattern extracts the regex pattern from a RegularExpressionLiteral text.
// The text is in the form /pattern/flags, so we strip the leading / and trailing /flags.
func extractPattern(text string) string {
	if len(text) < 2 || text[0] != '/' {
		return ""
	}
	// Find the last '/' which separates the pattern from flags
	lastSlash := strings.LastIndex(text[1:], "/")
	if lastSlash == -1 {
		return text[1:]
	}
	return text[1 : lastSlash+1]
}

// hasEmptyCharacterClass scans the regex pattern for empty character classes [].
// It uses a simple state machine:
//   - outside: normal scanning
//   - insideStart: just entered a character class with [
//   - insideAfterCaret: inside a negated class [^, next char decides
//   - inside: inside a character class, waiting for ]
func hasEmptyCharacterClass(pattern string) bool {
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
				// Skip escaped character
				i += 2
				continue
			}
			if ch == '[' {
				state = insideStart
			}
		case insideStart:
			if ch == ']' {
				// Found empty character class []
				return true
			}
			if ch == '^' {
				state = insideAfterCaret
			} else {
				state = inside
			}
		case insideAfterCaret:
			if ch == ']' {
				// This is [^], which is allowed (matches any character)
				state = outside
			} else {
				state = inside
			}
		case inside:
			if ch == '\\' {
				// Skip escaped character inside class
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

package no_octal_escape

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// https://eslint.org/docs/latest/rules/no-octal-escape
var NoOctalEscapeRule = rule.Rule{
	Name: "no-octal-escape",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindStringLiteral: func(node *ast.Node) {
				raw := scanner.GetSourceTextOfNodeFromSourceFile(ctx.SourceFile, node, false)
				if seq := findOctalEscape(raw); seq != "" {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "octalEscapeSequence",
						Description: fmt.Sprintf("Don't use octal: '\\%s'. Use '\\u....' instead.", seq),
					})
				}
			},
		}
	},
}

// findOctalEscape scans the raw source text of a string literal for the first
// octal escape sequence and returns the digit sequence (e.g. "01", "377", "0").
// Returns "" if no octal escape is found.
//
// Key behaviors matching ESLint:
//   - \0 alone (not followed by a digit) is a valid NULL character — skipped
//   - \0 followed by 8 or 9 — flagged as "0" (e.g. \08 → "0")
//   - \0 followed by octal digit — flagged (e.g. \01 → "01")
//   - \1 through \7 — always flagged
//   - \\ (escaped backslash) — skipped as a pair
//   - Only the first octal escape per string is reported
func findOctalEscape(raw string) string {
	n := len(raw)
	i := 0
	for i < n {
		if raw[i] != '\\' {
			i++
			continue
		}
		if i+1 >= n {
			break
		}
		next := raw[i+1]

		switch {
		case next == '\\':
			i += 2 // escaped backslash, skip pair
		case next >= '1' && next <= '7':
			return extractOctalSequence(raw, i+1)
		case next == '0' && i+2 < n && isOctalDigit(raw[i+2]):
			return extractOctalSequence(raw, i+1)
		case next == '0' && i+2 < n && (raw[i+2] == '8' || raw[i+2] == '9'):
			return "0" // \08 or \09
		default:
			i += 2 // other escape (\n, \t, \x, \u, \0 alone, etc.)
		}
	}
	return ""
}

// extractOctalSequence returns the maximal octal digit sequence starting at
// raw[start]. The maximum length depends on the first digit:
//   - 0-3: up to 3 digits total (e.g. \377)
//   - 4-7: up to 2 digits total (e.g. \77)
func extractOctalSequence(raw string, start int) string {
	n := len(raw)
	first := raw[start]
	end := start + 1

	if first <= '3' {
		// 0-3: up to 2 more octal digits
		if end < n && isOctalDigit(raw[end]) {
			end++
			if end < n && isOctalDigit(raw[end]) {
				end++
			}
		}
	} else {
		// 4-7: up to 1 more octal digit
		if end < n && isOctalDigit(raw[end]) {
			end++
		}
	}
	return raw[start:end]
}

func isOctalDigit(ch byte) bool {
	return ch >= '0' && ch <= '7'
}

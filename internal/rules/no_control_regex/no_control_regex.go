package no_control_regex

import (
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/no-control-regex
var NoControlRegexRule = rule.Rule{
	Name: "no-control-regex",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		report := func(node *ast.Node, controlChars []string) {
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "unexpected",
				Description: fmt.Sprintf("Unexpected control character(s) in regular expression: %s.", strings.Join(controlChars, ", ")),
			})
		}

		return rule.RuleListeners{
			ast.KindRegularExpressionLiteral: func(node *ast.Node) {
				pattern, flags := utils.ExtractRegexPatternAndFlags(node.Text())
				if chars := collectControlChars(pattern, flags); len(chars) > 0 {
					report(node, chars)
				}
			},
			ast.KindCallExpression: func(node *ast.Node) {
				callExpr := node.AsCallExpression()
				checkRegExpConstructor(callExpr.Expression, callExpr.Arguments, report)
			},
			ast.KindNewExpression: func(node *ast.Node) {
				newExpr := node.AsNewExpression()
				checkRegExpConstructor(newExpr.Expression, newExpr.Arguments, report)
			},
		}
	},
}

func checkRegExpConstructor(
	callee *ast.Node,
	args *ast.NodeList,
	report func(*ast.Node, []string),
) {
	callee = ast.SkipParentheses(callee)
	if callee == nil || callee.Kind != ast.KindIdentifier || callee.AsIdentifier().Text != "RegExp" {
		return
	}
	if args == nil || len(args.Nodes) == 0 {
		return
	}

	patternNode := args.Nodes[0]
	if patternNode.Kind != ast.KindStringLiteral {
		return
	}
	pattern := patternNode.AsStringLiteral().Text

	// ESLint only treats flags as known when the second argument is a string
	// literal; otherwise flags default to "" (neither u nor v).
	flags := ""
	if len(args.Nodes) >= 2 && args.Nodes[1].Kind == ast.KindStringLiteral {
		flags = args.Nodes[1].AsStringLiteral().Text
	}

	if chars := collectControlChars(pattern, flags); len(chars) > 0 {
		report(patternNode, chars)
	}
}

// collectControlChars scans a regex pattern and returns each code point in
// U+0000..U+001F that appears as:
//   - a raw character,
//   - a \xHH escape,
//   - a \uHHHH escape, or
//   - a \u{H...} escape (only under the u or v flag).
//
// Each hit is formatted as `\xHH` (lowercase, 2 digits). Symbolic control
// escapes (\t, \n, \r, \v, \f, \0, \cX) are NOT reported — matching ESLint.
//
// Note on syntax-invalid patterns: ESLint uses @eslint-community/regexpp's
// validatePattern inside a try/catch, so on a syntax error it keeps the
// characters collected before the error point and discards anything after.
// This scanner does not reproduce that behavior — on a malformed pattern it
// continues scanning to the end, which may over-report control characters
// that appear after the syntax error. Syntactically-invalid patterns are
// independently flagged by the `no-invalid-regexp` rule, so in practice a
// user running both rules sees every issue; the rule attribution differs.
func collectControlChars(pattern, flagsStr string) []string {
	flags := utils.ParseRegexFlags(flagsStr)
	uvMode := flags.UV()

	var results []string
	record := func(cp uint32) {
		results = append(results, fmt.Sprintf(`\x%02x`, cp))
	}

	i := 0
	for i < len(pattern) {
		c := pattern[i]

		if c == '\\' && i+1 < len(pattern) {
			switch pattern[i+1] {
			case 'x':
				// \xHH — 2 hex digits required.
				if i+3 < len(pattern) && utils.IsHexDigit(pattern[i+2]) && utils.IsHexDigit(pattern[i+3]) {
					cp := utils.ParseHexUint(pattern[i+2 : i+4])
					if cp <= 0x1f {
						record(cp)
					}
					i += 4
					continue
				}
			case 'u':
				// \u{H...} — only recognized under u / v flag.
				if uvMode && i+2 < len(pattern) && pattern[i+2] == '{' {
					if closeRel := strings.IndexByte(pattern[i+3:], '}'); closeRel > 0 {
						hex := pattern[i+3 : i+3+closeRel]
						if utils.AllHexDigits(hex) {
							cp := utils.ParseHexUint(hex)
							if cp <= 0x1f {
								record(cp)
							}
							i += 3 + closeRel + 1
							continue
						}
					}
				}
				// \uHHHH — 4 hex digits.
				if i+5 < len(pattern) && utils.AllHexDigits(pattern[i+2:i+6]) {
					cp := utils.ParseHexUint(pattern[i+2 : i+6])
					if cp <= 0x1f {
						record(cp)
					}
					i += 6
					continue
				}
			}
			// Any other escape (\t, \n, \\, \cI, \0, \d, \p{…}, \q{…}, etc.):
			// consume 2 bytes. Multi-byte escape bodies (e.g. `\q{…}` under v
			// flag) are NOT skipped as a unit — their contents may legitimately
			// contain nested escapes (`\u{1F}` etc.) that should be detected by
			// the subsequent loop iterations.
			i += 2
			continue
		}

		if c <= 0x1f {
			record(uint32(c))
		}
		i++
	}

	return results
}

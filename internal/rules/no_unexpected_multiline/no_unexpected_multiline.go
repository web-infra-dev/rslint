package no_unexpected_multiline

import (
	"unicode/utf8"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
)

var messageFunction = rule.RuleMessage{
	Id:          "function",
	Description: "Unexpected newline between function and ( of function call.",
}
var messageProperty = rule.RuleMessage{
	Id:          "property",
	Description: "Unexpected newline between object and [ of property access.",
}
var messageTaggedTemplate = rule.RuleMessage{
	Id:          "taggedTemplate",
	Description: "Unexpected newline between template tag and template literal.",
}
var messageDivision = rule.RuleMessage{
	Id:          "division",
	Description: "Unexpected newline between numerator and division operator.",
}

// https://eslint.org/docs/latest/rules/no-unexpected-multiline
var NoUnexpectedMultilineRule = rule.Rule{
	Name: "no-unexpected-multiline",
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		sf := ctx.SourceFile
		text := sf.Text()

		// reportTokenBreakAfter checks whether the next non-trivia token after
		// `expr` starts on a different line than the source position of
		// `expr.End()`. If so, it reports a diagnostic spanning that next
		// token. Mirrors ESLint's checkForBreakAfter — tsgo's `Expression.End()`
		// already accounts for trailing parens (parens are explicit nodes), so
		// we don't need ESLint's `isNotClosingParenToken` filter.
		reportTokenBreakAfter := func(expr *ast.Node, msg rule.RuleMessage) {
			exprEnd := expr.End()
			tokenStart := scanner.SkipTrivia(text, exprEnd)
			if tokenStart >= len(text) {
				return
			}
			if !hasLineTerminator(text, exprEnd, tokenStart) {
				return
			}
			// ESLint reports a single-character range for `[` / `(`. Don't use
			// scanner.GetRangeOfTokenAtPosition here — for the tagged-template
			// case it would expand to the entire `\`...\`` token, breaking
			// parity with ESLint's `column .. column+1` location.
			ctx.ReportRange(core.NewTextRange(tokenStart, tokenStart+1), msg)
		}

		return rule.RuleListeners{
			ast.KindElementAccessExpression: func(node *ast.Node) {
				// `node.optional` (ESTree) maps to IsOptionalChainRoot — only
				// the link with the literal `?.` token, not later
				// continuations. Without this guard, `b?.[c]` would be flagged.
				if ast.IsOptionalChainRoot(node) {
					return
				}
				expr := node.AsElementAccessExpression().Expression
				reportTokenBreakAfter(expr, messageProperty)
			},

			ast.KindCallExpression: func(node *ast.Node) {
				if ast.IsOptionalChainRoot(node) {
					return
				}
				call := node.AsCallExpression()
				if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
					return
				}
				reportTokenBreakAfter(call.Expression, messageFunction)
			},

			ast.KindTaggedTemplateExpression: func(node *ast.Node) {
				tte := node.AsTaggedTemplateExpression()
				template := tte.Template
				// Position of the backtick that opens the template literal.
				backtick := scanner.SkipTrivia(text, template.Pos())
				if backtick >= len(text) {
					return
				}
				// `template.Pos() - 1` lands on the last character of the
				// previous sibling (Tag, or the closing `>` of TypeArguments
				// when present). Avoids a separate token rewind.
				prevPos := template.Pos() - 1
				if prevPos < 0 {
					return
				}
				if !hasLineTerminator(text, prevPos+1, backtick) {
					return
				}
				ctx.ReportRange(core.NewTextRange(backtick, backtick+1), messageTaggedTemplate)
			},

			ast.KindBinaryExpression: func(node *ast.Node) {
				bin := node.AsBinaryExpression()
				if bin.OperatorToken == nil || bin.OperatorToken.Kind != ast.KindSlashToken {
					return
				}
				// A break before the first slash is required regardless of the
				// surrounding AST shape or the token after the second slash. Reject
				// ordinary same-line division before doing either of those checks.
				firstSlashRange := core.NewTextRange(bin.OperatorToken.End()-1, bin.OperatorToken.End())
				if !hasLineTerminator(text, bin.Left.End(), firstSlashRange.Pos()) {
					return
				}
				// Mirror ESLint's `BinaryExpression > BinaryExpression.left`
				// selector. ESTree drops parens, so `(a/b)/c` matches; tsgo
				// keeps them as explicit nodes. WalkUpParenthesizedExpressions
				// returns the topmost paren wrapper above `node` (or `node`
				// itself if unwrapped) — that's the entity whose direct
				// parent the selector consults.
				outerChild := ast.WalkUpParenthesizedExpressions(node)
				outer := outerChild.Parent
				if outer == nil || outer.Kind != ast.KindBinaryExpression {
					return
				}
				outerBin := outer.AsBinaryExpression()
				if outerBin.OperatorToken == nil || outerBin.OperatorToken.Kind != ast.KindSlashToken {
					return
				}
				if outerBin.Left != outerChild {
					return
				}

				// secondSlash = outer's `/`. Find the token immediately after.
				// ESLint requires `secondSlash.range[1] === tokenAfter.range[0]`
				// — no trivia between the slash and the flag identifier — so
				// we don't call SkipTrivia here.
				afterPos := outerBin.OperatorToken.End()
				if afterPos >= len(text) {
					return
				}
				if !matchesRegexFlagIdentifier(text, afterPos) {
					return
				}

				ctx.ReportRange(firstSlashRange, messageDivision)
			},
		}
	},
}

// hasLineTerminator reports whether text[start:end] contains an ECMAScript
// line terminator. Checking the short gap between two tokens directly avoids
// resolving both positions through the source file's full line map.
func hasLineTerminator(text string, start int, end int) bool {
	for i := start; i < end; i++ {
		switch text[i] {
		case '\r', '\n':
			return true
		case 0xe2:
			// U+2028 LINE SEPARATOR and U+2029 PARAGRAPH SEPARATOR are the
			// two non-ASCII ECMAScript line terminators (E2 80 A8/A9).
			if i+2 < end && text[i+1] == 0x80 && (text[i+2] == 0xa8 || text[i+2] == 0xa9) {
				return true
			}
		}
	}
	return false
}

// matchesRegexFlagIdentifier implements ESLint's Identifier-token plus
// /^[gimsuy]+$/ check. Plain identifiers stay on the allocation-free source
// scan; Unicode escapes are decoded in place so their identifier value (for
// example, g\u0069 -> gi) is compared without constructing a token string.
func matchesRegexFlagIdentifier(text string, pos int) bool {
	matched := false
	for pos < len(text) {
		ch := rune(text[pos])
		if isRegexFlag(ch) {
			matched = true
			pos++
			continue
		}

		if ch == '\\' {
			decoded, end, ok := scanIdentifierUnicodeEscape(text, pos)
			if !ok {
				return matched
			}
			if matched {
				if !scanner.IsIdentifierPart(decoded) {
					return true
				}
			} else if !scanner.IsIdentifierStart(decoded) {
				return false
			}
			if !isRegexFlag(decoded) {
				return false
			}
			matched = true
			pos = end
			continue
		}

		decoded, _ := utf8.DecodeRuneInString(text[pos:])
		if decoded != utf8.RuneError && scanner.IsIdentifierPart(decoded) {
			return false
		}
		return matched
	}
	return matched
}

func isRegexFlag(ch rune) bool {
	switch ch {
	case 'g', 'i', 'm', 's', 'u', 'y':
		return true
	default:
		return false
	}
}

// scanIdentifierUnicodeEscape decodes the identifier escape at pos and
// returns the first byte after it. Both ECMAScript forms are accepted:
// \uXXXX and \u{X...}.
func scanIdentifierUnicodeEscape(text string, pos int) (rune, int, bool) {
	const maxUnicodeCodePoint uint32 = 0x10ffff

	if pos+2 > len(text) || text[pos] != '\\' || text[pos+1] != 'u' {
		return 0, pos, false
	}

	pos += 2
	if pos < len(text) && text[pos] == '{' {
		pos++
		digitStart := pos
		value := uint32(0)
		validValue := true
		for pos < len(text) && isHexDigit(text[pos]) {
			if value > (maxUnicodeCodePoint-hexValue(text[pos]))/16 {
				validValue = false
			} else if validValue {
				value = value*16 + hexValue(text[pos])
			}
			pos++
		}
		if pos == digitStart || pos >= len(text) || text[pos] != '}' || !validValue {
			return 0, pos, false
		}
		return rune(value), pos + 1, true
	}

	if pos+4 > len(text) {
		return 0, pos, false
	}
	value := uint32(0)
	for end := pos + 4; pos < end; pos++ {
		if !isHexDigit(text[pos]) {
			return 0, pos, false
		}
		value = value*16 + hexValue(text[pos])
	}
	return rune(value), pos, true
}

func isHexDigit(ch byte) bool {
	return ch >= '0' && ch <= '9' || ch >= 'a' && ch <= 'f' || ch >= 'A' && ch <= 'F'
}

func hexValue(ch byte) uint32 {
	switch {
	case ch >= '0' && ch <= '9':
		return uint32(ch - '0')
	case ch >= 'a' && ch <= 'f':
		return uint32(ch-'a') + 10
	default:
		return uint32(ch-'A') + 10
	}
}

package no_unexpected_multiline

import (
	"regexp"
	"unicode/utf8"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var regexFlagMatcher = regexp.MustCompile(`^[gimsuy]+$`)

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
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		sf := ctx.SourceFile
		text := sf.Text()
		lineMap := sf.ECMALineMap()

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
			exprEndLine := scanner.ComputeLineOfPosition(lineMap, exprEnd)
			tokenLine := scanner.ComputeLineOfPosition(lineMap, tokenStart)
			if exprEndLine == tokenLine {
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
				prevLine := scanner.ComputeLineOfPosition(lineMap, prevPos)
				backtickLine := scanner.ComputeLineOfPosition(lineMap, backtick)
				if prevLine == backtickLine {
					return
				}
				ctx.ReportRange(core.NewTextRange(backtick, backtick+1), messageTaggedTemplate)
			},

			ast.KindBinaryExpression: func(node *ast.Node) {
				bin := node.AsBinaryExpression()
				if bin.OperatorToken == nil || bin.OperatorToken.Kind != ast.KindSlashToken {
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
				identEnd, ok := scanIdentifier(text, afterPos)
				if !ok {
					return
				}
				if !regexFlagMatcher.MatchString(text[afterPos:identEnd]) {
					return
				}

				// checkForBreakAfter(node.left): compare lines of
				// `bin.Left.End()` (= last char of the numerator, possibly the
				// closing `)` of a paren-wrapped left) and the first `/`
				// (= bin.OperatorToken). Use TrimNodeTextRange so we land on
				// the trimmed start of the slash, not its leading trivia.
				firstSlashRange := utils.TrimNodeTextRange(sf, bin.OperatorToken)
				leftEnd := bin.Left.End()
				leftLine := scanner.ComputeLineOfPosition(lineMap, leftEnd)
				slashLine := scanner.ComputeLineOfPosition(lineMap, firstSlashRange.Pos())
				if leftLine == slashLine {
					return
				}
				ctx.ReportRange(firstSlashRange, messageDivision)
			},
		}
	},
}

// scanIdentifier walks the identifier starting at `pos` (if any) and returns
// the position right after its last character. Uses the tsgo scanner's
// Unicode-aware identifier helpers and full UTF-8 rune decoding so we agree
// with the parser on what counts as an identifier — mirroring ESLint's
// `tokenAfterOperator.type === "Identifier"` check by construction. Returns
// ok=false when no identifier starts at `pos`.
func scanIdentifier(text string, pos int) (int, bool) {
	if pos >= len(text) {
		return pos, false
	}
	startCh, size := utf8.DecodeRuneInString(text[pos:])
	if startCh == utf8.RuneError || !scanner.IsIdentifierStart(startCh) {
		return pos, false
	}
	end := pos + size
	for end < len(text) {
		ch, sz := utf8.DecodeRuneInString(text[end:])
		if ch == utf8.RuneError || !scanner.IsIdentifierPart(ch) {
			break
		}
		end += sz
	}
	return end, true
}

// Package jsx_curly_newline ports `@stylistic/jsx-curly-newline` to rslint. It
// enforces consistent linebreaks immediately inside the curly braces of a JSX
// expression container (`{ ... }` in attribute values and children).
//
// For each container the rule looks at two boundaries: between `{` and the
// first token, and between the last token and `}`. Whether a newline is
// required there is decided per option (`singleline` / `multiline`, each
// `consistent` | `require` | `forbid`; the string forms `'consistent'` and
// `'never'` are shorthands), selected by whether the contained expression is
// single- or multi-line.
//
// Two tsgo↔ESTree shape differences are handled here:
//   - Spread children (`{...x}`) are JSXSpreadChild in ESTree and not visited by
//     this rule; tsgo folds them into JsxExpression with a DotDotDotToken, so
//     those are skipped.
//   - ESTree's `node.expression` unwraps parentheses, while tsgo keeps an
//     explicit ParenthesizedExpression. The single-/multi-line decision uses the
//     unwrapped expression, but the brace-adjacent token positions use the raw
//     (possibly parenthesized) expression — matching ESLint's token walk.
package jsx_curly_newline

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const (
	modeConsistent = "consistent"
	modeRequire    = "require"
	modeForbid     = "forbid"
)

var messages = map[string]string{
	"expectedBefore":   "Expected newline before '}'.",
	"expectedAfter":    "Expected newline after '{'.",
	"unexpectedBefore": "Unexpected newline before '}'.",
	"unexpectedAfter":  "Unexpected newline after '{'.",
}

// parseOptions resolves the (singleline, multiline) modes from every shape the
// loader can deliver for upstream's `string | { singleline?, multiline? }`
// schema: nil, a bare value (single-option CLI form), or a single-element array
// (rule-tester form). It mirrors upstream's destructuring exactly —
//
//	'never'      => { singleline: 'forbid',  multiline: 'forbid' }
//	'consistent' => { singleline: 'consistent', multiline: 'consistent' }
//	object       => fields with each side defaulting to 'consistent'
//
// Any value outside the enum is rejected by upstream's JSON schema before the
// rule runs; the `default` arm of the mode switch treats unknown strings as
// `consistent`, so hand-written rslint configs degrade the same way.
func parseOptions(options any) (singleline, multiline string) {
	singleline, multiline = modeConsistent, modeConsistent
	raw := options
	if arr, ok := options.([]interface{}); ok {
		if len(arr) == 0 {
			return
		}
		raw = arr[0]
	}
	switch v := raw.(type) {
	case string:
		if v == "never" {
			return modeForbid, modeForbid
		}
		// 'consistent' (the only other enum value) sets both sides; upstream:
		// `typeof options === 'string' ? { multiline: options, singleline: options }`.
		return v, v
	case map[string]interface{}:
		if s, ok := v["singleline"].(string); ok {
			singleline = s
		}
		if m, ok := v["multiline"].(string); ok {
			multiline = m
		}
	}
	return
}

// JsxCurlyNewlineRule enforces linebreaks inside JSX expression-container braces.
var JsxCurlyNewlineRule = rule.Rule{
	Name: "@stylistic/jsx-curly-newline",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		singleline, multiline := parseOptions(options)
		text := ctx.SourceFile.Text()
		lineStarts := ctx.SourceFile.ECMALineMap()
		lineOf := func(pos int) int {
			return scanner.ComputeLineOfPosition(lineStarts, pos)
		}

		return rule.RuleListeners{
			ast.KindJsxExpression: func(node *ast.Node) {
				jsxExpr := node.AsJsxExpression()
				// Spread children ({...x}) map to ESTree JSXSpreadChild, which
				// this rule does not handle (it only visits
				// JSXExpressionContainer).
				if jsxExpr.DotDotDotToken != nil {
					return
				}

				openBrace := utils.TrimNodeTextRange(ctx.SourceFile, node).Pos()
				closeBrace := node.End() - 1
				// Guard parser-recovery shapes (unterminated container,
				// synthesized nodes) where the braces aren't where we expect.
				if openBrace < 0 || closeBrace <= openBrace || closeBrace >= len(text) ||
					text[openBrace] != '{' || text[closeBrace] != '}' {
					return
				}

				openLine := lineOf(openBrace)
				closeLine := lineOf(closeBrace)

				// afterStart / beforeEnd are the positions of the first token
				// after `{` and the last token before `}`, mirroring ESLint's
				// getTokenAfter(leftCurly) / getTokenBefore(rightCurly) (both
				// skip comments). For a non-empty container these are the raw
				// expression's trimmed bounds, whose token scan already skips
				// leading/trailing comment trivia. For an empty container ({},
				// {/* c */}) the token after `{` is `}` and the token before
				// `}` is `{`, exactly where ESLint's walk lands.
				var afterStart, beforeEnd int
				var singleLine bool
				if jsxExpr.Expression != nil {
					exprRange := utils.TrimNodeTextRange(ctx.SourceFile, jsxExpr.Expression)
					afterStart = exprRange.Pos()
					beforeEnd = exprRange.End()
					// isSingleLine inspects ESTree's node.expression, which has
					// no node for parentheses; tsgo keeps an explicit
					// ParenthesizedExpression. Judge single-/multi-line by the
					// unwrapped expression so `{(\nfoo\n)}` is multiline by its
					// inner node — matching ESLint.
					inner := utils.TrimNodeTextRange(ctx.SourceFile, ast.SkipParentheses(jsxExpr.Expression))
					singleLine = lineOf(inner.Pos()) == lineOf(inner.End())
				} else {
					afterStart = closeBrace
					beforeEnd = openBrace + 1
					singleLine = openLine == closeLine
				}

				hasLeftNewline := openLine != lineOf(afterStart)
				hasRightNewline := lineOf(beforeEnd) != closeLine

				mode := multiline
				if singleLine {
					mode = singleline
				}
				var needsNewlines bool
				switch mode {
				case modeForbid:
					needsNewlines = false
				case modeRequire:
					needsNewlines = true
				default: // 'consistent' and any non-enum value
					needsNewlines = hasLeftNewline
				}

				// Left brace: the gap between `{` and the first token.
				switch {
				case hasLeftNewline && !needsNewlines:
					reportToken(ctx, openBrace, "unexpectedAfter",
						removalFix(ctx.SourceFile, openBrace+1, afterStart))
				case !hasLeftNewline && needsNewlines:
					reportToken(ctx, openBrace, "expectedAfter",
						&rule.RuleFix{Text: "\n", Range: core.NewTextRange(openBrace+1, openBrace+1)})
				}

				// Right brace: the gap between the last token and `}`.
				switch {
				case hasRightNewline && !needsNewlines:
					reportToken(ctx, closeBrace, "unexpectedBefore",
						removalFix(ctx.SourceFile, beforeEnd, closeBrace))
				case !hasRightNewline && needsNewlines:
					reportToken(ctx, closeBrace, "expectedBefore",
						&rule.RuleFix{Text: "\n", Range: core.NewTextRange(closeBrace, closeBrace)})
				}
			},
		}
	},
}

// reportToken reports a diagnostic anchored at the single-character brace token
// at pos. A nil fix emits the diagnostic without an autofix (upstream's
// `output: null`, used when a comment sits in the gap).
func reportToken(ctx rule.RuleContext, pos int, id string, fix *rule.RuleFix) {
	msg := rule.RuleMessage{Id: id, Description: messages[id]}
	tokenRange := core.NewTextRange(pos, pos+1)
	if fix != nil {
		ctx.ReportRangeWithFixes(tokenRange, msg, *fix)
	} else {
		ctx.ReportRange(tokenRange, msg)
	}
}

// removalFix builds a fix that deletes the [start, end) gap between a brace and
// the adjacent token, but returns nil when a comment sits in that gap — mirroring
// upstream's safeReplaceTextBetween, which refuses to delete comments.
// utils.HasCommentsInRange is rslint's commentsExistBetween equivalent.
func removalFix(sf *ast.SourceFile, start, end int) *rule.RuleFix {
	if utils.HasCommentsInRange(sf, core.NewTextRange(start, end)) {
		return nil
	}
	return &rule.RuleFix{Text: "", Range: core.NewTextRange(start, end)}
}

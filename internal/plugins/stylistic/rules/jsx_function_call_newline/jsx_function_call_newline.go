// Package jsx_function_call_newline ports `@stylistic/jsx-function-call-newline`
// to rslint. It enforces line breaks before and after a JSX element when the
// element is passed as an argument to a function call or `new` expression.
//
// Two option values control how much is enforced:
//
//	multiline (default) — only JSX arguments that themselves span multiple
//	                      lines must sit on their own line
//	always              — every JSX argument must sit on its own line
//
// A JSX argument "sits on its own line" when there is a line break between it
// and the previous token (the call's `(`, a `,`, or an opening paren) and
// between it and the next token — except that a trailing `,` after the
// argument is treated as already separating it, so no closing break is added.
//
// tsgo↔ESTree shape notes handled here:
//   - ESTree strips parentheses, so for `fn((<div/>))` the reported argument is
//     the bare JSXElement and getTokenBefore/After resolve to the inner parens.
//     tsgo keeps an explicit ParenthesizedExpression, so each argument is
//     unwrapped with ast.SkipParentheses and every position (start/end, the
//     previous-token line, the next-token scan) is taken from the inner JSX
//     node — landing on the same tokens ESLint walks.
//   - ESTree's single JSXElement kind (carrying a selfClosing flag) splits in
//     tsgo into JsxElement, JsxSelfClosingElement, and JsxFragment; isJSX here
//     accepts all three, matching upstream's `['JSXElement','JSXFragment']`.
package jsx_function_call_newline

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/stylisticutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const optionAlways = "always"

const missingLineBreak = "missingLineBreak"

var messages = map[string]string{
	missingLineBreak: "Missing line break around JSX",
}

// parseOption resolves the string option from every shape the loader can
// deliver: nil / empty (default 'multiline'), a bare string (single-option CLI
// form), or a single-element array (rule-tester form). Upstream's schema is the
// enum ['always', 'multiline'] defaulting to 'multiline'; only 'always' changes
// behavior, so every non-'always' value (including 'multiline' and any
// out-of-enum string the schema would have rejected) is treated as 'multiline'.
func parseOption(options any) string {
	raw := options
	if arr, ok := options.([]interface{}); ok {
		if len(arr) == 0 {
			return "multiline"
		}
		raw = arr[0]
	}
	if s, ok := raw.(string); ok {
		return s
	}
	return "multiline"
}

// isJSX mirrors upstream's `['JSXElement','JSXFragment'].includes(node.type)`.
// ESTree's single JSXElement (carrying a selfClosing flag) splits in tsgo into
// JsxElement (with children) and JsxSelfClosingElement, so both map to ESTree's
// JSXElement; JsxFragment maps to JSXFragment.
func isJSX(node *ast.Node) bool {
	return ast.IsJsxElement(node) || ast.IsJsxSelfClosingElement(node) || ast.IsJsxFragment(node)
}

// JsxFunctionCallNewlineRule enforces line breaks around JSX arguments of
// function calls and `new` expressions.
var JsxFunctionCallNewlineRule = rule.Rule{
	Name: "@stylistic/jsx-function-call-newline",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		option := parseOption(options)
		sf := ctx.SourceFile
		text := sf.Text()

		check := func(arg *ast.Node) {
			// ESTree's argument node is the parens-stripped expression; tsgo
			// keeps the ParenthesizedExpression, so unwrap to the node ESLint
			// inspects. All positions below come from this inner node, so the
			// previous/next token scans land on the inner parens (if any) just
			// like ESLint's getTokenBefore/After.
			node := ast.SkipParentheses(arg)
			if !isJSX(node) {
				return
			}

			r := utils.TrimNodeTextRange(sf, node)
			start, end := r.Pos(), r.End()
			// Defensive guard against parser-recovery / synthesized JSX nodes
			// (e.g. while editing in the LSP) whose trimmed range can be
			// degenerate; without it the text[start:end] slice below could panic.
			if start < 0 || end > len(text) || start >= end {
				return
			}

			// `option === 'always' || !isSingleLine(node)`: in multiline mode a
			// single-line JSX argument is never touched.
			if option != optionAlways && stylisticutil.SameLineByPos(sf, start, end) {
				return
			}

			// needsOpeningNewLine === isTokenOnSameLine(previousToken, node).
			// node.Pos() is the full-start, which under tsgo's trivia model is
			// exactly the previous token's end position; its line therefore
			// equals previousToken.loc.end.line. Comparing to the JSX start line
			// reproduces ESLint's same-line test, including the comment case
			// (a comment in the gap belongs to the JSX's leading trivia, so
			// node.Pos() still sits on the previous token's line).
			needsOpening := stylisticutil.SameLineByPos(sf, node.Pos(), start)

			// needsClosingNewLine: false when the next token is a `,` (the comma
			// already separates this argument); otherwise true when the JSX end
			// shares a line with the next token's end (i.e. no break before the
			// `)` / inner paren that follows). The scanner starts at `end`, so
			// the next real token always begins at/after it, satisfying
			// upstream's `nextToken.range[0] >= node.range[1]` by construction.
			needsClosing := false
			s := scanner.GetScannerForSourceFile(sf, end)
			if s.Token() != ast.KindEndOfFile && s.Token() != ast.KindCommaToken {
				needsClosing = stylisticutil.SameLineByPos(sf, end, s.TokenEnd())
			}

			if !needsOpening && !needsClosing {
				return
			}

			fixed := text[start:end]
			if needsOpening {
				fixed = "\n" + fixed
			}
			if needsClosing {
				fixed = fixed + "\n"
			}
			ctx.ReportRangeWithFixes(
				core.NewTextRange(start, end),
				rule.RuleMessage{Id: missingLineBreak, Description: messages[missingLineBreak]},
				rule.RuleFixReplaceRange(core.NewTextRange(start, end), fixed),
			)
		}

		handleArgs := func(args *ast.NodeList) {
			if args == nil || len(args.Nodes) == 0 {
				return
			}
			for _, arg := range args.Nodes {
				check(arg)
			}
		}

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				handleArgs(node.AsCallExpression().Arguments)
			},
			ast.KindNewExpression: func(node *ast.Node) {
				handleArgs(node.AsNewExpression().Arguments)
			},
		}
	},
}

package radix

import (
	"strconv"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var missingParametersMsg = rule.RuleMessage{
	Id:          "missingParameters",
	Description: "Missing parameters.",
}

var missingRadixMsg = rule.RuleMessage{
	Id:          "missingRadix",
	Description: "Missing radix parameter.",
}

var invalidRadixMsg = rule.RuleMessage{
	Id:          "invalidRadix",
	Description: "Invalid radix parameter, must be an integer between 2 and 36.",
}

var addRadixParameter10Msg = rule.RuleMessage{
	Id:          "addRadixParameter10",
	Description: "Add radix parameter `10` for parsing decimal numbers.",
}

// isValidRadixValue reports whether a numeric value is an integer in [2, 36].
func isValidRadixValue(f float64) bool {
	if f != float64(int64(f)) {
		return false
	}
	n := int64(f)
	return n >= 2 && n <= 36
}

// isValidRadix mirrors ESLint's isValidRadix: a radix argument is invalid if it
// is a literal whose value is not an integer in [2, 36], or a bare `undefined`
// identifier. Non-literal expressions (variables, calls, etc.) are assumed
// valid because their value cannot be determined statically.
//
// NOTE: NoSubstitutionTemplateLiteral (e.g. `parseInt("10", ` + "`10`" + `)`) is
// deliberately NOT treated as invalid — ESLint's AST models template literals
// as a separate node type from Literal, so its isValidRadix never rejects
// them.
func isValidRadix(radix *ast.Node) bool {
	switch radix.Kind {
	case ast.KindNumericLiteral:
		// tsgo normalizes NumericLiteral.Text to its decimal form at parse
		// time (e.g. `0x10` → "16", `0b10` → "2", `1.6e1` → "16"), so we
		// don't have to handle non-decimal prefixes ourselves.
		// NormalizeNumericLiteral stays as a defensive pass in case that
		// invariant ever changes.
		text := utils.NormalizeNumericLiteral(radix.AsNumericLiteral().Text)
		f, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return false
		}
		return isValidRadixValue(f)

	case ast.KindStringLiteral,
		ast.KindBigIntLiteral,
		ast.KindRegularExpressionLiteral,
		ast.KindNullKeyword,
		ast.KindTrueKeyword,
		ast.KindFalseKeyword:
		// Non-numeric literals and booleans/null can never be integers in [2, 36].
		return false

	case ast.KindIdentifier:
		// ESLint treats bare `undefined` as an invalid radix regardless of
		// whether it is shadowed — mirror that here.
		return radix.AsIdentifier().Text != "undefined"
	}
	return true
}

// isParseIntCallee reports whether a callee is a reference to the global
// `parseInt` or `Number.parseInt` function that is not shadowed. Mirrors
// ESLint's scope-based check, with the same constraints:
//
//   - `Number['parseInt']()` is NOT matched (ESLint's isParseIntMethod rejects
//     computed member access).
//   - `Number.#parseInt()` is NOT matched (PrivateIdentifier, not Identifier).
//   - TS-only wrappers (`!`, `as`, `<T>`, `satisfies`) on the callee or on
//     `Number` cause ESLint's check to fail because the callee node is no
//     longer an Identifier. We mirror that by only peeling parentheses, not
//     other outer expressions.
func isParseIntCallee(callee *ast.Node) bool {
	callee = ast.SkipParentheses(callee)
	if callee == nil {
		return false
	}

	switch callee.Kind {
	case ast.KindIdentifier:
		return callee.AsIdentifier().Text == "parseInt" &&
			!utils.IsShadowed(callee, "parseInt")

	case ast.KindPropertyAccessExpression:
		pae := callee.AsPropertyAccessExpression()
		name := pae.Name()
		if name == nil || !ast.IsIdentifier(name) {
			return false
		}
		if name.AsIdentifier().Text != "parseInt" {
			return false
		}
		obj := ast.SkipParentheses(pae.Expression)
		if obj == nil || !ast.IsIdentifier(obj) {
			return false
		}
		return obj.AsIdentifier().Text == "Number" &&
			!utils.IsShadowed(obj, "Number")
	}
	return false
}

// hasTrailingComma reports whether the CallExpression source text has a
// trailing comma after the last argument. It skips whitespace and comments
// between the last argument and the closing paren.
func hasTrailingComma(sourceText string, lastArgEnd, callEnd int) bool {
	pos := scanner.SkipTrivia(sourceText, lastArgEnd)
	return pos < callEnd && sourceText[pos] == ','
}

// https://eslint.org/docs/latest/rules/radix
var RadixRule = rule.Rule{
	Name: "radix",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				if call == nil {
					return
				}
				if !isParseIntCallee(call.Expression) {
					return
				}
				checkArguments(ctx, node, call)
			},
		}
	},
}

func checkArguments(ctx rule.RuleContext, node *ast.Node, call *ast.CallExpression) {
	var args []*ast.Node
	if call.Arguments != nil {
		args = call.Arguments.Nodes
	}

	switch len(args) {
	case 0:
		ctx.ReportNode(node, missingParametersMsg)

	case 1:
		lastArg := args[0]
		sourceText := ctx.SourceFile.Text()
		// TS AST node End() excludes trailing trivia, and the last token of a
		// CallExpression is always `)`, so End()-1 is the byte offset of `)`.
		// Byte offsets are the rule-layer convention — the LSP layer converts
		// them to UTF-16 code units via scanner.GetECMALineAndUTF16CharacterOfPosition.
		insertPos := node.End() - 1
		text := ", 10"
		if hasTrailingComma(sourceText, lastArg.End(), node.End()) {
			text = " 10,"
		}
		ctx.ReportNodeWithSuggestions(node, missingRadixMsg, rule.RuleSuggestion{
			Message: addRadixParameter10Msg,
			FixesArr: []rule.RuleFix{
				rule.RuleFixReplaceRange(core.NewTextRange(insertPos, insertPos), text),
			},
		})

	default:
		if !isValidRadix(args[1]) {
			ctx.ReportNode(node, invalidRadixMsg)
		}
	}
}

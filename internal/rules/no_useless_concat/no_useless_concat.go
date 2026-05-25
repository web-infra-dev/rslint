package no_useless_concat

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/no-useless-concat
var NoUselessConcatRule = rule.Rule{
	Name: "no-useless-concat",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		sourceFile := ctx.SourceFile
		lineMap := sourceFile.ECMALineMap()

		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				if !utils.IsPlusBinaryExpression(node) {
					return
				}
				bin := node.AsBinaryExpression()

				// For a chain like `a + 'b' + 'c'` the AST is left-associative
				// (`(a + 'b') + 'c'`); we need the leaf on the right edge of
				// the left operand and the leaf on the left edge of the right
				// operand to decide whether this particular `+` joins two
				// literals.
				left := getRightmostLeaf(bin.Left)
				right := getLeftmostLeaf(bin.Right)

				if !utils.IsStringLiteralOrTemplate(left) || !utils.IsStringLiteralOrTemplate(right) {
					return
				}

				// ESLint checks `left.loc.end.line === right.loc.start.line`.
				// `node.End()` already excludes trailing trivia, but `node.Pos()`
				// includes leading trivia, so trim the right leaf to find its
				// true starting line.
				leftEndLine := scanner.ComputeLineOfPosition(lineMap, left.End())
				rightStartLine := scanner.ComputeLineOfPosition(lineMap, utils.TrimNodeTextRange(sourceFile, right).Pos())
				if leftEndLine != rightStartLine {
					return
				}

				ctx.ReportRange(
					utils.TrimNodeTextRange(sourceFile, bin.OperatorToken),
					rule.RuleMessage{
						Id:          "unexpectedConcat",
						Description: "Unexpected string concatenation of literals.",
					},
				)
			},
		}
	},
}

// getRightmostLeaf descends into the right side of any nested `+` chain,
// transparently unwrapping `ParenthesizedExpression`. Mirrors ESLint's
// `getLeft` helper — for `foo + 'a' + 'b'` the leftmost AST neighbor of the
// outer `+` is the concatenation `foo + 'a'`, whose rightmost leaf is `'a'`.
func getRightmostLeaf(node *ast.Node) *ast.Node {
	node = ast.SkipParentheses(node)
	for utils.IsPlusBinaryExpression(node) {
		node = ast.SkipParentheses(node.AsBinaryExpression().Right)
	}
	return node
}

// getLeftmostLeaf is the mirror of getRightmostLeaf for the right side of a
// `+` chain. Mirrors ESLint's `getRight` helper.
func getLeftmostLeaf(node *ast.Node) *ast.Node {
	node = ast.SkipParentheses(node)
	for utils.IsPlusBinaryExpression(node) {
		node = ast.SkipParentheses(node.AsBinaryExpression().Left)
	}
	return node
}

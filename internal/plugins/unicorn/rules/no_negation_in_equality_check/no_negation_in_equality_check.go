// Ported from eslint-plugin-unicorn v75.0.0; see LICENSE.
package no_negation_in_equality_check

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var NoNegationInEqualityCheckRule = rule.Rule{
	Name:   "unicorn/no-negation-in-equality-check",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				binary := node.AsBinaryExpression()
				var replacement string
				switch binary.OperatorToken.Kind {
				case ast.KindEqualsEqualsEqualsToken:
					replacement = "!=="
				case ast.KindExclamationEqualsEqualsToken:
					replacement = "==="
				case ast.KindEqualsEqualsToken:
					replacement = "!="
				case ast.KindExclamationEqualsToken:
					replacement = "=="
				default:
					return
				}
				left := utils.ESTreeRuntimeExpression(binary.Left)
				if !isNegation(left) || isNegation(utils.ESTreeRuntimeExpression(left.AsPrefixUnaryExpression().Operand)) {
					return
				}
				start := utils.TrimNodeTextRange(ctx.SourceFile, left).Pos()
				bang := core.NewTextRange(start, start+1)
				ctx.ReportRangeWithDeferredSuggestions(bang, rule.RuleMessage{
					Id: "no-negation-in-equality-check/error", Description: "Negated expression is not allowed in equality check.",
				}, func() []rule.RuleSuggestion {
					fixes := unicornutil.SpaceAroundKeywordFixes(ctx.SourceFile, node)
					afterBang, _ := utils.TokenAtOrAfter(ctx.SourceFile, bang.End())
					fixes = append(fixes, returnOrThrowParentheses(ctx.SourceFile, node, afterBang.Start)...)
					prefix := ""
					leading := node
					for leading.Parent != nil && !ast.IsExpressionStatement(leading.Parent) && utils.TrimNodeTextRange(ctx.SourceFile, leading.Parent).Pos() == start {
						leading = leading.Parent
					}
					if unicornutil.NeedsSemicolonBefore(ctx.SourceFile, leading, afterBang.Text) {
						prefix = ";"
					}
					fixes = append(fixes, rule.RuleFixReplaceRange(bang, prefix), rule.RuleFixReplace(ctx.SourceFile, binary.OperatorToken, replacement))
					return []rule.RuleSuggestion{{Message: rule.RuleMessage{Id: "no-negation-in-equality-check/suggestion", Description: "Switch to '" + replacement + "' check.", Data: map[string]string{"operator": replacement}}, FixesArr: fixes}}
				})
			},
		}
	},
}

func isNegation(node *ast.Node) bool {
	return node.Kind == ast.KindPrefixUnaryExpression && node.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken
}

func returnOrThrowParentheses(sourceFile *ast.SourceFile, node *ast.Node, operandStart int) []rule.RuleFix {
	parent := node.Parent
	if parent == nil || (parent.Kind != ast.KindReturnStatement && parent.Kind != ast.KindThrowStatement) {
		return nil
	}
	keyword, _ := utils.TokenAtOrAfter(sourceFile, utils.TrimNodeTextRange(sourceFile, parent).Pos())
	keywordLine, _ := scanner.GetECMALineAndUTF16CharacterOfPosition(sourceFile, keyword.Start)
	operandLine, _ := scanner.GetECMALineAndUTF16CharacterOfPosition(sourceFile, operandStart)
	if keywordLine == operandLine {
		return nil
	}
	last, _ := utils.TokenBeforePosition(sourceFile, parent.End())
	end := parent.End()
	if last.Kind == ast.KindSemicolonToken {
		end = last.Start
	}
	return []rule.RuleFix{
		rule.RuleFixReplaceRange(core.NewTextRange(keyword.End, keyword.End), " ("),
		rule.RuleFixReplaceRange(core.NewTextRange(end, end), ")"),
	}
}

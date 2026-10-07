// Ported from eslint-plugin-unicorn v77.0.0; see LICENSE.
package no_unnecessary_await

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

var NoUnnecessaryAwaitRule = rule.Rule{
	Name:   "unicorn/no-unnecessary-await",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		// The parser rejects bare top-level `await !value` in JS before this
		// listener runs; see the documented limitation and skipped upstream case.
		return rule.RuleListeners{
			ast.KindAwaitExpression: func(node *ast.Node) {
				value := utils.ESTreeRuntimeExpression(node.Expression())
				if !notPromise(utils.SkipAssertionsAndParens(value)) {
					return
				}
				start := scanner.GetTokenPosOfNode(node, ctx.SourceFile, false)
				awaitRange := core.NewTextRange(start, start+len("await"))
				message := rule.RuleMessage{
					Id:          "no-unnecessary-await",
					Description: "Do not `await` non-promise value.",
				}
				ctx.ReportRangeWithDeferredFixes(awaitRange, message, func() []rule.RuleFix {
					// Removing await before a function/class can turn it into a
					// declaration. Keep upstream's conservative exclusion.
					if value.Kind == ast.KindFunctionExpression || value.Kind == ast.KindClassExpression || !isLastEvaluated(node) {
						return nil
					}
					return removeAwait(ctx.SourceFile, node, value, awaitRange)
				})
			},
		}
	},
}

func notPromise(node *ast.Node) bool {
	node = utils.ESTreeRuntimeExpression(node)
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindArrayLiteralExpression, ast.KindArrowFunction, ast.KindAwaitExpression,
		ast.KindClassExpression, ast.KindFunctionExpression,
		ast.KindJsxElement, ast.KindJsxSelfClosingElement, ast.KindJsxFragment,
		ast.KindNoSubstitutionTemplateLiteral, ast.KindTemplateExpression,
		ast.KindPrefixUnaryExpression, ast.KindPostfixUnaryExpression,
		ast.KindTypeOfExpression, ast.KindVoidExpression, ast.KindDeleteExpression:
		return true
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		operator := binary.OperatorToken.Kind
		if operator == ast.KindCommaToken {
			return notPromise(binary.Right)
		}
		return !ast.IsAssignmentOperator(operator) && !ast.IsLogicalOrCoalescingBinaryOperator(operator)
	default:
		return utils.IsESTreeLiteralKind(node.Kind)
	}
}

// Even a non-promise await suspends execution for a microtask. Match upstream's
// syntactic tail check so a fix cannot advance subsequent work in the body.
func isLastEvaluated(node *ast.Node) bool {
	parent := utils.ESTreeParent(node)
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindArrowFunction:
		return utils.ESTreeRuntimeExpression(parent.Body()) == node
	case ast.KindSourceFile:
		statements := parent.AsSourceFile().Statements.Nodes
		return statements[len(statements)-1] == node
	case ast.KindBlock:
		statements := parent.AsBlock().Statements.Nodes
		return statements[len(statements)-1] == node &&
			(ast.IsFunctionBlock(parent) || isLastEvaluated(parent))
	case ast.KindIfStatement:
		return utils.ESTreeRuntimeExpression(parent.Expression()) != node && isLastEvaluated(parent)
	case ast.KindVariableDeclarationList:
		declarations := parent.AsVariableDeclarationList().Declarations.Nodes
		return declarations[len(declarations)-1] == node && isLastEvaluated(parent)
	case ast.KindVariableStatement:
		// ESTree wraps exported declarations in ExportNamedDeclaration.
		return !ast.HasSyntacticModifier(parent, ast.ModifierFlagsExport) && isLastEvaluated(parent)
	case ast.KindExpressionStatement, ast.KindReturnStatement, ast.KindThrowStatement, ast.KindVariableDeclaration:
		return isLastEvaluated(parent)
	default:
		return false
	}
}

func removeAwait(sourceFile *ast.SourceFile, node, value *ast.Node, awaitRange core.TextRange) []rule.RuleFix {
	text := sourceFile.Text()
	end := ecmascript.SkipLeadingWhitespace(text, awaitRange.End(), len(text))
	fixes := []rule.RuleFix{rule.RuleFixRemoveRange(awaitRange.WithEnd(end))}

	valueStart := scanner.GetTokenPosOfNode(value, sourceFile, false)
	// A direct return/throw parent excludes existing parentheses, including
	// those whose expression is wrapped in a synthesized JSDoc assertion.
	parent := node.Parent
	if !utils.IsSameLine(sourceFile, awaitRange.Pos(), valueStart) &&
		(parent.Kind == ast.KindReturnStatement || parent.Kind == ast.KindThrowStatement) {
		// Preserve the return/throw argument across line comments and ASI.
		keyword, _ := utils.TokenAtOrAfter(sourceFile, parent.Pos())
		last, _ := utils.TokenBeforePosition(sourceFile, parent.End())
		closePosition := parent.End()
		if last.Kind == ast.KindSemicolonToken {
			closePosition = last.Start
		}
		fixes = append(fixes,
			rule.RuleFixReplaceRange(core.NewTextRange(keyword.End, keyword.End), " ("),
			rule.RuleFixReplaceRange(core.NewTextRange(closePosition, closePosition), ")"),
		)
	}
	argumentStart := scanner.GetTokenPosOfNode(node.Expression(), sourceFile, false)
	if unicornutil.NeedsSemicolonBefore(sourceFile, node, text[argumentStart:node.End()]) {
		fixes = append(fixes, rule.RuleFixReplaceRange(core.NewTextRange(argumentStart, argumentStart), ";"))
	}
	return fixes
}

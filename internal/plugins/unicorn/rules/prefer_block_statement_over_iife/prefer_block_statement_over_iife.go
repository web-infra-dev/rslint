// Ported from eslint-plugin-unicorn v79.0.0. See LICENSE.
package prefer_block_statement_over_iife

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var PreferBlockStatementOverIIFERule = rule.Rule{
	Name:   "unicorn/prefer-block-statement-over-iife",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		// Annex B can also expose a block function in a sloppy CommonJS wrapper.
		nonModule := ctx.LanguageOptions.EffectiveSourceType() != "module"
		return rule.RuleListeners{
			ast.KindExpressionStatement: func(statement *ast.Node) {
				expression := utils.ESTreeRuntimeExpression(statement.Expression())
				if expression == nil || !ast.IsCallExpression(expression) || ast.IsOptionalChain(expression) {
					return
				}
				call := expression.AsCallExpression()
				if len(call.Arguments.Nodes) != 0 {
					return
				}
				callee := utils.ESTreeCallCallee(call.Expression)
				if callee == nil || (!ast.IsFunctionExpression(callee) && !ast.IsArrowFunction(callee)) {
					return
				}
				ownContext := ast.IsFunctionExpression(callee)
				if (ownContext && (callee.Name() != nil || callee.AsFunctionExpression().AsteriskToken != nil)) ||
					ast.HasModifier(callee, ast.ModifierFlagsAsync) || len(callee.Parameters()) != 0 {
					return
				}

				body := callee.Body()
				if body == nil || !ast.IsBlock(body) {
					return
				}
				statements := body.Statements()
				if (len(statements) != 0 && ast.IsPrologueDirective(statements[0])) ||
					usesFunctionScope(body, nonModule, ownContext) {
					return
				}

				file := ctx.SourceFile
				expressionRange := core.NewTextRange(scanner.GetTokenPosOfNode(expression, file, false), expression.End())
				ctx.ReportRangeWithDeferredFixes(expressionRange, rule.RuleMessage{
					Id:          "prefer-block-statement-over-iife",
					Description: "Prefer a block statement over an IIFE used only for scoping.",
				}, func() []rule.RuleFix {
					// The block cannot retain the IIFE's type parameters.
					if len(callee.TypeParameters()) != 0 {
						return nil
					}
					// An ES5 target rejects function declarations moved into a block.
					if options := ctx.Program().Options(); options != nil &&
						options.GetEmitScriptTarget() < core.ScriptTargetES2015 &&
						slices.ContainsFunc(statements, ast.IsFunctionDeclaration) {
						return nil
					}
					statementStart := scanner.GetTokenPosOfNode(statement, file, false)
					bodyStart := scanner.GetTokenPosOfNode(body, file, false)
					comments := ctx.Comments.All()
					if utils.HasCommentInSpan(comments, statementStart, bodyStart) ||
						utils.HasCommentInSpan(comments, body.End(), statement.End()) {
						return nil
					}
					return []rule.RuleFix{rule.RuleFixReplaceRange(
						core.NewTextRange(statementStart, statement.End()),
						file.Text()[bodyStart:body.End()],
					)}
				})
			},
		}
	},
}

func usesFunctionScope(body *ast.Node, nonModule, ownContext bool) bool {
	var visit ast.Visitor
	inArrow := false
	visit = func(node *ast.Node) bool {
		if ast.IsFunctionLikeDeclaration(node) && node.Body() != nil {
			switch node.Kind {
			case ast.KindFunctionDeclaration:
				return nonModule && !inArrow
			case ast.KindFunctionExpression:
				return false
			case ast.KindArrowFunction:
				if !ownContext {
					return false
				}
				outerArrow := inArrow
				inArrow = true
				found := node.ForEachChild(visit)
				inArrow = outerArrow
				return found
			default:
				// A method's key and decorators evaluate outside its function scope.
				name := node.Name()
				visitDecorator := func(child *ast.Node) bool {
					return child.Kind == ast.KindDecorator && visit(child)
				}
				return node.ForEachChild(func(child *ast.Node) bool {
					if child.Kind == ast.KindParameter {
						return child.ForEachChild(visitDecorator)
					}
					return (child == name && visit(child)) || visitDecorator(child)
				})
			}
		}

		switch node.Kind {
		case ast.KindReturnStatement:
			if !inArrow {
				return true
			}
		case ast.KindVariableDeclarationList:
			if !inArrow && !ast.IsBlockOrCatchScoped(node) {
				return true
			}
		case ast.KindCallExpression:
			// A nested arrow can read the IIFE's this or arguments through eval.
			if unicornutil.IsDirectEvalCall(node) {
				return true
			}
		case ast.KindThisKeyword:
			return ownContext
		case ast.KindMetaProperty:
			return ownContext && node.AsMetaProperty().KeywordToken == ast.KindNewKeyword
		case ast.KindIdentifier:
			return ownContext && node.Text() == "arguments" && !isJSXLiteralName(node)
		}
		return node.ForEachChild(visit)
	}
	return visit(body)
}

func isJSXLiteralName(node *ast.Node) bool {
	parent := node.Parent
	switch parent.Kind {
	case ast.KindJsxAttribute, ast.KindJsxNamespacedName:
		return true
	case ast.KindPropertyAccessExpression:
		return parent.Name() == node && utils.IsInJsxTagName(parent)
	default:
		return ast.IsJsxTagName(node)
	}
}

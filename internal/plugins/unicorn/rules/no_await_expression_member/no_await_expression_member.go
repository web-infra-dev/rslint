package no_await_expression_member

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const messageID = "no-await-expression-member"

var message = rule.RuleMessage{
	Id:          messageID,
	Description: "Do not access a member directly from an await expression.",
}

// NoAwaitExpressionMemberRule disallows member access directly from await.
//
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/rules/no-await-expression-member.js
var NoAwaitExpressionMemberRule = rule.Rule{
	Name:   "unicorn/no-await-expression-member",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		check := func(node *ast.Node) {
			object := node.Expression()
			await := utils.ESTreeRuntimeExpression(object)
			if await == nil || await.Kind != ast.KindAwaitExpression {
				return
			}

			property := node.Name()
			if node.Kind == ast.KindElementAccessExpression {
				property = utils.ESTreeRuntimeExpression(node.AsElementAccessExpression().ArgumentExpression)
			}
			if property == nil {
				return
			}

			ctx.ReportNodeWithDeferredFixes(property, message, func() []rule.RuleFix {
				return destructuringFix(ctx, node, object, property)
			})
		}
		return rule.RuleListeners{
			ast.KindPropertyAccessExpression: check,
			ast.KindElementAccessExpression:  check,
		}
	},
}

func destructuringFix(ctx rule.RuleContext, member, object, property *ast.Node) []rule.RuleFix {
	parent := utils.ESTreeParent(member)
	if ast.IsOptionalChain(member) || parent == nil || parent.Kind != ast.KindVariableDeclaration {
		return nil
	}
	// NOTE: Unlike eslint-plugin-unicorn v74.0.0, do not fix using or await
	// using declarations: resource declarations cannot use binding patterns.
	declarationList := parent.Parent
	if declarationList != nil && declarationList.Kind == ast.KindVariableDeclarationList &&
		(ast.IsVarUsing(declarationList) || ast.IsVarAwaitUsing(declarationList)) {
		return nil
	}
	declaration := parent.AsVariableDeclaration()
	variable := declaration.Name()
	if utils.ESTreeRuntimeExpression(declaration.Initializer) != member ||
		variable == nil || variable.Kind != ast.KindIdentifier || declaration.Type != nil {
		return nil
	}

	var opening, closing string
	if member.Kind == ast.KindElementAccessExpression {
		if property.Kind != ast.KindNumericLiteral {
			return nil
		}
		switch utils.NormalizeNumericLiteral(property.AsNumericLiteral().Text) {
		case "0":
			opening, closing = "[", "]"
		case "1":
			opening, closing = "[, ", "]"
		default:
			return nil
		}
	} else {
		if property.Kind != ast.KindIdentifier || variable.Text() != property.Text() {
			return nil
		}
		opening, closing = "{", "}"
	}

	fixes := []rule.RuleFix{
		rule.RuleFixInsertBefore(ctx.SourceFile, variable, opening),
		rule.RuleFixInsertAfter(variable, closing),
		rule.RuleFixRemoveRange(core.NewTextRange(object.End(), member.End())),
	}
	// Remove only the parentheses around await, preserving their comments and
	// whitespace. Parentheses around the complete member expression stay intact.
	// JSDoc casts are transparent to ESTree; authored TS assertions are not.
	for current := object; current != nil; {
		if current.Kind == ast.KindParenthesizedExpression {
			span := utils.TrimNodeTextRange(ctx.SourceFile, current)
			fixes = append(fixes,
				rule.RuleFixRemoveRange(core.NewTextRange(span.Pos(), span.Pos()+1)),
				rule.RuleFixRemoveRange(core.NewTextRange(span.End()-1, span.End())),
			)
			current = current.AsParenthesizedExpression().Expression
			continue
		}
		if expression := utils.JSDocTypeCastExpression(current); expression != nil {
			current = expression
			continue
		}
		break
	}
	return fixes
}

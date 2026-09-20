package prefer_reflect_apply

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var message = rule.RuleMessage{
	Id:          "prefer-reflect-apply",
	Description: "Prefer `Reflect.apply()` over `Function#apply()`.",
}

// PreferReflectApplyRule prefers Reflect.apply over Function#apply.
//
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/rules/prefer-reflect-apply.js
var PreferReflectApplyRule = rule.Rule{
	Name:   "unicorn/prefer-reflect-apply",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		// Upstream deliberately does not resolve variables in computed keys.
		evaluator := utils.NewStaticStringEvaluatorWithoutScope()
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				if ast.IsOptionalChainRoot(node) {
					return
				}
				callee := utils.ESTreeCallCallee(node.Expression())
				if !isMember(callee) || ast.IsOptionalChainRoot(callee) {
					return
				}
				object := utils.ESTreeRuntimeExpression(callee.Expression())
				if object == nil {
					return
				}
				switch object.Kind {
				case ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindBigIntLiteral,
					ast.KindRegularExpressionLiteral, ast.KindTrueKeyword, ast.KindFalseKeyword,
					ast.KindNullKeyword, ast.KindArrayLiteralExpression, ast.KindObjectLiteralExpression:
					return
				}

				name, _ := evaluator.EvalAccessExpressionName(callee)
				args := node.Arguments()
				var target, receiver, argumentsList *ast.Node
				discardedMembers := [3]*ast.Node{callee}
				switch {
				case name == "apply" && len(args) == 2:
					target, receiver, argumentsList = object, args[0], args[1]
				case name == "call" && len(args) == 3:
					apply := utils.ESTreeCallCallee(callee.Expression())
					if !isMemberNamed(evaluator, apply, "apply") {
						return
					}
					prototype := utils.ESTreeCallCallee(apply.Expression())
					if !isMemberNamed(evaluator, prototype, "prototype") {
						return
					}
					constructor := utils.ESTreeRuntimeExpression(prototype.Expression())
					if constructor.Kind != ast.KindIdentifier || constructor.Text() != "Function" {
						return
					}
					target, receiver, argumentsList = utils.ESTreeRuntimeExpression(args[0]), args[1], args[2]
					discardedMembers[1], discardedMembers[2] = apply, prototype
				default:
					return
				}
				receiver = utils.ESTreeRuntimeExpression(receiver)
				argumentsList = utils.ESTreeRuntimeExpression(argumentsList)
				if (receiver.Kind != ast.KindNullKeyword && receiver.Kind != ast.KindThisKeyword) ||
					(argumentsList.Kind != ast.KindArrayLiteralExpression &&
						(argumentsList.Kind != ast.KindIdentifier || argumentsList.Text() != "arguments")) {
					return
				}

				ctx.ReportNodeWithDeferredFixes(node, message, func() []rule.RuleFix {
					// A spread target can shift the runtime receiver/argument-list
					// positions. Also avoid bare super and broken optional chains.
					if target.Kind == ast.KindSpreadElement || target.Kind == ast.KindSuperKeyword || ast.IsOptionalChain(node) {
						return nil
					}
					// Knowing a computed key's value does not make it safe to drop.
					// Check only removed keys: effects in the target and arguments
					// are retained by the replacement.
					for _, member := range discardedMembers {
						if member != nil && member.Kind == ast.KindElementAccessExpression {
							key := member.AsElementAccessExpression().ArgumentExpression
							if hasImplicitKeyEffects(evaluator, key) {
								return nil
							}
							if _, safe := evaluator.EvalControlFlowValue(key); !safe {
								return nil
							}
						}
					}
					targetText := utils.TrimmedNodeText(ctx.SourceFile, target)
					// A sequence must stay one argument after moving into the call.
					if utils.IsCommaOperator(target) {
						targetText = "(" + targetText + ")"
					}
					text := "Reflect.apply(" + targetText + ", " +
						utils.TrimmedNodeText(ctx.SourceFile, receiver) + ", " +
						utils.TrimmedNodeText(ctx.SourceFile, argumentsList) + ")"
					return []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, node, text)}
				})
			},
		}
	},
}

// Control-flow evaluation does not model implicit execution from class/JSX
// creation, tagged templates or spread iteration. Unknown references may also
// throw even when the evaluator can fold the containing sequence expression.
func hasImplicitKeyEffects(evaluator *utils.StaticStringEvaluator, node *ast.Node) bool {
	node = utils.SkipAssertionsAndParens(node)
	if node == nil || ast.IsTypeNode(node) {
		return false
	}
	switch node.Kind {
	case ast.KindClassExpression, ast.KindTaggedTemplateExpression,
		ast.KindJsxElement, ast.KindJsxSelfClosingElement, ast.KindJsxFragment,
		ast.KindSpreadElement, ast.KindSpreadAssignment,
		ast.KindThisKeyword, ast.KindSuperKeyword:
		return true
	case ast.KindIdentifier:
		return !utils.IsNonReferenceIdentifier(node)
	case ast.KindFunctionExpression, ast.KindArrowFunction:
		// Creating a function does not execute its body or default parameters.
		return false
	case ast.KindBinaryExpression, ast.KindPrefixUnaryExpression, ast.KindTemplateExpression:
		// Coercion can call user code or throw (for example, 1n / 0n).
		// Check each operand too, even if an enclosing sequence ignores it.
		if _, known := evaluator.EvalValue(node); !known {
			return true
		}
	}
	return node.ForEachChild(func(child *ast.Node) bool { return hasImplicitKeyEffects(evaluator, child) })
}

func isMember(node *ast.Node) bool {
	return node != nil && (node.Kind == ast.KindPropertyAccessExpression || node.Kind == ast.KindElementAccessExpression)
}

func isMemberNamed(evaluator *utils.StaticStringEvaluator, node *ast.Node, name string) bool {
	if !isMember(node) {
		return false
	}
	actual, ok := evaluator.EvalAccessExpressionName(node)
	return ok && actual == name
}

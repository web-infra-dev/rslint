package no_single_promise_in_promise_methods

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v77.0.0/rules/no-single-promise-in-promise-methods.js
var NoSinglePromiseInPromiseMethodsRule = rule.Rule{
	Name:   "unicorn/no-single-promise-in-promise-methods",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		oneArgument := 1
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{
					Methods:             []string{"all", "any", "race"},
					ArgumentsLength:     &oneArgument,
					RejectSpreadElement: true,
				})
				if !ok {
					return
				}
				object := utils.ESTreeRuntimeExpression(call.Object)
				if !ast.IsIdentifier(object) || object.Text() != "Promise" {
					return
				}
				array := utils.ESTreeRuntimeExpression(node.Arguments()[0])
				if array.Kind != ast.KindArrayLiteralExpression {
					return
				}
				elements := array.AsArrayLiteralExpression().Elements.Nodes
				if len(elements) != 1 || elements[0].Kind == ast.KindOmittedExpression || elements[0].Kind == ast.KindSpreadElement {
					return
				}
				element := elements[0]
				method := call.Property.Text()
				message := rule.RuleMessage{
					Id:          "no-single-promise-in-promise-methods/error",
					Description: "Wrapping single-element array with `Promise." + method + "()` is unnecessary.",
					Data:        map[string]string{"method": method},
				}
				// Unwrapping Promise.any changes rejection from AggregateError to
				// the input's rejection reason, even with only one input.
				if method == "any" {
					ctx.ReportNode(array, message)
					return
				}
				if method == "all" {
					ctx.ReportNodeWithDeferredFixes(array, message, func() []rule.RuleFix {
						return fixPromiseAll(ctx, node, element)
					})
					return
				}
				if utils.ESTreeParent(node).Kind == ast.KindAwaitExpression {
					ctx.ReportNodeWithDeferredFixes(array, message, func() []rule.RuleFix {
						if hasCommentsInside(ctx, node) {
							return nil
						}
						return []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, node, awaitedPromiseText(ctx.SourceFile, element))}
					})
					return
				}
				ctx.ReportNodeWithDeferredSuggestions(array, message, func() []rule.RuleSuggestion {
					var suggestions []rule.RuleSuggestion
					if !hasCommentsInside(ctx, node) {
						text := utils.TrimmedNodeText(ctx.SourceFile, element)
						expression := utils.ESTreeRuntimeExpression(element)
						canSkipParens := ast.IsIdentifier(expression) ||
							(ast.IsAccessExpression(expression) && !ast.IsOptionalChain(expression))
						switch ast.GetLeftmostExpression(expression, false).Kind {
						case ast.KindObjectLiteralExpression, ast.KindFunctionExpression, ast.KindClassExpression:
							canSkipParens = false
						}
						if element == expression && !canSkipParens {
							text = "(" + text + ")"
						}
						if unicornutil.NeedsSemicolonBefore(ctx.SourceFile, node, text) {
							text = ";" + text
						}
						suggestions = append(suggestions, rule.RuleSuggestion{
							Message: rule.RuleMessage{
								Id:          "no-single-promise-in-promise-methods/unwrap",
								Description: "Use the value directly.",
							},
							FixesArr: []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, node, text)},
						})
					}
					return append(suggestions, rule.RuleSuggestion{
						Message: rule.RuleMessage{
							Id:          "no-single-promise-in-promise-methods/use-promise-resolve",
							Description: "Switch to `Promise.resolve(…)`.",
						},
						FixesArr: switchToPromiseResolve(ctx.SourceFile, call.Property, array),
					})
				})
			},
		}
	},
}

func hasCommentsInside(ctx rule.RuleContext, node *ast.Node) bool {
	r := utils.TrimNodeTextRange(ctx.SourceFile, node)
	return utils.HasCommentInSpan(ctx.Comments.All(), r.Pos(), r.End())
}

// Check only the wrappers of this expression, not constraints on outer results.
func hasJSDocTypeCast(node *ast.Node) bool {
	for parent := node.Parent; parent != nil; parent = parent.Parent {
		if utils.IsJSDocTypeCastWrapper(parent) {
			return true
		}
		if parent.Kind != ast.KindParenthesizedExpression {
			break
		}
	}
	return false
}

func awaitedPromiseText(sourceFile *ast.SourceFile, element *ast.Node) string {
	text := utils.TrimmedNodeText(sourceFile, element)
	if ast.GetExpressionPrecedence(element) < ast.OperatorPrecedenceUnary {
		text = "(" + text + ")"
	}
	return text
}

func switchToPromiseResolve(sourceFile *ast.SourceFile, property, array *ast.Node) []rule.RuleFix {
	r := utils.TrimNodeTextRange(sourceFile, array)
	fixes := []rule.RuleFix{
		rule.RuleFixReplace(sourceFile, property, "resolve"),
		rule.RuleFixRemoveRange(core.NewTextRange(r.Pos(), r.Pos()+1)),
		rule.RuleFixRemoveRange(core.NewTextRange(r.End()-1, r.End())),
	}
	element := array.AsArrayLiteralExpression().Elements.Nodes[0]
	if token, ok := utils.TokenAtOrAfter(sourceFile, element.End()); ok && token.Kind == ast.KindCommaToken {
		fixes = append(fixes, rule.RuleFixRemoveRange(token.Range()))
	}
	return fixes
}

func fixPromiseAll(ctx rule.RuleContext, call, element *ast.Node) []rule.RuleFix {
	if hasCommentsInside(ctx, call) || hasJSDocTypeCast(call) {
		return nil
	}
	await := utils.ESTreeParent(call)
	if await.Kind != ast.KindAwaitExpression || hasJSDocTypeCast(await) {
		return nil
	}
	parent := utils.ESTreeParent(await)
	if parent.Kind == ast.KindExpressionStatement {
		return []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, call, awaitedPromiseText(ctx.SourceFile, element))}
	}
	var pattern *ast.Node
	if parent.Kind == ast.KindVariableDeclaration {
		declaration := parent.AsVariableDeclaration()
		// Both TypeScript and JSDoc annotations describe the original tuple.
		if declaration.Type == nil && utils.ESTreeRuntimeExpression(declaration.Initializer) == await {
			pattern = declaration.Name()
		}
	} else if ast.IsAssignmentExpression(parent, true) &&
		utils.ESTreeRuntimeExpression(parent.AsBinaryExpression().Right) == await &&
		utils.ESTreeParent(parent).Kind == ast.KindExpressionStatement && !hasJSDocTypeCast(parent) {
		pattern = utils.ESTreeRuntimeExpression(parent.AsBinaryExpression().Left)
	}
	if identifier := singleBindingIdentifier(pattern); identifier != nil && !hasCommentsInside(ctx, pattern) {
		return []rule.RuleFix{
			rule.RuleFixReplace(ctx.SourceFile, pattern, utils.TrimmedNodeText(ctx.SourceFile, identifier)),
			rule.RuleFixReplace(ctx.SourceFile, call, awaitedPromiseText(ctx.SourceFile, element)),
		}
	}

	// Replacing the whole indexed result with an await is safe only in the
	// initializer/assignment positions accepted by upstream.
	if parent.Kind != ast.KindElementAccessExpression || ast.IsOptionalChain(parent) {
		return nil
	}
	member := parent.AsElementAccessExpression()
	index := utils.ESTreeRuntimeExpression(member.ArgumentExpression)
	if utils.ESTreeRuntimeExpression(member.Expression) != await || index.Kind != ast.KindNumericLiteral {
		return nil
	}
	value, ok := ecmascript.StringToNumber(index.Text())
	if !ok || value != 0 || hasCommentsInside(ctx, parent) {
		return nil
	}
	outer := utils.ESTreeParent(parent)
	if (outer.Kind == ast.KindVariableDeclaration && utils.ESTreeRuntimeExpression(outer.AsVariableDeclaration().Initializer) == parent) ||
		(ast.IsAssignmentExpression(outer, false) && utils.ESTreeRuntimeExpression(outer.AsBinaryExpression().Right) == parent) {
		return []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, parent, "await "+awaitedPromiseText(ctx.SourceFile, element))}
	}
	return nil
}

func singleBindingIdentifier(pattern *ast.Node) *ast.Node {
	if pattern == nil {
		return nil
	}
	var elements []*ast.Node
	switch pattern.Kind {
	case ast.KindArrayBindingPattern:
		elements = pattern.AsBindingPattern().Elements.Nodes
	case ast.KindArrayLiteralExpression:
		elements = pattern.AsArrayLiteralExpression().Elements.Nodes
	default:
		return nil
	}
	if len(elements) != 1 {
		return nil
	}
	element := elements[0]
	if element.Kind == ast.KindBindingElement {
		binding := element.AsBindingElement()
		if binding.DotDotDotToken != nil || binding.Initializer != nil {
			return nil
		}
		element = binding.Name()
	}
	element = utils.ESTreeRuntimeExpression(element)
	if ast.IsIdentifier(element) {
		return element
	}
	return nil
}

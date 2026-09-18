// Ported from eslint-plugin-unicorn v75.0.0; see LICENSE.
package no_useless_error_capture_stack_trace

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var NoUselessErrorCaptureStackTraceRule = rule.Rule{
	Name:   "unicorn/no-useless-error-capture-stack-trace",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		argumentCount := 2
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{
					Method: "captureStackTrace", ArgumentsLength: &argumentCount, RejectSpreadElement: true, AllowOptionalCall: true,
				})
				if !ok {
					return
				}
				object := utils.ESTreeRuntimeExpression(call.Object)
				if !ast.IsIdentifier(object) || object.Text() != "Error" || !ctx.Globals.Access("Error").IsDeclared() || !ctx.Refs.IsGlobalReference(object) {
					return
				}
				class := constructorClass(node)
				if class == nil || !extendsBuiltinError(ctx, class) {
					return
				}
				args := node.Arguments()
				if utils.ESTreeRuntimeExpression(args[0]).Kind != ast.KindThisKeyword || !isClassReference(ctx, utils.ESTreeRuntimeExpression(args[1]), class) {
					return
				}
				ctx.ReportNodeWithDeferredFixes(node, rule.RuleMessage{
					Id: "no-useless-error-capture-stack-trace/error", Description: "Unnecessary `Error.captureStackTrace(…)` call.",
				}, func() []rule.RuleFix {
					statement := utils.ESTreeParent(node)
					if statement == nil || !ast.IsExpressionStatement(statement) || statement.Parent == nil || statement.Parent.Kind != ast.KindBlock {
						return nil
					}
					return []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, statement, "")}
				})
			},
		}
	},
}

// Arrow functions retain the enclosing constructor's this binding. Ordinary
// functions and nested classes terminate the search; no traversal stacks are needed.
func constructorClass(node *ast.Node) *ast.Node {
	child := node
	for parent := node.Parent; parent != nil; parent = parent.Parent {
		if ast.IsClassLike(parent) || parent.Kind == ast.KindClassStaticBlockDeclaration {
			return nil
		}
		if ast.IsFunctionLike(parent) && parent.Kind != ast.KindArrowFunction && parent.Body() != nil {
			// Computed keys and decorators sit outside the method function in ESTree.
			if child == parent.Name() || child.Kind == ast.KindDecorator {
				child = parent
				continue
			}
			if parent.Kind == ast.KindConstructor && !ast.HasStaticModifier(parent) && parent.Parent != nil && ast.IsClassLike(parent.Parent) {
				return parent.Parent
			}
			return nil
		}
		child = parent
	}
	return nil
}

func extendsBuiltinError(ctx rule.RuleContext, class *ast.Node) bool {
	heritage := ast.GetClassExtendsHeritageElement(class)
	if heritage == nil {
		return false
	}
	base := utils.ESTreeRuntimeExpression(heritage.AsExpressionWithTypeArguments().Expression)
	if !ast.IsIdentifier(base) || !ctx.Globals.Access(base.Text()).IsDeclared() || !ctx.Refs.IsGlobalReference(base) {
		return false
	}
	switch base.Text() {
	case "Error", "EvalError", "RangeError", "ReferenceError", "SyntaxError", "TypeError", "URIError", "AggregateError", "SuppressedError":
		return true
	default:
		return false
	}
}

func isClassReference(ctx rule.RuleContext, node, class *ast.Node) bool {
	if node.Kind == ast.KindMetaProperty {
		meta := node.AsMetaProperty()
		return meta.KeywordToken == ast.KindNewKeyword && node.Name().Text() == "target"
	}
	if ast.IsPropertyAccessExpression(node) && !ast.IsOptionalChainRoot(node) {
		access := node.AsPropertyAccessExpression()
		return ast.IsIdentifier(access.Name()) && access.Name().Text() == "constructor" && utils.ESTreeRuntimeExpression(access.Expression).Kind == ast.KindThisKeyword
	}
	if !ast.IsIdentifier(node) || class.Name() == nil {
		return false
	}
	symbol := ctx.Refs.ResolveInFile(node)
	return symbol != nil && ast.GetClassLikeDeclarationOfSymbol(symbol) == class
}

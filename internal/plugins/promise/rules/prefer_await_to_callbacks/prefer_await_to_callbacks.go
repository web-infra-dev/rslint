package prefer_await_to_callbacks

import (
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var callbackMessage = rule.RuleMessage{
	Id:          "error",
	Description: "Avoid callbacks. Prefer Async/Await.",
}

// parameterName mirrors ESTree's parameter.name: defaults, rest parameters,
// destructuring and TypeScript parameter properties are separate node kinds.
func parameterName(node *ast.Node) string {
	param := node.AsParameterDeclaration()
	if param.Initializer != nil || param.DotDotDotToken != nil ||
		ast.IsParameterPropertyDeclaration(node, node.Parent) {
		return ""
	}
	return propertyName(param.Name())
}

// Upstream reads property.name even for computed and private properties.
// String literals have no name, while PrivateIdentifier names omit the '#'.
func propertyName(node *ast.Node) string {
	if node == nil {
		return ""
	}
	switch node.Kind {
	case ast.KindIdentifier:
		return node.AsIdentifier().Text
	case ast.KindPrivateIdentifier:
		return strings.TrimPrefix(node.AsPrivateIdentifier().Text, "#")
	default:
		return ""
	}
}

func isCallbackName(name string) bool {
	return name == "cb" || name == "callback"
}

func isArrayMethod(name string) bool {
	switch name {
	case "map", "every", "forEach", "some", "find", "filter":
		return true
	default:
		return false
	}
}

func isInsideYieldOrAwait(node *ast.Node) bool {
	return ast.FindAncestor(node.Parent, func(parent *ast.Node) bool {
		return parent.Kind == ast.KindAwaitExpression || parent.Kind == ast.KindYieldExpression
	}) != nil
}

var PreferAwaitToCallbacksRule = rule.Rule{
	Name:   "promise/prefer-await-to-callbacks",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		checkLastParameter := func(node *ast.Node) {
			// Body-absent TypeScript declarations are TSDeclareFunction or
			// TSEmptyBodyFunctionExpression nodes, outside upstream's listeners.
			if node.Body() == nil {
				return
			}
			// Only inspect the last authored parameter; cloning/filtering the
			// entire list is unnecessary. JSDoc can synthesize a this parameter.
			for _, last := range slices.Backward(node.Parameters()) {
				if utils.IsJSDocSyntaxNode(last) {
					continue
				}
				if isCallbackName(parameterName(last)) {
					ctx.ReportRange(utils.GetESTreeBindingIdentifierRange(ctx.SourceFile, last.Name()), callbackMessage)
				}
				break
			}
		}
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				// Ordinary import() is an ESTree ImportExpression, not a call.
				// IsImportCall also includes import.defer(), which the pinned
				// reference parser still exposes as a CallExpression.
				if call.Expression.Kind == ast.KindImportKeyword {
					return
				}
				callee := utils.ESTreeCallCallee(call.Expression)
				if isCallbackName(propertyName(callee)) {
					ctx.ReportNode(node, callbackMessage)
					return
				}
				args := call.Arguments.Nodes
				if len(args) == 0 {
					return
				}
				arg := utils.ESTreeRuntimeExpression(args[len(args)-1])
				if arg.Kind != ast.KindFunctionExpression && arg.Kind != ast.KindArrowFunction {
					return
				}
				object, property := utils.MemberExpressionParts(callee)
				name := propertyName(utils.ESTreeRuntimeExpression(property))
				if name == "on" || name == "once" {
					return
				}
				objectName := propertyName(utils.ESTreeRuntimeExpression(object))
				isLodash := objectName == "lodash" || objectName == "underscore" || objectName == "_"
				if (isArrayMethod(name) && (len(args) == 1 || (len(args) == 2 && isLodash))) ||
					(isArrayMethod(propertyName(callee)) && len(args) == 2) {
					return
				}
				for _, first := range arg.Parameters() {
					if utils.IsJSDocSyntaxNode(first) {
						continue
					}
					firstName := parameterName(first)
					if (firstName == "err" || firstName == "error") && !isInsideYieldOrAwait(node) {
						ctx.ReportNode(arg, callbackMessage)
					}
					break
				}
			},
			ast.KindFunctionDeclaration: checkLastParameter,
			ast.KindFunctionExpression:  checkLastParameter,
			ast.KindArrowFunction:       checkLastParameter,
			// ESTree represents method/accessor/constructor values as functions.
			ast.KindMethodDeclaration: checkLastParameter,
			ast.KindGetAccessor:       checkLastParameter,
			ast.KindSetAccessor:       checkLastParameter,
			ast.KindConstructor:       checkLastParameter,
		}
	},
}

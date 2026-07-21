package no_promise_in_callback

import (
	_ "embed"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/promiseutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

//go:embed no_promise_in_callback.schema.json
var schemaJSON []byte

const skipTransparent = ast.OEKParentheses

type Options struct {
	ExemptDeclarations bool
}

func parseOptions(options []any) Options {
	opts := Options{}
	if len(options) == 0 {
		return opts
	}
	optsMap, _ := options[0].(map[string]interface{})
	opts.ExemptDeclarations, _ = optsMap["exemptDeclarations"].(bool)
	return opts
}

func messageAvoidPromiseInCallback() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "avoidPromiseInCallback",
		Description: "Avoid using promises inside of callbacks.",
	}
}

func isReturnExpression(node *ast.Node) bool {
	current := ast.WalkUpParenthesizedExpressions(node.Parent)
	return current != nil && current.Kind == ast.KindReturnStatement && !hasOptionalChain(node)
}

func hasOptionalChain(node *ast.Node) bool {
	node = ast.SkipOuterExpressions(node, skipTransparent)
	if node == nil {
		return false
	}
	if ast.IsOptionalChain(node) {
		return true
	}
	switch node.Kind {
	case ast.KindCallExpression:
		call := node.AsCallExpression()
		return call != nil && hasOptionalChain(call.Expression)
	case ast.KindPropertyAccessExpression:
		access := node.AsPropertyAccessExpression()
		return access != nil && hasOptionalChain(access.Expression)
	case ast.KindElementAccessExpression:
		access := node.AsElementAccessExpression()
		return access != nil && hasOptionalChain(access.Expression)
	case ast.KindNonNullExpression:
		expression := node.AsNonNullExpression()
		return expression != nil && hasOptionalChain(expression.Expression)
	default:
		return false
	}
}

func isPromiseCallback(node *ast.Node) bool {
	if node == nil || (node.Kind != ast.KindFunctionExpression && node.Kind != ast.KindArrowFunction) {
		return false
	}
	parent := ast.WalkUpParenthesizedExpressions(node.Parent)
	return promiseutil.IsMemberCall(parent, "then") || promiseutil.IsMemberCall(parent, "catch")
}

func firstParameterName(node *ast.Node) string {
	if node == nil || node.Parameters() == nil || len(node.Parameters()) == 0 {
		return ""
	}
	param := node.Parameters()[0]
	if param == nil || !ast.IsParameterDeclaration(param) {
		return ""
	}
	if param.Parent != nil && ast.IsParameterPropertyDeclaration(param, param.Parent) {
		return ""
	}
	decl := param.AsParameterDeclaration()
	if decl.Initializer != nil || decl.DotDotDotToken != nil {
		return ""
	}
	name := decl.Name()
	if name != nil && ast.IsIdentifier(name) {
		return name.AsIdentifier().Text
	}
	return ""
}

func isCallbackContainer(node *ast.Node, opts Options) bool {
	if node == nil || isPromiseCallback(node) {
		return false
	}

	// Object/class methods, accessors, and constructors are callback containers
	// too: in ESTree their bodies are FunctionExpressions, so upstream flags a
	// method whose first parameter is err/error. tsgo represents them as
	// dedicated declaration kinds, so enumerate them here. exemptDeclarations
	// only affects FunctionDeclaration, matching upstream.
	isFunction := node.Kind == ast.KindFunctionExpression || node.Kind == ast.KindArrowFunction ||
		node.Kind == ast.KindMethodDeclaration || node.Kind == ast.KindGetAccessor ||
		node.Kind == ast.KindSetAccessor || node.Kind == ast.KindConstructor ||
		(!opts.ExemptDeclarations && node.Kind == ast.KindFunctionDeclaration)
	if !isFunction {
		return false
	}

	name := firstParameterName(node)
	return name == "err" || name == "error"
}

func findCallbackAncestor(node *ast.Node, opts Options) *ast.Node {
	for current := node.Parent; current != nil; current = current.Parent {
		if isCallbackContainer(current, opts) {
			return current
		}
	}
	return nil
}

var NoPromiseInCallbackRule = rule.Rule{
	Name:   "promise/no-promise-in-callback",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				if !promiseutil.IsPromiseLikeCall(node) || isReturnExpression(node) {
					return
				}
				if findCallbackAncestor(node, opts) == nil {
					return
				}
				callee := ast.SkipOuterExpressions(node.AsCallExpression().Expression, skipTransparent)
				if callee == nil {
					callee = node
				}
				ctx.ReportNode(callee, messageAvoidPromiseInCallback())
			},
		}
	},
}

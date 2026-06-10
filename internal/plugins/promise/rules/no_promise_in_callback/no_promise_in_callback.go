package no_promise_in_callback

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/promiseutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const skipTransparent = ast.OEKParentheses

type Options struct {
	ExemptDeclarations bool
}

func parseOptions(options any) Options {
	opts := Options{}
	optsMap := utils.GetOptionsMap(options)
	if optsMap == nil {
		return opts
	}
	if value, ok := optsMap["exemptDeclarations"].(bool); ok {
		opts.ExemptDeclarations = value
	}
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
	return current != nil && current.Kind == ast.KindReturnStatement
}

func isPromiseCallback(node *ast.Node) bool {
	if node == nil || (node.Kind != ast.KindFunctionExpression && node.Kind != ast.KindArrowFunction) {
		return false
	}
	parent := ast.WalkUpParenthesizedExpressions(node.Parent)
	if parent == nil || !ast.IsCallExpression(parent) {
		return false
	}
	call := parent.AsCallExpression()
	if call == nil || call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
		return false
	}
	callee := ast.SkipOuterExpressions(call.Expression, skipTransparent)
	if callee == nil || !ast.IsPropertyAccessExpression(callee) {
		return false
	}
	name := callee.AsPropertyAccessExpression().Name()
	if name == nil || !ast.IsIdentifier(name) {
		return false
	}
	text := name.AsIdentifier().Text
	return text == "then" || text == "catch"
}

func firstParameterName(node *ast.Node) string {
	if node == nil || node.Parameters() == nil || len(node.Parameters()) == 0 {
		return ""
	}
	param := node.Parameters()[0]
	if param == nil || !ast.IsParameterDeclaration(param) {
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

	isFunction := node.Kind == ast.KindFunctionExpression || node.Kind == ast.KindArrowFunction ||
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
	Name: "promise/no-promise-in-callback",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
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

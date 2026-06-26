package no_return_wrap

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/promiseutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

const skipTransparent = ast.OEKParentheses

type Options struct {
	AllowReject bool
}

func buildResolveMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "resolve",
		Description: "Avoid wrapping return values in Promise.resolve",
	}
}

func buildRejectMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "reject",
		Description: "Expected throw instead of Promise.reject",
	}
}

func checkCallExpression(ctx rule.RuleContext, opts Options, callNode *ast.Node, reportNode *ast.Node) {
	if !isInPromise(reportNode) {
		return
	}
	// Bail on optional calls: Promise.resolve?.()
	if callNode.AsCallExpression().QuestionDotToken != nil {
		return
	}
	callee := ast.SkipOuterExpressions(callNode.AsCallExpression().Expression, skipTransparent)
	if callee == nil || !ast.IsPropertyAccessExpression(callee) {
		return
	}
	prop := callee.AsPropertyAccessExpression()
	if prop.QuestionDotToken != nil {
		return
	}
	object := ast.SkipOuterExpressions(prop.Expression, skipTransparent)
	if object == nil || !ast.IsIdentifier(object) || object.AsIdentifier().Text != "Promise" {
		return
	}
	name := prop.Name()
	if name == nil || !ast.IsIdentifier(name) {
		return
	}
	switch name.AsIdentifier().Text {
	case "resolve":
		ctx.ReportNode(reportNode, buildResolveMessage())
	case "reject":
		if !opts.AllowReject {
			ctx.ReportNode(reportNode, buildRejectMessage())
		}
	}
}

func isInPromise(node *ast.Node) bool {
	functionNode := nearestFunctionBoundary(node)
	if functionNode == nil {
		return false
	}
	for {
		parent := functionNode.Parent
		for parent != nil && ast.IsOuterExpression(parent, skipTransparent) {
			parent = parent.Parent
		}
		if parent == nil || !ast.IsPropertyAccessExpression(parent) {
			break
		}
		prop := parent.AsPropertyAccessExpression()
		if ast.SkipOuterExpressions(prop.Expression, skipTransparent) != functionNode {
			break
		}
		name := prop.Name()
		if name == nil || !ast.IsIdentifier(name) || name.AsIdentifier().Text != "bind" {
			break
		}
		call := parent.Parent
		for call != nil && ast.IsOuterExpression(call, skipTransparent) {
			call = call.Parent
		}
		if call == nil || !ast.IsCallExpression(call) || ast.SkipOuterExpressions(call.AsCallExpression().Expression, skipTransparent) != parent {
			break
		}
		functionNode = call
	}
	cur := functionNode.Parent
	for cur != nil && ast.IsOuterExpression(cur, skipTransparent) {
		cur = cur.Parent
	}
	return cur != nil && promiseutil.IsPromiseLikeCall(cur)
}

func nearestFunctionBoundary(node *ast.Node) *ast.Node {
	for cur := node.Parent; cur != nil; cur = cur.Parent {
		switch cur.Kind {
		case ast.KindFunctionExpression,
			ast.KindArrowFunction,
			ast.KindMethodDeclaration,
			ast.KindGetAccessor,
			ast.KindSetAccessor,
			ast.KindConstructor:
			return cur
		}
	}
	return nil
}

func isExpressionBodyArrowCall(node *ast.Node) bool {
	if node == nil || !ast.IsCallExpression(node) {
		return false
	}
	parent := node.Parent
	for parent != nil && ast.IsOuterExpression(parent, skipTransparent) {
		parent = parent.Parent
	}
	if parent == nil || parent.Kind != ast.KindArrowFunction {
		return false
	}
	return ast.SkipOuterExpressions(parent.Body(), skipTransparent) == node
}

var NoReturnWrapRule = rule.Rule{
	Name: "promise/no-return-wrap",
	Schema: rule.Tuple(rule.Object(map[string]rule.Schema{
		"allowReject": rule.Bool().Default(false),
	})),
	RunWithOptions: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		optsMap, _ := options[0].(map[string]any)
		allowReject, _ := optsMap["allowReject"].(bool)
		opts := Options{
			AllowReject: allowReject,
		}

		return rule.RuleListeners{
			ast.KindReturnStatement: func(node *ast.Node) {
				arg := node.AsReturnStatement().Expression
				if arg != nil {
					arg = ast.SkipOuterExpressions(arg, skipTransparent)
				}
				if arg == nil || !ast.IsCallExpression(arg) {
					return
				}
				checkCallExpression(ctx, opts, arg, node)
			},
			ast.KindCallExpression: func(node *ast.Node) {
				if isExpressionBodyArrowCall(node) {
					checkCallExpression(ctx, opts, node, node)
				}
			},
		}
	},
}

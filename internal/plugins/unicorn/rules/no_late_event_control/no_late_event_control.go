package no_late_event_control

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/no-late-event-control.js
var NoLateEventControlRule = rule.Rule{
	Name:   "unicorn/no-late-event-control",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		var timing handlerTiming
		var keys unicornutil.ComputedKeyEvaluator
		// Repeated calls on the same parameter reuse its declaration's type.
		var lastBinding *ast.Node
		var lastIsEvent bool
		return rule.RuleListeners{
			rule.ListenerOnExit(ast.KindAwaitExpression): timing.markSuspended,
			rule.ListenerOnExit(ast.KindYieldExpression): timing.markSuspended,
			rule.ListenerOnExit(ast.KindForOfStatement): func(node *ast.Node) {
				if isAsyncIteration(node) {
					timing.markSuspended(node)
				}
			},
			rule.ListenerOnExit(ast.KindCallExpression): func(node *ast.Node) {
				event, method := eventControlCall(node, &keys)
				if event == nil {
					return
				}
				parameter, binding := parameterBinding(ctx.Refs.ResolveInFile(event))
				if parameter == nil {
					return
				}

				message := timing.lateCall(node, parameter.Parent)
				if message.Id == "" {
					return
				}
				if binding != lastBinding {
					lastBinding = binding
					lastIsEvent = isEventParameter(ctx, parameter, binding)
				}
				if !lastIsEvent {
					return
				}
				message.Description = "`" + event.Text() + "." + method + "()` " + message.Description
				ctx.ReportNode(node, message)
			},
		}
	},
}

func eventControlCall(node *ast.Node, keys *unicornutil.ComputedKeyEvaluator) (*ast.Node, string) {
	callee := utils.ESTreeCallCallee(node.Expression())
	if callee == nil || !ast.IsAccessExpression(callee) {
		return nil, ""
	}

	var method string
	if ast.IsElementAccessExpression(callee) {
		method, _ = keys.Eval(callee.AsElementAccessExpression().ArgumentExpression)
	} else {
		method, _ = utils.AccessExpressionStaticName(callee)
	}
	switch method {
	case "preventDefault", "stopPropagation", "stopImmediatePropagation":
	default:
		return nil, ""
	}

	event := ast.SkipOuterExpressions(callee.Expression(), ast.OEKParentheses|ast.OEKAssertions|ast.OEKExpressionsWithTypeArguments)
	if event == nil || !ast.IsIdentifier(event) {
		return nil, ""
	}
	return event, method
}

func parameterBinding(symbol *ast.Symbol) (parameter, binding *ast.Node) {
	if symbol == nil {
		return nil, nil
	}
	for _, declaration := range symbol.Declarations {
		switch declaration.Kind {
		case ast.KindParameter:
			return declaration, declaration.Name()
		case ast.KindBindingElement:
			owner := ast.WalkUpBindingElementsAndPatterns(declaration)
			if owner != nil && owner.Kind == ast.KindParameter {
				return owner, declaration.Name()
			}
		}
	}
	return nil, nil
}

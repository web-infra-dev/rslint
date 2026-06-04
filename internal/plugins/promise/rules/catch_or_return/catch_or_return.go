package catch_or_return

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/promiseutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const skipTransparent = ast.OEKParentheses

type Options struct {
	AllowThen         bool
	AllowThenStrict   bool
	AllowFinally      bool
	TerminationMethod []string
}

func parseOptions(options any) Options {
	opts := Options{TerminationMethod: []string{"catch"}}
	optsMap := utils.GetOptionsMap(options)
	if optsMap == nil {
		return opts
	}
	if v, ok := optsMap["allowThen"].(bool); ok {
		opts.AllowThen = v
	}
	if v, ok := optsMap["allowThenStrict"].(bool); ok {
		opts.AllowThenStrict = v
	}
	if v, ok := optsMap["allowFinally"].(bool); ok {
		opts.AllowFinally = v
	}
	if v, ok := optsMap["terminationMethod"].(string); ok {
		opts.TerminationMethod = []string{v}
	} else if arr, ok := optsMap["terminationMethod"].([]interface{}); ok {
		methods := make([]string, 0, len(arr))
		for _, m := range arr {
			if s, ok := m.(string); ok {
				methods = append(methods, s)
			}
		}
		if len(methods) > 0 {
			opts.TerminationMethod = methods
		}
	}
	return opts
}

func buildTerminationMessage(methods []string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "terminationMethod",
		Description: "Expected " + strings.Join(methods, ",") + "() or return",
	}
}

// isAllowedPromiseTermination mirrors eslint-plugin-promise's isAllowedPromiseTermination.
func isAllowedPromiseTermination(expression *ast.Node, opts Options) bool {
	if expression == nil || !ast.IsCallExpression(expression) {
		return false
	}
	call := expression.AsCallExpression()
	callee := ast.SkipOuterExpressions(call.Expression, skipTransparent)

	// somePromise.then(a, b) — allowThen / allowThenStrict
	// somePromise.catch().finally(fn) — allowFinally
	// somePromise.catch() — matches terminationMethod list
	if callee != nil && ast.IsPropertyAccessExpression(callee) {
		prop := callee.AsPropertyAccessExpression()
		nameNode := prop.Name()
		if nameNode != nil && ast.IsIdentifier(nameNode) {
			name := nameNode.AsIdentifier().Text

			if (opts.AllowThen || opts.AllowThenStrict) && name == "then" && len(call.Arguments.Nodes) == 2 {
				if opts.AllowThen && !opts.AllowThenStrict {
					return true
				}
				// allowThenStrict: first argument must be null literal
				firstArg := ast.SkipOuterExpressions(call.Arguments.Nodes[0], skipTransparent)
				if firstArg != nil && firstArg.Kind == ast.KindNullKeyword {
					return true
				}
			}

			// somePromise.catch().finally(fn) — allowFinally falls through to terminationMethod
			// if the recursive check fails, matching upstream's sequential-if semantics.
			if opts.AllowFinally && name == "finally" {
				receiver := ast.SkipOuterExpressions(prop.Expression, skipTransparent)
				if receiver != nil && promiseutil.IsPromiseLikeCall(receiver) && isAllowedPromiseTermination(receiver, opts) {
					return true
				}
			}

			for _, method := range opts.TerminationMethod {
				if name == method {
					return true
				}
			}
		}
	}

	// somePromise['catch']() — element access with string literal 'catch' (upstream hardcodes this)
	if callee != nil && ast.IsElementAccessExpression(callee) {
		arg := ast.SkipOuterExpressions(callee.AsElementAccessExpression().ArgumentExpression, skipTransparent)
		if arg != nil && arg.Kind == ast.KindStringLiteral && arg.AsStringLiteral().Text == "catch" {
			return true
		}
	}

	// cy.get().then(a, b) — Cypress chains
	return promiseutil.IsMemberCallWithObjectName("cy", expression)
}

var CatchOrReturnRule = rule.Rule{
	Name: "promise/catch-or-return",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)
		return rule.RuleListeners{
			ast.KindExpressionStatement: func(node *ast.Node) {
				expr := ast.SkipOuterExpressions(node.AsExpressionStatement().Expression, skipTransparent)
				if expr == nil || !promiseutil.IsPromiseLikeCall(expr) {
					return
				}
				if isAllowedPromiseTermination(expr, opts) {
					return
				}
				ctx.ReportNode(node, buildTerminationMessage(opts.TerminationMethod))
			},
		}
	},
}

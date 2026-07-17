package catch_or_return

import (
	_ "embed"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/promiseutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

//go:embed catch-or-return.schema.json
var schemaJSON []byte

const skipTransparent = ast.OEKParentheses

type Options struct {
	AllowThen         bool
	AllowThenStrict   bool
	AllowFinally      bool
	TerminationMethod []string
}

func parseOptions(options []any) Options {
	opts := Options{TerminationMethod: []string{"catch"}}
	if len(options) == 0 {
		return opts
	}
	optsMap := options[0].(map[string]interface{})
	opts.AllowThen, _ = optsMap["allowThen"].(bool)
	opts.AllowThenStrict, _ = optsMap["allowThenStrict"].(bool)
	opts.AllowFinally, _ = optsMap["allowFinally"].(bool)
	switch v := optsMap["terminationMethod"].(type) {
	case string:
		opts.TerminationMethod = []string{v}
	case []interface{}:
		if len(v) > 0 {
			methods := make([]string, len(v))
			for i, m := range v {
				methods[i] = m.(string)
			}
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

	// cy.get().then(a, b) — Cypress chains
	return promiseutil.IsMemberCallWithObjectName("cy", expression)
}

var CatchOrReturnRule = rule.Rule{
	Name:   "promise/catch-or-return",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
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

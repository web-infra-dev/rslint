package no_new_object

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var NoNewObjectRule = rule.Rule{
	Name: "no-new-object",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindNewExpression: func(node *ast.Node) {
				callee := node.Expression()
				if callee == nil {
					return
				}

				// Unwrap parentheses and TS type assertions so that
				// new (Object)(), new (Object as any)() are caught.
				unwrapped := ast.SkipOuterExpressions(callee, ast.OEKParentheses|ast.OEKAssertions)
				if unwrapped.Kind != ast.KindIdentifier || unwrapped.Text() != "Object" {
					return
				}

				// If Object is shadowed by a local declaration (var, function, class, import),
				// it's not the global built-in — skip reporting.
				if utils.IsShadowed(unwrapped, "Object") {
					return
				}

				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "preferLiteral",
					Description: "The object literal notation {} is preferable.",
				})
			},
		}
	},
}

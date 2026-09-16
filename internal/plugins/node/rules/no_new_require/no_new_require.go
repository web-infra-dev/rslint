package no_new_require

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var noNewRequireMessage = rule.RuleMessage{
	Id:          "noNewRequire",
	Description: "Unexpected use of new with require.",
}

var NoNewRequireRule = rule.Rule{
	Name:   "node/no-new-require",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindNewExpression: func(node *ast.Node) {
				callee := utils.ESTreeRuntimeExpression(node.AsNewExpression().Expression)
				if callee != nil && callee.Kind == ast.KindIdentifier && callee.AsIdentifier().Text == "require" {
					// The expression already has its exact end; reuse tsgo's trivia
					// handling without allocating a scanner for each diagnostic.
					start := scanner.GetTokenPosOfNode(node, ctx.SourceFile, false)
					ctx.ReportRange(node.Loc.WithPos(start), noNewRequireMessage)
				}
			},
		}
	},
}

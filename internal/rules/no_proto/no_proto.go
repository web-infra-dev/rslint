package no_proto

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/no-proto
var NoProtoRule = rule.Rule{
	Name: "no-proto",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		msg := rule.RuleMessage{
			Id:          "unexpectedProto",
			Description: "The '__proto__' property is deprecated.",
		}

		return rule.RuleListeners{
			ast.KindPropertyAccessExpression: func(node *ast.Node) {
				propAccess := node.AsPropertyAccessExpression()
				if propAccess.Name().Text() == "__proto__" {
					ctx.ReportNode(node, msg)
				}
			},
			ast.KindElementAccessExpression: func(node *ast.Node) {
				elemAccess := node.AsElementAccessExpression()
				if utils.GetStaticStringValue(elemAccess.ArgumentExpression) == "__proto__" {
					ctx.ReportNode(node, msg)
				}
			},
		}
	},
}

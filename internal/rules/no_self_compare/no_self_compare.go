package no_self_compare

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/no-self-compare
var NoSelfCompareRule = rule.Rule{
	Name: "no-self-compare",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				bin := node.AsBinaryExpression()
				if bin == nil || bin.OperatorToken == nil {
					return
				}
				switch bin.OperatorToken.Kind {
				case ast.KindEqualsEqualsEqualsToken,
					ast.KindEqualsEqualsToken,
					ast.KindExclamationEqualsEqualsToken,
					ast.KindExclamationEqualsToken,
					ast.KindGreaterThanToken,
					ast.KindLessThanToken,
					ast.KindGreaterThanEqualsToken,
					ast.KindLessThanEqualsToken:
				default:
					return
				}
				if utils.HasSameTokens(ctx.SourceFile, bin.Left, bin.Right) {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "comparingToSelf",
						Description: "Comparing to itself is potentially pointless.",
					})
				}
			},
		}
	},
}

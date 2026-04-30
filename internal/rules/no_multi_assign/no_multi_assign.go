package no_multi_assign

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/no-multi-assign
var NoMultiAssignRule = rule.Rule{
	Name: "no-multi-assign",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				binExpr := node.AsBinaryExpression()
				if binExpr == nil || binExpr.OperatorToken == nil {
					return
				}
				if !ast.IsAssignmentOperator(binExpr.OperatorToken.Kind) {
					return
				}

				// ESTree drops ParenthesizedExpression nodes, so the rule's
				// `VariableDeclarator > AssignmentExpression.init`-style
				// selectors look through enclosing parens. tsgo keeps parens
				// explicit, so ascend past them to find the effective parent
				// and unwrap the parent's relevant child the same way.
				parent := ast.WalkUpParenthesizedExpressions(node.Parent)
				if parent == nil {
					return
				}

				switch parent.Kind {
				case ast.KindVariableDeclaration:
					decl := parent.AsVariableDeclaration()
					if decl != nil && decl.Initializer != nil &&
						ast.SkipParentheses(decl.Initializer) == node {
						ctx.ReportNode(node, buildMessage())
					}
				case ast.KindPropertyDeclaration:
					decl := parent.AsPropertyDeclaration()
					if decl != nil && decl.Initializer != nil &&
						ast.SkipParentheses(decl.Initializer) == node {
						ctx.ReportNode(node, buildMessage())
					}
				case ast.KindBinaryExpression:
					if opts.ignoreNonDeclaration {
						return
					}
					outer := parent.AsBinaryExpression()
					if outer == nil || outer.OperatorToken == nil {
						return
					}
					if !ast.IsAssignmentOperator(outer.OperatorToken.Kind) {
						return
					}
					if outer.Right != nil && ast.SkipParentheses(outer.Right) == node {
						ctx.ReportNode(node, buildMessage())
					}
				}
			},
		}
	},
}

type noMultiAssignOptions struct {
	ignoreNonDeclaration bool
}

func parseOptions(options any) noMultiAssignOptions {
	opts := noMultiAssignOptions{ignoreNonDeclaration: false}
	optsMap := utils.GetOptionsMap(options)
	if optsMap != nil {
		if v, ok := optsMap["ignoreNonDeclaration"].(bool); ok {
			opts.ignoreNonDeclaration = v
		}
	}
	return opts
}

func buildMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpectedChain",
		Description: "Unexpected chained assignment.",
	}
}

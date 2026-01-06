package no_console

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Message builders
func buildMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpected",
		Description: "Unexpected console statement.",
	}
}

var NoConsoleRule = rule.Rule{
	Name: "no-console",
	Run:  run,
}

type Options struct {
	Allow []string
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	var ruleOptions Options

	// Parsing options
	if opts, ok := options.([]interface{}); ok {
		if len(opts) > 0 {
			if config, ok := opts[0].(map[string]interface{}); ok {
				if allow, ok := config["allow"].([]interface{}); ok {
					for _, v := range allow {
						if s, ok := v.(string); ok {
							ruleOptions.Allow = append(ruleOptions.Allow, s)
						}
					}
				}
			}
		}
	} else if config, ok := options.(map[string]interface{}); ok {
		if allow, ok := config["allow"].([]interface{}); ok {
			for _, v := range allow {
				if s, ok := v.(string); ok {
					ruleOptions.Allow = append(ruleOptions.Allow, s)
				}
			}
		}
	}

	return rule.RuleListeners{
		ast.KindPropertyAccessExpression: func(node *ast.Node) {
			checkPropertyAccess(ctx, node, ruleOptions)
		},
		ast.KindElementAccessExpression: func(node *ast.Node) {
			checkElementAccess(ctx, node, ruleOptions)
		},
	}
}

func isConsole(node *ast.Node) bool {
	return node.Kind == ast.KindIdentifier && node.AsIdentifier().Text == "console"
}

func isAllowed(name string, options Options) bool {
	for _, allowed := range options.Allow {
		if allowed == name {
			return true
		}
	}
	return false
}

func checkPropertyAccess(ctx rule.RuleContext, node *ast.Node, options Options) {
	expr := node.AsPropertyAccessExpression()
	if !isConsole(expr.Expression) {
		return
	}

	nameNode := ast.GetElementOrPropertyAccessName(node)
	if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
		return
	}

	propName := nameNode.AsIdentifier().Text
	if isAllowed(propName, options) {
		return
	}

	ctx.ReportNode(node, buildMessage())
}

func checkElementAccess(ctx rule.RuleContext, node *ast.Node, options Options) {
	expr := node.AsElementAccessExpression()
	if !isConsole(expr.Expression) {
		return
	}

	arg := expr.ArgumentExpression
	if arg.Kind != ast.KindStringLiteral {
		ctx.ReportNode(node, buildMessage())
		return
	}

	propName := arg.AsStringLiteral().Text
	if isAllowed(propName, options) {
		return
	}

	ctx.ReportNode(node, buildMessage())
}

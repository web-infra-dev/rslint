package no_dupe_args

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// https://eslint.org/docs/latest/rules/no-dupe-args
var NoDupeArgsRule = rule.Rule{
	Name: "no-dupe-args",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		checkParams := func(node *ast.Node) {
			params := node.Parameters()
			if params == nil {
				return
			}

			seen := make(map[string]bool)
			for _, param := range params {
				name := getParamName(param)
				if name == "" {
					continue
				}
				if seen[name] {
					ctx.ReportNode(param, rule.RuleMessage{
						Id:          "unexpected",
						Description: fmt.Sprintf("Duplicate param '%s'.", name),
					})
				} else {
					seen[name] = true
				}
			}
		}

		return rule.RuleListeners{
			ast.KindFunctionDeclaration: checkParams,
			ast.KindFunctionExpression:  checkParams,
		}
	},
}

// getParamName extracts the parameter name from a parameter declaration
func getParamName(param *ast.Node) string {
	if param == nil {
		return ""
	}

	paramDecl := param.AsParameterDeclaration()
	if paramDecl == nil {
		return ""
	}

	nameNode := param.Name()
	if nameNode == nil {
		return ""
	}

	if nameNode.Kind == ast.KindIdentifier {
		return nameNode.AsIdentifier().Text
	}

	return ""
}

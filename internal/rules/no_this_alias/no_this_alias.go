package no_this_alias

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/typescript-eslint/rslint/internal/rule"
)

type NoThisAliasOptions struct {
	AllowDestructuring bool     `json:"allowDestructuring"`
	AllowedNames       []string `json:"allowedNames"`
}

var NoThisAliasRule = rule.Rule{
	Name: "no-this-alias",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := NoThisAliasOptions{
			AllowDestructuring: true,
			AllowedNames:       []string{},
		}
		if options != nil {
			if optsMap, ok := options.(map[string]interface{}); ok {
				if allowDestructuring, ok := optsMap["allowDestructuring"].(bool); ok {
					opts.AllowDestructuring = allowDestructuring
				}
				if allowedNames, ok := optsMap["allowedNames"].([]interface{}); ok {
					opts.AllowedNames = make([]string, len(allowedNames))
					for i, name := range allowedNames {
						if str, ok := name.(string); ok {
							opts.AllowedNames[i] = str
						}
					}
				}
			}
		}

		checkNode := func(node *ast.Node, id *ast.Node, init *ast.Node) {
			// Check if the init/right side is a ThisExpression
			if init == nil || init.Kind != ast.KindThisKeyword {
				return
			}

			// If destructuring is allowed and the id is not an Identifier, skip
			if opts.AllowDestructuring && id.Kind != ast.KindIdentifier {
				return
			}

			// Check if the name is in the allowed list
			hasAllowedName := false
			if id.Kind == ast.KindIdentifier {
				identifier := id.AsIdentifier()
				for _, allowedName := range opts.AllowedNames {
					if identifier.Text == allowedName {
						hasAllowedName = true
						break
					}
				}
			}

			if !hasAllowedName {
				messageId := "thisAssignment"
				if id.Kind != ast.KindIdentifier {
					messageId = "thisDestructure"
				}

				ctx.ReportNode(id, rule.RuleMessage{
					Id:          messageId,
					Description: getMessageText(messageId),
				})
			}
		}

		return rule.RuleListeners{
			ast.KindVariableDeclaration: func(node *ast.Node) {
				decl := node.AsVariableDeclaration()
				checkNode(node, decl.Name(), decl.Initializer)
			},
			ast.KindBinaryExpression: func(node *ast.Node) {
				expr := node.AsBinaryExpression()
				if expr.OperatorToken.Kind == ast.KindEqualsToken {
					checkNode(node, expr.Left, expr.Right)
				}
			},
		}
	},
}

func getMessageText(messageId string) string {
	switch messageId {
	case "thisAssignment":
		return "Unexpected aliasing of 'this' to local variable."
	case "thisDestructure":
		return "Unexpected aliasing of members of 'this' to local variables."
	default:
		return ""
	}
}
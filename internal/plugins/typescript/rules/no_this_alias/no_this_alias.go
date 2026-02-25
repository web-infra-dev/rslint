package no_this_alias

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type NoThisAliasOptions struct {
	AllowDestructuring bool     `json:"allowDestructuring"`
	AllowedNames       []string `json:"allowedNames"`
}

var NoThisAliasRule = rule.CreateRule(rule.Rule{
	Name: "no-this-alias",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := NoThisAliasOptions{
			AllowDestructuring: true,
			AllowedNames:       []string{},
		}

		// Parse options with dual-format support
		if options != nil {
			var optsMap map[string]interface{}
			var ok bool

			// Handle array format: [{ option: value }]
			if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
				optsMap, ok = optArray[0].(map[string]interface{})
			} else {
				// Handle direct object format: { option: value }
				optsMap, ok = options.(map[string]interface{})
			}

			if ok {
				if allowDestructuring, ok := optsMap["allowDestructuring"].(bool); ok {
					opts.AllowDestructuring = allowDestructuring
				}
				if allowedNames, ok := optsMap["allowedNames"].([]interface{}); ok {
					for _, name := range allowedNames {
						if nameStr, ok := name.(string); ok {
							opts.AllowedNames = append(opts.AllowedNames, nameStr)
						}
					}
				}
			}
		}

		// Helper to check if initializer is `this`
		isThisKeyword := func(node *ast.Node) bool {
			return node != nil && node.Kind == ast.KindThisKeyword
		}

		return rule.RuleListeners{
			ast.KindVariableDeclaration: func(node *ast.Node) {
				varDecl := node.AsVariableDeclaration()
				if varDecl == nil || varDecl.Initializer == nil {
					return
				}

				// Check if initializer is `this`
				if !isThisKeyword(varDecl.Initializer) {
					return
				}

				// Get the name node
				nameNode := varDecl.Name()
				if nameNode == nil {
					return
				}

				// Check if it's destructuring
				switch nameNode.Kind {
				case ast.KindObjectBindingPattern, ast.KindArrayBindingPattern:
					// Destructuring pattern
					if !opts.AllowDestructuring {
						ctx.ReportNode(nameNode, rule.RuleMessage{
							Id:          "thisDestructure",
							Description: "Destructuring `this` is not allowed.",
						})
					}
					return
				case ast.KindIdentifier:
					// Check if the identifier name is in allowedNames
					id := nameNode.AsIdentifier()
					if id != nil {
						idName := id.Text
						for _, allowedName := range opts.AllowedNames {
							if idName == allowedName {
								// Name is allowed, don't report
								return
							}
						}
					}
					// Regular identifier assignment - report it
					ctx.ReportNode(nameNode, rule.RuleMessage{
						Id:          "thisAssignment",
						Description: "Unexpected aliasing of `this` to local variable.",
					})
					return
				}
			},
			ast.KindBinaryExpression: func(node *ast.Node) {
				binExpr := node.AsBinaryExpression()
				if binExpr == nil {
					return
				}

				// Check for assignment (=) operator
				if binExpr.OperatorToken == nil || binExpr.OperatorToken.Kind != ast.KindEqualsToken {
					return
				}

				// Check if right side is `this`
				if !isThisKeyword(binExpr.Right) {
					return
				}

				// Check left side
				if binExpr.Left != nil {
					switch binExpr.Left.Kind {
					case ast.KindObjectLiteralExpression, ast.KindArrayLiteralExpression:
						// Destructuring pattern
						if !opts.AllowDestructuring {
							ctx.ReportNode(binExpr.Left, rule.RuleMessage{
								Id:          "thisDestructure",
								Description: "Destructuring `this` is not allowed.",
							})
						}
						return
					case ast.KindIdentifier:
						// Check if the identifier name is in allowedNames
						id := binExpr.Left.AsIdentifier()
						if id != nil {
							idName := id.Text
							for _, allowedName := range opts.AllowedNames {
								if idName == allowedName {
									// Name is allowed, don't report
									return
								}
							}
						}
						// Regular identifier assignment
						ctx.ReportNode(binExpr.Left, rule.RuleMessage{
							Id:          "thisAssignment",
							Description: "Unexpected aliasing of `this` to local variable.",
						})
						return
					}
				}
			},
		}
	},
})

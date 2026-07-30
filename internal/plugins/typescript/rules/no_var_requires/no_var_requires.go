package no_var_requires

import (
	_ "embed"
	"regexp"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

//go:embed no_var_requires.schema.json
var schemaJSON []byte

type Options struct {
	Allow []string `json:"allow"`
}

func parseOptions(options []any) *Options {
	opts := &Options{}
	if len(options) == 0 {
		return opts
	}
	if typed, ok := options[0].(*Options); ok {
		return typed
	}
	optMap, _ := options[0].(map[string]interface{})
	if allowList, exists := optMap["allow"]; exists {
		if allowSlice, ok := allowList.([]interface{}); ok {
			for _, item := range allowSlice {
				if str, ok := item.(string); ok {
					opts.Allow = append(opts.Allow, str)
				}
			}
		} else if allowStrSlice, ok := allowList.([]string); ok {
			opts.Allow = allowStrSlice
		}
	}
	return opts
}

var NoVarRequiresRule = rule.CreateRule(rule.Rule{
	Name:   "no-var-requires",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)

		// Compile allow patterns into regexes
		var allowPatterns []*regexp.Regexp
		for _, pattern := range opts.Allow {
			if re, err := regexp.Compile(pattern); err == nil {
				allowPatterns = append(allowPatterns, re)
			}
		}

		isImportPathAllowed := func(importPath string) bool {
			for _, pattern := range allowPatterns {
				if pattern.MatchString(importPath) {
					return true
				}
			}
			return false
		}

		isStringOrTemplateLiteral := func(node *ast.Node) bool {
			if node == nil {
				return false
			}
			if node.Kind == ast.KindStringLiteral {
				return true
			}
			if node.Kind == ast.KindTemplateExpression || node.Kind == ast.KindNoSubstitutionTemplateLiteral {
				return true
			}
			return false
		}

		// Helper to check if a node or any of its ancestors is a variable declaration
		isInVariableDeclaration := func(node *ast.Node) bool {
			current := node
			for current != nil {
				if current.Kind == ast.KindVariableDeclaration ||
					current.Kind == ast.KindVariableDeclarationList {
					return true
				}
				current = current.Parent
			}
			return false
		}

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				callExpr := node.AsCallExpression()
				if callExpr == nil {
					return
				}

				// Check if this is a require() call
				callee := callExpr.Expression
				if callee == nil || callee.Kind != ast.KindIdentifier {
					return
				}

				identifier := callee.AsIdentifier()
				if identifier == nil || identifier.Text != "require" {
					return
				}

				// Check if require is a local variable (not the global require)
				symbol := identifier.Symbol()
				if symbol != nil && len(symbol.Declarations) > 0 {
					// This is a local require variable, not the global one
					return
				}

				// Check arguments for allow patterns
				args := callExpr.Arguments
				if args != nil && len(args.Nodes) > 0 && isStringOrTemplateLiteral(args.Nodes[0]) {
					// Get string value from argument
					arg := args.Nodes[0]
					var argValue string
					if arg.Kind == ast.KindStringLiteral {
						stringLiteral := arg.AsStringLiteral()
						if stringLiteral != nil {
							// Remove quotes from the text
							text := stringLiteral.Text
							if len(text) >= 2 && ((text[0] == '"' && text[len(text)-1] == '"') ||
								(text[0] == '\'' && text[len(text)-1] == '\'')) {
								argValue = text[1 : len(text)-1]
							} else {
								argValue = text
							}
						}
					}
					if argValue != "" && isImportPathAllowed(argValue) {
						return
					}
				}

				// Get the parent, handling optional chaining
				parent := node.Parent
				if parent != nil && parent.Kind == ast.KindPropertyAccessExpression {
					// This handles optional chaining like require?.('foo')
					parent = parent.Parent
				}

				if parent == nil {
					return
				}

				// Check if this is part of a TypeScript import statement
				// import foo = require('foo') is allowed
				if parent.Kind == ast.KindImportEqualsDeclaration {
					return
				}

				// Standalone require() calls are allowed
				if parent.Kind == ast.KindExpressionStatement {
					return
				}

				// Check if require is used in contexts that are not allowed
				if isInVariableDeclaration(node) ||
					parent.Kind == ast.KindCallExpression ||
					parent.Kind == ast.KindPropertyAccessExpression ||
					parent.Kind == ast.KindNewExpression ||
					parent.Kind == ast.KindAsExpression ||
					parent.Kind == ast.KindTypeAssertionExpression {
					ctx.ReportNode(node, rule.RuleMessage{
						Description: "Require statement not part of import statement.",
						Id:          "noVarReqs",
					})
				}
			},
		}
	},
})

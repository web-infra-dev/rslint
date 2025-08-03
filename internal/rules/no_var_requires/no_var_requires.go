package no_var_requires

import (
	"regexp"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type Options struct {
	Allow []string `json:"allow"`
}

var NoVarRequiresRule = rule.Rule{
	Name: "no-var-requires",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := &Options{}
		if options != nil {
			if o, ok := options.(*Options); ok {
				opts = o
			}
		}

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

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				callExpr := node

				// Check if this is a require() call
				callee := callExpr.AsCallExpression().Expression
				if callee.Kind != ast.KindIdentifier {
					return
				}

				identifier := callee.AsIdentifier()
				if identifier.Text != "require" {
					return
				}

				// Check if require is a local variable (not the global require)
				// In RSLint, we check if the identifier has a symbol and if it's not the global require
				symbol := identifier.Symbol()
				if symbol != nil {
					// If require has a symbol and it's not the global require, it's a local variable
					// We need to check if the symbol is from a declaration (local variable)
					if len(symbol.Declarations) > 0 {
						// This is a local require variable, not the global one
						return
					}
				}

				// Check arguments for allow patterns
				args := callExpr.AsCallExpression().Arguments
				if len(args.Nodes) > 0 && isStringOrTemplateLiteral(args.Nodes[0]) {
					// Get string value from argument
					arg := args.Nodes[0]
					var argValue string
					if arg.Kind == ast.KindStringLiteral {
						argValue = arg.AsStringLiteral().Text
					}
					if argValue != "" && isImportPathAllowed(argValue) {
						return
					}
				}

				// Get the parent, handling ChainExpression
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
				invalidParentKinds := []ast.Kind{
					ast.KindVariableDeclaration,
					ast.KindCallExpression,
					ast.KindPropertyAccessExpression,
					ast.KindNewExpression,
					ast.KindAsExpression,
					ast.KindTypeAssertionExpression,
				}

				for _, kind := range invalidParentKinds {
					if parent.Kind == kind {
						ctx.ReportNode(node, rule.RuleMessage{
							Description: "Require statement not part of import statement.",
							Id:          "noVarReqs",
						})
						return
					}
				}

				// For variable declarations, we need to check the parent's parent
				if parent.Kind == ast.KindVariableDeclaration {
					ctx.ReportNode(node, rule.RuleMessage{
						Description: "Require statement not part of import statement.",
						Id:          "noVarReqs",
					})
				}
			},
		}
	},
}

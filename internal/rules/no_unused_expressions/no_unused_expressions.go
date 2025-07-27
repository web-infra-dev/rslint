package no_unused_expressions

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/typescript-eslint/rslint/internal/rule"
)

type NoUnusedExpressionsOptions struct {
	AllowShortCircuit     bool `json:"allowShortCircuit"`
	AllowTernary          bool `json:"allowTernary"`
	AllowTaggedTemplates  bool `json:"allowTaggedTemplates"`
}

func buildUnusedExpressionMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unusedExpression",
		Description: "Expected an assignment or function call and instead saw an expression.",
	}
}

var NoUnusedExpressionsRule = rule.Rule{
	Name: "no-unused-expressions",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := NoUnusedExpressionsOptions{
			AllowShortCircuit:    false,
			AllowTernary:         false,
			AllowTaggedTemplates: false,
		}
		
		if options != nil {
			if optsMap, ok := options.(map[string]interface{}); ok {
				if allowShortCircuit, ok := optsMap["allowShortCircuit"].(bool); ok {
					opts.AllowShortCircuit = allowShortCircuit
				}
				if allowTernary, ok := optsMap["allowTernary"].(bool); ok {
					opts.AllowTernary = allowTernary
				}
				if allowTaggedTemplates, ok := optsMap["allowTaggedTemplates"].(bool); ok {
					opts.AllowTaggedTemplates = allowTaggedTemplates
				}
			}
		}

		var isValidExpression func(node *ast.Node) bool
		isValidExpression = func(node *ast.Node) bool {
			// Binary expressions with side effects (short circuit)
			if node.Kind == ast.KindBinaryExpression {
				binaryExpr := node.AsBinaryExpression()
				// For logical operators (&&, ||), check if right side has side effects
				if binaryExpr.OperatorToken.Kind == ast.KindAmpersandAmpersandToken || 
				   binaryExpr.OperatorToken.Kind == ast.KindBarBarToken {
					// Allow if allowShortCircuit is true, or if right side has side effects
					if opts.AllowShortCircuit {
						return isValidExpression(binaryExpr.Right)
					}
					// Even without allowShortCircuit, allow if right side has side effects
					return isValidExpression(binaryExpr.Right)
				}
				// Other binary expressions (like arithmetic) are not valid
				return false
			}
			
			// Conditional expressions (ternary)
			if node.Kind == ast.KindConditionalExpression {
				conditionalExpr := node.AsConditionalExpression()
				// Allow if both branches have side effects, or if allowTernary is true
				if opts.AllowTernary {
					return isValidExpression(conditionalExpr.WhenTrue) && 
						   isValidExpression(conditionalExpr.WhenFalse)
				}
				// Even without allowTernary, allow if both branches have side effects
				return isValidExpression(conditionalExpr.WhenTrue) && 
					   isValidExpression(conditionalExpr.WhenFalse)
			}
			
			// ChainExpression with CallExpression (e.g., foo?.())
			if node.Kind == ast.KindCallExpression {
				callExpr := node.AsCallExpression()
				if callExpr.QuestionDotToken != nil {
					return true
				}
			}
			
			// ImportExpression (e.g., import('./foo'))
			if node.Kind == ast.KindImportKeyword {
				return true
			}
			
			// Check for call expressions within chain expressions
			if node.Kind == ast.KindPropertyAccessExpression || node.Kind == ast.KindElementAccessExpression {
				// Check if this is part of an optional chain that ends in a call
				parent := node.Parent
				for parent != nil {
					if parent.Kind == ast.KindCallExpression {
						return true
					}
					if parent.Kind != ast.KindPropertyAccessExpression && 
					   parent.Kind != ast.KindElementAccessExpression {
						break
					}
					parent = parent.Parent
				}
			}

			// Tagged template expressions
			if opts.AllowTaggedTemplates && node.Kind == ast.KindTaggedTemplateExpression {
				return true
			}
			
			// Only allow expressions with side effects
			return node.Kind == ast.KindCallExpression || 
			       node.Kind == ast.KindNewExpression ||
			       node.Kind == ast.KindPostfixUnaryExpression ||
			       node.Kind == ast.KindDeleteExpression ||
			       node.Kind == ast.KindAwaitExpression ||
			       node.Kind == ast.KindYieldExpression
		}

		return rule.RuleListeners{
			ast.KindExpressionStatement: func(node *ast.Node) {
				exprStmt := node.AsExpressionStatement()
				
				// Skip directive prologues (e.g., 'use strict')
				if ast.IsPrologueDirective(node) {
					// Get the string literal text to check if it's a known directive
					exprStmt := node.AsExpressionStatement()
					stringLiteral := exprStmt.Expression.AsStringLiteral()
					literalText := stringLiteral.Text
					
					// Debug: print the literal text to understand what we're getting
					// fmt.Printf("DEBUG: String literal text: '%s'\n", literalText)
					
					// Check if it's a known directive value
					// Note: The text includes quotes, so "use strict" becomes '"use strict"'
					if literalText == "use strict" || literalText == "use asm" {
						// Check if this is a directive by looking at its position
						parent := node.Parent
						if parent != nil && (parent.Kind == ast.KindSourceFile || 
						                    parent.Kind == ast.KindBlock || 
						                    parent.Kind == ast.KindModuleBlock) {
							// Check if it's at the beginning of the block
							var statements []*ast.Node
							switch parent.Kind {
							case ast.KindSourceFile:
								statements = parent.AsSourceFile().Statements.Nodes
							case ast.KindBlock:
								statements = parent.AsBlock().Statements.Nodes
							case ast.KindModuleBlock:
								statements = parent.AsModuleBlock().Statements.Nodes
							}
							
							// Check if this is in the directive prologue position
							// All statements from the start until the first non-string-literal statement form the directive prologue
							for _, stmt := range statements {
								if stmt == node {
									return // Found our node within the prologue, allow it
								}
								if !ast.IsPrologueDirective(stmt) {
									break // Hit non-directive, prologue ended
								}
							}
						}
					}
				}
				
				expression := exprStmt.Expression
				
				// Handle TypeScript-specific nodes by unwrapping them
				switch expression.Kind {
				case ast.KindAsExpression:
					expression = expression.AsAsExpression().Expression
				case ast.KindTypeAssertionExpression:
					expression = expression.AsTypeAssertion().Expression
				case ast.KindNonNullExpression:
					expression = expression.AsNonNullExpression().Expression
				case ast.KindSatisfiesExpression:
					expression = expression.AsSatisfiesExpression().Expression
				}
				
				// Check for instantiation expressions (e.g., Foo<string>)
				if expression.Kind == ast.KindExpressionWithTypeArguments {
					ctx.ReportNode(node, buildUnusedExpressionMessage())
					return
				}
				
				if isValidExpression(expression) {
					return
				}
				
				ctx.ReportNode(node, buildUnusedExpressionMessage())
			},
		}
	},
}
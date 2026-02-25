package prefer_as_const

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func buildPreferConstAssertionMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferConstAssertion",
		Description: "Expected a `const` instead of a literal type assertion.",
	}
}

func buildVariableConstAssertionMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "variableConstAssertion",
		Description: "Expected a `const` assertion instead of a literal type annotation.",
	}
}

func buildVariableSuggestMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "variableSuggest",
		Description: "You should use `as const` instead of type annotation.",
	}
}

var PreferAsConstRule = rule.CreateRule(rule.Rule{
	Name: "prefer-as-const",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {

		compareTypes := func(valueNode *ast.Node, typeNode *ast.Node, canFix bool) {
			if valueNode == nil || typeNode == nil {
				return
			}

			// Check if valueNode is a literal and typeNode is a literal type
			if !ast.IsLiteralExpression(valueNode) {
				return
			}

			var isLiteralType bool
			var literalNode *ast.Node

			if ast.IsLiteralTypeNode(typeNode) {
				literalTypeNode := typeNode.AsLiteralTypeNode()
				if literalTypeNode == nil {
					return
				}
				literalNode = literalTypeNode.Literal
				isLiteralType = true
			} else {
				return
			}

			if !isLiteralType || literalNode == nil {
				return
			}

			// Check if both are literals and have the same raw value
			if !ast.IsLiteralExpression(literalNode) {
				return
			}

			// Skip template literal types - they are different from regular literal types
			if literalNode.Kind == ast.KindNoSubstitutionTemplateLiteral {
				return
			}

			valueRange := utils.TrimNodeTextRange(ctx.SourceFile, valueNode)
			valueText := ctx.SourceFile.Text()[valueRange.Pos():valueRange.End()]
			typeRange := utils.TrimNodeTextRange(ctx.SourceFile, literalNode)
			typeText := ctx.SourceFile.Text()[typeRange.Pos():typeRange.End()]

			if valueText == typeText {
				if canFix {
					// For type assertions, we can directly fix to 'const'
					ctx.ReportNodeWithFixes(literalNode, buildPreferConstAssertionMessage(),
						rule.RuleFixReplace(ctx.SourceFile, typeNode, "const"))
				} else {
					// For variable declarations, suggest replacing with 'as const'
					// We need to find the colon token before the type annotation
					// and remove from there to the end of the type annotation
					parent := typeNode.Parent
					if parent != nil {
						// Find the colon token between the variable name and type
						s := scanner.GetScannerForSourceFile(ctx.SourceFile, parent.Pos())
						colonStart := -1
						for s.TokenStart() < typeNode.Pos() {
							if s.Token() == ast.KindColonToken {
								colonStart = s.TokenStart()
							}
							s.Scan()
						}

						if colonStart != -1 {
							ctx.ReportNodeWithSuggestions(literalNode, buildVariableConstAssertionMessage(),
								rule.RuleSuggestion{
									Message: buildVariableSuggestMessage(),
									FixesArr: []rule.RuleFix{
										rule.RuleFixReplaceRange(
											core.NewTextRange(colonStart, typeNode.End()),
											"",
										),
										rule.RuleFixInsertAfter(valueNode, " as const"),
									},
								})
						}
					}
				}
			}
		}

		return rule.RuleListeners{
			// PropertyDefinition in TypeScript corresponds to PropertyDeclaration in Go AST
			ast.KindPropertyDeclaration: func(node *ast.Node) {
				if node.Kind != ast.KindPropertyDeclaration {
					return
				}
				propDecl := node.AsPropertyDeclaration()
				if propDecl == nil {
					return
				}
				if propDecl.Initializer != nil && propDecl.Type != nil {
					compareTypes(propDecl.Initializer, propDecl.Type, false)
				}
			},

			ast.KindAsExpression: func(node *ast.Node) {
				if node.Kind != ast.KindAsExpression {
					return
				}
				asExpr := node.AsAsExpression()
				if asExpr == nil {
					return
				}
				compareTypes(asExpr.Expression, asExpr.Type, true)
			},

			ast.KindTypeAssertionExpression: func(node *ast.Node) {
				if node.Kind != ast.KindTypeAssertionExpression {
					return
				}
				typeAssertion := node.AsTypeAssertion()
				if typeAssertion == nil {
					return
				}
				compareTypes(typeAssertion.Expression, typeAssertion.Type, true)
			},

			// VariableDeclarator in TypeScript corresponds to VariableDeclaration in Go AST
			ast.KindVariableDeclaration: func(node *ast.Node) {
				if node.Kind != ast.KindVariableDeclaration {
					return
				}
				varDecl := node.AsVariableDeclaration()
				if varDecl == nil {
					return
				}
				if varDecl.Initializer != nil && varDecl.Type != nil {
					compareTypes(varDecl.Initializer, varDecl.Type, false)
				}
			},
		}
	},
})

package no_duplicate_enum_values

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

var NoDuplicateEnumValuesRule = rule.CreateRule(rule.Rule{
	Name: "no-duplicate-enum-values",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindEnumDeclaration: func(node *ast.Node) {
				enumDecl := node.AsEnumDeclaration()
				if enumDecl == nil || enumDecl.Members == nil {
					return
				}

				// Track seen values
				seenValues := make(map[string]*ast.Node)

				for _, memberNode := range enumDecl.Members.Nodes {
					member := memberNode.AsEnumMember()
					if member == nil || member.Initializer == nil {
						continue
					}

					// Get the initializer value
					initializer := member.Initializer
					var valueStr string
					var isLiteral bool

					switch initializer.Kind {
					case ast.KindNumericLiteral:
						numLit := initializer.AsNumericLiteral()
						if numLit != nil {
							valueStr = numLit.Text
							isLiteral = true
						}
					case ast.KindStringLiteral:
						strLit := initializer.AsStringLiteral()
						if strLit != nil {
							valueStr = strLit.Text
							isLiteral = true
						}
					case ast.KindNoSubstitutionTemplateLiteral:
						tmplLit := initializer.AsNoSubstitutionTemplateLiteral()
						if tmplLit != nil {
							valueStr = tmplLit.Text
							isLiteral = true
						}
					case ast.KindPrefixUnaryExpression:
						// Handle negative numbers
						unaryExpr := initializer.AsPrefixUnaryExpression()
						if unaryExpr != nil && unaryExpr.Operator == ast.KindMinusToken {
							if numLit := unaryExpr.Operand.AsNumericLiteral(); numLit != nil {
								valueStr = "-" + numLit.Text
								isLiteral = true
							}
						}
					}

					if !isLiteral {
						continue
					}

					// Check for duplicate
					if prevNode, exists := seenValues[valueStr]; exists {
						// Get the value for display
						displayValue := valueStr

						// For string literals and template literals, display the actual string content
						switch initializer.Kind {
						case ast.KindStringLiteral:
							strLit := initializer.AsStringLiteral()
							if strLit != nil {
								displayValue = strLit.Text
							}
						case ast.KindNoSubstitutionTemplateLiteral:
							tmplLit := initializer.AsNoSubstitutionTemplateLiteral()
							if tmplLit != nil {
								// Remove backticks for display
								displayValue = tmplLit.Text
								if len(displayValue) >= 2 && displayValue[0] == '`' && displayValue[len(displayValue)-1] == '`' {
									displayValue = displayValue[1 : len(displayValue)-1]
								}
							}
						}

						ctx.ReportNode(member.Name(), rule.RuleMessage{
							Id:          "duplicateValue",
							Description: fmt.Sprintf("Duplicate enum member value %s.", displayValue),
						})

						// Report the previous occurrence as well if needed
						if prevMember := prevNode.AsEnumMember(); prevMember != nil {
							// Only report once per duplicate
							continue
						}
					} else {
						seenValues[valueStr] = memberNode
					}
				}
			},
		}
	},
})

package no_duplicate_enum_values

import (
	"fmt"
	"strconv"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

var NoDuplicateEnumValuesRule = rule.Rule{
	Name: "no-duplicate-enum-values",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindEnumDeclaration: func(node *ast.Node) {
				enumDecl := node.AsEnumDeclaration()
				seenValues := make(map[interface{}]bool)

				for _, member := range enumDecl.Members.Nodes {
					enumMember := member.AsEnumMember()
					
					// Skip members without initializers
					if enumMember.Initializer == nil {
						continue
					}

					var value interface{}
					switch enumMember.Initializer.Kind {
					case ast.KindStringLiteral:
						// String literal - extract value without quotes
						stringLiteral := enumMember.Initializer.AsStringLiteral()
						text := stringLiteral.Text
						// Remove quotes to get the actual string value
						if len(text) >= 2 && ((text[0] == '"' && text[len(text)-1] == '"') || (text[0] == '\'' && text[len(text)-1] == '\'')) {
							value = text[1 : len(text)-1]
						} else {
							value = text
						}
					case ast.KindNumericLiteral:
						// Numeric literal - parse as number for proper comparison
						numericLiteral := enumMember.Initializer.AsNumericLiteral()
						text := numericLiteral.Text
						// Try to parse as float64 for proper numeric comparison
						if num, err := strconv.ParseFloat(text, 64); err == nil {
							value = num
						} else {
							// Fallback to text representation if parsing fails
							value = text
						}
					case ast.KindNoSubstitutionTemplateLiteral:
						// No substitution template literal (e.g., `A`)
						templateLiteral := enumMember.Initializer.AsNoSubstitutionTemplateLiteral()
						text := templateLiteral.Text
						// Remove backticks to get the actual template value
						if len(text) >= 2 && text[0] == '`' && text[len(text)-1] == '`' {
							value = text[1 : len(text)-1]
						} else {
							value = text
						}
					case ast.KindTemplateExpression:
						// Template literal - only handle static templates (no expressions)
						templateExpr := enumMember.Initializer.AsTemplateExpression()
						if templateExpr.TemplateSpans == nil || len(templateExpr.TemplateSpans.Nodes) == 0 {
							// Static template literal with no expressions
							value = templateExpr.Head.Text
						} else {
							// Skip template literals with expressions
							continue
						}
					default:
						// Skip other expression types (function calls, identifiers, etc.)
						continue
					}

					// Check if we've seen this value before
					if seenValues[value] {
						ctx.ReportNode(member, rule.RuleMessage{
							Id:          "duplicateValue",
							Description: fmt.Sprintf("Duplicate enum member value %v.", value),
						})
					} else {
						seenValues[value] = true
					}
				}
			},
		}
	},
}
package no_dupe_keys

import (
	"fmt"
	"strconv"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// https://eslint.org/docs/latest/rules/no-dupe-keys
var NoDupeKeysRule = rule.Rule{
	Name: "no-dupe-keys",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindObjectLiteralExpression: func(node *ast.Node) {
				objLit := node.AsObjectLiteralExpression()
				if objLit == nil || objLit.Properties == nil {
					return
				}

				type propInfo struct {
					get  bool
					set  bool
					init bool
				}
				properties := make(map[string]*propInfo)

				for _, prop := range objLit.Properties.Nodes {
					// Skip spread elements
					if prop.Kind == ast.KindSpreadAssignment {
						continue
					}

					nameNode := prop.Name()
					if nameNode == nil {
						continue
					}

					name, isStatic := getStaticPropertyName(nameNode)
					if !isStatic {
						continue
					}

					// Skip __proto__ setters (non-computed, non-shorthand, non-method init)
					if name == "__proto__" && prop.Kind == ast.KindPropertyAssignment {
						if nameNode.Kind != ast.KindComputedPropertyName {
							continue
						}
					}

					info, exists := properties[name]
					if !exists {
						info = &propInfo{}
						properties[name] = info
					}

					switch prop.Kind {
					case ast.KindGetAccessor:
						if info.get || info.init {
							ctx.ReportNode(prop, rule.RuleMessage{
								Id:          "unexpected",
								Description: fmt.Sprintf("Duplicate key '%s'.", name),
							})
						}
						info.get = true
					case ast.KindSetAccessor:
						if info.set || info.init {
							ctx.ReportNode(prop, rule.RuleMessage{
								Id:          "unexpected",
								Description: fmt.Sprintf("Duplicate key '%s'.", name),
							})
						}
						info.set = true
					default:
						// init (PropertyAssignment, ShorthandPropertyAssignment, MethodDeclaration)
						if info.init || info.get || info.set {
							ctx.ReportNode(prop, rule.RuleMessage{
								Id:          "unexpected",
								Description: fmt.Sprintf("Duplicate key '%s'.", name),
							})
						}
						info.init = true
					}
				}
			},
		}
	},
}

// normalizeNumericLiteral parses a numeric literal text and returns its
// normalized string representation, matching ESLint's String(node.value) behavior.
// e.g., "0x1" -> "1", "1.0" -> "1", "1e2" -> "100"
func normalizeNumericLiteral(text string) string {
	f, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return text
	}
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// getStaticPropertyName extracts the static name from a property name node.
// Returns the name and whether it's a static (non-computed) name.
func getStaticPropertyName(nameNode *ast.Node) (string, bool) {
	switch nameNode.Kind {
	case ast.KindIdentifier:
		return nameNode.AsIdentifier().Text, true
	case ast.KindStringLiteral:
		return nameNode.AsStringLiteral().Text, true
	case ast.KindNumericLiteral:
		return normalizeNumericLiteral(nameNode.AsNumericLiteral().Text), true
	case ast.KindComputedPropertyName:
		expr := nameNode.AsComputedPropertyName().Expression
		if expr.Kind == ast.KindStringLiteral {
			return expr.AsStringLiteral().Text, true
		}
		if expr.Kind == ast.KindNumericLiteral {
			return normalizeNumericLiteral(expr.AsNumericLiteral().Text), true
		}
		return "", false
	default:
		return "", false
	}
}

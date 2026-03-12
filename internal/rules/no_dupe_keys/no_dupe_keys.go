package no_dupe_keys

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
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

					name, isStatic := utils.GetStaticPropertyName(nameNode)
					if !isStatic {
						continue
					}

					// Skip __proto__ setters: non-computed __proto__ in PropertyAssignment is
					// a prototype setter, not a regular property. Computed ["__proto__"] is regular.
					if name == "__proto__" && prop.Kind == ast.KindPropertyAssignment && nameNode.Kind != ast.KindComputedPropertyName {
						continue
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

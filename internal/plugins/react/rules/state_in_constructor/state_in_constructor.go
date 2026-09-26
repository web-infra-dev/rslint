package state_in_constructor

import (
	_ "embed"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed state_in_constructor.schema.json
var schemaJSON []byte

var StateInConstructorRule = rule.Rule{
	Name:   "react/state-in-constructor",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		never := len(options) > 0 && options[0] == "never"
		pragma := reactutil.GetReactPragmaFromContext(ctx)
		inComponent := func(node *ast.Node) bool {
			component := reactutil.EnclosingClass(node)
			if component == nil {
				return false
			}
			return reactutil.ExtendsReactComponent(component, pragma) || reactutil.IsExplicitReactComponent(component)
		}

		if !never {
			return rule.RuleListeners{
				ast.KindPropertyDeclaration: func(node *ast.Node) {
					if !utils.IsPlainClassMember(node) || ast.HasStaticModifier(node) {
						return
					}
					name := node.Name()
					if name != nil && name.Kind == ast.KindComputedPropertyName {
						name = utils.ESTreeRuntimeExpression(name.AsComputedPropertyName().Expression)
					}
					// Upstream reads key.name, including private and computed
					// identifiers, but not string literals such as ['state'].
					if reactutil.IdentifierOrPrivateName(name) == "state" && inComponent(node) {
						ctx.ReportNode(node, rule.RuleMessage{
							Id:          "stateInitConstructor",
							Description: "State initialization should be in a constructor",
						})
					}
				},
			}
		}

		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				assignment := node.AsBinaryExpression()
				if !ast.IsAssignmentOperator(assignment.OperatorToken.Kind) {
					return
				}
				left := utils.ESTreeRuntimeExpression(assignment.Left)
				if ast.IsOptionalChain(left) {
					return
				}
				object, property := utils.MemberExpressionParts(left)
				object = utils.ESTreeRuntimeExpression(object)
				property = utils.ESTreeRuntimeExpression(property)
				if object == nil || object.Kind != ast.KindThisKeyword || reactutil.IdentifierOrPrivateName(property) != "state" {
					return
				}
				if inConstructor(node) && inComponent(node) {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "stateInitClassProp",
						Description: "State initialization should be in a class property",
					})
				}
			},
		}
	},
}

func inConstructor(node *ast.Node) bool {
	// Upstream searches all enclosing scopes, including nested functions
	// and classes, rather than stopping at the nearest function boundary.
	return ast.FindAncestor(node.Parent, func(parent *ast.Node) bool {
		return ast.IsConstructorDeclaration(parent) && !ast.HasStaticModifier(parent)
	}) != nil
}

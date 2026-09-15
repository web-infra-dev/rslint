package spec_only

import (
	_ "embed"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/promiseutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed spec_only.schema.json
var schemaJSON []byte

// identifierName mirrors ESTree's property.name, including private names.
func identifierName(property *ast.Node) string {
	switch property.Kind {
	case ast.KindIdentifier:
		return property.AsIdentifier().Text
	case ast.KindPrivateIdentifier:
		return strings.TrimPrefix(property.AsPrivateIdentifier().Text, "#")
	default:
		return ""
	}
}

func propertyName(property *ast.Node) string {
	if ast.IsMemberName(property) {
		return identifierName(property)
	}
	// ESTree templates have neither .name nor .value, even without substitutions.
	if property.Kind != ast.KindNoSubstitutionTemplateLiteral {
		if value, ok := utils.GetStaticExpressionValue(property); ok {
			return value
		}
	}
	return "undefined"
}

func isPermittedProperty(node, property *ast.Node, instance bool, allowedMethods map[string]bool) bool {
	var name string
	switch property.Kind {
	case ast.KindStringLiteral:
		name = property.AsStringLiteral().Text
	case ast.KindIdentifier:
		if node.Kind == ast.KindElementAccessExpression {
			return true
		}
		name = property.AsIdentifier().Text
	default:
		// Literal values are compared without coercion upstream, so numbers,
		// booleans, null, bigints and regexps cannot match an allowed string.
		return false
	}
	if allowedMethods[name] {
		return true
	}
	if instance {
		return name == "then" || name == "catch" || name == "finally"
	}
	return promiseutil.IsPromiseStatic(name) || name == "withResolvers"
}

// https://github.com/eslint-community/eslint-plugin-promise/blob/v7.3.0/rules/spec-only.js
var SpecOnlyRule = rule.Rule{
	Name:   "promise/spec-only",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		var allowedMethods map[string]bool
		if len(options) > 0 {
			if opts, ok := options[0].(map[string]any); ok {
				if methods, ok := opts["allowedMethods"].([]any); ok {
					allowedMethods = make(map[string]bool, len(methods))
					for _, method := range methods {
						if name, ok := method.(string); ok {
							allowedMethods[name] = true
						}
					}
				}
			}
		}

		checkMember := func(node *ast.Node) {
			object, property := utils.MemberExpressionParts(node)
			object = utils.ESTreeRuntimeExpression(object)
			if object == nil || !ast.IsIdentifier(object) || object.AsIdentifier().Text != "Promise" {
				return
			}
			// Only a Promise receiver needs the parent walk that excludes JSX tags.
			if utils.IsInJsxTagName(node) {
				return
			}
			property = utils.ESTreeRuntimeExpression(property)
			if property == nil {
				return
			}

			if identifierName(property) == "prototype" {
				parent := utils.ESTreeParent(node)
				parentObject, parentProperty := utils.MemberExpressionParts(parent)
				if parentProperty == nil {
					return
				}
				// A terminated optional chain has an ESTree ChainExpression
				// parent, even when tsgo exposes a member beyond parentheses.
				if ast.IsOptionalChain(node) && (node.Parent != parent || !ast.IsOptionalChain(parent) || parentObject != node) {
					return
				}
				if isPermittedProperty(parent, utils.ESTreeRuntimeExpression(parentProperty), true, allowedMethods) {
					return
				}
			} else if isPermittedProperty(node, property, false, allowedMethods) {
				return
			}

			name := propertyName(property)
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "avoidNonStandard",
				Description: "Avoid using non-standard 'Promise." + name + "'",
				Data:        map[string]string{"name": name},
			})
		}

		return rule.RuleListeners{
			ast.KindPropertyAccessExpression: checkMember,
			ast.KindElementAccessExpression:  checkMember,
			ast.KindQualifiedName:            checkMember,
		}
	},
}

package no_object_as_default_parameter

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const (
	messageIDIdentifier    = "identifier"
	messageIDNonIdentifier = "non-identifier"
)

func identifierMessage(parameter string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          messageIDIdentifier,
		Description: "Do not use an object literal as default for parameter `" + parameter + "`.",
		Data:        map[string]string{"parameter": parameter},
	}
}

var nonIdentifierMessage = rule.RuleMessage{
	Id:          messageIDNonIdentifier,
	Description: "Do not use an object literal as default.",
}

// NoObjectAsDefaultParameterRule disallows non-empty object literals as
// function-parameter defaults.
//
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/rules/no-object-as-default-parameter.js
var NoObjectAsDefaultParameterRule = rule.Rule{
	Name:   "unicorn/no-object-as-default-parameter",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindParameter: func(node *ast.Node) {
				parameter := node.AsParameterDeclaration()
				if parameter == nil || parameter.Initializer == nil || node.Parent == nil ||
					!utils.IsFunctionLikeContainer(node.Parent) || node.Parent.Body() == nil {
					return
				}

				// typescript-eslint wraps constructor parameter properties in a
				// TSParameterProperty, so their AssignmentPattern is not a direct
				// child of the FunctionExpression that upstream checks.
				if ast.IsParameterPropertyDeclaration(node, node.Parent) {
					return
				}

				// ESTree drops parentheses around the AssignmentPattern right-hand
				// side. TypeScript-only wrappers remain visible upstream and must not
				// be treated as a direct ObjectExpression.
				initializer := ast.SkipParentheses(parameter.Initializer)
				if initializer == nil || initializer.Kind != ast.KindObjectLiteralExpression {
					return
				}

				object := initializer.AsObjectLiteralExpression()
				if object == nil || object.Properties == nil || len(object.Properties.Nodes) == 0 {
					return
				}

				name := parameter.Name()
				if name == nil {
					return
				}

				if name.Kind == ast.KindIdentifier {
					parameterName := name.AsIdentifier().Text
					ctx.ReportRange(
						utils.GetESTreeBindingIdentifierRange(ctx.SourceFile, name),
						identifierMessage(parameterName),
					)
					return
				}

				ctx.ReportNode(initializer, nonIdentifierMessage)
			},
		}
	},
}

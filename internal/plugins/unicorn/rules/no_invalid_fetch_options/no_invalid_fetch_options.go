package no_invalid_fetch_options

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const messageID = "no-invalid-fetch-options"

func invalidBodyMessage(method string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          messageID,
		Description: "\"body\" is not allowed when method is \"" + method + "\".",
	}
}

func isIdentifierPropertyNamed(property *ast.Node, name string) bool {
	if property == nil {
		return false
	}

	switch property.Kind {
	case ast.KindPropertyAssignment,
		ast.KindShorthandPropertyAssignment,
		ast.KindMethodDeclaration,
		ast.KindGetAccessor,
		ast.KindSetAccessor:
		propertyName := property.Name()
		return propertyName != nil && propertyName.Kind == ast.KindIdentifier &&
			propertyName.AsIdentifier().Text == name
	default:
		return false
	}
}

func findLastIdentifierProperty(properties []*ast.Node, name string) *ast.Node {
	for i := len(properties) - 1; i >= 0; i-- {
		if isIdentifierPropertyNamed(properties[i], name) {
			return properties[i]
		}
	}
	return nil
}

func propertyValue(property *ast.Node) *ast.Node {
	if property == nil {
		return nil
	}

	switch property.Kind {
	case ast.KindPropertyAssignment:
		return property.AsPropertyAssignment().Initializer
	case ast.KindShorthandPropertyAssignment:
		return property.Name()
	default:
		// ESTree represents object methods and accessors as Property nodes whose
		// values are functions. Neither can be absent-body values or static strings.
		return nil
	}
}

func isEffectivelyAbsentBody(property *ast.Node) bool {
	value := propertyValue(property)
	if value == nil {
		return false
	}
	value = ast.SkipParentheses(value)

	switch value.Kind {
	case ast.KindNullKeyword, ast.KindVoidExpression:
		return true
	case ast.KindIdentifier:
		return value.AsIdentifier().Text == "undefined"
	default:
		return false
	}
}

func secondArgumentOfDirectCall(node *ast.Node, calleeName string, rejectOptional bool) *ast.Node {
	if node == nil {
		return nil
	}

	var expression *ast.Node
	var arguments []*ast.Node
	switch node.Kind {
	case ast.KindCallExpression:
		if rejectOptional && ast.IsOptionalChainRoot(node) {
			return nil
		}
		call := node.AsCallExpression()
		expression = call.Expression
		if call.Arguments != nil {
			arguments = call.Arguments.Nodes
		}
	case ast.KindNewExpression:
		newExpression := node.AsNewExpression()
		expression = newExpression.Expression
		if newExpression.Arguments != nil {
			arguments = newExpression.Arguments.Nodes
		}
	default:
		return nil
	}

	if len(arguments) < 2 {
		return nil
	}

	// ESTree removes parentheses around a callee, but keeps TypeScript assertion
	// wrappers. Skip only parentheses to preserve that distinction.
	expression = ast.SkipOuterExpressions(expression, ast.OEKParentheses)
	if expression == nil || expression.Kind != ast.KindIdentifier ||
		expression.AsIdentifier().Text != calleeName {
		return nil
	}

	return arguments[1]
}

func checkFetchOptions(ctx rule.RuleContext, staticStrings *utils.StaticStringEvaluator, node *ast.Node) {
	node = ast.SkipParentheses(node)
	if node == nil || node.Kind != ast.KindObjectLiteralExpression {
		return
	}

	object := node.AsObjectLiteralExpression()
	if object.Properties == nil {
		return
	}
	properties := object.Properties.Nodes

	bodyProperty := findLastIdentifierProperty(properties, "body")
	if bodyProperty == nil || isEffectivelyAbsentBody(bodyProperty) {
		return
	}

	methodProperty := findLastIdentifierProperty(properties, "method")
	if methodProperty == nil {
		for _, property := range properties {
			if property.Kind == ast.KindSpreadAssignment {
				return
			}
		}

		ctx.ReportNode(bodyProperty.Name(), invalidBodyMessage("GET"))
		return
	}

	method, valueKind := staticStrings.EvalStringValue(propertyValue(methodProperty))
	if valueKind != utils.StaticEvalString {
		return
	}

	method = strings.ToUpper(method)
	if method != "GET" && method != "HEAD" {
		return
	}

	ctx.ReportNode(bodyProperty.Name(), invalidBodyMessage(method))
}

// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v73.0.0/rules/no-invalid-fetch-options.js
var NoInvalidFetchOptionsRule = rule.Rule{
	Name:   "unicorn/no-invalid-fetch-options",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		staticStrings := utils.NewStaticStringEvaluatorWithReferenceResolver(ctx.TypeChecker, ctx.SourceFile, ctx.Refs)

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				options := secondArgumentOfDirectCall(node, "fetch", true)
				if options != nil {
					checkFetchOptions(ctx, staticStrings, options)
				}
			},
			ast.KindNewExpression: func(node *ast.Node) {
				options := secondArgumentOfDirectCall(node, "Request", false)
				if options != nil {
					checkFetchOptions(ctx, staticStrings, options)
				}
			},
		}
	},
}

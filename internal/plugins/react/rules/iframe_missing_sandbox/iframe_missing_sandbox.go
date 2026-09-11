package iframe_missing_sandbox

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

const (
	msgAttributeMissing   = "An iframe element is missing a sandbox attribute"
	msgInvalidValue       = "An iframe element defines a sandbox attribute with invalid value \"{{ value }}\""
	msgInvalidCombination = "An iframe element defines a sandbox attribute with both allow-scripts and allow-same-origin which is invalid"
)

// allowedValues is the sandbox token list used by eslint-plugin-react v7.37.5.
// Empty tokens are intentional: splitting a string literal on a space produces
// them for leading, trailing, and repeated spaces, and upstream accepts them.
var allowedValues = map[string]struct{}{
	"": {},
	"allow-downloads-without-user-activation": {},
	"allow-downloads":                         {},
	"allow-forms":                             {},
	"allow-modals":                            {},
	"allow-orientation-lock":                  {},
	"allow-pointer-lock":                      {},
	"allow-popups":                            {},
	"allow-popups-to-escape-sandbox":          {},
	"allow-presentation":                      {},
	"allow-same-origin":                       {},
	"allow-scripts":                           {},
	"allow-storage-access-by-user-activation": {},
	"allow-top-navigation":                    {},
	"allow-top-navigation-by-user-activation": {},
}

func message(id, description string) rule.RuleMessage {
	return rule.RuleMessage{Id: id, Description: description}
}

// validateSandboxValue mirrors upstream's string-literal-only validation.
// In particular, JSX expression containers and non-string object values are
// dynamic and therefore accepted without attempting static evaluation.
func validateSandboxValue(ctx rule.RuleContext, node *ast.Node, value string) {
	allowScripts := false
	allowSameOrigin := false
	for _, token := range strings.Split(value, " ") {
		token = ecmascript.StringTrim(token)
		if _, ok := allowedValues[token]; !ok {
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "invalidValue",
				Description: strings.ReplaceAll(msgInvalidValue, "{{ value }}", token),
				Data:        map[string]string{"value": token},
			})
		}
		if token == "allow-scripts" {
			allowScripts = true
		}
		if token == "allow-same-origin" {
			allowSameOrigin = true
		}
	}
	if allowScripts && allowSameOrigin {
		ctx.ReportNode(node, message("invalidCombination", msgInvalidCombination))
	}
}

// objectSandboxProperty returns the first object member that upstream sees as
// `x.type === 'Property' && x.key.name === 'sandbox'`. Identifier keys and
// computed identifier keys have a `name` in ESTree; quoted keys do not.
func objectSandboxProperty(member *ast.Node) (value *ast.Node, found bool) {
	if member == nil {
		return nil, false
	}
	var name *ast.Node
	switch member.Kind {
	case ast.KindPropertyAssignment:
		property := member.AsPropertyAssignment()
		name = property.Name()
		if !sandboxName(name) {
			return nil, false
		}
		return property.Initializer, true
	case ast.KindShorthandPropertyAssignment:
		name = member.AsShorthandPropertyAssignment().Name()
	case ast.KindMethodDeclaration:
		name = member.AsMethodDeclaration().Name()
	case ast.KindGetAccessor:
		name = member.AsGetAccessorDeclaration().Name()
	case ast.KindSetAccessor:
		name = member.AsSetAccessorDeclaration().Name()
	default:
		return nil, false
	}
	if !sandboxName(name) {
		return nil, false
	}
	// Shorthand properties and function-like members are never Literal values
	// in ESTree, so they mark the prop as found without value validation.
	return nil, true
}

func sandboxName(name *ast.Node) bool {
	if name == nil {
		return false
	}
	if name.Kind == ast.KindIdentifier {
		return name.AsIdentifier().Text == "sandbox"
	}
	if name.Kind != ast.KindComputedPropertyName {
		return false
	}
	expression := name.AsComputedPropertyName().Expression
	return expression != nil && expression.Kind == ast.KindIdentifier && expression.AsIdentifier().Text == "sandbox"
}

var IframeMissingSandboxRule = rule.Rule{
	Name:   "react/iframe-missing-sandbox",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		checkJSX := func(node *ast.Node) {
			if reactutil.GetJsxElementTypeString(node) != "iframe" {
				return
			}
			found := false
			for _, attribute := range reactutil.GetJsxElementAttributes(node) {
				if attribute.Kind != ast.KindJsxAttribute || reactutil.GetJsxPropName(attribute) != "sandbox" {
					continue
				}
				found = true
				initializer := attribute.AsJsxAttribute().Initializer
				if initializer != nil && initializer.Kind == ast.KindStringLiteral {
					value, ok := reactutil.GetJsxStringLiteralValue(ctx.SourceFile, initializer)
					if ok && value != "" {
						validateSandboxValue(ctx, node, value)
					}
				}
			}
			if !found {
				ctx.ReportNode(node, message("attributeMissing", msgAttributeMissing))
			}
		}

		isCreateElement := reactutil.NewCreateElementCallMatcher(ctx)
		return rule.RuleListeners{
			ast.KindJsxOpeningElement:     checkJSX,
			ast.KindJsxSelfClosingElement: checkJSX,
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				if !isCreateElement(call.Expression) || call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
					return
				}
				tag := ast.SkipParentheses(call.Arguments.Nodes[0])
				if tag == nil || tag.Kind != ast.KindStringLiteral || tag.AsStringLiteral().Text != "iframe" {
					return
				}
				found := false
				if len(call.Arguments.Nodes) > 1 {
					props := ast.SkipParentheses(call.Arguments.Nodes[1])
					if props != nil && props.Kind == ast.KindObjectLiteralExpression {
						for _, member := range props.AsObjectLiteralExpression().Properties.Nodes {
							value, hasSandbox := objectSandboxProperty(member)
							if !hasSandbox {
								continue
							}
							found = true
							value = ast.SkipParentheses(value)
							if value != nil && value.Kind == ast.KindStringLiteral {
								text := value.AsStringLiteral().Text
								if text != "" {
									validateSandboxValue(ctx, node, text)
								}
							}
							break
						}
					}
				}
				if !found {
					ctx.ReportNode(node, message("attributeMissing", msgAttributeMissing))
				}
			},
		}
	},
}

package button_has_type

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type buttonHasTypeOptions struct {
	button bool
	submit bool
	reset  bool
}

func parseOptions(opts any) buttonHasTypeOptions {
	cfg := buttonHasTypeOptions{button: true, submit: true, reset: true}
	optsMap := utils.GetOptionsMap(opts)
	if optsMap == nil {
		return cfg
	}
	if v, ok := optsMap["button"]; ok {
		if b, ok := v.(bool); ok {
			cfg.button = b
		}
	}
	if v, ok := optsMap["submit"]; ok {
		if b, ok := v.(bool); ok {
			cfg.submit = b
		}
	}
	if v, ok := optsMap["reset"]; ok {
		if b, ok := v.(bool); ok {
			cfg.reset = b
		}
	}
	return cfg
}

var ButtonHasTypeRule = rule.Rule{
	Name: "react/button-has-type",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		cfg := parseOptions(options)

		reportMissing := func(node *ast.Node) {
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "missingType",
				Description: "Missing an explicit type attribute for button",
			})
		}
		reportComplex := func(node *ast.Node) {
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "complexType",
				Description: "The button type attribute must be specified by a static string or a trivial ternary expression",
			})
		}
		checkValue := func(reportOn *ast.Node, value string) {
			allowed, known := false, false
			switch value {
			case "button":
				allowed, known = cfg.button, true
			case "submit":
				allowed, known = cfg.submit, true
			case "reset":
				allowed, known = cfg.reset, true
			}
			if !known {
				ctx.ReportNode(reportOn, rule.RuleMessage{
					Id:          "invalidValue",
					Description: `"` + value + `" is an invalid value for button type attribute`,
				})
			} else if !allowed {
				ctx.ReportNode(reportOn, rule.RuleMessage{
					Id:          "forbiddenValue",
					Description: `"` + value + `" is an invalid value for button type attribute`,
				})
			}
		}

		var checkExpression func(reportOn *ast.Node, expr *ast.Node)
		checkExpression = func(reportOn *ast.Node, expr *ast.Node) {
			if expr == nil {
				reportComplex(reportOn)
				return
			}
			inner := ast.SkipParentheses(expr)
			if val, ok := utils.GetStaticExpressionValue(inner); ok {
				checkValue(reportOn, val)
				return
			}
			switch inner.Kind {
			case ast.KindTrueKeyword:
				checkValue(reportOn, "true")
			case ast.KindFalseKeyword:
				checkValue(reportOn, "false")
			case ast.KindNullKeyword:
				checkValue(reportOn, "null")
			case ast.KindBigIntLiteral:
				checkValue(reportOn, utils.NormalizeBigIntLiteral(inner.AsBigIntLiteral().Text))
			case ast.KindConditionalExpression:
				cond := inner.AsConditionalExpression()
				checkExpression(reportOn, cond.WhenTrue)
				checkExpression(reportOn, cond.WhenFalse)
			default:
				reportComplex(inner)
			}
		}

		checkJsxButton := func(reportOn *ast.Node, tagName *ast.Node, attributes *ast.Node) {
			if tagName == nil || tagName.Kind != ast.KindIdentifier || tagName.AsIdentifier().Text != "button" {
				return
			}
			typeAttr := findJsxTypeAttribute(attributes)
			if typeAttr == nil {
				reportMissing(reportOn)
				return
			}
			initializer := typeAttr.AsJsxAttribute().Initializer
			if initializer == nil {
				checkValue(reportOn, "true")
				return
			}
			if initializer.Kind == ast.KindJsxExpression {
				expr := initializer.AsJsxExpression().Expression
				checkExpression(reportOn, expr)
				return
			}
			if initializer.Kind == ast.KindStringLiteral {
				checkValue(reportOn, initializer.AsStringLiteral().Text)
				return
			}
		}

		return rule.RuleListeners{
			ast.KindJsxOpeningElement: func(node *ast.Node) {
				parent := node.Parent
				if parent == nil || parent.Kind != ast.KindJsxElement {
					return
				}
				opening := node.AsJsxOpeningElement()
				checkJsxButton(parent, opening.TagName, opening.Attributes)
			},
			ast.KindJsxSelfClosingElement: func(node *ast.Node) {
				self := node.AsJsxSelfClosingElement()
				checkJsxButton(node, self.TagName, self.Attributes)
			},
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				if !reactutil.IsCreateElementCall(call.Expression, reactutil.GetReactPragma(ctx.Settings)) {
					return
				}
				if call.Arguments == nil || len(call.Arguments.Nodes) < 1 {
					return
				}
				firstArg := ast.SkipParentheses(call.Arguments.Nodes[0])
				if firstArg.Kind != ast.KindStringLiteral || firstArg.AsStringLiteral().Text != "button" {
					return
				}
				if len(call.Arguments.Nodes) < 2 {
					reportMissing(node)
					return
				}
				secondArg := ast.SkipParentheses(call.Arguments.Nodes[1])
				if secondArg.Kind != ast.KindObjectLiteralExpression {
					reportMissing(node)
					return
				}
				obj := secondArg.AsObjectLiteralExpression()
				typeProp, typeValue := findObjectTypeProperty(obj)
				if typeProp == nil {
					reportMissing(node)
					return
				}
				checkExpression(node, typeValue)
			},
		}
	},
}

func findJsxTypeAttribute(attributes *ast.Node) *ast.Node {
	if attributes == nil {
		return nil
	}
	attrs := attributes.AsJsxAttributes()
	if attrs.Properties == nil {
		return nil
	}
	for _, prop := range attrs.Properties.Nodes {
		if prop.Kind == ast.KindJsxAttribute && reactutil.GetJsxPropName(prop) == "type" {
			return prop
		}
	}
	return nil
}

func findObjectTypeProperty(obj *ast.ObjectLiteralExpression) (*ast.Node, *ast.Node) {
	if obj == nil || obj.Properties == nil {
		return nil, nil
	}
	for _, prop := range obj.Properties.Nodes {
		switch prop.Kind {
		case ast.KindPropertyAssignment:
			pa := prop.AsPropertyAssignment()
			name := pa.Name()
			if name != nil && name.Kind == ast.KindIdentifier && name.AsIdentifier().Text == "type" {
				return prop, pa.Initializer
			}
		case ast.KindShorthandPropertyAssignment:
			sa := prop.AsShorthandPropertyAssignment()
			name := sa.Name()
			if name != nil && name.Kind == ast.KindIdentifier && name.AsIdentifier().Text == "type" {
				return prop, name
			}
		}
	}
	return nil, nil
}

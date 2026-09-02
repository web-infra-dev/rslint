package jsx_props_no_spreading

import (
	_ "embed"
	"unicode/utf8"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

//go:embed jsx_props_no_spreading.schema.json
var schemaJSON []byte

type options struct {
	html, custom, explicitSpread string
	exceptions                   []string
}

var defaultOptions = options{
	html:           "enforce",
	custom:         "enforce",
	explicitSpread: "enforce",
}

// JsxPropsNoSpreadingRule disallows JSX prop spreading, with separate
// controls for intrinsic elements, custom components, and explicit object
// spreads.
var JsxPropsNoSpreadingRule = rule.Rule{
	Name:   "react/jsx-props-no-spreading",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, rawOptions []any) rule.RuleListeners {
		opts := parseOptions(rawOptions)

		return rule.RuleListeners{
			ast.KindJsxSpreadAttribute: func(node *ast.Node) {
				parent := reactutil.GetJsxParentElement(node)
				tagName, hasTagName := jsxTagName(parent)

				if hasTagName {
					first, _ := utf8.DecodeRuneInString(tagName)
					firstString := string(first)
					startsWithUpper := firstString == ecmascript.StringToUpperCase(firstString)
					isHTMLTag := !startsWithUpper
					isCustomTag := startsWithUpper || containsDot(tagName)
					isException := contains(opts.exceptions, tagName)

					if isHTMLTag && ((opts.html == "ignore" && !isException) || (opts.html != "ignore" && isException)) {
						return
					}
					if isCustomTag && ((opts.custom == "ignore" && !isException) || (opts.custom != "ignore" && isException)) {
						return
					}
				}

				if opts.explicitSpread == "ignore" {
					argument := node.AsJsxSpreadAttribute().Expression
					argument = ast.SkipParentheses(argument)
					if argument != nil && argument.Kind == ast.KindObjectLiteralExpression && allObjectProperties(argument) {
						return
					}
				}

				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "noSpreading",
					Description: "Prop spreading is forbidden",
				})
			},
		}
	},
}

func parseOptions(raw []any) options {
	opts := defaultOptions
	if len(raw) == 0 {
		return opts
	}
	config, _ := raw[0].(map[string]any)
	if config == nil {
		return opts
	}
	if value, ok := config["html"].(string); ok {
		opts.html = value
	}
	if value, ok := config["custom"].(string); ok {
		opts.custom = value
	}
	if value, ok := config["explicitSpread"].(string); ok {
		opts.explicitSpread = value
	}
	if values, ok := config["exceptions"].([]any); ok {
		opts.exceptions = make([]string, 0, len(values))
		for _, value := range values {
			if exception, ok := value.(string); ok {
				opts.exceptions = append(opts.exceptions, exception)
			}
		}
	}
	return opts
}

func jsxTagName(element *ast.Node) (string, bool) {
	tagName := reactutil.GetJsxTagName(element)
	if tagName == nil {
		return "", false
	}
	switch tagName.Kind {
	case ast.KindIdentifier:
		return tagName.AsIdentifier().Text, true
	case ast.KindPropertyAccessExpression:
		return reactutil.GetJsxElementTypeString(tagName), true
	default:
		// JSXNamespacedName and other tag forms are not JSXIdentifier or
		// JSXMemberExpression in ESTree, so upstream leaves tagName undefined.
		return "", false
	}
}

func allObjectProperties(object *ast.Node) bool {
	properties := object.AsObjectLiteralExpression().Properties
	if properties == nil {
		return true
	}
	for _, property := range properties.Nodes {
		switch property.Kind {
		case ast.KindPropertyAssignment, ast.KindMethodDeclaration,
			ast.KindShorthandPropertyAssignment, ast.KindGetAccessor, ast.KindSetAccessor:
		default:
			return false
		}
	}
	return true
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func containsDot(value string) bool {
	for _, char := range value {
		if char == '.' {
			return true
		}
	}
	return false
}

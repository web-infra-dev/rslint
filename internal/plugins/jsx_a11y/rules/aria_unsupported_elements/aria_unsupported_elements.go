package aria_unsupported_elements

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/jsxa11yutil"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// reservedElements mirrors the `reserved: true` entries in aria-query's
// domMap. These are HTML elements that do NOT support ARIA roles, states, or
// properties — adding any aria-* attribute or `role` to them is invalid.
//
// Source: https://github.com/A11yance/aria-query/blob/main/src/domMap.js
var reservedElements = map[string]struct{}{
	"base":     {},
	"col":      {},
	"colgroup": {},
	"head":     {},
	"html":     {},
	"link":     {},
	"meta":     {},
	"noembed":  {},
	"noscript": {},
	"param":    {},
	"picture":  {},
	"script":   {},
	"source":   {},
	"style":    {},
	"title":    {},
	"track":    {},
}

// invalidAttributes is the set of attribute names that must not appear on
// reserved elements. Mirrors `aria.keys().concat('role')` from aria-query —
// every defined ARIA state/property (sourced from jsxa11yutil.AriaPropertySet)
// plus the `role` attribute.
var invalidAttributes = func() map[string]struct{} {
	out := make(map[string]struct{}, len(jsxa11yutil.AriaPropertySet)+1)
	for k := range jsxa11yutil.AriaPropertySet {
		out[k] = struct{}{}
	}
	out["role"] = struct{}{}
	return out
}()

func errorMessage(invalidProp string) string {
	return "This element does not support ARIA roles, states and properties. Try removing the prop '" + invalidProp + "'."
}

var AriaUnsupportedElementsRule = rule.Rule{
	Name: "jsx-a11y/aria-unsupported-elements",
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		check := func(node *ast.Node) {
			nodeType := jsxa11yutil.GetElementType(node, ctx.Settings)
			if _, reserved := reservedElements[nodeType]; !reserved {
				return
			}
			for _, attr := range reactutil.GetJsxElementAttributes(node) {
				if attr.Kind != ast.KindJsxAttribute {
					continue
				}
				name := strings.ToLower(reactutil.GetJsxPropName(attr))
				if _, bad := invalidAttributes[name]; bad {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "unsupported",
						Description: errorMessage(name),
					})
				}
			}
		}
		return rule.RuleListeners{
			ast.KindJsxOpeningElement:     check,
			ast.KindJsxSelfClosingElement: check,
		}
	},
}

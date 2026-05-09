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
// every defined ARIA state/property plus the `role` attribute.
//
// Source: https://github.com/A11yance/aria-query/blob/main/src/ariaPropsMap.js
var invalidAttributes = map[string]struct{}{
	"role":                        {},
	"aria-activedescendant":       {},
	"aria-atomic":                 {},
	"aria-autocomplete":           {},
	"aria-braillelabel":           {},
	"aria-brailleroledescription": {},
	"aria-busy":                   {},
	"aria-checked":                {},
	"aria-colcount":               {},
	"aria-colindex":               {},
	"aria-colspan":                {},
	"aria-controls":               {},
	"aria-current":                {},
	"aria-describedby":            {},
	"aria-description":            {},
	"aria-details":                {},
	"aria-disabled":               {},
	"aria-dropeffect":             {},
	"aria-errormessage":           {},
	"aria-expanded":               {},
	"aria-flowto":                 {},
	"aria-grabbed":                {},
	"aria-haspopup":               {},
	"aria-hidden":                 {},
	"aria-invalid":                {},
	"aria-keyshortcuts":           {},
	"aria-label":                  {},
	"aria-labelledby":             {},
	"aria-level":                  {},
	"aria-live":                   {},
	"aria-modal":                  {},
	"aria-multiline":              {},
	"aria-multiselectable":        {},
	"aria-orientation":            {},
	"aria-owns":                   {},
	"aria-placeholder":            {},
	"aria-posinset":               {},
	"aria-pressed":                {},
	"aria-readonly":               {},
	"aria-relevant":               {},
	"aria-required":               {},
	"aria-roledescription":        {},
	"aria-rowcount":               {},
	"aria-rowindex":               {},
	"aria-rowspan":                {},
	"aria-selected":               {},
	"aria-setsize":                {},
	"aria-sort":                   {},
	"aria-valuemax":               {},
	"aria-valuemin":               {},
	"aria-valuenow":               {},
	"aria-valuetext":              {},
}

func errorMessage(invalidProp string) string {
	return "This element does not support ARIA roles, states and properties. Try removing the prop '" + invalidProp + "'."
}

var AriaUnsupportedElementsRule = rule.Rule{
	Name: "jsx-a11y/aria-unsupported-elements",
	Run: func(ctx rule.RuleContext, _ any) rule.RuleListeners {
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

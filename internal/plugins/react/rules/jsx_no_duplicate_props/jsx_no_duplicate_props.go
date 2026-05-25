package jsx_no_duplicate_props

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var JsxNoDuplicatePropsRule = rule.Rule{
	Name: "react/jsx-no-duplicate-props",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		ignoreCase := false
		if optsMap := utils.GetOptionsMap(options); optsMap != nil {
			if v, ok := optsMap["ignoreCase"].(bool); ok {
				ignoreCase = v
			}
		}

		check := func(node *ast.Node) {
			attrs := reactutil.GetJsxElementAttributes(node)
			if len(attrs) == 0 {
				return
			}

			seen := map[string]struct{}{}
			for _, attr := range attrs {
				// Spread attributes ({...x}) are filtered here; duplicate tracking
				// across a spread continues to flag names seen before the spread,
				// matching upstream eslint-plugin-react.
				if !ast.IsJsxAttribute(attr) {
					continue
				}
				nameNode := attr.AsJsxAttribute().Name()
				if nameNode == nil {
					continue
				}
				// Upstream: `if (typeof decl.name.name !== 'string') return;` —
				// JSXNamespacedName's `.name` is a JSXIdentifier node, not a
				// string, so namespaced attributes (e.g. `a:b`) are skipped
				// entirely and do not participate in duplicate tracking.
				if nameNode.Kind != ast.KindIdentifier {
					continue
				}
				name := nameNode.AsIdentifier().Text
				if ignoreCase {
					name = strings.ToLower(name)
				}
				if _, dup := seen[name]; dup {
					ctx.ReportNode(attr, rule.RuleMessage{
						Id:          "noDuplicateProps",
						Description: "No duplicate props allowed",
					})
					continue
				}
				seen[name] = struct{}{}
			}
		}

		return rule.RuleListeners{
			ast.KindJsxOpeningElement:     check,
			ast.KindJsxSelfClosingElement: check,
		}
	},
}

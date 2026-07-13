// Package lang ports eslint-plugin-jsx-a11y's `lang` rule. The rule extends
// `html-has-lang` by additionally enforcing that the value of an `<html>`
// element's `lang` prop is a recognized BCP-47 language tag, per
// [WCAG 3.1.1](https://www.w3.org/WAI/WCAG21/Understanding/language-of-page).
package lang

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/jsxa11yutil"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

const errorMessage = "lang attribute must have a valid value."

var LangRule = rule.Rule{
	Name: "jsx-a11y/lang",
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindJsxAttribute: func(attr *ast.Node) {
				// Upstream: `if (name && name.toUpperCase() !== 'LANG') return;`
				// — `name && ...` short-circuits on the empty-name defensive
				// path (tsgo always emits a non-empty Identifier or
				// JsxNamespacedName for legal source). Mirror exactly so we
				// don't accidentally skip cases upstream would still process.
				name := reactutil.GetJsxPropName(attr)
				if name != "" && !strings.EqualFold(name, "lang") {
					return
				}
				parent := reactutil.GetJsxParentElement(attr)
				if parent == nil {
					return
				}
				// Upstream: `if (type && type !== 'html') return;`. Logical-AND
				// with truthy type — empty type falls through (computed
				// member tag-name shapes tsgo can't resolve). Mirror byte-
				// for-byte so the listener gate matches upstream.
				tagType := jsxa11yutil.GetElementType(parent, ctx.Settings)
				if tagType != "" && tagType != "html" {
					return
				}
				// jsx-ast-utils' `getLiteralPropValue(node)`:
				//   - null      → caller `value === null` → SKIP
				//   - undefined → caller `value === undefined` → REPORT
				//   - string    → caller validates via tags.check
				//   - non-string literal → caller `tags.check(non-string)` throws
				//                          → we report instead of crashing
				val, state := jsxa11yutil.LiteralPropExtract(attr)
				switch state {
				case jsxa11yutil.ExtractUnresolvable:
					// Identifier (non-undefined), CallExpression,
					// MemberExpression, ConditionalExpression, TS-wrappers
					// (`undefined as any`), etc. → upstream null → no report.
					return
				case jsxa11yutil.ExtractUndefined, jsxa11yutil.ExtractOther:
					// `ExtractUndefined`: explicit `undefined` Identifier —
					// upstream's `value === undefined` arm.
					// `ExtractOther`: boolean attribute form, TrueKeyword,
					// FalseKeyword, "true"/"false" string coerced to bool,
					// NumericLiteral, BigIntLiteral, Array/ObjectExpression.
					// None can be a BCP-47 tag — upstream `tags.check` throws
					// `TypeError` on non-string input. We surface as a normal
					// diagnostic so the user sees the problem without losing
					// the rest of the file's reports.
					reportInvalid(ctx, attr)
				case jsxa11yutil.ExtractString:
					if !isValidBCP47Tag(val) {
						reportInvalid(ctx, attr)
					}
				}
			},
		}
	},
}

func reportInvalid(ctx rule.RuleContext, attr *ast.Node) {
	ctx.ReportNode(attr, rule.RuleMessage{
		Id:          "invalidLangValue",
		Description: errorMessage,
	})
}

// isValidBCP47Tag is defined in bcp47.go.

package jsxa11yutil

import "github.com/microsoft/typescript-go/shim/ast"

// IsDisabledElement mirrors upstream `isDisabledElement(attributes)`
// (eslint-plugin-jsx-a11y/src/util/isDisabledElement.js) verbatim:
//
//	const disabledAttr = getProp(attributes, 'disabled');
//	const disabledAttrValue = getPropValue(disabledAttr);
//	const isHTML5Disabled = disabledAttr && disabledAttrValue !== undefined;
//	if (isHTML5Disabled) return true;
//	const ariaDisabledAttr = getProp(attributes, 'aria-disabled');
//	const ariaDisabledAttrValue = getLiteralPropValue(ariaDisabledAttr);
//	if (ariaDisabledAttr && ariaDisabledAttrValue !== undefined
//	    && ariaDisabledAttrValue === true) return true;
//	return false;
//
// Notable upstream quirks we preserve:
//
//   - The `disabled` arm gates on `getPropValue !== undefined`, NOT
//     truthiness. `<X disabled={false} />` therefore counts as disabled
//     (getPropValue is the literal boolean `false`, not undefined). Only
//     `<X disabled={undefined} />` (and TS-wrapped variants) avoids the trip.
//     This is counter-intuitive but matches upstream byte-for-byte.
//   - The `disabled` arm uses `getPropValue` (full static eval), so
//     `disabled={someVar}` resolves to the identifier name string and
//     trips the trigger. The `aria-disabled` arm uses `getLiteralPropValue`
//     with strict `=== true`, so only literal boolean true (including the
//     boolean-attribute form `<X aria-disabled />` and the
//     literal-coerced-string `aria-disabled="true"`) counts.
func IsDisabledElement(attrs []*ast.Node) bool {
	if disabledAttr := FindAttributeByName(attrs, "disabled"); disabledAttr != nil {
		// upstream `getPropValue(disabledAttr) !== undefined`. The only
		// shapes that resolve to JS undefined are:
		//   - explicit `disabled={undefined}` (and TS-wrapped variants).
		// Boolean form, missing initializer (empty `{}`), string, number,
		// identifier, call, member, etc. all resolve to non-undefined and
		// therefore count as HTML5-disabled.
		if !AttributeIsExplicitUndefined(disabledAttr) {
			return true
		}
	}
	if ariaDisabledAttr := FindAttributeByName(attrs, "aria-disabled"); ariaDisabledAttr != nil {
		// upstream `getLiteralPropValue(...) === true`. LiteralPropIsExactlyTrue
		// covers the boolean-attribute form (extractValue null-attr → true),
		// `={true}`, and the literal-coerced string `="true"`. Anything else
		// (including `="false"`, `={false}`, `={1}`, identifiers, calls,
		// missing prop) falls through.
		if LiteralPropIsExactlyTrue(ariaDisabledAttr) {
			return true
		}
	}
	return false
}

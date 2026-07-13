// cspell:ignore eading

// Package role_has_required_aria_props ports eslint-plugin-jsx-a11y's
// `role-has-required-aria-props` rule. The rule listens on every JSX
// `role` attribute whose value is a literal-typed string, and reports
// when one of the named ARIA roles is missing its `requiredProps`
// (per aria-query's `roles.get(role).requiredProps`).
//
// Trigger order — mirrors upstream's `JSXAttribute` listener step-for-step:
//
//  1. `propName(attribute).toLowerCase() !== 'role'` → skip. We use
//     `strings.EqualFold` against "role" to match jsx-ast-utils'
//     case-insensitive comparison.
//  2. `!dom.get(elementType)` → skip. The element type is resolved via
//     `jsxa11yutil.GetElementType` (which honors the `components` and
//     `polymorphicPropName` jsx-a11y settings); the DOM-element membership
//     check uses `jsxa11yutil.IsDOMElement`. Custom components and
//     unknown element names are skipped.
//  3. Literal-value extraction in priority order — array → string →
//     bool/number → undef/NoLit:
//     a. `LiteralPropArrayAsString` handles `<div role={[...]}>`
//     shapes (upstream's LITERAL_TYPES.ArrayExpression evaluates
//     elements and `,`-joins via `String([...])`).
//     b. `LiteralPropStringValue` handles string-typed literal values
//     including the direct attribute form with HTML entity decoding
//     (`<div role="&#104;eading">` → "heading"), JsxExpression-wrapped
//     strings, the `null`-magic-string override, and template-synthesized
//     strings.
//     c. `LiteralPropAriaValue` is the fallback for boolean / number /
//     BigInt literal values (stringified via JS `String()`), and the
//     short-circuit `NoLit` / `Undef` arms that skip:
//     - boolean-form attribute `<div role />` returns
//     `Bool{true}` which stringifies to `"true"` (not a role).
//     - explicit `{undefined}` → `Undef` → skip.
//     - non-literal expression (Identifier, Logical, Conditional,
//     Call, Member, Binary, TS-wrapper, ...) → `NoLit` → skip.
//     Explicit `{null}` is NOT skipped here: it's handled by
//     step (b) via the LITERAL_TYPES.Literal override → magic
//     string `"null"`, which then falls through to step 4.
//  4. Lowercase and ASCII-space-split the role value. Then filter to the
//     tokens that exist in aria-query's non-abstract role set.
//  5. `isSemanticRoleElement(elementType, attributes)` → skip. The
//     check uses the un-normalized (i.e. NOT lowercased) `roleAttrValue`
//     for comparison against axobject-query's role-name table, so
//     `<select role="combobox" />` is skipped but `<select role="COMBOBOX" />`
//     is NOT — see Phase 1 step 6 below for the lock-in test.
//  6. For each valid role with non-empty `requiredProps`, check that
//     every required prop is present as a JSX attribute (via
//     `jsxa11yutil.FindAttributeByName`, which honors literal-spread
//     object expansion). If any required prop is missing, emit a
//     diagnostic on the role attribute node, with the role name
//     lowercased (`role.toLowerCase()`) and the required-props list
//     comma-joined (matching JS `String(array)` semantics).
//
// Phase 1 Step 6 — observable divergences from upstream:
//   - None. The rule is a thin port of the upstream `JSXAttribute`
//     listener. The only nuance is that upstream's `isSemanticRoleElement`
//     uses the raw (case-preserved) `roleAttrValue` while the
//     `validRoles` filter uses the lowercased form — so
//     `<select role="COMBOBOX" />` reports (validRoles → "combobox" via
//     lowercase split, but isSemanticRoleElement comparing "COMBOBOX" to
//     aria-query's lowercase "combobox" returns false). We mirror that
//     quirk verbatim.
package role_has_required_aria_props

import (
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/jsxa11yutil"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// errorMessage mirrors upstream's `errorMessage(role, requiredProps)`
// template byte-for-byte. `requiredProps` is `String(array)`-joined (i.e.
// JS `Array.prototype.toString()` semantics, which uses ",") and then
// lowercased — aria-query's prop names are already lowercase, so the
// `.toLowerCase()` call is a defensive no-op upstream that we mirror by
// constructing the joined string directly in lowercase.
func errorMessage(role string, requiredProps []string) string {
	return fmt.Sprintf(
		`Elements with the ARIA role "%s" must have the following attributes defined: %s`,
		role,
		strings.Join(requiredProps, ","),
	)
}

// isSemanticRoleElement mirrors upstream's `isSemanticRoleElement(type,
// attributes)`:
//
//	elementAXObjects.forEach((axObjects, concept) => {
//	  if (concept.name === type && concept.attributes.every(matches)) {
//	    for axObject in axObjects:
//	      for role in AXObjectRoles.get(axObject):
//	        if role.name === roleAttrValue: return true
//	  }
//	});
//
// We pre-joined the (concept → axObjects → roles) chain into
// `jsxa11yutil.SemanticRoleConcepts` so the runtime check is a flat
// concept-by-concept scan.
//
// Key parity notes:
//
//   - The concept-attribute name comparison is CASE-SENSITIVE
//     (`cAttr.name === propName(attr)` in upstream — `propName` returns
//     the raw AST name without case normalization). Upstream's outer
//     `name !== 'role'` check lowercases, but the inner concept-attribute
//     matcher does not. So `<input Type="checkbox" role="switch" />`
//     (capital T) does NOT match the input/type=checkbox concept.
//   - The concept-attribute value comparison uses upstream's
//     `getLiteralPropValue(attr)` — i.e. only literal-typed values
//     count; dynamic / identifier values silently fail to match.
//     Implemented here via [jsxa11yutil.LiteralPropStringValue], which
//     mirrors the LITERAL_TYPES extractor for string-typed results.
//   - The role-name comparison is the raw `roleAttrValue`, NOT the
//     lowercased / space-split form. So `<select role="COMBOBOX" />`
//     fails the semantic skip even though the validRoles filter sees
//     "combobox".
func isSemanticRoleElement(elementType string, attrs []*ast.Node, roleAttrValue string) bool {
	for i := range jsxa11yutil.SemanticRoleConcepts {
		concept := &jsxa11yutil.SemanticRoleConcepts[i]
		if concept.Name != elementType {
			continue
		}
		// All concept attributes must match an attribute on the JSX element.
		allMatch := true
		for _, cAttr := range concept.Attributes {
			matched := false
			for _, attr := range attrs {
				if attr.Kind != ast.KindJsxAttribute {
					// JsxSpreadAttribute is opaque — upstream's
					// `attribute.some(...)` filters on `attr.type ===
					// 'JSXAttribute'`. Match by skipping spreads here.
					continue
				}
				if reactutil.GetJsxPropName(attr) != cAttr.Name {
					continue
				}
				// Name matches. If the concept attribute carries a value,
				// the JSX attribute's literal value must match exactly
				// (case-sensitive). When the concept attribute has no
				// value (cAttr.Value == ""), name match alone is enough.
				if cAttr.Value != "" {
					attrValue, ok := jsxa11yutil.LiteralPropStringValue(attr)
					if !ok || attrValue != cAttr.Value {
						continue
					}
				}
				matched = true
				break
			}
			if !matched {
				allMatch = false
				break
			}
		}
		if !allMatch {
			continue
		}
		// Concept matches. Check if the role's name is among the concept's
		// implied roles.
		for _, r := range concept.Roles {
			if r == roleAttrValue {
				return true
			}
		}
	}
	return false
}

var RoleHasRequiredAriaPropsRule = rule.Rule{
	Name: "jsx-a11y/role-has-required-aria-props",
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindJsxAttribute: func(attr *ast.Node) {
				// Step 1: name === 'role' (case-insensitive — mirrors
				// upstream `propName(attribute).toLowerCase() !== 'role'`).
				if !strings.EqualFold(reactutil.GetJsxPropName(attr), "role") {
					return
				}

				// Step 2: resolve element type and gate on dom.get(type).
				parent := reactutil.GetJsxParentElement(attr)
				if parent == nil {
					return
				}
				elementType := jsxa11yutil.GetElementType(parent, ctx.Settings)
				if !jsxa11yutil.IsDOMElement(elementType) {
					return
				}

				// Step 3: literal value extraction. We need the upstream
				// `getLiteralPropValue` "undefined vs null vs other"
				// trichotomy:
				//   - boolean-form attribute → returns true (boolean)
				//   - explicit `{undefined}` → returns undefined
				//   - non-literal expression → returns null
				//   - direct string attribute → returns the parsed (entity-decoded) text
				//   - everything else → proceed (stringified via JS String())
				//
				// Three helpers cover this, in priority order:
				//
				//  1. `LiteralPropArrayAsString` — array values
				//     (`<div role={["x"]}>`). Upstream's
				//     LITERAL_TYPES.ArrayExpression evaluates elements via
				//     the full TYPES evaluator and `,`-joins per
				//     `String([...])`. `LiteralPropAriaValue` keeps arrays
				//     as NoLit, so the array path must be checked FIRST.
				//
				//  2. `LiteralPropStringValue` — string-typed results,
				//     including direct-attribute strings with HTML entity
				//     decoding (`<div role="&#104;eading">` → "heading"),
				//     JsxExpression-wrapped strings (`<div role={"x"}>`),
				//     `null` keyword (LITERAL_TYPES.Literal override →
				//     magic string `"null"`), and TemplateExpression
				//     synthesized strings.
				//
				//     CRITICAL: do NOT skip directly to LiteralPropAriaValue
				//     for the string case — that path reads
				//     `StringLiteral.Text` raw, so it misses HTML entity
				//     decoding on direct JSX attribute strings (where the
				//     `&…;` is part of the source). babel / @typescript-eslint
				//     JSX parsers expose decoded text on the AST; tsgo
				//     preserves the raw source. The decode happens inside
				//     `directAttributeStringValue` (called by
				//     `LiteralPropStringValue`).
				//
				//  3. `LiteralPropAriaValue` — fallback for non-string
				//     literal types (boolean, number, BigInt) and the
				//     short-circuit cases (NoLit / Undef → skip).
				//
				// `isStringValue` tracks whether the original literal was
				// a JS string (as opposed to array / number / boolean).
				// Upstream's `isSemanticRoleElement` compares
				// `role.name === roleAttrValue` with strict equality, so
				// `role.name` (always a string) only matches when the
				// extracted value is a string. We skip the semantic
				// check entirely for non-string values to preserve that
				// asymmetry.
				var roleAttrValue string
				var isStringValue bool
				if arrStr, isArr := jsxa11yutil.LiteralPropArrayAsString(attr); isArr {
					// Array — upstream's strict-equality check against
					// `role.name` (a string) is always false, so
					// isStringValue stays false.
					roleAttrValue = arrStr
				} else if str, ok := jsxa11yutil.LiteralPropStringValue(attr); ok {
					// String — direct or JsxExpression-wrapped or magic
					// "null" or template-synthesized. Entity-decoded by
					// directAttributeStringValue for the direct form.
					roleAttrValue = str
					isStringValue = true
				} else {
					value := jsxa11yutil.LiteralPropAriaValue(attr)
					if value.Kind == jsxa11yutil.AriaLiteralNoLit ||
						value.Kind == jsxa11yutil.AriaLiteralUndef {
						return
					}
					// Boolean / Number / BigInt — stringify per JS String()
					// semantics but keep isStringValue false because the
					// original literal wasn't a string.
					roleAttrValue = jsxa11yutil.AriaLiteralValueAsJSString(value)
				}

				// Step 4: normalize (lowercase + ASCII-space split) and
				// keep only the tokens that are recognized non-abstract
				// ARIA roles per aria-query's `roles.keys()`. Upstream
				// uses `.split(' ')` (single ASCII space, NOT `\s+`), so
				// tabs / newlines / multiple spaces produce empty or
				// non-matching tokens that fall out at this filter step.
				normalized := strings.Split(strings.ToLower(roleAttrValue), " ")
				validRoles := make([]string, 0, len(normalized))
				for _, tok := range normalized {
					if jsxa11yutil.IsValidAriaRole(tok) {
						validRoles = append(validRoles, tok)
					}
				}
				if len(validRoles) == 0 {
					return
				}

				// Step 5: semantic-role skip. Uses the un-normalized
				// (case-preserved) roleAttrValue — see the
				// isSemanticRoleElement docstring for the quirk
				// rationale. Skip entirely when the original literal
				// wasn't a string — upstream's strict-equality
				// comparison against `role.name` (always a string) is
				// always false for array / number / boolean values, so
				// the semantic skip never fires on those shapes.
				attrs := reactutil.GetJsxElementAttributes(parent)
				if isStringValue && isSemanticRoleElement(elementType, attrs, roleAttrValue) {
					return
				}

				// Step 6: required-props check. For each valid role with
				// non-empty requiredProps (per aria-query's
				// `requiredProps` map), verify every required prop is
				// present as a JSX attribute. Missing → report. Upstream
				// emits one diagnostic per failing role; we mirror by
				// reporting inside the per-role loop without breaking.
				for _, role := range validRoles {
					required, ok := jsxa11yutil.AriaRoleRequiredPropsFor(role)
					if !ok {
						continue
					}
					hasAll := true
					for _, prop := range required {
						if jsxa11yutil.FindAttributeByName(attrs, prop) == nil {
							hasAll = false
							break
						}
					}
					if hasAll {
						continue
					}
					ctx.ReportNode(attr, rule.RuleMessage{
						Id:          "role-has-required-aria-props",
						Description: errorMessage(role, required),
					})
				}
			},
		}
	},
}

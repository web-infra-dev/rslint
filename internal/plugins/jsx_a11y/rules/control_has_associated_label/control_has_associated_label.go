// Package control_has_associated_label ports eslint-plugin-jsx-a11y's
// `control-has-associated-label` rule. Enforces that an interactive control
// (DOM element, control component, or DOM element with an interactive role)
// has a discernible text label — either through visible text content, a
// labelling attribute (`alt` / `aria-label` / `aria-labelledby` / user-
// configured `labelAttributes`), or descendant content within a configured
// recursion depth.
//
// Upstream listener gate (`JSXElement`), checked in order:
//
//  1. `newIgnoreElements` (= user `ignoreElements` ∪ `['link']`) contains the
//     resolved tag → skip. The `link` exemption is hard-coded upstream and
//     cannot be disabled via config.
//  2. `getLiteralPropValue` of the `role` attribute is in `ignoreRoles` → skip.
//  3. Element is hidden from screen readers (`<input type="hidden">`,
//     `aria-hidden={true}`, `aria-hidden="true"`) → skip.
//  4. Trigger condition: `isInteractiveElement || (isDOMElement &&
//     isInteractiveRole) || controlComponents.indexOf(tag) > -1`.
//     - When false: no label requirement (rule does not apply).
//     - When true: run `mayHaveAccessibleLabel(root, min(options.depth ?? 2,
//     25), labelAttributes, getElementType, controlComponents)`.
//  5. If no label is found, report on the opening element.
//
// The message id is `controlHasAssociatedLabel`; the message text is taken
// verbatim from upstream's `errorMessage`.
package control_has_associated_label

import (
	"slices"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/jsxa11yutil"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// errorMessage mirrors upstream's `errorMessage` constant verbatim.
const errorMessage = "A control must be associated with a text label."

// defaultDepth mirrors upstream's `options.depth === undefined ? 2 : options.depth`.
const defaultDepth = 2

// maxDepthCap mirrors upstream's `Math.min(..., 25)` ceiling on the recursion
// budget — protects against pathological JSX trees regardless of user config.
const maxDepthCap = 25

type options struct {
	labelAttributes   []string
	controlComponents []string
	ignoreElements    []string
	ignoreRoles       []string
	depth             int
}

func parseOptions(raw any) options {
	opts := options{depth: defaultDepth}
	m := utils.GetOptionsMap(raw)
	if m == nil {
		return opts
	}
	opts.labelAttributes = jsxa11yutil.StringSliceOption(m["labelAttributes"])
	opts.controlComponents = jsxa11yutil.StringSliceOption(m["controlComponents"])
	opts.ignoreElements = jsxa11yutil.StringSliceOption(m["ignoreElements"])
	opts.ignoreRoles = jsxa11yutil.StringSliceOption(m["ignoreRoles"])
	// depth: number, default 2. JSON decodes numbers as float64.
	// Upstream truthiness: `options.depth === undefined ? 2 : options.depth`,
	// then `Math.min(depth, 25)`. We treat the absence of the key (`v == nil`)
	// as "undefined" and apply the default.
	if v, ok := m["depth"]; ok && v != nil {
		if f, ok := v.(float64); ok {
			opts.depth = int(f)
		}
	}
	if opts.depth > maxDepthCap {
		opts.depth = maxDepthCap
	}
	return opts
}

var ControlHasAssociatedLabelRule = rule.Rule{
	Name: "jsx-a11y/control-has-associated-label",
	Run: func(ctx rule.RuleContext, rawOptions any) rule.RuleListeners {
		opts := parseOptions(rawOptions)
		// Upstream `newIgnoreElements = new Set([].concat(ignoreElements,
		// ignoreList))` where `ignoreList = ['link']`. The `link` entry is
		// applied unconditionally — there is no way to opt out via config.
		newIgnoreElements := append([]string{}, opts.ignoreElements...)
		newIgnoreElements = append(newIgnoreElements, "link")

		getElementType := func(node *ast.Node) string {
			return jsxa11yutil.GetElementType(node, ctx.Settings)
		}

		check := func(elementNode *ast.Node) {
			opening, openingAttrs := openingElementOf(elementNode)
			if opening == nil {
				return
			}
			tag := getElementType(opening)

			// Step 1: newIgnoreElements (user ignoreElements ∪ ['link']).
			// Exact match — matches upstream's `Set.has(tag)`.
			if slices.Contains(newIgnoreElements, tag) {
				return
			}

			// Step 2: `getLiteralPropValue(getProp(attributes, 'role'))` in
			// `ignoreRoles`. Routes through `LiteralPropStringValue` (=
			// upstream's `getLiteralPropValue` for string-typed results) so
			// non-literal role expressions (Identifier, Call, Conditional)
			// fall through to ok=false and never satisfy `ignoreRoles`.
			roleAttr := jsxa11yutil.FindAttributeByName(openingAttrs, "role")
			if roleAttr != nil {
				if roleValue, ok := jsxa11yutil.LiteralPropStringValue(roleAttr); ok {
					if slices.Contains(opts.ignoreRoles, roleValue) {
						return
					}
				}
			}

			// Step 3: hidden from screen reader — same helper used across
			// the plugin (e.g. click-events-have-key-events). Treats
			// `<input type="hidden">` and `aria-hidden={true}` /
			// `aria-hidden="true"` as hidden.
			if jsxa11yutil.IsHiddenFromScreenReader(opening, getElementType) {
				return
			}

			// Step 4: trigger condition.
			// Upstream: `isInteractiveElement || (isDOMElement &&
			// isInteractiveRole) || controlComponents.indexOf(tag) > -1`.
			// `controlComponents.indexOf` is EXACT match (case-sensitive)
			// and does NOT use minimatch — the minimatch comparison only
			// appears inside `mayHaveAccessibleLabel`'s React-component
			// fallback. Preserve this asymmetry.
			nodeIsInteractiveElement := jsxa11yutil.IsInteractiveElement(tag, openingAttrs)
			nodeIsDOMElement := jsxa11yutil.IsDOMElement(tag)
			nodeIsInteractiveRole := jsxa11yutil.IsInteractiveRole(tag, openingAttrs)
			nodeIsControlComponent := slices.Contains(opts.controlComponents, tag)
			shouldCheck := nodeIsInteractiveElement ||
				(nodeIsDOMElement && nodeIsInteractiveRole) ||
				nodeIsControlComponent
			if !shouldCheck {
				return
			}

			if mayHaveAccessibleLabel(elementNode, 0, opts.depth, opts.labelAttributes, opts.controlComponents, getElementType) {
				return
			}

			ctx.ReportNode(opening, rule.RuleMessage{
				Id:          "controlHasAssociatedLabel",
				Description: errorMessage,
			})
		}

		// Listen on both paired (JsxElement) and self-closing
		// (JsxSelfClosingElement) — tsgo splits these into separate kinds
		// while ESTree (and therefore upstream) sees them as JSXElement with
		// `selfClosing` flag. Both forms must be classified independently.
		return rule.RuleListeners{
			ast.KindJsxElement:            check,
			ast.KindJsxSelfClosingElement: check,
		}
	},
}

// openingElementOf returns the focal "opening element" used for label-prop
// inspection and reporting, plus that element's attribute list. For a paired
// JsxElement, this is `node.openingElement`; for a JsxSelfClosingElement, the
// node itself plays both roles.
func openingElementOf(node *ast.Node) (*ast.Node, []*ast.Node) {
	if node == nil {
		return nil, nil
	}
	switch node.Kind {
	case ast.KindJsxElement:
		opening := node.AsJsxElement().OpeningElement
		if opening == nil {
			return nil, nil
		}
		return opening, reactutil.GetJsxElementAttributes(opening)
	case ast.KindJsxSelfClosingElement:
		return node, reactutil.GetJsxElementAttributes(node)
	}
	return nil, nil
}

// mayHaveAccessibleLabel mirrors upstream's
//
//	mayHaveAccessibleLabel(root, maxDepth, additionalLabellingProps,
//	                      getElementType, controlComponents)
//
// — a DFS that returns true if any of the following conditions hold within
// `maxDepth` levels of descent:
//
//   - The node is a JSX text / string literal whose trimmed value is non-empty.
//   - The node is a JsxExpression (`{expr}`) — assumed to render a label.
//   - The node has an opening element that declares a labelling prop with a
//     non-empty trimmed value (`alt`, `aria-label`, `aria-labelledby`, plus
//     any `labelAttributes` configured by the user). Spread attributes are
//     opaque and count as labelling.
//   - The node is a JSX element with no children whose tag name starts with
//     uppercase (= React component) AND is not in the `controlComponents`
//     minimatch list. This codifies "an opaque React component might render
//     a label".
//
// On every recursion the function fires before any of the above checks. The
// caller passes `depth=0` for the root.
func mayHaveAccessibleLabel(node *ast.Node, depth, maxDepth int, labelAttributes, controlComponents []string, getElementType func(*ast.Node) string) bool {
	if node == nil {
		return false
	}
	// Bail when maxDepth is exceeded. Upstream uses `depth > maxDepth`, so a
	// node at exactly `maxDepth` still gets inspected.
	if depth > maxDepth {
		return false
	}

	switch node.Kind {
	// JSXText — upstream's `node.type === 'JSXText' && !!tryTrim(node.value)`.
	// tsgo splits text into two kinds; both carry the raw text on `.Text`.
	case ast.KindJsxText, ast.KindJsxTextAllWhiteSpaces:
		if strings.TrimSpace(node.AsJsxText().Text) != "" {
			return true
		}
		return false
	// Literal text — upstream's `node.type === 'Literal' && !!tryTrim(node.value)`.
	// tsgo splits the ESTree `Literal` across several `Kind*Literal` kinds.
	// Only string-shaped literals can carry a textual label; other literal
	// kinds (NumericLiteral, etc.) are conceivable as JSX children only in
	// edge cases and fall to the no-recursion default.
	case ast.KindStringLiteral:
		if strings.TrimSpace(node.AsStringLiteral().Text) != "" {
			return true
		}
		return false
	case ast.KindNoSubstitutionTemplateLiteral:
		if strings.TrimSpace(node.AsNoSubstitutionTemplateLiteral().Text) != "" {
			return true
		}
		return false
	// JSXExpressionContainer — upstream returns true unconditionally, even
	// for `{undefined}`. We mirror; the only sound static analysis would
	// require full constant folding which jsx-ast-utils does not perform.
	case ast.KindJsxExpression:
		return true
	}

	// Labelling-prop check on the opening element. Upstream's `node.openingElement
	// && hasLabellingProp(...)` only applies to nodes that have an opening
	// element (JSXElement / JSXFragment-with-opening — but ESTree fragments
	// don't carry one). In tsgo, only JsxElement / JsxSelfClosingElement
	// expose an opening element.
	opening, openingAttrs := openingElementOf(node)
	if opening != nil && hasLabellingProp(openingAttrs, labelAttributes) {
		return true
	}

	// Empty-children React-component fallback. Upstream:
	//
	//   if (node.type === 'JSXElement' && node.children.length === 0 && node.openingElement) {
	//     const name = getElementType(node.openingElement);
	//     const isReactComponent = name.length > 0 && name[0] === name[0].toUpperCase();
	//     if (isReactComponent && !controlComponents.some((c) => minimatch(name, c))) {
	//       return true;
	//     }
	//   }
	//
	// Self-closing forms (`<Foo />`) match this branch upstream because
	// ESTree models them as JSXElement with `selfClosing: true` and empty
	// `children`. In tsgo, both KindJsxElement-with-empty-children and
	// KindJsxSelfClosingElement satisfy "JSXElement with no children".
	//
	// NOTE: `controlComponents` here is matched via `minimatch` — glob
	// patterns are supported. This is intentionally asymmetric with the
	// top-level trigger which uses exact `indexOf`. Mirror both.
	if opening != nil && childrenIsEmpty(node) {
		name := getElementType(opening)
		if isReactComponentName(name) && !anyMinimatch(name, controlComponents) {
			return true
		}
	}

	// Recurse into children. `reactutil.GetJsxChildren` covers both
	// JsxElement and JsxFragment (upstream treats fragments transparently
	// here — the switch in `checkElement` has no case for them, but the
	// `node.children` loop still walks them).
	for _, child := range reactutil.GetJsxChildren(node) {
		if mayHaveAccessibleLabel(child, depth+1, maxDepth, labelAttributes, controlComponents, getElementType) {
			return true
		}
	}
	return false
}

// childrenIsEmpty reports whether the JSX-element-like node has zero
// children. Self-closing elements always satisfy this; paired elements
// inspect their Children list.
func childrenIsEmpty(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindJsxElement:
		j := node.AsJsxElement()
		return j.Children == nil || len(j.Children.Nodes) == 0
	case ast.KindJsxSelfClosingElement:
		return true
	}
	return false
}

// isReactComponentName mirrors upstream's
//
//	name.length > 0 && name[0] === name[0].toUpperCase()
//
// The JS check is true for every first character EXCEPT lowercase letters
// (`'a'.toUpperCase()` is `'A'`, which is not strict-equal to `'a'`). For
// ASCII this is exactly `c < 'a' || c > 'z'` — uppercase letters, digits,
// `$`, `_`, and other symbols all satisfy `c === c.toUpperCase()` and thus
// classify as "React component" upstream. Aligning with upstream rather
// than the stricter "uppercase letter only" check matters for tags like
// `<_Custom />` / `<$Foo />` (rare but legal as React component-like
// expressions) — upstream applies the fallback to them too.
//
// Non-ASCII first characters fall under the same byte-level rule. JSX tag
// names are ASCII identifiers in practice, so the byte-level check is
// faithful to upstream's intent within that domain.
func isReactComponentName(name string) bool {
	if name == "" {
		return false
	}
	c := name[0]
	return c < 'a' || c > 'z'
}

// anyMinimatch returns true iff at least one pattern in `patterns` matches
// `name` per minimatch (via `reactutil.MatchGlob` — same glob engine used
// by `react/jsx-handler-names` and other ports). Mirrors upstream's
// `controlComponents.some((c) => minimatch(name, c))`.
func anyMinimatch(name string, patterns []string) bool {
	for _, pattern := range patterns {
		if reactutil.MatchGlob(name, pattern) {
			return true
		}
	}
	return false
}

// hasLabellingProp mirrors upstream's `hasLabellingProp(openingElement,
// additionalLabellingProps)`:
//
//	const labellingProps = ['alt', 'aria-label', 'aria-labelledby',
//	                       ...additionalLabellingProps];
//	return openingElement.attributes.some((attribute) => {
//	  if (attribute.type !== 'JSXAttribute') return true;  // spread → opaque
//	  if (labellingProps.includes(propName(attribute))
//	      && !!tryTrim(getPropValue(attribute))) return true;
//	  return false;
//	});
//
// Key semantics preserved:
//
//   - Spread attributes (`{...props}` and literal-resolvable variants like
//     `{...{title: 'x'}}`) ALWAYS count as labelling, regardless of
//     content. Upstream's first conditional short-circuits the entire
//     `some()` to true when any attribute isn't a JSXAttribute.
//   - The prop-name match is CASE-SENSITIVE — `includes()` uses `===`. So
//     `<X ALT="x" />` does not match `alt`. Upstream's `propName(attribute)`
//     returns the attribute's name verbatim (no casing).
//   - The value check is `tryTrim(getPropValue(attribute))` truthy — for
//     string values this is "non-empty after trim"; for non-string values
//     (boolean true from the boolean attribute form, numeric, synthesized
//     non-empty strings from Identifier / Call / Member) it's the raw
//     truthy check.
func hasLabellingProp(openingAttrs []*ast.Node, additionalLabellingProps []string) bool {
	for _, attr := range openingAttrs {
		// Spread (JsxSpreadAttribute) — upstream's "type !== 'JSXAttribute'"
		// arm returns true unconditionally. Even literal-resolvable spreads
		// (`{...{title: 'x'}}`) count, because upstream's `hasLabellingProp`
		// short-circuits before inspecting the spread's contents.
		if attr.Kind == ast.KindJsxSpreadAttribute {
			return true
		}
		if attr.Kind != ast.KindJsxAttribute {
			continue
		}
		name := reactutil.GetJsxPropName(attr)
		if !isLabellingPropName(name, additionalLabellingProps) {
			continue
		}
		if labellingValueIsPresent(attr) {
			return true
		}
	}
	return false
}

// isLabellingPropName performs the case-sensitive `===` check against the
// labelling-prop names. The base set is the upstream-fixed
// `['alt', 'aria-label', 'aria-labelledby']`; `additional` carries the
// user-configured `labelAttributes`.
func isLabellingPropName(name string, additional []string) bool {
	if name == "alt" || name == "aria-label" || name == "aria-labelledby" {
		return true
	}
	for _, n := range additional {
		if name == n {
			return true
		}
	}
	return false
}

// labellingValueIsPresent mirrors upstream's `!!tryTrim(getPropValue(attr))`.
//
// `tryTrim` only trims strings — non-string values pass through unchanged,
// so the final `!!` applies the JS truthiness rule. We split the
// implementation into two paths:
//
//  1. Boolean-attribute form (`<X aria-label />`) — upstream's `extractValue`
//     null-attribute-value path returns boolean `true`; `tryTrim(true)` is
//     `true`; `!!true` → true.
//  2. JsxExpression / direct-string-literal value — attempt to extract as a
//     string via `PropStaticStringValue`. When that succeeds the value
//     IS a string, and we apply the trim. When it fails (the value is a
//     non-string: boolean false/true, number, null, undefined, etc.), fall
//     back to `PropValueIsTruthy` which encodes upstream's full `getPropValue`
//     truthiness.
//
// Examples covered:
//   - `aria-label="Save"`     → "Save"  → trim → "Save"  → truthy → true
//   - `aria-label=""`         → ""      → trim → ""       → falsy  → false
//   - `aria-label="  "`       → "  "    → trim → ""       → falsy  → false
//   - `aria-label={"x"}`      → "x"     → truthy → true
//   - `aria-label={"  "}`     → "  "    → trim → ""       → falsy  → false
//   - `aria-label`            → boolean true (boolean form)          → true
//   - `aria-label={someVar}`  → Identifier → staticEval → "someVar"  → true
//   - `aria-label={true}`     → boolean true                         → true
//   - `aria-label={false}`    → boolean false → falsy                → false
//   - `aria-label={null}`     → null → falsy                         → false
//   - `aria-label={undefined}`→ undefined → falsy                    → false
func labellingValueIsPresent(attr *ast.Node) bool {
	if jsxa11yutil.AttributeIsBooleanForm(attr) {
		return true
	}
	if s, ok := jsxa11yutil.PropStaticStringValue(attr); ok {
		return strings.TrimSpace(s) != ""
	}
	return jsxa11yutil.PropValueIsTruthy(attr)
}

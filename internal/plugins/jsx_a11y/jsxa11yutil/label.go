package jsxa11yutil

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
)

// OpeningElementOf returns the focal "opening element" used for label-prop
// inspection and reporting, plus that element's attribute list.
//
// In ESTree, every paired JSXElement and every self-closing JSXElement
// (`<X />`) exposes a `node.openingElement`. tsgo splits these into two
// distinct kinds:
//
//   - KindJsxElement carries a nested KindJsxOpeningElement that owns the
//     tag name and attributes; `node` itself owns the children list.
//   - KindJsxSelfClosingElement carries the tag name and attributes
//     directly on the node — there is no nested opening element.
//
// To present the same "opening element + attributes" surface regardless
// of form, this helper returns:
//
//   - For KindJsxElement: the inner KindJsxOpeningElement and its
//     attribute list.
//   - For KindJsxSelfClosingElement: the node itself (it plays both
//     roles) and its attribute list.
//   - For any other kind (or nil): (nil, nil).
//
// Callers that report diagnostics on the opening element should use the
// first return value as `ReportNode` target; callers that inspect props
// should use the second.
func OpeningElementOf(node *ast.Node) (*ast.Node, []*ast.Node) {
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

// JsxElementHasNoChildren reports whether a JSX-element-like node has zero
// children. Self-closing elements always satisfy this; paired elements
// inspect their Children list.
//
// Mirrors upstream's `node.children.length === 0` check used inside
// `mayHaveAccessibleLabel`'s empty-React-component fallback. Self-closing
// elements (`<Foo />`) match this branch upstream because ESTree models
// them as JSXElement with `selfClosing: true` and empty `children`.
func JsxElementHasNoChildren(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindJsxElement:
		j := node.AsJsxElement()
		return j.Children == nil || len(j.Children.Nodes) == 0
	case ast.KindJsxSelfClosingElement:
		return true
	}
	return false
}

// IsReactComponentName mirrors upstream's
//
//	name.length > 0 && name[0] === name[0].toUpperCase()
//
// JS `String.prototype.toUpperCase` is Unicode-aware (Unicode Default Case
// Conversion). To stay byte-faithful with `c === c.toUpperCase()` across
// non-ASCII tag names, we decode the first rune and compare against
// `unicode.ToUpper(r)`:
//
//   - ASCII uppercase letters / digits / `$` / `_` / other symbols satisfy
//     `ToUpper(r) == r` → classified as React component (matches JS).
//   - ASCII lowercase letters: `ToUpper('a') = 'A' != 'a'` → not a React
//     component (matches JS).
//   - Non-ASCII lowercase like `'é'` / `'à'` / `'ı'`:
//     `ToUpper('é') = 'É' != 'é'` → not a React component (matches JS's
//     `'é'.toUpperCase() === 'É'`). A byte-level check on `name[0]`
//     would wrongly classify these as React component because the leading
//     UTF-8 byte (e.g. 0xC3 for `'é'`) is outside `'a'..'z'`.
//   - Non-ASCII characters without case (CJK, symbols, etc.):
//     `ToUpper('中') = '中'` → classified as React component (matches JS).
//
// Locale-sensitive Turkish dotted I (`'İ'` ↔ `'i'`): both Go's
// `unicode.ToUpper` and JS's default `toUpperCase` use the locale-
// independent Unicode mapping, so behavior aligns.
func IsReactComponentName(name string) bool {
	if name == "" {
		return false
	}
	r, _ := utf8.DecodeRuneInString(name)
	return unicode.ToUpper(r) == r
}

// AnyMinimatch returns true iff at least one pattern in `patterns` matches
// `name` per minimatch (via `reactutil.MatchGlob` — same glob engine used
// by `react/jsx-handler-names` and other ports). Mirrors upstream's
// `patterns.some((p) => minimatch(name, p))` predicate. Returns false on
// an empty patterns slice and on an empty name (MatchGlob's empty-text
// guard).
func AnyMinimatch(name string, patterns []string) bool {
	for _, pattern := range patterns {
		if reactutil.MatchGlob(name, pattern) {
			return true
		}
	}
	return false
}

// HasLabellingProp mirrors upstream's `hasLabellingProp(openingElement,
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
//
// Used by `control-has-associated-label`, `label-has-associated-control`,
// and any future a11y rule whose upstream form calls `hasLabellingProp`.
func HasLabellingProp(openingAttrs []*ast.Node, additionalLabellingProps []string) bool {
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
	if AttributeIsBooleanForm(attr) {
		return true
	}
	if s, ok := PropStaticStringValue(attr); ok {
		return strings.TrimSpace(s) != ""
	}
	return PropValueIsTruthy(attr)
}

// MayHaveAccessibleLabel mirrors upstream's
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
// caller passes `depth=0` for the root. Children of a JsxFragment are
// recursed into transparently — upstream's switch has no JSXFragment case
// for the label-text branches, but the children loop walks fragments.
func MayHaveAccessibleLabel(node *ast.Node, depth, maxDepth int, labelAttributes, controlComponents []string, getElementType func(*ast.Node) string) bool {
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
		return strings.TrimSpace(node.AsJsxText().Text) != ""
	// Literal text — upstream's `node.type === 'Literal' && !!tryTrim(node.value)`.
	// tsgo splits the ESTree `Literal` across several `Kind*Literal` kinds.
	// Only string-shaped literals can carry a textual label; other literal
	// kinds (NumericLiteral, etc.) are conceivable as JSX children only in
	// edge cases and fall to the no-recursion default.
	case ast.KindStringLiteral:
		return strings.TrimSpace(node.AsStringLiteral().Text) != ""
	case ast.KindNoSubstitutionTemplateLiteral:
		return strings.TrimSpace(node.AsNoSubstitutionTemplateLiteral().Text) != ""
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
	opening, openingAttrs := OpeningElementOf(node)
	if opening != nil && HasLabellingProp(openingAttrs, labelAttributes) {
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
	// NOTE: `controlComponents` here is matched via `minimatch` — glob
	// patterns are supported. This is intentionally asymmetric with the
	// `control-has-associated-label` top-level trigger which uses exact
	// `indexOf`. Mirror both there.
	if opening != nil && JsxElementHasNoChildren(node) {
		name := getElementType(opening)
		if IsReactComponentName(name) && !AnyMinimatch(name, controlComponents) {
			return true
		}
	}

	// Recurse into children. `reactutil.GetJsxChildren` covers both
	// JsxElement and JsxFragment (upstream treats fragments transparently
	// here — the switch in `checkElement` has no case for them, but the
	// `node.children` loop still walks them).
	for _, child := range reactutil.GetJsxChildren(node) {
		if MayHaveAccessibleLabel(child, depth+1, maxDepth, labelAttributes, controlComponents, getElementType) {
			return true
		}
	}
	return false
}

// MayContainChildComponent mirrors upstream's
//
//	mayContainChildComponent(root, componentName, maxDepth = 1,
//	                         elementType = rawElementType)
//
// — a DFS that returns true if any descendant within `maxDepth` levels is:
//
//   - A JsxExpression (`{expr}`) — assumed to render the target component.
//     This is upstream's "best we can do" given expression containers are
//     opaque to static analysis.
//   - A JsxElement / JsxSelfClosingElement whose resolved `elementType`
//     matches `componentName` per minimatch (so glob patterns like
//     `Custom*` work).
//
// The traversal starts at `root` with depth=1 (matching upstream's
// `traverseChildren(root, 1)` seed), so children of `root` are visited at
// depth 1, grandchildren at depth 2, etc. Upstream bails when
// `depth > maxDepth`, so a node at exactly `maxDepth` is still inspected.
//
// JsxFragment children are walked transparently — upstream's `node.children`
// loop has no fragment special-casing. JsxText / StringLiteral / other
// non-element children that aren't JsxExpression and don't match the
// component check fall through to the recursion (which finds nothing).
//
// Used by `label-has-associated-control` to detect a nested form control
// (`input`, `meter`, `output`, `progress`, `select`, `textarea`, plus any
// user-configured `controlComponents`) inside a `<label>`.
func MayContainChildComponent(root *ast.Node, componentName string, maxDepth int, getElementType func(*ast.Node) string) bool {
	return mayContainChildComponentWalk(root, componentName, 1, maxDepth, getElementType)
}

func mayContainChildComponentWalk(node *ast.Node, componentName string, depth, maxDepth int, getElementType func(*ast.Node) string) bool {
	if node == nil {
		return false
	}
	// Upstream's bail: `if (depth > maxDepth) return false;` — checked
	// before any child inspection at this level.
	if depth > maxDepth {
		return false
	}
	for _, child := range reactutil.GetJsxChildren(node) {
		// JsxExpression — upstream's `child.type === 'JSXExpressionContainer'`
		// short-circuits to true. Even `{undefined}` counts; static analysis
		// can't disprove the child renders the target component.
		if child.Kind == ast.KindJsxExpression {
			return true
		}
		// JSXElement match. Upstream gates on
		// `child.type === 'JSXElement' && child.openingElement &&
		// minimatch(elementType(child.openingElement), componentName)`.
		// tsgo splits the form: paired KindJsxElement carries
		// `OpeningElement`, self-closing KindJsxSelfClosingElement plays
		// both roles. Both are "JSXElement with an opening element" under
		// upstream's model, so we check both.
		switch child.Kind {
		case ast.KindJsxElement:
			opening := child.AsJsxElement().OpeningElement
			if opening != nil && reactutil.MatchGlob(getElementType(opening), componentName) {
				return true
			}
		case ast.KindJsxSelfClosingElement:
			if reactutil.MatchGlob(getElementType(child), componentName) {
				return true
			}
		}
		if mayContainChildComponentWalk(child, componentName, depth+1, maxDepth, getElementType) {
			return true
		}
	}
	return false
}

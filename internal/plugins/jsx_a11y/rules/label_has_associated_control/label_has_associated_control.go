// Package label_has_associated_control ports eslint-plugin-jsx-a11y's
// `label-has-associated-control` rule. Enforces that any `<label>` (or
// configured custom label component) has an accessible text label AND an
// associated form control — either via the `htmlFor` attribute, via a
// nested form-control descendant, or both (configurable via the `assert`
// option).
//
// Upstream listener (`JSXElement`, executed in this order):
//
//  1. Resolve the JSX tag name through `getElementType(context)` (which
//     consults `settings['jsx-a11y'].polymorphicPropName` and
//     `settings['jsx-a11y'].components`). Compare against
//     `['label'].concat(options.labelComponents)` via minimatch. If no
//     match, skip — the rule does not apply.
//  2. Compute:
//     - `hasHtmlFor` — true iff any attribute in
//     `settings['jsx-a11y'].attributes.for ?? ['htmlFor']` is present
//     AND its value is truthy under jsx-ast-utils' `getPropValue`.
//     - `hasNestedControl` — true iff any of the builtin form-control
//     names (`input`, `meter`, `output`, `progress`, `select`,
//     `textarea`) OR a user-configured `controlComponents` entry
//     matches a descendant within `min(options.depth ?? 2, 25)` levels
//     (per `mayContainChildComponent`'s minimatch / glob semantics).
//     - `hasAccessibleLabel` — true iff `mayHaveAccessibleLabel` finds a
//     text label within the same depth budget (also threading the
//     `labelAttributes` option and the builtin-`controlComponents`-plus-
//     user-list).
//  3. If `!hasAccessibleLabel`, report `accessibleLabel` and return.
//  4. Otherwise branch on `options.assert ?? 'either'`:
//     - `'htmlFor'`: report `htmlFor` if `!hasHtmlFor`.
//     - `'nesting'`: report `nesting` if `!hasNestedControl`.
//     - `'both'`: report `both` if `!hasHtmlFor || !hasNestedControl`.
//     - `'either'` (default): report `either` if neither holds.
//     - any other value: normalized to `'either'` in option parsing
//     (ESLint schema rejects + re-invokes with defaults; rslint emulates).
//
// All reports target the opening element (paired `<label>` → its
// `OpeningElement`; self-closing `<label />` → the node itself). Message
// strings mirror upstream `errorMessages` verbatim.
package label_has_associated_control

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/jsxa11yutil"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Message strings mirror upstream's `errorMessages` constant verbatim.
const (
	msgAccessibleLabel = "A form label must have accessible text."
	msgHtmlFor         = "A form label must have a valid htmlFor attribute."
	msgNesting         = "A form label must have an associated control as a descendant."
	msgEither          = "A form label must either have a valid htmlFor attribute or a control as a descendant."
	msgBoth            = "A form label must have a valid htmlFor attribute and a control as a descendant."
)

const (
	idAccessibleLabel = "accessibleLabel"
	idHtmlFor         = "htmlFor"
	idNesting         = "nesting"
	idEither          = "either"
	idBoth            = "both"
)

// defaultDepth mirrors upstream's `options.depth === undefined ? 2 : options.depth`.
const defaultDepth = 2

// maxDepthCap mirrors upstream's `Math.min(..., 25)` ceiling on the recursion
// budget — protects against pathological JSX trees regardless of user config.
const maxDepthCap = 25

// defaultAssert mirrors upstream's `options.assert || 'either'`. Non-enum
// values are normalized to this default by parseOptions (matches upstream
// ESLint's schema-fallback behavior; see parseOptions for verification).
const defaultAssert = "either"

// builtinControlComponents mirrors upstream's hard-coded form-control names.
// `controlComponents` from user config is APPENDED to this list, not
// substituted — so the builtins are always considered nested controls.
var builtinControlComponents = []string{"input", "meter", "output", "progress", "select", "textarea"}

// defaultHtmlForAttributes mirrors upstream's
// `settings['jsx-a11y']?.attributes?.for ?? ['htmlFor']` fallback. The
// settings list (when provided) REPLACES this default — upstream's `??`
// short-circuits at the first non-nullish value, so the user list does
// not have to include `htmlFor` to keep it active.
var defaultHtmlForAttributes = []string{"htmlFor"}

type options struct {
	labelComponents   []string
	labelAttributes   []string
	controlComponents []string
	assert            string
	depth             int
}

func parseOptions(raw any) options {
	opts := options{
		assert: defaultAssert,
		depth:  defaultDepth,
	}
	m := utils.GetOptionsMap(raw)
	if m == nil {
		return opts
	}
	opts.labelComponents = jsxa11yutil.StringSliceOption(m["labelComponents"])
	opts.labelAttributes = jsxa11yutil.StringSliceOption(m["labelAttributes"])
	opts.controlComponents = jsxa11yutil.StringSliceOption(m["controlComponents"])
	// `assert` must be one of `htmlFor` / `nesting` / `both` / `either` per
	// upstream's JSON schema enum. ESLint's schema rejects other values at
	// the framework layer, then re-invokes the rule with default options
	// (so the user observes a config-error AND `either`-mode behavior).
	// rslint has no schema layer, so we mirror that observable behavior
	// here: only enum values overwrite the default; any other string falls
	// back to `'either'`. Non-string values are likewise ignored.
	//
	// Verified against `npx eslint` on eslint-plugin-jsx-a11y@main:
	//
	//   {assert: "garbage"} on `<label>Save</label>` → "either" error
	//   {assert: ""}        on `<label>Save</label>` → "either" error
	//
	// upstream's literal `switch default: break` is dead code under the
	// schema, but ESLint's fallback path lands at `'either'` because
	// `options.assert || 'either'` reads the missing key from the default
	// options object.
	if v, ok := m["assert"].(string); ok {
		switch v {
		case "htmlFor", "nesting", "both", "either":
			opts.assert = v
		}
	}
	// `depth` upstream: `options.depth === undefined ? 2 : options.depth`,
	// then `Math.min(depth, 25)`. JSON numbers decode as float64; absence
	// of the key keeps the default. Negative values are NOT clamped here
	// because upstream's `Math.min(-1, 25)` is `-1` — every `depth > maxDepth`
	// check then bails immediately, producing the same behavior as
	// "no children visited" for the trigger heuristics (hasNestedControl
	// stays false; hasAccessibleLabel inspects only the root). Mirror
	// upstream's lack of clamping on the low end.
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

// LabelHasAssociatedControlRule is the rule definition exported to the
// plugin registry. The Name matches the rslint convention — the
// `jsx-a11y/` prefix is part of the registered key.
var LabelHasAssociatedControlRule = rule.Rule{
	Name: "jsx-a11y/label-has-associated-control",
	Run: func(ctx rule.RuleContext, _rawOptions []any) rule.RuleListeners {
		rawOptions := rule.UnwrapOptions(_rawOptions)
		opts := parseOptions(rawOptions)

		// `labelComponentNames = ['label'].concat(labelComponents)` upstream.
		// The order matters only for which name is reported first on a
		// debugging trace; for `.some(minimatch)` the result is identical.
		labelComponentNames := make([]string, 0, 1+len(opts.labelComponents))
		labelComponentNames = append(labelComponentNames, "label")
		labelComponentNames = append(labelComponentNames, opts.labelComponents...)

		// `controlComponents = [].concat('input', 'meter', 'output',
		// 'progress', 'select', 'textarea', options.controlComponents || [])`.
		// Builtin form-control tags are always present; the user list is
		// appended, NOT substituted. Used in two places: the
		// `hasNestedControl` minimatch loop AND `mayHaveAccessibleLabel`'s
		// empty-React-component fallback (where it filters out controls that
		// shouldn't classify as labels).
		controlComponents := make([]string, 0, len(builtinControlComponents)+len(opts.controlComponents))
		controlComponents = append(controlComponents, builtinControlComponents...)
		controlComponents = append(controlComponents, opts.controlComponents...)

		// `htmlForAttributes = settings['jsx-a11y']?.attributes?.for ?? ['htmlFor']`.
		// The user list REPLACES the default when set — upstream's `??`
		// short-circuits at the first non-nullish value, so a user list
		// that omits `htmlFor` will disable detection of the `htmlFor`
		// attribute entirely. Mirror that.
		htmlForAttributes := resolveHtmlForAttributes(ctx.Settings)

		getElementType := func(node *ast.Node) string {
			return jsxa11yutil.GetElementType(node, ctx.Settings)
		}

		check := func(elementNode *ast.Node) {
			opening, openingAttrs := jsxa11yutil.OpeningElementOf(elementNode)
			if opening == nil {
				return
			}
			tag := getElementType(opening)

			// Listener gate: the resolved tag name must match one of the
			// `labelComponentNames` patterns via minimatch (NOT exact
			// equality — upstream's `.some(name => minimatch(tag, name))`
			// supports glob entries like `*Label` or `????Label`). The
			// builtin `'label'` is also passed through minimatch, which
			// for a plain lowercase string degenerates to exact equality.
			if !jsxa11yutil.AnyMinimatch(tag, labelComponentNames) {
				return
			}

			// `hasHtmlFor`: upstream's `validateHtmlFor`. Walk the
			// settings-defined list in order; on the FIRST present attribute
			// (case-insensitive, matching jsx-ast-utils' default
			// `ignoreCase: true`), return whether its value is truthy under
			// `getPropValue`. Absent attribute → keep looking; first present
			// attribute decides the result. Mirror exactly — do not OR the
			// results across attributes.
			hasHtmlFor := validateHtmlFor(openingAttrs, htmlForAttributes)

			// `hasNestedControl`: upstream's `controlComponents.some((name)
			// => mayContainChildComponent(node, name, recursionDepth,
			// elementType))`. Each name is glob-matched against descendant
			// tag names (so user-configured patterns like `Custom*` work).
			// `mayContainChildComponent` seeds with depth=1, so children of
			// the root are visited at depth 1 — a `recursionDepth=0` user
			// override therefore bails before the first child.
			hasNestedControl := false
			for _, name := range controlComponents {
				if jsxa11yutil.MayContainChildComponent(elementNode, name, opts.depth, getElementType) {
					hasNestedControl = true
					break
				}
			}

			// `hasAccessibleLabel`: upstream's `mayHaveAccessibleLabel(node,
			// recursionDepth, options.labelAttributes, elementType,
			// controlComponents)`. The depth here measures recursion from
			// the root (depth=0); a `recursionDepth=0` user override
			// inspects only the root node and any of its labelling props,
			// without descending.
			hasAccessibleLabel := jsxa11yutil.MayHaveAccessibleLabel(elementNode, 0, opts.depth, opts.labelAttributes, controlComponents, getElementType)

			if !hasAccessibleLabel {
				ctx.ReportNode(opening, rule.RuleMessage{
					Id:          idAccessibleLabel,
					Description: msgAccessibleLabel,
				})
				return
			}

			switch opts.assert {
			case idHtmlFor:
				if !hasHtmlFor {
					ctx.ReportNode(opening, rule.RuleMessage{
						Id:          idHtmlFor,
						Description: msgHtmlFor,
					})
				}
			case idNesting:
				if !hasNestedControl {
					ctx.ReportNode(opening, rule.RuleMessage{
						Id:          idNesting,
						Description: msgNesting,
					})
				}
			case idBoth:
				if !hasHtmlFor || !hasNestedControl {
					ctx.ReportNode(opening, rule.RuleMessage{
						Id:          idBoth,
						Description: msgBoth,
					})
				}
			case idEither:
				if !hasHtmlFor && !hasNestedControl {
					ctx.ReportNode(opening, rule.RuleMessage{
						Id:          idEither,
						Description: msgEither,
					})
				}
				// default: silently no-op — upstream's `default: break;`.
				// Unknown `assert` values do not raise; valid values are
				// already enforced by JSON schema on the upstream side.
			}
		}

		// tsgo splits paired and self-closing JSX into separate kinds; ESTree
		// merges them under one JSXElement. Listen on both so `<label />`
		// and `<label>…</label>` are classified independently.
		return rule.RuleListeners{
			ast.KindJsxElement:            check,
			ast.KindJsxSelfClosingElement: check,
		}
	},
}

// resolveHtmlForAttributes mirrors upstream's
//
//	settings['jsx-a11y']?.attributes?.for ?? ['htmlFor']
//
// Note: upstream's `??` only falls back when the user value is `null` /
// `undefined`. An explicit empty array (`[]`) is preserved and silences the
// htmlFor check entirely — we mirror that by returning an empty (non-nil)
// slice in that case. Non-string entries (defensive) are dropped.
func resolveHtmlForAttributes(settings map[string]interface{}) []string {
	a11y, _ := settings["jsx-a11y"].(map[string]interface{})
	if a11y == nil {
		return defaultHtmlForAttributes
	}
	attrs, _ := a11y["attributes"].(map[string]interface{})
	if attrs == nil {
		return defaultHtmlForAttributes
	}
	raw, ok := attrs["for"]
	if !ok || raw == nil {
		return defaultHtmlForAttributes
	}
	arr, ok := raw.([]interface{})
	if !ok {
		return defaultHtmlForAttributes
	}
	out := make([]string, 0, len(arr))
	for _, item := range arr {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

// validateHtmlFor mirrors upstream's `validateHtmlFor`:
//
//	for (let i = 0; i < htmlForAttributes.length; i += 1) {
//	  const attribute = htmlForAttributes[i];
//	  if (hasProp(node.attributes, attribute)) {
//	    const htmlForAttr = getProp(node.attributes, attribute);
//	    const htmlForValue = getPropValue(htmlForAttr);
//	    return htmlForAttr !== false && !!htmlForValue;
//	  }
//	}
//	return false;
//
// Two jsx-ast-utils helpers with DIFFERENT defaults are composed here, and
// the asymmetry matters for spread attributes. We mirror it exactly:
//
//  1. `hasProp` default `{ spreadStrict: true, ignoreCase: true }` — spread
//     attributes (even literal-resolvable `{...{htmlFor: "x"}}`) are
//     SKIPPED for the presence check. The loop falls through to the next
//     `attributes.for` name as if the attribute were absent.
//
//  2. `getProp` has no `spreadStrict` option — it ALWAYS walks literal
//     ObjectExpression spreads, and uses `Array.prototype.find` to pick
//     the FIRST matching attribute in source order. So when a `direct`
//     and a literal-spread both carry the same prop name, source order
//     decides which value is taken — and the picked value may come from
//     either side. The synthetic JSXAttribute returned for a spread
//     match is then fed to `getPropValue` for truthiness.
//
// Concrete cases this implementation must align on:
//
//   - `<label {...{htmlFor: "y"}} htmlFor="">` — hasProp(strict)=true via
//     the direct `htmlFor=""`. getProp scans in order, the spread is first
//     and carries htmlFor → returns the spread's literal "y" → truthy.
//   - `<label {...{htmlFor: ""}} htmlFor="x">` — hasProp(strict)=true. getProp
//     finds the spread first → value is "" → falsy.
//   - `<label {...{htmlFor: "y"}}>` (spread only, no direct) — hasProp(strict)
//     =false → fall through; the loop returns false even though getProp
//     would have found a value.
//   - `<label htmlFor="x" {...{htmlFor: "y"}}>` — direct first, getProp
//     returns the direct → "x" → truthy.
//
// Implementation: scan once with strict semantics (direct only) to decide
// `hasProp`. If present, use `FindAttributeByName` — which walks direct +
// literal-spread in source order — to mirror `getProp`. Hand the result
// to `PropValueIsTruthy` which encodes upstream's full `getPropValue`
// truthiness extractor.
func validateHtmlFor(openingAttrs []*ast.Node, htmlForAttributes []string) bool {
	for _, name := range htmlForAttributes {
		// (1) hasProp(spreadStrict: true): does ANY direct JsxAttribute
		// carry this name? If not, the attribute is absent and the loop
		// must continue to the next configured name — even if a literal
		// spread carries it.
		hasDirect := false
		for _, a := range openingAttrs {
			if a.Kind != ast.KindJsxAttribute {
				continue
			}
			if strings.EqualFold(reactutil.GetJsxPropName(a), name) {
				hasDirect = true
				break
			}
		}
		if !hasDirect {
			continue
		}
		// (2) getProp (no spreadStrict): source-order find across direct
		// + literal-spread. `FindAttributeByName` already implements this
		// — for a spread match it returns the inner PropertyAssignment /
		// ShorthandPropertyAssignment node, which `PropValueIsTruthy`
		// resolves via `attributeInnerExpression` + staticEval.
		attr := jsxa11yutil.FindAttributeByName(openingAttrs, name)
		if attr == nil {
			// Defensive: hasDirect implied a match, so FindAttributeByName
			// should also find one. If it ever doesn't (e.g. a future
			// helper change), treat as absent and try the next name.
			continue
		}
		return jsxa11yutil.PropValueIsTruthy(attr)
	}
	return false
}

// Package no_noninteractive_element_interactions ports
// eslint-plugin-jsx-a11y's `no-noninteractive-element-interactions` rule.
// The rule enforces that visible, non-interactive DOM elements do not carry
// mouse / keyboard / focus / image event handlers — assigning handlers to a
// non-interactive element (e.g. `<li onClick={…} />`, `<div role="article"
// onClick={…} />`) is a strong signal of misuse of the platform a11y
// semantics, since screen-reader / keyboard users get no productive
// interaction once focus lands on the element.
//
// Upstream signature:
//
//	options: {
//	  handlers?:        string[]            (default: focus + image + keyboard + mouse)
//	  <element-name>?:  string[]            (per-element handler allow-list)
//	}
//
// `handlers` (when present, even when explicitly `[]`) overrides the default
// interactive-handler list. Any OTHER key is interpreted as a per-element
// allow-list: a JsxAttribute whose name appears in `config[type]` is
// filtered out of the trigger check, so e.g. recommended's
// `iframe: ['onError', 'onLoad']` makes `<iframe onLoad={…} />` valid while
// `<iframe onClick={…} />` still reports.
//
// Trigger sequence — each predicate is checked in order against the JSX
// opening / self-closing element. Bail-outs return without reporting:
//
//  1. Element type not in aria-query's `dom` map → bail (custom
//     components; we don't second-guess what low-level DOM they render).
//  2. After per-element filter, no remaining `handlers` prop has a
//     non-nullish value → bail (`getPropValue == null` upstream covers
//     `prop={null}` / `prop={undefined}` / missing).
//  3. `contentEditable="true"` (RAW source equality — see
//     [jsxa11yutil.IsContentEditable]) → bail.
//  4. Element is hidden from screen readers (`aria-hidden={true}` or
//     `<input type="hidden">`) → bail.
//  5. Element role resolves to `presentation` / `none` → bail.
//  6. Element is inherently interactive, has an interactive role, is
//     NEITHER inherently nor explicitly non-interactive, or has an
//     ABSTRACT role → bail.
//
// Otherwise, report at the JSX opening element. The single message id is
// `noNoninteractiveElementInteractions`, mirroring upstream's single
// `errorMessage`.
package no_noninteractive_element_interactions

import (
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/jsxa11yutil"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// errorMessage mirrors upstream's `errorMessage` string verbatim.
const errorMessage = "Non-interactive elements should not be assigned mouse or keyboard event listeners."

// options captures the parsed configuration.
//
// Handlers carries the override for the default interactive-handler list
// (`config.handlers || defaultInteractiveProps`). A non-nil slice — INCLUDING
// the explicit empty slice — wins; nil signals "absent" so we fall back to
// the default. Mirrors JS `[] || x` truthiness (empty array is truthy → no
// fallback).
//
// Elements maps tag-name → handler-allow-list. Matches upstream's
// `hasOwn(config, type)` lookup: a JsxAttribute whose name is in
// `Elements[type]` is filtered out before the interactive-prop check.
type options struct {
	Handlers *[]string
	Elements map[string][]string
}

// defaultInteractiveProps mirrors upstream's
// `[].concat(eventHandlersByType.focus, eventHandlersByType.image,
//
//	eventHandlersByType.keyboard, eventHandlersByType.mouse)` — used when
//
// `options.handlers` is absent. Order matches upstream concat for
// auditability; semantically the slice acts as a set (lookup via .some()).
var defaultInteractiveProps = func() []string {
	out := make([]string, 0,
		len(jsxa11yutil.EventHandlersFocus)+
			len(jsxa11yutil.EventHandlersImage)+
			len(jsxa11yutil.EventHandlersKeyboard)+
			len(jsxa11yutil.EventHandlersMouse))
	out = append(out, jsxa11yutil.EventHandlersFocus...)
	out = append(out, jsxa11yutil.EventHandlersImage...)
	out = append(out, jsxa11yutil.EventHandlersKeyboard...)
	out = append(out, jsxa11yutil.EventHandlersMouse...)
	return out
}()

func parseOptions(raw any) options {
	opts := options{}
	m := utils.GetOptionsMap(raw)
	if m == nil {
		return opts
	}
	if rawHandlers, ok := m["handlers"]; ok {
		// Upstream `config.handlers || defaultInteractiveProps`. JS arrays
		// (including empty) are truthy, so we treat any present `[]string`
		// — even empty — as an override. StringSliceOption returns nil for
		// non-array values, in which case we fall through to the default
		// (mirrors JS's null/undefined being falsy).
		if parsed := jsxa11yutil.StringSliceOption(rawHandlers); parsed != nil {
			opts.Handlers = &parsed
		}
	}
	for key, v := range m {
		if key == "handlers" {
			continue
		}
		if parsed := jsxa11yutil.StringSliceOption(v); parsed != nil {
			if opts.Elements == nil {
				opts.Elements = make(map[string][]string)
			}
			opts.Elements[key] = parsed
		}
	}
	return opts
}

var NoNoninteractiveElementInteractionsRule = rule.Rule{
	Name: "jsx-a11y/no-noninteractive-element-interactions",
	Run: func(ctx rule.RuleContext, _rawOptions []any) rule.RuleListeners {
		rawOptions := rule.LegacyUnwrapOptions(_rawOptions)
		opts := parseOptions(rawOptions)
		interactiveProps := defaultInteractiveProps
		if opts.Handlers != nil {
			interactiveProps = *opts.Handlers
		}

		check := func(node *ast.Node) {
			rawAttrs := reactutil.GetJsxElementAttributes(node)
			elementType := jsxa11yutil.GetElementType(node, ctx.Settings)

			// Per-element allow-list filter. Upstream:
			//   attributes.filter(attr =>
			//     attr.type !== 'JSXSpreadAttribute' &&
			//     !includes(config[type], propName(attr)))
			// — drop a non-spread attribute when its propName is in the
			// element's allow-list. Spread attributes are always retained
			// (the allow-list cannot prove what they carry), so a literal
			// spread `{...{onLoad: foo}}` slips past the filter and trips
			// the trigger downstream. Mirror exactly.
			//
			// Upstream's downstream calls — `isContentEditable(type, attrs)`,
			// `isHiddenFromScreenReader(type, attrs)`,
			// `isPresentationRole(type, attrs)`, `isInteractive*` /
			// `isNonInteractive*` / `isAbstractRole(type, attrs)` — all
			// consume the FILTERED `attributes` variable. We mirror that
			// byte-for-byte by passing `attrs` to every downstream call
			// (including [IsHiddenFromScreenReaderFromTagAttrs], the
			// attrs-based variant added for exactly this case).
			attrs := rawAttrs
			if allow, ok := opts.Elements[elementType]; ok {
				filtered := make([]*ast.Node, 0, len(rawAttrs))
				for _, attr := range rawAttrs {
					if attr.Kind != ast.KindJsxAttribute {
						filtered = append(filtered, attr)
						continue
					}
					if slices.Contains(allow, reactutil.GetJsxPropName(attr)) {
						continue
					}
					filtered = append(filtered, attr)
				}
				attrs = filtered
			}

			// Upstream `interactiveProps.some(prop => hasProp(attrs, prop)
			// && getPropValue(getProp(attrs, prop)) != null)`. `hasProp` /
			// `getProp` default options walk LITERAL ObjectLiteral spreads,
			// so `<div {...{onClick: foo}} />` counts here — matches
			// [FindAttributeByName] semantics. `!= null` covers null and
			// undefined (and the empty `{}` JsxExpression that
			// `getPropValue` resolves to null for).
			hasInteractiveProps := false
			for _, prop := range interactiveProps {
				attr := jsxa11yutil.FindAttributeByName(attrs, prop)
				if attr == nil {
					continue
				}
				if jsxa11yutil.PropValueIsNullish(attr) {
					continue
				}
				hasInteractiveProps = true
				break
			}

			if !jsxa11yutil.IsDOMElement(elementType) {
				return
			}
			if !hasInteractiveProps {
				return
			}
			if jsxa11yutil.IsContentEditable(attrs) {
				return
			}
			if jsxa11yutil.IsHiddenFromScreenReaderFromTagAttrs(elementType, attrs) {
				return
			}
			if jsxa11yutil.IsPresentationRole(attrs) {
				return
			}

			// Upstream "no opinion" bail-out. Any ONE of these short-circuits
			// the report:
			//   - element is inherently interactive
			//   - element has an interactive role
			//   - element is NEITHER inherently nor explicitly non-interactive
			//     (the "we don't know enough to call this non-interactive" arm)
			//   - role is abstract
			if jsxa11yutil.IsInteractiveElement(elementType, attrs) ||
				jsxa11yutil.IsInteractiveRole(elementType, attrs) ||
				(!jsxa11yutil.IsNonInteractiveElement(elementType, attrs) &&
					!jsxa11yutil.IsNonInteractiveRole(elementType, attrs)) ||
				jsxa11yutil.IsAbstractRole(elementType, attrs) {
				return
			}

			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "noNoninteractiveElementInteractions",
				Description: errorMessage,
			})
		}

		return rule.RuleListeners{
			ast.KindJsxOpeningElement:     check,
			ast.KindJsxSelfClosingElement: check,
		}
	},
}

// Package anchor_is_valid ports eslint-plugin-jsx-a11y's `anchor-is-valid`
// rule. The rule enforces that <a> elements (and any configured custom
// components) function as proper hyperlinks — every anchor should either
// carry a valid, navigable href, or be replaced with a <button> when used
// as a click target.
//
// Three independent aspects can be enabled / disabled via options:
//
//   - noHref         — require an href-like prop to be present.
//   - invalidHref    — reject `""`, `"#"`, and `javascript:` URL strings.
//   - preferButton   — when an onClick is present, recommend <button>.
//
// All three are active by default. The rule reports at most one diagnostic
// per element; aspect ordering and the onClick interaction follow upstream
// (see the comments inside the listener for the exact precedence).
package anchor_is_valid

import (
	"regexp"
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/jsxa11yutil"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const (
	preferButtonErrorMessage = "Anchor used as a button. Anchors are primarily expected to navigate. Use the button element instead. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md"
	noHrefErrorMessage       = "The href attribute is required for an anchor to be keyboard accessible. Provide a valid, navigable address as the href value. If you cannot provide an href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md"
	invalidHrefErrorMessage  = "The href attribute requires a valid value to be accessible. Provide a valid, navigable address as the href value. If you cannot provide a valid href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md"
)

// jsHrefPattern mirrors upstream's `safeRegexTest(/^\W*?javascript:/)`. The
// lazy `\W*?` lets a stray non-word prefix (whitespace, `#`, etc.) precede
// `javascript:` and still match — this is what catches obfuscation attempts
// like ` javascript:void(0)` while still accepting plain identifiers
// such as `javascriptFoo`. Both JS and Go default `\W` to ASCII (`[^A-Za-z0-9_]`),
// so the matched character class is identical.
var jsHrefPattern = regexp.MustCompile(`^\W*?javascript:`)

const (
	aspectNoHref       = "noHref"
	aspectInvalidHref  = "invalidHref"
	aspectPreferButton = "preferButton"
)

// options mirrors the upstream JSON shape. Each field maps directly to the
// schema entry of the same name; defaults are applied at use-sites because
// upstream's per-aspect default depends on whether `aspects` was provided
// at all (an empty / absent `aspects` enables all three; an explicit list
// enables only the listed names).
type options struct {
	components  []string
	specialLink []string
	// activeAspects mirrors upstream's `activeAspects[name] = aspects.indexOf(name) !== -1`
	// after defaulting. We pre-compute the booleans during options parsing
	// so the listener never needs to re-walk the `aspects` array.
	activeAspects map[string]bool
}

func parseOptions(raw any) options {
	opts := options{
		activeAspects: map[string]bool{
			aspectNoHref:       true,
			aspectInvalidHref:  true,
			aspectPreferButton: true,
		},
	}
	m := utils.GetOptionsMap(raw)
	if m == nil {
		return opts
	}
	opts.components = jsxa11yutil.StringSliceOption(m["components"])
	opts.specialLink = jsxa11yutil.StringSliceOption(m["specialLink"])
	if rawAspects, ok := m["aspects"]; ok {
		// An explicit aspects array — including an empty one — replaces the
		// default. Upstream's schema requires `minItems: 1`, but the rule
		// itself doesn't re-validate, so we mirror the runtime behavior:
		// any provided array, including an empty one, deactivates all
		// aspects until names are listed.
		//
		// Distinguishing "absent" vs "explicit []" is what the outer `ok`
		// guards: absent → keep the all-true default; present → start
		// from all-false and re-enable the listed names. StringSliceOption
		// flattens both cases into a (possibly empty) `[]string`, so the
		// `ok` check remains the single source of the absent/present
		// distinction.
		if aspects := jsxa11yutil.StringSliceOption(rawAspects); aspects != nil {
			opts.activeAspects = map[string]bool{
				aspectNoHref:       false,
				aspectInvalidHref:  false,
				aspectPreferButton: false,
			}
			for _, name := range aspects {
				if _, known := opts.activeAspects[name]; known {
					opts.activeAspects[name] = true
				}
			}
		}
	}
	return opts
}

var AnchorIsValidRule = rule.Rule{
	Name: "jsx-a11y/anchor-is-valid",
	Run: func(ctx rule.RuleContext, rawOptions any) rule.RuleListeners {
		opts := parseOptions(rawOptions)
		// typeCheck = ['a'].concat(componentOptions). Resolved against the
		// element's effective name via `jsxa11yutil.GetElementType` so both
		// `settings['jsx-a11y'].components` remap (Link → a) and the
		// `polymorphicPropName` setting are honored — same as upstream's
		// `getElementType(context)(node)` curry.
		typeCheck := append([]string{"a"}, opts.components...)
		// propsToValidate = ['href'].concat(propOptions). Each prop is
		// looked up case-insensitively via FindAttributeByName and walked
		// through any literal-spread, mirroring `getProp(attrs, name)`.
		propsToValidate := append([]string{"href"}, opts.specialLink...)

		check := func(node *ast.Node) {
			nodeType := jsxa11yutil.GetElementType(node, ctx.Settings)
			if !slices.Contains(typeCheck, nodeType) {
				return
			}

			attrs := reactutil.GetJsxElementAttributes(node)

			// First pass: walk every href-like prop, classify each as
			// "missing/null", "invalid string", or "present and OK".
			//
			// Upstream's logic is:
			//
			//   const values = propsToValidate.map(p => getPropValue(getProp(attrs, p)));
			//   const hasAnyHref = values.some(v => v != null);
			//   const invalidHrefValues = values.filter(v =>
			//     v != null && typeof v === 'string' &&
			//     (!v.length || v === '#' || /^\W*?javascript:/.test(v)));
			//
			// We don't need to materialize the values array — a single
			// pass tracking the two flags is equivalent.
			hasAnyHref := false
			hasInvalidHref := false
			for _, propName := range propsToValidate {
				attr := jsxa11yutil.FindAttributeByName(attrs, propName)
				if jsxa11yutil.PropValueIsNullish(attr) {
					// `attr == nil` (absent prop), `prop={null}`,
					// `prop={undefined}`, and TS-wrapped variants of those
					// all collapse to `value == null` in upstream — they
					// contribute nothing to hasAnyHref.
					continue
				}
				hasAnyHref = true
				if val, ok := jsxa11yutil.PropStaticStringValue(attr); ok {
					// Only string values are eligible for the invalid-href
					// check. Booleans (`prop={true}`, boolean form), numbers,
					// and any non-string truthy synthesized values (member
					// access, calls, etc.) all skip the typeof === 'string'
					// guard upstream and stay valid.
					if val == "" || val == "#" || jsHrefPattern.MatchString(val) {
						hasInvalidHref = true
					}
				}
			}

			// `attributes.some(a => a.type === 'JSXSpreadAttribute')` — any
			// spread, literal or not, suppresses the noHref / preferButton
			// branches when there is no explicit href. Crucially, this is
			// independent from FindAttributeByName's literal-spread walk:
			// `<a {...{href: ''}}/>` has hasSpread=true AND will populate
			// the href-prop branch via the literal walk.
			hasSpread := false
			for _, a := range attrs {
				if a.Kind == ast.KindJsxSpreadAttribute {
					hasSpread = true
					break
				}
			}
			onClick := jsxa11yutil.FindAttributeByName(attrs, "onClick")

			if !hasAnyHref {
				// Upstream:
				//
				//   if (!hasSpreadOperator && activeAspects.noHref
				//     && (!onClick || (onClick && !activeAspects.preferButton))) {
				//     report noHref
				//   }
				//   if (!hasSpreadOperator && onClick && activeAspects.preferButton) {
				//     report preferButton
				//   }
				//
				// The `(onClick && !preferButton)` arm is reachable: when
				// only `noHref` is enabled and an onClick is present, the
				// element should still be flagged as missing href (because
				// preferButton is off, the onClick doesn't redirect the
				// report into a button-recommendation). Lock-in tests exist
				// upstream for this combination — see __tests__/.../anchor-is-valid-test.js.
				if !hasSpread && opts.activeAspects[aspectNoHref] &&
					(onClick == nil || !opts.activeAspects[aspectPreferButton]) {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "noHref",
						Description: noHrefErrorMessage,
					})
				}
				if !hasSpread && onClick != nil && opts.activeAspects[aspectPreferButton] {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "preferButton",
						Description: preferButtonErrorMessage,
					})
				}
				return
			}

			if !hasInvalidHref {
				return
			}
			// Hrefs are present but at least one is invalid. preferButton
			// takes precedence over invalidHref when both are active and an
			// onClick exists — upstream's `if (onClick && activeAspects.preferButton)`
			// followed by `else if (activeAspects.invalidHref)`.
			if onClick != nil && opts.activeAspects[aspectPreferButton] {
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "preferButton",
					Description: preferButtonErrorMessage,
				})
			} else if opts.activeAspects[aspectInvalidHref] {
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "invalidHref",
					Description: invalidHrefErrorMessage,
				})
			}
		}

		// Listen on both opening kinds. tsgo splits ESTree's
		// JSXOpeningElement into KindJsxOpeningElement (paired tags) and
		// KindJsxSelfClosingElement (`<a/>`); upstream's single
		// `JSXOpeningElement` listener fires on both forms.
		return rule.RuleListeners{
			ast.KindJsxOpeningElement:     check,
			ast.KindJsxSelfClosingElement: check,
		}
	},
}

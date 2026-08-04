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
	_ "embed"
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/jsxa11yutil"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

//go:embed anchor_is_valid.schema.json
var schemaJSON []byte

const (
	preferButtonErrorMessage = "Anchor used as a button. Anchors are primarily expected to navigate. Use the button element instead. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md"
	noHrefErrorMessage       = "The href attribute is required for an anchor to be keyboard accessible. Provide a valid, navigable address as the href value. If you cannot provide an href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md"
	invalidHrefErrorMessage  = "The href attribute requires a valid value to be accessible. Provide a valid, navigable address as the href value. If you cannot provide a valid href, but still need the element to resemble a link, use a button and change it with appropriate styles. Learn more: https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/anchor-is-valid.md"
)

// hasJavaScriptHref mirrors upstream's `/^\W*?javascript:/` without invoking
// Go's regexp engine for every static href. JavaScript's `\w` is the ASCII
// class `[A-Za-z0-9_]`, so bytes from a UTF-8 encoded non-ASCII prefix are all
// non-word bytes and are skipped exactly as the upstream regexp skips their
// code points. The first word character must begin the literal
// `javascript:`; a word prefix such as `xjavascript:` therefore stays valid.
func hasJavaScriptHref(value string) bool {
	start := 0
	for start < len(value) {
		c := value[start]
		if c == '_' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' {
			break
		}
		start++
	}
	const prefix = "javascript:"
	return len(value)-start >= len(prefix) && value[start:start+len(prefix)] == prefix
}

func isInvalidHrefValue(value string) bool {
	return value == "" || value == "#" || hasJavaScriptHref(value)
}

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
	// These booleans mirror upstream's activeAspects object without requiring
	// three map lookups for every matching JSX element.
	noHref       bool
	invalidHref  bool
	preferButton bool
}

func parseOptions(raw []any) options {
	opts := options{
		noHref:       true,
		invalidHref:  true,
		preferButton: true,
	}
	if len(raw) == 0 {
		return opts
	}
	m, _ := raw[0].(map[string]interface{})
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
			opts.noHref = false
			opts.invalidHref = false
			opts.preferButton = false
			for _, name := range aspects {
				switch name {
				case aspectNoHref:
					opts.noHref = true
				case aspectInvalidHref:
					opts.invalidHref = true
				case aspectPreferButton:
					opts.preferButton = true
				}
			}
		}
	}
	return opts
}

var AnchorIsValidRule = rule.Rule{
	Name:   "jsx-a11y/anchor-is-valid",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, rawOptions []any) rule.RuleListeners {
		opts := parseOptions(rawOptions)
		// typeCheck = ['a'].concat(componentOptions). Resolved against the
		// element's effective name via `jsxa11yutil.GetElementType` so both
		// `settings['jsx-a11y'].components` remap (Link → a) and the
		// `polymorphicPropName` setting are honored — same as upstream's
		// `getElementType(context)(node)` curry.
		isTargetType := func(elementType string) bool {
			return elementType == "a" || slices.Contains(opts.components, elementType)
		}
		// With no element-type settings, the raw tag name is already the
		// effective name. This is the overwhelmingly common configuration and
		// avoids re-reading nested settings maps for every JSX element.
		hasElementTypeSettings := false
		if a11y, ok := ctx.Settings["jsx-a11y"].(map[string]interface{}); ok {
			polymorphicPropName, _ := a11y["polymorphicPropName"].(string)
			components, _ := a11y["components"].(map[string]interface{})
			hasElementTypeSettings = polymorphicPropName != "" || len(components) != 0
		}
		// propsToValidate = ['href'].concat(propOptions). Each prop is
		// looked up case-insensitively via FindAttributeByName and walked
		// through any literal-spread, mirroring `getProp(attrs, name)`.
		propsToValidate := append([]string{"href"}, opts.specialLink...)
		checkInvalidHref := opts.invalidHref || opts.preferButton
		sourceText := ctx.SourceFile.Text()
		reportNode := func(node *ast.Node, message rule.RuleMessage) {
			// ReportNode trims leading trivia by constructing a fresh scanner.
			// JSX opening nodes always end at node.End(); scanner.SkipTrivia gives
			// the identical first-token position without allocating a scanner.
			ctx.ReportRange(core.NewTextRange(scanner.SkipTrivia(sourceText, node.Pos()), node.End()), message)
		}

		check := func(node *ast.Node) {
			var nodeType string
			if hasElementTypeSettings {
				nodeType = jsxa11yutil.GetElementType(node, ctx.Settings)
			} else {
				nodeType = reactutil.GetJsxElementTypeString(node)
			}
			if !isTargetType(nodeType) {
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
				if attr == nil {
					continue
				}
				if jsxa11yutil.AttributeIsBooleanForm(attr) {
					hasAnyHref = true
					continue
				}
				if checkInvalidHref {
					// LiteralStringValue covers the common direct-literal path and
					// decodes JSX entities before the security-sensitive URL check.
					// PropStaticStringValue is the full expression fallback. Trying
					// the string classification before the nullish classification
					// evaluates string-valued expressions only once instead of twice.
					if val, ok := jsxa11yutil.LiteralStringValue(attr); ok {
						hasAnyHref = true
						hasInvalidHref = isInvalidHrefValue(val)
						if hasInvalidHref {
							break
						}
						continue
					}
					if val, ok := jsxa11yutil.PropStaticStringValue(attr); ok {
						hasAnyHref = true
						hasInvalidHref = isInvalidHrefValue(val)
						if hasInvalidHref {
							break
						}
						continue
					}
				}
				if jsxa11yutil.PropValueIsNullish(attr) {
					// `attr == nil` (absent prop), `prop={null}`,
					// `prop={undefined}`, and TS-wrapped variants of those
					// all collapse to `value == null` in upstream — they
					// contribute nothing to hasAnyHref.
					continue
				}
				hasAnyHref = true
			}

			if !hasAnyHref {
				// `attributes.some(a => a.type === 'JSXSpreadAttribute')` — any
				// spread, literal or not, suppresses both missing-href branches.
				// Defer this scan until it matters; valid hrefs return without it.
				for _, attr := range attrs {
					if attr.Kind == ast.KindJsxSpreadAttribute {
						return
					}
				}
				onClick := jsxa11yutil.FindAttributeByName(attrs, "onClick")
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
				if opts.noHref && (onClick == nil || !opts.preferButton) {
					reportNode(node, rule.RuleMessage{
						Id:          "noHref",
						Description: noHrefErrorMessage,
					})
				}
				if onClick != nil && opts.preferButton {
					reportNode(node, rule.RuleMessage{
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
			if opts.preferButton {
				if onClick := jsxa11yutil.FindAttributeByName(attrs, "onClick"); onClick != nil {
					reportNode(node, rule.RuleMessage{
						Id:          "preferButton",
						Description: preferButtonErrorMessage,
					})
					return
				}
			}
			if opts.invalidHref {
				reportNode(node, rule.RuleMessage{
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

// cspell:words bday impp

package autocomplete_valid

import (
	_ "embed"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/jsxa11yutil"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed autocomplete_valid.schema.json
var schemaJSON []byte

// failMessage is the message axe-core's `autocomplete-valid` check emits when
// the autocomplete attribute fails validation. The upstream ESLint rule reads
// `violations[0].nodes[0].all[0].message` directly from axe-core; the message
// text comes from `lib/checks/forms/autocomplete-valid.json`'s `messages.fail`.
// Source: https://github.com/dequelabs/axe-core/blob/develop/lib/checks/forms/autocomplete-valid.json
const failMessage = "the autocomplete attribute is incorrectly formatted"

// Token sets mirror axe-core's `autocomplete` constant in
// `lib/commons/text/is-valid-autocomplete.js`, plus the `stateTerms` and
// `ignoredValues` extensions defined in `lib/checks/forms/autocomplete-valid.json`.
//
// stateTerms: terms whose presence ALONE (without any other tokens) makes the
// autocomplete value valid. The base `["on", "off"]` come from the WHATWG HTML
// spec; axe-core's autocomplete-valid.json extends with eight more pragmatic
// extras (`none`, `false`, `true`, etc.) to accept common author misuses.
var stateTerms = map[string]struct{}{
	// axe-core base
	"on":  {},
	"off": {},
	// axe-core autocomplete-valid.json extension
	"none":      {},
	"false":     {},
	"true":      {},
	"disabled":  {},
	"enabled":   {},
	"undefined": {},
	"null":      {},
	"xoff":      {},
	"xon":       {},
}

// standaloneTerms — the autofill field-name tokens that are valid on their
// own (without a contact qualifier). Source: WHATWG HTML autofill detail
// tokens, mirrored from axe-core's `autocomplete.standaloneTerms`.
var standaloneTerms = map[string]struct{}{
	"name":                 {},
	"honorific-prefix":     {},
	"given-name":           {},
	"additional-name":      {},
	"family-name":          {},
	"honorific-suffix":     {},
	"nickname":             {},
	"username":             {},
	"new-password":         {},
	"current-password":     {},
	"organization-title":   {},
	"organization":         {},
	"street-address":       {},
	"address-line1":        {},
	"address-line2":        {},
	"address-line3":        {},
	"address-level4":       {},
	"address-level3":       {},
	"address-level2":       {},
	"address-level1":       {},
	"country":              {},
	"country-name":         {},
	"postal-code":          {},
	"cc-name":              {},
	"cc-given-name":        {},
	"cc-additional-name":   {},
	"cc-family-name":       {},
	"cc-number":            {},
	"cc-exp":               {},
	"cc-exp-month":         {},
	"cc-exp-year":          {},
	"cc-csc":               {},
	"cc-type":              {},
	"transaction-currency": {},
	"transaction-amount":   {},
	"language":             {},
	"bday":                 {},
	"bday-day":             {},
	"bday-month":           {},
	"bday-year":            {},
	"sex":                  {},
	"url":                  {},
	"photo":                {},
	"one-time-code":        {},
}

// qualifiers — contact-channel qualifiers that may precede a qualifiedTerm
// (e.g. `home email`, `work tel`). Mirrors axe-core's `autocomplete.qualifiers`.
var qualifiers = map[string]struct{}{
	"home":   {},
	"work":   {},
	"mobile": {},
	"fax":    {},
	"pager":  {},
}

// qualifiedTerms — autofill field-name tokens that are valid only AFTER a
// qualifier (and also valid on their own, since they're additionally accepted
// in the no-qualifier case). Mirrors axe-core's `autocomplete.qualifiedTerms`.
var qualifiedTerms = map[string]struct{}{
	"tel":              {},
	"tel-country-code": {},
	"tel-national":     {},
	"tel-area-code":    {},
	"tel-local":        {},
	"tel-local-prefix": {},
	"tel-local-suffix": {},
	"tel-extension":    {},
	"email":            {},
	"impp":             {},
}

// locations — address grouping tokens (`billing` / `shipping`) that may
// optionally precede the rest of the value. Mirrors axe-core's
// `autocomplete.locations`.
var locations = map[string]struct{}{
	"billing":  {},
	"shipping": {},
}

// ignoredValues — tokens that axe-core's autocomplete-valid extension marks
// as "incomplete" (returns `undefined`) rather than failing. The ESLint rule
// only reports violations (`false` returns), so these effectively pass.
// Source: `lib/checks/forms/autocomplete-valid.json` `options.ignoredValues`.
var ignoredValues = map[string]struct{}{
	"text":     {},
	"pronouns": {},
	"gender":   {},
	"message":  {},
	"content":  {},
}

// excludedInputTypes — input `type` values that axe-core's
// `autocomplete-matches` (lib/commons/forms/autocomplete-matches.js) filters
// out before running any check. For these, the rule produces no violation
// regardless of the autocomplete value. Source comparison is
// case-insensitive: axe-core's SerialVirtualNode lowercases `attributes.type`
// into `virtualNode.props.type`.
//
// The other gates in autocomplete-matches (`readonly`, `disabled`,
// `aria-readonly`, `aria-disabled`, `tabindex`+role) are unreachable in this
// port because the upstream ESLint rule passes ONLY `{autocomplete, type}`
// into `runVirtualRule` — every other attribute lookup in matches sees
// `null` and falls through. Only the `type` filter affects observable
// behavior here.
var excludedInputTypes = map[string]struct{}{
	"submit": {},
	"reset":  {},
	"button": {},
	"hidden": {},
}

type options struct {
	inputComponents []string
}

func parseOptions(raw []any) options {
	opts := options{}
	if len(raw) == 0 {
		return opts
	}
	m, _ := raw[0].(map[string]interface{})
	opts.inputComponents = utils.ToStringSlice(m["inputComponents"])
	return opts
}

var AutocompleteValidRule = rule.Rule{
	Name:   "jsx-a11y/autocomplete-valid",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, rawOptions []any) rule.RuleListeners {
		opts := parseOptions(rawOptions)

		// inputTypes is the union of `["input"]` and the user-provided
		// inputComponents. Upstream builds this fresh inside the listener;
		// hoisting it here is a no-op semantically because options can't
		// change during a single run.
		inputTypes := map[string]struct{}{"input": {}}
		for _, c := range opts.inputComponents {
			inputTypes[c] = struct{}{}
		}

		check := func(node *ast.Node) {
			elType := jsxa11yutil.GetElementType(node, ctx.Settings)
			if _, ok := inputTypes[elType]; !ok {
				return
			}

			attrs := reactutil.GetJsxElementAttributes(node)
			autocompleteAttr := jsxa11yutil.FindAttributeByName(attrs, "autocomplete")
			if autocompleteAttr == nil {
				return
			}
			// Upstream gates on `typeof autocomplete !== 'string'` after
			// `getLiteralPropValue`. LiteralPropStringValue mirrors that —
			// boolean form, dynamic identifiers, logical / conditional
			// expressions, etc. all return ("", false) and we exit early.
			autocompleteValue, ok := jsxa11yutil.LiteralPropStringValue(autocompleteAttr)
			if !ok {
				return
			}

			// Mirror axe-core's `autocomplete-matches` filter
			// (lib/commons/forms/autocomplete-matches.js): when the literal
			// `type` value is in `excludedInputTypes` (submit / reset /
			// button / hidden), the rule must NOT report. axe-core's
			// matches() runs before evaluate(), so the violation list is
			// empty for these inputs regardless of the autocomplete value.
			//
			// Applies to every element that passed the inputTypes gate —
			// the ESLint plugin always hardcodes `nodeName: 'input'` when
			// calling `runVirtualRule`, so the matches gate `nodeName !==
			// 'input'` upstream never fires for custom inputComponents or
			// components-map entries; the type filter is the only matches
			// branch that affects observable behavior here.
			//
			// Comparison is case-insensitive because axe-core's
			// SerialVirtualNode lowercases `attributes.type` into
			// `virtualNode.props.type`.
			if typeAttr := jsxa11yutil.FindAttributeByName(attrs, "type"); typeAttr != nil {
				if typeValue, ok := jsxa11yutil.LiteralPropStringValue(typeAttr); ok {
					if _, excluded := excludedInputTypes[strings.ToLower(typeValue)]; excluded {
						return
					}
				}
			}

			if isValidAutocomplete(autocompleteValue) {
				return
			}
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "autocompleteValid",
				Description: failMessage,
			})
		}

		return rule.RuleListeners{
			ast.KindJsxOpeningElement:     check,
			ast.KindJsxSelfClosingElement: check,
		}
	},
}

// isValidAutocomplete ports axe-core's `isValidAutocomplete` from
// `lib/commons/text/is-valid-autocomplete.js`, with the extended `stateTerms`
// and `ignoredValues` from the rule's check options baked in.
//
// Returns true when the value matches one of:
//   - empty string (after trim)
//   - a stateTerm (e.g. "on", "off", or one of the autocomplete-valid extras)
//   - the autofill detail-token grammar:
//     [section-*]? [billing|shipping]? [home|work|mobile|fax|pager]? <field-name> [webauthn]?
//     where <field-name> is in standaloneTerms when no qualifier was used,
//     OR in qualifiedTerms when a qualifier was used.
//   - any value whose final token (after stripping the optional `webauthn`,
//     section-, location, qualifier prefixes) is in `ignoredValues` —
//     axe-core returns `undefined` here, treated as "incomplete" by the rule
//     engine and therefore not a violation.
//
// Returns false for any value that doesn't fit the grammar.
func isValidAutocomplete(value string) bool {
	value = strings.TrimSpace(strings.ToLower(value))
	if _, ok := stateTerms[value]; ok {
		return true
	}
	if value == "" {
		return true
	}

	// `\s+` split per axe-core; strings.Fields collapses any whitespace run.
	terms := strings.Fields(value)
	if len(terms) == 0 {
		return true // unreachable: covered by `value == ""` above, but kept defensive
	}

	// Optional trailing `webauthn`. After popping, if no terms remain the
	// value was just `"webauthn"` alone — that's invalid.
	if terms[len(terms)-1] == "webauthn" {
		terms = terms[:len(terms)-1]
		if len(terms) == 0 {
			return false
		}
	}

	// Optional `section-*` prefix. Upstream guards on length > 8 so that the
	// bare token `"section-"` (8 chars) is NOT treated as a prefix. Mirror
	// that exactly to lock the boundary.
	if len(terms[0]) > 8 && strings.HasPrefix(terms[0], "section-") {
		terms = terms[1:]
	}
	if len(terms) == 0 {
		return false
	}

	// Optional `billing` / `shipping` location.
	if _, ok := locations[terms[0]]; ok {
		terms = terms[1:]
	}
	if len(terms) == 0 {
		return false
	}

	// Optional contact qualifier. When present, the field name must come from
	// qualifiedTerms — standaloneTerms become invalid (matches upstream's
	// `standaloneTerms = []` reset).
	qualifierUsed := false
	if _, ok := qualifiers[terms[0]]; ok {
		terms = terms[1:]
		qualifierUsed = true
	}

	// Exactly one term must remain — the field-name purpose.
	if len(terms) != 1 {
		return false
	}
	purposeTerm := terms[0]

	// `ignoredValues` short-circuit: axe-core returns `undefined` (incomplete);
	// the rule treats anything other than `false` (violation) as passing.
	if _, ok := ignoredValues[purposeTerm]; ok {
		return true
	}

	if qualifierUsed {
		_, ok := qualifiedTerms[purposeTerm]
		return ok
	}
	if _, ok := standaloneTerms[purposeTerm]; ok {
		return true
	}
	if _, ok := qualifiedTerms[purposeTerm]; ok {
		return true
	}
	return false
}

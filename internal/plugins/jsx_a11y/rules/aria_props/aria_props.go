// cspell:ignore Damerau labeledby

// Package aria_props ports eslint-plugin-jsx-a11y's `aria-props` rule. The
// rule flags any JSX attribute whose name starts with `aria-` but is NOT a
// recognized ARIA state / property as defined by `aria-query`'s
// `ariaPropsMap` (e.g. typos like `aria-labeledby` for `aria-labelledby`).
//
// Upstream signature: no options — the schema is `generateObjSchema()`
// (an empty object).
//
// Trigger: a JsxAttribute whose name passes a CASE-SENSITIVE `indexOf('aria-')
// === 0` check upstream — only a literal lowercase `aria-` prefix qualifies.
// `<div ARIA-HIDDEN />` does NOT enter the validation branch (early-return)
// and is therefore valid; `<div aria- />` enters the branch, fails the
// `aria.has` check, and reports.
//
// Diagnostic shape: the message lists the offending attribute name and, when
// the canonical-list lookup yields close matches (Damerau-Levenshtein OSA
// distance ≤ 2 after upper-casing both sides), appends a "Did you mean to
// use ..." suffix. Empty suggestion list omits the suffix entirely.
package aria_props

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/jsxa11yutil"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// suggestionDistanceThreshold mirrors upstream's `THRESHOLD = 2` in
// `getSuggestion.js`. Only candidates within this Damerau-Levenshtein OSA
// distance are surfaced.
const suggestionDistanceThreshold = 2

// suggestionLimit mirrors upstream's `limit = 2` default — at most this many
// suggestions are joined into the diagnostic message.
const suggestionLimit = 2

// errorMessage mirrors upstream's `errorMessage(name)` exactly: the base
// "<name>: This attribute is an invalid ARIA attribute." with an optional
// "Did you mean to use ...?" suffix when suggestions are available.
//
// The suggestions list is joined with the default Array.prototype.toString
// separator (a bare comma — no space), which is what JavaScript's template
// literal `${suggestions}` produces. This is intentional and visible in
// upstream tests, so we mirror byte-for-byte.
func errorMessage(name string) string {
	base := name + ": This attribute is an invalid ARIA attribute."
	suggestions := getSuggestion(name, jsxa11yutil.AriaPropertyNames, jsxa11yutil.AriaPropertyNamesUpper, suggestionLimit)
	if len(suggestions) == 0 {
		return base
	}
	return base + " Did you mean to use " + strings.Join(suggestions, ",") + "?"
}

var AriaPropsRule = rule.Rule{
	Name:   "jsx-a11y/aria-props",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindJsxAttribute: func(attr *ast.Node) {
				name := reactutil.GetJsxPropName(attr)
				// Upstream uses `String.prototype.indexOf` which is
				// case-sensitive. `ARIA-` / `Aria-` / mixed-case prefixes
				// do NOT enter the validation branch.
				if !strings.HasPrefix(name, "aria-") {
					return
				}
				if _, ok := jsxa11yutil.AriaPropertySet[name]; ok {
					return
				}
				ctx.ReportNode(attr, rule.RuleMessage{
					Id:          "invalidAriaProp",
					Description: errorMessage(name),
				})
			},
		}
	},
}

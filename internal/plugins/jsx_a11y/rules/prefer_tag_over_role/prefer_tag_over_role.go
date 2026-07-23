// cspell:ignore heckbox

// Package prefer_tag_over_role ports eslint-plugin-jsx-a11y's
// `prefer-tag-over-role` rule. The rule flags a JSX element whose explicit
// `role` attribute matches an ARIA role for which a semantic HTML element
// already exists — e.g. `<div role="checkbox" />` should be `<input
// type="checkbox">`, `<span role="heading" />` should be `<h1>` ... `<h6>`.
//
// The rule has no options. The diagnostic message is verbatim:
//
//	Use {{tag}} instead of the "{{role}}" role to ensure accessibility across all devices.
//
// Trigger: a JsxOpeningElement / JsxSelfClosingElement whose
//
//  1. has a literal-resolvable `role` attribute (the LAST whitespace-separated
//     token is consulted, mirroring upstream `getLastPropValue`),
//  2. that token names a non-abstract ARIA role with at least one semantic
//     HTML element mapping in `aria-query`'s `roleElements`, AND
//  3. the element's effective tag (after `getElementType` —
//     `polymorphicPropName` + `components` map resolution) is NOT the `name`
//     of any of those semantic elements.
//
// Listener gate parity: upstream listens on `JSXOpeningElement` only — ESTree
// wraps both paired and self-closing forms under that node. tsgo splits them
// into KindJsxOpeningElement (paired) and KindJsxSelfClosingElement
// (self-closing), so we register both kinds — see [no_redundant_roles] for
// the same pattern.
package prefer_tag_over_role

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	jsxtx "github.com/microsoft/typescript-go/shim/transformers/jsxtransforms"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/jsxa11yutil"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type roleEntry struct {
	// formatted is the comma-and-or-joined tag list rendered exactly as
	// upstream `formatTag` + the `, or `-trailing join produces. Pre-computed
	// here because the table is data-only and never depends on user input.
	formatted string
	// tagNames is the set of HTML element names this role's semantic
	// implementations span, in insertion order with duplicates removed.
	// Upstream's matched-tag check is `matchedTags.some(tag.name ===
	// elementType(node))` — names only, attributes are NOT inspected for the
	// match (they only appear in the diagnostic message). Duplicates would
	// repeat the same `name === ...` comparison, so we deduplicate.
	tagNames []string
}

// roleElementTable is derived from aria-query@5.x roleElementMap.
// Each entry pre-computes:
//  1. formatted: the tag list rendered exactly as upstream prefer-tag-over-role.js
//     composes it for the diagnostic message (formatTag + ", or " join).
//  2. tagNames: the set of HTML element names (ordered, deduplicated) used for
//     the upstream matchedTags.some(tag.name === elementType(node)) check.
//
// Source: aria-query 5.x — `roleElementMap` is built from `rolesMap` by
// concatenating each role's baseConcepts + relatedConcepts and keeping only
// the concepts whose module === "HTML". `formatTag` in upstream takes ONLY
// the first element of `attributes` — the rest are ignored even when the
// role concept lists multiple constraints (e.g. textbox's input variants).
// We mirror byte-for-byte; quirks like `<input list=...>` for textbox /
// searchbox / combobox or `<img alt=...>` for the img role are upstream-
// observable behavior, not bugs we're "fixing" silently.
var roleElementTable = map[string]roleEntry{
	"article":       {formatted: `<article>`, tagNames: []string{"article"}},
	"banner":        {formatted: `<header>`, tagNames: []string{"header"}},
	"blockquote":    {formatted: `<blockquote>`, tagNames: []string{"blockquote"}},
	"button":        {formatted: `<input type="button">, <input type="image">, <input type="reset">, <input type="submit">, or <button>`, tagNames: []string{"input", "button"}},
	"caption":       {formatted: `<caption>`, tagNames: []string{"caption"}},
	"cell":          {formatted: `<td>`, tagNames: []string{"td"}},
	"checkbox":      {formatted: `<input type="checkbox">`, tagNames: []string{"input"}},
	"code":          {formatted: `<code>`, tagNames: []string{"code"}},
	"columnheader":  {formatted: `<th>, <th scope="col">, or <th scope="colgroup">`, tagNames: []string{"th"}},
	"combobox":      {formatted: `<input list=...>, <input list=...>, <input list=...>, <input list=...>, <input list=...>, <input list=...>, or <select multiple=...>`, tagNames: []string{"input", "select"}},
	"complementary": {formatted: `<aside>, <aside aria-label=...>, or <aside aria-labelledby=...>`, tagNames: []string{"aside"}},
	"contentinfo":   {formatted: `<footer>`, tagNames: []string{"footer"}},
	"definition":    {formatted: `<dd>`, tagNames: []string{"dd"}},
	"deletion":      {formatted: `<del>`, tagNames: []string{"del"}},
	"dialog":        {formatted: `<dialog>`, tagNames: []string{"dialog"}},
	"document":      {formatted: `<html>`, tagNames: []string{"html"}},
	"emphasis":      {formatted: `<em>`, tagNames: []string{"em"}},
	"figure":        {formatted: `<figure>`, tagNames: []string{"figure"}},
	"form":          {formatted: `<form aria-label=...>, <form aria-labelledby=...>, or <form name=...>`, tagNames: []string{"form"}},
	"generic":       {formatted: `<a>, <area>, <aside>, <b>, <bdo>, <body>, <data>, <div>, <footer>, <header>, <hgroup>, <i>, <pre>, <q>, <samp>, <section>, <small>, <span>, or <u>`, tagNames: []string{"a", "area", "aside", "b", "bdo", "body", "data", "div", "footer", "header", "hgroup", "i", "pre", "q", "samp", "section", "small", "span", "u"}},
	"gridcell":      {formatted: `<td>`, tagNames: []string{"td"}},
	"group":         {formatted: `<details>, <fieldset>, <optgroup>, or <address>`, tagNames: []string{"details", "fieldset", "optgroup", "address"}},
	"heading":       {formatted: `<h1>, <h2>, <h3>, <h4>, <h5>, or <h6>`, tagNames: []string{"h1", "h2", "h3", "h4", "h5", "h6"}},
	"img":           {formatted: `<img alt=...>, or <img alt=...>`, tagNames: []string{"img"}},
	"insertion":     {formatted: `<ins>`, tagNames: []string{"ins"}},
	"link":          {formatted: `<a href=...>, or <area href=...>`, tagNames: []string{"a", "area"}},
	"list":          {formatted: `<menu>, <ol>, or <ul>`, tagNames: []string{"menu", "ol", "ul"}},
	"listbox":       {formatted: `<select size=...>, <select multiple=...>, or <datalist>`, tagNames: []string{"select", "datalist"}},
	"listitem":      {formatted: `<li>`, tagNames: []string{"li"}},
	"main":          {formatted: `<main>`, tagNames: []string{"main"}},
	"mark":          {formatted: `<mark>`, tagNames: []string{"mark"}},
	"math":          {formatted: `<math>`, tagNames: []string{"math"}},
	"meter":         {formatted: `<meter>`, tagNames: []string{"meter"}},
	"navigation":    {formatted: `<nav>`, tagNames: []string{"nav"}},
	"option":        {formatted: `<option>`, tagNames: []string{"option"}},
	"paragraph":     {formatted: `<p>`, tagNames: []string{"p"}},
	"presentation":  {formatted: `<img alt=...>`, tagNames: []string{"img"}},
	"progressbar":   {formatted: `<progress>`, tagNames: []string{"progress"}},
	"radio":         {formatted: `<input type="radio">`, tagNames: []string{"input"}},
	"region":        {formatted: `<section aria-label=...>, or <section aria-labelledby=...>`, tagNames: []string{"section"}},
	"row":           {formatted: `<tr>`, tagNames: []string{"tr"}},
	"rowgroup":      {formatted: `<tbody>, <tfoot>, or <thead>`, tagNames: []string{"tbody", "tfoot", "thead"}},
	"rowheader":     {formatted: `<th scope="row">, or <th scope="rowgroup">`, tagNames: []string{"th"}},
	"searchbox":     {formatted: `<input list=...>`, tagNames: []string{"input"}},
	"separator":     {formatted: `<hr>`, tagNames: []string{"hr"}},
	"slider":        {formatted: `<input type="range">`, tagNames: []string{"input"}},
	"spinbutton":    {formatted: `<input type="number">`, tagNames: []string{"input"}},
	"status":        {formatted: `<output>`, tagNames: []string{"output"}},
	"strong":        {formatted: `<strong>`, tagNames: []string{"strong"}},
	"subscript":     {formatted: `<sub>`, tagNames: []string{"sub"}},
	"superscript":   {formatted: `<sup>`, tagNames: []string{"sup"}},
	"table":         {formatted: `<table>`, tagNames: []string{"table"}},
	"term":          {formatted: `<dfn>, or <dt>`, tagNames: []string{"dfn", "dt"}},
	"textbox":       {formatted: `<input type=...>, <input list=...>, <input list=...>, <input list=...>, <input list=...>, or <textarea>`, tagNames: []string{"input", "textarea"}},
	"time":          {formatted: `<time>`, tagNames: []string{"time"}},
}

func errorMessage(tag, role string) string {
	return "Use " + tag + ` instead of the "` + role + `" role to ensure accessibility across all devices.`
}

// extractRoleString mirrors upstream `getPropValue(role)` (full
// jsx-ast-utils staticEval), with one tsgo-specific patch: on the direct
// `<X role="...">` shape, decode HTML entities so `role="&#99;heckbox"`
// resolves to "checkbox" like ESTree's JSX parser would expose it.
// PropStaticStringValue itself reads StringLiteral.Text verbatim, but
// every other direct-attribute consumer in jsxa11yutil already routes
// through DecodeEntities for parity. Non-direct shapes (`role={...}`,
// spread literal, binary, call, member, ...) carry JS string literals
// and need no decode — delegate to PropStaticStringValue.
func extractRoleString(roleAttr *ast.Node) (string, bool) {
	if roleAttr.Kind == ast.KindJsxAttribute {
		init := roleAttr.AsJsxAttribute().Initializer
		if init != nil && init.Kind == ast.KindStringLiteral {
			return jsxtx.DecodeEntities(init.AsStringLiteral().Text), true
		}
	}
	return jsxa11yutil.PropStaticStringValue(roleAttr)
}

// lastSpaceSeparatedToken mirrors upstream `getLastPropValue`'s tail
// extraction: split a non-empty role string on whitespace and take the last
// token. ESLint's roleElements is keyed by single role names (`checkbox`,
// `heading`, …); a multi-role attribute like `role="button checkbox"` is
// matched against the LAST token only. Returns the input unchanged when no
// space is present.
//
// Note that `strings.LastIndexByte(s, ' ')` is intentionally used (NOT
// `strings.LastIndexFunc(s, unicode.IsSpace)`): jsx-ast-utils' upstream
// equivalent (`String.prototype.lastIndexOf(' ')`) only matches the ASCII
// SPACE U+0020, not other whitespace such as tab / NBSP. Mirror byte-for-byte.
func lastSpaceSeparatedToken(s string) string {
	if i := strings.LastIndexByte(s, ' '); i >= 0 {
		return s[i+1:]
	}
	return s
}

var PreferTagOverRoleRule = rule.Rule{
	Name:   "jsx-a11y/prefer-tag-over-role",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		check := func(node *ast.Node) {
			attrs := reactutil.GetJsxElementAttributes(node)
			roleAttr := jsxa11yutil.FindAttributeByName(attrs, "role")
			if roleAttr == nil {
				return
			}
			rawRole, ok := extractRoleString(roleAttr)
			if !ok || rawRole == "" {
				return
			}
			// `getLastPropValue` returns the empty string when the input ends
			// with a space (`"button "` → ""). Upstream then fails the
			// `roleElements.get("")` lookup and skips. We do the same — empty
			// keys are not present in roleElementTable.
			role := lastSpaceSeparatedToken(rawRole)
			entry, ok := roleElementTable[role]
			if !ok {
				return
			}
			elementType := jsxa11yutil.GetElementType(node, ctx.Settings)
			for _, name := range entry.tagNames {
				if name == elementType {
					return
				}
			}
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "preferTagOverRole",
				Description: errorMessage(entry.formatted, role),
			})
		}

		return rule.RuleListeners{
			ast.KindJsxOpeningElement:     check,
			ast.KindJsxSelfClosingElement: check,
		}
	},
}

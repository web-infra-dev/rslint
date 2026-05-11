package scope

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// polymorphicSettings mirrors `polymorphicPropName: 'as'` — the as-prop value
// becomes the resolved tag name. `<Box as="th" scope />` should resolve to
// "th" and skip; `<Box as="div" scope />` should resolve to "div" and report.
var polymorphicSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"polymorphicPropName": "as",
	},
}

// polymorphicAllowListSettings restricts the `as` swap to a specific raw tag
// list — locks in that GetElementType honors the allow-list.
var polymorphicAllowListSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"polymorphicPropName":  "as",
		"polymorphicAllowList": []interface{}{"Box"},
	},
}

// componentsToCustomSettings maps a custom component to ANOTHER custom
// component (still not in the dom set) — verifies the rule skips when the
// final resolved name stays outside the dom set.
var componentsToCustomSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"components": map[string]interface{}{
			"Foo": "Bar",
		},
	},
}

// emptyJsxA11ySettings has the `jsx-a11y` key but no inner config —
// exercises the GetElementType / IsDOMElement defensive paths when the
// settings tree exists but is empty.
var emptyJsxA11ySettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{},
}

// TestScopeExtras is the catch-all for everything beyond upstream's 10-case
// suite (which lives in scope_upstream_test.go). Cases here fall in groups,
// kept in one suite so a regression bisects easily:
//
//  1. **Dimension 4 universal edge shapes** — namespaced attribute names,
//     paired vs self-closing element kinds, multiple attributes in one
//     element, listener boundary across nested elements, member-expression
//     and namespaced tag names.
//  2. **Case-sensitivity matrix** — both directions: case-insensitive prop
//     name match (`SCOPE`, `Scope`) AND case-sensitive dom-set lookup
//     (`<TH scope />` is silently SKIPPED because aria-query's dom map is
//     keyed by lowercase).
//  3. **polymorphicPropName × components × polymorphicAllowList resolution
//     matrix** — every combination that GetElementType has to reconcile.
//  4. **Real-world React / TS patterns** — TS generics, hyphenated tags,
//     hooks / forwardRef / memo / HOC wrappers, fragments + portals,
//     conditional rendering, multi-component files, generator/async/IIFE
//     bodies. These don't lock new semantics; they certify the listener
//     fires reliably across the AST shapes a real codebase produces.
func TestScopeExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ScopeRule, []rule_tester.ValidTestCase{
		// ============================================================
		// Group 1: Case-insensitive prop name matching
		// ============================================================
		// Upstream uses `name.toUpperCase() !== 'SCOPE'` so any case variant
		// of "scope" matches. On `<th>`, all variants are exempt via the th
		// branch; the listener still fires (matches), then the th branch
		// short-circuits.
		{Code: `<th SCOPE />`, Tsx: true},
		{Code: `<th Scope />`, Tsx: true},
		{Code: `<th SCoPE />`, Tsx: true},
		{Code: `<th scOpe="row" />`, Tsx: true},
		// Non-scope attribute names with similar spellings — propName !== scope.
		{Code: `<div scoped />`, Tsx: true},
		{Code: `<div scope-of />`, Tsx: true},

		// ============================================================
		// Group 2: Namespaced attribute names — composite name !== "scope"
		// ============================================================
		// reactutil.GetJsxPropName returns the composite "ns:name" string,
		// which uppercases to "NS:NAME" — never equal to "SCOPE". The
		// listener fires but the case-insensitive prop-name gate returns
		// early. Locks in that the JsxNamespacedName branch produces the
		// composite form upstream expects.
		{Code: `<div xml:scope />`, Tsx: true},
		{Code: `<div xml:scope="row" />`, Tsx: true},
		{Code: `<th xml:scope />`, Tsx: true},

		// ============================================================
		// Group 3: th case-sensitivity — DOM-set lookup is lowercase
		// ============================================================
		// aria-query's `dom` map is keyed by lowercase HTML names.
		// `<TH scope />` resolves to "TH" — not in the dom set → silently
		// SKIPPED via `!dom.has(tagName)`. This is upstream's actual behavior
		// (the case-insensitive `tagName.toUpperCase() === 'TH'` guard never
		// runs because the case-sensitive dom check rejects "TH" first).
		// Lock in this surprising-but-faithful behavior.
		{Code: `<TH scope />`, Tsx: true},
		{Code: `<Th scope="row" />`, Tsx: true},

		// ============================================================
		// Group 4: Spread attributes — different AST kind, listener never fires
		// ============================================================
		// JsxSpreadAttribute is KindJsxSpreadAttribute, not KindJsxAttribute.
		// Even `{...{scope: 'row'}}` cannot trigger this rule because no
		// JsxAttribute named scope is materialized in the tree.
		{Code: `<div {...{scope: 'row'}} />`, Tsx: true},
		{Code: `<div {...props} />`, Tsx: true},
		{Code: `<div {...{scope: 'row'}} {...props} />`, Tsx: true},

		// ============================================================
		// Group 5: components map — non-DOM target stays skipped
		// ============================================================
		// `Foo` → `Bar`, neither in dom set → IsDOMElement gate skips.
		{Code: `<Foo scope="row" />`, Tsx: true, Settings: componentsToCustomSettings},
		// Empty `jsx-a11y` settings — defensive: rawType "Foo" not in dom →
		// skipped.
		{Code: `<Foo scope />`, Tsx: true, Settings: emptyJsxA11ySettings},

		// ============================================================
		// Group 6: polymorphicPropName resolving to th — exempt
		// ============================================================
		// `<Box as="th" scope />` resolves to "th" → exempt.
		{Code: `<Box as="th" scope />`, Tsx: true, Settings: polymorphicSettings},
		{Code: `<Box as="th" scope="col" />`, Tsx: true, Settings: polymorphicSettings},

		// ============================================================
		// Group 7: polymorphicPropName resolving to non-DOM — skipped
		// ============================================================
		// `<Box as="ComponentName" scope />` → rawType "ComponentName" not in
		// dom → IsDOMElement gate skips.
		{Code: `<Box as="ComponentName" scope />`, Tsx: true, Settings: polymorphicSettings},

		// ============================================================
		// Group 8: polymorphicAllowList restricts the `as` swap
		// ============================================================
		// `<Box as="th" />` IS in allow-list → swap applies → rawType "th" → exempt.
		{Code: `<Box as="th" scope="row" />`, Tsx: true, Settings: polymorphicAllowListSettings},
		// `<Other as="div" />` is NOT in allow-list → no swap → rawType "Other"
		// not in dom → skipped (regardless of scope value).
		{Code: `<Other as="div" scope="row" />`, Tsx: true, Settings: polymorphicAllowListSettings},

		// ============================================================
		// Group 9: Member-expression / namespaced tag names — non-DOM
		// ============================================================
		// `<UX.Layout>` resolves to "UX.Layout" — not a single lowercase HTML
		// name → not in dom → skipped.
		{Code: `<UX.Layout scope />`, Tsx: true},
		{Code: `<this.Foo scope="col" />`, Tsx: true},
		// `<svg:circle>` resolves to "svg:circle" — composite namespaced name
		// not in aria-query's lowercase dom map → skipped.
		{Code: `<svg:circle scope="row" />`, Tsx: true},

		// ============================================================
		// Group 10: th in nested / paired forms (still exempt)
		// ============================================================
		{Code: `<table><tr><th scope="col">Header</th></tr></table>`, Tsx: true},
		{Code: `<table><thead><tr><th scope="col" /></tr></thead></table>`, Tsx: true},
		// Paired form (non-self-closing) — listener fires on JsxAttribute
		// inside the JsxOpeningElement; same exemption applies.
		{Code: `<th scope="row">Header text</th>`, Tsx: true},

		// ============================================================
		// Group 11: scope value variants on th — all exempt
		// ============================================================
		{Code: `<th scope="col" />`, Tsx: true},
		{Code: `<th scope="row" />`, Tsx: true},
		{Code: `<th scope="rowgroup" />`, Tsx: true},
		{Code: `<th scope="colgroup" />`, Tsx: true},
		// Non-string value — the rule doesn't inspect the value; only the
		// name + tag matter.
		{Code: `<th scope={someVar} />`, Tsx: true},
		{Code: `<th scope={cond ? 'row' : 'col'} />`, Tsx: true},
		{Code: `<th scope={getScope()} />`, Tsx: true},
		{Code: "<th scope={`${dynamic}`} />", Tsx: true},

		// ============================================================
		// Group 12: TS generic JSX components → th
		// ============================================================
		// `<List<string>>` parses as JsxOpeningElement with type args; the
		// scope attribute remains a plain JsxAttribute. With components map
		// `List → th`, exempt.
		{
			Code: `<List<string> scope="col" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"components": map[string]interface{}{
						"List": "th",
					},
				},
			},
		},

		// ============================================================
		// Group 13: TS wrappers around the value — irrelevant to rule
		// ============================================================
		// The rule never inspects the value, so TS wrappers (`as` / `!` /
		// `satisfies`) don't change behavior on th-exempt cases.
		{Code: `<th scope={"col" as string} />`, Tsx: true},
		{Code: `<th scope={"col"!} />`, Tsx: true},
		{Code: `<th scope={"col" satisfies string} />`, Tsx: true},
		{Code: `<th scope={("col")} />`, Tsx: true},

		// ============================================================
		// Group 14: Hyphenated DOM tags (web components)
		// ============================================================
		// `<my-element>` resolves to "my-element" — not in aria-query's dom
		// map → IsDOMElement skips → no report regardless of scope.
		{Code: `<my-element scope />`, Tsx: true},
		{Code: `<my-element scope="row" />`, Tsx: true},

		// ============================================================
		// Group 15: Real-world th in semantic table patterns
		// ============================================================
		{
			Code: `function DataTable() { return <table><thead><tr><th scope="col">Name</th><th scope="col">Age</th></tr></thead></table>; }`,
			Tsx:  true,
		},
		{
			Code: `function PivotTable({ rows }) { return <table>{rows.map(r => <tr><th scope="row" key={r.id}>{r.label}</th></tr>)}</table>; }`,
			Tsx:  true,
		},

		// ============================================================
		// Group 16: Comments around / inside the prop don't break extraction
		// ============================================================
		{Code: `<th /* before */ scope="col" /* after */ />`, Tsx: true},
		{Code: `<th scope={/* col */ "col"} />`, Tsx: true},

		// ============================================================
		// Group 17: Listener fires for every variant of "scope" but exempts via th
		// ============================================================
		// Multiple scope attributes on th — each fires independently; both
		// short-circuit at the th branch.
		{Code: `<th scope="row" scope="col" />`, Tsx: true},
		{Code: `<th scope scope="col" />`, Tsx: true},

		// ============================================================
		// Group 18: SVG / MathML primitives — NOT in aria-query's dom map
		// ============================================================
		// aria-query's `dom` map only contains the bare `svg` and `math`
		// roots — primitive shape/text children (rect, circle, path, mn,
		// mo, mi, ...) are absent. Listener fires on scope, IsDOMElement
		// gate skips. Lock in this surprising-but-faithful behavior so a
		// future aria-query upgrade can't silently flip the verdict on a
		// real codebase.
		{Code: `<rect scope />`, Tsx: true},
		{Code: `<circle scope="row" />`, Tsx: true},
		{Code: `<path scope />`, Tsx: true},
		{Code: `<polygon scope />`, Tsx: true},
		{Code: `<polyline scope />`, Tsx: true},
		{Code: `<g scope="col" />`, Tsx: true},
		{Code: `<defs scope />`, Tsx: true},
		{Code: `<use scope />`, Tsx: true},
		{Code: `<ellipse scope />`, Tsx: true},
		{Code: `<line scope />`, Tsx: true},
		{Code: `<text scope />`, Tsx: true},
		// MathML primitives.
		{Code: `<math scope />`, Tsx: true},
		{Code: `<mn scope />`, Tsx: true},
		{Code: `<mo scope />`, Tsx: true},
		{Code: `<mi scope />`, Tsx: true},
		// Modern HTML not in aria-query's older dom map.
		{Code: `<template scope />`, Tsx: true},
		{Code: `<slot scope />`, Tsx: true},

		// ============================================================
		// Group 19: Spread + scope attribute mixing
		// ============================================================
		// Spread BEFORE scope on th — scope still resolves to literal,
		// th-exempt.
		{Code: `<th {...props} scope />`, Tsx: true},
		// Spread AFTER scope on th — same outcome.
		{Code: `<th scope {...props} />`, Tsx: true},
		// Spread BETWEEN multiple attrs on th.
		{Code: `<th id="x" {...props} scope="row" className="y" />`, Tsx: true},
		// Spread with literal object that happens to contain `scope` key on
		// th — the JsxSpreadAttribute is opaque (different AST kind), and
		// the literal scope on the th-exempt element is exempt anyway.
		{Code: `<th {...{name: 'x'}} scope />`, Tsx: true},
		// Multiple spreads + scope on th.
		{Code: `<th {...a} {...b} scope />`, Tsx: true},

		// ============================================================
		// Group 20: Custom container with components map but NO mapping for it
		// ============================================================
		// `<DataTable scope />` with `components: { TableCell: 'td' }` —
		// DataTable not mapped → stays "DataTable" → not in dom → skipped.
		{
			Code: `<DataTable scope />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"components": map[string]interface{}{
						"TableCell": "td",
					},
				},
			},
		},

		// ============================================================
		// Group 21: components map: th value with extra unrelated keys
		// ============================================================
		// Multiple entries in components map — only TableHeader is exempt
		// via th; OtherFoo would map elsewhere if it appeared.
		{
			Code: `<TableHeader scope="col" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"components": map[string]interface{}{
						"TableHeader": "th",
						"OtherFoo":    "div",
						"AnotherBar":  "span",
					},
				},
			},
		},

		// ============================================================
		// Group 22: Component library patterns (sans component map remap)
		// ============================================================
		// Material-UI / Ant Design style — custom components with
		// `component` prop carrying "th" string, but without a
		// polymorphicPropName configured. The rule sees TableCell (not in
		// dom) → skip, regardless of the component prop. Locks in that we
		// don't accidentally inspect `component` / `as` style props
		// without explicit polymorphic configuration.
		{Code: `<TableCell component="th" scope="row" />`, Tsx: true},
		// Same shape with the `as` prop — without polymorphicPropName, the
		// rule should still see TableCell and skip.
		{Code: `<TableCell as="th" scope="row" />`, Tsx: true},

		// ============================================================
		// Group 23: JSX as render prop / children-as-function
		// ============================================================
		// Inner `<th scope>` exempt (it's th); outer DataTable's "render"
		// attribute name !== scope.
		{Code: `<DataTable render={() => <th scope="col" />} />`, Tsx: true},
		// Children-as-function with th-exempt body.
		{Code: `<DataTable>{() => <th scope="row" />}</DataTable>`, Tsx: true},
		// Higher-order render prop with cell-rendering function.
		{
			Code: `<Provider cellRender={(c) => <th scope={c.scope} key={c.id} />} />`,
			Tsx:  true,
		},

		// ============================================================
		// Group 24: th in array literals / list expressions
		// ============================================================
		{Code: `const items = [<th scope="col" key="1" />, <th scope="col" key="2" />];`, Tsx: true},
		{Code: `function App() { return [<th scope="col" key="1" />]; }`, Tsx: true},

		// ============================================================
		// Group 25: th inside higher-level JSX wrappers
		// ============================================================
		// cloneElement / React.createElement-style wrappers don't change
		// the inner element's AST shape. listener still sees the th and exempts.
		{Code: `cloneElement(<th scope />);`, Tsx: true},
		{Code: `Object.freeze(<th scope />);`, Tsx: true},
		// Provider / Context wrapping — listener inspects each JsxAttribute.
		{Code: `<Provider value={data}>{<th scope="row" />}</Provider>`, Tsx: true},
		{Code: `<ErrorBoundary>{<th scope="col" />}</ErrorBoundary>`, Tsx: true},

		// ============================================================
		// Group 26: th with JsxText / JsxExpression children
		// ============================================================
		// JsxText body is irrelevant to the rule; only attribute presence matters.
		{Code: `<th scope="row">Some text</th>`, Tsx: true},
		// JsxExpression body with content.
		{Code: `<th scope="row">{title}</th>`, Tsx: true},
		// Mixed JsxText + JsxExpression body.
		{Code: `<th scope="row">Hello, {name}!</th>`, Tsx: true},
		// Empty body.
		{Code: `<th scope="row"></th>`, Tsx: true},
		// Whitespace-only body.
		{Code: `<th scope="row">   </th>`, Tsx: true},

		// ============================================================
		// Group 27: Whitespace / formatting variations on th
		// ============================================================
		// Extra whitespace around the attribute name / equals.
		{Code: `<th  scope  />`, Tsx: true},
		{Code: `<th scope = "row" />`, Tsx: true},
		{Code: `<th scope ="row"/>`, Tsx: true},
		{Code: "<th\n\tscope\n/>", Tsx: true},
		{Code: "<th\n\tscope =\n\t\t\"row\"\n/>", Tsx: true},
		// JSX with trailing whitespace before closing.
		{Code: `<th scope="row"  />`, Tsx: true},

		// ============================================================
		// Group 28: th in deeply-nested table structures
		// ============================================================
		// 4-level nested table — listener still hits the deeply nested th.
		{
			Code: `<table><thead><tr><th scope="col" /></tr></thead><tbody><tr><th scope="row" /></tr></tbody></table>`,
			Tsx:  true,
		},
		// Conditional inside table body.
		{
			Code: `<table><thead><tr>{cond ? <th scope="col" /> : <th scope="row" />}</tr></thead></table>`,
			Tsx:  true,
		},
		// Map / iteration over headers.
		{
			Code: `<table><thead><tr>{headers.map(h => <th scope="col" key={h.id}>{h.label}</th>)}</tr></thead></table>`,
			Tsx:  true,
		},

		// ============================================================
		// Group 29: th with onclick / event-handler attrs alongside scope
		// ============================================================
		// scope is exempt regardless of sibling attributes; listener still
		// fires for each JsxAttribute but only matches `scope`.
		{Code: `<th scope="col" onClick={fn} onKeyDown={fn} />`, Tsx: true},
		{Code: `<th onClick={fn} scope="col" />`, Tsx: true},
		{Code: `<th id="x" scope="col" className="y" data-test="z" />`, Tsx: true},

		// ============================================================
		// Group 30: Non-trivial polymorphic allowList shapes
		// ============================================================
		// Multi-entry allowList — Box and Container both allow as-swap.
		{
			Code: `<Container as="th" scope="row" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"polymorphicPropName":  "as",
					"polymorphicAllowList": []interface{}{"Box", "Container"},
				},
			},
		},
		// Empty allowList — no as-swap ever applies; Container stays
		// "Container", not in dom → skipped.
		{
			Code: `<Container as="div" scope />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"polymorphicPropName":  "as",
					"polymorphicAllowList": []interface{}{},
				},
			},
		},

		// ============================================================
		// Group 31: polymorphic + components chain (`as` then components map)
		// ============================================================
		// `<Box as="TableHeader" scope="row" />` with polymorphic + components:
		// polymorphic runs first → rawType = "TableHeader"; components map
		// then remaps "TableHeader" → "th" → exempt.
		{
			Code: `<Box as="TableHeader" scope="row" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"polymorphicPropName": "as",
					"components": map[string]interface{}{
						"TableHeader": "th",
					},
				},
			},
		},

		// ============================================================
		// Group 32: th inside JSX expression in attribute value (slot pattern)
		// ============================================================
		// `<Container header={<th scope="col">Name</th>} />` — listener fires
		// on header (not scope) and on scope (th-exempt). Container is custom
		// → header attr doesn't match scope → no report.
		{Code: `<Container header={<th scope="col">Name</th>} />`, Tsx: true},

		// ============================================================
		// Group 33: Non-DOM resolved name via components → th
		// ============================================================
		// Custom component AND non-DOM intermediary that ultimately
		// matches th. Lock in that the chain doesn't break.
		{
			Code: `<CustomHeader scope="row" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"components": map[string]interface{}{
						"CustomHeader": "th",
					},
				},
			},
		},
		// Component map directly: `<Th scope />` mapping Th→th still
		// resolves to "th". Locks in that case-different-but-mapped key
		// works (components map keys are case-sensitive, but the value
		// "th" matches IsDOMElement directly).
		{
			Code: `<Th scope />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"components": map[string]interface{}{
						"Th": "th",
					},
				},
			},
		},

		// ============================================================
		// Group 34: Settings entirely absent (no jsx-a11y key)
		// ============================================================
		// Settings with unrelated keys / no jsx-a11y entry — the rule must
		// not crash and must use raw tag name.
		{
			Code: `<th scope="row" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"other-plugin": map[string]interface{}{"foo": "bar"},
			},
		},
		// Foo without any mapping under non-empty settings (no jsx-a11y) → custom → skipped.
		{
			Code: `<Foo scope />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"some-other": "value",
			},
		},

		// ============================================================
		// Group 35: TS-only JSX shapes — type-args + JSX
		// ============================================================
		// JSX with TS type arguments and th tag. The rule never inspects
		// type args; the attribute is still a plain JsxAttribute.
		{Code: `<List<string, number> scope />`, Tsx: true},
		// Components map keyed on the dotted member-expression string —
		// upstream's `componentMap[finalType]` is a plain key lookup, so
		// dotted keys ARE matched. `<DataGrid.Header>` resolves to the
		// dotted string "DataGrid.Header", the map remaps it to "th",
		// and the th branch exempts it. (See the matching invalid case
		// below where the same shape mapped to "div" reports.)
		{
			Code: `<DataGrid.Header scope="col" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"components": map[string]interface{}{
						"DataGrid.Header": "th",
					},
				},
			},
		},
		// polymorphicAllowList containing only non-string entries — no swap
		// applies; "Other" stays as-is → not in dom → SKIPPED.
		{
			Code: `<Other as="div" scope />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"polymorphicPropName":  "as",
					"polymorphicAllowList": []interface{}{123, true, nil},
				},
			},
		},
		// Resolved-to-th via components map under generic JSX.
		{
			Code: `<List<Header> scope="col" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"components": map[string]interface{}{
						"List": "th",
					},
				},
			},
		},
	}, []rule_tester.InvalidTestCase{
		// ============================================================
		// Group 1: Position assertions — JsxAttribute node range
		// ============================================================
		// Boolean form — JsxAttribute spans columns 6..11 (1-based,
		// end-exclusive), covering exactly `scope`.
		{
			Code: `<div scope />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "scopeOnTh",
				Message:   errorMessage,
				Line:      1, Column: 6, EndLine: 1, EndColumn: 11,
			}},
		},
		// String value — JsxAttribute spans `scope="row"` (11 chars),
		// columns 6..17 (1-based, end-exclusive).
		{
			Code: `<div scope="row" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "scopeOnTh",
				Message:   errorMessage,
				Line:      1, Column: 6, EndLine: 1, EndColumn: 17,
			}},
		},
		// Multi-line attribute — position spans the entire attribute.
		{
			Code: "<div\n  scope={\n    'row'\n  } />",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "scopeOnTh",
				Message:   errorMessage,
				Line:      2, Column: 3, EndLine: 4, EndColumn: 4,
			}},
		},
		// Paired (non-self-closing) element — listener still fires on the
		// JsxAttribute inside the JsxOpeningElement.
		{
			Code: `<div scope>child</div>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "scopeOnTh",
				Message:   errorMessage,
				Line:      1, Column: 6, EndLine: 1, EndColumn: 11,
			}},
		},

		// ============================================================
		// Group 2: Case-insensitive prop name matching → still reports
		// ============================================================
		// On a non-th element, all case variants of "scope" trigger the
		// reporting branch.
		{Code: `<div SCOPE />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div Scope />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div SCoPe="row" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<span scope />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ============================================================
		// Group 3: Listener boundary — nested elements report independently
		// ============================================================
		// Outer `<div scope>` and inner `<span scope />` each emit a
		// diagnostic. Locks in that the listener doesn't dedupe and doesn't
		// bleed across the nesting boundary.
		{
			Code: `<div scope><span scope /></div>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "scopeOnTh", Message: errorMessage, Line: 1, Column: 6, EndLine: 1, EndColumn: 11},
				{MessageId: "scopeOnTh", Message: errorMessage, Line: 1, Column: 18, EndLine: 1, EndColumn: 23},
			},
		},
		// Nested th wrapping non-th: the th's scope attribute is exempt; the
		// inner div's scope attribute reports.
		{
			Code: `<th scope="col"><div scope /></th>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "scopeOnTh", Message: errorMessage, Line: 1, Column: 22, EndLine: 1, EndColumn: 27},
			},
		},
		// Multiple scope attributes on one non-th element each report. tsgo
		// preserves duplicate attributes (legal source, typically a typo);
		// the JsxAttribute listener fires once per matching attribute.
		{
			Code: `<div scope scope />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "scopeOnTh", Message: errorMessage, Line: 1, Column: 6, EndLine: 1, EndColumn: 11},
				{MessageId: "scopeOnTh", Message: errorMessage, Line: 1, Column: 12, EndLine: 1, EndColumn: 17},
			},
		},

		// ============================================================
		// Group 4: Element kind survey — every DOM tag fires
		// ============================================================
		{Code: `<a scope />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<button scope />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<table scope />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<tr scope />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<td scope />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// `<thead>`, `<tbody>`, `<tfoot>` are common confusions — none of
		// these accept the scope attribute per HTML spec.
		{Code: `<thead scope />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<tbody scope />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<tfoot scope />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<input scope />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ============================================================
		// Group 5: scope value variants on a non-th element — all report
		// ============================================================
		// The rule never inspects the value; any present scope attribute on
		// a non-th DOM element triggers regardless of value shape.
		{Code: `<div scope="col" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div scope={"row"} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div scope={someVar} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div scope={null} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div scope={undefined} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div scope={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div scope={false} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div scope={""} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div scope={fn()} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div scope={obj.x} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div scope={obj?.x} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: "<div scope={`row`} />", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: "<div scope={`${x}`} />", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ============================================================
		// Group 6: TS wrappers around the value — irrelevant to rule
		// ============================================================
		{Code: `<div scope={"col" as string} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div scope={("col")} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div scope={"col"!} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div scope={"col" satisfies string} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ============================================================
		// Group 7: components map flips custom INTO scope-rule coverage
		// ============================================================
		// Each of these resolves to a non-th DOM element via the components
		// map → reports.
		{
			Code: `<Cell scope="col" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"components": map[string]interface{}{
						"Cell": "td",
					},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code: `<Wrapper scope="row" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"components": map[string]interface{}{
						"Wrapper": "section",
					},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Group 8: polymorphicPropName resolving to non-th DOM → reports
		// ============================================================
		// `<Box as="div" scope />` resolves to "div" (in dom set, not th) → reports.
		{Code: `<Box as="div" scope />`, Tsx: true, Settings: polymorphicSettings, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<Box as="span" scope="col" />`, Tsx: true, Settings: polymorphicSettings, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// polymorphicAllowList covers Box → swap applies → "div" → reports.
		{Code: `<Box as="div" scope />`, Tsx: true, Settings: polymorphicAllowListSettings, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ============================================================
		// Group 9: TypeScript generic JSX components
		// ============================================================
		// `<List<string> scope />` with components map promoting to a
		// non-th DOM element → reports. Lock-in for tsgo type-arg parsing
		// not breaking JsxAttribute detection.
		{
			Code: `<List<string> scope />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"components": map[string]interface{}{
						"List": "div",
					},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Group 10: Comments around / inside the prop don't suppress
		// ============================================================
		{
			Code: `<div /* a */ scope /* b */ />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "scopeOnTh",
				Message:   errorMessage,
				Line:      1, Column: 14, EndLine: 1, EndColumn: 19,
			}},
		},
		{
			Code:   `<div scope={/* row */ "row"} />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Group 11: Real-world misuse patterns
		// ============================================================
		{
			Code:   `function Header({ title }) { return <h1 scope="col">{title}</h1>; }`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// `<div role="columnheader" scope="col">` — role is the right way to
		// mark a non-th columnheader, but scope is still illegal there.
		{
			Code:   `<div role="columnheader" scope="col">Header</div>`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// `<td scope="col">` — confusion with `<th>`; td does NOT accept scope.
		{
			Code:   `function Cell({ value }) { return <td scope="col">{value}</td>; }`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Group 12: Multiple components in one file
		// ============================================================
		{
			Code: "function A() { return <div scope />; }\nfunction B() { return <span scope />; }",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				expectedError, expectedError,
			},
		},

		// ============================================================
		// Group 13: List-rendered JSX with offending scope
		// ============================================================
		{
			Code:   `const items = arr.map((x, i) => <div key={i} scope={x.scope} />);`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Group 14: Class component with state-driven JSX
		// ============================================================
		{
			Code: `class T extends React.Component { render() { return this.state.ready ? <div scope="row" /> : null; } }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Group 15: Generator / async / IIFE bodies
		// ============================================================
		{
			Code: `function* render() { yield <div scope />; yield <span scope />; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				expectedError, expectedError,
			},
		},
		{
			Code:   `async function render() { return <div scope />; }`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   `const x = (() => <div scope />)();`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Group 16: Fragment + conditional rendering
		// ============================================================
		{
			Code:   `const x = <>{cond && <div scope />}</>;`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code: `function Foo({a, b}) { return a ? <div scope /> : <span scope />; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				expectedError, expectedError,
			},
		},

		// ============================================================
		// Group 17: HOC / forwardRef / memo wrappers carrying scope
		// ============================================================
		{
			Code:   `const Enhanced = withTracking(({ value }) => <div value={value} scope />);`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   `const FocusInput = React.forwardRef((props, ref) => <div ref={ref} scope {...props} />);`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   `const Item = React.memo(({ id }) => <li id={id} scope>{id}</li>);`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Group 18: Extended DOM element survey — every dom-set entry reports
		// ============================================================
		// Sectioning / structural elements.
		{Code: `<article scope />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<section scope />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<aside scope />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<header scope />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<footer scope />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<main scope />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<nav scope />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Headings.
		{Code: `<h1 scope />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<h2 scope />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<h6 scope />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Form elements.
		{Code: `<form scope />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<fieldset scope />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<legend scope />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<label scope />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<select scope />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<option scope />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<textarea scope />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<output scope />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<datalist scope />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<optgroup scope />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<progress scope />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<meter scope />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Media / embedded.
		{Code: `<img scope />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<video scope />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<audio scope />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<picture scope />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<canvas scope />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<iframe scope />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<object scope />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Interactive widgets.
		{Code: `<dialog scope />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<details scope />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<summary scope />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Table-related neighbors of th — all confusable, all illegal.
		{Code: `<caption scope />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<colgroup scope />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<col scope />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ============================================================
		// Group 19: Spread + scope mixing (literal scope still reports)
		// ============================================================
		// Spread before scope on non-th — listener fires on the literal scope.
		{Code: `<div {...props} scope />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Spread after scope on non-th.
		{Code: `<div scope {...props} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Spread sandwich.
		{Code: `<div {...a} scope {...b} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Spread containing literal scope key + a literal scope attribute. Spread
		// is opaque (different AST kind), so only the literal scope reports.
		{Code: `<div {...{scope: 'row'}} scope />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Reverse: literal scope FIRST then spread containing the same key.
		{Code: `<div scope {...{scope: 'col'}} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ============================================================
		// Group 20: Multi-element multi-error scenarios
		// ============================================================
		// 3 sibling elements: th-exempt + 2 reporting → 2 errors.
		{
			Code: `<><th scope /><div scope /><span scope /></>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				expectedError, expectedError,
			},
		},
		// `<thead scope>` itself reports (thead is not th). Inner `<th scope>`
		// is exempt — only one error from thead.
		{
			Code: `<thead scope><tr><th scope="col" /></tr></thead>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// 3-deep wrapper containing both reporting and exempt forms.
		{
			Code: `<table><thead scope><tr><th scope="col" /><td scope /></tr></thead></table>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				expectedError, expectedError, // thead + td
			},
		},
		// JsxFragment containing 3 reporting + 1 exempt.
		{
			Code: `<><div scope /><span scope /><th scope /><a scope /></>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				expectedError, expectedError, expectedError,
			},
		},

		// ============================================================
		// Group 21: JSX as render prop / children-as-function (offending)
		// ============================================================
		// Inner non-th JSX inside render prop reports; outer DataTable's "render"
		// attribute is irrelevant (name !== scope).
		{
			Code:   `<DataTable render={() => <div scope />} />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// Children-as-function with offending body.
		{
			Code:   `<DataTable>{() => <div scope />}</DataTable>`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// Render prop with both exempt + offending in arms.
		{
			Code: `<Provider cellRender={(c) => c.header ? <th scope="col" /> : <div scope="col" />} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Group 22: JSX in array / iterator literals
		// ============================================================
		{
			Code: `const items = [<div scope key="1" />, <span scope key="2" />, <th scope key="3" />];`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				expectedError, expectedError,
			},
		},
		{
			Code: `function App() { return [<div scope key="a" />, <th scope key="b" />]; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Group 23: cloneElement / wrapper patterns with offending JSX
		// ============================================================
		{
			Code:   `cloneElement(<div scope />);`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   `Object.freeze(<div scope />);`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// Wrapped in Provider → inner div reports.
		{
			Code:   `<Provider value={data}>{<div scope />}</Provider>`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// JSX as object literal value.
		{
			Code:   `const obj = { content: <div scope /> };`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Group 24: Whitespace / formatting variations on non-th
		// ============================================================
		{
			Code:   `<div  scope  />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   `<div scope = "row" />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   "<div\n\tscope\n/>",
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   "<div\n\tscope =\n\t\t\"row\"\n/>",
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Group 25: scope attribute at varied positions
		// ============================================================
		// scope at start.
		{
			Code:   `<div scope id="x" className="y" />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// scope in the middle.
		{
			Code:   `<div id="x" scope className="y" />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// scope at end.
		{
			Code:   `<div id="x" className="y" scope />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// scope between event handlers.
		{
			Code:   `<div onClick={fn} scope onKeyDown={fn} />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Group 26: th with offending non-scope sibling element
		// ============================================================
		// `<th scope>` is exempt; sibling `<td scope>` reports.
		{
			Code:   `<><th scope="col" /><td scope="row" /></>`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// Sibling-of-th tags don't inherit th-exemption.
		{
			Code:   `<><th scope /><tr scope /><td scope /></>`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError, expectedError},
		},

		// ============================================================
		// Group 27: Component library patterns with components map
		// ============================================================
		// Material-UI style: TableCell mapped to td → reports.
		{
			Code: `<TableCell scope="col" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"components": map[string]interface{}{
						"TableCell": "td",
					},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// ============================================================
		// Group 28: polymorphic + components chain (`as` then components → non-th)
		// ============================================================
		// `<Box as="DataCell" scope />` with `components: { DataCell: 'td' }`:
		// polymorphic → "DataCell" → components → "td" (in dom, not th) → reports.
		{
			Code: `<Box as="DataCell" scope />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"polymorphicPropName": "as",
					"components": map[string]interface{}{
						"DataCell": "td",
					},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Group 29: th with sibling spread + scope on non-th sibling
		// ============================================================
		{
			Code:   `<><th {...props} scope /><div {...props} scope /></>`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Group 30: Settings with empty components / polymorphic config
		// ============================================================
		// Empty components map — defensive; rawType stays as raw tag name.
		{
			Code: `<div scope />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"components": map[string]interface{}{},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// Empty polymorphic config — no swap; div-scope still reports.
		{
			Code: `<div as="th" scope />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"polymorphicPropName": "",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Group 31: Suspense / lazy-loaded boundaries
		// ============================================================
		// scope inside Suspense fallback / boundary — listener still fires.
		{
			Code:   `<Suspense fallback={<div scope />}><div>x</div></Suspense>`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   `<Suspense><div scope /></Suspense>`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Group 32: ErrorBoundary / Provider wrapping an offending element
		// ============================================================
		{
			Code:   `<ErrorBoundary><div scope /></ErrorBoundary>`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   `<Context.Provider value={data}><div scope /></Context.Provider>`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Group 33: TS-only JSX shapes — reports through type wrappers
		// ============================================================
		// JSX-as-expression in const assertion: scope still reachable.
		{
			Code:   `const x = <div scope /> as React.ReactNode;`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// JSX in satisfies clause.
		{
			Code:   `const x = (<div scope />) satisfies any;`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// JSX in non-null assertion.
		{
			Code:   `const x = (<div scope />)!;`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Group 34: TypeScript generic JSX with components remap
		// ============================================================
		// `<List<Header> scope />` with `components: { List: 'div' }` →
		// rawType "List" → "div" → not th → reports.
		{
			Code: `<List<Header> scope />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"components": map[string]interface{}{
						"List": "div",
					},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// Multi-arg type generic.
		{
			Code: `<Cell<Row, Col> scope />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"components": map[string]interface{}{
						"Cell": "td",
					},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Group 35: Complex realistic table cell misuse patterns
		// ============================================================
		// Simulated form table — every cell incorrectly carries scope.
		{
			Code: `<table><tr><td scope="col" /><td scope="col" /></tr><tr><td scope="row" /><td scope="row" /></tr></table>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				expectedError, expectedError, expectedError, expectedError,
			},
		},
		// Mixed cells: some th (exempt), some td (reports).
		{
			Code: `<table><tr><th scope="col" /><td scope="col" /><th scope="col" /><td scope="col" /></tr></table>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError, expectedError},
		},

		// ============================================================
		// Group 36: scope inside iterator / map with destructured items
		// ============================================================
		{
			Code: `const cells = data.map(({ scope, ...rest }, i) => <td key={i} scope={scope} {...rest} />);`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// flatMap producing offending elements.
		{
			Code: `const cells = rows.flatMap(r => [<td scope="col" key={r.id} />, <span />]);`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Group 37: try / catch / finally JSX bodies
		// ============================================================
		{
			Code:   `function App() { try { return <div scope />; } catch (e) { return <span scope />; } }`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError, expectedError},
		},

		// ============================================================
		// Group 38: switch statement returning JSX
		// ============================================================
		{
			Code: `function App({type}) { switch(type) { case 'a': return <div scope />; case 'b': return <span scope />; default: return null; } }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError, expectedError},
		},

		// ============================================================
		// Group 39: Object methods / class methods returning JSX
		// ============================================================
		{
			Code:   `const obj = { render() { return <div scope />; } };`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   `class C { static factory() { return <div scope />; } }`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// Class with class-field arrow returning JSX.
		{
			Code:   `class C { render = () => <div scope />; }`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Group 40: Computed JSX expressions in attribute values + offending sibling
		// ============================================================
		{
			Code: `<div onClick={() => alert(<span scope />)} scope />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				expectedError, expectedError,
			},
		},
		// Two offending elements where one is in attribute and one is the parent.
		{
			Code: `<div header={<span scope />} scope />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				expectedError, expectedError,
			},
		},

		// ============================================================
		// Group 41: Member-expression tag remapped via dotted components key
		// ============================================================
		// Counterpart to the `<DataGrid.Header>` valid case above: the
		// components map's key lookup is plain `map[finalType]`, which
		// works for any string including dotted ones. Mapping the dotted
		// rawType to a non-th DOM element flips it INTO scope-rule coverage.
		{
			Code: `<DataGrid.Header scope="col" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"components": map[string]interface{}{
						"DataGrid.Header": "div",
					},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// Same with multi-segment dotted name.
		{
			Code: `<UI.Table.Cell scope="row" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"components": map[string]interface{}{
						"UI.Table.Cell": "td",
					},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// Namespaced tag mapped via the composite key — same plain-lookup
		// path; `<svg:circle>` → "svg:circle" → "div" → reports.
		{
			Code: `<svg:circle scope />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"components": map[string]interface{}{
						"svg:circle": "div",
					},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Group 42: scope on Web Components (hyphenated tags) via components
		// ============================================================
		// `<my-table scope />` with components map `my-table → td`. Even
		// though hyphenated tags are valid JSX, the components map keys
		// are upstream-checked via `tag in components`, which works for
		// hyphenated strings. Locks in we don't accidentally normalize away
		// the hyphen.
		{
			Code: `<my-table scope />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"components": map[string]interface{}{
						"my-table": "td",
					},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Group 43: Nested polymorphic + various tag-name AST shapes
		// ============================================================
		// `<Box as={"th"} />` — JsxExpression with StringLiteral as the as
		// value. polymorphic resolves the string → "th" → exempt → no report.
		// (Locked in valid section already.)
		// Reverse: `<Box as={"div"} scope />` resolves to "div" → reports.
		{
			Code: `<Box as={"div"} scope />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"polymorphicPropName": "as",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// `<Box as={\`th\`} />` — NoSubstitutionTemplateLiteral, also
		// extracts to literal "th" via literal pop value. Lock in via
		// the reverse direction (div).
		{
			Code: "<Box as={`div`} scope />",
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"polymorphicPropName": "as",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Group 44: Defensive — settings that would crash if not handled
		// ============================================================
		// `components` is not a map — IsDOMElement falls back to raw tag.
		{
			Code: `<div scope />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"components": "invalid",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// `polymorphicPropName` is a number — silently skip polymorphic.
		{
			Code: `<div as="th" scope />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"polymorphicPropName": 123,
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// ============================================================
		// Group 45: Newline-heavy attribute spans (multi-line position)
		// ============================================================
		{
			Code: "<div\nscope=\n  {value}\n/>",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "scopeOnTh",
				Message:   errorMessage,
				Line:      2, Column: 1, EndLine: 3, EndColumn: 10,
			}},
		},
		// Tab-indented multi-line.
		{
			Code: "<div\n\t\tscope={value}\n\t/>",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "scopeOnTh",
				Message:   errorMessage,
				Line:      2, Column: 3, EndLine: 2, EndColumn: 16,
			}},
		},
	})
}

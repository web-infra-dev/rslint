package no_noninteractive_tabindex

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// polymorphicSettings exercises the `<Box as="...">` resolution path inside
// jsxa11yutil.GetElementType. `<Box as="article" tabIndex="0" />` becomes
// effectively `<article tabIndex="0" />` and trips the rule.
var polymorphicSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"polymorphicPropName": "as",
	},
}

// polymorphicAllowListSettings restricts the `as` swap to a specific list.
// Outside the list, the original tag name is kept.
var polymorphicAllowListSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"polymorphicPropName":  "as",
		"polymorphicAllowList": []interface{}{"Box"},
	},
}

// componentsToCustomSettings remaps a custom component to ANOTHER custom
// (non-DOM) name — the rule still skips because the resolved name isn't
// in dom set.
var componentsToCustomSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"components": map[string]interface{}{
			"Foo": "Bar",
		},
	},
}

// emptyJsxA11ySettings exercises the defensive paths in
// jsxa11yutil.GetElementType / IsDOMElement when the settings tree exists
// but is empty.
var emptyJsxA11ySettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{},
}

// tagsExemptOption exercises the `tags` escape hatch — `<div tabIndex="0" />`
// would normally report, but `tags: ["div"]` short-circuits.
var tagsExemptOption = []interface{}{
	map[string]interface{}{
		"tags": []interface{}{"div"},
	},
}

// rolesExemptTabpanelOption mirrors `roles: ["tabpanel"]` alone (no other
// option). Used to verify the per-option default semantics — `tags` absent
// means the `tags && includes(...)` guard short-circuits via a falsy LHS.
var rolesExemptTabpanelOption = []interface{}{
	map[string]interface{}{
		"roles": []interface{}{"tabpanel"},
	},
}

// allOptionsCombo exercises all three options at once on top of the
// recommended preset shape, mirroring real user configs that override all
// dimensions.
var allOptionsCombo = []interface{}{
	map[string]interface{}{
		"tags":                  []interface{}{"div", "section"},
		"roles":                 []interface{}{"tabpanel", "alertdialog"},
		"allowExpressionValues": true,
	},
}

// TestNoNoninteractiveTabindexExtras is the catch-all suite for cases
// beyond upstream's small valid/invalid matrix. Groups span:
//
//  1. **Inherent-interactive element survey** — every `dom` schema that
//     matches an interactive role / AX-object table entry.
//  2. **Interactive role survey** — every widget role + multi-token /
//     case-insensitive variants + non-interactive equivalents.
//  3. **tabIndex value survey** — literal-numeric, literal-string,
//     boolean, empty, NaN, Infinity, BigInt, hex / octal / binary, numeric
//     separator, scientific notation, expression-fallback (BinaryExpression,
//     ConditionalExpression, LogicalExpression, NullishCoalescing,
//     PrefixUnary), TS wrappers, parens.
//  4. **Position assertions per container** — JsxAttribute span on
//     self-closing / paired / multi-line forms.
//  5. **Listener boundary** — repeats across nested elements; per-element
//     diagnostic isolation; multi-attribute repeats.
//  6. **Options coverage matrix** — bare-object / array-wrapped / empty /
//     malformed; the three options individually + combined.
//  7. **Settings × resolution** — components, polymorphicPropName,
//     polymorphicAllowList, and combinations.
//  8. **Real-world a11y patterns** — Modal/Dialog, Tabs widget,
//     Combobox/Listbox, Custom dropdown, Tree, Toolbar, focus management.
//  9. **Real-world component patterns** — forwardRef / memo / HOC /
//     generators / async / IIFE / class / fragment / map / multi-component.
//  10. **Spread literal** — `{...{role: "..."}}` shapes.
//  11. **TypeScript JSX** — generics, satisfies, non-null, namespaced,
//      long-chain member expression.
func TestNoNoninteractiveTabindexExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoNoninteractiveTabindexRule, []rule_tester.ValidTestCase{
		// ============================================================
		// Group 1: Inherent-interactive elements (skip via IsInteractiveElement)
		// ============================================================
		// interactiveElementRoleSchemas matches.
		{Code: `<button tabIndex={0} />`, Tsx: true},
		{Code: `<button type="submit" tabIndex={0} />`, Tsx: true},
		{Code: `<input tabIndex={0} />`, Tsx: true},
		{Code: `<input type="text" tabIndex={0} />`, Tsx: true},
		{Code: `<input type="button" tabIndex={0} />`, Tsx: true},
		{Code: `<input type="submit" tabIndex={0} />`, Tsx: true},
		{Code: `<input type="reset" tabIndex={0} />`, Tsx: true},
		{Code: `<input type="image" tabIndex={0} />`, Tsx: true},
		{Code: `<input type="checkbox" tabIndex={0} />`, Tsx: true},
		{Code: `<input type="radio" tabIndex={0} />`, Tsx: true},
		{Code: `<input type="range" tabIndex={0} />`, Tsx: true},
		{Code: `<input type="number" tabIndex={0} />`, Tsx: true},
		{Code: `<input type="email" list="x" tabIndex={0} />`, Tsx: true},
		{Code: `<input type="search" list="x" tabIndex={0} />`, Tsx: true},
		{Code: `<input type="tel" list="x" tabIndex={0} />`, Tsx: true},
		{Code: `<input type="url" list="x" tabIndex={0} />`, Tsx: true},
		{Code: `<textarea tabIndex={0} />`, Tsx: true},
		{Code: `<select tabIndex={0} />`, Tsx: true},
		{Code: `<select multiple size={2} tabIndex={0} />`, Tsx: true},
		{Code: `<select multiple tabIndex={0} />`, Tsx: true},
		{Code: `<select size={2} tabIndex={0} />`, Tsx: true},
		// `<a href>` is interactive — `<a />` (no href) is NON-interactive
		// and would report; covered in invalid below.
		{Code: `<a href="https://example.com" tabIndex={0} />`, Tsx: true},
		{Code: `<a href="" tabIndex={0} />`, Tsx: true},
		{Code: `<a href={url} tabIndex={0} />`, Tsx: true},
		// Boolean form `<a href />` — schema needs only "href present", which
		// boolean-form satisfies (interactive).
		{Code: `<a href tabIndex={0} />`, Tsx: true},
		// `<area href>` is interactive (link).
		{Code: `<area href="x" tabIndex={0} />`, Tsx: true},
		// Tabular elements — th/td/tr all match interactive role schemas.
		{Code: `<th tabIndex={0} />`, Tsx: true},
		{Code: `<th scope="col" tabIndex={0} />`, Tsx: true},
		{Code: `<th scope="colgroup" tabIndex={0} />`, Tsx: true},
		{Code: `<th scope="row" tabIndex={0} />`, Tsx: true},
		{Code: `<th scope="rowgroup" tabIndex={0} />`, Tsx: true},
		{Code: `<td tabIndex={0} />`, Tsx: true},
		{Code: `<tr tabIndex={0} />`, Tsx: true},
		// `<datalist>`, `<option>` — interactive role schemas.
		{Code: `<datalist tabIndex={0} />`, Tsx: true},
		{Code: `<option tabIndex={0} />`, Tsx: true},
		// AX-only tagged interactive (audio/video/canvas/embed/menuitem/summary).
		{Code: `<audio tabIndex={0} />`, Tsx: true},
		{Code: `<video tabIndex={0} />`, Tsx: true},
		{Code: `<canvas tabIndex={0} />`, Tsx: true},
		{Code: `<embed tabIndex={0} />`, Tsx: true},
		{Code: `<menuitem tabIndex={0} />`, Tsx: true},
		{Code: `<summary tabIndex={0} />`, Tsx: true},
		{Code: `<input type="color" tabIndex={0} />`, Tsx: true},
		{Code: `<input type="date" tabIndex={0} />`, Tsx: true},
		{Code: `<input type="time" tabIndex={0} />`, Tsx: true},
		{Code: `<input type="datetime" tabIndex={0} />`, Tsx: true},

		// ============================================================
		// Group 2: Interactive role overrides on non-interactive elements
		// ============================================================
		// Every concrete widget role from interactiveRolesSet.
		{Code: `<div role="button" tabIndex={0} />`, Tsx: true},
		{Code: `<div role="checkbox" tabIndex={0} />`, Tsx: true},
		{Code: `<div role="columnheader" tabIndex={0} />`, Tsx: true},
		{Code: `<div role="combobox" tabIndex={0} />`, Tsx: true},
		{Code: `<div role="grid" tabIndex={0} />`, Tsx: true},
		{Code: `<div role="gridcell" tabIndex={0} />`, Tsx: true},
		{Code: `<div role="link" tabIndex={0} />`, Tsx: true},
		{Code: `<div role="listbox" tabIndex={0} />`, Tsx: true},
		{Code: `<div role="menu" tabIndex={0} />`, Tsx: true},
		{Code: `<div role="menubar" tabIndex={0} />`, Tsx: true},
		{Code: `<div role="menuitem" tabIndex={0} />`, Tsx: true},
		{Code: `<div role="menuitemcheckbox" tabIndex={0} />`, Tsx: true},
		{Code: `<div role="menuitemradio" tabIndex={0} />`, Tsx: true},
		{Code: `<div role="option" tabIndex={0} />`, Tsx: true},
		{Code: `<div role="radio" tabIndex={0} />`, Tsx: true},
		{Code: `<div role="radiogroup" tabIndex={0} />`, Tsx: true},
		{Code: `<div role="row" tabIndex={0} />`, Tsx: true},
		{Code: `<div role="rowheader" tabIndex={0} />`, Tsx: true},
		{Code: `<div role="scrollbar" tabIndex={0} />`, Tsx: true},
		{Code: `<div role="searchbox" tabIndex={0} />`, Tsx: true},
		{Code: `<div role="slider" tabIndex={0} />`, Tsx: true},
		{Code: `<div role="spinbutton" tabIndex={0} />`, Tsx: true},
		{Code: `<div role="switch" tabIndex={0} />`, Tsx: true},
		{Code: `<div role="tab" tabIndex={0} />`, Tsx: true},
		{Code: `<div role="tablist" tabIndex={0} />`, Tsx: true},
		{Code: `<div role="textbox" tabIndex={0} />`, Tsx: true},
		{Code: `<div role="tree" tabIndex={0} />`, Tsx: true},
		{Code: `<div role="treegrid" tabIndex={0} />`, Tsx: true},
		{Code: `<div role="treeitem" tabIndex={0} />`, Tsx: true},
		// `toolbar` is interactive per upstream (treated as widget despite
		// not descending from widget — supports aria-activedescendant).
		{Code: `<div role="toolbar" tabIndex={0} />`, Tsx: true},
		// Multi-token role: first token wins. `button menu` → button (interactive).
		{Code: `<div role="button menu" tabIndex={0} />`, Tsx: true},
		{Code: `<div role="link button" tabIndex={0} />`, Tsx: true},
		// Trailing space → split produces ["button", ""], "button" matches first.
		{Code: `<div role="button " tabIndex={0} />`, Tsx: true},
		// Case-insensitive on the role value (lowercased before match).
		{Code: `<div role="BUTTON" tabIndex={0} />`, Tsx: true},
		{Code: `<div role="Button" tabIndex={0} />`, Tsx: true},
		{Code: `<div role="MENUITEM" tabIndex={0} />`, Tsx: true},

		// ============================================================
		// Group 3: tabIndex < 0 — always allowed
		// ============================================================
		{Code: `<div tabIndex={-1} />`, Tsx: true},
		{Code: `<div tabIndex={-100} />`, Tsx: true},
		{Code: `<div tabIndex={-2147483648} />`, Tsx: true}, // INT32_MIN
		{Code: `<article tabIndex="-2" />`, Tsx: true},
		{Code: `<section tabIndex="-1" />`, Tsx: true},
		// String "-1" with surrounding whitespace — Number coerces.
		{Code: `<div tabIndex=" -1 " />`, Tsx: true},
		// Numeric expression resolved by staticEval.
		{Code: `<div tabIndex={-1 - 0} />`, Tsx: true},
		// Conditional with both branches < 0.
		{Code: `<div tabIndex={cond ? -1 : -2} />`, Tsx: true},
		// LogicalExpression resolving to negative.
		{Code: `<div tabIndex={x || -1} />`, Tsx: true},

		// ============================================================
		// Group 4: tabIndex shapes that resolve to "undefined" (skip)
		// ============================================================
		// Boolean attribute form `<div tabIndex />` extracts to JS true →
		// upstream's getTabIndex routes that to `=== true` → undefined.
		{Code: `<div tabIndex />`, Tsx: true},
		// Explicit boolean values → undefined (step-1 boolean arm).
		{Code: `<div tabIndex={true} />`, Tsx: true},
		{Code: `<div tabIndex={false} />`, Tsx: true},
		// String "true" / "false" coerces to boolean via jsxAstUtilsLiteralCoerce
		// → step-1 boolean → undefined.
		{Code: `<div tabIndex="true" />`, Tsx: true},
		{Code: `<div tabIndex="false" />`, Tsx: true},
		{Code: `<div tabIndex="True" />`, Tsx: true},
		{Code: `<div tabIndex="FALSE" />`, Tsx: true},
		// Empty-string value → upstream's `literalValue.length === 0` → undefined.
		{Code: `<div tabIndex="" />`, Tsx: true},
		{Code: `<div tabIndex={""} />`, Tsx: true},
		// Non-integer numbers → upstream's `Number.isInteger(value) ? : undefined`.
		{Code: `<div tabIndex={1.5} />`, Tsx: true},
		{Code: `<div tabIndex={-0.5} />`, Tsx: true},
		{Code: `<div tabIndex={1.5e0} />`, Tsx: true},
		{Code: `<div tabIndex={1e-5} />`, Tsx: true},
		{Code: `<div tabIndex="1.5" />`, Tsx: true},
		{Code: `<div tabIndex="0.5" />`, Tsx: true},
		// BigInt — upstream Number(BigInt) coerces to a usable Number,
		// so `0n` / `1n` reach `>= 0` true → REPORT. Lock-ins in the
		// invalid section. Only negative BigInts skip here.
		{Code: `<div tabIndex={-1n} />`, Tsx: true}, // Number(-1n) = -1, -1 >= 0 false → skip
		// Identifier / runtime expression — staticEval returns string fallback,
		// parseFloat fails → undefined.
		{Code: `<div tabIndex={someVar} />`, Tsx: true},
		{Code: `<div tabIndex={fn()} />`, Tsx: true},
		{Code: `<div tabIndex={obj.x} />`, Tsx: true},
		{Code: `<div tabIndex={obj?.x} />`, Tsx: true},
		// Pure non-numeric string.
		{Code: `<div tabIndex="abc" />`, Tsx: true},
		{Code: `<div tabIndex="1abc" />`, Tsx: true}, // ParseFloat partial-match returns 1 in JS but err in Go strconv → false
		// `-Infinity` < 0 → not reported. `Infinity` is reported (covered
		// in invalid below — `Infinity >= 0` is true so step-2 ToNumber
		// path triggers). `NaN` identifier resolves to non-numeric jvString
		// "NaN" → ParseFloat fails → undefined → not reported.
		{Code: `<div tabIndex={-Infinity} />`, Tsx: true},
		{Code: `<div tabIndex={NaN} />`, Tsx: true},
		// Signed hex / octal / binary strings: JS Number rejects sign on
		// these prefixes → NaN → undefined → not reported.
		{Code: `<div tabIndex="-0x10" />`, Tsx: true},
		{Code: `<div tabIndex="+0x10" />`, Tsx: true},
		{Code: `<div tabIndex="-0o10" />`, Tsx: true},
		{Code: `<div tabIndex="-0b10" />`, Tsx: true},
		// Malformed hex / oct / bin → ParseUint fails → undefined.
		{Code: `<div tabIndex="0x" />`, Tsx: true},
		{Code: `<div tabIndex="0xZZ" />`, Tsx: true},
		{Code: `<div tabIndex="0o89" />`, Tsx: true}, // 8/9 not valid octal
		{Code: `<div tabIndex="0b12" />`, Tsx: true}, // 2 not valid binary
		// Non-integer that ParseFloat decodes successfully but isInteger
		// rejects (step-1 path enforces integer).
		{Code: `<div tabIndex="0.5" />`, Tsx: true},
		// ArrayLiteralExpression — toString → ToNumber yields negative or NaN.
		{Code: `<div tabIndex={[1, 2]} />`, Tsx: true},   // "1,2" → NaN
		{Code: `<div tabIndex={[-1]} />`, Tsx: true},     // "-1" → -1
		{Code: `<div tabIndex={[-1.5]} />`, Tsx: true},   // "-1.5" → -1.5
		{Code: `<div tabIndex={[NaN]} />`, Tsx: true},    // "NaN" → NaN
		{Code: `<div tabIndex={[5, "x"]} />`, Tsx: true}, // "5,x" → NaN
		// Unary `+`/`-` on string operand (step-2 fallback, since
		// staticEvalUnary's number-only guard returns null for strings).
		{Code: `<div tabIndex={-'5'} />`, Tsx: true}, // -5
		{Code: `<div tabIndex={+'abc'} />`, Tsx: true},
		{Code: `<div tabIndex={-'abc'} />`, Tsx: true},
		{Code: `<div tabIndex={+'-5'} />`, Tsx: true}, // +(-5) = -5
		// Bitwise NOT — staticEval already handles, just lock the negative result.
		{Code: `<div tabIndex={~5} />`, Tsx: true}, // -6
		// `<div tabIndex={} />` — empty JsxExpression / JSXEmptyExpression.
		// Upstream's TYPES has no entry → noop → null → `null >= 0` true →
		// REPORT. Moved to invalid section below.
		// undefined identifier → step 1 jvUndef (literalPropValue) → not in
		// switch → step 2 staticEval → jvUndef → not handled → skip.
		{Code: `<div tabIndex={undefined} />`, Tsx: true},
		// null literal → step 1 jvString "null" (LITERAL_TYPES special-case) →
		// "null" parseFloat fails → false.
		{Code: `<div tabIndex={null} />`, Tsx: true},

		// ============================================================
		// Group 5: TS wrappers around the tabIndex value (negative)
		// ============================================================
		// Parens unwrap inside attributeInnerExpression; staticEval skips
		// TS-wrapper kinds via skipTransparent so the inner literal is reached.
		{Code: `<div tabIndex={(-1)} />`, Tsx: true},
		{Code: `<div tabIndex={((-1))} />`, Tsx: true},
		{Code: `<div tabIndex={(-1) as number} />`, Tsx: true},
		{Code: `<div tabIndex={-1 as number} />`, Tsx: true},
		{Code: `<div tabIndex={(-1)!} />`, Tsx: true},
		// Note: `<div tabIndex={-1 satisfies number} />` is INVALID per
		// upstream (TSSatisfiesExpression → null → `null >= 0` true →
		// REPORT). It's covered in the invalid section as part of the
		// opaque-expression-types lock-in group.
		// Parenthesized string literal coerces to integer (negative).
		{Code: `<div tabIndex={("-1")} />`, Tsx: true},
		{Code: `<div tabIndex={("-1") as string} />`, Tsx: true},
		// Type assertion via legacy `<T>x` — also skipTransparent.
		// (Not legal in .tsx so not tested directly; covered conceptually
		//  via OEKTypeAssertions in the skipTransparent constant.)

		// ============================================================
		// Group 6: tags option exempts the resolved element name
		// ============================================================
		{Code: `<div tabIndex="0" />`, Tsx: true, Options: tagsExemptOption},
		{Code: `<div role="article" tabIndex="0" />`, Tsx: true, Options: tagsExemptOption},
		// tags + components map: `<MyDiv>` resolves to `div`, then tags exempt.
		{
			Code:    `<MyDiv tabIndex="0" />`,
			Tsx:     true,
			Options: tagsExemptOption,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"components": map[string]interface{}{"MyDiv": "div"},
				},
			},
		},
		// Multiple entries in tags.
		{Code: `<section tabIndex="0" />`, Tsx: true, Options: []interface{}{
			map[string]interface{}{"tags": []interface{}{"div", "section"}},
		}},

		// ============================================================
		// Group 7: roles option exempts a literal role value
		// ============================================================
		{Code: `<div role="tabpanel" tabIndex="0" />`, Tsx: true, Options: rolesExemptTabpanelOption},
		// Multiple roles in the option list.
		{Code: `<div role="presentation" tabIndex="0" />`, Tsx: true, Options: []interface{}{
			map[string]interface{}{"roles": []interface{}{"tabpanel", "presentation"}},
		}},
		// roles option matches even on non-interactive elements that would
		// otherwise report — the role escape hatch beats element classification.
		{Code: `<article role="tabpanel" tabIndex="0" />`, Tsx: true, Options: rolesExemptTabpanelOption},

		// ============================================================
		// Group 8: allowExpressionValues escape hatch
		// ============================================================
		{Code: `<div role={SOME_ROLE} tabIndex="0" />`, Tsx: true, Options: allowExpressionValuesTrueOptions},
		{Code: `<div role={getRole()} tabIndex="0" />`, Tsx: true, Options: allowExpressionValuesTrueOptions},
		{Code: `<div role={obj.role} tabIndex="0" />`, Tsx: true, Options: allowExpressionValuesTrueOptions},
		{Code: `<div role={obj?.role} tabIndex="0" />`, Tsx: true, Options: allowExpressionValuesTrueOptions},
		{Code: "<div role={`button`} tabIndex=\"0\" />", Tsx: true, Options: allowExpressionValuesTrueOptions},
		{Code: "<div role={`${dyn}`} tabIndex=\"0\" />", Tsx: true, Options: allowExpressionValuesTrueOptions},
		// Conditional with mixed arms — exempt under allowExpressionValues=true.
		{Code: `<div role={cond ? "button" : OTHER} tabIndex="0" />`, Tsx: true, Options: allowExpressionValuesTrueOptions},
		{Code: `<div role={cond ? Roles.A : Roles.B} tabIndex="0" />`, Tsx: true, Options: allowExpressionValuesTrueOptions},
		// JsxExpression with Literal inside is non-literal per upstream
		// `JSXExpressionContainer` arm.
		{Code: `<div role={"button"} tabIndex="0" />`, Tsx: true, Options: allowExpressionValuesTrueOptions},
		// Logical / Nullish coalescing — also non-literal.
		{Code: `<div role={r || "button"} tabIndex="0" />`, Tsx: true, Options: allowExpressionValuesTrueOptions},
		{Code: `<div role={r ?? "button"} tabIndex="0" />`, Tsx: true, Options: allowExpressionValuesTrueOptions},
		{Code: `<div role={r && "button"} tabIndex="0" />`, Tsx: true, Options: allowExpressionValuesTrueOptions},

		// ============================================================
		// Group 9: Spread attribute — literal-spread covered by FindAttributeByName
		// ============================================================
		// Plain spread of an unknown identifier — listener still walks
		// attribute list, no JsxAttribute named tabIndex matches → skip.
		{Code: `<div {...props} />`, Tsx: true},
		// Multiple spreads + an explicit tabIndex below 0.
		{Code: `<div {...a} {...b} tabIndex={-1} />`, Tsx: true},
		// Literal spread defining tabIndex with negative value — FindAttributeByName
		// walks into the ObjectLiteralExpression and finds the property; staticEval
		// resolves -1.
		{Code: `<div {...{tabIndex: -1}} />`, Tsx: true},
		// Literal spread defining role on an interactive role.
		{Code: `<div {...{role: "button", tabIndex: 0}} />`, Tsx: true},

		// ============================================================
		// Group 10: Comments around / inside the prop
		// ============================================================
		{Code: `<div /* before */ tabIndex={-1} /* after */ />`, Tsx: true},
		{Code: `<div tabIndex={/* explicit */ -1} />`, Tsx: true},
		{Code: `<div role={/* literal-string */ "button"} tabIndex={0} />`, Tsx: true},
		{Code: "<div\n  tabIndex={\n    -1 // negative\n  }\n/>", Tsx: true},

		// ============================================================
		// Group 11: TypeScript generic JSX components
		// ============================================================
		{Code: `<List<string> tabIndex={0} />`, Tsx: true},
		{Code: `<Cell<{a: number}> tabIndex={0} />`, Tsx: true},
		{Code: `<Map<string, number> tabIndex={0} />`, Tsx: true},

		// ============================================================
		// Group 12: Long-chain member-expression / namespaced tags
		// ============================================================
		// `<UX.Layout>`, `<svg:circle>`, `<this.Foo>` — the resolved tag is
		// the dotted / colon string, which is not in the dom set → skip.
		{Code: `<UX.Layout tabIndex={0} />`, Tsx: true},
		{Code: `<svg:circle tabIndex={0} />`, Tsx: true},
		{Code: `<this.Foo tabIndex={0} />`, Tsx: true},
		{Code: `<Foo.Bar.Baz tabIndex={0} />`, Tsx: true},
		{Code: `<a.b.c.d.e tabIndex={0} />`, Tsx: true},

		// ============================================================
		// Group 13: Hyphenated DOM tags (web components)
		// ============================================================
		// `<my-element>` is lowercased; `dom.get('my-element')` is undefined
		// → not in dom set → skip.
		{Code: `<my-element tabIndex={0} />`, Tsx: true},
		{Code: `<x-y-z tabIndex={0} />`, Tsx: true},

		// ============================================================
		// Group 14: polymorphicPropName remap to interactive tag
		// ============================================================
		{Code: `<Box as="button" tabIndex={0} />`, Tsx: true, Settings: polymorphicSettings},
		{Code: `<Box as="a" href="x" tabIndex={0} />`, Tsx: true, Settings: polymorphicSettings},
		{Code: `<Box as="textarea" tabIndex={0} />`, Tsx: true, Settings: polymorphicSettings},
		{Code: `<Box as="select" tabIndex={0} />`, Tsx: true, Settings: polymorphicSettings},
		{Code: `<Box as="input" type="checkbox" tabIndex={0} />`, Tsx: true, Settings: polymorphicSettings},
		// polymorphicAllowList: `Other` not in list → swap skipped → resolved
		// stays `Other` → not in dom set → skip.
		{Code: `<Other as="article" tabIndex={0} />`, Tsx: true, Settings: polymorphicAllowListSettings},
		// polymorphicAllowList: `Box` in list, `as="button"` → resolved button → interactive.
		{Code: `<Box as="button" tabIndex={0} />`, Tsx: true, Settings: polymorphicAllowListSettings},
		// components remapping `Foo → Bar` (still custom) → skip.
		{Code: `<Foo tabIndex={0} />`, Tsx: true, Settings: componentsToCustomSettings},

		// ============================================================
		// Group 15: empty `jsx-a11y` settings
		// ============================================================
		// settings present but inner empty — falls to raw element name; `Foo`
		// not in dom set → skip.
		{Code: `<Foo tabIndex={0} />`, Tsx: true, Settings: emptyJsxA11ySettings},

		// ============================================================
		// Group 16: Real-world a11y patterns (no-report)
		// ============================================================
		// Tabs widget — every interactive role + tabIndex is correct usage.
		{
			Code: `function Tabs() { return <div role="tablist"><button role="tab" tabIndex={0}>One</button><button role="tab" tabIndex={-1}>Two</button></div>; }`,
			Tsx:  true,
		},
		// Tabpanel under recommended config (roles: ["tabpanel"]).
		{
			Code:    `<div role="tabpanel" tabIndex={0}>Content</div>`,
			Tsx:     true,
			Options: rolesExemptTabpanelOption,
		},
		// Combobox / Listbox.
		{
			Code: `<div role="combobox"><input role="searchbox" tabIndex={0} /><ul role="listbox"><li role="option" tabIndex={-1}>A</li></ul></div>`,
			Tsx:  true,
		},
		// Tree.
		{
			Code: `<ul role="tree"><li role="treeitem" tabIndex={0}>Root</li></ul>`,
			Tsx:  true,
		},
		// Menu.
		{
			Code: `<ul role="menu"><li role="menuitem" tabIndex={0}>Open</li></ul>`,
			Tsx:  true,
		},
		// Modal/Dialog with focus trap on a button (not a div).
		{
			Code: `function Modal({ open }) { return open ? <dialog open><button autoFocus tabIndex={0}>Close</button></dialog> : null; }`,
			Tsx:  true,
		},
		// useRef + manual focus — no tabIndex prop.
		{
			Code: `function Search() { const ref = useRef(null); useEffect(() => ref.current?.focus(), []); return <input ref={ref} />; }`,
			Tsx:  true,
		},
		// React.forwardRef on an interactive element.
		{
			Code: `const Inp = React.forwardRef((props, ref) => <input ref={ref} tabIndex={0} {...props} />);`,
			Tsx:  true,
		},
		// React.memo with tabIndex on interactive button.
		{
			Code: `const Item = React.memo(({ id }) => <button id={id} tabIndex={0}>{id}</button>);`,
			Tsx:  true,
		},
		// HOC pattern.
		{
			Code: `const Enhanced = withTracking(({ value }) => <input value={value} tabIndex={0} />);`,
			Tsx:  true,
		},
		// Generic forwardRef.
		{
			Code: `const TypedInput = React.forwardRef<HTMLInputElement, Props>((props, ref) => <input ref={ref} tabIndex={0} {...props} />);`,
			Tsx:  true,
		},

		// ============================================================
		// Group 17: Default-options shapes — JSON path coverage
		// ============================================================
		// Bare object (single-option CLI shape) — exercises GetOptionsMap's
		// `opts.(map[string]interface{})` arm.
		{Code: `<div tabIndex={-1} />`, Tsx: true, Options: map[string]interface{}{}},
		// Array-wrapped (rule_tester / multi-element shape) — exercises the
		// `[]interface{}` arm.
		{Code: `<div tabIndex={-1} />`, Tsx: true, Options: []interface{}{map[string]interface{}{}}},
		// Empty array — defaults to all-falsy (no escape hatches).
		{Code: `<div tabIndex={-1} />`, Tsx: true, Options: []interface{}{}},
		// Malformed option types are silently dropped → defaults apply.
		{Code: `<div tabIndex={-1} />`, Tsx: true, Options: []interface{}{map[string]interface{}{"tags": "not-an-array"}}},
		{Code: `<div tabIndex={-1} />`, Tsx: true, Options: []interface{}{map[string]interface{}{"roles": 123}}},
		{Code: `<div tabIndex={-1} />`, Tsx: true, Options: []interface{}{map[string]interface{}{"allowExpressionValues": "yes"}}},
		// Combined options on top of recommendedOptions shape.
		{Code: `<div tabIndex={0} />`, Tsx: true, Options: allOptionsCombo},
		// `tags: []` (explicit empty list) — JS truthy but `includes([], type)`
		// always false. Doesn't exempt anything.
		{Code: `<div tabIndex={-1} />`, Tsx: true, Options: []interface{}{map[string]interface{}{"tags": []interface{}{}}}},

		// ============================================================
		// Group 18: Conditional with both branches resulting in skip
		// ============================================================
		// Both branches are negative → step 2 picks one (test truthy → consequent),
		// negative → skip.
		{Code: `<div tabIndex={true ? -1 : -2} />`, Tsx: true},
		// Both branches resolve to non-numeric → step 2 → not in jvNumber/jvString-numeric.
		{Code: `<div tabIndex={true ? someVar : otherVar} />`, Tsx: true},
		// LogicalExpression with falsy-string value.
		{Code: `<div tabIndex={false || -1} />`, Tsx: true},

		// progressbar role — aria-query treats it as interactive (widget
		// descendant via `range`), so isInteractiveRole skips. Lock-in.
		{Code: `<div role="progressbar" tabIndex={0} />`, Tsx: true},

		// TSNonNullExpression on tabIndex — upstream stringifies to "0!" /
		// "5!" → Number(...) = NaN → step-1 undefined → no-non skips.
		// Lock-in for Cluster A.
		{Code: `<div tabIndex={0!} />`, Tsx: true},
		{Code: `<div tabIndex={(0)!} />`, Tsx: true},
		{Code: `<div tabIndex={(5)!} />`, Tsx: true},

		// BigInt with negative value — Number(-1n) = -1, -1 >= 0 false → skip.
		{Code: `<div tabIndex={-1n} />`, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// Empty JsxExpression `tabIndex={}` — JSXEmptyExpression, upstream
		// TYPES no entry → null. `null >= 0` true → REPORT. Lock-in for
		// the Cluster H fix.
		{Code: `<div tabIndex={} />`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveTabindex", Message: errorMessage}}},

		// BigInt — upstream Number(BigInt) coerces. 0n → 0 → `0>=0` REPORT;
		// 1n → 1 → REPORT; 2n → 2 → REPORT. Lock-in for Cluster B.
		{Code: `<div tabIndex={0n} />`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveTabindex", Message: errorMessage}}},
		{Code: `<div tabIndex={1n} />`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveTabindex", Message: errorMessage}}},
		{Code: `<div tabIndex={2n} />`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveTabindex", Message: errorMessage}}},
		{Code: `<div tabIndex={5n} />`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveTabindex", Message: errorMessage}}},
		{Code: `<div tabIndex={true ? 1n : 0n} />`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveTabindex", Message: errorMessage}}},

		// ============================================================
		// Group 0: Opaque expression types — upstream returns null which
		// passes the `typeof === 'undefined'` guard, then `null >= 0` is
		// true (ToNumber-coerces to 0). Aligned via GetTabIndexEx's
		// nullLike arm. Locks against accidental regression to the lossy
		// pre-Ex behavior that silently skipped these.
		// ============================================================
		{Code: `<div tabIndex={-1 satisfies number} />`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveTabindex", Message: errorMessage}}},
		{Code: `<div tabIndex={5 satisfies number} />`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveTabindex", Message: errorMessage}}},
		{Code: `async function f() { return <div tabIndex={await p} />; }`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveTabindex", Message: errorMessage}}},
		{Code: `function* g() { yield <div tabIndex={yield 0} />; }`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveTabindex", Message: errorMessage}}},

		// ============================================================
		// Group 1: Position assertions
		// ============================================================
		// `<div tabIndex="0" />` — JsxAttribute spans columns 6..18.
		{
			Code: `<div tabIndex="0" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noNoninteractiveTabindex",
				Message:   errorMessage,
				Line:      1, Column: 6, EndLine: 1, EndColumn: 18,
			}},
		},
		// Numeric form — JsxAttribute spans `tabIndex={0}` (12 chars), columns 6..18.
		{
			Code: `<div tabIndex={0} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noNoninteractiveTabindex",
				Message:   errorMessage,
				Line:      1, Column: 6, EndLine: 1, EndColumn: 18,
			}},
		},
		// Multi-line attribute — position spans the entire attribute.
		{
			Code: "<div\n  tabIndex={\n    0\n  } />",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noNoninteractiveTabindex",
				Message:   errorMessage,
				Line:      2, Column: 3, EndLine: 4, EndColumn: 4,
			}},
		},
		// Paired (non-self-closing) element.
		{
			Code: `<div tabIndex="0">child</div>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noNoninteractiveTabindex",
				Message:   errorMessage,
				Line:      1, Column: 6, EndLine: 1, EndColumn: 18,
			}},
		},
		// Position with role attribute before tabIndex.
		{
			Code: `<div role="article" tabIndex="0" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noNoninteractiveTabindex",
				Message:   errorMessage,
				Line:      1, Column: 21, EndLine: 1, EndColumn: 33,
			}},
		},

		// ============================================================
		// Group 2: Listener boundary — nested elements report independently
		// ============================================================
		// Outer + inner each emit a diagnostic.
		{
			Code: `<article tabIndex={0}><span tabIndex={0} /></article>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noNoninteractiveTabindex", Message: errorMessage},
				{MessageId: "noNoninteractiveTabindex", Message: errorMessage},
			},
		},
		// Three levels deep — outer/mid/inner all report.
		{
			Code: `<article tabIndex={0}><section tabIndex={0}><span tabIndex={0} /></section></article>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noNoninteractiveTabindex", Message: errorMessage},
				{MessageId: "noNoninteractiveTabindex", Message: errorMessage},
				{MessageId: "noNoninteractiveTabindex", Message: errorMessage},
			},
		},
		// Mixed: outer interactive, inner non-interactive — only inner reports.
		{
			Code: `<button tabIndex={0}><span tabIndex={0} /></button>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noNoninteractiveTabindex", Message: errorMessage},
			},
		},
		// Two siblings — both report.
		{
			Code: `<><div tabIndex={0} /><span tabIndex={0} /></>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				expectedError, expectedError,
			},
		},

		// ============================================================
		// Group 3: Non-interactive element survey (without role override)
		// ============================================================
		{Code: `<div tabIndex={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<span tabIndex={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<section tabIndex={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<article tabIndex={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<header tabIndex={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<footer tabIndex={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<main tabIndex={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<aside tabIndex={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<nav tabIndex={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<p tabIndex={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<h1 tabIndex={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<h2 tabIndex={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<h3 tabIndex={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<h4 tabIndex={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<h5 tabIndex={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<h6 tabIndex={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<ul tabIndex={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<ol tabIndex={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<li tabIndex={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<table tabIndex={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<dl tabIndex={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<dt tabIndex={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<dd tabIndex={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<dialog tabIndex={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<details tabIndex={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<fieldset tabIndex={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// `<a />` (no href) is non-interactive (matches `{name: "a"}` schema).
		{Code: `<a tabIndex={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<area tabIndex={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ============================================================
		// Group 4: Non-interactive role values
		// ============================================================
		// First-token rule: even if a later token is interactive, the first
		// (article) wins — non-interactive → reports.
		{Code: `<div role="article button" tabIndex={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Various non-interactive roles.
		{Code: `<div role="article" tabIndex={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div role="document" tabIndex={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div role="img" tabIndex={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div role="heading" tabIndex={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div role="region" tabIndex={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div role="tabpanel" tabIndex={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Note: `<div role="progressbar" tabIndex={0} />` is VALID per upstream.
		// progressbar's superClass chain in aria-query contains `widget` (via
		// `range`), so isInteractiveRole returns true and the rule skips.
		// Moved to valid section above.
		// More non-interactive roles.
		{Code: `<div role="alert" tabIndex={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div role="banner" tabIndex={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div role="complementary" tabIndex={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div role="contentinfo" tabIndex={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div role="form" tabIndex={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div role="navigation" tabIndex={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div role="search" tabIndex={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div role="status" tabIndex={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div role="presentation" tabIndex={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div role="none" tabIndex={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Invalid / unknown role names — first valid arises in iteration; with
		// no valid roles, treated as "no interactive role".
		{Code: `<div role="madeUpRole" tabIndex={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div role="" tabIndex={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div role="   " tabIndex={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ============================================================
		// Group 5: Non-literal role + allowExpressionValues=false
		// ============================================================
		// Without the escape hatch, IsInteractiveRole can't read the role →
		// returns false → reports.
		{Code: `<div role={SOME_ROLE} tabIndex={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Same with explicit `false`.
		{Code: `<div role={SOME_ROLE} tabIndex={0} />`, Tsx: true, Options: allowExpressionValuesFalseOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// `role={undefined}` is special-cased to NOT count as non-literal,
		// so allowExpressionValues=true does NOT skip; rule still reaches
		// IsInteractiveRole which can't extract a string → reports.
		{Code: `<div role={undefined} tabIndex={0} />`, Tsx: true, Options: allowExpressionValuesTrueOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// `role={null}` is JSXExpressionContainer with NullKeyword — upstream
		// `JSXExpressionContainer` arm: NOT Identifier 'undefined', NOT JSXText
		// → IsNonLiteralProperty returns true → exempt under
		// allowExpressionValues=true. Without the option, falls through to
		// IsInteractiveRole which returns false → reports.
		{Code: `<div role={null} tabIndex={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Role with non-string literal — upstream LITERAL_TYPES.Literal
		// special-case (null → "null"); other literal types yield non-string
		// values that don't match any role.
		{Code: `<div role={0} tabIndex={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div role={true} tabIndex={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div role={false} tabIndex={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ============================================================
		// Group 6: tabIndex value variants that resolve to number ≥ 0
		// ============================================================
		{Code: `<div tabIndex={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div tabIndex={1} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div tabIndex={5} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div tabIndex={2147483647} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}}, // INT32_MAX
		// Hex / Octal / Binary integer literals — upstream `Number()` of the
		// pre-normalized text returns the integer value.
		{Code: `<div tabIndex={0x10} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div tabIndex={0o10} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div tabIndex={0b10} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Hex / octal / binary STRING forms — upstream `Number("0x10") = 16`
		// (etc.) → integer ≥ 0 → reports. Locks in the parseLiteralTabIndexString
		// `0x` / `0o` / `0b` prefix dispatch.
		{Code: `<div tabIndex="0x10" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div tabIndex="0X10" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}}, // uppercase X
		{Code: `<div tabIndex="0o10" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div tabIndex="0O10" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}}, // uppercase O
		{Code: `<div tabIndex="0b10" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div tabIndex="0B10" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}}, // uppercase B
		{Code: `<div tabIndex="0x0" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}}, // hex 0
		{Code: `<div tabIndex="0xff" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}}, // mixed-case hex digits
		{Code: `<div tabIndex="0xFF" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Step-2 fallback hex via ConditionalExpression — staticEval picks
		// the truthy branch's jvString and runs the same hex / oct / bin
		// dispatch in staticEvalToTabIndex.
		{Code: `<div tabIndex={true ? "0x10" : "-1"} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// `Infinity` identifier — staticEval returns jvNumber +Inf. Step 2
		// no longer rejects ±Infinity (upstream `Infinity >= 0` is true →
		// reports). Locks in that staticEvalToTabIndex's IsNaN-only guard
		// matches upstream.
		{Code: `<div tabIndex={Infinity} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// ArrayLiteralExpression — Array.prototype.join + ToNumber yields a
		// non-negative finite number. Locks in arrayToTabIndex's element
		// stringify + comma-join + ToNumber pipeline.
		{Code: `<div tabIndex={[]} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},     // "" → 0
		{Code: `<div tabIndex={[5]} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},    // "5" → 5
		{Code: `<div tabIndex={[0]} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},    // "0" → 0
		{Code: `<div tabIndex={[null]} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}}, // "" → 0
		{Code: `<div tabIndex={[undefined]} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div tabIndex={[Infinity]} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}}, // "Infinity" → +Inf
		// Unary `+`/`-` on string operand → ToNumber. Locks in
		// unaryStringToTabIndex.
		{Code: `<div tabIndex={+"5"} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},    // +5
		{Code: `<div tabIndex={-"-5"} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},   // -(-5) = 5
		{Code: `<div tabIndex={+"0x10"} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}}, // +16
		// Numeric separator (ES2021) — TS NumericLiteral text strips them.
		{Code: `<div tabIndex={1_000} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Scientific notation that lands on integer.
		{Code: `<div tabIndex={1e2} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Negative zero — `-0 >= 0` is true in JS (`-0 == 0`).
		{Code: `<div tabIndex={-0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// String forms.
		{Code: `<div tabIndex="5" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div tabIndex="0" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div tabIndex="100" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// String "0" with surrounding whitespace — Number coerces to 0.
		// JSX direct string-attribute values are RAW (`\t` is the two
		// literal characters `\` `t`, not a tab); the JsxExpression form
		// `{"..."}` runs through ECMAScript escape parsing instead. Both
		// shapes Number-coerce to 0 after our TrimSpace.
		{Code: `<div tabIndex=" 0 " />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: "<div tabIndex={\" 0 \"} />", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// String "+0" / "+1" — leading + parses successfully.
		{Code: `<div tabIndex="+0" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div tabIndex="+1" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// String "-0" — JS Number("-0") = -0, `-0 >= 0` is true.
		{Code: `<div tabIndex="-0" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Numeric expression resolved by staticEval (the getPropValue fallback).
		{Code: `<div tabIndex={1+1} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div tabIndex={2-1} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div tabIndex={2*0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div tabIndex={5%3} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// String concatenation evaluates to "00" → ToNumber → 0 → ≥0 → reports.
		{Code: `<div tabIndex={"0" + "0"} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Unary plus / minus on numeric literals.
		{Code: `<div tabIndex={+0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div tabIndex={+5} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div tabIndex={-(-1)} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}}, // 1
		// NoSubstitutionTemplateLiteral → string "0" → integer 0.
		{Code: "<div tabIndex={`0`} />", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: "<div tabIndex={`5`} />", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Parenthesized number literal.
		{Code: `<div tabIndex={(0)} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div tabIndex={((0))} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// TS wrapper around a number literal — staticEval skips the wrapper.
		{Code: `<div tabIndex={0 as number} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div tabIndex={(0) as number} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div tabIndex={0 as any} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Note: `<div tabIndex={(0)!} />` is VALID per upstream
		// (TSNonNullExpression stringifies to "0!" → NaN → step-1 undefined
		// → no-non skips). Moved to valid section.
		// ConditionalExpression with non-negative arms — staticEval picks one.
		{Code: `<div tabIndex={cond ? 0 : -1} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div tabIndex={true ? 0 : 1} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div tabIndex={true ? "0" : "-1"} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// LogicalExpression resolving to non-negative.
		{Code: `<div tabIndex={true && 0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div tabIndex={false || 1} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// NullishCoalescing.
		{Code: `<div tabIndex={null ?? 0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Boolean fallback in step-2 (very rare in practice — `cond ? true : false`).
		{Code: `<div tabIndex={true ? true : false} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ============================================================
		// Group 7: tags / roles option does NOT exempt
		// ============================================================
		// `tags: ["span"]` doesn't exempt `<div>`.
		{Code: `<div tabIndex={0} />`, Tsx: true, Options: []interface{}{
			map[string]interface{}{"tags": []interface{}{"span"}},
		}, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// `roles: ["tabpanel"]` doesn't exempt `role="article"`.
		{Code: `<div role="article" tabIndex={0} />`, Tsx: true, Options: rolesExemptTabpanelOption, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Empty `roles: []` — `[]` is JS truthy but `includes([], anything)`
		// false → no exemption.
		{Code: `<div role="article" tabIndex={0} />`, Tsx: true, Options: []interface{}{
			map[string]interface{}{"roles": []interface{}{}},
		}, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// allOptionsCombo allows `roles: [tabpanel, alertdialog]` but div is
		// itself in `tags: [div, section]` → exempt regardless of role.
		// Use `<article>` (not in tags) with role="article" (not in roles)
		// to verify the rule still reports when neither escape hatch matches.
		{Code: `<article role="article" tabIndex={0} />`, Tsx: true, Options: allOptionsCombo, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ============================================================
		// Group 8: Multiple tabIndex props (degenerate but legal)
		// ============================================================
		{
			Code:   `<div tabIndex={0} tabIndex={5} />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// FindAttributeByName returns the FIRST match — the second (5) is
		// ignored, but the first (0) trips the rule.

		// ============================================================
		// Group 9: polymorphicPropName remap to non-interactive
		// ============================================================
		{
			Code:     `<Box as="div" tabIndex={0} />`,
			Tsx:      true,
			Settings: polymorphicSettings,
			Errors:   []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:     `<Box as="article" tabIndex={0} />`,
			Tsx:      true,
			Settings: polymorphicSettings,
			Errors:   []rule_tester.InvalidTestCaseError{expectedError},
		},
		// polymorphicAllowList: `Box` in list, swap to non-interactive → reports.
		{
			Code:     `<Box as="article" tabIndex={0} />`,
			Tsx:      true,
			Settings: polymorphicAllowListSettings,
			Errors:   []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Group 10: Components map → non-interactive
		// ============================================================
		{
			Code: `<MyArticle tabIndex={0} />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"components": map[string]interface{}{"MyArticle": "article"},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code: `<MyDiv tabIndex={0} />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"components": map[string]interface{}{"MyDiv": "div"},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Group 11: Real-world component patterns (offending tabIndex)
		// ============================================================
		{
			Code:   `function Outer() { return <div tabIndex={0}>focusable but not interactive</div>; }`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   `const items = arr.map(item => <li key={item.id} tabIndex={0}>{item.name}</li>);`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   `function Foo({cond}) { return cond ? <article tabIndex={0} /> : null; }`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// React.forwardRef on a non-interactive div with tabIndex.
		{
			Code:   `const Pane = React.forwardRef((props, ref) => <div ref={ref} tabIndex={0} {...props} />);`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// Multiple components in one file.
		{
			Code: "function A() { return <div tabIndex={0} />; }\nfunction B() { return <article tabIndex={0} />; }",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				expectedError, expectedError,
			},
		},
		// HOC pattern with non-interactive element + tabIndex.
		{
			Code:   `const Enhanced = withTracking(({ value }) => <div data-value={value} tabIndex={0} />);`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// React.memo with non-interactive element.
		{
			Code:   `const Pane = React.memo(({ id }) => <section id={id} tabIndex={0} />);`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Group 12: Real-world a11y pattern misuse (typical mistakes)
		// ============================================================
		// "Custom button" via div + onClick + tabIndex but no role — reports
		// as expected (developer should use <button> or add role="button").
		{
			Code:   `function FakeButton({ onClick, children }) { return <div onClick={onClick} tabIndex={0}>{children}</div>; }`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// "Custom dropdown" missing role.
		{
			Code:   `function Dropdown({ items }) { return <ul tabIndex={0}>{items.map(x => <li key={x}>{x}</li>)}</ul>; }`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// Click-card pattern.
		{
			Code:   `function ClickCard() { return <article tabIndex={0} onClick={handler}>Title</article>; }`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Group 13: Generator / async / IIFE bodies
		// ============================================================
		{
			Code: `function* render() { yield <div tabIndex={0} />; yield <article tabIndex={0} />; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				expectedError, expectedError,
			},
		},
		{
			Code:   `async function render() { return <div tabIndex={0} />; }`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   `const x = (() => <article tabIndex={0} />)();`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Group 14: Class component render
		// ============================================================
		{
			Code:   `class Form extends React.Component { render() { return <div tabIndex={0}>ready</div>; } }`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// Class field arrow render.
		{
			Code:   `class Pane extends React.Component { render = () => <section tabIndex={0} />; }`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Group 15: Fragment + conditional rendering
		// ============================================================
		{
			Code:   `const x = <>{cond && <div tabIndex={0} />}</>`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   `const x = <React.Fragment><div tabIndex={0} /></React.Fragment>`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// Switch-case rendering.
		{
			Code: `function Foo({type}) { switch(type) { case 'a': return <div tabIndex={0} />; default: return null; } }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Group 16: Comments around / inside the prop don't suppress
		// ============================================================
		{
			Code: `<div /* a */ tabIndex={0} /* b */ />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noNoninteractiveTabindex",
				Message:   errorMessage,
				Line:      1, Column: 14, EndLine: 1, EndColumn: 26,
			}},
		},
		{
			Code:   `<div tabIndex={/* truthy */ 0} />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   "<div\n  tabIndex={\n    0 // explicit\n  }\n/>",
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Group 17: Spread literal — role / tabIndex flowing through
		// ============================================================
		// FindAttributeByName walks literal spread; the rule sees `role: "article"`
		// and `tabIndex: 0`. role="article" is non-interactive → reports.
		{
			Code:   `<div {...{role: "article", tabIndex: 0}} />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// Spread `{tabIndex: 0}` only — no role override → div is non-interactive → reports.
		{
			Code:   `<div {...{tabIndex: 0}} />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Group 18: TypeScript generic JSX with non-interactive resolved tag
		// ============================================================
		// `<List<string>>` — list is custom (no dom set match) → skip.
		// But the inner `<div>` is non-interactive → reports.
		{
			Code:   `function makeCell() { return <Cell><div tabIndex={0}>x</div></Cell>; }`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Group 19: Multiple errors in a single file across patterns
		// ============================================================
		{
			Code: `function App() {
				return (
					<>
						<div tabIndex={0}>A</div>
						<button tabIndex={0}>B</button>
						<article tabIndex={0}>C</article>
						<input tabIndex={0} />
						<section tabIndex={0}>D</section>
					</>
				);
			}`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				expectedError, // div
				expectedError, // article
				expectedError, // section (button and input are interactive)
			},
		},
	})
}

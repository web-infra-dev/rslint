// TestJsxCurlySpacingExtras locks in branches and edge shapes the upstream
// @stylistic suite doesn't exercise. Each case carries an inline comment
// pointing at the branch / Dimension 4 row / tsgo AST quirk it covers, so a
// future refactor can't silently regress it. The implementation is shared with
// react/jsx-curly-spacing via BuildRule, so these augmentation cases apply
// equally to both variants. Shared const/option-alias fixtures live in the
// sibling jsx_curly_spacing_upstream_test.go (same package).
//
// Groups: (1) Dimension 4 — the tsgo↔ESTree shape surface the rule treats as
// opaque (paren / TS-wrapper / optional-chain / Kind*Literal / element-vs-
// fragment nesting); (2) branch lock-ins for every spacing / objectLiterals /
// allowMultiline arm; (3) real-user shapes from the upstream issue tracker;
// (4) diagnostic contract — exact message text + Line/Column/EndLine/EndColumn;
// (5) the options JSON path — array-wrapped (CLI / rule-tester) shapes.
package jsx_curly_spacing

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestJsxCurlySpacingExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxCurlySpacingRule, []rule_tester.ValidTestCase{
		// ===== Dimension 4 lock-in: tsgo-specific shapes & real-world =====
		// These tests are not in the upstream suite. They exercise AST shapes
		// and content patterns that are easy to break in a port (string
		// content, regex literals, TS-only operators, parenthesized wrappers,
		// nested containers). The contract is: rules of brace-spacing depend
		// ONLY on the source text immediately surrounding `{` and `}` — the
		// container's inner expression is opaque.

		// ---- TS expression wrappers — should not trigger "object literal" ----
		{Code: `<App foo={(bar)} />`, Tsx: true},             // ParenthesizedExpression
		{Code: `<App foo={((bar))} />`, Tsx: true},           // multi-level paren
		{Code: `<App foo={bar!} />`, Tsx: true},              // non-null assertion
		{Code: `<App foo={bar as any} />`, Tsx: true},        // type assertion
		{Code: `<App foo={bar satisfies Foo} />`, Tsx: true}, // satisfies
		{Code: `<App foo={bar?.baz} />`, Tsx: true},          // optional chain
		{Code: `<App foo={bar?.()} />`, Tsx: true},           // optional call
		{Code: `<App foo={(bar)} />`, Tsx: true, Options: []interface{}{opts{"when": "never"}}},
		{Code: `<App foo={ (bar) } />`, Tsx: true, Options: []interface{}{opts{"when": "always"}}},

		// ---- Non-`{` second tokens — must use `when`, NOT objectLiteralSpaces ----
		{Code: `<App foo={[1, 2]} />`, Tsx: true, Options: []interface{}{opts{"when": "never", "spacing": spc{"objectLiterals": "always"}}}},
		{Code: `<App foo={<Bar />} />`, Tsx: true, Options: []interface{}{opts{"when": "never", "spacing": spc{"objectLiterals": "always"}}}},
		{Code: `<App foo={() => {}} />`, Tsx: true, Options: []interface{}{opts{"when": "never", "spacing": spc{"objectLiterals": "always"}}}},
		{Code: `<App foo={typeof bar} />`, Tsx: true, Options: []interface{}{opts{"when": "never", "spacing": spc{"objectLiterals": "always"}}}},
		{Code: `<App foo={!bar} />`, Tsx: true, Options: []interface{}{opts{"when": "never", "spacing": spc{"objectLiterals": "always"}}}},
		{Code: `<App foo={-1} />`, Tsx: true, Options: []interface{}{opts{"when": "never", "spacing": spc{"objectLiterals": "always"}}}},
		{Code: `<App foo={1 + 2} />`, Tsx: true, Options: []interface{}{opts{"when": "never", "spacing": spc{"objectLiterals": "always"}}}},

		// ---- Strings / templates / regex containing trivia-like text ----
		// These exercise the scanner-based body scan: the inner content
		// includes characters that a naive byte-scanner could mistake for
		// comments or unbalanced braces.
		{Code: `<App foo={"// not a comment"} />`, Tsx: true},
		{Code: `<App foo={"/* not a comment */"} />`, Tsx: true},
		{Code: `<App foo={"{ not an object }"} />`, Tsx: true},
		{Code: `<App foo={'single \'quote\''} />`, Tsx: true},
		{Code: "<App foo={`tpl ${a} ${b}`} />", Tsx: true},
		{Code: "<App foo={`with } brace inside`} />", Tsx: true},
		{Code: `<App foo={/regex/g} />`, Tsx: true},
		{Code: `<App foo={ /regex/g } />`, Tsx: true, Options: []interface{}{opts{"when": "always"}}},

		// ---- Nested containers ----
		// `children` defaults to false: only inner attribute container is checked.
		{Code: `<App><Bar foo={baz} /></App>`, Tsx: true},
		// fragment + element + attribute container — all clean.
		{Code: `<><Bar foo={baz} /></>`, Tsx: true},
		// nested fragment.
		{Code: `<><></></>`, Tsx: true},
		// attribute container with JSX expression inside.
		{Code: `<App foo={<Bar baz={qux} />} />`, Tsx: true},
		// children container with nested element with attribute container.
		{Code: `<App>{<Bar foo={baz} />}</App>`, Tsx: true, Options: []interface{}{opts{"children": opts{"when": "never"}}}},

		// ---- Spread variants ----
		// Children spread (`<App>{...arr}</App>`) is JSXSpreadChild in ESTree
		// and NOT covered by upstream's listener — verify our skip keeps the
		// rule silent across all configs, including ones that would otherwise
		// fire on the surrounding spaces.
		{Code: `<Foo>{...arr}</Foo>`, Tsx: true},
		{Code: `<Foo>{...arr}</Foo>`, Tsx: true, Options: []interface{}{opts{"children": opts{"when": "never"}}}},
		{Code: `<Foo>{ ...arr }</Foo>`, Tsx: true, Options: []interface{}{opts{"children": opts{"when": "never"}}}},
		{Code: `<Foo>{ ...arr }</Foo>`, Tsx: true, Options: []interface{}{opts{"children": opts{"when": "always"}}}},
		{Code: `<Foo>{...arr}</Foo>`, Tsx: true, Options: []interface{}{opts{"children": opts{"when": "always"}}}},
		{Code: `<App {...obj1} {...obj2} />`, Tsx: true},
		// Spread of a parenthesized object — second token is `.`, NOT `{`.
		{Code: `<App {...({a: 1, ...rest})} />`, Tsx: true},

		// ---- Real-world idioms ----
		{Code: `<div className={isActive ? 'active' : 'inactive'} />`, Tsx: true},
		{Code: `<button onClick={() => handleClick()}>Click</button>`, Tsx: true},
		{Code: `<App data-x={JSON.stringify({a: 1})} />`, Tsx: true},
		{Code: `<List items={items.map((item) => <Item key={item.id} />)} />`, Tsx: true},
		// Generic call inside attribute container.
		{Code: `<App foo={callMe<T>()} />`, Tsx: true},
		// Object literal as attribute (default never spacing for objectLiterals via lastPass).
		{Code: `<div style={{color: 'red'}} />`, Tsx: true},
		// Line comment immediately before `}` on the next line — allowed
		// under the default allowMultiline=true. Verifies the scanner-based
		// body scan correctly recognizes the comment as the previous
		// "thing" (so the multiline detection / reporting path agrees with
		// upstream's getTokenBefore({includeComments:true})).
		{Code: "<App>{foo // c\n}</App>", Tsx: true, Options: []interface{}{opts{"children": opts{"when": "never"}}}},
		// Same shape with a BLOCK comment immediately before `}` on the
		// next line — also allowed by default.
		{Code: "<App>{foo /* c */\n}</App>", Tsx: true, Options: []interface{}{opts{"children": opts{"when": "never"}}}},
		// Empty container `{}` and whitespace-only `{ }` — `never` accepts
		// both `{}` (no inner space) and same with `always` accepts `{ }`
		// (a single space). Locks in the empty-body branch where
		// `secondPos == innerHigh` and `isObjectLiteral` is correctly false.
		{Code: `<App>{}</App>`, Tsx: true, Options: []interface{}{opts{"children": opts{"when": "never"}}}},
		// Triple-nested attribute containers — listener triggers
		// independently for each JsxExpression, regardless of depth.
		{Code: `<Outer foo={<Mid bar={<Inner baz={qux} />} />} />`, Tsx: true},
		// Mixed normal + spread + normal attributes (ESLint passes them
		// through three separate listeners; we use one callback that runs
		// per node in source order).
		{Code: `<App foo={bar} {...spread} qux={quux} />`, Tsx: true},
		// CRLF line endings — newline detection must accept `\r\n`.
		{Code: "<App foo={\r\n  bar\r\n} />", Tsx: true},

		// ===== Robustness: complex inner contents under multi-line `always` =====
		// All of the following are multi-line attribute containers under
		// `always` mode with the default `allowMultiline: true`. The body
		// contains characters that a token-level scan would mis-classify
		// (template `}`, regex `}`, string `}`, nested template, nested
		// object literal `}`, JSX in body). The rule must remain SILENT —
		// brace spacing only depends on trivia immediately around the
		// outer `{` / `}`.

		// regex literal containing `}` in a character class.
		{Code: "<X foo={\n  /[}]+/g\n} />", Tsx: true, Options: []interface{}{"always"}},
		// regex literal containing escaped slash and `}`.
		{Code: "<X foo={\n  /a\\/b}c/g\n} />", Tsx: true, Options: []interface{}{"always"}},
		// String literal containing `}`.
		{Code: "<X foo={\n  \"hello}world\"\n} />", Tsx: true, Options: []interface{}{"always"}},
		// String literal containing `{` and `}`.
		{Code: "<X foo={\n  \"{ not an object }\"\n} />", Tsx: true, Options: []interface{}{"always"}},
		// Nested template literal: outer template's substitution itself contains a template.
		{Code: "<X foo={\n  `outer ${`inner ${x}`}`\n} />", Tsx: true, Options: []interface{}{"always"}},
		// Multiple `${ }` substitutions in one template (the regression shape).
		{Code: "<X foo={\n  `${a}${b}${c}`\n} />", Tsx: true, Options: []interface{}{"always"}},
		// Arrow function body with block statement (raw `{` `}` inside).
		{Code: "<X foo={\n  () => { return 1; }\n} />", Tsx: true, Options: []interface{}{"always"}},
		// Multi-line object literal as body (not "always" promotion of objectLiterals — body has surrounding newlines).
		{Code: "<X foo={\n  {\n    a: 1,\n    b: 2\n  }\n} />", Tsx: true, Options: []interface{}{"always", opts{"spacing": spc{"objectLiterals": "never"}}}},
		// Conditional with two object literals.
		{Code: "<X foo={\n  cond ? { a: 1 } : { b: 2 }\n} />", Tsx: true, Options: []interface{}{"always"}},
		// Fragment in body.
		{Code: "<X foo={\n  <>{inner}</>\n} />", Tsx: true, Options: []interface{}{"always"}},
		// arr.map returning JSX with nested attribute container — inner
		// `{ x.id }` carries spaces so the whole tree is `always`-clean.
		{Code: "<X foo={\n  arr.map((x) => <Item key={ x.id } />)\n} />", Tsx: true, Options: []interface{}{"always"}},
		// Tagged template with substitutions.
		{Code: "<X foo={\n  tag`hello ${name}`\n} />", Tsx: true, Options: []interface{}{"always"}},
		// Class expression body.
		{Code: "<X foo={\n  class { method() { return 1; } }\n} />", Tsx: true, Options: []interface{}{"always"}},

		// ===== Same-range edge cases (empty body with whitespace) =====
		// `<App>{ }</App>` body is whitespace-only — both noSpaceAfter and
		// noSpaceBefore fixes target the same range; rule_tester must apply
		// them without conflict.
		// Skipped: rule_tester rejects overlapping autofix ranges, even
		// when both replace the same span with the same content. The rule
		// itself does the right thing semantically (matches upstream's
		// double-report on `{ }`); we just can't assert the autofix
		// output through this harness when both fixes point at the same
		// span. Documented and asserted via the no-fix invalid case below.

		// A single tab between `{` and `}` should report just like a single space.
		{Code: "<X>{}</X>", Tsx: true, Options: []interface{}{opts{"children": opts{"when": "never"}}}},

		// ===== BOM + Unicode lock-in =====
		// rslint reports columns as 1-based UTF-16 character offsets (via
		// scanner.GetECMALineAndUTF16CharacterOfPosition), matching ESLint.
		// These tests lock that behaviour against three classes of source:
		//   1. BOM byte at file start — must NOT be counted as a column,
		//   2. BMP non-ASCII identifiers/strings/comments (e.g. `中`) —
		//      must each count as exactly 1 UTF-16 character,
		//   3. SMP characters represented by UTF-16 surrogate pairs (e.g.
		//      `🚀`) — must each count as 2 UTF-16 characters.

		// BOM at file start: column for `{` and `}` should match the
		// non-BOM equivalent.
		{Code: "\uFEFF<App foo={bar} />", Tsx: true},
		// BMP non-ASCII identifier inside braces — no surrounding spaces, valid under default never.
		{Code: `<App foo={中文} />`, Tsx: true},
		// String literal containing non-ASCII characters.
		{Code: `<App foo={"中文 with spaces"} />`, Tsx: true},
		// Block comment with non-ASCII text — both sides flush, valid.
		{Code: `<App foo={/* 中文注释 */ bar} />`, Tsx: true},
		// Single emoji (SMP, UTF-16 surrogate pair) inside braces.
		{Code: `<App foo={"🚀"} />`, Tsx: true},
		// `always` mode with surrounding spaces and non-ASCII content — valid.
		{Code: `<App foo={ 中文 } />`, Tsx: true, Options: []interface{}{"always"}},
		{Code: `<App foo={ "🚀" } />`, Tsx: true, Options: []interface{}{"always"}},

		// ===== Unicode WhiteSpace + LineTerminator (ECMAScript §12.2/§12.3) =====
		// NBSP (U+00A0) counts as ECMAScript WhiteSpace → satisfies `always`,
		// triggers `never` extra. Parity with ESLint verified via local probe.
		{Code: "<App foo={\u00A0bar\u00A0} />", Tsx: true, Options: []interface{}{"always"}},
		// LS (U+2028) / PS (U+2029) count as LineTerminator → cross-line
		// short-circuit on that side. Both modes valid when the OTHER side
		// hugs the brace.
		{Code: "<App foo={bar\u2028} />", Tsx: true, Options: []interface{}{"never"}},
		{Code: "<App foo={\u2028bar} />", Tsx: true, Options: []interface{}{"never"}},
		{Code: "<App foo={bar\u2029} />", Tsx: true, Options: []interface{}{"never"}},
		// Template literal with `${ … }` substitutions — earlier impl using
		// the bare tsgo Scanner without parser context mis-tokenized the
		// closing `}` of a substitution as a real `}` token, corrupting
		// `penultimateEnd` for any enclosing multi-line attribute container.
		// Locks in: a multi-line `always` attribute whose body contains a
		// template with substitutions must not falsely report
		// `spaceNeededBefore` on the closing brace.
		{
			Code:    "<App\n  foo={\n    a + `/${b}${c}`\n  }\n/>",
			Tsx:     true,
			Options: []interface{}{"always"},
		},
		{
			Code:    "<App\n  foo={\n    `${x}-${y}-${z}`\n  }\n/>",
			Tsx:     true,
			Options: []interface{}{"always"},
		},
		// Nested attribute container whose body is a JSX subtree that
		// itself contains template-literal attributes — the same shape
		// that triggered the regression in rsbuild website/theme/index.tsx.
		// Inner `href={ ... }` carries surrounding spaces so it complies
		// with `always` on its own; the test asserts the OUTER multi-line
		// `beforeNav={ ... }` doesn't falsely report due to template `}`s.
		{
			Code:    "<Outer\n  beforeNav={\n    cond ? (\n      <Inner href={ `/${pre}${post}` } />\n    ) : null\n  }\n/>",
			Tsx:     true,
			Options: []interface{}{"always"},
		},

		// ---- JS-truthy semantics on schema-invalid `attributes` / `children` ----
		// ESLint's JSON-schema validator rejects these before the rule runs;
		// rslint does not validate schemas, so the rule must reproduce JS's
		// `value ? cfg : null` semantics (0/null/"" → disabled).
		{Code: `<App foo={ bar } />;`, Tsx: true, Options: []interface{}{opts{"attributes": nil}}},
		{Code: `<App foo={ bar } />;`, Tsx: true, Options: []interface{}{opts{"attributes": 0}}},
		{Code: `<App foo={ bar } />;`, Tsx: true, Options: []interface{}{opts{"attributes": ""}}},
		{Code: `<App>{ bar }</App>;`, Tsx: true, Options: []interface{}{opts{"children": nil}}},
		{Code: `<App>{ bar }</App>;`, Tsx: true, Options: []interface{}{opts{"children": 0}}},

		// ===== @stylistic option-normalization scope (stylisticScope=true) =====
		// Locks in the per-side `spacing` vs top-level `spacing.objectLiterals`
		// interaction that distinguishes @stylistic from react. @stylistic's
		// `spacing = perSide.spacing ?? defaultSpacing` has three branches:
		//   (a) per-side empty `spacing:{}` shadows the top-level default →
		//       objectLiteralSpaces falls back to `when` (react would keep the
		//       top-level value — this is the one cross-fork delta).
		//   (b) per-side WITHOUT a spacing key inherits the top-level default.
		//   (c) per-side `spacing.objectLiterals` explicitly overrides.
		// (a) empty per-side spacing → objectLiterals == when:'never' → flush object valid
		{
			Code:    `<App foo={{a: 1}} />`,
			Tsx:     true,
			Options: []interface{}{opts{"spacing": spc{"objectLiterals": "always"}, "attributes": opts{"when": "never", "spacing": spc{}}}},
		},
		// (b) no per-side spacing key → inherits top-level objectLiterals:'always' → spaced object valid
		{
			Code:    `<App foo={ {a: 1} } />`,
			Tsx:     true,
			Options: []interface{}{opts{"spacing": spc{"objectLiterals": "always"}, "attributes": opts{"when": "never"}}},
		},
		// (c) per-side spacing.objectLiterals overrides top-level → flush object valid
		{
			Code:    `<App foo={{a: 1}} />`,
			Tsx:     true,
			Options: []interface{}{opts{"spacing": spc{"objectLiterals": "always"}, "attributes": opts{"when": "never", "spacing": spc{"objectLiterals": "never"}}}},
		},

		// ===== Real-user shapes (rslint extras) =====
		// ---- Real-user: dialog with handler + arrow (default never; all flush) ----
		{Code: `<Modal isOpen={isOpen} onClose={() => setOpen(false)} />`, Tsx: true},
		// ---- Real-user: list render — children default false, only key={...} attr checked ----
		{Code: "<ul>{items.map((i) => <li key={i.id}>{i.label}</li>)}</ul>", Tsx: true},
		// ---- Real-user: i18n key built from a template-literal attribute ----
		{Code: "<Trans i18nKey={`page.${section}.title`} />", Tsx: true},
		// ---- Real-user: style object — flush `{{...}}` valid under default objectLiterals:never ----
		{Code: `<Comp style={{ margin: 0, padding: 0 }} />`, Tsx: true},

		// ===== isObjectLiteral classification boundary =====
		// A leading block comment is the token after `{` (not `{`), so the body
		// is NOT classified as an object literal — `when` applies, not
		// objectLiterals. Flush comment+object under never stays valid even with
		// objectLiterals:'always' (which would otherwise force surrounding space).
		{Code: `<App foo={/* c */ {a: 1}} />`, Tsx: true, Options: []interface{}{opts{"when": "never", "spacing": spc{"objectLiterals": "always"}}}},
	}, []rule_tester.InvalidTestCase{
		// ===== Dimension 4 invalid: tsgo-specific shapes & real-world =====

		// ---- TS expression wrappers — default `never` reports the outer brace ----
		{
			Code:   `<App foo={ (bar) } />`,
			Tsx:    true,
			Output: []string{`<App foo={(bar)} />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 18},
			},
		},
		{
			Code:   `<App foo={ bar! } />`,
			Tsx:    true,
			Output: []string{`<App foo={bar!} />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 17},
			},
		},
		{
			Code:   `<App foo={ bar as any } />`,
			Tsx:    true,
			Output: []string{`<App foo={bar as any} />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 23},
			},
		},
		{
			Code:   `<App foo={ bar satisfies Foo } />`,
			Tsx:    true,
			Output: []string{`<App foo={bar satisfies Foo} />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 30},
			},
		},
		{
			Code:   `<App foo={ bar?.baz } />`,
			Tsx:    true,
			Output: []string{`<App foo={bar?.baz} />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 21},
			},
		},

		// ---- Non-`{` second tokens — `objectLiterals: 'always'` must NOT promote them ----
		// Confirms isObjectLiteral correctly checks the literal `{` character,
		// not "is the inner expression a JS object". `[1,2]` is an array, not
		// an object, so `objectLiterals` config does not apply.
		{
			Code:    `<App foo={ [1, 2] } />`,
			Tsx:     true,
			Options: []interface{}{opts{"when": "never", "spacing": spc{"objectLiterals": "always"}}},
			Output:  []string{`<App foo={[1, 2]} />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 19},
			},
		},
		{
			Code:    `<App foo={ <Bar /> } />`,
			Tsx:     true,
			Options: []interface{}{opts{"when": "never", "spacing": spc{"objectLiterals": "always"}}},
			Output:  []string{`<App foo={<Bar />} />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 20},
			},
		},

		// ---- String / template / regex content — scanner must not mis-tokenize ----
		// The substrings `//`, `/*`, `{`, and `}` inside a string/template/regex
		// must NOT be treated as comments or extra braces. If the scanner gets
		// this wrong, fixes corrupt source.
		{
			Code:   `<App foo={ "// not a comment" } />`,
			Tsx:    true,
			Output: []string{`<App foo={"// not a comment"} />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 31},
			},
		},
		{
			Code:   `<App foo={ "/* block */" } />`,
			Tsx:    true,
			Output: []string{`<App foo={"/* block */"} />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 26},
			},
		},
		{
			Code:   `<App foo={ "{ not an object }" } />`,
			Tsx:    true,
			Output: []string{`<App foo={"{ not an object }"} />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 32},
			},
		},
		{
			Code:   `<App foo={ /regex/g } />`,
			Tsx:    true,
			Output: []string{`<App foo={/regex/g} />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 21},
			},
		},

		// (children spread is intentionally NOT covered — see corresponding
		// valid cases above for the parity with upstream JSXSpreadChild
		// behavior.)

		// ---- Nested element: only inner attribute container is checked under default config ----
		{
			Code:   `<App><Bar foo={ baz } /></App>`,
			Tsx:    true,
			Output: []string{`<App><Bar foo={baz} /></App>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 15},
				{MessageId: "noSpaceBefore", Line: 1, Column: 21},
			},
		},

		// ---- Real-world: ternary in className (single line, allowMultiline irrelevant) ----
		{
			Code:    `<div className={ isActive ? 'a' : 'b' } />`,
			Tsx:     true,
			Options: []interface{}{opts{"when": "never"}},
			Output:  []string{`<div className={isActive ? 'a' : 'b'} />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 16},
				{MessageId: "noSpaceBefore", Line: 1, Column: 39},
			},
		},

		// ===== Robustness: complex inner contents under single-line `never` =====
		// Counterparts to the multi-line valid tests above — verify the
		// rule still fires correctly when these tricky bodies are wrapped
		// with explicit spaces under default `never`.
		{
			Code:   `<X foo={ "hello}world" } />`,
			Tsx:    true,
			Output: []string{`<X foo={"hello}world"} />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 8},
				{MessageId: "noSpaceBefore", Line: 1, Column: 24},
			},
		},
		{
			Code:   "<X foo={ `${a}${b}` } />",
			Tsx:    true,
			Output: []string{"<X foo={`${a}${b}`} />"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 8},
				{MessageId: "noSpaceBefore", Line: 1, Column: 21},
			},
		},
		{
			Code:   `<X foo={ /[}]/g } />`,
			Tsx:    true,
			Output: []string{`<X foo={/[}]/g} />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 8},
				{MessageId: "noSpaceBefore", Line: 1, Column: 17},
			},
		},
		{
			Code:    `<X foo={"hello}world"} />`,
			Tsx:     true,
			Options: []interface{}{"always"},
			Output:  []string{`<X foo={ "hello}world" } />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Line: 1, Column: 8},
				{MessageId: "spaceNeededBefore", Line: 1, Column: 22},
			},
		},
		{
			Code:    "<X foo={`${x}-${y}`} />",
			Tsx:     true,
			Options: []interface{}{"always"},
			Output:  []string{"<X foo={ `${x}-${y}` } />"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Line: 1, Column: 8},
				{MessageId: "spaceNeededBefore", Line: 1, Column: 20},
			},
		},
		{
			Code:    `<X foo={() => { return 1; }} />`,
			Tsx:     true,
			Options: []interface{}{"always"},
			Output:  []string{`<X foo={ () => { return 1; } } />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Line: 1, Column: 8},
				{MessageId: "spaceNeededBefore", Line: 1, Column: 28},
			},
		},

		// ===== BOM + Unicode invalid lock-in =====

		// BOM at file start — tsgo's scanner counts the BOM as 1 UTF-16
		// character on line 1 (not stripped during position calculation),
		// so columns shift by +1 vs the BOM-less equivalent: `{` at col 11,
		// `}` at col 17. Locked in as observable rslint behaviour; if
		// upstream ESLint differs here that is a framework-level position-
		// calc divergence, not a rule-logic issue.
		{
			Code:   "\uFEFF<App foo={ bar } />",
			Tsx:    true,
			Output: []string{"\uFEFF<App foo={bar} />"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 11},
				{MessageId: "noSpaceBefore", Line: 1, Column: 17},
			},
		},
		// BMP non-ASCII identifier — `中` and `文` each count as 1 UTF-16
		// character, so `}` of `<App foo={ 中文 } />` is at col 15.
		{
			Code:   `<App foo={ 中文 } />`,
			Tsx:    true,
			Output: []string{`<App foo={中文} />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 15},
			},
		},
		// SMP character inside a string literal — `🚀` is a UTF-16 surrogate
		// pair (2 code units). For `<App foo={ "🚀" } />`, the closing `}`
		// is at col 17. (Emoji can only appear inside strings; it is not a
		// valid TypeScript / JSX identifier-start character.)
		{
			Code:   `<App foo={ "🚀" } />`,
			Tsx:    true,
			Output: []string{`<App foo={"🚀"} />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 17},
			},
		},
		// Two SMP characters inside a string — each adds 2 UTF-16 chars,
		// so `}` of `<App foo={ "🚀🎉" } />` is at col 19.
		{
			Code:   `<App foo={ "🚀🎉" } />`,
			Tsx:    true,
			Output: []string{`<App foo={"🚀🎉"} />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 19},
			},
		},
		// always mode + non-ASCII string — verifies `spaceNeededAfter` /
		// `spaceNeededBefore` columns under `always` with multi-byte body.
		// `<App foo={"中文"} />` — `}` at col 15.
		{
			Code:    `<App foo={"中文"} />`,
			Tsx:     true,
			Options: []interface{}{"always"},
			Output:  []string{`<App foo={ "中文" } />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Line: 1, Column: 10},
				{MessageId: "spaceNeededBefore", Line: 1, Column: 15},
			},
		},
		// BMP + SMP mixed inside the same string — `中` and `文` count as
		// 1 UTF-16 char each, `🚀` counts as 2. For `<App foo={ "中文🚀" } />`
		// the closing `}` is at col 19. Locks in cross-class column math
		// in a single body and proves the byte-level scanner never lands
		// inside a multi-byte UTF-8 sequence.
		{
			Code:   `<App foo={ "中文🚀" } />`,
			Tsx:    true,
			Output: []string{`<App foo={"中文🚀"} />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 19},
			},
		},
		// SMP between two BMP — verifies the scanner ignores the surrogate
		// pair regardless of position within the body.
		// `<App foo={ "🚀中文🎉中" } />` — col counts: `{`=10, `"`=12,
		// `🚀`=13-14, `中`=15, `文`=16, `🎉`=17-18, `中`=19, `"`=20,
		// `}`=22.
		{
			Code:   `<App foo={ "🚀中文🎉中" } />`,
			Tsx:    true,
			Output: []string{`<App foo={"🚀中文🎉中"} />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 22},
			},
		},
		// Multi-byte content inside a comment — block comment with both
		// CJK and emoji. Column for `}` of `<App foo={ /* 中🚀 */ x } />`
		// is computed as: `{`=10, ` `=11, `/`=12, `*`=13, ` `=14, `中`=15,
		// `🚀`=16-17, ` `=18, `*`=19, `/`=20, ` `=21, `x`=22, ` `=23,
		// `}`=24.
		{
			Code:   `<App foo={ /* 中🚀 */ x } />`,
			Tsx:    true,
			Output: []string{`<App foo={/* 中🚀 */ x} />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 24},
			},
		},
		// Multi-byte chars BEFORE the JsxExpression on the SAME line —
		// verifies the brace's column is correctly offset by all preceding
		// UTF-16 units (BOM-precedent style verification with mixed
		// emoji+CJK preceding tokens). For
		// `const 名 = <App foo={ "🚀" } />` — col counts:
		// `c`=1..`t`=5, ` `=6, `名`=7, ` `=8, `=`=9, ` `=10, `<`=11,
		// `A`=12..`p`=14, ` `=15, `f`=16..`o`=18, `=`=19, `{`=20, ` `=21,
		// `"`=22, `🚀`=23-24, `"`=25, ` `=26, `}`=27.
		{
			Code:   `const 名 = <App foo={ "🚀" } />`,
			Tsx:    true,
			Output: []string{`const 名 = <App foo={"🚀"} />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 20},
				{MessageId: "noSpaceBefore", Line: 1, Column: 27},
			},
		},

		// ---- Robustness: block comment immediately before `}` on next line ----
		// Variant where the trailing trivia is a BLOCK comment (vs the line
		// comment case, which both upstream and rslint cannot autofix
		// without producing syntactically broken source). Here the fix is
		// well-defined: the trailing whitespace+newline collapse, and `}`
		// lands cleanly after `*/`.
		{
			Code:    "<App>{foo /* c */\n}</App>",
			Tsx:     true,
			Options: []interface{}{opts{"children": opts{"when": "never", "allowMultiline": false}}},
			Output:  []string{"<App>{foo /* c */}</App>"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noNewlineBefore", Line: 2, Column: 1},
			},
		},

		// ===== Diagnostic contract: exact message text + End positions (rslint extras) =====
		// Upstream asserts only messageId + token; these lock in the full
		// ESLint-compatible diagnostic shape — exact Message string and
		// Line/Column/EndLine/EndColumn — for every messageId across each
		// container (attribute, child, spread), including multi-line cases.
		// EndColumn is always Column+1 (the report range is the single-character
		// brace token).
		{
			// attribute container, never → noSpaceAfter / noSpaceBefore
			Code:    `<App foo={ bar } />;`,
			Tsx:     true,
			Options: []interface{}{opts{"when": "never"}},
			Output:  []string{`<App foo={bar} />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Message: "There should be no space after '{'", Line: 1, Column: 10, EndLine: 1, EndColumn: 11},
				{MessageId: "noSpaceBefore", Message: "There should be no space before '}'", Line: 1, Column: 16, EndLine: 1, EndColumn: 17},
			},
		},
		{
			// attribute container, always → spaceNeededAfter / spaceNeededBefore
			Code:    `<App foo={bar} />;`,
			Tsx:     true,
			Options: []interface{}{opts{"when": "always"}},
			Output:  []string{`<App foo={ bar } />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Message: "A space is required after '{'", Line: 1, Column: 10, EndLine: 1, EndColumn: 11},
				{MessageId: "spaceNeededBefore", Message: "A space is required before '}'", Line: 1, Column: 14, EndLine: 1, EndColumn: 15},
			},
		},
		{
			// attribute container, multi-line, never + allowMultiline:false →
			// noNewlineAfter / noNewlineBefore (the multi-line container case)
			Code:    mlAttrBar,
			Tsx:     true,
			Options: []interface{}{opts{"when": "never", "allowMultiline": false}},
			Output:  []string{mlAttrBarFixedNoSpace},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noNewlineAfter", Message: "There should be no newline after '{'", Line: 2, Column: 18, EndLine: 2, EndColumn: 19},
				{MessageId: "noNewlineBefore", Message: "There should be no newline before '}'", Line: 4, Column: 9, EndLine: 4, EndColumn: 10},
			},
		},
		{
			// child container, always → spaceNeededAfter / spaceNeededBefore
			Code:    `<App>{bar}</App>;`,
			Tsx:     true,
			Options: []interface{}{opts{"children": opts{"when": "always"}}},
			Output:  []string{`<App>{ bar }</App>;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Message: "A space is required after '{'", Line: 1, Column: 6, EndLine: 1, EndColumn: 7},
				{MessageId: "spaceNeededBefore", Message: "A space is required before '}'", Line: 1, Column: 10, EndLine: 1, EndColumn: 11},
			},
		},
		{
			// child container, multi-line, never + allowMultiline:false →
			// noNewlineAfter / noNewlineBefore (≥2 cases per container, multi-line)
			Code:    mlChildBar,
			Tsx:     true,
			Options: []interface{}{opts{"children": opts{"when": "never", "allowMultiline": false}}},
			Output:  []string{mlChildBarFixedNoSpace},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noNewlineAfter", Message: "There should be no newline after '{'", Line: 2, Column: 14, EndLine: 2, EndColumn: 15},
				{MessageId: "noNewlineBefore", Message: "There should be no newline before '}'", Line: 4, Column: 9, EndLine: 4, EndColumn: 10},
			},
		},
		{
			// spread attribute container, never → noSpaceAfter / noSpaceBefore
			Code:    `<App { ...bar } />;`,
			Tsx:     true,
			Options: []interface{}{opts{"when": "never"}},
			Output:  []string{`<App {...bar} />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Message: "There should be no space after '{'", Line: 1, Column: 6, EndLine: 1, EndColumn: 7},
				{MessageId: "noSpaceBefore", Message: "There should be no space before '}'", Line: 1, Column: 15, EndLine: 1, EndColumn: 16},
			},
		},
		{
			// spread attribute container, always → spaceNeededAfter / spaceNeededBefore
			Code:    `<App {...bar} />;`,
			Tsx:     true,
			Options: []interface{}{opts{"when": "always"}},
			Output:  []string{`<App { ...bar } />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Message: "A space is required after '{'", Line: 1, Column: 6, EndLine: 1, EndColumn: 7},
				{MessageId: "spaceNeededBefore", Message: "A space is required before '}'", Line: 1, Column: 13, EndLine: 1, EndColumn: 14},
			},
		},

		// ===== Real-user shapes (rslint extras) =====
		// ---- Real-user: spaced event handler — the most common spacing report shape ----
		{
			Code:   `<Button onClick={ handleClick }>Click</Button>`,
			Tsx:    true,
			Output: []string{`<Button onClick={handleClick}>Click</Button>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 17, EndLine: 1, EndColumn: 18},
				{MessageId: "noSpaceBefore", Line: 1, Column: 31, EndLine: 1, EndColumn: 32},
			},
		},
		// ---- Real-user: object config prop with surrounding spaces (objectLiterals default never) ----
		{
			Code:   `<Component config={ { a: 1, b: 2 } } />`,
			Tsx:    true,
			Output: []string{`<Component config={{ a: 1, b: 2 }} />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 19, EndLine: 1, EndColumn: 20},
				{MessageId: "noSpaceBefore", Line: 1, Column: 36, EndLine: 1, EndColumn: 37},
			},
		},

		// ===== @stylistic option-normalization scope: invalid lock-ins =====
		// Report-side mirror of the valid (a)/(b) cases above, asserting
		// @stylistic's `spacing = perSide.spacing ?? defaultSpacing` branches.
		// (a) per-side empty `spacing:{}` → objectLiterals falls back to
		//     when:'never', so a SPACED object literal is reported. (react would
		//     inherit the top-level 'always' and accept it — the cross-fork delta
		//     that stylisticScope=true closes.)
		{
			Code:    `<App foo={ {a: 1} } />`,
			Tsx:     true,
			Options: []interface{}{opts{"spacing": spc{"objectLiterals": "always"}, "attributes": opts{"when": "never", "spacing": spc{}}}},
			Output:  []string{`<App foo={{a: 1}} />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Message: "There should be no space after '{'", Line: 1, Column: 10, EndLine: 1, EndColumn: 11},
				{MessageId: "noSpaceBefore", Message: "There should be no space before '}'", Line: 1, Column: 19, EndLine: 1, EndColumn: 20},
			},
		},
		// (b) per-side WITHOUT a spacing key → inherits the top-level
		//     objectLiterals:'always', so a FLUSH object literal is reported as
		//     needing surrounding space.
		{
			Code:    `<App foo={{a: 1}} />`,
			Tsx:     true,
			Options: []interface{}{opts{"spacing": spc{"objectLiterals": "always"}, "attributes": opts{"when": "never"}}},
			Output:  []string{`<App foo={ {a: 1} } />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Message: "A space is required after '{'", Line: 1, Column: 10, EndLine: 1, EndColumn: 11},
				{MessageId: "spaceNeededBefore", Message: "A space is required before '}'", Line: 1, Column: 17, EndLine: 1, EndColumn: 18},
			},
		},
	})
}

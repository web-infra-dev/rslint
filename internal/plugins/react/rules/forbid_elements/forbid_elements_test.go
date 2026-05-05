package forbid_elements

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// optsForbid* helpers wrap the option object in a `[]interface{}` so the
// JSON-array path (used by both the JS rule_tester and the CLI) is always
// exercised. Bare-map shapes are covered by dedicated cases under the
// "Options-shape coverage" section.
func optsForbid(items ...interface{}) []interface{} {
	return []interface{}{map[string]interface{}{"forbid": items}}
}

func TestForbidElementsRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ForbidElementsRule, []rule_tester.ValidTestCase{
		// ============================================================
		// Upstream valid cases (all migrated)
		// ============================================================

		// No options at all → rule is a no-op (nothing in forbid list).
		{Code: `<button />`, Tsx: true},

		// Empty forbid array → rule is a no-op.
		{Code: `<button />`, Tsx: true, Options: optsForbid()},

		// PascalCase JSX tag does not match the lowercase forbid entry.
		{Code: `<Button />`, Tsx: true, Options: optsForbid("button")},

		// Object form — same as above.
		{
			Code:    `<Button />`,
			Tsx:     true,
			Options: optsForbid(map[string]interface{}{"element": "button"}),
		},

		// Identifier callee whose name starts lowercase falls through —
		// upstream's `/^[A-Z_]/` check rejects `button`.
		{Code: `React.createElement(button)`, Tsx: true, Options: optsForbid("button")},

		// Bare `createElement(...)` (no React.) is not recognized as a
		// createElement call without an import binding.
		{Code: `createElement("button")`, Tsx: true, Options: optsForbid("button")},

		// Non-React member-access callee is not treated as createElement.
		{Code: `NotReact.createElement("button")`, Tsx: true, Options: optsForbid("button")},

		// String literal whose first char is `_` doesn't match `/^[a-z]/`,
		// so the literal branch silently skips it.
		{Code: `React.createElement("_thing")`, Tsx: true, Options: optsForbid("_thing")},

		// String literal whose first char is uppercase doesn't match
		// `/^[a-z]/`. (Use the Identifier form instead for components.)
		{Code: `React.createElement("Modal")`, Tsx: true, Options: optsForbid("Modal")},

		// String literal containing `.` is rejected by `/^[a-z][^.]*$/`.
		{
			Code:    `React.createElement("dotted.component")`,
			Tsx:     true,
			Options: optsForbid("dotted.component"),
		},

		// Non-Identifier / non-Literal / non-MemberExpression argument shapes
		// — function expression, object literal, numeric literal — silently
		// fall through.
		{Code: `React.createElement(function() {})`, Tsx: true, Options: optsForbid("button")},
		{Code: `React.createElement({})`, Tsx: true, Options: optsForbid("button")},
		{Code: `React.createElement(1)`, Tsx: true, Options: optsForbid("button")},

		// No arguments → bail out.
		{Code: `React.createElement()`, Tsx: true},

		// ============================================================
		// Extra Go-side coverage (lock-in cases for branches upstream
		// doesn't exercise)
		// ============================================================

		// Identifier starting with `$` is not in `[A-Z_]`, so the bare
		// identifier branch is skipped. Locks in upstream's regex.
		{Code: `React.createElement($Foo)`, Tsx: true, Options: optsForbid("$Foo")},

		// Lowercase digit-mixed identifier — fails `/^[A-Z_]/`.
		{Code: `React.createElement(button2)`, Tsx: true, Options: optsForbid("button2")},

		// Single lowercase letter — matches `/^[a-z][^.]*$/` — but if forbid
		// doesn't match exactly, no report.
		{Code: `React.createElement("a")`, Tsx: true, Options: optsForbid("button")},

		// Numeric-leading literal (parsed as StringLiteral) — fails `/^[a-z]/`.
		{Code: `React.createElement("0abc")`, Tsx: true, Options: optsForbid("0abc")},

		// Template literal (NoSubstitutionTemplateLiteral) — upstream's
		// `argument.type === 'Literal'` doesn't match TemplateLiteral.
		{Code: "React.createElement(`button`)", Tsx: true, Options: optsForbid("button")},

		// Argument is a `(Modal)` parenthesized identifier — `SkipParentheses`
		// unwraps to the bare Identifier and applies the regex check.
		// `Modal` matches but the forbid list is empty for it, so no report.
		{Code: `React.createElement((Modal))`, Tsx: true, Options: optsForbid("Other")},

		// `Modal!` (TS non-null assertion) — not an Identifier after parens
		// skip, so falls through unreported. Locks in the no-unwrap behavior.
		{Code: `React.createElement(Modal!)`, Tsx: true, Options: optsForbid("Modal")},

		// `Modal as any` — TS as-expression, also falls through unreported.
		{Code: `React.createElement(Modal as any)`, Tsx: true, Options: optsForbid("Modal")},

		// MemberExpression but forbid doesn't match.
		{
			Code:    `React.createElement(other.Component)`,
			Tsx:     true,
			Options: optsForbid("dotted.Component"),
		},

		// JSX dotted form, forbid doesn't match.
		{
			Code:    `<Other.Component />`,
			Tsx:     true,
			Options: optsForbid("dotted.Component"),
		},

		// Optional-chain member access — upstream wraps it in ChainExpression
		// and the `argument.type === 'MemberExpression'` check fails. tsgo
		// encodes the optional chain as a flag on the same kind, so we must
		// gate via ast.IsOptionalChain to keep parity. Locks in the skip.
		{
			Code:    `React.createElement(a?.b)`,
			Tsx:     true,
			Options: optsForbid("a?.b", "a.b"),
		},

		// Optional-chain through dotted chain — same skip.
		{
			Code:    `React.createElement(a.b?.c)`,
			Tsx:     true,
			Options: optsForbid("a.b?.c", "a.b.c"),
		},

		// Optional in chain interior — `a?.b.c`. tsgo propagates
		// `NodeFlagsOptionalChain` up the chain, so the outer
		// PropertyAccessExpression is also flagged. Mirrors upstream's
		// ChainExpression wrapping. Skipped.
		{
			Code:    `React.createElement(a?.b.c)`,
			Tsx:     true,
			Options: optsForbid("a?.b.c", "a.b.c"),
		},

		// Optional in chain middle — `a.b?.c.d`.
		{
			Code:    `React.createElement(a.b?.c.d)`,
			Tsx:     true,
			Options: optsForbid("a.b?.c.d", "a.b.c.d"),
		},

		// Optional in chain tail — `a.b.c?.d`.
		{
			Code:    `React.createElement(a.b.c?.d)`,
			Tsx:     true,
			Options: optsForbid("a.b.c?.d", "a.b.c.d"),
		},

		// Optional in element-access form — `a?.["b"]`.
		{
			Code:    `React.createElement(a?.["b"])`,
			Tsx:     true,
			Options: optsForbid(`a?.["b"]`, `a["b"]`),
		},

		// Inner optional-chain wrapped in parens — `(a?.b)`. Inner has
		// optional flag, SkipParentheses unwraps, IsOptionalChain returns
		// true, skipped.
		{
			Code:    `React.createElement((a?.b))`,
			Tsx:     true,
			Options: optsForbid("a?.b", "a.b"),
		},

		// JSX namespaced form whose `ns:tag` doesn't match the forbid list.
		{
			Code:    `<ns:other />`,
			Tsx:     true,
			Options: optsForbid("ns:foo"),
		},

		// JSX three-level chain whose canonical text doesn't match.
		{
			Code:    `<Foo.Bar.Baz />`,
			Tsx:     true,
			Options: optsForbid("Foo.Bar.Other"),
		},

		// `<this.Foo />` — JSX dotted with `this` base; upstream's
		// `getText` returns `"this.Foo"`. Lock-in: forbid that doesn't match
		// stays valid.
		{
			Code:    `<this.Foo />`,
			Tsx:     true,
			Options: optsForbid("self.Foo"),
		},

		// Generic createElement — `React.createElement<Props>(Modal)` —
		// the type-args list is on `call.TypeArguments`, separate from the
		// argument list, so the listener still treats `Modal` as the first
		// argument. Locks in that we don't accidentally read the type-args
		// list as the value-args list.
		{
			Code:    `React.createElement<{x:number}>(Other)`,
			Tsx:     false,
			Options: optsForbid("Modal"),
		},

		// ============================================================
		// Options-shape coverage — bare-map (single-option CLI shape).
		// ============================================================

		// Bare-map options (matches the post-unwrap CLI shape) with no
		// matching element → no report.
		{
			Code:    `<Button />`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"button"}},
		},

		// ============================================================
		// settings.react.pragma — custom pragma support.
		// ============================================================

		// Custom pragma → `React.createElement` no longer matches when
		// `pragma = "Preact"`. Locks in pragma routing.
		{
			Code:     `React.createElement("button")`,
			Tsx:      true,
			Options:  optsForbid("button"),
			Settings: map[string]interface{}{"react": map[string]interface{}{"pragma": "Preact"}},
		},

		// Renamed-import + custom-pragma combo: `import { h } from 'preact'`
		// with `pragma = "h"`. Upstream's `isCreateElement` checks
		// literal name `createElement` on the bare callee, so `h('button')`
		// does NOT match — same in our port. Locks in literal-name gate.
		{
			Code: `
import { h } from 'preact';
h('button');
`,
			Tsx:      true,
			Options:  optsForbid("button"),
			Settings: map[string]interface{}{"react": map[string]interface{}{"pragma": "h"}},
		},

		// ============================================================
		// Argument shapes that fall through unreported (real user code)
		// ============================================================

		// Spread first argument — `KindSpreadElement` is not a candidate
		// kind, falls through.
		{Code: `React.createElement(...args)`, Tsx: true, Options: optsForbid("button")},

		// ConditionalExpression first argument — falls through (upstream
		// `argument.type` matches none of Identifier/Literal/MemberExpression).
		{
			Code:    `React.createElement(condition ? "button" : "span")`,
			Tsx:     true,
			Options: optsForbid("button", "span"),
		},

		// LogicalExpression first argument — falls through.
		{
			Code:    `React.createElement(maybe || "button")`,
			Tsx:     true,
			Options: optsForbid("button"),
		},

		// CallExpression first argument — falls through.
		{
			Code:    `React.createElement(getTag())`,
			Tsx:     true,
			Options: optsForbid("button"),
		},

		// Variable-bound lowercase identifier — `const tag = 'button';
		// React.createElement(tag)` → bare Identifier 'tag' fails
		// `/^[A-Z_]/`, no report. Locks in upstream regex (which checks
		// the identifier's *name*, not the resolved value).
		{
			Code:    `const tag = "button"; React.createElement(tag);`,
			Tsx:     true,
			Options: optsForbid("button", "tag"),
		},

		// Tagged template first argument — `KindTaggedTemplateExpression`
		// not in switch.
		{
			Code:    "React.createElement(html`button`)",
			Tsx:     true,
			Options: optsForbid("button"),
		},

		// Template literal with substitution — `KindTemplateExpression`
		// not in switch (only NoSubstitutionTemplateLiteral could be a
		// candidate, and even that is excluded to match upstream).
		{
			Code:    "React.createElement(`button${x}`)",
			Tsx:     true,
			Options: optsForbid("button"),
		},

		// Renamed import: `import { createElement as h } from 'react'; h('button')`.
		// The bare callee is `h`, not `createElement`, so upstream's literal-name
		// check (`callee.name === 'createElement'`) fails — same in our port.
		// Locks in: rename does NOT participate.
		{
			Code: `
import { createElement as h } from 'react';
h('button');
`,
			Tsx:     true,
			Options: optsForbid("button"),
		},

		// `a.b.createElement('button')` — non-pragma deep member-access
		// callee falls through (only `<pragma>.createElement` is matched).
		{
			Code:    `a.b.createElement("button")`,
			Tsx:     true,
			Options: optsForbid("button"),
		},

		// `React.createElement` as method reference (no call) —
		// CallExpression listener doesn't fire on PropertyAccess alone.
		{Code: `const ref = React.createElement;`, Tsx: true, Options: optsForbid("button")},

		// JSX inside JSX expression container — outer `<div>` is fine,
		// inner `<Foo>` is not in forbid list.
		{
			Code:    `<div>{condition && <Foo />}</div>`,
			Tsx:     true,
			Options: optsForbid("button"),
		},

		// JSX fragment containing only non-forbidden tags.
		{
			Code:    `<><Foo /><Bar /></>`,
			Tsx:     true,
			Options: optsForbid("button"),
		},

		// Class component method returning forbidden tag of a different
		// name — only the method body's JSX would fire. Forbid set to a
		// non-matching name, so all valid.
		{
			Code: `
class C {
  render() {
    return <Foo />;
  }
}
`,
			Tsx:     true,
			Options: optsForbid("button"),
		},

		// JSX inside arrow inside JSX inside class — deep nesting with no
		// forbidden tag.
		{
			Code: `
class C {
  render() {
    return <ul>{items.map(item => <Foo key={item.id}>{item.label}</Foo>)}</ul>;
  }
}
`,
			Tsx:     true,
			Options: optsForbid("button"),
		},

		// ============================================================
		// String-literal regex boundaries — `/^[a-z][^.]*$/`.
		// ============================================================

		// Uppercase first char fails `/^[a-z]/`.
		{Code: `React.createElement("UPPER")`, Tsx: true, Options: optsForbid("UPPER")},

		// Leading whitespace fails `/^[a-z]/` (space, not letter).
		{Code: `React.createElement("  button")`, Tsx: true, Options: optsForbid("  button", "button")},

		// Identifier `undefined` (lowercase) fails `/^[A-Z_]/`.
		{Code: `React.createElement(undefined)`, Tsx: true, Options: optsForbid("undefined")},

		// String containing dot fails `/[^.]*$/`.
		{Code: `React.createElement("a.b")`, Tsx: true, Options: optsForbid("a.b")},

		// ============================================================
		// TypeScript-specific arg wrappers.
		// ============================================================

		// `Modal satisfies Component` — TS SatisfiesExpression, falls
		// through unreported (only Identifier / StringLiteral /
		// MemberExpression are matched, matching upstream).
		{
			Code:    `React.createElement(Modal satisfies Component)`,
			Tsx:     false,
			Options: optsForbid("Modal"),
		},

		// JSX generic `<Modal<Props> />` — TS generic JSX. Tag name is the
		// PropertyAccess / Identifier; the type arguments hang off
		// separately. Forbid: ['Other'] doesn't match `Modal`, valid.
		{
			Code:    `<Modal<{x: number}> />`,
			Tsx:     true,
			Options: optsForbid("Other"),
		},

		// Type-only import — `import type { createElement } from 'react'`.
		// In TS, type imports don't introduce a value binding. Our
		// `IsDestructuredFromPragmaImport` syntax-fallback would ideally
		// match this (it's purely structural), but at runtime the
		// reference would error. Locks in: even if matched, the user's
		// code is degenerate; report behavior should follow whatever
		// `IsDestructuredFromPragmaImport` decides — covered by the
		// existing import test elsewhere.
		// (Test left as forbid-non-matching valid to avoid asserting
		//  syntax-fallback specifics.)
		{
			Code: `
import type { createElement } from 'react';
const x = createElement;
`,
			Tsx:     true,
			Options: optsForbid("Modal"),
		},

		// ============================================================
		// Boolean / null keyword literals as createElement arg.
		// ============================================================

		// Boolean keyword arg — non-matching forbid → no report.
		{Code: `React.createElement(true)`, Tsx: true, Options: optsForbid("button")},
		{Code: `React.createElement(false)`, Tsx: true, Options: optsForbid("button")},
		{Code: `React.createElement(null)`, Tsx: true, Options: optsForbid("button")},

		// ============================================================
		// Callee shapes that don't match `<pragma>.createElement`.
		// ============================================================

		// Bracket-notation callee — `React['createElement']` is not
		// `<pragma>.createElement` (PropertyAccessExpression), it's an
		// ElementAccessExpression. Both upstream and ours skip.
		{Code: `React["createElement"]("button")`, Tsx: true, Options: optsForbid("button")},

		// `.bind(...)`-wrapped — outer call's callee is itself a CallExpression,
		// not PropertyAccess. Skipped.
		{
			Code:    `React.createElement.bind(null)("button")`,
			Tsx:     true,
			Options: optsForbid("button"),
		},

		// `.apply(...)`-wrapped — same shape.
		{
			Code:    `React.createElement.apply(null, ["button"])`,
			Tsx:     true,
			Options: optsForbid("button"),
		},

	}, []rule_tester.InvalidTestCase{
		// ============================================================
		// Upstream invalid cases (all migrated)
		// ============================================================

		// JSX intrinsic tag matches lowercase forbid entry.
		{
			Code:    `<button />`,
			Tsx:     true,
			Options: optsForbid("button"),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   "<button> is forbidden",
					Line:      1,
					Column:    2,
				},
			},
		},

		// Two JSX elements in one expression — both reported, in source order.
		{
			Code:    `[<Modal />, <button />]`,
			Tsx:     true,
			Options: optsForbid("button", "Modal"),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   "<Modal> is forbidden",
					Line:      1,
					Column:    3,
				},
				{
					MessageId: "forbiddenElement",
					Message:   "<button> is forbidden",
					Line:      1,
					Column:    14,
				},
			},
		},

		// JSX dotted tag name.
		{
			Code:    `<dotted.component />`,
			Tsx:     true,
			Options: optsForbid("dotted.component"),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   "<dotted.component> is forbidden",
					Line:      1,
					Column:    2,
				},
			},
		},

		// JSX dotted with custom message.
		{
			Code: `<dotted.Component />`,
			Tsx:  true,
			Options: optsForbid(map[string]interface{}{
				"element": "dotted.Component",
				"message": "that ain't cool",
			}),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement_message",
					Message:   "<dotted.Component> is forbidden, that ain't cool",
					Line:      1,
					Column:    2,
				},
			},
		},

		// JSX intrinsic with custom message.
		{
			Code: `<button />`,
			Tsx:  true,
			Options: optsForbid(map[string]interface{}{
				"element": "button",
				"message": "use <Button> instead",
			}),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement_message",
					Message:   "<button> is forbidden, use <Button> instead",
					Line:      1,
					Column:    2,
				},
			},
		},

		// Nested JSX (opening element + child) — both fire when forbid lists
		// hit them.
		{
			Code: `<button><input /></button>`,
			Tsx:  true,
			Options: optsForbid(
				map[string]interface{}{"element": "button"},
				map[string]interface{}{"element": "input"},
			),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenElement", Message: "<button> is forbidden", Line: 1, Column: 2},
				{MessageId: "forbiddenElement", Message: "<input> is forbidden", Line: 1, Column: 10},
			},
		},

		// Nested JSX with mixed string + object forbid entries.
		{
			Code:    `<button><input /></button>`,
			Tsx:     true,
			Options: optsForbid(map[string]interface{}{"element": "button"}, "input"),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenElement", Message: "<button> is forbidden", Line: 1, Column: 2},
				{MessageId: "forbiddenElement", Message: "<input> is forbidden", Line: 1, Column: 10},
			},
		},

		// Same again but reverse order — order in `forbid` doesn't matter,
		// only AST order matters.
		{
			Code:    `<button><input /></button>`,
			Tsx:     true,
			Options: optsForbid("input", map[string]interface{}{"element": "button"}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenElement", Message: "<button> is forbidden", Line: 1, Column: 2},
				{MessageId: "forbiddenElement", Message: "<input> is forbidden", Line: 1, Column: 10},
			},
		},

		// Duplicate forbid entries for the same element — last write wins,
		// so the message is "use <Button2> instead".
		{
			Code: `<button />`,
			Tsx:  true,
			Options: optsForbid(
				map[string]interface{}{"element": "button", "message": "use <Button> instead"},
				map[string]interface{}{"element": "button", "message": "use <Button2> instead"},
			),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement_message",
					Message:   "<button> is forbidden, use <Button2> instead",
					Line:      1,
					Column:    2,
				},
			},
		},

		// React.createElement string literal with extra arguments.
		{
			Code:    `React.createElement("button", {}, child)`,
			Tsx:     true,
			Options: optsForbid("button"),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   "<button> is forbidden",
					Line:      1,
					Column:    21,
				},
			},
		},

		// React.createElement Identifier and Literal forms in one array.
		{
			Code:    `[React.createElement(Modal), React.createElement("button")]`,
			Tsx:     true,
			Options: optsForbid("button", "Modal"),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   "<Modal> is forbidden",
					Line:      1,
					Column:    22,
				},
				{
					MessageId: "forbiddenElement",
					Message:   "<button> is forbidden",
					Line:      1,
					Column:    50,
				},
			},
		},

		// React.createElement MemberExpression with custom message.
		{
			Code: `React.createElement(dotted.Component)`,
			Tsx:  true,
			Options: optsForbid(map[string]interface{}{
				"element": "dotted.Component",
				"message": "that ain't cool",
			}),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement_message",
					Message:   "<dotted.Component> is forbidden, that ain't cool",
					Line:      1,
					Column:    21,
				},
			},
		},

		// React.createElement MemberExpression with lowercase rightmost.
		{
			Code:    `React.createElement(dotted.component)`,
			Tsx:     true,
			Options: optsForbid("dotted.component"),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   "<dotted.component> is forbidden",
					Line:      1,
					Column:    21,
				},
			},
		},

		// `_`-prefixed identifier matches the `[A-Z_]` regex.
		{
			Code:    `React.createElement(_comp)`,
			Tsx:     true,
			Options: optsForbid("_comp"),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   "<_comp> is forbidden",
					Line:      1,
					Column:    21,
				},
			},
		},

		// React.createElement string literal with custom message.
		{
			Code: `React.createElement("button")`,
			Tsx:  true,
			Options: optsForbid(map[string]interface{}{
				"element": "button",
				"message": "use <Button> instead",
			}),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement_message",
					Message:   "<button> is forbidden, use <Button> instead",
					Line:      1,
					Column:    21,
				},
			},
		},

		// Nested React.createElement calls — both fire.
		{
			Code: `React.createElement("button", {}, React.createElement("input"))`,
			Tsx:  true,
			Options: optsForbid(
				map[string]interface{}{"element": "button"},
				map[string]interface{}{"element": "input"},
			),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   "<button> is forbidden",
					Line:      1,
					Column:    21,
				},
				{
					MessageId: "forbiddenElement",
					Message:   "<input> is forbidden",
					Line:      1,
					Column:    55,
				},
			},
		},

		// ============================================================
		// Extra Go-side coverage (lock-in cases for branches upstream
		// doesn't exercise)
		// ============================================================

		// Bare-map options (single-option CLI shape) — confirms GetOptionsMap
		// is exercised end-to-end for the JSON-object path.
		{
			Code:    `<button />`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"button"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   "<button> is forbidden",
					Line:      1,
					Column:    2,
				},
			},
		},

		// Three-segment dotted JSX tag — locks in `tagNameSourceText` raw
		// source extraction (as opposed to ESTree-style canonicalization).
		{
			Code:    `<a.b.c />`,
			Tsx:     true,
			Options: optsForbid("a.b.c"),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   "<a.b.c> is forbidden",
					Line:      1,
					Column:    2,
				},
			},
		},

		// Multi-line JSX — the report position uses the tag-name's line.
		{
			Code: `<div>
  <button />
</div>`,
			Tsx:     true,
			Options: optsForbid("button"),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   "<button> is forbidden",
					Line:      2,
					Column:    4,
				},
			},
		},

		// Multi-line React.createElement — the report position is at the
		// (string-literal) argument.
		{
			Code: `React.createElement(
  "button"
)`,
			Tsx:     true,
			Options: optsForbid("button"),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   "<button> is forbidden",
					Line:      2,
					Column:    3,
				},
			},
		},

		// Bare `createElement(...)` after destructured import — locks in the
		// `IsCreateElementCallWithChecker` second branch (parity with
		// upstream's `isCreateElement` which handles destructured imports).
		{
			Code: `
import { createElement } from 'react';
createElement('button');
`,
			Tsx:     true,
			Options: optsForbid("button"),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   "<button> is forbidden",
					Line:      3,
					Column:    15,
				},
			},
		},

		// Bare `createElement(...)` after `const { createElement } = React` —
		// also locks in the second branch.
		{
			Code: `
const { createElement } = React;
createElement(Modal);
`,
			Tsx:     true,
			Options: optsForbid("Modal"),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   "<Modal> is forbidden",
					Line:      3,
					Column:    15,
				},
			},
		},

		// Argument wrapped in parens — `SkipParentheses` unwraps to the
		// underlying Identifier and we report at the identifier (no parens).
		{
			Code:    `React.createElement((Modal))`,
			Tsx:     true,
			Options: optsForbid("Modal"),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   "<Modal> is forbidden",
					Line:      1,
					Column:    22,
				},
			},
		},

		// Element-access form `a["b"]` — upstream's `argument.type ===
		// 'MemberExpression'` matches computed access too, and `getText`
		// returns the literal source `a["b"]` (with brackets/quotes). We
		// catch `KindElementAccessExpression` and use the same source-text
		// extraction.
		{
			Code:    `React.createElement(a["b"])`,
			Tsx:     true,
			Options: optsForbid(`a["b"]`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   `<a["b"]> is forbidden`,
					Line:      1,
					Column:    21,
				},
			},
		},

		// JSX namespaced name — `<ns:foo />` matches forbid: ['ns:foo'].
		{
			Code:    `<ns:foo />`,
			Tsx:     true,
			Options: optsForbid("ns:foo"),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   "<ns:foo> is forbidden",
					Line:      1,
					Column:    2,
				},
			},
		},

		// JSX `<this.Foo />` — locks in the `this`-rooted dotted form.
		{
			Code:    `<this.Foo />`,
			Tsx:     true,
			Options: optsForbid("this.Foo"),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   "<this.Foo> is forbidden",
					Line:      1,
					Column:    2,
				},
			},
		},

		// Three-segment dotted JSX in createElement form.
		{
			Code:    `React.createElement(Foo.Bar.Baz)`,
			Tsx:     true,
			Options: optsForbid("Foo.Bar.Baz"),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   "<Foo.Bar.Baz> is forbidden",
					Line:      1,
					Column:    21,
				},
			},
		},

		// Deeply nested JSX — opening + closing form (`<button>...</button>`)
		// triggers the JsxOpeningElement listener; the inner self-closing
		// `<input />` triggers JsxSelfClosingElement. Both report.
		{
			Code: `<div><span><button /></span><input /></div>`,
			Tsx:  true,
			Options: optsForbid(
				map[string]interface{}{"element": "button"},
				map[string]interface{}{"element": "input"},
			),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   "<button> is forbidden",
					Line:      1,
					Column:    13,
				},
				{
					MessageId: "forbiddenElement",
					Message:   "<input> is forbidden",
					Line:      1,
					Column:    30,
				},
			},
		},

		// Generic createElement — `React.createElement<Props>(Modal)` —
		// the type-args list is on `call.TypeArguments` (separate from the
		// value-args list), so `Modal` is still the first value argument.
		{
			Code:    `React.createElement<{x:number}>(Modal)`,
			Tsx:     false,
			Options: optsForbid("Modal"),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   "<Modal> is forbidden",
					Line:      1,
					Column:    33,
				},
			},
		},

		// ============================================================
		// settings.react.pragma — custom pragma routes correctly.
		// ============================================================

		// `Preact.createElement('button')` with `pragma = "Preact"` →
		// reports.
		{
			Code:     `Preact.createElement("button")`,
			Tsx:      true,
			Options:  optsForbid("button"),
			Settings: map[string]interface{}{"react": map[string]interface{}{"pragma": "Preact"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   "<button> is forbidden",
					Line:      1,
					Column:    22,
				},
			},
		},

		// Mixed JSX + createElement in one expression — both fire in
		// AST traversal order.
		{
			Code:    `<div>{React.createElement("button")}</div>`,
			Tsx:     true,
			Options: optsForbid("button"),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   "<button> is forbidden",
					Line:      1,
					Column:    27,
				},
			},
		},

		// JSX fragment with a forbidden child.
		{
			Code:    `<><button /></>`,
			Tsx:     true,
			Options: optsForbid("button"),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   "<button> is forbidden",
					Line:      1,
					Column:    4,
				},
			},
		},

		// Conditional render with forbidden tag — JSX listener fires.
		{
			Code:    `<div>{condition && <button />}</div>`,
			Tsx:     true,
			Options: optsForbid("button"),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   "<button> is forbidden",
					Line:      1,
					Column:    21,
				},
			},
		},

		// Array map with forbidden tag.
		{
			Code:    `<ul>{items.map(item => <li key={item.id}>{item.label}</li>)}</ul>`,
			Tsx:     true,
			Options: optsForbid("li"),
			Errors: []rule_tester.InvalidTestCaseError{
				// opening <li ...>
				{
					MessageId: "forbiddenElement",
					Message:   "<li> is forbidden",
					Line:      1,
					Column:    25,
				},
			},
		},

		// JSX with attributes — attribute presence does not affect the
		// tag-name match. `<button id="foo" />` reports just like `<button />`.
		{
			Code:    `<button id="foo" />`,
			Tsx:     true,
			Options: optsForbid("button"),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   "<button> is forbidden",
					Line:      1,
					Column:    2,
				},
			},
		},

		// Class component method returning forbidden JSX — JSX listener
		// fires inside the class method body.
		{
			Code: `
class C {
  render() {
    return <button id="x" />;
  }
}
`,
			Tsx:     true,
			Options: optsForbid("button"),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   "<button> is forbidden",
					Line:      4,
					Column:    13,
				},
			},
		},

		// Deep nesting — JSX inside arrow inside JSX inside class — must
		// fire for every forbidden tag, regardless of containment depth.
		{
			Code: `
class C {
  render() {
    return <ul>{items.map(i => <li><button>{i}</button></li>)}</ul>;
  }
}
`,
			Tsx:     true,
			Options: optsForbid("button", "li"),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   "<li> is forbidden",
					Line:      4,
					Column:    33,
				},
				{
					MessageId: "forbiddenElement",
					Message:   "<button> is forbidden",
					Line:      4,
					Column:    37,
				},
			},
		},

		// Deeply-dotted MemberExpression as createElement argument.
		{
			Code:    `React.createElement(Foo.Bar.Baz.Qux)`,
			Tsx:     true,
			Options: optsForbid("Foo.Bar.Baz.Qux"),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   "<Foo.Bar.Baz.Qux> is forbidden",
					Line:      1,
					Column:    21,
				},
			},
		},

		// Mixed-form: MemberExpression with computed step.
		// `React.createElement(a.b["c"])` — element-access wraps a member.
		// Source text `a.b["c"]` matches forbid.
		{
			Code:    `React.createElement(a.b["c"])`,
			Tsx:     true,
			Options: optsForbid(`a.b["c"]`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   `<a.b["c"]> is forbidden`,
					Line:      1,
					Column:    21,
				},
			},
		},

		// ============================================================
		// Boolean / null literal as createElement arg — `String(value)`
		// matches `/^[a-z][^.]*$/` for "true" / "false" / "null".
		// ============================================================

		{
			Code:    `React.createElement(true)`,
			Tsx:     true,
			Options: optsForbid("true"),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   "<true> is forbidden",
					Line:      1,
					Column:    21,
				},
			},
		},
		{
			Code:    `React.createElement(false)`,
			Tsx:     true,
			Options: optsForbid("false"),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   "<false> is forbidden",
					Line:      1,
					Column:    21,
				},
			},
		},
		{
			Code:    `React.createElement(null)`,
			Tsx:     true,
			Options: optsForbid("null"),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   "<null> is forbidden",
					Line:      1,
					Column:    21,
				},
			},
		},

		// ============================================================
		// Optional-chain forms on the createElement callee — both branches
		// match upstream's `isCreateElement` and report.
		// ============================================================

		// Optional-chain on the call itself: `React.createElement?.('button')`.
		// Only the call is optional; the callee MemberExpression is not.
		// Reports. `"button"` literal starts at column 23 (after `?.(`).
		{
			Code:    `React.createElement?.("button")`,
			Tsx:     true,
			Options: optsForbid("button"),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   "<button> is forbidden",
					Line:      1,
					Column:    23,
				},
			},
		},

		// Optional-chain on the access: `React?.createElement('button')`.
		// Upstream's `isCreateElement` doesn't inspect the optional flag,
		// so this is a `createElement` call; we report. (We use a local
		// helper so the rule isn't subject to the shared
		// `IsCreateElementCallWithChecker` conservative-skip on optional
		// callees.) `"button"` literal starts at column 22 (after `(`).
		{
			Code:    `React?.createElement("button")`,
			Tsx:     true,
			Options: optsForbid("button"),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   "<button> is forbidden",
					Line:      1,
					Column:    22,
				},
			},
		},

		// Both optional: `React?.createElement?.('button')` — also reports.
		{
			Code:    `React?.createElement?.("button")`,
			Tsx:     true,
			Options: optsForbid("button"),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   "<button> is forbidden",
					Line:      1,
					Column:    24,
				},
			},
		},

		// Optional access on createElement Identifier arg —
		// `React?.createElement(Modal)`.
		{
			Code:    `React?.createElement(Modal)`,
			Tsx:     true,
			Options: optsForbid("Modal"),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   "<Modal> is forbidden",
					Line:      1,
					Column:    22,
				},
			},
		},

		// ============================================================
		// String-literal regex boundary — values that DO match
		// `/^[a-z][^.]*$/`.
		// ============================================================

		// Hyphen in name — passes `[^.]*$` (no dot), matches `[a-z]` start.
		// Real example: web-component custom elements.
		{
			Code:    `React.createElement("custom-element")`,
			Tsx:     true,
			Options: optsForbid("custom-element"),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   "<custom-element> is forbidden",
					Line:      1,
					Column:    21,
				},
			},
		},

		// Digit-bearing name — `button2`. Matches.
		{
			Code:    `React.createElement("button2")`,
			Tsx:     true,
			Options: optsForbid("button2"),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   "<button2> is forbidden",
					Line:      1,
					Column:    21,
				},
			},
		},

		// Single-char element name — `a`.
		{
			Code:    `React.createElement("a")`,
			Tsx:     true,
			Options: optsForbid("a"),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   "<a> is forbidden",
					Line:      1,
					Column:    21,
				},
			},
		},

		// ============================================================
		// forbid configuration boundaries.
		// ============================================================

		// forbid object with empty message — falls back to forbiddenElement
		// (no `_message` suffix), matching upstream's truthiness check.
		{
			Code: `<button />`,
			Tsx:  true,
			Options: optsForbid(map[string]interface{}{
				"element": "button",
				"message": "",
			}),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   "<button> is forbidden",
					Line:      1,
					Column:    2,
				},
			},
		},

		// forbid object with extra unknown keys — silently ignored, the
		// `element` and `message` fields still work. Lock-in: extra
		// keys do not break parsing.
		{
			Code: `<button />`,
			Tsx:  true,
			Options: optsForbid(map[string]interface{}{
				"element": "button",
				"message": "use Button",
				"extra":   "ignored",
				"flag":    true,
			}),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement_message",
					Message:   "<button> is forbidden, use Button",
					Line:      1,
					Column:    2,
				},
			},
		},

		// Mixed forbid list — string, object-with-message, object-without
		// — within one config; each entry tested in one expression.
		{
			Code: `<><button /><input /><img /></>`,
			Tsx:  true,
			Options: optsForbid(
				"img",
				map[string]interface{}{"element": "button", "message": "use Button"},
				map[string]interface{}{"element": "input"},
			),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement_message",
					Message:   "<button> is forbidden, use Button",
					Line:      1,
					Column:    4,
				},
				{
					MessageId: "forbiddenElement",
					Message:   "<input> is forbidden",
					Line:      1,
					Column:    14,
				},
				{
					MessageId: "forbiddenElement",
					Message:   "<img> is forbidden",
					Line:      1,
					Column:    23,
				},
			},
		},

		// ============================================================
		// Real-world codebase patterns.
		// ============================================================

		// Module-level `export default <jsx />` — JSX listener fires.
		{
			Code:    `export default <button />;`,
			Tsx:     true,
			Options: optsForbid("button"),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   "<button> is forbidden",
					Line:      1,
					Column:    17,
				},
			},
		},

		// Arrow body returning JSX — common in real React code.
		{
			Code:    `const App = () => <button />;`,
			Tsx:     true,
			Options: optsForbid("button"),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   "<button> is forbidden",
					Line:      1,
					Column:    20,
				},
			},
		},

		// Conditional render with early return.
		{
			Code: `
function App({ flag }) {
  if (flag) return <button />;
  return null;
}
`,
			Tsx:     true,
			Options: optsForbid("button"),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   "<button> is forbidden",
					Line:      3,
					Column:    21,
				},
			},
		},

		// `React.Fragment` as createElement first arg — MemberExpression,
		// source text `React.Fragment` matches forbid: ['React.Fragment'].
		{
			Code:    `React.createElement(React.Fragment, null)`,
			Tsx:     true,
			Options: optsForbid("React.Fragment"),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   "<React.Fragment> is forbidden",
					Line:      1,
					Column:    21,
				},
			},
		},

		// Real-world: button with attributes including spread and
		// expression children.
		{
			Code:    `<button type="submit" disabled={isLoading} {...rest}>{label}</button>`,
			Tsx:     true,
			Options: optsForbid("button"),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   "<button> is forbidden",
					Line:      1,
					Column:    2,
				},
			},
		},

		// React.createElement nested in a useMemo-style hook callback.
		{
			Code: `
const Memoed = useMemo(() => React.createElement("button"), []);
`,
			Tsx:     true,
			Options: optsForbid("button"),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   "<button> is forbidden",
					Line:      2,
					Column:    50,
				},
			},
		},

		// JSX returned from a memo / forwardRef-shaped HOC call. The JSX
		// listener fires regardless of the surrounding wrapper.
		{
			Code: `
const X = forwardRef((props, ref) => <button ref={ref} {...props} />);
`,
			Tsx:     true,
			Options: optsForbid("button"),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   "<button> is forbidden",
					Line:      2,
					Column:    39,
				},
			},
		},

		// Adjacent JSX siblings inside a fragment — both fire in source
		// order.
		{
			Code: `
function Toolbar() {
  return (
    <>
      <button>Save</button>
      <button>Cancel</button>
    </>
  );
}
`,
			Tsx:     true,
			Options: optsForbid("button"),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   "<button> is forbidden",
					Line:      5,
					Column:    8,
				},
				{
					MessageId: "forbiddenElement",
					Message:   "<button> is forbidden",
					Line:      6,
					Column:    8,
				},
			},
		},

		// 4-segment JSX dotted tag (deep namespace).
		{
			Code:    `<a.b.c.d />`,
			Tsx:     true,
			Options: optsForbid("a.b.c.d"),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   "<a.b.c.d> is forbidden",
					Line:      1,
					Column:    2,
				},
			},
		},

		// CommonJS-style: `var React = require('react'); React.createElement('button')`.
		// Pragma access via Identifier `React` whose binding is unrelated
		// to the require — but our syntactic check only looks at the
		// AST shape, so this still matches `<pragma>.createElement`.
		{
			Code: `
var React = require('react');
React.createElement('button');
`,
			Tsx:     false,
			Options: optsForbid("button"),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   "<button> is forbidden",
					Line:      3,
					Column:    21,
				},
			},
		},

		// Default import — `import React from 'react'` — pragma identifier
		// is "React", which matches the default-imported binding's name.
		{
			Code: `
import React from 'react';
React.createElement('button');
`,
			Tsx:     true,
			Options: optsForbid("button"),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   "<button> is forbidden",
					Line:      3,
					Column:    21,
				},
			},
		},

		// Namespace import — `import * as React from 'react'`.
		{
			Code: `
import * as React from 'react';
React.createElement('button');
`,
			Tsx:     true,
			Options: optsForbid("button"),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   "<button> is forbidden",
					Line:      3,
					Column:    21,
				},
			},
		},

		// Aliased pragma — `import R from 'react'` with `pragma = "R"`.
		// The shape match looks at identifier text, so `R.createElement`
		// is recognized when pragma is set to "R".
		{
			Code: `
import R from 'react';
R.createElement('button');
`,
			Tsx:      true,
			Options:  optsForbid("button"),
			Settings: map[string]interface{}{"react": map[string]interface{}{"pragma": "R"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   "<button> is forbidden",
					Line:      3,
					Column:    17,
				},
			},
		},

		// JSX in switch/case body — JSX listener fires regardless of
		// surrounding statement.
		{
			Code: `
function Renderer({ kind }) {
  switch (kind) {
    case 'a': return <button />;
    case 'b': return <input />;
    default: return null;
  }
}
`,
			Tsx:     true,
			Options: optsForbid("button", "input"),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   "<button> is forbidden",
					Line:      4,
					Column:    23,
				},
				{
					MessageId: "forbiddenElement",
					Message:   "<input> is forbidden",
					Line:      5,
					Column:    23,
				},
			},
		},

		// JSX inside try/catch body.
		{
			Code: `
function Safe() {
  try {
    return <button />;
  } catch {
    return <input />;
  }
}
`,
			Tsx:     true,
			Options: optsForbid("button", "input"),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   "<button> is forbidden",
					Line:      4,
					Column:    13,
				},
				{
					MessageId: "forbiddenElement",
					Message:   "<input> is forbidden",
					Line:      6,
					Column:    13,
				},
			},
		},

		// Async function returning JSX.
		{
			Code: `
async function fetchAndRender() {
  await Promise.resolve();
  return <button />;
}
`,
			Tsx:     true,
			Options: optsForbid("button"),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   "<button> is forbidden",
					Line:      4,
					Column:    11,
				},
			},
		},

		// IIFE returning JSX.
		{
			Code:    `const x = (() => <button />)();`,
			Tsx:     true,
			Options: optsForbid("button"),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   "<button> is forbidden",
					Line:      1,
					Column:    19,
				},
			},
		},

		// JSX nested in an attribute expression.
		{
			Code:    `<Outer prop={<button />} />`,
			Tsx:     true,
			Options: optsForbid("button"),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   "<button> is forbidden",
					Line:      1,
					Column:    15,
				},
			},
		},

		// MemberExpression with PrivateIdentifier — `obj.#priv`.
		// `getText` returns `"obj.#priv"`. forbid match still works.
		{
			Code: `
class C {
  static #x = 1;
  static run() {
    React.createElement(C.#x);
  }
}
`,
			Tsx:     true,
			Options: optsForbid("C.#x"),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenElement",
					Message:   "<C.#x> is forbidden",
					Line:      5,
					Column:    25,
				},
			},
		},
	})
}

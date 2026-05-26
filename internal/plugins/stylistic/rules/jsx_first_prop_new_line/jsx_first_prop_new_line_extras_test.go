// TestJsxFirstPropNewLineExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row / tsgo AST quirk it covers,
// so future refactors can't silently regress them without breaking a named
// lock-in.
//
// Dimension 4 walk (rule inspects JSX opening elements + their first attribute):
//   - Receiver/expression wrappers (paren / non-null / as / optional chain):
//     N/A — the rule targets JSX elements, it never reads a member-access
//     receiver, so there is no child expression to wrap.
//   - Access/key forms: the prop forms (identifier attr / boolean shorthand /
//     spread / member-or-generic tag name) are covered below; computed keys do
//     not exist on JSX attributes.
//   - Declaration/container forms (function / class variants): N/A — the rule
//     does not target functions or classes.
//   - Nesting/traversal boundaries: covered (nested JSX, opening-vs-self-closing,
//     opening tag single-line while element spans lines via children).
//   - Graceful degradation: covered (no attributes, fragment container).
package jsx_first_prop_new_line

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestJsxFirstPropNewLineExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxFirstPropNewLineRule, []rule_tester.ValidTestCase{
		// Locks in 'multiline' arm: a single-line opening tag is never checked,
		// even with a prop on the same line.
		{Code: `<Foo bar />`, Tsx: true, Options: []interface{}{"multiline"}},
		// Locks in 'multiline-multiprop' arm: a single prop is not checked even
		// when the tag spans multiple lines (the len>1 guard short-circuits).
		{Code: `
        <Foo bar={{
        }} />
      `, Tsx: true, Options: []interface{}{"multiline-multiprop"}},
		// Locks in 'multiprop' arm: a single prop on a single-line tag is fine.
		{Code: `<Foo bar />`, Tsx: true, Options: []interface{}{"multiprop"}},
		// ---- Dimension 4: opening tag is single-line but the element spans
		// lines via children — isMultiline reads the opening tag (not the whole
		// element), so multiline-multiprop does not fire here.
		{Code: `
        <Foo bar baz>
          x
        </Foo>
      `, Tsx: true, Options: []interface{}{"multiline-multiprop"}},
		// ---- Dimension 4: JSX fragment container has no attributes; the inner
		// single-line multi-prop element is valid under multiline-multiprop.
		{Code: `
        <>
          <Foo bar baz />
        </>
      `, Tsx: true, Options: []interface{}{"multiline-multiprop"}},
		// Default option (no options) == 'multiline-multiprop': single-line tag
		// with multiple props is valid (locks in the default-resolution path).
		{Code: `<Foo bar baz />`, Tsx: true},
		// Default option: explicit empty options array resolves to the same
		// default; a single-line multi-prop tag stays valid.
		{Code: `<Foo bar baz />`, Tsx: true, Options: []interface{}{}},
		// ---- Dimension 4 / nesting: JSX in an attribute value. The outer Foo
		// has one prop (not reported under multiprop) and the inner Baz has none,
		// so nothing is reported. Locks in that descending into attribute-value
		// JSX does not produce a spurious report on either element.
		{Code: `<Foo bar={<Baz />} />`, Tsx: true, Options: []interface{}{"multiprop"}},
	}, []rule_tester.InvalidTestCase{
		// Locks in 'always' arm: unconditional — a single prop on a single-line
		// tag is still reported (no multiline / multi-prop requirement). Asserts
		// the full report span (Line/Column/EndLine/EndColumn) for the prop node.
		{
			Code:    `<Foo bar baz />`,
			Tsx:     true,
			Options: []interface{}{"always"},
			Output:  []string{"<Foo\nbar baz />"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propOnNewLine", Message: msgPropOnNewLine, Line: 1, Column: 6, EndLine: 1, EndColumn: 9},
			},
		},
		// Locks in 'multiprop' propOnNewLine arm: >1 prop on a single-line tag is
		// reported even though the tag is NOT multiline (distinguishes multiprop
		// from multiline-multiprop).
		{
			Code:    `<Foo aaa bbb />`,
			Tsx:     true,
			Options: []interface{}{"multiprop"},
			Output:  []string{"<Foo\naaa bbb />"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propOnNewLine", Message: msgPropOnNewLine, Line: 1, Column: 6},
			},
		},
		// Locks in 'never' arm with a single prop (len>0, not >1): a lone prop on
		// a new line is pulled back. Multi-line code; asserts full report span.
		{
			Code: `<Foo
bar />`,
			Tsx:     true,
			Options: []interface{}{"never"},
			Output:  []string{"<Foo bar />"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propOnSameLine", Message: msgPropOnSameLine, Line: 2, Column: 1, EndLine: 2, EndColumn: 4},
			},
		},
		// ---- Dimension 4: JsxOpeningElement with children (not self-closing) —
		// the listener handles both element kinds.
		{
			Code:    `<Foo bar>x</Foo>`,
			Tsx:     true,
			Options: []interface{}{"always"},
			Output:  []string{"<Foo\nbar>x</Foo>"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propOnNewLine", Message: msgPropOnNewLine, Line: 1, Column: 6},
			},
		},
		// ---- Dimension 4: member-expression tag name (`<Foo.Bar>`).
		// ---- Real-user: namespaced/dotted component is a common React shape.
		{
			Code:    `<Foo.Bar baz qux />`,
			Tsx:     true,
			Options: []interface{}{"multiprop"},
			Output:  []string{"<Foo.Bar\nbaz qux />"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propOnNewLine", Message: msgPropOnNewLine, Line: 1, Column: 10},
			},
		},
		// ---- Dimension 3 (autofix boundary): a comment sits between the tag name
		// and the first prop. rslint's fix only replaces the whitespace directly
		// before the prop, so the comment is preserved. NOTE: this diverges from
		// ESLint, which replaces the whole name→prop range and drops the comment.
		{
			Code:    `<Foo /* c */ bar />`,
			Tsx:     true,
			Options: []interface{}{"always"},
			Output:  []string{"<Foo /* c */\nbar />"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propOnNewLine", Message: msgPropOnNewLine, Line: 1, Column: 14},
			},
		},
		// ---- Dimension 3 (autofix boundary): multiple spaces before the prop are
		// all collapsed into the inserted newline.
		{
			Code:    `<Foo    bar />`,
			Tsx:     true,
			Options: []interface{}{"always"},
			Output:  []string{"<Foo\nbar />"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propOnNewLine", Message: msgPropOnNewLine, Line: 1, Column: 9},
			},
		},
		// ---- Nesting: each element is judged independently. Only the outer
		// element (>1 prop, same line) is reported; the inner childless element
		// is skipped — the listener does not bleed across the boundary.
		{
			Code:    `<Parent prop1 prop2><Child /></Parent>`,
			Tsx:     true,
			Options: []interface{}{"multiprop"},
			Output:  []string{"<Parent\nprop1 prop2><Child /></Parent>"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propOnNewLine", Message: msgPropOnNewLine, Line: 1, Column: 9},
			},
		},
		// ---- Real-user: TypeScript generic component under multiline-multiprop —
		// the fix anchors after the type arguments (`<T>`), not after the bare name.
		{
			Code: `<Box<T> aaa bbb
/>`,
			Tsx:     true,
			Options: []interface{}{"multiline-multiprop"},
			Output: []string{`<Box<T>
aaa bbb
/>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propOnNewLine", Message: msgPropOnNewLine, Line: 1, Column: 9},
			},
		},
		// ---- Dimension 3 (autofix boundary): generic component, propOnSameLine
		// fix keeps the type arguments (`<T>`). NOTE: this diverges from ESLint,
		// whose propOnSameLine fix replaces from the bare tag name and drops the
		// type arguments. rslint's fix only consumes whitespace adjacent to the
		// prop, so `<T>` survives.
		{
			Code: `<Foo<T>
bar />`,
			Tsx:     true,
			Options: []interface{}{"never"},
			Output:  []string{"<Foo<T> bar />"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propOnSameLine", Message: msgPropOnSameLine, Line: 2, Column: 1},
			},
		},
		// ---- Dimension 3 (autofix boundary): a comment between the tag name and
		// the first prop blocks the same-line pull-back. Collapsing only the
		// adjacent whitespace would leave the prop below the tag and the autofix
		// would never converge, so the violation is reported WITHOUT a fix (no
		// Output). NOTE: ESLint instead removes the comment and collapses the line.
		{
			Code: `<Foo
/* c */
bar />`,
			Tsx:     true,
			Options: []interface{}{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propOnSameLine", Message: msgPropOnSameLine, Line: 3, Column: 1},
			},
		},
		// ---- Dimension 4: spread attribute as the first prop — props[0] is a
		// JsxSpreadAttribute (not a JsxAttribute); it is still reported/anchored.
		{
			Code:    `<Foo {...x} bar />`,
			Tsx:     true,
			Options: []interface{}{"always"},
			Output: []string{`<Foo
{...x} bar />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propOnNewLine", Message: msgPropOnNewLine, Line: 1, Column: 6},
			},
		},
		// ---- Dimension 4 / nesting: JSX element inside an attribute value. The
		// outer Foo has a single prop (not reported under multiprop); only the
		// inner Baz (>1 prop, same line) is reported — the listener descends but
		// keeps the two elements' judgments independent.
		{
			Code:    `<Foo bar={<Baz a b />} />`,
			Tsx:     true,
			Options: []interface{}{"multiprop"},
			Output: []string{`<Foo bar={<Baz
a b />} />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propOnNewLine", Message: msgPropOnNewLine, Line: 1, Column: 16},
			},
		},
		// ---- Dimension 4: multibyte characters before the element. The column is
		// counted in UTF-16 code units (`变量` = 2 units), matching ESLint — not
		// byte offset.
		{
			Code:    `const 变量 = <Foo aaa bbb />`,
			Tsx:     true,
			Options: []interface{}{"multiprop"},
			Output: []string{`const 变量 = <Foo
aaa bbb />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propOnNewLine", Message: msgPropOnNewLine, Line: 1, Column: 17},
			},
		},
		// ---- Dimension 4: CRLF line endings — the same-line fix consumes the
		// `\r\n` pair and line numbers are computed correctly across it.
		{
			Code:    "<Foo\r\na />",
			Tsx:     true,
			Options: []interface{}{"never"},
			Output:  []string{"<Foo a />"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propOnSameLine", Message: msgPropOnSameLine, Line: 2, Column: 1},
			},
		},
		// ---- Real-user: lowercase HTML element with multiple attributes.
		{
			Code:    `<div className="x" id="y" />`,
			Tsx:     true,
			Options: []interface{}{"multiprop"},
			Output: []string{`<div
className="x" id="y" />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propOnNewLine", Message: msgPropOnNewLine, Line: 1, Column: 6},
			},
		},
		// ---- Real-user: JSX returned from an Array.map callback.
		{
			Code:    `items.map(i => <Item k={i} a b />)`,
			Tsx:     true,
			Options: []interface{}{"multiprop"},
			Output: []string{`items.map(i => <Item
k={i} a b />)`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propOnNewLine", Message: msgPropOnNewLine, Line: 1, Column: 22},
			},
		},
	})
}

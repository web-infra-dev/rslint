// TestJsxIndentPropsExtras locks in branches and edge shapes that the upstream
// test suite doesn't exercise. Each case carries an inline comment pointing at
// the specific branch / Dimension 4 row / tsgo AST quirk it covers, so future
// refactors can't silently regress them without breaking a named lock-in.
//
// The implementation is shared with react/jsx-indent-props via BuildRule, so
// these cases double as guards for both registered variants.
package jsx_indent_props

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestJsxIndentPropsExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxIndentPropsRule, []rule_tester.ValidTestCase{
		// ---- Dimension 4: access/key form — JsxSpreadAttribute as first prop
		// (default 4-space). `{...rest}` must obey the same indent contract as
		// a plain JsxAttribute. ----
		{Code: "\n        <App\n            {...rest}\n            foo\n        />\n      ", Tsx: true},

		// ---- Dimension 4: spread as first prop under 'first' — the column
		// anchor is the leading `{` of `{...rest}`, not any inner identifier. ----
		{Code: "\n        <App {...rest}\n             foo\n        />\n      ", Tsx: true, Options: []interface{}{"first"}},
		{Code: "\n        <App {...rest}\n             id=\"x\"\n             onClick={fn}\n        />\n      ", Tsx: true, Options: []interface{}{"first"}},

		// ---- Dimension 4: boolean attribute as first prop under 'first' —
		// anchor comes from the boolean's identifier start. ----
		{Code: "\n        <App disabled\n             onClick={fn}\n             id=\"x\"\n        />\n      ", Tsx: true, Options: []interface{}{"first"}},

		// ---- Dimension 4: container forms — generic tag `<Foo<T>>` type
		// arguments don't appear in Attributes.Properties.Nodes, so prop
		// processing ignores them. ----
		{Code: "\n        <Foo<string>\n            a\n            b\n        />\n      ", Tsx: true},

		// ---- Dimension 4: member-expression tag `<Foo.Bar>` — the dotted name
		// must be transparent to prop processing. ----
		{Code: "\n        <Foo.Bar\n            a\n            b\n        />\n      ", Tsx: true},

		// ---- Dimension 4: namespaced tag `<svg:rect>` — same shape as member,
		// transparent to prop processing. ----
		{Code: "\n        <svg:rect\n            a\n            b\n        />\n      ", Tsx: true},

		// ---- Dimension 4: graceful degradation — single-line JSX inside a
		// ternary must NOT trigger the bump (the `?` sits on the same line as
		// `<` but the line does not START with `?`/`:` after whitespace). ----
		{Code: "\n        const x = cond ? <App foo bar /> : null\n      ", Tsx: true, Options: []interface{}{float64(2)}},

		// ---- Dimension 4: nesting boundary — adjacent ternary JSX + non-ternary
		// JSX. The per-element state reset must keep `isUsingOperator` from
		// leaking from the first element to the second. ----
		{Code: "\n        const a = c1\n          ? <X\n              a\n            /> : null;\n        const b = <App\n          foo\n        />;\n      ", Tsx: true, Options: []interface{}{float64(2)}},

		// ---- Dimension 4: multi-line JSX literal as an attr value — first line
		// of `foo={` does NOT contain `<`, so upstream's `useBracket` reset does
		// NOT fire and the ternary bump still applies to subsequent props. ----
		{Code: "\n        c\n          ? <App\n              foo={\n                <Inner/>\n              }\n              bar\n            />\n          : null\n      ", Tsx: true, Options: []interface{}{float64(2)}},

		// ---- Dimension 4: single-line attr value containing a JSX literal —
		// `foo={<Inner/>}`'s first line contains `<`, which cancels the ternary
		// bump (upstream's `useBracket` reset). Both props align WITHOUT it. ----
		{Code: "\n        c\n          ? <App\n            foo={<Inner/>}\n            bar\n          />\n          : null\n      ", Tsx: true, Options: []interface{}{float64(2)}},

		// ---- Dimension 4: TS type-expression wrapper around the ternary side
		// (`as`). The `useOperator` check is line-content-based, so the wrapper
		// doesn't move which line `?` sits on and the bump still applies. ----
		{Code: "\n        const x = c\n          ? (<App\n              foo\n            />) as React.ReactElement\n          : null\n      ", Tsx: true, Options: []interface{}{float64(2)}},

		// N/A: optional-chain / non-null receiver wrappers — the rule inspects
		// JSX attribute positions, never an arbitrary member/call receiver, so
		// `?.` / `!` receiver shapes don't reach this rule's child accesses.

		// ---- Real-user: inline arrow event handler with a multi-line body. The
		// first source line of `onClick` ends with `={() => {` (no `<`), so the
		// bracket-reset is a no-op and there is no ternary. ----
		{Code: "\n        <button\n          onClick={() => {\n            setX(true);\n            log();\n          }}\n          disabled\n        />\n      ", Tsx: true, Options: []interface{}{float64(2)}},

		// ---- Real-user: className template literal whose interpolation itself
		// contains a `?:` — the `?` is inside `${ ... }`, not at line start, so
		// it must NOT trigger the operator bump. ----
		{Code: "\n        <div\n          className={`foo ${cond ? 'a' : 'b'} bar`}\n          id=\"x\"\n        />\n      ", Tsx: true, Options: []interface{}{float64(2)}},

		// ---- Real-user: spread + named override (`<Input {...props} value={x} />`). ----
		{Code: "\n        <Input\n          {...props}\n          value={localValue}\n        />\n      ", Tsx: true, Options: []interface{}{float64(2)}},

		// ---- Real-user: multiple consecutive spread attributes followed by named. ----
		{Code: "\n        <Input\n          {...props1}\n          {...props2}\n          value={x}\n        />\n      ", Tsx: true, Options: []interface{}{float64(2)}},

		// ---- Real-user: render-prop / children-as-function — JSX nested inside
		// a JSX expression container that itself sits in JSX children. ----
		{Code: "\n        <Query>\n          {(data) => <Result\n            value={data}\n            onClick={handler}\n          />}\n        </Query>\n      ", Tsx: true, Options: []interface{}{float64(2)}},

		// ---- Real-user: conditional rendering with `&&` wrapping multi-line JSX.
		// The line of `<Modal` is not led by `?`/`:`, so no bump. ----
		{Code: "\n        {isOpen && (\n          <Modal\n            onClose={close}\n            title=\"Welcome\"\n          />\n        )}\n      ", Tsx: true, Options: []interface{}{float64(2)}},

		// ---- Real-user: `.map(...)` render with a multi-line JSX body. ----
		{Code: "\n        {items.map((item) => (\n          <Item\n            key={item.id}\n            value={item.value}\n          />\n        ))}\n      ", Tsx: true, Options: []interface{}{float64(2)}},

		// ---- Real-user: hyphenated and dotted attribute names exercising the
		// JSX identifier parser; `ref` is just an attribute. ----
		{Code: "\n        <input\n          ref={inputRef}\n          type=\"text\"\n          aria-label=\"name\"\n          data-testid=\"input\"\n        />\n      ", Tsx: true, Options: []interface{}{float64(2)}},

		// ---- Real-user: full-form component — function declaration, return,
		// multi-line opening tag, hyphenated and template-literal props. ----
		{Code: "\n        function Form({ submit }) {\n          return (\n            <form\n              onSubmit={submit}\n              noValidate\n              autoComplete=\"off\"\n              className={`form ${variant}`}\n            >\n              <input\n                ref={inputRef}\n                type=\"text\"\n              />\n            </form>\n          );\n        }\n      ", Tsx: true, Options: []interface{}{float64(2)}},

		// ---- Real-user: comment between attributes — leading trivia on `b`;
		// the trimmed start must skip the comment so the indent reading lands on
		// `b`'s line, not the comment's. ----
		{Code: "\n        <App\n          a={1}\n          // separator\n          b={2}\n        />\n      ", Tsx: true, Options: []interface{}{float64(2)}},

		// ---- Branch lock-in: parseOptions object form with ONLY
		// `ignoreTernaryOperator` (no indentMode) falls back to default 4-space,
		// and the flag still suppresses the bump. ----
		{Code: "\n        {cond\n          ? <App\n              foo\n            />\n          : null}\n      ", Tsx: true, Options: []interface{}{map[string]interface{}{"ignoreTernaryOperator": true}}},

		// ---- Branch lock-in: parseOptions object form with `indentMode: 'first'`
		// — string value inside the object form behaves as the bare `'first'`. ----
		{Code: "\n        <App aaa\n             b\n             cc\n        />\n      ", Tsx: true, Options: []interface{}{map[string]interface{}{"indentMode": "first"}}},

		// ---- Branch lock-in: large indent (8-space) — option size only feeds
		// arithmetic and the string repeat. ----
		{Code: "\n        <App\n                foo\n                bar\n        />\n      ", Tsx: true, Options: []interface{}{float64(8)}},

		// ---- Branch lock-in: empty options array — equivalent to default 4-space. ----
		{Code: "\n        <App\n            foo\n        />\n      ", Tsx: true, Options: []interface{}{}},

		// ---- Branch lock-in: unrecognised string in object form
		// (`indentMode: 'spaces'`) silently falls through to default 4-space. ----
		{Code: "\n        <App\n            foo\n        />\n      ", Tsx: true, Options: []interface{}{map[string]interface{}{"indentMode": "spaces"}}},

		// ---- Branch lock-in: ternary bump when the FIRST prop shares the
		// opening tag's `?`/`:` line. Upstream's getNodeIndent gives useOperator
		// (line starts with the operator) priority over useBracket (line
		// contains `<`), so the operator state survives the `<` on that line and
		// the bump carries to the new-line props. Confirmed valid via ESLint
		// differential against @stylistic 5.10.0. (Regression guard: an earlier
		// port reset the state on `<` unconditionally and reported false positives here.) ----
		{Code: "\n        const x = cond\n          ? <App foo\n              bar\n            />\n          : null\n      ", Tsx: true, Options: []interface{}{float64(2)}},
		// `:` alternate is symmetric.
		{Code: "\n        const x = cond\n          ? null\n          : <App foo\n              bar\n            />\n      ", Tsx: true, Options: []interface{}{float64(2)}},
		// Multiple new-line props after the operator-line first prop — all bumped.
		{Code: "\n        const y = cond\n          ? <App foo\n              bar\n              baz\n            />\n          : null\n      ", Tsx: true, Options: []interface{}{float64(2)}},

		// N/A: ArrayExpression / ObjectExpression node-type skip — upstream's
		// `node.type !== 'ArrayExpression' && node.type !== 'ObjectExpression'`
		// guard is vestigial in the JSX context: JSX attributes are always
		// JsxAttribute / JsxSpreadAttribute, never those kinds, so the guard is
		// unreachable here (inherited from the generic indent rule it forked).
	}, []rule_tester.InvalidTestCase{
		// ---- Dimension 4: spread + named prop both at wrong indent — both must
		// report. ----
		{
			Code:   "\n        <App\n          {...rest}\n          foo\n        />\n      ",
			Output: []string{"\n        <App\n            {...rest}\n            foo\n        />\n      "},
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent"},
				{MessageId: "wrongIndent"},
			},
		},

		// ---- Branch lock-in: bracket-reset cancels the ternary bump. The props
		// are indented as if the bump applied, but the first prop's first source
		// line contains `<` (`foo={<Inner/>}`), so upstream's `useBracket` reset
		// fires and the expected indent is propIndent (NOT propIndent+indentSize).
		// Without the reset this would silently accept props at indent 10. ----
		{
			Code:    "\n        c\n          ? <App\n              foo={<Inner/>}\n              bar\n            />\n          : null\n      ",
			Output:  []string{"\n        c\n          ? <App\n            foo={<Inner/>}\n            bar\n            />\n          : null\n      "},
			Tsx:     true,
			Options: []interface{}{float64(2)},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent"},
				{MessageId: "wrongIndent"},
			},
		},

		// ---- Branch lock-in: ternary state must NOT leak to the next JSX. The
		// first element's `?` line sets `isUsingOperator`, but the second
		// element's own `<` line isn't led by `?`/`:` — it must fall back to the
		// non-bumped expected indent. ----
		{
			Code:    "\n        const a = c1 ? <X\n          a\n        /> : null;\n        <Y\n            foo\n        />\n      ",
			Output:  []string{"\n        const a = c1 ? <X\n          a\n        /> : null;\n        <Y\n          foo\n        />\n      "},
			Tsx:     true,
			Options: []interface{}{float64(2)},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent"},
			},
		},

		// ---- Diagnostic contract: default 4-space, full diagnostic shape —
		// messageId, exact message text, and 1-based Line/Column/EndLine/
		// EndColumn anchored at the JsxAttribute's trimmed range. ----
		{
			Code:   "\n        <App\n          foo\n        />\n      ",
			Output: []string{"\n        <App\n            foo\n        />\n      "},
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 12 space characters but found 10.", Line: 3, Column: 11, EndLine: 3, EndColumn: 14},
			},
		},

		// ---- Diagnostic contract: tab option, zero leading tabs — needed=1 must
		// use the SINGULAR "character"; locks in the pluralization branch. ----
		{
			Code:    "\n        <App1\n            foo\n        />\n      ",
			Output:  []string{"\n        <App1\n\tfoo\n        />\n      "},
			Tsx:     true,
			Options: []interface{}{"tab"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 1 tab character but found 0.", Line: 3, Column: 13, EndLine: 3, EndColumn: 16},
			},
		},

		// ---- Diagnostic contract: 'first' alignment mismatch — expected indent
		// is the first prop's column (`a` at column 13). ----
		{
			Code:    "\n        <App a\n          b\n        />\n      ",
			Output:  []string{"\n        <App a\n             b\n        />\n      "},
			Tsx:     true,
			Options: []interface{}{"first"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 13 space characters but found 10.", Line: 3, Column: 11, EndLine: 3, EndColumn: 12},
			},
		},

		// ---- Diagnostic contract: multi-error sequencing — two mis-indented
		// props on different lines each report on their OWN line. ----
		{
			Code:   "\n        <App\n          foo\n          bar\n        />\n      ",
			Output: []string{"\n        <App\n            foo\n            bar\n        />\n      "},
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 12 space characters but found 10.", Line: 3, Column: 11},
				{MessageId: "wrongIndent", Message: "Expected indentation of 12 space characters but found 10.", Line: 4, Column: 11},
			},
		},

		// ---- Branch lock-in: 'first' with three mismatched props — Line
		// increments stay in source order (iteration follows
		// Attributes.Properties.Nodes order). ----
		{
			Code:    "\n        <App\n          a\n         b\n           c\n        />\n      ",
			Output:  []string{"\n        <App\n          a\n          b\n          c\n        />\n      "},
			Tsx:     true,
			Options: []interface{}{"first"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Line: 4, Column: 10},
				{MessageId: "wrongIndent", Line: 5, Column: 12},
			},
		},

		// ---- Branch lock-in: negative indent option `[-2]` makes propIndent < 0.
		// (1) The rule MUST NOT panic on the negative repeat count (omitted
		// Output asserts no fix was applied — matching ESLint, whose repeat()
		// throws and is silently dropped). (2) The message preserves the
		// negative `needed` verbatim. ----
		{
			Code:    "<App\nfoo\n/>",
			Tsx:     true,
			Options: []interface{}{float64(-2)},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of -2 space characters but found 0.", Line: 2, Column: 1, EndLine: 2, EndColumn: 4},
			},
		},

		// ---- Dimension 4: multi-byte tag name in 'first' mode — `<中Foo>` adds
		// 1 UTF-16 code unit but 3 UTF-8 bytes. propIndent must use the UTF-16
		// character column (matches ESLint's loc.start.column), not byte offset;
		// otherwise this would expect indent 16 instead of 14. ----
		{
			Code:    "\n        <中Foo a=\"x\"\n           b=\"y\"\n        />\n      ",
			Output:  []string{"\n        <中Foo a=\"x\"\n              b=\"y\"\n        />\n      "},
			Tsx:     true,
			Options: []interface{}{"first"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 14 space characters but found 11.", Line: 3, Column: 12, EndLine: 3, EndColumn: 17},
			},
		},

		// ---- Diagnostic contract: multi-line attribute value — the `beforeNav`
		// attribute spans 4 lines; the diagnostic's end-of-range must land on the
		// CLOSING line/column of the JsxAttribute, not on its first line. ----
		{
			Code:   "\n        <BaseLayout\n          beforeNav={\n            <Banner />\n          }\n        />\n      ",
			Output: []string{"\n        <BaseLayout\n            beforeNav={\n            <Banner />\n          }\n        />\n      "},
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 12 space characters but found 10.", Line: 3, Column: 11, EndLine: 5, EndColumn: 12},
			},
		},

		// ---- Branch lock-in: same operator-line-first-prop shape, but the
		// new-line prop sits at the UN-bumped column (12). Must report needs 14
		// (propIndent 12 + bump 2), proving the ternary bump IS applied — the
		// direct regression guard for useOperator preceding useBracket. ----
		{
			Code:    "\n        const x = cond\n          ? <App foo\n            bar\n            />\n          : null\n      ",
			Output:  []string{"\n        const x = cond\n          ? <App foo\n              bar\n            />\n          : null\n      "},
			Tsx:     true,
			Options: []interface{}{float64(2)},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 14 space characters but found 12.", Line: 4, Column: 13},
			},
		},
	})
}

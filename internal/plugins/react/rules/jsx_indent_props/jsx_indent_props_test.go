package jsx_indent_props

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestJsxIndentPropsRule mirrors the upstream eslint-plugin-react test suite
// at `tests/lib/rules/jsx-indent-props.js`. Each upstream case is migrated
// 1:1; cases that exercise non-tsgo features are kept as `Skip: true` with a
// short reason so the file still maps to the upstream layout.
//
// Layout note: upstream's tagged-template-literal cases anchor their JSX
// with an 8-space outer indent that the template literal itself injects
// (the source of the test file is indented inside the JS module). We keep
// the same 8-space prefix verbatim so the indent arithmetic continues to
// match upstream's expectations.
func TestJsxIndentPropsRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxIndentPropsRule, []rule_tester.ValidTestCase{
		// ---- Default 4-space indent (no options) ----
		{
			Code: "\n        <App foo\n        />\n      ",
			Tsx:  true,
		},

		// ---- 2-space indent ----
		{
			Code:    "\n        <App\n          foo\n        />\n      ",
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},

		// ---- Complex multi-element array literal (options: [2]) ----
		{
			Code:    "\n        const Test = () => ([\n          (x\n            ? <div key=\"1\" />\n            : <div key=\"2\" />),\n          <div\n            key=\"3\"\n            align=\"left\"\n          />,\n          <div\n            key=\"4\"\n            align=\"left\"\n          />,\n        ]);\n      ",
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},

		// ---- 0-indent (props flush to column 0) ----
		{
			Code:    "\n        <App\n        foo\n        />\n      ",
			Tsx:     true,
			Options: []interface{}{float64(0)},
		},

		// ---- Negative indent (props sit two cols to the LEFT of `<App`) ----
		{
			Code:    "\n          <App\n        foo\n          />\n      ",
			Tsx:     true,
			Options: []interface{}{float64(-2)},
		},

		// ---- Tab indentation ----
		{
			Code:    "\n\t\t\t\t<App\n\t\t\t\t\tfoo\n\t\t\t\t/>\n\t\t\t",
			Tsx:     true,
			Options: []interface{}{"tab"},
		},

		// ---- 'first' option, no props ----
		{
			Code:    "\n        <App/>\n      ",
			Tsx:     true,
			Options: []interface{}{"first"},
		},

		// ---- 'first' option, props aligned with first prop's column ----
		{
			Code:    "\n        <App aaa\n             b\n             cc\n        />\n      ",
			Tsx:     true,
			Options: []interface{}{"first"},
		},
		{
			Code:    "\n        <App   aaa\n               b\n               cc\n        />\n      ",
			Tsx:     true,
			Options: []interface{}{"first"},
		},
		{
			Code:    "\n        const test = <App aaa\n                          b\n                          cc\n                     />\n      ",
			Tsx:     true,
			Options: []interface{}{"first"},
		},
		{
			Code:    "\n        <App aaa x\n             b y\n             cc\n        />\n      ",
			Tsx:     true,
			Options: []interface{}{"first"},
		},
		{
			Code:    "\n        const test = <App aaa x\n                          b y\n                          cc\n                     />\n      ",
			Tsx:     true,
			Options: []interface{}{"first"},
		},
		{
			Code:    "\n        <App aaa\n             b\n        >\n            <Child c\n                   d/>\n        </App>\n      ",
			Tsx:     true,
			Options: []interface{}{"first"},
		},
		{
			Code:    "\n        <Fragment>\n          <App aaa\n               b\n               cc\n          />\n          <OtherApp a\n                    bbb\n                    c\n          />\n        </Fragment>\n      ",
			Tsx:     true,
			Options: []interface{}{"first"},
		},

		// ---- 'first' with all props on new lines: each prop must match
		// the first prop's column (which is the open-tag's own line indent
		// plus the gap before the first prop). ----
		{
			Code:    "\n        <App\n          a\n          b\n        />\n      ",
			Tsx:     true,
			Options: []interface{}{"first"},
		},

		// ---- ignoreTernaryOperator: false (default), input already has
		// the bump applied. ----
		{
			Code:    "\n        {this.props.ignoreTernaryOperatorFalse\n          ? <span\n              className=\"value\"\n              some={{aaa}}\n            />\n          : null}\n      ",
			Tsx:     true,
			Options: []interface{}{map[string]interface{}{"indentMode": float64(2), "ignoreTernaryOperator": false}},
		},

		// ---- Function returning JSX nested in conditional, default
		// indentMode=2, ignoreTernaryOperator=false. ----
		{
			Code:    "\n        const F = () => {\n          const foo = true\n            ? <div id=\"id\">test</div>\n            : false;\n\n          return <div\n            id=\"id\"\n          >\n            test\n          </div>\n        }\n      ",
			Tsx:     true,
			Options: []interface{}{map[string]interface{}{"indentMode": float64(2), "ignoreTernaryOperator": false}},
		},

		// ---- Same shape, ignoreTernaryOperator=true — single-line JSX
		// inside the conditional doesn't trigger a bump either way. ----
		{
			Code:    "\n        const F = () => {\n          const foo = true\n            ? <div id=\"id\">test</div>\n            : false;\n\n          return <div\n            id=\"id\"\n          >\n            test\n          </div>\n        }\n      ",
			Tsx:     true,
			Options: []interface{}{map[string]interface{}{"indentMode": float64(2), "ignoreTernaryOperator": true}},
		},

		// ---- Tab indent with conditional + return JSX ----
		{
			Code:    "\n\t\t\t\tconst F = () => {\n\t\t\t\t\tconst foo = true\n\t\t\t\t\t\t? <div id=\"id\">test</div>\n\t\t\t\t\t\t: false;\n\n\t\t\t\t\treturn <div\n\t\t\t\t\t\tid=\"id\"\n\t\t\t\t\t>\n\t\t\t\t\t\ttest\n\t\t\t\t\t</div>\n\t\t\t\t}\n",
			Tsx:     true,
			Options: []interface{}{map[string]interface{}{"indentMode": "tab", "ignoreTernaryOperator": false}},
		},
		{
			Code:    "\n\t\t\t\tconst F = () => {\n\t\t\t\t\tconst foo = true\n\t\t\t\t\t\t? <div id=\"id\">test</div>\n\t\t\t\t\t\t: false;\n\n\t\t\t\t\treturn <div\n\t\t\t\t\t\tid=\"id\"\n\t\t\t\t\t>\n\t\t\t\t\t\ttest\n\t\t\t\t\t</div>\n\t\t\t\t}\n",
			Tsx:     true,
			Options: []interface{}{map[string]interface{}{"indentMode": "tab", "ignoreTernaryOperator": true}},
		},

		// ---- ignoreTernaryOperator: true — props inside ternary alternate/
		// consequent do NOT receive an extra indent bump. ----
		{
			Code:    "\n        {this.props.ignoreTernaryOperatorTrue\n          ? <span\n            className=\"value\"\n            some={{aaa}}\n            />\n          : null}\n      ",
			Tsx:     true,
			Options: []interface{}{map[string]interface{}{"indentMode": float64(2), "ignoreTernaryOperator": true}},
		},

		// ---- Realistic anchor element with mixed prop kinds (object form
		// with only indentMode). ----
		{
			Code:    "\n        <a\n          role={'button'}\n          className={`navbar-burger ${open ? 'is-active' : ''}`}\n          href={'#'}\n          aria-label={'menu'}\n          aria-expanded={false}\n          onClick={openMenu}>\n          <span aria-hidden={'true'}/>\n          <span aria-hidden={'true'}/>\n          <span aria-hidden={'true'}/>\n        </a>\n      ",
			Tsx:     true,
			Options: []interface{}{map[string]interface{}{"indentMode": float64(2)}},
		},

		// ---- Same realistic element under 'first' alignment ----
		{
			Code:    "\n        <a role={'button'}\n           className={`navbar-burger ${open ? 'is-active' : ''}`}\n           href={'#'}\n           aria-label={'menu'}\n           aria-expanded={false}\n           onClick={openMenu}>\n          <span aria-hidden={'true'}/>\n          <span aria-hidden={'true'}/>\n          <span aria-hidden={'true'}/>\n        </a>\n      ",
			Tsx:     true,
			Options: []interface{}{"first"},
		},

		// ---- tsgo-specific lock-ins (NOT in upstream) ----
		// Spread attribute as the first prop under default 4-space:
		// `{...rest}` is a JsxSpreadAttribute and must obey the same
		// indent contract as a JsxAttribute.
		{
			Code: "\n        <App\n            {...rest}\n            foo\n        />\n      ",
			Tsx:  true,
		},

		// Spread attribute as first prop with 'first' alignment — column
		// of the first prop is the column of the leading `{`, and the
		// second prop must match.
		{
			Code:    "\n        <App {...rest}\n             foo\n        />\n      ",
			Tsx:     true,
			Options: []interface{}{"first"},
		},

		// Single-line JSX inside a ternary must NOT trigger the bump —
		// upstream's `useOperator` regex only fires when the `<` line
		// itself starts with `?`/`:` after WS, not when the `?` sits
		// upstream on the same line.
		{
			Code:    "\n        const x = cond ? <App foo bar /> : null\n      ",
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},

		// Adjacent ternary JSX + non-ternary JSX in the same source —
		// the per-element state reset must prevent `isUsingOperator`
		// from leaking across elements (the state was set by the first
		// element's `?` line, but the second element's `<App` line is
		// not a ternary side and must NOT pick up the bump).
		{
			Code:    "\n        const a = c1\n          ? <X\n              a\n            /> : null;\n        const b = <App\n          foo\n        />;\n      ",
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},

		// Multi-line JSX literal as an attr value, first line of the
		// prop does NOT contain `<` (it ends with `={`), so upstream's
		// `useBracket` reset does NOT fire and the ternary bump still
		// applies to subsequent props.
		{
			Code:    "\n        c\n          ? <App\n              foo={\n                <Inner/>\n              }\n              bar\n            />\n          : null\n      ",
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},

		// Single-line attr value containing a JSX literal: first line
		// of the prop contains `<`, which cancels the ternary bump
		// (upstream's `useBracket` reset). Both `foo={...}` and the
		// following `bar` must align WITHOUT the bump.
		{
			Code:    "\n        c\n          ? <App\n            foo={<Inner/>}\n            bar\n          />\n          : null\n      ",
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},

		// TS wrappers around the ternary side (`as`, `satisfies`, `!`,
		// type assertion) are explicit nodes in tsgo — but the
		// `useOperator` check is line-content-based, so wrappers don't
		// change which line `?` sits on, and the bump still applies.
		{
			Code:    "\n        const x = c\n          ? (<App\n              foo\n            />) as React.ReactElement\n          : null\n      ",
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},

		// Generic JSX tag `<Foo<T>>` — type arguments don't appear in
		// `Attributes.Properties.Nodes`, so prop processing must ignore
		// them and treat the multi-line props normally.
		{
			Code:    "\n        <Foo<string>\n            a\n            b\n        />\n      ",
			Tsx:     true,
		},

		// Member-expression tag `<Foo.Bar>` — the dotted name must not
		// disturb prop processing.
		{
			Code: "\n        <Foo.Bar\n            a\n            b\n        />\n      ",
			Tsx:  true,
		},

		// Namespaced tag `<svg:rect>` — same shape as member, should be
		// transparent to prop processing.
		{
			Code: "\n        <svg:rect\n            a\n            b\n        />\n      ",
			Tsx:  true,
		},

		// ---- Real user scenarios (valid) ----
		// Inline arrow event handler with a multi-line body. The first
		// source line of `onClick` ends with `={() => {` — no `<` — so
		// the bracket-reset is a no-op (and there's no ternary anyway).
		{
			Code:    "\n        <button\n          onClick={() => {\n            setX(true);\n            log();\n          }}\n          disabled\n        />\n      ",
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},

		// className with template literal interpolation that itself
		// contains a `?:` — the `?` is inside `${ ... }`, not at line
		// start, so it must NOT trigger the operator bump.
		{
			Code:    "\n        <div\n          className={`foo ${cond ? 'a' : 'b'} bar`}\n          id=\"x\"\n        />\n      ",
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},

		// Spread + named override pattern (`<Input {...props} value={x} />`)
		// — both forms of attribute share the same indent contract.
		{
			Code:    "\n        <Input\n          {...props}\n          value={localValue}\n        />\n      ",
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},

		// Multiple consecutive spread attributes followed by named.
		{
			Code:    "\n        <Input\n          {...props1}\n          {...props2}\n          value={x}\n        />\n      ",
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},

		// Render-prop / children-as-function — JSX nested inside a JSX
		// expression container that itself sits in JSX children.
		{
			Code:    "\n        <Query>\n          {(data) => <Result\n            value={data}\n            onClick={handler}\n          />}\n        </Query>\n      ",
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},

		// Conditional rendering with `&&` (logical-and) wrapping a
		// multi-line JSX. Same as map / IIFE — line of `<Modal` is not
		// led by `?`/`:`, so no bump.
		{
			Code:    "\n        {isOpen && (\n          <Modal\n            onClose={close}\n            title=\"Welcome\"\n          />\n        )}\n      ",
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},

		// `.map(...)` render with a multi-line JSX body.
		{
			Code:    "\n        {items.map((item) => (\n          <Item\n            key={item.id}\n            value={item.value}\n          />\n        ))}\n      ",
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},

		// `ref` is just an attribute — no special handling. Combined
		// with hyphenated and dotted attribute names that exercise the
		// JSX identifier parser.
		{
			Code:    "\n        <input\n          ref={inputRef}\n          type=\"text\"\n          aria-label=\"name\"\n          data-testid=\"input\"\n        />\n      ",
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},

		// Realistic full-form component: function declaration, return
		// statement, multi-line opening tag, hyphenated and template-
		// literal-valued props.
		{
			Code:    "\n        function Form({ submit }) {\n          return (\n            <form\n              onSubmit={submit}\n              noValidate\n              autoComplete=\"off\"\n              className={`form ${variant}`}\n            >\n              <input\n                ref={inputRef}\n                type=\"text\"\n              />\n            </form>\n          );\n        }\n      ",
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},

		// Comment between attributes — leading trivia on the second
		// attr; trimmed.Pos() must skip the comment so the line-indent
		// reading lands on `b`'s line, not the comment's line.
		{
			Code:    "\n        <App\n          a={1}\n          // separator\n          b={2}\n        />\n      ",
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},

		// `'first'` mode with a SPREAD as the first prop — column
		// anchor must come from the leading `{` of `{...rest}`, not
		// from any inner identifier.
		{
			Code:    "\n        <App {...rest}\n             id=\"x\"\n             onClick={fn}\n        />\n      ",
			Tsx:     true,
			Options: []interface{}{"first"},
		},

		// `'first'` mode with a BOOLEAN attribute as the first prop —
		// column anchor must come from the boolean's identifier start.
		{
			Code:    "\n        <App disabled\n             onClick={fn}\n             id=\"x\"\n        />\n      ",
			Tsx:     true,
			Options: []interface{}{"first"},
		},

		// ---- Configuration edge cases ----
		// Object form with ONLY `ignoreTernaryOperator` (no
		// `indentMode`): should fall back to default 4-space, and the
		// flag must still take effect (no bump on ternary side).
		{
			Code:    "\n        {cond\n          ? <App\n              foo\n            />\n          : null}\n      ",
			Tsx:     true,
			Options: []interface{}{map[string]interface{}{"ignoreTernaryOperator": true}},
		},

		// Object form with `indentMode: 'first'` — string value inside
		// the object form must work the same as the bare `'first'`.
		{
			Code:    "\n        <App aaa\n             b\n             cc\n        />\n      ",
			Tsx:     true,
			Options: []interface{}{map[string]interface{}{"indentMode": "first"}},
		},

		// Large indent (8-space) — option size only feeds arithmetic
		// and string repeat, must work regardless of magnitude.
		{
			Code:    "\n        <App\n                foo\n                bar\n        />\n      ",
			Tsx:     true,
			Options: []interface{}{float64(8)},
		},

		// Empty options array — equivalent to default 4-space.
		{
			Code:    "\n        <App\n            foo\n        />\n      ",
			Tsx:     true,
			Options: []interface{}{},
		},

		// Unrecognised string in object form (`indentMode: 'spaces'`)
		// silently falls through to default 4-space — must not crash
		// or produce false reports.
		{
			Code:    "\n        <App\n            foo\n        />\n      ",
			Tsx:     true,
			Options: []interface{}{map[string]interface{}{"indentMode": "spaces"}},
		},
	}, []rule_tester.InvalidTestCase{
		// ---- Default 4-space indent — child at 10 must be 12. ----
		{
			Code:   "\n        <App\n          foo\n        />\n      ",
			Output: []string{"\n        <App\n            foo\n        />\n      "},
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent"},
			},
		},

		// ---- 2-space indent — child at 12 must be 10. ----
		{
			Code:    "\n        <App\n            foo\n        />\n      ",
			Output:  []string{"\n        <App\n          foo\n        />\n      "},
			Tsx:     true,
			Options: []interface{}{float64(2)},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent"},
			},
		},

		// ---- Conditional with JSX in both branches — bump applies on
		// each ternary side. ----
		{
			Code:    "\n        const test = true\n          ? <span\n            attr=\"value\"\n            />\n          : <span\n            attr=\"otherValue\"\n            />\n      ",
			Output:  []string{"\n        const test = true\n          ? <span\n              attr=\"value\"\n            />\n          : <span\n              attr=\"otherValue\"\n            />\n      "},
			Tsx:     true,
			Options: []interface{}{float64(2)},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent"},
				{MessageId: "wrongIndent"},
			},
		},

		// ---- Alternate wrapped in its own parens — indent expected
		// without ternary bump because the `(` resets context. ----
		{
			Code:    "\n        const test = true\n          ? <span attr=\"value\" />\n          : (\n            <span\n                attr=\"otherValue\"\n            />\n          )\n      ",
			Output:  []string{"\n        const test = true\n          ? <span attr=\"value\" />\n          : (\n            <span\n              attr=\"otherValue\"\n            />\n          )\n      "},
			Tsx:     true,
			Options: []interface{}{float64(2)},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent"},
			},
		},

		// ---- Ternary alternate: only one prop. ----
		{
			Code:    "\n        {test.isLoading\n          ? <Value/>\n          : <OtherValue\n            some={aaa}/>\n        }\n      ",
			Output:  []string{"\n        {test.isLoading\n          ? <Value/>\n          : <OtherValue\n              some={aaa}/>\n        }\n      "},
			Tsx:     true,
			Options: []interface{}{float64(2)},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent"},
			},
		},

		// ---- Ternary alternate: two props, both bumped. ----
		{
			Code:    "\n        {test.isLoading\n          ? <Value/>\n          : <OtherValue\n            some={aaa}\n            other={bbb}/>\n        }\n      ",
			Output:  []string{"\n        {test.isLoading\n          ? <Value/>\n          : <OtherValue\n              some={aaa}\n              other={bbb}/>\n        }\n      "},
			Tsx:     true,
			Options: []interface{}{float64(2)},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent"},
				{MessageId: "wrongIndent"},
			},
		},

		// ---- Ternary consequent inside JSXExpressionContainer: bump
		// applies. ----
		{
			Code:    "\n        {this.props.test\n          ? <span\n            className=\"value\"\n            some={{aaa}}\n            />\n          : null}\n      ",
			Output:  []string{"\n        {this.props.test\n          ? <span\n              className=\"value\"\n              some={{aaa}}\n            />\n          : null}\n      "},
			Tsx:     true,
			Options: []interface{}{float64(2)},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent"},
				{MessageId: "wrongIndent"},
			},
		},

		// ---- Tab option, prop has no leading tab — needed=1 character. ----
		{
			Code:    "\n        <App1\n            foo\n        />\n      ",
			Output:  []string{"\n        <App1\n\tfoo\n        />\n      "},
			Tsx:     true,
			Options: []interface{}{"tab"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent"},
			},
		},

		// ---- Tab option, too many tabs — needed=5 characters. ----
		{
			Code:    "\n\t\t\t\t<App\n\t\t\t\t\t\t\tfoo\n\t\t\t\t/>\n\t\t\t",
			Output:  []string{"\n\t\t\t\t<App\n\t\t\t\t\tfoo\n\t\t\t\t/>\n\t\t\t"},
			Tsx:     true,
			Options: []interface{}{"tab"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent"},
			},
		},

		// ---- 'first': prop at col 10 must match firstPropCol=13. ----
		{
			Code:    "\n        <App a\n          b\n        />\n      ",
			Output:  []string{"\n        <App a\n             b\n        />\n      "},
			Tsx:     true,
			Options: []interface{}{"first"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent"},
			},
		},

		// ---- 'first': prop at col 11 must match firstPropCol=14. ----
		{
			Code:    "\n        <App  a\n           b\n        />\n      ",
			Output:  []string{"\n        <App  a\n              b\n        />\n      "},
			Tsx:     true,
			Options: []interface{}{"first"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent"},
			},
		},

		// ---- 'first', first prop on new line: subsequent props must
		// match the FIRST prop's column. ----
		{
			Code:    "\n        <App\n              a\n           b\n        />\n      ",
			Output:  []string{"\n        <App\n              a\n              b\n        />\n      "},
			Tsx:     true,
			Options: []interface{}{"first"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent"},
			},
		},

		// ---- 'first', three props with two mis-aligned. ----
		{
			Code:    "\n        <App\n          a\n         b\n           c\n        />\n      ",
			Output:  []string{"\n        <App\n          a\n          b\n          c\n        />\n      "},
			Tsx:     true,
			Options: []interface{}{"first"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent"},
				{MessageId: "wrongIndent"},
			},
		},

		// ---- ignoreTernaryOperator:false, return JSX one indent off. ----
		{
			Code:    "\n        const F = () => {\n          const foo = true\n            ? <div id=\"id\">test</div>\n            : false;\n\n          return <div\n              id=\"id\"\n          >\n            test\n          </div>\n        }\n      ",
			Output:  []string{"\n        const F = () => {\n          const foo = true\n            ? <div id=\"id\">test</div>\n            : false;\n\n          return <div\n            id=\"id\"\n          >\n            test\n          </div>\n        }\n      "},
			Tsx:     true,
			Options: []interface{}{map[string]interface{}{"indentMode": float64(2), "ignoreTernaryOperator": false}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent"},
			},
		},

		// ---- Same shape, ignoreTernaryOperator:true (return JSX is not in
		// a ternary so behaviour is identical). ----
		{
			Code:    "\n        const F = () => {\n          const foo = true\n            ? <div id=\"id\">test</div>\n            : false;\n\n          return <div\n              id=\"id\"\n          >\n            test\n          </div>\n        }\n      ",
			Output:  []string{"\n        const F = () => {\n          const foo = true\n            ? <div id=\"id\">test</div>\n            : false;\n\n          return <div\n            id=\"id\"\n          >\n            test\n          </div>\n        }\n      "},
			Tsx:     true,
			Options: []interface{}{map[string]interface{}{"indentMode": float64(2), "ignoreTernaryOperator": true}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent"},
			},
		},

		// ---- Tab indent + return JSX off by one tab. ----
		{
			Code:    "\n\t\t\t\tconst F = () => {\n\t\t\t\t\tconst foo = true\n\t\t\t\t\t\t? <div id=\"id\">test</div>\n\t\t\t\t\t\t: false;\n\n\t\t\t\t\treturn <div\n\t\t\t\t\t\t\tid=\"id\"\n\t\t\t\t\t>\n\t\t\t\t\t\ttest\n\t\t\t\t\t</div>\n\t\t\t\t}\n",
			Output:  []string{"\n\t\t\t\tconst F = () => {\n\t\t\t\t\tconst foo = true\n\t\t\t\t\t\t? <div id=\"id\">test</div>\n\t\t\t\t\t\t: false;\n\n\t\t\t\t\treturn <div\n\t\t\t\t\t\tid=\"id\"\n\t\t\t\t\t>\n\t\t\t\t\t\ttest\n\t\t\t\t\t</div>\n\t\t\t\t}\n"},
			Tsx:     true,
			Options: []interface{}{map[string]interface{}{"indentMode": "tab", "ignoreTernaryOperator": false}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent"},
			},
		},
		{
			Code:    "\n\t\t\t\tconst F = () => {\n\t\t\t\t\tconst foo = true\n\t\t\t\t\t\t? <div id=\"id\">test</div>\n\t\t\t\t\t\t: false;\n\n\t\t\t\t\treturn <div\n\t\t\t\t\t\t\tid=\"id\"\n\t\t\t\t\t>\n\t\t\t\t\t\ttest\n\t\t\t\t\t</div>\n\t\t\t\t}\n",
			Output:  []string{"\n\t\t\t\tconst F = () => {\n\t\t\t\t\tconst foo = true\n\t\t\t\t\t\t? <div id=\"id\">test</div>\n\t\t\t\t\t\t: false;\n\n\t\t\t\t\treturn <div\n\t\t\t\t\t\tid=\"id\"\n\t\t\t\t\t>\n\t\t\t\t\t\ttest\n\t\t\t\t\t</div>\n\t\t\t\t}\n"},
			Tsx:     true,
			Options: []interface{}{map[string]interface{}{"indentMode": "tab", "ignoreTernaryOperator": true}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent"},
			},
		},

		// ---- tsgo-specific lock-ins (NOT in upstream) ----
		// Spread + named prop both at wrong indent — both must report.
		{
			Code:   "\n        <App\n          {...rest}\n          foo\n        />\n      ",
			Output: []string{"\n        <App\n            {...rest}\n            foo\n        />\n      "},
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent"},
				{MessageId: "wrongIndent"},
			},
		},

		// Bracket-reset cancels the ternary bump: actual props are
		// indented as if the bump applied, but because the first prop's
		// first source line contains `<`, upstream's `useBracket` reset
		// fires and the expected indent is propIndent (NOT
		// propIndent + indentSize). Locks in: without the reset, this
		// would silently accept props at indent 10.
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

		// State-leak lock-in: ternary state from the first JSX must NOT
		// affect the second JSX. The first element's `?` line sets
		// `isUsingOperator` upstream, but the second element's own `<`
		// line doesn't lead with `?`/`:` — it should fall back to the
		// non-bumped expected indent.
		{
			Code:    "\n        const a = c1 ? <X\n          a\n        /> : null;\n        <Y\n            foo\n        />\n      ",
			Output:  []string{"\n        const a = c1 ? <X\n          a\n        /> : null;\n        <Y\n          foo\n        />\n      "},
			Tsx:     true,
			Options: []interface{}{float64(2)},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent"},
			},
		},

		// ---- Message text + diagnostic position lock-ins ----
		// Default 4-space, foo at indent 10 must be 12. Asserts the
		// FULL diagnostic shape: messageId, exact message text
		// ("12 space characters"), and 1-based Line/Column/EndLine/
		// EndColumn anchored at the JsxAttribute's trimmed range.
		{
			Code:   "\n        <App\n          foo\n        />\n      ",
			Output: []string{"\n        <App\n            foo\n        />\n      "},
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "wrongIndent",
					Message:   "Expected indentation of 12 space characters but found 10.",
					Line:      3,
					Column:    11,
					EndLine:   3,
					EndColumn: 14,
				},
			},
		},

		// Tab option, prop has zero leading tabs — needed=1 must use
		// the SINGULAR "character" (not "characters"); locks in the
		// pluralization branch in the message.
		{
			Code:    "\n        <App1\n            foo\n        />\n      ",
			Output:  []string{"\n        <App1\n\tfoo\n        />\n      "},
			Tsx:     true,
			Options: []interface{}{"tab"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "wrongIndent",
					Message:   "Expected indentation of 1 tab character but found 0.",
					Line:      3,
					Column:    13,
					EndLine:   3,
					EndColumn: 16,
				},
			},
		},

		// 'first' alignment mismatch — expected indent is computed from
		// the first prop's column (here `a` sits at byte column 13,
		// which becomes the expected indent count).
		{
			Code:    "\n        <App a\n          b\n        />\n      ",
			Output:  []string{"\n        <App a\n             b\n        />\n      "},
			Tsx:     true,
			Options: []interface{}{"first"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "wrongIndent",
					Message:   "Expected indentation of 13 space characters but found 10.",
					Line:      3,
					Column:    11,
					EndLine:   3,
					EndColumn: 12,
				},
			},
		},

		// Multi-error sequencing — two mis-indented props on different
		// lines must each report on their OWN line, not on a single
		// merged location.
		{
			Code:   "\n        <App\n          foo\n          bar\n        />\n      ",
			Output: []string{"\n        <App\n            foo\n            bar\n        />\n      "},
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "wrongIndent",
					Message:   "Expected indentation of 12 space characters but found 10.",
					Line:      3,
					Column:    11,
				},
				{
					MessageId: "wrongIndent",
					Message:   "Expected indentation of 12 space characters but found 10.",
					Line:      4,
					Column:    11,
				},
			},
		},

		// 'first' with three mismatched props — Line increments must
		// stay in source order; locks in that the iteration follows
		// `Attributes.Properties.Nodes` order, not visit order.
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

		// Negative indent option — `[-2]` makes propIndent < 0. Locks in:
		//   1. The rule MUST NOT panic (early `strings.Repeat(' ', -2)`
		//      crash before fix). The omitted `Output` field asserts no
		//      fix was applied (matching upstream ESLint, whose `repeat()`
		//      throws and is silently dropped).
		//   2. The diagnostic message preserves the negative `needed`
		//      verbatim — same shape upstream produces before its fix
		//      lambda blows up.
		{
			Code:    "<App\nfoo\n/>",
			Tsx:     true,
			Options: []interface{}{float64(-2)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "wrongIndent",
					Message:   "Expected indentation of -2 space characters but found 0.",
					Line:      2,
					Column:    1,
					EndLine:   2,
					EndColumn: 4,
				},
			},
		},

		// Multi-byte tag name in `'first'` mode — `<中Foo>` adds 1 UTF-16
		// code unit but 3 UTF-8 bytes. Locks in that propIndent uses
		// UTF-16 character column (matches ESLint's `loc.start.column`),
		// not byte offset. Without the UTF-16 fix, this case would
		// erroneously expect indent 16 (byte offset) instead of 14.
		{
			Code:    "\n        <中Foo a=\"x\"\n           b=\"y\"\n        />\n      ",
			Output:  []string{"\n        <中Foo a=\"x\"\n              b=\"y\"\n        />\n      "},
			Tsx:     true,
			Options: []interface{}{"first"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "wrongIndent",
					Message:   "Expected indentation of 14 space characters but found 11.",
					Line:      3,
					Column:    12,
					EndLine:   3,
					EndColumn: 17,
				},
			},
		},

		// Multi-line attribute value position assertion (PORT_RULE.md
		// Phase 4 Step 6 requires ≥1 multi-line case with full
		// Line+Column+EndLine+EndColumn). The `beforeNav` attribute
		// spans 4 lines; the diagnostic's end-of-range must land on the
		// CLOSING line/column of the JsxAttribute, not on its first line.
		{
			Code: "\n        <BaseLayout\n          beforeNav={\n            <Banner />\n          }\n        />\n      ",
			Output: []string{
				"\n        <BaseLayout\n            beforeNav={\n            <Banner />\n          }\n        />\n      ",
			},
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "wrongIndent",
					Message:   "Expected indentation of 12 space characters but found 10.",
					Line:      3,
					Column:    11,
					EndLine:   5,
					EndColumn: 12,
				},
			},
		},
	})
}

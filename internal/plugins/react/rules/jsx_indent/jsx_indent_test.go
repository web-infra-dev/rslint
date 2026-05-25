package jsx_indent

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestJsxIndentRule mirrors the upstream eslint-plugin-react test suite at
// `tests/lib/rules/jsx-indent.js`. Each upstream case is migrated 1:1; cases
// that exercise stage-1 / non-tsgo features are kept as `Skip: true` with a
// short reason so the file still maps to the upstream layout.
//
// Layout note: most upstream `valid` cases anchor their JSX with the same
// 8-space outer indent ESLint's tagged-template literals introduce, so we
// keep that prefix verbatim. Switching them to flush-left would break the
// indent arithmetic.
func TestJsxIndentRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxIndentRule, []rule_tester.ValidTestCase{
		// ---- Single-line elements / fragments — no inner content to check ----
		{Code: "\n        <App></App>\n      ", Tsx: true},
		{Code: "\n        <></>\n      ", Tsx: true},

		// ---- Multi-line element with no children ----
		{Code: "\n        <App>\n        </App>\n      ", Tsx: true},
		{Code: "\n        <>\n        </>\n      ", Tsx: true},

		// ---- 2-space option ----
		{
			Code:    "\n        <App>\n          <Foo />\n        </App>\n      ",
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		{
			Code:    "\n        <App>\n          <></>\n        </App>\n      ",
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		{
			Code:    "\n        <>\n          <Foo />\n        </>\n      ",
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},

		// ---- 0-indent option ----
		{
			Code:    "\n        <App>\n        <Foo />\n        </App>\n      ",
			Tsx:     true,
			Options: []interface{}{float64(0)},
		},

		// ---- Negative offset (rare upstream test — child sits two cols
		// to the LEFT of the open / close). ----
		{
			Code:    "\n          <App>\n        <Foo />\n          </App>\n      ",
			Tsx:     true,
			Options: []interface{}{float64(-2)},
		},

		// ---- Tab indentation ----
		{
			Code:    "\n\t\t\t\t<App>\n\t\t\t\t\t<Foo />\n\t\t\t\t</App>\n\t\t\t",
			Tsx:     true,
			Options: []interface{}{"tab"},
		},

		// ---- Function with `return <jsx>` (no parens) — return / closing
		// tag share a column. ----
		{
			Code:    "\n        function App() {\n          return <App>\n            <Foo />\n          </App>;\n        }\n      ",
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		{
			Code:    "\n        function App() {\n          return <App>\n            <></>\n          </App>;\n        }\n      ",
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},

		// ---- Function with `return (<jsx>)` — opening paren on return line. ----
		{
			Code:    "\n        function App() {\n          return (<App>\n            <Foo />\n          </App>);\n        }\n      ",
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		{
			Code:    "\n        function App() {\n          return (<App>\n            <></>\n          </App>);\n        }\n      ",
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},

		// ---- Function with `return (\n  <jsx>\n)` — JSX on its own indent
		// inside the parens. ----
		{
			Code:    "\n        function App() {\n          return (\n            <App>\n              <Foo />\n            </App>\n          );\n        }\n      ",
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		{
			Code:    "\n        function App() {\n          return (\n            <App>\n              <></>\n            </App>\n          );\n        }\n      ",
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},

		// ---- Call argument with parenthesized JSX ----
		{
			Code:    "\n        it(\n          (\n            <div>\n              <span />\n            </div>\n          )\n        )\n      ",
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		{
			Code:    "\n        it(\n          (\n            <div>\n              <></>\n            </div>\n          )\n        )\n      ",
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		{
			Code:    "\n        it(\n          (<div>\n            <span />\n            <span />\n            <span />\n          </div>)\n        )\n      ",
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		{
			Code:    "\n        (\n          <div>\n            <span />\n          </div>\n        )\n      ",
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},

		// ---- Logical `&&` in expression-statement context — JSX may
		// share parent indent (default indentLogicalExpressions: false). ----
		{
			Code:    "\n        {\n          head.title &&\n          <h1>\n            {head.title}\n          </h1>\n        }\n      ",
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		{
			Code:    "\n        {\n          head.title &&\n          <>\n            {head.title}\n          </>\n        }\n      ",
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		{
			Code:    "\n        {\n          head.title &&\n            <h1>\n              {head.title}\n            </h1>\n        }\n      ",
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		{
			Code:    "\n        {\n          head.title && (\n          <h1>\n            {head.title}\n          </h1>)\n        }\n      ",
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		{
			Code:    "\n        {\n          head.title && (\n            <h1>\n              {head.title}\n            </h1>\n          )\n        }\n      ",
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},

		// ---- Array literal of JSX (comma-anchored to array start). ----
		{
			Code:    "\n        [\n          <div />,\n          <div />\n        ]\n      ",
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		{
			Code:    "\n        [\n          <></>,\n          <></>\n        ]\n      ",
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},

		// ---- Default 4-space, deeply nested array inside JsxExpression ----
		{Code: "\n        <div>\n            {\n                [\n                    <Foo />,\n                    <Bar />\n                ]\n            }\n        </div>\n      ", Tsx: true},
		{Code: "\n        <div>\n            {foo &&\n                [\n                    <Foo />,\n                    <Bar />\n                ]\n            }\n        </div>\n      ", Tsx: true},
		{Code: "\n        <div>\n            {foo &&\n                [\n                    <></>,\n                    <></>\n                ]\n            }\n        </div>\n      ", Tsx: true},

		// ---- Default 4-space, JsxText with mixed inline JSX ----
		{Code: "\n        <div>\n            bar <div>\n                bar\n                bar {foo}\n                bar </div>\n        </div>\n      ", Tsx: true},
		{Code: "\n        <>\n            bar <>\n                bar\n                bar {foo}\n                bar </>\n        </>\n      ", Tsx: true},

		// ---- Multiline ternary — colon at end of consequent ----
		{Code: "\n        foo ?\n            <Foo /> :\n            <Bar />\n      ", Tsx: true},
		{Code: "\n        foo ?\n            <></> :\n            <></>\n      ", Tsx: true},

		// ---- Multiline ternary — colon at start of second expr ----
		{Code: "\n        foo ?\n            <Foo />\n            : <Bar />\n      ", Tsx: true},
		{Code: "\n        foo ?\n            <></>\n            : <></>\n      ", Tsx: true},

		// ---- Multiline ternary — colon on its own line ----
		{Code: "\n        foo ?\n            <Foo />\n        :\n            <Bar />\n      ", Tsx: true},
		{Code: "\n        foo ?\n            <></>\n        :\n            <></>\n      ", Tsx: true},

		// ---- Multiline JSX inside ternary, colon on its own line ----
		{Code: "\n        {!foo ?\n            <Foo\n                onClick={this.onClick}\n            />\n        :\n            <Bar\n                onClick={this.onClick}\n            />\n        }\n      ", Tsx: true},

		// ---- Test expr on first line, colon at end of consequent ----
		{Code: "\n        foo ? <Foo /> :\n        <Bar />\n      ", Tsx: true},
		{Code: "\n        foo ? <></> :\n        <></>\n      ", Tsx: true},

		// ---- Test expr on first line, colon on second line ----
		{Code: "\n        foo ? <Foo />\n        : <Bar />\n      ", Tsx: true},
		{Code: "\n        foo ? <></>\n        : <></>\n      ", Tsx: true},

		// ---- Test expr on first line, colon on its own line ----
		{Code: "\n        foo ? <Foo />\n        :\n        <Bar />\n      ", Tsx: true},
		{Code: "\n        foo ? <></>\n        :\n        <></>\n      ", Tsx: true},

		// ---- Parenthesized first expression ----
		{Code: "\n        foo ? (\n            <Foo />\n        ) :\n            <Bar />\n      ", Tsx: true},
		{Code: "\n        foo ? (\n            <></>\n        ) :\n            <></>\n      ", Tsx: true},
		{Code: "\n        foo ? (\n            <Foo />\n        )\n            : <Bar />\n      ", Tsx: true},
		{Code: "\n        foo ? (\n            <></>\n        )\n            : <></>\n      ", Tsx: true},
		{Code: "\n        foo ? (\n            <Foo />\n        )\n        :\n            <Bar />\n      ", Tsx: true},
		{Code: "\n        foo ? (\n            <></>\n        )\n        :\n            <></>\n      ", Tsx: true},

		// ---- Parenthesized second expression ----
		{Code: "\n        foo ?\n            <Foo /> : (\n                <Bar />\n            )\n      ", Tsx: true},
		{Code: "\n        foo ?\n            <></> : (\n                <></>\n            )\n      ", Tsx: true},
		{Code: "\n        foo ?\n            <Foo />\n        : (\n            <Bar />\n        )\n      ", Tsx: true},
		{Code: "\n        foo ?\n            <></>\n        : (\n            <></>\n        )\n      ", Tsx: true},
		{Code: "\n        foo ?\n            <Foo />\n            : (\n                <Bar />\n            )\n      ", Tsx: true},
		{Code: "\n        foo ?\n            <></>\n            : (\n                <></>\n            )\n      ", Tsx: true},

		// ---- Both branches parenthesized ----
		{Code: "\n        foo ? (\n            <Foo />\n        ) : (\n            <Bar />\n        )\n      ", Tsx: true},
		{Code: "\n        foo ? (\n            <></>\n        ) : (\n            <></>\n        )\n      ", Tsx: true},
		{Code: "\n        foo ? (\n            <Foo />\n        )\n        : (\n            <Bar />\n        )\n      ", Tsx: true},
		{Code: "\n        foo ? (\n            <></>\n        )\n        : (\n            <></>\n        )\n      ", Tsx: true},
		{Code: "\n        foo ? (\n            <Foo />\n        )\n        :\n        (\n            <Bar />\n        )\n      ", Tsx: true},
		{Code: "\n        foo ? (\n            <></>\n        )\n        :\n        (\n            <></>\n        )\n      ", Tsx: true},

		// ---- Mixed test-on-first-line + paren second branch ----
		{Code: "\n        foo ? <Foo /> : (\n            <Bar />\n        )\n      ", Tsx: true},
		{Code: "\n        foo ? <></> : (\n            <></>\n        )\n      ", Tsx: true},
		{Code: "\n        foo ? <Foo />\n        : (<Bar />)\n      ", Tsx: true},
		{Code: "\n        foo ? <></>\n        : (<></>)\n      ", Tsx: true},
		{Code: "\n        foo ? <Foo />\n        : (\n            <Bar />\n        )\n      ", Tsx: true},
		{Code: "\n        foo ? <></>\n        : (\n            <></>\n        )\n      ", Tsx: true},

		// ---- JsxExpression-wrapped ternary, JSX with attrs spanning
		// multiple lines (opts [2]) ----
		{
			Code:    "\n        <span>\n          {condition ?\n            <Thing\n              foo={`bar`}\n            /> :\n            <Thing/>\n          }\n        </span>\n      ",
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		{
			Code:    "\n        <span>\n          {condition ?\n            <Thing\n              foo={\"bar\"}\n            /> :\n            <Thing/>\n          }\n        </span>\n      ",
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		{
			Code:    "\n        function foo() {\n          <span>\n            {condition ?\n              <Thing\n                foo={superFoo}\n              /> :\n              <Thing/>\n            }\n          </span>\n        }\n      ",
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		{
			Code:    "\n        function foo() {\n          <span>\n            {condition ?\n              <Thing\n                foo={superFoo}\n              /> :\n              <></>\n            }\n          </span>\n        }\n      ",
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},

		// ---- do { } expression (stage-1) — tsgo has no native node;
		// upstream uses the babel-eslint parser. Skip with reason. ----
		{
			Code: "\n        <span>\n            {do {\n                const num = rollDice();\n                <Thing num={num} />;\n            }}\n        </span>\n      ",
			Tsx:  true,
			Skip: true, // SKIP: rslint's tsgo parser does not support do-expressions.
		},
		{
			Code: "\n        <span>\n            {(do {\n                const num = rollDice();\n                <Thing num={num} />;\n            })}\n        </span>\n      ",
			Tsx:  true,
			Skip: true, // SKIP: rslint's tsgo parser does not support do-expressions.
		},
		{
			Code: "\n        <span>\n            {do {\n                const purposeOfLife = getPurposeOfLife();\n                if (purposeOfLife == 42) {\n                    <Thing />;\n                } else {\n                    <AnotherThing />;\n                }\n            }}\n        </span>\n      ",
			Tsx:  true,
			Skip: true, // SKIP: rslint's tsgo parser does not support do-expressions.
		},
		{
			Code: "\n        <span>\n            {(do {\n                const purposeOfLife = getPurposeOfLife();\n                if (purposeOfLife == 42) {\n                    <Thing />;\n                } else {\n                    <AnotherThing />;\n                }\n            })}\n        </span>\n      ",
			Tsx:  true,
			Skip: true, // SKIP: rslint's tsgo parser does not support do-expressions.
		},
		{
			Code: "\n        <span>\n            {do {\n                <Thing num={rollDice()} />;\n            }}\n        </span>\n      ",
			Tsx:  true,
			Skip: true, // SKIP: rslint's tsgo parser does not support do-expressions.
		},
		{
			Code: "\n        <span>\n            {(do {\n                <Thing num={rollDice()} />;\n            })}\n        </span>\n      ",
			Tsx:  true,
			Skip: true, // SKIP: rslint's tsgo parser does not support do-expressions.
		},
		{
			Code: "\n        <span>\n            {do {\n                <Thing num={rollDice()} />;\n                <Thing num={rollDice()} />;\n            }}\n        </span>\n      ",
			Tsx:  true,
			Skip: true, // SKIP: rslint's tsgo parser does not support do-expressions.
		},
		{
			Code: "\n        <span>\n            {(do {\n                <Thing num={rollDice()} />;\n                <Thing num={rollDice()} />;\n            })}\n        </span>\n      ",
			Tsx:  true,
			Skip: true, // SKIP: rslint's tsgo parser does not support do-expressions.
		},
		{
			Code: "\n        <span>\n            {do {\n                const purposeOfLife = 42;\n                <Thing num={purposeOfLife} />;\n                <Thing num={purposeOfLife} />;\n            }}\n        </span>\n      ",
			Tsx:  true,
			Skip: true, // SKIP: rslint's tsgo parser does not support do-expressions.
		},
		{
			Code: "\n        <span>\n            {(do {\n                const purposeOfLife = 42;\n                <Thing num={purposeOfLife} />;\n                <Thing num={purposeOfLife} />;\n            })}\n        </span>\n      ",
			Tsx:  true,
			Skip: true, // SKIP: rslint's tsgo parser does not support do-expressions.
		},

		// ---- Class component with deeply nested return ----
		{
			Code:    "\n        class Test extends React.Component {\n          render() {\n            return (\n              <div>\n                <div />\n                <div />\n              </div>\n            );\n          }\n        }\n      ",
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		{
			Code:    "\n        class Test extends React.Component {\n          render() {\n            return (\n              <>\n                <></>\n                <></>\n              </>\n            );\n          }\n        }\n      ",
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},

		// ---- Drift inside attribute is allowed by default (checkAttributes:false) ----
		{
			Code:    "\n        const Component = () => (\n          <View\n            ListFooterComponent={(\n              <View\n                rowSpan={3}\n                placeholder=\"placeholder text here\"\n              />\n        )}\n          />\n        );\n      ",
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		{
			Code:    "\nconst Component = () => (\n\t<View\n\t\tListFooterComponent={(\n\t\t\t<View\n\t\t\t\trowSpan={3}\n\t\t\t\tplaceholder=\"placeholder text here\"\n\t\t\t/>\n)}\n\t/>\n);\n    ",
			Tsx:     true,
			Options: []interface{}{"tab"},
		},
		{
			Code:    "\n        const Component = () => (\n          <View\n            ListFooterComponent={(\n              <View\n                rowSpan={3}\n                placeholder=\"placeholder text here\"\n              />\n        )}\n          />\n        );\n      ",
			Tsx:     true,
			Options: []interface{}{float64(2), map[string]interface{}{"checkAttributes": false}},
		},
		{
			Code:    "\nconst Component = () => (\n\t<View\n\t\tListFooterComponent={(\n\t\t\t<View\n\t\t\t\trowSpan={3}\n\t\t\t\tplaceholder=\"placeholder text here\"\n\t\t\t/>\n)}\n\t/>\n);\n    ",
			Tsx:     true,
			Options: []interface{}{"tab", map[string]interface{}{"checkAttributes": false}},
		},

		// ---- checkAttributes:true with already-correct indentation ----
		{
			Code:    "\n        function Foo() {\n          return (\n            <input\n              type=\"radio\"\n              defaultChecked\n            />\n          );\n        }\n      ",
			Tsx:     true,
			Options: []interface{}{float64(2), map[string]interface{}{"checkAttributes": true}},
		},

		// ---- indentLogicalExpressions:true with already-correct nested indent ----
		{
			Code:    "\n        function Foo() {\n          return (\n            <div>\n              {condition && (\n                <p>Bar</p>\n              )}\n            </div>\n          );\n        }\n      ",
			Tsx:     true,
			Options: []interface{}{float64(2), map[string]interface{}{"indentLogicalExpressions": true}},
		},

		// ---- JsxText content (default 4-space) ----
		{Code: "\n        <App>\n            text\n        </App>\n      ", Tsx: true},
		{Code: "\n        <App>\n            text\n            text\n            text\n        </App>\n      ", Tsx: true},

		// ---- Tab + JsxText ----
		{
			Code:    "\n\t\t\t\t<App>\n\t\t\t\t\ttext\n\t\t\t\t</App>\n\t\t\t",
			Tsx:     true,
			Options: []interface{}{"tab"},
		},
		{
			Code:    "\n\t\t\t\t<App>\n\t\t\t\t\t{undefined}\n\t\t\t\t\t{null}\n\t\t\t\t\t{true}\n\t\t\t\t\t{false}\n\t\t\t\t\t{42}\n\t\t\t\t\t{NaN}\n\t\t\t\t\t{\"foo\"}\n\t\t\t\t</App>\n\t\t\t",
			Tsx:     true,
			Options: []interface{}{"tab"},
		},

		// ---- Literals outside JSX must NOT be checked (#2563) ----
		{Code: "\n        function foo() {\n          const a = `aa`;\n          const b = `b\nb`;\n        }\n      ", Tsx: true},

		// ---- Arrow body returns / various function shapes ----
		{
			Code:    "\n        function App() {\n          return (\n            <App />\n          );\n        }\n      ",
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		{
			Code:    "\n        function App() {\n          return <App>\n            <Foo />\n          </App>;\n        }\n      ",
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		{
			Code:    "\n        const myFunction = () => (\n          [\n            <Tag\n              {...properties}\n            />,\n            <Tag\n              {...properties}\n            />,\n            <Tag\n              {...properties}\n            />,\n          ]\n        )\n      ",
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		{
			Code:    "\n        const Item = ({ id, name, onSelect }) => <div onClick={onSelect}>\n          {id}: {name}\n        </div>;\n      ",
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},

		// ---- Flow-typed prop pattern. Skip — Flow not supported. ----
		{
			Code: "\n        type Props = {\n          email: string,\n          password: string,\n          error: string,\n        }\n\n        const SomeFormComponent = ({\n          email,\n          password,\n          error,\n        }: Props) => (\n          // JSX\n        );\n      ",
			Tsx:  true,
			Skip: true, // SKIP: Flow type syntax is not supported by rslint's tsgo parser.
		},

		// ---- Multi-line opening tag with many attrs ----
		{
			Code:    "\n        <a role={'button'}\n          className={`navbar-burger ${open ? 'is-active' : ''}`}\n          href={'#'}\n          aria-label={'menu'}\n          aria-expanded={false}\n          onClick={openMenu}>\n          <span aria-hidden={'true'}/>\n          <span aria-hidden={'true'}/>\n          <span aria-hidden={'true'}/>\n        </a>\n      ",
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},

		// ---- Class component with class fields (TS supports the same shape) ----
		{
			Code:    "\n        export default class App extends React.Component {\n          state = {\n            name: '',\n          }\n\n          componentDidMount() {\n            this.fetchName()\n              .then(name => {\n                this.setState({name})\n              });\n          }\n\n          fetchName = () => {\n            const url = 'https://api.github.com/users/job13er'\n            return fetch(url)\n              .then(resp => resp.json())\n              .then(json => json.name)\n          }\n\n          render() {\n            const {name} = this.state\n            return (\n              <h1>Hello, {name}</h1>\n            )\n          }\n        }\n      ",
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},

		// ---- Large indentSize with non-JSX return body — must not flag. ----
		{
			Code:    "\n        function test (foo) {\n          return foo != null\n            ? Math.max(0, Math.min(1, 10))\n            : 0\n        }\n      ",
			Tsx:     true,
			Options: []interface{}{float64(99)},
		},
		{
			Code:    "\n        function test (foo) {\n          return foo != null\n            ? <div>foo</div>\n            : <div>bar</div>\n        }\n      ",
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},

		// =================================================================
		// Below: real-world / edge-shape coverage beyond the upstream
		// suite. These exercise tsgo↔ESTree shape differences and common
		// React patterns the upstream suite never gets to.
		// =================================================================

		// ---- TS `as` cast around a multi-line JSX expression — anchor
		// for closing tag must skip the wrapping `(... as T)`. ----
		{
			Code: `var x = (
  <App>
    <Foo />
  </App>
) as React.ReactElement;
`,
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		// ---- TS `satisfies` wrapper around JSX. Same structural shape
		// — skipExpressionWrappers must transparently drop it. ----
		{
			Code: `var x = (
  <App>
    <Foo />
  </App>
) satisfies React.ReactElement;
`,
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		// ---- TS non-null `!` on a JSX expression. tsgo represents this
		// as `NonNullExpression` wrapping the JSX; the rule must treat
		// it as transparent. ----
		{
			Code: `var x = (
  <App>
    <Foo />
  </App>
)!;
`,
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		// ---- TS-`as` wrapper on the alternate of a `?:` — locks
		// `walkUpThroughExprWrappers` cooperating with
		// `isAlternateInConditionalExp`. ----
		{
			Code: `var x = condition ? (
  <App>
    <Foo />
  </App>
) : (
  <Bar /> as React.ReactElement
);
`,
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		// ---- Generic JSX `<Foo<T>>...</Foo>` (TS-specific syntax). The
		// `<T>` type-arguments list sits inside the opening tag but the
		// opening tag's `<` is still at the line start; child indent
		// follows the parent's line. ----
		{
			Code: `var x = (
  <Foo<string>>
    <Bar />
  </Foo>
);
`,
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		// ---- JSX member-expression tag name (`<Foo.Bar>`). Tag-name
		// shape is irrelevant to indentation arithmetic; this locks
		// in that behaviour. ----
		{
			Code: `var x = (
  <Foo.Bar>
    <Foo.Baz />
  </Foo.Bar>
);
`,
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		// ---- Three-level JSX member expression. ----
		{
			Code: `var x = (
  <Foo.Bar.Baz>
    <Quux />
  </Foo.Bar.Baz>
);
`,
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		// ---- Namespaced JSX tag (`<svg:rect />`). ----
		{
			Code: `var x = (
  <svg:svg>
    <svg:rect />
  </svg:svg>
);
`,
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		// ---- 6-level deeply nested element tree. Locks the
		// jsxParentOpening anchor for arbitrary depth. ----
		{
			Code: `var x = (
  <A>
    <B>
      <C>
        <D>
          <E>
            <F />
          </E>
        </D>
      </C>
    </B>
  </A>
);
`,
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		// ---- Sibling fragment + element children at the same level. ----
		{
			Code: `var x = (
  <App>
    <>
      <span />
    </>
    <span />
  </App>
);
`,
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		// ---- Common React pattern: `React.memo(() => <jsx>)`. The JSX
		// sits inside a CallExpression argument list; the anchor for
		// `<App>` is the call argument's start line. ----
		{
			Code: `var Wrapped = React.memo(() => (
  <App>
    <Foo />
  </App>
));
`,
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		// ---- React.forwardRef wrap. ----
		{
			Code: `var Wrapped = React.forwardRef((props, ref) => (
  <App ref={ref}>
    <Foo />
  </App>
));
`,
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		// ---- map() returning JSX with a complex callback body. ----
		{
			Code: `var els = items.map((item, i) => (
  <Item key={i}>
    {item.name}
  </Item>
));
`,
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		// ---- Spread attribute in opening tag (`{...props}`). The
		// JsxSpreadAttribute is treated like a normal attribute by the
		// rule. ----
		{
			Code: `var x = (
  <App
    {...props}
    foo="bar"
  >
    <Foo />
  </App>
);
`,
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		// ---- CRLF line endings, 2-space indent. The rule must treat
		// `\r\n` like `\n` when locating start-of-line. ----
		{
			Code:    "var x = (\r\n  <App>\r\n    <Foo />\r\n  </App>\r\n);\r\n",
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		// ---- HTML-entity content in a JsxText line. The entity is
		// regular `\S` content — locks that the rule's literal scan
		// doesn't mis-handle the `&`. ----
		{
			Code: `var x = (
  <App>
    Hello&nbsp;World
  </App>
);
`,
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		// ---- CJK content in a JsxText line. ----
		{
			Code: `var x = (
  <App>
    你好，世界
  </App>
);
`,
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		// ---- Arrow function with parenthesized JSX body. ----
		{
			Code: `var Make = () => (
  <App>
    <Foo />
  </App>
);
`,
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		// ---- async function returning a JSX-only expression — not
		// flagged by the ReturnStatement listener (it requires
		// FunctionDeclaration/FunctionExpression, not async-arrow). ----
		{
			Code: `var f = async () => (
  <App>
    <Foo />
  </App>
);
`,
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		// ---- React.createContext + Provider with multi-line children. ----
		{
			Code: `var x = (
  <ThemeContext.Provider value={theme}>
    <App>
      <Foo />
    </App>
  </ThemeContext.Provider>
);
`,
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		// ---- compose / HOC pattern: nested call expressions wrapping a
		// JSX-returning arrow. ----
		{
			Code: `var Wrapped = compose(withRouter, connect(mapState))(({ items }) => (
  <App>
    <Foo />
  </App>
));
`,
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		// ---- Conditional rendering: ternary + nullish-coalesce mix. ----
		{
			Code: `var x = (
  <App>
    {loading
      ? <Spinner />
      : data ?? <Empty />}
  </App>
);
`,
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		// ---- Multiple JsxText siblings separated by JsxExpressions. ----
		{
			Code: `var x = (
  <App>
    Hello
    {", "}
    World
  </App>
);
`,
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		// ---- JsxText that starts with non-whitespace (no leading
		// newline). The first JsxText after an element opening with
		// content directly inline. ----
		{
			Code: `var x = <App>inline-text<Foo />tail-text</App>;
`,
			Tsx: true,
		},
		// ---- Empty JsxElement / JsxFragment / no-attribute self-closing. ----
		{
			Code: `var x = (
  <App>
    <></>
    <Foo />
    <Bar></Bar>
  </App>
);
`,
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		// ---- Multi-line comments INSIDE a multi-line JSX
		// expression don't break anchor calculations. ----
		{
			Code: `var x = (
  <App>
    {/* block comment */}
    <Foo />
    {
      // line comment
      bar
    }
  </App>
);
`,
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		// ---- Tag with type-cast attribute value: `<Foo bar={x as number} />`.
		// Type-cast wrapper inside the attr expression is irrelevant. ----
		{
			Code: `var x = (
  <App>
    <Foo bar={x as number} />
  </App>
);
`,
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		// ---- JSX as a default parameter value. ----
		{
			Code: `function Wrap(child = (
  <Default>
    <Foo />
  </Default>
)) {
  return child;
}
`,
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		// ---- JSX as a value in an object literal. ----
		{
			Code: `var x = {
  child: (
    <App>
      <Foo />
    </App>
  ),
};
`,
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		// ---- React.lazy + Suspense pattern. ----
		{
			Code: `var Page = React.lazy(() => import('./Page'));
var x = (
  <Suspense fallback={<Spinner />}>
    <Page />
  </Suspense>
);
`,
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		// ---- Hooks / functional component pattern with useState. ----
		{
			Code: `function Counter() {
  const [n, setN] = useState(0);
  return (
    <div>
      <span>{n}</span>
      <button onClick={() => setN(n + 1)}>+</button>
    </div>
  );
}
`,
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		// ---- Tab option, deep nesting, mixed empty lines. ----
		{
			Code:    "var x = (\n\t<App>\n\n\t\t<Foo />\n\n\t</App>\n);\n",
			Tsx:     true,
			Options: []interface{}{"tab"},
		},
		// ---- Switch case returning JSX (FunctionExpression context). ----
		{
			Code: `function pick(kind) {
  switch (kind) {
    case 'a':
      return (
        <A>
          <Foo />
        </A>
      );
    default:
      return null;
  }
}
`,
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		// ---- Triple-nested ternary returning JSX. Locks colonAnchor
		// + isAlternateInConditionalExp at multiple depths. ----
		{
			Code: `var x = (
  a ?
    <A />
    : b ?
      <B />
      : c ?
        <C />
        : <D />
);
`,
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		// ---- checkAttributes:true with deeply nested JSX inside the
		// attribute value, plus matching indentation. ----
		{
			Code:    "\n        const x = (\n          <View\n            ListFooterComponent={(\n              <View\n                rowSpan={3}\n              />\n            )}\n          />\n        );\n      ",
			Tsx:     true,
			Options: []interface{}{float64(2), map[string]interface{}{"checkAttributes": true}},
		},
		// ---- indentLogicalExpressions:true on nested logical chains.
		// Locks that the rule still allows the inner JSX to align with
		// the outer JSX when the option is on. ----
		{
			Code: `var x = (
  <App>
    {a && b && (
      <p>both</p>
    )}
  </App>
);
`,
			Tsx:     true,
			Options: []interface{}{float64(2), map[string]interface{}{"indentLogicalExpressions": true}},
		},
		// ---- Whitespace-only JsxText between sibling elements — must
		// not produce diagnostics (no `\S` for the literal-line
		// regex). ----
		{
			Code: "var x = (\n  <App>\n    <Foo />\n\n    <Bar />\n  </App>\n);\n",
			Tsx:  true,
			Options: []interface{}{float64(2)},
		},
		// ---- Inline JSX as attribute prop value (no checkAttributes
		// — drift not enforced). ----
		{
			Code: `var x = (
  <App
    icon={<Icon />}
    title={<Title>foo</Title>}
  >
    <Foo />
  </App>
);
`,
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		// ---- Both boolean options on at the same time. Locks that
		// `checkAttributes` and `indentLogicalExpressions` cooperate
		// rather than interfere. ----
		{
			Code: `var x = (
  <App>
    {condition && (
      <View
        Foo={(
          <Inner />
        )}
      />
    )}
  </App>
);
`,
			Tsx: true,
			Options: []interface{}{
				float64(2),
				map[string]interface{}{
					"checkAttributes":          true,
					"indentLogicalExpressions": true,
				},
			},
		},
		// ---- Empty JsxExpression `{}` — JsxExpression with no inner
		// expression. The handler must not panic; expected indent for
		// the container `{` is the parent's line indent + indentSize. ----
		{
			Code: `var x = (
  <App>
    {}
  </App>
);
`,
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		// ---- `/* eslint-disable */` block at JS-top level suppresses
		// diagnostics that would otherwise fire inside the JSX. Locks
		// that the rule's emitted diagnostics route through the disable
		// manager. JSX-internal `{/* */}` content is JsxText /
		// JsxExpression, NOT a JS trivia comment — only JS-level
		// comments are recognised. The bare-disable form is used here
		// because rule_tester registers the rule under the name "test",
		// not its real "react/jsx-indent" identifier; bare disable
		// applies to all rules unconditionally. ----
		{
			Code: `/* eslint-disable */
var x = (
  <App>
        <Foo />
  </App>
);
`,
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		// ---- Comma in a SequenceExpression mustn't be treated like
		// the array-sibling comma branch — anchor falls through to the
		// non-comma case. ----
		{
			Code: `var x = (a, (
  <App>
    <Foo />
  </App>
));
`,
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		// ---- Boolean attribute (no value) — must not throw in the
		// attribute handler under checkAttributes:true. ----
		{
			Code: `var x = (
  <App>
    <input
      type="radio"
      defaultChecked
      disabled
    />
  </App>
);
`,
			Tsx:     true,
			Options: []interface{}{float64(2), map[string]interface{}{"checkAttributes": true}},
		},
		// ---- Block comment inside an attribute value — anchor must
		// land on the value, not the comment. ----
		{
			Code: `var x = (
  <App>
    <Foo bar={/* annotated */ 42} />
  </App>
);
`,
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		// ---- A JsxText line containing only an NBSP-equivalent entity
		// is treated as whitespace-only, mirroring upstream where
		// `&nbsp;` decodes to U+00A0 and `\S` doesn't match. The
		// surrounding lines indented "wrong" against the strict child
		// expectation should NOT report — locks the entity-aware
		// scanLiteralIndents branch. ----
		{
			Code: `var x = (
  <>
    <Link>hello</Link>
    &nbsp;
    <span>x</span>
  </>
);
`,
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		// ---- Numeric NBSP entity variants. ----
		{
			Code: `var x = (
  <>
    <Link>hello</Link>
    &#160;
    <span>x</span>
  </>
);
`,
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		{
			Code: `var x = (
  <>
    <Link>hello</Link>
    &#xA0;
    <span>x</span>
  </>
);
`,
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		// ---- A JsxText line whose first non-whitespace content is a
		// non-NBSP entity (e.g. `&amp;`, `&copy;`) DOES report — those
		// decode to real `\S` chars. ----
		{
			Code: `var x = (
  <>
    <Link>a</Link>
    &amp;
    <span>b</span>
  </>
);
`,
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		// ---- ReturnStatement gate alignment: TS `as` cast on a multi-line
		// JSX argument with mismatched opening/closing column. Upstream's
		// `jsxUtil.isJSX(node.argument)` rejects TSAsExpression so the
		// ReturnStatement listener never enters; we mirror that with
		// `ast.SkipParentheses` (paren-only, not TS-wrapper-stripping).
		// Even though the closing `)` is over-indented relative to the
		// `return`, NO diagnostic fires. ----
		{
			Code: `function App() {
  return (
    <App>
      <Foo />
    </App>
    ) as React.ReactElement;
}
`,
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		// ---- Same, with `satisfies`. ----
		{
			Code: `function App() {
  return (
    <App>
      <Foo />
    </App>
    ) satisfies React.ReactElement;
}
`,
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
		// ---- Same, with non-null assertion `!`. ----
		{
			Code: `function App() {
  return (
    <App>
      <Foo />
    </App>
    )!;
}
`,
			Tsx:     true,
			Options: []interface{}{float64(2)},
		},
	}, []rule_tester.InvalidTestCase{
		// ---- JsxText + nested element drift. Multi-pass: pass 1 fixes
		// outer/inner JsxText against the ORIGINAL parent indent (8); pass
		// 2 re-evaluates with the new inner-`<div>` line indent (12) and
		// pushes inner JsxText another 4 cols. Upstream's single-pass
		// fixer stops after pass 1; rslint's multipass converges. ----
		{
			Code: "\n        <div>\n        bar <div>\n           bar\n           bar {foo}\n           bar </div>\n        </div>\n      ",
			Output: []string{
				"\n        <div>\n            bar <div>\n            bar\n            bar {foo}\n            bar </div>\n        </div>\n      ",
				"\n        <div>\n            bar <div>\n                bar\n                bar {foo}\n                bar </div>\n        </div>\n      ",
			},
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 12 space characters but found 8."},
				{MessageId: "wrongIndent", Message: "Expected indentation of 12 space characters but found 11."},
				{MessageId: "wrongIndent", Message: "Expected indentation of 12 space characters but found 11."},
				{MessageId: "wrongIndent", Message: "Expected indentation of 12 space characters but found 11."},
			},
		},
		// ---- Default 4-space, child under-indented ----
		{
			Code: "\n        <App>\n          <Foo />\n        </App>\n      ",
			Output: []string{
				"\n        <App>\n            <Foo />\n        </App>\n      ",
			},
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 12 space characters but found 10."},
			},
		},
		// ---- Fragment variant ----
		{
			Code: "\n        <App>\n          <></>\n        </App>\n      ",
			Output: []string{
				"\n        <App>\n            <></>\n        </App>\n      ",
			},
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 12 space characters but found 10."},
			},
		},
		{
			Code: "\n        <>\n          <Foo />\n        </>\n      ",
			Output: []string{
				"\n        <>\n            <Foo />\n        </>\n      ",
			},
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 12 space characters but found 10."},
			},
		},
		// ---- Over-indented child with 2-space option ----
		{
			Code: "\n        <App>\n            <Foo />\n        </App>\n      ",
			Output: []string{
				"\n        <App>\n          <Foo />\n        </App>\n      ",
			},
			Tsx:     true,
			Options: []interface{}{float64(2)},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 10 space characters but found 12."},
			},
		},
		// ---- Tab option, child gets one tab ----
		{
			Code: "\n        <App>\n            <Foo />\n        </App>\n      ",
			Output: []string{
				"\n        <App>\n\t<Foo />\n        </App>\n      ",
			},
			Tsx:     true,
			Options: []interface{}{"tab"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 1 tab character but found 0."},
			},
		},
		// ---- Closing tag mis-indented (return + closing — 2 errors) ----
		{
			Code: "\n        function App() {\n          return <App>\n            <Foo />\n                 </App>;\n        }\n      ",
			Output: []string{
				"\n        function App() {\n          return <App>\n            <Foo />\n          </App>;\n        }\n      ",
			},
			Tsx:     true,
			Options: []interface{}{float64(2)},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Line: 3, Message: "Expected indentation of 10 space characters but found 17."},
				{MessageId: "wrongIndent", Line: 5, Message: "Expected indentation of 10 space characters but found 17."},
			},
		},
		// ---- Parenthesized JSX with closing mis-indented ----
		{
			Code: "\n        function App() {\n          return (<App>\n            <Foo />\n            </App>);\n        }\n      ",
			Output: []string{
				"\n        function App() {\n          return (<App>\n            <Foo />\n          </App>);\n        }\n      ",
			},
			Tsx:     true,
			Options: []interface{}{float64(2)},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Line: 3, Message: "Expected indentation of 10 space characters but found 12."},
				{MessageId: "wrongIndent", Line: 5, Message: "Expected indentation of 10 space characters but found 12."},
			},
		},
		// ---- Upstream "only flags one of three lines" case (#608).
		// Pass 1 reports just `<App>` mis-indent (the comment in the
		// upstream source explains the diagnostic gap); pass 2 then
		// catches `<Foo />` and `</App>` against the now-shifted parent
		// — multipass converges to the fully aligned output. ----
		{
			Code: "\n        function App() {\n          return (\n        <App>\n          <Foo />\n        </App>\n          );\n        }\n      ",
			Output: []string{
				"\n        function App() {\n          return (\n            <App>\n          <Foo />\n        </App>\n          );\n        }\n      ",
				"\n        function App() {\n          return (\n            <App>\n              <Foo />\n            </App>\n          );\n        }\n      ",
			},
			Tsx:     true,
			Options: []interface{}{float64(2)},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 12 space characters but found 8."},
			},
		},
		// ---- Three-space child JsxExpression ----
		{
			Code: "\n        <App>\n           {test}\n        </App>\n      ",
			Output: []string{
				"\n        <App>\n            {test}\n        </App>\n      ",
			},
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 12 space characters but found 11."},
			},
		},
		// ---- Three-space deeply nested JsxExpression ----
		{
			Code: "\n        <App>\n            {options.map((option, index) => (\n                <option key={index} value={option.key}>\n                   {option.name}\n                </option>\n            ))}\n        </App>\n      ",
			Output: []string{
				"\n        <App>\n            {options.map((option, index) => (\n                <option key={index} value={option.key}>\n                    {option.name}\n                </option>\n            ))}\n        </App>\n      ",
			},
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 20 space characters but found 19."},
			},
		},
		// ---- Tab option with JsxExpression at col 0 ----
		{
			Code: "\n        <App>\n        {test}\n        </App>\n      ",
			Output: []string{
				"\n        <App>\n\t{test}\n        </App>\n      ",
			},
			Tsx:     true,
			Options: []interface{}{"tab"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 1 tab character but found 0."},
			},
		},
		// ---- Tab option, deeply nested JsxExpression ----
		{
			Code: "\n\t\t\t\t<App>\n\t\t\t\t\t{options.map((option, index) => (\n\t\t\t\t\t\t<option key={index} value={option.key}>\n\t\t\t\t\t\t{option.name}\n\t\t\t\t\t\t</option>\n\t\t\t\t\t))}\n\t\t\t\t</App>\n\t\t\t",
			Output: []string{
				"\n\t\t\t\t<App>\n\t\t\t\t\t{options.map((option, index) => (\n\t\t\t\t\t\t<option key={index} value={option.key}>\n\t\t\t\t\t\t\t{option.name}\n\t\t\t\t\t\t</option>\n\t\t\t\t\t))}\n\t\t\t\t</App>\n\t\t\t",
			},
			Tsx:     true,
			Options: []interface{}{"tab"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 7 tab characters but found 6."},
			},
		},
		// ---- Tab option, JsxText literal containing escape `\n` ----
		{
			Code: "\n\t\t\t\t<App>\n\n\t\t\t\t<Foo />\n\n\t\t\t\t</App>\n\t\t\t",
			Output: []string{
				"\n\t\t\t\t<App>\n\n\t\t\t\t\t<Foo />\n\n\t\t\t\t</App>\n\t\t\t",
			},
			Tsx:     true,
			Options: []interface{}{"tab"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 5 tab characters but found 4."},
			},
		},
		// ---- Array literal: 2nd element over-indented ----
		{
			Code: "\n        [\n          <div />,\n            <div />\n        ]\n      ",
			Output: []string{
				"\n        [\n          <div />,\n          <div />\n        ]\n      ",
			},
			Tsx:     true,
			Options: []interface{}{float64(2)},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 10 space characters but found 12."},
			},
		},
		{
			Code: "\n        [\n          <div />,\n            <></>\n        ]\n      ",
			Output: []string{
				"\n        [\n          <div />,\n          <></>\n        ]\n      ",
			},
			Tsx:     true,
			Options: []interface{}{float64(2)},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 10 space characters but found 12."},
			},
		},
		// ---- Tab option, blank line between siblings ----
		{
			Code: "\n        <App>\n\n         <Foo />\n\n        </App>\n      ",
			Output: []string{
				"\n        <App>\n\n\t<Foo />\n\n        </App>\n      ",
			},
			Tsx:     true,
			Options: []interface{}{"tab"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 1 tab character but found 0."},
			},
		},
		// ---- 2-space option, mixed tab/space child ----
		{
			Code: "\n        <App>\n\n        \t<Foo />\n\n        </App>\n      ",
			Output: []string{
				"\n        <App>\n\n          <Foo />\n\n        </App>\n      ",
			},
			Tsx:     true,
			Options: []interface{}{float64(2)},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 10 space characters but found 8."},
			},
		},
		// ---- Default 4-space, deeply nested array under-indented ----
		{
			Code: "\n        <div>\n            {\n                [\n                    <Foo />,\n                <Bar />\n                ]\n            }\n        </div>\n      ",
			Output: []string{
				"\n        <div>\n            {\n                [\n                    <Foo />,\n                    <Bar />\n                ]\n            }\n        </div>\n      ",
			},
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 20 space characters but found 16."},
			},
		},
		// ---- Default 4-space, deeply nested array under `&&` ----
		{
			Code: "\n        <div>\n            {foo &&\n                [\n                    <Foo />,\n                <Bar />\n                ]\n            }\n        </div>\n      ",
			Output: []string{
				"\n        <div>\n            {foo &&\n                [\n                    <Foo />,\n                    <Bar />\n                ]\n            }\n        </div>\n      ",
			},
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 20 space characters but found 16."},
			},
		},
		// ---- Multiline ternary — alternate at col 0 ----
		{
			Code: "\n        foo ?\n            <Foo /> :\n        <Bar />\n      ",
			Output: []string{
				"\n        foo ?\n            <Foo /> :\n            <Bar />\n      ",
			},
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 12 space characters but found 8."},
			},
		},
		{
			Code: "\n        foo ?\n            <Foo /> :\n        <></>\n      ",
			Output: []string{
				"\n        foo ?\n            <Foo /> :\n            <></>\n      ",
			},
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 12 space characters but found 8."},
			},
		},
		// ---- Multiline ternary — colon on its own line, alternate at col 0 ----
		{
			Code: "\n        foo ?\n            <Foo />\n        :\n        <Bar />\n      ",
			Output: []string{
				"\n        foo ?\n            <Foo />\n        :\n            <Bar />\n      ",
			},
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 12 space characters but found 8."},
			},
		},
		// ---- First expr on test line, colon at end → alternate gets too much ----
		{
			Code: "\n        foo ? <Foo /> :\n            <Bar />\n      ",
			Output: []string{
				"\n        foo ? <Foo /> :\n        <Bar />\n      ",
			},
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 8 space characters but found 12."},
			},
		},
		{
			Code: "\n        foo ?\n            <Foo />\n        :\n        <></>\n      ",
			Output: []string{
				"\n        foo ?\n            <Foo />\n        :\n            <></>\n      ",
			},
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 12 space characters but found 8."},
			},
		},
		// ---- Test on first line, colon on its own line, alternate over-indented ----
		{
			Code: "\n        foo ? <Foo />\n        :\n              <Bar />\n      ",
			Output: []string{
				"\n        foo ? <Foo />\n        :\n        <Bar />\n      ",
			},
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 8 space characters but found 14."},
			},
		},
		// ---- Parenthesized first expr, colon at end ----
		{
			Code: "\n        foo ? (\n            <Foo />\n        ) :\n        <Bar />\n      ",
			Output: []string{
				"\n        foo ? (\n            <Foo />\n        ) :\n            <Bar />\n      ",
			},
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 12 space characters but found 8."},
			},
		},
		{
			Code: "\n        foo ? (\n            <Foo />\n        ) :\n        <></>\n      ",
			Output: []string{
				"\n        foo ? (\n            <Foo />\n        ) :\n            <></>\n      ",
			},
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 12 space characters but found 8."},
			},
		},
		// ---- Parenthesized first expr, colon on own line ----
		{
			Code: "\n        foo ? (\n            <Foo />\n        )\n        :\n        <Bar />\n      ",
			Output: []string{
				"\n        foo ? (\n            <Foo />\n        )\n        :\n            <Bar />\n      ",
			},
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 12 space characters but found 8."},
			},
		},
		// ---- Parenthesized second expr, colon at end of consequent ----
		{
			Code: "\n        foo ?\n            <Foo /> : (\n            <Bar />\n            )\n      ",
			Output: []string{
				"\n        foo ?\n            <Foo /> : (\n                <Bar />\n            )\n      ",
			},
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 16 space characters but found 12."},
			},
		},
		{
			Code: "\n        foo ?\n            <Foo /> : (\n            <></>\n            )\n      ",
			Output: []string{
				"\n        foo ?\n            <Foo /> : (\n                <></>\n            )\n      ",
			},
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 16 space characters but found 12."},
			},
		},
		// ---- Parenthesized second expr, colon on own line ----
		{
			Code: "\n        foo ?\n            <Foo />\n        : (\n        <Bar />\n        )\n      ",
			Output: []string{
				"\n        foo ?\n            <Foo />\n        : (\n            <Bar />\n        )\n      ",
			},
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 12 space characters but found 8."},
			},
		},
		// ---- Parenthesized second expr, colon indented on its own line ----
		{
			Code: "\n        foo ?\n            <Foo />\n            : (\n            <Bar />\n            )\n      ",
			Output: []string{
				"\n        foo ?\n            <Foo />\n            : (\n                <Bar />\n            )\n      ",
			},
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 16 space characters but found 12."},
			},
		},
		{
			Code: "\n        foo ?\n            <Foo />\n            : (\n            <></>\n            )\n      ",
			Output: []string{
				"\n        foo ?\n            <Foo />\n            : (\n                <></>\n            )\n      ",
			},
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 16 space characters but found 12."},
			},
		},
		// ---- Both branches parenthesized, colon at end ----
		{
			Code: "\n        foo ? (\n        <Foo />\n        ) : (\n        <Bar />\n        )\n      ",
			Output: []string{
				"\n        foo ? (\n            <Foo />\n        ) : (\n            <Bar />\n        )\n      ",
			},
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 12 space characters but found 8."},
				{MessageId: "wrongIndent", Message: "Expected indentation of 12 space characters but found 8."},
			},
		},
		{
			Code: "\n        foo ? (\n        <></>\n        ) : (\n        <></>\n        )\n      ",
			Output: []string{
				"\n        foo ? (\n            <></>\n        ) : (\n            <></>\n        )\n      ",
			},
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 12 space characters but found 8."},
				{MessageId: "wrongIndent", Message: "Expected indentation of 12 space characters but found 8."},
			},
		},
		// ---- Both parenthesized, colon on own line ----
		{
			Code: "\n        foo ? (\n        <Foo />\n        )\n        : (\n        <Bar />\n        )\n      ",
			Output: []string{
				"\n        foo ? (\n            <Foo />\n        )\n        : (\n            <Bar />\n        )\n      ",
			},
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 12 space characters but found 8."},
				{MessageId: "wrongIndent", Message: "Expected indentation of 12 space characters but found 8."},
			},
		},
		{
			Code: "\n        foo ? (\n        <Foo />\n        )\n        :\n        (\n        <Bar />\n        )\n      ",
			Output: []string{
				"\n        foo ? (\n            <Foo />\n        )\n        :\n        (\n            <Bar />\n        )\n      ",
			},
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 12 space characters but found 8."},
				{MessageId: "wrongIndent", Message: "Expected indentation of 12 space characters but found 8."},
			},
		},
		{
			Code: "\n        foo ? (\n        <></>\n        )\n        :\n        (\n        <></>\n        )\n      ",
			Output: []string{
				"\n        foo ? (\n            <></>\n        )\n        :\n        (\n            <></>\n        )\n      ",
			},
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 12 space characters but found 8."},
				{MessageId: "wrongIndent", Message: "Expected indentation of 12 space characters but found 8."},
			},
		},
		// ---- Test on first line, colon at end, parenthesized second expr ----
		{
			Code: "\n        foo ? <Foo /> : (\n        <Bar />\n        )\n      ",
			Output: []string{
				"\n        foo ? <Foo /> : (\n            <Bar />\n        )\n      ",
			},
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 12 space characters but found 8."},
			},
		},
		{
			Code: "\n        foo ? <Foo /> : (\n        <></>\n        )\n      ",
			Output: []string{
				"\n        foo ? <Foo /> : (\n            <></>\n        )\n      ",
			},
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 12 space characters but found 8."},
			},
		},
		// ---- Test on first line, colon on its own line, parenthesized second expr ----
		{
			Code: "\n        foo ? <Foo />\n        : (\n        <Bar />\n        )\n      ",
			Output: []string{
				"\n        foo ? <Foo />\n        : (\n            <Bar />\n        )\n      ",
			},
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 12 space characters but found 8."},
			},
		},
		{
			Code: "\n        foo ? <Foo />\n        : (\n        <></>\n        )\n      ",
			Output: []string{
				"\n        foo ? <Foo />\n        : (\n            <></>\n        )\n      ",
			},
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 12 space characters but found 8."},
			},
		},
		// ---- Inline closing `</div>` after content; nested over-indented closing ----
		{
			Code: "\n        <p>\n            <div>\n                <SelfClosingTag />Text\n          </div>\n        </p>\n      ",
			Output: []string{
				"\n        <p>\n            <div>\n                <SelfClosingTag />Text\n            </div>\n        </p>\n      ",
			},
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 12 space characters but found 10."},
			},
		},
		// ---- checkAttributes:true — `)}` realigned to attribute name ----
		{
			Code: "\n        const Component = () => (\n          <View\n            ListFooterComponent={(\n              <View\n                rowSpan={3}\n                placeholder=\"placeholder text here\"\n              />\n        )}\n          />\n        );\n      ",
			Output: []string{
				"\n        const Component = () => (\n          <View\n            ListFooterComponent={(\n              <View\n                rowSpan={3}\n                placeholder=\"placeholder text here\"\n              />\n            )}\n          />\n        );\n      ",
			},
			Tsx:     true,
			Options: []interface{}{float64(2), map[string]interface{}{"checkAttributes": true}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 12 space characters but found 8."},
			},
		},
		// ---- checkAttributes:true with tab option ----
		{
			Code: "\nconst Component = () => (\n\t<View\n\t\tListFooterComponent={(\n\t\t\t<View\n\t\t\t\trowSpan={3}\n\t\t\t\tplaceholder=\"placeholder text here\"\n\t\t\t/>\n)}\n\t/>\n);\n    ",
			Output: []string{
				"\nconst Component = () => (\n\t<View\n\t\tListFooterComponent={(\n\t\t\t<View\n\t\t\t\trowSpan={3}\n\t\t\t\tplaceholder=\"placeholder text here\"\n\t\t\t/>\n\t\t)}\n\t/>\n);\n    ",
			},
			Tsx:     true,
			Options: []interface{}{"tab", map[string]interface{}{"checkAttributes": true}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 2 tab characters but found 0."},
			},
		},
		// ---- indentLogicalExpressions:true — inner JSX must indent past `&&` ----
		{
			Code: "\n        function Foo() {\n          return (\n            <div>\n              {condition && (\n              <p>Bar</p>\n              )}\n            </div>\n          );\n        }\n      ",
			Output: []string{
				"\n        function Foo() {\n          return (\n            <div>\n              {condition && (\n                <p>Bar</p>\n              )}\n            </div>\n          );\n        }\n      ",
			},
			Tsx:     true,
			Options: []interface{}{float64(2), map[string]interface{}{"indentLogicalExpressions": true}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 16 space characters but found 14."},
			},
		},
		// ---- do-expression invalid cases (Skip — see valid skip block) ----
		{
			Code: "\n        <span>\n            {do {\n                const num = rollDice();\n                    <Thing num={num} />;\n            }}\n        </span>\n      ",
			Tsx:  true,
			Skip: true, // SKIP: rslint's tsgo parser does not support do-expressions.
		},
		{
			Code: "\n        <span>\n            {(do {\n                const num = rollDice();\n                    <Thing num={num} />;\n            })}\n        </span>\n      ",
			Tsx:  true,
			Skip: true, // SKIP: rslint's tsgo parser does not support do-expressions.
		},
		{
			Code: "\n        <span>\n            {do {\n            <Thing num={getPurposeOfLife()} />;\n            }}\n        </span>\n      ",
			Tsx:  true,
			Skip: true, // SKIP: rslint's tsgo parser does not support do-expressions.
		},
		{
			Code: "\n        <span>\n            {(do {\n            <Thing num={getPurposeOfLife()} />;\n            })}\n        </span>\n      ",
			Tsx:  true,
			Skip: true, // SKIP: rslint's tsgo parser does not support do-expressions.
		},
		// ---- JsxText fixer — single line of text mis-indented ----
		{
			Code: "\n        <div>\n        text\n        </div>\n      ",
			Output: []string{
				"\n        <div>\n            text\n        </div>\n      ",
			},
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 12 space characters but found 8."},
			},
		},
		// ---- JsxText fixer — multi-line text, all lines mis-indented ----
		{
			Code: "\n        <div>\n          text\n        text\n        </div>\n      ",
			Output: []string{
				"\n        <div>\n            text\n            text\n        </div>\n      ",
			},
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 12 space characters but found 10."},
				{MessageId: "wrongIndent", Message: "Expected indentation of 12 space characters but found 8."},
			},
		},
		// ---- JsxText fixer — mixed tab + space ----
		{
			Code: "\n        <div>\n        \t  text\n          \t  text\n        </div>\n      ",
			Output: []string{
				"\n        <div>\n            text\n            text\n        </div>\n      ",
			},
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 12 space characters but found 8."},
				{MessageId: "wrongIndent", Message: "Expected indentation of 12 space characters but found 10."},
			},
		},
		// ---- Tab option, JsxText with tab characters ----
		{
			Code: "\n        <div>\n        \t\ttext\n        </div>\n      ",
			Output: []string{
				"\n        <div>\n\ttext\n        </div>\n      ",
			},
			Tsx:     true,
			Options: []interface{}{"tab"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 1 tab character but found 0."},
			},
		},
		// ---- Fragment JsxText mis-indented ----
		{
			Code: "\n        <>\n        aaa\n        </>\n      ",
			Output: []string{
				"\n        <>\n            aaa\n        </>\n      ",
			},
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 12 space characters but found 8."},
			},
		},
		// ---- Stateless component, return inside if(...) ----
		{
			Code: "\n        const StatelessComponent = () => {\n          if (new Date() % 2) {\n              return (\n        <div>Hello</div>\n              );\n          }\n          return null;\n        };\n      ",
			Output: []string{
				"\n        const StatelessComponent = () => {\n          if (new Date() % 2) {\n              return (\n                  <div>Hello</div>\n              );\n          }\n          return null;\n        };\n      ",
			},
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 18 space characters but found 8."},
			},
		},
		// ---- Closing paren of return mis-indented (message text only) ----
		{
			Code: "\n        function App() {\n          return (\n            <App />\n            );\n        }\n      ",
			Output: []string{
				"\n        function App() {\n          return (\n            <App />\n          );\n        }\n      ",
			},
			Tsx:     true,
			Options: []interface{}{float64(2)},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 10 space characters but found 12."},
			},
		},
		{
			Code: "\n        function App() {\n          return (\n            <App />\n        );\n        }\n      ",
			Output: []string{
				"\n        function App() {\n          return (\n            <App />\n          );\n        }\n      ",
			},
			Tsx:     true,
			Options: []interface{}{float64(2)},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 10 space characters but found 8."},
			},
		},
		// ---- Array of JSX with arrow-bodied callbacks ----
		{
			Code: "\n        {condition && [\n            <Tag key=\"a\" onClick={() => {\n              // some code\n            }} />,\n            <Tag key=\"b\" onClick={() => {\n              // some code\n            }} />,\n          ]\n        }\n      ",
			Output: []string{
				"\n        {condition && [\n          <Tag key=\"a\" onClick={() => {\n              // some code\n            }} />,\n          <Tag key=\"b\" onClick={() => {\n              // some code\n            }} />,\n          ]\n        }\n      ",
			},
			Tsx:     true,
			Options: []interface{}{float64(2)},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Line: 3, Message: "Expected indentation of 10 space characters but found 12."},
				{MessageId: "wrongIndent", Line: 6, Message: "Expected indentation of 10 space characters but found 12."},
			},
		},
		// ---- Mixed JsxExpression / JsxText / closing ----
		{
			Code: "\n        const IndexPage = () => (\n          <h1>\n        {\"Hi people\"}\n        <button/>\n        </h1>\n        );\n      ",
			Output: []string{
				"\n        const IndexPage = () => (\n          <h1>\n            {\"Hi people\"}\n            <button/>\n          </h1>\n        );\n      ",
			},
			Tsx:     true,
			Options: []interface{}{float64(2)},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 12 space characters but found 8."},
				{MessageId: "wrongIndent", Message: "Expected indentation of 12 space characters but found 8."},
				{MessageId: "wrongIndent", Message: "Expected indentation of 10 space characters but found 8."},
			},
		},
		// ---- Multipass: mixed JsxText / JSX / closing — JsxText fix
		// range overlaps with `<button/>` indent fix, so pass 1 only
		// applies the JsxText + closing-tag fixes; pass 2 fixes
		// `<button/>` once the overlap is gone. ----
		{
			Code: "\n        const IndexPage = () => (\n          <h1>\n        Hi people\n        <button/>\n        </h1>\n        );\n      ",
			Output: []string{
				"\n        const IndexPage = () => (\n          <h1>\n            Hi people\n        <button/>\n          </h1>\n        );\n      ",
				"\n        const IndexPage = () => (\n          <h1>\n            Hi people\n            <button/>\n          </h1>\n        );\n      ",
			},
			Tsx:     true,
			Options: []interface{}{float64(2)},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 12 space characters but found 8."},
				{MessageId: "wrongIndent", Message: "Expected indentation of 12 space characters but found 8."},
				{MessageId: "wrongIndent", Message: "Expected indentation of 10 space characters but found 8."},
			},
		},
		// ---- Followup pass: only the JsxElement child remains mis-indented ----
		{
			Code: "\n        const IndexPage = () => (\n          <h1>\n            Hi people\n        <button/>\n          </h1>\n        );\n      ",
			Output: []string{
				"\n        const IndexPage = () => (\n          <h1>\n            Hi people\n            <button/>\n          </h1>\n        );\n      ",
			},
			Tsx:     true,
			Options: []interface{}{float64(2)},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 12 space characters but found 8."},
			},
		},
		// ---- import + multi-line indent (4-space option). Same overlap
		// pattern as invalid-61: pass 1 fixes JsxText 1 + closing
		// peer-anchor; pass 2 fixes `<p>` once the JsxText fix range
		// no longer covers its line. Upstream's single-pass output is
		// our pass 1; rslint's multipass yields the converged result. ----
		{
			Code: "\n        import React from 'react';\n\n        export default function () {\n            return (\n                <div>\n                            Test1\n\n                      <p>Test2</p>\n                </div>\n            );\n        }\n      ",
			Output: []string{
				"\n        import React from 'react';\n\n        export default function () {\n            return (\n                <div>\n                    Test1\n\n                      <p>Test2</p>\n                </div>\n            );\n        }\n      ",
				"\n        import React from 'react';\n\n        export default function () {\n            return (\n                <div>\n                    Test1\n\n                    <p>Test2</p>\n                </div>\n            );\n        }\n      ",
			},
			Tsx:     true,
			Options: []interface{}{float64(4)},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent"},
				{MessageId: "wrongIndent"},
			},
		},

		// =================================================================
		// Below: invalid-side coverage of the real-world / edge shapes.
		// =================================================================

		// ---- TS `as` cast: closing tag mis-indented. The wrapper must
		// be transparent so the closing tag still reports against the
		// parent's opening. ----
		{
			Code: `var x = (
  <App>
    <Foo />
      </App>
) as React.ReactElement;
`,
			Output: []string{`var x = (
  <App>
    <Foo />
  </App>
) as React.ReactElement;
`},
			Tsx:     true,
			Options: []interface{}{float64(2)},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 2 space characters but found 6."},
			},
		},
		// ---- TS `satisfies` wrapper: child mis-indented. ----
		{
			Code: `var x = (
  <App>
      <Foo />
  </App>
) satisfies React.ReactElement;
`,
			Output: []string{`var x = (
  <App>
    <Foo />
  </App>
) satisfies React.ReactElement;
`},
			Tsx:     true,
			Options: []interface{}{float64(2)},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 4 space characters but found 6."},
			},
		},
		// ---- Generic JSX with mis-indented child (`<Foo<T>><Bar />`). ----
		{
			Code: `var x = (
  <Foo<string>>
      <Bar />
  </Foo>
);
`,
			Output: []string{`var x = (
  <Foo<string>>
    <Bar />
  </Foo>
);
`},
			Tsx:     true,
			Options: []interface{}{float64(2)},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 4 space characters but found 6."},
			},
		},
		// ---- JSX member-expression tag with mis-indented child. ----
		{
			Code: `var x = (
  <Foo.Bar>
        <Foo.Baz />
  </Foo.Bar>
);
`,
			Output: []string{`var x = (
  <Foo.Bar>
    <Foo.Baz />
  </Foo.Bar>
);
`},
			Tsx:     true,
			Options: []interface{}{float64(2)},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 4 space characters but found 8."},
			},
		},
		// ---- Namespaced JSX with mis-indented closing. ----
		{
			Code: `var x = (
  <svg:svg>
    <svg:rect />
      </svg:svg>
);
`,
			Output: []string{`var x = (
  <svg:svg>
    <svg:rect />
  </svg:svg>
);
`},
			Tsx:     true,
			Options: []interface{}{float64(2)},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 2 space characters but found 6."},
			},
		},
		// ---- Deep nesting (5 levels) with one bad child mid-tree. The
		// rule must report exactly that one mis-indent — not propagate
		// to ancestors / descendants. ----
		{
			Code: `var x = (
  <A>
    <B>
      <C>
            <D>
              <E />
            </D>
      </C>
    </B>
  </A>
);
`,
			Output: []string{`var x = (
  <A>
    <B>
      <C>
        <D>
              <E />
            </D>
      </C>
    </B>
  </A>
);
`,
				`var x = (
  <A>
    <B>
      <C>
        <D>
          <E />
        </D>
      </C>
    </B>
  </A>
);
`},
			Tsx:     true,
			Options: []interface{}{float64(2)},
			Errors: []rule_tester.InvalidTestCaseError{
				// Pass 1 sees only `<D>` over-indented at 12 (expected 8)
				// — its closing tag still aligns with the (wrong) opening
				// and `<E />` aligns one level deeper, so they're locally
				// consistent. After pass 1 moves `<D>` to col 8, pass 2
				// re-evaluates the children against the new parent indent
				// and converges.
				{MessageId: "wrongIndent", Message: "Expected indentation of 8 space characters but found 12."},
			},
		},
		// ---- React.memo wrap with mis-indented child. ----
		{
			Code: `var Wrapped = React.memo(() => (
  <App>
        <Foo />
  </App>
));
`,
			Output: []string{`var Wrapped = React.memo(() => (
  <App>
    <Foo />
  </App>
));
`},
			Tsx:     true,
			Options: []interface{}{float64(2)},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 4 space characters but found 8."},
			},
		},
		// ---- forwardRef wrap with mis-indented closing. ----
		{
			Code: `var Wrapped = React.forwardRef((props, ref) => (
  <App ref={ref}>
    <Foo />
        </App>
));
`,
			Output: []string{`var Wrapped = React.forwardRef((props, ref) => (
  <App ref={ref}>
    <Foo />
  </App>
));
`},
			Tsx:     true,
			Options: []interface{}{float64(2)},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 2 space characters but found 8."},
			},
		},
		// ---- map() callback returning JSX, mis-indented child. ----
		{
			Code: `var els = items.map((item, i) => (
  <Item key={i}>
      {item.name}
  </Item>
));
`,
			Output: []string{`var els = items.map((item, i) => (
  <Item key={i}>
    {item.name}
  </Item>
));
`},
			Tsx:     true,
			Options: []interface{}{float64(2)},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 4 space characters but found 6."},
			},
		},
		// ---- CRLF line endings, mis-indented child. The rule must
		// recognize `\r\n` so it locates the line start correctly. ----
		{
			Code: "var x = (\r\n  <App>\r\n      <Foo />\r\n  </App>\r\n);\r\n",
			Output: []string{
				"var x = (\r\n  <App>\r\n    <Foo />\r\n  </App>\r\n);\r\n",
			},
			Tsx:     true,
			Options: []interface{}{float64(2)},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 4 space characters but found 6."},
			},
		},
		// ---- HTML-entity content with wrong line indent. ----
		{
			Code: `var x = (
  <App>
        Hello&nbsp;World
  </App>
);
`,
			Output: []string{`var x = (
  <App>
    Hello&nbsp;World
  </App>
);
`},
			Tsx:     true,
			Options: []interface{}{float64(2)},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 4 space characters but found 8."},
			},
		},
		// ---- CJK content with wrong indent. Locks UTF-8 byte
		// arithmetic in lineIndentAt: leading whitespace counting must
		// not get confused by multi-byte runes elsewhere on the line. ----
		{
			Code: `var x = (
  <App>
        你好世界
  </App>
);
`,
			Output: []string{`var x = (
  <App>
    你好世界
  </App>
);
`},
			Tsx:     true,
			Options: []interface{}{float64(2)},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 4 space characters but found 8."},
			},
		},
		// ---- Sibling-fragment-then-element pair, second child mis-indented. ----
		{
			Code: `var x = (
  <App>
    <>
      <span />
    </>
        <span />
  </App>
);
`,
			Output: []string{`var x = (
  <App>
    <>
      <span />
    </>
    <span />
  </App>
);
`},
			Tsx:     true,
			Options: []interface{}{float64(2)},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 4 space characters but found 8."},
			},
		},
		// ---- Spread attribute is mis-aligned and has nothing to do
		// with this rule (the rule only checks attribute *value*
		// indent under checkAttributes). Still no diagnostic. ----
		{
			Code: `var x = (
  <App
    {...props}
    foo="bar"
  >
        <Foo />
  </App>
);
`,
			Output: []string{`var x = (
  <App
    {...props}
    foo="bar"
  >
    <Foo />
  </App>
);
`},
			Tsx:     true,
			Options: []interface{}{float64(2)},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 4 space characters but found 8."},
			},
		},
		// ---- async arrow returning JSX with mis-indented child. ----
		{
			Code: `var f = async () => (
  <App>
        <Foo />
  </App>
);
`,
			Output: []string{`var f = async () => (
  <App>
    <Foo />
  </App>
);
`},
			Tsx:     true,
			Options: []interface{}{float64(2)},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 4 space characters but found 8."},
			},
		},
		// ---- Provider pattern with mis-indented inner child. ----
		{
			Code: `var x = (
  <ThemeContext.Provider value={theme}>
    <App>
        <Foo />
    </App>
  </ThemeContext.Provider>
);
`,
			Output: []string{`var x = (
  <ThemeContext.Provider value={theme}>
    <App>
      <Foo />
    </App>
  </ThemeContext.Provider>
);
`},
			Tsx:     true,
			Options: []interface{}{float64(2)},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 6 space characters but found 8."},
			},
		},
		// ---- compose / HOC double-call with mis-indented child. ----
		{
			Code: `var Wrapped = compose(withRouter, connect(mapState))(({ items }) => (
  <App>
        <Foo />
  </App>
));
`,
			Output: []string{`var Wrapped = compose(withRouter, connect(mapState))(({ items }) => (
  <App>
    <Foo />
  </App>
));
`},
			Tsx:     true,
			Options: []interface{}{float64(2)},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 4 space characters but found 8."},
			},
		},
		// ---- JSX in object literal value with mis-indented closing. ----
		{
			Code: `var x = {
  child: (
    <App>
      <Foo />
        </App>
  ),
};
`,
			Output: []string{`var x = {
  child: (
    <App>
      <Foo />
    </App>
  ),
};
`},
			Tsx:     true,
			Options: []interface{}{float64(2)},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 4 space characters but found 8."},
			},
		},
		// ---- React.lazy + Suspense with mis-indented child. ----
		{
			Code: `var x = (
  <Suspense fallback={<Spinner />}>
        <Page />
  </Suspense>
);
`,
			Output: []string{`var x = (
  <Suspense fallback={<Spinner />}>
    <Page />
  </Suspense>
);
`},
			Tsx:     true,
			Options: []interface{}{float64(2)},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 4 space characters but found 8."},
			},
		},
		// ---- Hooks pattern: function returning multi-line JSX, child
		// mis-indented. Locks ReturnStatement listener vs jsxParentOpening
		// interaction. ----
		{
			Code: `function Counter() {
  const [n, setN] = useState(0);
  return (
    <div>
        <span>{n}</span>
      <button onClick={() => setN(n + 1)}>+</button>
    </div>
  );
}
`,
			Output: []string{`function Counter() {
  const [n, setN] = useState(0);
  return (
    <div>
      <span>{n}</span>
      <button onClick={() => setN(n + 1)}>+</button>
    </div>
  );
}
`},
			Tsx:     true,
			Options: []interface{}{float64(2)},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 6 space characters but found 8."},
			},
		},
		// ---- switch case returning multi-line JSX, child mis-indented. ----
		{
			Code: `function pick(kind) {
  switch (kind) {
    case 'a':
      return (
        <A>
            <Foo />
        </A>
      );
    default:
      return null;
  }
}
`,
			Output: []string{`function pick(kind) {
  switch (kind) {
    case 'a':
      return (
        <A>
          <Foo />
        </A>
      );
    default:
      return null;
  }
}
`},
			Tsx:     true,
			Options: []interface{}{float64(2)},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 10 space characters but found 12."},
			},
		},
		// ---- Tab option with multi-line JSX, child mis-indented. ----
		{
			Code:    "var x = (\n\t<App>\n\t\t\t<Foo />\n\t</App>\n);\n",
			Output:  []string{"var x = (\n\t<App>\n\t\t<Foo />\n\t</App>\n);\n"},
			Tsx:     true,
			Options: []interface{}{"tab"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 2 tab characters but found 3."},
			},
		},
		// ---- Block-comment between JSX siblings, second sibling
		// mis-indented. The comment shouldn't affect anchor logic. ----
		{
			Code: `var x = (
  <App>
    {/* block comment */}
        <Foo />
  </App>
);
`,
			Output: []string{`var x = (
  <App>
    {/* block comment */}
    <Foo />
  </App>
);
`},
			Tsx:     true,
			Options: []interface{}{float64(2)},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 4 space characters but found 8."},
			},
		},
		// ---- CRLF JsxText fix: the fixer must preserve `\r` carriage
		// returns. Locks `replaceLeadingIndentInText` against the
		// regression of stripping `\r` while it consumes leading
		// whitespace. ----
		{
			Code:   "var x = (\r\n  <App>\r\nHello\r\n  </App>\r\n);\r\n",
			Output: []string{"var x = (\r\n  <App>\r\n    Hello\r\n  </App>\r\n);\r\n"},
			Tsx:    true,
			Options: []interface{}{float64(2)},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 4 space characters but found 0."},
			},
		},

		// =================================================================
		// Position-assertion suite. Each entry below pins
		// Line + Column + EndLine + EndColumn so future range / Pos
		// changes get caught. Containers covered: JsxOpeningElement,
		// JsxClosingElement, JsxExpression, JsxText (multi-line range),
		// ReturnStatement (multi-line range), JsxAttribute (custom
		// anchor range).
		// =================================================================

		// ---- JsxOpeningElement at exact line+col, single-line node. ----
		{
			Code: "<App>\n  <Foo />\n</App>",
			Output: []string{
				"<App>\n    <Foo />\n</App>",
			},
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "wrongIndent",
					Message:   "Expected indentation of 4 space characters but found 2.",
					Line:      2, Column: 3, EndLine: 2, EndColumn: 10,
				},
			},
		},
		// ---- JsxClosingElement at exact line+col. ----
		{
			Code: "<App>\n    <Foo />\n     </App>",
			Output: []string{
				"<App>\n    <Foo />\n</App>",
			},
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "wrongIndent",
					Message:   "Expected indentation of 0 space characters but found 5.",
					Line:      3, Column: 6, EndLine: 3, EndColumn: 12,
				},
			},
		},
		// ---- JsxExpression `{...}` exact range. ----
		{
			Code: "<App>\n  {value}\n</App>",
			Output: []string{
				"<App>\n    {value}\n</App>",
			},
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "wrongIndent",
					Message:   "Expected indentation of 4 space characters but found 2.",
					Line:      2, Column: 3, EndLine: 2, EndColumn: 10,
				},
			},
		},
		// ---- ReturnStatement multi-line: opening on its line, closing
		// on a different line. The diagnostic range spans the whole
		// statement (multi-line range). ----
		{
			Code: `function App() {
  return <App>
    <Foo />
       </App>;
}`,
			Output: []string{
				`function App() {
  return <App>
    <Foo />
  </App>;
}`,
			},
			Tsx:     true,
			Options: []interface{}{float64(2)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "wrongIndent",
					Message:   "Expected indentation of 2 space characters but found 7.",
					Line:      2, Column: 3, EndLine: 4, EndColumn: 15,
				},
				{
					MessageId: "wrongIndent",
					Message:   "Expected indentation of 2 space characters but found 7.",
					Line:      4, Column: 8, EndLine: 4, EndColumn: 14,
				},
			},
		},
		// ---- JsxText multi-line raw range. The diagnostic anchors at
		// the JsxText raw start (right after the parent's `>`) and
		// ends at the next sibling's start. ----
		{
			Code: `<div>
text
text
</div>`,
			Output: []string{
				`<div>
    text
    text
</div>`,
			},
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "wrongIndent",
					Message:   "Expected indentation of 4 space characters but found 0.",
					// JsxText raw range: from `>` of `<div>` end to `<` of `</div>`.
					Line: 1, Column: 6, EndLine: 4, EndColumn: 1,
				},
				{
					MessageId: "wrongIndent",
					Message:   "Expected indentation of 4 space characters but found 0.",
					Line:      1, Column: 6, EndLine: 4, EndColumn: 1,
				},
			},
		},
		// ---- Triple-nested ternary, deepest alternate on its OWN line at
		// wrong indent. (Same line as `:` would NOT report — upstream's
		// `isNodeFirstInLine` excludes the JSX from indent checking when
		// the colon precedes it on the same line.) ----
		{
			Code: `var x = (
  a ?
    <A />
    : b ?
      <B />
      : c ?
        <C />
        :
            <D />
);
`,
			Output: []string{`var x = (
  a ?
    <A />
    : b ?
      <B />
      : c ?
        <C />
        :
        <D />
);
`},
			Tsx:     true,
			Options: []interface{}{float64(2)},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 8 space characters but found 12."},
			},
		},
		// ---- indentLogicalExpressions:true on nested logical chain,
		// inner JSX mis-indented. ----
		{
			Code: `var x = (
  <App>
    {a && b && (
    <p>both</p>
    )}
  </App>
);
`,
			Output: []string{`var x = (
  <App>
    {a && b && (
      <p>both</p>
    )}
  </App>
);
`},
			Tsx:     true,
			Options: []interface{}{float64(2), map[string]interface{}{"indentLogicalExpressions": true}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 6 space characters but found 4."},
			},
		},
		// ---- checkAttributes:true with mis-indented `)}` after a
		// multi-line attribute-value JSX. ----
		{
			Code: `const x = (
  <View
    Foo={(
      <Inner />
)}
  />
);
`,
			Output: []string{`const x = (
  <View
    Foo={(
      <Inner />
    )}
  />
);
`},
			Tsx:     true,
			Options: []interface{}{float64(2), map[string]interface{}{"checkAttributes": true}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 4 space characters but found 0."},
			},
		},
		// ---- JsxOpeningFragment mis-indented relative to its
		// container's opening anchor. Pass 1 fixes `<>` against `<App>`
		// (now col 4) AND `<Foo />` / `</>` against the *original* `<>`
		// line indent (8) — they shift to 10 / 8. Pass 2 then
		// re-evaluates against the moved `<>` (line indent 4) and
		// converges on the final layout (6 / 4). ----
		{
			Code: `var x = (
  <App>
        <>
      <Foo />
    </>
  </App>
);
`,
			Output: []string{
				"var x = (\n  <App>\n    <>\n          <Foo />\n        </>\n  </App>\n);\n",
				"var x = (\n  <App>\n    <>\n      <Foo />\n    </>\n  </App>\n);\n",
			},
			Tsx:     true,
			Options: []interface{}{float64(2)},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongIndent", Message: "Expected indentation of 4 space characters but found 8."},
				{MessageId: "wrongIndent", Message: "Expected indentation of 10 space characters but found 6."},
				{MessageId: "wrongIndent", Message: "Expected indentation of 8 space characters but found 4."},
			},
		},
	})
}

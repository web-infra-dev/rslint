// cspell:ignore renderder

package no_render_return_value

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoRenderReturnValueRule(t *testing.T) {
	// Shared settings fixtures mirroring upstream's `settings: { react: { version: '…' } }`.
	react0140 := map[string]interface{}{"react": map[string]interface{}{"version": "0.14.0"}}
	react0130 := map[string]interface{}{"react": map[string]interface{}{"version": "0.13.0"}}
	react0001 := map[string]interface{}{"react": map[string]interface{}{"version": "0.0.1"}}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoRenderReturnValueRule, []rule_tester.ValidTestCase{
		// ---- Upstream: render with no return-value consumption ----
		{Code: `ReactDOM.render(<div />, document.body);`, Tsx: true},
		{Code: `
        let node;
        ReactDOM.render(<div ref={ref => node = ref}/>, document.body);
      `, Tsx: true},

		// ---- Upstream: version-gated callee-object matching ----
		{Code: `ReactDOM.render(<div ref={ref => this.node = ref}/>, document.body);`, Tsx: true, Settings: react0140},
		{Code: `React.render(<div ref={ref => this.node = ref}/>, document.body);`, Tsx: true, Settings: react0140},
		{Code: `React.render(<div ref={ref => this.node = ref}/>, document.body);`, Tsx: true, Settings: react0130},

		// ---- Upstream: version 0.0.1 falls back to default (ReactDOM-only) — React doesn't match ----
		{Code: `var foo = React.render(<div />, root);`, Tsx: true, Settings: react0001},

		// ---- Upstream: bare `render(...)` — callee not a MemberExpression ----
		{Code: `var foo = render(<div />, root)`, Tsx: true},

		// ---- Upstream: lowercase `ReactDom.renderder` — regex is case-sensitive, neither name nor property match ----
		{Code: `var foo = ReactDom.renderder(<div />, root)`, Tsx: true},

		// ---- Edge: bracket access with a string-literal key — upstream's `'name' in property` guard is false for Literal nodes (no `.name`), so no fire ----
		{Code: `var x = ReactDOM['render'](<div />, document.body);`, Tsx: true},

		// ---- Edge: bracket access with a numeric-literal key — same as above ----
		{Code: `var x = ReactDOM[0](<div />, document.body);`, Tsx: true},

		// ---- Edge: bracket access with an Identifier named NOT "render" — `property.name !== 'render'` → no fire ----
		{Code: `var renderFn; var x = ReactDOM[renderFn](<div />, document.body);`, Tsx: true},

		// ---- Edge: bracket access with a template-literal / computed expression — not an Identifier, no `.name`, no fire ----
		{Code: "var x = ReactDOM[`render`](<div />, document.body);", Tsx: true},

		// ---- Edge: `React.render` at default version (>= 15.0.0) — object must be `ReactDOM`, not `React` ----
		{Code: `var x = React.render(<div />, document.body);`, Tsx: true},

		// ---- Edge: call as a non-return-consuming argument position ----
		{Code: `doSomething(ReactDOM.render(<div />, document.body));`, Tsx: true},

		// ---- Edge: conditional expression — not in the allow-list ----
		{Code: `var x = cond ? ReactDOM.render(<div />, document.body) : null;`, Tsx: true},

		// ---- Edge: logical-operand position — not in the allow-list ----
		{Code: `var x = y || ReactDOM.render(<div />, document.body);`, Tsx: true},

		// ---- Edge: nullish-coalesce operand ----
		{Code: `var x = y ?? ReactDOM.render(<div />, document.body);`, Tsx: true},

		// ---- Edge: chained member access — callee.object is a MemberExpression, not an Identifier ----
		{Code: `var x = obj.ReactDOM.render(<div />, document.body);`, Tsx: true},

		// ---- Edge: `new ReactDOM.render()` — a NewExpression, not a CallExpression — listener doesn't fire ----
		{Code: `var x = new ReactDOM.render(<div />, document.body);`, Tsx: true},

		// ---- Edge: `await` / `yield` / `throw` / `typeof` — none of them are in the allow-list ----
		{Code: `async function f() { await ReactDOM.render(<div />, document.body); }`, Tsx: true},
		{Code: `function* g() { yield ReactDOM.render(<div />, document.body); }`, Tsx: true},
		{Code: `function h() { throw ReactDOM.render(<div />, document.body); }`, Tsx: true},
		{Code: `var x = typeof ReactDOM.render(<div />, document.body);`, Tsx: true},

		// ---- Edge: array / spread / template — consumed into a container, not via the allow-list ----
		{Code: `var xs = [ReactDOM.render(<div />, document.body)];`, Tsx: true},
		{Code: `var o = { ...ReactDOM.render(<div />, document.body) };`, Tsx: true},
		{Code: `var s = ` + "`${ReactDOM.render(<div />, document.body)}`" + `;`, Tsx: true},

		// ---- Edge: method-chained from the call — `.then` / `.foo` parent is a PropertyAccess ----
		{Code: `ReactDOM.render(<div />, document.body).then(inst => inst);`, Tsx: true},
		{Code: `var x = ReactDOM.render(<div />, document.body).foo;`, Tsx: true},

		// ---- Edge: SequenceExpression (comma) — only the last operand is consumed by its parent ----
		{Code: `var x = (foo(), ReactDOM.render(<div />, document.body), 1);`, Tsx: true},

		// ---- Edge: TS type assertion / non-null — wrapped in AsExpression / NonNullExpression; matches upstream's parent-type guard miss ----
		{Code: `var x = ReactDOM.render(<div />, document.body) as any;`, Tsx: true},
		{Code: `var x = ReactDOM.render(<div />, document.body)!;`, Tsx: true},

		// ---- Edge: shadowed `ReactDOM` identifier still matches — rule is name-based, not scope-aware (matches upstream) ----
		// Sibling valid: a top-level call with no consumption still doesn't fire.
		{Code: `{ let ReactDOM; ReactDOM.render(<div />, document.body); }`, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// ---- Upstream: VariableDeclarator (default version → ReactDOM) ----
		{
			Code: `var Hello = ReactDOM.render(<div />, document.body);`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noReturnValue",
					Message:   "Do not depend on the return value from ReactDOM.render",
					Line:      1, Column: 13, EndLine: 1, EndColumn: 28,
				},
			},
		},

		// ---- Upstream: Object property value ----
		{
			Code: `
        var o = {
          inst: ReactDOM.render(<div />, document.body)
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noReturnValue", Line: 3, Column: 17},
			},
		},

		// ---- Upstream: ReturnStatement ----
		{
			Code: `
        function render () {
          return ReactDOM.render(<div />, document.body)
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noReturnValue", Line: 3, Column: 18},
			},
		},

		// ---- Upstream: ArrowFunctionExpression body ----
		{
			Code: `var render = (a, b) => ReactDOM.render(a, b)`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noReturnValue", Line: 1, Column: 24, EndLine: 1, EndColumn: 39},
			},
		},

		// ---- Upstream: AssignmentExpression (member-expression LHS) ----
		{
			Code: `this.o = ReactDOM.render(<div />, document.body);`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noReturnValue", Line: 1, Column: 10},
			},
		},

		// ---- Upstream: AssignmentExpression (identifier LHS) ----
		{
			Code: `var v; v = ReactDOM.render(<div />, document.body);`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noReturnValue", Line: 1, Column: 12},
			},
		},

		// ---- Upstream: React matches at ^0.14.0 ----
		{
			Code:     `var inst = React.render(<div />, document.body);`,
			Tsx:      true,
			Settings: react0140,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noReturnValue",
					Message:   "Do not depend on the return value from React.render",
					Line:      1, Column: 12,
				},
			},
		},

		// ---- Upstream: ReactDOM matches at ^0.14.0 ----
		{
			Code:     `var inst = ReactDOM.render(<div />, document.body);`,
			Tsx:      true,
			Settings: react0140,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noReturnValue",
					Message:   "Do not depend on the return value from ReactDOM.render",
					Line:      1, Column: 12,
				},
			},
		},

		// ---- Upstream: React matches at ^0.13.0 ----
		{
			Code:     `var inst = React.render(<div />, document.body);`,
			Tsx:      true,
			Settings: react0130,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noReturnValue",
					Message:   "Do not depend on the return value from React.render",
					Line:      1, Column: 12,
				},
			},
		},

		// ---- Edge: parens around the init — ESTree flattens, tsgo preserves; walk transparently ----
		{
			Code: `var x = (ReactDOM.render(<div />, document.body));`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noReturnValue", Line: 1, Column: 10},
			},
		},

		// ---- Edge: double parens ----
		{
			Code: `var x = ((ReactDOM.render(<div />, document.body)));`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noReturnValue", Line: 1, Column: 11},
			},
		},

		// ---- Edge: parens around the arrow body ----
		{
			Code: `var f = () => (ReactDOM.render(<div />, document.body))`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noReturnValue", Line: 1, Column: 16},
			},
		},

		// ---- Edge: nested arrow — inner arrow's body is the call ----
		{
			Code: `var f = () => () => ReactDOM.render(<div />, document.body);`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noReturnValue", Line: 1, Column: 21},
			},
		},

		// ---- Edge: compound assignment (`+=`) — upstream's `parent.type === 'AssignmentExpression'` matches all operators ----
		{
			Code: `var v = 0; v += ReactDOM.render(<div />, document.body);`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noReturnValue", Line: 1, Column: 17},
			},
		},

		// ---- Edge: logical assignment (`??=`) — also an AssignmentExpression ----
		{
			Code: `var v; v ??= ReactDOM.render(<div />, document.body);`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noReturnValue", Line: 1, Column: 14},
			},
		},

		// ---- Edge: chained assignment — inner `=` fires too ----
		{
			Code: `var a, b; a = b = ReactDOM.render(<div />, document.body);`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noReturnValue", Line: 1, Column: 19},
			},
		},

		// ---- Edge: destructuring-assignment RHS — still an AssignmentExpression ----
		{
			Code: `var o; ({ a: o } = ReactDOM.render(<div />, document.body));`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noReturnValue", Line: 1, Column: 20},
			},
		},

		// ---- Edge: shorthand property value ----
		{
			Code: `var o = { inst: ReactDOM.render(<div />, document.body) };`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noReturnValue", Line: 1, Column: 17},
			},
		},

		// ---- Edge: class-method return ----
		{
			Code: `
        class App {
          foo() {
            return ReactDOM.render(<div />, document.body);
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noReturnValue", Line: 4, Column: 20},
			},
		},

		// ---- Edge: class-field initializer arrow — body is the call ----
		{
			Code: `
        class App {
          cb = () => ReactDOM.render(<div />, document.body);
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noReturnValue", Line: 3, Column: 22},
			},
		},

		// ---- Edge: nested inside IIFE returning the render — the return itself consumes ----
		{
			Code: `var x = (() => { return ReactDOM.render(<div />, document.body); })();`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noReturnValue", Line: 1, Column: 25},
			},
		},

		// ---- Edge: TSX function-component return ----
		{
			Code: `
        function App() {
          return ReactDOM.render(<div />, document.body);
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noReturnValue", Line: 3, Column: 18},
			},
		},

		// ---- Edge: bracket access with an Identifier literally named `render` —
		// upstream's `'name' in callee.property && property.name === 'render'`
		// is true for Identifier(render), regardless of what the variable
		// actually references at runtime. Matches upstream exactly.
		{
			Code: `var render; var x = ReactDOM[render](<div />, document.body);`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noReturnValue", Line: 1, Column: 21},
			},
		},
	})
}

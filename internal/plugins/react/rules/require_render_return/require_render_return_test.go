package require_render_return

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestRequireRenderReturnRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &RequireRenderReturnRule, []rule_tester.ValidTestCase{
		// ---- Upstream: ES6 class ----
		{Code: `
        class Hello extends React.Component {
          render() {
            return <div>Hello {this.props.name}</div>;
          }
        }
      `, Tsx: true},

		// ---- Upstream: ES6 class with render property (arrow + block + return) ----
		{Code: `
        class Hello extends React.Component {
          render = () => {
            return <div>Hello {this.props.name}</div>;
          }
        }
      `, Tsx: true},

		// ---- Upstream: ES6 class with render property (implicit return) ----
		{Code: `
        class Hello extends React.Component {
          render = () => (
            <div>Hello {this.props.name}</div>
          )
        }
      `, Tsx: true},

		// ---- Upstream: ES5 class ----
		{Code: `
        var Hello = createReactClass({
          displayName: 'Hello',
          render: function() {
            return <div></div>
          }
        });
      `, Tsx: true},

		// ---- Upstream: Stateless function ----
		{Code: `
        function Hello() {
          return <div></div>;
        }
      `, Tsx: true},

		// ---- Upstream: Stateless arrow function ----
		{Code: `
        var Hello = () => (
          <div></div>
        );
      `, Tsx: true},

		// ---- Upstream: Return in a switch...case ----
		{Code: `
        var Hello = createReactClass({
          render: function() {
            switch (this.props.name) {
              case 'Foo':
                return <div>Hello Foo</div>;
              default:
                return <div>Hello {this.props.name}</div>;
            }
          }
        });
      `, Tsx: true},

		// ---- Upstream: Return in a if...else ----
		{Code: `
        var Hello = createReactClass({
          render: function() {
            if (this.props.name === 'Foo') {
              return <div>Hello Foo</div>;
            } else {
              return <div>Hello {this.props.name}</div>;
            }
          }
        });
      `, Tsx: true},

		// ---- Upstream: Not a React component (class doesn't extend Component) ----
		{Code: `
        class Hello {
          render() {}
        }
      `, Tsx: true},

		// ---- Upstream: ES6 class without a render method ----
		{Code: `class Hello extends React.Component {}`, Tsx: true},

		// ---- Upstream: ES5 class without a render method ----
		{Code: `var Hello = createReactClass({});`, Tsx: true},

		// ---- Upstream: ES5 class with imported (shorthand) render method ----
		{Code: `
        var render = require('./render');
        var Hello = createReactClass({
          render
        });
      `, Tsx: true},

		// ---- Upstream: Invalid render method (field without initializer) ----
		{Code: `
        class Foo extends Component {
          render
        }
      `, Tsx: true},

		// ---- Edge: PureComponent is also a React component ----
		{Code: `
        class Hello extends React.PureComponent {
          render() {
            return <div/>;
          }
        }
      `, Tsx: true},

		// ---- Edge: bare Component identifier ----
		{Code: `
        class Hello extends Component {
          render() {
            return <div/>;
          }
        }
      `, Tsx: true},

		// ---- Edge: pragma-qualified createReactClass ----
		{Code: `
        var Hello = React.createReactClass({
          render: function() {
            return <div/>;
          }
        });
      `, Tsx: true},

		// ---- Edge: render as FunctionExpression inside class field (not arrow) ----
		{Code: `
        class Hello extends React.Component {
          render = function() {
            return <div/>;
          };
        }
      `, Tsx: true},

		// ---- Edge: ES5 class with shorthand render method ----
		{Code: `
        var Hello = createReactClass({
          render() {
            return <div/>;
          }
        });
      `, Tsx: true},

		// ---- Edge: returning a ternary still counts ----
		{Code: `
        class Hello extends React.Component {
          render() {
            return this.props.ok ? <div/> : null;
          }
        }
      `, Tsx: true},

		// ---- Edge: return statement nested inside an if block ----
		{Code: `
        class Hello extends React.Component {
          render() {
            if (!this.props.ready) {
              return null;
            }
            return <div/>;
          }
        }
      `, Tsx: true},

		// ---- Edge: multiple render-named members — any one with a return marks the component OK ----
		// Mirrors ESLint's markReturnStatementPresent tracking the whole
		// component, not any particular render method.
		{Code: `
        class Hello extends React.Component {
          static render() {}
          render() { return <div/>; }
        }
      `, Tsx: true},

		// ---- Edge: string-literal key "render" is NOT matched (ESLint's getPropertyName only reads nameNode.name) ----
		// So this component is treated as having NO render method → not flagged.
		{Code: `
        class Hello extends React.Component {
          "render"() {}
        }
      `, Tsx: true},

		// ---- Edge: non-React class with an inner React component — only the React inner class is examined ----
		{Code: `
        class Outer {
          render() {}
          inner() {
            class Inner extends React.Component {
              render() {
                return <div/>;
              }
            }
          }
        }
      `, Tsx: true},

		// ---- Edge: createReactClass invoked with zero arguments is safely ignored ----
		{Code: `createReactClass();`, Tsx: true},

		// ---- Edge: createReactClass invoked with a non-object argument is safely ignored ----
		{Code: `createReactClass(1);`, Tsx: true},

		// ---- Edge: custom pragma — React-like name configured via settings ----
		{
			Code: `
        class Hello extends Preact.Component {
          render() {
            return <div/>;
          }
        }
      `,
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"pragma": "Preact"}},
		},

		// ---- Edge: string-literal key in ES5 object is NOT matched — component has no render by ESLint's rules ----
		{Code: `
        var Hello = createReactClass({
          "render": function() {}
        });
      `, Tsx: true},

		// ---- Edge: computed key `['render']` — Literal inside ComputedPropertyName, name is undefined → not matched ----
		{Code: `
        class Hello extends React.Component {
          ['render']() {}
        }
      `, Tsx: true},

		// ---- Edge: template-literal key (NoSubstitutionTemplateLiteral) — no .name → not matched ----
		{Code: `
        class Hello extends React.Component {
          [` + "`render`" + `]() {}
        }
      `, Tsx: true},

		// ---- Edge: getter render with a return ----
		{Code: `
        class Hello extends React.Component {
          get render() {
            return () => <div/>;
          }
        }
      `, Tsx: true},

		// ---- Edge: bare `return;` still counts as having a return ----
		{Code: `
        class Hello extends React.Component {
          render() {
            return;
          }
        }
      `, Tsx: true},

		// ---- Edge: return inside try / catch / finally ----
		{Code: `
        class Hello extends React.Component {
          render() {
            try {
              return <div/>;
            } catch (e) {}
          }
        }
      `, Tsx: true},
		{Code: `
        class Hello extends React.Component {
          render() {
            try {} finally { return <div/>; }
          }
        }
      `, Tsx: true},

		// ---- Edge: return inside for / while ----
		{Code: `
        class Hello extends React.Component {
          render() {
            for (let i = 0; i < 1; i++) { return <div/>; }
          }
        }
      `, Tsx: true},

		// ---- Edge: parens-wrapped ClassExpression assignment ----
		{Code: `
        var Hello = (class extends React.Component {
          render() {
            return <div/>;
          }
        });
      `, Tsx: true},

		// ---- Edge: parens-wrapped createReactClass callee ----
		{Code: `
        var Hello = (createReactClass)({
          render: function() {
            return <div/>;
          }
        });
      `, Tsx: true},

		// ---- Edge: parens-wrapped object argument to createReactClass ----
		{Code: `
        var Hello = createReactClass(({
          render: function() {
            return <div/>;
          }
        }));
      `, Tsx: true},

		// ---- Edge: class extending an intermediate (not Component) — not a React component ----
		{Code: `
        class Middle extends React.Component {
          render() { return <div/>; }
        }
        class Hello extends Middle {
          render() {}
        }
      `, Tsx: true},

		// ---- Edge: `declare class` render overload (no body) — skipped ----
		{Code: `
        declare class Hello extends React.Component {
          render(): JSX.Element;
        }
      `, Tsx: true},

		// ---- Edge: React.memo wrapping a bare class (not extending Component) — not a React component ----
		{Code: `const Foo = React.memo(class { render() {} });`, Tsx: true},

		// ---- Edge: aliased `createReactClass` via a variable — not detected (literal call only) ----
		{Code: `
        const c = createReactClass;
        c({ render: function() {} });
      `, Tsx: true},

		// ---- Edge: inner-class static block with IIFE — outer render body has no direct return → fires on outer, not inner ----
		// (Pure regression lock: walker must stop at the IIFE's FunctionExpression.)
		// NOTE: this fires — see invalid section below.

		// ---- H3: numeric-literal key ----
		{Code: `
        class Hello extends React.Component {
          42() {}
        }
      `, Tsx: true},

		// ---- L1: async render with return ----
		{Code: `
        class Hello extends React.Component {
          async render() { return <div/>; }
        }
      `, Tsx: true},

		// ---- K2: getter render with a nested function expression's return ----
		// getter body contains a return of a FunctionExpression whose inner
		// return is attached to that FE — the outer getter has its own return,
		// so the component passes.
		{Code: `
        class Hello extends React.Component {
          get render() { return function() { return <div/>; }; }
        }
      `, Tsx: true},

		// ---- M2: labeled block containing return — return is directly in render body ----
		{Code: `
        class Hello extends React.Component {
          render() {
            outer: {
              return <div/>;
            }
          }
        }
      `, Tsx: true},

		// ---- M3: return inside catch clause ----
		{Code: `
        class Hello extends React.Component {
          render() {
            try {} catch (e) { return <div/>; }
          }
        }
      `, Tsx: true},

		// ---- M8: arrow expression body returning a ternary ----
		{Code: `
        class Hello extends React.Component {
          render = () => (this.props.ok ? <div/> : null);
        }
      `, Tsx: true},

		// ---- M9: render = someFactory() — value is CallExpression, not function-like ----
		{Code: `
        class Hello extends React.Component {
          render = someFactory();
        }
      `, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// ---- Upstream: Missing return in ES5 class ----
		{
			Code: `
        var Hello = createReactClass({
          displayName: 'Hello',
          render: function() {}
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noRenderReturn",
					Message:   "Your render method should have a return statement",
					Line:      4, Column: 11,
				},
			},
		},

		// ---- Upstream: Missing return in ES6 class ----
		{
			Code: `
        class Hello extends React.Component {
          render() {}
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noRenderReturn", Line: 3, Column: 11},
			},
		},

		// ---- Upstream: Missing return (but one is present in a sub-function) ----
		{
			Code: `
        class Hello extends React.Component {
          render() {
            const names = this.props.names.map(function(name) {
              return <div>{name}</div>
            });
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noRenderReturn", Line: 3, Column: 11},
			},
		},

		// ---- Upstream: Missing return ES6 class render property ----
		{
			Code: `
        class Hello extends React.Component {
          render = () => {
            <div>Hello {this.props.name}</div>
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noRenderReturn", Line: 3, Column: 11},
			},
		},

		// ---- Edge: Missing return in shorthand method inside createReactClass ----
		{
			Code: `
        var Hello = createReactClass({
          render() {}
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noRenderReturn", Line: 3, Column: 11},
			},
		},

		// ---- Edge: Missing return with return inside a nested FunctionDeclaration ----
		{
			Code: `
        class Hello extends React.Component {
          render() {
            function inner() { return <div/>; }
            inner();
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noRenderReturn", Line: 3, Column: 11},
			},
		},

		// ---- Edge: Missing return in bare-Component ES6 class ----
		{
			Code: `
        class Hello extends Component {
          render() {}
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noRenderReturn", Line: 3, Column: 11},
			},
		},

		// ---- Edge: Missing return in PureComponent subclass ----
		{
			Code: `
        class Hello extends React.PureComponent {
          render() {}
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noRenderReturn", Line: 3, Column: 11},
			},
		},

		// ---- Edge: Missing return with render = function(){} (FunctionExpression initializer) ----
		{
			Code: `
        class Hello extends React.Component {
          render = function() {};
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noRenderReturn", Line: 3, Column: 11},
			},
		},

		// ---- Edge: pragma-qualified createReactClass with missing render ----
		{
			Code: `
        var Hello = React.createReactClass({
          render: function() {}
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noRenderReturn", Line: 3, Column: 11},
			},
		},

		// ---- Edge: ClassExpression with missing return ----
		{
			Code: `
        var Hello = class extends React.Component {
          render() {}
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noRenderReturn", Line: 3, Column: 11},
			},
		},

		// ---- Edge: custom pragma via settings — Preact.Component ----
		{
			Code: `
        class Hello extends Preact.Component {
          render() {}
        }
      `,
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"pragma": "Preact"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noRenderReturn", Line: 3, Column: 11},
			},
		},

		// ---- Edge: getter render with no return ----
		{
			Code: `
        class Hello extends React.Component {
          get render() {
            const f = () => <div/>;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noRenderReturn", Line: 3, Column: 11},
			},
		},

		// ---- Edge: return inside a nested FunctionExpression IIFE — does NOT count ----
		{
			Code: `
        class Hello extends React.Component {
          render() {
            (function() { return <div/>; })();
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noRenderReturn", Line: 3, Column: 11},
			},
		},

		// ---- Edge: return inside a nested ArrowFunction (block body) — also does NOT count ----
		// ESLint's unanchored regex `/Function(Expression|Declaration)$/` matches
		// `ArrowFunctionExpression` too, so arrows bump depth past 1.
		{
			Code: `
        class Hello extends React.Component {
          render() {
            const f = () => { return <div/>; };
            f();
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noRenderReturn", Line: 3, Column: 11},
			},
		},

		// ---- Edge: arrow IIFE inside render — return inside arrow does NOT count ----
		{
			Code: `
        class Hello extends React.Component {
          render() {
            (() => { return <div/>; })();
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noRenderReturn", Line: 3, Column: 11},
			},
		},

		// ---- Edge: render = () => { nested arrow with return, no top-level return } ----
		{
			Code: `
        class Hello extends React.Component {
          render = () => {
            const f = () => { return <div/>; };
            f();
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noRenderReturn", Line: 3, Column: 11},
			},
		},

		// ---- Edge: return inside a nested FunctionDeclaration — does NOT count ----
		{
			Code: `
        class Hello extends React.Component {
          render() {
            function inner() { return <div/>; }
            inner();
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noRenderReturn", Line: 3, Column: 11},
			},
		},

		// ---- Edge: return inside a nested inner-class method — does NOT count ----
		{
			Code: `
        class Hello extends React.Component {
          render() {
            class Inner {
              foo() { return <div/>; }
            }
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noRenderReturn", Line: 3, Column: 11},
			},
		},

		// ---- Edge: inner ClassDeclaration is itself a React component and gets reported ----
		{
			Code: `
        class Outer {
          foo() {
            class Inner extends React.Component {
              render() {}
            }
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noRenderReturn", Line: 5, Column: 15},
			},
		},

		// ---- Edge: render with only a throw (no return) ----
		{
			Code: `
        class Hello extends React.Component {
          render() {
            throw new Error('nope');
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noRenderReturn", Line: 3, Column: 11},
			},
		},

		// ---- Edge: parens-wrapped ClassExpression with missing return ----
		{
			Code: `
        var Hello = (class extends React.Component {
          render() {}
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noRenderReturn", Line: 3, Column: 11},
			},
		},

		// ---- Edge: parens-wrapped object arg to createReactClass with missing return ----
		{
			Code: `
        var Hello = createReactClass(({
          render: function() {}
        }));
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noRenderReturn", Line: 3, Column: 11},
			},
		},

		// ---- Edge: static render method (rule does not filter by static) ----
		{
			Code: `
        class Hello extends React.Component {
          static render() {}
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noRenderReturn", Line: 3, Column: 11},
			},
		},

		// ---- Edge: setter named render (no return) ----
		{
			Code: `
        class Hello extends React.Component {
          set render(v) {}
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noRenderReturn", Line: 3, Column: 11},
			},
		},

		// ---- Edge: ES5 object getter with no return ----
		{
			Code: `
        var Hello = createReactClass({
          get render() { const f = () => <div/>; }
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noRenderReturn", Line: 3, Column: 11},
			},
		},

		// ---- Edge: HOC-wrapped class expression that still extends Component ----
		{
			Code: `const Foo = withRouter(class extends React.Component { render() {} });`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noRenderReturn", Line: 1, Column: 56},
			},
		},

		// ---- Edge: spread in createReactClass object, render has no return ----
		{
			Code: `createReactClass({ ...config, render: function() {} });`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noRenderReturn", Line: 1, Column: 31},
			},
		},

		// ---- Edge: F1 — inner-class static block with deeply nested IIFE return ----
		// Walker must stop at the FunctionExpression IIFE; the IIFE's return
		// does not mark the outer render as returning.
		{
			Code: `
        class Hello extends React.Component {
          render() {
            class Inner {
              static {
                const x = (function() { return 1; })();
              }
            }
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noRenderReturn", Line: 3, Column: 11},
			},
		},

		// ---- L2: generator *render() with no return — rule does NOT skip generators ----
		{
			Code: `
        class Hello extends React.Component {
          *render() {}
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noRenderReturn", Line: 3, Column: 11},
			},
		},

		// ---- M1: deeply nested IIFE inside do/while + try/catch — IIFE FE stops walker ----
		{
			Code: `
        class Hello extends React.Component {
          render() {
            do {
              (function() {
                try { return <div/>; } catch (e) {}
              })();
            } while (false);
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noRenderReturn", Line: 3, Column: 11},
			},
		},

		// ---- M5: arrow with expression body as local var (no ReturnStatement anywhere) ----
		// No ReturnStatement in render; nested arrow's implicit return does NOT
		// mark anything because the ArrowFunctionExpression listener only marks
		// when arrow.parent is the render property (here parent is VariableDeclarator).
		{
			Code: `
        class Hello extends React.Component {
          render() {
            const f = () => <div/>;
            f();
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noRenderReturn", Line: 3, Column: 11},
			},
		},

		// ---- M6: class expression nested as property value — only the OUTER fires ----
		// Outer render has no return → fire. Inner class is itself a React
		// component whose render DOES return → no fire on inner.
		{
			Code: `
        class Outer extends React.Component {
          render() {
            const obj = {
              inner: class extends React.Component {
                render() { return <div/>; }
              }
            };
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noRenderReturn", Line: 3, Column: 11},
			},
		},

		// ---- M10: two createReactClass calls; only the one missing return fires ----
		{
			Code: `var A = createReactClass({ render: function() { return 1; } });
var B = createReactClass({ render: function() {} });`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noRenderReturn", Line: 2, Column: 28},
			},
		},

		// ---- M11: class inside an array literal — BOTH outer and inner are React components with missing return ----
		{
			Code: `
        class Hello extends React.Component {
          render() {
            const arr = [class Inner extends React.Component { render() {} }];
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noRenderReturn", Line: 3, Column: 11},
				{MessageId: "noRenderReturn", Line: 4, Column: 64},
			},
		},
	})
}

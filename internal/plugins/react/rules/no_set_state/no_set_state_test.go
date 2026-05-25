package no_set_state

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoSetStateRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoSetStateRule, []rule_tester.ValidTestCase{
		// ---- Upstream: bare function (not a component — no JSX return) ----
		{Code: `
        var Hello = function() {
          this.setState({})
        };
      `, Tsx: true},

		// ---- Upstream: createReactClass with render only, no setState ----
		{Code: `
        var Hello = createReactClass({
          render: function() {
            return <div>Hello {this.props.name}</div>;
          }
        });
      `, Tsx: true},

		// ---- Upstream: `this.setState` reference (not call) is allowed ----
		// `this.someHandler = this.setState;` — the rule listens to
		// CallExpression, so a bare property read never fires.
		{Code: `
        var Hello = createReactClass({
          componentDidUpdate: function() {
            someNonMemberFunction(arg);
            this.someHandler = this.setState;
          },
          render: function() {
            return <div>Hello {this.props.name}</div>;
          }
        });
      `, Tsx: true},

		// ---- Edge: plain class (no `extends Component`) — not a component ----
		{Code: `
        class Hello {
          someMethod() {
            this.setState({});
          }
        }
      `, Tsx: true},

		// ---- Edge: bare `this.setState({})` outside any component / class ----
		{Code: `this.setState({});`, Tsx: true},

		// ---- Edge: receiver is not `this` ----
		{Code: `
        class Hello extends React.Component {
          componentDidMount() {
            other.setState({});
          }
        }
      `, Tsx: true},

		// ---- Edge: `this['setState']({})` — element access, no name match ----
		// ESLint checks `callee.property.name === 'setState'`, which is
		// undefined on a computed property; we mirror by gating on
		// KindIdentifier. Element-access never reaches the
		// PropertyAccessExpression branch at all.
		{Code: `
        class Hello extends React.Component {
          componentDidMount() {
            this['setState']({});
          }
        }
      `, Tsx: true},

		// ---- Edge: `super.setState({})` — receiver is SuperKeyword, not ThisKeyword ----
		{Code: `
        class Hello extends React.Component {
          componentDidMount() {
            super.setState({});
          }
        }
      `, Tsx: true},

		// ---- Edge: `(this as any).setState({})` — TS as-expression breaks the receiver match ----
		// SkipParentheses unwraps Parens only; `as` / `satisfies` / `!`
		// remain as wrappers, so the receiver-check fails (matches the
		// upstream behavior — it does the equivalent textual check before
		// any TS wrapper has been stripped).
		{Code: `
        class Hello extends React.Component {
          componentDidMount() {
            (this as any).setState({});
          }
        }
      `, Tsx: true},

		// ---- Edge: `this!.setState({})` — TS non-null assertion breaks the receiver match ----
		{Code: `
        class Hello extends React.Component {
          componentDidMount() {
            this!.setState({});
          }
        }
      `, Tsx: true},

		// ---- Edge: `(this satisfies React.Component).setState({})` — TS satisfies wrapper ----
		{Code: `
        class Hello extends React.Component {
          componentDidMount() {
            (this satisfies React.Component).setState({});
          }
        }
      `, Tsx: true},

		// ---- Edge: ConditionalExpression receiver — not a bare ThisKeyword ----
		{Code: `
        class Hello extends React.Component {
          componentDidMount() {
            (cond ? this : other).setState({});
          }
        }
      `, Tsx: true},

		// ---- Edge: `setState` reference assigned to a variable — no CallExpression fires ----
		{Code: `
        class Hello extends React.Component {
          componentDidMount() {
            const fn = this.setState;
          }
        }
      `, Tsx: true},

		// ---- Edge: tagged template `this.setState\`{}\`` — TaggedTemplateExpression, not CallExpression ----
		{Code: "\n        class Hello extends React.Component {\n          componentDidMount() {\n            this.setState`{}`;\n          }\n        }\n      ", Tsx: true},

		// ---- Edge: `new this.setState({})` — NewExpression, not CallExpression ----
		{Code: `
        class Hello extends React.Component {
          componentDidMount() {
            new this.setState({});
          }
        }
      `, Tsx: true},

		// ---- Edge: comma-operator callee `(0, this.setState)({})` — callee is BinaryExpression, not PropertyAccess ----
		{Code: `
        class Hello extends React.Component {
          componentDidMount() {
            (0, this.setState)({});
          }
        }
      `, Tsx: true},

		// ---- Edge: indirect call `this.setState.call(this, {})` — outer call's callee is `this.setState.call`, .call's receiver is `this.setState` (PropertyAccess), not ThisKeyword ----
		{Code: `
        class Hello extends React.Component {
          componentDidMount() {
            this.setState.call(this, {});
            this.setState.bind(this);
            this.setState.apply(this, [{}]);
          }
        }
      `, Tsx: true},

		// ---- Edge: lowercased function — not a stateless component ----
		// `IsStatelessReactComponent` requires a capital-cased binding
		// (or anonymous + export default).
		{Code: `
        function hello() {
          this.setState({});
          return <div/>;
        }
      `, Tsx: true},

		// ---- Edge: capital-cased function NOT returning JSX — not a component ----
		{Code: `
        function Hello() {
          this.setState({});
          return 42;
        }
      `, Tsx: true},

		// ---- Edge: TS abstract class with body-absent componentDidMount ----
		// No body → no setState possible. Locks that the rule traverses
		// signature-only members without crashing.
		{Code: `
        abstract class Hello extends React.Component {
          abstract componentDidMount(): void;
        }
      `, Tsx: true},

		// ---- Cross-helper lock: `extends mixin(React.Component)` — extends
		// clause is a CallExpression (not Identifier or PropertyAccess);
		// ExtendsReactComponent does not recognize it. Empirically verified
		// against ESLint master (NOT reported). ----
		{Code: `
        class Hello extends mixin(React.Component) {
          someMethod() {
            this.setState({});
          }
          render() { return <div/>; }
        }
      `, Tsx: true},

		// ---- Edge: extends arbitrary base class — not a React component ----
		{Code: `
        class Hello extends MyOwnBase {
          someMethod() {
            this.setState({});
          }
        }
      `, Tsx: true},

		// ---- Edge: extends React.SomeOther — only Component / PureComponent match ----
		{Code: `
        class Hello extends React.SomeOther {
          someMethod() {
            this.setState({});
          }
        }
      `, Tsx: true},

		// ---- Edge: JSX onClick={this.setState} — reference, not call ----
		// Argument to onClick is a bare PropertyAccessExpression (no
		// CallExpression around it), so the listener never fires.
		{Code: `
        class Hello extends React.Component {
          render() {
            return <button onClick={this.setState}/>;
          }
        }
      `, Tsx: true},

		// ---- Edge: chained `this.setState.bind(this, {})` — outer call is on .bind, not setState ----
		// JSX usage: `onClick={this.setState.bind(this, {})}`. The outer
		// CallExpression's callee is `this.setState.bind`, whose
		// PropertyAccess receiver is `this.setState` (PropertyAccess),
		// not ThisKeyword. Locked as VALID — `.bind(...)` calls are not
		// the setState call site.
		{Code: `
        class Hello extends React.Component {
          render() {
            return <button onClick={this.setState.bind(this, {})}/>;
          }
        }
      `, Tsx: true},

		// ---- Edge: `this.x.setState({})` — multi-level receiver, base is not bare `this` ----
		{Code: `
        class Hello extends React.Component {
          someMethod() {
            this.x.setState({});
          }
        }
      `, Tsx: true},

		// ---- Edge: `this.constructor.setState({})` — same shape as above ----
		{Code: `
        class Hello extends React.Component {
          someMethod() {
            this.constructor.setState({});
          }
        }
      `, Tsx: true},

		// ---- Edge: legacy type assertion `<T>this.setState({})` (non-TSX file) ----
		// In TSX this would be parsed as JSX; in plain TS it's a type
		// assertion that wraps `this.setState({})`. The wrapper is the
		// outer expression, but the inner `this.setState({})` is still a
		// CallExpression that the listener visits. The wrapper around
		// the OUTER expression doesn't break the receiver check (the
		// receiver of the inner setState call is bare `this`).
		{Code: `
class Hello {
  someMethod() {
    <any>this.setState({});
  }
}
`, Tsx: false},

		// ---- Edge: declare class — ambient declaration, no body ----
		// declare class is purely a type signature; no method bodies, no
		// setState possible. Locks no-crash on traversal.
		{Code: `
        declare class Hello extends React.Component {
          someMethod(): void;
        }
      `, Tsx: true},

		// ---- Edge: settings.react is not a map (boolean) — must NOT crash ----
		// GetReactPragma defensive guard: settings.react.(map[..]) ok=false
		// → fall back to default. Locks parity with the existing
		// no_did_mount_set_state test that exercises the same shape.
		{Code: `
        class Hello extends React.Component {
          someMethod() {
            other.setState({});
          }
        }
      `, Tsx: true, Settings: map[string]interface{}{"react": true}},

		// ---- Edge: settings.react.pragma is not a string (number) — falls back to default ----
		{Code: `
        class Hello extends React.Component {
          someMethod() {
            other.setState({});
          }
        }
      `, Tsx: true, Settings: map[string]interface{}{"react": map[string]interface{}{"pragma": 42}}},

		// ---- Edge: custom pragma — `class Hello extends Preact.Component` is NOT React ----
		// With `settings.react.pragma = "Preact"`, `React.Component` is
		// no longer the recognized base — only `Preact.Component` (or bare
		// `Component`/`PureComponent`) is. So the inner class extending
		// `React.Component` is NOT a detected component under the Preact
		// pragma, and `this.setState` inside is NOT reported.
		{Code: `
        class Hello extends React.Component {
          someMethod() {
            this.setState({});
          }
        }
      `, Tsx: true, Settings: map[string]interface{}{"react": map[string]interface{}{"pragma": "Preact"}}},

		// ---- Edge: custom createClass — `var X = createReactClass({...})` is NOT a component ----
		// With `settings.react.createClass = "myCreate"`, only
		// `myCreate({...})` calls are recognized as ES5 components.
		{Code: `
        var Hello = createReactClass({
          someMethod: function() {
            this.setState({});
          },
          render: function() { return <div/>; }
        });
      `, Tsx: true, Settings: map[string]interface{}{"react": map[string]interface{}{"createClass": "myCreate"}}},


		// ---- Edge: logical-OR receiver `(this || other).setState({})` ----
		// `this || other` is a BinaryExpression(||). SkipParentheses
		// unwraps the outer parens, but the receiver is still a
		// BinaryExpression — not ThisKeyword. Locked as VALID, mirroring
		// the same shape lock in no_did_mount_set_state.
		{Code: `
        class Hello extends React.Component {
          someMethod() {
            (this || other).setState({});
          }
        }
      `, Tsx: true},

		// ---- Edge: logical-AND receiver `(this && other).setState({})` ----
		// Same reasoning as the OR variant — symmetric lock.
		{Code: `
        class Hello extends React.Component {
          someMethod() {
            (this && other).setState({});
          }
        }
      `, Tsx: true},

		// ---- Edge: comma-operator receiver `(this, other).setState({})` ----
		// `(this, other)` is a BinaryExpression(comma). Same as above.
		{Code: `
        class Hello extends React.Component {
          someMethod() {
            (this, other).setState({});
          }
        }
      `, Tsx: true},

		// ---- Edge: ElementAccess on setState `this.setState[0]({})` ----
		// Outer call's callee is `this.setState[0]`
		// (ElementAccessExpression), not PropertyAccessExpression. The
		// kind check rejects.
		{Code: `
        class Hello extends React.Component {
          someMethod() {
            (this.setState as any)[0]({});
          }
        }
      `, Tsx: true},

		// ---- Edge: PropertyAccess chained off setState `this.setState.fn({})` ----
		// Outer call's callee is `this.setState.fn` (PropertyAccess).
		// Receiver is `this.setState` (PropertyAccess), not ThisKeyword.
		{Code: `
        class Hello extends React.Component {
          someMethod() {
            this.setState.fn({});
          }
        }
      `, Tsx: true},

		// ---- Edge: with-statement — bare `setState(...)` callee is Identifier, not PropertyAccess ----
		// Even though `with (this)` would resolve `setState` to
		// `this.setState` at runtime, textually the callee is a bare
		// Identifier. Both upstream and rslint require explicit `this.`.
		{Code: `
        // @ts-nocheck
        class Hello extends React.Component {
          someMethod() {
            with (this as any) {
              setState({});
            }
          }
        }
      `, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// ---- Upstream #1: createReactClass componentDidUpdate ----
		{
			Code: `
        var Hello = createReactClass({
          componentDidUpdate: function() {
            this.setState({
              name: this.props.name.toUpperCase()
            });
          },
          render: function() {
            return <div>Hello {this.state.name}</div>;
          }
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},

		// ---- Upstream #2: createReactClass someMethod ----
		{
			Code: `
        var Hello = createReactClass({
          someMethod: function() {
            this.setState({
              name: this.props.name.toUpperCase()
            });
          },
          render: function() {
            return <div onClick={this.someMethod.bind(this)}>Hello {this.state.name}</div>;
          }
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},

		// ---- Upstream #3: ES6 class someMethod ----
		{
			Code: `
        class Hello extends React.Component {
          someMethod() {
            this.setState({
              name: this.props.name.toUpperCase()
            });
          }
          render() {
            return <div onClick={this.someMethod.bind(this)}>Hello {this.state.name}</div>;
          }
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},

		// ---- Upstream #4: class field arrow ----
		{
			Code: `
        class Hello extends React.Component {
          someMethod = () => {
            this.setState({
              name: this.props.name.toUpperCase()
            });
          }
          render() {
            return <div onClick={this.someMethod.bind(this)}>Hello {this.state.name}</div>;
          }
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},

		// ---- Upstream #5: setState inside JSX onMouseEnter callback ----
		{
			Code: `
        class Hello extends React.Component {
          render() {
            return <div onMouseEnter={() => this.setState({dropdownIndex: index})} />;
          }
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 45},
			},
		},

		// ---- Edge: parenthesized `(this).setState({})` ----
		{
			Code: `
        class Hello extends React.Component {
          render() {
            (this).setState({});
            return <div/>;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},

		// ---- Edge: multi-paren receiver `((this)).setState({})` ----
		{
			Code: `
        class Hello extends React.Component {
          render() {
            ((this)).setState({});
            return <div/>;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},

		// ---- Edge: optional-chain receiver `this?.setState({})` ----
		{
			Code: `
        class Hello extends React.Component {
          render() {
            this?.setState({});
            return <div/>;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},

		// ---- Edge: optional-call `this.setState?.({})` ----
		{
			Code: `
        class Hello extends React.Component {
          render() {
            this.setState?.({});
            return <div/>;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},

		// ---- Edge: parens around the entire callee `(this.setState)({})` ----
		// SkipParentheses on `call.Expression` unwraps to the underlying
		// PropertyAccessExpression; column lands at `this` (after the
		// stripped `(`).
		{
			Code: `
        class Hello extends React.Component {
          render() {
            (this.setState)({});
            return <div/>;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 14},
			},
		},

		// ---- Edge: spread arguments — call shape unaffected ----
		{
			Code: `
        class Hello extends React.Component {
          render() {
            this.setState(...args);
            return <div/>;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},

		// ---- Edge: setState inside setTimeout callback (still inside the component) ----
		{
			Code: `
        class Hello extends React.Component {
          componentDidMount() {
            setTimeout(() => {
              this.setState({});
            }, 100);
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 5, Column: 15},
			},
		},

		// ---- Edge: setState inside JSX onClick arrow inside render ----
		{
			Code: `
        class Hello extends React.Component {
          render() {
            return <button onClick={() => this.setState({})}/>;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 43},
			},
		},

		// ---- Edge: setState inside constructor ----
		// Upstream's rule does NOT carve out the constructor (unlike
		// no-direct-mutation-state); a setState call there is still
		// flagged. Locks behavior parity.
		{
			Code: `
        class Hello extends React.Component {
          constructor(props) {
            super(props);
            this.setState({});
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 5, Column: 13},
			},
		},

		// ---- Edge: multiple setState calls in the same component, each reported ----
		{
			Code: `
        class Hello extends React.Component {
          render() {
            this.setState({});
            if (x) { this.setState({}); }
            return <div/>;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
				{MessageId: "noSetState", Line: 5, Column: 22},
			},
		},

		// ---- Edge: async render-like method ----
		{
			Code: `
        class Hello extends React.Component {
          async load() {
            await Promise.resolve();
            this.setState({});
          }
          render() { return <div/>; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 5, Column: 13},
			},
		},

		// ---- Edge: generator method ----
		{
			Code: `
        class Hello extends React.Component {
          *load() {
            yield this.setState({});
          }
          render() { return <div/>; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 19},
			},
		},

		// ---- Edge: static method on a React component ----
		// `static someMethod` isn't typically called with a real component
		// `this`, but the rule is purely structural — `this.setState` here
		// is still inside a class extending React.Component, so it
		// reports. Locks that we are not gating on `static`.
		{
			Code: `
        class Hello extends React.Component {
          static someMethod() {
            this.setState({});
          }
          render() { return <div/>; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},

		// ---- Edge: class expression assigned to a const ----
		{
			Code: `
        const Hello = class extends React.Component {
          someMethod() {
            this.setState({});
          }
          render() { return <div/>; }
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},

		// ---- Edge: bare `extends Component` (no React. prefix) — also a component ----
		{
			Code: `
        class Hello extends Component {
          someMethod() {
            this.setState({});
          }
          render() { return <div/>; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},

		// ---- Edge: extends PureComponent ----
		{
			Code: `
        class Hello extends React.PureComponent {
          someMethod() {
            this.setState({});
          }
          render() { return <div/>; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},

		// ---- Edge: nested class component — inner class is the component owner ----
		// The walk from the inner setState reaches the inner Class first;
		// `GetEnclosingReactComponent` stops at the nearest enclosing
		// class scope, so reporting attributes the call to Inner.
		{
			Code: `
        class Outer extends React.Component {
          render() {
            class Inner extends React.Component {
              someMethod() {
                this.setState({});
              }
              render() { return <div/>; }
            }
            return <div/>;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 6, Column: 17},
			},
		},

		// ---- Edge: try/catch wrapping setState ----
		{
			Code: `
        class Hello extends React.Component {
          componentDidMount() {
            try {
              this.setState({});
            } catch (e) {}
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 5, Column: 15},
			},
		},

		// ---- Edge: switch/case wrapping setState ----
		{
			Code: `
        class Hello extends React.Component {
          componentDidMount() {
            switch (this.props.kind) {
              case 1:
                this.setState({});
                break;
            }
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 6, Column: 17},
			},
		},

		// ---- Edge: comment between `this` and `.setState` — AST shape preserved ----
		{
			Code: `
        class Hello extends React.Component {
          render() {
            this/* hi */.setState({});
            return <div/>;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},

		// ---- Edge: newline between `this` and `.setState` ----
		{
			Code: `
        class Hello extends React.Component {
          render() {
            this
              .setState({});
            return <div/>;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},

		// ---- Edge: stateless functional component returning JSX ----
		// `IsStatelessReactComponent` matches Capital-cased function that
		// returns JSX → `getParentComponent` would resolve here, so
		// `this.setState` (semantically meaningless but textually present)
		// still reports. Locks parity with upstream.
		{
			Code: `
        function Hello() {
          this.setState({});
          return <div/>;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 3, Column: 11},
			},
		},

		// ---- Edge: stateless arrow component assigned to capitalized const ----
		{
			Code: `
        const Hello = () => {
          this.setState({});
          return <div/>;
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 3, Column: 11},
			},
		},

		// ---- Edge: PrivateIdentifier callee `this.#setState()` ----
		// ESLint's PrivateIdentifier surfaces `.name` with the leading
		// `#` stripped, so upstream's `callee.property.name === 'setState'`
		// matches. We mirror that — `propertyName` strips the `#`, so a
		// private `#setState` reports just like a public one.
		{
			Code: `
        class Hello extends React.Component {
          someMethod() {
            this.#setState({});
          }
          #setState(_x: unknown) {}
          render() { return <div/>; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},

		// ---- Edge: setState in default-export class ----
		{
			Code: `
        export default class extends React.Component {
          someMethod() {
            this.setState({});
          }
          render() { return <div/>; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},

		// ---- Edge: TS overload signatures + body-bearing implementation ----
		{
			Code: `
        class Hello extends React.Component {
          load(): void;
          load(arg: number): void;
          load(_arg?: number) {
            this.setState({});
          }
          render() { return <div/>; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 6, Column: 13},
			},
		},

		// ---- Edge: setState in createReactClass with custom pragma still works under default ----
		// (Custom pragma settings are exercised in plugin-wide tests; here
		// we lock the default-pragma path with multiple lifecycle hooks
		// reporting independently.)
		{
			Code: `
        var Hello = createReactClass({
          componentDidMount: function() { this.setState({}); },
          componentDidUpdate: function() { this.setState({}); },
          render: function() { return <div/>; }
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 3, Column: 43},
				{MessageId: "noSetState", Line: 4, Column: 44},
			},
		},

		// ---- Edge: getter `get foo() { this.setState({}); }` inside a component ----
		{
			Code: `
        class Hello extends React.Component {
          get foo() {
            this.setState({});
            return null;
          }
          render() { return <div/>; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},

		// ---- Edge: setter `set foo(v) { this.setState({}); }` inside a component ----
		{
			Code: `
        class Hello extends React.Component {
          set foo(_v: any) {
            this.setState({});
          }
          render() { return <div/>; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},

		// ---- Edge: arrow with concise (expression) body returning setState call ----
		{
			Code: `
        class Hello extends React.Component {
          someMethod = () => this.setState({});
          render() { return <div/>; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 3, Column: 30},
			},
		},

		// ---- Edge: namespace-wrapped class component ----
		{
			Code: `
        namespace App {
          export class Hello extends React.Component {
            someMethod() {
              this.setState({});
            }
            render() { return <div/>; }
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 5, Column: 15},
			},
		},

		// ---- Edge: createReactClass wrapped in extra parens ----
		// `createReactClass(({...}))` — tsgo preserves parens; the
		// component-detection helper unwraps them when matching the
		// argument-position. Locks that the rule sees through them.
		{
			Code: `
        var Hello = createReactClass(({
          someMethod: function() {
            this.setState({});
          },
          render: function() { return <div/>; }
        }));
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},

		// ---- Edge: setState chained on a new instance of the component ----
		// `new Hello().setState({})` — outer is NewExpression, but the
		// trailing `.setState({})` is a CallExpression on
		// PropertyAccessExpression whose receiver is the NewExpression,
		// not ThisKeyword. Even when this whole expression sits inside
		// another class, the call is NOT matched (receiver != `this`).
		// Locks the receiver guard against false positives from
		// "constructor + chained call" shapes. (Kept here as comment —
		// already covered by the `other.setState({})` valid case above.)

		// ---- Edge: createReactClass with computed key (non-Identifier method name) ----
		// `[methodName]: function() { this.setState({}); }` — the method
		// name is a ComputedPropertyName, but the rule doesn't gate on
		// the method name at all (only on the call). The setState call
		// inside still reports. Locks that the rule is purely call-site
		// based, not stopper-name based (unlike no-did-mount-set-state).
		{
			Code: `
        var Hello = createReactClass({
          [methodName]: function() {
            this.setState({});
          },
          render: function() { return <div/>; }
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},

		// ---- Edge: TS auto-accessor field `accessor someMethod = () => { ... }` ----
		{
			Code: `
        class Hello extends React.Component {
          accessor someMethod = () => {
            this.setState({});
          };
          render() { return <div/>; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},

		// ---- Edge: ClassStaticBlockDeclaration — `this` here is the class itself, ----
		// not an instance, but the rule is structural — `this.setState`
		// inside a class extending React.Component still reports.
		// Locks that `GetEnclosingReactComponent` traverses through static
		// blocks (which are not function-like and not a component scope).
		{
			Code: `
        class Hello extends React.Component {
          static {
            this.setState({});
          }
          render() { return <div/>; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},

		// ---- Edge: generic class component `class Hello extends React.Component<Props, State>` ----
		// TS type-argument list doesn't affect heritage detection.
		{
			Code: `
        class Hello extends React.Component<{}, {}> {
          someMethod() {
            this.setState({});
          }
          render() { return <div/>; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},

		// ---- Edge: HOC-wrapped class component (`withRouter(class extends React.Component {...})`) ----
		// The inner ClassExpression is what extends React.Component;
		// `GetEnclosingReactComponent` walks up from the setState call
		// and finds the inner class first — wrapper position is irrelevant.
		{
			Code: `
        const Hello = withRouter(class extends React.Component {
          someMethod() {
            this.setState({});
          }
          render() { return <div/>; }
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},

		// ---- Edge: setState inside an arrow callback nested inside the constructor ----
		// Walk: ArrowFunction (depth 1), Constructor (depth 2 — function-like),
		// ClassDeclaration. Reports on the ClassDeclaration (the React component).
		{
			Code: `
        class Hello extends React.Component {
          constructor(props) {
            super(props);
            setTimeout(() => this.setState({}), 100);
          }
          render() { return <div/>; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 5, Column: 30},
			},
		},

		// ---- Edge: functional-form setState `this.setState(prev => ({...}))` ----
		// Argument shape is irrelevant to the rule; the callee shape is what matters.
		{
			Code: `
        class Hello extends React.Component {
          someMethod() {
            this.setState(prev => ({ count: prev.count + 1 }));
          }
          render() { return <div/>; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},

		// ---- Edge: deeply-nested setState (Promise.then chain inside setTimeout inside method) ----
		// 4-deep nesting (setTimeout-arrow → Promise-callback → then-callback → setState).
		// All levels should still resolve to the same enclosing class.
		{
			Code: `
        class Hello extends React.Component {
          load() {
            setTimeout(() => {
              new Promise((resolve) => {
                fetch("/api").then((data) => {
                  this.setState({ data });
                  resolve(data);
                });
              });
            }, 100);
          }
          render() { return <div/>; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 7, Column: 19},
			},
		},

		// ---- Edge: setState as the right operand of `&&` ----
		// `cond && this.setState({})` — short-circuit gating. The rule is
		// listener-based, fires regardless of guarding context.
		{
			Code: `
        class Hello extends React.Component {
          someMethod() {
            cond && this.setState({});
          }
          render() { return <div/>; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 21},
			},
		},

		// ---- Edge: bare `extends PureComponent` (no React. prefix) ----
		// Locks `isComponentName` covers BOTH "Component" and "PureComponent"
		// in the bare-Identifier branch (line 1051 of reactutil.go).
		{
			Code: `
        class Hello extends PureComponent {
          someMethod() {
            this.setState({});
          }
          render() { return <div/>; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},

		// ---- Edge: custom pragma — `extends Preact.Component` IS a component ----
		// Inverse of the matching valid case above: with the same custom
		// pragma, `extends Preact.Component` IS recognized.
		{
			Code: `
        class Hello extends Preact.Component {
          someMethod() {
            this.setState({});
          }
          render() { return <div/>; }
        }
      `,
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"pragma": "Preact"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},

		// ---- Edge: custom createClass — `var X = myCreate({...})` IS a component ----
		// Inverse of the matching valid case above: with the same
		// `settings.react.createClass`, the custom factory call is
		// recognized as ES5 component.
		{
			Code: `
        var Hello = myCreate({
          someMethod: function() {
            this.setState({});
          },
          render: function() { return <div/>; }
        });
      `,
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"createClass": "myCreate"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},

		// ---- Real-world: `React.memo(class extends React.Component {...})` ----
		// memo wraps a class component. The class itself is the
		// detected ES6 component; the memo() call is just a wrapper at
		// the outer expression. setState inside the inner class still
		// reports. (Locks that wrapper-call positioning doesn't mask
		// the inner React-extending class.)
		{
			Code: `
        const Hello = React.memo(class extends React.Component {
          someMethod() {
            this.setState({});
          }
          render() { return <div/>; }
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},

		// ---- Real-world: redux `connect(...)(class extends React.Component {...})` ----
		// connect()(...) is a curried HOC. The inner class is what
		// extends React.Component. Same reasoning as memo above.
		{
			Code: `
        const Hello = connect(mapState)(class extends React.Component {
          someMethod() {
            this.setState({});
          }
          render() { return <div/>; }
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},

		// ---- Multi-component file: only the React component reports ----
		// Locks no-cross-fire across declarations: A's setState (not a
		// React class) is NOT reported; B's setState IS reported.
		{
			Code: `
        class A {
          foo() {
            this.setState({});
          }
        }
        class B extends React.Component {
          foo() {
            this.setState({});
          }
          render() { return <div/>; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 9, Column: 13},
			},
		},

		// ---- Real-world: addEventListener callback inside componentDidMount ----
		{
			Code: `
        class Hello extends React.Component {
          componentDidMount() {
            window.addEventListener("resize", () => {
              this.setState({});
            });
          }
          render() { return <div/>; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 5, Column: 15},
			},
		},

		// ---- Real-world: array.forEach + condition + setState ----
		{
			Code: `
        class Hello extends React.Component {
          process(items) {
            items.forEach((item) => {
              if (item.dirty) {
                this.setState({ value: item.v });
              }
            });
          }
          render() { return <div/>; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 6, Column: 17},
			},
		},

		// ---- Real-world: lifecycle hooks beyond componentDidMount ----
		// Each lifecycle method's setState reports independently.
		// The exact column is not the point of this test — what we lock
		// is "9 calls → 9 reports" (no de-dup, no skip across lifecycle
		// stoppers). Per-method column assertions are covered by the
		// dedicated single-method tests above.
		{
			Code: `
        class Hello extends React.Component {
          componentWillMount() {
            this.setState({});
          }
          componentDidMount() {
            this.setState({});
          }
          componentWillReceiveProps() {
            this.setState({});
          }
          shouldComponentUpdate() {
            this.setState({});
            return true;
          }
          componentWillUpdate() {
            this.setState({});
          }
          componentDidUpdate() {
            this.setState({});
          }
          componentWillUnmount() {
            this.setState({});
          }
          componentDidCatch() {
            this.setState({});
          }
          getSnapshotBeforeUpdate() {
            this.setState({});
            return null;
          }
          render() { return <div/>; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
				{MessageId: "noSetState", Line: 7, Column: 13},
				{MessageId: "noSetState", Line: 10, Column: 13},
				{MessageId: "noSetState", Line: 13, Column: 13},
				{MessageId: "noSetState", Line: 17, Column: 13},
				{MessageId: "noSetState", Line: 20, Column: 13},
				{MessageId: "noSetState", Line: 23, Column: 13},
				{MessageId: "noSetState", Line: 26, Column: 13},
				{MessageId: "noSetState", Line: 29, Column: 13},
			},
		},

		// ---- Real-world: Promise.then + setState in arrow ----
		{
			Code: `
        class Hello extends React.Component {
          load() {
            Promise.resolve().then(() => this.setState({}));
          }
          render() { return <div/>; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 42},
			},
		},

		// ---- Real-world: TS modifiers on the method (override / public / readonly arrow) ----
		// `override` requires TS >= 4.3, `readonly` is field-only — the
		// modifiers don't change the method's stopper / function-like
		// classification. Locks parity.
		{
			Code: `
        class Hello extends React.Component {
          public override componentDidMount(): void {
            this.setState({});
          }
          render() { return <div/>; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},

		// ---- Real-world: protected / private TS modifier ----
		{
			Code: `
        class Hello extends React.Component {
          protected someMethod() {
            this.setState({});
          }
          render() { return <div/>; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},

		// ---- Real-world: assignment-as-callee `(x = this.setState)({})` ----
		// `x = this.setState` is BinaryExpression(=) wrapped by
		// CallExpression. SkipParentheses doesn't unwrap binary ops, so
		// the callee Kind is BinaryExpression — not PropertyAccess.
		// Locked as VALID by the `cond ? this : x` shape above; here we
		// add the assignment variant for completeness. (Kept here as a
		// comment — overlaps with the conditional-receiver valid case.)

		// ---- Real-world: setState inside class method calling another method that calls setState ----
		// Indirect calls don't mask the direct setState call.
		{
			Code: `
        class Hello extends React.Component {
          someMethod() {
            this.setState({});
            this.otherMethod();
          }
          otherMethod() {
            this.setState({});
          }
          render() { return <div/>; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
				{MessageId: "noSetState", Line: 8, Column: 13},
			},
		},

		// ---- Real-world: setState inside a returned arrow (event-handler factory) ----
		{
			Code: `
        class Hello extends React.Component {
          makeHandler() {
            return () => this.setState({});
          }
          render() { return <button onClick={this.makeHandler()}/>; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 26},
			},
		},

		// ---- Real-world: ChainExpression-equivalent — `this?.x?.setState?.()` (deep optional) ----
		// `this?.x?.setState?.({})`: outermost call is `this?.x?.setState`
		// chained to `?.()`. After SkipParens, callee is the
		// PropertyAccessExpression `this?.x?.setState`. prop.Expression
		// is `this?.x` (PropertyAccessExpression), NOT ThisKeyword.
		// Locked as VALID via the `this.x.setState` valid case above —
		// here we add the optional-chain variant as a real-world shape.
		// (Kept as comment — overlaps with `this.x.setState` valid case.)

		// ---- Real-world: TS class with `implements` clause ----
		// `implements` is a sibling of `extends`; it doesn't affect
		// `ExtendsReactComponent`. setState still reports.
		{
			Code: `
        interface Foo {}
        class Hello extends React.Component implements Foo {
          someMethod() {
            this.setState({});
          }
          render() { return <div/>; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 5, Column: 13},
			},
		},

		// ---- Real-world: useEffect-style — setState inside setInterval callback ----
		{
			Code: `
        class Hello extends React.Component {
          componentDidMount() {
            this.timer = setInterval(() => {
              this.setState({ tick: Date.now() });
            }, 1000);
          }
          componentWillUnmount() { clearInterval(this.timer); }
          render() { return <div/>; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 5, Column: 15},
			},
		},

		// ---- Real-world: TS-only `as const` argument — argument shape doesn't matter ----
		{
			Code: `
        class Hello extends React.Component {
          someMethod() {
            this.setState({ x: 1 } as const);
          }
          render() { return <div/>; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},

		// ---- Real-world: setState inside a ternary expression ----
		{
			Code: `
        class Hello extends React.Component {
          someMethod() {
            cond ? this.setState({}) : null;
          }
          render() { return <div/>; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 20},
			},
		},

		// ---- Real-world: `await this.setState({})` inside async method ----
		// AwaitExpression is neither stopper nor function-like for the
		// component walk; the inner `this.setState({})` CallExpression
		// triggers the listener as usual.
		{
			Code: `
        class Hello extends React.Component {
          async load() {
            await this.setState({});
          }
          render() { return <div/>; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 19},
			},
		},

		// ---- Real-world: `void this.setState({})` ----
		// VoidExpression is a unary wrapper around the call. Same as
		// await — the inner CallExpression still fires.
		{
			Code: `
        class Hello extends React.Component {
          someMethod() {
            void this.setState({});
          }
          render() { return <div/>; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 18},
			},
		},

		// ---- Real-world: setState inside a `for` loop body ----
		{
			Code: `
        class Hello extends React.Component {
          process() {
            for (let i = 0; i < 10; i++) {
              this.setState({});
            }
          }
          render() { return <div/>; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 5, Column: 15},
			},
		},

		// ---- Real-world: setState inside a `while` loop body ----
		{
			Code: `
        class Hello extends React.Component {
          process() {
            while (cond) this.setState({});
          }
          render() { return <div/>; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 26},
			},
		},

		// ---- Real-world: TS legacy decorator + lifecycle method ----
		// Decorator is a sibling node on the MethodDeclaration; the
		// stopper match (function-like + class-extends-React) is
		// unaffected by decorator presence.
		{
			Code: `
        function readonlyDec(_t: any, _k: any, d: any) { return d; }
        class Hello extends React.Component {
          @readonlyDec
          componentDidMount() {
            this.setState({});
          }
          render() { return <div/>; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 6, Column: 13},
			},
		},

		// ---- Real-world: pragma-qualified createReactClass `React.createReactClass({...})` ----
		// upstream's `IsCreateClassCall` accepts both bare
		// `createReactClass(...)` and `<pragma>.<createClass>(...)` —
		// e.g. `React.createReactClass({...})` works under default settings.
		{
			Code: `
        var Hello = React.createReactClass({
          someMethod: function() {
            this.setState({});
          },
          render: function() { return <div/>; }
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},

		// ---- Real-world: named default-export class `export default class Foo extends React.Component` ----
		// Symmetric to the anonymous default-export case above; locks
		// that the class name is irrelevant to detection.
		{
			Code: `
        export default class Hello extends React.Component {
          someMethod() {
            this.setState({});
          }
          render() { return <div/>; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},

		// ---- Real-world: setState followed by chained `.then(callback)` ----
		// `this.setState({}).then(...)` — although React's setState
		// returns void, the textual call shape is preserved. The outer
		// `.then` is a separate CallExpression on a separate
		// PropertyAccess. Both setStates should report (the inner
		// one in this snippet, plus the optional second in callback).
		{
			Code: `
        class Hello extends React.Component {
          someMethod() {
            this.setState({}, () => this.setState({}));
          }
          render() { return <div/>; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
				{MessageId: "noSetState", Line: 4, Column: 37},
			},
		},

		// ---- Real-world: setState inside `do/while` body ----
		{
			Code: `
        class Hello extends React.Component {
          process() {
            do {
              this.setState({});
            } while (cond);
          }
          render() { return <div/>; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 5, Column: 15},
			},
		},

		// ---- Real-world: setState inside `for-of` loop body ----
		{
			Code: `
        class Hello extends React.Component {
          process(items) {
            for (const item of items) {
              this.setState({ value: item });
            }
          }
          render() { return <div/>; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 5, Column: 15},
			},
		},

		// ---- Real-world: `!this.setState({})` — unary-not on the call ----
		// Symmetric to `void` and `await` — the unary wrapper is on the
		// outer expression; the inner CallExpression still fires.
		{
			Code: `
        class Hello extends React.Component {
          someMethod() {
            !this.setState({});
          }
          render() { return <div/>; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 14},
			},
		},

		// ---- Real-world: forward-compat — Options array with future option strings ----
		// upstream `schema: []` accepts no options. We don't read
		// `options` either, so passing a future-shape value must be a
		// no-op (still report the call). Locks future-proofing.
		{
			Code: `
        class Hello extends React.Component {
          someMethod() {
            this.setState({});
          }
          render() { return <div/>; }
        }
      `,
			Tsx:     true,
			Options: []interface{}{"some-future-option"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},

		// ---- Real-world: empty Options array — same as default ----
		{
			Code: `
        class Hello extends React.Component {
          someMethod() {
            this.setState({});
          }
          render() { return <div/>; }
        }
      `,
			Tsx:     true,
			Options: []interface{}{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},

		// ---- Cross-helper lock: nested non-React class inside createReactClass arg ----
		// Empirically verified against ESLint master: ESLint's
		// `Components.set(node, ...)` walks `node.parent` to find any
		// node in the detected-component list. Non-React inner classes
		// are NOT in the list, so the walk passes through them and
		// reaches the outer createReactClass arg ObjectLiteral. Reports.
		// rslint's `GetEnclosingReactComponent` mirrors via the same
		// AST ancestor walk.
		{
			Code: `
var Hello = createReactClass({
  someMethod: function() {
    class Helper {
      doStuff() {
        this.setState({});
      }
    }
  },
  render: function() { return <div/>; }
});
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 6, Column: 9},
			},
		},

		// ---- Cross-helper lock: setState as createReactClass arg
		// top-level property value (not inside any method/function).
		// Empirically verified against ESLint master: components.set's
		// free parent walk attributes the call to the createReactClass
		// arg ObjectLiteral, so this reports. Locks the helper-layer
		// fix that removes the previous `seenEnclosingFunction` gate. ----
		{
			Code: `
var Hello = createReactClass({
  config: this.setState({}),
  render: function() { return <div/>; }
});
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 3, Column: 11},
			},
		},

		// ---- Cross-helper lock: nested non-React class inside an ES6
		// React component — same reasoning, walks past the non-React
		// inner class and reaches the outer React class. ----
		{
			Code: `
class Outer extends React.Component {
  render() {
    class Helper {
      doStuff() {
        this.setState({});
      }
    }
    return <div/>;
  }
}
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 6, Column: 9},
			},
		},

		// ---- Cross-helper lock: settings.componentWrapperFunctions ----
		// User configures `myObserver` as a custom component wrapper
		// (e.g. mobx-style HOC). Upstream's stateless detection
		// recognizes `myObserver(arrow)` as a stateless component when
		// the inner arrow returns JSX, and reports `this.setState({})`
		// inside it.
		//
		// rslint mirrors this via `GetComponentWrapperFunctions(settings, pragma)`
		// + `IsStatelessReactComponentWithWrappers`. Locked here.
		{
			Code: `
        const Hello = myObserver((props) => {
          this.setState({});
          return <div/>;
        });
      `,
			Tsx:      true,
			Settings: map[string]interface{}{"componentWrapperFunctions": []interface{}{"myObserver"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 3, Column: 11},
			},
		},

		// ---- Cross-helper lock: settings.componentWrapperFunctions object form ----
		// Object-shape entry `{property: "observer", object: "<pragma>"}`
		// matches `React.observer(arrow)` (and substitutes pragma per
		// configured `settings.react.pragma`).
		{
			Code: `
        const Hello = React.observer((props) => {
          this.setState({});
          return <div/>;
        });
      `,
			Tsx: true,
			Settings: map[string]interface{}{
				"componentWrapperFunctions": []interface{}{
					map[string]interface{}{"property": "observer", "object": "<pragma>"},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 3, Column: 11},
			},
		},

		// ---- Real-world: `throw this.setState({})` — ThrowStatement wrapping the call ----
		// ThrowStatement is neither stopper nor function-like. The
		// inner CallExpression still fires.
		{
			Code: `
        class Hello extends React.Component {
          someMethod() {
            throw this.setState({});
          }
          render() { return <div/>; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 19},
			},
		},

		// ---- Real-world: class decorator + method setState ----
		// Class decorators sit on the ClassDeclaration as a sibling
		// modifier; they don't change `extends` heritage detection.
		{
			Code: `
        function classDec<T extends new (...a: any[]) => any>(c: T) { return c; }
        @classDec
        class Hello extends React.Component {
          someMethod() {
            this.setState({});
          }
          render() { return <div/>; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 6, Column: 13},
			},
		},

		// ---- Real-world: multiple stacked decorators on the method ----
		{
			Code: `
        function dec1(_t: any, _k: any, d: any) { return d; }
        function dec2(_t: any, _k: any, d: any) { return d; }
        class Hello extends React.Component {
          @dec1
          @dec2
          componentDidMount() {
            this.setState({});
          }
          render() { return <div/>; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 8, Column: 13},
			},
		},

		// ---- Real-world: setState as JSX attribute value (immediate call, not callback) ----
		// `<div data-x={this.setState({})}/>` — the setState call is
		// the JSXExpression value, not wrapped in an arrow. Listener
		// fires on the call.
		{
			Code: `
        class Hello extends React.Component {
          render() {
            return <div data-x={this.setState({})}/>;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 33},
			},
		},

		// ---- Real-world: named function expression assigned to const ----
		// `const Hello = function Hello() { ... }` — the inner Identifier
		// `Hello` makes this a NamedFunctionExpression. Stateless
		// component detection should match the binding name.
		{
			Code: `
        const Hello = function Hello() {
          this.setState({});
          return <div/>;
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 3, Column: 11},
			},
		},

	})
}

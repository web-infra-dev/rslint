package no_access_state_in_setstate

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoAccessStateInSetstateRule(t *testing.T) {
	// Upstream's suite runs with `settings.react.createClass = 'createClass'`,
	// which wires the createReactClass matcher to `<pragma>.createClass(...)`
	// (e.g. `React.createClass`). rslint's default is `createReactClass`;
	// to keep the upstream test code verbatim we pass the same settings.
	legacyCreateClass := map[string]interface{}{
		"react": map[string]interface{}{"createClass": "createClass"},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoAccessStateInSetstateRule, []rule_tester.ValidTestCase{
		// ---- Upstream: setState called with callback (the correct form) ----
		{Code: `
        var Hello = React.createClass({
          onClick: function() {
            this.setState(state => ({value: state.value + 1}))
          }
        });
      `, Tsx: true, Settings: legacyCreateClass},

		// ---- Upstream: this.state read outside setState first arg ----
		{Code: `
        var Hello = React.createClass({
          multiplyValue: function(obj) {
            return obj.value*2
          },
          onClick: function() {
            var value = this.state.value
            this.multiplyValue({ value: value })
          }
        });
      `, Tsx: true, Settings: legacyCreateClass},

		// ---- Upstream: issue 1559 — IIFE with .call(this) inside render ----
		{Code: `
        var SearchForm = createReactClass({
          render: function () {
            return (
              <div>
                {(function () {
                  if (this.state.prompt) {
                    return <div>{this.state.prompt}</div>
                  }
                }).call(this)}
              </div>
            );
          }
        });
      `, Tsx: true},

		// ---- Upstream: issue 1604 — this.state inside setState's callback argument (2nd arg) is allowed ----
		{Code: `
        var Hello = React.createClass({
          onClick: function() {
            this.setState({}, () => console.log(this.state));
          }
        });
      `, Tsx: true, Settings: legacyCreateClass},

		// ---- Upstream: setState with empty object and no state access in callback ----
		{Code: `
        var Hello = React.createClass({
          onClick: function() {
            this.setState({}, () => 1 + 1);
          }
        });
      `, Tsx: true, Settings: legacyCreateClass},

		// ---- Upstream: var captured from this.state is unused; setState uses an unrelated var ----
		{Code: `
        var Hello = React.createClass({
          onClick: function() {
            var nextValueNotUsed = this.state.value + 1
            var nextValue = 2
            this.setState({value: nextValue})
          }
        });
      `, Tsx: true, Settings: legacyCreateClass},

		// ---- Upstream: unrelated top-level destructuring in a non-component function ----
		{Code: `
        function testFunction({a, b}) {
        };
      `, Tsx: true},

		// ---- Upstream: class field arrow that reads this.state outside setState ----
		{Code: `
        class ComponentA extends React.Component {
          state = {
            greeting: 'hello',
          };

          myFunc = () => {
            this.setState({ greeting: 'hi' }, () => this.doStuff());
          };

          doStuff = () => {
            console.log(this.state.greeting);
          };
        }
      `, Tsx: true},

		// ---- Upstream: call expression that takes this.state as arg but is not setState ----
		{Code: `
        class Foo extends Abstract {
          update = () => {
            const result = this.getResult ( this.state.foo );
            return this.setState ({ result });
          };
        }
      `, Tsx: true},

		// ---- Upstream: non-Component parent — rule does not fire at all ----
		{Code: `
        class StateContainer extends Container {
          anything() {
            return this.setState({value: this.state.value + 1})
          }
        };
      `, Tsx: true},

		// ---- Edge: this.state via bracket access is NOT matched (upstream's
		// `property.name === 'state'` gate excludes Literal property keys) ----
		{Code: `
        class Hello extends React.Component {
          onClick() {
            this.setState({value: this['state'].value + 1});
          }
        }
      `, Tsx: true},

		// ---- Edge: setState via bracket access is NOT matched (same gate
		// on the outer call) ----
		{Code: `
        class Hello extends React.Component {
          onClick() {
            this['setState']({value: this.state.value + 1});
          }
        }
      `, Tsx: true},

		// ---- Edge: renamed destructuring — upstream tracks variableName='state'
		// (property key), but subsequent use is `aliased`, which does not match ----
		{Code: `
        class Hello extends React.Component {
          onClick() {
            var {state: aliased} = this;
            this.setState({value: aliased.value + 1});
          }
        }
      `, Tsx: true},

		// ---- Edge: destructuring a local variable (not `this`) is NOT tracked ----
		{Code: `
        class Hello extends React.Component {
          onClick() {
            var obj = {state: {value: 1}};
            var {state} = obj;
            this.setState({value: state.value + 1});
          }
        }
      `, Tsx: true},

		// ---- Edge: setState with no arguments — no first-arg to search ----
		{Code: `
        class Hello extends React.Component {
          onClick() {
            this.setState();
          }
        }
      `, Tsx: true},

		// ---- Edge: this.state used only in a non-setState call's argument ----
		{Code: `
        class Hello extends React.Component {
          onClick() {
            console.log(this.state.value);
          }
        }
      `, Tsx: true},

		// ---- Edge: tracked method is called as `this.foo()` — upstream's
		// `'name' in callee` guard excludes MemberExpression callees, so the
		// propagation never triggers and no report is emitted. Mirrors the
		// upstream quirk rather than "fixing" it. ----
		{Code: `
        class Hello extends React.Component {
          nextState() { return this.state.value + 1 }
          onClick() {
            this.setState({value: this.nextState()});
          }
        }
      `, Tsx: true},

		// ---- Edge: this.state as an index key `obj[this.state]` — the
		// Identifier inside ElementAccessExpression.ArgumentExpression is in
		// ESTree's `property` position, not `object`, so the Identifier
		// listener correctly ignores it. Here we verify the PropertyAccess
		// listener also doesn't fire when wrapping `this.state.x[this.state]`
		// is read OUTSIDE a setState first arg. ----
		{Code: `
        class Hello extends React.Component {
          onClick() {
            var x = this.state.value;
            console.log(x);
          }
        }
      `, Tsx: true},

		// ---- Edge: class static block — NOT a function-body context that
		// can receive `this.setState`; walk reaches no container stopper and
		// falls off the top. ----
		{Code: `
        class Hello extends React.Component {
          static {
            const s = this.state;
            void s;
          }
        }
      `, Tsx: true},

		// ---- Edge: arrow-value object literal property — upstream gates the
		// FE branch on `FunctionExpression`, so ArrowFunctionExpression values
		// don't register as a container. We mirror that: tracking stops at
		// VariableDeclaration above the arrow (if any) or reaches program
		// root. ----
		{Code: `
        var Hello = createReactClass({
          onClick: () => {
            this.setState({ value: state.value + 1 });
          }
        });
      `, Tsx: true},

		// ---- Edge: inner class inside outer component — the inner class's
		// `nextState` does read this.state, but it isn't called from the
		// outer setState call, so no propagation fires. ----
		{Code: `
        class Outer extends React.Component {
          onClick() {
            class Inner { nextState() { return this.state; } }
            this.setState({ value: 1 });
          }
        }
      `, Tsx: true},

		// ---- Edge: computed-static class-body method key
		// (`['nextState']() {...}`) — upstream evaluates
		// `'name' in current.key ? current.key.name : undefined`; a
		// StringLiteral-wrapped computed key has no `.name`, so methodName
		// becomes undefined and no propagation ever fires against a real
		// callee string. We mirror this strictly (empty-string methodName
		// never matches any callee). ----
		{Code: `
        class Hello extends React.Component {
          ['nextState']() { return this.state.value + 1; }
          onClick() {
            this.setState({ value: nextState() });
          }
        }
      `, Tsx: true},

		// ---- Edge: template-literal-wrapped computed key — same reasoning
		// as the string-literal form. Locks in upstream-aligned non-match. ----
		{Code: "\n        class Hello extends React.Component {\n          [`nextState`]() { return this.state.value + 1; }\n          onClick() {\n            this.setState({ value: nextState() });\n          }\n        }\n      ", Tsx: true},

		// ---- Edge: string-literal key in object literal
		// (`'nextState': function() {...}`) — ESTree's Property.key is a
		// Literal, `'name' in key` is false, methodName undefined, no
		// propagation. ----
		{Code: `
        var Hello = createReactClass({
          'nextState': function() {
            return { value: this.state.value + 1 };
          },
          onClick: function() {
            this.setState(nextState());
          }
        });
      `, Tsx: true},

		// ---- Edge: destructuring with computed-static key
		// (`{ ['state']: local } = this`) — upstream's ObjectPattern listener
		// guards with `'name' in property.key`; a StringLiteral-wrapped
		// computed key has no `.name`, so the binding is not tracked. We
		// mirror strictly. ----
		{Code: `
        class Hello extends React.Component {
          onClick() {
            var { ['state']: state } = this;
            this.setState({ value: state.value + 1 });
          }
        }
      `, Tsx: true},

		// ---- Edge: destructuring with string-literal key — same reasoning. ----
		{Code: `
        class Hello extends React.Component {
          onClick() {
            var { 'state': state } = this;
            this.setState({ value: state.value + 1 });
          }
        }
      `, Tsx: true},

		// ---- Edge: block-scope precision — two sibling `if` blocks each
		// declare a `const nextValue`. Only the first block's `nextValue`
		// captures `this.state`; the second (used inside setState) is a
		// distinct binding. Symbol-identity matching correctly distinguishes
		// them (aligns with ESLint's block-scope manager), so no report
		// fires. ----
		{Code: `
        class Hello extends React.Component {
          onClick() {
            if (foo) {
              const nextValue = this.state.value + 1;
              void nextValue;
            }
            if (bar) {
              const nextValue = 2;
              this.setState({ value: nextValue });
            }
          }
        }
      `, Tsx: true},

		// ---- Edge: block-scope precision via destructuring — inner block
		// destructuring of `state` from `this`; outer setState uses a
		// same-named sibling binding that is NOT derived from `this`. The
		// two `state` bindings resolve to distinct symbols, so no match. ----
		{Code: `
        class Hello extends React.Component {
          onClick() {
            if (foo) {
              const { state } = this;
              void state;
            }
            const state = { value: 2 };
            this.setState({ value: state.value });
          }
        }
      `, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// ---- Upstream #1: direct this.state inside setState first arg ----
		{
			Code: `
        var Hello = React.createClass({
          onClick: function() {
            this.setState({value: this.state.value + 1})
          }
        });
      `,
			Tsx:      true,
			Settings: legacyCreateClass,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useCallback",
					Message:   "Use callback in setState when referencing the previous state.",
					Line:      4, Column: 35,
				},
			},
		},

		// ---- Upstream #2: this.state inside an arrow passed as first arg ----
		// Upstream reports this because the arrow IS args[0] and the
		// MemberExpression walk reaches the setState call via arg[0]. ----
		{
			Code: `
        var Hello = React.createClass({
          onClick: function() {
            this.setState(() => ({value: this.state.value + 1}))
          }
        });
      `,
			Tsx:      true,
			Settings: legacyCreateClass,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useCallback", Line: 4, Column: 42},
			},
		},

		// ---- Upstream #3: var captures this.state, used inside setState ----
		{
			Code: `
        var Hello = React.createClass({
          onClick: function() {
            var nextValue = this.state.value + 1
            this.setState({value: nextValue})
          }
        });
      `,
			Tsx:      true,
			Settings: legacyCreateClass,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useCallback", Line: 4, Column: 29},
			},
		},

		// ---- Upstream #4: destructuring from `this`, state used inside setState ----
		{
			Code: `
        var Hello = React.createClass({
          onClick: function() {
            var {state, ...rest} = this
            this.setState({value: state.value + 1})
          }
        });
      `,
			Tsx:      true,
			Settings: legacyCreateClass,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useCallback", Line: 4, Column: 18},
			},
		},

		// ---- Upstream #5: external nextState() consumes this.state ----
		// The `nextState` function itself is a top-level FunctionDeclaration,
		// so its `this.state` read isn't tracked per-method; but the
		// `this.state` passed as an argument to `nextState(this.state)` IS
		// inside the setState first arg, so the direct-access branch reports.
		{
			Code: `
        function nextState(state) {
          return {value: state.value + 1}
        }
        var Hello = React.createClass({
          onClick: function() {
            this.setState(nextState(this.state))
          }
        });
      `,
			Tsx:      true,
			Settings: legacyCreateClass,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useCallback", Line: 7, Column: 37},
			},
		},

		// ---- Upstream #6: this.state passed as the first argument to setState ----
		{
			Code: `
        var Hello = React.createClass({
          onClick: function() {
            this.setState(this.state, () => 1 + 1);
          }
        });
      `,
			Tsx:      true,
			Settings: legacyCreateClass,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useCallback", Line: 4, Column: 27},
			},
		},

		// ---- Upstream #7: this.state passed directly, plus this.state in
		// the 2nd-arg callback — only the first arg triggers a report
		// (upstream emits a single error). ----
		{
			Code: `
        var Hello = React.createClass({
          onClick: function() {
            this.setState(this.state, () => console.log(this.state));
          }
        });
      `,
			Tsx:      true,
			Settings: legacyCreateClass,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useCallback", Line: 4, Column: 27},
			},
		},

		// ---- Upstream #8: method in object literal reads this.state; another
		// method calls it inside setState. Mirrors upstream's FE-with-key
		// tracking branch. ----
		{
			Code: `
        var Hello = React.createClass({
          nextState: function() {
            return {value: this.state.value + 1}
          },
          onClick: function() {
            this.setState(nextState())
          }
        });
      `,
			Tsx:      true,
			Settings: legacyCreateClass,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useCallback", Line: 4, Column: 28},
			},
		},

		// ---- Upstream #9: this.state as first arg, plus 2nd-arg callback —
		// ES6 class variant. ----
		{
			Code: `
        class Hello extends React.Component {
          onClick() {
            this.setState(this.state, () => console.log(this.state));
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useCallback", Line: 4, Column: 27},
			},
		},

		// ---- Edge: class-body method tracking — class method reads this.state,
		// called inside setState by bare identifier (not this.foo()). ----
		{
			Code: `
        class Hello extends React.Component {
          nextState() {
            return this.state.value + 1;
          }
          onClick() {
            this.setState({value: nextState()});
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useCallback", Line: 4, Column: 20},
			},
		},

		// ---- Edge: this.state inside if-block inside setState first arg ----
		{
			Code: `
        class Hello extends React.Component {
          onClick() {
            this.setState({value: (() => { if (true) return this.state.value; return 0; })()});
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useCallback", Line: 4, Column: 61},
			},
		},

		// ---- Edge: const + let capture this.state, used in setState ----
		{
			Code: `
        class Hello extends React.Component {
          onClick() {
            const nextValue = this.state.value + 1;
            this.setState({value: nextValue});
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useCallback", Line: 4, Column: 31},
			},
		},

		// ---- Edge: parenthesized `(this).state` inside setState — tsgo
		// preserves the paren wrapper on the PropertyAccessExpression so the
		// reported span includes the leading `(`. Locked in to document the
		// divergence from ESLint (which strips parens). ----
		{
			Code: `
        class Hello extends React.Component {
          onClick() {
            this.setState({value: (this).state.value + 1});
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useCallback", Line: 4, Column: 35},
			},
		},

		// ---- Edge: optional chain `this?.state` in setState first arg —
		// PropertyAccessExpression Kind is the same, flag-only difference,
		// so the rule still matches. ----
		{
			Code: `
        class Hello extends React.Component {
          onClick() {
            this.setState({value: this?.state.value + 1});
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useCallback", Line: 4, Column: 35},
			},
		},

		// ---- Edge: two setState calls each with this.state — two reports ----
		{
			Code: `
        class Hello extends React.Component {
          onClick() {
            this.setState({value: this.state.value + 1});
            this.setState({value: this.state.value + 2});
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useCallback", Line: 4, Column: 35},
				{MessageId: "useCallback", Line: 5, Column: 35},
			},
		},

		// ---- Edge: destructuring with shorthand use in setState ----
		{
			Code: `
        class Hello extends React.Component {
          onClick() {
            var {state} = this;
            this.setState({state});
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useCallback", Line: 4, Column: 18},
			},
		},

		// ---- Edge: class expression (anonymous class assigned to const) ----
		{
			Code: `
        const Hello = class extends React.Component {
          onClick() {
            this.setState({value: this.state.value + 1});
          }
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useCallback", Line: 4, Column: 35},
			},
		},

		// ---- Edge: async method reads this.state inside setState first arg ----
		{
			Code: `
        class Hello extends React.Component {
          async onClick() {
            await Promise.resolve();
            this.setState({value: this.state.value + 1});
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useCallback", Line: 5, Column: 35},
			},
		},

		// ---- Edge: shorthand method in object literal reads this.state,
		// called from sibling method inside setState. Exercises the
		// MethodDeclaration-in-ObjectLiteralExpression branch of container
		// detection. ----
		{
			Code: `
        var Hello = createReactClass({
          nextState() {
            return { value: this.state.value + 1 };
          },
          onClick() {
            this.setState(nextState());
          }
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useCallback", Line: 4, Column: 29},
			},
		},

		// ---- Edge: two nested setState calls — `this.setState(this.setState(this.state))`.
		// The inner `this.state` reaches the INNER setState's first-arg
		// position first — the walk reports there and returns, so a single
		// diagnostic fires. ----
		{
			Code: `
        class Hello extends React.Component {
          onClick() {
            this.setState(this.setState(this.state));
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useCallback", Line: 4, Column: 41},
			},
		},

		// ---- Edge: destructuring with rest + aliased sibling — rest is
		// skipped (no PropertyName); aliased `state: x` tracks variableName
		// 'state' so a later `state` Identifier (matching original key) is
		// not found, but `x` is. Mirrors upstream's property-name-based
		// tracking which effectively keeps aliased bindings opaque. ----
		{
			Code: `
        class Hello extends React.Component {
          onClick() {
            var { state, ...rest } = this;
            this.setState({ state, rest });
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useCallback", Line: 4, Column: 19},
			},
		},

		// ---- Edge: this.state in arrow-body concise expression inside
		// setState first arg (arrow `() => this.state.x + 1`). Walk goes
		// through ArrowFunction → CallExpression (setState). ----
		{
			Code: `
        class Hello extends React.Component {
          onClick() {
            this.setState(() => this.state.value + 1);
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useCallback", Line: 4, Column: 33},
			},
		},

		// ---- Edge: nested setState-of-setState — outer is setState call
		// whose first arg is itself a setState call. The OUTER setState's
		// first-arg is the inner CallExpression; `this.state` inside the
		// inner setState's first arg reports AT the inner setState boundary
		// (walk matches inner first, returns). ----
		{
			Code: `
        class Hello extends React.Component {
          onClick() {
            this.setState(this.setState({ value: this.state.value }));
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useCallback", Line: 4, Column: 50},
			},
		},

		// ---- Edge: typed method signature `onClick(): void { ... }` — tsgo
		// still emits MethodDeclaration; type annotations don't disturb
		// container detection. ----
		{
			Code: `
        class Hello extends React.Component<{}, { value: number }> {
          onClick(): void {
            this.setState({ value: this.state.value + 1 });
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useCallback", Line: 4, Column: 36},
			},
		},
	})
}

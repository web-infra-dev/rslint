package no_will_update_set_state

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoWillUpdateSetStateRule(t *testing.T) {
	react162 := map[string]interface{}{"react": map[string]interface{}{"version": "16.2.0"}}
	react163 := map[string]interface{}{"react": map[string]interface{}{"version": "16.3.0"}}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoWillUpdateSetStateRule, []rule_tester.ValidTestCase{
		// ---- Upstream: createReactClass render only, no componentWillUpdate ----
		{Code: `
        var Hello = createReactClass({
          render: function() {
            return <div>Hello {this.props.name}</div>;
          }
        });
      `, Tsx: true},

		// ---- Upstream: empty componentWillUpdate body ----
		{Code: `
        var Hello = createReactClass({
          componentWillUpdate: function() {}
        });
      `, Tsx: true},

		// ---- Upstream: componentWillUpdate with non-member call and property assignment ----
		{Code: `
        var Hello = createReactClass({
          componentWillUpdate: function() {
            someNonMemberFunction(arg);
            this.someHandler = this.setState;
          }
        });
      `, Tsx: true},

		// ---- Upstream: setState inside a nested regular function (default mode) ----
		{Code: `
        var Hello = createReactClass({
          componentWillUpdate: function() {
            someClass.onSomeEvent(function(data) {
              this.setState({
                data: data
              });
            })
          }
        });
      `, Tsx: true},

		// ---- Upstream: setState inside a nested named function declaration (default mode) ----
		{Code: `
        var Hello = createReactClass({
          componentWillUpdate: function() {
            function handleEvent(data) {
              this.setState({
                data: data
              });
            }
            someClass.onSomeEvent(handleEvent)
          }
        });
      `, Tsx: true},

		// ---- Upstream: UNSAFE_componentWillUpdate is NOT matched when react < 16.3.0 ----
		{Code: `
        class Hello extends React.Component {
          UNSAFE_componentWillUpdate() {
            this.setState({
              data: data
            });
          }
        }
      `, Tsx: true, Settings: react162},

		// ---- Default-mode: setState inside nested arrow (depth > 1) ----
		{Code: `
        class Hello extends React.Component {
          componentWillUpdate() {
            someClass.onSomeEvent((data) => this.setState({data: data}));
          }
        }
      `, Tsx: true},

		// ---- Edge: setState receiver is not `this` ----
		{Code: `
        class Hello extends React.Component {
          componentWillUpdate() {
            other.setState({});
          }
        }
      `, Tsx: true},

		// ---- Edge: method name mismatch (componentDidUpdate, not componentWillUpdate) ----
		{Code: `
        class Hello extends React.Component {
          componentDidUpdate() {
            this.setState({});
          }
        }
      `, Tsx: true},

		// ---- Edge: computed key `[`componentWillUpdate`]()` — non-Identifier key never matches ----
		{Code: "\n        class Hello extends React.Component {\n          [`componentWillUpdate`]() {\n            this.setState({});\n          }\n        }\n      ", Tsx: true},

		// ---- Edge: string-literal key in object literal — non-Identifier key never matches ----
		{Code: `
        var Hello = createReactClass({
          "componentWillUpdate": function() {
            this.setState({});
          }
        });
      `, Tsx: true},

		// ---- Edge: bracketed setState access ('name' in property guard) ----
		{Code: `
        class Hello extends React.Component {
          componentWillUpdate() {
            this['setState']({});
          }
        }
      `, Tsx: true},

		// ---- Edge: top-level call outside any method/class/object ----
		{Code: `this.setState({});`, Tsx: true},

		// ---- Edge: setState in a sibling render() stays untouched — rule is method-scoped ----
		{Code: `
        class Hello extends React.Component {
          componentWillUpdate() {}
          render() {
            this.setState({});
            return <div/>;
          }
        }
      `, Tsx: true},

		// ---- Edge: setState inside JSX callback within componentWillUpdate — default mode allows (depth > 1) ----
		{Code: `
        class Hello extends React.Component {
          componentWillUpdate() {
            return <button onClick={() => this.setState({})}/>;
          }
        }
      `, Tsx: true},

		// ---- Edge: TS non-null / as expressions break the ThisKeyword match, so no report ----
		{Code: `
        class Hello extends React.Component {
          componentWillUpdate() {
            (this as any).setState({});
          }
        }
      `, Tsx: true},

		// ---- Edge: aliased setState via destructuring — receiver is not a MemberExpression ----
		{Code: `
        class Hello extends React.Component {
          componentWillUpdate() {
            const { setState } = this;
            setState({});
          }
        }
      `, Tsx: true},

		// ---- Edge: outer componentWillUpdate, setState inside an inner class's constructor (default mode) ----
		// Walk: Constructor (depth 1, stopper name="constructor" — no match), ClassDeclaration,
		// outer MethodDeclaration (depth 2, stopper match — skipped because depth > 1).
		{Code: `
        class Outer extends React.Component {
          componentWillUpdate() {
            class Inner { constructor() { this.setState({}); } }
            new Inner();
          }
        }
      `, Tsx: true},

		// ---- Edge: UNSAFE_ alias NOT matched when version explicitly pinned below 16.3.0 ----
		// Upstream's `shouldCheckUnsafeCb` returns false for < 16.3.0.
		{Code: `
        var Hello = createReactClass({
          UNSAFE_componentWillUpdate: function() {
            this.setState({});
          }
        });
      `, Tsx: true, Settings: react162},

		// ---- Edge: TS non-null assertion `this!.setState()` breaks ThisKeyword match ----
		// The NonNullExpression wraps `this`, not a ParenthesizedExpression, so
		// SkipParentheses leaves it intact. Matches upstream (TSNonNullExpression
		// is not a ThisExpression).
		{Code: `
        class Hello extends React.Component {
          componentWillUpdate() {
            this!.setState({});
          }
        }
      `, Tsx: true},

		// ---- Edge: UNSAFE_ alias NOT matched even under default mode when
		// called from a non-matching sibling method ----
		{Code: `
        class Hello extends React.Component {
          someOtherMethod() {
            this.setState({});
          }
        }
      `, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// ---- Upstream #1: createReactClass componentWillUpdate → setState ----
		{
			Code: `
        var Hello = createReactClass({
          componentWillUpdate: function() {
            this.setState({
              data: data
            });
          }
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noSetState",
					Message:   "Do not use setState in componentWillUpdate",
					Line:      4, Column: 13,
				},
			},
		},

		// ---- Upstream #2: ES6 class method componentWillUpdate → setState ----
		{
			Code: `
        class Hello extends React.Component {
          componentWillUpdate() {
            this.setState({
              data: data
            });
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noSetState",
					Message:   "Do not use setState in componentWillUpdate",
					Line:      4, Column: 13,
				},
			},
		},

		// ---- Upstream #3: disallow-in-func + createReactClass (direct call) ----
		{
			Code: `
        var Hello = createReactClass({
          componentWillUpdate: function() {
            this.setState({
              data: data
            });
          }
        });
      `,
			Tsx:     true,
			Options: []interface{}{"disallow-in-func"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},

		// ---- Upstream #4: disallow-in-func + ES6 class (direct call) ----
		{
			Code: `
        class Hello extends React.Component {
          componentWillUpdate() {
            this.setState({
              data: data
            });
          }
        }
      `,
			Tsx:     true,
			Options: []interface{}{"disallow-in-func"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},

		// ---- Upstream #5: disallow-in-func + setState in nested function (createReactClass) ----
		{
			Code: `
        var Hello = createReactClass({
          componentWillUpdate: function() {
            someClass.onSomeEvent(function(data) {
              this.setState({
                data: data
              });
            })
          }
        });
      `,
			Tsx:     true,
			Options: []interface{}{"disallow-in-func"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 5, Column: 15},
			},
		},

		// ---- Upstream #6: disallow-in-func + setState in nested function (ES6 class) ----
		{
			Code: `
        class Hello extends React.Component {
          componentWillUpdate() {
            someClass.onSomeEvent(function(data) {
              this.setState({
                data: data
              });
            })
          }
        }
      `,
			Tsx:     true,
			Options: []interface{}{"disallow-in-func"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 5, Column: 15},
			},
		},

		// ---- Upstream #7: setState inside if-block (createReactClass) ----
		{
			Code: `
        var Hello = createReactClass({
          componentWillUpdate: function() {
            if (true) {
              this.setState({
                data: data
              });
            }
          }
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 5, Column: 15},
			},
		},

		// ---- Upstream #8: setState inside if-block (ES6 class) ----
		{
			Code: `
        class Hello extends React.Component {
          componentWillUpdate() {
            if (true) {
              this.setState({
                data: data
              });
            }
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 5, Column: 15},
			},
		},

		// ---- Upstream #9: disallow-in-func + arrow callback (createReactClass) ----
		{
			Code: `
        var Hello = createReactClass({
          componentWillUpdate: function() {
            someClass.onSomeEvent((data) => this.setState({data: data}));
          }
        });
      `,
			Tsx:     true,
			Options: []interface{}{"disallow-in-func"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 45},
			},
		},

		// ---- Upstream #10: disallow-in-func + arrow callback (ES6 class) ----
		{
			Code: `
        class Hello extends React.Component {
          componentWillUpdate() {
            someClass.onSomeEvent((data) => this.setState({data: data}));
          }
        }
      `,
			Tsx:     true,
			Options: []interface{}{"disallow-in-func"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 45},
			},
		},

		// ---- Upstream #11: UNSAFE_componentWillUpdate matched when react >= 16.3.0 (class) ----
		{
			Code: `
        class Hello extends React.Component {
          UNSAFE_componentWillUpdate() {
            this.setState({
              data: data
            });
          }
        }
      `,
			Tsx:      true,
			Settings: react163,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noSetState",
					Message:   "Do not use setState in UNSAFE_componentWillUpdate",
					Line:      4, Column: 13,
				},
			},
		},

		// ---- Upstream #12: UNSAFE_componentWillUpdate matched when react >= 16.3.0 (createReactClass) ----
		{
			Code: `
        var Hello = createReactClass({
          UNSAFE_componentWillUpdate: function() {
            this.setState({
              data: data
            });
          }
        });
      `,
			Tsx:      true,
			Settings: react163,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noSetState",
					Message:   "Do not use setState in UNSAFE_componentWillUpdate",
					Line:      4, Column: 13,
				},
			},
		},

		// ---- Edge: parenthesized `(this).setState(...)` receiver ----
		{
			Code: `
        class Hello extends React.Component {
          componentWillUpdate() {
            (this).setState({});
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},

		// ---- Edge: optional-chain on `this` — `this?.setState()` ----
		{
			Code: `
        class Hello extends React.Component {
          componentWillUpdate() {
            this?.setState({});
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},

		// ---- Edge: optional-call `this.setState?.()` ----
		{
			Code: `
        class Hello extends React.Component {
          componentWillUpdate() {
            this.setState?.({});
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},

		// ---- Edge: object-literal shorthand method (tsgo: MethodDeclaration with no PropertyAssignment wrapper) ----
		{
			Code: `
        var Hello = {
          componentWillUpdate() {
            this.setState({});
          }
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},

		// ---- Edge: class field with arrow initializer ----
		{
			Code: `
        class Hello extends React.Component {
          componentWillUpdate = () => {
            this.setState({});
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},

		// ---- Edge: class field with function expression initializer ----
		{
			Code: `
        class Hello extends React.Component {
          componentWillUpdate = function() {
            this.setState({});
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},

		// ---- Edge: setState under explicit React version < 16.3 — rule active, UNSAFE_ NOT matched ----
		// This case targets the non-UNSAFE name, so it fires regardless of version.
		{
			Code: `
        class Hello extends React.Component {
          componentWillUpdate() {
            this.setState({});
          }
        }
      `,
			Tsx:      true,
			Settings: react162,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},

		// ---- Edge: multiple setState calls in same componentWillUpdate ----
		{
			Code: `
        class Hello extends React.Component {
          componentWillUpdate() {
            this.setState({});
            if (x) { this.setState({}); }
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
				{MessageId: "noSetState", Line: 5, Column: 22},
			},
		},

		// ---- Edge: async componentWillUpdate ----
		{
			Code: `
        class Hello extends React.Component {
          async componentWillUpdate() {
            await Promise.resolve();
            this.setState({});
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 5, Column: 13},
			},
		},

		// ---- Edge: generator componentWillUpdate ----
		{
			Code: `
        class Hello extends React.Component {
          *componentWillUpdate() {
            yield this.setState({});
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 19},
			},
		},

		// ---- Edge: class expression (not ClassDeclaration) ----
		{
			Code: `
        const Hello = class extends React.Component {
          componentWillUpdate() {
            this.setState({});
          }
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},

		// ---- Edge: arrow-function value in object literal (createReactClass-style) ----
		{
			Code: `
        var Hello = createReactClass({
          componentWillUpdate: () => {
            this.setState({});
          }
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},

		// ---- Edge: JSX callback + disallow-in-func (symmetric pair to the default-mode valid case above) ----
		{
			Code: `
        class Hello extends React.Component {
          componentWillUpdate() {
            return <button onClick={() => this.setState({})}/>;
          }
        }
      `,
			Tsx:     true,
			Options: []interface{}{"disallow-in-func"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 43},
			},
		},

		// ---- Edge: nested class whose inner method is also componentWillUpdate — inner takes precedence ----
		{
			Code: `
        class Outer extends React.Component {
          componentWillUpdate() {
            class Inner extends React.Component {
              componentWillUpdate() {
                this.setState({});
              }
            }
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 6, Column: 17},
			},
		},

		// ---- Edge: disallow-in-func — setState inside a nested arrow inside a nested function ----
		{
			Code: `
        class Hello extends React.Component {
          componentWillUpdate() {
            someClass.onSomeEvent(function() {
              Promise.resolve().then(() => this.setState({}));
            });
          }
        }
      `,
			Tsx:     true,
			Options: []interface{}{"disallow-in-func"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 5, Column: 44},
			},
		},

		// ---- Edge: getter named `componentWillUpdate` — stopper via IsMethodOrAccessor, depth 1 from the getter itself ----
		{
			Code: `
        class Hello extends React.Component {
          get componentWillUpdate() {
            this.setState({});
            return null;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},

		// ---- Edge: PrivateIdentifier stopper `#componentWillUpdate` ----
		{
			Code: `
        class Hello extends React.Component {
          #componentWillUpdate() {
            this.setState({});
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},

		// ---- Edge: PrivateIdentifier callee `this.#setState()` ----
		{
			Code: `
        class Hello extends React.Component {
          componentWillUpdate() {
            this.#setState({});
          }
          #setState(_x: unknown) {}
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},

		// ---- Edge: static componentWillUpdate (upstream doesn't discriminate static) ----
		{
			Code: `
        class Hello extends React.Component {
          static componentWillUpdate() {
            this.setState({});
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},

		// ---- Edge: setter named `componentWillUpdate` — stopper via IsMethodOrAccessor ----
		{
			Code: `
        class Hello extends React.Component {
          set componentWillUpdate(v: unknown) {
            this.setState({});
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},

		// ---- Edge: UNSAFE_ alias as object-literal shorthand method (default version) ----
		{
			Code: `
        var Hello = {
          UNSAFE_componentWillUpdate() {
            this.setState({});
          }
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noSetState",
					Message:   "Do not use setState in UNSAFE_componentWillUpdate",
					Line:      4, Column: 13,
				},
			},
		},

		// ---- Edge: UNSAFE_ nested inside componentWillUpdate — innermost stopper wins ----
		// Walk: inner MethodDeclaration UNSAFE_componentWillUpdate (depth 1, stopper match)
		// → reported against the inner alias, not the outer name.
		{
			Code: `
        class Outer extends React.Component {
          componentWillUpdate() {
            class Inner extends React.Component {
              UNSAFE_componentWillUpdate() {
                this.setState({});
              }
            }
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noSetState",
					Message:   "Do not use setState in UNSAFE_componentWillUpdate",
					Line:      6, Column: 17,
				},
			},
		},

		// ---- Edge: React version absent (defaults to 999.999.999 ≥ 16.3.0) — UNSAFE_ alias IS matched ----
		// Upstream's `testReactVersion` treats missing version as latest, so the
		// UNSAFE_ callback returns true and the alias fires.
		{
			Code: `
        class Hello extends React.Component {
          UNSAFE_componentWillUpdate() {
            this.setState({});
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noSetState",
					Message:   "Do not use setState in UNSAFE_componentWillUpdate",
					Line:      4, Column: 13,
				},
			},
		},
	})
}

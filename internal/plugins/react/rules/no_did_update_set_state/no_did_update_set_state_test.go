package no_did_update_set_state

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoDidUpdateSetStateRule(t *testing.T) {
	// Shared helper: upstream's test harness takes every invalid case and
	// re-asserts it as valid once `settings.react.version` is "16.3.0", because
	// the upstream rule becomes a no-op from that version on (see
	// `shouldBeNoop`). We mirror that — feeding the same code paths through
	// the version-gate so a future regression on the gate is caught.
	react163 := map[string]interface{}{"react": map[string]interface{}{"version": "16.3.0"}}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoDidUpdateSetStateRule, []rule_tester.ValidTestCase{
		// ---- Upstream: createReactClass render only, no componentDidUpdate ----
		{Code: `
        var Hello = createReactClass({
          render: function() {
            return <div>Hello {this.props.name}</div>;
          }
        });
      `, Tsx: true},

		// ---- Upstream: empty componentDidUpdate body ----
		{Code: `
        var Hello = createReactClass({
          componentDidUpdate: function() {}
        });
      `, Tsx: true},

		// ---- Upstream: componentDidUpdate with non-member call and property assignment ----
		{Code: `
        var Hello = createReactClass({
          componentDidUpdate: function() {
            someNonMemberFunction(arg);
            this.someHandler = this.setState;
          }
        });
      `, Tsx: true},

		// ---- Upstream: setState inside a nested regular function (default mode) ----
		{Code: `
        var Hello = createReactClass({
          componentDidUpdate: function() {
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
          componentDidUpdate: function() {
            function handleEvent(data) {
              this.setState({
                data: data
              });
            }
            someClass.onSomeEvent(handleEvent)
          }
        });
      `, Tsx: true},

		// ---- Upstream: setState inside nested arrow (default mode) ----
		// Default mode allows setState in nested functions (depth > 1), so these
		// mirror upstream's `disallow-in-func`-only invalids being valid by default.
		{Code: `
        var Hello = createReactClass({
          componentDidUpdate: function() {
            someClass.onSomeEvent((data) => this.setState({data: data}));
          }
        });
      `, Tsx: true},
		{Code: `
        class Hello extends React.Component {
          componentDidUpdate() {
            someClass.onSomeEvent((data) => this.setState({data: data}));
          }
        }
      `, Tsx: true},

		// ---- Upstream: version-gated — every invalid case becomes valid at react >= 16.3.0 ----
		{Code: `
        var Hello = createReactClass({
          componentDidUpdate: function() {
            this.setState({
              data: data
            });
          }
        });
      `, Tsx: true, Settings: react163},
		{Code: `
        class Hello extends React.Component {
          componentDidUpdate() {
            this.setState({
              data: data
            });
          }
        }
      `, Tsx: true, Settings: react163},
		{Code: `
        class Hello extends React.Component {
          componentDidUpdate = () => {
            this.setState({
              data: data
            });
          }
        }
      `, Tsx: true, Settings: react163},
		{Code: `
        var Hello = createReactClass({
          componentDidUpdate: function() {
            this.setState({
              data: data
            });
          }
        });
      `, Tsx: true, Options: []interface{}{"disallow-in-func"}, Settings: react163},
		{Code: `
        class Hello extends React.Component {
          componentDidUpdate() {
            this.setState({
              data: data
            });
          }
        }
      `, Tsx: true, Options: []interface{}{"disallow-in-func"}, Settings: react163},
		{Code: `
        var Hello = createReactClass({
          componentDidUpdate: function() {
            someClass.onSomeEvent(function(data) {
              this.setState({
                data: data
              });
            })
          }
        });
      `, Tsx: true, Options: []interface{}{"disallow-in-func"}, Settings: react163},
		{Code: `
        class Hello extends React.Component {
          componentDidUpdate() {
            someClass.onSomeEvent(function(data) {
              this.setState({
                data: data
              });
            })
          }
        }
      `, Tsx: true, Options: []interface{}{"disallow-in-func"}, Settings: react163},
		{Code: `
        var Hello = createReactClass({
          componentDidUpdate: function() {
            if (true) {
              this.setState({
                data: data
              });
            }
          }
        });
      `, Tsx: true, Settings: react163},
		{Code: `
        class Hello extends React.Component {
          componentDidUpdate() {
            if (true) {
              this.setState({
                data: data
              });
            }
          }
        }
      `, Tsx: true, Settings: react163},
		{Code: `
        var Hello = createReactClass({
          componentDidUpdate: function() {
            someClass.onSomeEvent((data) => this.setState({data: data}));
          }
        });
      `, Tsx: true, Options: []interface{}{"disallow-in-func"}, Settings: react163},
		{Code: `
        class Hello extends React.Component {
          componentDidUpdate() {
            someClass.onSomeEvent((data) => this.setState({data: data}));
          }
        }
      `, Tsx: true, Options: []interface{}{"disallow-in-func"}, Settings: react163},

		// ---- Edge: setState receiver is not `this` ----
		{Code: `
        class Hello extends React.Component {
          componentDidUpdate() {
            other.setState({});
          }
        }
      `, Tsx: true},

		// ---- Edge: method name mismatch (componentDidMount, not componentDidUpdate) ----
		{Code: `
        class Hello extends React.Component {
          componentDidMount() {
            this.setState({});
          }
        }
      `, Tsx: true},

		// ---- Edge: UNSAFE_componentDidUpdate is NOT matched (factory passed no shouldCheckUnsafeCb) ----
		{Code: `
        class Hello extends React.Component {
          UNSAFE_componentDidUpdate() {
            this.setState({});
          }
        }
      `, Tsx: true},

		// ---- Edge: computed key `[\`componentDidUpdate\`]()` — non-Identifier key never matches ----
		{Code: "\n        class Hello extends React.Component {\n          [`componentDidUpdate`]() {\n            this.setState({});\n          }\n        }\n      ", Tsx: true},

		// ---- Edge: string-literal key in object literal — non-Identifier key never matches ----
		{Code: `
        var Hello = createReactClass({
          "componentDidUpdate": function() {
            this.setState({});
          }
        });
      `, Tsx: true},

		// ---- Edge: bracketed setState access ('name' in property guard) ----
		{Code: `
        class Hello extends React.Component {
          componentDidUpdate() {
            this['setState']({});
          }
        }
      `, Tsx: true},

		// ---- Edge: top-level call outside any method/class/object ----
		{Code: `this.setState({});`, Tsx: true},

		// ---- Edge: setState in a sibling render() stays untouched — rule is method-scoped ----
		{Code: `
        class Hello extends React.Component {
          componentDidUpdate() {}
          render() {
            this.setState({});
            return <div/>;
          }
        }
      `, Tsx: true},

		// ---- Edge: setState inside JSX callback within componentDidUpdate — default mode allows (depth > 1) ----
		{Code: `
        class Hello extends React.Component {
          componentDidUpdate() {
            return <button onClick={() => this.setState({})}/>;
          }
        }
      `, Tsx: true},

		// ---- Edge: TS non-null / as expressions break the ThisKeyword match, so no report ----
		{Code: `
        class Hello extends React.Component {
          componentDidUpdate() {
            (this as any).setState({});
          }
        }
      `, Tsx: true},

		// ---- Edge: aliased setState via destructuring — receiver is not a MemberExpression ----
		{Code: `
        class Hello extends React.Component {
          componentDidUpdate() {
            const { setState } = this;
            setState({});
          }
        }
      `, Tsx: true},

		// ---- Edge: outer componentDidUpdate, setState inside an inner class's constructor (default mode) ----
		// Walk: Constructor (depth 1, stopper name="constructor" — no match), ClassDeclaration,
		// outer MethodDeclaration (depth 2, stopper match — skipped because depth > 1).
		{Code: `
        class Outer extends React.Component {
          componentDidUpdate() {
            class Inner { constructor() { this.setState({}); } }
            new Inner();
          }
        }
      `, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// ---- Upstream #1: createReactClass componentDidUpdate → setState ----
		{
			Code: `
        var Hello = createReactClass({
          componentDidUpdate: function() {
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
					Message:   "Do not use setState in componentDidUpdate",
					Line:      4, Column: 13,
				},
			},
		},

		// ---- Upstream #2: ES6 class method componentDidUpdate → setState ----
		{
			Code: `
        class Hello extends React.Component {
          componentDidUpdate() {
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
					Line:      4, Column: 13,
				},
			},
		},

		// ---- Upstream #3: class field arrow `componentDidUpdate = () => { ... }` ----
		{
			Code: `
        class Hello extends React.Component {
          componentDidUpdate = () => {
            this.setState({
              data: data
            });
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},

		// ---- Upstream #4: disallow-in-func + createReactClass (direct call) ----
		{
			Code: `
        var Hello = createReactClass({
          componentDidUpdate: function() {
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

		// ---- Upstream #5: disallow-in-func + ES6 class (direct call) ----
		{
			Code: `
        class Hello extends React.Component {
          componentDidUpdate() {
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

		// ---- Upstream #6: disallow-in-func + setState in nested function (createReactClass) ----
		{
			Code: `
        var Hello = createReactClass({
          componentDidUpdate: function() {
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

		// ---- Upstream #7: disallow-in-func + setState in nested function (ES6 class) ----
		{
			Code: `
        class Hello extends React.Component {
          componentDidUpdate() {
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

		// ---- Upstream #8: setState inside if-block (createReactClass) ----
		{
			Code: `
        var Hello = createReactClass({
          componentDidUpdate: function() {
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

		// ---- Upstream #9: setState inside if-block (ES6 class) ----
		{
			Code: `
        class Hello extends React.Component {
          componentDidUpdate() {
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

		// ---- Upstream #10: disallow-in-func + arrow callback (createReactClass) ----
		{
			Code: `
        var Hello = createReactClass({
          componentDidUpdate: function() {
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

		// ---- Upstream #11: disallow-in-func + arrow callback (ES6 class) ----
		{
			Code: `
        class Hello extends React.Component {
          componentDidUpdate() {
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

		// ---- Edge: parenthesized `(this).setState(...)` receiver ----
		{
			Code: `
        class Hello extends React.Component {
          componentDidUpdate() {
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
          componentDidUpdate() {
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
          componentDidUpdate() {
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
          componentDidUpdate() {
            this.setState({});
          }
        };
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
          componentDidUpdate = function() {
            this.setState({});
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},

		// ---- Edge: setState under explicit React version < 16.3 (rule still active) ----
		{
			Code: `
        class Hello extends React.Component {
          componentDidUpdate() {
            this.setState({});
          }
        }
      `,
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"version": "16.2.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},

		// ---- Edge: multiple setState calls in same componentDidUpdate ----
		{
			Code: `
        class Hello extends React.Component {
          componentDidUpdate() {
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

		// ---- Edge: async componentDidUpdate ----
		{
			Code: `
        class Hello extends React.Component {
          async componentDidUpdate() {
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

		// ---- Edge: generator componentDidUpdate ----
		{
			Code: `
        class Hello extends React.Component {
          *componentDidUpdate() {
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
          componentDidUpdate() {
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
          componentDidUpdate: () => {
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
          componentDidUpdate() {
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

		// ---- Edge: nested class whose inner method is also componentDidUpdate — inner takes precedence ----
		// Walk from the setState in Inner's method: Inner MethodDeclaration (depth 1,
		// stopper match) → report at depth 1, regardless of disallow-in-func mode.
		{
			Code: `
        class Outer extends React.Component {
          componentDidUpdate() {
            class Inner extends React.Component {
              componentDidUpdate() {
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
		// Walk: ArrowFunction (1), FunctionExpression (2), MethodDeclaration (3, stopper match).
		// Default mode: depth > 1 → skip. disallow-in-func: report.
		{
			Code: `
        class Hello extends React.Component {
          componentDidUpdate() {
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

		// ---- Edge: getter named `componentDidUpdate` — stopper via IsMethodOrAccessor, depth 1 from the getter itself ----
		{
			Code: `
        class Hello extends React.Component {
          get componentDidUpdate() {
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

		// ---- Edge: PrivateIdentifier stopper `#componentDidUpdate` ----
		// ESLint's PrivateIdentifier.name is spec'd to exclude the leading #,
		// so `key.name === 'componentDidUpdate'` matches. Aligning with
		// upstream: we strip the # from tsgo's PrivateIdentifier.Text.
		{
			Code: `
        class Hello extends React.Component {
          #componentDidUpdate() {
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
		// Same reasoning: upstream's `'name' in property && property.name
		// === 'setState'` matches PrivateIdentifier with spec-stripped name.
		{
			Code: `
        class Hello extends React.Component {
          componentDidUpdate() {
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

		// ---- Edge: React version explicitly set to sentinel "999.999.999" — rule active ----
		// Upstream `shouldBeNoop` has an explicit `!testReactVersion('>= 999.999.999')`
		// upper bound — a user who pins that sentinel is treated as "version
		// unpinned" and the rule stays active.
		{
			Code: `
        class Hello extends React.Component {
          componentDidUpdate() {
            this.setState({});
          }
        }
      `,
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"version": "999.999.999"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},
	})
}

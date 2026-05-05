package no_did_mount_set_state

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoDidMountSetStateRule(t *testing.T) {
	// Shared helper: upstream's test harness takes every invalid case and
	// re-asserts it as valid once `settings.react.version` is "16.3.0", because
	// the upstream rule becomes a no-op from that version on (see
	// `shouldBeNoop`). We mirror that — feeding the same code paths through
	// the version-gate so a future regression on the gate is caught.
	react163 := map[string]interface{}{"react": map[string]interface{}{"version": "16.3.0"}}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoDidMountSetStateRule, []rule_tester.ValidTestCase{
		// ---- Upstream: createReactClass render only, no componentDidMount ----
		{Code: `
        var Hello = createReactClass({
          render: function() {
            return <div>Hello {this.props.name}</div>;
          }
        });
      `, Tsx: true},

		// ---- Upstream: empty componentDidMount body ----
		{Code: `
        var Hello = createReactClass({
          componentDidMount: function() {}
        });
      `, Tsx: true},

		// ---- Upstream: componentDidMount with non-member call and property assignment ----
		{Code: `
        var Hello = createReactClass({
          componentDidMount: function() {
            someNonMemberFunction(arg);
            this.someHandler = this.setState;
          }
        });
      `, Tsx: true},

		// ---- Upstream: setState inside a nested regular function (default mode) ----
		{Code: `
        var Hello = createReactClass({
          componentDidMount: function() {
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
          componentDidMount: function() {
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
          componentDidMount: function() {
            someClass.onSomeEvent((data) => this.setState({data: data}));
          }
        });
      `, Tsx: true},
		{Code: `
        class Hello extends React.Component {
          componentDidMount() {
            someClass.onSomeEvent((data) => this.setState({data: data}));
          }
        }
      `, Tsx: true},

		// ---- Upstream: version-gated — every invalid case becomes valid at react >= 16.3.0 ----
		{Code: `
        var Hello = createReactClass({
          componentDidMount: function() {
            this.setState({
              data: data
            });
          }
        });
      `, Tsx: true, Settings: react163},
		{Code: `
        class Hello extends React.Component {
          componentDidMount() {
            this.setState({
              data: data
            });
          }
        }
      `, Tsx: true, Settings: react163},
		{Code: `
        class Hello extends React.Component {
          componentDidMount = () => {
            this.setState({
              data: data
            });
          }
        }
      `, Tsx: true, Settings: react163},
		{Code: `
        var Hello = createReactClass({
          componentDidMount: function() {
            this.setState({
              data: data
            });
          }
        });
      `, Tsx: true, Options: []interface{}{"disallow-in-func"}, Settings: react163},
		{Code: `
        class Hello extends React.Component {
          componentDidMount() {
            this.setState({
              data: data
            });
          }
        }
      `, Tsx: true, Options: []interface{}{"disallow-in-func"}, Settings: react163},
		{Code: `
        var Hello = createReactClass({
          componentDidMount: function() {
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
          componentDidMount() {
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
          componentDidMount: function() {
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
          componentDidMount() {
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
          componentDidMount: function() {
            someClass.onSomeEvent((data) => this.setState({data: data}));
          }
        });
      `, Tsx: true, Options: []interface{}{"disallow-in-func"}, Settings: react163},
		{Code: `
        class Hello extends React.Component {
          componentDidMount() {
            someClass.onSomeEvent((data) => this.setState({data: data}));
          }
        }
      `, Tsx: true, Options: []interface{}{"disallow-in-func"}, Settings: react163},

		// ---- Edge: setState receiver is not `this` ----
		{Code: `
        class Hello extends React.Component {
          componentDidMount() {
            other.setState({});
          }
        }
      `, Tsx: true},

		// ---- Edge: method name mismatch (componentDidUpdate, not componentDidMount) ----
		{Code: `
        class Hello extends React.Component {
          componentDidUpdate() {
            this.setState({});
          }
        }
      `, Tsx: true},

		// ---- Edge: UNSAFE_componentDidMount is NOT matched (factory passed no shouldCheckUnsafeCb) ----
		{Code: `
        class Hello extends React.Component {
          UNSAFE_componentDidMount() {
            this.setState({});
          }
        }
      `, Tsx: true},

		// ---- Edge: computed key `[\`componentDidMount\`]()` — non-Identifier key never matches ----
		{Code: "\n        class Hello extends React.Component {\n          [`componentDidMount`]() {\n            this.setState({});\n          }\n        }\n      ", Tsx: true},

		// ---- Edge: string-literal key in object literal — non-Identifier key never matches ----
		{Code: `
        var Hello = createReactClass({
          "componentDidMount": function() {
            this.setState({});
          }
        });
      `, Tsx: true},

		// ---- Edge: bracketed setState access ('name' in property guard) ----
		{Code: `
        class Hello extends React.Component {
          componentDidMount() {
            this['setState']({});
          }
        }
      `, Tsx: true},

		// ---- Edge: top-level call outside any method/class/object ----
		{Code: `this.setState({});`, Tsx: true},

		// ---- Edge: setState in a sibling render() stays untouched — rule is method-scoped ----
		{Code: `
        class Hello extends React.Component {
          componentDidMount() {}
          render() {
            this.setState({});
            return <div/>;
          }
        }
      `, Tsx: true},

		// ---- Edge: setState inside JSX callback within componentDidMount — default mode allows (depth > 1) ----
		{Code: `
        class Hello extends React.Component {
          componentDidMount() {
            return <button onClick={() => this.setState({})}/>;
          }
        }
      `, Tsx: true},

		// ---- Edge: TS non-null / as expressions break the ThisKeyword match, so no report ----
		{Code: `
        class Hello extends React.Component {
          componentDidMount() {
            (this as any).setState({});
          }
        }
      `, Tsx: true},

		// ---- Edge: aliased setState via destructuring — receiver is not a MemberExpression ----
		{Code: `
        class Hello extends React.Component {
          componentDidMount() {
            const { setState } = this;
            setState({});
          }
        }
      `, Tsx: true},

		// ---- Edge: outer componentDidMount, setState inside an inner class's constructor (default mode) ----
		// Walk: Constructor (depth 1, stopper name="constructor" — no match), ClassDeclaration,
		// outer MethodDeclaration (depth 2, stopper match — skipped because depth > 1).
		{Code: `
        class Outer extends React.Component {
          componentDidMount() {
            class Inner { constructor() { this.setState({}); } }
            new Inner();
          }
        }
      `, Tsx: true},

		// ---- rslint hardening: real-world scopes & tsgo shape boundaries ----
		// Lock that NonNullExpression on `this` breaks the ThisKeyword check
		// (SkipParentheses unwraps Parens only, not `!` / `as` / `satisfies`).
		{Code: `
        class Hello extends React.Component {
          componentDidMount() {
            this!.setState({});
          }
        }
      `, Tsx: true},
		// Lock that SatisfiesExpression on `this` breaks the ThisKeyword check.
		{Code: `
        class Hello extends React.Component {
          componentDidMount() {
            (this satisfies React.Component).setState({});
          }
        }
      `, Tsx: true},
		// `super.setState(...)` inside componentDidMount — receiver is SuperKeyword, not ThisKeyword.
		{Code: `
        class Hello extends React.Component {
          componentDidMount() {
            super.setState({});
          }
        }
      `, Tsx: true},
		// Indirect call: `this.setState.call(this, ...)` — outer call's prop.Expression is
		// `this.setState` (PropertyAccessExpression), not ThisKeyword, so no match.
		// `.bind(...)` etc. follow the same shape.
		{Code: `
        class Hello extends React.Component {
          componentDidMount() {
            this.setState.call(this, {});
            this.setState.bind(this);
            this.setState.apply(this, [{}]);
          }
        }
      `, Tsx: true},
		// Tagged template (TaggedTemplateExpression, not CallExpression — listener never fires).
		{Code: "\n        class Hello extends React.Component {\n          componentDidMount() {\n            this.setState`{}`;\n          }\n        }\n      ", Tsx: true},
		// `new this.setState(...)` — NewExpression, not CallExpression.
		{Code: `
        class Hello extends React.Component {
          componentDidMount() {
            new this.setState({});
          }
        }
      `, Tsx: true},
		// Comma operator callee: `(0, this.setState)(...)` — callee is BinaryExpression, not PropertyAccess.
		{Code: `
        class Hello extends React.Component {
          componentDidMount() {
            (0, this.setState)({});
          }
        }
      `, Tsx: true},
		// Static block: ClassStaticBlockDeclaration is NOT a stopper and NOT IsFunctionLikeDeclaration.
		// Walk reaches no stopper above it (the sibling componentDidMount is unrelated).
		{Code: `
        class Hello extends React.Component {
          componentDidMount() {}
          static {
            this.setState({});
          }
        }
      `, Tsx: true},
		// Top-level free-standing `function componentDidMount() { ... }` — FunctionDeclaration is
		// NOT a stopper (only Property/Method/ClassProperty/PropertyDefinition are). Walk reaches
		// root with no stopper found. Locks that the rule is stopper-keyed, not name-on-function-keyed.
		{Code: `
        function componentDidMount() {
          this.setState({});
        }
      `, Tsx: true},
		// `setState` reference without call — no CallExpression listener fires for bare reference.
		{Code: `
        class Hello extends React.Component {
          componentDidMount() {
            const fn = this.setState;
          }
        }
      `, Tsx: true},
		// React version <16.3 with patch & pre-release-ish form — rule remains active so we
		// expect the call to be flagged; tested separately under invalid below for setState.
		// Here we lock the inverse: a noop version 16.4.0 (well above 16.3.0) keeps it valid.
		{Code: `
        class Hello extends React.Component {
          componentDidMount() {
            this.setState({});
          }
        }
      `, Tsx: true, Settings: map[string]interface{}{"react": map[string]interface{}{"version": "16.4.0"}}},
		// settings.react is not a map (e.g. boolean). reactVersionNoop must not crash and the rule
		// stays active (no version pinned ⇒ same as unset ⇒ rule active under our gate). Tested
		// separately under invalid because the call should still flag.

		// ConditionalExpression as receiver — `SkipParens` doesn't unwrap `?:`, so the
		// `prop.Expression` is a ConditionalExpression, not ThisKeyword. No match.
		{Code: `
        class Hello extends React.Component {
          componentDidMount() {
            (cond ? this : other).setState({});
          }
        }
      `, Tsx: true},
		// LogicalOr receiver — BinaryExpression, not ThisKeyword. No match.
		{Code: `
        class Hello extends React.Component {
          componentDidMount() {
            (this || other).setState({});
          }
        }
      `, Tsx: true},
		// (TS legacy `<T>expr` type assertion is ambiguous with JSX in TSX mode and
		// not relevant here; the `as any` form already covers the wrapper-breaks-ThisKeyword
		// behavior in the existing valid set.)
		// Sibling method (non-componentDidMount) in the same class containing setState —
		// must NOT match: walk reaches MethodDeclaration `other` (not the matched name)
		// then climbs past the class, reaching root with no stopper match.
		{Code: `
        class Hello extends React.Component {
          componentDidMount() {}
          other() {
            this.setState({});
          }
        }
      `, Tsx: true},
		// disallow-in-func option spelled in wrong case — upstream's enum is exact, so
		// the value is treated as not-disallow-in-func and default mode applies. The
		// nested-callback setState below should therefore stay valid.
		{Code: `
        class Hello extends React.Component {
          componentDidMount() {
            someClass.onSomeEvent(function(data) {
              this.setState({ data: data });
            });
          }
        }
      `, Tsx: true, Options: []interface{}{"DISALLOW-IN-FUNC"}},

		// Empirically-verified rslint divergence: pre-release React version
		// "16.3.0-rc.1" is parsed by reactutil.ParseReactVersion as (16, 3, 0),
		// which makes the noop gate fire (range is [16.3.0, 999.999.999)). Upstream
		// uses semver where "16.3.0-rc.1" ranks BELOW "16.3.0", so upstream stays
		// active. Locked as VALID here — this is a Phase-1-Step-6.B language-natural
		// divergence stemming from rslint's integer-tuple version parsing (shared
		// across every version-gated React rule). See the rule's `.md` "Differences
		// from ESLint" section.
		{Code: `
        class Hello extends React.Component {
          componentDidMount() {
            this.setState({});
          }
        }
      `, Tsx: true, Settings: map[string]interface{}{"react": map[string]interface{}{"version": "16.3.0-rc.1"}}},

		// Empirically-verified: `with` statement makes a bare `setState(...)` call
		// resolve to `this.setState` at runtime, but textually the callee is a bare
		// Identifier, not a PropertyAccessExpression on ThisKeyword. Both upstream
		// and rslint require the explicit `this.` — neither matches.
		{Code: `
        // @ts-nocheck
        class Hello extends React.Component {
          componentDidMount() {
            with (this as any) {
              setState({});
            }
          }
        }
      `, Tsx: true},

		// ---- Real-world async / callback patterns (all empirically verified) ----
		// IIFE inside componentDidMount — depth=2 in default mode, allowed.
		{Code: `
        class Hello extends React.Component {
          componentDidMount() {
            (function() {
              this.setState({});
            })();
          }
        }
      `, Tsx: true},
		// setTimeout callback (default mode) — depth=2.
		{Code: `
        class Hello extends React.Component {
          componentDidMount() {
            setTimeout(() => {
              this.setState({});
            }, 100);
          }
        }
      `, Tsx: true},
		// `globalThis.setState({})` — receiver is Identifier `globalThis`, not
		// ThisKeyword, never matches regardless of mode.
		{Code: `
        class Hello extends React.Component {
          componentDidMount() {
            (globalThis as any).setState({});
          }
        }
      `, Tsx: true},
		// TS abstract method — body-absent, no setState possible. Locks that the
		// rule does not crash when traversing class bodies that contain
		// abstract / signature-only members.
		{Code: `
        abstract class Hello extends React.Component {
          abstract componentDidMount(): void;
        }
      `, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// ---- Upstream #1: createReactClass componentDidMount → setState ----
		{
			Code: `
        var Hello = createReactClass({
          componentDidMount: function() {
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
					Message:   "Do not use setState in componentDidMount",
					Line:      4, Column: 13,
				},
			},
		},

		// ---- Upstream #2: ES6 class method componentDidMount → setState ----
		{
			Code: `
        class Hello extends React.Component {
          componentDidMount() {
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

		// ---- Upstream #3: class field arrow `componentDidMount = () => { ... }` ----
		{
			Code: `
        class Hello extends React.Component {
          componentDidMount = () => {
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
          componentDidMount: function() {
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
          componentDidMount() {
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
          componentDidMount: function() {
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
          componentDidMount() {
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
          componentDidMount: function() {
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
          componentDidMount() {
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
          componentDidMount: function() {
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
          componentDidMount() {
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
          componentDidMount() {
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
          componentDidMount() {
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
          componentDidMount() {
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
          componentDidMount() {
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
          componentDidMount = function() {
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
          componentDidMount() {
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

		// ---- Edge: multiple setState calls in same componentDidMount ----
		{
			Code: `
        class Hello extends React.Component {
          componentDidMount() {
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

		// ---- Edge: async componentDidMount ----
		{
			Code: `
        class Hello extends React.Component {
          async componentDidMount() {
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

		// ---- Edge: generator componentDidMount ----
		{
			Code: `
        class Hello extends React.Component {
          *componentDidMount() {
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
          componentDidMount() {
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
          componentDidMount: () => {
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
          componentDidMount() {
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

		// ---- Edge: nested class whose inner method is also componentDidMount — inner takes precedence ----
		// Walk from the setState in Inner's method: Inner MethodDeclaration (depth 1,
		// stopper match) → report at depth 1, regardless of disallow-in-func mode.
		{
			Code: `
        class Outer extends React.Component {
          componentDidMount() {
            class Inner extends React.Component {
              componentDidMount() {
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
          componentDidMount() {
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

		// ---- Edge: getter named `componentDidMount` — stopper via IsMethodOrAccessor, depth 1 from the getter itself ----
		{
			Code: `
        class Hello extends React.Component {
          get componentDidMount() {
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

		// ---- Edge: PrivateIdentifier stopper `#componentDidMount` ----
		// ESLint's PrivateIdentifier.name is spec'd to exclude the leading #,
		// so `key.name === 'componentDidMount'` matches. Aligning with
		// upstream: we strip the # from tsgo's PrivateIdentifier.Text.
		{
			Code: `
        class Hello extends React.Component {
          #componentDidMount() {
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
          componentDidMount() {
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
          componentDidMount() {
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

		// ---- rslint hardening: tsgo shape boundaries ----
		// Multi-level parens around the receiver — SkipParentheses must unwrap recursively.
		{
			Code: `
        class Hello extends React.Component {
          componentDidMount() {
            ((this)).setState({});
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},
		// Parens wrapping the entire callee — SkipParentheses on `call.Expression` must
		// unwrap to the underlying PropertyAccessExpression. We report on the unwrapped
		// callee, so the column lands on `this` (after the stripped `(`), not on the
		// `(` itself. Locking col 14 here; the multi-paren case above (where the parens
		// wrap the receiver only, leaving the outer PropertyAccess intact) stays at 13.
		{
			Code: `
        class Hello extends React.Component {
          componentDidMount() {
            (this.setState)({});
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 14},
			},
		},
		// Spread arguments — call shape unaffected; rule only inspects callee.
		{
			Code: `
        class Hello extends React.Component {
          componentDidMount() {
            this.setState(...args);
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},

		// ---- rslint hardening: stopper / scope variants ----
		// Static method `static componentDidMount()` — upstream is purely textual on
		// `key.name`; the static modifier doesn't matter. tsgo MethodDeclaration is the
		// stopper; `static` is just a modifier flag. Locks that we're not gating on it.
		{
			Code: `
        class Hello extends React.Component {
          static componentDidMount() {
            this.setState({});
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},
		// Static class field arrow: `static componentDidMount = () => ...` — PropertyDeclaration
		// stopper with the same name; arrow body counts depth=1.
		{
			Code: `
        class Hello extends React.Component {
          static componentDidMount = () => {
            this.setState({});
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},
		// componentDidMount nested in unrelated object literal — locks that the rule is
		// purely textual: it doesn't care whether the enclosing thing is actually a
		// React component, it only matches the key name.
		{
			Code: `
        var x = {
          foo: {
            componentDidMount() {
              this.setState({});
            }
          }
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 5, Column: 15},
			},
		},
		// try/catch wrapping setState — locks that StatementList nodes don't interfere
		// with the ancestor walk.
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
		// switch/case wrapping setState — same as above for SwitchStatement / CaseClause.
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
		// Default-export anonymous class expression with componentDidMount.
		{
			Code: `
        export default class extends React.Component {
          componentDidMount() {
            this.setState({});
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},
		// Multiple classes in the same source — only the matching componentDidMount reports.
		{
			Code: `
        class A {
          foo() {
            this.setState({});
          }
        }
        class B extends React.Component {
          componentDidMount() {
            this.setState({});
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 9, Column: 13},
			},
		},
		// Empty options array — equivalent to default mode, must still report direct calls.
		{
			Code: `
        class Hello extends React.Component {
          componentDidMount() {
            this.setState({});
          }
        }
      `,
			Tsx:     true,
			Options: []interface{}{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},
		// settings.react is not a map (e.g. boolean) — rule must NOT crash and stays active
		// because version is not pinned. Locks the defensive `_, ok := …(map[…])` guard.
		{
			Code: `
        class Hello extends React.Component {
          componentDidMount() {
            this.setState({});
          }
        }
      `,
			Tsx:      true,
			Settings: map[string]interface{}{"react": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},
		// settings.react.version is a non-string (e.g. number) — rule must NOT crash and
		// stays active. Locks the defensive `_, ok := rs["version"].(string)` guard.
		{
			Code: `
        class Hello extends React.Component {
          componentDidMount() {
            this.setState({});
          }
        }
      `,
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"version": 16.3}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},
		// Optional-chain receiver wrapped in parens: `(this)?.setState({})` — combines
		// SkipParentheses on the receiver with the optional-chain flag.
		{
			Code: `
        class Hello extends React.Component {
          componentDidMount() {
            (this)?.setState({});
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},
		// Decorated class method (TS legacy decorator) — decorator is just a sibling node
		// on the MethodDeclaration; stopper match is unaffected.
		{
			Code: `
        function readonly(_t: any, _k: any, d: any) { return d; }
        class Hello extends React.Component {
          @readonly
          componentDidMount() {
            this.setState({});
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 6, Column: 13},
			},
		},
		// `await this.setState({})` inside async componentDidMount — AwaitExpression is
		// neither stopper nor function-like, so the ancestor walk passes through it
		// unchanged. (setState doesn't actually return a Promise, but rule semantics
		// are purely structural.)
		{
			Code: `
        class Hello extends React.Component {
          async componentDidMount() {
            await this.setState({});
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 19},
			},
		},
		// Class field arrow with EXPRESSION body (no block) — symmetric to the existing
		// block-bodied class-field arrow; locks that ArrowFunction with concise body
		// is still treated as function-like for depth counting.
		{
			Code: `
        class Hello extends React.Component {
          componentDidMount = () => this.setState({});
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 3, Column: 37},
			},
		},
		// Setter `set componentDidMount(v) {…}` — symmetric to the existing getter case.
		// SetAccessor satisfies IsMethodOrAccessor and matches name="componentDidMount".
		{
			Code: `
        class Hello extends React.Component {
          set componentDidMount(_v: any) {
            this.setState({});
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},
		// Comment between `this` and `.setState` — tsgo treats comments as trivia, the
		// PropertyAccess shape is preserved; rule still matches.
		{
			Code: `
        class Hello extends React.Component {
          componentDidMount() {
            this/* hi */.setState({});
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},
		// Newline between `this` and `.setState` — same as above; the AST shape is
		// independent of formatting.
		{
			Code: `
        class Hello extends React.Component {
          componentDidMount() {
            this
              .setState({});
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},
		// Inner getter NAMED componentDidMount inside an outer componentDidMount —
		// inner stopper match wins (depth=1 at the GetAccessor itself), regardless
		// of disallow-in-func mode.
		{
			Code: `
        class Outer extends React.Component {
          componentDidMount() {
            const Inner = class {
              get componentDidMount() {
                this.setState({});
                return null;
              }
            };
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 6, Column: 17},
			},
		},

		// ---- Empirically-verified user scenarios (locked from probe results) ----
		// Plain TS file (Tsx: false) — rule must work in non-JSX TS code, not just .tsx.
		{
			Code: `
class Hello {
  componentDidMount() {
    this.setState({});
  }
}
`,
			Tsx: false,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 5},
			},
		},
		// TS auto-accessor `accessor componentDidMount = () => { … }` — represented
		// as PropertyDeclaration with `accessor` modifier; stopper match works.
		{
			Code: `
        class Hello extends React.Component {
          accessor componentDidMount = () => {
            this.setState({});
          };
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 4, Column: 13},
			},
		},
		// Duplicate `componentDidMount` declarations in the same class — TS emits a
		// diagnostic but tsgo still parses both as siblings; rule fires on the
		// second one. Locks tolerance against malformed-but-parsable user code.
		{
			Code: `
        class Hello extends React.Component {
          componentDidMount() {}
          componentDidMount() {
            this.setState({});
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 5, Column: 13},
			},
		},
		// Class with mixin extends — `extends mixin()` is a CallExpression in the
		// extends clause; doesn't affect the stopper-based match for the method below.
		{
			Code: `
        function mixin(): any { return React.Component; }
        class Hello extends mixin() {
          componentDidMount() {
            this.setState({});
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 5, Column: 13},
			},
		},

		// ---- Real-world async / callback patterns + TS structural cases (verified) ----
		// TS overload signatures + body-bearing implementation — walk reaches the
		// impl's MethodDeclaration as the stopper; signature siblings are unrelated.
		{
			Code: `
        class Hello extends React.Component {
          componentDidMount(): void;
          componentDidMount(arg: number): void;
          componentDidMount(_arg?: number) {
            this.setState({});
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 6, Column: 13},
			},
		},
		// `namespace X { class Y { … } }` — ModuleDeclaration ancestor is non-stopper,
		// non-function-like; walk passes through.
		{
			Code: `
        namespace App {
          export class Hello extends React.Component {
            componentDidMount() {
              this.setState({});
            }
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 5, Column: 15},
			},
		},
		// new Promise(callback) — common async pattern. depth=2 default mode (valid),
		// invalid under disallow-in-func.
		{
			Code: `
        class Hello extends React.Component {
          componentDidMount() {
            new Promise((_resolve) => {
              this.setState({});
            });
          }
        }
      `,
			Tsx:     true,
			Options: []interface{}{"disallow-in-func"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 5, Column: 15},
			},
		},
		// setTimeout callback under disallow-in-func — paired with the valid
		// default-mode case above.
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
			Tsx:     true,
			Options: []interface{}{"disallow-in-func"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 5, Column: 15},
			},
		},
		// fetch().then().then(callback) — promise chain pattern with 2 nested arrows.
		// Walk: ArrowFunction (depth=1), ArrowFunction (depth=2 — but the outer one
		// returns a non-setState expression, so depth at MethodDeclaration is 3).
		// disallow-in-func reports at any depth.
		{
			Code: `
        class Hello extends React.Component {
          componentDidMount() {
            fetch("/api").then((res) => res.json()).then((data) => {
              this.setState({ data });
            });
          }
        }
      `,
			Tsx:     true,
			Options: []interface{}{"disallow-in-func"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSetState", Line: 5, Column: 15},
			},
		},
	})
}

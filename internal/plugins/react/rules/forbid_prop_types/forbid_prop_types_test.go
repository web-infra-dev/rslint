package forbid_prop_types

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestForbidPropTypesRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ForbidPropTypesRule, []rule_tester.ValidTestCase{
		// ---- Upstream valid cases (propTypes) ----
		{
			Code: `
        var First = createReactClass({
          render: function() {
            return <div />;
          }
        });
      `,
			Tsx: true,
		},
		// External (Identifier) propTypes — variable not declared in file →
		// nameToObject lookup misses → no recursion → no report.
		{
			Code: `
        var First = createReactClass({
          propTypes: externalPropTypes,
          render: function() {
            return <div />;
          }
        });
      `,
			Tsx: true,
		},
		// Default forbid does not include string/number/bool.
		{
			Code: `
        var First = createReactClass({
          propTypes: {
            s: PropTypes.string,
            n: PropTypes.number,
            i: PropTypes.instanceOf,
            b: PropTypes.bool
          },
          render: function() {
            return <div />;
          }
        });
      `,
			Tsx: true,
		},
		// Custom forbid omitting `array` → array stays valid.
		{
			Code: `
        var First = createReactClass({
          propTypes: {
            a: PropTypes.array
          },
          render: function() {
            return <div />;
          }
        })
      `,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"any", "object"}},
		},
		// Custom forbid omitting `object` → object stays valid.
		{
			Code: `
        var First = createReactClass({
          propTypes: {
            o: PropTypes.object
          },
          render: function() {
            return <div />;
          }
        });
      `,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"any", "array"}},
		},
		// Class-form `Foo.propTypes = {...}` with strings only.
		{
			Code: `
        class First extends React.Component {
          render() {
            return <div />;
          }
        }
        First.propTypes = {
          a: PropTypes.string,
          b: PropTypes.string
        };
        First.propTypes.justforcheck = PropTypes.string;
      `,
			Tsx: true,
		},
		// `instanceOf(...)` does not unwrap to a known forbidden name.
		{
			Code: `
        class First extends React.Component {
          render() {
            return <div />;
          }
        }
        First.propTypes = {
          elem: PropTypes.instanceOf(HTMLElement)
        };
      `,
			Tsx: true,
		},
		// String-literal key (not Identifier).
		{
			Code: `
        class Hello extends React.Component {
          render() {
            return <div>Hello</div>;
          }
        }
        Hello.propTypes = {
          "aria-controls": PropTypes.string
        };
      `,
			Tsx: true,
		},
		// Spread inside object literal — silently skipped.
		{
			Code: `
        var Hello = createReactClass({
          render: function() {
            let { a, ...b } = obj;
            let c = { ...d };
            return <div />;
          }
        });
      `,
			Tsx: true,
		},
		// `instanceOf(Map).isRequired` → unwrap `.isRequired` then unwrap
		// CallExpression: callee is `PropTypes.instanceOf`, arguments are
		// `Map` (Identifier), name is "Map" — not forbidden.
		{
			Code: `
        var Hello = createReactClass({
          propTypes: {
            retailer: PropTypes.instanceOf(Map).isRequired,
            requestRetailer: PropTypes.func.isRequired
          },
          render: function() {
            return <div />;
          }
        });
      `,
			Tsx: true,
		},
		// Spread in class field's object literal.
		{
			Code: `
        class Test extends React.component {
          static propTypes = {
            intl: React.propTypes.number,
            ...propTypes
          };
        }
      `,
			Tsx: true,
		},
		// Static getter returning an object literal with spread.
		{
			Code: `
        class Test extends React.component {
          static get propTypes() {
            return {
              intl: React.propTypes.number,
              ...propTypes
            };
          };
        }
      `,
			Tsx: true,
		},

		// ---- contextTypes / childContextTypes — valid (option off) ----
		// Without `checkContextTypes`, contextTypes / childContextTypes are
		// invisible to the rule.
		{
			Code: `
        var First = createReactClass({
          childContextTypes: externalPropTypes,
          render: function() {
            return <div />;
          }
        });
      `,
			Tsx:     true,
			Options: map[string]interface{}{"checkContextTypes": true},
		},
		{
			Code: `
        var First = createReactClass({
          childContextTypes: {
            s: PropTypes.string,
            n: PropTypes.number,
            i: PropTypes.instanceOf,
            b: PropTypes.bool
          },
          render: function() {
            return <div />;
          }
        });
      `,
			Tsx:     true,
			Options: map[string]interface{}{"checkContextTypes": true},
		},
		{
			Code: `
        var First = createReactClass({
          childContextTypes: {
            a: PropTypes.array
          },
          render: function() {
            return <div />;
          }
        });
      `,
			Tsx: true,
			Options: map[string]interface{}{
				"forbid":            []interface{}{"any", "object"},
				"checkContextTypes": true,
			},
		},
		{
			Code: `
        var First = createReactClass({
          childContextTypes: {
            o: PropTypes.object
          },
          render: function() {
            return <div />;
          }
        });
      `,
			Tsx: true,
			Options: map[string]interface{}{
				"forbid":            []interface{}{"any", "array"},
				"checkContextTypes": true,
			},
		},
		{
			Code: `
        class First extends React.Component {
          render() {
            return <div />;
          }
        }
        First.childContextTypes = {
          a: PropTypes.string,
          b: PropTypes.string
        };
        First.childContextTypes.justforcheck = PropTypes.string;
      `,
			Tsx:     true,
			Options: map[string]interface{}{"checkContextTypes": true},
		},
		{
			Code: `
        class First extends React.Component {
          render() {
            return <div />;
          }
        }
        First.childContextTypes = {
          elem: PropTypes.instanceOf(HTMLElement)
        };
      `,
			Tsx:     true,
			Options: map[string]interface{}{"checkContextTypes": true},
		},
		{
			Code: `
        class Hello extends React.Component {
          render() {
            return <div>Hello</div>;
          }
        }
        Hello.childContextTypes = {
          "aria-controls": PropTypes.string
        };
      `,
			Tsx:     true,
			Options: map[string]interface{}{"checkContextTypes": true},
		},
		// `static childContextTypes = {...}` with checkContextTypes:true →
		// childContextTypes is gated by checkChildContextTypes, not
		// checkContextTypes; the field is invisible to the rule. Reports
		// from the `React.childContextTypes.number` member access do NOT
		// occur because it's value=PropTypes-package member access whose
		// `.property.name` is `number`, not in default forbid.
		{
			Code: `
        class Test extends React.component {
          static childContextTypes = {
            intl: React.childContextTypes.number,
            ...childContextTypes
          };
        }
      `,
			Tsx:     true,
			Options: map[string]interface{}{"checkContextTypes": true},
		},
		{
			Code: `
        class Test extends React.component {
          static get childContextTypes() {
            return {
              intl: React.childContextTypes.number,
              ...childContextTypes
            };
          };
        }
      `,
			Tsx:     true,
			Options: map[string]interface{}{"checkContextTypes": true},
		},
		// Same group with `checkChildContextTypes: true` — but values are
		// strings/numbers/bools/instanceOf, none in default forbid.
		{
			Code: `
        var First = createReactClass({
          childContextTypes: externalPropTypes,
          render: function() {
            return <div />;
          }
        });
      `,
			Tsx:     true,
			Options: map[string]interface{}{"checkChildContextTypes": true},
		},
		{
			Code: `
        var First = createReactClass({
          childContextTypes: {
            s: PropTypes.string,
            n: PropTypes.number,
            i: PropTypes.instanceOf,
            b: PropTypes.bool
          },
          render: function() {
            return <div />;
          }
        });
      `,
			Tsx:     true,
			Options: map[string]interface{}{"checkChildContextTypes": true},
		},
		{
			Code: `
        var First = createReactClass({
          childContextTypes: {
            a: PropTypes.array
          },
          render: function() {
            return <div />;
          }
        });
      `,
			Tsx: true,
			Options: map[string]interface{}{
				"forbid":                 []interface{}{"any", "object"},
				"checkChildContextTypes": true,
			},
		},
		{
			Code: `
        var First = createReactClass({
          childContextTypes: {
            o: PropTypes.object
          },
          render: function() {
            return <div />;
          }
        });
      `,
			Tsx: true,
			Options: map[string]interface{}{
				"forbid":                 []interface{}{"any", "array"},
				"checkChildContextTypes": true,
			},
		},
		{
			Code: `
        class First extends React.Component {
          render() {
            return <div />;
          }
        }
        First.childContextTypes = {
          a: PropTypes.string,
          b: PropTypes.string
        };
        First.childContextTypes.justforcheck = PropTypes.string;
      `,
			Tsx:     true,
			Options: map[string]interface{}{"checkChildContextTypes": true},
		},
		{
			Code: `
        class First extends React.Component {
          render() {
            return <div />;
          }
        }
        First.childContextTypes = {
          elem: PropTypes.instanceOf(HTMLElement)
        };
      `,
			Tsx:     true,
			Options: map[string]interface{}{"checkChildContextTypes": true},
		},
		{
			Code: `
        class Hello extends React.Component {
          render() {
            return <div>Hello</div>;
          }
        }
        Hello.childContextTypes = {
          "aria-controls": PropTypes.string
        };
      `,
			Tsx:     true,
			Options: map[string]interface{}{"checkChildContextTypes": true},
		},
		{
			Code: `
        class Test extends React.component {
          static childContextTypes = {
            intl: React.childContextTypes.number,
            ...childContextTypes
          };
        }
      `,
			Tsx:     true,
			Options: map[string]interface{}{"checkChildContextTypes": true},
		},
		{
			Code: `
        class Test extends React.component {
          static get childContextTypes() {
            return {
              intl: React.childContextTypes.number,
              ...childContextTypes
            };
          };
        }
      `,
			Tsx:     true,
			Options: map[string]interface{}{"checkChildContextTypes": true},
		},
		// IIFE returning an object — class field initializer is a CallExpression
		// not a propWrapper, so checkNode falls through and nothing is reported.
		{
			Code: `
        class TestComponent extends React.Component {
          static defaultProps = function () {
            const date = new Date();
            return {
              date
            };
          }();
        }
      `,
			Tsx: true,
		},
		// `Object.assign(...)` is not in propWrapperFunctions by default →
		// checkNode CallExpression branch matches no propWrapper → no recursion.
		{
			Code: `
        class HeroTeaserList extends React.Component {
          render() { return null; }
        }
        HeroTeaserList.propTypes = Object.assign({
          heroIndex: PropTypes.number,
          preview: PropTypes.bool,
        }, componentApi, teaserListProps);
      `,
			Tsx: true,
		},
		// `PropTypes.shape(Foo)` where Foo is an Identifier → CallExpression
		// listener checks `arg0` is ObjectLiteral, but Foo is an Identifier →
		// no recursion. ObjectExpression listener fires on the surrounding
		// `{foo: PropTypes.string}` and `{bar: PropTypes.shape(Foo)}`, but
		// the keys are not propTypes / contextTypes / childContextTypes.
		{
			Code: `
        import PropTypes from "prop-types";
        const Foo = {
          foo: PropTypes.string,
        };
        const Bar = {
          bar: PropTypes.shape(Foo),
        };
      `,
			Tsx: true,
		},
		// `Yup.object().shape({...})` — outer .shape callee receives from a
		// CallExpression (not a propTypes-package Identifier), so the
		// upstream gate returns early; no recursion.
		{
			Code: `
        import yup from "yup"
        const formValidation = Yup.object().shape({
          name: Yup.string(),
          customer_ids: Yup.array()
        });
      `,
			Tsx: true,
		},
		// Same shape with explicit forbid — Yup-prefixed calls are still
		// gated out at the .shape() level.
		{
			Code: `
        import yup from "Yup"
        const validation = yup.object().shape({
          address: yup.object({
            city: yup.string(),
            zip: yup.string(),
          })
        })
      `,
			Tsx: true,
			Options: map[string]interface{}{
				"forbid": []interface{}{"string", "object"},
			},
		},
		// `Yup.array(Yup.object().shape({...}))` — outer .array call's
		// callee `Yup.array` is a member expression with bare Identifier
		// receiver Yup, but `Yup` IS treated as propTypes-package by default
		// (no foreign import yet). The shape match only fires on .shape(),
		// which is gated out (object is CallExpression).
		{
			Code: `
        import yup from "yup"
        Yup.array(
          Yup.object().shape({
            value: Yup.number()
          })
        )
      `,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"number"}},
		},
		// Custom prop-types alias: `import CustomPropTypes from "prop-types"`
		// → propTypesPackageName = 'CustomPropTypes'. Member access
		// `CustomPropTypes.shape(...)` is recognized as propTypes-package.
		// Inner `CustomPropTypes.String` and `CustomPropTypes.number` aren't
		// in default forbid.
		{
			Code: `
        import CustomPropTypes from "prop-types";
        class Component extends React.Component {};
        Component.propTypes = {
          a: CustomPropTypes.shape({
            b: CustomPropTypes.String,
            c: CustomPropTypes.number.isRequired,
          })
        }
      `,
			Tsx: true,
		},
		// `import CustomReact from "react"` → reactPackageName = 'CustomReact'.
		// `CustomReact.PropTypes.string` → MemberExpression with object
		// `CustomReact.PropTypes` (nested ME) → object.name undefined →
		// !isForeign branch passes → recognized as propTypes-package.
		{
			Code: `
        import CustomReact from "react"
        class Component extends React.Component {};
        Component.propTypes = {
          b: CustomReact.PropTypes.string,
        }
      `,
			Tsx: true,
		},
		// Foreign package importing PropTypes — `import PropTypes from "yup"`
		// (default import). Default specifier's local name is 'PropTypes' →
		// triggers the foreign-package branch. Then `PropTypes.array()` is
		// NOT recognized as propTypes-package (since PropTypes name doesn't
		// match propTypesPackageName=='', and isForeign=true).
		{
			Code: `
        import PropTypes from "yup"
        class Component extends React.Component {};
        Component.propTypes = {
          b: PropTypes.array(),
        }
      `,
			Tsx: true,
		},
		// Foreign package importing PropTypes via named import.
		{
			Code: `
        import { PropTypes, shape, any } from "yup"
        class Component extends React.Component {};
        Component.propTypes = {
          b: PropTypes.any,
        }
      `,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"any"}},
		},
		// Foreign package importing PropTypes from "not-react" — array() is
		// gated out.
		{
			Code: `
        import { PropTypes } from "not-react"
        class Component extends React.Component {};
        Component.propTypes = {
          b: PropTypes.array(),
        }
      `,
			Tsx: true,
		},

		// ---- Edge cases (Dimension 4 universal shapes) ----
		// Empty forbid → nothing forbidden, even default targets.
		{
			Code: `
        var First = createReactClass({
          propTypes: {
            a: PropTypes.any,
            b: PropTypes.array,
            c: PropTypes.object
          },
          render: function() { return <div />; }
        });
      `,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{}},
		},
		// Numeric / string-literal property keys.
		{
			Code: `
        class C extends React.Component {}
        C.propTypes = {
          0: PropTypes.string,
          "1": PropTypes.string,
        };
      `,
			Tsx: true,
		},
		// Computed property keys with static literals — not handled by
		// upstream's `getPropertyName(property.key)`; we mirror by accepting
		// only Identifier / StringLiteral / NumericLiteral keys via
		// `GetStaticPropertyName`. Computed `[k]: PropTypes.any` is NOT
		// reported when the computed expression is not a static literal.
		{
			Code: `
        const k = "any";
        var First = createReactClass({
          propTypes: {
            [k]: PropTypes.string,  // dynamic key
          },
          render: function() { return <div />; }
        });
      `,
			Tsx: true,
		},
		// Nested ObjectLiteral inside a non-propTypes property — the inner
		// `propTypes` key looks like a propTypes declaration syntactically,
		// so upstream's ObjectExpression listener WILL match it. Mirror that
		// permissive behavior with a Skip-marked test below if behavior diverges.
		// Note: upstream's listener fires on EVERY ObjectExpression so this
		// particular nesting is treated as a valid propTypes declaration.
		// String values aren't forbidden by default → valid.
		{
			Code: `
        const config = {
          propTypes: {
            a: PropTypes.string,
          },
        };
      `,
			Tsx: true,
		},

		// ---- Options shape coverage (CLI vs. multi-element) ----
		// Bare object form (single-option CLI shape).
		{
			Code:    `class C extends React.Component {} C.propTypes = { a: PropTypes.string };`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"any"}},
		},
		// Array-wrapped form (rule-tester / multi-element CLI shape).
		{
			Code:    `class C extends React.Component {} C.propTypes = { a: PropTypes.string };`,
			Tsx:     true,
			Options: []interface{}{map[string]interface{}{"forbid": []interface{}{"any"}}},
		},
		// `null` options → defaults apply, but `string` not in defaults.
		{
			Code:    `class C extends React.Component {} C.propTypes = { a: PropTypes.string };`,
			Tsx:     true,
			Options: nil,
		},

		// ---- Robustness: TS expression wrappers on the LHS / RHS ----
		// `Foo!.propTypes = {a: PropTypes.string}` — non-null assertion on
		// the receiver. Outer PropertyAccessExpression name is still
		// 'propTypes'. Right side is valid (string).
		{
			Code: `
        declare const Foo: any;
        Foo!.propTypes = { a: PropTypes.string };
      `,
			Tsx: true,
		},
		// `(Foo as any).propTypes = {a: PropTypes.string}`.
		{
			Code: `
        declare const Foo: any;
        (Foo as any).propTypes = { a: PropTypes.string };
      `,
			Tsx: true,
		},
		// `(Foo.propTypes) = {a: PropTypes.string}` — paren wrapping LHS.
		{
			Code: `
        class Foo extends React.Component {}
        (Foo.propTypes) = { a: PropTypes.string };
      `,
			Tsx: true,
		},
		// Value-side wrapper: `(PropTypes.string as any)`. SkipExpressionWrappers
		// strips wrappers before the .name read.
		{
			Code: `
        class C extends React.Component {}
        C.propTypes = { a: (PropTypes.string as any) };
      `,
			Tsx: true,
		},
		// `(PropTypes.string)` — pure parens.
		{
			Code: `
        class C extends React.Component {}
        C.propTypes = { a: (PropTypes.string) };
      `,
			Tsx: true,
		},
		// Optional chain: `PropTypes?.string`. Same Identifier/PAE shape, just
		// IsOptional flagged. Not in default forbid → valid.
		{
			Code: `
        class C extends React.Component {}
        C.propTypes = { a: PropTypes?.string };
      `,
			Tsx: true,
		},
		// `PropTypes.string!` — non-null assertion on the value side.
		{
			Code: `
        class C extends React.Component {}
        C.propTypes = { a: PropTypes.string! };
      `,
			Tsx: true,
		},

		// ---- Robustness: declaration / nesting forms ----
		// Class expression: `const X = class extends React.Component {...}`.
		// PropertyDeclaration listener still fires on the static field.
		{
			Code: `
        const X = class extends React.Component {
          static propTypes = { a: PropTypes.string };
        };
      `,
			Tsx: true,
		},
		// Anonymous class expression assigned to const.
		{
			Code: `
        const Foo = class extends React.Component {};
        Foo.propTypes = { a: PropTypes.string };
      `,
			Tsx: true,
		},
		// Nested classes — outer's static propTypes valid, inner's also valid.
		{
			Code: `
        class Outer extends React.Component {
          static propTypes = { a: PropTypes.string };
          render() {
            class Inner extends React.Component {
              static propTypes = { b: PropTypes.string };
            }
            return null;
          }
        }
      `,
			Tsx: true,
		},
		// Static getter inside class expression.
		{
			Code: `
        const X = class extends React.Component {
          static get propTypes() {
            return { a: PropTypes.string };
          }
        };
      `,
			Tsx: true,
		},
		// Method body returns from inside conditional — last return wins.
		{
			Code: `
        class C extends React.Component {
          static get propTypes() {
            if (foo) {
              return { a: PropTypes.string };
            }
            return { a: PropTypes.bool };
          };
        }
      `,
			Tsx: true,
		},

		// ---- Robustness: assignment operator variations ----
		// `Foo.propTypes ||= {...}` — compound assign. RHS is valid (string).
		{
			Code: `
        class Foo extends React.Component {}
        Foo.propTypes ||= { a: PropTypes.string };
      `,
			Tsx: true,
		},
		// `Foo.propTypes ??= {...}`.
		{
			Code: `
        class Foo extends React.Component {}
        Foo.propTypes ??= { a: PropTypes.string };
      `,
			Tsx: true,
		},

		// ---- Robustness: bracket vs dot access ----
		// `Foo['propTypes'] = {...}` — upstream's `getPropertyName` returns
		// '' for computed string-literal access on MemberExpression. Mirror
		// upstream: NOT detected, so even forbidden values are valid.
		{
			Code: `
        class Foo extends React.Component {}
        Foo['propTypes'] = { a: PropTypes.any };
      `,
			Tsx: true,
		},

		// ---- Robustness: scope / Identifier resolution edge cases ----
		// Identifier value whose NAME is not in forbid — name='helper' is
		// not in default ['any','array','object']. Source of the binding
		// doesn't matter (matches upstream's bare-Identifier handling
		// which keys solely on the identifier text).
		{
			Code: `
        const helper = {};
        var Hello = createReactClass({
          propTypes: {
            retailer: helper,
          },
          render: function() { return <div />; }
        });
      `,
			Tsx: true,
		},
		// `Component.propTypes = somePropTypes` where somePropTypes is NOT
		// declared in the file — scope walk fails → no recursion → no report.
		{
			Code: `
        class C extends React.Component {}
        C.propTypes = somePropTypes;  // unresolved
      `,
			Tsx: true,
		},
		// Scope semantics — child-scope binding is NOT visible from
		// outer scope. Mirrors ESLint scope manager: `scope.upper` chain
		// walks UP, never DOWN into child scopes. The outer `C.propTypes
		// = propTypes` reference resolves at top-level scope where no
		// `propTypes` is bound (the function-local one is invisible).
		{
			Code: `
        function build() {
          const propTypes = { a: PropTypes.any };
          return propTypes;
        }
        class C extends React.Component {}
        C.propTypes = propTypes;  // resolves to nothing → no recursion
      `,
			Tsx: true,
		},
		// Scope shadowing — closest binding wins. Outer `propTypes` is a
		// safe object; inner `propTypes` (shadowed) has any but is bound
		// in a different scope from the assignment site.
		{
			Code: `
        const propTypes = { a: PropTypes.string };
        function f() {
          const propTypes = { a: PropTypes.any };  // shadowed, not visible from outer assignment
        }
        class C extends React.Component {}
        C.propTypes = propTypes;  // resolves to outer (safe) binding
      `,
			Tsx: true,
		},
		// Re-assignment after declaration — upstream's
		// `variableUtil.findVariableByName` returns the variable's
		// `defs[0].node.init` (declaration's initializer), NOT the last
		// assigned value. So a reassignment to a forbidden shape after
		// the declaration is INVISIBLE to the rule. Lock that in.
		{
			Code: `
        let propTypes = { a: PropTypes.string };
        propTypes = { a: PropTypes.any };  // invisible to lookup
        class C extends React.Component {}
        C.propTypes = propTypes;
      `,
			Tsx: true,
		},
		// Switch-statement return: NO forbidden value in last case.
		{
			Code: `
        class C extends React.Component {
          static get propTypes() {
            switch (mode) {
              case "a":
                return { x: PropTypes.string };
              default:
                return { x: PropTypes.string };
            }
          };
        }
      `,
			Tsx: true,
		},
		// `Component.propTypes = somePropTypes` where somePropTypes IS
		// declared but to a non-object literal → no recursion.
		{
			Code: `
        const somePropTypes = makeProps();  // CallExpression init
        class C extends React.Component {}
        C.propTypes = somePropTypes;
      `,
			Tsx: true,
		},

		// ---- Robustness: deeply nested object literals (real user shapes) ----
		// Deeply nested shape with no forbidden types.
		{
			Code: `
        class C extends React.Component {}
        C.propTypes = {
          user: PropTypes.shape({
            profile: PropTypes.shape({
              name: PropTypes.string,
              age: PropTypes.number,
            }),
          }),
        };
      `,
			Tsx: true,
		},
		// `arrayOf(shape({...}))` — outer arrayOf is unwrapped via
		// CallExpression.callee, inner shape recurses via CallExpression
		// listener.
		{
			Code: `
        class C extends React.Component {}
        C.propTypes = {
          users: PropTypes.arrayOf(PropTypes.shape({
            name: PropTypes.string,
          })),
        };
      `,
			Tsx: true,
		},
		// `oneOfType([...])` — array literal argument, not ObjectLiteral, so
		// CallExpression listener doesn't recurse.
		{
			Code: `
        class C extends React.Component {}
        C.propTypes = {
          x: PropTypes.oneOfType([PropTypes.string, PropTypes.number]),
        };
      `,
			Tsx: true,
		},

		// ---- Robustness: real-world component patterns ----
		// `module.exports = ...` with propTypes assignment.
		{
			Code: `
        function Foo(props) { return null; }
        Foo.propTypes = {
          name: PropTypes.string,
        };
        module.exports = Foo;
      `,
			Tsx: true,
		},
		// HOC return value pattern — `withTheme(Foo)` then assign propTypes
		// to the original Foo, not the wrapped one.
		{
			Code: `
        function Foo(props) { return null; }
        Foo.propTypes = {
          name: PropTypes.string,
        };
        const Themed = withTheme(Foo);
      `,
			Tsx: true,
		},
		// `forwardRef`-wrapped component with propTypes assignment after.
		{
			Code: `
        const Foo = React.forwardRef((props, ref) => null);
        Foo.propTypes = {
          name: PropTypes.string,
        };
      `,
			Tsx: true,
		},

		// ---- Robustness: import variations ----
		// Namespace import from prop-types: `import * as PT from "prop-types"`.
		// First specifier's local is 'PT' → propTypesPackageName='PT'.
		// `PT.string` recognized via reactPackageName check (object='PT' →
		// matches !isForeign branch).
		{
			Code: `
        import * as PT from "prop-types";
        class C extends React.Component {}
        C.propTypes = { a: PT.string };
      `,
			Tsx: true,
		},
		// Default + named imports from prop-types.
		{
			Code: `
        import PropTypes, { shape } from "prop-types";
        class C extends React.Component {}
        C.propTypes = { a: shape({ b: PropTypes.string }) };
      `,
			Tsx: true,
		},
		// Aliased named import: `import { PropTypes as PT } from "react"`.
		{
			Code: `
        import { PropTypes as PT } from "react";
        class C extends React.Component {}
        C.propTypes = { a: PT.string };
      `,
			Tsx: true,
		},
		// Side-effect-only import (no specifiers).
		{
			Code: `
        import "prop-types";
        class C extends React.Component {}
        C.propTypes = { a: PropTypes.string };
      `,
			Tsx: true,
		},
		// Multiple imports — propTypesPackageName updated each time the
		// 'prop-types' branch is hit. Only last import's specifiers[0] wins.
		{
			Code: `
        import a from "prop-types";
        import b from "react";
        class C extends React.Component {}
        C.propTypes = { x: PropTypes.string };
      `,
			Tsx: true,
		},

		// ---- Robustness: createReactClass argument shape variations ----
		// `createReactClass({propTypes: forbidExtraProps({...})})` — upstream
		// ObjectExpression listener doesn't recurse into CallExpression
		// values. Match upstream literally: NOT detected, so a forbidden
		// type inside is missed. Lock that behavior in.
		{
			Code: `
        var First = createReactClass({
          propTypes: forbidExtraProps({
            a: PropTypes.any
          }),
          render: function() { return <div />; }
        });
      `,
			Tsx: true,
			Settings: map[string]interface{}{
				"propWrapperFunctions": []interface{}{"forbidExtraProps"},
			},
		},
		// `createReactClass({propTypes: externalThing})` — Identifier value,
		// not ObjectExpression → not recursed.
		{
			Code: `
        var First = createReactClass({
          propTypes: externalPropTypes,
          render: function() { return <div />; }
        });
      `,
			Tsx: true,
		},

		// ---- Robustness: forbid list edge values ----
		// Non-string entries in forbid array are silently dropped.
		{
			Code: `
        class C extends React.Component {}
        C.propTypes = { a: PropTypes.string };
      `,
			Tsx: true,
			Options: map[string]interface{}{
				"forbid": []interface{}{42, nil, []interface{}{"any"}},
			},
		},
		// Empty string in forbid list — never matches a real type name.
		{
			Code: `
        class C extends React.Component {}
        C.propTypes = { a: PropTypes.string };
      `,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{""}},
		},
		// Forbid list with duplicates — slices.Contains tolerates them.
		{
			Code: `
        class C extends React.Component {}
        C.propTypes = { a: PropTypes.string };
      `,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"any", "any", "object"}},
		},

		// ---- Robustness: shape({...}) edge shapes ----
		// `PropTypes.shape()` with no arguments — checkProperties never
		// called (Arguments empty).
		{
			Code: `
        class C extends React.Component {}
        C.propTypes = { a: PropTypes.shape() };
      `,
			Tsx: true,
		},
		// `PropTypes.shape(nonObjectArg)` — argument is not ObjectLiteral,
		// CallExpression listener bails on the arg type check.
		{
			Code: `
        class C extends React.Component {}
        C.propTypes = { a: PropTypes.shape(somethingElse) };
      `,
			Tsx: true,
		},

		// ---- Babel-eslint legacy cases (skipped: TS-class-prop-without-init
		// shape that valid TS parsers reject) ----
		{
			// Upstream: `class Component { propTypes: { a: PropTypes.any } }`
			// — class field WITHOUT initializer, with a type annotation
			// containing object. TS strict parser doesn't accept this as a
			// class field with object value. Upstream gates this whole
			// case behind `semver.satisfies(babelEslintVersion, '< 9')`.
			Code: `
        class Component {}
      `,
			Tsx: true,
			Skip: true, // SKIP: tsgo parser doesn't parse this babel-eslint-only shape
		},
	}, []rule_tester.InvalidTestCase{
		// ---- Upstream invalid cases (propTypes) ----
		{
			Code: `
        var First = createReactClass({
          propTypes: {
            a: PropTypes.any
          },
          render: function() {
            return <div />;
          }
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "any" is forbidden`, Line: 4, Column: 13},
			},
		},
		// Custom forbid: `number`.
		{
			Code: `
        var First = createReactClass({
          propTypes: {
            n: PropTypes.number
          },
          render: function() {
            return <div />;
          }
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "number" is forbidden`, Line: 4, Column: 13},
			},
			Options: map[string]interface{}{"forbid": []interface{}{"number"}},
		},
		// `.isRequired` chain unwraps.
		{
			Code: `
        var First = createReactClass({
          propTypes: {
            a: PropTypes.any.isRequired
          },
          render: function() {
            return <div />;
          }
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "any" is forbidden`, Line: 4, Column: 13},
			},
		},
		{
			Code: `
        var First = createReactClass({
          propTypes: {
            a: PropTypes.array
          },
          render: function() {
            return <div />;
          }
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "array" is forbidden`, Line: 4, Column: 13},
			},
		},
		{
			Code: `
        var First = createReactClass({
          propTypes: {
            a: PropTypes.array.isRequired
          },
          render: function() {
            return <div />;
          }
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "array" is forbidden`, Line: 4, Column: 13},
			},
		},
		{
			Code: `
        var First = createReactClass({
          propTypes: {
            a: PropTypes.object
          },
          render: function() {
            return <div />;
          }
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "object" is forbidden`, Line: 4, Column: 13},
			},
		},
		{
			Code: `
        var First = createReactClass({
          propTypes: {
            a: PropTypes.object.isRequired
          },
          render: function() {
            return <div />;
          }
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "object" is forbidden`, Line: 4, Column: 13},
			},
		},
		// Two forbidden in one object → 2 errors.
		{
			Code: `
        var First = createReactClass({
          propTypes: {
            a: PropTypes.array,
            o: PropTypes.object
          },
          render: function() {
            return <div />;
          }
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "array" is forbidden`, Line: 4, Column: 13},
				{MessageId: "forbiddenPropType", Message: `Prop type "object" is forbidden`, Line: 5, Column: 13},
			},
		},
		// Two separate components, one forbidden each.
		{
			Code: `
        var First = createReactClass({
          propTypes: {
            a: PropTypes.array
          },
          render: function() {
            return <div />;
          }
        });
        var Second = createReactClass({
          propTypes: {
            o: PropTypes.object
          },
          render: function() {
            return <div />;
          }
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "array" is forbidden`, Line: 4, Column: 13},
				{MessageId: "forbiddenPropType", Message: `Prop type "object" is forbidden`, Line: 12, Column: 13},
			},
		},
		// Two classes, two propTypes each → 4 errors.
		{
			Code: `
        class First extends React.Component {
          render() {
            return <div />;
          }
        }
        First.propTypes = {
            a: PropTypes.array,
            o: PropTypes.object
        };
        class Second extends React.Component {
          render() {
            return <div />;
          }
        }
        Second.propTypes = {
            a: PropTypes.array,
            o: PropTypes.object
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "array" is forbidden`},
				{MessageId: "forbiddenPropType", Message: `Prop type "object" is forbidden`},
				{MessageId: "forbiddenPropType", Message: `Prop type "array" is forbidden`},
				{MessageId: "forbiddenPropType", Message: `Prop type "object" is forbidden`},
			},
		},
		// `forbidExtraProps({...})` propWrapper unwrap → reports.
		{
			Code: `
        class First extends React.Component {
          render() {
            return <div />;
          }
        }
        First.propTypes = forbidExtraProps({
            a: PropTypes.array
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "array" is forbidden`},
			},
			Settings: map[string]interface{}{
				"propWrapperFunctions": []interface{}{"forbidExtraProps"},
			},
		},
		// Identifier propTypes → look up file-level binding → checkProperties.
		{
			Code: `
        import { forbidExtraProps } from "airbnb-prop-types";
        export const propTypes = {dpm: PropTypes.any};
        export default function Component() {}
        Component.propTypes = propTypes;
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "any" is forbidden`},
			},
		},
		// Identifier propTypes wrapped via propWrapper.
		{
			Code: `
        import { forbidExtraProps } from "airbnb-prop-types";
        export const propTypes = {a: PropTypes.any};
        export default function Component() {}
        Component.propTypes = forbidExtraProps(propTypes);
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "any" is forbidden`},
			},
			Settings: map[string]interface{}{
				"propWrapperFunctions": []interface{}{"forbidExtraProps"},
			},
		},
		// Class field `static propTypes = {...}`.
		{
			Code: `
        class Component extends React.Component {
          static propTypes = {
            a: PropTypes.array,
            o: PropTypes.object
          };
          render() {
            return <div />;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "array" is forbidden`},
				{MessageId: "forbiddenPropType", Message: `Prop type "object" is forbidden`},
			},
		},
		// `static get propTypes()` → return statement is the object literal.
		{
			Code: `
        class Component extends React.Component {
          static get propTypes() {
            return {
              a: PropTypes.array,
              o: PropTypes.object
            };
          };
          render() {
            return <div />;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "array" is forbidden`},
				{MessageId: "forbiddenPropType", Message: `Prop type "object" is forbidden`},
			},
		},
		// Class field with propWrapper.
		{
			Code: `
        class Component extends React.Component {
          static propTypes = forbidExtraProps({
            a: PropTypes.array,
            o: PropTypes.object
          });
          render() {
            return <div />;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "array" is forbidden`},
				{MessageId: "forbiddenPropType", Message: `Prop type "object" is forbidden`},
			},
			Settings: map[string]interface{}{
				"propWrapperFunctions": []interface{}{"forbidExtraProps"},
			},
		},
		// `static get propTypes() { return forbidExtraProps({...}) }`.
		{
			Code: `
        class Component extends React.Component {
          static get propTypes() {
            return forbidExtraProps({
              a: PropTypes.array,
              o: PropTypes.object
            });
          }
          render() {
            return <div />;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "array" is forbidden`},
				{MessageId: "forbiddenPropType", Message: `Prop type "object" is forbidden`},
			},
			Settings: map[string]interface{}{
				"propWrapperFunctions": []interface{}{"forbidExtraProps"},
			},
		},
		// Custom forbid: `instanceOf` — argument-walk catches it.
		{
			Code: `
        var Hello = createReactClass({
          propTypes: {
            retailer: PropTypes.instanceOf(Map).isRequired,
            requestRetailer: PropTypes.func.isRequired
          },
          render: function() {
            return <div />;
          }
        });
      `,
			Tsx: true,
			Options: map[string]interface{}{"forbid": []interface{}{"instanceOf"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "instanceOf" is forbidden`},
			},
		},
		// Identifier as value (e.g. `retailer: object`) — value's name is the
		// target.
		{
			Code: `
        var object = PropTypes.object;
        var Hello = createReactClass({
          propTypes: {
            retailer: object,
          },
          render: function() {
            return <div />;
          }
        });
      `,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"object"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "object" is forbidden`},
			},
		},

		// ---- contextTypes — invalid (option ON) ----
		{
			Code: `
        var First = createReactClass({
          contextTypes: {
            a: PropTypes.any
          },
          render: function() {
            return <div />;
          }
        });
      `,
			Tsx:     true,
			Options: map[string]interface{}{"checkContextTypes": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "any" is forbidden`, Line: 4, Column: 13},
			},
		},
		{
			Code: `
        class Foo extends Component {
          static contextTypes = {
            a: PropTypes.any
          }
          render() {
            return <div />;
          }
        }
      `,
			Tsx:     true,
			Options: map[string]interface{}{"checkContextTypes": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "any" is forbidden`, Line: 4, Column: 13},
			},
		},
		{
			Code: `
        class Foo extends Component {
          static get contextTypes() {
            return {
              a: PropTypes.any
            };
          }
          render() {
            return <div />;
          }
        }
      `,
			Tsx:     true,
			Options: map[string]interface{}{"checkContextTypes": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "any" is forbidden`, Line: 5, Column: 15},
			},
		},
		{
			Code: `
        class Foo extends Component {
          render() {
            return <div />;
          }
        }
        Foo.contextTypes = {
          a: PropTypes.any
        };
      `,
			Tsx:     true,
			Options: map[string]interface{}{"checkContextTypes": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "any" is forbidden`, Line: 8, Column: 11},
			},
		},
		{
			Code: `
        function Foo(props) {
          return <div />;
        }
        Foo.contextTypes = {
          a: PropTypes.any
        };
      `,
			Tsx:     true,
			Options: map[string]interface{}{"checkContextTypes": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "any" is forbidden`, Line: 6, Column: 11},
			},
		},
		{
			Code: `
        const Foo = (props) => {
          return <div />;
        };
        Foo.contextTypes = {
          a: PropTypes.any
        };
      `,
			Tsx:     true,
			Options: map[string]interface{}{"checkContextTypes": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "any" is forbidden`, Line: 6, Column: 11},
			},
		},
		// contextTypes with propWrapper.
		{
			Code: `
        class Component extends React.Component {
          static contextTypes = forbidExtraProps({
            a: PropTypes.array,
            o: PropTypes.object
          });
          render() {
            return <div />;
          }
        }
      `,
			Tsx:     true,
			Options: map[string]interface{}{"checkContextTypes": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "array" is forbidden`},
				{MessageId: "forbiddenPropType", Message: `Prop type "object" is forbidden`},
			},
			Settings: map[string]interface{}{
				"propWrapperFunctions": []interface{}{"forbidExtraProps"},
			},
		},
		{
			Code: `
        class Component extends React.Component {
          static get contextTypes() {
            return forbidExtraProps({
              a: PropTypes.array,
              o: PropTypes.object
            });
          }
          render() {
            return <div />;
          }
        }
      `,
			Tsx:     true,
			Options: map[string]interface{}{"checkContextTypes": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "array" is forbidden`},
				{MessageId: "forbiddenPropType", Message: `Prop type "object" is forbidden`},
			},
			Settings: map[string]interface{}{
				"propWrapperFunctions": []interface{}{"forbidExtraProps"},
			},
		},
		{
			Code: `
        class Component extends React.Component {
          render() {
            return <div />;
          }
        }
        Component.contextTypes = forbidExtraProps({
          a: PropTypes.array,
          o: PropTypes.object
        });
      `,
			Tsx:     true,
			Options: map[string]interface{}{"checkContextTypes": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "array" is forbidden`},
				{MessageId: "forbiddenPropType", Message: `Prop type "object" is forbidden`},
			},
			Settings: map[string]interface{}{
				"propWrapperFunctions": []interface{}{"forbidExtraProps"},
			},
		},
		{
			Code: `
        function Component(props) {
          return <div />;
        }
        Component.contextTypes = forbidExtraProps({
          a: PropTypes.array,
          o: PropTypes.object
        });
      `,
			Tsx:     true,
			Options: map[string]interface{}{"checkContextTypes": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "array" is forbidden`},
				{MessageId: "forbiddenPropType", Message: `Prop type "object" is forbidden`},
			},
			Settings: map[string]interface{}{
				"propWrapperFunctions": []interface{}{"forbidExtraProps"},
			},
		},
		{
			Code: `
        const Component = (props) => {
          return <div />;
        };
        Component.contextTypes = forbidExtraProps({
          a: PropTypes.array,
          o: PropTypes.object
        });
      `,
			Tsx:     true,
			Options: map[string]interface{}{"checkContextTypes": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "array" is forbidden`},
				{MessageId: "forbiddenPropType", Message: `Prop type "object" is forbidden`},
			},
			Settings: map[string]interface{}{
				"propWrapperFunctions": []interface{}{"forbidExtraProps"},
			},
		},
		// contextTypes custom forbid: `instanceOf`.
		{
			Code: `
        var Hello = createReactClass({
          contextTypes: {
            retailer: PropTypes.instanceOf(Map).isRequired,
            requestRetailer: PropTypes.func.isRequired
          },
          render: function() {
            return <div />;
          }
        });
      `,
			Tsx: true,
			Options: map[string]interface{}{
				"forbid":            []interface{}{"instanceOf"},
				"checkContextTypes": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "instanceOf" is forbidden`},
			},
		},
		{
			Code: `
        class Component extends React.Component {
          static contextTypes = {
            retailer: PropTypes.instanceOf(Map).isRequired,
            requestRetailer: PropTypes.func.isRequired
          }
          render() {
            return <div />;
          }
        }
      `,
			Tsx: true,
			Options: map[string]interface{}{
				"forbid":            []interface{}{"instanceOf"},
				"checkContextTypes": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "instanceOf" is forbidden`},
			},
		},
		{
			Code: `
        class Component extends React.Component {
          static get contextTypes() {
            return {
              retailer: PropTypes.instanceOf(Map).isRequired,
              requestRetailer: PropTypes.func.isRequired
            };
          }
          render() {
            return <div />;
          }
        }
      `,
			Tsx: true,
			Options: map[string]interface{}{
				"forbid":            []interface{}{"instanceOf"},
				"checkContextTypes": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "instanceOf" is forbidden`},
			},
		},
		{
			Code: `
        class Component extends React.Component {
          render() {
            return <div />;
          }
        }
        Component.contextTypes = {
          retailer: PropTypes.instanceOf(Map).isRequired,
          requestRetailer: PropTypes.func.isRequired
        };
      `,
			Tsx: true,
			Options: map[string]interface{}{
				"forbid":            []interface{}{"instanceOf"},
				"checkContextTypes": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "instanceOf" is forbidden`},
			},
		},
		{
			Code: `
        function Component(props) {
          return <div />;
        }
        Component.contextTypes = {
          retailer: PropTypes.instanceOf(Map).isRequired,
          requestRetailer: PropTypes.func.isRequired
        };
      `,
			Tsx: true,
			Options: map[string]interface{}{
				"forbid":            []interface{}{"instanceOf"},
				"checkContextTypes": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "instanceOf" is forbidden`},
			},
		},
		{
			Code: `
        const Component = (props) => {
          return <div />;
        };
        Component.contextTypes = {
          retailer: PropTypes.instanceOf(Map).isRequired,
          requestRetailer: PropTypes.func.isRequired
        }
      `,
			Tsx: true,
			Options: map[string]interface{}{
				"forbid":            []interface{}{"instanceOf"},
				"checkContextTypes": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "instanceOf" is forbidden`},
			},
		},

		// ---- childContextTypes — invalid (option ON) ----
		{
			Code: `
        var First = createReactClass({
          childContextTypes: {
            a: PropTypes.any
          },
          render: function() {
            return <div />;
          }
        });
      `,
			Tsx:     true,
			Options: map[string]interface{}{"checkChildContextTypes": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "any" is forbidden`, Line: 4, Column: 13},
			},
		},
		{
			Code: `
        class Foo extends Component {
          static childContextTypes = {
            a: PropTypes.any
          }
          render() {
            return <div />;
          }
        }
      `,
			Tsx:     true,
			Options: map[string]interface{}{"checkChildContextTypes": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "any" is forbidden`, Line: 4, Column: 13},
			},
		},
		{
			Code: `
        class Foo extends Component {
          static get childContextTypes() {
            return {
              a: PropTypes.any
            };
          }
          render() {
            return <div />;
          }
        }
      `,
			Tsx:     true,
			Options: map[string]interface{}{"checkChildContextTypes": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "any" is forbidden`, Line: 5, Column: 15},
			},
		},
		{
			Code: `
        class Foo extends Component {
          render() {
            return <div />;
          }
        }
        Foo.childContextTypes = {
          a: PropTypes.any
        };
      `,
			Tsx:     true,
			Options: map[string]interface{}{"checkChildContextTypes": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "any" is forbidden`, Line: 8, Column: 11},
			},
		},
		{
			Code: `
        function Foo(props) {
          return <div />;
        }
        Foo.childContextTypes = {
          a: PropTypes.any
        };
      `,
			Tsx:     true,
			Options: map[string]interface{}{"checkChildContextTypes": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "any" is forbidden`, Line: 6, Column: 11},
			},
		},
		{
			Code: `
        const Foo = (props) => {
          return <div />;
        };
        Foo.childContextTypes = {
          a: PropTypes.any
        };
      `,
			Tsx:     true,
			Options: map[string]interface{}{"checkChildContextTypes": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "any" is forbidden`, Line: 6, Column: 11},
			},
		},
		// childContextTypes with propWrapper.
		{
			Code: `
        class Component extends React.Component {
          static childContextTypes = forbidExtraProps({
            a: PropTypes.array,
            o: PropTypes.object
          });
          render() {
            return <div />;
          }
        }
      `,
			Tsx:     true,
			Options: map[string]interface{}{"checkChildContextTypes": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "array" is forbidden`},
				{MessageId: "forbiddenPropType", Message: `Prop type "object" is forbidden`},
			},
			Settings: map[string]interface{}{
				"propWrapperFunctions": []interface{}{"forbidExtraProps"},
			},
		},
		{
			Code: `
        class Component extends React.Component {
          static get childContextTypes() {
            return forbidExtraProps({
              a: PropTypes.array,
              o: PropTypes.object
            });
          }
          render() {
            return <div />;
          }
        }
      `,
			Tsx:     true,
			Options: map[string]interface{}{"checkChildContextTypes": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "array" is forbidden`},
				{MessageId: "forbiddenPropType", Message: `Prop type "object" is forbidden`},
			},
			Settings: map[string]interface{}{
				"propWrapperFunctions": []interface{}{"forbidExtraProps"},
			},
		},
		{
			Code: `
        class Component extends React.Component {
          render() {
            return <div />;
          }
        }
        Component.childContextTypes = forbidExtraProps({
          a: PropTypes.array,
          o: PropTypes.object
        });
      `,
			Tsx:     true,
			Options: map[string]interface{}{"checkChildContextTypes": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "array" is forbidden`},
				{MessageId: "forbiddenPropType", Message: `Prop type "object" is forbidden`},
			},
			Settings: map[string]interface{}{
				"propWrapperFunctions": []interface{}{"forbidExtraProps"},
			},
		},
		{
			Code: `
        function Component(props) {
          return <div />;
        }
        Component.childContextTypes = forbidExtraProps({
          a: PropTypes.array,
          o: PropTypes.object
        });
      `,
			Tsx:     true,
			Options: map[string]interface{}{"checkChildContextTypes": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "array" is forbidden`},
				{MessageId: "forbiddenPropType", Message: `Prop type "object" is forbidden`},
			},
			Settings: map[string]interface{}{
				"propWrapperFunctions": []interface{}{"forbidExtraProps"},
			},
		},
		{
			Code: `
        const Component = (props) => {
          return <div />;
        };
        Component.childContextTypes = forbidExtraProps({
          a: PropTypes.array,
          o: PropTypes.object
        });
      `,
			Tsx:     true,
			Options: map[string]interface{}{"checkChildContextTypes": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "array" is forbidden`},
				{MessageId: "forbiddenPropType", Message: `Prop type "object" is forbidden`},
			},
			Settings: map[string]interface{}{
				"propWrapperFunctions": []interface{}{"forbidExtraProps"},
			},
		},
		// childContextTypes custom forbid: `instanceOf`.
		{
			Code: `
        var Hello = createReactClass({
          childContextTypes: {
            retailer: PropTypes.instanceOf(Map).isRequired,
            requestRetailer: PropTypes.func.isRequired
          },
          render: function() {
            return <div />;
          }
        });
      `,
			Tsx: true,
			Options: map[string]interface{}{
				"forbid":                 []interface{}{"instanceOf"},
				"checkChildContextTypes": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "instanceOf" is forbidden`},
			},
		},
		{
			Code: `
        class Component extends React.Component {
          static childContextTypes = {
            retailer: PropTypes.instanceOf(Map).isRequired,
            requestRetailer: PropTypes.func.isRequired
          }
          render() {
            return <div />;
          }
        }
      `,
			Tsx: true,
			Options: map[string]interface{}{
				"forbid":                 []interface{}{"instanceOf"},
				"checkChildContextTypes": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "instanceOf" is forbidden`},
			},
		},
		{
			Code: `
        class Component extends React.Component {
          render() {
            return <div />;
          }
        }
        Component.childContextTypes = {
          retailer: PropTypes.instanceOf(Map).isRequired,
          requestRetailer: PropTypes.func.isRequired
        };
      `,
			Tsx: true,
			Options: map[string]interface{}{
				"forbid":                 []interface{}{"instanceOf"},
				"checkChildContextTypes": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "instanceOf" is forbidden`},
			},
		},
		{
			Code: `
        function Component(props) {
          return <div />;
        }
        Component.childContextTypes = {
          retailer: PropTypes.instanceOf(Map).isRequired,
          requestRetailer: PropTypes.func.isRequired
        };
      `,
			Tsx: true,
			Options: map[string]interface{}{
				"forbid":                 []interface{}{"instanceOf"},
				"checkChildContextTypes": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "instanceOf" is forbidden`},
			},
		},
		{
			Code: `
        const Component = (props) => {
          return <div />;
        };
        Component.childContextTypes = {
          retailer: PropTypes.instanceOf(Map).isRequired,
          requestRetailer: PropTypes.func.isRequired
        };
      `,
			Tsx: true,
			Options: map[string]interface{}{
				"forbid":                 []interface{}{"instanceOf"},
				"checkChildContextTypes": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "instanceOf" is forbidden`},
			},
		},

		// ---- Imported PropTypes / packages ----
		// `import { object, string } from "prop-types"` →
		// propTypesPackageName = first specifier's local ('object'); when used
		// as bare Identifier value, isPropTypesPackage(Identifier) matches
		// via name OR `!isForeign`.
		{
			Code: `
        import { object, string } from "prop-types";
        function C({ a, b }) { return [a, b]; }
        C.propTypes = {
          a: object,
          b: string
        };
      `,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"object"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "object" is forbidden`},
			},
		},
		// `objectOf(any)` — argument-walk reports `any`.
		{
			Code: `
        import { objectOf, any } from "prop-types";
        function C({ a }) { return a; }
        C.propTypes = {
          a: objectOf(any)
        };
      `,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"any"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "any" is forbidden`},
			},
		},
		// `objectOf(any)` with `forbid: ['objectOf']` — callee unwrap reports.
		{
			Code: `
        import { objectOf, any } from "prop-types";
        function C({ a }) { return a; }
        C.propTypes = {
          a: objectOf(any)
        };
      `,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"objectOf"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "objectOf" is forbidden`},
			},
		},
		// `shape({b: any})` — CallExpression listener handles bare `shape`.
		{
			Code: `
        import { shape, any } from "prop-types";
        function C({ a }) { return a; }
        C.propTypes = {
          a: shape({
            b: any
          })
        };
      `,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"any"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "any" is forbidden`},
			},
		},
		// `PropTypes.shape({b: any})`.
		{
			Code: `
        import { any } from "prop-types";
        function C({ a }) { return a; }
        C.propTypes = {
          a: PropTypes.shape({
            b: any
          })
        };
      `,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"any"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "any" is forbidden`},
			},
		},
		// Default forbid catches nested `PropTypes.object` in `shape({...})`.
		{
			Code: `
        var First = createReactClass({
          propTypes: {
            s: PropTypes.shape({
              o: PropTypes.object
            })
          },
          render: function() {
            return <div />;
          }
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "object" is forbidden`},
			},
		},
		// `arrayOf(object)` — argument-walk reports `object`.
		{
			Code: `
        import React from './React';

        import { arrayOf, object } from 'prop-types';

        const App = ({ foo }) => (
          <div>
            Hello world {foo}
          </div>
        );

        App.propTypes = {
          foo: arrayOf(object)
        }

        export default App;
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "object" is forbidden`},
			},
		},
		// `arrayOf(PropTypes.object)` — argument is a MemberExpression →
		// argument-walk reads `.property.name` ('object').
		{
			Code: `
        import React from './React';

        import PropTypes, { arrayOf } from 'prop-types';

        const App = ({ foo }) => (
          <div>
            Hello world {foo}
          </div>
        );

        App.propTypes = {
          foo: arrayOf(PropTypes.object)
        }

        export default App;
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "object" is forbidden`},
			},
		},
		// `CustomPropTypes.shape({c: CustomPropTypes.object.isRequired})` —
		// `.isRequired` strip → `.object` → forbidden.
		{
			Code: `
        import CustomPropTypes from "prop-types";
        class Component extends React.Component {};
        Component.propTypes = {
          a: CustomPropTypes.shape({
            b: CustomPropTypes.String,
            c: CustomPropTypes.object.isRequired,
          })
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "object" is forbidden`},
			},
		},
		// `import { PropTypes as CustomPropTypes } from "react"` — react path
		// with named PropTypes specifier; alias becomes propTypesPackageName.
		{
			Code: `
        import { PropTypes as CustomPropTypes } from "react";
        class Component extends React.Component {};
        Component.propTypes = {
          a: CustomPropTypes.shape({
            b: CustomPropTypes.String,
            c: CustomPropTypes.object.isRequired,
          })
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "object" is forbidden`},
			},
		},
		// `CustomReact.PropTypes.object` — nested member-expression receiver
		// triggers the `object.name === undefined` arm of isPropTypesPackage
		// for the outer `.object` MemberExpression.
		{
			Code: `
        import CustomReact from "react"
        class Component extends React.Component {};
        Component.propTypes = {
          b: CustomReact.PropTypes.object,
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "object" is forbidden`},
			},
		},

		// ---- Locks in upstream branches that upstream tests don't cover ----
		// Locks in upstream `checkProperties` arm: `value` is a bare
		// Identifier whose name matches a forbidden type without needing an
		// import (because `!isForeign` is true by default).
		{
			Code: `
        class C extends React.Component {}
        C.propTypes = {
          a: any,
        };
      `,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"any"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "any" is forbidden`},
			},
		},
		// Locks in upstream's CallExpression listener for bare-identifier
		// callee `shape({...})` (no MemberExpression gate to skip).
		{
			Code: `
        class C extends React.Component {}
        C.propTypes = {
          a: shape({ b: any }),
        };
      `,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"any"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "any" is forbidden`},
			},
		},
		// Locks in upstream's `Object.assign({}, X, {...})` first-arg-only
		// recursion: when `Object.assign` is configured as a propWrapper,
		// `checkNode` on the CallExpression takes ONLY `arguments[0]`. The
		// `{a: PropTypes.any}` in the third position is invisible. Reports
		// from this code only fire if the FIRST argument contains a
		// forbidden type.
		{
			Code: `
        function Foo() { return null; }
        Foo.propTypes = Object.assign({a: PropTypes.any}, X.propTypes, {b: PropTypes.array});
      `,
			Tsx: true,
			Settings: map[string]interface{}{
				"propWrapperFunctions": []interface{}{
					map[string]interface{}{"object": "Object", "property": "assign"},
				},
			},
			// Only the first arg's forbidden value is reported; `b: PropTypes.array`
			// in arg index 2 is silently skipped — match upstream literally.
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "any" is forbidden`},
			},
		},
		// Locks in upstream's single-pass `.isRequired` strip: only ONE
		// level is unwrapped. `PropTypes.any.isRequired.isRequired` (legal
		// syntactically; nonsense semantically) strips the outer
		// `.isRequired` once, leaving `PropTypes.any.isRequired` whose
		// `.property.name === 'isRequired'`. Then `isPropTypesPackage`
		// matches the inner MemberExpression, target = 'isRequired' — but
		// we report on `any` because the original arg-walk reads the
		// first `.property.name` after one strip. Lock the actual upstream
		// behavior: outer strip → value = `PropTypes.any.isRequired`,
		// then the value.property.name='isRequired' branch is NOT
		// re-applied. target ends up 'isRequired'. So 'any' is NOT
		// reported — but 'isRequired' would be reported if forbidden.
		{
			Code: `
        class C extends React.Component {}
        C.propTypes = {
          a: PropTypes.any.isRequired.isRequired,
        };
      `,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"isRequired"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "isRequired" is forbidden`},
			},
		},
		// Locks in upstream's "any function named `shape` is treated as a
		// PropTypes shape call" behavior. A user-defined local function
		// called `shape` with NO connection to PropTypes still triggers
		// the CallExpression listener's shape branch — and recurses into
		// its first argument's properties. Forbidden values inside ARE
		// reported.
		{
			Code: `
        function shape(obj) { return obj; }
        const x = shape({ a: PropTypes.any });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "any" is forbidden`},
			},
		},
		// Locks in upstream's "ObjectExpression listener fires on ANY
		// object literal" behavior. A non-component config object that
		// happens to contain a `propTypes` key with object value is
		// reported, even though it's unrelated to React.
		{
			Code: `
        const config = {
          propTypes: {
            a: PropTypes.any,
          },
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "any" is forbidden`},
			},
		},
		// Locks in upstream's "last `prop-types` import wins for
		// propTypesPackageName" behavior. Two imports from the same
		// module overwrite each other; the second import's first
		// specifier is the resolved name.
		{
			Code: `
        import x from "prop-types";
        import { PropTypes } from "prop-types";
        class C extends React.Component {}
        C.propTypes = { a: PropTypes.any };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "any" is forbidden`},
			},
		},

		// ---- Strict Line/Column/EndLine/EndColumn assertions ----
		// Reasoning: upstream reports on the Property node (the entire
		// `key: value` pair). tsgo's PropertyAssignment Pos/End match
		// the same trimmed range. Lock the exact 1-based positions for
		// every container shape the rule emits into.

		// `createReactClass({propTypes: {...}})` — single forbidden, default forbid.
		{
			Code: `
        var First = createReactClass({
          propTypes: {
            a: PropTypes.any
          },
          render: function() { return <div />; }
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenPropType",
					Message:   `Prop type "any" is forbidden`,
					Line:      4, Column: 13, EndLine: 4, EndColumn: 29,
				},
			},
		},
		// Class field `static propTypes = {...}` — column anchored to property start.
		{
			Code: `
        class C extends React.Component {
          static propTypes = {
            a: PropTypes.any
          };
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenPropType",
					Message:   `Prop type "any" is forbidden`,
					Line:      4, Column: 13, EndLine: 4, EndColumn: 29,
				},
			},
		},
		// Static getter — return-statement value's properties.
		{
			Code: `
        class C extends React.Component {
          static get propTypes() {
            return {
              a: PropTypes.any
            };
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenPropType",
					Message:   `Prop type "any" is forbidden`,
					Line:      5, Column: 15, EndLine: 5, EndColumn: 31,
				},
			},
		},
		// `Foo.propTypes = {...}` after class — separate statement.
		{
			Code: `
        class Foo extends React.Component {}
        Foo.propTypes = {
          a: PropTypes.any
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenPropType",
					Message:   `Prop type "any" is forbidden`,
					Line:      4, Column: 11, EndLine: 4, EndColumn: 27,
				},
			},
		},
		// contextTypes — same column rules apply.
		{
			Code: `
        class Foo extends React.Component {}
        Foo.contextTypes = {
          a: PropTypes.any
        };
      `,
			Tsx:     true,
			Options: map[string]interface{}{"checkContextTypes": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenPropType",
					Message:   `Prop type "any" is forbidden`,
					Line:      4, Column: 11, EndLine: 4, EndColumn: 27,
				},
			},
		},
		// childContextTypes — anchored to the same shape.
		{
			Code: `
        class Foo extends React.Component {}
        Foo.childContextTypes = {
          a: PropTypes.any
        };
      `,
			Tsx:     true,
			Options: map[string]interface{}{"checkChildContextTypes": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenPropType",
					Message:   `Prop type "any" is forbidden`,
					Line:      4, Column: 11, EndLine: 4, EndColumn: 27,
				},
			},
		},
		// `.isRequired` chain — Property node's range is unchanged by the
		// inner unwrap. Position still anchors to the full `a: ...` range.
		{
			Code: `
        class C extends React.Component {}
        C.propTypes = {
          a: PropTypes.any.isRequired
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenPropType",
					Message:   `Prop type "any" is forbidden`,
					Line:      4, Column: 11, EndLine: 4, EndColumn: 38,
				},
			},
		},
		// `shape({...})` — inner property reported at inner location, not outer.
		{
			Code: `
        class C extends React.Component {}
        C.propTypes = {
          a: PropTypes.shape({
            b: PropTypes.any
          })
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenPropType",
					Message:   `Prop type "any" is forbidden`,
					Line:      5, Column: 13, EndLine: 5, EndColumn: 29,
				},
			},
		},
		// Identifier value position — entire `retailer: object` Property range.
		{
			Code: `
        class C extends React.Component {}
        C.propTypes = {
          retailer: object,
        };
      `,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"object"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenPropType",
					Message:   `Prop type "object" is forbidden`,
					Line:      4, Column: 11, EndLine: 4, EndColumn: 27,
				},
			},
		},
		// `arrayOf(PropTypes.any)` — argument-walk reports on the
		// outer Property node, not the inner argument.
		{
			Code: `
        class C extends React.Component {}
        C.propTypes = {
          a: arrayOf(PropTypes.any)
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenPropType",
					Message:   `Prop type "any" is forbidden`,
					Line:      4, Column: 11, EndLine: 4, EndColumn: 36,
				},
			},
		},

		// ---- Strict ordering across multiple errors ----
		// Reasoning: upstream visits properties in source order. Two
		// forbidden in one object → errors must be in source order.
		// Lock the exact line/column for each.

		// Two forbidden, same object — order: first by line, then by column.
		{
			Code: `
        class C extends React.Component {}
        C.propTypes = {
          a: PropTypes.array,
          o: PropTypes.object
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenPropType",
					Message:   `Prop type "array" is forbidden`,
					Line:      4, Column: 11, EndLine: 4, EndColumn: 29,
				},
				{
					MessageId: "forbiddenPropType",
					Message:   `Prop type "object" is forbidden`,
					Line:      5, Column: 11, EndLine: 5, EndColumn: 30,
				},
			},
		},
		// Two separate components — outer source order.
		{
			Code: `
        class A extends React.Component {}
        A.propTypes = {
          a: PropTypes.any
        };
        class B extends React.Component {}
        B.propTypes = {
          b: PropTypes.array
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenPropType",
					Message:   `Prop type "any" is forbidden`,
					Line:      4, Column: 11, EndLine: 4, EndColumn: 27,
				},
				{
					MessageId: "forbiddenPropType",
					Message:   `Prop type "array" is forbidden`,
					Line:      8, Column: 11, EndLine: 8, EndColumn: 29,
				},
			},
		},
		// Three forbidden in nested shape — outer first, then inner two
		// in source order.
		{
			Code: `
        class C extends React.Component {}
        C.propTypes = {
          a: PropTypes.any,
          b: PropTypes.shape({
            c: PropTypes.array,
            d: PropTypes.object
          })
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenPropType",
					Message:   `Prop type "any" is forbidden`,
					Line:      4, Column: 11, EndLine: 4, EndColumn: 27,
				},
				{
					MessageId: "forbiddenPropType",
					Message:   `Prop type "array" is forbidden`,
					Line:      6, Column: 13, EndLine: 6, EndColumn: 31,
				},
				{
					MessageId: "forbiddenPropType",
					Message:   `Prop type "object" is forbidden`,
					Line:      7, Column: 13, EndLine: 7, EndColumn: 32,
				},
			},
		},
		// Mixed propTypes + contextTypes + childContextTypes on one class.
		// Source order: propTypes (line 3) → contextTypes (line 6) →
		// childContextTypes (line 9).
		{
			Code: `
        class Foo extends React.Component {
          static propTypes = {
            a: PropTypes.any
          };
          static contextTypes = {
            b: PropTypes.array
          };
          static childContextTypes = {
            c: PropTypes.object
          };
        }
      `,
			Tsx: true,
			Options: map[string]interface{}{
				"checkContextTypes":      true,
				"checkChildContextTypes": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenPropType",
					Message:   `Prop type "any" is forbidden`,
					Line:      4, Column: 13, EndLine: 4, EndColumn: 29,
				},
				{
					MessageId: "forbiddenPropType",
					Message:   `Prop type "array" is forbidden`,
					Line:      7, Column: 13, EndLine: 7, EndColumn: 31,
				},
				{
					MessageId: "forbiddenPropType",
					Message:   `Prop type "object" is forbidden`,
					Line:      10, Column: 13, EndLine: 10, EndColumn: 32,
				},
			},
		},
		// Argument-walk reports preserve order: `oneOf(string, any, array)`
		// emits 'any' before 'array' (and skips non-forbidden 'string').
		{
			Code: `
        class C extends React.Component {}
        C.propTypes = {
          a: oneOf(string, any, array)
        };
      `,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"any", "array"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenPropType",
					Message:   `Prop type "any" is forbidden`,
					Line:      4, Column: 11, EndLine: 4, EndColumn: 39,
				},
				{
					MessageId: "forbiddenPropType",
					Message:   `Prop type "array" is forbidden`,
					Line:      4, Column: 11, EndLine: 4, EndColumn: 39,
				},
			},
		},
		// Multi-line property — EndLine/EndColumn span the whole property.
		{
			Code: `
        class C extends React.Component {}
        C.propTypes = {
          a:
            PropTypes
              .any
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenPropType",
					Message:   `Prop type "any" is forbidden`,
					Line:      4, Column: 11, EndLine: 6, EndColumn: 19,
				},
			},
		},

		// ---- Robustness: TS expression wrappers — invalid ----
		// `Foo!.propTypes = {...}` with forbidden value.
		{
			Code: `
        declare const Foo: any;
        Foo!.propTypes = { a: PropTypes.any };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "any" is forbidden`},
			},
		},
		// `(Foo as any).propTypes = {...}`.
		{
			Code: `
        declare const Foo: any;
        (Foo as any).propTypes = { a: PropTypes.any };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "any" is forbidden`},
			},
		},
		// Paren-wrapped LHS.
		{
			Code: `
        class Foo extends React.Component {}
        (Foo.propTypes) = { a: PropTypes.any };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "any" is forbidden`},
			},
		},
		// Value-side wrapper: `(PropTypes.any as any)`.
		{
			Code: `
        class C extends React.Component {}
        C.propTypes = { a: (PropTypes.any as any) };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "any" is forbidden`},
			},
		},
		// Value-side double-paren.
		{
			Code: `
        class C extends React.Component {}
        C.propTypes = { a: ((PropTypes.any)) };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "any" is forbidden`},
			},
		},
		// Optional chain on value: `PropTypes?.any`. tsgo flags optional
		// chain on the PAE; getPropertyAccessName still reads the name.
		{
			Code: `
        class C extends React.Component {}
        C.propTypes = { a: PropTypes?.any };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "any" is forbidden`},
			},
		},
		// Non-null assertion on value: `PropTypes.any!`.
		{
			Code: `
        class C extends React.Component {}
        C.propTypes = { a: PropTypes.any! };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "any" is forbidden`},
			},
		},
		// `.isRequired` chain on a wrapped expression.
		{
			Code: `
        class C extends React.Component {}
        C.propTypes = { a: (PropTypes.any).isRequired };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "any" is forbidden`},
			},
		},

		// ---- Robustness: declaration / nesting forms — invalid ----
		// Class expression with static field.
		{
			Code: `
        const X = class extends React.Component {
          static propTypes = { a: PropTypes.any };
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "any" is forbidden`},
			},
		},
		// Anonymous class expression assigned to variable, then propTypes
		// assignment via .propTypes.
		{
			Code: `
        const Foo = class extends React.Component {};
        Foo.propTypes = { a: PropTypes.array };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "array" is forbidden`},
			},
		},
		// Nested classes — both report independently.
		{
			Code: `
        class Outer extends React.Component {
          static propTypes = { a: PropTypes.any };
          render() {
            class Inner extends React.Component {
              static propTypes = { b: PropTypes.array };
            }
            return null;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "any" is forbidden`},
				{MessageId: "forbiddenPropType", Message: `Prop type "array" is forbidden`},
			},
		},
		// Static getter inside class expression.
		{
			Code: `
        const X = class extends React.Component {
          static get propTypes() {
            return { a: PropTypes.any };
          }
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "any" is forbidden`},
			},
		},
		// Method body with multiple returns — last return wins (matches
		// upstream's `loopNodes` which iterates from last to first).
		{
			Code: `
        class C extends React.Component {
          static get propTypes() {
            if (foo) {
              return { a: PropTypes.string };  // not reached by lookup
            }
            return { a: PropTypes.any };  // last return — checked
          };
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "any" is forbidden`},
			},
		},

		// ---- Robustness: assignment operators — invalid ----
		// `Foo.propTypes ||= {...}` — compound assign with forbidden value.
		{
			Code: `
        class Foo extends React.Component {}
        Foo.propTypes ||= { a: PropTypes.any };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "any" is forbidden`},
			},
		},
		// `Foo.propTypes ??= {...}`.
		{
			Code: `
        class Foo extends React.Component {}
        Foo.propTypes ??= { a: PropTypes.any };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "any" is forbidden`},
			},
		},

		// ---- Robustness: scope / Identifier resolution — invalid ----
		// Top-level let with object literal.
		{
			Code: `
        let propTypes = { a: PropTypes.any };
        class C extends React.Component {}
        C.propTypes = propTypes;
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "any" is forbidden`},
			},
		},
		// Scope shadowing — closest binding wins. Both assignments and
		// the inner `propTypes` are in the same function scope, so the
		// inner shadow IS what the assignment resolves to (any is reported).
		{
			Code: `
        const propTypes = { a: PropTypes.string };
        function f() {
          const propTypes = { a: PropTypes.any };
          class C extends React.Component {}
          C.propTypes = propTypes;  // resolves to inner (shadowed) binding
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "any" is forbidden`},
			},
		},
		// Switch-statement return: forbidden in last case wins.
		// Mirrors upstream's `loopNodes` recursion into the last case
		// clause's consequent.
		{
			Code: `
        class C extends React.Component {
          static get propTypes() {
            switch (mode) {
              case "a":
                return { x: PropTypes.string };
              default:
                return { x: PropTypes.any };
            }
          };
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "any" is forbidden`},
			},
		},
		// Switch with no default — last case clause's return is reached.
		{
			Code: `
        class C extends React.Component {
          static get propTypes() {
            switch (mode) {
              case "a":
                return { x: PropTypes.string };
              case "b":
                return { x: PropTypes.array };
            }
          };
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "array" is forbidden`},
			},
		},

		// ---- Robustness: deeply nested (real user shapes) — invalid ----
		// Deeply nested shape with one forbidden type at depth 3.
		{
			Code: `
        class C extends React.Component {}
        C.propTypes = {
          user: PropTypes.shape({
            profile: PropTypes.shape({
              metadata: PropTypes.any,
            }),
          }),
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "any" is forbidden`},
			},
		},
		// `arrayOf(shape({...with forbidden inside...}))`.
		{
			Code: `
        class C extends React.Component {}
        C.propTypes = {
          users: PropTypes.arrayOf(PropTypes.shape({
            metadata: PropTypes.object,
          })),
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "object" is forbidden`},
			},
		},
		// Multiple forbidden inside one shape.
		{
			Code: `
        class C extends React.Component {}
        C.propTypes = {
          x: PropTypes.shape({
            a: PropTypes.any,
            b: PropTypes.array,
            c: PropTypes.object,
          }),
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "any" is forbidden`},
				{MessageId: "forbiddenPropType", Message: `Prop type "array" is forbidden`},
				{MessageId: "forbiddenPropType", Message: `Prop type "object" is forbidden`},
			},
		},
		// Two-level shape with forbidden at outer + inner.
		{
			Code: `
        class C extends React.Component {}
        C.propTypes = {
          a: PropTypes.any,
          b: PropTypes.shape({
            c: PropTypes.object,
          }),
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "any" is forbidden`},
				{MessageId: "forbiddenPropType", Message: `Prop type "object" is forbidden`},
			},
		},

		// ---- Robustness: import variations — invalid ----
		// Namespace import: `import * as PT from "prop-types"`.
		{
			Code: `
        import * as PT from "prop-types";
        class C extends React.Component {}
        C.propTypes = { a: PT.any };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "any" is forbidden`},
			},
		},
		// Aliased named import from react.
		{
			Code: `
        import { PropTypes as PT } from "react";
        class C extends React.Component {}
        C.propTypes = { a: PT.any };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "any" is forbidden`},
			},
		},
		// Default + named imports from prop-types.
		{
			Code: `
        import PropTypes, { shape } from "prop-types";
        class C extends React.Component {}
        C.propTypes = { a: shape({ b: PropTypes.any }) };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "any" is forbidden`},
			},
		},
		// Side-effect-only import doesn't update propTypesPackageName but
		// `!isForeign` keeps the rule on by default.
		{
			Code: `
        import "prop-types";
        class C extends React.Component {}
        C.propTypes = { a: PropTypes.any };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "any" is forbidden`},
			},
		},

		// ---- Robustness: real-world component patterns — invalid ----
		// Functional component with .propTypes assignment + forbidden type.
		{
			Code: `
        function Foo(props) { return null; }
        Foo.propTypes = {
          data: PropTypes.any,
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "any" is forbidden`},
			},
		},
		// Arrow function component with .propTypes assignment.
		{
			Code: `
        const Foo = (props) => null;
        Foo.propTypes = {
          data: PropTypes.array,
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "array" is forbidden`},
			},
		},
		// `forwardRef`-wrapped component with forbidden propType.
		{
			Code: `
        const Foo = React.forwardRef((props, ref) => null);
        Foo.propTypes = {
          data: PropTypes.object,
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "object" is forbidden`},
			},
		},

		// ---- Robustness: forbid list / options — invalid ----
		// Custom forbid list including stack of types.
		{
			Code: `
        class C extends React.Component {}
        C.propTypes = {
          a: PropTypes.func,
          b: PropTypes.symbol,
          c: PropTypes.element,
        };
      `,
			Tsx: true,
			Options: map[string]interface{}{
				"forbid": []interface{}{"func", "symbol", "element"},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "func" is forbidden`},
				{MessageId: "forbiddenPropType", Message: `Prop type "symbol" is forbidden`},
				{MessageId: "forbiddenPropType", Message: `Prop type "element" is forbidden`},
			},
		},
		// Custom forbid alongside non-string entries (drop-and-keep semantics).
		{
			Code: `
        class C extends React.Component {}
        C.propTypes = { a: PropTypes.any };
      `,
			Tsx: true,
			Options: map[string]interface{}{
				"forbid": []interface{}{42, "any", nil},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "any" is forbidden`},
			},
		},

		// ---- Robustness: combined options ----
		// Both checkContextTypes AND checkChildContextTypes ON; both fire.
		{
			Code: `
        class Foo extends React.Component {}
        Foo.contextTypes = { a: PropTypes.any };
        Foo.childContextTypes = { b: PropTypes.array };
      `,
			Tsx: true,
			Options: map[string]interface{}{
				"checkContextTypes":      true,
				"checkChildContextTypes": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "any" is forbidden`},
				{MessageId: "forbiddenPropType", Message: `Prop type "array" is forbidden`},
			},
		},
		// All three (propTypes + both contexts) on a single class.
		{
			Code: `
        class Foo extends React.Component {
          static propTypes = { a: PropTypes.any };
          static contextTypes = { b: PropTypes.array };
          static childContextTypes = { c: PropTypes.object };
          render() { return null; }
        }
      `,
			Tsx: true,
			Options: map[string]interface{}{
				"checkContextTypes":      true,
				"checkChildContextTypes": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "any" is forbidden`},
				{MessageId: "forbiddenPropType", Message: `Prop type "array" is forbidden`},
				{MessageId: "forbiddenPropType", Message: `Prop type "object" is forbidden`},
			},
		},
		// Multiple separate components in same file with mixed shapes.
		{
			Code: `
        class A extends React.Component {
          static propTypes = { x: PropTypes.any };
        }
        function B(props) { return null; }
        B.propTypes = { x: PropTypes.array };
        const C = (props) => null;
        C.propTypes = { x: PropTypes.object };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "any" is forbidden`},
				{MessageId: "forbiddenPropType", Message: `Prop type "array" is forbidden`},
				{MessageId: "forbiddenPropType", Message: `Prop type "object" is forbidden`},
			},
		},

		// ---- Robustness: argument-walk variations ----
		// `arrayOf(any)` — argument-walk reports `any` (Identifier arg).
		{
			Code: `
        import { arrayOf, any } from "prop-types";
        class C extends React.Component {}
        C.propTypes = { a: arrayOf(any) };
      `,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"any"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "any" is forbidden`},
			},
		},
		// `oneOf` with multiple Identifier arguments — only one is forbidden.
		{
			Code: `
        import { oneOf } from "prop-types";
        class C extends React.Component {}
        C.propTypes = { a: oneOf(string, any, number) };
      `,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"any"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "any" is forbidden`},
			},
		},

		// ---- Robustness: createReactClass + checkContextTypes ----
		{
			Code: `
        var First = createReactClass({
          contextTypes: {
            a: PropTypes.array,
            b: PropTypes.object
          },
          render: function() { return <div />; }
        });
      `,
			Tsx: true,
			Options: map[string]interface{}{
				"checkContextTypes": true,
				"forbid":            []interface{}{"array", "object"},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Message: `Prop type "array" is forbidden`},
				{MessageId: "forbiddenPropType", Message: `Prop type "object" is forbidden`},
			},
		},
	})
}

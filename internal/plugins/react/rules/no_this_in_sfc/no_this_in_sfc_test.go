package no_this_in_sfc

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoThisInSfcRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoThisInSfcRule, []rule_tester.ValidTestCase{
		// ---- Upstream: function component without `this` ----
		{Code: `
        function Foo(props) {
          const { foo } = props;
          return <div bar={foo} />;
        }
      `, Tsx: true},

		// ---- Upstream: destructured-parameter function component ----
		{Code: `
        function Foo({ foo }) {
          return <div bar={foo} />;
        }
      `, Tsx: true},

		// ---- Upstream: ES6 class component using `this.props` legitimately ----
		{Code: `
        class Foo extends React.Component {
          render() {
            const { foo } = this.props;
            return <div bar={foo} />;
          }
        }
      `, Tsx: true},

		// ---- Upstream: createReactClass (default ES5 factory) ----
		{Code: `
        const Foo = createReactClass({
          render: function() {
            return <div>{this.props.foo}</div>;
          }
        });
      `, Tsx: true},

		// ---- Upstream: pragma-qualified createClass via settings ----
		{
			Code: `
        const Foo = React.createClass({
          render: function() {
            return <div>{this.props.foo}</div>;
          }
        });
      `,
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"createClass": "createClass"}},
		},

		// ---- Upstream: regular non-component function may use `this` freely ----
		{Code: `
        function foo(bar) {
          this.bar = bar;
          this.props = 'baz';
          this.getFoo = function() {
            return this.bar + this.props;
          }
        }
      `, Tsx: true},

		// ---- Upstream: ConditionalExpression returning JSX-or-null is OK without this ----
		{Code: `
        function Foo(props) {
          return props.foo ? <span>{props.bar}</span> : null;
        }
      `, Tsx: true},

		// ---- Upstream: if/return JSX, no this ----
		{Code: `
        function Foo(props) {
          if (props.foo) {
            return <div>{props.bar}</div>;
          }
          return null;
        }
      `, Tsx: true},

		// ---- Upstream: if branches with side effects, no this ----
		{Code: `
        function Foo(props) {
          if (props.foo) {
            something();
          }
          return null;
        }
      `, Tsx: true},

		// ---- Upstream: arrow expression body — no this ----
		{Code: `const Foo = (props) => <span>{props.foo}</span>`, Tsx: true},
		{Code: `const Foo = ({ foo }) => <span>{foo}</span>`, Tsx: true},
		{Code: `const Foo = (props) => props.foo ? <span>{props.bar}</span> : null;`, Tsx: true},
		{Code: `const Foo = ({ foo, bar }) => foo ? <span>{bar}</span> : null;`, Tsx: true},

		// ---- Upstream: arrow inside non-React class method — `this` allowed ----
		{Code: `
        class Foo {
          bar() {
            () => {
              this.something();
              return null;
            };
          }
        }
      `, Tsx: true},

		// ---- Upstream: class-field arrow (TS class fields) ----
		{Code: `
        class Foo {
          bar = () => {
            this.something();
            return null;
          };
        }
      `, Tsx: true},

		// ---- Upstream: arrow returns object whose method uses `this` —
		// the inner method has lowercase key (`renderNode`) so it isn't an SFC,
		// and the outer arrow returns an object (not JSX/null) so it isn't an
		// SFC either. ----
		{Code: `
        export const Example = ({ prop }) => {
          return {
            handleClick: () => {},
            renderNode() {
              return <div onClick={this.handleClick} />;
            },
          };
        };
      `, Tsx: true},

		// ---- Upstream: Meteor ValidatedMethod-style — `run()` is a property method
		// of a non-component CallExpression argument; outer call isn't a wrapper. ----
		{Code: `
        export const prepareLogin = new ValidatedMethod({
          name: "user.prepare",
          validate: new SimpleSchema({
          }).validator(),
          run({ remember }) {
              if (Meteor.isServer) {
                  const connectionId = this.connection.id;
                  return Methods.prepareLogin(connectionId, remember);
              }
              return null;
          },
        });
      `, Tsx: true},

		// ---- Upstream: assignment of FE returning `this.a || null` to obj.notAComponent —
		// rightmost MemberExpression name is lowercase → not a component. ----
		{Code: `
        obj.notAComponent = function () {
          return this.a || null;
        };
      `, Tsx: true},

		// ---- Upstream: jQuery plugin idiom with TS return type — `$.fn.x = function...` ----
		{Code: `
        $.fn.getValueAsStringWeak = function (): string | null {
          const val = this.length === 1 ? this.val() : null;

          return typeof val === 'string' ? val : null;
        };
      `, Tsx: true},

		// ---- Edge (universal): lowercase function declaration cannot be SFC ----
		{Code: `
        function foo(props) {
          return <div>{this.props.foo}</div>;
        }
      `, Tsx: true},

		// ---- Edge (universal): capitalized function returning a string is not SFC ----
		{Code: `
        function Foo(props) {
          return this.props.foo;
        }
      `, Tsx: true},

		// ---- Edge (tsgo Dimension 4): `(this).foo` — paren-wrapped receiver still skipped via SkipParentheses
		// in non-component context, no report ----
		{Code: `
        function regular() {
          return (this).foo;
        }
      `, Tsx: true},

		// ---- Edge (universal): `this` in inner FunctionDeclaration of non-component outer
		// — no enclosing SFC up the chain ----
		{Code: `
        function regular() {
          function inner() {
            return this.foo;
          }
          return inner();
        }
      `, Tsx: true},

		// ---- Edge: PropertyAssignment with capitalized key holding a function returning JSX
		// makes the function-value an SFC, but its parent is PropertyAssignment, so the
		// "Property" carve-out skips reporting (mirrors upstream test #16). ----
		{Code: `
        export const obj = {
          Renderer: function() {
            return <div>{this.x}</div>;
          },
        };
      `, Tsx: true},

		// ---- Edge: capitalized MethodDeclaration on object literal — also under the
		// "Property" carve-out (MethodDeclaration parent is ObjectLiteralExpression). ----
		{Code: `
        export const obj = {
          Renderer() {
            return <div>{this.x}</div>;
          },
        };
      `, Tsx: true},

		// ---- Boundary (TS-only): `(this as any).x` — receiver is TS as-expression.
		// tsgo wraps in KindAsExpression; ESTree wraps in TSAsExpression; neither
		// is a ThisKeyword/ThisExpression, so neither tool reports. Locked in. ----
		{Code: `
        function Foo(props) {
          return <div>{(this as any).x}</div>;
        }
      `, Tsx: true},

		// ---- Boundary (TS-only): `this!.x` — non-null assertion wraps the receiver.
		// tsgo: KindNonNullExpression; ESTree: TSNonNullExpression. Neither is
		// ThisExpression/ThisKeyword → no report. ----
		{Code: `
        function Foo(props) {
          return <div>{this!.x}</div>;
        }
      `, Tsx: true},

		// ---- Boundary: SFC returning a non-JSX value (string) is not an SFC,
		// so `this.x` inside is not flagged. ----
		{Code: `
        function Foo() {
          this.x;
          return "hi";
        }
      `, Tsx: true},

	}, []rule_tester.InvalidTestCase{
		// ---- Upstream: `const { foo } = this.props` in SFC ----
		{
			Code: `
        function Foo(props) {
          const { foo } = this.props;
          return <div>{foo}</div>;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noThisInSFC", Message: "Stateless functional components should not use `this`", Line: 3, Column: 27, EndLine: 3, EndColumn: 37},
			},
		},

		// ---- Upstream: `this.props.foo` in JSX expression ----
		{
			Code: `
        function Foo(props) {
          return <div>{this.props.foo}</div>;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noThisInSFC", Line: 3, Column: 24, EndLine: 3, EndColumn: 34},
			},
		},

		// ---- Upstream: `this.state.foo` in JSX ----
		{
			Code: `
        function Foo(props) {
          return <div>{this.state.foo}</div>;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noThisInSFC", Line: 3, Column: 24, EndLine: 3, EndColumn: 34},
			},
		},

		// ---- Upstream: destructure from this.state ----
		{
			Code: `
        function Foo(props) {
          const { foo } = this.state;
          return <div>{foo}</div>;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noThisInSFC"},
			},
		},

		// ---- Upstream: ConditionalExpression with `this.props` in WhenTrue ----
		{
			Code: `
        function Foo(props) {
          return props.foo ? <div>{this.props.bar}</div> : null;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noThisInSFC"},
			},
		},

		// ---- Upstream: if branch with this.props ----
		{
			Code: `
        function Foo(props) {
          if (props.foo) {
            return <div>{this.props.bar}</div>;
          }
          return null;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noThisInSFC"},
			},
		},

		// ---- Upstream: this.props inside if test ----
		{
			Code: `
        function Foo(props) {
          if (this.props.foo) {
            something();
          }
          return null;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noThisInSFC"},
			},
		},

		// ---- Upstream: arrow returning JSX with this.props in JSX ----
		{
			Code: `const Foo = (props) => <span>{this.props.foo}</span>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noThisInSFC", Line: 1, Column: 31, EndLine: 1, EndColumn: 41},
			},
		},

		// ---- Upstream: arrow returning ConditionalExpression with this.props in test ----
		{
			Code: `const Foo = (props) => this.props.foo ? <span>{props.bar}</span> : null;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noThisInSFC", Line: 1, Column: 24, EndLine: 1, EndColumn: 34},
			},
		},

		// ---- Upstream: nested non-SFC function inside SFC — both `this.props` accesses
		// resolve to the outer SFC. Two reports (one inside `onClick`, one in JSX). ----
		{
			Code: `
        function Foo(props) {
          function onClick(bar) {
            this.props.onClick();
          }
          return <div onClick={onClick}>{this.props.foo}</div>;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noThisInSFC"},
				{MessageId: "noThisInSFC"},
			},
		},

		// ---- Lock-in: bracket access `this['props']` is also flagged (ESTree
		// MemberExpression covers both forms; tsgo splits into PropertyAccessExpression
		// and ElementAccessExpression — listener on both keeps parity). ----
		{
			Code: `
        function Foo(props) {
          return <div>{this['props'].foo}</div>;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noThisInSFC"},
			},
		},

		// ---- Lock-in: paren-wrapped receiver `(this).x` — tsgo preserves parens,
		// SkipParentheses unwraps, so this matches just like ESTree's flattened form. ----
		{
			Code: `
        function Foo(props) {
          return <div>{((this)).props.foo}</div>;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noThisInSFC"},
			},
		},

		// ---- Lock-in: optional chain `this?.foo` — tsgo flags the access via flag, no
		// ChainExpression wrapper; receiver is still ThisKeyword. ----
		{
			Code: `
        function Foo(props) {
          return <div>{this?.props}</div>;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noThisInSFC"},
			},
		},

		// ---- Lock-in: anonymous `export default function() { ... }` is an SFC
		// (matches upstream's `!node.id || capitalized(node.id.name)` branch). ----
		{
			Code: `
        export default function() {
          return <div>{this.props.foo}</div>;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noThisInSFC"},
			},
		},

		// ---- Lock-in: pragma wrapper `React.memo(() => <div>{this.x}</div>)`
		// — wrapped functions are SFCs regardless of position. ----
		{
			Code: `
        const Foo = React.memo(() => <div>{this.props.foo}</div>);
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noThisInSFC"},
			},
		},

		// ---- Lock-in: bare `forwardRef(...)` wrapper. ----
		{
			Code: `
        const Foo = forwardRef((props, ref) => <div ref={ref}>{this.props.foo}</div>);
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noThisInSFC"},
			},
		},

		// ---- Boundary: forwardRef wrapping a NAMED FunctionExpression.
		// Wrapper-call branch in IsStatelessReactComponent classifies regardless
		// of name. ----
		{
			Code: `
        const Foo = forwardRef(function Inner(props, ref) {
          return <div ref={ref}>{this.props.foo}</div>;
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noThisInSFC"},
			},
		},

		// ---- Boundary: SFC returning a JSX Fragment (not just an Element).
		// `<>` is KindJsxFragment in tsgo and JSXFragment in ESTree — both
		// classify as JSX. ----
		{
			Code: `
        function Foo() {
          return <>{this.props.foo}</>;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noThisInSFC"},
			},
		},

		// ---- Boundary: nested SFCs — inner `Inner` is itself an SFC, so
		// `this.x` inside Inner resolves to Inner (not Outer). ----
		{
			Code: `
        function Outer() {
          function Inner() {
            return <span>{this.x}</span>;
          }
          return <Inner />;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noThisInSFC"},
			},
		},

		// ---- Boundary: deep member chain `this.foo.bar.baz` — only the
		// innermost `this.foo` MemberExpression has ThisKeyword as receiver,
		// so exactly one report fires. ----
		{
			Code: `
        function Foo() {
          return <div>{this.foo.bar.baz}</div>;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noThisInSFC"},
			},
		},

		// ---- Boundary: ElementAccess via computed key — `this[Symbol.iterator]`
		// — the bracket form is also a MemberExpression in ESTree and an
		// ElementAccessExpression in tsgo; both flag it. ----
		{
			Code: `
        function Foo() {
          return <div>{this[Symbol.iterator]}</div>;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noThisInSFC"},
			},
		},

		// ---- Boundary: `obj.Foo = function () { return <div>{this.x}</div> }`
		// — MemberExpression LHS with capitalized rightmost name; isMEAssign
		// path classifies as SFC and parent is BinaryExpression (not Property),
		// so it is reported. ----
		{
			Code: `
        obj.Foo = function () {
          return <div>{this.x}</div>;
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noThisInSFC"},
			},
		},

		// ---- Boundary: anonymous arrow exported as default — arrow has
		// ExportAssignment parent; Branch 1 classifies on strict isReturningJSX. ----
		{
			Code: `
        export default () => <div>{this.props.foo}</div>;
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noThisInSFC"},
			},
		},

		// ---- Boundary: SFC inside an Outer non-component arrow that returns
		// the SFC — the inner arrow is in a `ReturnStatement` allowed position,
		// strictly returns JSX, and Outer wrapping it is non-component (returns
		// the inner arrow, not JSX). Inner classified as SFC; report fires. ----
		{
			Code: `
        const make = () => {
          const Foo = () => <div>{this.props.x}</div>;
          return Foo;
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noThisInSFC"},
			},
		},

		// ---- Boundary: capitalized FunctionDeclaration returning ONLY `null`
		// with a `this` access. Upstream auto-registers the SFC via the
		// FunctionDeclaration detection instruction (no JSX is required for
		// registration), so `components.get(sfc)` is truthy and the rule
		// reports. rslint reaches the same conclusion through
		// IsStatelessReactComponent (capitalized + isReturningJSXOrNull). ----
		{
			Code: `
        function Foo() {
          if (this.x) return null;
          return null;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noThisInSFC"},
			},
		},
	})
}

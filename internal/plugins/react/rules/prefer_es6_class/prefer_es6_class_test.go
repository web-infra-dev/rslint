package prefer_es6_class

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferEs6ClassRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferEs6ClassRule, []rule_tester.ValidTestCase{
		// ---- Upstream: default (always) — ES6 class is fine ----
		{Code: `
        class Hello extends React.Component {
          render() {
            return <div>Hello {this.props.name}</div>;
          }
        }
        Hello.displayName = 'Hello'
      `, Tsx: true},

		// ---- Upstream: default (always) — export default ES6 class is fine ----
		{Code: `
        export default class Hello extends React.Component {
          render() {
            return <div>Hello {this.props.name}</div>;
          }
        }
        Hello.displayName = 'Hello'
      `, Tsx: true},

		// ---- Upstream: no component at all ----
		{Code: `
        var Hello = "foo";
        module.exports = {};
      `, Tsx: true},

		// ---- Upstream: createReactClass is fine when mode is "never" ----
		{
			Code: `
        var Hello = createReactClass({
          render: function() {
            return <div>Hello {this.props.name}</div>;
          }
        });
      `,
			Tsx:     true,
			Options: []any{"never"},
		},

		// ---- Upstream: ES6 class is fine when mode is "always" (explicit) ----
		{
			Code: `
        class Hello extends React.Component {
          render() {
            return <div>Hello {this.props.name}</div>;
          }
        }
      `,
			Tsx:     true,
			Options: []any{"always"},
		},

		// ---- Edge: mode "never" + non-component class (no extends) — not reported ----
		{
			Code: `
        class Hello {
          render() {
            return <div>Hello</div>;
          }
        }
      `,
			Tsx:     true,
			Options: []any{"never"},
		},

		// ---- Edge: mode "never" + class extending something other than
		// React.Component / React.PureComponent — not reported ----
		{
			Code: `
        class Hello extends SomethingElse {
          render() {
            return <div>Hello</div>;
          }
        }
      `,
			Tsx:     true,
			Options: []any{"never"},
		},

		// ---- Edge: mode "never" + ClassExpression extending React.Component —
		// upstream only subscribes to ClassDeclaration, so ClassExpressions must
		// NOT be reported. Locks the upstream behavior. ----
		{
			Code: `
        const Hello = class extends React.Component {
          render() {
            return <div>Hello</div>;
          }
        };
      `,
			Tsx:     true,
			Options: []any{"never"},
		},
		{
			Code: `
        const Hello = class extends React.PureComponent {
          render() {
            return <div>Hello</div>;
          }
        };
      `,
			Tsx:     true,
			Options: []any{"never"},
		},

		// ---- Edge: mode "always" + ObjectExpression passed to a non-createReactClass
		// callee — not reported. ----
		{Code: `
        var x = something({ foo: 1 });
      `, Tsx: true},

		// ---- Edge: mode "always" + pragma-mismatched member-expression callee
		// `Other.createClass({...})` — not reported when pragma is default
		// ("React"). ----
		{Code: `
        var Hello = Other.createClass({
          render: function() {
            return <div/>;
          }
        });
      `, Tsx: true},

		// ---- Edge: mode "always" + pragma-mismatched member-expression method
		// `React.somethingElse({...})` — not reported (property name mismatch). ----
		{Code: `
        var x = React.somethingElse({ foo: 1 });
      `, Tsx: true},

		// ---- Edge: mode "always" + object literal that is NOT a direct argument
		// of a CallExpression (nested inside another object) — not reported. ----
		{Code: `
        var x = {
          inner: { foo: 1 }
        };
      `, Tsx: true},

		// ---- Edge: mode "always" + object literal as value of a PropertyAssignment
		// inside createReactClass — only the OUTER object (the createReactClass
		// arg) should report, not the INNER one. Here we lock the inner-object
		// non-report by putting it on the valid side with mode "never". ----
		{
			Code: `
        var Hello = createReactClass({
          defaultProps: { foo: 1 },
          render: function() {
            return <div/>;
          }
        });
      `,
			Tsx:     true,
			Options: []any{"never"},
		},

		// ---- Edge: mode "always" + IIFE `({foo:1})()` — callee is the object
		// itself, not a createClass identifier. Not reported. ----
		{Code: `
        var x = ({ foo: 1 })();
      `, Tsx: true},

		// ---- Edge: custom `settings.react.createClass` — createFoo({...})
		// NOT reported with default pragma settings. ----
		{Code: `
        var Hello = createFoo({
          render: function() { return <div/>; }
        });
      `, Tsx: true},

		// ---- Edge: custom `settings.react.createClass` — createReactClass
		// is NOT a matching callee when settings points elsewhere. ----
		{
			Code: `
        var Hello = createReactClass({
          render: function() { return <div/>; }
        });
      `,
			Tsx: true,
			Settings: map[string]interface{}{
				"react": map[string]interface{}{
					"createClass": "createFoo",
				},
			},
		},

		// ---- Edge: custom `settings.react.pragma` — `React.createClass`
		// NOT matching once pragma is overridden. ----
		{
			Code: `
        var Hello = React.createClass({
          render: function() { return <div/>; }
        });
      `,
			Tsx: true,
			Settings: map[string]interface{}{
				"react": map[string]interface{}{
					"pragma": "Foo",
				},
			},
		},

		// ---- Edge: mode "always" + default settings + `React.createClass({...})`
		// — the default `createClass` name is `createReactClass`, NOT
		// `createClass`, so this is NOT flagged. Locks the upstream default. ----
		{Code: `
        var Hello = React.createClass({
          render: function() { return <div/>; }
        });
      `, Tsx: true},

		// ---- Edge: computed-property access `React['createReactClass']({...})`
		// — upstream reads `callee.property.name` which is undefined for a
		// computed/string-literal property. Not reported. ----
		{Code: `
        var Hello = React['createReactClass']({
          render: function() { return <div/>; }
        });
      `, Tsx: true},

		// ---- Edge: mode "always" + tagged template ``createReactClass` ``` —
		// TaggedTemplateExpression has no `.callee` in ESTree; upstream skips.
		// No object literal here so the listener is irrelevant, but lock that
		// no false positive fires from the surrounding ObjectLiteral either. ----
		{Code: `
        var Hello = createReactClass` + "`ignored`" + `;
        var Other = foo({ nested: { inner: 1 } });
      `, Tsx: true},

		// ---- Edge: `class X extends A.B.C {}` (C happens to be "Component")
		// — upstream's superClass check only matches a 2-segment
		// `<pragma>.Component` member access. Deeper chains don't match. ----
		{
			Code: `
        class Hello extends A.B.Component {
          render() { return <div/>; }
        }
      `,
			Tsx:     true,
			Options: []any{"never"},
		},

		// ---- Edge: `class X extends (React.Component as any) {}` —
		// TS type-assertion wraps the superClass as `TSAsExpression` in ESTree
		// (AsExpression in tsgo); upstream's superClass type check matches
		// neither Identifier nor MemberExpression. Not reported. ----
		{
			Code: `
        class Hello extends (React.Component as any) {
          render() { return null; }
        }
      `,
			Tsx:     true,
			Options: []any{"never"},
		},

		// ---- Edge: `class X extends React.Component.foo {}` — rightmost
		// property name is not `Component` / `PureComponent`; upstream's
		// `/^(Pure)?Component$/.test(property.name)` fails. Not reported. ----
		{
			Code: `
        class Hello extends React.Component.foo {
          render() { return null; }
        }
      `,
			Tsx:     true,
			Options: []any{"never"},
		},

		// ---- Edge: `class X extends Other.Component {}` — object name does
		// NOT match the default pragma ("React"); not reported. ----
		{
			Code: `
        class Hello extends Other.Component {
          render() { return null; }
        }
      `,
			Tsx:     true,
			Options: []any{"never"},
		},

		// ---- Edge: `class X extends fn() {}` — superClass is a
		// CallExpression; matches neither Identifier nor MemberExpression.
		// Not reported. ----
		{
			Code: `
        class Hello extends getBase() {
          render() { return null; }
        }
      `,
			Tsx:     true,
			Options: []any{"never"},
		},

		// ---- Edge: `class X extends React.Component<P, {}> {}` under mode
		// "always" — the ClassDeclaration listener never fires in "always"
		// mode. Not reported. ----
		{Code: `
        class Hello<P> extends React.Component<P, {}> {
          render() { return null; }
        }
      `, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// ---- Upstream: default (always) — createReactClass should be flagged ----
		{
			Code: `
        var Hello = createReactClass({
          displayName: 'Hello',
          render: function() {
            return <div>Hello {this.props.name}</div>;
          }
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "shouldUseES6Class", Message: "Component should use es6 class instead of createClass", Line: 2, Column: 38},
			},
		},

		// ---- Upstream: explicit mode "always" — createReactClass should be flagged ----
		{
			Code: `
        var Hello = createReactClass({
          render: function() {
            return <div>Hello {this.props.name}</div>;
          }
        });
      `,
			Tsx:     true,
			Options: []any{"always"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "shouldUseES6Class", Line: 2, Column: 38},
			},
		},

		// ---- Upstream: mode "never" — ES6 class extending React.Component flagged ----
		{
			Code: `
        class Hello extends React.Component {
          render() {
            return <div>Hello {this.props.name}</div>;
          }
        }
      `,
			Tsx:     true,
			Options: []any{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "shouldUseCreateClass", Message: "Component should use createClass instead of es6 class", Line: 2, Column: 9},
			},
		},

		// ---- Edge: mode "always" + pragma-qualified
		// `React.createReactClass({...})`. Default pragma is "React" and
		// default createClass name is "createReactClass", so this is the
		// literal member-expression form the rule recognizes out of the box. ----
		{
			Code: `
        var Hello = React.createReactClass({
          render: function() { return <div/>; }
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "shouldUseES6Class", Line: 2, Column: 44},
			},
		},

		// ---- Edge: mode "always" + paren-wrapped object `createReactClass(({...}))`
		// — tsgo preserves parens; ESTree flattens them. Must still report. ----
		{
			Code: `
        var Hello = createReactClass(({
          render: function() { return <div/>; }
        }));
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "shouldUseES6Class", Line: 2, Column: 39},
			},
		},

		// ---- Edge: mode "always" + paren-wrapped callee `(createReactClass)({...})` ----
		{
			Code: `
        var Hello = (createReactClass)({
          render: function() { return <div/>; }
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "shouldUseES6Class", Line: 2, Column: 40},
			},
		},

		// ---- Edge: mode "never" + bare `Component` (no pragma) ----
		{
			Code: `
        class Hello extends Component {
          render() {
            return <div/>;
          }
        }
      `,
			Tsx:     true,
			Options: []any{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "shouldUseCreateClass", Line: 2, Column: 9},
			},
		},

		// ---- Edge: mode "never" + bare `PureComponent` (no pragma) ----
		{
			Code: `
        class Hello extends PureComponent {
          render() {
            return <div/>;
          }
        }
      `,
			Tsx:     true,
			Options: []any{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "shouldUseCreateClass", Line: 2, Column: 9},
			},
		},

		// ---- Edge: mode "never" + pragma-qualified `React.PureComponent` ----
		{
			Code: `
        class Hello extends React.PureComponent {
          render() {
            return <div/>;
          }
        }
      `,
			Tsx:     true,
			Options: []any{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "shouldUseCreateClass", Line: 2, Column: 9},
			},
		},

		// ---- Edge: mode "never" + `export class X ...` — report starts at
		// the `class` keyword, past the `export` modifier. ----
		{
			Code: `
        export class Hello extends React.Component {
          render() {
            return <div/>;
          }
        }
      `,
			Tsx:     true,
			Options: []any{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "shouldUseCreateClass", Line: 2, Column: 16},
			},
		},

		// ---- Edge: mode "never" + `export default class ...` — still a
		// ClassDeclaration, so it fires. The report starts at the `class`
		// keyword, mirroring ESLint (whose ESTree ClassDeclaration begins at
		// `class`; tsgo inlines the `export default` modifiers into the
		// ClassDeclaration, so we skip past them before reporting). ----
		{
			Code: `
        export default class Hello extends React.Component {
          render() {
            return <div/>;
          }
        }
      `,
			Tsx:     true,
			Options: []any{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "shouldUseCreateClass", Line: 2, Column: 24, EndLine: 6, EndColumn: 10},
			},
		},

		// ---- Edge: custom `settings.react.pragma` — pragma-qualified
		// `Foo.createReactClass({...})` reports when pragma is set to "Foo". ----
		{
			Code: `
        var Hello = Foo.createReactClass({
          render: function() { return <div/>; }
        });
      `,
			Tsx: true,
			Settings: map[string]interface{}{
				"react": map[string]interface{}{
					"pragma": "Foo",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "shouldUseES6Class", Line: 2, Column: 42},
			},
		},

		// ---- Edge: custom `settings.react.createClass` — bare `createFoo({...})`
		// reports when createClass is set to "createFoo". ----
		{
			Code: `
        var Hello = createFoo({
          render: function() { return <div/>; }
        });
      `,
			Tsx: true,
			Settings: map[string]interface{}{
				"react": map[string]interface{}{
					"createClass": "createFoo",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "shouldUseES6Class", Line: 2, Column: 31},
			},
		},

		// ---- Edge: custom pragma + custom createClass — `Foo.createFoo({...})`
		// reports when both settings are overridden. ----
		{
			Code: `
        var Hello = Foo.createFoo({
          render: function() { return <div/>; }
        });
      `,
			Tsx: true,
			Settings: map[string]interface{}{
				"react": map[string]interface{}{
					"pragma":      "Foo",
					"createClass": "createFoo",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "shouldUseES6Class", Line: 2, Column: 35},
			},
		},

		// ---- Edge: `new createReactClass({...})` — NewExpression also has
		// `.callee` in ESTree so upstream reports. Locks the alignment. ----
		{
			Code: `
        var Hello = new createReactClass({
          render: function() { return <div/>; }
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "shouldUseES6Class", Line: 2, Column: 42},
			},
		},

		// ---- Edge: `new React.createReactClass({...})` — pragma-qualified
		// NewExpression. Upstream reports. ----
		{
			Code: `
        var Hello = new React.createReactClass({
          render: function() { return <div/>; }
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "shouldUseES6Class", Line: 2, Column: 48},
			},
		},

		// ---- Edge: mode "always" + second-argument object literal —
		// upstream's `node.parent.callee` check fires regardless of argument
		// position, so a createReactClass call with multiple object-literal
		// arguments reports on each. Also locks in that a non-object first
		// argument doesn't gate later object-literal arguments. ----
		{
			Code: `
        var Hello = createReactClass(mixin, {
          render: function() { return <div/>; }
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "shouldUseES6Class", Line: 2, Column: 45},
			},
		},

		// ---- Edge: mode "always" + nested object-literal as property value —
		// upstream reports ONLY on the outer (direct-argument) object, not
		// on the inner property-value object. Lock with errors length = 1. ----
		{
			Code: `
        var Hello = createReactClass({
          defaultProps: { foo: 1 },
          render: function() { return <div/>; }
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "shouldUseES6Class", Line: 2, Column: 38},
			},
		},

		// ---- Edge: mode "never" + custom pragma on ES6 class — pragma-qualified
		// `Foo.Component` fires when `settings.react.pragma: "Foo"`. ----
		{
			Code: `
        class Hello extends Foo.Component {
          render() { return null; }
        }
      `,
			Tsx:     true,
			Options: []any{"never"},
			Settings: map[string]interface{}{
				"react": map[string]interface{}{
					"pragma": "Foo",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "shouldUseCreateClass", Line: 2, Column: 9},
			},
		},

		// ---- Edge: mode "never" + custom pragma + PureComponent ----
		{
			Code: `
        class Hello extends Foo.PureComponent {
          render() { return null; }
        }
      `,
			Tsx:     true,
			Options: []any{"never"},
			Settings: map[string]interface{}{
				"react": map[string]interface{}{
					"pragma": "Foo",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "shouldUseCreateClass", Line: 2, Column: 9},
			},
		},

		// ---- Edge: mode "never" + class with type parameters on heritage —
		// `class X<P> extends React.Component<P, S>` still fires; type-args
		// wrap via ExpressionWithTypeArguments and ExtendsReactComponent
		// unwraps to the underlying PropertyAccessExpression. ----
		{
			Code: `
        class Hello<P> extends React.Component<P, {}> {
          render() { return null; }
        }
      `,
			Tsx:     true,
			Options: []any{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "shouldUseCreateClass", Line: 2, Column: 9},
			},
		},

		// ---- Edge: mode "never" + two nested ClassDeclarations both
		// extending React.Component — each ClassDeclaration is visited
		// independently, so both fire. ----
		{
			Code: `
        class Outer extends React.Component {
          render() {
            class Inner extends React.Component {
              render() { return null; }
            }
            return null;
          }
        }
      `,
			Tsx:     true,
			Options: []any{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "shouldUseCreateClass", Line: 2, Column: 9},
				{MessageId: "shouldUseCreateClass", Line: 4, Column: 13},
			},
		},

		// ---- Edge: mode "never" + `abstract class X extends React.Component`
		// — `abstract` is a TS modifier kept on the TSESTree ClassDeclaration
		// (unlike `export`, which lives on an outer wrapper), so ESLint
		// reports starting at `abstract`, not at `class`. Lock alignment. ----
		{
			Code: `
        abstract class Hello extends React.Component {
          render() { return null; }
        }
      `,
			Tsx:     true,
			Options: []any{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "shouldUseCreateClass", Line: 2, Column: 9},
			},
		},

		// ---- Edge: mode "never" + `declare class X extends React.Component`
		// — `declare` also stays on the ClassDeclaration in TSESTree. Report
		// starts at `declare`. ----
		{
			Code: `
        declare class Hello extends React.Component {
          render(): any;
        }
      `,
			Tsx:     true,
			Options: []any{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "shouldUseCreateClass", Line: 2, Column: 9},
			},
		},

		// ---- Edge: mode "never" + decorator (`@foo class X ...`) — decorators
		// are part of the ClassDeclaration range in TSESTree. Report starts
		// at the `@`. ----
		{
			Code: `
@foo
class Hello extends React.Component {
  render() { return null; }
}
      `,
			Tsx:     true,
			Options: []any{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "shouldUseCreateClass", Line: 2, Column: 1},
			},
		},

		// ---- Edge: mode "never" + `export default abstract class X ...` —
		// `export default` are stripped; `abstract` is kept. Report starts
		// at `abstract`. ----
		{
			Code: `
        export default abstract class Hello extends React.Component {
          render() { return null; }
        }
      `,
			Tsx:     true,
			Options: []any{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "shouldUseCreateClass", Line: 2, Column: 24},
			},
		},

		// ---- Edge: mode "never" + `export default class` with NO name — the
		// anonymous form is a valid ClassDeclaration in tsgo (the rspack/rsbuild
		// patterns also look like this in practice). Report still at `class`. ----
		{
			Code: `
        export default class extends React.Component {
          render() { return null; }
        }
      `,
			Tsx:     true,
			Options: []any{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "shouldUseCreateClass", Line: 2, Column: 24},
			},
		},

		// ---- Edge: mode "always" + paren-wrapped object inside `new` —
		// `new createReactClass(({...}))` combines NewExpression parent
		// detection with paren unwrapping. ----
		{
			Code: `
        var Hello = new createReactClass(({
          render: function() { return <div/>; }
        }));
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "shouldUseES6Class", Line: 2, Column: 43},
			},
		},
	})
}

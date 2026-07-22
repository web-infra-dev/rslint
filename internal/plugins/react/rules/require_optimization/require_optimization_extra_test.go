package require_optimization

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/tspath"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// TestRequireOptimizationRuleExtra collects rslint-specific edge cases that
// extend coverage beyond the upstream eslint-plugin-react test suite (kept
// intact in `require_optimization_test.go`):
//
//   - tsgo AST quirks (parens, optional chain, generic type args, satisfies,
//     ClassExpression-vs-ClassDeclaration asymmetries, …)
//   - Walk-up boundary lock-ins for the `mark-SCU-as-declared` semantics
//     (nested detected components absorb the SCU; outer reports
//     independently)
//   - SFC detection branches (in-class-method body capitalized fns; class
//     field arrow vs class method arrow; PropertyAssignment value arrow;
//     anonymous nested arrows; class static block; …)
//   - TS-specific syntax (abstract method gate, declare class, namespace /
//     module, generic SCU, overload signatures, access modifiers)
//   - Settings overrides (`react.pragma`, `react.createClass`)
//   - Decorator boundary cases (deep member access, optional chain, spread
//     args, factory call, paren-wrapped Identifier, multi-decorator chains)
//   - Robustness (malformed options, null entries, options shape variations)
//   - Real-codebase regressions (rspack class component, etc.)
//
// Each case here was verified against eslint-plugin-react@latest via
// differential validation before being locked in.
func TestRequireOptimizationRuleExtra(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &RequireOptimizationRule, []rule_tester.ValidTestCase{
		// ---- Edge: ClassExpression assigned to var — same React.Component check applies ----
		{Code: `
        const YourComponent = class extends React.Component {
          shouldComponentUpdate() { return true; }
        };
      `, Tsx: true},

		// ---- Edge: PrivateIdentifier #shouldComponentUpdate — upstream strips '#' on .name ----
		{Code: `
        class YourComponent extends React.Component {
          #shouldComponentUpdate() { return true; }
        }
      `, Tsx: true},

		// ---- Edge: getter shouldComponentUpdate — MethodDefinition listener fires on accessors ----
		{Code: `
        class YourComponent extends React.Component {
          get shouldComponentUpdate() { return () => true; }
        }
      `, Tsx: true},

		// ---- Edge: setter shouldComponentUpdate — same ----
		{Code: `
        class YourComponent extends React.Component {
          set shouldComponentUpdate(v) {}
        }
      `, Tsx: true},

		// ---- Edge: static shouldComponentUpdate — upstream doesn't filter static, neither do we ----
		{Code: `
        class YourComponent extends React.Component {
          static shouldComponentUpdate() {}
        }
      `, Tsx: true},

		// ---- Edge: paren-wrapped extends — paren-transparent (matches ESTree) ----
		{Code: `
        class YourComponent extends (React.Component) {
          shouldComponentUpdate() {}
        }
      `, Tsx: true},

		// ---- Edge: parens around decorator receiver — `@(reactMixin).decorate(PureRenderMixin)` ----
		{Code: `
        @(reactMixin).decorate(PureRenderMixin)
        class DecoratedComponent extends Component {}
      `, Tsx: true},

		// ---- Edge: parens around PureRenderMixin argument — `[(PureRenderMixin)]` in mixins ----
		{Code: `
        createReactClass({
          mixins: [(PureRenderMixin)]
        })
      `, Tsx: true},

		// ---- Edge: createReactClass with ParenthesizedExpression around obj arg ----
		{Code: `
        createReactClass(({
          mixins: [PureRenderMixin]
        }))
      `, Tsx: true},

		// ---- Edge: createReactClass with shorthand `shouldComponentUpdate()` method ----
		{Code: `
        createReactClass({
          shouldComponentUpdate() {}
        })
      `, Tsx: true},

		// ---- Edge: React.createClass (pragma + createClass) qualified call ----
		{Code: `
        React.createClass({
          shouldComponentUpdate: function () {}
        })
      `, Tsx: true},

		// ---- Edge: createReactClass with mixins containing PureRenderMixin among others ----
		{Code: `
        createReactClass({
          mixins: [SomeMixin, PureRenderMixin, AnotherMixin]
        })
      `, Tsx: true},

		// ---- Edge: ObjectLiteral used outside createReactClass — never reported (not detected) ----
		{Code: `
        const config = {};
      `, Tsx: true},

		// ---- Edge: ObjectLiteral with shouldComponentUpdate in non-createReactClass position — never reported ----
		{Code: `
        const obj = {
          // not a createReactClass arg, irrelevant whether SCU is here
        };
      `, Tsx: true},

		// ---- Edge: empty `class X extends React.PureComponent {}` is allowed (PureComponent-only path) ----
		{Code: `
        class YourComponent extends React.PureComponent {}
      `, Tsx: true},

		// ---- Edge: anonymous default-exported ClassDeclaration with
		// PureComponent. tsgo classifies `export default class …` as a
		// ClassDeclaration (with export/default modifiers) so the
		// PureComponent shortcut applies. Verified empirically against
		// eslint-plugin-react. ----
		{Code: `
        export default class extends React.PureComponent {}
      `, Tsx: true},

		// ---- Edge: nested class — only the React-extending class is checked ----
		{Code: `
        class Outer {
          method() {
            class Inner extends React.Component {
              shouldComponentUpdate() {}
            }
          }
        }
      `, Tsx: true},

		// ---- Edge: settings.react.pragma="Preact" + Preact.PureComponent ----
		{Code: `
        class YourComponent extends Preact.PureComponent {}
      `, Tsx: true, Settings: map[string]interface{}{"react": map[string]interface{}{"pragma": "Preact"}}},

		// ---- Edge: settings.react.createClass="myCreateClass" custom factory ----
		{Code: `
        myCreateClass({
          shouldComponentUpdate: function() {}
        })
      `, Tsx: true, Settings: map[string]interface{}{"react": map[string]interface{}{"createClass": "myCreateClass"}}},

		// Locks in upstream `hasCustomDecorator` arm: empty allowDecorators must
		// not match anything, so the rule reports plain-decorator classes.
		// (This valid case proves the inverse — when allow contains the name, it passes.)
		{Code: `
        @custom
        class DecoratedComponent extends Component {}
      `, Tsx: true, Options: map[string]interface{}{"allowDecorators": []interface{}{"custom"}}},

		// ---- Branch: allowDecorators item is the EXACT identifier name —
		// CallExpression decorator like `@custom()` does NOT match (upstream's
		// `expression.name` is undefined on CallExpression). Locked in here. ----
		{Code: `
        @custom
        @customCall
        class DecoratedComponent extends Component {}
      `, Tsx: true, Options: map[string]interface{}{"allowDecorators": []interface{}{"customCall"}}},

		// ---- Branch: ExtendsReactComponent matches `extends Component` from any source —
		// classes extending PureComponent take the PureComponent shortcut. ----
		{Code: `
        class A extends PureComponent {}
      `, Tsx: true},

		// ---- Edge (TS): generic type args on PureComponent extends — type
		// args are stripped via ExpressionWithTypeArguments wrapper. ----
		{Code: `
        class YourComponent extends React.PureComponent<Props, State> {}
      `, Tsx: true},

		// ---- Edge (TS): bare PureComponent with type args ----
		{Code: `
        class YourComponent extends PureComponent<Props> {
          render() { return null; }
        }
      `, Tsx: true},

		// ---- Edge (TS): abstract React.Component subclass with SCU ----
		{Code: `
        abstract class YourComponent extends React.Component {
          shouldComponentUpdate() {}
        }
      `, Tsx: true},

		// ---- Edge (TS): TypeAssertion `(React.Component as any)` — text-regex
		// would not match either, so we mirror that miss; the class is not
		// detected as a React component and never reported. ----
		{Code: `
        class YourComponent extends (React.Component as any) {
          // no SCU — NOT detected as a React component, so NOT reported
        }
      `, Tsx: true},

		// ---- Edge: sibling components in same file — each evaluated independently ----
		{Code: `
        class A extends React.Component {
          shouldComponentUpdate() {}
        }
        class B extends React.PureComponent {}
      `, Tsx: true},

		// ---- Edge: doubly-nested class — only inner extends React.Component, with SCU ----
		{Code: `
        class Outer {
          method() {
            class Middle {
              another() {
                class Inner extends React.Component {
                  shouldComponentUpdate() { return true; }
                }
              }
            }
          }
        }
      `, Tsx: true},

		// ---- Edge: the SCU method is the LAST declared sibling — iteration
		// must not stop early on a non-matching member. ----
		{Code: `
        class YourComponent extends React.Component {
          a() {}
          b() {}
          render() { return null; }
          shouldComponentUpdate() {}
        }
      `, Tsx: true},

		// ---- Real user: `export default class extends React.PureComponent {}`
		// (anonymous default-exported class). PureComponent shortcut applies. ----
		{Code: `
        export default class extends React.PureComponent {}
      `, Tsx: true},

		// ---- Real user: `export default class C extends React.Component { ... }`
		// with shouldComponentUpdate. ----
		{Code: `
        export default class C extends React.Component {
          shouldComponentUpdate() { return true; }
        }
      `, Tsx: true},

		// ---- Real user: `export class C extends React.Component { ... }` named
		// export with shouldComponentUpdate. ----
		{Code: `
        export class C extends React.Component {
          shouldComponentUpdate() { return true; }
        }
      `, Tsx: true},

		// ---- Real user: TS access modifiers on shouldComponentUpdate
		// (`public` / `protected` / `private`) — modifier list does not
		// change member name detection. ----
		{Code: `
        class C extends React.Component {
          public shouldComponentUpdate() { return true; }
        }
      `, Tsx: true},
		{Code: `
        class C extends React.Component {
          protected shouldComponentUpdate() { return true; }
        }
      `, Tsx: true},
		{Code: `
        class C extends React.Component {
          private shouldComponentUpdate() { return true; }
        }
      `, Tsx: true},

		// ---- Real user: TS `override` modifier on SCU (TS 4.3+) ----
		{Code: `
        class C extends React.Component {
          override shouldComponentUpdate() { return true; }
        }
      `, Tsx: true},

		// ---- Real user: SCU with parameter list (props, state) — body shape
		// is irrelevant, only the name matters. ----
		{Code: `
        class C extends React.Component<P, S> {
          shouldComponentUpdate(nextProps: P, nextState: S): boolean {
            return true;
          }
        }
      `, Tsx: true},

		// ---- Real user: namespace-imported React (`import * as React from "react"`) ----
		{Code: `
        import * as React from "react";
        class C extends React.PureComponent {}
      `, Tsx: true},

		// ---- Real user: PureRender decorator immediately above class with no
		// blank line — whitespace is irrelevant. ----
		{Code: `@reactMixin.decorate(PureRenderMixin)
class C extends Component {}`, Tsx: true},

		// ---- Real user: redux/mobx-style decorator factory with @observer
		// allow-listed via call form — does NOT match (Identifier-only),
		// so must extend PureComponent or have SCU. We test the bare-Identifier
		// allow-list path here. ----
		{Code: `
        @observer
        class Store extends React.Component {
          shouldComponentUpdate() { return true; }
        }
      `, Tsx: true},

		// ---- Real user: connect HOC wrapping a class — the inner class still
		// extends React.Component so SCU is required. SCU present → valid. ----
		{Code: `
        class Inner extends React.Component {
          shouldComponentUpdate() { return true; }
        }
        const Connected = connect(mapStateToProps)(Inner);
      `, Tsx: true},

		// ---- Real user: TS declaration merging — class + namespace ----
		{Code: `
        class C extends React.Component {
          shouldComponentUpdate() { return true; }
        }
        namespace C {}
      `, Tsx: true},

		// ---- Real user: SCU defined on a class extending an aliased React
		// (`import { Component as Comp } from "react"`) — upstream's regex
		// ONLY matches "Component"/"PureComponent" via text, so an aliased
		// `Comp` would not be detected as a React component → not reported.
		// We mirror that miss. ----
		{Code: `
        import { Component as Comp } from "react";
        class C extends Comp {}
      `, Tsx: true},

		// ---- Real user: settings-driven createClass + custom pragma ----
		{Code: `
        Preact.createClass({ shouldComponentUpdate() {} })
      `, Tsx: true, Settings: map[string]interface{}{"react": map[string]interface{}{"pragma": "Preact"}}},

		// ---- Edge: createReactClass with mixins value being a CONST array
		// reference — value isn't ArrayLiteral inline, gate fails. The
		// createReactClass call still gets reported (counterexample below);
		// here we add the mirror VALID — same call but inline mixins. ----
		{Code: `
        createReactClass({
          mixins: [PureRenderMixin],
          render() { return null; }
        })
      `, Tsx: true},

		// ---- Edge: SCU defined as the very first member ----
		{Code: `
        class C extends React.Component {
          shouldComponentUpdate() {}
          render() { return null; }
        }
      `, Tsx: true},

		// ---- Edge: SCU with TS overload signatures + implementation —
		// any of the Identifier-keyed signatures matches. ----
		{Code: `
        class C extends React.Component {
          shouldComponentUpdate(): boolean;
          shouldComponentUpdate(nextProps: any): boolean;
          shouldComponentUpdate() { return true; }
        }
      `, Tsx: true},

		// ===== Walk-up semantics (verified empirically against upstream) =====
		// Upstream's `mark-SCU-as-declared(node)` calls `components.set(node, ...)`
		// which walks up the parent chain to the nearest detected component
		// and stamps `hasSCU=true`. So a SCU/PureRenderMixin signal
		// ANYWHERE inside a component's subtree silences the report —
		// even nested arbitrarily deep, even across non-React class /
		// function boundaries.

		// ---- walk-up: ObjectExpression with SCU inside class method body ----
		{Code: `
        class C extends React.Component {
          init() {
            const cfg = { shouldComponentUpdate() {} };
          }
        }
      `, Tsx: true},

		// ---- walk-up: ObjectExpression with SCU inside class field initializer ----
		{Code: `
        class C extends React.Component {
          static defaults = {
            shouldComponentUpdate: function() {}
          };
          render() { return null; }
        }
      `, Tsx: true},

		// ---- walk-up: ObjectExpression with SCU inside arrow class field ----
		{Code: `
        class C extends React.Component {
          build = () => ({ shouldComponentUpdate() {} });
        }
      `, Tsx: true},

		// ---- walk-up: arrow returning object with SCU inside class method ----
		{Code: `
        class C extends React.Component {
          build() {
            return () => ({ shouldComponentUpdate() {} });
          }
        }
      `, Tsx: true},

		// ---- walk-up: PureRender decorator on non-React nested class
		// silences the OUTER React component (the inner class isn't a
		// detected component, so its markSCU walks up). ----
		{Code: `
        class C extends React.Component {
          build() {
            @reactMixin.decorate(PureRenderMixin)
            class Helper extends NonReact {}
          }
        }
      `, Tsx: true},

		// ---- walk-up: PureRender decorator on inner class without extends ----
		{Code: `
        class C extends React.Component {
          build() {
            @reactMixin.decorate(PureRenderMixin)
            class Helper {}
          }
        }
      `, Tsx: true},

		// ---- walk-up: SCU method on nested non-React class silences outer ----
		{Code: `
        class C extends React.Component {
          build() {
            class Helper {
              shouldComponentUpdate() {}
            }
          }
        }
      `, Tsx: true},

		// ---- walk-up: SCU obj inside helper function body silences outer ----
		{Code: `
        class C extends React.Component {
          build() {
            function helper() {
              return { shouldComponentUpdate() {} };
            }
          }
        }
      `, Tsx: true},

		// ---- walk-up: nested ObjectExpression containing SCU silences outer ----
		{Code: `
        createReactClass({
          config: {
            shouldComponentUpdate: function() {}
          }
        })
      `, Tsx: true},

		// ---- walk-up: nested mixins silences outer createReactClass ----
		{Code: `
        createReactClass({
          config: {
            mixins: [PureRenderMixin]
          }
        })
      `, Tsx: true},

		// ===== Stateless functional component (SFC) =====
		// Top-level / module-level SFC function/arrow are auto-silenced by
		// upstream's `mark-SCU-as-declared(self)` (only `isFunctionInClass`
		// negative path); rslint achieves the same by simply not
		// classifying them as detected components.

		// ---- top-level FunctionDeclaration with capitalized name ----
		{Code: `
        function Comp(props) { return <div />; }
      `, Tsx: true},

		// ---- top-level arrow assigned to const ----
		{Code: `
        const Comp = (props) => <div />;
      `, Tsx: true},

		// ---- top-level FunctionExpression assigned to const ----
		{Code: `
        const Comp = function(props) { return <div />; };
      `, Tsx: true},

		// ---- top-level arrow returning null — markSCU(self) silences ----
		{Code: `
        const Comp = () => null;
      `, Tsx: true},

		// ---- top-level function returning non-JSX — markSCU(self) silences ----
		{Code: `
        function Comp() { return 1; }
      `, Tsx: true},

		// ---- lowercase-named function — never classified as SFC ----
		{Code: `
        function comp(props) { return <div />; }
      `, Tsx: true},

		// ---- nested top-level SFCs both auto-silenced ----
		{Code: `
        function Outer(props) {
          function Inner(p) { return <div />; }
          return <Inner />;
        }
      `, Tsx: true},

		// ---- SFC + class component, class with SCU → both silenced ----
		{Code: `
        function Comp(props) { return <div />; }
        class C extends React.Component { shouldComponentUpdate() {} }
      `, Tsx: true},

		// ---- export default function — top-level, silenced ----
		{Code: `
        export default function Comp(props) { return <div />; }
      `, Tsx: true},

		// ---- export default anonymous function — top-level, silenced ----
		{Code: `
        export default function(props) { return <div />; }
      `, Tsx: true},

		// ---- SFC inside ClassExpression's method is auto-silenced
		// (upstream `isFunctionInClass` only matches ClassDeclaration).
		// Outer ClassExpression silenced via `shouldComponentUpdate`. ----
		{Code: `
        const X = class extends React.Component {
          shouldComponentUpdate() {}
          build() {
            function Inner(p) { return <div />; }
            return <Inner />;
          }
        };
      `, Tsx: true},

		// ---- class field arrow — even capitalized + returns JSX is NOT
		// detected as SFC (Components.detect skips class-field initializer). ----
		{Code: `
        class C extends React.Component {
          shouldComponentUpdate() {}
          Inner = (p) => <div />;
          render() { return <this.Inner />; }
        }
      `, Tsx: true},

		// ---- in-class-method SFC with SCU obj inside it absorbs SCU
		// to itself; outer class still reports separately (in invalid). ----
		{Code: `
        class C extends React.Component {
          shouldComponentUpdate() {}
          build() {
            function Inner() {
              const cfg = { shouldComponentUpdate() {} };
              return <div />;
            }
          }
        }
      `, Tsx: true},

		// ===== Settings overrides =====

		// ---- pragma="Preact" + Preact.PureComponent — silences ----
		{Code: `
        class C extends Preact.PureComponent {}
      `, Tsx: true, Settings: map[string]interface{}{"react": map[string]interface{}{"pragma": "Preact"}}},

		// ---- pragma="Preact" + Preact.createClass({...}) — recognized ----
		{Code: `
        Preact.createClass({});
      `, Tsx: true, Settings: map[string]interface{}{"react": map[string]interface{}{"pragma": "Preact"}}},

		// ---- pragma="Preact" + React.Component — NOT recognized as React
		// component anymore; not reported. ----
		{Code: `
        class C extends React.Component {}
      `, Tsx: true, Settings: map[string]interface{}{"react": map[string]interface{}{"pragma": "Preact"}}},

		// ---- createClass="myCreateClass" + custom factory + SCU ----
		{Code: `
        myCreateClass({ shouldComponentUpdate() {} });
      `, Tsx: true, Settings: map[string]interface{}{"react": map[string]interface{}{"createClass": "myCreateClass"}}},

		// ---- createClass="myCreateClass" + default name `createReactClass`
		// is no longer recognized → not detected → not reported. ----
		{Code: `
        createReactClass({});
      `, Tsx: true, Settings: map[string]interface{}{"react": map[string]interface{}{"createClass": "myCreateClass"}}},

		// ===== Misc verified-against-upstream edge cases =====

		// ---- comma expression in extends — not a React component shape ----
		{Code: `
        class C extends (React, Component) {}
      `, Tsx: true},

		// ---- HOC mixin call in extends — not a React component shape ----
		{Code: `
        class C extends Mixin(Component) {}
      `, Tsx: true},

		// ---- multi-level member access in extends — regex anchors at start ----
		{Code: `
        class C extends MyLib.React.Component {}
      `, Tsx: true},

		// ---- extends a CallExpression — not a component ----
		{Code: `
        class C extends getBase() {
          shouldComponentUpdate() {}
        }
      `, Tsx: true},

		// ---- ObjectExpression in class field initializer with SCU silences class ----
		{Code: `
        class C extends React.Component {
          cfg = { shouldComponentUpdate() {} };
        }
      `, Tsx: true},

		// ---- ObjectExpression in static class field with SCU silences class ----
		{Code: `
        class C extends React.Component {
          static cfg = { shouldComponentUpdate() {} };
        }
      `, Tsx: true},

		// ---- nested ClassExpression with SCU silences itself; outer also has SCU ----
		{Code: `
        class Outer extends React.Component {
          shouldComponentUpdate() {}
          build() {
            return class extends React.Component { shouldComponentUpdate() {} };
          }
        }
      `, Tsx: true},

		// ---- export default class with PureRender decorator ----
		{Code: `
        @reactMixin.decorate(PureRenderMixin)
        export default class extends Component {}
      `, Tsx: true},

		// ---- top-level export const arrow SFC auto-silenced ----
		{Code: `
        export const Comp = (props) => <div />;
      `, Tsx: true},

		// ---- multi-args to PureRender decorator — first arg matches → silences ----
		{Code: `
        @reactMixin.decorate(PureRenderMixin, OtherMixin)
        class C extends Component {}
      `, Tsx: true},

		// ---- TS overload signature + concrete impl — concrete is the SCU. ----
		{Code: `
        class C extends React.Component {
          shouldComponentUpdate(): boolean;
          shouldComponentUpdate() { return true; }
        }
      `, Tsx: true},

		// ---- TS abstract `shouldComponentUpdate` alongside concrete — concrete
		// satisfies the SCU shortcut; abstract is rejected by the gate. ----
		{Code: `
        abstract class C extends React.Component {
          abstract shouldComponentUpdate(): boolean;
          shouldComponentUpdate() { return true; }
        }
      `, Tsx: true},

		// ---- TS generic PureComponent extends — type args don't break detection. ----
		{Code: `
        class C<P> extends React.PureComponent<P> {}
      `, Tsx: true},

		// ---- TS satisfies in extends — non-Identifier wrapper, class is not
		// detected as a React component → not reported. ----
		{Code: `
        class C extends (React.Component satisfies any) {
          shouldComponentUpdate() {}
        }
      `, Tsx: true},

		// ---- TS multi-overload signatures + concrete impl ----
		{Code: `
        class C extends React.Component {
          shouldComponentUpdate(): boolean;
          shouldComponentUpdate(p: any): boolean;
          shouldComponentUpdate(p?: any) { return true; }
        }
      `, Tsx: true},

		// ---- TS generic SCU method ----
		{Code: `
        class C extends React.Component {
          shouldComponentUpdate<T>(): boolean { return true; }
        }
      `, Tsx: true},

		// ---- TypeChecker-resolved binding: `let v; v = <div/>; return v;`
		// — `v` has no initializer, so the resolver can't prove it returns
		// JSX → Outer is NOT detected as SFC. ----
		{Code: `
        class C extends React.Component {
          shouldComponentUpdate() {}
          method() {
            function Outer() {
              let v;
              v = <div />;
              return v;
            }
          }
        }
      `, Tsx: true},

		// ---- Named FunctionExpression assigned to a lowercase const —
		// upstream resolves the binding name (`x`), not the inner function
		// id, so `Inner` is not classified as an SFC. ----
		{Code: `
        class C extends React.Component {
          shouldComponentUpdate() {}
          build() {
            const x = function Inner() { return <div />; };
          }
        }
      `, Tsx: true},

		// ---- ObjectExpression with SCU as JSX expression child silences
		// outer class via walk-up. ----
		{Code: `
        class C extends React.Component {
          render() {
            return <div>{({ shouldComponentUpdate() {} })}</div>;
          }
        }
      `, Tsx: true},

		// ---- ObjectExpression with SCU as JSX attribute value silences
		// outer class via walk-up. ----
		{Code: `
        class C extends React.Component {
          render() {
            return <Comp config={{ shouldComponentUpdate() {} }} />;
          }
        }
      `, Tsx: true},

		// ---- mixins: [...other, PureRenderMixin] — has Identifier element
		// matching PureRenderMixin → silences. ----
		{Code: `
        createReactClass({
          mixins: [...other, PureRenderMixin]
        });
      `, Tsx: true},

		// ---- Multiple walk-up sources in same component: any one suffices. ----
		{Code: `
        class C extends React.Component {
          build() {
            const a = { shouldComponentUpdate() {} };
            @reactMixin.decorate(PureRenderMixin)
            class Helper {}
          }
        }
      `, Tsx: true},

		// ---- Deeply nested ObjectExpression with SCU silences createReactClass. ----
		{Code: `
        createReactClass({
          level1: { level2: { level3: { shouldComponentUpdate: function() {} } } }
        });
      `, Tsx: true},

		// ---- Static block inside class: walk-up still bubbles SCU to outer class. ----
		{Code: `
        class C extends React.Component {
          static {
            const cfg = { shouldComponentUpdate() {} };
          }
        }
      `, Tsx: true},

		// ---- Capitalized class-field arrow is NOT an SFC (class-field
		// initializer position rejects). ----
		{Code: `
        class C extends React.Component {
          shouldComponentUpdate() {}
          Build = () => <div />;
        }
      `, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// ---- Edge: ClassExpression empty body extending React.Component ----
		{
			Code: `
        const X = class extends React.Component {};
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 19},
			},
		},

		// ---- Edge: anonymous ClassExpression as `module.exports = class extends React.Component { ... }` ----
		{
			Code: `
        module.exports = class extends React.Component {};
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 26},
			},
		},

		// ---- Edge: class with `shouldComponentUpdate = () => {}` class field —
		// upstream's MethodDefinition listener does NOT fire on PropertyDefinition,
		// so this still reports. Locks in upstream behavior. ----
		{
			Code: `
        class YourComponent extends React.Component {
          shouldComponentUpdate = () => true;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 9},
			},
		},

		// ---- Edge: shouldComponentUpdate as STRING-LITERAL key — upstream's
		// `key.name` is undefined for non-Identifier keys, so still reports. ----
		{
			Code: `
        class YourComponent extends React.Component {
          "shouldComponentUpdate"() {}
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 9},
			},
		},

		// ---- Edge: shouldComponentUpdate as COMPUTED key — same, no match ----
		{
			Code: "\n        class YourComponent extends React.Component {\n          [`shouldComponentUpdate`]() {}\n        }\n      ",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 9},
			},
		},

		// Locks in upstream `hasPureRenderDecorator` arm: optional-chain
		// `reactMixin?.decorate(PureRenderMixin)` does NOT match. tsgo flags
		// the chain via QuestionDotToken; we explicitly reject it.
		{
			Code: `
        @reactMixin?.decorate(PureRenderMixin)
        class DecoratedComponent extends Component {}
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 9},
			},
		},

		// Locks in upstream `hasPureRenderDecorator` arm: deeper member access
		// `outer.reactMixin.decorate(PureRenderMixin)` does NOT match
		// (`callee.object.name` is undefined on nested PropertyAccess).
		{
			Code: `
        @outer.reactMixin.decorate(PureRenderMixin)
        class DecoratedComponent extends Component {}
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 9},
			},
		},

		// Locks in upstream `hasPureRenderDecorator` arm: zero-argument call
		// `@reactMixin.decorate()` does NOT match (`arguments[0]` is undefined).
		{
			Code: `
        @reactMixin.decorate()
        class DecoratedComponent extends Component {}
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 9},
			},
		},

		// Locks in upstream `hasPureRenderDecorator` arm: non-Identifier first
		// arg (`@reactMixin.decorate(getMixin())`) does NOT match.
		{
			Code: `
        @reactMixin.decorate(getMixin())
        class DecoratedComponent extends Component {}
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 9},
			},
		},

		// Locks in upstream `hasPureRenderDecorator` arm: wrong wrapper name
		// `@otherWrapper.decorate(PureRenderMixin)` does NOT match.
		{
			Code: `
        @otherWrapper.decorate(PureRenderMixin)
        class DecoratedComponent extends Component {}
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 9},
			},
		},

		// Locks in upstream `hasPureRenderDecorator` arm: wrong method name
		// `@reactMixin.combine(PureRenderMixin)` does NOT match.
		{
			Code: `
        @reactMixin.combine(PureRenderMixin)
        class DecoratedComponent extends Component {}
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 9},
			},
		},

		// Locks in upstream `hasCustomDecorator` arm: CallExpression decorator
		// `@pureRender()` does NOT match (upstream `expression.name` is
		// undefined on CallExpression — only bare Identifier matches).
		{
			Code: `
        @pureRender()
        class DecoratedComponent extends Component {}
      `,
			Tsx:     true,
			Options: map[string]interface{}{"allowDecorators": []interface{}{"pureRender"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 9},
			},
		},

		// Locks in upstream `isPureRenderDeclared` arm: `mixins` value is NOT
		// an ArrayExpression — `node.value.elements` is undefined → no match.
		{
			Code: `
        createReactClass({
          mixins: PureRenderMixin
        })
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 26},
			},
		},

		// Locks in upstream `isPureRenderDeclared` arm: array element is a
		// CallExpression / non-Identifier — `element.name` is undefined → no match.
		{
			Code: `
        createReactClass({
          mixins: [resolveMixin()]
        })
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 26},
			},
		},

		// ---- Edge: settings.react.pragma="Preact" — `Preact.Component` IS
		// detected as a component (pragma matches), but no SCU → reports. ----
		{
			Code: `
        class YourComponent extends Preact.Component {}
      `,
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"pragma": "Preact"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 9},
			},
		},

		// ---- Edge: nested `class extends React.Component` inside an outer non-
		// component class body — listener fires per ClassDeclaration, so the
		// inner class is independently checked. Locks in tree traversal. ----
		{
			Code: `
        class Outer {
          method() {
            class Inner extends React.Component {}
            return Inner;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 4, Column: 13},
			},
		},

		// ---- Edge: parens around obj arg — `createReactClass(({}))` is still
		// detected (paren-transparent on createReactClass detection). ----
		{
			Code: `
        createReactClass(({}))
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 27},
			},
		},

		// ---- Edge: createReactClass with non-Identifier mixin element (template literal) ----
		{
			Code: "\n        createReactClass({\n          mixins: [`PureRenderMixin`]\n        })\n      ",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 26},
			},
		},

		// ---- Edge: createReactClass with `shouldComponentUpdate` on COMPUTED key ----
		{
			Code: "\n        createReactClass({\n          [`shouldComponentUpdate`]: function() {}\n        })\n      ",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 26},
			},
		},

		// ---- Edge: createReactClass with elision-only mixins array — no Identifier element ----
		{
			Code: `
        createReactClass({
          mixins: [,,,]
        })
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 26},
			},
		},

		// Locks in upstream `Components.detect` ClassExpression arm: anonymous
		// ClassExpression in CallExpression argument position is still a
		// React.Component subclass, still reported.
		{
			Code: `
        register(class extends React.Component {});
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 18},
			},
		},

		// ---- Options coverage: bare-object option shape (CLI-shaped) — locks
		// in GetOptionsMap integration. Without GetOptionsMap (typed-struct-only)
		// this would silently default and allow @pure to be reported. ----
		{
			Code: `
        @pure
        class DecoratedComponent extends Component {}
      `,
			Tsx:     true,
			Options: map[string]interface{}{"allowDecorators": []interface{}{"renderPure"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 9},
			},
		},

		// ---- Options coverage: array-wrapped option shape (rule_tester / multi-element CLI shape) ----
		{
			Code: `
        @pure
        class DecoratedComponent extends Component {}
      `,
			Tsx:     true,
			Options: []interface{}{map[string]interface{}{"allowDecorators": []interface{}{"renderPure"}}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 9},
			},
		},

		// ---- Edge (TS): generic type args + no SCU ----
		{
			Code: `
        class YourComponent extends React.Component<Props, State> {
          render() { return null; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 9},
			},
		},

		// ---- Edge (TS): abstract React.Component subclass without SCU ----
		{
			Code: `
        abstract class YourComponent extends React.Component {
          render() { return null; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 9},
			},
		},

		// ---- Edge: SpreadElement first arg in decorator — `@reactMixin.decorate(...args)`
		// — first argument is a SpreadAssignment, not an Identifier → does NOT match. ----
		{
			Code: `
        @reactMixin.decorate(...args)
        class DecoratedComponent extends Component {}
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 9},
			},
		},

		// ---- Edge: doubly-nested class — only inner extends React.Component, no SCU ----
		{
			Code: `
        class Outer {
          method() {
            class Inner extends React.Component {
              render() { return null; }
            }
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 4, Column: 13},
			},
		},

		// ---- Edge: SpreadAssignment in createReactClass arg — has no key,
		// must not match SCU/PureRenderMixin gate. ----
		{
			Code: `
        createReactClass({
          ...rest
        })
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 26},
			},
		},

		// ---- Edge: ShorthandPropertyAssignment `mixins` — no value, can't be
		// an array, gate fails. ----
		{
			Code: `
        const mixins = [PureRenderMixin];
        createReactClass({
          mixins
        })
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 3, Column: 26},
			},
		},

		// ---- Edge: multiple class components in same file — each is evaluated independently ----
		{
			Code: `
        class A extends React.Component {}
        class B extends React.Component {
          shouldComponentUpdate() {}
        }
        class C extends React.Component {}
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 9},
				{MessageId: "noShouldComponentUpdate", Line: 6, Column: 9},
			},
		},

		// ---- Position: `export default class extends React.Component {}` —
		// upstream ESTree puts ExportDefaultDeclaration outside ClassDeclaration,
		// so the report anchors on `class`. Locks in ClassKeywordStart trim.
		// `        export default class …` → Column 24. ----
		{
			Code: `
        export default class extends React.Component {}
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 24},
			},
		},

		// ---- Position: `export class C extends React.Component {}` —
		// `        export class …` → Column 16. ----
		{
			Code: `
        export class YourComponent extends React.Component {}
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 16},
			},
		},

		// ---- Position: `export default class C extends React.Component {}`
		// (named default-exported) — same Column 24 as anonymous form. ----
		{
			Code: `
        export default class C extends React.Component {}
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 24},
			},
		},

		// ---- Position: `abstract class …` — abstract IS part of ClassDeclaration's
		// range upstream (TSESTree keeps it on the class), so report starts at
		// `abstract`. `        abstract class …` → Column 9. ----
		{
			Code: `
        abstract class YourComponent extends React.Component {}
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 9},
			},
		},

		// ---- Position: `declare class …` — `declare` is also a TS modifier
		// that belongs to the ClassDeclaration's range upstream. ----
		// SKIP: tsgo errors out on a declare class with a body — this is a
		// declare-only feature and the body content makes the input invalid.
		// (Documented for completeness; no observable rule behavior.)

		// ---- Position: `export default abstract class …` — export/default
		// trimmed, `abstract` retained.
		// `        export default abstract …` → `abstract` at Column 24. ----
		{
			Code: `
        export default abstract class extends React.Component {}
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 24},
			},
		},

		// ---- Position: decorator + export default — decorator is BEFORE
		// `export default`. tsgo orders modifiers by source position so the
		// trim must walk all leading export/default and stop at decorator. ----
		// Some bundlers / babel outputs emit `@dec\nexport default class`.
		// In TS, decorators after `export default` are valid for class only
		// in Stage-3 form; `@dec export default class` is the legacy form.
		// We test only the normalized "decorator first" form (TS standard). ----
		{
			Code: `
        @observer
        export default class extends React.Component {}
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				// `@observer` decorator at Line 2 Column 9, included in the range.
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 9},
			},
		},

		// ---- Real user: connect HOC + inline ClassExpression without SCU ----
		{
			Code: `
        const Connected = connect(mapStateToProps)(class extends React.Component {});
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				// `class extends React.Component {}` — `class` keyword position.
				// `        const Connected = connect(mapStateToProps)(class …` →
				// 8 + len("const Connected = connect(mapStateToProps)(") = 8 + 43 = 51 + 1 = Column 52.
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 52},
			},
		},

		// ---- Real user: ClassExpression in expression statement — `(class … {});` ----
		{
			Code: `
        (class extends React.Component {});
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				// ParenthesizedExpression wraps; ClassExpression itself starts
				// at `class`. `        (class …` → 8 + 1 = Column 10.
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 10},
			},
		},

		// ---- Real user: multiple createReactClass calls in same file ----
		{
			Code: `
        const A = createReactClass({});
        const B = createReactClass({ shouldComponentUpdate() {} });
        const C = createReactClass({ mixins: [RandomMixin] });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 36},
				{MessageId: "noShouldComponentUpdate", Line: 4, Column: 36},
			},
		},

		// ---- Real user: deeply nested class inside object inside class
		// — listener fires on each ClassDeclaration independently. ----
		{
			Code: `
        const factory = {
          make() {
            return class Inner extends React.Component {};
          }
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				// `class Inner extends …` — class keyword position.
				// `            return class …` → 12 + len("return ") = 19 + 1 = Column 20.
				{MessageId: "noShouldComponentUpdate", Line: 4, Column: 20},
			},
		},

		// ---- Real user: TS access modifier on SCU — modifier doesn't change name match.
		// COUNTEREXAMPLE: `private otherMethod()` doesn't satisfy SCU. ----
		{
			Code: `
        class C extends React.Component {
          private otherMethod() {}
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 9},
			},
		},

		// ---- Edge: class with computed-key static field that LOOKS like SCU
		// — computed key is not Identifier, so doesn't match. ----
		{
			Code: "\n        class C extends React.Component {\n          static [`shouldComponentUpdate`] = () => true;\n        }\n      ",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 9},
			},
		},

		// ---- Robustness: options malformed — `allowDecorators` is a string,
		// not array → must fall back to default (empty) and report normally. ----
		{
			Code: `
        @pure
        class C extends Component {}
      `,
			Tsx:     true,
			Options: map[string]interface{}{"allowDecorators": "pure"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 9},
			},
		},

		// ---- Robustness: options has null/non-string entries in
		// allowDecorators — non-string elements ignored, valid ones still apply. ----
		{
			Code: `
        @observer
        class C extends Component {}
      `,
			Tsx:     true,
			Options: map[string]interface{}{"allowDecorators": []interface{}{nil, 42, "pure"}},
			Errors: []rule_tester.InvalidTestCaseError{
				// "observer" not in allow-list ("pure" is the only string), so reports.
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 9},
			},
		},

		// ---- Robustness: options is `[null]` — falls back to default. ----
		{
			Code: `
        class C extends React.Component {}
      `,
			Tsx:     true,
			Options: []interface{}{nil},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 9},
			},
		},

		// ---- Robustness: options is `[{}]` empty object — defaults apply. ----
		{
			Code: `
        @pure
        class C extends Component {}
      `,
			Tsx:     true,
			Options: []interface{}{map[string]interface{}{}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 9},
			},
		},

		// ---- Robustness: options is `[{ allowDecorators: [] }]` empty array
		// — same as defaults. ----
		{
			Code: `
        @pure
        class C extends Component {}
      `,
			Tsx:     true,
			Options: []interface{}{map[string]interface{}{"allowDecorators": []interface{}{}}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 9},
			},
		},

		// ---- Real user: ClassExpression assigned to module-level const,
		// without SCU — class reports on its `class` keyword. ----
		{
			Code: `
        const C = class extends React.Component {};
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				// `        const C = class …` → 8 + len("const C = ") = 18 + 1 = Column 19.
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 19},
			},
		},

		// ---- Real user: deeply nested SIBLING component reports — outer
		// is detected, inner is detected, both reported. ----
		{
			Code: `
        function outer() {
          class A extends React.Component {}
          function inner() {
            class B extends React.Component {}
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 3, Column: 11},
				{MessageId: "noShouldComponentUpdate", Line: 5, Column: 13},
			},
		},

		// ---- Locks in upstream behavior: ClassExpression extending
		// React.PureComponent is REPORTED. Upstream registers
		// `ClassDeclaration(node)` only — its PureComponent shortcut never
		// fires for ClassExpression. Verified empirically against
		// eslint-plugin-react. ----
		{
			Code: `
        const X = class extends React.PureComponent {};
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				// `        const X = class …` → 8 + len("const X = ") = 18 + 1 = Column 19.
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 19},
			},
		},

		// ---- Locks in upstream: ClassExpression PureComponent in CallExpression
		// argument position — still reported. ----
		{
			Code: `
        register(class extends React.PureComponent {});
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				// `        register(class …` → 8 + len("register(") = 17 + 1 = Column 18.
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 18},
			},
		},

		// ---- Locks in upstream: ClassExpression PureComponent on RHS of
		// `module.exports =` AssignmentExpression — still reported. ----
		{
			Code: `
        module.exports = class extends React.PureComponent {};
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				// `        module.exports = class …` → 8 + len("module.exports = ") = 25 + 1 = Column 26.
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 26},
			},
		},

		// ---- Locks in upstream: ClassExpression PureComponent as object
		// property value — still reported. ----
		{
			Code: `
        const obj = { Comp: class extends React.PureComponent {} };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				// `        const obj = { Comp: class …` → 8 + len("const obj = { Comp: ") = 28 + 1 = Column 29.
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 29},
			},
		},

		// ---- Locks in upstream: ClassExpression PureComponent wrapped in HOC
		// (real-world Redux `connect(...)(Component)` pattern) — still reported. ----
		{
			Code: `
        const Connected = connect(mapStateToProps)(class extends React.PureComponent {});
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				// `        const Connected = connect(mapStateToProps)(class …` →
				// 8 + len("const Connected = connect(mapStateToProps)(") = 8 + 43 = 51 + 1 = Column 52.
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 52},
			},
		},

		// ---- Locks in upstream: bare ClassExpression PureComponent — same
		// `<bare>PureComponent` regex match, still reported. ----
		{
			Code: `
        const X = class extends PureComponent {};
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				// `        const X = class …` → Column 19.
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 19},
			},
		},

		// ---- Locks in upstream: ClassExpression with allow-listed decorator
		// — `hasCustomDecorator` shortcut also gated to ClassDeclaration only,
		// so decorated ClassExpression still reports. ----
		// SKIP: tsgo's parser rejects `@dec class` decorator syntax on
		// ClassExpression (legacy decorators are class-statement only).
		// `const X = @dec class extends Component {}` is a parse error.
		// Documented for completeness; no observable rule case.

		// ===== Walk-up boundary: nested detected component absorbs the markSCU =====

		// ---- nested React.Component class with SCU absorbs SCU; outer
		// React.Component still reports because outer has no SCU of its own. ----
		{
			Code: `
        class Outer extends React.Component {
          build() {
            class Inner extends React.Component {
              shouldComponentUpdate() {}
            }
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				// Outer at 'class' keyword.
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 9},
			},
		},

		// ---- nested PureComponent absorbs its own (self) SCU; outer
		// React.Component still reports. ----
		{
			Code: `
        class Outer extends React.Component {
          build() {
            class Inner extends React.PureComponent {}
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 9},
			},
		},

		// ---- nested obj-with-SCU inside nested React.Component absorbs the
		// SCU into INNER, so OUTER still reports. ----
		{
			Code: `
        class Outer extends React.Component {
          build() {
            class Inner extends React.Component {
              method() {
                return { shouldComponentUpdate() {} };
              }
            }
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 9},
			},
		},

		// ---- inner createReactClass absorbs its own SCU; outer class
		// still reports. ----
		{
			Code: `
        class C extends React.Component {
          init() {
            const cfg = { shouldComponentUpdate() {} };
            createReactClass({});
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				// outer has cfg with SCU → C is silenced
				// createReactClass({}) without SCU → reported at the {} object.
				// `            createReactClass({})` → 12 + len("createReactClass(") = 29 + 1 = Column 30.
				{MessageId: "noShouldComponentUpdate", Line: 5, Column: 30},
			},
		},

		// ---- two ClassDeclaration components, one with walk-up obj scu,
		// one without — only the second reports. ----
		{
			Code: `
        class A extends React.Component {
          init() {
            const cfg = { shouldComponentUpdate() {} };
          }
        }
        class B extends React.Component {}
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				// `        class B extends …` at Line 7 Column 9.
				{MessageId: "noShouldComponentUpdate", Line: 7, Column: 9},
			},
		},

		// ---- top-level obj with SCU does NOT walk up (no enclosing
		// component); class A is still reported. ----
		{
			Code: `
        const cfg = { shouldComponentUpdate() {} };
        class A extends React.Component {}
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 3, Column: 9},
			},
		},

		// ---- ClassExpression PureComponent inside class field static —
		// inner ClassExpression reports (not silenced by PureComponent — it's
		// ClassExpression), outer C also reports (no SCU of its own). ----
		{
			Code: `
        class C extends React.Component {
          static Inner = class extends React.PureComponent {};
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 9},
				// Inner ClassExpression at `class` keyword: 18 col
				// (`          static Inner = class …` → 10 + len("static Inner = ") = 25 + 1 = Col 26).
				{MessageId: "noShouldComponentUpdate", Line: 3, Column: 26},
			},
		},

		// ---- inner ClassExpression with SCU silences itself, outer C still
		// reports. ----
		{
			Code: `
        class C extends React.Component {
          static Inner = class extends React.Component {
            shouldComponentUpdate() {}
          };
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 9},
			},
		},

		// ---- ClassExpression returned from helper function inside class
		// method — Inner is detected (extends React.PureComponent), but
		// PureComponent shortcut doesn't fire (ClassExpression), so Inner
		// reports. Outer C also reports. ----
		{
			Code: `
        class C extends React.Component {
          build() {
            function inner() {
              return class extends React.PureComponent {};
            }
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 9},
				// Inner ClassExpression: `              return class …` → 14 + len("return ") = 21 + 1 = Col 22.
				{MessageId: "noShouldComponentUpdate", Line: 5, Column: 22},
			},
		},

		// ---- inner class with @PureRender decorator absorbs SCU; outer
		// React.Component still reports. ----
		{
			Code: `
        class Outer extends React.Component {
          build() {
            @reactMixin.decorate(PureRenderMixin)
            class Inner extends Component {}
            return Inner;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 9},
			},
		},

		// ---- createReactClass with nested ClassExpression that has SCU
		// — outer createReactClass has no SCU of its own, ClassExpression
		// absorbs its SCU, so only outer reports. ----
		{
			Code: `
        createReactClass({
          inner: class extends React.Component {
            shouldComponentUpdate() {}
          }
        })
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				// `        createReactClass({...` → 8 + len("createReactClass(") = 25 + 1 = Col 26.
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 26},
			},
		},

		// ---- top-level mixins var doesn't walk up to createReactClass({})
		// — Identifier path; createReactClass has empty obj, reports. ----
		{
			Code: `
        const x = { mixins: [PureRenderMixin] };
        createReactClass({});
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				// `        createReactClass({})` → 8 + len("createReactClass(") = 25 + 1 = Col 26.
				{MessageId: "noShouldComponentUpdate", Line: 3, Column: 26},
			},
		},

		// ---- outer class method with empty createReactClass — both
		// outer and inner have no SCU, both report. ----
		{
			Code: `
        class C extends React.Component {
          buildConfig() {
            return createReactClass({});
          }
          render() { return null; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 9},
				// inner createReactClass({}) at the {} object position:
				// `            return createReactClass({})` → 12 + len("return createReactClass(") = 36 + 1 = Col 37.
				{MessageId: "noShouldComponentUpdate", Line: 4, Column: 37},
			},
		},

		// ===== SFC inside ClassDeclaration method body =====
		// upstream's `isFunctionInClass` blocks `mark-SCU-as-declared(self)`,
		// so a SFC inside a ClassDeclaration method body falls into
		// `components.list()` without being silenced and reports.

		// ---- function declaration inside class method ----
		{
			Code: `
        class C extends React.Component {
          shouldComponentUpdate() {}
          build() {
            function Inner(p) { return <div />; }
            return <Inner />;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				// `            function Inner …` → 12 + 1 = Column 13.
				{MessageId: "noShouldComponentUpdate", Line: 5, Column: 13},
			},
		},

		// ---- arrow assigned to const inside class method ----
		{
			Code: `
        class C extends React.Component {
          shouldComponentUpdate() {}
          build() {
            const Inner = (p) => <div />;
            return <Inner />;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				// arrow ReportNode anchors at `(p) =>` — `            const Inner = (p) => …` →
				// 12 + len("const Inner = ") = 26 + 1 = Column 27.
				{MessageId: "noShouldComponentUpdate", Line: 5, Column: 27},
			},
		},

		// ---- function declaration inside block scope inside method ----
		{
			Code: `
        class C extends React.Component {
          shouldComponentUpdate() {}
          build() {
            {
              function Inner() { return <div />; }
            }
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				// `              function Inner …` → 14 + 1 = Column 15.
				{MessageId: "noShouldComponentUpdate", Line: 6, Column: 15},
			},
		},

		// ---- in-class-method SFC absorbs an obj-with-SCU sibling
		// (Inner.something = { scu() {} }) → the sibling assigns SCU to
		// outer C (walk-up to nearest detected). Inner itself has no SCU
		// inside its body, so Inner reports. ----
		{
			Code: `
        class C extends React.Component {
          build() {
            function Inner() {
              return <div />;
            }
            Inner.something = { shouldComponentUpdate() {} };
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				// Inner reports: `            function Inner …` → 12 + 1 = Col 13, Line 4.
				{MessageId: "noShouldComponentUpdate", Line: 4, Column: 13},
			},
		},

		// ---- class with no SCU + SFC inside method body — both report ----
		{
			Code: `
        class C extends React.Component {
          build() {
            function Inner() { return <div />; }
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 9},
				// `            function Inner …` → Col 13, Line 4.
				{MessageId: "noShouldComponentUpdate", Line: 4, Column: 13},
			},
		},

		// ===== Settings: pragma + bare component =====
		// ---- pragma="Preact" + Preact.Component without SCU ----
		{
			Code: `
        class C extends Preact.Component {}
      `,
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"pragma": "Preact"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 9},
			},
		},

		// ---- pragma="Preact" + bare Component matches (regex `(Preact\.)?Component`) ----
		{
			Code: `
        class C extends Component {}
      `,
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"pragma": "Preact"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 9},
			},
		},

		// ---- createClass="myCreateClass" + custom factory empty obj ----
		{
			Code: `
        myCreateClass({});
      `,
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"createClass": "myCreateClass"}},
			Errors: []rule_tester.InvalidTestCaseError{
				// `        myCreateClass({})` → 8 + len("myCreateClass(") = 22 + 1 = Col 23.
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 23},
			},
		},

		// ---- ClassExpression with SFC inside method — outer ClassExpression
		// still reports (PureComponent shortcut doesn't apply to
		// ClassExpression and there's no SCU). Inner SFC is auto-silenced
		// because upstream's `isFunctionInClass` doesn't match ClassExpression. ----
		{
			Code: `
        const X = class extends React.Component {
          build() {
            function Inner(p) { return <div />; }
            return <Inner />;
          }
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				// `        const X = class …` → 8 + len("const X = ") = 18 + 1 = Col 19.
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 19},
			},
		},

		// ===== Misc verified-against-upstream invalid cases =====

		// ---- IIFE in class method body containing SFC reports both C and Inner ----
		{
			Code: `
        class C extends React.Component {
          build() {
            (function() {
              function Inner() { return <div />; }
            })();
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate"},
				{MessageId: "noShouldComponentUpdate"},
			},
		},

		// ---- nested function chain in class method body — all 3 report ----
		{
			Code: `
        class C extends React.Component {
          method() {
            function Outer() {
              function Inner() { return <div />; }
              return <Inner />;
            }
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate"},
				{MessageId: "noShouldComponentUpdate"},
				{MessageId: "noShouldComponentUpdate"},
			},
		},

		// ---- arrow callback in class method (no name binding) — not detected
		// as SFC; outer C still reports (no SCU). ----
		{
			Code: `
        class C extends React.Component {
          build() {
            return list.map((item) => <div key={item} />);
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate"},
			},
		},

		// ---- nested class with PureRender decorator extending React.Component
		// (Helper is detected, self-marked); outer reports. ----
		{
			Code: `
        class Outer extends React.Component {
          build() {
            @reactMixin.decorate(PureRenderMixin)
            class Helper extends React.Component {}
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate"},
			},
		},

		// ---- array of 3 ClassExpressions, all 3 report ----
		{
			Code: `[class A extends React.Component {}, class B extends React.PureComponent {}, class C extends Component {}];`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate"},
				{MessageId: "noShouldComponentUpdate"},
				{MessageId: "noShouldComponentUpdate"},
			},
		},

		// ---- top-level obj-with-SCU bound to const, used as class field
		// initializer — does not walk up through binding. ----
		{
			Code: `
        const obj = { shouldComponentUpdate: () => true };
        class C extends React.Component {
          use = obj;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate"},
			},
		},

		// ---- helper SFC declared inside render method body reports ----
		{
			Code: `
        class C extends React.Component {
          shouldComponentUpdate() {}
          render() {
            function Helper() { return <span />; }
            return <div><Helper /></div>;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate"},
			},
		},

		// ---- class field arrow returning JSX is NOT detected as SFC,
		// outer class still reports (no SCU). ----
		{
			Code: `
        class C extends React.Component {
          render = () => <div />;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 9},
			},
		},

		// ---- nested ClassExpression PureComponent inside ClassDecl method
		// (outer has SCU, inner reports because PureComponent shortcut
		// doesn't apply to ClassExpression). ----
		{
			Code: `
        class Outer extends React.Component {
          shouldComponentUpdate() {}
          build() {
            return class extends React.PureComponent {};
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate"},
			},
		},

		// ---- decorator factory (CallExpression) with PureRenderMixin arg —
		// not the recognized `@reactMixin.decorate(PureRenderMixin)` shape. ----
		{
			Code: `
        @createDecorator(PureRenderMixin)
        class C extends Component {}
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate"},
			},
		},

		// ---- decorator with PropertyAccess first argument — not bare Identifier
		// `PureRenderMixin`, doesn't silence. ----
		{
			Code: `
        @reactMixin.decorate(MyMixin.PureRenderMixin)
        class C extends Component {}
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate"},
			},
		},

		// ---- TS abstract shouldComponentUpdate-only (no concrete impl) —
		// upstream's TSESTree maps `abstract foo()` to
		// `TSAbstractMethodDefinition` (not `MethodDefinition`), so the
		// MethodDefinition listener never fires. We mirror the rejection
		// via the abstract-modifier gate. ----
		{
			Code: `
        abstract class C extends React.Component {
          abstract shouldComponentUpdate(): boolean;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 9},
			},
		},

		// ---- TS declare class — no body, no SCU, still reports. ----
		{
			Code: `
        declare class C extends React.Component {}
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 9},
			},
		},

		// ---- TS abstract class without SCU — still reports. ----
		{
			Code: `
        abstract class C extends React.Component {
          abstract render(): JSX.Element;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 9},
			},
		},

		// ---- TS namespace-wrapped class — detected, no SCU, reports. ----
		{
			Code: `
        namespace Foo {
          export class C extends React.Component {}
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate"},
			},
		},

		// ---- TS module-wrapped class (legacy `module Foo { ... }`). ----
		{
			Code: `
        module Foo {
          class C extends React.Component {}
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate"},
			},
		},

		// ---- TS interface extending React.Component is NOT a class component;
		// the sibling class still reports. ----
		{
			Code: `
        interface I extends React.Component {}
        class C extends React.Component {}
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 3, Column: 9},
			},
		},

		// ---- PropertyDeclaration named SCU with no initializer is NOT a method;
		// class still reports. ----
		{
			Code: `
        class C extends React.Component {
          shouldComponentUpdate;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 9},
			},
		},

		// ---- decorator with nested member access deeper than the matched
		// `<Identifier>.<name>` shape — does NOT match. ----
		{
			Code: `
        @reactMixin.decorate.fn(PureRenderMixin)
        class C extends Component {}
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate"},
			},
		},

		// ---- paren-wrapped bare-Identifier decorator — not a CallExpression,
		// so `hasPureRenderDecorator` doesn't match. ----
		{
			Code: `
        @(PureRenderMixin)
        class C extends Component {}
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate"},
			},
		},

		// ---- class with only a constructor — constructor key is "constructor",
		// not SCU; reports. ----
		{
			Code: `
        class C extends React.Component {
          constructor() { super(); }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 9},
			},
		},

		// ---- TypeChecker-resolved binding: `const view = <div/>; return view;`
		// — `view` initializer is JSX, resolver classifies Outer as SFC,
		// Outer is in class method body → reports. Lock-in for the
		// `FunctionReturnsJSXOrNullWithChecker` path. ----
		{
			Code: `
        class C extends React.Component {
          shouldComponentUpdate() {}
          method() {
            function Outer() {
              const view = <div />;
              return view;
            }
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				// Outer at `function` keyword:
				// `            function Outer() {` → 12 + 1 = Col 13.
				{MessageId: "noShouldComponentUpdate", Line: 5, Column: 13},
			},
		},

		// ---- TypeChecker-resolved cross-scope binding: `view` declared in
		// outer method scope, referenced from `Outer` body. Without a
		// TypeChecker, the local-block scan misses it; with the
		// TypeChecker, the symbol resolves to its declaration in the
		// outer scope and Outer classifies as SFC. ----
		{
			Code: `
        class C extends React.Component {
          shouldComponentUpdate() {}
          method() {
            const view = <div />;
            function Outer() {
              return view;
            }
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				// Outer at `function` keyword:
				// `            function Outer() {` → 12 + 1 = Col 13.
				{MessageId: "noShouldComponentUpdate", Line: 6, Column: 13},
			},
		},

		// ---- SFC inside class static block — reaches ClassDeclaration via
		// ClassStaticBlockDeclaration sentinel. ----
		{
			Code: `
        class C extends React.Component {
          shouldComponentUpdate() {}
          static {
            function Inner() { return <div />; }
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate"},
			},
		},

		// ---- Named FunctionExpression returned from method body, no
		// outer binding — Inner reported via its own id. ----
		{
			Code: `
        class C extends React.Component {
          shouldComponentUpdate() {}
          build() {
            return function Inner() { return <div />; };
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate"},
			},
		},

		// ---- Anonymous arrow as object property value with capitalized
		// key (Inner) — reported via PropertyAssignment-key binding name. ----
		{
			Code: `
        class C extends React.Component {
          shouldComponentUpdate() {}
          build() {
            const obj = { Inner: () => <div /> };
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate"},
			},
		},

		// ---- Anonymous nested arrow as outer arrow's body inside a
		// class field — inner arrow's enclosing scope walks through the
		// class field (PropertyDeclaration sentinel) up to the class. ----
		{
			Code: `
        class C extends React.Component {
          shouldComponentUpdate() {}
          build = () => () => <div />;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate"},
			},
		},

		// ---- Anonymous arrow returned from method body — reported. ----
		{
			Code: `
        class C extends React.Component {
          shouldComponentUpdate() {}
          build() {
            return () => <div />;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate"},
			},
		},

		// ---- Double anonymous arrow returned from method body — inner-most
		// reported. ----
		{
			Code: `
        class C extends React.Component {
          shouldComponentUpdate() {}
          build() {
            return () => () => <div />;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate"},
			},
		},
	})
}

// TestRequireOptimizationRule_NilTypeChecker verifies the rule operates
// without panicking when the TypeChecker is unavailable. rslint schedules
// rules without `RequiresTypeInfo: true` against "gap files" (files in the
// program but not in `typeInfoFiles`) with a nil checker; the rule must
// degrade gracefully — the only TC-aware path (SFC classification via
// `reactutil.IsStatelessReactComponentWithChecker`) falls back to a
// local-block scan when `tc == nil`.
//
// The fixture below covers every detection path of the rule:
//   - ClassDeclaration extending React.Component (no SCU → reports)
//   - ClassExpression extending React.PureComponent (no SCU shortcut, reports)
//   - createReactClass({}) ObjectLiteralExpression
//   - SFC inside ClassDeclaration method body (TC-aware classification)
//   - obj-with-SCU walk-up (silences class)
//   - decorator + extends Component
//
// Success criterion: the rule's Run callback executes end-to-end without
// panicking. Reports may differ from the TC-enabled path; we don't assert
// specific messages, only non-panic behavior.
func TestRequireOptimizationRule_NilTypeChecker(t *testing.T) {
	t.Parallel()
	rootDir := fixtures.GetRootDir()
	filePath := tspath.ResolvePath(rootDir, "react.tsx")
	code := `
class A extends React.Component {}
class B extends React.PureComponent { shouldComponentUpdate() {} }
class C extends React.Component {
  shouldComponentUpdate() {}
  build() {
    function Inner() { return <div />; }
    const Sibling = (p) => <div />;
    return <Inner />;
  }
}
class D extends React.Component {
  static cfg = { shouldComponentUpdate() {} };
}
const E = class extends React.Component {
  shouldComponentUpdate() {}
};
createReactClass({});
createReactClass({ shouldComponentUpdate: function () {} });
createReactClass({ mixins: [PureRenderMixin] });
@reactMixin.decorate(PureRenderMixin)
class F extends Component {}
function TopLevelComp(props) { return <div />; }
const ArrowComp = (p) => <div />;
`
	fs := utils.NewOverlayVFSForFile(filePath, code)
	program, err := utils.CreateProgram(
		true, fs, rootDir, "tsconfig.json", utils.CreateCompilerHost(rootDir, fs),
	)
	if err != nil {
		t.Fatalf("CreateProgram: %v", err)
	}
	sourceFile := program.GetSourceFile(filePath)
	if sourceFile == nil {
		t.Fatalf("source file not found for %s", filePath)
		return
	}

	ctx := (rule.RuleContext{
		SourceFile:  sourceFile,
		Program:     program,
		Settings:    map[string]interface{}{},
		TypeChecker: nil, // explicitly nil — this is the path under test
	}).WithReporter("test/require-optimization", rule.SeverityWarning, func(rule.RuleDiagnostic) {})

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("RequireOptimizationRule.Run() panicked with nil TypeChecker: %v", r)
		}
	}()

	// The rule does its full walk inside Run() (single-pass collect-and-
	// report) and returns an empty listener map. Calling Run() is enough
	// to exercise every code path against the fixture.
	_ = RequireOptimizationRule.Run(ctx, nil)
}

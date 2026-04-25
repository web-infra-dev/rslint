package prefer_stateless_function

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferStatelessFunctionRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferStatelessFunctionRule, []rule_tester.ValidTestCase{
		// ---- Upstream valid: already a stateless function ----
		{Code: `
        const Foo = function(props) {
          return <div>{props.foo}</div>;
        };
      `, Tsx: true},

		// ---- Upstream valid: already a stateless arrow function ----
		{Code: `const Foo = ({foo}) => <div>{foo}</div>;`, Tsx: true},

		// ---- Upstream valid: PureComponent + props + ignorePureComponents ----
		{Code: `
        class Foo extends React.PureComponent {
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `, Tsx: true, Options: map[string]interface{}{"ignorePureComponents": true}},

		// ---- Upstream valid: PureComponent + context + ignorePureComponents ----
		{Code: `
        class Foo extends React.PureComponent {
          render() {
            return <div>{this.context.foo}</div>;
          }
        }
      `, Tsx: true, Options: map[string]interface{}{"ignorePureComponents": true}},

		// ---- Upstream valid: PureComponent in expression context + ignorePureComponents ----
		{Code: `
        const Foo = class extends React.PureComponent {
          render() {
            return <div>{this.props.foo}</div>;
          }
        };
      `, Tsx: true, Options: map[string]interface{}{"ignorePureComponents": true}},

		// ---- Upstream valid: has lifecycle method ----
		{Code: `
        class Foo extends React.Component {
          shouldComponentUpdate() {
            return false;
          }
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `, Tsx: true},

		// ---- Upstream valid: has state — uses this.setState / this.state ----
		{Code: `
        class Foo extends React.Component {
          changeState() {
            this.setState({foo: "clicked"});
          }
          render() {
            return <div onClick={this.changeState.bind(this)}>{this.state.foo || "bar"}</div>;
          }
        }
      `, Tsx: true},

		// ---- Upstream valid: uses this.refs ----
		{Code: `
        class Foo extends React.Component {
          doStuff() {
            this.refs.foo.style.backgroundColor = "red";
          }
          render() {
            return <div ref="foo" onClick={this.doStuff}>{this.props.foo}</div>;
          }
        }
      `, Tsx: true},

		// ---- Upstream valid: has additional method ----
		{Code: `
        class Foo extends React.Component {
          doStuff() {}
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `, Tsx: true},

		// ---- Upstream valid: empty (no super) constructor — not "useless" ----
		{Code: `
        class Foo extends React.Component {
          constructor() {}
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `, Tsx: true},

		// ---- Upstream valid: constructor with non-super body ----
		{Code: `
        class Foo extends React.Component {
          constructor() {
            doSpecialStuffs();
          }
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `, Tsx: true},

		// ---- Upstream valid: constructor with non-super body (2) ----
		{Code: `
        class Foo extends React.Component {
          constructor() {
            foo;
          }
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `, Tsx: true},

		// ---- Upstream valid: uses this.bar — useThis ----
		{Code: `
        class Foo extends React.Component {
          render() {
            return <div>{this.bar}</div>;
          }
        }
      `, Tsx: true},

		// ---- Upstream valid: destructures this.bar — useThis ----
		{Code: `
        class Foo extends React.Component {
          render() {
            let {props:{foo}, bar} = this;
            return <div>{foo}</div>;
          }
        }
      `, Tsx: true},

		// ---- Upstream valid: this[bar] — element access with Identifier key ----
		{Code: `
        class Foo extends React.Component {
          render() {
            return <div>{this[bar]}</div>;
          }
        }
      `, Tsx: true},

		// ---- Upstream valid: this['bar'] — element access with string-literal key ----
		{Code: `
        class Foo extends React.Component {
          render() {
            return <div>{this['bar']}</div>;
          }
        }
      `, Tsx: true},

		// ---- Upstream valid: render returns null with React 0.14.0 ----
		{Code: `
        class Foo extends React.Component {
          render() {
            if (!this.props.foo) {
              return null;
            }
            return <div>{this.props.foo}</div>;
          }
        }
      `, Tsx: true, Settings: map[string]interface{}{"react": map[string]interface{}{"version": "0.14.0"}}},

		// ---- Upstream valid: ES5 createReactClass returns null with React 0.14.0 ----
		{Code: `
        var Foo = createReactClass({
          render: function() {
            if (!this.props.foo) {
              return null;
            }
            return <div>{this.props.foo}</div>;
          }
        });
      `, Tsx: true, Settings: map[string]interface{}{"react": map[string]interface{}{"version": "0.14.0"}}},

		// ---- Upstream valid: shorthand-if returning null with React 0.14.0 ----
		{Code: `
        class Foo extends React.Component {
          render() {
            return true ? <div /> : null;
          }
        }
      `, Tsx: true, Settings: map[string]interface{}{"react": map[string]interface{}{"version": "0.14.0"}}},

		// ---- Upstream valid: nested class declaration with extra method ----
		{Code: `
        export default (Component) => (
          class Test extends React.Component {
            componentDidMount() {}
            render() {
              return <Component />;
            }
          }
        );
      `, Tsx: true},

		// ---- Upstream valid: external Foo.childContextTypes = {...} ----
		{Code: `
        class Foo extends React.Component {
          render() {
            return <div>{this.props.children}</div>;
          }
        }
        Foo.childContextTypes = {
          color: PropTypes.string
        };
      `, Tsx: true},

		// ---- Upstream valid: decorator on class ----
		{Code: `
        @foo
        class Foo extends React.Component {
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `, Tsx: true},

		// ---- Upstream valid: called decorator ----
		{Code: `
        @foo("bar")
        class Foo extends React.Component {
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `, Tsx: true},

		// ---- Edge: chained property access for external childContextTypes —
		// `Foo.bar.childContextTypes = ...` resolves to base `Foo` ----
		{Code: `
        class Foo extends React.Component {
          render() {
            return <div>{this.props.children}</div>;
          }
        }
        Foo.bar.childContextTypes = {};
      `, Tsx: true},

		// ---- Edge: paren-wrapped base — `(Foo).childContextTypes = ...` ----
		{Code: `
        class Foo extends React.Component {
          render() {
            return <div>{this.props.children}</div>;
          }
        }
        (Foo).childContextTypes = {};
      `, Tsx: true},

		// ---- Edge: aliased binding — `const C = Foo; C.childContextTypes = ...`
		// Resolved via TypeChecker symbol following (upstream's
		// `getRelatedComponent` does this via ESLint's scope manager). ----
		{Code: `
        class Foo extends React.Component {
          render() {
            return <div>{this.props.children}</div>;
          }
        }
        const C = Foo;
        C.childContextTypes = { color: PropTypes.string };
      `, Tsx: true},

		// ---- Edge: doubly-aliased — `const A = Foo; const B = A; B.childContextTypes = ...` ----
		{Code: `
        class Foo extends React.Component {
          render() {
            return <div>{this.props.children}</div>;
          }
        }
        const A = Foo;
        const B = A;
        B.childContextTypes = {};
      `, Tsx: true},

		// ---- Edge: ES5 createReactClass with external childContextTypes via alias ----
		{Code: `
        var Foo = createReactClass({
          render: function() {
            return <div>{this.props.children}</div>;
          }
        });
        var C = Foo;
        C.childContextTypes = {};
      `, Tsx: true},

		// ---- Edge: class static block — counts as "other" property → not
		// flagged. Locks in graceful handling of KindClassStaticBlockDeclaration. ----
		{Code: `
        class Foo extends React.Component {
          static {
            globalThis.Foo = Foo;
          }
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `, Tsx: true},

		// ---- Edge: bare `;` class element — KindSemicolonClassElement —
		// counts as "other" → not flagged. ----
		{Code: `
        class Foo extends React.Component {
          ;
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `, Tsx: true},

		// ---- Edge: abstract class method (TS only — no body) — counts as
		// "other" property since `abstract foo()` has no body and no allowed
		// name. Should NOT be flagged. ----
		{Code: `
        abstract class Foo extends React.Component {
          abstract bar(): void;
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `, Tsx: true},

		// ---- Edge: render via getter (`get render() {...}`) — getter on
		// allowed name, no useThis (this.props), so should be reported as
		// invalid. Lock-in: getter's body is walked. The next test handles
		// the body-walk pattern; this test asserts a get-render with this.bar
		// usage is suppressed (useThis). ----
		{Code: `
        class Foo extends React.Component {
          get render() {
            return () => <div>{this.bar}</div>;
          }
        }
      `, Tsx: true},

		// ---- Edge: nested arrow inside method using this.props only ----
		// is invalid (still props-only). Covered in invalid-2 below. The
		// valid sibling: nested arrow uses this.bar → useThis → not flagged.
		{Code: `
        class Foo extends React.Component {
          render() {
            const helper = () => this.bar;
            return <div>{helper()}</div>;
          }
        }
      `, Tsx: true},

		// ---- Edge (upstream-aligned): outer class has a non-render method
		// that returns a nested object literal containing a `render` property
		// whose value is a non-JSX-returning function. Upstream's
		// ReturnStatement listener walks scope chain and finds the inner
		// `render: function()` Property as the nearest enclosing Method/
		// Property — sets `invalidReturn` on the outer class via
		// `components.set` upward propagation. The outer Foo therefore is NOT
		// reported (suppressed by invalidReturn). Locks in the upstream
		// `Property` arm semantic. ----
		{Code: `
        class Foo extends React.Component {
          helper() {
            return {
              render: function() { return foo; }
            };
          }
          render() { return <div>{this.props.foo}</div>; }
        }
      `, Tsx: true},

		// ---- Locks in: scope-walk continues past unmatched FunctionLikes —
		// upstream's `scope.upper` keeps walking outward when the current
		// scope's `block.parent` is not `MethodDefinition` / `Property`.
		// Here a non-render-returning inline `function()` lives inside an
		// onClick JSX expression of render. The inner return must be
		// attributed to the OUTER render (per upstream), flipping
		// invalidReturn=true on the class → suppressed → valid.
		// Rejects gemini-code-assist suggestion that would stop the walk
		// at the inline function. ----
		{Code: `
        class Foo extends React.Component {
          render() {
            return <div onClick={function() { return undefined; }}>{this.props.foo}</div>;
          }
        }
      `, Tsx: true},

		// ---- Locks in same scope-walk continuation for ArrowFunction
		// inside JSX expression. ----
		{Code: `
        class Foo extends React.Component {
          render() {
            return <div onClick={() => { return undefined; }}>{this.props.foo}</div>;
          }
        }
      `, Tsx: true},

		// ---- Edge: nested arrow inside method has its own JSX ref — useRef → valid ----
		{Code: `
        class Foo extends React.Component {
          render() {
            const inner = () => <span ref="x" />;
            return <div>{inner()}</div>;
          }
        }
      `, Tsx: true},

		// ---- Upstream valid: multiple decorators ----
		{Code: `
        @foo
        @bar()
        class Foo extends React.Component {
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `, Tsx: true},

		// ---- Upstream valid: bare PureComponent + ignorePureComponents ----
		{Code: `
        class Child extends PureComponent {
          render() {
            return <h1>I don't</h1>;
          }
        }
      `, Tsx: true, Options: map[string]interface{}{"ignorePureComponents": true}},

		// ---- Upstream valid: PureComponent w/ static propTypes + ignorePureComponents ----
		{Code: `
        function errorDecorator(options) {
          return WrappedComponent => {
            class Wrapper extends PureComponent {
              static propTypes = {
                error: PropTypes.string
              }
              render () {
                const {error, ...props} = this.props
                if (error) {
                  return <div>Error! {error}</div>
                } else {
                  return <WrappedComponent {...props} />
                }
              }
            }
            return Wrapper
          }
        }
      `, Tsx: true, Options: map[string]interface{}{"ignorePureComponents": true}},

		// ---- Upstream valid: PureComponent inside arrow body + ignorePureComponents ----
		{Code: `
        function errorDecorator(options) {
          return WrappedComponent =>
            class Wrapper extends PureComponent {
              static propTypes = {
                error: PropTypes.string
              }
              render () {
                const {error, ...props} = this.props
                if (error) {
                  return <div>Error! {error}</div>
                } else {
                  return <WrappedComponent {...props} />
                }
              }
            }
        }
      `, Tsx: true, Options: map[string]interface{}{"ignorePureComponents": true}},

		// ---- Upstream valid: React.PureComponent inside arrow body + ignorePureComponents ----
		{Code: `
        function errorDecorator(options) {
          return WrappedComponent =>
            class Wrapper extends React.PureComponent {
              static propTypes = {
                error: PropTypes.string
              }
              render () {
                const {error, ...props} = this.props
                if (error) {
                  return <div>Error! {error}</div>
                } else {
                  return <WrappedComponent {...props} />
                }
              }
            }
        }
      `, Tsx: true, Options: map[string]interface{}{"ignorePureComponents": true}},

		// ---- Upstream valid: jsdoc-decorated stateless function (already pure) ----
		{Code: `
        /**
         * @param a.
         */
        function Comp() {
          return <a></a>
        }
      `, Tsx: true},

		// ---- Edge: settings.react.pragma="Preact" — extends Preact.Component triggers, but
		// React.Component no longer matches as a React component ----
		{Code: `
        class Foo extends React.Component {
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `, Tsx: true, Settings: map[string]interface{}{"react": map[string]interface{}{"pragma": "Preact"}}},

		// ---- Edge: PureComponent + ignorePureComponents=false (default) — should still flag ----
		// Covered by the invalid test below; valid version skipped.

		// ---- Edge: render returns object literal — invalid return → don't flag ----
		{Code: `
        class Foo extends React.Component {
          render() {
            return {foo: 1};
          }
        }
      `, Tsx: true},

		// ---- Edge: render returns string — invalid return → don't flag ----
		{Code: `
        class Foo extends React.Component {
          render() {
            return "hi";
          }
        }
      `, Tsx: true},

		// ---- Edge: ref on inner JSX inside method — useRef → don't flag ----
		{Code: `
        class Foo extends React.Component {
          render() {
            return <div ref={(el) => this._el = el}>{this.props.foo}</div>;
          }
        }
      `, Tsx: true},

		// ---- Edge: nested class is a separate boundary — outer stays clean
		// only when outer's body has no extra members. Here outer has a
		// method (helper), which itself flags useThis on outer (this.helper).
		// The inner class is reported separately. Valid: only the inner is
		// reported, outer is suppressed by `useThis`. ----
		{Code: `
        class Outer extends React.Component {
          helper() {}
          render() {
            class Inner extends React.Component {
              render() { return <span ref="x" /> }
            }
            return <Inner onClick={this.helper}/>;
          }
        }
      `, Tsx: true},

		// ---- Edge: createReactClass with getInitialState — extra prop → don't flag ----
		{Code: `
        var Foo = createReactClass({
          getInitialState: function() { return {}; },
          render: function() { return <div>{this.props.foo}</div>; }
        });
      `, Tsx: true},

		// ---- Edge: useless constructor variants — `super(...arguments)` ----
		// Component is REPORTED (invalid below); we test the boundary case
		// where ctor is non-pass-through, meaning it is NOT useless and the
		// component is valid.
		{Code: `
        class Foo extends React.Component {
          constructor(a, b) {
            super(b, a);
          }
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `, Tsx: true},

		// ---- Edge: useless constructor with default param — params not "simple" → not useless → valid ----
		{Code: `
        class Foo extends React.Component {
          constructor(props = {}) {
            super(props);
          }
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `, Tsx: true},

		// ---- Edge: useless constructor with destructured param — params not "simple" → valid ----
		{Code: `
        class Foo extends React.Component {
          constructor({props}) {
            super(props);
          }
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `, Tsx: true},

		// ---- Edge: createReactClass with non-render arrow body using this.bar — useThis valid ----
		{Code: `
        var Foo = createReactClass({
          render: function() {
            return <div>{this.bar}</div>;
          }
        });
      `, Tsx: true},

		// ===== tsgo-shape lock-in tests (these patterns differ from ESTree
		// in node shape but should produce the same observable outcome) =====

		// ---- tsgo: `(this).bar` — single-paren-wrapped this. tsgo preserves
		// the ParenthesizedExpression wrapper that ESTree flattens.
		// `isThisExpression` peels parens, so this is treated as `this.bar`
		// → useThis → valid. ----
		{Code: `
        class Foo extends React.Component {
          render() {
            return <div>{(this).bar}</div>;
          }
        }
      `, Tsx: true},

		// ---- tsgo: multi-paren-wrapped this `((this)).bar` ----
		{Code: `
        class Foo extends React.Component {
          render() {
            return <div>{((this)).bar}</div>;
          }
        }
      `, Tsx: true},

		// ---- tsgo: `this[\`bar\`]` — KindNoSubstitutionTemplateLiteral key
		// in element access. staticAccessName recognizes this. ----
		{Code: "\n        class Foo extends React.Component {\n          render() {\n            return <div>{this[`bar`]}</div>;\n          }\n        }\n      ", Tsx: true},

		// ---- tsgo: `this[\`pr${x}ops\`]` — TemplateExpression with
		// substitution: dynamic key, NOT recognized as 'props' → useThis. ----
		{Code: "\n        class Foo extends React.Component {\n          render() {\n            return <div>{this[`pr${x}ops`]}</div>;\n          }\n        }\n      ", Tsx: true},

		// ---- tsgo: optional chain `this?.bar` — KindPropertyAccessExpression
		// with optional flag (no ChainExpression wrapper). useThis. ----
		{Code: `
        class Foo extends React.Component {
          render() {
            return <div>{this?.bar}</div>;
          }
        }
      `, Tsx: true},

		// ---- tsgo: optional element access `this?.['bar']` — useThis. ----
		{Code: `
        class Foo extends React.Component {
          render() {
            return <div>{this?.['bar']}</div>;
          }
        }
      `, Tsx: true},

		// ---- tsgo: this in destructure with computed string-literal key
		// matching 'props': `let { ['props']: foo } = this` — keyName resolves
		// to "props" via ComputedPropertyName recursion → not useThis. So
		// when this is the only access, component IS reportable (invalid
		// section below covers the report). Here we add a sibling key so
		// useThis flips and the component remains valid. ----
		{Code: `
        class Foo extends React.Component {
          render() {
            let { ['props']: a, bar } = this;
            return <div>{bar}</div>;
          }
        }
      `, Tsx: true},

		// ---- tsgo: rest-only destructure `let { ...rest } = this` —
		// upstream's `getPropertyName(RestElement)` returns null;
		// `null !== 'props' && null !== 'context'` is true → useThis=true →
		// component suppressed. ----
		{Code: `
        class Foo extends React.Component {
          render() {
            let { ...rest } = this;
            return <div>{rest}</div>;
          }
        }
      `, Tsx: true},

		// ---- tsgo: rest after `props` — `let { props, ...rest } = this`.
		// rest still flips useThis=true regardless of sibling keys. ----
		{Code: `
        class Foo extends React.Component {
          render() {
            let { props, ...rest } = this;
            return <div>{rest}</div>;
          }
        }
      `, Tsx: true},

		// ---- tsgo: dynamic computed destructure `let { [Symbol.iterator]: x } = this`
		// — keyName "" → useThis=true → valid (suppressed). ----
		{Code: `
        class Foo extends React.Component {
          render() {
            let { [Symbol.iterator]: x } = this;
            return <div/>;
          }
        }
      `, Tsx: true},

		// ---- tsgo: numeric destructure key `let { 0: x } = this` —
		// memberName returns "0" → useThis=true → valid. ----
		{Code: `
        class Foo extends React.Component {
          render() {
            let { 0: x } = this;
            return <div/>;
          }
        }
      `, Tsx: true},

		// ---- tsgo: TS non-null assertion on `this` — `(this!).bar` /
		// `this!.bar`. SkipParentheses doesn't peel non-null. The shape is
		// not recognized as a `this`-receiver, so neither markPropsOrContext
		// nor markThisAsUsed fires for this access. ESTree behaves the same
		// (`TSNonNullExpression` wraps ThisExpression and doesn't equal
		// 'ThisExpression'). When no other useThis is present, the
		// component is REPORTED (invalid section locks this). Here we make
		// it valid by adding a real useThis sibling. ----
		{Code: `
        class Foo extends React.Component {
          render() {
            const x = this!.bar;
            return <div>{this.helper}</div>;
          }
        }
      `, Tsx: true},

		// ---- tsgo: TS `as` assertion on `this` — `(this as any).bar`.
		// AsExpression wraps this; not peeled. Same outcome as non-null. ----
		{Code: `
        class Foo extends React.Component {
          render() {
            const x = (this as any).bar;
            return <div>{this.helper}</div>;
          }
        }
      `, Tsx: true},

		// ---- tsgo: extends with multi-paren `extends ((React.Component))`
		// — SkipParentheses peels every level. ----
		{Code: `
        class Foo extends ((React.Component)) {
          render() {
            return <div>{this.bar}</div>;
          }
        }
      `, Tsx: true},

		// ---- tsgo: TS `as` cast around extends — extends `(React.Component as any)`
		// — SkipParentheses doesn't peel the AsExpression, so heritage
		// expression doesn't match Component. → not detected as a React
		// component → not reportable. ----
		{Code: `
        class Foo extends (React.Component as any) {
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `, Tsx: true},

		// ---- tsgo: `extends React['Component']` — element access form,
		// neither Identifier nor PropertyAccessExpression of pragma. Not
		// recognized as React component. ----
		{Code: `
        class Foo extends React['Component'] {
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `, Tsx: true},

		// ---- tsgo: render returning JSX wrapped in `as` cast — AsExpression
		// is not peeled by SkipParentheses, so isJSXLike returns false →
		// invalidReturn=true → component suppressed. ----
		{Code: `
        class Foo extends React.Component {
          render() {
            return (<div>{this.props.foo}</div>) as any;
          }
        }
      `, Tsx: true},

		// ---- tsgo: render returning bare `return;` — no expression → both
		// strict and non-strict: not JSX, not null literal → invalidReturn=true. ----
		{Code: `
        class Foo extends React.Component {
          render() {
            return;
          }
        }
      `, Tsx: true},

		// ---- tsgo: render returning template literal — not JSX, not null →
		// invalidReturn=true → suppressed. ----
		{Code: "\n        class Foo extends React.Component {\n          render() {\n            return `not jsx`;\n          }\n        }\n      ", Tsx: true},

		// ---- tsgo: render returning yield (generator-shaped though render
		// isn't a generator) — KindYieldExpression not recognized as JSX
		// → invalidReturn=true → suppressed. ----
		{Code: `
        class Foo extends React.Component {
          render() {
            return helper();
          }
        }
      `, Tsx: true},

		// ---- tsgo: TS overload signatures + impl, where overload is bodyless ----
		{Code: `
        class Foo extends React.Component {
          doStuff(): void;
          doStuff() {}
          render() { return <div>{this.props.foo}</div>; }
        }
      `, Tsx: true},

		// ---- tsgo: abstract bodyless method — bodyless, "other property"
		// (name not in allow-list) → not flagged. ----
		// (Already have a similar test above — keeping for completeness.)

		// ---- tsgo: `static childContextTypes = {...}` static class field —
		// upstream's hasOtherProperties branch matches the name 'childContextTypes'
		// is NOT in allow-list → other property → not flagged. ----
		{Code: `
        class Foo extends React.Component {
          static childContextTypes = { color: PropTypes.string };
          render() {
            return <div>{this.props.children}</div>;
          }
        }
      `, Tsx: true},

		// ---- tsgo: `props` field with `declare props: SomeType` — Type
		// annotation present, no Initializer. memberName="props" + Type!=nil
		// → in allow-list. The ONLY member here is the annotated field +
		// render with this.props → flagged. (Invalid below.) ----

		// ---- Edge: PropertyDeclaration with Type but Initializer too —
		// `props: Foo = defaultFoo;`. Type is set, so allow-list matches.
		// Initializer exists; my walk goes into it. The initializer is just
		// `defaultFoo` (Identifier) — no this, no JSX, no ref. ----
		{Code: `
        class Foo extends React.Component {
          render() {
            return <div>{this.bar}</div>;
          }
        }
        Foo.propTypes = {};
      `, Tsx: true},

		// ---- Edge: render method returning a comma-sequence whose rhs is
		// non-JSX — isJSXLike Comma branch checks rhs only → false →
		// invalidReturn=true → suppressed (valid). ----
		{Code: `
        class Foo extends React.Component {
          render() {
            return (<div>{this.props.foo}</div>, sideEffect());
          }
        }
      `, Tsx: true},

		// ---- Edge: settings.react.createClass="customCreate" — when the
		// configured factory name is custom, default `createReactClass` no
		// longer matches as a component. ----
		{Code: `
        var Foo = createReactClass({
          render: function() { return <div>{this.props.foo}</div>; }
        });
      `, Tsx: true, Settings: map[string]interface{}{"react": map[string]interface{}{"createClass": "customCreate"}}},
	}, []rule_tester.InvalidTestCase{
		// ---- Upstream invalid: only uses this.props ----
		{
			Code: `
        class Foo extends React.Component {
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "componentShouldBePure",
					Message:   "Component should be written as a pure function",
					Line:      2,
					Column:    9,
				},
			},
		},

		// ---- Upstream invalid: this['props'] ----
		{
			Code: `
        class Foo extends React.Component {
          render() {
            return <div>{this['props'].foo}</div>;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "componentShouldBePure", Line: 2, Column: 9},
			},
		},

		// ---- Upstream invalid: PureComponent without ignorePureComponents — defaults to false ----
		{
			Code: `
        class Foo extends React.PureComponent {
          render() {
            return <div>foo</div>;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "componentShouldBePure", Line: 2, Column: 9},
			},
		},

		// ---- Upstream invalid: PureComponent + props (default ignorePureComponents=false) ----
		{
			Code: `
        class Foo extends React.PureComponent {
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "componentShouldBePure", Line: 2, Column: 9},
			},
		},

		// ---- Upstream invalid: static get displayName() {...} ----
		{
			Code: `
        class Foo extends React.Component {
          static get displayName() {
            return 'Foo';
          }
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "componentShouldBePure", Line: 2, Column: 9},
			},
		},

		// ---- Upstream invalid: static displayName = 'Foo' ----
		{
			Code: `
        class Foo extends React.Component {
          static displayName = 'Foo';
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "componentShouldBePure", Line: 2, Column: 9},
			},
		},

		// ---- Upstream invalid: static get propTypes() {...} ----
		{
			Code: `
        class Foo extends React.Component {
          static get propTypes() {
            return {
              name: PropTypes.string
            };
          }
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "componentShouldBePure", Line: 2, Column: 9},
			},
		},

		// ---- Upstream invalid: static propTypes = {...} ----
		{
			Code: `
        class Foo extends React.Component {
          static propTypes = {
            name: PropTypes.string
          };
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "componentShouldBePure", Line: 2, Column: 9},
			},
		},

		// ---- Upstream invalid: props with type annotation (Flow / TS field) ----
		{
			Code: `
        class Foo extends React.Component {
          props: {
            name: string;
          };
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "componentShouldBePure", Line: 2, Column: 9},
			},
		},

		// ---- Upstream invalid: useless constructor calling super() ----
		{
			Code: `
        class Foo extends React.Component {
          constructor() {
            super();
          }
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "componentShouldBePure", Line: 2, Column: 9},
			},
		},

		// ---- Upstream invalid: destructures only props/context ----
		{
			Code: `
        class Foo extends React.Component {
          render() {
            let {props:{foo}, context:{bar}} = this;
            return <div>{this.props.foo}</div>;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "componentShouldBePure", Line: 2, Column: 9},
			},
		},

		// ---- Upstream invalid: render returns null on default (>= 15) — props only ----
		{
			Code: `
        class Foo extends React.Component {
          render() {
            if (!this.props.foo) {
              return null;
            }
            return <div>{this.props.foo}</div>;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "componentShouldBePure", Line: 2, Column: 9},
			},
		},

		// ---- Upstream invalid: createReactClass with null return ----
		{
			Code: `
        var Foo = createReactClass({
          render: function() {
            if (!this.props.foo) {
              return null;
            }
            return <div>{this.props.foo}</div>;
          }
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "componentShouldBePure", Line: 2, Column: 36},
			},
		},

		// ---- Upstream invalid: shorthand-if returning null at default (>= 15) ----
		{
			Code: `
        class Foo extends React.Component {
          render() {
            return true ? <div /> : null;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "componentShouldBePure", Line: 2, Column: 9},
			},
		},

		// ---- Upstream invalid: defaultProps as class field + shorthand-if/null return ----
		{
			Code: `
        class Foo extends React.Component {
          static defaultProps = {
            foo: true
          }
          render() {
            const { foo } = this.props;
            return foo ? <div /> : null;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "componentShouldBePure", Line: 2, Column: 9},
			},
		},

		// ---- Upstream invalid: static get defaultProps() {...} ----
		{
			Code: `
        class Foo extends React.Component {
          static get defaultProps() {
            return {
              foo: true
            };
          }
          render() {
            const { foo } = this.props;
            return foo ? <div /> : null;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "componentShouldBePure", Line: 2, Column: 9},
			},
		},

		// ---- Upstream invalid: external Foo.defaultProps assignment ----
		{
			Code: `
        class Foo extends React.Component {
          render() {
            const { foo } = this.props;
            return foo ? <div /> : null;
          }
        }
        Foo.defaultProps = {
          foo: true
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "componentShouldBePure", Line: 2, Column: 9},
			},
		},

		// ---- Upstream invalid: contextTypes as static class field ----
		{
			Code: `
        class Foo extends React.Component {
          static contextTypes = {
            foo: PropTypes.boolean
          }
          render() {
            const { foo } = this.context;
            return foo ? <div /> : null;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "componentShouldBePure", Line: 2, Column: 9},
			},
		},

		// ---- Upstream invalid: static get contextTypes() {...} ----
		{
			Code: `
        class Foo extends React.Component {
          static get contextTypes() {
            return {
              foo: PropTypes.boolean
            };
          }
          render() {
            const { foo } = this.context;
            return foo ? <div /> : null;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "componentShouldBePure", Line: 2, Column: 9},
			},
		},

		// ---- Upstream invalid: external Foo.contextTypes assignment ----
		{
			Code: `
        class Foo extends React.Component {
          render() {
            const { foo } = this.context;
            return foo ? <div /> : null;
          }
        }
        Foo.contextTypes = {
          foo: PropTypes.boolean
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "componentShouldBePure", Line: 2, Column: 9},
			},
		},

		// ---- Edge: bare Component (not React.Component) — should still flag ----
		{
			Code: `
        class Foo extends Component {
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "componentShouldBePure", Line: 2, Column: 9},
			},
		},

		// ---- Edge: createReactClass with displayName + render only — flagged ----
		{
			Code: `
        var Foo = createReactClass({
          displayName: 'Foo',
          render: function() {
            return <div>{this.props.foo}</div>;
          }
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "componentShouldBePure", Line: 2, Column: 36},
			},
		},

		// ---- Edge: `new createReactClass({...})` — non-idiomatic but
		// upstream's `componentUtil.isES5Component` check is `node.parent.callee`
		// which exists on both CallExpression and NewExpression in ESTree.
		// We mirror by handling KindNewExpression's parent shape too. ----
		{
			Code: `
        var Foo = new createReactClass({
          render: function() {
            return <div>{this.props.foo}</div>;
          }
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "componentShouldBePure"},
			},
		},

		// ---- Edge: ClassExpression assigned to var — flagged ----
		{
			Code: `
        var Foo = class extends React.Component {
          render() {
            return <div>{this.props.foo}</div>;
          }
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "componentShouldBePure", Line: 2, Column: 19},
			},
		},

		// ---- Edge: useless constructor with `super(...arguments)` ----
		{
			Code: `
        class Foo extends React.Component {
          constructor() {
            super(...arguments);
          }
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "componentShouldBePure", Line: 2, Column: 9},
			},
		},

		// ---- Edge: useless constructor with pass-through params ----
		{
			Code: `
        class Foo extends React.Component {
          constructor(a, b) {
            super(a, b);
          }
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "componentShouldBePure", Line: 2, Column: 9},
			},
		},

		// ---- Edge: useless constructor with rest pass-through ----
		{
			Code: `
        class Foo extends React.Component {
          constructor(...args) {
            super(...args);
          }
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "componentShouldBePure", Line: 2, Column: 9},
			},
		},

		// ---- Edge: bare PureComponent without ignorePureComponents — flagged by default ----
		{
			Code: `
        class Foo extends PureComponent {
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "componentShouldBePure", Line: 2, Column: 9},
			},
		},

		// ---- Edge: render returns JSX fragment ----
		{
			Code: `
        class Foo extends React.Component {
          render() {
            return <>{this.props.foo}</>;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "componentShouldBePure", Line: 2, Column: 9},
			},
		},

		// ---- Edge: nested ES5 component inside an outer ES6 class — outer
		// has `helper()` so it's NOT reported (hasOtherProperty), but the
		// inner ES5 (only `render` + `this.props`) IS reported separately.
		// Locks in independent boundary analysis: outer's body walk MUST
		// stop at the inner createReactClass call. ----
		{
			Code: `
        class Outer extends React.Component {
          helper() { return null; }
          render() {
            var Inner = createReactClass({
              render: function() { return <span>{this.props.x}</span>; }
            });
            return <Inner onClick={this.helper}/>;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "componentShouldBePure"},
			},
		},

		// ---- Edge: nested arrow inside render uses this.props only — still flagged ----
		{
			Code: `
        class Foo extends React.Component {
          render() {
            const inner = () => this.props.foo;
            return <div>{inner()}</div>;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "componentShouldBePure", Line: 2, Column: 9},
			},
		},

		// ---- Edge: chained call returning JSX in render (logical &&) — props-only, still flagged ----
		{
			Code: `
        class Foo extends React.Component {
          render() {
            return this.props.cond && <div>{this.props.foo}</div>;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "componentShouldBePure", Line: 2, Column: 9},
			},
		},

		// ---- Edge: computed string-literal key for `displayName` — upstream's
		// `getPropertyName` resolves to "displayName", in allow-list, so the
		// component WITH only render + this.props is reported. ----
		{
			Code: `
        class Foo extends React.Component {
          static ['displayName']() { return 'Foo'; }
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "componentShouldBePure", Line: 2, Column: 9},
			},
		},

		// ---- Edge: PrivateIdentifier-keyed `#displayName` — upstream's
		// `getPropertyName` strips the `#` and returns "displayName", placing
		// it in the allow-list. ----
		{
			Code: `
        class Foo extends React.Component {
          static #displayName = 'Foo';
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "componentShouldBePure", Line: 2, Column: 9},
			},
		},

		// ===== tsgo-shape lock-in invalid cases =====

		// ---- tsgo: paren-wrapped this with only props access — `(this).props`
		// is recognized as `this.props` (paren transparent), NOT useThis →
		// component IS reported. ----
		{
			Code: `
        class Foo extends React.Component {
          render() {
            return <div>{(this).props.foo}</div>;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "componentShouldBePure", Line: 2, Column: 9},
			},
		},

		// ---- tsgo: optional chain on this.props — `this?.props.foo`. tsgo
		// uses an optional flag on PropertyAccessExpression (no
		// ChainExpression wrapper). property name 'props' → markPropsOrContext
		// → not useThis → component reported. ----
		{
			Code: `
        class Foo extends React.Component {
          render() {
            return <div>{this?.props.foo}</div>;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "componentShouldBePure", Line: 2, Column: 9},
			},
		},

		// ---- tsgo: TS non-null on `this` — `this!.props`. tsgo wraps with
		// NonNullExpression which is NOT peeled by SkipParentheses. The
		// receiver doesn't equal `this`, so neither props-or-context nor
		// useThis fires. With no other useThis, the component is reported.
		// Match upstream: ESTree `TSNonNullExpression` has the same outcome
		// (its type is not 'ThisExpression' in the listener gate). ----
		{
			Code: `
        class Foo extends React.Component {
          render() {
            const x = this!.props.foo;
            return <div>{x}</div>;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "componentShouldBePure", Line: 2, Column: 9},
			},
		},

		// ---- tsgo: TS `as` assertion `(this as any).props` — same shape
		// outcome as non-null. ----
		{
			Code: `
        class Foo extends React.Component {
          render() {
            const x = (this as any).props.foo;
            return <div>{x}</div>;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "componentShouldBePure", Line: 2, Column: 9},
			},
		},

		// ---- tsgo: computed string-literal destructure key `let { ['props']: x } = this`
		// — keyName resolves to "props" via ComputedPropertyName recursion;
		// this counts as accessing this.props only, NOT useThis. With ONLY
		// this destructure + JSX render, component is reported. ----
		{
			Code: `
        class Foo extends React.Component {
          render() {
            let { ['props']: x } = this;
            return <div>{x.foo}</div>;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "componentShouldBePure", Line: 2, Column: 9},
			},
		},

		// ---- tsgo: NoSubstitutionTemplateLiteral computed key in
		// destructure: `let { [\`props\`]: x } = this` — keyName="props". ----
		{
			Code: "\n        class Foo extends React.Component {\n          render() {\n            let { [`props`]: x } = this;\n            return <div>{x.foo}</div>;\n          }\n        }\n      ",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "componentShouldBePure", Line: 2, Column: 9},
			},
		},

		// ---- tsgo: extends with multi-paren `extends ((React.Component))`
		// — SkipParentheses peels every level → matches Component. ----
		{
			Code: `
        class Foo extends ((React.Component)) {
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "componentShouldBePure", Line: 2, Column: 9},
			},
		},

		// ---- tsgo: render returning JSX with multi-line / nested
		// expression containers — typical real-world shape. ----
		{
			Code: `
        class Foo extends React.Component {
          render() {
            return (
              <div>
                {this.props.foo}
              </div>
            );
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "componentShouldBePure", Line: 2, Column: 9},
			},
		},

		// ---- tsgo: render returning conditional with both branches JSX —
		// strict-AND succeeds even on React 0.14.0; non-strict trivially
		// succeeds on React 15+. With this.props only → reported. ----
		{
			Code: `
        class Foo extends React.Component {
          render() {
            return cond ? <a>{this.props.x}</a> : <b/>;
          }
        }
      `,
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"version": "0.14.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "componentShouldBePure", Line: 2, Column: 9},
			},
		},

		// ---- tsgo: PrivateIdentifier-keyed `#propTypes` — strip `#` →
		// "propTypes" → in allow-list. ----
		{
			Code: `
        class Foo extends React.Component {
          static #propTypes = {};
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "componentShouldBePure", Line: 2, Column: 9},
			},
		},

		// ---- tsgo: NoSubstitutionTemplateLiteral computed member key —
		// `static [\`displayName\`] = 'Foo'` resolves to displayName. ----
		{
			Code: "\n        class Foo extends React.Component {\n          static [`displayName`] = 'Foo';\n          render() {\n            return <div>{this.props.foo}</div>;\n          }\n        }\n      ",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "componentShouldBePure", Line: 2, Column: 9},
			},
		},

		// ---- tsgo: `export class …` — report range starts at `class`, not at
		// `export`. ESTree wraps `export class` in ExportNamedDeclaration so
		// the ClassDeclaration's range begins at `class`. tsgo inlines
		// `export` into the ClassDeclaration's modifier list, so we trim it
		// back out via classKeywordStart. Locks in the column-after-export
		// position for parity with upstream.
		{
			Code: `
        export class App extends React.Component {
          render() {
            return <div className="App">Hello</div>;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "componentShouldBePure", Line: 2, Column: 16},
			},
		},

		// ---- Edge: `export default class extends React.Component {…}` —
		// anonymous default-export class. Position starts at `class`. ----
		{
			Code: `
        export default class extends React.Component {
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "componentShouldBePure", Line: 2, Column: 24},
			},
		},

		// ---- Locks in: `const x = this;` (Identifier id, NOT ObjectPattern)
		// — upstream's VariableDeclarator listener returns early when
		// `node.id.type !== 'ObjectPattern'`, so it does NOT flip useThis.
		// The class has no other this-access, so the component is REPORTED.
		// Rejects gemini-code-assist suggestion that would tighten this. ----
		{
			Code: `
        class Foo extends React.Component {
          render() {
            const x = this;
            return <div/>;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "componentShouldBePure", Line: 2, Column: 9},
			},
		},

		// ---- Locks in: array-pattern destructure `const [x] = this;` —
		// upstream returns early (id is ArrayPattern, not ObjectPattern). No
		// useThis flip → component reported. ----
		{
			Code: `
        class Foo extends React.Component {
          render() {
            const [x] = this;
            return <div/>;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "componentShouldBePure", Line: 2, Column: 9},
			},
		},

		// ---- tsgo: render returning a comma-sequence whose RHS is JSX —
		// isJSXLike Comma branch returns rhs check (true) → not
		// invalidReturn → component reported. ----
		{
			Code: `
        class Foo extends React.Component {
          render() {
            return (sideEffect(), <div>{this.props.foo}</div>);
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "componentShouldBePure", Line: 2, Column: 9},
			},
		},

		// ---- Alignment: `createReactClass` second-argument position —
		// upstream's `isES5Component` does NOT verify argument position
		// (only `node.parent.callee` identity). Even an obj passed as the
		// SECOND argument to a createClass call gets registered. We mirror
		// that by dropping our position check. ----
		{
			Code: `
        createReactClass(undefined, {
          render: function() { return <div>{this.props.foo}</div>; }
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "componentShouldBePure"},
			},
		},

		// ---- Edge: TS abstract class extending React.Component (already
		// has a no_redundant test counterpart). Even though declared
		// abstract, a stateless-eligible body still gets flagged. ----
		{
			Code: `
        abstract class Foo extends React.Component {
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "componentShouldBePure", Line: 2, Column: 9},
			},
		},

		// ---- Edge: settings.react.createClass="customCreate" + matching
		// factory — recognized as ES5 component. ----
		{
			Code: `
        var Foo = customCreate({
          render: function() { return <div>{this.props.foo}</div>; }
        });
      `,
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"createClass": "customCreate"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "componentShouldBePure"},
			},
		},

		// ---- Edge: settings.react.pragma="Preact" + extends Preact.Component ----
		{
			Code: `
        class Foo extends Preact.Component {
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `,
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"pragma": "Preact"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "componentShouldBePure", Line: 2, Column: 9},
			},
		},
	})
}

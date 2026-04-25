package no_unused_class_component_methods

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnusedClassComponentMethodsRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnusedClassComponentMethodsRule, []rule_tester.ValidTestCase{
		// ---- Upstream: SmockTestForTypeOfNullError — used method + used field ----
		{Code: `
        class SmockTestForTypeOfNullError extends React.Component {
          handleClick() {}
          foo;
          render() {
            let a;
            return <button disabled onClick={this.handleClick} foo={this.foo}>Text</button>;
          }
        }
      `, Tsx: true},

		// ---- Upstream: handler referenced in JSX event ----
		{Code: `
        class Foo extends React.Component {
          handleClick() {}
          render() {
            return <button onClick={this.handleClick}>Text</button>;
          }
        }
      `, Tsx: true},

		// ---- Upstream: createReactClass handler referenced in JSX event ----
		{Code: `
        var Foo = createReactClass({
          handleClick() {},
          render() {
            return <button onClick={this.handleClick}>Text</button>;
          },
        })
      `, Tsx: true},

		// ---- Upstream: method called from lifecycle method ----
		{Code: `
        class Foo extends React.Component {
          action() {}
          componentDidMount() {
            this.action();
          }
          render() {
            return null;
          }
        }
      `, Tsx: true},

		// ---- Upstream: createReactClass method called from lifecycle ----
		{Code: `
        var Foo = createReactClass({
          action() {},
          componentDidMount() {
            this.action();
          },
          render() {
            return null;
          },
        })
      `, Tsx: true},

		// ---- Upstream: method reference aliased locally ----
		{Code: `
        class Foo extends React.Component {
          action() {}
          componentDidMount() {
            const action = this.action;
            action();
          }
          render() {
            return null;
          }
        }
      `, Tsx: true},

		// ---- Upstream: method called and result assigned ----
		{Code: `
        class Foo extends React.Component {
          getValue() {}
          componentDidMount() {
            const action = this.getValue();
          }
          render() {
            return null;
          }
        }
      `, Tsx: true},

		// ---- Upstream: class-field arrow handler ----
		{Code: `
        class Foo extends React.Component {
          handleClick = () => {}
          render() {
            return <button onClick={this.handleClick}>Button</button>;
          }
        }
      `, Tsx: true},

		// ---- Upstream: method called within JSX expression ----
		{Code: `
        class Foo extends React.Component {
          renderContent() {}
          render() {
            return <div>{this.renderContent()}</div>;
          }
        }
      `, Tsx: true},

		// ---- Upstream: method called in nested JSX ----
		{Code: `
        class Foo extends React.Component {
          renderContent() {}
          render() {
            return (
              <div>
                <div>{this.renderContent()}</div>;
              </div>
            );
          }
        }
      `, Tsx: true},

		// ---- Upstream: class field object used in JSX ----
		{Code: `
        class Foo extends React.Component {
          property = {}
          render() {
            return <div property={this.property}>Example</div>;
          }
        }
      `, Tsx: true},

		// ---- Upstream: arrow field referenced from another arrow field ----
		{Code: `
        class Foo extends React.Component {
          action = () => {}
          anotherAction = () => {
            this.action();
          }
          render() {
            return <button onClick={this.anotherAction}>Example</button>;
          }
        }
      `, Tsx: true},

		// ---- Upstream: arrow expression-body referencing sibling arrow ----
		{Code: `
        class Foo extends React.Component {
          action = () => {}
          anotherAction = () => this.action()
          render() {
            return <button onClick={this.anotherAction}>Example</button>;
          }
        }
      `, Tsx: true},

		// ---- Upstream: chained field initializer invokes sibling arrow ----
		{Code: `
        class Foo extends React.Component {
          getValue = () => {}
          value = this.getValue()
          render() {
            return this.value;
          }
        }
      `, Tsx: true},

		// ---- Upstream: non-Component class is skipped entirely ----
		{Code: `
        class Foo {
          action = () => {}
          anotherAction = () => this.action()
        }
      `, Tsx: true},

		// ---- Upstream: async arrow method reference ----
		{Code: `
        class Foo extends React.Component {
          action = async () => {}
          render() {
            return <button onClick={this.action}>Click</button>;
          }
        }
      `, Tsx: true},

		// ---- Upstream: async method referenced via nested arrow ----
		{Code: `
        class Foo extends React.Component {
          async action() {
            console.log('error');
          }
          render() {
            return <button onClick={() => this.action()}>Click</button>;
          }
        }
      `, Tsx: true},

		// ---- Upstream: generator method referenced via nested arrow ----
		{Code: `
        class Foo extends React.Component {
          * action() {
            console.log('error');
          }
          render() {
            return <button onClick={() => this.action()}>Click</button>;
          }
        }
      `, Tsx: true},

		// ---- Upstream: async generator method referenced via nested arrow ----
		{Code: `
        class Foo extends React.Component {
          async * action() {
            console.log('error');
          }
          render() {
            return <button onClick={() => this.action()}>Click</button>;
          }
        }
      `, Tsx: true},

		// ---- Upstream: function-expression class field referenced via nested arrow ----
		{Code: `
        class Foo extends React.Component {
          action = function() {
            console.log('error');
          }
          render() {
            return <button onClick={() => this.action()}>Click</button>;
          }
        }
      `, Tsx: true},

		// ---- Upstream: this.X = ... in constructor (defined + used) ----
		{Code: `
        class ClassAssignPropertyInMethodTest extends React.Component {
          constructor() {
            this.foo = 3;;
          }
          render() {
            return <SomeComponent foo={this.foo} />;
          }
        }
      `, Tsx: true},

		// ---- Upstream: declared class property without initializer ----
		{Code: `
        class ClassPropertyTest extends React.Component {
          foo;
          render() {
            return <SomeComponent foo={this.foo} />;
          }
        }
      `, Tsx: true},

		// ---- Upstream: class property with initializer ----
		{Code: `
        class ClassPropertyTest extends React.Component {
          foo = a;
          render() {
            return <SomeComponent foo={this.foo} />;
          }
        }
      `, Tsx: true},

		// ---- Upstream: computed string-literal key with initializer referenced via element access ----
		{Code: `
        class Foo extends React.Component {
          ['foo'] = a;
          render() {
            return <SomeComponent foo={this['foo']} />;
          }
        }
      `, Tsx: true},

		// ---- Upstream: computed string-literal key without initializer, referenced via element access ----
		{Code: `
        class Foo extends React.Component {
          ['foo'];
          render() {
            return <SomeComponent foo={this['foo']} />;
          }
        }
      `, Tsx: true},

		// ---- Upstream: computed template-literal key, referenced via element access ----
		{Code: "\n        class ClassComputedTemplatePropertyTest extends React.Component {\n          [`foo`] = a;\n          render() {\n            return <SomeComponent foo={this[`foo`]} />;\n          }\n        }\n      ", Tsx: true},

		// ---- Upstream: state class field is a lifecycle name (never reported) ----
		{Code: `
        class ClassComputedTemplatePropertyTest extends React.Component {
          state = {}
          render() {
            return <div />;
          }
        }
      `, Tsx: true},

		// ---- Upstream: computed string-literal method name, referenced via plain member access ----
		{Code: `
        class ClassLiteralComputedMemberTest extends React.Component {
          ['foo']() {}
          render() {
            return <SomeComponent foo={this.foo} />;
          }
        }
      `, Tsx: true},

		// ---- Upstream: computed template-literal method name, referenced via plain member access ----
		{Code: "\n        class ClassComputedTemplateMemberTest extends React.Component {\n          [`foo`]() {}\n          render() {\n            return <SomeComponent foo={this.foo} />;\n          }\n        }\n      ", Tsx: true},

		// ---- Upstream: method referenced via a bare `this.foo;` statement ----
		{Code: `
        class ClassUseAssignTest extends React.Component {
          foo() {}
          render() {
            this.foo;
            return <SomeComponent />;
          }
        }
      `, Tsx: true},

		// ---- Upstream: method referenced via shorthand destructuring of `this` ----
		{Code: `
        class ClassUseAssignTest extends React.Component {
          foo() {}
          render() {
            const { foo } = this;
            return <SomeComponent />;
          }
        }
      `, Tsx: true},

		// ---- Upstream: same pattern, different name — destructuring from `this` ----
		{Code: `
        class ClassUseDestructuringTest extends React.Component {
          foo() {}
          render() {
            const { foo } = this;
            return <SomeComponent />;
          }
        }
      `, Tsx: true},

		// ---- Upstream: destructuring with explicit string key alias ----
		{Code: `
        class ClassUseDestructuringTest extends React.Component {
          ['foo']() {}
          render() {
            const { 'foo': bar } = this;
            return <SomeComponent />;
          }
        }
      `, Tsx: true},

		// ---- Upstream: fully-computed method key ([foo]) is not tracked ----
		{Code: `
        class ClassComputedMemberTest extends React.Component {
          [foo]() {}
          render() {
            return <SomeComponent />;
          }
        }
      `, Tsx: true},

		// ---- Upstream: every canonical lifecycle method ----
		{Code: `
        class ClassWithLifecyleTest extends React.Component {
          constructor(props) {
            super(props);
          }
          static getDerivedStateFromProps() {}
          componentWillMount() {}
          UNSAFE_componentWillMount() {}
          componentDidMount() {}
          componentWillReceiveProps() {}
          UNSAFE_componentWillReceiveProps() {}
          shouldComponentUpdate() {}
          componentWillUpdate() {}
          UNSAFE_componentWillUpdate() {}
          static getSnapshotBeforeUpdate() {}
          componentDidUpdate() {}
          componentDidCatch() {}
          componentWillUnmount() {}
          getChildContext() {}
          render() {
            return <SomeComponent />;
          }
        }
      `, Tsx: true},

		// ---- Upstream: every canonical ES5 lifecycle method on createReactClass ----
		{Code: `
        var ClassWithLifecyleTest = createReactClass({
          mixins: [],
          constructor(props) {
          },
          getDefaultProps() {
            return {}
          },
          getInitialState: function() {
            return {x: 0};
          },
          componentWillMount() {},
          UNSAFE_componentWillMount() {},
          componentDidMount() {},
          componentWillReceiveProps() {},
          UNSAFE_componentWillReceiveProps() {},
          shouldComponentUpdate() {},
          componentWillUpdate() {},
          UNSAFE_componentWillUpdate() {},
          componentDidUpdate() {},
          componentDidCatch() {},
          componentWillUnmount() {},
          getChildContext() {},
          render() {
            return <SomeComponent />;
          },
        })
      `, Tsx: true},

		// ---- rslint: non-React class is ignored regardless of member status ----
		{Code: `
        class NotAComponent {
          handleClick() {}
          render() { return null; }
        }
      `, Tsx: true},

		// ---- rslint: parenthesized `this` still matches (ESTree flattens, tsgo preserves) ----
		{Code: `
        class Foo extends React.Component {
          handleClick() {}
          render() {
            return <button onClick={(this).handleClick}>Text</button>;
          }
        }
      `, Tsx: true},

		// ---- rslint: TS non-null `this!.X` should count as a use ----
		{Code: `
        class Foo extends React.Component {
          handleClick() {}
          render() {
            return <button onClick={this!.handleClick}>Text</button>;
          }
        }
      `, Tsx: true},

		// ---- rslint: TS `as` cast on `this` still counts as a use ----
		{Code: `
        class Foo extends React.Component {
          handleClick() {}
          render() {
            return <button onClick={(this as any).handleClick}>Text</button>;
          }
        }
      `, Tsx: true},

		// ---- rslint: PropertyAccess LHS with inner `this.X.Y = z` does not define X ----
		// Only direct `this.X = …` (LHS is member of `this`) counts as a definition.
		// The nested access marks X as USED, not defined.
		{Code: `
        class Foo extends React.Component {
          foo = {}
          bar() {
            this.foo.x = 1;
          }
          componentDidMount() {
            this.bar();
          }
          render() { return null; }
        }
      `, Tsx: true},

		// ---- rslint: method referenced from inside a nested normal function
		// (lexical `this` does NOT apply, but upstream still counts it) ----
		{Code: `
        class Foo extends React.Component {
          action() {}
          componentDidMount() {
            (function() { this.action(); }).call(this);
          }
          render() { return null; }
        }
      `, Tsx: true},

		// ---- rslint: class-in-class boundary — the inner (non-Component)
		// class is visited as its own class body; its `this.x` must NOT leak
		// into the outer classInfo ----
		{Code: `
        class Foo extends React.Component {
          handleClick() {}
          componentDidMount() {
            class Inner { x() { this.y; } }
          }
          render() {
            return <button onClick={this.handleClick} />;
          }
        }
      `, Tsx: true},

		// ---- rslint: element access with numeric literal key (parity with
		// upstream isKeyLiteralLike for number literals) ----
		{Code: `
        class Foo extends React.Component {
          [0]() {}
          render() {
            return <SomeComponent foo={this[0]} />;
          }
        }
      `, Tsx: true},

		// ---- rslint: destructuring with numeric-literal explicit key ----
		{Code: `
        class Foo extends React.Component {
          [0]() {}
          render() {
            const { 0: first } = this;
            return <SomeComponent foo={first} />;
          }
        }
      `, Tsx: true},

		// ---- rslint: ES5 shorthand property initialized to a variable ----
		{Code: `
        var Foo = createReactClass({
          data: {},
          render() {
            return <SomeComponent data={this.data} />;
          },
        })
      `, Tsx: true},

		// ---- rslint: get/set accessor pair — setter references a property
		// defined via the same class ----
		{Code: `
        class Foo extends React.Component {
          _x = 0
          get x() { return this._x; }
          set x(v) { this._x = v; }
          render() { return <SomeComponent x={this.x} />; }
        }
      `, Tsx: true},

		// ---- rslint: extends React.PureComponent (second pragma class name) ----
		{Code: `
        class Foo extends React.PureComponent {
          handleClick() {}
          render() {
            return <button onClick={this.handleClick}>Text</button>;
          }
        }
      `, Tsx: true},

		// ---- rslint: bare `Component` base (no pragma prefix, reactutil allows) ----
		{Code: `
        class Foo extends Component {
          handleClick() {}
          render() {
            return <button onClick={this.handleClick}>Text</button>;
          }
        }
      `, Tsx: true},

		// ---- rslint: bare `PureComponent` base ----
		{Code: `
        class Foo extends PureComponent {
          handleClick() {}
          render() {
            return <button onClick={this.handleClick}>Text</button>;
          }
        }
      `, Tsx: true},

		// ---- rslint: constructor defines via `this.X = …`, render uses via
		// dotted access — mixed definition paths converge on the same name ----
		{Code: `
        class Foo extends React.Component {
          constructor(props) {
            super(props);
            this.foo = 1;
            this.bar = 2;
          }
          render() {
            return <SomeComponent foo={this.foo} bar={this.bar} />;
          }
        }
      `, Tsx: true},

		// ---- rslint: usage via element access with string-literal key —
		// `this['handleClick']()` must mark `handleClick` as used ----
		{Code: `
        class Foo extends React.Component {
          handleClick() {}
          componentDidMount() {
            this['handleClick']();
          }
          render() { return null; }
        }
      `, Tsx: true},

		// ---- rslint: usage via element access with template-literal key ----
		{Code: "\n        class Foo extends React.Component {\n          handleClick() {}\n          componentDidMount() {\n            this[`handleClick`]();\n          }\n          render() { return null; }\n        }\n      ", Tsx: true},

		// ---- rslint: computed boolean-literal method key is recognized by
		// both the definition side (computed `[true]`) and the element-access
		// side (`this[true]`). Upstream treats `String(true) === 'true'`. ----
		{Code: `
        class Foo extends React.Component {
          [true]() {}
          render() {
            return <SomeComponent foo={this[true]} />;
          }
        }
      `, Tsx: true},

		// ---- rslint: computed null-literal method key + access parity ----
		{Code: `
        class Foo extends React.Component {
          [null]() {}
          render() {
            return <SomeComponent foo={this[null]} />;
          }
        }
      `, Tsx: true},

		// ---- rslint: optional chain `this?.X` must still mark X as used
		// (tsgo encodes optional chain as a flag on PropertyAccessExpression —
		// no ChainExpression wrapper to unwrap) ----
		{Code: `
        class Foo extends React.Component {
          handleClick() {}
          render() {
            const fn = this?.handleClick;
            return <button onClick={fn}>Text</button>;
          }
        }
      `, Tsx: true},

		// ---- rslint: optional element-access `this?.['X']` must also count ----
		{Code: `
        class Foo extends React.Component {
          handleClick() {}
          render() {
            const fn = this?.['handleClick'];
            return <button onClick={fn}>Text</button>;
          }
        }
      `, Tsx: true},

		// ---- rslint: nested object destructuring `{ foo: { bar } } = this`
		// — upstream only marks TOP-LEVEL keys as used (matches the filter
		// `prop.type === 'Property' && isKeyLiteralLike(prop, prop.key)` +
		// `addUsedProperty(prop.key)`). The nested `bar` is NOT a direct
		// property of `this` — it would be `this.foo.bar`. So `foo` is used,
		// and `bar` is irrelevant to the outer class. ----
		{Code: `
        class Foo extends React.Component {
          foo = { bar: 1 }
          render() {
            const { foo: { bar } } = this;
            return <SomeComponent bar={bar} />;
          }
        }
      `, Tsx: true},

		// ---- rslint: rest pattern `{ foo, ...rest } = this` — rest is not a
		// Property in ESTree; upstream filter excludes it. `foo` is still
		// picked up. ----
		{Code: `
        class Foo extends React.Component {
          foo() {}
          render() {
            const { foo, ...rest } = this;
            return <SomeComponent rest={rest} />;
          }
        }
      `, Tsx: true},

		// ---- rslint: createReactClass with getter/setter shorthand ----
		{Code: `
        var Foo = createReactClass({
          get value() { return this._v; },
          set value(v) { this._v = v; },
          render() {
            return <SomeComponent v={this.value} />;
          },
        })
      `, Tsx: true},

		// ---- rslint: createReactClass method referenced via destructuring
		// from `this` ----
		{Code: `
        var Foo = createReactClass({
          action() {},
          componentDidMount() {
            const { action } = this;
            action();
          },
          render() { return null; },
        })
      `, Tsx: true},

		// ---- rslint: JSX spread attribute `<C {...this}>` — upstream's
		// MemberExpression handler only fires on dotted access, not spread.
		// This is NOT treated as "all members used" — but if no decl is
		// declared either, it stays valid trivially. Tests the no-crash
		// property on spread shapes. ----
		{Code: `
        class Foo extends React.Component {
          render() {
            return <SomeComponent {...this} />;
          }
        }
      `, Tsx: true},

		// ---- rslint: declared method + conditional rendering inside render ----
		{Code: `
        class Foo extends React.Component {
          handleClick() {}
          renderContent(flag) {
            return flag ? <button onClick={this.handleClick} /> : null;
          }
          render() {
            return this.renderContent(true);
          }
        }
      `, Tsx: true},

		// ---- rslint: declared method used inside a nested arrow inside
		// render, through multiple JSX layers + conditional expression ----
		{Code: `
        class Foo extends React.Component {
          items = [1, 2, 3]
          render() {
            return (
              <ul>
                {this.items.map((i) => (
                  <li key={i}>{i}</li>
                ))}
              </ul>
            );
          }
        }
      `, Tsx: true},

		// ---- rslint: static method's body contains `this.X` — should NOT
		// count as a use because static members are skipped entirely
		// (upstream's inStatic gate). Instance `foo` stays used from render. ----
		{Code: `
        class Foo extends React.Component {
          foo() {}
          static init() {
            // body is skipped — this pattern doesn't mark foo as used.
          }
          render() {
            return <SomeComponent foo={this.foo} />;
          }
        }
      `, Tsx: true},

		// ---- rslint: abstract / declare class member with no body — walker
		// must not crash when traversing a member whose Body is nil ----
		{Code: `
        abstract class Foo extends React.Component {
          abstract action(): void;
          render() {
            return <button onClick={this.action} />;
          }
        }
      `, Tsx: true},

		// ---- rslint: overloaded method signatures — tsgo emits multiple
		// MethodDeclaration members for overloads; the one with body is the
		// implementation, others are signature-only (body nil). Must not
		// crash and must not double-report. ----
		{Code: `
        class Foo extends React.Component {
          action(x: string): void;
          action(x: number): void;
          action(x: any): void { console.log(x); }
          render() {
            return <button onClick={() => this.action(1)} />;
          }
        }
      `, Tsx: true},

		// ---- rslint: `this.state = {…}` in constructor — `state` is
		// ES6_LIFECYCLE and always filtered, even when it's only assigned ----
		{Code: `
        class Foo extends React.Component {
          constructor(props) {
            super(props);
            this.state = { x: 0 };
          }
          render() { return <div>{this.state.x}</div>; }
        }
      `, Tsx: true},

		// ---- rslint: `defaultProps` / `propTypes` as static fields — static
		// is skipped entirely so they're never reported regardless of name ----
		{Code: `
        class Foo extends React.Component {
          static defaultProps = { x: 0 };
          static propTypes = {};
          render() { return <div />; }
        }
      `, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// ---- Upstream: non-standard lifecycle method on class ----
		{
			Code: `
        class Foo extends React.Component {
          getDerivedStateFromProps() {}
          render() {
            return <div>Example</div>;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedWithClass", Message: `Unused method or property "getDerivedStateFromProps" of class "Foo"`, Line: 3, Column: 11},
			},
		},

		// ---- Upstream: unused class property with empty-object initializer ----
		{
			Code: `
        class Foo extends React.Component {
          property = {}
          render() {
            return <div>Example</div>;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedWithClass", Message: `Unused method or property "property" of class "Foo"`, Line: 3, Column: 11},
			},
		},

		// ---- Upstream: unused method on class ----
		{
			Code: `
        class Foo extends React.Component {
          handleClick() {}
          render() {
            return null;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedWithClass", Message: `Unused method or property "handleClick" of class "Foo"`, Line: 3, Column: 11},
			},
		},

		// ---- Upstream: unused method on createReactClass ----
		{
			Code: `
        var Foo = createReactClass({
          handleClick() {},
          render() {
            return null;
          },
        })
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unused", Message: `Unused method or property "handleClick"`, Line: 3, Column: 11},
			},
		},

		// ---- Upstream: unused property (numeric value) on createReactClass ----
		{
			Code: `
        var Foo = createReactClass({
          a: 3,
          render() {
            return null;
          },
        })
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unused", Message: `Unused method or property "a"`, Line: 3, Column: 11},
			},
		},

		// ---- Upstream: multiple unused methods reported in source order ----
		{
			Code: `
        class Foo extends React.Component {
          handleScroll() {}
          handleClick() {}
          render() {
            return null;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedWithClass", Message: `Unused method or property "handleScroll" of class "Foo"`, Line: 3, Column: 11},
				{MessageId: "unusedWithClass", Message: `Unused method or property "handleClick" of class "Foo"`, Line: 4, Column: 11},
			},
		},

		// ---- Upstream: unused class-field arrow ----
		{
			Code: `
        class Foo extends React.Component {
          handleClick = () => {}
          render() {
            return null;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedWithClass", Message: `Unused method or property "handleClick" of class "Foo"`, Line: 3, Column: 11},
			},
		},

		// ---- Upstream: unused async arrow class field ----
		{
			Code: `
        class Foo extends React.Component {
          action = async () => {}
          render() {
            return null;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedWithClass", Message: `Unused method or property "action" of class "Foo"`, Line: 3, Column: 11},
			},
		},

		// ---- Upstream: unused async method — position is the identifier, after `async ` ----
		{
			Code: `
        class Foo extends React.Component {
          async action() {
            console.log('error');
          }
          render() {
            return null;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedWithClass", Message: `Unused method or property "action" of class "Foo"`, Line: 3, Column: 17},
			},
		},

		// ---- Upstream: unused generator method — position is the identifier, after `* ` ----
		{
			Code: `
        class Foo extends React.Component {
          * action() {
            console.log('error');
          }
          render() {
            return null;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedWithClass", Message: `Unused method or property "action" of class "Foo"`, Line: 3, Column: 13},
			},
		},

		// ---- Upstream: unused async generator method — position is the identifier, after `async * ` ----
		{
			Code: `
        class Foo extends React.Component {
          async * action() {
            console.log('error');
          }
          render() {
            return null;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedWithClass", Message: `Unused method or property "action" of class "Foo"`, Line: 3, Column: 19},
			},
		},

		// ---- Upstream: getInitialState on ES6 class is not an ES6 lifecycle name → reported ----
		{
			Code: `
        class Foo extends React.Component {
          getInitialState() {}
          render() {
            return null;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedWithClass", Message: `Unused method or property "getInitialState" of class "Foo"`, Line: 3, Column: 11},
			},
		},

		// ---- Upstream: unused `function` class field ----
		{
			Code: `
        class Foo extends React.Component {
          action = function() {
            console.log('error');
          }
          render() {
            return null;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedWithClass", Message: `Unused method or property "action" of class "Foo"`, Line: 3, Column: 11},
			},
		},

		// ---- Upstream: `this.foo = …` in constructor reported when unused; position is the identifier ----
		{
			Code: `
         class ClassAssignPropertyInMethodTest extends React.Component {
           constructor() {
             this.foo = 3;
           }
           render() {
             return <SomeComponent />;
           }
         }
       `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedWithClass", Message: `Unused method or property "foo" of class "ClassAssignPropertyInMethodTest"`, Line: 4, Column: 19},
			},
		},

		// ---- Upstream: bare class field declaration (no initializer) reported as unused ----
		{
			Code: `
         class Foo extends React.Component {
           foo;
           render() {
             return <SomeComponent />;
           }
         }
       `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedWithClass", Message: `Unused method or property "foo" of class "Foo"`, Line: 3, Column: 12},
			},
		},

		// ---- Upstream: class field with initializer reported as unused ----
		{
			Code: `
         class Foo extends React.Component {
           foo = a;
           render() {
             return <SomeComponent />;
           }
         }
       `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedWithClass", Message: `Unused method or property "foo" of class "Foo"`, Line: 3, Column: 12},
			},
		},

		// ---- Upstream: computed string-literal field (no initializer) reported at the literal ----
		{
			Code: `
         class Foo extends React.Component {
           ['foo'];
           render() {
             return <SomeComponent />;
           }
         }
       `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedWithClass", Message: `Unused method or property "foo" of class "Foo"`, Line: 3, Column: 13},
			},
		},

		// ---- Upstream: computed string-literal field (with initializer) reported at the literal ----
		{
			Code: `
         class Foo extends React.Component {
           ['foo'] = a;
           render() {
             return <SomeComponent />;
           }
         }
       `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedWithClass", Message: `Unused method or property "foo" of class "Foo"`, Line: 3, Column: 13},
			},
		},

		// ---- Upstream: access via a dynamic key (`this[foo]`) does NOT count as use ----
		{
			Code: `
         class Foo extends React.Component {
           foo = a;
           render() {
             return <SomeComponent foo={this[foo]} />;
           }
         }
       `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedWithClass", Message: `Unused method or property "foo" of class "Foo"`, Line: 3, Column: 12},
			},
		},

		// ---- Upstream: TS `private` field — position is the identifier, after `private ` ----
		{
			Code: `
         class Foo extends React.Component {
           private foo;
           render() {
             return <SomeComponent />;
           }
         }
       `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedWithClass", Message: `Unused method or property "foo" of class "Foo"`, Line: 3, Column: 20},
			},
		},

		// ---- Upstream: TS `private` method ----
		{
			Code: `
         class Foo extends React.Component {
           private foo() {}
           render() {
             return <SomeComponent />;
           }
         }
       `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedWithClass", Message: `Unused method or property "foo" of class "Foo"`, Line: 3, Column: 20},
			},
		},

		// ---- Upstream: TS `private` field with initializer ----
		{
			Code: `
         class Foo extends React.Component {
           private foo = 3;
           render() {
             return <SomeComponent />;
           }
         }
       `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedWithClass", Message: `Unused method or property "foo" of class "Foo"`, Line: 3, Column: 20},
			},
		},

		// ---- rslint: anonymous ClassExpression — no class name, uses `unused` messageId ----
		{
			Code: `
        const MakeIt = () => class extends React.Component {
          handleClick() {}
          render() { return null; }
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unused", Message: `Unused method or property "handleClick"`, Line: 3, Column: 11},
			},
		},

		// ---- rslint: `this.X` inside a non-constructor method still registers as a definition ----
		// Mirrors upstream's MemberExpression handler: any `this.X = …` inside any
		// (non-static) method adds X as a property. Not just constructor.
		{
			Code: `
        class Foo extends React.Component {
          notInLifecycle() {
            this.bar = 1;
          }
          componentDidMount() {
            this.notInLifecycle();
          }
          render() { return null; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedWithClass", Message: `Unused method or property "bar" of class "Foo"`, Line: 4, Column: 18},
			},
		},

		// ---- rslint: `this['X'] = …` in constructor — definition reported at
		// the inner string literal, not the `[` token ----
		{
			Code: `
        class Foo extends React.Component {
          constructor() {
            this['bar'] = 1;
          }
          render() { return null; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedWithClass", Message: `Unused method or property "bar" of class "Foo"`, Line: 4, Column: 18},
			},
		},

		// ---- rslint: static member must NOT mask a same-name instance use.
		// Static `foo` is skipped entirely; instance `foo` is still unused. ----
		{
			Code: `
        class Foo extends React.Component {
          static foo = 1;
          foo() {}
          render() { return null; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedWithClass", Message: `Unused method or property "foo" of class "Foo"`, Line: 4, Column: 11},
			},
		},

		// ---- rslint: fully-computed identifier key (`[foo]`) is not tracked
		// as a definition — but a literal-keyed sibling method still is ----
		{
			Code: `
        class Foo extends React.Component {
          [foo]() {}
          bar() {}
          render() { return null; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedWithClass", Message: `Unused method or property "bar" of class "Foo"`, Line: 4, Column: 11},
			},
		},

		// ---- rslint: compound assignment `this.X += …` is treated by upstream
		// as a DEFINITION (its AssignmentExpression check does not restrict
		// operator). So a class field `counter` plus a `this.counter += 1`
		// yields two "unused" reports — since neither assignment counts as a
		// read, nothing marks `counter` used. This is a known upstream quirk;
		// lock the behavior in so future refactors cannot silently flip it. ----
		{
			Code: `
        class Foo extends React.Component {
          counter = 0
          componentDidMount() {
            this.counter += 1;
          }
          render() { return null; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedWithClass", Message: `Unused method or property "counter" of class "Foo"`, Line: 3, Column: 11},
				{MessageId: "unusedWithClass", Message: `Unused method or property "counter" of class "Foo"`, Line: 5, Column: 18},
			},
		},

		// ---- rslint: shorthand ES5 ShorthandPropertyAssignment as a property ----
		{
			Code: `
        var Foo = createReactClass({
          data,
          render() {
            return null;
          },
        })
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unused", Message: `Unused method or property "data"`, Line: 3, Column: 11},
			},
		},

		// ---- rslint: extends PureComponent — still a React component; rule
		// must fire on unused methods ----
		{
			Code: `
        class Foo extends React.PureComponent {
          handleClick() {}
          render() {
            return null;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedWithClass", Message: `Unused method or property "handleClick" of class "Foo"`, Line: 3, Column: 11},
			},
		},

		// ---- rslint: setter body's `this._x = v` is treated as a DEFINITION
		// (any AssignmentExpression LHS counts). And `set x(v)` is itself a
		// declared accessor. `render`'s `this.x = v` also counts as another
		// definition for x. None of these names are ever READ → 3 reports
		// total. Locks in the "assignment-LHS = definition" semantics plus
		// its cross-member accumulation behavior. ----
		{
			Code: `
        class Foo extends React.Component {
          set x(v) {
            this._x = v;
          }
          render() {
            return <SomeComponent set={v => this.x = v} />;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedWithClass", Message: `Unused method or property "x" of class "Foo"`, Line: 3, Column: 15},
				{MessageId: "unusedWithClass", Message: `Unused method or property "_x" of class "Foo"`, Line: 4, Column: 18},
				{MessageId: "unusedWithClass", Message: `Unused method or property "x" of class "Foo"`, Line: 7, Column: 50},
			},
		},

		// ---- rslint: ES5 PropertyAssignment whose initializer is an arrow
		// function — the arrow body's `this.foo()` must mark `foo` as used.
		// Here `foo` IS used that way, but `helper` is unused. ----
		{
			Code: `
        var Foo = createReactClass({
          foo() {},
          helper: () => {},
          componentDidMount: function() { this.foo(); },
          render() { return null; },
        })
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unused", Message: `Unused method or property "helper"`, Line: 4, Column: 11},
			},
		},

		// ---- rslint: anonymous ClassExpression extending PureComponent;
		// uses `unused` messageId (no className) ----
		{
			Code: `
        export default class extends React.PureComponent {
          handleClick() {}
          render() { return null; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unused", Message: `Unused method or property "handleClick"`, Line: 3, Column: 11},
			},
		},

		// ---- rslint: multi-line method report position is the IDENTIFIER
		// (not the method opener). Verifies Line/Column precision on a
		// member whose name spans across leading modifiers / whitespace. ----
		{
			Code: `
        class Foo extends React.Component {
          async
          /* comment between modifier and name */
          handler() {}
          render() { return null; }
        }
      `,
			Tsx: true,
			// tsgo treats the bare `async` on its own line as a standalone
			// identifier-typed property declaration (not a modifier for the
			// following method). The "handler" method is therefore a plain
			// method on the next line; both `async` (the property) and
			// `handler` (the method) are unused.
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedWithClass", Message: `Unused method or property "async" of class "Foo"`, Line: 3, Column: 11},
				{MessageId: "unusedWithClass", Message: `Unused method or property "handler" of class "Foo"`, Line: 5, Column: 11},
			},
		},

		// ---- rslint: JSX spread `<C {...this}>` does NOT count as use for
		// declared methods. Upstream's MemberExpression handler only fires on
		// dotted access, not SpreadElement. ----
		{
			Code: `
        class Foo extends React.Component {
          handleClick() {}
          render() {
            return <SomeComponent {...this} />;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedWithClass", Message: `Unused method or property "handleClick" of class "Foo"`, Line: 3, Column: 11},
			},
		},

		// ---- rslint: string-literal property assignment on ES5 ----
		{
			Code: `
        var Foo = createReactClass({
          "foo-bar": 3,
          render() { return null; },
        })
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unused", Message: `Unused method or property "foo-bar"`, Line: 3, Column: 11},
			},
		},

		// ---- rslint: this.X used as a Call callee on a property that was
		// never declared — no report (nothing to report); but sibling declared
		// method `foo` stays unused ----
		{
			Code: `
        class Foo extends React.Component {
          foo() {}
          render() {
            this.nonExistent();
            return null;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedWithClass", Message: `Unused method or property "foo" of class "Foo"`, Line: 3, Column: 11},
			},
		},
	})
}

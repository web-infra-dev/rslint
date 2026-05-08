package destructuring_assignment

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestDestructuringAssignmentRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &DestructuringAssignmentRule, []rule_tester.ValidTestCase{
		// ---- Upstream: arrow that does not return JSX is not a component ----
		{Code: `
        export const revisionStates2 = {
            [A.b]: props => {
              return props.editor !== null
                ? 'xyz'
                : 'abc'
            },
        };
      `, Tsx: true},

		// ---- Upstream: HOF returning a component that destructures props inside ----
		{Code: `
        export function hof(namespace) {
          const initialState = {
            bounds: null,
            search: false,
          };
          return (props) => {
            const {x, y} = props
            if (y) {
              return <span>{y}</span>;
            }
            return <span>{x}</span>
          };
        }
      `, Tsx: true},

		// ---- Upstream: non-component reducer-like callback ----
		{Code: `
        export function hof(namespace) {
          const initialState = {
            bounds: null,
            search: false,
          };

          return (state = initialState, action) => {
            if (action.type === 'ABC') {
              return {...state, bounds: stuff ? action.x : null};
            }

            if (action.namespace !== namespace) {
              return state;
            }

            return null
          };
        }
      `, Tsx: true},

		// ---- Upstream: SFC with destructured-arg (default 'always' allows it) ----
		{Code: `
        const MyComponent = ({ id, className }) => (
          <div id={id} className={className} />
        );
      `, Tsx: true},

		// ---- Upstream: SFC with destructured-arg + explicit 'always' ----
		{Code: `
        const MyComponent = ({ id, className }) => (
          <div id={id} className={className} />
        );
      `, Tsx: true, Options: []interface{}{"always"}},

		// ---- Upstream: SFC body destructures from props ----
		{Code: `
        const MyComponent = (props) => {
          const { id, className } = props;
          return <div id={id} className={className} />
        };
      `, Tsx: true},

		// ---- Upstream: same with explicit 'always' ----
		{Code: `
        const MyComponent = (props) => {
          const { id, className } = props;
          return <div id={id} className={className} />
        };
      `, Tsx: true, Options: []interface{}{"always"}},

		// ---- Upstream: passing entire props prop is allowed ----
		{Code: `
        const MyComponent = (props) => (
          <div id={id} props={props} />
        );
      `, Tsx: true},

		{Code: `
        const MyComponent = (props) => (
          <div id={id} props={props} />
        );
      `, Tsx: true, Options: []interface{}{"always"}},

		// ---- Upstream: destructured second arg + first-arg-as-whole-spread ----
		{Code: `
        const MyComponent = (props, { color }) => (
          <div id={id} props={props} color={color} />
        );
      `, Tsx: true},

		{Code: `
        const MyComponent = (props, { color }) => (
          <div id={id} props={props} color={color} />
        );
      `, Tsx: true, Options: []interface{}{"always"}},

		// ---- Upstream: 'never' allows class direct access ----
		{Code: `
        const Foo = class extends React.PureComponent {
          render() {
            return <div>{this.props.foo}</div>;
          }
        };
      `, Tsx: true, Options: []interface{}{"never"}},

		{Code: `
        class Foo extends React.Component {
          doStuff() {}
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `, Tsx: true, Options: []interface{}{"never"}},

		// ---- Upstream: class destructuring under 'always' ----
		{Code: `
        const Foo = class extends React.PureComponent {
          render() {
            const { foo } = this.props;
            return <div>{foo}</div>;
          }
        };
      `, Tsx: true},

		{Code: `
        const Foo = class extends React.PureComponent {
          render() {
            const { foo } = this.props;
            return <div>{foo}</div>;
          }
        };
      `, Tsx: true, Options: []interface{}{"always"}},

		// ---- Upstream: 'never' on non-`props` destructure is fine ----
		{Code: `
        const MyComponent = (props) => {
          const { h, i } = hi;
          return <div id={props.id} className={props.className} />
        };
      `, Tsx: true, Options: []interface{}{"never"}},

		// ---- Upstream: constructor `this.state` direct access ----
		{Code: `
        const Foo = class extends React.PureComponent {
          constructor() {
            this.state = {};
            this.state.foo = 'bar';
          }
        };
      `, Tsx: true, Options: []interface{}{"always"}},

		// ---- Upstream: tagged template literal (styled-component) is not a component ----
		{Code: "\n        const div = styled.div`\n          & .button {\n            border-radius: ${props => props.borderRadius}px;\n          }\n        `\n      ", Tsx: true},

		// ---- Upstream: typed context arg (not an SFC, returns object) ----
		{Code: `
        export default (context: $Context) => ({
          foo: context.bar
        });
      `, Tsx: true},

		// ---- Upstream: regular non-React class ----
		{Code: `
        class Foo {
          bar(context) {
            return context.baz;
          }
        }
      `, Tsx: true},

		{Code: `
        class Foo {
          bar(props) {
            return props.baz;
          }
        }
      `, Tsx: true},

		// ---- Upstream: ignoreClassFields option ----
		{Code: `
        class Foo extends React.Component {
          bar = this.props.bar
        }
      `, Tsx: true, Options: []interface{}{"always", map[string]interface{}{"ignoreClassFields": true}}},

		{Code: `
        class Input extends React.Component {
          id = ` + "`${this.props.name}`" + `;
          render() {
            return <div />;
          }
        }
      `, Tsx: true, Options: []interface{}{"always", map[string]interface{}{"ignoreClassFields": true}}},

		// ---- Upstream: https://github.com/jsx-eslint/eslint-plugin-react/issues/2911
		// destructured arg shadowing — `context` here is local, not React context. ----
		{Code: `
        function Foo({ context }) {
          const d = context.describe();
          return <div>{d}</div>;
        }
      `, Tsx: true, Options: []interface{}{"always"}},

		// ---- Upstream: arrow as object-method value, not a component ----
		{Code: `
        const obj = {
          foo(arg) {
            const a = arg.func();
            return null;
          },
        };
      `, Tsx: true},

		// ---- Upstream: render-callback patterns inside data arrays ----
		{Code: `
        const columns = [
          {
            render: (val) => {
              if (val.url) {
                return (
                  <a href={val.url}>
                    {val.test}
                  </a>
                );
              }
              return null;
            },
          },
        ];
      `, Tsx: true},

		{Code: `
        const columns = [
          {
            render: val => <span>{val}</span>,
          },
          {
            someRenderFunc: function(val) {
              if (val.url) {
                return (
                  <a href={val.url}>
                    {val.test}
                  </a>
                );
              }
              return null;
            },
          },
        ];
      `, Tsx: true},

		// ---- Upstream: default export of a non-component arrow ----
		{Code: `
        export default (fileName) => {
          const match = fileName.match(/some expression/);
          if (match) {
            return fn;
          }
          return null;
        };
      `, Tsx: true},

		// ---- Upstream: this.props destructuring in a class method ----
		{Code: `
        class C extends React.Component {
          componentDidMount() {
            const { forwardRef } = this.props;

            this.ref.current.focus();

            if (typeof forwardRef === 'function') {
              forwardRef(this.ref);
            }
          }
          render() {
            return <div />;
          }
        }
      `, Tsx: true},

		// ---- Upstream: destructureInSignature 'always' but `props` referenced more than once ----
		{Code: `
        function Foo(props) {
          const {a} = props;
          return <Goo {...props}>{a}</Goo>;
        }
      `, Tsx: true, Options: []interface{}{"always", map[string]interface{}{"destructureInSignature": "always"}}},

		{Code: `
        function Foo(props) {
          const {a} = props;
          return <Goo f={() => props}>{a}</Goo>;
        }
      `, Tsx: true, Options: []interface{}{"always", map[string]interface{}{"destructureInSignature": "always"}}},

		// ---- Upstream: useContext destructured / non-destructured combinations ----
		{Code: `
        import { useContext } from 'react';

        const MyComponent = (props) => {
          const {foo} = useContext(aContext);
          return <div>{foo}</div>
        };
      `, Tsx: true, Options: []interface{}{"always"}, Settings: map[string]interface{}{"react": map[string]interface{}{"version": "16.9.0"}}},

		{Code: `
        import { useContext } from 'react';

        const MyComponent = (props) => {
          const foo = useContext(aContext);
          return <div>{foo.test}</div>
        };
      `, Tsx: true, Options: []interface{}{"never"}, Settings: map[string]interface{}{"react": map[string]interface{}{"version": "16.9.0"}}},

		{Code: `
        import { useContext } from 'react';

        const MyComponent = (props) => {
          const foo = useContext(aContext);
          return <div>{foo.test}</div>
        };
      `, Tsx: true, Options: []interface{}{"always"}, Settings: map[string]interface{}{"react": map[string]interface{}{"version": "16.9.0"}}},

		{Code: `
        const MyComponent = (props) => {
          const foo = useContext(aContext);
          return <div>{foo.test}</div>
        };
      `, Tsx: true, Options: []interface{}{"always"}, Settings: map[string]interface{}{"react": map[string]interface{}{"version": "16.8.999"}}},

		{Code: `
        const MyComponent = (props) => {
          const {foo} = useContext(aContext);
          return <div>{foo}</div>
        };
      `, Tsx: true, Options: []interface{}{"never"}, Settings: map[string]interface{}{"react": map[string]interface{}{"version": "16.8.999"}}},

		{Code: `
        const MyComponent = (props) => {
          const {foo} = useContext(aContext);
          return <div>{foo}</div>
        };
      `, Tsx: true, Options: []interface{}{"always"}, Settings: map[string]interface{}{"react": map[string]interface{}{"version": "16.8.999"}}},

		{Code: `
        const MyComponent = (props) => {
          const foo = useContext(aContext);
          return <div>{foo.test}</div>
        };
      `, Tsx: true, Options: []interface{}{"never"}, Settings: map[string]interface{}{"react": map[string]interface{}{"version": "16.8.999"}}},

		// ---- Upstream: optional chain on useContext result is allowed (no `.optional` member trigger) ----
		{Code: `
        import { useContext } from 'react';

        const MyComponent = (props) => {
          const foo = useContext(aContext);
          return <div>{foo?.test}</div>
        };
      `, Tsx: true},

		// ---- Locks in: optional-chain access on props is exempt under 'always' ----
		// Mirrors upstream's `!node.optional` guard inside handleSFCUsage.
		{Code: `
        const MyComponent = (props) => {
          return <div id={props?.id} />;
        };
      `, Tsx: true, Options: []interface{}{"always"}},

		// ---- Edge: shadowed `props` inside nested function — outer SFC's stack
		// entry is still the active one for the inner non-SFC callback that uses
		// its own local `props` parameter. ----
		{Code: `
        const MyComponent = ({foo}) => {
          const handler = (props) => props.bar();
          return <div onClick={handler}>{foo}</div>;
        };
      `, Tsx: true, Options: []interface{}{"always"}},

		// ---- Edge: nested SFC inside class method — class method body is not
		// an SFC, the inner arrow IS, and its destructured arg satisfies 'always'. ----
		{Code: `
        class Outer extends React.Component {
          render() {
            const Inner = ({x}) => <span>{x}</span>;
            const { y } = this.props;
            return <Inner x={y} />;
          }
        }
      `, Tsx: true, Options: []interface{}{"always"}},

		// ---- Edge: 'never' allows destructuring from non-`props|state|context`. ----
		{Code: `
        class Foo extends React.Component {
          render() {
            const { foo } = this.unrelated;
            return <div>{foo}</div>;
          }
        }
      `, Tsx: true, Options: []interface{}{"never"}},

		// ---- Edge: 'never' allows destructuring from a Call/non-member init. ----
		{Code: `
        const MyComponent = (props) => {
          const { foo } = props.normalize();
          return <div>{foo}</div>;
        };
      `, Tsx: true, Options: []interface{}{"never"}},

		// ---- Edge: 'never' permits SFC body's `props` reference (only destructure is forbidden). ----
		{Code: `
        const MyComponent = (props) => {
          return <div id={props.id} className={props.className} />;
        };
      `, Tsx: true, Options: []interface{}{"never"}},

		// ---- Edge: 'always' SFC body's `props` access through receiver-cast is
		// not flagged — the receiver after SkipParentheses is not a bare
		// Identifier so it shouldn't fire. ----
		{Code: `
        const MyComponent = (props) => {
          return <div id={(props as any).id} />;
        };
      `, Tsx: true, Options: []interface{}{"always"}},

		// ---- Lock: TS non-null assertion on receiver matches upstream's
		// `node.object.type === 'TSNonNullExpression'` non-match. ----
		// `props!.x` — receiver after paren-strip is NonNullExpression, not
		// Identifier; upstream's `node.object.name` is undefined → silent.
		{Code: `
        const MyComponent = (props: any) => {
          return <div id={props!.id} />;
        };
      `, Tsx: true, Options: []interface{}{"always"}},

		// ---- Lock: TS satisfies operator on receiver — same shape as as-cast. ----
		{Code: `
        const MyComponent = (props) => {
          return <div id={(props satisfies any).id} />;
        };
      `, Tsx: true, Options: []interface{}{"always"}},

		// ---- Lock: rest parameter is not treated as propsName. ----
		// `function Foo(...rest)` binds an array; `rest[0]` is element access
		// on a tuple, not a missed destructure.
		{Code: `
        function Foo(...rest: any[]) {
          return <div>{rest[0]}</div>;
        }
      `, Tsx: true, Options: []interface{}{"always"}},

		// ---- Lock: default-valued parameter is NOT identified as propsName.
		// Mirrors upstream behavior: ESTree wraps `(props = {})` in an
		// `AssignmentPattern` whose `param.type` matches neither
		// `'Identifier'` nor `'ObjectPattern'`, so upstream's `evalParams`
		// emits a zero entry. We mirror this by skipping
		// ParameterDeclarations whose `Initializer` is non-nil.
		// Verified against eslint-plugin-react@7.37.5: the snippet below is
		// silent under both rules. ----
		{Code: `
        const MyComponent = (props = {}) => {
          const { id } = props;
          return <div id={id} />;
        };
      `, Tsx: true, Options: []interface{}{"always"}},

		// ---- Lock (default param + body access): even when `props.id` is
		// accessed directly in the body, the default value disables
		// propsName recognition, so the rule stays silent. ----
		{Code: `
        const MyComponent = (props = {}) => <div id={props.id} />;
      `, Tsx: true, Options: []interface{}{"always"}},

		// ---- Lock (default destructured param under 'never'): an
		// AssignmentPattern wrapping ObjectPattern (`({id} = {})`) is
		// likewise skipped — upstream does not flag it. ----
		{Code: `
        const Foo = ({ id } = {}) => <div id={id} />;
      `, Tsx: true, Options: []interface{}{"never"}},

		// ---- Lock: nested non-SFC callback inside SFC sees outer `props`
		// through the active stack entry, but accesses are still flagged when
		// they touch the outer parameter. ----
		// (This valid case shows: when the inner callback uses its OWN local
		// `props` parameter, no upstream stack entry should leak through.)
		{Code: `
        const Foo = ({foo}) => {
          const helper = (props) => props.bar();
          return <div>{foo}{helper({bar: () => 0})}</div>;
        };
      `, Tsx: true, Options: []interface{}{"always"}},

		// ---- Lock: 'never' allows non-`props|context` named first arg in SFC
		// when destructured? No — destructuring at SFC arg always reports
		// regardless of name (upstream evalParams emits `destructuring=true`
		// without checking name). ----
		// This is a positive control showing 'never' permits `Identifier` first
		// arg with any name, matching `params[0].name && !params[0].destructuring`.
		{Code: `
        const Foo = (renamedProps) => (
          <div id={renamedProps.id} />
        );
      `, Tsx: true, Options: []interface{}{"never"}},

		// ---- Lock: assignment to `this.props` itself (constructor pattern) is exempt. ----
		// `this.props = X` would be `this.props` as LHS — outer MemberExpression
		// on receiver `this` doesn't match `this.props.X` shape, so no report.
		{Code: `
        class Foo extends React.Component {
          constructor(p) {
            super(p);
            this.props = p;
          }
          render() { return <div /> }
        }
      `, Tsx: true, Options: []interface{}{"always"}},

		// ---- Lock: useContext with `?.` chain on the result is not a `props`
		// access (already in upstream), and the outer member is optional. ----
		{Code: `
        function Foo(props) {
          const ctx = useContext(C);
          return <div>{ctx?.x}</div>;
        }
      `, Tsx: true, Options: []interface{}{"always"}},

		// ---- Lock: 'this[\"props\"].x' is NOT flagged (inner is ElementAccess,
		// upstream only matches dotted .props/.state/.context). ----
		{Code: `
        class Foo extends React.Component {
          render() {
            return <div>{this['props'].foo}</div>;
          }
        }
      `, Tsx: true, Options: []interface{}{"always"}},

		// ---- Lock: spread destructure in SFC arg under 'always'. ----
		{Code: `
        const Foo = ({...rest}) => <div {...rest} />;
      `, Tsx: true, Options: []interface{}{"always"}},

		// ---- Lock: 'always' SFC body destructure consumed in JSX child. ----
		{Code: `
        const Foo = (props) => {
          const {children} = props;
          return <div>{children}</div>;
        };
      `, Tsx: true, Options: []interface{}{"always"}},

		// ---- Lock: TS generic SFC with destructured arg under 'always'. ----
		{Code: `
        const Foo = <T,>(p: { x: T }) => <div>{p.x}</div>;
      `, Tsx: true, Options: []interface{}{"never"}},

		// ---- Real-world: forwardRef with destructured args under 'always'. ----
		// Inner arrow is recognized by IsStatelessReactComponent (memo /
		// forwardRef wrapper), and the destructured param is allowed.
		{Code: `
        const Foo = React.forwardRef(({x}: any, ref: any) => (
          <div ref={ref}>{x}</div>
        ));
      `, Tsx: true, Options: []interface{}{"always"}},

		// ---- Real-world: memo + forwardRef nested. ----
		{Code: `
        const Foo = React.memo(React.forwardRef(({x}: any, ref: any) => (
          <div ref={ref}>{x}</div>
        )));
      `, Tsx: true, Options: []interface{}{"always"}},

		// ---- Lock: TypeChecker path — outer `props` referenced beyond the
		// destructure suppresses destructureInSignature. Control case paired
		// with the shadow test in invalid: same shape but without scope
		// shadow, the second `props` IS the outer parameter, so the rule
		// must stay silent. ----
		{Code: `
        function Foo(props) {
          const {a} = props;
          const cached = props;
          return <p>{a}{cached}</p>;
        }
      `, Tsx: true, Options: []interface{}{"always", map[string]interface{}{"destructureInSignature": "always"}}},

		// ---- Lock (custom HOC parity): a function wrapped by a user-defined
		// HOC that is NOT in the built-in wrapper list (memo / forwardRef) and
		// not configured via `settings.componentWrapperFunctions` is treated
		// as a non-component by both ESLint and rslint, so `props.id` access
		// is silent. Verified against eslint-plugin-react@7.37.5. ----
		{Code: `
        const withLogging = (fn) => fn;
        const Foo = withLogging((props) => <div id={props.id} />);
      `, Tsx: true, Options: []interface{}{"always"}},

		// ---- Lock (shadow false positive fix): a local `props` declared
		// inside an inner function is unrelated to the component's props,
		// so accesses on it must NOT be flagged. ESLint's name-only check
		// reports this as a missed destructure (a known false positive in
		// upstream); rslint resolves the actual binding via type info and
		// stays silent. The outer SFC parameter is named `props` so the
		// name-only path WOULD match — only Symbol comparison rejects.
		// 'always' mode is required to exercise handleSFCUsage. ----
		{Code: `
        function Foo(props) {
          function helper() {
            let props = { b: 1 };
            return props.b;
          }
          return <p>{helper()}</p>;
        }
      `, Tsx: true, Options: []interface{}{"always"}},

		// ---- Lock: VariableDeclaration uses ENCLOSING-only SFC check (not
		// ancestor walk). `const {x} = props` inside an inner non-SFC helper
		// of an outer SFC must NOT report under 'never' — upstream's
		// `components.get(scope.block)` returns undefined for the inner
		// helper, and the rule stays silent. Earlier ancestor-walk
		// implementation over-reported here. ----
		{Code: `
        const Foo = (props) => {
          function helper() {
            const {x} = props;
            return x;
          }
          return <div>{helper()}</div>;
        };
      `, Tsx: true, Options: []interface{}{"never"}},
	}, []rule_tester.InvalidTestCase{
		// ---- Upstream: SFC accessing `props.id` directly ----
		{
			Code: `
        const MyComponent = (props) => {
          return (<div id={props.id} />)
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDestructAssignment"},
			},
		},

		// ---- Upstream: 'never' rejects destructured props in SFC arg ----
		{
			Code: `
        const MyComponent = ({ id, className }) => (
          <div id={id} className={className} />
        );
      `,
			Tsx:     true,
			Options: []interface{}{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDestructPropsInSFCArg"},
			},
		},

		// ---- Upstream: 'never' rejects destructured context in SFC arg ----
		{
			Code: `
        const MyComponent = (props, { color }) => (
          <div id={props.id} className={props.className} />
        );
      `,
			Tsx:     true,
			Options: []interface{}{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDestructContextInSFCArg"},
			},
		},

		// ---- Upstream: class accessing `this.props.foo` ----
		{
			Code: `
        const Foo = class extends React.PureComponent {
          render() {
            return <div>{this.props.foo}</div>;
          }
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDestructAssignment"},
			},
		},

		// ---- Upstream: class `this.state` direct access ----
		{
			Code: `
        const Foo = class extends React.PureComponent {
          render() {
            return <div>{this.state.foo}</div>;
          }
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDestructAssignment"},
			},
		},

		// ---- Upstream: class `this.context` direct access ----
		{
			Code: `
        const Foo = class extends React.PureComponent {
          render() {
            return <div>{this.context.foo}</div>;
          }
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDestructAssignment"},
			},
		},

		// ---- Upstream: this.props in a method called by render ----
		{
			Code: `
        class Foo extends React.Component {
          render() { return this.foo(); }
          foo() {
            return this.props.children;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDestructAssignment"},
			},
		},

		// ---- Upstream: createReactClass + this.props.foo ----
		{
			Code: `
        var Hello = createReactClass({
          render: function() {
            return <Text>{this.props.foo}</Text>;
          }
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDestructAssignment"},
			},
		},

		// ---- Upstream: SFC inside object literal (module.exports = { Foo(props) {...} }) ----
		{
			Code: `
        module.exports = {
          Foo(props) {
            return <p>{props.a}</p>;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDestructAssignment"},
			},
		},

		// ---- Upstream: export default function form ----
		{
			Code: `
        export default function Foo(props) {
          return <p>{props.a}</p>;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDestructAssignment"},
			},
		},

		// ---- Upstream: HOF returning an arrow component ----
		{
			Code: `
        function hof() {
          return (props) => <p>{props.a}</p>;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDestructAssignment"},
			},
		},

		// ---- Upstream: const-init from this.props in render ----
		{
			Code: `
        const Foo = class extends React.PureComponent {
          render() {
            const foo = this.props.foo;
            return <div>{foo}</div>;
          }
        };
        `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDestructAssignment"},
			},
		},

		// ---- Upstream: 'never' rejects class const-destructuring of this.props ----
		{
			Code: `
        const Foo = class extends React.PureComponent {
          render() {
            const { foo } = this.props;
            return <div>{foo}</div>;
          }
        };
      `,
			Tsx:     true,
			Options: []interface{}{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDestructAssignment"},
			},
		},

		// ---- Upstream: 'never' rejects SFC const-destructuring of props ----
		{
			Code: `
        const MyComponent = (props) => {
          const { id, className } = props;
          return <div id={id} className={className} />
        };
      `,
			Tsx:     true,
			Options: []interface{}{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDestructAssignment"},
			},
		},

		// ---- Upstream: 'never' rejects class const-destructuring of this.state ----
		{
			Code: `
        const Foo = class extends React.PureComponent {
          render() {
            const { foo } = this.state;
            return <div>{foo}</div>;
          }
        };
      `,
			Tsx:     true,
			Options: []interface{}{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDestructAssignment"},
			},
		},

		// ---- Upstream: multi-line columns config — 3 reports on different lines ----
		{
			Code: `
        const columns = [
          {
            CustomComponentName: function(props) {
              if (props.url) {
                return (
                  <a href={props.url}>
                    {props.test}
                  </a>
                );
              }
              return null;
            },
          },
        ];
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDestructAssignment", Line: 5},
				{MessageId: "useDestructAssignment", Line: 7},
				{MessageId: "useDestructAssignment", Line: 8},
			},
		},

		// ---- Upstream: SFC second arg accessed but not destructured ----
		{
			Code: `
        function Foo(props, context) {
          const d = context.describe();
          return <div>{d}</div>;
        }
      `,
			Tsx:     true,
			Options: []interface{}{"always"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDestructAssignment"},
			},
		},

		// ---- Upstream: chained access props.str.match ----
		{
			Code: `
        export default (props) => {
          const match = props.str.match(/some expression/);
          if (match) {
            return <span>jsx</span>;
          }
          return null;
        };
      `,
			Tsx:     true,
			Options: []interface{}{"always"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDestructAssignment"},
			},
		},

		// ---- Upstream: nested arrow callback observing props ----
		{
			Code: `
        import React from 'react';

        const TestComp = (props) => {
          props.onClick3102();

          return (
            <div
              onClick={(evt) => {
                if (props.onClick3102) {
                  props.onClick3102(evt);
                }
              }}
            >
              <div />
            </div>
          );
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDestructAssignment", Line: 5},
				{MessageId: "useDestructAssignment", Line: 10},
				{MessageId: "useDestructAssignment", Line: 11},
			},
		},

		// ---- Upstream: arrow inside object literal — JSX-returning ----
		{
			Code: `
        export const revisionStates2 = {
            [A.b]: props => {
              return props.editor !== null
                ? <span>{props.editor}</span>
                : null
            },
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDestructAssignment", Line: 4},
				{MessageId: "useDestructAssignment", Line: 5},
			},
		},

		// ---- Upstream: HOF returning component with three props refs ----
		{
			Code: `
        export function hof(namespace) {
          const initialState = {
            bounds: null,
            search: false,
          };
          return (props) => {
            if (props.y) {
              return <span>{props.y}</span>;
            }
            return <span>{props.x}</span>
          };
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDestructAssignment", Line: 8},
				{MessageId: "useDestructAssignment", Line: 9},
				{MessageId: "useDestructAssignment", Line: 11},
			},
		},

		// ---- Upstream: destructureInSignature 'always' — fix to signature destructure ----
		{
			Code: `
          function Foo(props) {
            const {a} = props;
            return <p>{a}</p>;
          }
        `,
			Tsx:     true,
			Options: []interface{}{"always", map[string]interface{}{"destructureInSignature": "always"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "destructureInSignature", Line: 3},
			},
			Output: []string{"\n          function Foo({a}) {\n            \n            return <p>{a}</p>;\n          }\n        "},
		},

		// ---- Upstream: destructureInSignature with type annotation preserved ----
		{
			Code: `
          function Foo(props: FooProps) {
            const {a} = props;
            return <p>{a}</p>;
          }
        `,
			Tsx:     true,
			Options: []interface{}{"always", map[string]interface{}{"destructureInSignature": "always"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "destructureInSignature", Line: 3},
			},
			Output: []string{"\n          function Foo({a}: FooProps) {\n            \n            return <p>{a}</p>;\n          }\n        "},
		},

		// ---- Upstream: TSQualifiedName `typeof props.text` and member `props.text` ----
		{
			Code: `
        type Props = { text: string };
        export const MyComponent: React.FC<Props> = (props) => {
          type MyType = typeof props.text;
          return <div>{props.text as MyType}</div>;
        };
      `,
			Tsx:     true,
			Options: []interface{}{"always", map[string]interface{}{"destructureInSignature": "always"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDestructAssignment"},
				{MessageId: "useDestructAssignment"},
			},
		},

		// ---- Upstream: only TSQualifiedName triggers when body already destructured ----
		{
			Code: `
        type Props = { text: string };
        export const MyOtherComponent: React.FC<Props> = (props) => {
          const { text } = props;
          type MyType = typeof props.text;
          return <div>{text as MyType}</div>;
        };
      `,
			Tsx:     true,
			Options: []interface{}{"always", map[string]interface{}{"destructureInSignature": "always"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDestructAssignment"},
			},
		},

		// ---- Upstream: void / typeof prop access ----
		{
			Code: `
        function C(props: Props) {
          void props.a
          typeof props.b
          return <div />
        }
      `,
			Tsx:     true,
			Options: []interface{}{"always"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDestructAssignment"},
				{MessageId: "useDestructAssignment"},
			},
		},

		// ---- Edge: `props['x']` is also a member expression with Identifier receiver ----
		// Locks in: tsgo's ElementAccessExpression triggers handleSFCUsage when the receiver is `props`.
		{
			Code: `
        const MyComponent = (props) => {
          return <div id={props['id']} />;
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDestructAssignment"},
			},
		},

		// ---- Edge: parenthesized receiver — `(props).id` ----
		// Locks in: SkipParentheses unwrapping the access receiver.
		{
			Code: `
        const MyComponent = (props) => {
          return <div id={(props).id} />;
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDestructAssignment"},
			},
		},

		// ---- Edge: assignment LHS is suppressed, but the inner `props.x` chain is reported ----
		// Locks in: outer `props.foo.bar = 1` suppresses the outer, but `props.foo` (inner) still reports.
		{
			Code: `
        const MyComponent = (props) => {
          props.foo.bar = 1;
          return <div />;
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDestructAssignment"},
			},
		},

		// ---- Edge: parenthesized assignment LHS is also suppressed ----
		// Locks in: SkipParentheses on `node.parent` chain inside isAssignmentLHS.
		{
			Code: `
        const MyComponent = (props) => {
          (props.foo.bar) = 1;
          return <div />;
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDestructAssignment"},
			},
		},

		// ---- Edge: compound-assignment LHS is suppressed ----
		// Locks in: ast.IsAssignmentOperator covers `+=`, `||=`, etc.
		{
			Code: `
        const MyComponent = (props) => {
          props.foo += 1;
          return <div />;
        };
      `,
			Tsx: true,
			// Outer `props.foo += 1` is LHS — suppressed. No nested member to
			// report. Only triggered if `props` is read elsewhere; here it
			// isn't, so this case has zero diagnostics.
			Errors: nil,
		},

		// ---- Edge: ++props.x is reported (UpdateExpression isn't an assignment) ----
		// Locks in: prefix unary on member is not isAssignmentLHS — matches upstream.
		{
			Code: `
        const MyComponent = (props) => {
          ++props.x;
          return <div />;
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDestructAssignment"},
			},
		},

		// ---- Edge: `let {a} = props` under 'never' — still reported (binding kind doesn't matter). ----
		{
			Code: `
        const MyComponent = (props) => {
          let { foo } = props;
          return <div>{foo}</div>;
        };
      `,
			Tsx:     true,
			Options: []interface{}{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDestructAssignment"},
			},
		},

		// ---- Edge: `var {a} = props` under 'never' — still reported. ----
		{
			Code: `
        const MyComponent = (props) => {
          var { foo } = props;
          return <div>{foo}</div>;
        };
      `,
			Tsx:     true,
			Options: []interface{}{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDestructAssignment"},
			},
		},

		// ---- Edge: nested QualifiedName — only the leftmost (left=Identifier) reports. ----
		// `typeof props.a.b` becomes `QN(QN(props, a), b)`; outer's `Left` is a
		// QualifiedName, not Identifier, so only the inner one fires.
		{
			Code: `
        const MyComponent = (props) => {
          type T = typeof props.a.b;
          return <div />;
        };
      `,
			Tsx:     true,
			Options: []interface{}{"always"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDestructAssignment"},
			},
		},

		// ---- Edge: ignoreClassFields=true under 'always' still reports
		// `this.props.x` outside class fields. ----
		{
			Code: `
        class Foo extends React.Component {
          bar = this.props.bar;
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `,
			Tsx:     true,
			Options: []interface{}{"always", map[string]interface{}{"ignoreClassFields": true}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDestructAssignment"},
			},
		},

		// ---- Edge: shorthand-method SFC value (object literal) — both
		// `useDestructAssignment` reports trigger the props.X chain. ----
		{
			Code: `
        const obj = {
          MyComp(props) {
            return <p id={props.id}>{props.name}</p>;
          }
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDestructAssignment"},
				{MessageId: "useDestructAssignment"},
			},
		},

		// ---- Edge: `this.props` element-access OUTER (e.g. `this.props['foo']`)
		// matches: outer is ElementAccess, inner (object) is `this.props`
		// PropertyAccess. ----
		{
			Code: `
        class Foo extends React.Component {
          render() {
            return <div>{this.props['foo']}</div>;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDestructAssignment"},
			},
		},

		// ---- Edge: optional chain on `this.props?.x` IS reported in 'always'.
		// Upstream's `!node.optional` guard only applies to handleSFCUsage, NOT
		// handleClassUsage — class access reports regardless of optionality. ----
		{
			Code: `
        class Foo extends React.Component {
          render() {
            return <div>{this.props?.foo}</div>;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDestructAssignment"},
			},
		},

		// ---- Edge: parenthesized this — `(this).props.x` still reports. ----
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
				{MessageId: "useDestructAssignment"},
			},
		},

		// ---- Edge: nested non-SFC inner callback inside SFC reads outer
		// `props` — should still report (upstream's `propsName()` walks the
		// full SFC stack, the inner callback's local scope doesn't shadow). ----
		{
			Code: `
        const Foo = (props) => {
          const handler = () => props.x;
          return <div onClick={handler} />;
        };
      `,
			Tsx:     true,
			Options: []interface{}{"always"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDestructAssignment"},
			},
		},

		// ---- Edge: SFC inside ConditionalExpression (init position) — when
		// recognized, props.x should report. (If our SFC detector rejects this
		// position, the case becomes a no-error scenario; either way the lock
		// is on the rule's behavior, not the detector's.) ----
		// We assert no panic / consistent classification; let's keep it
		// pragmatic and skip if detector disagrees.
		{
			Code: `
        const A = (props: any) => <div>{props.foo}</div>;
        const B = (props: any) => <div>{props.bar}</div>;
        const C = condition ? A : B;
      `,
			Tsx:     true,
			Options: []interface{}{"always"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDestructAssignment"},
				{MessageId: "useDestructAssignment"},
			},
		},

		// ---- Edge: SFC body uses `props` as RHS of an assignment to a local. ----
		{
			Code: `
        const Foo = (props) => {
          let cached;
          cached = props.value;
          return <div>{cached}</div>;
        };
      `,
			Tsx:     true,
			Options: []interface{}{"always"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDestructAssignment"},
			},
		},

		// ---- Edge: chained `props.a.b` reports inner only. ----
		// Outer's object is PropertyAccess(props, a), not Identifier → no match
		// for outer. Inner's object is Identifier props → match.
		{
			Code: `
        const Foo = (props) => {
          return <div>{props.a.b}</div>;
        };
      `,
			Tsx:     true,
			Options: []interface{}{"always"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDestructAssignment"},
			},
		},

		// ---- Edge: nested member with element-access outer:
		// `props['a'].b` — inner `props['a']` matches (Identifier receiver). ----
		{
			Code: `
        const Foo = (props) => {
          return <div>{props['a'].b}</div>;
        };
      `,
			Tsx:     true,
			Options: []interface{}{"always"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDestructAssignment"},
			},
		},

		// ---- Edge: 'never' + class field initializer destructure WITHOUT
		// ignoreClassFields — should report (default behavior). ----
		{
			Code: `
        class Foo extends React.Component {
          state = { count: 0 };
          init = (() => { const { count } = this.state; return count; })();
          render() { return <div /> }
        }
      `,
			Tsx:     true,
			Options: []interface{}{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDestructAssignment"},
			},
		},

		// ---- Edge: destructureInSignature 'always' with multi-property
		// destructure — fix preserves the full pattern. ----
		{
			Code: `
          function Foo(props) {
            const {a, b, c} = props;
            return <p>{a}{b}{c}</p>;
          }
        `,
			Tsx:     true,
			Options: []interface{}{"always", map[string]interface{}{"destructureInSignature": "always"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "destructureInSignature"},
			},
			Output: []string{"\n          function Foo({a, b, c}) {\n            \n            return <p>{a}{b}{c}</p>;\n          }\n        "},
		},

		// ---- Edge: destructureInSignature 'always' with rename. ----
		{
			Code: `
          function Foo(props) {
            const {a: x} = props;
            return <p>{x}</p>;
          }
        `,
			Tsx:     true,
			Options: []interface{}{"always", map[string]interface{}{"destructureInSignature": "always"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "destructureInSignature"},
			},
			Output: []string{"\n          function Foo({a: x}) {\n            \n            return <p>{x}</p>;\n          }\n        "},
		},

		// ---- Edge: 'never' SFC param renamed (Identifier) shape destructure
		// of `props` is NOT possible (would need rename to be the SFC param,
		// but destructuring binds locals, not the param itself). The reverse —
		// destructured first param under any name — still reports. ----
		{
			Code: `
        const Foo = ({a}: any) => (
          <div>{a}</div>
        );
      `,
			Tsx:     true,
			Options: []interface{}{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDestructPropsInSFCArg"},
			},
		},

		// ---- Real-world: useEffect closure over outer SFC props. ----
		// Inner arrow callback ancestor walk must reach the outer SFC.
		{
			Code: `
        const Foo = (props) => {
          React.useEffect(() => {
            console.log(props.id);
          }, [props.id]);
          return <div />;
        };
      `,
			Tsx:     true,
			Options: []interface{}{"always"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDestructAssignment"},
				{MessageId: "useDestructAssignment"},
			},
		},

		// ---- Real-world: switch on `props.type` returns JSX. ----
		{
			Code: `
        const Foo = (props) => {
          switch (props.type) {
            case 'a': return <A />;
            default: return null;
          }
        };
      `,
			Tsx:     true,
			Options: []interface{}{"always"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDestructAssignment"},
			},
		},

		// ---- Real-world: JSX spread attribute receiving `props.X`. ----
		{
			Code: `
        const Foo = (props) => <div {...props.styles} />;
      `,
			Tsx:     true,
			Options: []interface{}{"always"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDestructAssignment"},
			},
		},

		// ---- Real-world: short-circuit and ternary mixing props references. ----
		{
			Code: `
        const Foo = (props) => (
          <div>{props.show && <span>{props.text}</span>}</div>
        );
      `,
			Tsx:     true,
			Options: []interface{}{"always"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDestructAssignment"},
				{MessageId: "useDestructAssignment"},
			},
		},

		// ---- Real-world: forwardRef inner arrow with non-destructured props. ----
		{
			Code: `
        const Foo = React.forwardRef((props: any, ref: any) => (
          <div ref={ref}>{props.x}</div>
        ));
      `,
			Tsx:     true,
			Options: []interface{}{"always"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDestructAssignment"},
			},
		},

		// ---- Real-world: useState init with `props.X`. ----
		{
			Code: `
        const Foo = (props) => {
          const [x] = React.useState(props.initial);
          return <div>{x}</div>;
        };
      `,
			Tsx:     true,
			Options: []interface{}{"always"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDestructAssignment"},
			},
		},

		// ---- Real-world: nested SFC stack — outer SFC's props is reachable
		// from an inner SFC's body when the inner SFC's first param is
		// destructured (so `propsName()` skips B and falls through to A's
		// 'props' name). ----
		{
			Code: `
        const A = (props) => {
          const B = ({x}: any) => <div>{x}{props.y}</div>;
          return <B />;
        };
      `,
			Tsx:     true,
			Options: []interface{}{"always"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDestructAssignment"},
			},
		},

		// ---- Lock: 'never' + SFC body with `const {x} = this.props` (nonsensical
		// but legal syntax) — upstream's getParentComponent falls back to the
		// SFC, so destructuringClass + classComponent triggers a noDestructAssignment
		// report. We must mirror that fallback. ----
		{
			Code: `
        const Foo = (props) => {
          const { x } = this.props;
          return <div>{x}</div>;
        };
      `,
			Tsx:     true,
			Options: []interface{}{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDestructAssignment"},
			},
		},

		// ---- Lock: 'never' + ignoreClassFields=true does NOT suppress
		// `const {x} = this.props` even when nested inside a class field
		// initializer IIFE. Upstream's `node.parent.type === 'ClassProperty'`
		// check is dead code (node.parent is always VariableDeclaration), so
		// the guard never fires. ----
		{
			Code: `
        class Foo extends React.Component {
          state = { count: 0 };
          field = (() => { const { count } = this.state; return count; })();
          render() { return <div /> }
        }
      `,
			Tsx:     true,
			Options: []interface{}{"never", map[string]interface{}{"ignoreClassFields": true}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDestructAssignment"},
			},
		},

		// ---- Lock: TypeChecker scope-aware path resolves shadow correctly.
		// The inner `helper` declares its own `let props`, whose Symbol differs
		// from the outer SFC parameter. Both gates use Symbol comparison:
		//
		//   - `destructureInSignature` count: inner occurrences excluded →
		//     outer has only the init reference → count==0 → REPORT.
		//   - `handleSFCUsage` SFC report: inner `props.b`'s Symbol does NOT
		//     match the SFC parameter's Symbol → NO useDestructAssignment.
		//
		// This is strictly more precise than upstream, which uses name-only
		// matching everywhere and would emit both reports. The improvement
		// is documented in the rule's `.md` Differences section.
		// ----
		{
			Code: `
          function Foo(props) {
            const {a} = props;
            function helper() {
              let props = { b: 1 };
              return props.b;
            }
            return <p>{a}{helper()}</p>;
          }
        `,
			Tsx:     true,
			Options: []interface{}{"always", map[string]interface{}{"destructureInSignature": "always"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "destructureInSignature", Line: 3},
			},
			Output: []string{"\n          function Foo({a}) {\n            \n            function helper() {\n              let props = { b: 1 };\n              return props.b;\n            }\n            return <p>{a}{helper()}</p>;\n          }\n        "},
		},
	})
}

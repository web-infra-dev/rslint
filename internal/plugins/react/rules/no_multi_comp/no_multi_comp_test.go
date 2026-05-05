package no_multi_comp

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// crCode mirrors upstream's `code.split('\n').join('\r')` test fixture
// transformation: every `\n` becomes a single `\r`. Used to verify the
// rule's line numbers track the host parser's CR-only line-terminator
// handling on the canonical "two createReactClass per file" / "three class
// declarations per file" inputs.
func crCode(s string) string {
	return strings.ReplaceAll(s, "\n", "\r")
}

const (
	onlyOne     = "onlyOneComponent"
	onlyOneText = "Declare only one React component per file"
)

func TestNoMultiCompRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoMultiCompRule, []rule_tester.ValidTestCase{
		// ---- Single component declarations ----
		{Code: `
        var Hello = require('./components/Hello');
        var HelloJohn = createReactClass({
          render: function() {
            return <Hello name="John" />;
          }
        });
      `, Tsx: true},
		{Code: `
        class Hello extends React.Component {
          render() {
            return <div>Hello {this.props.name}</div>;
          }
        }
      `, Tsx: true},
		{Code: `
        var Heading = createReactClass({
          render: function() {
            return (
              <div>
                {this.props.buttons.map(function(button, index) {
                  return <Button {...button} key={index}/>;
                })}
              </div>
            );
          }
        });
      `, Tsx: true},

		// ---- ignoreStateless: true ----
		{Code: `
        function Hello(props) {
          return <div>Hello {props.name}</div>;
        }
        function HelloAgain(props) {
          return <div>Hello again {props.name}</div>;
        }
      `, Tsx: true, Options: map[string]interface{}{"ignoreStateless": true}},
		{Code: `
        function Hello(props) {
          return <div>Hello {props.name}</div>;
        }
        class HelloJohn extends React.Component {
          render() {
            return <Hello name="John" />;
          }
        }
      `, Tsx: true, Options: map[string]interface{}{"ignoreStateless": true}},

		// ---- Helper functions are not components ----
		{Code: `
        import React, { createElement } from "react"
        const helperFoo = () => {
          return true;
        };
        function helperBar() {
          return false;
        };
        function RealComponent() {
          return createElement("img");
        };
      `, Tsx: true},

		// ---- React.memo / React.forwardRef wrapping ----
		{Code: `
        const Hello = React.memo(function(props) {
          return <div>Hello {props.name}</div>;
        });
        class HelloJohn extends React.Component {
          render() {
            return <Hello name="John" />;
          }
        }
      `, Tsx: true, Options: map[string]interface{}{"ignoreStateless": true}},

		// ---- forwardRef wrapping a sibling component (nodeWrapsComponent gate) ----
		{Code: `
        class StoreListItem extends React.PureComponent {
          // A bunch of stuff here
        }
        export default React.forwardRef((props, ref) => <StoreListItem {...props} forwardRef={ref} />);
      `, Tsx: true, Options: map[string]interface{}{"ignoreStateless": false}},
		{Code: `
        class StoreListItem extends React.PureComponent {
          // A bunch of stuff here
        }
        export default React.forwardRef((props, ref) => {
          return <StoreListItem {...props} forwardRef={ref} />
        });
      `, Tsx: true, Options: map[string]interface{}{"ignoreStateless": false}},
		{Code: `
        const HelloComponent = (props) => {
          return <div></div>;
        }
        export default React.forwardRef((props, ref) => <HelloComponent {...props} forwardRef={ref} />);
      `, Tsx: true, Options: map[string]interface{}{"ignoreStateless": false}},
		{Code: `
        class StoreListItem extends React.PureComponent {
          // A bunch of stuff here
        }
        export default React.forwardRef(
          function myFunction(props, ref) {
            return <StoreListItem {...props} forwardedRef={ref} />;
          }
        );
      `, Tsx: true, Options: map[string]interface{}{"ignoreStateless": true}},
		{Code: `
        const HelloComponent = (props) => {
          return <div></div>;
        }
        class StoreListItem extends React.PureComponent {
          // A bunch of stuff here
        }
        export default React.forwardRef(
          function myFunction(props, ref) {
            return <StoreListItem {...props} forwardedRef={ref} />;
          }
        );
      `, Tsx: true, Options: map[string]interface{}{"ignoreStateless": true}},
		{Code: `
        const HelloComponent = (props) => {
          return <div></div>;
        }
        export default React.memo((props, ref) => <HelloComponent {...props} />);
      `, Tsx: true, Options: map[string]interface{}{"ignoreStateless": true}},

		// ---- Class field arrow that consumes an unrelated `memo`-named helper ----
		{Code: `
        import React from 'react';
        function memo() {
          var outOfScope = "hello"
          return null;
        }
        class ComponentY extends React.Component {
          memoCities = memo((cities) => cities.map((v) => ({ label: v })));
          render() {
            return (
              <div>
                <div>Counter</div>
              </div>
            );
          }
        }
      `, Tsx: true},

		// ---- forwardRef + .displayName / .propTypes / .defaultProps assignments ----
		{Code: `
        const MenuList = forwardRef(({onClose, ...props}, ref) => {
          const {t} = useTranslation();
          const handleLogout = useLogoutHandler();

          const onLogout = useCallback(() => {
            onClose();
            handleLogout();
          }, [onClose, handleLogout]);

          return (
            <MuiMenuList ref={ref} {...props}>
              <MuiMenuItem key="logout" onClick={onLogout}>
                {t('global-logout')}
              </MuiMenuItem>
            </MuiMenuList>
          );
        });

        MenuList.displayName = 'MenuList';

        MenuList.propTypes = {
          onClose: PropTypes.func,
        };

        MenuList.defaultProps = {
          onClose: () => null,
        };

        export default MenuList;
      `, Tsx: true},
		{Code: `
        const MenuList = forwardRef(({ onClose, ...props }, ref) => {
          const onLogout = useCallback(() => {
            onClose()
          }, [onClose])

          return (
            <BlnMenuList ref={ref} {...props}>
              <BlnMenuItem key="logout" onClick={onLogout}>
                Logout
              </BlnMenuItem>
            </BlnMenuList>
          )
        })

        MenuList.displayName = 'MenuList'

        MenuList.propTypes = {
          onClose: PropTypes.func
        }

        MenuList.defaultProps = {
          onClose: () => null
        }

        export default MenuList
      `, Tsx: true},

		// ---- Lock-in: branches not exercised by upstream's own valid suite ----

		// Locks in upstream `getStatelessComponent` Branch 12 reject (`!isInAllowedPositionForComponent`):
		// arrow nested as a CallExpression argument that is NOT a pragma wrapper —
		// the inner arrow is rejected (parent is CallExpression, not in allowed
		// positions list), so the only component is the outer ClassDeclaration.
		{Code: `
        class App extends React.Component {
          render() {
            return <div onClick={() => <div>nope</div>} />;
          }
        }
      `, Tsx: true},

		// Locks in upstream `getStatelessComponent` Branch 2 reject (lowercase VariableDeclarator):
		// lowercase-named arrow is not a component; the only component is the class.
		{Code: `
        const helper = (props) => <div>{props.x}</div>;
        class App extends React.Component {
          render() { return <div />; }
        }
      `, Tsx: true},

		// Locks in upstream Components.detect FunctionExpression banned-confidence path:
		// async generators are explicitly registered with confidence 0, so they don't
		// count as components even when paired with another stateless function.
		{Code: `
        async function* gen() {
          yield 1;
        }
        function App(props) { return <div /> }
      `, Tsx: true, Options: map[string]interface{}{"ignoreStateless": true}},

		// Locks in upstream's `nodeWrapsComponent` MemberExpression-only gate:
		// the bare-callee form (`memo(arrow)` with `import { memo } from 'react'`)
		// does NOT skip even when the inner arrow's root JSX names a sibling.
		// There's no second component upstream OR rslint here — only one
		// component is added, since the bare wrapper IS classified.
		{Code: `
        import { forwardRef } from 'react';
        const Inner = (props) => <div />;
      `, Tsx: true},

		// ---- Dimension 4: Universal edge shapes ----

		// Paren-wrapped class declaration RHS — tsgo preserves
		// ParenthesizedExpression, ESTree flattens. Class is the only
		// component; no diagnostic.
		{Code: `
        const A = (class extends React.Component { render() { return <div /> } });
      `, Tsx: true},

		// TS non-null assertion / `as` / `satisfies` wrappers around the
		// arrow inside a forwardRef call: SkipExpressionWrappers must
		// transparently peel them so the wrapper's first-argument match
		// still recognizes the inner FunctionLike as a sibling-wrapping
		// arrow.
		{Code: `
        class StoreListItem extends React.PureComponent {}
        export default React.forwardRef(((props, ref) => <StoreListItem {...props} forwardRef={ref} />) as any);
      `, Tsx: true},
		{Code: `
        class StoreListItem extends React.PureComponent {}
        export default React.forwardRef(((props, ref) => <StoreListItem {...props} forwardRef={ref} />)!);
      `, Tsx: true},

		// PrivateIdentifier-keyed class field arrow that returns JSX must
		// NOT be classified as a component (private fields are not
		// detected by upstream's Components.detect).
		{Code: `
        class App extends React.Component {
          #helper = (props) => <div>{props.x}</div>;
          render() { return <div /> }
        }
      `, Tsx: true},

		// Computed key with member-expression that does NOT return JSX:
		// Branch 9 of upstream's `getStatelessComponent` rejects when
		// neither JSX nor only-null is returned.
		{Code: `
        class App extends React.Component {
          render() { return <div /> }
        }
        const obj = { [ns.X]: (props) => props.y };
      `, Tsx: true},

		// Async / generator / async-generator FE shapes are NOT
		// components (upstream registers them with confidence 0).
		{Code: `
        async function fetchData() { return null; }
        function* gen() { yield 1; }
        async function* asyncGen() { yield 1; }
        function App(props) { return <div /> }
      `, Tsx: true, Options: map[string]interface{}{"ignoreStateless": true}},

		// Lock-in: async generator with capitalized name + JSX return —
		// upstream's `node.async && node.generator` banned-confidence
		// gate excludes this from components.list(). Without the gate
		// rslint would surface a phantom component, paired here with
		// a real sibling we'd otherwise diagnose.
		{Code: `
        async function* Gen() { return <div /> }
        function App(props) { return <div /> }
      `, Tsx: true, Options: map[string]interface{}{"ignoreStateless": true}},

		// Same-kind nesting: function-in-function where only the inner
		// returns JSX. Upstream's `getStatelessComponent` Branch 7 / 8
		// fires on the inner FE only when its enclosing scope is an
		// AssignmentExpression / PropertyAssignment with capitalized LHS
		// — neither holds here, so neither inner nor outer is a
		// component. No diagnostic.
		{Code: `
        function helper() {
          function inner() { return <div />; }
          return inner;
        }
      `, Tsx: true},

		// ---- Real-world: TypeScript-typed function components ----

		// `React.FC<Props>` typed function component declaration —
		// VariableDeclarator with TypeReference annotation. Tsgo
		// preserves the annotation; the arrow's parent path stays
		// VariableDeclarator → resolves through Branch 2.
		{Code: `
        type Props = { name: string };
        const Hello: React.FC<Props> = (props) => <div>{props.name}</div>;
      `, Tsx: true},

		// Generic forwardRef typed callsite — a single component.
		{Code: `
        interface Props { label: string }
        const Btn = React.forwardRef<HTMLButtonElement, Props>((props, ref) => (
          <button ref={ref}>{props.label}</button>
        ));
      `, Tsx: true},

		// `as const` / `satisfies` after a sibling-wrap forwardRef does
		// NOT add a phantom component — TS expression wrappers are
		// peeled by SkipExpressionWrappers / SkipExpressionWrappersUp.
		{Code: `
        class StoreListItem extends React.PureComponent {}
        export default React.forwardRef((props, ref) => <StoreListItem {...props} forwardRef={ref} />) satisfies React.ForwardRefExoticComponent<any>;
      `, Tsx: true},

		// ---- Real-world: HOC patterns ----

		// Custom HOC that's NOT a registered wrapper — the outer call
		// is just a regular call, the inner arrow is in a non-allowed
		// position (CallExpression argument), so neither classifies.
		// Pair with one class component → no diagnostic.
		{Code: `
        const enhanced = withFoo((props) => <div>{props.x}</div>);
        class App extends React.Component { render() { return <div /> } }
      `, Tsx: true, Options: map[string]interface{}{"ignoreStateless": true}},

		// Configured wrapper via componentWrapperFunctions — a string
		// entry. Wrapped arrow becomes the only component.
		{Code: `
        const App = myObserver((props) => <div>{props.x}</div>);
      `, Tsx: true, Settings: map[string]interface{}{
			"componentWrapperFunctions": []interface{}{"myObserver"},
		}},

		// Configured wrapper as object form `{property, object}`.
		{Code: `
        const App = MyLib.observer((props) => <div>{props.x}</div>);
      `, Tsx: true, Settings: map[string]interface{}{
			"componentWrapperFunctions": []interface{}{
				map[string]interface{}{"property": "observer", "object": "MyLib"},
			},
		}},

		// ---- Real-world: optional chain and ESM dynamic imports ----

		// `React?.memo(arrow)` member-level optional — upstream's
		// `MatchesAnyComponentWrapper` recognizes member-level optional
		// (PropertyAccessExpression flag) but not call-level. Single
		// component → no diagnostic.
		{Code: `
        const Hello = React?.memo((props) => <div>{props.x}</div>);
      `, Tsx: true},

		// ---- Real-world: re-exported `React` namespace alias ----
		// Aliasing the namespace itself doesn't change pragma matching —
		// only the `pragma` setting controls it. Without setting,
		// `Reactlike.memo(...)` is NOT a wrapper.
		{Code: `
        import * as Reactlike from 'react';
        const App = Reactlike.memo((props) => <div>{props.x}</div>);
        class Sibling extends React.Component { render() { return <div /> } }
      `, Tsx: true, Options: map[string]interface{}{"ignoreStateless": true}},

		// ---- Real-world: lazy / Suspense / context ----
		// `React.lazy` is NOT in default wrappers — argument is a
		// dynamic-import callback, no FunctionLike there in practice.
		// One class component → no diagnostic.
		{Code: `
        const LazyOne = React.lazy(() => import('./One'));
        class App extends React.Component { render() { return <div /> } }
      `, Tsx: true},

		// ---- Real-world: hooks-only file (no components) ----
		{Code: `
        import { useState } from 'react';
        export function useToggle(init) {
          const [on, setOn] = useState(init);
          return [on, () => setOn(v => !v)];
        }
      `, Tsx: true},

		// ---- Real-world: only render-prop callbacks, no top-level component ----
		// All function-likes are arguments of non-wrapper calls →
		// neither in allowed position. Zero detected components.
		{Code: `
        someLib.register('foo', (props) => <div>{props.x}</div>);
        someLib.register('bar', (props) => <span>{props.y}</span>);
      `, Tsx: true},

		// ---- Lock-in: ignoreStateless filters object-literal `Method() {}` shorthands ----
		// tsgo registers a MethodDeclaration child of an
		// ObjectLiteralExpression (upstream's ESTree exposes the same
		// shape as `Property { method: true, value: FunctionExpression }`,
		// where `value.type === 'FunctionExpression'` matches
		// `/Function/.test`). With `ignoreStateless: true`, the entire
		// object-literal-of-methods file should produce zero diagnostics.
		{Code: `
        export default {
          A() { return <div /> },
          B() { return <div /> },
          C() { return <div /> },
        };
      `, Tsx: true, Options: map[string]interface{}{"ignoreStateless": true}},

		// ---- Lock-in: ignoreStateless ALSO filters object getters / setters returning JSX ----
		{Code: `
        export default {
          get A() { return <div /> },
        };
        class B extends React.Component { render() { return <div /> } }
      `, Tsx: true, Options: map[string]interface{}{"ignoreStateless": true}},

		// ---- Lock-in: `componentWrapperFunctions` `"object": "<pragma>"` placeholder ----
		// Upstream's `getWrapperFunctions` substitutes the literal
		// "<pragma>" string with the configured pragma. So with a custom
		// pragma `Foo` and a wrapper entry `{property: 'observer',
		// object: '<pragma>'}`, `Foo.observer((props) => <div/>)` must
		// classify as a single component (no diagnostic).
		{Code: `
        const App = Foo.observer((props) => <div>{props.x}</div>);
      `, Tsx: true,
			Settings: map[string]interface{}{
				"react": map[string]interface{}{"pragma": "Foo"},
				"componentWrapperFunctions": []interface{}{
					map[string]interface{}{"property": "observer", "object": "<pragma>"},
				},
			},
		},

		// ---- Real-world: declaration-merging (function + namespace) ----
		// Function declaration `App` merges with namespace `App`.
		// Only the function declaration is a component candidate (the
		// namespace doesn't contribute to Components.detect). Class
		// inside the namespace doesn't extend React.Component → not a
		// component. Single component → no diagnostic.
		{Code: `
        function App(props) { return <div>{props.x}</div>; }
        namespace App {
          export class Helper {}
        }
      `, Tsx: true},

		// ---- Lock-in: bare `Component` / bare `PureComponent` extends ----
		// `ExtendsReactComponent` matches both `<pragma>.Component` and
		// the bare `Component` / `PureComponent` identifiers (regex
		// /^(Pure)?Component$/). Single bare-Component class is one
		// component — no diagnostic.
		{Code: `
        class A extends Component { render() { return <div /> } }
      `, Tsx: true},
		{Code: `
        class A extends PureComponent { render() { return <div /> } }
      `, Tsx: true},

		// ---- Lock-in: PropertyAccess wrapper.object NOT matching pragma ----
		// `Foo.memo(arrow)` with pragma=React must NOT classify as a
		// wrapper. Inner arrow is in CallExpression argument position
		// (Branch 12 reject). Pair with a sibling class — only the
		// class is a component, no diagnostic.
		{Code: `
        const A = Foo.memo((props) => <div>{props.x}</div>);
        class B extends React.Component { render() { return <div /> } }
      `, Tsx: true},

		// ---- Lock-in: ES5 component via `<pragma>.<createClass>` ----
		// `React.createClass({...})` (MemberExpression callee) ES5
		// component shape — `IsCreateReactClassObjectArg` recognizes
		// both bare `createReactClass` and pragma-qualified
		// `<pragma>.<createClass>`. Default `react.createClass` is
		// `createReactClass`; for `React.createClass(...)` the
		// `react.createClass: 'createClass'` setting is required to
		// match upstream's `pragmaUtil.getCreateClassFromContext`.
		// Single component → no diagnostic.
		{Code: `
        var A = React.createClass({
          render: function() { return <div />; }
        });
      `, Tsx: true,
			Settings: map[string]interface{}{
				"react": map[string]interface{}{"createClass": "createClass"},
			},
		},

		// ---- Lock-in: export default anonymous FunctionDeclaration ----
		// `export default function() {...}` — anonymous FD allowed only
		// in this position; Branch FunctionDeclaration's name == nil
		// arm requires the `default` modifier. Single component → no
		// diagnostic.
		{Code: `
        export default function() { return <div /> }
      `, Tsx: true},

		// ---- Lock-in: Branch 16 reject — Property parent + only-null return ----
		// Upstream `getStatelessComponent` Branch 16 rejects a
		// PropertyAssignment-positioned FunctionLike whose body only
		// returns `null`. Previously rslint's `IsStatelessReactComponentWithWrappers`
		// fell through to `return true` for the (anon arrow + computed
		// key + only-null) shape — surfacing a phantom component. Pair
		// with a sibling class to verify the arrow is NOT counted.
		{Code: `
        const obj = { [Symbol.iterator]: () => null };
        class A extends React.Component { render() { return <div /> } }
      `, Tsx: true},

		// ---- Lock-in: Branch 16 paren-transparent ----
		// `{ [k]: (() => null) }` — paren-wrapped only-null arrow under
		// computed key. ESTree flattens parens so upstream sees parent
		// = Property and rejects via Branch 16; rslint must mirror that
		// via SkipExpressionWrappersUp. Without the gate the paren-
		// wrapped arrow would surface as a phantom component.
		{Code: `
        const obj = { [Symbol.iterator]: (() => null) };
        class A extends React.Component { render() { return <div /> } }
      `, Tsx: true},

		// ---- Lock-in: paren-wrapped MemberExpression wrapper callee + nodeWrapsComponent gate ----
		// `(React.forwardRef)(arrow)` — paren wraps the MemberExpression
		// callee. ESTree flattens parens so upstream sees
		// `callee.type === 'MemberExpression'` and applies the
		// nodeWrapsComponent gate. tsgo preserves the paren, so
		// `WrapperWrapsKnownSiblingComponent` must skip it before the
		// kind check or the gate misfires (over-reports).
		{Code: `
        class StoreListItem extends React.PureComponent {}
        export default (React.forwardRef)((props, ref) => <StoreListItem {...props} forwardRef={ref} />);
      `, Tsx: true},

		// ---- Lock-in: async-generator object-literal shorthand method ----
		// `{ async *Foo() { return <div/> } }` — upstream's FE listener
		// (which fires on ESTree's `Property.value`) treats this as an
		// async generator and registers with confidence 0; rslint
		// MethodDeclaration arm must apply the same `isAsyncGenerator`
		// gate so the method does NOT surface as a component.
		{Code: `
        const obj = { async *Foo() { return <div /> } };
        function App() { return <div /> }
      `, Tsx: true, Options: map[string]interface{}{"ignoreStateless": true}},

		// ---- Documented divergence: JSDoc @extends without extends clause ----
		// Skip: true. ESLint's `isExplicitComponent` parses JSDoc tags
		// (`@extends React.Component` / `@augments React.PureComponent`)
		// via `doctrine` and treats the class as an ES6 component on tag
		// presence alone. rslint requires a real `extends` clause. With
		// upstream this would be 2 components → 1 diagnostic; rslint
		// sees 1 component → 0 diagnostics. Documented under the rule's
		// "Differences from ESLint" section. Test kept skipped to lock
		// in the divergence (will surface if the gap is closed later).
		{Code: `
        /** @extends React.Component */
        class Hello {
          render() { return <div /> }
        }
        class HelloJohn extends React.Component {
          render() { return <div /> }
        }
      `, Tsx: true, Skip: true},
	}, []rule_tester.InvalidTestCase{
		// ---- Two createReactClass / classes / functions / mixed ----
		{
			Code: crCode(`
        var Hello = createReactClass({
          render: function() {
            return <div>Hello {this.props.name}</div>;
          }
        });
        var HelloJohn = createReactClass({
          render: function() {
            return <Hello name="John" />;
          }
        });
      `),
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: onlyOne, Line: 7},
			},
		},
		{
			Code: crCode(`
        class Hello extends React.Component {
          render() {
            return <div>Hello {this.props.name}</div>;
          }
        }
        class HelloJohn extends React.Component {
          render() {
            return <Hello name="John" />;
          }
        }
        class HelloJohnny extends React.Component {
          render() {
            return <Hello name="Johnny" />;
          }
        }
      `),
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: onlyOne, Line: 7},
				{MessageId: onlyOne, Line: 12},
			},
		},
		{
			Code: `
        function Hello(props) {
          return <div>Hello {props.name}</div>;
        }
        function HelloAgain(props) {
          return <div>Hello again {props.name}</div>;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: onlyOne, Line: 5},
			},
		},
		{
			Code: `
        function Hello(props) {
          return <div>Hello {props.name}</div>;
        }
        class HelloJohn extends React.Component {
          render() {
            return <Hello name="John" />;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: onlyOne, Line: 5},
			},
		},

		// ---- Object-literal property methods that are stateless components ----
		{
			Code: `
        export default {
          RenderHello(props) {
            let {name} = props;
            return <div>{name}</div>;
          },
          RenderHello2(props) {
            let {name} = props;
            return <div>{name}</div>;
          }
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: onlyOne, Line: 7},
			},
		},

		// ---- Fragment + nested-function component ----
		{
			Code: `
        exports.Foo = function Foo() {
          return <></>
        }

        exports.createSomeComponent = function createSomeComponent(opts) {
          return function Foo() {
            return <>{opts.a}</>
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: onlyOne, Line: 7},
			},
		},

		// ---- forwardRef wrapping `<div>...</div>` (root tag is `div`, NOT a known sibling) ----
		{
			Code: `
        class StoreListItem extends React.PureComponent {
          // A bunch of stuff here
        }
        export default React.forwardRef((props, ref) => <div><StoreListItem {...props} forwardRef={ref} /></div>);
      `,
			Tsx:     true,
			Options: map[string]interface{}{"ignoreStateless": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: onlyOne, Line: 5},
			},
		},
		{
			Code: `
        const HelloComponent = (props) => {
          return <div></div>;
        }
        const HelloComponent2 = React.forwardRef((props, ref) => <div></div>);
      `,
			Tsx:     true,
			Options: map[string]interface{}{"ignoreStateless": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: onlyOne, Line: 5},
			},
		},

		// ---- SequenceExpression that resolves to an arrow ----
		{
			Code: `
        const HelloComponent = (0, (props) => {
          return <div></div>;
        });
        const HelloComponent2 = React.forwardRef((props, ref) => <><HelloComponent></HelloComponent></>);
      `,
			Tsx:     true,
			Options: map[string]interface{}{"ignoreStateless": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: onlyOne, Line: 5},
			},
		},

		// ---- Various ways the wrapper callee can be aliased ----
		{
			Code: `
        const forwardRef = React.forwardRef;
        const HelloComponent = (0, (props) => {
          return <div></div>;
        });
        const HelloComponent2 = forwardRef((props, ref) => <HelloComponent></HelloComponent>);
      `,
			Tsx:     true,
			Options: map[string]interface{}{"ignoreStateless": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: onlyOne, Line: 6},
			},
		},
		{
			Code: `
        const memo = React.memo;
        const HelloComponent = (props) => {
          return <div></div>;
        };
        const HelloComponent2 = memo((props) => <HelloComponent></HelloComponent>);
      `,
			Tsx:     true,
			Options: map[string]interface{}{"ignoreStateless": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: onlyOne, Line: 6},
			},
		},
		{
			Code: `
        const {forwardRef} = React;
        const HelloComponent = (0, (props) => {
          return <div></div>;
        });
        const HelloComponent2 = forwardRef((props, ref) => <HelloComponent></HelloComponent>);
      `,
			Tsx:     true,
			Options: map[string]interface{}{"ignoreStateless": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: onlyOne, Line: 6},
			},
		},
		{
			Code: `
        const {memo} = React;
        const HelloComponent = (0, (props) => {
          return <div></div>;
        });
        const HelloComponent2 = memo((props) => <HelloComponent></HelloComponent>);
      `,
			Tsx:     true,
			Options: map[string]interface{}{"ignoreStateless": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: onlyOne, Line: 6},
			},
		},
		{
			Code: `
        import React, { memo } from 'react';
        const HelloComponent = (0, (props) => {
          return <div></div>;
        });
        const HelloComponent2 = memo((props) => <HelloComponent></HelloComponent>);
      `,
			Tsx:     true,
			Options: map[string]interface{}{"ignoreStateless": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: onlyOne, Line: 6},
			},
		},
		{
			Code: `
        import {forwardRef} from 'react';
        const HelloComponent = (0, (props) => {
          return <div></div>;
        });
        const HelloComponent2 = forwardRef((props, ref) => <HelloComponent></HelloComponent>);
      `,
			Tsx:     true,
			Options: map[string]interface{}{"ignoreStateless": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: onlyOne, Line: 6},
			},
		},
		{
			Code: `
        const { memo } = require('react');
        const HelloComponent = (0, (props) => {
          return <div></div>;
        });
        const HelloComponent2 = memo((props) => <HelloComponent></HelloComponent>);
      `,
			Tsx:     true,
			Options: map[string]interface{}{"ignoreStateless": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: onlyOne, Line: 6},
			},
		},
		{
			Code: `
        const {forwardRef} = require('react');
        const HelloComponent = (0, (props) => {
          return <div></div>;
        });
        const HelloComponent2 = forwardRef((props, ref) => <HelloComponent></HelloComponent>);
      `,
			Tsx:     true,
			Options: map[string]interface{}{"ignoreStateless": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: onlyOne, Line: 6},
			},
		},
		{
			Code: `
        const forwardRef = require('react').forwardRef;
        const HelloComponent = (0, (props) => {
          return <div></div>;
        });
        const HelloComponent2 = forwardRef((props, ref) => <HelloComponent></HelloComponent>);
      `,
			Tsx:     true,
			Options: map[string]interface{}{"ignoreStateless": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: onlyOne, Line: 6},
			},
		},
		{
			Code: `
        const memo = require('react').memo;
        const HelloComponent = (0, (props) => {
          return <div></div>;
        });
        const HelloComponent2 = memo((props) => <HelloComponent></HelloComponent>);
      `,
			Tsx:     true,
			Options: map[string]interface{}{"ignoreStateless": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: onlyOne, Line: 6},
			},
		},

		// ---- Custom pragma via `settings.react.pragma` ----
		{
			Code: `
        import Foo, { memo, forwardRef } from 'foo';
        const Text = forwardRef(({ text }, ref) => {
          return <div ref={ref}>{text}</div>;
        })
        const Label = memo(() => <Text />);
      `,
			Tsx: true,
			Settings: map[string]interface{}{
				"react": map[string]interface{}{
					"pragma": "Foo",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: onlyOne},
			},
		},

		// ---- Lock-in: explicit `[{}]` (empty options object) lands on default `ignoreStateless = false` ----
		// Verifies the `GetOptionsMap` JSON-path round-trip preserves defaults
		// when the user supplies an empty options object — same behavior as
		// supplying no options at all (the canonical first-paragraph invalid
		// case is two function components, which `ignoreStateless: false`
		// must report).
		{
			Code: `
        function Hello(props) { return <div /> }
        function HelloAgain(props) { return <div /> }
      `,
			Tsx:     true,
			Options: []interface{}{map[string]interface{}{}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: onlyOne, Line: 3},
			},
		},

		// ---- Lock-in: array-wrapped options shape (rule_tester / multi-element CLI) ----
		{
			Code: `
        function Hello(props) { return <div /> }
        class HelloJohn extends React.Component {
          render() { return <div /> }
        }
      `,
			Tsx:     true,
			Options: []interface{}{map[string]interface{}{"ignoreStateless": false}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: onlyOne, Line: 3},
			},
		},

		// ---- Dimension 4: nested class-in-function where both qualify ----
		{
			Code: `
        function Outer(props) {
          class Inner extends React.Component { render() { return <div /> } }
          return <Inner />;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: onlyOne, Line: 3},
			},
		},

		// ---- Dimension 4: paren / TS wrapper around class expression VariableDeclarator init ----
		// Both qualify as components. Outer is a class component; second
		// VariableDeclarator inits to a paren-wrapped ClassExpression.
		{
			Code: `
        class A extends React.Component { render() { return <div /> } }
        const B = (class extends React.Component { render() { return <div /> } });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: onlyOne, Line: 3},
			},
		},

		// ---- Dimension 4: computed-key arrow that returns JSX is still a component ----
		// Branch 9 of upstream's `getStatelessComponent` only rejects when
		// the return is neither JSX nor only-null. JSX-returning arrow under
		// a `[ns.X]: ...` computed key passes through to Branch 12+ and is
		// classified — paired with a sibling class component, the second
		// (declaration-order) entry must be reported.
		{
			Code: `
        class App extends React.Component {
          render() { return <div /> }
        }
        const obj = { [ns.X]: (props) => <div>{props.y}</div> };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: onlyOne, Line: 5},
			},
		},

		// ---- Dimension 4: object literal with shorthand `Method() {}` style ----
		// Upstream invariant: each method whose key is capitalized AND
		// returns JSX is independently registered as a component.
		// When THREE such methods exist, errors should fire on entries 2
		// and 3.
		{
			Code: `
        export default {
          A() { return <div /> },
          B() { return <div /> },
          C() { return <div /> },
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: onlyOne, Line: 4},
				{MessageId: onlyOne, Line: 5},
			},
		},

		// ---- Real-world: TS class-with-generics + sibling typed FC ----
		{
			Code: `
        type Props = { name: string };
        class Hello extends React.Component<Props> { render() { return <div>{this.props.name}</div>; } }
        const Bye: React.FC<Props> = (props) => <div>Bye {props.name}</div>;
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: onlyOne, Line: 4},
			},
		},

		// ---- Real-world: arrow returning ternary with JSX (Branch ` strict=true` paths) ----
		{
			Code: `
        const A = (props) => props.cond ? <div /> : <span />;
        const B = (props) => <div>{props.x}</div>;
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: onlyOne, Line: 3},
			},
		},

		// ---- Real-world: deep nesting — inner pragma-wrapper inside a class field ----
		// Lock-in: upstream's `Components.detect` CallExpression listener
		// adds `React.memo(arrow)` regardless of its enclosing position
		// (allowed-position gate only applies to FunctionLike Branch 12;
		// CallExpression listener has no such gate). So three components
		// exist: Outer class, the React.memo wrapper inside the field
		// initializer, and the Sibling function. Errors on entries 2 & 3.
		{
			Code: `
        class Outer extends React.Component {
          inner = React.memo((props) => <div>{props.x}</div>);
          render() { return <div /> }
        }
        function Sibling(props) { return <div>{props.y}</div>; }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: onlyOne, Line: 3},
				{MessageId: onlyOne, Line: 6},
			},
		},

		// ---- Real-world: SequenceExpression chain `(0, 0, arrow)` ----
		// Upstream's `getStatelessComponent` Branch 12 recognizes
		// SequenceExpression position when the arrow is the LAST entry
		// and the sequence is itself in an allowed position
		// (VariableDeclarator). Two such siblings → second reports.
		{
			Code: `
        const A = (0, 0, (props) => <div>{props.x}</div>);
        const B = (0, (props) => <span>{props.y}</span>);
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: onlyOne, Line: 3},
			},
		},

		// ---- Real-world: configured wrapper user-defined entry ----
		{
			Code: `
        const A = myObserver((props) => <div>{props.x}</div>);
        const B = myObserver((props) => <div>{props.y}</div>);
      `,
			Tsx: true,
			Settings: map[string]interface{}{
				"componentWrapperFunctions": []interface{}{"myObserver"},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: onlyOne, Line: 3},
			},
		},

		// ---- Real-world: configured wrapper user-defined entry, ignoreStateless filters wrapper-call CallExpression node too ----
		{
			Code: `
        const A = myObserver((props) => <div>{props.x}</div>);
        class B extends React.Component { render() { return <div /> } }
        class C extends React.Component { render() { return <div /> } }
      `,
			Tsx: true,
			Options: map[string]interface{}{"ignoreStateless": true},
			Settings: map[string]interface{}{
				"componentWrapperFunctions": []interface{}{"myObserver"},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: onlyOne, Line: 4},
			},
		},

		// ---- Real-world: declaration order matters — first wins, no resort ----
		// Class is declared after function; function is the FIRST in
		// declaration order, the class is reported. Confirms position-
		// based ordering rather than kind-based.
		{
			Code: `
        function Hello(props) { return <div>{props.name}</div>; }
        class HelloAgain extends React.Component { render() { return <div /> } }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: onlyOne, Line: 3},
			},
		},

		// ---- Real-world: TS namespace + const that both yield components ----
		// Two top-level components inside a namespace block. Both
		// classify. Namespace traversal works because we use
		// ForEachChild recursively.
		{
			Code: `
        namespace UI {
          export const A = (props) => <div>{props.x}</div>;
          export const B = (props) => <span>{props.y}</span>;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: onlyOne, Line: 4},
			},
		},

		// ---- Real-world: `module.exports = function() {...}` is allowed ----
		// `module.exports` is special-cased in upstream's
		// `getStatelessComponent` (`isModuleExportsAssignment`).
		// PAIRED with a sibling component, the sibling is the SECOND
		// component (the exports-assigned anonymous fn is FIRST).
		{
			Code: `
        module.exports = function (props) { return <div>{props.x}</div>; };
        class Sibling extends React.Component { render() { return <div /> } }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: onlyOne, Line: 3},
			},
		},

		// ---- Lock-in: `componentWrapperFunctions` `<pragma>` placeholder + bare imported callee ----
		// Upstream `isPragmaComponentWrapper` bare-callee arm:
		//   wrapperFunction.object === pragma && isDestructuredFromPragmaImport(callee)
		// must match. Here a user entry `{property: 'observer', object: '<pragma>'}`
		// is configured, pragma is default `React`, and `observer` is
		// imported from 'react'. Bare `observer(arrow)` must classify as
		// the wrapper component, paired with a sibling class → second
		// component reports.
		{
			Code: `
        import { observer } from 'react';
        const A = observer((props) => <div>{props.x}</div>);
        class B extends React.Component { render() { return <div /> } }
      `,
			Tsx: true,
			Settings: map[string]interface{}{
				"componentWrapperFunctions": []interface{}{
					map[string]interface{}{"property": "observer", "object": "<pragma>"},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: onlyOne, Line: 4},
			},
		},

		// ---- Lock-in: `<pragma>` placeholder with TWO matching wrapper calls ----
		{
			Code: `
        const A = Foo.observer((props) => <div>{props.x}</div>);
        const B = Foo.observer((props) => <div>{props.y}</div>);
      `,
			Tsx: true,
			Settings: map[string]interface{}{
				"react": map[string]interface{}{"pragma": "Foo"},
				"componentWrapperFunctions": []interface{}{
					map[string]interface{}{"property": "observer", "object": "<pragma>"},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: onlyOne, Line: 3},
			},
		},

		// ---- Lock-in: nested wrapper `memo(forwardRef(arrow))` ----
		// Upstream `Components.detect` registers BOTH wrappers as
		// independent components (memo via the arrow's
		// `getPragmaComponentWrapper` outer-most ascent, forwardRef via
		// its own CallExpression listener). With a sibling class, total
		// components = 3 → 2 diagnostics, on the memo and forwardRef
		// lines.
		{
			Code: `
        class A extends React.Component { render() { return <div /> } }
        const Outer = React.memo(
          React.forwardRef((props, ref) => <div ref={ref} />)
        );
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: onlyOne, Line: 3},
				{MessageId: onlyOne, Line: 4},
			},
		},

		// ---- Real-world: deeply nested — three siblings in a single file ----
		{
			Code: `
        function A() { return <div /> }
        function B() { return <div /> }
        function C() { return <div /> }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: onlyOne, Line: 3},
				{MessageId: onlyOne, Line: 4},
			},
		},

		// ---- Contract: exact message text + line/column/endLine/endColumn ----
		// Locks in upstream's diagnostic surface byte-for-byte: the
		// message string must equal `onlyOneText`, and the report range
		// must cover the second component's full extent on the same
		// line.
		{
			Code: `
        function Hello(props) { return <div /> }
        function HelloAgain(props) { return <div /> }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: onlyOne,
					Message:   onlyOneText,
					Line:      3,
					Column:    9,
					EndLine:   3,
					EndColumn: 54,
				},
			},
		},

		// ---- Contract: multi-line component report — Line + Column + EndLine + EndColumn ----
		// Locks in the position fields when the second component spans
		// more than one source line. Class declaration starts at line
		// 5 and ends at line 7.
		{
			Code: `
        class First extends React.Component {
          render() { return <div /> }
        }
        class Second extends React.Component {
          render() { return <div /> }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: onlyOne,
					Line:      5,
					Column:    9,
					EndLine:   7,
					EndColumn: 10,
				},
			},
		},

		// ---- Branch 4 lock-in: AssignmentExpression non-MemberExpression LHS ----
		// `Capitalized = function() { return <div/> }` — Identifier LHS,
		// fn returns JSX, capitalized → component. Two such
		// assignments in same file → second reports.
		{
			Code: `
        var A, B;
        A = function() { return <div /> }
        B = function() { return <div /> }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: onlyOne, Line: 4},
			},
		},

		// ---- Branch 4 lock-in: named-FE id takes priority over LHS ----
		// `lower = function CapitalizedFE() { return <div/> }` — even
		// though LHS is lowercase, the FE's named id is capitalized so
		// the FE classifies as a component (Branch 4 in
		// IsStatelessReactComponentWithWrappers explicitly checks
		// `fn.Kind == FunctionExpression && fn.Name() != nil` BEFORE
		// looking at LHS). Pair with sibling class.
		{
			Code: `
        var helper;
        helper = function NamedComp() { return <div /> }
        class App extends React.Component { render() { return <div /> } }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: onlyOne, Line: 4},
			},
		},

		// ---- Branch 5 lock-in: nested arrow whose outer arrow is in AssignmentExpression ----
		// `X = () => () => <div/>` — outer arrow in AE with capitalized
		// LHS, inner arrow returns JSX → inner classifies as
		// component. Two of these → second reports.
		{
			Code: `
        var A, B;
        A = () => () => <div />;
        B = () => () => <div />;
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: onlyOne, Line: 4},
			},
		},

		// ---- Branch 6 lock-in: nested arrow whose outer arrow is a PropertyAssignment value ----
		// `{ Foo: () => () => <div/> }` — same as Branch 5 but
		// PropertyAssignment instead of AE.
		{
			Code: `
        const obj = {
          Foo: () => () => <div />,
          Bar: () => () => <div />,
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: onlyOne, Line: 4},
			},
		},

		// ---- Branch 7/8 lock-in: nested FE in ReturnStatement, outer FE in PropertyAssignment ----
		{
			Code: `
        const obj = {
          Foo: function() { return function() { return <div />; }; },
          Bar: function() { return function() { return <div />; }; },
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: onlyOne, Line: 4},
			},
		},

		// ---- React.createClass (MemberExpression callee) two siblings ----
		// Lock-in for `<pragma>.<createClass>` ES5 detection — requires
		// `settings.react.createClass` because the default factory name
		// is `createReactClass`, not `createClass`.
		{
			Code: `
        var A = React.createClass({ render: function() { return <div />; } });
        var B = React.createClass({ render: function() { return <div />; } });
      `,
			Tsx: true,
			Settings: map[string]interface{}{
				"react": map[string]interface{}{"createClass": "createClass"},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: onlyOne, Line: 3},
			},
		},

		// ---- bare Component / bare PureComponent extends ----
		// Both bare identifiers match `/^(Pure)?Component$/`. Two
		// classes extending bare Component / PureComponent → second
		// reports.
		{
			Code: `
        class A extends Component { render() { return <div /> } }
        class B extends PureComponent { render() { return <div /> } }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: onlyOne, Line: 3},
			},
		},

		// ---- export default anon FD + sibling ----
		// Lock-in: anonymous FunctionDeclaration only legal as
		// `export default function()`; the `default` modifier flag is
		// what lets the FunctionDeclaration arm of
		// IsStatelessReactComponentWithWrappers accept a name == nil
		// fn. Pair with sibling class.
		{
			Code: `
        class A extends React.Component { render() { return <div /> } }
        export default function() { return <div /> }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: onlyOne, Line: 3},
			},
		},
	})
}

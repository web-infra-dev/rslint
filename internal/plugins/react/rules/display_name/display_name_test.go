package display_name

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const (
	noDisplayName        = "noDisplayName"
	noContextDisplayName = "noContextDisplayName"
)

func TestDisplayNameRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &DisplayNameRule, []rule_tester.ValidTestCase{
		// ---- Shadowed wrapper identifiers (React/memo/forwardRef) ----
		{Code: `
        import React, { memo, forwardRef } from 'react'

        const TestComponent = function () {
          {
            const memo = (cb) => cb()
            const forwardRef = (cb) => cb()
            const React = { memo, forwardRef }

            const BlockMemo = memo(() => <div>shadowed</div>)
            const BlockForwardRef = forwardRef(() => <div>shadowed</div>)
            const BlockReactMemo = React.memo(() => <div>shadowed</div>)
          }
          return null
        }
      `, Tsx: true},
		{Code: `
        import React, { memo, forwardRef } from 'react'

        const Test1 = function (memo) {
          return memo(() => <div>param shadowed</div>)
        }

        const Test2 = function ({ forwardRef }) {
          return forwardRef(() => <div>destructured param</div>)
        }
      `, Tsx: true},
		{Code: `
        import React, { memo, forwardRef } from 'react'

        const TestComponent = function () {
          function innerFunction() {
            const memo = (cb) => cb()
            const React = { forwardRef }

            const Comp = memo(() => <div>nested</div>)
            const ForwardComp = React.forwardRef(() => <div>nested</div>)
            return [Comp, ForwardComp]
          }
          return innerFunction()
        }
      `, Tsx: true},
		{Code: `
        import React, { memo, forwardRef } from 'react'

        const MixedShadowed = function () {
          const memo = (cb) => cb()
          const { forwardRef } = { forwardRef: () => null }
          const [React] = [{ memo, forwardRef }]

          const Comp = memo(() => <div>shadowed</div>)
          const ReactMemo = React.memo(() => null)
          const ReactForward = React.forwardRef((props, ref) => ` + "`${props} ${ref}`" + `)
          const OtherComp = forwardRef((props, ref) => ` + "`${props} ${ref}`" + `)

          return [Comp, ReactMemo, ReactForward, OtherComp]
        }
      `, Tsx: true},

		// ---- ES5 createReactClass with displayName + ignoreTranspilerName ----
		{Code: `
        var Hello = createReactClass({
          displayName: 'Hello',
          render: function() {
            return <div>Hello {this.props.name}</div>;
          }
        });
      `, Tsx: true, Options: map[string]interface{}{"ignoreTranspilerName": true}},
		{Code: `
        var Hello = React.createClass({
          displayName: 'Hello',
          render: function() {
            return <div>Hello {this.props.name}</div>;
          }
        });
      `, Tsx: true, Options: map[string]interface{}{"ignoreTranspilerName": true},
			Settings: map[string]interface{}{
				"react": map[string]interface{}{"createClass": "createClass"},
			},
		},
		{Code: `
        class Hello extends React.Component {
          render() {
            return <div>Hello {this.props.name}</div>;
          }
        }
        Hello.displayName = 'Hello'
      `, Tsx: true, Options: map[string]interface{}{"ignoreTranspilerName": true}},

		// ---- Class without React heritage ----
		{Code: `
        class Hello {
          render() {
            return 'Hello World';
          }
        }
      `, Tsx: true},
		{Code: `
        class Hello extends Greetings {
          static text = 'Hello World';
          render() {
            return Hello.text;
          }
        }
      `, Tsx: true},
		{Code: `
        class Hello {
          method;
        }
      `, Tsx: true},

		// ---- Class with displayName as static property / getter ----
		{Code: `
        class Hello extends React.Component {
          static get displayName() {
            return 'Hello';
          }
          render() {
            return <div>Hello {this.props.name}</div>;
          }
        }
      `, Tsx: true, Options: map[string]interface{}{"ignoreTranspilerName": true}},
		{Code: `
        class Hello extends React.Component {
          static displayName = 'Widget';
          render() {
            return <div>Hello {this.props.name}</div>;
          }
        }
      `, Tsx: true, Options: map[string]interface{}{"ignoreTranspilerName": true}},

		// ---- ES5 createReactClass — has transpiler name (default ignoreTranspilerName=false) ----
		{Code: `
        var Hello = createReactClass({
          render: function() {
            return <div>Hello {this.props.name}</div>;
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

		// ---- export default class with binding name ----
		{Code: `
        export default class Hello {
          render() {
            return <div>Hello {this.props.name}</div>;
          }
        }
      `, Tsx: true},

		// ---- Reassignment of declared variable ----
		{Code: `
        var Hello;
        Hello = createReactClass({
          render: function() {
            return <div>Hello {this.props.name}</div>;
          }
        });
      `, Tsx: true},

		// ---- module.exports with displayName property ----
		{Code: `
        module.exports = createReactClass({
          "displayName": "Hello",
          "render": function() {
            return <div>Hello {this.props.name}</div>;
          }
        });
      `, Tsx: true},

		// ---- ES5 + spread + ignoreTranspilerName ----
		{Code: `
        var Hello = createReactClass({
          displayName: 'Hello',
          render: function() {
            let { a, ...b } = obj;
            let c = { ...d };
            return <div />;
          }
        });
      `, Tsx: true, Options: map[string]interface{}{"ignoreTranspilerName": true}},

		// ---- Anonymous default-exported class ----
		{Code: `
        export default class {
          render() {
            return <div>Hello {this.props.name}</div>;
          }
        }
      `, Tsx: true},

		// ---- Named function expression inside React.memo ----
		{Code: `
        export const Hello = React.memo(function Hello() {
          return <p />;
        })
      `, Tsx: true},

		// ---- Named function expressions / declarations / arrows with binding ----
		{Code: `
        var Hello = function() {
          return <div>Hello {this.props.name}</div>;
        }
      `, Tsx: true},
		{Code: `
        function Hello() {
          return <div>Hello {this.props.name}</div>;
        }
      `, Tsx: true},
		{Code: `
        var Hello = () => {
          return <div>Hello {this.props.name}</div>;
        }
      `, Tsx: true},
		{Code: `
        module.exports = function Hello() {
          return <div>Hello {this.props.name}</div>;
        }
      `, Tsx: true},

		// ---- Functional + Hello.displayName + ignoreTranspilerName ----
		{Code: `
        function Hello() {
          return <div>Hello {this.props.name}</div>;
        }
        Hello.displayName = 'Hello';
      `, Tsx: true, Options: map[string]interface{}{"ignoreTranspilerName": true}},
		{Code: `
        var Hello = () => {
          return <div>Hello {this.props.name}</div>;
        }
        Hello.displayName = 'Hello';
      `, Tsx: true, Options: map[string]interface{}{"ignoreTranspilerName": true}},
		{Code: `
        var Hello = function() {
          return <div>Hello {this.props.name}</div>;
        }
        Hello.displayName = 'Hello';
      `, Tsx: true, Options: map[string]interface{}{"ignoreTranspilerName": true}},

		// ---- Deep MemberExpression displayName ----
		{Code: `
        var Mixins = {
          Greetings: {
            Hello: function() {
              return <div>Hello {this.props.name}</div>;
            }
          }
        }
        Mixins.Greetings.Hello.displayName = 'Hello';
      `, Tsx: true, Options: map[string]interface{}{"ignoreTranspilerName": true}},

		// ---- ES5 createReactClass with helper render ----
		{Code: `
        var Hello = createReactClass({
          render: function() {
            return <div>{this._renderHello()}</div>;
          },
          _renderHello: function() {
            return <span>Hello {this.props.name}</span>;
          }
        });
      `, Tsx: true},
		{Code: `
        var Hello = createReactClass({
          displayName: 'Hello',
          render: function() {
            return <div>{this._renderHello()}</div>;
          },
          _renderHello: function() {
            return <span>Hello {this.props.name}</span>;
          }
        });
      `, Tsx: true, Options: map[string]interface{}{"ignoreTranspilerName": true}},

		// ---- Object literal with shorthand method (capitalized — Mixin button) ----
		{Code: `
        const Mixin = {
          Button() {
            return (
              <button />
            );
          }
        };
      `, Tsx: true},

		// ---- Lowercase shorthand — not a component ----
		{Code: `
        var obj = {
          pouf: function() {
            return any
          }
        };
      `, Tsx: true, Options: map[string]interface{}{"ignoreTranspilerName": true}},
		{Code: `
        var obj = {
          pouf: function() {
            return any
          }
        };
      `, Tsx: true},

		// ---- export default object literal with method that returns JSX ----
		{Code: `
        export default {
          renderHello() {
            let {name} = this.props;
            return <div>{name}</div>;
          }
        };
      `, Tsx: true},

		// ---- import + createClass settings ----
		{Code: `
        import React, { createClass } from 'react';
        export default createClass({
          displayName: 'Foo',
          render() {
            return <h1>foo</h1>;
          }
        });
      `, Tsx: true, Options: map[string]interface{}{"ignoreTranspilerName": true},
			Settings: map[string]interface{}{
				"react": map[string]interface{}{"createClass": "createClass"},
			},
		},

		// ---- Decorator-style class returned from a factory ----
		{Code: `
        import React, {Component} from "react";
        function someDecorator(ComposedComponent) {
          return class MyDecorator extends Component {
            render() {return <ComposedComponent {...this.props} />;}
          };
        }
        module.exports = someDecorator;
      `, Tsx: true},

		// ---- Capitalized SomeComponent with createElement (not a component) ----
		{Code: `
        import React, {createElement} from "react";
        const SomeComponent = (props) => {
          const {foo, bar} = props;
          return someComponentFactory({
            onClick: () => foo(bar("x"))
          });
        };
      `, Tsx: true},

		// ---- Render-prop arrow inside JSX (not a component) ----
		{Code: `
        const element = (
          <Media query={query} render={() => {
            renderWasCalled = true
            return <div/>
          }}/>
        )
      `, Tsx: true},
		{Code: `
        const element = (
          <Media query={query} render={function() {
            renderWasCalled = true
            return <div/>
          }}/>
        )
      `, Tsx: true},

		// ---- Object method named createElement (not React.createElement) ----
		{Code: `
        module.exports = {
          createElement: tagName => document.createElement(tagName)
        };
      `, Tsx: true},
		{Code: `
        const { createElement } = document;
        createElement("a");
      `, Tsx: true},

		// ---- Component + propTypes + React.memo wrapping ----
		{Code: `
        import React from 'react'
        import { string } from 'prop-types'

        function Component({ world }) {
          return <div>Hello {world}</div>
        }

        Component.propTypes = {
          world: string,
        }

        export default React.memo(Component)
      `, Tsx: true},

		// ---- React.memo wrapping named function (transpiler name) ----
		{Code: `
        import React from 'react'

        const ComponentWithMemo = React.memo(function Component({ world }) {
          return <div>Hello {world}</div>
        })
      `, Tsx: true},
		{Code: `
        import React from 'react';

        const Hello = React.memo(function Hello() {
          return;
        });
      `, Tsx: true},

		// ---- React.forwardRef wrapping named function ----
		{Code: `
        import React from 'react'

        const ForwardRefComponentLike = React.forwardRef(function ComponentLike({ world }, ref) {
          return <div ref={ref}>Hello {world}</div>
        })
      `, Tsx: true},

		// ---- Loop / array helpers / unrelated displayName field ----
		{Code: `
        function F() {
          let items = [];
          let testData = [
            {a: "test1", displayName: "test2"}, {a: "test1", displayName: "test2"}];
          for (let item of testData) {
              items.push({a: item.a, b: item.displayName});
          }
          return <div>{items}</div>;
        }
      `, Tsx: true},

		// ---- Empty class body (Flow type imports) ----
		// SKIP: rslint does not support Flow object type spreads.

		// ---- Object literal Cell render ----
		{Code: `
        const x = {
          title: "URL",
          dataIndex: "url",
          key: "url",
          render: url => (
            <a href={url} target="_blank" rel="noopener noreferrer">
              <p>lol</p>
            </a>
          )
        }
      `, Tsx: true},

		// ---- Higher-order arrow returning named function (issue #2920) ----
		{Code: `
        const renderer = a => function Component(listItem) {
          return <div>{a} {listItem}</div>;
        };
      `, Tsx: true},

		// ---- React.forwardRef + Comp.displayName ----
		{Code: `
        const Comp = React.forwardRef((props, ref) => <main />);
        Comp.displayName = 'MyCompName';
      `, Tsx: true},

		// ---- React.forwardRef + TS as expression + Comp.displayName ----
		{Code: `
        const Comp = React.forwardRef((props, ref) => <main data-as="yes" />) as SomeComponent;
        Comp.displayName = 'MyCompNameAs';
      `, Tsx: true},

		// ---- Cell column (issue #3300 / #3289 / #3329 / #3334 / #3346) ----
		{Code: `
        function Test() {
          const data = [
            {
              name: 'Bob',
            },
          ];

          const columns = [
            {
              Header: 'Name',
              accessor: 'name',
              Cell: ({ value }) => <div>{value}</div>,
            },
          ];

          return <ReactTable columns={columns} data={data} />;
        }
      `, Tsx: true},

		{Code: `
        const f = (a) => () => {
          if (a) {
            return null;
          }
          return 1;
        };
      `, Tsx: true},
		{Code: `
        class Test {
          render() {
            const data = [
              {
                name: 'Bob',
              },
            ];

            const columns = [
              {
                Header: 'Name',
                accessor: 'name',
                Cell: ({ value }) => <div>{value}</div>,
              },
            ];

            return <ReactTable columns={columns} data={data} />;
          }
        }
      `, Tsx: true},
		{Code: `
        export const demo = (a) => (b) => {
          if (a == null) return null;
          return b;
        }
      `, Tsx: true},
		{Code: `
        let demo = null;
        demo = (a) => {
          if (a == null) return null;
          return f(a);
        };
      `, Tsx: true},
		{Code: `
        obj._property = (a) => {
          if (a == null) return null;
          return f(a);
        };
      `, Tsx: true},
		{Code: `
        _variable = (a) => {
          if (a == null) return null;
          return f(a);
        };
      `, Tsx: true},
		{Code: `
        demo = () => () => null;
      `, Tsx: true},
		{Code: `
        demo = {
          property: () => () => null
        }
      `, Tsx: true},
		{Code: `
        demo = function() {return function() {return null;};};
      `, Tsx: true},
		{Code: `
        demo = {
          property: function() {return function() {return null;};}
        }
      `, Tsx: true},

		// ---- React.memo with multiple args (issue #3303) ----
		{Code: `
        function MyComponent(props) {
          return <b>{props.name}</b>;
        }

        const MemoizedMyComponent = React.memo(
          MyComponent,
          (prevProps, nextProps) => prevProps.name === nextProps.name
        )
      `, Tsx: true},

		// ---- Nested memo+forwardRef accepted in supported React versions ----
		{Code: `
        import React from 'react'

        const MemoizedForwardRefComponentLike = React.memo(
          React.forwardRef(function({ world }, ref) {
            return <div ref={ref}>Hello {world}</div>
        })
        )
      `, Tsx: true,
			Settings: map[string]interface{}{
				"react": map[string]interface{}{"version": "16.14.0"},
			},
		},
		{Code: `
        import React from 'react'

        const MemoizedForwardRefComponentLike = React.memo(
          React.forwardRef(({ world }, ref) => {
            return <div ref={ref}>Hello {world}</div>
          })
        )
      `, Tsx: true,
			Settings: map[string]interface{}{
				"react": map[string]interface{}{"version": "15.7.0"},
			},
		},
		{Code: `
        import React from 'react'

        const MemoizedForwardRefComponentLike = React.memo(
          React.forwardRef(function ComponentLike({ world }, ref) {
            return <div ref={ref}>Hello {world}</div>
          })
        )
      `, Tsx: true,
			Settings: map[string]interface{}{
				"react": map[string]interface{}{"version": "16.12.1"},
			},
		},
		{Code: `
        export const ComponentWithForwardRef = React.memo(
          React.forwardRef(function Component({ world }) {
            return <div>Hello {world}</div>
          })
        )
      `, Tsx: true,
			Settings: map[string]interface{}{
				"react": map[string]interface{}{"version": "0.14.11"},
			},
		},
		{Code: `
        import React from 'react'

        const MemoizedForwardRefComponentLike = React.memo(
          React.forwardRef(function({ world }, ref) {
            return <div ref={ref}>Hello {world}</div>
          })
        )
      `, Tsx: true,
			Settings: map[string]interface{}{
				"react": map[string]interface{}{"version": "15.7.1"},
			},
		},

		// ---- checkContextObjects: true with explicit displayName ----
		{Code: `
        import React from 'react';

        const Hello = React.createContext();
        Hello.displayName = "HelloContext"
      `, Tsx: true, Options: map[string]interface{}{"checkContextObjects": true}},
		{Code: `
        import { createContext } from 'react';

        const Hello = createContext();
        Hello.displayName = "HelloContext"
      `, Tsx: true, Options: map[string]interface{}{"checkContextObjects": true}},
		{Code: `
        import { createContext } from 'react';

        const Hello = createContext();

        const obj = {};
        obj.displayName = "False positive";

        Hello.displayName = "HelloContext"
      `, Tsx: true, Options: map[string]interface{}{"checkContextObjects": true}},
		{Code: `
        import * as React from 'react';

        const Hello = React.createContext();

        const obj = {};
        obj.displayName = "False positive";

        Hello.displayName = "HelloContext";
      `, Tsx: true, Options: map[string]interface{}{"checkContextObjects": true}},
		{Code: `
        const obj = {};
        obj.displayName = "False positive";
      `, Tsx: true, Options: map[string]interface{}{"checkContextObjects": true}},

		// ---- React version too old for context check (silently disabled) ----
		{Code: `
        import { createContext } from 'react';

        const Hello = createContext();
      `, Tsx: true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"version": "16.2.0"}},
			Options:  map[string]interface{}{"checkContextObjects": true},
		},

		// ---- React version >16.3.0 with context displayName ----
		{Code: `
        import { createContext } from 'react';

        const Hello = createContext();
        Hello.displayName = "HelloContext";
      `, Tsx: true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"version": ">16.3.0"}},
			Options:  map[string]interface{}{"checkContextObjects": true},
		},

		// ---- let / var assignments creating contexts ----
		{Code: `
        import { createContext } from 'react';

        let Hello;
        Hello = createContext();
        Hello.displayName = "HelloContext";
      `, Tsx: true, Options: map[string]interface{}{"checkContextObjects": true}},

		// ---- checkContextObjects: false leaves context unchecked ----
		{Code: `
        import { createContext } from 'react';

        const Hello = createContext();
      `, Tsx: true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"version": ">16.3.0"}},
			Options:  map[string]interface{}{"checkContextObjects": false},
		},

		// ---- var Hello reassignment with createContext ----
		{Code: `
        import { createContext } from 'react';

        var Hello;
        Hello = createContext();
        Hello.displayName = "HelloContext";
      `, Tsx: true, Options: map[string]interface{}{"checkContextObjects": true}},
		{Code: `
        import { createContext } from 'react';

        var Hello;
        Hello = React.createContext();
        Hello.displayName = "HelloContext";
      `, Tsx: true, Options: map[string]interface{}{"checkContextObjects": true}},

		// ---- Lock-in: bare options object (CLI-shape) for ignoreTranspilerName ----
		// Locks in upstream config.go:414-420 single-element option-array
		// unwrapping. Without GetOptionsMap this would silently fall back to
		// defaults and the test would fail.
		{Code: `
        var Hello = createReactClass({
          displayName: 'Hello',
          render: function() {
            return <div>Hello</div>;
          }
        });
      `, Tsx: true, Options: map[string]interface{}{"ignoreTranspilerName": true}},

		// ---- Lock-in: array-wrapped options shape ([{...}]) ----
		// Locks in `utils.GetOptionsMap` array-form unwrap. The CLI sends
		// `["error", {...}]` → after the level-1 unwrap, options arrives as
		// an interface array; GetOptionsMap takes [0]. This shape exercises
		// the same code path JS rule-tester uses.
		{Code: `
        var Hello = createReactClass({
          displayName: 'Hello',
          render: function() {
            return <div>Hello</div>;
          }
        });
      `, Tsx: true, Options: []interface{}{map[string]interface{}{"ignoreTranspilerName": true}}},

		// ---- Lock-in: bracket-access displayName property (tsgo ElementAccessExpression) ----
		// `Foo['displayName'] = 'Foo'` is upstream's MemberExpression with
		// `computed: true`. tsgo splits it into ElementAccessExpression
		// (vs PropertyAccessExpression for dot-access). Both must mark the
		// component — this case locks in the ElementAccessExpression arm.
		{Code: `
        function Hello() {
          return <div>Hello</div>;
        }
        Hello['displayName'] = 'Hello';
      `, Tsx: true, Options: map[string]interface{}{"ignoreTranspilerName": true}},

		// ---- Lock-in: computed displayName key on class field (string literal) ----
		// `class Foo extends React.Component { ['displayName'] = 'Foo' }` —
		// tsgo wraps the string literal in a ComputedPropertyName. Upstream
		// `propsUtil.isDisplayNameDeclaration` matches a string-literal
		// `'displayName'` key. We must peek through ComputedPropertyName.
		{Code: `
        class Hello extends React.Component {
          ['displayName'] = 'Hello';
          render() { return <div>Hello</div>; }
        }
      `, Tsx: true, Options: map[string]interface{}{"ignoreTranspilerName": true}},

		// ---- Lock-in: TS abstract class extends React.Component ----
		// Upstream's namedClass requires `node.id.name`. Both
		// `abstract class Foo extends React.Component {}` and the TS-only
		// abstract modifier are recognized. ESTree keeps `abstract` on the
		// ClassDeclaration; rslint must too.
		{Code: `
        abstract class Hello extends React.Component {
          render() { return <div>Hello</div>; }
        }
      `, Tsx: true},

		// ---- Lock-in: declare class extends Component (body-absent) ----
		// `declare` classes have no runtime body — upstream's component
		// detection treats them like any other named class extending
		// Component (named id → no diagnostic). Lock that in.
		{Code: `
        declare class Hello extends React.Component {
          render(): JSX.Element;
        }
      `, Tsx: true},

		// ---- Lock-in: forwardRef wrapping a known sibling component ----
		// `nodeWrapsComponent` gate: when forwardRef wraps an arrow whose
		// JSX root tag names an existing component, the forwardRef call is
		// NOT itself a component (it's a passthrough). No diagnostic.
		{Code: `
        class StoreListItem extends React.PureComponent {
          render() { return <div />; }
        }
        export default React.forwardRef((props, ref) => <StoreListItem {...props} forwardRef={ref} />);
      `, Tsx: true},

		// ---- Lock-in: triply-nested wrapper with all layers named ----
		// `memo(memo(memo(named)))` — upstream's `getStatelessComponent`
		// returns the OUTERMOST wrapper. With named inner FN, transpiler
		// name is on the inner; the CallExpression listener for the
		// outermost (no further outer wrapper, FunctionLike arg = inner
		// memo, NOT FunctionLikeExpression in upstream's check) would
		// behave differently. With this React version, all named → 0
		// reports. Locks in component-identity dedup across multiple
		// wrapper layers.
		{Code: `
        const MemoMemo = React.memo(React.memo(function Inner() { return <div/>; }));
      `, Tsx: true,
			Settings: map[string]interface{}{
				"react": map[string]interface{}{"version": "16.14.0"},
			},
		},

		// ---- Lock-in: TS satisfies / non-null wrappers around forwardRef call ----
		// `React.forwardRef(...) satisfies SomeT` and `React.forwardRef(...)!`
		// — both are TS-only expression wrappers that ESTree doesn't have.
		// `bindingNameForCallExpression` walks past these so `Comp.displayName
		// = ...` resolves through them.
		{Code: `
        const Comp = React.forwardRef((props, ref) => <main />) satisfies SomeT;
        Comp.displayName = 'MyCompName';
      `, Tsx: true},
		{Code: `
        const Comp = React.forwardRef((props, ref) => <main />)!;
        Comp.displayName = 'MyCompName';
      `, Tsx: true},

		// ---- Lock-in: parenthesized createReactClass argument ----
		// Upstream flattens parens; tsgo preserves them. The createReactClass
		// detection must walk past parens around the object literal.
		{Code: `
        var Hello = createReactClass(({
          displayName: 'Hello',
          render: function() { return <div/>; }
        }));
      `, Tsx: true, Options: map[string]interface{}{"ignoreTranspilerName": true}},

		// ---- Lock-in: getter / setter named displayName on class body ----
		// Upstream's `MethodDefinition` listener marks a `displayName`
		// getter / setter on class body as a displayName decl. tsgo splits
		// these into KindGetAccessor / KindSetAccessor; both must qualify.
		{Code: `
        class Hello extends React.Component {
          set displayName(v) {}
          render() { return <div/>; }
        }
      `, Tsx: true, Options: map[string]interface{}{"ignoreTranspilerName": true}},

		// ---- Lock-in: MethodDeclaration shorthand inside ObjectLiteral ----
		// `{ Hello() { return <div/> } }` — tsgo collapses ESTree's
		// `Property { method: true, value: FunctionExpression }` into a
		// MethodDeclaration directly under ObjectLiteralExpression. The
		// method's transpiler name comes from the parent's key (uppercase
		// → component candidate; matches `parent.method === true` arm).
		{Code: `
        const Mixin = {
          Hello() {
            return <div/>;
          }
        };
      `, Tsx: true},

		// ---- Lock-in: non-React class doesn't classify (no extends) ----
		// `class Foo {}` without React heritage is never a component.
		{Code: `
        class Hello {
          render() { return <div/>; }
        }
      `, Tsx: true},

		// ---- Lock-in: createReactClass without ignoreTranspilerName + reassignment ----
		// `var X; X = createReactClass(...)` — Identifier-LHS assignment.
		// `hasTranspilerNameForObject` must walk through the createReactClass
		// CallExpression to reach the BinaryExpression and check that LHS
		// is not module.exports.
		{Code: `
        var Hello;
        Hello = createReactClass({
          render: function() { return <div/>; }
        });
      `, Tsx: true},

		// ---- Lock-in: deep MemberExpression with object-literal nesting ----
		// Upstream's `getRelatedComponent` walks `Mixins.Greetings.Hello`
		// through nested ObjectLiteralExpression properties. Lock in the
		// path-resolution behavior — Hello inside Greetings inside Mixins.
		{Code: `
        var Mixins = {
          Greetings: {
            Hello: function() {
              return <div>Hello</div>;
            }
          }
        };
        Mixins.Greetings.Hello.displayName = 'Hello';
      `, Tsx: true, Options: map[string]interface{}{"ignoreTranspilerName": true}},

		// ---- Lock-in: shadowing via let-rebinding inside the function ----
		// Beyond const/var, `let memo = ...` also shadows.
		{Code: `
        import { memo } from 'react'
        const TestComponent = function () {
          let memo = (cb) => cb()
          const X = memo(() => <div>shadowed</div>)
          return X
        }
      `, Tsx: true},

		// ---- Real-world: TS generic arrow component ----
		// `const Comp = <T,>(props: { v: T }) => <div/>` — a common TS
		// pattern. Has transpiler name from the binding `Comp`. The
		// `<T,>` syntax is parsed as a TypeParameter list; the arrow's
		// classification must not be confused by it.
		{Code: `
        const Comp = <T,>(props: { value: T }) => <div>{props.value}</div>;
      `, Tsx: true},

		// ---- Real-world: TypeScript class with type parameters ----
		// `class Foo<P> extends React.Component<P>` — generic class.
		// `ExtendsReactComponent` reads through TypeArguments. Has id.
		{Code: `
        class Hello<P extends { name: string }> extends React.Component<P> {
          render() { return <div>{this.props.name}</div>; }
        }
      `, Tsx: true},

		// ---- Real-world: React.FC type annotation ----
		// `const Comp: React.FC = () => <div/>` — TS type annotation on the
		// VariableDeclaration doesn't change the binding shape. Transpiler
		// name still derives from `Comp`.
		{Code: `
        const Hello: React.FC = () => <div>Hello</div>;
      `, Tsx: true},
		{Code: `
        const Hello: React.FC<{ name: string }> = ({ name }) => <div>Hello {name}</div>;
      `, Tsx: true},

		// ---- Real-world: arrow assigned via destructured rebinding ----
		// `const { Foo } = { Foo: () => <div/> }` — the Foo binding is via
		// destructuring, not a direct `var Foo = ...`. Upstream `Components.detect`
		// won't classify the inner arrow on its own (parent is PropertyAssignment
		// in a non-createReactClass object). Treat as no component → no report.
		{Code: `
        const { Hello } = { Hello: () => <div>Hello</div> };
      `, Tsx: true},

		// ---- Real-world: Component + propTypes + memo wrapping ----
		// Common React + TypeScript pattern with named function. Upstream
		// & rslint: function Component is named, transpiler name applies,
		// React.memo doesn't add a separate component (first arg is
		// Identifier, not FunctionLikeExpression).
		{Code: `
        function Hello({ name }: { name: string }) { return <div>{name}</div>; }
        Hello.propTypes = { name: () => null };
        export default React.memo(Hello);
      `, Tsx: true},

		// ---- Real-world: forwardRef with named inner function + type parameters ----
		// Named inner FN `Inner` carries the transpiler name; the wrapper
		// inherits it via `getStatelessComponent → outermostWrapper` redirect.
		// Type arguments on forwardRef don't affect classification.
		{Code: `
        const Hello = React.forwardRef<HTMLDivElement, { world: string }>(
          function Inner(props, ref) {
            return <div ref={ref}>Hello {props.world}</div>;
          }
        );
      `, Tsx: true},

		// ---- Real-world: deeply nested wrapper with named inner ----
		// `memo(forwardRef(memo(function Inner() {...})))` — three layers,
		// inner is named. Outermost has transpiler name (from inner's id),
		// no report.
		{Code: `
        const Hello = React.memo(
          React.forwardRef(
            React.memo(function Inner({ world }, ref) {
              return <div ref={ref}>Hello {world}</div>;
            })
          )
        );
      `, Tsx: true,
			Settings: map[string]interface{}{
				"react": map[string]interface{}{"version": "16.14.0"},
			},
		},

		// ---- Real-world: memo with displayName via assignment ----
		// `const Comp = React.memo(...); Comp.displayName = '...'` — the
		// wrapper-call component is identified via `Comp` binding name; the
		// PropertyAccessExpression listener marks it.
		{Code: `
        const Hello = React.memo((props) => <div>{props.name}</div>);
        Hello.displayName = 'HelloMemo';
      `, Tsx: true, Options: map[string]interface{}{"ignoreTranspilerName": true}},

		// ---- Real-world: ES5 + multiple property assignments to displayName ----
		// `Foo.displayName = 'A'; Foo.displayName = 'B';` — both fire the
		// MemberExpression listener; both mark. End state: marked.
		{Code: `
        var Hello = createReactClass({
          render: function() { return <div/>; }
        });
        Hello.displayName = 'A';
        Hello.displayName = 'B';
      `, Tsx: true, Options: map[string]interface{}{"ignoreTranspilerName": true}},

		// ---- Real-world: arrow in JSX render prop (not a component) ----
		// `<Comp render={() => <div/>}/>` — the inner arrow is in JSX
		// attribute position; not in an allowed component position.
		// Branch 12 reject. No report.
		{Code: `
        const x = <Comp render={() => <div/>} renderItem={(item) => <span>{item}</span>}/>;
      `, Tsx: true},

		// ---- Real-world: arrow in array literal (not a component) ----
		// `[() => <div/>]` — bare arrow in array literal isn't an allowed
		// position. No detection.
		{Code: `
        const arr = [() => <div/>, function() { return <div/>; }];
      `, Tsx: true},

		// ---- Real-world: HOC pattern with bare Identifier arg ----
		// `withRouter(SomeNamedFn)` — first arg is Identifier, not
		// FunctionLike. The HOC call is NOT a component; the named SomeNamedFn
		// is the component (if extended/etc.). Lock-in: no extra component
		// added for the call.
		{Code: `
        function NamedComp({ x }) { return <div>{x}</div>; }
        const Wrapped = withRouter(NamedComp);
      `, Tsx: true},

		// ---- Real-world: createContext with default value argument ----
		// `createContext('defaultValue')` — first arg is a string, but
		// upstream still treats this as createContext. With displayName set,
		// no diagnostic.
		{Code: `
        import { createContext } from 'react';
        const Hello = createContext('defaultValue');
        Hello.displayName = 'HelloContext';
      `, Tsx: true, Options: map[string]interface{}{"checkContextObjects": true}},

		// ---- Real-world: TypeScript abstract method with displayName ----
		// `class Foo extends React.Component { abstract render(): JSX.Element; }`
		// — abstract member has no body; class has id. Has transpiler name.
		{Code: `
        abstract class Hello extends React.Component {
          abstract render(): JSX.Element;
        }
      `, Tsx: true},

		// ---- Real-world: PropertyAccess on `this` (NOT a component) ----
		// `this.displayName = 'Foo'` inside a method — the receiver is
		// `this`, not an Identifier matching a component. Should NOT mark
		// any component (and not crash).
		{Code: `
        class Hello extends React.Component {
          constructor(props) {
            super(props);
            this.displayName = 'Hello';
          }
          render() { return <div/>; }
        }
      `, Tsx: true},

		// ---- Real-world: optional-chain wrapper + named inner ----
		// `React?.memo(function Inner() {...})` — both member-level optional
		// AND named inner FN. Named inner gives transpiler name; the
		// optional chain on the member access doesn't disrupt detection.
		{Code: `
        const Hello = React?.memo(function Inner(props) { return <div>{props.name}</div>; });
      `, Tsx: true},

		// ---- Real-world: wrapper assigned to module.exports ----
		// `module.exports = React.memo(function Inner() {...})` — wrapper
		// not directly bound to a variable, but inner FN has own id →
		// transpiler name.
		{Code: `
        module.exports = React.memo(function Hello() { return <div/>; });
      `, Tsx: true},

		// ---- Real-world: ES5 createReactClass via React.createClass + setting ----
		// `React.createClass({...})` requires `react.createClass: 'createClass'`
		// setting — `IsCreateClassCall` matches `<pragma>.<createClass>` only
		// when the configured createClass name matches.
		{Code: `
        var Hello = React.createClass({
          render: function() { return <div/>; }
        });
      `, Tsx: true,
			Settings: map[string]interface{}{
				"react": map[string]interface{}{"createClass": "createClass"},
			},
		},

		// ---- Real-world: TS class with multiple modifiers (export default abstract) ----
		// `export default abstract class Foo extends React.Component`
		// — modifier list = [export, default, abstract]. Has id `Foo`,
		// transpiler name. No diagnostic. Locks in modifier-skipping order.
		{Code: `
        export default abstract class Hello extends React.Component {
          abstract render(): JSX.Element;
        }
      `, Tsx: true},

		// ---- Real-world: declaration merging — function + namespace ----
		// `function Foo() {} namespace Foo { ... }` — only the FN classifies
		// as a component (namespace doesn't contribute). FN is named, has
		// transpiler name. No diagnostic.
		{Code: `
        function Hello(props: { name: string }) { return <div>{props.name}</div>; }
        namespace Hello {
          export const defaultProps = { name: 'World' };
        }
      `, Tsx: true},

		// ---- Real-world: import-then-extend (createReactClass aliased) ----
		// `import { createClass as createReactClass } from 'react'` — TS
		// import alias. Default `createClass` setting still matches
		// `createReactClass` callee name. We're testing that aliasing
		// doesn't break detection (same call shape after binding).
		{Code: `
        import { createClass as createReactClass } from 'react';
        var Hello = createReactClass({
          render: function() { return <div/>; }
        });
      `, Tsx: true},

		// ---- Real-world: chained method-call wrapper ----
		// `module.exports = React.memo()(component)` — confused chain that
		// upstream's `isPragmaComponentWrapper` does NOT match (callee is
		// CallExpression `React.memo()`, not MemberExpression / Identifier).
		// Inner arrow is in CallExpression-arg position → Branch 12 reject.
		// No detection. No diagnostic.
		{Code: `
        module.exports = React.memo()((props) => <div>{props.name}</div>);
      `, Tsx: true},

		// ---- Real-world: createContext with TS as wrapper ----
		// `const X = createContext() as Context<T>` — `resolveCreateContextCall`
		// SkipExpressionWrappers handles the outer `as`. Hello.displayName set
		// → no diagnostic.
		{Code: `
        import { createContext } from 'react';
        const Hello = (createContext() as React.Context<{ name: string }>);
        Hello.displayName = 'HelloContext';
      `, Tsx: true, Options: map[string]interface{}{"checkContextObjects": true}},

		// ---- Real-world: createContext via call-level TS satisfies ----
		// Same shape as `as` but using `satisfies`. Both are TS-only
		// expression wrappers; `SkipExpressionWrappers` peels both.
		{Code: `
        import { createContext } from 'react';
        const Hello = (createContext() satisfies React.Context<unknown>);
        Hello.displayName = 'HelloContext';
      `, Tsx: true, Options: map[string]interface{}{"checkContextObjects": true}},

		// ---- Real-world: empty source file ----
		// No content → no components → no errors.
		{Code: ``, Tsx: true},

		// ---- Real-world: pure-type code (no runtime components) ----
		// Type-only declarations — none classify as components.
		{Code: `
        type Props = { name: string };
        interface Helper { foo(): void; }
        type Comp = React.FC<Props>;
      `, Tsx: true},

		// ---- Real-world: extreme nesting — class component containing arrow ----
		// `class Outer { render() { const handleClick = () => {...}; return <div/>; } }`
		// — handleClick is lowercase arrow inside a class method. NOT a component.
		// Class is named → no diagnostic.
		{Code: `
        class Hello extends React.Component {
          render() {
            const handleClick = () => { console.log('click'); };
            const renderHelper = () => <span>helper</span>;
            return <div onClick={handleClick}>{renderHelper()}</div>;
          }
        }
      `, Tsx: true},

		// ---- Real-world: forwardRef + named inner inside TS `as` cast ----
		// `const Comp = forwardRef(function Inner(...) {...}) as ForwardRefExoticComponent`
		// — has TS `as` wrapper outside the wrapper call. Inner FN is named
		// → transpiler name carries through.
		{Code: `
        const Hello = React.forwardRef(function Inner(props, ref) {
          return <div ref={ref}>{props.name}</div>;
        }) as React.ForwardRefExoticComponent<any>;
      `, Tsx: true},

		// ---- Real-world: parameter destructuring with default in shadowed wrapper ----
		// `function f({memo = x} = {}) { memo(...) }` — destructured param
		// with default value still binds `memo`. Shadowing applies.
		{Code: `
        import { memo } from 'react'
        function Test({ memo: localMemo = (cb) => cb() } = {}) {
          return localMemo(() => <div>shadowed</div>);
        }
      `, Tsx: true},

		// ---- Real-world: ignore lowercase function declaration ----
		// `function helper() { return <div/>; }` — lowercase name. Even
		// though it returns JSX, lowercase prefix excludes it from
		// `isComponentName` checks. Branch 12 reject. NOT a component.
		{Code: `
        function helper() { return <div/>; }
        export default function App() { return helper(); }
      `, Tsx: true},

		// ---- Real-world: class component with abstract render ----
		// `abstract class Foo extends React.Component { abstract render(): JSX.Element }`
		// — body-absent abstract method. Class has id `Foo`, has transpiler name.
		{Code: `
        abstract class Hello extends React.Component {
          abstract render(): JSX.Element;
        }
      `, Tsx: true},

		// ---- Real-world: class with declare modifier ----
		// `declare class Foo extends React.Component {}` — has id, no body.
		// Type-only declaration but still classifies as a named component.
		{Code: `
        declare class Hello extends React.Component {
          render(): JSX.Element;
        }
      `, Tsx: true},

		// ---- Real-world: forwardRef inside a custom HOC factory ----
		// `withHOC(forwardRef(...))` — the inner forwardRef has wrapper status
		// only if `withHOC` is configured. With default settings, withHOC
		// isn't a wrapper, so the forwardRef call is just a wrapped FN that
		// classifies. Inner arrow anonymous → reports the forwardRef call.
		// Wait, that would be invalid. Move to invalid section if so.
		// Actually: the parent of forwardRef is CallExpression (withHOC(...)).
		// `Components.detect`'s CallExpression arm classifies forwardRef IF its
		// first arg is FunctionLike. Yes. Then redirect to outermost wrapper:
		// `outermostWrapperCall(arrow)` returns forwardRef (since withHOC
		// isn't a wrapper). Register forwardRef. No transpiler name, no
		// displayName decl... → would report. So move to invalid.

		// ---- Real-world: forwardRef + memo nested with named inner (backwards) ----
		// `forwardRef(memo(function Inner() {...}))` — opposite nesting from
		// usual `memo(forwardRef(...))`. Inner is named. With nested-memo
		// React version, both layers redirect to outermost. Has transpiler
		// name from inner.
		{Code: `
        const Hello = React.forwardRef(React.memo(function Inner({ x }, ref) {
          return <div ref={ref}>{x}</div>;
        }));
      `, Tsx: true,
			Settings: map[string]interface{}{
				"react": map[string]interface{}{"version": "16.14.0"},
			},
		},

		// ---- Real-world: forwardRef returning null ----
		// `forwardRef((props, ref) => null)` — returns null, not JSX. With
		// `IsStatelessReactComponentWithWrappers`, null-only return WITH
		// pragma wrapper IS classified (via the special null-return path).
		// Inner arrow has no transpiler name; the wrapper has no binding
		// here either... add a binding to make it valid:
		{Code: `
        const Hello = React.forwardRef(function Inner(props, ref) {
          return null;
        });
      `, Tsx: true},

		// ---- Real-world: class with displayName via static block ----
		// `class Foo extends React.Component { static { Foo.displayName = 'Foo'; } }`
		// — static initialization block sets displayName at class init time.
		// Upstream's `MemberExpression` listener catches `Foo.displayName = ...`
		// regardless of where it appears; the deep / nameToComponent path
		// resolves `Foo`.
		{Code: `
        class Hello extends React.Component {
          static {
            Hello.displayName = 'Hello';
          }
          render() { return <div/>; }
        }
      `, Tsx: true, Options: map[string]interface{}{"ignoreTranspilerName": true}},

		// ---- Lock-in: `(arrow) as Type` in default-export position is NOT classified ----
		// `export default ((arrow) as React.FC)` — both upstream's
		// `isInAllowedPositionForComponent` and rslint's parent-walk look
		// at the IMMEDIATE parent (skipping ParenthesizedExpression only).
		// TS expression wrappers (`as` / `satisfies` / `!`) sit between the
		// arrow and the ExportAssignment, so the arrow doesn't classify
		// as a component and the lint doesn't fire. Lock in this
		// (admittedly conservative) behavior to align with upstream.
		{Code: `
        export default ((props: { name: string }) => <div>{props.name}</div>) as React.FC<any>;
      `, Tsx: true},
		{Code: `
        export default ((props: { x: string }) => <div>{props.x}</div>) satisfies React.FC<any>;
      `, Tsx: true},
		{Code: `
        export default (((props: { x: string }) => <div>{props.x}</div>)!);
      `, Tsx: true},

		// ---- Lock-in: async generator banned — never classifies as component ----
		// `async function* X() { return <div/>; }` — upstream's
		// `Components.detect` registers async generators at confidence 0,
		// which is permanently banned from `components.list()`. Without
		// our explicit gate, the generator's `return <div/>` would slip
		// through as JSX-returning and the rule would falsely report.
		// In real code generators are never React components.
		{Code: `
        async function* Hello() {
          yield 1;
          return <div>not a component</div>;
        }
        export const Real = () => <div/>;
      `, Tsx: true},

		// FunctionExpression form of async generator (also banned).
		{Code: `
        const gen = async function* () {
          return <div>not a component</div>;
        };
        export const Hello = () => <div/>;
      `, Tsx: true},

		// MethodDeclaration shorthand form of async generator (also banned).
		{Code: `
        const obj = {
          async *Hello() {
            return <div/>;
          }
        };
        export const Real = () => <div/>;
      `, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// ---- Shadowed-but-only-some-paths shadowed: the un-shadowed ones report ----
		{Code: `
        import React, { memo, forwardRef } from 'react'

        const TestComponent = function () {
          {
            const BlockReactMemo = React.memo(() => {
              return <div>not shadowed</div>
            })

            const BlockMemo = memo(() => {
              return <div>not shadowed</div>
            })

            const BlockForwardRef = forwardRef((props, ref) => {
              return ` + "`${props} ${ref}`" + `
            })
          }

          return null
        }
      `, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: noDisplayName},
				{MessageId: noDisplayName},
				{MessageId: noDisplayName},
			},
		},
		{Code: `
        import React, { memo, forwardRef } from 'react'

        const Test1 = function () {
          const Comp = memo(() => <div>not param shadowed</div>)
          return Comp
        }

        const Test2 = function () {
          function innerFunction() {
            const Comp = memo(() => <div>nested not shadowed</div>)
            const ForwardComp = React.forwardRef(() => <div>nested</div>)
            return [Comp, ForwardComp]
          }
          return innerFunction()
        }
      `, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: noDisplayName},
				{MessageId: noDisplayName},
				{MessageId: noDisplayName},
			},
		},
		{Code: `
        import React, { memo, forwardRef } from 'react'

        const MixedNotShadowed = function () {
          const Comp = memo(() => {
            return <div>not shadowed</div>
          })
          const ReactMemo = React.memo(() => null)
          const ReactForward = React.forwardRef((props, ref) => {
            return ` + "`${props} ${ref}`" + `
          })
          const OtherComp = forwardRef((props, ref) => ` + "`${props} ${ref}`" + `)

          return [Comp, ReactMemo, ReactForward, OtherComp]
        }
      `, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: noDisplayName},
				{MessageId: noDisplayName},
				{MessageId: noDisplayName},
				{MessageId: noDisplayName},
			},
		},

		// ---- ES5 createReactClass without displayName + ignoreTranspilerName ----
		{Code: `
        var Hello = createReactClass({
          render: function() {
            return React.createElement("div", {}, "text content");
          }
        });
      `, Tsx: true, Options: map[string]interface{}{"ignoreTranspilerName": true},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: noDisplayName}},
		},
		{Code: `
        var Hello = React.createClass({
          render: function() {
            return React.createElement("div", {}, "text content");
          }
        });
      `, Tsx: true, Options: map[string]interface{}{"ignoreTranspilerName": true},
			Settings: map[string]interface{}{
				"react": map[string]interface{}{"createClass": "createClass"},
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: noDisplayName}},
		},
		{Code: `
        var Hello = createReactClass({
          render: function() {
            return <div>Hello {this.props.name}</div>;
          }
        });
      `, Tsx: true, Options: map[string]interface{}{"ignoreTranspilerName": true},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: noDisplayName}},
		},

		// ---- Class without displayName + ignoreTranspilerName ----
		{Code: `
        class Hello extends React.Component {
          render() {
            return <div>Hello {this.props.name}</div>;
          }
        }
      `, Tsx: true, Options: map[string]interface{}{"ignoreTranspilerName": true},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: noDisplayName}},
		},

		// ---- Function returning createReactClass result ----
		{Code: `
        function HelloComponent() {
          return createReactClass({
            render: function() {
              return <div>Hello {this.props.name}</div>;
            }
          });
        }
        module.exports = HelloComponent();
      `, Tsx: true, Options: map[string]interface{}{"ignoreTranspilerName": true},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: noDisplayName}},
		},

		// ---- module.exports anonymous arrow / function ----
		{Code: `
        module.exports = () => {
          return <div>Hello {props.name}</div>;
        }
      `, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: noDisplayName}},
		},
		{Code: `
        module.exports = function() {
          return <div>Hello {props.name}</div>;
        }
      `, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: noDisplayName}},
		},

		// ---- module.exports of createReactClass without displayName ----
		{Code: `
        module.exports = createReactClass({
          render() {
            return <div>Hello {this.props.name}</div>;
          }
        });
      `, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: noDisplayName}},
		},

		// ---- ES5 with helper render method ----
		{Code: `
        var Hello = createReactClass({
          _renderHello: function() {
            return <span>Hello {this.props.name}</span>;
          },
          render: function() {
            return <div>{this._renderHello()}</div>;
          }
        });
      `, Tsx: true, Options: map[string]interface{}{"ignoreTranspilerName": true},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: noDisplayName}},
		},

		// ---- ES5 with custom pragma + createClass ----
		{Code: `
        var Hello = Foo.createClass({
          _renderHello: function() {
            return <span>Hello {this.props.name}</span>;
          },
          render: function() {
            return <div>{this._renderHello()}</div>;
          }
        });
      `, Tsx: true, Options: map[string]interface{}{"ignoreTranspilerName": true},
			Settings: map[string]interface{}{
				"react": map[string]interface{}{
					"pragma":      "Foo",
					"createClass": "createClass",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: noDisplayName}},
		},
		{Code: `
        /** @jsx Foo */
        var Hello = Foo.createClass({
          _renderHello: function() {
            return <span>Hello {this.props.name}</span>;
          },
          render: function() {
            return <div>{this._renderHello()}</div>;
          }
        });
      `, Tsx: true, Options: map[string]interface{}{"ignoreTranspilerName": true},
			Settings: map[string]interface{}{
				"react": map[string]interface{}{"createClass": "createClass"},
			},
			// SKIP: rslint does not support ESLint's `/** @jsx ... */` directive
			// comments. The pragma is read solely from settings.react.pragma.
			Skip:   true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: noDisplayName}},
		},

		// ---- Object literal shorthand method as component + ignoreTranspilerName ----
		{Code: `
        const Mixin = {
          Button() {
            return (
              <button />
            );
          }
        };
      `, Tsx: true, Options: map[string]interface{}{"ignoreTranspilerName": true},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: noDisplayName}},
		},

		// ---- Higher-order anonymous function returning JSX ----
		{Code: `
        function Hof() {
          return function () {
            return <div />
          }
        }
      `, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: noDisplayName}},
		},

		// ---- Anonymous arrow exporting createElement ----
		{Code: `
        import React, { createElement } from "react";
        export default (props) => {
          return createElement("div", {}, "hello");
        };
      `, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: noDisplayName}},
		},

		// ---- React.memo / React.forwardRef with anonymous arrow ----
		{Code: `
        import React from 'react'

        const ComponentWithMemo = React.memo(({ world }) => {
          return <div>Hello {world}</div>
        })
      `, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: noDisplayName}},
		},
		{Code: `
        import React from 'react'

        const ComponentWithMemo = React.memo(function() {
          return <div>Hello {world}</div>
        })
      `, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: noDisplayName}},
		},
		{Code: `
        import React from 'react'

        const ForwardRefComponentLike = React.forwardRef(({ world }, ref) => {
          return <div ref={ref}>Hello {world}</div>
        })
      `, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: noDisplayName}},
		},
		{Code: `
        import React from 'react'

        const ForwardRefComponentLike = React.forwardRef(function({ world }, ref) {
          return <div ref={ref}>Hello {world}</div>
        })
      `, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: noDisplayName}},
		},

		// ---- Nested memo+forwardRef NOT supported in older React versions ----
		{Code: `
        import React from 'react'

        const MemoizedForwardRefComponentLike = React.memo(
          React.forwardRef(({ world }, ref) => {
            return <div ref={ref}>Hello {world}</div>
          })
        )
      `, Tsx: true,
			Settings: map[string]interface{}{
				"react": map[string]interface{}{"version": "15.6.0"},
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: noDisplayName}},
		},
		{Code: `
        import React from 'react'

        const MemoizedForwardRefComponentLike = React.memo(
          React.forwardRef(function({ world }, ref) {
            return <div ref={ref}>Hello {world}</div>
          })
        )
      `, Tsx: true,
			Settings: map[string]interface{}{
				"react": map[string]interface{}{"version": "0.14.2"},
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: noDisplayName}},
		},
		{Code: `
        import React from 'react'

        const MemoizedForwardRefComponentLike = React.memo(
          React.forwardRef(function ComponentLike({ world }, ref) {
            return <div ref={ref}>Hello {world}</div>
          })
        )
      `, Tsx: true,
			Settings: map[string]interface{}{
				"react": map[string]interface{}{"version": "15.0.1"},
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: noDisplayName}},
		},

		// ---- Destructured createElement from React ----
		{Code: `
        import React from "react";
        const { createElement } = React;
        export default (props) => {
          return createElement("div", {}, "hello");
        };
      `, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: noDisplayName}},
		},
		{Code: `
        import React from "react";
        const createElement = React.createElement;
        export default (props) => {
          return createElement("div", {}, "hello");
        };
      `, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: noDisplayName}},
		},

		// ---- Inner non-component functions don't shadow outer report ----
		{Code: `
        module.exports = function () {
          function a () {}
          const b = function b () {}
          const c = function () {}
          const d = () => {}
          const obj = {
            a: function a () {},
            b: function b () {},
            c () {},
            d: () => {},
          }
          return React.createElement("div", {}, "text content");
        }
      `, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: noDisplayName}},
		},
		{Code: `
        module.exports = () => {
          function a () {}
          const b = function b () {}
          const c = function () {}
          const d = () => {}
          const obj = {
            a: function a () {},
            b: function b () {},
            c () {},
            d: () => {},
          }

          return React.createElement("div", {}, "text content");
        }
      `, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: noDisplayName}},
		},
		{Code: `
        export default class extends React.Component {
          render() {
            function a () {}
            const b = function b () {}
            const c = function () {}
            const d = () => {}
            const obj = {
              a: function a () {},
              b: function b () {},
              c () {},
              d: () => {},
            }
            return <div>Hello {this.props.name}</div>;
          }
        }
      `, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: noDisplayName}},
		},

		// ---- Two unrelated anonymous components — line/column assertions ----
		{Code: `
        export default class extends React.PureComponent {
          render() {
            return <Card />;
          }
        }

        const Card = (() => {
          return React.memo(({ }) => (
            <div />
          ));
        })();
      `, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: noDisplayName, Line: 2, Column: 24},
				{MessageId: noDisplayName, Line: 9, Column: 18},
			},
		},

		// ---- Curried arrow returning JSX (issue) ----
		{Code: `
        const renderer = a => listItem => (
          <div>{a} {listItem}</div>
        );
      `, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: noDisplayName, Message: "Component definition is missing display name"},
			},
		},

		// ---- componentWrapperFunctions setting ----
		{Code: `
        const processData = (options?: { value: string }) => options?.value || 'no data';

        export const Component = observer(() => {
          const data = processData({ value: 'data' });
          return <div>{data}</div>;
        });

        export const Component2 = observer(() => {
          const data = processData();
          return <div>{data}</div>;
        });
      `, Tsx: true,
			Settings: map[string]interface{}{
				"componentWrapperFunctions": []interface{}{"observer"},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: noDisplayName, Message: "Component definition is missing display name", Line: 4},
				{MessageId: noDisplayName, Message: "Component definition is missing display name", Line: 9},
			},
		},

		// ---- checkContextObjects: missing context displayName ----
		{Code: `
        import React from 'react';

        const Hello = React.createContext();
      `, Tsx: true, Options: map[string]interface{}{"checkContextObjects": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: noContextDisplayName, Line: 4},
			},
		},
		{Code: `
        import * as React from 'react';

        const Hello = React.createContext();
      `, Tsx: true, Options: map[string]interface{}{"checkContextObjects": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: noContextDisplayName, Line: 4},
			},
		},
		{Code: `
        import { createContext } from 'react';

        const Hello = createContext();
      `, Tsx: true, Options: map[string]interface{}{"checkContextObjects": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: noContextDisplayName, Line: 4},
			},
		},
		{Code: `
        import { createContext } from 'react';

        var Hello;
        Hello = createContext();
      `, Tsx: true, Options: map[string]interface{}{"checkContextObjects": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: noContextDisplayName, Line: 5},
			},
		},
		{Code: `
        import { createContext } from 'react';

        var Hello;
        Hello = React.createContext();
      `, Tsx: true, Options: map[string]interface{}{"checkContextObjects": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: noContextDisplayName, Line: 5},
			},
		},

		// ---- Lock-in: anonymous arrow returning createElement ----
		// Reachable via Branch 12 path: Arrow whose immediate parent is
		// `module.exports = ...`. Has no transpiler name (LHS is
		// member-expression `module.exports`, not an Identifier). Returns
		// JSX. Reports.
		{Code: `
        module.exports = (props) => React.createElement("div", {}, "hello");
      `, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: noDisplayName}},
		},

		// ---- Lock-in: ignoreTranspilerName + ES5 reassignment ----
		// `var X; X = createReactClass(...)` — has transpiler name from the
		// assignment, but with ignoreTranspilerName we still require an
		// explicit displayName property.
		{Code: `
        var Hello;
        Hello = createReactClass({
          render: function() { return <div/>; }
        });
      `, Tsx: true, Options: map[string]interface{}{"ignoreTranspilerName": true},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: noDisplayName}},
		},

		// ---- Lock-in: nested triple-memo without inner name (older React) ----
		// `memo(memo(() => <div/>))` on React 15.6 — three layers, none
		// named, none in nested-memo-supported range. The outermost
		// reports (single component identity preserved by
		// outermostWrapperCall redirect; inner layers don't double-report).
		{Code: `
        import React from 'react'

        const T = React.memo(React.memo(() => <div/>))
      `, Tsx: true,
			Settings: map[string]interface{}{
				"react": map[string]interface{}{"version": "15.6.0"},
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: noDisplayName}},
		},

		// ---- Lock-in: ES5 module.exports with createReactClass ----
		// LHS `module.exports` disqualifies the named-object-assignment arm.
		// No transpiler name even though the right-hand side is recognized.
		// With ignoreTranspilerName=false, still reports because the rule
		// checks transpiler name AND module.exports gates it out.
		{Code: `
        module.exports = createReactClass({
          render: function() { return <div/>; }
        });
      `, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: noDisplayName}},
		},

		// ---- Lock-in: bare-Identifier wrapper without import ----
		// `memo(() => <div/>)` — `memo` is a bare identifier with no
		// React import. `MatchesAnyComponentWrapperWithChecker` gates the
		// hardcoded `{Property: "memo"}` default behind
		// `IsDestructuredFromPragmaImport`. Without the import, this is
		// NOT a wrapper call, so the arrow is in CallExpression-arg
		// position (Branch 12 reject); it doesn't classify. Pair with a
		// valid component to verify only the valid one would normally
		// register — here neither classifies, no errors.
		// SKIP — covered indirectly: when memo isn't imported,
		// the arrow doesn't classify as a component, so there's nothing
		// to report. Add to invalid only if/when reactutil's import
		// detection mode changes.

		// ---- Lock-in: CallExpression args is empty ----
		// `React.memo()` — the call has zero arguments. Detection must
		// not crash and must skip. Pair with an anonymous arrow assigned
		// to module.exports so the run still has an expected error and
		// the empty-call short-circuit gets exercised on the same file.
		{Code: `
        const empty = React.memo();
        module.exports = () => <div/>;
      `, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: noDisplayName}},
		},

		// ---- Lock-in: CallExpression first arg is non-FunctionLike literal ----
		// `React.memo(SomeOtherComp)` where SomeOtherComp is an Identifier
		// — first arg is NOT a FunctionLikeExpression, the CallExpression
		// arm of Components.detect requires `IsFunctionLikeForComponent`.
		// No detection on the memo call. Pair with a real invalid one.
		{Code: `
        const SomeComp = function () { return <div/>; }
        const Memoized = React.memo(SomeComp);
        const Anon = React.memo(() => <div/>);
      `, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: noDisplayName}},
		},

		// ---- Real-world: anonymous arrow returned from function ----
		// `function makeComp() { return () => <div/>; }` — the inner arrow
		// is in ReturnStatement position, no transpiler name. The outer
		// `makeComp` returns it. Reports the inner arrow.
		{Code: `
        function makeComp() {
          return () => <div/>;
        }
      `, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: noDisplayName}},
		},

		// ---- Real-world: TS generic anonymous arrow without binding ----
		// Without a binding, `<T,>(props) => <div/>` has no transpiler name.
		// In ExportAssignment position, gets registered → reports.
		{Code: `
        export default <T,>(props: { value: T }) => <div>{props.value}</div>;
      `, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: noDisplayName}},
		},

		// ---- Real-world: forwardRef anonymous arrow + propTypes only ----
		// `const Comp = React.forwardRef((props, ref) => ...)` then
		// `Comp.propTypes = {...}` — no displayName, only propTypes. Reports.
		{Code: `
        const Hello = React.forwardRef((props, ref) => <div ref={ref}>{props.x}</div>);
        Hello.propTypes = { x: () => null };
      `, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: noDisplayName}},
		},

		// ---- Real-world: shadowed wrapper inside fn, but not at top level ----
		// Mixed: shadowed memo inside a component, but a top-level memo that
		// IS the imported one. Top-level reports; shadowed doesn't.
		{Code: `
        import { memo } from 'react'
        const Outer = memo(() => <div>top-level</div>);

        function inner() {
          const memo = (cb) => cb();
          return memo(() => <div>shadowed</div>);
        }
      `, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: noDisplayName}},
		},

		// ---- Real-world: ignoreTranspilerName + memo.displayName ----
		// `const Comp = React.memo(...); Comp.displayName = '...'` —
		// even with ignoreTranspilerName, explicit displayName satisfies.
		// Inverse case: forgot to set displayName → reports.
		{Code: `
        const Hello = React.memo((props) => <div>{props.name}</div>);
      `, Tsx: true, Options: map[string]interface{}{"ignoreTranspilerName": true},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: noDisplayName}},
		},

		// ---- Real-world: createContext registered both in valid & invalid ----
		// Multiple createContext calls in one file. checkContextObjects: true.
		// One has displayName, the other doesn't. Order in source is preserved.
		{Code: `
        import { createContext } from 'react';
        const A = createContext();
        A.displayName = 'A';
        const B = createContext();
        const C = createContext();
        C.displayName = 'C';
      `, Tsx: true, Options: map[string]interface{}{"checkContextObjects": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: noContextDisplayName},
			},
		},

		// ---- Real-world: createContext via React.X without displayName ----
		// React 16.3+ — checkContextObjects fires.
		{Code: `
        const Hello = React.createContext();
      `, Tsx: true, Options: map[string]interface{}{"checkContextObjects": true},
			Settings: map[string]interface{}{"react": map[string]interface{}{"version": "16.3.0"}},
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: noContextDisplayName}},
		},

		// ---- Real-world: nested wrapper, NONE named, OLD React ----
		// `memo(memo(memo(arrow)))` — three layers, none named, on React
		// 15.6 (BEFORE nested-memo support). Outermost is reported once
		// (inner layers redirect to outermost via `outermostWrapperCall`,
		// preserving single-component identity).
		{Code: `
        const Hello = React.memo(React.memo(React.memo(() => <div/>)));
      `, Tsx: true,
			Settings: map[string]interface{}{
				"react": map[string]interface{}{"version": "15.6.0"},
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: noDisplayName}},
		},

		// ---- Real-world: deep MemberExpression with displayName missing ----
		// `Mixins.Greetings.Hello` inner FN — without displayName declared,
		// upstream reports the inner FN. ignoreTranspilerName disables the
		// transpiler-name fallback.
		{Code: `
        var Mixins = {
          Greetings: {
            Hello: function() {
              return <div>Hello</div>;
            }
          }
        };
      `, Tsx: true, Options: map[string]interface{}{"ignoreTranspilerName": true},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: noDisplayName}},
		},

		// ---- Real-world: arrow in IIFE result returned from arrow ----
		// `const x = (() => () => <div/>)()` — outer IIFE, inner returns
		// arrow. The innermost arrow has no binding, in ReturnStatement
		// position of an arrow body, with capitalized binding `x`?
		// Actually `x` is lowercase — Branch 12 reject? Let me check.
		// `Comp` capitalized binding makes the inner arrow classify.
		{Code: `
        const Hello = (() => () => <div/>)();
      `, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: noDisplayName}},
		},

		// ---- Real-world: anonymous class component with null-only render ----
		// `class extends React.Component { render() { return null; } }` —
		// classifies (extends Component). Anonymous → no transpiler name.
		// `ignoreTranspilerName` doesn't matter; it's anonymous.
		{Code: `
        export default class extends React.Component {
          render() { return null; }
        }
      `, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: noDisplayName}},
		},

		// ---- Real-world: anonymous class assigned to identifier (no class id) ----
		// `var Foo = class extends React.Component {}` — anonymous class,
		// but `Foo` binding gives transpiler name? In upstream, namedClass
		// requires `node.id.name`, which is absent. So no transpiler name.
		// VARIABLE binding gives namedFunctionExpression-style name only for
		// FE/Arrow, NOT classes. Reports.
		{Code: `
        var Hello = class extends React.Component {
          render() { return <div/>; }
        };
      `, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: noDisplayName}},
		},

		// ---- Real-world: ES5 createReactClass via call to non-pragma ----
		// `var X = MyOwnFactory({ render: ... })` — not the configured
		// createClass. NOT classified as ES5. The arrow inside might still
		// classify though… actually there's no arrow returning JSX directly
		// here, just an object literal. Not a component. No report.
		// Inverse: with no component pattern, no diagnostic.
		// Use an actual real-world FN to keep this list invalid:
		{Code: `
        var Helper = MyOwnFactory({ render: function() { return <div/>; } });
        export default function() { return <div>I am the only component</div>; }
      `, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: noDisplayName}},
		},

		// ---- Real-world: forwardRef with TS as wrapper outside the call ----
		// `const Comp = React.forwardRef(...) as ForwardRefExoticComponent<P>`
		// — inner arrow is anonymous; outer wrapper bound via VariableDeclaration
		// `Comp`. Has transpiler name. NO error (this is in valid).
		// Inverse: WITHOUT outer binding → report:
		{Code: `
        export default React.forwardRef((props, ref) => <main ref={ref} />) as React.ForwardRefExoticComponent<any>;
      `, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: noDisplayName}},
		},

		// ---- Real-world: anonymous memo arg in CallExpression position ----
		// `const X = memo(() => <div/>)` — wrapper bound to `X`, but
		// anonymous arrow has no transpiler name. Wrapper is registered
		// as component; no transpiler name to inherit. Reports.
		{Code: `
        import { memo } from 'react'
        const Hello = memo(() => <div/>);
      `, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: noDisplayName}},
		},

		// ---- Real-world: forwardRef anonymous bound to const ----
		// Same shape as above. Component identity is the wrapper call;
		// inner anonymous arrow doesn't transfer name to wrapper.
		{Code: `
        import { forwardRef } from 'react'
        const Hello = forwardRef((props, ref) => <div ref={ref} />);
      `, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: noDisplayName}},
		},

		// ---- Real-world: memo+forwardRef nested, both anonymous ----
		// `memo(forwardRef(arrow))` — neither named, on React 15.0.0
		// (before nested-memo support). Outermost reports.
		{Code: `
        const Hello = React.memo(React.forwardRef((props, ref) => <div ref={ref} />));
      `, Tsx: true,
			Settings: map[string]interface{}{
				"react": map[string]interface{}{"version": "15.0.0"},
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: noDisplayName}},
		},

		// ---- Real-world: ES5 createReactClass without binding (immediate use) ----
		// `(createReactClass({...})).propTypes = ...` — IIFE-style use,
		// no binding. ES5 component IS detected (via `IsCreateReactClassObjectArg`).
		// Without displayName property → reports.
		{Code: `
        export default createReactClass({
          render: function() { return <div/>; }
        }).someProp;
      `, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: noDisplayName}},
		},

		// ---- Real-world: anonymous default-export arrow returning JSX ----
		// `export default () => <div/>` — default export of anonymous arrow,
		// no binding. Reports.
		{Code: `
        export default () => <div>Hello</div>;
      `, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: noDisplayName}},
		},

		// ---- TC-aware lock-in: cross-scope same-name binding disambiguation ----
		// Two `Inner` bindings in sibling scopes — only outerB's gets a
		// displayName assignment. Upstream's scope manager resolves the
		// `Inner.displayName = 'B'` reference to outerB's Inner, marks it,
		// and reports outerA's Inner (line 4). Without TC-aware resolution,
		// the syntactic `nameToComponent` would map "Inner" to whichever
		// was discovered first, leading to wrong-node reports. The TC path
		// in `resolveAndMarkComponentRef` keeps both directions correct;
		// the line/column assertion locks in the precise node.
		{Code: `
        function outerA() {
          const Inner = function() { return <div>A</div>; };
          return Inner;
        }
        function outerB() {
          const Inner = function() { return <div>B</div>; };
          Inner.displayName = 'B';
          return Inner;
        }
      `, Tsx: true, Options: map[string]interface{}{"ignoreTranspilerName": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: noDisplayName, Line: 3},
			},
		},

		// ---- Real-world: capitalized property-assignment arrow + anonymous default ----
		// `NS.Bar = () => <div/>` — property-assignment with capitalized
		// property name classifies (Branch 13). `export default function()`
		// — anonymous default export classifies. Both report.
		{Code: `
        const NS: { Bar?: () => JSX.Element } = {};
        NS.Bar = () => <div/>;
        export default function() { return <div/>; }
      `, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: noDisplayName},
				{MessageId: noDisplayName},
			},
		},

		// ---- Real-world: exporting anonymous arrow ----
		// `export default (props) => <div/>` — default export of anonymous
		// arrow. No transpiler name, no displayName property. Reports.
		{Code: `
        export default (props) => <div>{props.name}</div>;
      `, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: noDisplayName}},
		},

		// ---- Real-world: ignoreTranspilerName disables FN binding ----
		// `var Foo = () => <div/>` — transpiler name from binding `Foo`,
		// but `ignoreTranspilerName: true` requires explicit displayName.
		{Code: `
        var Hello = () => <div>Hello</div>;
      `, Tsx: true, Options: map[string]interface{}{"ignoreTranspilerName": true},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: noDisplayName}},
		},

		// ---- Real-world: ignoreTranspilerName disables class id ----
		// `class Foo extends React.Component {}` — has class id, but
		// `ignoreTranspilerName: true` requires explicit displayName.
		{Code: `
        class Hello extends React.Component {
          render() { return <div/>; }
        }
      `, Tsx: true, Options: map[string]interface{}{"ignoreTranspilerName": true},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: noDisplayName}},
		},

		// ---- Real-world: anonymous Cell-style arrow in JSX prop ----
		// `<Table cellRenderer={(value) => <div>{value}</div>}/>` — the
		// arrow is in JSX-attribute position, NOT an allowed component
		// position. NO detection. Pair with a real anonymous component.
		{Code: `
        const x = <Table cellRenderer={(value) => <div>{value}</div>}/>;
        export default () => <div>main</div>;
      `, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: noDisplayName}},
		},

		// ---- Real-world: anonymous arrow + propTypes assignment alone ----
		// Without setting displayName, only propTypes — anonymous wrapper
		// reports.
		{Code: `
        export const Hello = React.memo(({ x }) => <div>{x}</div>);
        Hello.propTypes = { x: () => null };
      `, Tsx: true, Options: map[string]interface{}{"ignoreTranspilerName": true},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: noDisplayName}},
		},

		// ---- Real-world: anonymous CallExpression of factory ----
		// `module.exports = createMyComponent({ render: () => <div/> })`
		// — not a recognized createReactClass; no React extends; no
		// pragma wrapper. The render callback's parent is PropertyAssignment
		// in a non-createReactClass object → not a component. The outer
		// shape isn't a component either. No detection on either.
		// Without a fallback component, no errors at all → not invalid.
		// Make this an invalid by adding a real anonymous arrow:
		{Code: `
        const helper = createMyComponent({ render: () => <div/> });
        export default function() { return <div/>; }
      `, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: noDisplayName}},
		},

		// ---- Real-world: createContext without React >= 16.3.0 setting falls back ----
		// Default React version (`999.999.999`) — context check IS active.
		// Without displayName, reports `noContextDisplayName`.
		{Code: `
        import { createContext } from 'react';
        const Hello = createContext({ defaultValue: 'world' });
      `, Tsx: true, Options: map[string]interface{}{"checkContextObjects": true},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: noContextDisplayName}},
		},

		// ---- Real-world: displayName key with ignored case-mismatch ----
		// `Foo.DisplayName = 'X'` — capitalized D doesn't match `displayName`.
		// upstream's `isDisplayNameDeclaration` is exact match. Should
		// NOT mark, so the component still reports.
		{Code: `
        function Hello() { return <div/>; }
        Hello.DisplayName = 'Hello';
      `, Tsx: true, Options: map[string]interface{}{"ignoreTranspilerName": true},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: noDisplayName}},
		},

		// ---- Real-world: displayName via different identifier name ----
		// `Foo.name = 'Foo'` — `name` ≠ `displayName`. Doesn't satisfy.
		{Code: `
        function Hello() { return <div/>; }
        Hello.name = 'Hello';
      `, Tsx: true, Options: map[string]interface{}{"ignoreTranspilerName": true},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: noDisplayName}},
		},

		// ---- Real-world: createContext via aliased import ----
		// `import { createContext as ctx } from 'react'; const X = ctx()`
		// — upstream's `isCreateContext` matches by callee NAME, not by
		// import resolution. Aliased `ctx` has callee.name = 'ctx', not
		// 'createContext'. So it's NOT recognized as createContext.
		// The X is just a value; not a context. With checkContextObjects
		// the regular rule's report is also avoided.
		// Make this valid (no diagnostic) to lock in the alias behavior:
		// SKIP — this is a VALID case, not invalid. Move test out of this section.

		// ---- Real-world: anonymous default-export class wrapping arrow ----
		// `export default class { static comp = () => <div/>; }` — class
		// is anonymous (no id), no displayName, extends nothing. Doesn't
		// classify as React component. Inner arrow `comp` is lowercase →
		// not a component either.
		// To make this invalid, add a real anonymous arrow:
		{Code: `
        export default class { static x = 1; }
        export const Hello = (props: any) => <div/>;
      `, Tsx: true, Options: map[string]interface{}{"ignoreTranspilerName": true},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: noDisplayName}},
		},
	})
}

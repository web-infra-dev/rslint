import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, upstreamModuleOptions } from '../config';
import { diagnostic } from '../expectations';

function expectNoDisplayNameError(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic(
    'noDisplayName',
    'Component definition is missing display name',
    range,
  );
}

function expectNoContextDisplayNameError(
  range: Parameters<typeof diagnostic>[2],
) {
  return diagnostic(
    'noContextDisplayName',
    'Context definition is missing display name',
    range,
  );
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('display-name', {
  valid: [
    // ---- Shadowed wrapper identifiers ----
    {
      code: `
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
      `,
    },
    {
      code: `
        import React, { memo, forwardRef } from 'react'

        const Test1 = function (memo) {
          return memo(() => <div>param shadowed</div>)
        }

        const Test2 = function ({ forwardRef }) {
          return forwardRef(() => <div>destructured param</div>)
        }
      `,
    },

    // ---- ES5 createReactClass with displayName + ignoreTranspilerName ----
    {
      code: `
        var Hello = createReactClass({
          displayName: 'Hello',
          render: function() {
            return <div>Hello {this.props.name}</div>;
          }
        });
      `,
      options: [{ ignoreTranspilerName: true }],
    },
    {
      code: `
        class Hello extends React.Component {
          render() {
            return <div>Hello {this.props.name}</div>;
          }
        }
        Hello.displayName = 'Hello'
      `,
      options: [{ ignoreTranspilerName: true }],
    },

    // ---- Class without React heritage ----
    {
      code: `
        class Hello {
          render() {
            return 'Hello World';
          }
        }
      `,
    },

    // ---- ES5 createReactClass — has transpiler name ----
    {
      code: `
        var Hello = createReactClass({
          render: function() {
            return <div>Hello {this.props.name}</div>;
          }
        });
      `,
    },
    {
      code: `
        class Hello extends React.Component {
          render() {
            return <div>Hello {this.props.name}</div>;
          }
        }
      `,
    },

    // ---- export default class with binding name ----
    {
      code: `
        export default class Hello {
          render() {
            return <div>Hello {this.props.name}</div>;
          }
        }
      `,
    },

    // ---- Reassignment of declared variable ----
    {
      code: `
        var Hello;
        Hello = createReactClass({
          render: function() {
            return <div>Hello {this.props.name}</div>;
          }
        });
      `,
    },

    // ---- module.exports with displayName property ----
    {
      code: `
        module.exports = createReactClass({
          "displayName": "Hello",
          "render": function() {
            return <div>Hello {this.props.name}</div>;
          }
        });
      `,
    },

    // ---- Anonymous default-exported class ----
    {
      code: `
        export default class {
          render() {
            return <div>Hello {this.props.name}</div>;
          }
        }
      `,
    },

    // ---- Named function expression inside React.memo ----
    {
      code: `
        export const Hello = React.memo(function Hello() {
          return <p />;
        })
      `,
    },

    // ---- Named function expressions / declarations / arrows with binding ----
    {
      code: `
        function Hello() {
          return <div>Hello {this.props.name}</div>;
        }
      `,
    },
    {
      code: `
        var Hello = () => {
          return <div>Hello {this.props.name}</div>;
        }
      `,
    },
    {
      code: `
        module.exports = function Hello() {
          return <div>Hello {this.props.name}</div>;
        }
      `,
    },

    // ---- Functional + Hello.displayName + ignoreTranspilerName ----
    {
      code: `
        function Hello() {
          return <div>Hello {this.props.name}</div>;
        }
        Hello.displayName = 'Hello';
      `,
      options: [{ ignoreTranspilerName: true }],
    },

    // ---- Deep MemberExpression displayName ----
    {
      code: `
        var Mixins = {
          Greetings: {
            Hello: function() {
              return <div>Hello {this.props.name}</div>;
            }
          }
        }
        Mixins.Greetings.Hello.displayName = 'Hello';
      `,
      options: [{ ignoreTranspilerName: true }],
    },

    // ---- ES5 createReactClass with helper render ----
    {
      code: `
        var Hello = createReactClass({
          render: function() {
            return <div>{this._renderHello()}</div>;
          },
          _renderHello: function() {
            return <span>Hello {this.props.name}</span>;
          }
        });
      `,
    },

    // ---- Object literal with shorthand method (Mixin Button) ----
    {
      code: `
        const Mixin = {
          Button() {
            return (
              <button />
            );
          }
        };
      `,
    },

    // ---- Component + propTypes + React.memo wrapping ----
    {
      code: `
        import React from 'react'
        import { string } from 'prop-types'

        function Component({ world }) {
          return <div>Hello {world}</div>
        }

        Component.propTypes = {
          world: string,
        }

        export default React.memo(Component)
      `,
    },

    // ---- React.memo wrapping named function ----
    {
      code: `
        import React from 'react'

        const ComponentWithMemo = React.memo(function Component({ world }) {
          return <div>Hello {world}</div>
        })
      `,
    },

    // ---- React.forwardRef wrapping named function ----
    {
      code: `
        import React from 'react'

        const ForwardRefComponentLike = React.forwardRef(function ComponentLike({ world }, ref) {
          return <div ref={ref}>Hello {world}</div>
        })
      `,
    },

    // ---- Comp.displayName + React.forwardRef ----
    {
      code: `
        const Comp = React.forwardRef((props, ref) => <main />);
        Comp.displayName = 'MyCompName';
      `,
    },
    {
      code: `
        const Comp = React.forwardRef((props, ref) => <main data-as="yes" />) as SomeComponent;
        Comp.displayName = 'MyCompNameAs';
      `,
    },

    // ---- Curried arrows / callbacks ----
    {
      code: `
        const f = (a) => () => {
          if (a) {
            return null;
          }
          return 1;
        };
      `,
    },

    // ---- React.memo with two args ----
    {
      code: `
        function MyComponent(props) {
          return <b>{props.name}</b>;
        }

        const MemoizedMyComponent = React.memo(
          MyComponent,
          (prevProps, nextProps) => prevProps.name === nextProps.name
        )
      `,
    },

    // ---- Nested memo+forwardRef accepted in supported React versions ----
    {
      code: `
        import React from 'react'

        const MemoizedForwardRefComponentLike = React.memo(
          React.forwardRef(({ world }, ref) => {
            return <div ref={ref}>Hello {world}</div>
          })
        )
      `,
      settings: {
        react: { version: '16.14.0' },
      },
    },

    // ---- checkContextObjects: true with explicit displayName ----
    {
      code: `
        import React from 'react';

        const Hello = React.createContext();
        Hello.displayName = "HelloContext"
      `,
      options: [{ checkContextObjects: true }],
    },
    {
      code: `
        import { createContext } from 'react';

        const Hello = createContext();
        Hello.displayName = "HelloContext"
      `,
      options: [{ checkContextObjects: true }],
    },
    {
      code: `
        import { createContext } from 'react';

        var Hello;
        Hello = createContext();
        Hello.displayName = "HelloContext";
      `,
      options: [{ checkContextObjects: true }],
    },

    // ---- React version too old for context check (silently disabled) ----
    {
      code: `
        import { createContext } from 'react';

        const Hello = createContext();
      `,
      settings: { react: { version: '16.2.0' } },
      options: [{ checkContextObjects: true }],
    },

    // ---- checkContextObjects: false leaves context unchecked ----
    {
      code: `
        import { createContext } from 'react';

        const Hello = createContext();
      `,
      options: [{ checkContextObjects: false }],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 1 (7.37.5)',
      code: `
        var Hello = React.createClass({
          displayName: 'Hello',
          render: function() {
            return <div>Hello {this.props.name}</div>;
          }
        });
      `,
      options: [{ ignoreTranspilerName: true }],
      settings: { react: { createClass: 'createClass' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 4 (7.37.5)',
      code: `
        class Hello extends Greetings {
          static text = 'Hello World';
          render() {
            return Hello.text;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 5 (7.37.5)',
      code: `
        class Hello {
          method;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 6 (7.37.5)',
      code: `
        class Hello extends React.Component {
          static get displayName() {
            return 'Hello';
          }
          render() {
            return <div>Hello {this.props.name}</div>;
          }
        }
      `,
      options: [{ ignoreTranspilerName: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 7 (7.37.5)',
      code: `
        class Hello extends React.Component {
          static displayName = 'Widget';
          render() {
            return <div>Hello {this.props.name}</div>;
          }
        }
      `,
      options: [{ ignoreTranspilerName: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 13 (7.37.5)',
      code: `
        var Hello = createReactClass({
          displayName: 'Hello',
          render: function() {
            let { a, ...b } = obj;
            let c = { ...d };
            return <div />;
          }
        });
      `,
      options: [{ ignoreTranspilerName: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 16 (7.37.5)',
      code: `
        var Hello = function() {
          return <div>Hello {this.props.name}</div>;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 21 (7.37.5)',
      code: `
        var Hello = () => {
          return <div>Hello {this.props.name}</div>;
        }
        Hello.displayName = 'Hello';
      `,
      options: [{ ignoreTranspilerName: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 22 (7.37.5)',
      code: `
        var Hello = function() {
          return <div>Hello {this.props.name}</div>;
        }
        Hello.displayName = 'Hello';
      `,
      options: [{ ignoreTranspilerName: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 25 (7.37.5)',
      code: `
        var Hello = createReactClass({
          displayName: 'Hello',
          render: function() {
            return <div>{this._renderHello()}</div>;
          },
          _renderHello: function() {
            return <span>Hello {this.props.name}</span>;
          }
        });
      `,
      options: [{ ignoreTranspilerName: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 27 (7.37.5)',
      code: `
        var obj = {
          pouf: function() {
            return any
          }
        };
      `,
      options: [{ ignoreTranspilerName: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 28 (7.37.5)',
      code: `
        var obj = {
          pouf: function() {
            return any
          }
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 29 (7.37.5)',
      code: `
        export default {
          renderHello() {
            let {name} = this.props;
            return <div>{name}</div>;
          }
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 30 (7.37.5)',
      code: `
        import React, { createClass } from 'react';
        export default createClass({
          displayName: 'Foo',
          render() {
            return <h1>foo</h1>;
          }
        });
      `,
      options: [{ ignoreTranspilerName: true }],
      settings: { react: { createClass: 'createClass' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 31 (7.37.5)',
      code: `
        import React, {Component} from "react";
        function someDecorator(ComposedComponent) {
          return class MyDecorator extends Component {
            render() {return <ComposedComponent {...this.props} />;}
          };
        }
        module.exports = someDecorator;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 32 (7.37.5)',
      code: `
        import React, {createElement} from "react";
        const SomeComponent = (props) => {
          const {foo, bar} = props;
          return someComponentFactory({
            onClick: () => foo(bar("x"))
          });
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 33 (7.37.5)',
      code: `
        const element = (
          <Media query={query} render={() => {
            renderWasCalled = true
            return <div/>
          }}/>
        )
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 34 (7.37.5)',
      code: `
        const element = (
          <Media query={query} render={function() {
            renderWasCalled = true
            return <div/>
          }}/>
        )
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 35 (7.37.5)',
      code: `
        module.exports = {
          createElement: tagName => document.createElement(tagName)
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 36 (7.37.5)',
      code: `
        const { createElement } = document;
        createElement("a");
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 39 (7.37.5)',
      code: `
        import React from 'react';

        const Hello = React.memo(function Hello() {
          return;
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 41 (7.37.5)',
      code: `
        function F() {
          let items = [];
          let testData = [
            {a: "test1", displayName: "test2"}, {a: "test1", displayName: "test2"}];
          for (let item of testData) {
              items.push({a: item.a, b: item.displayName});
          }
          return <div>{items}</div>;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 42 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream flow syntax (Property or signature expected.).',
      code: `
        import {Component} from "react";
        type LinkProps = {...{}};
        class Link extends Component<LinkProps> {}
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 43 (7.37.5)',
      code: `
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
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 44 (7.37.5)',
      code: `
        const renderer = a => function Component(listItem) {
          return <div>{a} {listItem}</div>;
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 47 (7.37.5)',
      code: `
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
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 49 (7.37.5)',
      code: `
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
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 50 (7.37.5)',
      code: `
        export const demo = (a) => (b) => {
          if (a == null) return null;
          return b;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 51 (7.37.5)',
      code: `
        let demo = null;
        demo = (a) => {
          if (a == null) return null;
          return f(a);
        };`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 52 (7.37.5)',
      code: `
        obj._property = (a) => {
          if (a == null) return null;
          return f(a);
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 53 (7.37.5)',
      code: `
        _variable = (a) => {
          if (a == null) return null;
          return f(a);
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 54 (7.37.5)',
      code: `
        demo = () => () => null;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 55 (7.37.5)',
      code: `
        demo = {
          property: () => () => null
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 56 (7.37.5)',
      code: `
        demo = function() {return function() {return null;};};
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 57 (7.37.5)',
      code: `
        demo = {
          property: function() {return function() {return null;};}
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 59 (7.37.5)',
      code: `
        import React from 'react'

        const MemoizedForwardRefComponentLike = React.memo(
          React.forwardRef(function({ world }, ref) {
            return <div ref={ref}>Hello {world}</div>
        })
        )
      `,
      settings: { react: { version: '16.14.0' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 60 (7.37.5)',
      code: `
        import React from 'react'

        const MemoizedForwardRefComponentLike = React.memo(
          React.forwardRef(({ world }, ref) => {
            return <div ref={ref}>Hello {world}</div>
          })
        )
      `,
      settings: { react: { version: '15.7.0' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 61 (7.37.5)',
      code: `
        import React from 'react'

        const MemoizedForwardRefComponentLike = React.memo(
          React.forwardRef(function ComponentLike({ world }, ref) {
            return <div ref={ref}>Hello {world}</div>
          })
        )
      `,
      settings: { react: { version: '16.12.1' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 62 (7.37.5)',
      code: `
        export const ComponentWithForwardRef = React.memo(
          React.forwardRef(function Component({ world }) {
            return <div>Hello {world}</div>
          })
        )
      `,
      settings: { react: { version: '0.14.11' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 63 (7.37.5)',
      code: `
        import React from 'react'

        const MemoizedForwardRefComponentLike = React.memo(
          React.forwardRef(function({ world }, ref) {
            return <div ref={ref}>Hello {world}</div>
          })
        )
      `,
      settings: { react: { version: '15.7.1' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 66 (7.37.5)',
      code: `
        import { createContext } from 'react';

        const Hello = createContext();

        const obj = {};
        obj.displayName = "False positive";

        Hello.displayName = "HelloContext"
      `,
      options: [{ checkContextObjects: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 67 (7.37.5)',
      code: `
        import * as React from 'react';

        const Hello = React.createContext();

        const obj = {};
        obj.displayName = "False positive";

        Hello.displayName = "HelloContext";
      `,
      options: [{ checkContextObjects: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 68 (7.37.5)',
      code: `
        const obj = {};
        obj.displayName = "False positive";
      `,
      options: [{ checkContextObjects: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 70 (7.37.5)',
      code: `
        import { createContext } from 'react';

        const Hello = createContext();
        Hello.displayName = "HelloContext";
      `,
      options: [{ checkContextObjects: true }],
      settings: { react: { version: '>16.3.0' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 71 (7.37.5)',
      code: `
        import { createContext } from 'react';

        let Hello;
        Hello = createContext();
        Hello.displayName = "HelloContext";
      `,
      options: [{ checkContextObjects: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 72 (7.37.5)',
      code: `
        import { createContext } from 'react';

        const Hello = createContext();
      `,
      options: [{ checkContextObjects: false }],
      settings: { react: { version: '>16.3.0' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 74 (7.37.5)',
      code: `
        import { createContext } from 'react';

        var Hello;
        Hello = React.createContext();
        Hello.displayName = "HelloContext";
      `,
      options: [{ checkContextObjects: true }],
      languageOptions: upstreamModuleOptions,
    },
  ],

  invalid: [
    {
      code: `
        /** @jsx Foo */
        var Hello = Foo.createClass({
          _renderHello: function() {
            return <span>Hello {this.props.name}</span>;
          },
          render: function() {
            return <div>{this._renderHello()}</div>;
          }
        });
      `,
      options: [{ ignoreTranspilerName: true }],
      settings: { react: { createClass: 'createClass' } },
      errors: [expectNoDisplayNameError([3, 37, 10, 10])],
    },
    // ---- Shadowed-but-only-some-paths shadowed ----
    {
      code: `
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
              return \`\${props} \${ref}\`
            })
          }

          return null
        }
      `,
      errors: [
        expectNoDisplayNameError([6, 36, 8, 15]),
        expectNoDisplayNameError([10, 31, 12, 15]),
        expectNoDisplayNameError([14, 37, 16, 15]),
      ],
    },

    // ---- ES5 createReactClass without displayName + ignoreTranspilerName ----
    {
      code: `
        var Hello = createReactClass({
          render: function() {
            return React.createElement("div", {}, "text content");
          }
        });
      `,
      options: [{ ignoreTranspilerName: true }],
      errors: [expectNoDisplayNameError([2, 38, 6, 10])],
    },

    // ---- Class without displayName + ignoreTranspilerName ----
    {
      code: `
        class Hello extends React.Component {
          render() {
            return <div>Hello {this.props.name}</div>;
          }
        }
      `,
      options: [{ ignoreTranspilerName: true }],
      errors: [expectNoDisplayNameError([2, 9, 6, 10])],
    },

    // ---- module.exports anonymous arrow / function ----
    {
      code: `
        module.exports = () => {
          return <div>Hello {props.name}</div>;
        }
      `,
      errors: [expectNoDisplayNameError([2, 26, 4, 10])],
    },
    {
      code: `
        module.exports = function() {
          return <div>Hello {props.name}</div>;
        }
      `,
      errors: [expectNoDisplayNameError([2, 26, 4, 10])],
    },

    // ---- module.exports of createReactClass without displayName ----
    {
      code: `
        module.exports = createReactClass({
          render() {
            return <div>Hello {this.props.name}</div>;
          }
        });
      `,
      errors: [expectNoDisplayNameError([2, 43, 6, 10])],
    },

    // ---- Higher-order anonymous function returning JSX ----
    {
      code: `
        function Hof() {
          return function () {
            return <div />
          }
        }
      `,
      errors: [expectNoDisplayNameError([3, 18, 5, 12])],
    },

    // ---- React.memo / React.forwardRef with anonymous arrow ----
    {
      code: `
        import React from 'react'

        const ComponentWithMemo = React.memo(({ world }) => {
          return <div>Hello {world}</div>
        })
      `,
      errors: [expectNoDisplayNameError([4, 35, 6, 11])],
    },
    {
      code: `
        import React from 'react'

        const ComponentWithMemo = React.memo(function() {
          return <div>Hello {world}</div>
        })
      `,
      errors: [expectNoDisplayNameError([4, 35, 6, 11])],
    },
    {
      code: `
        import React from 'react'

        const ForwardRefComponentLike = React.forwardRef(({ world }, ref) => {
          return <div ref={ref}>Hello {world}</div>
        })
      `,
      errors: [expectNoDisplayNameError([4, 41, 6, 11])],
    },

    // ---- Nested memo+forwardRef NOT supported in older React versions ----
    {
      code: `
        import React from 'react'

        const MemoizedForwardRefComponentLike = React.memo(
          React.forwardRef(({ world }, ref) => {
            return <div ref={ref}>Hello {world}</div>
          })
        )
      `,
      settings: {
        react: { version: '15.6.0' },
      },
      errors: [expectNoDisplayNameError([4, 49, 8, 10])],
    },

    // ---- Inner non-component functions don't shadow outer report ----
    {
      code: `
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
      `,
      errors: [expectNoDisplayNameError([2, 26, 14, 10])],
    },

    // ---- componentWrapperFunctions setting ----
    {
      code: `
        const processData = (options?: { value: string }) => options?.value || 'no data';

        export const Component = observer(() => {
          const data = processData({ value: 'data' });
          return <div>{data}</div>;
        });

        export const Component2 = observer(() => {
          const data = processData();
          return <div>{data}</div>;
        });
      `,
      settings: {
        componentWrapperFunctions: ['observer'],
      },
      errors: [
        expectNoDisplayNameError([4, 34, 7, 11]),
        expectNoDisplayNameError([9, 35, 12, 11]),
      ],
    },

    // ---- checkContextObjects: missing context displayName ----
    {
      code: `
        import React from 'react';

        const Hello = React.createContext();
      `,
      options: [{ checkContextObjects: true }],
      errors: [expectNoContextDisplayNameError([4, 15, 4, 44])],
    },
    {
      code: `
        import { createContext } from 'react';

        const Hello = createContext();
      `,
      options: [{ checkContextObjects: true }],
      errors: [expectNoContextDisplayNameError([4, 15, 4, 38])],
    },
    {
      code: `
        import { createContext } from 'react';

        var Hello;
        Hello = createContext();
      `,
      options: [{ checkContextObjects: true }],
      errors: [expectNoContextDisplayNameError([5, 9, 5, 33])],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 1 (7.37.5)',
      code: `
        var Hello = React.createClass({
          render: function() {
            return React.createElement("div", {}, "text content");
          }
        });
      `,
      options: [{ ignoreTranspilerName: true }],
      settings: { react: { createClass: 'createClass' } },
      errors: [expectNoDisplayNameError([2, 39, 6, 10])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 2 (7.37.5)',
      code: `
        var Hello = createReactClass({
          render: function() {
            return <div>Hello {this.props.name}</div>;
          }
        });
      `,
      options: [{ ignoreTranspilerName: true }],
      errors: [expectNoDisplayNameError([2, 38, 6, 10])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 4 (7.37.5)',
      code: `
        function HelloComponent() {
          return createReactClass({
            render: function() {
              return <div>Hello {this.props.name}</div>;
            }
          });
        }
        module.exports = HelloComponent();
      `,
      options: [{ ignoreTranspilerName: true }],
      errors: [expectNoDisplayNameError([3, 35, 7, 12])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 8 (7.37.5)',
      code: `
        var Hello = createReactClass({
          _renderHello: function() {
            return <span>Hello {this.props.name}</span>;
          },
          render: function() {
            return <div>{this._renderHello()}</div>;
          }
        });
      `,
      options: [{ ignoreTranspilerName: true }],
      errors: [expectNoDisplayNameError([2, 38, 9, 10])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 9 (7.37.5)',
      code: `
        var Hello = Foo.createClass({
          _renderHello: function() {
            return <span>Hello {this.props.name}</span>;
          },
          render: function() {
            return <div>{this._renderHello()}</div>;
          }
        });
      `,
      options: [{ ignoreTranspilerName: true }],
      settings: { react: { pragma: 'Foo', createClass: 'createClass' } },
      errors: [expectNoDisplayNameError([2, 37, 9, 10])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 11 (7.37.5)',
      code: `
        const Mixin = {
          Button() {
            return (
              <button />
            );
          }
        };
      `,
      options: [{ ignoreTranspilerName: true }],
      errors: [expectNoDisplayNameError([3, 11, 7, 12])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 13 (7.37.5)',
      code: `
        import React, { createElement } from "react";
        export default (props) => {
          return createElement("div", {}, "hello");
        };
      `,
      errors: [expectNoDisplayNameError([3, 24, 5, 10])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 17 (7.37.5)',
      code: `
        import React from 'react'

        const ForwardRefComponentLike = React.forwardRef(function({ world }, ref) {
          return <div ref={ref}>Hello {world}</div>
        })
      `,
      errors: [expectNoDisplayNameError([4, 41, 6, 11])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 19 (7.37.5)',
      code: `
        import React from 'react'

        const MemoizedForwardRefComponentLike = React.memo(
          React.forwardRef(function({ world }, ref) {
            return <div ref={ref}>Hello {world}</div>
          })
        )
      `,
      settings: { react: { version: '0.14.2' } },
      errors: [expectNoDisplayNameError([4, 49, 8, 10])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 20 (7.37.5)',
      code: `
        import React from 'react'

        const MemoizedForwardRefComponentLike = React.memo(
          React.forwardRef(function ComponentLike({ world }, ref) {
            return <div ref={ref}>Hello {world}</div>
          })
        )
      `,
      settings: { react: { version: '15.0.1' } },
      errors: [expectNoDisplayNameError([4, 49, 8, 10])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 21 (7.37.5)',
      code: `
        import React from "react";
        const { createElement } = React;
        export default (props) => {
          return createElement("div", {}, "hello");
        };
      `,
      errors: [expectNoDisplayNameError([4, 24, 6, 10])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 22 (7.37.5)',
      code: `
        import React from "react";
        const createElement = React.createElement;
        export default (props) => {
          return createElement("div", {}, "hello");
        };
      `,
      errors: [expectNoDisplayNameError([4, 24, 6, 10])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 24 (7.37.5)',
      code: `
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
      `,
      errors: [expectNoDisplayNameError([2, 26, 15, 10])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 25 (7.37.5)',
      code: `
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
      `,
      errors: [expectNoDisplayNameError([2, 24, 16, 10])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 26 (7.37.5)',
      code: `
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
      `,
      errors: [
        expectNoDisplayNameError([2, 24, 6, 10]),
        expectNoDisplayNameError([9, 18, 11, 13]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 27 (7.37.5)',
      code: `
        const renderer = a => listItem => (
          <div>{a} {listItem}</div>
        );
      `,
      errors: [expectNoDisplayNameError([2, 31, 4, 10])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 30 (7.37.5)',
      code: `
        import * as React from 'react';

        const Hello = React.createContext();
      `,
      options: [{ checkContextObjects: true }],
      errors: [expectNoContextDisplayNameError([4, 15, 4, 44])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 33 (7.37.5)',
      code: `
        import { createContext } from 'react';

        var Hello;
        Hello = React.createContext();
      `,
      options: [{ checkContextObjects: true }],
      errors: [expectNoContextDisplayNameError([5, 9, 5, 39])],
      languageOptions: upstreamModuleOptions,
    },
  ],
});

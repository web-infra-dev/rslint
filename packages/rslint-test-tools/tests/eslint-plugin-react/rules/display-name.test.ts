import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('display-name', {} as never, {
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
  ],

  invalid: [
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
        { messageId: 'noDisplayName' },
        { messageId: 'noDisplayName' },
        { messageId: 'noDisplayName' },
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
      errors: [{ messageId: 'noDisplayName' }],
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
      errors: [{ messageId: 'noDisplayName' }],
    },

    // ---- module.exports anonymous arrow / function ----
    {
      code: `
        module.exports = () => {
          return <div>Hello {props.name}</div>;
        }
      `,
      errors: [{ messageId: 'noDisplayName' }],
    },
    {
      code: `
        module.exports = function() {
          return <div>Hello {props.name}</div>;
        }
      `,
      errors: [{ messageId: 'noDisplayName' }],
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
      errors: [{ messageId: 'noDisplayName' }],
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
      errors: [{ messageId: 'noDisplayName' }],
    },

    // ---- React.memo / React.forwardRef with anonymous arrow ----
    {
      code: `
        import React from 'react'

        const ComponentWithMemo = React.memo(({ world }) => {
          return <div>Hello {world}</div>
        })
      `,
      errors: [{ messageId: 'noDisplayName' }],
    },
    {
      code: `
        import React from 'react'

        const ComponentWithMemo = React.memo(function() {
          return <div>Hello {world}</div>
        })
      `,
      errors: [{ messageId: 'noDisplayName' }],
    },
    {
      code: `
        import React from 'react'

        const ForwardRefComponentLike = React.forwardRef(({ world }, ref) => {
          return <div ref={ref}>Hello {world}</div>
        })
      `,
      errors: [{ messageId: 'noDisplayName' }],
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
      errors: [{ messageId: 'noDisplayName' }],
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
      errors: [{ messageId: 'noDisplayName' }],
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
      errors: [{ messageId: 'noDisplayName' }, { messageId: 'noDisplayName' }],
    },

    // ---- checkContextObjects: missing context displayName ----
    {
      code: `
        import React from 'react';

        const Hello = React.createContext();
      `,
      options: [{ checkContextObjects: true }],
      errors: [{ messageId: 'noContextDisplayName' }],
    },
    {
      code: `
        import { createContext } from 'react';

        const Hello = createContext();
      `,
      options: [{ checkContextObjects: true }],
      errors: [{ messageId: 'noContextDisplayName' }],
    },
    {
      code: `
        import { createContext } from 'react';

        var Hello;
        Hello = createContext();
      `,
      options: [{ checkContextObjects: true }],
      errors: [{ messageId: 'noContextDisplayName' }],
    },
  ],
});

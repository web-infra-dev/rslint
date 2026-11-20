import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, upstreamModuleOptions } from '../config';

function expectPropsNotSortedError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'propsNotSorted',
    message: 'Prop types declarations should be sorted alphabetically',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectCallbackPropsLastError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'callbackPropsLast',
    message: 'Callback prop types must be listed after all other prop types',
    line,
    column,
    endLine,
    endColumn,
  };
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('sort-prop-types', {
  valid: [
    {
      code: `Component.propTypes = { a: PropTypes.string, b: PropTypes.number };`,
    },
    {
      code: `Component.propTypes = { z: PropTypes.string, a: PropTypes.number };`,
      options: [{ noSortAlphabetically: true }],
    },
    {
      code: `Component.propTypes = { a: PropTypes.string, onChange: PropTypes.func };`,
      options: [{ callbacksLast: true }],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 0 (7.37.5)',
      code: `
        var First = createReactClass({
          render: function() {
            return <div />;
          }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 1 (7.37.5)',
      code: `
        var First = createReactClass({
          propTypes: externalPropTypes,
          render: function() {
            return <div />;
          }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 2 (7.37.5)',
      code: `
        var First = createReactClass({
          propTypes: {
            A: PropTypes.any,
            Z: PropTypes.string,
            a: PropTypes.any,
            z: PropTypes.string
          },
          render: function() {
            return <div />;
          }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 3 (7.37.5)',
      code: `
        var First = createReactClass({
          propTypes: {
            a: PropTypes.any,
            A: PropTypes.any,
            z: PropTypes.string,
            Z: PropTypes.string
          },
          render: function() {
            return <div />;
          }
        });
      `,
      options: [{ ignoreCase: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 4 (7.37.5)',
      code: `
        var First = createReactClass({
          propTypes: {
            a: PropTypes.any,
            z: PropTypes.string
          },
          render: function() {
            return <div />;
          }
        });
        var Second = createReactClass({
          propTypes: {
            AA: PropTypes.any,
            ZZ: PropTypes.string
          },
          render: function() {
            return <div />;
          }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 5 (7.37.5)',
      code: `
        class First extends React.Component {
          render() {
            return <div />;
          }
        }
        First.propTypes = {
          a: PropTypes.string,
          z: PropTypes.string
        };
        First.propTypes.justforcheck = PropTypes.string;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 6 (7.37.5)',
      code: `
        class First extends React.Component {
          render() {
            return <div />;
          }
        }
        First.propTypes = {
          a: PropTypes.any,
          A: PropTypes.any,
          z: PropTypes.string,
          Z: PropTypes.string
        };
      `,
      options: [{ ignoreCase: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 7 (7.37.5)',
      code: `
        class Component extends React.Component {
          static propTypes = {
            a: PropTypes.any,
            b: PropTypes.any,
            c: PropTypes.any
          };
          render() {
            return <div />;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 8 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            return <div>Hello</div>;
          }
        }
        Hello.propTypes = {
          "aria-controls": PropTypes.string
        };
      `,
      options: [{ ignoreCase: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 9 (7.37.5)',
      code: `
        var Hello = createReactClass({
          render: function() {
            let { a, ...b } = obj;
            let c = { ...d };
            return <div />;
          }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 10 (7.37.5)',
      code: `
        var First = createReactClass({
          propTypes: {
            barRequired: PropTypes.func.isRequired,
            onBar: PropTypes.func,
            z: PropTypes.any
          },
          render: function() {
            return <div />;
          }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 11 (7.37.5)',
      code: `
        var First = createReactClass({
          propTypes: {
            a: PropTypes.any,
            z: PropTypes.string,
            onBar: PropTypes.func,
            onFoo: PropTypes.func
          },
          render: function() {
            return <div />;
          }
        });
      `,
      options: [{ callbacksLast: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 12 (7.37.5)',
      code: `
        class Component extends React.Component {
          static propTypes = {
            a: PropTypes.any,
            z: PropTypes.string,
            onBar: PropTypes.func,
            onFoo: PropTypes.func
          };
          render() {
            return <div />;
          }
        }
      `,
      options: [{ callbacksLast: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 13 (7.37.5)',
      code: `
        class First extends React.Component {
          render() {
            return <div />;
          }
        }
        First.propTypes = {
            a: PropTypes.any,
            z: PropTypes.string,
            onBar: PropTypes.func,
            onFoo: PropTypes.func
        };
      `,
      options: [{ callbacksLast: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 14 (7.37.5)',
      code: `
        class First extends React.Component {
          render() {
            return <div />;
          }
        }
        First.propTypes = {
            barRequired: PropTypes.string.isRequired,
            a: PropTypes.any
        };
      `,
      options: [{ requiredFirst: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 15 (7.37.5)',
      code: `
        class First extends React.Component {
          render() {
            return <div />;
          }
        }
        First.propTypes = {
            fooRequired: MyPropType,
        };
      `,
      options: [{ requiredFirst: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 16 (7.37.5)',
      code: `
        class First extends React.Component {
          render() {
            return <div />;
          }
        }
        First.propTypes = {
            barRequired: PropTypes.string.isRequired,
            fooRequired: PropTypes.any.isRequired,
            a: PropTypes.any,
            z: PropTypes.string,
            onBar: PropTypes.func,
            onFoo: PropTypes.func
        };
      `,
      options: [{ requiredFirst: true, callbacksLast: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 17 (7.37.5)',
      code: `
        export default class ClassWithSpreadInPropTypes extends BaseClass {
          static propTypes = {
            b: PropTypes.string,
            ...c.propTypes,
            a: PropTypes.string
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 18 (7.37.5)',
      code: `
        const propTypes = require('./externalPropTypes')
        const TextFieldLabel = (props) => {
          return <div />;
        };
        TextFieldLabel.propTypes = propTypes;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 19 (7.37.5)',
      code: `
        const First = (props) => <div />;
        export const propTypes = {
            a: PropTypes.any,
            z: PropTypes.string,
        };
        First.propTypes = propTypes;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 20 (7.37.5)',
      code: `
        class Component extends React.Component {
          render() {
            return <div />;
          }
        }
        Component.propTypes = {
          x: PropTypes.any,
          y: PropTypes.any,
          z: PropTypes.shape({
            c: PropTypes.any,
            C: PropTypes.string,
            a: PropTypes.any,
            b: PropTypes.bool,
          }),
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 21 (7.37.5)',
      code: `
        class Component extends React.Component {
          render() {
            return <div />;
          }
        }
        Component.propTypes = {
          a: PropTypes.any,
          b: PropTypes.any,
          c: PropTypes.shape({
            c: PropTypes.any,
            ...otherPropTypes,
            a: PropTypes.any,
            b: PropTypes.bool,
          }),
        };
      `,
      options: [{ sortShapeProp: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 22 (7.37.5)',
      code: `
        class Component extends React.Component {
          render() {
            return <div />;
          }
        }
        Component.propTypes = {
          a: PropTypes.any,
          b: PropTypes.any,
          c: PropTypes.shape(
            importedPropType,
          ),
        };
      `,
      options: [{ sortShapeProp: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 23 (7.37.5)',
      code: `
        class Component extends React.Component {
          render() {
            return <div />;
          }
        }
        Component.propTypes = {
          a: PropTypes.any,
          z: PropTypes.any,
        };
      `,
      options: [{ noSortAlphabetically: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 24 (7.37.5)',
      code: `
        class Component extends React.Component {
          render() {
            return <div />;
          }
        }
        Component.propTypes = {
          z: PropTypes.any,
          a: PropTypes.any,
        };
      `,
      options: [{ noSortAlphabetically: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 25 (7.37.5)',
      code: `
        class Component extends React.Component {
          render() {
            return <div />;
          }
        }
        Component.propTypes = {
          0: PropTypes.any,
          1: PropTypes.any,
        };
      `,
      options: [{ ignoreCase: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 26 (7.37.5)',
      code: `
        const shape = {
          a: PropTypes.any,
          b: PropTypes.bool,
          c: PropTypes.any,
        };
        class Component extends React.Component {
          static propTypes = {
            x: PropTypes.shape(shape),
          };
          render() {
            return <div />;
          }
        }
      `,
      options: [{ sortShapeProp: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 27 (7.37.5)',
      code: `
        const shape = {
          a: PropTypes.any,
          b: PropTypes.bool,
          c: PropTypes.any,
        };
        class Component extends React.Component {
          render() {
            return <div />;
          }
        }
        Component.propTypes = {
          x: PropTypes.shape(shape)
        };
      `,
      options: [{ sortShapeProp: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 28 (7.37.5)',
      code: `
        var Component = createReactClass({
          propTypes: {
            a: React.PropTypes.string,
            c: React.PropTypes.string,
            b: React.PropTypes.string,
            onChange: React.PropTypes.func,
          }
        });
      `,
      options: [{ callbacksLast: true, noSortAlphabetically: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 29 (7.37.5)',
      code: `
        type Props = {
          zzz: string;
          aaa: string;
        }
        function Foo(props: Props) {
          return null;
        }
      `,
      options: [{ checkTypes: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 30 (7.37.5)',
      code: `
        function Foo() {
          return <div />;
        }
      `,
      options: [{ checkTypes: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 31 (7.37.5)',
      code: `
        const Foo = (props: {
          aaa: string,
          zzz: string
        }) => {
          return null;
        }
      `,
      options: [{ checkTypes: true }],
      languageOptions: upstreamModuleOptions,
    },
  ],
  invalid: [
    {
      code: `Component.propTypes = { z: PropTypes.string, a: PropTypes.number };`,
      errors: [expectPropsNotSortedError([1, 46, 1, 65])],
    },
    {
      code: `Component.propTypes = { onChange: PropTypes.func, name: PropTypes.string };`,
      options: [{ callbacksLast: true }],
      errors: [expectCallbackPropsLastError([1, 25, 1, 49])],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 0 (7.37.5)',
      code: `
        var First = createReactClass({
          propTypes: {
            z: PropTypes.string,
            a: PropTypes.any
          },
          render: function() {
            return <div />;
          }
        });
      `,
      errors: [expectPropsNotSortedError([5, 13, 5, 29])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 1 (7.37.5)',
      code: `
        var First = createReactClass({
          propTypes: {
            /* z */
            z: PropTypes.string,
            /* a */
            a: PropTypes.any
          },
          render: function() {
            return <div />;
          }
        });
      `,
      errors: [expectPropsNotSortedError([7, 13, 7, 29])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 2 (7.37.5)',
      code: `
        var First = createReactClass({
          propTypes: {
            z: PropTypes.any,
            Z: PropTypes.any
          },
          render: function() {
            return <div />;
          }
        });
      `,
      errors: [expectPropsNotSortedError([5, 13, 5, 29])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 3 (7.37.5)',
      code: `
        var First = createReactClass({
          propTypes: {
            Z: PropTypes.any,
            a: PropTypes.any
          },
          render: function() {
            return <div />;
          }
        });
      `,
      options: [{ ignoreCase: true }],
      errors: [expectPropsNotSortedError([5, 13, 5, 29])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 4 (7.37.5)',
      code: `
        var First = createReactClass({
          propTypes: {
            a: PropTypes.any,
            A: PropTypes.any,
            z: PropTypes.string,
            Z: PropTypes.string
          },
          render: function() {
            return <div />;
          }
        });
      `,
      errors: [
        expectPropsNotSortedError([5, 13, 5, 29]),
        expectPropsNotSortedError([7, 13, 7, 32]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 5 (7.37.5)',
      code: `
        var First = createReactClass({
          propTypes: {
            a: PropTypes.any,
            Zz: PropTypes.string
          },
          render: function() {
            return <div />;
          }
        });
        var Second = createReactClass({
          propTypes: {
            aAA: PropTypes.any,
            ZZ: PropTypes.string
          },
          render: function() {
            return <div />;
          }
        });
      `,
      errors: [
        expectPropsNotSortedError([5, 13, 5, 33]),
        expectPropsNotSortedError([14, 13, 14, 33]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 6 (7.37.5)',
      code: `
        class First extends React.Component {
          render() {
            return <div />;
          }
        }
        First.propTypes = {
            yy: PropTypes.any,
            bb: PropTypes.string
        };
        class Second extends React.Component {
          render() {
            return <div />;
          }
        }
        Second.propTypes = {
            aAA: PropTypes.any,
            ZZ: PropTypes.string
        };
      `,
      errors: [
        expectPropsNotSortedError([9, 13, 9, 33]),
        expectPropsNotSortedError([18, 13, 18, 33]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 7 (7.37.5)',
      code: `
        class Component extends React.Component {
          static propTypes = {
            z: PropTypes.any,
            y: PropTypes.any,
            a: PropTypes.any
          };
          render() {
            return <div />;
          }
        }
      `,
      errors: [
        expectPropsNotSortedError([5, 13, 5, 29]),
        expectPropsNotSortedError([6, 13, 6, 29]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 8 (7.37.5)',
      code: `
        class Component extends React.Component {
          static propTypes = forbidExtraProps({
            z: PropTypes.any,
            y: PropTypes.any,
            a: PropTypes.any
          });
          render() {
            return <div />;
          }
        }
      `,
      settings: { propWrapperFunctions: ['forbidExtraProps'] },
      errors: [
        expectPropsNotSortedError([5, 13, 5, 29]),
        expectPropsNotSortedError([6, 13, 6, 29]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 9 (7.37.5)',
      code: `
        var First = createReactClass({
          propTypes: {
            a: PropTypes.any,
            z: PropTypes.string,
            onFoo: PropTypes.func,
            onBar: PropTypes.func
          },
          render: function() {
            return <div />;
          }
        });
      `,
      options: [{ callbacksLast: true }],
      errors: [expectPropsNotSortedError([7, 13, 7, 34])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 10 (7.37.5)',
      code: `
        class Component extends React.Component {
          static propTypes = {
            a: PropTypes.any,
            z: PropTypes.string,
            onFoo: PropTypes.func,
            onBar: PropTypes.func
          };
          render() {
            return <div />;
          }
        }
      `,
      options: [{ callbacksLast: true }],
      errors: [expectPropsNotSortedError([7, 13, 7, 34])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 11 (7.37.5)',
      code: `
        class First extends React.Component {
          render() {
            return <div />;
          }
        }
        First.propTypes = {
            a: PropTypes.any,
            z: PropTypes.string,
            onFoo: PropTypes.func,
            onBar: PropTypes.func
        };
      `,
      options: [{ callbacksLast: true }],
      errors: [expectPropsNotSortedError([11, 13, 11, 34])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 12 (7.37.5)',
      code: `
        class First extends React.Component {
          render() {
            return <div />;
          }
        }
        First.propTypes = forbidExtraProps({
            a: PropTypes.any,
            z: PropTypes.string,
            onFoo: PropTypes.func,
            onBar: PropTypes.func
        });
      `,
      options: [{ callbacksLast: true }],
      settings: { propWrapperFunctions: ['forbidExtraProps'] },
      errors: [expectPropsNotSortedError([11, 13, 11, 34])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 13 (7.37.5)',
      code: `
        const First = (props) => <div />;
        const propTypes = {
            z: PropTypes.string,
            a: PropTypes.any,
        };
        First.propTypes = forbidExtraProps(propTypes);
      `,
      settings: { propWrapperFunctions: ['forbidExtraProps'] },
      errors: [expectPropsNotSortedError([5, 13, 5, 29])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 14 (7.37.5)',
      code: `
        const First = (props) => <div />;
        const propTypes = {
            z: PropTypes.string,
            a: PropTypes.any,
        };
        First.propTypes = propTypes;
      `,
      settings: { propWrapperFunctions: ['forbidExtraProps'] },
      errors: [expectPropsNotSortedError([5, 13, 5, 29])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 15 (7.37.5)',
      code: `
        var First = createReactClass({
          propTypes: {
            a: PropTypes.any,
            onBar: PropTypes.func,
            onFoo: PropTypes.func,
            z: PropTypes.string
          },
          render: function() {
            return <div />;
          }
        });
      `,
      options: [{ callbacksLast: true }],
      errors: [expectCallbackPropsLastError([6, 13, 6, 34])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 16 (7.37.5)',
      code: `
        var First = createReactClass({
          propTypes: {
            fooRequired: PropTypes.string.isRequired,
            barRequired: PropTypes.string.isRequired,
            a: PropTypes.any
          },
          render: function() {
            return <div />;
          }
        });
      `,
      options: [{ requiredFirst: true }],
      errors: [expectPropsNotSortedError([5, 13, 5, 53])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 17 (7.37.5)',
      code: `
        var First = createReactClass({
          propTypes: {
            a: PropTypes.any,
            barRequired: PropTypes.string.isRequired,
            onFoo: PropTypes.func
          },
          render: function() {
            return <div />;
          }
        });
      `,
      options: [{ requiredFirst: true }],
      errors: [
        {
          messageId: 'requiredPropsFirst',
          message:
            'Required prop types must be listed before all other prop types',
          line: 5,
          column: 13,
          endLine: 5,
          endColumn: 53,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 18 (7.37.5)',
      code: `
        export default class ClassWithSpreadInPropTypes extends BaseClass {
          static propTypes = {
            b: PropTypes.string,
            ...a.propTypes,
            d: PropTypes.string,
            c: PropTypes.string
          }
        }
      `,
      errors: [expectPropsNotSortedError([7, 13, 7, 32])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 19 (7.37.5)',
      code: `
        export default class ClassWithSpreadInPropTypes extends BaseClass {
          static propTypes = {
            b: PropTypes.string,
            ...a.propTypes,
            f: PropTypes.string,
            d: PropTypes.string,
            ...e.propTypes,
            c: PropTypes.string
          }
        }
      `,
      errors: [expectPropsNotSortedError([7, 13, 7, 32])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 20 (7.37.5)',
      code: `
        const propTypes = {
          b: PropTypes.string,
          a: PropTypes.string,
        };
        const TextFieldLabel = (props) => {
          return <div />;
        };
        TextFieldLabel.propTypes = propTypes;
      `,
      errors: [expectPropsNotSortedError([4, 11, 4, 30])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 21 (7.37.5)',
      code: `
        class Component extends React.Component {
          render() {
            return <div />;
          }
        }
        Component.propTypes = {
          x: PropTypes.any,
          y: PropTypes.any,
          z: PropTypes.shape({
            c: PropTypes.any,
            a: PropTypes.any,
            b: PropTypes.bool,
          }),
        };
      `,
      options: [{ sortShapeProp: true }],
      errors: [
        expectPropsNotSortedError([12, 13, 12, 29]),
        expectPropsNotSortedError([13, 13, 13, 30]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 22 (7.37.5)',
      code: `
        class Component extends React.Component {
          render() {
            return <div />;
          }
        }
        Component.propTypes = {
          x: PropTypes.any,
          z: PropTypes.shape(),
          y: PropTypes.any,
        };
      `,
      options: [{ sortShapeProp: true }],
      errors: [expectPropsNotSortedError([10, 11, 10, 27])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 23 (7.37.5)',
      code: `
        class Component extends React.Component {
          render() {
            return <div />;
          }
        }
        Component.propTypes = {
          x: PropTypes.any,
          z: PropTypes.shape(someType),
          y: PropTypes.any,
        };
      `,
      options: [{ sortShapeProp: true }],
      errors: [expectPropsNotSortedError([10, 11, 10, 27])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 24 (7.37.5)',
      code: `
        class Component extends React.Component {
          render() {
            return <div />;
          }
        }
        Component.propTypes = {
          z: PropTypes.any,
          y: PropTypes.any,
          a: PropTypes.shape({
            c: PropTypes.any,
            C: PropTypes.string,
            a: PropTypes.any,
            b: PropTypes.bool,
          }),
        };
      `,
      options: [{ sortShapeProp: true }],
      errors: [
        expectPropsNotSortedError([9, 11, 9, 27]),
        expectPropsNotSortedError([10, 11, 15, 13]),
        expectPropsNotSortedError([12, 13, 12, 32]),
        expectPropsNotSortedError([13, 13, 13, 29]),
        expectPropsNotSortedError([14, 13, 14, 30]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 25 (7.37.5)',
      code: `
        class Component extends React.Component {
          render() {
            return <div />;
          }
        }
        Component.propTypes = {
          x: PropTypes.any,
          y: PropTypes.any,
          z: PropTypes.shape({
            c: PropTypes.any,
            C: PropTypes.string,
            a: PropTypes.any,
            b: PropTypes.bool,
          }),
        };
      `,
      options: [{ sortShapeProp: true, ignoreCase: true }],
      errors: [
        expectPropsNotSortedError([13, 13, 13, 29]),
        expectPropsNotSortedError([14, 13, 14, 30]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 26 (7.37.5)',
      code: `
        class Component extends React.Component {
          render() {
            return <div />;
          }
        }
        Component.propTypes = {
          x: PropTypes.any,
          y: PropTypes.any,
          z: PropTypes.shape({
            a: PropTypes.string,
            c: PropTypes.number.isRequired,
            b: PropTypes.any,
            d: PropTypes.bool,
          }),
        };
      `,
      options: [{ sortShapeProp: true, requiredFirst: true }],
      errors: [
        {
          messageId: 'requiredPropsFirst',
          message:
            'Required prop types must be listed before all other prop types',
          line: 12,
          column: 13,
          endLine: 12,
          endColumn: 43,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 27 (7.37.5)',
      code: `
        class Component extends React.Component {
          render() {
            return <div />;
          }
        }
        Component.propTypes = {
          x: PropTypes.any,
          y: PropTypes.any,
          z: PropTypes.shape({
            a: PropTypes.string,
            c: PropTypes.number.isRequired,
            b: PropTypes.any,
            onFoo: PropTypes.func,
            d: PropTypes.bool,
          }),
        };
      `,
      options: [{ sortShapeProp: true, callbacksLast: true }],
      errors: [
        expectPropsNotSortedError([13, 13, 13, 29]),
        expectCallbackPropsLastError([14, 13, 14, 34]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 28 (7.37.5)',
      code: `
        class Component extends React.Component {
          render() {
            return <div />;
          }
        }
        Component.propTypes = {
          x: PropTypes.any,
          y: PropTypes.any,
          z: PropTypes.shape({
            a: PropTypes.string,
            c: PropTypes.number.isRequired,
            b: PropTypes.any,
            ...otherPropTypes,
            f: PropTypes.bool,
            d: PropTypes.string,
          }),
        };
      `,
      options: [{ sortShapeProp: true }],
      errors: [
        expectPropsNotSortedError([13, 13, 13, 29]),
        expectPropsNotSortedError([16, 13, 16, 32]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 29 (7.37.5)',
      code: `
        class Component extends React.Component {
          static propTypes = {
            z: PropTypes.any,
            y: PropTypes.any,
            a: PropTypes.shape({
              c: PropTypes.any,
              a: PropTypes.any,
              b: PropTypes.bool,
            }),
          };
          render() {
            return <div />;
          }
        }
      `,
      options: [{ sortShapeProp: true }],
      errors: [
        expectPropsNotSortedError([5, 13, 5, 29]),
        expectPropsNotSortedError([6, 13, 10, 15]),
        expectPropsNotSortedError([8, 15, 8, 31]),
        expectPropsNotSortedError([9, 15, 9, 32]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 30 (7.37.5)',
      code: `
        var First = createReactClass({
          propTypes: {
            z: PropTypes.string,
            a: PropTypes.any
          },
          render: function() {
            return <div />;
          }
        });
      `,
      options: [{ noSortAlphabetically: false }],
      errors: [expectPropsNotSortedError([5, 13, 5, 29])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 31 (7.37.5)',
      code: `
        var First = createReactClass({
          propTypes: {
            'data-letter': PropTypes.string,
            a: PropTypes.any,
            e: PropTypes.any
          },
          render: function() {
            return <div />;
          }
        });
      `,
      options: [{ noSortAlphabetically: false }],
      errors: [expectPropsNotSortedError([5, 13, 5, 29])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 32 (7.37.5)',
      code: `
        class Component extends React.Component {
          render() {
            return <div />;
          }
        }
        Component.propTypes = {
          1: PropTypes.any,
          0: PropTypes.any,
        };
      `,
      options: [{ ignoreCase: true }],
      errors: [expectPropsNotSortedError([9, 11, 9, 27])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 33 (7.37.5)',
      code: `
        const shape = {
          c: PropTypes.any,
          a: PropTypes.any,
          b: PropTypes.bool,
        };
        class Component extends React.Component {
          static propTypes = {
            x: PropTypes.shape(shape),
          };

          render() {
            return <div />;
          }
        }
      `,
      options: [{ sortShapeProp: true }],
      errors: [
        expectPropsNotSortedError([4, 11, 4, 27]),
        expectPropsNotSortedError([5, 11, 5, 28]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 34 (7.37.5)',
      code: `
        const shape = {
          c: PropTypes.any,
          a: PropTypes.any,
          b: PropTypes.bool,
        };
        class Component extends React.Component {
          render() {
            return <div />;
          }
        }
        Component.propTypes = {
          x: PropTypes.shape(shape)
        };
      `,
      options: [{ sortShapeProp: true }],
      errors: [
        expectPropsNotSortedError([4, 11, 4, 27]),
        expectPropsNotSortedError([5, 11, 5, 28]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 35 (7.37.5)',
      code: `
        var Component = React.createClass({
          propTypes: {
            onChange: React.PropTypes.func,
            a: React.PropTypes.string,
            c: React.PropTypes.string,
            b: React.PropTypes.string,
          }
        });
      `,
      options: [{ callbacksLast: true }],
      errors: [expectCallbackPropsLastError([4, 13, 4, 43])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 36 (7.37.5)',
      code: `
        var Component = createReactClass({
          propTypes: {
            onChange: React.PropTypes.func,
            a: React.PropTypes.string,
            c: React.PropTypes.string,
            b: React.PropTypes.string,
          }
        });
      `,
      options: [{ callbacksLast: true, noSortAlphabetically: true }],
      errors: [expectCallbackPropsLastError([4, 13, 4, 43])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 37 (7.37.5)',
      code: `
        var First = createReactClass({
          propTypes: {
            z: PropTypes.string /* z */,
            a: PropTypes.any /* a */,
            b: PropTypes.any /* b */
          },
          render: function() {
            return <div />;
          }
        });
      `,
      errors: [
        expectPropsNotSortedError([5, 13, 5, 29]),
        expectPropsNotSortedError([6, 13, 6, 29]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 38 (7.37.5)',
      code: `
        var First = createReactClass({
          propTypes: {
            /* z */ z: PropTypes.string,
            /* a */ a: PropTypes.any,
            /* b */ b: PropTypes.any
          },
          render: function() {
            return <div />;
          }
        });
      `,
      errors: [
        expectPropsNotSortedError([5, 21, 5, 37]),
        expectPropsNotSortedError([6, 21, 6, 37]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 39 (7.37.5)',
      code: `
        var First = createReactClass({
          propTypes: {
            /* z */ z: PropTypes.string /* z */,
            /* a */ a: PropTypes.any /* a */,
            /* b */ b: PropTypes.any /* b */
          },
          render: function() {
            return <div />;
          }
        });
      `,
      errors: [
        expectPropsNotSortedError([5, 21, 5, 37]),
        expectPropsNotSortedError([6, 21, 6, 37]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 40 (7.37.5)',
      code: `
        var First = createReactClass({
          propTypes: {
            /* z */ z: PropTypes.string, /* a */ a: PropTypes.any, /* b */ b: PropTypes.any
          },
          render: function() {
            return <div />;
          }
        });
      `,
      errors: [
        expectPropsNotSortedError([4, 50, 4, 66]),
        expectPropsNotSortedError([4, 76, 4, 92]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 41 (7.37.5)',
      code: `
        class Component extends React.Component {
          render() {
            return <div />;
          }
        }
        Component.propTypes = {
          x: PropTypes.any,
          y: PropTypes.any,
          z: PropTypes.shape({
            a: PropTypes.string,
            c: PropTypes.number.isRequired /* c */,
            b: PropTypes.any,
            ...otherPropTypes,
            f: PropTypes.bool,
            /* d */
            d: PropTypes.string,
          }),
        };
      `,
      options: [{ sortShapeProp: true }],
      errors: [
        expectPropsNotSortedError([13, 13, 13, 29]),
        expectPropsNotSortedError([17, 13, 17, 32]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 42 (7.37.5)',
      code: `
        var First = createReactClass({
          propTypes: {
            /* z */
            /* z */
            z: PropTypes.string /* z */,
            /* a */
            a: PropTypes.any /* a */
            /* a */
            /* a */,
            b: PropTypes.any
          },
          render: function() {
            return <div />;
          }
        });
      `,
      errors: [
        expectPropsNotSortedError([8, 13, 8, 29]),
        expectPropsNotSortedError([11, 13, 11, 29]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 43 (7.37.5)',
      code: `
        var Component = createReactClass({
          propTypes: {
            /* onChange */ onChange: React.PropTypes.func,
            /* a */ a: React.PropTypes.string,
            /* c */ c: React.PropTypes.string,
            /* b */ b: React.PropTypes.string,
          }
        });
      `,
      options: [{ callbacksLast: true, noSortAlphabetically: true }],
      errors: [expectCallbackPropsLastError([4, 28, 4, 58])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 44 (7.37.5)',
      code: `
        var Component = createReactClass({
          propTypes: {
            /* onChange */ onChange: React.PropTypes.func /* onChange */,
            /* a */ a: React.PropTypes.string /* a */,
            /* c */ c: React.PropTypes.string /* c */,
            /* b */ b: React.PropTypes.string /* b */,
          }
        });
      `,
      options: [{ callbacksLast: true, noSortAlphabetically: true }],
      errors: [expectCallbackPropsLastError([4, 28, 4, 58])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 45 (7.37.5)',
      code: `
        type Props = {
          zzz: string;
          aaa: string;
        }
        function Foo(props: Props) {
          return null;
        }
      `,
      options: [{ checkTypes: true }],
      errors: [expectPropsNotSortedError([4, 11, 4, 23])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 46 (7.37.5)',
      code: `
        type Props = {
          zzz: string;
          aaa: string;
        }
        const Foo = (props: Props) => {
          return null;
        }
      `,
      options: [{ checkTypes: true }],
      errors: [expectPropsNotSortedError([4, 11, 4, 23])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 47 (7.37.5)',
      code: `
        const Foo = (props: {
          zzz: string,
          aaa: string,
        }) => {
          return null;
        }
      `,
      options: [{ checkTypes: true }],
      errors: [expectPropsNotSortedError([4, 11, 4, 23])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 48 (7.37.5)',
      code: `
        type CustomProps = { onChange: () => void; name: string };
        const Foo = (props: CustomProps) => {
          return null;
        }
      `,
      options: [{ checkTypes: true }],
      errors: [expectPropsNotSortedError([2, 52, 2, 64])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 49 (7.37.5)',
      code: `
        type CustomProps = { onChange: (event: { target: { name: string; value: string } }) => void; name: string };
        const Foo = (props: CustomProps) => {
          return null;
        }
      `,
      options: [{ checkTypes: true }],
      errors: [expectPropsNotSortedError([2, 102, 2, 114])],
      languageOptions: upstreamModuleOptions,
    },
  ],
});

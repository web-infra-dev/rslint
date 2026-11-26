import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, upstreamModuleOptions } from '../config';

function expectNotStaticClassPropError1(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'notStaticClassProp',
    message: "'propTypes' should be declared as a static class property.",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectNotGetterClassFuncError2(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'notGetterClassFunc',
    message:
      "'propTypes' should be declared as a static getter class function.",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectDeclareOutsideClassError3(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'declareOutsideClass',
    message: "'displayName' should be declared outside the class body.",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectDeclareOutsideClassError4(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'declareOutsideClass',
    message: "'propTypes' should be declared outside the class body.",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectNotStaticClassPropError5(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'notStaticClassProp',
    message:
      "'childContextTypes' should be declared as a static class property.",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectNotStaticClassPropError6(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'notStaticClassProp',
    message: "'contextTypes' should be declared as a static class property.",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectNotStaticClassPropError7(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'notStaticClassProp',
    message: "'contextType' should be declared as a static class property.",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectNotStaticClassPropError8(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'notStaticClassProp',
    message: "'displayName' should be declared as a static class property.",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectNotStaticClassPropError9(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'notStaticClassProp',
    message: "'defaultProps' should be declared as a static class property.",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectDeclareOutsideClassError10(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'declareOutsideClass',
    message: "'childContextTypes' should be declared outside the class body.",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectDeclareOutsideClassError11(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'declareOutsideClass',
    message: "'contextTypes' should be declared outside the class body.",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectDeclareOutsideClassError12(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'declareOutsideClass',
    message: "'contextType' should be declared outside the class body.",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectDeclareOutsideClassError13(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'declareOutsideClass',
    message: "'defaultProps' should be declared outside the class body.",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectNotGetterClassFuncError14(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'notGetterClassFunc',
    message:
      "'childContextTypes' should be declared as a static getter class function.",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectNotGetterClassFuncError15(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'notGetterClassFunc',
    message:
      "'contextTypes' should be declared as a static getter class function.",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectNotGetterClassFuncError16(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'notGetterClassFunc',
    message:
      "'contextType' should be declared as a static getter class function.",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectNotGetterClassFuncError17(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'notGetterClassFunc',
    message:
      "'displayName' should be declared as a static getter class function.",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectNotGetterClassFuncError18(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'notGetterClassFunc',
    message:
      "'defaultProps' should be declared as a static getter class function.",
    line,
    column,
    endLine,
    endColumn,
  };
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('static-property-placement', {
  valid: [
    {
      code: `class MyComponent extends React.Component {
        static propTypes = {};
        static defaultProps = {};
        static contextType = Context;
        static contextTypes = {};
        static childContextTypes = {};
        static displayName = 'MyComponent';
      }`,
    },
    {
      code: `class MyComponent extends React.Component {
        static get propTypes() { return {}; }
        static get defaultProps() { return {}; }
        static get contextType() { return Context; }
      }`,
      options: ['static getter'],
    },
    {
      code: `class MyComponent extends React.Component {}
        MyComponent.propTypes = {};
        MyComponent.defaultProps = {};`,
      options: ['property assignment'],
    },
    {
      code: `const Box = { C: class extends React.Component {} };
        Box.C = {};
        Box.C.propTypes = {};`,
    },
    {
      code: `const Box = { C: class extends React.Component {} } as const;
        Box.C.propTypes = {};`,
    },
    {
      code: `/** @extends {React.Component} */
        class C { propTypes = {}; }`,
      filename: 'files/static-property-placement.jsx',
    },
    {
      code: `/** @extends React.Component */
        const C = class {
          // eslint-disable-next-line react/static-property-placement
          propTypes = {};
        };`,
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 0 (7.37.5)',
      code: `
        var MyComponent = createReactClass({
          childContextTypes: {
            something: PropTypes.bool
          },

          contextTypes: {
            something: PropTypes.bool
          },

          getDefaultProps: function() {
            name: 'Bob'
          },

          displayName: 'Hello',

          propTypes: {
            something: PropTypes.bool
          },

          render: function() {
            return null;
          },
        });
      `,
      options: ['property assignment'],
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 1 (7.37.5)',
      code: `
        var MyComponent = React.createClass({
          childContextTypes: {
            something: PropTypes.bool
          },

          contextTypes: {
            something: PropTypes.bool
          },

          getDefaultProps: function() {
            name: 'Bob'
          },

          displayName: 'Hello',

          propTypes: {
            something: PropTypes.bool
          },

          render: function() {
            return null;
          },
        });
      `,
      options: ['property assignment'],
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 2 (7.37.5)',
      code: `
        const MyComponent = () => {
            return <div>Hello</div>;
        };

        MyComponent.childContextTypes = {
          something: PropTypes.bool
        };

        MyComponent.contextTypes = {
          something: PropTypes.bool
        };

        MyComponent.defaultProps = {
          something: 'Bob'
        };

        MyComponent.displayName = 'Hello';

        MyComponent.propTypes = {
          something: PropTypes.bool
        };
      `,
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 3 (7.37.5)',
      code: `
        const MyComponent = () => (<div>Hello</div>);

        MyComponent.childContextTypes = {
          something: PropTypes.bool
        };

        MyComponent.contextTypes = {
          something: PropTypes.bool
        };

        MyComponent.defaultProps = {
          something: 'Bob'
        };

        MyComponent.displayName = 'Hello';

        MyComponent.propTypes = {
          something: PropTypes.bool
        };
      `,
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 4 (7.37.5)',
      code: `
        export function MyComponent () {
            return <div>Hello</div>;
        };

        MyComponent.childContextTypes = {
          something: PropTypes.bool
        };

        MyComponent.contextTypes = {
          something: PropTypes.bool
        };

        MyComponent.defaultProps = {
          something: 'Bob'
        };

        MyComponent.displayName = 'Hello';

        MyComponent.propTypes = {
          something: PropTypes.bool
        };
      `,
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 5 (7.37.5)',
      code: `
        class Foo {
          static get propTypes() {}
        }
      `,
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 6 (7.37.5)',
      code: `
        class Foo {
          static propTypes = {}
        }
      `,
      options: ['property assignment'],
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 7 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          render() {
            return null;
          }
        }
      `,
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 8 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          static randomlyNamed = {
            name: 'random'
          }
        }
      `,
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 9 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          static randomlyNamed = {
            name: 'random'
          }
        }
      `,
      options: ['property assignment'],
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 10 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          render() {
            return null;
          }
        }

        MyComponent.randomlyNamed = {
          name: 'random'
        }
      `,
      options: ['property assignment'],
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 11 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          render() {
            return null;
          }
        }

        MyComponent.randomlyNamed = {
          name: 'random'
        }
      `,
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 12 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          static childContextTypes = {
            something: PropTypes.bool
          };
        }
      `,
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 13 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          static childContextTypes = {
            something: PropTypes.bool
          };
        }
      `,
      options: [
        'property assignment',
        { childContextTypes: 'static public field' },
      ],
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 14 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          static get childContextTypes() {
            return {
              something: PropTypes.bool
            };
          }
        }
      `,
      options: ['static getter'],
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 15 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          static get childContextTypes() {
            return {
              something: PropTypes.bool
            };
          }
        }
      `,
      options: ['property assignment', { childContextTypes: 'static getter' }],
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 16 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          render() {
            return null;
          }
        }

        MyComponent.childContextTypes = {
          name: PropTypes.string.isRequired
        }
      `,
      options: ['property assignment'],
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 17 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          render() {
            return null;
          }
        }

        MyComponent.childContextTypes = {
          name: PropTypes.string.isRequired
        }
      `,
      options: [
        'static public field',
        { childContextTypes: 'property assignment' },
      ],
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 18 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          static contextTypes = {
            something: PropTypes.bool
          };
        }
      `,
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 19 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          static contextTypes = {
            something: PropTypes.bool
          };
        }
      `,
      options: ['property assignment', { contextTypes: 'static public field' }],
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 20 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          static get contextTypes() {
            return {
              something: PropTypes.bool
            };
          }
        }
      `,
      options: ['static getter'],
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 21 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          static get contextTypes() {
            return {
              something: PropTypes.bool
            };
          }
        }
      `,
      options: ['property assignment', { contextTypes: 'static getter' }],
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 22 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          render() {
            return null;
          }
        }

        MyComponent.contextTypes = {
          name: PropTypes.string.isRequired
        }
      `,
      options: ['property assignment'],
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 23 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          render() {
            return null;
          }
        }

        MyComponent.contextTypes = {
          name: PropTypes.string.isRequired
        }
      `,
      options: ['static public field', { contextTypes: 'property assignment' }],
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 24 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          static contextType = MyContext;
        }
      `,
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 25 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          static contextType = MyContext;
        }
      `,
      options: ['property assignment', { contextType: 'static public field' }],
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 26 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          static get contextType() {
             return MyContext;
          }
        }
      `,
      options: ['static getter'],
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 27 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          static get contextType() {
             return MyContext;
          }
        }
      `,
      options: ['property assignment', { contextType: 'static getter' }],
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 28 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          render() {
            return null;
          }
        }

        MyComponent.contextType = MyContext;
      `,
      options: ['property assignment'],
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 29 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          render() {
            return null;
          }
        }

        MyComponent.contextType = MyContext;
      `,
      options: ['static public field', { contextType: 'property assignment' }],
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 30 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          static displayName = "Hello";
        }
      `,
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 31 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          static displayName = "Hello";
        }
      `,
      options: ['property assignment', { displayName: 'static public field' }],
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 32 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          static get displayName() {
            return "Hello";
          }
        }
      `,
      options: ['static getter'],
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 33 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          static get displayName() {
            return "Hello";
          }
        }
      `,
      options: ['property assignment', { displayName: 'static getter' }],
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 34 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          render() {
            return null;
          }
        }

        MyComponent.displayName = "Hello";
      `,
      options: ['property assignment'],
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 35 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          render() {
            return null;
          }
        }

        MyComponent.displayName = "Hello";
      `,
      options: ['static public field', { displayName: 'property assignment' }],
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 36 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          static defaultProps = {
            something: 'Bob'
          };
        }
      `,
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 37 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          static defaultProps = {
            something: 'Bob'
          };
        }
      `,
      options: ['property assignment', { defaultProps: 'static public field' }],
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 38 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          static get defaultProps() {
            return {
              something: 'Bob'
            };
          }
        }
      `,
      options: ['static getter'],
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 39 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          static get defaultProps() {
            return {
              something: 'Bob'
            };
          }
        }
      `,
      options: ['property assignment', { defaultProps: 'static getter' }],
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 40 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          render() {
            return null;
          }
        }

        MyComponent.defaultProps = {
          name: 'Bob'
        }
      `,
      options: ['property assignment'],
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 41 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          render() {
            return null;
          }
        }

        MyComponent.defaultProps = {
          name: 'Bob'
        }
      `,
      options: ['static public field', { defaultProps: 'property assignment' }],
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 42 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          static propTypes = {
            something: PropTypes.bool
          };
        }
      `,
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 43 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          static propTypes = {
            something: PropTypes.bool
          };
        }
      `,
      options: ['property assignment', { propTypes: 'static public field' }],
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 44 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          static get propTypes() {
            return {
              something: PropTypes.bool
            };
          }
        }
      `,
      options: ['static getter'],
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 45 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          static get propTypes() {
            return {
              something: PropTypes.bool
            };
          }
        }
      `,
      options: ['property assignment', { propTypes: 'static getter' }],
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 46 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          render() {
            return null;
          }
        }

        MyComponent.propTypes = {
          name: PropTypes.string.isRequired
        }
      `,
      options: ['property assignment'],
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 47 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          render() {
            return null;
          }
        }

        MyComponent.propTypes = {
          name: PropTypes.string.isRequired
        }
      `,
      options: ['static public field', { propTypes: 'property assignment' }],
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 48 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          static childContextTypes = {
            something: PropTypes.bool
          };

          static contextTypes = {
            something: PropTypes.bool
          };

          static contextType = MyContext;

          static displayName = "Hello";

          static defaultProps = {
            something: 'Bob'
          };

          static propTypes = {
            something: PropTypes.bool
          };
        }
      `,
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 49 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          static childContextTypes = {
            something: PropTypes.bool
          };

          static contextTypes = {
            something: PropTypes.bool
          };

          static contextType = MyContext;

          static displayName = "Hello";

          static defaultProps = {
            something: 'Bob'
          };

          static propTypes = {
            something: PropTypes.bool
          };
        }
      `,
      options: [
        'property assignment',
        {
          childContextTypes: 'static public field',
          contextTypes: 'static public field',
          contextType: 'static public field',
          displayName: 'static public field',
          defaultProps: 'static public field',
          propTypes: 'static public field',
        },
      ],
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 50 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          static get childContextTypes() {
            return {
              something: PropTypes.bool
            };
          }

          static get contextTypes() {
            return {
              something: PropTypes.bool
            };
          }

          static get contextType() {
            return MyContext;
          }

          static get displayName() {
            return "Hello";
          }

          static get defaultProps() {
            return {
              something: PropTypes.bool
            };
          }

          static get propTypes() {
            return {
              something: PropTypes.bool
            };
          }
        }
      `,
      options: ['static getter'],
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 51 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          static get childContextTypes() {
            return {
              something: PropTypes.bool
            };
          }

          static get contextTypes() {
            return {
              something: PropTypes.bool
            };
          }

          static get contextType() {
            return MyContext;
          }

          static get displayName() {
            return "Hello";
          }

          static get defaultProps() {
            return {
              something: PropTypes.bool
            };
          }

          static get propTypes() {
            return {
              something: PropTypes.bool
            };
          }
        }
      `,
      options: [
        'property assignment',
        {
          childContextTypes: 'static getter',
          contextTypes: 'static getter',
          contextType: 'static getter',
          displayName: 'static getter',
          defaultProps: 'static getter',
          propTypes: 'static getter',
        },
      ],
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 52 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          render() {
            return null;
          }
        }

        MyComponent.childContextTypes = {
          name: PropTypes.string.isRequired
        }

        MyComponent.contextTypes = {
          name: PropTypes.string.isRequired
        }

        MyComponent.displayName = "Hello";

        MyComponent.defaultProps = {
          name: 'Bob'
        }

        MyComponent.propTypes = {
          name: PropTypes.string.isRequired
        }
      `,
      options: ['property assignment'],
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 53 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          render() {
            return null;
          }
        }

        MyComponent.childContextTypes = {
          name: PropTypes.string.isRequired
        }

        MyComponent.contextTypes = {
          name: PropTypes.string.isRequired
        }

        MyComponent.displayName = "Hello";

        MyComponent.defaultProps = {
          name: 'Bob'
        }

        MyComponent.propTypes = {
          name: PropTypes.string.isRequired
        }
      `,
      options: [
        'static public field',
        {
          childContextTypes: 'property assignment',
          contextTypes: 'property assignment',
          displayName: 'property assignment',
          defaultProps: 'property assignment',
          propTypes: 'property assignment',
        },
      ],
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 54 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          static childContextTypes = {
            name: PropTypes.string.isRequired
          }

          static contextTypes = {
            name: PropTypes.string.isRequired
          }

          static get displayName() {
            return "Hello"
          }
        }

        MyComponent.defaultProps = {
          name: 'Bob'
        }

        MyComponent.propTypes = {
          name: PropTypes.string.isRequired
        }
      `,
      options: [
        'static public field',
        {
          displayName: 'static getter',
          defaultProps: 'property assignment',
          propTypes: 'property assignment',
        },
      ],
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 55 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          static childContextTypes = {
            name: PropTypes.string.isRequired
          }

          static contextTypes = {
            name: PropTypes.string.isRequired
          }

          static get displayName() {
            return "Hello"
          }
        }

        MyComponent.defaultProps = {
          name: 'Bob'
        }

        MyComponent.propTypes = {
          name: PropTypes.string.isRequired
        }
      `,
      options: [
        'property assignment',
        {
          childContextTypes: 'static public field',
          contextTypes: 'static public field',
          displayName: 'static getter',
        },
      ],
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 56 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          static childContextTypes = {
            name: PropTypes.string.isRequired
          }

          static contextTypes = {
            name: PropTypes.string.isRequired
          }

          static displayName = "Hello";
        }

        const OtherComponent = () => (<div>Hello</div>);

        OtherComponent.defaultProps = {
          name: 'Bob'
        }

        OtherComponent.propTypes = {
          name: PropTypes.string.isRequired
        }
      `,
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 57 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          static childContextTypes = {
            name: PropTypes.string.isRequired
          }

          static contextTypes = {
            name: PropTypes.string.isRequired
          }

          static displayName = "Hello";
        }

        class OtherComponent extends React.Component {
          static defaultProps = {
            name: 'Bob'
          }

          static propTypes = {
            name: PropTypes.string.isRequired
          }
        }
      `,
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 58 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          static displayName = "Hello";

          myMethod() {
            console.log(MyComponent.displayName);
          }
        }
      `,
      options: ['static public field'],
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 59 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          static displayName = "Hello";

          myMethod() {
            MyComponent.displayName = "Bonjour";
          }
        }
      `,
      options: ['static public field'],
      settings: { react: { version: '15' } },
      languageOptions: upstreamModuleOptions,
    },
  ],
  invalid: [
    {
      code: `class MyComponent extends React.Component {
        static get propTypes() { return {}; }
      }`,
      errors: [expectNotStaticClassPropError1([2, 9, 2, 46])],
    },
    {
      code: `class MyComponent extends React.Component {
        static propTypes = {};
      }`,
      options: ['static getter'],
      errors: [expectNotGetterClassFuncError2([2, 9, 2, 31])],
    },
    {
      code: `class MyComponent extends React.Component {}
        MyComponent.propTypes = {};`,
      errors: [expectNotStaticClassPropError1([2, 9, 2, 30])],
    },
    {
      code: `class MyComponent extends React.Component {
        static displayName = 'x';
        static get propTypes() { return {}; }
      }
      MyComponent.defaultProps = {};`,
      options: ['property assignment'],
      errors: [
        expectDeclareOutsideClassError3([2, 9, 2, 34]),
        expectDeclareOutsideClassError4([3, 9, 3, 46]),
      ],
    },
    {
      code: `/** @extends React.Component */
        const C = class { propTypes = {}; };`,
      filename: 'files/static-property-placement.jsx',
      errors: [expectNotStaticClassPropError1([2, 27, 2, 42])],
    },
    {
      code: `const Box = { C: class extends React.Component {} };
        (Box).C.propTypes = {};`,
      options: ['static getter'],
      errors: [expectNotGetterClassFuncError2([2, 9, 2, 26])],
    },
    {
      code: `/* @jsx Preact.h */
        class C extends Preact.Component { propTypes = {}; }`,
      errors: [expectNotStaticClassPropError1([2, 44, 2, 59])],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 0 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          render() {
            return null;
          }
        }

        MyComponent.childContextTypes = {
          name: PropTypes.string.isRequired
        }

        MyComponent.contextTypes = {
          name: PropTypes.string.isRequired
        }

        MyComponent.contextType = MyContext;

        MyComponent.displayName = "Hello";

        MyComponent.defaultProps = {
          name: 'Bob'
        }

        MyComponent.propTypes = {
          name: PropTypes.string.isRequired
        }
      `,
      settings: { react: { version: '15' } },
      errors: [
        expectNotStaticClassPropError5([8, 9, 8, 38]),
        expectNotStaticClassPropError6([12, 9, 12, 33]),
        expectNotStaticClassPropError7([16, 9, 16, 32]),
        expectNotStaticClassPropError8([18, 9, 18, 32]),
        expectNotStaticClassPropError9([20, 9, 20, 33]),
        expectNotStaticClassPropError1([24, 9, 24, 30]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 1 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          render() {
            return null;
          }
        }

        MyComponent.childContextTypes = {
          name: PropTypes.string.isRequired
        }

        MyComponent.contextTypes = {
          name: PropTypes.string.isRequired
        }

        MyComponent.contextType = MyContext;

        MyComponent.displayName = "Hello";

        MyComponent.defaultProps = {
          name: 'Bob'
        }

        MyComponent.propTypes = {
          name: PropTypes.string.isRequired
        }
      `,
      options: [
        'property assignment',
        {
          childContextTypes: 'static public field',
          contextTypes: 'static public field',
          contextType: 'static public field',
          displayName: 'static public field',
          defaultProps: 'static public field',
          propTypes: 'static public field',
        },
      ],
      settings: { react: { version: '15' } },
      errors: [
        expectNotStaticClassPropError5([8, 9, 8, 38]),
        expectNotStaticClassPropError6([12, 9, 12, 33]),
        expectNotStaticClassPropError7([16, 9, 16, 32]),
        expectNotStaticClassPropError8([18, 9, 18, 32]),
        expectNotStaticClassPropError9([20, 9, 20, 33]),
        expectNotStaticClassPropError1([24, 9, 24, 30]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 2 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          static get childContextTypes() {
            return {
              something: PropTypes.bool
            };
          }

          static get contextTypes() {
            return {
              something: PropTypes.bool
            };
          }

          static get contextType() {
            return MyContext;
          }

          static get displayName() {
            return "Hello";
          }

          static get defaultProps() {
            return {
              something: PropTypes.bool
            };
          }

          static get propTypes() {
            return {
              something: PropTypes.bool
            };
          }
        }
      `,
      settings: { react: { version: '15' } },
      errors: [
        expectNotStaticClassPropError5([3, 11, 7, 12]),
        expectNotStaticClassPropError6([9, 11, 13, 12]),
        expectNotStaticClassPropError7([15, 11, 17, 12]),
        expectNotStaticClassPropError8([19, 11, 21, 12]),
        expectNotStaticClassPropError9([23, 11, 27, 12]),
        expectNotStaticClassPropError1([29, 11, 33, 12]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 3 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          static get childContextTypes() {
            return {
              something: PropTypes.bool
            };
          }

          static get contextTypes() {
            return {
              something: PropTypes.bool
            };
          }

          static get contextType() {
            return MyContext;
          }

          static get displayName() {
            return "Hello";
          }

          static get defaultProps() {
            return {
              something: PropTypes.bool
            };
          }

          static get propTypes() {
            return {
              something: PropTypes.bool
            };
          }
        }
      `,
      options: [
        'property assignment',
        {
          childContextTypes: 'static public field',
          contextTypes: 'static public field',
          contextType: 'static public field',
          displayName: 'static public field',
          defaultProps: 'static public field',
          propTypes: 'static public field',
        },
      ],
      settings: { react: { version: '15' } },
      errors: [
        expectNotStaticClassPropError5([3, 11, 7, 12]),
        expectNotStaticClassPropError6([9, 11, 13, 12]),
        expectNotStaticClassPropError7([15, 11, 17, 12]),
        expectNotStaticClassPropError8([19, 11, 21, 12]),
        expectNotStaticClassPropError9([23, 11, 27, 12]),
        expectNotStaticClassPropError1([29, 11, 33, 12]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 4 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          static childContextTypes = {
            something: PropTypes.bool
          };

          static contextTypes = {
            something: PropTypes.bool
          };

          static contextType = MyContext;

          static displayName = "Hello";

          static defaultProps = {
            something: 'Bob'
          };

          static propTypes = {
            something: PropTypes.bool
          };
        }
      `,
      options: ['property assignment'],
      settings: { react: { version: '15' } },
      errors: [
        expectDeclareOutsideClassError10([3, 11, 5, 13]),
        expectDeclareOutsideClassError11([7, 11, 9, 13]),
        expectDeclareOutsideClassError12([11, 11, 11, 42]),
        expectDeclareOutsideClassError3([13, 11, 13, 40]),
        expectDeclareOutsideClassError13([15, 11, 17, 13]),
        expectDeclareOutsideClassError4([19, 11, 21, 13]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 5 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          static childContextTypes = {
            something: PropTypes.bool
          };

          static contextTypes = {
            something: PropTypes.bool
          };

          static contextType = MyContext;

          static displayName = "Hello";

          static defaultProps = {
            something: 'Bob'
          };

          static propTypes = {
            something: PropTypes.bool
          };
        }
      `,
      options: [
        'static public field',
        {
          childContextTypes: 'property assignment',
          contextTypes: 'property assignment',
          contextType: 'property assignment',
          displayName: 'property assignment',
          defaultProps: 'property assignment',
          propTypes: 'property assignment',
        },
      ],
      settings: { react: { version: '15' } },
      errors: [
        expectDeclareOutsideClassError10([3, 11, 5, 13]),
        expectDeclareOutsideClassError11([7, 11, 9, 13]),
        expectDeclareOutsideClassError12([11, 11, 11, 42]),
        expectDeclareOutsideClassError3([13, 11, 13, 40]),
        expectDeclareOutsideClassError13([15, 11, 17, 13]),
        expectDeclareOutsideClassError4([19, 11, 21, 13]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 6 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          static get childContextTypes() {
            return {
              something: PropTypes.bool
            };
          }

          static get contextTypes() {
            return {
              something: PropTypes.bool
            };
          }

          static get contextType() {
            return MyContext;
          }

          static get displayName() {
            return "Hello";
          }

          static get defaultProps() {
            return {
              something: PropTypes.bool
            };
          }

          static get propTypes() {
            return {
              something: PropTypes.bool
            };
          }
        }
      `,
      options: ['property assignment'],
      settings: { react: { version: '15' } },
      errors: [
        expectDeclareOutsideClassError10([3, 11, 7, 12]),
        expectDeclareOutsideClassError11([9, 11, 13, 12]),
        expectDeclareOutsideClassError12([15, 11, 17, 12]),
        expectDeclareOutsideClassError3([19, 11, 21, 12]),
        expectDeclareOutsideClassError13([23, 11, 27, 12]),
        expectDeclareOutsideClassError4([29, 11, 33, 12]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 7 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          static get childContextTypes() {
            return {
              something: PropTypes.bool
            };
          }

          static get contextTypes() {
            return {
              something: PropTypes.bool
            };
          }

          static get contextType() {
            return MyContext;
          }

          static get displayName() {
            return "Hello";
          }

          static get defaultProps() {
            return {
              something: PropTypes.bool
            };
          }

          static get propTypes() {
            return {
              something: PropTypes.bool
            };
          }
        }
      `,
      options: [
        'static getter',
        {
          childContextTypes: 'property assignment',
          contextTypes: 'property assignment',
          contextType: 'property assignment',
          displayName: 'property assignment',
          defaultProps: 'property assignment',
          propTypes: 'property assignment',
        },
      ],
      settings: { react: { version: '15' } },
      errors: [
        expectDeclareOutsideClassError10([3, 11, 7, 12]),
        expectDeclareOutsideClassError11([9, 11, 13, 12]),
        expectDeclareOutsideClassError12([15, 11, 17, 12]),
        expectDeclareOutsideClassError3([19, 11, 21, 12]),
        expectDeclareOutsideClassError13([23, 11, 27, 12]),
        expectDeclareOutsideClassError4([29, 11, 33, 12]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 8 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          static childContextTypes = {
            something: PropTypes.bool
          };

          static contextTypes = {
            something: PropTypes.bool
          };

          static contextType = MyContext;

          static displayName = "Hello";

          static defaultProps = {
            something: 'Bob'
          };

          static propTypes = {
            something: PropTypes.bool
          };
        }
      `,
      options: ['static getter'],
      settings: { react: { version: '15' } },
      errors: [
        expectNotGetterClassFuncError14([3, 11, 5, 13]),
        expectNotGetterClassFuncError15([7, 11, 9, 13]),
        expectNotGetterClassFuncError16([11, 11, 11, 42]),
        expectNotGetterClassFuncError17([13, 11, 13, 40]),
        expectNotGetterClassFuncError18([15, 11, 17, 13]),
        expectNotGetterClassFuncError2([19, 11, 21, 13]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 9 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          static childContextTypes = {
            something: PropTypes.bool
          };

          static contextTypes = {
            something: PropTypes.bool
          };

          static contextType = MyContext;

          static displayName = "Hello";

          static defaultProps = {
            something: 'Bob'
          };

          static propTypes = {
            something: PropTypes.bool
          };
        }
      `,
      options: [
        'static public field',
        {
          childContextTypes: 'static getter',
          contextTypes: 'static getter',
          contextType: 'static getter',
          displayName: 'static getter',
          defaultProps: 'static getter',
          propTypes: 'static getter',
        },
      ],
      settings: { react: { version: '15' } },
      errors: [
        expectNotGetterClassFuncError14([3, 11, 5, 13]),
        expectNotGetterClassFuncError15([7, 11, 9, 13]),
        expectNotGetterClassFuncError16([11, 11, 11, 42]),
        expectNotGetterClassFuncError17([13, 11, 13, 40]),
        expectNotGetterClassFuncError18([15, 11, 17, 13]),
        expectNotGetterClassFuncError2([19, 11, 21, 13]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 10 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          render() {
            return null;
          }
        }

        MyComponent.childContextTypes = {
          name: PropTypes.string.isRequired
        }

        MyComponent.contextTypes = {
          name: PropTypes.string.isRequired
        }

        MyComponent.contextType = MyContext;

        MyComponent.displayName = "Hello";

        MyComponent.defaultProps = {
          name: 'Bob'
        }

        MyComponent.propTypes = {
          name: PropTypes.string.isRequired
        }
      `,
      options: ['static getter'],
      settings: { react: { version: '15' } },
      errors: [
        expectNotGetterClassFuncError14([8, 9, 8, 38]),
        expectNotGetterClassFuncError15([12, 9, 12, 33]),
        expectNotGetterClassFuncError16([16, 9, 16, 32]),
        expectNotGetterClassFuncError17([18, 9, 18, 32]),
        expectNotGetterClassFuncError18([20, 9, 20, 33]),
        expectNotGetterClassFuncError2([24, 9, 24, 30]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 11 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          render() {
            return null;
          }
        }

        MyComponent.childContextTypes = {
          name: PropTypes.string.isRequired
        }

        MyComponent.contextTypes = {
          name: PropTypes.string.isRequired
        }

        MyComponent.contextType = MyContext;

        MyComponent.displayName = "Hello";

        MyComponent.defaultProps = {
          name: 'Bob'
        }

        MyComponent.propTypes = {
          name: PropTypes.string.isRequired
        }
      `,
      options: [
        'property assignment',
        {
          childContextTypes: 'static getter',
          contextTypes: 'static getter',
          contextType: 'static getter',
          displayName: 'static getter',
          defaultProps: 'static getter',
          propTypes: 'static getter',
        },
      ],
      settings: { react: { version: '15' } },
      errors: [
        expectNotGetterClassFuncError14([8, 9, 8, 38]),
        expectNotGetterClassFuncError15([12, 9, 12, 33]),
        expectNotGetterClassFuncError16([16, 9, 16, 32]),
        expectNotGetterClassFuncError17([18, 9, 18, 32]),
        expectNotGetterClassFuncError18([20, 9, 20, 33]),
        expectNotGetterClassFuncError2([24, 9, 24, 30]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 12 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          static childContextTypes = {
            name: PropTypes.string.isRequired
          }

          static contextTypes = {
            name: PropTypes.string.isRequired
          }

          static contextType = MyContext;

          static get displayName() {
            return "Hello";
          }
        }

        MyComponent.defaultProps = {
          name: 'Bob'
        }

        MyComponent.propTypes = {
          name: PropTypes.string.isRequired
        }
      `,
      options: [
        'property assignment',
        {
          defaultProps: 'static getter',
          propTypes: 'static public field',
          displayName: 'static public field',
        },
      ],
      settings: { react: { version: '15' } },
      errors: [
        expectDeclareOutsideClassError10([3, 11, 5, 12]),
        expectDeclareOutsideClassError11([7, 11, 9, 12]),
        expectDeclareOutsideClassError12([11, 11, 11, 42]),
        expectNotStaticClassPropError8([13, 11, 15, 12]),
        expectNotGetterClassFuncError18([18, 9, 18, 33]),
        expectNotStaticClassPropError1([22, 9, 22, 30]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 13 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          static childContextTypes = {
            name: PropTypes.string.isRequired
          }

          static contextTypes = {
            name: PropTypes.string.isRequired
          }

          static contextType = MyContext;

          static get displayName() {
            return "Hello";
          }
        }

        MyComponent.defaultProps = {
          name: 'Bob'
        }

        MyComponent.propTypes = {
          name: PropTypes.string.isRequired
        }
      `,
      options: [
        'static getter',
        {
          childContextTypes: 'property assignment',
          contextTypes: 'property assignment',
          contextType: 'property assignment',
          displayName: 'property assignment',
        },
      ],
      settings: { react: { version: '15' } },
      errors: [
        expectDeclareOutsideClassError10([3, 11, 5, 12]),
        expectDeclareOutsideClassError11([7, 11, 9, 12]),
        expectDeclareOutsideClassError12([11, 11, 11, 42]),
        expectDeclareOutsideClassError3([13, 11, 15, 12]),
        expectNotGetterClassFuncError18([18, 9, 18, 33]),
        expectNotGetterClassFuncError2([22, 9, 22, 30]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 14 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          static childContextTypes = {
            name: PropTypes.string.isRequired
          }

          static contextTypes = {
            name: PropTypes.string.isRequired
          }

          static contextType = MyContext;

          static get displayName() {
            return "Hello";
          }
        }

        const OtherComponent = () => (<div>Hello</div>);

        OtherComponent.defaultProps = {
          name: 'Bob'
        }

        OtherComponent.propTypes = {
          name: PropTypes.string.isRequired
        }
      `,
      options: [
        'property assignment',
        { defaultProps: 'static public field', propTypes: 'static getter' },
      ],
      settings: { react: { version: '15' } },
      errors: [
        expectDeclareOutsideClassError10([3, 11, 5, 12]),
        expectDeclareOutsideClassError11([7, 11, 9, 12]),
        expectDeclareOutsideClassError12([11, 11, 11, 42]),
        expectDeclareOutsideClassError3([13, 11, 15, 12]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 15 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          static childContextTypes = {
            name: PropTypes.string.isRequired
          }

          static contextTypes = {
            name: PropTypes.string.isRequired
          }

          static contextType = MyContext;

          static displayName = "Hello";
        }

        class OtherComponent extends React.Component {
          static contextTypes = {
            name: PropTypes.string.isRequired
          }

          static defaultProps = {
            name: 'Bob'
          }

          static propTypes = {
            name: PropTypes.string.isRequired
          }

          static get displayName() {
            return "Hello";
          }
        }
      `,
      options: ['property assignment'],
      settings: { react: { version: '15' } },
      errors: [
        expectDeclareOutsideClassError10([3, 11, 5, 12]),
        expectDeclareOutsideClassError11([7, 11, 9, 12]),
        expectDeclareOutsideClassError12([11, 11, 11, 42]),
        expectDeclareOutsideClassError3([13, 11, 13, 40]),
        expectDeclareOutsideClassError11([17, 11, 19, 12]),
        expectDeclareOutsideClassError13([21, 11, 23, 12]),
        expectDeclareOutsideClassError4([25, 11, 27, 12]),
        expectDeclareOutsideClassError3([29, 11, 31, 12]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 16 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          displayName = 'Foo';
        }
      `,
      options: ['static public field'],
      settings: { react: { version: '15' } },
      errors: [expectNotStaticClassPropError8([3, 11, 3, 31])],
      languageOptions: upstreamModuleOptions,
    },
  ],
});

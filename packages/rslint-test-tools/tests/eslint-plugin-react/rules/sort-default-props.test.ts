import { RuleTester } from '../rule-tester';

// Migrated from eslint-plugin-react v7.37.5.
const ruleTester = new RuleTester();
ruleTester.run('sort-default-props', {} as never, {
  valid: [
    {
      code: '\n        var First = createReactClass({\n          render: function() {\n            return <div />;\n          }\n        });\n      ',
    },
    {
      code: '\n        var First = createReactClass({\n          propTypes: {\n            A: PropTypes.any,\n            Z: PropTypes.string,\n            a: PropTypes.any,\n            z: PropTypes.string\n          },\n          getDefaultProps: function() {\n            return {\n              A: "A",\n              Z: "Z",\n              a: "a",\n              z: "z"\n            };\n          },\n          render: function() {\n            return <div />;\n          }\n        });\n      ',
    },
    {
      code: '\n        var First = createReactClass({\n          propTypes: {\n            a: PropTypes.any,\n            A: PropTypes.any,\n            z: PropTypes.string,\n            Z: PropTypes.string\n          },\n          getDefaultProps: function() {\n            return {\n              a: "a",\n              A: "A",\n              z: "z",\n              Z: "Z"\n            };\n          },\n          render: function() {\n            return <div />;\n          }\n        });\n      ',
      options: [
        {
          ignoreCase: true,
        },
      ],
    },
    {
      code: '\n        var First = createReactClass({\n          propTypes: {\n            a: PropTypes.any,\n            z: PropTypes.string\n          },\n          getDefaultProps: function() {\n            return {\n              a: "a",\n              z: "z"\n            };\n          },\n          render: function() {\n            return <div />;\n          }\n        });\n        var Second = createReactClass({\n          propTypes: {\n            AA: PropTypes.any,\n            ZZ: PropTypes.string\n          },\n          getDefaultProps: function() {\n            return {\n              AA: "AA",\n              ZZ: "ZZ"\n            };\n          },\n          render: function() {\n            return <div />;\n          }\n        });\n      ',
    },
    {
      code: '\n        class First extends React.Component {\n          render() {\n            return <div />;\n          }\n        }\n        First.propTypes = {\n          a: PropTypes.string,\n          z: PropTypes.string\n        };\n        First.propTypes.justforcheck = PropTypes.string;\n        First.defaultProps = {\n          a: a,\n          z: z\n        };\n        First.defaultProps.justforcheck = "justforcheck";\n      ',
    },
    {
      code: '\n        class First extends React.Component {\n          render() {\n            return <div />;\n          }\n        }\n        First.propTypes = {\n          a: PropTypes.any,\n          A: PropTypes.any,\n          z: PropTypes.string,\n          Z: PropTypes.string\n        };\n        First.defaultProps = {\n          a: "a",\n          A: "A",\n          z: "z",\n          Z: "Z"\n        };\n      ',
      options: [
        {
          ignoreCase: true,
        },
      ],
    },
    {
      code: '\n        class Component extends React.Component {\n          static propTypes = {\n            a: PropTypes.any,\n            b: PropTypes.any,\n            c: PropTypes.any\n          };\n          static defaultProps = {\n            a: "a",\n            b: "b",\n            c: "c"\n          };\n          render() {\n            return <div />;\n          }\n        }\n      ',
    },
    {
      code: '\n        class Hello extends React.Component {\n          render() {\n            return <div>Hello</div>;\n          }\n        }\n        Hello.propTypes = {\n          "aria-controls": PropTypes.string\n        };\n        Hello.defaultProps = {\n          "aria-controls": "aria-controls"\n        };\n      ',
      options: [
        {
          ignoreCase: true,
        },
      ],
    },
    {
      code: '\n        var Hello = createReactClass({\n          render: function() {\n            let { a, ...b } = obj;\n            let c = { ...d };\n            return <div />;\n          }\n        });\n      ',
    },
    {
      code: '\n        var First = createReactClass({\n          propTypes: {\n            barRequired: PropTypes.func.isRequired,\n            onBar: PropTypes.func,\n            z: PropTypes.any\n          },\n          getDefaultProps: function() {\n            return {\n              barRequired: "barRequired",\n              onBar: "onBar",\n              z: "z"\n            };\n          },\n          render: function() {\n            return <div />;\n          }\n        });\n      ',
    },
    {
      code: '\n        export default class ClassWithSpreadInPropTypes extends BaseClass {\n          static propTypes = {\n            b: PropTypes.string,\n            ...c.propTypes,\n            a: PropTypes.string\n          }\n          static defaultProps = {\n            b: "b",\n            ...c.defaultProps,\n            a: "a"\n          }\n        }\n      ',
    },
    {
      code: '\n        export default class ClassWithSpreadInPropTypes extends BaseClass {\n          static propTypes = {\n            a: PropTypes.string,\n            b: PropTypes.string,\n            c: PropTypes.string,\n            d: PropTypes.string,\n            e: PropTypes.string,\n            f: PropTypes.string\n          }\n          static defaultProps = {\n            a: "a",\n            b: "b",\n            ...c.defaultProps,\n            e: "e",\n            f: "f",\n            ...d.defaultProps\n          }\n        }\n      ',
    },
    {
      code: '\n        const defaults = {\n          b: "b"\n        };\n        const types = {\n          a: PropTypes.string,\n          b: PropTypes.string,\n          c: PropTypes.string\n        };\n        function StatelessComponentWithSpreadInPropTypes({ a, b, c }) {\n          return <div>{a}{b}{c}</div>;\n        }\n        StatelessComponentWithSpreadInPropTypes.propTypes = types;\n        StatelessComponentWithSpreadInPropTypes.defaultProps = {\n          c: "c",\n          ...defaults,\n          a: "a"\n        };\n      ',
    },
    {
      code: "\n        const propTypes = require('./externalPropTypes')\n        const defaultProps = require('./externalDefaultProps')\n        const TextFieldLabel = (props) => {\n          return <div />;\n        };\n        TextFieldLabel.propTypes = propTypes;\n        TextFieldLabel.defaultProps = defaultProps;\n      ",
    },
    {
      code: '\n        const First = (props) => <div />;\n        export const propTypes = {\n            a: PropTypes.any,\n            z: PropTypes.string,\n        };\n        export const defaultProps = {\n            a: "a",\n            z: "z",\n        };\n        First.propTypes = propTypes;\n        First.defaultProps = defaultProps;\n      ',
    },
    {
      code: '\n        const defaults = {\n          b: "b"\n        };\n        const First = (props) => <div />;\n        export const propTypes = {\n            a: PropTypes.string,\n            b: PropTypes.string,\n            z: PropTypes.string,\n        };\n        export const defaultProps = {\n            ...defaults,\n            a: "a",\n            z: "z",\n        };\n        First.propTypes = propTypes;\n        First.defaultProps = defaultProps;\n      ',
    },
    {
      code: '\n        class First extends React.Component {\n          render() {\n            return <div />;\n          }\n        }\n\n        First.defaultProps = {\n            a: PropTypes.any,\n            onBar: PropTypes.func,\n            onFoo: PropTypes.func,\n            z: PropTypes.string,\n        };\n      ',
    },
  ],
  invalid: [
    {
      code: '\n        class Component extends React.Component {\n          static propTypes = {\n            a: PropTypes.any,\n            b: PropTypes.any,\n            c: PropTypes.any\n          };\n          static defaultProps = {\n            a: "a",\n            c: "c",\n            b: "b"\n          };\n          render() {\n            return <div />;\n          }\n        }\n      ',
      errors: [
        {
          messageId: 'propsNotSorted',
          message:
            'Default prop types declarations should be sorted alphabetically',
          line: 11,
          column: 13,
          endLine: 11,
          endColumn: 19,
        },
      ],
    },
    {
      code: '\n        class Component extends React.Component {\n          static propTypes = {\n            a: PropTypes.any,\n            b: PropTypes.any,\n            c: PropTypes.any\n          };\n          static defaultProps = {\n            c: "c",\n            b: "b",\n            a: "a"\n          };\n          render() {\n            return <div />;\n          }\n        }\n      ',
      errors: [
        {
          messageId: 'propsNotSorted',
          message:
            'Default prop types declarations should be sorted alphabetically',
          line: 10,
          column: 13,
          endLine: 10,
          endColumn: 19,
        },
        {
          messageId: 'propsNotSorted',
          message:
            'Default prop types declarations should be sorted alphabetically',
          line: 11,
          column: 13,
          endLine: 11,
          endColumn: 19,
        },
      ],
    },
    {
      code: '\n        class Component extends React.Component {\n          static propTypes = {\n            a: PropTypes.any,\n            b: PropTypes.any\n          };\n          static defaultProps = {\n            Z: "Z",\n            a: "a",\n          };\n          render() {\n            return <div />;\n          }\n        }\n      ',
      options: [
        {
          ignoreCase: true,
        },
      ],
      errors: [
        {
          messageId: 'propsNotSorted',
          message:
            'Default prop types declarations should be sorted alphabetically',
          line: 9,
          column: 13,
          endLine: 9,
          endColumn: 19,
        },
      ],
    },
    {
      code: '\n        class Component extends React.Component {\n          static propTypes = {\n            a: PropTypes.any,\n            z: PropTypes.any\n          };\n          static defaultProps = {\n            a: "a",\n            Z: "Z",\n          };\n          render() {\n            return <div />;\n          }\n        }\n      ',
      errors: [
        {
          messageId: 'propsNotSorted',
          message:
            'Default prop types declarations should be sorted alphabetically',
          line: 9,
          column: 13,
          endLine: 9,
          endColumn: 19,
        },
      ],
    },
    {
      code: '\n        class Hello extends React.Component {\n          render() {\n            return <div>Hello</div>;\n          }\n        }\n        Hello.propTypes = {\n          "a": PropTypes.string,\n          "b": PropTypes.string\n        };\n        Hello.defaultProps = {\n          "b": "b",\n          "a": "a"\n        };\n      ',
      errors: [
        {
          messageId: 'propsNotSorted',
          message:
            'Default prop types declarations should be sorted alphabetically',
          line: 13,
          column: 11,
          endLine: 13,
          endColumn: 19,
        },
      ],
    },
    {
      code: '\n        class Hello extends React.Component {\n          render() {\n            return <div>Hello</div>;\n          }\n        }\n        Hello.propTypes = {\n          "a": PropTypes.string,\n          "b": PropTypes.string,\n          "c": PropTypes.string\n        };\n        Hello.defaultProps = {\n          "c": "c",\n          "b": "b",\n          "a": "a"\n        };\n      ',
      errors: [
        {
          messageId: 'propsNotSorted',
          message:
            'Default prop types declarations should be sorted alphabetically',
          line: 14,
          column: 11,
          endLine: 14,
          endColumn: 19,
        },
        {
          messageId: 'propsNotSorted',
          message:
            'Default prop types declarations should be sorted alphabetically',
          line: 15,
          column: 11,
          endLine: 15,
          endColumn: 19,
        },
      ],
    },
    {
      code: '\n        class Hello extends React.Component {\n          render() {\n            return <div>Hello</div>;\n          }\n        }\n        Hello.propTypes = {\n          "a": PropTypes.string,\n          "B": PropTypes.string,\n        };\n        Hello.defaultProps = {\n          "a": "a",\n          "B": "B",\n        };\n      ',
      errors: [
        {
          messageId: 'propsNotSorted',
          message:
            'Default prop types declarations should be sorted alphabetically',
          line: 13,
          column: 11,
          endLine: 13,
          endColumn: 19,
        },
      ],
    },
    {
      code: '\n        class Hello extends React.Component {\n          render() {\n            return <div>Hello</div>;\n          }\n        }\n        Hello.propTypes = {\n          "a": PropTypes.string,\n          "B": PropTypes.string,\n        };\n        Hello.defaultProps = {\n          "B": "B",\n          "a": "a",\n        };\n      ',
      options: [
        {
          ignoreCase: true,
        },
      ],
      errors: [
        {
          messageId: 'propsNotSorted',
          message:
            'Default prop types declarations should be sorted alphabetically',
          line: 13,
          column: 11,
          endLine: 13,
          endColumn: 19,
        },
      ],
    },
    {
      code: '\n        const First = (props) => <div />;\n        const propTypes = {\n          z: PropTypes.string,\n          a: PropTypes.any,\n        };\n        const defaultProps = {\n          z: "z",\n          a: "a",\n        };\n        First.propTypes = propTypes;\n        First.defaultProps = defaultProps;\n      ',
      errors: [
        {
          messageId: 'propsNotSorted',
          message:
            'Default prop types declarations should be sorted alphabetically',
          line: 9,
          column: 11,
          endLine: 9,
          endColumn: 17,
        },
      ],
    },
    {
      code: '\n        export default class ClassWithSpreadInPropTypes extends BaseClass {\n          static propTypes = {\n            b: PropTypes.string,\n            ...c.propTypes,\n            a: PropTypes.string\n          }\n          static defaultProps = {\n            b: "b",\n            a: "a",\n            ...c.defaultProps\n          }\n        }\n      ',
      errors: [
        {
          messageId: 'propsNotSorted',
          message:
            'Default prop types declarations should be sorted alphabetically',
          line: 10,
          column: 13,
          endLine: 10,
          endColumn: 19,
        },
      ],
    },
    {
      code: '\n        export default class ClassWithSpreadInPropTypes extends BaseClass {\n          static propTypes = {\n            a: PropTypes.string,\n            b: PropTypes.string,\n            c: PropTypes.string,\n            d: PropTypes.string,\n            e: PropTypes.string,\n            f: PropTypes.string\n          }\n          static defaultProps = {\n            b: "b",\n            a: "a",\n            ...c.defaultProps,\n            f: "f",\n            e: "e",\n            ...d.defaultProps\n          }\n        }\n      ',
      errors: [
        {
          messageId: 'propsNotSorted',
          message:
            'Default prop types declarations should be sorted alphabetically',
          line: 13,
          column: 13,
          endLine: 13,
          endColumn: 19,
        },
        {
          messageId: 'propsNotSorted',
          message:
            'Default prop types declarations should be sorted alphabetically',
          line: 16,
          column: 13,
          endLine: 16,
          endColumn: 19,
        },
      ],
    },
    {
      code: '\n        const defaults = {\n          b: "b"\n        };\n        const types = {\n          a: PropTypes.string,\n          b: PropTypes.string,\n          c: PropTypes.string\n        };\n        function StatelessComponentWithSpreadInPropTypes({ a, b, c }) {\n          return <div>{a}{b}{c}</div>;\n        }\n        StatelessComponentWithSpreadInPropTypes.propTypes = types;\n        StatelessComponentWithSpreadInPropTypes.defaultProps = {\n          c: "c",\n          a: "a",\n          ...defaults,\n        };\n      ',
      errors: [
        {
          messageId: 'propsNotSorted',
          message:
            'Default prop types declarations should be sorted alphabetically',
          line: 16,
          column: 11,
          endLine: 16,
          endColumn: 17,
        },
      ],
    },
    {
      code: '\n        class First extends React.Component {\n          render() {\n            return <div />;\n          }\n        }\n\n        First.defaultProps = {\n            a: PropTypes.any,\n            z: PropTypes.string,\n            onFoo: PropTypes.func,\n            onBar: PropTypes.func,\n        };\n      ',
      errors: [
        {
          messageId: 'propsNotSorted',
          message:
            'Default prop types declarations should be sorted alphabetically',
          line: 11,
          column: 13,
          endLine: 11,
          endColumn: 34,
        },
        {
          messageId: 'propsNotSorted',
          message:
            'Default prop types declarations should be sorted alphabetically',
          line: 12,
          column: 13,
          endLine: 12,
          endColumn: 34,
        },
      ],
    },
  ],
});

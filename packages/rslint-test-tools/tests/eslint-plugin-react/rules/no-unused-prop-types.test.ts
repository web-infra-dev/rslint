import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, upstreamModuleOptions } from '../config';

function expectUnusedPropTypeError1(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'unusedPropType',
    message: "'name' PropType is defined but prop is never used",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectUnusedPropTypeError2(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'unusedPropType',
    message: "'unused' PropType is defined but prop is never used",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectUnusedPropTypeError3(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'unusedPropType',
    message: "'zap' PropType is defined but prop is never used",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectUnusedPropTypeError4(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'unusedPropType',
    message: "'something' PropType is defined but prop is never used",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectUnusedPropTypeError5(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'unusedPropType',
    message: "'bar' PropType is defined but prop is never used",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectUnusedPropTypeError6(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'unusedPropType',
    message: "'baz' PropType is defined but prop is never used",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectUnusedPropTypeError7(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'unusedPropType',
    message: "'foo' PropType is defined but prop is never used",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectUnusedPropTypeError8(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'unusedPropType',
    message: "'prop1' PropType is defined but prop is never used",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectUnusedPropTypeError9(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'unusedPropType',
    message: "'prop2' PropType is defined but prop is never used",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectUnusedPropTypeError10(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'unusedPropType',
    message: "'prop2.*' PropType is defined but prop is never used",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectUnusedPropTypeError11(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'unusedPropType',
    message: "'lastname' PropType is defined but prop is never used",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectUnusedPropTypeError12(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'unusedPropType',
    message: "'z' PropType is defined but prop is never used",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectUnusedPropTypeError13(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'unusedPropType',
    message: "'y' PropType is defined but prop is never used",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectUnusedPropTypeError14(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'unusedPropType',
    message: "'books' PropType is defined but prop is never used",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectUnusedPropTypeError15(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'unusedPropType',
    message: "'username' PropType is defined but prop is never used",
    line,
    column,
    endLine,
    endColumn,
  };
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('no-unused-prop-types', {
  valid: [
    {
      code: `class Hello extends React.Component { render() { return <div>{this.props.name}</div>; } } Hello.propTypes = { name: PropTypes.string };`,
    },
    {
      code: `function Hello(props) { return <div>{props.name}</div>; } Hello.propTypes = { name: PropTypes.string };`,
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 0 (7.37.5)',
      code: `
        var Hello = createReactClass({
          propTypes: {
            name: PropTypes.string.isRequired
          },
          render: function() {
            return <div>Hello {this.props.name}</div>;
          }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 1 (7.37.5)',
      code: `
        var Hello = createReactClass({
          propTypes: {
            name: PropTypes.object.isRequired
          },
          render: function() {
            return <div>Hello {this.props.name.firstname}</div>;
          }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 2 (7.37.5)',
      code: `
        var Hello = createReactClass({
          render: function() {
            return <div>Hello World</div>;
          }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 3 (7.37.5)',
      code: `
        var Hello = createReactClass({
          render: function() {
            var props = this.props;
            return <div>Hello World</div>;
          }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 4 (7.37.5)',
      code: `
        var Hello = createReactClass({
          render: function() {
            var propName = "foo";
            return <div>Hello World {this.props[propName]}</div>;
          }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 5 (7.37.5)',
      code: `
        var Hello = createReactClass({
          propTypes: externalPropTypes,
          render: function() {
            return <div>Hello {this.props.name}</div>;
          }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 6 (7.37.5)',
      code: `
        var Hello = createReactClass({
          propTypes: externalPropTypes.mySharedPropTypes,
          render: function() {
            return <div>Hello {this.props.name}</div>;
          }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 7 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            return <div>Hello World</div>;
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
            return <div>Hello {this.props.firstname} {this.props.lastname}</div>;
          }
        }
        Hello.propTypes = {
          firstname: PropTypes.string
        };
        Hello.propTypes.lastname = PropTypes.string;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 9 (7.37.5)',
      code: `
        var Hello = createReactClass({
          propTypes: {
            name: PropTypes.object.isRequired
          },
          render: function() {
            var user = {
              name: this.props.name
            };
            return <div>Hello {user.name}</div>;
          }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 10 (7.37.5)',
      code: `
        class Hello {
          render() {
            return 'Hello' + this.props.name;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 11 (7.37.5)',
      code: `
        class Hello {
          method;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 12 (7.37.5)',
      code: `
        class Hello extends React.Component {
          static get propTypes() {
            return {
              name: PropTypes.string
            };
          }
          render() {
            return <div>Hello {this.props.name}</div>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 13 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            var { firstname, ...props } = this.props;
            var { category, icon } = props;
            return <div>Hello {firstname}</div>;
          }
        }
        Hello.propTypes = {
          firstname: PropTypes.string,
          category: PropTypes.string,
          icon: PropTypes.bool
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 14 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            var {firstname, lastname} = this.state, something = this.props;
            return <div>Hello {firstname}</div>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 15 (7.37.5)',
      code: `
        class Hello extends React.Component {
          static propTypes = {
            name: PropTypes.string
          };
          render() {
            return <div>Hello {this.props.name}</div>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 16 (7.37.5)',
      code: `
        class Hello extends React.Component {
          static propTypes = {
            name: PropTypes.string
          };
          render() {
            return <div>Hello {this.props?.name}</div>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 17 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            return <div>Hello {this.props.firstname}</div>;
          }
        }
        Hello.propTypes = {
          'firstname': PropTypes.string
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 18 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            if (Object.prototype.hasOwnProperty.call(this.props, 'firstname')) {
              return <div>Hello {this.props.firstname}</div>;
            }
            return <div>Hello</div>;
          }
        }
        Hello.propTypes = {
          'firstname': PropTypes.string
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 19 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            this.props.a.b
            return <div>Hello</div>;
          }
        }
        Hello.propTypes = {};
        Hello.propTypes.a = PropTypes.shape({
          b: PropTypes.string
        });
      `,
      options: [{ skipShapeProps: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 20 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            this.props.a.b.c;
            return <div>Hello</div>;
          }
        }
        Hello.propTypes = {
          a: PropTypes.shape({
            b: PropTypes.shape({
            })
          })
        };
        Hello.propTypes.a.b.c = PropTypes.number;
      `,
      options: [{ skipShapeProps: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 21 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            this.props.a.b.c;
            this.props.a.__.d.length;
            this.props.a.anything.e[2];
            return <div>Hello</div>;
          }
        }
        Hello.propTypes = {
          a: PropTypes.objectOf(
            PropTypes.shape({
              c: PropTypes.number,
              d: PropTypes.string,
              e: PropTypes.array
            })
          )
        };
      `,
      options: [{ skipShapeProps: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 22 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            var i = 3;
            this.props.a[2].c;
            this.props.a[i].d.length;
            this.props.a[i + 2].e[2];
            this.props.a.length;
            return <div>Hello</div>;
          }
        }
        Hello.propTypes = {
          a: PropTypes.arrayOf(
            PropTypes.shape({
              c: PropTypes.number,
              d: PropTypes.string,
              e: PropTypes.array
            })
          )
        };
      `,
      options: [{ skipShapeProps: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 23 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            this.props.a.length;
            return <div>Hello</div>;
          }
        }
        Hello.propTypes = {
          a: PropTypes.oneOfType([
            PropTypes.array,
            PropTypes.string
          ])
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 24 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            this.props.a.c;
            this.props.a[2] === true;
            this.props.a.e[2];
            this.props.a.length;
            return <div>Hello</div>;
          }
        }
        Hello.propTypes = {
          a: PropTypes.oneOfType([
            PropTypes.shape({
              c: PropTypes.number,
              e: PropTypes.array
            }).isRequired,
            PropTypes.arrayOf(
              PropTypes.bool
            )
          ])
        };
      `,
      options: [{ skipShapeProps: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 25 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            this.props.a.render;
            this.props.a.c;
            return <div>Hello</div>;
          }
        }
        Hello.propTypes = {
          a: PropTypes.instanceOf(Hello)
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 26 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            this.props.arr;
            this.props.arr[3];
            this.props.arr.length;
            this.props.arr.push(3);
            this.props.bo;
            this.props.bo.toString();
            this.props.fu;
            this.props.fu.bind(this);
            this.props.numb;
            this.props.numb.toFixed();
            this.props.stri;
            this.props.stri.length();
            return <div>Hello</div>;
          }
        }
        Hello.propTypes = {
          arr: PropTypes.array,
          bo: PropTypes.bool.isRequired,
          fu: PropTypes.func,
          numb: PropTypes.number,
          stri: PropTypes.string
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 27 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            var {
              propX,
              "aria-controls": ariaControls,
              ...props } = this.props;
            return <div>Hello</div>;
          }
        }
        Hello.propTypes = {
          "propX": PropTypes.string,
          "aria-controls": PropTypes.string
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 28 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            this.props["some.value"];
            return <div>Hello</div>;
          }
        }
        Hello.propTypes = {
          "some.value": PropTypes.string
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 29 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            this.props["arr"][1];
            return <div>Hello</div>;
          }
        }
        Hello.propTypes = {
          "arr": PropTypes.array
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 30 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            this.props["arr"][1]["some.value"];
            return <div>Hello</div>;
          }
        }
        Hello.propTypes = {
          "arr": PropTypes.arrayOf(
            PropTypes.shape({"some.value": PropTypes.string})
          )
        };
      `,
      options: [{ skipShapeProps: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 31 (7.37.5)',
      code: `
        var TestComp1 = createReactClass({
          propTypes: {
            size: PropTypes.string
          },
          render: function() {
            var foo = {
              baz: 'bar'
            };
            var icons = foo[this.props.size].salut;
            return <div>{icons}</div>;
          }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 32 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            const {firstname, lastname} = this.props.name;
            return <div>{firstname} {lastname}</div>;
          }
        }
        Hello.propTypes = {
          name: PropTypes.shape({
            firstname: PropTypes.string,
            lastname: PropTypes.string
          })
        };
      `,
      options: [{ skipShapeProps: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 33 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            let {firstname} = this;
            return <div>{firstname}</div>;
          }
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 34 (7.37.5)',
      code: `
        function Foo({ a }) {
          return <>{ a.b }</>
        }
        Foo.propTypes = {
          a: PropTypes.shape({
            b: PropType.string,
          })
        }
      `,
      options: [{ skipShapeProps: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 35 (7.37.5)',
      code: `
        function Foo({ a }) {
          const { b } = a
          return <>{ b.c }</>
        }
        Foo.propTypes = {
          a: PropTypes.shape({
            b: PropType.shape({
              c: PropTypes.string,
            }),
          })
        }
      `,
      options: [{ skipShapeProps: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 36 (7.37.5)',
      code: `
        class Thing extends React.Component {
          static propTypes = {
            i18n: PropTypes.shape({
              gettext: PropTypes.func,
            }),
          }

          render() {
            const { i18n } = this.props;
            return (
              <p>{i18n.gettext('Some Text')}</p>
            );
          }
        }
      `,
      options: [{ skipShapeProps: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 37 (7.37.5)',
      code: `
        class Thing extends React.Component {
          static propTypes = {
            a: PropTypes.shape({
              b: PropTypes.string,
            }),
          }

          render() {
            const { a } = this.props;
            return (
              <p>{ a.b }</p>
            );
          }
        }
      `,
      options: [{ skipShapeProps: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 38 (7.37.5)',
      code: `
        var Hello = createReactClass({
          propTypes: {
            router: PropTypes.func
          },
          render: function() {
            var nextPath = this.props.router.getCurrentQuery().nextPath;
            return <div>{nextPath}</div>;
          }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 39 (7.37.5)',
      code: `
        var Hello = createReactClass({
          propTypes: {
            firstname: CustomValidator.string
          },
          render: function() {
            return <div>{this.props.firstname}</div>;
          }
        });
      `,
      options: [{ customValidators: ['CustomValidator'] }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 40 (7.37.5)',
      code: `
        var Hello = createReactClass({
          propTypes: {
            outer: CustomValidator.shape({
              inner: CustomValidator.map
            })
          },
          render: function() {
            return <div>{this.props.outer.inner}</div>;
          }
        });
      `,
      options: [
        { customValidators: ['CustomValidator'], skipShapeProps: false },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 41 (7.37.5)',
      code: `
        var Hello = createReactClass({
          propTypes: {
            outer: PropTypes.shape({
              inner: CustomValidator.string
            })
          },
          render: function() {
            return <div>{this.props.outer.inner}</div>;
          }
        });
      `,
      options: [
        { customValidators: ['CustomValidator'], skipShapeProps: false },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 42 (7.37.5)',
      code: `
        var Hello = createReactClass({
          propTypes: {
            outer: CustomValidator.shape({
              inner: PropTypes.string
            })
          },
          render: function() {
            return <div>{this.props.outer.inner}</div>;
          }
        });
      `,
      options: [
        { customValidators: ['CustomValidator'], skipShapeProps: false },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 43 (7.37.5)',
      code: `
        var Hello = createReactClass({
          propTypes: {
            name: PropTypes.string
          },
          render: function() {
            return <div>{this.props.name.get("test")}</div>;
          }
        });
      `,
      options: [{ customValidators: ['CustomValidator'] }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 44 (7.37.5)',
      code: `
        const SomeComponent = createReactClass({
          propTypes: SomeOtherComponent.propTypes
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 45 (7.37.5)',
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
      name: 'Upstream valid 46 (7.37.5)',
      code: `
        class Hello extends React.Component {
          static get propTypes() {}
          render() {
            return <div>Hello World</div>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 47 (7.37.5)',
      code: `
        class Hello extends React.Component {
          static get propTypes() {}
          render() {
            var users = this.props.users.find(user => user.name === 'John');
            return <div>Hello you {users.length}</div>;
          }
        }
        Hello.propTypes = {
          users: PropTypes.arrayOf(PropTypes.object)
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 48 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            const {} = this.props;
            return <div>Hello</div>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 49 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            var foo = 'fullname';
            var { [foo]: firstname } = this.props;
            return <div>Hello {firstname}</div>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 50 (7.37.5)',
      code: `
        class Hello extends React.Component {
          constructor(props, context) {
            super(props, context)
            this.state = { status: props.source.uri }
          }
          static propTypes = {
            source: PropTypes.object
          };
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 51 (7.37.5)',
      code: `
        class Hello extends React.Component {
          constructor(props, context) {
            super(props, context)
            this.state = { status: props?.source?.uri }
          }
          static propTypes = {
            source: PropTypes.object
          };
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 52 (7.37.5)',
      code: `
        class Hello extends React.Component {
          constructor(props, context) {
            super(props, context)
            this.state = { status: this.props.source.uri }
          }
          static propTypes = {
            source: PropTypes.object
          };
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 53 (7.37.5)',
      code: `
        HelloJohn.prototype.render = function() {
          return React.createElement(Hello, {
            name: this.props.firstname
          });
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 54 (7.37.5)',
      code: `
        HelloJohn.prototype.render = function() {
          return React.createElement(Hello, {
            name: this.props?.firstname
          });
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 55 (7.37.5)',
      code: `
        function HelloComponent() {
          class Hello extends React.Component {
            render() {
              return <div>Hello {this.props.name}</div>;
            }
          }
          Hello.propTypes = { name: PropTypes.string };
          return Hello;
        }
        module.exports = HelloComponent();
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 56 (7.37.5)',
      code: `
        function HelloComponent() {
          class Hello extends React.Component {
            render() {
              return <div>Hello {this.props?.name}</div>;
            }
          }
          Hello.propTypes = { name: PropTypes.string };
          return Hello;
        }
        module.exports = HelloComponent();
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 57 (7.37.5)',
      code: `
        function HelloComponent() {
          var Hello = createReactClass({
            propTypes: { name: PropTypes.string },
            render: function() {
              return <div>Hello {this.props.name}</div>;
            }
          });
          return Hello;
        }
        module.exports = HelloComponent();
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 58 (7.37.5)',
      code: `
        function HelloComponent() {
          var Hello = createReactClass({
            propTypes: { name: PropTypes.string },
            render: function() {
              return <div>Hello {this.props?.name}</div>;
            }
          });
          return Hello;
        }
        module.exports = HelloComponent();
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 59 (7.37.5)',
      code: `
        class DynamicHello extends Component {
          render() {
            const {firstname} = this.props;
            class Hello extends Component {
              render() {
                const {name} = this.props;
                return <div>Hello {name}</div>;
              }
            }
            Hello.propTypes = {
              name: PropTypes.string
            };
            Hello = connectReduxForm({name: firstname})(Hello);
            return <Hello />;
          }
        }
        DynamicHello.propTypes = {
          firstname: PropTypes.string,
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 60 (7.37.5)',
      code: `
        const Hello = (props) => {
          let team = props.names.map((name) => {
              return <li>{name}, {props.company}</li>;
            });
          return <ul>{team}</ul>;
        };
        Hello.propTypes = {
          names: PropTypes.array,
          company: PropTypes.string
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 61 (7.37.5)',
      code: `
        const Hello = (props) => {
          let team = props?.names.map((name) => {
              return <li>{name}, {props?.company}</li>;
            });
          return <ul>{team}</ul>;
        };
        Hello.propTypes = {
          names: PropTypes.array,
          company: PropTypes.string
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 62 (7.37.5)',
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
      name: 'Upstream valid 63 (7.37.5)',
      code: `
        export class Hello extends Component {
          render() {
            const props = this.props;
            return <div>Hello {props.name.firstname} {props['name'].lastname}</div>
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 64 (7.37.5)',
      code: `
        export default function FooBar(props) {
          const bar = props.bar;
          return (<div bar={bar}><div {...props}/></div>);
        }
        if (process.env.NODE_ENV !== 'production') {
          FooBar.propTypes = {
            bar: PropTypes.string
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 65 (7.37.5)',
      code: `
        var Hello = createReactClass({
          render: function() {
            var {...other} = this.props;
            return (
              <div {...other} />
            );
          }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 66 (7.37.5)',
      code: `
        function notAComponent({ something }) {
          return something + 1;
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 67 (7.37.5)',
      code: `
        const notAComponent = function({ something }) {
          return something + 1;
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 68 (7.37.5)',
      code: `
        const notAComponent = ({ something }) => {
          return something + 1;
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 69 (7.37.5)',
      code: `
        const statelessComponent = (props) => {
          let newProps = props;
          return <span>{newProps.someProp}</span>;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 70 (7.37.5)',
      code: `
        class Hello extends React.Component {
          props: {
            name: string;
          };
          render () {
            return <div>Hello {this.props.name}</div>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 71 (7.37.5)',
      code: `
        class Hello extends React.Component {
          props: {
            name: Object;
          };
          render () {
            return <div>Hello {this.props.name.firstname}</div>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 72 (7.37.5)',
      code: `
        type Props = {name: Object;};
        class Hello extends React.Component {
          props: Props;
          render () {
            return <div>Hello {this.props.name.firstname}</div>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 73 (7.37.5)',
      code: `
        type PropsA = { a: string }
        type PropsB = { b: string }
        type Props = PropsA & PropsB;

        class MyComponent extends React.Component {
          props: Props;

          render() {
            return <div>{this.props.a} - {this.props.b}</div>
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 74 (7.37.5)',
      code: `
        type PropsA = { foo: string };
        type PropsB = { bar: string };
        type PropsC = { zap: string };
        type Props = PropsA & PropsB;

        class Bar extends React.Component {
          props: Props & PropsC;

          render() {
            return <div>{this.props.foo} - {this.props.bar} - {this.props.zap}</div>
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 75 (7.37.5)',
      code: `
        import type { PropsA } from "./myPropsA";
        type PropsB = { bar: string };
        type PropsC = { zap: string };
        type Props = PropsA & PropsB;

        class Bar extends React.Component {
          props: Props & PropsC;

          render() {
            return <div>{this.props.foo} - {this.props.bar} - {this.props.zap}</div>
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 76 (7.37.5)',
      code: `
        type PropsB = { foo: string };
        type PropsC = { bar: string };
        type Props = PropsB & {
          zap: string
        };

        class Bar extends React.Component {
          props: Props & PropsC;

          render() {
            return <div>{this.props.foo} - {this.props.bar} - {this.props.zap}</div>
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 77 (7.37.5)',
      code: `
        type PropsB = { foo: string };
        type PropsC = { bar: string };
        type Props = {
          zap: string
        } & PropsB;

        class Bar extends React.Component {
          props: Props & PropsC;

          render() {
            return <div>{this.props.foo} - {this.props.bar} - {this.props.zap}</div>
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 78 (7.37.5)',
      code: `
        import type Props from "fake";
        class Hello extends React.Component {
          props: Props;
          render () {
            return <div>Hello {this.props.name.firstname}</div>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 79 (7.37.5)',
      code: `
        class Hello extends React.Component {
          props: {
            name: {
              firstname: string;
            }
          };
          render () {
            return <div>Hello {this.props.name.firstname}</div>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 80 (7.37.5)',
      code: `
        type Props = {name: {firstname: string;};};
        class Hello extends React.Component {
          props: Props;
          render () {
            return <div>Hello {this.props.name.firstname}</div>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 81 (7.37.5)',
      code: `
        type Person = {name: {firstname: string;}};
        class Hello extends React.Component {
          props: {people: Person[];};
          render () {
            var names = [];
            for (var i = 0; i < this.props.people.length; i++) {
              names.push(this.props.people[i].name.firstname);
            }
            return <div>Hello {names.join(', ')}</div>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 82 (7.37.5)',
      code: `
        type Person = {name: {firstname: string;}};
        type Props = {people: Person[];};
        class Hello extends React.Component {
          props: Props;
          render () {
            var names = [];
            for (var i = 0; i < this.props.people.length; i++) {
              names.push(this.props.people[i].name.firstname);
            }
            return <div>Hello {names.join(', ')}</div>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 83 (7.37.5)',
      code: `
        type Person = {name: {firstname: string;}};
        type Props = {people: Person[]|Person;};
        class Hello extends React.Component {
          props: Props;
          render () {
            var names = [];
            if (Array.isArray(this.props.people)) {
              for (var i = 0; i < this.props.people.length; i++) {
                names.push(this.props.people[i].name.firstname);
              }
            } else {
              names.push(this.props.people.name.firstname);
            }
            return <div>Hello {names.join(', ')}</div>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 84 (7.37.5)',
      code: `
        type Props = {ok: string | boolean;};
        class Hello extends React.Component {
          props: Props;
          render () {
            return <div>Hello {this.props.ok}</div>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 85 (7.37.5)',
      code: `
        type Props = {result: {ok: string | boolean;}|{ok: number | Array}};
        class Hello extends React.Component {
          props: Props;
          render () {
            return <div>Hello {this.props.result.ok}</div>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 86 (7.37.5)',
      code: `
        type Props = {result?: {ok?: ?string | boolean;}|{ok?: ?number | Array}};
        class Hello extends React.Component {
          props: Props;
          render () {
            return <div>Hello {this.props.result.ok}</div>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 87 (7.37.5)',
      code: `
        type Props = {notTarget: string};
        class Hello extends React.Component {
          props: Props;
          onEvent({ target }: { target: Object }) {};
          render () {
            return <div>Hello {this.props.notTarget}</div>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 88 (7.37.5)',
      code: `
        type Props = {notTarget: string};
        class Hello extends React.Component {
          props: Props;
          onEvent(infos: { target: Object }) {};
          render () {
            return <div>Hello {this.props.notTarget}</div>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 89 (7.37.5)',
      code: `
        class Hello extends React.Component {
          props = {a: 123};
          render () {
            return <div>Hello</div>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 90 (7.37.5)',
      code: `
        class Hello extends React.Component {
            render() {
                return  <div>Hello {this.props.firstName} {this.props.lastName}</div>;
            }
        };
        const otherPropTypes = {
            lastName: PropTypes.string
        };
        Hello.propTypes = {
            ...otherPropTypes,
            firstName: PropTypes.string
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 91 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render () {
            return ["string"].map(({length}) => <div>{length}</div>);
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 92 (7.37.5)',
      code: `
        type Props = {
          firstname: string;
          lastname: string;
        };
        function Hello(props: Props): React.Element {
          const {firstname, lastname} = props;
          return <div>Hello {firstname} {lastname}</div>
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 93 (7.37.5)',
      code: `
        type Props = {
          firstname: string;
          lastname: string;
        };
        const Hello = function(props: Props): React.Element {
          const {firstname, lastname} = props;
          return <div>Hello {firstname} {lastname}</div>
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 94 (7.37.5)',
      code: `
        type Props = {
          firstname: string;
          lastname: string;
        };
        const Hello = (props: Props): React.Element => {
          const {firstname, lastname} = props;
          return <div>Hello {firstname} {lastname}</div>
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 95 (7.37.5)',
      code: `
        type PropsUnionA = {
          a: string,
          b?: void,
        };
        type PropsUnionB = {
          a?: void,
          b: string,
        };
        type Props = {
          name: string,
        } & (PropsUnionA | PropsUnionB);
        class Hello extends React.Component {
          props: Props;
          render() {
            const {name} = this.props;
            return name;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 96 (7.37.5)',
      code: `
        import type { FieldProps } from "redux-form"

        type Props = {
        label: string,
          type: string,
          options: Array<SelectOption>
        } & FieldProps
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 97 (7.37.5)',
      code: `
        Card.propTypes = {
          title: PropTypes.string.isRequired,
          children: PropTypes.element.isRequired,
          footer: PropTypes.node
        }
        function Card ({ title, children, footer }) {
          return (
            <div/>
          )
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 98 (7.37.5)',
      code: `
        function JobList(props) {
          props
          .jobs
          .forEach(() => {});
          return <div></div>;
        }
        JobList.propTypes = {
          jobs: PropTypes.array
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 99 (7.37.5)',
      code: `
        type Props = {
          firstname: ?string,
        };
        function Hello({firstname}: Props): React$Element {
          return <div>Hello {firstname}</div>;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 100 (7.37.5)',
      code: `
        function Greetings() {
          return <div>{({name}) => <Hello name={name} />}</div>
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 101 (7.37.5)',
      code: `
        function Greetings() {
          return <div>{function({name}) { return <Hello name={name} />; }}</div>
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 102 (7.37.5)',
      code: `
        export default (ComposedComponent) => class Something extends SomeOtherComponent {
          someMethod = ({width}) => {}
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 103 (7.37.5)',
      code: `
        @asyncConnect([{
          promise: ({dispatch}) => {}
        }])
        class Something extends Component {}
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 104 (7.37.5)',
      code: `
        class Hello extends Component {
          static propTypes = {
            params: PropTypes.shape({
              id: PropTypes.string
            })
           }
          render () {
            const {params} = this.props
            const id = (params || {}).id;
            return <span>{id}</span>
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 105 (7.37.5)',
      code: `
        class Hello extends Component {
          static propTypes = {
            something: PropTypes.bool
          }
          componentWillReceiveProps (nextProps) {
            const {something} = nextProps;
            doSomething(something);
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 106 (7.37.5)',
      code: `
        class Hello extends Component {
          componentWillReceiveProps (nextProps) {
            const {something} = nextProps;
            doSomething(something);
          }
        }
        Hello.propTypes = {
          something: PropTypes.bool,
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 107 (7.37.5)',
      code: `
        var Hello = createReactClass({
          propTypes: {
            something: PropTypes.bool,
          },
          componentWillReceiveProps (nextProps) {
            const {something} = nextProps;
            doSomething(something);
          }
        })
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 108 (7.37.5)',
      code: `
        class Hello extends Component {
          static propTypes = {
            something: PropTypes.bool
          }
          componentWillReceiveProps ({something}) {
            doSomething(something);
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 109 (7.37.5)',
      code: `
        class Hello extends Component {
          componentWillReceiveProps ({something}) {
            doSomething(something);
          }
        }
        Hello.propTypes = {
          something: PropTypes.bool,
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 110 (7.37.5)',
      code: `
        var Hello = createReactClass({
          propTypes: {
            something: PropTypes.bool,
          },
          componentWillReceiveProps ({something}) {
            doSomething(something);
          }
        })
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 111 (7.37.5)',
      code: `
        class Hello extends Component {
          static propTypes = {
            something: PropTypes.bool
          }
          constructor (props) {
            super(props);
            const {something} = props;
            doSomething(something);
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 112 (7.37.5)',
      code: `
        class Hello extends Component {
          constructor (props) {
            super(props);
            const {something} = props;
            doSomething(something);
          }
        }
        Hello.propTypes = {
          something: PropTypes.bool,
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 113 (7.37.5)',
      code: `
        class Hello extends Component {
          static propTypes = {
            something: PropTypes.bool
          }
          constructor ({something}) {
            super({something});
            doSomething(something);
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 114 (7.37.5)',
      code: `
        class Hello extends Component {
          constructor ({something}) {
            super({something});
            doSomething(something);
          }
        }
        Hello.propTypes = {
          something: PropTypes.bool,
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 115 (7.37.5)',
      code: `
        // Destructured props in the \`shouldComponentUpdate\` method shouldn’t throw errors
        class Hello extends Component {
          static propTypes = {
            something: PropTypes.bool
          }
          shouldComponentUpdate(nextProps, nextState) {
            const {something} = nextProps;
            return something;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 116 (7.37.5)',
      code: `
        // Destructured props in the \`shouldComponentUpdate\` method shouldn’t throw errors
        class Hello extends Component {
          shouldComponentUpdate(nextProps, nextState) {
            const {something} = nextProps;
            return something;
          }
        }
        Hello.propTypes = {
          something: PropTypes.bool,
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 117 (7.37.5)',
      code: `
        // Destructured props in \`shouldComponentUpdate\` shouldn’t throw errors when used with createReactClass
        var Hello = createReactClass({
          propTypes: {
            something: PropTypes.bool,
          },
          shouldComponentUpdate(nextProps, nextState) {
            const {something} = nextProps;
            return something;
          }
        })
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 118 (7.37.5)',
      code: `
        var Hello = createReactClass({
          propTypes: {
            something: PropTypes.bool,
          },
          shouldComponentUpdate (nextProps, nextState) {
            const {something} = nextProps;
            return something;
          }
        })
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 119 (7.37.5)',
      code: `
        class Hello extends Component {
          static propTypes = {
            something: PropTypes.bool
          }
          shouldComponentUpdate ({something}, nextState) {
            return something;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 120 (7.37.5)',
      code: `
        class Hello extends Component {
          shouldComponentUpdate ({something}, nextState) {
            return something;
          }
        }
        Hello.propTypes = {
          something: PropTypes.bool,
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 121 (7.37.5)',
      code: `
        var Hello = createReactClass({
          propTypes: {
            something: PropTypes.bool,
          },
          shouldComponentUpdate ({something}, nextState) {
            return something;
          }
        })
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 122 (7.37.5)',
      code: `
        class Hello extends Component {
          static propTypes = {
            something: PropTypes.bool
          }
          componentWillUpdate (nextProps, nextState) {
            const {something} = nextProps;
            return something;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 123 (7.37.5)',
      code: `
        class Hello extends Component {
          componentWillUpdate (nextProps, nextState) {
            const {something} = nextProps;
            return something;
          }
        }
        Hello.propTypes = {
          something: PropTypes.bool,
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 124 (7.37.5)',
      code: `
        var Hello = createReactClass({
          propTypes: {
            something: PropTypes.bool,
          },
          componentWillUpdate (nextProps, nextState) {
            const {something} = nextProps;
            return something;
          }
        })
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 125 (7.37.5)',
      code: `
        class Hello extends Component {
          static propTypes = {
            something: PropTypes.bool
          }
          componentWillUpdate ({something}, nextState) {
            return something;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 126 (7.37.5)',
      code: `
        class Hello extends Component {
          componentWillUpdate ({something}, nextState) {
            return something;
          }
        }
        Hello.propTypes = {
          something: PropTypes.bool,
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 127 (7.37.5)',
      code: `
        var Hello = createReactClass({
          propTypes: {
            something: PropTypes.bool,
          },
          componentWillUpdate ({something}, nextState) {
            return something;
          }
        })
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 128 (7.37.5)',
      code: `
        class Hello extends Component {
          static propTypes = {
            something: PropTypes.bool
          }
          componentDidUpdate (prevProps, prevState) {
            const {something} = prevProps;
            return something;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 129 (7.37.5)',
      code: `
        class Hello extends Component {
          componentDidUpdate (prevProps, prevState) {
            const {something} = prevProps;
            return something;
          }
        }
        Hello.propTypes = {
          something: PropTypes.bool,
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 130 (7.37.5)',
      code: `
        var Hello = createReactClass({
          propTypes: {
            something: PropTypes.bool,
          },
          componentDidUpdate (prevProps, prevState) {
            const {something} = prevProps;
            return something;
          }
        })
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 131 (7.37.5)',
      code: `
        class Hello extends Component {
          static propTypes = {
            something: PropTypes.bool
          }
          componentDidUpdate ({something}, prevState) {
            return something;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 132 (7.37.5)',
      code: `
        class Hello extends Component {
          componentDidUpdate ({something}, prevState) {
            return something;
          }
        }
        Hello.propTypes = {
          something: PropTypes.bool,
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 133 (7.37.5)',
      code: `
        var Hello = createReactClass({
          propTypes: {
            something: PropTypes.bool,
          },
          componentDidUpdate ({something}, prevState) {
            return something;
          }
        })
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 134 (7.37.5)',
      code: `
        class Hello extends Component {
          static propTypes = {
            something: PropTypes.bool
          }
          componentDidUpdate ({something}, {state1, state2}) {
            return something;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 135 (7.37.5)',
      code: `
        class Hello extends Component {
          componentDidUpdate ({something}, {state1, state2}) {
            return something;
          }
        }
        Hello.propTypes = {
          something: PropTypes.bool,
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 136 (7.37.5)',
      code: `
        var Hello = createReactClass({
          propTypes: {
            something: PropTypes.bool,
          },
          componentDidUpdate ({something}, {state1, state2}) {
            return something;
          }
        })
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 137 (7.37.5)',
      code: `
        var Hello = React.Component({
          propTypes: {
            something: PropTypes.bool
          },
          componentDidUpdate: function ({something}, {state1, state2}) {
            return something;
          }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 138 (7.37.5)',
      code: `
        var Hello = createReactClass({
          propTypes: {
            something: PropTypes.bool,
          },
          componentDidUpdate: function ({something}, {state1, state2}) {
            return something;
          }
        })
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 139 (7.37.5)',
      code: `
        const Hello = (props) => {
          const {...rest} = props;
          return <div />;
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 140 (7.37.5)',
      code: `
        function SomeComponent(props) {
          const {aaa: {bbb}} = props;
          return <p>{bbb}</p>;
        }

        SomeComponent.propTypes = {
          aaa: somePropType,
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 141 (7.37.5)',
      code: `
        const PagingBlock = function(props) {
          return (
            <span>
              <a onClick={() => props.previousPage()}/>
              <a onClick={() => props.nextPage()}/>
            </span>
         );
        };

        PagingBlock.propTypes = {
          nextPage: PropTypes.func.isRequired,
          previousPage: PropTypes.func.isRequired,
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 142 (7.37.5)',
      code: `
        const PagingBlock = function(props) {
          return (
            <SomeChild {...props} />
         );
        };

        PagingBlock.propTypes = {
          nextPage: PropTypes.func.isRequired,
          previousPage: PropTypes.func.isRequired,
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 143 (7.37.5)',
      code: `
        function Foo(props) {
          useEffect(() => {
            const { a } = props;
            document.title = a;
          });

          return <p/>;
        }

        Foo.propTypes = {
          a: PropTypes.string,
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 144 (7.37.5)',
      code: `
        class Hello extends Component {
          componentWillReceiveProps (nextProps) {
            if (nextProps.foo) {
              doSomething(this.props.bar);
            }
          }
        }

        Hello.propTypes = {
          foo: PropTypes.bool,
          bar: PropTypes.bool
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 145 (7.37.5)',
      code: `
        export default function SomeComponent(props) {
            const callback = () => {
                props.a(props.b);
            };

            const anotherCallback = () => {};

            return (
                <SomeOtherComponent
                    name={props.c}
                    callback={callback}
                />
            );
        }

        SomeComponent.propTypes = {
            a: React.PropTypes.func.isRequired,
            b: React.PropTypes.string.isRequired,
            c: React.PropTypes.string.isRequired,
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 146 (7.37.5)',
      code: `
        export default function SomeComponent(props) {
            const callback = () => {
                props.a(props.b);
            };

            return (
                <SomeOtherComponent
                    name={props.c}
                    callback={callback}
                />
            );
        }

        SomeComponent.propTypes = {
            a: React.PropTypes.func.isRequired,
            b: React.PropTypes.string.isRequired,
            c: React.PropTypes.string.isRequired,
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 147 (7.37.5)',
      code: `
        class Hello extends Component {
          static propTypes = {
            foo: PropTypes.string,
            bar: PropTypes.string,
          };

          shouldComponentUpdate (props) {
            if (props.foo) {
              return true;
            }
          }

          render() {
            return (<div>{this.props.bar}</div>);
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 148 (7.37.5)',
      code: `
        class HelloFooBar extends Component {
          shouldComponentUpdate(props) {
            if (props.foo) {
              return true;
            }
          }

          render() {
            return (<div>{this.props.bar}</div>);
          }
        }
        HelloFooBar.propTypes = {
          foo: PropTypes.string,
          bar: PropTypes.string,
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 149 (7.37.5)',
      code: `
        class Hello extends Component {
          static propTypes = {
            foo: PropTypes.string,
            bar: PropTypes.string,
          };

          componentWillUpdate (props) {
            if (props.foo) {
              return true;
            }
          }

          render() {
            return (<div>{this.props.bar}</div>);
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 150 (7.37.5)',
      code: `
        class Hello extends Component {
          componentWillUpdate (props) {
            if (props.foo) {
              return true;
            }
          }

          render() {
            return (<div>{this.props.bar}</div>);
          }
        }
        Hello.propTypes = {
          foo: PropTypes.string,
          bar: PropTypes.string,
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 151 (7.37.5)',
      code: `
        class Hello extends Component {
          static propTypes = {
            foo: PropTypes.string,
            bar: PropTypes.string,
          };

          componentWillReceiveProps (nextProps) {
            const {foo} = nextProps;
            if (foo) {
              return true;
            }
          }

          render() {
            return (<div>{this.props.bar}</div>);
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 152 (7.37.5)',
      code: `
        export class Example extends Component {
          static propTypes = {
            foo: PropTypes.func,
          }
          classProperty = async () => {
            await this.props.foo();
          };
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 153 (7.37.5)',
      code: `
        export class Example extends Component {
          static propTypes = {
            foo: PropTypes.func,
            bar: PropTypes.func,
            baz: PropTypes.func,
          }
          classProperty = async () => {
            await this.props.foo();
            await this.props.bar();
            await this.props.baz();
          };
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 154 (7.37.5)',
      code: `
        class Hello extends Component {
          componentWillReceiveProps (props) {
            if (props.foo) {
              return true;
            }
          }

          render() {
            return (<div>{this.props.bar}</div>);
          }
        }
        Hello.propTypes = {
          foo: PropTypes.string,
          bar: PropTypes.string,
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 155 (7.37.5)',
      code: `
        class Hello extends Component {
          static propTypes = {
            foo: PropTypes.string,
            bar: PropTypes.string,
          };

          shouldComponentUpdate (nextProps) {
            if (nextProps.foo) {
              return true;
            }
          }

          render() {
            return (<div>{this.props.bar}</div>);
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 156 (7.37.5)',
      code: `
        export class Example extends Component {
          static propTypes = {
            foo: PropTypes.func,
          }
          classProperty = async () => {
            const { foo } = this.props;
            await foo();
          };
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 157 (7.37.5)',
      code: `
        export class Example extends Component {
          static propTypes = {
            foo: PropTypes.func,
            bar: PropTypes.func,
            baz: PropTypes.func,
          }
          classProperty = async () => {
            const { foo, bar, baz } = this.props;
            await foo();
            await bar();
            await baz();
          };
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 158 (7.37.5)',
      code: `
        class Hello extends Component {
          shouldComponentUpdate (nextProps) {
            if (nextProps.foo) {
              return true;
            }
          }

          render() {
            return (<div>{this.props.bar}</div>);
          }
        }
        Hello.propTypes = {
          foo: PropTypes.string,
          bar: PropTypes.string,
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 159 (7.37.5)',
      code: `
        class Hello extends Component {
          static propTypes = {
            foo: PropTypes.string,
            bar: PropTypes.string,
          };

          componentWillUpdate (nextProps) {
            if (nextProps.foo) {
              return true;
            }
          }

          render() {
            return (<div>{this.props.bar}</div>);
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 160 (7.37.5)',
      code: `
        export class Example extends Component {
          static propTypes = {
            foo: PropTypes.func,
          }
          async method() {
            await this.props.foo();
          };
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 161 (7.37.5)',
      code: `
        export class Example extends Component {
          static propTypes = {
            foo: PropTypes.func,
            bar: PropTypes.func,
            baz: PropTypes.func,
          }
          async method() {
            await this.props.foo();
            await this.props.bar();
            await this.props.baz();
          };
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 162 (7.37.5)',
      code: `
        export class Example extends Component {
          static propTypes = {
            foo: PropTypes.func,
          }
          async method() {
            const { foo } = this.props;
            await foo();
          };
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 163 (7.37.5)',
      code: `
        class Hello extends Component {
          componentWillUpdate (nextProps) {
            if (nextProps.foo) {
              return true;
            }
          }

          render() {
            return (<div>{this.props.bar}</div>);
          }
        }
        Hello.propTypes = {
          foo: PropTypes.string,
          bar: PropTypes.string,
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 164 (7.37.5)',
      code: `
        class Hello extends Component {
          static propTypes = {
            foo: PropTypes.string,
            bar: PropTypes.string,
          };

          componentDidUpdate (prevProps) {
            if (prevProps.foo) {
              return true;
            }
          }

          render() {
            return (<div>{this.props.bar}</div>);
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 165 (7.37.5)',
      code: `
        export class Example extends Component {
          static propTypes = {
            foo: PropTypes.func,
            bar: PropTypes.func,
            baz: PropTypes.func,
          }
          async method() {
            const { foo, bar, baz } = this.props;
            await foo();
            await bar();
            await baz();
          };
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 166 (7.37.5)',
      code: `
        export class Example extends Component {
          static propTypes = {
            foo: PropTypes.func,
            bar: PropTypes.func,
            baz: PropTypes.func,
          }
          factory() {
            return async () => {
              await this.props.foo();
              await this.props.bar();
              await this.props.baz();
            };
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 167 (7.37.5)',
      code: `
        export class Example extends Component {
          static propTypes = {
            foo: PropTypes.func,
            bar: PropTypes.func,
            baz: PropTypes.func,
          }
          factory() {
            return async function onSubmit() {
              await this.props.foo();
              await this.props.bar();
              await this.props.baz();
            };
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 168 (7.37.5)',
      code: `
        class Hello extends Component {
          componentDidUpdate (prevProps) {
            if (prevProps.foo) {
              return true;
            }
          }

          render() {
            return (<div>{this.props.bar}</div>);
          }
        }
        Hello.propTypes = {
          foo: PropTypes.string,
          bar: PropTypes.string,
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 169 (7.37.5)',
      code: `
        class Example extends Component {
          async method() {
            await this.props.foo();
            await this.props.bar();
          };
        }
        Example.propTypes = {
          foo: PropTypes.func,
          bar: PropTypes.func,
        }
      `,
      languageOptions: {
        ecmaVersion: 2017,
        sourceType: 'module',
        parserOptions: { ecmaFeatures: { jsx: true } },
      },
    },
    {
      name: 'Upstream valid 170 (7.37.5)',
      code: `
        class Example extends Component {
          render() {
            async function onSubmit() {
              await this.props.foo();
              await this.props.bar();
            }
            return <Form onSubmit={onSubmit} />
          };
        }
        Example.propTypes = {
          foo: PropTypes.func,
          bar: PropTypes.func,
        }
      `,
      languageOptions: {
        ecmaVersion: 2017,
        sourceType: 'module',
        parserOptions: { ecmaFeatures: { jsx: true } },
      },
    },
    {
      name: 'Upstream valid 171 (7.37.5)',
      code: `
        class Example extends Component {
          render() {
            const onSubmit = async () => {
              await this.props.foo();
              await this.props.bar();
            }
            return <Form onSubmit={onSubmit} />
          };
        }
        Example.propTypes = {
          foo: PropTypes.func,
          bar: PropTypes.func,
        }
      `,
      languageOptions: {
        ecmaVersion: 2017,
        sourceType: 'module',
        parserOptions: { ecmaFeatures: { jsx: true } },
      },
    },
    {
      name: 'Upstream valid 172 (7.37.5)',
      code: `
        export default class NavigationButton extends React.Component {
         static propTypes = {
           route: PropTypes.shape({
            getBarTintColor: PropTypes.func.isRequired,
          }).isRequired,
         };

         renderTitle() {
          const { route } = this.props;
           return <Title tintColor={route.getBarTintColor()}>TITLE</Title>;
         }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 173 (7.37.5)',
      code: `
        const Component = ({ children: aNode }) => (
         <div>{aNode}</div>
        );

        Component.defaultProps = {
         children: null,
        };

        Component.propTypes = {
         children: React.PropTypes.node,
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 174 (7.37.5)',
      code: `
        const Thing = (props) => (
            <div>
              {(() => {
                    if(props.enabled && props.test){
                        return (
                            <span>Enabled!</span>
                        )
                    }
                    return (
                        <span>Disabled..</span>
                    )
                })()}
            </div>
        );

        Thing.propTypes = {
            enabled: React.PropTypes.bool,
            test: React.PropTypes.bool
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 175 (7.37.5)',
      code: `
        const Test = props => <div>
          {someArray.map(l => <div
            key={l}>
              {props.property + props.property2}
            </div>)}
        </div>

        Test.propTypes = {
          property: React.propTypes.string.isRequired,
          property2: React.propTypes.string.isRequired
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 176 (7.37.5)',
      code: `
        const Button = React.createClass({
        displayName: "Button",

        propTypes: {
            name: React.PropTypes.string.isRequired,
            isEnabled: React.PropTypes.bool.isRequired
        },

        render() {
            const item = this.props;
            const disabled = !this.props.isEnabled;
            return (
                <div>
                    <button type="button" disabled={disabled}>{item.name}</button>
                </div>
            );
        }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 177 (7.37.5)',
      code: `
        class Foo extends React.Component {
         static propTypes = {
         foo: PropTypes.func.isRequired,
        }

        constructor(props) {
          super(props);

          const { foo } = props;
          this.message = foo("blablabla");
        }

        render() {
          return <div>{this.message}</div>;
        }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 178 (7.37.5)',
      code: `
        class HelloGraphQL extends Component {
          render() {
              return <div>Hello</div>;
          }
        }

        const HellowQueries = graphql(queryDetails, {
          options: ownProps => ({
          variables: ownProps.aProp
          }),
        })(HelloGraphQL)

        HellowQueries.propTypes = {
          aProp: PropTypes.string.isRequired
        }

        export default connect(mapStateToProps, mapDispatchToProps)(HellowQueries)
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 179 (7.37.5)',
      code: `
        type Props = {
         foo: {
          bar: boolean
         }
        };

        class DigitalServices extends React.Component {
         props: Props
         render() {
           const { foo } = this.props;
           return <div>{foo.bar}</div>;
         }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 180 (7.37.5)',
      code: `
        const foo = {};
        class Hello extends React.Component {
          render() {
            const {firstname, lastname} = this.props.name;
            return <div>{firstname} {lastname}</div>;
          }
        }
        Hello.propTypes = {
          name: PropTypes.shape(foo)
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 181 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            const {foo: {bar}} = this.props;
            return <div>{bar}</div>;
          }
        }
        Hello.propTypes = {
          foo: PropTypes.shape({
            bar: PropTypes.string,
          })
        };
      `,
      options: [{ skipShapeProps: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 182 (7.37.5)',
      code: `
        type Props = {
         onMouseOver: Function,
         onClick: Function,
        };

        const MyComponent = (props: Props) => (
        <div>
          <button onMouseOver={() => props.onMouseOver()} />
          <button onClick={() => props.onClick()} />
        </div>
        );
      `,
      options: [{ skipShapeProps: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 183 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          onFoo() {
            this.setState((prevState, props) => {
              props.doSomething();
            });
          }
          render() {
            return (
               <div onClick={this.onFoo}>Test</div>
            );
          }
        }
        MyComponent.propTypes = {
          doSomething: PropTypes.func
        };
        var tempVar2;
      `,
      options: [{ skipShapeProps: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 184 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          onFoo() {
            this.setState((prevState, { doSomething }) => {
              doSomething();
            });
          }
          render() {
            return (
               <div onClick={this.onFoo}>Test</div>
            );
          }
        }
        MyComponent.propTypes = {
          doSomething: PropTypes.func
        };
      `,
      options: [{ skipShapeProps: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 185 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          onFoo() {
            this.setState((prevState, obj) => {
              obj.doSomething();
            });
          }
          render() {
            return (
               <div onClick={this.onFoo}>Test</div>
            );
          }
        }
        MyComponent.propTypes = {
          doSomething: PropTypes.func
        };
        var tempVar2;
      `,
      options: [{ skipShapeProps: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 186 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          onFoo() {
            this.setState(() => {
              this.props.doSomething();
            });
          }
          render() {
            return (
               <div onClick={this.onFoo}>Test</div>
            );
          }
        }
        MyComponent.propTypes = {
          doSomething: PropTypes.func
        };
        var tempVar;
      `,
      options: [{ skipShapeProps: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 187 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          onFoo() {
            this.setState((prevState) => {
              this.props.doSomething();
            });
          }
          render() {
            return (
               <div onClick={this.onFoo}>Test</div>
            );
          }
        }
        MyComponent.propTypes = {
          doSomething: PropTypes.func
        };
      `,
      options: [{ skipShapeProps: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 188 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          onFoo() {
            this.setState(({ something }) => {
              this.props.doSomething();
            });
          }
          render() {
            return (
               <div onClick={this.onFoo}>Test</div>
            );
          }
        }
        MyComponent.propTypes = {
          doSomething: PropTypes.func
        };
      `,
      options: [{ skipShapeProps: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 189 (7.37.5)',
      code: `
        import React from 'react';
        import SharedPropTypes from './SharedPropTypes';

        export default class A extends React.Component {
          render() {
            return (
              <span
                a={this.props.a}
                b={this.props.b}
                c={this.props.c}>
                {this.props.children}
              </span>
            );
          }
        }

        A.propTypes = {
          a: React.PropTypes.string,
          ...SharedPropTypes // eslint-disable-line object-shorthand
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 190 (7.37.5)',
      code: `
        type Props = $ReadOnly<{foo: number}>;
        class Hello extends React.Component {
          props: Props;
          render () {
            return <div>{this.props.foo}</div>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 191 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream flow syntax (Property or signature expected.).',
      code: `
        type Props = {
          +foo: number
        }
        class MyComponent extends React.Component {
          render() {
            return <div>{this.props.foo}</div>
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 192 (7.37.5)',
      code: `
        type Props = {
          'completed?': boolean,
        };
        const Hello = (props: Props): React.Element => {
          return <div>{props['completed?']}</div>;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 193 (7.37.5)',
      code: `
        type Person = {
          firstname: string
        }
        class MyComponent extends React.Component<void, Props, void> {
          render() {
            return <div>Hello {this.props.firstname}</div>
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 194 (7.37.5)',
      code: `
        type Person = {
          firstname: string
        }
        class MyComponent extends React.Component<void, Props, void> {
          render() {
            return <div>Hello {this.props.firstname}</div>
          }
        }
      `,
      settings: { react: { flowVersion: '0.52' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 195 (7.37.5)',
      code: `
        type Person = {
          firstname: string
        }
        class MyComponent extends React.Component<Props> {
          render() {
            return <div>Hello {this.props.firstname}</div>
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 196 (7.37.5)',
      code: `
        type Person = {
          firstname: string
        }
        class MyComponent extends React.Component<Props> {
          render() {
            return <div>Hello {this.props.firstname}</div>
          }
        }
      `,
      settings: { react: { flowVersion: '0.53' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 197 (7.37.5)',
      code: `
      class MyComponent extends Component {
        static propTypes = {
          validate: PropTypes.bool,
          options: PropTypes.array,
          value: ({options, value, validate}) => {
            if (!validate) return;
            if (options.indexOf(value) < 0)
              throw new Errow('oops');
          }
        }

        render() {
          return <ul>
            {this.props.options.map(option =>
              <li className={this.props.value == option && "active"}>{option}</li>
            )}
          </ul>
        }
      }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 198 (7.37.5)',
      code: `
      class MyComponent extends Component {
        static propTypes = {
          validate: PropTypes.bool,
          options: PropTypes.array,
          value: function ({options, value, validate}) {
            if (!validate) return;
            if (options.indexOf(value) < 0)
              throw new Errow('oops');
          }
        }

        render() {
          return <ul>
            {this.props.options.map(option =>
              <li className={this.props.value == option && "active"}>{option}</li>
            )}
          </ul>
        }
      }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 199 (7.37.5)',
      code: `
      class MyComponent extends Component {
        static propTypes = {
          validate: PropTypes.bool,
          options: PropTypes.array,
          value({options, value, validate}) {
            if (!validate) return;
            if (options.indexOf(value) < 0)
              throw new Errow('oops');
          }
        }

        render() {
          return <ul>
            {this.props.options.map(option =>
              <li className={this.props.value == option && "active"}>{option}</li>
            )}
          </ul>
        }
      }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 200 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          render() {
            return <div>{ this.props.other }</div>
          }
        }
        MyComponent.propTypes = { other: () => {} };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 201 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          render() {
            return <div>{ this.props.other }</div>
          }
        }
        MyComponent.propTypes = { other() {} };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 202 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          render() {
            return <div>{ this.props.other }</div>
          }
        }
        MyComponent.propTypes = { other: function () {} };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 203 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          render() {
            return <div>{ this.props.other }</div>
          }
        }
        MyComponent.propTypes = { * other() {} };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 204 (7.37.5)',
      code: `
        class Hello extends Component {
          static propTypes = {
            something: PropTypes.bool
          };
          UNSAFE_componentWillReceiveProps (nextProps) {
            const {something} = nextProps;
            doSomething(something);
          }
        }
      `,
      settings: { react: { version: '16.3.0' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 205 (7.37.5)',
      code: `
        class Hello extends Component {
          static propTypes = {
            something: PropTypes.bool
          };
          UNSAFE_componentWillUpdate (nextProps, nextState) {
            const {something} = nextProps;
            return something;
          }
        }
      `,
      settings: { react: { version: '16.3.0' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 206 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          static propTypes = {
            defaultValue: 'bar'
          };
          state = {
            currentValue: null
          };
          static getDerivedStateFromProps(nextProps, prevState) {
            if (prevState.currentValue === null) {
              return {
                currentValue: nextProps.defaultValue,
              }
            }
            return null;
          }
          render() {
            return <div>{ this.state.currentValue }</div>
          }
        }
      `,
      settings: { react: { version: '16.3.0' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 207 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          static propTypes = {
            defaultValue: PropTypes.string
          };
          getSnapshotBeforeUpdate(prevProps, prevState) {
            if (prevProps.defaultValue === null) {
              return 'snapshot';
            }
            return null;
          }
          render() {
            return <div />
          }
        }
      `,
      settings: { react: { version: '16.3.0' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 208 (7.37.5)',
      code: `
        import React from 'react';
        type Props = string & {
          fullname: string
        };
        class Test extends React.PureComponent<Props> {
          render() {
            return <div>Hello {this.props.fullname}</div>
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 209 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream flow syntax (Property or signature expected.).',
      code: `
        import type {BasePerson} from './types'
        type Props = {
          person: {
           ...$Exact<BasePerson>,
           lastname: string
          }
        };
        class Hello extends React.Component {
          props: Props;
          render () {
            return <div>Hello {this.props.person.firstname}</div>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 210 (7.37.5)',
      code: `
        import BasePerson from './types'
        class Hello extends React.Component {
          render () {
            return <div>Hello {this.props.person.firstname}</div>;
          }
        }
        Hello.propTypes = {
          person: ProTypes.shape({
            ...BasePerson,
            lastname: PropTypes.string
          })
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 211 (7.37.5)',
      code: `
        type Props = {
          used: string,
        }
        class Hello extends React.Component<Props> {
          renderHelper = ({notAProp}: {notAProp: string}) => {
            return <div />;
          }
          render() {
            return <div>{this.props.used}</div>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 212 (7.37.5)',
      code: `
        type Props = {
          used: string,
        }
        class Hello extends React.Component<Props> {
          componentDidMount() {
            foo(
              ({notAProp}: {notAProp: string}) => (<div />)
            );
          }
          render() {
            return <div>{this.props.used}</div>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 213 (7.37.5)',
      code: `
        type Props = {
          used: string,
        }
        class Hello extends React.Component<Props> {
          render() {
            return <QueryRenderer
              render={({notAProp}: {notAProp: string}) => <div>{this.props.used}</div>}
            />;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 214 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            return <div>{this.props.used}</div>;
          }
        }
        Hello.propTypes = {
          used: PropTypes.string,
          foo: PropTypes.string
        };
      `,
      options: [{ ignore: ['foo'] }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 215 (7.37.5)',
      code: `
        type Props = {
          used: string,
          foo: string,
        }
        class Hello extends React.Component<Props> {
          render() {
            return <div>{this.props.used}</div>;
          }
        }
      `,
      options: [{ ignore: ['foo'] }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 216 (7.37.5)',
      code: `
        export default class Foo extends React.Component {
          render() {
            return null;
          }
        }

        Foo.defaultProps = Object.assign({});
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 217 (7.37.5)',
      code: `
        const Hello = ({a}) => (
          <div>
            {a.map(({b}) => (
              <div>{b}</div>
            ))}
          </div>
        );
        Hello.propTypes = {
          a: PropTypes.arrayOf(
            PropTypes.exact({
              b: PropTypes.string,
            })
          ),
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 218 (7.37.5)',
      code: `
        const Foo = (props) => {
          const { foo } = props as unknown;
          (props as unknown).bar as unknown;

          return <></>;
        };

        Foo.propTypes = {
          foo,
          bar,
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 219 (7.37.5)',
      code: `
        class Foo extends React.Component {
          static propTypes = {
            prevProp,
            nextProp,
            setStateProp,
            thisPropsAliasDestructProp,
            thisPropsAliasProp,
            thisDestructPropsAliasDestructProp,
            thisDestructPropsAliasProp,
            thisDestructPropsDestructProp,
            thisPropsDestructProp,
            thisPropsProp,
          };

          componentDidUpdate(prevProps) {
            (prevProps as unknown).prevProp as unknown;
          }

          shouldComponentUpdate(nextProps) {
            (nextProps as unknown).nextProp as unknown;
          }

          stateProps() {
            ((this as unknown).setState as unknown)((_, props) => (props as unknown).setStateProp as unknown);
          }

          thisPropsAlias() {
            const props = (this as unknown).props as unknown;

            const { thisPropsAliasDestructProp } = props as unknown;
            (props as unknown).thisPropsAliasProp as unknown;
          }

          thisDestructPropsAlias() {
            const { props } = this as unknown;

            const { thisDestructPropsAliasDestructProp } = props as unknown;
            (props as unknown).thisDestructPropsAliasProp as unknown;
          }

          render() {
            const { props: { thisDestructPropsDestructProp } } = this as unknown;
            const { thisPropsDestructProp } = (this as unknown).props as unknown;
            ((this as unknown).props as unknown).thisPropsProp as unknown;

            return null;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 220 (7.37.5)',
      code: `
        declare class Thing {
          constructor({ id }: { id: string });
        }
        export default Thing;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 221 (7.37.5)',
      code: `
        const Hello = (props: {firstname: string, lastname: string}) => {
            return <div {...props}></div>;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 222 (7.37.5)',
      code: `
        const Hello = (props: UnfoundProps) => {
            return <div {...props}></div>;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 223 (7.37.5)',
      code: `
        const Hello = (props: Omit<{a: string, b: string, c: string}, "a">) => {
            return <div>{props.b}</div>;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 224 (7.37.5)',
      code: `
        const Hello = (props: () => any) => {
            return <div>{props.firstname}</div>;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 225 (7.37.5)',
      code: `
        const Hello = (props: () => any) => {
            return <div>{props?.firstname}</div>;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 226 (7.37.5)',
      code: `
        interface Props {
          'aria-label': string;
        }
        export default function Component({
          'aria-label': ariaLabel,
        }: Props): JSX.Element {
          return <div aria-label={ariaLabel} />;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 227 (7.37.5)',
      code: `
        interface Props {
          ['aria-label']: string;
        }
        export default function Component({
          ['aria-label']: ariaLabel,
        }: Props): JSX.Element {
          return <div aria-label={ariaLabel} />;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 228 (7.37.5)',
      code: `
        interface Props {
          [1234]: string;
        }
        export default function Component(
          props
        : Props): JSX.Element {
          return <div aria-label={props[1234]} />;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 229 (7.37.5)',
      code: `
        interface Props {
          ['1234']: string;
        }
        export default function Component(
          props
        : Props): JSX.Element {
          return <div aria-label={props[1234]} />;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 230 (7.37.5)',
      code: `
        interface Props {
          [1234]: string;
        }
        export default function Component(
          props
        : Props): JSX.Element {
          return <div aria-label={props['1234']} />;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 231 (7.37.5)',
      code: `
        interface Props {
          [1234]: string;
        }
        export default function Component(
          props
        : Props): JSX.Element {
        const handleVerifySubmit = ({
          otp,
          }) => {
          dispatch(
            verifyOTPPhone({
              otp,
            }),
          );
        };
        return <div aria-label={props['1234']} />;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 232 (7.37.5)',
      code: `
        interface Props {
          foo: string;
        }
        const Component = (props: Props) => (
          <div>{(()=> {return props.foo})()}</div>
        )
        export default Component
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 233 (7.37.5)',
      code: `
        type User = {
          user: string;
        }

        type Props = User;

        export default (props: Props) => {
          return <div><span>{props.user}</span></div>;
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 234 (7.37.5)',
      code: `
        type User = {
          user: string;
        }

        type Props = User & UserProps;

        export default (props: Props) => {
          return <div><span>{props.user}</span></div>;
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 235 (7.37.5)',
      code: `
        interface Person {
          firstname: string
          lastname?: never
        };

        const Hello = ({firstname}: Person) => {
          return <div><span>{firstname}</span></div>;
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 236 (7.37.5)',
      code: `
        class App extends Component {
          static propTypes = {
            notifications: PropTypes.array.isRequired
          };
          customizeNotifications() {
            const props = this.props;
            return props.notifications.map((notification) => notification);
          }
          render() {
            const notifications = this.customizeNotifications();
            return (
              <View>
                {notifications.map((notification) => <Text>{notification}</Text>)}
              </View>
            );
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 237 (7.37.5)',
      code: `
        const Home = () => {
          const renderStaticList = ({
              item,
          }: {
              item: IContent;
          }) => (
              <Section
                  icon={<FlashImage />}
                  title={item.title}
                  titleFontSize={theme.typography.FONT_SIZE_24}
              >
                  <StaticFlatList
                      data={item.pointsOfSale}
                      renderItem={renderStaticItem}
                      keyExtractor={staticItemKeyExtractor}
                  />
              </Section>
          );

          return (
              <SafeAreaViewWrapper>
                  <LightStatusBar />
                  <HomeFlatList
                      ListHeaderComponent={listHeaderComponent}
                      data={home?.static}
                      renderItem={renderStaticList}
                      keyExtractor={staticListKeyExtractor}
                  />
              </SafeAreaViewWrapper>
          );
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 238 (7.37.5)',
      code: `
        const Home = () => {
          const renderStaticList = function({
              item,
          }: {
              item: IContent;
          }) {
            return (
              <Section
                  icon={<FlashImage />}
                  title={item.title}
                  titleFontSize={theme.typography.FONT_SIZE_24}
              >
                  <StaticFlatList
                      data={item.pointsOfSale}
                      renderItem={renderStaticItem}
                      keyExtractor={staticItemKeyExtractor}
                  />
              </Section>
            )
          };

          return (
              <SafeAreaViewWrapper>
                  <LightStatusBar />
                  <HomeFlatList
                      ListHeaderComponent={listHeaderComponent}
                      data={home?.static}
                      renderItem={renderStaticList}
                      keyExtractor={staticListKeyExtractor}
                  />
              </SafeAreaViewWrapper>
          );
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 239 (7.37.5)',
      code: `
        type ConnectedProps = DispatchProps &
          StateProps

        const Component = ({ prop1, prop2, prop3 }: ConnectedProps) => {
          // Do stuff
          return (
            <StyledComponent>...</StyledComponent>
          )
        }

        const mapDispatchToProps = (dispatch: ThunkDispatch<State, null, Action>) => ({
          ...bindActionCreators<ActionCreatorsMapObject<Types.RootAction>>(
            { prop1: importedAction, prop2: anotherImportedAction },
            dispatch,
          ),
        })

        const mapStateToProps = (state: State) => ({
          prop3: Selector.value(state),
        })

        type StateProps = ReturnType<typeof mapStateToProps>
        type DispatchProps = ReturnType<typeof mapDispatchToProps>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 240 (7.37.5)',
      code: `
        type ConnectedProps = DispatchProps &
          StateProps

        const Component = ({ prop1, prop2, prop3 }: ConnectedProps) => {
          // Do stuff
          return (
            <StyledComponent>...</StyledComponent>
          )
        }

        const mapDispatchToProps = (dispatch: ThunkDispatch<State, null, Action>) => ({
          ...bindActionCreators(
            { prop1: importedAction, prop2: anotherImportedAction },
            dispatch,
          ),
        })

        const mapStateToProps = (state: State) => ({
          prop3: Selector.value(state),
        })

        type StateProps = ReturnType<typeof mapStateToProps>
        type DispatchProps = ReturnType<typeof mapDispatchToProps>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 241 (7.37.5)',
      code: `
        type ConnectedProps = DispatchProps &
          StateProps

        const Component = ({ prop1, prop2, prop3 }: ConnectedProps) => {
          // Do stuff
          return (
            <StyledComponent>...</StyledComponent>
          )
        }

        const mapDispatchToProps = (dispatch: ThunkDispatch<State, null, Action>) =>
          bindActionCreators(
            { prop1: importedAction, prop2: anotherImportedAction },
            dispatch,
          )

        const mapStateToProps = (state: State) => ({
          prop3: Selector.value(state),
        })

        type StateProps = ReturnType<typeof mapStateToProps>
        type DispatchProps = ReturnType<typeof mapDispatchToProps>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 242 (7.37.5)',
      code: `
        import React from "react";

        interface Props {
          name: string;
        }

        const MyComponent: React.FC<Props> = ({ name }) => {
          return <div>{name}</div>;
        };

        export default MyComponent;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 243 (7.37.5)',
      code: `
        type Props = {
          history: {
            push: Function,
          },
          match?: {
            params?: {
              date?: string
            },
          },
        };

        export const Daily = ({ history, match: {
          params: {
            date = moment().toISOString(),
          } = {},
        } = {} }: Props) => (
              <div>
                <div style={{ textAlign: 'right', paddingRight: '25px' }}>
                  Show <DatePicker
                    selected={moment(date)}
                    className="datetime"
                    onChange={d => history.push(\`./\${d.toISOString()}\`)}
                  />
                </div>
                <WithData url="/payments/daily" body={{ date: moment(date).toISOString() }}>
                  <Flashcards date={moment(date)} />
                </WithData>
              </div>
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 244 (7.37.5)',
      code: `
        import React from 'react';

        type Props = {
          test: string,
          callback: () => void,
        };

        export default function Foo(props: Props) {
          return (
            <div>
              {[1, 2, 3].map(e => (
                <div key={e} onClick={() => props.callback()}>
                  {props.test}
                </div>
              ))}
            </div>
          );
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 245 (7.37.5)',
      code: `
        class Component extends Component <{}, {updateQueue: Array<string>}> {
          debouncedUpdate = debounce((cellsetId, options) => {
            this.setState((prevState, props) => {
              const {updateQueue} = prevState;
              const {updateCells} = props;
              updateCells(cellsetId, updateQueue, options);
              return {
                updateQueue: [],
              };
            });
          }, 2000);

          render() {
            return <div />
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 246 (7.37.5)',
      code: `
        class Test extends Component<{}, {selectedId: string}> {
          constructor(props: *) {
            super(props);
            this.state = {
              selectedId: '',
            };
          }

          onChange = ({id}: {id: string}) => { // This will say: 'id' PropType is defined but prop is never used (react/no-unused-prop-types)
            this.setState({
              selectedId: id,
            });
          };

          render() {
            const {selectedId} = this.state;
            return (
              <div>
                {selectedId}
                <select onChange={() => this.onChange({id: '1'})}>
                  <option value='1'>1</option>
                </select>
              </div>
            );
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 247 (7.37.5)',
      code: `
        function Foo (props) {
          return <div>{ renderPhoto() }</div>;

          function renderPhoto () {
            return <div>{ props.renderPhotoTools() }</div>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 248 (7.37.5)',
      code: `
        const Wrapper = featureToggle
        ? ({ children }: { children: Node }) => (
          <FeatureToggledComponent
              featureToggle={featureToggle}
              defaultValue
          >
              {children}
          </FeatureToggledComponent>
        )
        : React.Fragment;
      `,
      languageOptions: upstreamModuleOptions,
    },
  ],
  invalid: [
    {
      code: `class Hello extends React.Component { render() { return <div />; } } Hello.propTypes = { name: PropTypes.string };`,
      errors: [expectUnusedPropTypeError1([1, 90, 1, 94])],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 0 (7.37.5)',
      code: `
        var Hello = createReactClass({
          propTypes: {
            unused: PropTypes.string
          },
          render: function() {
            return React.createElement("div", {}, this.props.value);
          }
        });
      `,
      errors: [expectUnusedPropTypeError2([4, 13, 4, 19])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 1 (7.37.5)',
      code: `
        var Hello = createReactClass({
          propTypes: {
            name: PropTypes.string
          },
          render: function() {
            return <div>Hello {this.props.value}</div>;
          }
        });
      `,
      errors: [expectUnusedPropTypeError1([4, 13, 4, 17])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 2 (7.37.5)',
      code: `
        class Hello extends React.Component {
          static propTypes = {
            name: PropTypes.string
          }
          render() {
            return <div>Hello {this.props.value}</div>;
          }
        }
      `,
      errors: [expectUnusedPropTypeError1([4, 13, 4, 17])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 3 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            return <div>Hello {this.props.firstname} {this.props.lastname}</div>;
          }
        }
        Hello.propTypes = {
          unused: PropTypes.string
        };
      `,
      errors: [expectUnusedPropTypeError2([8, 11, 8, 17])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 4 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            return <div>Hello {this.props.name}</div>;
          }
        }
        Hello.propTypes = {
          unused: PropTypes.string
        };
        class HelloBis extends React.Component {
          render() {
            return <div>Hello {this.props.name}</div>;
          }
        }
      `,
      errors: [expectUnusedPropTypeError2([8, 11, 8, 17])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 5 (7.37.5)',
      code: `
        var Hello = createReactClass({
          propTypes: {
            unused: PropTypes.string.isRequired,
            anotherunused: PropTypes.string.isRequired
          },
          render: function() {
            return <div>Hello {this.props.name} and {this.props.propWithoutTypeDefinition}</div>;
          }
        });
        var Hello2 = createReactClass({
          render: function() {
            return <div>Hello {this.props.name}</div>;
          }
        });
      `,
      errors: [
        expectUnusedPropTypeError2([4, 13, 4, 19]),
        {
          messageId: 'unusedPropType',
          message: "'anotherunused' PropType is defined but prop is never used",
          line: 5,
          column: 13,
          endLine: 5,
          endColumn: 26,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 6 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            var { firstname, lastname } = this.props;
            return <div>Hello</div>;
          }
        }
        Hello.propTypes = {
          unused: PropTypes.string
        };
      `,
      errors: [expectUnusedPropTypeError2([9, 11, 9, 17])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 7 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            this.props.a.z
            return <div>Hello</div>;
          }
        }
        Hello.propTypes = {
          a: PropTypes.shape({
            b: PropTypes.string
          })
        };
      `,
      options: [{ skipShapeProps: false }],
      errors: [
        {
          messageId: 'unusedPropType',
          message: "'a.b' PropType is defined but prop is never used",
          line: 10,
          column: 13,
          endLine: 10,
          endColumn: 14,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 8 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            this.props.a.b.z;
            return <div>Hello</div>;
          }
        }
        Hello.propTypes = {
          a: PropTypes.shape({
            b: PropTypes.shape({
              c: PropTypes.string
            })
          })
        };
      `,
      options: [{ skipShapeProps: false }],
      errors: [
        {
          messageId: 'unusedPropType',
          message: "'a.b.c' PropType is defined but prop is never used",
          line: 11,
          column: 15,
          endLine: 11,
          endColumn: 16,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 9 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            this.props.a.b.c;
            this.props.a.__.d.length;
            this.props.a.anything.e[2];
            return <div>Hello</div>;
          }
        }
        Hello.propTypes = {
          a: PropTypes.objectOf(
            PropTypes.shape({
              unused: PropTypes.string
            })
          )
        };
      `,
      options: [{ skipShapeProps: false }],
      errors: [
        {
          messageId: 'unusedPropType',
          message: "'a.*.unused' PropType is defined but prop is never used",
          line: 13,
          column: 15,
          endLine: 13,
          endColumn: 21,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 10 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            var i = 3;
            this.props.a[2].c;
            this.props.a[i].d.length;
            this.props.a[i + 2].e[2];
            this.props.a.length;
            return <div>Hello</div>;
          }
        }
        Hello.propTypes = {
          a: PropTypes.arrayOf(
            PropTypes.shape({
              unused: PropTypes.string
            })
          )
        };
      `,
      options: [{ skipShapeProps: false }],
      errors: [
        {
          messageId: 'unusedPropType',
          message: "'a.*.unused' PropType is defined but prop is never used",
          line: 15,
          column: 15,
          endLine: 15,
          endColumn: 21,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 11 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            this.props.a.length;
            this.props.a.b;
            this.props.a.e.length;
            this.props.a.e.anyProp;
            this.props.a.c.toString();
            this.props.a.c.someThingElse();
            return <div>Hello</div>;
          }
        }
        Hello.propTypes = {
          a: PropTypes.oneOfType([
            PropTypes.shape({
              unused: PropTypes.number,
              anotherunused: PropTypes.array
            })
          ])
        };
      `,
      options: [{ skipShapeProps: false }],
      errors: [
        {
          messageId: 'unusedPropType',
          message: "'a.unused' PropType is defined but prop is never used",
          line: 16,
          column: 15,
          endLine: 16,
          endColumn: 21,
        },
        {
          messageId: 'unusedPropType',
          message:
            "'a.anotherunused' PropType is defined but prop is never used",
          line: 17,
          column: 15,
          endLine: 17,
          endColumn: 28,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 12 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            this.props["some.value"];
            return <div>Hello</div>;
          }
        }
        Hello.propTypes = {
          "some.unused": PropTypes.string
        };
      `,
      errors: [
        {
          messageId: 'unusedPropType',
          message: "'some.unused' PropType is defined but prop is never used",
          line: 9,
          column: 11,
          endLine: 9,
          endColumn: 24,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 13 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            this.props["arr"][1]["some.value"];
            return <div>Hello</div>;
          }
        }
        Hello.propTypes = {
          "arr": PropTypes.arrayOf(
            PropTypes.shape({
              "some.unused": PropTypes.string
        })
          )
        };
      `,
      options: [{ skipShapeProps: false }],
      errors: [
        {
          messageId: 'unusedPropType',
          message:
            "'arr.*.some.unused' PropType is defined but prop is never used",
          line: 11,
          column: 15,
          endLine: 11,
          endColumn: 28,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 14 (7.37.5)',
      code: `
        class Hello extends React.Component {
          static propTypes = {
            unused: PropTypes.string
          }
          render() {
            var text;
            text = 'Hello ';
            let {props: {firstname}} = this;
            return <div>{text} {firstname}</div>;
          }
        }
      `,
      errors: [expectUnusedPropTypeError2([4, 13, 4, 19])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 15 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            if (true) {
              return <span>{this.props.firstname}</span>
            } else {
              return <span>{this.props.lastname}</span>
            }
          }
        }
        Hello.propTypes = {
          unused: PropTypes.string
        }
      `,
      errors: [expectUnusedPropTypeError2([12, 11, 12, 17])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 16 (7.37.5)',
      code: `
        var Hello = function(props) {
          return <div>Hello {props.name}</div>;
        }
        Hello.prototype.propTypes = {unused: PropTypes.string};
      `,
      errors: [expectUnusedPropTypeError2([5, 38, 5, 44])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 17 (7.37.5)',
      code: `
        function Hello(props) {
          return <div>Hello {props.name}</div>;
        }
        Hello.prototype.propTypes = {unused: PropTypes.string};
      `,
      errors: [expectUnusedPropTypeError2([5, 38, 5, 44])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 18 (7.37.5)',
      code: `
        var Hello = (props) => {
          return <div>Hello {props.name}</div>;
        }
        Hello.prototype.propTypes = {unused: PropTypes.string};
      `,
      errors: [expectUnusedPropTypeError2([5, 38, 5, 44])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 19 (7.37.5)',
      code: `
        var Hello = (props) => {
          const {name} = props;
          return <div>Hello {name}</div>;
        }
        Hello.prototype.propTypes = {unused: PropTypes.string};
      `,
      errors: [expectUnusedPropTypeError2([6, 38, 6, 44])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 20 (7.37.5)',
      code: `
        function Hello({ name }) {
          return <div>Hello {name}</div>;
        }
        Hello.prototype.propTypes = {unused: PropTypes.string};
      `,
      errors: [expectUnusedPropTypeError2([5, 38, 5, 44])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 21 (7.37.5)',
      code: `
        const Hello = function({ name }) {
          return <div>Hello {name}</div>;
        }
        Hello.prototype.propTypes = {unused: PropTypes.string};
      `,
      errors: [expectUnusedPropTypeError2([5, 38, 5, 44])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 22 (7.37.5)',
      code: `
        const Hello = ({ name }) => {
          return <div>Hello {name}</div>;
        }
        Hello.prototype.propTypes = {unused: PropTypes.string};
      `,
      errors: [expectUnusedPropTypeError2([5, 38, 5, 44])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 23 (7.37.5)',
      code: `
        class Hello extends React.Component {
          static propTypes = {unused: PropTypes.string}
          render() {
            var props = {firstname: 'John'};
            return <div>Hello {props.firstname} {this.props.lastname}</div>;
          }
        }
      `,
      errors: [expectUnusedPropTypeError2([3, 31, 3, 37])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 24 (7.37.5)',
      code: `
        class Hello extends React.Component {
          static propTypes = {unused: PropTypes.string}
          constructor(props, context) {
            super(props, context)
            this.state = { status: props.source }
          }
        }
      `,
      errors: [expectUnusedPropTypeError2([3, 31, 3, 37])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 25 (7.37.5)',
      code: `
        class Hello extends React.Component {
          static propTypes = {unused: PropTypes.string}
          constructor(props, context) {
            super(props, context)
            this.state = { status: props.source.uri }
          }
        }
      `,
      errors: [expectUnusedPropTypeError2([3, 31, 3, 37])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 26 (7.37.5)',
      code: `
        function HelloComponent() {
          var Hello = createReactClass({
            propTypes: {unused: PropTypes.string},
            render: function() {
              return <div>Hello {this.props.name}</div>;
            }
          });
          return Hello;
        }
        module.exports = HelloComponent();
      `,
      errors: [expectUnusedPropTypeError2([4, 25, 4, 31])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 27 (7.37.5)',
      code: `
        const Hello = (props) => {
          let team = props.names.map((name) => {
              return <li>{name}, {props.company}</li>;
            });
          return <ul>{team}</ul>;
        };
        Hello.prototype.propTypes = {unused: PropTypes.string};
      `,
      errors: [expectUnusedPropTypeError2([8, 38, 8, 44])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 28 (7.37.5)',
      code: `
        const Annotation = props => (
          <div>
            {props.text}
          </div>
        )
        Annotation.prototype.propTypes = {unused: PropTypes.string};
      `,
      errors: [expectUnusedPropTypeError2([7, 43, 7, 49])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 29 (7.37.5)',
      code: `
        for (var key in foo) {
          var Hello = createReactClass({
            propTypes: {unused: PropTypes.string},
            render: function() {
              return <div>Hello {this.props.name}</div>;
            }
          });
        }
      `,
      errors: [expectUnusedPropTypeError2([4, 25, 4, 31])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 30 (7.37.5)',
      code: `
        var propTypes = {
          unused: PropTypes.string
        };
        class Test extends React.Component {
          render() {
            return (
              <div>{this.props.firstname} {this.props.lastname}</div>
            );
          }
        }
        Test.propTypes = propTypes;
      `,
      errors: [expectUnusedPropTypeError2([3, 11, 3, 17])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 31 (7.37.5)',
      code: `
        class Test extends Foo.Component {
          render() {
            return (
              <div>{this.props.firstname} {this.props.lastname}</div>
            );
          }
        }
        Test.propTypes = {
          unused: PropTypes.string
        };
      `,
      settings: { react: { pragma: 'Foo' } },
      errors: [expectUnusedPropTypeError2([10, 11, 10, 17])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 32 (7.37.5)',
      code: `
        /** @jsx Foo */
        class Test extends Foo.Component {
          render() {
            return (
              <div>{this.props.firstname} {this.props.lastname}</div>
            );
          }
        }
        Test.propTypes = {
          unused: PropTypes.string
        };
      `,
      errors: [expectUnusedPropTypeError2([11, 11, 11, 17])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 33 (7.37.5)',
      code: `
        /** @jsx Foo */
        /** @jsx React */
        class Test extends Foo.Component {
          render() {
            return (
              <div>{this.props.firstname} {this.props.lastname}</div>
            );
          }
        }
        Test.propTypes = {
          unused: PropTypes.string
        };
      `,
      errors: [expectUnusedPropTypeError2([12, 11, 12, 17])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 34 (7.37.5)',
      code: `
        /**
         * Copyright ....
         * @jsx Foo
         */
        /** @jsx React */
        class Test extends Foo.Component {
          render() {
            return (
              <div>{this.props.firstname} {this.props.lastname}</div>
            );
          }
        }
        Test.propTypes = {
          unused: PropTypes.string
        };
      `,
      errors: [expectUnusedPropTypeError2([15, 11, 15, 17])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 35 (7.37.5)',
      code: `
        class Hello extends React.Component {
          props: {
            unused: string
          };
          render () {
            return <div>Hello {this.props.name}</div>;
          }
        }
      `,
      errors: [expectUnusedPropTypeError2([4, 13, 4, 19])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 36 (7.37.5)',
      code: `
        type Props = {unused: Object;};
        class Hello extends React.Component {
          props: Props;
          render () {
            return <div>Hello {this.props.firstname}</div>;
          }
        }
      `,
      errors: [expectUnusedPropTypeError2([2, 23, 2, 29])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 37 (7.37.5)',
      skip: 'Upstream expands the Flow $ReadOnly utility; the native TypeScript model does not interpret this Flow utility.',
      code: `
        type Props = $ReadOnly<{unused: Object;}>;
        class Hello extends React.Component {
          props: Props;
          render () {
            return <div>Hello {this.props.firstname}</div>;
          }
        }
      `,
      errors: [expectUnusedPropTypeError2([2, 33, 2, 47])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 38 (7.37.5)',
      code: `
        type PropsA = { a: string }
        type PropsB = { b: string }
        type Props = PropsA & PropsB;

        class MyComponent extends React.Component {
          props: Props;

          render() {
            return <div>{this.props.a}</div>
          }
        }
      `,
      errors: [
        {
          messageId: 'unusedPropType',
          message: "'b' PropType is defined but prop is never used",
          line: 3,
          column: 25,
          endLine: 3,
          endColumn: 26,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 39 (7.37.5)',
      code: `
        type PropsA = { foo: string };
        type PropsB = { bar: string };
        type PropsC = { zap: string };
        type Props = PropsA & PropsB;

        class Bar extends React.Component {
          props: Props & PropsC;

          render() {
            return <div>{this.props.foo} - {this.props.bar}</div>
          }
        }
      `,
      errors: [expectUnusedPropTypeError3([4, 25, 4, 28])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 40 (7.37.5)',
      code: `
        type PropsB = { foo: string };
        type PropsC = { bar: string };
        type Props = PropsB & {
          zap: string
        };

        class Bar extends React.Component {
          props: Props & PropsC;

          render() {
            return <div>{this.props.foo} - {this.props.bar}</div>
          }
        }
      `,
      errors: [expectUnusedPropTypeError3([5, 11, 5, 14])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 41 (7.37.5)',
      code: `
        type PropsB = { foo: string };
        type PropsC = { bar: string };
        type Props = {
          zap: string
        } & PropsB;

        class Bar extends React.Component {
          props: Props & PropsC;

          render() {
            return <div>{this.props.foo} - {this.props.bar}</div>
          }
        }
      `,
      errors: [expectUnusedPropTypeError3([5, 11, 5, 14])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 42 (7.37.5)',
      skip: 'Upstream checks Flow prop annotations for no-unused-prop-types; the native TypeScript annotation path reports 0 diagnostics instead of 1.',
      code: `
        class Hello extends React.Component {
          props: {
            name: {
              unused: string;
            }
          };
          render () {
            return <div>Hello {this.props.name.lastname}</div>;
          }
        }
      `,
      options: [{ skipShapeProps: false }],
      errors: [
        {
          messageId: 'unusedPropType',
          message: "'name.unused' PropType is defined but prop is never used",
          line: 5,
          column: 15,
          endLine: 5,
          endColumn: 29,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 43 (7.37.5)',
      skip: 'Upstream checks Flow prop annotations for no-unused-prop-types; the native TypeScript annotation path reports 0 diagnostics instead of 1.',
      code: `
        type Props = {name: {unused: string;};};
        class Hello extends React.Component {
          props: Props;
          render () {
            return <div>Hello {this.props.name.lastname}</div>;
          }
        }
      `,
      options: [{ skipShapeProps: false }],
      errors: [
        {
          messageId: 'unusedPropType',
          message: "'name.unused' PropType is defined but prop is never used",
          line: 2,
          column: 30,
          endLine: 2,
          endColumn: 44,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 44 (7.37.5)',
      skip: 'Upstream checks Flow prop annotations for no-unused-prop-types; the native TypeScript annotation path reports 0 diagnostics instead of 1.',
      code: `
        class Hello extends React.Component {
          props: {person: {name: {unused: string;};};};
          render () {
            return <div>Hello {this.props.person.name.lastname}</div>;
          }
        }
      `,
      options: [{ skipShapeProps: false }],
      errors: [
        {
          messageId: 'unusedPropType',
          message:
            "'person.name.unused' PropType is defined but prop is never used",
          line: 3,
          column: 35,
          endLine: 3,
          endColumn: 49,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 45 (7.37.5)',
      skip: 'Upstream checks Flow prop annotations for no-unused-prop-types; the native TypeScript annotation path reports 0 diagnostics instead of 1.',
      code: `
        type Props = {person: {name: {unused: string;};};};
        class Hello extends React.Component {
          props: Props;
          render () {
            return <div>Hello {this.props.person.name.lastname}</div>;
          }
        }
      `,
      options: [{ skipShapeProps: false }],
      errors: [
        {
          messageId: 'unusedPropType',
          message:
            "'person.name.unused' PropType is defined but prop is never used",
          line: 2,
          column: 39,
          endLine: 2,
          endColumn: 53,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 46 (7.37.5)',
      skip: 'Upstream checks Flow prop annotations for no-unused-prop-types; the native TypeScript annotation path reports 0 diagnostics instead of 1.',
      code: `
        type Person = {name: {unused: string;}};
        class Hello extends React.Component {
          props: {people: Person[];};
          render () {
            var names = [];
            for (var i = 0; i < this.props.people.length; i++) {
              names.push(this.props.people[i].name.lastname);
            }
            return <div>Hello {names.join(', ')}</div>;
          }
        }
      `,
      options: [{ skipShapeProps: false }],
      errors: [
        {
          messageId: 'unusedPropType',
          message:
            "'people.*.name.unused' PropType is defined but prop is never used",
          line: 2,
          column: 31,
          endLine: 2,
          endColumn: 45,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 47 (7.37.5)',
      skip: 'Upstream checks Flow prop annotations for no-unused-prop-types; the native TypeScript annotation path reports 0 diagnostics instead of 1.',
      code: `
        type Person = {name: {unused: string;}};
        type Props = {people: Person[];};
        class Hello extends React.Component {
          props: Props;
          render () {
            var names = [];
            for (var i = 0; i < this.props.people.length; i++) {
              names.push(this.props.people[i].name.lastname);
            }
            return <div>Hello {names.join(', ')}</div>;
          }
        }
      `,
      options: [{ skipShapeProps: false }],
      errors: [
        {
          messageId: 'unusedPropType',
          message:
            "'people.*.name.unused' PropType is defined but prop is never used",
          line: 2,
          column: 31,
          endLine: 2,
          endColumn: 45,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 48 (7.37.5)',
      skip: 'Upstream checks Flow prop annotations for no-unused-prop-types; the native TypeScript annotation path reports 0 diagnostics instead of 2.',
      code: `
        type Props = {result?: {ok: string | boolean;}|{ok: number | Array}};
        class Hello extends React.Component {
          props: Props;
          render () {
            return <div>Hello {this.props.result.notok}</div>;
          }
        }
      `,
      options: [{ skipShapeProps: false }],
      errors: [
        {
          messageId: 'unusedPropType',
          message: "'result.ok' PropType is defined but prop is never used",
          line: 2,
          column: 33,
          endLine: 2,
          endColumn: 53,
        },
        {
          messageId: 'unusedPropType',
          message: "'result.ok' PropType is defined but prop is never used",
          line: 2,
          column: 57,
          endLine: 2,
          endColumn: 75,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 49 (7.37.5)',
      code: `
        function Greetings({names}) {
          names = names.map(({firstname, lastname}) => <div>{firstname} {lastname}</div>);
          return <Hello>{names}</Hello>;
        }
        Greetings.propTypes = {unused: Object};
      `,
      errors: [expectUnusedPropTypeError2([6, 32, 6, 38])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 50 (7.37.5)',
      code: `
        const MyComponent = props => (
          <div onClick={() => props.toggle()}></div>
        )
        MyComponent.propTypes = {unused: Object};
      `,
      errors: [expectUnusedPropTypeError2([5, 34, 5, 40])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 51 (7.37.5)',
      code: `
        const MyComponent = props => props.test ? <div /> : <span />
        MyComponent.propTypes = {unused: Object};
      `,
      errors: [expectUnusedPropTypeError2([3, 34, 3, 40])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 52 (7.37.5)',
      code: `
        type Props = {
          unused: ?string,
        };
        function Hello({firstname, lastname}: Props): React$Element {
          return <div>Hello {firstname} {lastname}</div>;
        }
      `,
      errors: [expectUnusedPropTypeError2([3, 11, 3, 17])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 53 (7.37.5)',
      code: `
        class Hello extends Component {
          static propTypes = {
            unused: PropTypes.bool
          }
          constructor (props) {
            super(props);
            const {something} = props;
            doSomething(something);
          }
        }
      `,
      errors: [expectUnusedPropTypeError2([4, 13, 4, 19])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 54 (7.37.5)',
      code: `
        class Hello extends Component {
          static propTypes = {
            unused: PropTypes.bool
          }
          constructor ({something}) {
            super({something});
            doSomething(something);
          }
        }
      `,
      errors: [expectUnusedPropTypeError2([4, 13, 4, 19])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 55 (7.37.5)',
      code: `
        class Hello extends Component {
          static propTypes = {
            unused: PropTypes.bool
          }
          componentWillReceiveProps (nextProps, nextState) {
            const {something} = nextProps;
            return something;
          }
        }
      `,
      errors: [expectUnusedPropTypeError2([4, 13, 4, 19])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 56 (7.37.5)',
      code: `
        class Hello extends Component {
          static propTypes = {
            unused: PropTypes.bool
          }
          componentWillReceiveProps ({something}, nextState) {
            return something;
          }
        }
      `,
      errors: [expectUnusedPropTypeError2([4, 13, 4, 19])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 57 (7.37.5)',
      code: `
        class Hello extends Component {
          static propTypes = {
            unused: PropTypes.bool
          }
          shouldComponentUpdate (nextProps, nextState) {
            const {something} = nextProps;
            return something;
          }
        }
      `,
      errors: [expectUnusedPropTypeError2([4, 13, 4, 19])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 58 (7.37.5)',
      code: `
        class Hello extends Component {
          static propTypes = {
            unused: PropTypes.bool
          }
          shouldComponentUpdate ({something}, nextState) {
            return something;
          }
        }
      `,
      errors: [expectUnusedPropTypeError2([4, 13, 4, 19])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 59 (7.37.5)',
      code: `
        class Hello extends Component {
          static propTypes = {
            unused: PropTypes.bool
          }
          componentWillUpdate (nextProps, nextState) {
            const {something} = nextProps;
            return something;
          }
        }
      `,
      errors: [expectUnusedPropTypeError2([4, 13, 4, 19])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 60 (7.37.5)',
      code: `
        class Hello extends Component {
          static propTypes = {
            unused: PropTypes.bool
          }
          componentWillUpdate ({something}, nextState) {
            return something;
          }
        }
      `,
      errors: [expectUnusedPropTypeError2([4, 13, 4, 19])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 61 (7.37.5)',
      code: `
        class Hello extends Component {
          static propTypes = {
            unused: PropTypes.bool
          }
          componentDidUpdate (prevProps, prevState) {
            const {something} = prevProps;
            return something;
          }
        }
      `,
      errors: [expectUnusedPropTypeError2([4, 13, 4, 19])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 62 (7.37.5)',
      code: `
        class Hello extends Component {
          static propTypes = {
            unused: PropTypes.bool
          }
          componentDidUpdate ({something}, prevState) {
            return something;
          }
        }
      `,
      errors: [expectUnusedPropTypeError2([4, 13, 4, 19])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 63 (7.37.5)',
      code: `
        class Hello extends Component {
          static propTypes = {
            something: PropTypes.bool
          }
          componentDidUpdate (prevProps, {state1, state2}) {
            return something;
          }
        }
      `,
      errors: [expectUnusedPropTypeError4([4, 13, 4, 22])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 64 (7.37.5)',
      code: `
        var Hello = createReactClass({
          propTypes: {
            something: PropTypes.bool
          },
          componentDidUpdate: function (prevProps, {state1, state2}) {
            return something;
          }
        })
      `,
      errors: [expectUnusedPropTypeError4([4, 13, 4, 22])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 65 (7.37.5)',
      code: `
        class Hello extends Component {
          static propTypes = {
            foo: PropTypes.string,
            bar: PropTypes.string,
          };

          componentWillUpdate (nextProps) {
            if (nextProps.foo) {
              return true;
            }
          }
        }
      `,
      errors: [expectUnusedPropTypeError5([5, 13, 5, 16])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 66 (7.37.5)',
      code: `
        export class Example extends Component {
          static propTypes = {
            foo: PropTypes.func,
            bar: PropTypes.func,
            baz: PropTypes.func,
          }
          classProperty = async () => {
            await this.props.foo();
            await this.props.bar();
          };
        }
      `,
      errors: [expectUnusedPropTypeError6([6, 13, 6, 16])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 67 (7.37.5)',
      code: `
        class Hello extends Component {
          componentWillUpdate (nextProps) {
            if (nextProps.foo) {
              return true;
            }
          }
        }
        Hello.propTypes = {
          foo: PropTypes.string,
          bar: PropTypes.string,
        };
      `,
      errors: [expectUnusedPropTypeError5([11, 11, 11, 14])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 68 (7.37.5)',
      code: `
        class Hello extends Component {
          static propTypes = {
            foo: PropTypes.string,
            bar: PropTypes.string,
          };

          shouldComponentUpdate (nextProps) {
            if (nextProps.foo) {
              return true;
            }
          }
        }
      `,
      errors: [expectUnusedPropTypeError5([5, 13, 5, 16])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 69 (7.37.5)',
      code: `
        export class Example extends Component {
          static propTypes = {
            foo: PropTypes.func,
            bar: PropTypes.func,
            baz: PropTypes.func,
          }
          classProperty = async () => {
            const { bar, baz } = this.props;
            await bar();
            await baz();
          };
        }
      `,
      errors: [expectUnusedPropTypeError7([4, 13, 4, 16])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 70 (7.37.5)',
      code: `
        export class Example extends Component {
          static propTypes = {
            foo: PropTypes.func,
            bar: PropTypes.func,
            baz: PropTypes.func,
          }
          async method() {
            await this.props.foo();
            await this.props.baz();
          };
        }
      `,
      errors: [expectUnusedPropTypeError5([5, 13, 5, 16])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 71 (7.37.5)',
      code: `
        class Hello extends Component {
          shouldComponentUpdate (nextProps) {
            if (nextProps.foo) {
              return true;
            }
          }
        }
        Hello.propTypes = {
          foo: PropTypes.string,
          bar: PropTypes.string,
        };
      `,
      errors: [expectUnusedPropTypeError5([11, 11, 11, 14])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 72 (7.37.5)',
      code: `
        class Hello extends Component {
          static propTypes = {
            foo: PropTypes.string,
            bar: PropTypes.string,
          };

          componentDidUpdate (nextProps) {
            if (nextProps.foo) {
              return true;
            }
          }
        }
      `,
      errors: [expectUnusedPropTypeError5([5, 13, 5, 16])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 73 (7.37.5)',
      code: `
        export class Example extends Component {
          static propTypes = {
            foo: PropTypes.func,
            bar: PropTypes.func,
            baz: PropTypes.func,
          }
          async method() {
            const { foo, bar } = this.props;
            await foo();
            await bar();
          };
        }
      `,
      errors: [expectUnusedPropTypeError6([6, 13, 6, 16])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 74 (7.37.5)',
      code: `
        export class Example extends Component {
          static propTypes = {
            foo: PropTypes.func,
            bar: PropTypes.func,
            baz: PropTypes.func,
          }
          factory() {
            return async () => {
              await this.props.foo();
              await this.props.bar();
            };
          }
        }
      `,
      errors: [expectUnusedPropTypeError6([6, 13, 6, 16])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 75 (7.37.5)',
      code: `
        class Hello extends Component {
          componentDidUpdate (nextProps) {
            if (nextProps.foo) {
              return true;
            }
          }
        }
        Hello.propTypes = {
          foo: PropTypes.string,
          bar: PropTypes.string,
        };
      `,
      errors: [expectUnusedPropTypeError5([11, 11, 11, 14])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 76 (7.37.5)',
      code: `
        class Hello extends Component {
          componentDidUpdate (nextProps) {
            if (nextProps.foo) {
              return true;
            }
          }
        }
        Hello.propTypes = forbidExtraProps({
          foo: PropTypes.string,
          bar: PropTypes.string,
        });
      `,
      settings: { propWrapperFunctions: ['forbidExtraProps'] },
      errors: [expectUnusedPropTypeError5([11, 11, 11, 14])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 77 (7.37.5)',
      code: `
        class Hello extends Component {
          propTypes = forbidExtraProps({
            foo: PropTypes.string,
            bar: PropTypes.string
          });
          componentDidUpdate (nextProps) {
            if (nextProps.foo) {
              return true;
            }
          }
        };
      `,
      settings: { propWrapperFunctions: ['forbidExtraProps'] },
      errors: [expectUnusedPropTypeError5([5, 13, 5, 16])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 78 (7.37.5)',
      code: `
        export class Example extends Component {
          static propTypes = {
            foo: PropTypes.func,
            bar: PropTypes.func,
            baz: PropTypes.func,
          }
          factory() {
            return async function onSubmit() {
              await this.props.bar();
              await this.props.baz();
            };
          }
        }
      `,
      errors: [expectUnusedPropTypeError7([4, 13, 4, 16])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 79 (7.37.5)',
      code: `
        class Example extends Component {
          render() {
            async function onSubmit() {
              await this.props.foo();
              await this.props.bar();
            }
            return <Form onSubmit={onSubmit} />
          };
        }
        Example.propTypes = {
          foo: PropTypes.func,
          bar: PropTypes.func,
          baz: PropTypes.func,
        }
      `,
      errors: [expectUnusedPropTypeError6([14, 11, 14, 14])],
      languageOptions: {
        ecmaVersion: 2017,
        sourceType: 'module',
        parserOptions: { ecmaFeatures: { jsx: true } },
      },
    },
    {
      name: 'Upstream invalid 80 (7.37.5)',
      code: `
        class Example extends Component {
          render() {
            const onSubmit = async () => {
              await this.props.bar();
              await this.props.baz();
            }
            return <Form onSubmit={onSubmit} />
          };
        }
        Example.propTypes = {
          foo: PropTypes.func,
          bar: PropTypes.func,
          baz: PropTypes.func,
        }
      `,
      errors: [expectUnusedPropTypeError7([12, 11, 12, 14])],
      languageOptions: {
        ecmaVersion: 2017,
        sourceType: 'module',
        parserOptions: { ecmaFeatures: { jsx: true } },
      },
    },
    {
      name: 'Upstream invalid 81 (7.37.5)',
      code: `
        import React from "react";
        var Hello = React.createReactClass({
         propTypes: {
           name: React.PropTypes.string
         },
         render: function() {
           return <div>Hello Bob</div>;
          }
        });
      `,
      errors: [expectUnusedPropTypeError1([5, 12, 5, 16])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 82 (7.37.5)',
      code: `
        class Comp1 extends Component {
          render() {
            return <span />;
          }
        }
        Comp1.propTypes = {
          prop1: PropTypes.number
        };
        class Comp2 extends Component {
          render() {
            return <span />;
          }
        }
        Comp2.propTypes = {
          prop2: PropTypes.arrayOf(Comp1.propTypes.prop1)
        };
      `,
      errors: [
        expectUnusedPropTypeError8([8, 11, 8, 16]),
        expectUnusedPropTypeError9([16, 11, 16, 16]),
        expectUnusedPropTypeError10([16, 36, 16, 57]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 83 (7.37.5)',
      code: `
        class Comp1 extends Component {
          render() {
            return <span />;
          }
        }
        Comp1.propTypes = {
          prop1: PropTypes.number
        };
        class Comp2 extends Component {
          static propTypes = {
            prop2: PropTypes.arrayOf(Comp1.propTypes.prop1)
          }
          render() {
            return <span />;
          }
        }
      `,
      errors: [
        expectUnusedPropTypeError8([8, 11, 8, 16]),
        expectUnusedPropTypeError9([12, 13, 12, 18]),
        expectUnusedPropTypeError10([12, 38, 12, 59]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 84 (7.37.5)',
      code: `
        class Comp1 extends Component {
          render() {
            return <span />;
          }
        }
        Comp1.propTypes = {
          prop1: PropTypes.number
        };
        var Comp2 = createReactClass({
          propTypes: {
            prop2: PropTypes.arrayOf(Comp1.propTypes.prop1)
          },
          render() {
            return <span />;
          }
        });
      `,
      errors: [
        expectUnusedPropTypeError8([8, 11, 8, 16]),
        expectUnusedPropTypeError9([12, 13, 12, 18]),
        expectUnusedPropTypeError10([12, 38, 12, 59]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 85 (7.37.5)',
      code: `
        class HelloGraphQL extends Component {
          render() {
              return <div>Hello</div>;
          }
        }
        HelloGraphQL.propTypes = {
          aProp: PropTypes.string.isRequired
        }

        const HellowQueries = graphql(queryDetails, {
          options: ownProps => ({
          variables: ownProps.aProp
          }),
        })(HelloGraphQL)

        export default connect(mapStateToProps, mapDispatchToProps)(HellowQueries)
      `,
      errors: [
        {
          messageId: 'unusedPropType',
          message: "'aProp' PropType is defined but prop is never used",
          line: 8,
          column: 11,
          endLine: 8,
          endColumn: 16,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 86 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream flow syntax (Property or signature expected.).',
      code: `
        type UsedProps = {|
          usedProp: number,
        |};

        type UnusedProps = {|
          unusedProp: number,
        |};

        type Props = {| ...UsedProps, ...UnusedProps |};

        function MyComponent({ usedProp, notOne }: Props) {
          return <div>{usedProp}</div>;
        }
      `,
      errors: [
        {
          messageId: 'unusedPropType',
          message: "'unusedProp' PropType is defined but prop is never used",
          line: 7,
          column: 11,
          endLine: 7,
          endColumn: 29,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 87 (7.37.5)',
      skip: 'Upstream Flow React.Component<void, Props, State> puts props in the second type argument; the native TypeScript model uses the first.',
      code: `
        type Props = {
          firstname: string,
          lastname: string,
        }
        class MyComponent extends React.Component<void, Props, void> {
          render() {
            return <div>Hello {this.props.firstname}</div>
          }
        }
      `,
      errors: [expectUnusedPropTypeError11([4, 11, 4, 27])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 88 (7.37.5)',
      skip: 'Upstream Flow React.Component<void, Props, State> puts props in the second type argument; the native TypeScript model uses the first.',
      code: `
        type Props = {
          firstname: string,
          lastname: string,
        }
        class MyComponent extends React.Component<void, Props, void> {
          render() {
            return <div>Hello {this.props.firstname}</div>
          }
        }
      `,
      settings: { react: { flowVersion: '0.52' } },
      errors: [expectUnusedPropTypeError11([4, 11, 4, 27])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 89 (7.37.5)',
      code: `
        type Props = {
          firstname: string,
          lastname: string,
        }
        class MyComponent extends React.Component<Props> {
          render() {
            return <div>Hello {this.props.firstname}</div>
          }
        }
      `,
      errors: [expectUnusedPropTypeError11([4, 11, 4, 19])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 90 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          render() {
            return <div>Hello {this.props.firstname}</div>
          }
        }
        MyComponent.propTypes = {
          firstname: PropTypes.string,
          lastname: PropTypes.string,
          foo: PropTypes.string,
        };
      `,
      options: [{ ignore: ['foo'] }],
      errors: [expectUnusedPropTypeError11([9, 11, 9, 19])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 91 (7.37.5)',
      code: `
        type Props = {
          firstname: string,
          lastname: string,
          foo: string,
        }
        class MyComponent extends React.Component<Props> {
          render() {
            return <div>Hello {this.props.firstname}</div>
          }
        }
      `,
      options: [{ ignore: ['foo'] }],
      errors: [expectUnusedPropTypeError11([4, 11, 4, 19])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 92 (7.37.5)',
      code: `
        type Person = string;
        class Hello extends React.Component<{ person: Person }> {
          render () {
            return <div />;
          }
        }
      `,
      settings: { react: { flowVersion: '0.53' } },
      errors: [
        {
          messageId: 'unusedPropType',
          message: "'person' PropType is defined but prop is never used",
          line: 3,
          column: 47,
          endLine: 3,
          endColumn: 53,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 93 (7.37.5)',
      skip: 'Upstream Flow React.Component<void, Props, State> puts props in the second type argument; the native TypeScript model uses the first.',
      code: `
        type Person = string;
        class Hello extends React.Component<void, { person: Person }, void> {
          render () {
            return <div />;
          }
        }
      `,
      settings: { react: { flowVersion: '0.52' } },
      errors: [
        {
          messageId: 'unusedPropType',
          message: "'person' PropType is defined but prop is never used",
          line: 3,
          column: 53,
          endLine: 3,
          endColumn: 67,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 94 (7.37.5)',
      skip: "The native TypeScript parser cannot parse this upstream flow syntax (',' expected.).",
      code: `
          function higherOrderComponent<P: { foo: string }>() {
            return class extends React.Component<P> {
              render() {
                return <div />;
              }
            }
          }
        `,
      errors: [expectUnusedPropTypeError7([2, 46, 2, 57])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 95 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          onFoo() {
            this.setState(({ doSomething }, props) => {
              return { doSomething: doSomething + 1 };
            });
          }
          render() {
            return (
               <div onClick={this.onFoo}>Test</div>
            );
          }
        }
        MyComponent.propTypes = {
          doSomething: PropTypes.func
        };
      `,
      errors: [
        {
          messageId: 'unusedPropType',
          message: "'doSomething' PropType is defined but prop is never used",
          line: 15,
          column: 11,
          endLine: 15,
          endColumn: 22,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 96 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          onFoo() {
            this.setState(prevState => ({
              doSomething: prevState.doSomething + 1,
            }));
          }
          render() {
            return (
               <div onClick={this.onFoo}>Test</div>
            );
          }
        }
        MyComponent.propTypes = {
          doSomething: PropTypes.func
        };
      `,
      errors: [
        {
          messageId: 'unusedPropType',
          message: "'doSomething' PropType is defined but prop is never used",
          line: 15,
          column: 11,
          endLine: 15,
          endColumn: 22,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 97 (7.37.5)',
      code: `
        type Props = {
          firstname: string,
          lastname: string,
        }
        class MyComponent extends React.Component<Props> {
          render() {
            return <div>Hello {this.props.firstname}</div>
          }
        }
      `,
      settings: { react: { flowVersion: '0.53' } },
      errors: [expectUnusedPropTypeError11([4, 11, 4, 19])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 98 (7.37.5)',
      code: `
        class Hello extends Component {
          static propTypes = {
            something: PropTypes.bool
          };
          UNSAFE_componentWillReceiveProps (nextProps) {
            const {something} = nextProps;
            doSomething(something);
          }
        }
      `,
      settings: { react: { version: '16.2.0' } },
      errors: [expectUnusedPropTypeError4([4, 13, 4, 22])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 99 (7.37.5)',
      code: `
        class Hello extends Component {
          static propTypes = {
            something: PropTypes.bool
          };
          UNSAFE_componentWillUpdate (nextProps, nextState) {
            const {something} = nextProps;
            return something;
          }
        }
      `,
      settings: { react: { version: '16.2.0' } },
      errors: [expectUnusedPropTypeError4([4, 13, 4, 22])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 100 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          static propTypes = {
            defaultValue: 'bar'
          };
          state = {
            currentValue: null
          };
          static getDerivedStateFromProps(nextProps, prevState) {
            if (prevState.currentValue === null) {
              return {
                currentValue: nextProps.defaultValue,
              }
            }
            return null;
          }
          render() {
            return <div>{ this.state.currentValue }</div>
          }
        }
      `,
      settings: { react: { version: '16.2.0' } },
      errors: [
        {
          messageId: 'unusedPropType',
          message: "'defaultValue' PropType is defined but prop is never used",
          line: 4,
          column: 13,
          endLine: 4,
          endColumn: 25,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 101 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          static propTypes = {
            defaultValue: PropTypes.string
          };
          getSnapshotBeforeUpdate(prevProps, prevState) {
            if (prevProps.defaultValue === null) {
              return 'snapshot';
            }
            return null;
          }
          render() {
            return <div />
          }
        }
      `,
      settings: { react: { version: '16.2.0' } },
      errors: [
        {
          messageId: 'unusedPropType',
          message: "'defaultValue' PropType is defined but prop is never used",
          line: 4,
          column: 13,
          endLine: 4,
          endColumn: 25,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 102 (7.37.5)',
      code: `
        import React from 'react';
        type OtherProps = {
          firstname: string,
          lastname: string,
        } | {
          fullname: string
        };
        type Props = OtherProps & {
          age: number
        };
        class Test extends React.PureComponent<Props> {
          render() {
            return <div>Hello {this.props.firstname}</div>
          }
        }
      `,
      errors: [
        {
          messageId: 'unusedPropType',
          message: "'age' PropType is defined but prop is never used",
          line: 10,
          column: 11,
          endLine: 10,
          endColumn: 14,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 103 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            return <div>Hello</div>;
          }
        }
        Hello.propTypes = {
          a: PropTypes.shape({
            b: PropTypes.shape({
            })
          })
        };
        Hello.propTypes.a.b.c = PropTypes.number;
      `,
      options: [{ skipShapeProps: false }],
      errors: [
        {
          messageId: 'unusedPropType',
          message: "'a' PropType is defined but prop is never used",
          line: 8,
          column: 11,
          endLine: 8,
          endColumn: 12,
        },
        {
          messageId: 'unusedPropType',
          message: "'a.b' PropType is defined but prop is never used",
          line: 9,
          column: 13,
          endLine: 9,
          endColumn: 14,
        },
        {
          messageId: 'unusedPropType',
          message: "'a.b.c' PropType is defined but prop is never used",
          line: 13,
          column: 9,
          endLine: 13,
          endColumn: 49,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 104 (7.37.5)',
      code: `
        type Props = { foo: string }
        function higherOrderComponent<Props>() {
          return class extends React.Component<Props> {
            render() {
              return <div />;
            }
          }
        }
      `,
      errors: [expectUnusedPropTypeError7([2, 24, 2, 27])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 105 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream flow syntax (Property or signature expected.).',
      code: `
        type Person = {
          ...data,
          lastname: string
        };
        class Hello extends React.Component {
          props: Person;
          render () {
            return <div>Hello {this.props.firstname}</div>;
          }
        }
      `,
      errors: [expectUnusedPropTypeError11([4, 11, 4, 27])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 106 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream flow syntax (Property or signature expected.).',
      code: `
        type Person = {|
          ...data,
          lastname: string
        |};
        class Hello extends React.Component {
          props: Person;
          render () {
            return <div>Hello {this.props.firstname}</div>;
          }
        }
      `,
      errors: [expectUnusedPropTypeError11([4, 11, 4, 27])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 107 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream flow syntax (Property or signature expected.).',
      code: `
        type Person = {
          ...$Exact<data>,
          lastname: string
        };
        class Hello extends React.Component {
          props: Person;
          render () {
            return <div>Hello {this.props.firstname}</div>;
          }
        }
      `,
      errors: [expectUnusedPropTypeError11([4, 11, 4, 27])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 108 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream flow syntax (Property or signature expected.).',
      code: `
        import type {Data} from './Data'
        type Person = {
          ...Data,
          lastname: string
        };
        class Hello extends React.Component {
          props: Person;
          render () {
            return <div>Hello {this.props.bar}</div>;
          }
        }
      `,
      errors: [expectUnusedPropTypeError11([5, 11, 5, 27])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 109 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream flow syntax (Property or signature expected.).',
      code: `
        import type {Data} from 'some-libdef-like-flow-typed-provides'
        type Person = {
          ...Data,
          lastname: string
        };
        class Hello extends React.Component {
          props: Person;
          render () {
            return <div>Hello {this.props.bar}</div>;
          }
        }
      `,
      errors: [expectUnusedPropTypeError11([5, 11, 5, 27])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 110 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render () {
            return <div>Hello {this.props.firstname}</div>;
          }
        }
        Hello.propTypes = {
          ...BasePerson,
          lastname: PropTypes.string
        };
      `,
      errors: [expectUnusedPropTypeError11([9, 11, 9, 19])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 111 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream flow syntax (Property or signature expected.).',
      code: `
        import type {BasePerson} from './types'
        type Props = {
          person: {
           ...$Exact<BasePerson>,
           lastname: string
          }
        };
        class Hello extends React.Component {
          props: Props;
          render () {
            return <div>Hello {this.props.person.firstname}</div>;
          }
        }
      `,
      options: [{ skipShapeProps: false }],
      errors: [
        {
          messageId: 'unusedPropType',
          message:
            "'person.lastname' PropType is defined but prop is never used",
          line: 6,
          column: 12,
          endLine: 6,
          endColumn: 28,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 112 (7.37.5)',
      code: `
        import BasePerson from './types'
        class Hello extends React.Component {
          render () {
            return <div>Hello {this.props.person.firstname}</div>;
          }
        }
        Hello.propTypes = {
          person: ProTypes.shape({
            ...BasePerson,
            lastname: PropTypes.string
          })
        };
      `,
      options: [{ skipShapeProps: false }],
      errors: [
        {
          messageId: 'unusedPropType',
          message:
            "'person.lastname' PropType is defined but prop is never used",
          line: 11,
          column: 13,
          endLine: 11,
          endColumn: 21,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 113 (7.37.5)',
      code: `
        type Props = {notTarget: string, unused: string};
        class Hello extends React.Component {
          props: Props;
          onEvent = ({ target }: { target: Object }) => {};
          render () {
            return <div>Hello {this.props.notTarget}</div>;
          }
        }
      `,
      errors: [expectUnusedPropTypeError2([2, 42, 2, 48])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 114 (7.37.5)',
      code: `
        import PropTypes from 'prop-types';
        import React from 'react';

        const MyComponent= (props) => {
          switch (props.usedProp) {
            case 1:
              return (<div />);
            default:
              return <div />;
          }
        };

        MyComponent.propTypes = {
          usedProp: PropTypes.string,
          unUsedProp: PropTypes.string,
        };

        export default MyComponent;
      `,
      errors: [
        {
          messageId: 'unusedPropType',
          message: "'unUsedProp' PropType is defined but prop is never used",
          line: 16,
          column: 11,
          endLine: 16,
          endColumn: 21,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 115 (7.37.5)',
      code: `
        const Foo = (props) => {
          const { foo } = props as unknown;
          (props as unknown).bar as unknown;

          return <></>;
        };

        Foo.propTypes = {
          fooUnused,
          barUnused,
        };
      `,
      errors: [
        {
          messageId: 'unusedPropType',
          message: "'fooUnused' PropType is defined but prop is never used",
          line: 10,
          column: 11,
          endLine: 10,
          endColumn: 20,
        },
        {
          messageId: 'unusedPropType',
          message: "'barUnused' PropType is defined but prop is never used",
          line: 11,
          column: 11,
          endLine: 11,
          endColumn: 20,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 116 (7.37.5)',
      code: `
        class Foo extends React.Component {
          static propTypes = {
            prevPropUnused,
            nextPropUnused,
            setStatePropUnused,
            thisPropsAliasDestructPropUnused,
            thisPropsAliasPropUnused,
            thisDestructPropsAliasDestructPropUnused,
            thisDestructPropsAliasPropUnused,
            thisDestructPropsDestructPropUnused,
            thisPropsDestructPropUnused,
            thisPropsPropUnused,
          };

          componentDidUpdate(prevProps) {
            (prevProps as unknown).prevProp as unknown;
          }

          shouldComponentUpdate(nextProps) {
            (nextProps as unknown).nextProp as unknown;
          }

          stateProps() {
            ((this as unknown).setState as unknown)((_, props) => (props as unknown).setStateProp as unknown);
          }

          thisPropsAlias() {
            const props = (this as unknown).props as unknown;

            const { thisPropsAliasDestructProp } = props as unknown;
            (props as unknown).thisPropsAliasProp as unknown;
          }

          thisDestructPropsAlias() {
            const { props } = this as unknown;

            const { thisDestructPropsAliasDestructProp } = props as unknown;
            (props as unknown).thisDestructPropsAliasProp as unknown;
          }

          render() {
            const { props: { thisDestructPropsDestructProp } } = this as unknown;
            const { thisPropsDestructProp } = (this as unknown).props as unknown;
            ((this as unknown).props as unknown).thisPropsProp as unknown;

            return null;
          }
        }
      `,
      errors: [
        {
          messageId: 'unusedPropType',
          message:
            "'prevPropUnused' PropType is defined but prop is never used",
          line: 4,
          column: 13,
          endLine: 4,
          endColumn: 27,
        },
        {
          messageId: 'unusedPropType',
          message:
            "'nextPropUnused' PropType is defined but prop is never used",
          line: 5,
          column: 13,
          endLine: 5,
          endColumn: 27,
        },
        {
          messageId: 'unusedPropType',
          message:
            "'setStatePropUnused' PropType is defined but prop is never used",
          line: 6,
          column: 13,
          endLine: 6,
          endColumn: 31,
        },
        {
          messageId: 'unusedPropType',
          message:
            "'thisPropsAliasDestructPropUnused' PropType is defined but prop is never used",
          line: 7,
          column: 13,
          endLine: 7,
          endColumn: 45,
        },
        {
          messageId: 'unusedPropType',
          message:
            "'thisPropsAliasPropUnused' PropType is defined but prop is never used",
          line: 8,
          column: 13,
          endLine: 8,
          endColumn: 37,
        },
        {
          messageId: 'unusedPropType',
          message:
            "'thisDestructPropsAliasDestructPropUnused' PropType is defined but prop is never used",
          line: 9,
          column: 13,
          endLine: 9,
          endColumn: 53,
        },
        {
          messageId: 'unusedPropType',
          message:
            "'thisDestructPropsAliasPropUnused' PropType is defined but prop is never used",
          line: 10,
          column: 13,
          endLine: 10,
          endColumn: 45,
        },
        {
          messageId: 'unusedPropType',
          message:
            "'thisDestructPropsDestructPropUnused' PropType is defined but prop is never used",
          line: 11,
          column: 13,
          endLine: 11,
          endColumn: 48,
        },
        {
          messageId: 'unusedPropType',
          message:
            "'thisPropsDestructPropUnused' PropType is defined but prop is never used",
          line: 12,
          column: 13,
          endLine: 12,
          endColumn: 40,
        },
        {
          messageId: 'unusedPropType',
          message:
            "'thisPropsPropUnused' PropType is defined but prop is never used",
          line: 13,
          column: 13,
          endLine: 13,
          endColumn: 32,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 117 (7.37.5)',
      code: `
        type Person = {
          lastname: string
        };
        const Hello = (props: Person) => {
            return <div>Hello {props.firstname}</div>;
        }
      `,
      errors: [expectUnusedPropTypeError11([3, 11, 3, 19])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 118 (7.37.5)',
      code: `
        type Person = {
          lastname: string
        };
        const Hello = (props: Person) => {
            return <div>Hello {props?.firstname}</div>;
        }
      `,
      errors: [expectUnusedPropTypeError11([3, 11, 3, 19])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 120 (7.37.5)',
      code: `
        type Person = {
          lastname?: string
        };
        const Hello = (props: Person) => {
            return <div>Hello {props.firstname}</div>;
        }
      `,
      errors: [expectUnusedPropTypeError11([3, 11, 3, 19])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 121 (7.37.5)',
      code: `
        type Person = {
          lastname?: string
        };
        const Hello = (props: Person) => {
            return <div>Hello {props?.firstname}</div>;
        }
      `,
      errors: [expectUnusedPropTypeError11([3, 11, 3, 19])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 122 (7.37.5)',
      code: `
        type Person = {
          firstname: string
          lastname: string
        };
        const Hello = ({firstname}: Person) => {
            return <div>Hello {firstname}</div>;
        }
      `,
      errors: [expectUnusedPropTypeError11([4, 11, 4, 19])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 123 (7.37.5)',
      code: `
        interface Person {
          firstname: string
          lastname: string
        };
        const Hello = ({firstname}: Person) => {
            return <div>Hello {firstname}</div>;
        }
      `,
      errors: [expectUnusedPropTypeError11([4, 11, 4, 19])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 124 (7.37.5)',
      code: `
        interface Foo {
          foo: string;
          [blah: string]: number;
        }
        const Hello = ({bar}: Foo) => {
            return <div>Hello {bar}</div>;
        }
      `,
      errors: [expectUnusedPropTypeError7([3, 11, 3, 14])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 125 (7.37.5)',
      code: `
        interface Props {
          'aria-label': string;
        }
        export default function Component(props: Props): JSX.Element {
          return <div />;
        }
      `,
      errors: [
        {
          messageId: 'unusedPropType',
          message: "'aria-label' PropType is defined but prop is never used",
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 23,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 126 (7.37.5)',
      code: `
        interface Props {
        [1234]: string;
        }
        export default function Component(props: Props): JSX.Element {
          return <div />;
        }
      `,
      errors: [
        {
          messageId: 'unusedPropType',
          message: "'1234' PropType is defined but prop is never used",
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 14,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 127 (7.37.5)',
      code: `
        const Hello = ({firstname}: {firstname: string, lastname: string}) => {
            return <div>Hello {firstname}</div>;
        }
      `,
      errors: [expectUnusedPropTypeError11([2, 57, 2, 65])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 128 (7.37.5)',
      code: `
        interface Foo {
          x: number;
        }

        interface Foo {
          z: string;
        }

        interface Bar extends Foo {
          y: string;
        }

        const Baz = ({ x, y }: Bar) => (
          <span>
              {x}
              {y}
          </span>
        );
      `,
      errors: [expectUnusedPropTypeError12([7, 11, 7, 12])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 129 (7.37.5)',
      code: `
        interface Foo {
          x: number;
        }

        interface Bar extends Foo {
          y: string;
        }

        const Baz = ({ x }: Bar) => (
          <span>
              {x}
          </span>
        );
      `,
      errors: [expectUnusedPropTypeError13([7, 11, 7, 12])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 130 (7.37.5)',
      code: `
        interface Foo {
          x: number;
        }

        interface Bar {
          y: string;
        }

        interface Baz {
          z:string;
        }

        const Baz = ({ x }: Bar & Foo & Baz) => (
          <span>
              {x}
          </span>
        );
      `,
      errors: [
        expectUnusedPropTypeError13([7, 11, 7, 12]),
        expectUnusedPropTypeError12([11, 11, 11, 12]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 131 (7.37.5)',
      code: `
        interface Foo {
          x: number;
        }

        interface Foo {
          z: string;
        }

        interface Bar extends Foo {
          y: string;
        }

        const Baz = ({ x }: Bar) => (
          <span>
              {x}
          </span>
        );
      `,
      errors: [
        expectUnusedPropTypeError12([7, 11, 7, 12]),
        expectUnusedPropTypeError13([11, 11, 11, 12]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 132 (7.37.5)',
      code: `
        interface Foo {
          x: number;
        }

        interface Foo {
          z: string;
        }

        interface Foo {
          y: string;
        }

        const Baz = ({ x }: Foo) => (
          <span>
              {x}
          </span>
        );
      `,
      errors: [
        expectUnusedPropTypeError12([7, 11, 7, 12]),
        expectUnusedPropTypeError13([11, 11, 11, 12]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 133 (7.37.5)',
      code: `
        type User = {
          user: string;
        }

        type UserProps = {
          userId: string;
        }

        type AgeProps = {
          age: number;
        }

        type BirthdayProps = {
          birthday: string;
        }

        type intersectionUserProps = AgeProps & BirthdayProps;

        type Props = User & UserProps & intersectionUserProps;

        export default (props: Props) => {
          const { userId, user } = props;

          if (userId === 0) {
            return <p>userId is 0</p>;
          }

          return null;
        };
      `,
      errors: [
        {
          messageId: 'unusedPropType',
          message: "'age' PropType is defined but prop is never used",
          line: 11,
          column: 11,
          endLine: 11,
          endColumn: 14,
        },
        {
          messageId: 'unusedPropType',
          message: "'birthday' PropType is defined but prop is never used",
          line: 15,
          column: 11,
          endLine: 15,
          endColumn: 19,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 134 (7.37.5)',
      code: `
        const mapStateToProps = state => ({
          books: state.books
        });

        interface InfoLibTableProps extends ReturnType<typeof mapStateToProps> {
        }

        const App = (props: InfoLibTableProps) => {
          return <div></div>;
        }
      `,
      errors: [expectUnusedPropTypeError14([3, 11, 3, 16])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 135 (7.37.5)',
      code: `
        const mapStateToProps = state => ({
          books: state.books,
        });

        interface BooksTable extends ReturnType<typeof mapStateToProps> {
          username: string;
        }

        const App = (props: BooksTable) => {
          return <div />;
        }
      `,
      errors: [
        expectUnusedPropTypeError14([3, 11, 3, 16]),
        expectUnusedPropTypeError15([7, 11, 7, 19]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 136 (7.37.5)',
      code: `
        interface BooksTable extends ReturnType<() => {books:Array<string>}> {
          username: string;
        }

        const App = (props: BooksTable) => {
          return <div></div>;
        }
      `,
      errors: [
        expectUnusedPropTypeError14([2, 56, 2, 61]),
        expectUnusedPropTypeError15([3, 11, 3, 19]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 137 (7.37.5)',
      code: `
        type BooksTable = ReturnType<() => {books:Array<string>}> & {
          username: string;
        }

        const App = (props: BooksTable) => {
          return <div></div>;
        }
      `,
      errors: [
        expectUnusedPropTypeError14([2, 45, 2, 50]),
        expectUnusedPropTypeError15([3, 11, 3, 19]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 138 (7.37.5)',
      code: `
        type mapStateToProps = ReturnType<() => {books:Array<string>}>;

        type Props = {
          username: string;
        }

        type BooksTable = mapStateToProps & Props;

        const App = (props: BooksTable) => {
          return <div></div>;
        }
      `,
      errors: [
        expectUnusedPropTypeError14([2, 50, 2, 55]),
        expectUnusedPropTypeError15([5, 11, 5, 19]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 139 (7.37.5)',
      code: `
        type ConnectedProps = DispatchProps &
          StateProps

        const Component = ({ prop2, prop3 }: ConnectedProps) => {
          // Do stuff
          return (
            <StyledComponent>...</StyledComponent>
          )
        }

        const mapDispatchToProps = (dispatch: ThunkDispatch<State, null, Action>) => ({
          ...bindActionCreators<{prop1: ()=>void,prop2: ()=>string}>(
            { prop1: importedAction, prop2: anotherImportedAction },
            dispatch,
          ),
        })

        const mapStateToProps = (state: State) => ({
          prop3: Selector.value(state),
        })

        type StateProps = ReturnType<typeof mapStateToProps>
        type DispatchProps = ReturnType<typeof mapDispatchToProps>
      `,
      errors: [expectUnusedPropTypeError8([13, 34, 13, 39])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 140 (7.37.5)',
      code: `
        function Foo(props) {
          const { bar: { nope } } = props;
          return <div test={nope} />;
        }
        Foo.propTypes = {
          foo: PropTypes.number,
          bar: PropTypes.shape({
            faz: PropTypes.number,
            qaz: PropTypes.object,
          }),
        };
      `,
      errors: [expectUnusedPropTypeError7([7, 11, 7, 14])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 141 (7.37.5)',
      code: `
        interface Props {
          readonly firstname: string;
          readonly lastname: string;
        }

        class TestComponent extends React.Component<Props> {
          public render() {
            return <div>{this.props.firstname}</div>;
          }
        }
      `,
      errors: [expectUnusedPropTypeError11([4, 20, 4, 28])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 142 (7.37.5)',
      code: `
        import React from "react";

        var Hello = React.createClass({
          propTypes: {
            name: React.PropTypes.string,
            foo: React.PropTypes.string,
            propTypes: React.PropTypes.string
          },
          render: function() {
            return <div>Hello {this.props.name}</div>;
          }
        });
      `,
      settings: { react: { createClass: 'createClass' } },
      errors: [
        expectUnusedPropTypeError7([7, 13, 7, 16]),
        {
          messageId: 'unusedPropType',
          message: "'propTypes' PropType is defined but prop is never used",
          line: 8,
          column: 13,
          endLine: 8,
          endColumn: 22,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 143 (7.37.5)',
      code: `
        import React from "react";

        type props = {
          foo: string;
          bar: string;
        };

        const Demo: React.FC<props> = ({ foo }) => {
          return <div {...{}}>{foo}</div>;
        };

        export default Demo;
      `,
      errors: [expectUnusedPropTypeError5([6, 11, 6, 14])],
      languageOptions: upstreamModuleOptions,
    },
  ],
});

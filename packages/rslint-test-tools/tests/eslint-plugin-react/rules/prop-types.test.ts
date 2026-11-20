import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, upstreamModuleOptions } from '../config';

function expectMissingPropTypeError1(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'missingPropType',
    message: "'name' is missing in props validation",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectMissingPropTypeError2(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'missingPropType',
    message: "'foo' is missing in props validation",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectMissingPropTypeError3(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'missingPropType',
    message: "'foo.bar' is missing in props validation",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectMissingPropTypeError4(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'missingPropType',
    message: "'a.nope' is missing in props validation",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectMissingPropTypeError5(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'missingPropType',
    message: "'lastname' is missing in props validation",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectMissingPropTypeError6(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'missingPropType',
    message: "'a.b' is missing in props validation",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectMissingPropTypeError7(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'missingPropType',
    message: "'a.b.c' is missing in props validation",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectMissingPropTypeError8(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'missingPropType',
    message: "'firstname' is missing in props validation",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectMissingPropTypeError9(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'missingPropType',
    message: "'b' is missing in props validation",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectMissingPropTypeError10(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'missingPropType',
    message: "'source' is missing in props validation",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectMissingPropTypeError11(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'missingPropType',
    message: "'text' is missing in props validation",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectMissingPropTypeError12(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'missingPropType',
    message: "'name.lastname' is missing in props validation",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectMissingPropTypeError13(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'missingPropType',
    message: "'result.notok' is missing in props validation",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectMissingPropTypeError14(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'missingPropType',
    message: "'test' is missing in props validation",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectMissingPropTypeError15(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'missingPropType',
    message: "'bar' is missing in props validation",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectMissingPropTypeError16(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'missingPropType',
    message: "'person.lastname' is missing in props validation",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectMissingPropTypeError17(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'missingPropType',
    message: "'fooBar' is missing in props validation",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectMissingPropTypeError18(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'missingPropType',
    message: "'foo.baz' is missing in props validation",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectMissingPropTypeError19(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'missingPropType',
    message: "'cryptoCurrency' is missing in props validation",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectMissingPropTypeError20(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'missingPropType',
    message: "'first.name' is missing in props validation",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectMissingPropTypeError21(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'missingPropType',
    message: "'username' is missing in props validation",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectMissingPropTypeError22(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'missingPropType',
    message: "'foo.c' is missing in props validation",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectMissingPropTypeError23(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'missingPropType',
    message: "'l' is missing in props validation",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectMissingPropTypeError24(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'missingPropType',
    message: "'nonExistent' is missing in props validation",
    line,
    column,
    endLine,
    endColumn,
  };
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('prop-types', {
  valid: [
    {
      code: `function Hello({ name }) { return <div>{name}</div>; } Hello.propTypes = { name: PropTypes.string };`,
    },
    { code: `function Hello() { return <div>Hello</div>; }` },

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
            return <div>Hello World {this.props.children}</div>;
          }
        });
      `,
      options: [{ ignore: ['children'] }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 4 (7.37.5)',
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
      name: 'Upstream valid 5 (7.37.5)',
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
      name: 'Upstream valid 6 (7.37.5)',
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
      name: 'Upstream valid 7 (7.37.5)',
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
      name: 'Upstream valid 8 (7.37.5)',
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
      name: 'Upstream valid 9 (7.37.5)',
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
      name: 'Upstream valid 10 (7.37.5)',
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
      name: 'Upstream valid 11 (7.37.5)',
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
      name: 'Upstream valid 12 (7.37.5)',
      code: `
        class Hello {
          method;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 13 (7.37.5)',
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
      name: 'Upstream valid 14 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            var { firstname, ...other } = this.props;
            return <div>Hello {firstname}</div>;
          }
        }
        Hello.propTypes = {
          firstname: PropTypes.string
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 15 (7.37.5)',
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
      name: 'Upstream valid 16 (7.37.5)',
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 25 (7.37.5)',
      code: `
        class Component extends React.Component {
          render() {
            return <div>{this.props.foo.baz}</div>;
          }
        }
        Component.propTypes = {
          foo: PropTypes.oneOfType([
            PropTypes.shape({
              bar: PropTypes.string
            }),
            PropTypes.shape({
              baz: PropTypes.string
            })
          ])
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 26 (7.37.5)',
      code: `
        class Component extends React.Component {
          render() {
            return <div>{this.props.foo.baz}</div>;
          }
        }
        Component.propTypes = {
          foo: PropTypes.oneOfType([
            PropTypes.shape({
              bar: PropTypes.string
            }),
            PropTypes.instanceOf(Baz)
          ])
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 27 (7.37.5)',
      code: `
        class Component extends React.Component {
          render() {
            return <div>{this.props.foo.baz}</div>;
          }
        }
        Component.propTypes = {
          foo: PropTypes.oneOf(['bar', 'baz'])
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 28 (7.37.5)',
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
      name: 'Upstream valid 29 (7.37.5)',
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
      name: 'Upstream valid 30 (7.37.5)',
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
      name: 'Upstream valid 31 (7.37.5)',
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
      name: 'Upstream valid 32 (7.37.5)',
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
      name: 'Upstream valid 33 (7.37.5)',
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 34 (7.37.5)',
      code: `
        class ArrayLengthTest extends React.Component {
          render() {
            use(this.props.arr.length)
            use(this.props.arr2.length)
            return <div />
          }
        }

        ArrayLengthTest.propTypes = {
          arr: PropTypes.array,
          arr2: PropTypes.arrayOf(PropTypes.number),
        }`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 35 (7.37.5)',
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
      name: 'Upstream valid 36 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            return <div>{this.props.name.firstname}</div>;
          }
        }
      `,
      options: [{ ignore: ['name'] }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 37 (7.37.5)',
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 38 (7.37.5)',
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
      name: 'Upstream valid 39 (7.37.5)',
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
      name: 'Upstream valid 40 (7.37.5)',
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
      name: 'Upstream valid 41 (7.37.5)',
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
      name: 'Upstream valid 42 (7.37.5)',
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
      options: [{ customValidators: ['CustomValidator'] }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 43 (7.37.5)',
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
      options: [{ customValidators: ['CustomValidator'] }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 44 (7.37.5)',
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
      options: [{ customValidators: ['CustomValidator'] }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 45 (7.37.5)',
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
      name: 'Upstream valid 46 (7.37.5)',
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 47 (7.37.5)',
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 48 (7.37.5)',
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 49 (7.37.5)',
      code: `
        const SomeComponent = createReactClass({
          propTypes: SomeOtherComponent.propTypes
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 50 (7.37.5)',
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
      name: 'Upstream valid 51 (7.37.5)',
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
      name: 'Upstream valid 52 (7.37.5)',
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
      name: 'Upstream valid 53 (7.37.5)',
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
      name: 'Upstream valid 54 (7.37.5)',
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
      name: 'Upstream valid 55 (7.37.5)',
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
      name: 'Upstream valid 56 (7.37.5)',
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
      name: 'Upstream valid 57 (7.37.5)',
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
      name: 'Upstream valid 58 (7.37.5)',
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
      name: 'Upstream valid 59 (7.37.5)',
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
      name: 'Upstream valid 60 (7.37.5)',
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
      name: 'Upstream valid 61 (7.37.5)',
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
      name: 'Upstream valid 64 (7.37.5)',
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
      name: 'Upstream valid 65 (7.37.5)',
      code: `
        const statelessComponent = (props) => {
          const subRender = () => {
            return <span>{props.someProp}</span>;
          };
          return <div>{subRender()}</div>;
        };
        statelessComponent.propTypes = {
          someProp: PropTypes.string
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 66 (7.37.5)',
      code: `
        const statelessComponent = ({ someProp }) => {
          const subRender = () => {
            return <span>{someProp}</span>;
          };
          return <div>{subRender()}</div>;
        };
        statelessComponent.propTypes = {
          someProp: PropTypes.string
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 67 (7.37.5)',
      code: `
        const statelessComponent = function({ someProp }) {
          const subRender = () => {
            return <span>{someProp}</span>;
          };
          return <div>{subRender()}</div>;
        };
        statelessComponent.propTypes = {
          someProp: PropTypes.string
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 68 (7.37.5)',
      code: `
        function statelessComponent({ someProp }) {
          const subRender = () => {
            return <span>{someProp}</span>;
          };
          return <div>{subRender()}</div>;
        };
        statelessComponent.propTypes = {
          someProp: PropTypes.string
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 69 (7.37.5)',
      code: `
        function notAComponent({ something }) {
          return something + 1;
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 70 (7.37.5)',
      code: `
        const notAComponent = function({ something }) {
          return something + 1;
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 71 (7.37.5)',
      code: `
        const notAComponent = ({ something }) => {
          return something + 1;
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 72 (7.37.5)',
      code: `
        const statelessComponent = (props) => {
          let newProps = props;
          return <span>{newProps.someProp}</span>;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 73 (7.37.5)',
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
      name: 'Upstream valid 74 (7.37.5)',
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
      name: 'Upstream valid 75 (7.37.5)',
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
      name: 'Upstream valid 76 (7.37.5)',
      code: `
        type Props = {'data-action': string};
        function Button({ 'data-action': dataAction }: Props) {
          return <div data-action={dataAction} />;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 77 (7.37.5)',
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
      name: 'Upstream valid 78 (7.37.5)',
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
      name: 'Upstream valid 79 (7.37.5)',
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
      name: 'Upstream valid 80 (7.37.5)',
      code: `
        type Props = {name: {firstname: string; lastname: string;};};
        class Hello extends React.Component {
          props: Props;
          render () {
            return <div>Hello {this.props.name}</div>;
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
            return <div>Hello {names.join(
        )}</div>;
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
            return <div>Hello {names.join(
        )}</div>;
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
            return <div>Hello {names.join(
        )}</div>;
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
      name: 'Upstream valid 88 (7.37.5)',
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
      name: 'Upstream valid 89 (7.37.5)',
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
      name: 'Upstream valid 90 (7.37.5)',
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
      name: 'Upstream valid 91 (7.37.5)',
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
      name: 'Upstream valid 92 (7.37.5)',
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
      name: 'Upstream valid 93 (7.37.5)',
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
      name: 'Upstream valid 94 (7.37.5)',
      code: `
        type Props = {
          name: string,
        };
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
      name: 'Upstream valid 98 (7.37.5)',
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
      name: 'Upstream valid 99 (7.37.5)',
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
      name: 'Upstream valid 100 (7.37.5)',
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
      name: 'Upstream valid 101 (7.37.5)',
      code: `
        function Greetings() {
          return <div>{({name}) => <Hello name={name} />}</div>
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 102 (7.37.5)',
      code: `
        function Greetings() {
          return <div>{function({name}) { return <Hello name={name} />; }}</div>
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 103 (7.37.5)',
      code: `
        export default (ComposedComponent) => class Something extends SomeOtherComponent {
          someMethod = ({width}) => {}
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 104 (7.37.5)',
      code: `
        @asyncConnect([{
          promise: ({dispatch}) => {}
        }])
        class Something extends Component {}
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 105 (7.37.5)',
      code: `
        function Hello(props) {
          const {...rest} = props;
          return <div>Hello</div>;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 106 (7.37.5)',
      code: `
        let Greetings = class extends React.Component {
          render () {
            return <div>Hello {this.props.name}</div>;
          }
        };
        Greetings.propTypes = {
          name: PropTypes.string
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 107 (7.37.5)',
      code: `
        let Greetings = {
          Hello: class extends React.Component {
            render () {
              return <div>Hello {this.props.name}</div>;
            }
          }
        }
        Greetings.Hello.propTypes = {
          name: PropTypes.string
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 108 (7.37.5)',
      code: `
        let Greetings = {};
        Greetings.Hello = class extends React.Component {
          render () {
            return <div>Hello {this.props.name}</div>;
          }
        };
        Greetings.Hello.propTypes = {
          name: PropTypes.string
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 109 (7.37.5)',
      code: `
        function Hello({names}) {
          return names.map((name) => {
            return <div>{name}</div>;
          });
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 110 (7.37.5)',
      code: `
        type Target = { target: EventTarget }
        class MyComponent extends Component {
          static propTypes = {
            children: PropTypes.any,
          }
          handler({ target }: Target) {}
          render() {
            return <div>{this.props.children}</div>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 111 (7.37.5)',
      code: `
        class Hello extends Component {}
        Hello.Foo = ({foo}) => (
          <div>Hello {foo}</div>
        )
        Hello.Foo.propTypes = {
          foo: PropTypes.node
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 112 (7.37.5)',
      code: `
        var Hello = createReactClass({
          render: function() {
            return <div>{this.props.name}</div>;
          }
        });
      `,
      options: [{ skipUndeclared: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 113 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            return <div>{this.props.name}</div>;
          }
        }
      `,
      options: [{ skipUndeclared: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 114 (7.37.5)',
      code: `
        var Hello = createReactClass({
          propTypes: {
            name: PropTypes.object.isRequired
          },
          render: function() {
            return <div>{this.props.name}</div>;
          }
        });
      `,
      options: [{ skipUndeclared: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 115 (7.37.5)',
      code: `
        var Hello = createReactClass({
          propTypes: {
            name: PropTypes.object.isRequired
          },
          render: function() {
            return <div>{this.props.name}</div>;
          }
        });
      `,
      options: [{ skipUndeclared: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 116 (7.37.5)',
      code: `
        var Hello = async function* (props) {
          yield null;
          return <div>Hello {props.name}</div>;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 117 (7.37.5)',
      code: `
        async function* Hello(props) {
          yield null;
          return <div>Hello {props.name}</div>;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 118 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream flow syntax (Property or signature expected.).',
      code: `
        type Props = {
          +firstname: string;
          -lastname: string;
        };
        function Hello(props: Props): React.Element {
          const {firstname, lastname} = props;
          return <div>Hello {firstname} {lastname}</div>
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 119 (7.37.5)',
      code: `
        class Hello extends React.Component {
          async onSelect({ name }) {
            return null;
          }
          render() {
            return <Greeting onSelect={this.onSelect} />;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 120 (7.37.5)',
      code: `
        export class Example extends Component {
          static propTypes = {
            onDelete: PropTypes.func.isRequired
          }
          handleDeleteConfirm = () => {
            this.props.onDelete();
          };
          handleSubmit = async ({ certificate, key }) => {};
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 121 (7.37.5)',
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 122 (7.37.5)',
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 123 (7.37.5)',
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 124 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream flow syntax (Property or signature expected.).',
      code: `
        import type { Data } from './Data'
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 125 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream flow syntax (Property or signature expected.).',
      code: `
        import type { Data } from 'some-libdef-like-flow-typed-provides'
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 126 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream flow syntax (Property or signature expected.).',
      code: `
        import type { BasePerson } from './types'
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
      name: 'Upstream valid 127 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream flow syntax (Property or signature expected.).',
      code: `
        type OtherProps = {
          firstname: string,
        };
        type Props = {
           ...OtherProps,
           lastname: string
        };
        class Hello extends React.Component {
          props: Props;
          render () {
            return <div>Hello {this.props.firstname}</div>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 128 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream flow syntax (Property or signature expected.).',
      code: `
        type FooProps = {
          ...any,
          ...42,
        }
        function Foo(props: FooProps) {
          return <p />;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 129 (7.37.5)',
      code: `
        type Person = {
          firstname: string
        };
        class Hello extends React.Component<void, Person, void> {
          render () {
            return <div>Hello {this.props.firstname}</div>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 130 (7.37.5)',
      code: `
        type Person = {
          firstname: string
        };
        class Hello extends React.Component<void, Person, void> {
          render () {
            return <div>Hello {this.props.firstname}</div>;
          }
        }
      `,
      settings: { react: { flowVersion: '0.52' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 131 (7.37.5)',
      code: `
        type Person = {
          firstname: string
        };
        class Hello extends React.Component<void, Person, void> {
          render () {
            const { firstname } = this.props;
            return <div>Hello {firstname}</div>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 132 (7.37.5)',
      code: `
        type Person = {
          firstname: string
        };
        class Hello extends React.Component<void, Person, void> {
          render () {
            const { firstname } = this.props;
            return <div>Hello {firstname}</div>;
          }
        }
      `,
      settings: { react: { flowVersion: '0.52' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 133 (7.37.5)',
      code: `
        type Props = { name: { firstname: string; }; };
        class Hello extends React.Component<void, Props, void> {
          render () {
            return <div>Hello {this.props.name.firstname}</div>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 134 (7.37.5)',
      code: `
        type Props = { name: { firstname: string; }; };
        class Hello extends React.Component<void, Props, void> {
          render () {
            return <div>Hello {this.props.name.firstname}</div>;
          }
        }
      `,
      settings: { react: { flowVersion: '0.52' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 135 (7.37.5)',
      code: `
        type Note = { text: string, children?: Note[] };
        type Props = {
          notes: Note[];
        };
        class Hello extends React.Component<void, Props, void> {
          render () {
            return <div>Hello {this.props.notes[0].text}</div>;
          }
        }
      `,
      settings: { react: { flowVersion: '0.52' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 136 (7.37.5)',
      code: `
        import type Props from "fake";
        class Hello extends React.Component<void, Props, void> {
          render () {
            return <div>Hello {this.props.firstname}</div>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 137 (7.37.5)',
      code: `
        import type Props from "fake";
        class Hello extends React.Component<void, Props, void> {
          render () {
            return <div>Hello {this.props.firstname}</div>;
          }
        }
      `,
      settings: { react: { flowVersion: '0.52' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 138 (7.37.5)',
      code: `
        type Person = {
          firstname: string
        };
        class Hello extends React.Component<void, { person: Person }, void> {
          render () {
            return <div>Hello {this.props.person.firstname}</div>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 139 (7.37.5)',
      code: `
        type Person = {
          firstname: string
        };
        class Hello extends React.Component<void, { person: Person }, void> {
          render () {
            return <div>Hello {this.props.person.firstname}</div>;
          }
        }
      `,
      settings: { react: { flowVersion: '0.52' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 140 (7.37.5)',
      code: `
        type Props = { result?: { ok?: ?string | boolean; } | { ok?: ?number | Array } };
        class Hello extends React.Component<void, Props, void> {
          render () {
            return <div>Hello {this.props.result.ok}</div>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 141 (7.37.5)',
      code: `
        type Props = { result?: { ok?: ?string | boolean;} | { ok?: ?number | Array } };
        class Hello extends React.Component<void, Props, void> {
          render () {
            return <div>Hello {this.props.result.ok}</div>;
          }
        }
      `,
      settings: { react: { flowVersion: '0.52' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 142 (7.37.5)',
      code: `
        type Props = {
          foo: string,
        };

        class Bar extends React.Component<Props> {
          render() {
            return <div>{this.props.foo}</div>
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 143 (7.37.5)',
      code: `
        type Props = {
          foo: string,
        };

        class Bar extends React.Component<Props> {
          render() {
            return <div>{this.props.foo}</div>
          }
        }
      `,
      settings: { react: { flowVersion: '0.52' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 144 (7.37.5)',
      code: `
        type Props = {
          foo: string,
        };

        class Bar extends React.Component<Props> {
          render() {
            return <div>{this.props.foo}</div>
          }
        }
      `,
      settings: { react: { flowVersion: '0.53' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 145 (7.37.5)',
      code: `
        type FancyProps = {
          foo: string,
        };

        class Bar extends React.Component<FancyProps> {
          render() {
            return <div>{this.props.foo}</div>
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 146 (7.37.5)',
      code: `
        type FancyProps = {
          foo: string,
        };

        class Bar extends React.Component<FancyProps> {
          render() {
            return <div>{this.props.foo}</div>
          }
        }
      `,
      settings: { react: { flowVersion: '0.53' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 147 (7.37.5)',
      code: `
        type PropsA = { foo: string };
        type PropsB = { bar: string };
        type Props = PropsA & PropsB;

        class Bar extends React.Component {
          props: Props;

          render() {
            return <div>{this.props.foo} - {this.props.bar}</div>
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 148 (7.37.5)',
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
      name: 'Upstream valid 149 (7.37.5)',
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
      name: 'Upstream valid 150 (7.37.5)',
      code: `
        type PropsA = { bar: string };
        type PropsB = { zap: string };
        type Props = PropsA & {
          baz: string
        };

        class Bar extends React.Component {
          props: Props & PropsB;

          render() {
            return <div>{this.props.bar} - {this.props.zap} - {this.props.baz}</div>
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 151 (7.37.5)',
      code: `
        type PropsA = { bar: string };
        type PropsB = { zap: string };
        type Props =  {
          baz: string
        } & PropsA;

        class Bar extends React.Component {
          props: Props & PropsB;

          render() {
            return <div>{this.props.bar} - {this.props.zap} - {this.props.baz}</div>
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 152 (7.37.5)',
      code: `
        type Props = { foo: string }
        function higherOrderComponent<Props>() {
          return class extends React.Component<Props> {
            render() {
              return <div>{this.props.foo}</div>
            }
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 153 (7.37.5)',
      skip: "The native TypeScript parser cannot parse this upstream flow syntax (',' expected.).",
      code: `
        function higherOrderComponent<P: { foo: string }>() {
          return class extends React.Component<P> {
            render() {
              return <div>{this.props.foo}</div>
            }
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 154 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream flow syntax (Identifier expected.).',
      code: `
        const withOverlayState = <P: { foo: string }>(WrappedComponent: ComponentType<P>): ComponentType<P> => (
          class extends React.Component<P> {
            constructor(props) {
              super(props);
              this.state = { foo: props.foo };
            }
            render() {
              return <div>Hello World</div>;
            }
          }
        )
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 155 (7.37.5)',
      code: `function Foo() {
      const props = {};
      props.bar = 'bar';
      return <div {...props} />;
    }`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 156 (7.37.5)',
      code: `function Foo(props) {
      props.bar = 'bar';
      return <div {...props} />;
    }`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 157 (7.37.5)',
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
      name: 'Upstream valid 158 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream flow syntax (Identifier expected.).',
      code: `
      // @flow
      import * as React from 'react';

      type Props = {};

      const func = <OP: *>(arg) => arg;

      const hoc = <OP>() => () => {
        class Inner extends React.Component<Props & OP> {
          render() {
            return <div />;
          }
        }
      };
    `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 159 (7.37.5)',
      code: `
        const Slider = props => (
          <RcSlider {...props} />
        );

        Slider.propTypes = RcSlider.propTypes;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 160 (7.37.5)',
      code: `
        const Slider = props => (
          <RcSlider foo={props.bar} />
        );

        Slider.propTypes = RcSlider.propTypes;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 161 (7.37.5)',
      code: `
        class Foo extends React.Component {
          bar() {
            this.setState((state, props) => ({ current: props.current }));
          }
          render() {
            return <div />;
          }
        }

        Foo.propTypes = {
          current: PropTypes.number.isRequired,
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 162 (7.37.5)',
      code: `
        class Foo extends React.Component {
          bar() {
            this.setState((state, props) => {
              function f(_, { aaaaaaa }) {}
            });
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 163 (7.37.5)',
      code: `
        class Foo extends React.Component {
          static getDerivedStateFromProps(props) {
            const { foo } = props;
            return {
              foobar: foo
            };
          }

          render() {
            const { foobar } = this.state;
            return <div>{foobar}</div>;
          }
        }

        Foo.propTypes = {
          foo: PropTypes.func.isRequired,
        };
      `,
      settings: { react: { version: '16.3.0' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 164 (7.37.5)',
      code: `
        const HeaderBalance = React.memo(({ cryptoCurrency }) => (
          <div className="header-balance">
            <div className="header-balance__balance">
              BTC
              {cryptoCurrency}
            </div>
          </div>
        ));
        HeaderBalance.propTypes = {
          cryptoCurrency: PropTypes.string
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 165 (7.37.5)',
      code: `
        import React, { memo } from 'react';
        const HeaderBalance = memo(({ cryptoCurrency }) => (
          <div className="header-balance">
            <div className="header-balance__balance">
              BTC
              {cryptoCurrency}
            </div>
          </div>
        ));
        HeaderBalance.propTypes = {
          cryptoCurrency: PropTypes.string
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 166 (7.37.5)',
      code: `
        import Foo, { memo } from 'foo';
        const HeaderBalance = memo(({ cryptoCurrency }) => (
          <div className="header-balance">
            <div className="header-balance__balance">
              BTC
              {cryptoCurrency}
            </div>
          </div>
        ));
        HeaderBalance.propTypes = {
          cryptoCurrency: PropTypes.string
        };
      `,
      settings: { react: { pragma: 'Foo' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 167 (7.37.5)',
      code: `
        const Label = React.forwardRef(({ text }, ref) => {
          return <div ref={ref}>{text}</div>;
        });
        Label.propTypes = {
          text: PropTypes.string,
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 168 (7.37.5)',
      code: `
        const Label = Foo.forwardRef(({ text }, ref) => {
          return <div ref={ref}>{text}</div>;
        });
        Label.propTypes = {
          text: PropTypes.string,
        };
      `,
      settings: { react: { pragma: 'Foo' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 169 (7.37.5)',
      code: `
        import React, { forwardRef } from 'react';
        const Label = forwardRef(({ text }, ref) => {
          return <div ref={ref}>{text}</div>;
        });
        Label.propTypes = {
          text: PropTypes.string,
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 170 (7.37.5)',
      code: `
        import Foo, { forwardRef } from 'foo';
        const Label = forwardRef(({ text }, ref) => {
          return <div ref={ref}>{text}</div>;
        });
        Label.propTypes = {
          text: PropTypes.string,
        };
      `,
      settings: { react: { pragma: 'Foo' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 171 (7.37.5)',
      code: `
      class Foo extends React.Component {
        propTypes = {
          actions: PropTypes.object.isRequired,
        };
        componentWillReceiveProps (nextProps) {
          this.props.actions.doSomething();
        }

        componentWillUnmount () {
          this.props.actions.doSomething();
        }

        render() {
          return <div>foo</div>;
        }
      }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 172 (7.37.5)',
      code: `
        class Foo extends React.Component {
          componentDidUpdate() { this.inputRef.focus(); }
          render() {
            return (
              <div>
                <input ref={(node) => { this.inputRef = node; }} />
              </div>
            )
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 173 (7.37.5)',
      code: `
        class Foo extends React.Component {
          componentDidUpdate(nextProps, nextState) {
            const {
                first_organization,
                second_organization,
            } = this.state;
            return true;
          }
          render() {
            return <div>hi</div>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 174 (7.37.5)',
      code: `
      class Foo extends React.Component {
        shouldComponentUpdate(nextProps) {
          if (this.props.search !== nextProps.search) {
            let query = nextProps.query;
            let result = nextProps.list.filter(item => {
              return (item.name.toLowerCase().includes(query.trim().toLowerCase()));
            });

            this.setState({ result });

            return true;
          }
        }
        render() {
          return <div>foo</div>;
        }
      }
      Foo.propTypes = {
        search: PropTypes.object,
        list: PropTypes.array,
        query: PropTypes.string,
      };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 175 (7.37.5)',
      code: `
        export default class LazyLoader extends Component {
          static propTypes = {
            children: PropTypes.node,
            load: PropTypes.any,
          };
          state = { mod: null };
          shouldComponentUpdate(prevProps) {
            assert(prevProps.load === this.props.load);
            return true;
          }
          load() {
            this.props.load(mod => {
              this.setState({
                mod: mod.default ? mod.default : mod
              });
            });
          }
          render() {
            if (this.state.mod !== null) {
              return this.props.children(this.state.mod);
            }
            this.load();
            return null;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 176 (7.37.5)',
      code: `
        import React from 'react';
        import PropTypes from 'prop-types';
        import { connect } from 'react-redux';

        class Foo extends React.Component {
          render() {
            return this.props.children;
          }
        }

        Foo.propTypes = {
          children: PropTypes.element.isRequired
        };

        export const Unconnected = Foo;
        export default connect(Foo);
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 177 (7.37.5)',
      skip: "The native TypeScript parser cannot parse this upstream bind operator syntax (',' expected.).",
      code: `
        const a = {};
        function fn1() {}
        const b = a::fn1();
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 178 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream flow syntax (Property or signature expected.).',
      code: `
        type UsedProps = {|
          usedProp: number,
        |};

        type UnusedProps = {|
          unusedProp: number,
        |};

        type Props = {| ...UsedProps, ...UnusedProps |};

        function MyComponent({ usedProp }: Props) {
          return <div>{usedProp}</div>;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 179 (7.37.5)',
      code: `
        const Hello = props => {
          const { notInProp } = dict[props.foo];
          return <div />
        }

        Hello.propTypes = {
          foo: PropTypes.number,
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 180 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream flow syntax (Property or signature expected.).',
      code: `
      type Props = {|
        firstname?: string
      |};

      function Hello({ firstname = 'John' }: Props = {}) {
        return <div>Hello {firstname}</div>
      }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 181 (7.37.5)',
      code: `
        for (const { result } of results) {}
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 182 (7.37.5)',
      code: `
        type Props = {
          user: {
            [string]: string
          }
        };

        export function Bar(props: Props) {
          return (
            <div>
              {props.user}
              {props.user.name}
            </div>
          );
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 183 (7.37.5)',
      code: `
        const Label = React.memo(React.forwardRef(({ text }, ref) => {
          return <div ref={ref}>{text}</div>;
        }));
        Label.propTypes = {
          text: PropTypes.string
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 184 (7.37.5)',
      code: `
        import React, { memo, forwardRef } from 'react';
        const Label = memo(forwardRef(({ text }, ref) => {
          return <div ref={ref}>{text}</div>;
        }));
        Label.propTypes = {
          text: PropTypes.string
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 185 (7.37.5)',
      code: `
        import Foo, { memo, forwardRef } from 'foo';
        const Label = memo(forwardRef(({ text }, ref) => {
          return <div ref={ref}>{text}</div>;
        }));
        Label.propTypes = {
          text: PropTypes.string
        };
      `,
      settings: { react: { pragma: 'Foo' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 186 (7.37.5)',
      code: `
        const Foo = ({ length, ordering }) => (
          length > 0 && (
            <Paginator items={ordering} pageSize={10} />
          )
        );
        Foo.propTypes = {
          length: PropTypes.number,
          ordering: PropTypes.array
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 187 (7.37.5)',
      code: `
      function noAComponent(props) {
        return <div>{props.text}</div>
      }
    `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 188 (7.37.5)',
      code: `
      export default function() {}
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 189 (7.37.5)',
      code: `
        function Component(props) {
          return 0,
          <div>
            Hello, { props.name }!
          </div>
        }

        Component.propTypes = {
          name: PropTypes.string.isRequired
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 190 (7.37.5)',
      code: `
      const SideMenu = styled(
        ({ componentId }) => (
          <S.Container>
            <S.Head />
            <UserInfo />
            <Separator />
            <MenuList componentId={componentId} />
          </S.Container>
        ),
      );
      SideMenu.propTypes = {
        componentId: PropTypes.string.isRequired
      }
      `,
      settings: { componentWrapperFunctions: [{ property: 'styled' }] },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 191 (7.37.5)',
      code: `
        interface Props {
          'aria-label': string // 'undefined' PropType is defined but prop is never used eslint(react/no-unused-prop-types)
          // 'undefined' PropType is defined but prop is never used eslint(react-redux/no-unused-prop-types)
        }

        export default function Component({
          'aria-label': ariaLabel, // 'aria-label' is missing in props validation eslint(react/prop-types)
        }: Props): JSX.Element {
          return <div aria-label={ariaLabel} />
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 192 (7.37.5)',
      code: `
        interface Props {
          value?: string;
        }

        // without the | null, all ok, with it, it is broken
        function Test ({ value }: Props): React.ReactElement<Props> | null {
          if (!value) {
            return null;
          }

          return <div>{value}</div>;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 193 (7.37.5)',
      code: `
        interface Props {
          value?: string;
        }

        // without the | null, all ok, with it, it is broken
        function Test ({ value }: Props): React.ReactElement<Props> | null {
          if (!value) {
            return <div>{value}</div>;;
          }

          return null;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 194 (7.37.5)',
      code: `
        interface Props {
          value?: string;
        }
        const Hello = (props: Props) => {
          if (props.value) {
            return <div></div>;
          }
          return null;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 195 (7.37.5)',
      code: `
      const MyComponent = (props) => {
        const render = () => {
          return <test>{props.hello}</test>;
        }
        return render();
      };
      MyComponent.propTypes = {
        hello: PropTypes.string.isRequired,
      };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 196 (7.37.5)',
      code: `
        interface Props {
          value?: string;
        }

        // without the | null, all ok, with it, it is broken
        function Test ({ value }: Props): React.ReactElement<Props> | null {
          if (!value) {
            return null;
          }
          return <div>{value}</div>;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 197 (7.37.5)',
      code: `
        import * as React from 'react';

        interface Props {
          text: string;
        }

        export const Test: React.FC<Props> = (props: Props) => {
          const createElement = (text: string) => {
            return (
              <div>
                {text}
              </div>
            );
          };

          return <>{createElement(props.text)}</>;
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 198 (7.37.5)',
      code: `
        const mapStateToProps = state => ({
          books: state.books
        });

        interface InfoLibTableProps extends ReturnType<typeof mapStateToProps> {
        }

        const App = (props: InfoLibTableProps) => {
          props.books();
          return <div></div>;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 199 (7.37.5)',
      code: `
        const mapStateToProps = state => ({
          books: state.books
        });

        interface BooksTable extends ReturnType<typeof mapStateToProps> {
          username: string;
        }

        const App = (props: BooksTable) => {
          props.books();
          return <div><span>{props.username}</span></div>;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 200 (7.37.5)',
      code: `
        interface infoLibTable {
          removeCollection(): Array<string>;
        }

        interface InfoLibTableProps extends ReturnType<(dispatch: storeDispatch) => infoLibTable> {
        }

        const App = (props: InfoLibTableProps) => {
          props.removeCollection();
          return <div></div>;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 201 (7.37.5)',
      code: `
        interface addTable {
          createCollection: () => Array<string>
        }

        type infoLibTable = addTable & {
          removeCollection: () => Array<string>
        }

        interface InfoLibTableProps extends ReturnType<(dispatch: storeDispatch) => infoLibTable> {
        }

        const App = (props: InfoLibTableProps) => {
          props.createCollection();
          props.removeCollection();
          return <div></div>;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 202 (7.37.5)',
      code: `
        interface InfoLibTableProps extends ReturnType<(dispatch: storeDispatch) => {
          removeCollection:  () => Array<string>,
        }> {
        }

        const App = (props: InfoLibTableProps) => {
          props.removeCollection();
          return <div></div>;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 203 (7.37.5)',
      code: `
        interface addTable {
          createCollection: () => Array<string>;
        }

        type infoLibTable = {
          removeCollection: () => Array<string>,
        };

        interface InfoLibTableProps extends  ReturnType<
        (dispatch: storeDispatch) => infoLibTable & addTable,
        >{}

        const App = (props: InfoLibTableProps) => {
          props.createCollection();
          props.removeCollection();
          return <div></div>;
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 204 (7.37.5)',
      code: `
        interface addTable {
          createCollection: () => Array<string>;
        }

        type infoLibTable = ReturnType<(dispatch: storeDispatch) => infoLibTable & addTable> & {
          removeCollection: () => Array<string>,
        };

        interface InfoLibTableProps {}

        const App = (props: infoLibTable) => {
          props.createCollection();
          props.removeCollection();
          return <div></div>;
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 205 (7.37.5)',
      code: `
        interface InfoLibTableProps extends ReturnType<(dispatch: storeDispatch) => ({
          removeCollection:  () => Array<string>,
        })> {
        }

        const App = (props: InfoLibTableProps) => {
          props.removeCollection();
          return <div></div>;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 206 (7.37.5)',
      code: `
        interface ThingProps extends React.HTMLAttributes<HTMLDivElement> {
          thing?: number
        }

        export const Thing = ({ thing = 1, style, ...props }: ThingProps) => {
          return <div />;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 207 (7.37.5)',
      code: `
        interface ThingProps {
          thing?: number
        }

        export const Thing = ({ thing = 1, style, ...props }: ThingProps & React.HTMLAttributes<HTMLDivElement>) => {
          return <div />;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 208 (7.37.5)',
      code: `
        type User = {
          user: string;
        }

        type Props = User & UserProps;

        export default (props: Props) => {
          const { userId, user } = props;

          if (userId === 0) {
            return <p>userId is 0</p>;
          }

          return null;
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 209 (7.37.5)',
      code: `
        type User = {
        }

        type Props = User & UserProps;

        export default (props: Props) => {
          const { user } = props;

          if (user === 0) {
            return <p>user is 0</p>;
          }

          return null;
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 210 (7.37.5)',
      code: `
        interface GenericProps {
          onClose: () => void
        }

        interface ImplementationProps extends GenericProps {
          onClick: () => void
        }

        export const Implementation: FC<ImplementationProps> = (
          {
            onClick,
            onClose,
          }: ImplementationProps
        ) => (<div />)
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 211 (7.37.5)',
      code: `
        interface Props extends V2SocialLoginProps {
          autoLoad: boolean;
        }

        export const DKFacebookButton = ({
          autoLoad,
          authAction,
          errorHandler,
          redirectUrl,
          isSignup,
        }: Props): JSX.Element | null => {
          if (!APP_ID) {
            rollbar.error('Missing Facebook OAuth App Id');
            return null;
          }

          const fbButtonText = isSignup ? 'Sign up with Facebook' : 'Log in with Facebook';

          const responseCallback = async ({
            accessToken,
            email = '',
            name = '',
          }: ReactFacebookLoginInfo) => {
            const [firstName, lastName] = name.split(' ');

            const requestData: DK.SocialLogin = {
              accessToken,
              email,
              firstName,
              lastName,
              intercomUserId: intercomService.getVisitorId(),
            };

            try {
              await authAction(requestData, redirectUrl);
            } catch (err) {
              errorHandler(err.message);
            }
          };

          const FacebookIcon = () => (
            <img
              style={{ marginRight: '8px' }}
              src={facebookIcon}
              alt='Facebook Login'
            />
          );

          return (
            <FacebookLogin
              cssClass='ant-btn dk-button dk-facebook-button dk-button--secondary ant-btn-primary ant-btn-lg'
              autoLoad={autoLoad}
              textButton={fbButtonText}
              size='small'
              icon={<FacebookIcon />}
              appId={APP_ID}
              fields='name,email'
              callback={responseCallback}
              data-testId='dk-facebook-button'
            />
          );
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 212 (7.37.5)',
      code: `
        function Foo({ bar = "" }: { bar: string }): JSX.Element {
          return <div>{bar}</div>;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 213 (7.37.5)',
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
        ...bindActionCreators<{ prop1: ()=>void, prop2: ()=>string }>(
          { prop1: importedAction, prop2: anotherImportedAction },
          dispatch,
        ),
      })

      const mapStateToProps = (state: State) => ({
        prop3: Selector.value(state),
      })

      type StateProps = ReturnType<typeof mapStateToProps>
      type DispatchProps = ReturnType<typeof mapDispatchToProps>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 214 (7.37.5)',
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
      type DispatchProps = ReturnType<typeof mapDispatchToProps>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 215 (7.37.5)',
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
        bindActionCreators<{ prop1: () => void, prop2: () => string }>(
          { prop1: importedAction, prop2: anotherImportedAction },
          dispatch,
        )

      const mapStateToProps = (state: State) => ({
        prop3: Selector.value(state),
      })

      type StateProps = ReturnType<typeof mapStateToProps>
      type DispatchProps = ReturnType<typeof mapDispatchToProps>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 216 (7.37.5)',
      code: `
      import React from 'react'

      interface Meta {
        touched: boolean,
        error: string;
      }

      interface Props {
        input: string,
        meta: Meta,
        cssClasses: object
      }
      const InputField = ({ input, meta: { touched, error }, cssClasses = {}, ...restProps }: Props) => {
        restProps.className = cssClasses.base

        if (cssClasses.custom) {
          restProps.className += 'cssClasses.custom'
        }
        if (touched && error) {
          restProps.className += 'cssClasses.error'
        }

        return(
          <input
            {...input}
            {...restProps}
          />
        )
      }
      export default InputField`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 217 (7.37.5)',
      code: `
        import React from 'react'

        type ComponentProps = {
          name: string
        }

        class Factory {
          getComponent() {
            return function Component({ name }: ComponentProps) {
              return <div>Hello {name}</div>
            }
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 218 (7.37.5)',
      code: `
        import React from 'react';

        interface PersonProps {
            username: string;
        }
        const Person: React.FunctionComponent<PersonProps> = (props): React.ReactElement => (
            <div>{props.username}</div>
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 219 (7.37.5)',
      code: `
        import React from 'react';

        const Person: React.FunctionComponent<PersonProps> = (props): React.ReactElement => (
            <div>{props.username}</div>
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 220 (7.37.5)',
      code: `
        import React from 'react';

        interface PersonProps {
            username: string;
        }
        const Person: React.FC<PersonProps> = (props): React.ReactElement => (
            <div>{props.username}</div>
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 221 (7.37.5)',
      code: `
        import React from 'react';

        export interface PersonProps {
            username: string;
        }
        const Person: React.FC<PersonProps> = (props): React.ReactElement => (
            <div>{props.username}</div>
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 222 (7.37.5)',
      code: `
        import React from 'react';
        const Person: React.FunctionComponent<{ username: string }> = (props): React.ReactElement => (
            <div>{props.username}</div>
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 223 (7.37.5)',
      code: `
        import React from 'react';
        type PersonProps = {
            username: string;
        }
        const Person: React.FunctionComponent<PersonProps> = (props): React.ReactElement => (
            <div>{props.username}</div>
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 224 (7.37.5)',
      code: `
        import { FunctionComponent } from 'react';

        type PersonProps = {
            username: string;
        }
        const Person: FunctionComponent<PersonProps> = (props): React.ReactElement => (
            <div>{props.username}</div>
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 225 (7.37.5)',
      code: `
        import { FC } from 'react';
        type PersonProps = {
            username: string;
        }
        const Person: FC<PersonProps> = (props): React.ReactElement => (
            <div>{props.username}</div>
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 226 (7.37.5)',
      code: `
        import type { FC } from 'react';
        type PersonProps = {
            username: string;
        }
        const Person: FC<PersonProps> = (props): React.ReactElement => (
            <div>{props.username}</div>
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 227 (7.37.5)',
      code: `
        import { FC as X } from 'react';
        interface PersonProps {
            username: string;
        }
        const Person: X<PersonProps> = (props): React.ReactElement => (
            <div>{props.username}</div>
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 228 (7.37.5)',
      code: `
        import * as X from 'react';
        interface PersonProps {
            username: string;
        }
        const Person: X.FC<PersonProps> = (props): React.ReactElement => (
            <div>{props.username}</div>
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 229 (7.37.5)',
      code: `
        import React from 'react';

        interface Props {
          item: any;
        }

        const SomeComponent: React.FC<Props> = ({ item }: Props) => {
          return item ? <></> : <></>;
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 230 (7.37.5)',
      code: `
        import React from "react";

        const returnTypeProp = (someProp: any) => ({ someProp });

        const SomeComponent: React.FunctionComponent<
          ReturnType<typeof returnTypeProp>
        > = ({ someProp }) => {
          return <div>{someProp}</div>;
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 231 (7.37.5)',
      code: `
        export const EuiSuperSelectControl: <T extends string>(
          props: EuiSuperSelectControlProps<T>
        ) => ReturnType<FunctionComponent<EuiSuperSelectControlProps<T>>> = ({
          ...rest
        }) => {
          return null;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 232 (7.37.5)',
      code: `
        import React from 'react';

        const MyComponent = (props: React.PropsWithChildren<{ username: string }>): React.ReactElement => {
          return <>{props.children}{props.username}</>;
        };

      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 233 (7.37.5)',
      code: `
        type Props<ValueType> = {
          value: ValueType;
          onClick: (value: ValueType) => void;
        };

        const Button = <T,>({ onClick, value }: Props<T>) => {
          return <button type="button" onClick={() => onClick(value)}>BUTTON</button>;
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 234 (7.37.5)',
      code: `
          import React, { VoidFunctionComponent } from 'react'

          interface Props {
          age: number
          }
          const Hello: VoidFunctionComponent<Props> = function Hello(props) {
          const { age } = props;

          return <div>Hello {age}</div>;
          }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 235 (7.37.5)',
      code: `
          import React from 'react'

          interface Props {
          age: number
          }
          const Hello: React.VoidFunctionComponent<Props> = function Hello(props) {
          const { age } = props;

          return <div>Hello {age}</div>;
          }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 236 (7.37.5)',
      code: `
          import React from 'react'

          export interface Props {
          age: number
          }
          const Hello: React.VoidFunctionComponent<Props> = function Hello(props) {
          const { age } = props;

          return <div>Hello {age}</div>;
          }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 237 (7.37.5)',
      code: `
          import React, { VFC } from 'react'

          interface Props {
          age: number
          }
          const Hello: VFC<Props> = function Hello(props) {
          const { age } = props;

          return <div>Hello {age}</div>;
          }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 238 (7.37.5)',
      code: `
          import React from 'react'

          interface Props {
          age: number
          }
          const Hello: React.VFC<Props> = function Hello(props) {
          const { age } = props;

          return <div>Hello {age}</div>;
          }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 239 (7.37.5)',
      code: `
          import React from 'react'

          export interface Props {
          age: number
          }
          const Hello: React.VFC<Props> = function Hello(props) {
          const { age } = props;

          return <div>Hello {age}</div>;
          }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 240 (7.37.5)',
      code: `
        import React, { ForwardRefRenderFunction  as X } from 'react'

        type IfooProps = { e: string };
          const Foo: X<HTMLDivElement, IfooProps> = function Foo (props, ref) {
          const { e } = props;
          return  <div ref={ref}>hello</div>;
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 241 (7.37.5)',
      code: `
        import React, { ForwardRefRenderFunction  as X } from 'react'

        export type IfooProps = { e: string };
          const Foo: X<HTMLDivElement, IfooProps> = function Foo (props, ref) {
          const { e } = props;
          return  <div ref={ref}>hello</div>;
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 242 (7.37.5)',
      code: `
        import React, { ForwardRefRenderFunction } from 'react'

        type IfooProps = { e: string };
          const Foo: ForwardRefRenderFunction<HTMLDivElement, IfooProps> = function Foo (props, ref) {
          const { e } = props;
          return  <div ref={ref}>hello</div>;
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 243 (7.37.5)',
      code: `
        import React, { ForwardRefRenderFunction } from 'react'

        type IfooProps = { e: string };
          const Foo: ForwardRefRenderFunction<HTMLDivElement, IfooProps> = (props, ref) => {
          const { e } = props;
          return  <div ref={ref}>hello</div>;
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 244 (7.37.5)',
      code: `
        import React from 'react'

        type IfooProps = { e: string };
        const Foo= React.forwardRef<HTMLDivElement, IfooProps>((props, ref) => {
          const { e } = props;
          return  <div ref={ref}>hello</div>;
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 245 (7.37.5)',
      code: `
        import React, { forwardRef } from 'react'

        type IfooProps = { e: string };
        const Foo= forwardRef<HTMLDivElement, IfooProps>((props, ref) => {
          const { e } = props;
          return  <div ref={ref}>hello</div>;
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 246 (7.37.5)',
      code: `
        import React from 'react'
        type IfooProps = { e: string };
        const Foo= React.forwardRef<HTMLDivElement, IfooProps>(function Foo(props, ref) {
          const { e } = props;
          return  <div ref={ref}>hello</div>;
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 247 (7.37.5)',
      code: `
        import React from 'react'
        interface IfooProps { e: string }
        const Foo= React.forwardRef<HTMLDivElement, IfooProps>(function Foo(props, ref) {
          const { e } = props;
          return  <div ref={ref}>hello</div>;
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 248 (7.37.5)',
      code: `
        import React, { forwardRef } from 'react'
        interface IfooProps { e: string }
        const Foo= forwardRef<HTMLDivElement, IfooProps>(function Foo(props, ref) {
          const { e } = props;
          return  <div ref={ref}>hello</div>;
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 249 (7.37.5)',
      code: `
        import React, { forwardRef as X } from 'react'

        type IfooProps = { e: string };
        const Foo= X<HTMLDivElement, IfooProps>((props, ref) => {
          const { e } = props;
          return  <div ref={ref}>hello</div>;
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 250 (7.37.5)',
      code: `
        import React from 'react'

        class Factory {
          getRenderFunction() {
            return function renderFunction({ name }) {
              return <div>Hello {name}</div>
            }
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 251 (7.37.5)',
      code: `
        class Test extends React.Component {
          static propTypes = {
            value: PropTypes.string
          };

          render() {
            return <span>{this.props.value}</span>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 252 (7.37.5)',
      code: `
        const styles = { width: 0 };

        function Foo(props) {
          const { styles: x } = props;
          return (
            <p>
              {styles.width} {x._}
            </p>
          );
        }

        Foo.propTypes = {
          styles: PropTypes.shape({
            _: PropTypes.number,
          }),
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 253 (7.37.5)',
      code: `
        import React, { forwardRef } from "react";

        export type Props = { children: React.ReactNode; type: "submit" | "button" };

        export const FancyButton = forwardRef<HTMLButtonElement, Props>((props, ref) => (
          <button type="button" ref={ref} className="MyClassName" type={props.type}>
            {props.children}
          </button>
        ));
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 254 (7.37.5)',
      code: `
        import React, { forwardRef } from "react";

        export type X = { num: number };
        export type Props = { children: React.ReactNode; type: "submit" | "button" } & X;

        export const FancyButton = forwardRef<HTMLButtonElement, Props>((props, ref) => (
          <button type="button" ref={ref} className="MyClassName" type={props.type} num={props.num}>
            {props.children}
          </button>
        ));
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 255 (7.37.5)',
      code: `
        import React, { forwardRef } from "react";

        export interface IProps {
          children: React.ReactNode;
          type: "submit" | "button"
        }

        export const FancyButton = forwardRef<HTMLButtonElement, IProps>((props, ref) => (
          <button type="button" ref={ref} className="MyClassName" type={props.type}>
            {props.children}
          </button>
        ));
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 256 (7.37.5)',
      code: `
        import React, { forwardRef } from "react";

        export interface X {
          num: number
        }
        export interface IProps extends X {
          children: React.ReactNode;
          type: "submit" | "button"
        }

        export const FancyButton = forwardRef<HTMLButtonElement, IProps>((props, ref) => (
          <button type="button" ref={ref} className="MyClassName" type={props.type} num={props.num}>
            {props.children}
          </button>
        ));
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 257 (7.37.5)',
      code: `
        type ButtonProps = ({
            children: React.Node,
        } | {
            icon: React.Element,
        }) & {
            label: string,
            disabled?: boolean
        } & {};

        const Button: React.FC<ButtonProps> = ({ children, icon, label, disabled }: ButtonProps) => {
          return <div />;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 258 (7.37.5)',
      code: `
        import React, { PropTypes } from 'react'

        function Foo({ bar }) {
          const { baz } = Foo;
          return <div>{baz} {bar}</div>
        }

        Foo.propTypes = {
          bar: PropTypes.string.isRequired,
        };

        Foo.baz = 'hi';
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 259 (7.37.5)',
      code: `
        type Props = {
          'data-hover': string
        }

        function MyComponent({
          'data-hover': dataHover
        }: Props) {
          return <span data-hover={dataHover} />
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 260 (7.37.5)',
      code: `
        import React, { PropTypes } from 'react';

        function MyComponent({ name }) {
          const someParams = {
            userName: name,
            greeting: 'Hi there',
          };

          function getMessage({ userName, greeting }) {
            return <span>\`\${greeting} \${userName}\`</span>;
          }

          return (
            <div>
              { getMessage(someParams) }
            </div>
          );
        }

        MyComponent.propTypes = {
          name: PropTypes.string.isRequired,
        };

        export default MyComponent;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 261 (7.37.5)',
      skip: "The native TypeScript parser cannot parse this upstream flow syntax (';' expected.).",
      code: `
        type Field = {
            value: string,
            error: string,
        }

        type Form = {
            fields: {
                [string]: Field,
            },
            formError: string,
        }

        type Props = {
            bankDetails: Form,
            onBankDetailsUpdate: any => void,
        }

        const Provider = (props:Props) =>
                    <Input
                        label={'Account Name'}
                        value={props.bankDetails.fields.accountName.value}
                        error={props.bankDetails.fields.accountName.error}
                        onChange={newVal => props.onBankDetailsUpdate({ accountName: newVal })}
                    />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 262 (7.37.5)',
      code: `
        class App extends React.Component {
          render() {
            return (
                <h1>Open the console</h1>
            )
          }
        }

        // does not work
        App.propTypes = PropTypes.exact({
          foo: PropTypes.object,
        });
      `,
      settings: {
        propWrapperFunctions: [
          { object: 'PropTypes', property: 'exact', exact: true },
        ],
      },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 263 (7.37.5)',
      code: `
        class Test extends React.Component {
          componentDidUpdate() {
              const { bar } = this.state;
              console.log(bar);
          }

          render() {
              return null;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 264 (7.37.5)',
      code: `
        const DisplayName = (props) => {
            const getNameDiv = () => {
                return <div>{props.name}</div>;
            };

            return getNameDiv();
        };

        DisplayName.propTypes = {
            name: PropTypes.string.isRequired,
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 265 (7.37.5)',
      code: `
        const SomeComponent = (props) => {
          function renderComponent() {
            return <div>{props.name}</div>
          }

          return renderComponent();
        }

        SomeComponent.propTypes = {
          name: PropTypes.string
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 266 (7.37.5)',
      code: `
        import React from 'react';
        import { MyType } from './types';

        function Component(props: Props): JSX.Element | null {
          const { array } = props;

          function renderFound(): JSX.Element | null {
            const found = array.find(x => x.id === id); // issue here

            if (!found) {
              return null;
            }

            return (
              <div>{found.id}</div>
            );
          }

          return renderFound();
        }

        interface Props {
          array: MyType[];
        }

        export default Component;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 267 (7.37.5)',
      code: `
        export default React.memo(function AppIcon({
          icon,
          className,
          title,
          spinning,
        }: {
          icon: string | IconDefinition;
          className?: string;
          title?: string;
          spinning?: boolean;
        }) {
          if (typeof icon === 'string') {
            return (
              <span
                className={clsx(
                  icon,
                  'app-icon',
                  'no-pointer-events',
                  className,
                  spinning ? 'fa-spin' : false
                )}
                title={title}
              />
            );
          } else {
            return (
              <FontAwesomeIcon
                className={className ? 'app-icon ' + className : 'app-icon'}
                icon={icon}
                title={title}
                spin={spinning}
              />
            );
          }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 268 (7.37.5)',
      code: `
        import PropTypes from 'prop-types';
        import React from 'react';

        import { Link } from '..';

        const LinkWrapper = ({ primaryLinks }) => (
          <>
            {primaryLinks.map((x, index) => (
              <Link key={index} href={x.link}>
                {x.text}
              </Link>
            ))}
          </>
        );

        LinkWrapper.propTypes = {
          primaryLinks: PropTypes.arrayOf(
            PropTypes.shape({
              text: PropTypes.string,
            })
          ),
        };

        export default LinkWrapper;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 269 (7.37.5)',
      code: `
        const projectType = PropTypes.shape({
          id: PropTypes.string.isRequired,
          name: PropTypes.string.isRequired,
        });

        const nodesType = PropTypes.arrayOf(
          PropTypes.shape({ project: projectType, children: nodesType }),
        );

        class ProjectNode extends React.Component {
          render() {
            return <div />;
          }
        }

        ProjectNode.propTypes = {
          project: projectType.isRequired,
          nodes: nodesType,
          depth: PropTypes.number.isRequired,
          setActiveProject: PropTypes.func.isRequired,
          activeProject: projectType,
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 270 (7.37.5)',
      code: `
        import React from 'react';

        interface SomeType<ContextType = any> {
          renderValue: (context: ContextType) => React.ReactNode;
        }

        interface DataObject {
          id: string,
          title: string,
          value: string,
        }

        const someType: SomeType<DataObject> = {
          renderValue: ({ title }) => <div>{title}</div>,
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 271 (7.37.5)',
      code: `
        <MyComponent
          options={{
            tabBarIcon: ({ color, size }) => (
              <Icon name="dumbbell" size={size} color={color} />
            )
          }}
        />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 272 (7.37.5)',
      code: `
        import React, { forwardRef } from 'react';
        import { ControlProps, NamedProps } from './ext';

        type ButtonProps = ControlProps & NamedProps & {
          onClick?: (() => void) | undefined;
          onMouseDown?: (() => void) | undefined;
          onMouseUp?: (() => void) | undefined;
          disabled?: boolean | undefined;
          width?: number;
          type?: 'submit' | 'reset' | 'button' | undefined;
        };

        const BaseButton = forwardRef<HTMLButtonElement, ButtonProps>((
          {
            name,
            className,
            onClick,
            onMouseDown,
            onMouseUp,
            children,
            disabled,
            width,
            type,
          },
          ref,
        ): JSX.Element => {
          return <span>{width}</span>;
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 273 (7.37.5)',
      code: `
        import React, { forwardRef } from 'react';
        import { ControlProps, NamedProps } from './ext';

        interface ButtonProps extends NamedProps {
          onClick?: (() => void) | undefined;
          onMouseDown?: (() => void) | undefined;
          onMouseUp?: (() => void) | undefined;
          disabled?: boolean | undefined;
          width?: number;
          type?: 'submit' | 'reset' | 'button' | undefined;
        };

        const BaseButton = forwardRef<HTMLButtonElement, ButtonProps>((
          {
            name,
            className,
            onClick,
            onMouseDown,
            onMouseUp,
            children,
            disabled,
            width,
            type,
          },
          ref,
        ): JSX.Element => {
          return <span>{width}</span>;
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 274 (7.37.5)',
      code: `
        import { ComponentPropsWithoutRef, forwardRef } from 'react';

        export const FancyButton = forwardRef<HTMLButtonElement, ComponentPropsWithoutRef<"button">>(
          ({ className, children, ...props }, ref) => (
            <button type="button" ref={ref} className={className} {...props}>
              {children}
            </button>
          ),
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 275 (7.37.5)',
      code: `
        import { ComponentPropsWithRef, forwardRef } from "react";

        export const FancyButton = forwardRef<HTMLButtonElement, ComponentPropsWithRef<"button">>(
          ({ className, children, ...props }, ref) => (
            <button type="button" ref={ref} className={className} {...props}>
              {children}
            </button>
          ),
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 276 (7.37.5)',
      code: `
        import { ComponentProps, forwardRef } from "react";

        export const FancyButton = forwardRef<HTMLButtonElement, ComponentProps<"button">>(
          ({ className, children, ...props }, ref) => (
            <button type="button" ref={ref} className={className} {...props}>
              {children}
            </button>
          ),
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 277 (7.37.5)',
      code: `
        import { ComponentPropsWithoutRef, ElementRef, forwardRef } from "react";

        const BaseButton = forwardRef<HTMLButtonElement, ComponentPropsWithoutRef<"button">>(
          ({ children, className, ...props }, ref) => (
            <button type="button" ref={ref} className={className} {...props}>
                {children}
            </button>
          ),
        );

        export const FancyButton = forwardRef<ElementRef<typeof BaseButton>, ComponentPropsWithoutRef<typeof BaseButton>>(
          ({ children, className, ...props }, ref) => (
            <BaseButton ref={ref} className={className} {...props}>
              {children}
            </BaseButton>
          ),
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 278 (7.37.5)',
      code: `
        import { ComponentPropsWithRef, ElementRef, forwardRef } from "react";

        const BaseButton = forwardRef<HTMLButtonElement, ComponentPropsWithRef<"button">>(
          ({ children, className, ...props }, ref) => (
            <button type="button" ref={ref} className={className} {...props}>
                {children}
            </button>
          ),
        );

        export const FancyButton = forwardRef<ElementRef<typeof BaseButton>, ComponentPropsWithRef<typeof BaseButton>>(
          ({ children, className, ...props }, ref) => (
            <BaseButton ref={ref} className={className} {...props}>
              {children}
            </BaseButton>
          ),
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 279 (7.37.5)',
      code: `
        import { ComponentProps, ElementRef, forwardRef } from "react";

        const BaseButton = forwardRef<HTMLButtonElement, ComponentProps<"button">>(
          ({ children, className, ...props }, ref) => (
            <button type="button" ref={ref} className={className} {...props}>
                {children}
            </button>
          ),
        );

        export const FancyButton = forwardRef<ElementRef<typeof BaseButton>, ComponentProps<typeof BaseButton>>(
          ({ children, className, ...props }, ref) => (
            <BaseButton ref={ref} className={className} {...props}>
              {children}
            </BaseButton>
          ),
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 280 (7.37.5)',
      code: `
        import { ComponentProps, ComponentPropsWithoutRef, ElementRef, forwardRef } from "react";

        const BaseButton = forwardRef<HTMLButtonElement, ComponentProps<"button">>(
          ({ children, className, ...props }, ref) => (
            <button type="button" ref={ref} className={className} {...props}>
                {children}
            </button>
          ),
        );

        export const FancyButton = forwardRef<ElementRef<typeof BaseButton>, ComponentPropsWithoutRef<typeof BaseButton>>(
          ({ children, className, ...props }, ref) => (
            <BaseButton ref={ref} className={className} {...props}>
              {children}
            </BaseButton>
          ),
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 281 (7.37.5)',
      code: `
        import { ComponentProps, ComponentPropsWithoutRef, ElementRef, forwardRef } from "react";

        const BaseButton = forwardRef<HTMLButtonElement, ComponentPropsWithoutRef<"button">>(
          ({ children, className, ...props }, ref) => (
            <button type="button" ref={ref} className={className} {...props}>
                {children}
            </button>
          ),
        );

        export const FancyButton = forwardRef<ElementRef<typeof BaseButton>, ComponentProps<typeof BaseButton>>(
          ({ children, className, ...props }, ref) => (
            <BaseButton ref={ref} className={className} {...props}>
              {children}
            </BaseButton>
          ),
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 282 (7.37.5)',
      code: `
        import { ComponentProps, ComponentPropsWithRef, ElementRef, forwardRef } from "react";

        const BaseButton = forwardRef<HTMLButtonElement, ComponentProps<"button">>(
          ({ children, className, ...props }, ref) => (
            <button type="button" ref={ref} className={className} {...props}>
                {children}
            </button>
          ),
        );

        export const FancyButton = forwardRef<ElementRef<typeof BaseButton>, ComponentPropsWithRef<typeof BaseButton>>(
          ({ children, className, ...props }, ref) => (
            <BaseButton ref={ref} className={className} {...props}>
              {children}
            </BaseButton>
          ),
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 283 (7.37.5)',
      code: `
        import { ComponentProps, ComponentPropsWithRef, ElementRef, forwardRef } from 'react';

        const BaseButton = forwardRef<HTMLButtonElement, ComponentPropsWithRef<"button">>(
          ({ children, className, ...props }, ref) => (
            <button type="button" ref={ref} className={className} {...props}>
                {children}
            </button>
          ),
        );

        export const FancyButton = forwardRef<ElementRef<typeof BaseButton>, ComponentProps<typeof BaseButton>>(
          ({ children, className, ...props }, ref) => (
            <BaseButton ref={ref} className={className} {...props}>
              {children}
            </BaseButton>
          ),
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 284 (7.37.5)',
      code: `
        import React from "react";

        export const FancyButton = React.forwardRef<HTMLButtonElement, React.ComponentProps<"button">>(
          ({ className, children, ...props }, ref) => (
            <button type="button" ref={ref} className={className} {...props}>
              {children}
            </button>
          ),
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 285 (7.37.5)',
      code: `
        import React from "react";

        export const FancyButton = React.forwardRef<HTMLButtonElement, React.ComponentPropsWithRef<"button">>(
          ({ className, children, ...props }, ref) => (
            <button type="button" ref={ref} className={className} {...props}>
              {children}
            </button>
          ),
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 286 (7.37.5)',
      code: `
        import React from "react";

        export const FancyButton = React.forwardRef<HTMLButtonElement, React.ComponentPropsWithoutRef<"button">>(
          ({ className, children, ...props }, ref) => (
            <button type="button" ref={ref} className={className} {...props}>
              {children}
            </button>
          ),
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 287 (7.37.5)',
      code: `
        import React from "react";

        const BaseButton = React.forwardRef<HTMLButtonElement, React.ComponentProps<"button">>(
          ({ children, className, ...props }, ref) => (
            <button type="button" ref={ref} className={className} {...props}>
                {children}
            </button>
          ),
        );

        export const FancyButton = React.forwardRef<React.ElementRef<typeof BaseButton>, React.ComponentProps<typeof BaseButton>>(
          ({ children, className, ...props }, ref) => (
            <BaseButton ref={ref} className={className} {...props}>
              {children}
            </BaseButton>
          ),
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 288 (7.37.5)',
      code: `
        import React from "react";

        const BaseButton = React.forwardRef<HTMLButtonElement, React.ComponentPropsWithRef<"button">>(
          ({ children, className, ...props }, ref) => (
            <button type="button" ref={ref} className={className} {...props}>
                {children}
            </button>
          ),
        );

        export const FancyButton = React.forwardRef<React.ElementRef<typeof BaseButton>, React.ComponentPropsWithRef<typeof BaseButton>>(
          ({ children, className, ...props }, ref) => (
            <BaseButton ref={ref} className={className} {...props}>
              {children}
            </BaseButton>
          ),
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 289 (7.37.5)',
      code: `
        import React from "react";

        const BaseButton = React.forwardRef<HTMLButtonElement, React.ComponentPropsWithoutRef<"button">>(
          ({ children, className, ...props }, ref) => (
            <button type="button" ref={ref} className={className} {...props}>
                {children}
            </button>
          ),
        );

        export const FancyButton = React.forwardRef<React.ElementRef<typeof BaseButton>, React.ComponentPropsWithoutRef<typeof BaseButton>>(
          ({ children, className, ...props }, ref) => (
            <BaseButton ref={ref} className={className} {...props}>
              {children}
            </BaseButton>
          ),
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 290 (7.37.5)',
      code: `
        import React from "react";

        const BaseButton = React.forwardRef<HTMLButtonElement, React.ComponentProps<"button">>(
          ({ children, className, ...props }, ref) => (
            <button type="button" ref={ref} className={className} {...props}>
                {children}
            </button>
          ),
        );

        export const FancyButton = React.forwardRef<React.ElementRef<typeof BaseButton>, React.ComponentPropsWithRef<typeof BaseButton>>(
          ({ children, className, ...props }, ref) => (
            <BaseButton ref={ref} className={className} {...props}>
              {children}
            </BaseButton>
          ),
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 291 (7.37.5)',
      code: `
        import React from "react";

        const BaseButton = React.forwardRef<HTMLButtonElement, React.ComponentPropsWithRef<"button">>(
          ({ children, className, ...props }, ref) => (
            <button type="button" ref={ref} className={className} {...props}>
                {children}
            </button>
          ),
        );

        export const FancyButton = React.forwardRef<React.ElementRef<typeof BaseButton>, React.ComponentProps<typeof BaseButton>>(
          ({ children, className, ...props }, ref) => (
            <BaseButton ref={ref} className={className} {...props}>
              {children}
            </BaseButton>
          ),
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 292 (7.37.5)',
      code: `
        import React from "react";

        const BaseButton = React.forwardRef<HTMLButtonElement, React.ComponentProps<"button">>(
          ({ children, className, ...props }, ref) => (
            <button type="button" ref={ref} className={className} {...props}>
                {children}
            </button>
          ),
        );

        export const FancyButton = React.forwardRef<React.ElementRef<typeof BaseButton>, React.ComponentPropsWithoutRef<typeof BaseButton>>(
          ({ children, className, ...props }, ref) => (
            <BaseButton ref={ref} className={className} {...props}>
              {children}
            </BaseButton>
          ),
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 293 (7.37.5)',
      code: `
        import React from "react";

        const BaseButton = React.forwardRef<HTMLButtonElement, React.ComponentPropsWithoutRef<"button">>(
          ({ children, className, ...props }, ref) => (
            <button type="button" ref={ref} className={className} {...props}>
                {children}
            </button>
          ),
        );

        export const FancyButton = React.forwardRef<React.ElementRef<typeof BaseButton>, React.ComponentProps<typeof BaseButton>>(
          ({ children, className, ...props }, ref) => (
            <BaseButton ref={ref} className={className} {...props}>
              {children}
            </BaseButton>
          ),
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 294 (7.37.5)',
      code: `
        import * as React from "react";

        export const FancyButton = React.forwardRef<HTMLButtonElement, React.ComponentProps<"button">>(
          ({ className, children, ...props }, ref) => (
            <button type="button" ref={ref} className={className} {...props}>
              {children}
            </button>
          ),
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 295 (7.37.5)',
      code: `
        import * as React from "react";

        export const FancyButton = React.forwardRef<HTMLButtonElement, React.ComponentPropsWithRef<"button">>(
          ({ className, children, ...props }, ref) => (
            <button type="button" ref={ref} className={className} {...props}>
              {children}
            </button>
          ),
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 296 (7.37.5)',
      code: `
        import * as React from "react";

        export const FancyButton = React.forwardRef<HTMLButtonElement, React.ComponentPropsWithoutRef<"button">>(
          ({ className, children, ...props }, ref) => (
            <button type="button" ref={ref} className={className} {...props}>
              {children}
            </button>
          ),
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 297 (7.37.5)',
      code: `
        import * as React from "react";

        const BaseButton = React.forwardRef<HTMLButtonElement, React.ComponentProps<"button">>(
          ({ children, className, ...props }, ref) => (
            <button type="button" ref={ref} className={className} {...props}>
                {children}
            </button>
          ),
        );

        export const FancyButton = React.forwardRef<React.ElementRef<typeof BaseButton>, React.ComponentProps<typeof BaseButton>>(
          ({ children, className, ...props }, ref) => (
            <BaseButton ref={ref} className={className} {...props}>
              {children}
            </BaseButton>
          ),
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 298 (7.37.5)',
      code: `
        import * as React from "react";

        const BaseButton = React.forwardRef<HTMLButtonElement, React.ComponentPropsWithRef<"button">>(
          ({ children, className, ...props }, ref) => (
            <button type="button" ref={ref} className={className} {...props}>
                {children}
            </button>
          ),
        );

        export const FancyButton = React.forwardRef<React.ElementRef<typeof BaseButton>, React.ComponentPropsWithRef<typeof BaseButton>>(
          ({ children, className, ...props }, ref) => (
            <BaseButton ref={ref} className={className} {...props}>
              {children}
            </BaseButton>
          ),
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 299 (7.37.5)',
      code: `
        import * as React from "react";

        const BaseButton = React.forwardRef<HTMLButtonElement, React.ComponentPropsWithoutRef<"button">>(
          ({ children, className, ...props }, ref) => (
            <button type="button" ref={ref} className={className} {...props}>
                {children}
            </button>
          ),
        );

        export const FancyButton = React.forwardRef<React.ElementRef<typeof BaseButton>, React.ComponentPropsWithoutRef<typeof BaseButton>>(
          ({ children, className, ...props }, ref) => (
            <BaseButton ref={ref} className={className} {...props}>
              {children}
            </BaseButton>
          ),
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 300 (7.37.5)',
      code: `
        import * as React from "react";

        const BaseButton = React.forwardRef<HTMLButtonElement, React.ComponentProps<"button">>(
          ({ children, className, ...props }, ref) => (
            <button type="button" ref={ref} className={className} {...props}>
                {children}
            </button>
          ),
        );

        export const FancyButton = React.forwardRef<React.ElementRef<typeof BaseButton>, React.ComponentPropsWithoutRef<typeof BaseButton>>(
          ({ children, className, ...props }, ref) => (
            <BaseButton ref={ref} className={className} {...props}>
              {children}
            </BaseButton>
          ),
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 301 (7.37.5)',
      code: `
        import * as React from "react";

        const BaseButton = React.forwardRef<HTMLButtonElement, React.ComponentPropsWithoutRef<"button">>(
          ({ children, className, ...props }, ref) => (
            <button type="button" ref={ref} className={className} {...props}>
                {children}
            </button>
          ),
        );

        export const FancyButton = React.forwardRef<React.ElementRef<typeof BaseButton>, React.ComponentProps<typeof BaseButton>>(
          ({ children, className, ...props }, ref) => (
            <BaseButton ref={ref} className={className} {...props}>
              {children}
            </BaseButton>
          ),
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 302 (7.37.5)',
      code: `
        import * as React from "react";

        const BaseButton = React.forwardRef<HTMLButtonElement, React.ComponentProps<"button">>(
          ({ children, className, ...props }, ref) => (
            <button type="button" ref={ref} className={className} {...props}>
                {children}
            </button>
          ),
        );

        export const FancyButton = React.forwardRef<React.ElementRef<typeof BaseButton>, React.ComponentPropsWithRef<typeof BaseButton>>(
          ({ children, className, ...props }, ref) => (
            <BaseButton ref={ref} className={className} {...props}>
              {children}
            </BaseButton>
          ),
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 303 (7.37.5)',
      code: `
        import * as React from "react";

        const BaseButton = React.forwardRef<HTMLButtonElement, React.ComponentPropsWithRef<"button">>(
          ({ children, className, ...props }, ref) => (
            <button type="button" ref={ref} className={className} {...props}>
                {children}
            </button>
          ),
        );

        export const FancyButton = React.forwardRef<React.ElementRef<typeof BaseButton>, React.ComponentProps<typeof BaseButton>>(
          ({ children, className, ...props }, ref) => (
            <BaseButton ref={ref} className={className} {...props}>
              {children}
            </BaseButton>
          ),
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 304 (7.37.5)',
      code: `
        import React, {ComponentPropsWithoutRef } from 'react';

        export const FancyButton = React.forwardRef<HTMLButtonElement, ComponentPropsWithoutRef<"button">>(
          ({ className, children, ...props }, ref) => (
            <button type="button" ref={ref} className={className} {...props}>
              {children}
            </button>
          ),
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 305 (7.37.5)',
      code: `
        import React, { ComponentPropsWithRef } from 'react';

        export const FancyButton = React.forwardRef<HTMLButtonElement, ComponentPropsWithRef<"button">>(
          ({ className, children, ...props }, ref) => (
            <button type="button" ref={ref} className={className} {...props}>
              {children}
            </button>
          ),
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 306 (7.37.5)',
      code: `
        import React, { ComponentProps } from 'react';

        export const FancyButton = React.forwardRef<HTMLButtonElement, ComponentProps<"button">>(
          ({ className, children, ...props }, ref) => (
            <button type="button" ref={ref} className={className} {...props}>
              {children}
            </button>
          ),
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 307 (7.37.5)',
      code: `
        import React, { ComponentPropsWithoutRef, ElementRef } from 'react';

        const BaseButton = React.forwardRef<HTMLButtonElement, ComponentPropsWithoutRef<"button">>(
          ({ children, className, ...props }, ref) => (
            <button type="button" ref={ref} className={className} {...props}>
                {children}
            </button>
          ),
        );

        export const FancyButton = React.forwardRef<ElementRef<typeof BaseButton>, ComponentPropsWithoutRef<typeof BaseButton>>(
          ({ children, className, ...props }, ref) => (
            <BaseButton ref={ref} className={className} {...props}>
              {children}
            </BaseButton>
          ),
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 308 (7.37.5)',
      code: `
        import React, { ComponentPropsWithRef, ElementRef } from 'react';

        const BaseButton = React.forwardRef<HTMLButtonElement, ComponentPropsWithRef<"button">>(
          ({ children, className, ...props }, ref) => (
            <button type="button" ref={ref} className={className} {...props}>
                {children}
            </button>
          ),
        );

        export const FancyButton = React.forwardRef<ElementRef<typeof BaseButton>, ComponentPropsWithRef<typeof BaseButton>>(
          ({ children, className, ...props }, ref) => (
            <BaseButton ref={ref} className={className} {...props}>
              {children}
            </BaseButton>
          ),
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 309 (7.37.5)',
      code: `
        import React, { ComponentProps, ElementRef } from 'react';

        const BaseButton = React.forwardRef<HTMLButtonElement, ComponentProps<"button">>(
          ({ children, className, ...props }, ref) => (
            <button type="button" ref={ref} className={className} {...props}>
                {children}
            </button>
          ),
        );

        export const FancyButton = React.forwardRef<ElementRef<typeof BaseButton>, ComponentProps<typeof BaseButton>>(
          ({ children, className, ...props }, ref) => (
            <BaseButton ref={ref} className={className} {...props}>
              {children}
            </BaseButton>
          ),
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 310 (7.37.5)',
      code: `
        import React, { ComponentProps, ComponentPropsWithoutRef, ElementRef } from 'react';

        const BaseButton = React.forwardRef<HTMLButtonElement, ComponentPropsWithoutRef<"button">>(
          ({ children, className, ...props }, ref) => (
            <button type="button" ref={ref} className={className} {...props}>
                {children}
            </button>
          ),
        );

        export const FancyButton = React.forwardRef<ElementRef<typeof BaseButton>, ComponentProps<typeof BaseButton>>(
          ({ children, className, ...props }, ref) => (
            <BaseButton ref={ref} className={className} {...props}>
              {children}
            </BaseButton>
          ),
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 311 (7.37.5)',
      code: `
        import React, { ComponentProps, ComponentPropsWithoutRef, ElementRef } from 'react';

        const BaseButton = React.forwardRef<HTMLButtonElement, ComponentProps<"button">>(
          ({ children, className, ...props }, ref) => (
            <button type="button" ref={ref} className={className} {...props}>
                {children}
            </button>
          ),
        );

        export const FancyButton = React.forwardRef<ElementRef<typeof BaseButton>, ComponentPropsWithoutRef<typeof BaseButton>>(
          ({ children, className, ...props }, ref) => (
            <BaseButton ref={ref} className={className} {...props}>
              {children}
            </BaseButton>
          ),
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 312 (7.37.5)',
      code: `
        import React, { ComponentProps, ComponentPropsWithRef, ElementRef } from 'react';

        const BaseButton = React.forwardRef<HTMLButtonElement, ComponentPropsWithRef<"button">>(
          ({ children, className, ...props }, ref) => (
            <button type="button" ref={ref} className={className} {...props}>
                {children}
            </button>
          ),
        );

        export const FancyButton = React.forwardRef<ElementRef<typeof BaseButton>, ComponentProps<typeof BaseButton>>(
          ({ children, className, ...props }, ref) => (
            <BaseButton ref={ref} className={className} {...props}>
              {children}
            </BaseButton>
          ),
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 313 (7.37.5)',
      code: `
        import React, { ComponentProps, ComponentPropsWithRef, ElementRef } from 'react';

        const BaseButton = React.forwardRef<HTMLButtonElement, ComponentProps<"button">>(
          ({ children, className, ...props }, ref) => (
            <button type="button" ref={ref} className={className} {...props}>
                {children}
            </button>
          ),
        );

        export const FancyButton = React.forwardRef<ElementRef<typeof BaseButton>, ComponentPropsWithRef<typeof BaseButton>>(
          ({ children, className, ...props }, ref) => (
            <BaseButton ref={ref} className={className} {...props}>
              {children}
            </BaseButton>
          ),
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 314 (7.37.5)',
      code: `
        import React, { forwardRef } from 'react';
        import { IExt1 } from './ext';

        interface IProps extends IExt1 {
          onClick?: (() => void) | undefined;
          onMouseDown?: (() => void) | undefined;
          onMouseUp?: (() => void) | undefined;
          disabled?: boolean | undefined;
          width?: number;
          type?: 'submit' | 'reset' | 'button' | undefined;
        };

        const Button: React.FC<IProps> = ({
          name,
          className,
          onClick,
          onMouseDown,
          onMouseUp,
          children,
          disabled,
          width,
          type,
        }): JSX.Element => {
          return <span>{width}</span>;
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 315 (7.37.5)',
      code: `
        import React, { memo } from 'react';
        interface Props1 {
            age: number;
        }
        const HelloTemp = memo(({ age }: Props1) => {
            return <div>Hello {age}</div>;
        });
        export const Hello = HelloTemp
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 316 (7.37.5)',
      code: `
        import React, { forwardRef, memo } from 'react';
        interface Props1 {
            age: number;
        }
        const HelloTemp = forwardRef(({ age }: Props1) => {
            return <div>Hello {age}</div>;
        });
        export const Hello = memo(HelloTemp);
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 317 (7.37.5)',
      code: `
        import React, { forwardRef, memo } from 'react';
        interface Props1 {
            age: number;
        }
        export const Hello = memo(
            forwardRef(({ age }: Props1) => {
                return <div>Hello {age}</div>;
            }),
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 318 (7.37.5)',
      code: `
        import React from "react"

        export function Heading({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
          return <div className={cn("font-semibold text-lg", className)} {...props} />
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 319 (7.37.5)',
      code: `
        import React from 'react';
        type TDelIconProps = React.HTMLProps<HTMLDivElement>;

        const DelIcon: React.FC<TDelIconProps> = ({ className, ...rest }) => (
            <div className={classNames('del flex-center', className)} {...rest}>
                <i className="iconfont icon-del f12" />
            </div>
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
  ],
  invalid: [
    {
      code: `function Hello({ name }) { return <div>{name}</div>; }`,
      errors: [expectMissingPropTypeError1([1, 18, 1, 22])],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 0 (7.37.5)',
      code: `
        type Props = {
          name: string,
        };
        class Hello extends React.Component {
          foo(props: Props) {}
          render() {
            return this.props.name;
          }
        }
      `,
      errors: [expectMissingPropTypeError1([8, 31, 8, 35])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 1 (7.37.5)',
      code: `
        var Hello = createReactClass({
          render: function() {
            return React.createElement("div", {}, this.props.name);
          }
        });
      `,
      errors: [expectMissingPropTypeError1([4, 62, 4, 66])],
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
      errors: [expectMissingPropTypeError1([4, 43, 4, 47])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 3 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            return <div>Hello {this.props.name}</div>;
          }
        }
      `,
      errors: [expectMissingPropTypeError1([4, 43, 4, 47])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 4 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            const { foo: { bar } } = this.props;
            return <p>{bar}</p>
          }
        }
      `,
      errors: [
        expectMissingPropTypeError2([4, 21, 4, 24]),
        expectMissingPropTypeError3([4, 28, 4, 31]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 5 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            const { foo: { bar } } = this.props;
            return <p>{bar}</p>
          }
        }

        Hello.propTypes = {
          foo: PropTypes.shape({
            _: PropTypes.string,
          })
        }
      `,
      errors: [expectMissingPropTypeError3([4, 28, 4, 31])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 6 (7.37.5)',
      code: `
        function Foo({ foo: { bar } }) {
          return <p>{bar}</p>
        }
      `,
      errors: [
        expectMissingPropTypeError2([2, 24, 2, 27]),
        expectMissingPropTypeError3([2, 31, 2, 34]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 7 (7.37.5)',
      code: `
        function Foo({ a }) {
          return <p>{ a.nope }</p>
        }

        Foo.propTypes = {
          a: PropTypes.shape({
            _: PropType.string,
          })
        }
      `,
      errors: [expectMissingPropTypeError4([3, 25, 3, 29])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 8 (7.37.5)',
      code: `
        function Foo({ a }) {
          const { b } = a
          return <p>{ b.nope }</p>
        }

        Foo.propTypes = {
          a: PropTypes.shape({
            b: PropType.shape({
              _: PropType.string,
            }),
          })
        }
      `,
      errors: [
        {
          messageId: 'missingPropType',
          message: "'a.b.nope' is missing in props validation",
          line: 4,
          column: 25,
          endLine: 4,
          endColumn: 29,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 9 (7.37.5)',
      code: `
        function Foo(props) {
          const { a } = props
          return <p>{ a.nope }</p>
        }

        Foo.propTypes = {
          a: PropTypes.shape({
            _: PropType.string,
          })
        }
      `,
      errors: [expectMissingPropTypeError4([4, 25, 4, 29])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 10 (7.37.5)',
      code: `
        function Foo(props) {
          const a = props.a
          return <p>{ a.nope }</p>
        }

        Foo.propTypes = {
          a: PropTypes.shape({
            _: PropType.string,
          })
        }
      `,
      errors: [expectMissingPropTypeError4([4, 25, 4, 29])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 11 (7.37.5)',
      code: `
        class Foo extends Component {
          render() {
            const props = this.props
            return <div>{props.cat}</div>
          }
        }
      `,
      errors: [
        {
          messageId: 'missingPropType',
          message: "'cat' is missing in props validation",
          line: 5,
          column: 32,
          endLine: 5,
          endColumn: 35,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 12 (7.37.5)',
      code: `
        class Foo extends Component {
          render() {
            const { props } = this;
            return <div>{props.cat}</div>;
          }
        }
      `,
      errors: [
        {
          messageId: 'missingPropType',
          message: "'cat' is missing in props validation",
          line: 5,
          column: 32,
          endLine: 5,
          endColumn: 35,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 13 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            const { name, ...rest } = this.props
            return <div>Hello</div>;
          }
        }
      `,
      errors: [expectMissingPropTypeError1([4, 21, 4, 25])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 14 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            const { name, title, ...rest } = this.props
            return <div>Hello</div>;
          }
        }
        Hello.propTypes = {
          name: PropTypes.string
        }
      `,
      errors: [
        {
          messageId: 'missingPropType',
          message: "'title' is missing in props validation",
          line: 4,
          column: 27,
          endLine: 4,
          endColumn: 32,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 15 (7.37.5)',
      code: `
        class Hello extends React.Component {
           renderStuff() {
            const { name, ...rest } = this.props
            return (<div {...rest}>{name}</div>);
          }
          render() {
            this.renderStuff()
          }
        }
        Hello.propTypes = {};
      `,
      errors: [expectMissingPropTypeError1([4, 21, 4, 25])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 16 (7.37.5)',
      code: `
        /** @extends React.Component */
        class Hello extends ChildComponent {
          render() {
            return <div>Hello {this.props.name}</div>;
          }
        }
      `,
      errors: [expectMissingPropTypeError1([5, 43, 5, 47])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 17 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            return <div>Hello {this.props.firstname} {this.props.lastname}</div>;
          }
        }
        Hello.propTypes = {
          firstname: PropTypes.string
        };
      `,
      errors: [expectMissingPropTypeError5([4, 66, 4, 74])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 18 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            return <div>Hello {this.props.name}</div>;
          }
        }
        Hello.propTypes = {
          name: PropTypes.string
        };
        class HelloBis extends React.Component {
          render() {
            return <div>Hello {this.props.name}</div>;
          }
        }
      `,
      errors: [expectMissingPropTypeError1([12, 43, 12, 47])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 19 (7.37.5)',
      code: `
        var Hello = createReactClass({
          propTypes: {
            name: PropTypes.string.isRequired
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
        {
          messageId: 'missingPropType',
          message: "'propWithoutTypeDefinition' is missing in props validation",
          line: 7,
          column: 65,
          endLine: 7,
          endColumn: 90,
        },
        expectMissingPropTypeError1([12, 43, 12, 47]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 20 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            var { firstname, lastname } = this.props;
            return <div>Hello</div>;
          }
        }
        Hello.propTypes = {
          firstname: PropTypes.string
        };
      `,
      errors: [expectMissingPropTypeError5([4, 30, 4, 38])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 21 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            this.props.a.b
            return <div>Hello</div>;
          }
        }
        Hello.propTypes = {
          a: PropTypes.shape({
          })
        };
      `,
      errors: [expectMissingPropTypeError6([4, 26, 4, 27])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 22 (7.37.5)',
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
      `,
      errors: [expectMissingPropTypeError7([4, 28, 4, 29])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 23 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            this.props.a.b.c;
            return <div>Hello</div>;
          }
        }
        Hello.propTypes = {
          a: PropTypes.shape({})
        };
        Hello.propTypes.a.b = PropTypes.shape({});
      `,
      errors: [expectMissingPropTypeError7([4, 28, 4, 29])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 24 (7.37.5)',
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
            })
          )
        };
      `,
      errors: [
        expectMissingPropTypeError7([4, 28, 4, 29]),
        {
          messageId: 'missingPropType',
          message: "'a.__.d' is missing in props validation",
          line: 5,
          column: 29,
          endLine: 5,
          endColumn: 30,
        },
        {
          messageId: 'missingPropType',
          message: "'a.__.d.length' is missing in props validation",
          line: 5,
          column: 31,
          endLine: 5,
          endColumn: 37,
        },
        {
          messageId: 'missingPropType',
          message: "'a.anything.e' is missing in props validation",
          line: 6,
          column: 35,
          endLine: 6,
          endColumn: 36,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 25 (7.37.5)',
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
            })
          )
        };
      `,
      errors: [
        {
          messageId: 'missingPropType',
          message: "'a[].c' is missing in props validation",
          line: 5,
          column: 29,
          endLine: 5,
          endColumn: 30,
        },
        {
          messageId: 'missingPropType',
          message: "'a[].d' is missing in props validation",
          line: 6,
          column: 29,
          endLine: 6,
          endColumn: 30,
        },
        {
          messageId: 'missingPropType',
          message: "'a[].d.length' is missing in props validation",
          line: 6,
          column: 31,
          endLine: 6,
          endColumn: 37,
        },
        {
          messageId: 'missingPropType',
          message: "'a[].e' is missing in props validation",
          line: 7,
          column: 33,
          endLine: 7,
          endColumn: 34,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 26 (7.37.5)',
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
              c: PropTypes.number,
              e: PropTypes.array
            })
          ])
        };
      `,
      errors: [
        {
          messageId: 'missingPropType',
          message: "'a.length' is missing in props validation",
          line: 4,
          column: 26,
          endLine: 4,
          endColumn: 32,
        },
        expectMissingPropTypeError6([5, 26, 5, 27]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 27 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            var {
              "aria-controls": ariaControls,
              propX,
              ...props } = this.props;
            return <div>Hello</div>;
          }
        }
        Hello.propTypes = {
          "aria-controls": PropTypes.string
        };
      `,
      errors: [
        {
          messageId: 'missingPropType',
          message: "'propX' is missing in props validation",
          line: 6,
          column: 15,
          endLine: 6,
          endColumn: 20,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 28 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            this.props["some.value"];
            return <div>Hello</div>;
          }
        }
        Hello.propTypes = {
        };
      `,
      errors: [
        {
          messageId: 'missingPropType',
          message: "'some.value' is missing in props validation",
          line: 4,
          column: 24,
          endLine: 4,
          endColumn: 36,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 29 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            this.props["arr"][1];
            return <div>Hello</div>;
          }
        }
        Hello.propTypes = {
        };
      `,
      errors: [
        {
          messageId: 'missingPropType',
          message: "'arr' is missing in props validation",
          line: 4,
          column: 24,
          endLine: 4,
          endColumn: 29,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 30 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            this.props["arr"][1]["some.value"];
            return <div>Hello</div>;
          }
        }
        Hello.propTypes = {
          "arr": PropTypes.arrayOf(
            PropTypes.shape({})
          )
        };
      `,
      errors: [
        {
          messageId: 'missingPropType',
          message: "'arr[].some.value' is missing in props validation",
          line: 4,
          column: 34,
          endLine: 4,
          endColumn: 46,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 31 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            var text;
            text = 'Hello ';
            let { props: { firstname } } = this;
            return <div>{text} {firstname}</div>;
          }
        }
      `,
      errors: [expectMissingPropTypeError8([6, 28, 6, 37])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 32 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            var { 'props': { firstname } } = this;
            return <div>Hello {firstname}</div>;
          }
        }
      `,
      errors: [expectMissingPropTypeError8([4, 30, 4, 39])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 33 (7.37.5)',
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
          lastname: PropTypes.string
        }
      `,
      errors: [expectMissingPropTypeError8([5, 40, 5, 49])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 34 (7.37.5)',
      code: `
        var Hello = function(props) {
          return <div>Hello {props.name}</div>;
        }
      `,
      errors: [expectMissingPropTypeError1([3, 36, 3, 40])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 35 (7.37.5)',
      code: `
        function Hello(props) {
          return <div>Hello {props.name}</div>;
        }
      `,
      errors: [expectMissingPropTypeError1([3, 36, 3, 40])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 36 (7.37.5)',
      code: `
        var Hello = (props) => {
          return <div>Hello {props.name}</div>;
        }
      `,
      errors: [expectMissingPropTypeError1([3, 36, 3, 40])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 37 (7.37.5)',
      code: `
        var Hello = (props) => {
          const { name } = props;
          return <div>Hello {name}</div>;
        }
      `,
      errors: [expectMissingPropTypeError1([3, 19, 3, 23])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 38 (7.37.5)',
      code: `
        function Hello({ name }) {
          return <div>Hello {name}</div>;
        }
      `,
      errors: [expectMissingPropTypeError1([2, 26, 2, 30])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 39 (7.37.5)',
      code: `
        const Hello = function({ name }) {
          return <div>Hello {name}</div>;
        }
      `,
      errors: [expectMissingPropTypeError1([2, 34, 2, 38])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 40 (7.37.5)',
      code: `
        const Hello = ({ name }) => {
          return <div>Hello {name}</div>;
        }
      `,
      errors: [expectMissingPropTypeError1([2, 26, 2, 30])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 41 (7.37.5)',
      skip: 'Upstream parsers.all excludes this fixture for every parser (no-default, no-ts and no-babel).',
      code: `
        class Test extends Foo.Component {
          render() {
            return (
              <div>{this.props.firstname} {this.props.lastname}</div>
            );
          }
        };
        const propTypes = forbidExtraProps({
          firstname: PropTypes.string
        });
        Test.propTypes = propTypes;
      `,
      settings: { propWrapperFunctions: ['forbidExtraProps'] },
      errors: [expectMissingPropTypeError5([5, 55, 5, 63])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 42 (7.37.5)',
      code: `
        class Test extends Foo.Component {
          render() {
            return (
              <div>{this.props.firstname} {this.props.lastname}</div>
            );
          }
        }
        Test.propTypes = {
          firstname: PropTypes.string
        };
      `,
      settings: { react: { pragma: 'Foo' } },
      errors: [expectMissingPropTypeError5([5, 55, 5, 63])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 43 (7.37.5)',
      code: `
        type MyComponentProps = {
          a: number,
        };
        function MyComponent({ a, b }: MyComponentProps) {
          return <div />;
        }
      `,
      errors: [expectMissingPropTypeError9([5, 35, 5, 36])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 44 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            var props = { firstname: 'John' };
            return <div>Hello {props.firstname} {this.props.lastname}</div>;
          }
        }
      `,
      errors: [expectMissingPropTypeError5([5, 61, 5, 69])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 45 (7.37.5)',
      code: `
        class Hello extends React.Component {
          constructor(props, context) {
            super(props, context)
            this.state = { status: props.source }
          }
        }
      `,
      errors: [expectMissingPropTypeError10([5, 42, 5, 48])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 46 (7.37.5)',
      code: `
        class Hello extends React.Component {
          constructor(props, context) {
            super(props, context)
            this.state = { status: props.source.uri }
          }
        }
      `,
      errors: [
        expectMissingPropTypeError10([5, 42, 5, 48]),
        {
          messageId: 'missingPropType',
          message: "'source.uri' is missing in props validation",
          line: 5,
          column: 49,
          endLine: 5,
          endColumn: 52,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 47 (7.37.5)',
      code: `
        class Hello extends React.Component {
          constructor(props, context) {
            super(props, context)
            this.state = { status: this.props.source }
          }
        }
      `,
      errors: [expectMissingPropTypeError10([5, 47, 5, 53])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 48 (7.37.5)',
      code: `
        class Hello extends React.Component {
          constructor(props, context) {
            super(props, context)
            this.state = { status: this.props.source.uri }
          }
        }
      `,
      errors: [
        expectMissingPropTypeError10([5, 47, 5, 53]),
        {
          messageId: 'missingPropType',
          message: "'source.uri' is missing in props validation",
          line: 5,
          column: 54,
          endLine: 5,
          endColumn: 57,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 49 (7.37.5)',
      code: `
        function HelloComponent() {
          class Hello extends React.Component {
            render() {
              return <div>Hello {this.props.name}</div>;
            }
          }
          return Hello;
        }
        module.exports = HelloComponent();
      `,
      errors: [expectMissingPropTypeError1([5, 45, 5, 49])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 50 (7.37.5)',
      code: `
        function HelloComponent() {
          var Hello = createReactClass({
            render: function() {
              return <div>Hello {this.props.name}</div>;
            }
          });
          return Hello;
        }
        module.exports = HelloComponent();
      `,
      errors: [expectMissingPropTypeError1([5, 45, 5, 49])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 51 (7.37.5)',
      code: `
        class DynamicHello extends Component {
          render() {
            const { firstname } = this.props;
            class Hello extends Component {
              render() {
                const { name } = this.props;
                return <div>Hello {name}</div>;
              }
            }
            Hello = connectReduxForm({ name: firstname })(Hello);
            return <Hello />;
          }
        }
      `,
      errors: [
        expectMissingPropTypeError8([4, 21, 4, 30]),
        expectMissingPropTypeError1([7, 25, 7, 29]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 52 (7.37.5)',
      code: `
        const Hello = (props) => {
          let team = props.names.map((name) => {
              return <li>{name}, {props.company}</li>;
            });
          return <ul>{team}</ul>;
        };
      `,
      errors: [
        {
          messageId: 'missingPropType',
          message: "'names' is missing in props validation",
          line: 3,
          column: 28,
          endLine: 3,
          endColumn: 33,
        },
        {
          messageId: 'missingPropType',
          message: "'names.map' is missing in props validation",
          line: 3,
          column: 34,
          endLine: 3,
          endColumn: 37,
        },
        {
          messageId: 'missingPropType',
          message: "'company' is missing in props validation",
          line: 4,
          column: 41,
          endLine: 4,
          endColumn: 48,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 53 (7.37.5)',
      code: `
        const Annotation = props => (
          <div>
            {props.text}
          </div>
        )
      `,
      errors: [expectMissingPropTypeError11([4, 20, 4, 24])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 54 (7.37.5)',
      code: `
        for (var key in foo) {
          var Hello = createReactClass({
            render: function() {
              return <div>Hello {this.props.name}</div>;
            }
          });
        }
      `,
      errors: [expectMissingPropTypeError1([5, 45, 5, 49])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 55 (7.37.5)',
      code: `
        var propTypes = {
          firstname: PropTypes.string
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
      errors: [expectMissingPropTypeError5([8, 55, 8, 63])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 56 (7.37.5)',
      code: `
        class Test extends Foo.Component {
          render() {
            return (
              <div>{this.props.firstname} {this.props.lastname}</div>
            );
          }
        }
        Test.propTypes = forbidExtraProps({
          firstname: PropTypes.string
        });
      `,
      settings: {
        react: { pragma: 'Foo' },
        propWrapperFunctions: ['forbidExtraProps'],
      },
      errors: [expectMissingPropTypeError5([5, 55, 5, 63])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 57 (7.37.5)',
      code: `
        class Test extends Foo.Component {
          render() {
            return (
              <div>{this.props.firstname} {this.props.lastname}</div>
            );
          }
        }
        Test.propTypes = Object.freeze({
          firstname: PropTypes.string
        });
      `,
      settings: {
        react: { pragma: 'Foo' },
        propWrapperFunctions: ['Object.freeze'],
      },
      errors: [expectMissingPropTypeError5([5, 55, 5, 63])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 58 (7.37.5)',
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
          firstname: PropTypes.string
        };
      `,
      errors: [expectMissingPropTypeError5([6, 55, 6, 63])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 59 (7.37.5)',
      code: `
        class Hello extends React.Component {
          props: {};
          render () {
            return <div>Hello {this.props.name}</div>;
          }
        }
      `,
      errors: [expectMissingPropTypeError1([5, 43, 5, 47])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 60 (7.37.5)',
      code: `
        class Hello extends React.Component {
          props: {
            name: Object;
          };
          render () {
            return <div>Hello {this.props.firstname}</div>;
          }
        }
      `,
      errors: [expectMissingPropTypeError8([7, 43, 7, 52])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 61 (7.37.5)',
      code: `
        type Props = { name: Object; };
        class Hello extends React.Component {
          props: Props;
          render () {
            return <div>Hello {this.props.firstname}</div>;
          }
        }
      `,
      errors: [expectMissingPropTypeError8([6, 43, 6, 52])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 62 (7.37.5)',
      skip: 'Upstream checks Flow prop annotations for prop-types; the native TypeScript annotation path reports 0 diagnostics instead of 1.',
      code: `
        class Hello extends React.Component {
          props: {
            name: {
              firstname: string;
            }
          };
          render () {
            return <div>Hello {this.props.name.lastname}</div>;
          }
        }
      `,
      errors: [expectMissingPropTypeError12([9, 48, 9, 56])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 63 (7.37.5)',
      skip: 'Upstream checks Flow prop annotations for prop-types; the native TypeScript annotation path reports 0 diagnostics instead of 1.',
      code: `
        type Props = { name: { firstname: string; }; };
        class Hello extends React.Component {
          props: Props;
          render () {
            return <div>Hello {this.props.name.lastname}</div>;
          }
        }
      `,
      errors: [expectMissingPropTypeError12([6, 48, 6, 56])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 64 (7.37.5)',
      skip: 'Upstream checks Flow prop annotations for prop-types; the native TypeScript annotation path reports 0 diagnostics instead of 1.',
      code: `
        class Hello extends React.Component {
          props: { person: { name: { firstname: string; }; }; };
          render () {
            return <div>Hello {this.props.person.name.lastname}</div>;
          }
        }
      `,
      errors: [
        {
          messageId: 'missingPropType',
          message: "'person.name.lastname' is missing in props validation",
          line: 5,
          column: 55,
          endLine: 5,
          endColumn: 63,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 65 (7.37.5)',
      skip: 'Upstream checks Flow prop annotations for prop-types; the native TypeScript annotation path reports 0 diagnostics instead of 1.',
      code: `
        type Props = { person: { name: { firstname: string; }; }; };
        class Hello extends React.Component {
          props: Props;
          render () {
            return <div>Hello {this.props.person.name.lastname}</div>;
          }
        }
      `,
      errors: [
        {
          messageId: 'missingPropType',
          message: "'person.name.lastname' is missing in props validation",
          line: 6,
          column: 55,
          endLine: 6,
          endColumn: 63,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 66 (7.37.5)',
      skip: 'Upstream checks Flow prop annotations for prop-types; the native TypeScript annotation path reports 0 diagnostics instead of 1.',
      code: `
        type Person = { name: { firstname: string; } };
        class Hello extends React.Component {
          props: { people: Person[]; };
          render () {
            var names = [];
            for (var i = 0; i < this.props.people.length; i++) {
              names.push(this.props.people[i].name.lastname);
            }
            return <div>Hello {names.join(
        )}</div>;
          }
        }
      `,
      errors: [
        {
          messageId: 'missingPropType',
          message: "'people[].name.lastname' is missing in props validation",
          line: 8,
          column: 52,
          endLine: 8,
          endColumn: 60,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 67 (7.37.5)',
      skip: 'Upstream checks Flow prop annotations for prop-types; the native TypeScript annotation path reports 0 diagnostics instead of 1.',
      code: `
        type Person = { name: { firstname: string } };
        type Props = { people: Person[] };
        class Hello extends React.Component {
          props: Props;
          render () {
            var names = [];
            for (var i = 0; i < this.props.people.length; i++) {
              names.push(this.props.people[i].name.lastname);
            }
            return <div>Hello {names.join(
        )}</div>;
          }
        }
      `,
      errors: [
        {
          messageId: 'missingPropType',
          message: "'people[].name.lastname' is missing in props validation",
          line: 9,
          column: 52,
          endLine: 9,
          endColumn: 60,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 68 (7.37.5)',
      skip: 'Upstream checks Flow prop annotations for prop-types; the native TypeScript annotation path reports 0 diagnostics instead of 1.',
      code: `
        type Props = { result?: { ok: string | boolean; } | { ok: number | Array } };
        class Hello extends React.Component {
          props: Props;
          render () {
            return <div>Hello {this.props.result.notok}</div>;
          }
        }
      `,
      errors: [expectMissingPropTypeError13([6, 50, 6, 55])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 69 (7.37.5)',
      code: `
        let Greetings = class extends React.Component {
          render () {
            return <div>Hello {this.props.name}</div>;
          }
        }
        Greetings.propTypes = {};
      `,
      errors: [expectMissingPropTypeError1([4, 43, 4, 47])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 70 (7.37.5)',
      code: `
        let Greetings = {
          Hello: class extends React.Component {
            render () {
              return <div>Hello {this.props.name}</div>;
            }
          }
        }
        Greetings.Hello.propTypes = {};
      `,
      errors: [expectMissingPropTypeError1([5, 45, 5, 49])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 71 (7.37.5)',
      code: `
        let Greetings = {};
        Greetings.Hello = class extends React.Component {
          render () {
            return <div>Hello {this.props.name}</div>;
          }
        }
        Greetings.Hello.propTypes = {};
      `,
      errors: [expectMissingPropTypeError1([5, 43, 5, 47])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 72 (7.37.5)',
      code: `
        function Greetings({ names }) {
          names = names.map(({ firstname, lastname }) => <div>{firstname} {lastname}</div>);
          return <Hello>{names}</Hello>;
        }
      `,
      errors: [
        {
          messageId: 'missingPropType',
          message: "'names' is missing in props validation",
          line: 2,
          column: 30,
          endLine: 2,
          endColumn: 35,
        },
        {
          messageId: 'missingPropType',
          message: "'names.map' is missing in props validation",
          line: 3,
          column: 25,
          endLine: 3,
          endColumn: 28,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 73 (7.37.5)',
      code: `
        const MyComponent = props => (
          <div onClick={() => props.toggle()}></div>
        )
      `,
      errors: [
        {
          messageId: 'missingPropType',
          message: "'toggle' is missing in props validation",
          line: 3,
          column: 37,
          endLine: 3,
          endColumn: 43,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 74 (7.37.5)',
      code: `
        const MyComponent = props => props.test ? <div /> : <span />
      `,
      errors: [expectMissingPropTypeError14([2, 44, 2, 48])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 75 (7.37.5)',
      code: `
        const TestComponent = props =>
          <div onClick={() => props.test()} />
        const mapStateToProps = (_, props) => ({
          otherProp: props.otherProp,
        })
      `,
      errors: [expectMissingPropTypeError14([3, 37, 3, 41])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 76 (7.37.5)',
      code: `
        type Props = {
          firstname: ?string,
        };
        function Hello({ firstname, lastname }: Props): React$Element {
          return <div>Hello {firstname} {lastname}</div>;
        }
      `,
      errors: [expectMissingPropTypeError5([5, 37, 5, 45])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 77 (7.37.5)',
      code: `
        class Hello extends React.Component {
          constructor(props, context) {
            super(props, context)
            const firstname = props.firstname;
            const { lastname } = props;
            this.state = {
              firstname,
              lastname
            }
          }
        }
      `,
      errors: [
        expectMissingPropTypeError8([5, 37, 5, 46]),
        expectMissingPropTypeError5([6, 21, 6, 29]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 78 (7.37.5)',
      code: `
        function Hello(props) {
          return <div>{props.name.constructor.firstname}</div>
        }
        Hello.propTypes = {
          name: PropTypes.shape({
            firstname: PropTypes.object
          })
        };
      `,
      errors: [
        {
          messageId: 'missingPropType',
          message:
            "'name.constructor.firstname' is missing in props validation",
          line: 3,
          column: 47,
          endLine: 3,
          endColumn: 56,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 79 (7.37.5)',
      code: `
      function Hello({ foo = '' }) {
        return <p>{foo}</p>
      }
    `,
      errors: [expectMissingPropTypeError2([2, 24, 2, 27])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 80 (7.37.5)',
      code: `
        function SomeComponent({bar}) {
          function f({foo}) {}
          return <div className={f()}>{bar}</div>;
        }
      `,
      errors: [expectMissingPropTypeError15([2, 33, 2, 36])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 81 (7.37.5)',
      code: `
        function SomeComponent({bar} = baz) {
          function f({foo}) {}
          return <div className={f()}>{bar}</div>;
        }
      `,
      errors: [expectMissingPropTypeError15([2, 33, 2, 36])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 82 (7.37.5)',
      code: `
        class Hello extends React.PureComponent {
          render() {
            return <div>Hello {this.props.name}</div>;
          }
        }
      `,
      errors: [expectMissingPropTypeError1([4, 43, 4, 47])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 83 (7.37.5)',
      code: `
        class Hello extends PureComponent {
          render() {
            return <div>Hello {this.props.name}</div>;
          }
        }
      `,
      errors: [expectMissingPropTypeError1([4, 43, 4, 47])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 85 (7.37.5)',
      code: `
        var Hello = createReactClass({
          propTypes: {},
          render: function() {
            return <div>{this.props.firstname}</div>;
          }
        });
      `,
      options: [{ skipUndeclared: true }],
      errors: [expectMissingPropTypeError8([5, 37, 5, 46])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 86 (7.37.5)',
      code: `
        var Hello = function(props) {
          return <div>{props.firstname}</div>;
        };
        Hello.propTypes = {}
      `,
      options: [{ skipUndeclared: true }],
      errors: [expectMissingPropTypeError8([3, 30, 3, 39])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 87 (7.37.5)',
      code: `
        class Hello extends React.Component {
          static get propTypes() {
            return {};
          }
          render() {
            return <div>{this.props.firstname}</div>;
          }
        }
      `,
      options: [{ skipUndeclared: true }],
      errors: [expectMissingPropTypeError8([7, 37, 7, 46])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 88 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            return <div>{this.props.firstname}</div>;
          }
        }
        Hello.propTypes = {};
      `,
      options: [{ skipUndeclared: true }],
      errors: [expectMissingPropTypeError8([4, 37, 4, 46])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 89 (7.37.5)',
      code: `
        var Hello = createReactClass({
          render: function() {
            return <div>{this.props.firstname}</div>;
          }
        });
      `,
      options: [{ skipUndeclared: false }],
      errors: [expectMissingPropTypeError8([4, 37, 4, 46])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 90 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream flow syntax (Property or signature expected.).',
      code: `
        type MyComponentProps = {
          +a: number,
        };
        function MyComponent({ a, b }: MyComponentProps) {
          return <div />;
        }
      `,
      errors: [expectMissingPropTypeError9([5, 35, 5, 36])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 91 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream flow syntax (Property or signature expected.).',
      code: `
        type MyComponentProps = {
          -a: number,
        };
        function MyComponent({ a, b }: MyComponentProps) {
          return <div />;
        }
      `,
      errors: [expectMissingPropTypeError9([5, 35, 5, 36])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 92 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream flow syntax (Property or signature expected.).',
      code: `
        type Props = {+name: Object;};
        class Hello extends React.Component {
          props: Props;
          render () {
            return <div>Hello {this.props.firstname}</div>;
          }
        }
      `,
      errors: [expectMissingPropTypeError8([6, 43, 6, 52])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 93 (7.37.5)',
      code: `
        class Hello extends React.Component {
          onSelect = async ({ name }) => {
            return this.props.foo;
          }
          render() {
            return <Greeting onSelect={this.onSelect} />;
          }
        }
      `,
      errors: [expectMissingPropTypeError2([4, 31, 4, 34])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 94 (7.37.5)',
      code: `
        class Hello extends Component {
          static propTypes = forbidExtraProps({
            bar: PropTypes.func
          })
          componentWillReceiveProps(nextProps) {
            if (nextProps.foo) {
              return;
            }
          }
          render() {
            return <div bar={this.props.bar} />;
          }
        }
      `,
      settings: {
        react: { pragma: 'Foo' },
        propWrapperFunctions: ['forbidExtraProps'],
      },
      errors: [expectMissingPropTypeError2([7, 27, 7, 30])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 95 (7.37.5)',
      code: `
        class Hello extends Component {
          static propTypes = {
            bar: PropTypes.func
          }
          componentWillReceiveProps(nextProps) {
            if (nextProps.foo) {
              return;
            }
          }
          render() {
            return <div bar={this.props.bar} />;
          }
        }
      `,
      errors: [expectMissingPropTypeError2([7, 27, 7, 30])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 96 (7.37.5)',
      code: `
        class Hello extends Component {
          static propTypes = {
            bar: PropTypes.func
          }
          componentWillReceiveProps(nextProps) {
            const {foo} = nextProps;
            if (foo) {
              return;
            }
          }
          render() {
            return <div bar={this.props.bar} />;
          }
        }
      `,
      errors: [expectMissingPropTypeError2([7, 20, 7, 23])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 97 (7.37.5)',
      code: `
        class Hello extends Component {
          static propTypes = {
            bar: PropTypes.func
          }
          componentWillReceiveProps({foo}) {
            if (foo) {
              return;
            }
          }
          render() {
            return <div bar={this.props.bar} />;
          }
        }
      `,
      errors: [expectMissingPropTypeError2([6, 38, 6, 41])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 98 (7.37.5)',
      code: `
        class Hello extends Component {
          componentWillReceiveProps({foo}) {
            if (foo) {
              return;
            }
          }
          render() {
            return <div bar={this.props.bar} />;
          }
        }
        Hello.propTypes = {
            bar: PropTypes.func
          }
      `,
      errors: [expectMissingPropTypeError2([3, 38, 3, 41])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 99 (7.37.5)',
      code: `
        class Hello extends Component {
          static propTypes = forbidExtraProps({
            bar: PropTypes.func
          })
          shouldComponentUpdate(nextProps) {
            if (nextProps.foo) {
              return;
            }
          }
          render() {
            return <div bar={this.props.bar} />;
          }
        }
      `,
      settings: {
        react: { pragma: 'Foo' },
        propWrapperFunctions: ['forbidExtraProps'],
      },
      errors: [expectMissingPropTypeError2([7, 27, 7, 30])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 100 (7.37.5)',
      code: `
        class Hello extends Component {
          shouldComponentUpdate({foo}) {
            if (foo) {
              return;
            }
          }
          render() {
            return <div bar={this.props.bar} />;
          }
        }
        Hello.propTypes = {
            bar: PropTypes.func
          }
      `,
      errors: [expectMissingPropTypeError2([3, 34, 3, 37])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 101 (7.37.5)',
      code: `
        class Hello extends React.Component {
          static propTypes() {
            return {
              name: PropTypes.string
            };
          }
          render() {
            return <div>Hello {this.props.name}</div>;
          }
        }
      `,
      errors: [expectMissingPropTypeError1([9, 43, 9, 47])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 102 (7.37.5)',
      skip: 'Upstream Flow React.Component<void, Props, State> puts props in the second type argument; the native TypeScript model uses the first.',
      code: `
        type Person = {
          firstname: string
        };
        class Hello extends React.Component<void, Person, void> {
          render () {
            return <div>Hello {this.props.lastname}</div>;
          }
        }
      `,
      errors: [expectMissingPropTypeError5([7, 43, 7, 51])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 103 (7.37.5)',
      skip: 'Upstream Flow React.Component<void, Props, State> puts props in the second type argument; the native TypeScript model uses the first.',
      code: `
        type Person = {
          firstname: string
        };
        class Hello extends React.Component<void, Person, void> {
          render () {
            return <div>Hello {this.props.lastname}</div>;
          }
        }
      `,
      settings: { react: { flowVersion: '0.52' } },
      errors: [expectMissingPropTypeError5([7, 43, 7, 51])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 104 (7.37.5)',
      skip: 'Upstream Flow React.Component<void, Props, State> puts props in the second type argument; the native TypeScript model uses the first.',
      code: `
        type Person = {
          firstname: string
        };
        class Hello extends React.Component<void, Person, void> {
          render () {
            const {
              lastname,
            } = this.props;
            return <div>Hello {lastname}</div>;
          }
        }
      `,
      errors: [expectMissingPropTypeError5([8, 15, 8, 23])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 105 (7.37.5)',
      skip: 'Upstream Flow React.Component<void, Props, State> puts props in the second type argument; the native TypeScript model uses the first.',
      code: `
        type Person = {
          firstname: string
        };
        class Hello extends React.Component<void, Person, void> {
          render () {
            const {
              lastname,
            } = this.props;
            return <div>Hello {lastname}</div>;
          }
        }
      `,
      settings: { react: { flowVersion: '0.52' } },
      errors: [expectMissingPropTypeError5([8, 15, 8, 23])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 106 (7.37.5)',
      skip: 'Upstream Flow React.Component<void, Props, State> puts props in the second type argument; the native TypeScript model uses the first.',
      code: `
        type Props = {name: {firstname: string;};};
        class Hello extends React.Component<void, Props, void> {
          render () {
            return <div>Hello {this.props.name.lastname}</div>;
          }
        }
      `,
      errors: [expectMissingPropTypeError12([5, 48, 5, 56])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 107 (7.37.5)',
      skip: 'Upstream Flow React.Component<void, Props, State> puts props in the second type argument; the native TypeScript model uses the first.',
      code: `
        type Props = {name: {firstname: string;};};
        class Hello extends React.Component<void, Props, void> {
          render () {
            return <div>Hello {this.props.name.lastname}</div>;
          }
        }
      `,
      settings: { react: { flowVersion: '0.52' } },
      errors: [expectMissingPropTypeError12([5, 48, 5, 56])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 108 (7.37.5)',
      skip: 'Upstream Flow React.Component<void, Props, State> puts props in the second type argument; the native TypeScript model uses the first.',
      code: `
        type Props = {result?: {ok: string | boolean;}|{ok: number | Array}};
        class Hello extends React.Component<void, Props, void> {
          render () {
            return <div>Hello {this.props.result.notok}</div>;
          }
        }
      `,
      errors: [expectMissingPropTypeError13([5, 50, 5, 55])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 109 (7.37.5)',
      skip: 'Upstream Flow React.Component<void, Props, State> puts props in the second type argument; the native TypeScript model uses the first.',
      code: `
        type Props = {result?: {ok: string | boolean;}|{ok: number | Array}};
        class Hello extends React.Component<void, Props, void> {
          render () {
            return <div>Hello {this.props.result.notok}</div>;
          }
        }
      `,
      settings: { react: { flowVersion: '0.52' } },
      errors: [expectMissingPropTypeError13([5, 50, 5, 55])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 110 (7.37.5)',
      skip: 'Upstream Flow React.Component<void, Props, State> puts props in the second type argument; the native TypeScript model uses the first.',
      code: `
        type Person = {
          firstname: string
        };
        class Hello extends React.Component<void, { person: Person }, void> {
          render () {
            return <div>Hello {this.props.person.lastname}</div>;
          }
        }
      `,
      errors: [expectMissingPropTypeError16([7, 50, 7, 58])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 111 (7.37.5)',
      skip: 'Upstream Flow React.Component<void, Props, State> puts props in the second type argument; the native TypeScript model uses the first.',
      code: `
        type Person = {
          firstname: string
        };
        class Hello extends React.Component<void, { person: Person }, void> {
          render () {
            return <div>Hello {this.props.person.lastname}</div>;
          }
        }
      `,
      settings: { react: { flowVersion: '0.52' } },
      errors: [expectMissingPropTypeError16([7, 50, 7, 58])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 112 (7.37.5)',
      code: `
        type Props = {
          foo: string,
        };

        class Bar extends React.Component<Props> {
          render() {
            return <div>{this.props.bar}</div>
          }
        }
      `,
      errors: [expectMissingPropTypeError15([8, 37, 8, 40])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 113 (7.37.5)',
      code: `
        type Props = {
          foo: string,
        };

        class Bar extends React.Component<Props> {
          render() {
            return <div>{this.props.bar}</div>
          }
        }
      `,
      settings: { react: { flowVersion: '0.53' } },
      errors: [expectMissingPropTypeError15([8, 37, 8, 40])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 114 (7.37.5)',
      code: `
        type FancyProps = {
          foo: string,
        };

        class Bar extends React.Component<FancyProps> {
          render() {
            return <div>{this.props.bar}</div>
          }
        }
      `,
      errors: [expectMissingPropTypeError15([8, 37, 8, 40])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 115 (7.37.5)',
      code: `
        type FancyProps = {
          foo: string,
        };

        class Bar extends React.Component<FancyProps> {
          render() {
            return <div>{this.props.bar}</div>
          }
        }
      `,
      settings: { react: { flowVersion: '0.53' } },
      errors: [expectMissingPropTypeError15([8, 37, 8, 40])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 116 (7.37.5)',
      skip: 'Upstream checks Flow prop annotations for prop-types; the native TypeScript annotation path reports 0 diagnostics instead of 1.',
      code: `
        type Person = {
          firstname: string
        };
        class Hello extends React.Component<{ person: Person }> {
          render () {
            return <div>Hello {this.props.person.lastname}</div>;
          }
        }
      `,
      errors: [expectMissingPropTypeError16([7, 50, 7, 58])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 117 (7.37.5)',
      skip: 'Upstream checks Flow prop annotations for prop-types; the native TypeScript annotation path reports 0 diagnostics instead of 1.',
      code: `
        type Person = {
          firstname: string
        };
        class Hello extends React.Component<{ person: Person }> {
          render () {
            return <div>Hello {this.props.person.lastname}</div>;
          }
        }
      `,
      settings: { react: { flowVersion: '0.53' } },
      errors: [expectMissingPropTypeError16([7, 50, 7, 58])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 118 (7.37.5)',
      code: `
        type Props = { foo: string }
        function higherOrderComponent<Props>() {
          return class extends React.Component<Props> {
            render() {
              return <div>{this.props.foo} - {this.props.bar}</div>
            }
          }
        }
      `,
      errors: [expectMissingPropTypeError15([6, 58, 6, 61])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 119 (7.37.5)',
      code: `
        type PropsA = {foo: string };
        type PropsB = { bar: string };
        type Props = PropsA & PropsB;

        class MyComponent extends React.Component {
          props: Props;

          render() {
            return <div>{this.props.foo} - {this.props.bar} - {this.props.fooBar}</div>
          }
        }
      `,
      errors: [expectMissingPropTypeError17([10, 75, 10, 81])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 120 (7.37.5)',
      code: `
        type PropsA = { foo: string };
        type PropsB = { bar: string };
        type PropsC = { zap: string };
        type Props = PropsA & PropsB;

        class Bar extends React.Component {
          props: Props & PropsC;

          render() {
            return <div>{this.props.foo} - {this.props.bar} - {this.props.zap} - {this.props.fooBar}</div>
          }
        }
      `,
      errors: [expectMissingPropTypeError17([11, 94, 11, 100])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 121 (7.37.5)',
      code: `
        type PropsB = { bar: string };
        type PropsC = { zap: string };
        type Props = PropsB & {
          baz: string
        };

        class Bar extends React.Component {
          props: Props & PropsC;

          render() {
            return <div>{this.props.bar} - {this.props.baz} - {this.props.fooBar}</div>
          }
        }
      `,
      errors: [expectMissingPropTypeError17([12, 75, 12, 81])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 122 (7.37.5)',
      code: `
        type PropsB = { bar: string };
        type PropsC = { zap: string };
        type Props = {
          baz: string
        } & PropsB;

        class Bar extends React.Component {
          props: Props & PropsC;

          render() {
            return <div>{this.props.bar} - {this.props.baz} - {this.props.fooBar}</div>
          }
        }
      `,
      errors: [expectMissingPropTypeError17([12, 75, 12, 81])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 123 (7.37.5)',
      skip: "The native TypeScript parser cannot parse this upstream flow syntax (',' expected.).",
      code: `
      type ReduxState = {bar: number};

      const mapStateToProps = (state: ReduxState) => ({
          foo: state.bar,
      });
      // utility to extract the return type from a function
      type ExtractReturn_<R, Fn: (...args: any[]) => R> = R;
      type ExtractReturn<T> = ExtractReturn_<*, T>;

      type PropsFromRedux = ExtractReturn<typeof mapStateToProps>;

      type OwnProps = {
          baz: string,
      }

      // I want my Props to be {baz: string, foo: number}
      type Props = PropsFromRedux & OwnProps;

      const Component = (props: Props) => (
        <div>
            {props.baz}
            {props.bad}
        </div>
      );
    `,
      errors: [
        {
          messageId: 'missingPropType',
          message: "'bad' is missing in props validation",
          line: 23,
          column: 20,
          endLine: 23,
          endColumn: 23,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 124 (7.37.5)',
      code: `
        class Component extends React.Component {
          render() {
            return <div>{this.props.foo.baz}</div>;
          }
        }
        Component.propTypes = {
          foo: PropTypes.oneOfType([
            PropTypes.shape({
              bar: PropTypes.string
            })
          ])
        };
      `,
      errors: [expectMissingPropTypeError18([4, 41, 4, 44])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 125 (7.37.5)',
      code: `
        class Foo extends React.Component {
          bar() {
            this.setState((state, props) => ({ current: props.current, bar: props.bar }));
          }
          render() {
            return <div />;
          }
        }

        Foo.propTypes = {
          current: PropTypes.number.isRequired,
        };
      `,
      errors: [expectMissingPropTypeError15([4, 83, 4, 86])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 126 (7.37.5)',
      code: `
        class Foo extends React.Component {
          setZoo() {
            this.setState((state, {zoo}) => ({ zoo }));
          }
          render() {
            return <div />;
          }
        }
      `,
      errors: [
        {
          messageId: 'missingPropType',
          message: "'zoo' is missing in props validation",
          line: 4,
          column: 36,
          endLine: 4,
          endColumn: 39,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 127 (7.37.5)',
      code: `
        class Foo extends React.Component {
          static getDerivedStateFromProps(props) {
            const { foo, bar } = props;
            return {
              foobar: foo + bar
            };
          }

          render() {
            const { foobar } = this.state;
            return <div>{foobar}</div>;
          }
        }

        Foo.propTypes = {
          foo: PropTypes.func.isRequired,
        };
      `,
      settings: { react: { version: '16.3.0' } },
      errors: [expectMissingPropTypeError15([4, 26, 4, 29])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 128 (7.37.5)',
      code: `
        const ForAttendees = ({ page }) => (
          <>
            <section>{page}</section>
          </>
        );

        export default ForAttendees;
      `,
      errors: [
        {
          messageId: 'missingPropType',
          message: "'page' is missing in props validation",
          line: 2,
          column: 33,
          endLine: 2,
          endColumn: 37,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 129 (7.37.5)',
      code: `
        const HeaderBalance = React.memo(({ cryptoCurrency }) => (
          <div className="header-balance">
            <div className="header-balance__balance">
              BTC
              {cryptoCurrency}
            </div>
          </div>
        ));
      `,
      errors: [expectMissingPropTypeError19([2, 45, 2, 59])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 130 (7.37.5)',
      code: `
        import React, { memo } from 'react';
        const HeaderBalance = memo(({ cryptoCurrency }) => (
          <div className="header-balance">
            <div className="header-balance__balance">
              BTC
              {cryptoCurrency}
            </div>
          </div>
        ));
      `,
      errors: [expectMissingPropTypeError19([3, 39, 3, 53])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 131 (7.37.5)',
      code: `
        const HeaderBalance = Foo.memo(({ cryptoCurrency }) => (
          <div className="header-balance">
            <div className="header-balance__balance">
              BTC
              {cryptoCurrency}
            </div>
          </div>
        ));
      `,
      settings: { react: { pragma: 'Foo' } },
      errors: [expectMissingPropTypeError19([2, 43, 2, 57])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 132 (7.37.5)',
      code: `
        import Foo, { memo } from 'foo';
        const HeaderBalance = memo(({ cryptoCurrency }) => (
          <div className="header-balance">
            <div className="header-balance__balance">
              BTC
              {cryptoCurrency}
            </div>
          </div>
        ));
      `,
      settings: { react: { pragma: 'Foo' } },
      errors: [expectMissingPropTypeError19([3, 39, 3, 53])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 133 (7.37.5)',
      code: `
        const Label = React.forwardRef(({ text }, ref) => {
          return <div ref={ref}>{text}</div>;
        });
      `,
      errors: [expectMissingPropTypeError11([2, 43, 2, 47])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 134 (7.37.5)',
      code: `
        import React, { forwardRef } from 'react';
        const Label = forwardRef(({ text }, ref) => {
          return <div ref={ref}>{text}</div>;
        });
      `,
      errors: [expectMissingPropTypeError11([3, 37, 3, 41])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 135 (7.37.5)',
      code: `
        const Label = Foo.forwardRef(({ text }, ref) => {
          return <div ref={ref}>{text}</div>;
        });
      `,
      settings: { react: { pragma: 'Foo' } },
      errors: [expectMissingPropTypeError11([2, 41, 2, 45])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 136 (7.37.5)',
      code: `
        import Foo, { forwardRef } from 'foo';
        const Label = forwardRef(({ text }, ref) => {
          return <div ref={ref}>{text}</div>;
        });
      `,
      settings: { react: { pragma: 'Foo' } },
      errors: [expectMissingPropTypeError11([3, 37, 3, 41])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 137 (7.37.5)',
      code: `
        import PropTypes from 'prop-types';
        import React from 'react';

        const MyComponent = (props) => {
          switch (props.usedProp) {
            case 1:
              return (<div />);
            default:
              return <div />;
          }
        };

        export default MyComponent;
      `,
      errors: [
        {
          messageId: 'missingPropType',
          message: "'usedProp' is missing in props validation",
          line: 6,
          column: 25,
          endLine: 6,
          endColumn: 33,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 138 (7.37.5)',
      code: `
        export default class Controller extends React.Component {
          handleAdd = id => {
            this.setState((state, { name }) => {
                const item = this.buildItem(id);
            });
          };
        }
      `,
      errors: [expectMissingPropTypeError1([4, 37, 4, 41])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 139 (7.37.5)',
      code: `
        export default class extends React.Component {
          onSubmit = () => {
            this.setState((state, { a }) => {
              a.b.c();
              return null;
            });
          };

          render() {
            return null;
          }
        }
      `,
      errors: [
        {
          messageId: 'missingPropType',
          message: "'a' is missing in props validation",
          line: 4,
          column: 37,
          endLine: 4,
          endColumn: 38,
        },
        expectMissingPropTypeError6([5, 17, 5, 18]),
        expectMissingPropTypeError7([5, 19, 5, 20]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 140 (7.37.5)',
      code: `
        class Foo extends React.Component {
          constructor(props) {
            super(props);
            this.initialValues = {
              test: '',
            };
          }

          render = () => {
            return (
              <Component
                initialValues={this.props.initialValues || this.initialValues}
              >
                {formikProps => (
                  <Input {...formikProps} />
                )}
              </Component>
            );
          }
        }
      `,
      errors: [
        {
          messageId: 'missingPropType',
          message: "'initialValues' is missing in props validation",
          line: 13,
          column: 43,
          endLine: 13,
          endColumn: 56,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 141 (7.37.5)',
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
          messageId: 'missingPropType',
          message: "'notOne' is missing in props validation",
          line: 12,
          column: 42,
          endLine: 12,
          endColumn: 48,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 142 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream flow syntax (Property or signature expected.).',
      code: `
        type Props = {|
          firstname?: string
        |};

        function Hello({ firstname = 'John', lastname = 'Doe' }: Props = {}) {
          return <div>Hello {firstname} {lastname}</div>
        }
      `,
      errors: [expectMissingPropTypeError5([6, 46, 6, 62])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 143 (7.37.5)',
      skip: 'Upstream checks Flow prop annotations for prop-types; the native TypeScript annotation path reports 0 diagnostics instead of 1.',
      code: `
        type Props = {
          user: {
            name: {
              firstname: string,
              [string]: string
            }
          }
        };

        export function Bar(props: Props) {
          return (
            <div>
              {props.user}
              {props.user.name.firstname}
              {props.user.name.lastname}
              {props.user.age}
            </div>
          );
        }
      `,
      errors: [
        {
          messageId: 'missingPropType',
          message: "'user.age' is missing in props validation",
          line: 17,
          column: 27,
          endLine: 17,
          endColumn: 30,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 144 (7.37.5)',
      code: `
        const Label = React.memo(React.forwardRef(({ text }, ref) => {
          return <div ref={ref}>{text}</div>;
        }));
      `,
      errors: [expectMissingPropTypeError11([2, 54, 2, 58])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 145 (7.37.5)',
      code: `
        import React, { memo, forwardRef } from 'react';
        const Label = memo(forwardRef(({ text }, ref) => {
          return <div ref={ref}>{text}</div>;
        }));
      `,
      errors: [expectMissingPropTypeError11([3, 42, 3, 46])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 146 (7.37.5)',
      code: `
        import Foo, { memo, forwardRef } from 'foo';
        const Label = memo(forwardRef(({ text }, ref) => {
          return <div ref={ref}>{text}</div>;
        }));
      `,
      settings: { react: { pragma: 'Foo' } },
      errors: [expectMissingPropTypeError11([3, 42, 3, 46])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 147 (7.37.5)',
      code: `
        function Foo({
          foo: {
            bar: foo,
            baz
          },
        }) {
          return <p>{foo.reduce(() => 5)}</p>;
        }

        Foo.propTypes = {
          foo: PropTypes.shape({
            bar: PropTypes.arrayOf(PropTypes.string).isRequired,
          }).isRequired,
        };
      `,
      errors: [expectMissingPropTypeError18([5, 13, 5, 16])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 148 (7.37.5)',
      code: `
        const Foo = ({length, ordering}) => (
          length > 0 && (
            <Paginator items={ordering} pageSize={10} />
          )
        );
      `,
      errors: [
        {
          messageId: 'missingPropType',
          message: "'length' is missing in props validation",
          line: 2,
          column: 23,
          endLine: 2,
          endColumn: 29,
        },
        {
          messageId: 'missingPropType',
          message: "'ordering' is missing in props validation",
          line: 2,
          column: 31,
          endLine: 2,
          endColumn: 39,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 149 (7.37.5)',
      code: `
        const firstType = PropTypes.shape({
          id: PropTypes.number,
        });
        class ComponentX extends React.Component {
          static propTypes = {
            first: firstType.isRequired,
          };
          render() {
            return (
              <div>
                <div>Counter = {this.props.first.name}</div>
              </div>
            );
          }
        }
      `,
      errors: [expectMissingPropTypeError20([12, 50, 12, 54])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 150 (7.37.5)',
      code: `
        const firstType = PropTypes.shape({
          id: PropTypes.number,
        }).isRequired;
        class ComponentX extends React.Component {
          static propTypes = {
            first: firstType,
          };
          render() {
            return (
              <div>
                <div>Counter = {this.props.first.name}</div>
              </div>
            );
          }
        }
      `,
      errors: [expectMissingPropTypeError20([12, 50, 12, 54])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 151 (7.37.5)',
      code: `
        const firstType = PropTypes.shape({
          id: PropTypes.number,
        });
        class ComponentX extends React.Component {
          static propTypes = {
            first: firstType,
          };
          render() {
            return (
              <div>
                <div>Counter = {this.props.first.name}</div>
              </div>
            );
          }
        }
      `,
      errors: [expectMissingPropTypeError20([12, 50, 12, 54])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 152 (7.37.5)',
      code: `
        function Foo({
          foo: {
            bar: foo,
            baz
          },
        }) {
          return <p>{foo.reduce(() => 5)}</p>;
        }
        const fooType = PropTypes.shape({
          bar: PropTypes.arrayOf(PropTypes.string).isRequired,
        }).isRequired
        Foo.propTypes = {
          foo: fooType,
        };
      `,
      errors: [expectMissingPropTypeError18([5, 13, 5, 16])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 153 (7.37.5)',
      code: `
        function Foo({
          foo: {
            bar: foo,
            baz
          },
        }) {
          return <p>{foo.reduce(() => 5)}</p>;
        }
        const fooType = PropTypes.shape({
          bar: PropTypes.arrayOf(PropTypes.string).isRequired,
        })
        Foo.propTypes = {
          foo: fooType.isRequired,
        };
      `,
      errors: [expectMissingPropTypeError18([5, 13, 5, 16])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 154 (7.37.5)',
      code: `
        function Foo({
          foo: {
            bar: foo,
            baz
          },
        }) {
          return <p>{foo.reduce(() => 5)}</p>;
        }
        const fooType = PropTypes.shape({
          bar: PropTypes.arrayOf(PropTypes.string).isRequired,
        })
        Foo.propTypes = {
          foo: fooType,
        };
      `,
      errors: [expectMissingPropTypeError18([5, 13, 5, 16])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 155 (7.37.5)',
      code: `
        function Component(props) {
          return 0,
          <div>
            Hello, { props.name }!
          </div>
        }
      `,
      errors: [expectMissingPropTypeError1([5, 28, 5, 32])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 156 (7.37.5)',
      code: `
      const SideMenu = observer(
        ({ componentId }) => (
          <S.Container>
            <S.Head />
            <UserInfo />
            <Separator />
            <MenuList componentId={componentId} />
          </S.Container>
        ),
      );`,
      settings: { componentWrapperFunctions: ['observer'] },
      errors: [
        {
          messageId: 'missingPropType',
          message: "'componentId' is missing in props validation",
          line: 3,
          column: 12,
          endLine: 3,
          endColumn: 23,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 157 (7.37.5)',
      code: `
      const SideMenu = Mobx.observer(
        ({ id }) => (
          <S.Container>
            <S.Head />
            <UserInfo />
            <Separator />
            <MenuList componentId={id} />
          </S.Container>
        ),
      );
      `,
      settings: {
        componentWrapperFunctions: [{ property: 'observer', object: 'Mobx' }],
      },
      errors: [
        {
          messageId: 'missingPropType',
          message: "'id' is missing in props validation",
          line: 3,
          column: 12,
          endLine: 3,
          endColumn: 14,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 158 (7.37.5)',
      code: `
        interface Props {
        }
        const Hello = (props: Props) => {
          if (props.value) {
            return <div></div>;
          }
          return null;
        }
      `,
      errors: [
        {
          messageId: 'missingPropType',
          message: "'value' is missing in props validation",
          line: 5,
          column: 21,
          endLine: 5,
          endColumn: 26,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 159 (7.37.5)',
      code: `
        type User = {
          user: string;
        }

        type Props = User & {
        };

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
          messageId: 'missingPropType',
          message: "'userId' is missing in props validation",
          line: 10,
          column: 19,
          endLine: 10,
          endColumn: 25,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 160 (7.37.5)',
      code: `
        type User = {
        }

        type Props = User & {
          userId
        };

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
          messageId: 'missingPropType',
          message: "'user' is missing in props validation",
          line: 10,
          column: 27,
          endLine: 10,
          endColumn: 31,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 161 (7.37.5)',
      code: `
        type User = {
          user: string;
        }
        type UserProps = {
        }

        type Props = User & UserProps;

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
          messageId: 'missingPropType',
          message: "'userId' is missing in props validation",
          line: 11,
          column: 19,
          endLine: 11,
          endColumn: 25,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 162 (7.37.5)',
      code: `
        interface GenericProps {
          onClose: () => void
        }

        interface ImplementationProps extends GenericProps {
        }

        export const Implementation: FC<ImplementationProps> = (
          {
            onClick,
            onClose,
          }: ImplementationProps
        ) => (<div />)
      `,
      errors: [
        {
          messageId: 'missingPropType',
          message: "'onClick' is missing in props validation",
          line: 11,
          column: 13,
          endLine: 11,
          endColumn: 20,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 163 (7.37.5)',
      code: `
        const mapStateToProps = state => ({
        });

        interface BooksTable extends ReturnType<typeof mapStateToProps> {
          username: string;
        }

        const App = (props: BooksTable) => {
          props.books();
          return <div><span>{props.username}</span></div>;
        }
      `,
      errors: [
        {
          messageId: 'missingPropType',
          message: "'books' is missing in props validation",
          line: 10,
          column: 17,
          endLine: 10,
          endColumn: 22,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 164 (7.37.5)',
      code: `
        const mapStateToProps = state => ({
          books: state.books,
        });

        interface BooksTable extends ReturnType<typeof mapStateToProps> {
        }

        const App = (props: BooksTable) => {
          props.books();
          return <div><span>{props.username}</span></div>;
        }
      `,
      errors: [expectMissingPropTypeError21([11, 36, 11, 44])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 165 (7.37.5)',
      code: `
        type Event = {
            name: string;
            type: string;
        }

        interface UserEvent extends Event {
            UserId: string;
        }
        const App = (props: UserEvent) => {
          props.name();
          props.type;
          props.UserId;
          return <div><span>{props.dateCreated}</span></div>;
        }
      `,
      errors: [
        {
          messageId: 'missingPropType',
          message: "'dateCreated' is missing in props validation",
          line: 14,
          column: 36,
          endLine: 14,
          endColumn: 47,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 166 (7.37.5)',
      code: `
        function Zoo(props) {
          return (
            <>
              {props.foo.c}
            </>
          );
        }

        Zoo.propTypes = {
          foo: PropTypes.exact({
            a: PropTypes.number,
            b: PropTypes.number,
          }),
        };
      `,
      errors: [expectMissingPropTypeError22([5, 26, 5, 27])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 167 (7.37.5)',
      code: `
        function Zoo(props) {
          return (
            <>
              {props.foo.c}
            </>
          );
        }

        Zoo.propTypes = {
          foo: React.PropTypes.exact({
            a: PropTypes.number,
            b: PropTypes.number,
          }),
        };
      `,
      errors: [expectMissingPropTypeError22([5, 26, 5, 27])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 168 (7.37.5)',
      code: `
        function Zoo(props) {
          return (
            <>
              {props.foo.c}
            </>
          );
        }

        Zoo.propTypes = {
          foo: Foo.PropTypes.exact({
            a: PropTypes.number,
            b: PropTypes.number,
          }),
        };
      `,
      settings: { react: { pragma: 'Foo' } },
      errors: [expectMissingPropTypeError22([5, 26, 5, 27])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 169 (7.37.5)',
      code: `
        const Foo: JSX.Element = ({ bar }) => {
          return <div>{bar}</div>;
        }
      `,
      settings: { react: { pragma: 'Foo' } },
      errors: [expectMissingPropTypeError15([2, 37, 2, 40])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 170 (7.37.5)',
      code: `
        const Foo: JSX.Element = function foo ({ bar }) {
          return <div>{bar}</div>;
        }
      `,
      settings: { react: { pragma: 'Foo' } },
      errors: [expectMissingPropTypeError15([2, 50, 2, 53])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 171 (7.37.5)',
      code: `
        import React from 'react'

        const InputField = ({ type, ...restProps }) => {

          return(
            <input
              type={type}
              {...restProps}
            />
          )
        }

        export default InputField;
      `,
      errors: [
        {
          messageId: 'missingPropType',
          message: "'type' is missing in props validation",
          line: 4,
          column: 31,
          endLine: 4,
          endColumn: 35,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 172 (7.37.5)',
      code: `
        const Foo: JSX.Element = ({ bar = "" }) => {
          return <div>{bar}</div>;
        }
      `,
      errors: [expectMissingPropTypeError15([2, 37, 2, 40])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 173 (7.37.5)',
      code: `
        function Foo({ foo = "" }): JSX.Element {
          return <div>{foo}</div>;
        }
      `,
      errors: [expectMissingPropTypeError2([2, 24, 2, 27])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 174 (7.37.5)',
      code: `
      const Foo: JSX.Element = function foo ({ bar = "" }) {
        return <div>{bar}</div>;
      }
    `,
      errors: [expectMissingPropTypeError15([2, 48, 2, 51])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 175 (7.37.5)',
      code: `
      function Foo({ bar = "" as string }): JSX.Element {
        return <div>{bar}</div>;
      }
    `,
      errors: [expectMissingPropTypeError15([2, 22, 2, 25])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 176 (7.37.5)',
      code: `
        import React from 'react';
        interface PersonProps {
            test: string;
        }
        const Person: React.FunctionComponent<PersonProps> = (props): React.ReactElement => (
            <div>{props.username}</div>
        );
      `,
      errors: [expectMissingPropTypeError21([7, 25, 7, 33])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 177 (7.37.5)',
      code: `
        import React from 'react';
        interface PersonProps {
            username: string;
        }
        const Person: React.FunctionComponent<PersonProps> = (props): React.ReactElement => (
            <div>{props.test}</div>
        );
      `,
      errors: [expectMissingPropTypeError14([7, 25, 7, 29])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 178 (7.37.5)',
      code: `
        import React from "react";

        const returnTypeProp = (someProp: any) => ({ someProp });

        const SomeComponent: React.FunctionComponent<
          ReturnType<typeof returnTypeProp>
        > = ({ someIncorrectProp }) => {
          return <div>{someProp}</div>;
        };
      `,
      errors: [
        {
          messageId: 'missingPropType',
          message: "'someIncorrectProp' is missing in props validation",
          line: 8,
          column: 16,
          endLine: 8,
          endColumn: 33,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 179 (7.37.5)',
      code: `
        import React from 'react';
        interface PersonProps {
            username: string;
        }
        const Person: FunctionComponent<PersonProps> = (props): React.ReactElement => (
            <div>{props.test}</div>
        );
      `,
      errors: [expectMissingPropTypeError14([7, 25, 7, 29])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 180 (7.37.5)',
      code: `
        interface PersonProps {
            username: string;
        }
        const Person: X.FC<PersonProps> = (props): React.ReactElement => (
            <div>{props.username}</div>
        );
      `,
      errors: [expectMissingPropTypeError21([6, 25, 6, 33])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 181 (7.37.5)',
      code: `
        interface PersonProps {
            username: string;
        }
        const Person: FC<PersonProps> = (props): React.ReactElement => (
            <div>{props.username}</div>
        );
      `,
      errors: [expectMissingPropTypeError21([6, 25, 6, 33])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 182 (7.37.5)',
      code: `
        type PersonProps<T> = {
            x: T;
        }
        const Person = (props: PersonProps<string>): React.ReactElement => (
            <div>{props.username}</div>
        );
      `,
      errors: [expectMissingPropTypeError21([6, 25, 6, 33])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 183 (7.37.5)',
      code: `
        import React, { VoidFunctionComponent } from 'react'

        interface Props {
        age: number
        }
        const Hello: VoidFunctionComponent<Props> = function Hello(props) {
        const { test } = props;

        return <div>Hello {name}</div>;
        }
      `,
      errors: [expectMissingPropTypeError14([8, 17, 8, 21])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 184 (7.37.5)',
      code: `
        import React, { VFC } from 'react'

        interface Props {
        age: number
        }
        const Hello: VFC<Props> = function Hello(props) {
        const { test } = props;

        return <div>Hello {name}</div>;
        }
      `,
      errors: [expectMissingPropTypeError14([8, 17, 8, 21])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 185 (7.37.5)',
      code: `
        import React from 'react'

        interface Props {
        age: number
        }
        const Hello: React.VoidFunctionComponent<Props> = function Hello(props) {
        const { test } = props;

        return <div>Hello {name}</div>;
        }
      `,
      errors: [expectMissingPropTypeError14([8, 17, 8, 21])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 186 (7.37.5)',
      code: `
        import React from 'react'

        interface Props {
        age: number
        }
        const Hello: React.VFC<Props> = function Hello(props) {
        const { test } = props;

        return <div>Hello {name}</div>;
        }
      `,
      errors: [expectMissingPropTypeError14([8, 17, 8, 21])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 187 (7.37.5)',
      code: `
        import React from 'react'

        type IfooProps = { e: string };
          const Foo: React.ForwardRefRenderFunction<HTMLDivElement, IfooProps> = function Foo (props, ref) {
          const { name } = props;
          return  <div ref={ref}>{name}</div>;
        };
      `,
      errors: [expectMissingPropTypeError1([6, 19, 6, 23])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 188 (7.37.5)',
      code: `
        import React from 'react'
        type IfooProps = { k: string, a: number }
        const Foo= React.forwardRef<HTMLDivElement, IfooProps>((props, ref) => {
          return  <div ref={ref}>{props.l}</div>;
        });
      `,
      errors: [expectMissingPropTypeError23([5, 41, 5, 42])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 189 (7.37.5)',
      code: `
        import React from 'react'

        type IfooProps = { e: string };
        const Foo= React.forwardRef<HTMLDivElement, IfooProps>(function Foo(props, ref) {
          const { l } = props;
          return  <div ref={ref}>hello</div>;
        });
      `,
      errors: [expectMissingPropTypeError23([6, 19, 6, 20])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 190 (7.37.5)',
      code: `
        import React, { forwardRef } from 'react'

        type IfooProps = { e: string };
        const Foo= forwardRef<HTMLDivElement, IfooProps>(function Foo(props, ref) {
          const { l } = props;
          return  <div ref={ref}>hello</div>;
        });
      `,
      errors: [expectMissingPropTypeError23([6, 19, 6, 20])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 191 (7.37.5)',
      code: `
        import React, { forwardRef } from "react";

        export type Props = { children: React.ReactNode; type: "submit" | "button" };

        export const FancyButton = forwardRef<HTMLButtonElement, Props>((props, ref) => (
          <button type="button" ref={ref} className="MyClassName" type={props.nonExistent}>
            {props.children}
          </button>
        ));
      `,
      errors: [expectMissingPropTypeError24([7, 79, 7, 90])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 192 (7.37.5)',
      code: `
        import React, { forwardRef } from "react";

        export interface IProps { children: React.ReactNode; type: "submit" | "button" };

        export const FancyButton = forwardRef<HTMLButtonElement, IProps>((props, ref) => (
          <button type="button" ref={ref} className="MyClassName" type={props.nonExistent}>
            {props.children}
          </button>
        ));
      `,
      errors: [expectMissingPropTypeError24([7, 79, 7, 90])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 193 (7.37.5)',
      code: `
        import React from 'react';

        export interface PersonProps {
            username: string;
        }
        const Person: React.FC<PersonProps> = (props): React.ReactElement => (
            <div>{props.nonExistent}</div>
        );
      `,
      errors: [expectMissingPropTypeError24([8, 25, 8, 36])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 194 (7.37.5)',
      code: `
        import React, { FC } from 'react';

        export interface PersonProps {
            username: string;
        }
        const Person: FC<PersonProps> = (props): React.ReactElement => (
            <div>{props.nonExistent}</div>
        );
      `,
      errors: [expectMissingPropTypeError24([8, 25, 8, 36])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 195 (7.37.5)',
      code: `
        import React, { FC as X } from 'react';

        export interface PersonProps {
            username: string;
        }
        const Person: X<PersonProps> = (props): React.ReactElement => (
            <div>{props.nonExistent}</div>
        );
      `,
      errors: [expectMissingPropTypeError24([8, 25, 8, 36])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 196 (7.37.5)',
      code: `
        import React, { ForwardRefRenderFunction  as X } from 'react'

        export type IfooProps = { e: string };
        const Foo: X<HTMLDivElement, IfooProps> = function Foo (props, ref) {
          const { nonExistent } = props;
          return  <div ref={ref}>hello</div>;
        };
      `,
      errors: [expectMissingPropTypeError24([6, 19, 6, 30])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 197 (7.37.5)',
      code: `
        import React from 'react';

        export interface PersonProps {
            username: string;
        }
        const Person: React.VoidFunctionComponent<PersonProps> = (props): React.ReactElement => (
            <div>{props.nonExistent}</div>
        );
      `,
      errors: [expectMissingPropTypeError24([8, 25, 8, 36])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 198 (7.37.5)',
      code: `
        import React from 'react';

        export interface PersonProps {
            username: string;
        }
        const Person: React.VFC<PersonProps> = (props): React.ReactElement => (
            <div>{props.nonExistent}</div>
        );
      `,
      errors: [expectMissingPropTypeError24([8, 25, 8, 36])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 199 (7.37.5)',
      code: `
        const Foo = ({ foo }) => {
          return <SomeJSX />;
        }
      `,
      errors: [expectMissingPropTypeError2([2, 24, 2, 27])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 200 (7.37.5)',
      code: `
        const Foo = ({ foo }) => {
          const returnValue = <SomeJSX />;
          return returnValue;
        }
      `,
      errors: [expectMissingPropTypeError2([2, 24, 2, 27])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 201 (7.37.5)',
      code: `
        function _EventsList({ prop_ }) {
            return (
                <div>
                    {prop_.events.map((event) => (
                        <Event key={event.id} eventId={event.id} />
                    ))}
                </div>
            );
        }

        function $EventsList({ prop$ }) {
          return (
              <div>
                  {prop$.events.map((event) => (
                      <Event key={event.id} eventId={event.id} />
                  ))}
              </div>
          );
      }
      `,
      errors: [
        {
          messageId: 'missingPropType',
          message: "'prop_' is missing in props validation",
          line: 2,
          column: 32,
          endLine: 2,
          endColumn: 37,
        },
        {
          messageId: 'missingPropType',
          message: "'prop_.events' is missing in props validation",
          line: 5,
          column: 28,
          endLine: 5,
          endColumn: 34,
        },
        {
          messageId: 'missingPropType',
          message: "'prop_.events.map' is missing in props validation",
          line: 5,
          column: 35,
          endLine: 5,
          endColumn: 38,
        },
        {
          messageId: 'missingPropType',
          message: "'prop$' is missing in props validation",
          line: 12,
          column: 32,
          endLine: 12,
          endColumn: 37,
        },
        {
          messageId: 'missingPropType',
          message: "'prop$.events' is missing in props validation",
          line: 15,
          column: 26,
          endLine: 15,
          endColumn: 32,
        },
        {
          messageId: 'missingPropType',
          message: "'prop$.events.map' is missing in props validation",
          line: 15,
          column: 33,
          endLine: 15,
          endColumn: 36,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
  ],
});

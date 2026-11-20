import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, upstreamModuleOptions } from '../config';

function expectTypoStaticClassPropError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'typoStaticClassProp',
    message: 'Typo in static class property declaration',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectTypoPropTypeError2(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'typoPropType',
    message: 'Typo in declared prop type: Number',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectTypoPropTypeChainError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'typoPropTypeChain',
    message: 'Typo in prop type chain qualifier: isrequired',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectTypoPropTypeError4(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'typoPropType',
    message: 'Typo in declared prop type: bools',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectTypoPropTypeError5(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'typoPropType',
    message: 'Typo in declared prop type: Array',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectTypoPropTypeError6(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'typoPropType',
    message: 'Typo in declared prop type: function',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectTypoPropTypeError7(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'typoPropType',
    message: 'Typo in declared prop type: objectof',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectTypoPropDeclarationError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'typoPropDeclaration',
    message: 'Typo in property declaration',
    line,
    column,
    endLine,
    endColumn,
  };
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('no-typos', {
  valid: [
    // ---- Upstream: non-component classes / functions are ignored ----
    {
      code: `
          import createReactClass from 'create-react-class'
          function hello (extra = {}) {
            return createReactClass({
              noteType: 'hello',
              renderItem () {
                return null
              },
              ...extra
            })
          }
        `,
    },
    {
      code: `
          class First {
            static PropTypes = {key: "myValue"};
            static ContextTypes = {key: "myValue"};
            static ChildContextTypes = {key: "myValue"};
            static DefaultProps = {key: "myValue"};
          }
        `,
    },
    {
      code: `
          class First {}
          First.PropTypes = {key: "myValue"};
          First.ContextTypes = {key: "myValue"};
          First.ChildContextTypes = {key: "myValue"};
          First.DefaultProps = {key: "myValue"};
        `,
    },
    // ---- Upstream: exact casing on static class properties ----
    {
      code: `
          class First extends React.Component {
            static propTypes = {key: "myValue"};
            static contextTypes = {key: "myValue"};
            static childContextTypes = {key: "myValue"};
            static defaultProps = {key: "myValue"};
          }
        `,
    },
    {
      code: `
          class First extends React.Component {}
          First.propTypes = {key: "myValue"};
          First.contextTypes = {key: "myValue"};
          First.childContextTypes = {key: "myValue"};
          First.defaultProps = {key: "myValue"};
        `,
    },
    // ---- Upstream: non-static members of non-component class are fine ----
    {
      code: `
          class MyClass {
            propTypes = {key: "myValue"};
            contextTypes = {key: "myValue"};
            childContextTypes = {key: "myValue"};
            defaultProps = {key: "myValue"};
          }
        `,
    },
    {
      code: `
          class MyClass {
            PropTypes = {key: "myValue"};
            ContextTypes = {key: "myValue"};
            ChildContextTypes = {key: "myValue"};
            DefaultProps = {key: "myValue"};
          }
        `,
    },
    {
      code: `
          class MyClass {
            proptypes = {key: "myValue"};
            contexttypes = {key: "myValue"};
            childcontextypes = {key: "myValue"};
            defaultprops = {key: "myValue"};
          }
        `,
    },
    {
      code: `
          class MyClass {
            static PropTypes() {};
            static ContextTypes() {};
            static ChildContextTypes() {};
            static DefaultProps() {};
          }
        `,
    },
    {
      code: `
          class MyClass {
            static proptypes() {};
            static contexttypes() {};
            static childcontexttypes() {};
            static defaultprops() {};
          }
        `,
    },
    {
      code: `
          class MyClass {}
          MyClass.prototype.PropTypes = function() {};
          MyClass.prototype.ContextTypes = function() {};
          MyClass.prototype.ChildContextTypes = function() {};
          MyClass.prototype.DefaultProps = function() {};
        `,
    },
    {
      code: `
          class MyClass {}
          MyClass.PropTypes = function() {};
          MyClass.ContextTypes = function() {};
          MyClass.ChildContextTypes = function() {};
          MyClass.DefaultProps = function() {};
        `,
    },
    {
      code: `
          function MyRandomFunction() {}
          MyRandomFunction.PropTypes = {};
          MyRandomFunction.ContextTypes = {};
          MyRandomFunction.ChildContextTypes = {};
          MyRandomFunction.DefaultProps = {};
        `,
    },
    // ---- Upstream: unsupported dynamic computed keys (bracket notation) ----
    {
      code: `
          class First extends React.Component {}
          First["prop" + "Types"] = {};
          First["context" + "Types"] = {};
          First["childContext" + "Types"] = {};
          First["default" + "Props"] = {};
        `,
    },
    {
      code: `
          class First extends React.Component {}
          First["PROP" + "TYPES"] = {};
          First["CONTEXT" + "TYPES"] = {};
          First["CHILDCONTEXT" + "TYPES"] = {};
          First["DEFAULT" + "PROPS"] = {};
        `,
    },
    {
      code: `
          const propTypes = "PROPTYPES"
          const contextTypes = "CONTEXTTYPES"
          const childContextTypes = "CHILDCONTEXTTYPES"
          const defaultProps = "DEFAULTPROPS"

          class First extends React.Component {}
          First[propTypes] = {};
          First[contextTypes] = {};
          First[childContextTypes] = {};
          First[defaultProps] = {};
        `,
    },
    // ---- Upstream: well-cased lifecycle methods ----
    {
      code: `
          class Hello extends React.Component {
            static getDerivedStateFromProps() { }
            componentWillMount() { }
            componentDidMount() { }
            componentWillReceiveProps() { }
            shouldComponentUpdate() { }
            componentWillUpdate() { }
            componentDidUpdate() { }
            componentWillUnmount() { }
            render() {
              return <div>Hello {this.props.name}</div>;
            }
          }
        `,
    },
    {
      code: `
          class Hello extends React.Component {
            "componentDidMount"() { }
            "my-method"() { }
          }
        `,
    },
    {
      code: `
          class MyClass {
            componentWillMount() { }
            componentDidMount() { }
            componentWillReceiveProps() { }
            shouldComponentUpdate() { }
            componentWillUpdate() { }
            componentDidUpdate() { }
            componentWillUnmount() { }
            render() { }
          }
        `,
    },
    {
      code: `
          class MyClass {
            componentwillmount() { }
            componentdidmount() { }
            componentwillreceiveprops() { }
            shouldcomponentupdate() { }
            componentwillupdate() { }
            componentdidupdate() { }
            componentwillUnmount() { }
            render() { }
          }
        `,
    },
    {
      code: `
          class MyClass {
            Componentwillmount() { }
            Componentdidmount() { }
            Componentwillreceiveprops() { }
            Shouldcomponentupdate() { }
            Componentwillupdate() { }
            Componentdidupdate() { }
            ComponentwillUnmount() { }
            Render() { }
          }
        `,
    },
    // ---- Upstream: issue #1353 — unrelated .bind ----
    {
      code: `
          function test(b) {
            return a.bind(b);
          }
          function a() {}
        `,
    },
    // ---- Upstream: well-formed PropTypes usage ----
    {
      code: `
          import PropTypes from "prop-types";
          class Component extends React.Component {};
          Component.propTypes = {
            a: PropTypes.number.isRequired
          }
        `,
    },
    {
      code: `
          import PropTypes from "prop-types";
          class Component extends React.Component {};
          Component.propTypes = {
            e: PropTypes.shape({
              ea: PropTypes.string,
            })
          }
        `,
    },
    {
      code: `
          import PropTypes from "prop-types";
          class Component extends React.Component {};
          Component.propTypes = {
            a: PropTypes.string,
            b: PropTypes.string.isRequired,
            c: PropTypes.shape({
              d: PropTypes.string,
              e: PropTypes.number.isRequired,
            }).isRequired
          }
        `,
    },
    {
      code: `
          import PropTypes from "prop-types";
          class Component extends React.Component {};
          Component.propTypes = {
            a: PropTypes.oneOfType([
              PropTypes.string,
              PropTypes.number
            ])
          }
        `,
    },
    {
      code: `
          import PropTypes from "prop-types";
          class Component extends React.Component {};
          Component.propTypes = {
            a: PropTypes.oneOf([
              'hello',
              'hi'
            ])
          }
        `,
    },
    {
      code: `
          import PropTypes from "prop-types";
          class Component extends React.Component {};
          Component.childContextTypes = {
            a: PropTypes.string,
            b: PropTypes.string.isRequired,
            c: PropTypes.shape({
              d: PropTypes.string,
              e: PropTypes.number.isRequired,
            }).isRequired
          }
        `,
    },
    {
      code: `
          import PropTypes from "prop-types";
          class Component extends React.Component {};
          Component.contextTypes = {
            a: PropTypes.string,
            b: PropTypes.string.isRequired,
            c: PropTypes.shape({
              d: PropTypes.string,
              e: PropTypes.number.isRequired,
            }).isRequired
          }
        `,
    },
    // ---- Upstream: external types alongside prop-types ----
    {
      code: `
          import PropTypes from 'prop-types'
          import * as MyPropTypes from 'lib/my-prop-types'
          class Component extends React.Component {};
          Component.propTypes = {
            a: PropTypes.string,
            b: MyPropTypes.MYSTRING,
            c: MyPropTypes.MYSTRING.isRequired,
          }
        `,
    },
    {
      code: `
          import PropTypes from "prop-types"
          import * as MyPropTypes from 'lib/my-prop-types'
          class Component extends React.Component {};
          Component.propTypes = {
            b: PropTypes.string,
            a: MyPropTypes.MYSTRING,
          }
        `,
    },
    {
      code: `
          import CustomReact from "react"
          class Component extends React.Component {};
          Component.propTypes = {
            b: CustomReact.PropTypes.string,
          }
        `,
    },
    // ---- Upstream: absent arg to PropTypes.shape must not crash ----
    {
      code: `
          class Component extends React.Component {};
          Component.propTypes = {
            a: PropTypes.shape(),
          };
          Component.contextTypes = {
            a: PropTypes.shape(),
          };
        `,
    },
    // ---- Upstream: unrelated patterns ----
    {
      code: `
          const fn = (err, res) => {
            const { body: data = {} } = { ...res };
            data.time = data.time || {};
          };
        `,
    },
    {
      code: `
          class Component extends React.Component {};
          Component.propTypes = {
            b: string.isRequired,
            c: PropTypes.shape({
              d: number.isRequired,
            }).isRequired
          }
        `,
    },
    // ---- Upstream: createReactClass with PropTypes ----
    {
      code: `
          import React from 'react';
          import PropTypes from 'prop-types';
          const Component = React.createReactClass({
            propTypes: {
              a: PropTypes.string.isRequired,
              b: PropTypes.shape({
                c: PropTypes.number
              }).isRequired
            }
          });
        `,
    },
    {
      code: `
          import React from 'react';
          import PropTypes from 'prop-types';
          const Component = React.createReactClass({
            childContextTypes: {
              a: PropTypes.bool,
              b: PropTypes.array,
              c: PropTypes.func,
              d: PropTypes.object,
            }
          });
        `,
    },
    {
      code: `
          import React from 'react';
          const Component = React.createReactClass({
            propTypes: {},
            childContextTypes: {},
            contextTypes: {},
            componentWillMount() { },
            componentDidMount() { },
            componentWillReceiveProps() { },
            shouldComponentUpdate() { },
            componentWillUpdate() { },
            componentDidUpdate() { },
            componentWillUnmount() { },
            render() {
              return <div>Hello {this.props.name}</div>;
            }
          });
        `,
    },
    // ---- Upstream: destructured named imports from prop-types ----
    {
      code: `
          import { string, element } from "prop-types";

          class Sample extends React.Component {
            render() { return null; }
          }

          Sample.propTypes = {
            title: string.isRequired,
            body: element.isRequired
          };
        `,
    },
    // ---- Upstream: computed key with PropertyAccessExpression base ----
    {
      code: `
          import React from 'react';

          const A = { B: 'C' };

          export default class MyComponent extends React.Component {
            [A.B] () {
              return null
            }
          }
        `,
    },
    // ---- Upstream: React.forwardRef / styled-components escape hatches ----
    {
      code: `
          const MyComponent = React.forwardRef((props, ref) => <div />);
          MyComponent.defaultProps = { value: "" };
        `,
    },
    {
      code: `
          import styled from "styled-components";

          const MyComponent = styled.div;
          MyComponent.defaultProps = { value: "" };
        `,
    },
    // ---- Upstream: private method should not trigger lifecycle typo ----
    {
      code: `
          class Editor extends React.Component {
              #somethingPrivate() {
                // ...
              }

              render() {
              const { value = '' } = this.props;

              return (
                <textarea>
                  {value}
                </textarea>
              );
            }
          }
        `,
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 0 (7.37.5)',
      code: `
          import createReactClass from 'create-react-class'
          function hello (extra = {}) {
            return createReactClass({
              noteType: 'hello',
              renderItem () {
                return null
              },
              ...extra
            })
          }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 1 (7.37.5)',
      code: `
        class First {
          static PropTypes = {key: "myValue"};
          static ContextTypes = {key: "myValue"};
          static ChildContextTypes = {key: "myValue"};
          static DefaultProps = {key: "myValue"};
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 2 (7.37.5)',
      code: `
        class First {}
        First.PropTypes = {key: "myValue"};
        First.ContextTypes = {key: "myValue"};
        First.ChildContextTypes = {key: "myValue"};
        First.DefaultProps = {key: "myValue"};
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 3 (7.37.5)',
      code: `
        class First extends React.Component {
          static propTypes = {key: "myValue"};
          static contextTypes = {key: "myValue"};
          static childContextTypes = {key: "myValue"};
          static defaultProps = {key: "myValue"};
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 4 (7.37.5)',
      code: `
        class First extends React.Component {}
        First.propTypes = {key: "myValue"};
        First.contextTypes = {key: "myValue"};
        First.childContextTypes = {key: "myValue"};
        First.defaultProps = {key: "myValue"};
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 5 (7.37.5)',
      code: `
        class MyClass {
          propTypes = {key: "myValue"};
          contextTypes = {key: "myValue"};
          childContextTypes = {key: "myValue"};
          defaultProps = {key: "myValue"};
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 6 (7.37.5)',
      code: `
        class MyClass {
          PropTypes = {key: "myValue"};
          ContextTypes = {key: "myValue"};
          ChildContextTypes = {key: "myValue"};
          DefaultProps = {key: "myValue"};
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 7 (7.37.5)',
      code: `
        class MyClass {
          proptypes = {key: "myValue"};
          contexttypes = {key: "myValue"};
          childcontextypes = {key: "myValue"};
          defaultprops = {key: "myValue"};
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 8 (7.37.5)',
      code: `
        class MyClass {
          static PropTypes() {};
          static ContextTypes() {};
          static ChildContextTypes() {};
          static DefaultProps() {};
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 9 (7.37.5)',
      code: `
        class MyClass {
          static proptypes() {};
          static contexttypes() {};
          static childcontexttypes() {};
          static defaultprops() {};
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 10 (7.37.5)',
      code: `
        class MyClass {}
        MyClass.prototype.PropTypes = function() {};
        MyClass.prototype.ContextTypes = function() {};
        MyClass.prototype.ChildContextTypes = function() {};
        MyClass.prototype.DefaultProps = function() {};
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 11 (7.37.5)',
      code: `
        class MyClass {}
        MyClass.PropTypes = function() {};
        MyClass.ContextTypes = function() {};
        MyClass.ChildContextTypes = function() {};
        MyClass.DefaultProps = function() {};
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 12 (7.37.5)',
      code: `
        function MyRandomFunction() {}
        MyRandomFunction.PropTypes = {};
        MyRandomFunction.ContextTypes = {};
        MyRandomFunction.ChildContextTypes = {};
        MyRandomFunction.DefaultProps = {};
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 13 (7.37.5)',
      code: `
        class First extends React.Component {}
        First["prop" + "Types"] = {};
        First["context" + "Types"] = {};
        First["childContext" + "Types"] = {};
        First["default" + "Props"] = {};
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 14 (7.37.5)',
      code: `
        class First extends React.Component {}
        First["PROP" + "TYPES"] = {};
        First["CONTEXT" + "TYPES"] = {};
        First["CHILDCONTEXT" + "TYPES"] = {};
        First["DEFAULT" + "PROPS"] = {};
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 15 (7.37.5)',
      code: `
        const propTypes = "PROPTYPES"
        const contextTypes = "CONTEXTTYPES"
        const childContextTypes = "CHILDCONTEXTTYPES"
        const defaultProps = "DEFAULTPROPS"

        class First extends React.Component {}
        First[propTypes] = {};
        First[contextTypes] = {};
        First[childContextTypes] = {};
        First[defaultProps] = {};
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 16 (7.37.5)',
      code: `
        class Hello extends React.Component {
          static getDerivedStateFromProps() { }
          componentWillMount() { }
          componentDidMount() { }
          componentWillReceiveProps() { }
          shouldComponentUpdate() { }
          componentWillUpdate() { }
          componentDidUpdate() { }
          componentWillUnmount() { }
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
          "componentDidMount"() { }
          "my-method"() { }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 18 (7.37.5)',
      code: `
        class MyClass {
          componentWillMount() { }
          componentDidMount() { }
          componentWillReceiveProps() { }
          shouldComponentUpdate() { }
          componentWillUpdate() { }
          componentDidUpdate() { }
          componentWillUnmount() { }
          render() { }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 19 (7.37.5)',
      code: `
        class MyClass {
          componentwillmount() { }
          componentdidmount() { }
          componentwillreceiveprops() { }
          shouldcomponentupdate() { }
          componentwillupdate() { }
          componentdidupdate() { }
          componentwillUnmount() { }
          render() { }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 20 (7.37.5)',
      code: `
        class MyClass {
          Componentwillmount() { }
          Componentdidmount() { }
          Componentwillreceiveprops() { }
          Shouldcomponentupdate() { }
          Componentwillupdate() { }
          Componentdidupdate() { }
          ComponentwillUnmount() { }
          Render() { }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 21 (7.37.5)',
      code: `
        function test(b) {
          return a.bind(b);
        }
        function a() {}
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 22 (7.37.5)',
      code: `
        import PropTypes from "prop-types";
        class Component extends React.Component {};
        Component.propTypes = {
          a: PropTypes.number.isRequired
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 23 (7.37.5)',
      code: `
        import PropTypes from "prop-types";
        class Component extends React.Component {};
        Component.propTypes = {
          e: PropTypes.shape({
            ea: PropTypes.string,
          })
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 24 (7.37.5)',
      code: `
        import PropTypes from "prop-types";
        class Component extends React.Component {};
        Component.propTypes = {
          a: PropTypes.string,
          b: PropTypes.string.isRequired,
          c: PropTypes.shape({
            d: PropTypes.string,
            e: PropTypes.number.isRequired,
          }).isRequired
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 25 (7.37.5)',
      code: `
        import PropTypes from "prop-types";
        class Component extends React.Component {};
        Component.propTypes = {
          a: PropTypes.oneOfType([
            PropTypes.string,
            PropTypes.number
          ])
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 26 (7.37.5)',
      code: `
        import PropTypes from "prop-types";
        class Component extends React.Component {};
        Component.propTypes = {
          a: PropTypes.oneOf([
            'hello',
            'hi'
          ])
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 27 (7.37.5)',
      code: `
        import PropTypes from "prop-types";
        class Component extends React.Component {};
        Component.childContextTypes = {
          a: PropTypes.string,
          b: PropTypes.string.isRequired,
          c: PropTypes.shape({
            d: PropTypes.string,
            e: PropTypes.number.isRequired,
          }).isRequired
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 28 (7.37.5)',
      code: `
        import PropTypes from "prop-types";
        class Component extends React.Component {};
        Component.contextTypes = {
          a: PropTypes.string,
          b: PropTypes.string.isRequired,
          c: PropTypes.shape({
            d: PropTypes.string,
            e: PropTypes.number.isRequired,
          }).isRequired
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 29 (7.37.5)',
      code: `
        import PropTypes from 'prop-types'
        import * as MyPropTypes from 'lib/my-prop-types'
        class Component extends React.Component {};
        Component.propTypes = {
          a: PropTypes.string,
          b: MyPropTypes.MYSTRING,
          c: MyPropTypes.MYSTRING.isRequired,
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 30 (7.37.5)',
      code: `
        import PropTypes from "prop-types"
        import * as MyPropTypes from 'lib/my-prop-types'
        class Component extends React.Component {};
        Component.propTypes = {
          b: PropTypes.string,
          a: MyPropTypes.MYSTRING,
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 31 (7.37.5)',
      code: `
        import CustomReact from "react"
        class Component extends React.Component {};
        Component.propTypes = {
          b: CustomReact.PropTypes.string,
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 32 (7.37.5)',
      code: `
        class Component extends React.Component {};
        Component.propTypes = {
          a: PropTypes.shape(),
        };
        Component.contextTypes = {
          a: PropTypes.shape(),
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 33 (7.37.5)',
      code: `
        const fn = (err, res) => {
          const { body: data = {} } = { ...res };
          data.time = data.time || {};
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 34 (7.37.5)',
      code: `
        class Component extends React.Component {};
        Component.propTypes = {
          b: string.isRequired,
          c: PropTypes.shape({
            d: number.isRequired,
          }).isRequired
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 35 (7.37.5)',
      code: `
        import React from 'react';
        import PropTypes from 'prop-types';
        const Component = React.createReactClass({
          propTypes: {
            a: PropTypes.string.isRequired,
            b: PropTypes.shape({
              c: PropTypes.number
            }).isRequired
          }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 36 (7.37.5)',
      code: `
        import React from 'react';
        import PropTypes from 'prop-types';
        const Component = React.createReactClass({
          childContextTypes: {
            a: PropTypes.bool,
            b: PropTypes.array,
            c: PropTypes.func,
            d: PropTypes.object,
          }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 37 (7.37.5)',
      code: `
        import React from 'react';
        const Component = React.createReactClass({
          propTypes: {},
          childContextTypes: {},
          contextTypes: {},
          componentWillMount() { },
          componentDidMount() { },
          componentWillReceiveProps() { },
          shouldComponentUpdate() { },
          componentWillUpdate() { },
          componentDidUpdate() { },
          componentWillUnmount() { },
          render() {
            return <div>Hello {this.props.name}</div>;
          }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 38 (7.37.5)',
      code: `
        import { string, element } from "prop-types";

        class Sample extends React.Component {
          render() { return null; }
        }

        Sample.propTypes = {
          title: string.isRequired,
          body: element.isRequired
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 39 (7.37.5)',
      code: `
        import React from 'react';

        const A = { B: 'C' };

        export default class MyComponent extends React.Component {
          [A.B] () {
            return null
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 40 (7.37.5)',
      code: `
        const MyComponent = React.forwardRef((props, ref) => <div />);
        MyComponent.defaultProps = { value: "" };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 41 (7.37.5)',
      code: `
        import styled from "styled-components";

        const MyComponent = styled.div;
        MyComponent.defaultProps = { value: "" };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 42 (7.37.5)',
      code: `
        class Editor extends React.Component {
            #somethingPrivate() {
              // ...
            }

            render() {
            const { value = '' } = this.props;

            return (
              <textarea>
                {value}
              </textarea>
            );
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
  ],
  invalid: [
    // ---- Upstream: typoStaticClassProp — PropTypes variants ----
    {
      code: `
        class Component extends React.Component {
          static PropTypes = {};
        }
      `,
      errors: [expectTypoStaticClassPropError([3, 18, 3, 27])],
    },
    {
      code: `
        class Component extends React.Component {}
        Component.PropTypes = {}
      `,
      errors: [expectTypoStaticClassPropError([3, 19, 3, 28])],
    },
    {
      code: `
        function MyComponent() { return (<div>{this.props.myProp}</div>) }
        MyComponent.PropTypes = {}
      `,
      errors: [expectTypoStaticClassPropError([3, 21, 3, 30])],
    },
    {
      code: `
        class Component extends React.Component {
          static proptypes = {};
        }
      `,
      errors: [expectTypoStaticClassPropError([3, 18, 3, 27])],
    },
    {
      code: `
        class Component extends React.Component {}
        Component.proptypes = {}
      `,
      errors: [expectTypoStaticClassPropError([3, 19, 3, 28])],
    },
    {
      code: `
        function MyComponent() { return (<div>{this.props.myProp}</div>) }
        MyComponent.proptypes = {}
      `,
      errors: [expectTypoStaticClassPropError([3, 21, 3, 30])],
    },
    // ---- Upstream: typoStaticClassProp — ContextTypes variants ----
    {
      code: `
        class Component extends React.Component {
          static ContextTypes = {};
        }
      `,
      errors: [expectTypoStaticClassPropError([3, 18, 3, 30])],
    },
    {
      code: `
        class Component extends React.Component {}
        Component.ContextTypes = {}
      `,
      errors: [expectTypoStaticClassPropError([3, 19, 3, 31])],
    },
    {
      code: `
        function MyComponent() { return (<div>{this.props.myProp}</div>) }
        MyComponent.ContextTypes = {}
      `,
      errors: [expectTypoStaticClassPropError([3, 21, 3, 33])],
    },
    {
      code: `
        class Component extends React.Component {
          static contexttypes = {};
        }
      `,
      errors: [expectTypoStaticClassPropError([3, 18, 3, 30])],
    },
    {
      code: `
        class Component extends React.Component {}
        Component.contexttypes = {}
      `,
      errors: [expectTypoStaticClassPropError([3, 19, 3, 31])],
    },
    {
      code: `
        function MyComponent() { return (<div>{this.props.myProp}</div>) }
        MyComponent.contexttypes = {}
      `,
      errors: [expectTypoStaticClassPropError([3, 21, 3, 33])],
    },
    // ---- Upstream: typoStaticClassProp — ChildContextTypes variants ----
    {
      code: `
        class Component extends React.Component {
          static ChildContextTypes = {};
        }
      `,
      errors: [expectTypoStaticClassPropError([3, 18, 3, 35])],
    },
    {
      code: `
        class Component extends React.Component {}
        Component.ChildContextTypes = {}
      `,
      errors: [expectTypoStaticClassPropError([3, 19, 3, 36])],
    },
    {
      code: `
        function MyComponent() { return (<div>{this.props.myProp}</div>) }
        MyComponent.ChildContextTypes = {}
      `,
      errors: [expectTypoStaticClassPropError([3, 21, 3, 38])],
    },
    {
      code: `
        class Component extends React.Component {
          static childcontexttypes = {};
        }
      `,
      errors: [expectTypoStaticClassPropError([3, 18, 3, 35])],
    },
    {
      code: `
        class Component extends React.Component {}
        Component.childcontexttypes = {}
      `,
      errors: [expectTypoStaticClassPropError([3, 19, 3, 36])],
    },
    {
      code: `
        function MyComponent() { return (<div>{this.props.myProp}</div>) }
        MyComponent.childcontexttypes = {}
      `,
      errors: [expectTypoStaticClassPropError([3, 21, 3, 38])],
    },
    // ---- Upstream: typoStaticClassProp — DefaultProps variants ----
    {
      code: `
        class Component extends React.Component {
          static DefaultProps = {};
        }
      `,
      errors: [expectTypoStaticClassPropError([3, 18, 3, 30])],
    },
    {
      code: `
        class Component extends React.Component {}
        Component.DefaultProps = {}
      `,
      errors: [expectTypoStaticClassPropError([3, 19, 3, 31])],
    },
    {
      code: `
        function MyComponent() { return (<div>{this.props.myProp}</div>) }
        MyComponent.DefaultProps = {}
      `,
      errors: [expectTypoStaticClassPropError([3, 21, 3, 33])],
    },
    {
      code: `
        class Component extends React.Component {
          static defaultprops = {};
        }
      `,
      errors: [expectTypoStaticClassPropError([3, 18, 3, 30])],
    },
    {
      code: `
        class Component extends React.Component {}
        Component.defaultprops = {}
      `,
      errors: [expectTypoStaticClassPropError([3, 19, 3, 31])],
    },
    {
      code: `
        function MyComponent() { return (<div>{this.props.myProp}</div>) }
        MyComponent.defaultprops = {}
      `,
      errors: [expectTypoStaticClassPropError([3, 21, 3, 33])],
    },
    // ---- Upstream: typoStaticClassProp — assignment before class definition ----
    {
      code: `
        Component.defaultprops = {}
        class Component extends React.Component {}
      `,
      errors: [expectTypoStaticClassPropError([2, 19, 2, 31])],
    },
    // ---- Upstream: @extends JSDoc tag ----
    {
      code: `
        /** @extends React.Component */
        class MyComponent extends BaseComponent {}
        MyComponent.PROPTYPES = {}
      `,
      errors: [expectTypoStaticClassPropError([4, 21, 4, 30])],
    },
    // ---- Upstream: typoLifecycleMethod — PascalCase variant ----
    {
      code: `
        class Hello extends React.Component {
          static GetDerivedStateFromProps()  { }
          ComponentWillMount() { }
          UNSAFE_ComponentWillMount() { }
          ComponentDidMount() { }
          ComponentWillReceiveProps() { }
          UNSAFE_ComponentWillReceiveProps() { }
          ShouldComponentUpdate() { }
          ComponentWillUpdate() { }
          UNSAFE_ComponentWillUpdate() { }
          GetSnapshotBeforeUpdate() { }
          ComponentDidUpdate() { }
          ComponentDidCatch() { }
          ComponentWillUnmount() { }
          render() {
            return <div>Hello {this.props.name}</div>;
          }
        }
      `,
      errors: [
        {
          messageId: 'typoLifecycleMethod',
          message:
            'Typo in component lifecycle method declaration: GetDerivedStateFromProps should be getDerivedStateFromProps',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 49,
        },
        {
          messageId: 'typoLifecycleMethod',
          message:
            'Typo in component lifecycle method declaration: ComponentWillMount should be componentWillMount',
          line: 4,
          column: 11,
          endLine: 4,
          endColumn: 35,
        },
        {
          messageId: 'typoLifecycleMethod',
          message:
            'Typo in component lifecycle method declaration: UNSAFE_ComponentWillMount should be UNSAFE_componentWillMount',
          line: 5,
          column: 11,
          endLine: 5,
          endColumn: 42,
        },
        {
          messageId: 'typoLifecycleMethod',
          message:
            'Typo in component lifecycle method declaration: ComponentDidMount should be componentDidMount',
          line: 6,
          column: 11,
          endLine: 6,
          endColumn: 34,
        },
        {
          messageId: 'typoLifecycleMethod',
          message:
            'Typo in component lifecycle method declaration: ComponentWillReceiveProps should be componentWillReceiveProps',
          line: 7,
          column: 11,
          endLine: 7,
          endColumn: 42,
        },
        {
          messageId: 'typoLifecycleMethod',
          message:
            'Typo in component lifecycle method declaration: UNSAFE_ComponentWillReceiveProps should be UNSAFE_componentWillReceiveProps',
          line: 8,
          column: 11,
          endLine: 8,
          endColumn: 49,
        },
        {
          messageId: 'typoLifecycleMethod',
          message:
            'Typo in component lifecycle method declaration: ShouldComponentUpdate should be shouldComponentUpdate',
          line: 9,
          column: 11,
          endLine: 9,
          endColumn: 38,
        },
        {
          messageId: 'typoLifecycleMethod',
          message:
            'Typo in component lifecycle method declaration: ComponentWillUpdate should be componentWillUpdate',
          line: 10,
          column: 11,
          endLine: 10,
          endColumn: 36,
        },
        {
          messageId: 'typoLifecycleMethod',
          message:
            'Typo in component lifecycle method declaration: UNSAFE_ComponentWillUpdate should be UNSAFE_componentWillUpdate',
          line: 11,
          column: 11,
          endLine: 11,
          endColumn: 43,
        },
        {
          messageId: 'typoLifecycleMethod',
          message:
            'Typo in component lifecycle method declaration: GetSnapshotBeforeUpdate should be getSnapshotBeforeUpdate',
          line: 12,
          column: 11,
          endLine: 12,
          endColumn: 40,
        },
        {
          messageId: 'typoLifecycleMethod',
          message:
            'Typo in component lifecycle method declaration: ComponentDidUpdate should be componentDidUpdate',
          line: 13,
          column: 11,
          endLine: 13,
          endColumn: 35,
        },
        {
          messageId: 'typoLifecycleMethod',
          message:
            'Typo in component lifecycle method declaration: ComponentDidCatch should be componentDidCatch',
          line: 14,
          column: 11,
          endLine: 14,
          endColumn: 34,
        },
        {
          messageId: 'typoLifecycleMethod',
          message:
            'Typo in component lifecycle method declaration: ComponentWillUnmount should be componentWillUnmount',
          line: 15,
          column: 11,
          endLine: 15,
          endColumn: 37,
        },
      ],
    },
    // ---- Upstream: typoLifecycleMethod — First-letter-uppercase variant ----
    {
      code: `
        class Hello extends React.Component {
          static Getderivedstatefromprops() { }
          Componentwillmount() { }
          UNSAFE_Componentwillmount() { }
          Componentdidmount() { }
          Componentwillreceiveprops() { }
          UNSAFE_Componentwillreceiveprops() { }
          Shouldcomponentupdate() { }
          Componentwillupdate() { }
          UNSAFE_Componentwillupdate() { }
          Getsnapshotbeforeupdate() { }
          Componentdidupdate() { }
          Componentdidcatch() { }
          Componentwillunmount() { }
          Render() {
            return <div>Hello {this.props.name}</div>;
          }
        }
      `,
      errors: [
        {
          messageId: 'typoLifecycleMethod',
          message:
            'Typo in component lifecycle method declaration: Getderivedstatefromprops should be getDerivedStateFromProps',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 48,
        },
        {
          messageId: 'typoLifecycleMethod',
          message:
            'Typo in component lifecycle method declaration: Componentwillmount should be componentWillMount',
          line: 4,
          column: 11,
          endLine: 4,
          endColumn: 35,
        },
        {
          messageId: 'typoLifecycleMethod',
          message:
            'Typo in component lifecycle method declaration: UNSAFE_Componentwillmount should be UNSAFE_componentWillMount',
          line: 5,
          column: 11,
          endLine: 5,
          endColumn: 42,
        },
        {
          messageId: 'typoLifecycleMethod',
          message:
            'Typo in component lifecycle method declaration: Componentdidmount should be componentDidMount',
          line: 6,
          column: 11,
          endLine: 6,
          endColumn: 34,
        },
        {
          messageId: 'typoLifecycleMethod',
          message:
            'Typo in component lifecycle method declaration: Componentwillreceiveprops should be componentWillReceiveProps',
          line: 7,
          column: 11,
          endLine: 7,
          endColumn: 42,
        },
        {
          messageId: 'typoLifecycleMethod',
          message:
            'Typo in component lifecycle method declaration: UNSAFE_Componentwillreceiveprops should be UNSAFE_componentWillReceiveProps',
          line: 8,
          column: 11,
          endLine: 8,
          endColumn: 49,
        },
        {
          messageId: 'typoLifecycleMethod',
          message:
            'Typo in component lifecycle method declaration: Shouldcomponentupdate should be shouldComponentUpdate',
          line: 9,
          column: 11,
          endLine: 9,
          endColumn: 38,
        },
        {
          messageId: 'typoLifecycleMethod',
          message:
            'Typo in component lifecycle method declaration: Componentwillupdate should be componentWillUpdate',
          line: 10,
          column: 11,
          endLine: 10,
          endColumn: 36,
        },
        {
          messageId: 'typoLifecycleMethod',
          message:
            'Typo in component lifecycle method declaration: UNSAFE_Componentwillupdate should be UNSAFE_componentWillUpdate',
          line: 11,
          column: 11,
          endLine: 11,
          endColumn: 43,
        },
        {
          messageId: 'typoLifecycleMethod',
          message:
            'Typo in component lifecycle method declaration: Getsnapshotbeforeupdate should be getSnapshotBeforeUpdate',
          line: 12,
          column: 11,
          endLine: 12,
          endColumn: 40,
        },
        {
          messageId: 'typoLifecycleMethod',
          message:
            'Typo in component lifecycle method declaration: Componentdidupdate should be componentDidUpdate',
          line: 13,
          column: 11,
          endLine: 13,
          endColumn: 35,
        },
        {
          messageId: 'typoLifecycleMethod',
          message:
            'Typo in component lifecycle method declaration: Componentdidcatch should be componentDidCatch',
          line: 14,
          column: 11,
          endLine: 14,
          endColumn: 34,
        },
        {
          messageId: 'typoLifecycleMethod',
          message:
            'Typo in component lifecycle method declaration: Componentwillunmount should be componentWillUnmount',
          line: 15,
          column: 11,
          endLine: 15,
          endColumn: 37,
        },
        {
          messageId: 'typoLifecycleMethod',
          message:
            'Typo in component lifecycle method declaration: Render should be render',
          line: 16,
          column: 11,
          endLine: 18,
          endColumn: 12,
        },
      ],
    },
    // ---- Upstream: typoLifecycleMethod — lowercase variant ----
    {
      code: `
        class Hello extends React.Component {
          static getderivedstatefromprops() { }
          componentwillmount() { }
          unsafe_componentwillmount() { }
          componentdidmount() { }
          componentwillreceiveprops() { }
          unsafe_componentwillreceiveprops() { }
          shouldcomponentupdate() { }
          componentwillupdate() { }
          unsafe_componentwillupdate() { }
          getsnapshotbeforeupdate() { }
          componentdidupdate() { }
          componentdidcatch() { }
          componentwillunmount() { }
          render() {
            return <div>Hello {this.props.name}</div>;
          }
        }
      `,
      errors: [
        {
          messageId: 'typoLifecycleMethod',
          message:
            'Typo in component lifecycle method declaration: getderivedstatefromprops should be getDerivedStateFromProps',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 48,
        },
        {
          messageId: 'typoLifecycleMethod',
          message:
            'Typo in component lifecycle method declaration: componentwillmount should be componentWillMount',
          line: 4,
          column: 11,
          endLine: 4,
          endColumn: 35,
        },
        {
          messageId: 'typoLifecycleMethod',
          message:
            'Typo in component lifecycle method declaration: unsafe_componentwillmount should be UNSAFE_componentWillMount',
          line: 5,
          column: 11,
          endLine: 5,
          endColumn: 42,
        },
        {
          messageId: 'typoLifecycleMethod',
          message:
            'Typo in component lifecycle method declaration: componentdidmount should be componentDidMount',
          line: 6,
          column: 11,
          endLine: 6,
          endColumn: 34,
        },
        {
          messageId: 'typoLifecycleMethod',
          message:
            'Typo in component lifecycle method declaration: componentwillreceiveprops should be componentWillReceiveProps',
          line: 7,
          column: 11,
          endLine: 7,
          endColumn: 42,
        },
        {
          messageId: 'typoLifecycleMethod',
          message:
            'Typo in component lifecycle method declaration: unsafe_componentwillreceiveprops should be UNSAFE_componentWillReceiveProps',
          line: 8,
          column: 11,
          endLine: 8,
          endColumn: 49,
        },
        {
          messageId: 'typoLifecycleMethod',
          message:
            'Typo in component lifecycle method declaration: shouldcomponentupdate should be shouldComponentUpdate',
          line: 9,
          column: 11,
          endLine: 9,
          endColumn: 38,
        },
        {
          messageId: 'typoLifecycleMethod',
          message:
            'Typo in component lifecycle method declaration: componentwillupdate should be componentWillUpdate',
          line: 10,
          column: 11,
          endLine: 10,
          endColumn: 36,
        },
        {
          messageId: 'typoLifecycleMethod',
          message:
            'Typo in component lifecycle method declaration: unsafe_componentwillupdate should be UNSAFE_componentWillUpdate',
          line: 11,
          column: 11,
          endLine: 11,
          endColumn: 43,
        },
        {
          messageId: 'typoLifecycleMethod',
          message:
            'Typo in component lifecycle method declaration: getsnapshotbeforeupdate should be getSnapshotBeforeUpdate',
          line: 12,
          column: 11,
          endLine: 12,
          endColumn: 40,
        },
        {
          messageId: 'typoLifecycleMethod',
          message:
            'Typo in component lifecycle method declaration: componentdidupdate should be componentDidUpdate',
          line: 13,
          column: 11,
          endLine: 13,
          endColumn: 35,
        },
        {
          messageId: 'typoLifecycleMethod',
          message:
            'Typo in component lifecycle method declaration: componentdidcatch should be componentDidCatch',
          line: 14,
          column: 11,
          endLine: 14,
          endColumn: 34,
        },
        {
          messageId: 'typoLifecycleMethod',
          message:
            'Typo in component lifecycle method declaration: componentwillunmount should be componentWillUnmount',
          line: 15,
          column: 11,
          endLine: 15,
          endColumn: 37,
        },
      ],
    },
    // ---- Upstream: typoPropType (single) ----
    {
      code: `
        import PropTypes from "prop-types";
        class Component extends React.Component {};
        Component.propTypes = {
            a: PropTypes.Number.isRequired
        }
      `,
      errors: [expectTypoPropTypeError2([5, 26, 5, 32])],
    },
    // ---- Upstream: typoPropTypeChain (single) ----
    {
      code: `
        import PropTypes from "prop-types";
        class Component extends React.Component {};
        Component.propTypes = {
            a: PropTypes.number.isrequired
        }
      `,
      errors: [expectTypoPropTypeChainError([5, 33, 5, 43])],
    },
    {
      code: `
        import PropTypes from "prop-types";
        class Component extends React.Component {
          static propTypes = {
            a: PropTypes.number.isrequired
          }
        };
      `,
      errors: [expectTypoPropTypeChainError([5, 33, 5, 43])],
    },
    // ---- Upstream: typoPropType — static field ----
    {
      code: `
        import PropTypes from "prop-types";
        class Component extends React.Component {
          static propTypes = {
            a: PropTypes.Number
          }
        };
      `,
      errors: [expectTypoPropTypeError2([5, 26, 5, 32])],
    },
    {
      code: `
        import PropTypes from "prop-types";
        class Component extends React.Component {};
        Component.propTypes = {
            a: PropTypes.Number
        }
      `,
      errors: [expectTypoPropTypeError2([5, 26, 5, 32])],
    },
    // ---- Upstream: typoPropType — inside shape ----
    {
      code: `
        import PropTypes from "prop-types";
        class Component extends React.Component {};
        Component.propTypes = {
          a: PropTypes.shape({
            b: PropTypes.String,
            c: PropTypes.number.isRequired,
          })
        }
      `,
      errors: [
        {
          messageId: 'typoPropType',
          message: 'Typo in declared prop type: String',
          line: 6,
          column: 26,
          endLine: 6,
          endColumn: 32,
        },
      ],
    },
    // ---- Upstream: typoPropType — inside oneOfType ----
    {
      code: `
        import PropTypes from "prop-types";
        class Component extends React.Component {};
        Component.propTypes = {
          a: PropTypes.oneOfType([
            PropTypes.bools,
            PropTypes.number,
          ])
        }
      `,
      errors: [expectTypoPropTypeError4([6, 23, 6, 28])],
    },
    // ---- Upstream: typoPropType — multiple at top level ----
    {
      code: `
        import PropTypes from "prop-types";
        class Component extends React.Component {};
        Component.propTypes = {
          a: PropTypes.bools,
          b: PropTypes.Array,
          c: PropTypes.function,
          d: PropTypes.objectof,
        }
      `,
      errors: [
        expectTypoPropTypeError4([5, 24, 5, 29]),
        expectTypoPropTypeError5([6, 24, 6, 29]),
        expectTypoPropTypeError6([7, 24, 7, 32]),
        expectTypoPropTypeError7([8, 24, 8, 32]),
      ],
    },
    {
      code: `
        import PropTypes from "prop-types";
        class Component extends React.Component {};
        Component.childContextTypes = {
          a: PropTypes.bools,
          b: PropTypes.Array,
          c: PropTypes.function,
          d: PropTypes.objectof,
        }
      `,
      errors: [
        expectTypoPropTypeError4([5, 24, 5, 29]),
        expectTypoPropTypeError5([6, 24, 6, 29]),
        expectTypoPropTypeError6([7, 24, 7, 32]),
        expectTypoPropTypeError7([8, 24, 8, 32]),
      ],
    },
    {
      code: `
        import PropTypes from 'prop-types';
        class Component extends React.Component {};
        Component.childContextTypes = {
          a: PropTypes.bools,
          b: PropTypes.Array,
          c: PropTypes.function,
          d: PropTypes.objectof,
        }
      `,
      errors: [
        expectTypoPropTypeError4([5, 24, 5, 29]),
        expectTypoPropTypeError5([6, 24, 6, 29]),
        expectTypoPropTypeError6([7, 24, 7, 32]),
        expectTypoPropTypeError7([8, 24, 8, 32]),
      ],
    },
    // ---- Upstream: typoPropTypeChain — multiple .isrequired ----
    {
      code: `
        import PropTypes from 'prop-types';
        class Component extends React.Component {};
        Component.propTypes = {
          a: PropTypes.string.isrequired,
          b: PropTypes.shape({
            c: PropTypes.number
          }).isrequired
        }
      `,
      errors: [
        expectTypoPropTypeChainError([5, 31, 5, 41]),
        expectTypoPropTypeChainError([8, 14, 8, 24]),
      ],
    },
    // ---- Upstream: typoPropType — aliased default import ----
    {
      code: `
        import RealPropTypes from 'prop-types';
        class Component extends React.Component {};
        Component.childContextTypes = {
          a: RealPropTypes.bools,
          b: RealPropTypes.Array,
          c: RealPropTypes.function,
          d: RealPropTypes.objectof,
        }
      `,
      errors: [
        expectTypoPropTypeError4([5, 28, 5, 33]),
        expectTypoPropTypeError5([6, 28, 6, 33]),
        expectTypoPropTypeError6([7, 28, 7, 36]),
        expectTypoPropTypeError7([8, 28, 8, 36]),
      ],
    },
    // ---- Upstream: typoPropTypeChain — React.PropTypes ----
    {
      code: `
      import React from 'react';
      class Component extends React.Component {};
      Component.propTypes = {
        a: React.PropTypes.string.isrequired,
        b: React.PropTypes.shape({
          c: React.PropTypes.number
        }).isrequired
      }
    `,
      errors: [
        expectTypoPropTypeChainError([5, 35, 5, 45]),
        expectTypoPropTypeChainError([8, 12, 8, 22]),
      ],
    },
    {
      code: `
        import React from 'react';
        class Component extends React.Component {};
        Component.childContextTypes = {
          a: React.PropTypes.bools,
          b: React.PropTypes.Array,
          c: React.PropTypes.function,
          d: React.PropTypes.objectof,
        }
      `,
      errors: [
        expectTypoPropTypeError4([5, 30, 5, 35]),
        expectTypoPropTypeError5([6, 30, 6, 35]),
        expectTypoPropTypeError6([7, 30, 7, 38]),
        expectTypoPropTypeError7([8, 30, 8, 38]),
      ],
    },
    // ---- Upstream: typoPropTypeChain — destructured { PropTypes } from react ----
    {
      code: `
      import { PropTypes } from 'react';
      class Component extends React.Component {};
      Component.propTypes = {
        a: PropTypes.string.isrequired,
        b: PropTypes.shape({
          c: PropTypes.number
        }).isrequired
      }
    `,
      errors: [
        expectTypoPropTypeChainError([5, 29, 5, 39]),
        expectTypoPropTypeChainError([8, 12, 8, 22]),
      ],
    },
    // ---- Upstream: noReactBinding ----
    {
      code: `
      import 'react';
      class Component extends React.Component {};
    `,
      errors: [
        {
          messageId: 'noReactBinding',
          message: "`'react'` imported without a local `React` binding.",
          line: 2,
          column: 7,
          endLine: 2,
          endColumn: 22,
        },
      ],
    },
    // ---- Upstream: typoPropType — destructured from react ----
    {
      code: `
        import { PropTypes } from 'react';
        class Component extends React.Component {};
        Component.childContextTypes = {
          a: PropTypes.bools,
          b: PropTypes.Array,
          c: PropTypes.function,
          d: PropTypes.objectof,
        }
      `,
      errors: [
        expectTypoPropTypeError4([5, 24, 5, 29]),
        expectTypoPropTypeError5([6, 24, 6, 29]),
        expectTypoPropTypeError6([7, 24, 7, 32]),
        expectTypoPropTypeError7([8, 24, 8, 32]),
      ],
    },
    // ---- Upstream: typoPropTypeChain in createReactClass ----
    {
      code: `
        import React from 'react';
        import PropTypes from 'prop-types';
        const Component = React.createReactClass({
          propTypes: {
            a: PropTypes.string.isrequired,
            b: PropTypes.shape({
              c: PropTypes.number
            }).isrequired
          }
        });
      `,
      errors: [
        expectTypoPropTypeChainError([6, 33, 6, 43]),
        expectTypoPropTypeChainError([9, 16, 9, 26]),
      ],
    },
    // ---- Upstream: typoPropType in createReactClass childContextTypes ----
    {
      code: `
        import React from 'react';
        import PropTypes from 'prop-types';
        const Component = React.createReactClass({
          childContextTypes: {
            a: PropTypes.bools,
            b: PropTypes.Array,
            c: PropTypes.function,
            d: PropTypes.objectof,
          }
        });
      `,
      errors: [
        expectTypoPropTypeError4([6, 26, 6, 31]),
        expectTypoPropTypeError5([7, 26, 7, 31]),
        expectTypoPropTypeError6([8, 26, 8, 34]),
        expectTypoPropTypeError7([9, 26, 9, 34]),
      ],
    },
    // ---- Upstream: createReactClass — prop declarations + lifecycle typos ----
    {
      code: `
        import React from 'react';
        const Component = React.createReactClass({
          proptypes: {},
          childcontexttypes: {},
          contexttypes: {},
          getdefaultProps() { },
          getinitialState() { },
          getChildcontext() { },
          ComponentWillMount() { },
          ComponentDidMount() { },
          ComponentWillReceiveProps() { },
          ShouldComponentUpdate() { },
          ComponentWillUpdate() { },
          ComponentDidUpdate() { },
          ComponentWillUnmount() { },
          render() {
            return <div>Hello {this.props.name}</div>;
          }
        });
      `,
      errors: [
        expectTypoPropDeclarationError([4, 11, 4, 20]),
        expectTypoPropDeclarationError([5, 11, 5, 28]),
        expectTypoPropDeclarationError([6, 11, 6, 23]),
        {
          messageId: 'typoLifecycleMethod',
          message:
            'Typo in component lifecycle method declaration: getdefaultProps should be getDefaultProps',
          line: 7,
          column: 11,
          endLine: 7,
          endColumn: 32,
        },
        {
          messageId: 'typoLifecycleMethod',
          message:
            'Typo in component lifecycle method declaration: getinitialState should be getInitialState',
          line: 8,
          column: 11,
          endLine: 8,
          endColumn: 32,
        },
        {
          messageId: 'typoLifecycleMethod',
          message:
            'Typo in component lifecycle method declaration: getChildcontext should be getChildContext',
          line: 9,
          column: 11,
          endLine: 9,
          endColumn: 32,
        },
        {
          messageId: 'typoLifecycleMethod',
          message:
            'Typo in component lifecycle method declaration: ComponentWillMount should be componentWillMount',
          line: 10,
          column: 11,
          endLine: 10,
          endColumn: 35,
        },
        {
          messageId: 'typoLifecycleMethod',
          message:
            'Typo in component lifecycle method declaration: ComponentDidMount should be componentDidMount',
          line: 11,
          column: 11,
          endLine: 11,
          endColumn: 34,
        },
        {
          messageId: 'typoLifecycleMethod',
          message:
            'Typo in component lifecycle method declaration: ComponentWillReceiveProps should be componentWillReceiveProps',
          line: 12,
          column: 11,
          endLine: 12,
          endColumn: 42,
        },
        {
          messageId: 'typoLifecycleMethod',
          message:
            'Typo in component lifecycle method declaration: ShouldComponentUpdate should be shouldComponentUpdate',
          line: 13,
          column: 11,
          endLine: 13,
          endColumn: 38,
        },
        {
          messageId: 'typoLifecycleMethod',
          message:
            'Typo in component lifecycle method declaration: ComponentWillUpdate should be componentWillUpdate',
          line: 14,
          column: 11,
          endLine: 14,
          endColumn: 36,
        },
        {
          messageId: 'typoLifecycleMethod',
          message:
            'Typo in component lifecycle method declaration: ComponentDidUpdate should be componentDidUpdate',
          line: 15,
          column: 11,
          endLine: 15,
          endColumn: 35,
        },
        {
          messageId: 'typoLifecycleMethod',
          message:
            'Typo in component lifecycle method declaration: ComponentWillUnmount should be componentWillUnmount',
          line: 16,
          column: 11,
          endLine: 16,
          endColumn: 37,
        },
      ],
    },
    // ---- Upstream: staticLifecycleMethod ----
    {
      code: `
        class Hello extends React.Component {
          getDerivedStateFromProps() { }
        }
      `,
      errors: [
        {
          messageId: 'staticLifecycleMethod',
          message:
            'Lifecycle method should be static: getDerivedStateFromProps',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 41,
        },
      ],
    },
    // ---- Upstream: staticLifecycleMethod + typoLifecycleMethod combined ----
    {
      code: `
        class Hello extends React.Component {
          GetDerivedStateFromProps() { }
        }
      `,
      errors: [
        {
          messageId: 'staticLifecycleMethod',
          message:
            'Lifecycle method should be static: GetDerivedStateFromProps',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 41,
        },
        {
          messageId: 'typoLifecycleMethod',
          message:
            'Typo in component lifecycle method declaration: GetDerivedStateFromProps should be getDerivedStateFromProps',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 41,
        },
      ],
    },
    // ---- Upstream: noPropTypesBinding ----
    {
      code: `
        import 'prop-types'
      `,
      errors: [
        {
          messageId: 'noPropTypesBinding',
          message:
            "`'prop-types'` imported without a local `PropTypes` binding.",
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 28,
        },
      ],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 46 (7.37.5)',
      code: `
      import PropTypes from 'prop-types';
      class Component extends React.Component {};
      Component.propTypes = {
        a: PropTypes.string.isrequired,
        b: PropTypes.shape({
          c: PropTypes.number
        }).isrequired
      }
      `,
      errors: [
        expectTypoPropTypeChainError([5, 29, 5, 39]),
        expectTypoPropTypeChainError([8, 12, 8, 22]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 47 (7.37.5)',
      code: `
      import PropTypes from 'prop-types';
      class Component extends React.Component {};
      Component.propTypes = {
        a: PropTypes.string.isrequired,
        b: PropTypes.shape({
          c: PropTypes.number
        }).isrequired
      }
    `,
      errors: [
        expectTypoPropTypeChainError([5, 29, 5, 39]),
        expectTypoPropTypeChainError([8, 12, 8, 22]),
      ],
      languageOptions: upstreamModuleOptions,
    },
  ],
});

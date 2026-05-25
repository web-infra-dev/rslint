import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-typos', {} as never, {
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
  ],
  invalid: [
    // ---- Upstream: typoStaticClassProp — PropTypes variants ----
    {
      code: `
        class Component extends React.Component {
          static PropTypes = {};
        }
      `,
      errors: [{ messageId: 'typoStaticClassProp' }],
    },
    {
      code: `
        class Component extends React.Component {}
        Component.PropTypes = {}
      `,
      errors: [{ messageId: 'typoStaticClassProp' }],
    },
    {
      code: `
        function MyComponent() { return (<div>{this.props.myProp}</div>) }
        MyComponent.PropTypes = {}
      `,
      errors: [{ messageId: 'typoStaticClassProp' }],
    },
    {
      code: `
        class Component extends React.Component {
          static proptypes = {};
        }
      `,
      errors: [{ messageId: 'typoStaticClassProp' }],
    },
    {
      code: `
        class Component extends React.Component {}
        Component.proptypes = {}
      `,
      errors: [{ messageId: 'typoStaticClassProp' }],
    },
    {
      code: `
        function MyComponent() { return (<div>{this.props.myProp}</div>) }
        MyComponent.proptypes = {}
      `,
      errors: [{ messageId: 'typoStaticClassProp' }],
    },
    // ---- Upstream: typoStaticClassProp — ContextTypes variants ----
    {
      code: `
        class Component extends React.Component {
          static ContextTypes = {};
        }
      `,
      errors: [{ messageId: 'typoStaticClassProp' }],
    },
    {
      code: `
        class Component extends React.Component {}
        Component.ContextTypes = {}
      `,
      errors: [{ messageId: 'typoStaticClassProp' }],
    },
    {
      code: `
        function MyComponent() { return (<div>{this.props.myProp}</div>) }
        MyComponent.ContextTypes = {}
      `,
      errors: [{ messageId: 'typoStaticClassProp' }],
    },
    {
      code: `
        class Component extends React.Component {
          static contexttypes = {};
        }
      `,
      errors: [{ messageId: 'typoStaticClassProp' }],
    },
    {
      code: `
        class Component extends React.Component {}
        Component.contexttypes = {}
      `,
      errors: [{ messageId: 'typoStaticClassProp' }],
    },
    {
      code: `
        function MyComponent() { return (<div>{this.props.myProp}</div>) }
        MyComponent.contexttypes = {}
      `,
      errors: [{ messageId: 'typoStaticClassProp' }],
    },
    // ---- Upstream: typoStaticClassProp — ChildContextTypes variants ----
    {
      code: `
        class Component extends React.Component {
          static ChildContextTypes = {};
        }
      `,
      errors: [{ messageId: 'typoStaticClassProp' }],
    },
    {
      code: `
        class Component extends React.Component {}
        Component.ChildContextTypes = {}
      `,
      errors: [{ messageId: 'typoStaticClassProp' }],
    },
    {
      code: `
        function MyComponent() { return (<div>{this.props.myProp}</div>) }
        MyComponent.ChildContextTypes = {}
      `,
      errors: [{ messageId: 'typoStaticClassProp' }],
    },
    {
      code: `
        class Component extends React.Component {
          static childcontexttypes = {};
        }
      `,
      errors: [{ messageId: 'typoStaticClassProp' }],
    },
    {
      code: `
        class Component extends React.Component {}
        Component.childcontexttypes = {}
      `,
      errors: [{ messageId: 'typoStaticClassProp' }],
    },
    {
      code: `
        function MyComponent() { return (<div>{this.props.myProp}</div>) }
        MyComponent.childcontexttypes = {}
      `,
      errors: [{ messageId: 'typoStaticClassProp' }],
    },
    // ---- Upstream: typoStaticClassProp — DefaultProps variants ----
    {
      code: `
        class Component extends React.Component {
          static DefaultProps = {};
        }
      `,
      errors: [{ messageId: 'typoStaticClassProp' }],
    },
    {
      code: `
        class Component extends React.Component {}
        Component.DefaultProps = {}
      `,
      errors: [{ messageId: 'typoStaticClassProp' }],
    },
    {
      code: `
        function MyComponent() { return (<div>{this.props.myProp}</div>) }
        MyComponent.DefaultProps = {}
      `,
      errors: [{ messageId: 'typoStaticClassProp' }],
    },
    {
      code: `
        class Component extends React.Component {
          static defaultprops = {};
        }
      `,
      errors: [{ messageId: 'typoStaticClassProp' }],
    },
    {
      code: `
        class Component extends React.Component {}
        Component.defaultprops = {}
      `,
      errors: [{ messageId: 'typoStaticClassProp' }],
    },
    {
      code: `
        function MyComponent() { return (<div>{this.props.myProp}</div>) }
        MyComponent.defaultprops = {}
      `,
      errors: [{ messageId: 'typoStaticClassProp' }],
    },
    // ---- Upstream: typoStaticClassProp — assignment before class definition ----
    {
      code: `
        Component.defaultprops = {}
        class Component extends React.Component {}
      `,
      errors: [{ messageId: 'typoStaticClassProp' }],
    },
    // ---- Upstream: @extends JSDoc tag ----
    {
      code: `
        /** @extends React.Component */
        class MyComponent extends BaseComponent {}
        MyComponent.PROPTYPES = {}
      `,
      errors: [{ messageId: 'typoStaticClassProp' }],
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
      errors: Array(13).fill({ messageId: 'typoLifecycleMethod' }),
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
      errors: Array(14).fill({ messageId: 'typoLifecycleMethod' }),
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
      errors: Array(13).fill({ messageId: 'typoLifecycleMethod' }),
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
      errors: [{ messageId: 'typoPropType' }],
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
      errors: [{ messageId: 'typoPropTypeChain' }],
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
      errors: [{ messageId: 'typoPropTypeChain' }],
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
      errors: [{ messageId: 'typoPropType' }],
    },
    {
      code: `
        import PropTypes from "prop-types";
        class Component extends React.Component {};
        Component.propTypes = {
            a: PropTypes.Number
        }
      `,
      errors: [{ messageId: 'typoPropType' }],
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
      errors: [{ messageId: 'typoPropType' }],
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
      errors: [{ messageId: 'typoPropType' }],
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
      errors: Array(4).fill({ messageId: 'typoPropType' }),
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
      errors: Array(4).fill({ messageId: 'typoPropType' }),
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
      errors: Array(4).fill({ messageId: 'typoPropType' }),
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
      errors: Array(2).fill({ messageId: 'typoPropTypeChain' }),
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
      errors: Array(4).fill({ messageId: 'typoPropType' }),
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
      errors: Array(2).fill({ messageId: 'typoPropTypeChain' }),
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
      errors: Array(4).fill({ messageId: 'typoPropType' }),
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
      errors: Array(2).fill({ messageId: 'typoPropTypeChain' }),
    },
    // ---- Upstream: noReactBinding ----
    {
      code: `
      import 'react';
      class Component extends React.Component {};
    `,
      errors: [{ messageId: 'noReactBinding' }],
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
      errors: Array(4).fill({ messageId: 'typoPropType' }),
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
      errors: Array(2).fill({ messageId: 'typoPropTypeChain' }),
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
      errors: Array(4).fill({ messageId: 'typoPropType' }),
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
        { messageId: 'typoPropDeclaration' },
        { messageId: 'typoPropDeclaration' },
        { messageId: 'typoPropDeclaration' },
        { messageId: 'typoLifecycleMethod' },
        { messageId: 'typoLifecycleMethod' },
        { messageId: 'typoLifecycleMethod' },
        { messageId: 'typoLifecycleMethod' },
        { messageId: 'typoLifecycleMethod' },
        { messageId: 'typoLifecycleMethod' },
        { messageId: 'typoLifecycleMethod' },
        { messageId: 'typoLifecycleMethod' },
        { messageId: 'typoLifecycleMethod' },
        { messageId: 'typoLifecycleMethod' },
      ],
    },
    // ---- Upstream: staticLifecycleMethod ----
    {
      code: `
        class Hello extends React.Component {
          getDerivedStateFromProps() { }
        }
      `,
      errors: [{ messageId: 'staticLifecycleMethod' }],
    },
    // ---- Upstream: staticLifecycleMethod + typoLifecycleMethod combined ----
    {
      code: `
        class Hello extends React.Component {
          GetDerivedStateFromProps() { }
        }
      `,
      errors: [
        { messageId: 'staticLifecycleMethod' },
        { messageId: 'typoLifecycleMethod' },
      ],
    },
    // ---- Upstream: noPropTypesBinding ----
    {
      code: `
        import 'prop-types'
      `,
      errors: [{ messageId: 'noPropTypesBinding' }],
    },
  ],
});

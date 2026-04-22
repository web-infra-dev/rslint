// cspell:disable — test cases deliberately contain casing-typo identifiers
// (e.g. `componentwillmount`, `isrequired`, `objectof`) that this rule is
// designed to flag. Disabling cspell for the whole file keeps the intent
// obvious and avoids sprinkling per-line pragmas across ~600 lines.
package no_typos

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoTyposRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoTyposRule, []rule_tester.ValidTestCase{
		// ---- Upstream: non-component classes / functions are ignored ----
		{Code: `
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
        `, Tsx: true},
		{Code: `
          class First {
            static PropTypes = {key: "myValue"};
            static ContextTypes = {key: "myValue"};
            static ChildContextTypes = {key: "myValue"};
            static DefaultProps = {key: "myValue"};
          }
        `, Tsx: true},
		{Code: `
          class First {}
          First.PropTypes = {key: "myValue"};
          First.ContextTypes = {key: "myValue"};
          First.ChildContextTypes = {key: "myValue"};
          First.DefaultProps = {key: "myValue"};
        `, Tsx: true},

		// ---- Upstream: exact casing on static class properties ----
		{Code: `
          class First extends React.Component {
            static propTypes = {key: "myValue"};
            static contextTypes = {key: "myValue"};
            static childContextTypes = {key: "myValue"};
            static defaultProps = {key: "myValue"};
          }
        `, Tsx: true},
		{Code: `
          class First extends React.Component {}
          First.propTypes = {key: "myValue"};
          First.contextTypes = {key: "myValue"};
          First.childContextTypes = {key: "myValue"};
          First.defaultProps = {key: "myValue"};
        `, Tsx: true},

		// ---- Upstream: non-static members of non-component class are fine ----
		{Code: `
          class MyClass {
            propTypes = {key: "myValue"};
            contextTypes = {key: "myValue"};
            childContextTypes = {key: "myValue"};
            defaultProps = {key: "myValue"};
          }
        `, Tsx: true},
		{Code: `
          class MyClass {
            PropTypes = {key: "myValue"};
            ContextTypes = {key: "myValue"};
            ChildContextTypes = {key: "myValue"};
            DefaultProps = {key: "myValue"};
          }
        `, Tsx: true},
		{Code: `
          class MyClass {
            proptypes = {key: "myValue"};
            contexttypes = {key: "myValue"};
            childcontextypes = {key: "myValue"};
            defaultprops = {key: "myValue"};
          }
        `, Tsx: true},
		{Code: `
          class MyClass {
            static PropTypes() {};
            static ContextTypes() {};
            static ChildContextTypes() {};
            static DefaultProps() {};
          }
        `, Tsx: true},
		{Code: `
          class MyClass {
            static proptypes() {};
            static contexttypes() {};
            static childcontexttypes() {};
            static defaultprops() {};
          }
        `, Tsx: true},
		{Code: `
          class MyClass {}
          MyClass.prototype.PropTypes = function() {};
          MyClass.prototype.ContextTypes = function() {};
          MyClass.prototype.ChildContextTypes = function() {};
          MyClass.prototype.DefaultProps = function() {};
        `, Tsx: true},
		{Code: `
          class MyClass {}
          MyClass.PropTypes = function() {};
          MyClass.ContextTypes = function() {};
          MyClass.ChildContextTypes = function() {};
          MyClass.DefaultProps = function() {};
        `, Tsx: true},
		{Code: `
          function MyRandomFunction() {}
          MyRandomFunction.PropTypes = {};
          MyRandomFunction.ContextTypes = {};
          MyRandomFunction.ChildContextTypes = {};
          MyRandomFunction.DefaultProps = {};
        `, Tsx: true},

		// ---- Upstream: unsupported dynamic computed keys (bracket notation) ----
		{Code: `
          class First extends React.Component {}
          First["prop" + "Types"] = {};
          First["context" + "Types"] = {};
          First["childContext" + "Types"] = {};
          First["default" + "Props"] = {};
        `, Tsx: true},
		{Code: `
          class First extends React.Component {}
          First["PROP" + "TYPES"] = {};
          First["CONTEXT" + "TYPES"] = {};
          First["CHILDCONTEXT" + "TYPES"] = {};
          First["DEFAULT" + "PROPS"] = {};
        `, Tsx: true},
		{Code: `
          const propTypes = "PROPTYPES"
          const contextTypes = "CONTEXTTYPES"
          const childContextTypes = "CHILDCONTEXTTYPES"
          const defaultProps = "DEFAULTPROPS"

          class First extends React.Component {}
          First[propTypes] = {};
          First[contextTypes] = {};
          First[childContextTypes] = {};
          First[defaultProps] = {};
        `, Tsx: true},

		// ---- Upstream: well-cased lifecycle methods ----
		{Code: `
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
        `, Tsx: true},
		{Code: `
          class Hello extends React.Component {
            "componentDidMount"() { }
            "my-method"() { }
          }
        `, Tsx: true},
		{Code: `
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
        `, Tsx: true},
		{Code: `
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
        `, Tsx: true},
		{Code: `
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
        `, Tsx: true},

		// ---- Upstream: issue #1353 — unrelated .bind ----
		{Code: `
          function test(b) {
            return a.bind(b);
          }
          function a() {}
        `, Tsx: true},

		// ---- Upstream: well-formed PropTypes usage ----
		{Code: `
          import PropTypes from "prop-types";
          class Component extends React.Component {};
          Component.propTypes = {
            a: PropTypes.number.isRequired
          }
        `, Tsx: true},
		{Code: `
          import PropTypes from "prop-types";
          class Component extends React.Component {};
          Component.propTypes = {
            e: PropTypes.shape({
              ea: PropTypes.string,
            })
          }
        `, Tsx: true},
		{Code: `
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
        `, Tsx: true},
		{Code: `
          import PropTypes from "prop-types";
          class Component extends React.Component {};
          Component.propTypes = {
            a: PropTypes.oneOfType([
              PropTypes.string,
              PropTypes.number
            ])
          }
        `, Tsx: true},
		{Code: `
          import PropTypes from "prop-types";
          class Component extends React.Component {};
          Component.propTypes = {
            a: PropTypes.oneOf([
              'hello',
              'hi'
            ])
          }
        `, Tsx: true},
		{Code: `
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
        `, Tsx: true},
		{Code: `
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
        `, Tsx: true},

		// ---- Upstream: external types alongside prop-types ----
		{Code: `
          import PropTypes from 'prop-types'
          import * as MyPropTypes from 'lib/my-prop-types'
          class Component extends React.Component {};
          Component.propTypes = {
            a: PropTypes.string,
            b: MyPropTypes.MYSTRING,
            c: MyPropTypes.MYSTRING.isRequired,
          }
        `, Tsx: true},
		{Code: `
          import PropTypes from "prop-types"
          import * as MyPropTypes from 'lib/my-prop-types'
          class Component extends React.Component {};
          Component.propTypes = {
            b: PropTypes.string,
            a: MyPropTypes.MYSTRING,
          }
        `, Tsx: true},
		{Code: `
          import CustomReact from "react"
          class Component extends React.Component {};
          Component.propTypes = {
            b: CustomReact.PropTypes.string,
          }
        `, Tsx: true},

		// ---- Upstream: absent arg to PropTypes.shape must not crash ----
		{Code: `
          class Component extends React.Component {};
          Component.propTypes = {
            a: PropTypes.shape(),
          };
          Component.contextTypes = {
            a: PropTypes.shape(),
          };
        `, Tsx: true},

		// ---- Upstream: unrelated patterns ----
		{Code: `
          const fn = (err, res) => {
            const { body: data = {} } = { ...res };
            data.time = data.time || {};
          };
        `, Tsx: true},
		{Code: `
          class Component extends React.Component {};
          Component.propTypes = {
            b: string.isRequired,
            c: PropTypes.shape({
              d: number.isRequired,
            }).isRequired
          }
        `, Tsx: true},

		// ---- Upstream: createReactClass with PropTypes ----
		{Code: `
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
        `, Tsx: true},
		{Code: `
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
        `, Tsx: true},
		{Code: `
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
        `, Tsx: true},

		// ---- Upstream: destructured named imports from prop-types ----
		{Code: `
          import { string, element } from "prop-types";

          class Sample extends React.Component {
            render() { return null; }
          }

          Sample.propTypes = {
            title: string.isRequired,
            body: element.isRequired
          };
        `, Tsx: true},

		// ---- Upstream: computed key with PropertyAccessExpression base ----
		{Code: `
          import React from 'react';

          const A = { B: 'C' };

          export default class MyComponent extends React.Component {
            [A.B] () {
              return null
            }
          }
        `, Tsx: true},

		// ---- Upstream: React.forwardRef / styled-components escape hatches ----
		{Code: `
          const MyComponent = React.forwardRef((props, ref) => <div />);
          MyComponent.defaultProps = { value: "" };
        `, Tsx: true},
		{Code: `
          import styled from "styled-components";

          const MyComponent = styled.div;
          MyComponent.defaultProps = { value: "" };
        `, Tsx: true},

		// ---- Upstream: private method should not trigger lifecycle typo ----
		{Code: `
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
        `, Tsx: true},

		// ---- Edge: PureComponent is also a React component (well-cased) ----
		{Code: `
          class Hello extends React.PureComponent {
            componentDidMount() {}
            render() { return <div/>; }
          }
          Hello.propTypes = {};
        `, Tsx: true},

		// ---- Edge: bare Component / PureComponent identifiers ----
		{Code: `
          class Hello extends Component {
            componentDidMount() {}
            render() { return <div/>; }
          }
          Hello.propTypes = {};
        `, Tsx: true},
		{Code: `
          class Hello extends PureComponent {}
          Hello.defaultProps = {};
        `, Tsx: true},

		// ---- Edge: ClassExpression assigned to const ----
		{Code: `
          const Foo = class extends React.Component {
            render() { return <div/>; }
          };
          Foo.propTypes = {};
        `, Tsx: true},

		// ---- Edge: arrow function component is tracked for the Foo.Prop path ----
		{Code: `
          const Foo = () => <div/>;
          Foo.defaultProps = {};
        `, Tsx: true},

		// ---- Edge: parenthesized extends clause ----
		{Code: `
          class Hello extends (React.Component) {
            componentDidMount() {}
          }
        `, Tsx: true},

		// ---- Edge: deeply nested shape + oneOfType (all correct) ----
		{Code: `
          import PropTypes from 'prop-types';
          class C extends React.Component {}
          C.propTypes = {
            x: PropTypes.shape({
              y: PropTypes.oneOfType([
                PropTypes.shape({
                  z: PropTypes.number.isRequired,
                }),
                PropTypes.string,
              ]).isRequired,
            }).isRequired,
          };
        `, Tsx: true},

		// ---- Edge: spread in propTypes object is tolerated ----
		{Code: `
          import PropTypes from 'prop-types';
          const extra = { b: PropTypes.string };
          class C extends React.Component {}
          C.propTypes = {
            a: PropTypes.number,
            ...extra,
          };
        `, Tsx: true},

		// ---- Edge: string key on a class method with correct casing ----
		{Code: `
          class Hello extends React.Component {
            "componentWillMount"() {}
          }
        `, Tsx: true},

		// ---- Edge: aliased named import of PropTypes from react ----
		{Code: `
          import { PropTypes as PT } from 'react';
          class C extends React.Component {}
          C.propTypes = {
            a: PT.number.isRequired,
          };
        `, Tsx: true},

		// ---- Edge: namespace import of prop-types ----
		{Code: `
          import * as PT from 'prop-types';
          class C extends React.Component {}
          C.propTypes = {
            a: PT.string.isRequired,
          };
        `, Tsx: true},

		// ---- Edge: numeric key should be ignored by lifecycle check ----
		{Code: `
          class Hello extends React.Component {
            1() {}
          }
        `, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// ---- Upstream: typoStaticClassProp — PropTypes variants ----
		{Code: `
        class Component extends React.Component {
          static PropTypes = {};
        }
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "typoStaticClassProp"}}},
		{Code: `
        class Component extends React.Component {}
        Component.PropTypes = {}
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "typoStaticClassProp"}}},
		{Code: `
        function MyComponent() { return (<div>{this.props.myProp}</div>) }
        MyComponent.PropTypes = {}
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "typoStaticClassProp"}}},
		{Code: `
        class Component extends React.Component {
          static proptypes = {};
        }
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "typoStaticClassProp"}}},
		{Code: `
        class Component extends React.Component {}
        Component.proptypes = {}
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "typoStaticClassProp"}}},
		{Code: `
        function MyComponent() { return (<div>{this.props.myProp}</div>) }
        MyComponent.proptypes = {}
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "typoStaticClassProp"}}},

		// ---- Upstream: typoStaticClassProp — ContextTypes variants ----
		{Code: `
        class Component extends React.Component {
          static ContextTypes = {};
        }
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "typoStaticClassProp"}}},
		{Code: `
        class Component extends React.Component {}
        Component.ContextTypes = {}
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "typoStaticClassProp"}}},
		{Code: `
        function MyComponent() { return (<div>{this.props.myProp}</div>) }
        MyComponent.ContextTypes = {}
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "typoStaticClassProp"}}},
		{Code: `
        class Component extends React.Component {
          static contexttypes = {};
        }
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "typoStaticClassProp"}}},
		{Code: `
        class Component extends React.Component {}
        Component.contexttypes = {}
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "typoStaticClassProp"}}},
		{Code: `
        function MyComponent() { return (<div>{this.props.myProp}</div>) }
        MyComponent.contexttypes = {}
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "typoStaticClassProp"}}},

		// ---- Upstream: typoStaticClassProp — ChildContextTypes variants ----
		{Code: `
        class Component extends React.Component {
          static ChildContextTypes = {};
        }
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "typoStaticClassProp"}}},
		{Code: `
        class Component extends React.Component {}
        Component.ChildContextTypes = {}
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "typoStaticClassProp"}}},
		{Code: `
        function MyComponent() { return (<div>{this.props.myProp}</div>) }
        MyComponent.ChildContextTypes = {}
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "typoStaticClassProp"}}},
		{Code: `
        class Component extends React.Component {
          static childcontexttypes = {};
        }
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "typoStaticClassProp"}}},
		{Code: `
        class Component extends React.Component {}
        Component.childcontexttypes = {}
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "typoStaticClassProp"}}},
		{Code: `
        function MyComponent() { return (<div>{this.props.myProp}</div>) }
        MyComponent.childcontexttypes = {}
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "typoStaticClassProp"}}},

		// ---- Upstream: typoStaticClassProp — DefaultProps variants ----
		{Code: `
        class Component extends React.Component {
          static DefaultProps = {};
        }
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "typoStaticClassProp"}}},
		{Code: `
        class Component extends React.Component {}
        Component.DefaultProps = {}
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "typoStaticClassProp"}}},
		{Code: `
        function MyComponent() { return (<div>{this.props.myProp}</div>) }
        MyComponent.DefaultProps = {}
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "typoStaticClassProp"}}},
		{Code: `
        class Component extends React.Component {
          static defaultprops = {};
        }
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "typoStaticClassProp"}}},
		{Code: `
        class Component extends React.Component {}
        Component.defaultprops = {}
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "typoStaticClassProp"}}},
		{Code: `
        function MyComponent() { return (<div>{this.props.myProp}</div>) }
        MyComponent.defaultprops = {}
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "typoStaticClassProp"}}},

		// ---- Upstream: typoStaticClassProp — assignment before class definition ----
		{Code: `
        Component.defaultprops = {}
        class Component extends React.Component {}
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "typoStaticClassProp"}}},

		// ---- Upstream: @extends JSDoc tag ----
		{Code: `
        /** @extends React.Component */
        class MyComponent extends BaseComponent {}
        MyComponent.PROPTYPES = {}
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "typoStaticClassProp"}}},

		// ---- Upstream: typoLifecycleMethod — PascalCase variant ----
		{Code: `
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
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "typoLifecycleMethod", Line: 3},
			{MessageId: "typoLifecycleMethod", Line: 4},
			{MessageId: "typoLifecycleMethod", Line: 5},
			{MessageId: "typoLifecycleMethod", Line: 6},
			{MessageId: "typoLifecycleMethod", Line: 7},
			{MessageId: "typoLifecycleMethod", Line: 8},
			{MessageId: "typoLifecycleMethod", Line: 9},
			{MessageId: "typoLifecycleMethod", Line: 10},
			{MessageId: "typoLifecycleMethod", Line: 11},
			{MessageId: "typoLifecycleMethod", Line: 12},
			{MessageId: "typoLifecycleMethod", Line: 13},
			{MessageId: "typoLifecycleMethod", Line: 14},
			{MessageId: "typoLifecycleMethod", Line: 15},
		}},

		// ---- Upstream: typoLifecycleMethod — First-letter-uppercase variant ----
		{Code: `
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
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "typoLifecycleMethod", Line: 3},
			{MessageId: "typoLifecycleMethod", Line: 4},
			{MessageId: "typoLifecycleMethod", Line: 5},
			{MessageId: "typoLifecycleMethod", Line: 6},
			{MessageId: "typoLifecycleMethod", Line: 7},
			{MessageId: "typoLifecycleMethod", Line: 8},
			{MessageId: "typoLifecycleMethod", Line: 9},
			{MessageId: "typoLifecycleMethod", Line: 10},
			{MessageId: "typoLifecycleMethod", Line: 11},
			{MessageId: "typoLifecycleMethod", Line: 12},
			{MessageId: "typoLifecycleMethod", Line: 13},
			{MessageId: "typoLifecycleMethod", Line: 14},
			{MessageId: "typoLifecycleMethod", Line: 15},
			{MessageId: "typoLifecycleMethod", Line: 16},
		}},

		// ---- Upstream: typoLifecycleMethod — lowercase variant ----
		{Code: `
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
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "typoLifecycleMethod"},
			{MessageId: "typoLifecycleMethod"},
			{MessageId: "typoLifecycleMethod"},
			{MessageId: "typoLifecycleMethod"},
			{MessageId: "typoLifecycleMethod"},
			{MessageId: "typoLifecycleMethod"},
			{MessageId: "typoLifecycleMethod"},
			{MessageId: "typoLifecycleMethod"},
			{MessageId: "typoLifecycleMethod"},
			{MessageId: "typoLifecycleMethod"},
			{MessageId: "typoLifecycleMethod"},
			{MessageId: "typoLifecycleMethod"},
			{MessageId: "typoLifecycleMethod"},
		}},

		// ---- Upstream: typoPropType (single) ----
		{Code: `
        import PropTypes from "prop-types";
        class Component extends React.Component {};
        Component.propTypes = {
            a: PropTypes.Number.isRequired
        }
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "typoPropType"}}},

		// ---- Upstream: typoPropTypeChain (single) ----
		{Code: `
        import PropTypes from "prop-types";
        class Component extends React.Component {};
        Component.propTypes = {
            a: PropTypes.number.isrequired
        }
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "typoPropTypeChain"}}},
		{Code: `
        import PropTypes from "prop-types";
        class Component extends React.Component {
          static propTypes = {
            a: PropTypes.number.isrequired
          }
        };
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "typoPropTypeChain"}}},

		// ---- Upstream: typoPropType — static field ----
		{Code: `
        import PropTypes from "prop-types";
        class Component extends React.Component {
          static propTypes = {
            a: PropTypes.Number
          }
        };
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "typoPropType"}}},
		{Code: `
        import PropTypes from "prop-types";
        class Component extends React.Component {};
        Component.propTypes = {
            a: PropTypes.Number
        }
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "typoPropType"}}},

		// ---- Upstream: typoPropType — inside shape ----
		{Code: `
        import PropTypes from "prop-types";
        class Component extends React.Component {};
        Component.propTypes = {
          a: PropTypes.shape({
            b: PropTypes.String,
            c: PropTypes.number.isRequired,
          })
        }
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "typoPropType"}}},

		// ---- Upstream: typoPropType — inside oneOfType ----
		{Code: `
        import PropTypes from "prop-types";
        class Component extends React.Component {};
        Component.propTypes = {
          a: PropTypes.oneOfType([
            PropTypes.bools,
            PropTypes.number,
          ])
        }
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "typoPropType"}}},

		// ---- Upstream: typoPropType — multiple at top level ----
		{Code: `
        import PropTypes from "prop-types";
        class Component extends React.Component {};
        Component.propTypes = {
          a: PropTypes.bools,
          b: PropTypes.Array,
          c: PropTypes.function,
          d: PropTypes.objectof,
        }
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "typoPropType"},
			{MessageId: "typoPropType"},
			{MessageId: "typoPropType"},
			{MessageId: "typoPropType"},
		}},
		{Code: `
        import PropTypes from "prop-types";
        class Component extends React.Component {};
        Component.childContextTypes = {
          a: PropTypes.bools,
          b: PropTypes.Array,
          c: PropTypes.function,
          d: PropTypes.objectof,
        }
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "typoPropType"},
			{MessageId: "typoPropType"},
			{MessageId: "typoPropType"},
			{MessageId: "typoPropType"},
		}},
		{Code: `
        import PropTypes from 'prop-types';
        class Component extends React.Component {};
        Component.childContextTypes = {
          a: PropTypes.bools,
          b: PropTypes.Array,
          c: PropTypes.function,
          d: PropTypes.objectof,
        }
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "typoPropType"},
			{MessageId: "typoPropType"},
			{MessageId: "typoPropType"},
			{MessageId: "typoPropType"},
		}},

		// ---- Upstream: typoPropTypeChain — multiple .isrequired ----
		{Code: `
        import PropTypes from 'prop-types';
        class Component extends React.Component {};
        Component.propTypes = {
          a: PropTypes.string.isrequired,
          b: PropTypes.shape({
            c: PropTypes.number
          }).isrequired
        }
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "typoPropTypeChain"},
			{MessageId: "typoPropTypeChain"},
		}},

		// ---- Upstream: typoPropType — aliased default import ----
		{Code: `
        import RealPropTypes from 'prop-types';
        class Component extends React.Component {};
        Component.childContextTypes = {
          a: RealPropTypes.bools,
          b: RealPropTypes.Array,
          c: RealPropTypes.function,
          d: RealPropTypes.objectof,
        }
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "typoPropType"},
			{MessageId: "typoPropType"},
			{MessageId: "typoPropType"},
			{MessageId: "typoPropType"},
		}},

		// ---- Upstream: typoPropTypeChain — React.PropTypes ----
		{Code: `
      import React from 'react';
      class Component extends React.Component {};
      Component.propTypes = {
        a: React.PropTypes.string.isrequired,
        b: React.PropTypes.shape({
          c: React.PropTypes.number
        }).isrequired
      }
    `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "typoPropTypeChain"},
			{MessageId: "typoPropTypeChain"},
		}},
		{Code: `
        import React from 'react';
        class Component extends React.Component {};
        Component.childContextTypes = {
          a: React.PropTypes.bools,
          b: React.PropTypes.Array,
          c: React.PropTypes.function,
          d: React.PropTypes.objectof,
        }
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "typoPropType"},
			{MessageId: "typoPropType"},
			{MessageId: "typoPropType"},
			{MessageId: "typoPropType"},
		}},

		// ---- Upstream: typoPropTypeChain — destructured { PropTypes } from react ----
		{Code: `
      import { PropTypes } from 'react';
      class Component extends React.Component {};
      Component.propTypes = {
        a: PropTypes.string.isrequired,
        b: PropTypes.shape({
          c: PropTypes.number
        }).isrequired
      }
    `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "typoPropTypeChain"},
			{MessageId: "typoPropTypeChain"},
		}},

		// ---- Upstream: noReactBinding ----
		{Code: `
      import 'react';
      class Component extends React.Component {};
    `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noReactBinding"}}},

		// ---- Upstream: typoPropType — destructured from react ----
		{Code: `
        import { PropTypes } from 'react';
        class Component extends React.Component {};
        Component.childContextTypes = {
          a: PropTypes.bools,
          b: PropTypes.Array,
          c: PropTypes.function,
          d: PropTypes.objectof,
        }
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "typoPropType"},
			{MessageId: "typoPropType"},
			{MessageId: "typoPropType"},
			{MessageId: "typoPropType"},
		}},

		// ---- Upstream: typoPropTypeChain — leading whitespace variants ----
		{Code: `
      import PropTypes from 'prop-types';
      class Component extends React.Component {};
      Component.propTypes = {
        a: PropTypes.string.isrequired,
        b: PropTypes.shape({
          c: PropTypes.number
        }).isrequired
      }
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "typoPropTypeChain"},
			{MessageId: "typoPropTypeChain"},
		}},

		// ---- Upstream: typoPropTypeChain in createReactClass ----
		{Code: `
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
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "typoPropTypeChain"},
			{MessageId: "typoPropTypeChain"},
		}},

		// ---- Upstream: typoPropType in createReactClass childContextTypes ----
		{Code: `
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
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "typoPropType"},
			{MessageId: "typoPropType"},
			{MessageId: "typoPropType"},
			{MessageId: "typoPropType"},
		}},

		// ---- Upstream: createReactClass — prop declarations + lifecycle typos ----
		{Code: `
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
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "typoPropDeclaration"},
			{MessageId: "typoPropDeclaration"},
			{MessageId: "typoPropDeclaration"},
			{MessageId: "typoLifecycleMethod"},
			{MessageId: "typoLifecycleMethod"},
			{MessageId: "typoLifecycleMethod"},
			{MessageId: "typoLifecycleMethod"},
			{MessageId: "typoLifecycleMethod"},
			{MessageId: "typoLifecycleMethod"},
			{MessageId: "typoLifecycleMethod"},
			{MessageId: "typoLifecycleMethod"},
			{MessageId: "typoLifecycleMethod"},
			{MessageId: "typoLifecycleMethod"},
		}},

		// ---- Upstream: staticLifecycleMethod ----
		{Code: `
        class Hello extends React.Component {
          getDerivedStateFromProps() { }
        }
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "staticLifecycleMethod"}}},

		// ---- Upstream: staticLifecycleMethod + typoLifecycleMethod combined ----
		{Code: `
        class Hello extends React.Component {
          GetDerivedStateFromProps() { }
        }
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "staticLifecycleMethod"},
			{MessageId: "typoLifecycleMethod"},
		}},

		// ---- Upstream: noPropTypesBinding ----
		{Code: `
        import 'prop-types'
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noPropTypesBinding"}}},

		// ---- Contract: Line/Column + Message assertions (typoStaticClassProp) ----
		{Code: "class C extends React.Component {\n  static ProPtYpEs = {};\n}",
			Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "typoStaticClassProp",
				Message:   "Typo in static class property declaration",
				Line:      2, Column: 10, EndLine: 2, EndColumn: 19,
			}}},
		{Code: "class C extends React.Component {}\nC.DefaultProps = {};",
			Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "typoStaticClassProp",
				Message:   "Typo in static class property declaration",
				Line:      2, Column: 3, EndLine: 2, EndColumn: 15,
			}}},

		// ---- Contract: Line/Column + Message assertions (typoPropType) ----
		{Code: "import PropTypes from 'prop-types';\nclass C extends React.Component {}\nC.propTypes = { a: PropTypes.Number };",
			Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "typoPropType",
				Message:   "Typo in declared prop type: Number",
				Line:      3, Column: 30, EndLine: 3, EndColumn: 36,
			}}},

		// ---- Contract: Line/Column + Message assertions (typoPropTypeChain) ----
		{Code: "import PropTypes from 'prop-types';\nclass C extends React.Component {}\nC.propTypes = { a: PropTypes.number.isrequired };",
			Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "typoPropTypeChain",
				Message:   "Typo in prop type chain qualifier: isrequired",
				Line:      3, Column: 37, EndLine: 3, EndColumn: 47,
			}}},

		// ---- Contract: Line/Column + Message assertions (typoLifecycleMethod) ----
		{Code: "class C extends React.Component {\n  ComponentDidMount() {}\n}",
			Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "typoLifecycleMethod",
				Message:   "Typo in component lifecycle method declaration: ComponentDidMount should be componentDidMount",
				Line:      2, Column: 3, EndLine: 2, EndColumn: 25,
			}}},

		// ---- Contract: Line/Column + Message assertions (staticLifecycleMethod) ----
		{Code: "class C extends React.Component {\n  getDerivedStateFromProps() {}\n}",
			Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "staticLifecycleMethod",
				Message:   "Lifecycle method should be static: getDerivedStateFromProps",
				Line:      2, Column: 3, EndLine: 2, EndColumn: 32,
			}}},

		// ---- Edge: typoStaticClassProp reported on the key, not the class ----
		{Code: "class Hello extends React.Component {\n  static propTypes = {};\n  static proptypes = {};\n}",
			Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "typoStaticClassProp", Line: 3, Column: 10}}},

		// ---- Edge: PureComponent variants are detected ----
		{Code: "class C extends React.PureComponent {\n  ComponentDidMount() {}\n}",
			Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "typoLifecycleMethod", Line: 2}}},
		{Code: "class C extends PureComponent {}\nC.PropTypes = {};",
			Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "typoStaticClassProp", Line: 2}}},

		// ---- Edge: ClassExpression assigned to variable ----
		{Code: "const Foo = class extends React.Component {};\nFoo.PropTypes = {};",
			Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "typoStaticClassProp", Line: 2}}},

		// ---- Edge: arrow function component is detected for Foo.Prop path ----
		{Code: "const Foo = () => <div/>;\nFoo.DefaultProps = {};",
			Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "typoStaticClassProp", Line: 2}}},

		// ---- Edge: deeply nested typo inside oneOfType inside shape ----
		// Emission order follows the recursive descent: the `.isrequired`
		// qualifier on the outer oneOfType is reported before descending
		// into the oneOfType's element array.
		{Code: `
        import PropTypes from 'prop-types';
        class C extends React.Component {}
        C.propTypes = {
          x: PropTypes.shape({
            y: PropTypes.oneOfType([
              PropTypes.Bool,
              PropTypes.shape({ z: PropTypes.Number })
            ]).isrequired
          })
        };
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "typoPropTypeChain"}, // .isrequired
			{MessageId: "typoPropType"},      // PropTypes.Bool
			{MessageId: "typoPropType"},      // PropTypes.Number inside nested shape
		}},

		// ---- Edge: parenthesized LHS target ----
		{Code: "class C extends React.Component {}\n(C).PropTypes = {};",
			Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "typoStaticClassProp"}}},

		// ---- Edge: @augments (alias of @extends) JSDoc tag ----
		{Code: `
        /** @augments React.PureComponent */
        class MyComponent extends BaseComponent {}
        MyComponent.PROPTYPES = {}
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "typoStaticClassProp"}}},

		// ---- Edge: aliased React namespace — React.PropTypes still detected ----
		{Code: `
        import * as MyReact from 'react';
        class C extends React.Component {}
        C.propTypes = {
          a: MyReact.PropTypes.string.isrequired,
        };
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "typoPropTypeChain"}}},

		// ---- Edge: string-literal lifecycle method key in class ----
		{Code: `
        class C extends React.Component {
          "ComponentDidMount"() {}
        }
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "typoLifecycleMethod"}}},

		// ---- Edge: string-literal lifecycle method key in ES5 createReactClass ----
		{Code: `
        import React from 'react';
        const C = React.createReactClass({
          "ComponentDidMount": function() {},
        });
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "typoLifecycleMethod"}}},
	})
}

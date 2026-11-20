import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, upstreamModuleOptions } from '../config';

function expectBindCallError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'bindCall',
    message: 'JSX props should not use .bind()',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectArrowFuncError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'arrowFunc',
    message: 'JSX props should not use arrow functions',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectFuncError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'func',
    message: 'JSX props should not use functions',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectBindExpressionError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'bindExpression',
    message: 'JSX props should not use ::',
    line,
    column,
    endLine,
    endColumn,
  };
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('jsx-no-bind', {
  valid: [
    { code: `var x = <div onClick={this._handleClick}></div>;` },
    { code: `var x = <Foo onClick={this._handleClick} />;` },
    { code: `var x = <div meaningOfLife={42}></div>;` },
    { code: `var x = <div onClick={getHandler()}></div>;` },

    // ignoreRefs
    {
      code: `var x = <div ref={c => (this._input = c)}></div>;`,
      options: [{ ignoreRefs: true }],
    },
    {
      code: `var x = <div ref={this._refCallback.bind(this)}></div>;`,
      options: [{ ignoreRefs: true }],
    },

    // allowBind
    {
      code: `var x = <div onClick={this._handleClick.bind(this)}></div>;`,
      options: [{ allowBind: true }],
    },

    // allowArrowFunctions
    {
      code: `var x = <div onClick={() => alert("1337")}></div>;`,
      options: [{ allowArrowFunctions: true }],
    },

    // allowFunctions
    {
      code: `var x = <div onClick={function () { alert("1337"); }}></div>;`,
      options: [{ allowFunctions: true }],
    },

    // ignoreDOMComponents
    {
      code: `var x = <div onClick={this._handleClick.bind(this)}></div>;`,
      options: [{ ignoreDOMComponents: true }],
    },
    {
      code: `var x = <div onClick={() => alert("1337")}></div>;`,
      options: [{ ignoreDOMComponents: true }],
    },

    // Not attached to JSX
    {
      code: `
        class Hello extends Component {
          render() {
            const click = this.onTap.bind(this);
            return <div onClick={onClick}>Hello</div>;
          }
        }
      `,
    },

    // Uninitialized variable should not crash
    {
      code: `
        class Hello extends Component {
          render() {
            let click;
            return <div onClick={onClick}>Hello</div>;
          }
        }
      `,
    },

    // Top-level function declarations are not tracked
    {
      code: `
        function click() { return true; }
        class Hello23 extends React.Component {
          renderDiv() {
            return <div onClick={click}>Hello</div>;
          }
        }
      `,
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 0 (7.37.5)',
      code: `<div onClick={this._handleClick}></div>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 1 (7.37.5)',
      code: `<div onClick={this._handleClick}></div>`,
      options: [{}],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 2 (7.37.5)',
      code: `<Foo onClick={this._handleClick} />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 3 (7.37.5)',
      code: `<Foo onClick={this._handleClick} />`,
      options: [{}],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 4 (7.37.5)',
      code: `<div meaningOfLife={42}></div>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 5 (7.37.5)',
      code: `<div onClick={getHandler()}></div>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 6 (7.37.5)',
      code: `<div ref={c => this._input = c}></div>`,
      options: [{ ignoreRefs: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 7 (7.37.5)',
      code: `<div ref={this._refCallback.bind(this)}></div>`,
      options: [{ ignoreRefs: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 8 (7.37.5)',
      code: `<div ref={function (c) {this._input = c}}></div>`,
      options: [{ ignoreRefs: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 9 (7.37.5)',
      code: `<div onClick={this._handleClick.bind(this)}></div>`,
      options: [{ allowBind: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 10 (7.37.5)',
      code: `<div onClick={() => alert("1337")}></div>`,
      options: [{ allowArrowFunctions: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 11 (7.37.5)',
      code: `<div onClick={async () => alert("1337")}></div>`,
      options: [{ allowArrowFunctions: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 12 (7.37.5)',
      code: `<div onClick={function () { alert("1337") }}></div>`,
      options: [{ allowFunctions: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 13 (7.37.5)',
      code: `<div onClick={function * () { alert("1337") }}></div>`,
      options: [{ allowFunctions: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 14 (7.37.5)',
      code: `<div onClick={async function () { alert("1337") }}></div>`,
      options: [{ allowFunctions: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 15 (7.37.5)',
      code: `
        class Hello extends Component {
          render() {
            return <div>Hello</div>;
          }
        }
        export default connect()(Hello);
      `,
      options: [{ allowBind: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 16 (7.37.5)',
      code: `
        var DocumentRow = Backbone.View.extend({
          tagName: "li",
          render: function() {
            this.onTap.bind(this);
          }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 17 (7.37.5)',
      code: `
        const foo = {
          render: function() {
            this.onTap.bind(this);
            return true;
          }
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 18 (7.37.5)',
      code: `
        const foo = {
          render() {
            this.onTap.bind(this);
            return true;
          }
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 19 (7.37.5)',
      code: `
        class Hello extends Component {
          render() {
            const click = this.onTap.bind(this);
            return <div onClick={onClick}>Hello</div>;
          }
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 20 (7.37.5)',
      code: `
        class Hello extends Component {
          render() {
            foo.onClick = this.onTap.bind(this);
            return <div onClick={onClick}>Hello</div>;
          }
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 21 (7.37.5)',
      code: `
        class Hello extends Component {
          render() {
            return (<div>{
              this.props.list.map(this.wrap.bind(this, "span"))
            }</div>);
          }
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 22 (7.37.5)',
      code: `
        class Hello extends Component {
          render() {
            const click = () => true;
            return <div onClick={onClick}>Hello</div>;
          }
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 23 (7.37.5)',
      code: `
        class Hello extends Component {
          render() {
            return (<div>{
              this.props.list.map(item => <item hello="true"/>)
            }</div>);
          }
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 24 (7.37.5)',
      skip: "The native TypeScript parser cannot parse this upstream bind operator syntax (',' expected.).",
      code: `
        class Hello extends Component {
          render() {
            const click = this.bar::baz
            return <div onClick={onClick}>Hello</div>;
          }
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 25 (7.37.5)',
      skip: "The native TypeScript parser cannot parse this upstream bind operator syntax (',' expected.).",
      code: `
        class Hello extends Component {
          render() {
            return (<div>{
              this.props.list.map(this.bar::baz)
            }</div>);
          }
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 26 (7.37.5)',
      code: `
        var Hello = React.createClass({
          render: function() {
            return (<div>{
              this.props.list.map(this.wrap.bind(this, "span"))
            }</div>);
          }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 27 (7.37.5)',
      skip: "The native TypeScript parser cannot parse this upstream bind operator syntax (',' expected.).",
      code: `
        var Hello = React.createClass({
          render: function() {
            const click = this.bar::baz
            return <div onClick={onClick}>Hello</div>;
          }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 28 (7.37.5)',
      code: `
        var Hello = React.createClass({
          render: function() {
            const click = () => true
            return <div onClick={onClick}>Hello</div>;
          }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 29 (7.37.5)',
      code: `
        class Hello23 extends React.Component {
          renderDiv = () => {
            const onClick = this.doSomething.bind(this, "no")
            return <div onClick={click}>Hello</div>;
          }
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 30 (7.37.5)',
      code: `
        class Hello23 extends React.Component {
          renderDiv = async () => {
            return (<div>{
              this.props.list.map(this.wrap.bind(this, "span"))
            }</div>);
          }
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 32 (7.37.5)',
      code: `<div onClick={this._handleClick.bind(this)}></div>`,
      options: [{ ignoreDOMComponents: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 33 (7.37.5)',
      code: `<div onClick={() => alert("1337")}></div>`,
      options: [{ ignoreDOMComponents: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 34 (7.37.5)',
      code: `<div onClick={function () { alert("1337") }}></div>`,
      options: [{ ignoreDOMComponents: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 35 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream bind operator syntax (Expression expected.).',
      code: `<div foo={::this.onChange} />`,
      options: [{ ignoreDOMComponents: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 36 (7.37.5)',
      code: `
        function click() { return true; }
        class Hello23 extends React.Component {
          renderDiv() {
            return <div onClick={click}>Hello</div>;
          }
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
  ],

  invalid: [
    {
      code: `var x = <div onClick={this._handleClick.bind(this)}></div>;`,
      errors: [expectBindCallError([1, 14, 1, 52])],
    },
    {
      code: `var x = <div onClick={someGlobalFunction.bind(this)}></div>;`,
      errors: [expectBindCallError([1, 14, 1, 53])],
    },
    {
      code: `var x = <div ref={this._refCallback.bind(this)}></div>;`,
      errors: [expectBindCallError([1, 14, 1, 48])],
    },
    {
      code: `var x = <div onClick={() => alert("1337")}></div>;`,
      errors: [expectArrowFuncError([1, 14, 1, 43])],
    },
    {
      code: `var x = <div onClick={async () => alert("1337")}></div>;`,
      errors: [expectArrowFuncError([1, 14, 1, 49])],
    },
    {
      code: `var x = <div onClick={function () { alert("1337"); }}></div>;`,
      errors: [expectFuncError([1, 14, 1, 54])],
    },
    {
      code: `var x = <div onClick={async function () { alert("1337"); }}></div>;`,
      errors: [expectFuncError([1, 14, 1, 60])],
    },
    {
      code: `var x = <div onClick={cond ? onClick.bind(this) : handleClick()}></div>;`,
      errors: [expectBindCallError([1, 14, 1, 65])],
    },
    // ignoreDOMComponents: user components still flagged
    {
      code: `var x = <Foo onClick={this._handleClick.bind(this)} />;`,
      options: [{ ignoreDOMComponents: true }],
      errors: [expectBindCallError([1, 14, 1, 52])],
    },
    // Variable tracking
    {
      code: `
        class Hello23 extends React.Component {
          render() {
            const click = this.someMethod.bind(this);
            return <div onClick={click}>Hello</div>;
          }
        }
      `,
      errors: [expectBindCallError([5, 25, 5, 40])],
    },
    {
      code: `
        class Hello23 extends React.Component {
          render() {
            const click = () => true;
            return <div onClick={click}>Hello</div>;
          }
        }
      `,
      errors: [expectArrowFuncError([5, 25, 5, 40])],
    },
    {
      code: `
        class Hello23 extends React.Component {
          renderDiv() {
            function click() { return true; }
            return <div onClick={click}>Hello</div>;
          }
        }
      `,
      errors: [expectFuncError([5, 25, 5, 40])],
    },
    // Nested blocks
    {
      code: `
        class Hello23 extends React.Component {
          renderDiv() {
            const click = () => true;
            const renderStuff = () => {
              const click = this.doSomething.bind(this, "hey");
              return <div onClick={click} />;
            };
            return <div onClick={click}>Hello</div>;
          }
        }
      `,
      errors: [
        expectBindCallError([7, 27, 7, 42]),
        expectArrowFuncError([9, 25, 9, 40]),
      ],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 0 (7.37.5)',
      code: `<div onClick={this._handleClick.bind(this)}></div>`,
      errors: [expectBindCallError([1, 6, 1, 44])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 1 (7.37.5)',
      code: `<div onClick={someGlobalFunction.bind(this)}></div>`,
      errors: [expectBindCallError([1, 6, 1, 45])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 2 (7.37.5)',
      code: `<div onClick={window.lol.bind(this)}></div>`,
      errors: [expectBindCallError([1, 6, 1, 37])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 3 (7.37.5)',
      code: `<div ref={this._refCallback.bind(this)}></div>`,
      errors: [expectBindCallError([1, 6, 1, 40])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 4 (7.37.5)',
      code: `
        var Hello = createReactClass({
          render: function() {
            const click = this.someMethod.bind(this);
            return <div onClick={click}>Hello {this.state.name}</div>;
          }
        });
      `,
      errors: [expectBindCallError([5, 25, 5, 40])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 5 (7.37.5)',
      code: `
        class Hello23 extends React.Component {
          render() {
            const click = this.someMethod.bind(this);
            return <div onClick={click}>Hello {this.state.name}</div>;
          }
        };
      `,
      errors: [expectBindCallError([5, 25, 5, 40])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 6 (7.37.5)',
      code: `
        class Hello23 extends React.Component {
          renderDiv() {
            const click = this.doSomething.bind(this, "no")
            return <div onClick={click}>Hello</div>;
          }
        };
      `,
      errors: [expectBindCallError([5, 25, 5, 40])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 7 (7.37.5)',
      code: `
        class Hello23 extends React.Component {
          renderDiv = () => {
            const click = this.doSomething.bind(this, "no")
            return <div onClick={click}>Hello</div>;
          }
        };
      `,
      errors: [expectBindCallError([5, 25, 5, 40])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 8 (7.37.5)',
      code: `
        class Hello23 extends React.Component {
          renderDiv = async () => {
            const click = this.doSomething.bind(this, "no")
            return <div onClick={click}>Hello</div>;
          }
        };
      `,
      errors: [expectBindCallError([5, 25, 5, 40])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 9 (7.37.5)',
      code: `
        const foo = {
          render: ({onClick}) => (
            <div onClick={onClick.bind(this)}>Hello</div>
          )
        };
      `,
      errors: [expectBindCallError([4, 18, 4, 46])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 10 (7.37.5)',
      code: `
        var Hello = React.createClass({
          render: function() {
          return <div onClick={this.doSomething.bind(this, "hey")} />
          }
        });
      `,
      errors: [expectBindCallError([4, 23, 4, 67])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 11 (7.37.5)',
      code: `
        var Hello = React.createClass({
          render: function() {
            const doThing = this.doSomething.bind(this, "hey")
            return <div onClick={doThing} />
          }
        });
      `,
      errors: [expectBindCallError([5, 25, 5, 42])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 12 (7.37.5)',
      code: `
        class Hello23 extends React.Component {
          renderDiv = () => {
            const click = () => true
            const renderStuff = () => {
              const click = this.doSomething.bind(this, "hey")
              return <div onClick={click} />
            }
            return <div onClick={click}>Hello</div>;
          }
        };
      `,
      errors: [
        expectBindCallError([7, 27, 7, 42]),
        expectArrowFuncError([9, 25, 9, 40]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 13 (7.37.5)',
      code: `
        const foo = {
          render: ({onClick}) => (
            <div onClick={(returningBoolean()) ? onClick.bind(this) : onClick.bind(this)}>Hello</div>
          )
        };
      `,
      errors: [expectBindCallError([4, 18, 4, 90])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 14 (7.37.5)',
      code: `
        const foo = {
          render: ({onClick}) => (
            <div onClick={(returningBoolean()) ? onClick.bind(this) : handleClick()}>Hello</div>
          )
        };
      `,
      errors: [expectBindCallError([4, 18, 4, 85])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 15 (7.37.5)',
      code: `
        const foo = {
          render: ({onClick}) => (
            <div onClick={(returningBoolean()) ? handleClick() : this.onClick.bind(this)}>Hello</div>
          )
        };
      `,
      errors: [expectBindCallError([4, 18, 4, 90])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 16 (7.37.5)',
      code: `
        const foo = {
          render: ({onClick}) => (
            <div onClick={returningBoolean.bind(this) ? handleClick() : onClick()}>Hello</div>
          )
        };
      `,
      errors: [expectBindCallError([4, 18, 4, 83])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 17 (7.37.5)',
      code: `<div onClick={() => alert("1337")}></div>`,
      errors: [expectArrowFuncError([1, 6, 1, 35])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 18 (7.37.5)',
      code: `<div onClick={async () => alert("1337")}></div>`,
      errors: [expectArrowFuncError([1, 6, 1, 41])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 19 (7.37.5)',
      code: `<div onClick={() => 42}></div>`,
      errors: [expectArrowFuncError([1, 6, 1, 24])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 20 (7.37.5)',
      code: `<div onClick={param => { first(); second(); }}></div>`,
      errors: [expectArrowFuncError([1, 6, 1, 47])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 21 (7.37.5)',
      code: `<div ref={c => this._input = c}></div>`,
      errors: [expectArrowFuncError([1, 6, 1, 32])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 22 (7.37.5)',
      code: `
        class Hello23 extends React.Component {
          renderDiv = () => {
            const click = () => true
            return <div onClick={click}>Hello</div>;
          }
        };
      `,
      errors: [expectArrowFuncError([5, 25, 5, 40])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 23 (7.37.5)',
      code: `
        class Hello23 extends React.Component {
          renderDiv = async () => {
            const click = () => true
            return <div onClick={click}>Hello</div>;
          }
        };
      `,
      errors: [expectArrowFuncError([5, 25, 5, 40])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 24 (7.37.5)',
      code: `
        class Hello23 extends React.Component {
          renderDiv = async () => {
            const click = async () => true
            return <div onClick={click}>Hello</div>;
          }
        };
      `,
      errors: [expectArrowFuncError([5, 25, 5, 40])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 25 (7.37.5)',
      code: `
        var Hello = React.createClass({
          render: function() {
          return <div onClick={() => true} />
          }
        });
      `,
      errors: [expectArrowFuncError([4, 23, 4, 43])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 26 (7.37.5)',
      code: `
        var Hello = React.createClass({
          render: function() {
          return <div onClick={async () => true} />
          }
        });
      `,
      errors: [expectArrowFuncError([4, 23, 4, 49])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 27 (7.37.5)',
      code: `
        var Hello = React.createClass({
          render: function() {
            const doThing = () => true
            return <div onClick={doThing} />
          }
        });
      `,
      errors: [expectArrowFuncError([5, 25, 5, 42])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 28 (7.37.5)',
      code: `
        var Hello = React.createClass({
          render: function() {
            const doThing = async () => true
            return <div onClick={doThing} />
          }
        });
      `,
      errors: [expectArrowFuncError([5, 25, 5, 42])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 29 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream bind operator syntax (Expression expected.).',
      code: `
        class Hello23 extends React.Component {
          renderDiv = () => {
            const click = ::this.onChange
            const renderStuff = () => {
              const click = () => true
              return <div onClick={click} />
            }
            return <div onClick={click}>Hello</div>;
          }
        };
      `,
      errors: [
        expectArrowFuncError([7, 27, 7, 42]),
        expectBindExpressionError([9, 25, 9, 40]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 30 (7.37.5)',
      code: `<div onClick={function () { alert("1337") }}></div>`,
      errors: [expectFuncError([1, 6, 1, 45])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 31 (7.37.5)',
      code: `<div onClick={function * () { alert("1337") }}></div>`,
      errors: [expectFuncError([1, 6, 1, 47])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 32 (7.37.5)',
      code: `<div onClick={async function () { alert("1337") }}></div>`,
      errors: [expectFuncError([1, 6, 1, 51])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 33 (7.37.5)',
      code: `<div ref={function (c) { this._input = c }}></div>`,
      errors: [expectFuncError([1, 6, 1, 44])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 34 (7.37.5)',
      code: `
        class Hello23 extends React.Component {
          renderDiv = () => {
            const click = function () { return true }
            return <div onClick={click}>Hello</div>;
          }
        };
      `,
      errors: [expectFuncError([5, 25, 5, 40])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 35 (7.37.5)',
      code: `
        class Hello23 extends React.Component {
          renderDiv = () => {
            const click = function * () { return true }
            return <div onClick={click}>Hello</div>;
          }
        };
      `,
      errors: [expectFuncError([5, 25, 5, 40])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 36 (7.37.5)',
      code: `
        class Hello23 extends React.Component {
          renderDiv = async () => {
            const click = function () { return true }
            return <div onClick={click}>Hello</div>;
          }
        };
      `,
      errors: [expectFuncError([5, 25, 5, 40])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 37 (7.37.5)',
      code: `
        class Hello23 extends React.Component {
          renderDiv = async () => {
            const click = async function () { return true }
            return <div onClick={click}>Hello</div>;
          }
        };
      `,
      errors: [expectFuncError([5, 25, 5, 40])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 38 (7.37.5)',
      code: `
        var Hello = React.createClass({
          render: function() {
          return <div onClick={function () { return true }} />
          }
        });
      `,
      errors: [expectFuncError([4, 23, 4, 60])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 39 (7.37.5)',
      code: `
        var Hello = React.createClass({
          render: function() {
          return <div onClick={function * () { return true }} />
          }
        });
      `,
      errors: [expectFuncError([4, 23, 4, 62])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 40 (7.37.5)',
      code: `
        var Hello = React.createClass({
          render: function() {
          return <div onClick={async function () { return true }} />
          }
        });
      `,
      errors: [expectFuncError([4, 23, 4, 66])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 41 (7.37.5)',
      code: `
        var Hello = React.createClass({
          render: function() {
            const doThing = function () { return true }
            return <div onClick={doThing} />
          }
        });
      `,
      errors: [expectFuncError([5, 25, 5, 42])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 42 (7.37.5)',
      code: `
        var Hello = React.createClass({
          render: function() {
            const doThing = async function () { return true }
            return <div onClick={doThing} />
          }
        });
      `,
      errors: [expectFuncError([5, 25, 5, 42])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 43 (7.37.5)',
      code: `
        var Hello = React.createClass({
          render: function() {
            const doThing = function * () { return true }
            return <div onClick={doThing} />
          }
        });
      `,
      errors: [expectFuncError([5, 25, 5, 42])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 44 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream bind operator syntax (Expression expected.).',
      code: `
        class Hello23 extends React.Component {
          renderDiv = () => {
            const click = ::this.onChange
            const renderStuff = () => {
              const click = function () { return true }
              return <div onClick={click} />
            }
            return <div onClick={click}>Hello</div>;
          }
        };
      `,
      errors: [
        expectFuncError([7, 27, 7, 42]),
        expectBindExpressionError([9, 25, 9, 40]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 45 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream bind operator syntax (Expression expected.).',
      code: `<div foo={::this.onChange} />`,
      errors: [expectBindExpressionError([1, 6, 1, 27])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 46 (7.37.5)',
      skip: "The native TypeScript parser cannot parse this upstream bind operator syntax ('}' expected.).",
      code: `<div foo={foo.bar::baz} />`,
      errors: [expectBindExpressionError([1, 6, 1, 24])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 47 (7.37.5)',
      skip: "The native TypeScript parser cannot parse this upstream bind operator syntax ('}' expected.).",
      code: `<div foo={foo::bar} />`,
      errors: [expectBindExpressionError([1, 6, 1, 20])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 48 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream bind operator syntax (Expression expected.).',
      code: `
        class Hello23 extends React.Component {
          renderDiv() {
            const click = ::this.onChange
            return <div onClick={click}>Hello</div>;
          }
        };
      `,
      errors: [expectBindExpressionError([5, 25, 5, 40])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 49 (7.37.5)',
      skip: "The native TypeScript parser cannot parse this upstream bind operator syntax (',' expected.).",
      code: `
        class Hello23 extends React.Component {
          renderDiv() {
            const click = this.bar::baz
            return <div onClick={click}>Hello</div>;
          }
        };
      `,
      errors: [expectBindExpressionError([5, 25, 5, 40])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 50 (7.37.5)',
      skip: "The native TypeScript parser cannot parse this upstream bind operator syntax (',' expected.).",
      code: `
        class Hello23 extends React.Component {
          renderDiv = async () => {
            const click = this.bar::baz
            return <div onClick={click}>Hello</div>;
          }
        };
      `,
      errors: [expectBindExpressionError([5, 25, 5, 40])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 51 (7.37.5)',
      skip: "The native TypeScript parser cannot parse this upstream bind operator syntax (',' expected.).",
      code: `
        class Hello23 extends React.Component {
          renderDiv = () => {
            const click = true
            const renderStuff = () => {
              const click = this.bar::baz
              return <div onClick={click} />
            }
            return <div onClick={click}>Hello</div>;
          }
        };
      `,
      errors: [expectBindExpressionError([7, 27, 7, 42])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 52 (7.37.5)',
      code: `
        class Hello23 extends React.Component {
          renderDiv() {
            function click() { return true; }
            return <div onClick={click}>Hello</div>;
          }
        };
      `,
      errors: [expectFuncError([5, 25, 5, 40])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 53 (7.37.5)',
      code: `<Foo onClick={this._handleClick.bind(this)} />`,
      options: [{ ignoreDOMComponents: true }],
      errors: [expectBindCallError([1, 6, 1, 44])],
      languageOptions: upstreamModuleOptions,
    },
  ],
});

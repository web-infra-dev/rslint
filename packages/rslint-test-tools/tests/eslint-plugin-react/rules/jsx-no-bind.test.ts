import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('jsx-no-bind', {} as never, {
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
  ],

  invalid: [
    {
      code: `var x = <div onClick={this._handleClick.bind(this)}></div>;`,
      errors: [{ messageId: 'bindCall' }],
    },
    {
      code: `var x = <div onClick={someGlobalFunction.bind(this)}></div>;`,
      errors: [{ messageId: 'bindCall' }],
    },
    {
      code: `var x = <div ref={this._refCallback.bind(this)}></div>;`,
      errors: [{ messageId: 'bindCall' }],
    },
    {
      code: `var x = <div onClick={() => alert("1337")}></div>;`,
      errors: [{ messageId: 'arrowFunc' }],
    },
    {
      code: `var x = <div onClick={async () => alert("1337")}></div>;`,
      errors: [{ messageId: 'arrowFunc' }],
    },
    {
      code: `var x = <div onClick={function () { alert("1337"); }}></div>;`,
      errors: [{ messageId: 'func' }],
    },
    {
      code: `var x = <div onClick={async function () { alert("1337"); }}></div>;`,
      errors: [{ messageId: 'func' }],
    },
    {
      code: `var x = <div onClick={cond ? onClick.bind(this) : handleClick()}></div>;`,
      errors: [{ messageId: 'bindCall' }],
    },
    // ignoreDOMComponents: user components still flagged
    {
      code: `var x = <Foo onClick={this._handleClick.bind(this)} />;`,
      options: [{ ignoreDOMComponents: true }],
      errors: [{ messageId: 'bindCall' }],
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
      errors: [{ messageId: 'bindCall' }],
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
      errors: [{ messageId: 'arrowFunc' }],
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
      errors: [{ messageId: 'func' }],
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
      errors: [{ messageId: 'bindCall' }, { messageId: 'arrowFunc' }],
    },
  ],
});

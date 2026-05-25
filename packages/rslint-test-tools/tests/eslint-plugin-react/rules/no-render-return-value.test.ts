import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-render-return-value', {} as never, {
  valid: [
    // ---- Upstream: render with no return-value consumption ----
    { code: `ReactDOM.render(<div />, document.body);` },
    {
      code: `
        let node;
        ReactDOM.render(<div ref={ref => node = ref}/>, document.body);
      `,
    },
    { code: `var foo = render(<div />, root)` },
    { code: `var foo = ReactDom.renderder(<div />, root)` },

    // ---- Edge: bracket access with string / numeric / template / other-identifier key — upstream's `'name' in property && property.name === 'render'` guard fails ----
    { code: `var x = ReactDOM['render'](<div />, document.body);` },
    { code: `var x = ReactDOM[0](<div />, document.body);` },
    {
      code: `var renderFn; var x = ReactDOM[renderFn](<div />, document.body);`,
    },
    { code: 'var x = ReactDOM[`render`](<div />, document.body);' },

    // ---- Edge: React at default version (>= 15.0.0) — must be ReactDOM ----
    { code: `var x = React.render(<div />, document.body);` },

    // ---- Non-consuming positions ----
    { code: `doSomething(ReactDOM.render(<div />, document.body));` },
    { code: `var x = cond ? ReactDOM.render(<div />, document.body) : null;` },
    { code: `var x = y || ReactDOM.render(<div />, document.body);` },
    { code: `var x = y ?? ReactDOM.render(<div />, document.body);` },
    { code: `var x = obj.ReactDOM.render(<div />, document.body);` },
    { code: `var x = new ReactDOM.render(<div />, document.body);` },
    {
      code: `async function f() { await ReactDOM.render(<div />, document.body); }`,
    },
    {
      code: `function* g() { yield ReactDOM.render(<div />, document.body); }`,
    },
    { code: `function h() { throw ReactDOM.render(<div />, document.body); }` },
    { code: `var x = typeof ReactDOM.render(<div />, document.body);` },
    { code: `var xs = [ReactDOM.render(<div />, document.body)];` },
    { code: `var o = { ...ReactDOM.render(<div />, document.body) };` },
    { code: `ReactDOM.render(<div />, document.body).then(inst => inst);` },
    { code: `var x = ReactDOM.render(<div />, document.body).foo;` },
    {
      code: `var x = (foo(), ReactDOM.render(<div />, document.body), 1);`,
    },

    // ---- TS wrappers (AsExpression / NonNullExpression) — upstream skips, we match ----
    { code: `var x = ReactDOM.render(<div />, document.body) as any;` },
    { code: `var x = ReactDOM.render(<div />, document.body)!;` },
  ],
  invalid: [
    // ---- Upstream: VariableDeclarator ----
    {
      code: `var Hello = ReactDOM.render(<div />, document.body);`,
      errors: [{ messageId: 'noReturnValue' }],
    },
    // ---- Upstream: Object property value ----
    {
      code: `
        var o = {
          inst: ReactDOM.render(<div />, document.body)
        };
      `,
      errors: [{ messageId: 'noReturnValue' }],
    },
    // ---- Upstream: ReturnStatement ----
    {
      code: `
        function render () {
          return ReactDOM.render(<div />, document.body)
        }
      `,
      errors: [{ messageId: 'noReturnValue' }],
    },
    // ---- Upstream: ArrowFunctionExpression body ----
    {
      code: `var render = (a, b) => ReactDOM.render(a, b)`,
      errors: [{ messageId: 'noReturnValue' }],
    },
    // ---- Upstream: AssignmentExpression (member-expression LHS) ----
    {
      code: `this.o = ReactDOM.render(<div />, document.body);`,
      errors: [{ messageId: 'noReturnValue' }],
    },
    // ---- Upstream: AssignmentExpression (identifier LHS) ----
    {
      code: `var v; v = ReactDOM.render(<div />, document.body);`,
      errors: [{ messageId: 'noReturnValue' }],
    },

    // ---- Parens transparent ----
    {
      code: `var x = (ReactDOM.render(<div />, document.body));`,
      errors: [{ messageId: 'noReturnValue' }],
    },
    {
      code: `var x = ((ReactDOM.render(<div />, document.body)));`,
      errors: [{ messageId: 'noReturnValue' }],
    },
    {
      code: `var f = () => (ReactDOM.render(<div />, document.body))`,
      errors: [{ messageId: 'noReturnValue' }],
    },

    // ---- Nested arrow — inner body is the call ----
    {
      code: `var f = () => () => ReactDOM.render(<div />, document.body);`,
      errors: [{ messageId: 'noReturnValue' }],
    },

    // ---- Compound / logical / chained / destructuring assignment ----
    {
      code: `var v = 0; v += ReactDOM.render(<div />, document.body);`,
      errors: [{ messageId: 'noReturnValue' }],
    },
    {
      code: `var v; v ??= ReactDOM.render(<div />, document.body);`,
      errors: [{ messageId: 'noReturnValue' }],
    },
    {
      code: `var a, b; a = b = ReactDOM.render(<div />, document.body);`,
      errors: [{ messageId: 'noReturnValue' }],
    },
    {
      code: `var o; ({ a: o } = ReactDOM.render(<div />, document.body));`,
      errors: [{ messageId: 'noReturnValue' }],
    },

    // ---- Class-method return + class-field arrow body ----
    {
      code: `
        class App {
          foo() {
            return ReactDOM.render(<div />, document.body);
          }
        }
      `,
      errors: [{ messageId: 'noReturnValue' }],
    },
    {
      code: `
        class App {
          cb = () => ReactDOM.render(<div />, document.body);
        }
      `,
      errors: [{ messageId: 'noReturnValue' }],
    },

    // ---- TSX function-component return ----
    {
      code: `
        function App() {
          return ReactDOM.render(<div />, document.body);
        }
      `,
      errors: [{ messageId: 'noReturnValue' }],
    },

    // ---- Bracket access with an Identifier literally named `render` — upstream's `property.name === 'render'` matches ----
    {
      code: `var render; var x = ReactDOM[render](<div />, document.body);`,
      errors: [{ messageId: 'noReturnValue' }],
    },
  ],
});

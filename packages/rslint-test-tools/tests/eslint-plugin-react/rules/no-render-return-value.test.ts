import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, upstreamModuleOptions } from '../config';

function expectNoReturnValueError1(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'noReturnValue',
    message: 'Do not depend on the return value from ReactDOM.render',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectNoReturnValueError2(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'noReturnValue',
    message: 'Do not depend on the return value from React.render',
    line,
    column,
    endLine,
    endColumn,
  };
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('no-render-return-value', {
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

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 2 (7.37.5)',
      code: `ReactDOM.render(<div ref={ref => this.node = ref}/>, document.body);`,
      settings: { react: { version: '0.14.0' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 3 (7.37.5)',
      code: `React.render(<div ref={ref => this.node = ref}/>, document.body);`,
      settings: { react: { version: '0.14.0' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 4 (7.37.5)',
      code: `React.render(<div ref={ref => this.node = ref}/>, document.body);`,
      settings: { react: { version: '0.13.0' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 5 (7.37.5)',
      code: `var foo = React.render(<div />, root);`,
      settings: { react: { version: '0.0.1' } },
      languageOptions: upstreamModuleOptions,
    },
  ],
  invalid: [
    // ---- Upstream: VariableDeclarator ----
    {
      code: `var Hello = ReactDOM.render(<div />, document.body);`,
      errors: [expectNoReturnValueError1([1, 13, 1, 28])],
    },
    // ---- Upstream: Object property value ----
    {
      code: `
        var o = {
          inst: ReactDOM.render(<div />, document.body)
        };
      `,
      errors: [expectNoReturnValueError1([3, 17, 3, 32])],
    },
    // ---- Upstream: ReturnStatement ----
    {
      code: `
        function render () {
          return ReactDOM.render(<div />, document.body)
        }
      `,
      errors: [expectNoReturnValueError1([3, 18, 3, 33])],
    },
    // ---- Upstream: ArrowFunctionExpression body ----
    {
      code: `var render = (a, b) => ReactDOM.render(a, b)`,
      errors: [expectNoReturnValueError1([1, 24, 1, 39])],
    },
    // ---- Upstream: AssignmentExpression (member-expression LHS) ----
    {
      code: `this.o = ReactDOM.render(<div />, document.body);`,
      errors: [expectNoReturnValueError1([1, 10, 1, 25])],
    },
    // ---- Upstream: AssignmentExpression (identifier LHS) ----
    {
      code: `var v; v = ReactDOM.render(<div />, document.body);`,
      errors: [expectNoReturnValueError1([1, 12, 1, 27])],
    },

    // ---- Upstream: defaultVersion is used when version is omitted ----
    {
      code: `var instance = React.render(<div />, document.body);`,
      settings: { react: { defaultVersion: '0.13.0' } },
      errors: [expectNoReturnValueError2([1, 16, 1, 28])],
    },

    // ---- Upstream: a computed property key consumes the render result ----
    {
      code: `var value = { [ReactDOM.render(<div />, document.body)]: true };`,
      errors: [expectNoReturnValueError1([1, 16, 1, 31])],
    },

    // ---- Parens transparent ----
    {
      code: `var x = (ReactDOM.render(<div />, document.body));`,
      errors: [expectNoReturnValueError1([1, 10, 1, 25])],
    },
    {
      code: `var x = ((ReactDOM.render(<div />, document.body)));`,
      errors: [expectNoReturnValueError1([1, 11, 1, 26])],
    },
    {
      code: `var f = () => (ReactDOM.render(<div />, document.body))`,
      errors: [expectNoReturnValueError1([1, 16, 1, 31])],
    },

    // ---- Nested arrow — inner body is the call ----
    {
      code: `var f = () => () => ReactDOM.render(<div />, document.body);`,
      errors: [expectNoReturnValueError1([1, 21, 1, 36])],
    },

    // ---- Compound / logical / chained / destructuring assignment ----
    {
      code: `var v = 0; v += ReactDOM.render(<div />, document.body);`,
      errors: [expectNoReturnValueError1([1, 17, 1, 32])],
    },
    {
      code: `var v; v ??= ReactDOM.render(<div />, document.body);`,
      errors: [expectNoReturnValueError1([1, 14, 1, 29])],
    },
    {
      code: `var a, b; a = b = ReactDOM.render(<div />, document.body);`,
      errors: [expectNoReturnValueError1([1, 19, 1, 34])],
    },
    {
      code: `var o; ({ a: o } = ReactDOM.render(<div />, document.body));`,
      errors: [expectNoReturnValueError1([1, 20, 1, 35])],
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
      errors: [expectNoReturnValueError1([4, 20, 4, 35])],
    },
    {
      code: `
        class App {
          cb = () => ReactDOM.render(<div />, document.body);
        }
      `,
      errors: [expectNoReturnValueError1([3, 22, 3, 37])],
    },

    // ---- TSX function-component return ----
    {
      code: `
        function App() {
          return ReactDOM.render(<div />, document.body);
        }
      `,
      errors: [expectNoReturnValueError1([3, 18, 3, 33])],
    },

    // ---- Bracket access with an Identifier literally named `render` — upstream's `property.name === 'render'` matches ----
    {
      code: `var render; var x = ReactDOM[render](<div />, document.body);`,
      errors: [expectNoReturnValueError1([1, 21, 1, 37])],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 6 (7.37.5)',
      code: `var inst = React.render(<div />, document.body);`,
      settings: { react: { version: '0.14.0' } },
      errors: [expectNoReturnValueError2([1, 12, 1, 24])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 7 (7.37.5)',
      code: `var inst = ReactDOM.render(<div />, document.body);`,
      settings: { react: { version: '0.14.0' } },
      errors: [expectNoReturnValueError1([1, 12, 1, 27])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 8 (7.37.5)',
      code: `var inst = React.render(<div />, document.body);`,
      settings: { react: { version: '0.13.0' } },
      errors: [expectNoReturnValueError2([1, 12, 1, 24])],
      languageOptions: upstreamModuleOptions,
    },
  ],
});

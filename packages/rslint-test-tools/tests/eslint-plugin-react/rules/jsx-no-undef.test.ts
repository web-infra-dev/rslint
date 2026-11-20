import { RuleTester } from '@rslint/rule-tester';
import {
  fixtureConfig,
  upstreamScriptOptions,
  upstreamModuleOptions,
} from '../config';
import { diagnostic } from '../expectations';

function expectUndefinedError1(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic('undefined', "'App' is not defined.", range);
}

function expectUndefinedError2(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic('undefined', "'appp' is not defined.", range);
}

function expectUndefinedError3(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic('undefined', "'Foo' is not defined.", range);
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('jsx-no-undef', {
  valid: [
    // ---- Upstream valid cases ----
    { code: `var React, App; React.render(<App />);` },
    { code: `var React; React.render(<img />);` },
    { code: `var React; React.render(<x-gif />);` },
    { code: `var React, app; React.render(<app.Foo />);` },
    { code: `var React, app; React.render(<app.foo.Bar />);` },
    { code: `var React; React.render(<Apppp:Foo />);` },
    {
      code: `
        var React: any;
        class Hello extends React.Component {
          render() {
            return <this.props.tag />
          }
        }
      `,
    },
    {
      code: `
        import Text from "cool-module";
        const TextWrapper = function (props: any) {
          return (
            <Text />
          );
        };
      `,
      options: [{ allowGlobals: false }],
    },

    // ---- ThisKeyword forms ----
    { code: `var x = <this />;` },
    { code: `var x = <this.Foo />;` },
    { code: `var x = <this.a.b.c />;` },

    // ---- DOM / lowercase ----
    { code: `var x = <foo-bar />;` },
    { code: `var x = <my-elem-2 />;` },

    // ---- Declaration kinds ----
    { code: `const _Foo = () => null; var x = <_Foo />;` },
    { code: `let Foo: any; var x = <Foo />;` },
    { code: `const Foo: any = null; var x = <Foo />;` },
    { code: `class Foo {} var x = <Foo />;` },
    { code: `function Foo() { return null; } var x = <Foo />;` },
    { code: `enum Foo { A } var x = <Foo />;` },
    {
      code: `namespace Foo { export const Bar: any = null; } var x = <Foo.Bar />;`,
    },
    { code: `import { Foo } from "cool-module"; var x = <Foo />;` },
    { code: `import * as NS from "cool-module"; var x = <NS.Foo />;` },
    { code: `import Foo from "cool-module"; var x = <Foo />;` },
    { code: `const foo: any = {}; var x = <foo.Bar />;` },
    { code: `const a: any = {}; var x = <a.b.c.d.e.f />;` },

    // ---- Nested scopes ----
    {
      code: `function render() { const App = () => null; return <App />; }`,
    },
    {
      code: `function outer() { const App: any = null; return function inner() { return <App />; }; }`,
    },
    { code: `function render(app: any) { return <app.Foo />; }` },
    { code: `try {} catch (e: any) { var x = <e.Foo />; }` },
    { code: `for (let Foo: any; false; ) { var x = <Foo />; }` },
    {
      code: `const xs: any[] = []; for (const Foo of xs) { var x = <Foo />; }`,
    },

    // ---- Fragment ----
    { code: `var x = <></>;` },
    { code: `var x = <><div /></>;` },

    // ---- JSX in non-top positions ----
    {
      code: `const Outer: any = null; const Inner: any = null; var x = <Outer attr={<Inner />} />;`,
    },
    {
      code: `const A: any = null; const B: any = null; var x = true ? <A /> : <B />;`,
    },
    { code: `const f: any = null; const A: any = null; f(<A />);` },
    { code: `const A: any = null; var x = { k: <A /> };` },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 0 (7.37.5)',
      code: `/*eslint no-undef:1*/ var React, App; React.render(<App />);`,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 1 (7.37.5)',
      code: `/*eslint no-undef:1*/ var React; React.render(<img />);`,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 2 (7.37.5)',
      code: `/*eslint no-undef:1*/ var React; React.render(<x-gif />);`,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 3 (7.37.5)',
      code: `/*eslint no-undef:1*/ var React, app; React.render(<app.Foo />);`,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 4 (7.37.5)',
      code: `/*eslint no-undef:1*/ var React, app; React.render(<app.foo.Bar />);`,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 5 (7.37.5)',
      code: `/*eslint no-undef:1*/ var React; React.render(<Apppp:Foo />);`,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 6 (7.37.5)',
      code: `
        /*eslint no-undef:1*/
        var React;
        class Hello extends React.Component {
          render() {
            return <this.props.tag />
          }
        }
      `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 7 (7.37.5)',
      code: `var React; React.render(<Text />);`,
      languageOptions: {
        ecmaVersion: 2018,
        sourceType: 'script',
        parserOptions: { ecmaFeatures: { jsx: true } },
        globals: { Text: true },
      },
    },
    {
      name: 'Upstream valid 8 (7.37.5)',
      code: `
        import Text from "cool-module";
        const TextWrapper = function (props) {
          return (
            <Text />
          );
        };
      `,
      options: [{ allowGlobals: false }],
      languageOptions: upstreamModuleOptions,
    },
  ],
  invalid: [
    // ---- Upstream invalid cases ----
    {
      code: `var React: any; React.render(<App />);`,
      errors: [expectUndefinedError1([1, 31, 1, 34])],
    },
    {
      code: `var React: any; React.render(<Appp.Foo />);`,
      errors: [
        diagnostic('undefined', "'Appp' is not defined.", [1, 31, 1, 35]),
      ],
    },
    {
      code: `var React: any; React.render(<appp.Foo />);`,
      errors: [expectUndefinedError2([1, 31, 1, 35])],
    },
    {
      code: `var React: any; React.render(<appp.foo.Bar />);`,
      errors: [expectUndefinedError2([1, 31, 1, 35])],
    },
    {
      code: `
        const TextWrapper = function (props: any) {
          return (
            <Text />
          );
        };
        export default TextWrapper;
      `,
      options: [{ allowGlobals: false }],
      errors: [
        diagnostic('undefined', "'Text' is not defined.", [4, 14, 4, 18]),
      ],
    },
    {
      code: `var React: any; React.render(<Foo />);`,
      errors: [expectUndefinedError3([1, 31, 1, 34])],
    },

    // ---- Additional edge cases ----
    {
      code: `<Bar></Bar>;`,
      errors: [diagnostic('undefined', "'Bar' is not defined.", [1, 2, 1, 5])],
    },
    {
      code: `var x = <Outer.Inner />;`,
      errors: [
        diagnostic('undefined', "'Outer' is not defined.", [1, 10, 1, 15]),
      ],
    },
    {
      code: `var x = <_Undeclared />;`,
      errors: [
        diagnostic(
          'undefined',
          "'_Undeclared' is not defined.",
          [1, 10, 1, 21],
        ),
      ],
    },
    {
      code: `var x = <Undeclared />;`,
      options: [{ allowGlobals: true }],
      errors: [
        diagnostic('undefined', "'Undeclared' is not defined.", [1, 10, 1, 20]),
      ],
    },
    {
      code: `const Outer: any = null; var x = <Outer attr={<InnerBad />}><ChildBad /></Outer>;`,
      errors: [
        diagnostic('undefined', "'InnerBad' is not defined.", [1, 48, 1, 56]),
        diagnostic('undefined', "'ChildBad' is not defined.", [1, 62, 1, 70]),
      ],
    },
    {
      code: `var x = true ? <A /> : <B />;`,
      errors: [
        diagnostic('undefined', "'A' is not defined.", [1, 17, 1, 18]),
        diagnostic('undefined', "'B' is not defined.", [1, 25, 1, 26]),
      ],
    },
    {
      code: `var x = <><Undef /></>;`,
      errors: [
        diagnostic('undefined', "'Undef' is not defined.", [1, 12, 1, 17]),
      ],
    },
    {
      code: `function other() { const App: any = null; } function render() { return <App />; }`,
      errors: [expectUndefinedError1([1, 73, 1, 76])],
    },
    {
      code: `try {} catch (e: any) {} var x = <e.Foo />;`,
      errors: [diagnostic('undefined', "'e' is not defined.", [1, 35, 1, 36])],
    },
    {
      code: `for (let Foo: any; false; ) {} var x = <Foo />;`,
      errors: [expectUndefinedError3([1, 41, 1, 44])],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 0 (7.37.5)',
      code: `/*eslint no-undef:1*/ var React; React.render(<App />);`,
      errors: [expectUndefinedError1([1, 48, 1, 51])],
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream invalid 1 (7.37.5)',
      code: `/*eslint no-undef:1*/ var React; React.render(<Appp.Foo />);`,
      errors: [
        diagnostic('undefined', "'Appp' is not defined.", [1, 48, 1, 52]),
      ],
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream invalid 2 (7.37.5)',
      code: `/*eslint no-undef:1*/ var React; React.render(<appp.Foo />);`,
      errors: [expectUndefinedError2([1, 48, 1, 52])],
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream invalid 3 (7.37.5)',
      code: `/*eslint no-undef:1*/ var React; React.render(<appp.foo.Bar />);`,
      errors: [expectUndefinedError2([1, 48, 1, 52])],
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream invalid 4 (7.37.5)',
      code: `
        const TextWrapper = function (props) {
          return (
            <Text />
          );
        };
        export default TextWrapper;
      `,
      options: [{ allowGlobals: false }],
      languageOptions: {
        ecmaVersion: 2018,
        sourceType: 'module',
        parserOptions: { ecmaFeatures: { jsx: true } },
        globals: { Text: true },
      },
      errors: [
        diagnostic('undefined', "'Text' is not defined.", [4, 14, 4, 18]),
      ],
    },
    {
      name: 'Upstream invalid 5 (7.37.5)',
      code: `/*eslint no-undef:1*/ var React; React.render(<Foo />);`,
      errors: [expectUndefinedError3([1, 48, 1, 51])],
      languageOptions: upstreamScriptOptions,
    },
  ],
});

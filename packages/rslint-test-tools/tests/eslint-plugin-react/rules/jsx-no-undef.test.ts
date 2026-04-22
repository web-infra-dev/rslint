import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('jsx-no-undef', {} as never, {
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
  ],
  invalid: [
    // ---- Upstream invalid cases ----
    {
      code: `var React: any; React.render(<App />);`,
      errors: [{ message: `'App' is not defined.` }],
    },
    {
      code: `var React: any; React.render(<Appp.Foo />);`,
      errors: [{ message: `'Appp' is not defined.` }],
    },
    {
      code: `var React: any; React.render(<appp.Foo />);`,
      errors: [{ message: `'appp' is not defined.` }],
    },
    {
      code: `var React: any; React.render(<appp.foo.Bar />);`,
      errors: [{ message: `'appp' is not defined.` }],
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
      errors: [{ message: `'Text' is not defined.` }],
    },
    {
      code: `var React: any; React.render(<Foo />);`,
      errors: [{ message: `'Foo' is not defined.` }],
    },

    // ---- Additional edge cases ----
    { code: `<Bar></Bar>;`, errors: [{ message: `'Bar' is not defined.` }] },
    {
      code: `var x = <Outer.Inner />;`,
      errors: [{ message: `'Outer' is not defined.` }],
    },
    {
      code: `var x = <_Undeclared />;`,
      errors: [{ message: `'_Undeclared' is not defined.` }],
    },
    {
      code: `var x = <Undeclared />;`,
      options: [{ allowGlobals: true }],
      errors: [{ message: `'Undeclared' is not defined.` }],
    },
    {
      code: `const Outer: any = null; var x = <Outer attr={<InnerBad />}><ChildBad /></Outer>;`,
      errors: [
        { message: `'InnerBad' is not defined.` },
        { message: `'ChildBad' is not defined.` },
      ],
    },
    {
      code: `var x = true ? <A /> : <B />;`,
      errors: [
        { message: `'A' is not defined.` },
        { message: `'B' is not defined.` },
      ],
    },
    {
      code: `var x = <><Undef /></>;`,
      errors: [{ message: `'Undef' is not defined.` }],
    },
    {
      code: `function other() { const App: any = null; } function render() { return <App />; }`,
      errors: [{ message: `'App' is not defined.` }],
    },
    {
      code: `try {} catch (e: any) {} var x = <e.Foo />;`,
      errors: [{ message: `'e' is not defined.` }],
    },
    {
      code: `for (let Foo: any; false; ) {} var x = <Foo />;`,
      errors: [{ message: `'Foo' is not defined.` }],
    },
  ],
});

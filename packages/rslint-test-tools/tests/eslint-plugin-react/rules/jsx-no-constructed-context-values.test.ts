// Upstream: https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/tests/lib/rules/jsx-no-constructed-context-values.js
// All upstream cases and documentation examples; parser variants share cases.
// The React wrapper checks counts and messages; Go tests assert IDs and ranges.
import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();
ruleTester.run('jsx-no-constructed-context-values', {} as never, {
  valid: [
    // Upstream valid 1.
    {
      code: 'const Component = () => <Context.Provider value={props}></Context.Provider>',
    },
    // Upstream valid 2.
    {
      code: 'const Component = () => <Context.Provider value={100}></Context.Provider>',
    },
    // Upstream valid 3.
    {
      code: 'const Component = () => <Context.Provider value="Some string"></Context.Provider>',
    },
    // Upstream valid 4.
    {
      code: 'function Component() { const foo = useMemo(() => { return {} }, []); return (<Context.Provider value={foo}></Context.Provider>)}',
      options: [
        {
          allowArrowFunctions: true,
        },
      ],
    },
    // Upstream valid 5.
    {
      code: '\n        function Component({oneProp, twoProp, redProp, blueProp,}) {\n          return (\n            <NewContext.Provider value={twoProp}></NewContext.Provider>\n          );\n        }\n      ',
    },
    // Upstream valid 6.
    {
      code: '\n        function Foo(section) {\n          const foo = section.section_components?.edges;\n\n          return (\n            <Context.Provider value={foo}></Context.Provider>\n          )\n        }\n      ',
    },
    // Upstream valid 7.
    {
      code: "\n        import foo from 'foo';\n        function innerContext() {\n          return (\n            <Context.Provider value={foo.something}></Context.Provider>\n          )\n        }\n      ",
    },
    // Upstream valid 8.
    {
      code: "\n        // Passes because the lint rule doesn't handle JSX spread attributes\n        function innerContext() {\n          const foo = {value: 'something'}\n          return (\n            <Context.Provider {...foo}></Context.Provider>\n          )\n        }\n      ",
    },
    // Upstream valid 9.
    {
      code: "\n        // Passes because the lint rule doesn't handle JSX spread attributes\n        function innerContext() {\n          const foo = useMemo(() => {\n            return bar;\n          })\n          return (\n            <Context.Provider value={foo}></Context.Provider>\n          )\n        }\n      ",
    },
    // Upstream valid 10.
    {
      code: "\n        // Passes because we can't statically check if it's using the default value\n        function Component({ a = {} }) {\n          return (<Context.Provider value={a}></Context.Provider>);\n        }\n      ",
    },
    // Upstream valid 11.
    {
      code: "\n          import React from 'react';\n          import MyContext from './MyContext';\n\n          const value = '';\n\n          function ContextProvider(props) {\n              return (\n                  <MyContext.Provider value={value as any}>\n                      {props.children}\n                  </MyContext.Provider>\n              )\n          }\n        ",
    },
    // Upstream valid 12.
    {
      code: "\n        import React from 'react';\n        import BooleanContext from './BooleanContext';\n\n        function ContextProvider(props) {\n            return (\n                <BooleanContext.Provider value>\n                    {props.children}\n                </BooleanContext.Provider>\n            )\n        }\n      ",
    },
    // Upstream valid 13.
    {
      code: "\n        const root = ReactDOM.createRoot(document.getElementById('root'));\n        root.render(\n          <AppContext.Provider value={{}}>\n            <AppView />\n          </AppContext.Provider>\n        );\n      ",
    },
    // Upstream valid 14.
    {
      code: "\n        // Passes because the context is not a provider\n        function Component() {\n          return <MyContext.Consumer value={{ foo: 'bar' }} />;\n        }\n      ",
    },
    // Upstream valid 15.
    {
      code: "\n        import React from 'react';\n\n        const MyContext = React.createContext();\n        const Component = () => <MyContext value={props}></MyContext>;\n      ",
    },
    // Upstream valid 16.
    {
      code: "\n        import React from 'react';\n\n        const MyContext = React.createContext();\n        const Component = () => <MyContext value={100}></MyContext>;\n      ",
    },
    // Upstream valid 17.
    {
      code: '\n        const SomeContext = createContext();\n        const Component = () => <SomeContext value="Some string"></SomeContext>;\n      ',
    },
    // Upstream valid 18.
    {
      code: '\n        // Passes because MyContext is not a variable declarator\n        function Component({ MyContext }) {\n          return <MyContext value={{ foo: "bar" }} />;\n        }\n      ',
    },
    // Documentation: memoized value (return fragment wrapped in a component).
    {
      code: "function Component() {\n  const foo = useMemo(() => ({foo: 'bar'}), []);\n  return (\n    <SomeContext.Provider value={foo}>\n      ...\n    </SomeContext.Provider>\n  );\n}",
    },
    // Documentation: string value (closing-tag typo corrected).
    {
      code: 'const SomeContext = createContext();\nconst Component = () => <SomeContext value="Some string"></SomeContext>;',
    },
  ],
  invalid: [
    // Upstream invalid 1.
    {
      code: 'function Component() { const foo = {}; return (<Context.Provider value={foo}></Context.Provider>) }',
      errors: [
        {
          messageId: 'withIdentifierMsg',
          message:
            "The 'foo' object (at line 1) passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.",
          line: 1,
          column: 36,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    // Upstream invalid 2.
    {
      code: 'function Component() { const foo = []; return (<Context.Provider value={foo}></Context.Provider>) }',
      errors: [
        {
          messageId: 'withIdentifierMsg',
          message:
            "The 'foo' array (at line 1) passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.",
          line: 1,
          column: 36,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    // Upstream invalid 3.
    {
      code: 'function Component() { const foo = () => {}; return (<Context.Provider value={foo}></Context.Provider>)}',
      errors: [
        {
          messageId: 'withIdentifierMsgFunc',
          message:
            "The 'foo' function expression (at line 1) passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useCallback hook.",
          line: 1,
          column: 36,
          endLine: 1,
          endColumn: 44,
        },
      ],
    },
    // Upstream invalid 4.
    {
      code: 'function Component() { const foo = function bar(){}; return (<Context.Provider value={foo}></Context.Provider>)}',
      errors: [
        {
          messageId: 'withIdentifierMsgFunc',
          message:
            "The 'foo' function expression (at line 1) passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useCallback hook.",
          line: 1,
          column: 36,
          endLine: 1,
          endColumn: 52,
        },
      ],
    },
    // Upstream invalid 5.
    {
      code: 'function Component() { const foo = class SomeClass{}; return (<Context.Provider value={foo}></Context.Provider>)}',
      errors: [
        {
          messageId: 'withIdentifierMsg',
          message:
            "The 'foo' class expression (at line 1) passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.",
          line: 1,
          column: 36,
          endLine: 1,
          endColumn: 53,
        },
      ],
    },
    // Upstream invalid 6.
    {
      code: 'function Component() { const foo = new SomeClass(); return (<Context.Provider value={foo}></Context.Provider>)}',
      errors: [
        {
          messageId: 'withIdentifierMsg',
          message:
            "The 'foo' new expression (at line 1) passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.",
          line: 1,
          column: 36,
          endLine: 1,
          endColumn: 51,
        },
      ],
    },
    // Upstream invalid 7.
    {
      code: 'function Component() { function foo() {}; return (<Context.Provider value={foo}></Context.Provider>)}',
      errors: [
        {
          messageId: 'withIdentifierMsgFunc',
          message:
            "The 'foo' function declaration (at line 1) passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useCallback hook.",
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 41,
        },
      ],
    },
    // Upstream invalid 8.
    {
      code: 'function Component() { const foo = true ? {} : "fine"; return (<Context.Provider value={foo}></Context.Provider>)}',
      errors: [
        {
          messageId: 'withIdentifierMsg',
          message:
            "The 'foo' object (at line 1) passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.",
          line: 1,
          column: 43,
          endLine: 1,
          endColumn: 45,
        },
      ],
    },
    // Upstream invalid 9.
    {
      code: 'function Component() { const foo = bar || {}; return (<Context.Provider value={foo}></Context.Provider>)}',
      errors: [
        {
          messageId: 'withIdentifierMsg',
          message:
            "The 'foo' object (at line 1) passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.",
          line: 1,
          column: 43,
          endLine: 1,
          endColumn: 45,
        },
      ],
    },
    // Upstream invalid 10.
    {
      code: 'function Component() { const foo = bar && {}; return (<Context.Provider value={foo}></Context.Provider>)}',
      errors: [
        {
          messageId: 'withIdentifierMsg',
          message:
            "The 'foo' object (at line 1) passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.",
          line: 1,
          column: 43,
          endLine: 1,
          endColumn: 45,
        },
      ],
    },
    // Upstream invalid 11.
    {
      code: 'function Component() { const foo = bar ? baz ? {} : null : null; return (<Context.Provider value={foo}></Context.Provider>)}',
      errors: [
        {
          messageId: 'withIdentifierMsg',
          message:
            "The 'foo' object (at line 1) passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.",
          line: 1,
          column: 48,
          endLine: 1,
          endColumn: 50,
        },
      ],
    },
    // Upstream invalid 12.
    {
      code: 'function Component() { let foo = {}; return (<Context.Provider value={foo}></Context.Provider>) }',
      errors: [
        {
          messageId: 'withIdentifierMsg',
          message:
            "The 'foo' object (at line 1) passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.",
          line: 1,
          column: 34,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },
    // Upstream invalid 13.
    {
      code: 'function Component() { var foo = {}; return (<Context.Provider value={foo}></Context.Provider>)}',
      errors: [
        {
          messageId: 'withIdentifierMsg',
          message:
            "The 'foo' object (at line 1) passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.",
          line: 1,
          column: 34,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },
    // Upstream invalid 14.
    {
      code: '\n        function Component() {\n          let a = {};\n          a = 10;\n          return (<Context.Provider value={a}></Context.Provider>);\n        }\n      ',
      errors: [
        {
          messageId: 'withIdentifierMsg',
          message:
            "The 'a' object (at line 3) passed as the value prop to the Context provider (at line 5) changes every render. To fix this consider wrapping it in a useMemo hook.",
          line: 3,
          column: 19,
          endLine: 3,
          endColumn: 21,
        },
      ],
    },
    // Upstream invalid 15.
    {
      code: '\n        function Component() {\n          const foo = {};\n          const bar = foo;\n          return (<Context.Provider value={bar}></Context.Provider>);\n        }\n      ',
      errors: [
        {
          messageId: 'withIdentifierMsg',
          message:
            "The 'bar' object (at line 3) passed as the value prop to the Context provider (at line 5) changes every render. To fix this consider wrapping it in a useMemo hook.",
          line: 3,
          column: 23,
          endLine: 3,
          endColumn: 25,
        },
      ],
    },
    // Upstream invalid 16.
    {
      code: '\n        function Component(foo) {\n          let bar = true ? foo : {};\n          return (<Context.Provider value={bar}></Context.Provider>);\n        }\n      ',
      errors: [
        {
          messageId: 'withIdentifierMsg',
          message:
            "The 'bar' object (at line 3) passed as the value prop to the Context provider (at line 4) changes every render. To fix this consider wrapping it in a useMemo hook.",
          line: 3,
          column: 34,
          endLine: 3,
          endColumn: 36,
        },
      ],
    },
    // Upstream invalid 17.
    {
      code: 'function Component() { return (<Context.Provider value={{foo: "bar"}}></Context.Provider>);}',
      errors: [
        {
          messageId: 'defaultMsg',
          message:
            'The object passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.',
          line: 1,
          column: 57,
          endLine: 1,
          endColumn: 69,
        },
      ],
    },
    // Upstream invalid 18.
    {
      code: 'function Component() { const Wrapper = (<SomeComp />); return (<Context.Provider value={Wrapper}></Context.Provider>);}',
      errors: [
        {
          messageId: 'withIdentifierMsg',
          message:
            "The 'Wrapper' JSX element (at line 1) passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.",
          line: 1,
          column: 41,
          endLine: 1,
          endColumn: 53,
        },
      ],
    },
    // Upstream invalid 19.
    {
      code: 'function Component() { const someRegex = /HelloWorld/; return (<Context.Provider value={someRegex}></Context.Provider>);}',
      errors: [
        {
          messageId: 'withIdentifierMsg',
          message:
            "The 'someRegex' regular expression (at line 1) passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.",
          line: 1,
          column: 42,
          endLine: 1,
          endColumn: 54,
        },
      ],
    },
    // Upstream invalid 20.
    {
      code: '\n        function Component() {\n          let foo = null;\n          let bar = x = () => {};\n          return (<Context.Provider value={bar}></Context.Provider>);\n        }\n      ',
      errors: [
        {
          messageId: 'withIdentifierMsg',
          message:
            "The 'bar' assignment expression (at line 4) passed as the value prop to the Context provider (at line 5) changes every render. To fix this consider wrapping it in a useMemo hook.",
          line: 4,
          column: 25,
          endLine: 4,
          endColumn: 33,
        },
      ],
    },
    // Upstream invalid 21.
    {
      code: "\n        import React from 'react';\n\n        const Context = React.createContext();\n        function Component() {\n          function foo() {};\n          return (<Context value={foo}></Context>)\n        }\n      ",
      errors: [
        {
          messageId: 'withIdentifierMsgFunc',
          message:
            "The 'foo' function declaration (at line 6) passed as the value prop to the Context provider (at line 7) changes every render. To fix this consider wrapping it in a useCallback hook.",
          line: 6,
          column: 11,
          endLine: 6,
          endColumn: 28,
        },
      ],
    },
    // Upstream invalid 22.
    {
      code: '\n        const MyContext = createContext();\n        function Component() { const foo = {}; return (<MyContext value={foo}></MyContext>) }\n      ',
      errors: [
        {
          messageId: 'withIdentifierMsg',
          message:
            "The 'foo' object (at line 3) passed as the value prop to the Context provider (at line 3) changes every render. To fix this consider wrapping it in a useMemo hook.",
          line: 3,
          column: 44,
          endLine: 3,
          endColumn: 46,
        },
      ],
    },
    // Upstream invalid 23.
    {
      code: '\n        const MyContext = createContext();\n        function Component() { return (<MyContext value={{foo: "bar"}}></MyContext>); }\n      ',
      errors: [
        {
          messageId: 'defaultMsg',
          message:
            'The object passed as the value prop to the Context provider (at line 3) changes every render. To fix this consider wrapping it in a useMemo hook.',
          line: 3,
          column: 58,
          endLine: 3,
          endColumn: 70,
        },
      ],
    },
    // Documentation: inline object (return fragment wrapped in a component).
    {
      code: "function Component() {\n  return (\n    <SomeContext.Provider value={{foo: 'bar'}}>\n      ...\n    </SomeContext.Provider>\n  );\n}",
      errors: [
        {
          messageId: 'defaultMsg',
          message:
            'The object passed as the value prop to the Context provider (at line 3) changes every render. To fix this consider wrapping it in a useMemo hook.',
          line: 3,
          column: 34,
          endLine: 3,
          endColumn: 46,
        },
      ],
    },
    // Documentation: React 19 context.
    {
      code: "import React from 'react';\nconst MyContext = React.createContext();\nfunction Component() {\n  function foo() {}\n  return (<MyContext value={foo}></MyContext>);\n}",
      errors: [
        {
          messageId: 'withIdentifierMsgFunc',
          message:
            "The 'foo' function declaration (at line 4) passed as the value prop to the Context provider (at line 5) changes every render. To fix this consider wrapping it in a useCallback hook.",
          line: 4,
          column: 3,
          endLine: 4,
          endColumn: 20,
        },
      ],
    },
  ],
});

package jsx_no_constructed_context_values

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// All 18 valid and 23 invalid cases from eslint-plugin-react v7.37.5,
// followed by all four documentation examples. Parser variants share cases.
// https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/tests/lib/rules/jsx-no-constructed-context-values.js
// https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/docs/rules/jsx-no-constructed-context-values.md
func TestJsxNoConstructedContextValuesUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxNoConstructedContextValuesRule, []rule_tester.ValidTestCase{
		// Upstream valid 1.
		{
			Code: `const Component = () => <Context.Provider value={props}></Context.Provider>`,
			Tsx:  true,
		},
		// Upstream valid 2.
		{
			Code: `const Component = () => <Context.Provider value={100}></Context.Provider>`,
			Tsx:  true,
		},
		// Upstream valid 3.
		{
			Code: `const Component = () => <Context.Provider value="Some string"></Context.Provider>`,
			Tsx:  true,
		},
		// Upstream valid 4.
		{
			Code:    `function Component() { const foo = useMemo(() => { return {} }, []); return (<Context.Provider value={foo}></Context.Provider>)}`,
			Tsx:     true,
			Options: []any{map[string]any{"allowArrowFunctions": true}},
		},
		// Upstream valid 5.
		{
			Code: `
        function Component({oneProp, twoProp, redProp, blueProp,}) {
          return (
            <NewContext.Provider value={twoProp}></NewContext.Provider>
          );
        }
      `,
			Tsx: true,
		},
		// Upstream valid 6.
		{
			Code: `
        function Foo(section) {
          const foo = section.section_components?.edges;

          return (
            <Context.Provider value={foo}></Context.Provider>
          )
        }
      `,
			Tsx: true,
		},
		// Upstream valid 7.
		{
			Code: `
        import foo from 'foo';
        function innerContext() {
          return (
            <Context.Provider value={foo.something}></Context.Provider>
          )
        }
      `,
			Tsx: true,
		},
		// Upstream valid 8.
		{
			Code: `
        // Passes because the lint rule doesn't handle JSX spread attributes
        function innerContext() {
          const foo = {value: 'something'}
          return (
            <Context.Provider {...foo}></Context.Provider>
          )
        }
      `,
			Tsx: true,
		},
		// Upstream valid 9.
		{
			Code: `
        // Passes because the lint rule doesn't handle JSX spread attributes
        function innerContext() {
          const foo = useMemo(() => {
            return bar;
          })
          return (
            <Context.Provider value={foo}></Context.Provider>
          )
        }
      `,
			Tsx: true,
		},
		// Upstream valid 10.
		{
			Code: `
        // Passes because we can't statically check if it's using the default value
        function Component({ a = {} }) {
          return (<Context.Provider value={a}></Context.Provider>);
        }
      `,
			Tsx: true,
		},
		// Upstream valid 11.
		{
			Code: `
          import React from 'react';
          import MyContext from './MyContext';

          const value = '';

          function ContextProvider(props) {
              return (
                  <MyContext.Provider value={value as any}>
                      {props.children}
                  </MyContext.Provider>
              )
          }
        `,
			Tsx: true,
		},
		// Upstream valid 12.
		{
			Code: `
        import React from 'react';
        import BooleanContext from './BooleanContext';

        function ContextProvider(props) {
            return (
                <BooleanContext.Provider value>
                    {props.children}
                </BooleanContext.Provider>
            )
        }
      `,
			Tsx: true,
		},
		// Upstream valid 13.
		{
			Code: `
        const root = ReactDOM.createRoot(document.getElementById('root'));
        root.render(
          <AppContext.Provider value={{}}>
            <AppView />
          </AppContext.Provider>
        );
      `,
			Tsx: true,
		},
		// Upstream valid 14.
		{
			Code: `
        // Passes because the context is not a provider
        function Component() {
          return <MyContext.Consumer value={{ foo: 'bar' }} />;
        }
      `,
			Tsx: true,
		},
		// Upstream valid 15.
		{
			Code: `
        import React from 'react';

        const MyContext = React.createContext();
        const Component = () => <MyContext value={props}></MyContext>;
      `,
			Tsx: true,
		},
		// Upstream valid 16.
		{
			Code: `
        import React from 'react';

        const MyContext = React.createContext();
        const Component = () => <MyContext value={100}></MyContext>;
      `,
			Tsx: true,
		},
		// Upstream valid 17.
		{
			Code: `
        const SomeContext = createContext();
        const Component = () => <SomeContext value="Some string"></SomeContext>;
      `,
			Tsx: true,
		},
		// Upstream valid 18.
		{
			Code: `
        // Passes because MyContext is not a variable declarator
        function Component({ MyContext }) {
          return <MyContext value={{ foo: "bar" }} />;
        }
      `,
			Tsx: true,
		},
		// Documentation: memoized value (return fragment wrapped in a component).
		{
			Code: `function Component() {
  const foo = useMemo(() => ({foo: 'bar'}), []);
  return (
    <SomeContext.Provider value={foo}>
      ...
    </SomeContext.Provider>
  );
}`,
			Tsx: true,
		},
		// Documentation: string value (closing-tag typo corrected).
		{
			Code: `const SomeContext = createContext();
const Component = () => <SomeContext value="Some string"></SomeContext>;`,
			Tsx: true,
		},
	}, []rule_tester.InvalidTestCase{
		// Upstream invalid 1.
		{
			Code: `function Component() { const foo = {}; return (<Context.Provider value={foo}></Context.Provider>) }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "withIdentifierMsg", Message: "The 'foo' object (at line 1) passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 1, Column: 36, EndLine: 1, EndColumn: 38},
			},
		},
		// Upstream invalid 2.
		{
			Code: `function Component() { const foo = []; return (<Context.Provider value={foo}></Context.Provider>) }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "withIdentifierMsg", Message: "The 'foo' array (at line 1) passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 1, Column: 36, EndLine: 1, EndColumn: 38},
			},
		},
		// Upstream invalid 3.
		{
			Code: `function Component() { const foo = () => {}; return (<Context.Provider value={foo}></Context.Provider>)}`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "withIdentifierMsgFunc", Message: "The 'foo' function expression (at line 1) passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useCallback hook.", Line: 1, Column: 36, EndLine: 1, EndColumn: 44},
			},
		},
		// Upstream invalid 4.
		{
			Code: `function Component() { const foo = function bar(){}; return (<Context.Provider value={foo}></Context.Provider>)}`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "withIdentifierMsgFunc", Message: "The 'foo' function expression (at line 1) passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useCallback hook.", Line: 1, Column: 36, EndLine: 1, EndColumn: 52},
			},
		},
		// Upstream invalid 5.
		{
			Code: `function Component() { const foo = class SomeClass{}; return (<Context.Provider value={foo}></Context.Provider>)}`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "withIdentifierMsg", Message: "The 'foo' class expression (at line 1) passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 1, Column: 36, EndLine: 1, EndColumn: 53},
			},
		},
		// Upstream invalid 6.
		{
			Code: `function Component() { const foo = new SomeClass(); return (<Context.Provider value={foo}></Context.Provider>)}`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "withIdentifierMsg", Message: "The 'foo' new expression (at line 1) passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 1, Column: 36, EndLine: 1, EndColumn: 51},
			},
		},
		// Upstream invalid 7.
		{
			Code: `function Component() { function foo() {}; return (<Context.Provider value={foo}></Context.Provider>)}`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "withIdentifierMsgFunc", Message: "The 'foo' function declaration (at line 1) passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useCallback hook.", Line: 1, Column: 24, EndLine: 1, EndColumn: 41},
			},
		},
		// Upstream invalid 8.
		{
			Code: `function Component() { const foo = true ? {} : "fine"; return (<Context.Provider value={foo}></Context.Provider>)}`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "withIdentifierMsg", Message: "The 'foo' object (at line 1) passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 1, Column: 43, EndLine: 1, EndColumn: 45},
			},
		},
		// Upstream invalid 9.
		{
			Code: `function Component() { const foo = bar || {}; return (<Context.Provider value={foo}></Context.Provider>)}`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "withIdentifierMsg", Message: "The 'foo' object (at line 1) passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 1, Column: 43, EndLine: 1, EndColumn: 45},
			},
		},
		// Upstream invalid 10.
		{
			Code: `function Component() { const foo = bar && {}; return (<Context.Provider value={foo}></Context.Provider>)}`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "withIdentifierMsg", Message: "The 'foo' object (at line 1) passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 1, Column: 43, EndLine: 1, EndColumn: 45},
			},
		},
		// Upstream invalid 11.
		{
			Code: `function Component() { const foo = bar ? baz ? {} : null : null; return (<Context.Provider value={foo}></Context.Provider>)}`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "withIdentifierMsg", Message: "The 'foo' object (at line 1) passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 1, Column: 48, EndLine: 1, EndColumn: 50},
			},
		},
		// Upstream invalid 12.
		{
			Code: `function Component() { let foo = {}; return (<Context.Provider value={foo}></Context.Provider>) }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "withIdentifierMsg", Message: "The 'foo' object (at line 1) passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 1, Column: 34, EndLine: 1, EndColumn: 36},
			},
		},
		// Upstream invalid 13.
		{
			Code: `function Component() { var foo = {}; return (<Context.Provider value={foo}></Context.Provider>)}`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "withIdentifierMsg", Message: "The 'foo' object (at line 1) passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 1, Column: 34, EndLine: 1, EndColumn: 36},
			},
		},
		// Upstream invalid 14.
		{
			Code: `
        function Component() {
          let a = {};
          a = 10;
          return (<Context.Provider value={a}></Context.Provider>);
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "withIdentifierMsg", Message: "The 'a' object (at line 3) passed as the value prop to the Context provider (at line 5) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 3, Column: 19, EndLine: 3, EndColumn: 21},
			},
		},
		// Upstream invalid 15.
		{
			Code: `
        function Component() {
          const foo = {};
          const bar = foo;
          return (<Context.Provider value={bar}></Context.Provider>);
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "withIdentifierMsg", Message: "The 'bar' object (at line 3) passed as the value prop to the Context provider (at line 5) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 3, Column: 23, EndLine: 3, EndColumn: 25},
			},
		},
		// Upstream invalid 16.
		{
			Code: `
        function Component(foo) {
          let bar = true ? foo : {};
          return (<Context.Provider value={bar}></Context.Provider>);
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "withIdentifierMsg", Message: "The 'bar' object (at line 3) passed as the value prop to the Context provider (at line 4) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 3, Column: 34, EndLine: 3, EndColumn: 36},
			},
		},
		// Upstream invalid 17.
		{
			Code: `function Component() { return (<Context.Provider value={{foo: "bar"}}></Context.Provider>);}`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "defaultMsg", Message: "The object passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 1, Column: 57, EndLine: 1, EndColumn: 69},
			},
		},
		// Upstream invalid 18.
		{
			Code: `function Component() { const Wrapper = (<SomeComp />); return (<Context.Provider value={Wrapper}></Context.Provider>);}`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "withIdentifierMsg", Message: "The 'Wrapper' JSX element (at line 1) passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 1, Column: 41, EndLine: 1, EndColumn: 53},
			},
		},
		// Upstream invalid 19.
		{
			Code: `function Component() { const someRegex = /HelloWorld/; return (<Context.Provider value={someRegex}></Context.Provider>);}`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "withIdentifierMsg", Message: "The 'someRegex' regular expression (at line 1) passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 1, Column: 42, EndLine: 1, EndColumn: 54},
			},
		},
		// Upstream invalid 20.
		{
			Code: `
        function Component() {
          let foo = null;
          let bar = x = () => {};
          return (<Context.Provider value={bar}></Context.Provider>);
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "withIdentifierMsg", Message: "The 'bar' assignment expression (at line 4) passed as the value prop to the Context provider (at line 5) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 4, Column: 25, EndLine: 4, EndColumn: 33},
			},
		},
		// Upstream invalid 21.
		{
			Code: `
        import React from 'react';

        const Context = React.createContext();
        function Component() {
          function foo() {};
          return (<Context value={foo}></Context>)
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "withIdentifierMsgFunc", Message: "The 'foo' function declaration (at line 6) passed as the value prop to the Context provider (at line 7) changes every render. To fix this consider wrapping it in a useCallback hook.", Line: 6, Column: 11, EndLine: 6, EndColumn: 28},
			},
		},
		// Upstream invalid 22.
		{
			Code: `
        const MyContext = createContext();
        function Component() { const foo = {}; return (<MyContext value={foo}></MyContext>) }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "withIdentifierMsg", Message: "The 'foo' object (at line 3) passed as the value prop to the Context provider (at line 3) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 3, Column: 44, EndLine: 3, EndColumn: 46},
			},
		},
		// Upstream invalid 23.
		{
			Code: `
        const MyContext = createContext();
        function Component() { return (<MyContext value={{foo: "bar"}}></MyContext>); }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "defaultMsg", Message: "The object passed as the value prop to the Context provider (at line 3) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 3, Column: 58, EndLine: 3, EndColumn: 70},
			},
		},
		// Documentation: inline object (return fragment wrapped in a component).
		{
			Code: `function Component() {
  return (
    <SomeContext.Provider value={{foo: 'bar'}}>
      ...
    </SomeContext.Provider>
  );
}`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "defaultMsg", Message: "The object passed as the value prop to the Context provider (at line 3) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 3, Column: 34, EndLine: 3, EndColumn: 46},
			},
		},
		// Documentation: React 19 context.
		{
			Code: `import React from 'react';
const MyContext = React.createContext();
function Component() {
  function foo() {}
  return (<MyContext value={foo}></MyContext>);
}`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "withIdentifierMsgFunc", Message: "The 'foo' function declaration (at line 4) passed as the value prop to the Context provider (at line 5) changes every render. To fix this consider wrapping it in a useCallback hook.", Line: 4, Column: 3, EndLine: 4, EndColumn: 20},
			},
		},
	})
}

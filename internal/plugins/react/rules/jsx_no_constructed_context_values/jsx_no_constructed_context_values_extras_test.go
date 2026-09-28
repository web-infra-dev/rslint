package jsx_no_constructed_context_values

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Upstream recurses indefinitely on cyclic aliases. Stop that path while
// still checking other branches of the same expression.
func TestJsxNoConstructedContextValuesCycles(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxNoConstructedContextValuesRule, []rule_tester.ValidTestCase{
		{Code: `function Component() { let value = value; return <Context.Provider value={value} />; }`, Tsx: true},
		{Code: `function Component() { let a = b; let b = a; return <Context.Provider value={a} />; }`, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `function Component() { let value = value || {}; return <Context.Provider value={value} />; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "withIdentifierMsg",
				Message:   "The 'value' object (at line 1) passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.",
				Line:      1, Column: 45, EndLine: 1, EndColumn: 47,
			}},
		},
	})
}

// Expected diagnostics were compared with eslint-plugin-react v7.37.5 and
// @typescript-eslint/parser v8.65.0, including complete ranges and messages.
func TestJsxNoConstructedContextValuesExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxNoConstructedContextValuesRule, []rule_tester.ValidTestCase{
		// Loop block excludes the initializer scope.
		{Code: `function Component() { for (let value = [];;) { return <Context.Provider value={value} />; } }`, Tsx: true},

		// Optional chain boundary in member receiver.
		{Code: `function Component() { const value = {}; return <Context.Provider value={(value?.outer).inner} />; }`, Tsx: true},

		// TypeScript instantiation expressions stay opaque.
		{Code: `function Component() { function value<T>() {} return <Context.Provider value={value<string>} />; }`, Tsx: true},

		// A later declaration shadows a context factory.
		{Code: `const Context = createContext(); function Component() { return <Context value={{}} />; const Context = null; }`, Tsx: true},

		// Initializer suppression follows the reporting node.
		{Code: `function Component() {
 // eslint-disable-next-line
 const value = {};
 return <Context.Provider value={value} />;
}`, Tsx: true,
		},
		// Outer bindings are not searched for values.
		{
			Code: `const value = {}; function Component() {  return <Context.Provider value={value} />; }`,
			Tsx:  true,
		},
		// A nested block has its own invocation scope.
		{
			Code: `function Component() { const value = {}; { return <Context.Provider value={value} />; } }`,
			Tsx:  true,
		},
		// Var bindings are hoisted out of the nested block.
		{
			Code: `function Component() { { var value = {}; return <Context.Provider value={value} />; } }`,
			Tsx:  true,
		},
		// A nested closure cannot see an outer construction.
		{
			Code: `function Component() { const value = {}; return <div>{items.map(() => <Context.Provider value={value} />)}</div>; }`,
			Tsx:  true,
		},
		// Last variable declaration wins.
		{
			Code: `function Component() { var value = {}; var value = 1; return <Context.Provider value={value} />; }`,
			Tsx:  true,
		},
		// Last declaration may have no initializer.
		{
			Code: `function Component() { var value = {}; var value; return <Context.Provider value={value} />; }`,
			Tsx:  true,
		},
		// Writes after an uninitialized declaration are ignored.
		{
			Code: `function Component() { let value; value = {}; return <Context.Provider value={value} />; }`,
			Tsx:  true,
		},
		// Destructuring defaults are not initializers.
		{
			Code: `function Component() { const {value = {}} = props; return <Context.Provider value={value} />; }`,
			Tsx:  true,
		},
		// Function-expression self name has no initializer.
		{
			Code: `const Component = function value() { return <Context.Provider value={value} />; };`,
			Tsx:  true,
		},
		// Catch parameters are not constructions.
		{
			Code: `function Component() { try {} catch (value) { return <Context.Provider value={value} />; } }`,
			Tsx:  true,
		},
		// Class declarations are not expressions.
		{
			Code: `function Component() { class Value {} return <Context.Provider value={Value} />; }`,
			Tsx:  true,
		},
		// TypeScript declarations without bodies.
		{
			Code: `function Component() { declare function value(): void; return <Context.Provider value={value} />; }`,
			Tsx:  true,
		},
		// Imported React 19 context cannot be inferred.
		{
			Code: `import Context from './context'; const Component = () => <Context value={{}} />;`,
			Tsx:  true,
		},
		// Unknown React 19 context.
		{
			Code: `const Component = () => <Context value={{}} />;`,
			Tsx:  true,
		},
		// React 19 context shadowing.
		{
			Code: `const Context = createContext(); const Component = (Context) => <Context value={{}} />;`,
			Tsx:  true,
		},
		// React 19 context ignores a later createContext definition.
		{
			Code: `var Context; var Context = createContext(); const Component = () => <Context value={{}} />;`,
			Tsx:  true,
		},
		// React 19 context with string property is not inferred.
		{
			Code: `const Context = React['createContext'](); const Component = () => <Context value={{}} />;`,
			Tsx:  true,
		},
		// Optional createContext calls are not inferred.
		{
			Code: `const Context = React?.createContext(); const Component = () => <Context value={{}} />;`,
			Tsx:  true,
		},
		// Parenthesized optional callee is not inferred.
		{
			Code: `const Context = (React?.createContext)(); const Component = () => <Context value={{}} />;`,
			Tsx:  true,
		},
		// React 19 context initializer assertions remain opaque.
		{
			Code: `const Context = createContext() as any; const Component = () => <Context value={{}} />;`,
			Tsx:  true,
		},
		// Non-React factory is not inferred.
		{
			Code: `const Context = Other.createContext(); const Component = () => <Context value={{}} />;`,
			Tsx:  true,
		},
		// Namespaced JSX tags and props are not provider values.
		{
			Code: `const Component = () => <><ns:Provider value={{}} /><Context.Provider ns:value={{}} /></>;`,
			Tsx:  true,
		},
		// Only the first value attribute is inspected.
		{
			Code: `const Component = () => <Context.Provider value="stable" value={{}} />;`,
			Tsx:  true,
		},
		// Empty JSX expression is ignored.
		{
			Code: `const Component = () => <Context.Provider value={/* comment */} />;`,
			Tsx:  true,
		},
		// Direct JSX attribute value is not an expression container.
		{
			Code: `const Component = () => <Context.Provider value=<Child /> />;`,
			Tsx:  true,
		},
		// Unrecognized function is not a component.
		{
			Code: `function renderContext() { return <Context.Provider value={{}} />; }`,
			Tsx:  true,
		},
		// Construction in a non-component class is ignored.
		{
			Code: `class Thing { render() { return <Context.Provider value={{}} />; } }`,
			Tsx:  true,
		},
		// TypeScript satisfies expressions remain opaque.
		{
			Code: `function Component() {  return <Context.Provider value={({} satisfies object)} />; }`,
			Tsx:  true,
		},
		// TypeScript non-null assertions remain opaque.
		{
			Code: `function Component() { const value = {}; return <Context.Provider value={value!} />; }`,
			Tsx:  true,
		},
		// Optional member access remains opaque.
		{
			Code: `function Component() { const value = {}; return <Context.Provider value={value?.property} />; }`,
			Tsx:  true,
		},
		// Conditional only checks result branches.
		{
			Code: `function Component() {  return <Context.Provider value={{} ? 1 : 2} />; }`,
			Tsx:  true,
		},
		// Assignments without constructed right sides are ignored.
		{
			Code: `function Component() { let value; return <Context.Provider value={value = props} />; }`,
			Tsx:  true,
		},
		// Sequence expressions are not inspected.
		{
			Code: `function Component() {  return <Context.Provider value={(0, {})} />; }`,
			Tsx:  true,
		},
		// Ordinary binary expressions are not inspected.
		{
			Code: `function Component() {  return <Context.Provider value={[] + []} />; }`,
			Tsx:  true,
		},
		// Template literals are primitive.
		{
			Code: "function Component() {  return <Context.Provider value={`value`} />; }",
			Tsx:  true,
		},
	}, []rule_tester.InvalidTestCase{
		// Direct JSX class field.
		{Code: `class Component extends React.Component { content = <Context.Provider value={{}} />; render() { return this.content; } }`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "defaultMsg", Message: "The object passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 1, Column: 78, EndLine: 1, EndColumn: 80},
			},
		},

		// Static class initialization.
		{Code: `class Component extends React.Component { static content = <Context.Provider value={{}} />; render() { return null; } }`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "defaultMsg", Message: "The object passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 1, Column: 85, EndLine: 1, EndColumn: 87},
			},
		},

		// Provider in class getter.
		{Code: `class Component extends React.Component { get content() { return <Context.Provider value={{}} />; } render() { return this.content; } }`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "defaultMsg", Message: "The object passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 1, Column: 91, EndLine: 1, EndColumn: 93},
			},
		},

		// Class static block scope.
		{Code: `class Component extends React.Component { static { const value = {}; const content = <Context.Provider value={value} />; } render() { return null; } }`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "withIdentifierMsg", Message: "The 'value' object (at line 1) passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 1, Column: 66, EndLine: 1, EndColumn: 68},
			},
		},

		// Loop variable in an unbraced body.
		{Code: `function Component() { for (let value = [];;) return <Context.Provider value={value} />; }`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "withIdentifierMsg", Message: "The 'value' array (at line 1) passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 1, Column: 41, EndLine: 1, EndColumn: 43},
			},
		},

		// Array destructuring uses the declarator.
		{Code: `function Component() { const [value = null] = [1]; return <Context.Provider value={value} />; }`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "withIdentifierMsg", Message: "The 'value' array (at line 1) passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 1, Column: 47, EndLine: 1, EndColumn: 50},
			},
		},

		// Anonymous default export component.
		{Code: `export default function () { return <Context.Provider value={{}} />; }`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "defaultMsg", Message: "The object passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 1, Column: 62, EndLine: 1, EndColumn: 64},
			},
		},

		// Async generator returning JSX.
		{Code: `async function* Component() { return <Context.Provider value={{}} />; }`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "defaultMsg", Message: "The object passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 1, Column: 63, EndLine: 1, EndColumn: 65},
			},
		},

		// Escaped identifier uses its decoded name.
		{Code: `function Component() { const v\u0061lue = {}; return <Context.Provider value={v\u0061lue} />; }`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "withIdentifierMsg", Message: "The 'value' object (at line 1) passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 1, Column: 43, EndLine: 1, EndColumn: 45},
			},
		},

		// Typed function declarations.
		{Code: `function Component() { function value(input: string): string; function value(input: any) { return input; } return <Context.Provider value={value} />; }`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "withIdentifierMsgFunc", Message: "The 'value' function declaration (at line 1) passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useCallback hook.", Line: 1, Column: 63, EndLine: 1, EndColumn: 107},
			},
		},

		// CRLF lines and leading comments.
		{Code: "function Component() {\r\n  const value = /* object */ ({\r\n    text: \"😀\"\r\n  });\r\n  return <Context.Provider value={value} />;\r\n}", Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "withIdentifierMsg", Message: "The 'value' object (at line 2) passed as the value prop to the Context provider (at line 5) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 2, Column: 31, EndLine: 4, EndColumn: 4},
			},
		},

		// Use site suppression does not hide initializer diagnostic.
		{Code: `function Component() {
 const value = {};
 // eslint-disable-next-line
 return <Context.Provider value={value} />;
}`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "withIdentifierMsg", Message: "The 'value' object (at line 2) passed as the value prop to the Context provider (at line 4) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 2, Column: 16, EndLine: 2, EndColumn: 18},
			},
		},
		// JavaScript JSDoc casts are transparent to ESTree in both lookups.
		{
			FileName: "context.jsx",
			TSConfig: "tsconfig.allow-js.json",
			Code: `const Context = /** @type {any} */ (React.createContext());
function Component() {
  const value = /** @satisfies {object} */ ({foo: 1});
  return <Context value={value} />;
}`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "withIdentifierMsg",
				Message:   "The 'value' object (at line 3) passed as the value prop to the Context provider (at line 4) changes every render. To fix this consider wrapping it in a useMemo hook.",
				Line:      3, Column: 45, EndLine: 3, EndColumn: 53,
			}},
		},
		// Block-local construction.
		{
			Code: `function Component() { { const value = {}; return <Context.Provider value={value} />; } }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "withIdentifierMsg", Message: "The 'value' object (at line 1) passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 1, Column: 40, EndLine: 1, EndColumn: 42},
			},
		},
		// Inline construction in a nested callback.
		{
			Code: `function Component() { return <div>{items.map(() => <Context.Provider value={{}} />)}</div>; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "defaultMsg", Message: "The object passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 1, Column: 78, EndLine: 1, EndColumn: 80},
			},
		},
		// Last declaration may be a function.
		{
			Code: `function Component() { var value = 1; function value() {} return <Context.Provider value={value} />; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "withIdentifierMsgFunc", Message: "The 'value' function declaration (at line 1) passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useCallback hook.", Line: 1, Column: 39, EndLine: 1, EndColumn: 58},
			},
		},
		// Destructuring inspects the declarator initializer.
		{
			Code: `function Component() { const {value} = {}; return <Context.Provider value={value} />; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "withIdentifierMsg", Message: "The 'value' object (at line 1) passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 1, Column: 40, EndLine: 1, EndColumn: 42},
			},
		},
		// Nested destructuring keeps the complete initializer.
		{
			Code: `function Component() { const {outer: {value = []}} = {outer: {}}; return <Context.Provider value={value} />; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "withIdentifierMsg", Message: "The 'value' object (at line 1) passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 1, Column: 54, EndLine: 1, EndColumn: 65},
			},
		},
		// React 19 context uses first definition.
		{
			Code: `var Context = createContext(); var Context; const Component = () => <Context value={{}} />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "defaultMsg", Message: "The object passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 1, Column: 85, EndLine: 1, EndColumn: 87},
			},
		},
		// React 19 context with computed identifier.
		{
			Code: `const Context = React[createContext](); const Component = () => <Context value={{}} />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "defaultMsg", Message: "The object passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 1, Column: 81, EndLine: 1, EndColumn: 83},
			},
		},
		// Parentheses in a React 19 context initializer.
		{
			Code: `const Context = ((React).createContext)(); const Component = () => <Context value={{}} />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "defaultMsg", Message: "The object passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 1, Column: 84, EndLine: 1, EndColumn: 86},
			},
		},
		// Destructured React 19 context.
		{
			Code: `const {Context} = createContext(); const Component = () => <Context value={{}} />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "defaultMsg", Message: "The object passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 1, Column: 76, EndLine: 1, EndColumn: 78},
			},
		},
		// JSX member providers need no factory definition.
		{
			Code: `const Component = () => <Library.Context.Provider value={{}} />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "defaultMsg", Message: "The object passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 1, Column: 58, EndLine: 1, EndColumn: 60},
			},
		},
		// Spreads do not hide an explicit value.
		{
			Code: `const Component = () => <Context.Provider {...props} value={{}} {...other} />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "defaultMsg", Message: "The object passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 1, Column: 61, EndLine: 1, EndColumn: 63},
			},
		},
		// Class render component.
		{
			Code: `class Component extends React.Component { render() { return <Context.Provider value={{}} />; } }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "defaultMsg", Message: "The object passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 1, Column: 86, EndLine: 1, EndColumn: 88},
			},
		},
		// Class field arrow component.
		{
			Code: `class Component extends React.Component { render = () => <Context.Provider value={{}} />; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "defaultMsg", Message: "The object passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 1, Column: 83, EndLine: 1, EndColumn: 85},
			},
		},
		// ES5 component.
		{
			Code: `const Component = createReactClass({render() { return <Context.Provider value={{}} />; }});`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "defaultMsg", Message: "The object passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 1, Column: 80, EndLine: 1, EndColumn: 82},
			},
		},
		// Memo component.
		{
			Code: `const Component = React.memo(() => <Context.Provider value={{}} />);`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "defaultMsg", Message: "The object passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 1, Column: 61, EndLine: 1, EndColumn: 63},
			},
		},
		// Forward ref component.
		{
			Code: `const Component = React.forwardRef((props, ref) => <Context.Provider value={{}} />);`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "defaultMsg", Message: "The object passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 1, Column: 77, EndLine: 1, EndColumn: 79},
			},
		},
		// Configured component wrapper.
		{
			Code:     `const Component = observer(() => <Context.Provider value={{}} />);`,
			Tsx:      true,
			Settings: map[string]any{"componentWrapperFunctions": []any{"observer"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "defaultMsg", Message: "The object passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 1, Column: 59, EndLine: 1, EndColumn: 61},
			},
		},
		// Configured React pragma.
		{
			Code:     `class Component extends Preact.Component { render() { return <Context.Provider value={{}} />; } }`,
			Tsx:      true,
			Settings: map[string]any{"react": map[string]any{"pragma": "Preact"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "defaultMsg", Message: "The object passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 1, Column: 87, EndLine: 1, EndColumn: 89},
			},
		},
		// Comment React pragma.
		{
			Code: `/** @jsx Preact.h */
class Component extends Preact.Component { render() { return <Context.Provider value={{}} />; } }`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "defaultMsg", Message: "The object passed as the value prop to the Context provider (at line 2) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 2, Column: 87, EndLine: 2, EndColumn: 89},
			},
		},
		// Inline function uses useCallback.
		{
			Code: `function Component() {  return <Context.Provider value={() => {}} />; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "defaultMsgFunc", Message: "The function expression passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useCallback hook.", Line: 1, Column: 57, EndLine: 1, EndColumn: 65},
			},
		},
		// Inline function expression.
		{
			Code: `function Component() {  return <Context.Provider value={(function () {})} />; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "defaultMsgFunc", Message: "The function expression passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useCallback hook.", Line: 1, Column: 58, EndLine: 1, EndColumn: 72},
			},
		},
		// Inline fragment.
		{
			Code: `function Component() {  return <Context.Provider value={<><Child /></>} />; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "defaultMsg", Message: "The JSX fragment passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 1, Column: 57, EndLine: 1, EndColumn: 71},
			},
		},
		// Inline JSX element.
		{
			Code: `function Component() {  return <Context.Provider value={<Child></Child>} />; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "defaultMsg", Message: "The JSX element passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 1, Column: 57, EndLine: 1, EndColumn: 72},
			},
		},
		// TypeScript as expressions and parentheses are transparent.
		{
			Code: `function Component() {  return <Context.Provider value={(({} as const) as object)} />; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "defaultMsg", Message: "The object passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 1, Column: 59, EndLine: 1, EndColumn: 61},
			},
		},
		// Computed member access follows its object.
		{
			Code: `function Component() { const value = {}; return <Context.Provider value={value[unknown]} />; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "withIdentifierMsg", Message: "The 'value' object (at line 1) passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 1, Column: 38, EndLine: 1, EndColumn: 40},
			},
		},
		// Member access preserves the immediate object usage.
		{
			Code: `function Component() { const value = {}; return <Context.Provider value={value.outer.inner} />; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "withIdentifierMsg", Message: "The 'undefined' object (at line 1) passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 1, Column: 38, EndLine: 1, EndColumn: 40},
			},
		},
		// Parenthesized member receiver.
		{
			Code: `function Component() { const value = {}; return <Context.Provider value={(value).property} />; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "withIdentifierMsg", Message: "The 'value' object (at line 1) passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 1, Column: 38, EndLine: 1, EndColumn: 40},
			},
		},
		// Inline object receiver.
		{
			Code: `function Component() {  return <Context.Provider value={({property: 1}).property} />; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "withIdentifierMsg", Message: "The 'undefined' object (at line 1) passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 1, Column: 58, EndLine: 1, EndColumn: 71},
			},
		},
		// Private member follows its receiver.
		{
			Code: `class Component extends React.Component { #value; render() { const value = new Component(); return <Context.Provider value={value.#value} />; } }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "withIdentifierMsg", Message: "The 'value' new expression (at line 1) passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 1, Column: 76, EndLine: 1, EndColumn: 91},
			},
		},
		// React 19 private context factory follows upstream property name.
		{
			Code: `class Component extends React.Component { #createContext; render() { const React = this; const Context = React.#createContext(); return <Context value={{}} />; } }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "defaultMsg", Message: "The object passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 1, Column: 153, EndLine: 1, EndColumn: 155},
			},
		},
		// Logical operators prefer the left construction.
		{
			Code: `function Component() {  return <Context.Provider value={{} || []} />; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "defaultMsg", Message: "The object passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 1, Column: 57, EndLine: 1, EndColumn: 59},
			},
		},
		// Nullish coalescing examines the right value.
		{
			Code: `function Component() {  return <Context.Provider value={props ?? []} />; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "defaultMsg", Message: "The array passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 1, Column: 66, EndLine: 1, EndColumn: 68},
			},
		},
		// Conditional falls back to alternate branch.
		{
			Code: `function Component() {  return <Context.Provider value={flag ? props : (() => {})} />; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "defaultMsgFunc", Message: "The function expression passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useCallback hook.", Line: 1, Column: 73, EndLine: 1, EndColumn: 81},
			},
		},
		// Direct assignment retains assignment message.
		{
			Code: `function Component() { let value; return <Context.Provider value={value = {}} />; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "withIdentifierMsg", Message: "The 'undefined' assignment expression (at line 1) passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 1, Column: 75, EndLine: 1, EndColumn: 77},
			},
		},
		// Compound assignment inspects its right side.
		{
			Code: `function Component() { let value; return <Context.Provider value={value ||= () => {}} />; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "withIdentifierMsg", Message: "The 'undefined' assignment expression (at line 1) passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 1, Column: 77, EndLine: 1, EndColumn: 85},
			},
		},
		// Ignored options do not suppress functions.
		{
			Code:    `function Component() {  return <Context.Provider value={() => {}} />; }`,
			Tsx:     true,
			Options: []any{map[string]any{"allowArrowFunctions": true}, false, "ignored", nil},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "defaultMsgFunc", Message: "The function expression passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useCallback hook.", Line: 1, Column: 57, EndLine: 1, EndColumn: 65},
			},
		},
		// Multiline and UTF-16 diagnostic range.
		{
			Code: `function Component() {
  const 标签 = '😀', value = (
    {标签: '😀'}
  );
  return <Context.Provider value={value} />;
}`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "withIdentifierMsg", Message: "The 'value' object (at line 3) passed as the value prop to the Context provider (at line 5) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 3, Column: 5, EndLine: 3, EndColumn: 15},
			},
		},
		// UTF-16 start column after an astral character.
		{
			Code: `function Component() { const 标签 = '😀', value = {}; return <Context.Provider value={value} />; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "withIdentifierMsg", Message: "The 'value' object (at line 1) passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 1, Column: 49, EndLine: 1, EndColumn: 51},
			},
		},
		// Two providers report the same construction twice.
		{
			Code: `function Component() { const value = {}; return <><A.Provider value={value} /><B.Provider value={value} /></>; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "withIdentifierMsg", Message: "The 'value' object (at line 1) passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 1, Column: 38, EndLine: 1, EndColumn: 40},
				{MessageId: "withIdentifierMsg", Message: "The 'value' object (at line 1) passed as the value prop to the Context provider (at line 1) changes every render. To fix this consider wrapping it in a useMemo hook.", Line: 1, Column: 38, EndLine: 1, EndColumn: 40},
			},
		},
	})
}

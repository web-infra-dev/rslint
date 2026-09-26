package no_object_type_as_default_prop

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// These expectations were checked against eslint-plugin-react v7.37.5.
func TestNoObjectTypeAsDefaultPropExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoObjectTypeAsDefaultPropRule, []rule_tester.ValidTestCase{
		// Authored this remains the first TypeScript parameter.
		{Code: `function Foo(this: unknown, {items = []}: any) { return null; }`, Tsx: true},
		// Tagged templates and comma expressions are outside the forbidden syntax set.
		{Code: "function Foo({a = tag`x`, b = (0, {}), c = (Symbol.for)('x')}) { return null; }", Tsx: true},
		// Optional typed calls remain optional chains.
		{Code: `function Foo({id = Symbol?.<unknown>('x')}) { return null; }`, Tsx: true},
		// Authored component assertions do not identify function components.
		{Code: `const Foo = (({items = []}) => null) satisfies Function;`, Tsx: true},
		// Declared async generators remain excluded when assigned to properties.
		{Code: `const views = { async *Foo({items = []}) { return <div />; } };`, Tsx: true},
		// Parenthesized default export must still return JSX.
		{Code: `export default (({items = []}) => null);`, Tsx: true},
		// Scope suppression remains effective.
		{Code: `/* eslint-disable */
function Foo({items = []}) { return null; }`, Tsx: true},
		// Line suppression remains effective.
		{Code: `// eslint-disable-next-line
const Foo = ({items = []}) => null;`, Tsx: true},
		// Assignment names take precedence over a function expression's own name.
		{Code: `let lower; lower = function Foo({items = []}) { return null; };`, Tsx: true},
		// Primitives, references, and expressions outside the forbidden syntax set.
		{Code: "function Foo({a = null, b = undefined, c = true, d = 1n, e = -1, f = `text`, g = defaults.items, h = makeItems(), i = Symbol.for('id'), j = cond ? {} : [], k = <></>, l = obj?.value}) { return <div />; }", Tsx: true},
		// Only direct defaults in the first object pattern are inspected.
		{Code: `function Foo({nested: {items = []}, ...rest}, {later = {}}) { return null; }`, Tsx: true},
		// An assignment pattern around the whole parameter is ignored upstream.
		{Code: `function Foo({items = []} = {}) { return null; }`, Tsx: true},
		// Array and rest parameters are not object patterns.
		{Code: `function Foo([items = []]) { return null; } function Bar(...[{items = []}]) { return null; }`, Tsx: true},
		// Body destructuring and class methods are outside this rule.
		{Code: `function Foo(props) { const {items = []} = props; return null; } class Bar extends React.Component { render({items = []}) { return <div />; } }`, Tsx: true},
		// Functions without component returns, lowercase functions, and nested returns.
		{Code: `function Foo({items = []}) { function Nested() { return <div />; } } const lower = ({items = []}) => null; function lowerCase({items = []}) { return <div />; }`, Tsx: true},
		// Optional Symbol calls and member calls are not direct Symbol calls.
		{Code: `function Foo({a = Symbol?.('id'), b = Symbol.for('id'), c = globalThis.Symbol('id'), d = (Symbol?.())()}) { return null; }`, Tsx: true},
		// Object methods returning only null and private class methods are not components.
		{Code: `const views = { Foo({data = {}}) { return null; } }; class View { #Foo({data = {}}) { return <div />; } }`, Tsx: true},
		// Built-in wrapper calls replace the inner function in the component list.
		{Code: `import React, {memo, forwardRef} from 'react'; const Foo = React.memo(({items = []}) => <div />); const Bar = forwardRef(({items = []}, ref) => <div />); const Baz = memo(({items = []}) => null);`, Tsx: true},
		// Configured wrappers follow the same component registration behavior.
		{Code: `const Foo = observer(({items = []}) => <div />);`, Tsx: true, Settings: map[string]any{"componentWrapperFunctions": []any{"observer"}}},
		// Unrecognized callback and IIFE positions are ignored.
		{Code: `wrap(({items = []}) => <div />); (function Foo({items = []}) { return <div />; })();`, Tsx: true},
		// Default-exported arrows require JSX; null-only arrows are ignored.
		{Code: `export default ({items = []}) => null;`, Tsx: true},
		// A parenthesized lowercase binding must retain the component naming gate.
		{Code: `const lower = (({items = []}) => null);`, Tsx: true},
		// A parenthesized object property retains its capitalization gate.
		{Code: `const views = { lower: (({items = []}) => <div />) };`, Tsx: true},
		// Authored TypeScript assertions remain visible to the upstream rule.
		{Code: `function Foo({a = {} as object, b = [] satisfies unknown[], c = (() => {})!, d = (Symbol as any)('id')}) { return null; }`, Tsx: true},
		// Body-absent declarations have no component return.
		{Code: `declare function Foo({items = []}: any): unknown;`, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// JSDoc this does not hide the first authored props parameter.
		{Code: `/** @this {unknown} */
function Foo({items = []}) { return null; }`, FileName: "review.jsx", TSConfig: "tsconfig.allow-js.json",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenTypeDefaultParam", Message: "items has a/an array literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of array literal.", Line: 2, Column: 15, EndLine: 2, EndColumn: 25},
			}},
		// JSDoc cast around the component preserves its position.
		{Code: `const Foo = /** @type {Function} */ (({items = []}) => null);`, FileName: "review.jsx", TSConfig: "tsconfig.allow-js.json",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenTypeDefaultParam", Message: "items has a/an array literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of array literal.", Line: 1, Column: 40, EndLine: 1, EndColumn: 50},
			}},
		// Escaped identifier keys use their decoded name with CRLF source.
		{Code: "function Foo({\\u0061: value = []}) {\r\n  return <div />;\r\n}", Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenTypeDefaultParam", Message: "a has a/an array literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of array literal.", Line: 1, Column: 23, EndLine: 1, EndColumn: 33},
			}},
		// Nested parenthesized bindings keep their complete assignment range.
		{Code: `function Foo({data: {x} = (/* before */ {x: 1} /* after */), list: [item] = ([1])}) { return null; }`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenTypeDefaultParam", Message: "data has a/an object literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of object literal.", Line: 1, Column: 21, EndLine: 1, EndColumn: 60},
				{MessageId: "forbiddenTypeDefaultParam", Message: "list has a/an array literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of array literal.", Line: 1, Column: 68, EndLine: 1, EndColumn: 82},
			}},
		// Trailing comments stay outside the assignment range.
		{Code: `function Foo({items = [] /* comment */, other = new Map}) { return null; }`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenTypeDefaultParam", Message: "items has a/an array literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of array literal.", Line: 1, Column: 15, EndLine: 1, EndColumn: 25},
				{MessageId: "forbiddenTypeDefaultParam", Message: "other has a/an construction expression as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of construction expression.", Line: 1, Column: 41, EndLine: 1, EndColumn: 56},
			}},
		// Call type arguments do not hide a direct Symbol call.
		{Code: `function Foo({id = Symbol<unknown>('x')}) { return null; }`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenTypeDefaultParam", Message: "id has a/an Symbol literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of Symbol literal.", Line: 1, Column: 15, EndLine: 1, EndColumn: 40},
			}},
		// Setter methods are represented as function expressions upstream.
		{Code: `const views = { set Foo({items = []}) { return <div />; } };`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenTypeDefaultParam", Message: "items has a/an array literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of array literal.", Line: 1, Column: 26, EndLine: 1, EndColumn: 36},
			}},
		// Member assignments use the assigned component name.
		{Code: `const views = {}; views.Foo = (({items = []}) => <div />); views.lower = (({items = []}) => <div />);`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenTypeDefaultParam", Message: "items has a/an array literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of array literal.", Line: 1, Column: 34, EndLine: 1, EndColumn: 44},
			}},
		// Parenthesized nested returned functions preserve their component boundary.
		{Code: `let Foo; Foo = (() => (({items = []}) => <div />));`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenTypeDefaultParam", Message: "items has a/an array literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of array literal.", Line: 1, Column: 26, EndLine: 1, EndColumn: 36},
			}},
		// Later forbidden defaults are checked after ordinary props.
		{Code: `function Foo({first, count = 1, data = [], callback = () => {}}) { return null; }`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenTypeDefaultParam", Message: "data has a/an array literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of array literal.", Line: 1, Column: 33, EndLine: 1, EndColumn: 42},
				{MessageId: "forbiddenTypeDefaultParam", Message: "callback has a/an arrow function as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of arrow function.", Line: 1, Column: 44, EndLine: 1, EndColumn: 63},
			}},
		// Parentheses preserve all forbidden initializer types.
		{Code: `function Foo({a = ({}), b = ([]), c = (/x/u), d = (() => {}), e = (function () {}), f = (class {}), g = (new Map()), h = (<div>text</div>), i = ((Symbol)('id'))}) { return null; }`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenTypeDefaultParam", Message: "a has a/an object literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of object literal.", Line: 1, Column: 15, EndLine: 1, EndColumn: 23},
				{MessageId: "forbiddenTypeDefaultParam", Message: "b has a/an array literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of array literal.", Line: 1, Column: 25, EndLine: 1, EndColumn: 33},
				{MessageId: "forbiddenTypeDefaultParam", Message: "c has a/an regex literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of regex literal.", Line: 1, Column: 35, EndLine: 1, EndColumn: 45},
				{MessageId: "forbiddenTypeDefaultParam", Message: "d has a/an arrow function as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of arrow function.", Line: 1, Column: 47, EndLine: 1, EndColumn: 61},
				{MessageId: "forbiddenTypeDefaultParam", Message: "e has a/an function expression as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of function expression.", Line: 1, Column: 63, EndLine: 1, EndColumn: 83},
				{MessageId: "forbiddenTypeDefaultParam", Message: "f has a/an class expression as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of class expression.", Line: 1, Column: 85, EndLine: 1, EndColumn: 99},
				{MessageId: "forbiddenTypeDefaultParam", Message: "g has a/an construction expression as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of construction expression.", Line: 1, Column: 101, EndLine: 1, EndColumn: 116},
				{MessageId: "forbiddenTypeDefaultParam", Message: "h has a/an JSX element as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of JSX element.", Line: 1, Column: 118, EndLine: 1, EndColumn: 139},
				{MessageId: "forbiddenTypeDefaultParam", Message: "i has a/an Symbol literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of Symbol literal.", Line: 1, Column: 141, EndLine: 1, EndColumn: 161},
			}},
		// Aliased keys report only the assignment pattern, using the original key name.
		{Code: `const Foo = ({items: values = [], callback: handler = () => {}}) => <div />;`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenTypeDefaultParam", Message: "items has a/an array literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of array literal.", Line: 1, Column: 22, EndLine: 1, EndColumn: 33},
				{MessageId: "forbiddenTypeDefaultParam", Message: "callback has a/an arrow function as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of arrow function.", Line: 1, Column: 45, EndLine: 1, EndColumn: 63},
			}},
		// Literal and computed keys follow upstream key.name, not static property names.
		{Code: `function Foo({'items': values = [], 1: one = {}, [key]: dynamic = /x/, ['literal']: fixed = new Map(), [obj.key]: member = []}) { return null; }`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenTypeDefaultParam", Message: "undefined has a/an array literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of array literal.", Line: 1, Column: 24, EndLine: 1, EndColumn: 35},
				{MessageId: "forbiddenTypeDefaultParam", Message: "undefined has a/an object literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of object literal.", Line: 1, Column: 40, EndLine: 1, EndColumn: 48},
				{MessageId: "forbiddenTypeDefaultParam", Message: "key has a/an regex literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of regex literal.", Line: 1, Column: 57, EndLine: 1, EndColumn: 70},
				{MessageId: "forbiddenTypeDefaultParam", Message: "undefined has a/an construction expression as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of construction expression.", Line: 1, Column: 85, EndLine: 1, EndColumn: 102},
				{MessageId: "forbiddenTypeDefaultParam", Message: "undefined has a/an array literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of array literal.", Line: 1, Column: 115, EndLine: 1, EndColumn: 126},
			}},
		// Nested binding patterns with their own default are reported as a whole.
		{Code: `function Foo({items: [first] = [], options: {enabled} = {}}) { return null; }`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenTypeDefaultParam", Message: "items has a/an array literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of array literal.", Line: 1, Column: 22, EndLine: 1, EndColumn: 34},
				{MessageId: "forbiddenTypeDefaultParam", Message: "options has a/an object literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of object literal.", Line: 1, Column: 45, EndLine: 1, EndColumn: 59},
			}},
		// Multiline ranges start at the renamed binding and include the default.
		{Code: `function Foo({
  items: values = (
    [1, 2]
  ),
}) { return null; }`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenTypeDefaultParam", Message: "items has a/an array literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of array literal.", Line: 2, Column: 10, EndLine: 4, EndColumn: 4},
			}},
		// UTF-16 positions include astral characters before the binding.
		{Code: `function Foo({ /* 🍎 */ 数据: 值 = {}, [键]: 项 = []}) { return null; }`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenTypeDefaultParam", Message: "数据 has a/an object literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of object literal.", Line: 1, Column: 29, EndLine: 1, EndColumn: 35},
				{MessageId: "forbiddenTypeDefaultParam", Message: "键 has a/an array literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of array literal.", Line: 1, Column: 42, EndLine: 1, EndColumn: 48},
			}},
		// Symbol is matched syntactically even when shadowed.
		{Code: `function outer(Symbol) { function Foo({id = Symbol('id')}) { return null; } }`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenTypeDefaultParam", Message: "id has a/an Symbol literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of Symbol literal.", Line: 1, Column: 40, EndLine: 1, EndColumn: 57},
			}},
		// Function expressions and shorthand object methods can be components.
		{Code: `const Foo = function ({items = []}) { return <div />; }; const views = { Bar({data = {}}) { return <span />; } };`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenTypeDefaultParam", Message: "items has a/an array literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of array literal.", Line: 1, Column: 24, EndLine: 1, EndColumn: 34},
				{MessageId: "forbiddenTypeDefaultParam", Message: "data has a/an object literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of object literal.", Line: 1, Column: 79, EndLine: 1, EndColumn: 88},
			}},
		// Async generators are excluded; ordinary async and generator functions remain eligible.
		{Code: `async function* Hidden({items = []}) { return <div />; } async function Async({items = []}) { return null; } function* Generator({items = []}) { return null; }`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenTypeDefaultParam", Message: "items has a/an array literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of array literal.", Line: 1, Column: 80, EndLine: 1, EndColumn: 90},
				{MessageId: "forbiddenTypeDefaultParam", Message: "items has a/an array literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of array literal.", Line: 1, Column: 131, EndLine: 1, EndColumn: 141},
			}},
		// Configured pragma participates in component return detection.
		{Code: `function Foo({items = []}) { return h.createElement('div'); }`, Tsx: true, Settings: map[string]any{"react": map[string]any{"pragma": "h", "version": "18.3"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenTypeDefaultParam", Message: "items has a/an array literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of array literal.", Line: 1, Column: 15, EndLine: 1, EndColumn: 25},
			}},
		// Anonymous default-exported declarations are components even when returning null.
		{Code: `export default function ({items = []}) { return null; }`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenTypeDefaultParam", Message: "items has a/an array literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of array literal.", Line: 1, Column: 27, EndLine: 1, EndColumn: 37},
			}},
		// Default-exported JSX arrows are components.
		{Code: `export default ({items = []}) => <div />;`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenTypeDefaultParam", Message: "items has a/an array literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of array literal.", Line: 1, Column: 18, EndLine: 1, EndColumn: 28},
			}},
		// Parenthesized uppercase bindings and assignments retain their component position.
		{Code: `let Foo; Foo = (({items = []}) => null); const Bar = (function ({items = []}) { return <div />; });`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenTypeDefaultParam", Message: "items has a/an array literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of array literal.", Line: 1, Column: 19, EndLine: 1, EndColumn: 29},
				{MessageId: "forbiddenTypeDefaultParam", Message: "items has a/an array literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of array literal.", Line: 1, Column: 66, EndLine: 1, EndColumn: 76},
			}},
		// JSX returned through a local or outer variable still identifies a component.
		{Code: `const view = <div />; function Foo({items = []}) { return view; } function Bar({data = {}}) { const content = <span />; return content; }`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenTypeDefaultParam", Message: "items has a/an array literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of array literal.", Line: 1, Column: 37, EndLine: 1, EndColumn: 47},
				{MessageId: "forbiddenTypeDefaultParam", Message: "data has a/an object literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of object literal.", Line: 1, Column: 81, EndLine: 1, EndColumn: 90},
			}},
		// TypeScript parameter annotations preserve defaults and their locations.
		{Code: `function Foo<T>({items = [], data: value = {}}: {items?: T[]; data?: object}) { return <div />; }`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenTypeDefaultParam", Message: "items has a/an array literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of array literal.", Line: 1, Column: 18, EndLine: 1, EndColumn: 28},
				{MessageId: "forbiddenTypeDefaultParam", Message: "data has a/an object literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of object literal.", Line: 1, Column: 36, EndLine: 1, EndColumn: 46},
			}},
		// JavaScript JSDoc casts remain transparent.
		{Code: `function Foo({items = /** @type {any[]} */ ([]), callback = /** @type {Function} */ (() => {})}) { return null; }`, FileName: "component.jsx", TSConfig: "tsconfig.allow-js.json",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenTypeDefaultParam", Message: "items has a/an array literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of array literal.", Line: 1, Column: 15, EndLine: 1, EndColumn: 48},
				{MessageId: "forbiddenTypeDefaultParam", Message: "callback has a/an arrow function as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of arrow function.", Line: 1, Column: 50, EndLine: 1, EndColumn: 95},
			}},
		// Documentation array example with a component return.
		{Code: `function Component({items = []}) { return null; }`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenTypeDefaultParam", Message: "items has a/an array literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of array literal.", Line: 1, Column: 21, EndLine: 1, EndColumn: 31},
			}},
		// Documentation object and callback examples with component returns.
		{Code: `const Component = ({items = {}}) => <div />; const Other = ({items = () => {}}) => null;`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenTypeDefaultParam", Message: "items has a/an object literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of object literal.", Line: 1, Column: 21, EndLine: 1, EndColumn: 31},
				{MessageId: "forbiddenTypeDefaultParam", Message: "items has a/an arrow function as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of arrow function.", Line: 1, Column: 62, EndLine: 1, EndColumn: 78},
			}},
	})
}

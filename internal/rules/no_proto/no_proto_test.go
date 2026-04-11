package no_proto

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoProtoRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoProtoRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			// --- Object literal contexts (not member access) ---
			{Code: `var a = { __proto__: [] }`},
			{Code: `var a = { __proto__ }`},
			{Code: `var a = { ["__proto__"]: [] }`},
			{Code: `var a = { __proto__() {} }`},
			{Code: `var a = { get __proto__() { return 1; } }`},
			{Code: `var a = { set __proto__(v) {} }`},

			// --- __proto__ as declaration name ---
			{Code: `var __proto__ = 1;`},
			{Code: `let __proto__ = 2;`},
			{Code: `const __proto__ = 3;`},
			{Code: `function __proto__() {}`},
			{Code: `function foo(__proto__) {}`},

			// --- Destructuring (binding pattern, not property access) ---
			{Code: `var { __proto__ } = obj;`},
			{Code: `var { __proto__: proto } = obj;`},
			{Code: `var { a: { __proto__ } } = obj;`},
			{Code: `function foo({ __proto__ }) {}`},
			// Array destructuring binding
			{Code: `var [__proto__] = arr;`},
			// Rest element in destructuring
			{Code: `var { a, ...__proto__ } = obj;`},
			{Code: `var [a, ...__proto__] = arr;`},

			// --- Catch clause binding ---
			{Code: `try {} catch (__proto__) {}`},

			// --- For-in / for-of declarations ---
			{Code: `for (var __proto__ in obj) {}`},
			{Code: `for (var __proto__ of arr) {}`},

			// --- Import / Export ---
			{Code: `import { __proto__ } from 'mod';`},
			{Code: `var __proto__ = 1; export { __proto__ };`},

			// --- TypeScript type-level constructs ---
			{Code: `interface I { __proto__: string }`},
			{Code: `type T = { __proto__: string }`},
			{Code: `declare class C { __proto__: string }`},

			// --- Class member declarations ---
			{Code: `class C { __proto__ = 1 }`},
			{Code: `class C { __proto__() {} }`},
			{Code: `class C { get __proto__() { return 1; } }`},
			{Code: `class C { static __proto__ = 1 }`},

			// --- Enum member ---
			{Code: `enum E { __proto__ = 1 }`},

			// --- __proto__ as type-level names ---
			{Code: `class __proto__ {}`},
			{Code: `type __proto__ = string;`},
			{Code: `namespace __proto__ {}`},
			{Code: `function foo<__proto__>() {}`},
			{Code: `declare function __proto__(): void;`},
			{Code: `abstract class C { abstract __proto__(): void }`},

			// --- Label ---
			{Code: `__proto__: for (;;) { break __proto__; }`},

			// --- String / non-member-access usage ---
			{Code: `var s = "__proto__";`},
			{Code: "var s = `__proto__`;"},

			// --- Different property name ---
			{Code: `obj.prototype`},
			{Code: `obj.__proto`},
			{Code: `obj.proto__`},
			{Code: `obj.__PROTO__`},

			// --- Recommended alternatives ---
			{Code: `var a = Object.getPrototypeOf(obj);`},
			{Code: `Object.setPrototypeOf(obj, b);`},

			// --- Dynamic / non-static property access ---
			{Code: `var x = "__proto__"; obj[x];`},
			{Code: "obj[`__${'proto'}__`]"},
			{Code: `var k = "__proto__"; obj[k] = 1;`},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			// === Dot notation ===
			{
				Code: `var a = obj.__proto__;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 9},
				},
			},
			{
				Code: `obj.__proto__ = b;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 1},
				},
			},

			// === Bracket notation ===
			{
				Code: `var a = obj["__proto__"];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 9},
				},
			},
			{
				Code: `obj["__proto__"] = b;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 1},
				},
			},
			{
				Code: `var a = obj['__proto__'];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 9},
				},
			},
			// Template literal (no substitution)
			{
				Code: "var a = obj[`__proto__`];",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 9},
				},
			},

			// === Optional chaining ===
			{
				Code: `obj?.__proto__;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 1},
				},
			},
			{
				Code: `obj?.["__proto__"];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 1},
				},
			},

			// === Chained / nested access ===
			{
				Code: `var a = foo.bar.__proto__;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 9},
				},
			},
			{
				Code: `obj.__proto__.hasOwnProperty("foo");`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 1},
				},
			},
			// Double __proto__ — two errors
			{
				Code: `obj.__proto__.__proto__;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 1},
					{MessageId: "unexpectedProto", Line: 1, Column: 1},
				},
			},
			// Deeply chained
			{
				Code: `a.b.c.d.__proto__;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 1},
				},
			},
			// Optional chain continuing after __proto__
			{
				Code: `obj?.__proto__?.toString();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 1},
				},
			},

			// === Calling __proto__ as function ===
			{
				Code: `obj.__proto__();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 1},
				},
			},
			{
				Code: `obj["__proto__"]();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 1},
				},
			},

			// === this ===
			{
				Code: `var a = this.__proto__;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 9},
				},
			},
			{
				Code: `this["__proto__"];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 1},
				},
			},

			// === Parenthesized expression ===
			{
				Code: `(obj).__proto__;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 1},
				},
			},
			{
				Code: `(obj)["__proto__"];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 1},
				},
			},

			// === TypeScript expressions ===
			// as assertion
			{
				Code: `(obj as any).__proto__;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 1},
				},
			},
			// Angle-bracket assertion
			{
				Code: `(<any>obj).__proto__;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 1},
				},
			},
			// Non-null assertion
			{
				Code: `obj!.__proto__;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 1},
				},
			},
			// Satisfies expression
			{
				Code: `(obj satisfies any).__proto__;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 1},
				},
			},

			// === As function argument ===
			{
				Code: `foo(obj.__proto__);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 5},
				},
			},
			{
				Code: `console.log(obj["__proto__"]);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 13},
				},
			},

			// === new / await / yield ===
			{
				Code: `new (obj.__proto__)();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 6},
				},
			},
			{
				Code: `async function f() { await obj.__proto__; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 28},
				},
			},
			{
				Code: `function* g() { yield obj.__proto__; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 23},
				},
			},

			// === Tagged template ===
			{
				Code: "obj.__proto__`tagged`;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 1},
				},
			},

			// === Update expressions (++/--) ===
			{
				Code: `++obj.__proto__;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 3},
				},
			},
			{
				Code: `obj.__proto__++;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 1},
				},
			},

			// === In expressions ===
			// Template literal expression
			{
				Code: "var s = `${obj.__proto__}`;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 12},
				},
			},
			// Ternary
			{
				Code: `var a = x ? obj.__proto__ : null;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 13},
				},
			},
			// Logical OR / AND
			{
				Code: `var a = obj.__proto__ || default_;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 9},
				},
			},
			{
				Code: `var a = x && obj.__proto__;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 14},
				},
			},
			// Nullish coalescing
			{
				Code: `var a = obj.__proto__ ?? default_;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 9},
				},
			},
			// Comma operator
			{
				Code: `(0, obj.__proto__);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 5},
				},
			},
			// in expression
			{
				Code: `"x" in obj.__proto__;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 8},
				},
			},
			// instanceof
			{
				Code: `obj.__proto__ instanceof Object;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 1},
				},
			},

			// === Unary operators ===
			{
				Code: `typeof obj.__proto__;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 8},
				},
			},
			{
				Code: `void obj.__proto__;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 6},
				},
			},
			{
				Code: `delete obj.__proto__;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 8},
				},
			},

			// === Spread ===
			{
				Code: `var a = { ...obj.__proto__ };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 14},
				},
			},
			{
				Code: `var a = [...obj.__proto__];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 13},
				},
			},

			// === Assignment patterns ===
			// Compound assignment
			{
				Code: `obj.__proto__ += "";`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 1},
				},
			},
			// Logical assignment operators
			{
				Code: `obj.__proto__ ||= x;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 1},
				},
			},
			{
				Code: `obj.__proto__ &&= x;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 1},
				},
			},
			{
				Code: `obj.__proto__ ??= x;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 1},
				},
			},
			// Destructuring default value
			{
				Code: `var { a = obj.__proto__ } = b;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 11},
				},
			},
			// Array destructuring assignment target
			{
				Code: `[obj.__proto__] = [1];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 2},
				},
			},

			// === As value in object / array literal ===
			// Object property value
			{
				Code: `var a = { key: obj.__proto__ };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 16},
				},
			},
			// Array element
			{
				Code: `var a = [obj.__proto__];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 10},
				},
			},
			// Computed property key
			{
				Code: `var a = { [obj.__proto__]: 1 };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 12},
				},
			},

			// === for-in / for-of with member expression as target ===
			{
				Code: `for (obj.__proto__ in x) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 6},
				},
			},
			{
				Code: `for (obj.__proto__ of x) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 6},
				},
			},

			// === Function / class contexts ===
			{
				Code: `function f() { return obj.__proto__; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 23},
				},
			},
			{
				Code: `var f = () => obj.__proto__;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 15},
				},
			},
			{
				Code: `async function f() { return obj.__proto__; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 29},
				},
			},
			{
				Code: `class C { method() { return this.__proto__; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 29},
				},
			},
			{
				Code: `class C { constructor() { this.__proto__ = null; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 27},
				},
			},
			{
				Code: `class C { get p() { return obj.__proto__; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 28},
				},
			},
			{
				Code: `class C { static { obj.__proto__; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 20},
				},
			},
			// Class field initializer value
			{
				Code: `class C { x = obj.__proto__ }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 15},
				},
			},
			// IIFE
			{
				Code: `(function() { return obj.__proto__; })();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 22},
				},
			},
			// Arrow returning object
			{
				Code: `var f = () => ({ a: obj.__proto__ });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 21},
				},
			},

			// === Namespace / module scope ===
			{
				Code: `namespace N { var a = obj.__proto__; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 23},
				},
			},

			// === Export ===
			{
				Code: `export default obj.__proto__;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 16},
				},
			},

			// === Control flow ===
			{
				Code: `if (obj.__proto__) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 5},
				},
			},
			{
				Code: `for (var x in obj.__proto__) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 15},
				},
			},
			{
				Code: `for (var x of obj.__proto__) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 15},
				},
			},
			{
				Code: `while (obj.__proto__) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 8},
				},
			},
			{
				Code: `switch (obj.__proto__) { case 0: break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 9},
				},
			},
			{
				Code: `throw obj.__proto__;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 7},
				},
			},
			// do-while
			{
				Code: `do {} while (obj.__proto__);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 14},
				},
			},
			// try / catch / finally
			{
				Code: `try { obj.__proto__; } catch(e) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 7},
				},
			},
			{
				Code: `try {} catch(e) { obj.__proto__; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 19},
				},
			},
			{
				Code: `try {} finally { obj.__proto__; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 18},
				},
			},
			// For init / update
			{
				Code: `for (obj.__proto__ = 0;;) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 6},
				},
			},

			// === Multiple occurrences ===
			{
				Code: "a.__proto__;\nb.__proto__;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 1},
					{MessageId: "unexpectedProto", Line: 2, Column: 1},
				},
			},
			// Mixed dot and bracket
			{
				Code: `obj.__proto__; obj["__proto__"];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 1},
					{MessageId: "unexpectedProto", Line: 1, Column: 16},
				},
			},

			// === Multi-byte characters ===
			{
				Code: "/* 🚀 */ obj.__proto__",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedProto", Line: 1, Column: 10},
				},
			},
		},
	)
}

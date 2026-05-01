package max_lines_per_function

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestMaxLinesPerFunction exercises three overlapping corpora in one run:
//
//  1. ESLint parity — ports of every case from
//     eslint/tests/lib/rules/max-lines-per-function.js, asserting matching
//     messageId + Description (the data fields exercised by upstream).
//  2. Additional coverage on top of upstream — TS-specific syntax (private
//     methods, async generators), tsgo AST quirks (multi-paren IIFE wrappers,
//     class-field arrows), and full Line/Column position assertions.
//  3. Locks in upstream branches that ESLint's own suite doesn't cover (e.g.
//     `function expression with own id` name extraction; method on an object
//     literal vs a class body).
func TestMaxLinesPerFunction(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&MaxLinesPerFunctionRule,
		[]rule_tester.ValidTestCase{
			// ============================================================
			// 1. ESLint parity — mirrors tests/lib/rules/max-lines-per-function.js
			// ============================================================

			// Test code in global scope doesn't count
			{Code: "var x = 5;\nvar x = 2;\n", Options: 1},

			// Test single line standalone function
			{Code: "function name() {}", Options: 1},

			// Test standalone function with lines of code
			{Code: "function name() {\nvar x = 5;\nvar x = 2;\n}", Options: 4},

			// Test inline arrow function
			{Code: "const bar = () => 2", Options: 1},

			// Test arrow function
			{Code: "const bar = () => {\nconst x = 2 + 1;\nreturn x;\n}", Options: 4},

			// skipBlankLines: false with simple standalone function
			{
				Code:    "function name() {\nvar x = 5;\n\t\n \n\nvar x = 2;\n}",
				Options: map[string]interface{}{"max": 7, "skipComments": false, "skipBlankLines": false},
			},

			// skipBlankLines: true with simple standalone function
			{
				Code:    "function name() {\nvar x = 5;\n\t\n \n\nvar x = 2;\n}",
				Options: map[string]interface{}{"max": 4, "skipComments": false, "skipBlankLines": true},
			},

			// skipComments: true with an individual single line comment
			{
				Code:    "function name() {\nvar x = 5;\nvar x = 2; // end of line comment\n}",
				Options: map[string]interface{}{"max": 4, "skipComments": true, "skipBlankLines": false},
			},

			// skipComments: true with an individual single line comment
			{
				Code:    "function name() {\nvar x = 5;\n// a comment on it's own line\nvar x = 2; // end of line comment\n}",
				Options: map[string]interface{}{"max": 4, "skipComments": true, "skipBlankLines": false},
			},

			// skipComments: true with single line comments
			{
				Code:    "function name() {\nvar x = 5;\n// a comment on it's own line\n// and another line comment\nvar x = 2; // end of line comment\n}",
				Options: map[string]interface{}{"max": 4, "skipComments": true, "skipBlankLines": false},
			},

			// skipComments: true test with multiple different comment types
			{
				Code:    "function name() {\nvar x = 5;\n/* a \n multi \n line \n comment \n*/\n\nvar x = 2; // end of line comment\n}",
				Options: map[string]interface{}{"max": 5, "skipComments": true, "skipBlankLines": false},
			},

			// skipComments: true with multiple different comment types, including trailing and leading whitespace
			{
				Code:    "function name() {\nvar x = 5;\n\t/* a comment with leading whitespace */\n/* a comment with trailing whitespace */\t\t\n\t/* a comment with trailing and leading whitespace */\t\t\n/* a \n multi \n line \n comment \n*/\t\t\n\nvar x = 2; // end of line comment\n}",
				Options: map[string]interface{}{"max": 5, "skipComments": true, "skipBlankLines": false},
			},

			// Multiple params on separate lines test
			{
				Code: `function foo(
    aaa = 1,
    bbb = 2,
    ccc = 3
) {
    return aaa + bbb + ccc
}`,
				Options: map[string]interface{}{"max": 7, "skipComments": true, "skipBlankLines": false},
			},

			// IIFE validity test
			{
				Code: `(
function
()
{
}
)
()`,
				Options: map[string]interface{}{"max": 4, "skipComments": true, "skipBlankLines": false, "IIFEs": true},
			},

			// Nested function validity test
			{
				Code: `function parent() {
var x = 0;
function nested() {
    var y = 0;
    x = 2;
}
if ( x === y ) {
    x++;
}
}`,
				Options: map[string]interface{}{"max": 10, "skipComments": true, "skipBlankLines": false},
			},

			// Class method validity test
			{
				Code: `class foo {
    method() {
        let y = 10;
        let x = 20;
        return y + x;
    }
}`,
				Options: map[string]interface{}{"max": 5, "skipComments": true, "skipBlankLines": false},
			},

			// IIFEs should be recognized if IIFEs: true
			{
				Code: `(function(){
    let x = 0;
    let y = 0;
    let z = x + y;
    let foo = {};
    return bar;
}());`,
				Options: map[string]interface{}{"max": 7, "skipComments": true, "skipBlankLines": false, "IIFEs": true},
			},

			// IIFEs should not be recognized if IIFEs: false
			{
				Code: `(function(){
    let x = 0;
    let y = 0;
    let z = x + y;
    let foo = {};
    return bar;
}());`,
				Options: map[string]interface{}{"max": 2, "skipComments": true, "skipBlankLines": false, "IIFEs": false},
			},

			// Arrow IIFEs should be recognized if IIFEs: true
			{
				Code: `(() => {
    let x = 0;
    let y = 0;
    let z = x + y;
    let foo = {};
    return bar;
})();`,
				Options: map[string]interface{}{"max": 7, "skipComments": true, "skipBlankLines": false, "IIFEs": true},
			},

			// Arrow IIFEs should not be recognized if IIFEs: false
			{
				Code: `(() => {
    let x = 0;
    let y = 0;
    let z = x + y;
    let foo = {};
    return bar;
})();`,
				Options: map[string]interface{}{"max": 2, "skipComments": true, "skipBlankLines": false, "IIFEs": false},
			},

			// ============================================================
			// 2. Additional edge cases
			// ============================================================

			// --- Option shapes ---
			{Code: "function f() {}", Options: 1},
			{Code: "function f() {}", Options: []interface{}{1}},
			{Code: "function f() {}", Options: []interface{}{map[string]interface{}{"max": 1}}},
			{Code: "function f() {}"}, // No options → default max=50
			{Code: "function f() {\nreturn 1;\n}", Options: []interface{}{}},
			// Bare object form (single-option CLI shape)
			{Code: "function f() {}", Options: map[string]interface{}{"max": 1}},
			// Array-wrapped empty object → defaults
			{Code: "function f() {\nreturn 1;\n}", Options: []interface{}{map[string]interface{}{}}},

			// --- Multi-paren IIFE wrappers (tsgo preserves parens; ESLint flattens) ---
			{
				Code:    "((function () {\nlet x = 0;\nlet y = 0;\n}))();",
				Options: map[string]interface{}{"max": 1, "IIFEs": false},
			},
			{
				Code:    "((() => {\nlet x = 0;\nlet y = 0;\n}))();",
				Options: map[string]interface{}{"max": 1, "IIFEs": false},
			},

			// --- Function expression passed as argument is NOT an IIFE ---
			// (callee is `setTimeout`, not the function — must be checked)
			{
				Code:    "setTimeout(function () {\nlet x = 0;\nlet y = 0;\nlet z = 0;\n}, 100);",
				Options: map[string]interface{}{"max": 100},
			},

			// --- Async / generator variants (under default max) ---
			{Code: "async function f() {\nawait g();\n}", Options: 3},
			{Code: "function* gen() {\nyield 1;\nyield 2;\n}", Options: 4},
			{Code: "async function* gen() {\nyield await g();\n}", Options: 3},

			// --- Class-field arrow (parent is PropertyDeclaration, not method) ---
			{
				Code: `class A {
foo = () => {
return 1;
}
}`,
				Options: 4,
			},

			// --- Object property arrow (parent is PropertyAssignment, not method) ---
			{
				Code:    "var o = { foo: () => {\nreturn 1;\n} }",
				Options: 3,
			},

			// --- Private method ---
			{
				Code: `class A {
#foo() {
return 1;
}
}`,
				Options: 3,
			},

			// --- Class expression ---
			{
				Code: `var A = class {
method() {
return 1;
}
};`,
				Options: 3,
			},

			// --- Empty function body ---
			{Code: "function f() {}", Options: 1},
			{Code: "var f = () => {};", Options: 1},
			{Code: "class A { method() {} }", Options: 1},
			{Code: "class A { constructor() {} }", Options: 1},

			// --- Functions in default parameters / type contexts ---
			{
				Code:    "function f(cb = () => {\nreturn 1;\n}) {\nreturn cb();\n}",
				Options: 5,
			},

			// --- Methods on object literals (not classes) — counts as Method ---
			{
				Code: `var o = {
shorthand() {
return 1;
}
};`,
				Options: 3,
			},

			// --- TS body-absent declarations: overload signatures, abstract,
			//     declare. ESLint never visits these (no FunctionExpression
			//     value); we mirror by filtering on `node.Body() == nil`.
			{
				Code: `class A {
foo(): void;
foo(x: string): void;
foo(x?: any) { return x; }
}`,
				Options: 1, // Only the implementation is counted; it's 1 line.
			},
			{
				Code: `abstract class A {
abstract foo(): void;
}`,
				Options: 0, // Abstract method has no body → not visited.
			},
			{
				Code: `declare class A {
foo(): void;
}`,
				Options: 0,
			},
			{
				Code: `interface I {
foo(): void;
bar(x: string): void;
}`,
				Options: 0, // MethodSignature isn't in our listeners at all.
			},

			// --- Class static blocks are NOT visited (ESLint also doesn't).
			{
				Code: `class A {
static {
let x = 1;
let y = 2;
let z = 3;
let w = 4;
}
}`,
				Options: 0,
			},

			// --- Generator method (non-async)
			{
				Code: `class A {
*gen() {
yield 1;
}
}`,
				Options: 3,
			},

			// --- Async method
			{
				Code: `class A {
async foo() {
await g();
}
}`,
				Options: 3,
			},

			// --- Computed method with string literal name → "Method 'foo'"
			{
				Code: `class A {
['foo']() {
return 1;
}
}`,
				Options: 3,
			},

			// --- Static getter / private getter / static setter
			{
				Code: `class A {
static get x() {
return 1;
}
}`,
				Options: 3,
			},
			{
				Code: `class A {
get #x() {
return 1;
}
}`,
				Options: 3,
			},
			{
				Code: `class A {
static set x(v) {
this._x = v;
}
}`,
				Options: 3,
			},

			// --- IIFE in return position (still IIFE — callee of CallExpression)
			//     The inner function is skipped under IIFEs:false; the outer
			//     spans 5 lines, so max=5 keeps the case valid.
			{
				Code: `function outer() {
return (function () {
return 1;
})();
}`,
				Options: map[string]interface{}{"max": 5, "IIFEs": false},
			},

			// --- Deeply nested arrow chains (each arrow visited independently)
			{
				Code:    "var f = () => () => () => 1;",
				Options: 1,
			},

			// --- Function inside Class field arrow inside method
			//     The MethodDeclaration / inner ArrowFunction / inner
			//     FunctionExpression are all visited; under generous max,
			//     none should fire.
			{
				Code: `class A {
method() {
const cb = () => function () {
return 1;
};
return cb;
}
}`,
				Options: 100,
			},

			// --- TS type wrappers in IIFE callee position: should still be
			//     recognized as IIFE. tsgo represents `<T>(fn)<U>()` differently
			//     from ESTree, but the call's `Expression` should still trace
			//     back through paren wrappers to the function.
			{
				Code:    "(function () {\nlet x = 0;\nlet y = 0;\n})();",
				Options: map[string]interface{}{"max": 1, "IIFEs": false},
			},

			// --- TS parameter properties (constructor with `public`/`private`)
			{
				Code: `class A {
constructor(public x: number, private y: string) {
this.z = x;
}
}`,
				Options: 4,
			},

			// --- Default exports
			{
				Code: `export default function () {
return 1;
}`,
				Options: 3,
			},
			{
				Code: `export default () => {
return 1;
};`,
				Options: 3,
			},
			{
				Code: `export default function named() {
return 1;
}`,
				Options: 3,
			},

			// --- Decorator on method: tsgo includes decorator in the node range
			//     (matching ESLint's MethodDefinition.loc which also starts at
			//     the decorator). The line count includes the decorator line.
			{
				Code: `class A {
@dec
method() {
return 1;
}
}`,
				Options: 4,
			},

			// --- Class with multiple modifiers stacked (TS-specific)
			{
				Code: `class A {
public static async foo() {
await g();
return 1;
}
}`,
				Options: 4,
			},

			// --- Multibyte source — column/line counting must use byte offsets
			//     consistently, multi-byte chars don't break full-line-comment
			//     detection.
			{
				Code: `function 日本語() {
const 中文 = 1;
return 中文;
}`,
				Options: 4,
			},

			// --- Emoji in source / comments (skipComments path)
			{
				Code: `function f() {
// 🎉 celebrate
return 1;
}`,
				Options: map[string]interface{}{"max": 3, "skipComments": true},
			},

			// --- BOM / NBSP / U+2028 / U+2029 in blank line detection
			{
				Code:    "function f() {\n\u00a0\u00a0\n\ufeff\nreturn 1;\n}",
				Options: map[string]interface{}{"max": 3, "skipBlankLines": true},
			},

			// --- CRLF line endings in non-trivial layouts: 5 lines, 2 blank
			//     → 3 counted, max=3 → valid.
			{
				Code:    "function name() {\r\n\r\n\r\nreturn 1;\r\n}",
				Options: map[string]interface{}{"max": 3, "skipBlankLines": true},
			},

			// --- Mixed line terminators (LF + CR + CRLF + LS + PS)
			{
				Code:    "function f() {\na;\rb;\r\nc; d; e;\n}",
				Options: 7,
			},

			// --- Hashbang isn't a function-line concern (always at file start),
			//     but the rule shouldn't choke on files with one.
			{
				Code:    "#!/usr/bin/env node\nfunction f() {\nreturn 1;\n}",
				Options: 3,
			},

			// --- Realistic user-shape: object literal with 5 method-shorthand
			//     entries — each Method is reported independently. Each fits
			//     under generous max.
			{
				Code: `const api = {
get() { return 1; },
post() { return 2; },
put() { return 3; },
del() { return 4; },
patch() { return 5; },
};`,
				Options: 1,
			},

			// --- Realistic user-shape: React-style class with lifecycle methods,
			//     all individually short.
			{
				Code: `class Component {
componentDidMount() { this.fetch(); }
componentWillUnmount() { this.cleanup(); }
render() { return null; }
}`,
				Options: 1,
			},

			// --- Realistic user-shape: builder pattern — chained methods on
			//     class expression, returning `this`.
			{
				Code: `var Builder = class {
withFoo() { return this; }
withBar() { return this; }
build() { return {}; }
};`,
				Options: 1,
			},

			// --- Realistic user-shape: callback-heavy code — function as last
			//     argument is NOT an IIFE.
			{
				Code: `arr.reduce(function (acc, v) {
acc.push(v);
return acc;
}, []);`,
				Options: 4,
			},

			// --- skipComments + JSDoc above the function — JSDoc is OUTSIDE
			//     the function range, so doesn't affect the count. (Function
			//     range starts at `function`/`name`, not at the JSDoc.)
			{
				Code: `/**
 * Description
 * @param x number
 */
function f(x) {
return x;
}`,
				Options: map[string]interface{}{"max": 3, "skipComments": true},
			},

			// --- Comment at exact column boundary (ESLint's isFullLineComment
			//     uses .trim() === "" — equivalent to IsECMABlankLine in our port)
			{
				Code: `function f() {
   /* leading whitespace */
return 1;
}`,
				Options: map[string]interface{}{"max": 3, "skipComments": true},
			},
			{
				Code: `function f() {
/* trailing whitespace */   ` + `
return 1;
}`,
				Options: map[string]interface{}{"max": 3, "skipComments": true},
			},

			// --- Tagged template: function expression as tag is NOT IIFE
			//     (TaggedTemplateExpression, not CallExpression).
			{
				Code: "(function f() {\nreturn 1;\n})`tpl`;",
				Options: map[string]interface{}{"max": 3, "IIFEs": false},
			},

			// --- new (function(){})() — NewExpression callee is the function,
			//     but ESLint's isIIFE checks CallExpression specifically. Same
			//     in our port — NewExpression is its own kind, so the inner
			//     function IS visited even with IIFEs:false. Function spans
			//     3 lines, so max=3 keeps the case valid.
			{
				Code:    "new (function () {\nthis.x = 1;\n})();",
				Options: map[string]interface{}{"max": 3, "IIFEs": false},
			},

			// --- Bound function call: `(function(){}).bind(this)()` is NOT
			//     an IIFE — the outer call's callee is `.bind(this)`, not the
			//     function. Function should be checked normally.
			{
				Code: `var fn = (function () {
return 1;
}).bind(this);`,
				Options: 3,
			},

			// --- Empty class with empty constructor
			{Code: "class A { constructor() {} }", Options: 1},

			// --- Malformed-edge: function with empty parameter list and body
			//     declared on same line
			{Code: "function f(){}", Options: 1},
			{Code: "()=>{}", Options: 1},
			{Code: "()=>1", Options: 1},

			// --- Function expression in array
			{
				Code: `var arr = [function () {
return 1;
}];`,
				Options: 3,
			},

			// --- Function as default export and re-exported
			{
				Code:    "export const f = function () {\nreturn 1;\n};",
				Options: 3,
			},

			// --- Class expression method (private/static accessor identification
			//     must work even though parent is ClassExpression, not Class
			//     Declaration). Verifies inClassBody handles both kinds.
			{
				Code: `var A = class {
static #priv() {
return 1;
}
};`,
				Options: 4,
			},

			// --- TS type-only constructs are NOT visited:
			//   - call signature: `type F = (x) => void`
			//   - method signature: `interface I { m(): void }`
			//   - construct signature: `interface I { new (): void }`
			//   - function type literal as parameter type
			{
				Code: `type F = (x: number) => void;
type G = { m(): void };
type H = { new (): H };
function f(cb: (x: number) => void) {
cb(1);
}`,
				Options: 3,
			},

			// --- Trailing comment inside function body (just before `})`).
			//     skipComments should treat the trailing-comment line as comment.
			{
				Code: `function f() {
let x = 1;
// trailing
}`,
				Options: map[string]interface{}{"max": 3, "skipComments": true},
			},

			// --- Comments at exact word boundaries: `var/*c*/x` — comment
			//     interleaved between tokens on same line. NOT a full-line
			//     comment because there's code on both sides.
			{
				Code: `function f() {
var/*c*/x = 1;
return x;
}`,
				Options: map[string]interface{}{"max": 4, "skipComments": true},
			},

			// --- Callback nested inside component-style return: arrow as
			//     argument inside an object literal returned from another
			//     arrow. Inner arrow visited; not IIFE (it's a value, not a
			//     callee).
			{
				Code: `const Button = () => ({ onClick: () => {
console.log('clicked');
console.log('twice');
} });`,
				Options: map[string]interface{}{"max": 100},
			},

			// --- Generator method on object literal (not class)
			{
				Code: `var obj = {
*gen() {
yield 1;
}
};`,
				Options: 3,
			},

			// --- Class expression with computed key + static modifier
			{
				Code: `var A = class {
static [Symbol.iterator]() {
return this;
}
};`,
				Options: 4,
			},

			// --- Optional method (TS): `foo?(): void` is a TSMethodSignature in
			//     interface contexts. Should not be visited.
			{
				Code: `interface I {
foo?(): void;
}`,
				Options: 0,
			},

			// --- Function expression with own id matching outer var name (the
			//     function's name token wins, not the var).
			{
				Code:    "var f = function f() {\nreturn 1;\n};",
				Options: 3,
			},

			// --- Module-level top-level await wrapped in IIFE pattern.
			{
				Code: `(async () => {
const x = await fetch('/api');
console.log(x);
})();`,
				Options: map[string]interface{}{"max": 1, "IIFEs": false},
			},

			// --- Two IIFEs in same file, different forms.
			{
				Code: `(function () { return 1; })();
(() => 2)();
(function () { return 3; }());`,
				Options: map[string]interface{}{"max": 1, "IIFEs": false},
			},

			// --- Nested object literal with method shorthand inside method.
			{
				Code: `class A {
configure() {
return {
build() { return 1; },
};
}
}`,
				Options: 5,
			},
		},
		[]rule_tester.InvalidTestCase{
			// ============================================================
			// 1. ESLint parity — mirrors tests/lib/rules/max-lines-per-function.js
			// ============================================================

			// Test simple standalone function is recognized
			{
				Code:    "function name() {\n}",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Function 'name' has too many lines (2). Maximum allowed is 1.",
					},
				},
			},

			// Test anonymous function assigned to variable is recognized — name
			// resolved via parent VariableDeclaration to match ESLint's getName.
			{
				Code:    "var func = function() {\n}",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Function 'func' has too many lines (2). Maximum allowed is 1.",
					},
				},
			},

			// Test arrow functions are recognized — name resolved via parent
			// VariableDeclaration to match ESLint's getName.
			{
				Code:    "const bar = () => {\nconst x = 2 + 1;\nreturn x;\n}",
				Options: 3,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Arrow function 'bar' has too many lines (4). Maximum allowed is 3.",
					},
				},
			},

			// Test inline arrow functions are recognized
			{
				Code:    "const bar = () =>\n 2",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Arrow function 'bar' has too many lines (2). Maximum allowed is 1.",
					},
				},
			},

			// Test that option defaults work as expected
			{
				Code:    "() => {" + strings.Repeat("foo\n", 60) + "}",
				Options: map[string]interface{}{},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Arrow function has too many lines (61). Maximum allowed is 50.",
					},
				},
			},

			// Test skipBlankLines: false
			{
				Code:    "function name() {\nvar x = 5;\n\t\n \n\nvar x = 2;\n}",
				Options: map[string]interface{}{"max": 6, "skipComments": false, "skipBlankLines": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Function 'name' has too many lines (7). Maximum allowed is 6.",
					},
				},
			},

			// Test skipBlankLines: false with CRLF line endings
			{
				Code:    "function name() {\r\nvar x = 5;\r\n\t\r\n \r\n\r\nvar x = 2;\r\n}",
				Options: map[string]interface{}{"max": 6, "skipComments": true, "skipBlankLines": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Function 'name' has too many lines (7). Maximum allowed is 6.",
					},
				},
			},

			// Test skipBlankLines: true
			{
				Code:    "function name() {\nvar x = 5;\n\t\n \n\nvar x = 2;\n}",
				Options: map[string]interface{}{"max": 2, "skipComments": true, "skipBlankLines": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Function 'name' has too many lines (4). Maximum allowed is 2.",
					},
				},
			},

			// Test skipBlankLines: true with CRLF line endings
			{
				Code:    "function name() {\r\nvar x = 5;\r\n\t\r\n \r\n\r\nvar x = 2;\r\n}",
				Options: map[string]interface{}{"max": 2, "skipComments": true, "skipBlankLines": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Function 'name' has too many lines (4). Maximum allowed is 2.",
					},
				},
			},

			// Test skipComments: true and skipBlankLines: false for multiple types of comment
			{
				Code:    "function name() { // end of line comment\nvar x = 5; /* mid line comment */\n\t// single line comment taking up whole line\n\t\n \n\nvar x = 2;\n}",
				Options: map[string]interface{}{"max": 6, "skipComments": true, "skipBlankLines": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Function 'name' has too many lines (7). Maximum allowed is 6.",
					},
				},
			},

			// Test skipComments: true and skipBlankLines: true for multiple types of comment
			{
				Code:    "function name() { // end of line comment\nvar x = 5; /* mid line comment */\n\t// single line comment taking up whole line\n\t\n \n\nvar x = 2;\n}",
				Options: map[string]interface{}{"max": 1, "skipComments": true, "skipBlankLines": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Function 'name' has too many lines (4). Maximum allowed is 1.",
					},
				},
			},

			// Test skipComments: false and skipBlankLines: true for multiple types of comment
			{
				Code:    "function name() { // end of line comment\nvar x = 5; /* mid line comment */\n\t// single line comment taking up whole line\n\t\n \n\nvar x = 2;\n}",
				Options: map[string]interface{}{"max": 1, "skipComments": false, "skipBlankLines": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Function 'name' has too many lines (5). Maximum allowed is 1.",
					},
				},
			},

			// Test simple standalone function with params on separate lines
			{
				Code: `function foo(
    aaa = 1,
    bbb = 2,
    ccc = 3
) {
    return aaa + bbb + ccc
}`,
				Options: map[string]interface{}{"max": 2, "skipComments": true, "skipBlankLines": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Function 'foo' has too many lines (7). Maximum allowed is 2.",
					},
				},
			},

			// Test IIFE "function" keyword is included in the count
			{
				Code: `(
function
()
{
}
)
()`,
				Options: map[string]interface{}{"max": 2, "skipComments": true, "skipBlankLines": false, "IIFEs": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Function has too many lines (4). Maximum allowed is 2.",
					},
				},
			},

			// Nested functions are included in their parent's function count.
			{
				Code: `function parent() {
var x = 0;
function nested() {
    var y = 0;
    x = 2;
}
if ( x === y ) {
    x++;
}
}`,
				Options: map[string]interface{}{"max": 9, "skipComments": true, "skipBlankLines": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Function 'parent' has too many lines (10). Maximum allowed is 9.",
					},
				},
			},

			// Both parent and nested when both exceed max
			{
				Code: `function parent() {
var x = 0;
function nested() {
    var y = 0;
    x = 2;
}
if ( x === y ) {
    x++;
}
}`,
				Options: map[string]interface{}{"max": 2, "skipComments": true, "skipBlankLines": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Function 'parent' has too many lines (10). Maximum allowed is 2.",
					},
					{
						MessageId: "exceed",
						Message:   "Function 'nested' has too many lines (4). Maximum allowed is 2.",
					},
				},
			},

			// Test regular methods are recognized
			{
				Code: `class foo {
    method() {
        let y = 10;
        let x = 20;
        return y + x;
    }
}`,
				Options: map[string]interface{}{"max": 2, "skipComments": true, "skipBlankLines": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Method 'method' has too many lines (5). Maximum allowed is 2.",
					},
				},
			},

			// Test static methods are recognized
			{
				Code: `class A {
    static
    foo
    (a) {
        return a
    }
}`,
				Options: map[string]interface{}{"max": 2, "skipComments": true, "skipBlankLines": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Static method 'foo' has too many lines (5). Maximum allowed is 2.",
					},
				},
			},

			// Test getters are recognized as properties
			{
				Code: `var obj = {
    get
    foo
    () {
        return 1
    }
}`,
				Options: map[string]interface{}{"max": 2, "skipComments": true, "skipBlankLines": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Getter 'foo' has too many lines (5). Maximum allowed is 2.",
					},
				},
			},

			// Test setters are recognized as properties
			{
				Code: `var obj = {
    set
    foo
    ( val ) {
        this._foo = val;
    }
}`,
				Options: map[string]interface{}{"max": 2, "skipComments": true, "skipBlankLines": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Setter 'foo' has too many lines (5). Maximum allowed is 2.",
					},
				},
			},

			// Test computed property names
			{
				Code: `class A {
    static
    [
        foo +
            bar
    ]
    (a) {
        return a
    }
}`,
				Options: map[string]interface{}{"max": 2, "skipComments": true, "skipBlankLines": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Static method has too many lines (8). Maximum allowed is 2.",
					},
				},
			},

			// Test the IIFEs option includes IIFEs
			{
				Code: `(function(){
    let x = 0;
    let y = 0;
    let z = x + y;
    let foo = {};
    return bar;
}());`,
				Options: map[string]interface{}{"max": 2, "skipComments": true, "skipBlankLines": false, "IIFEs": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Function has too many lines (7). Maximum allowed is 2.",
					},
				},
			},

			// Test the IIFEs option includes arrow IIFEs
			{
				Code: `(() => {
    let x = 0;
    let y = 0;
    let z = x + y;
    let foo = {};
    return bar;
})();`,
				Options: map[string]interface{}{"max": 2, "skipComments": true, "skipBlankLines": false, "IIFEs": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Arrow function has too many lines (7). Maximum allowed is 2.",
					},
				},
			},

			// ============================================================
			// 2. Additional edge cases — TS-specific syntax + position assertions
			// ============================================================

			// Position assertions: function declaration spans 2 lines, max=1.
			// Range: from `f` (col 1, line 1) through the closing `}` (col 1, line 2).
			{
				Code:    "function f() {\n}",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Function 'f' has too many lines (2). Maximum allowed is 1.",
						Line:      1,
						Column:    1,
						EndLine:   2,
						EndColumn: 2,
					},
				},
			},

			// Position assertions: arrow function in const decl. Range covers `() => { ... }`.
			{
				Code:    "const bar = () => {\nreturn 1;\n}",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Arrow function 'bar' has too many lines (3). Maximum allowed is 1.",
						Line:      1,
						Column:    13,
						EndLine:   3,
						EndColumn: 2,
					},
				},
			},

			// Async arrow function — description includes "async" + parent-walked name
			{
				Code:    "const bar = async () => {\nawait g();\nreturn 1;\n}",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Async arrow function 'bar' has too many lines (4). Maximum allowed is 2.",
					},
				},
			},

			// Generator function declaration — "Generator function 'gen'"
			{
				Code:    "function* gen() {\nyield 1;\nyield 2;\n}",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Generator function 'gen' has too many lines (4). Maximum allowed is 2.",
					},
				},
			},

			// Async function declaration — "Async function 'foo'"
			{
				Code:    "async function foo() {\nawait g();\n}",
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Async function 'foo' has too many lines (3). Maximum allowed is 1.",
					},
				},
			},

			// Async generator method
			{
				Code: `class A {
async *foo() {
yield await g();
}
}`,
				Options: map[string]interface{}{"max": 2},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Async generator method 'foo' has too many lines (3). Maximum allowed is 2.",
					},
				},
			},

			// Private method
			{
				Code: `class A {
#foo() {
return 1;
}
}`,
				Options: map[string]interface{}{"max": 2},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Private method '#foo' has too many lines (3). Maximum allowed is 2.",
					},
				},
			},

			// Constructor
			{
				Code: `class A {
constructor() {
this.x = 1;
this.y = 2;
}
}`,
				Options: map[string]interface{}{"max": 2},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Constructor has too many lines (4). Maximum allowed is 2.",
					},
				},
			},

			// Function expression with own id name
			{
				Code:    "var x = function bar() {\nreturn 1;\nreturn 2;\n};",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Function 'bar' has too many lines (4). Maximum allowed is 2.",
					},
				},
			},

			// Class expression method
			{
				Code: `var A = class {
method() {
return 1;
return 2;
}
};`,
				Options: map[string]interface{}{"max": 2},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Method 'method' has too many lines (4). Maximum allowed is 2.",
					},
				},
			},

			// Class field arrow — counts as arrow function (parent is
			// PropertyDeclaration, not a method); name is resolved from the
			// property key, matching ESLint's getName which inspects
			// PropertyDefinition.key.
			{
				Code: `class A {
foo = () => {
return 1;
return 2;
}
}`,
				Options: map[string]interface{}{"max": 2},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Arrow function 'foo' has too many lines (4). Maximum allowed is 2.",
					},
				},
			},

			// Object literal arrow assignment — name resolved from property key.
			{
				Code:    "var o = { foo: () => {\nreturn 1;\nreturn 2;\n} }",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Arrow function 'foo' has too many lines (4). Maximum allowed is 2.",
					},
				},
			},

			// Object literal method shorthand — counts as Method 'foo'
			{
				Code: `var o = {
foo() {
return 1;
return 2;
}
};`,
				Options: map[string]interface{}{"max": 2},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Method 'foo' has too many lines (4). Maximum allowed is 2.",
					},
				},
			},

			// Multi-paren-wrapped IIFE recognized when IIFEs: true.
			// `((function(){}))()` has two ParenthesizedExpression wrappers
			// in tsgo; ESTree flattens them, but our isIIFE walks past both.
			{
				Code:    "((function () {\nlet x = 0;\nlet y = 0;\nlet z = 0;\n}))();",
				Options: map[string]interface{}{"max": 2, "IIFEs": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Function has too many lines (5). Maximum allowed is 2.",
					},
				},
			},

			// Function expression argument is NOT an IIFE (callee is `cb`, not the function).
			{
				Code:    "cb(function () {\nlet x = 0;\nlet y = 0;\nlet z = 0;\n});",
				Options: map[string]interface{}{"max": 2, "IIFEs": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Function has too many lines (5). Maximum allowed is 2.",
					},
				},
			},

			// Locks in upstream `processFunction` arms:
			// 1. function-in-function (parent and nested both exceed) — covered above.
			// 2. isFullLineComment two-arms quirk: `/* a */ /* b */` on a line
			//    between code → NOT full-line (last comment is /* b */, /* a */
			//    precedes it on the same line, so isFirstTokenOnLine=false).
			{
				Code: `function f() {
/* a */ /* b */
return 1;
}`,
				Options: map[string]interface{}{"max": 2, "skipComments": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Function 'f' has too many lines (4). Maximum allowed is 2.",
					},
				},
			},

			// Locks in: a multi-line block comment with following code on the
			// last line — the last line is NOT skipped (code follows comment).
			{
				Code: `function f() {
/* multi
line comment */ var x = 1;
return x;
}`,
				Options: map[string]interface{}{"max": 3, "skipComments": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Function 'f' has too many lines (4). Maximum allowed is 3.",
					},
				},
			},

			// Generator method (non-async): "Generator method 'gen'"
			{
				Code: `class A {
*gen() {
yield 1;
yield 2;
}
}`,
				Options: map[string]interface{}{"max": 2},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Generator method 'gen' has too many lines (4). Maximum allowed is 2.",
					},
				},
			},

			// Async method: "Async method 'foo'"
			{
				Code: `class A {
async foo() {
await g();
return 1;
}
}`,
				Options: map[string]interface{}{"max": 2},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Async method 'foo' has too many lines (4). Maximum allowed is 2.",
					},
				},
			},

			// Computed method with string literal name → "Method 'foo'"
			// (GetStaticPropertyName resolves "['foo']" to "foo".)
			{
				Code: `class A {
['foo']() {
return 1;
return 2;
}
}`,
				Options: map[string]interface{}{"max": 2},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Method 'foo' has too many lines (4). Maximum allowed is 2.",
					},
				},
			},

			// Private getter
			{
				Code: `class A {
get #x() {
return 1;
return 2;
}
}`,
				Options: map[string]interface{}{"max": 2},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Private getter '#x' has too many lines (4). Maximum allowed is 2.",
					},
				},
			},

			// Static setter
			{
				Code: `class A {
static set x(v) {
this._x = v;
this._y = v;
}
}`,
				Options: map[string]interface{}{"max": 2},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Static setter 'x' has too many lines (4). Maximum allowed is 2.",
					},
				},
			},

			// IIFE in return position is still IIFE — callee of CallExpression
			// even though wrapped in a `return` statement.
			{
				Code: `function outer() {
return (function () {
let a = 1;
let b = 2;
})();
}`,
				Options: map[string]interface{}{"max": 2, "IIFEs": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Function 'outer' has too many lines (6). Maximum allowed is 2.",
					},
					{
						MessageId: "exceed",
						Message:   "Function has too many lines (4). Maximum allowed is 2.",
					},
				},
			},

			// Triple-nested arrow chain — each arrow gets its own visit. The
			// outermost spans 4 lines; max=1 should report all three. Only the
			// outermost is bound to a variable, so only it carries a name.
			{
				Code: `var f = () =>
() =>
() =>
1;`,
				Options: 1,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Arrow function 'f' has too many lines (4). Maximum allowed is 1.",
					},
					{
						MessageId: "exceed",
						Message:   "Arrow function has too many lines (3). Maximum allowed is 1.",
					},
					{
						MessageId: "exceed",
						Message:   "Arrow function has too many lines (2). Maximum allowed is 1.",
					},
				},
			},

			// ============================================================
			// 3. Real-user shapes & structural edge cases
			// ============================================================

			// Decorator on method — line count includes the decorator line
			// (tsgo's MethodDeclaration node range starts at `@`, matching
			// ESLint's MethodDefinition.loc).
			{
				Code: `class A {
@dec
method() {
return 1;
}
}`,
				Options: map[string]interface{}{"max": 3},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Method 'method' has too many lines (4). Maximum allowed is 3.",
					},
				},
			},

			// Multibyte source — `name` includes non-ASCII identifier
			{
				Code: `function 日本語() {
const x = 1;
return x;
const y = 2;
}`,
				Options: 4,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Function '日本語' has too many lines (5). Maximum allowed is 4.",
					},
				},
			},

			// Multibyte in skipComments path — emoji in comment shouldn't
			// throw off byte-offset slicing in isFullLineCommentLine.
			{
				Code: `function f() {
// 🎉 line 1
// 🎊 line 2
let x = 1;
let y = 2;
return x + y;
}`,
				Options: map[string]interface{}{"max": 3, "skipComments": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Function 'f' has too many lines (5). Maximum allowed is 3.",
					},
				},
			},

			// CRLF + skipBlankLines
			{
				Code:    "function name() {\r\nlet a = 1;\r\n\r\n\r\nlet b = 2;\r\nreturn a + b;\r\n}",
				Options: map[string]interface{}{"max": 3, "skipBlankLines": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Function 'name' has too many lines (5). Maximum allowed is 3.",
					},
				},
			},

			// Mixed terminators (LF + CR + CRLF + U+2028 + U+2029)
			{
				Code:    "function f() {\na;\rb;\r\nc; d; e;\n}",
				Options: 4,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Function 'f' has too many lines (7). Maximum allowed is 4.",
					},
				},
			},

			// Hashbang + function — hashbang is on a separate line, function
			// range starts at line 2.
			{
				Code:    "#!/usr/bin/env node\nfunction f() {\nlet a = 1;\nlet b = 2;\nreturn a + b;\n}",
				Options: 3,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Function 'f' has too many lines (5). Maximum allowed is 3.",
					},
				},
			},

			// JSDoc above function is OUTSIDE the function range → does NOT
			// inflate the count. The function spans only its body lines.
			{
				Code: `/**
 * Description
 * @param x number
 * @returns x squared
 */
function f(x) {
let y = x * x;
return y;
}`,
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Function 'f' has too many lines (4). Maximum allowed is 2.",
					},
				},
			},

			// Tagged template — function as tag is NOT IIFE; rule fires on the
			// function regardless of IIFEs option.
			{
				Code:    "(function f() {\nlet a = 1;\nlet b = 2;\nreturn a;\n})`tpl`;",
				Options: map[string]interface{}{"max": 2, "IIFEs": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Function 'f' has too many lines (5). Maximum allowed is 2.",
					},
				},
			},

			// `new (function(){})()` — NewExpression's callee is the function
			// but ESLint's isIIFE only matches CallExpression. We do the same.
			// So with IIFEs:false the function is still checked.
			{
				Code: `new (function () {
this.x = 1;
this.y = 2;
this.z = 3;
})();`,
				Options: map[string]interface{}{"max": 2, "IIFEs": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Function has too many lines (5). Maximum allowed is 2.",
					},
				},
			},

			// `(function(){}).bind(this)()` — outermost CallExpression's
			// callee is `.bind(this)`, NOT the function. Inner function should
			// be visited normally even with IIFEs:false.
			{
				Code: `var fn = (function () {
let a = 1;
let b = 2;
let c = 3;
}).bind(this);`,
				Options: map[string]interface{}{"max": 2, "IIFEs": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Function has too many lines (5). Maximum allowed is 2.",
					},
				},
			},

			// React-style class — render method exceeds; sibling methods
			// don't. Verifies independent per-method visiting.
			{
				Code: `class Component {
componentDidMount() { this.fetch(); }
render() {
let a = 1;
let b = 2;
let c = 3;
return null;
}
componentWillUnmount() { this.cleanup(); }
}`,
				Options: map[string]interface{}{"max": 3},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Method 'render' has too many lines (6). Maximum allowed is 3.",
					},
				},
			},

			// Reduce callback NOT IIFE — function as 1st argument
			{
				Code: `arr.reduce(function (acc, v) {
acc.x = v;
acc.y = v;
acc.z = v;
return acc;
}, {});`,
				Options: map[string]interface{}{"max": 2, "IIFEs": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Function has too many lines (6). Maximum allowed is 2.",
					},
				},
			},

			// Multi-paren-wrapped IIFE recognized when IIFEs: true.
			// Same code as a valid case earlier with max=1+IIFEs:false; here
			// max=2+IIFEs:true → the function IS checked and exceeds.
			{
				Code:    "((function () {\nlet x = 0;\nlet y = 0;\nlet z = 0;\n}))();",
				Options: map[string]interface{}{"max": 2, "IIFEs": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Function has too many lines (5). Maximum allowed is 2.",
					},
				},
			},

			// TS class with all modifier permutations stacked: public + static + async
			{
				Code: `class A {
public static async foo() {
await g();
return 1;
}
}`,
				Options: map[string]interface{}{"max": 2},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Static async method 'foo' has too many lines (4). Maximum allowed is 2.",
					},
				},
			},

			// TS parameter properties — constructor still counts normally.
			// "Constructor" is reported (no name modifier).
			{
				Code: `class A {
constructor(public x: number, private y: string) {
this.z = x;
this.w = y;
}
}`,
				Options: map[string]interface{}{"max": 2},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Constructor has too many lines (4). Maximum allowed is 2.",
					},
				},
			},

			// max=0 fires on EVERY function (including 1-line). Trips both
			// the outer arrow and the inner method.
			{
				Code: `var f = () => {
class A { method() { return 1; } }
};`,
				Options: map[string]interface{}{"max": 0},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Arrow function 'f' has too many lines (3). Maximum allowed is 0.",
					},
					{
						MessageId: "exceed",
						Message:   "Method 'method' has too many lines (1). Maximum allowed is 0.",
					},
				},
			},

			// Negative max fires on everything (defensive — ESLint's schema
			// rejects this, but our parser accepts the value, fires on >= 0).
			// Ports of similar guard rails in max-lines.
			{
				Code:    "function f() {\nreturn 1;\n}",
				Options: map[string]interface{}{"max": -1},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Function 'f' has too many lines (3). Maximum allowed is -1.",
					},
				},
			},

			// Function expression in array literal — visited and reported
			// normally.
			{
				Code: `var arr = [function () {
let a = 1;
let b = 2;
}];`,
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Function has too many lines (4). Maximum allowed is 2.",
					},
				},
			},

			// Single-line method spanning the whole class on one line —
			// when max=0 the lineCount=1 still exceeds.
			{
				Code:    "class A { method() { return 1; } }",
				Options: 0,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Method 'method' has too many lines (1). Maximum allowed is 0.",
					},
				},
			},

			// Function with name that is reserved word coerced to identifier
			// in tsgo (parsing as function name)
			{
				Code:    "var x = function async() {\nreturn 1;\nreturn 2;\n};",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Function 'async' has too many lines (4). Maximum allowed is 2.",
					},
				},
			},

			// Position assertions: multi-line arrow with type annotation on
			// param (TS-only). Range starts at `(` (with leading paren), not
			// at the type. Catches accidental inclusion of the type-only
			// position.
			{
				Code:    "const f = (x: number): number => {\nreturn x;\nreturn x + 1;\n};",
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Arrow function 'f' has too many lines (4). Maximum allowed is 2.",
						Line:      1,
						Column:    11,
						EndLine:   4,
						EndColumn: 2,
					},
				},
			},

			// Position assertions: function declaration after JSDoc — the
			// function range starts at `function`, not at the JSDoc.
			{
				Code: `/** docs */
function f() {
return 1;
return 2;
}`,
				Options: 2,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Function 'f' has too many lines (4). Maximum allowed is 2.",
						Line:      2,
						Column:    1,
						EndLine:   5,
						EndColumn: 2,
					},
				},
			},

			// Class expression with private static method (both modifiers):
			// "Static private method '#priv'". Verifies inClassBody works for
			// ClassExpression too, not just ClassDeclaration.
			{
				Code: `var A = class {
static #priv() {
return 1;
return 2;
}
};`,
				Options: map[string]interface{}{"max": 2},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Static private method '#priv' has too many lines (4). Maximum allowed is 2.",
					},
				},
			},

			// Object literal arrow nested in returned object: outer arrow +
			// inner arrow both reportable. (The inner arrow should NOT be
			// treated as IIFE since its parent is PropertyAssignment, not
			// CallExpression.)
			{
				Code: `const Button = () => ({ onClick: () => {
console.log('clicked');
console.log('twice');
console.log('thrice');
} });`,
				Options: map[string]interface{}{"max": 2, "IIFEs": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Arrow function 'Button' has too many lines (5). Maximum allowed is 2.",
					},
					{
						MessageId: "exceed",
						Message:   "Arrow function 'onClick' has too many lines (5). Maximum allowed is 2.",
					},
				},
			},

			// Generator method on object literal — "Generator method 'gen'"
			// (verifies object-literal generator method is identified the
			// same as class generator method).
			{
				Code: `var obj = {
*gen() {
yield 1;
yield 2;
yield 3;
}
};`,
				Options: map[string]interface{}{"max": 2},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Generator method 'gen' has too many lines (5). Maximum allowed is 2.",
					},
				},
			},

			// Trailing-comment line inside function body — under skipComments,
			// the trailing comment line is skipped. With max=2, code-only
			// count = 3 (signature + body line + closing brace) → 3 > 2.
			{
				Code: `function f() {
let x = 1;
// trailing
}`,
				Options: map[string]interface{}{"max": 2, "skipComments": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Function 'f' has too many lines (3). Maximum allowed is 2.",
					},
				},
			},

			// Two IIFEs in the same file, different forms — when IIFEs:true,
			// all three should fire under max=0.
			{
				Code: `(function () { return 1; })();
(() => 2)();
(function () { return 3; }());`,
				Options: map[string]interface{}{"max": 0, "IIFEs": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Function has too many lines (1). Maximum allowed is 0.",
					},
					{
						MessageId: "exceed",
						Message:   "Arrow function has too many lines (1). Maximum allowed is 0.",
					},
					{
						MessageId: "exceed",
						Message:   "Function has too many lines (1). Maximum allowed is 0.",
					},
				},
			},

			// Async arrow IIFE (top-level await pattern). With IIFEs:true
			// it's checked.
			{
				Code: `(async () => {
const x = await f();
const y = await g();
return x + y;
})();`,
				Options: map[string]interface{}{"max": 2, "IIFEs": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "exceed",
						Message:   "Async arrow function has too many lines (5). Maximum allowed is 2.",
					},
				},
			},
		},
	)
}

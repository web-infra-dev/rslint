package prefer_math_min_max_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_math_min_max"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/prefer-math-min-max.js
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/prefer-math-min-max.md
// Diagnostic ranges and suggestions use Unicorn 79.0.0 and ESLint 10.14.0.
func TestPreferMathMinMaxUpstream(t *testing.T) {
	// Repeated TypeScript cases use the Go tester's compiler program here.
	// The JS mirror runs them with and without a project.
	valid := []rule_tester.ValidTestCase{
		// String initializers and parameter defaults
		{FileName: "input.js", Code: `const value = "1"; value > 2 ? 2 : value;`},
		{FileName: "input.js", Code: "const value = `1`; value > 2 ? 2 : value;"},
		{FileName: "input.js", Code: "const value = ``; value < 2 ? value : 2;"},
		{FileName: "input.js", Code: "const value = `\\u0031`; value > 2 ? value : 2;"},
		{FileName: "input.js", Code: `function foo(value = "1") { return value > 2 ? 2 : value; }`},
		{FileName: "input.js", Code: "function foo(value = `1`) { return value > 2 ? 2 : value; }"},
		{FileName: "input.js", Code: "function foo(value = ``) { return value < 2 ? value : 2; }"},

		// Preserved and discarded comments
		{FileName: "input.js", Code: `const value = 1n; (/* keep */ value) > 0 ? value : 0;`},

		// BigInt and number JSDoc in JavaScript
		{FileName: "input.js", Code: `/** @param {bigint} input */ function run(input) { const value = input * 2n; return value > 0 ? value : 0; }`},
		{FileName: "input.js", Code: `/** @type {bigint} */ let value; value > 0 ? 0 : value;`},
		{FileName: "input.js", Code: `/** @param {bigint} value */ function run(value) { return value > 0 ? value : 0; }`},
		{FileName: "input.js", Code: `/** @param {number | bigint} value */ function run(value, limit) { return value < limit ? limit : value; }`},
		{FileName: "input.js", Code: `/** @param {?bigint} value */ function run(value) { return value <= 0 ? value : 0; }`},
		{FileName: "input.js", Code: `/** @param {bigint} [value] */ function run(value) { return value >= 0 ? 0 : value; }`},
		{FileName: "input.js", Code: `/** @param {number | bigint} limit */ function run(value, limit) { return limit > value ? limit : value; }`},
		{FileName: "input.js", Code: `/** @returns {number | bigint} */ function read() { return source; } let value = read(); value > 0 ? value : 0;`},
		{FileName: "input.js", Code: `/** @returns {number | bigint} */ function read() { return source; } function run(value = read()) { return value > 0 ? value : 0; }`},
		{FileName: "input.js", Code: `/** @param {number | bigint} input */ function run(input) { const value = -input; return value < 0 ? value : 0; }`},

		// JSDoc on TypeScript parameters
		{FileName: "input.ts", Code: `/** @param {bigint} value */ function run(value) { return value > 0 ? value : 0; }`},
		{FileName: "input.ts", Code: `/** @param {number | bigint} limit */ function run(value, limit) { return value < limit ? value : limit; }`},

		// JSDoc with the JavaScript parser
		{FileName: "input.js", Code: `/** @returns {bigint} */ function read() { return source; } const value = read(); value > 0 ? value : 0;`},
		{FileName: "input.js", Code: `/** @param {() => bigint} read */ function run(read) { const value = read(); return value > 0 ? value : 0; }`},
		{FileName: "input.js", Code: `/** @type {bigint} */ const value = source; const alias = value; alias < 0 ? alias : 0;`},
		{FileName: "input.js", Code: `/** @returns {bigint} */ function read() { return source; } function run(value = read()) { return value > 0 ? value : 0; }`},
		{FileName: "input.js", Code: `/** @param {bigint} first
 * @param {bigint} second */
function run(first, second) { const value = condition ? first : second; return value > 0 ? value : 0; }`},

		// JSDoc in TypeScript syntax mode
		{FileName: "input.ts", Code: `/** @returns {bigint} */ function read() { return source; } const value = read(); value > 0 ? value : 0;`},
		{FileName: "input.ts", Code: `/** @param {() => bigint} read */ function run(read) { const value = read(); return value > 0 ? value : 0; }`},
		{FileName: "input.ts", Code: `/** @type {bigint} */ const value = source; const alias = value; alias < 0 ? alias : 0;`},
		{FileName: "input.ts", Code: `/** @returns {bigint} */ function read() { return source; } function run(value = read()) { return value > 0 ? value : 0; }`},
		{FileName: "input.ts", Code: `/** @param {bigint} first
 * @param {bigint} second */
function run(first, second) { const value = condition ? first : second; return value > 0 ? value : 0; }`},

		// Native return type precedence in TypeScript syntax mode
		{FileName: "input.ts", Code: `/** @returns {number} */ function read(): bigint { return source; } const value = read(); value > 0 ? value : 0;`},

		// JavaScript snapshots
		{FileName: "input.js", Code: `height > 10 ? height : 20`},
		{FileName: "input.js", Code: `height > 50 ? Math.min(50, height) : height`},
		{FileName: "input.js", Code: `foo ? foo : bar`},
		{FileName: "input.js", Code: `foo > 10n ? 10n : foo`},
		{FileName: "input.js", Code: `foo > BigInt(10) ? BigInt(10) : foo`},
		{FileName: "input.js", Code: `new Date() > foo ? foo : new Date()`},
		{FileName: "input.js", Code: `foo > new Date(0) ? new Date(0) : foo`},
		{FileName: "input.js", Code: `var now = new Date();
var later = new Date();
var value = now > later ? later : now;`},
		{FileName: "input.js", Code: `const start = new Date();
var value = start > foo ? foo : start;`},
		{FileName: "input.js", Code: `const start = new Date();
function foo() {
	return start > foo ? foo : start;
}`},
		{FileName: "input.js", Code: `const start = 'text';
function foo() {
	return start > foo ? foo : start;
}`},
		{FileName: "input.js", Code: `const a = BigInt(1);
const value = a > 0 ? 0 : a;`},
		{FileName: "input.js", Code: `const a = -1n;
const value = a > 0 ? 0 : a;`},
		{FileName: "input.js", Code: `function foo(a = BigInt(1)) {
	return a > 0 ? 0 : a;
}`},
		{FileName: "input.js", Code: `function foo(a = 'string', b) {
  return a > b ? a : b;
}`},

		// TypeScript snapshots
		{FileName: "input.ts", Code: `function foo(a, b) {
	return (a as bigint) > b ? a : b;
}`},
		{FileName: "input.ts", Code: `function foo(a, b) {
	return (a as string) > b ? a : b;
}`},
		{FileName: "input.ts", Code: `function foo(a, b) {
	return (<bigint>a) > b ? a : b;
}`},
		{FileName: "input.ts", Code: `function foo(a, b) {
	return (<string>a) > b ? a : b;
}`},
		{FileName: "input.ts", Code: `function foo(a: string, b) {
	return a > b ? a : b;
}`},
		{FileName: "input.ts", Code: `function foo(a, b: string) {
	return a > b ? a : b;
}`},
		{FileName: "input.ts", Code: `function foo(a: bigint, b: bigint) {
	return a > b ? a : b;
}`},
		{FileName: "input.ts", Code: `function foo(a: Date, b: Date) {
	return a > b ? a : b;
}`},
		{FileName: "input.ts", Code: `var foo: Date;
var bar: Date;
var value = foo > bar ? bar : foo;`},
		{FileName: "input.ts", Code: `var foo = 10;
var bar = '20';

var value = foo > bar ? bar : foo;`},
		{FileName: "input.ts", Code: `var foo = 10;
var bar: string;

var value = foo > bar ? bar : foo;`},
		{FileName: "input.ts", Code: `function foo(a, b) {
	return (a as string)! > b ? (a as string)! : b;
}`},
		{FileName: "input.ts", Code: `function foo(a, b) {
	return (<string>a)! > b ? (<string>a)! : b;
}`},
		{FileName: "input.ts", Code: `function foo(a, b) {
	return (a as Foo) > b ? a : b;
}`},

		// Documentation examples
		{FileName: "input.js", Code: `Math.min(height, 50);`},
		{FileName: "input.js", Code: `Math.max(height, 50);`},

		// Runtime regression inputs
		{FileName: "input.js", Code: `/** @param {bigint} value */ function run(value) { return value < 0 ? value : 0; } run(-5n);`},
		{FileName: "input.js", Code: `/** @param {bigint} value */ function run(value) { return value > 0 ? value : 0; } run(5n);`},
	}
	invalid := []rule_tester.InvalidTestCase{
		// Comments inside operand parentheses
		minMaxCase("input.js", `(/* keep */ height) > 50 ? 50 : height;`, "min", 1, 1, 1, 39, `Math.min((/* keep */ height), 50);`),
		minMaxCase("input.js", `height < (50 /* keep */) ? 50 : height;`, "max", 1, 1, 1, 39, `Math.max(height, (50 /* keep */));`),

		// Preserved and discarded comments
		minMaxCase("input.js", `((height /* keep */)) >= 50 ? 50 : height;`, "min", 1, 1, 1, 42, `Math.min(((height /* keep */)), 50);`),
		minMaxCase("input.js", `(/* left */ height) > (50 /* right */) ? height : 50;`, "max", 1, 1, 1, 53, `Math.max((/* left */ height), (50 /* right */));`),
		minMaxCase("input.js", "height < (50 // keep\r\n) ? height : 50;", "min", 1, 1, 2, 16, "Math.min(height, (50 // keep\r\n));"),
		minMaxCase("input.js", `(/* keep */ first, height) < 50 ? 50 : (first, height);`, "max", 1, 1, 1, 55, `Math.max((/* keep */ first, height), 50);`),
		minMaxCase("input.ts", `(height as /* keep */ number) > 50 ? 50 : height;`, "min", 1, 1, 1, 49, `Math.min(height as /* keep */ number, 50);`),
		minMaxCase("input.js", `(height /* discard */ > 50) ? 50 : height;`, "min", 1, 1, 1, 42, ""),
		minMaxCase("input.js", `(/* discard */ height > 50) ? 50 : height;`, "min", 1, 1, 1, 42, ""),
		minMaxCase("input.js", `height > 50 ? 50 : (/* discard */ height);`, "min", 1, 1, 1, 42, ""),
		minMaxCase("input.js", `(/* keep */ read()) > 50 ? 50 : read();`, "min", 1, 1, 1, 39, ""),

		// BigInt and number JSDoc in JavaScript
		minMaxCase("input.js", `/** @param {number} value */ function run(value) { return value > 0 ? 0 : value; }`, "min", 1, 59, 1, 80, `/** @param {number} value */ function run(value) { return Math.min(value, 0); }`),
		minMaxCase("input.js", `/** @param {number} value */ function run(value) { return value > 0 ? value : 0; }`, "max", 1, 59, 1, 80, `/** @param {number} value */ function run(value) { return Math.max(value, 0); }`),

		// JSDoc on TypeScript parameters
		minMaxCase("input.ts", `/** @param {bigint} value */ function run(value: number) { return value > 0 ? value : 0; }`, "max", 1, 67, 1, 88, `/** @param {bigint} value */ function run(value: number) { return Math.max(value, 0); }`),

		// JSDoc with the JavaScript parser
		minMaxCase("input.js", `/** @returns {number} */ function read() { return source; } const value = read(); value > 0 ? value : 0;`, "max", 1, 83, 1, 104, `/** @returns {number} */ function read() { return source; } const value = read(); Math.max(value, 0);`),

		// JSDoc in TypeScript syntax mode
		minMaxCase("input.ts", `/** @returns {number} */ function read() { return source; } const value = read(); value > 0 ? value : 0;`, "max", 1, 83, 1, 104, `/** @returns {number} */ function read() { return source; } const value = read(); Math.max(value, 0);`),

		// Native return type precedence in TypeScript syntax mode
		minMaxCase("input.ts", `/** @returns {bigint} */ function read(): number { return source; } const value = read(); value > 0 ? value : 0;`, "max", 1, 91, 1, 112, `/** @returns {bigint} */ function read(): number { return source; } const value = read(); Math.max(value, 0);`),

		// NaN and signed zero
		minMaxCase("input.js", `const value = NaN; value > 0 ? value : 0;`, "max", 1, 20, 1, 41, `const value = NaN; Math.max(value, 0);`),
		minMaxCase("input.js", `const value = -0; value < 0 ? value : 0;`, "min", 1, 19, 1, 40, `const value = -0; Math.min(value, 0);`),

		// JavaScript snapshots
		minMaxCase("input.js", `height > 50 ? 50 : height`, "min", 1, 1, 1, 26, `Math.min(height, 50)`),
		minMaxCase("input.js", `height >= 50 ? 50 : height`, "min", 1, 1, 1, 27, `Math.min(height, 50)`),
		minMaxCase("input.js", `height < 50 ? height : 50`, "min", 1, 1, 1, 26, `Math.min(height, 50)`),
		minMaxCase("input.js", `height <= 50 ? height : 50`, "min", 1, 1, 1, 27, `Math.min(height, 50)`),
		minMaxCase("input.js", `height > maxHeight ? maxHeight : height`, "min", 1, 1, 1, 40, `Math.min(height, maxHeight)`),
		minMaxCase("input.js", `height < maxHeight ? height : maxHeight`, "min", 1, 1, 1, 40, `Math.min(height, maxHeight)`),
		minMaxCase("input.js", `window.height > 50 ? 50 : window.height`, "min", 1, 1, 1, 40, `Math.min(window.height, 50)`),
		minMaxCase("input.js", `window.height < 50 ? window.height : 50`, "min", 1, 1, 1, 40, `Math.min(window.height, 50)`),
		minMaxCase("input.js", `height > 50 ? height : 50`, "max", 1, 1, 1, 26, `Math.max(height, 50)`),
		minMaxCase("input.js", `height >= 50 ? height : 50`, "max", 1, 1, 1, 27, `Math.max(height, 50)`),
		minMaxCase("input.js", `height < 50 ? 50 : height`, "max", 1, 1, 1, 26, `Math.max(height, 50)`),
		minMaxCase("input.js", `height <= 50 ? 50 : height`, "max", 1, 1, 1, 27, `Math.max(height, 50)`),
		minMaxCase("input.js", `height > maxHeight ? height : maxHeight`, "max", 1, 1, 1, 40, `Math.max(height, maxHeight)`),
		minMaxCase("input.js", `height < maxHeight ? maxHeight : height`, "max", 1, 1, 1, 40, `Math.max(height, maxHeight)`),
		minMaxCase("input.js", `function a() {
	return +foo > 10 ? 10 : +foo
}`, "min", 2, 9, 2, 30, `function a() {
	return Math.min(+foo, 10)
}`),
		minMaxCase("input.js", `function a() {
	return+foo > 10 ? 10 : +foo
}`, "min", 2, 8, 2, 29, `function a() {
	return Math.min(+foo, 10)
}`),
		minMaxCase("input.js", `(0,foo) > 10 ? 10 : (0,foo)`, "min", 1, 1, 1, 28, `Math.min((0,foo), 10)`),
		minMaxCase("input.js", `foo > (0,bar) ? foo : (0,bar)`, "max", 1, 1, 1, 30, `Math.max(foo, (0,bar))`),
		minMaxCase("input.js", `Number(x) > 50 ? 50 : Number(x)`, "min", 1, 1, 1, 32, ""),
		minMaxCase("input.js", `foo.bar() > 10 ? 10 : foo.bar()`, "min", 1, 1, 1, 32, ""),
		minMaxCase("input.js", `async function foo() {
	return await foo.bar() > 10 ? 10 : await foo.bar()
}`, "min", 2, 9, 2, 52, ""),
		minMaxCase("input.js", `async function foo() {
	await(+foo > 10 ? 10 : +foo)
}`, "min", 2, 8, 2, 29, `async function foo() {
	await (Math.min(+foo, 10))
}`),
		minMaxCase("input.js", `function foo() {
	return(foo.bar() > 10) ? 10 : foo.bar()
}`, "min", 2, 8, 2, 41, ""),
		minMaxCase("input.js", `function* foo() {
	yield+foo > 10 ? 10 : +foo
}`, "min", 2, 7, 2, 28, `function* foo() {
	yield Math.min(+foo, 10)
}`),
		minMaxCase("input.js", `export default+foo > 10 ? 10 : +foo`, "min", 1, 15, 1, 36, `export default Math.min(+foo, 10)`),
		minMaxCase("input.js", `foo.length > bar.length ? bar.length : foo.length`, "min", 1, 1, 1, 50, `Math.min(foo.length, bar.length)`),

		// Comments outside comparison operands
		minMaxCase("input.js", `const value = height > // keep
	50 ? 50 : height;`, "min", 1, 15, 2, 18, ""),
		minMaxCase("input.js", `const value = height > /* keep */ 50 ? 50 : height;`, "min", 1, 15, 1, 51, ""),
		minMaxCase("input.js", `const value = height > 50 /* keep */ ? 50 : height;`, "min", 1, 15, 1, 51, ""),
		minMaxCase("input.js", `const value = height > 50 ? 50 /* keep */ : height;`, "min", 1, 15, 1, 51, ""),
		minMaxCase("input.js", `const value = height > 50 ? /* keep */ 50 : height;`, "min", 1, 15, 1, 51, ""),
		minMaxCase("input.js", `const value = height > 50 ? 50 : height;`, "min", 1, 15, 1, 40, `const value = Math.min(height, 50);`),

		// TypeScript snapshots
		minMaxCase("input.ts", `function foo(a, b) {
	return (a as number) > b ? a : b;
}`, "max", 2, 9, 2, 34, `function foo(a, b) {
	return Math.max(a as number, b);
}`),
		minMaxCase("input.ts", `function foo(a, b) {
	return (a as number) > b ? a : b;
}`, "max", 2, 9, 2, 34, `function foo(a, b) {
	return Math.max(a as number, b);
}`),
		minMaxCase("input.ts", `function foo(a, b) {
	return (a as unknown as number) > b ? a : b;
}`, "max", 2, 9, 2, 45, `function foo(a, b) {
	return Math.max(a as unknown as number, b);
}`),
		minMaxCase("input.ts", `var foo = 10;

var value = foo > bar ? bar : foo;`, "min", 3, 13, 3, 34, `var foo = 10;

var value = Math.min(foo, bar);`),
		minMaxCase("input.ts", `var foo = 10;
var bar = 20;

var value = foo > bar ? bar : foo;`, "min", 4, 13, 4, 34, `var foo = 10;
var bar = 20;

var value = Math.min(foo, bar);`),
		minMaxCase("input.ts", `var foo: number;
var bar: number;

var value = foo > bar ? bar : foo;`, "min", 4, 13, 4, 34, `var foo: number;
var bar: number;

var value = Math.min(foo, bar);`),
		minMaxCase("input.ts", `function foo(a, b) {
	return (a as number) > b ? (a as number) : b;
}`, "max", 2, 9, 2, 46, `function foo(a, b) {
	return Math.max(a as number, b);
}`),
		minMaxCase("input.ts", `function foo(a, b) {
	return (<number>a) > b ? (<number>a) : b;
}`, "max", 2, 9, 2, 42, `function foo(a, b) {
	return Math.max(<number>a, b);
}`),
		minMaxCase("input.ts", `function foo(a, b) {
	return a! > b ? a! : b;
}`, "max", 2, 9, 2, 24, `function foo(a, b) {
	return Math.max(a!, b);
}`),
		minMaxCase("input.ts", `function foo(a, b) {
	return (a as Number) > b ? a : b;
}`, "max", 2, 9, 2, 34, `function foo(a, b) {
	return Math.max(a as Number, b);
}`),

		// Repeated evaluation of operands
		minMaxCase("input.js", `const a = f() > 0 ? f() : 0;`, "max", 1, 11, 1, 28, ""),
		minMaxCase("input.js", `const a = (x++, y) > 0 ? 0 : (x++, y);`, "min", 1, 11, 1, 38, ""),
		minMaxCase("input.js", `const a = (x = 1, y) > 0 ? 0 : (x = 1, y);`, "min", 1, 11, 1, 42, ""),
		minMaxCase("input.js", `const a = (x || f()) > 0 ? 0 : (x || f());`, "min", 1, 11, 1, 42, ""),
		minMaxCase("input.js", `const a = (x ? f() : g()) > 0 ? 0 : (x ? f() : g());`, "min", 1, 11, 1, 52, ""),
		minMaxCase("input.js", `const a = [f()] > 0 ? 0 : [f()];`, "min", 1, 11, 1, 32, ""),
		minMaxCase("input.js", `const a = ({a: f()}) > 0 ? 0 : ({a: f()});`, "min", 1, 11, 1, 42, ""),
		minMaxCase("input.js", `const a = (x, y) > 0 ? 0 : (x, y);`, "min", 1, 11, 1, 34, `const a = Math.min((x, y), 0);`),
		minMaxCase("input.js", "const a = `${x}` > 0 ? 0 : `${x}`;", "min", 1, 11, 1, 34, ""),
		minMaxCase("input.js", "const a = `1` > 0 ? 0 : `1`;", "min", 1, 11, 1, 28, "const a = Math.min(`1`, 0);"),
		minMaxCase("input.js", `const a = [() => f()] > 0 ? 0 : [() => f()];`, "min", 1, 11, 1, 44, `const a = Math.min([() => f()], 0);`),
		minMaxCase("input.js", `import foo from "foo"; const a = foo > 0 ? 0 : foo;`, "min", 1, 34, 1, 51, `import foo from "foo"; const a = Math.min(foo, 0);`),

		// Documentation examples
		minMaxCase("input.js", `height > 50 ? 50 : height;`, "min", 1, 1, 1, 26, `Math.min(height, 50);`),
		minMaxCase("input.js", `height > 50 ? height : 50;`, "max", 1, 1, 1, 26, `Math.max(height, 50);`),
		minMaxCase("input.js", `height >= 50 ? 50 : height;`, "min", 1, 1, 1, 27, `Math.min(height, 50);`),
		minMaxCase("input.js", `height >= 50 ? height : 50;`, "max", 1, 1, 1, 27, `Math.max(height, 50);`),
		minMaxCase("input.js", `height < 50 ? height : 50;`, "min", 1, 1, 1, 26, `Math.min(height, 50);`),
		minMaxCase("input.js", `height < 50 ? 50 : height;`, "max", 1, 1, 1, 26, `Math.max(height, 50);`),
		minMaxCase("input.js", `height <= 50 ? height : 50;`, "min", 1, 1, 1, 27, `Math.min(height, 50);`),
		minMaxCase("input.js", `height <= 50 ? 50 : height;`, "max", 1, 1, 1, 27, `Math.max(height, 50);`),

		// Runtime regression inputs
		minMaxCase("input.js", `(/* keep */ value) > limit ? value : limit;`, "max", 1, 1, 1, 43, `Math.max((/* keep */ value), limit);`),
		minMaxCase("input.js", `(/* keep */ value) < limit ? value : limit;`, "min", 1, 1, 1, 43, `Math.min((/* keep */ value), limit);`),
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_math_min_max.PreferMathMinMaxRule, valid, invalid)
}

func minMaxCase(filename, code, method string, line, column, endLine, endColumn int, suggestion string) rule_tester.InvalidTestCase {
	diagnostic := rule_tester.InvalidTestCaseError{
		MessageId: "prefer-math-min-max",
		Message:   "Prefer `Math." + method + "()` to simplify ternary expressions.",
		Line:      line, Column: column, EndLine: endLine, EndColumn: endColumn,
	}
	if suggestion != "" {
		diagnostic.Suggestions = []rule_tester.InvalidTestCaseSuggestion{{
			MessageId: "replace-with-math-min-max",
			Output:    suggestion,
		}}
	}
	return rule_tester.InvalidTestCase{FileName: filename, Code: code, Errors: []rule_tester.InvalidTestCaseError{diagnostic}}
}

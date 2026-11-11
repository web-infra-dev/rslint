package no_return_array_push_test

// Ported from eslint-plugin-unicorn v79.0.0. See LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/no-return-array-push.js

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_return_array_push"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoReturnArrayPushUpstream(t *testing.T) {
	run := func(t *testing.T, valid []rule_tester.ValidTestCase, invalid []rule_tester.InvalidTestCase) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_return_array_push.NoReturnArrayPushRule, valid, invalid)
	}
	// The three upstream parsers give the same results for these JSDoc inputs.
	t.Run("JSDoc annotations", func(t *testing.T) {
		run(t, validCases("file.ts",
			`/** @param {Map<string, number>} values */ function run(values) { return values.push(value); }`,
			`/** @returns {Set<number>} */ function read() { return source; } function run() { return read().unshift(value); }`,
			`/** @param {() => string} read */ function run(read) { return read().push(value); }`,
			`/** @param {?Array<number>} stream */ function run(stream) { return stream.push(value); }`,
			`/** @param {Array<number>} values */ function run(values) { return values.push(value).then(consume); }`,
			`/** @param {Array<number>} values */ function run(values) { return void values.push(value); }`,
		), []rule_tester.InvalidTestCase{
			invalidCase("file.ts", `/** @param {Array<number>} stream */ function run(stream) { return stream.push(value); }`, "push", 1, 75, `/** @param {Array<number>} stream */ function run(stream) { stream.push(value); return; }`),
			invalidCase("file.ts", `/** @returns {ReadonlyArray<number>} */ function read() { return source; } function run() { return read().unshift(value); }`, "unshift", 1, 107, `/** @returns {ReadonlyArray<number>} */ function read() { return source; } function run() { read().unshift(value); return; }`),
			invalidCase("file.ts", `/** @param {?Array<number>} values */ function run(values) { return values.push(value); }`, "push", 1, 76, `/** @param {?Array<number>} values */ function run(values) { values.push(value); return; }`),
			invalidCase("file.ts", `/** @param {Uint8Array} values */ function run(values) { return values.push(value); }`, "push", 1, 72, `/** @param {Uint8Array} values */ function run(values) { values.push(value); return; }`),
			invalidCase("file.ts", `/** @param {Array<number>} values */ const run = values => values.push(value);`, "push", 1, 67, ""),
			invalidCase("file.ts", `/** @param {Array<number>} values */ function run(values) { return /* keep */ values.unshift(value); }`, "unshift", 1, 86, ""),
		})
	})
	// The syntax-only and type-aware TypeScript parsers agree here.
	t.Run("TypeScript annotations take precedence over JSDoc", func(t *testing.T) {
		run(t, validCases("file.ts",
			`/** @param {Array<number>} values */ function run(values: Set<number>) { return values.push(value); }`,
		), []rule_tester.InvalidTestCase{
			invalidCase("file.ts", `/** @param {Map<string, number>} stream */ function run(stream: number[]) { return stream.push(value); }`, "push", 1, 91, `/** @param {Map<string, number>} stream */ function run(stream: number[]) { stream.push(value); return; }`),
		})
	})
	t.Run("JavaScript", func(t *testing.T) {
		run(t, validCases("file.js",
			`function foo() {
	array.push(value);
	return;
}`,
			`function foo() {
	array.unshift(value);
	return;
}`,
			`function foo() {
	return array.length + 1;
}`,
			`function foo() {
	return array.push();
}`,
			`function foo() {
	return array.unshift();
}`,
			`function foo() {
	return stream.push(chunk);
}`,
			`function foo() {
	return this.push(chunk);
}`,
			`function foo() {
	return this.stream.push(chunk);
}`,
			`function foo() {
	return process.stdin.push(chunk);
}`,
			`function foo() {
	return process.stdout.push(chunk);
}`,
			`function foo() {
	return process.stderr.push(chunk);
}`,
			`function foo() {
	return router.push(to);
}`,
			`function foo() {
	return this.router.push(to);
}`,
			`function foo() {
	return this.$router.push(to);
}`,
			`function foo() {
	return stream?.push(chunk);
}`,
			`function foo() {
	return this?.push(chunk);
}`,
			`function foo() {
	return process.stdin?.push(chunk);
}`,
			`function foo() {
	return array['push'](value);
}`,
			`function foo() {
	return array['unshift'](value);
}`,
			`function foo() {
	return array[push](value);
}`,
			`function foo() {
	return array[unshift](value);
}`,
			`function foo() {
	return push(value);
}`,
			`function foo() {
	return unshift(value);
}`,
			`function foo() {
	array?.push(value);
}`,
			`function foo() {
	array?.unshift(value);
}`,
			`void array.push(value);`,
			`void array.unshift(value);`,
			`void array?.push(value);`,
			`void array?.unshift(value);`,
			`const length = void array.push(value);`,
			`const foo = value => void array.push(value);`,
			`function foo() {
	return void array.push(value);
}`,
			`function foo() {
	return router.push(to).catch(() => {});
}`,
			`function foo() {
	return router.push(to).then(onFulfilled);
}`,
			`const promise = router.push(to).finally(cleanup);`,
			`function foo() {
	return array.push(value).toString();
}`,
			`function foo() {
	return array.unshift(value).foo;
}`,
			`function foo() {
	return router.push(to)?.catch(() => {});
}`,
			`array.push(value)[0];`,
		), []rule_tester.InvalidTestCase{
			invalidCase("file.js", `const length = array.push(value);`, "push", 1, 22, ""),
			invalidCase("file.js", `const length = routerItems.push(value);`, "push", 1, 28, ""),
			invalidCase("file.js", `const length = array.unshift(value);`, "unshift", 1, 22, ""),
			invalidCase("file.js", `const length = stream.unshift(chunk);`, "unshift", 1, 23, ""),
			invalidCase("file.js", `console.log(array.push(value));`, "push", 1, 19, ""),
			invalidCase("file.js", `console.log(array.unshift(value));`, "unshift", 1, 19, ""),
			invalidCase("file.js", `const length = await array.push(value);`, "push", 1, 28, ""),
			invalidCase("file.js", `const length = await array.unshift(value);`, "unshift", 1, 28, ""),
			invalidCase("file.js", `async function foo() { await array.push(value); }`, "push", 1, 36, ""),
			invalidCase("file.js", `async function foo() { await array.unshift(value); }`, "unshift", 1, 36, ""),
			invalidCase("file.js", `(array.push(value), sideEffect());`, "push", 1, 8, ""),
			invalidCase("file.js", `(array.unshift(value), sideEffect());`, "unshift", 1, 8, ""),
			invalidCase("file.js", `(sideEffect(), array.push(value));`, "push", 1, 22, ""),
			invalidCase("file.js", `(sideEffect(), array.unshift(value));`, "unshift", 1, 22, ""),
			invalidCase("file.js", `condition && array.push(value);`, "push", 1, 20, ""),
			invalidCase("file.js", `condition && array.unshift(value);`, "unshift", 1, 20, ""),
			invalidCase("file.js", `array.push(value) && sideEffect();`, "push", 1, 7, ""),
			invalidCase("file.js", `array.unshift(value) && sideEffect();`, "unshift", 1, 7, ""),
			invalidCase("file.js", `condition || array.push(value);`, "push", 1, 20, ""),
			invalidCase("file.js", `condition || array.unshift(value);`, "unshift", 1, 20, ""),
			invalidCase("file.js", `array.push(value) || sideEffect();`, "push", 1, 7, ""),
			invalidCase("file.js", `array.unshift(value) || sideEffect();`, "unshift", 1, 7, ""),
			invalidCase("file.js", `condition ?? array.push(value);`, "push", 1, 20, ""),
			invalidCase("file.js", `condition ?? array.unshift(value);`, "unshift", 1, 20, ""),
			invalidCase("file.js", `array.push(value) ?? sideEffect();`, "push", 1, 7, ""),
			invalidCase("file.js", `array.unshift(value) ?? sideEffect();`, "unshift", 1, 7, ""),
			invalidCase("file.js", `condition ? array.push(value) : sideEffect();`, "push", 1, 19, ""),
			invalidCase("file.js", `condition ? array.unshift(value) : sideEffect();`, "unshift", 1, 19, ""),
			invalidCase("file.js", `condition ? sideEffect() : array.push(value);`, "push", 1, 34, ""),
			invalidCase("file.js", `condition ? sideEffect() : array.unshift(value);`, "unshift", 1, 34, ""),
			invalidCase("file.js", `array.push(value) ? sideEffect() : other();`, "push", 1, 7, ""),
			invalidCase("file.js", `array.unshift(value) ? sideEffect() : other();`, "unshift", 1, 7, ""),
			invalidCase("file.js", `for (array.push(value); condition; ) {}`, "push", 1, 12, ""),
			invalidCase("file.js", `for (; array.push(value); ) {}`, "push", 1, 14, ""),
			invalidCase("file.js", `for (; condition; array.unshift(value)) {}`, "unshift", 1, 25, ""),
			invalidCase("file.js", `for (; array.unshift(value); ) {}`, "unshift", 1, 14, ""),
			invalidCase("file.js", `function foo() {
	return array.push(value);
}`, "push", 2, 15, `function foo() {
	array.push(value); return;
}`),
			invalidCase("file.js", `function foo() {
	return array.unshift(value);
}`, "unshift", 2, 15, `function foo() {
	array.unshift(value); return;
}`),
			invalidCase("file.js", `function foo() {
	return (array.push(value));
}`, "push", 2, 16, `function foo() {
	array.push(value); return;
}`),
			invalidCase("file.js", `function foo() {
	return (array.unshift(value));
}`, "unshift", 2, 16, `function foo() {
	array.unshift(value); return;
}`),
			invalidCase("file.js", `function foo() {
	return array?.push(value);
}`, "push", 2, 16, ""),
			invalidCase("file.js", `function foo() {
	return array?.unshift(value);
}`, "unshift", 2, 16, ""),
			invalidCase("file.js", `function foo() {
	return array.push?.(value);
}`, "push", 2, 15, ""),
			invalidCase("file.js", `function foo() {
	return array.unshift?.(value);
}`, "unshift", 2, 15, ""),
			invalidCase("file.js", `async function foo() {
	return await array.push(value);
}`, "push", 2, 21, ""),
			invalidCase("file.js", `async function foo() {
	return await array.unshift(value);
}`, "unshift", 2, 21, ""),
			invalidCase("file.js", `const foo = value => array.push(value);`, "push", 1, 28, ""),
			invalidCase("file.js", `const foo = value => array.unshift(value);`, "unshift", 1, 28, ""),
			invalidCase("file.js", `const foo = async value => array.push(await value);`, "push", 1, 34, ""),
			invalidCase("file.js", `const foo = async value => array.unshift(await value);`, "unshift", 1, 34, ""),
			invalidCase("file.js", `function foo() {
	return /* comment */ array.push(value);
}`, "push", 2, 29, ""),
			invalidCase("file.js", `function foo() {
	return /* comment */ array.unshift(value);
}`, "unshift", 2, 29, ""),
			invalidCase("file.js", `function foo() {
	return array.push(/* comment */ value);
}`, "push", 2, 15, `function foo() {
	array.push(/* comment */ value); return;
}`),
			invalidCase("file.js", `function foo() {
	return array.unshift(/* comment */ value);
}`, "unshift", 2, 15, `function foo() {
	array.unshift(/* comment */ value); return;
}`),
			invalidCase("file.js", `function foo() {
	if (condition) return array.push(value);
}`, "push", 2, 30, ""),
			invalidCase("file.js", `function foo() {
	if (condition) return array.unshift(value);
}`, "unshift", 2, 30, ""),
			invalidCase("file.js", `function foo() {
	return condition && array.push(value);
}`, "push", 2, 28, ""),
			invalidCase("file.js", `function foo() {
	return condition && array.unshift(value);
}`, "unshift", 2, 28, ""),
			invalidCase("file.js", `function foo() {
	return condition ? array.push(value) : value;
}`, "push", 2, 27, ""),
			invalidCase("file.js", `function foo() {
	return condition ? array.unshift(value) : value;
}`, "unshift", 2, 27, ""),
			invalidCase("file.js", `function foo() {
	return condition ? value : array.push(value);
}`, "push", 2, 35, ""),
			invalidCase("file.js", `function foo() {
	return condition ? value : array.unshift(value);
}`, "unshift", 2, 35, ""),
			invalidCase("file.js", `function foo() {
	return (sideEffect(), array.push(value));
}`, "push", 2, 30, ""),
			invalidCase("file.js", `function foo() {
	return (sideEffect(), array.unshift(value));
}`, "unshift", 2, 30, ""),
			invalidCase("file.js", `function foo() {
	foo()
	return [array].push(value);
}`, "push", 3, 17, `function foo() {
	foo()
	;[array].push(value); return;
}`),
			invalidCase("file.js", `function foo() {
	foo()
	return [array].unshift(value);
}`, "unshift", 3, 17, `function foo() {
	foo()
	;[array].unshift(value); return;
}`),
		})
	})
	t.Run("TypeScript syntax", func(t *testing.T) {
		run(t, validCases("file.ts",
			`array.push(value) as number;`,
			`array.unshift(value) as number;`,
			`<number>array.push(value);`,
			`<number>array.unshift(value);`,
			`array.push(value)!;`,
			`array.unshift(value)!;`,
			`array.push(value) satisfies number;`,
			`array.unshift(value) satisfies number;`,
			`void (array.push(value) as number);`,
			`void (array.unshift(value) as number);`,
			"const onClick = () => router.push(`/${region}/customers/${String(result.data)}` as Route);",
			`function foo() {
	return (array.push(value) as Foo).bar;
}`,
			`interface Router {
	push(to: string): Promise<void>;
}

function foo(router: Router) {
	return router.push(to);
}`,
			`class Queue {
	unshift(value: unknown) {
		return value;
	}
}

function foo(queue: Queue) {
	return queue.unshift(value);
}`,
		), []rule_tester.InvalidTestCase{
			invalidCase("file.ts", `const foo = (value: string) => array.push(value);`, "push", 1, 38, ""),
			invalidCase("file.ts", `const foo = (value: string) => array.unshift(value);`, "unshift", 1, 38, ""),
			invalidCase("file.ts", `function foo() {
	return array.push(value) as number;
}`, "push", 2, 15, ""),
			invalidCase("file.ts", `function foo() {
	return array.unshift(value) as number;
}`, "unshift", 2, 15, ""),
			invalidCase("file.ts", `function foo() {
	return <number>array.push(value);
}`, "push", 2, 23, ""),
			invalidCase("file.ts", `function foo() {
	return <number>array.unshift(value);
}`, "unshift", 2, 23, ""),
			invalidCase("file.ts", `function foo() {
	return array.push(value)!;
}`, "push", 2, 15, ""),
			invalidCase("file.ts", `function foo() {
	return array.unshift(value)!;
}`, "unshift", 2, 15, ""),
			invalidCase("file.ts", `function foo() {
	return array.push(value) satisfies number;
}`, "push", 2, 15, ""),
			invalidCase("file.ts", `function foo() {
	return array.unshift(value) satisfies number;
}`, "unshift", 2, 15, ""),
			invalidCase("file.ts", `const foo = (value: string) => array.push(value) satisfies number;`, "push", 1, 38, ""),
			invalidCase("file.ts", `const foo = (value: string) => array.unshift(value) satisfies number;`, "unshift", 1, 38, ""),
		})
	})
	t.Run("Type information", func(t *testing.T) {
		run(t, validCases("file.ts",
			`declare const router: {push(to: string): Promise<void>}; function foo() { return router.push(to); }`,
			`function foo(navigation: {push(to: string): Promise<void>}) { return navigation.push(to); }`,
			`declare const queue: {unshift(value: unknown): number}; function foo() { return queue.unshift(value); }`,
		), []rule_tester.InvalidTestCase{
			invalidCase("file.ts", `declare const array: number[]; function foo() { return array.push(value); }`, "push", 1, 62, `declare const array: number[]; function foo() { array.push(value); return; }`),
			invalidCase("file.ts", `function foo(stream: string[]) { return stream.push(value); }`, "push", 1, 48, `function foo(stream: string[]) { stream.push(value); return; }`),
			invalidCase("file.ts", `function foo(this: string[]) { return this.push(value); }`, "push", 1, 44, `function foo(this: string[]) { this.push(value); return; }`),
			invalidCase("file.ts", `function foo(router: string[]) { return router.push(value); }`, "push", 1, 48, `function foo(router: string[]) { router.push(value); return; }`),
			invalidCase("file.ts", `function foo(this: {router: string[]}) { return this.router.push(value); }`, "push", 1, 61, `function foo(this: {router: string[]}) { this.router.push(value); return; }`),
			invalidCase("file.ts", `function foo(this: {$router: string[]}) { return this.$router.push(value); }`, "push", 1, 63, `function foo(this: {$router: string[]}) { this.$router.push(value); return; }`),
			invalidCase("file.ts", `function foo(this: {stream: string[]}) { return this.stream.push(value); }`, "push", 1, 61, `function foo(this: {stream: string[]}) { this.stream.push(value); return; }`),
			invalidCase("file.ts", `declare const process: {stdin: string[]}; function foo() { return process.stdin.push(value); }`, "push", 1, 81, `declare const process: {stdin: string[]}; function foo() { process.stdin.push(value); return; }`),
			invalidCase("file.ts", `declare const process: {stdout: string[]}; function foo() { return process.stdout.push(value); }`, "push", 1, 83, `declare const process: {stdout: string[]}; function foo() { process.stdout.push(value); return; }`),
			invalidCase("file.ts", `declare const process: {stderr: string[]}; function foo() { return process.stderr.push(value); }`, "push", 1, 83, `declare const process: {stderr: string[]}; function foo() { process.stderr.push(value); return; }`),
		})
	})
	t.Run("IIFE receivers", func(t *testing.T) {
		run(t, nil, []rule_tester.InvalidTestCase{
			invalidCase("file.js", `function run() { return function () { return []; }().push(1); }`, "push", 1, 54, `function run() { (function () { return []; }().push(1)); return; }`),
			invalidCase("file.js", `function run() { return (function () { return []; }().push(1)); }`, "push", 1, 55, `function run() { (function () { return []; }().push(1)); return; }`),
			invalidCase("file.js", `function run() { return function () { return []; }().unshift(1); }`, "unshift", 1, 54, `function run() { (function () { return []; }().unshift(1)); return; }`),
			invalidCase("file.js", `function run() { return (function () { return []; }().unshift(1)); }`, "unshift", 1, 55, `function run() { (function () { return []; }().unshift(1)); return; }`),
		})
	})
	t.Run("IIFE comments and semicolons", func(t *testing.T) {
		run(t, nil, []rule_tester.InvalidTestCase{
			invalidCase("file.js", `function run() { setup()
return function () { return []; }().push(1); }`, "push", 2, 37, `function run() { setup()
;(function () { return []; }().push(1)); return; }`),
			invalidCase("file.js", `function run() { return function () { return []; }().push(/* keep */ 1); }`, "push", 1, 54, `function run() { (function () { return []; }().push(/* keep */ 1)); return; }`),
			invalidCase("file.ts", `function run() { return function () { return []; }()!.push(1); }`, "push", 1, 55, `function run() { (function () { return []; }()!.push(1)); return; }`),
		})
	})
	t.Run("IIFE suggestion preserves the side effect", func(t *testing.T) {
		const code = `const array = []; function run() { return function () { return array; }().push(1); } run(); array.join();`
		const suggested = `const array = []; function run() { (function () { return array; }().push(1)); return; } run(); array.join();`
		run(t, validCases("file.js", suggested), []rule_tester.InvalidTestCase{
			invalidCase("file.js", code, "push", 1, 75, suggested),
		})
	})
	t.Run("Documentation examples", func(t *testing.T) {
		run(t, validCases("file.js",
			`function add(item) {
	items.push(item);
	return;
}`,
			`const add = item => {
	items.push(item);
};`,
			`items.unshift(item);
const length = items.length;`,
			`function getNextLength(item) {
	items.push(item);
	return items.length;
}`,
			`void items.push(item);`,
			`target.push(...source);`,
		), []rule_tester.InvalidTestCase{
			invalidCase("file.js", `function add(item) {
	return items.push(item);
}`, "push", 2, 15, `function add(item) {
	items.push(item); return;
}`),
			invalidCase("file.js", `const add = item => items.push(item);`, "push", 1, 27, ""),
			invalidCase("file.js", `const length = items.unshift(item);`, "unshift", 1, 22, ""),
			invalidCase("file.js", `source.forEach(item => target.push(item));`, "push", 1, 31, ""),
		})
	})
}

func validCases(filename string, codes ...string) []rule_tester.ValidTestCase {
	cases := make([]rule_tester.ValidTestCase, 0, len(codes))
	for _, code := range codes {
		cases = append(cases, rule_tester.ValidTestCase{Code: code, FileName: filename})
	}
	return cases
}

func invalidCase(filename, code, method string, line, column int, suggested string) rule_tester.InvalidTestCase {
	diagnostic := rule_tester.InvalidTestCaseError{
		MessageId: "no-return-array-push/error",
		Message:   "Do not use the return value of `." + method + "(…)`.",
		Line:      line, Column: column, EndLine: line, EndColumn: column + len(method),
	}
	if suggested != "" {
		diagnostic.Suggestions = []rule_tester.InvalidTestCaseSuggestion{
			{MessageId: "no-return-array-push/suggestion", Output: suggested},
		}
	}
	return rule_tester.InvalidTestCase{Code: code, FileName: filename, Errors: []rule_tester.InvalidTestCaseError{diagnostic}}
}

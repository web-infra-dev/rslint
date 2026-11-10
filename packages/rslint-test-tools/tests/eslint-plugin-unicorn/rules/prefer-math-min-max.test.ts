import path from 'node:path';
import { runInNewContext } from 'node:vm';

import { lint } from '@rslint/core/internal';

import { RuleTester, type ValidTestCase } from '../rule-tester';

// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/prefer-math-min-max.js
const ruleName = 'prefer-math-min-max';
const jsFilename = 'src/virtual.js';
const tester = new RuleTester();

function sourceCase(
  code: string,
  filename = jsFilename,
  project = false,
): ValidTestCase {
  return {
    code,
    filename,
    languageOptions: {
      parserOptions: { project: project ? './tsconfig.virtual.json' : false },
    },
  };
}

function diagnosticCase(
  method: string,
  code: string,
  filename?: string,
  project?: boolean,
) {
  return {
    ...sourceCase(code, filename, project),
    errors: [
      {
        messageId: ruleName,
        message: `Prefer \`Math.${method}()\` to simplify ternary expressions.`,
      },
    ],
  };
}

// The Unicorn wrapper checks diagnostics. The Go mirror checks all ranges and
// applied suggestions, and the runtime regressions below exercise suggestions
// through IPC.
tester.run(ruleName, {} as never, {
  valid: [
    // String initializers and parameter defaults
    sourceCase(`const value = "1"; value > 2 ? 2 : value;`),
    sourceCase('const value = `1`; value > 2 ? 2 : value;'),
    sourceCase('const value = ``; value < 2 ? value : 2;'),
    sourceCase('const value = `\\u0031`; value > 2 ? value : 2;'),
    sourceCase(`function foo(value = "1") { return value > 2 ? 2 : value; }`),
    sourceCase('function foo(value = `1`) { return value > 2 ? 2 : value; }'),
    sourceCase('function foo(value = ``) { return value < 2 ? value : 2; }'),

    // Preserved and discarded comments
    sourceCase(`const value = 1n; (/* keep */ value) > 0 ? value : 0;`),

    // BigInt and number JSDoc in JavaScript
    sourceCase(
      `/** @param {bigint} input */ function run(input) { const value = input * 2n; return value > 0 ? value : 0; }`,
    ),
    sourceCase(`/** @type {bigint} */ let value; value > 0 ? 0 : value;`),
    sourceCase(
      `/** @param {bigint} value */ function run(value) { return value > 0 ? value : 0; }`,
    ),
    sourceCase(
      `/** @param {number | bigint} value */ function run(value, limit) { return value < limit ? limit : value; }`,
    ),
    sourceCase(
      `/** @param {?bigint} value */ function run(value) { return value <= 0 ? value : 0; }`,
    ),
    sourceCase(
      `/** @param {bigint} [value] */ function run(value) { return value >= 0 ? 0 : value; }`,
    ),
    sourceCase(
      `/** @param {number | bigint} limit */ function run(value, limit) { return limit > value ? limit : value; }`,
    ),
    sourceCase(
      `/** @returns {number | bigint} */ function read() { return source; } let value = read(); value > 0 ? value : 0;`,
    ),
    sourceCase(
      `/** @returns {number | bigint} */ function read() { return source; } function run(value = read()) { return value > 0 ? value : 0; }`,
    ),
    sourceCase(
      `/** @param {number | bigint} input */ function run(input) { const value = -input; return value < 0 ? value : 0; }`,
    ),

    // JSDoc on TypeScript parameters
    sourceCase(
      `/** @param {bigint} value */ function run(value) { return value > 0 ? value : 0; }`,
      'src/virtual.ts',
    ),
    sourceCase(
      `/** @param {number | bigint} limit */ function run(value, limit) { return value < limit ? value : limit; }`,
      'src/virtual.ts',
    ),

    // JSDoc on TypeScript parameters with project service
    sourceCase(
      `/** @param {bigint} value */ function run(value) { return value > 0 ? value : 0; }`,
      'src/virtual.ts',
      true,
    ),
    sourceCase(
      `/** @param {number | bigint} limit */ function run(value, limit) { return value < limit ? value : limit; }`,
      'src/virtual.ts',
      true,
    ),

    // JSDoc with the JavaScript parser
    sourceCase(
      `/** @returns {bigint} */ function read() { return source; } const value = read(); value > 0 ? value : 0;`,
    ),
    sourceCase(
      `/** @param {() => bigint} read */ function run(read) { const value = read(); return value > 0 ? value : 0; }`,
    ),
    sourceCase(
      `/** @type {bigint} */ const value = source; const alias = value; alias < 0 ? alias : 0;`,
    ),
    sourceCase(
      `/** @returns {bigint} */ function read() { return source; } function run(value = read()) { return value > 0 ? value : 0; }`,
    ),
    sourceCase(`/** @param {bigint} first
 * @param {bigint} second */
function run(first, second) { const value = condition ? first : second; return value > 0 ? value : 0; }`),

    // JSDoc in TypeScript syntax mode
    sourceCase(
      `/** @returns {bigint} */ function read() { return source; } const value = read(); value > 0 ? value : 0;`,
      'src/virtual.ts',
    ),
    sourceCase(
      `/** @param {() => bigint} read */ function run(read) { const value = read(); return value > 0 ? value : 0; }`,
      'src/virtual.ts',
    ),
    sourceCase(
      `/** @type {bigint} */ const value = source; const alias = value; alias < 0 ? alias : 0;`,
      'src/virtual.ts',
    ),
    sourceCase(
      `/** @returns {bigint} */ function read() { return source; } function run(value = read()) { return value > 0 ? value : 0; }`,
      'src/virtual.ts',
    ),
    sourceCase(
      `/** @param {bigint} first
 * @param {bigint} second */
function run(first, second) { const value = condition ? first : second; return value > 0 ? value : 0; }`,
      'src/virtual.ts',
    ),

    // JSDoc in TypeScript project mode
    sourceCase(
      `/** @returns {bigint} */ function read() { return source; } const value = read(); value > 0 ? value : 0;`,
      'src/virtual.ts',
      true,
    ),
    sourceCase(
      `/** @param {() => bigint} read */ function run(read) { const value = read(); return value > 0 ? value : 0; }`,
      'src/virtual.ts',
      true,
    ),
    sourceCase(
      `/** @type {bigint} */ const value = source; const alias = value; alias < 0 ? alias : 0;`,
      'src/virtual.ts',
      true,
    ),
    sourceCase(
      `/** @returns {bigint} */ function read() { return source; } function run(value = read()) { return value > 0 ? value : 0; }`,
      'src/virtual.ts',
      true,
    ),
    sourceCase(
      `/** @param {bigint} first
 * @param {bigint} second */
function run(first, second) { const value = condition ? first : second; return value > 0 ? value : 0; }`,
      'src/virtual.ts',
      true,
    ),

    // Native return type precedence in TypeScript syntax mode
    sourceCase(
      `/** @returns {number} */ function read(): bigint { return source; } const value = read(); value > 0 ? value : 0;`,
      'src/virtual.ts',
    ),

    // Native return type precedence in TypeScript project mode
    sourceCase(
      `/** @returns {number} */ function read(): bigint { return source; } const value = read(); value > 0 ? value : 0;`,
      'src/virtual.ts',
      true,
    ),

    // JavaScript snapshots
    sourceCase(`height > 10 ? height : 20`),
    sourceCase(`height > 50 ? Math.min(50, height) : height`),
    sourceCase(`foo ? foo : bar`),
    sourceCase(`foo > 10n ? 10n : foo`),
    sourceCase(`foo > BigInt(10) ? BigInt(10) : foo`),
    sourceCase(`new Date() > foo ? foo : new Date()`),
    sourceCase(`foo > new Date(0) ? new Date(0) : foo`),
    sourceCase(`var now = new Date();
var later = new Date();
var value = now > later ? later : now;`),
    sourceCase(`const start = new Date();
var value = start > foo ? foo : start;`),
    sourceCase(`const start = new Date();
function foo() {
	return start > foo ? foo : start;
}`),
    sourceCase(`const start = 'text';
function foo() {
	return start > foo ? foo : start;
}`),
    sourceCase(`const a = BigInt(1);
const value = a > 0 ? 0 : a;`),
    sourceCase(`const a = -1n;
const value = a > 0 ? 0 : a;`),
    sourceCase(`function foo(a = BigInt(1)) {
	return a > 0 ? 0 : a;
}`),
    sourceCase(`function foo(a = 'string', b) {
  return a > b ? a : b;
}`),

    // TypeScript snapshots
    sourceCase(
      `function foo(a, b) {
	return (a as bigint) > b ? a : b;
}`,
      'src/virtual.ts',
    ),
    sourceCase(
      `function foo(a, b) {
	return (a as string) > b ? a : b;
}`,
      'src/virtual.ts',
    ),
    sourceCase(
      `function foo(a, b) {
	return (<bigint>a) > b ? a : b;
}`,
      'src/virtual.ts',
    ),
    sourceCase(
      `function foo(a, b) {
	return (<string>a) > b ? a : b;
}`,
      'src/virtual.ts',
    ),
    sourceCase(
      `function foo(a: string, b) {
	return a > b ? a : b;
}`,
      'src/virtual.ts',
    ),
    sourceCase(
      `function foo(a, b: string) {
	return a > b ? a : b;
}`,
      'src/virtual.ts',
    ),
    sourceCase(
      `function foo(a: bigint, b: bigint) {
	return a > b ? a : b;
}`,
      'src/virtual.ts',
    ),
    sourceCase(
      `function foo(a: Date, b: Date) {
	return a > b ? a : b;
}`,
      'src/virtual.ts',
    ),
    sourceCase(
      `var foo: Date;
var bar: Date;
var value = foo > bar ? bar : foo;`,
      'src/virtual.ts',
    ),
    sourceCase(
      `var foo = 10;
var bar = '20';

var value = foo > bar ? bar : foo;`,
      'src/virtual.ts',
    ),
    sourceCase(
      `var foo = 10;
var bar: string;

var value = foo > bar ? bar : foo;`,
      'src/virtual.ts',
    ),
    sourceCase(
      `function foo(a, b) {
	return (a as string)! > b ? (a as string)! : b;
}`,
      'src/virtual.ts',
    ),
    sourceCase(
      `function foo(a, b) {
	return (<string>a)! > b ? (<string>a)! : b;
}`,
      'src/virtual.ts',
    ),
    sourceCase(
      `function foo(a, b) {
	return (a as Foo) > b ? a : b;
}`,
      'src/virtual.ts',
    ),

    // Documentation examples
    sourceCase(`Math.min(height, 50);`),
    sourceCase(`Math.max(height, 50);`),
  ],
  invalid: [
    // Comments inside operand parentheses
    diagnosticCase('min', `(/* keep */ height) > 50 ? 50 : height;`),
    diagnosticCase('max', `height < (50 /* keep */) ? 50 : height;`),

    // Preserved and discarded comments
    diagnosticCase('min', `((height /* keep */)) >= 50 ? 50 : height;`),
    diagnosticCase(
      'max',
      `(/* left */ height) > (50 /* right */) ? height : 50;`,
    ),
    diagnosticCase('min', 'height < (50 // keep\r\n) ? height : 50;'),
    diagnosticCase(
      'max',
      `(/* keep */ first, height) < 50 ? 50 : (first, height);`,
    ),
    diagnosticCase(
      'min',
      `(height as /* keep */ number) > 50 ? 50 : height;`,
      'src/virtual.ts',
    ),
    diagnosticCase('min', `(height /* discard */ > 50) ? 50 : height;`),
    diagnosticCase('min', `(/* discard */ height > 50) ? 50 : height;`),
    diagnosticCase('min', `height > 50 ? 50 : (/* discard */ height);`),
    diagnosticCase('min', `(/* keep */ read()) > 50 ? 50 : read();`),

    // BigInt and number JSDoc in JavaScript
    diagnosticCase(
      'min',
      `/** @param {number} value */ function run(value) { return value > 0 ? 0 : value; }`,
    ),
    diagnosticCase(
      'max',
      `/** @param {number} value */ function run(value) { return value > 0 ? value : 0; }`,
    ),

    // JSDoc on TypeScript parameters
    diagnosticCase(
      'max',
      `/** @param {bigint} value */ function run(value: number) { return value > 0 ? value : 0; }`,
      'src/virtual.ts',
    ),

    // JSDoc on TypeScript parameters with project service
    diagnosticCase(
      'max',
      `/** @param {bigint} value */ function run(value: number) { return value > 0 ? value : 0; }`,
      'src/virtual.ts',
      true,
    ),

    // JSDoc with the JavaScript parser
    diagnosticCase(
      'max',
      `/** @returns {number} */ function read() { return source; } const value = read(); value > 0 ? value : 0;`,
    ),

    // JSDoc in TypeScript syntax mode
    diagnosticCase(
      'max',
      `/** @returns {number} */ function read() { return source; } const value = read(); value > 0 ? value : 0;`,
      'src/virtual.ts',
    ),

    // JSDoc in TypeScript project mode
    diagnosticCase(
      'max',
      `/** @returns {number} */ function read() { return source; } const value = read(); value > 0 ? value : 0;`,
      'src/virtual.ts',
      true,
    ),

    // Native return type precedence in TypeScript syntax mode
    diagnosticCase(
      'max',
      `/** @returns {bigint} */ function read(): number { return source; } const value = read(); value > 0 ? value : 0;`,
      'src/virtual.ts',
    ),

    // Native return type precedence in TypeScript project mode
    diagnosticCase(
      'max',
      `/** @returns {bigint} */ function read(): number { return source; } const value = read(); value > 0 ? value : 0;`,
      'src/virtual.ts',
      true,
    ),

    // NaN and signed zero
    diagnosticCase('max', `const value = NaN; value > 0 ? value : 0;`),
    diagnosticCase('min', `const value = -0; value < 0 ? value : 0;`),

    // JavaScript snapshots
    diagnosticCase('min', `height > 50 ? 50 : height`),
    diagnosticCase('min', `height >= 50 ? 50 : height`),
    diagnosticCase('min', `height < 50 ? height : 50`),
    diagnosticCase('min', `height <= 50 ? height : 50`),
    diagnosticCase('min', `height > maxHeight ? maxHeight : height`),
    diagnosticCase('min', `height < maxHeight ? height : maxHeight`),
    diagnosticCase('min', `window.height > 50 ? 50 : window.height`),
    diagnosticCase('min', `window.height < 50 ? window.height : 50`),
    diagnosticCase('max', `height > 50 ? height : 50`),
    diagnosticCase('max', `height >= 50 ? height : 50`),
    diagnosticCase('max', `height < 50 ? 50 : height`),
    diagnosticCase('max', `height <= 50 ? 50 : height`),
    diagnosticCase('max', `height > maxHeight ? height : maxHeight`),
    diagnosticCase('max', `height < maxHeight ? maxHeight : height`),
    diagnosticCase(
      'min',
      `function a() {
	return +foo > 10 ? 10 : +foo
}`,
    ),
    diagnosticCase(
      'min',
      `function a() {
	return+foo > 10 ? 10 : +foo
}`,
    ),
    diagnosticCase('min', `(0,foo) > 10 ? 10 : (0,foo)`),
    diagnosticCase('max', `foo > (0,bar) ? foo : (0,bar)`),
    diagnosticCase('min', `Number(x) > 50 ? 50 : Number(x)`),
    diagnosticCase('min', `foo.bar() > 10 ? 10 : foo.bar()`),
    diagnosticCase(
      'min',
      `async function foo() {
	return await foo.bar() > 10 ? 10 : await foo.bar()
}`,
    ),
    diagnosticCase(
      'min',
      `async function foo() {
	await(+foo > 10 ? 10 : +foo)
}`,
    ),
    diagnosticCase(
      'min',
      `function foo() {
	return(foo.bar() > 10) ? 10 : foo.bar()
}`,
    ),
    diagnosticCase(
      'min',
      `function* foo() {
	yield+foo > 10 ? 10 : +foo
}`,
    ),
    diagnosticCase('min', `export default+foo > 10 ? 10 : +foo`),
    diagnosticCase('min', `foo.length > bar.length ? bar.length : foo.length`),

    // Comments outside comparison operands
    diagnosticCase(
      'min',
      `const value = height > // keep
	50 ? 50 : height;`,
    ),
    diagnosticCase(
      'min',
      `const value = height > /* keep */ 50 ? 50 : height;`,
    ),
    diagnosticCase(
      'min',
      `const value = height > 50 /* keep */ ? 50 : height;`,
    ),
    diagnosticCase(
      'min',
      `const value = height > 50 ? 50 /* keep */ : height;`,
    ),
    diagnosticCase(
      'min',
      `const value = height > 50 ? /* keep */ 50 : height;`,
    ),
    diagnosticCase('min', `const value = height > 50 ? 50 : height;`),

    // TypeScript snapshots
    diagnosticCase(
      'max',
      `function foo(a, b) {
	return (a as number) > b ? a : b;
}`,
      'src/virtual.ts',
    ),
    diagnosticCase(
      'max',
      `function foo(a, b) {
	return (a as number) > b ? a : b;
}`,
      'src/virtual.ts',
    ),
    diagnosticCase(
      'max',
      `function foo(a, b) {
	return (a as unknown as number) > b ? a : b;
}`,
      'src/virtual.ts',
    ),
    diagnosticCase(
      'min',
      `var foo = 10;

var value = foo > bar ? bar : foo;`,
      'src/virtual.ts',
    ),
    diagnosticCase(
      'min',
      `var foo = 10;
var bar = 20;

var value = foo > bar ? bar : foo;`,
      'src/virtual.ts',
    ),
    diagnosticCase(
      'min',
      `var foo: number;
var bar: number;

var value = foo > bar ? bar : foo;`,
      'src/virtual.ts',
    ),
    diagnosticCase(
      'max',
      `function foo(a, b) {
	return (a as number) > b ? (a as number) : b;
}`,
      'src/virtual.ts',
    ),
    diagnosticCase(
      'max',
      `function foo(a, b) {
	return (<number>a) > b ? (<number>a) : b;
}`,
      'src/virtual.ts',
    ),
    diagnosticCase(
      'max',
      `function foo(a, b) {
	return a! > b ? a! : b;
}`,
      'src/virtual.ts',
    ),
    diagnosticCase(
      'max',
      `function foo(a, b) {
	return (a as Number) > b ? a : b;
}`,
      'src/virtual.ts',
    ),

    // Repeated evaluation of operands
    diagnosticCase('max', `const a = f() > 0 ? f() : 0;`),
    diagnosticCase('min', `const a = (x++, y) > 0 ? 0 : (x++, y);`),
    diagnosticCase('min', `const a = (x = 1, y) > 0 ? 0 : (x = 1, y);`),
    diagnosticCase('min', `const a = (x || f()) > 0 ? 0 : (x || f());`),
    diagnosticCase(
      'min',
      `const a = (x ? f() : g()) > 0 ? 0 : (x ? f() : g());`,
    ),
    diagnosticCase('min', `const a = [f()] > 0 ? 0 : [f()];`),
    diagnosticCase('min', `const a = ({a: f()}) > 0 ? 0 : ({a: f()});`),
    diagnosticCase('min', `const a = (x, y) > 0 ? 0 : (x, y);`),
    diagnosticCase('min', 'const a = `${x}` > 0 ? 0 : `${x}`;'),
    diagnosticCase('min', 'const a = `1` > 0 ? 0 : `1`;'),
    diagnosticCase('min', `const a = [() => f()] > 0 ? 0 : [() => f()];`),
    diagnosticCase(
      'min',
      `import foo from "foo"; const a = foo > 0 ? 0 : foo;`,
    ),

    // Documentation examples
    diagnosticCase('min', `height > 50 ? 50 : height;`),
    diagnosticCase('max', `height > 50 ? height : 50;`),
    diagnosticCase('min', `height >= 50 ? 50 : height;`),
    diagnosticCase('max', `height >= 50 ? height : 50;`),
    diagnosticCase('min', `height < 50 ? height : 50;`),
    diagnosticCase('max', `height < 50 ? 50 : height;`),
    diagnosticCase('min', `height <= 50 ? height : 50;`),
    diagnosticCase('max', `height <= 50 ? 50 : height;`),
  ],
});

const configDirectory = path.resolve(import.meta.dirname, '..');
const absoluteFilename = path.resolve(configDirectory, jsFilename);

function lintCode(code: string) {
  return lint({
    config: [
      { plugins: ['unicorn'], rules: { [`unicorn/${ruleName}`]: 'error' } },
    ],
    configDirectory,
    workingDirectory: process.cwd(),
    fileContents: { [absoluteFilename]: code },
    fix: true,
  });
}

test.each([
  {
    method: 'max',
    code: '(/* keep */ value) > limit ? value : limit;',
    suggested: 'Math.max((/* keep */ value), limit);',
    cases: [
      [5, 0, 5, 5],
      [-5, 0, 0, 0],
      [NaN, 0, 0, NaN],
    ],
  },
  {
    method: 'min',
    code: '(/* keep */ value) < limit ? value : limit;',
    suggested: 'Math.min((/* keep */ value), limit);',
    cases: [
      [5, 0, 0, 0],
      [-5, 0, -5, -5],
      [-0, 0, 0, -0],
    ],
  },
])(
  'keeps Math.$method rewrites manual',
  async ({ method, code, suggested, cases }) => {
    const result = await lintCode(code);
    expect(result).toMatchObject({
      fileCount: 1,
      ruleCount: 1,
      fixableErrorCount: 0,
    });
    expect(result.output ?? {}).toEqual({});
    expect(result.diagnostics).toHaveLength(1);
    const [diagnostic] = result.diagnostics;
    expect(diagnostic.fixes).toBeUndefined();
    expect(diagnostic.suggestions).toHaveLength(1);
    const suggestion = diagnostic.suggestions?.[0];
    assert.ok(suggestion?.fixes);
    expect(suggestion).toMatchObject({
      messageId: 'replace-with-math-min-max',
      message: `Replace with \`Math.${method}()\`.`,
      data: { method },
    });
    const fixes = [...suggestion.fixes].sort((a, b) => b.startPos - a.startPos);
    const rewritten = fixes.reduce(
      (text, fix) =>
        text.slice(0, fix.startPos) + fix.text + text.slice(fix.endPos),
      code,
    );
    expect(rewritten).toBe(suggested);
    expect((await lintCode(rewritten)).diagnostics).toEqual([]);
    for (const [value, limit, originalResult, suggestedResult] of cases) {
      expect(runInNewContext(code, { value, limit })).toBe(originalResult);
      expect(runInNewContext(rewritten, { value, limit })).toBe(
        suggestedResult,
      );
    }
  },
);

test('leaves documented BigInt comparisons intact', async () => {
  for (const [operator, value] of [
    ['<', -5n],
    ['>', 5n],
  ] as const) {
    const code = `/** @param {bigint} value */ function run(value) { return value ${operator} 0 ? value : 0; } run(${value}n);`;
    const result = await lintCode(code);
    expect(result).toMatchObject({
      fileCount: 1,
      ruleCount: 1,
      diagnostics: [],
    });
    expect(result.output ?? {}).toEqual({});
    expect(runInNewContext(code)).toBe(value);
  }
  for (const method of ['min', 'max']) {
    // The thrown error comes from a separate JavaScript realm.
    expect(() => runInNewContext(`Math.${method}(5n, 0)`)).toThrow(
      expect.objectContaining({ name: 'TypeError' }),
    );
  }
});

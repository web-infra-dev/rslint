// Upstream tests and documentation: eslint-plugin-unicorn v79.0.0.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/prefer-global-this.js
import assert from 'node:assert/strict';
import path from 'node:path';
import { runInNewContext } from 'node:vm';
import { transpileModule, ScriptTarget } from 'typescript';
import { lint } from '@rslint/core/internal';
import { RuleTester, type ValidTestCase } from '../rule-tester';

function valid(
  code: string,
  filename = 'src/virtual.js',
  languageOptions?: ValidTestCase['languageOptions'],
) {
  return { code, filename, languageOptions };
}
function invalid(
  code: string,
  output: string | null,
  locations: [
    value: string,
    line: number,
    column: number,
    endLine: number,
    endColumn: number,
  ][],
  filename = 'src/virtual.js',
  languageOptions?: ValidTestCase['languageOptions'],
) {
  return {
    ...valid(code, filename, languageOptions),
    output,
    errors: locations.map(([value, line, column, endLine, endColumn]) => ({
      messageId: 'prefer-global-this/error',
      message: `Prefer \`globalThis\` over \`${value}\`.`,
      line,
      column,
      endLine,
      endColumn,
    })),
  };
}
const cases = {
  valid: [
    // Upstream.
    valid('globalThis'),
    valid('globalThis.foo'),
    valid('globalThis[foo]'),
    valid('globalThis.foo()'),
    valid('const { foo } = globalThis'),
    valid('function foo (window) {}'),
    valid('function foo (global) {}'),
    valid('var foo = function foo (window) {}'),
    valid('var foo = function foo (global) {}'),
    valid('var window = {}'),
    valid('let global = {}'),
    valid('const global = {}'),
    valid('function foo (window) {\n\twindow.foo();\n}'),
    valid('var window = {};\nfunction foo () {\n\twindow.foo();\n}'),
    valid('foo.window'),
    valid('foo.global'),
    valid('import window from "xxx"'),
    valid('import * as window from "xxx"'),
    valid('import window, {foo} from "xxx"'),
    valid('export { window }  from "xxx"'),
    valid('export * as window from "xxx";'),
    valid('try {\n\n} catch (window) {}'),
    valid('window.name = "foo"'),
    valid('window.addEventListener'),
    valid('window.innerWidth'),
    valid('window.innerHeight'),
    valid('window.addEventListener("resize", () => {})'),
    valid('window.addEventListener(`resize`, () => {})'),
    valid('window.removeEventListener(`resize`, () => {})'),
    valid('window.onresize = function () {}'),
    valid('\'open\' in window; window.open("https://example.com")'),
    valid('`open` in window; window.open("https://example.com")'),
    valid('const {window} = jsdom()\nwindow.jQuery = jQuery;'),
    valid('({ foo: window.name } =  {})'),
    valid('[window.name] = []'),
    valid('window[foo]'),
    valid('window[title]'),
    valid('window["foo"]'),
    valid('typeof window'),
    valid('typeof window !== "undefined"'),
    valid('typeof self !== "undefined"'),
    valid('typeof global !== "undefined"'),
    valid('typeof window === "undefined"'),
    valid('typeof window === "object"'),
    valid('type Window = typeof window;', 'src/virtual.ts'),
    valid('let foo: typeof self;', 'src/virtual.ts'),
    valid('type Foo = typeof window.foo;', 'src/virtual.ts'),
    valid('type Foo = typeof window.foo.bar;', 'src/virtual.ts'),
    valid('function foo(bar: typeof self.foo) {}', 'src/virtual.ts'),
    // TypeScript.
    valid('typeof (window as unknown);', 'src/virtual.ts'),
    valid('typeof (window satisfies unknown);', 'src/virtual.ts'),
    valid('typeof (<unknown>window);', 'src/virtual.ts'),
    valid('typeof window!;', 'src/virtual.ts'),
    valid('typeof (self as unknown);', 'src/virtual.ts'),
    valid('typeof (self satisfies unknown);', 'src/virtual.ts'),
    valid('typeof (<unknown>self);', 'src/virtual.ts'),
    valid('typeof self!;', 'src/virtual.ts'),
    valid('typeof (global as unknown);', 'src/virtual.ts'),
    valid('typeof (global satisfies unknown);', 'src/virtual.ts'),
    valid('typeof (<unknown>global);', 'src/virtual.ts'),
    valid('typeof global!;', 'src/virtual.ts'),
    // Documentation.
    valid('globalThis;'),
    valid('globalThis.foo;'),
    valid('globalThis;'),
    valid('globalThis.foo;'),
    valid('const {foo} = globalThis;'),
    valid('globalThis.navigator;'),
    valid('globalThis.location;'),
    valid('window.innerWidth;'),
    valid('window.innerHeight;'),
    valid("typeof window === 'undefined';"),
    valid("typeof window === 'object';"),
    valid("globalThis.addEventListener('click', () => {});"),
    valid("window.addEventListener('resize', () => {});"),
    valid("window.addEventListener('load', () => {});"),
    valid("window.addEventListener('unload', () => {});"),
    valid('window[foo]'),
    valid("typeof window === 'undefined'"),
  ],
  invalid: [
    // Upstream.
    invalid('global', 'globalThis', [['global', 1, 1, 1, 7]]),
    invalid('self', 'globalThis', [['self', 1, 1, 1, 5]]),
    invalid('window', 'globalThis', [['window', 1, 1, 1, 7]]),
    invalid('window.foo', 'globalThis.foo', [['window', 1, 1, 1, 7]]),
    invalid('window.foo()', 'globalThis.foo()', [['window', 1, 1, 1, 7]]),
    invalid('window > 10', 'globalThis > 10', [['window', 1, 1, 1, 7]]),
    invalid('10 > window', '10 > globalThis', [['window', 1, 6, 1, 12]]),
    invalid('window ?? 10', 'globalThis ?? 10', [['window', 1, 1, 1, 7]]),
    invalid('10 ?? window', '10 ?? globalThis', [['window', 1, 7, 1, 13]]),
    invalid('window.foo = 123', 'globalThis.foo = 123', [
      ['window', 1, 1, 1, 7],
    ]),
    invalid('window = 123', null, [['window', 1, 1, 1, 7]]),
    invalid('obj.a = window', 'obj.a = globalThis', [['window', 1, 9, 1, 15]]),
    invalid(
      'function* gen() {\n  yield window\n}',
      'function* gen() {\n  yield globalThis\n}',
      [['window', 2, 9, 2, 15]],
    ),
    invalid(
      'async function gen() {\n  await window\n}',
      'async function gen() {\n  await globalThis\n}',
      [['window', 2, 9, 2, 15]],
    ),
    invalid('window ? foo : bar', 'globalThis ? foo : bar', [
      ['window', 1, 1, 1, 7],
    ]),
    invalid('foo ? window : bar', 'foo ? globalThis : bar', [
      ['window', 1, 7, 1, 13],
    ]),
    invalid('foo ? bar : window', 'foo ? bar : globalThis', [
      ['window', 1, 13, 1, 19],
    ]),
    invalid(
      'function foo() {\n  return window\n}',
      'function foo() {\n  return globalThis\n}',
      [['window', 2, 10, 2, 16]],
    ),
    invalid('new window()', 'new globalThis()', [['window', 1, 5, 1, 11]]),
    invalid(
      'const obj = {\n\tfoo: window.foo,\n\tbar: window.bar,\n\twindow: window\n}',
      'const obj = {\n\tfoo: globalThis.foo,\n\tbar: globalThis.bar,\n\twindow: globalThis\n}',
      [
        ['window', 2, 7, 2, 13],
        ['window', 3, 7, 3, 13],
        ['window', 4, 10, 4, 16],
      ],
    ),
    invalid(
      'function sequenceTest() {\n\tlet x, y;\n\tx = (y = 10, y + 5, window);\n\tconsole.log(x, y);\n}',
      'function sequenceTest() {\n\tlet x, y;\n\tx = (y = 10, y + 5, globalThis);\n\tconsole.log(x, y);\n}',
      [['window', 3, 22, 3, 28]],
    ),
    invalid('window`Hello ${42} World`', 'globalThis`Hello ${42} World`', [
      ['window', 1, 1, 1, 7],
    ]),
    invalid(
      'tag`Hello ${window.foo} World`',
      'tag`Hello ${globalThis.foo} World`',
      [['window', 1, 13, 1, 19]],
    ),
    invalid(
      'var str = `hello ${window.foo} world!`',
      'var str = `hello ${globalThis.foo} world!`',
      [['window', 1, 20, 1, 26]],
    ),
    invalid('delete window.foo', 'delete globalThis.foo', [
      ['window', 1, 8, 1, 14],
    ]),
    invalid('++window', null, [['window', 1, 3, 1, 9]]),
    invalid('++window.foo', '++globalThis.foo', [['window', 1, 3, 1, 9]]),
    invalid(
      'for (var attr in window) {\n\n}',
      'for (var attr in globalThis) {\n\n}',
      [['window', 1, 18, 1, 24]],
    ),
    invalid(
      'for (window.foo = 0; i < 10; window.foo++) {\n\n}',
      'for (globalThis.foo = 0; i < 10; globalThis.foo++) {\n\n}',
      [
        ['window', 1, 6, 1, 12],
        ['window', 1, 30, 1, 36],
      ],
    ),
    invalid(
      'for (const item of window.foo) {\n}',
      'for (const item of globalThis.foo) {\n}',
      [['window', 1, 20, 1, 26]],
    ),
    invalid(
      'for (const item of window) {\n}',
      'for (const item of globalThis) {\n}',
      [['window', 1, 20, 1, 26]],
    ),
    invalid('switch (window) {}', 'switch (globalThis) {}', [
      ['window', 1, 9, 1, 15],
    ]),
    invalid(
      'switch (true) {\n\tcase window:\n\t\tbreak;\n}',
      'switch (true) {\n\tcase globalThis:\n\t\tbreak;\n}',
      [['window', 2, 7, 2, 13]],
    ),
    invalid(
      'switch (true) {\n\tcase window.foo:\n\t\tbreak;\n}',
      'switch (true) {\n\tcase globalThis.foo:\n\t\tbreak;\n}',
      [['window', 2, 7, 2, 13]],
    ),
    invalid('while (window) {\n}', 'while (globalThis) {\n}', [
      ['window', 1, 8, 1, 14],
    ]),
    invalid('do {} while (window) {}', 'do {} while (globalThis) {}', [
      ['window', 1, 14, 1, 20],
    ]),
    invalid('if (window) {}', 'if (globalThis) {}', [['window', 1, 5, 1, 11]]),
    invalid('throw window', 'throw globalThis', [['window', 1, 7, 1, 13]]),
    invalid('var foo = window', 'var foo = globalThis', [
      ['window', 1, 11, 1, 17],
    ]),
    invalid(
      'function foo (name = window) {\n\n}',
      'function foo (name = globalThis) {\n\n}',
      [['window', 1, 22, 1, 28]],
    ),
    invalid('self.innerWidth', 'globalThis.innerWidth', [['self', 1, 1, 1, 5]]),
    invalid('self.innerHeight', 'globalThis.innerHeight', [
      ['self', 1, 1, 1, 5],
    ]),
    invalid('self.location', 'globalThis.location', [['self', 1, 1, 1, 5]]),
    invalid('self.navigator', 'globalThis.navigator', [['self', 1, 1, 1, 5]]),
    invalid('self.name', 'globalThis.name', [['self', 1, 1, 1, 5]]),
    invalid('self.postMessage("Hello")', 'globalThis.postMessage("Hello")', [
      ['self', 1, 1, 1, 5],
    ]),
    invalid(
      'self.onconnect = function () {}',
      'globalThis.onconnect = function () {}',
      [['self', 1, 1, 1, 5]],
    ),
    invalid(
      'self.onerror = function () {}',
      'globalThis.onerror = function () {}',
      [['self', 1, 1, 1, 5]],
    ),
    invalid(
      'self.onlanguagechange = function () {}',
      'globalThis.onlanguagechange = function () {}',
      [['self', 1, 1, 1, 5]],
    ),
    invalid(
      'self.onoffline = function () {}',
      'globalThis.onoffline = function () {}',
      [['self', 1, 1, 1, 5]],
    ),
    invalid(
      'self.ononline = function () {}',
      'globalThis.ononline = function () {}',
      [['self', 1, 1, 1, 5]],
    ),
    invalid(
      'self.addEventListener("message", () => {})',
      'globalThis.addEventListener("message", () => {})',
      [['self', 1, 1, 1, 5]],
    ),
    invalid(
      'self.removeEventListener("message", () => {})',
      'globalThis.removeEventListener("message", () => {})',
      [['self', 1, 1, 1, 5]],
    ),
    invalid(
      'self.dispatchEvent(new Event("message"))',
      'globalThis.dispatchEvent(new Event("message"))',
      [['self', 1, 1, 1, 5]],
    ),
    invalid('self.self', 'globalThis.self', [['self', 1, 1, 1, 5]]),
    invalid(
      'self.onrejectionhandled = function () {}',
      'globalThis.onrejectionhandled = function () {}',
      [['self', 1, 1, 1, 5]],
    ),
    invalid(
      'self.onunhandledrejection = function () {}',
      'globalThis.onunhandledrejection = function () {}',
      [['self', 1, 1, 1, 5]],
    ),
    invalid("'postMessage' in self", "'postMessage' in globalThis", [
      ['self', 1, 18, 1, 22],
    ]),
    invalid('window.crypto', 'globalThis.crypto', [['window', 1, 1, 1, 7]]),
    invalid(
      'window.addEventListener("play", () => {})',
      'globalThis.addEventListener("play", () => {})',
      [['window', 1, 1, 1, 7]],
    ),
    invalid(
      'window.onplay = function () {}',
      'globalThis.onplay = function () {}',
      [['window', 1, 1, 1, 7]],
    ),
    invalid(
      'function greet({ name = window.foo }) {}',
      'function greet({ name = globalThis.foo }) {}',
      [['window', 1, 25, 1, 31]],
    ),
    invalid('({ foo: window.foo } =  {})', '({ foo: globalThis.foo } =  {})', [
      ['window', 1, 9, 1, 15],
    ]),
    invalid('[window.foo] = []', '[globalThis.foo] = []', [
      ['window', 1, 2, 1, 8],
    ]),
    invalid('foo[window]', 'foo[globalThis]', [['window', 1, 5, 1, 11]]),
    invalid('foo[window.foo]', 'foo[globalThis.foo]', [
      ['window', 1, 5, 1, 11],
    ]),
    invalid("'foo' in window", "'foo' in globalThis", [
      ['window', 1, 10, 1, 16],
    ]),
    invalid("'foo' in global", "'foo' in globalThis", [
      ['global', 1, 10, 1, 16],
    ]),
    invalid(
      'typeof window.something === "function"',
      'typeof globalThis.something === "function"',
      [['window', 1, 8, 1, 14]],
    ),
    invalid(
      'typeof self.something === "function"',
      'typeof globalThis.something === "function"',
      [['self', 1, 8, 1, 12]],
    ),
    invalid(
      'typeof global.something === "function"',
      'typeof globalThis.something === "function"',
      [['global', 1, 8, 1, 14]],
    ),
    invalid('self[foo]', 'globalThis[foo]', [['self', 1, 1, 1, 5]]),
    invalid('global[foo]', 'globalThis[foo]', [['global', 1, 1, 1, 7]]),
    invalid(
      'const foo: typeof window = window;',
      'const foo: typeof window = globalThis;',
      [['window', 1, 28, 1, 34]],
      'src/virtual.ts',
    ),
    invalid('const foo = {window};', 'const foo = {window: globalThis};', [
      ['window', 1, 14, 1, 20],
    ]),
    invalid(
      'const foo = {window, self};',
      'const foo = {window: globalThis, self: globalThis};',
      [
        ['window', 1, 14, 1, 20],
        ['self', 1, 22, 1, 26],
      ],
    ),
    invalid('foo({window});', 'foo({window: globalThis});', [
      ['window', 1, 6, 1, 12],
    ]),
    invalid('const foo = {[window]: 1};', 'const foo = {[globalThis]: 1};', [
      ['window', 1, 15, 1, 21],
    ]),
    invalid('({window} = foo);', null, [['window', 1, 3, 1, 9]]),
    invalid('({window = 1} = foo);', null, [['window', 1, 3, 1, 9]]),
    // GlobalsOff.
    invalid(
      'global.global_did_not_declare_in_language_options',
      'globalThis.global_did_not_declare_in_language_options',
      [['global', 1, 1, 1, 7]],
      'src/virtual.js',
      { globals: { global: 'off', window: 'off', self: 'off' } },
    ),
    invalid(
      'window.window_did_not_declare_in_language_options',
      'globalThis.window_did_not_declare_in_language_options',
      [['window', 1, 1, 1, 7]],
      'src/virtual.js',
      { globals: { global: 'off', window: 'off', self: 'off' } },
    ),
    invalid(
      'self.self_did_not_declare_in_language_options',
      'globalThis.self_did_not_declare_in_language_options',
      [['self', 1, 1, 1, 5]],
      'src/virtual.js',
      { globals: { global: 'off', window: 'off', self: 'off' } },
    ),
    // Writes.
    invalid('window = 123;', null, [['window', 1, 1, 1, 7]]),
    invalid('self = 1;', null, [['self', 1, 1, 1, 5]]),
    invalid('window += 1;', null, [['window', 1, 1, 1, 7]]),
    invalid('[window] = foo;', null, [['window', 1, 2, 1, 8]]),
    invalid('({window} = foo);', null, [['window', 1, 3, 1, 9]]),
    invalid('for (global of foo);', null, [['global', 1, 6, 1, 12]]),
    invalid('window.foo = window;', 'globalThis.foo = globalThis;', [
      ['window', 1, 1, 1, 7],
      ['window', 1, 14, 1, 20],
    ]),
    // TypeScript.
    invalid(
      'const value = (window as unknown);',
      'const value = (globalThis as unknown);',
      [['window', 1, 16, 1, 22]],
      'src/virtual.ts',
    ),
    invalid(
      'typeof (window as unknown).value;',
      'typeof (globalThis as unknown).value;',
      [['window', 1, 9, 1, 15]],
      'src/virtual.ts',
    ),
    invalid(
      'const value = (self as unknown);',
      'const value = (globalThis as unknown);',
      [['self', 1, 16, 1, 20]],
      'src/virtual.ts',
    ),
    invalid(
      'typeof (self as unknown).value;',
      'typeof (globalThis as unknown).value;',
      [['self', 1, 9, 1, 13]],
      'src/virtual.ts',
    ),
    invalid(
      'const value = (global as unknown);',
      'const value = (globalThis as unknown);',
      [['global', 1, 16, 1, 22]],
      'src/virtual.ts',
    ),
    invalid(
      'typeof (global as unknown).value;',
      'typeof (globalThis as unknown).value;',
      [['global', 1, 9, 1, 15]],
      'src/virtual.ts',
    ),
    // Documentation.
    invalid('window;', 'globalThis;', [['window', 1, 1, 1, 7]]),
    invalid('window.foo;', 'globalThis.foo;', [['window', 1, 1, 1, 7]]),
    invalid('global;', 'globalThis;', [['global', 1, 1, 1, 7]]),
    invalid('global.foo;', 'globalThis.foo;', [['global', 1, 1, 1, 7]]),
    invalid('const {foo} = window;', 'const {foo} = globalThis;', [
      ['window', 1, 15, 1, 21],
    ]),
    invalid('window.navigator;', 'globalThis.navigator;', [
      ['window', 1, 1, 1, 7],
    ]),
    invalid('window.location;', 'globalThis.location;', [
      ['window', 1, 1, 1, 7],
    ]),
    invalid(
      "window.addEventListener('click', () => {});",
      "globalThis.addEventListener('click', () => {});",
      [['window', 1, 1, 1, 7]],
    ),
  ],
};
new RuleTester().run('prefer-global-this', null as never, cases);

async function lintCase(item: ReturnType<typeof valid>) {
  const filename = path.resolve(import.meta.dirname, item.filename);
  const result = await lint({
    config: [
      {
        plugins: ['unicorn'],
        languageOptions: {
          parserOptions: { project: false, projectService: false },
          ...item.languageOptions,
        },
        rules: { 'unicorn/prefer-global-this': 'error' },
      },
    ],
    configDirectory: import.meta.dirname,
    workingDirectory: process.cwd(),
    fileContents: { [filename]: item.code },
  });
  assert.equal(result.fileCount, 1, item.code);
  assert.equal(result.ruleCount, 1, item.code);
  return result;
}

// The suite wrapper checks messages; verify complete ranges and edits through IPC.
test('upstream diagnostic ranges and autofixes', async () => {
  for (const item of cases.invalid) {
    const result = await lintCase(item);
    assert.equal(result.diagnostics.length, item.errors.length, item.code);
    for (const [index, diagnostic] of result.diagnostics.entries()) {
      const expected = item.errors[index];
      assert.equal(
        diagnostic.ruleName,
        'unicorn/prefer-global-this',
        item.code,
      );
      assert.equal(diagnostic.messageId, expected.messageId, item.code);
      assert.equal(diagnostic.message, expected.message, item.code);
      assert.deepEqual(
        diagnostic.range,
        {
          start: { line: expected.line, column: expected.column },
          end: { line: expected.endLine, column: expected.endColumn },
        },
        item.code,
      );
      assert.equal(diagnostic.suggestions?.length ?? 0, 0, item.code);
    }
    const fixes = result.diagnostics
      .flatMap((d) => d.fixes ?? [])
      .sort((a, b) => b.startPos - a.startPos);
    let output = item.code;
    for (const fix of fixes)
      output =
        output.slice(0, fix.startPos) + fix.text + output.slice(fix.endPos);
    assert.equal(fixes.length ? output : null, item.output, item.code);
  }
});

// test/prefer-global-this.js: wrapped typeof checks still run with missing globals.
test('wrapped typeof checks remain portable', async () => {
  for (const name of ['window', 'self', 'global']) {
    for (const expression of [
      '(' + name + ' as unknown)',
      '(' + name + ' satisfies unknown)',
      '(<unknown>' + name + ')',
      name + '!',
    ]) {
      const code = 'typeof ' + expression + ';';
      const result = await lintCase(valid(code, 'src/virtual.ts'));
      assert.deepEqual(result.diagnostics, [], code);
      const { outputText, diagnostics } = transpileModule(code, {
        compilerOptions: { target: ScriptTarget.ESNext },
        reportDiagnostics: true,
      });
      assert.deepEqual(diagnostics, [], code);
      assert.equal(runInNewContext(outputText), 'undefined', code);
    }
  }
});

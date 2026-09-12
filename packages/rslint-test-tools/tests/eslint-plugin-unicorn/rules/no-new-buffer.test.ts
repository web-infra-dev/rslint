import path from 'node:path';

import { lint } from '@rslint/core/internal';

import { RuleTester } from '../rule-tester';
import { buildConfigForSettings } from '../../src/util/load-test-config';

const ruleTester = new RuleTester();

const unknown =
  '`new Buffer()` is deprecated, use `Buffer.alloc()` or `Buffer.from()` instead.';
const from = '`new Buffer()` is deprecated, use `Buffer.from()` instead.';
const alloc = '`new Buffer()` is deprecated, use `Buffer.alloc()` instead.';
const valid = (code: string) => ({ code, filename: 'file.js' });
const invalid = (code: string, message: string, filename = 'file.js') => ({
  code,
  filename,
  errors: [{ message }],
});

// Complete semantic mirror of eslint-plugin-unicorn v74.0.0 test/no-new-buffer.js.
ruleTester.run('no-new-buffer', null as never, {
  valid: [
    valid('const buffer = Buffer'),
    valid('const buffer = new NotBuffer(1)'),
    valid("const buffer = Buffer.from('buf')"),
    valid(
      "const buffer = Buffer.from('7468697320697320612074c3a97374', 'hex')",
    ),
    valid('const buffer = Buffer.from([0x62, 0x75, 0x66, 0x66, 0x65, 0x72])'),
    valid('const buffer = Buffer.alloc(10)'),
  ],
  invalid: [
    invalid(
      "const modes = new Set(['foo']);\nmodes.clear();\nnew Buffer(modes.size ? 'x' : 1);",
      unknown,
    ),
    invalid(
      'const modes = new Set(["foo"]); modes.clear(); new Buffer((modes.size && "x") || value);',
      unknown,
    ),
    invalid(
      'const alias = condition; var condition = true; new Buffer(alias ? "x" : 1);',
      unknown,
    ),
    invalid(
      'const object = {value: true}; Object.defineProperty(object, "value", {get() { return false; }}); new Buffer(object.value ? "x" : value);',
      unknown,
    ),
    invalid(
      'const modes = new Set(["foo"]); modes.clear(); new Buffer((modes.size ? 1 : "x") as string);',
      unknown,
      'file.ts',
    ),

    invalid(
      'const buffer = new Buffer([0x62, 0x75, 0x66, 0x66, 0x65, 0x72])',
      from,
    ),
    invalid('const buffer = new Buffer([0x62, bar])', from),
    invalid('const array = [0x62];\nconst buffer = new Buffer(array);', from),
    invalid(
      'const arrayBuffer = new ArrayBuffer(10);\nconst buffer = new Buffer(arrayBuffer);',
      unknown,
    ),
    invalid(
      'const arrayBuffer = new ArrayBuffer(10);\nconst buffer = new Buffer(arrayBuffer, 0, );',
      from,
    ),
    invalid(
      'const arrayBuffer = new ArrayBuffer(10);\nconst buffer = new Buffer(arrayBuffer, 0, 2);',
      from,
    ),

    invalid('const buffer = new Buffer(10);', alloc),
    invalid('const size = 10;\nconst buffer = new Buffer(size);', alloc),
    invalid('new Buffer(foo.length)', alloc),
    invalid('new Buffer(Math.min(foo, bar))', alloc),
    invalid('new Buffer(Math.unknown())', unknown),
    invalid('new Buffer(foo?.length)', unknown),
    invalid('new Buffer(-1)', alloc),
    invalid('new Buffer(1 + 2)', alloc),
    invalid('new Buffer(Number.parseInt(value))', alloc),
    invalid('new Buffer(value as number)', alloc, 'file.ts'),
    invalid('new Buffer("a" + "b")', from),

    invalid('const buffer = new Buffer("string");', from),
    invalid(
      'const buffer = new Buffer("7468697320697320612074c3a97374", "hex")',
      from,
    ),
    invalid(
      'const string = "string";\nconst buffer = new Buffer(string);',
      from,
    ),
    invalid('const buffer = new Buffer(`${unknown}`)', from),

    invalid('const buffer = new (Buffer)(unknown)', unknown),
    invalid('const buffer = new Buffer(unknown, 2)', from),
    invalid('const buffer = new Buffer(...unknown)', unknown),

    invalid('() => {\n\treturn new // 1\n\t\tBuffer();\n}', from),
    invalid('() => {\n\treturn (\n\t\tnew // 2\n\t\t\tBuffer()\n\t);\n}', from),
    invalid('() => {\n\treturn new // 3\n\t\t(Buffer);\n}', from),
    invalid('() => {\n\treturn new // 4\n\t\tBuffer;\n}', from),
    invalid('() => {\n\treturn (\n\t\tnew // 5\n\t\t\tBuffer\n\t);\n}', from),
    invalid('() => {\n\treturn (\n\t\tnew // 6\n\t\t\t(Buffer)\n\t);\n}', from),
    invalid('const buffer = new /* comment */ Buffer()', from),
    invalid('const buffer = new /* comment */ Buffer', from),

    invalid('new Buffer(input, encoding);', from),
  ],
});

const runtimeFile = path.resolve(
  import.meta.dirname,
  'no-new-buffer.runtime.js',
);
const configFile = path.resolve(import.meta.dirname, '../rslint.config.mjs');
let runtimeConfig: ReturnType<typeof buildConfigForSettings> | undefined;

const getRuntimeConfig = () =>
  (runtimeConfig ??= buildConfigForSettings(configFile, undefined));

const lintRuntime = async (code: string, fix = false) => {
  const { config, configDirectory } = await getRuntimeConfig();
  return lint({
    workingDirectory: process.cwd(),
    configDirectory,
    config: [...config, { rules: { 'unicorn/no-new-buffer': 'error' } }],
    fileContents: { [runtimeFile]: code },
    fix,
  });
};

const applySuggestion = (
  source: string,
  fixes: { startPos: number; endPos: number; text: string }[],
) =>
  [...fixes]
    .sort((left, right) => right.startPos - left.startPos)
    .reduce(
      (output, fix) =>
        output.slice(0, fix.startPos) + fix.text + output.slice(fix.endPos),
      source,
    );

describe('unicorn/no-new-buffer runtime integration', () => {
  test('reports the exact range and converges an automatic fix', async () => {
    const source = 'new Buffer(foo.length);';
    const before = await lintRuntime(source);
    expect(
      before.diagnostics.map(({ ruleName, messageId, message, range }) => ({
        ruleName,
        messageId,
        message,
        range,
      })),
    ).toEqual([
      {
        ruleName: 'unicorn/no-new-buffer',
        messageId: 'error',
        message: alloc,
        range: {
          start: { line: 1, column: 1 },
          end: { line: 1, column: 23 },
        },
      },
    ]);

    const fixed = await lintRuntime(source, true);
    const output = Object.values(fixed.output ?? {})[0];
    expect(output).toBe('Buffer.alloc(foo.length);');
    await expect(lintRuntime(output ?? source)).resolves.toMatchObject({
      diagnostics: [],
    });
  });

  test('applies each unknown-case suggestion and converges', async () => {
    const source = 'new Buffer(unknown);';
    const result = await lintRuntime(source);
    const [diagnostic] = result.diagnostics;
    expect({
      ruleName: diagnostic.ruleName,
      messageId: diagnostic.messageId,
      message: diagnostic.message,
      range: diagnostic.range,
    }).toEqual({
      ruleName: 'unicorn/no-new-buffer',
      messageId: 'error-unknown',
      message: unknown,
      range: {
        start: { line: 1, column: 1 },
        end: { line: 1, column: 20 },
      },
    });
    expect(
      diagnostic.suggestions?.map(({ messageId, message }) => ({
        messageId,
        message,
      })),
    ).toEqual([
      { messageId: 'suggestion', message: 'Switch to `Buffer.from()`.' },
      { messageId: 'suggestion', message: 'Switch to `Buffer.alloc()`.' },
    ]);

    for (const suggestion of diagnostic.suggestions ?? []) {
      const output = applySuggestion(source, suggestion.fixes ?? []);
      await expect(lintRuntime(output)).resolves.toMatchObject({
        diagnostics: [],
      });
    }
  });

  test('preserves multiline generator yield operands after fixing', async () => {
    const cases = [
      {
        name: 'yield',
        source: 'function* values() {\n\tyield new // yield\n\t\tBuffer(1);\n}',
        output:
          'function* values() {\n\tyield ( // yield\n\t\tBuffer.alloc(1));\n}',
        buffer: { alloc: (size: number) => ({ size }) },
        evaluate: 'return [...values()];',
        expected: [{ size: 1 }],
      },
      {
        name: 'yield*',
        source:
          'function* values() {\n\tyield* new // yield-star\n\t\tBuffer([1, 2]);\n}',
        output:
          'function* values() {\n\tyield* ( // yield-star\n\t\tBuffer.from([1, 2]));\n}',
        buffer: { from: (values: number[]) => values },
        evaluate: 'return [...values()];',
        expected: [1, 2],
      },
      {
        name: 'return with line separator',
        source: 'function value() {\n\treturn new\u2028\tBuffer(1);\n}',
        output: 'function value() {\n\treturn ( Buffer.alloc(1));\n}',
        buffer: { alloc: (size: number) => ({ size }) },
        evaluate: 'return value();',
        expected: { size: 1 },
      },
      {
        name: 'throw with paragraph separator',
        source:
          'function value() {\n\ttry {\n\t\tthrow new\u2029\tBuffer(1);\n\t} catch (error) {\n\t\treturn error;\n\t}\n}',
        output:
          'function value() {\n\ttry {\n\t\tthrow ( Buffer.alloc(1));\n\t} catch (error) {\n\t\treturn error;\n\t}\n}',
        buffer: { alloc: (size: number) => ({ size }) },
        evaluate: 'return value();',
        expected: { size: 1 },
      },
    ];

    for (const { name, source, output, buffer, evaluate, expected } of cases) {
      const fixed = await lintRuntime(source, true);
      const actual = Object.values(fixed.output ?? {})[0];
      expect(actual, name).toBe(output);
      await expect(lintRuntime(actual ?? source), name).resolves.toMatchObject({
        diagnostics: [],
      });
      expect(
        new Function('Buffer', `${actual ?? source}\n${evaluate}`)(buffer),
        name,
      ).toEqual(expected);
    }
  });

  test('covers numeric helpers, static control flow, and mutable updates', async () => {
    const automaticCases = [
      {
        name: 'literal conditional',
        source: 'new Buffer(true ? 1 : value);',
        output: 'Buffer.alloc(true ? 1 : value);',
      },
      {
        name: 'const conditional',
        source: 'const enabled = true; new Buffer(enabled ? 1 : value);',
        output: 'const enabled = true; Buffer.alloc(enabled ? 1 : value);',
      },
      {
        name: 'const logical expression',
        source: 'const size = 1; new Buffer(size || value);',
        output: 'const size = 1; Buffer.alloc(size || value);',
      },
      {
        name: 'assignment expression',
        source: 'let value; new Buffer(value = 1);',
        output: 'let value; Buffer.alloc(value = 1);',
      },
      {
        name: 'string numeric method',
        source: 'new Buffer("x".indexOf(value));',
        output: 'Buffer.alloc("x".indexOf(value));',
      },
      {
        name: 'shadowed Math',
        source: 'function f(Math) { new Buffer(Math.min(x, y)); }',
        output: 'function f(Math) { Buffer.alloc(Math.min(x, y)); }',
      },
      {
        name: 'shadowed Number',
        source: 'function f(Number) { new Buffer(Number(value)); }',
        output: 'function f(Number) { Buffer.alloc(Number(value)); }',
      },
      {
        name: 'literal logical expression',
        source: 'new Buffer(1 || value);',
        output: 'Buffer.alloc(1 || value);',
      },
      {
        name: 'const number binding',
        source: 'const value = 1; new Buffer(value);',
        output: 'const value = 1; Buffer.alloc(value);',
      },
      {
        name: 'const string binding',
        source: 'const value = "x"; new Buffer(value);',
        output: 'const value = "x"; Buffer.from(value);',
        message: from,
      },
      {
        name: 'const array binding',
        source: 'const value = [1]; new Buffer(value);',
        output: 'const value = [1]; Buffer.from(value);',
        message: from,
      },
      {
        name: 'static equality conditional',
        source: 'new Buffer((1 === 1) ? 1 : value);',
        output: 'Buffer.alloc((1 === 1) ? 1 : value);',
      },
      {
        name: 'undefined nullish expression',
        source: 'new Buffer(undefined ?? 1);',
        output: 'Buffer.alloc(undefined ?? 1);',
      },
      {
        name: 'object logical expression',
        source: 'new Buffer({} && 1);',
        output: 'Buffer.alloc({} && 1);',
      },
    ];

    for (const { name, source, output, message = alloc } of automaticCases) {
      const context = `${name}: ${source}`;
      const before = await lintRuntime(source);
      expect(
        before.diagnostics.map(({ ruleName, messageId, message }) => ({
          ruleName,
          messageId,
          message,
        })),
        context,
      ).toEqual([
        {
          ruleName: 'unicorn/no-new-buffer',
          messageId: 'error',
          message,
        },
      ]);
      const fixed = await lintRuntime(source, true);
      expect(Object.values(fixed.output ?? {})[0], context).toBe(output);
      await expect(lintRuntime(output), context).resolves.toMatchObject({
        diagnostics: [],
      });
    }

    const mutable = 'let value = 1; new Buffer(value++);';
    for (const source of [
      mutable,
      'let value = 1; new Buffer(value);',
      'var value = "x"; new Buffer(value);',
    ]) {
      const result = await lintRuntime(source);
      expect(result.diagnostics[0]).toMatchObject({
        ruleName: 'unicorn/no-new-buffer',
        messageId: 'error-unknown',
        message: unknown,
      });
      for (const suggestion of result.diagnostics[0]?.suggestions ?? []) {
        const output = applySuggestion(source, suggestion.fixes ?? []);
        await expect(lintRuntime(output)).resolves.toMatchObject({
          diagnostics: [],
        });
      }
    }
  });
});

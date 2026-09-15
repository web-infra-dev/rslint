import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const valid = (code: string, filename = 'file.js') => ({ code, filename });
const invalid = (
  code: string,
  output: string,
  method: 'trimLeft' | 'trimRight',
  replacement: 'trimStart' | 'trimEnd',
  filename = 'file.js',
) => ({
  code,
  filename,
  output,
  errors: [
    {
      messageId: 'prefer-string-trim-start-end',
      message: `Prefer \`String#${replacement}()\` over \`String#${method}()\`.`,
      data: { method, replacement },
    },
  ],
});

ruleTester.run('prefer-string-trim-start-end', null as never, {
  valid: [
    valid('function f(foo: number[]) { foo.trimLeft(); }', 'file.ts'),
    valid('foo.trimStart()'),
    valid('foo.trimStart?.()'),
    valid('foo.trimEnd()'),
    valid('new foo.trimLeft();'),
    valid('trimLeft();'),
    valid("foo['trimLeft']();"),
    valid('foo[trimLeft]();'),
    valid('foo.bar();'),
    valid('foo.trimLeft(extra);'),
    valid('foo.trimLeft(...argumentsArray)'),
    valid('foo.bar(trimLeft)'),
    valid('foo.bar(foo.trimLeft)'),
    valid('trimLeft.foo()'),
    valid('foo.trimLeft.bar()'),
  ],
  invalid: [
    invalid(
      'function f(foo: string) { foo.trimLeft(); }',
      'function f(foo: string) { foo.trimStart(); }',
      'trimLeft',
      'trimStart',
      'file.ts',
    ),
    invalid('foo.trimLeft()', 'foo.trimStart()', 'trimLeft', 'trimStart'),
    invalid('foo.trimRight()', 'foo.trimEnd()', 'trimRight', 'trimEnd'),
    invalid(
      'trimLeft.trimRight()',
      'trimLeft.trimEnd()',
      'trimRight',
      'trimEnd',
    ),
    invalid(
      'foo.trimLeft.trimRight()',
      'foo.trimLeft.trimEnd()',
      'trimRight',
      'trimEnd',
    ),
    invalid('"foo".trimLeft()', '"foo".trimStart()', 'trimLeft', 'trimStart'),
    invalid(
      'foo\n\t// comment\n\t.trimRight/* comment */(\n\t\t/* comment */\n\t)',
      'foo\n\t// comment\n\t.trimEnd/* comment */(\n\t\t/* comment */\n\t)',
      'trimRight',
      'trimEnd',
    ),
    invalid('foo?.trimLeft()', 'foo?.trimStart()', 'trimLeft', 'trimStart'),
  ],
});

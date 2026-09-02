import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const valid = (code: string, filename = 'file.js') => ({ code, filename });
const invalid = (code: string, output: string, filename = 'file.js') => ({
  code,
  filename,
  output,
  errors: [
    {
      messageId: 'no-unnecessary-array-flat-depth',
      message: 'Passing `1` as the `depth` argument is unnecessary.',
    },
  ],
});

ruleTester.run('no-unnecessary-array-flat-depth', null as never, {
  valid: [
    valid(
      'function f(foo: {flat(depth: number): void}) { foo.flat(1); }',
      'file.ts',
    ),
    valid('foo.flat()'),
    valid('foo.flat?.(1)'),
    valid('foo?.flat()'),
    valid('foo.flat(1, extra)'),
    valid('flat(1)'),
    valid('new foo.flat(1)'),
    valid('const ONE = 1; foo.flat(ONE)'),
    valid('foo.notFlat(1)'),
  ],
  invalid: [
    invalid('foo.flat(1)', 'foo.flat()'),
    invalid('foo.flat(1.0)', 'foo.flat()'),
    invalid('foo.flat(0b01)', 'foo.flat()'),
    invalid('foo?.flat(1)', 'foo?.flat()'),
    invalid(
      'function f(foo: number[][]) { foo.flat(1); }',
      'function f(foo: number[][]) { foo.flat(); }',
      'file.ts',
    ),
  ],
});

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const valid = (code: string) => ({ code, filename: 'file.mjs' });

const invalid = (code: string) => ({
  code,
  filename: 'file.mjs',
  errors: [{ message: 'Magic number as depth is not allowed.' }],
});

ruleTester.run('no-magic-array-flat-depth', null as never, {
  valid: [
    // depth is 1 (the default).
    valid('array.flat(1)'),
    valid('array.flat(1.0)'),
    valid('array.flat(0x01)'),

    // depth is not a numeric literal.
    valid('array.flat(unknown)'),
    valid('array.flat(Number.POSITIVE_INFINITY)'),
    valid('array.flat(Infinity)'),

    // comments around the depth explain the magic number.
    valid('array.flat(/* explanation */2)'),
    valid('array.flat(2/* explanation */)'),

    // no argument or wrong number of arguments.
    valid('array.flat()'),
    valid('array.flat(2, extraArgument)'),

    // not a CallExpression of `.flat`.
    valid('new array.flat(2)'),
    valid('array.flat?.(2)'),
    valid('array.notFlat(2)'),
    valid('flat(2)'),

    // Statically known non-array receivers.
    valid('Math.abs(-2).flat(2)'),
    valid('Array.isArray([]).flat(2)'),
    valid("parseInt('2', 10).flat(2)"),
    valid('Number(1).flat(2)'),
    valid('const value = 1; value.flat(2)'),
  ],
  invalid: [
    invalid('array.flat(2)'),
    invalid('array?.flat(2)'),
    invalid('array.flat(99,)'),
    invalid('array.flat(0b10,)'),

    // Known arrays, unknown/throwing coercions, and source-only recursion.
    invalid('Array.of(1).flat(2)'),
    invalid('Object.freeze([]).flat(2)'),
    invalid('Number(value).flat(2)'),
    invalid('BigInt().flat(2)'),
    invalid('const value = Math.PI; value.flat(2)'),
    invalid('(flag ? Math.PI : 1).flat(2)'),

    // A comment before the real opening parenthesis is outside the arguments.
    invalid('array.flat /* explanation */ (2)'),
  ],
});

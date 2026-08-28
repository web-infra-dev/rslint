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
  ],
  invalid: [
    invalid('array.flat(2)'),
    invalid('array?.flat(2)'),
    invalid('array.flat(99,)'),
    invalid('array.flat(0b10,)'),
  ],
});

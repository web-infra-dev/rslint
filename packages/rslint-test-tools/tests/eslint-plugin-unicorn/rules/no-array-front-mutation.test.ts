import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const valid = (code: string, filename = 'file.mjs') => ({ code, filename });
const invalid = (
  code: string,
  method: 'shift' | 'unshift',
  filename = 'file.mjs',
) => ({
  code,
  filename,
  errors: [
    {
      messageId: 'no-array-front-mutation',
      message: `Avoid front-of-array mutation with \`Array#${method}()\`.`,
      data: { method },
    },
  ],
});

ruleTester.run('no-array-front-mutation', null as never, {
  valid: [
    valid('function f(foo: Set<number>) { foo.shift(); }', 'file.ts'),
    valid('function f(foo: Set<number>) { foo.unshift(value); }', 'file.ts'),
    valid('array.shift'),
    valid('array.unshift'),
    valid('array.shift?.()'),
    valid('array.unshift?.(value)'),
    valid('array?.shift?.()'),
    valid('array?.unshift?.(value)'),
    valid('array["shift"]()'),
    valid('array["unshift"](value)'),
    valid('shift(array)'),
    valid('unshift(array, value)'),
    valid('Array.prototype.shift.call(array)'),
    valid('Array.prototype.unshift.call(array, value)'),
    valid('stream.unshift(chunk)'),
    valid('this.unshift(chunk)'),
    valid('this.stream.unshift(chunk)'),
    valid('process.stdin.unshift(chunk)'),
    valid('process.stdout.unshift(chunk)'),
    valid('process.stderr.unshift(chunk)'),
  ],
  invalid: [
    invalid('array.shift()', 'shift'),
    invalid('array.shift(extraArgument)', 'shift'),
    invalid('array?.shift()', 'shift'),
    invalid('array.unshift()', 'unshift'),
    invalid('array.unshift(value)', 'unshift'),
    invalid('array.unshift(...values)', 'unshift'),
    invalid('array?.unshift(value)', 'unshift'),
    invalid('stream.shift()', 'shift'),
    invalid('const item = array.shift()', 'shift'),
    invalid('const length = array.unshift(value)', 'unshift'),
    invalid('function getItem() { return array.shift(); }', 'shift'),
    invalid('while (array.shift()) {}', 'shift'),
    invalid('if (array.unshift(value)) {}', 'unshift'),
    invalid('for (; array.shift(); ) {}', 'shift'),
    invalid('(array as string[]).shift()', 'shift', 'file.ts'),
    invalid('array!.unshift(value)', 'unshift', 'file.ts'),
  ],
});

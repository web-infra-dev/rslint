import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const valid = (code: string) => ({ code });
const invalid = (code: string) => ({
  code,
  errors: [
    {
      messageId: 'no-array-from-fill',
      message:
        'Use the `Array.from(…, mapFunction)` argument instead of chaining `.fill()`.',
    },
  ],
});

ruleTester.run('no-array-from-fill', null as never, {
  valid: [
    valid('Array.from({length: 3})'),
    valid('Array.from({length: 3}, (_, index) => index)'),
    valid('Array.from({length: 3}).map((_, index) => index)'),
    valid('Array.from(items).fill(0)'),
    valid('Array.from({length: 3, 0: "value"}).fill(0)'),
    valid('Array.from({...length}).fill(0)'),
    valid('Array.from({["length"]: 3}).fill(0)'),
    valid('Array.from({length: 3}).fill(0, 1)'),
    valid('Array.from({length: 3}).fill(0, 1, 2)'),
    valid('Array.from({length: 3}).fill(...value)'),
    valid('Array.from?.({length: 3}).fill(0)'),
    valid('Array.from({length: 3})?.fill(0)'),
    valid('Array.from({length: 3}).fill?.(0)'),
    valid('NotArray.from({length: 3}).fill(0)'),
    valid('Array.notFrom({length: 3}).fill(0)'),
    valid('Array.from({length: 3}).slice().fill(0)'),
    valid(
      'const Array = {from() { return {fill() { return {map() {}}; }}; }}; Array.from({length: 3}).fill().map((_, index) => index)',
    ),
    valid(
      'function unicorn(Array) { return Array.from({length: 3}).fill(0); }',
    ),
  ],
  invalid: [
    invalid('Array.from({length: 3}).fill(0)'),
    invalid('Array.from({length: 3}).fill()'),
    invalid('Array.from({length}).fill(null)'),
    invalid('Array.from({"length": 3}).fill(0)'),
    invalid('Array.from({length: 3}).fill({})'),
    invalid('Array.from({length: 3}).fill(0).map((_, index) => index)'),
    invalid('Array.from({length: 3}).fill().map((value, index) => index)'),
    invalid('Array.from({length: 3}).fill(0).flatMap((_, index) => [index])'),
    invalid('Array.from({length: 3}).fill().flatMap(value => [value])'),
    invalid('Array.from({length: 3}).fill(0).filter(Boolean)'),
    invalid(
      'Array.from({length: 3})\n\t.fill(0)\n\t.map((_, index) => index);',
    ),
    invalid(
      'Array.from(\n\t{length: 3}\n)\n\t.fill(0)\n\t.map((_, index) => index);',
    ),
  ],
});

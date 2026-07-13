import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const message = 'Missing the digits argument.';
const valid = (code: string) => ({ code, filename: 'file.js' });
const invalid = (code: string) => ({
  code,
  filename: 'file.js',
  errors: [{ message }],
});

ruleTester.run('require-number-to-fixed-digits-argument', null as never, {
  valid: [
    valid('number.toFixed(0)'),
    valid('number.toFixed(...[])'),
    valid('number.toFixed(2)'),
    valid('number.toFixed(1,2,3)'),
    valid('number[toFixed]()'),
    valid('number["toFixed"]()'),
    valid('number.toFixed?.()'),
    valid('number.notToFixed();'),

    // `callee` is a `NewExpression`.
    valid('new BigNumber(1).toFixed()'),
    valid('new Number(1).toFixed()'),
  ],
  invalid: [
    invalid('const string = number.toFixed();'),
    invalid('const string = number?.toFixed() ?? "";'),
    invalid('const string = number.toFixed( /* comment */ );'),
    invalid('Number(1).toFixed()'),

    // False positive cases.
    invalid(
      'const bigNumber = new BigNumber(1); const string = bigNumber.toFixed();',
    ),
  ],
});

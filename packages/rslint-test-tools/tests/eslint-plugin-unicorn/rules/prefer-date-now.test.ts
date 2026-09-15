import { RuleTester } from '../rule-tester';

const dateMessage = 'Prefer `Date.now()` over `new Date()`.';
const invalid = (
  code: string,
  output: string,
  messageId = 'prefer-date',
  message = dateMessage,
  count = 1,
) => ({
  code,
  filename: 'file.js',
  output,
  errors: Array.from({ length: count }, () => ({ messageId, message })),
});
const method = (code: string, output: string, name: string) =>
  invalid(
    code,
    output,
    'prefer-date-now-over-methods',
    `Prefer \`Date.now()\` over \`Date#${name}()\`.`,
  );
const number = (code: string, output: string) =>
  invalid(
    code,
    output,
    'prefer-date-now-over-number-data-object',
    'Prefer `Date.now()` over `Number(new Date())`.',
  );

// Complete upstream suite and documentation examples:
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/test/prefer-date-now.js
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/docs/rules/prefer-date-now.md
new RuleTester().run('prefer-date-now', null as never, {
  valid: [
    'const ts = Date.now()',
    // Constructor shape and arguments.
    '+Date()',
    '+ Date',
    '+ new window.Date()',
    '+ new Moments()',
    '+ new Date(0)',
    '+ new Date(...[])',
    // Method calls.
    'new Date.getTime()',
    'valueOf()',
    'new Date()[getTime]()',
    'new Date()["valueOf"]()',
    'new Date().notListed(0)',
    'new Date().getTime(0)',
    'new Date().valueOf(...[])',
    // Number and BigInt calls.
    'new Number(new Date())',
    'window.BigInt(new Date())',
    'toNumber(new Date())',
    'BigInt()',
    'Number(new Date(), extraArgument)',
    'BigInt([...new Date()])',
    // Unary, assignment and binary expressions.
    'throw new Date()',
    'typeof new Date()',
    'const foo = () => {return new Date()}',
    'foo += new Date()',
    'function * foo() {yield new Date()}',
    'new Date() + new Date()',
    'foo = new Date() | 0',
    'foo &= new Date()',
    'foo = new Date() >> 0',
    // Documentation.
    'const foo = Date.now();',
    'const foo = Date.now() * 2;',
  ].map((code) => ({ code, filename: 'file.js' })),
  invalid: [
    method(
      'const ts = new Date().getTime();',
      'const ts = Date.now();',
      'getTime',
    ),
    method(
      'const ts = (new Date).getTime();',
      'const ts = Date.now();',
      'getTime',
    ),
    method(
      'const ts = (new Date()).getTime();',
      'const ts = Date.now();',
      'getTime',
    ),
    method(
      'const ts = new Date().valueOf();',
      'const ts = Date.now();',
      'valueOf',
    ),
    method(
      'const ts = (new Date).valueOf();',
      'const ts = Date.now();',
      'valueOf',
    ),
    method(
      'const ts = (new Date()).valueOf();',
      'const ts = Date.now();',
      'valueOf',
    ),
    number(
      'const ts = /* 1 */ Number(/* 2 */ new /* 3 */ Date( /* 4 */ ) /* 5 */) /* 6 */',
      'const ts = /* 1 */ Date.now() /* 6 */',
    ),
    invalid(
      'const tsBigInt = /* 1 */ BigInt(/* 2 */ new /* 3 */ Date( /* 4 */ ) /* 5 */) /* 6 */',
      'const tsBigInt = /* 1 */ BigInt(/* 2 */ Date.now() /* 5 */) /* 6 */',
    ),
    invalid('const ts = + /* 1 */ new Date;', 'const ts = Date.now();'),
    invalid(
      'const ts = - /* 1 */ new Date();',
      'const ts = - /* 1 */ Date.now();',
    ),
    invalid('const ts = +(new Date());', 'const ts = Date.now();'),
    invalid('const ts = -(new Date());', 'const ts = -(Date.now());'),
    invalid('const ts = new Date() - 0', 'const ts = Date.now() - 0'),
    invalid('const foo = bar - new Date', 'const foo = bar - Date.now()'),
    invalid('const foo = new Date() * bar', 'const foo = Date.now() * bar'),
    invalid('const ts = new Date() / 1', 'const ts = Date.now() / 1'),
    invalid(
      'const ts = new Date() % Infinity',
      'const ts = Date.now() % Infinity',
    ),
    invalid('const ts = new Date() ** 1', 'const ts = Date.now() ** 1'),
    invalid(
      'const zero = (new Date(/* 1 */) /* 2 */) /* 3 */ - /* 4 */new Date',
      'const zero = (Date.now() /* 2 */) /* 3 */ - /* 4 */Date.now()',
      'prefer-date',
      dateMessage,
      2,
    ),
    invalid('foo -= new Date()', 'foo -= Date.now()'),
    invalid('foo *= new Date()', 'foo *= Date.now()'),
    invalid('foo /= new Date', 'foo /= Date.now()'),
    invalid('foo %= new Date()', 'foo %= Date.now()'),
    invalid('foo **= new Date()', 'foo **= Date.now()'),
    invalid(
      'function foo(){return+new Date}',
      'function foo(){return Date.now()}',
    ),
    invalid(
      'function foo(){return-new Date}',
      'function foo(){return-Date.now()}',
    ),
    // Documentation.
    method(
      'const foo = new Date().getTime();',
      'const foo = Date.now();',
      'getTime',
    ),
    method(
      'const foo = new Date().valueOf();',
      'const foo = Date.now();',
      'valueOf',
    ),
    invalid('const foo = +new Date;', 'const foo = Date.now();'),
    number('const foo = Number(new Date());', 'const foo = Date.now();'),
    invalid('const foo = new Date() * 2;', 'const foo = Date.now() * 2;'),
  ],
});

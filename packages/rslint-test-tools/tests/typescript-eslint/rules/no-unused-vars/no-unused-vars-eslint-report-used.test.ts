// The following tests are adapted from the tests in eslint.
// Original Code: https://github.com/eslint/eslint/blob/eb76282e0a2db8aa10a3d5659f5f9237d9729121/tests/lib/rules/no-unused-vars.js
// License      : https://github.com/eslint/eslint/blob/eb76282e0a2db8aa10a3d5659f5f9237d9729121/LICENSE

import {
  assignedError,
  definedError,
  ruleTester,
  usedIgnoredError,
} from './eslint-test-helpers';

ruleTester.run('no-unused-vars', {
  invalid: [
    // https://github.com/eslint/eslint/issues/17568
    {
      code: `
const _a = 5;
const _b = _a + 5;
      `,
      errors: [usedIgnoredError('_a', '. Used vars must not match /^_/u')],
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
      options: [
        { args: 'all', reportUsedIgnorePattern: true, varsIgnorePattern: '^_' },
      ],
    },
    {
      code: `
const _a = 42;
foo(() => _a);
      `,
      errors: [usedIgnoredError('_a', '. Used vars must not match /^_/u')],
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
      options: [
        { args: 'all', reportUsedIgnorePattern: true, varsIgnorePattern: '^_' },
      ],
    },
    {
      code: `
(function foo(_a) {
  return _a + 5;
})(5);
      `,
      errors: [usedIgnoredError('_a', '. Used args must not match /^_/u')],
      options: [
        { args: 'all', argsIgnorePattern: '^_', reportUsedIgnorePattern: true },
      ],
    },
    {
      code: `
const [a, _b] = items;
console.log(a + _b);
      `,
      errors: [
        usedIgnoredError(
          '_b',
          '. Used elements of array destructuring must not match /^_/u',
        ),
      ],
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
      options: [
        { destructuredArrayIgnorePattern: '^_', reportUsedIgnorePattern: true },
      ],
    },
    {
      // TODO: ESLint considers assignment-side destructuring `[_x] = arr` as matching
      code: `
declare const arr: any[];
declare function foo(x: any): void;
let _x;
[_x] = arr;
foo(_x);
      `,
      errors: [
        usedIgnoredError(
          '_x',
          '. Used elements of array destructuring must not match /^_/u',
        ),
      ],
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
      options: [
        {
          destructuredArrayIgnorePattern: '^_',
          reportUsedIgnorePattern: true,
          varsIgnorePattern: '[iI]gnored',
        },
      ],
    },
    {
      code: `
const [ignored] = arr;
foo(ignored);
      `,
      errors: [
        usedIgnoredError('ignored', '. Used vars must not match /[iI]gnored/u'),
      ],
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
      options: [
        {
          destructuredArrayIgnorePattern: '^_',
          reportUsedIgnorePattern: true,
          varsIgnorePattern: '[iI]gnored',
        },
      ],
    },
    {
      code: `
try {
} catch (_err) {
  console.error(_err);
}
      `,
      errors: [
        usedIgnoredError('_err', '. Used caught errors must not match /^_/u'),
      ],
      options: [
        {
          caughtErrors: 'all',
          caughtErrorsIgnorePattern: '^_',
          reportUsedIgnorePattern: true,
        },
      ],
    },
    {
      code: `
try {
} catch (_) {
  _ = 'foo';
}
      `,
      errors: [
        assignedError('_', '. Allowed unused caught errors must match /foo/u'),
      ],
      options: [{ caughtErrorsIgnorePattern: 'foo' }],
    },
    {
      code: `
try {
} catch (_) {
  _ = 'foo';
}
      `,
      errors: [
        assignedError(
          '_',
          '. Allowed unused caught errors must match /ignored/u',
        ),
      ],
      options: [
        {
          caughtErrorsIgnorePattern: 'ignored',
          varsIgnorePattern: '_',
        },
      ],
    },
    {
      code: `
try {
} catch ({ message, errors: [firstError] }) {}
      `,
      errors: [
        {
          ...definedError(
            'message',
            '. Allowed unused caught errors must match /foo/u',
          ),
        },
        {
          ...definedError(
            'firstError',
            '. Allowed unused caught errors must match /foo/u',
          ),
        },
      ],
      options: [{ caughtErrorsIgnorePattern: 'foo' }],
    },
    {
      code: `
try {
} catch ({ stack: $ }) {
  $ = 'Something broke: ' + $;
}
      `,
      errors: [
        {
          ...assignedError(
            '$',
            '. Allowed unused caught errors must match /\\w/u',
          ),
        },
      ],
      options: [{ caughtErrorsIgnorePattern: '\\w' }],
    },
    {
      code: `
_ => {
  _ = _ + 1;
};
      `,
      errors: [
        assignedError('_', '. Allowed unused args must match /ignored/u'),
      ],
      options: [
        {
          argsIgnorePattern: 'ignored',
          varsIgnorePattern: '_',
        },
      ],
    },
  ],
  valid: [
    // https://github.com/eslint/eslint/issues/17568
    {
      code: `
const a = 5;
const _c = a + 5;
      `,
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
      options: [
        { args: 'all', reportUsedIgnorePattern: true, varsIgnorePattern: '^_' },
      ],
    },
    {
      code: `
(function foo(a, _b) {
  return a + 5;
})(5);
      `,
      options: [
        { args: 'all', argsIgnorePattern: '^_', reportUsedIgnorePattern: true },
      ],
    },
    {
      code: `
const [a, _b, c] = items;
console.log(a + c);
      `,
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
      options: [
        { destructuredArrayIgnorePattern: '^_', reportUsedIgnorePattern: true },
      ],
    },
  ],
});

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('valid-typeof', {
  valid: [
    // All valid typeof comparison values
    "typeof foo === 'string'",
    "typeof foo === 'object'",
    "typeof foo === 'function'",
    "typeof foo === 'undefined'",
    "typeof foo === 'boolean'",
    "typeof foo === 'number'",
    "typeof foo === 'bigint'",
    "typeof foo === 'symbol'",

    // Reversed operands
    "'string' === typeof foo",
    "'object' === typeof foo",

    // typeof compared to typeof (always valid)
    'typeof foo === typeof bar',

    // Non-equality operators are not checked
    "typeof foo > 'string'",

    // Without requireStringLiterals, non-string comparisons are OK
    'typeof foo === baz',
    'typeof foo === Object',

    // Not a comparison
    'var x = typeof foo',
    'typeof foo',

    // With requireStringLiterals: valid cases still valid
    {
      code: "typeof foo === 'string'",
      options: { requireStringLiterals: true },
    },
    {
      code: 'typeof foo === typeof bar',
      options: { requireStringLiterals: true },
    },
    {
      code: "'undefined' === typeof foo",
      options: { requireStringLiterals: true },
    },

    // != and !== with valid strings
    "typeof foo !== 'string'",
    "typeof foo != 'function'",
    "typeof foo == 'number'",

    // Static template literals with valid values
    'typeof foo === `string`',
    'typeof foo === `object`',
    'typeof foo === `undefined`',

    // Parenthesized expressions with valid values
    '(typeof foo) === "string"',
    'typeof foo === ("string")',
    '((typeof foo)) === "string"',
    '(typeof foo) === ("string")',

    // Locally shadowed undefined — not reported without requireStringLiterals
    'function f(undefined: string) { typeof foo === undefined }',
    '{ const undefined = "test"; typeof foo === undefined }',
  ],
  invalid: [
    // Invalid typeof comparison value (misspelled)
    {
      code: "typeof foo === 'strnig'",
      errors: [{ messageId: 'invalidValue' }],
    },
    // Reversed operands with invalid value
    {
      code: "'strnig' === typeof foo",
      errors: [{ messageId: 'invalidValue' }],
    },
    // !== with invalid value
    {
      code: "typeof foo !== 'strnig'",
      errors: [{ messageId: 'invalidValue' }],
    },
    // == with invalid value
    {
      code: "typeof foo == 'strnig'",
      errors: [{ messageId: 'invalidValue' }],
    },
    // != with invalid value
    {
      code: "typeof foo != 'strnig'",
      errors: [{ messageId: 'invalidValue' }],
    },
    // Bare undefined identifier without requireStringLiterals → invalidValue
    {
      code: 'typeof foo === undefined',
      errors: [{ messageId: 'invalidValue' }],
    },
    // Bare undefined identifier with requireStringLiterals → notString
    {
      code: 'typeof foo === undefined',
      options: { requireStringLiterals: true },
      errors: [{ messageId: 'notString' }],
    },
    // Non-string identifier with requireStringLiterals → notString
    {
      code: 'typeof foo === Object',
      options: { requireStringLiterals: true },
      errors: [{ messageId: 'notString' }],
    },
    // Completely invalid string
    {
      code: "typeof foo === 'foobar'",
      errors: [{ messageId: 'invalidValue' }],
    },
    // Empty string
    {
      code: "typeof foo === ''",
      errors: [{ messageId: 'invalidValue' }],
    },
    // Static template literal with invalid value
    {
      code: 'typeof foo === `strnig`',
      errors: [{ messageId: 'invalidValue' }],
    },
    // Static template literal with completely invalid value
    {
      code: 'typeof foo === `foobar`',
      errors: [{ messageId: 'invalidValue' }],
    },
    // Non-string literals: null, number, boolean, regex
    {
      code: 'typeof foo === null',
      errors: [{ messageId: 'invalidValue' }],
    },
    {
      code: 'typeof foo === 42',
      errors: [{ messageId: 'invalidValue' }],
    },
    {
      code: 'typeof foo === true',
      errors: [{ messageId: 'invalidValue' }],
    },
    {
      code: 'typeof foo === false',
      errors: [{ messageId: 'invalidValue' }],
    },
    // Shadowed undefined with requireStringLiterals → notString (no suggestion)
    {
      code: 'function f(undefined: string) { typeof foo === undefined }',
      options: { requireStringLiterals: true },
      errors: [{ messageId: 'notString' }],
    },
    // Parenthesized expressions with invalid values
    {
      code: '(typeof foo) === "strnig"',
      errors: [{ messageId: 'invalidValue' }],
    },
    {
      code: 'typeof foo === ("strnig")',
      errors: [{ messageId: 'invalidValue' }],
    },
    {
      code: '((typeof foo)) === "strnig"',
      errors: [{ messageId: 'invalidValue' }],
    },
    {
      code: 'typeof foo === (("strnig"))',
      errors: [{ messageId: 'invalidValue' }],
    },
    {
      code: '(typeof foo) === ("strnig")',
      errors: [{ messageId: 'invalidValue' }],
    },
  ],
});

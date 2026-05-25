import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-useless-concat', {
  valid: [
    // Non-`+` operators.
    'var a = 1 + 1;',
    'var a = 1 - 2;',
    "var a = 1 * '2';",
    "var a = 'a' * 'b';",
    "var a = 'a' - 'b';",

    // At least one side is not a string literal.
    'var a = foo + bar;',
    "var a = 'foo' + bar;",
    "var a = 1 + '1';",
    'var a = 1 + `1`;',
    'var a = `1` + 1;',
    "var a = 'a' + 1;",
    "var a = foo + 'a' + bar;",
    "var a = 'a' + 1 + 'b';",
    "var a = a + 'b' + c + 'd';",

    // Unary operators.
    'var a = (1 + +2) + `b`;',
    "var a = +'1' + 100;",
    "var a = -'a' + 'b';",

    // Multi-line concatenation is allowed.
    "var foo = 'foo' +\n 'bar';",
    "var a = ('a') +\n ('b');",
    "var a = 'a' +\n// comment\n'b';",

    // Compound assignment.
    "x += 'y';",

    // TaggedTemplateExpression.
    'var a = tag`a` + "b";',

    // `as` / `satisfies`.
    "var a = 'a' + (b as string);",
    "var a = ('a' as const) + 'b';",
  ],
  invalid: [
    // Basic two-literal.
    { code: "'a' + 'b'", errors: [{ messageId: 'unexpectedConcat' }] },
    { code: "`a` + 'b'", errors: [{ messageId: 'unexpectedConcat' }] },
    { code: '`a` + `b`', errors: [{ messageId: 'unexpectedConcat' }] },
    { code: "'' + ''", errors: [{ messageId: 'unexpectedConcat' }] },

    // Templates with substitutions.
    { code: "`a${x}` + 'b'", errors: [{ messageId: 'unexpectedConcat' }] },
    { code: '`${x}a` + `b${y}`', errors: [{ messageId: 'unexpectedConcat' }] },

    // Chains.
    { code: "foo + 'a' + 'b'", errors: [{ messageId: 'unexpectedConcat' }] },
    {
      code: "'a' + 'b' + 'c'",
      errors: [
        { messageId: 'unexpectedConcat' },
        { messageId: 'unexpectedConcat' },
      ],
    },
    {
      code: "'a' + 'b' + 'c' + 'd'",
      errors: [
        { messageId: 'unexpectedConcat' },
        { messageId: 'unexpectedConcat' },
        { messageId: 'unexpectedConcat' },
      ],
    },
    {
      code: "'a' + 'b' + foo",
      errors: [{ messageId: 'unexpectedConcat' }],
    },

    // Parentheses.
    {
      code: "(foo + 'a') + ('b' + 'c')",
      errors: [
        { messageId: 'unexpectedConcat' },
        { messageId: 'unexpectedConcat' },
      ],
    },
    {
      code: "'a' + ('b' + 'c')",
      errors: [
        { messageId: 'unexpectedConcat' },
        { messageId: 'unexpectedConcat' },
      ],
    },
    {
      code: "('a' + 'b') + ('c' + 'd')",
      errors: [
        { messageId: 'unexpectedConcat' },
        { messageId: 'unexpectedConcat' },
        { messageId: 'unexpectedConcat' },
      ],
    },
    {
      code: "'a' + ('b' + 'c') + 'd'",
      errors: [
        { messageId: 'unexpectedConcat' },
        { messageId: 'unexpectedConcat' },
        { messageId: 'unexpectedConcat' },
      ],
    },
    {
      code: "'a' + ('b' + ('c' + 'd'))",
      errors: [
        { messageId: 'unexpectedConcat' },
        { messageId: 'unexpectedConcat' },
        { messageId: 'unexpectedConcat' },
      ],
    },

    // Multi-line mixed.
    {
      code: "'a' +\n'b' + 'c'",
      errors: [{ messageId: 'unexpectedConcat' }],
    },
    {
      code: "'a' + 'b' +\n'c'",
      errors: [{ messageId: 'unexpectedConcat' }],
    },

    // Template chain.
    { code: 'foo + `a` + `b`', errors: [{ messageId: 'unexpectedConcat' }] },

    // Nested inside other expressions.
    { code: "foo('a' + 'b')", errors: [{ messageId: 'unexpectedConcat' }] },
    {
      code: "var x = {a: 'a' + 'b'};",
      errors: [{ messageId: 'unexpectedConcat' }],
    },
    {
      code: "var x = ['a' + 'b'];",
      errors: [{ messageId: 'unexpectedConcat' }],
    },
    {
      code: "`${'a' + 'b'}`",
      errors: [{ messageId: 'unexpectedConcat' }],
    },
    {
      code: "x += 'a' + 'b'",
      errors: [{ messageId: 'unexpectedConcat' }],
    },
  ],
});

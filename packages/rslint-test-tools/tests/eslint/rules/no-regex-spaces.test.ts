import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-regex-spaces', {
  valid: [
    // Baseline: no consecutive spaces
    'var foo = /foo/;',
    "var foo = RegExp('foo')",
    'var foo = / /;',
    "var foo = RegExp(' ')",
    'var foo = / a b c d /;',

    // Single space followed by explicit quantifier
    'var foo = /bar {3}baz/g;',
    "var foo = RegExp('bar {3}baz', 'g')",
    "var foo = new RegExp('bar {3}baz')",
    'var foo = /  +/;',
    'var foo = /  ?/;',
    'var foo = /  */;',
    'var foo = /  {2}/;',

    // Tabs / non-space whitespace don't count
    'var foo = /bar\t\t\tbaz/;',
    "var foo = RegExp('bar\t\t\tbaz');",
    "var foo = new RegExp('bar\t\t\tbaz');",

    // RegExp shadowed in the enclosing scope
    "var RegExp = function() {}; var foo = new RegExp('bar   baz');",
    "var RegExp = function() {}; var foo = RegExp('bar   baz');",

    // No consecutive spaces in the source
    String.raw`var foo = /bar \ baz/;`,
    String.raw`var foo = /bar\ \ baz/;`,
    String.raw`var foo = /bar \u0020 baz/;`,
    String.raw`var foo = /bar\u0020\u0020baz/;`,
    String.raw`var foo = new RegExp('bar \ baz')`,
    String.raw`var foo = new RegExp('bar\ \ baz')`,
    String.raw`var foo = new RegExp('bar \\ baz')`,
    String.raw`var foo = new RegExp('bar \u0020 baz')`,
    String.raw`var foo = new RegExp('bar\u0020\u0020baz')`,
    String.raw`var foo = new RegExp('bar \\u0020 baz')`,

    // Spaces inside character classes
    'var foo = /[  ]/;',
    'var foo = /[   ]/;',
    'var foo = / [  ] /;',
    'var foo = / [  ] [  ] /;',
    "var foo = new RegExp('[  ]');",
    "var foo = new RegExp('[   ]');",
    "var foo = new RegExp(' [  ] ');",
    "var foo = RegExp(' [  ] [  ] ');",
    String.raw`var foo = new RegExp(' \[   ');`,
    String.raw`var foo = new RegExp(' \[   \] ');`,

    // ES2024 v flag
    'var foo = /  {2}/v;',
    String.raw`var foo = /[\q{    }]/v;`,

    // Invalid regex — skipped to match ESLint's parsePattern try/catch
    "var foo = new RegExp('[  ');",
    "var foo = new RegExp('{  ', 'u');",
    "var foo = new RegExp('{  ', 'v');",

    // Flags cannot be determined
    "new RegExp('  ', flags)",
    "new RegExp('[[abc]  ]', flags + 'v')",
    String.raw`new RegExp('[[abc]\q{  }]', flags + 'v')`,
  ],
  invalid: [
    {
      code: 'var foo = /bar  baz/;',
      errors: [{ messageId: 'multipleSpaces' }],
    },
    {
      code: 'var foo = /bar    baz/;',
      errors: [{ messageId: 'multipleSpaces' }],
    },
    {
      code: 'var foo = / a b  c d /;',
      errors: [{ messageId: 'multipleSpaces' }],
    },

    {
      code: "var foo = RegExp(' a b c d  ');",
      errors: [{ messageId: 'multipleSpaces' }],
    },
    {
      code: "var foo = RegExp('bar    baz');",
      errors: [{ messageId: 'multipleSpaces' }],
    },
    {
      code: "var foo = new RegExp('bar    baz');",
      errors: [{ messageId: 'multipleSpaces' }],
    },

    // RegExp not shadowed where it's called
    {
      code: "{ let RegExp = function() {}; } var foo = RegExp('bar    baz');",
      errors: [{ messageId: 'multipleSpaces' }],
    },

    // Space runs followed by a quantifier
    {
      code: 'var foo = /bar   {3}baz/;',
      errors: [{ messageId: 'multipleSpaces' }],
    },
    {
      code: 'var foo = /bar    ?baz/;',
      errors: [{ messageId: 'multipleSpaces' }],
    },
    {
      code: "var foo = new RegExp('bar   *baz')",
      errors: [{ messageId: 'multipleSpaces' }],
    },
    {
      code: "var foo = RegExp('bar   +baz')",
      errors: [{ messageId: 'multipleSpaces' }],
    },
    {
      code: "var foo = new RegExp('bar    ');",
      errors: [{ messageId: 'multipleSpaces' }],
    },

    // Escaped backslash + spaces in regex literal
    {
      code: String.raw`var foo = /bar\  baz/;`,
      errors: [{ messageId: 'multipleSpaces' }],
    },

    // Spaces outside character classes
    { code: 'var foo = /[   ]  /;', errors: [{ messageId: 'multipleSpaces' }] },
    {
      code: 'var foo = /  [   ] /;',
      errors: [{ messageId: 'multipleSpaces' }],
    },
    {
      code: "var foo = new RegExp('[   ]  ');",
      errors: [{ messageId: 'multipleSpaces' }],
    },
    {
      code: "var foo = RegExp('  [ ]');",
      errors: [{ messageId: 'multipleSpaces' }],
    },

    // Escaped brackets don't open a class
    {
      code: String.raw`var foo = /\[  /;`,
      errors: [{ messageId: 'multipleSpaces' }],
    },
    {
      code: String.raw`var foo = /\[  \]/;`,
      errors: [{ messageId: 'multipleSpaces' }],
    },

    // Non-capturing groups / assertions
    { code: 'var foo = /(?:  )/;', errors: [{ messageId: 'multipleSpaces' }] },
    {
      code: "var foo = RegExp('^foo(?=   )');",
      errors: [{ messageId: 'multipleSpaces' }],
    },

    // Escape of the space character
    {
      code: String.raw`var foo = /\  /`,
      errors: [{ messageId: 'multipleSpaces' }],
    },
    {
      code: String.raw`var foo = / \  /`,
      errors: [{ messageId: 'multipleSpaces' }],
    },

    // Report only the first occurrence of consecutive spaces
    {
      code: 'var foo = /  foo   /;',
      errors: [{ messageId: 'multipleSpaces' }],
    },

    // Strings containing escape sequences — report but no fix
    {
      code: String.raw`var foo = new RegExp('\d  ')`,
      errors: [{ messageId: 'multipleSpaces' }],
    },
    {
      code: String.raw`var foo = RegExp('\u0041   ')`,
      errors: [{ messageId: 'multipleSpaces' }],
    },
    {
      code: String.raw`var foo = new RegExp('\\[  \\]');`,
      errors: [{ messageId: 'multipleSpaces' }],
    },

    // ES2024 v-flag: nested character classes
    {
      code: 'var foo = /[[    ]    ]    /v;',
      errors: [{ messageId: 'multipleSpaces' }],
    },
    {
      code: "var foo = new RegExp('[[    ]    ]    ', 'v');",
      errors: [{ messageId: 'multipleSpaces' }],
    },
  ],
});

// Upstream: ESLint v10.14.0 tests/lib/rules/no-invalid-regexp.js (all cases).
// Full ranges verified with ESLint v10.14.0; no fixes or suggestions.
import { RuleTester } from '@rslint/rule-tester';

new RuleTester().run('no-invalid-regexp', {
  valid: [
    "RegExp('')",
    'RegExp()',
    "RegExp('.', 'g')",
    "new RegExp('.')",
    'new RegExp',
    "new RegExp('.', 'im')",
    "global.RegExp('\\\\')",
    "function foo(RegExp) { RegExp('['); }",
    "function foo(RegExp) { new RegExp('['); }",
    "let RegExp; RegExp('.', 'z');",
    {
      code: "RegExp('[', 'z');",
      languageOptions: {
        globals: {
          RegExp: 'off',
        },
      },
    },
    "new RegExp('.', y)",
    "new RegExp('.', 'y')",
    "new RegExp('.', 'u')",
    "new RegExp('.', 'yu')",
    "new RegExp('/', 'yu')",
    "new RegExp('\\/', 'yu')",
    "new RegExp('\\\\u{65}', 'u')",
    "new RegExp('\\\\u{65}*', 'u')",
    "new RegExp('[\\\\u{0}-\\\\u{1F}]', 'u')",
    "new RegExp('.', 's')",
    "new RegExp('(?<=a)b')",
    "new RegExp('(?<!a)b')",
    "new RegExp('(?<a>b)\\k<a>')",
    "new RegExp('(?<a>b)\\k<a>', 'u')",
    "new RegExp('\\\\p{Letter}', 'u')",
    "RegExp('{', flags)",
    "new RegExp('{', flags)",
    "RegExp('\\\\u{0}*', flags)",
    "new RegExp('\\\\u{0}*', flags)",
    {
      code: "RegExp('{', flags)",
      options: [
        {
          allowConstructorFlags: ['u'],
        },
      ],
    },
    {
      code: "RegExp('\\\\u{0}*', flags)",
      options: [
        {
          allowConstructorFlags: ['a'],
        },
      ],
    },
    "new RegExp(pattern, 'g')",
    "new RegExp('.' + '', 'g')",
    "new RegExp(pattern, '')",
    'new RegExp(pattern)',
    "new RegExp('(?<\\\\ud835\\\\udc9c>.)', 'g')",
    "new RegExp('(?<\\\\u{1d49c}>.)', 'g')",
    "new RegExp('(?<𝒜>.)', 'g');",
    "new RegExp('\\\\p{Script=Nandinagari}', 'u');",
    "new RegExp('a+(?<Z>z)?', 'd')",
    "new RegExp('\\\\p{Script=Cpmn}', 'u')",
    "new RegExp('\\\\p{Script=Cypro_Minoan}', 'u')",
    "new RegExp('\\\\p{Script=Old_Uyghur}', 'u')",
    "new RegExp('\\\\p{Script=Ougr}', 'u')",
    "new RegExp('\\\\p{Script=Tangsa}', 'u')",
    "new RegExp('\\\\p{Script=Tnsa}', 'u')",
    "new RegExp('\\\\p{Script=Toto}', 'u')",
    "new RegExp('\\\\p{Script=Vith}', 'u')",
    "new RegExp('\\\\p{Script=Vithkuqi}', 'u')",
    "new RegExp('[A--B]', 'v')",
    "new RegExp('[A&&B]', 'v')",
    "new RegExp('[A--[0-9]]', 'v')",
    "new RegExp('[\\\\p{Basic_Emoji}--\\\\q{a|bc|def}]', 'v')",
    "new RegExp('[A--B]', flags)",
    "new RegExp('[[]\\\\u{0}*', flags)",
    "new RegExp('((?<k>a)|(?<k>b))')",
    "new RegExp('(?ims:foo)')",
    "new RegExp('(?ims-:foo)')",
    "new RegExp('(?-ims:foo)')",
    "new RegExp('(?s-i:foo)')",
    {
      code: "new RegExp('.', 'g')",
      options: [
        {
          allowConstructorFlags: [],
        },
      ],
    },
    {
      code: "new RegExp('.', 'g')",
      options: [
        {
          allowConstructorFlags: ['a'],
        },
      ],
    },
    {
      code: "new RegExp('.', 'a')",
      options: [
        {
          allowConstructorFlags: ['a'],
        },
      ],
    },
    {
      code: "new RegExp('.', 'ag')",
      options: [
        {
          allowConstructorFlags: ['a'],
        },
      ],
    },
    {
      code: "new RegExp('.', 'ga')",
      options: [
        {
          allowConstructorFlags: ['a'],
        },
      ],
    },
    {
      code: "new RegExp(pattern, 'ga')",
      options: [
        {
          allowConstructorFlags: ['a'],
        },
      ],
    },
    {
      code: "new RegExp('.' + '', 'ga')",
      options: [
        {
          allowConstructorFlags: ['a'],
        },
      ],
    },
    {
      code: "new RegExp('.', 'a')",
      options: [
        {
          allowConstructorFlags: ['a', 'z'],
        },
      ],
    },
    {
      code: "new RegExp('.', 'z')",
      options: [
        {
          allowConstructorFlags: ['a', 'z'],
        },
      ],
    },
    {
      code: "new RegExp('.', 'az')",
      options: [
        {
          allowConstructorFlags: ['a', 'z'],
        },
      ],
    },
    {
      code: "new RegExp('.', 'za')",
      options: [
        {
          allowConstructorFlags: ['a', 'z'],
        },
      ],
    },
    {
      code: "new RegExp('.', 'agz')",
      options: [
        {
          allowConstructorFlags: ['a', 'z'],
        },
      ],
    },
  ],
  invalid: [
    {
      code: "RegExp('[');",
      errors: [
        {
          messageId: 'regexMessage',
          message:
            'Invalid regular expression: /[/: Unterminated character class.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    {
      code: "RegExp('.', 'z');",
      errors: [
        {
          messageId: 'regexMessage',
          message: "Invalid flags supplied to RegExp constructor 'z'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: "RegExp('.', 'a');",
      options: [{}],
      errors: [
        {
          messageId: 'regexMessage',
          message: "Invalid flags supplied to RegExp constructor 'a'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: "new RegExp('.', 'a');",
      options: [
        {
          allowConstructorFlags: [],
        },
      ],
      errors: [
        {
          messageId: 'regexMessage',
          message: "Invalid flags supplied to RegExp constructor 'a'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: "new RegExp('.', 'z');",
      options: [
        {
          allowConstructorFlags: ['a'],
        },
      ],
      errors: [
        {
          messageId: 'regexMessage',
          message: "Invalid flags supplied to RegExp constructor 'z'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: "RegExp('.', 'a');",
      options: [
        {
          allowConstructorFlags: ['A'],
        },
      ],
      errors: [
        {
          messageId: 'regexMessage',
          message: "Invalid flags supplied to RegExp constructor 'a'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: "RegExp('.', 'A');",
      options: [
        {
          allowConstructorFlags: ['a'],
        },
      ],
      errors: [
        {
          messageId: 'regexMessage',
          message: "Invalid flags supplied to RegExp constructor 'A'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: "new RegExp('.', 'az');",
      options: [
        {
          allowConstructorFlags: ['z'],
        },
      ],
      errors: [
        {
          messageId: 'regexMessage',
          message: "Invalid flags supplied to RegExp constructor 'a'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: "new RegExp('.', 'aa');",
      options: [
        {
          allowConstructorFlags: ['a'],
        },
      ],
      errors: [
        {
          messageId: 'regexMessage',
          message: "Duplicate flags ('a') supplied to RegExp constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: "new RegExp('.', 'aA');",
      options: [
        {
          allowConstructorFlags: ['a'],
        },
      ],
      errors: [
        {
          messageId: 'regexMessage',
          message: "Invalid flags supplied to RegExp constructor 'A'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: "new RegExp('.', 'aaz');",
      options: [
        {
          allowConstructorFlags: ['a', 'z'],
        },
      ],
      errors: [
        {
          messageId: 'regexMessage',
          message: "Duplicate flags ('a') supplied to RegExp constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: "new RegExp('.', 'azz');",
      options: [
        {
          allowConstructorFlags: ['a', 'z'],
        },
      ],
      errors: [
        {
          messageId: 'regexMessage',
          message: "Duplicate flags ('z') supplied to RegExp constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: "new RegExp('.', 'aga');",
      options: [
        {
          allowConstructorFlags: ['a'],
        },
      ],
      errors: [
        {
          messageId: 'regexMessage',
          message: "Duplicate flags ('a') supplied to RegExp constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: "new RegExp('.', 'uu');",
      options: [
        {
          allowConstructorFlags: ['u'],
        },
      ],
      errors: [
        {
          messageId: 'regexMessage',
          message: "Duplicate flags ('u') supplied to RegExp constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: "new RegExp('.', 'ouo');",
      options: [
        {
          allowConstructorFlags: ['u'],
        },
      ],
      errors: [
        {
          messageId: 'regexMessage',
          message: "Invalid flags supplied to RegExp constructor 'oo'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: "new RegExp(')');",
      errors: [
        {
          messageId: 'regexMessage',
          message: "Invalid regular expression: /)/: Unmatched ')'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: "new RegExp('\\\\a', 'u');",
      errors: [
        {
          messageId: 'regexMessage',
          message: 'Invalid regular expression: /\\a/u: Invalid escape.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: "new RegExp('\\\\a', 'u');",
      options: [
        {
          allowConstructorFlags: ['u'],
        },
      ],
      errors: [
        {
          messageId: 'regexMessage',
          message: 'Invalid regular expression: /\\a/u: Invalid escape.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: "RegExp('\\\\u{0}*');",
      errors: [
        {
          messageId: 'regexMessage',
          message: 'Invalid regular expression: /\\u{0}*/: Nothing to repeat.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: "new RegExp('\\\\u{0}*');",
      errors: [
        {
          messageId: 'regexMessage',
          message: 'Invalid regular expression: /\\u{0}*/: Nothing to repeat.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: "new RegExp('\\\\u{0}*', '');",
      errors: [
        {
          messageId: 'regexMessage',
          message: 'Invalid regular expression: /\\u{0}*/: Nothing to repeat.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: "new RegExp('\\\\u{0}*', 'a');",
      options: [
        {
          allowConstructorFlags: ['a'],
        },
      ],
      errors: [
        {
          messageId: 'regexMessage',
          message: 'Invalid regular expression: /\\u{0}*/: Nothing to repeat.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: "RegExp('\\\\u{0}*');",
      options: [
        {
          allowConstructorFlags: ['a'],
        },
      ],
      errors: [
        {
          messageId: 'regexMessage',
          message: 'Invalid regular expression: /\\u{0}*/: Nothing to repeat.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: "new RegExp('\\\\');",
      errors: [
        {
          messageId: 'regexMessage',
          message: 'Invalid regular expression: /\\/: \\ at end of pattern.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: "RegExp(')' + '', 'a');",
      errors: [
        {
          messageId: 'regexMessage',
          message: "Invalid flags supplied to RegExp constructor 'a'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: "new RegExp('.' + '', 'az');",
      options: [
        {
          allowConstructorFlags: ['z'],
        },
      ],
      errors: [
        {
          messageId: 'regexMessage',
          message: "Invalid flags supplied to RegExp constructor 'a'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: "new RegExp(pattern, 'az');",
      options: [
        {
          allowConstructorFlags: ['a'],
        },
      ],
      errors: [
        {
          messageId: 'regexMessage',
          message: "Invalid flags supplied to RegExp constructor 'z'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: "new RegExp('[[]', 'v');",
      errors: [
        {
          messageId: 'regexMessage',
          message:
            'Invalid regular expression: /[[]/v: Unterminated character class.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: "new RegExp('.', 'uv');",
      errors: [
        {
          messageId: 'regexMessage',
          message: "Regex 'u' and 'v' flags cannot be used together.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: "new RegExp(pattern, 'uv');",
      errors: [
        {
          messageId: 'regexMessage',
          message: "Regex 'u' and 'v' flags cannot be used together.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: "new RegExp('[A--B]' /* valid only with `v` flag */, 'u')",
      errors: [
        {
          messageId: 'regexMessage',
          message:
            'Invalid regular expression: /[A--B]/u: Range out of order in character class.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 57,
        },
      ],
    },
    {
      code: "new RegExp('[[]\\\\u{0}*' /* valid only with `u` flag */, 'v')",
      errors: [
        {
          messageId: 'regexMessage',
          message:
            'Invalid regular expression: /[[]\\u{0}*/v: Unterminated character class.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 61,
        },
      ],
    },
    {
      code: "new RegExp('(?<k>a)(?<k>b)')",
      errors: [
        {
          messageId: 'regexMessage',
          message:
            'Invalid regular expression: /(?<k>a)(?<k>b)/: Duplicate capture group name.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: "new RegExp('(?ii:foo)')",
      errors: [
        {
          messageId: 'regexMessage',
          message:
            "Invalid regular expression: /(?ii:foo)/: Duplicated flag 'i'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: "new RegExp('(?-ii:foo)')",
      errors: [
        {
          messageId: 'regexMessage',
          message:
            "Invalid regular expression: /(?-ii:foo)/: Duplicated flag 'i'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: "new RegExp('(?i-i:foo)')",
      errors: [
        {
          messageId: 'regexMessage',
          message:
            "Invalid regular expression: /(?i-i:foo)/: Duplicated flag 'i'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: "new RegExp('(?-:foo)')",
      errors: [
        {
          messageId: 'regexMessage',
          message:
            'Invalid regular expression: /(?-:foo)/: Invalid empty flags.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: "new RegExp('(?g:foo)')",
      errors: [
        {
          messageId: 'regexMessage',
          message: 'Invalid regular expression: /(?g:foo)/: Invalid group.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: "new RegExp('(?-u:foo)')",
      errors: [
        {
          messageId: 'regexMessage',
          message: 'Invalid regular expression: /(?-u:foo)/: Invalid group.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
  ],
});

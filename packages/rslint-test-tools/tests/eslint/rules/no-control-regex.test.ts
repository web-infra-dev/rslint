import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-control-regex', {
  valid: [
    // Baseline: no control chars
    'var regex = /x1f/',
    String.raw`var regex = /\\x1f/`,
    "var regex = new RegExp('x1f')",
    "var regex = RegExp('x1f')",
    "new RegExp('[')",
    "RegExp('[')",
    String.raw`/\u{20}/u`,
    String.raw`/\u{1F}/`,
    String.raw`/\u{1F}/g`,
    String.raw`new RegExp("\\u{20}", "u")`,
    String.raw`new RegExp("\\u{1F}")`,
    String.raw`new RegExp("\\u{1F}", "g")`,
    String.raw`new RegExp("\\u{1F}", flags)`,
    String.raw`new RegExp("[\\q{\\u{20}}]", "v")`,
    String.raw`/[\u{20}--B]/v`,
    // Symbolic escapes — allowed
    String.raw`/\t/`,
    String.raw`/\n/`,
    String.raw`/\r/`,
    String.raw`/\v/`,
    String.raw`/\f/`,
    String.raw`/\0/`,
    String.raw`/\b/`,
    String.raw`/\cI/`,
    // Non-RegExp callees — should not match
    "foo.RegExp('\\x1f')",
    "window.RegExp('\\x1f')",
    "this.RegExp('\\x1f')",
    "regexp('\\x1f')",
    "new (function foo(){})('\\x1f')",
    // Non-string first argument
    'RegExp(pattern)',
    'RegExp(/x20/)',
    "RegExp('a' + 'b')",
    'RegExp(123)',
    'RegExp(null)',
    // No-args
    'new RegExp',
    'RegExp()',
    'new RegExp()',
    // Spread first argument — not a StringLiteral, skip
    'new RegExp(...args)',
    // Surrogate pair (non-control)
    String.raw`/\uD83D\uDC7F/`,
    String.raw`new RegExp("\\uD83D\\uDC7F")`,
    // Legacy octal in regex literal
    String.raw`/\01/`,
    String.raw`/\012/`,
    // \p{...} unicode property
    String.raw`/\p{Letter}/u`,
    String.raw`new RegExp("\\p{Letter}", "u")`,
  ],
  invalid: [
    // Regex literals: \xHH
    {
      code: String.raw`var regex = /\x1f/`,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: String.raw`var regex = /\x00/`,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: String.raw`var regex = /\\\x1f\\x1e/`,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: String.raw`var regex = /\\\x1fFOO\\x00/`,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: String.raw`var regex = /FOO\\\x1fFOO\\x1f/`,
      errors: [{ messageId: 'unexpected' }],
    },
    // Regex literal: \uHHHH
    {
      code: String.raw`/\u000C/`,
      errors: [{ messageId: 'unexpected' }],
    },
    // Regex literal: \u{H} under u/v flag
    {
      code: String.raw`/\u{C}/u`,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: String.raw`/\u{1F}/u`,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: String.raw`/\u{1F}/gui`,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: String.raw`/\u{1111}*\x1F/u`,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: String.raw`/[\u{1F}--B]/v`,
      errors: [{ messageId: 'unexpected' }],
    },
    // Regex literal: named capture + control
    {
      code: 'var regex = /(?<a>\\x1f)/',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: String.raw`var regex = /(?<\u{1d49c}>.)\x1f/`,
      errors: [{ messageId: 'unexpected' }],
    },
    // Constructor: raw / escaped strings
    {
      code: "var regex = new RegExp('\\x1f\\x1e')",
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "var regex = new RegExp('\\x1fFOO\\x00')",
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "var regex = new RegExp('FOO\\x1fFOO\\x1f')",
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "var regex = RegExp('\\x1f')",
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: String.raw`new RegExp("\\u001F", flags)`,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: String.raw`new RegExp("\\u{1111}*\\x1F", "u")`,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: String.raw`new RegExp("\\u{1F}", "u")`,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: String.raw`new RegExp("\\u{1F}", "gui")`,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: String.raw`new RegExp("[\\q{\\u{1F}}]", "v")`,
      errors: [{ messageId: 'unexpected' }],
    },
    // Parenthesized callee
    {
      code: "(RegExp)('\\x1f')",
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "((RegExp))('\\x1f')",
      errors: [{ messageId: 'unexpected' }],
    },
    // Nesting contexts
    {
      code: String.raw`foo(/\x1f/)`,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: String.raw`[/\x1f/]`,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: String.raw`({ re: /\x1f/ })`,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'const fn = () => /\\x1f/;',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "RegExp(RegExp('\\x1f'))",
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'class C { m() { /\\x1f/; } }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'function f(x = /\\x1f/) {}',
      errors: [{ messageId: 'unexpected' }],
    },
    // Character class range of controls — one diagnostic
    {
      code: String.raw`/[\x00-\x1f]/`,
      errors: [{ messageId: 'unexpected' }],
    },
    // Multiple statements — only the bad one reports
    {
      code: String.raw`/\x11/; RegExp("foo", "uv");`,
      errors: [{ messageId: 'unexpected' }],
    },
  ],
});

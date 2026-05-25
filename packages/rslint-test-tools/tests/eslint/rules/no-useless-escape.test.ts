import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-useless-escape', {
  valid: [
    // ---- Regex outside character classes ----
    'var foo = /\\./;',
    'var foo = /\\//g;',
    'var foo = /""/;',
    "var foo = /''/;",
    'var foo = /\\D/;',
    'var foo = /\\W/;',
    'var foo = /\\w/;',
    'var foo = /\\\\/g;',
    'var foo = /\\w\\$\\*\\./;',
    'var foo = /\\^\\+\\./;',

    // ---- String literals: valid escape sequences ----
    'var foo = "\\x12";',
    'var foo = "\\u00a9";',
    'var foo = "\\"";',
    'var foo = "foo \\\\ bar";',
    'var foo = "\\t";',
    'var foo = "foo \\b bar";',
    "var foo = '\\n';",
    "var foo = 'foo \\r bar';",
    "var foo = '\\v';",
    "var foo = '\\f';",

    // ---- Template literals: valid escape sequences ----
    'var foo = `\\x12`;',
    'var foo = `\\u00a9`;',
    'var foo = `xs\\u2111`;',
    'var foo = `foo \\\\ bar`;',
    'var foo = `\\t`;',
    'var foo = `\\n`;',
    'var foo = `\\r`;',
    'var foo = `\\v`;',
    'var foo = `\\f`;',

    // ---- Quote escape inside template ----
    'var foo = `\\``;',
    'var foo = `\\`${foo}\\``;',
    // `\$` followed by `{` and `\{` preceded by `$` are necessary.
    'var foo = `\\${{${foo}`;',
    'var foo = `$\\{{${foo}`;',
    // Tagged template — escapes are exposed via the `raw` array.
    'var foo = String.raw`\\.`;',
    'var foo = myFunc`\\.`;',

    // ---- Regex character classes ----
    'var foo = /[\\d]/;',
    'var foo = /[a\\-b]/;',
    'var foo = /foo\\?/;',
    'var foo = /example\\.com/;',
    'var foo = /foo\\|bar/;',
    'var foo = /[\\^bar]/;',
    'var foo = /\\(bar\\)/;',
    'var foo = /[[\\]]/;',
    'var foo = /[\\]\\]]/;',
    'var foo = /\\[abc]/;',

    // ---- Special regex escapes ----
    'var foo = /\\0/;',
    'var foo = /\\1/;',
    'var foo = /(a)\\1/;',
    'var foo = /(a)\\12/;',
    'var foo = /[\\0]/;',

    // ---- Carets ----
    '/[^^]/;',
    '/[^^]/u;',

    // ---- v-flag character class escapes ----
    '/[\\q{abc}]/v;',
    '/[\\(]/v;',
    '/[\\-]/v;',
    '/[\\&&]/v;',
    '/[\\&&&\\&]/v;',
    '/[\\^]/v;',

    // ---- allowRegexCharacters option ----
    {
      code: 'var foo = /\\#/;',
      options: { allowRegexCharacters: ['#'] },
    },
    {
      code: 'var foo = /[ab\\-]/;',
      options: { allowRegexCharacters: ['-'] },
    },
    {
      code: 'var foo = /\\-/;',
      options: { allowRegexCharacters: ['-'] },
    },
  ],
  invalid: [
    // ---- Regex outside character class ----
    {
      code: 'var foo = /\\#/;',
      errors: [
        {
          messageId: 'unnecessaryEscape',
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      code: 'var foo = /\\;/;',
      errors: [
        {
          messageId: 'unnecessaryEscape',
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    // ---- String literals: identity escapes that aren't valid ----
    {
      code: 'var foo = "\\\'";',
      errors: [{ messageId: 'unnecessaryEscape', line: 1, column: 12 }],
    },
    {
      code: 'var foo = "\\#/";',
      errors: [{ messageId: 'unnecessaryEscape', line: 1, column: 12 }],
    },
    {
      code: 'var foo = "\\a";',
      errors: [{ messageId: 'unnecessaryEscape', line: 1, column: 12 }],
    },
    {
      code: 'var foo = "\\B";',
      errors: [{ messageId: 'unnecessaryEscape', line: 1, column: 12 }],
    },
    {
      code: 'var foo = "\\@";',
      errors: [{ messageId: 'unnecessaryEscape', line: 1, column: 12 }],
    },
    {
      code: 'var foo = "foo \\a bar";',
      errors: [{ messageId: 'unnecessaryEscape', line: 1, column: 16 }],
    },
    {
      code: "var foo = '\\\"';",
      errors: [{ messageId: 'unnecessaryEscape', line: 1, column: 12 }],
    },
    {
      code: "var foo = '\\#';",
      errors: [{ messageId: 'unnecessaryEscape', line: 1, column: 12 }],
    },
    {
      code: "var foo = '\\$';",
      errors: [{ messageId: 'unnecessaryEscape', line: 1, column: 12 }],
    },
    {
      code: "var foo = '\\p';",
      errors: [{ messageId: 'unnecessaryEscape', line: 1, column: 12 }],
    },
    {
      code: "var foo = '\\p\\a\\@';",
      errors: [
        { messageId: 'unnecessaryEscape', line: 1, column: 12 },
        { messageId: 'unnecessaryEscape', line: 1, column: 14 },
        { messageId: 'unnecessaryEscape', line: 1, column: 16 },
      ],
    },
    {
      code: "var foo = '\\`';",
      errors: [{ messageId: 'unnecessaryEscape', line: 1, column: 12 }],
    },

    // ---- Template literals ----
    {
      code: 'var foo = `\\"`;',
      errors: [{ messageId: 'unnecessaryEscape', line: 1, column: 12 }],
    },
    {
      code: "var foo = `\\'`;",
      errors: [{ messageId: 'unnecessaryEscape', line: 1, column: 12 }],
    },
    {
      code: 'var foo = `\\#`;',
      errors: [{ messageId: 'unnecessaryEscape', line: 1, column: 12 }],
    },
    {
      code: "var foo = '\\`foo\\`';",
      errors: [
        { messageId: 'unnecessaryEscape', line: 1, column: 12 },
        { messageId: 'unnecessaryEscape', line: 1, column: 17 },
      ],
    },
    {
      code: 'var foo = `\\#${foo}`;',
      errors: [{ messageId: 'unnecessaryEscape', line: 1, column: 12 }],
    },
    {
      code: 'var foo = `\\$\\{{${foo}`;',
      errors: [{ messageId: 'unnecessaryEscape', line: 1, column: 12 }],
    },
    {
      code: 'var foo = `\\$a${foo}`;',
      errors: [{ messageId: 'unnecessaryEscape', line: 1, column: 12 }],
    },
    {
      code: 'var foo = `a\\{{${foo}`;',
      errors: [{ messageId: 'unnecessaryEscape', line: 1, column: 13 }],
    },

    // ---- Regex character class ----
    {
      code: 'var foo = /[ab\\-]/;',
      errors: [{ messageId: 'unnecessaryEscape', line: 1, column: 15 }],
    },
    {
      code: 'var foo = /[\\-ab]/;',
      errors: [{ messageId: 'unnecessaryEscape', line: 1, column: 13 }],
    },
    {
      code: 'var foo = /[ab\\?]/;',
      errors: [{ messageId: 'unnecessaryEscape', line: 1, column: 15 }],
    },
    {
      code: 'var foo = /[ab\\.]/;',
      errors: [{ messageId: 'unnecessaryEscape', line: 1, column: 15 }],
    },
    {
      code: 'var foo = /\\-/;',
      errors: [{ messageId: 'unnecessaryEscape', line: 1, column: 12 }],
    },
    {
      code: 'var foo = /[\\-]/;',
      errors: [{ messageId: 'unnecessaryEscape', line: 1, column: 13 }],
    },
    {
      code: 'var foo = /[\\B]/;',
      errors: [{ messageId: 'unnecessaryEscape', line: 1, column: 13 }],
    },
    {
      code: 'var foo = /[a\\^]/;',
      errors: [{ messageId: 'unnecessaryEscape', line: 1, column: 14 }],
    },

    // ---- Caret in negated class ----
    {
      code: '/[^\\^]/;',
      errors: [{ messageId: 'unnecessaryEscape', line: 1, column: 4 }],
    },

    // ---- Directive prologue ----
    {
      code: '"use\\ strict";',
      errors: [{ messageId: 'unnecessaryEscape', line: 1, column: 5 }],
    },

    // ---- v-flag identity escapes ----
    {
      code: '/[\\$]/v;',
      errors: [{ messageId: 'unnecessaryEscape', line: 1, column: 3 }],
    },
    {
      code: '/[\\&\\&]/v;',
      errors: [{ messageId: 'unnecessaryEscape', line: 1, column: 3 }],
    },

    // ---- v-flag set-operation escapes (no escapeBackslash suggestion) ----
    {
      code: '/[\\p{ASCII}--\\.]/v;',
      errors: [{ messageId: 'unnecessaryEscape', line: 1, column: 14 }],
    },
    {
      code: '/[\\.--\\.--\\.]/v;',
      errors: [
        { messageId: 'unnecessaryEscape', line: 1, column: 3 },
        { messageId: 'unnecessaryEscape', line: 1, column: 7 },
        { messageId: 'unnecessaryEscape', line: 1, column: 11 },
      ],
    },
    {
      code: '/[[\\.&]--[\\.&]]/v;',
      errors: [
        { messageId: 'unnecessaryEscape', line: 1, column: 4 },
        { messageId: 'unnecessaryEscape', line: 1, column: 11 },
      ],
    },

    // ---- allowRegexCharacters: only allows specific chars ----
    {
      code: 'var foo = /\\#\\@/;',
      options: { allowRegexCharacters: ['#'] },
      errors: [{ messageId: 'unnecessaryEscape', line: 1, column: 14 }],
    },
    {
      code: 'var foo = /[a\\@b]/;',
      options: { allowRegexCharacters: ['#'] },
      errors: [{ messageId: 'unnecessaryEscape', line: 1, column: 14 }],
    },

    // ---- rslint-specific extensions (tsgo / Go-port edge cases) ----

    // ---- Multi-span template literals: each span reports independently ----
    {
      code: 'var foo = `\\#${a}\\@${b}\\!`;',
      errors: [
        { messageId: 'unnecessaryEscape', line: 1, column: 12 },
        { messageId: 'unnecessaryEscape', line: 1, column: 18 },
        { messageId: 'unnecessaryEscape', line: 1, column: 24 },
      ],
    },

    // ---- TS surface around string literals — non-null, parens, as, satisfies ----
    {
      code: 'var foo = ("\\#")!;',
      errors: [{ messageId: 'unnecessaryEscape', line: 1, column: 13 }],
    },
    {
      code: 'var foo = "\\@" as string;',
      errors: [{ messageId: 'unnecessaryEscape', line: 1, column: 12 }],
    },
    {
      code: 'var foo = "\\@" satisfies string;',
      errors: [{ messageId: 'unnecessaryEscape', line: 1, column: 12 }],
    },

    // ---- String literals in real TS contexts ----
    {
      code: 'import foo from "./a\\#";',
      errors: [{ messageId: 'unnecessaryEscape', line: 1, column: 21 }],
    },
    {
      code: 'enum E { A = "\\#" }',
      errors: [{ messageId: 'unnecessaryEscape', line: 1, column: 15 }],
    },
    {
      code: 'type T = "\\#";',
      errors: [{ messageId: 'unnecessaryEscape', line: 1, column: 11 }],
    },
    {
      code: 'const obj = { ["\\#"]: 1 };',
      errors: [{ messageId: 'unnecessaryEscape', line: 1, column: 17 }],
    },

    // ---- Function-body directive prologue ----
    {
      code: 'function f() { "use\\ strict"; return 1; }',
      errors: [{ messageId: 'unnecessaryEscape', line: 1, column: 20 }],
    },

    // ---- v-mode triple-nested classes ----
    // Outer class has `&&` set-op → escape directly under it has no escapeBackslash.
    {
      code: '/[\\.&&[a&&[b]]]/v;',
      errors: [{ messageId: 'unnecessaryEscape', line: 1, column: 3 }],
    },
    // Innermost class has no set-ops → escape gets escapeBackslash.
    {
      code: '/[a&&[b&&[\\.]]]/v;',
      errors: [{ messageId: 'unnecessaryEscape', line: 1, column: 11 }],
    },

    // ---- v-mode `\^` not at class start in negated class ----
    {
      code: '/[^a\\^]/v;',
      errors: [{ messageId: 'unnecessaryEscape', line: 1, column: 5 }],
    },

    // ---- Multi-byte UTF-8 escape ----
    {
      code: 'var foo = "\\é";',
      errors: [{ messageId: 'unnecessaryEscape', line: 1, column: 12 }],
    },

    // ---- Multiple escapes interleaved with valid escapes in regex ----
    {
      code: 'var foo = /\\d\\@\\w\\#\\s/;',
      errors: [
        { messageId: 'unnecessaryEscape', line: 1, column: 14 },
        { messageId: 'unnecessaryEscape', line: 1, column: 18 },
      ],
    },

    // ---- Surrogate-pair preceding the escape (UTF-16 column counting) ----
    // 👍 occupies cols 12-13; \# lands at col 14.
    {
      code: "var foo = '👍\\#';",
      errors: [{ messageId: 'unnecessaryEscape', line: 1, column: 14 }],
    },

    // ---- Directive-container precision: only function-body Blocks ----
    {
      code: 'if (true) { "ba\\z"; }',
      errors: [{ messageId: 'unnecessaryEscape', line: 1, column: 16 }],
    },
    {
      code: '{ "ba\\z"; }',
      errors: [{ messageId: 'unnecessaryEscape', line: 1, column: 6 }],
    },
    // Class static block IS a directive container.
    {
      code: 'class C { static { "use\\ strict"; } }',
      errors: [{ messageId: 'unnecessaryEscape', line: 1, column: 24 }],
    },

    // ---- v-mode \q{...} body: nested escapes are flagged ----
    {
      code: '/[\\q{a\\.b}]/v;',
      errors: [{ messageId: 'unnecessaryEscape', line: 1, column: 7 }],
    },
    {
      code: '/[\\q{\\@a|\\#b}]/v;',
      errors: [
        { messageId: 'unnecessaryEscape', line: 1, column: 6 },
        { messageId: 'unnecessaryEscape', line: 1, column: 10 },
      ],
    },
  ],
});

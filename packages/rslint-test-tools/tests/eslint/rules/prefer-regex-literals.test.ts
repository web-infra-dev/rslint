import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

// Migrated from the existing suite. Each active case pins the full diagnostic and edits.
new RuleTester(config).run('prefer-regex-literals', {
  valid: [
    {
      code: '/abc/',
    },
    {
      code: '/abc/g',
    },
    {
      code: 'new RegExp(pattern)',
    },
    {
      code: "new RegExp('\\\\p{Emoji_Presentation}\\\\P{Script_Extensions=Latin}' + '', `ug`)",
    },
    {
      code: "new RegExp('\\\\cA' + '')",
    },
    {
      code: "RegExp(pattern, 'g')",
    },
    {
      code: "new RegExp(f('a'))",
    },
    {
      code: "RegExp(prefix + 'a')",
    },
    {
      code: "new RegExp('a' + suffix)",
    },
    {
      code: 'RegExp(`a` + suffix);',
    },
    {
      code: 'new RegExp(String.raw`a` + suffix);',
    },
    {
      code: "RegExp('a', flags)",
    },
    {
      code: "const flags = 'gu';RegExp('a', flags)",
    },
    {
      code: "RegExp('a', 'g' + flags)",
    },
    {
      code: 'new RegExp(String.raw`a`, flags);',
    },
    {
      code: 'RegExp(`${prefix}abc`)',
    },
    {
      code: 'new RegExp(`a${b}c`);',
    },
    {
      code: "new RegExp(`a${''}c`);",
    },
    {
      code: 'new RegExp(String.raw`a${b}c`);',
    },
    {
      code: "new RegExp(String.raw`a${''}c`);",
    },
    {
      code: "new RegExp('a' + 'b')",
    },
    {
      code: 'RegExp(1)',
    },
    {
      code: "new RegExp('(\\\\p{Emoji_Presentation})\\\\1' + '', `ug`)",
    },
    {
      code: "RegExp(String.raw`\\78\\126` + '\\\\5934', '' + `g` + '')",
    },
    {
      code: "func(new RegExp(String.raw`a${''}c\\d`, 'u'),new RegExp(String.raw`a${''}c\\d`, 'u'))",
    },
    {
      code: 'new RegExp(\'\\\\[\' + "b\\\\]")',
    },
    {
      code: 'new RegExp(/a/, flags);',
      options: [
        {
          disallowRedundantWrapping: true,
        },
      ],
    },
    {
      code: 'new RegExp(/a/, `u${flags}`);',
      options: [
        {
          disallowRedundantWrapping: true,
        },
      ],
    },
    {
      code: 'new RegExp(/a/);',
      options: [{}],
    },
    {
      code: 'new RegExp(/a/);',
      options: [
        {
          disallowRedundantWrapping: false,
        },
      ],
    },
    {
      code: 'new RegExp;',
    },
    {
      code: 'new RegExp();',
    },
    {
      code: 'RegExp();',
    },
    {
      code: "new RegExp('a', 'g', 'b');",
    },
    {
      code: "RegExp('a', 'g', 'b');",
    },
    {
      code: 'new RegExp(`a`, `g`, `b`);',
    },
    {
      code: 'RegExp(`a`, `g`, `b`);',
    },
    {
      code: 'new RegExp(String.raw`a`, String.raw`g`, String.raw`b`);',
    },
    {
      code: 'RegExp(String.raw`a`, String.raw`g`, String.raw`b`);',
    },
    {
      code: "new RegExp(/a/, 'u', 'foo');",
      options: [
        {
          disallowRedundantWrapping: true,
        },
      ],
    },
    {
      code: 'new RegExp(String`a`);',
    },
    {
      code: 'RegExp(raw`a`);',
    },
    {
      code: 'new RegExp(f(String.raw)`a`);',
    },
    {
      code: 'RegExp(string.raw`a`);',
    },
    {
      code: 'new RegExp(String.Raw`a`);',
    },
    {
      code: 'new RegExp(String[raw]`a`);',
    },
    {
      code: 'RegExp(String.raw.foo`a`);',
    },
    {
      code: 'new RegExp(String.foo.raw`a`);',
    },
    {
      code: 'RegExp(foo.String.raw`a`);',
    },
    {
      code: 'new RegExp(String.raw);',
    },
    {
      code: 'let String; new RegExp(String.raw`a`);',
    },
    {
      code: 'function foo() { var String; new RegExp(String.raw`a`); }',
    },
    {
      code: 'function foo(String) { RegExp(String.raw`a`); }',
    },
    {
      code: 'if (foo) { const String = bar; RegExp(String.raw`a`); }',
    },
    {
      code: '/* globals String:off */ new RegExp(String.raw`a`);',
    },
    {
      code: "RegExp('a', String.raw`g`);",
      languageOptions: {
        globals: {
          String: 'off',
        },
      },
    },
    {
      code: "new Regexp('abc');",
    },
    {
      code: 'Regexp(`a`);',
    },
    {
      code: 'new Regexp(String.raw`a`);',
    },
    {
      code: "let RegExp; new RegExp('a');",
    },
    {
      code: "function foo() { var RegExp; RegExp('a', 'g'); }",
    },
    {
      code: 'function foo(RegExp) { new RegExp(String.raw`a`); }',
    },
    {
      code: "if (foo) { const RegExp = bar; RegExp('a'); }",
    },
    {
      code: "/* globals RegExp:off */ new RegExp('a');",
    },
    {
      code: "RegExp('a');",
      languageOptions: {
        globals: {
          RegExp: 'off',
        },
      },
    },
    {
      code: "new globalThis.RegExp('a');",
      languageOptions: {
        ecmaVersion: 5,
      },
    },
    {
      code: "new globalThis.RegExp('a');",
      languageOptions: {
        ecmaVersion: 2015,
      },
    },
    {
      code: "new globalThis.RegExp('a');",
      languageOptions: {
        ecmaVersion: 2017,
      },
    },
    {
      code: "class C { #RegExp; foo() { globalThis.#RegExp('a'); } }",
    },
    {
      code: "new RegExp('[[A--B]]' + a, 'v')",
    },
  ],
  invalid: [
    {
      code: "new RegExp('abc');",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 18,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/abc/;',
            },
          ],
        },
      ],
    },
    {
      code: "RegExp('abc');",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/abc/;',
            },
          ],
        },
      ],
    },
    {
      code: "new RegExp('abc', 'g');",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/abc/g;',
            },
          ],
        },
      ],
    },
    {
      code: "RegExp('abc', 'g');",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/abc/g;',
            },
          ],
        },
      ],
    },
    {
      code: 'new RegExp(`abc`);',
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 18,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/abc/;',
            },
          ],
        },
      ],
    },
    {
      code: 'RegExp(`abc`);',
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/abc/;',
            },
          ],
        },
      ],
    },
    {
      code: 'new RegExp(`abc`, `g`);',
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/abc/g;',
            },
          ],
        },
      ],
    },
    {
      code: 'RegExp(`abc`, `g`);',
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/abc/g;',
            },
          ],
        },
      ],
    },
    {
      code: 'new RegExp(String.raw`abc`);',
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 28,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/abc/;',
            },
          ],
        },
      ],
    },
    {
      code: `new RegExp(String.raw\`abc
abc\`);`,
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 2,
          endColumn: 6,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/abc\\nabc/;',
            },
          ],
        },
      ],
    },
    {
      code: `new RegExp(String.raw\`	abc
abc\`);`,
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 2,
          endColumn: 6,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/\\tabc\\nabc/;',
            },
          ],
        },
      ],
    },
    {
      code: 'RegExp(String.raw`abc`);',
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/abc/;',
            },
          ],
        },
      ],
    },
    {
      code: 'new RegExp(String.raw`abc`, String.raw`g`);',
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 43,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/abc/g;',
            },
          ],
        },
      ],
    },
    {
      code: 'RegExp(String.raw`abc`, String.raw`g`);',
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 39,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/abc/g;',
            },
          ],
        },
      ],
    },
    {
      code: "new RegExp(String['raw']`a`);",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/a/;',
            },
          ],
        },
      ],
    },
    {
      code: "new RegExp('');",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 15,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/(?:)/;',
            },
          ],
        },
      ],
    },
    {
      code: "RegExp('', '');",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 15,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/(?:)/;',
            },
          ],
        },
      ],
    },
    {
      code: 'new RegExp(String.raw``);',
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/(?:)/;',
            },
          ],
        },
      ],
    },
    {
      code: "new RegExp('a', `g`);",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/a/g;',
            },
          ],
        },
      ],
    },
    {
      code: "RegExp(`a`, 'g');",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 17,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/a/g;',
            },
          ],
        },
      ],
    },
    {
      code: "RegExp(String.raw`a`, 'g');",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 27,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/a/g;',
            },
          ],
        },
      ],
    },
    {
      code: 'new RegExp(String.raw`\\d`, `g`);',
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 32,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/\\d/g;',
            },
          ],
        },
      ],
    },
    {
      code: 'new RegExp(String.raw`\\\\d`, `g`);',
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 33,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/\\\\d/g;',
            },
          ],
        },
      ],
    },
    {
      code: "new RegExp(String['raw']`\\\\d`, `g`);",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 36,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/\\\\d/g;',
            },
          ],
        },
      ],
    },
    {
      code: 'new RegExp(String["raw"]`\\\\d`, `g`);',
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 36,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/\\\\d/g;',
            },
          ],
        },
      ],
    },
    {
      code: "RegExp('a', String.raw`g`);",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 27,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/a/g;',
            },
          ],
        },
      ],
    },
    {
      code: "new globalThis.RegExp('a');",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 27,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/a/;',
            },
          ],
        },
      ],
    },
    {
      code: "globalThis.RegExp('a');",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/a/;',
            },
          ],
        },
      ],
    },
    {
      code: 'new RegExp(/a/);',
      options: [
        {
          disallowRedundantWrapping: true,
        },
      ],
      errors: [
        {
          messageId: 'unexpectedRedundantRegExp',
          message:
            "Regular expression literal is unnecessarily wrapped within a 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 16,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/a/;',
            },
          ],
        },
      ],
    },
    {
      code: "new RegExp(/a/, 'u');",
      options: [
        {
          disallowRedundantWrapping: true,
        },
      ],
      errors: [
        {
          messageId: 'unexpectedRedundantRegExpWithFlags',
          message:
            "Use regular expression literal with flags instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'replaceWithLiteralAndFlags',
              desc: "Replace with an equivalent regular expression literal with flags 'u'.",
              output: '/a/u;',
            },
          ],
        },
      ],
    },
    {
      code: "new RegExp(/a/g, '');",
      options: [
        {
          disallowRedundantWrapping: true,
        },
      ],
      errors: [
        {
          messageId: 'unexpectedRedundantRegExpWithFlags',
          message:
            "Use regular expression literal with flags instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'replaceWithLiteralAndFlags',
              desc: "Replace with an equivalent regular expression literal with flags ''.",
              output: '/a/;',
            },
            {
              messageId: 'replaceWithIntendedLiteralAndFlags',
              desc: "Replace with a regular expression literal with flags 'g'.",
              output: '/a/g;',
            },
          ],
        },
      ],
    },
    {
      code: "new RegExp(/a/g, 'g');",
      options: [
        {
          disallowRedundantWrapping: true,
        },
      ],
      errors: [
        {
          messageId: 'unexpectedRedundantRegExpWithFlags',
          message:
            "Use regular expression literal with flags instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
          suggestions: [
            {
              messageId: 'replaceWithLiteralAndFlags',
              desc: "Replace with an equivalent regular expression literal with flags 'g'.",
              output: '/a/g;',
            },
          ],
        },
      ],
    },
    {
      code: "new RegExp(/a/ig, 'g');",
      options: [
        {
          disallowRedundantWrapping: true,
        },
      ],
      errors: [
        {
          messageId: 'unexpectedRedundantRegExpWithFlags',
          message:
            "Use regular expression literal with flags instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
          suggestions: [
            {
              messageId: 'replaceWithLiteralAndFlags',
              desc: "Replace with an equivalent regular expression literal with flags 'g'.",
              output: '/a/g;',
            },
            {
              messageId: 'replaceWithIntendedLiteralAndFlags',
              desc: "Replace with a regular expression literal with flags 'ig'.",
              output: '/a/ig;',
            },
          ],
        },
      ],
    },
    {
      code: "new RegExp(/a/g, 'ig');",
      options: [
        {
          disallowRedundantWrapping: true,
        },
      ],
      errors: [
        {
          messageId: 'unexpectedRedundantRegExpWithFlags',
          message:
            "Use regular expression literal with flags instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
          suggestions: [
            {
              messageId: 'replaceWithLiteralAndFlags',
              desc: "Replace with an equivalent regular expression literal with flags 'ig'.",
              output: '/a/ig;',
            },
          ],
        },
      ],
    },
    {
      code: "new RegExp(/a/i, 'g');",
      options: [
        {
          disallowRedundantWrapping: true,
        },
      ],
      errors: [
        {
          messageId: 'unexpectedRedundantRegExpWithFlags',
          message:
            "Use regular expression literal with flags instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
          suggestions: [
            {
              messageId: 'replaceWithLiteralAndFlags',
              desc: "Replace with an equivalent regular expression literal with flags 'g'.",
              output: '/a/g;',
            },
            {
              messageId: 'replaceWithIntendedLiteralAndFlags',
              desc: "Replace with a regular expression literal with flags 'ig'.",
              output: '/a/ig;',
            },
          ],
        },
      ],
    },
    {
      code: "new RegExp(/a/i, 'i');",
      options: [
        {
          disallowRedundantWrapping: true,
        },
      ],
      errors: [
        {
          messageId: 'unexpectedRedundantRegExpWithFlags',
          message:
            "Use regular expression literal with flags instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
          suggestions: [
            {
              messageId: 'replaceWithLiteralAndFlags',
              desc: "Replace with an equivalent regular expression literal with flags 'i'.",
              output: '/a/i;',
            },
          ],
        },
      ],
    },
    {
      code: 'new RegExp(/a/, `u`);',
      options: [
        {
          disallowRedundantWrapping: true,
        },
      ],
      errors: [
        {
          messageId: 'unexpectedRedundantRegExpWithFlags',
          message:
            "Use regular expression literal with flags instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'replaceWithLiteralAndFlags',
              desc: "Replace with an equivalent regular expression literal with flags 'u'.",
              output: '/a/u;',
            },
          ],
        },
      ],
    },
    {
      code: 'new RegExp(/a/, `gi`);',
      options: [
        {
          disallowRedundantWrapping: true,
        },
      ],
      errors: [
        {
          messageId: 'unexpectedRedundantRegExpWithFlags',
          message:
            "Use regular expression literal with flags instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
          suggestions: [
            {
              messageId: 'replaceWithLiteralAndFlags',
              desc: "Replace with an equivalent regular expression literal with flags 'gi'.",
              output: '/a/gi;',
            },
          ],
        },
      ],
    },
    {
      code: "new RegExp('a');",
      options: [
        {
          disallowRedundantWrapping: true,
        },
      ],
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 16,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/a/;',
            },
          ],
        },
      ],
    },
    {
      code: 'new RegExp(/a/, String.raw`u`);',
      options: [
        {
          disallowRedundantWrapping: true,
        },
      ],
      errors: [
        {
          messageId: 'unexpectedRedundantRegExpWithFlags',
          message:
            "Use regular expression literal with flags instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 31,
          suggestions: [
            {
              messageId: 'replaceWithLiteralAndFlags',
              desc: "Replace with an equivalent regular expression literal with flags 'u'.",
              output: '/a/u;',
            },
          ],
        },
      ],
    },
    {
      code: 'new RegExp(/a/ /* comment */);',
      options: [
        {
          disallowRedundantWrapping: true,
        },
      ],
      errors: [
        {
          messageId: 'unexpectedRedundantRegExp',
          message:
            "Regular expression literal is unnecessarily wrapped within a 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: "new RegExp(/a/, 'd');",
      options: [
        {
          disallowRedundantWrapping: true,
        },
      ],
      errors: [
        {
          messageId: 'unexpectedRedundantRegExpWithFlags',
          message:
            "Use regular expression literal with flags instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'replaceWithLiteralAndFlags',
              desc: "Replace with an equivalent regular expression literal with flags 'd'.",
              output: '/a/d;',
            },
          ],
        },
      ],
    },
    {
      code: `(a)
new RegExp(/b/);`,
      options: [
        {
          disallowRedundantWrapping: true,
        },
      ],
      errors: [
        {
          messageId: 'unexpectedRedundantRegExp',
          message:
            "Regular expression literal is unnecessarily wrapped within a 'RegExp' constructor.",
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 16,
        },
      ],
    },
    {
      code: `(a)
new RegExp(/b/, 'g');`,
      options: [
        {
          disallowRedundantWrapping: true,
        },
      ],
      errors: [
        {
          messageId: 'unexpectedRedundantRegExpWithFlags',
          message:
            "Use regular expression literal with flags instead of the 'RegExp' constructor.",
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 21,
        },
      ],
    },
    {
      code: 'a/RegExp(/foo/);',
      options: [
        {
          disallowRedundantWrapping: true,
        },
      ],
      errors: [
        {
          messageId: 'unexpectedRedundantRegExp',
          message:
            "Regular expression literal is unnecessarily wrapped within a 'RegExp' constructor.",
          line: 1,
          column: 3,
          endLine: 1,
          endColumn: 16,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: 'a/ /foo/;',
            },
          ],
        },
      ],
    },
    {
      code: 'RegExp(/foo/)in a;',
      options: [
        {
          disallowRedundantWrapping: true,
        },
      ],
      errors: [
        {
          messageId: 'unexpectedRedundantRegExp',
          message:
            "Regular expression literal is unnecessarily wrapped within a 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/foo/ in a;',
            },
          ],
        },
      ],
    },
    {
      code: 'new RegExp((String?.raw)`a`);',
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/a/;',
            },
          ],
        },
      ],
    },
    {
      code: "new RegExp('+');",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: "new RegExp('*');",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: "RegExp('+');",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    {
      code: "RegExp('*');",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    {
      code: "new RegExp('+', 'g');",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: "new RegExp('*', 'g');",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: "RegExp('+', 'g');",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: "RegExp('*', 'g');",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: "RegExp('abc', 'u');",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/abc/u;',
            },
          ],
        },
      ],
    },
    {
      code: "new RegExp('abc', 'd');",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/abc/d;',
            },
          ],
        },
      ],
    },
    {
      code: "RegExp('abc', 'd');",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/abc/d;',
            },
          ],
        },
      ],
    },
    {
      code: "RegExp('\\\\\\\\', '');",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/\\\\/;',
            },
          ],
        },
      ],
    },
    {
      code: "RegExp('\\n', '');",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 17,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/\\n/;',
            },
          ],
        },
      ],
    },
    {
      code: "RegExp('\\n\\n', '');",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/\\n\\n/;',
            },
          ],
        },
      ],
    },
    {
      code: "RegExp('\\t', '');",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 17,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/\\t/;',
            },
          ],
        },
      ],
    },
    {
      code: "RegExp('\\t\\t', '');",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/\\t\\t/;',
            },
          ],
        },
      ],
    },
    {
      code: "RegExp('\\r\\n', '');",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/\\r\\n/;',
            },
          ],
        },
      ],
    },
    {
      code: "RegExp('\\u1234', 'g')",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: "RegExp('\\u{1234}', 'g')",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: "RegExp('\\u{11111}', 'g')",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: "RegExp('\\v', '');",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 17,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/\\v/;',
            },
          ],
        },
      ],
    },
    {
      code: "RegExp('\\v\\v', '');",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/\\v\\v/;',
            },
          ],
        },
      ],
    },
    {
      code: "RegExp('\\f', '');",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 17,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/\\f/;',
            },
          ],
        },
      ],
    },
    {
      code: "RegExp('\\f\\f', '');",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/\\f\\f/;',
            },
          ],
        },
      ],
    },
    {
      code: "RegExp('\\\\b', '');",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 18,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/\\b/;',
            },
          ],
        },
      ],
    },
    {
      code: "RegExp('\\\\b\\\\b', '');",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/\\b\\b/;',
            },
          ],
        },
      ],
    },
    {
      code: "new RegExp('\\\\B\\\\b', '');",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/\\B\\b/;',
            },
          ],
        },
      ],
    },
    {
      code: "RegExp('\\\\w', '');",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 18,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/\\w/;',
            },
          ],
        },
      ],
    },
    {
      code: "new globalThis.RegExp('\\\\W', '');",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 33,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/\\W/;',
            },
          ],
        },
      ],
    },
    {
      code: "RegExp('\\\\s', '');",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 18,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/\\s/;',
            },
          ],
        },
      ],
    },
    {
      code: "new RegExp('\\\\S', '')",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/\\S/',
            },
          ],
        },
      ],
    },
    {
      code: "globalThis.RegExp('\\\\d', '');",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/\\d/;',
            },
          ],
        },
      ],
    },
    {
      code: "globalThis.RegExp('\\\\D', '')",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/\\D/',
            },
          ],
        },
      ],
    },
    {
      code: "globalThis.RegExp('\\\\\\\\\\\\D', '')",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 33,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/\\\\\\D/',
            },
          ],
        },
      ],
    },
    {
      code: "new RegExp('\\\\D\\\\D', '')",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/\\D\\D/',
            },
          ],
        },
      ],
    },
    {
      code: "new globalThis.RegExp('\\\\0\\\\0', '');",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 36,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/\\0\\0/;',
            },
          ],
        },
      ],
    },
    {
      code: "new RegExp('\\\\0\\\\0', '');",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/\\0\\0/;',
            },
          ],
        },
      ],
    },
    {
      code: "new RegExp('\\0\\0', 'g');",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: "RegExp('\\\\0\\\\0\\\\0', '')",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/\\0\\0\\0/',
            },
          ],
        },
      ],
    },
    {
      code: "RegExp('\\\\78\\\\126\\\\5934', '')",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/\\78\\126\\5934/',
            },
          ],
        },
      ],
    },
    {
      code: "new window['RegExp']('\\\\x56\\\\x78\\\\x45', '');",
      languageOptions: {
        globals: {
          window: 'readonly',
        },
      },
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 44,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/\\x56\\x78\\x45/;',
            },
          ],
        },
      ],
    },
    {
      code: "a in(RegExp('abc'))",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 19,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: 'a in(/abc/)',
            },
          ],
        },
      ],
    },
    {
      code: `x = y
            RegExp("foo").test(x) ? bar() : baz()`,
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 2,
          column: 13,
          endLine: 2,
          endColumn: 26,
        },
      ],
    },
    {
      code: "func(new RegExp(String.raw`\\w{1, 2`, 'u'),new RegExp(String.raw`\\w{1, 2`, 'u'))",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 42,
        },
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 43,
          endLine: 1,
          endColumn: 79,
        },
      ],
    },
    {
      code: `x = y;
            RegExp("foo").test(x) ? bar() : baz()`,
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 2,
          column: 13,
          endLine: 2,
          endColumn: 26,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: `x = y;
            /foo/.test(x) ? bar() : baz()`,
            },
          ],
        },
      ],
    },
    {
      code: 'typeof RegExp("foo")',
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: 'typeof /foo/',
            },
          ],
        },
      ],
    },
    {
      code: "RegExp(\"foo\") instanceof RegExp(String.raw`blahblah`, 'g') ? typeof new RegExp('(\\\\p{Emoji_Presentation})\\\\1', `ug`) : false",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output:
                "/foo/ instanceof RegExp(String.raw`blahblah`, 'g') ? typeof new RegExp('(\\\\p{Emoji_Presentation})\\\\1', `ug`) : false",
            },
          ],
        },
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 59,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output:
                'RegExp("foo") instanceof /blahblah/g ? typeof new RegExp(\'(\\\\p{Emoji_Presentation})\\\\1\', `ug`) : false',
            },
          ],
        },
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 69,
          endLine: 1,
          endColumn: 117,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output:
                'RegExp("foo") instanceof RegExp(String.raw`blahblah`, \'g\') ? typeof /(\\p{Emoji_Presentation})\\1/ug : false',
            },
          ],
        },
      ],
    },
    {
      code: '[   new RegExp(`someregular`)]',
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 30,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '[   /someregular/]',
            },
          ],
        },
      ],
    },
    {
      code: `const totallyValidatesEmails = new RegExp("\\\\S+@(\\\\S+\\\\.)+\\\\S+")
            if (typeof totallyValidatesEmails === 'object') {
                runSomethingThatExists(Regexp('stuff'))
            }`,
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 32,
          endLine: 1,
          endColumn: 65,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: `const totallyValidatesEmails = /\\S+@(\\S+\\.)+\\S+/
            if (typeof totallyValidatesEmails === 'object') {
                runSomethingThatExists(Regexp('stuff'))
            }`,
            },
          ],
        },
      ],
    },
    {
      code: "!new RegExp('^Hey, ', 'u') && new RegExp('jk$') && ~new RegExp('^Sup, ') || new RegExp('hi') + new RegExp('person') === -new RegExp('hi again') ? 5 * new RegExp('abc') : 'notregbutstring'",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 2,
          endLine: 1,
          endColumn: 27,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output:
                "!/^Hey, /u && new RegExp('jk$') && ~new RegExp('^Sup, ') || new RegExp('hi') + new RegExp('person') === -new RegExp('hi again') ? 5 * new RegExp('abc') : 'notregbutstring'",
            },
          ],
        },
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 31,
          endLine: 1,
          endColumn: 48,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output:
                "!new RegExp('^Hey, ', 'u') && /jk$/ && ~new RegExp('^Sup, ') || new RegExp('hi') + new RegExp('person') === -new RegExp('hi again') ? 5 * new RegExp('abc') : 'notregbutstring'",
            },
          ],
        },
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 53,
          endLine: 1,
          endColumn: 73,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output:
                "!new RegExp('^Hey, ', 'u') && new RegExp('jk$') && ~/^Sup, / || new RegExp('hi') + new RegExp('person') === -new RegExp('hi again') ? 5 * new RegExp('abc') : 'notregbutstring'",
            },
          ],
        },
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 77,
          endLine: 1,
          endColumn: 93,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output:
                "!new RegExp('^Hey, ', 'u') && new RegExp('jk$') && ~new RegExp('^Sup, ') || /hi/ + new RegExp('person') === -new RegExp('hi again') ? 5 * new RegExp('abc') : 'notregbutstring'",
            },
          ],
        },
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 96,
          endLine: 1,
          endColumn: 116,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output:
                "!new RegExp('^Hey, ', 'u') && new RegExp('jk$') && ~new RegExp('^Sup, ') || new RegExp('hi') + /person/ === -new RegExp('hi again') ? 5 * new RegExp('abc') : 'notregbutstring'",
            },
          ],
        },
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 122,
          endLine: 1,
          endColumn: 144,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output:
                "!new RegExp('^Hey, ', 'u') && new RegExp('jk$') && ~new RegExp('^Sup, ') || new RegExp('hi') + new RegExp('person') === -/hi again/ ? 5 * new RegExp('abc') : 'notregbutstring'",
            },
          ],
        },
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 151,
          endLine: 1,
          endColumn: 168,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output:
                "!new RegExp('^Hey, ', 'u') && new RegExp('jk$') && ~new RegExp('^Sup, ') || new RegExp('hi') + new RegExp('person') === -new RegExp('hi again') ? 5 * /abc/ : 'notregbutstring'",
            },
          ],
        },
      ],
    },
    {
      code: `#!/usr/bin/sh
            RegExp("foo")`,
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 2,
          column: 13,
          endLine: 2,
          endColumn: 26,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: `#!/usr/bin/sh
            /foo/`,
            },
          ],
        },
      ],
    },
    {
      code: 'async function abc(){await new RegExp("foo")}',
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 28,
          endLine: 1,
          endColumn: 45,
        },
      ],
    },
    {
      code: 'function* abc(){yield new RegExp("foo")}',
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      code: 'function* abc(){yield* new RegExp("foo")}',
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 41,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: 'function* abc(){yield* /foo/}',
            },
          ],
        },
      ],
    },
    {
      code: "console.log({ ...new RegExp('a') })",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 33,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: 'console.log({ .../a/ })',
            },
          ],
        },
      ],
    },
    {
      code: "delete RegExp('a');",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 19,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: 'delete /a/;',
            },
          ],
        },
      ],
    },
    {
      code: "void RegExp('a');",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 17,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: 'void /a/;',
            },
          ],
        },
      ],
    },
    {
      code: 'new RegExp("\\\\S+@(\\\\S+\\\\.)+\\\\S+")**RegExp(\'a\')',
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 34,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: "/\\S+@(\\S+\\.)+\\S+/**RegExp('a')",
            },
          ],
        },
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 36,
          endLine: 1,
          endColumn: 47,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: 'new RegExp("\\\\S+@(\\\\S+\\\\.)+\\\\S+")**/a/',
            },
          ],
        },
      ],
    },
    {
      code: 'new RegExp("\\\\S+@(\\\\S+\\\\.)+\\\\S+")%RegExp(\'a\')',
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 34,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: "/\\S+@(\\S+\\.)+\\S+/%RegExp('a')",
            },
          ],
        },
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 35,
          endLine: 1,
          endColumn: 46,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: 'new RegExp("\\\\S+@(\\\\S+\\\\.)+\\\\S+")%/a/',
            },
          ],
        },
      ],
    },
    {
      code: "a in RegExp('abc')",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 19,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: 'a in /abc/',
            },
          ],
        },
      ],
    },
    {
      code: "\n            /abc/ == new RegExp('cba');\n            ",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 2,
          column: 22,
          endLine: 2,
          endColumn: 39,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '\n            /abc/ == /cba/;\n            ',
            },
          ],
        },
      ],
    },
    {
      code: "\n            /abc/ === new RegExp('cba');\n            ",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 2,
          column: 23,
          endLine: 2,
          endColumn: 40,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '\n            /abc/ === /cba/;\n            ',
            },
          ],
        },
      ],
    },
    {
      code: "\n            /abc/ != new RegExp('cba');\n            ",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 2,
          column: 22,
          endLine: 2,
          endColumn: 39,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '\n            /abc/ != /cba/;\n            ',
            },
          ],
        },
      ],
    },
    {
      code: "\n            /abc/ !== new RegExp('cba');\n            ",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 2,
          column: 23,
          endLine: 2,
          endColumn: 40,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '\n            /abc/ !== /cba/;\n            ',
            },
          ],
        },
      ],
    },
    {
      code: "\n            /abc/ > new RegExp('cba');\n            ",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 2,
          column: 21,
          endLine: 2,
          endColumn: 38,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '\n            /abc/ > /cba/;\n            ',
            },
          ],
        },
      ],
    },
    {
      code: "\n            /abc/ < new RegExp('cba');\n            ",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 2,
          column: 21,
          endLine: 2,
          endColumn: 38,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '\n            /abc/ < /cba/;\n            ',
            },
          ],
        },
      ],
    },
    {
      code: "\n            /abc/ >= new RegExp('cba');\n            ",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 2,
          column: 22,
          endLine: 2,
          endColumn: 39,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '\n            /abc/ >= /cba/;\n            ',
            },
          ],
        },
      ],
    },
    {
      code: "\n            /abc/ <= new RegExp('cba');\n            ",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 2,
          column: 22,
          endLine: 2,
          endColumn: 39,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '\n            /abc/ <= /cba/;\n            ',
            },
          ],
        },
      ],
    },
    {
      code: "\n            /abc/ << new RegExp('cba');\n            ",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 2,
          column: 22,
          endLine: 2,
          endColumn: 39,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '\n            /abc/ << /cba/;\n            ',
            },
          ],
        },
      ],
    },
    {
      code: "\n            /abc/ >> new RegExp('cba');\n            ",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 2,
          column: 22,
          endLine: 2,
          endColumn: 39,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '\n            /abc/ >> /cba/;\n            ',
            },
          ],
        },
      ],
    },
    {
      code: "\n            /abc/ >>> new RegExp('cba');\n            ",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 2,
          column: 23,
          endLine: 2,
          endColumn: 40,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '\n            /abc/ >>> /cba/;\n            ',
            },
          ],
        },
      ],
    },
    {
      code: "\n            /abc/ ^ new RegExp('cba');\n            ",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 2,
          column: 21,
          endLine: 2,
          endColumn: 38,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '\n            /abc/ ^ /cba/;\n            ',
            },
          ],
        },
      ],
    },
    {
      code: "\n            /abc/ & new RegExp('cba');\n            ",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 2,
          column: 21,
          endLine: 2,
          endColumn: 38,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '\n            /abc/ & /cba/;\n            ',
            },
          ],
        },
      ],
    },
    {
      code: "\n            /abc/ | new RegExp('cba');\n            ",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 2,
          column: 21,
          endLine: 2,
          endColumn: 38,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '\n            /abc/ | /cba/;\n            ',
            },
          ],
        },
      ],
    },
    {
      code: "\n            null ?? new RegExp('blah')\n            ",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 2,
          column: 21,
          endLine: 2,
          endColumn: 39,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '\n            null ?? /blah/\n            ',
            },
          ],
        },
      ],
    },
    {
      code: "\n            abc *= new RegExp('blah')\n            ",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 2,
          column: 20,
          endLine: 2,
          endColumn: 38,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '\n            abc *= /blah/\n            ',
            },
          ],
        },
      ],
    },
    {
      code: "\n            console.log({a: new RegExp('sup')})\n            ",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 2,
          column: 29,
          endLine: 2,
          endColumn: 46,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '\n            console.log({a: /sup/})\n            ',
            },
          ],
        },
      ],
    },
    {
      code: "\n            console.log(() => {new RegExp('sup')})\n            ",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 2,
          column: 32,
          endLine: 2,
          endColumn: 49,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '\n            console.log(() => {/sup/})\n            ',
            },
          ],
        },
      ],
    },
    {
      code: "\n            function abc() {new RegExp('sup')}\n            ",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 2,
          column: 29,
          endLine: 2,
          endColumn: 46,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '\n            function abc() {/sup/}\n            ',
            },
          ],
        },
      ],
    },
    {
      code: "\n            function abc() {return new RegExp('sup')}\n            ",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 2,
          column: 36,
          endLine: 2,
          endColumn: 53,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output:
                '\n            function abc() {return /sup/}\n            ',
            },
          ],
        },
      ],
    },
    {
      code: "\n            abc <<= new RegExp('cba');\n            ",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 2,
          column: 21,
          endLine: 2,
          endColumn: 38,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '\n            abc <<= /cba/;\n            ',
            },
          ],
        },
      ],
    },
    {
      code: "\n            abc >>= new RegExp('cba');\n            ",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 2,
          column: 21,
          endLine: 2,
          endColumn: 38,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '\n            abc >>= /cba/;\n            ',
            },
          ],
        },
      ],
    },
    {
      code: "\n            abc >>>= new RegExp('cba');\n            ",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 2,
          column: 22,
          endLine: 2,
          endColumn: 39,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '\n            abc >>>= /cba/;\n            ',
            },
          ],
        },
      ],
    },
    {
      code: "\n            abc ^= new RegExp('cba');\n            ",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 2,
          column: 20,
          endLine: 2,
          endColumn: 37,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '\n            abc ^= /cba/;\n            ',
            },
          ],
        },
      ],
    },
    {
      code: "\n            abc &= new RegExp('cba');\n            ",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 2,
          column: 20,
          endLine: 2,
          endColumn: 37,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '\n            abc &= /cba/;\n            ',
            },
          ],
        },
      ],
    },
    {
      code: "\n            abc |= new RegExp('cba');\n            ",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 2,
          column: 20,
          endLine: 2,
          endColumn: 37,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '\n            abc |= /cba/;\n            ',
            },
          ],
        },
      ],
    },
    {
      code: "\n            abc ??= new RegExp('cba');\n            ",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 2,
          column: 21,
          endLine: 2,
          endColumn: 38,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '\n            abc ??= /cba/;\n            ',
            },
          ],
        },
      ],
    },
    {
      code: "\n            abc &&= new RegExp('cba');\n            ",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 2,
          column: 21,
          endLine: 2,
          endColumn: 38,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '\n            abc &&= /cba/;\n            ',
            },
          ],
        },
      ],
    },
    {
      code: "\n            abc ||= new RegExp('cba');\n            ",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 2,
          column: 21,
          endLine: 2,
          endColumn: 38,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '\n            abc ||= /cba/;\n            ',
            },
          ],
        },
      ],
    },
    {
      code: "\n            abc **= new RegExp('blah')\n            ",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 2,
          column: 21,
          endLine: 2,
          endColumn: 39,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '\n            abc **= /blah/\n            ',
            },
          ],
        },
      ],
    },
    {
      code: "\n            abc /= new RegExp('blah')\n            ",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 2,
          column: 20,
          endLine: 2,
          endColumn: 38,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '\n            abc /= /blah/\n            ',
            },
          ],
        },
      ],
    },
    {
      code: "\n            abc += new RegExp('blah')\n            ",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 2,
          column: 20,
          endLine: 2,
          endColumn: 38,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '\n            abc += /blah/\n            ',
            },
          ],
        },
      ],
    },
    {
      code: "\n            abc -= new RegExp('blah')\n            ",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 2,
          column: 20,
          endLine: 2,
          endColumn: 38,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '\n            abc -= /blah/\n            ',
            },
          ],
        },
      ],
    },
    {
      code: "\n            abc %= new RegExp('blah')\n            ",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 2,
          column: 20,
          endLine: 2,
          endColumn: 38,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '\n            abc %= /blah/\n            ',
            },
          ],
        },
      ],
    },
    {
      code: "\n            () => new RegExp('blah')\n            ",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 2,
          column: 19,
          endLine: 2,
          endColumn: 37,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '\n            () => /blah/\n            ',
            },
          ],
        },
      ],
    },
    {
      code: 'a/RegExp("foo")in b',
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 3,
          endLine: 1,
          endColumn: 16,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: 'a/ /foo/ in b',
            },
          ],
        },
      ],
    },
    {
      code: 'a/RegExp("foo")instanceof b',
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 3,
          endLine: 1,
          endColumn: 16,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: 'a/ /foo/ instanceof b',
            },
          ],
        },
      ],
    },
    {
      code: `do RegExp("foo")
while (true);`,
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 17,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: `do /foo/
while (true);`,
            },
          ],
        },
      ],
    },
    {
      code: `for(let i;i<5;i++) { break
new RegExp('search')}`,
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: `for(let i;i<5;i++) { break
/search/}`,
            },
          ],
        },
      ],
    },
    {
      code: `for(let i;i<5;i++) { continue
new RegExp('search')}`,
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: `for(let i;i<5;i++) { continue
/search/}`,
            },
          ],
        },
      ],
    },
    {
      code: "\n            switch (value) {\n                case \"possibility\":\n                    console.log('possibility matched')\n                case RegExp('myReg').toString():\n                    console.log('matches a regexp\\' toString value')\n                    break;\n            }\n            ",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 5,
          column: 22,
          endLine: 5,
          endColumn: 37,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output:
                "\n            switch (value) {\n                case \"possibility\":\n                    console.log('possibility matched')\n                case /myReg/.toString():\n                    console.log('matches a regexp\\' toString value')\n                    break;\n            }\n            ",
            },
          ],
        },
      ],
    },
    {
      code: "throw new RegExp('abcdefg') // fail with a regular expression",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 28,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: 'throw /abcdefg/ // fail with a regular expression',
            },
          ],
        },
      ],
    },
    {
      code: "for (value of new RegExp('something being searched')) { console.log(value) }",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 53,
        },
      ],
    },
    {
      code: "(async function(){for await (value of new RegExp('something being searched')) { console.log(value) }})()",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 39,
          endLine: 1,
          endColumn: 77,
        },
      ],
    },
    {
      code: "for (value in new RegExp('something being searched')) { console.log(value) }",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 53,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output:
                'for (value in /something being searched/) { console.log(value) }',
            },
          ],
        },
      ],
    },
    {
      code: "if (condition1 && condition2) new RegExp('avalue').test(str);",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 31,
          endLine: 1,
          endColumn: 51,
        },
      ],
    },
    {
      code: `debugger
new RegExp('myReg')`,
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 20,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: `debugger
/myReg/`,
            },
          ],
        },
      ],
    },
    {
      code: 'RegExp("\\\\\\n")',
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 15,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/\\n/',
            },
          ],
        },
      ],
    },
    {
      code: 'RegExp("\\\\\\t")',
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 15,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/\\t/',
            },
          ],
        },
      ],
    },
    {
      code: 'RegExp("\\\\\\f")',
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 15,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/\\f/',
            },
          ],
        },
      ],
    },
    {
      code: 'RegExp("\\\\\\v")',
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 15,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/\\v/',
            },
          ],
        },
      ],
    },
    {
      code: 'RegExp("\\\\\\r")',
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 15,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/\\r/',
            },
          ],
        },
      ],
    },
    {
      code: 'new RegExp("\t")',
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 16,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/\\t/',
            },
          ],
        },
      ],
    },
    {
      code: 'new RegExp("/")',
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 16,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/\\//',
            },
          ],
        },
      ],
    },
    {
      code: 'new RegExp("\\.")',
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 17,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/./',
            },
          ],
        },
      ],
    },
    {
      code: 'new RegExp("\\\\.")',
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 18,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/\\./',
            },
          ],
        },
      ],
    },
    {
      code: 'new RegExp("\\\\\\n\\\\\\n")',
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/\\n\\n/',
            },
          ],
        },
      ],
    },
    {
      code: 'new RegExp("\\\\\\n\\\\\\f\\\\\\n")',
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 27,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/\\n\\f\\n/',
            },
          ],
        },
      ],
    },
    {
      code: 'new RegExp("\\u000A\\u000A");',
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 27,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/\\n\\n/;',
            },
          ],
        },
      ],
    },
    {
      code: "new RegExp('mysafereg' /* comment explaining its safety */)",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 60,
        },
      ],
    },
    {
      code: "new RegExp('[[A--B]]', 'v')",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 28,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/[[A--B]]/v',
            },
          ],
        },
      ],
    },
    {
      code: "new RegExp('[[A--B]]', 'v')",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 28,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/[[A--B]]/v',
            },
          ],
        },
      ],
    },
    {
      code: "new RegExp('[[A&&&]]', 'v')",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: "new RegExp('a', 'uv')",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: "new RegExp(/a/, 'v')",
      options: [
        {
          disallowRedundantWrapping: true,
        },
      ],
      errors: [
        {
          messageId: 'unexpectedRedundantRegExpWithFlags',
          message:
            "Use regular expression literal with flags instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'replaceWithLiteralAndFlags',
              desc: "Replace with an equivalent regular expression literal with flags 'v'.",
              output: '/a/v',
            },
          ],
        },
      ],
    },
    {
      code: "new RegExp(/a/, 'v')",
      options: [
        {
          disallowRedundantWrapping: true,
        },
      ],
      errors: [
        {
          messageId: 'unexpectedRedundantRegExpWithFlags',
          message:
            "Use regular expression literal with flags instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'replaceWithLiteralAndFlags',
              desc: "Replace with an equivalent regular expression literal with flags 'v'.",
              output: '/a/v',
            },
          ],
        },
      ],
    },
    {
      code: "new RegExp(/a/g, 'v')",
      options: [
        {
          disallowRedundantWrapping: true,
        },
      ],
      errors: [
        {
          messageId: 'unexpectedRedundantRegExpWithFlags',
          message:
            "Use regular expression literal with flags instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
          suggestions: [
            {
              messageId: 'replaceWithLiteralAndFlags',
              desc: "Replace with an equivalent regular expression literal with flags 'v'.",
              output: '/a/v',
            },
            {
              messageId: 'replaceWithIntendedLiteralAndFlags',
              desc: "Replace with a regular expression literal with flags 'gv'.",
              output: '/a/gv',
            },
          ],
        },
      ],
    },
    {
      code: "new RegExp(/[[A--B]]/v, 'g')",
      options: [
        {
          disallowRedundantWrapping: true,
        },
      ],
      errors: [
        {
          messageId: 'unexpectedRedundantRegExpWithFlags',
          message:
            "Use regular expression literal with flags instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
          suggestions: [
            {
              messageId: 'replaceWithIntendedLiteralAndFlags',
              desc: "Replace with a regular expression literal with flags 'vg'.",
              output: '/[[A--B]]/vg',
            },
          ],
        },
      ],
    },
    {
      code: "new RegExp(/a/u, 'v')",
      options: [
        {
          disallowRedundantWrapping: true,
        },
      ],
      errors: [
        {
          messageId: 'unexpectedRedundantRegExpWithFlags',
          message:
            "Use regular expression literal with flags instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
          suggestions: [
            {
              messageId: 'replaceWithLiteralAndFlags',
              desc: "Replace with an equivalent regular expression literal with flags 'v'.",
              output: '/a/v',
            },
          ],
        },
      ],
    },
    {
      code: "new RegExp(/a/v, 'u')",
      options: [
        {
          disallowRedundantWrapping: true,
        },
      ],
      errors: [
        {
          messageId: 'unexpectedRedundantRegExpWithFlags',
          message:
            "Use regular expression literal with flags instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
          suggestions: [
            {
              messageId: 'replaceWithLiteralAndFlags',
              desc: "Replace with an equivalent regular expression literal with flags 'u'.",
              output: '/a/u',
            },
          ],
        },
      ],
    },
    {
      code: "new RegExp(/[[A--B]]/v, 'u')",
      options: [
        {
          disallowRedundantWrapping: true,
        },
      ],
      errors: [
        {
          messageId: 'unexpectedRedundantRegExpWithFlags',
          message:
            "Use regular expression literal with flags instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: "new RegExp('(?i:foo)bar')",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/(?i:foo)bar/',
            },
          ],
        },
      ],
    },
    {
      code: "new RegExp('(?i:foo)bar')",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: '/(?i:foo)bar/',
            },
          ],
        },
      ],
    },
    {
      code: "var regex = new RegExp('foo', 'u');",
      errors: [
        {
          messageId: 'unexpectedRegExp',
          message:
            "Use a regular expression literal instead of the 'RegExp' constructor.",
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 35,
          suggestions: [
            {
              messageId: 'replaceWithLiteral',
              desc: 'Replace with an equivalent regular expression literal.',
              output: 'var regex = /foo/u;',
            },
          ],
        },
      ],
    },
  ],
});

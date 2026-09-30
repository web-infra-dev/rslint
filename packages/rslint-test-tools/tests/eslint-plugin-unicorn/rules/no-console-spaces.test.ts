// Ported from eslint-plugin-unicorn v76.0.0 tests, snapshots and documentation.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v76.0.0/test/no-console-spaces.js
import { RuleTester } from '../rule-tester';

new RuleTester().run('no-console-spaces', {} as never, {
  valid: [
    // Upstream
    'console.log("abc");',
    'console.log("abc", "def");',
    'console.log(\'abc\', "def");',
    'console.log(`abc`, "def");',
    'console.log(`\nabc\ndef\n`);',
    'console.log(\' \', "def");',
    'console.log(" ");',
    'console.log(" ", "b");',
    'console.log("a", " ");',
    'console.log(" ", "b", "c");',
    'console.log("a", " ", "c");',
    'console.log("a", "b", " ");',
    'console.log(\'  \', "def");',
    'console.log("abc  ", "def");',
    'console.log("abc\\t", "def");',
    'console.log("abc\\n", "def");',
    'console.log("  abc", "def");',
    'console.log(" abc", "def");',
    'console.log("abc", "def ");',
    'console.log();',
    'console.log("");',
    'console.log(123);',
    'console.log(null);',
    'console.log(undefined);',
    'console.dir("abc ");',
    'new console.log(" a ", " b ");',
    'new console.debug(" a ", " b ");',
    'new console.info(" a ", " b ");',
    'new console.warn(" a ", " b ");',
    'new console.error(" a ", " b ");',
    'log(" a ", " b ");',
    'debug(" a ", " b ");',
    'info(" a ", " b ");',
    'warn(" a ", " b ");',
    'error(" a ", " b ");',
    'console["log"](" a ", " b ");',
    'console["debug"](" a ", " b ");',
    'console["info"](" a ", " b ");',
    'console["warn"](" a ", " b ");',
    'console["error"](" a ", " b ");',
    'console[log](" a ", " b ");',
    'console[debug](" a ", " b ");',
    'console[info](" a ", " b ");',
    'console[warn](" a ", " b ");',
    'console[error](" a ", " b ");',
    'console.foo(" a ", " b ");',
    'foo.log(" a ", " b ");',
    'foo.debug(" a ", " b ");',
    'foo.info(" a ", " b ");',
    'foo.warn(" a ", " b ");',
    'foo.error(" a ", " b ");',
    'lib.console.log(" a ", " b ");',
    'lib.console.debug(" a ", " b ");',
    'lib.console.info(" a ", " b ");',
    'lib.console.warn(" a ", " b ");',
    'lib.console.error(" a ", " b ");',
    // Snapshots

    // Documentation
    "console.log('abc', 'def');",
    "console.debug('abc', 'def');",
    "console.info('abc', 'def');",
    "console.warn('abc', 'def');",
    "console.error('abc', 'def');",
    "console.log('abc ');",
    "console.log(' abc');",
    "console.log('abc  ', 'def');",
    "console.log('abc\\t', 'def');",
    "console.log('abc\\n', 'def');",
  ].map((code) => ({ code, filename: 'src/virtual.js' })),
  invalid: [
    // Upstream
    {
      code: 'console.log("abc ", "def");',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-console-spaces',
          message:
            'Do not use trailing space between `console.log` parameters.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 18,
        },
      ],
      output: 'console.log("abc", "def");',
    },
    {
      code: 'console.log("abc", " def");',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-console-spaces',
          message: 'Do not use leading space between `console.log` parameters.',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 22,
        },
      ],
      output: 'console.log("abc", "def");',
    },
    {
      code: 'console.log(" abc ", "def");',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-console-spaces',
          message:
            'Do not use trailing space between `console.log` parameters.',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 19,
        },
      ],
      output: 'console.log(" abc", "def");',
    },
    {
      code: 'console.debug("abc ", "def");',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-console-spaces',
          message:
            'Do not use trailing space between `console.debug` parameters.',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 20,
        },
      ],
      output: 'console.debug("abc", "def");',
    },
    {
      code: 'console.info("abc ", "def");',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-console-spaces',
          message:
            'Do not use trailing space between `console.info` parameters.',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 19,
        },
      ],
      output: 'console.info("abc", "def");',
    },
    {
      code: 'console.warn("abc ", "def");',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-console-spaces',
          message:
            'Do not use trailing space between `console.warn` parameters.',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 19,
        },
      ],
      output: 'console.warn("abc", "def");',
    },
    {
      code: 'console.error("abc ", "def");',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-console-spaces',
          message:
            'Do not use trailing space between `console.error` parameters.',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 20,
        },
      ],
      output: 'console.error("abc", "def");',
    },
    {
      code: 'console.log("abc", " def ", "ghi");',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-console-spaces',
          message: 'Do not use leading space between `console.log` parameters.',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 22,
        },
        {
          messageId: 'no-console-spaces',
          message:
            'Do not use trailing space between `console.log` parameters.',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 26,
        },
      ],
      output: 'console.log("abc", "def", "ghi");',
    },
    {
      code: 'console.log("abc ", "def ", "ghi");',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-console-spaces',
          message:
            'Do not use trailing space between `console.log` parameters.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 18,
        },
        {
          messageId: 'no-console-spaces',
          message:
            'Do not use trailing space between `console.log` parameters.',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 26,
        },
      ],
      output: 'console.log("abc", "def", "ghi");',
    },
    {
      code: 'console.log(\'abc \', "def");',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-console-spaces',
          message:
            'Do not use trailing space between `console.log` parameters.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 18,
        },
      ],
      output: 'console.log(\'abc\', "def");',
    },
    {
      code: 'console.log(`abc `, "def");',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-console-spaces',
          message:
            'Do not use trailing space between `console.log` parameters.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 18,
        },
      ],
      output: 'console.log(`abc`, "def");',
    },
    {
      code: 'console.log(`abc ${1 + 2} `, "def");',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-console-spaces',
          message:
            'Do not use trailing space between `console.log` parameters.',
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 27,
        },
      ],
      output: 'console.log(`abc ${1 + 2}`, "def");',
    },
    {
      code: "console.log(\n\t'abc',\n\t'def ',\n\t'ghi'\n);",
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-console-spaces',
          message:
            'Do not use trailing space between `console.log` parameters.',
          line: 3,
          column: 6,
          endLine: 3,
          endColumn: 7,
        },
      ],
      output: "console.log(\n\t'abc',\n\t'def',\n\t'ghi'\n);",
    },
    {
      code: "console.error(\n\ttheme.error('✗'),\n\t'Verifying \"packaging\" fixture\\n ',\n\ttheme.error(errorMessage)\n);",
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-console-spaces',
          message:
            'Do not use trailing space between `console.error` parameters.',
          line: 3,
          column: 34,
          endLine: 3,
          endColumn: 35,
        },
      ],
      output:
        "console.error(\n\ttheme.error('✗'),\n\t'Verifying \"packaging\" fixture\\n',\n\ttheme.error(errorMessage)\n);",
    },
    // Snapshots
    {
      code: 'console.log("abc", " def ", "ghi");',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-console-spaces',
          message: 'Do not use leading space between `console.log` parameters.',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 22,
        },
        {
          messageId: 'no-console-spaces',
          message:
            'Do not use trailing space between `console.log` parameters.',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 26,
        },
      ],
      output: 'console.log("abc", "def", "ghi");',
    },
    {
      code: "console.error(\n\ttheme.error('✗'),\n\t'Verifying \"packaging\" fixture\\n ',\n\ttheme.error(errorMessage)\n);",
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-console-spaces',
          message:
            'Do not use trailing space between `console.error` parameters.',
          line: 3,
          column: 34,
          endLine: 3,
          endColumn: 35,
        },
      ],
      output:
        "console.error(\n\ttheme.error('✗'),\n\t'Verifying \"packaging\" fixture\\n',\n\ttheme.error(errorMessage)\n);",
    },
    {
      code: "console.log(\n\t'abc',\n\t'def ',\n\t'ghi'\n);",
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-console-spaces',
          message:
            'Do not use trailing space between `console.log` parameters.',
          line: 3,
          column: 6,
          endLine: 3,
          endColumn: 7,
        },
      ],
      output: "console.log(\n\t'abc',\n\t'def',\n\t'ghi'\n);",
    },
    {
      code: 'console.log("_", " leading", "_")',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-console-spaces',
          message: 'Do not use leading space between `console.log` parameters.',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 20,
        },
      ],
      output: 'console.log("_", "leading", "_")',
    },
    {
      code: 'console.log("_", "trailing ", "_")',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-console-spaces',
          message:
            'Do not use trailing space between `console.log` parameters.',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 28,
        },
      ],
      output: 'console.log("_", "trailing", "_")',
    },
    {
      code: 'console.log("_", " leading and trailing ", "_")',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-console-spaces',
          message: 'Do not use leading space between `console.log` parameters.',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 20,
        },
        {
          messageId: 'no-console-spaces',
          message:
            'Do not use trailing space between `console.log` parameters.',
          line: 1,
          column: 40,
          endLine: 1,
          endColumn: 41,
        },
      ],
      output: 'console.log("_", "leading and trailing", "_")',
    },
    {
      code: 'console.log("_", " log ", "_")',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-console-spaces',
          message: 'Do not use leading space between `console.log` parameters.',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 20,
        },
        {
          messageId: 'no-console-spaces',
          message:
            'Do not use trailing space between `console.log` parameters.',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 24,
        },
      ],
      output: 'console.log("_", "log", "_")',
    },
    {
      code: 'console.debug("_", " debug ", "_")',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-console-spaces',
          message:
            'Do not use leading space between `console.debug` parameters.',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 22,
        },
        {
          messageId: 'no-console-spaces',
          message:
            'Do not use trailing space between `console.debug` parameters.',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 28,
        },
      ],
      output: 'console.debug("_", "debug", "_")',
    },
    {
      code: 'console.info("_", " info ", "_")',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-console-spaces',
          message:
            'Do not use leading space between `console.info` parameters.',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 21,
        },
        {
          messageId: 'no-console-spaces',
          message:
            'Do not use trailing space between `console.info` parameters.',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 26,
        },
      ],
      output: 'console.info("_", "info", "_")',
    },
    {
      code: 'console.warn("_", " warn ", "_")',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-console-spaces',
          message:
            'Do not use leading space between `console.warn` parameters.',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 21,
        },
        {
          messageId: 'no-console-spaces',
          message:
            'Do not use trailing space between `console.warn` parameters.',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 26,
        },
      ],
      output: 'console.warn("_", "warn", "_")',
    },
    {
      code: 'console.error("_", " error ", "_")',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-console-spaces',
          message:
            'Do not use leading space between `console.error` parameters.',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 22,
        },
        {
          messageId: 'no-console-spaces',
          message:
            'Do not use trailing space between `console.error` parameters.',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 28,
        },
      ],
      output: 'console.error("_", "error", "_")',
    },
    // Documentation
    {
      code: "console.log('abc ', 'def');",
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-console-spaces',
          message:
            'Do not use trailing space between `console.log` parameters.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 18,
        },
      ],
      output: "console.log('abc', 'def');",
    },
    {
      code: "console.log('abc', ' def');",
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-console-spaces',
          message: 'Do not use leading space between `console.log` parameters.',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 22,
        },
      ],
      output: "console.log('abc', 'def');",
    },
    {
      code: 'console.log("abc ", " def");',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-console-spaces',
          message:
            'Do not use trailing space between `console.log` parameters.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 18,
        },
        {
          messageId: 'no-console-spaces',
          message: 'Do not use leading space between `console.log` parameters.',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 23,
        },
      ],
      output: 'console.log("abc", "def");',
    },
    {
      code: 'console.log(`abc `, ` def`);',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-console-spaces',
          message:
            'Do not use trailing space between `console.log` parameters.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 18,
        },
        {
          messageId: 'no-console-spaces',
          message: 'Do not use leading space between `console.log` parameters.',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 23,
        },
      ],
      output: 'console.log(`abc`, `def`);',
    },
    {
      code: "console.debug('abc ', 'def');",
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-console-spaces',
          message:
            'Do not use trailing space between `console.debug` parameters.',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 20,
        },
      ],
      output: "console.debug('abc', 'def');",
    },
    {
      code: "console.info('abc ', 'def');",
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-console-spaces',
          message:
            'Do not use trailing space between `console.info` parameters.',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 19,
        },
      ],
      output: "console.info('abc', 'def');",
    },
    {
      code: "console.warn('abc ', 'def');",
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-console-spaces',
          message:
            'Do not use trailing space between `console.warn` parameters.',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 19,
        },
      ],
      output: "console.warn('abc', 'def');",
    },
    {
      code: "console.error('abc ', 'def');",
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-console-spaces',
          message:
            'Do not use trailing space between `console.error` parameters.',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 20,
        },
      ],
      output: "console.error('abc', 'def');",
    },
  ],
});

import { RuleTester } from '@rslint/rule-tester';

// Upstream: typescript-eslint v8.67.0 (10 valid, 17 invalid cases).
// This rule supplies neither autofixes nor suggestions.

const ruleTester = new RuleTester({
  plugin: '@typescript-eslint',
  filename: 'input.ts',
});

ruleTester.run('no-var-requires', {
  valid: [
    "import foo = require('foo');",
    "require('foo');",
    "require?.('foo');",
    `
import { createRequire } from 'module';
const require = createRequire('foo');
const json = require('./some.json');
    `,
    {
      code: "const pkg = require('./package.json');",
      options: [{ allow: ['/package\\.json$'] }],
    },
    {
      code: "const pkg = require('../package.json');",
      options: [{ allow: ['/package\\.json$'] }],
    },
    {
      code: "const pkg = require('../packages/package.json');",
      options: [{ allow: ['/package\\.json$'] }],
    },
    {
      code: "const pkg = require('data.json');",
      options: [{ allow: ['\\.json$'] }],
    },
    {
      code: "const pkg = require('some-package');",
      options: [{ allow: ['^some-package$'] }],
    },
    {
      code: 'const pkg = require(`some-package`);',
      options: [{ allow: ['^some-package$'] }],
    },
  ],
  invalid: [
    {
      code: "var foo = require('foo');",
      errors: [
        {
          column: 11,
          endColumn: 25,
          endLine: 1,
          line: 1,
          messageId: 'noVarReqs',
          message: 'Require statement not part of import statement.',
        },
      ],
    },
    {
      code: "const foo = require('foo');",
      errors: [
        {
          column: 13,
          endColumn: 27,
          endLine: 1,
          line: 1,
          messageId: 'noVarReqs',
          message: 'Require statement not part of import statement.',
        },
      ],
    },
    {
      code: "let foo = require('foo');",
      errors: [
        {
          column: 11,
          endColumn: 25,
          endLine: 1,
          line: 1,
          messageId: 'noVarReqs',
          message: 'Require statement not part of import statement.',
        },
      ],
    },
    {
      code: "let foo = trick(require('foo'));",
      errors: [
        {
          column: 17,
          endColumn: 31,
          endLine: 1,
          line: 1,
          messageId: 'noVarReqs',
          message: 'Require statement not part of import statement.',
        },
      ],
    },
    {
      code: "var foo = require?.('foo');",
      errors: [
        {
          column: 11,
          endColumn: 27,
          endLine: 1,
          line: 1,
          messageId: 'noVarReqs',
          message: 'Require statement not part of import statement.',
        },
      ],
    },
    {
      code: "const foo = require?.('foo');",
      errors: [
        {
          column: 13,
          endColumn: 29,
          endLine: 1,
          line: 1,
          messageId: 'noVarReqs',
          message: 'Require statement not part of import statement.',
        },
      ],
    },
    {
      code: "let foo = require?.('foo');",
      errors: [
        {
          column: 11,
          endColumn: 27,
          endLine: 1,
          line: 1,
          messageId: 'noVarReqs',
          message: 'Require statement not part of import statement.',
        },
      ],
    },
    {
      code: "let foo = trick(require?.('foo'));",
      errors: [
        {
          column: 17,
          endColumn: 33,
          endLine: 1,
          line: 1,
          messageId: 'noVarReqs',
          message: 'Require statement not part of import statement.',
        },
      ],
    },
    {
      code: "let foo = trick?.(require('foo'));",
      errors: [
        {
          column: 19,
          endColumn: 33,
          endLine: 1,
          line: 1,
          messageId: 'noVarReqs',
          message: 'Require statement not part of import statement.',
        },
      ],
    },
    {
      code: "const foo = require('./foo.json') as Foo;",
      errors: [
        {
          column: 13,
          endColumn: 34,
          endLine: 1,
          line: 1,
          messageId: 'noVarReqs',
          message: 'Require statement not part of import statement.',
        },
      ],
    },
    {
      code: "const foo = <Foo>require('./foo.json');",
      errors: [
        {
          column: 18,
          endColumn: 39,
          endLine: 1,
          line: 1,
          messageId: 'noVarReqs',
          message: 'Require statement not part of import statement.',
        },
      ],
    },
    {
      code: "const foo: Foo = require('./foo.json').default;",
      errors: [
        {
          column: 18,
          endColumn: 39,
          endLine: 1,
          line: 1,
          messageId: 'noVarReqs',
          message: 'Require statement not part of import statement.',
        },
      ],
    },
    {
      // https://github.com/typescript-eslint/typescript-eslint/issues/3883
      code: `
const configValidator = new Validator(require('./a.json'));
configValidator.addSchema(require('./a.json'));
      `,
      errors: [
        {
          column: 39,
          endColumn: 58,
          endLine: 2,
          line: 2,
          messageId: 'noVarReqs',
          message: 'Require statement not part of import statement.',
        },
        {
          column: 27,
          endColumn: 46,
          endLine: 3,
          line: 3,
          messageId: 'noVarReqs',
          message: 'Require statement not part of import statement.',
        },
      ],
    },
    {
      code: "const pkg = require('./package.json');",
      errors: [
        {
          column: 13,
          endColumn: 38,
          endLine: 1,
          line: 1,
          messageId: 'noVarReqs',
          message: 'Require statement not part of import statement.',
        },
      ],
    },
    {
      code: "const pkg = require('./package.jsonc');",
      errors: [
        {
          column: 13,
          endColumn: 39,
          endLine: 1,
          line: 1,
          messageId: 'noVarReqs',
          message: 'Require statement not part of import statement.',
        },
      ],
      options: [{ allow: ['/package\\.json$'] }],
    },
    {
      code: "const pkg = require('./package.json');",
      errors: [
        {
          column: 13,
          endColumn: 38,
          endLine: 1,
          line: 1,
          messageId: 'noVarReqs',
          message: 'Require statement not part of import statement.',
        },
      ],
      options: [{ allow: ['^some-package$'] }],
    },
    {
      code: 'const pkg = require(`./package.json`);',
      errors: [
        {
          column: 13,
          endColumn: 38,
          endLine: 1,
          line: 1,
          messageId: 'noVarReqs',
          message: 'Require statement not part of import statement.',
        },
      ],
      options: [{ allow: ['^some-package$'] }],
    },
  ],
});

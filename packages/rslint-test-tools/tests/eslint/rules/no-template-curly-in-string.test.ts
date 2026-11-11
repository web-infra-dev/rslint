// Every case from ESLint v10.8.1, with messages and full ranges
// additionally pinned from the same upstream version. No upstream expectations are relaxed.
// https://github.com/eslint/eslint/blob/v10.8.1/tests/lib/rules/no-template-curly-in-string.js
import { RuleTester } from '@rslint/rule-tester';

const ruleTester = new RuleTester({
  plugin: '',
});

ruleTester.run('no-template-curly-in-string', {
  valid: [
    {
      code: '`Hello, ${name}`;',
    },
    {
      code: 'templateFunction`Hello, ${name}`;',
    },
    {
      code: '`Hello, name`;',
    },
    {
      code: "'Hello, name';",
    },
    {
      code: "'Hello, ' + name;",
    },
    {
      code: '`Hello, ${index + 1}`',
    },
    {
      code: '`Hello, ${name + " foo"}`',
    },
    {
      code: '`Hello, ${name || "foo"}`',
    },
    {
      code: '`Hello, ${{foo: "bar"}.foo}`',
    },
    {
      code: "'$2'",
    },
    {
      code: "'${'",
    },
    {
      code: "'$}'",
    },
    {
      code: "'{foo}'",
    },
    {
      code: '\'{foo: "bar"}\'',
    },
    {
      code: 'const number = 3',
    },
  ],
  invalid: [
    {
      code: "'Hello, ${name}'",
      errors: [
        {
          messageId: 'unexpectedTemplateExpression',
          message: 'Unexpected template string expression.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: '"Hello, ${name}"',
      errors: [
        {
          messageId: 'unexpectedTemplateExpression',
          message: 'Unexpected template string expression.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: "'${greeting}, ${name}'",
      errors: [
        {
          messageId: 'unexpectedTemplateExpression',
          message: 'Unexpected template string expression.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: "'Hello, ${index + 1}'",
      errors: [
        {
          messageId: 'unexpectedTemplateExpression',
          message: 'Unexpected template string expression.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '\'Hello, ${name + " foo"}\'',
      errors: [
        {
          messageId: 'unexpectedTemplateExpression',
          message: 'Unexpected template string expression.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '\'Hello, ${name || "foo"}\'',
      errors: [
        {
          messageId: 'unexpectedTemplateExpression',
          message: 'Unexpected template string expression.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '\'Hello, ${{foo: "bar"}.foo}\'',
      errors: [
        {
          messageId: 'unexpectedTemplateExpression',
          message: 'Unexpected template string expression.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
  ],
});

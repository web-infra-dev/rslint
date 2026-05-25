import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('eol-last', null as never, {
  valid: [
    // default 'always'
    { code: '' },
    { code: '\n' },
    { code: 'var a = 123;\n' },
    { code: 'var a = 123;\n\n' },
    { code: 'var a = 123;\n   \n' },

    { code: '\r\n' },
    { code: 'var a = 123;\r\n' },
    { code: 'var a = 123;\r\n\r\n' },
    { code: 'var a = 123;\r\n   \r\n' },

    // 'never'
    { code: 'var a = 123;', options: ['never'] },
    { code: 'var a = 123;\nvar b = 456;', options: ['never'] },
    { code: 'var a = 123;\r\nvar b = 456;', options: ['never'] },
  ],

  invalid: [
    // default 'always' — missing trailing newline
    {
      code: 'var a = 123;',
      output: 'var a = 123;\n',
      errors: [{ messageId: 'missing', line: 1, column: 13 }],
    },
    {
      code: 'var a = 123;\n   ',
      output: 'var a = 123;\n   \n',
      errors: [{ messageId: 'missing', line: 2, column: 4 }],
    },

    // 'never' — unexpected trailing newline
    {
      code: 'var a = 123;\n',
      output: 'var a = 123;',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpected',
          line: 1,
          column: 13,
          endLine: 2,
          endColumn: 1,
        },
      ],
    },
    {
      code: 'var a = 123;\r\n',
      output: 'var a = 123;',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpected',
          line: 1,
          column: 13,
          endLine: 2,
          endColumn: 1,
        },
      ],
    },
    {
      code: 'var a = 123;\r\n\r\n',
      output: 'var a = 123;',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpected',
          line: 2,
          column: 1,
          endLine: 3,
          endColumn: 1,
        },
      ],
    },
    {
      code: 'var a = 123;\nvar b = 456;\n',
      output: 'var a = 123;\nvar b = 456;',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpected',
          line: 2,
          column: 13,
          endLine: 3,
          endColumn: 1,
        },
      ],
    },
    {
      code: 'var a = 123;\r\nvar b = 456;\r\n',
      output: 'var a = 123;\r\nvar b = 456;',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpected',
          line: 2,
          column: 13,
          endLine: 3,
          endColumn: 1,
        },
      ],
    },
    {
      code: 'var a = 123;\n\n',
      output: 'var a = 123;',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpected',
          line: 2,
          column: 1,
          endLine: 3,
          endColumn: 1,
        },
      ],
    },
  ],
});

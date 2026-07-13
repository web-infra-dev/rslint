import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-empty-static-block', {
  valid: [
    'class Foo { static { bar(); } }',
    'class Foo { static { /* comments */ } }',
    'class Foo { static {\n// comment\n} }',
    'class Foo { static { bar(); } static { bar(); } }',
  ],
  invalid: [
    {
      code: 'class Foo { static {} }',
      errors: [
        {
          messageId: 'unexpected',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: 'class Foo { static { } }',
      errors: [
        {
          messageId: 'unexpected',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: 'class Foo { static { \n\n } }',
      errors: [
        {
          messageId: 'unexpected',
          line: 1,
          column: 20,
          endLine: 3,
          endColumn: 3,
        },
      ],
    },
    {
      code: 'class Foo { static { bar(); } static {} }',
      errors: [
        {
          messageId: 'unexpected',
          line: 1,
          column: 38,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      code: 'class Foo { static // comment\n {} }',
      errors: [
        {
          messageId: 'unexpected',
          line: 2,
          column: 2,
          endLine: 2,
          endColumn: 4,
        },
      ],
    },
    {
      code: 'class Foo { static /* empty */ {} /* empty */ }',
      errors: [
        {
          messageId: 'unexpected',
          line: 1,
          column: 32,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
  ],
});

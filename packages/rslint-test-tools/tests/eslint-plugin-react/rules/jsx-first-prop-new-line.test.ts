import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('jsx-first-prop-new-line', {} as never, {
  valid: [
    {
      code: `<Foo bar="baz" />`,
      options: ['never'],
    },
    {
      code: `<Foo\n  bar="baz"\n/>`,
      options: ['always'],
    },
  ],
  invalid: [
    {
      code: `<Foo bar="baz"\n  qux="quux"\n/>`,
      options: ['always'],
      errors: [{ message: 'Property should be placed on a new line' }],
    },
    {
      code: `<Foo\n  bar="baz"\n/>`,
      options: ['never'],
      errors: [
        {
          message:
            'Property should be placed on the same line as the component declaration',
        },
      ],
    },
  ],
});

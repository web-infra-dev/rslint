import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('jsx-max-props-per-line', {} as never, {
  valid: [
    {
      code: `<Foo bar="baz" />`,
    },
    {
      code: `<Foo bar="baz" qux="quux" />`,
      options: [{ maximum: 2 }],
    },
  ],
  invalid: [
    {
      code: `<Foo bar="baz" qux="quux" />`,
      errors: [{ message: 'Prop `qux` must be placed on a new line' }],
    },
    {
      // ESLint reports one error per line at the first excess prop
      code: `<Foo bar="baz" qux="quux" abc="def" />`,
      options: [{ maximum: 1 }],
      errors: [{ message: 'Prop `qux` must be placed on a new line' }],
    },
  ],
});

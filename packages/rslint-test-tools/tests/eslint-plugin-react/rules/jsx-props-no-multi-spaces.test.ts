import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('jsx-props-no-multi-spaces', {} as never, {
  valid: [
    {
      code: `<Foo bar="baz" />`,
    },
    {
      code: `<Foo bar="baz" qux="quux" />`,
    },
  ],
  invalid: [
    {
      code: `<Foo  bar="baz" />`,
      errors: [{ message: 'Expected only one space between "Foo" and "bar"' }],
    },
    {
      code: `<Foo bar="baz"  qux="quux" />`,
      errors: [{ message: 'Expected only one space between "bar" and "qux"' }],
    },
  ],
});

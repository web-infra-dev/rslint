import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('jsx-equals-spacing', {} as never, {
  valid: [
    {
      code: `<Foo name="value" />`,
      options: ['never'],
    },
    {
      code: `<Foo name = "value" />`,
      options: ['always'],
    },
  ],
  invalid: [
    {
      code: `<Foo name ="value" />`,
      options: ['never'],
      errors: [{ message: "There should be no space before '='" }],
    },
    {
      code: `<Foo name="value" />`,
      options: ['always'],
      errors: [
        { message: "A space is required before '='" },
        { message: "A space is required after '='" },
      ],
    },
  ],
});

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('jsx-boolean-value', {} as never, {
  valid: [
    {
      code: `<Foo disabled />`,
    },
    {
      code: `<Foo disabled />`,
      options: ['never'],
    },
    {
      code: `<Foo disabled={true} />`,
      options: ['always'],
    },
  ],
  invalid: [
    {
      code: `<Foo disabled={true} />`,
      options: ['never'],
      errors: [
        { message: 'Value must be omitted for boolean attribute `disabled`' },
      ],
    },
    {
      code: `<Foo disabled />`,
      options: ['always'],
      errors: [
        { message: 'Value must be set for boolean attribute `disabled`' },
      ],
    },
  ],
});

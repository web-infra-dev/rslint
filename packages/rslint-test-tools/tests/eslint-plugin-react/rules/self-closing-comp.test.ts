import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('self-closing-comp', {} as never, {
  valid: [
    {
      code: `<Foo>bar</Foo>`,
    },
    {
      code: `<div>children</div>`,
    },
    {
      code: `<Foo />`,
    },
  ],
  invalid: [
    {
      code: `<Foo></Foo>`,
      errors: [{ message: 'Empty components are self-closing' }],
    },
    {
      code: `<div></div>`,
      errors: [{ message: 'Empty components are self-closing' }],
    },
  ],
});

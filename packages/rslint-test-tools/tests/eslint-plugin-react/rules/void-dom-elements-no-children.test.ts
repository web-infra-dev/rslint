import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('void-dom-elements-no-children', {} as never, {
  valid: [
    {
      code: `<div>Children</div>`,
    },
    {
      code: `<br />`,
    },
    {
      code: `<img src="image.png" />`,
    },
  ],
  invalid: [
    {
      code: `<br>Children</br>`,
      errors: [{ message: 'Void DOM element <br /> cannot receive children.' }],
    },
    {
      code: `<img children="foo" />`,
      errors: [
        { message: 'Void DOM element <img /> cannot receive children.' },
      ],
    },
  ],
});

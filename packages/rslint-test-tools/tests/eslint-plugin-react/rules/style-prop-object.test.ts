import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('style-prop-object', {} as never, {
  valid: [
    {
      code: `<div style={{ color: 'red' }} />`,
    },
    {
      code: `const s = { color: 'red' }; <div style={s} />`,
    },
    {
      code: `<div className="foo" />`,
    },
  ],
  invalid: [
    {
      code: `<div style="color: red" />`,
      errors: [{ message: 'Style prop value must be an object' }],
    },
    {
      code: `<div style={42} />`,
      errors: [{ message: 'Style prop value must be an object' }],
    },
  ],
});

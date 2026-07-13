import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('jsx-props-no-spread-multi', {} as never, {
  valid: [
    {
      code: `
        const a = {};
        <App {...a} />
      `,
    },
    {
      code: `
        const a = {};
        const b = {};
        <App {...a} {...b} />
      `,
    },
  ],
  invalid: [
    {
      code: `
        const props = {};
        <App {...props} {...props} />
      `,
      errors: [
        {
          message: 'Spreading the same expression multiple times is forbidden',
        },
      ],
    },
    {
      code: `
        const props = {};
        <div {...props} a="a" {...props} />
      `,
      errors: [
        {
          message: 'Spreading the same expression multiple times is forbidden',
        },
      ],
    },
    {
      code: `
        const props = {};
        <div {...props} {...props} {...props} />
      `,
      errors: [
        {
          message: 'Spreading the same expression multiple times is forbidden',
        },
        {
          message: 'Spreading the same expression multiple times is forbidden',
        },
      ],
    },
  ],
});

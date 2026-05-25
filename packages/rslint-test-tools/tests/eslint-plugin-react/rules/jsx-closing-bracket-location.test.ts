import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('jsx-closing-bracket-location', {} as never, {
  valid: [
    { code: `<App />` },
    { code: `<App foo />` },
    { code: `<App\n  foo\n/>` },
    {
      code: `<App\n  foo />`,
      options: ['after-props'],
    },
    {
      code: `<App\n  foo\n  />`,
      options: ['props-aligned'],
    },
    {
      code: `<App foo></App>`,
    },
    {
      code: `<App>\n  <Foo\n    bar\n  >\n  </Foo>\n  <Foo\n    bar />\n</App>`,
      options: [{ nonEmpty: false, selfClosing: 'after-props' }],
    },
  ],
  invalid: [
    {
      code: `<App\n/>`,
      options: ['tag-aligned'],
      errors: [
        { message: 'The closing bracket must be placed after the opening tag' },
      ],
    },
    {
      code: `<App foo\n/>`,
      options: ['tag-aligned'],
      errors: [
        { message: 'The closing bracket must be placed after the last prop' },
      ],
    },
    {
      code: `<App\n  foo />`,
      options: ['tag-aligned'],
      errors: [
        {
          message:
            'The closing bracket must be aligned with the opening tag (expected column 1 on the next line)',
        },
      ],
    },
    {
      code: `<App\n  foo />`,
      options: ['props-aligned'],
      errors: [
        {
          message:
            'The closing bracket must be aligned with the last prop (expected column 3 on the next line)',
        },
      ],
    },
  ],
});

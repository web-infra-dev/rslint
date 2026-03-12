import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('jsx-closing-tag-location', {} as never, {
  valid: [
    {
      code: `<Foo>bar</Foo>`,
    },
    {
      code: `<Foo>\n  bar\n</Foo>`,
    },
  ],
  invalid: [
    {
      code: `<Foo>\n  bar\n  </Foo>`,
      errors: [
        { message: 'Expected closing tag to match indentation of opening.' },
      ],
    },
  ],
});

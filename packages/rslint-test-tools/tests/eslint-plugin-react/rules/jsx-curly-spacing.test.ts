import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('jsx-curly-spacing', {} as never, {
  valid: [
    { code: `<App foo={bar} />;` },
    { code: `<App foo={bar} />;`, options: [{ when: 'never' }] },
    { code: `<App foo={ bar } />;`, options: [{ when: 'always' }] },
    { code: `<App foo={ bar } />;`, options: ['always'] },
    { code: `<App foo={bar} />;`, options: ['never'] },
    { code: `<App {...bar} />;` },
    { code: `<App { ...bar } />;`, options: ['always'] },
    {
      code: `<App>{bar}</App>;`,
      options: [{ children: { when: 'never' } }],
    },
    {
      code: `<App>{ bar }</App>;`,
      options: [{ children: { when: 'always' } }],
    },
    {
      code: `<App foo={{ a: 1 }} />;`,
      options: [{ when: 'never' }],
    },
    {
      code: `<App foo={ {a: 1} } />;`,
      options: ['never', { spacing: { objectLiterals: 'always' } }],
    },
  ],
  invalid: [
    {
      code: `<App foo={ bar } />;`,
      options: ['never'],
      errors: [
        { message: "There should be no space after '{'" },
        { message: "There should be no space before '}'" },
      ],
    },
    {
      code: `<App foo={bar} />;`,
      options: ['always'],
      errors: [
        { message: "A space is required after '{'" },
        { message: "A space is required before '}'" },
      ],
    },
    {
      code: `<App {...bar} />;`,
      options: ['always'],
      errors: [
        { message: "A space is required after '{'" },
        { message: "A space is required before '}'" },
      ],
    },
    {
      code: `<App>{ bar }</App>;`,
      options: [{ children: { when: 'never' } }],
      errors: [
        { message: "There should be no space after '{'" },
        { message: "There should be no space before '}'" },
      ],
    },
    {
      code: `<App foo={{ a: 1 }} />;`,
      options: ['never', { spacing: { objectLiterals: 'always' } }],
      errors: [
        { message: "A space is required after '{'" },
        { message: "A space is required before '}'" },
      ],
    },
  ],
});

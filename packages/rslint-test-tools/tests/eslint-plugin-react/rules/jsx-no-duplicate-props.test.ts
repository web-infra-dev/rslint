import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('jsx-no-duplicate-props', {} as never, {
  valid: [
    { code: `<App />;` },
    { code: `<App {...this.props} />;` },
    { code: `<App a b c />;` },
    { code: `<App a b c A />;` },
    { code: `<App {...this.props} a b c />;` },
    { code: `<App c {...this.props} a b />;` },
    { code: `<App a="c" b="b" c="a" />;` },
    { code: `<App {...this.props} a="c" b="b" c="a" />;` },
    { code: `<App c="a" {...this.props} a="c" b="b" />;` },
    { code: `<App A a />;` },
    { code: `<App A b a />;` },
    { code: `<App A="a" b="b" B="B" />;` },
    { code: `<App a:b="c" />;`, options: [{ ignoreCase: true }] },
    // Non-self-closing element without duplicates.
    { code: `<App a b></App>;` },
    // Spread between distinct props is fine.
    { code: `<App a {...x} b />;` },
  ],
  invalid: [
    {
      code: `<App a a />;`,
      errors: [{ message: 'No duplicate props allowed' }],
    },
    {
      code: `<App A b c A />;`,
      errors: [{ message: 'No duplicate props allowed' }],
    },
    {
      code: `<App a="a" b="b" a="a" />;`,
      errors: [{ message: 'No duplicate props allowed' }],
    },
    {
      code: `<App A a />;`,
      options: [{ ignoreCase: true }],
      errors: [{ message: 'No duplicate props allowed' }],
    },
    {
      code: `<App a b c A />;`,
      options: [{ ignoreCase: true }],
      errors: [{ message: 'No duplicate props allowed' }],
    },
    {
      code: `<App A="a" b="b" B="B" />;`,
      options: [{ ignoreCase: true }],
      errors: [{ message: 'No duplicate props allowed' }],
    },
    // Non-self-closing element with duplicates.
    {
      code: `<App a a></App>;`,
      errors: [{ message: 'No duplicate props allowed' }],
    },
    // Spread does not reset duplicate tracking.
    {
      code: `<App a {...x} a />;`,
      errors: [{ message: 'No duplicate props allowed' }],
    },
    // Three duplicates — each subsequent dup reports.
    {
      code: `<App a a a />;`,
      errors: [
        { message: 'No duplicate props allowed' },
        { message: 'No duplicate props allowed' },
      ],
    },
  ],
});

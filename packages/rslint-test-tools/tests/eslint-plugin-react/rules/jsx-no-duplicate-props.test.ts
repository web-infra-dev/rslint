import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

function expectNoDuplicatePropsError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'noDuplicateProps',
    message: 'No duplicate props allowed',
    line,
    column,
    endLine,
    endColumn,
  };
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('jsx-no-duplicate-props', {
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
      errors: [expectNoDuplicatePropsError([1, 8, 1, 9])],
    },
    {
      code: `<App A b c A />;`,
      errors: [expectNoDuplicatePropsError([1, 12, 1, 13])],
    },
    {
      code: `<App a="a" b="b" a="a" />;`,
      errors: [expectNoDuplicatePropsError([1, 18, 1, 23])],
    },
    {
      code: `<App A a />;`,
      options: [{ ignoreCase: true }],
      errors: [expectNoDuplicatePropsError([1, 8, 1, 9])],
    },
    {
      code: `<App a b c A />;`,
      options: [{ ignoreCase: true }],
      errors: [expectNoDuplicatePropsError([1, 12, 1, 13])],
    },
    {
      code: `<App A="a" b="b" B="B" />;`,
      options: [{ ignoreCase: true }],
      errors: [expectNoDuplicatePropsError([1, 18, 1, 23])],
    },
    // Non-self-closing element with duplicates.
    {
      code: `<App a a></App>;`,
      errors: [expectNoDuplicatePropsError([1, 8, 1, 9])],
    },
    // Spread does not reset duplicate tracking.
    {
      code: `<App a {...x} a />;`,
      errors: [expectNoDuplicatePropsError([1, 15, 1, 16])],
    },
    // Three duplicates — each subsequent dup reports.
    {
      code: `<App a a a />;`,
      errors: [
        expectNoDuplicatePropsError([1, 8, 1, 9]),
        expectNoDuplicatePropsError([1, 10, 1, 11]),
      ],
    },
  ],
});

import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

function expectNoMultiSpreadingError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'noMultiSpreading',
    message: 'Spreading the same expression multiple times is forbidden',
    line,
    column,
    endLine,
    endColumn,
  };
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('jsx-props-no-spread-multi', {
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
      errors: [expectNoMultiSpreadingError([3, 25, 3, 35])],
    },
    {
      code: `
        const props = {};
        <div {...props} a="a" {...props} />
      `,
      errors: [expectNoMultiSpreadingError([3, 31, 3, 41])],
    },
    {
      code: `
        const props = {};
        <div {...props} {...props} {...props} />
      `,
      errors: [
        expectNoMultiSpreadingError([3, 25, 3, 35]),
        expectNoMultiSpreadingError([3, 36, 3, 46]),
      ],
    },
  ],
});

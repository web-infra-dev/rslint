import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

new RuleTester(config).run('padding-around-all', {
  valid: [{ code: `setup();\n\nbeforeAll(connect);\n\ntest('works', run);` }],
  invalid: [
    {
      code: `setup();\nbeforeAll(connect);\ntest('works', run);`,
      output: `setup();\n\nbeforeAll(connect);\n\ntest('works', run);`,
      errors: [
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 20,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 20,
        },
      ],
    },
  ],
});

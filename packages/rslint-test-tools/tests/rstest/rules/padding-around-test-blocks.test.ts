import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

new RuleTester(config).run('padding-around-test-blocks', {
  valid: [{ code: `setup();\n\ntest('works', run);` }],
  invalid: [
    {
      code: `setup();\ntest('works', run);`,
      output: `setup();\n\ntest('works', run);`,
      errors: [
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 20,
        },
      ],
    },
  ],
});

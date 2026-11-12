import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

new RuleTester(config).run('padding-around-describe-blocks', {
  valid: [{ code: `setup();\n\ndescribe('suite', run);` }],
  invalid: [
    {
      code: `setup();\ndescribe('suite', run);`,
      output: `setup();\n\ndescribe('suite', run);`,
      errors: [
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 24,
        },
      ],
    },
  ],
});

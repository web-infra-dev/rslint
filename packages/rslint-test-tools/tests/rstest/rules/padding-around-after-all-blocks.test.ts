import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

new RuleTester(config).run('padding-around-after-all-blocks', {
  valid: [{ code: 'setup();\n\nafterAll(cleanup);' }],
  invalid: [
    {
      code: 'setup();\nafterAll(cleanup);',
      output: 'setup();\n\nafterAll(cleanup);',
      errors: [
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 19,
        },
      ],
    },
  ],
});

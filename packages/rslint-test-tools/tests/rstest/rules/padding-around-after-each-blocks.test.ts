import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

new RuleTester(config).run('padding-around-after-each-blocks', {
  valid: [{ code: 'setup();\n\nafterEach(cleanup);' }],
  invalid: [
    {
      code: 'setup();\nafterEach(cleanup);',
      output: 'setup();\n\nafterEach(cleanup);',
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

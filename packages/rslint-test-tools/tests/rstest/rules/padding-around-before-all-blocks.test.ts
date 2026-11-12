import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

new RuleTester(config).run('padding-around-before-all-blocks', {
  valid: [{ code: 'setup();\n\nbeforeAll(connect);' }],
  invalid: [
    {
      code: 'setup();\nbeforeAll(connect);',
      output: 'setup();\n\nbeforeAll(connect);',
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

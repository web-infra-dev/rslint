import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

new RuleTester(config).run('padding-around-before-each-blocks', {
  valid: [{ code: 'setup();\n\nbeforeEach(reset);' }],
  invalid: [
    {
      code: 'setup();\nbeforeEach(reset);',
      output: 'setup();\n\nbeforeEach(reset);',
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

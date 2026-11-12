import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

new RuleTester(config).run('padding-around-expect-groups', {
  valid: [{ code: 'const value = load();\n\nexpect(value).toBeDefined();' }],
  invalid: [
    {
      code: 'const value = load();\nexpect(value).toBeDefined();',
      output: 'const value = load();\n\nexpect(value).toBeDefined();',
      errors: [
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 29,
        },
      ],
    },
  ],
});

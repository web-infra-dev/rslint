// Runs @rslint/rule-tester's suggestion checks against a rule that reports
// suggestions. Expectations follow eslint-plugin-unicorn v72.0.0.
import { RuleTester } from '@rslint/rule-tester';

new RuleTester({ plugin: 'unicorn' }).run('no-negation-in-equality-check', {
  valid: ['foo === bar', { code: '!(foo === bar)' }],
  invalid: [
    {
      code: '!foo === bar',
      errors: [
        {
          messageId: 'no-negation-in-equality-check/error',
          message: 'Negated expression is not allowed in equality check.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 2,
          suggestions: [
            {
              messageId: 'no-negation-in-equality-check/suggestion',
              desc: "Switch to '!==' check.",
              output: 'foo !== bar',
            },
          ],
        },
      ],
    },
  ],
});

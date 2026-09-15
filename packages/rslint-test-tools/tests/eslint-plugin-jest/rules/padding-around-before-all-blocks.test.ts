import { RuleTester } from '../rule-tester';

new RuleTester().run('padding-around-before-all-blocks', {} as never, {
  valid: ['setup();\n\nbeforeAll(connect);'],
  invalid: [
    {
      code: 'setup();\nbeforeAll(connect);',
      output: 'setup();\n\nbeforeAll(connect);',
      errors: 1,
    },
  ],
});

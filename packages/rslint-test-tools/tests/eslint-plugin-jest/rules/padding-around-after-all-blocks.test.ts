import { RuleTester } from '../rule-tester';

new RuleTester().run('padding-around-after-all-blocks', {} as never, {
  valid: ['setup();\n\nafterAll(cleanup);'],
  invalid: [
    {
      code: 'setup();\nafterAll(cleanup);',
      output: 'setup();\n\nafterAll(cleanup);',
      errors: 1,
    },
  ],
});

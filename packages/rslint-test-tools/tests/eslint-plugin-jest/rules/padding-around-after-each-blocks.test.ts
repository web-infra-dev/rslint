import { RuleTester } from '../rule-tester';

new RuleTester().run('padding-around-after-each-blocks', {} as never, {
  valid: ['setup();\n\nafterEach(cleanup);'],
  invalid: [
    {
      code: 'setup();\nafterEach(cleanup);',
      output: 'setup();\n\nafterEach(cleanup);',
      errors: 1,
    },
  ],
});

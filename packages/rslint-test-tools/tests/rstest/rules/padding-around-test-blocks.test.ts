import { RuleTester } from '../rule-tester';

new RuleTester().run('padding-around-test-blocks', {} as never, {
  valid: [{ code: `setup();\n\ntest('works', run);` }],
  invalid: [
    {
      code: `setup();\ntest('works', run);`,
      output: `setup();\n\ntest('works', run);`,
      errors: 1,
    },
  ],
});

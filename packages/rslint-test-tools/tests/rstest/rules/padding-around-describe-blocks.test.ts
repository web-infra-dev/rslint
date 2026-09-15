import { RuleTester } from '../rule-tester';

new RuleTester().run('padding-around-describe-blocks', {} as never, {
  valid: [{ code: `setup();\n\ndescribe('suite', run);` }],
  invalid: [
    {
      code: `setup();\ndescribe('suite', run);`,
      output: `setup();\n\ndescribe('suite', run);`,
      errors: 1,
    },
  ],
});

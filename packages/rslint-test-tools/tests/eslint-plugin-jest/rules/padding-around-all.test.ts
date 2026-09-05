import { RuleTester } from '../rule-tester';

new RuleTester().run('padding-around-all', {} as never, {
  valid: [`setup();\n\nbeforeAll(connect);\n\ntest('works', run);`],
  invalid: [
    {
      code: `setup();\nbeforeAll(connect);\ntest('works', run);`,
      output: `setup();\n\nbeforeAll(connect);\n\ntest('works', run);`,
      errors: 2,
    },
  ],
});

import { RuleTester } from '../rule-tester';

new RuleTester().run('padding-around-before-each-blocks', {} as never, {
  valid: [{ code: 'setup();\n\nbeforeEach(reset);' }],
  invalid: [
    {
      code: 'setup();\nbeforeEach(reset);',
      output: 'setup();\n\nbeforeEach(reset);',
      errors: 1,
    },
  ],
});

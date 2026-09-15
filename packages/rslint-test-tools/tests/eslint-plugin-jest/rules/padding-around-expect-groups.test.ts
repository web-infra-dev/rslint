import { RuleTester } from '../rule-tester';

new RuleTester().run('padding-around-expect-groups', {} as never, {
  valid: ['const value = load();\n\nexpect(value).toBeDefined();'],
  invalid: [
    {
      code: 'const value = load();\nexpect(value).toBeDefined();',
      output: 'const value = load();\n\nexpect(value).toBeDefined();',
      errors: 1,
    },
  ],
});

import { RuleTester } from '@typescript-eslint/rule-tester';
import { getFixturesRootDir } from '../RuleTester.ts';

const rootPath = getFixturesRootDir();

const ruleTester = new RuleTester({
  // @ts-ignore
  languageOptions: {
    parserOptions: {
      project: './tsconfig.json',
      tsconfigRootDir: rootPath,
    },
  },
});

ruleTester.run('no-inferrable-types', {
  valid: ['const a = 10;', 'const b = true;'],
  invalid: [
    {
      code: 'const a: number = 10;',
      errors: [
        {
          column: 7,
          data: {
            type: 'number',
          },
          line: 1,
          messageId: 'noInferrableType',
        },
      ],
      output: 'const a = 10;',
    },
  ],
});

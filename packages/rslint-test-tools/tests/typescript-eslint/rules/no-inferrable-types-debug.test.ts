import { RuleTester, getFixturesRootDir } from '../RuleTester.ts';

const ruleTester = new RuleTester({
  languageOptions: {
    parserOptions: {
      project: './tsconfig.json',
      tsconfigRootDir: getFixturesRootDir(),
    },
  },
});

// Test just the first few valid cases to see where it hangs
ruleTester.run('no-inferrable-types', {
  valid: ['const a = 10n;'],
  invalid: [],
});

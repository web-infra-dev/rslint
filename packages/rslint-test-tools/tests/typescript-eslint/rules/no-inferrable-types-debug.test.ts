import { describe, test, expect } from '@rstest/core';
import { RuleTester } from '@typescript-eslint/rule-tester';
import { getFixturesRootDir } from '../RuleTester.ts';

const rootPath = getFixturesRootDir();

const ruleTester = new RuleTester({
  languageOptions: {
    parserOptions: {
      project: './tsconfig.json',
      tsconfigRootDir: rootPath,
    },
  },
});

describe('no-inferrable-types', () => {
  test('run rule', () => {
    // Test just the first few valid cases to see where it hangs
    ruleTester.run('no-inferrable-types', {
      valid: ['const a = 10n;'],
      invalid: [],
    });
  });
});

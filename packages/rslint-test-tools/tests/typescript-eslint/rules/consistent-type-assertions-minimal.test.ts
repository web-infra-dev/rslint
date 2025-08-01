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

// Type declarations for test cases
declare class Generic<T> {
  constructor();
}
declare type Foo = any;
declare type A = any;

ruleTester.run('consistent-type-assertions', {
  valid: [
    {
      code: 'const x = new Generic<number>() as Foo;',
      options: [
        {
          assertionStyle: 'as',
          objectLiteralTypeAssertions: 'allow',
        },
      ],
    },
    {
      code: 'const x = b as A;',
      options: [
        {
          assertionStyle: 'as',
          objectLiteralTypeAssertions: 'allow',
        },
      ],
    },
    // Test that type declarations are not flagged by this rule
    {
      code: 'type T = { x: number };',
      options: [
        {
          assertionStyle: 'as',
          objectLiteralTypeAssertions: 'allow',
        },
      ],
    },
  ],
  invalid: [
    {
      code: 'const x = new Generic<number>() as Foo;',
      errors: [
        {
          line: 1,
          messageId: 'angle-bracket',
        },
      ],
      options: [{ assertionStyle: 'angle-bracket' }],
    },
    {
      code: 'const x = b as A;',
      errors: [
        {
          line: 1,
          messageId: 'angle-bracket',
        },
      ],
      options: [{ assertionStyle: 'angle-bracket' }],
    },
  ],
});

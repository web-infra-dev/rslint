import { RuleTester, getFixturesRootDir } from '../RuleTester.ts';

const rootPath = getFixturesRootDir();

const ruleTester = new RuleTester({
  languageOptions: {
    parserOptions: {
      project: './tsconfig.json',
      tsconfigRootDir: rootPath,
    },
  },
});

ruleTester.run('no-unsafe-function-type', {
  valid: [
    'let value: () => void;',
    'let value: <T>(t: T) => T;',
    `
      // create a scope since it's illegal to declare a duplicate identifier
      // 'Function' in the global script scope.
      {
        type Function = () => void;
        let value: Function;
      }
    `,
  ],
  invalid: [
    {
      code: 'let value: Function;',
      errors: [
        {
          column: 12,
          line: 1,
          messageId: 'bannedFunctionType',
        },
      ],
    },
    {
      code: 'let value: Function[];',
      errors: [
        {
          column: 12,
          line: 1,
          messageId: 'bannedFunctionType',
        },
      ],
    },
    {
      code: 'let value: Function | number;',
      errors: [
        {
          column: 12,
          line: 1,
          messageId: 'bannedFunctionType',
        },
      ],
    },
    {
      code: `
        class Weird implements Function {
          // ...
        }
      `,
      errors: [
        {
          column: 32,
          line: 2,
          messageId: 'bannedFunctionType',
        },
      ],
    },
    {
      code: `
        interface Weird extends Function {
          // ...
        }
      `,
      errors: [
        {
          column: 33,
          line: 2,
          messageId: 'bannedFunctionType',
        },
      ],
    },
  ],
});
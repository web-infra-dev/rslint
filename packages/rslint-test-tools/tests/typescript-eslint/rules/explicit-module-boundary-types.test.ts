import { RuleTester, getFixturesRootDir } from '../RuleTester.ts';

const rootPath = getFixturesRootDir();

const ruleTester = new RuleTester({
  languageOptions: {
    parserOptions: {
      project: './tsconfig.json',
      projectService: false,
      tsconfigRootDir: rootPath,
    },
  },
});

ruleTester.run('explicit-module-boundary-types', {
  valid: [
    'function test(): void { return; }',
    'export function test(): void { return; }',
    'export var fn = function (): number { return 1; };',
    'export var arrowFn = (): string => "test";',
    'class Test { method(): void { return; } }',
  ],
  invalid: [
    {
      code: 'export function test() { return; }',
      errors: [{ messageId: 'missingReturnType' }],
    },
    {
      code: 'export var fn = function () { return 1; };',
      errors: [{ messageId: 'missingReturnType' }],
    },
  ],
});
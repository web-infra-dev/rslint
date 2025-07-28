import { lint } from '@rslint/core';
import path from 'node:path';

const tsconfig = path.resolve('./tests/typescript-eslint/fixtures/tsconfig.virtual.json');
const virtual_entry = path.resolve('./tests/typescript-eslint/fixtures/src/virtual.ts');

console.log('Testing no-empty-function with options...');

const result = await lint({
  tsconfig: path.basename(tsconfig),
  workingDirectory: path.dirname(tsconfig),
  fileContents: {
    [virtual_entry]: 'const foo = () => {};',
  },
  ruleOptions: {
    'no-empty-function': ['error', { allow: ['arrowFunctions'] }],
  },
});

console.log('Result:', JSON.stringify(result, null, 2));
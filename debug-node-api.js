import { lint } from '@rslint/core';
import path from 'path';

async function testRule() {
  const cwd = path.resolve('./packages/rslint-test-tools/tests/typescript-eslint/fixtures');
  const tsconfig = path.resolve(cwd, 'tsconfig.virtual.json');
  const virtual_entry = path.resolve(cwd, 'src/virtual.ts');
  
  console.log('Testing no-unnecessary-type-constraint...');
  console.log('Working directory:', path.dirname(tsconfig));
  console.log('TSConfig:', path.basename(tsconfig));
  console.log('Virtual entry:', virtual_entry);
  
  const result = await lint({
    tsconfig: path.basename(tsconfig),
    workingDirectory: path.dirname(tsconfig),
    fileContents: {
      [virtual_entry]: 'function data<T extends any>() {}'
    },
    ruleOptions: {
      'no-unnecessary-type-constraint': 'error'
    }
  });
  
  console.log('Result:', JSON.stringify(result, null, 2));
}

testRule().catch(console.error);
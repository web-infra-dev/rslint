import { lint } from '@rslint/core';
import path from 'path';

async function testArrowFunction() {
  console.log('Testing arrow function case...');
  
  const cwd = path.resolve('./packages/rslint-test-tools/tests/typescript-eslint/fixtures');
  const tsconfig = path.resolve(cwd, 'tsconfig.virtual.json');
  const virtual_entry = path.resolve(cwd, 'src/virtual.ts');
  
  const result = await lint({
    tsconfig: path.basename(tsconfig),
    workingDirectory: path.dirname(tsconfig),
    fileContents: {
      [virtual_entry]: 'const data = <T extends any>() => {};'
    },
    ruleOptions: {
      'no-unnecessary-type-constraint': 'error'
    }
  });
  
  console.log('Arrow function result:', JSON.stringify(result, null, 2));
  
  // Test regular function too
  const result2 = await lint({
    tsconfig: path.basename(tsconfig),
    workingDirectory: path.dirname(tsconfig),
    fileContents: {
      [virtual_entry]: 'function data<T extends any>() {}'
    },
    ruleOptions: {
      'no-unnecessary-type-constraint': 'error'
    }
  });
  
  console.log('Regular function result:', JSON.stringify(result2, null, 2));
}

testArrowFunction().catch(console.error);
import { lint } from '@rslint/core';
import path from 'path';

async function testLikeTestRunner() {
  console.log('Testing exactly like the test runner...');
  
  // Mimic exactly what the test runner does
  const cwd = path.resolve('./packages/rslint-test-tools/tests/typescript-eslint/fixtures');
  const tsconfig = path.resolve(cwd, 'tsconfig.virtual.json');
  const virtual_entry = path.resolve(cwd, 'src/virtual.ts');
  
  console.log('Working directory:', path.dirname(tsconfig));
  console.log('TSConfig basename:', path.basename(tsconfig));
  console.log('Virtual entry:', virtual_entry);
  
  const ruleName = 'no-unnecessary-type-constraint';
  const code = 'function data<T extends any>() {}';
  const ruleConfig = 'error';
  
  console.log('Rule name:', ruleName);
  console.log('Code:', code);
  console.log('Rule config:', ruleConfig);
  
  const result = await lint({
    tsconfig: path.basename(tsconfig),
    workingDirectory: path.dirname(tsconfig),
    fileContents: {
      [virtual_entry]: code,
    },
    ruleOptions: {
      [ruleName]: ruleConfig,
    },
  });
  
  console.log('Result:');
  console.log(JSON.stringify(result, null, 2));
}

testLikeTestRunner().catch(console.error);
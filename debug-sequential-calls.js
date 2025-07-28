import { lint } from '@rslint/core';
import path from 'path';

async function testSequentialCalls() {
  console.log('Testing multiple sequential calls like test runner...');
  
  const cwd = path.resolve('./packages/rslint-test-tools/tests/typescript-eslint/fixtures');
  const tsconfig = path.resolve(cwd, 'tsconfig.virtual.json');
  const virtual_entry = path.resolve(cwd, 'src/virtual.ts');
  
  const testCases = [
    'function data<T extends any>() {}',
    'function data<T extends any, U>() {}',
    'function data<T, U extends any>() {}',
    'const data = <T extends any>() => {};'
  ];
  
  console.log('Running tests sequentially...');
  
  for (let i = 0; i < testCases.length; i++) {
    const code = testCases[i];
    console.log(`\nTest ${i + 1}: ${code}`);
    
    try {
      const result = await Promise.race([
        lint({
          tsconfig: path.basename(tsconfig),
          workingDirectory: path.dirname(tsconfig),
          fileContents: {
            [virtual_entry]: code,
          },
          ruleOptions: {
            'no-unnecessary-type-constraint': 'error',
          },
        }),
        new Promise((_, reject) => 
          setTimeout(() => reject(new Error(`Timeout after 30s for test ${i + 1}`)), 30000)
        )
      ]);
      
      console.log(`Result ${i + 1}: ${result.diagnostics.length} diagnostics`);
      if (result.diagnostics.length > 0) {
        console.log(JSON.stringify(result.diagnostics[0], null, 2));
      } else {
        console.log('NO DIAGNOSTICS - This is the bug!');
      }
    } catch (error) {
      console.error(`Error in test ${i + 1}:`, error);
    }
  }
}

testSequentialCalls().catch(console.error);
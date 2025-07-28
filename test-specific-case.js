import { lint } from '@rslint/core';

const result = await lint({
  fileContents: {
    '/test.ts': `let test = 1; type TestType = typeof test; type Func = (test: string) => typeof test;`
  },
  ruleOptions: {
    'no-shadow': JSON.stringify(['error', { ignoreFunctionTypeParameterNameValueShadow: true }])
  }
});

console.log('Result:', JSON.stringify(result, null, 2));
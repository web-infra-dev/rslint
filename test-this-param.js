import { lint } from '@rslint/core';

const result = await lint({
  fileContents: {
    '/test.ts': `function foo(this: any, a: number) { function bar(this: any, a: string) {} }`
  },
  ruleOptions: {
    'no-shadow': JSON.stringify(['error'])
  }
});

console.log('Result:', JSON.stringify(result, null, 2));
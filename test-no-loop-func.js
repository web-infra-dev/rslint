import { lint } from '@rslint/core';

const result = await lint({
  fileContents: {
    '/virtual.ts': `
for (let i = 0; i < 10; i++) {
  function foo() {
    console.log('A');
  }
}
    `
  },
  ruleOptions: {
    'no-loop-func': 'error'
  }
});

console.log('Result:', JSON.stringify(result, null, 2));
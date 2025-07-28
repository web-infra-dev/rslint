import { lint } from '@rslint/core';

const result = await lint({
  fileContents: {
    '/virtual.ts': `
for (let i = 0; i < l; i++) {
  (function () {
    i;
  });
}
    `
  },
  ruleOptions: {
    'no-loop-func': 'error'
  }
});

console.log('Let loop result:', JSON.stringify(result, null, 2));
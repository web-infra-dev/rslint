import { lint } from '@rslint/core';

const result = await lint({
  fileContents: {
    '/virtual.ts': `
for (var i = 0; i < l; i++) {
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

console.log('Simple var result:', JSON.stringify(result, null, 2));
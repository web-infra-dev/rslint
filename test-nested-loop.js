import { lint } from '@rslint/core';

const result = await lint({
  fileContents: {
    '/virtual.ts': `
for (var i = 0; i < l; i++) {
  for (var j = 0; j < m; j++) {
    (function () {
      i + j;
    });
  }
}
    `
  },
  ruleOptions: {
    'no-loop-func': 'error'
  }
});

console.log('Nested loop result:', JSON.stringify(result, null, 2));
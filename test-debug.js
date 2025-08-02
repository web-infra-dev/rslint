const { lint } = require('./packages/rslint/dist/index.js');

async function test() {
  const result = await lint({
    workingDirectory:
      '/Users/bytedance/dev/rslint/packages/rslint-test-tools/tests/typescript-eslint/fixtures',
    config:
      '/Users/bytedance/dev/rslint/packages/rslint-test-tools/tests/typescript-eslint/fixtures/rslint.json',
    files: [
      '/Users/bytedance/dev/rslint/packages/rslint-test-tools/tests/typescript-eslint/fixtures/src/virtual.ts',
    ],
    fileContents: {
      '/Users/bytedance/dev/rslint/packages/rslint-test-tools/tests/typescript-eslint/fixtures/src/virtual.ts': `
function foo(s: string);
function foo(n: number);
function bar(): void {}
function baz(): void {}
function foo(sn: string | number) {}
      `,
    },
    ruleOptions: {
      'adjacent-overload-signatures': [],
    },
  });

  console.log('Result:', JSON.stringify(result, null, 2));
}

test().catch(console.error);

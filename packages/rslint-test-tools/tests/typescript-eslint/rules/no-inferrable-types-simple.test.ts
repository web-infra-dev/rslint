import { lint } from '@rslint/core';
import test from 'node:test';

test('no-inferrable-types simple', async () => {
  console.log('Starting simple test...');

  try {
    const result = await Promise.race([
      lint({
        config:
          '/Users/bytedance/dev/rslint/packages/rslint-test-tools/tests/typescript-eslint/fixtures/rslint.json',
        workingDirectory:
          '/Users/bytedance/dev/rslint/packages/rslint-test-tools/tests/typescript-eslint/fixtures',
        fileContents: {
          'src/virtual.ts': 'const a = 10;',
        },
        ruleOptions: {
          'no-inferrable-types': 'error',
        },
      }),
      new Promise((_, reject) =>
        setTimeout(() => reject(new Error('Timeout after 5s')), 5000),
      ),
    ]);

    console.log('Result:', JSON.stringify(result, null, 2));
  } catch (error) {
    console.error(
      'Error:',
      error instanceof Error ? error.message : String(error),
    );
    throw error;
  }
});

import { describe, test, expect } from '@rstest/core';
import { lint } from '@rslint/core';

describe('no-inferrable-types simple', () => {
  test('should work', async () => {
    console.log('Starting simple test...');

    try {
      const result = await Promise.race([
        lint({
          config:
            '/Users/bytedance/dev/rslint/packages/rslint-test-tools/tests/typescript-eslint/fixtures/rslint.json',
          workingDirectory:
            '/Users/bytedance/dev/rslint/packages/rslint-test-tools/tests/typescript-eslint/fixtures',
          files: [
            '/Users/bytedance/dev/rslint/packages/rslint-test-tools/tests/typescript-eslint/fixtures/src/virtual.ts',
          ],
          fileContents: {
            '/Users/bytedance/dev/rslint/packages/rslint-test-tools/tests/typescript-eslint/fixtures/src/virtual.ts':
              'const a: number = 10;',
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
});

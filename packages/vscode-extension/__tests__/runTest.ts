import * as path from 'path';

import { runTests } from '@vscode/test-electron';

async function main() {
  try {
    const extensionDevelopmentPath = path.resolve(__dirname, '..');
    const testWorkspace = path.resolve(
      require.resolve('@rslint/core'),
      '../..',
      'fixtures',
    );

    const extensionTestsPath = path.resolve(__dirname, './suite');

    await runTests({
      extensionDevelopmentPath,
      extensionTestsPath,
      launchArgs: ['--disable-extensions', testWorkspace],
      version: 'stable',
    });

    await runTests({
      extensionDevelopmentPath,
      extensionTestsPath,
      launchArgs: ['--disable-extensions', testWorkspace],
      version: '1.106.3',
    });
  } catch (err) {
    console.error(err);
    console.error('Failed to run tests');
    process.exit(1);
  }
}

void main();

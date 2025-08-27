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
      launchArgs: [
        '--disable-extensions',
        testWorkspace,
        '--disable-features=NetworkService,OutOfBlinkCors',
        '--disable-integrated-auth',
        '--auth-server-allowlist="_"',
      ],
    });
  } catch (err) {
    console.error(err);
    console.error('Failed to run tests');
    process.exit(1);
  }
}

main();

import * as path from 'path';

import { runTests } from '@vscode/test-electron';

async function main() {
  try {
    const extensionDevelopmentPath = path.resolve(__dirname, '..');
    const testWorkspace = path.resolve(process.cwd(), '__tests__', 'fixtures');

    const extensionTestsPath = path.resolve(__dirname, './suite');

    await runTests({
      extensionDevelopmentPath,
      extensionTestsPath,
      launchArgs: ['--disable-extensions', testWorkspace]
    });
  } catch (err) {
    console.error(err);
    console.error('Failed to run tests');
    process.exit(1);
  }
}

main();

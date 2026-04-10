import * as path from 'path';

import { runTests } from '@vscode/test-electron';

function resolveFixture(name: string): string {
  return path.resolve(require.resolve('@rslint/core'), '../..', name);
}

async function main() {
  const extensionDevelopmentPath = path.resolve(__dirname, '..');
  let failed = false;

  // --- Existing tests (JSON config workspace) ---
  const testWorkspace = resolveFixture('fixtures');
  const extensionTestsPath = path.resolve(__dirname, './suite');

  try {
    await runTests({
      extensionDevelopmentPath,
      extensionTestsPath,
      launchArgs: ['--disable-extensions', testWorkspace],
      version: 'stable',
    });
  } catch (err) {
    console.error('JSON config tests (stable) failed:', err);
    failed = true;
  }

  try {
    await runTests({
      extensionDevelopmentPath,
      extensionTestsPath,
      launchArgs: ['--disable-extensions', testWorkspace],
      version: '1.106.3',
    });
  } catch (err) {
    console.error('JSON config tests (1.106.3) failed:', err);
    failed = true;
  }

  // --- JS config tests ---
  const testsSourceDir = path.resolve(extensionDevelopmentPath, '__tests__');
  const jsConfigWorkspace = path.resolve(testsSourceDir, 'fixtures-jsconfig');
  const jsConfigTestsPath = path.resolve(__dirname, './suite-jsconfig');

  try {
    await runTests({
      extensionDevelopmentPath,
      extensionTestsPath: jsConfigTestsPath,
      launchArgs: ['--disable-extensions', jsConfigWorkspace],
      version: 'stable',
    });
  } catch (err) {
    console.error('JS config tests failed:', err);
    failed = true;
  }

  // --- Monorepo multi-config tests ---
  const monorepoWorkspace = path.resolve(testsSourceDir, 'fixtures-monorepo');
  const monorepoTestsPath = path.resolve(__dirname, './suite-monorepo');

  try {
    await runTests({
      extensionDevelopmentPath,
      extensionTestsPath: monorepoTestsPath,
      launchArgs: ['--disable-extensions', monorepoWorkspace],
      version: 'stable',
    });
  } catch (err) {
    console.error('Monorepo config tests failed:', err);
    failed = true;
  }

  // --- No config fallback tests ---
  const noConfigWorkspace = path.resolve(testsSourceDir, 'fixtures-noconfig');
  const noConfigTestsPath = path.resolve(__dirname, './suite-noconfig');

  try {
    await runTests({
      extensionDevelopmentPath,
      extensionTestsPath: noConfigTestsPath,
      launchArgs: ['--disable-extensions', noConfigWorkspace],
      version: 'stable',
    });
  } catch (err) {
    console.error('No config tests failed:', err);
    failed = true;
  }

  // --- Type-aware rule scope tests ---
  const typeAwareScopeWorkspace = path.resolve(
    testsSourceDir,
    'fixtures-type-aware-scope',
  );
  const typeAwareScopeTestsPath = path.resolve(
    __dirname,
    './suite-type-aware-scope',
  );

  try {
    await runTests({
      extensionDevelopmentPath,
      extensionTestsPath: typeAwareScopeTestsPath,
      launchArgs: ['--disable-extensions', typeAwareScopeWorkspace],
      version: 'stable',
    });
  } catch (err) {
    console.error('Type-aware scope tests failed:', err);
    failed = true;
  }

  if (failed) {
    console.error('Some test suites failed');
    process.exit(1);
  }
}

main();

import * as path from 'path';

import { runTests } from '@vscode/test-electron';

function resolveFixture(name: string): string {
  return path.resolve(require.resolve('@rslint/core'), '../..', name);
}

async function main() {
  const extensionDevelopmentPath = path.resolve(__dirname, '..');
  let failed = false;

  const launchArgs = ['--no-sandbox', '--disable-gpu', '--disable-extensions'];

  // --- Existing tests (JSON config workspace) ---
  const testWorkspace = resolveFixture('fixtures');
  const extensionTestsPath = path.resolve(__dirname, './suite');

  try {
    await runTests({
      extensionDevelopmentPath,
      extensionTestsPath,
      launchArgs: [...launchArgs, testWorkspace],
      version: 'stable',
    });
  } catch (err) {
    console.error('JSON config tests (stable) failed:', err);
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
      launchArgs: [...launchArgs, jsConfigWorkspace],
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
      launchArgs: [...launchArgs, monorepoWorkspace],
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
      launchArgs: [...launchArgs, noConfigWorkspace],
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
      launchArgs: [...launchArgs, typeAwareScopeWorkspace],
      version: 'stable',
    });
  } catch (err) {
    console.error('Type-aware scope tests failed:', err);
    failed = true;
  }

  // --- projectService type-aware scope tests ---
  const projectServiceScopeWorkspace = path.resolve(
    testsSourceDir,
    'fixtures-project-service-scope',
  );
  const projectServiceScopeTestsPath = path.resolve(
    __dirname,
    './suite-project-service-scope',
  );

  try {
    await runTests({
      extensionDevelopmentPath,
      extensionTestsPath: projectServiceScopeTestsPath,
      launchArgs: [...launchArgs, projectServiceScopeWorkspace],
      version: 'stable',
    });
  } catch (err) {
    console.error('projectService scope tests failed:', err);
    failed = true;
  }

  // --- eslintPlugins reverse-dispatch tests ---
  const eslintPluginsWorkspace = path.resolve(
    testsSourceDir,
    'fixtures-eslint-plugins',
  );
  const eslintPluginsTestsPath = path.resolve(
    __dirname,
    './suite-eslint-plugins',
  );

  try {
    await runTests({
      extensionDevelopmentPath,
      extensionTestsPath: eslintPluginsTestsPath,
      launchArgs: [...launchArgs, eslintPluginsWorkspace],
      version: 'stable',
    });
  } catch (err) {
    console.error('eslintPlugins tests failed:', err);
    failed = true;
  }

  if (failed) {
    console.error('Some test suites failed');
    process.exit(1);
  }
}

main();

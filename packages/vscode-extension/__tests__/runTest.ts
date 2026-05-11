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

  // try {
  //   await runTests({
  //     extensionDevelopmentPath,
  //     extensionTestsPath,
  //     launchArgs: ['--disable-extensions', testWorkspace],
  //     version: '1.106.3',
  //   });
  // } catch (err) {
  //   console.error('JSON config tests (1.106.3) failed:', err);
  //   failed = true;
  // }

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

  // --- Single-config ESLint plugin smoke test ---
  // Verifies the full chain from a user config that declares
  // `eslintPlugins` down to plugin diagnostics arriving in
  // vscode.languages.getDiagnostics. Uses a self-contained fake
  // plugin under the workspace so the fixture has no npm deps.
  const eslintPluginWorkspace = path.resolve(
    testsSourceDir,
    'fixtures-eslint-plugin',
  );
  const eslintPluginTestsPath = path.resolve(
    __dirname,
    './suite-eslint-plugin',
  );

  try {
    await runTests({
      extensionDevelopmentPath,
      extensionTestsPath: eslintPluginTestsPath,
      launchArgs: ['--disable-extensions', eslintPluginWorkspace],
      version: 'stable',
    });
  } catch (err) {
    console.error('ESLint plugin (single-config) tests failed:', err);
    failed = true;
  }

  // --- Multi-config ESLint plugin routing test ---
  // Two nested packages, each with a disjoint plugin (different
  // prefix, different module). Verifies the worker's per-config
  // `loadedPluginsByDir` map keeps the two configs isolated end-to-
  // end through the LSP layer.
  const eslintPluginMonorepoWorkspace = path.resolve(
    testsSourceDir,
    'fixtures-eslint-plugin-monorepo',
  );
  const eslintPluginMonorepoTestsPath = path.resolve(
    __dirname,
    './suite-eslint-plugin-monorepo',
  );

  try {
    await runTests({
      extensionDevelopmentPath,
      extensionTestsPath: eslintPluginMonorepoTestsPath,
      launchArgs: ['--disable-extensions', eslintPluginMonorepoWorkspace],
      version: 'stable',
    });
  } catch (err) {
    console.error('ESLint plugin (multi-config) tests failed:', err);
    failed = true;
  }

  // --- Mixed native+plugin & gap-file+plugin tests ---
  // The config enables a native rule (no-debugger) AND a plugin rule
  // (fx/no-forbidden) together; tsconfig include is src/ only so
  // scripts/gap.ts is a gap file. Covers the two LSP-side combinations
  // the pure-plugin suites miss: native+plugin on the same file, and a
  // gap file still receiving plugin diagnostics.
  const eslintPluginMixedWorkspace = path.resolve(
    testsSourceDir,
    'fixtures-eslint-plugin-mixed',
  );
  const eslintPluginMixedTestsPath = path.resolve(
    __dirname,
    './suite-eslint-plugin-mixed',
  );

  try {
    await runTests({
      extensionDevelopmentPath,
      extensionTestsPath: eslintPluginMixedTestsPath,
      launchArgs: ['--disable-extensions', eslintPluginMixedWorkspace],
      version: 'stable',
    });
  } catch (err) {
    console.error(
      'ESLint plugin (mixed native+plugin / gap) tests failed:',
      err,
    );
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
      launchArgs: ['--disable-extensions', projectServiceScopeWorkspace],
      version: 'stable',
    });
  } catch (err) {
    console.error('projectService scope tests failed:', err);
    failed = true;
  }

  if (failed) {
    console.error('Some test suites failed');
    process.exit(1);
  }
}

main();

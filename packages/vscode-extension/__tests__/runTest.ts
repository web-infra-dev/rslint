import * as path from 'path';
import * as fs from 'node:fs';
import * as os from 'node:os';

import { runTests } from '@vscode/test-electron';

function resolveFixture(name: string): string {
  return path.resolve(require.resolve('@rslint/core'), '../..', name);
}

function copyToIsolatedWorkspace(source: string): {
  workspace: string;
  userDataDir: string;
  extensionsDir: string;
  dispose(): void;
} {
  // Go discovery intentionally searches strict cwd ancestors. Keeping an E2E
  // fixture under this repository would therefore inherit the repository's
  // root config (whose global ignores exclude __tests__) and correctly prune
  // the fixture config. A physical temp copy gives the fixture the same clean
  // ancestry as a real standalone user workspace.
  // VS Code and the language client create Unix sockets below these paths.
  // macOS's os.tmpdir() is already long enough to exceed the 103-byte socket
  // limit once VS Code appends its own names, so keep every test-owned path
  // directly below a short /tmp container. Windows has no Unix socket limit.
  const tempRoot = process.platform === 'win32' ? os.tmpdir() : '/tmp';
  const container = fs.mkdtempSync(path.join(tempRoot, 'rsv-'));
  const workspace = path.join(container, 'w');
  const userDataDir = path.join(container, 'u');
  const extensionsDir = path.join(container, 'e');
  try {
    fs.cpSync(source, workspace, { recursive: true });
    fs.mkdirSync(userDataDir);
    fs.mkdirSync(extensionsDir);
  } catch (error) {
    fs.rmSync(container, { recursive: true, force: true });
    throw error;
  }
  return {
    workspace,
    userDataDir,
    extensionsDir,
    dispose() {
      fs.rmSync(container, { recursive: true, force: true });
    },
  };
}

async function runIsolatedFixtureTests(options: {
  source: string;
  name: string;
  extensionDevelopmentPath: string;
  extensionTestsPath: string;
}): Promise<boolean> {
  let isolatedWorkspace: ReturnType<typeof copyToIsolatedWorkspace> | undefined;
  try {
    isolatedWorkspace = copyToIsolatedWorkspace(options.source);
    await runTests({
      extensionDevelopmentPath: options.extensionDevelopmentPath,
      extensionTestsPath: options.extensionTestsPath,
      launchArgs: [
        '--disable-extensions',
        `--user-data-dir=${isolatedWorkspace.userDataDir}`,
        `--extensions-dir=${isolatedWorkspace.extensionsDir}`,
        isolatedWorkspace.workspace,
      ],
      version: 'stable',
    });
    return true;
  } catch (err) {
    console.error(`${options.name} failed:`, err);
    return false;
  } finally {
    isolatedWorkspace?.dispose();
  }
}

async function main() {
  const extensionDevelopmentPath = path.resolve(__dirname, '../..');
  let failed = false;

  // --- Existing tests (JSON config workspace) ---
  const extensionTestsPath = path.resolve(__dirname, './suite');
  if (
    !(await runIsolatedFixtureTests({
      source: resolveFixture('fixtures'),
      name: 'JSON config tests (stable)',
      extensionDevelopmentPath,
      extensionTestsPath,
    }))
  ) {
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
  const jsConfigTestsPath = path.resolve(__dirname, './suite-jsconfig');
  if (
    !(await runIsolatedFixtureTests({
      source: path.resolve(testsSourceDir, 'fixtures-jsconfig'),
      name: 'JS config tests',
      extensionDevelopmentPath,
      extensionTestsPath: jsConfigTestsPath,
    }))
  ) {
    failed = true;
  }

  // --- Monorepo multi-config tests ---
  const monorepoTestsPath = path.resolve(__dirname, './suite-monorepo');
  if (
    !(await runIsolatedFixtureTests({
      source: path.resolve(testsSourceDir, 'fixtures-monorepo'),
      name: 'Monorepo config tests',
      extensionDevelopmentPath,
      extensionTestsPath: monorepoTestsPath,
    }))
  ) {
    failed = true;
  }

  // --- No config fallback tests ---
  const noConfigTestsPath = path.resolve(__dirname, './suite-noconfig');
  if (
    !(await runIsolatedFixtureTests({
      source: path.resolve(testsSourceDir, 'fixtures-noconfig'),
      name: 'No config tests',
      extensionDevelopmentPath,
      extensionTestsPath: noConfigTestsPath,
    }))
  ) {
    failed = true;
  }

  // --- Type-aware rule scope tests ---
  const typeAwareScopeTestsPath = path.resolve(
    __dirname,
    './suite-type-aware-scope',
  );
  if (
    !(await runIsolatedFixtureTests({
      source: path.resolve(testsSourceDir, 'fixtures-type-aware-scope'),
      name: 'Type-aware scope tests',
      extensionDevelopmentPath,
      extensionTestsPath: typeAwareScopeTestsPath,
    }))
  ) {
    failed = true;
  }

  // --- projectService type-aware scope tests ---
  const projectServiceScopeTestsPath = path.resolve(
    __dirname,
    './suite-project-service-scope',
  );
  if (
    !(await runIsolatedFixtureTests({
      source: path.resolve(testsSourceDir, 'fixtures-project-service-scope'),
      name: 'projectService scope tests',
      extensionDevelopmentPath,
      extensionTestsPath: projectServiceScopeTestsPath,
    }))
  ) {
    failed = true;
  }

  // --- eslintPlugins reverse-dispatch tests ---
  const eslintPluginsTestsPath = path.resolve(
    __dirname,
    './suite-eslint-plugins',
  );
  if (
    !(await runIsolatedFixtureTests({
      source: path.resolve(testsSourceDir, 'fixtures-eslint-plugins'),
      name: 'eslintPlugins tests',
      extensionDevelopmentPath,
      extensionTestsPath: eslintPluginsTestsPath,
    }))
  ) {
    failed = true;
  }

  if (failed) {
    console.error('Some test suites failed');
    process.exit(1);
  }
}

main();

import { randomUUID } from 'node:crypto';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';

import { downloadAndUnzipVSCode, runTests } from '@vscode/test-electron';

interface TestSuite {
  name: string;
  workspace: string;
  tests: string;
  workspaceEntry?: string;
  workspaceFolders?: string[];
}

const workspaceMarkerFile = '.rslint-vscode-test-sandbox.json';

function resolveFixture(name: string): string {
  return path.resolve(require.resolve('@rslint/core'), '../..', name);
}

function resolveSandboxEntry(workspaceRoot: string, entry: string): string {
  const resolved = path.resolve(workspaceRoot, entry);
  const relative = path.relative(workspaceRoot, resolved);
  if (
    relative === '..' ||
    relative.startsWith(`..${path.sep}`) ||
    path.isAbsolute(relative)
  ) {
    throw new Error(`Workspace entry escapes its isolated sandbox: ${entry}`);
  }
  return resolved;
}

async function findPackageRoot(startPath: string): Promise<string> {
  let current = path.resolve(startPath);
  while (true) {
    try {
      await fs.promises.access(path.join(current, 'package.json'));
      return current;
    } catch {
      const parent = path.dirname(current);
      if (parent === current) {
        throw new Error(`Could not find a package boundary above ${startPath}`);
      }
      current = parent;
    }
  }
}

async function runIsolatedSuite(
  extensionDevelopmentPath: string,
  vscodeExecutablePath: string,
  suite: TestSuite,
): Promise<void> {
  // Go discovery intentionally searches strict cwd ancestors, so keep each
  // fixture in a clean physical workspace outside this repository. Use short
  // paths on Unix as VS Code and the language client append socket names that
  // can otherwise exceed macOS's Unix-domain socket path limit.
  const tempRoot = process.platform === 'win32' ? os.tmpdir() : '/tmp';
  const profileRoot = await fs.promises.mkdtemp(path.join(tempRoot, 'rsv-'));
  const userDataDir = path.join(profileRoot, 'u');
  const extensionsDir = path.join(profileRoot, 'e');
  const workspaceCopy = path.join(profileRoot, 'w');

  let testError: unknown;
  try {
    // Suites intentionally create, rewrite, and delete config files. Run them
    // against a private copy so even an Extension Host crash cannot mutate a
    // tracked fixture in the checkout.
    await fs.promises.cp(suite.workspace, workspaceCopy, {
      recursive: true,
      force: false,
      errorOnExist: true,
    });
    const expectedWorkspaceFolders = await Promise.all(
      (suite.workspaceFolders ?? ['.']).map((folder) =>
        fs.promises.realpath(resolveSandboxEntry(workspaceCopy, folder)),
      ),
    );
    await fs.promises.writeFile(
      path.join(workspaceCopy, workspaceMarkerFile),
      JSON.stringify({
        version: 2,
        nonce: randomUUID(),
        sourceWorkspace: await fs.promises.realpath(suite.workspace),
        expectedWorkspace: await fs.promises.realpath(workspaceCopy),
        expectedWorkspaceFolders,
      }),
      { encoding: 'utf8', flag: 'wx', mode: 0o600 },
    );

    // Preserve the source workspace's package boundary and dependency lookup
    // without placing a writable node_modules link inside the test workspace.
    const packageRoot = await findPackageRoot(suite.workspace);
    await fs.promises.copyFile(
      path.join(packageRoot, 'package.json'),
      path.join(profileRoot, 'package.json'),
    );
    await fs.promises.symlink(
      path.join(packageRoot, 'node_modules'),
      path.join(profileRoot, 'node_modules'),
      process.platform === 'win32' ? 'junction' : 'dir',
    );

    await runTests({
      extensionDevelopmentPath,
      extensionTestsPath: suite.tests,
      vscodeExecutablePath,
      launchArgs: [
        suite.workspaceEntry
          ? resolveSandboxEntry(workspaceCopy, suite.workspaceEntry)
          : workspaceCopy,
        '--disable-extensions',
        '--disable-updates',
        '--disable-workspace-trust',
        '--force-disable-user-env',
        '--skip-release-notes',
        '--skip-welcome',
        `--user-data-dir=${userDataDir}`,
        `--extensions-dir=${extensionsDir}`,
      ],
    });
  } catch (error) {
    testError = error;
  }

  let cleanupError: unknown;
  try {
    await fs.promises.rm(profileRoot, {
      recursive: true,
      force: true,
      maxRetries: 10,
      retryDelay: 200,
    });
  } catch (error) {
    cleanupError = error;
  }

  if (testError && cleanupError) {
    throw new AggregateError(
      [testError, cleanupError],
      `${suite.name} failed and its isolated sandbox could not be removed`,
    );
  }
  if (testError) throw testError;
  if (cleanupError) throw cleanupError;
}

async function main(): Promise<void> {
  const extensionDevelopmentPath = path.resolve(__dirname, '../..');
  // Resolve the compatible stable release once so every suite explicitly uses
  // the same executable and repeated runTests() calls avoid redundant
  // compatibility/cache resolution.
  const vscodeExecutablePath = await downloadAndUnzipVSCode({
    extensionDevelopmentPath,
  });
  const testsSourceDir = path.resolve(extensionDevelopmentPath, '__tests__');
  const suites: TestSuite[] = [
    {
      name: 'JSON config tests',
      workspace: resolveFixture('fixtures'),
      tests: path.resolve(__dirname, './suite'),
    },
    {
      name: 'JS config tests',
      workspace: path.resolve(testsSourceDir, 'fixtures-jsconfig'),
      tests: path.resolve(__dirname, './suite-jsconfig'),
    },
    {
      name: 'Monorepo config tests',
      workspace: path.resolve(testsSourceDir, 'fixtures-monorepo'),
      tests: path.resolve(__dirname, './suite-monorepo'),
    },
    {
      name: 'Multi-root dynamic ownership tests',
      workspace: path.resolve(testsSourceDir, 'fixtures-multiroot'),
      tests: path.resolve(__dirname, './suite-multiroot'),
      workspaceEntry: 'multiroot.code-workspace',
      workspaceFolders: [
        'parent',
        'sentinel',
        'twins/left/app',
        'twins/right/app',
      ],
    },
    {
      name: 'Multi-root initial parent-child tests',
      workspace: path.resolve(testsSourceDir, 'fixtures-multiroot'),
      tests: path.resolve(__dirname, './suite-multiroot'),
      workspaceEntry: 'nested-initial.code-workspace',
      workspaceFolders: ['parent', 'parent/nested', 'sentinel'],
    },
    {
      name: 'No config tests',
      workspace: path.resolve(testsSourceDir, 'fixtures-noconfig'),
      tests: path.resolve(__dirname, './suite-noconfig'),
    },
    {
      name: 'Type-aware scope tests',
      workspace: path.resolve(testsSourceDir, 'fixtures-type-aware-scope'),
      tests: path.resolve(__dirname, './suite-type-aware-scope'),
    },
    {
      name: 'projectService scope tests',
      workspace: path.resolve(testsSourceDir, 'fixtures-project-service-scope'),
      tests: path.resolve(__dirname, './suite-project-service-scope'),
    },
    {
      name: 'eslintPlugins tests',
      workspace: path.resolve(testsSourceDir, 'fixtures-eslint-plugins'),
      tests: path.resolve(__dirname, './suite-eslint-plugins'),
    },
    {
      name: 'Generated rule-option-types tests',
      workspace: path.resolve(testsSourceDir, 'fixtures-rule-option-types'),
      tests: path.resolve(__dirname, './suite-rule-option-types'),
    },
  ];

  const failures: unknown[] = [];
  for (const suite of suites) {
    try {
      await runIsolatedSuite(
        extensionDevelopmentPath,
        vscodeExecutablePath,
        suite,
      );
    } catch (error) {
      console.error(`${suite.name} failed:`, error);
      failures.push(error);
    }
  }

  if (failures.length > 0) {
    throw new AggregateError(
      failures,
      `${failures.length} VS Code suite(s) failed`,
    );
  }
}

void main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});

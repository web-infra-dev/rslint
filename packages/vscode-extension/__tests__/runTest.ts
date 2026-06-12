import * as path from 'path';
import * as fs from 'fs';

import { runTests, downloadAndUnzipVSCode } from '@vscode/test-electron';

function resolveFixture(name: string): string {
  return path.resolve(require.resolve('@rslint/core'), '../..', name);
}

interface TestSuite {
  name: string;
  workspace: string;
  testsPath: string;
}

async function main() {
  const extensionDevelopmentPath = path.resolve(__dirname, '..');
  const testsSourceDir = path.resolve(extensionDevelopmentPath, '__tests__');

  const suites: TestSuite[] = [
    {
      name: 'json-config',
      workspace: resolveFixture('fixtures'),
      testsPath: path.resolve(__dirname, './suite'),
    },
    {
      name: 'js-config',
      workspace: path.resolve(testsSourceDir, 'fixtures-jsconfig'),
      testsPath: path.resolve(__dirname, './suite-jsconfig'),
    },
    {
      name: 'monorepo',
      workspace: path.resolve(testsSourceDir, 'fixtures-monorepo'),
      testsPath: path.resolve(__dirname, './suite-monorepo'),
    },
    {
      name: 'noconfig',
      workspace: path.resolve(testsSourceDir, 'fixtures-noconfig'),
      testsPath: path.resolve(__dirname, './suite-noconfig'),
    },
    {
      name: 'type-aware-scope',
      workspace: path.resolve(testsSourceDir, 'fixtures-type-aware-scope'),
      testsPath: path.resolve(__dirname, './suite-type-aware-scope'),
    },
    {
      name: 'project-service-scope',
      workspace: path.resolve(testsSourceDir, 'fixtures-project-service-scope'),
      testsPath: path.resolve(__dirname, './suite-project-service-scope'),
    },
    {
      name: 'eslint-plugins',
      workspace: path.resolve(testsSourceDir, 'fixtures-eslint-plugins'),
      testsPath: path.resolve(__dirname, './suite-eslint-plugins'),
    },
  ];

  console.log(`Starting ${suites.length} test suites in parallel...`);

  let failed = false;

  let vscodeExecutablePath: string | undefined;
  try {
    console.log('Downloading and unzipping VS Code...');
    vscodeExecutablePath = await downloadAndUnzipVSCode({ version: 'stable' });
    console.log(`VS Code executable path resolved: ${vscodeExecutablePath}`);
  } catch (err) {
    console.error('Failed to download VS Code:', err);
    process.exit(1);
  }

  await Promise.allSettled(
    suites.map(async (suite, index) => {
      // Create isolated user-data and extensions directories for each suite to prevent conflicts
      // Avoid UNIX socket path length limit (108 chars) by keeping paths short.
      // Also avoid os.tmpdir() (/tmp) on CI containers where it's mapped to a small tmpfs.
      // We place the directories under the workspace root where there is plenty of disk space.
      const repoRoot = path.resolve(extensionDevelopmentPath, '../..');
      const tmpBase = path.join(repoRoot, `.tmp-test-dir-${index}`);
      const userDataDir = path.join(tmpBase, 'user-data');
      const extensionsDir = path.join(tmpBase, 'extensions');

      try {
        await runTests({
          vscodeExecutablePath,
          extensionDevelopmentPath,
          extensionTestsPath: suite.testsPath,
          launchArgs: [
            '--disable-extensions',
            `--user-data-dir=${userDataDir}`,
            `--extensions-dir=${extensionsDir}`,
            suite.workspace,
          ],
          version: 'stable',
        });
        console.log(`Suite "${suite.name}" passed.`);
      } catch (err) {
        console.error(`Suite "${suite.name}" failed:`, err);
        failed = true;
      } finally {
        // Clean up temp directories
        try {
          fs.rmSync(tmpBase, { recursive: true, force: true });
        } catch (cleanupErr) {
          // Ignore cleanup errors
        }
      }
    }),
  );

  if (failed) {
    console.error('Some test suites failed');
    process.exit(1);
  } else {
    console.log('All test suites passed successfully!');
  }
}

main();

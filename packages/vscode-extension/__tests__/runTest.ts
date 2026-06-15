import * as path from 'path';
import * as fs from 'fs';
import * as os from 'os';

import { runTests, downloadAndUnzipVSCode } from '@vscode/test-electron';

function resolveFixture(name: string): string {
  return path.resolve(require.resolve('@rslint/core'), '../..', name);
}

interface TestSuite {
  name: string;
  workspace: string;
  testsPath: string;
}

async function runWithLimit<T>(
  tasks: (() => Promise<T>)[],
  limit: number,
): Promise<PromiseSettledResult<T>[]> {
  const results: PromiseSettledResult<T>[] = [];
  const executing = new Set<Promise<void>>();

  for (let i = 0; i < tasks.length; i++) {
    const taskIndex = i;
    const p = (async () => {
      try {
        const val = await tasks[taskIndex]();
        results[taskIndex] = { status: 'fulfilled', value: val };
      } catch (err) {
        results[taskIndex] = { status: 'rejected', reason: err };
      }
    })();

    executing.add(p);
    const clean = () => executing.delete(p);
    p.then(clean, clean);

    if (executing.size >= limit) {
      await Promise.race(executing);
    }
  }
  await Promise.all(executing);
  return results;
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

  console.log(`Starting ${suites.length} test suites...`);

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

  const limit = os.platform() === 'win32' ? 1 : 4;
  console.log(`Running test suites with concurrency limit of ${limit}...`);

  const tasks = suites.map((suite, index) => async () => {
    // Create isolated user-data and extensions directories for each suite to prevent conflicts
    // Avoid UNIX socket path length limit (108 chars) by keeping paths short.
    // On Linux, use workspace root (repoRoot) to avoid /tmp space limits.
    // On Windows, use os.tmpdir() to avoid STATUS_ACCESS_VIOLATION crashes on container volume mounts.
    const tmpBase =
      process.platform === 'win32'
        ? path.join(os.tmpdir(), `rslint-test-${index}`)
        : path.join(
            path.resolve(extensionDevelopmentPath, '../..'),
            `.tmp-test-dir-${index}`,
          );
    const userDataDir = path.join(tmpBase, 'user-data');
    const extensionsDir = path.join(tmpBase, 'extensions');

    const extraArgs: string[] = [];
    if (process.platform === 'linux') {
      extraArgs.push(
        '--disable-gpu',
        '--disable-updates',
        '--no-sandbox',
        '--disable-dev-shm-usage',
      );
    }

    try {
      await runTests({
        vscodeExecutablePath,
        extensionDevelopmentPath,
        extensionTestsPath: suite.testsPath,
        launchArgs: [
          '--disable-extensions',
          ...extraArgs,
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
  });

  await runWithLimit(tasks, limit);

  if (failed) {
    console.error('Some test suites failed');
    process.exit(1);
  } else {
    console.log('All test suites passed successfully!');
  }
}

main();

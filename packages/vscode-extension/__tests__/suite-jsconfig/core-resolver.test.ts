import * as assert from 'node:assert';
import fs from 'node:fs/promises';
import os from 'node:os';
import path from 'node:path';
import {
  Uri,
  workspace,
  type TextDocument,
  type WorkspaceFolder,
} from 'vscode';
import {
  CoreNotFoundError,
  CoreResolver,
  resolveCorePackageDirectory,
} from '../../src/CoreResolver';

// require.resolve can retain a Windows 8.3 alias (RUNNER~1) while fs.realpath
// returns its long spelling. Compare the physical identity used by production.
async function canonicalPath(filePath: string): Promise<string> {
  const realPath = path.normalize(await fs.realpath(filePath));
  return process.platform === 'win32' ? realPath.toLowerCase() : realPath;
}

async function assertSamePhysicalPath(
  actual: string,
  expected: string,
): Promise<void> {
  assert.strictEqual(
    await canonicalPath(actual),
    await canonicalPath(expected),
  );
}

suite('local core resolver', () => {
  let temporaryDirectory: string;

  setup(async () => {
    temporaryDirectory = await fs.mkdtemp(
      path.join(os.tmpdir(), 'rslint-core-resolver-'),
    );
  });

  teardown(async function () {
    this.timeout(10_000);
    // Keep transient Windows EBUSY handling bounded and surface the final
    // failure instead of turning cleanup into a best-effort operation.
    await fs.rm(temporaryDirectory, {
      recursive: true,
      force: true,
      maxRetries: 10,
      retryDelay: 100,
    });
  });

  async function createPackageDirectory(root: string): Promise<string> {
    const packageDirectory = path.join(root, 'node_modules', '@rslint', 'core');
    await fs.mkdir(packageDirectory, { recursive: true });
    await fs.writeFile(
      path.join(packageDirectory, 'package.json'),
      JSON.stringify({ name: '@rslint/core', version: '1.0.0' }),
    );
    return packageDirectory;
  }

  function temporaryWorkspaceFolder(): WorkspaceFolder {
    return {
      uri: Uri.file(temporaryDirectory),
      name: 'temporary',
      index: 0,
    };
  }

  async function createSource(
    relativePath: string,
  ): Promise<Pick<TextDocument, 'uri'>> {
    const source = path.join(temporaryDirectory, relativePath);
    await fs.mkdir(path.dirname(source), { recursive: true });
    await fs.writeFile(source, 'const value = 1;\n');
    // Core resolution consumes only uri. A real VS Code TextDocument adds an
    // unrelated editor/file-watcher lifecycle and made Windows cleanup racy.
    return { uri: Uri.file(source) };
  }

  async function createLoadableCore(
    packageDirectory: string,
    version: string,
  ): Promise<void> {
    await fs.mkdir(packageDirectory, { recursive: true });
    const binaryPath = path.join(packageDirectory, 'rslint');
    await fs.writeFile(binaryPath, '#!/bin/sh\nexit 0\n', { mode: 0o755 });
    await Promise.all([
      fs.writeFile(
        path.join(packageDirectory, 'package.json'),
        JSON.stringify({
          name: '@rslint/core',
          version,
          type: 'module',
          exports: {
            './package.json': './package.json',
            './config-loader': './config-loader.js',
            './eslint-plugin': './eslint-plugin.js',
          },
        }),
      ),
      fs.writeFile(
        path.join(packageDirectory, 'config-loader.js'),
        [
          'export const CONFIG_DISCOVERY_PROTOCOL_VERSION = 3;',
          'export class ConfigModuleHost {}',
          `export function resolveRslintBinary() { return ${JSON.stringify(binaryPath)}; }`,
        ].join('\n'),
      ),
      fs.writeFile(
        path.join(packageDirectory, 'eslint-plugin.js'),
        'export async function createPluginLintHost() { throw new Error("not used"); }\n',
      ),
    ]);
  }

  test('selects the nearest node_modules installation', async () => {
    const rootCore = await createPackageDirectory(temporaryDirectory);
    const nestedRoot = path.join(temporaryDirectory, 'packages', 'app');
    const nestedCore = await createPackageDirectory(nestedRoot);
    const sourceDirectory = path.join(nestedRoot, 'src');
    await fs.mkdir(sourceDirectory, { recursive: true });

    await assertSamePhysicalPath(
      resolveCorePackageDirectory(sourceDirectory),
      nestedCore,
    );
    assert.notStrictEqual(nestedCore, rootCore);
  });

  test('uses an explicit core package directory verbatim', async () => {
    const configured = path.join(temporaryDirectory, 'vendor', 'rslint-core');
    await fs.mkdir(configured, { recursive: true });

    assert.strictEqual(
      resolveCorePackageDirectory(temporaryDirectory, configured),
      configured,
    );
  });

  test('loads a root installation for nested monorepo documents', async () => {
    const packageDirectory = path.join(
      temporaryDirectory,
      'node_modules',
      '@rslint',
      'core',
    );
    await createLoadableCore(packageDirectory, '1.2.3');
    const first = await createSource('packages/first/src/index.ts');
    const second = await createSource('packages/second/src/index.ts');

    const resolver = new CoreResolver();
    const [firstResolution, secondResolution] = await Promise.all([
      resolver.resolve(first, temporaryWorkspaceFolder()),
      resolver.resolve(second, temporaryWorkspaceFolder()),
    ]);

    assert.strictEqual(
      firstResolution.installation,
      secondResolution.installation,
    );
    assert.strictEqual(firstResolution.installation.version, '1.2.3');
    await assertSamePhysicalPath(
      firstResolution.installation.packageDirectory,
      packageDirectory,
    );
  });

  test('selects independent nested installations for different documents', async () => {
    const firstCore = path.join(
      temporaryDirectory,
      'packages',
      'first',
      'node_modules',
      '@rslint',
      'core',
    );
    const secondCore = path.join(
      temporaryDirectory,
      'packages',
      'second',
      'node_modules',
      '@rslint',
      'core',
    );
    await Promise.all([
      createLoadableCore(firstCore, '1.0.0'),
      createLoadableCore(secondCore, '2.0.0'),
    ]);
    const first = await createSource('packages/first/src/index.ts');
    const second = await createSource('packages/second/src/index.ts');

    const resolver = new CoreResolver();
    const [firstResolution, secondResolution] = await Promise.all([
      resolver.resolve(first, temporaryWorkspaceFolder()),
      resolver.resolve(second, temporaryWorkspaceFolder()),
    ]);

    assert.notStrictEqual(firstResolution.key, secondResolution.key);
    assert.strictEqual(firstResolution.installation.version, '1.0.0');
    assert.strictEqual(secondResolution.installation.version, '2.0.0');
  });

  test('does not merge separate package copies by version text alone', async () => {
    const firstCore = path.join(
      temporaryDirectory,
      'packages',
      'first',
      'node_modules',
      '@rslint',
      'core',
    );
    const secondCore = path.join(
      temporaryDirectory,
      'packages',
      'second',
      'node_modules',
      '@rslint',
      'core',
    );
    await Promise.all([
      createLoadableCore(firstCore, '1.0.0'),
      createLoadableCore(secondCore, '1.0.0'),
    ]);
    const first = await createSource('packages/first/src/index.ts');
    const second = await createSource('packages/second/src/index.ts');

    const resolver = new CoreResolver();
    const [firstResolution, secondResolution] = await Promise.all([
      resolver.resolve(first, temporaryWorkspaceFolder()),
      resolver.resolve(second, temporaryWorkspaceFolder()),
    ]);

    assert.notStrictEqual(firstResolution.key, secondResolution.key);
    assert.notStrictEqual(
      firstResolution.installation,
      secondResolution.installation,
    );
  });

  test('loads an exact relative corePath outside node_modules', async () => {
    const packageDirectory = path.join(
      temporaryDirectory,
      'vendor',
      'rslint-core',
    );
    await createLoadableCore(packageDirectory, '3.0.0');
    const document = await createSource('src/index.ts');

    const resolved = await new CoreResolver().resolve(
      document,
      temporaryWorkspaceFolder(),
      'vendor/rslint-core',
    );

    await assertSamePhysicalPath(
      resolved.installation.packageDirectory,
      packageDirectory,
    );
    assert.strictEqual(resolved.installation.version, '3.0.0');
  });

  test('rejects a core package whose binary is missing', async () => {
    const packageDirectory = path.join(
      temporaryDirectory,
      'node_modules',
      '@rslint',
      'core',
    );
    await createLoadableCore(packageDirectory, '1.0.0');
    await fs.rm(path.join(packageDirectory, 'rslint'));
    const document = await createSource('src/index.ts');

    await assert.rejects(
      new CoreResolver().resolve(document, temporaryWorkspaceFolder()),
      /Rslint binary does not exist/,
    );
  });

  test('does not execute a Yarn PnP resolver as a fallback', async () => {
    await fs.writeFile(
      path.join(temporaryDirectory, '.pnp.cjs'),
      'throw new Error("PnP resolver must not execute");\n',
    );

    assert.throws(
      () => resolveCorePackageDirectory(temporaryDirectory),
      CoreNotFoundError,
    );
  });

  test('reuses symlinks that point at one physical installation', async () => {
    const folder = workspace.workspaceFolders?.[0];
    assert.ok(folder, 'test requires a workspace folder');
    const actualCore = path.dirname(
      require.resolve('@rslint/core/package.json'),
    );
    const documents: Array<Pick<TextDocument, 'uri'>> = [];
    for (const name of ['a', 'b']) {
      const project = path.join(temporaryDirectory, name);
      const packageScope = path.join(project, 'node_modules', '@rslint');
      await fs.mkdir(packageScope, { recursive: true });
      await fs.symlink(
        actualCore,
        path.join(packageScope, 'core'),
        process.platform === 'win32' ? 'junction' : 'dir',
      );
      const source = path.join(project, 'index.ts');
      await fs.writeFile(source, `const ${name} = 1;\n`);
      documents.push({ uri: Uri.file(source) });
    }

    const resolver = new CoreResolver();
    const first = await resolver.resolve(documents[0], folder);
    const second = await resolver.resolve(documents[1], folder);
    assert.strictEqual(first.key, second.key);
    assert.strictEqual(first.installation, second.installation);
    assert.strictEqual(first.installation.version.length > 0, true);
    assert.strictEqual(path.isAbsolute(first.installation.binaryPath), true);
  });
});

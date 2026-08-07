import * as assert from 'node:assert';
import fs from 'node:fs/promises';
import os from 'node:os';
import path from 'node:path';
import {
  commands,
  Uri,
  window,
  workspace,
  type TextDocument,
  type WorkspaceFolder,
} from 'vscode';
import {
  CoreNotFoundError,
  CoreResolver,
  resolveCorePackageDirectory,
} from '../../src/CoreResolver';

suite('local core resolver', () => {
  let temporaryDirectory: string;
  let openedDocuments: TextDocument[];

  setup(async () => {
    temporaryDirectory = await fs.mkdtemp(
      path.join(os.tmpdir(), 'rslint-core-resolver-'),
    );
    openedDocuments = [];
  });

  teardown(async () => {
    for (const document of openedDocuments) {
      await window.showTextDocument(document, { preview: false });
      await commands.executeCommand('workbench.action.closeActiveEditor');
    }
    await fs.rm(temporaryDirectory, { recursive: true, force: true });
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

  async function openSource(relativePath: string): Promise<TextDocument> {
    const source = path.join(temporaryDirectory, relativePath);
    await fs.mkdir(path.dirname(source), { recursive: true });
    await fs.writeFile(source, 'const value = 1;\n');
    const document = await workspace.openTextDocument(Uri.file(source));
    openedDocuments.push(document);
    return document;
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
          'export const CONFIG_DISCOVERY_PROTOCOL_VERSION = 1;',
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

    assert.strictEqual(
      resolveCorePackageDirectory(sourceDirectory),
      await fs.realpath(nestedCore),
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
    const first = await openSource('packages/first/src/index.ts');
    const second = await openSource('packages/second/src/index.ts');

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
    assert.strictEqual(
      firstResolution.installation.packageDirectory,
      await fs.realpath(packageDirectory),
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
    const first = await openSource('packages/first/src/index.ts');
    const second = await openSource('packages/second/src/index.ts');

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
    const first = await openSource('packages/first/src/index.ts');
    const second = await openSource('packages/second/src/index.ts');

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
    const document = await openSource('src/index.ts');

    const resolved = await new CoreResolver().resolve(
      document,
      temporaryWorkspaceFolder(),
      'vendor/rslint-core',
    );

    assert.strictEqual(
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
    const document = await openSource('src/index.ts');

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
    const documents: TextDocument[] = [];
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
      documents.push(await workspace.openTextDocument(Uri.file(source)));
    }
    openedDocuments.push(...documents);

    const resolver = new CoreResolver();
    const first = await resolver.resolve(documents[0], folder);
    const second = await resolver.resolve(documents[1], folder);
    assert.strictEqual(first.key, second.key);
    assert.strictEqual(first.installation, second.installation);
    assert.strictEqual(first.installation.version.length > 0, true);
    assert.strictEqual(path.isAbsolute(first.installation.binaryPath), true);
  });
});

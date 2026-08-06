import assert from 'node:assert/strict';
import fs from 'node:fs/promises';
import path from 'node:path';

import * as vscode from 'vscode';

import { resolveRuntimeForDocument } from '../../src/RuntimeResolver';

function fakeDocument(filePath: string): vscode.TextDocument {
  return {
    uri: vscode.Uri.file(filePath),
    languageId: 'typescript',
  } as vscode.TextDocument;
}

async function setRuntimePath(
  folder: vscode.WorkspaceFolder,
  value: string,
): Promise<void> {
  await vscode.workspace
    .getConfiguration('rslint', folder.uri)
    .update('runtime.path', value, vscode.ConfigurationTarget.WorkspaceFolder);
}

suite('document-local runtime resolution', () => {
  test('resolves the core editor runtime from the document issuer', async () => {
    const folder = vscode.workspace.workspaceFolders?.[0];
    assert.ok(folder);
    const document = await vscode.workspace.openTextDocument(
      vscode.Uri.file(path.join(folder.uri.fsPath, 'src/index.ts')),
    );

    const first = await resolveRuntimeForDocument(document);
    const second = await resolveRuntimeForDocument(document);
    assert.ok(first, 'expected the sandbox-local @rslint/core to resolve');
    assert.ok(second);
    assert.equal(first.key, second.key);
    assert.equal(first.workspaceFolder.uri.toString(), folder.uri.toString());
    assert.match(first.entryPath, /editor-runtime\.js$/);
    assert.equal(first.source, 'node-modules');
  });

  test('pools one physical install but keeps equal-version nested installs separate', async () => {
    const folder = vscode.workspace.workspaceFolders?.[0];
    assert.ok(folder);
    const fixture = path.join(
      folder.uri.fsPath,
      `.runtime-resolution-${String(process.pid)}`,
    );
    const makeCore = async (project: string): Promise<void> => {
      const packageDirectory = path.join(
        fixture,
        project,
        'node_modules/@rslint/core',
      );
      await fs.mkdir(path.join(packageDirectory, 'dist'), { recursive: true });
      await fs.writeFile(
        path.join(packageDirectory, 'package.json'),
        JSON.stringify({
          name: '@rslint/core',
          version: '9.9.9-test',
          exports: {
            './editor-runtime': './dist/editor-runtime.js',
            './package.json': './package.json',
          },
        }),
      );
      await fs.writeFile(
        path.join(packageDirectory, 'dist/editor-runtime.js'),
        'export {}\n',
      );
    };
    const document = (relativePath: string): vscode.TextDocument =>
      fakeDocument(path.join(fixture, relativePath));

    try {
      await Promise.all([makeCore('a'), makeCore('b')]);
      const a1 = await resolveRuntimeForDocument(document('a/src/one.ts'));
      const a2 = await resolveRuntimeForDocument(document('a/src/two.ts'));
      const b = await resolveRuntimeForDocument(document('b/src/one.ts'));
      const root1 = await resolveRuntimeForDocument(
        document('without-local-core/one.ts'),
      );
      const root2 = await resolveRuntimeForDocument(
        document('without-local-core/two.ts'),
      );

      assert.ok(a1);
      assert.ok(a2);
      assert.ok(b);
      assert.ok(root1);
      assert.ok(root2);
      assert.equal(a1.key, a2.key, 'one physical install should be pooled');
      assert.notEqual(
        a1.key,
        b.key,
        'equal semver must not merge distinct dependency graphs',
      );
      assert.equal(
        root1.key,
        root2.key,
        'one root install should serve the whole remaining monorepo',
      );
    } finally {
      await fs.rm(fixture, { recursive: true, force: true });
    }
  });

  test('canonicalizes symlinked installs and changes identity when runtime bytes change', async () => {
    const folder = vscode.workspace.workspaceFolders?.[0];
    assert.ok(folder);
    const fixture = path.join(
      folder.uri.fsPath,
      `.runtime-symlink-${String(process.pid)}`,
    );
    const shared = path.join(fixture, 'shared-core');
    const entry = path.join(shared, 'dist/editor-runtime.js');
    const support = path.join(shared, 'dist/runtime-support.js');
    const bin = path.join(shared, 'bin/rslint.js');
    try {
      await fs.mkdir(path.dirname(entry), { recursive: true });
      await fs.mkdir(path.dirname(bin), { recursive: true });
      await fs.writeFile(
        path.join(shared, 'package.json'),
        JSON.stringify({
          name: '@rslint/core',
          version: '9.9.9-test',
          exports: {
            './editor-runtime': './dist/editor-runtime.js',
            './package.json': './package.json',
          },
        }),
      );
      await fs.writeFile(
        entry,
        "import './runtime-support.js';\nexport const generation = 1;\n",
      );
      await fs.writeFile(support, 'export const support = 1;\n');
      await fs.writeFile(bin, 'export const bin = 1;\n');
      for (const project of ['a', 'b']) {
        const scope = path.join(fixture, project, 'node_modules/@rslint');
        await fs.mkdir(scope, { recursive: true });
        await fs.symlink(
          shared,
          path.join(scope, 'core'),
          process.platform === 'win32' ? 'junction' : 'dir',
        );
      }

      const a = await resolveRuntimeForDocument(
        fakeDocument(path.join(fixture, 'a/src/index.ts')),
      );
      const b = await resolveRuntimeForDocument(
        fakeDocument(path.join(fixture, 'b/src/index.ts')),
      );
      assert.ok(a);
      assert.ok(b);
      assert.equal(a.key, b.key, 'symlinks to one physical core must pool');

      // Change both size and mtime so this assertion is stable even on file
      // systems with coarse timestamp resolution.
      await fs.writeFile(entry, 'export const generation = 222222;\n');
      const changed = await resolveRuntimeForDocument(
        fakeDocument(path.join(fixture, 'a/src/index.ts')),
      );
      assert.ok(changed);
      assert.notEqual(
        changed.key,
        a.key,
        'an in-place editor-runtime update must create a new generation',
      );

      await fs.writeFile(support, 'export const support = 222222222222222;\n');
      const changedChunk = await resolveRuntimeForDocument(
        fakeDocument(path.join(fixture, 'a/src/index.ts')),
      );
      assert.ok(changedChunk);
      assert.notEqual(
        changedChunk.key,
        changed.key,
        'a shared runtime chunk update must create a new generation',
      );
      assert.ok(changedChunk.watchPaths.includes(await fs.realpath(support)));

      await fs.writeFile(bin, 'export const bin = 333333333333333333333;\n');
      const changedBin = await resolveRuntimeForDocument(
        fakeDocument(path.join(fixture, 'b/src/index.ts')),
      );
      assert.ok(changedBin);
      assert.notEqual(
        changedBin.key,
        changedChunk.key,
        'a selected core bin update must create a new generation',
      );
      assert.ok(changedBin.watchPaths.includes(await fs.realpath(bin)));
    } finally {
      await fs.rm(fixture, { recursive: true, force: true });
    }
  });

  test('changes one shared install generation when its dependency lock changes', async () => {
    const folder = vscode.workspace.workspaceFolders?.[0];
    assert.ok(folder);
    const fixture = path.join(
      folder.uri.fsPath,
      `.runtime-lock-generation-${String(process.pid)}`,
    );
    const packageDirectory = path.join(fixture, 'node_modules/@rslint/core');
    const entryPath = path.join(packageDirectory, 'editor-runtime.js');
    const lockPath = path.join(fixture, 'pnpm-lock.yaml');
    try {
      await fs.mkdir(packageDirectory, { recursive: true });
      await fs.writeFile(
        path.join(fixture, 'package.json'),
        '{"name":"runtime-lock-generation","private":true}\n',
      );
      await fs.writeFile(
        path.join(packageDirectory, 'package.json'),
        JSON.stringify({
          name: '@rslint/core',
          version: '9.9.9-test',
          exports: { './editor-runtime': './editor-runtime.js' },
        }),
      );
      await fs.writeFile(entryPath, 'export {};\n');
      await fs.writeFile(lockPath, 'lockfileVersion: 1\n');

      const firstDocument = fakeDocument(
        path.join(fixture, 'packages/a/src/index.ts'),
      );
      const secondDocument = fakeDocument(
        path.join(fixture, 'packages/b/src/index.ts'),
      );
      const first = await resolveRuntimeForDocument(firstDocument);
      const shared = await resolveRuntimeForDocument(secondDocument);
      assert.ok(first);
      assert.ok(shared);
      assert.equal(first.key, shared.key);

      await fs.writeFile(lockPath, 'lockfileVersion: 222222\n');
      const changed = await resolveRuntimeForDocument(firstDocument);
      assert.ok(changed);
      assert.notEqual(
        changed.key,
        first.key,
        'transitive dependency changes must replace a cached sidecar',
      );
      const stillShared = await resolveRuntimeForDocument(secondDocument);
      assert.ok(stillShared);
      assert.equal(changed.key, stillShared.key);
    } finally {
      await fs.rm(fixture, { recursive: true, force: true });
    }
  });

  test('supports absolute and workspace-relative configured core paths', async function () {
    // Updating a workspace setting also exercises the live RuntimeManager
    // watcher. On a loaded extension host that reconciliation can legitimately
    // overlap both explicit resolutions, so Mocha's 2s default is too tight.
    this.timeout(10_000);
    const folder = vscode.workspace.workspaceFolders?.[0];
    assert.ok(folder);
    const document = await vscode.workspace.openTextDocument(
      vscode.Uri.file(path.join(folder.uri.fsPath, 'src/index.ts')),
    );
    const discovered = await resolveRuntimeForDocument(document);
    assert.ok(discovered);
    const packageDirectory = path.dirname(discovered.packagePath);

    try {
      await setRuntimePath(folder, packageDirectory);
      const absolute = await resolveRuntimeForDocument(document);
      assert.ok(absolute);
      assert.equal(absolute.source, 'configured');
      assert.equal(absolute.entryPath, discovered.entryPath);

      await setRuntimePath(
        folder,
        path.relative(folder.uri.fsPath, packageDirectory),
      );
      const relative = await resolveRuntimeForDocument(document);
      assert.ok(relative);
      assert.equal(relative.source, 'configured');
      assert.equal(relative.entryPath, discovered.entryPath);
    } finally {
      await setRuntimePath(folder, '');
    }
  });

  test('rejects a configured directory that is not a complete @rslint/core package', async () => {
    const folder = vscode.workspace.workspaceFolders?.[0];
    assert.ok(folder);
    const document = await vscode.workspace.openTextDocument(
      vscode.Uri.file(path.join(folder.uri.fsPath, 'src/index.ts')),
    );
    const fixture = path.join(
      folder.uri.fsPath,
      `.runtime-configured-invalid-${String(process.pid)}`,
    );
    try {
      await fs.mkdir(fixture, { recursive: true });
      await fs.writeFile(
        path.join(fixture, 'package.json'),
        JSON.stringify({ name: '@not-rslint/core', version: '1.0.0' }),
      );
      await setRuntimePath(folder, fixture);
      await assert.rejects(
        resolveRuntimeForDocument(document),
        /must point to an @rslint\/core package directory/,
      );

      await fs.writeFile(
        path.join(fixture, 'package.json'),
        JSON.stringify({ name: '@rslint/core', version: '1.0.0' }),
      );
      await assert.rejects(
        resolveRuntimeForDocument(document),
        /editor[- ]runtime|exports|Cannot find module|Package subpath/,
      );
    } finally {
      await setRuntimePath(folder, '');
      await fs.rm(fixture, { recursive: true, force: true });
    }
  });

  test('configured core overrides a different nearest node_modules install', async () => {
    const folder = vscode.workspace.workspaceFolders?.[0];
    assert.ok(folder);
    const fixture = path.join(
      folder.uri.fsPath,
      `.runtime-configured-priority-${String(process.pid)}`,
    );
    const writeCore = async (
      directory: string,
      label: string,
    ): Promise<void> => {
      await fs.mkdir(directory, { recursive: true });
      await fs.writeFile(
        path.join(directory, 'package.json'),
        JSON.stringify({
          name: '@rslint/core',
          version: label,
          exports: { './editor-runtime': './editor-runtime.js' },
        }),
      );
      await fs.writeFile(
        path.join(directory, 'editor-runtime.js'),
        `export const label = ${JSON.stringify(label)};\n`,
      );
    };
    const configured = path.join(fixture, 'configured-core');
    const nearest = path.join(fixture, 'project/node_modules/@rslint/core');
    try {
      await Promise.all([
        writeCore(configured, 'configured'),
        writeCore(nearest, 'nearest'),
      ]);
      await setRuntimePath(folder, configured);
      const resolved = await resolveRuntimeForDocument(
        fakeDocument(path.join(fixture, 'project/src/index.ts')),
      );
      assert.ok(resolved);
      assert.equal(resolved.source, 'configured');
      assert.equal(
        resolved.entryPath,
        await fs.realpath(path.join(configured, 'editor-runtime.js')),
      );
    } finally {
      await setRuntimePath(folder, '');
      await fs.rm(fixture, { recursive: true, force: true });
    }
  });

  test('configured core keeps document-local Yarn PnP execution domains', async function () {
    this.timeout(10_000);
    const folder = vscode.workspace.workspaceFolders?.[0];
    assert.ok(folder);
    const seed = await vscode.workspace.openTextDocument(
      vscode.Uri.file(path.join(folder.uri.fsPath, 'src/index.ts')),
    );
    const installed = await resolveRuntimeForDocument(seed);
    assert.ok(installed);
    const configured = path.dirname(installed.packagePath);
    const fixture = path.join(
      folder.uri.fsPath,
      `.runtime-configured-pnp-${String(process.pid)}`,
    );
    const domainA = path.join(fixture, 'a');
    const domainB = path.join(fixture, 'b');
    try {
      for (const domain of [domainA, domainB]) {
        await fs.mkdir(path.join(domain, 'src'), { recursive: true });
        await fs.writeFile(
          path.join(domain, '.pnp.cjs'),
          'module.exports = { resolveRequest() { return null; } };\n',
        );
      }
      await fs.writeFile(
        path.join(domainA, '.pnp.loader.mjs'),
        'export async function resolve(s, c, next) { return next(s, c); }\n',
      );
      await setRuntimePath(folder, configured);

      const a1 = await resolveRuntimeForDocument(
        fakeDocument(path.join(domainA, 'src/one.ts')),
      );
      const a2 = await resolveRuntimeForDocument(
        fakeDocument(path.join(domainA, 'src/two.ts')),
      );
      const b = await resolveRuntimeForDocument(
        fakeDocument(path.join(domainB, 'src/one.ts')),
      );
      const canonicalDomainA = path.dirname(
        await fs.realpath(path.join(domainA, '.pnp.cjs')),
      );
      const canonicalDomainB = path.dirname(
        await fs.realpath(path.join(domainB, '.pnp.cjs')),
      );
      assert.ok(a1);
      assert.ok(a2);
      assert.ok(b);
      assert.equal(a1.entryPath, b.entryPath);
      assert.equal(a1.key, a2.key, 'one PnP issuer domain should still pool');
      assert.notEqual(
        a1.key,
        b.key,
        'different PnP graphs must not share one configured-core process',
      );
      assert.equal(a1.workingDirectory, canonicalDomainA);
      assert.equal(a2.workingDirectory, canonicalDomainA);
      assert.equal(b.workingDirectory, canonicalDomainB);
      assert.equal(
        a1.pnpPath,
        await fs.realpath(path.join(domainA, '.pnp.cjs')),
      );
      assert.equal(
        b.pnpPath,
        await fs.realpath(path.join(domainB, '.pnp.cjs')),
      );
      assert.deepEqual(a1.nodeArgs, [
        '--require',
        await fs.realpath(path.join(domainA, '.pnp.cjs')),
        '--experimental-loader',
        await fs.realpath(path.join(domainA, '.pnp.loader.mjs')),
      ]);
      assert.deepEqual(b.nodeArgs, [
        '--require',
        await fs.realpath(path.join(domainB, '.pnp.cjs')),
      ]);
    } finally {
      await setRuntimePath(folder, '');
      await fs.rm(fixture, { recursive: true, force: true });
    }
  });

  test('does not bypass a broken nearest node_modules core with an ancestor copy', async () => {
    const folder = vscode.workspace.workspaceFolders?.[0];
    assert.ok(folder);
    const fixture = path.join(
      folder.uri.fsPath,
      `.runtime-broken-nearest-${String(process.pid)}`,
    );
    const packageDirectory = path.join(fixture, 'node_modules/@rslint/core');
    try {
      await fs.mkdir(packageDirectory, { recursive: true });
      await fs.writeFile(
        path.join(packageDirectory, 'package.json'),
        JSON.stringify({ name: '@rslint/core', version: 'broken' }),
      );
      await assert.rejects(
        resolveRuntimeForDocument(
          fakeDocument(path.join(fixture, 'src/index.ts')),
        ),
        /does not export.*editor-runtime/,
      );
    } finally {
      await fs.rm(fixture, { recursive: true, force: true });
    }
  });

  test('carries the complete Yarn PnP ESM domain into the sidecar', async () => {
    const folder = vscode.workspace.workspaceFolders?.[0];
    assert.ok(folder);
    const fixture = path.join(
      folder.uri.fsPath,
      `.runtime-pnp-${String(process.pid)}`,
    );
    const entryPath = path.join(fixture, 'core/dist/editor-runtime.js');
    const packagePath = path.join(fixture, 'core/package.json');
    const pnpPath = path.join(fixture, '.pnp.cjs');
    const loaderPath = path.join(fixture, '.pnp.loader.mjs');
    const dataPath = path.join(fixture, '.pnp.data.json');
    try {
      await fs.mkdir(path.dirname(entryPath), { recursive: true });
      await fs.mkdir(path.join(fixture, 'src'), { recursive: true });
      await fs.writeFile(entryPath, 'export {}\n');
      await fs.writeFile(
        packagePath,
        JSON.stringify({ name: '@rslint/core', version: '9.9.9-pnp' }),
      );
      await fs.writeFile(loaderPath, 'export async function resolve() {}\n');
      await fs.writeFile(dataPath, '{"generation":1}\n');
      await fs.writeFile(
        pnpPath,
        `module.exports = { resolveRequest(request) {\n` +
          `  if (request === '@rslint/core/editor-runtime') return ${JSON.stringify(entryPath)};\n` +
          `  if (request === '@rslint/core/package.json') return ${JSON.stringify(packagePath)};\n` +
          `  return null;\n` +
          `} };\n`,
      );

      const resolved = await resolveRuntimeForDocument({
        uri: vscode.Uri.file(path.join(fixture, 'src/index.ts')),
        languageId: 'typescript',
      } as vscode.TextDocument);
      const [canonicalPnpPath, canonicalLoaderPath] = await Promise.all([
        fs.realpath(pnpPath),
        fs.realpath(loaderPath),
      ]);

      assert.ok(resolved);
      assert.equal(resolved.source, 'pnp');
      assert.equal(resolved.workingDirectory, path.dirname(canonicalPnpPath));
      assert.equal(resolved.pnpPath, canonicalPnpPath);
      assert.deepEqual(resolved.nodeArgs, [
        '--require',
        canonicalPnpPath,
        '--experimental-loader',
        canonicalLoaderPath,
      ]);

      await fs.writeFile(dataPath, '{"generation":222222}\n');
      const changedData = await resolveRuntimeForDocument(
        fakeDocument(path.join(fixture, 'src/index.ts')),
      );
      assert.ok(changedData);
      assert.notEqual(
        changedData.key,
        resolved.key,
        'a PnP data update must invalidate the runtime generation',
      );

      await fs.rm(loaderPath);
      const withoutLoader = await resolveRuntimeForDocument(
        fakeDocument(path.join(fixture, 'src/index.ts')),
      );
      assert.ok(withoutLoader);
      assert.deepEqual(withoutLoader.nodeArgs, ['--require', canonicalPnpPath]);
      assert.notEqual(withoutLoader.key, changedData.key);
    } finally {
      await fs.rm(fixture, { recursive: true, force: true });
    }
  });

  test('reloads a changed PnP API instead of using Node require cache', async () => {
    const folder = vscode.workspace.workspaceFolders?.[0];
    assert.ok(folder);
    const fixture = path.join(
      folder.uri.fsPath,
      `.runtime-pnp-refresh-${String(process.pid)}`,
    );
    const pnpPath = path.join(fixture, '.pnp.cjs');
    const loadMarker = path.join(fixture, 'pnp-loads.log');
    const makeCore = async (name: string): Promise<[string, string]> => {
      const directory = path.join(fixture, name);
      const entry = path.join(directory, 'editor-runtime.js');
      const manifest = path.join(directory, 'package.json');
      await fs.mkdir(directory, { recursive: true });
      await fs.writeFile(
        entry,
        `export const name = ${JSON.stringify(name)};\n`,
      );
      await fs.writeFile(
        manifest,
        JSON.stringify({ name: '@rslint/core', version: name }),
      );
      return [entry, manifest];
    };
    const writePnp = async (entry: string, manifest: string): Promise<void> => {
      await fs.writeFile(
        pnpPath,
        `require('node:fs').appendFileSync(${JSON.stringify(loadMarker)}, 'load\\n');\n` +
          `module.exports = { resolveRequest(request) {\n` +
          `  if (request === '@rslint/core/editor-runtime') return ${JSON.stringify(entry)};\n` +
          `  if (request === '@rslint/core/package.json') return ${JSON.stringify(manifest)};\n` +
          `  return null;\n` +
          `} };\n`,
      );
    };

    try {
      await fs.mkdir(path.join(fixture, 'src'), { recursive: true });
      const firstCore = await makeCore('first');
      const secondCore = await makeCore('second-longer');
      await writePnp(...firstCore);
      const document = fakeDocument(path.join(fixture, 'src/index.ts'));
      const first = await resolveRuntimeForDocument(document);
      assert.ok(first);
      assert.equal(first.entryPath, await fs.realpath(firstCore[0]));
      const unchanged = await resolveRuntimeForDocument(document);
      assert.ok(unchanged);
      assert.equal(unchanged.key, first.key);
      assert.equal(
        (await fs.readFile(loadMarker, 'utf8')).trim().split('\n').length,
        1,
        'an unchanged PnP graph must not be rehydrated for every document',
      );

      await writePnp(...secondCore);
      const second = await resolveRuntimeForDocument(document);
      assert.ok(second);
      assert.equal(second.entryPath, await fs.realpath(secondCore[0]));
      assert.notEqual(second.key, first.key);
      assert.equal(
        (await fs.readFile(loadMarker, 'utf8')).trim().split('\n').length,
        2,
        'a changed PnP graph must load exactly one new API generation',
      );
    } finally {
      await fs.rm(fixture, { recursive: true, force: true });
    }
  });

  test('bounds PnP graph caching and rehydrates an evicted domain', async function () {
    this.timeout(10_000);
    const folder = vscode.workspace.workspaceFolders?.[0];
    assert.ok(folder);
    const fixture = path.join(
      folder.uri.fsPath,
      `.runtime-pnp-lru-${String(process.pid)}`,
    );
    const firstLoadMarker = path.join(fixture, 'first-loads.log');
    const makeDomain = async (index: number): Promise<vscode.TextDocument> => {
      const domain = path.join(fixture, `domain-${String(index)}`);
      const core = path.join(domain, 'core');
      const entry = path.join(core, 'editor-runtime.js');
      const manifest = path.join(core, 'package.json');
      const source = path.join(domain, 'src/index.ts');
      await fs.mkdir(path.dirname(source), { recursive: true });
      await fs.mkdir(core, { recursive: true });
      await fs.writeFile(entry, `export const domain = ${String(index)};\n`);
      await fs.writeFile(
        manifest,
        JSON.stringify({ name: '@rslint/core', version: `domain-${index}` }),
      );
      await fs.writeFile(
        path.join(domain, '.pnp.cjs'),
        (index === 0
          ? `require('node:fs').appendFileSync(${JSON.stringify(firstLoadMarker)}, 'load\\n');\n`
          : '') +
          `module.exports = { resolveRequest(request) {\n` +
          `  if (request === '@rslint/core/editor-runtime') return ${JSON.stringify(entry)};\n` +
          `  if (request === '@rslint/core/package.json') return ${JSON.stringify(manifest)};\n` +
          `  return null;\n` +
          `} };\n`,
      );
      return fakeDocument(source);
    };

    try {
      const firstDocument = await makeDomain(0);
      const first = await resolveRuntimeForDocument(firstDocument);
      assert.ok(first);
      // The cache holds 16 APIs. Seventeen untouched domains guarantee the
      // original one is cold-evicted regardless of earlier tests' entries.
      for (let index = 1; index <= 17; index++) {
        const resolved = await resolveRuntimeForDocument(
          await makeDomain(index),
        );
        assert.ok(resolved);
      }

      const rehydrated = await resolveRuntimeForDocument(firstDocument);
      assert.ok(rehydrated);
      assert.equal(
        rehydrated.key,
        first.key,
        'cache eviction must affect only performance, not runtime identity',
      );
      assert.equal(
        (await fs.readFile(firstLoadMarker, 'utf8')).trim().split('\n').length,
        2,
        'the cold domain should be loaded once initially and once after eviction',
      );
    } finally {
      await fs.rm(fixture, { recursive: true, force: true });
    }
  });

  test('accepts runtime and manifest paths inside Yarn ZipFS', async () => {
    const folder = vscode.workspace.workspaceFolders?.[0];
    assert.ok(folder);
    const fixture = path.join(
      folder.uri.fsPath,
      `.runtime-pnp-zipfs-${String(process.pid)}`,
    );
    const entryPath = path.join(
      fixture,
      '.yarn/cache/core.zip/node_modules/@rslint/core/dist/editor-runtime.js',
    );
    const packagePath = path.join(
      fixture,
      '.yarn/cache/core.zip/node_modules/@rslint/core/package.json',
    );
    const archivePath = path.join(fixture, '.yarn/cache/core.zip');
    const pnpPath = path.join(fixture, '.pnp.cjs');
    try {
      await Promise.all([
        fs.mkdir(path.join(fixture, 'src'), { recursive: true }),
        fs.mkdir(path.dirname(archivePath), { recursive: true }),
      ]);
      await fs.writeFile(archivePath, 'zip-generation-one');
      await fs.writeFile(
        pnpPath,
        `module.exports = { resolveRequest(request) {\n` +
          `  if (request === '@rslint/core/editor-runtime') return ${JSON.stringify(entryPath)};\n` +
          `  if (request === '@rslint/core/package.json') return ${JSON.stringify(packagePath)};\n` +
          `  return null;\n` +
          `} };\n`,
      );

      const resolved = await resolveRuntimeForDocument(
        fakeDocument(path.join(fixture, 'src/index.ts')),
      );
      assert.ok(resolved);
      assert.equal(resolved.source, 'pnp');
      assert.equal(resolved.entryPath, path.normalize(entryPath));
      assert.equal(resolved.packagePath, packagePath);
      assert.deepEqual(resolved.nodeArgs, [
        '--require',
        await fs.realpath(pnpPath),
      ]);
      assert.ok(
        resolved.watchPaths.includes(await fs.realpath(archivePath)),
        'the native ZipFS archive must invalidate its virtual entry files',
      );
      await assert.rejects(fs.access(entryPath), /ENOENT|ENOTDIR/);

      await fs.writeFile(archivePath, 'zip-generation-two-is-different');
      const updated = await resolveRuntimeForDocument(
        fakeDocument(path.join(fixture, 'src/index.ts')),
      );
      assert.ok(updated);
      assert.notEqual(
        updated.key,
        resolved.key,
        'replacing ZipFS bytes in place must create a new runtime generation',
      );
    } finally {
      await fs.rm(fixture, { recursive: true, force: true });
    }
  });

  test('treats a PnP boundary without @rslint/core as authoritative', async () => {
    const folder = vscode.workspace.workspaceFolders?.[0];
    assert.ok(folder);
    const fixture = path.join(
      folder.uri.fsPath,
      `.runtime-pnp-missing-${String(process.pid)}`,
    );
    try {
      await fs.mkdir(path.join(fixture, 'src'), { recursive: true });
      await fs.writeFile(
        path.join(fixture, '.pnp.cjs'),
        'module.exports = { resolveRequest() { return null; } };\n',
      );
      const resolved = await resolveRuntimeForDocument(
        fakeDocument(path.join(fixture, 'src/index.ts')),
      );
      assert.equal(
        resolved,
        undefined,
        'must not escape PnP into an ancestor node_modules install',
      );

      const entry = path.join(fixture, 'editor-runtime.js');
      await fs.writeFile(entry, 'export {};\n');
      await fs.writeFile(
        path.join(fixture, '.pnp.cjs'),
        `module.exports = { resolveRequest(request) {\n` +
          `  return request === '@rslint/core/editor-runtime' ? ${JSON.stringify(entry)} : null;\n` +
          `} };\n`,
      );
      await assert.rejects(
        resolveRuntimeForDocument(
          fakeDocument(path.join(fixture, 'src/index.ts')),
        ),
        /resolved an incomplete @rslint\/core package/,
      );
    } finally {
      await fs.rm(fixture, { recursive: true, force: true });
    }
  });
});

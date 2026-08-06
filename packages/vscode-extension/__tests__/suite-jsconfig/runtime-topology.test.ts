import assert from 'node:assert/strict';
import fs from 'node:fs/promises';
import os from 'node:os';
import path from 'node:path';
import { pathToFileURL } from 'node:url';

import * as vscode from 'vscode';

import { resolveRuntimeForDocument } from '../../src/RuntimeResolver';
import {
  getRslintDiagnostics,
  waitForRslintDiagnostics,
} from '../utils/diagnostics';

function markerEvents(contents: string, kind: 'start' | 'stop'): string[] {
  return contents.split(/\r?\n/u).filter((line) => line.startsWith(`${kind}:`));
}

async function readMarker(markerPath: string): Promise<string> {
  try {
    return await fs.readFile(markerPath, 'utf8');
  } catch (error) {
    if ((error as NodeJS.ErrnoException).code === 'ENOENT') return '';
    throw error;
  }
}

async function waitForMarker(
  markerPath: string,
  predicate: (contents: string) => boolean,
  description: string,
  timeoutMs = 30_000,
): Promise<string> {
  const deadline = Date.now() + timeoutMs;
  let contents = '';
  while (Date.now() < deadline) {
    contents = await readMarker(markerPath);
    if (predicate(contents)) return contents;
    await new Promise((resolve) => setTimeout(resolve, 50));
  }
  assert.fail(`${description}; marker contents:\n${contents}`);
}

async function writeCoreWrapper(
  packageDirectory: string,
  markerPath: string,
  actualEntryPath: string,
  version: string,
  crashOncePath?: string,
  configGatePath?: string,
): Promise<void> {
  await fs.mkdir(packageDirectory, { recursive: true });
  await fs.writeFile(
    path.join(packageDirectory, 'package.json'),
    JSON.stringify({
      name: '@rslint/core',
      version,
      type: 'module',
      exports: {
        './editor-runtime': './editor-runtime.js',
        './package.json': './package.json',
      },
    }),
  );
  await fs.writeFile(
    path.join(packageDirectory, 'runtime-support.js'),
    `export const runtimeSupportGeneration = ${JSON.stringify(version)};\n`,
  );
  await fs.writeFile(
    path.join(packageDirectory, 'editor-runtime.js'),
    `import fs from 'node:fs';\n` +
      `import './runtime-support.js';\n` +
      `import { runEditorRuntime } from ${JSON.stringify(pathToFileURL(actualEntryPath).href)};\n` +
      `fs.appendFileSync(${JSON.stringify(markerPath)}, 'start:' + String(process.pid) + '\\n');\n` +
      `fs.appendFileSync(${JSON.stringify(markerPath)}, 'cwd:' + process.cwd() + '\\n');\n` +
      (crashOncePath
        ? `if (!fs.existsSync(${JSON.stringify(crashOncePath)})) {\n` +
          `  fs.writeFileSync(${JSON.stringify(crashOncePath)}, 'crashed\\n');\n` +
          `  setTimeout(() => process.exit(23), 1500);\n` +
          `}\n`
        : '') +
      (configGatePath
        ? `process.env.RSLINT_TEST_CONFIG_GATE = ${JSON.stringify(configGatePath)};\n` +
          `process.env.RSLINT_TEST_CONFIG_MARKER = ${JSON.stringify(markerPath)};\n`
        : '') +
      `try { process.exitCode = await runEditorRuntime(); } finally {\n` +
      `  fs.appendFileSync(${JSON.stringify(markerPath)}, 'stop:' + String(process.pid) + '\\n');\n` +
      `}\n`,
  );
}

async function closeDocuments(
  documents: readonly vscode.TextDocument[],
): Promise<void> {
  for (const document of documents) {
    if (!vscode.workspace.textDocuments.includes(document)) continue;
    await vscode.window.showTextDocument(document, {
      preview: false,
      preserveFocus: false,
    });
    await vscode.commands.executeCommand('workbench.action.closeActiveEditor');
  }
  await vscode.commands.executeCommand('workbench.action.closeAllEditors');
}

suite('runtime topology lifecycle', function () {
  this.timeout(90_000);

  test('pools root core, starts PnP, restarts crashes, and migrates both ways', async () => {
    const folder = vscode.workspace.workspaceFolders?.[0];
    assert.ok(folder);
    const seedDocument = await vscode.workspace.openTextDocument(
      vscode.Uri.file(path.join(folder.uri.fsPath, 'src/index.ts')),
    );
    const installed = await resolveRuntimeForDocument(seedDocument);
    assert.ok(installed, 'the test profile must provide a complete core');

    const fixture = path.join(
      folder.uri.fsPath,
      `.runtime-topology-${String(process.pid)}`,
    );
    const rootMarker = path.join(fixture, 'root-runtime.log');
    const nestedMarker = path.join(fixture, 'nested-runtime.log');
    const migratingMarker = path.join(fixture, 'migrating-runtime.log');
    const crashMarker = path.join(fixture, 'crash-runtime.log');
    const pnpMarker = path.join(fixture, 'pnp-runtime.log');
    const pnpRequireMarker = path.join(fixture, 'pnp-require.log');
    const pnpLoaderMarker = path.join(fixture, 'pnp-loader.log');
    const crashOncePath = path.join(fixture, 'crash-once.state');
    const rootCore = path.join(fixture, 'node_modules/@rslint/core');
    const nestedCore = path.join(
      fixture,
      'packages/nested/node_modules/@rslint/core',
    );
    const migratingCore = path.join(
      fixture,
      'packages/migrating/node_modules/@rslint/core',
    );
    const crashCore = path.join(
      fixture,
      'packages/crash/node_modules/@rslint/core',
    );
    const pnpProject = path.join(fixture, 'packages/pnp');
    const pnpCore = path.join(pnpProject, '.core');
    const documentPaths = [
      path.join(fixture, 'packages/root-a/src/index.ts'),
      path.join(fixture, 'packages/root-b/src/index.ts'),
      path.join(fixture, 'packages/nested/src/index.ts'),
      path.join(fixture, 'packages/migrating/src/index.ts'),
      path.join(fixture, 'packages/crash/src/index.ts'),
      path.join(pnpProject, 'src/index.ts'),
    ];
    const documents: vscode.TextDocument[] = [];
    let managerDisabled = false;

    try {
      await Promise.all([
        writeCoreWrapper(
          rootCore,
          rootMarker,
          installed.entryPath,
          '9.9.9-equal-version',
        ),
        writeCoreWrapper(
          nestedCore,
          nestedMarker,
          installed.entryPath,
          '9.9.9-equal-version',
        ),
        writeCoreWrapper(
          crashCore,
          crashMarker,
          installed.entryPath,
          '9.9.9-equal-version',
          crashOncePath,
        ),
        writeCoreWrapper(pnpCore, pnpMarker, installed.entryPath, '11.0.0-pnp'),
      ]);
      const pnpEntry = path.join(pnpCore, 'editor-runtime.js');
      const pnpManifest = path.join(pnpCore, 'package.json');
      await fs.writeFile(
        path.join(pnpProject, '.pnp.cjs'),
        `const fs = require('node:fs');\n` +
          `fs.appendFileSync(${JSON.stringify(pnpRequireMarker)}, 'required:' + String(process.pid) + '\\n');\n` +
          `module.exports = { resolveRequest(request) {\n` +
          `  if (request === '@rslint/core/editor-runtime') return ${JSON.stringify(pnpEntry)};\n` +
          `  if (request === '@rslint/core/package.json') return ${JSON.stringify(pnpManifest)};\n` +
          `  return null;\n` +
          `} };\n`,
      );
      await fs.writeFile(
        path.join(pnpProject, '.pnp.loader.mjs'),
        `import fs from 'node:fs';\n` +
          `fs.appendFileSync(${JSON.stringify(pnpLoaderMarker)}, 'loaded:' + String(process.pid) + '\\n');\n` +
          `export async function resolve(specifier, context, nextResolve) {\n` +
          `  return nextResolve(specifier, context);\n` +
          `}\n` +
          `export async function load(url, context, nextLoad) {\n` +
          `  return nextLoad(url, context);\n` +
          `}\n`,
      );
      await Promise.all(
        documentPaths.map(async (filePath) => {
          await fs.mkdir(path.dirname(filePath), { recursive: true });
          await fs.writeFile(filePath, 'const value: any = 1;\n');
        }),
      );

      for (const filePath of documentPaths) {
        documents.push(
          await vscode.workspace.openTextDocument(vscode.Uri.file(filePath)),
        );
      }

      const rootContents = await waitForMarker(
        rootMarker,
        (contents) => markerEvents(contents, 'start').length === 1,
        'the shared root runtime did not start exactly once',
      );
      const nestedContents = await waitForMarker(
        nestedMarker,
        (contents) => markerEvents(contents, 'start').length === 1,
        'the nested runtime did not start exactly once',
      );
      assert.equal(
        markerEvents(rootContents, 'start').length,
        1,
        'three documents using one physical root install must share a process',
      );
      assert.equal(
        markerEvents(nestedContents, 'start').length,
        1,
        'an equal-semver but physically distinct install needs its own process',
      );

      // Local builds and package-manager updates can replace core's exported
      // entry without touching package.json or a lockfile. Exact runtime-file
      // watching must roll the generation even though the install path stays
      // the same.
      await fs.appendFile(
        path.join(nestedCore, 'editor-runtime.js'),
        '// rebuilt entry generation\n',
      );
      await waitForMarker(
        nestedMarker,
        (contents) =>
          markerEvents(contents, 'start').length === 2 &&
          markerEvents(contents, 'stop').length >= 1,
        'an in-workspace editor-runtime rebuild did not replace its sidecar',
      );
      const restartedContents = await waitForMarker(
        crashMarker,
        (contents) => markerEvents(contents, 'start').length >= 2,
        'the language client did not restart a crashed core sidecar',
      );
      assert.equal(
        markerEvents(restartedContents, 'start').length,
        2,
        'one crash must create exactly one replacement generation',
      );
      const pnpContents = await waitForMarker(
        pnpMarker,
        (contents) => markerEvents(contents, 'start').length === 1,
        'the PnP-resolved sidecar did not start with its require/loader domain',
      );
      assert.equal(markerEvents(pnpContents, 'start').length, 1);
      const canonicalPnpProject = await fs.realpath(pnpProject);
      assert.ok(
        pnpContents.includes(`cwd:${canonicalPnpProject}\n`),
        `PnP sidecar used the wrong working directory:\n${pnpContents}`,
      );
      const pnpPid = markerEvents(pnpContents, 'start')[0].slice(
        'start:'.length,
      );
      await waitForMarker(
        pnpRequireMarker,
        (contents) => contents.includes(`required:${pnpPid}`),
        'the PnP require hook was not installed in the sidecar',
      );
      await waitForMarker(
        pnpLoaderMarker,
        (contents) => contents.includes(`loaded:${pnpPid}`),
        'the PnP ESM loader was not installed in the sidecar',
      );

      // Adding a nearer complete core and changing the owning package manifest
      // is a real topology event. The already-open document must move without
      // requiring an editor reopen.
      await writeCoreWrapper(
        migratingCore,
        migratingMarker,
        installed.entryPath,
        '10.0.0-nearer',
      );
      const migratingProject = path.join(fixture, 'packages/migrating');
      await fs.writeFile(
        path.join(migratingProject, 'package.json'),
        '{"name":"migrating","private":true}\n',
      );
      let migratingContents = await waitForMarker(
        migratingMarker,
        (contents) => markerEvents(contents, 'start').length === 1,
        'the open document did not migrate to its new nearest core',
      );
      assert.equal(markerEvents(migratingContents, 'start').length, 1);

      // Removing that install must hand the same document back to the already
      // running root runtime. The nested sidecar exiting proves the manager
      // released it; the root marker remaining at one start proves reuse.
      await fs.rm(migratingCore, { recursive: true, force: true });
      await fs.writeFile(
        path.join(migratingProject, 'package.json'),
        '{"name":"migrating","private":true,"generation":2}\n',
      );
      migratingContents = await waitForMarker(
        migratingMarker,
        (contents) => markerEvents(contents, 'stop').length === 1,
        'the removed nested runtime was not released',
      );
      assert.equal(markerEvents(migratingContents, 'stop').length, 1);
      const reusedRootContents = await readMarker(rootMarker);
      assert.equal(
        markerEvents(reusedRootContents, 'start').length,
        1,
        'falling back must reuse the existing root runtime',
      );

      const returned = await resolveRuntimeForDocument(documents[3]);
      assert.ok(returned);
      assert.equal(
        returned.entryPath,
        await fs.realpath(path.join(rootCore, 'editor-runtime.js')),
      );
    } finally {
      // VS Code may retain programmatically opened TextDocuments after their
      // visible editors close. Exercise the real enable/disable lifecycle so
      // cleanup never relies on undocumented document-cache eviction.
      await vscode.workspace
        .getConfiguration('rslint', folder.uri)
        .update('enable', false, vscode.ConfigurationTarget.WorkspaceFolder);
      managerDisabled = true;
      await closeDocuments(documents);
      if (await readMarker(nestedMarker)) {
        await waitForMarker(
          nestedMarker,
          (contents) => markerEvents(contents, 'stop').length >= 1,
          'the nested test runtime did not stop during cleanup',
        );
      }
      if (await readMarker(rootMarker)) {
        await waitForMarker(
          rootMarker,
          (contents) => markerEvents(contents, 'stop').length >= 1,
          'the root test runtime did not stop during cleanup',
        );
      }
      if (await readMarker(crashMarker)) {
        await waitForMarker(
          crashMarker,
          (contents) => markerEvents(contents, 'stop').length >= 1,
          'the restarted test runtime did not stop during cleanup',
        );
      }
      if (await readMarker(pnpMarker)) {
        await waitForMarker(
          pnpMarker,
          (contents) => markerEvents(contents, 'stop').length >= 1,
          'the PnP test runtime did not stop during cleanup',
        );
      }
      await fs.rm(fixture, { recursive: true, force: true });
      if (managerDisabled) {
        await vscode.workspace
          .getConfiguration('rslint', folder.uri)
          .update('enable', true, vscode.ConfigurationTarget.WorkspaceFolder);
      }
    }
  });

  test('keeps the last-good owner until a replacement commits config', async () => {
    const folder = vscode.workspace.workspaceFolders?.[0];
    assert.ok(folder);
    const seedDocument = await vscode.workspace.openTextDocument(
      vscode.Uri.file(path.join(folder.uri.fsPath, 'src/index.ts')),
    );
    const installed = await resolveRuntimeForDocument(seedDocument);
    assert.ok(installed);

    const fixture = path.join(
      folder.uri.fsPath,
      `.runtime-ready-gate-${String(process.pid)}`,
    );
    const rootMarker = path.join(fixture, 'root-runtime.log');
    const replacementMarker = path.join(fixture, 'replacement-runtime.log');
    const startGate = path.join(fixture, 'replacement-ready.gate');
    const rootCore = path.join(fixture, 'node_modules/@rslint/core');
    const project = path.join(fixture, 'packages/app');
    const replacementCore = path.join(project, 'node_modules/@rslint/core');
    const documentPath = path.join(project, 'src/index.ts');
    let document: vscode.TextDocument | undefined;
    let managerDisabled = false;

    try {
      await writeCoreWrapper(
        rootCore,
        rootMarker,
        installed.entryPath,
        '20.0.0-root',
      );
      await fs.mkdir(path.dirname(documentPath), { recursive: true });
      await fs.writeFile(documentPath, 'debugger;\n');
      await fs.writeFile(
        path.join(fixture, 'rslint.config.mjs'),
        `import fs from 'node:fs';\n` +
          `const gate = process.env.RSLINT_TEST_CONFIG_GATE;\n` +
          `if (gate) {\n` +
          `  fs.appendFileSync(process.env.RSLINT_TEST_CONFIG_MARKER, 'config-wait:' + String(process.pid) + '\\n');\n` +
          `  while (!fs.existsSync(gate)) await new Promise((resolve) => setTimeout(resolve, 25));\n` +
          `}\n` +
          `export default [{ files: ['**/*.ts'], rules: { 'no-debugger': 'error' } }];\n`,
      );
      document = await vscode.workspace.openTextDocument(
        vscode.Uri.file(documentPath),
      );
      await waitForMarker(
        rootMarker,
        (contents) => markerEvents(contents, 'start').length === 1,
        'the last-good root runtime did not start',
      );
      await waitForRslintDiagnostics(
        document,
        (diagnostics) => diagnostics.length > 0,
        30_000,
      );

      // Package managers can expose a replacement manifest before its entry
      // file is in place. That resolution error is not the same as a clean
      // uninstall: the current owner and its diagnostics must survive until a
      // later topology event completes the package.
      await fs.mkdir(replacementCore, { recursive: true });
      await fs.writeFile(
        path.join(replacementCore, 'package.json'),
        JSON.stringify({
          name: '@rslint/core',
          version: '21.0.0-incomplete',
          type: 'module',
          exports: {
            './editor-runtime': './editor-runtime.js',
            './package.json': './package.json',
          },
        }),
      );
      await fs.writeFile(
        path.join(project, 'package.json'),
        '{"name":"ready-gate","private":true,"phase":"incomplete"}\n',
      );
      await new Promise((resolve) => setTimeout(resolve, 1_000));
      assert.equal(
        markerEvents(await readMarker(rootMarker), 'stop').length,
        0,
        'a half-installed replacement must not release the last-good owner',
      );
      assert.ok(
        getRslintDiagnostics(document).length > 0,
        'a half-installed replacement must not clear last-good diagnostics',
      );

      await writeCoreWrapper(
        replacementCore,
        replacementMarker,
        installed.entryPath,
        '21.0.0-replacement',
        undefined,
        startGate,
      );
      await fs.writeFile(
        path.join(project, 'package.json'),
        '{"name":"ready-gate","private":true,"phase":"complete"}\n',
      );
      await waitForMarker(
        replacementMarker,
        (contents) => contents.includes('config-wait:'),
        'the replacement runtime did not reach its post-initialize config gate',
      );
      assert.equal(
        markerEvents(await readMarker(rootMarker), 'stop').length,
        0,
        'LSP initialize is not config-ready: the old owner must remain live at the gate',
      );
      assert.ok(
        getRslintDiagnostics(document).length > 0,
        'the last-good diagnostics must remain published while replacement initialize is pending',
      );

      await fs.writeFile(startGate, 'ready\n');
      await waitForMarker(
        rootMarker,
        (contents) => markerEvents(contents, 'stop').length === 1,
        'the old owner was not released after replacement config commit',
      );
    } finally {
      await fs.writeFile(startGate, 'cleanup\n').catch(() => undefined);
      await vscode.workspace
        .getConfiguration('rslint', folder.uri)
        .update('enable', false, vscode.ConfigurationTarget.WorkspaceFolder);
      managerDisabled = true;
      if (document) await closeDocuments([document]);
      if (await readMarker(replacementMarker)) {
        await waitForMarker(
          replacementMarker,
          (contents) => markerEvents(contents, 'stop').length >= 1,
          'the gated replacement did not stop during cleanup',
        );
      }
      await fs.rm(fixture, { recursive: true, force: true });
      if (managerDisabled) {
        await vscode.workspace
          .getConfiguration('rslint', folder.uri)
          .update('enable', true, vscode.ConfigurationTarget.WorkspaceFolder);
      }
    }
  });

  test('restarts a configured core when files outside the workspace change', async () => {
    const folder = vscode.workspace.workspaceFolders?.[0];
    assert.ok(folder);
    const document = await vscode.workspace.openTextDocument(
      vscode.Uri.file(path.join(folder.uri.fsPath, 'src/index.ts')),
    );
    const installed = await resolveRuntimeForDocument(document);
    assert.ok(installed);

    const externalRoot = await fs.mkdtemp(
      path.join(os.tmpdir(), 'rslint-vscode-external-core-'),
    );
    const packageDirectory = path.join(externalRoot, 'not-created-yet', 'core');
    const marker = path.join(externalRoot, 'runtime.log');
    let configured = false;
    try {
      await vscode.workspace
        .getConfiguration('rslint', folder.uri)
        .update(
          'runtime.path',
          packageDirectory,
          vscode.ConfigurationTarget.WorkspaceFolder,
        );
      configured = true;
      await new Promise((resolve) => setTimeout(resolve, 750));
      assert.equal(
        await readMarker(marker),
        '',
        'a missing configured package must not start a sidecar',
      );

      await fs.mkdir(packageDirectory, { recursive: true });
      await fs.writeFile(
        path.join(packageDirectory, 'package.json'),
        JSON.stringify({ name: '@rslint/core', version: 'incomplete' }),
      );
      await new Promise((resolve) => setTimeout(resolve, 750));
      assert.equal(
        await readMarker(marker),
        '',
        'an incomplete configured package must not start a sidecar',
      );

      await writeCoreWrapper(
        packageDirectory,
        marker,
        installed.entryPath,
        '30.0.0-external',
      );
      await waitForMarker(
        marker,
        (contents) => markerEvents(contents, 'start').length === 1,
        'the configured external runtime did not start',
      );

      await fs.writeFile(
        path.join(packageDirectory, 'runtime-support.js'),
        `export const runtimeSupportGeneration = '30.0.1-external-payload-only';\n`,
      );
      await waitForMarker(
        marker,
        (contents) =>
          markerEvents(contents, 'start').length >= 2 &&
          markerEvents(contents, 'stop').length >= 1,
        'an external core payload-only update did not replace the running generation',
      );
    } finally {
      if (configured) {
        await vscode.workspace
          .getConfiguration('rslint', folder.uri)
          .update(
            'runtime.path',
            '',
            vscode.ConfigurationTarget.WorkspaceFolder,
          );
      }
      if (await readMarker(marker)) {
        await waitForMarker(
          marker,
          (contents) => markerEvents(contents, 'stop').length >= 2,
          'the external runtime was not released after clearing its setting',
          45_000,
        );
      }
      await fs.rm(externalRoot, { recursive: true, force: true });
    }
  });
});

import * as assert from 'assert';
import * as vscode from 'vscode';
import path from 'node:path';
import fs from 'node:fs';
import { selectUnavailableConfigBoundaryDirectories } from '../../src/Rslint';

suite('rslint JS config support', function () {
  this.timeout(120_000);

  function getWorkspaceRoot(): string {
    return vscode.workspace.workspaceFolders![0].uri.fsPath;
  }

  /**
   * Race-free `onDidChangeDiagnostics` waiter.
   *
   * Compared to a polling `getDiagnostics` + `setTimeout` loop, this
   * subscribes BEFORE checking the initial state, eliminating the
   * subscribe-after-check window where an event fired between the
   * synchronous read and the listener registration would be lost.
   *
   * The waiter returns as soon as `predicate` accepts a snapshot — no
   * file edits, no nudges, no sleeps. It relies on the server's own
   * push channel: `internal/lsp/service.go::handleConfigUpdate` ends
   * with `s.RefreshDiagnostics(ctx)`, which signals `refreshCh`; the
   * dispatch loop in `internal/lsp/server.go` consumes that signal and
   * calls `pushDiagnostics(uri)` for every open document. So the LSP
   * server already publishes new diagnostics after a config reload —
   * tests just need to listen for them.
   *
   * Default 60s timeout (vs the older helper's 30s × 20 polling
   * budget) gives headroom for slow CI runners where the JS-config
   * reload chain (ESM dynamic import + IO) can take 10s+.
   */
  function waitForDiagnostics(
    doc: vscode.TextDocument,
    predicate?: (diags: vscode.Diagnostic[]) => boolean,
    timeoutMs = 60_000,
  ): Promise<vscode.Diagnostic[]> {
    const matches = predicate ?? ((ds) => ds.length > 0);
    return new Promise<vscode.Diagnostic[]>((resolve, reject) => {
      let done = false;
      const finish = (
        value: vscode.Diagnostic[] | undefined,
        err?: Error,
      ): void => {
        if (done) return;
        done = true;
        sub.dispose();
        clearTimeout(timer);
        if (err) reject(err);
        else resolve(value!);
      };
      const sub = vscode.languages.onDidChangeDiagnostics((e) => {
        if (!e.uris.some((u) => u.toString() === doc.uri.toString())) return;
        const ds = vscode.languages.getDiagnostics(doc.uri);
        if (matches(ds)) finish(ds);
      });
      const timer = setTimeout(
        () =>
          finish(
            undefined,
            new Error('waitForDiagnostics: timeout waiting for predicate'),
          ),
        timeoutMs,
      );
      // Check current state in case the matching publish already
      // happened before subscription. Safe because finish() is
      // idempotent.
      const initial = vscode.languages.getDiagnostics(doc.uri);
      if (matches(initial)) finish(initial);
    });
  }

  async function openFixture(filename: string): Promise<vscode.TextDocument> {
    const filePath = path.join(getWorkspaceRoot(), 'src', filename);
    return vscode.workspace.openTextDocument(filePath);
  }

  async function triggerRelint(doc: vscode.TextDocument): Promise<void> {
    const editor = await vscode.window.showTextDocument(doc);
    await editor.edit((edit) => edit.insert(new vscode.Position(0, 0), ' '));
    await editor.edit((edit) => edit.delete(new vscode.Range(0, 0, 0, 1)));
  }

  test('unavailable config boundaries preserve JS ownership without hiding valid ancestors', () => {
    const root = path.resolve('/workspace');
    const app = path.join(root, 'packages', 'app');
    const broken = path.join(root, 'packages', 'broken');

    assert.deepStrictEqual(
      selectUnavailableConfigBoundaryDirectories([app], [root]),
      [root],
      'a valid descendant must not expose JSON through an unavailable root',
    );
    assert.deepStrictEqual(
      selectUnavailableConfigBoundaryDirectories([root], [broken]),
      [],
      'an unavailable child must fall back to its usable JS ancestor',
    );
    assert.deepStrictEqual(
      selectUnavailableConfigBoundaryDirectories([], [broken]),
      [broken],
      'a lone unavailable nested config protects only its own subtree',
    );
    assert.deepStrictEqual(
      selectUnavailableConfigBoundaryDirectories([], [root, broken]),
      [root],
      'an outer empty boundary makes a nested duplicate unnecessary',
    );
  });

  test('JS config should produce diagnostics', async () => {
    const doc = await openFixture('index.ts');
    await vscode.window.showTextDocument(doc);

    // Wait specifically for JS config diagnostics. JSON config may load first
    // (server-side) before JS config arrives via rslint/configUpdate, so we
    // must use a predicate to avoid returning early with JSON config results.
    const diagnostics = await waitForDiagnostics(doc, (diags) =>
      diags.some((d) => d.message.includes('no-unsafe-member-access')),
    );
    assert.ok(
      diagnostics.length > 0,
      `Expected diagnostics but got ${diagnostics.length}`,
    );
    assert.ok(
      diagnostics.some((d) => d.message.includes('no-unsafe-member-access')),
      'Expected no-unsafe-member-access diagnostic from JS config',
    );
  });

  test('JS config should take priority over JSON config', async () => {
    const doc = await openFixture('index.ts');
    await vscode.window.showTextDocument(doc);

    // Wait specifically for JS config diagnostics (no-unsafe-member-access).
    // JSON config may load first with no-explicit-any, but JS config should
    // override via rslint/configUpdate notification.
    const diagnostics = await waitForDiagnostics(doc, (diags) =>
      diags.some((d) => d.message.includes('no-unsafe-member-access')),
    );

    assert.ok(
      diagnostics.some((d) => d.message.includes('no-unsafe-member-access')),
      'Expected no-unsafe-member-access from JS config',
    );
    assert.ok(
      !diagnostics.some((d) => d.message.includes('no-explicit-any')),
      'Should NOT see no-explicit-any because JS config takes priority over JSON',
    );
  });

  test('config hot reload should update diagnostics', async () => {
    const doc = await openFixture('index.ts');
    await vscode.window.showTextDocument(doc);

    // 1. Verify initial diagnostics have no-unsafe-member-access.
    await waitForDiagnostics(doc, (diags) =>
      diags.some((d) => d.message.includes('no-unsafe-member-access')),
    );

    // 2. Subscribe BEFORE writing the new config — eliminates the
    //    "publish fires between write and subscribe" race window. The
    //    waiter listens; the server pushes after handleConfigUpdate
    //    via RefreshDiagnostics; we resolve.
    const configPath = path.join(getWorkspaceRoot(), 'rslint.config.js');
    const originalConfig = fs.readFileSync(configPath, 'utf8');
    const newConfig = `export default [
  {
    files: ['**/*.ts'],
    languageOptions: {
      parserOptions: {
        projectService: false,
        project: ['./tsconfig.json'],
      },
    },
    rules: {
      '@typescript-eslint/no-explicit-any': 'error',
      '@typescript-eslint/no-unsafe-member-access': 'off',
    },
    plugins: ['@typescript-eslint'],
  },
];
`;
    const reloaded = waitForDiagnostics(
      doc,
      (diags) =>
        diags.some((d) => d.message.includes('no-explicit-any')) &&
        !diags.some((d) => d.message.includes('no-unsafe-member-access')),
    );

    try {
      fs.writeFileSync(configPath, newConfig, 'utf8');
      const updatedDiags = await reloaded;
      assert.ok(
        updatedDiags.some((d) => d.message.includes('no-explicit-any')),
        'After hot reload, diagnostics should include no-explicit-any',
      );
      assert.ok(
        !updatedDiags.some((d) =>
          d.message.includes('no-unsafe-member-access'),
        ),
        'After hot reload, no-unsafe-member-access should be gone',
      );
    } finally {
      fs.writeFileSync(configPath, originalConfig, 'utf8');
    }
  });

  test('CJS and CTS configs support create and hot reload', async () => {
    const root = getWorkspaceRoot();
    const jsConfigPath = path.join(root, 'rslint.config.js');
    const originalJSConfig = fs.readFileSync(jsConfigPath, 'utf8');
    const probePath = path.join(root, 'src', 'config-extension-probe.ts');
    const configSource = (ruleName: string): string => `module.exports = [{
  files: ['**/*.ts'],
  rules: { ${JSON.stringify(ruleName)}: 'error' },
}];
`;

    fs.writeFileSync(probePath, 'debugger;\nconsole.log("probe");\n', 'utf8');
    const doc = await vscode.workspace.openTextDocument(probePath);
    await vscode.window.showTextDocument(doc);
    fs.unlinkSync(jsConfigPath);

    try {
      for (const extension of ['cjs', 'cts']) {
        const configPath = path.join(root, `rslint.config.${extension}`);
        const debuggerDiagnostics = waitForDiagnostics(doc, (diagnostics) =>
          diagnostics.some((diagnostic) =>
            diagnostic.message.includes("Unexpected 'debugger' statement"),
          ),
        );
        fs.writeFileSync(configPath, configSource('no-debugger'), 'utf8');
        await debuggerDiagnostics;

        const consoleDiagnostics = waitForDiagnostics(
          doc,
          (diagnostics) =>
            diagnostics.some((diagnostic) =>
              diagnostic.message.includes('Unexpected console statement'),
            ) &&
            !diagnostics.some((diagnostic) =>
              diagnostic.message.includes("Unexpected 'debugger' statement"),
            ),
        );
        fs.writeFileSync(configPath, configSource('no-console'), 'utf8');
        await consoleDiagnostics;
        fs.unlinkSync(configPath);
      }
    } finally {
      for (const extension of ['cjs', 'cts']) {
        fs.rmSync(path.join(root, `rslint.config.${extension}`), {
          force: true,
        });
      }
      fs.writeFileSync(jsConfigPath, originalJSConfig, 'utf8');
      fs.rmSync(probePath, { force: true });
    }
  });

  test('deleting JS config should clear diagnostics', async () => {
    const doc = await openFixture('index.ts');
    await vscode.window.showTextDocument(doc);

    await waitForDiagnostics(doc, (diags) =>
      diags.some((d) => d.message.includes('no-unsafe-member-access')),
    );

    const configPath = path.join(getWorkspaceRoot(), 'rslint.config.js');
    const originalConfig = fs.readFileSync(configPath, 'utf8');

    const cleared = waitForDiagnostics(
      doc,
      (diags) =>
        !diags.some((d) => d.message.includes('no-unsafe-member-access')),
    );

    try {
      fs.unlinkSync(configPath);
      const afterDeleteDiags = await cleared;
      assert.ok(
        !afterDeleteDiags.some((d) =>
          d.message.includes('no-unsafe-member-access'),
        ),
        'After deleting JS config, no-unsafe-member-access should be gone',
      );
    } finally {
      // Restore config and wait for the reload to settle so the next
      // test starts from a known state. Subscribe first, then write.
      const restored = waitForDiagnostics(doc, (diags) =>
        diags.some((d) => d.message.includes('no-unsafe-member-access')),
      ).catch(() => undefined);
      fs.writeFileSync(configPath, originalConfig, 'utf8');
      await restored;
    }
  });

  test('creating a new JS config should load it and produce diagnostics', async () => {
    const configPath = path.join(getWorkspaceRoot(), 'rslint.config.js');
    const originalConfig = fs.readFileSync(configPath, 'utf8');
    const doc = await openFixture('index.ts');
    await vscode.window.showTextDocument(doc);

    // Step 1: delete existing config and wait until its diagnostics drop.
    const cleared = waitForDiagnostics(doc, (diags) =>
      diags.every((d) => !d.message.includes('no-unsafe-member-access')),
    );
    fs.unlinkSync(configPath);
    await cleared;

    try {
      // Step 2: subscribe for the new rule, then create the file.
      const newConfig = `export default [
  {
    files: ['**/*.ts'],
    languageOptions: {
      parserOptions: {
        projectService: false,
        project: ['./tsconfig.json'],
      },
    },
    rules: {
      '@typescript-eslint/no-explicit-any': 'error',
    },
    plugins: ['@typescript-eslint'],
  },
];
`;
      const created = waitForDiagnostics(doc, (diags) =>
        diags.some((d) => d.message.includes('no-explicit-any')),
      );
      fs.writeFileSync(configPath, newConfig, 'utf8');
      const afterCreateDiags = await created;
      assert.ok(
        afterCreateDiags.some((d) => d.message.includes('no-explicit-any')),
        'After creating new JS config, should see no-explicit-any diagnostic',
      );
    } finally {
      const restored = waitForDiagnostics(doc, (diags) =>
        diags.some((d) => d.message.includes('no-unsafe-member-access')),
      ).catch(() => undefined);
      fs.writeFileSync(configPath, originalConfig, 'utf8');
      await restored;
    }
  });

  test('a newly discovered broken nested config keeps valid ancestor config active', async () => {
    const root = getWorkspaceRoot();
    const rootConfigPath = path.join(root, 'rslint.config.js');
    const originalRootConfig = fs.readFileSync(rootConfigPath, 'utf8');
    const nestedDir = path.join(root, 'broken-nested-config');
    const nestedFilePath = path.join(nestedDir, 'index.ts');
    const nestedConfigPath = path.join(nestedDir, 'rslint.config.js');
    const rootDoc = await openFixture('index.ts');

    const rootConfigWithMarker = `export default [{
  files: ['**/*.ts'],
  languageOptions: {
    parserOptions: { projectService: false, project: ['./tsconfig.json'] },
  },
  rules: {
    '@typescript-eslint/no-unsafe-member-access': 'error',
    'no-debugger': 'error',
  },
  plugins: ['@typescript-eslint'],
}];
`;

    try {
      fs.mkdirSync(nestedDir, { recursive: true });
      fs.writeFileSync(nestedFilePath, 'debugger;\n', 'utf8');
      fs.writeFileSync(rootConfigPath, rootConfigWithMarker, 'utf8');

      await vscode.window.showTextDocument(rootDoc);
      const nestedDoc = await vscode.workspace.openTextDocument(nestedFilePath);
      await vscode.window.showTextDocument(nestedDoc);
      await waitForDiagnostics(nestedDoc, (diags) =>
        diags.some((d) => d.message.includes('no-debugger')),
      );

      fs.writeFileSync(
        nestedConfigPath,
        'export default [BROKEN SYNTAX;',
        'utf8',
      );
      await new Promise((resolve) => setTimeout(resolve, 1500));
      await triggerRelint(nestedDoc);

      const nestedDiagnostics = await waitForDiagnostics(nestedDoc, (diags) =>
        diags.some((d) => d.message.includes('no-debugger')),
      );
      assert.ok(
        nestedDiagnostics.some((d) => d.message.includes('no-debugger')),
        'The broken nested config should be skipped so the root config remains active',
      );

      const rootDiagnostics = await waitForDiagnostics(rootDoc, (diags) =>
        diags.some((d) => d.message.includes('no-unsafe-member-access')),
      );
      assert.ok(
        rootDiagnostics.some((d) =>
          d.message.includes('no-unsafe-member-access'),
        ),
        'The valid root config must remain active outside the broken nested config',
      );
    } finally {
      const rootRestored = waitForDiagnostics(rootDoc, (diags) =>
        diags.some((d) => d.message.includes('no-unsafe-member-access')),
      ).catch(() => undefined);
      fs.writeFileSync(rootConfigPath, originalRootConfig, 'utf8');
      fs.rmSync(nestedDir, { recursive: true, force: true });
      await rootRestored;
    }
  });

  test('parent global ignores remove nested configs from the effective catalog', async () => {
    const root = getWorkspaceRoot();
    const rootConfigPath = path.join(root, 'rslint.config.js');
    const originalRootConfig = fs.readFileSync(rootConfigPath, 'utf8');
    const nestedDir = path.join(root, 'parent-ignore-catalog-probe');
    const nestedConfigPath = path.join(nestedDir, 'rslint.config.mjs');
    const nestedFilePath = path.join(nestedDir, 'index.ts');
    const loadMarkerPath = path.join(nestedDir, 'config-loads.txt');
    const rootDoc = await openFixture('index.ts');

    const ignoredRootConfig = originalRootConfig
      .replace(
        'export default [',
        "export default [{ ignores: ['parent-ignore-catalog-probe/**'] },",
      )
      .replace(
        "'@typescript-eslint/no-explicit-any': 'off'",
        "'@typescript-eslint/no-explicit-any': 'warn'",
      );

    try {
      fs.mkdirSync(nestedDir, { recursive: true });
      fs.writeFileSync(nestedFilePath, 'console.log("nested");\n', 'utf8');
      fs.writeFileSync(
        nestedConfigPath,
        `import fs from 'node:fs';
fs.appendFileSync(${JSON.stringify(loadMarkerPath)}, 'x');
export default [{ files: ['**/*.ts'], rules: { 'no-console': 'error' } }];
`,
        'utf8',
      );

      const nestedDoc = await vscode.workspace.openTextDocument(nestedFilePath);
      await vscode.window.showTextDocument(nestedDoc);
      await waitForDiagnostics(nestedDoc, (diagnostics) =>
        diagnostics.some((diagnostic) =>
          diagnostic.message.includes('Unexpected console statement'),
        ),
      );

      await vscode.window.showTextDocument(rootDoc);
      const parentApplied = waitForDiagnostics(rootDoc, (diagnostics) =>
        diagnostics.some((diagnostic) =>
          diagnostic.message.includes('no-explicit-any'),
        ),
      );
      const nestedCleared = waitForDiagnostics(
        nestedDoc,
        (diagnostics) =>
          !diagnostics.some((diagnostic) =>
            diagnostic.message.includes('Unexpected console statement'),
          ),
      );

      fs.writeFileSync(rootConfigPath, ignoredRootConfig, 'utf8');
      await Promise.all([parentApplied, nestedCleared]);

      assert.ok(
        fs.readFileSync(loadMarkerPath, 'utf8').length >= 2,
        'The ignored nested candidate should still be scanned and loaded',
      );
      assert.ok(
        !vscode.languages
          .getDiagnostics(nestedDoc.uri)
          .some((diagnostic) =>
            diagnostic.message.includes('Unexpected console statement'),
          ),
        'The ignored nested config must not be sent in the effective catalog',
      );
    } finally {
      const restored = waitForDiagnostics(
        rootDoc,
        (diagnostics) =>
          diagnostics.some((diagnostic) =>
            diagnostic.message.includes('no-unsafe-member-access'),
          ) &&
          !diagnostics.some((diagnostic) =>
            diagnostic.message.includes('no-explicit-any'),
          ),
      ).catch(() => undefined);
      fs.rmSync(nestedDir, { recursive: true, force: true });
      fs.writeFileSync(rootConfigPath, originalRootConfig, 'utf8');
      await restored;
    }
  });

  test('config search excludes node_modules and .git', async () => {
    const root = getWorkspaceRoot();
    const rootConfigPath = path.join(root, 'rslint.config.js');
    const originalRootConfig = fs.readFileSync(rootConfigPath, 'utf8');
    const rootDoc = await openFixture('index.ts');
    const gitDirectory = path.join(root, '.git');
    const gitDirectoryExisted = fs.existsSync(gitDirectory);
    const probes = [
      path.join(root, 'node_modules', 'rslint-config-search-probe'),
      path.join(gitDirectory, 'rslint-config-search-probe'),
    ];
    const markerPaths = probes.map((probe) => path.join(probe, 'loaded.txt'));
    const changedRootConfig = originalRootConfig.replace(
      "'@typescript-eslint/no-explicit-any': 'off'",
      "'@typescript-eslint/no-explicit-any': 'warn'",
    );

    try {
      for (let index = 0; index < probes.length; index++) {
        fs.mkdirSync(probes[index], { recursive: true });
        fs.writeFileSync(
          path.join(probes[index], 'rslint.config.mjs'),
          `import fs from 'node:fs'; fs.writeFileSync(${JSON.stringify(markerPaths[index])}, 'loaded'); export default [];`,
          'utf8',
        );
      }

      await vscode.window.showTextDocument(rootDoc);
      const reloaded = waitForDiagnostics(rootDoc, (diagnostics) =>
        diagnostics.some((diagnostic) =>
          diagnostic.message.includes('no-explicit-any'),
        ),
      );
      fs.writeFileSync(rootConfigPath, changedRootConfig, 'utf8');
      await reloaded;

      for (const markerPath of markerPaths) {
        assert.ok(
          !fs.existsSync(markerPath),
          `Excluded config was unexpectedly loaded: ${markerPath}`,
        );
      }
    } finally {
      const restored = waitForDiagnostics(
        rootDoc,
        (diagnostics) =>
          diagnostics.some((diagnostic) =>
            diagnostic.message.includes('no-unsafe-member-access'),
          ) &&
          !diagnostics.some((diagnostic) =>
            diagnostic.message.includes('no-explicit-any'),
          ),
      ).catch(() => undefined);
      for (const probe of probes) {
        fs.rmSync(probe, { recursive: true, force: true });
      }
      if (!gitDirectoryExisted) {
        try {
          fs.rmdirSync(gitDirectory);
        } catch {
          // Leave a concurrently populated directory intact.
        }
      }
      fs.writeFileSync(rootConfigPath, originalRootConfig, 'utf8');
      await restored;
    }
  });

  test('same-directory configs use .js > .mjs > .cjs > .ts > .mts > .cts priority', async () => {
    const root = getWorkspaceRoot();
    const jsPath = path.join(root, 'rslint.config.js');
    const mjsPath = path.join(root, 'rslint.config.mjs');
    const cjsPath = path.join(root, 'rslint.config.cjs');
    const tsPath = path.join(root, 'rslint.config.ts');
    const mtsPath = path.join(root, 'rslint.config.mts');
    const ctsPath = path.join(root, 'rslint.config.cts');
    const originalJS = fs.readFileSync(jsPath, 'utf8');
    const doc = await openFixture('index.ts');
    await vscode.window.showTextDocument(doc);

    const configFor = (
      enabledRule:
        | '@typescript-eslint/no-explicit-any'
        | '@typescript-eslint/no-unsafe-member-access',
      commonJS = false,
    ): string => `${commonJS ? 'module.exports =' : 'export default'} [
  {
    files: ['**/*.ts'],
    languageOptions: {
      parserOptions: {
        projectService: false,
        project: ['./tsconfig.json'],
      },
    },
    rules: {
      '@typescript-eslint/no-explicit-any': '${enabledRule.endsWith('no-explicit-any') ? 'warn' : 'off'}',
      '@typescript-eslint/no-unsafe-member-access': '${enabledRule.endsWith('no-unsafe-member-access') ? 'warn' : 'off'}',
    },
    plugins: ['@typescript-eslint'],
  },
];
`;

    const expectWarning = async (ruleName: string): Promise<void> => {
      const diagnostics = await waitForDiagnostics(doc, (diags) =>
        diags.some(
          (d) =>
            d.message.includes(ruleName) &&
            d.severity === vscode.DiagnosticSeverity.Warning,
        ),
      );
      const diagnostic = diagnostics.find((d) => d.message.includes(ruleName));
      assert.strictEqual(
        diagnostic?.severity,
        vscode.DiagnosticSeverity.Warning,
        `Expected ${ruleName} from the selected config`,
      );
    };

    try {
      fs.writeFileSync(
        mjsPath,
        configFor('@typescript-eslint/no-explicit-any'),
        'utf8',
      );
      fs.writeFileSync(
        tsPath,
        configFor('@typescript-eslint/no-explicit-any'),
        'utf8',
      );
      fs.writeFileSync(
        mtsPath,
        configFor('@typescript-eslint/no-unsafe-member-access'),
        'utf8',
      );
      fs.writeFileSync(
        cjsPath,
        configFor('@typescript-eslint/no-unsafe-member-access', true),
        'utf8',
      );
      fs.writeFileSync(
        ctsPath,
        configFor('@typescript-eslint/no-explicit-any', true),
        'utf8',
      );
      await new Promise((resolve) => setTimeout(resolve, 1500));

      let diagnostics = vscode.languages.getDiagnostics(doc.uri);
      assert.ok(
        diagnostics.some(
          (d) =>
            d.message.includes('no-unsafe-member-access') &&
            d.severity === vscode.DiagnosticSeverity.Error,
        ),
        '.js should remain selected while lower-priority configs coexist',
      );

      fs.unlinkSync(jsPath);
      await expectWarning('no-explicit-any');

      fs.unlinkSync(mjsPath);
      await expectWarning('no-unsafe-member-access');

      fs.unlinkSync(cjsPath);
      await expectWarning('no-explicit-any');

      fs.unlinkSync(tsPath);
      await expectWarning('no-unsafe-member-access');

      fs.unlinkSync(mtsPath);
      await expectWarning('no-explicit-any');

      diagnostics = vscode.languages.getDiagnostics(doc.uri);
      assert.ok(
        !diagnostics.some((d) => d.message.includes('no-unsafe-member-access')),
        '.cts should govern after all higher-priority variants are removed',
      );
    } finally {
      const restored = waitForDiagnostics(doc, (diags) =>
        diags.some(
          (d) =>
            d.message.includes('no-unsafe-member-access') &&
            d.severity === vscode.DiagnosticSeverity.Error,
        ),
      ).catch(() => undefined);
      fs.writeFileSync(jsPath, originalJS, 'utf8');
      fs.rmSync(mjsPath, { force: true });
      fs.rmSync(cjsPath, { force: true });
      fs.rmSync(tsPath, { force: true });
      fs.rmSync(mtsPath, { force: true });
      fs.rmSync(ctsPath, { force: true });
      await restored;
    }
  });

  test('broken higher-priority config preserves last-good and does not load a lower variant', async () => {
    const root = getWorkspaceRoot();
    const jsPath = path.join(root, 'rslint.config.js');
    const mjsPath = path.join(root, 'rslint.config.mjs');
    const originalJS = fs.readFileSync(jsPath, 'utf8');
    const doc = await openFixture('index.ts');
    await vscode.window.showTextDocument(doc);
    await waitForDiagnostics(doc, (diags) =>
      diags.some((d) => d.message.includes('no-unsafe-member-access')),
    );

    const lowerPriorityConfig = `export default [{
  languageOptions: { parserOptions: { project: ['./tsconfig.json'] } },
  rules: {
    '@typescript-eslint/no-explicit-any': 'warn',
    '@typescript-eslint/no-unsafe-member-access': 'off',
  },
  plugins: ['@typescript-eslint'],
}];
`;

    try {
      fs.writeFileSync(mjsPath, lowerPriorityConfig, 'utf8');
      fs.writeFileSync(jsPath, 'export default [BROKEN SYNTAX;', 'utf8');
      await new Promise((resolve) => setTimeout(resolve, 2000));
      await triggerRelint(doc);

      const diagnostics = await waitForDiagnostics(doc, (diags) =>
        diags.some((d) => d.message.includes('no-unsafe-member-access')),
      );
      assert.ok(
        diagnostics.some((d) => d.message.includes('no-unsafe-member-access')),
        'The last-good .js config should remain active',
      );
      assert.ok(
        !diagnostics.some((d) => d.message.includes('no-explicit-any')),
        'A broken .js must not fall through to .mjs or JSON',
      );
    } finally {
      const restored = waitForDiagnostics(doc, (diags) =>
        diags.some((d) => d.message.includes('no-unsafe-member-access')),
      ).catch(() => undefined);
      fs.writeFileSync(jsPath, originalJS, 'utf8');
      fs.rmSync(mjsPath, { force: true });
      await restored;
    }
  });
});

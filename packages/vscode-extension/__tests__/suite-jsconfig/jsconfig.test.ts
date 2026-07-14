import * as assert from 'assert';
import * as vscode from 'vscode';
import path from 'node:path';
import fs from 'node:fs';

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
   * push channel: a committed config-discovery transaction refreshes open
   * document diagnostics, so tests only need to listen for that publication.
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

  test('JS config should produce diagnostics', async () => {
    const doc = await openFixture('index.ts');
    await vscode.window.showTextDocument(doc);

    // Wait specifically for JS config diagnostics. The startup snapshot may
    // publish JSON fallback results before JS config activation commits.
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
    // JSON config may load first with no-explicit-any, but the committed JS
    // config catalog should override it.
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
    //    waiter listens; the server pushes after committing the refresh.
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

  test('one workspace config edit evaluates exactly one transaction', async () => {
    const root = getWorkspaceRoot();
    const configPath = path.join(root, 'rslint.config.js');
    const markerPath = path.join(root, '.rslint-config-refresh-count');
    const originalConfig = fs.readFileSync(configPath, 'utf8');
    const doc = await openFixture('index.ts');
    await vscode.window.showTextDocument(doc);
    await waitForDiagnostics(doc, (diags) =>
      diags.some((d) => d.message.includes('no-unsafe-member-access')),
    );

    const countedConfig = `import fs from 'node:fs';
fs.appendFileSync(${JSON.stringify(markerPath)}, 'x');
export default [{
  files: ['**/*.ts'],
  languageOptions: {
    parserOptions: { projectService: false, project: ['./tsconfig.json'] },
  },
  rules: {
    '@typescript-eslint/no-explicit-any': 'error',
    '@typescript-eslint/no-unsafe-member-access': 'off',
  },
  plugins: ['@typescript-eslint'],
}];
`;

    const reloaded = waitForDiagnostics(
      doc,
      (diags) =>
        diags.some((d) => d.message.includes('no-explicit-any')) &&
        !diags.some((d) => d.message.includes('no-unsafe-member-access')),
    );

    try {
      fs.rmSync(markerPath, { force: true });
      fs.writeFileSync(configPath, countedConfig, 'utf8');
      await reloaded;
      // A duplicate didChangeWatchedFiles transaction used to race the direct
      // watcher. Let any queued 300ms debounce complete before inspecting
      // the config module's observable evaluation count.
      await new Promise((resolve) => setTimeout(resolve, 1500));
      assert.strictEqual(
        fs.readFileSync(markerPath, 'utf8'),
        'x',
        'one edit must evaluate one config-discovery transaction',
      );
    } finally {
      const restored = waitForDiagnostics(doc, (diags) =>
        diags.some((d) => d.message.includes('no-unsafe-member-access')),
      ).catch(() => undefined);
      fs.writeFileSync(configPath, originalConfig, 'utf8');
      await restored;
      fs.rmSync(markerPath, { force: true });
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

      assert.strictEqual(
        fs.readFileSync(loadMarkerPath, 'utf8'),
        'x',
        'The ignored nested candidate must not be evaluated again',
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

  test('same-directory configs use .js > .mjs > .ts > .mts priority', async () => {
    const root = getWorkspaceRoot();
    const jsPath = path.join(root, 'rslint.config.js');
    const mjsPath = path.join(root, 'rslint.config.mjs');
    const tsPath = path.join(root, 'rslint.config.ts');
    const mtsPath = path.join(root, 'rslint.config.mts');
    const originalJS = fs.readFileSync(jsPath, 'utf8');
    const doc = await openFixture('index.ts');
    await vscode.window.showTextDocument(doc);

    const configFor = (
      enabledRule:
        | '@typescript-eslint/no-explicit-any'
        | '@typescript-eslint/no-unsafe-member-access',
    ): string => `export default [
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
        configFor('@typescript-eslint/no-unsafe-member-access'),
        'utf8',
      );
      fs.writeFileSync(
        mtsPath,
        configFor('@typescript-eslint/no-explicit-any'),
        'utf8',
      );
      await new Promise((resolve) => setTimeout(resolve, 1500));

      const diagnostics = vscode.languages.getDiagnostics(doc.uri);
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

      fs.unlinkSync(tsPath);
      await expectWarning('no-explicit-any');
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
      fs.rmSync(tsPath, { force: true });
      fs.rmSync(mtsPath, { force: true });
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

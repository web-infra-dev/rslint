import * as assert from 'assert';
import * as vscode from 'vscode';
import path from 'node:path';
import fs from 'node:fs';
import { waitForRslintDiagnostics as waitForDiagnostics } from '../utils/diagnostics';
import { closeTextEditor, revertTextDocument } from '../utils/documents';

suite('rslint JS config support', function () {
  this.timeout(120_000);

  function getWorkspaceRoot(): string {
    return vscode.workspace.workspaceFolders![0].uri.fsPath;
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

  async function waitForFile(filePath: string, timeoutMs = 10_000) {
    const deadline = Date.now() + timeoutMs;
    while (Date.now() < deadline) {
      if (fs.existsSync(filePath)) return;
      await new Promise((resolve) => setTimeout(resolve, 25));
    }
    throw new Error(`Timed out waiting for file: ${filePath}`);
  }

  async function withFailClosedCleanup(
    testFn: () => Promise<void>,
    cleanupFn: () => Promise<void>,
    description: string,
  ): Promise<void> {
    let testError: unknown;
    try {
      await testFn();
    } catch (error) {
      testError = error;
    }

    let cleanupError: unknown;
    try {
      await cleanupFn();
    } catch (error) {
      cleanupError = error;
    }

    if (testError && cleanupError) {
      throw new AggregateError(
        [testError, cleanupError],
        `${description}: test and cleanup both failed`,
      );
    }
    if (testError) throw testError;
    if (cleanupError) throw cleanupError;
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
    await withFailClosedCleanup(
      async () => {
        const reloaded = waitForDiagnostics(
          doc,
          (diags) =>
            diags.some((d) => d.message.includes('no-explicit-any')) &&
            !diags.some((d) => d.message.includes('no-unsafe-member-access')),
        );
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
      },
      async () => {
        const restored = waitForDiagnostics(doc, (diags) =>
          diags.some((d) => d.message.includes('no-unsafe-member-access')),
        );
        fs.writeFileSync(configPath, originalConfig, 'utf8');
        await restored;
      },
      'Config hot-reload test',
    );
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

    await withFailClosedCleanup(
      async () => {
        const reloaded = waitForDiagnostics(
          doc,
          (diags) =>
            diags.some((d) => d.message.includes('no-explicit-any')) &&
            !diags.some((d) => d.message.includes('no-unsafe-member-access')),
        );
        fs.rmSync(markerPath, { force: true });
        fs.writeFileSync(configPath, countedConfig, 'utf8');
        await reloaded;

        // A duplicate didChangeWatchedFiles transaction used to race the
        // direct watcher. Let any queued 300ms debounce finish before reading
        // the config module's observable evaluation count.
        await new Promise((resolve) => setTimeout(resolve, 1500));
        assert.strictEqual(
          fs.readFileSync(markerPath, 'utf8'),
          'x',
          'one edit must evaluate one config-discovery transaction',
        );
      },
      async () => {
        await withFailClosedCleanup(
          async () => {
            const restored = waitForDiagnostics(doc, (diags) =>
              diags.some((d) => d.message.includes('no-unsafe-member-access')),
            );
            fs.writeFileSync(configPath, originalConfig, 'utf8');
            await restored;
          },
          async () => fs.rmSync(markerPath, { force: true }),
          'Config-transaction cleanup',
        );
      },
      'Config-transaction test',
    );
  });

  test('deleting JS config should clear diagnostics', async () => {
    const doc = await openFixture('index.ts');
    await vscode.window.showTextDocument(doc);

    await waitForDiagnostics(doc, (diags) =>
      diags.some((d) => d.message.includes('no-unsafe-member-access')),
    );

    const configPath = path.join(getWorkspaceRoot(), 'rslint.config.js');
    const originalConfig = fs.readFileSync(configPath, 'utf8');

    await withFailClosedCleanup(
      async () => {
        const cleared = waitForDiagnostics(
          doc,
          (diags) =>
            !diags.some((d) => d.message.includes('no-unsafe-member-access')),
        );
        fs.unlinkSync(configPath);
        const afterDeleteDiags = await cleared;
        assert.ok(
          !afterDeleteDiags.some((d) =>
            d.message.includes('no-unsafe-member-access'),
          ),
          'After deleting JS config, no-unsafe-member-access should be gone',
        );
      },
      async () => {
        // Subscribe before restoring so the next test starts from a known state.
        const restored = waitForDiagnostics(doc, (diags) =>
          diags.some((d) => d.message.includes('no-unsafe-member-access')),
        );
        fs.writeFileSync(configPath, originalConfig, 'utf8');
        await restored;
      },
      'JS-config deletion test',
    );
  });

  test('creating a new JS config should load it and produce diagnostics', async () => {
    const configPath = path.join(getWorkspaceRoot(), 'rslint.config.js');
    const originalConfig = fs.readFileSync(configPath, 'utf8');
    const doc = await openFixture('index.ts');
    await vscode.window.showTextDocument(doc);

    // Establish a positive publication first, so clearing cannot pass on the
    // document's not-yet-linted initial empty snapshot.
    await waitForDiagnostics(doc, (diags) =>
      diags.some((d) => d.message.includes('no-unsafe-member-access')),
    );

    await withFailClosedCleanup(
      async () => {
        // Step 1: delete existing config and observe its diagnostics drop.
        const cleared = waitForDiagnostics(doc, (diags) =>
          diags.every((d) => !d.message.includes('no-unsafe-member-access')),
        );
        fs.unlinkSync(configPath);
        await cleared;

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
      },
      async () => {
        const restored = waitForDiagnostics(doc, (diags) =>
          diags.some((d) => d.message.includes('no-unsafe-member-access')),
        );
        fs.writeFileSync(configPath, originalConfig, 'utf8');
        await restored;
      },
      'JS-config creation test',
    );
  });

  test('a newly discovered broken nested config keeps valid ancestor config active', async () => {
    const root = getWorkspaceRoot();
    const rootConfigPath = path.join(root, 'rslint.config.js');
    const originalRootConfig = fs.readFileSync(rootConfigPath, 'utf8');
    const nestedDir = path.join(root, 'broken-nested-config');
    const nestedFilePath = path.join(nestedDir, 'index.ts');
    const nestedConfigPath = path.join(nestedDir, 'rslint.config.js');
    const attemptedLoadPath = path.join(nestedDir, 'config-load-attempted');
    const postFailureFilePath = path.join(nestedDir, 'post-failure.ts');
    const rootDoc = await openFixture('index.ts');
    let nestedDoc: vscode.TextDocument | undefined;
    let postFailureDoc: vscode.TextDocument | undefined;

    const rootConfigWithMarker = `export default [{
  files: ['**/*.ts'],
  languageOptions: {
    parserOptions: { projectService: false, project: ['./tsconfig.json'] },
  },
  rules: {
    '@typescript-eslint/no-unsafe-member-access': 'warn',
    'no-debugger': 'error',
  },
  plugins: ['@typescript-eslint'],
}];
`;

    await withFailClosedCleanup(
      async () => {
        fs.mkdirSync(nestedDir, { recursive: true });
        fs.writeFileSync(nestedFilePath, 'debugger;\n', 'utf8');
        fs.writeFileSync(rootConfigPath, rootConfigWithMarker, 'utf8');

        await vscode.window.showTextDocument(rootDoc);
        await waitForDiagnostics(rootDoc, (diags) =>
          diags.some(
            (diagnostic) =>
              diagnostic.message.includes('no-unsafe-member-access') &&
              diagnostic.severity === vscode.DiagnosticSeverity.Warning,
          ),
        );
        nestedDoc = await vscode.workspace.openTextDocument(nestedFilePath);
        await vscode.window.showTextDocument(nestedDoc);
        await waitForDiagnostics(nestedDoc, (diags) =>
          diags.some(
            (diagnostic) =>
              diagnostic.message.includes('no-debugger') &&
              diagnostic.severity === vscode.DiagnosticSeverity.Error,
          ),
        );
        await closeTextEditor(nestedDoc);
        nestedDoc = undefined;

        fs.writeFileSync(
          nestedConfigPath,
          `import fs from 'node:fs';
fs.writeFileSync(${JSON.stringify(attemptedLoadPath)}, 'attempted', 'utf8');
throw new Error('intentional broken nested config');
export default [];
`,
          'utf8',
        );
        await waitForFile(attemptedLoadPath);
        assert.strictEqual(
          fs.readFileSync(attemptedLoadPath, 'utf8'),
          'attempted',
          'The broken nested config must be evaluated before fallback is asserted',
        );

        // This URI does not exist until after the broken module has executed,
        // so its diagnostics cannot be a stale snapshot from before the failed
        // refresh. Its first lint must run after the blocking config transaction
        // and resolve through the still-valid ancestor config.
        fs.writeFileSync(postFailureFilePath, 'debugger;\n', 'utf8');
        postFailureDoc =
          await vscode.workspace.openTextDocument(postFailureFilePath);
        await vscode.window.showTextDocument(postFailureDoc);
        const postFailureDiagnostics = await waitForDiagnostics(
          postFailureDoc,
          (diags) =>
            diags.some(
              (diagnostic) =>
                diagnostic.message.includes('no-debugger') &&
                diagnostic.severity === vscode.DiagnosticSeverity.Error,
            ),
        );
        assert.deepStrictEqual(
          postFailureDiagnostics
            .filter((diagnostic) => diagnostic.message.includes('no-debugger'))
            .map((diagnostic) => diagnostic.severity),
          [vscode.DiagnosticSeverity.Error],
          'The valid ancestor must lint a file opened after the broken child was evaluated',
        );
      },
      async () => {
        const temporaryDocuments = [nestedDoc, postFailureDoc].filter(
          (document): document is vscode.TextDocument => document !== undefined,
        );
        await withFailClosedCleanup(
          async () => {
            await Promise.all(
              temporaryDocuments.map((document) => closeTextEditor(document)),
            );
          },
          async () => {
            // The temporary root config deliberately lowers this rule to a
            // warning. Waiting for Error makes restoration an observable
            // config-transaction barrier before the next test starts.
            const rootRestored = waitForDiagnostics(rootDoc, (diags) =>
              diags.some(
                (diagnostic) =>
                  diagnostic.message.includes('no-unsafe-member-access') &&
                  diagnostic.severity === vscode.DiagnosticSeverity.Error,
              ),
            );
            fs.writeFileSync(rootConfigPath, originalRootConfig, 'utf8');
            fs.rmSync(nestedDir, { recursive: true, force: true });
            assert.ok(
              !fs.existsSync(nestedDir),
              'Broken nested-config fixtures must be deleted during cleanup',
            );
            await rootRestored;
          },
          'Broken nested-config resource cleanup',
        );
      },
      'Broken nested-config test',
    );
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

    await withFailClosedCleanup(
      async () => {
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

        const nestedDoc =
          await vscode.workspace.openTextDocument(nestedFilePath);
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
      },
      async () => {
        const restored = waitForDiagnostics(
          rootDoc,
          (diagnostics) =>
            diagnostics.some((diagnostic) =>
              diagnostic.message.includes('no-unsafe-member-access'),
            ) &&
            !diagnostics.some((diagnostic) =>
              diagnostic.message.includes('no-explicit-any'),
            ),
        );
        fs.rmSync(nestedDir, { recursive: true, force: true });
        fs.writeFileSync(rootConfigPath, originalRootConfig, 'utf8');
        await restored;
      },
      'Parent-ignore catalog test',
    );
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

    await withFailClosedCleanup(
      async () => {
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
      },
      async () => {
        const restored = waitForDiagnostics(
          rootDoc,
          (diagnostics) =>
            diagnostics.some((diagnostic) =>
              diagnostic.message.includes('no-unsafe-member-access'),
            ) &&
            !diagnostics.some((diagnostic) =>
              diagnostic.message.includes('no-explicit-any'),
            ),
        );
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
      },
      'Excluded-config search test',
    );
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

    const mutateAndExpectWarning = async (
      ruleName: string,
      mutate: () => void,
    ): Promise<void> => {
      const result = waitForDiagnostics(doc, (diags) =>
        diags.some(
          (d) =>
            d.message.includes(ruleName) &&
            d.severity === vscode.DiagnosticSeverity.Warning,
        ),
      );
      mutate();
      const diagnostics = await result;
      const diagnostic = diagnostics.find((d) => d.message.includes(ruleName));
      assert.strictEqual(
        diagnostic?.severity,
        vscode.DiagnosticSeverity.Warning,
        `Expected ${ruleName} from the selected config`,
      );
    };

    await waitForDiagnostics(doc, (diags) =>
      diags.some(
        (diagnostic) =>
          diagnostic.message.includes('no-unsafe-member-access') &&
          diagnostic.severity === vscode.DiagnosticSeverity.Error,
      ),
    );

    await withFailClosedCleanup(
      async () => {
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

        await mutateAndExpectWarning('no-explicit-any', () =>
          fs.unlinkSync(jsPath),
        );
        await mutateAndExpectWarning('no-unsafe-member-access', () =>
          fs.unlinkSync(mjsPath),
        );
        await mutateAndExpectWarning('no-explicit-any', () =>
          fs.unlinkSync(tsPath),
        );
      },
      async () => {
        const restored = waitForDiagnostics(doc, (diags) =>
          diags.some(
            (d) =>
              d.message.includes('no-unsafe-member-access') &&
              d.severity === vscode.DiagnosticSeverity.Error,
          ),
        );
        fs.writeFileSync(jsPath, originalJS, 'utf8');
        fs.rmSync(mjsPath, { force: true });
        fs.rmSync(tsPath, { force: true });
        fs.rmSync(mtsPath, { force: true });
        await restored;
      },
      'Same-directory config-priority test',
    );
  });

  test('broken higher-priority config preserves last-good and does not load a lower variant', async () => {
    const root = getWorkspaceRoot();
    const jsPath = path.join(root, 'rslint.config.js');
    const mjsPath = path.join(root, 'rslint.config.mjs');
    const attemptedLoadPath = path.join(root, 'broken-config-attempted.txt');
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

    await withFailClosedCleanup(
      async () => {
        fs.writeFileSync(mjsPath, lowerPriorityConfig, 'utf8');
        fs.writeFileSync(
          jsPath,
          `import fs from 'node:fs';
fs.writeFileSync(${JSON.stringify(attemptedLoadPath)}, 'attempted');
throw new Error('intentional broken higher-priority config');
export default [];
`,
          'utf8',
        );
        await waitForFile(attemptedLoadPath);

        // Request fresh diagnostics only after the failing module was actually
        // evaluated. A stale pre-reload snapshot cannot satisfy both content
        // transitions, and falling through to .mjs would report another rule.
        const originalContent = doc.getText();
        const editor = await vscode.window.showTextDocument(doc);
        const cleared = waitForDiagnostics(
          doc,
          (diagnostics) => diagnostics.length === 0,
        );
        assert.ok(
          await editor.edit((edit) => {
            edit.replace(
              new vscode.Range(
                doc.positionAt(0),
                doc.positionAt(doc.getText().length),
              ),
              'const safe = 1;\n',
            );
          }),
          'Editing the last-good config probe to clean content should succeed',
        );
        await cleared;

        const lastGoodApplied = waitForDiagnostics(doc, (diagnostics) =>
          diagnostics.some((diagnostic) =>
            diagnostic.message.includes('no-unsafe-member-access'),
          ),
        );
        assert.ok(
          await editor.edit((edit) => {
            edit.replace(
              new vscode.Range(
                doc.positionAt(0),
                doc.positionAt(doc.getText().length),
              ),
              originalContent,
            );
          }),
          'Restoring the last-good config probe content should succeed',
        );
        const diagnostics = await lastGoodApplied;
        assert.ok(
          !diagnostics.some((d) => d.message.includes('no-explicit-any')),
          'A broken .js must not fall through to .mjs or JSON',
        );
      },
      async () => {
        await revertTextDocument(doc);
        const restored = waitForDiagnostics(doc, (diags) =>
          diags.some((d) => d.message.includes('no-unsafe-member-access')),
        );
        fs.writeFileSync(jsPath, originalJS, 'utf8');
        fs.rmSync(mjsPath, { force: true });
        fs.rmSync(attemptedLoadPath, { force: true });
        await restored;
      },
      'Broken higher-priority config test',
    );
  });
});

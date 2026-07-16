import * as assert from 'node:assert';
import {
  commands,
  Range,
  Uri,
  window,
  workspace,
  type TextDocument,
  type TextDocumentChangeEvent,
  type WorkspaceFolder,
} from 'vscode';
import {
  WorkspaceDocumentRouter,
  type DocumentRoutingRuntime,
} from '../../src/WorkspaceDocumentRouter';

class FakeRoutingRuntime implements DocumentRoutingRuntime {
  readonly events: string[] = [];
  failOpen = false;

  constructor(
    readonly rootKey: string,
    readonly workspaceFolder: WorkspaceFolder,
  ) {}

  async sendDocumentOpen(document: TextDocument): Promise<void> {
    this.events.push(`open:${document.uri}:${document.getText()}`);
    if (this.failOpen) throw new Error(`open failed for ${this.rootKey}`);
  }

  async sendDocumentClose(document: TextDocument): Promise<void> {
    this.events.push(`close:${document.uri}`);
  }

  clearDocumentDiagnostics(uri: Uri): void {
    this.events.push(`clear:${uri}`);
  }
}

function detachedDocument(uri: Uri): TextDocument {
  return {
    uri,
    languageId: 'typescript',
    getText: () => 'const stale = 1;\n',
  } as TextDocument;
}

suite('workspace document router', () => {
  let document: TextDocument;
  let parentFolder: WorkspaceFolder;
  let childFolder: WorkspaceFolder;
  let testDirectory: Uri;

  suiteSetup(async () => {
    const workspaceFolder = workspace.workspaceFolders?.[0];
    assert.ok(workspaceFolder, 'test requires a workspace folder');
    parentFolder = workspaceFolder;
    testDirectory = Uri.joinPath(
      workspaceFolder.uri,
      `.router-test-${Date.now()}`,
    );
    await workspace.fs.createDirectory(testDirectory);
    const file = Uri.joinPath(testDirectory, 'nested.ts');
    await workspace.fs.writeFile(file, Buffer.from('const value = 1;\n'));
    document = await workspace.openTextDocument(file);
    childFolder = {
      uri: testDirectory,
      name: 'nested',
      index: workspaceFolder.index + 1,
    };
  });

  suiteTeardown(async () => {
    await window.showTextDocument(document, { preview: false });
    await commands.executeCommand('workbench.action.closeActiveEditor');
    await workspace.fs.delete(testDirectory, { recursive: true });
  });

  test('hands an open document to the longest active root', async () => {
    const router = new WorkspaceDocumentRouter();
    const parent = new FakeRoutingRuntime(
      parentFolder.uri.toString(),
      parentFolder,
    );
    const child = new FakeRoutingRuntime(
      childFolder.uri.toString(),
      childFolder,
    );

    await router.activate(parent);
    assert.strictEqual(router.getServerOpenOwner(document), parent.rootKey);
    parent.events.length = 0;

    await router.activate(child);
    assert.strictEqual(router.getServerOpenOwner(document), child.rootKey);
    assert.deepStrictEqual(
      parent.events.filter((event) => event.includes(document.uri.toString())),
      [`close:${document.uri}`, `clear:${document.uri}`],
    );
    assert.ok(
      child.events.some((event) =>
        event.startsWith(`open:${document.uri}:const value = 1;`),
      ),
    );

    child.events.length = 0;
    parent.events.length = 0;
    await router.deactivate(child.rootKey);
    assert.strictEqual(router.getServerOpenOwner(document), parent.rootKey);
    assert.deepStrictEqual(
      child.events.filter((event) => event.includes(document.uri.toString())),
      [`close:${document.uri}`, `clear:${document.uri}`],
    );
    assert.ok(
      parent.events.some((event) =>
        event.startsWith(`open:${document.uri}:const value = 1;`),
      ),
    );
    await router.closeAll();
  });

  test('forwards text changes only through the server-open owner', async () => {
    const router = new WorkspaceDocumentRouter();
    const parent = new FakeRoutingRuntime(
      parentFolder.uri.toString(),
      parentFolder,
    );
    const child = new FakeRoutingRuntime(
      childFolder.uri.toString(),
      childFolder,
    );
    await router.activate(parent);
    await router.activate(child);

    const event: TextDocumentChangeEvent = {
      document,
      reason: undefined,
      contentChanges: [
        {
          range: new Range(0, 0, 0, 0),
          rangeOffset: 0,
          rangeLength: 0,
          text: 'x',
        },
      ],
    };
    let parentChanges = 0;
    let childChanges = 0;
    await Promise.all([
      router.createMiddleware(parent).didChange?.(event, async () => {
        parentChanges++;
      }),
      router.createMiddleware(child).didChange?.(event, async () => {
        childChanges++;
      }),
    ]);
    assert.strictEqual(parentChanges, 0);
    assert.strictEqual(childChanges, 1);
    await router.closeAll();
  });

  test('does not duplicate didOpen after a topology handoff opened the document', async () => {
    const router = new WorkspaceDocumentRouter();
    const parent = new FakeRoutingRuntime(
      parentFolder.uri.toString(),
      parentFolder,
    );
    await router.activate(parent);

    let forwarded = 0;
    await router.createMiddleware(parent).didOpen?.(document, async () => {
      forwarded++;
    });

    assert.strictEqual(forwarded, 0);
    assert.strictEqual(
      parent.events.filter((event) => event.startsWith(`open:${document.uri}:`))
        .length,
      1,
    );
    await router.closeAll();
  });

  test('clears ownership and diagnostics when a normal didClose fails', async () => {
    const router = new WorkspaceDocumentRouter();
    const parent = new FakeRoutingRuntime(
      parentFolder.uri.toString(),
      parentFolder,
    );
    await router.activate(parent);
    parent.events.length = 0;

    const didClose = router.createMiddleware(parent).didClose;
    assert.ok(didClose);
    await assert.rejects(
      didClose(document, async () => {
        throw new Error('server close failed');
      }),
      /server close failed/,
    );

    assert.deepStrictEqual(parent.events, [`clear:${document.uri}`]);
    assert.strictEqual(router.getServerOpenOwner(document), undefined);
    await router.closeAll();
  });

  test('resets a restarted server session before its didOpen replay', async () => {
    const router = new WorkspaceDocumentRouter();
    const parent = new FakeRoutingRuntime(
      parentFolder.uri.toString(),
      parentFolder,
    );
    await router.activate(parent);
    parent.events.length = 0;

    const didOpen = router.createMiddleware(parent).didOpen;
    assert.ok(didOpen);
    const reset = router.resetServerSession(parent);
    const replay = didOpen(document, async () => {
      parent.events.push(`replay-open:${document.uri}`);
    });
    await Promise.all([reset, replay]);

    assert.deepStrictEqual(
      parent.events.filter((event) => event.includes(document.uri.toString())),
      [`clear:${document.uri}`, `replay-open:${document.uri}`],
    );
    assert.strictEqual(router.getServerOpenOwner(document), parent.rootKey);
    await router.closeAll();
  });

  test('lets a nested root activate during the restart gap after early reset', async () => {
    const router = new WorkspaceDocumentRouter();
    const parent = new FakeRoutingRuntime(
      parentFolder.uri.toString(),
      parentFolder,
    );
    const child = new FakeRoutingRuntime(
      childFolder.uri.toString(),
      childFolder,
    );
    await router.activate(parent);
    await router.resetServerSession(parent);
    parent.events.length = 0;

    await router.activate(child);

    assert.strictEqual(router.getServerOpenOwner(document), child.rootKey);
    assert.strictEqual(
      parent.events.some((event) => event === `close:${document.uri}`),
      false,
      'a cleared old transport must not receive didClose',
    );
    assert.ok(
      child.events.some((event) =>
        event.startsWith(`open:${document.uri}:const value = 1;`),
      ),
    );
    await router.closeAll();
  });

  test('forgets a document closed during restart before the same URI reopens', async () => {
    const file = Uri.joinPath(testDirectory, 'restart-closed.ts');
    const staleDocument = detachedDocument(file);
    const router = new WorkspaceDocumentRouter();
    const parent = new FakeRoutingRuntime(
      parentFolder.uri.toString(),
      parentFolder,
    );
    await router.activate(parent);
    const didOpen = router.createMiddleware(parent).didOpen;
    assert.ok(didOpen);
    await didOpen(staleDocument, async () => undefined);
    assert.strictEqual(
      router.getServerOpenOwner(staleDocument),
      parent.rootKey,
    );

    // Simulate the LanguageClient feature-listener gap: VS Code closes the
    // document, but this router never receives that generation's didClose. A
    // detached TextDocument models the retained old session while the URI is
    // absent from workspace.textDocuments.
    assert.strictEqual(
      workspace.textDocuments.some(
        (candidate) => candidate.uri.toString() === file.toString(),
      ),
      false,
    );

    parent.events.length = 0;
    await router.resetServerSession(parent);
    assert.deepStrictEqual(
      parent.events.filter((event) => event.includes(file.toString())),
      [`clear:${file}`],
    );

    const reopenedDocument = detachedDocument(file);
    let forwarded = 0;
    await didOpen(reopenedDocument, async () => {
      forwarded++;
    });
    assert.strictEqual(forwarded, 1);
    assert.strictEqual(
      router.getServerOpenOwner(reopenedDocument),
      parent.rootKey,
    );

    await router.closeAll();
  });

  test('drains a closed restart-gap session when its root is removed', async () => {
    const file = Uri.joinPath(testDirectory, 'removed-root-stale.ts');
    const staleDocument = detachedDocument(file);
    const router = new WorkspaceDocumentRouter();
    const parent = new FakeRoutingRuntime(
      parentFolder.uri.toString(),
      parentFolder,
    );
    await router.activate(parent);
    const firstDidOpen = router.createMiddleware(parent).didOpen;
    assert.ok(firstDidOpen);
    await firstDidOpen(staleDocument, async () => undefined);
    assert.strictEqual(
      workspace.textDocuments.some(
        (candidate) => candidate.uri.toString() === file.toString(),
      ),
      false,
    );

    parent.events.length = 0;
    await router.deactivate(parent.rootKey);
    assert.deepStrictEqual(
      parent.events.filter((event) => event.includes(file.toString())),
      [`clear:${file}`],
    );

    await router.activate(parent);
    const reopenedDocument = detachedDocument(file);
    const didOpen = router.createMiddleware(parent).didOpen;
    assert.ok(didOpen);
    let forwarded = 0;
    await didOpen(reopenedDocument, async () => {
      forwarded++;
    });
    assert.strictEqual(forwarded, 1);
    assert.strictEqual(
      router.getServerOpenOwner(reopenedDocument),
      parent.rootKey,
    );

    await router.closeAll();
  });

  test('rejects diagnostics from a replaced runtime with the same root URI', async () => {
    const router = new WorkspaceDocumentRouter();
    const oldRuntime = new FakeRoutingRuntime(
      parentFolder.uri.toString(),
      parentFolder,
    );
    const replacement = new FakeRoutingRuntime(
      parentFolder.uri.toString(),
      parentFolder,
    );
    const oldMiddleware = router.createMiddleware(oldRuntime);
    const replacementMiddleware = router.createMiddleware(replacement);
    await router.activate(oldRuntime);
    await router.deactivate(oldRuntime.rootKey);
    await router.activate(replacement);

    let oldDiagnostics = 0;
    let replacementDiagnostics = 0;
    oldMiddleware.handleDiagnostics?.(document.uri, [], () => {
      oldDiagnostics++;
    });
    replacementMiddleware.handleDiagnostics?.(document.uri, [], () => {
      replacementDiagnostics++;
    });

    assert.strictEqual(oldDiagnostics, 0);
    assert.strictEqual(replacementDiagnostics, 1);
    await router.closeAll();
  });

  test('drops a code action whose ownership changes while awaiting', async () => {
    const router = new WorkspaceDocumentRouter();
    const parent = new FakeRoutingRuntime(
      parentFolder.uri.toString(),
      parentFolder,
    );
    const child = new FakeRoutingRuntime(
      childFolder.uri.toString(),
      childFolder,
    );
    await router.activate(parent);

    let release!: () => void;
    const pending = new Promise<void>((resolve) => {
      release = resolve;
    });
    const action = router.createMiddleware(parent).provideCodeActions?.(
      document,
      new Range(0, 0, 0, 0),
      { diagnostics: [], only: undefined, triggerKind: 1 },
      {
        isCancellationRequested: false,
        onCancellationRequested: () => ({ dispose() {} }),
      },
      async () => {
        await pending;
        return [];
      },
    );

    await router.activate(child);
    release();
    assert.strictEqual(await Promise.resolve(action), undefined);
    await router.closeAll();
  });

  test('rolls back to the prior owner when a new owner cannot open', async () => {
    const router = new WorkspaceDocumentRouter();
    const parent = new FakeRoutingRuntime(
      parentFolder.uri.toString(),
      parentFolder,
    );
    const child = new FakeRoutingRuntime(
      childFolder.uri.toString(),
      childFolder,
    );
    child.failOpen = true;
    await router.activate(parent);

    await assert.rejects(
      router.activate(child),
      /failed to activate document owner/,
    );
    assert.strictEqual(router.getServerOpenOwner(document), parent.rootKey);
    assert.strictEqual(router.ownerKeyForDocument(document), parent.rootKey);
    await router.closeAll();
  });
});

import assert from 'node:assert/strict';
import path from 'node:path';

import * as vscode from 'vscode';

import { EditorRuntimeClient } from '../../src/EditorRuntimeClient';
import {
  RuntimeDocumentRouter,
  type RuntimeRoutingTarget,
} from '../../src/RuntimeDocumentRouter';

suite('runtime document router', () => {
  test('hands a live document between explicit runtime owners exactly once', async () => {
    const folder = vscode.workspace.workspaceFolders?.[0];
    assert.ok(folder);
    const document = await vscode.workspace.openTextDocument(
      vscode.Uri.file(path.join(folder.uri.fsPath, 'src/index.ts')),
    );
    const events: string[] = [];
    const runtime = (runtimeKey: string): RuntimeRoutingTarget => ({
      runtimeKey,
      workspaceFolder: folder,
      isRunning: () => true,
      sendDocumentOpen: async () => {
        events.push(`open:${runtimeKey}`);
      },
      sendDocumentClose: async () => {
        events.push(`close:${runtimeKey}`);
      },
      clearDocumentDiagnostics: () => {
        events.push(`clear:${runtimeKey}`);
      },
    });
    const a = runtime('a');
    const b = runtime('b');
    const router = new RuntimeDocumentRouter();
    router.register(a);
    router.register(b);

    await router.assign(document, a);
    await router.assign(document, a);
    await router.assign(document, b);
    await router.assign(document, undefined);

    assert.deepEqual(events, [
      'open:a',
      'close:a',
      'clear:a',
      'open:b',
      'close:b',
      'clear:b',
    ]);
    await router.closeAll();
  });

  test('replaces an open session when the same URI gets a new document instance', async () => {
    const folder = vscode.workspace.workspaceFolders?.[0];
    assert.ok(folder);
    const document = await vscode.workspace.openTextDocument(
      vscode.Uri.file(path.join(folder.uri.fsPath, 'src/index.ts')),
    );
    const replacement = { uri: document.uri } as vscode.TextDocument;
    const events: string[] = [];
    const runtime: RuntimeRoutingTarget = {
      runtimeKey: 'runtime',
      workspaceFolder: folder,
      isRunning: () => true,
      sendDocumentOpen: async (opened) => {
        events.push(opened === replacement ? 'open:new' : 'open:old');
      },
      sendDocumentClose: async (closed) => {
        events.push(closed === replacement ? 'close:new' : 'close:old');
      },
      clearDocumentDiagnostics: () => {
        events.push('clear');
      },
    };
    const router = new RuntimeDocumentRouter();
    router.register(runtime);

    await router.assign(document, runtime);
    await router.assign(replacement, runtime);

    assert.deepEqual(events, ['open:old', 'close:old', 'clear', 'open:new']);
    await router.closeAll();
  });

  test('does not let a same-URI replacement borrow another document permit', async () => {
    const folder = vscode.workspace.workspaceFolders?.[0];
    assert.ok(folder);
    const document = await vscode.workspace.openTextDocument(
      vscode.Uri.file(path.join(folder.uri.fsPath, 'src/index.ts')),
    );
    const replacement = { uri: document.uri } as vscode.TextDocument;
    const events: string[] = [];
    let releaseFirstOpen!: () => void;
    const firstOpenGate = new Promise<void>((resolve) => {
      releaseFirstOpen = resolve;
    });
    let firstOpenStarted!: () => void;
    const firstOpen = new Promise<void>((resolve) => {
      firstOpenStarted = resolve;
    });
    let middleware!: ReturnType<RuntimeDocumentRouter['createMiddleware']>;
    const runtime: RuntimeRoutingTarget = {
      runtimeKey: 'permit-identity',
      workspaceFolder: folder,
      isRunning: () => true,
      sendDocumentOpen: async (opened) => {
        await middleware.didOpen?.(opened, async () => {
          events.push(opened === replacement ? 'next:new' : 'next:old');
          if (opened === document) {
            firstOpenStarted();
            await firstOpenGate;
          }
        });
      },
      sendDocumentClose: async () => undefined,
      clearDocumentDiagnostics: () => undefined,
    };
    const router = new RuntimeDocumentRouter();
    router.register(runtime);
    middleware = router.createMiddleware(runtime);

    const assignment = router.assign(document, runtime);
    await firstOpen;
    const staleDidOpen = middleware.didOpen?.(replacement, async () => {
      events.push('next:new');
    });
    await Promise.resolve();
    assert.deepEqual(
      events,
      ['next:old'],
      'the permit must be tied to TextDocument identity, not just its URI',
    );
    releaseFirstOpen();
    await assignment;
    await staleDidOpen;
    assert.deepEqual(events, ['next:old']);
    await router.closeAll();
  });

  test('opens deferred assignments when a runtime starts and reopens after restart', async () => {
    const folder = vscode.workspace.workspaceFolders?.[0];
    assert.ok(folder);
    const document = await vscode.workspace.openTextDocument(
      vscode.Uri.file(path.join(folder.uri.fsPath, 'src/index.ts')),
    );
    let running = false;
    const events: string[] = [];
    const runtime: RuntimeRoutingTarget = {
      runtimeKey: 'deferred',
      workspaceFolder: folder,
      isRunning: () => running,
      sendDocumentOpen: async () => {
        events.push('open');
      },
      sendDocumentClose: async () => {
        events.push('close');
      },
      clearDocumentDiagnostics: () => {
        events.push('clear');
      },
    };
    const router = new RuntimeDocumentRouter();
    router.register(runtime);

    await router.assign(document, runtime);
    const middleware = router.createMiddleware(runtime);
    await middleware.didOpen?.(document, async () => {
      events.push('automatic-open');
    });
    assert.deepEqual(events, []);
    running = true;
    await router.runtimeBecameReady(runtime);
    await router.runtimeBecameReady(runtime);
    assert.deepEqual(events, ['open']);

    await router.resetServerSession(runtime);
    assert.deepEqual(events, ['open', 'clear']);
    await router.runtimeBecameReady(runtime);
    assert.deepEqual(events, ['open', 'clear', 'open']);
    await router.closeAll();
  });

  test('continues reopening other documents when one ready-time didOpen fails', async () => {
    const folder = vscode.workspace.workspaceFolders?.[0];
    assert.ok(folder);
    const opened = await Promise.all([
      vscode.workspace.openTextDocument(
        vscode.Uri.file(path.join(folder.uri.fsPath, 'src/index.ts')),
      ),
      vscode.workspace.openTextDocument(
        vscode.Uri.file(
          path.join(folder.uri.fsPath, 'parent-ignore-catalog-probe/index.ts'),
        ),
      ),
    ]);
    const documents = [...opened].sort(
      (left, right) =>
        vscode.workspace.textDocuments.indexOf(left) -
        vscode.workspace.textDocuments.indexOf(right),
    );
    let running = false;
    const attempts: string[] = [];
    const failedDocument = documents[0];
    assert.ok(failedDocument);
    const runtime: RuntimeRoutingTarget = {
      runtimeKey: 'ready-open-isolation',
      workspaceFolder: folder,
      isRunning: () => running,
      sendDocumentOpen: async (document) => {
        attempts.push(document.uri.toString());
        if (document === failedDocument) {
          throw new Error('synthetic first didOpen failure');
        }
      },
      sendDocumentClose: async () => undefined,
      clearDocumentDiagnostics: () => undefined,
    };
    const router = new RuntimeDocumentRouter();
    router.register(runtime);
    for (const document of documents) await router.assign(document, runtime);

    running = true;
    await assert.rejects(
      router.runtimeBecameReady(runtime),
      /failed to open one or more routed documents/,
    );
    assert.deepEqual(
      new Set(attempts),
      new Set(documents.map((document) => document.uri.toString())),
      'a failed first didOpen must not starve later documents',
    );
    await router.closeAll();
  });

  test('filters middleware traffic and drops a stale code-action result after handoff', async () => {
    const folder = vscode.workspace.workspaceFolders?.[0];
    assert.ok(folder);
    const document = await vscode.workspace.openTextDocument(
      vscode.Uri.file(path.join(folder.uri.fsPath, 'src/index.ts')),
    );
    const events: string[] = [];
    const runtime = (runtimeKey: string): RuntimeRoutingTarget => ({
      runtimeKey,
      workspaceFolder: folder,
      isRunning: () => true,
      sendDocumentOpen: async () => {
        events.push(`open:${runtimeKey}`);
      },
      sendDocumentClose: async () => {
        events.push(`close:${runtimeKey}`);
      },
      clearDocumentDiagnostics: () => {
        events.push(`clear:${runtimeKey}`);
      },
    });
    const a = runtime('a');
    const b = runtime('b');
    const router = new RuntimeDocumentRouter();
    router.register(a);
    router.register(b);
    await router.assign(document, a);
    const middlewareA = router.createMiddleware(a);
    const middlewareB = router.createMiddleware(b);

    await middlewareA.didChange?.(
      {
        document,
        contentChanges: [],
        reason: vscode.TextDocumentChangeReason.Undo,
      },
      async () => {
        events.push('change:a');
      },
    );
    await middlewareB.didChange?.(
      {
        document,
        contentChanges: [],
        reason: vscode.TextDocumentChangeReason.Undo,
      },
      async () => {
        events.push('change:b');
      },
    );
    await middlewareA.didSave?.(document, async () => {
      events.push('save:a');
    });
    await middlewareB.didSave?.(document, async () => {
      events.push('save:b');
    });
    middlewareA.handleDiagnostics?.(document.uri, [], () => {
      events.push('diagnostics:a');
    });
    middlewareB.handleDiagnostics?.(document.uri, [], () => {
      events.push('diagnostics:b');
    });

    let resolveCodeActions!: (value: vscode.CodeAction[]) => void;
    const stale = middlewareA.provideCodeActions?.(
      document,
      new vscode.Range(0, 0, 0, 0),
      {
        diagnostics: [],
        only: undefined,
        triggerKind: vscode.CodeActionTriggerKind.Invoke,
      },
      new vscode.CancellationTokenSource().token,
      async () =>
        new Promise<vscode.CodeAction[]>((resolve) => {
          resolveCodeActions = resolve;
        }),
    );
    await router.assign(document, b);
    resolveCodeActions([new vscode.CodeAction('stale')]);
    assert.equal(await stale, undefined);

    assert.deepEqual(events, [
      'open:a',
      'change:a',
      'save:a',
      'diagnostics:a',
      'close:a',
      'clear:a',
      'open:b',
    ]);
    await router.closeAll();
  });

  test('completes ownership handoff even when old close and diagnostic cleanup fail', async () => {
    const folder = vscode.workspace.workspaceFolders?.[0];
    assert.ok(folder);
    const document = await vscode.workspace.openTextDocument(
      vscode.Uri.file(path.join(folder.uri.fsPath, 'src/index.ts')),
    );
    const events: string[] = [];
    const broken: RuntimeRoutingTarget = {
      runtimeKey: 'broken',
      workspaceFolder: folder,
      isRunning: () => true,
      sendDocumentOpen: async () => {
        events.push('open:broken');
      },
      sendDocumentClose: async () => {
        events.push('close:broken');
        throw new Error('close failed');
      },
      clearDocumentDiagnostics: () => {
        events.push('clear:broken');
        throw new Error('clear failed');
      },
    };
    const healthy: RuntimeRoutingTarget = {
      runtimeKey: 'healthy',
      workspaceFolder: folder,
      isRunning: () => true,
      sendDocumentOpen: async () => {
        events.push('open:healthy');
      },
      sendDocumentClose: async () => {
        events.push('close:healthy');
      },
      clearDocumentDiagnostics: () => {
        events.push('clear:healthy');
      },
    };
    const router = new RuntimeDocumentRouter();
    router.register(broken);
    router.register(healthy);
    await router.assign(document, broken);
    await assert.rejects(
      router.assign(document, healthy),
      /failed to transfer routed document/,
    );

    let healthyChange = false;
    await router.createMiddleware(healthy).didChange?.(
      {
        document,
        contentChanges: [],
        reason: vscode.TextDocumentChangeReason.Undo,
      },
      async () => {
        healthyChange = true;
      },
    );
    assert.equal(healthyChange, true, 'the new runtime must own the document');
    assert.deepEqual(events, [
      'open:broken',
      'close:broken',
      'clear:broken',
      'open:healthy',
    ]);
    await router.closeAll();
  });

  test('rolls back to the last-good owner when replacement didOpen fails', async () => {
    const folder = vscode.workspace.workspaceFolders?.[0];
    assert.ok(folder);
    const document = await vscode.workspace.openTextDocument(
      vscode.Uri.file(path.join(folder.uri.fsPath, 'src/index.ts')),
    );
    const events: string[] = [];
    const previous: RuntimeRoutingTarget = {
      runtimeKey: 'last-good',
      workspaceFolder: folder,
      isRunning: () => true,
      sendDocumentOpen: async () => {
        events.push('open:last-good');
      },
      sendDocumentClose: async () => {
        events.push('close:last-good');
      },
      clearDocumentDiagnostics: () => {
        events.push('clear:last-good');
      },
    };
    const replacement: RuntimeRoutingTarget = {
      runtimeKey: 'replacement',
      workspaceFolder: folder,
      isRunning: () => true,
      sendDocumentOpen: async () => {
        events.push('open:replacement');
        throw new Error('replacement didOpen failed');
      },
      sendDocumentClose: async () => undefined,
      clearDocumentDiagnostics: () => undefined,
    };
    const router = new RuntimeDocumentRouter();
    router.register(previous);
    router.register(replacement);
    await router.assign(document, previous);

    await assert.rejects(
      router.assign(document, replacement),
      /replacement didOpen failed/,
    );
    let previousChange = false;
    let replacementChange = false;
    await router.createMiddleware(previous).didChange?.(
      {
        document,
        contentChanges: [],
        reason: vscode.TextDocumentChangeReason.Undo,
      },
      async () => {
        previousChange = true;
      },
    );
    await router.createMiddleware(replacement).didChange?.(
      {
        document,
        contentChanges: [],
        reason: vscode.TextDocumentChangeReason.Undo,
      },
      async () => {
        replacementChange = true;
      },
    );
    assert.equal(previousChange, true);
    assert.equal(replacementChange, false);
    assert.deepEqual(events, [
      'open:last-good',
      'close:last-good',
      'clear:last-good',
      'open:replacement',
      'open:last-good',
    ]);
    await router.closeAll();
  });

  test('rolls back a same-URI replacement to its current document instance', async () => {
    const folder = vscode.workspace.workspaceFolders?.[0];
    assert.ok(folder);
    const oldDocument = await vscode.workspace.openTextDocument(
      vscode.Uri.file(path.join(folder.uri.fsPath, 'src/index.ts')),
    );
    const currentDocument = {
      uri: oldDocument.uri,
    } as vscode.TextDocument;
    const events: string[] = [];
    const identity = (document: vscode.TextDocument): string =>
      document === currentDocument ? 'current' : 'old';
    const previous: RuntimeRoutingTarget = {
      runtimeKey: 'same-uri-last-good',
      workspaceFolder: folder,
      isRunning: () => true,
      sendDocumentOpen: async (document) => {
        events.push(`open:last-good:${identity(document)}`);
      },
      sendDocumentClose: async (document) => {
        events.push(`close:last-good:${identity(document)}`);
      },
      clearDocumentDiagnostics: () => undefined,
    };
    const replacement: RuntimeRoutingTarget = {
      runtimeKey: 'same-uri-replacement',
      workspaceFolder: folder,
      isRunning: () => true,
      sendDocumentOpen: async (document) => {
        events.push(`open:replacement:${identity(document)}`);
        throw new Error('replacement didOpen failed');
      },
      sendDocumentClose: async () => undefined,
      clearDocumentDiagnostics: () => undefined,
    };
    const router = new RuntimeDocumentRouter();
    router.register(previous);
    router.register(replacement);
    await router.assign(oldDocument, previous);

    await assert.rejects(
      router.assign(currentDocument, replacement),
      /replacement didOpen failed/,
    );
    let currentChange = false;
    await router.createMiddleware(previous).didChange?.(
      {
        document: currentDocument,
        contentChanges: [],
        reason: vscode.TextDocumentChangeReason.Undo,
      },
      async () => {
        currentChange = true;
      },
    );
    assert.equal(currentChange, true);
    assert.deepEqual(events, [
      'open:last-good:old',
      'close:last-good:old',
      'open:replacement:current',
      'open:last-good:current',
    ]);
    await router.closeAll();
  });

  test('retires a restarted generation that violates the readiness protocol', async () => {
    let terminalFailures = 0;
    type ReadyClientHarness = {
      generationReady: boolean;
      handleRuntimeReady(params: unknown): Promise<void>;
    };
    const client = Object.assign(Object.create(EditorRuntimeClient.prototype), {
      options: {
        onTerminalFailure: () => {
          terminalFailures++;
        },
      },
      closing: false,
      terminalFailureReported: false,
      firstRuntimeReadySettled: true,
      generationReady: true,
    }) as ReadyClientHarness;

    await client.handleRuntimeReady({ protocolVersion: 999 });
    await new Promise<void>((resolve) => queueMicrotask(resolve));

    assert.equal(client.generationReady, false);
    assert.equal(terminalFailures, 1);
  });
});

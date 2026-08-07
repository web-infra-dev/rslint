import * as assert from 'node:assert';
import {
  commands,
  Uri,
  window,
  workspace,
  type TextDocument,
  type WorkspaceFolder,
} from 'vscode';
import type {
  CoreInstallation,
  ResolvedCoreRuntime,
} from '../../src/CoreResolver';
import {
  RuntimeManager,
  type ManagedRslintRuntime,
  type RuntimeCoreResolver,
  type RuntimeManagerLogger,
} from '../../src/RuntimeManager';
import { WorkspaceDocumentRouter } from '../../src/WorkspaceDocumentRouter';

class FakeResolver implements RuntimeCoreResolver {
  readonly keys = new Map<string, string>();
  readonly failures = new Set<string>();
  readonly gates = new Map<string, Deferred<void>>();
  readonly started: string[] = [];
  clearCalls = 0;

  clear(): void {
    this.clearCalls++;
  }

  async resolve(
    document: TextDocument,
    workspaceFolder: WorkspaceFolder,
  ): Promise<ResolvedCoreRuntime> {
    const identity = this.keys.get(document.uri.toString()) ?? 'shared-core';
    this.started.push(identity);
    await this.gates.get(identity)?.promise;
    if (this.failures.has(identity)) {
      throw new Error(`resolution failed for ${identity}`);
    }
    // The fake runtime never consumes module factories or the binary.
    // rslint-disable-next-line @typescript-eslint/no-unsafe-type-assertion
    const installation = {
      identity,
      packageDirectory: `/core/${identity}`,
      version: identity,
      binaryPath: `/core/${identity}/rslint`,
      protocolVersion: 1,
    } as CoreInstallation;
    return {
      key: `${workspaceFolder.uri}\0${identity}`,
      workspaceFolder,
      installation,
    };
  }
}

class FakeRuntime implements ManagedRslintRuntime {
  readonly opened: string[] = [];
  readonly closedDocuments: string[] = [];
  startCalls = 0;
  closeCalls = 0;

  constructor(
    readonly rootKey: string,
    readonly workspaceFolder: WorkspaceFolder,
    readonly identity: string,
    private readonly behavior: RuntimeBehavior,
  ) {}

  async start(signal: AbortSignal): Promise<void> {
    this.startCalls++;
    if (this.behavior.failStart) {
      throw new Error(`start failed for ${this.rootKey}`);
    }
    if (this.behavior.pendingStart) {
      await new Promise<void>((_resolve, reject) => {
        if (signal.aborted) {
          reject(signal.reason);
          return;
        }
        signal.addEventListener('abort', () => reject(signal.reason), {
          once: true,
        });
      });
    }
  }

  async close(): Promise<void> {
    this.closeCalls++;
    if (this.behavior.failClose) {
      throw new Error(`close failed for ${this.rootKey}`);
    }
  }

  async sendDocumentOpen(document: TextDocument): Promise<void> {
    if (this.behavior.failOpen) {
      throw new Error(`didOpen failed for ${this.rootKey}`);
    }
    this.opened.push(document.uri.toString());
  }

  async sendDocumentClose(document: TextDocument): Promise<void> {
    this.closedDocuments.push(document.uri.toString());
  }

  clearDocumentDiagnostics(): void {}
}

const silentLogger: RuntimeManagerLogger = {
  debug() {},
  info() {},
  error() {},
};

interface Deferred<T> {
  readonly promise: Promise<T>;
  resolve(value: T): void;
}

interface RuntimeBehavior {
  readonly failFactory?: boolean;
  readonly failStart?: boolean;
  readonly pendingStart?: boolean;
  readonly failOpen?: boolean;
  readonly failClose?: boolean;
}

function deferred<T>(): Deferred<T> {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((resolvePromise) => {
    resolve = resolvePromise;
  });
  return { promise, resolve };
}

async function eventually(
  predicate: () => boolean,
  message: string,
): Promise<void> {
  const deadline = Date.now() + 2_000;
  while (Date.now() < deadline) {
    if (predicate()) return;
    await new Promise((resolve) => setTimeout(resolve, 5));
  }
  assert.fail(message);
}

async function reconcileAll(
  manager: RuntimeManager,
  documents: readonly TextDocument[],
): Promise<void> {
  await Promise.all(documents.map((document) => manager.reconcile(document)));
}

suite('local-core runtime manager', () => {
  let testDirectory: Uri;
  let first: TextDocument;
  let second: TextDocument;
  let workspaceFolder: WorkspaceFolder;

  setup(async () => {
    const folder = workspace.workspaceFolders?.[0];
    assert.ok(folder, 'test requires a workspace folder');
    workspaceFolder = folder;
    testDirectory = Uri.joinPath(folder.uri, `.runtime-manager-${Date.now()}`);
    await workspace.fs.createDirectory(testDirectory);
    const firstUri = Uri.joinPath(testDirectory, 'first.ts');
    const secondUri = Uri.joinPath(testDirectory, 'second.ts');
    await workspace.fs.writeFile(firstUri, Buffer.from('const first = 1;\n'));
    await workspace.fs.writeFile(secondUri, Buffer.from('const second = 2;\n'));
    first = await workspace.openTextDocument(firstUri);
    second = await workspace.openTextDocument(secondUri);
  });

  teardown(async () => {
    for (const document of [first, second]) {
      await window.showTextDocument(document, { preview: false });
      await commands.executeCommand('workbench.action.closeActiveEditor');
    }
    await workspace.fs.delete(testDirectory, { recursive: true });
  });

  function harness(
    behaviorFor: (identity: string) => RuntimeBehavior = () => ({}),
  ) {
    const resolver = new FakeResolver();
    const router = new WorkspaceDocumentRouter();
    const runtimes: FakeRuntime[] = [];
    const manager = new RuntimeManager(
      router,
      resolver,
      (resolved) => {
        const identity = resolved.installation.identity;
        const behavior = behaviorFor(identity);
        if (behavior.failFactory) {
          throw new Error(`factory failed for ${identity}`);
        }
        const runtime = new FakeRuntime(
          resolved.key,
          resolved.workspaceFolder,
          identity,
          behavior,
        );
        runtimes.push(runtime);
        return runtime;
      },
      silentLogger,
    );
    return { manager, resolver, router, runtimes };
  }

  test('reuses one physical core installation for multiple documents', async () => {
    const { manager, router, runtimes } = harness();
    await reconcileAll(manager, [first, second]);

    assert.strictEqual(runtimes.length, 1);
    assert.strictEqual(router.getServerOpenOwner(first), runtimes[0].rootKey);
    assert.strictEqual(router.getServerOpenOwner(second), runtimes[0].rootKey);
    assert.deepStrictEqual(
      new Set(runtimes[0].opened),
      new Set([first.uri.toString(), second.uri.toString()]),
    );
    await manager.close();
    assert.strictEqual(runtimes[0].closeCalls, 1);
  });

  test('does not resolve or retain runtimes for unsupported documents', async () => {
    const notesUri = Uri.joinPath(testDirectory, 'notes.md');
    await workspace.fs.writeFile(notesUri, Buffer.from('# Notes\n'));
    const notes = await workspace.openTextDocument(notesUri);
    const { manager, resolver, runtimes } = harness();

    await manager.reconcile(notes);

    assert.deepStrictEqual(resolver.started, []);
    assert.deepStrictEqual(runtimes, []);
    await manager.close();
  });

  test('keeps distinct core installations isolated', async () => {
    const { manager, resolver, router, runtimes } = harness();
    resolver.keys.set(first.uri.toString(), 'core-a');
    resolver.keys.set(second.uri.toString(), 'core-b');
    await reconcileAll(manager, [first, second]);

    assert.strictEqual(runtimes.length, 2);
    assert.notStrictEqual(
      router.getServerOpenOwner(first),
      router.getServerOpenOwner(second),
    );
    await manager.close();
  });

  test('keeps the last-good owner when a replacement fails to start', async () => {
    const { manager, resolver, router, runtimes } = harness((identity) => ({
      failStart: identity === 'broken-core',
    }));
    resolver.keys.set(first.uri.toString(), 'working-core');
    await reconcileAll(manager, [first]);
    const working = runtimes[0];

    resolver.keys.set(first.uri.toString(), 'broken-core');
    await manager.reconcile(first);

    assert.strictEqual(router.getServerOpenOwner(first), working.rootKey);
    assert.strictEqual(working.closeCalls, 0);
    assert.strictEqual(runtimes.length, 2);
    assert.strictEqual(runtimes[1].closeCalls, 1);
    await manager.close();
  });

  test('isolates an initial start failure from a document using another core', async () => {
    const { manager, resolver, router, runtimes } = harness((identity) => ({
      failStart: identity === 'broken-core',
    }));
    resolver.keys.set(first.uri.toString(), 'broken-core');
    resolver.keys.set(second.uri.toString(), 'healthy-core');

    await reconcileAll(manager, [first, second]);

    const broken = runtimes.find(
      (runtime) => runtime.identity === 'broken-core',
    );
    const healthy = runtimes.find(
      (runtime) => runtime.identity === 'healthy-core',
    );
    assert.ok(broken);
    assert.ok(healthy);
    assert.strictEqual(router.getServerOpenOwner(first), undefined);
    assert.strictEqual(router.getServerOpenOwner(second), healthy.rootKey);
    assert.strictEqual(broken.closeCalls, 1);
    assert.strictEqual(healthy.closeCalls, 0);
    await manager.close();
    assert.strictEqual(healthy.closeCalls, 1);
  });

  test('keeps the last-good owner when replacement resolution fails', async () => {
    const { manager, resolver, router, runtimes } = harness();
    resolver.keys.set(first.uri.toString(), 'working-core');
    await reconcileAll(manager, [first]);

    resolver.keys.set(first.uri.toString(), 'missing-core');
    resolver.failures.add('missing-core');
    await manager.reconcile(first);

    assert.strictEqual(router.getServerOpenOwner(first), runtimes[0].rootKey);
    assert.strictEqual(runtimes[0].closeCalls, 0);
    await manager.close();
  });

  test('restores the last-good owner when replacement didOpen fails', async () => {
    const { manager, resolver, router, runtimes } = harness((identity) => ({
      failOpen: identity === 'broken-core',
    }));
    resolver.keys.set(first.uri.toString(), 'working-core');
    await reconcileAll(manager, [first]);

    resolver.keys.set(first.uri.toString(), 'broken-core');
    await manager.reconcile(first);

    assert.strictEqual(router.getServerOpenOwner(first), runtimes[0].rootKey);
    assert.strictEqual(runtimes[0].closeCalls, 0);
    assert.strictEqual(runtimes[1].closeCalls, 1);
    await manager.close();
  });

  test('switches one shared document without restarting the remaining owner', async () => {
    const { manager, resolver, router, runtimes } = harness();
    await reconcileAll(manager, [first, second]);
    const shared = runtimes[0];

    resolver.keys.set(first.uri.toString(), 'nested-core');
    await manager.reconcile(first);

    assert.strictEqual(shared.closeCalls, 0);
    assert.strictEqual(router.getServerOpenOwner(second), shared.rootKey);
    assert.strictEqual(router.getServerOpenOwner(first), runtimes[1].rootKey);

    resolver.keys.set(second.uri.toString(), 'nested-core');
    await manager.reconcile(second);
    await eventually(
      () => shared.closeCalls === 1,
      'the unreferenced shared runtime should close',
    );
    assert.strictEqual(runtimes.length, 2);
    assert.strictEqual(
      router.getServerOpenOwner(first),
      router.getServerOpenOwner(second),
    );
    await manager.close();
  });

  test('releases a runtime only after its last document closes', async () => {
    const { manager, router, runtimes } = harness();
    await reconcileAll(manager, [first, second]);

    manager.documentClosed(first);
    await eventually(
      () => router.getServerOpenOwner(first) === undefined,
      'the first document should detach',
    );
    assert.strictEqual(runtimes[0].closeCalls, 0);

    manager.documentClosed(second);
    await eventually(
      () => runtimes[0].closeCalls === 1,
      'the last document should release the runtime',
    );
    await manager.close();
  });

  test('discards a stale resolution before it starts a runtime', async () => {
    const { manager, resolver, router, runtimes } = harness();
    resolver.keys.set(first.uri.toString(), 'slow-core');
    const gate = deferred<void>();
    resolver.gates.set('slow-core', gate);

    const slowReconcile = manager.reconcile(first);
    await eventually(
      () => resolver.started.includes('slow-core'),
      'the slow resolution should begin',
    );
    resolver.keys.set(first.uri.toString(), 'current-core');
    const currentReconcile = manager.reconcile(first);
    gate.resolve();
    await Promise.all([slowReconcile, currentReconcile]);

    assert.deepStrictEqual(
      runtimes.map((runtime) => runtime.identity),
      ['current-core'],
    );
    assert.strictEqual(router.getServerOpenOwner(first), runtimes[0].rootKey);
    await manager.close();
  });

  test('does not undo a committed switch when old-runtime shutdown fails', async () => {
    const { manager, resolver, router, runtimes } = harness((identity) => ({
      failClose: identity === 'old-core',
    }));
    resolver.keys.set(first.uri.toString(), 'old-core');
    await reconcileAll(manager, [first]);

    resolver.keys.set(first.uri.toString(), 'new-core');
    await manager.reconcile(first);
    await eventually(
      () => runtimes[0].closeCalls === 1,
      'the old runtime should attempt shutdown',
    );

    assert.strictEqual(router.getServerOpenOwner(first), runtimes[1].rootKey);
    assert.strictEqual(runtimes[1].closeCalls, 0);
    await assert.rejects(manager.close(), /failed to close runtime manager/);
    assert.strictEqual(runtimes[1].closeCalls, 1);
  });

  test('does not start a same-key replacement after shutdown fails', async () => {
    const { manager, resolver, router, runtimes } = harness((identity) => ({
      failClose: identity === 'quarantined-core',
    }));
    resolver.keys.set(first.uri.toString(), 'quarantined-core');
    await manager.reconcile(first);

    resolver.keys.set(first.uri.toString(), 'healthy-core');
    await manager.reconcile(first);
    await eventually(
      () => runtimes[0].closeCalls === 1,
      'the superseded runtime should attempt shutdown',
    );

    resolver.keys.set(first.uri.toString(), 'quarantined-core');
    await manager.reconcile(first);

    assert.strictEqual(runtimes.length, 3);
    assert.strictEqual(runtimes[2].identity, 'quarantined-core');
    assert.strictEqual(
      runtimes[2].startCalls,
      0,
      'a failed shutdown must remain a barrier for the same runtime key',
    );
    assert.strictEqual(router.getServerOpenOwner(first), runtimes[1].rootKey);
    await assert.rejects(manager.close(), /failed to close runtime manager/);
    assert.strictEqual(runtimes[1].closeCalls, 1);
  });

  test('isolates a factory failure from a document using another core', async () => {
    const { manager, resolver, router, runtimes } = harness((identity) => ({
      failFactory: identity === 'broken-core',
    }));
    resolver.keys.set(first.uri.toString(), 'broken-core');
    resolver.keys.set(second.uri.toString(), 'healthy-core');

    await reconcileAll(manager, [first, second]);

    assert.deepStrictEqual(
      runtimes.map((runtime) => runtime.identity),
      ['healthy-core'],
    );
    assert.strictEqual(router.getServerOpenOwner(first), undefined);
    assert.strictEqual(router.getServerOpenOwner(second), runtimes[0].rootKey);
    await manager.close();
  });

  test('recovers after an initially missing core becomes available', async () => {
    const { manager, resolver, router, runtimes } = harness();
    resolver.keys.set(first.uri.toString(), 'installed-later');
    resolver.failures.add('installed-later');

    await manager.reconcile(first);
    assert.strictEqual(router.getServerOpenOwner(first), undefined);
    assert.strictEqual(runtimes.length, 0);

    resolver.failures.delete('installed-later');
    manager.clearResolutionCache();
    await manager.reconcile(first);

    assert.strictEqual(resolver.clearCalls, 1);
    assert.strictEqual(runtimes.length, 1);
    assert.strictEqual(router.getServerOpenOwner(first), runtimes[0].rootKey);
    await manager.close();
  });

  test('lets a healthy core become ready while another initial core is pending', async () => {
    const { manager, resolver, router, runtimes } = harness((identity) => ({
      pendingStart: identity === 'pending-core',
    }));
    resolver.keys.set(first.uri.toString(), 'pending-core');
    resolver.keys.set(second.uri.toString(), 'healthy-core');

    manager.initialize([first, second]);
    await eventually(
      () => router.getServerOpenOwner(second) !== undefined,
      'the healthy runtime should not wait for the pending runtime',
    );

    assert.strictEqual(router.getServerOpenOwner(first), undefined);
    assert.deepStrictEqual(
      new Set(runtimes.map((runtime) => runtime.identity)),
      new Set(['pending-core', 'healthy-core']),
    );
    await manager.close();
  });

  test('aborts a pending start during terminal shutdown', async () => {
    const { manager, resolver, runtimes } = harness((identity) => ({
      pendingStart: identity === 'pending-core',
    }));
    resolver.keys.set(first.uri.toString(), 'pending-core');

    const reconciling = manager.reconcile(first);
    await eventually(
      () => runtimes.length === 1,
      'the pending runtime should be created',
    );
    await manager.close();
    await reconciling;

    assert.strictEqual(runtimes[0].closeCalls, 1);
  });

  test('aborts a pending start when its only document closes', async () => {
    const { manager, resolver, runtimes } = harness((identity) => ({
      pendingStart: identity === 'pending-core',
    }));
    resolver.keys.set(first.uri.toString(), 'pending-core');

    const reconciling = manager.reconcile(first);
    await eventually(
      () => runtimes.length === 1,
      'the pending runtime should be created',
    );
    manager.documentClosed(first);
    await reconciling;
    await eventually(
      () => runtimes[0].closeCalls === 1,
      'closing the document should cancel its pending runtime',
    );

    await manager.close();
  });

  test('clears resolver state without replacing an unchanged runtime', async () => {
    const { manager, resolver, runtimes } = harness();
    await reconcileAll(manager, [first]);
    manager.clearResolutionCache();
    await manager.reconcile(first);

    assert.strictEqual(resolver.clearCalls, 1);
    assert.strictEqual(runtimes.length, 1);
    await manager.close();
  });

  test('closes every active runtime when one terminal close fails', async () => {
    const { manager, resolver, runtimes } = harness((identity) => ({
      failClose: identity === 'broken-core',
    }));
    resolver.keys.set(first.uri.toString(), 'broken-core');
    resolver.keys.set(second.uri.toString(), 'healthy-core');
    await reconcileAll(manager, [first, second]);

    await assert.rejects(manager.close(), /failed to close runtime manager/);

    assert.strictEqual(runtimes.length, 2);
    assert.deepStrictEqual(
      new Map(
        runtimes.map((runtime) => [runtime.identity, runtime.closeCalls]),
      ),
      new Map([
        ['broken-core', 1],
        ['healthy-core', 1],
      ]),
    );
  });
});

import * as assert from 'node:assert';
import {
  Uri,
  type TextDocument,
  type WorkspaceFolder,
  type WorkspaceFoldersChangeEvent,
} from 'vscode';
import {
  WorkspaceRslintCoordinator,
  workspaceRootKey,
  type WorkspaceCoordinatorLogger,
  type WorkspaceRootRouter,
  type WorkspaceRuntime,
} from '../../src/WorkspaceRslintCoordinator';

type StartMode = 'ready' | 'fail' | 'pending' | 'factory-fail';

class FakeRuntime implements WorkspaceRuntime {
  readonly opened: string[] = [];
  readonly closedDocuments: string[] = [];
  closeCalls = 0;
  failClose = false;

  constructor(
    readonly workspaceFolder: WorkspaceFolder,
    readonly rootKey: string,
    private readonly startMode: StartMode,
  ) {}

  async start(signal: AbortSignal): Promise<void> {
    if (this.startMode === 'ready') return;
    if (this.startMode === 'fail') throw new Error(`failed ${this.rootKey}`);
    await new Promise<void>((resolve, reject) => {
      const onAbort = () => reject(signal.reason);
      signal.addEventListener('abort', onAbort, { once: true });
      void resolve;
    });
  }

  async close(): Promise<void> {
    this.closeCalls++;
    if (this.failClose) throw new Error(`close failed ${this.rootKey}`);
  }

  async sendDocumentOpen(document: TextDocument): Promise<void> {
    this.opened.push(document.uri.toString());
  }

  async sendDocumentClose(document: TextDocument): Promise<void> {
    this.closedDocuments.push(document.uri.toString());
  }

  clearDocumentDiagnostics(): void {}
}

class FakeRouter implements WorkspaceRootRouter {
  readonly active = new Map<string, WorkspaceRuntime>();

  async activate(runtime: WorkspaceRuntime): Promise<void> {
    this.active.set(runtime.rootKey, runtime);
  }

  async deactivate(rootKey: string): Promise<void> {
    this.active.delete(rootKey);
  }

  async closeAll(): Promise<void> {
    this.active.clear();
  }
}

const silentLogger: WorkspaceCoordinatorLogger = {
  debug() {},
  info() {},
  warn() {},
  error() {},
};

function folder(fsPath: string, name = 'app', index = 0): WorkspaceFolder {
  return { uri: Uri.file(fsPath), name, index };
}

function changeEvent(
  added: readonly WorkspaceFolder[],
  removed: readonly WorkspaceFolder[],
): WorkspaceFoldersChangeEvent {
  return { added, removed };
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

function coordinatorHarness(
  modeFor: (rootKey: string) => StartMode = () => 'ready',
) {
  const router = new FakeRouter();
  const runtimes: FakeRuntime[] = [];
  const coordinator = new WorkspaceRslintCoordinator(
    router,
    (workspaceFolder, rootKey) => {
      const mode = modeFor(rootKey);
      if (mode === 'factory-fail') {
        throw new Error(`factory failed ${rootKey}`);
      }
      const runtime = new FakeRuntime(workspaceFolder, rootKey, mode);
      runtimes.push(runtime);
      return runtime;
    },
    silentLogger,
  );
  return { coordinator, router, runtimes };
}

suite('workspace runtime coordinator', () => {
  test('uses URI identity for same-name roots', async () => {
    const first = folder('/workspace/first/app', 'app', 0);
    const second = folder('/workspace/second/app', 'app', 1);
    const { coordinator, router, runtimes } = coordinatorHarness();

    await coordinator.initialize([first, second]);
    await eventually(
      () => router.active.size === 2,
      'both same-name roots should become active',
    );

    assert.notStrictEqual(workspaceRootKey(first), workspaceRootKey(second));
    assert.strictEqual(runtimes.length, 2);
    await coordinator.close();
  });

  test('isolates initial root failures', async () => {
    const broken = folder('/workspace/broken', 'broken', 0);
    const healthy = folder('/workspace/healthy', 'healthy', 1);
    const { coordinator, router } = coordinatorHarness((key) =>
      key === workspaceRootKey(broken) ? 'fail' : 'ready',
    );

    await coordinator.initialize([broken, healthy]);
    await eventually(
      () => router.active.has(workspaceRootKey(healthy)),
      'healthy root should remain active',
    );
    assert.strictEqual(router.active.has(workspaceRootKey(broken)), false);
    await coordinator.close();
  });

  test('isolates runtime factory failures', async () => {
    const broken = folder('/workspace/broken', 'broken', 0);
    const healthy = folder('/workspace/healthy', 'healthy', 1);
    const { coordinator, router } = coordinatorHarness((key) =>
      key === workspaceRootKey(broken) ? 'factory-fail' : 'ready',
    );

    await coordinator.initialize([broken, healthy]);
    await eventually(
      () => router.active.has(workspaceRootKey(healthy)),
      'healthy root should survive a sibling factory failure',
    );
    await coordinator.close();
  });

  test('does not let a pending root block another root or removal', async () => {
    const pending = folder('/workspace/pending', 'pending', 0);
    const healthy = folder('/workspace/healthy', 'healthy', 1);
    const { coordinator, router, runtimes } = coordinatorHarness((key) =>
      key === workspaceRootKey(pending) ? 'pending' : 'ready',
    );

    await coordinator.initialize([pending, healthy]);
    coordinator.handleWorkspaceFoldersChanged(changeEvent([], [pending]), [
      healthy,
    ]);
    await eventually(
      () =>
        runtimes.find(
          (runtime) => runtime.rootKey === workspaceRootKey(pending),
        )?.closeCalls === 1,
      'removed pending root should close',
    );
    assert.strictEqual(router.active.has(workspaceRootKey(healthy)), true);
    await coordinator.close();
  });

  test('follows a topology replacement while activation is still pending', async () => {
    const pending = folder('/workspace/pending', 'pending', 0);
    const replacement = folder('/workspace/replacement', 'replacement', 0);
    const { coordinator, router } = coordinatorHarness((key) =>
      key === workspaceRootKey(pending) ? 'pending' : 'ready',
    );

    const initializing = coordinator.initialize([pending]);
    coordinator.handleWorkspaceFoldersChanged(
      changeEvent([replacement], [pending]),
      [replacement],
    );
    await initializing;

    assert.strictEqual(router.active.has(workspaceRootKey(pending)), false);
    assert.strictEqual(router.active.has(workspaceRootKey(replacement)), true);
    await coordinator.close();
  });

  test('lets an added healthy root unblock a pending initial root', async () => {
    const pending = folder('/workspace/pending', 'pending', 0);
    const healthy = folder('/workspace/healthy', 'healthy', 1);
    const { coordinator, router } = coordinatorHarness((key) =>
      key === workspaceRootKey(pending) ? 'pending' : 'ready',
    );

    const initializing = coordinator.initialize([pending]);
    coordinator.handleWorkspaceFoldersChanged(changeEvent([healthy], []), [
      pending,
      healthy,
    ]);
    await initializing;

    assert.strictEqual(router.active.has(workspaceRootKey(healthy)), true);
    await coordinator.close();
  });

  test('replaces a renamed root even when its URI is unchanged', async () => {
    const original = folder('/workspace/app', 'old-name', 0);
    const renamed = folder('/workspace/app', 'new-name', 0);
    const { coordinator, router, runtimes } = coordinatorHarness();
    await coordinator.initialize([original]);

    coordinator.handleWorkspaceFoldersChanged(
      changeEvent([renamed], [original]),
      [renamed],
    );
    await eventually(
      () => runtimes.length === 2 && runtimes[0].closeCalls === 1,
      'rename should replace and close the old runtime',
    );
    assert.strictEqual(
      router.active.get(workspaceRootKey(renamed))?.workspaceFolder.name,
      'new-name',
    );
    await coordinator.close();
  });

  test('quarantines a close-failed runtime instead of overlapping its replacement', async () => {
    const original = folder('/workspace/app', 'old-name', 0);
    const renamed = folder('/workspace/app', 'new-name', 0);
    const { coordinator, router, runtimes } = coordinatorHarness();
    await coordinator.initialize([original]);
    runtimes[0].failClose = true;

    coordinator.handleWorkspaceFoldersChanged(
      changeEvent([renamed], [original]),
      [renamed],
    );
    await eventually(
      () => runtimes[0].closeCalls === 1,
      'replacement should attempt to close the old runtime',
    );

    assert.strictEqual(runtimes.length, 1, 'replacement must not overlap');
    assert.strictEqual(router.active.has(workspaceRootKey(original)), false);
    await assert.rejects(
      coordinator.close(),
      /failed to close workspace coordinator/,
    );
    assert.strictEqual(
      runtimes[0].closeCalls,
      2,
      'terminal close should retry',
    );
  });

  test('does not restart a root when only its positional index changes', async () => {
    const original = folder('/workspace/app', 'app', 0);
    const shifted = folder('/workspace/app', 'app', 1);
    const inserted = folder('/workspace/inserted', 'inserted', 0);
    const { coordinator, router, runtimes } = coordinatorHarness();
    await coordinator.initialize([original]);

    coordinator.handleWorkspaceFoldersChanged(changeEvent([inserted], []), [
      inserted,
      shifted,
    ]);
    await eventually(
      () => router.active.size === 2,
      'new root should become active',
    );
    assert.strictEqual(
      runtimes.filter(
        (runtime) => runtime.rootKey === workspaceRootKey(original),
      ).length,
      1,
    );
    await coordinator.close();
  });

  test('rejects activation when every root fails', async () => {
    const first = folder('/workspace/first', 'first', 0);
    const second = folder('/workspace/second', 'second', 1);
    const { coordinator } = coordinatorHarness(() => 'fail');

    await assert.rejects(
      coordinator.initialize([first, second]),
      /All Rslint workspace roots failed/,
    );
    await coordinator.close();
  });

  test('closes every root and reports terminal close failures', async () => {
    const first = folder('/workspace/first', 'first', 0);
    const second = folder('/workspace/second', 'second', 1);
    const { coordinator, router, runtimes } = coordinatorHarness();
    await coordinator.initialize([first, second]);
    await eventually(
      () => router.active.size === 2,
      'both roots should become active',
    );
    runtimes[0].failClose = true;

    await assert.rejects(
      coordinator.close(),
      /failed to close workspace coordinator/,
    );
    assert.deepStrictEqual(
      runtimes.map((runtime) => runtime.closeCalls),
      [1, 1],
    );
  });
});

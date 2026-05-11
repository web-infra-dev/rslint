/**
 * Unit tests for the workspace-folder add/remove handler that
 * `Extension.activate()` wires up to `workspace.onDidChangeWorkspaceFolders`.
 *
 * These tests run via rstest (no VS Code runtime required). The
 * end-to-end path is implicit: the helper signature matches the event
 * the VS Code API delivers, and the mocks here stand in for what the
 * real Extension binds to (`createRslintInstance` / `removeRslintInstance`
 * on the Extension class, both of which are already exercised by the
 * vscode-test integration suites for the single-folder happy path).
 */
import { describe, test, expect, beforeEach } from '@rstest/core';

// Tiny hand-rolled spy (rstest doesn't ship a `vi.fn` equivalent;
// existing unit/compat-pool.test.ts pattern is to write
// instrumented stubs inline).
function spy<TArgs extends unknown[], TReturn>(
  impl: (...args: TArgs) => TReturn,
): ((...args: TArgs) => TReturn) & { calls: TArgs[] } {
  const calls: TArgs[] = [];
  const fn = ((...args: TArgs) => {
    calls.push(args);
    return impl(...args);
  }) as ((...args: TArgs) => TReturn) & { calls: TArgs[] };
  fn.calls = calls;
  return fn;
}
import type { WorkspaceFolder, Uri } from 'vscode';

import {
  applyFolderChange,
  folderKey,
  type RslintInstanceManager,
  type FolderChangeLogger,
} from '../src/workspace-folders';

function mkFolder(name: string, fsPath = `/tmp/${name}`): WorkspaceFolder {
  return {
    name,
    index: 0,
    uri: {
      fsPath,
      scheme: 'file',
      path: fsPath,
      authority: '',
      query: '',
      fragment: '',
      with: () => null as unknown as Uri,
      toJSON: () => null,
      toString: () => `file://${fsPath}`,
    } as Uri,
  };
}

function mkManager(): RslintInstanceManager & {
  created: WorkspaceFolder[];
  removed: string[];
  existing: Set<string>;
} {
  const existing = new Set<string>();
  const created: WorkspaceFolder[] = [];
  const removed: string[] = [];
  return {
    existing,
    created,
    removed,
    has(id: string) {
      return existing.has(id);
    },
    async create(folder: WorkspaceFolder) {
      created.push(folder);
      existing.add(folderKey(folder));
    },
    async remove(id: string) {
      removed.push(id);
      existing.delete(id);
    },
  };
}

function mkLogger(): FolderChangeLogger & {
  warns: string[];
  errors: Array<{ msg: string; err: unknown }>;
} {
  const warns: string[] = [];
  const errors: Array<{ msg: string; err: unknown }> = [];
  return {
    warns,
    errors,
    warn: (m) => warns.push(m),
    error: (m, e) => errors.push({ msg: m, err: e }),
  };
}

describe('applyFolderChange', () => {
  let mgr: ReturnType<typeof mkManager>;
  let log: ReturnType<typeof mkLogger>;

  beforeEach(() => {
    mgr = mkManager();
    log = mkLogger();
  });

  test('added folder triggers manager.create with the WorkspaceFolder', async () => {
    const folder = mkFolder('proj-a');
    await applyFolderChange({ added: [folder], removed: [] }, mgr, log);
    expect(mgr.created).toHaveLength(1);
    expect(mgr.created[0].name).toBe('proj-a');
    expect(mgr.removed).toHaveLength(0);
    expect(log.errors).toHaveLength(0);
  });

  test('removed folder triggers manager.remove by id', async () => {
    mgr.existing.add(folderKey(mkFolder('proj-a')));
    await applyFolderChange(
      { added: [], removed: [mkFolder('proj-a')] },
      mgr,
      log,
    );
    expect(mgr.removed).toEqual([folderKey(mkFolder('proj-a'))]);
    expect(mgr.created).toHaveLength(0);
    expect(log.errors).toHaveLength(0);
  });

  test('mixed event processes removals BEFORE additions', async () => {
    // Same-name remove + add in one event (e.g. user re-keyed a
    // folder). The remove must finish before create runs so the
    // create path doesn't trip the "instance already exists" guard.
    mgr.existing.add(folderKey(mkFolder('shared')));
    const trace: string[] = [];
    const tracedMgr: RslintInstanceManager = {
      has: (id) => mgr.has(id),
      async create(f) {
        trace.push(`create:${folderKey(f)}`);
        await mgr.create(f);
      },
      async remove(id) {
        trace.push(`remove:${id}`);
        await mgr.remove(id);
      },
    };
    await applyFolderChange(
      { added: [mkFolder('shared')], removed: [mkFolder('shared')] },
      tracedMgr,
      log,
    );
    const sharedKey = folderKey(mkFolder('shared'));
    expect(trace).toEqual([`remove:${sharedKey}`, `create:${sharedKey}`]);
    expect(log.errors).toHaveLength(0);
  });

  test('duplicate add (instance already exists) logs warn, skips create', async () => {
    mgr.existing.add(folderKey(mkFolder('proj-a')));
    await applyFolderChange(
      { added: [mkFolder('proj-a')], removed: [] },
      mgr,
      log,
    );
    expect(mgr.created).toHaveLength(0);
    expect(log.warns).toHaveLength(1);
    expect(log.warns[0]).toContain('proj-a');
    expect(log.warns[0]).toMatch(/already exists/);
  });

  test('create() error is caught and logged; subsequent adds still run', async () => {
    const createSpy = spy(async (f: WorkspaceFolder) => {
      if (f.name === 'bad') throw new Error('boom');
      mgr.created.push(f);
    });
    const failing: RslintInstanceManager = {
      has: () => false,
      create: createSpy,
      remove: async () => {},
    };
    await applyFolderChange(
      {
        added: [mkFolder('bad'), mkFolder('good')],
        removed: [],
      },
      failing,
      log,
    );
    // Both adds were attempted.
    expect(createSpy.calls).toHaveLength(2);
    // Good one landed in the underlying manager.
    expect(mgr.created.map((f) => f.name)).toEqual(['good']);
    // The bad one surfaced as a single error log; not a thrown error.
    expect(log.errors).toHaveLength(1);
    expect(log.errors[0].msg).toContain('bad');
    expect((log.errors[0].err as Error).message).toBe('boom');
  });

  test('remove() error is caught; subsequent removes + adds still run', async () => {
    const removeSpy = spy(async (id: string) => {
      if (id === folderKey(mkFolder('bad'))) throw new Error('boom-remove');
      mgr.removed.push(id);
    });
    const failing: RslintInstanceManager = {
      has: () => false,
      create: async (f) => {
        mgr.created.push(f);
      },
      remove: removeSpy,
    };
    await applyFolderChange(
      {
        added: [mkFolder('proj-c')],
        removed: [mkFolder('bad'), mkFolder('proj-b')],
      },
      failing,
      log,
    );
    expect(removeSpy.calls).toHaveLength(2);
    expect(mgr.removed).toEqual([folderKey(mkFolder('proj-b'))]);
    expect(mgr.created.map((f) => f.name)).toEqual(['proj-c']);
    expect(log.errors).toHaveLength(1);
    expect(log.errors[0].msg).toContain('bad');
  });

  test('empty event is a noop (no manager calls, no logs)', async () => {
    const hasSpy = spy((_id: string) => false);
    const createSpy = spy(async (_f: WorkspaceFolder) => {});
    const removeSpy = spy(async (_id: string) => {});
    const noop: RslintInstanceManager = {
      has: hasSpy,
      create: createSpy,
      remove: removeSpy,
    };
    await applyFolderChange({ added: [], removed: [] }, noop, log);
    expect(hasSpy.calls).toHaveLength(0);
    expect(createSpy.calls).toHaveLength(0);
    expect(removeSpy.calls).toHaveLength(0);
    expect(log.warns).toHaveLength(0);
    expect(log.errors).toHaveLength(0);
  });
});

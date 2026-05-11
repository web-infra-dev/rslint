import { describe, test, expect } from '@rstest/core';
import {
  applyFolderChange,
  folderKey,
  type RslintInstanceManager,
  type WorkspaceFoldersChangeEventLike,
} from '../src/workspace-folders';

// #7 regression: folders were keyed by `folder.name` (the basename),
// which collides across folders sharing a last path segment (e.g. two
// `pkg` dirs) — the second was skipped and never linted. Keys must be the
// globally unique folder URI.
describe('workspace folder keying (#7)', () => {
  const fakeFolder = (name: string, uri: string) =>
    ({ name, uri: { toString: () => uri } }) as never;
  const noopLogger = { warn: () => {}, error: () => {} };

  test('folderKey uses the unique URI, not the basename', () => {
    expect(folderKey(fakeFolder('pkg', 'file:///a/pkg'))).toBe('file:///a/pkg');
  });

  test('applyFolderChange keys has()/remove() by URI, not name', async () => {
    const hasKeys: string[] = [];
    const removeKeys: string[] = [];
    const manager: RslintInstanceManager = {
      has: (id) => {
        hasKeys.push(id);
        return false;
      },
      create: async () => {},
      remove: async (id) => {
        removeKeys.push(id);
      },
    };
    const event: WorkspaceFoldersChangeEventLike = {
      added: [fakeFolder('pkg', 'file:///a/pkg')],
      removed: [fakeFolder('old', 'file:///x/old')],
    } as never;

    await applyFolderChange(event, manager, noopLogger);

    expect(hasKeys).toEqual(['file:///a/pkg']); // URI, not 'pkg'
    expect(removeKeys).toEqual(['file:///x/old']); // URI, not 'old'
  });
});

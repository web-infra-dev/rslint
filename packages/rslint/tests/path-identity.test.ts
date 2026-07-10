import { describe, expect, test } from '@rstest/core';
import path from 'node:path';

import {
  AncestorPathIndex,
  createCachedAncestorFinder,
  createPathIdentity,
} from '../src/api/path-identity.js';

describe('path identity', () => {
  test('uses case-sensitive POSIX identity', () => {
    const identity = createPathIdentity(path.posix, true);

    expect(identity.equals('/repo/App', '/repo/app')).toBe(false);
    expect(identity.isSameOrChild('/repo/App', '/repo/app/file.ts')).toBe(
      false,
    );
    expect(identity.isSameOrChild('/repo/App', '/repo/App/file.ts')).toBe(true);

    const caseInsensitive = createPathIdentity(path.posix, false);
    expect(
      caseInsensitive.isSameOrChild('/Repo/App', '/repo/app/file.ts'),
    ).toBe(true);
  });

  test('normalizes case, separators, and drive letters for Windows paths', () => {
    const identity = createPathIdentity(path.win32, false);

    expect(identity.equals('C:/Repo/App/', 'c:\\repo\\app')).toBe(true);
    expect(identity.isSameOrChild('C:\\Repo', 'c:/REPO/src/file.ts')).toBe(
      true,
    );
    expect(identity.isSameOrChild('C:\\Repo', 'D:\\Repo\\file.ts')).toBe(false);

    const index = new AncestorPathIndex(
      [
        ['C:\\Repo', 'root'],
        ['C:\\Repo\\Packages\\App', 'app'],
      ],
      identity,
    );
    expect(index.find('c:/repo/packages/APP/src')).toBe('app');
    expect(index.find('C:\\REPO\\other')).toBe('root');
    expect(index.find('D:\\Repo')).toBeUndefined();
  });

  test('routes UNC paths with Windows case semantics', () => {
    const identity = createPathIdentity(path.win32, false);
    const index = new AncestorPathIndex(
      [
        ['\\\\Server\\Share\\Repo', 'root'],
        ['\\\\server\\share\\repo\\packages\\app', 'app'],
      ],
      identity,
    );

    expect(index.find('\\\\SERVER\\SHARE\\Repo\\Packages\\App\\src')).toBe(
      'app',
    );
    expect(index.find('\\\\server\\other\\repo')).toBeUndefined();
  });

  test('caches shared ancestor probes by directory', () => {
    const identity = createPathIdentity(path.posix, true);
    const probes: string[] = [];
    const find = createCachedAncestorFinder((directory) => {
      probes.push(directory);
      return directory === '/repo' ? '/repo/rslint.config.mjs' : undefined;
    }, identity);

    expect(find('/repo/packages/app/src')).toBe('/repo/rslint.config.mjs');
    expect(find('/repo/packages/app/test')).toBe('/repo/rslint.config.mjs');
    expect(probes).toEqual([
      '/repo/packages/app/src',
      '/repo/packages/app',
      '/repo/packages',
      '/repo',
      '/repo/packages/app/test',
    ]);
  });
});

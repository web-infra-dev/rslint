import { describe, expect, test } from '@rstest/core';
import path from 'node:path';

import { createPathIdentity } from '../src/api/path-identity.js';

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
    const caseSensitive = createPathIdentity(path.win32, true);
    expect(caseSensitive.equals('C:\\Repo', 'C:\\repo')).toBe(false);
    expect(
      caseSensitive.isSameOrChild('C:\\Repo', 'C:\\repo\\src\\file.ts'),
    ).toBe(false);
  });

  test('routes UNC paths with Windows case semantics', () => {
    const identity = createPathIdentity(path.win32, false);

    expect(
      identity.isSameOrChild(
        '\\\\server\\share\\repo\\packages\\app',
        '\\\\SERVER\\SHARE\\Repo\\Packages\\App\\src',
      ),
    ).toBe(true);
    expect(
      identity.isSameOrChild(
        '\\\\server\\share\\repo',
        '\\\\server\\other\\repo',
      ),
    ).toBe(false);
  });
});

import path from 'node:path';

import type { InvalidTestCase, ValidTestCase } from './rule-tester.js';

// Port from https://github.com/import-js/eslint-plugin-import/blob/01c9eb04331d2efa8d63f2d7f4bfec3bc44c94f3/tests/src/utils.js

export function test<T extends ValidTestCase | InvalidTestCase>(t: T): T {
  return t;
}

export function testFilePath(relativePath: string): string {
  return path.join(import.meta.dirname, 'files', relativePath);
}

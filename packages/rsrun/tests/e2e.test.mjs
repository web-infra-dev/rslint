import { test } from 'node:test';
import assert from 'node:assert/strict';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { spawnSync } from 'node:child_process';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const registerPath = path.resolve(__dirname, '..', 'dist', 'register.js');
const fixtureOkDir = path.resolve(__dirname, 'fixtures', 'ok');
const fixtureErrorDir = path.resolve(__dirname, 'fixtures', 'error');
const fixtureOk = path.resolve(fixtureOkDir, 'index.ts');
const fixtureError = path.resolve(fixtureErrorDir, 'index.ts');

function runRsrun(entry, cwd) {
  return spawnSync('node', ['-r', registerPath, entry], {
    cwd,
    env: {
      ...process.env,
      RSRUN_TYPECHECK: 'true',
    },
    encoding: 'utf8',
  });
}

test('rsrun typecheck executes fixture', () => {
  const result = runRsrun(fixtureOk, fixtureOkDir);
  assert.equal(result.status, 0);
  assert.match(result.stdout ?? '', /rsrun fixture/);
});

test('rsrun typecheck reports errors', () => {
  const result = runRsrun(fixtureError, fixtureErrorDir);
  assert.equal(result.status, 1);
  assert.match(result.stderr ?? '', /TS\d+/);
});

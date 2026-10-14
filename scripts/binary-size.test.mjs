import assert from 'node:assert/strict';
import { execFileSync } from 'node:child_process';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const script = fileURLToPath(new URL('./binary-size.mjs', import.meta.url));

function temporaryDirectory(t) {
  const directory = fs.mkdtempSync(
    path.join(os.tmpdir(), 'rslint-binary-size-'),
  );
  t.after(() => fs.rmSync(directory, { recursive: true, force: true }));
  return directory;
}

function run(args, cwd, env = {}) {
  return execFileSync(process.execPath, [script, ...args], {
    cwd,
    encoding: 'utf8',
    env: { ...process.env, ...env },
  });
}

function writeJson(file, value) {
  fs.writeFileSync(file, `${JSON.stringify(value, null, 2)}\n`);
}

test('record stores the executable, native addon, and compatible CLI size', (t) => {
  const directory = temporaryDirectory(t);
  const binary = path.join(directory, 'rslint');
  const nativeAddon = path.join(directory, 'rslint.node');
  const output = path.join(directory, 'measurement.json');
  fs.writeFileSync(binary, Buffer.alloc(2048));
  fs.writeFileSync(nativeAddon, Buffer.alloc(512));

  const stdout = run(['record', binary, nativeAddon, output], directory, {
    BINARY_SIZE_GO_VERSION: '1.27.0',
    GITHUB_SHA: 'head-sha',
  });
  const measurement = JSON.parse(fs.readFileSync(output, 'utf8'));

  assert.equal(measurement.schemaVersion, 2);
  assert.equal(measurement.size, 2048);
  assert.equal(measurement.nativeAddonSize, 512);
  assert.equal(measurement.goVersion, '1.27.0');
  assert.match(stdout, /rslint: 2048 bytes \(2\.00 KiB\)/);
  assert.match(stdout, /rslint\.node: 512 bytes \(0\.50 KiB\)/);
});

test('report keeps the CLI comparison when the base predates addon data', (t) => {
  const directory = temporaryDirectory(t);
  const headPath = path.join(directory, 'head.json');
  const basePath = path.join(directory, 'base.json');
  writeJson(headPath, {
    schemaVersion: 2,
    sha: 'merge-sha',
    headSha: 'head-sha',
    runId: '123',
    target: 'linux-x64-gnu',
    goVersion: '1.27.0',
    size: 2 * 1024 * 1024,
    nativeAddonSize: 1024 * 1024,
    measuredAt: '2026-10-14T04:03:52.689Z',
  });
  writeJson(basePath, {
    schemaVersion: 1,
    sha: 'base-sha',
    size: 1024 * 1024,
  });
  writeJson(path.join(directory, 'base-commit.json'), {
    sha: 'base-sha',
  });

  const report = run(['report', headPath, basePath], directory, {
    GITHUB_REPOSITORY: 'web-infra-dev/rslint',
    GITHUB_SERVER_URL: 'https://github.com',
  });

  assert.match(
    report,
    /\| `rslint` \(linux-x64-gnu\) \| 1\.00 MiB \| 2\.00 MiB \| \+1\.00 MiB \(\+100\.00%\) \|/,
  );
  assert.match(
    report,
    /\| `rslint\.node` \(linux-x64-gnu\) \| — \| 1\.00 MiB \| — \|/,
  );
  assert.match(report, /base artifact predates native-addon measurement/);
});

test('report compares the executable and native addon independently', (t) => {
  const directory = temporaryDirectory(t);
  const headPath = path.join(directory, 'head.json');
  const basePath = path.join(directory, 'base.json');
  writeJson(headPath, {
    schemaVersion: 2,
    sha: 'merge-sha',
    headSha: 'head-sha',
    target: 'linux-x64-gnu',
    size: 2 * 1024 * 1024,
    nativeAddonSize: 1024 * 1024,
  });
  writeJson(basePath, {
    schemaVersion: 2,
    sha: 'base-sha',
    size: 1024 * 1024,
    nativeAddonSize: 512 * 1024,
  });
  writeJson(path.join(directory, 'base-commit.json'), {
    sha: 'base-sha',
  });

  const report = run(['report', headPath, basePath], directory);

  assert.match(
    report,
    /\| `rslint\.node` \(linux-x64-gnu\) \| 512\.00 KiB \| 1\.00 MiB \| \+512\.00 KiB \(\+100\.00%\) \|/,
  );
  assert.doesNotMatch(report, /base artifact predates/);
});

import assert from 'node:assert/strict';
import { spawnSync } from 'node:child_process';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { test } from 'node:test';
import { fileURLToPath, pathToFileURL } from 'node:url';

const root = fileURLToPath(new URL('../', import.meta.url));
const binarySizeScript = path.join(root, 'scripts/binary-size.mjs');
const moveArtifactsScript = path.join(
  root,
  '.github/actions/move-artifacts/move-artifacts.js',
);
const tuples = [
  'darwin-arm64',
  'darwin-x64',
  'linux-arm64-gnu',
  'linux-arm64-musl',
  'linux-x64-gnu',
  'linux-x64-musl',
  'win32-arm64-msvc',
  'win32-x64-msvc',
];

function temporaryDirectory(t) {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), 'rslint-ci-delivery-'));
  t.after(() => fs.rmSync(dir, { recursive: true, force: true }));
  return dir;
}

function run(script, dir, args = []) {
  return spawnSync(process.execPath, [script, ...args], {
    cwd: dir,
    encoding: 'utf8',
    env: {
      ...process.env,
      GITHUB_SHA: 'head-commit',
      GITHUB_EVENT_PATH: path.join(dir, 'event.json'),
    },
  });
}

function renderReport(t, head, base) {
  const dir = temporaryDirectory(t);
  fs.writeFileSync(
    path.join(dir, 'base-commit.json'),
    JSON.stringify({ sha: 'base-commit' }),
  );
  fs.writeFileSync(path.join(dir, 'head.json'), JSON.stringify(head));
  fs.writeFileSync(
    path.join(dir, 'base.json'),
    JSON.stringify({ sha: 'base-commit', ...base }),
  );
  const result = run(binarySizeScript, dir, [
    'report',
    'head.json',
    'base.json',
  ]);
  assert.equal(result.status, 0, result.stderr);
  assert.doesNotMatch(result.stdout, /NaN|undefined/);
  return result.stdout;
}

test('record measures the only shipped native addon and rejects missing artifacts', (t) => {
  const dir = temporaryDirectory(t);
  const filename = 'rslint.linux-x64-gnu.node';
  fs.writeFileSync(path.join(dir, filename), Buffer.alloc(1024));
  const result = run(binarySizeScript, dir, ['record', filename, 'size.json']);
  assert.equal(result.status, 0, result.stderr);
  const measured = JSON.parse(fs.readFileSync(path.join(dir, 'size.json')));
  assert.equal(measured.schemaVersion, 4);
  assert.equal(measured.target, 'linux-x64-gnu');
  assert.equal(measured.nativeAddonSize, 1024);
  assert.equal(measured.totalSize, 1024);
  assert.equal('size' in measured, false);
  assert.equal('compilerServiceSize' in measured, false);
  assert.equal('goVersion' in measured, false);
  assert.match(result.stdout, /total native artifacts: 1024 bytes/);
  fs.unlinkSync(path.join(dir, filename));
  const missing = run(binarySizeScript, dir, [
    'record',
    filename,
    'missing.json',
  ]);
  assert.notEqual(missing.status, 0);
  assert.equal(fs.existsSync(path.join(dir, 'missing.json')), false);
  fs.writeFileSync(path.join(dir, filename), '');
  const empty = run(binarySizeScript, dir, ['record', filename, 'empty.json']);
  assert.notEqual(empty.status, 0);
  assert.equal(fs.existsSync(path.join(dir, 'empty.json')), false);
  fs.writeFileSync(path.join(dir, 'rslint'), 'legacy executable');
  const legacy = run(binarySizeScript, dir, [
    'record',
    'rslint',
    'legacy.json',
  ]);
  assert.notEqual(legacy.status, 0);
  assert.equal(fs.existsSync(path.join(dir, 'legacy.json')), false);
});

test('report compares native totals only for the same architecture and target', (t) => {
  const output = renderReport(
    t,
    { schemaVersion: 4, target: 'linux-x64-gnu', nativeAddonSize: 1024 },
    { schemaVersion: 4, target: 'linux-x64-gnu', nativeAddonSize: 2048 },
  );
  assert.match(
    output,
    /\| \*\*Total native artifacts\*\* \| 2.00 KiB \| 1.00 KiB \| -1.00 KiB \(-50.00%\) \|/,
  );
  for (const base of [
    { schemaVersion: 2, nativeAddonSize: 2048, size: 8192 },
    {
      schemaVersion: 3,
      nativeAddonSize: 2048,
      size: 8192,
      compilerServiceSize: 4096,
    },
    { schemaVersion: 4, target: 'darwin-arm64', nativeAddonSize: 2048 },
  ]) {
    const historical = renderReport(
      t,
      { schemaVersion: 4, target: 'linux-x64-gnu', nativeAddonSize: 1024 },
      base,
    );
    assert.match(historical, /complete native size is not comparable/);
    assert.match(
      historical,
      /\| \*\*Total native artifacts\*\* \| — \| 1.00 KiB \| — \|/,
    );
  }
});

test('historical head artifacts remain readable with explicit partial scope', (t) => {
  const output = renderReport(
    t,
    {
      schemaVersion: 2,
      target: 'linux-x64-gnu',
      nativeAddonSize: 1024,
      size: 8192,
    },
    { schemaVersion: 2, nativeAddonSize: 2048, size: 8192 },
  );
  assert.match(output, /does not establish its complete native size/);
  assert.doesNotMatch(output, /\| \*\*Total native artifacts\*\*/);
});

// Header-only fixtures exercise package assembly validation, not executable
// compatibility. Cross-platform builds and addon tests are separate CI jobs.
function packageArtifacts(dir) {
  for (const tuple of tuples) {
    const filename = `rslint.${tuple}.node`;
    const artifact = path.join(dir, 'binaries', tuple, filename);
    fs.mkdirSync(path.dirname(artifact), { recursive: true });
    const magic = tuple.startsWith('darwin-')
      ? 'cffaedfe'
      : tuple.startsWith('linux-')
        ? '7f454c46'
        : '4d5a0000';
    fs.writeFileSync(artifact, Buffer.from(magic, 'hex'));
    const destination = path.join(dir, 'npm/rslint', tuple);
    fs.mkdirSync(destination, { recursive: true });
    fs.copyFileSync(
      path.join(root, 'npm/rslint', tuple, 'package.json'),
      path.join(destination, 'package.json'),
    );
  }
}

test('release assembly validates and copies all eight Rust addon packages', (t) => {
  const dir = temporaryDirectory(t);
  packageArtifacts(dir);
  const result = run(moveArtifactsScript, dir);
  assert.equal(result.status, 0, result.stderr);
  for (const tuple of tuples) {
    const filename = `rslint.${tuple}.node`;
    assert.deepEqual(
      fs.readFileSync(path.join(dir, 'npm/rslint', tuple, filename)),
      fs.readFileSync(path.join(dir, 'binaries', tuple, filename)),
    );
    const metadata = JSON.parse(
      fs.readFileSync(path.join(dir, 'npm/rslint', tuple, 'package.json')),
    );
    assert.deepEqual(metadata.files, [filename]);
    assert.equal(metadata.bin, undefined);
    assert.equal(metadata.publishConfig?.executableFiles, undefined);
  }
});

test('release assembly rejects incomplete, duplicate, invalid and legacy artifacts before copying', (t) => {
  for (const failure of [
    'missing',
    'duplicate',
    'invalid',
    'unexpected',
    'legacy',
    'metadata',
  ]) {
    const dir = temporaryDirectory(t);
    packageArtifacts(dir);
    const tuple = tuples.at(-1);
    const filename = `rslint.${tuple}.node`;
    const artifact = path.join(dir, 'binaries', tuple, filename);
    const destination = path.join(dir, 'npm/rslint', tuple);
    if (failure === 'missing') fs.unlinkSync(artifact);
    if (failure === 'duplicate') {
      fs.mkdirSync(path.join(dir, 'binaries/duplicate'));
      fs.copyFileSync(artifact, path.join(dir, 'binaries/duplicate', filename));
    }
    if (failure === 'invalid') fs.writeFileSync(artifact, 'not a native addon');
    if (failure === 'unexpected')
      fs.writeFileSync(path.join(dir, 'binaries/rslint'), 'old executable');
    if (failure === 'legacy')
      fs.writeFileSync(
        path.join(destination, 'rslint-tsrs-service.exe'),
        'old service',
      );
    if (failure === 'metadata')
      fs.writeFileSync(
        path.join(destination, 'package.json'),
        JSON.stringify({ name: '@rslint/wrong', exports: {} }),
      );
    const result = run(moveArtifactsScript, dir);
    assert.notEqual(result.status, 0, failure);
    for (const candidate of tuples) {
      assert.equal(
        fs.existsSync(
          path.join(dir, 'npm/rslint', candidate, `rslint.${candidate}.node`),
        ),
        false,
        `${failure}: no partial output`,
      );
    }
  }
});

test('CI uses Rust and JS on three hosts without a Go or submodule setup path', () => {
  const workflow = fs.readFileSync(
    path.join(root, '.github/workflows/ci.yml'),
    'utf8',
  );
  assert.match(workflow, /os: \[ubuntu-22\.04, macos-latest, windows-latest\]/);
  for (const command of [
    'node scripts/tsrs/build-rust.mjs --test',
    'pnpm build',
    'node --test crates/rslint-native/tests/rust-lint.test.mjs',
    'pnpm --filter @rslint/core test:rust',
  ])
    assert.ok(workflow.includes(command), command);
  assert.doesNotMatch(
    workflow,
    /setup-go|test-go|go test|go build|submodules:/,
  );
  assert.match(workflow, /needs: \[static, test-rust-js\]/);
});

test('release matrix builds all native targets and only assembles frozen Rust artifacts', () => {
  const workflow = fs.readFileSync(
    path.join(root, '.github/workflows/release.yml'),
    'utf8',
  );
  const matrixTuples = [...workflow.matchAll(/^\s+tuple: ([\w-]+)$/gm)].map(
    (match) => match[1],
  );
  assert.deepEqual([...matrixTuples].sort(), [...tuples].sort());
  assert.match(
    workflow,
    /node scripts\/tsrs\/build-rust\.mjs --platform \$\{\{ matrix\.tuple \}\}/,
  );
  assert.match(
    workflow,
    /ref: \$\{\{ needs\.resolve-release\.outputs\.sha \}\}/,
  );
  assert.match(workflow, /uses: \.\/\.github\/actions\/move-artifacts/);
  assert.match(workflow, /if-no-files-found: error/);
  assert.match(workflow, /Publishing is unsupported/);
  assert.doesNotMatch(
    workflow,
    /setup-go|build-host|build-service|go build|submodules:/,
  );
  assert.match(
    workflow,
    /pnpm --filter @rslint\/core --filter @rslint\/api --filter '@rslint\/native-\*' -r pack/,
  );
});

test('legacy package subpaths reject instead of bundling compiler or worker bridges', async () => {
  const metadata = JSON.parse(
    fs.readFileSync(path.join(root, 'packages/rslint/package.json')),
  );
  for (const subpath of ['./service', './internal', './eslint-plugin']) {
    assert.equal(metadata.exports[subpath].default, './unsupported.mjs');
    assert.equal(metadata.exports[subpath].types, './unsupported.d.ts');
  }
  const shim = await import(
    pathToFileURL(path.join(root, 'packages/rslint/unsupported.mjs')).href
  );
  const unsupported = { code: 'ERR_RSLINT_UNSUPPORTED_ENTRYPOINT' };
  for (const Constructor of [
    shim.RSLintService,
    shim.NodeRslintService,
    shim.WorkerPool,
  ]) {
    assert.throws(() => new Constructor(), unsupported);
  }
  await assert.rejects(shim.lint(), unsupported);
  for (const operation of [
    shim.createPluginLintHost,
    shim.lintFile,
    shim.readAttachmentBytes,
  ]) {
    assert.throws(operation, unsupported);
  }
  const buildConfig = fs.readFileSync(
    path.join(root, 'packages/rslint/rstack.config.ts'),
    'utf8',
  );
  assert.match(buildConfig, /service: '\.\/unsupported\.mjs'/);
  assert.match(buildConfig, /internal: '\.\/unsupported\.mjs'/);
  assert.doesNotMatch(
    buildConfig,
    /src\/service|src\/internal|src\/eslint-plugin|setup-worker-entry/,
  );
});

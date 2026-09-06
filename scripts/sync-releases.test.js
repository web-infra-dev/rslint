// cspell:ignore NOSYSTEM gpgsign deinit
const assert = require('node:assert/strict');
const { execFileSync } = require('node:child_process');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const { test } = require('node:test');

const env = {
  ...process.env,
  GIT_CONFIG_GLOBAL: os.devNull,
  GIT_CONFIG_NOSYSTEM: '1',
};

function git(root, ...args) {
  return execFileSync('git', args, {
    cwd: root,
    encoding: 'utf8',
    stdio: ['ignore', 'pipe', 'pipe'],
    env,
  }).trim();
}

function write(root, file, value) {
  const target = path.join(root, file);
  fs.mkdirSync(path.dirname(target), { recursive: true });
  fs.writeFileSync(
    target,
    typeof value === 'string' ? value : JSON.stringify(value),
  );
}

function init(root) {
  fs.mkdirSync(root, { recursive: true });
  git(root, 'init', '--quiet');
  git(root, 'config', 'user.name', 'Release Test');
  git(root, 'config', 'user.email', 'test@example.com');
}

function commit(root) {
  git(root, 'add', '.');
  git(root, '-c', 'commit.gpgsign=false', 'commit', '--quiet', '-m', 'fixture');
  return git(root, 'rev-parse', 'HEAD');
}

function tag(root, name, ...args) {
  git(root, '-c', 'tag.gpgsign=false', 'tag', name, ...args);
}

function fixture(t) {
  const directory = fs.mkdtempSync(
    path.join(os.tmpdir(), 'rslint-release-sync-'),
  );
  t.after(() => fs.rmSync(directory, { recursive: true, force: true }));
  const upstream = path.join(directory, 'upstream');
  init(upstream);
  write(upstream, 'compiler.go', 'package compiler\n');
  const released = commit(upstream);
  tag(upstream, 'v7.0.2', '-a', '-m', 'release');
  write(upstream, 'next.go', 'package compiler\n');
  const latest = commit(upstream);

  const root = path.join(directory, 'rslint');
  init(root);
  git(
    root,
    '-c',
    'protocol.file.allow=always',
    'submodule',
    'add',
    '--quiet',
    upstream,
    'typescript-go',
  );
  git(
    root,
    'config',
    `url.${upstream}.insteadOf`,
    'https://github.com/microsoft/TypeScript.git',
  );
  write(root, 'internal/rules/old/old.go', 'package old\n');
  write(
    root,
    'internal/plugins/typescript/plugin.go',
    'const PLUGIN_NAME = "@typescript-eslint"\n',
  );
  write(
    root,
    'internal/plugins/typescript/rules/old_ts/rule.go',
    'package old_ts\n',
  );
  write(root, 'package.json', { version: '0.9.1' });
  const original = [
    { version: '0.9.1', rules: ['@typescript-eslint:old-ts', 'eslint:old'] },
  ];
  const file = path.join(root, 'website/releases.json');
  write(root, 'website/releases.json', original);
  write(
    root,
    'scripts/sync-releases.js',
    fs.readFileSync(path.join(__dirname, 'sync-releases.js'), 'utf8'),
  );
  commit(root);
  tag(root, 'v0.9.1');
  return {
    root,
    upstream,
    released,
    latest,
    original,
    text: () => fs.readFileSync(file, 'utf8'),
    read: () => JSON.parse(fs.readFileSync(file, 'utf8')),
    version: (version) => write(root, 'package.json', { version }),
    run: (...args) =>
      execFileSync(process.execPath, ['scripts/sync-releases.js', ...args], {
        cwd: root,
        env,
        encoding: 'utf8',
        stdio: ['ignore', 'pipe', 'pipe'],
      }),
    pin: (revision) => {
      git(
        root,
        '-C',
        'typescript-go',
        'checkout',
        '--quiet',
        '--detach',
        revision,
      );
      git(root, 'add', 'typescript-go');
    },
  };
}

test('main records upcoming rules and the current compiler without backfilling old releases', (t) => {
  const f = fixture(t);
  write(f.root, 'internal/rules/new/new.go', 'package added\n');
  f.run();
  assert.deepEqual(f.read(), [
    ...f.original,
    {
      version: 'main',
      rules: ['eslint:new'],
      typescript: { commit: f.latest, releaseVersion: null },
    },
  ]);
  const first = f.text();
  f.run();
  assert.equal(f.text(), first);
});

test('a new stable version records rules and compiler together, then stays fixed after publishing', (t) => {
  const f = fixture(t);
  write(f.root, 'internal/rules/new/new.go', 'package added\n');
  f.run();
  f.version('0.9.2');
  f.run();
  const binding = { commit: f.latest, releaseVersion: null };
  assert.deepEqual(f.read(), [
    ...f.original,
    { version: '0.9.2', rules: ['eslint:new'], typescript: binding },
    { version: 'main', rules: [], typescript: binding },
  ]);
  write(
    f.root,
    'internal/plugins/typescript/rules/new_ts/rule.go',
    'package new_ts\n',
  );
  f.run();
  assert.deepEqual(f.read()[1].rules, [
    '@typescript-eslint:new-ts',
    'eslint:new',
  ]);
  const published = f.read().slice(0, -1);
  commit(f.root);
  tag(f.root, 'v0.9.2');
  f.pin(f.released);
  f.run();
  assert.deepEqual(f.read().slice(0, -1), published);
  assert.deepEqual(f.read().at(-1).typescript, {
    commit: f.released,
    releaseVersion: '7.0.2',
  });
});

test('release lookup resolves annotated and lightweight upstream tags without local tag history', (t) => {
  const f = fixture(t);
  f.pin(f.released);
  git(f.root, '-C', 'typescript-go', 'tag', '-d', 'v7.0.2');
  f.run();
  assert.equal(f.read().at(-1).typescript.releaseVersion, '7.0.2');
  f.pin(f.latest);
  tag(f.upstream, 'v7.1.0-rc.1');
  f.run();
  assert.equal(f.read().at(-1).typescript.releaseVersion, '7.1.0-rc.1');
  tag(f.upstream, 'v7.1.0');
  f.run();
  assert.equal(f.read().at(-1).typescript.releaseVersion, '7.1.0');
});

test('an upstream failure does not write either main or a new release', (t) => {
  const f = fixture(t);
  const before = f.text();
  fs.renameSync(f.upstream, `${f.upstream}-offline`);
  assert.throws(() => f.run(), /Release sync failed/);
  assert.equal(f.text(), before);
  f.version('0.9.2');
  assert.throws(() => f.run(), /Release sync failed/);
  assert.equal(f.text(), before);
});

test('an unstaged compiler revision is rejected, while an uninitialized submodule uses its recorded pin', (t) => {
  const f = fixture(t);
  const before = f.text();
  git(
    f.root,
    '-C',
    'typescript-go',
    'checkout',
    '--quiet',
    '--detach',
    f.released,
  );
  assert.throws(() => f.run(), /Stage the TypeScript submodule revision/);
  assert.equal(f.text(), before);
  git(f.root, 'submodule', 'deinit', '--force', '--', 'typescript-go');
  f.run();
  assert.equal(f.read().at(-1).typescript.commit, f.latest);
});

test('full sync preserves main and release bindings without adding old compiler data', (t) => {
  const f = fixture(t);
  f.version('0.9.2');
  f.run();
  commit(f.root);
  tag(f.root, 'v0.9.2');
  const before = f.read();
  f.run('full');
  assert.deepEqual(f.read(), before);
  git(f.root, 'tag', '-d', 'v0.9.1');
  const text = f.text();
  assert.throws(() => f.run('full'), /Fetch all recorded stable release tags/);
  assert.equal(f.text(), text);
});

test('rule discovery excludes helpers and fixtures in the checkout and tagged history', (t) => {
  const f = fixture(t);
  for (const prefix of [
    'internal/rules',
    'internal/plugins/typescript/rules',
  ]) {
    write(f.root, `${prefix}/helpers/utils.go`, 'package helpers\n');
    write(f.root, `${prefix}/testdata/rule.go`, 'package fixture\n');
    write(f.root, `${prefix}/docs/README.md`, 'Rule documentation\n');
  }
  for (const plugin of ['jest', 'react']) {
    // Identical blobs without a declared name still need distinct fallbacks.
    write(f.root, `internal/plugins/${plugin}/plugin.go`, 'package plugin\n');
    write(
      f.root,
      `internal/plugins/${plugin}/rules/example/rule.go`,
      'package example\n',
    );
  }
  f.version('0.9.2');
  f.run();
  const expected = [
    'eslint-plugin-jest:example',
    'eslint-plugin-react:example',
  ];
  assert.deepEqual(f.read()[1].rules, expected);
  commit(f.root);
  tag(f.root, 'v0.9.2');
  f.run('full');
  assert.deepEqual(f.read()[1].rules, expected);
});

test('canary versions and unknown arguments retain the previous command restrictions', (t) => {
  const f = fixture(t);
  const before = f.text();
  f.version('0.9.2-canary.1');
  assert.throws(() => f.run(), /not a stable version/);
  assert.throws(() => f.run('--check'), /only supported argument/);
  assert.equal(f.text(), before);
});

test('missing local tags cannot cause an old release to acquire a new compiler binding', (t) => {
  const f = fixture(t);
  tag(f.root, 'v0.9.0');
  git(f.root, 'tag', '-d', 'v0.9.1');
  const before = f.text();
  assert.throws(
    () => f.run(),
    /Refusing to add a compiler binding to historical version/,
  );
  assert.equal(f.text(), before);
});

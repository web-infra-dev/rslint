const assert = require('node:assert/strict');
const { spawnSync } = require('node:child_process');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const { test } = require('node:test');

const repository = path.resolve(__dirname, '../../..');
const script = path.join(__dirname, 'move-artifacts.js');
const packageRoot = path.join(repository, 'npm/rslint');
const packages = fs
  .readdirSync(packageRoot, { withFileTypes: true })
  .filter((entry) => entry.isDirectory())
  .map((entry) => ({
    tuple: entry.name,
    manifest: JSON.parse(
      fs.readFileSync(path.join(packageRoot, entry.name, 'package.json')),
    ),
  }));

function fixture(t) {
  const root = fs.mkdtempSync(path.join(os.tmpdir(), 'rslint-artifacts-'));
  t.after(() => fs.rmSync(root, { recursive: true, force: true }));
  const files = new Map();
  const write = (relative, contents) => {
    const file = path.join(root, relative);
    fs.mkdirSync(path.dirname(file), { recursive: true });
    fs.writeFileSync(file, contents);
    return file;
  };
  for (const { tuple, manifest } of packages) {
    write(`npm/rslint/${tuple}/package.json`, JSON.stringify(manifest));
    const platform = `${manifest.os[0]}-${manifest.cpu[0]}`;
    const go = `${platform}-rslint`;
    const native = path.basename(manifest.exports['.']);
    files.set(go, write(`binaries/${go}/${go}`, `Go ${platform}`));
    files.set(
      native,
      write(`binaries/native-${tuple}/${native}`, `Native ${tuple}`),
    );
  }
  files.set(
    'rule-schemas.json',
    write('binaries/rule-schemas/rule-schemas.json', '{"rules":{}}'),
  );
  fs.mkdirSync(path.join(root, 'packages/rslint'), { recursive: true });
  return { root, files, write };
}

function run(root) {
  return spawnSync(process.execPath, [script], {
    cwd: root,
    encoding: 'utf8',
  });
}

function assertNoOutputs(root) {
  for (const { tuple, manifest } of packages) {
    for (const entry of ['.', './bin']) {
      assert.equal(
        fs.existsSync(
          path.join(root, 'npm/rslint', tuple, manifest.exports[entry]),
        ),
        false,
      );
    }
  }
}

test('assembles each platform package at its manifest exports', (t) => {
  const { root, files } = fixture(t);
  const result = run(root);
  assert.equal(result.status, 0, result.stderr);
  for (const { tuple, manifest } of packages) {
    const directory = path.join(root, 'npm/rslint', tuple);
    const go = path.join(directory, manifest.exports['./bin']);
    const native = path.join(directory, manifest.exports['.']);
    const platform = `${manifest.os[0]}-${manifest.cpu[0]}`;
    assert.equal(fs.readFileSync(go, 'utf8'), `Go ${platform}`);
    assert.equal(fs.readFileSync(native, 'utf8'), `Native ${tuple}`);
    if (process.platform !== 'win32') assert.ok(fs.statSync(go).mode & 0o111);
  }
  assert.deepEqual(
    fs.readFileSync(path.join(root, 'packages/rslint/rule-schemas.json')),
    fs.readFileSync(files.get('rule-schemas.json')),
  );
});

const required = [
  'linux-arm64-rslint',
  'rslint.linux-x64-musl.node',
  'rule-schemas.json',
];
for (const filename of required) {
  test(`rejects missing ${filename} before writing outputs`, (t) => {
    const { root, files } = fixture(t);
    fs.unlinkSync(files.get(filename));
    const result = run(root);
    assert.notEqual(result.status, 0);
    assert.ok(result.stderr.includes(filename), result.stderr);
    assertNoOutputs(root);
  });

  test(`rejects empty ${filename} before writing outputs`, (t) => {
    const { root, files } = fixture(t);
    fs.writeFileSync(files.get(filename), '');
    const result = run(root);
    assert.notEqual(result.status, 0);
    assert.ok(result.stderr.includes(filename), result.stderr);
    assertNoOutputs(root);
  });
}

for (const filename of required.slice(0, 2)) {
  test(`rejects duplicate ${filename} before writing outputs`, (t) => {
    const { root, files, write } = fixture(t);
    write(
      `binaries/duplicate/${filename}`,
      fs.readFileSync(files.get(filename)),
    );
    const result = run(root);
    assert.notEqual(result.status, 0);
    assert.ok(result.stderr.includes(filename), result.stderr);
    assertNoOutputs(root);
  });
}

for (const [filename, target] of [
  ['linux-arm64-rslint', 'linux-arm64-gnu/rslint'],
  ['rslint.linux-x64-musl.node', 'linux-x64-musl/rslint.linux-x64-musl.node'],
]) {
  test(`stale package output cannot conceal missing ${filename}`, (t) => {
    const { root, files, write } = fixture(t);
    fs.unlinkSync(files.get(filename));
    const stale = write(`npm/rslint/${target}`, 'previous build');
    const result = run(root);
    assert.notEqual(result.status, 0);
    assert.ok(result.stderr.includes(filename), result.stderr);
    assert.equal(fs.readFileSync(stale, 'utf8'), 'previous build');
  });
}

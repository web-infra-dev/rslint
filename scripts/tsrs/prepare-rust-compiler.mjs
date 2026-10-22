#!/usr/bin/env node
// Prepare a reproducible Rust-only compiler graph with the frozen baseline libs.
// No Go SDK, Go source generation, or mutable external checkout is used.
import assert from 'node:assert/strict';
import crypto from 'node:crypto';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { execFileSync } from 'node:child_process';
import { fileURLToPath } from 'node:url';

const repository = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  '../..',
);
const origin = 'https://github.com/oxc-project/tsrs.git';
const revision = '280b426b89cd4ecd6153cf48228826bd60a99c61';
const libraryRevision = '1f70213d4922b434345f639b441681e470c7cfc1';
const vendor = path.join(repository, 'vendor/typescript-libs');
const patchDirectory = path.join(
  repository,
  'scripts/tsrs/rust-compiler-patches',
);
const options = {};
const flags = new Set(['check', 'validate-inputs', 'verify-upstream']);
const values = new Set([
  'mirror',
  'typescript-mirror',
  'source-cache',
  'config-out',
]);
for (let i = 2; i < process.argv.length; i++) {
  const argument = process.argv[i];
  if (argument === '--help') {
    process.stdout.write(
      `Usage: node scripts/tsrs/prepare-rust-compiler.mjs [options]\n\n--mirror DIR              Read committed tsrs objects from an optional local clone\n--typescript-mirror DIR   Read committed TypeScript objects for --verify-upstream\n--source-cache DIR        Isolated source cache, outside this repository\n--config-out FILE         Write this explicit Cargo --config file\n--validate-inputs         Verify vendored hashes, library references and patches only\n--verify-upstream         Also compare all vendor bytes and aliases to the public pin\n--check                   Verify an existing prepared graph; do not create or edit it\n`,
    );
    process.exit(0);
  }
  assert(argument.startsWith('--'), `unexpected argument: ${argument}`);
  const key = argument.slice(2);
  assert(!Object.hasOwn(options, key), `duplicate option: ${argument}`);
  if (flags.has(key)) options[key] = true;
  else {
    assert(values.has(key), `unknown option: ${argument}`);
    assert(
      process.argv[i + 1] && !process.argv[i + 1].startsWith('--'),
      `missing value: ${argument}`,
    );
    options[key] = path.resolve(process.argv[++i]);
  }
}
const hash = (value) => crypto.createHash('sha256').update(value).digest('hex');
const read = (name) => fs.readFileSync(name);
const json = (name) => JSON.parse(read(name));
function git(directory, ...args) {
  return execFileSync('git', ['-C', directory, ...args], {
    maxBuffer: 64 * 1024 * 1024,
    stdio: ['ignore', 'pipe', 'pipe'],
  });
}
function relativeFile(directory, name) {
  assert(
    typeof name === 'string' && name.length > 0 && !path.isAbsolute(name),
    `invalid relative file: ${name}`,
  );
  assert(!name.split(/[\\/]/).includes('..'), `file escapes source: ${name}`);
  return path.join(directory, name);
}
function exactFile(file, expected) {
  const stat = fs.lstatSync(file);
  assert(
    stat.isFile() && !stat.isSymbolicLink(),
    `expected regular file: ${file}`,
  );
  const bytes = read(file);
  assert.equal(bytes.length, expected.bytes, `byte count changed: ${file}`);
  assert.equal(hash(bytes), expected.sha256, `content changed: ${file}`);
  return bytes;
}
function atomicWrite(file, contents) {
  fs.mkdirSync(path.dirname(file), { recursive: true });
  const temporary = `${file}.tmp-${process.pid}-${crypto.randomBytes(8).toString('hex')}`;
  try {
    fs.writeFileSync(temporary, contents, { flag: 'wx' });
    fs.renameSync(temporary, file);
  } finally {
    fs.rmSync(temporary, { force: true });
  }
}
const manifestPath = path.join(vendor, 'manifest.json');
const manifest = json(manifestPath);
assert.equal(manifest.schemaVersion, 1);
assert.equal(
  manifest.repository,
  'https://github.com/microsoft/TypeScript.git',
);
assert.equal(manifest.revision, libraryRevision);
const inputs = {
  'vendor/typescript-libs/manifest.json': hash(read(manifestPath)),
};
const names = new Set();
for (const file of manifest.files) {
  assert(!names.has(file.path), `duplicate resource: ${file.path}`);
  names.add(file.path);
  const bytes = exactFile(relativeFile(vendor, file.path), file);
  inputs[`vendor/typescript-libs/${file.path}`] = hash(bytes);
}
const libraryFiles = manifest.files.filter((file) =>
  file.path.startsWith('lib/'),
);
assert.equal(libraryFiles.length, manifest.libraryCount);
assert.equal(
  libraryFiles.reduce((total, file) => total + file.bytes, 0),
  manifest.libraryBytes,
);
assert.deepEqual(
  fs.readdirSync(path.join(vendor, 'lib')).sort(),
  libraryFiles.map((file) => path.basename(file.path)).sort(),
);
const mapPath = relativeFile(vendor, manifest.libMap.path);
assert.equal(
  hash(read(mapPath)),
  manifest.libMap.sha256,
  'library map changed',
);
inputs[`vendor/typescript-libs/${manifest.libMap.path}`] =
  manifest.libMap.sha256;
const libraryMap = json(mapPath);
assert.equal(libraryMap.length, manifest.libMap.entries);
assert.equal(
  new Set(libraryMap.map((entry) => entry.name)).size,
  libraryMap.length,
);
const aliases = new Map(libraryMap.map((entry) => [entry.name, entry.file]));
for (const entry of libraryMap)
  assert(
    names.has(`lib/${entry.file}`),
    `alias target is absent: ${entry.name}`,
  );
for (const file of libraryFiles) {
  const text = read(path.join(vendor, file.path)).toString('utf8');
  assert(
    text.startsWith(
      '/*! *****************************************************************************\nCopyright (c) Microsoft Corporation.',
    ),
    `missing original copyright: ${file.path}`,
  );
  for (const match of text.matchAll(/\/\/\/\s*<reference\s+lib="([^"]+)"/g)) {
    assert(
      aliases.has(match[1]),
      `unknown library reference ${match[1]} in ${file.path}`,
    );
  }
}
const patchNames = fs
  .readdirSync(patchDirectory)
  .filter((name) => name.endsWith('.patch'))
  .sort();
assert(patchNames.length > 0, 'compiler patches are absent');
for (const name of patchNames)
  inputs[`scripts/tsrs/rust-compiler-patches/${name}`] = hash(
    read(path.join(patchDirectory, name)),
  );
inputs['scripts/tsrs/prepare-rust-compiler.mjs'] = hash(
  read(fileURLToPath(import.meta.url)),
);
const fingerprint = hash(JSON.stringify({ origin, revision, inputs }));
const cacheRoot =
  options['source-cache'] ??
  path.join(
    process.env.XDG_CACHE_HOME || path.join(os.homedir(), '.cache'),
    'rslint/tsrs-rust-compiler',
    hash(repository).slice(0, 16),
  );
const configPath =
  options['config-out'] ??
  path.join(
    repository,
    'node_modules/.cache/rslint-rust-compiler/cargo-config.toml',
  );
assert.notEqual(
  configPath,
  path.join(repository, '.cargo/config.toml'),
  'the existing workspace Cargo configuration must not be overwritten',
);
function assertOutsideRepository(directory) {
  const relative = path.relative(
    fs.realpathSync(repository),
    fs.realpathSync(directory),
  );
  assert(
    relative.startsWith(`..${path.sep}`) || path.isAbsolute(relative),
    'compiler source cache must be outside the consuming repository',
  );
}
function checkout(directory, url, pin, mirror) {
  fs.mkdirSync(directory);
  git(directory, 'init', '--quiet');
  git(directory, 'config', 'core.autocrlf', 'false');
  git(directory, 'remote', 'add', 'origin', url);
  git(directory, 'fetch', '--quiet', '--depth=1', mirror ?? 'origin', pin);
  git(directory, 'checkout', '--quiet', '--detach', pin);
  assert.equal(git(directory, 'rev-parse', 'HEAD').toString().trim(), pin);
  assert.equal(git(directory, 'status', '--porcelain').length, 0);
}
function verifyUpstream() {
  const temporary = fs.mkdtempSync(
    path.join(os.tmpdir(), 'rslint-library-origin-'),
  );
  try {
    const source =
      options['typescript-mirror'] ?? path.join(temporary, 'source');
    if (!options['typescript-mirror'])
      checkout(source, manifest.repository, manifest.revision);
    // A mirror contributes committed objects only. Dirty/untracked files and
    // its current branch cannot affect the resource comparison.
    assert.equal(
      git(source, 'rev-parse', `${manifest.revision}^{commit}`)
        .toString()
        .trim(),
      manifest.revision,
    );
    for (const file of manifest.files) {
      const contents = git(
        source,
        'show',
        `${manifest.revision}:${file.upstreamPath}`,
      );
      assert.equal(
        hash(contents),
        file.sha256,
        `upstream bytes differ: ${file.upstreamPath}`,
      );
    }
    const enumSource = git(
      source,
      'show',
      `${manifest.revision}:${manifest.libMap.upstreamPath}`,
    );
    assert.equal(hash(enumSource), manifest.libMap.upstreamSha256);
    const block = enumSource
      .toString()
      .split('var LibMap =')[1]
      ?.split('\n})')[0];
    assert(block, 'pinned ordered library map was not found');
    const upstreamMap = Array.from(
      block.matchAll(/\{Key: "([^"]+)", Value: "([^"]+)"\}/g),
      (match) => ({ name: match[1], file: match[2] }),
    );
    assert.deepEqual(
      libraryMap,
      upstreamMap,
      'alias values or load order differ from baseline',
    );
  } finally {
    fs.rmSync(temporary, { recursive: true, force: true });
  }
}
if (options['verify-upstream']) verifyUpstream();
if (options['validate-inputs']) {
  process.stdout.write(
    `${JSON.stringify({ status: 'inputs-verified', fingerprint, libraryRevision, libraryCount: manifest.libraryCount, libraryBytes: manifest.libraryBytes, upstreamVerified: Boolean(options['verify-upstream']) }, null, 2)}\n`,
  );
  process.exit(0);
}
if (options.check)
  assert(fs.existsSync(cacheRoot), 'prepared cache does not exist');
else fs.mkdirSync(cacheRoot, { recursive: true });
assertOutsideRepository(cacheRoot);
const prepared = path.join(
  cacheRoot,
  `${revision.slice(0, 12)}-${fingerprint.slice(0, 16)}`,
);
const source = path.join(prepared, 'source');
const receiptPath = path.join(prepared, 'prepared.json');
function verifyLibraryMap(source) {
  const text = read(
    path.join(source, 'crates/tsrs_tsoptions/src/enummaps.rs'),
  ).toString();
  const block = text.slice(
    text.indexOf('pub static LIB_MAP'),
    text.indexOf('pub static LIBS'),
  );
  const actual = Array.from(
    block.matchAll(
      /\(\s*"([^"\n]+)",\s*CompilerOptionsValue::String\("([^"\n]+)"\.to_string\(\)\),?\s*\)/g,
    ),
    (match) => ({ name: match[1], file: match[2] }),
  );
  assert.deepEqual(
    actual,
    libraryMap,
    'prepared alias table or its load order changed',
  );
  assert(
    !text.includes('ScriptTarget::ES2026'),
    'post-baseline target is still exposed',
  );
}
function generatedLibraryIndex() {
  const names = libraryFiles.map((file) => path.basename(file.path)).sort();
  const header =
    '// Generated by rslint scripts/tsrs/prepare-rust-compiler.mjs.\n// Microsoft TypeScript library revision: ' +
    libraryRevision +
    '\n\n';
  return {
    'libs_generated.rs': `${header}pub static LIB_NAMES: &[&str] = &[\n${names.map((name) => `    "${name}",`).join('\n')}\n];\n`,
    'embed_generated.rs': `${header}pub(crate) static EMBEDDED_CONTENTS: &[(&str, &str)] = &[\n${names.map((name) => `    ("libs/${name}", include_str!("../../libs/${name}")),`).join('\n')}\n];\n`,
  };
}
if (!fs.existsSync(prepared)) {
  assert(!options.check, 'prepared graph is missing; run preparation first');
  const temporary = fs.mkdtempSync(path.join(cacheRoot, '.preparing-'));
  try {
    const stageSource = path.join(temporary, 'source');
    checkout(stageSource, origin, revision, options.mirror);
    for (const name of patchNames) {
      const patch = path.join(patchDirectory, name);
      git(stageSource, 'apply', '--check', patch);
      git(stageSource, 'apply', patch);
    }
    const libs = path.join(stageSource, 'crates/tsrs_vfs/libs');
    // This is a fresh task-owned checkout. Replace the entire resource set so
    // post-baseline ES2026 files cannot remain reachable through a direct path.
    fs.rmSync(libs, { recursive: true });
    fs.mkdirSync(libs);
    for (const file of libraryFiles)
      fs.copyFileSync(
        path.join(vendor, file.path),
        path.join(libs, path.basename(file.path)),
      );
    for (const [name, contents] of Object.entries(generatedLibraryIndex()))
      fs.writeFileSync(
        path.join(stageSource, 'crates/tsrs_vfs/src/bundled', name),
        contents,
      );
    verifyLibraryMap(stageSource);
    git(stageSource, 'add', '--intent-to-add', 'crates/tsrs_vfs/libs');
    const diff = git(stageSource, 'diff', '--binary', 'HEAD');
    const metadata = {
      schemaVersion: 1,
      origin,
      revision,
      libraryRevision,
      fingerprint,
      inputs,
      diffSha256: hash(diff),
    };
    fs.writeFileSync(
      path.join(temporary, 'prepared.json'),
      `${JSON.stringify(metadata, null, 2)}\n`,
    );
    try {
      fs.renameSync(temporary, prepared);
    } catch (error) {
      // Another preparation with these identical immutable inputs may win.
      // Never replace its checkout; the normal verification below checks it.
      if (!fs.existsSync(prepared)) throw error;
    }
  } finally {
    fs.rmSync(temporary, { recursive: true, force: true });
  }
}
const receipt = json(receiptPath);
assert.equal(
  receipt.fingerprint,
  fingerprint,
  'prepared receipt has different inputs',
);
assert.equal(receipt.revision, revision);
assert.equal(receipt.libraryRevision, libraryRevision);
assert.equal(
  git(source, 'rev-parse', 'HEAD').toString().trim(),
  revision,
  'prepared HEAD changed',
);
assert.equal(
  git(source, 'remote', 'get-url', 'origin').toString().trim(),
  origin,
  'prepared origin changed',
);
assert.equal(
  hash(git(source, 'diff', '--binary', 'HEAD')),
  receipt.diffSha256,
  'prepared compiler was modified; refusing to overwrite it',
);
assert.equal(
  git(source, 'ls-files', '--others', '--exclude-standard').length,
  0,
  'prepared compiler contains unknown untracked files',
);
assert.deepEqual(
  fs.readdirSync(path.join(source, 'crates/tsrs_vfs/libs')).sort(),
  libraryFiles.map((file) => path.basename(file.path)).sort(),
);
for (const file of libraryFiles)
  exactFile(
    path.join(source, 'crates/tsrs_vfs/libs', path.basename(file.path)),
    file,
  );
for (const [name, contents] of Object.entries(generatedLibraryIndex()))
  assert.equal(
    read(path.join(source, 'crates/tsrs_vfs/src/bundled', name)).toString(),
    contents,
  );
verifyLibraryMap(source);
const packageEntries = [];
for (const entry of fs.readdirSync(path.join(source, 'crates'), {
  withFileTypes: true,
})) {
  if (!entry.isDirectory()) continue;
  const directory = path.join(source, 'crates', entry.name);
  const cargo = path.join(directory, 'Cargo.toml');
  if (!fs.existsSync(cargo)) continue;
  const manifest = read(cargo).toString();
  const section = manifest.match(
    /^\[package\]\s*\n([\s\S]*?)(?=^\[|$(?![\s\S]))/m,
  )?.[1];
  const name = section?.match(/^name\s*=\s*"([^"]+)"/m)?.[1];
  assert(name, `package name is absent: ${cargo}`);
  packageEntries.push([name, directory]);
}
packageEntries.sort(([left], [right]) => left.localeCompare(right));
assert(packageEntries.length >= 30, 'incomplete compiler workspace');
assert.equal(
  new Set(packageEntries.map(([name]) => name)).size,
  packageEntries.length,
);
const config = `# Generated by scripts/tsrs/prepare-rust-compiler.mjs; use cargo --config explicitly.\n# tsrs ${revision}; TypeScript libraries ${libraryRevision}\n# All workspace packages share this source to avoid mixing incompatible arena types.\n[patch.${JSON.stringify(origin)}]\n${packageEntries.map(([name, directory]) => `${JSON.stringify(name)} = { path = ${JSON.stringify(directory)} }`).join('\n')}\n`;
if (options.check)
  assert.equal(
    read(configPath).toString(),
    config,
    'Cargo override is absent or stale',
  );
else atomicWrite(configPath, config);
process.stdout.write(
  `${JSON.stringify({ status: 'prepared', origin, revision, libraryRevision, fingerprint, source, cargoConfig: configPath, receipt: receiptPath, libraryCount: manifest.libraryCount, libraryBytes: manifest.libraryBytes, patchedPackages: packageEntries.length, upstreamVerified: Boolean(options['verify-upstream']) }, null, 2)}\n`,
);

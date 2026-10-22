#!/usr/bin/env node
import assert from 'node:assert/strict';
import crypto from 'node:crypto';
import fs from 'node:fs';
import path from 'node:path';
import { spawnSync } from 'node:child_process';
import { fileURLToPath, pathToFileURL } from 'node:url';

export const TARGETS = Object.freeze({
  'darwin-arm64': 'aarch64-apple-darwin',
  'darwin-x64': 'x86_64-apple-darwin',
  'linux-arm64-gnu': 'aarch64-unknown-linux-gnu',
  'linux-x64-gnu': 'x86_64-unknown-linux-gnu',
  'linux-arm64-musl': 'aarch64-unknown-linux-musl',
  'linux-x64-musl': 'x86_64-unknown-linux-musl',
  'win32-arm64-msvc': 'aarch64-pc-windows-msvc',
  'win32-x64-msvc': 'x86_64-pc-windows-msvc',
});
const REPOSITORY = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  '../..',
);
const COMPILER_REVISION = '280b426b89cd4ecd6153cf48228826bd60a99c61';
const LIBRARY_REVISION = '1f70213d4922b434345f639b441681e470c7cfc1';
const HELP = `Usage: node scripts/tsrs/build-rust.mjs [options]

Build the in-process Rust lint addon. No Go executable, compiler service or WASM
is built, copied or downloaded. The Cargo.lock must already match the prepared
compiler graph: --locked is always enabled and a mismatch fails the build.

  --test                   Run the Rust core and native crate tests instead of building
  --debug                  Use Cargo's development profile (default: release)
  --platform TUPLE         Select one native npm platform package
  --target TRIPLE          Select one of the supported Rust targets instead
  --jobs N                 Limit Cargo parallelism (default: CARGO_BUILD_JOBS or 2)
  --target-dir DIR         Cargo artifact directory (default: target/rust-js)
  --source-cache DIR       Isolated tsrs cache, outside the repository
  --mirror DIR             Read the pinned committed tsrs objects from this clone
  --plan                   Print commands and output paths without executing them
  --help                   Show this help

Cross compilation needs the selected Rust target and its linker on PATH; the
script never substitutes an artifact for a different target.
`;

export function parseArguments(args) {
  const options = {};
  const flags = new Set(['debug', 'plan', 'help', 'test']);
  const values = new Set([
    'platform',
    'target',
    'jobs',
    'target-dir',
    'source-cache',
    'mirror',
  ]);
  for (let index = 0; index < args.length; index++) {
    const match = /^--([^=]+)(?:=(.*))?$/.exec(args[index]);
    assert(match, `unexpected argument: ${args[index]}`);
    const [, key, inline] = match;
    assert(!Object.hasOwn(options, key), `duplicate option: --${key}`);
    if (flags.has(key)) {
      assert(inline === undefined, `--${key} does not accept a value`);
      options[key] = true;
    } else {
      assert(values.has(key), `unknown option: --${key}`);
      const value = inline ?? args[++index];
      assert(value && !value.startsWith('--'), `missing value: --${key}`);
      options[key] = value;
    }
  }
  assert(
    !(options.platform && options.target),
    'choose --platform or --target, not both',
  );
  if (options.jobs !== undefined)
    assert(
      /^[1-9]\d*$/.test(options.jobs),
      '--jobs must be a positive integer',
    );
  return options;
}

export function hostTuple({
  platform = process.platform,
  arch = process.arch,
  glibc = process.report?.getReport()?.header?.glibcVersionRuntime,
} = {}) {
  const tuple =
    platform === 'linux'
      ? `linux-${arch}-${glibc ? 'gnu' : 'musl'}`
      : platform === 'win32'
        ? `win32-${arch}-msvc`
        : `${platform}-${arch}`;
  assert(Object.hasOwn(TARGETS, tuple), `unsupported host platform: ${tuple}`);
  return tuple;
}

export function createBuildPlan(
  options,
  {
    repository = REPOSITORY,
    tuple = hostTuple(),
    node = process.execPath,
    environment = process.env,
    cwd = process.cwd(),
  } = {},
) {
  const selected =
    options.platform ??
    (options.target
      ? Object.keys(TARGETS).find((key) => TARGETS[key] === options.target)
      : tuple);
  assert(
    selected && Object.hasOwn(TARGETS, selected),
    `unsupported native target: ${options.platform ?? options.target ?? tuple}`,
  );
  const target = TARGETS[selected];
  const jobs = options.jobs ?? environment.CARGO_BUILD_JOBS ?? '2';
  assert(
    /^[1-9]\d*$/.test(String(jobs)),
    'CARGO_BUILD_JOBS must be a positive integer',
  );
  const profile = options.debug ? 'debug' : 'release';
  const targetDirectory = options['target-dir']
    ? path.resolve(cwd, options['target-dir'])
    : environment.CARGO_TARGET_DIR
      ? path.resolve(cwd, environment.CARGO_TARGET_DIR)
      : path.join(repository, 'target/rust-js');
  const cargoConfig = path.join(
    repository,
    'node_modules/.cache/rslint-rust-compiler/cargo-config.toml',
  );
  const prepareArgs = [
    path.join(repository, 'scripts/tsrs/prepare-rust-compiler.mjs'),
    '--config-out',
    cargoConfig,
  ];
  for (const key of ['source-cache', 'mirror'])
    if (options[key])
      prepareArgs.push(`--${key}`, path.resolve(cwd, options[key]));
  const cargoArgs = [
    '--config',
    cargoConfig,
    options.test ? 'test' : 'build',
    '--locked',
    ...(options.test
      ? [
          '--no-fail-fast',
          '--package',
          'rslint-core',
          '--package',
          'rslint-native',
        ]
      : ['--package', 'rslint-native', '--lib']),
    '--target',
    target,
    '--target-dir',
    targetDirectory,
    '--jobs',
    String(jobs),
  ];
  if (!options.debug) cargoArgs.push('--release');
  const library = selected.startsWith('win32-')
    ? 'rslint_native.dll'
    : selected.startsWith('darwin-')
      ? 'librslint_native.dylib'
      : 'librslint_native.so';
  return {
    schemaVersion: 1,
    mode: options.test ? 'test' : 'build',
    repository,
    tuple: selected,
    target,
    profile,
    jobs: Number(jobs),
    compilerRevision: COMPILER_REVISION,
    libraryRevision: LIBRARY_REVISION,
    cargoConfig,
    prepare: { command: node, args: prepareArgs },
    build: { command: 'cargo', args: cargoArgs },
    artifact: path.join(targetDirectory, target, profile, library),
    destination: path.join(
      repository,
      'npm/rslint',
      selected,
      `rslint.${selected}.node`,
    ),
    receipt: path.join(
      targetDirectory,
      `${options.test ? 'test' : 'build'}-${selected}-${profile}.json`,
    ),
  };
}

function command(command, args, { cwd, capture = false }) {
  const result = spawnSync(command, args, {
    cwd,
    stdio: ['ignore', capture ? 'pipe' : 'inherit', 'inherit'],
    encoding: 'utf8',
    maxBuffer: 8 * 1024 * 1024,
  });
  if (result.error) throw result.error;
  assert.equal(
    result.signal,
    null,
    `${command} was terminated by ${result.signal}`,
  );
  assert.equal(
    result.status,
    0,
    `${command} exited with status ${result.status}`,
  );
  return capture ? result.stdout : '';
}
const digest = (value) =>
  crypto.createHash('sha256').update(value).digest('hex');
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
function assertAddon(bytes, tuple) {
  assert(bytes.length >= 4, 'Cargo produced an empty native addon');
  const header = bytes.subarray(0, 4).toString('hex');
  if (tuple.startsWith('darwin-'))
    assert(
      ['cffaedfe', 'feedfacf'].includes(header),
      'expected a 64-bit Mach-O addon',
    );
  else if (tuple.startsWith('linux-'))
    assert.equal(header, '7f454c46', 'expected an ELF addon');
  else
    assert.equal(bytes.subarray(0, 2).toString(), 'MZ', 'expected a PE addon');
}

/** Execute one prepared Cargo graph, and publish only its successful cdylib. */
export function executeBuild(plan, { run = command } = {}) {
  const packagePath = path.join(path.dirname(plan.destination), 'package.json');
  const metadata = JSON.parse(fs.readFileSync(packagePath, 'utf8'));
  assert.equal(
    metadata.name,
    `@rslint/native-${plan.tuple}`,
    'wrong destination package',
  );
  assert.equal(
    metadata.exports['.'],
    `./${path.basename(plan.destination)}`,
    'native export does not match build target',
  );
  const lockPath = path.join(plan.repository, 'Cargo.lock');
  const lockHash = digest(fs.readFileSync(lockPath));
  const prepared = JSON.parse(
    run(plan.prepare.command, plan.prepare.args, {
      cwd: plan.repository,
      capture: true,
    }),
  );
  assert.equal(prepared.status, 'prepared');
  assert.equal(
    prepared.revision,
    plan.compilerRevision,
    'unexpected tsrs revision',
  );
  assert.equal(
    prepared.libraryRevision,
    plan.libraryRevision,
    'unexpected TypeScript library revision',
  );
  assert.equal(
    prepared.cargoConfig,
    plan.cargoConfig,
    'preparation returned a different Cargo graph',
  );
  assert(
    typeof prepared.fingerprint === 'string' &&
      /^[a-f\d]{64}$/.test(prepared.fingerprint),
    'preparation omitted its source fingerprint',
  );
  run(plan.build.command, plan.build.args, { cwd: plan.repository });
  assert.equal(
    digest(fs.readFileSync(lockPath)),
    lockHash,
    'Cargo.lock changed during a locked build',
  );
  let addon;
  if (plan.mode === 'build') {
    const file = fs.lstatSync(plan.artifact);
    assert(
      file.isFile() && !file.isSymbolicLink(),
      'native output must be a regular file',
    );
    addon = fs.readFileSync(plan.artifact);
    assertAddon(addon, plan.tuple);
  }
  const cargoVersion = run('cargo', ['--version'], {
    cwd: plan.repository,
    capture: true,
  }).trim();
  const rustcVersion = run('rustc', ['--version'], {
    cwd: plan.repository,
    capture: true,
  }).trim();
  const receipt = {
    ...plan,
    compilerFingerprint: prepared.fingerprint,
    cargoLockSha256: lockHash,
    cargoVersion,
    rustcVersion,
    nodeVersion: process.version,
    ...(addon
      ? { artifactSha256: digest(addon), artifactBytes: addon.length }
      : {}),
  };
  // All checks complete before replacing a previous addon; a failed build never
  // falls back to an older artifact and never publishes a partial output.
  if (addon) atomicWrite(plan.destination, addon);
  atomicWrite(plan.receipt, `${JSON.stringify(receipt, null, 2)}\n`);
  return receipt;
}

export function main(args = process.argv.slice(2)) {
  const options = parseArguments(args);
  if (options.help) {
    process.stdout.write(HELP);
    return;
  }
  const plan = createBuildPlan(options);
  const result = options.plan ? plan : executeBuild(plan);
  process.stdout.write(`${JSON.stringify(result, null, 2)}\n`);
}
if (
  process.argv[1] &&
  pathToFileURL(path.resolve(process.argv[1])).href === import.meta.url
) {
  try {
    main();
  } catch (error) {
    process.stderr.write(
      `Rust addon build failed: ${error instanceof Error ? error.message : String(error)}\n`,
    );
    process.exitCode = 1;
  }
}

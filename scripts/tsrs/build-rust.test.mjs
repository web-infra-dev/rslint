import assert from 'node:assert/strict';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { test } from 'node:test';
import {
  TARGETS,
  createBuildPlan,
  executeBuild,
  hostTuple,
  parseArguments,
} from './build-rust.mjs';

function fixture(t, options = {}) {
  const repository = fs.mkdtempSync(
    path.join(os.tmpdir(), 'rslint-rust-build-'),
  );
  t.after(() => fs.rmSync(repository, { recursive: true, force: true }));
  const plan = createBuildPlan(options, {
    repository,
    tuple: 'darwin-arm64',
    environment: {},
    cwd: repository,
  });
  fs.mkdirSync(path.dirname(plan.destination), { recursive: true });
  fs.writeFileSync(
    path.join(path.dirname(plan.destination), 'package.json'),
    JSON.stringify({
      name: '@rslint/native-darwin-arm64',
      exports: { '.': './rslint.darwin-arm64.node' },
    }),
  );
  fs.writeFileSync(path.join(repository, 'Cargo.lock'), '# fixed lock\n');
  fs.writeFileSync(plan.destination, 'previous artifact');
  const calls = [];
  const run = (command, args) => {
    calls.push({ command, args });
    if (command === plan.prepare.command)
      return JSON.stringify({
        status: 'prepared',
        revision: plan.compilerRevision,
        libraryRevision: plan.libraryRevision,
        cargoConfig: plan.cargoConfig,
        fingerprint: 'a'.repeat(64),
      });
    if (args[0] === '--version') return `${command} test-version\n`;
    assert.equal(command, 'cargo');
    fs.mkdirSync(path.dirname(plan.artifact), { recursive: true });
    fs.writeFileSync(plan.artifact, Buffer.from('cffaedfe00000000', 'hex'));
    return '';
  };
  return { repository, plan, calls, run };
}

test('all native targets use a locked Rust graph and a direct addon', () => {
  for (const [tuple, target] of Object.entries(TARGETS)) {
    const plan = createBuildPlan(
      { platform: tuple },
      { repository: '/fixture', environment: {}, tuple: 'darwin-arm64' },
    );
    assert.equal(plan.target, target);
    assert(plan.build.args.includes('--locked'));
    assert(plan.build.args.includes('--release'));
    assert(plan.build.args.includes('rslint-native'));
    assert(plan.destination.endsWith(`rslint.${tuple}.node`));
    assert.equal(plan.build.command, 'cargo');
    assert(!JSON.stringify(plan).includes('rslint-tsrs-service'));
  }
  assert.equal(
    hostTuple({ platform: 'linux', arch: 'x64', glibc: '2.31' }),
    'linux-x64-gnu',
  );
  assert.equal(
    hostTuple({ platform: 'linux', arch: 'x64', glibc: undefined }),
    'linux-x64-musl',
  );
  assert.throws(() => parseArguments(['--platform']), /missing value/);
  assert.throws(() => parseArguments(['--debug=true']), /does not accept/);
  assert.throws(() => createBuildPlan({ jobs: '0' }), /positive integer/);
});

test('a successful build atomically copies only the matching Rust cdylib', (t) => {
  const { plan, calls, run } = fixture(t);
  const receipt = executeBuild(plan, { run });
  assert.deepEqual(
    fs.readFileSync(plan.destination),
    fs.readFileSync(plan.artifact),
  );
  assert.equal(receipt.artifactBytes, 8);
  assert.equal(
    JSON.parse(fs.readFileSync(plan.receipt)).compilerFingerprint,
    'a'.repeat(64),
  );
  assert.deepEqual(
    calls.map(({ command }) => command),
    [process.execPath, 'cargo', 'cargo', 'rustc'],
  );
});

test('preparation pin mismatch and failed compilation leave the previous artifact intact', (t) => {
  const { plan, run } = fixture(t);
  assert.throws(
    () =>
      executeBuild(plan, {
        run: (command, args) =>
          command === plan.prepare.command
            ? JSON.stringify({ status: 'prepared', revision: 'wrong' })
            : run(command, args),
      }),
    /unexpected tsrs revision/,
  );
  assert.throws(
    () =>
      executeBuild(plan, {
        run: (command, args) =>
          command === 'cargo'
            ? (() => {
                throw new Error('compiler failed');
              })()
            : run(command, args),
      }),
    /compiler failed/,
  );
  assert.equal(fs.readFileSync(plan.destination, 'utf8'), 'previous artifact');
  assert(!fs.existsSync(plan.receipt));
});

test('wrong artifact format or a modified lock cannot publish an output', (t) => {
  const { repository, plan, run } = fixture(t);
  assert.throws(
    () =>
      executeBuild(plan, {
        run: (command, args) => {
          const result = run(command, args);
          if (command === 'cargo' && args[0] !== '--version')
            fs.writeFileSync(plan.artifact, 'not a native addon');
          return result;
        },
      }),
    /Mach-O/,
  );
  assert.equal(fs.readFileSync(plan.destination, 'utf8'), 'previous artifact');
  assert.throws(
    () =>
      executeBuild(plan, {
        run: (command, args) => {
          const result = run(command, args);
          if (command === 'cargo' && args[0] !== '--version')
            fs.appendFileSync(
              path.join(repository, 'Cargo.lock'),
              '# changed\n',
            );
          return result;
        },
      }),
    /Cargo.lock changed/,
  );
  assert.equal(fs.readFileSync(plan.destination, 'utf8'), 'previous artifact');
});

test('test mode cannot replace the native artifact', (t) => {
  const { plan, run } = fixture(t, { test: true });
  assert(plan.build.args.includes('--no-fail-fast'));
  const receipt = executeBuild(plan, { run });
  assert.equal(receipt.artifactSha256, undefined);
  assert.equal(fs.readFileSync(plan.destination, 'utf8'), 'previous artifact');
});

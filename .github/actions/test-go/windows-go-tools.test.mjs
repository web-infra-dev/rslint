// cspell:ignore importcfg packagefile toolexec
import assert from 'node:assert/strict';
import { EventEmitter, once } from 'node:events';
import { mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import path from 'node:path';
import { spawn, spawnSync } from 'node:child_process';
import { fileURLToPath } from 'node:url';
import { test } from 'node:test';
import {
  acquire,
  createScheduler,
  estimateLinkMemoryMiB,
  linkerInputs,
} from './windows-go-tools.mjs';

const MiB = 1024 * 1024;
const request = (kind, inputMiB, pid) => ({
  kind,
  input_bytes: inputMiB * MiB,
  pid,
});
async function release(socket) {
  const closed = once(socket, 'close');
  socket.end();
  await closed;
}

test('large links wait for memory; small tools retain concurrency; queued work stays fair', async () => {
  const events = [];
  const changes = new EventEmitter();
  const scheduler = await createScheduler(40 * 1024, 5 * 1024, (event) => {
    events.push(event);
    changes.emit(`${event.event}-${event.pid}`);
  });
  try {
    const first = await acquire(
      scheduler.endpoint,
      request('link', 24 * 1024, 1),
    );
    const second = await acquire(scheduler.endpoint, request('test', 0, 2));
    const third = await acquire(scheduler.endpoint, request('tool', 0, 3));
    assert.equal(scheduler.stats.max_reserved_mib, 40 * 1024);
    assert.equal(scheduler.stats.max_concurrent_processes, 3);
    assert.equal(scheduler.stats.max_concurrent_links, 1);
    assert.equal(scheduler.stats.max_link_estimate_mib, 30 * 1024);
    const queuedFourth = once(changes, 'queue-4');
    const waiting = acquire(scheduler.endpoint, request('link', 24 * 1024, 4));
    await queuedFourth;
    const queuedFifth = once(changes, 'queue-5');
    const small = acquire(scheduler.endpoint, request('tool', 0, 5));
    await queuedFifth;
    await release(second);
    await release(third);
    assert.equal(scheduler.stats.grants, 3);
    await release(first);
    const [fourth, fifth] = await Promise.all([waiting, small]);
    assert.deepEqual(
      events.filter((e) => e.event === 'grant').map((e) => e.pid),
      [1, 2, 3, 4, 5],
    );
    assert.ok(events.every((e) => e.used_mib >= 0 && e.used_mib <= 40 * 1024));
    await release(fourth);
    await release(fifth);
  } finally {
    await scheduler.close();
  }
});

test('an oversized linker runs alone, is recorded, and disconnect releases its reservation', async () => {
  const scheduler = await createScheduler(40 * 1024, 4 * 1024);
  try {
    const big = await acquire(
      scheduler.endpoint,
      request('link', 50 * 1024, 1),
    );
    const next = acquire(scheduler.endpoint, request('test', 0, 2));
    assert.equal(scheduler.stats.oversized_links, 1);
    big.destroy();
    const small = await next;
    assert.equal(scheduler.stats.grants, 2);
    await release(small);
  } finally {
    await scheduler.close();
  }
});

test('measured link sizes retain small-task concurrency and include large-task overhead', () => {
  const slot = 3686;
  const common = estimateLinkMemoryMiB(1897 * MiB, slot);
  assert.equal(common, slot);
  assert.equal(Math.floor(36864 / common), 10);
  assert.equal(estimateLinkMemoryMiB(30432757580, slot), 36279);
  assert.ok(estimateLinkMemoryMiB(30110032156, slot) > 33961.93);
  // A small per-process slot must not erase the fixed linker overhead.
  assert.ok(estimateLinkMemoryMiB(Math.ceil(2094.46 * MiB), 512) > 3085.75);
  for (const cores of [10, 20, 40]) {
    const budget = Math.floor(cores * 4 * 1024 * 0.9);
    assert.ok(Math.floor(budget / common) >= cores);
  }
});

test('link inputs include cached main archives, paths with spaces, and unique dependency files', () => {
  const directory = mkdtempSync(path.join(tmpdir(), 'go link input '));
  try {
    const dependency = path.join(directory, 'dependency.a');
    const main = path.join(directory, 'cached-main-d');
    const config = path.join(directory, 'importcfg.link');
    writeFileSync(dependency, Buffer.alloc(300));
    writeFileSync(main, Buffer.alloc(40));
    writeFileSync(
      config,
      `packagefile example/a=${dependency}\r\npackagefile example/b=${dependency}\r\n`,
    );
    assert.equal(
      linkerInputs(['-o', 'test.exe', '-importcfg', config, main]).input_bytes,
      340,
    );
    assert.equal(linkerInputs([`-importcfg=${config}`, main]).input_count, 2);
    assert.throws(() => linkerInputs(['-importcfg', 'missing', main]));
  } finally {
    rmSync(directory, { recursive: true });
  }
});

test('version probes preserve stdout, arguments, and exit status without needing the scheduler', () => {
  const script = fileURLToPath(
    new URL('./windows-go-tools.mjs', import.meta.url),
  );
  const args = [
    '-e',
    'console.log(JSON.stringify(process.argv.slice(1))); process.exitCode = 7',
    '--',
    '-V=full',
    'a b',
    '"quote"',
  ];
  const result = spawnSync(
    process.execPath,
    [script, 'tool', process.execPath, ...args],
    { encoding: 'utf8' },
  );
  assert.equal(result.status, 7, result.stderr);
  assert.deepEqual(JSON.parse(result.stdout), ['-V=full', 'a b', '"quote"']);
});

test('test wrapper forwards arguments and failure status and releases its reservation', async () => {
  const records = [];
  const scheduler = await createScheduler(40, 4, (record) =>
    records.push(record),
  );
  try {
    const script = fileURLToPath(
      new URL('./windows-go-tools.mjs', import.meta.url),
    );
    const child = spawn(
      process.execPath,
      [
        script,
        'exec',
        process.execPath,
        '-e',
        'console.log(JSON.stringify(process.argv.slice(1)));process.exitCode=3',
        '--',
        'a b',
        '中',
      ],
      {
        env: {
          ...process.env,
          RSLINT_GO_SCHEDULER: JSON.stringify(scheduler.endpoint),
        },
        stdio: ['ignore', 'pipe', 'pipe'],
      },
    );
    let output = '';
    let error = '';
    child.stdout.on('data', (chunk) => {
      output += chunk;
    });
    child.stderr.on('data', (chunk) => {
      error += chunk;
    });
    const [code] = await once(child, 'close');
    assert.equal(code, 3, error);
    const result = JSON.parse(output);
    assert.deepEqual(result, ['a b', '中']);
    assert.deepEqual(
      records.map((r) => r.event),
      ['queue', 'grant', 'release'],
    );
    assert.equal(records.at(-1).used_mib, 0);
  } finally {
    await scheduler.close();
  }
});

test(
  'losing the coordinator stops a running tool and fails the wrapper',
  { timeout: 15000 },
  async () => {
    const scheduler = await createScheduler(40, 4);
    const script = fileURLToPath(
      new URL('./windows-go-tools.mjs', import.meta.url),
    );
    const child = spawn(
      process.execPath,
      [
        script,
        'exec',
        process.execPath,
        '-e',
        'console.log("started"); setInterval(() => {}, 1000)',
      ],
      {
        env: {
          ...process.env,
          RSLINT_GO_SCHEDULER: JSON.stringify(scheduler.endpoint),
        },
        stdio: ['ignore', 'pipe', 'pipe'],
      },
    );
    const completed = once(child, 'close');
    try {
      await once(child.stdout, 'data');
      await scheduler.close();
      const [code] = await completed;
      assert.notEqual(code, 0);
      assert.equal(scheduler.stats.grants, 1);
    } finally {
      await scheduler.close();
      child.kill();
    }
  },
);

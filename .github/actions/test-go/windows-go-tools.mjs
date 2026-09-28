// Share an estimated memory budget across Windows Go tools over a named pipe.
// cspell:ignore toolexec importcfg packagefile DWARF gcflags
import { spawn } from 'node:child_process';
import { randomBytes } from 'node:crypto';
import { once } from 'node:events';
import { readFileSync, statSync } from 'node:fs';
import net from 'node:net';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const script = fileURLToPath(import.meta.url);
const MiB = 1024 * 1024;

function estimateLinkMemoryMiB(inputBytes, slotMiB) {
  const inputMiB = Math.ceil(inputBytes / MiB);
  // Observed linker overhead reached 991 MiB for small inputs and 18.3% for
  // large inputs. Leave room above both, without adding the full process slot
  // on top of every archive. These reservations are estimates, not limits.
  const overheadMiB = Math.max(1536, Math.ceil(inputMiB / 4));
  return Math.max(slotMiB, inputMiB + overheadMiB);
}

function linkerInputBytes(args) {
  const index = args.indexOf('-importcfg');
  const config =
    index >= 0
      ? args[index + 1]
      : args.find((arg) => arg.startsWith('-importcfg='))?.slice(11);
  if (!config) throw new Error('Linker invocation has no import configuration');
  const files = new Set();
  for (const line of readFileSync(config, 'utf8').split(/\r?\n/)) {
    const match = /^packagefile ([^=]+)=(.+)$/.exec(line);
    if (match) files.add(match[2]);
  }
  // cmd/go places the main archive last. Cached archives have no .a suffix.
  const main = args.at(-1);
  if (!main || main.startsWith('-'))
    throw new Error('Linker invocation has no main archive');
  files.add(main);
  return [...files].reduce((sum, file) => sum + statSync(file).size, 0);
}

async function createScheduler(budgetMiB, slotMiB) {
  if (
    !Number.isSafeInteger(budgetMiB) ||
    budgetMiB < 1 ||
    !Number.isSafeInteger(slotMiB) ||
    slotMiB < 1
  ) {
    throw new Error('Expected positive integer memory budgets');
  }
  const token = randomBytes(24).toString('hex');
  const pipe = String.raw`\\.\pipe\rslint-go-${randomBytes(16).toString('hex')}`;
  const sockets = new Set();
  const queue = [];
  let used = 0;
  let active = 0;
  let activeLinks = 0;
  const stats = {
    grants: 0,
    links: 0,
    oversized_links: 0,
    max_reserved_mib: 0,
    max_concurrent_processes: 0,
    max_concurrent_links: 0,
    max_link_estimate_mib: 0,
    wait_ms: 0,
  };
  function drain() {
    while (queue.length) {
      const entry = queue[0];
      // An input larger than the entire estimate budget must run alone.
      // This is an explicit exception, not a claim of a hard memory bound.
      if (used + entry.reservedMiB > budgetMiB) return;
      queue.shift();
      entry.active = true;
      used += entry.reservedMiB;
      active++;
      activeLinks += Number(entry.kind === 'link');
      const waitMs = Date.now() - entry.queuedAt;
      stats.grants++;
      stats.links += Number(entry.kind === 'link');
      stats.oversized_links += Number(
        entry.kind === 'link' && entry.estimateMiB > budgetMiB,
      );
      stats.max_reserved_mib = Math.max(stats.max_reserved_mib, used);
      stats.max_concurrent_processes = Math.max(
        stats.max_concurrent_processes,
        active,
      );
      stats.max_concurrent_links = Math.max(
        stats.max_concurrent_links,
        activeLinks,
      );
      if (entry.kind === 'link') {
        stats.max_link_estimate_mib = Math.max(
          stats.max_link_estimate_mib,
          entry.estimateMiB,
        );
      }
      stats.wait_ms += waitMs;
      entry.socket.write('ready\n');
    }
  }
  const server = net.createServer((socket) => {
    sockets.add(socket);
    let entry;
    let buffer = '';
    socket.setEncoding('utf8');
    socket.on('error', () => socket.destroy());
    socket.on('data', (chunk) => {
      buffer += chunk;
      // Each connection requests one lease and holds it until the tool exits.
      if (entry || buffer.length > 16384) return socket.destroy();
      if (!buffer.endsWith('\n')) return;
      try {
        const request = JSON.parse(buffer);
        if (
          request.token !== token ||
          !['tool', 'link', 'test'].includes(request.kind) ||
          !Number.isSafeInteger(request.inputBytes) ||
          request.inputBytes < 0
        ) {
          throw new Error('Invalid scheduler request');
        }
        const estimateMiB =
          request.kind === 'link'
            ? estimateLinkMemoryMiB(request.inputBytes, slotMiB)
            : slotMiB;
        entry = {
          socket,
          kind: request.kind,
          estimateMiB,
          reservedMiB: Math.min(budgetMiB, estimateMiB),
          queuedAt: Date.now(),
          active: false,
        };
        queue.push(entry);
        drain();
      } catch {
        socket.destroy();
      }
    });
    socket.on('close', () => {
      sockets.delete(socket);
      if (entry?.active) {
        used -= entry.reservedMiB;
        active--;
        activeLinks -= Number(entry.kind === 'link');
      } else if (entry) {
        const index = queue.indexOf(entry);
        if (index >= 0) queue.splice(index, 1);
      }
      drain();
    });
  });
  // Windows owns the pipe lifetime; no TCP port or socket file to clean up.
  server.listen(pipe);
  await once(server, 'listening');
  return {
    endpoint: { pipe, token },
    stats,
    async close() {
      // The go command has finished or been interrupted; no new work may start.
      queue.length = 0;
      for (const socket of sockets) socket.destroy();
      await new Promise((resolve) => server.close(resolve));
    },
  };
}

async function acquire(endpoint, kind, inputBytes) {
  const socket = net.createConnection(endpoint.pipe);
  socket.setEncoding('utf8');
  await once(socket, 'connect');
  const ready = new Promise((resolve, reject) => {
    let response = '';
    socket.on('error', reject);
    socket.once('close', () =>
      reject(
        new Error('Go memory scheduler disconnected before granting memory'),
      ),
    );
    socket.on('data', (chunk) => {
      response += chunk;
      if (response === 'ready\n') resolve();
    });
  });
  socket.write(
    JSON.stringify({ kind, inputBytes, token: endpoint.token }) + '\n',
  );
  await ready;
  return socket;
}

async function execute(command, args, env = process.env, lease) {
  const child = spawn(command, args, { env, stdio: 'inherit' });
  const kill = () => child.kill();
  // A lost coordinator must not leave an unaccounted tool running.
  lease?.once('close', kill);
  process.once('SIGINT', kill);
  process.once('SIGTERM', kill);
  try {
    const [code] = await once(child, 'exit');
    return code ?? 1;
  } finally {
    process.removeListener('SIGINT', kill);
    process.removeListener('SIGTERM', kill);
    lease?.removeListener('close', kill);
    lease?.end();
  }
}

async function main([mode, ...args]) {
  if (mode === 'tool' || mode === 'exec') {
    const [tool, ...toolArgs] = args;
    // Tool version queries participate in Go's build cache identity. Forward
    // their output unchanged and never turn a version probe into a link job.
    if (mode === 'tool' && toolArgs.includes('-V=full'))
      return execute(tool, toolArgs);
    const kind =
      mode === 'exec'
        ? 'test'
        : /^link(?:\.exe)?$/i.test(path.basename(tool))
          ? 'link'
          : 'tool';
    const lease = await acquire(
      JSON.parse(process.env.RSLINT_GO_SCHEDULER),
      kind,
      kind === 'link' ? linkerInputBytes(toolArgs) : 0,
    );
    return execute(tool, toolArgs, process.env, lease);
  }
  if (mode !== 'run') throw new Error('Expected run, tool, or exec');
  const [budgetText, cpuText, command, ...goArgs] = args;
  if (!['build', 'test'].includes(command))
    throw new Error('Only Go build and test are supported');
  const budget = Number(budgetText);
  const cpu = Number(cpuText);
  if (!Number.isSafeInteger(cpu) || cpu < 1)
    throw new Error('Invalid CPU parallelism');
  // Missing memory quotas preserve the ordinary go command and its flags.
  if (budget === 0) return execute('go', [command, ...goArgs]);
  // Go's quoted.Split strips quotes without interpreting backslash escapes.
  const wrapper = [process.execPath, script]
    .map((value) => {
      if (value.includes('"'))
        throw new Error('Unsupported quote in tool wrapper path');
      return `"${value}"`;
    })
    .join(' ');
  const scheduler = await createScheduler(budget, Math.floor(budget / cpu));
  try {
    const flags = [`-toolexec=${wrapper} tool`];
    if (command === 'test') flags.push(`-exec=${wrapper} exec`);
    const code = await execute('go', [command, ...flags, ...goArgs], {
      ...process.env,
      RSLINT_GO_SCHEDULER: JSON.stringify(scheduler.endpoint),
    });
    console.log(
      'Windows Go scheduler: ' +
        JSON.stringify({
          phase: command,
          transport: 'named-pipe',
          budget_mib: budget,
          slot_mib: Math.floor(budget / cpu),
          ...scheduler.stats,
          exit_code: code,
        }),
    );
    return code;
  } finally {
    await scheduler.close();
  }
}

main(process.argv.slice(2))
  .then((code) => {
    process.exitCode = code;
  })
  .catch((error) => {
    console.error(error);
    process.exitCode = 1;
  });

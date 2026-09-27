// Schedule Go processes by estimated memory without reducing CPU parallelism.
// cspell:ignore toolexec importcfg packagefile TOOLEXEC IMPORTPATH DWARF gcflags
import { spawn } from 'node:child_process';
import { randomBytes } from 'node:crypto';
import { once } from 'node:events';
import { appendFileSync, readFileSync, statSync } from 'node:fs';
import net from 'node:net';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const script = fileURLToPath(import.meta.url);
const MiB = 1024 * 1024;

export function estimateLinkMemoryMiB(inputBytes, slotMiB) {
  const inputMiB = Math.ceil(inputBytes / MiB);
  // Observed linker overhead reached 991 MiB for small inputs and 18.3% for
  // large inputs. Leave room above both, without adding the full process slot
  // on top of every archive. These reservations are estimates, not limits.
  const overheadMiB = Math.max(1536, Math.ceil(inputMiB / 4));
  return Math.max(slotMiB, inputMiB + overheadMiB);
}

export function linkerInputs(args) {
  const index = args.indexOf('-importcfg');
  const config =
    index >= 0
      ? args[index + 1]
      : args.find((arg) => arg.startsWith('-importcfg='))?.slice(11);
  if (!config) throw new Error('Linker invocation has no import configuration');
  const files = new Map();
  for (const line of readFileSync(config, 'utf8').split(/\r?\n/)) {
    const match = /^packagefile ([^=]+)=(.+)$/.exec(line);
    if (match) files.set(match[2], match[1]);
  }
  // cmd/go places the main archive last. Cached archives have no .a suffix.
  const main = args.at(-1);
  if (!main || main.startsWith('-'))
    throw new Error('Linker invocation has no main archive');
  files.set(main, '(main)');
  const inputs = [...files].map(([file, name]) => ({
    name,
    bytes: statSync(file).size,
  }));
  return {
    input_bytes: inputs.reduce((sum, input) => sum + input.bytes, 0),
    input_count: inputs.length,
    largest_inputs: inputs.sort((a, b) => b.bytes - a.bytes).slice(0, 8),
  };
}

export async function createScheduler(budgetMiB, slotMiB, record = () => {}) {
  if (
    !Number.isSafeInteger(budgetMiB) ||
    budgetMiB < 1 ||
    !Number.isSafeInteger(slotMiB) ||
    slotMiB < 1
  ) {
    throw new Error('Expected positive integer memory budgets');
  }
  const token = randomBytes(24).toString('hex');
  const sockets = new Set();
  const queue = [];
  let used = 0;
  const stats = {
    grants: 0,
    links: 0,
    oversized_links: 0,
    max_reserved_mib: 0,
    wait_ms: 0,
  };
  function drain() {
    while (queue.length) {
      const entry = queue[0];
      // An input larger than the entire estimate budget must run alone.
      // This is an explicit exception, not a claim of a hard memory bound.
      if (used + entry.reserved_mib > budgetMiB) return;
      queue.shift();
      entry.active = true;
      used += entry.reserved_mib;
      const waitMs = Date.now() - entry.queuedAt;
      stats.grants++;
      stats.links += Number(entry.kind === 'link');
      stats.oversized_links += Number(
        entry.kind === 'link' && entry.estimate_mib > budgetMiB,
      );
      stats.max_reserved_mib = Math.max(stats.max_reserved_mib, used);
      stats.wait_ms += waitMs;
      record({
        event: 'grant',
        ...entry.details,
        reserved_mib: entry.reserved_mib,
        wait_ms: waitMs,
        used_mib: used,
      });
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
      if (buffer.length > 16384) return socket.destroy();
      try {
        while (buffer.includes('\n')) {
          const end = buffer.indexOf('\n');
          const request = JSON.parse(buffer.slice(0, end));
          buffer = buffer.slice(end + 1);
          if (entry) {
            if (
              !entry.active ||
              entry.details.tool_pid ||
              !Number.isSafeInteger(request.tool_pid) ||
              request.tool_pid < 1
            ) {
              throw new Error('Invalid child process record');
            }
            entry.details.tool_pid = request.tool_pid;
            record({ event: 'start', ...entry.details, used_mib: used });
            continue;
          }
          if (
            request.token !== token ||
            !['tool', 'link', 'test'].includes(request.kind) ||
            !Number.isSafeInteger(request.input_bytes) ||
            request.input_bytes < 0
          ) {
            throw new Error('Invalid scheduler request');
          }
          const estimate =
            request.kind === 'link'
              ? estimateLinkMemoryMiB(request.input_bytes, slotMiB)
              : slotMiB;
          const { token: _, ...details } = request;
          entry = {
            socket,
            kind: request.kind,
            estimate_mib: estimate,
            reserved_mib: Math.min(budgetMiB, estimate),
            details: { ...details, estimate_mib: estimate },
            queuedAt: Date.now(),
            active: false,
          };
          queue.push(entry);
          record({ event: 'queue', ...entry.details, used_mib: used });
          drain();
        }
      } catch {
        socket.destroy();
      }
    });
    socket.on('close', () => {
      sockets.delete(socket);
      if (entry?.active) {
        used -= entry.reserved_mib;
        record({
          event: 'release',
          pid: entry.details.pid,
          tool_pid: entry.details.tool_pid,
          used_mib: used,
        });
      } else if (entry) {
        const index = queue.indexOf(entry);
        if (index >= 0) queue.splice(index, 1);
      }
      drain();
    });
  });
  server.listen(0, '127.0.0.1');
  await once(server, 'listening');
  return {
    endpoint: { port: server.address().port, token },
    stats,
    async close() {
      // The go command has finished or been interrupted; no new work may start.
      queue.length = 0;
      for (const socket of sockets) socket.destroy();
      await new Promise((resolve) => server.close(resolve));
    },
  };
}

export async function acquire(endpoint, details) {
  const socket = net.createConnection({
    host: '127.0.0.1',
    port: endpoint.port,
  });
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
  socket.write(JSON.stringify({ ...details, token: endpoint.token }) + '\n');
  await ready;
  return socket;
}

export async function execute(command, args, env = process.env, lease) {
  const child = spawn(command, args, { env, stdio: 'inherit' });
  child.once('spawn', () =>
    lease?.write(JSON.stringify({ tool_pid: child.pid }) + '\n'),
  );
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
    const inputs =
      kind === 'link' ? linkerInputs(toolArgs) : { input_bytes: 0 };
    const lease = await acquire(JSON.parse(process.env.RSLINT_GO_SCHEDULER), {
      kind,
      pid: process.pid,
      package: process.env.TOOLEXEC_IMPORTPATH ?? path.basename(tool),
      ...inputs,
    });
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
  const log = path.join(process.env.RUNNER_TEMP, 'windows-go-tools.jsonl');
  const scheduler = await createScheduler(
    budget,
    Math.floor(budget / cpu),
    (data) => {
      appendFileSync(
        log,
        JSON.stringify({
          utc: new Date().toISOString(),
          phase: command,
          ...data,
        }) + '\n',
      );
    },
  );
  // Go's quoted.Split strips quotes without interpreting backslash escapes.
  const wrapper = [process.execPath, script]
    .map((value) => {
      if (value.includes('"'))
        throw new Error('Unsupported quote in tool wrapper path');
      return `"${value}"`;
    })
    .join(' ');
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

if (process.argv[1] && path.resolve(process.argv[1]) === script) {
  main(process.argv.slice(2))
    .then((code) => {
      process.exitCode = code;
    })
    .catch((error) => {
      console.error(error);
      process.exitCode = 1;
    });
}

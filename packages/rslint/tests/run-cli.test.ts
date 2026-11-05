import { spawn, spawnSync } from 'node:child_process';
import { once } from 'node:events';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { gunzipSync } from 'node:zlib';
import { describe, expect, test } from 'rstack/test';
import { runCLI } from '../src/index.js';

describe('runCLI', () => {
  test('is exported from the package root', () => {
    expect(typeof runCLI).toBe('function');
  });

  test.each([
    ['SIGHUP', 129],
    ['SIGINT', 130],
    ['SIGTERM', 143],
  ] as const)(
    'lets the host handle %s and continue after the run',
    (signal, exitCode) => {
      const cwd = fs.mkdtempSync(path.join(os.tmpdir(), 'rslint-host-signal-'));
      try {
        fs.writeFileSync(path.join(cwd, 'input.ts'), 'debugger;\n');
        fs.writeFileSync(
          path.join(cwd, 'rslint.config.mjs'),
          `process.emit('${signal}'); export default [];`,
        );
        const entry = new URL('../dist/index.js', import.meta.url).href;
        const host = `
        import { runCLI } from ${JSON.stringify(entry)};
        let interrupts = 0;
        process.on('${signal}', () => interrupts++);
        await runCLI({ argv: ['node', 'rslint', '--format=jsonline', 'input.ts'] });
        console.log(JSON.stringify([interrupts, process.exitCode]));
        process.exitCode = 0;
      `;
        const result = spawnSync(
          process.execPath,
          ['--input-type=module', '-e', host],
          {
            cwd,
            encoding: 'utf8',
            timeout: 30_000,
            killSignal: 'SIGKILL',
          },
        );
        expect(result.status).toBe(0);
        expect(JSON.parse(result.stdout)).toEqual([1, exitCode]);
        expect(result.stderr).not.toContain('error: ipc: channel closed');
      } finally {
        fs.rmSync(cwd, { recursive: true, force: true });
      }
    },
    40_000,
  );

  test('interrupts the CLI and releases its Go child while stdout stays paused', async () => {
    const cwd = fs.mkdtempSync(path.join(os.tmpdir(), 'rslint-interrupt-'));
    try {
      fs.writeFileSync(path.join(cwd, 'input.ts'), 'debugger;\n');
      // One row is larger than the OS pipe, so reading its first byte leaves
      // a real report write outstanding.
      fs.writeFileSync(
        path.join(cwd, 'rslint.config.mjs'),
        `export default [{ rules: { 'no-restricted-syntax': ['warn', {
          selector: 'DebuggerStatement', message: 'x'.repeat(4 * 1024 * 1024),
        }] } }];`,
      );
      const bin = fileURLToPath(new URL('../bin/rslint.js', import.meta.url));
      const child = spawn(
        process.execPath,
        [bin, '--format=jsonline', '--cpuprof=cpu.prof', 'input.ts'],
        {
          cwd,
          stdio: ['ignore', 'pipe', 'pipe'],
        },
      );
      const signal = AbortSignal.timeout(30_000);
      try {
        await once(child.stdout, 'readable', { signal });
        expect(child.stdout.read(1)?.toString()).toBe('{');
        child.stdout.on('readable', () => undefined); // Keep the pipe paused on host exit.
        // Go inherits stderr. Its EOF also proves that the child has released
        // its handles when only Node, not the foreground group, is signalled.
        const done = Promise.all([
          once(child, 'exit', { signal }),
          once(child.stderr, 'end', { signal }),
        ]);
        child.stderr.resume();
        child.kill('SIGINT');
        const [exit] = await done;
        expect(exit).toEqual([null, 'SIGINT']);
        // Windows child.kill() forcibly terminates Node and can kill its
        // entire job. On POSIX only Node gets SIGINT, so Go can flush on EOF.
        if (process.platform !== 'win32') {
          const profile = fs.readFileSync(path.join(cwd, 'cpu.prof'));
          expect(gunzipSync(profile).length).toBeGreaterThan(0);
        }
      } finally {
        child.stdout.destroy();
        if (child.exitCode === null && child.signalCode === null) {
          const stopped = once(child, 'exit');
          child.kill('SIGKILL');
          await stopped;
        }
      }
    } finally {
      fs.rmSync(cwd, { recursive: true, force: true });
    }
  }, 40_000);
});

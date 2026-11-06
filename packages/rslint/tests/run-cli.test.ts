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

  test('releases its Go child when JavaScript exits during config loading', () => {
    const cwd = fs.mkdtempSync(path.join(os.tmpdir(), 'rslint-exit-'));
    try {
      fs.writeFileSync(path.join(cwd, 'input.ts'), 'debugger;\n');
      fs.writeFileSync(
        path.join(cwd, 'rslint.config.mjs'),
        'process.exit(17); export default [];',
      );
      const bin = fileURLToPath(new URL('../bin/rslint.js', import.meta.url));
      // spawnSync also waits for inherited output pipes to close, so a Go
      // child left running after Node exits would time out here.
      const result = spawnSync(process.execPath, [bin, 'input.ts'], {
        cwd,
        encoding: 'utf8',
        timeout: 30_000,
        killSignal: 'SIGKILL',
      });
      expect(result.error).toBeUndefined();
      expect(result.status).toBe(17);
      expect(result.stdout).toBe('');
      expect(result.stderr).toBe('');
    } finally {
      fs.rmSync(cwd, { recursive: true, force: true });
    }
  }, 40_000);

  test.each<NodeJS.Signals>(
    process.platform === 'win32' ? ['SIGINT'] : ['SIGHUP', 'SIGINT', 'SIGTERM'],
  )(
    'exits on %s and releases its Go child while stdout stays paused',
    async (terminationSignal) => {
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
          child.kill(terminationSignal);
          const [exit] = await done;
          expect(exit).toEqual([null, terminationSignal]);
          // Windows child.kill() forcibly terminates Node and can kill its
          // entire job. On POSIX only Node is signalled, so Go can flush on EOF.
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
    },
    40_000,
  );
});

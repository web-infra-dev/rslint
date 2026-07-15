import { describe, test, expect } from '@rstest/core';
import fs from 'node:fs';
import os from 'node:os';
import { PassThrough } from 'node:stream';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { runEngine } from '../src/cli/engine.js';
import { ConfigModuleHost } from '../src/config/config-loader.js';
import { resolveRslintBinary } from '../src/internal/resolve-binary.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const FAKE_BIN = path.resolve(__dirname, './fixtures/fake-ipc-binary.cjs');
const CONFIG_RACE_BIN = path.resolve(
  __dirname,
  './fixtures/fake-config-activation-race.cjs',
);
const EXIT_DURING_CONFIG_ACTIVATION_BIN = path.resolve(
  __dirname,
  './fixtures/fake-exit-during-config-activation.cjs',
);

/**
 * Runs the engine against the fake IPC binary, which echoes the `init`
 * payload it received back through an `output` frame — letting the tests
 * assert on what actually crossed the wire, not on engine internals.
 */
async function runWithSink(sink: PassThrough): Promise<{
  exitCode: number;
  payload: { runtime?: { stdoutIsTTY?: boolean } };
}> {
  let captured = '';
  sink.on('data', (d: Buffer) => {
    captured += d.toString();
  });
  const exitCode = await runEngine({
    binPath: process.execPath,
    goArgs: [FAKE_BIN],
    stdout: sink,
    stderr: new PassThrough(),
  });
  return { exitCode, payload: JSON.parse(captured) };
}

describe('runEngine init payload TTY fact', () => {
  test('sends runtime.stdoutIsTTY=true when the output sink is a TTY', async () => {
    const sink = Object.assign(new PassThrough(), { isTTY: true });
    const { exitCode, payload } = await runWithSink(sink);
    expect(exitCode).toBe(0);
    expect(payload.runtime?.stdoutIsTTY).toBe(true);
  });

  test('sends runtime.stdoutIsTTY=false for a non-TTY sink', async () => {
    const { exitCode, payload } = await runWithSink(new PassThrough());
    expect(exitCode).toBe(0);
    expect(payload.runtime?.stdoutIsTTY).toBe(false);
  });
});

describe('runEngine live config predicates', () => {
  test('the real CLI evaluates each visited path once and reuses its selection', async () => {
    const root = fs.mkdtempSync(
      path.join(os.tmpdir(), 'rslint-cli-predicate-'),
    );
    const marker = path.join(root, 'predicate-calls.log');
    const stdout = new PassThrough();
    const stderr = new PassThrough();
    let output = '';
    let errors = '';
    stdout.on('data', (chunk: Buffer) => {
      output += chunk.toString();
    });
    stderr.on('data', (chunk: Buffer) => {
      errors += chunk.toString();
    });
    fs.writeFileSync(
      path.join(root, 'rslint.config.mjs'),
      [
        "import { appendFileSync } from 'node:fs';",
        "const marker = new URL('./predicate-calls.log', import.meta.url);",
        'export default [{',
        '  files: [(filePath) => {',
        '    appendFileSync(marker, `${filePath}\\n`);',
        "    return filePath.endsWith('keep.ts');",
        '  }],',
        "  rules: { 'no-debugger': 'error' },",
        '}];',
        '',
      ].join('\n'),
    );
    fs.writeFileSync(path.join(root, 'keep.ts'), 'debugger;\n');
    fs.writeFileSync(path.join(root, 'drop.ts'), 'debugger;\n');

    try {
      const exitCode = await runEngine({
        binPath: resolveRslintBinary(),
        goArgs: ['--start-time=0'],
        cwd: root,
        runtime: { singleThreaded: true },
        extraInit: {
          configDiscovery: { mode: 'auto', inputs: ['*.ts'] },
        },
        stdout,
        stderr,
      });
      expect(exitCode).toBe(1);
      expect(errors).toBe('');
      expect(output).toContain('keep.ts');
      expect(output).toContain('no-debugger');
      expect(output).not.toContain('drop.ts');

      const evaluated = fs
        .readFileSync(marker, 'utf8')
        .trim()
        .split('\n')
        .map((filePath) => path.basename(filePath))
        .sort();
      expect(evaluated).toEqual(['drop.ts', 'keep.ts', 'rslint.config.mjs']);
    } finally {
      fs.rmSync(root, { recursive: true, force: true });
    }
  });
});

describe('runEngine config activation', () => {
  test('disposes and never publishes a host whose prepare changes its config', async () => {
    const root = fs.mkdtempSync(
      path.join(os.tmpdir(), 'rslint-cli-config-activation-'),
    );
    const configPath = path.join(root, 'rslint.config.mjs');
    fs.writeFileSync(
      configPath,
      'export default [{ plugins: { local: { rules: { example: {} } } } }];\n',
    );
    const stdout = new PassThrough();
    let captured = '';
    let lintCalls = 0;
    let shutdownCalls = 0;
    stdout.on('data', (chunk: Buffer) => {
      captured += chunk.toString();
    });

    try {
      const exitCode = await runEngine({
        binPath: process.execPath,
        goArgs: [CONFIG_RACE_BIN, configPath],
        stdout,
        stderr: new PassThrough(),
        createPluginLintHost: async () => {
          fs.writeFileSync(configPath, '// changed by mocked worker prepare\n');
          return {
            async lint() {
              lintCalls++;
              return { results: ['stale-host-was-visible'] };
            },
            async shutdown() {
              shutdownCalls++;
            },
          };
        },
      });

      expect(exitCode).toBe(0);
      expect(captured).toContain('plugin host was being prepared');
      expect(captured).toContain(
        'pluginLint requested without an activated plugin host',
      );
      expect(lintCalls).toBe(0);
      expect(shutdownCalls).toBe(1);
    } finally {
      fs.rmSync(root, { recursive: true, force: true });
    }
  });

  test('disposes a plugin host that finishes after the Go child exits', async () => {
    const root = fs.mkdtempSync(
      path.join(os.tmpdir(), 'rslint-cli-config-exit-'),
    );
    const configPath = path.join(root, 'rslint.config.mjs');
    fs.writeFileSync(
      configPath,
      'export default [{ plugins: { local: { rules: { example: {} } } } }];\n',
    );
    let buildStarted = false;
    let shutdownCalls = 0;

    try {
      const exitCode = await runEngine({
        binPath: process.execPath,
        goArgs: [EXIT_DURING_CONFIG_ACTIVATION_BIN, configPath],
        stdout: new PassThrough(),
        stderr: new PassThrough(),
        createPluginLintHost: async () => {
          buildStarted = true;
          await new Promise((resolve) => setTimeout(resolve, 250));
          return {
            async lint() {
              return { results: [] };
            },
            async shutdown() {
              shutdownCalls++;
            },
          };
        },
      });

      expect(exitCode).toBe(0);
      expect(buildStarted).toBe(true);
      expect(shutdownCalls).toBe(1);
    } finally {
      fs.rmSync(root, { recursive: true, force: true });
    }
  });

  test('disposes a staged host before returning when Go exits during post-prepare verification', async () => {
    const root = fs.mkdtempSync(
      path.join(os.tmpdir(), 'rslint-cli-config-staged-exit-'),
    );
    const configPath = path.join(root, 'rslint.config.mjs');
    fs.writeFileSync(configPath, '// stable config bytes\n');
    let fingerprintReads = 0;
    let releasePostPrepare!: () => void;
    let markPostPrepareStarted!: () => void;
    const postPrepareStarted = new Promise<void>((resolve) => {
      markPostPrepareStarted = resolve;
    });
    const postPrepareGate = new Promise<void>((resolve) => {
      releasePostPrepare = resolve;
    });
    const configModuleHost = new ConfigModuleHost({
      loadCached: async () => [
        { plugins: { local: { rules: { example: {} } } } },
      ],
      readSource: async (sourcePath) => {
        fingerprintReads++;
        if (fingerprintReads === 4) {
          markPostPrepareStarted();
          await postPrepareGate;
        }
        return fs.promises.readFile(sourcePath);
      },
    });
    let shutdownCalls = 0;

    try {
      const run = runEngine({
        binPath: process.execPath,
        goArgs: [EXIT_DURING_CONFIG_ACTIVATION_BIN, configPath],
        stdout: new PassThrough(),
        stderr: new PassThrough(),
        configModuleHost,
        createPluginLintHost: async () => ({
          async lint() {
            return { results: [] };
          },
          async shutdown() {
            shutdownCalls++;
          },
        }),
      });

      await postPrepareStarted;
      const exitCode = await run;
      expect(exitCode).toBe(0);
      expect(shutdownCalls).toBe(1);
    } finally {
      releasePostPrepare();
      await new Promise<void>((resolve) => setImmediate(resolve));
      fs.rmSync(root, { recursive: true, force: true });
    }
  });
});

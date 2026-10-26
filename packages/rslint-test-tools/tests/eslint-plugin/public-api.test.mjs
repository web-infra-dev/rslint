import { Rslint } from '@rslint/core';
import { describe, test, expect } from 'rstack/test';
import { spawn } from 'node:child_process';
import { mkdtemp, mkdir, writeFile } from 'node:fs/promises';
import path from 'node:path';
import os from 'node:os';
import { pathToFileURL } from 'node:url';
import { cleanupTempDir } from '../cli/js-config/helpers.js';

const eslintPluginFixturesDir = path.resolve(import.meta.dirname, 'fixtures');
// These fixtures prove that the public API lets Node exit. Success is the
// validated close event; this only releases a process still alive after 30m.
const EXIT_FIXTURE_DEAD_PROCESS_WATCHDOG_MS = 30 * 60_000;
const EXIT_FIXTURE_OUTER_DEADLOCK_SENTINEL_MS = 35 * 60_000;
const EXIT_FIXTURE_OUTPUT_LIMIT_BYTES = 1024 * 1024;

async function runExitFixture(scriptName, successMarker) {
  const script = path.resolve(import.meta.dirname, '../fixtures', scriptName);
  const child = spawn(process.execPath, [script], {
    cwd: path.resolve(import.meta.dirname, '../..'),
    stdio: ['ignore', 'pipe', 'pipe'],
  });
  const stdoutChunks = [];
  const stderrChunks = [];
  let spawnError;
  let timedOut = false;
  let outputBytes = 0;
  let outputOverflow = false;

  child.stdout.on('data', (chunk) => {
    outputBytes += chunk.length;
    if (outputBytes > EXIT_FIXTURE_OUTPUT_LIMIT_BYTES) {
      outputOverflow = true;
      child.kill('SIGKILL');
      return;
    }
    stdoutChunks.push(chunk);
  });
  child.stderr.on('data', (chunk) => {
    outputBytes += chunk.length;
    if (outputBytes > EXIT_FIXTURE_OUTPUT_LIMIT_BYTES) {
      outputOverflow = true;
      child.kill('SIGKILL');
      return;
    }
    stderrChunks.push(chunk);
  });
  child.on('error', (error) => {
    spawnError = error;
  });

  const timer = setTimeout(() => {
    timedOut = true;
    child.kill('SIGKILL');
  }, EXIT_FIXTURE_DEAD_PROCESS_WATCHDOG_MS);
  timer.unref();

  const { code, signal } = await new Promise((resolve) => {
    child.on('close', (code, signal) => {
      resolve({ code, signal });
    });
  });
  clearTimeout(timer);
  const stdout = Buffer.concat(stdoutChunks).toString('utf8');
  const stderr = Buffer.concat(stderrChunks).toString('utf8');

  const details =
    `fixture=${scriptName}, code=${String(code)}, signal=${String(signal)}, ` +
    `timedOut=${String(timedOut)}, outputOverflow=${String(outputOverflow)}, ` +
    `spawnError=${String(spawnError)}\n` +
    `stdout:\n${stdout}\nstderr:\n${stderr}`;
  expect(timedOut, details).toBe(false);
  expect(outputOverflow, details).toBe(false);
  expect(spawnError, details).toBeUndefined();
  expect(signal, details).toBeNull();
  expect(code, details).toBe(0);
  expect(stdout.trim(), details).toBe(successMarker);
  expect(stderr.trim(), details).toBe('');
  expect(
    `${stdout}\n${stderr}`,
    `fixture emitted an abnormal worker marker\n${details}`,
  ).not.toMatch(
    /CLOSE_GATE_LEAKED|task_timeout|worker_crashed|timed out after|EXPECTED-NATIVE-ABORT/,
  );
}

describe('Rslint community plugin API', () => {
  test('lintText runs a local community plugin Program listener and returns its fix output', async () => {
    const tmp = await mkdtemp(path.join(os.tmpdir(), 'rslint-api-plugin-'));
    const pluginUrl = pathToFileURL(
      path.join(eslintPluginFixturesDir, 'local-plugin.mjs'),
    ).href;
    await writeFile(
      path.join(tmp, 'rslint.config.mjs'),
      [
        `import local from ${JSON.stringify(pluginUrl)};`,
        'export default [{',
        '  plugins: { local },',
        '  rules: {',
        "    'local/program-listener': 'error',",
        "    'local/prefer-array-some': 'error',",
        '  },',
        '}];',
        '',
      ].join('\n'),
    );

    const rslint = new Rslint({ cwd: tmp, fix: true });
    try {
      const [result] = await rslint.lintText(
        'const found = [1].filter(Boolean);\n',
        { filePath: 'probe.ts' },
      );
      expect(result.messages.map((message) => message.ruleId)).toEqual([
        'local/program-listener',
      ]);
      expect(result.fixableErrorCount).toBe(0);
      expect(result.output).toBe('const found = [1].some(Boolean);\n');
    } finally {
      await rslint.close();
      await cleanupTempDir(tmp);
    }
  });

  test('lintFiles routes community plugins through one multi-config host', async () => {
    const tmp = await mkdtemp(
      path.join(os.tmpdir(), 'rslint-api-plugin-routing-'),
    );
    const xDir = path.join(tmp, 'packages', 'x');
    const yDir = path.join(tmp, 'packages', 'y');
    const pluginXUrl = pathToFileURL(
      path.join(eslintPluginFixturesDir, 'cfgX', 'plugin-x.mjs'),
    ).href;
    const pluginYUrl = pathToFileURL(
      path.join(eslintPluginFixturesDir, 'cfgY', 'plugin-y.mjs'),
    ).href;
    try {
      await mkdir(xDir, { recursive: true });
      await mkdir(yDir, { recursive: true });
      await writeFile(
        path.join(xDir, 'rslint.config.mjs'),
        `import plugin from ${JSON.stringify(pluginXUrl)};\n` +
          "export default [{ plugins: { px: plugin }, rules: { 'px/no-foo': 'error' } }];\n",
      );
      await writeFile(
        path.join(yDir, 'rslint.config.mjs'),
        `import plugin from ${JSON.stringify(pluginYUrl)};\n` +
          "export default [{ plugins: { py: plugin }, rules: { 'py/no-bar': 'error' } }];\n",
      );
      const source = 'const foo = 1; const bar = 2;\n';
      await writeFile(path.join(xDir, 'index.ts'), source);
      await writeFile(path.join(yDir, 'index.ts'), source);

      const rslint = new Rslint({ cwd: tmp });
      try {
        const results = await rslint.lintFiles('**/*.ts');
        const byRelativePath = new Map(
          results.map((result) => [
            path.relative(tmp, result.filePath),
            result,
          ]),
        );
        expect(
          byRelativePath
            .get(path.join('packages', 'x', 'index.ts'))
            .messages.map((message) => message.ruleId),
        ).toEqual(['px/no-foo']);
        expect(
          byRelativePath
            .get(path.join('packages', 'y', 'index.ts'))
            .messages.map((message) => message.ruleId),
        ).toEqual(['py/no-bar']);
      } finally {
        await rslint.close();
      }
    } finally {
      await cleanupTempDir(tmp);
    }
  });

  test('repeated API calls fresh-load changed config and plugin topology', async () => {
    const tmp = await mkdtemp(
      path.join(os.tmpdir(), 'rslint-api-fresh-config-'),
    );
    const configPath = path.join(tmp, 'rslint.config.mjs');
    const pluginXUrl = pathToFileURL(
      path.join(eslintPluginFixturesDir, 'cfgX', 'plugin-x.mjs'),
    ).href;
    const pluginYUrl = pathToFileURL(
      path.join(eslintPluginFixturesDir, 'cfgY', 'plugin-y.mjs'),
    ).href;
    const writeConfig = async (pluginUrl, ruleName) => {
      await writeFile(
        configPath,
        `import plugin from ${JSON.stringify(pluginUrl)};\n` +
          `export default [{ plugins: { p: plugin }, rules: { ${JSON.stringify(`p/${ruleName}`)}: 'error' } }];\n`,
      );
    };
    try {
      await writeConfig(pluginXUrl, 'no-foo');
      const rslint = new Rslint({ cwd: tmp });
      try {
        const source = 'const foo = 1; const bar = 2;\n';
        const first = await rslint.lintText(source, { filePath: 'index.ts' });
        expect(first[0].messages.map((message) => message.ruleId)).toEqual([
          'p/no-foo',
        ]);

        await writeConfig(pluginYUrl, 'no-bar');
        const second = await rslint.lintText(source, {
          filePath: 'index.ts',
        });
        expect(second[0].messages.map((message) => message.ruleId)).toEqual([
          'p/no-bar',
        ]);
      } finally {
        await rslint.close();
      }
    } finally {
      await cleanupTempDir(tmp);
    }
  });

  // Regression for the unref() fix: a script that lints and never calls close()
  // must still let the Node process exit. Without unref the resident `--api`
  // child + its stdio pipes keep the event loop alive and the process hangs.
  test(
    'lintText without close() lets the process exit on its own (no hang)',
    async () => {
      await runExitFixture('no-close-exit.mjs', 'FIXTURE_OK:no-close');
    },
    EXIT_FIXTURE_OUTER_DEADLOCK_SENTINEL_MS,
  );

  test(
    'community plugin host initializes and shuts down before lintText returns',
    async () => {
      await runExitFixture(
        'no-close-plugin-exit.mjs',
        'FIXTURE_OK:no-close-plugin',
      );
    },
    EXIT_FIXTURE_OUTER_DEADLOCK_SENTINEL_MS,
  );

  test(
    'community plugin host shuts down when the Go lint request fails',
    async () => {
      await runExitFixture(
        'no-close-plugin-error-exit.mjs',
        'FIXTURE_OK:no-close-plugin-error',
      );
    },
    EXIT_FIXTURE_OUTER_DEADLOCK_SENTINEL_MS,
  );

  test(
    'close shuts down an active community plugin host',
    async () => {
      await runExitFixture(
        'close-active-plugin-exit.mjs',
        'FIXTURE_OK:close-active-plugin',
      );
    },
    EXIT_FIXTURE_OUTER_DEADLOCK_SENTINEL_MS,
  );
});

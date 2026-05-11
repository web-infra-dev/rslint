/**
 * End-to-end test for monorepo multi-version eslintPlugins dispatch.
 *
 * Layout under a temp dir:
 *
 *   <tmp>/
 *   ├─ packages/a/
 *   │  ├─ rslint.config.mjs               imports 'fake-plugin' (v1 here)
 *   │  ├─ src/index.ts                    `const x = 1; const y = 2;`
 *   │  └─ node_modules/fake-plugin/
 *   │     ├─ package.json                 v1.0.0
 *   │     └─ index.cjs                    rules['no-x'] reports
 *   │                                     'PluginA-v1: identifier x at path A'
 *   └─ packages/b/                       symmetric, v2.0.0 + 'PluginB-v2: ...'
 *
 * The two `packages/{a,b}/rslint.config.mjs` use the SAME `fake` prefix and
 * the SAME `fake-plugin` specifier, but each resolves the specifier to its
 * own `node_modules` — so the two configs end up with two PHYSICALLY
 * DIFFERENT plugin files.
 *
 * If per-file dispatch works, A's source file is linted with the v1 plugin
 * (its message text wins) and B's source is linted with v2. If the
 * pipeline collapses by prefix anywhere along the path (Go side, Node
 * merge, Worker rule resolution), one plugin's message would leak across
 * both files — that's the regression this test guards.
 */

import { describe, test, expect, beforeAll, afterAll } from '@rstest/core';
import { spawn } from 'node:child_process';
import path from 'node:path';
import fs from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { createRequire } from 'node:module';

// require.resolve('@rslint/core/bin') gives us the absolute path of
// `packages/rslint/bin/rslint.cjs`, which then locates the Go binary and
// hands off to engine.ts. Same entry the user invokes via `npx rslint`.
const requireFromHere = createRequire(import.meta.url);
const RSLINT_BIN = requireFromHere.resolve('@rslint/core/bin');

interface CliResult {
  exitCode: number;
  stdout: string;
  stderr: string;
}

function runRslint(args: string[], cwd: string): Promise<CliResult> {
  return new Promise((resolve) => {
    // Strip color env so jsonline output stays parseable byte-for-byte.
    const { GITHUB_ACTIONS, FORCE_COLOR, ...cleanEnv } = process.env;
    void GITHUB_ACTIONS;
    void FORCE_COLOR;
    const child = spawn(process.execPath, [RSLINT_BIN, ...args], {
      cwd,
      stdio: ['pipe', 'pipe', 'pipe'],
      env: { ...cleanEnv, NO_COLOR: '1' },
    });
    let stdout = '';
    let stderr = '';
    child.stdout?.on('data', (d: Buffer) => {
      stdout += d.toString();
    });
    child.stderr?.on('data', (d: Buffer) => {
      stderr += d.toString();
    });
    child.on('close', (code) => {
      resolve({ exitCode: code ?? 0, stdout, stderr });
    });
  });
}

/**
 * One fake plugin file, keyed by an arbitrary `label` so we can verify
 * which instance produced a diagnostic. Stays in CJS to keep the fixture
 * minimal — no `"type": "module"` package boundaries needed.
 */
function fakePluginCJS(label: string, version: string): string {
  return `
'use strict';
module.exports = {
  meta: { name: 'fake-plugin', version: '${version}' },
  rules: {
    'no-x': {
      meta: {
        type: 'problem',
        schema: [],
        messages: {
          found: '${label}: identifier x',
        },
      },
      create(ctx) {
        return {
          Identifier(node) {
            if (node.name === 'x') {
              ctx.report({ node, messageId: 'found' });
            }
          },
        };
      },
    },
  },
};
`;
}

const RSLINT_CONFIG = `
import fakePlugin from 'fake-plugin';
export default [
  {
    files: ['src/**/*.ts'],
    plugins: ['fake'],
    eslintPlugins: { fake: fakePlugin },
    rules: { 'fake/no-x': 'error' },
  },
];
`;

// Two identifiers `x` (the binding and the use). The fake plugin reports
// once per Identifier whose name === 'x', so each file should produce
// exactly 2 diagnostics. Using 2 (not 1) lets us catch a double-dispatch
// bug where a single file is processed by BOTH plugin instances — which
// would inflate the count without changing the per-message text and
// could otherwise sneak past a single-diagnostic-per-file assertion.
const SOURCE_TS = `const x = 1; foo(x);\n`;

async function writeFiles(
  root: string,
  files: Record<string, string>,
): Promise<void> {
  for (const [rel, content] of Object.entries(files)) {
    const full = path.join(root, rel);
    await fs.mkdir(path.dirname(full), { recursive: true });
    await fs.writeFile(full, content, 'utf8');
  }
}

describe('monorepo multi-version eslintPlugins', () => {
  let fixtureDir: string;

  beforeAll(async () => {
    fixtureDir = await fs.mkdtemp(path.join(tmpdir(), 'rslint-multi-version-'));
    await writeFiles(fixtureDir, {
      'package.json': JSON.stringify({
        name: 'rslint-multi-version-fixture',
        private: true,
      }),

      // packages/a — fake-plugin@1, message tagged 'PluginA-v1'
      'packages/a/node_modules/fake-plugin/package.json': JSON.stringify({
        name: 'fake-plugin',
        version: '1.0.0',
        main: 'index.cjs',
      }),
      'packages/a/node_modules/fake-plugin/index.cjs': fakePluginCJS(
        'PluginA-v1',
        '1.0.0',
      ),
      'packages/a/rslint.config.mjs': RSLINT_CONFIG,
      'packages/a/src/index.ts': SOURCE_TS,

      // packages/b — fake-plugin@2, message tagged 'PluginB-v2'
      'packages/b/node_modules/fake-plugin/package.json': JSON.stringify({
        name: 'fake-plugin',
        version: '2.0.0',
        main: 'index.cjs',
      }),
      'packages/b/node_modules/fake-plugin/index.cjs': fakePluginCJS(
        'PluginB-v2',
        '2.0.0',
      ),
      'packages/b/rslint.config.mjs': RSLINT_CONFIG,
      'packages/b/src/index.ts': SOURCE_TS,
    });
  });

  afterAll(async () => {
    if (fixtureDir) {
      await fs.rm(fixtureDir, { recursive: true, force: true });
    }
  });

  test("each file is linted by its OWN config's plugin instance", async () => {
    const result = await runRslint(
      [
        'packages/a/src/index.ts',
        'packages/b/src/index.ts',
        '--format=jsonline',
      ],
      fixtureDir,
    );

    if (result.exitCode !== 1) {
      throw new Error(
        `unexpected exit code ${result.exitCode}\nstdout:\n${result.stdout}\nstderr:\n${result.stderr}`,
      );
    }

    const lines = result.stdout.trim().split('\n').filter(Boolean);
    const diagnostics = lines.map(
      (l) =>
        JSON.parse(l) as {
          ruleName: string;
          message: string;
          filePath: string;
        },
    );

    // Source has 2 `x` identifiers (decl + use). With correct per-file
    // dispatch we expect exactly 4 diagnostics total — 2 from file A
    // (both produced by PluginA-v1) and 2 from file B (both PluginB-v2).
    //
    // Failure shapes this asserts against:
    //   - Total != 4 → either the dispatcher dropped a rule listener call,
    //     or a file got double-dispatched to BOTH plugin instances (which
    //     would make total = 6 or 8).
    //   - Per-file count != 2 → the same double-dispatch bug, but only
    //     hitting one file (count would be 4 on that file, 0 on the other).
    //   - Any single message text mismatch → cross-contamination: file A
    //     was linted by PluginB or vice versa.
    expect(diagnostics).toHaveLength(4);

    const aDiags = diagnostics.filter((d) =>
      d.filePath.endsWith('packages/a/src/index.ts'),
    );
    const bDiags = diagnostics.filter((d) =>
      d.filePath.endsWith('packages/b/src/index.ts'),
    );
    expect(aDiags).toHaveLength(2);
    expect(bDiags).toHaveLength(2);

    // Every diagnostic on A must be PluginA-v1's text — exact match, not
    // contains, because we want any drift caught.
    for (const d of aDiags) {
      expect(d.ruleName).toBe('fake/no-x');
      expect(d.message).toBe('PluginA-v1: identifier x');
    }
    for (const d of bDiags) {
      expect(d.ruleName).toBe('fake/no-x');
      expect(d.message).toBe('PluginB-v2: identifier x');
    }

    // Belt and suspenders: zero PluginB messages on file A, zero PluginA
    // messages on file B. If either of the per-message checks above
    // failed open (e.g. message ever became empty), these would catch it.
    expect(aDiags.filter((d) => d.message.includes('PluginB-v2'))).toHaveLength(
      0,
    );
    expect(bDiags.filter((d) => d.message.includes('PluginA-v1'))).toHaveLength(
      0,
    );
  });
});

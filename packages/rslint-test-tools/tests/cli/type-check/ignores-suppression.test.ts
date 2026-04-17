import { describe, test, expect } from '@rstest/core';
import { runRslint, createTempDir, cleanupTempDir, TS_CONFIG } from './helpers';

// `ignores` must hide files from --type-check, matching ESLint v10 semantics.
// These tests guard the end-to-end CLI path (config → buildFileFilters →
// RunLinter). Unit tests inside internal/linter only exercise the filter
// callback; they cannot catch regressions in the cmd wiring that composes it.

interface Diagnostic {
  ruleName: string;
  filePath: string;
  severity?: string;
}

/**
 * Run rslint --type-check twice: once with --format jsonline for exact
 * per-file diagnostic parsing, once with the default format for the summary
 * line ("linted N files"). The default format does not emit structured
 * diagnostics, and jsonline suppresses the summary, so both runs are needed
 * for precise assertions.
 */
async function lintTypeCheck(
  tempDir: string,
  extraArgs: string[] = [],
): Promise<{
  diagnostics: Diagnostic[];
  lintedFileCount: number;
  exitCode: number;
}> {
  const jsonRun = await runRslint(
    ['--type-check', '--format', 'jsonline', ...extraArgs],
    tempDir,
  );
  const diagnostics: Diagnostic[] = jsonRun.stdout
    .trim()
    .split('\n')
    .filter((l) => l.trim().startsWith('{'))
    .map((l) => JSON.parse(l) as Diagnostic);

  const summaryRun = await runRslint(['--type-check', ...extraArgs], tempDir);
  const combined = `${summaryRun.stdout}\n${summaryRun.stderr}`;
  const countMatch = combined.match(/linted (\d+) files?/);
  if (!countMatch) {
    throw new Error(
      `Could not parse linted-file count from summary output:\nSTDOUT:\n${summaryRun.stdout}\nSTDERR:\n${summaryRun.stderr}`,
    );
  }

  return {
    diagnostics,
    lintedFileCount: parseInt(countMatch[1]!, 10),
    exitCode: jsonRun.exitCode,
  };
}

/**
 * Build an rslint.config.mjs with a given `ignores` array. Keeps everything
 * else (files glob, parserOptions, no enabled rules) constant so the only
 * variable under test is the ignores behavior.
 */
function makeConfigWithIgnores(ignores: string[]): string {
  return `export default [
  { ignores: ${JSON.stringify(ignores)} },
  {
    files: ['**/*.ts'],
    languageOptions: { parserOptions: { project: ['./tsconfig.json'] } },
  }
];
`;
}

describe('--type-check + config ignores', () => {
  test('ignored file produces zero diagnostics; non-ignored file still does', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.mjs': makeConfigWithIgnores(['ignored/**']),
      // Ignored file: has a TS2322 that would otherwise fire
      'ignored/bad.ts': "const x: number = 'from-ignored';\n",
      // Control file: same kind of error, should still fire
      'src/bad.ts': "const y: number = 'from-src';\n",
    });
    try {
      const r = await lintTypeCheck(tempDir);

      // Guard against trivial pass: the non-ignored file MUST produce a TS
      // diagnostic. If it doesn't, the test setup is broken (e.g. tsconfig
      // not picked up) and any assertion about the ignored file is vacuous.
      const srcDiags = r.diagnostics.filter((d) =>
        d.filePath?.includes('src/bad.ts'),
      );
      expect(srcDiags.length).toBeGreaterThan(0);
      expect(srcDiags.some((d) => d.ruleName.includes('TS'))).toBe(true);

      // The actual assertion: zero diagnostics point at the ignored file.
      const ignoredDiags = r.diagnostics.filter((d) =>
        d.filePath?.includes('ignored/bad.ts'),
      );
      expect(ignoredDiags).toEqual([]);

      // Counts must reflect the filter: exactly one file was linted.
      // If ignores leaks, count would be 2 (or more if tsgolint pulled extras).
      expect(r.lintedFileCount).toBe(1);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('ignored file stays in TS program as import context but emits no own diagnostics', async () => {
    // Scenario: caller.ts imports from util.ts; util.ts is ignored. The TS
    // compiler must still load util.ts to type-check caller.ts's call site,
    // but util.ts itself must not emit diagnostics and must not count.
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.mjs': makeConfigWithIgnores(['lib/**']),
      // Ignored file: exports a typed function. Contains its own TS error
      // that would fire if it were not ignored.
      'lib/util.ts': [
        'export function greet(name: string): string {',
        '  return `hi ${name}`;',
        '}',
        "const shouldBeIgnored: number = 'not a number';", // TS2322 if linted
        '',
      ].join('\n'),
      // Non-ignored caller passes wrong arg type — TS must still catch this,
      // which proves util.ts is loaded into the program as context.
      'src/caller.ts': "import { greet } from '../lib/util';\ngreet(42);\n",
    });
    try {
      const r = await lintTypeCheck(tempDir);

      // Context check: caller.ts must get its TS2345 (wrong arg type). This
      // only works if util.ts's types were resolved → proves ignored file
      // stayed in the TS program.
      const callerDiags = r.diagnostics.filter((d) =>
        d.filePath?.includes('src/caller.ts'),
      );
      expect(callerDiags.length).toBeGreaterThan(0);
      const callerTsDiags = callerDiags.filter((d) =>
        d.ruleName.startsWith('TypeScript('),
      );
      expect(callerTsDiags.length).toBeGreaterThan(0);

      // Assertion: the ignored file's own TS2322 must NOT surface.
      const utilDiags = r.diagnostics.filter((d) =>
        d.filePath?.includes('lib/util.ts'),
      );
      expect(utilDiags).toEqual([]);

      // Count: only caller.ts counts as "linted".
      expect(r.lintedFileCount).toBe(1);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('multi-config: root ignores filter files covered by child tsconfig', async () => {
    // Root rslint.config.mjs globally ignores packages/child/**. The child
    // package has its own tsconfig that includes those files. The ignores
    // must still win — this is the ESLint v10 "global ignores" semantics.
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.mjs': `export default [
  { ignores: ['packages/child/**'] },
  {
    files: ['**/*.ts'],
    languageOptions: { parserOptions: { project: ['./tsconfig.json', './packages/child/tsconfig.json'] } },
  }
];
`,
      'packages/child/tsconfig.json': JSON.stringify({
        compilerOptions: {
          target: 'ES2020',
          module: 'ESNext',
          strict: true,
          moduleResolution: 'node',
        },
        include: ['**/*.ts'],
      }),
      // File covered by child's tsconfig, but globally ignored
      'packages/child/a.ts': "const x: number = 'ignored-by-root';\n",
      // File not covered by the root ignore — must still be linted
      'src/ok.ts': "const y: number = 'from-src';\n",
    });
    try {
      const r = await lintTypeCheck(tempDir);

      // Control: the non-ignored file must still produce diagnostics.
      const srcDiags = r.diagnostics.filter((d) =>
        d.filePath?.includes('src/ok.ts'),
      );
      expect(srcDiags.length).toBeGreaterThan(0);

      // Assertion: no diagnostic anywhere under packages/child/.
      const childDiags = r.diagnostics.filter((d) =>
        d.filePath?.includes('packages/child/'),
      );
      expect(childDiags).toEqual([]);

      // Count: only src/ok.ts, even though packages/child/a.ts is in the
      // child's tsconfig include.
      expect(r.lintedFileCount).toBe(1);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });
});

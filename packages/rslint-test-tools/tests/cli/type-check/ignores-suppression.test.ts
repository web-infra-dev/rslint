import { describe, test, expect } from '@rstest/core';
import { runRslint, createTempDir, cleanupTempDir, TS_CONFIG } from './helpers';

// `ignores` controls the LINT phase only. `--type-check` is program-level
// and aligned with `tsc --noEmit`: every file that the TypeScript program
// loads is checked, regardless of rslint's `ignores` configuration. This
// supersedes the earlier behaviour from PR #681 (where `ignores` also hid
// type-check diagnostics) — see "Alignment with tsc --noEmit" in
// website/docs/en/guide/type-checking.md.
//
// These tests guard the end-to-end CLI path (config → buildFileFilters →
// RunLinter): config `ignores` continue to remove a file from the lint-rule
// pass (LintedFileCount, lint-rule diagnostics) while leaving its
// type-check diagnostics intact.

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
  test('ignored file still produces type-check diagnostics; lint-file count excludes it', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.mjs': makeConfigWithIgnores(['ignored/**']),
      // Ignored file: has a TS2322 — under tsc-aligned semantics this MUST
      // surface because the file is in the TS program.
      'ignored/bad.ts': "const x: number = 'from-ignored';\n",
      // Control file: same kind of error, also reported.
      'src/bad.ts': "const y: number = 'from-src';\n",
    });
    try {
      const r = await lintTypeCheck(tempDir);

      // Control: the non-ignored file produces a TS diagnostic.
      const srcDiags = r.diagnostics.filter((d) =>
        d.filePath?.includes('src/bad.ts'),
      );
      expect(srcDiags.length).toBeGreaterThan(0);
      expect(srcDiags.some((d) => d.ruleName.includes('TS'))).toBe(true);

      // The new contract: the ignored file's type-check diagnostic still
      // surfaces (program-level check is not gated by ignores).
      const ignoredDiags = r.diagnostics.filter((d) =>
        d.filePath?.includes('ignored/bad.ts'),
      );
      expect(ignoredDiags.length).toBeGreaterThan(0);
      expect(ignoredDiags.some((d) => d.ruleName.includes('TS2322'))).toBe(
        true,
      );

      // LintedFileCount counts the lint-rule pass only — ignored files do
      // not contribute. Only src/bad.ts is "linted".
      expect(r.lintedFileCount).toBe(1);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('ignored file as import context: caller is checked AND ignored file own type errors surface', async () => {
    // Scenario: caller.ts imports from util.ts; util.ts is `ignores`d. The
    // TypeScript compiler loads util.ts to type-check caller's call site
    // (proving the import context works). Under the tsc-aligned design,
    // util.ts's own type errors are also reported.
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.mjs': makeConfigWithIgnores(['lib/**']),
      // Ignored file with its own TS error.
      'lib/util.ts': [
        'export function greet(name: string): string {',
        '  return `hi ${name}`;',
        '}',
        "const shouldBeReported: number = 'not a number';", // TS2322
        '',
      ].join('\n'),
      // Non-ignored caller: passes wrong arg type — TS must catch this,
      // which proves util.ts is loaded into the program as context.
      'src/caller.ts': "import { greet } from '../lib/util';\ngreet(42);\n",
    });
    try {
      const r = await lintTypeCheck(tempDir);

      // Context check: caller.ts gets its TS2345 (wrong arg type).
      const callerDiags = r.diagnostics.filter((d) =>
        d.filePath?.includes('src/caller.ts'),
      );
      expect(callerDiags.length).toBeGreaterThan(0);
      const callerTsDiags = callerDiags.filter((d) =>
        d.ruleName.startsWith('TypeScript('),
      );
      expect(callerTsDiags.length).toBeGreaterThan(0);

      // New contract: util.ts's own TS2322 surfaces despite `ignores`.
      const utilDiags = r.diagnostics.filter((d) =>
        d.filePath?.includes('lib/util.ts'),
      );
      expect(utilDiags.length).toBeGreaterThan(0);
      expect(utilDiags.some((d) => d.ruleName.includes('TS2322'))).toBe(true);

      // Lint-rule pass excludes the ignored file: count is 1 (caller.ts).
      expect(r.lintedFileCount).toBe(1);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('multi-config: root ignores excludes files from lint-rule pass but type-check still reports them', async () => {
    // Root rslint.config.mjs globally ignores packages/child/**. The child
    // package has its own tsconfig that includes those files. The ignores
    // remove the child files from lint rules, but type-check sees the full
    // program and reports their type errors — matching `tsc --noEmit`.
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
          moduleResolution: 'bundler',
        },
        include: ['**/*.ts'],
      }),
      // File covered by child's tsconfig, globally ignored.
      'packages/child/a.ts': "const x: number = 'ignored-by-root';\n",
      // File not covered by the root ignore.
      'src/ok.ts': "const y: number = 'from-src';\n",
    });
    try {
      const r = await lintTypeCheck(tempDir);

      // Control: the non-ignored file produces diagnostics.
      const srcDiags = r.diagnostics.filter((d) =>
        d.filePath?.includes('src/ok.ts'),
      );
      expect(srcDiags.length).toBeGreaterThan(0);

      // New contract: type-check diagnostics for the ignored child file
      // still surface.
      const childDiags = r.diagnostics.filter((d) =>
        d.filePath?.includes('packages/child/'),
      );
      expect(childDiags.length).toBeGreaterThan(0);
      expect(childDiags.some((d) => d.ruleName.includes('TS2322'))).toBe(true);

      // Lint-rule pass excludes the ignored child file: count is 1 (src/ok.ts).
      expect(r.lintedFileCount).toBe(1);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });
});

import { describe, test, expect } from '@rstest/core';
import path from 'node:path';

import {
  runRslint,
  createTempDir,
  cleanupTempDir,
} from '../js-config/helpers.js';
import fs from 'node:fs/promises';

// End-to-end for the type snapshot bridge with a third-party-style JS
// type-aware plugin running in the worker: Go builds the per-file snapshot,
// ships it (CLI binary frame), the worker reconstructs parserServices, and the
// rule's getTypeAtLocation / isUnion / getUndefinedType queries resolve through
// it. The rule reads context.sourceCode.parserServices directly — the substance
// of ESLintUtils.getParserServices (@typescript-eslint/utils is a stub in this
// repo). The native typescript-eslint rules (Go) don't exercise this path; this
// is the only e2e that drives a JS type-aware rule over the bridge.
const FIXTURE_DIR = path.resolve(import.meta.dirname);

interface Diag {
  ruleName: string;
  message: string;
  filePath: string;
  range: {
    start: { line: number; column: number };
    end: { line: number; column: number };
  };
}

function parseDiags(stdout: string): Diag[] {
  return stdout
    .trim()
    .split('\n')
    .filter((l) => l.trim().startsWith('{'))
    .map((l) => JSON.parse(l) as Diag);
}

describe('type-aware plugin over the worker snapshot bridge (e2e)', () => {
  test('reports exactly the union-with-undefined variable, via the bridge', async () => {
    const res = await runRslint(['--format', 'jsonline'], FIXTURE_DIR);

    // The worker must not have crashed (a parserServices/bridge failure would
    // surface as a non-zero exit or a stderr error, not a clean diagnostic).
    expect(res.stderr).not.toMatch(
      /parserServices|getParserServices|TypeError/,
    );

    const diags = parseDiags(res.stdout).filter(
      (d) =>
        d.ruleName === 'ta/no-undefined-union' &&
        d.filePath.endsWith('union-undefined.ts'),
    );

    // Only `withUndefined` (line 5) is a union containing undefined.
    // `withoutUndefined` is a union WITHOUT undefined; `plain` is not a union.
    expect(diags).toHaveLength(1);
    expect(diags[0].range.start.line).toBe(5);
  });

  test('report-type-shape round-trips array / intersection / callable layout via the bridge', async () => {
    const res = await runRslint(['--format', 'jsonline'], FIXTURE_DIR);
    expect(res.stderr).not.toMatch(
      /parserServices|getParserServices|TypeError/,
    );

    // name → tags, from each "name: tags" message on the type-shapes fixture.
    const byVar = new Map(
      parseDiags(res.stdout)
        .filter(
          (d) =>
            d.ruleName === 'ta/report-type-shape' &&
            d.filePath.endsWith('type-shapes.ts'),
        )
        .map((d) => {
          const idx = d.message.indexOf(': ');
          return [d.message.slice(0, idx), d.message.slice(idx + 2)];
        }),
    );

    // Each kind reads a distinct snapshot type-block layout; a decode drift
    // would change the tag or drop the report.
    expect(byVar.get('unionVar')).toBe('union:2');
    expect(byVar.get('interVar')).toBe('intersection:2');
    expect(byVar.get('arrayVar')).toBe('array:1');
    expect(byVar.get('fnVar')).toBe('callable:1');
    // `plainVar` (a string literal) has no structural shape → not reported.
    expect(byVar.has('plainVar')).toBe(false);
  });

  test('real rule across span-boundary identifiers: every variant HITs including the escaped identifier (decoded-length keying)', async () => {
    const res = await runRslint(['--format', 'jsonline'], FIXTURE_DIR);
    expect(res.stderr).not.toMatch(
      /parserServices|getParserServices|TypeError/,
    );

    const names = parseDiags(res.stdout)
      .filter(
        (d) =>
          d.ruleName === 'ta/no-undefined-union' &&
          d.filePath.endsWith('span-edge.ts'),
      )
      .map((d) => d.message.match(/Variable '(.+?)'/)?.[1])
      .filter((n): n is string => !!n)
      .sort();

    // Both Go (snapshot.go) and the worker key identifiers on the DECODED name
    // length (range[0]+name.length), so the REAL rule HITs every span-boundary
    // variant — annotated, no-initializer, the annotation-shape variants, AND the
    // escaped identifier `esc` whose source token overshoots its decoded name.
    // (esc once MISSed; decoded-length keying fixed it.) Asserted BY NAME so the
    // fixture's line numbers / formatting can shift without breaking this.
    expect(names).toEqual([
      'annotated',
      'esc',
      'multiLine',
      'noInit',
      'spaced',
    ]);
  });

  test('--fix multi-pass: a first-pass fix changes a value type and the second pass type-aware rule sees the NEW type from a fresh snapshot', async () => {
    const pluginPath = path.resolve(FIXTURE_DIR, 'type-aware-plugin.mjs');
    const dir = await createTempDir({
      'tsconfig.json': JSON.stringify({
        compilerOptions: { strict: true },
        include: ['*.ts'],
      }),
      'rslint.config.mjs': `import ta from ${JSON.stringify(pluginPath)};
export default [{
  files: ['*.ts'],
  languageOptions: { parserOptions: { projectService: false, project: ['./tsconfig.json'] } },
  plugins: { ta },
  rules: { 'ta/drop-any-annotation': 'error', 'ta/no-undefined-union': 'error' },
}];`,
      'fix-me.ts':
        'declare function maybeString(): string | undefined;\n' +
        'let x: any = maybeString();\n' +
        'export { x };\n',
    });
    try {
      const res = await runRslint(['--fix', '--format', 'jsonline'], dir);
      expect(res.stderr).not.toMatch(
        /parserServices|getParserServices|TypeError/,
      );

      // Pass 1's drop-any-annotation fix removed `: any`, so the file now infers
      // x's type instead of pinning it to any.
      const fixed = await fs.readFile(path.join(dir, 'fix-me.ts'), 'utf8');
      expect(fixed).toContain('let x = maybeString()');
      expect(fixed).not.toContain(': any');

      // Pass 2 rebuilt the type snapshot from a FRESH program of the fixed file,
      // so the type-aware no-undefined-union now sees the inferred
      // `string | undefined` and flags x. A snapshot cached across passes would
      // still carry `any` (not a union) and silently miss this — the staleness
      // regression this test guards.
      const diags = parseDiags(res.stdout).filter(
        (d) =>
          d.ruleName === 'ta/no-undefined-union' &&
          d.filePath.endsWith('fix-me.ts'),
      );
      expect(diags).toHaveLength(1);
      expect(diags[0].range.start.line).toBe(2);
    } finally {
      await cleanupTempDir(dir);
    }
  });

  test('gate is project-based, not requiresTypeChecking: an UNDECLARED type-aware rule still gets type info and flags the union', async () => {
    const pluginPath = path.resolve(FIXTURE_DIR, 'type-aware-plugin.mjs');
    const dir = await createTempDir({
      'tsconfig.json': JSON.stringify({
        compilerOptions: { strict: true },
        include: ['*.ts'],
      }),
      'rslint.config.mjs': `import ta from ${JSON.stringify(pluginPath)};
export default [{
  files: ['*.ts'],
  languageOptions: { parserOptions: { projectService: false, project: ['./tsconfig.json'] } },
  plugins: { ta },
  rules: { 'ta/no-undefined-union-undeclared': 'error' },
}];`,
      'u.ts':
        'declare function f(): string | undefined;\nexport const u = f();\n',
    });
    try {
      const res = await runRslint(['--format', 'jsonline'], dir);
      expect(res.stderr).not.toMatch(
        /parserServices|getParserServices|TypeError/,
      );
      const diags = parseDiags(res.stdout).filter(
        (d) =>
          d.ruleName === 'ta/no-undefined-union-undeclared' &&
          d.filePath.endsWith('u.ts'),
      );
      // The rule declares NO meta.docs.requiresTypeChecking, yet it still
      // receives a snapshot (the file has a program) and flags the union —
      // proving type-aware gating is project-based, not declaration-based.
      expect(diags).toHaveLength(1);
    } finally {
      await cleanupTempDir(dir);
    }
  });
});

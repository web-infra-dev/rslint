/**
 * Differential harness: run a community ESLint-plugin rule through BOTH
 *  - ESLint v10 (in-process `Linter.verify`), and
 *  - rslint's `plugins` feature (via the CLI),
 * then compare the normalized diagnostics. A rule "matches" only when both
 * engines emit a byte-identical set of {ruleId, message, line, column,
 * endLine, endColumn, severity}.
 *
 * Each plugin is mounted under a NON-native alias (rslint reserves `unicorn`,
 * `promise`, ... for natively-ported plugins). The SAME alias is used on both
 * engines, so the emitted `ruleId` is identical across the comparison.
 */
import { createRequire } from 'node:module';
import { pathToFileURL } from 'node:url';
import { spawnSync } from 'node:child_process';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { Linter } from 'eslint';
import tsParser from '@typescript-eslint/parser';

const require = createRequire(import.meta.url);

/** One differential case: a minimal trigger for `<alias>/<rule>`. */
export interface DiffCase {
  /** Plugin package, e.g. `eslint-plugin-unicorn`. */
  pkg: string;
  /** Bare rule name, e.g. `no-null`. */
  rule: string;
  /** TypeScript source the rule should report on (or stay silent for CLEAN). */
  code: string;
  /** Extra rule options (ESLint's post-severity args). */
  options?: unknown[];
  /** Force a fixture filename (e.g. for filename-sensitive rules). */
  filename?: string;
  /** Force JSX parsing (otherwise auto-detected from `code`). */
  jsx?: boolean;
}

/** Normalized diagnostic shape compared across the two engines. */
export interface NormDiag {
  ruleId: string | null;
  message: string;
  line: number | null;
  column: number | null;
  endLine: number | null;
  endColumn: number | null;
  severity: 'error' | 'warning';
}

/** Per-case comparison outcome. */
export interface Verdict {
  index: number;
  pkg: string;
  rule: string;
  ruleId: string;
  /** Both engines emitted an identical diagnostic set. */
  match: boolean;
  /** Neither engine reported — the trigger exercises nothing. */
  empty: boolean;
  eslint: NormDiag[];
  rslint: NormDiag[];
}

/**
 * Per-plugin alias. Must not collide with rslint's native plugin prefixes
 * (`@typescript-eslint`, `import`, `jest`, `jsx-a11y`, `promise`, `react`,
 * `react-hooks`, `unicorn`). Both engines mount the plugin under this alias.
 */
export const ALIASES: Record<string, string> = {
  'eslint-plugin-unicorn': 'u',
  'eslint-plugin-sonarjs': 's',
  'eslint-plugin-promise': 'p',
  '@stylistic/eslint-plugin': 'st',
  'eslint-plugin-simple-import-sort': 'sis',
  '@eslint-community/eslint-plugin-eslint-comments': 'ec',
  'eslint-plugin-security': 'sec',
  'eslint-plugin-es-x': 'esx',
  'eslint-plugin-cypress': 'cy',
};

/** Resolve + dynamically import a plugin package's live default export. */
const pluginCache = new Map<string, unknown>();
async function loadPlugin(pkg: string): Promise<unknown> {
  const cached = pluginCache.get(pkg);
  if (cached) return cached;
  const mod = await import(pathToFileURL(require.resolve(pkg)).href);
  const plugin = mod.default ?? mod;
  pluginCache.set(pkg, plugin);
  return plugin;
}

function aliasFor(pkg: string): string {
  const a = ALIASES[pkg];
  if (!a) throw new Error(`no alias registered for plugin package ${pkg}`);
  return a;
}

/**
 * Heuristic: does this snippet contain JSX? JSX in a `.ts` file is a
 * TypeScript SYNTAX error (TS1005), and rslint's native pass aborts JSONL
 * output for the WHOLE batch on any syntax error — so JSX fixtures MUST be
 * `.tsx`. A case may set `jsx: true` explicitly to override the heuristic.
 */
function needsJsx(code: string): boolean {
  return /<\/?[A-Za-z][\w.:-]*(?:\s|\/?>|>)/.test(code) || /<>/.test(code);
}

/**
 * Per-case fixture identity: filename (right extension) + JSX parse mode. The
 * SAME filename is used as the rslint fixture name and the ESLint `filename`,
 * so filename-sensitive rules and the JSX/TSX parser mode line up on both
 * engines.
 */
export function fixtureMeta(
  c: DiffCase,
  index: number,
): { jsx: boolean; filename: string } {
  const jsx = c.jsx ?? needsJsx(c.code);
  const ext = jsx ? 'tsx' : 'ts';
  return { jsx, filename: c.filename ?? `c${index}.${ext}` };
}

interface EslintMessage {
  ruleId: string | null;
  message: string;
  line?: number;
  column?: number;
  endLine?: number;
  endColumn?: number;
  severity: number;
}

function normEslint(m: EslintMessage): NormDiag {
  return {
    ruleId: m.ruleId ?? null,
    message: m.message,
    line: m.line ?? null,
    column: m.column ?? null,
    endLine: m.endLine ?? null,
    endColumn: m.endColumn ?? null,
    severity: m.severity === 1 ? 'warning' : 'error',
  };
}

interface RslintDiag {
  ruleName: string;
  message: string;
  range: {
    start: { line: number; column: number };
    end: { line: number; column: number };
  };
  severity: 'error' | 'warning';
  filePath: string;
}

function normRslint(d: RslintDiag): NormDiag {
  return {
    ruleId: d.ruleName,
    message: d.message,
    line: d.range?.start?.line ?? null,
    column: d.range?.start?.column ?? null,
    endLine: d.range?.end?.line ?? null,
    endColumn: d.range?.end?.column ?? null,
    severity: d.severity,
  };
}

function sortDiags(diags: NormDiag[]): NormDiag[] {
  return [...diags].sort(
    (a, b) =>
      (a.line ?? 0) - (b.line ?? 0) ||
      (a.column ?? 0) - (b.column ?? 0) ||
      (a.endLine ?? 0) - (b.endLine ?? 0) ||
      (a.endColumn ?? 0) - (b.endColumn ?? 0) ||
      String(a.ruleId).localeCompare(String(b.ruleId)) ||
      String(a.message).localeCompare(String(b.message)),
  );
}

const ruleValue = (c: DiffCase): Linter.RuleEntry =>
  (c.options ? ['error', ...c.options] : 'error') as Linter.RuleEntry;

/**
 * Run every case through ESLint v10 in-process. Returns Map<index, NormDiag[]>.
 * Uses @typescript-eslint/parser WITHOUT a `project` (no type information) —
 * symmetric with rslint's eslint-plugin runner, which has no type info either.
 */
export async function runEslintV10(
  cases: DiffCase[],
): Promise<Map<number, NormDiag[]>> {
  const linter = new Linter({ configType: 'flat' });
  const out = new Map<number, NormDiag[]>();
  for (let i = 0; i < cases.length; i++) {
    const c = cases[i];
    const alias = aliasFor(c.pkg);
    const plugin = await loadPlugin(c.pkg);
    const ruleId = `${alias}/${c.rule}`;
    const { jsx, filename } = fixtureMeta(c, i);
    const config = [
      {
        files: ['**/*.ts', '**/*.tsx'],
        languageOptions: {
          parser: tsParser,
          ecmaVersion: 'latest',
          sourceType: 'module',
          parserOptions: { ecmaFeatures: { jsx } },
        },
        plugins: { [alias]: plugin },
        rules: { [ruleId]: ruleValue(c) },
      },
    ] as unknown as Linter.Config[];
    const messages = linter.verify(c.code, config, { filename });
    out.set(
      i,
      sortDiags(
        (messages as unknown as EslintMessage[])
          .filter((m) => m.ruleId === ruleId)
          .map(normEslint),
      ),
    );
  }
  return out;
}

/**
 * Run cases through rslint, bucketing JSONL diagnostics back by file. Splits
 * into chunks (one CLI invocation each) so a single syntax-broken fixture can
 * only zero out its own chunk, not the whole suite, and so no single config /
 * IPC payload grows unbounded. Returns Map<globalIndex, NormDiag[]>.
 */
export function runRslintBatch(cases: DiffCase[]): Map<number, NormDiag[]> {
  const CHUNK = 60;
  const byIndex = new Map<number, NormDiag[]>(cases.map((_, i) => [i, []]));
  for (let start = 0; start < cases.length; start += CHUNK) {
    runRslintChunk(
      cases,
      start,
      Math.min(start + CHUNK, cases.length),
      byIndex,
    );
  }
  for (const [i, arr] of byIndex) byIndex.set(i, sortDiags(arr));
  return byIndex;
}

/** Run one chunk [start,end) in its own temp project, filling `byIndex` by the
 *  GLOBAL case index (fixture files are named `c<globalIndex>.<ext>`). */
function runRslintChunk(
  cases: DiffCase[],
  start: number,
  end: number,
  byIndex: Map<number, NormDiag[]>,
): void {
  const rslintBin = require.resolve('@rslint/core/bin');
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), 'rslint-v10-diff-'));
  try {
    // One import per distinct plugin package in this chunk, reused across its
    // entries so the alias maps to a single live instance.
    const pkgs = [...new Set(cases.slice(start, end).map((c) => c.pkg))];
    const pkgVar = new Map<string, string>();
    const importLines = pkgs.map((pkg, idx) => {
      pkgVar.set(pkg, `_p${idx}`);
      return `import _p${idx} from ${JSON.stringify(pathToFileURL(require.resolve(pkg)).href)};`;
    });

    const fileToIndex = new Map<string, number>();
    const entries: string[] = [];
    for (let i = start; i < end; i++) {
      const c = cases[i];
      const { filename } = fixtureMeta(c, i);
      fileToIndex.set(filename, i);
      fs.writeFileSync(path.join(dir, filename), c.code);
      const alias = aliasFor(c.pkg);
      const ruleId = `${alias}/${c.rule}`;
      const ruleVal = c.options
        ? JSON.stringify(['error', ...c.options])
        : "'error'";
      entries.push(
        `  { files: [${JSON.stringify(filename)}], ` +
          `plugins: { ${alias}: ${pkgVar.get(c.pkg)} }, ` +
          `rules: { ${JSON.stringify(ruleId)}: ${ruleVal} } }`,
      );
    }

    fs.writeFileSync(
      path.join(dir, 'rslint.config.mjs'),
      `${importLines.join('\n')}\nexport default [\n${entries.join(',\n')}\n];\n`,
    );
    fs.writeFileSync(
      path.join(dir, 'tsconfig.json'),
      JSON.stringify({
        compilerOptions: {
          target: 'esnext',
          module: 'esnext',
          jsx: 'preserve',
          skipLibCheck: true,
        },
        include: ['*.ts', '*.tsx'],
      }),
    );

    const res = spawnSync(
      process.execPath,
      [rslintBin, '--format', 'jsonline'],
      { cwd: dir, encoding: 'utf8', env: { ...process.env, NO_COLOR: '1' } },
    );

    if (
      /invalid config|Cannot find|failed to load/i.test(res.stderr) &&
      !res.stdout.trim()
    ) {
      throw new Error(`rslint failed:\n${res.stderr.slice(0, 2000)}`);
    }

    for (const line of res.stdout.split('\n')) {
      const t = line.trim();
      if (!t.startsWith('{')) continue;
      const d = JSON.parse(t) as RslintDiag;
      const idx = fileToIndex.get(d.filePath);
      if (idx === undefined) continue;
      byIndex.get(idx)!.push(normRslint(d));
    }
  } finally {
    fs.rmSync(dir, { recursive: true, force: true });
  }
}

/**
 * Compare both engines over `cases`. Returns one {@link Verdict} per case;
 * `match` = identical normalized diag sets, `empty` = neither engine reported.
 */
export async function compareCases(cases: DiffCase[]): Promise<Verdict[]> {
  const es = await runEslintV10(cases);
  const rs = runRslintBatch(cases);
  return cases.map((c, i) => {
    const eslint = es.get(i) ?? [];
    const rslint = rs.get(i) ?? [];
    return {
      index: i,
      pkg: c.pkg,
      rule: c.rule,
      ruleId: `${aliasFor(c.pkg)}/${c.rule}`,
      match: JSON.stringify(eslint) === JSON.stringify(rslint),
      empty: eslint.length === 0 && rslint.length === 0,
      eslint,
      rslint,
    };
  });
}

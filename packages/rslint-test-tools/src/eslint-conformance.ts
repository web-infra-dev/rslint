/**
 * ESLint conformance harness.
 *
 * For each fixture in `fixtureGlob`, run the chosen ESLint plugin twice:
 *
 *   1. through real ESLint v10 (the canonical execution)
 *   2. through rslint's runner WorkerPool (the same `@rslint/eslint-plugin-runner`
 *      pool that production uses end-to-end)
 *
 * Then structurally diff the diagnostics. Any mismatch outside the
 * caller-supplied allow-list is a conformance failure.
 *
 * This harness is the load-bearing safety net for the claim that rslint
 * produces diagnostics identical to ESLint — it's a CI verification, not
 * a manual assertion. Every plugin added to the compatibility matrix
 * runs through it.
 *
 * Scope:
 *   - diagnostic count
 *   - rule name
 *   - line / column (1-based, matching ESLint's `Linter.LintMessage`)
 *   - messageId
 *
 * Out of scope (intentionally — we don't re-implement ESLint):
 *   - exact byte offsets, since ESLint reports them via tokens which can
 *     drift without affecting user-facing line/column
 *   - applied-fix bytes — fix application uses different code paths in
 *     ESLint vs rslint; meaningful comparison needs equivalent
 *     ApplyRuleFixes-shaped pre-application data which the harness does
 *     not yet exercise.
 */

import { Linter } from 'eslint';
import { pathToFileURL } from 'node:url';
import { mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { createRequire } from 'node:module';
import os from 'node:os';
import path from 'node:path';

import { WorkerPool, type LintTask } from '@rslint/eslint-plugin-runner';

// ─── Types ───────────────────────────────────────────────────────────

/**
 * One configured plugin under test. Mirrors what users would put in their
 * rslint.config.ts's `eslintPlugins` field, plus the npm specifier so
 * the harness can resolve it.
 */
export interface ConformancePluginConfig {
  /** Namespace prefix used by rules: `<prefix>/<ruleName>`. */
  prefix: string;
  /** Unmodified ESLint plugin instance (loaded via `import()`). */
  plugin: {
    meta?: { name?: string; version?: string };
    name?: string;
    rules?: Record<string, unknown>;
  };
  /** npm specifier (used by the runner Worker for its own import). */
  specifier: string;
  /** Rule names under test. Keys of plugin.rules. */
  ruleNames: string[];
}

/**
 * One fixture file. The harness assumes the file has been read into `text`
 * already; callers can map a glob → fixtures with whatever IO strategy
 * (synchronous readFileSync is fine).
 */
export interface ConformanceFixture {
  filePath: string;
  text: string;
  /** rslint config rules block — same shape as user config. */
  rules: Record<string, 'error' | 'warn' | [string, ...unknown[]]>;
  /** TypeScript? */
  isTypeScript?: boolean;
}

/**
 * Allow-list entry for tolerated wording differences. Either an exact
 * `messageId` to ignore the message text on, or a regex to match strings
 * that should be considered equivalent.
 */
export interface AllowListEntry {
  ruleName: string;
  /** Skip message text comparison for this messageId. */
  messageId?: string;
  /** Skip the diagnostic entirely (ESLint emits but we don't, or vice versa). */
  ignore?: boolean;
}

export interface ConformanceOptions {
  plugin: ConformancePluginConfig;
  fixtures: ConformanceFixture[];
  /** URL anchor for plugin specifier resolution. */
  resolverBaseUrl: string;
  allowList?: AllowListEntry[];
  /** Pool worker count; default 2 for harness speed. */
  workerCount?: number;
}

/** A single normalized diagnostic the harness compares on. */
export interface NormalizedDiagnostic {
  ruleName: string;
  messageId: string | undefined;
  line: number;
  column: number;
  message: string;
}

/** Per-fixture comparison result. */
export interface FixtureComparison {
  filePath: string;
  match: boolean;
  reasons: string[];
  eslint: NormalizedDiagnostic[];
  rslint: NormalizedDiagnostic[];
}

export interface ConformanceReport {
  matched: number;
  mismatched: number;
  fixtureResults: FixtureComparison[];
}

// ─── Implementation ──────────────────────────────────────────────────

/**
 * Run the conformance harness. Returns a report; throwing is reserved
 * for harness-internal failures (pool init, missing plugin) — content
 * mismatches are reported, not thrown.
 */
export async function runConformance(
  opts: ConformanceOptions,
): Promise<ConformanceReport> {
  // ── ESLint side ──
  const eslintLinter = new Linter({ configType: 'flat' });

  // ── rslint side ──
  //
  // The runner Worker imports rslint config files directly to recover
  // live plugin instances (worker_threads can't postMessage functions).
  // The harness owns this side of the contract: we mktemp a real
  // `rslint.config.mjs` and have it import the plugin via the plugin
  // module's ABSOLUTE filesystem path (file:// URL). That breaks the
  // dependency on the temp directory having a `node_modules` next to
  // it — we can put the config in `/tmp/xxx/` cleanly without
  // polluting the package source tree, and Node's ESM resolver loads
  // the absolute URL without any `node_modules` walking.
  //
  // `resolverBaseUrl` is conventionally the consumer package's
  // `package.json` URL — we use it only to resolve the plugin
  // specifier to a concrete file path (via createRequire), nothing
  // about the test config lives at that location anymore.
  const requireFromBase = createRequire(opts.resolverBaseUrl);
  const pluginAbsPath = requireFromBase.resolve(opts.plugin.specifier);
  const pluginUrl = pathToFileURL(pluginAbsPath).href;

  const tmpDir = mkdtempSync(path.join(os.tmpdir(), 'rslint-conformance-'));
  const cfgPath = path.join(tmpDir, 'rslint.config.mjs');
  writeFileSync(
    cfgPath,
    [
      `import plugin from ${JSON.stringify(pluginUrl)};`,
      `const unwrapped = plugin && typeof plugin === 'object' && 'default' in plugin`,
      `  ? plugin.default ?? plugin`,
      `  : plugin;`,
      `export default [{ eslintPlugins: { ${JSON.stringify(opts.plugin.prefix)}: unwrapped } }];`,
      '',
    ].join('\n'),
  );

  const pool = new WorkerPool({
    configs: [{ configPath: cfgPath, configDirectory: tmpDir }],
    workerCount: opts.workerCount ?? 2,
  });

  await pool.init();

  try {
    const results: FixtureComparison[] = [];
    let matched = 0;
    let mismatched = 0;

    for (const fix of opts.fixtures) {
      const eslintDiagnostics = runEslint(eslintLinter, fix, opts.plugin);
      const rslintDiagnostics = await runRslint(pool, fix, opts.plugin, tmpDir);

      const cmp = compareDiagnostics(
        fix.filePath,
        eslintDiagnostics,
        rslintDiagnostics,
        opts.allowList ?? [],
      );
      results.push(cmp);
      if (cmp.match) matched++;
      else mismatched++;
    }

    return { matched, mismatched, fixtureResults: results };
  } finally {
    await pool.shutdown();
    try {
      rmSync(cfgPath, { force: true });
    } catch {}
    try {
      rmSync(tmpDir, { recursive: true, force: true });
    } catch {}
  }
}

/**
 * Lazy singleton: load @typescript-eslint/parser exactly once across
 * the suite. It's a 10+ MB transitive dep; loading per-fixture would
 * dominate test wall time and crash CI runners with low file-descriptor
 * limits. Cached on first TS fixture, undefined for JS-only suites.
 */
let _tsParserCache: unknown | undefined;
function getTsParser(): unknown {
  if (_tsParserCache !== undefined) return _tsParserCache;
  // Use require because @typescript-eslint/parser exports a CJS-shaped
  // module (`module.exports.parseForESLint`); ESM `import()` returns
  // a `{default: <obj>}` wrapper here which ESLint v10 won't accept as
  // a flat-config parser. The runtime require lookup matches what
  // ESLint itself does internally to resolve a config-injected parser.
  // eslint-disable-next-line @typescript-eslint/no-require-imports
  _tsParserCache = require('@typescript-eslint/parser');
  return _tsParserCache;
}

function runEslint(
  linter: Linter,
  fixture: ConformanceFixture,
  plugin: ConformancePluginConfig,
): NormalizedDiagnostic[] {
  // Wire the plugin into a flat-config block; let ESLint resolve everything.
  // In v10 flat-config you pass plugins as `{ <prefix>: pluginInstance }` and
  // rules as `{ <prefix>/<rule>: 'error' }`.
  //
  // `languageOptions.parserOptions` is set to an empty object so legacy
  // ESLint plugins (e.g. eslint-plugin-import) that probe
  // `'sourceType' in context.parserOptions` don't TypeError on
  // undefined. ecmaVersion / sourceType themselves are top-level
  // flat-config fields (v10 sets these explicitly for parity).
  //
  // `linterOptions.reportUnusedDisableDirectives: 'off'` matches
  // rslint's compat path which doesn't run ESLint's
  // disable-directive application step. Without this, ESLint
  // emits an extra `Unused eslint-disable directive` warning per
  // unmatched directive — pure linter-host noise that doesn't come
  // from the plugin rule itself, so we suppress it here for
  // byte-equal comparison.
  type FlatConfig = Record<string, unknown>;
  const block: FlatConfig = {
    files: [fixture.filePath],
    plugins: { [plugin.prefix]: plugin.plugin },
    rules: fixture.rules,
    linterOptions: {
      reportUnusedDisableDirectives: 'off',
    },
    languageOptions: {
      ecmaVersion: 'latest',
      sourceType: 'module',
      parserOptions: {},
    },
  };
  if (fixture.isTypeScript) {
    // ESLint v10's default parser (espree) can't read TS syntax. For TS
    // fixtures we plug in @typescript-eslint/parser so the ESLint side
    // produces an AST compatible with what rslint (via oxc-parser)
    // also emits — comparable diagnostics depend on comparable ASTs.
    //
    // Override `languageOptions` (set above for JS defaults) so the
    // parser slot is in place; deliberately leaving `parserOptions`
    // empty because most ts-eslint rules in scope are syntax-only
    // and don't need `project` / type info.
    block.languageOptions = {
      ecmaVersion: 'latest',
      sourceType: 'module',
      parserOptions: {},
      parser: getTsParser(),
    };
  }
  const cfg: FlatConfig[] = [block];
  const messages = linter.verify(fixture.text, cfg as never, fixture.filePath);
  return messages.map((m: Linter.LintMessage) => ({
    ruleName: m.ruleId ?? '(unknown)',
    messageId: m.messageId ?? undefined,
    line: m.line,
    column: m.column,
    message: m.message,
  }));
}

async function runRslint(
  pool: WorkerPool,
  fixture: ConformanceFixture,
  plugin: ConformancePluginConfig,
  configKey: string,
): Promise<NormalizedDiagnostic[]> {
  const tasks: LintTask[] = [
    {
      filePath: fixture.filePath,
      text: fixture.text,
      rules: Object.fromEntries(
        Object.entries(fixture.rules).map(([k, v]) => [
          k,
          {
            options: Array.isArray(v) ? v.slice(1) : [],
            meta: undefined,
          },
        ]),
      ),
      collectFixes: false,
      suggestionsMode: 'off',
      configKey,
    },
  ];

  const results = await pool.lintBatch(tasks);
  const out: NormalizedDiagnostic[] = [];
  for (const d of results[0].diagnostics) {
    // rslint reports byte offsets; convert to 1-based line/column to match ESLint.
    const lso = buildLineStartOffsets(fixture.text);
    const startLine = lineFromOffset(d.startPos, lso);
    const startCol = d.startPos - lso[startLine - 1] + 1; // ESLint: 1-based
    out.push({
      ruleName: d.ruleName,
      messageId: d.messageId,
      line: startLine,
      column: startCol,
      message: d.message,
    });
  }
  void plugin; // currently unused; reserved for future per-plugin shaping
  return out;
}

function buildLineStartOffsets(text: string): number[] {
  const offsets = [0];
  for (let i = 0; i < text.length; i++) {
    if (text.charCodeAt(i) === 10 /* \n */) offsets.push(i + 1);
  }
  return offsets;
}

function lineFromOffset(offset: number, lso: number[]): number {
  let lo = 0,
    hi = lso.length - 1;
  while (lo < hi) {
    const mid = (lo + hi + 1) >> 1;
    if (lso[mid] <= offset) lo = mid;
    else hi = mid - 1;
  }
  return lo + 1;
}

function compareDiagnostics(
  filePath: string,
  eslint: NormalizedDiagnostic[],
  rslint: NormalizedDiagnostic[],
  allowList: AllowListEntry[],
): FixtureComparison {
  const reasons: string[] = [];

  // Filter ignored entries per allow-list before counting
  const allowIgnoreSet = new Set(
    allowList
      .filter((e) => e.ignore)
      .map((e) => `${e.ruleName}::${e.messageId ?? ''}`),
  );
  const filterIgnored = (ds: NormalizedDiagnostic[]) =>
    ds.filter(
      (d) => !allowIgnoreSet.has(`${d.ruleName}::${d.messageId ?? ''}`),
    );
  const e = filterIgnored(eslint);
  const r = filterIgnored(rslint);

  if (e.length !== r.length) {
    reasons.push(`count mismatch: eslint=${e.length}, rslint=${r.length}`);
  }

  // Pair by (ruleName, line, column); sort both sides identically first.
  const sortKey = (d: NormalizedDiagnostic) =>
    `${d.ruleName}::${d.line}:${d.column}::${d.messageId ?? ''}`;
  const eSorted = [...e].sort((a, b) => sortKey(a).localeCompare(sortKey(b)));
  const rSorted = [...r].sort((a, b) => sortKey(a).localeCompare(sortKey(b)));

  const max = Math.max(eSorted.length, rSorted.length);
  for (let i = 0; i < max; i++) {
    const ed = eSorted[i];
    const rd = rSorted[i];
    if (!ed || !rd) {
      reasons.push(
        `diagnostic at index ${i}: ${ed ? 'ESLint only' : 'rslint only'}`,
      );
      continue;
    }
    if (ed.ruleName !== rd.ruleName) {
      reasons.push(
        `#${i} ruleName mismatch: eslint=${ed.ruleName}, rslint=${rd.ruleName}`,
      );
    }
    if (ed.line !== rd.line || ed.column !== rd.column) {
      reasons.push(
        `#${i} location mismatch on ${ed.ruleName}: eslint=${ed.line}:${ed.column}, rslint=${rd.line}:${rd.column}`,
      );
    }
    if (ed.messageId !== rd.messageId) {
      reasons.push(
        `#${i} messageId mismatch: eslint=${ed.messageId}, rslint=${rd.messageId}`,
      );
    }
    // message text is *not* compared by default — message wording can drift
    // across plugin versions. The allow-list mechanism could be extended
    // to enforce exact-match per ruleName for stricter regression checks.
  }

  return {
    filePath,
    match: reasons.length === 0,
    reasons,
    eslint,
    rslint,
  };
}

/** Convenience: format a ConformanceReport for console.log. */
export function formatReport(report: ConformanceReport): string {
  const lines: string[] = [];
  lines.push(
    `[conformance] matched=${report.matched} mismatched=${report.mismatched}`,
  );
  for (const r of report.fixtureResults) {
    if (r.match) {
      lines.push(`  ✓ ${r.filePath}`);
    } else {
      lines.push(`  ✗ ${r.filePath}`);
      for (const reason of r.reasons) lines.push(`      - ${reason}`);
    }
  }
  return lines.join('\n');
}

/** Test-friendly default base URL anchor. */
export function defaultResolverBaseUrl(): string {
  return pathToFileURL(path.resolve(process.cwd(), 'package.json')).href;
}

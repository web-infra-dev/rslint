import { spawnSync } from 'node:child_process';
import { createRequire } from 'node:module';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { pathToFileURL } from 'node:url';
import util from 'node:util';

// Shared alignment RuleTester for community ESLint plugins that have NO native
// Go port (@stylistic, simple-import-sort, eslint-comments, …). Such plugins
// cannot run through the `lint()` JS API: that path spawns `rslint --api`, a
// pure-Go server that only executes native rules and silently ignores
// eslint-plugin rules (see cmd/rslint/api.go). They run ONLY through the CLI,
// which loads the real plugin in a Node worker pool.
//
// So this tester drives the rslint CLI directly (the same engine path the
// conformance harness uses): each `run()` writes one fixture file per case plus
// a generated `rslint.config.mjs` that mounts the live plugin, invokes the CLI
// once, and buckets the JSONL diagnostics back per case. Upstream valid/invalid
// arrays are ported verbatim; expected messages are resolved from the plugin's
// own `meta.messages` (rslint's JSONL carries the rendered `message` text, not
// the `messageId`).
//
// `createRuleTester({ pkg, prefix, plugin })` binds one plugin; each
// `tests/eslint-plugin-<plugin>/rule-tester.ts` is a thin wrapper over it.

const require = createRequire(import.meta.url);

interface TestCaseError {
  message?: string | RegExp;
  messageId?: string;
  type?: string | undefined;
  data?: Record<string, unknown>;
  line?: number | undefined;
  column?: number | undefined;
  endLine?: number | undefined;
  endColumn?: number | undefined;
  suggestions?: unknown[] | undefined;
}

export interface ValidTestCase {
  name?: string;
  code: string;
  options?: any;
  filename?: string | undefined;
  only?: boolean;
  settings?: Record<string, any> | undefined;
  parserOptions?: Record<string, any> | undefined;
  output?: string | null | undefined;
}

export interface InvalidTestCase extends ValidTestCase {
  // Optional: upstream's eslint-vitest-rule-tester tolerates an invalid case
  // that pins ONLY `output` (it verifies the fix, not a diagnostic set). Such a
  // case is ported `errors`-less — we then assert the fix output plus a sanity
  // check that it genuinely reports (≥1 diagnostic), without inventing the
  // diagnostic positions from rslint's own output.
  errors?: number | (TestCaseError | string)[];
  output?: string | null | undefined;
}

/** The live plugin object, narrowed to what the tester reads (`meta.messages`
 *  per rule, for resolving `messageId` → rendered message text). */
export interface RuleTesterPlugin {
  rules: Record<string, { meta?: { messages?: Record<string, string> } }>;
}

export interface RuleTesterOptions {
  /** Plugin package name, e.g. `@stylistic/eslint-plugin`. Resolved to a file
   *  URL for the generated config's `import`. */
  pkg: string;
  /** Mounted plugin prefix, e.g. `@stylistic`. `ruleId = prefix/ruleName`. */
  prefix: string;
  /** The live plugin (its default export) — read for `meta.messages`. */
  plugin: RuleTesterPlugin;
}

/** One diagnostic as emitted by `rslint --format jsonline`. */
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

/** Per-case result: the rule's diagnostics, plus the `--fix` output when the
 *  rule was run with autofix (else `null`). */
interface CaseResult {
  diags: RslintDiag[];
  fixed: string | null;
}

/**
 * JSX in a `.ts` file is a TypeScript SYNTAX error (TS1005) and rslint aborts
 * JSONL for the whole batch on a syntax error, so JSX fixtures MUST be `.tsx`.
 *
 * Detect REAL JSX only — a closing tag (`</Tag`), a self-close (`/>`), or a
 * fragment (`<>`). The earlier "any `<Tag>`" heuristic mis-classified a TS
 * generic arrow `<T>(a) => b` (and type args like `foo<T>()`) as JSX and routed
 * it to `.tsx`, where ts-go reads the lone `<T>` as an unclosed JSX element
 * (TS17008). Genuine JSX always carries one of these three markers, so this
 * stays correct for real JSX while keeping generics on `.ts`.
 */
function needsJsx(code: string): boolean {
  return /<\/[A-Za-z]/.test(code) || /\/>/.test(code) || /<>/.test(code);
}

/** Render an ESLint message template (`{{x}}` → data.x), as the rule itself
 *  would. Returns the template unchanged when `data` lacks a referenced key —
 *  the caller treats a still-templated result as "message not assertable". */
function interpolate(
  template: string,
  data: Record<string, unknown> | undefined,
): string {
  if (!data) return template;
  return template.replace(/\{\{\s*([^{}]+?)\s*\}\}/g, (whole, key: string) =>
    Object.prototype.hasOwnProperty.call(data, key) ? String(data[key]) : whole,
  );
}

interface NormalizedCase {
  code: string;
  options: unknown[];
  filename: string;
  settings: Record<string, unknown> | undefined;
}

function normalize(
  raw: ValidTestCase | InvalidTestCase | string,
  index: number,
): NormalizedCase {
  const c = typeof raw === 'string' ? { code: raw } : raw;
  const jsx = needsJsx(c.code);
  const filename = c.filename ?? `case${index}.${jsx ? 'tsx' : 'ts'}`;
  return {
    code: c.code,
    options: c.options ?? [],
    filename,
    settings: c.settings,
  };
}

/** Build a RuleTester bound to one plugin (`pkg`/`prefix`/`plugin`). */
export function createRuleTester(opts: RuleTesterOptions) {
  const { pkg: PKG, prefix: PREFIX, plugin } = opts;

  /** The fully-rendered message an error expects, or `null` when the case pins
   *  no message (bare count) / the template can't be fully resolved from data. */
  function expectedMessage(
    ruleName: string,
    error: TestCaseError | string,
  ): string | RegExp | null {
    if (typeof error === 'string') return error;
    if (error.message != null) return error.message;
    if (error.messageId != null) {
      const template =
        plugin.rules[ruleName]?.meta?.messages?.[error.messageId];
      assert.ok(
        template != null,
        `Rule ${PREFIX}/${ruleName} has no messageId '${error.messageId}'`,
      );
      const rendered = interpolate(template, error.data);
      return rendered.includes('{{') ? null : rendered;
    }
    return null;
  }

  /**
   * Run every normalized case through the rslint CLI in one invocation and
   * bucket the rule's diagnostics back by case index. Each case is its own
   * fixture file with its own config entry, so per-case options/settings stay
   * isolated.
   */
  function runRslint(
    ruleId: string,
    cases: NormalizedCase[],
    needFix: boolean,
  ): CaseResult[] {
    const rslintBin = require.resolve('@rslint/core/bin');
    const pluginUrl = pathToFileURL(require.resolve(PKG)).href;
    const dir = fs.mkdtempSync(path.join(os.tmpdir(), 'rslint-plugin-test-'));
    try {
      const fileToIndex = new Map<string, number>();
      const fixtureNames: string[] = [];
      const entries: string[] = [];
      cases.forEach((c, i) => {
        // Distinct fixture name per case even if two share a `filename`.
        const ext = path.extname(c.filename) || '.ts';
        const fixture = `c${i}${ext}`;
        fixtureNames.push(fixture);
        fileToIndex.set(fixture, i);
        fs.writeFileSync(path.join(dir, fixture), c.code);
        const ruleVal =
          c.options.length > 0
            ? JSON.stringify(['error', ...c.options])
            : "'error'";
        const settings = c.settings
          ? `, settings: ${JSON.stringify(c.settings)}`
          : '';
        entries.push(
          `  { files: [${JSON.stringify(fixture)}], ` +
            `plugins: { ${JSON.stringify(PREFIX)}: _p }, ` +
            `rules: { ${JSON.stringify(ruleId)}: ${ruleVal} }${settings} }`,
        );
      });

      fs.writeFileSync(
        path.join(dir, 'rslint.config.mjs'),
        `import _p from ${JSON.stringify(pluginUrl)};\n` +
          `export default [\n${entries.join(',\n')}\n];\n`,
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
        throw new Error(
          `rslint CLI failed for ${ruleId}:\n${res.stderr.slice(0, 2000)}`,
        );
      }

      // A ts-go SYNTAX error (TS1xxx) on ANY fixture in this CLI run aborts the
      // whole batch: rslint prints human-format TS errors to stderr and emits
      // NO JSONL — so every case would read back 0 diagnostics, silently
      // passing the `valid` cases (a false green). Fail loudly instead, naming
      // the offending fixtures so they get isolated into KNOWN GAPS — they are
      // unparseable under ts-go's strict/module semantics, a real documented
      // gap, never hidden.
      const hasJsonl = res.stdout
        .split('\n')
        .some((l) => l.trim().startsWith('{'));
      if (!hasJsonl && /TypeScript\(TS\d+\)|error TS\d+/.test(res.stderr)) {
        const bad = new Set<string>();
        for (const m of res.stderr.matchAll(
          /\(\s*([\w./-]+\.tsx?):\d+:\d+\s*\)/g,
        )) {
          bad.add(path.basename(m[1]));
        }
        const badList = [...bad].map((f) => {
          const i = fileToIndex.get(f);
          return i != null
            ? `  case[${i}]: ${util.inspect(cases[i].code)}`
            : `  ${f}`;
        });
        throw new Error(
          `[${ruleId}] rslint aborted the batch on a ts-go syntax error; these ` +
            `fixtures are unparseable under strict/module semantics and MUST be ` +
            `isolated into KNOWN GAPS:\n${badList.join('\n') || '  (file not identified)'}` +
            `\n\nstderr:\n${res.stderr.slice(0, 1200)}`,
        );
      }

      const byIndex: RslintDiag[][] = cases.map(() => []);
      for (const line of res.stdout.split('\n')) {
        const t = line.trim();
        if (!t.startsWith('{')) continue;
        const d = JSON.parse(t) as RslintDiag;
        const idx = fileToIndex.get(d.filePath);
        if (idx === undefined) continue;
        if (d.ruleName !== ruleId) continue;
        byIndex[idx].push(d);
      }

      // Second pass for autofix: `--fix` rewrites each fixture in place; its new
      // contents are the fixed output. Only spawned when a case pins `output`.
      // NOTE: rslint fixes to a stable point (multi-pass), whereas ESLint's
      // RuleTester `output` is a single fix pass — for the rare rule where they
      // differ that surfaces as a real, documented gap (never silently hidden).
      const fixed: (string | null)[] = cases.map(() => null);
      if (needFix) {
        spawnSync(process.execPath, [rslintBin, '--fix'], {
          cwd: dir,
          encoding: 'utf8',
          env: { ...process.env, NO_COLOR: '1' },
        });
        cases.forEach((_, i) => {
          fixed[i] = fs.readFileSync(path.join(dir, fixtureNames[i]), 'utf8');
        });
      }

      return cases.map((_, i) => ({ diags: byIndex[i], fixed: fixed[i] }));
    } finally {
      fs.rmSync(dir, { recursive: true, force: true });
    }
  }

  return class RuleTester {
    run(
      ruleName: string,
      _rule: never,
      cases: {
        valid: (ValidTestCase | string)[];
        invalid: InvalidTestCase[];
      },
    ) {
      const ruleId = `${PREFIX}/${ruleName}`;
      describe(ruleId, () => {
        const valid = cases.valid;
        const invalid = cases.invalid;
        const all = [...valid, ...invalid];
        const normalized = all.map(normalize);

        // Any invalid case that pins `output` triggers the autofix pass.
        const needFix = invalid.some((ic) =>
          Object.prototype.hasOwnProperty.call(ic, 'output'),
        );
        // One CLI invocation for the whole rule, memoized across valid+invalid.
        let resultsCache: CaseResult[] | undefined;
        const results = () =>
          (resultsCache ??= runRslint(ruleId, normalized, needFix));

        test('valid', () => {
          const diagsByCase = results();
          valid.forEach((vc, i) => {
            const code = typeof vc === 'string' ? vc : vc.code;
            const diags = diagsByCase[i].diags;
            assert.strictEqual(
              diags.length,
              0,
              `Expected no ${ruleId} diagnostics for valid case but got ${diags.length}:\n` +
                `Code: ${util.inspect(code)}\n${util.inspect(diags)}`,
            );
          });
        });

        test('invalid', () => {
          const diagsByCase = results();
          invalid.forEach((ic, j) => {
            const result = diagsByCase[valid.length + j];
            const diags = result.diags;
            // Diagnostic-count assertion. An `errors`-less case (upstream pinned
            // ONLY `output`, not a diagnostic set) skips the exact count and
            // just sanity-checks that it genuinely reports (≥1); the `output`
            // check below is the real assertion for those — never invent positions.
            if (ic.errors == null) {
              assert.ok(
                Object.prototype.hasOwnProperty.call(ic, 'output'),
                `Invalid case for ${ruleId} pins neither 'errors' nor 'output': ${util.inspect(ic.code)}`,
              );
              assert.ok(
                diags.length >= 1,
                `Output-only invalid case for ${ruleId} produced no diagnostics: ${util.inspect(ic.code)}\n${util.inspect(diags)}`,
              );
            } else {
              const expectedCount =
                typeof ic.errors === 'number' ? ic.errors : ic.errors.length;
              assert.strictEqual(
                diags.length,
                expectedCount,
                util.format(
                  'Should have %d %s error%s for %s but had %d: %s',
                  expectedCount,
                  ruleId,
                  expectedCount === 1 ? '' : 's',
                  util.inspect(ic.code),
                  diags.length,
                  util.inspect(diags),
                ),
              );
            }

            // Autofix output (ESLint RuleTester semantics): `output` omitted ⇒
            // not checked; `output: null` ⇒ asserts the source is unchanged; a
            // string ⇒ asserts the fixed source equals it.
            if (Object.prototype.hasOwnProperty.call(ic, 'output')) {
              const expectedFixed = ic.output == null ? ic.code : ic.output;
              assert.strictEqual(
                result.fixed,
                expectedFixed,
                `Autofix output mismatch for ${ruleId} (${util.inspect(ic.code)})`,
              );
            }

            if (ic.errors == null || typeof ic.errors === 'number') return;

            ic.errors.forEach((error, i) => {
              const diag = diags[i];
              const expMsg = expectedMessage(ruleName, error);
              if (expMsg instanceof RegExp) {
                assert.ok(
                  expMsg.test(diag.message),
                  `Error ${i}: expected ${diag.message} to match ${expMsg}`,
                );
              } else if (typeof expMsg === 'string') {
                assert.strictEqual(
                  diag.message,
                  expMsg,
                  `Error ${i} message mismatch for ${ruleId}`,
                );
              }
              if (typeof error === 'object' && error !== null) {
                if (error.line != null) {
                  assert.strictEqual(
                    diag.range.start.line,
                    error.line,
                    `Error ${i} line mismatch for ${ruleId} (${util.inspect(ic.code)})`,
                  );
                }
                if (error.column != null) {
                  assert.strictEqual(
                    diag.range.start.column,
                    error.column,
                    `Error ${i} column mismatch for ${ruleId} (${util.inspect(ic.code)})`,
                  );
                }
                if (error.endLine != null) {
                  assert.strictEqual(
                    diag.range.end.line,
                    error.endLine,
                    `Error ${i} endLine mismatch for ${ruleId} (${util.inspect(ic.code)})`,
                  );
                }
                if (error.endColumn != null) {
                  assert.strictEqual(
                    diag.range.end.column,
                    error.endColumn,
                    `Error ${i} endColumn mismatch for ${ruleId} (${util.inspect(ic.code)})`,
                  );
                }
              }
            });
          });
        });
      });
    }
  };
}

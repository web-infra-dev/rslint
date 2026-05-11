/**
 * ESLint API audit — exhaustive compatibility audit for the `SourceCode`
 * and `RuleContext` surfaces that real plugin rules consume.
 *
 * For each API the harness runs the SAME probe rule through:
 *   1. real ESLint v10 (`new Linter({ configType: 'flat' }).verify(...)`)
 *   2. rslint's runner (`lintFile` from `@rslint/eslint-plugin-runner`)
 *
 * The probe is a stub rule whose listener calls the API under test,
 * serialises the return into a deterministic string, and reports it via
 * `ctx.report({ node, message: <serialised> })`. The harness then
 * compares the produced diagnostic message strings byte-for-byte.
 *
 * Equality → rslint mirrors ESLint for that API.
 * Difference → diff is captured in the failure report so a follow-up
 * can either fix the runner or update the `@experimental` matrix.
 *
 * NOT all 50+ public APIs are exhaustively pinned here — initial
 * coverage targets the surface most-used by community plugins (the
 * `SourceCode` token / scope / location helpers and the `RuleContext`
 * fields that the v8→v9 migration kept). Adding a new API to the
 * matrix is one entry in `apiChecks` below.
 */
import { describe, test, expect } from '@rstest/core';
import { Linter, type Rule } from 'eslint';

import { lintFile } from '@rslint/eslint-plugin-runner';
import type { LoadedPlugins } from '@rslint/eslint-plugin-runner';

// ─── Probe harness ────────────────────────────────────────────────────

interface ApiCheck {
  /** Display name used in failure assertions. */
  name: string;
  /** Source text to lint. Default: `const x = 1;`. */
  text?: string;
  /** AST node selector that triggers the probe. Default: `Program:exit`. */
  trigger?: string;
  /**
   * Probe body. Receives the rule context (typed loosely because both
   * ESLint and rslint contexts exist behind the wire) and returns the
   * stringified observation. Throwing falls through to a `THREW:` prefix
   * so the diagnostic still reports something comparable.
   */
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  probe: (ctx: any) => string;
  /**
   * Optional category for the audit matrix output. Plain free-form
   * label like 'tokens', 'scope', 'comments'.
   */
  category?: string;
}

interface AuditResult {
  name: string;
  category: string;
  eslint: string | null;
  rslint: string | null;
  match: boolean;
}

async function runApiCheck(check: ApiCheck): Promise<AuditResult> {
  const text = check.text ?? 'const x = 1;';
  const trigger = check.trigger ?? 'Program:exit';

  // One rule definition reused on both sides — IDENTITY-comparable, so a
  // divergence is necessarily a runner-level difference (no plugin-side
  // drift introduced by accident).
  const ruleSpec: Rule.RuleModule = {
    meta: { type: 'problem', schema: [], messages: { probe: '{{result}}' } },
    create(ctx) {
      return {
        [trigger](node: never) {
          let result: string;
          try {
            // eslint-disable-next-line @typescript-eslint/no-explicit-any
            result = check.probe(ctx as any);
          } catch (e) {
            result = `THREW: ${(e as Error)?.message ?? String(e)}`;
          }
          ctx.report({ node, message: result });
        },
      } as never;
    },
  };

  // ── ESLint side ──
  const linter = new Linter({ configType: 'flat' });
  // Pass the same filename as rslint so `context.filename` matches
  // — without `options.filename` ESLint defaults to `<input>` and
  // our audit would falsely report a filename drift.
  const eslintMsgs = linter.verify(
    text,
    [
      {
        plugins: { audit: { rules: { probe: ruleSpec } } },
        rules: { 'audit/probe': 'error' },
      },
    ],
    { filename: 'audit.js' },
  );
  const eslintMsg = eslintMsgs[0]?.message ?? null;

  // ── rslint side ──
  // Build a LoadedPlugins map without going through the real plugin
  // loader (the rule isn't on npm). The runner accepts a synthesised
  // map; the only fields it inspects on each rule are `meta` and
  // `create`, both present above.
  const loaded: LoadedPlugins = {
    plugins: [],
    rules: new Map<string, unknown>([['audit/probe', ruleSpec]]),
  };
  const rslintResult = lintFile(
    {
      filePath: 'audit.js',
      text,
      rules: { 'audit/probe': { options: [] } },
      collectFixes: false,
      suggestionsMode: 'off',
    },
    loaded,
  );
  const rslintMsg = rslintResult.diagnostics[0]?.message ?? null;

  return {
    name: check.name,
    category: check.category ?? 'misc',
    eslint: eslintMsg,
    rslint: rslintMsg,
    match: eslintMsg === rslintMsg,
  };
}

// ─── API matrix ───────────────────────────────────────────────────────
//
// Each entry exercises ONE public API. Probes prefer simple,
// deterministic outputs (counts, joined values, types) over full
// object dumps — full dumps tend to differ on incidental fields
// (Symbol references, host-allocated identity) that aren't load-bearing
// for plugin authors.

const SAMPLE =
  'const x = 1;\nfunction f() { return x; }\n// trailing comment\n';

const apiChecks: ApiCheck[] = [
  // ── SourceCode core fields ──
  {
    name: 'SourceCode.text',
    category: 'sourcecode',
    text: SAMPLE,
    probe: (ctx) => `len=${ctx.sourceCode.text.length}`,
  },
  {
    name: 'SourceCode.ast.type',
    category: 'sourcecode',
    probe: (ctx) => ctx.sourceCode.ast.type,
  },
  {
    name: 'SourceCode.lines (getter)',
    category: 'sourcecode',
    text: 'a\nb\nc',
    probe: (ctx) => ctx.sourceCode.lines.join('|'),
  },
  {
    name: 'SourceCode.hasBOM',
    category: 'sourcecode',
    probe: (ctx) => String(ctx.sourceCode.hasBOM),
  },

  // ── Text retrieval ──
  {
    name: 'SourceCode.getText() — whole source',
    category: 'text',
    probe: (ctx) => `len=${ctx.sourceCode.getText().length}`,
  },
  {
    name: 'SourceCode.getText(node)',
    category: 'text',
    probe: (ctx) => ctx.sourceCode.getText(ctx.sourceCode.ast.body[0]),
  },
  {
    name: 'SourceCode.getText(node, before, after)',
    category: 'text',
    text: '/*pre*/ const x = 1; /*post*/',
    probe: (ctx) => ctx.sourceCode.getText(ctx.sourceCode.ast.body[0], 2, 2),
  },
  {
    name: 'SourceCode.getLines()',
    category: 'text',
    text: 'a\nb\nc',
    probe: (ctx) => ctx.sourceCode.getLines().join('|'),
  },

  // ── Position helpers ──
  {
    name: 'SourceCode.getLocFromIndex(0)',
    category: 'location',
    probe: (ctx) => {
      const loc = ctx.sourceCode.getLocFromIndex(0);
      return `line=${loc.line},col=${loc.column}`;
    },
  },
  {
    name: 'SourceCode.getLocFromIndex(end)',
    category: 'location',
    text: 'a\nbc\nd',
    probe: (ctx) => {
      const loc = ctx.sourceCode.getLocFromIndex(4); // 'c' on line 2
      return `line=${loc.line},col=${loc.column}`;
    },
  },
  {
    name: 'SourceCode.getIndexFromLoc',
    category: 'location',
    text: 'a\nbc\nd',
    probe: (ctx) =>
      String(ctx.sourceCode.getIndexFromLoc({ line: 2, column: 1 })),
  },
  {
    name: 'SourceCode.getLoc(node)',
    category: 'location',
    probe: (ctx) => {
      const loc = ctx.sourceCode.getLoc(ctx.sourceCode.ast.body[0]);
      return `${loc.start.line}:${loc.start.column}-${loc.end.line}:${loc.end.column}`;
    },
  },
  {
    name: 'SourceCode.getRange(node)',
    category: 'location',
    probe: (ctx) => {
      const r = ctx.sourceCode.getRange(ctx.sourceCode.ast.body[0]);
      return `${r[0]}-${r[1]}`;
    },
  },

  // ── Comments ──
  {
    name: 'SourceCode.getAllComments() count',
    category: 'comments',
    text: '// a\nconst x = 1; /* b */',
    probe: (ctx) => String(ctx.sourceCode.getAllComments().length),
  },
  {
    name: 'SourceCode.getCommentsBefore(node)',
    category: 'comments',
    text: '/* lead */ const x = 1;',
    probe: (ctx) => {
      const cs = ctx.sourceCode.getCommentsBefore(ctx.sourceCode.ast.body[0]);
      return cs.map((c: { value: string }) => c.value.trim()).join(',');
    },
  },
  // R4 — regression for the "all-preceding-comments" bug. Pre-fix
  // `getCommentsBefore` filtered solely on `c.range[1] <= node.range[0]`
  // so the second `VariableDeclaration` saw THREE comments instead of
  // the two that are actually adjacent to it. ESLint only returns
  // the ones between the previous code token and this node.
  {
    name: 'SourceCode.getCommentsBefore — multi-statement, only adjacent comments',
    category: 'comments',
    text: '/* file leading */\nconst a = 1;\n// before b\n/* still before b */\nconst b = 2;',
    probe: (ctx) => {
      const cs = ctx.sourceCode.getCommentsBefore(ctx.sourceCode.ast.body[1]);
      return cs.map((c: { value: string }) => c.value.trim()).join('|');
    },
  },
  {
    name: 'SourceCode.getCommentsAfter — multi-statement, only adjacent comments',
    category: 'comments',
    text: 'const a = 1;\n// after a\n/* still after a */\nconst b = 2;\n// trailing',
    probe: (ctx) => {
      const cs = ctx.sourceCode.getCommentsAfter(ctx.sourceCode.ast.body[0]);
      return cs.map((c: { value: string }) => c.value.trim()).join('|');
    },
  },
  // NOTE: `getJSDocComment` was removed in ESLint v10. Both engines
  // now return `undefined` for `typeof sourceCode.getJSDocComment`; we
  // pin that absence below in the "v10-removed APIs" block.
  {
    name: 'SourceCode.getCommentsAfter(node)',
    category: 'comments',
    text: 'const x = 1; /* trail */',
    probe: (ctx) => {
      const cs = ctx.sourceCode.getCommentsAfter(ctx.sourceCode.ast.body[0]);
      return cs.map((c: { value: string }) => c.value.trim()).join(',');
    },
  },
  {
    name: 'SourceCode.getCommentsInside(node)',
    category: 'comments',
    text: 'function f() { /* inside */ return 1; }',
    probe: (ctx) => {
      const cs = ctx.sourceCode.getCommentsInside(ctx.sourceCode.ast.body[0]);
      return cs.map((c: { value: string }) => c.value.trim()).join(',');
    },
  },
  {
    // Fixture wraps in `[]` so `a/*c*/b` parses as two array elements
    // (the bare statement `a /*c*/ b` is invalid syntactically).
    name: 'SourceCode.commentsExistBetween — yes',
    category: 'comments',
    text: 'let _ = [a, /*c*/ b];',
    probe: (ctx) => {
      const t = ctx.sourceCode.getTokens(ctx.sourceCode.ast.body[0]);
      const a = t.find((x: { value: string }) => x.value === 'a');
      const b = t.find((x: { value: string }) => x.value === 'b');
      return String(ctx.sourceCode.commentsExistBetween(a, b));
    },
  },
  {
    name: 'SourceCode.commentsExistBetween — no',
    category: 'comments',
    text: 'let _ = a + b;',
    probe: (ctx) => {
      const t = ctx.sourceCode.getTokens(ctx.sourceCode.ast.body[0]);
      const a = t.find((x: { value: string }) => x.value === 'a');
      const b = t.find((x: { value: string }) => x.value === 'b');
      return String(ctx.sourceCode.commentsExistBetween(a, b));
    },
  },

  // ── Spacing ──
  {
    name: 'SourceCode.isSpaceBetween — flush comment',
    category: 'spacing',
    text: 'let _ = [a,/*c*/b];',
    probe: (ctx) => {
      const tokens = ctx.sourceCode.getTokens(ctx.sourceCode.ast);
      const a = tokens.find((t: { value: string }) => t.value === 'a');
      const b = tokens.find((t: { value: string }) => t.value === 'b');
      return String(ctx.sourceCode.isSpaceBetween(a, b));
    },
  },
  {
    name: 'SourceCode.isSpaceBetween — actual space',
    category: 'spacing',
    text: 'let _ = [a, /*c*/ b];',
    probe: (ctx) => {
      const tokens = ctx.sourceCode.getTokens(ctx.sourceCode.ast);
      const a = tokens.find((t: { value: string }) => t.value === 'a');
      const b = tokens.find((t: { value: string }) => t.value === 'b');
      return String(ctx.sourceCode.isSpaceBetween(a, b));
    },
  },
  {
    name: 'SourceCode.isSpaceBetween — adjacent',
    category: 'spacing',
    text: 'let _ = [a,b];',
    probe: (ctx) => {
      const tokens = ctx.sourceCode.getTokens(ctx.sourceCode.ast);
      const a = tokens.find((t: { value: string }) => t.value === 'a');
      const b = tokens.find((t: { value: string }) => t.value === 'b');
      return String(ctx.sourceCode.isSpaceBetween(a, b));
    },
  },

  // ── Ancestors ──
  {
    name: 'SourceCode.getAncestors(node)',
    category: 'tree',
    text: 'const x = 1;',
    trigger: 'Literal',
    probe: (ctx) => {
      const a = ctx.sourceCode.getAncestors(
        ctx.sourceCode.ast.body[0].declarations[0].init,
      );
      return a.map((n: { type: string }) => n.type).join(',');
    },
  },
  {
    name: 'SourceCode.getNodeByRangeIndex(idx)',
    category: 'tree',
    text: 'const xx = 1;',
    probe: (ctx) => {
      const n = ctx.sourceCode.getNodeByRangeIndex(7);
      return n?.type ?? 'null';
    },
  },

  // ── Tokens — singular ──
  {
    name: 'SourceCode.getFirstToken(node)',
    category: 'tokens',
    text: 'const x = 1;',
    probe: (ctx) =>
      ctx.sourceCode.getFirstToken(ctx.sourceCode.ast.body[0]).value,
  },
  // R3 — singular token API `skip` argument. Pre-fix, rslint
  // silently ignored the number; treating `getFirstToken(node, 2)`
  // as if no opts were passed.
  {
    name: 'SourceCode.getFirstToken(node, skip=1)',
    category: 'tokens',
    text: 'const x = 1;',
    probe: (ctx) =>
      ctx.sourceCode.getFirstToken(ctx.sourceCode.ast.body[0], 1).value,
  },
  {
    name: 'SourceCode.getFirstToken(node, skip=2)',
    category: 'tokens',
    text: 'const x = 1;',
    probe: (ctx) =>
      ctx.sourceCode.getFirstToken(ctx.sourceCode.ast.body[0], 2).value,
  },
  {
    name: 'SourceCode.getFirstToken(node, { skip: 2 })',
    category: 'tokens',
    text: 'const x = 1;',
    probe: (ctx) =>
      ctx.sourceCode.getFirstToken(ctx.sourceCode.ast.body[0], { skip: 2 })
        .value,
  },
  {
    name: 'SourceCode.getTokenAfter(token, skip=1)',
    category: 'tokens',
    text: 'a + b + c;',
    probe: (ctx) => {
      const decl = ctx.sourceCode.ast.body[0];
      const first = ctx.sourceCode.getFirstToken(decl);
      return ctx.sourceCode.getTokenAfter(first, 1).value;
    },
  },
  {
    name: 'SourceCode.getLastToken(node)',
    category: 'tokens',
    text: 'const x = 1;',
    probe: (ctx) =>
      ctx.sourceCode.getLastToken(ctx.sourceCode.ast.body[0]).value,
  },
  {
    name: 'SourceCode.getTokenBefore(token)',
    category: 'tokens',
    text: 'const x = 1;',
    probe: (ctx) => {
      const tokens = ctx.sourceCode.getTokens(ctx.sourceCode.ast);
      return ctx.sourceCode.getTokenBefore(tokens[2])?.value ?? 'null';
    },
  },
  {
    name: 'SourceCode.getTokenAfter(token)',
    category: 'tokens',
    text: 'const x = 1;',
    probe: (ctx) => {
      const tokens = ctx.sourceCode.getTokens(ctx.sourceCode.ast);
      return ctx.sourceCode.getTokenAfter(tokens[0])?.value ?? 'null';
    },
  },
  {
    name: 'SourceCode.getTokenByRangeStart(offset)',
    category: 'tokens',
    text: 'const x = 1;',
    probe: (ctx) => ctx.sourceCode.getTokenByRangeStart(0)?.value ?? 'null',
  },

  // ── Tokens — plural ──
  {
    name: 'SourceCode.getTokens(node) — count',
    category: 'tokens',
    text: 'const x = 1;',
    probe: (ctx) =>
      String(ctx.sourceCode.getTokens(ctx.sourceCode.ast.body[0]).length),
  },
  {
    name: 'SourceCode.getFirstTokens(node, count)',
    category: 'tokens',
    text: 'const x = 1; const y = 2;',
    probe: (ctx) =>
      ctx.sourceCode
        .getFirstTokens(ctx.sourceCode.ast, 3)
        .map((t: { value: string }) => t.value)
        .join(','),
  },
  {
    name: 'SourceCode.getLastTokens(node, count)',
    category: 'tokens',
    text: 'const x = 1; const y = 2;',
    probe: (ctx) =>
      ctx.sourceCode
        .getLastTokens(ctx.sourceCode.ast, 3)
        .map((t: { value: string }) => t.value)
        .join(','),
  },
  {
    name: 'SourceCode.getTokensBefore(token, count)',
    category: 'tokens',
    text: 'a + b + c',
    probe: (ctx) => {
      const tokens = ctx.sourceCode.getTokens(ctx.sourceCode.ast.body[0]);
      const target = tokens[tokens.length - 1];
      return ctx.sourceCode
        .getTokensBefore(target, 2)
        .map((t: { value: string }) => t.value)
        .join(',');
    },
  },
  {
    name: 'SourceCode.getTokensAfter(token, count)',
    category: 'tokens',
    text: 'a + b + c',
    probe: (ctx) => {
      const tokens = ctx.sourceCode.getTokens(ctx.sourceCode.ast.body[0]);
      return ctx.sourceCode
        .getTokensAfter(tokens[0], 2)
        .map((t: { value: string }) => t.value)
        .join(',');
    },
  },
  {
    name: 'SourceCode.getTokensBetween(left, right)',
    category: 'tokens',
    text: 'a + b + c',
    probe: (ctx) => {
      const tokens = ctx.sourceCode.getTokens(ctx.sourceCode.ast.body[0]);
      return ctx.sourceCode
        .getTokensBetween(tokens[0], tokens[tokens.length - 1])
        .map((t: { value: string }) => t.value)
        .join(',');
    },
  },
  {
    name: 'SourceCode.getFirstTokens with options.filter',
    category: 'tokens',
    text: 'const x = 1; const y = 2;',
    probe: (ctx) =>
      ctx.sourceCode
        .getFirstTokens(ctx.sourceCode.ast, {
          count: 10,
          filter: (t: { type: string }) => t.type === 'Identifier',
        })
        .map((t: { value: string }) => t.value)
        .join(','),
  },

  // ── Scope ──
  {
    name: 'SourceCode.getScope(node).type',
    category: 'scope',
    text: 'const x = 1;',
    probe: (ctx) => ctx.sourceCode.getScope(ctx.sourceCode.ast).type,
  },
  // R6 — isGlobalReference must EXIST and not throw. Widely used by
  // community rules (unicorn no-typeof-undefined, no-useless-error-
  // capture-stack-trace, ESLint built-in no-implied-eval, etc.).
  // Pre-fix rslint had no implementation on `SourceCode`; calling it
  // crashed with TypeError. Verifying parity here pins both
  // (a) the method exists and (b) its return for the common cases
  // matches ESLint exactly.
  {
    name: 'SourceCode.isGlobalReference — Literal returns false',
    category: 'scope',
    text: 'const x = 1;',
    probe: (ctx) => {
      const lit = ctx.sourceCode.ast.body[0].declarations[0].init;
      return String(ctx.sourceCode.isGlobalReference(lit));
    },
  },
  {
    name: 'SourceCode.isGlobalReference — Identifier with local binding',
    category: 'scope',
    text: 'const x = 1; x;',
    probe: (ctx) => {
      // The reference to `x` on the right-hand `x;` expression — has
      // an in-source `defs[]` entry, so isGlobalReference is false.
      const exprStmt = ctx.sourceCode.ast.body[1];
      return String(ctx.sourceCode.isGlobalReference(exprStmt.expression));
    },
  },
  {
    name: 'SourceCode.isGlobalReference — unresolved Identifier',
    category: 'scope',
    text: 'undeclaredFoo;',
    probe: (ctx) => {
      // Free Identifier with no global declaration nor in-source
      // binding. ESLint flat-config has no browser/node globals by
      // default, so this resolves to nothing → isGlobalReference
      // returns false on both sides.
      const exprStmt = ctx.sourceCode.ast.body[0];
      return String(ctx.sourceCode.isGlobalReference(exprStmt.expression));
    },
  },
  {
    name: 'SourceCode.getDeclaredVariables(node) — count',
    category: 'scope',
    text: 'const a = 1, b = 2;',
    probe: (ctx) =>
      String(
        ctx.sourceCode.getDeclaredVariables(ctx.sourceCode.ast.body[0]).length,
      ),
  },
  // markVariableAsUsed — the SourceCode-level method that walks the
  // scope chain from `refNode`'s scope upward, sets `eslintUsed=true`
  // on the matching Variable, and returns whether one was found.
  // ESLint v10's `lib/languages/js/source-code/source-code.js` is the
  // reference implementation. Plugins like react/jsx-uses-vars call
  // this to suppress `no-unused-vars` for generated bindings; pinning
  // parity here covers a code path that has no other audit-api probe.
  {
    name: 'SourceCode.markVariableAsUsed — existing var → true',
    category: 'scope',
    text: 'const x = 1; x;',
    probe: (ctx) => {
      // The reference site `x;` is the listener target; ask
      // markVariableAsUsed to mark `x`. Should resolve via the scope
      // chain to the declared const.
      const refNode = ctx.sourceCode.ast.body[1].expression;
      const ok = ctx.sourceCode.markVariableAsUsed('x', refNode);
      return String(ok);
    },
  },
  {
    name: 'SourceCode.markVariableAsUsed — unknown name → false',
    category: 'scope',
    text: 'const x = 1;',
    probe: (ctx) => {
      // No `nope` binding anywhere in the scope chain → returns false.
      // Plugins rely on the negative return to skip when their
      // generated identifier doesn't actually appear.
      const ok = ctx.sourceCode.markVariableAsUsed('nope', ctx.sourceCode.ast);
      return String(ok);
    },
  },

  // ── RuleContext fields ──
  {
    name: 'context.id',
    category: 'context',
    probe: (ctx) => ctx.id,
  },
  {
    name: 'context.options (empty)',
    category: 'context',
    probe: (ctx) => `len=${ctx.options.length}`,
  },
  {
    name: 'context.cwd is non-empty string',
    category: 'context',
    probe: (ctx) =>
      typeof ctx.cwd === 'string' && ctx.cwd.length > 0 ? 'ok' : 'bad',
  },
  {
    name: 'context.filename',
    category: 'context',
    probe: (ctx) => {
      // Both ESLint and rslint normalise filenames differently; we
      // assert the BASENAME so cross-platform paths still match.
      const name = ctx.filename;
      return name?.split(/[/\\]/).pop() ?? 'null';
    },
  },
  {
    name: 'context.physicalFilename (v10)',
    category: 'context',
    probe: (ctx) => {
      const name = (ctx as { physicalFilename?: string }).physicalFilename;
      return typeof name === 'string' && name.length > 0 ? 'ok' : 'missing';
    },
  },
  {
    name: 'context.languageOptions presence',
    category: 'context',
    probe: (ctx) =>
      ctx.languageOptions && typeof ctx.languageOptions === 'object'
        ? 'ok'
        : 'missing',
  },
  {
    name: 'context.settings is object',
    category: 'context',
    probe: (ctx) =>
      ctx.settings && typeof ctx.settings === 'object' ? 'ok' : 'missing',
  },
  // ── APIs removed in ESLint v10 ──
  //
  // v10 dropped the legacy getters on RuleContext as well as
  // `parserPath` / `parserOptions` (the latter survived on
  // languageOptions). rslint mirrors the v10 surface, so both engines
  // now return `typeof === 'undefined'` for these. The probes pin
  // parity on absence in both directions — a re-introduction on
  // either side fails the audit immediately.
  {
    name: 'context.getCwd is removed (v10 alignment)',
    category: 'context',
    probe: (ctx) => `typeof=${typeof (ctx as { getCwd?: unknown }).getCwd}`,
  },
  {
    name: 'context.getFilename is removed (v10 alignment)',
    category: 'context',
    probe: (ctx) =>
      `typeof=${typeof (ctx as { getFilename?: unknown }).getFilename}`,
  },
  {
    name: 'context.getSourceCode is removed (v10 alignment)',
    category: 'context',
    probe: (ctx) =>
      `typeof=${typeof (ctx as { getSourceCode?: unknown }).getSourceCode}`,
  },
  {
    name: 'context.parserPath is removed (v10 alignment)',
    category: 'context',
    probe: (ctx) =>
      `typeof=${typeof (ctx as { parserPath?: unknown }).parserPath}`,
  },
  {
    name: 'context.parserOptions is removed (v10 alignment)',
    category: 'context',
    probe: (ctx) =>
      `typeof=${typeof (ctx as { parserOptions?: unknown }).parserOptions}`,
  },
  // v9 already removed these from RuleContext; v10 keeps them absent.
  // Plugins must reach for the equivalents on `sourceCode`.
  {
    name: 'context.getScope is removed (v9+ alignment)',
    category: 'context',
    probe: (ctx) => `typeof=${typeof (ctx as { getScope?: unknown }).getScope}`,
  },
  {
    name: 'context.getAncestors is removed (v9+ alignment)',
    category: 'context',
    probe: (ctx) =>
      `typeof=${typeof (ctx as { getAncestors?: unknown }).getAncestors}`,
  },
  {
    name: 'context.getDeclaredVariables is removed (v9+ alignment)',
    category: 'context',
    probe: (ctx) =>
      `typeof=${typeof (ctx as { getDeclaredVariables?: unknown }).getDeclaredVariables}`,
  },
  {
    name: 'context.markVariableAsUsed is removed (v9+ alignment)',
    category: 'context',
    probe: (ctx) =>
      `typeof=${typeof (ctx as { markVariableAsUsed?: unknown }).markVariableAsUsed}`,
  },
  // ── SourceCode APIs removed in ESLint v10 ──
  //
  // `getJSDocComment(node)` (no replacement), `getTokenOrCommentBefore/After`,
  // `isSpaceBetweenTokens` — pin absence on both sides.
  {
    name: 'SourceCode.getJSDocComment is removed (v10 alignment)',
    category: 'sourceCode',
    probe: (ctx) =>
      `typeof=${typeof (ctx.sourceCode as { getJSDocComment?: unknown }).getJSDocComment}`,
  },
  {
    name: 'SourceCode.getTokenOrCommentBefore is removed (v10 alignment)',
    category: 'sourceCode',
    probe: (ctx) =>
      `typeof=${typeof (ctx.sourceCode as { getTokenOrCommentBefore?: unknown }).getTokenOrCommentBefore}`,
  },
  {
    name: 'SourceCode.getTokenOrCommentAfter is removed (v10 alignment)',
    category: 'sourceCode',
    probe: (ctx) =>
      `typeof=${typeof (ctx.sourceCode as { getTokenOrCommentAfter?: unknown }).getTokenOrCommentAfter}`,
  },
  {
    name: 'SourceCode.isSpaceBetweenTokens is removed (v10 alignment)',
    category: 'sourceCode',
    probe: (ctx) =>
      `typeof=${typeof (ctx.sourceCode as { isSpaceBetweenTokens?: unknown }).isSpaceBetweenTokens}`,
  },
  // ── SourceCode properties added (or surfaced) in v10 ──
  {
    name: 'SourceCode.visitorKeys is present and has ESTree keys',
    category: 'sourceCode',
    probe: (ctx) => {
      const vk = (
        ctx.sourceCode as { visitorKeys?: Record<string, readonly string[]> }
      ).visitorKeys;
      if (!vk || typeof vk !== 'object') return 'missing';
      // Spot-check two stable ESTree node types
      const id = Array.isArray(vk.Identifier);
      const be = Array.isArray(vk.BinaryExpression);
      return id && be ? 'ok' : 'wrong-shape';
    },
  },
  {
    name: 'SourceCode.parserServices is plain {} by default',
    category: 'sourceCode',
    probe: (ctx) => {
      const ps = (ctx.sourceCode as { parserServices?: unknown })
        .parserServices;
      return ps && typeof ps === 'object' ? 'object' : 'missing';
    },
  },
  // ── getInlineConfigNodes coverage ──
  //
  // Regression pin for the bug where rslint's filter recognized only the
  // `eslint-disable*` / `eslint-enable*` directive family and silently
  // dropped `eslint <rule>: ...` (inline rule-config), `global` /
  // `globals` declarations, and `exported` annotations. ESLint v10
  // returns every comment whose trimmed value is a recognized inline
  // directive — the probe lists them in source order so any drift on
  // either engine surfaces as a string diff.
  {
    // Bare `eslint-disable` is intentionally NOT in the fixture: it
    // would suppress the audit's own probe rule on the ESLint side
    // (it stops the rule from firing, so `runApiCheck` sees `null`
    // instead of a comparable string). The disable-family is covered
    // by `eslint-disable-next-line <named>` below, which only
    // suppresses that single rule on the next line.
    name: 'SourceCode.getInlineConfigNodes — disable-line/rule/global/exported',
    category: 'sourceCode',
    text:
      '/* eslint-disable-next-line no-console */\n' +
      'console.log(1);\n' +
      '/* eslint no-console: "error" */\n' +
      '/* global foo:writable */\n' +
      '/* globals bar:readonly, baz */\n' +
      '/* exported zot */\n' +
      'const a = 1;',
    probe: (ctx) =>
      ctx.sourceCode
        .getInlineConfigNodes()
        .map(
          (c: { type: string; value: string }) => `${c.type}:${c.value.trim()}`,
        )
        .join('|'),
  },
  // ── Template literal segmentation ──
  //
  // Regression pin: rslint's tokenizer used to emit the entire
  // template literal — including the embedded `${...}` expressions —
  // as a single `Template` token. ESLint splits it into one Template
  // token per literal segment with the expression tokens interleaved.
  // Rules that walk tokens (`getTokenBefore/After`, prefer-template,
  // spacing rules) saw a completely different shape.
  {
    name: 'tokenizer — template literal split into segments with interleaved expressions',
    category: 'tokens',
    text: 'const x = `a${1+2}b${`nested-${3}`}c`;',
    probe: (ctx) =>
      ctx.sourceCode
        .getTokens(ctx.sourceCode.ast)
        .map((t: { type: string; value: string }) => `${t.type}:${t.value}`)
        .join('|'),
  },
  // ── PrivateIdentifier token type ──
  //
  // Regression pin: `#privateField` was tokenized as `Identifier:priv`
  // (without the `#`), losing the type distinction ESLint uses to
  // identify class private members. The tokenizer now emits
  // `PrivateIdentifier:priv` with the `#` stripped from `value` and
  // the range spanning both `#` + name.
  {
    name: 'tokenizer — private class field emits PrivateIdentifier',
    category: 'tokens',
    text: 'class C { #priv = 1; get p() { return this.#priv; } }',
    probe: (ctx) =>
      ctx.sourceCode
        .getTokens(ctx.sourceCode.ast)
        .map((t: { type: string; value: string }) => `${t.type}:${t.value}`)
        .join('|'),
  },
  // ── Shebang comment ──
  //
  // Regression pin: the tokenizer didn't recognize a leading `#!`
  // hashbang and tokenized it as a sequence of garbage tokens (`#`,
  // `!`, `/`, ...) with NO comment recorded. ESLint v10 produces a
  // single Comment of type `Shebang` with the `#!` stripped from
  // `value`. Empirically pinned via `getAllComments` + token count.
  {
    name: 'tokenizer — leading shebang produces single Shebang comment',
    category: 'tokens',
    text: '#!/usr/bin/env node\nconst x = 1;',
    probe: (ctx) => {
      const comments = ctx.sourceCode
        .getAllComments()
        .map(
          (c: { type: string; value: string }) => `${c.type}:${c.value.trim()}`,
        )
        .join('|');
      const tokenCount = ctx.sourceCode.getTokens(ctx.sourceCode.ast).length;
      return `comments=[${comments}] tokens=${tokenCount}`;
    },
  },
  // ── Contextual keywords ──
  //
  // Regression pin: rslint's lexer used to lump every reserved word
  // — including the *contextual* ones `await`, `async`, `of`, `as` —
  // into the `Keyword` token type. ESLint v10 / espree emits them as
  // `Identifier` (with the parser disambiguating by surrounding
  // context). Rules that branched on `token.type === 'Identifier'`
  // for these names silently dropped matches.
  {
    name: 'tokenizer — contextual keywords (await/async/of) are Identifier',
    category: 'tokens',
    text: 'async function g() { for (const x of arr) { await f(); } }',
    probe: (ctx) =>
      ctx.sourceCode
        .getTokens(ctx.sourceCode.ast)
        .map((t: { type: string; value: string }) => `${t.type}:${t.value}`)
        .join('|'),
  },
  // ── ECMAScript line terminators ──
  //
  // Regression pin: rslint counted only `\n` as a line break. The
  // ECMAScript spec recognises four line terminators — LF, CR (legacy
  // Mac), U+2028 (LINE SEPARATOR), U+2029 (PARAGRAPH SEPARATOR) —
  // and `\r\n` collapses to one. Two compounding bugs: (1) the
  // line-start offsets array missed CR/LS/PS, and (2) `isWhitespace`
  // in the tokenizer didn't skip LS/PS, so they got swept up by
  // `isIdentStart`'s `ch > 127` fallback and glued onto the next
  // character. Files with any of these separators produced wrong
  // `loc.line` for tokens past the missed break.
  //
  // CR-alone source (`a;\\rb;`) probes the line-counting path.
  {
    name: 'tokenizer — CR alone counts as a line break',
    category: 'tokens',
    text: 'a;\rb;',
    probe: (ctx) =>
      JSON.stringify({
        lines: ctx.sourceCode.lines.length,
        tokens: ctx.sourceCode
          .getTokens(ctx.sourceCode.ast)
          .map(
            (t: { value: string; loc: { start: { line: number } } }) =>
              `${t.value}@L${t.loc.start.line}`,
          ),
      }),
  },
  // ── Unicode whitespace ──
  //
  // Regression pin: NBSP (U+00A0) and other Unicode space-class chars
  // weren't classified as whitespace by the tokenizer, so the
  // `isIdentStart` fallback `ch > 127` swept them into the next
  // identifier — `const`<NBSP>`a` tokenized as a single 7-char
  // `Identifier` value `const a`. Real source hits this when code is
  // pasted from web editors that convert plain spaces to NBSP.
  {
    name: 'tokenizer — NBSP between tokens is whitespace, not identifier glue',
    category: 'tokens',
    // Note: the space between `const` and `a` below is U+00A0 NBSP.
    text: 'const a = 1;',
    probe: (ctx) =>
      ctx.sourceCode
        .getTokens(ctx.sourceCode.ast)
        .map((t: { type: string; value: string }) => `${t.type}:${t.value}`)
        .join('|'),
  },
];

// ─── Test ─────────────────────────────────────────────────────────────

describe('ESLint API audit', () => {
  // One test per API so the failure report enumerates each gap.
  for (const check of apiChecks) {
    test(`[${check.category ?? 'misc'}] ${check.name}`, async () => {
      const r = await runApiCheck(check);
      if (!r.match) {
        throw new Error(
          `API DRIFT: ${r.name}\n  eslint:  ${JSON.stringify(r.eslint)}\n  rslint:  ${JSON.stringify(r.rslint)}`,
        );
      }
      expect(r.match).toBe(true);
    });
  }
});

// ─── Public matrix export — consumed by ESLINT_API_AUDIT.md generator
//
// `runFullAudit()` runs every check sequentially and returns the
// matrix as plain data. Not invoked from the test suite (that already
// pins per-check pass/fail), but kept here so a future doc-generation
// task can serialize the matrix to markdown without re-listing the
// probes.

export async function runFullAudit(): Promise<AuditResult[]> {
  const results: AuditResult[] = [];
  for (const c of apiChecks) results.push(await runApiCheck(c));
  return results;
}

/* rslint-disable @typescript-eslint/no-unsafe-type-assertion */
/**
 * ESLint-compatible RuleContext factory for plugins running inside the
 * runner Worker. Surface tracks **ESLint v10** — members removed in v10
 * (`getJSDocComment`,
 * `getTokenOrCommentBefore/After`, `isSpaceBetweenTokens`, plus the
 * legacy `context.parserPath`/`parserOptions`/`getSourceCode`/
 * `getFilename`/`getCwd`) are intentionally absent so misbehaving
 * plugins fail loudly instead of running against a degraded shim.
 *
 * This file owns the **RuleContext** layer:
 *
 *   - `createRuleContext(opts)` — per-rule, per-file context with
 *     `report()`, `options`, `settings`, `languageOptions`,
 *     `filename`, `physicalFilename`, `cwd`, `sourceCode`, and a
 *     private diagnostic collector drained via `_drainDiagnostics()`
 *     after the AST visit.
 *
 * The heavier sub-modules live next door:
 *
 *   - `source-code.ts`        — `createSourceCode` + all SourceCode/Token
 *                                APIs (token/comment/scope/spacing).
 *   - `diagnostic-builder.ts` — `buildDiagnostic` + `normalizeReportArgs`
 *                                + diagnostic-related types.
 *
 * All public symbols those files expose are re-exported here so
 * downstream code (`import { ... } from './context.js'`) keeps
 * working unchanged.
 */

import { makeFixer } from './fixer.js';
import { applyOptionDefaults, type RuleSchema } from './options-defaults.js';
import {
  type ESTreeNode,
  type SourceCode,
} from '../source-code/source-code.js';
import {
  buildDiagnostic,
  normalizeReportArgs,
  type Diagnostic,
  type MessagesMap,
  type ReportDescriptor,
  type SuggestionsMode,
} from './diagnostic-builder.js';

// ─────────────────────────────────────────────────────────────────────
// Public types
// ─────────────────────────────────────────────────────────────────────

/**
 * The forwarded subset of ESLint flat-config `languageOptions` exposed
 * to plugin rules as `context.languageOptions`. Plugin rules read it
 * to branch on `languageOptions.parserOptions.ecmaFeatures.jsx`,
 * `parserOptions.sourceType`, or `languageOptions.globals`.
 *
 * Field set matches what the wire payload from Go's
 * `linter.CompatLanguageOptions` carries; out-of-band fields the
 * runner doesn't reproduce (custom `parser`, `parserOptions.project`,
 * `parserOptions.tsconfigRootDir`) are intentionally absent. This is
 * the contract the linter package and the plugin runtime share — any
 * change here must update both `ecma-language-plugin.ts:LintFileRequest`
 * and `internal/linter/types.go:CompatLanguageOptions`.
 */
export interface LanguageOptions {
  /**
   * ECMAScript version target — top-level in ESLint v10. Plugin rules
   * read this as `ctx.languageOptions.ecmaVersion`. The v8-era
   * `parserOptions.ecmaVersion` nesting is intentionally NOT mirrored
   * here: rslint targets v10 cleanly, and we don't expose this field
   * to user config, so there is no legacy v8 surface to preserve.
   */
  ecmaVersion?: number | 'latest';
  /** Module / script / commonjs — top-level in ESLint v10. Same rationale as `ecmaVersion`. */
  sourceType?: 'module' | 'script' | 'commonjs';
  globals?: Record<string, 'readonly' | 'writable' | 'off'>;
  /**
   * Parser-specific extras. Only `ecmaFeatures` lives here in v10
   * (rules that gate on `jsx` / `globalReturn` / `impliedStrict` read
   * via `ctx.languageOptions.parserOptions.ecmaFeatures.*`). Custom
   * `parser` instances aren't supported by the runner, so we don't
   * include that field.
   */
  parserOptions?: {
    ecmaFeatures?: {
      jsx?: boolean;
      globalReturn?: boolean;
      impliedStrict?: boolean;
    };
  };
}

/**
 * Inputs to `createRuleContext`. Everything the rule needs to behave
 * identically to ESLint, plus the few knobs the runner uses to control
 * cost (suggestionsMode, collectFixes).
 */
export interface CreateContextOptions {
  ruleName: string; // e.g. 'unicorn/no-null'
  filePath: string;
  /** Raw user options (before schema defaults). */
  userOptions: readonly unknown[];
  /** rule.meta.schema for default-filling user options. */
  schema?: RuleSchema;
  /**
   * `rule.meta.defaultOptions` (ESLint v10): per-slot default objects
   * that get deep-merged with the user's values before invocation. Many
   * unicorn rules use this to ship per-property booleans (e.g.
   * `prefer-number-properties`'s `{ checkInfinity: false, checkNaN: true }`)
   * that are NOT expressible as `schema[i].default`. Without forwarding
   * this, rules silently see `{}` for their options and diverge from
   * ESLint's diagnostic set.
   */
  defaultOptions?: readonly unknown[];
  /** Merged flat-config settings. */
  settings: Record<string, unknown>;
  /**
   * Forwarded `languageOptions` for this file. Read by plugin rules
   * via `context.languageOptions`. Many real-world plugins branch on
   * `languageOptions.parserOptions.ecmaFeatures.jsx`,
   * `parserOptions.sourceType`, or `languageOptions.globals` — passing
   * undefined here is acceptable but loses that compatibility surface.
   */
  languageOptions?: LanguageOptions;
  /** Source text (the wire-supplied byte source — not re-read from disk). */
  text: string;
  /**
   * Per-file line-start-offsets cache (already built by `lintFile` for
   * normalize). Forwarded into `buildDiagnostic` so the `descriptor.loc`
   * report path doesn't rebuild lso per report — `no-unused-vars` paths
   * `loc:` into every descriptor, making this a 70+ms wall hit on
   * vscode/src without the cache.
   */
  lsoCache?: number[];
  /**
   * The SourceCode for this file, built once by `lintFile` (carrying the native parser's
   * token/comment streams) and shared across every rule + `applyDisableDirectives` so they
   * all hit the same lazy caches (tokens/comments/scopeManager). Required — per ESLint v10 a
   * RuleContext always has a SourceCode; there is no token-less mode.
   */
  sourceCode: SourceCode;
  /** rule.meta.messages map for messageId → template lookup. */
  messages?: MessagesMap;
  /**
   * Whether to materialise `descriptor.fix(fixer)` into the diagnostic's
   * `fixes` payload. The runner never applies fixes — application is the
   * caller's job (CLI fix-loop, LSP code-action). CLI sets this when
   * `--fix` is on; LSP sets it unconditionally so Quick Fix /
   * source.fixAll see plugin-rule fixes the same way they see native-rule
   * ones.
   */
  collectFixes: boolean;
  /** Suggestion handling: 'off' skips fix(fixer); 'eager' invokes it. */
  suggestionsMode: SuggestionsMode;
}

/**
 * The shape of the context object passed to `rule.create(ctx)`. ESLint's
 * actual type has many more fields; we expose what real plugins read.
 *
 * Surface tracks **ESLint v10** exactly — empirically pinned against
 * `eslint@10.x` so any v10 plugin reads the same property set whether
 * it runs under ESLint or rslint.
 */
export interface RuleContext {
  id: string;
  options: readonly unknown[];
  settings: Record<string, unknown>;
  /**
   * The full forwarded `languageOptions` object. Always present (never
   * undefined) — when the user's config didn't set anything, nested
   * fields stay undefined but the wrapper itself is here so a rule's
   * `ctx.languageOptions.globals` access doesn't crash.
   */
  languageOptions: LanguageOptions;
  filename: string;
  /**
   * Physical disk path of the file. For non-processor files this equals
   * `filename`. Distinct from `filename` only when ESLint applies a
   * processor that yields virtual sub-files; the runner doesn't run
   * processors today, so they're equal.
   */
  physicalFilename: string;
  cwd: string;
  sourceCode: SourceCode;

  // The reporter rules call.
  report(descriptor: ReportDescriptor): void;

  // Runner-only accessors (not part of ESLint's public surface).
  /** @internal */ _drainDiagnostics(): Diagnostic[];

  // NOTE: v10 removed the following from RuleContext entirely:
  //   parserPath, parserOptions,
  //   getSourceCode(), getFilename(), getCwd(), getPhysicalFilename(),
  //   getScope(), getAncestors(), getDeclaredVariables(),
  //   markVariableAsUsed().
  // Plugin code targeting v10 reads `sourceCode` / `filename` / `cwd`
  // directly or uses the `sourceCode.*` equivalents for scope helpers.
}

export type ListenerFn = (node: ESTreeNode) => void;

// ─────────────────────────────────────────────────────────────────────
// Re-exports — keep `./context.js` as a stable barrel for existing
// callers (`import { ... } from './context.js'`).
// ─────────────────────────────────────────────────────────────────────

export {
  createSourceCode,
  type ESTreeNode,
  type SourceCode,
  type SourceCodeBuildInput,
  type TokenCountOpts,
  type TokenFilterOpts,
  type TokenSkipOpts,
} from '../source-code/source-code.js';
export {
  buildDiagnostic,
  normalizeReportArgs,
  type Diagnostic,
  type MessagesMap,
  type ReportDescriptor,
  type SuggestionDescriptor,
  type SuggestionInput,
  type SuggestionsMode,
} from './diagnostic-builder.js';

// ─────────────────────────────────────────────────────────────────────
// RuleContext factory
// ─────────────────────────────────────────────────────────────────────

/**
 * Build a RuleContext for one rule on one file. The context captures a
 * private diagnostic collector; the runner drains it via
 * `_drainDiagnostics()` after `rule.create(ctx)` returns and the AST
 * visit completes. Listener registration goes through the standard
 * ESLint `rule.create(ctx)` return value — the earlier rslint-only
 * `ctx.on()` / `ctx.onExit()` extension has been removed in favor of
 * straight ESLint API parity.
 */
export function createRuleContext(opts: CreateContextOptions): RuleContext {
  const { ruleName, filePath, settings, text, collectFixes, suggestionsMode } =
    opts;
  const messages = opts.messages ?? {};
  const finalOptions = applyOptionDefaults(
    opts.userOptions,
    opts.schema,
    opts.defaultOptions,
  );

  // lintFile builds one SourceCode per file (carrying the native parser's token/comment
  // streams) and passes the same instance to every rule's createRuleContext, so the lazy
  // state (tokens/comments/scopeManager) primes once and is reused across rules. Per ESLint
  // v10 it is always supplied — there is no token-less RuleContext.
  const sourceCode = opts.sourceCode;

  const fixer = makeFixer();
  const diagnostics: Diagnostic[] = [];

  // Captured at construction so all callsites are stable references.
  const cwd = process.cwd();

  // Build the languageOptions view passed to plugin rules. The wrapper
  // is always present (never undefined) so a rule's
  // `ctx.languageOptions.globals` access doesn't crash; nested fields
  // stay undefined when the user didn't set them.
  //
  // We do NOT deep-clone the user's input — plugins are expected to
  // treat languageOptions as read-only per ESLint convention. A
  // misbehaving plugin that mutates the object would leak across
  // files within a worker; if that turns out to happen we can
  // shallow-freeze, but the cost is real on hot lint paths so we
  // defer until needed.
  const languageOptions: LanguageOptions = opts.languageOptions ?? {};

  const ctx: RuleContext = {
    id: ruleName,
    options: finalOptions,
    settings,
    languageOptions,
    filename: filePath,
    // The runner doesn't run processors today, so the virtual filename
    // and the physical disk path are identical. If processor support
    // lands, the producer side must pass the underlying physical path
    // separately and this assignment splits.
    physicalFilename: filePath,
    cwd,
    sourceCode,

    report(...args: unknown[]) {
      // Accept BOTH the modern descriptor form (`report({...})`) and
      // ESLint's legacy positional forms. Mirrors the behavior of
      // `normalizeMultiArgReportCall` in ESLint's
      // `lib/linter/file-report.js`:
      //
      //   1 arg              → object descriptor (shallow-cloned so a
      //                        rule mutating the descriptor mid-report
      //                        can't desync downstream code).
      //   2+ args, args[1] = string → (node, message, data?, fix?)
      //   2+ args, otherwise        → (node, loc, message, data?, fix?)
      //
      // Without this branch, plugins authored against older ESLint
      // versions silently drop diagnostics — `descriptor.node` /
      // `descriptor.message` are both undefined, and `buildDiagnostic`
      // returns null in that case.
      const descriptor = normalizeReportArgs(args);
      const diag = buildDiagnostic({
        ruleName,
        descriptor,
        text,
        messages,
        fixer,
        collectFixes,
        suggestionsMode,
        lsoCache: opts.lsoCache,
      });
      if (diag) diagnostics.push(diag);
    },

    // Returns a defensive copy AND clears the source. Multiple drains
    // are not expected (lintFile calls each ctx exactly once after the
    // visit), but if a future caller drains twice they'll correctly
    // see "no more diagnostics" rather than getting the same array
    // reference for the rest of the worker's lifetime. Splice with
    // index 0 + delete-count = length empties the array in place
    // without reallocation.
    _drainDiagnostics: () => {
      const out = diagnostics.slice();
      diagnostics.length = 0;
      return out;
    },
  };

  return ctx as unknown as RuleContext;
}

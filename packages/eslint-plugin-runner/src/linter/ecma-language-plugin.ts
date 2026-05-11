/* eslint-disable @typescript-eslint/no-unsafe-type-assertion -- AST / parser / scope-manager / plugin-API boundary casts. Each site projects from an `any` / `unknown` peer surface (oxc-parser output, user plugin objects, ESLint v10 wire shapes) into the typed shape this module uses; the contract is runtime-validated at the call boundaries above, not at the cast. Bulk-disabling here instead of per-line keeps the cast sites readable. */
/**
 * EcmaLanguagePlugin: the full per-file lint pipeline for JS/JSX/TS/TSX
 * files. Wires together (in dependency order):
 *
 *   plugin-loader     — loaded once at Worker init
 *   normalize-ast     — per-file
 *   scope-factory     — per-file lazy
 *   options-defaults  — per-rule per-file (inside context)
 *   context           — per-rule per-file
 *   tokenizer         — lazy inside context
 *   listener-merge    — per-file
 *   fixer             — per-rule per-file (inside context)
 *   suggestions       — gated by suggestionsMode (inside context)
 *
 * Designed to be called from a Worker thread. Stateless across files
 * except for the loaded-plugins cache, which is shared per Worker.
 *
 * Errors are surfaced via the result object's `parseError` and
 * `ruleErrors` fields rather than thrown — a single broken file must
 * not abort the batch.
 *
 * Caveat on `parseError`: it is set only when `parseSync` THROWS
 * (an unrecoverable failure). oxc is a RECOVERING parser — for most
 * syntax errors it returns a best-effort AST plus a `parsed.errors`
 * list rather than throwing. The runner does NOT read `parsed.errors`,
 * so such files are linted against the recovery AST instead of being
 * reported as a parse error the way espree (which aborts on any syntax
 * error) would. This is a deliberate oxc-vs-espree difference, not an
 * oversight; see `plugin-lint-flow.test.ts`'s "parse error captured
 * per-file" case.
 */

import { readFileSync } from 'node:fs';
import { parseSync } from 'oxc-parser';

import {
  buildLineStartOffsets,
  buildUtf16ToByteMap,
  normalizeAst,
} from '../ast/normalize-ast.js';
import {
  makeScopeManagerFactory,
  seedEcmaGlobals,
  seedGlobals,
} from '../ast/scope-factory.js';
import { mergeListeners, visit } from './listener-merge.js';
import {
  createRuleContext,
  createSourceCode,
  type Diagnostic,
  type ESTreeNode,
  type ListenerFn,
  type RuleContext,
  type SuggestionsMode,
} from './context.js';
import type { LoadedPlugins } from '../plugin/plugin-loader.js';
import type { Fix } from './fixer.js';
import { applyDisableDirectives } from './apply-disable-directives.js';

/**
 * Per-file rule resolution. The worker has already picked the right
 * `LoadedPlugins` for this file (via `configKey` → per-config map);
 * `lintFile` only consults that selected set's `rules` lookup.
 *
 * Returning `null` means "rule not found" — Go's compat dispatcher
 * surfaces that as a `ruleErrors` entry. Same prefix can appear under
 * multiple configs in a monorepo, but each config has its own
 * `LoadedPlugins`, so there are no cross-config collisions here.
 */
function resolveRule(ruleName: string, loadedPlugins: LoadedPlugins): unknown {
  return loadedPlugins.rules.get(ruleName) ?? null;
}

// ─── Public types ────────────────────────────────────────────────────

/** Per-rule configuration as it reaches the Worker. Already filtered to enabled rules; severity is reattached Go-side. */
export interface RuleConfig {
  /** Rule options — typically [optionsObject] or [], pre-schema-defaults. */
  options: readonly unknown[];
  /**
   * `rule.meta` is opaque to the runner except for:
   *   - schema      → applyOptionDefaults
   *   - messages    → context.report messageId lookup
   *   - fixable     → currently informational only; collectFixes gating is
   *                   done at the lintBatch level (internal/linter)
   */
  meta?: {
    schema?: unknown;
    messages?: Record<string, string>;
    fixable?: 'code' | 'whitespace';
    hasSuggestions?: boolean;
  };
}

/** Per-file lint request input (called by Worker dispatcher per task).
 *
 * The Worker reads source text from disk via `fs.readFileSync(filePath)`
 * by default — text is intentionally NOT carried over IPC. This drops
 * the structuredClone cost of shipping every file's contents across
 * the worker_threads boundary (~60 MB on a 5000-file repo).
 *
 * Multi-pass --fix coherence is preserved because cmd/rslint's
 * applyFixPass writes fixes to disk BEFORE re-dispatching the next
 * lint pass — the worker reads the post-fix contents.
 *
 * `text` here is an in-process override for unit tests that want to
 * exercise `lintFile` against an in-memory source. The wire shape
 * (engine.ts → worker postMessage) NEVER carries text.
 */
export interface LintFileRequest {
  filePath: string;
  /** In-process override for tests. The IPC wire shape never carries this. */
  text?: string;
  /**
   * Forwarded subset of user `languageOptions`. Only the fields the
   * runner actually consumes are typed here; the rest of the user's
   * `languageOptions` (parser custom hooks, plugin-specific extensions)
   * is intentionally dropped at the IPC boundary because the worker
   * doesn't reproduce ESLint's full language plugin pipeline.
   *
   * Per ESLint v10's flat-config spec
   * (https://eslint.org/docs/v10.x/use/configure/language-options),
   * `ecmaVersion` / `sourceType` / `globals` are TOP-LEVEL properties
   * of `languageOptions`. `parserOptions` is reserved for parser-
   * specific extras (`ecmaFeatures`, `allowReserved`). The v8-era
   * positions (`parserOptions.ecmaVersion` / `parserOptions.sourceType`)
   * are NOT accepted — rslint targets v10 cleanly.
   */
  languageOptions?: {
    ecmaVersion?: number | 'latest';
    sourceType?: 'module' | 'script' | 'commonjs';
    globals?: Record<string, 'readonly' | 'writable' | 'off'>;
    parserOptions?: {
      ecmaFeatures?: {
        jsx?: boolean;
        globalReturn?: boolean;
        impliedStrict?: boolean;
      };
    };
  };
  /** Merged flat-config `settings` for plugin consumption (e.g. `react.version`). */
  settings?: Record<string, unknown>;
  /** Map of fully-qualified rule name → config. Already enabled-filtered. */
  rules: Record<string, RuleConfig>;
  /**
   * Whether to materialise plugin `descriptor.fix(fixer)` into the
   * diagnostic's `fixes` payload. The runner never APPLIES fixes —
   * application is the caller's job (CLI fix-loop, LSP code-action /
   * fixAll). CLI sets this whenever `--fix` is on; LSP always sets it so
   * Quick Fix / source.fixAll see plugin-rule fixes the same way they
   * see native-rule ones.
   */
  collectFixes: boolean;
  suggestionsMode: SuggestionsMode;
  /** Optional Int32Array(SharedArrayBuffer) cancel flag, length-1, for per-node Atomics polling. */
  cancelFlag?: Int32Array;
  /**
   * Identity of the rslint config that owns THIS file — the
   * `configDirectory` Go writes into `CompatLintFile.ConfigKey`. The
   * worker uses this to pick the right `LoadedPlugins` from its
   * per-config map (`Map<configDirectory, LoadedPlugins>`).
   *
   * The `lintFile` pipeline never reads this field directly — the
   * worker has already selected `LoadedPlugins` by the time `lintFile`
   * runs. It exists on the request only so the worker dispatcher can
   * route the task before delegating.
   */
  configKey?: string;
}

/** Per-file lint response — one record per LintFileRequest. */
export interface LintFileResult {
  filePath: string;
  diagnostics: Diagnostic[];
  /** Aggregated fixes (sum of diagnostic.fixes for ApplyRuleFixes). For convenience; same data as diagnostics[].fixes. */
  fixes: Fix[];
  /** Aggregated suggestions, same convenience as fixes. */
  suggestionsCount: number;
  /** True iff the visit was cancelled mid-flight (cancelFlag observed). */
  cancelled: boolean;
  /** Set when oxc-parser failed; diagnostics is empty in this case. */
  parseError?: string;
  /** Per-rule errors caught during create() / listener execution. */
  ruleErrors?: Array<{ rule: string; message: string }>;
}

/**
 * Lint a single file. Stateless — `loadedPlugins` is owned by the
 * Worker; this function does not mutate it.
 */
export function lintFile(
  req: LintFileRequest,
  loadedPlugins: LoadedPlugins,
): LintFileResult {
  const result: LintFileResult = {
    filePath: req.filePath,
    diagnostics: [],
    fixes: [],
    suggestionsCount: 0,
    cancelled: false,
  };

  // Source resolution: prefer req.text when present, else read disk.
  // req.text carries an unsaved editor buffer's content on the LSP / --api
  // path (finding #3) — and synthetic source for in-process callers. On the
  // CLI path it's absent and the worker reads disk via readFileSync (disk
  // is authoritative there; see LintFileRequest's doc comment for the
  // multi-pass --fix coherence rationale).
  let sourceText: string;
  if (req.text !== undefined) {
    sourceText = req.text;
  } else {
    try {
      sourceText = readFileSync(req.filePath, 'utf8');
    } catch (err) {
      result.parseError = `worker fs.readFile failed for ${req.filePath}: ${(err as Error)?.message ?? String(err)}`;
      return result;
    }
  }

  // Strip a leading UTF-8 BOM before parsing. ESLint's `SourceCode`
  // constructor does the same — without this, oxc-parser keeps the
  // BOM byte in its source view and every node range comes back
  // shifted +1 UTF-16 / +3 UTF-8 from what ESLint would emit, so all
  // wire diagnostics on BOM-prefixed files would point one column
  // late. `SourceCode.hasBOM` still reflects the original file via
  // the `hasBOM` flag we thread through to createSourceCode below.
  const hadBOM = sourceText.charCodeAt(0) === 0xfeff;
  if (hadBOM) sourceText = sourceText.slice(1);

  const lso = buildLineStartOffsets(sourceText);

  // ── 1. parse ─────────────────────────────────────────────────────
  //
  // Determine oxc-parser `lang` from the file extension plus the
  // user's `parserOptions.ecmaFeatures.jsx`. By default oxc-parser
  // infers `lang` from the extension (.js → 'js', .jsx → 'jsx',
  // .ts → 'ts', .tsx → 'tsx'), which is the right answer for files
  // with conventional extensions. But ESLint flat-config users
  // routinely write JSX in `.js` files and tell ESLint about it via
  // `parserOptions.ecmaFeatures.jsx = true`. Without lifting that
  // flag into a `lang: 'jsx'` (or `'tsx'`) override, oxc-parser
  // refuses to parse `<Foo />` and the file silently fails with a
  // parse error.
  const jsxEnabled =
    req.languageOptions?.parserOptions?.ecmaFeatures?.jsx === true;
  // ESLint v10: `sourceType` lives at the top level of `languageOptions`.
  // Default is `'module'` (matches v10's flat-config default).
  const sourceTypeRaw = req.languageOptions?.sourceType ?? 'module';
  // KNOWN LIMITATION — `languageOptions.ecmaVersion` is NOT enforced at
  // parse time. oxc's `ParserOptions` has no `ecmaVersion` knob; it
  // always parses the latest grammar. So a file configured with a low
  // `ecmaVersion` that uses newer syntax (e.g. `a ??= b` under
  // `ecmaVersion: 2018`) parses cleanly here, whereas espree would
  // report a parse error. This is an under-reporting-only gap (the
  // runner never MIS-parses valid code); `ecmaVersion` still flows to
  // the scope analyzer and `context.languageOptions`. Enforcing it
  // would require a post-parse syntax-feature gate we don't currently
  // implement.
  const parserOpts: {
    sourceType: 'module' | 'script';
    lang?: 'js' | 'jsx' | 'ts' | 'tsx' | 'dts';
    preserveParens: boolean;
  } = {
    sourceType: sourceTypeRaw === 'commonjs' ? 'script' : sourceTypeRaw,
    // ESLint v10 + espree strip parentheses by default — the AST
    // doesn't carry `ParenthesizedExpression` wrapper nodes. oxc
    // defaults the other way (preserves the wrapper), so without this
    // override plugin rules that walk `member.object.type` see
    // `'ParenthesizedExpression'` instead of the inner expression's
    // type (e.g. `(await fetch()).foo` looks like a paren-of-await
    // member access, not a member access on AwaitExpression). Setting
    // `preserveParens: false` aligns the AST shape with v10.
    preserveParens: false,
    // `range: true` is supported but intentionally left off: oxc-parser
    // would emit a `[start, end]` array on every node, and the NAPI
    // marshalling cost of that extra per-node array outweighs the
    // JS-side allocation it would save inside normalize-ast.
    // normalize-ast already gates its `node.range = [start, end]`
    // assignment on `node.range == null`, so re-enabling here is a
    // one-line change if oxc's Rust→JS marshalling cost ever drops.
  };
  if (jsxEnabled) {
    // Promote .ts/.mts/.cts → tsx and .js/.mjs/.cjs → jsx so JSX
    // syntax parses cleanly. .jsx / .tsx / .dts already imply their
    // own lang; leave them alone so an explicit
    // `ecmaFeatures.jsx = false` on a `.tsx` file doesn't downgrade
    // (we ignore `false` here — explicit downgrade isn't a real
    // workflow and oxc treats .tsx as TSX regardless).
    //
    // Pre-fix the `.ts` branch was `/\.ts$/i` which doesn't match
    // `.mts` / `.cts`; the second branch only covered `.js`/.mjs`/
    // `.cjs`, so `Component.mts` with JSX silently fell through to
    // oxc's extension-inferred `ts` mode and emitted parse errors
    // for every JSX element — every rule on the file went blind.
    if (/\.[mc]?ts$/i.test(req.filePath)) {
      parserOpts.lang = 'tsx';
    } else if (/\.[mc]?js$/i.test(req.filePath)) {
      parserOpts.lang = 'jsx';
    }
  }
  let parsed: { program: unknown; comments?: unknown };
  try {
    parsed = parseSync(req.filePath, sourceText, parserOpts);
  } catch (err) {
    result.parseError = `parse: ${(err as Error)?.message ?? String(err)}`;
    return result;
  }

  // newer oxc-parser returns objects directly; older versions returned
  // JSON strings. Handle both for forward/backward compat.
  const ast = (
    typeof parsed.program === 'string'
      ? JSON.parse(parsed.program)
      : parsed.program
  ) as ESTreeNode;

  // ── 2. normalize ─────────────────────────────────────────────────
  try {
    normalizeAst(ast, lso, sourceText);
  } catch (err) {
    result.parseError = `normalize: ${(err as Error)?.message ?? String(err)}`;
    return result;
  }

  // ── 3. scope-manager factory (lazy) ──────────────────────────────
  // ESLint v10: sourceType / ecmaVersion / globals are top-level fields
  // of `languageOptions`. No legacy positions are accepted.
  const globals = req.languageOptions?.globals;
  const ecmaFeatures = req.languageOptions?.parserOptions?.ecmaFeatures;
  const scopeManagerFactory = (() => {
    const inner = makeScopeManagerFactory(ast, {
      filePath: req.filePath,
      sourceType: req.languageOptions?.sourceType ?? 'module',
      ecmaVersion: req.languageOptions?.ecmaVersion,
      globals,
      // Forward each ecmaFeatures-derived knob the scope analyzers
      // honour. Pre-fix, the runner hard-coded `impliedStrict: true`
      // and ignored `globalReturn` entirely, silently breaking
      // sourceType:'script' bundles and Node CJS scripts. Defaults
      // are filled in by `scope-factory` to match ESLint v10.
      impliedStrict: ecmaFeatures?.impliedStrict,
      globalReturn: ecmaFeatures?.globalReturn,
    });
    return () => {
      const sm = inner();
      // Seed lazily — only when the scope-manager is materialized.
      // Order matters:
      //   1. seedEcmaGlobals injects the ES built-in set (parseInt /
      //      Array / NaN / ...) into globalScope.variables and
      //      resolves matching `through` refs against them.
      //   2. seedGlobals layers the user's `languageOptions.globals`
      //      on top, possibly overriding readonly/writable mode and
      //      re-running the resolve pass for any new refs.
      // This mirrors ESLint flat-config's behavior — without step 1,
      // rules built on `ReferenceTracker.iterateGlobalReferences` see
      // an empty global scope and silently never fire.
      seedEcmaGlobals(sm);
      seedGlobals(sm, globals);
      return sm;
    };
  })();

  // ── 4. shared SourceCode + per-rule contexts ─────────────────────
  // Build ONE SourceCode for the file. Every rule context wires the
  // same instance via `opts.sourceCode`, so the lazy caches (tokens,
  // comments, scopeManager) prime once and are reused across rules.
  // `parsedComments` is the oxc-emitted compact comment list — when
  // present, `getAllComments` builds an ESLint-shape Comment[] directly
  // (cheap loc adapter) instead of invoking the full text tokenizer.
  const sharedSourceCode = createSourceCode({
    text: sourceText,
    ast,
    scopeManagerFactory,
    // BOM was already stripped above; thread the original-file flag so
    // `SourceCode.hasBOM` reflects whether the source on disk started
    // with a BOM (matches ESLint's API).
    hasBOM: hadBOM,
    // JSX lexing trigger: explicit `ecmaFeatures.jsx` flag OR a
    // `.jsx` / `.tsx` extension (oxc auto-infers JSX from those
    // extensions, so the tokenizer must enable JSX-aware lexing
    // even when the flag is absent).
    jsx:
      jsxEnabled ||
      parserOpts.lang === 'jsx' ||
      parserOpts.lang === 'tsx' ||
      /\.(jsx|tsx)$/i.test(req.filePath),
    parsedComments: parsed.comments as
      | ReadonlyArray<{
          type: 'Line' | 'Block';
          value: string;
          start: number;
          end: number;
        }>
      | undefined,
  });

  // ESLint v10's espree attaches `Program.tokens` (and `Program.comments`)
  // to the AST root directly. oxc-parser doesn't, so plugin rules that
  // read `program.tokens` (e.g. `eslint-plugin-import/no-empty-named-blocks`)
  // see `undefined` and throw. Attach a lazy getter that delegates to
  // the shared SourceCode's tokenizer — only pays the tokenize cost if
  // a rule actually consumes the field.
  if (!('tokens' in (ast as object))) {
    // Memoize on first access. `getTokens(ast)` rebuilds the array
    // every call (each call allocates a fresh `out` and re-walks the
    // tokens stream), so `program.tokens !== program.tokens` pre-fix.
    // Rules that use the tokens array as a WeakMap/Set key or rely on
    // `===` identity (e.g. caching token-derived metadata across
    // listener invocations) silently failed. ESLint v10 attaches the
    // array once on the AST root; we match by caching after the first
    // get and serving the same reference thereafter.
    let cached: ReturnType<typeof sharedSourceCode.getTokens> | undefined;
    Object.defineProperty(ast as object, 'tokens', {
      get() {
        if (cached === undefined) cached = sharedSourceCode.getTokens(ast);
        return cached;
      },
      configurable: true,
      enumerable: false,
    });
  }
  // Mirror `comments` onto the AST root for the same reason as
  // `tokens` above. `eslint-plugin-jsdoc`'s comment walker,
  // `eslint-plugin-comment-length`, and several `eslint-plugin-import`
  // top-level scanners read `program.comments` directly (rather than
  // going through `sourceCode.getAllComments()`). Without this getter
  // they get `undefined` and `.length` / `.forEach` throw a TypeError.
  if (!('comments' in (ast as object))) {
    // Memoize for the same `===` identity reasons as `tokens` above.
    // `getAllComments()` returns `_comments.slice()` (new array every
    // call) under the hood, so identity-sensitive consumers would see
    // a different array every read pre-fix.
    let cached: ReturnType<typeof sharedSourceCode.getAllComments> | undefined;
    Object.defineProperty(ast as object, 'comments', {
      get() {
        if (cached === undefined) cached = sharedSourceCode.getAllComments();
        return cached;
      },
      configurable: true,
      enumerable: false,
    });
  }

  const ruleContexts: Array<{ name: string; ctx: RuleContext }> = [];
  const taggedListenerMaps: Array<{
    ruleName: string;
    listeners: Record<string, ListenerFn | ListenerFn[]>;
  }> = [];
  const ruleErrors: Array<{ rule: string; message: string }> = [];

  for (const [ruleName, cfg] of Object.entries(req.rules)) {
    const ruleDef = resolveRule(ruleName, loadedPlugins);
    if (ruleDef == null) {
      // Rule reached us but the plugin doesn't expose it — record and skip.
      ruleErrors.push({
        rule: ruleName,
        message: `rule not found in any loaded plugin`,
      });
      continue;
    }
    const ruleAny = ruleDef as {
      meta?: RuleConfig['meta'] & { defaultOptions?: readonly unknown[] };
      create?: (ctx: RuleContext) => Record<string, ListenerFn | ListenerFn[]>;
    };

    const ctx = createRuleContext({
      ruleName,
      filePath: req.filePath,
      userOptions: cfg.options ?? [],
      schema: ruleAny.meta?.schema as never,
      defaultOptions: ruleAny.meta?.defaultOptions,
      settings: req.settings ?? {},
      // Forward `languageOptions` as-is. The wire shape mirrors
      // ESLint v10 exactly — `ecmaVersion` / `sourceType` / `globals`
      // at the top level, parser-specific `ecmaFeatures` nested under
      // `parserOptions` — so plugin rules reading
      // `ctx.languageOptions.ecmaVersion` or
      // `ctx.languageOptions.parserOptions.ecmaFeatures.jsx` see the
      // exact same paths they would in ESLint itself (FileContext in
      // ESLint just stores the reference unchanged).
      languageOptions: req.languageOptions,
      text: sourceText,
      lsoCache: lso,
      ast,
      scopeManagerFactory,
      sourceCode: sharedSourceCode,
      messages: ruleAny.meta?.messages,
      collectFixes: req.collectFixes,
      suggestionsMode: req.suggestionsMode,
    });

    let returnedListeners: Record<string, ListenerFn | ListenerFn[]> = {};
    try {
      // ruleAny is `any` so the RHS is `any` and assigns to the typed
      // target without a cast.
      returnedListeners = ruleAny.create?.(ctx) ?? {};
    } catch (err) {
      ruleErrors.push({
        rule: ruleName,
        message: `create: ${(err as Error)?.message ?? String(err)}`,
      });
      continue;
    }

    // Standard ESLint plugin API: `rule.create(ctx)` returns the listener
    // map. We no longer expose `ctx.on` / `ctx.onExit` — those were a
    // rslint-only extension that wasn't part of ESLint's public surface
    // and risked overriding standard returned listeners via spread
    // merge. Plugins write straight ESLint and we use that as-is.
    //
    // Listener wrapping previously existed to feed `ctx.getAncestors()`
    // its "current node" register (the v8 no-arg form). ESLint v9
    // removed `context.getAncestors()` outright — plugins should use
    // `sourceCode.getAncestors(node)` instead — so the wrap is gone
    // and listeners flow straight to the merger.
    ruleContexts.push({ name: ruleName, ctx });
    taggedListenerMaps.push({ ruleName, listeners: returnedListeners });
  }

  // ── 5. merge + visit ──────────────────────────────────────────────
  const mergedListeners = mergeListeners(taggedListenerMaps);
  // #3: surface selector-compile failures (an invalid esquery selector in a
  // plugin rule) as per-rule errors. `mergeListeners` skipped the bad
  // selectors and recorded them rather than throwing out of `lintFile`.
  for (const se of mergedListeners.selectorErrors) {
    if (se.ruleName) {
      ruleErrors.push({
        rule: se.ruleName,
        message: `invalid selector '${se.selector}': ${se.message}`,
      });
    }
  }
  const visitResult = visit(ast, mergedListeners, {
    cancelFlag: req.cancelFlag,
    onListenerError: ({ ruleName, selector, node, err }) => {
      // Structured rule-level attribution: write one ruleErrors entry
      // per (rule, listener-throw). This is what surfaces in the
      // per-file result back to Go and ultimately to the user's
      // diagnostic stream — without it the user only sees stderr,
      // which the LSP path doesn't even render.
      //
      // Stderr is still emitted as a dedup'd one-liner so a flood
      // of identical (selector, errMsg) pairs doesn't drown the
      // terminal during interactive runs. But ruleErrors is the
      // authoritative channel: every throw lands there.
      const message = (err as Error)?.message ?? String(err);
      if (ruleName) {
        ruleErrors.push({
          rule: ruleName,
          message: `listener for '${selector}' threw on ${node.type}: ${message}`,
        });
      }
      const key = `${ruleName ?? '?'}::${selector}::${message.slice(0, 100)}`;
      if (!_listenerErrorDedup.has(key)) {
        // Capped insertion. Without the cap, a long-lived worker in
        // an LSP session that lints many files with rotating errors
        // accumulates the Set monotonically; entries are useless old
        // (selector, errMsg) pairs that no longer reflect current
        // user code. The cap discards a single oldest entry when full
        // — FIFO via insertion order on Set (Set iterators yield
        // insertion order per the spec) — which is fine since
        // dedup is purely a stderr-spam guard, not a correctness
        // contract.
        if (_listenerErrorDedup.size >= LISTENER_ERROR_DEDUP_MAX) {
          // `.values().next().value` on a Set<string> is already typed
          // as `string | undefined`; no cast needed.
          const oldest = _listenerErrorDedup.values().next().value;
          if (oldest !== undefined) _listenerErrorDedup.delete(oldest);
        }
        _listenerErrorDedup.add(key);
        // Plugin `console.*` is intentionally NOT monkey-patched in
        // the worker (see `lint-worker.ts`). The host spawns the
        // worker with `{ stdout: true, stderr: true }` and forwards
        // both streams to `onLog`, so a plain `console.error` here
        // reaches the host's log channel under both CLI and LSP
        // without altering plugin-visible console behavior.
        console.error(
          `[ecma-language-plugin] ${ruleName ?? 'anonymous'} listener for ${selector} on ${node.type} threw: ${message}`,
        );
      }
    },
  });
  result.cancelled = visitResult.cancelled;

  // ── 6. drain diagnostics ─────────────────────────────────────────
  //
  // Convert every offset from runner-internal UTF-16 code-unit indices
  // to UTF-8 byte offsets. Two units of measurement collide at the
  // wire boundary:
  //
  //   - oxc-parser emits offsets in UTF-16 code units (matching JS
  //     string indexing). Token API, sourceCode.getText, fixer ranges
  //     — every piece of runner-internal code naturally operates in
  //     this unit.
  //   - Go's `scanner.GetECMALineAndUTF16CharacterOfPosition`, which
  //     turns the wire `startPos` into a final line/column for user
  //     display, takes UTF-8 byte offsets as input (empirically
  //     verified — see `TestScannerPosUnit_UTF8Bytes` in
  //     `internal/linter`).
  //
  // Without this conversion, files with any multi-byte UTF-8 char
  // (CJK text, emoji, arrow glyphs like `➜`) produce diagnostics
  // whose column is shifted back by (bytes_consumed - utf16_units)
  // for every char that appears before the diagnostic in the file.
  // The shift is invisible on pure-ASCII codebases — which is exactly
  // why the bug went unnoticed until a real i18n-heavy project was
  // linted.
  // Stage diagnostics in UTF-16 units first (the unit that comments,
  // tokens, and rule report sites all share). The disable-directive
  // filter compares diagnostic.startPos to comment range offsets — both
  // must be in the same unit. We delay the UTF-16→UTF-8 conversion (the
  // wire boundary) until AFTER filtering so we don't have to also
  // re-shift the directive offsets.
  const staged: Diagnostic[] = [];
  for (const { ctx } of ruleContexts) {
    const diags = ctx._drainDiagnostics();
    for (const d of diags) staged.push(d);
  }

  // ── 7. apply disable directives (UTF-16 units) ──────────────────
  //
  // Pulls comments via the first rule's SourceCode — every ruleContext
  // shares the same source text so any of them works. The lazy tokenize
  // inside SourceCode is cached, so this triggers at most one extra
  // tokenize per file (zero if a rule already called getAllComments /
  // similar during the visit pass).
  //
  // Skip entirely when no rule contexts exist: zero rules → zero
  // diagnostics → nothing to filter, and there's no SourceCode to pull
  // comments from anyway.
  const filteredDiagnostics =
    ruleContexts.length === 0
      ? staged
      : applyDisableDirectives({
          comments: ruleContexts[0].ctx.sourceCode.getAllComments(),
          diagnostics: staged,
          lineStartOffsets: lso,
          textLength: sourceText.length,
        });

  // ── 8. UTF-16 → UTF-8 conversion + wire push ────────────────────
  // Building the per-character UTF-16→UTF-8 map is O(N) over the
  // source text. When the file produced no diagnostics (empirically
  // ~60% of files on a clean codebase), every byte of that map is
  // unused. Gate the build on having something to convert.
  if (filteredDiagnostics.length > 0) {
    const u16ToByte = buildUtf16ToByteMap(sourceText);
    // `u16ToByte` indexes from 0 to `sourceText.length` inclusive (the
    // sentinel at `[length]` is the total byte count for one-past-end
    // `endPos`). Clamp every offset before lookup so a plugin that
    // reports `{ loc: { line: 9999, column: 5 } }` (or fixes that
    // reference a stale node range) gets a deterministic in-range byte
    // offset instead of `u16ToByte[undefined] === undefined` leaking
    // onto the wire and confusing Go's decoder.
    const clamp = (pos: number): number => {
      if (!Number.isFinite(pos) || pos < 0) return 0;
      if (pos > sourceText.length) return sourceText.length;
      // Truncate to an integer before indexing into `u16ToByte`.
      // Pre-fix a finite decimal (e.g. 2.5) passed through unchanged
      // and `u16ToByte[2.5] === undefined` leaked onto the wire,
      // contradicting the "no undefined past this point" contract the
      // wider comment block claims. `Math.trunc` is the right choice
      // (vs floor/ceil) so negatives can't surprise us — `pos < 0`
      // is already short-circuited above.
      return Math.trunc(pos);
    };
    const toByte = (pos: number): number => u16ToByte[clamp(pos)];
    for (const d of filteredDiagnostics) {
      d.startPos = toByte(d.startPos);
      d.endPos = toByte(d.endPos);
      if (d.fixes && d.fixes.length > 0) {
        for (const f of d.fixes) {
          f.range = [toByte(f.range[0]), toByte(f.range[1])];
          result.fixes.push(f);
        }
      }
      if (d.suggestions && d.suggestions.length > 0) {
        for (const s of d.suggestions) {
          if (s.fixes) {
            for (const f of s.fixes) {
              f.range = [toByte(f.range[0]), toByte(f.range[1])];
            }
          }
        }
        result.suggestionsCount += d.suggestions.length;
      }
      result.diagnostics.push(d);
    }
  }

  if (ruleErrors.length > 0) result.ruleErrors = ruleErrors;
  return result;
}

// Per-Worker dedup: prevents the same (selector, errorMessage) from
// flooding stderr if many nodes hit the same buggy listener.
// Capped dedup Set: at most LISTENER_ERROR_DEDUP_MAX (selector, errMsg)
// pairs are remembered to prevent stderr flooding. When full, the
// oldest entry is evicted (insertion-order via Set iteration).
//
// Cap = 1024 is well above the realistic worst case (each rule has
// maybe 5-10 distinct selectors × maybe 10 distinct error messages per
// rule = ~100 unique keys per rule). 1024 covers ~10 buggy rules
// simultaneously without forgetting recent failures, and bounds memory
// to a few hundred KB max even under pathological growth.
const LISTENER_ERROR_DEDUP_MAX = 1024;
const _listenerErrorDedup = new Set<string>();

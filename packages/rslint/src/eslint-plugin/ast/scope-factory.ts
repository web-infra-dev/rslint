/* rslint-disable @typescript-eslint/no-unsafe-type-assertion */
/**
 * Scope-manager factory: pick the right scope analyzer for each file.
 *
 * Two implementations are bundled as runtime deps:
 *
 *   - `eslint-scope`                            for plain JS files
 *   - `@typescript-eslint/scope-manager`        for TypeScript files
 *
 * Both expose a compatible-enough surface for ESLint plugins: a top-level
 * `ScopeManager` with `acquire(node)`, `globalScope`, etc. The TS version
 * understands TS-specific binding rules (enum members, namespace exports,
 * type-only imports), which the plain JS analyzer does not — using
 * eslint-scope on `.ts` files would silently miscount `Variable.references`
 * for the TS-only constructs.
 *
 * Selection is by file extension; languageOptions can override globals
 * (forwarded per-file in the wire request's `languageOptions`, which Go
 * computes via `GetConfigForFile`).
 */

import type { ScopeManager as EslintScopeManager } from 'eslint-scope';
// CJS-only scope analyzers, imported statically so rslib bundles them into the
// worker (consumers of @rslint/core don't need them at runtime). They run in
// the synchronous scope-analysis path, so a dynamic `import()` (async) isn't
// usable here. BOTH must use the namespace form (`import * as`): under rslib's
// CJS interop a DEFAULT import of these bundles to `undefined`, which crashed
// scope analysis on every .ts/.tsx file ("Cannot read properties of undefined
// (reading 'analyze')") while .js (eslint-scope) silently kept working.
import * as tsScopeManagerPkg from '@typescript-eslint/scope-manager';
import * as eslintScopePkg from 'eslint-scope';

import { VISITOR_KEYS } from './visitor-keys.js';
import { BUILTIN_GLOBAL_NAMES } from './builtin-globals.js';
import type { GlobalAccess } from '../types.js';

type NormalizedGlobalAccess = 'readonly' | 'writable' | 'off';

/** Subset of options accepted by both analyzers. */
export interface ScopeFactoryOptions {
  filePath: string;
  /** Optional overrides; sane defaults are used when missing. */
  sourceType?: 'module' | 'script' | 'commonjs';
  ecmaVersion?: number | 'latest';
  /**
   * `languageOptions.globals` from the user's flat-config — a map of
   * global names to ESLint-compatible access values. Used to seed the
   * scope-manager so user-declared globals don't false-positive `no-undef`
   * and similar rules.
   */
  globals?: Record<string, GlobalAccess>;
  /**
   * `languageOptions.parserOptions.ecmaFeatures.impliedStrict`. ESLint
   * v10 default is `false`; setting it `true` treats the source as if
   * it begins with `'use strict'`. Strict mode reserves additional
   * keywords (`yield` / `let` / `implements` ...) which affects scope
   * analysis. Without honouring the user setting, sourceType:'script'
   * legacy bundles get analysed as strict and rules like
   * `no-shadow-restricted-names` false-positive.
   */
  impliedStrict?: boolean;
  /**
   * `languageOptions.parserOptions.ecmaFeatures.globalReturn`. When
   * true (typical for Node CommonJS scripts that may contain top-level
   * `return`), eslint-scope wraps the Program in an implicit Function
   * scope, matching the runtime `(function(exports, require, module,
   * __filename, __dirname) { ... })` wrapper Node applies. Forwarded as
   * `nodejsScope` to eslint-scope.
   */
  globalReturn?: boolean;
  /**
   * `languageOptions.parserOptions.ecmaFeatures.jsxPragma`. The TS
   * scope-manager treats `<X />` as a reference to this identifier so
   * `no-unused-vars` on the JSX-pragma import doesn't false-positive.
   * Defaults to `'React'` to match the historical behaviour, but
   * Preact / Vue JSX configurations override it.
   */
  jsxPragma?: string;
  /**
   * `languageOptions.parserOptions.ecmaFeatures.jsxFragmentName`. The
   * TS scope-manager treats `<>...</>` as a reference to this
   * identifier (defaults to `'Fragment'`). Preact users set this to
   * `'Fragment'` already; some bundlers / TS configs use a custom
   * fragment factory.
   */
  jsxFragmentName?: string;
}

const TS_EXT = new Set(['.ts', '.tsx', '.mts', '.cts']);

/**
 * Returns true when the file should be analyzed by the TS scope-manager.
 * Uses the literal extension; we intentionally don't sniff content to
 * keep this O(1) — if a file is named `.js` but contains TS syntax, that's
 * a parser problem, not a scope-manager-selection problem.
 */
export function shouldUseTsScopeManager(filePath: string): boolean {
  const dot = filePath.lastIndexOf('.');
  if (dot < 0) return false;
  return TS_EXT.has(filePath.slice(dot));
}

/**
 * Produce a thunk that, when called, runs scope analysis on `ast` and
 * returns the ScopeManager. Lazy by design: rules that don't call
 * `getScope()` never trigger the analysis pass. (The analyzers are top-level
 * static imports so they bundle into the worker; "lazy" here defers the
 * analysis work, not the module load.)
 */
export function makeScopeManagerFactory(
  ast: object,
  opts: ScopeFactoryOptions,
): () => unknown {
  let cached: unknown = null;
  let computed = false;
  return () => {
    if (computed) return cached;
    computed = true;
    cached = analyze(ast, opts);
    return cached;
  };
}

function analyze(ast: object, opts: ScopeFactoryOptions): unknown {
  const useTs = shouldUseTsScopeManager(opts.filePath);
  if (useTs) {
    const tsScope = tsScopeManagerPkg as unknown as {
      analyze: (ast: object, options?: object) => unknown;
    };
    return tsScope.analyze(ast, {
      sourceType: opts.sourceType ?? 'module',
      // eslint-scope (the JS-side analyzer) auto-enables nodejsScope
      // when `sourceType === 'commonjs'` — see its `isImpliedStrict()`
      // which returns `nodejsScope || sourceType === 'commonjs'`.
      // ts-scope-manager has no such equivalence, so the same source
      // got `[global]` under .cts vs `[global, function]` under .cjs
      // (function wrapper). Route commonjs ⇒ globalReturn:true
      // ourselves here to keep TS and JS commonjs paths on the same
      // scope tree. Also forwards an explicit user-supplied
      // `globalReturn` — pre-fix the TS branch silently dropped it
      // entirely (the JS branch line ~169 honored it), breaking
      // sourceType:'script' + globalReturn:true on .ts/.tsx files.
      globalReturn:
        opts.globalReturn === true || opts.sourceType === 'commonjs',
      // ESLint v10 default is `false`. The previous hard-coded `true`
      // forced every sourceType:'script' file to be analysed as
      // strict, producing false-positives for `no-shadow-restricted-
      // names` and similar in legacy bundles.
      impliedStrict: opts.impliedStrict ?? false,
      // Defaults preserved for back-compat; Preact / Vue JSX users
      // override via languageOptions.parserOptions.ecmaFeatures.
      jsxPragma: opts.jsxPragma ?? 'React',
      jsxFragmentName: opts.jsxFragmentName ?? 'Fragment',
    });
  }

  const eslintScope = eslintScopePkg as unknown as {
    analyze: (ast: object, options?: object) => EslintScopeManager;
  };
  // `childVisitorKeys` MUST be supplied. Without it eslint-scope's
  // `fallback: 'iteration'` default blindly iterates every enumerable
  // property — including the `parent` backref and `loc`/`range` — and
  // on the FIRST JSX element walks `parent` upward then re-descends,
  // overflowing the stack ("Maximum call stack size exceeded");
  // empirically every `.jsx` file silently failed scope analysis.
  // We pass the runner's `VISITOR_KEYS` table (the same table the walker
  // traverses with — see `visitor-keys.ts`): it covers JSX/TS and is a
  // strict superset of ESLint's keys, so eslint-scope descends exactly
  // the right children. Verified equivalent to the former
  // `eslint-visitor-keys` `KEYS` on JS / JSX / decorator scope trees
  // before the switch (identical scope trees).
  return eslintScope.analyze(ast, {
    ecmaVersion:
      opts.ecmaVersion === 'latest' ? 2025 : (opts.ecmaVersion ?? 2025),
    sourceType: opts.sourceType ?? 'module',
    // ESLint v10 default `false` — see jsdoc on ScopeFactoryOptions.
    impliedStrict: opts.impliedStrict ?? false,
    // `nodejsScope` wraps Program in an implicit Function scope when
    // the user set `parserOptions.ecmaFeatures.globalReturn: true` —
    // matches Node CJS runtime, where the file is wrapped in
    // `(function(exports, require, module, __filename, __dirname) {})`.
    // Required so top-level `return` / `arguments` / `this` resolve to
    // the Function scope rather than crashing scope analysis or
    // shifting `variableScope.type` for rules that check it.
    nodejsScope: opts.globalReturn === true,
    fallback: 'iteration',
    childVisitorKeys: VISITOR_KEYS,
    // `opts.globals` is intentionally NOT passed here — neither eslint-scope
    // nor the TS analyzer takes globals at construct time, so both paths seed
    // them post-analysis via seedGlobals below.
  });
}

// Minimal shape of an eslint-scope Variable that the
// post-analyze resolution loop and downstream plugin rules read.
interface SyntheticGlobal {
  name: string;
  defs: unknown[];
  identifiers: unknown[];
  references: Array<{ identifier?: { name?: string }; resolved?: unknown }>;
  writeable: boolean;
  eslintImplicitGlobalSetting?: 'readonly' | 'writable';
  scope: unknown;
}

const syntheticGlobals = new WeakSet<object>();

interface GlobalScopeLike {
  variables: SyntheticGlobal[];
  set?: Map<string, SyntheticGlobal>;
  through?: Array<{ identifier?: { name?: string }; resolved?: unknown }>;
}

/**
 * Add or update a synthetic global Variable in `globalScope`. Returns
 * the (newly created or existing) Variable.
 *
 * Mode is synced to `mode` for existing synthetic entries. Source-declared
 * Variables are left unchanged. ESLint's
 * apply-environments + apply-globals pipeline lets user globals
 * (`languageOptions.globals: { Array: 'writable' }`) override the
 * built-in mode flags layered earlier by `seedEcmaGlobals`. Pre-fix
 * `ensureGlobal` short-circuited on existing entries, so user-supplied
 * `'writable'` on a built-in like `Array` had no effect — rules such
 * as `no-global-assign` would then false-positive when the user
 * reassigned the global they had explicitly opted in to.
 *
 * Idempotent: calling twice with the same name + mode is a no-op
 * after the first call.
 */
function ensureGlobal(
  gs: GlobalScopeLike,
  name: string,
  mode: 'readonly' | 'writable',
): SyntheticGlobal {
  const existing = gs.set?.get(name);
  if (existing) {
    // A source declaration can share a name with a configured/built-in global.
    // Config globals must not rewrite that lexical Variable's semantics.
    if (!syntheticGlobals.has(existing)) return existing;
    // Sync mode — see jsdoc. The Variable's identity stays; only the
    // mode flags shift. Downstream rules consult `writeable` (the
    // typo'd-as-eslint-compat field) and `eslintImplicitGlobalSetting`
    // (the modern name); both must move in lockstep.
    existing.writeable = mode === 'writable';
    existing.eslintImplicitGlobalSetting = mode;
    return existing;
  }
  const v: SyntheticGlobal = {
    name,
    defs: [],
    identifiers: [],
    references: [],
    writeable: mode === 'writable',
    eslintImplicitGlobalSetting: mode,
    scope: gs,
  };
  syntheticGlobals.add(v);
  gs.variables.push(v);
  gs.set?.set(name, v);
  return v;
}

/**
 * Walk `globalScope.through` (unresolved free references) and attach
 * each one whose name matches a global Variable to that variable's
 * `references` list, clearing it from `through`. This is what ESLint's
 * `lib/linter/apply-environments.js` does post-analyze — eslint-scope
 * itself never resolves cross-scope references against env-supplied
 * globals because it doesn't know about envs. Without this pass,
 * `globalScope.through` keeps every reference to a built-in (parseInt,
 * Array, etc.) as unresolved, and downstream tools like
 * `@eslint-community/eslint-utils`'s `ReferenceTracker` — which
 * iterates `variable.references` — see zero references and silently
 * skip every globally-tracked rule (unicorn's `prefer-number-properties`
 * et al). Empirically pinned via the combined-rules conformance suite.
 */
function resolveThroughReferences(gs: GlobalScopeLike): void {
  const through = gs.through;
  if (!through || through.length === 0) return;
  const remaining: typeof through = [];
  for (const ref of through) {
    const name = ref.identifier?.name;
    if (name == null) {
      remaining.push(ref);
      continue;
    }
    const v = gs.set?.get(name);
    if (v) {
      v.references.push(ref);
      ref.resolved = v;
    } else {
      remaining.push(ref);
    }
  }
  through.length = 0;
  for (const r of remaining) through.push(r);
}

/**
 * Seed ECMA built-in globals (`globals.builtin` from the `globals` npm
 * package: parseInt, NaN, Infinity, Array, Object, ...) into the
 * global scope as `readonly` Variables. ESLint's flat-config does this
 * by default before any user-supplied `languageOptions.globals` are
 * applied; rslint mirrors that here so plugin rules that walk the
 * global scope (e.g. via `ReferenceTracker.iterateGlobalReferences`)
 * see the same variable set on both engines.
 *
 * Idempotent: calling twice doesn't duplicate entries.
 */
export function seedEcmaGlobals(scopeManager: unknown): void {
  if (!scopeManager) return;
  const sm = scopeManager as { globalScope?: GlobalScopeLike };
  const gs = sm.globalScope;
  if (!gs) return;

  for (const name of BUILTIN_GLOBAL_NAMES) {
    ensureGlobal(gs, name, 'readonly');
  }
  resolveThroughReferences(gs);
}

/**
 * Seed user-supplied globals from `languageOptions.globals` into the
 * scope manager's global scope. Layered on top of `seedEcmaGlobals`
 * so user-declared globals override the built-ins' mode flags (e.g.
 * a project declaring `Array: 'writable'` lifts the readonly default).
 *
 * Both `eslint-scope` and `@typescript-eslint/scope-manager` expose
 * `globalScope.variables` and `globalScope.set` (the latter is a Map
 * of name → Variable).
 *
 * Typically called by the SourceCode builder right after the factory's
 * analyze() returns.
 */
export function seedGlobals(
  scopeManager: unknown,
  globals: Record<string, GlobalAccess> | undefined,
): void {
  if (!scopeManager) return;
  const sm = scopeManager as { globalScope?: GlobalScopeLike };
  const gs = sm.globalScope;
  if (!gs) return;
  if (!globals) {
    // Even with no user globals, we still want the resolution pass
    // (in case `seedEcmaGlobals` already ran but new `through` refs
    // appeared, or the caller chose to skip the ECMA seed).
    resolveThroughReferences(gs);
    return;
  }

  for (const [name, access] of Object.entries(globals)) {
    const mode = normalizeGlobalAccess(access);
    if (mode == null) continue;
    if (mode === 'off') {
      // 'off' explicitly removes the binding — drop any synthetic entry.
      const v = gs.set?.get(name);
      if (v && syntheticGlobals.has(v)) {
        // Restore any references that `seedEcmaGlobals` previously
        // moved from `gs.through` onto this Variable. Without this,
        // every ref is left dangling on a deleted Variable —
        // `Variable.references[]` still points at it, `ref.resolved`
        // still references it, but the Variable itself is no longer
        // in any scope. Plugin rules that walk
        // `globalScope.variables` to find name='Array' see nothing
        // AND plugins that walk `globalScope.through` for unresolved
        // refs ALSO see nothing — `Array.from(...)` becomes invisible
        // to every reference-tracking rule (`unicorn/prefer-number-
        // properties`, `@typescript-eslint/no-unnecessary-type-
        // assertion`, etc.). The semantically correct behavior for
        // `{ Array: 'off' }` is "pretend Array isn't a known
        // identifier" — i.e. its refs become unresolved-throughs,
        // exactly the state they'd be in if seedEcmaGlobals had
        // never added the Variable to begin with.
        if (Array.isArray(v.references) && v.references.length > 0) {
          const through = gs.through ?? (gs.through = []);
          for (const ref of v.references) {
            (ref as { resolved?: unknown }).resolved = null;
            through.push(ref);
          }
          v.references = [];
        }
        const idx = gs.variables.indexOf(v);
        if (idx >= 0) gs.variables.splice(idx, 1);
        gs.set?.delete(name);
      }
      continue;
    }
    ensureGlobal(gs, name, mode);
  }
  resolveThroughReferences(gs);
}

function normalizeGlobalAccess(
  access: GlobalAccess,
): NormalizedGlobalAccess | undefined {
  switch (access) {
    case true:
    case 'true':
    case 'writable':
    case 'writeable':
      return 'writable';
    case false:
    case null:
    case 'false':
    case 'readonly':
    case 'readable':
      return 'readonly';
    case 'off':
      return 'off';
    default:
      // Config validation rejects unknown values before plugin execution.
      // Keep the scope boundary defensive so a runtime value that bypassed
      // validation cannot become an invalid eslintImplicitGlobalSetting.
      return undefined;
  }
}

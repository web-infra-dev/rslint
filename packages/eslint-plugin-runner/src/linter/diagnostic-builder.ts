/* eslint-disable @typescript-eslint/no-unsafe-type-assertion -- AST / parser / scope-manager / plugin-API boundary casts. Each site projects from an `any` / `unknown` peer surface (oxc-parser output, user plugin objects, ESLint v10 wire shapes) into the typed shape this module uses; the contract is runtime-validated at the call boundaries above, not at the cast. Bulk-disabling here instead of per-line keeps the cast sites readable. */
/**
 * Diagnostic construction for `context.report(...)` calls.
 *
 * Split out from `context.ts` because the build path is largely
 * self-contained: it normalizes the four argument shapes ESLint
 * accepts, resolves messageId templates, gates fix() / suggest[i].fix()
 * by the runner's `collectFixes` / `suggestionsMode` knobs, and
 * returns a wire-shaped {@link Diagnostic} (or `null` for a malformed
 * report). `context.ts` stays focused on RuleContext wiring.
 */

import { type Fix, type RuleFixer } from './fixer.js';
import { type LocPosition } from '../ast/normalize-ast.js';
import {
  buildLineStartOffsetsLocal,
  type ESTreeNode,
} from '../source-code/source-code.js';
import { lineStartOffset } from '../source-code/source-code-helpers.js';

// ESLint v10 `lib/linter/interpolate.js` uses this exact regex (lazy
// match between `{{ }}` pairs, ECMAScript Unicode mode) plus a function
// replacer + `term.trim()`. Two reasons we can't just use
// `String.prototype.replaceAll('{{key}}', v)`:
//   1. literal `{{key}}` won't match `{{ key }}` (space-padded templates
//      like `{{ propName }}` are common in react-*/jsdoc rules);
//   2. the string-form `replaceAll(needle, replacement)` interprets
//      `$&`, `$$`, `` $` ``, `$'` in `replacement` as match-group
//      back-references — values like `$&`, `$$`, jQuery `$(...)` or
//      price strings get silently mangled. The function form bypasses
//      back-reference interpretation entirely.
const INTERPOLATE_RE = /\{\{([^{}]+?)\}\}/gu;

function interpolateMessage(
  template: string,
  data: Record<string, string | number>,
): string {
  return template.replace(INTERPOLATE_RE, (match, term: string) => {
    const key = term.trim();
    // ESLint v10 `lib/linter/interpolate.js` uses `key in data`
    // (prototype chain included), not an own-key check — a placeholder
    // resolves if the key is reachable on `data` at all. We match v10
    // verbatim so a `{{ propName }}` template interpolates identically
    // under rslint and ESLint.
    return key in data ? String(data[key]) : match;
  });
}

// ─────────────────────────────────────────────────────────────────────
// Public types
// ─────────────────────────────────────────────────────────────────────

/**
 * Suggestion descriptor as returned to Go. `fixes` is null in `'off'`
 * mode (we record the descriptor but didn't run `fix(fixer)`); a
 * populated array in `'eager'` mode.
 */
export interface SuggestionDescriptor {
  messageId?: string;
  desc?: string;
  fixes: Fix[] | null;
}

/** A single diagnostic emitted by a rule call to `context.report`. */
export interface Diagnostic {
  ruleName: string;
  messageId?: string;
  message: string;
  /**
   * Position offsets into the source text. INSIDE the runner these
   * are UTF-16 code-unit indices (matching native JS string
   * indexing). On the IPC boundary in `ecma-language-plugin.ts` they
   * are converted to UTF-8 BYTE offsets before shipping to Go — Go's
   * `scanner.GetECMALineAndUTF16CharacterOfPosition` takes bytes.
   * Callers that drain diagnostics via the public lintFile path
   * therefore observe byte offsets; the in-process runner code paths
   * still work in UTF-16 units.
   */
  startPos: number;
  endPos: number;
  fixes?: Fix[];
  suggestions?: SuggestionDescriptor[];
}

/** Mode controlling whether suggestion `fix(fixer)` is invoked at report time. */
export type SuggestionsMode = 'off' | 'eager';

/** Optional message-id mapping from `rule.meta.messages`. */
export type MessagesMap = Record<string, string>;

/**
 * Modern descriptor form: `context.report({ node | loc, message |
 * messageId, data?, fix?, suggest? })`. This is the recommended form.
 *
 * For the legacy positional form ESLint still accepts —
 * `report(node, message, data?, fix?)` or
 * `report(node, loc, message, data?, fix?)` — see
 * `normalizeReportArgs` below. ESLint's
 * `lib/linter/file-report.js`'s `normalizeMultiArgReportCall` is our
 * reference implementation.
 */
/**
 * `loc` shape accepted on a report descriptor. Wider than the AST's
 * own {@link SourceLocation} on purpose: ESLint's `context.report({ loc })`
 * permits any of:
 *
 *   - `{ line, column }`           — a single position (treated as zero-width)
 *   - `{ start, end }`             — a full range
 *   - `{ start }`                  — partial range; `end` defaults to `start`
 *
 * The third form is what real plugins write when reporting at a single
 * point but using the `{start: ...}` shape they already build for ranges
 * elsewhere. ESLint v9's `lib/linter/file-report.js` accepts it; rslint
 * matches.
 */
export type ReportLoc = LocPosition | { start: LocPosition; end?: LocPosition };

export interface ReportDescriptor {
  node?: ESTreeNode;
  loc?: ReportLoc;
  message?: string;
  messageId?: string;
  data?: Record<string, string | number>;
  fix?: (fixer: RuleFixer) => Fix | Fix[] | null | undefined | Iterable<Fix>;
  suggest?: SuggestionInput[];
}

export interface SuggestionInput {
  messageId?: string;
  desc?: string;
  data?: Record<string, string | number>;
  fix: (fixer: RuleFixer) => Fix | Fix[] | null | undefined | Iterable<Fix>;
}

// ─────────────────────────────────────────────────────────────────────
// Internal types
// ─────────────────────────────────────────────────────────────────────

export interface BuildDiagnosticArgs {
  ruleName: string;
  descriptor: ReportDescriptor;
  text: string;
  messages: MessagesMap;
  fixer: RuleFixer;
  collectFixes: boolean;
  suggestionsMode: SuggestionsMode;
  /**
   * Per-file line-start-offsets cache (1 array per file). When supplied,
   * the `descriptor.loc` path reuses this instead of rebuilding the
   * lso from `text` on every report. This is critical for rules like
   * `@typescript-eslint/no-unused-vars` which include `loc:` on every
   * descriptor — rebuilding lso per report was ~96µs each (file-text
   * scan) × thousands of reports per run.
   *
   * Optional only because legacy unit tests construct a fresh
   * `BuildDiagnosticArgs` without wiring SourceCode; the production
   * `createRuleContext` always passes it.
   */
  lsoCache?: number[];
}

// ─────────────────────────────────────────────────────────────────────
// Diagnostic builder — handles fix / suggest gating
// ─────────────────────────────────────────────────────────────────────

export function buildDiagnostic(args: BuildDiagnosticArgs): Diagnostic | null {
  const {
    ruleName,
    descriptor,
    text,
    messages,
    fixer,
    collectFixes,
    suggestionsMode,
    lsoCache,
  } = args;

  let startPos: number;
  let endPos: number;
  // ESLint v10's `lib/linter/file-report.js` prioritises `loc` over
  // `node` when both are present — the descriptor is interpreted as
  // "use the same node for AST attribution but report at this
  // specific location" (commonly used to point at a specific token
  // inside a larger node). The previous order branched on `node`
  // first, silently ignoring an explicit `loc` and reporting at the
  // node's range instead. Empirically pinned against ESLint v10.
  if (descriptor.loc) {
    // Reuse the per-file lso cache when wired up by createRuleContext.
    // Falls back to building a transient one only for legacy callers
    // that didn't supply `lsoCache` (unit tests, direct API consumers).
    const lso = lsoCache ?? buildLineStartOffsetsLocal(text);
    if ('start' in descriptor.loc) {
      const startLoc = descriptor.loc.start;
      // `end` is optional on real plugin reports — ESLint v9's
      // `lib/linter/file-report.js` defaults a missing `end` to `start`
      // (treats the report as zero-width at `start`). Without this
      // fallback rslint crashed with `Cannot read properties of
      // undefined (reading 'line')`, dropping the diagnostic entirely
      // for any plugin that passed `{ loc: { start: {...} } }`. The
      // worker's listener wrapper catches the throw and routes it to
      // `ruleErrors`, so the user only saw "rule failed" with no
      // information about the intended report — empirically pinned
      // against ESLint v9.
      const endLoc = descriptor.loc.end ?? startLoc;
      // `lineStartOffset` clamps BOTH ends — a prior `lso[Math.max(0, line - 1)]`
      // only clamped the low end, so reports past EOF (rare but possible
      // for synthesized trailing positions) produced `undefined`/`NaN`
      // diagnostic ranges. Clamping to `text.length` resolves to a safe
      // zero-width position at EOF instead.
      startPos =
        lineStartOffset(lso, startLoc.line ?? 1, text.length) +
        (startLoc.column ?? 0);
      endPos =
        lineStartOffset(lso, endLoc.line ?? 1, text.length) +
        (endLoc.column ?? 0);
    } else {
      // descriptor.loc is a Position (line/column) — treat as zero-width.
      // `'start' in descriptor.loc` ruled out the range form, so
      // descriptor.loc is narrowed to `LocPosition` here; no cast needed.
      //
      // `line ?? 1` + `column ?? 0` mirror the range-form branch above
      // and ESLint v10's `lib/linter/file-report.js` defaults. Plugin
      // code that reports `{ loc: { line: 5 } }` (omitting column) was
      // otherwise computing `lineStartOffset(...) + undefined === NaN`
      // and shipping NaN positions to the wire.
      startPos =
        lineStartOffset(lso, descriptor.loc.line ?? 1, text.length) +
        (descriptor.loc.column ?? 0);
      endPos = startPos;
    }
  } else if (descriptor.node) {
    // Prefer range — that's the common path (oxc-emitted nodes always
    // have it). Fall back to `node.loc` for synthetic / autogenerated
    // nodes (`{ type, loc }` shapes that ESLint v10 also accepts).
    // Pre-fix the bare destructure assumed `.range` exists and threw
    // TypeError on loc-only nodes; the throw was swallowed by the
    // listener try/catch and the diagnostic vanished silently.
    const nodeRange = (descriptor.node as { range?: [number, number] }).range;
    if (nodeRange) {
      [startPos, endPos] = nodeRange;
    } else if (descriptor.node.loc) {
      const lso = lsoCache ?? buildLineStartOffsetsLocal(text);
      const sLoc = descriptor.node.loc.start;
      const eLoc = descriptor.node.loc.end ?? sLoc;
      startPos =
        lineStartOffset(lso, sLoc.line, text.length) + (sLoc.column ?? 0);
      endPos =
        lineStartOffset(lso, eLoc.line, text.length) + (eLoc.column ?? 0);
    } else {
      // A node with neither `.range` nor `.loc` yields no derivable
      // position. ESLint's `assertValidNodeInfo`
      // (lib/linter/file-report.js) demands a usable location and
      // throws this exact string when none is provided; pre-fix the
      // runner silently `return null`ed, so `context.report()` produced
      // no diagnostic AND no `ruleErrors` entry — the report just
      // vanished. Throwing routes it to `ruleErrors` like ESLint.
      throw new TypeError(
        'Node must be provided when reporting error if location is not provided',
      );
    }
  } else {
    // Neither `node` nor `loc` — ESLint's `assertValidNodeInfo`
    // (lib/linter/file-report.js) throws this exact string. Pre-fix the
    // runner silently dropped the report (no diagnostic, no error).
    throw new TypeError(
      'Node must be provided when reporting error if location is not provided',
    );
  }

  // Resolve the message, mirroring ESLint's
  // `computeMessageFromDescriptor` (lib/linter/file-report.js,
  // eslint@9.32.0 — the installed copy). The throw conditions and the
  // exact throw strings are copied verbatim so a plugin that misuses
  // `report()` fails identically under rslint and ESLint:
  //
  //   1. message + messageId BOTH present → throw (ambiguous intent).
  //      Pre-fix the runner silently kept `message`.
  //   2. messageId not in the rule's `messages` map → throw. Pre-fix
  //      the runner fabricated a `(${messageId})` placeholder, which
  //      hid the authoring mistake and shipped a useless diagnostic.
  //
  // These throws propagate out of `context.report()`, which runs INSIDE
  // the per-rule listener (or `create()`) try/catch in
  // `ecma-language-plugin.ts` → `listener-merge.ts` (`onListenerError`),
  // so a buggy rule's misuse lands as a `ruleErrors` entry for that one
  // rule and does NOT abort the file or any sibling rule — matching how
  // ESLint isolates a thrown rule.
  let message: string | undefined;
  if (descriptor.messageId) {
    const id = descriptor.messageId;
    if (descriptor.message) {
      throw new TypeError(
        'context.report() called with a message and a messageId. Please only pass one.',
      );
    }
    // `messages` defaults to `{}` from createRuleContext, so the
    // "no messages were present" branch ESLint has for an undefined
    // `messages` can only matter to direct callers that pass it; guard
    // both ways to stay faithful.
    if (!messages || !Object.hasOwn(messages, id)) {
      throw new TypeError(
        `context.report() called with a messageId of '${id}' which is not present in the 'messages' config: ${JSON.stringify(messages, null, 2)}`,
      );
    }
    message = messages[id];
  } else if (descriptor.message) {
    message = descriptor.message;
  }
  if (descriptor.data && message) {
    message = interpolateMessage(message, descriptor.data);
  }
  if (message == null) {
    // Neither `message` nor `messageId` was supplied. ESLint's
    // `computeMessageFromDescriptor` (lib/linter/file-report.js) throws
    // this exact string for that case. Pre-fix the runner silently
    // `return null`ed — the report vanished with no diagnostic and no
    // `ruleErrors` entry. Throwing routes it to `ruleErrors` like the
    // message+messageId / unknown-messageId throws above.
    throw new TypeError(
      'Missing `message` property in report() call; add a message that describes the linting problem.',
    );
  }

  // ── fix(fixer) — only when the caller asked us to collect fixes. ──
  //
  // collectFixes is true when the consumer (CLI --fix, LSP code-actions)
  // intends to surface the fix payload to the user. `false` skips the
  // descriptor.fix(fixer) call entirely so a plugin whose fix() is
  // expensive doesn't pay the cost on plain lint passes.
  let fixes: Fix[] | undefined;
  if (collectFixes && typeof descriptor.fix === 'function') {
    try {
      const result = descriptor.fix(fixer);
      fixes = materializeFixes(result, text);
    } catch (err) {
      // A buggy fix must not turn a real diagnostic into a runtime error.
      // Drop the fix; keep the diagnostic. `console.error` is forwarded
      // through the worker's captured stderr stream to the host's
      // `onLog` channel (see `lint-worker.ts` — plugin `console.*` is
      // intentionally NOT monkey-patched).
      console.error(
        `[context] rule ${ruleName} fix() threw: ${(err as Error)?.message ?? err}`,
      );
    }
  }

  // ── suggest[i].fix(fixer) — only when suggestionsMode === 'eager' ──
  let suggestions: SuggestionDescriptor[] | undefined;
  if (descriptor.suggest && descriptor.suggest.length > 0) {
    // Validate every suggestion BEFORE building any of them, mirroring
    // ESLint's `validateSuggestions` (lib/linter/file-report.js, run
    // before `mapSuggestions`). The throw strings are copied verbatim so
    // a plugin that misuses `suggest[]` fails identically under rslint
    // and ESLint. Pre-fix the runner did zero validation and fabricated
    // `messages[s.messageId] ?? `(${s.messageId})`` for an unknown
    // messageId — shipping a blank/placeholder suggestion that hid the
    // authoring mistake. These throws propagate out of `context.report()`
    // and land as a `ruleErrors` entry like the message validations above.
    for (const s of descriptor.suggest) {
      if (s.messageId) {
        const { messageId } = s;
        if (!messages) {
          throw new TypeError(
            `context.report() called with a suggest option with a messageId '${messageId}', but no messages were present in the rule metadata.`,
          );
        }
        if (!messages[messageId]) {
          throw new TypeError(
            `context.report() called with a suggest option with a messageId '${messageId}' which is not present in the 'messages' config: ${JSON.stringify(messages, null, 2)}`,
          );
        }
        if (s.desc) {
          throw new TypeError(
            "context.report() called with a suggest option that defines both a 'messageId' and an 'desc'. Please only pass one.",
          );
        }
      } else if (!s.desc) {
        throw new TypeError(
          "context.report() called with a suggest option that doesn't have either a `desc` or `messageId`",
        );
      }
    }
    suggestions = descriptor.suggest.map((s) => {
      let suggestionMessage = s.desc;
      if (!suggestionMessage && s.messageId) {
        suggestionMessage = messages[s.messageId] ?? `(${s.messageId})`;
      }
      if (s.data && suggestionMessage) {
        suggestionMessage = interpolateMessage(suggestionMessage, s.data);
      }
      let sFixes: Fix[] | null = null;
      if (suggestionsMode === 'eager' && typeof s.fix === 'function') {
        try {
          sFixes = materializeFixes(s.fix(fixer), text) ?? null;
        } catch (err) {
          console.error(
            `[context] rule ${ruleName} suggestion fix() threw: ${(err as Error)?.message ?? err}`,
          );
          sFixes = null;
        }
      }
      return { messageId: s.messageId, desc: suggestionMessage, fixes: sFixes };
    });
  }

  return {
    ruleName,
    messageId: descriptor.messageId,
    message,
    startPos,
    endPos,
    fixes,
    suggestions,
  };
}

/**
 * Translate the variadic forms ESLint's `context.report(...)` accepts
 * into a canonical `ReportDescriptor`. Modeled on
 * `normalizeMultiArgReportCall` in ESLint's `lib/linter/file-report.js`
 * (verified against ESLint main).
 *
 * Forms handled:
 *
 *   - `report({ ... })`             → shallow-clone of the descriptor
 *   - `report(node, message, ...)`  → 2nd arg is a string
 *   - `report(node, loc, message, data?, fix?)` → 2nd arg isn't a string
 *
 * Anything that doesn't match falls through to the modern form via
 * `args[0]` — the worst case is a `null` descriptor that
 * `buildDiagnostic` will reject, matching today's behavior for
 * malformed reports.
 */
export function normalizeReportArgs(args: unknown[]): ReportDescriptor {
  if (args.length <= 1) {
    return { ...((args[0] as ReportDescriptor) ?? {}) };
  }
  // 2+ args → legacy positional form.
  if (typeof args[1] === 'string') {
    return {
      node: args[0] as ESTreeNode,
      // `typeof === 'string'` already narrowed args[1] to string.
      message: args[1],
      data: args[2] as Record<string, string | number> | undefined,
      fix: args[3] as ReportDescriptor['fix'],
    };
  }
  return {
    node: args[0] as ESTreeNode,
    loc: args[1] as ReportDescriptor['loc'],
    message: args[2] as string,
    data: args[3] as Record<string, string | number> | undefined,
    fix: args[4] as ReportDescriptor['fix'],
  };
}

function materializeFixes(
  result: Fix | Fix[] | null | undefined | Iterable<Fix>,
  text: string,
): Fix[] | undefined {
  if (result == null) return undefined;
  if (Array.isArray(result)) {
    // Per-element `isFix` filter. The single-fix and bottom-iterable
    // branches already validate; the array branch used to trust the
    // array verbatim. A plugin returning `[validFix, junk, validFix]`
    // then crashed the diagnostic post-processor at `f.range[0]` —
    // turning a single buggy fix into a parseError for the whole file.
    const filtered = result.filter(isFix);
    if (filtered.length !== result.length) {
      console.error(
        `[rslint] fix(fixer) returned an array with ${
          result.length - filtered.length
        } invalid element(s); dropped them.`,
      );
    }
    return mergeFixes(filtered, text);
  }
  // Reject strings BEFORE the iterable branch — strings are iterable
  // (`Array.from('foo') = ['f','o','o']`), so without this guard the
  // bottom branch turned a buggy `fix: () => 'replacement'` into three
  // character-shaped pseudo-Fix entries; the diagnostic post-processor
  // then crashed at `f.range[0]` (each 'f'/'o'/'o' has no .range). The
  // user's intended fix wasn't applicable anyway, but the cascade
  // wiped the diagnostic entirely.
  // `typeof === 'string'` runtime check via `unknown` cast — strings
  // aren't in the static return-type union, but JS plugins can return
  // anything and we have to defend at runtime.
  if (typeof (result as unknown) === 'string') {
    const s = result as unknown as string;
    console.error(
      `[rslint] fix(fixer) returned a string ("${s.slice(0, 32)}${
        s.length > 32 ? '…' : ''
      }"); expected a Fix object, Fix[], or iterable. Fix dropped.`,
    );
    return undefined;
  }
  if (isFix(result)) return [result];
  // Iterable of fixes — `Array.isArray` + `isFix` already ruled out the
  // single-fix and array cases, so what's left of the original union
  // (`Iterable<Fix>`) is the only branch that can reach here. TS
  // narrows accordingly; no cast needed for `Array.from`.
  //
  // Per-element `isFix` filter on the materialised array, matching the
  // hardening on the Array branch above. Pre-fix `Array.from(result)`
  // accepted whatever the iterable yielded — a `Set([junk, validFix])`
  // or `function*(){ yield junk; yield validFix }` slipped junk
  // through, which then crashed at `f.range[0]` in the diagnostic
  // post-processor (no try/catch there) and wiped EVERY diagnostic on
  // the file via the worker's catch-all. Filter symmetric with the
  // Array branch keeps the contract uniform across all
  // legal-input shapes.
  if (
    typeof (result as { [Symbol.iterator]?: unknown })[Symbol.iterator] ===
    'function'
  ) {
    const arr = Array.from(result);
    const filtered = arr.filter(isFix);
    if (filtered.length !== arr.length) {
      console.error(
        `[rslint] fix(fixer) returned an iterable with ${
          arr.length - filtered.length
        } invalid element(s); dropped them.`,
      );
    }
    return mergeFixes(filtered, text);
  }
  // Anything else is a contract violation by the plugin author — most
  // commonly returning a non-Fix object from `fix(fixer)`. Previously
  // we silently returned undefined here, so the user's fix simply
  // didn't apply, with no clue why. Surface a stderr note so the
  // author can correct it (we still return undefined because there's
  // nothing meaningful to apply; the diagnostic itself still fires).
  console.error(
    `[rslint] fix(fixer) returned an unsupported value (got ${
      result === null ? 'null' : typeof result
    }); expected a Fix object, Fix[], or iterable. Fix dropped.`,
  );
  return undefined;
}

/**
 * Merge the fixes a single `report()` returned into ONE atomic Fix,
 * matching ESLint v10 `lib/linter/file-report.js::mergeFixes`. ESLint
 * treats every report as producing at most one edit: it sorts the
 * fixes by range, then stitches a single `{range:[min,max], text}` by
 * splicing the ORIGINAL source between adjacent fixes. Downstream
 * appliers (`--fix`, LSP code-actions) assume one fix per report, so
 * returning the un-merged `Fix[]` (as the runner did before) risked
 * the multiple edits being conflict-dropped or applied wrong.
 *
 * 0 fixes → undefined; 1 fix → returned as-is (no allocation); ≥2 →
 * the merged single-element array.
 *
 * Overlap handling diverges from ESLint deliberately. ESLint v10
 * (`file-report.js::mergeFixes`) `assert`s `fix.range[0] >= lastPos`
 * after sorting and THROWS `"Fix objects must not be overlapped in a
 * report."` on a violation — verified against eslint@10.4.0. In the
 * runner, throwing here would propagate out of the per-rule listener
 * and the worker's catch-all would wipe EVERY diagnostic on the file,
 * turning one rule's buggy fix into a whole-file failure. Instead we
 * DROP the fix (return `undefined`) and warn on stderr, so the
 * diagnostic still fires but no bogus merged edit is emitted — the same
 * net outcome as ESLint (the fix is not applied) while preserving the
 * runner's per-rule failure isolation.
 */
function mergeFixes(fixes: Fix[], text: string): Fix[] | undefined {
  if (fixes.length === 0) return undefined;
  if (fixes.length === 1) return fixes;
  const sorted = [...fixes].sort(
    (a, b) => a.range[0] - b.range[0] || a.range[1] - b.range[1],
  );
  const start = sorted[0].range[0];
  const end = sorted[sorted.length - 1].range[1];
  let merged = '';
  let lastPos = start;
  for (const f of sorted) {
    // Overlap detection, mirroring ESLint's `fix.range[0] >= lastPos`
    // invariant (here `lastPos` is the previous fix's end). Adjacent
    // (`[0,3]`+`[3,5]`) and touching zero-width (`[3,3]`+`[3,5]`) fixes
    // are NOT overlaps and pass; a contained or crossing range
    // (`[0,8]`+`[3,5]`, `[0,5]`+`[3,8]`) trips it. On the first element
    // `f.range[0] === lastPos` (== start), so this never false-fires.
    if (f.range[0] < lastPos) {
      console.error(
        '[rslint] fix(fixer) returned overlapping fixes in one report; ' +
          'ESLint rejects these ("Fix objects must not be overlapped in a ' +
          'report."). Dropped the fix; the diagnostic still reports.',
      );
      return undefined;
    }
    // Splice the untouched original source between the previous fix's
    // end and this fix's start (empty for adjacent fixes).
    merged += text.slice(lastPos, f.range[0]);
    merged += f.text;
    lastPos = f.range[1];
  }
  merged += text.slice(lastPos, end);
  return [{ range: [start, end], text: merged }];
}

function isFix(x: unknown): x is Fix {
  return (
    x != null &&
    typeof x === 'object' &&
    Array.isArray((x as Fix).range) &&
    typeof (x as Fix).text === 'string'
  );
}

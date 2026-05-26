/* eslint-disable @typescript-eslint/no-unsafe-type-assertion -- AST / parser / scope-manager / plugin-API boundary casts. Each site projects from an `any` / `unknown` peer surface (oxc-parser output, user plugin objects, ESLint v10 wire shapes) into the typed shape this module uses; the contract is runtime-validated at the call boundaries above, not at the cast. Bulk-disabling here instead of per-line keeps the cast sites readable. */
/**
 * Listener merging + AST traversal.
 *
 * Given N rules each returning their own listener map (from
 * `rule.create(ctx)` — the standard ESLint plugin API), merge them
 * into one dispatch table indexed by node type, and walk the AST
 * exactly once.
 *
 * Two selector classes:
 *
 *   - **Simple selector** — a bare node-type name like `'Identifier'` or
 *     `'CallExpression'`. Matched by exact equality on `node.type`. O(1).
 *   - **ESQuery selector** — anything more (e.g.
 *     `'CallExpression > MemberExpression'`, `'[name="foo"]'`). Matched
 *     by walking an esquery AST against the current node + ancestor stack.
 *     O(selectors × per-node).
 *
 * Per-node cancellation poll: the visitor checks `Atomics.load(cancelFlag, 0)`
 * on every node visit. The Atomics.load is cheap enough to do per-node
 * (single-digit ns over a baseline visit) so we don't bother batching;
 * when the flag flips, the walk bails immediately and the caller treats
 * the file as cancelled.
 */

import { createRequire } from 'node:module';
import type { ESTreeNode, ListenerFn } from './context.js';
import { VISITOR_KEYS, getVisitorKeys } from '../ast/visitor-keys.js';

const require = createRequire(import.meta.url);

/** Lazily-loaded esquery — only loaded when an ESQuery selector is registered. */
let _esquery: {
  matches: (node: object, selector: object, ancestry: object[]) => boolean;
  parse: (s: string) => object;
} | null = null;

function esquery() {
  if (_esquery == null) {
    // eslint-disable-next-line @typescript-eslint/no-var-requires -- ESM module loading a CJS-only package (esquery ships CJS); createRequire is the standard ESM↔CJS bridge.
    _esquery = require('esquery') as typeof _esquery extends infer T
      ? T
      : never;
  }
  return _esquery!;
}

/**
 * One listener with rule attribution. A merged-map entry is a list of
 * these instead of a list of raw `ListenerFn`s — when the listener
 * throws, the visitor reports `ruleName` to the error handler so the
 * caller can write a structured `ruleErrors` record back to the
 * per-file lint result. Without attribution, the error reaches stderr
 * with only a selector string, and users can't tell which plugin
 * rule failed (multiple rules can register the same selector).
 */
interface TaggedListener {
  fn: ListenerFn;
  /** Source rule name, e.g. `'unicorn/no-null'`. Undefined when the
   *  listener was injected outside the per-rule pipeline (used by
   *  test fixtures / future internal traversal needs). */
  ruleName?: string;
}

/**
 * Input form for `mergeListeners`. Each entry pairs a rule's listener
 * map with the rule name used for error attribution. `ruleName`
 * omitted is acceptable but loses attribution: listener throws will
 * still be reported, but without telling the caller which rule was
 * responsible — that should only happen for non-rule traversal needs.
 */
export interface TaggedListenerMap {
  ruleName?: string;
  listeners: Record<string, ListenerFn | ListenerFn[]>;
}

/**
 * A compiled esquery-style entry, keyed by listeners + raw text. Pre-
 * computed specificity lets the bucketed dispatch table sort once at
 * merge time instead of per visit — same fields ESLint v10's
 * `ESQueryParsedSelector` stores in `lib/linter/esquery.js`.
 *
 * Bare-type selectors (`Identifier`, `VariableDeclaration`, …) are
 * filed as entries too with `isSimple: true`. ESLint v10 unifies bare
 * and esquery into the same NodeEventGenerator dispatch so they
 * specificity-sort against each other; we mirror that here. `isSimple`
 * lets `fireEnter` skip the `esquery.matches` call (the anchor-type
 * bucket already filtered by `node.type`).
 */
interface EsqEntry {
  /** The parsed esquery selector. `null` for bare-type entries — no esquery match needed. */
  selector: object | null;
  raw: string;
  listeners: TaggedListener[];
  /** Number of attribute / field / nth-child selectors in the parsed AST. */
  attributeCount: number;
  /** Number of `identifier` selectors (a literal node-type name) in the parsed AST. */
  identifierCount: number;
  /** True for bare-type selectors — skip `esquery.matches` and fire directly. */
  isSimple: boolean;
}

/**
 * Result of merging multiple rules' listener maps.
 *
 * `esqueryByType` is the hot-path bucket — same shape ESLint v10 uses
 * in `lib/linter/node-event-generator.js`. Each entry (bare-type OR
 * esquery) is filed under every node type its selector COULD match,
 * derived statically from the parsed selector AST (see
 * `extractAnchorTypes`, modelled on v10's `getPossibleTypes`). The
 * visitor then only runs `esquery.matches` on entries whose anchor set
 * contains the current `node.type`. Selectors with no extractable
 * anchor (e.g. `:not(...)`, attribute-only filters, raw class
 * selectors) fall into `byType.wildcard` and run on every node — same
 * fallback semantics as v10.
 *
 * Bare-type and esquery entries live in the SAME bucket so they share
 * the specificity sort. Pre-fix they had separate pipes (`simple` ran
 * first, then `esquery`), inverting cross-rule observation order vs
 * ESLint on mark-and-sweep style rules (where a `:matches(Identifier)`
 * raw sorts BEFORE bare `Identifier` because `':' < 'I'`).
 *
 * `esqueryList` is the flat catalog kept for diagnostics / debug / hot
 * reload — never iterated on the hot path. Maintained because the
 * bucketed map drops the "first-registration" iteration order needed
 * for cross-rule listener determinism, which is what the flat list
 * preserves.
 */
interface MergedListeners {
  esqueryList: {
    enter: EsqEntry[];
    exit: EsqEntry[];
  };
  esqueryByType: {
    /** anchor-type → entries known to potentially match nodes of that type. */
    enter: Map<string, EsqEntry[]>;
    exit: Map<string, EsqEntry[]>;
    /** entries whose anchor types could not be statically extracted (wildcards). */
    enterWildcard: EsqEntry[];
    exitWildcard: EsqEntry[];
  };
  /**
   * #3: selectors that failed to compile (an invalid esquery selector in a
   * plugin rule). Recorded here and skipped rather than thrown out of
   * `mergeListeners`; the caller surfaces them as per-rule `ruleErrors`.
   */
  selectorErrors: Array<{
    ruleName?: string;
    selector: string;
    message: string;
  }>;
}

/**
 * Merge an array of tagged listener maps into one dispatch table. Each
 * input is a `selector → ListenerFn` map paired with the rule name
 * used for error attribution.
 */
export function mergeListeners(
  maps: Array<TaggedListenerMap>,
): MergedListeners {
  const merged: MergedListeners = {
    esqueryList: { enter: [], exit: [] },
    esqueryByType: {
      enter: new Map(),
      exit: new Map(),
      enterWildcard: [],
      exitWildcard: [],
    },
    selectorErrors: [],
  };

  // Dedup index: `raw → EsqEntry` per direction. ESLint v10's
  // NodeEventGenerator builds ONE EsqueryParsedSelector per (selector,
  // direction) pair, sharing it across every rule that registers the
  // selector. Pre-fix runner inserted a separate entry per rule, so a
  // selector registered by N rules produced N bucket entries with one
  // listener each. V8's `Array.sort` is unstable for buckets >10 with
  // equal `raw` strings, so cross-rule observation order could flip
  // depending on bucket size. Deduplicating to one entry whose
  // `listeners[]` preserves registration order matches ESLint exactly.
  const enterByRaw = new Map<string, EsqEntry>();
  const exitByRaw = new Map<string, EsqEntry>();

  for (const m of maps) {
    if (!m || !m.listeners) continue;
    const ruleName = m.ruleName;
    const map = m.listeners;
    for (const rawSelector of Object.keys(map)) {
      const isExit = rawSelector.endsWith(':exit');
      const cleanSelector = isExit ? rawSelector.slice(0, -5) : rawSelector;
      const fn = map[rawSelector];
      const rawListeners: ListenerFn[] = Array.isArray(fn)
        ? fn
        : fn
          ? [fn]
          : [];
      if (rawListeners.length === 0) continue;
      const tagged: TaggedListener[] = rawListeners.map((f) => ({
        fn: f,
        ruleName,
      }));

      const byRaw = isExit ? exitByRaw : enterByRaw;
      const existing = byRaw.get(cleanSelector);
      if (existing) {
        // Same raw selector already registered (possibly by a
        // different rule) — extend its listener list in registration
        // order. Both bare-type and esquery dedup through here.
        existing.listeners.push(...tagged);
        continue;
      }

      let entry: EsqEntry;
      if (isSimpleSelector(cleanSelector)) {
        // Bare-type selector. Specificity equivalent to a single
        // `identifier` esquery (attributeCount=0, identifierCount=1)
        // — ESLint v10 computes the same numbers via esquery's parse.
        // `selector: null` flags "skip esquery.matches" in fireEnter.
        entry = {
          selector: null,
          raw: cleanSelector,
          attributeCount: 0,
          identifierCount: 1,
          listeners: [...tagged],
          isSimple: true,
        };
      } else {
        let compiled: EsqSelector;
        try {
          compiled = esquery().parse(cleanSelector) as EsqSelector;
        } catch (err) {
          // #3: a buggy plugin rule's invalid esquery selector must not
          // throw out of `mergeListeners` → `lintFile` (whose contract is
          // to never throw; per-rule failures surface via `ruleErrors`).
          // Record it against the owning rule(s) and skip this selector —
          // ESLint isolates a broken rule the same way, leaving sibling
          // rules running.
          for (const tl of tagged) {
            merged.selectorErrors.push({
              ruleName: tl.ruleName,
              selector: cleanSelector,
              message: (err as Error).message,
            });
          }
          continue;
        }
        const { attributeCount, identifierCount } =
          analyzeSpecificity(compiled);
        entry = {
          selector: compiled,
          raw: cleanSelector,
          attributeCount,
          identifierCount,
          listeners: [...tagged],
          isSimple: false,
        };
      }
      byRaw.set(cleanSelector, entry);
      const target = isExit
        ? merged.esqueryList.exit
        : merged.esqueryList.enter;
      target.push(entry);
    }
  }

  // Build the per-anchor-type dispatch table from the flat lists.
  // For bare-type entries the anchor is the type name directly; for
  // esquery entries `bucketByAnchor` derives it via `extractAnchorTypes`.
  const enterBuckets = bucketByAnchor(merged.esqueryList.enter);
  const exitBuckets = bucketByAnchor(merged.esqueryList.exit);
  merged.esqueryByType.enter = enterBuckets.byType;
  merged.esqueryByType.exit = exitBuckets.byType;
  merged.esqueryByType.enterWildcard = enterBuckets.wildcard;
  merged.esqueryByType.exitWildcard = exitBuckets.wildcard;

  return merged;
}

/**
 * A single AST node's `type` is "simple" iff it matches `^[A-Z][a-zA-Z]+$`
 * — covering all standard ESTree / TS-ESLint node types. Anything more
 * complicated must go through esquery.
 */
function isSimpleSelector(s: string): boolean {
  // Reject empty, non-letter starting char, or anything containing
  // characters esquery would interpret (e.g. spaces, brackets, > +).
  return /^[A-Z][a-zA-Z]*$/.test(s);
}

/**
 * Esquery selector AST node — minimal shape we need for type analysis.
 * Mirrors what `esquery.parse` returns. `value` is the identifier name
 * for `type: 'identifier'`; `selectors` is the child set for `not` /
 * `matches` / `compound`; `left` / `right` are the operands for binary
 * combinators (`child` / `descendant` / `sibling` / `adjacent`).
 */
interface EsqSelector {
  type: string;
  value?: string;
  name?: string;
  selectors?: EsqSelector[];
  left?: EsqSelector;
  right?: EsqSelector;
}

/**
 * Static analysis of a parsed esquery selector AST. Returns the set of
 * node types this selector COULD match against — `null` means "could
 * match any type" (fall through to wildcard bucket).
 *
 * Mirrors ESLint v10's `analyzeSelector` in `lib/linter/esquery.js`
 * bit-for-bit (case set, union/intersection semantics, `:function`
 * pseudo-class expansion, attribute / nth-* returning null). Linked
 * to ESLint behavior so plugin rules registering esquery selectors
 * get the same per-type firing semantics whether they're hosted by
 * ESLint or rslint.
 */
/**
 * @internal — exported only for unit-test access. Mirrors ESLint v10's
 * `analyzeSelector` in `lib/linter/esquery.js`; callers should treat the
 * shape as private and route through `mergeListeners` in production.
 */
export function extractAnchorTypes(
  selector: EsqSelector,
): readonly string[] | null {
  switch (selector.type) {
    case 'identifier':
      return selector.value != null ? [selector.value] : null;

    case 'wildcard':
      return null;

    case 'not':
      // Negation can match any type — caller falls through to wildcard.
      return null;

    case 'matches': {
      const subs = selector.selectors ?? [];
      const each = subs.map(extractAnchorTypes);
      // If every sub-selector is typed, the union is the answer; if
      // any sub is wildcard, the whole match becomes wildcard.
      if (each.every((t) => t != null)) {
        // TS 5.5+ infers `(t) => t != null` as a `t is NonNullable<…>`
        // type predicate, so `each` is narrowed to `readonly string[][]`
        // here — no per-element cast needed.
        const merged = new Set<string>();
        for (const t of each) for (const v of t) merged.add(v);
        return [...merged];
      }
      return null;
    }

    case 'compound': {
      const subs = selector.selectors ?? [];
      const typed = subs
        .map(extractAnchorTypes)
        .filter((t): t is readonly string[] => t != null);
      // All components untyped → wildcard.
      if (typed.length === 0) return null;
      // Intersection across typed sub-components — a compound only
      // matches a node iff every typed sub-component matches it.
      let result = new Set<string>(typed[0]);
      for (let i = 1; i < typed.length; i++) {
        const next = new Set(typed[i]);
        result = new Set([...result].filter((v) => next.has(v)));
      }
      return [...result];
    }

    case 'child':
    case 'descendant':
    case 'sibling':
    case 'adjacent':
      // Binary combinators — the anchor is the RIGHT side (the node
      // being matched). `left` is the qualifier (e.g. `A > B` means
      // "B whose parent is A" → anchor = B).
      return selector.right ? extractAnchorTypes(selector.right) : null;

    case 'class':
      // Pseudo-class. v10 specifically expands `:function` to the
      // three function-bearing node types; everything else is wildcard.
      if (selector.name === 'function') {
        return [
          'FunctionDeclaration',
          'FunctionExpression',
          'ArrowFunctionExpression',
        ];
      }
      return null;

    // attribute / field / nth-child / nth-last-child / has — all
    // run on top of an outer node, no type information to extract.
    case 'attribute':
    case 'field':
    case 'nth-child':
    case 'nth-last-child':
    case 'has':
    default:
      return null;
  }
}

/**
 * File esquery entries into per-anchor-type buckets + wildcard list.
 * Returns a function callable as `getMatches(nodeType)` that yields
 * the entries the visitor should run through.
 *
 * Build is O(N_selectors × selector-AST-size); each entry is placed
 * in (#anchor-types) buckets, or in the wildcard list if no anchor
 * could be extracted. Lookup is then O(1) per node via Map.get plus
 * the wildcard array.
 */
function bucketByAnchor(entries: EsqEntry[]): {
  byType: Map<string, EsqEntry[]>;
  wildcard: EsqEntry[];
} {
  const byType = new Map<string, EsqEntry[]>();
  const wildcard: EsqEntry[] = [];
  for (const entry of entries) {
    // Bare-type entry: anchor is the raw type name directly. Bypass
    // `extractAnchorTypes` since selector is null.
    const anchors = entry.isSimple
      ? [entry.raw]
      : extractAnchorTypes(entry.selector as EsqSelector);
    if (anchors == null) {
      wildcard.push(entry);
      continue;
    }
    for (const t of anchors) {
      const list = byType.get(t);
      if (list) list.push(entry);
      else byType.set(t, [entry]);
    }
  }
  // Sort each bucket + the wildcard list by ESLint v10's selector
  // specificity rule (`lib/linter/esquery.js::compare`):
  // attributeCount → identifierCount → raw string lexicographic.
  // Without this, rules that depend on a more-specific selector
  // firing AFTER a less-specific one on the same node (mark-and-sweep
  // patterns common in `@typescript-eslint` and `eslint-plugin-
  // unicorn`) get the wrong observation order.
  for (const list of byType.values()) list.sort(compareSpecificity);
  wildcard.sort(compareSpecificity);
  return { byType, wildcard };
}

function compareSpecificity(a: EsqEntry, b: EsqEntry): number {
  return (
    a.attributeCount - b.attributeCount ||
    a.identifierCount - b.identifierCount ||
    (a.raw <= b.raw ? -1 : 1)
  );
}

/**
 * Mirrors ESLint v10's `lib/linter/esquery.js::analyzeParsedSelector` —
 * counts attribute / field / nth-child / nth-last-child as attributes,
 * and bare `identifier` selectors (literal node-type names) as
 * identifiers. Both totals feed the specificity comparator.
 *
 * The walk shape matches `extractAnchorTypes` deliberately — same
 * recursion structure, different per-node accumulator. Keep the two
 * in lockstep when adding new selector grammar handling.
 */
function analyzeSpecificity(selector: EsqSelector): {
  attributeCount: number;
  identifierCount: number;
} {
  let attributeCount = 0;
  let identifierCount = 0;
  function walk(s: EsqSelector | undefined | null): void {
    if (!s) return;
    switch (s.type) {
      case 'identifier':
        identifierCount++;
        return;
      case 'wildcard':
        return;
      case 'attribute':
      case 'field':
      case 'nth-child':
      case 'nth-last-child':
        attributeCount++;
        return;
      case 'matches':
      case 'not':
      case 'compound':
        if (Array.isArray(s.selectors))
          for (const sub of s.selectors) walk(sub);
        return;
      case 'has':
        // ESLint v10's `analyzeParsedSelector` does NOT recurse into
        // `:has()` — the function only counts identifiers reachable
        // along the OUTER selector path. Treating `:has()` like
        // matches/not double-counts the inner identifiers and
        // inflates specificity (e.g.
        // `CallExpression:has(Identifier MemberExpression)` came out
        // at id=3 vs ESLint's id=1), reversing tie-breaks against
        // sibling selectors.
        return;
      case 'child':
      case 'descendant':
      case 'sibling':
      case 'adjacent':
        walk(s.left);
        walk(s.right);
        return;
      case 'class':
      default:
        // class / unknown — no contribution to counts.
        return;
    }
  }
  walk(selector);
  return { attributeCount, identifierCount };
}

/**
 * Linear merge of two pre-sorted EsqEntry arrays by specificity. Both
 * inputs are already sorted at `mergeListeners` time via
 * `compareSpecificity`; this merge keeps the combined sequence sorted
 * without allocating an intermediate array. Used by fireEnter /
 * fireExit to interleave byType + wildcard bucket events in the order
 * ESLint v10's NodeEventGenerator produces them.
 */
function mergeByLinearMerge(
  byType: EsqEntry[] | undefined,
  wild: EsqEntry[],
  fire: (e: EsqEntry) => void,
): void {
  const bt = byType ?? [];
  let i = 0;
  let j = 0;
  while (i < bt.length && j < wild.length) {
    if (compareSpecificity(bt[i], wild[j]) <= 0) fire(bt[i++]);
    else fire(wild[j++]);
  }
  while (i < bt.length) fire(bt[i++]);
  while (j < wild.length) fire(wild[j++]);
}

/**
 * Cancellation: the runner allocates an Int32Array of length 1 in a
 * SharedArrayBuffer (one slot per task) and hands it to the worker. The
 * worker passes that view here as `cancelFlag`. Set to non-zero to
 * cancel; the next node visit observes it via Atomics.load and the
 * walker returns early.
 */
export interface VisitOptions {
  cancelFlag?: Int32Array;
  /**
   * Caught listener errors are reported here. The visitor itself does
   * NOT throw — a buggy rule must not abort the file. If `null`, errors
   * are swallowed silently; pass a function to log them.
   */
  onListenerError?: (info: {
    /** Source rule name, when known. Undefined for orphaned listeners
     *  (no rule attribution wired in by the caller). */
    ruleName?: string;
    /** Selector text that triggered this listener (raw, pre-parse). */
    selector: string;
    /** AST node visited when the listener threw. */
    node: ESTreeNode;
    /** Thrown value. */
    err: unknown;
  }) => void;
}

/** Result of a single-file visit. `cancelled` true means we bailed early. */
export interface VisitResult {
  cancelled: boolean;
  visitedNodes: number;
}

/**
 * Walk `root` once, dispatching to `merged`'s listeners. Maintains an
 * ancestor stack for ESQuery to use.
 */
export function visit(
  root: ESTreeNode,
  merged: MergedListeners,
  opts: VisitOptions = {},
): VisitResult {
  // `ancestry` here is the stack pushed by DFS (root pushed first,
  // immediate parent last). esquery's contract, on the other hand,
  // expects `ancestry[0]` to be the IMMEDIATE PARENT (innermost first,
  // root last) — its `>` (child) combinator does
  // `matches(ancestry[0], left, ancestry.slice(1))`. Passing the stack
  // verbatim would invert every parent-walking selector and break
  // combinators like `FunctionDeclaration > ReturnStatement`,
  // `MemberExpression > Identifier.property`, etc. The pre-fix runner
  // had this exact bug — every esquery selector with `>` `+` or `~`
  // silently never fired. Empirically pinned against ESLint v10 by
  // running the same `'FunctionDeclaration > BlockStatement > ReturnStatement'`
  // selector through both engines.
  const ancestry: ESTreeNode[] = [];
  const reversedAncestry = (): ESTreeNode[] => {
    if (ancestry.length === 0) return ancestry;
    const out = new Array<ESTreeNode>(ancestry.length);
    for (let i = 0; i < ancestry.length; i++)
      out[i] = ancestry[ancestry.length - 1 - i];
    return out;
  };
  const result: VisitResult = { cancelled: false, visitedNodes: 0 };
  const errH = opts.onListenerError;
  const cancelFlag = opts.cancelFlag;

  // No-ESQuery fast path. When the merged set contains no esquery
  // selectors at all — every registered selector is a bare node type
  // — the DFS does not need to maintain an ancestor stack, because
  // ancestry is consumed ONLY by `esquery.matches`. Skipping the
  // per-node `ancestry.push(node)` / `ancestry.pop()` removes one
  // array op pair per node visit (≈O(N) over the AST). Modelled on
  // ESLint v10's NodeEventGenerator which similarly elides ancestry
  // maintenance when its `currentSelectors` list is empty. The
  // typescript-eslint rule set is dominated by bare-type selectors,
  // so this branch fires for most real-world configs.
  const hasEsquery =
    merged.esqueryByType.enter.size > 0 ||
    merged.esqueryByType.exit.size > 0 ||
    merged.esqueryByType.enterWildcard.length > 0 ||
    merged.esqueryByType.exitWildcard.length > 0;

  // Inline helpers — closures over loop-stable state for hot path.
  // Bare-type entries (`isSimple: true`, `selector: null`) skip the
  // `esquery.matches` step entirely — the by-type bucket lookup
  // already filtered to nodes whose type matches the selector.
  const fireEnter = (node: ESTreeNode): void => {
    const enterByType = merged.esqueryByType.enter.get(node.type);
    const enterWild = merged.esqueryByType.enterWildcard;
    if ((enterByType && enterByType.length > 0) || enterWild.length > 0) {
      const esq = esquery();
      const rev = reversedAncestry();
      const fire = (e: EsqEntry): void => {
        // Bare-type entries: by-type bucket guarantees a match, fire
        // listeners directly.
        if (e.isSimple) {
          for (const tl of e.listeners) {
            try {
              tl.fn(node);
            } catch (err) {
              errH?.({ ruleName: tl.ruleName, selector: e.raw, node, err });
            }
          }
          return;
        }
        // `esq.matches` itself can throw for poorly-formed selectors
        // (e.g. unknown pseudo-class names). Without this outer
        // try/catch the throw escapes `fire` → `fireEnter` → `dfs` →
        // `visit` → `lintFile`, killing every diagnostic on the file
        // for what is really a single bad selector. Surface it via
        // `errH` (per-rule error channel) and skip the entry so the
        // remaining selectors and listeners still run.
        //
        // `e.selector` is non-null on this branch — `isSimple` is the
        // only flag that sets it to null and that case returned above.
        let matched: boolean;
        try {
          matched = esq.matches(node, e.selector!, rev);
        } catch (err) {
          // Attribute the failure to the FIRST listener registered
          // under the bad selector — every listener in `e.listeners`
          // came in via the same selector, so any of them is a fair
          // pointer back to the rule that registered it.
          errH?.({
            ruleName: e.listeners[0]?.ruleName,
            selector: e.raw,
            node,
            err,
          });
          return;
        }
        if (matched) {
          for (const tl of e.listeners) {
            try {
              tl.fn(node);
            } catch (err) {
              errH?.({ ruleName: tl.ruleName, selector: e.raw, node, err });
            }
          }
        }
      };
      // Merge byType and wildcard buckets by specificity (both are
      // already individually sorted at mergeListeners time). ESLint
      // v10's NodeEventGenerator dispatches all matching selectors
      // through ONE specificity-sorted list per node — pre-fix runner
      // ran byType fully and then wildcard fully, inverting the
      // observation order for any wildcard with lower specificity
      // than a sibling by-type selector on the same node.
      mergeByLinearMerge(enterByType, enterWild, fire);
    }
  };

  const fireExit = (node: ESTreeNode): void => {
    const exitByType = merged.esqueryByType.exit.get(node.type);
    const exitWild = merged.esqueryByType.exitWildcard;
    if ((exitByType && exitByType.length > 0) || exitWild.length > 0) {
      const esq = esquery();
      const rev = reversedAncestry();
      const fire = (e: EsqEntry): void => {
        if (e.isSimple) {
          for (const tl of e.listeners) {
            try {
              tl.fn(node);
            } catch (err) {
              errH?.({ ruleName: tl.ruleName, selector: e.raw, node, err });
            }
          }
          return;
        }
        // Same throw-isolation rationale as fireEnter — see comment above.
        let matched: boolean;
        try {
          matched = esq.matches(node, e.selector!, rev);
        } catch (err) {
          errH?.({
            ruleName: e.listeners[0]?.ruleName,
            selector: e.raw,
            node,
            err,
          });
          return;
        }
        if (matched) {
          for (const tl of e.listeners) {
            try {
              tl.fn(node);
            } catch (err) {
              errH?.({ ruleName: tl.ruleName, selector: e.raw, node, err });
            }
          }
        }
      };
      mergeByLinearMerge(exitByType, exitWild, fire);
    }
  };

  // Child-key lookup mirrors ESLint v10's `lib/shared/traverser.js`
  // exactly: `visitorKeys[node.type]` is consulted first (O(1));
  // unknown types fall back to `getVisitorKeys` which delegates to
  // `eslint-visitor-keys`'s enumerable-key filter. This both
  //   (a) aligns the runner's recursion order with espree /
  //       @typescript-eslint/parser bit-for-bit, and
  //   (b) eliminates the prior `Object.keys + 'parent'/'loc'/'range'`
  //       blacklist per node, removing ~N short-lived allocs per file.
  const dfs = (node: ESTreeNode): boolean /* keep going */ => {
    // Per-node cancel check. The Atomics.load is cheap enough to run
    // unconditionally, so we don't bother amortizing across node count.
    if (cancelFlag && Atomics.load(cancelFlag, 0) !== 0) {
      result.cancelled = true;
      return false;
    }
    result.visitedNodes++;

    fireEnter(node);

    const keys =
      VISITOR_KEYS[(node as { type: string }).type] ??
      getVisitorKeys(node as { type?: string });
    if (keys.length > 0) {
      // `hasEsquery` is hoisted from the enclosing closure and constant
      // for the entire walk; V8's JIT promotes it to a hot constant so
      // the per-node branch is effectively free (verified empirically
      // — pure simple-selector configs see net wins, mixed configs no
      // measurable regression).
      if (hasEsquery) ancestry.push(node);
      for (let i = 0; i < keys.length; i++) {
        const v = (node as Record<string, unknown>)[keys[i]];
        if (v == null) continue;
        if (Array.isArray(v)) {
          for (let j = 0; j < v.length; j++) {
            const c = v[j];
            if (c && typeof c === 'object' && (c as ESTreeNode).type) {
              if (!dfs(c as ESTreeNode)) {
                if (hasEsquery) ancestry.pop();
                return false;
              }
            }
          }
        } else if (typeof v === 'object' && (v as ESTreeNode).type) {
          if (!dfs(v as ESTreeNode)) {
            if (hasEsquery) ancestry.pop();
            return false;
          }
        }
      }
      if (hasEsquery) ancestry.pop();
    }

    fireExit(node);
    return true;
  };

  dfs(root);
  return result;
}

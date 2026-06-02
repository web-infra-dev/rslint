/**
 * The `fixer` object passed to ESLint plugin code via `descriptor.fix(fixer)`
 * (and `descriptor.suggest[i].fix(fixer)`). Plugins call methods on this
 * object to express their intended edit; each method returns a `Fix`
 * record that the runner ships back to Go for application.
 *
 * The full ESLint fixer surface is 8 methods. Most plugins use the
 * insert/remove variants, not just `replaceText`, so leaving any out
 * breaks real plugins.
 */

/**
 * A single fix edit. Mirrors ESLint's `Fix` interface: byte range +
 * replacement text. Empty text + non-empty range = remove. Empty range
 * (start === end) + non-empty text = insert.
 */
export interface Fix {
  /**
   * Offset range `[start, end)` into the source text.
   *
   * INSIDE the runner these are UTF-16 code-unit indices (matching
   * native JS string indexing). On the IPC boundary they get
   * converted to UTF-8 BYTE offsets — Go's TextRange consumes bytes.
   * The conversion lives in `ecma-language-plugin.ts` next to the
   * diagnostic drain; do not double-convert from rule code.
   */
  range: [number, number];
  /** Replacement text (may be empty for remove). */
  text: string;
}

/**
 * The fixer object plugins call into. Eight methods, one record per call.
 * Stateless — each method returns a fresh Fix; the caller (RuleContext's
 * `report`) collects them into the diagnostic.
 */
export interface RuleFixer {
  // ── Replacement ──
  replaceText(node: HasRange, text: string): Fix;
  replaceTextRange(range: [number, number], text: string): Fix;

  // ── Insertion (zero-width range) ──
  insertTextBefore(node: HasRange, text: string): Fix;
  insertTextBeforeRange(range: [number, number], text: string): Fix;
  insertTextAfter(node: HasRange, text: string): Fix;
  insertTextAfterRange(range: [number, number], text: string): Fix;

  // ── Removal (empty replacement) ──
  remove(node: HasRange): Fix;
  removeRange(range: [number, number]): Fix;
}

/** Minimal node shape — the fixer only needs `range`. */
interface HasRange {
  range: [number, number];
}

/**
 * Build the fixer object. It is stateless and shareable across all rules
 * in one file; we return a singleton per file rather than a new object
 * per rule, but per-rule is also fine — the cost is tiny.
 */
/**
 * Throws a clear, attributable error when a rule passes a synthetic /
 * incomplete node (no `range`) to a fixer node-API method. Without
 * this guard, `copyRange(undefined)` produced a generic TypeError
 * that the diagnostic-builder's outer try/catch absorbed into a
 * `console.error` line — the diagnostic still fired but its fix was
 * silently dropped, and the user running `--fix` saw nothing change
 * with no recoverable signal. The error class lets the wrapping
 * try/catch surface it via `ruleErrors` (per-rule channel) the same
 * way an exception inside the rule body is surfaced.
 */
function assertNodeRange(node: HasRange | undefined, method: string): void {
  if (!node || !node.range) {
    throw new TypeError(
      `fixer.${method}: node has no \`range\` — was a synthetic / virtual node passed in?`,
    );
  }
}

export function makeFixer(): RuleFixer {
  return {
    replaceText: (node, text) => {
      assertNodeRange(node, 'replaceText');
      return { range: copyRange(node.range), text };
    },
    replaceTextRange: (range, text) => ({ range: copyRange(range), text }),

    insertTextBefore: (node, text) => {
      assertNodeRange(node, 'insertTextBefore');
      return { range: [node.range[0], node.range[0]], text };
    },
    insertTextBeforeRange: (range, text) => ({
      range: [range[0], range[0]],
      text,
    }),
    insertTextAfter: (node, text) => {
      assertNodeRange(node, 'insertTextAfter');
      return { range: [node.range[1], node.range[1]], text };
    },
    insertTextAfterRange: (range, text) => ({
      range: [range[1], range[1]],
      text,
    }),

    remove: (node) => {
      assertNodeRange(node, 'remove');
      return { range: copyRange(node.range), text: '' };
    },
    removeRange: (range) => ({ range: copyRange(range), text: '' }),
  };
}

/**
 * Defensive copy: callers may mutate the returned Fix's `range`, and
 * mutating the original node's `range` would break invariants downstream.
 */
function copyRange(r: readonly [number, number]): [number, number] {
  return [r[0], r[1]];
}

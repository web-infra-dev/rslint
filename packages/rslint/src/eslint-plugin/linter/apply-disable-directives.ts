/**
 * Per-file `eslint-disable` / `eslint-enable` filter.
 *
 * Two-pass algorithm mirroring ESLint v10's
 * `lib/linter/apply-disable-directives.js`:
 *
 *   1. Block directives (`/* eslint-disable rule *\/` ... `/* eslint-enable
 *      rule *\/`) form open-ended disable regions.
 *   2. Line directives (`// eslint-disable-line`,
 *      `// eslint-disable-next-line`) are equivalent to a single-line
 *      disable region whose `enable` lands at the next line's start.
 *
 * Crucially the two kinds are applied in SEPARATE passes — running them
 * combined-sorted would let a line-form `enable` clear an outstanding
 * block-form disable (or vice versa), changing the meaning. See
 * ESLint's source for the same partition.
 *
 * We diverge from ESLint in three places, intentionally:
 *
 *   - **No "unused-disable-directive" reporting.** That feature is
 *     specific to ESLint's linter; the rslint compat path simply drops
 *     suppressed diagnostics. The `reportUnusedDisableDirectives`
 *     setting is configured in flat config but rslint compat doesn't
 *     surface lint-core meta-errors (an aligned design choice — see
 *     bench-eslint-3.config.mjs note).
 *   - **No `fix` / `suggest` removal for unused directives.** Same
 *     reason as above.
 *   - **Position comparison uses char offsets** instead of (line,
 *     column). The two are isomorphic — offset monotonically maps to
 *     ESLint's `compareLocations` order — and saves us from carrying a
 *     parallel column-base bookkeeping across the runner.
 *
 * `disable-next-line`'s target line is derived from `comment.loc.end.line
 *  + 1`, not `start.line + 1`. ESLint's `createDisableDirectives` uses
 *  `loc.end` for this kind for a reason: a multi-line `/* eslint-disable
 *  -next-line foo ... *\/` should suppress the line AFTER the comment
 *  ends, not the line after the comment starts. Other directive kinds
 *  use `loc.start.line` for the same reason — they target the comment's
 *  beginning.
 *
 * Reference algorithm (post-simplification of the unused-reporting path):
 * for each diagnostic D in sorted directives list, walk directives whose
 * pos ≤ D.startPos; track `active` count of disable-directives that
 * match D.ruleName; a matching `enable` resets `active` to 0 (NOT a
 * decrement — ESLint's `disableDirectivesForProblem = []` semantics).
 * Diagnostic is suppressed iff `active > 0` after the walk.
 */

import { ConfigCommentParser } from '@eslint/plugin-kit';
import type { Comment } from '../source-code/token-builder.js';
import type { Diagnostic } from './diagnostic-builder.js';
import { lineStartOffset } from '../source-code/source-code-helpers.js';

// Reused across files (parser is stateless after construction; cheap to
// share — plugin-kit's parser doesn't hold per-call state).
const parser = new ConfigCommentParser();

interface BlockDirective {
  type: 'disable' | 'enable';
  pos: number;
  // null = applies to all rules; otherwise the specific rule id this
  // directive targets. `/* eslint-disable foo, bar */` expands to two
  // BlockDirectives (one per rule) so the per-diagnostic walk can do a
  // straight `ruleId === d.ruleName` test.
  ruleId: string | null;
}

export interface ApplyDisableOptions {
  comments: readonly Comment[];
  diagnostics: readonly Diagnostic[];
  lineStartOffsets: readonly number[];
  textLength: number;
}

/**
 * Returns a new diagnostic array with every diagnostic suppressed by an
 * `eslint-disable` / `eslint-disable-line` / `eslint-disable-next-line`
 * directive removed.
 *
 * Pure function — does not mutate inputs.
 */
export function applyDisableDirectives(
  opts: ApplyDisableOptions,
): Diagnostic[] {
  if (opts.diagnostics.length === 0) return [];

  const { blockDirectives, lineDirectives } = collectDirectives(
    opts.comments,
    opts.lineStartOffsets,
    opts.textLength,
  );

  // Fast path: zero directives in either kind. ESLint's
  // `applyDirectives` short-circuits the same way (the inner for-loop
  // simply emits every problem).
  if (blockDirectives.length === 0 && lineDirectives.length === 0) {
    return opts.diagnostics.slice();
  }

  // Block pass first. The output then feeds the line pass — the SAME
  // two-pass structure ESLint uses. See module-level comment on why.
  const afterBlock =
    blockDirectives.length > 0
      ? filterOnce(opts.diagnostics, blockDirectives)
      : opts.diagnostics.slice();
  return lineDirectives.length > 0
    ? filterOnce(afterBlock, lineDirectives)
    : afterBlock;
}

// ─────────────────────────────────────────────────────────────────────
// Internals
// ─────────────────────────────────────────────────────────────────────

interface CollectedDirectives {
  blockDirectives: BlockDirective[];
  lineDirectives: BlockDirective[];
}

function collectDirectives(
  comments: readonly Comment[],
  lso: readonly number[],
  textLength: number,
): CollectedDirectives {
  const blockDirectives: BlockDirective[] = [];
  const lineDirectives: BlockDirective[] = [];

  for (const c of comments) {
    // Shebang and any other non-comment node has nothing to parse.
    if (c.type !== 'Line' && c.type !== 'Block') continue;

    const parsed = parser.parseDirective(c.value);
    if (!parsed) continue;

    // Filter to the 4 disable/enable labels. Both `eslint-` and
    // `rslint-` prefixes are accepted and treated as equivalent —
    // matches the Go-side `internal/rule/disable_manager.go:37-38`
    // contract and the user-facing documentation at
    // `website/docs/en/guide/inline-directives.md`, which lets users
    // mix `rslint-disable` with `eslint-enable` (and vice versa)
    // freely. The plugin-kit parser also recognizes `eslint`,
    // `global`, `globals`, `exported` — none of those affect
    // diagnostic suppression, so we skip silently.
    let kind:
      | 'block-disable'
      | 'block-enable'
      | 'line-disable'
      | 'line-next-disable';
    switch (parsed.label) {
      case 'eslint-disable':
      case 'rslint-disable':
      case 'eslint-enable':
      case 'rslint-enable':
        // ESLint v10 (`lib/linter/linter.js:445`) requires these four
        // labels to be carried by a BLOCK comment. A line carrier is
        // silently ignored — `lineCommentSupported = /^eslint-disable-
        // (next-)?line$/` rejects them. Pre-fix the switch matched on
        // label only, so `// eslint-disable foo` over-suppressed and
        // `// eslint-enable foo` could re-open a block-disable region.
        if (c.type !== 'Block') continue;
        kind =
          parsed.label === 'eslint-disable' || parsed.label === 'rslint-disable'
            ? 'block-disable'
            : 'block-enable';
        break;
      case 'eslint-disable-line':
      case 'rslint-disable-line':
        // `disable-line` directives are spec-bound to a single line.
        // ESLint rejects multi-line block comments carrying this
        // label as INVALID — it reports a problem AND does not
        // suppress. `source-code.ts::getDisableDirectives` already
        // applies this validation when surfaced to user rules;
        // mirroring it here keeps the suppression engine consistent
        // with the inspection API.
        if (c.type === 'Block' && c.loc.start.line !== c.loc.end.line) {
          continue;
        }
        kind = 'line-disable';
        break;
      case 'eslint-disable-next-line':
      case 'rslint-disable-next-line':
        kind = 'line-next-disable';
        break;
      default:
        continue;
    }

    const ruleIds = parseRuleList(parsed.value);

    if (kind === 'block-disable' || kind === 'block-enable') {
      const type = kind === 'block-disable' ? 'disable' : 'enable';
      const pos = c.range[0];
      pushPerRule(blockDirectives, type, pos, ruleIds);
    } else {
      // disable-next-line uses end.line so a multi-line block comment
      // targets the line AFTER the comment closes, not after it opens.
      // Other line directives anchor on start.line.
      const targetLine =
        kind === 'line-disable' ? c.loc.start.line : c.loc.end.line + 1;
      const disablePos = lineStartOffset(lso, targetLine, textLength);
      const enablePos = lineStartOffset(lso, targetLine + 1, textLength);
      pushPerRule(lineDirectives, 'disable', disablePos, ruleIds);
      pushPerRule(lineDirectives, 'enable', enablePos, ruleIds);
    }
  }

  blockDirectives.sort(byPos);
  lineDirectives.sort(byPos);
  return { blockDirectives, lineDirectives };
}

function byPos(a: BlockDirective, b: BlockDirective): number {
  return a.pos - b.pos;
}

function pushPerRule(
  dst: BlockDirective[],
  type: 'disable' | 'enable',
  pos: number,
  ruleIds: string[] | null,
): void {
  if (ruleIds === null) {
    dst.push({ type, pos, ruleId: null });
    return;
  }
  for (const r of ruleIds) dst.push({ type, pos, ruleId: r });
}

function parseRuleList(value: string): string[] | null {
  // `parser.parseDirective` returns `.value` as the substring AFTER the
  // label and BEFORE the `--` justification — but it may include
  // leading whitespace. parseListConfig is whitespace-tolerant, but
  // returning `null` for "no rules listed" requires the trim check.
  const trimmed = value.trim();
  if (trimmed === '') return null;
  const map = parser.parseListConfig(trimmed);
  const ids = Object.keys(map);
  return ids.length === 0 ? null : ids;
}

function filterOnce(
  diagnostics: readonly Diagnostic[],
  sortedDirectives: readonly BlockDirective[],
): Diagnostic[] {
  // O(P × D) inner loop intentionally — ESLint does the same (each
  // problem rewalks directives from index 0 to maintain a fresh active
  // stack per diagnostic). The directives list is short in practice
  // (single-digit-to-low-hundred for vscode-sized files).
  const out: Diagnostic[] = [];
  for (const d of diagnostics) {
    let active = 0;
    for (let i = 0; i < sortedDirectives.length; i++) {
      const dir = sortedDirectives[i];
      if (dir.pos > d.startPos) break;
      // ruleId === null is the unscoped form (`/* eslint-disable */`),
      // which applies to every rule.
      if (dir.ruleId === null || dir.ruleId === d.ruleName) {
        if (dir.type === 'disable') {
          active++;
        } else {
          // ESLint's `disableDirectivesForProblem = []` — any matching
          // `enable` resets the active count to 0 (not a decrement).
          active = 0;
        }
      }
    }
    if (active === 0) out.push(d);
  }
  return out;
}

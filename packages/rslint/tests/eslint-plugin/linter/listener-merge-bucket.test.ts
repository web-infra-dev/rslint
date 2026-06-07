/**
 * Anchor-type extraction + bucketed dispatch tests for listener-merge.
 *
 * Two layers of coverage:
 *
 *   1. `extractAnchorTypes` unit tests — verify the static analysis of
 *      an esquery selector AST against every grammar shape in ESLint
 *      v10's `lib/linter/esquery.js::analyzeSelector`. Each test takes
 *      a raw selector string, parses it via the `esquery` package, and
 *      asserts the anchor-type set the function returns. A `null` return
 *      means "wildcard — could match any node type"; an array means "only
 *      these node types can match".
 *
 *   2. `mergeListeners` integration tests — verify the bucketed dispatch
 *      table the visitor consumes. For each selector class we check
 *      both that the right `esqueryByType.enter` bucket is populated AND
 *      that wildcards land in `enterWildcard`. These tests double as the
 *      regression net for the visitor's `fireEnter` per-type lookup.
 *
 * Why both: the unit tests pin the algorithm's case set; the integration
 * tests pin the runtime data structure shape. A bug in either path
 * (helper mis-classifying a case, or `bucketByAnchor` mis-filing the
 * result) would be caught by one of the two.
 */

import { describe, test, expect } from '@rstest/core';
import { createRequire } from 'node:module';

import {
  extractAnchorTypes,
  mergeListeners,
  visit,
} from '../../../src/eslint-plugin/linter/listener-merge.js';
import type { ESTreeNode } from '../../../src/eslint-plugin/linter/context.js';

const require = createRequire(import.meta.url);
const esquery = require('esquery') as { parse: (s: string) => object };

describe('visit: esquery match throw is isolated, does not kill file', () => {
  // Regression for review P0 #3 — `esq.matches` can throw at runtime
  // for selectors that parsed cleanly but reference an unknown
  // pseudo-class (`Identifier:enter` etc.). Without the outer
  // try/catch in `fire`, the throw cascaded up through dfs → visit →
  // lintFile and the file's diagnostics came back empty. The fix
  // wraps `esq.matches` and reports the throw via the
  // `onListenerError` callback, leaving the remaining listeners free
  // to fire.
  test('selector that throws on matches is reported via onListenerError; siblings still fire', () => {
    const errors: Array<{
      ruleName: string | undefined;
      selector: string;
      message: string;
    }> = [];
    const goodSelectorHits: ESTreeNode[] = [];

    // Hand-built tiny AST. visit() doesn't need a parsed program — it
    // walks via visitorKeys / Object.keys, and an Identifier inside an
    // ExpressionStatement triggers both bucketed dispatch + wildcard.
    const ast = {
      type: 'Program',
      body: [
        {
          type: 'ExpressionStatement',
          expression: { type: 'Identifier', name: 'foo' },
        },
      ],
    } as unknown as ESTreeNode;

    const merged = mergeListeners([
      {
        ruleName: 'bad-rule',
        listeners: {
          // `Identifier:enter` parses as a class-pseudo esquery
          // doesn't recognise → `esq.matches` throws "Unknown class
          // name: enter" at runtime.
          'Identifier:enter'() {
            /* never reached — esq.matches throws first */
          },
        },
      },
      {
        ruleName: 'good-rule',
        listeners: {
          Identifier(node) {
            goodSelectorHits.push(node);
          },
        },
      },
    ]);

    const result = visit(ast, merged, {
      onListenerError: (rec) => {
        errors.push({
          ruleName: rec.ruleName,
          selector: rec.selector,
          message: rec.err instanceof Error ? rec.err.message : String(rec.err),
        });
      },
    });

    // Pre-fix: throw propagated, visit returned early, goodSelectorHits
    // was empty. Post-fix: bad selector goes to onListenerError, good
    // selector still fires.
    expect(result.visitedNodes).toBeGreaterThan(0);
    expect(goodSelectorHits.length).toBeGreaterThan(0);
    expect(goodSelectorHits[0].type).toBe('Identifier');

    // Bad selector reported with rule attribution.
    const badErr = errors.find((e) => e.selector === 'Identifier:enter');
    expect(badErr).toBeDefined();
    expect(badErr?.ruleName).toBe('bad-rule');
  });
});

/**
 * Helper: parse a selector via esquery and run the anchor extractor on it.
 * Throws if the selector is unparseable so tests fail loudly rather than
 * silently treating bad input as wildcard.
 */
function anchorsOf(selector: string): readonly string[] | null {
  const ast = esquery.parse(selector) as Parameters<
    typeof extractAnchorTypes
  >[0];
  return extractAnchorTypes(ast);
}

/**
 * Helper: build a MergedListeners from a single rule's listener map and
 * return the post-bucket dispatch table for assertions.
 */
function mergeAndGetBuckets(listeners: Record<string, () => void>) {
  return mergeListeners([{ ruleName: 'test', listeners }]).esqueryByType;
}

// ─────────────────────────────────────────────────────────────────────
// extractAnchorTypes — unit
// ─────────────────────────────────────────────────────────────────────

describe('extractAnchorTypes — esquery grammar coverage', () => {
  test('bare identifier → single anchor', () => {
    expect(anchorsOf('Identifier')).toEqual(['Identifier']);
    expect(anchorsOf('TSEnumDeclaration')).toEqual(['TSEnumDeclaration']);
  });

  test('wildcard `*` → null (wildcard bucket)', () => {
    expect(anchorsOf('*')).toBeNull();
  });

  test('child combinator `A > B` → right side', () => {
    expect(anchorsOf('Program > FunctionDeclaration')).toEqual([
      'FunctionDeclaration',
    ]);
    // Right side is the matched node, not the parent — `Program` itself
    // is just the qualifier and should NOT appear in the anchor set.
    expect(anchorsOf('Program > FunctionDeclaration')).not.toContain('Program');
  });

  test('descendant combinator `A B` → right side', () => {
    expect(anchorsOf('FunctionDeclaration ReturnStatement')).toEqual([
      'ReturnStatement',
    ]);
  });

  test('adjacent sibling `A + B` → right side', () => {
    expect(anchorsOf('VariableDeclaration + ExpressionStatement')).toEqual([
      'ExpressionStatement',
    ]);
  });

  test('general sibling `A ~ B` → right side', () => {
    expect(anchorsOf('ImportDeclaration ~ ExportDeclaration')).toEqual([
      'ExportDeclaration',
    ]);
  });

  test(':matches(A, B) → union of all sub-anchors', () => {
    const anchors = anchorsOf(
      ':matches(TSInterfaceDeclaration, TSTypeAliasDeclaration)',
    );
    expect(new Set(anchors)).toEqual(
      new Set(['TSInterfaceDeclaration', 'TSTypeAliasDeclaration']),
    );
  });

  test(':matches with any wildcard sub → null', () => {
    // If even one branch of `:matches` is wildcard-typed, the whole
    // selector becomes wildcard (the union with "anything" is "anything").
    expect(anchorsOf(':matches(Identifier, *)')).toBeNull();
  });

  test(':not(B) standalone → null (negation matches any type)', () => {
    expect(anchorsOf(':not(FunctionDeclaration)')).toBeNull();
  });

  test('compound A:has(B) → anchor narrows to A', () => {
    // `:has` itself is wildcard, but the compound with `A` constrains
    // matching to only nodes of type A. Same as `A` alone for anchor
    // extraction purposes.
    expect(anchorsOf('FunctionDeclaration:has(ReturnStatement)')).toEqual([
      'FunctionDeclaration',
    ]);
  });

  test('compound A:not(B) → anchor narrows to A', () => {
    expect(anchorsOf('Identifier:not([name="self"])')).toEqual(['Identifier']);
  });

  test('compound with attribute filter → anchor narrows to type', () => {
    expect(anchorsOf('BinaryExpression[operator="+"]')).toEqual([
      'BinaryExpression',
    ]);
  });

  test('compound with multiple typed components → intersection', () => {
    // Pathological but valid: `A B` is not compound, but `A.label` and
    // similar use compound. For a literal compound of two types, the
    // intersection of their anchor sets is what matches both — typically
    // empty in practice, but the algorithm must compute it correctly.
    // `:matches(A, B):matches(B, C)` → intersection({A,B}, {B,C}) = {B}.
    expect(anchorsOf(':matches(A, B):matches(B, C)')).toEqual(['B']);
  });

  test(':function pseudo-class → 3 function-bearing types', () => {
    // ESLint v10's special-case: `:function` matches any function-form
    // node. Without this, rules using `:function` would silently fail
    // since esquery treats it as an unrecognized class otherwise.
    expect(new Set(anchorsOf(':function'))).toEqual(
      new Set([
        'FunctionDeclaration',
        'FunctionExpression',
        'ArrowFunctionExpression',
      ]),
    );
  });

  test('other pseudo-classes → null (wildcard)', () => {
    // Any `:xxx` pseudo-class that isn't `:function` falls through to
    // null so the selector is treated as wildcard — same as ESLint v10.
    expect(anchorsOf(':statement')).toBeNull();
  });

  test('attribute-only selector → null (wildcard)', () => {
    // `[name="foo"]` alone has no type identifier, so it could match
    // any node with the right attribute.
    expect(anchorsOf('[name="foo"]')).toBeNull();
  });

  test('nth-child pseudo → null', () => {
    // Structural pseudo without a type identifier → wildcard.
    expect(anchorsOf(':nth-child(2)')).toBeNull();
  });

  test('nested combinator `A > B > C` → deepest right side', () => {
    expect(anchorsOf('Program > BlockStatement > ReturnStatement')).toEqual([
      'ReturnStatement',
    ]);
  });

  test('nested matches in child `Program > :matches(A, B)` → union', () => {
    const anchors = anchorsOf(
      'Program > :matches(ClassDeclaration, TSEnumDeclaration)',
    );
    expect(new Set(anchors)).toEqual(
      new Set(['ClassDeclaration', 'TSEnumDeclaration']),
    );
  });
});

// ─────────────────────────────────────────────────────────────────────
// mergeListeners — bucketed dispatch table
// ─────────────────────────────────────────────────────────────────────

describe('mergeListeners — esqueryByType bucket assembly', () => {
  test('simple selector files into the by-type bucket as an isSimple entry', () => {
    // ESLint v10 unifies bare-type and esquery selectors into one
    // NodeEventGenerator dispatch so they specificity-sort against
    // each other. The runner mirrors this: `Identifier` produces a
    // by-type bucket entry with `isSimple: true` (no `esquery.matches`
    // is called at dispatch time — the bucket key already filtered to
    // matching nodes). Pre-fix bare-type selectors lived in a separate
    // `simple` map and always ran BEFORE every esquery selector,
    // inverting cross-rule observation order vs ESLint.
    const buckets = mergeAndGetBuckets({ Identifier: () => {} });
    const entries = buckets.enter.get('Identifier') ?? [];
    expect(entries).toHaveLength(1);
    expect(entries[0].raw).toBe('Identifier');
    expect(entries[0].isSimple).toBe(true);
    expect(entries[0].selector).toBeNull();
    expect(entries[0].attributeCount).toBe(0);
    expect(entries[0].identifierCount).toBe(1);
    expect(buckets.enterWildcard).toEqual([]);
  });

  test('child-combinator selector files under right-side type', () => {
    const buckets = mergeAndGetBuckets({
      'Program > FunctionDeclaration': () => {},
    });
    expect(buckets.enter.get('FunctionDeclaration')?.length).toBe(1);
    // Crucially, `Program` is NOT in the bucket — selector only fires
    // when the visitor reaches a FunctionDeclaration node.
    expect(buckets.enter.has('Program')).toBe(false);
    expect(buckets.enterWildcard).toEqual([]);
  });

  test(':matches(A, B) files under both A and B', () => {
    const buckets = mergeAndGetBuckets({
      ':matches(ClassDeclaration, FunctionDeclaration)': () => {},
    });
    expect(buckets.enter.get('ClassDeclaration')?.length).toBe(1);
    expect(buckets.enter.get('FunctionDeclaration')?.length).toBe(1);
    expect(buckets.enterWildcard).toEqual([]);
  });

  test(':not(B) standalone goes to wildcard', () => {
    const buckets = mergeAndGetBuckets({
      ':not(FunctionDeclaration)': () => {},
    });
    expect(buckets.enter.size).toBe(0);
    expect(buckets.enterWildcard.length).toBe(1);
  });

  test('A:not(B) compound narrows to A', () => {
    const buckets = mergeAndGetBuckets({
      'Identifier:not([name="self"])': () => {},
    });
    expect(buckets.enter.get('Identifier')?.length).toBe(1);
    expect(buckets.enterWildcard).toEqual([]);
  });

  test('attribute-only selector goes to wildcard', () => {
    const buckets = mergeAndGetBuckets({ '[name="foo"]': () => {} });
    expect(buckets.enter.size).toBe(0);
    expect(buckets.enterWildcard.length).toBe(1);
  });

  test(':exit selectors land in the exit bucket, not enter', () => {
    const buckets = mergeAndGetBuckets({
      'Program > FunctionDeclaration:exit': () => {},
    });
    expect(buckets.enter.size).toBe(0);
    expect(buckets.exit.get('FunctionDeclaration')?.length).toBe(1);
    expect(buckets.enterWildcard).toEqual([]);
    expect(buckets.exitWildcard).toEqual([]);
  });

  test('multiple selectors targeting the same anchor share one bucket entry list', () => {
    const buckets = mergeAndGetBuckets({
      'Program > FunctionDeclaration': () => {},
      'Module > FunctionDeclaration': () => {},
    });
    // Both file under FunctionDeclaration — the bucket has 2 entries.
    expect(buckets.enter.get('FunctionDeclaration')?.length).toBe(2);
    // No other anchor types pulled in.
    expect(buckets.enter.size).toBe(1);
  });

  test(':function pseudo-class fans out to all 3 function types', () => {
    const buckets = mergeAndGetBuckets({ ':function': () => {} });
    // Each of the 3 function types has its own bucket entry pointing
    // at the same listener.
    expect(buckets.enter.get('FunctionDeclaration')?.length).toBe(1);
    expect(buckets.enter.get('FunctionExpression')?.length).toBe(1);
    expect(buckets.enter.get('ArrowFunctionExpression')?.length).toBe(1);
    expect(buckets.enterWildcard).toEqual([]);
  });
});

// ─────────────────────────────────────────────────────────────────────
// visit() no-ESQuery fast path
// ─────────────────────────────────────────────────────────────────────

describe('visit — no-ESQuery fast path', () => {
  // Synthetic AST: Program { body: [ExpressionStatement { expression: Identifier "x" }] }
  // We use a literal object tree because visit() only looks at `type` and
  // visitor-key children; we don't need a real parse.
  // Synthetic nodes — we deliberately omit `range`/`loc` because visit()
  // does not read them; only `type` + visitor-key children matter. Cast
  // through `unknown` to satisfy the stricter ESTreeNode interface used
  // by visit's signature without forcing a real parse.
  const x = { type: 'Identifier', name: 'x', start: 0, end: 1 };
  const expr = {
    type: 'ExpressionStatement',
    expression: x,
    start: 0,
    end: 2,
  };
  const program = {
    type: 'Program',
    body: [expr],
    start: 0,
    end: 2,
  } as unknown as ESTreeNode;

  test('simple-only listeners still fire on matching nodes', () => {
    // Post-M4: bare-type listeners file into `esqueryByType.enter` as
    // `isSimple: true` entries — `fireEnter` skips `esquery.matches`
    // for them (the bucket key already filtered to matching node
    // types), so dispatch cost is the same as the pre-refactor
    // separate `simple` map but ordering is unified with esquery.
    const seen: string[] = [];
    const merged = mergeListeners([
      {
        ruleName: 'test',
        listeners: {
          Program: () => seen.push('Program'),
          Identifier: () => seen.push('Identifier'),
        },
      },
    ]);
    // Bare-type entries land in the by-type bucket (one per type).
    expect(merged.esqueryByType.enter.size).toBe(2);
    expect(merged.esqueryByType.enter.get('Program')?.[0].isSimple).toBe(true);
    expect(merged.esqueryByType.enter.get('Identifier')?.[0].isSimple).toBe(
      true,
    );
    expect(merged.esqueryByType.enterWildcard.length).toBe(0);
    const result = visit(program, merged);
    expect(result.cancelled).toBe(false);
    expect(seen).toEqual(['Program', 'Identifier']);
  });

  test('esquery selector still resolves ancestor combinators when registered', () => {
    // Sanity: when an esquery selector IS present, the non-fast-path
    // branch runs and ancestry is maintained — the `Program > Identifier`
    // child combinator must still fire on the descendant Identifier.
    const seen: string[] = [];
    const merged = mergeListeners([
      {
        ruleName: 'test',
        listeners: {
          // Child combinator: requires ancestry to evaluate correctly.
          // Identifier is two levels deep so this should NOT fire
          // (Program > ExpressionStatement > Identifier is the structure).
          'Program > Identifier': () => seen.push('direct-child-of-program'),
          'Program Identifier': () => seen.push('descendant-of-program'),
        },
      },
    ]);
    expect(merged.esqueryByType.enter.size).toBeGreaterThan(0);
    visit(program, merged);
    // The Identifier is NOT a direct child of Program (it's nested under
    // ExpressionStatement.expression), so `>` should not fire — but the
    // descendant combinator should. This proves ancestry is being
    // maintained correctly in the non-fast-path.
    expect(seen).toEqual(['descendant-of-program']);
  });

  test('mixed simple + esquery: ancestry maintained, simple fires too', () => {
    const seen: string[] = [];
    const merged = mergeListeners([
      {
        ruleName: 'test',
        listeners: {
          Program: () => seen.push('Program'),
          'Program > ExpressionStatement': () => seen.push('direct-expr-stmt'),
        },
      },
    ]);
    expect(merged.esqueryByType.enter.size).toBeGreaterThan(0);
    visit(program, merged);
    expect(seen).toContain('Program');
    expect(seen).toContain('direct-expr-stmt');
  });
});

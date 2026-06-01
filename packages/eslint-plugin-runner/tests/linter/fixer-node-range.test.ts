/**
 * Regression test for review P1 #11.
 *
 * `RuleFixer` node-API methods (`replaceText` / `insertTextBefore` /
 * `insertTextAfter` / `remove`) used to read `node.range` without
 * checking whether the field was actually present. Plugins that
 * construct synthetic AST nodes (helpers, codemod-style migrations,
 * `react-internal`-style extension nodes) and hand them to the fixer
 * would trigger `copyRange(undefined)` → generic `TypeError` → the
 * fixer's outer try/catch in `diagnostic-builder` absorbed the throw
 * into a single `console.error` line and dropped the fix on the
 * floor. User running `--fix` saw no edit and had no recoverable
 * signal pointing at the rule.
 *
 * Post-fix: each method calls `assertNodeRange(node, methodName)`
 * which throws a clearly attributed `TypeError` referencing the
 * fixer method. The wrapping try/catch then surfaces the throw
 * normally — for the runtime path it still ends in the rule-error
 * channel, but the message is now actionable.
 */
import { describe, test, expect } from '@rstest/core';

import { makeFixer } from '../../src/linter/fixer.js';

describe('RuleFixer: guards against synthetic node without range', () => {
  const fixer = makeFixer();

  test('replaceText on a node without range throws a clearly-attributed TypeError', () => {
    const synthetic = { type: 'Identifier', name: 'foo' } as never;
    expect(() => fixer.replaceText(synthetic, 'bar')).toThrow(TypeError);
    try {
      fixer.replaceText(synthetic, 'bar');
    } catch (err) {
      expect((err as Error).message).toContain('replaceText');
      expect((err as Error).message).toContain('range');
    }
  });

  test('insertTextBefore on a node without range throws', () => {
    const synthetic = { type: 'CallExpression' } as never;
    expect(() => fixer.insertTextBefore(synthetic, 'x')).toThrow(TypeError);
  });

  test('insertTextAfter on a node without range throws', () => {
    const synthetic = { type: 'CallExpression' } as never;
    expect(() => fixer.insertTextAfter(synthetic, 'x')).toThrow(TypeError);
  });

  test('remove on a node without range throws', () => {
    const synthetic = { type: 'CallExpression' } as never;
    expect(() => fixer.remove(synthetic)).toThrow(TypeError);
  });

  test('replaceText on a node WITH valid range works (no regression)', () => {
    const node = { type: 'Identifier', range: [5, 10] as const } as never;
    const fix = fixer.replaceText(node, 'baz');
    expect(fix.range).toEqual([5, 10]);
    expect(fix.text).toBe('baz');
  });

  test('range-API variants are unaffected by the guard (range passed directly)', () => {
    // The *-Range methods don't go through node.range, so they
    // should keep working even with synthetic / inferred ranges.
    expect(fixer.replaceTextRange([1, 3], 'ok')).toEqual({
      range: [1, 3],
      text: 'ok',
    });
    expect(fixer.removeRange([4, 7])).toEqual({ range: [4, 7], text: '' });
    expect(fixer.insertTextBeforeRange([2, 5], '/* */')).toEqual({
      range: [2, 2],
      text: '/* */',
    });
    expect(fixer.insertTextAfterRange([2, 5], '/* */')).toEqual({
      range: [5, 5],
      text: '/* */',
    });
  });
});

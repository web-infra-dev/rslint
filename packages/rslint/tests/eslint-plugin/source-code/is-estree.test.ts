/**
 * `SourceCode#isESTree` mirrors ESLint v10's
 *   `this.isESTree = (ast.type === "Program")`
 * (eslint/lib/languages/js/source-code/source-code.js:316).
 *
 * The property exists so plugin rules that guard on it behave under rslint as
 * they do under ESLint. Stylistic's `indent` rule, for instance, bails with
 * `if (!isESTreeSourceCode(...)) return {}` — without this property it would
 * register zero listeners and silently report nothing. These tests pin the
 * exact predicate (the Program / non-Program split), not just a truthy value.
 */
import { describe, test, expect } from '@rstest/core';

import { createSourceCode } from '../../../src/eslint-plugin/source-code/source-code.js';
import type { ESTreeNode } from '../../../src/eslint-plugin/source-code/source-code.js';

/** Minimal root node — only `type`/`range`/`loc` matter for `isESTree`. */
function mkRoot(type: string, text: string): ESTreeNode {
  return {
    type,
    range: [0, text.length],
    loc: {
      start: { line: 1, column: 0 },
      end: { line: 1, column: text.length },
    },
    start: 0,
    end: text.length,
  };
}

describe('SourceCode#isESTree', () => {
  test('is true when the AST root is a Program (ESTree-backed)', () => {
    const text = 'const x = 1;';
    const sc = createSourceCode({
      text,
      ast: mkRoot('Program', text),
      scopeManagerFactory: () => ({}),
    });
    expect(sc.isESTree).toBe(true);
  });

  test('is false when the AST root is not a Program', () => {
    const text = '{}';
    const sc = createSourceCode({
      text,
      ast: mkRoot('JSONProgram', text),
      scopeManagerFactory: () => ({}),
    });
    expect(sc.isESTree).toBe(false);
  });
});

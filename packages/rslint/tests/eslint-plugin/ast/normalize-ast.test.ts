/**
 * normalize-ast — TS shape adjustments + parent linking.
 *
 * oxc-parser 0.132 emits `TSEnumDeclaration` ALREADY in the
 * typescript-eslint v8+ shape: members live inside
 * `body: { type: 'TSEnumBody', members: [...] }` and there is NO
 * top-level `members` field on the declaration. normalizeAst does not
 * transform enums — it only adds `parent` links and (for nodes lacking
 * them) `loc` while walking. The tests below pin oxc's native enum shape
 * that the runner consumes directly, so a future oxc version that drifts
 * (reverting to a flat `members`, changing the body range, or renaming
 * the wrapper) fails here and forces a review.
 */

import { describe, test, expect } from '@rstest/core';
import { parse as nativeParse } from '../../../src/eslint-plugin/native/load-binding.js';

import {
  normalizeAst,
  buildLineStartOffsets,
  buildUtf16ToByteMap,
} from '../../../src/eslint-plugin/ast/normalize-ast.js';
import type { ESTreeNode } from '../../../src/eslint-plugin/linter/context.js';

function parseAndNormalize(filePath: string, text: string): ESTreeNode {
  // Use the native (napi) parser — the same one the runtime uses — so these
  // tests exercise the real AST the runner normalizes, not npm oxc-parser.
  const jsx = /\.[jt]sx$/.test(filePath);
  const parsed = nativeParse(filePath, text, 'module', jsx);
  const ast = JSON.parse(parsed.program) as ESTreeNode;
  const lso = buildLineStartOffsets(text);
  normalizeAst(ast, lso, text);
  return ast;
}

describe('oxc-parser native TSEnumDeclaration shape', () => {
  test('members live in body.members; no top-level members field', () => {
    const ast = parseAndNormalize(
      'enum.ts',
      'enum Color { Red, Green, Blue }\n',
    );

    // Locate the TSEnumDeclaration. Programs always have `body` as an
    // array of top-level statements.
    const program = ast as ESTreeNode & { body: ESTreeNode[] };
    const enumDecl = program.body.find(
      (n) => n.type === 'TSEnumDeclaration',
    ) as (ESTreeNode & { body?: ESTreeNode; members?: unknown }) | undefined;
    expect(enumDecl).toBeDefined();

    // oxc emits the members under a `body: TSEnumBody` wrapper.
    expect(enumDecl?.body).toBeDefined();
    expect((enumDecl?.body as ESTreeNode).type).toBe('TSEnumBody');
    const bodyAsUnknown = enumDecl?.body as unknown;
    const body = bodyAsUnknown as { members: unknown[] };
    expect(Array.isArray(body.members)).toBe(true);
    expect(body.members.length).toBe(3);

    // oxc does NOT also put a flat `members` array on the declaration —
    // so the walker visits each member exactly once (via body.members)
    // with the correct parent. A future oxc that reintroduces a flat
    // `members` would fail here.
    expect('members' in enumDecl!).toBe(false);
  });

  // oxc emits TSEnumBody.range as strictly the `{ ... }` portion — NOT
  // the whole declaration (which would include the `enum Foo ` keyword +
  // identifier prefix). Tools that read `body.range` to highlight the
  // body — codemods, formatting rules, source-text slicing in
  // conformance harnesses — rely on this matching typescript-eslint v8,
  // which also defines TSEnumBody as strictly the braces and contents.
  test('TSEnumBody.range covers the {...} portion only, not the keyword/id', () => {
    const text = 'enum Color { Red, Green, Blue }';
    const ast = parseAndNormalize('enum.ts', text);
    const program = ast as ESTreeNode & { body: ESTreeNode[] };
    const enumDecl = program.body.find(
      (n) => n.type === 'TSEnumDeclaration',
    ) as unknown as {
      range: [number, number];
      body: {
        type: string;
        range: [number, number];
        loc: {
          start: { line: number; column: number };
          end: { line: number; column: number };
        };
      };
    };
    expect(enumDecl).toBeDefined();

    const bodyRange = enumDecl.body.range;
    // `{` is at index 11 (`enum Color ` is 11 chars), `}` is the last
    // char so range[1] === text.length.
    expect(text[bodyRange[0]]).toBe('{');
    expect(text[bodyRange[1] - 1]).toBe('}');
    expect(bodyRange[0]).toBe(text.indexOf('{'));
    expect(bodyRange[1]).toBe(text.length);

    // loc must mirror range (1-based line, 0-based column).
    expect(enumDecl.body.loc.start.column).toBe(text.indexOf('{'));
  });

  test('TSEnumBody.range survives a multi-line enum with leading comment', () => {
    const text = '// comment\nexport enum Bigger {\n  A,\n  B,\n  C,\n}\n';
    const ast = parseAndNormalize('enum.ts', text);
    const program = ast as ESTreeNode & { body: ESTreeNode[] };
    // Wrapped in ExportNamedDeclaration; descend one level.
    const exportDecl = program.body.find(
      (n) => n.type === 'ExportNamedDeclaration',
    ) as unknown as {
      declaration: { type: string; body: { range: [number, number] } };
    };
    expect(exportDecl).toBeDefined();
    expect(exportDecl.declaration.type).toBe('TSEnumDeclaration');
    const bodyRange = exportDecl.declaration.body.range;
    expect(text[bodyRange[0]]).toBe('{');
    expect(text[bodyRange[1] - 1]).toBe('}');
  });

  test('TSEnumBody.range is valid for an empty enum body', () => {
    // Empty enum body — oxc still emits a TSEnumBody whose range is the
    // `{}` span. Pins that the degenerate (zero-member) case keeps a
    // valid brace-to-brace range.
    const text = 'enum X {}';
    const ast = parseAndNormalize('enum.ts', text);
    const program = ast as ESTreeNode & { body: ESTreeNode[] };
    const enumDecl = program.body.find(
      (n) => n.type === 'TSEnumDeclaration',
    ) as unknown as { body: { range: [number, number] } };
    expect(enumDecl).toBeDefined();
    expect(text[enumDecl.body.range[0]]).toBe('{');
    expect(text[enumDecl.body.range[1] - 1]).toBe('}');
  });

  test('enum member parent points at TSEnumBody (single-visit walk)', () => {
    // The walker sets `parent` during DFS. Because oxc puts members only
    // under `body.members` (no flat `members`), each member is visited
    // exactly once and its parent is the TSEnumBody wrapper —
    // typescript-eslint v8+ semantics.
    const ast = parseAndNormalize(
      'enum.ts',
      'enum Status { Active = "A", Inactive = "I" }\n',
    );
    const program = ast as ESTreeNode & { body: ESTreeNode[] };
    const enumDecl = program.body.find(
      (n) => n.type === 'TSEnumDeclaration',
    ) as unknown as
      | { body: { type: string; members: ESTreeNode[] } }
      | undefined;
    expect(enumDecl).toBeDefined();

    const body = enumDecl!.body;
    expect(body.type).toBe('TSEnumBody');
    const firstMember = body.members[0];
    expect(firstMember).toBeDefined();
    expect(
      (firstMember as ESTreeNode & { parent?: ESTreeNode }).parent?.type,
    ).toBe('TSEnumBody');
  });
});

// R2 regression. The runner internally uses UTF-16 code-unit offsets
// (matching JS string indexing + oxc-parser output), but the Go side
// passes the wire `startPos` to `scanner.GetECMALineAnd
// UTF16CharacterOfPosition`, whose `pos` parameter is UTF-8 BYTES.
// Without conversion at the IPC boundary, any file containing a
// multi-byte UTF-8 character (CJK, emoji, the `➜` symbol, etc.)
// reports diagnostic columns shifted back by N for every following
// diagnostic, where N = bytes_consumed_by_multi_byte_chars
// - utf16_units_consumed_by_them.
describe('buildUtf16ToByteMap (R2 UTF-16 → UTF-8 byte conversion)', () => {
  test('ASCII-only is identity', () => {
    const text = 'const x = 1;';
    const map = buildUtf16ToByteMap(text);
    expect(map[0]).toBe(0);
    expect(map[6]).toBe(6); // ` ` after `const`
    expect(map[text.length]).toBe(text.length);
  });

  test('3-byte UTF-8 char (➜ U+279C) shifts byte offsets by 2', () => {
    const text = '➜ab';
    const map = buildUtf16ToByteMap(text);
    expect(map[0]).toBe(0); // ➜ starts at byte 0
    expect(map[1]).toBe(3); // a starts at byte 3 (➜ is 3 bytes)
    expect(map[2]).toBe(4); // b
    expect(map[3]).toBe(5); // end sentinel
  });

  test('CJK chars (3 bytes each)', () => {
    const text = '中文a';
    const map = buildUtf16ToByteMap(text);
    expect(map[0]).toBe(0);
    expect(map[1]).toBe(3);
    expect(map[2]).toBe(6); // a
    expect(map[3]).toBe(7); // end
  });

  test('surrogate pair (emoji, 4 UTF-8 bytes) maps both halves to start byte', () => {
    const text = '😀ab'; // 😀 U+1F600 = surrogate pair in UTF-16
    expect(text.length).toBe(4); // 2 surrogate units + 2 ASCII
    const map = buildUtf16ToByteMap(text);
    expect(map[0]).toBe(0); // high surrogate → byte 0
    expect(map[1]).toBe(0); // low surrogate → also byte 0 (mid of 4-byte seq)
    expect(map[2]).toBe(4); // a
    expect(map[3]).toBe(5); // b
    expect(map[4]).toBe(6); // end
  });

  test('matches manual TextEncoder count for arbitrary mixed text', () => {
    const text = '// ➜ 中文 😀\nconst x = "草";';
    const map = buildUtf16ToByteMap(text);
    // End sentinel must equal the UTF-8 byte length.
    expect(map[text.length]).toBe(new TextEncoder().encode(text).length);
    // Every map[i] must equal the UTF-8 byte length of `text.slice(0, i)`.
    // Spot-check a handful of indices.
    for (const i of [0, 3, 5, 8, text.length]) {
      const sliceBytes = new TextEncoder().encode(text.slice(0, i)).length;
      expect(map[i]).toBe(sliceBytes);
    }
  });
});

describe('R2: oxc UTF-16 offset → Go-friendly UTF-8 byte offset', () => {
  test('diagnostic startPos is in UTF-8 bytes after lintFile drains', async () => {
    // We exercise the runner end-to-end (lintFile path) to confirm
    // the byte conversion happens at the IPC boundary, not just in
    // the helper. Source: 2 leading lines with `➜` (3 UTF-8 bytes
    // each, 1 utf-16 unit each) so byte/utf-16 indices diverge by 4
    // before line 3.
    const text = '// ➜\n// ➜\n    let z = null;\n';
    const { lintFile } =
      await import('../../../src/eslint-plugin/linter/ecma-language-plugin.js');
    // Stub unicorn/no-null in a hand-rolled plugin map. The rule
    // implementation is the simplest possible: report on every
    // `Literal` whose `value` is null.
    const stub = {
      meta: { name: 'stub/no-null' },
      create(
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        ctx: any,
      ) {
        return {
          Literal(node: { value: unknown }) {
            if (node.value === null) {
              ctx.report({ node, message: 'no null' });
            }
          },
        };
      },
    };
    const result = lintFile(
      {
        filePath: 'enc.ts',
        text,
        rules: { 'stub/no-null': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      {
        plugins: [],
        rules: new Map<string, unknown>([['stub/no-null', stub]]),
      },
    );

    expect(result.diagnostics.length).toBe(1);
    const d = result.diagnostics[0];
    // The `null` literal sits on line 3 at utf-16 col 12 (1-based) /
    // byte col 12 within line 3 (no multi-byte chars on line 3
    // itself). What we care about: startPos points at UTF-8 byte
    // offset of `n` of `null`, NOT the UTF-16 idx.
    //
    // Compute expected byte offset manually:
    //   `// ➜\n` = `/`+`/`+` `+`➜`(3)+`\n` = 7 bytes
    //   `// ➜\n` again = 7 bytes (total 14)
    //   `    let z = ` = 12 bytes (4 spaces + `let z = `)
    //   total = 26 bytes → byte offset of `n` of null
    const expectedByte = 26;
    expect(d.startPos).toBe(expectedByte);
    expect(d.endPos).toBe(expectedByte + 4); // null = 4 bytes ASCII

    // UTF-16 alternative would have been 22 (4 fewer than 26). Pin
    // that the runner ISN'T reporting that.
    expect(d.startPos).not.toBe(22);
  });
});

// ── buildLineStartOffsets — every ECMAScript LineTerminator ──────
//
// Spec § 11.3 LineTerminator: LF, CR, CRLF (one terminator), LS, PS.
// Pre-fix this function only split on LF, so files using any of the
// other terminators got wrong `node.loc.line` for every node after
// the missed break — and disagreed with `sourceCode.getLocFromIndex`
// (which used a separate fuller implementation). Both surfaces now
// share this one implementation; the tests below pin every
// terminator family.

describe('buildLineStartOffsets — ECMAScript line terminators', () => {
  test('LF alone produces one offset per line', () => {
    expect(buildLineStartOffsets('a\nb\nc')).toEqual([0, 2, 4]);
  });

  test('CRLF counts as ONE terminator (not two)', () => {
    expect(buildLineStartOffsets('a\r\nb\r\nc')).toEqual([0, 3, 6]);
  });

  test('bare CR (legacy Mac files) counts as a terminator', () => {
    expect(buildLineStartOffsets('a\rb\rc')).toEqual([0, 2, 4]);
  });

  test('U+2028 (LINE SEPARATOR) counts as a terminator', () => {
    expect(buildLineStartOffsets('a b')).toEqual([0, 2]);
  });

  test('U+2029 (PARAGRAPH SEPARATOR) counts as a terminator', () => {
    expect(buildLineStartOffsets('a b')).toEqual([0, 2]);
  });

  test('mixed terminators in one file all contribute offsets', () => {
    const offsets = buildLineStartOffsets('a\nb\rc\r\nd e f');
    // 6 lines: starts at 0, after a-LF, after b-CR, after c-CRLF,
    // after d-LS, after e-PS.
    expect(offsets).toEqual([0, 2, 4, 7, 9, 11]);
  });

  test('source-code re-export points at the same implementation', async () => {
    const { buildLineStartOffsetsLocal } =
      await import('../../../src/eslint-plugin/source-code/source-code.js');
    for (const sample of [
      'a',
      'a\nb',
      'a\r\nb',
      'a\rb',
      'a b',
      'a b',
      'a\nb\rc\r\nd e f',
    ]) {
      expect(buildLineStartOffsetsLocal(sample)).toEqual(
        buildLineStartOffsets(sample),
      );
    }
  });
});

// ─────────────────────────────────────────────────────────────────────
// normalize derives loc — the native parser emits neither range nor loc.
// ─────────────────────────────────────────────────────────────────────

describe('normalize-ast — derives loc (the native parser emits neither range nor loc)', () => {
  test('normalize adds loc with correct line/column', () => {
    const text = 'const x = 1;\nconst y = 2;';
    const parsed = nativeParse('a.ts', text, 'module', false);
    const ast = JSON.parse(parsed.program) as ESTreeNode;
    const decl = (ast as unknown as { body: Array<Record<string, unknown>> })
      .body[1];
    // The native parser emits neither range nor loc — normalize derives both.
    expect(decl.loc).toBeUndefined();
    const lso = buildLineStartOffsets(text);
    normalizeAst(ast, lso, text);
    expect(decl.loc).toEqual({
      start: { line: 2, column: 0 },
      end: { line: 2, column: 12 },
    });
  });
});

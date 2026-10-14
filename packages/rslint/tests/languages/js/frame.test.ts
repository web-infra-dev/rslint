/** Final ESTree shape, positions, and hostile frame validation. */

import { describe, test, expect } from 'rstack/test';
import { parse as parseUpstream } from '@typescript-eslint/parser';
import {
  decodeFrame,
  FrameCancelled,
} from '../../../src/languages/js/frame.js';
import { FRAME_SCHEMAS } from '../../../src/languages/js/frame-schema.js';
import { VISITOR_KEYS } from '../../../src/languages/js/visitor-keys.js';
import { lintFile as lintFrame } from '../../../src/eslint-plugin/linter/ecma-language-plugin.js';
import { projectFrame, projectSyntax } from './project-frame.js';

import {
  buildLineStartOffsets,
  buildUtf16ToByteMap,
} from '../../../src/languages/js/positions.js';
import type { ESTreeNode } from '../../../src/languages/js/types.js';
import type { RuleContext } from '../../../src/eslint-plugin/linter/context.js';

function projectAST(filePath: string, text: string): ESTreeNode {
  const parsed = projectSyntax(filePath, text, 'module');
  const ast = parsed.ast as ESTreeNode;
  return ast;
}

describe('Go-projected TSEnumDeclaration shape', () => {
  test('members live in body.members; no top-level members field', () => {
    const ast = projectAST('enum.ts', 'enum Color { Red, Green, Blue }\n');

    // Locate the TSEnumDeclaration. Programs always have `body` as an
    // array of top-level statements.
    const program = ast as ESTreeNode & { body: ESTreeNode[] };
    const enumDecl = program.body.find(
      (n) => n.type === 'TSEnumDeclaration',
    ) as (ESTreeNode & { body?: ESTreeNode; members?: unknown }) | undefined;
    expect(enumDecl).toBeDefined();

    // ESTree exposes the members under a `body: TSEnumBody` wrapper.
    expect(enumDecl?.body).toBeDefined();
    expect((enumDecl?.body as ESTreeNode).type).toBe('TSEnumBody');
    const bodyAsUnknown = enumDecl?.body as unknown;
    const body = bodyAsUnknown as { members: unknown[] };
    expect(Array.isArray(body.members)).toBe(true);
    expect(body.members.length).toBe(3);

    // ESTree does not also put a flat `members` array on the declaration —
    // so the walker visits each member exactly once (via body.members)
    // with the correct parent. A future projector that reintroduces a flat
    // `members` would fail here.
    expect('members' in enumDecl!).toBe(false);
  });

  // ESTree exposes TSEnumBody.range as strictly the `{ ... }` portion — NOT
  // the whole declaration (which would include the `enum Foo ` keyword +
  // identifier prefix). Tools that read `body.range` to highlight the
  // body — codemods, formatting rules, source-text slicing in
  // conformance harnesses — rely on this matching typescript-eslint v8,
  // which also defines TSEnumBody as strictly the braces and contents.
  test('TSEnumBody.range covers the {...} portion only, not the keyword/id', () => {
    const text = 'enum Color { Red, Green, Blue }';
    const ast = projectAST('enum.ts', text);
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
    const ast = projectAST('enum.ts', text);
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
    // Empty enum body — ESTree still exposes a TSEnumBody whose range is the
    // `{}` span. Pins that the degenerate (zero-member) case keeps a
    // valid brace-to-brace range.
    const text = 'enum X {}';
    const ast = projectAST('enum.ts', text);
    const program = ast as ESTreeNode & { body: ESTreeNode[] };
    const enumDecl = program.body.find(
      (n) => n.type === 'TSEnumDeclaration',
    ) as unknown as { body: { range: [number, number] } };
    expect(enumDecl).toBeDefined();
    expect(text[enumDecl.body.range[0]]).toBe('{');
    expect(text[enumDecl.body.range[1] - 1]).toBe('}');
  });

  test('enum member parent points at TSEnumBody (single-visit walk)', () => {
    // The walker sets `parent` during DFS. Because ESTree puts members only
    // under `body.members` (no flat `members`), each member is visited
    // exactly once and its parent is the TSEnumBody wrapper —
    // typescript-eslint v8+ semantics.
    const ast = projectAST(
      'enum.ts',
      'enum Status { Active = "A", Inactive = "I" }\n',
    );
    const program = ast as ESTreeNode & { body: ESTreeNode[] };
    const enumDecl = program.body.find(
      (n) => n.type === 'TSEnumDeclaration',
    ) as unknown as
      { body: { type: string; members: ESTreeNode[] } } | undefined;
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
// (matching JS string indexing + ESTree output), but the Go side
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

describe('R2: ESTree UTF-16 offset → Go-friendly UTF-8 byte offset', () => {
  test('diagnostic startPos is in UTF-8 bytes after lintFile drains', async () => {
    // We exercise the runner end-to-end (lintFile path) to confirm
    // the byte conversion happens at the IPC boundary, not just in
    // the helper. Source: 2 leading lines with `➜` (3 UTF-8 bytes
    // each, 1 utf-16 unit each) so byte/utf-16 indices diverge by 4
    // before line 3.
    const text = '// ➜\n// ➜\n    let z = null;\n';
    const { lintFile } = await import('../../eslint-plugin/project-frame.js');
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

  test('SourceCode and decoded nodes share line positions across terminators', async () => {
    const { createSourceCode } =
      await import('../../../src/eslint-plugin/source-code/source-code.js');
    for (const terminator of ['\n', '\r\n', '\r', '\u2028', '\u2029']) {
      const text = `const first = 1;${terminator}const next = 2;`;
      const decoded = projectSyntax('lines.js', text);
      const sourceCode = createSourceCode({
        text: decoded.text,
        ast: decoded.ast,
        hasBOM: decoded.hasBOM,
        scopeManagerFactory: () => null,
      });
      const next = (decoded.ast.body as ESTreeNode[])[1];
      expect(next.loc.start).toEqual({ line: 2, column: 0 });
      expect(sourceCode.getLocFromIndex(next.range[0])).toEqual(next.loc.start);
      expect(sourceCode.getIndexFromLoc({ line: 2, column: 0 })).toBe(
        next.range[0],
      );
    }
  });
});

// ─────────────────────────────────────────────────────────────────────
// The decoder derives loc from projected source ranges.
// ─────────────────────────────────────────────────────────────────────

describe('frame decoder — derives loc', () => {
  test('decoder adds loc with correct line/column', () => {
    const text = 'const x = 1;\nconst y = 2;';
    const parsed = projectSyntax('a.ts', text, 'module');
    const ast = parsed.ast as ESTreeNode;
    const decl = (ast as unknown as { body: Array<Record<string, unknown>> })
      .body[1];
    // The decoder materializes range and loc from the columns.
    expect(decl.range).toEqual([13, 25]);
    const lso = buildLineStartOffsets(text);

    expect(decl.loc).toEqual({
      start: { line: 2, column: 0 },
      end: { line: 2, column: 12 },
    });
  });
});

// Compare all rule-visible child structure, ranges and scalar values with the
// canonical TypeScript ESTree contract. Deprecated parser aliases are excluded.
function canonical(value: unknown): unknown {
  if (value == null) return null;
  if (value instanceof RegExp) return String(value);
  if (typeof value === 'bigint') return `${value}n`;
  if (Array.isArray(value)) return value.map(canonical);
  if (typeof value !== 'object') return value;
  const node = value as Record<string, unknown>;
  const result: Record<string, unknown> = {};
  const type = typeof node.type === 'string' ? node.type : undefined;
  const fields = type
    ? FRAME_SCHEMAS.find(([name]) => name === type)?.[1]
    : undefined;
  for (const key of fields ? ['type', 'range', ...fields] : Object.keys(node)) {
    if (key.startsWith('$')) continue;
    result[key] = canonical(node[key]);
  }
  if (type === 'Literal' && node.regex) result.regex = canonical(node.regex);
  if (type === 'TemplateElement') result.value = canonical(node.value);
  return result;
}

describe('Go projection matches canonical ESTree', () => {
  const cases = [
    'const {source: local = 1, [key]: other = 2, nested: {value} = fallback} = input; function f({source: local = 1}) {} const [item = 1] = list;',
    '// 😀\nclass C { private /* constructor in trivia */ constructor /* ( */ () {} } class D { protected\nconstructor() {} }',
    'class C { private constructor(x: number); private constructor(x: number) {} }',
    String.raw`class C { "constructor"() {} } class D { 'constr\u0075ctor'() {} }`,
    'class C { static constructor() {} static "constructor"() {} }',
    'function f(...xs: string[]) {} function g(x?: number, y: number = 1) {} class C { constructor(public readonly x: number = 1) {} }',
    'abstract class C<T> extends Base<T> implements ns.I<T> { abstract f(x?: T): T; accessor x = 1; readonly y!: T; constructor(public a: T) {} static { init() } get z() { return this.y; } set z(v:T) {} }',
    'export interface I<T extends string = string> extends ns.Base<T> { readonly x?: T; f<U>(x: U): T; [key: string]: T; new(x: T): I<T> } type A<T> = { -readonly [K in keyof T as K]-?: T[K] }; type B = (x: number) => string; type C = [a: string, ...rest: number[], b?: string]; type D<T> = T extends infer U ? U : never; type E = (string | null); type G = new<T>(x:T) => I<T>;',
    "import type { Foo as F } from 'a'; export type { F }; export * as ns from 'a'; export as namespace N; import X = require('x'); export = X; namespace A.B { export const x = 1 } enum E { A, B = 2 } declare module 'a' {} declare global { interface A {} }",
    "import x, * as y from 'a'; import {default as a, b as c, type D} from 'b' with {type:'json'}; export {c as d}; export {a} from 'b'; const p = import('b', {with:{type:'json'}}); type T = typeof import('a').X; type U = import('a', {with:{'resolution-mode':'import'},}).X<string>;",
    'using a = f(); await using b = g(); class C { #x = 1; static accessor y: number = 1; f<T>(x:T):T { return x } } const x = f<string>; const y = a?.b!.c?.<number>(1);',
    'const {a, b = 1, c: d, ...rest}: T = obj; const [a,,...b]:T = values; ({x = 1, y:z} = obj); for ({x} of values) {} const f = ({x}:T = v, ...rest: T[]) => x;',
    "'use strict'; const x = (a?.b).c; const o = { f() {}, get x() {}, set x(v) {}, async *g() {} }; 'late'; if (a) { 'not strict'; } function f(){ '\\x75se strict'; return 1; }",
    "const x = `a${ `b${c}` }d`; tag`\\unicode`; const r = /a[\\/]/giu; const n = 0x20000000000001; const b = 0x10n; const s = '\\uD800'; type T = `foo-${number}`;",
    'export default interface I {} export type T = number; function f(x?) {} type F = (x: /* comment */ number) => /* comment */ string; class C { f /* < ( */ <T /* comment */ >(x:T):T { return x } }',
    'function f(x: unknown): x is string { return true } function g(x:unknown): asserts x is string {} function h(x): asserts x {} interface I { f(x: unknown): x is string }',
    'declare function f<T>(x:T):T; function f(x:number):number; function f(x:number){ return x } declare abstract class C { abstract accessor foo:string; abstract get bar(): string; abstract set bar(v: string); private x?: number; override f() {} } type U<T> = { [K in keyof T]+?: T[K] };',
    '@foo export class C { @dec x = 1; constructor(@dec public readonly p: number) {} @dec f(@dec x: number) {} } export default abstract class A {} declare namespace A.B.C { export interface D {} }',
    'type A = unique symbol; type B = readonly [a?: string, ...b: number[]]; type C = typeof ns.x<string>; type D = { (x:number): void; <T>(x:T):T; readonly [x: string]: T; get x(): string; set x(v: string); }; type E = this;',
    "import defer * as ns from 'a'; const x = import('a');",
  ];
  for (const [index, text] of cases.entries()) {
    test(`syntax matrix ${index + 1}`, () => {
      const actual = projectSyntax('matrix.ts', text).ast;
      const expected = parseUpstream(text, {
        filePath: 'matrix.ts',
        sourceType: 'module',
        range: true,
        loc: true,
      });
      expect(canonical(actual)).toEqual(canonical(expected));
      const stack = [actual];
      while (stack.length) {
        const node = stack.pop()! as ESTreeNode & Record<string, unknown>;
        for (const key of VISITOR_KEYS[node.type]) {
          const value = node[key];
          for (const child of Array.isArray(value) ? value : [value]) {
            if (child == null) continue;
            expect((child as ESTreeNode).parent).toBe(node);
            stack.push(child as ESTreeNode);
          }
        }
      }
    });
  }
});

describe('projected ranges used by plugin diagnostics and fixes', () => {
  const cases = [
    {
      text: '// 😀\nconst { source: local = 1 } = input;',
      selector: 'AssignmentPattern',
      target: 'local = 1',
      replacement: 'local',
      expected: '// 😀\nconst { source: local } = input;',
    },
    {
      text: 'const { [key]: local = 1 } = input;',
      selector: 'AssignmentPattern',
      target: 'local = 1',
      replacement: 'local',
      expected: 'const { [key]: local } = input;',
    },
    {
      text: 'const [local = 1] = input;',
      selector: 'AssignmentPattern',
      target: 'local = 1',
      replacement: 'local',
      expected: 'const [local] = input;',
    },
    {
      text: '// 😀\nclass C { private /* constructor */ constructor() {} }',
      selector: 'MethodDefinition[kind="constructor"]',
      target: 'constructor',
      replacement: '"constructor"',
      expected:
        '// 😀\nclass C { private /* constructor */ "constructor"() {} }',
    },
    {
      text: 'const valid = left >= right;',
      selector: 'BinaryExpression',
      target: '>=',
      replacement: '>',
      expected: 'const valid = left > right;',
    },
  ];
  for (const fixture of cases) {
    test(fixture.text, async () => {
      const { lintFile } = await import('../../eslint-plugin/project-frame.js');
      const result = lintFile(
        {
          filePath: 'ranges.ts',
          text: fixture.text,
          rules: { 'probe/range': { options: [] } },
          collectFixes: true,
          suggestionsMode: 'off',
        },
        {
          plugins: [],
          rules: new Map([
            [
              'probe/range',
              {
                meta: { fixable: 'code' },
                create(ctx: RuleContext) {
                  return {
                    [fixture.selector](candidate: ESTreeNode) {
                      const node =
                        candidate.type === 'MethodDefinition'
                          ? (candidate.key as ESTreeNode)
                          : candidate.type === 'BinaryExpression'
                            ? (ctx.sourceCode.getFirstTokenBetween(
                                candidate.left as ESTreeNode,
                                candidate.right as ESTreeNode,
                              )! as unknown as ESTreeNode)
                            : candidate;
                      expect(ctx.sourceCode.getText(node)).toBe(fixture.target);
                      ctx.report({
                        node,
                        message: 'selected range',
                        fix: (fixer) =>
                          fixer.replaceText(node, fixture.replacement),
                      });
                    },
                  };
                },
              },
            ],
          ]),
        },
      );
      expect(result.diagnostics).toHaveLength(1);
      const start = Buffer.byteLength(
        fixture.text.slice(0, fixture.text.lastIndexOf(fixture.target)),
      );
      const end = start + Buffer.byteLength(fixture.target);
      expect(result.diagnostics[0].startPos).toBe(start);
      expect(result.diagnostics[0].endPos).toBe(end);
      expect(result.fixes).toEqual([
        { range: [start, end], text: fixture.replacement },
      ]);
      const bytes = Buffer.from(fixture.text);
      const applied = Buffer.concat([
        bytes.subarray(0, start),
        Buffer.from(result.fixes[0].text),
        bytes.subarray(end),
      ]).toString();
      expect(applied).toBe(fixture.expected);
      expect(() =>
        parseUpstream(applied, { filePath: 'ranges.ts' }),
      ).not.toThrow();
    });
  }

  test('JSX attribute boundaries preserve subsequent tokens and comments', async () => {
    const { lintFile } = await import('../../eslint-plugin/project-frame.js');
    const text = String.raw`const e = <A value="\" />; /* after */ next();`;
    let visited = false;
    const result = lintFile(
      {
        filePath: 'attribute.jsx',
        text,
        rules: { 'probe/tokens': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      {
        plugins: [],
        rules: new Map([
          [
            'probe/tokens',
            {
              create(ctx: RuleContext) {
                return {
                  JSXOpeningElement(node: ESTreeNode) {
                    visited = true;
                    const attributes = node.attributes as ESTreeNode[];
                    const value = attributes[0].value as ESTreeNode;
                    expect(ctx.sourceCode.getText(value)).toBe(String.raw`"\"`);
                    expect(ctx.sourceCode.getTokenAfter(value)!.value).toBe(
                      '/',
                    );
                    expect(ctx.sourceCode.getLastToken(node)!.value).toBe('>');
                    const comments = ctx.sourceCode.getAllComments();
                    expect(comments.map((comment) => comment.value)).toEqual([
                      ' after ',
                    ]);
                    expect(
                      ctx.sourceCode.getTokenAfter(node, { skip: 1 })!.value,
                    ).toBe('next');
                    expect(
                      ctx.sourceCode.getTokenAfter(node, { skip: 1 })!.loc.start
                        .column,
                    ).toBe(text.indexOf('next'));
                  },
                };
              },
            },
          ],
        ]),
      },
    );
    expect(visited).toBe(true);
    expect(result.ruleErrors ?? []).toEqual([]);
  });
});

test('parenthesized constructor type arguments stay on NewExpression', () => {
  const ast = projectSyntax('file.ts', 'new (f<T>)(x)').ast as ESTreeNode & {
    body: Array<{ expression: Record<string, unknown> }>;
  };
  const node = ast.body[0].expression;
  expect((node.callee as ESTreeNode).type).toBe('Identifier');
  expect((node.typeArguments as ESTreeNode).type).toBe(
    'TSTypeParameterInstantiation',
  );
});

describe('hostile frames and ownership', () => {
  test('a shared empty wire list decodes into independent mutable arrays', () => {
    const { ast } = projectSyntax(
      'file.ts',
      'const first = 1; const second = 2;',
    );
    const program = ast as ESTreeNode & {
      body: Array<{
        declarations: Array<{ id: { decorators: unknown[] } }>;
      }>;
    };
    const first = program.body[0].declarations[0].id.decorators;
    const second = program.body[1].declarations[0].id.decorators;
    expect(first).toEqual([]);
    expect(second).toEqual([]);
    expect(first).not.toBe(second);
    first.push('retained');
    expect(second).toEqual([]);
  });

  test('schema constructors produce mutable plain objects with absent optional fields', () => {
    const original = projectFrame(
      'file.js',
      "const name = 'stable😀'; function read(value) { return value + name; } const pattern = /a/g; const cooked = `text${name}`;",
    );
    const storage = new Uint8Array(original.length + 11);
    storage.set(original, 7);
    const bytes = storage.subarray(7, 7 + original.length);
    const { ast } = decodeFrame(bytes);
    const visit = (node: ESTreeNode): void => {
      expect(Object.getPrototypeOf(node)).toBe(Object.prototype);
      for (const key of Object.keys(node)) {
        const descriptor = Object.getOwnPropertyDescriptor(node, key)!;
        expect(descriptor.writable).toBe(true);
        expect(descriptor.enumerable).toBe(true);
        expect(descriptor.configurable).toBe(true);
      }
      if (node.type === 'Identifier') {
        expect(Object.hasOwn(node, 'typeAnnotation')).toBe(false);
        expect(Object.hasOwn(node, 'optional')).toBe(false);
        if (node.name === 'name')
          expect(Object.hasOwn(node, 'decorators')).toBe(false);
      }
      for (const key of VISITOR_KEYS[node.type]) {
        const value = node[key];
        const children = Array.isArray(value) ? value : [value];
        for (const child of children) {
          if (!child) continue;
          expect((child as ESTreeNode).parent).toBe(node);
          expect(Object.keys(child).at(-1)).toBe('parent');
          visit(child as ESTreeNode);
        }
      }
    };
    visit(ast);
    expect(Object.hasOwn(ast, 'parent')).toBe(false);
    bytes.fill(0);
    expect(ast.type).toBe('Program');
    const program = ast as ESTreeNode & {
      body: Array<{
        declarations: Array<{ id: { name: string }; init: { value: string } }>;
      }>;
    };
    expect(program.body[0].declarations[0].id.name).toBe('name');
    expect(program.body[0].declarations[0].init.value).toBe('stable😀');
    ast.type = 'retained';
    expect(ast.type).toBe('retained');
  });

  test('metadata initialization does not invoke an inherited setter', () => {
    const bytes = projectFrame('file.js', 'const value = 1;');
    const old = Object.getOwnPropertyDescriptor(Object.prototype, 'start');
    let writes = 0;
    Object.defineProperty(Object.prototype, 'start', {
      configurable: true,
      set() {
        writes++;
      },
    });
    try {
      const { ast } = decodeFrame(bytes);
      expect(Object.hasOwn(ast, 'start')).toBe(true);
      expect(ast.start).toBe(0);
      expect(writes).toBe(0);
    } finally {
      if (old) Object.defineProperty(Object.prototype, 'start', old);
      else Reflect.deleteProperty(Object.prototype, 'start');
    }
  });

  test.each(['setter', 'readonly'] as const)(
    'wire fields become own data properties over an inherited %s',
    (kind) => {
      const bytes = projectFrame('file.js', 'const value = 1;');
      const old = Object.getOwnPropertyDescriptor(
        Object.prototype,
        'declarations',
      );
      let writes = 0;
      Object.defineProperty(
        Object.prototype,
        'declarations',
        kind === 'setter'
          ? {
              configurable: true,
              set() {
                writes++;
              },
            }
          : { configurable: true, value: 'inherited', writable: false },
      );
      let ast: ESTreeNode;
      try {
        ast = decodeFrame(bytes).ast;
      } finally {
        if (old) Object.defineProperty(Object.prototype, 'declarations', old);
        else Reflect.deleteProperty(Object.prototype, 'declarations');
      }
      const declaration = (ast as ESTreeNode & { body: ESTreeNode[] }).body[0];
      const descriptor = Object.getOwnPropertyDescriptor(
        declaration,
        'declarations',
      )!;
      expect(writes).toBe(0);
      expect(descriptor).toMatchObject({
        writable: true,
        enumerable: true,
        configurable: true,
      });
      expect(descriptor.value).toHaveLength(1);
      expect(descriptor.get).toBeUndefined();
      expect(descriptor.set).toBeUndefined();
    },
  );

  test('parent linking preserves inherited setter assignment', () => {
    const bytes = projectFrame('file.js', 'const value = 1;');
    const old = Object.getOwnPropertyDescriptor(Object.prototype, 'parent');
    const parents = new Map<object, ESTreeNode>();
    Object.defineProperty(Object.prototype, 'parent', {
      configurable: true,
      set(this: ESTreeNode, value: ESTreeNode) {
        expect(this.type).toBeDefined();
        parents.set(this, value);
      },
    });
    try {
      const { ast } = decodeFrame(bytes);
      const visit = (node: ESTreeNode): void => {
        expect(Object.hasOwn(node, 'parent')).toBe(false);
        for (const key of VISITOR_KEYS[node.type]) {
          const value = node[key];
          const children = Array.isArray(value) ? value : [value];
          for (const child of children) {
            if (!child) continue;
            expect(parents.get(child as ESTreeNode)).toBe(node);
            visit(child as ESTreeNode);
          }
        }
      };
      visit(ast);
      expect(parents.size).toBe(4);
    } finally {
      if (old) Object.defineProperty(Object.prototype, 'parent', old);
      else Reflect.deleteProperty(Object.prototype, 'parent');
    }
  });

  test('records reached only through non-structural fields have no parent', () => {
    const bytes = projectFrame('file.js', 'first(); second();');
    const count = bytes.readUInt32LE(24);
    const starts = Math.ceil((64 + count * 2) / 4) * 4;
    const valueStarts = starts + count * 8;
    const values = valueStarts + count * 4;
    const lists = values + bytes.readUInt32LE(28) * 4;
    const fieldOffset = (id: number, key: string): number => {
      const type = bytes.readUInt16LE(64 + (id - 1) * 2);
      const field = FRAME_SCHEMAS[type][1].indexOf(key);
      expect(field).toBeGreaterThanOrEqual(0);
      return (
        values + (bytes.readUInt32LE(valueStarts + (id - 1) * 4) + field) * 4
      );
    };
    const nodeID = (offset: number): number =>
      bytes.readUInt32LE(offset) & 0x1fffffff;
    const root = bytes.readUInt32LE(20);
    const body = lists + nodeID(fieldOffset(root, 'body')) * 4;
    const first = bytes.readUInt32LE(body + 4);
    const second = bytes.readUInt32LE(body + 8);
    const call = nodeID(fieldOffset(first, 'expression'));
    const callee = nodeID(fieldOffset(call, 'callee'));
    // Remove the second statement's structural edge while exposing the same
    // record through a scalar field. Both forms were accepted by the decoder.
    bytes.writeUInt32LE(0, body + 8);
    bytes.writeUInt32LE(0x20000000 | second, fieldOffset(callee, 'name'));
    const { ast } = decodeFrame(bytes);
    const program = ast as ESTreeNode & {
      body: Array<{ expression: { callee: { name: ESTreeNode } } }>;
    };
    const detached = program.body[0].expression.callee.name;
    expect(detached.type).toBe('ExpressionStatement');
    expect(Object.hasOwn(detached, 'parent')).toBe(false);
    expect(Object.hasOwn(detached.expression as ESTreeNode, 'parent')).toBe(
      false,
    );
  });

  test('decodes long UTF-16 values across chunks without changing code units', () => {
    const value = 'a'.repeat(4095) + '\ud800\udc00\udfff' + 'b'.repeat(4096);
    const decoded = projectSyntax(
      'file.ts',
      `const value = ${JSON.stringify(value)};`,
    );
    const program = decoded.ast as ESTreeNode & {
      body: Array<{ declarations: Array<{ init: { value: string } }> }>;
    };
    expect(program.body[0].declarations[0].init.value).toBe(value);
  });

  test('rejects concurrently mutable frame storage before validating fields', () => {
    const original = projectFrame('file.ts', 'const value = 1;');
    const shared = new Uint8Array(new SharedArrayBuffer(original.length));
    shared.set(original);
    expect(() => decodeFrame(shared)).toThrow('invalid ESTree frame');
  });

  test.each([0, 4, 8, 12, 16, 20, 24, 28, 32, 36, 40, 44, 48, 52, 56, 60])(
    'rejects corrupt header word %s',
    (offset) => {
      const bytes = projectFrame('file.ts', 'const a = 1; const b = 2;');
      bytes.writeUInt32LE(0xffffffff, offset);
      expect(() => decodeFrame(bytes)).toThrow();
    },
  );

  test('rejects truncated frames, bad UTF-8 and cyclic or shared children', () => {
    const original = projectFrame('file.ts', 'a(); b();');
    expect(() =>
      decodeFrame(original.subarray(0, original.length - 1)),
    ).toThrow();
    const badText = Buffer.from(original);
    badText[badText.length - 1] = 0xff;
    expect(() => decodeFrame(badText)).toThrow();
    const count = original.readUInt32LE(24);
    const starts = Math.ceil((64 + count * 2) / 4) * 4;
    const values = starts + count * 12;
    const lists = values + original.readUInt32LE(28) * 4;
    const root = original.readUInt32LE(20);
    const rootValues = original.readUInt32LE(
      starts + count * 8 + (root - 1) * 4,
    );
    const bodyOffset =
      original.readUInt32LE(values + rootValues * 4) & 0x1fffffff;
    const body = lists + bodyOffset * 4;
    // A payload must point to a list header, never an element or beyond the
    // list column. This also exercises bounds checks in the compact index.
    for (const offset of [
      bodyOffset + 1,
      original.readUInt32LE(32),
      0x1fffffff,
    ]) {
      const badList = Buffer.from(original);
      badList.writeUInt32LE(0x40000000 | offset, values + rootValues * 4);
      expect(() => decodeFrame(badList)).toThrow('invalid ESTree frame');
    }
    const cycle = Buffer.from(original);
    cycle.writeUInt32LE(root, body + 4);
    expect(() => decodeFrame(cycle)).toThrow('invalid ESTree frame');
    const duplicate = Buffer.from(original);
    duplicate.writeUInt32LE(duplicate.readUInt32LE(body + 4), body + 8);
    expect(() => decodeFrame(duplicate)).toThrow('invalid ESTree frame');
    const wrongBody = Buffer.from(original);
    wrongBody.writeUInt32LE(
      0x20000000 | original.readUInt32LE(body + 4),
      values + rootValues * 4,
    );
    expect(() => decodeFrame(wrongBody)).toThrow('invalid ESTree frame');
  });

  test('retained AST and token columns own their data after the frame is overwritten', () => {
    const text = "const s = '\\uD800'; const n = 0x10n; // hi\n";
    const bytes = projectFrame('file.ts', text);
    const padded = Buffer.concat([Buffer.alloc(7), bytes, Buffer.alloc(9)]);
    const decoded = decodeFrame(padded.subarray(7, 7 + bytes.length));
    const saved = canonical(decoded.ast);
    const tokens = [...decoded.tokenStarts];
    padded.fill(0);
    expect(decoded.text).toBe(text);
    expect(canonical(decoded.ast)).toEqual(saved);
    expect([...decoded.tokenStarts]).toEqual(tokens);
    expect(decoded.comments[0].value).toBe(' hi');
  });

  test('strips exactly one BOM and checks cancellation before reading an invalid lease', () => {
    const decoded = projectSyntax('file.ts', '\ufeff\ufeffconst x = 1;');
    expect(decoded.hasBOM).toBe(true);
    expect(decoded.text).toBe('\ufeffconst x = 1;');
    expect(decoded.ast.range[0]).toBe(1);
    const flag = new Int32Array(new SharedArrayBuffer(4));
    Atomics.store(flag, 0, 1);
    expect(() => decodeFrame(new Uint8Array(), flag)).toThrow(FrameCancelled);
    const result = lintFrame(
      {
        filePath: 'file.ts',
        frame: { lease: 0xffffffff, offset: 0, length: 1 },
        cancelFlag: flag,
        rules: {},
        collectFixes: false,
        suggestionsMode: 'off',
      },
      { plugins: [], rules: new Map() },
    );
    expect(result.cancelled).toBe(true);
    expect(result.parseError).toBeUndefined();
  });
});

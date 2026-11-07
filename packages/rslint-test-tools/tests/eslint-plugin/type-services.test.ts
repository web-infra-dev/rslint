import { beforeAll, afterAll, describe, expect, test } from 'rstack/test';
import { execFileSync } from 'node:child_process';
import { mkdtempSync, rmSync, writeFileSync, readFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { resolve } from 'node:path';
import { Worker } from 'node:worker_threads';
import { pathToFileURL } from 'node:url';
import { Linter } from 'eslint';
import tsParser from '@typescript-eslint/parser';
import typedPlugin from '@typescript-eslint/eslint-plugin';
import sonarPlugin from 'eslint-plugin-sonarjs';
import effectorPlugin from 'eslint-plugin-effector';
import ts from 'typescript';
import {
  TypeServiceClient,
  typeQueryKey,
  type TypeServiceDescriptor,
} from '../../../rslint/src/languages/js/type-service.js';
import {
  SharedTypeCacheOwner,
  SharedTypeCacheReader,
} from '../../../rslint/src/languages/js/shared-type-cache.js';
import { lintFile } from '../../../rslint/src/eslint-plugin/linter/ecma-language-plugin.js';
import { decodeFrame } from '../../../rslint/src/languages/js/frame.js';
import { VISITOR_KEYS } from '../../../rslint/src/languages/js/visitor-keys.js';

const root = resolve(import.meta.dirname, '../../../..');
const enums = {
  SyntaxKind: ts.SyntaxKind,
  TypeFlags: ts.TypeFlags,
  NodeFlags: ts.NodeFlags,
  JsxEmit: ts.JsxEmit,
  ModuleDetectionKind: ts.ModuleDetectionKind,
};
let temp: string, binary: string, readerModule: string;
beforeAll(() => {
  temp = mkdtempSync(resolve(tmpdir(), 'rslint-types-'));
  binary = resolve(temp, process.platform === 'win32' ? 'types.exe' : 'types');
  readerModule = pathToFileURL(resolve(temp, 'shared-cache.mjs')).href;
  writeFileSync(
    new URL(readerModule),
    ts.transpileModule(
      readFileSync(
        resolve(root, 'packages/rslint/src/languages/js/shared-type-cache.ts'),
        'utf8',
      ),
      {
        compilerOptions: {
          module: ts.ModuleKind.ESNext,
          target: ts.ScriptTarget.ES2022,
        },
      },
    ).outputText,
  );
  execFileSync(
    'go',
    ['build', '-o', binary, './internal/typeservice/testdata/server'],
    { cwd: root, stdio: 'pipe' },
  );
});
afterAll(() => rmSync(temp, { recursive: true, force: true }));

class Host {
  mailbox = new SharedArrayBuffer(16 << 20);
  control = new Int32Array(this.mailbox, 0, 2);
  worker: Worker;
  requests = 0;
  constructor() {
    this.worker = new Worker(
      resolve(
        root,
        'packages/rslint-test-tools/tests/eslint-plugin/type-service-broker.mjs',
      ),
      { workerData: { binary, cwd: root, mailbox: this.mailbox } },
    );
  }
  request(request: object): any {
    Atomics.store(this.control, 0, 0);
    this.worker.postMessage(request);
    if (Atomics.wait(this.control, 0, 0, 60_000) === 'timed-out')
      throw new Error('type service host timed out');
    const response = JSON.parse(
      Buffer.from(this.mailbox, 8, Atomics.load(this.control, 1)).toString(),
    );
    if (response.error) throw new Error(response.error);
    return response.result;
  }
  async close() {
    this.request({ method: 'close' });
    const stopped = new Promise<void>((done, reject) => {
      this.worker.once('message', () => done());
      this.worker.once('error', reject);
    });
    this.worker.postMessage({ method: 'stop' });
    await stopped;
    await this.worker.terminate();
  }
}

interface Case {
  name: string;
  plugin: any;
  rule: string;
  code: string;
  options?: unknown[];
  reports: number;
  dependencies?: Record<string, string>;
  requiredOps?: string[];
}
// Focused valid/invalid examples follow the original rule fixtures and exercise
// their actual npm implementations, including real type-utils dependencies.
const cases: Case[] = [
  {
    name: 'await nonthenable and Promise',
    plugin: typedPlugin,
    rule: 'await-thenable',
    code: 'const text = "😀中文"; async function run() { await 42; await Promise.resolve(1); }',
    reports: 1,
  },
  {
    name: 'custom thenable and constrained generic',
    plugin: typedPlugin,
    rule: 'await-thenable',
    code: 'async function run<T extends PromiseLike<number>>(value: T) { await value; await { then(resolve: (x: number) => unknown) { return resolve(1); } }; }',
    reports: 0,
  },
  {
    name: 'union operands',
    plugin: typedPlugin,
    rule: 'restrict-plus-operands',
    code: 'declare const value: string | {}; const bad = value + 1; const good = 1 + 2; const big = 1n + 2n;',
    reports: 1,
  },
  {
    name: 'nondefault plus options',
    plugin: typedPlugin,
    rule: 'restrict-plus-operands',
    options: [{ allowNumberAndString: false }],
    code: 'const value = "x" + 1;',
    reports: 1,
  },
  {
    name: 'return-await fix and TS parents',
    plugin: typedPlugin,
    rule: 'return-await',
    options: ['always'],
    code: 'async function run() { return Promise.resolve(1); }',
    reports: 1,
  },
  {
    name: 'cross-file union value',
    plugin: typedPlugin,
    rule: 'await-thenable',
    code: 'import { value } from "./dependency"; async function run() { await value; }',
    dependencies: {
      'dependency.ts': 'export const value: number | string = 1;',
    },
    reports: 1,
  },
  {
    name: 'collection size semantics',
    plugin: sonarPlugin,
    rule: 'no-collection-size-mischeck',
    code: 'declare const list: number[]; declare const set: Set<number>; declare const other: { length: number }; if (list.length < 0) {} if (set.size < 0) {} if (other.length < 0) {}',
    reports: 2,
  },
  {
    name: 'Effector store naming',
    plugin: effectorPlugin,
    rule: 'enforce-store-naming-convention',
    code: 'import { createStore } from "effector"; const count = createStore(0); const $good = createStore(1); const local: { getState(): number } = { getState: () => 0 }; local.getState();',
    reports: 1,
  },
  {
    name: 'Effector no-getState cross-file',
    plugin: effectorPlugin,
    rule: 'no-getState',
    code: 'import { $count } from "./dependency"; const value = $count.getState();',
    dependencies: {
      'dependency.ts':
        'import { createStore } from "effector"; export const $count = createStore(0);',
    },
    reports: 1,
  },
  {
    name: 'Effector no-watch',
    plugin: effectorPlugin,
    rule: 'no-watch',
    code: 'import { createEvent } from "effector"; const changed = createEvent<number>(); changed.watch(value => console.log(value)); interface Event<T> { watch(cb: (value: T) => void): void } declare const fake: Event<number>; fake.watch(() => {});',
    reports: 1,
  },
  {
    name: 'Effector callback signatures',
    plugin: effectorPlugin,
    rule: 'no-unnecessary-combination',
    code: 'import { combine, createStore, sample } from "effector"; const $value = createStore(1); const transform = (n: number) => n + 1; sample({ source: combine($value, transform) }); sample({ source: combine({ $value }) });',
    reports: 1,
    requiredOps: ['signatures'],
  },
  {
    name: 'Effector contextual component signature',
    plugin: effectorPlugin,
    rule: 'mandatory-scope-binding',
    code: 'import type { FC } from "react"; import { createEvent } from "effector"; const clicked = createEvent(); const Button: FC = () => { clicked(); return null; }; declare function render(callback: () => null): void; render(() => null);',
    reports: 1,
    requiredOps: ['signature', 'returnType', 'context'],
  },
];

function normMessages(messages: any[]) {
  return messages.map((m) => ({
    message: m.message,
    line: m.line,
    column: m.column,
    endLine: m.endLine,
    endColumn: m.endColumn,
    fix: m.fix ? { range: m.fix.range, text: m.fix.text } : undefined,
    suggestions: m.suggestions?.map((s: any) => ({
      desc: s.desc,
      fix: { range: s.fix.range, text: s.fix.text },
    })),
  }));
}
function normResult(result: ReturnType<typeof lintFile>, code: string) {
  const byteToUTF16 = (offset: number): number =>
    Buffer.from(code).subarray(0, offset).toString().length;
  const fix = (fixes: any[] | null | undefined) => {
    if (!fixes?.length) return undefined;
    const sorted = [...fixes].sort((a, b) => a.range[0] - b.range[0]);
    const start = byteToUTF16(sorted[0].range[0]),
      end = byteToUTF16(sorted.at(-1).range[1]);
    let text = '',
      cursor = start;
    for (const f of sorted) {
      const from = byteToUTF16(f.range[0]),
        to = byteToUTF16(f.range[1]);
      text += code.slice(cursor, from) + f.text;
      cursor = to;
    }
    return { range: [start, end], text: text + code.slice(cursor, end) };
  };
  const loc = (offset: number) => {
    const prefix = code.slice(0, byteToUTF16(offset));
    const lines = prefix.split('\n');
    return { line: lines.length, column: lines.at(-1)!.length + 1 };
  };
  return result.diagnostics.map((d) => ({
    message: d.message,
    line: loc(d.startPos).line,
    column: loc(d.startPos).column,
    endLine: loc(d.endPos).line,
    endColumn: loc(d.endPos).column,
    fix: fix(d.fixes),
    suggestions: d.suggestions?.map((s) => ({
      desc: s.desc,
      fix: fix(s.fixes),
    })),
  }));
}

describe('original npm typed rules over Go type services', () => {
  for (const c of cases)
    test(c.name, async () => {
      // Keep fixtures below the workspace so both compilers resolve the exact
      // installed npm declarations, without copying or faking library types.
      const directory = mkdtempSync(
        resolve(root, 'packages/rslint-test-tools/.type-services-'),
      );
      const file = resolve(directory, 'input.ts'),
        config = resolve(directory, 'tsconfig.json');
      writeFileSync(file, c.code);
      for (const [name, text] of Object.entries(c.dependencies ?? {}))
        writeFileSync(resolve(directory, name), text);
      writeFileSync(
        config,
        JSON.stringify({
          compilerOptions: {
            strict: true,
            target: 'ES2022',
            module: 'ESNext',
            moduleResolution: 'Bundler',
            skipLibCheck: true,
          },
          include: ['*.ts'],
        }),
      );
      const host = new Host();
      try {
        const descriptor = host.request({
          method: 'init',
          config,
        }) as TypeServiceDescriptor;
        const encoded = host.request({ method: 'project', file });
        const frame = Buffer.from(encoded.frame, 'base64');
        const projection = {
          ...encoded,
          tree: Buffer.from(encoded.tree, 'base64'),
        };
        const rules = new Map([
          [`community/${c.rule}`, c.plugin.rules[c.rule]],
        ]);
        const request = {
          filePath: file,
          frame,
          rules: { [`community/${c.rule}`]: { options: c.options ?? [] } },
          collectFixes: true,
          suggestionsMode: 'eager' as const,
        };
        const eslint = new Linter();
        const expected = eslint.verify(
          c.code,
          [
            {
              files: ['**/*.ts'],
              languageOptions: {
                parser: tsParser,
                parserOptions: { project: config, tsconfigRootDir: directory },
              },
              plugins: { community: c.plugin },
              rules: {
                [`community/${c.rule}`]: ['error', ...(c.options ?? [])],
              },
            },
          ],
          { filename: file },
        );
        expect(expected.filter((m) => m.fatal)).toEqual([]);
        expect(expected).toHaveLength(c.reports);
        expect(descriptor.sources.some((source) => source.defaultLibrary)).toBe(
          true,
        );
        const operations = new Set<string>();
        for (const budget of [
          undefined,
          { bytes: 4 << 20, entries: 8192 },
          { bytes: 64, entries: 1 },
        ]) {
          const cache =
            budget &&
            new SharedTypeCacheOwner(
              descriptor.generation,
              budget.bytes,
              budget.entries,
            );
          for (let pass = 0; pass < 2; pass++) {
            const client = new TypeServiceClient(
              enums,
              descriptor,
              (query, key) => {
                operations.add(query.op);
                host.requests++;
                const packet = Buffer.from(
                  host.request({ method: 'query', query }),
                  'base64',
                );
                cache?.publish(key, packet);
                return packet;
              },
              {
                shared: cache
                  ? new SharedTypeCacheReader(cache.descriptor)
                  : undefined,
                maxFacts: 4096,
              },
            );
            const before = host.requests;
            const result = lintFile(request, { rules } as any, (decoded) =>
              client.parserServices(
                decoded,
                projection,
                file.replaceAll('\\', '/'),
              ),
            );
            expect(result.parseError).toBeUndefined();
            expect(result.ruleErrors).toBeUndefined();
            expect(normResult(result, c.code)).toEqual(normMessages(expected));
            if (pass === 1 && budget?.entries === 8192) {
              expect(client.stats.sharedHits).toBeGreaterThan(0);
              expect(host.requests - before).toBe(0);
            }
            expect(
              client.stats.queries + client.stats.sharedHits,
            ).toBeGreaterThan(0);
            client.close();
          }
          if (cache) {
            expect(cache.readers).toBe(0);
            cache.close();
          }
        }
        for (const op of c.requiredOps ?? [])
          expect(operations.has(op), op).toBe(true);
      } finally {
        await host.close();
        rmSync(directory, { recursive: true, force: true });
      }
    });
});

test('exact origins, aliases, metadata and disposal', async () => {
  const directory = mkdtempSync(
    resolve(root, 'packages/rslint-test-tools/.type-services-'),
  );
  const file = resolve(directory, 'mapping.tsx'),
    config = resolve(directory, 'tsconfig.json');
  const code = `\uFEFF\uFEFF"use strict";
import Default, { createStore as named } from "effector";
export function f<T>(plain: string, value = 1, ...rest: number[]) {
  return ((plain));
}
class C {
  #value = 1;
  m(x: string) { return x; }
  read() { return this.#value; }
}
declare const a: any;
a?.b.c;
f<string>("x");
const __name = "\\uD800" as const;
const truth = true as const;
const view = <Box />;
declare const Box: () => null;
declare namespace N { const value: number; interface Item<T> {} }
declare global { interface GlobalMarker {} }
type Query = typeof N.value;
namespace A.B {}
class Base<T> {}
class Derived extends Base<string> implements N.Item<number> {}
interface Child extends N.Item<boolean> {}
function meta() { return new.target; }
import.meta;
const first = 1, second = 2, third = 3;
first, second, third;
undefined;
`;
  writeFileSync(file, code);
  writeFileSync(resolve(directory, 'other.tsx'), code);
  writeFileSync(
    config,
    JSON.stringify({
      compilerOptions: {
        strict: false,
        strictNullChecks: true,
        target: 'ES2022',
        module: 'ESNext',
        moduleResolution: 'Bundler',
        jsx: 'react-native',
        moduleDetection: 'legacy',
        newLine: 'crlf',
        experimentalDecorators: true,
        types: [],
      },
      include: ['*.tsx'],
    }),
  );
  const host = new Host();
  try {
    const descriptor = host.request({
      method: 'init',
      config,
    }) as TypeServiceDescriptor;
    const encoded = host.request({ method: 'project', file });
    const projection = {
      ...encoded,
      tree: Buffer.from(encoded.tree, 'base64'),
    };
    const frame = decodeFrame(Buffer.from(encoded.frame, 'base64'));
    const cache = new SharedTypeCacheOwner(
      descriptor.generation,
      1 << 20,
      1024,
    );
    const published = new Map<string, Uint8Array>();
    const client = new TypeServiceClient(
      enums,
      descriptor,
      (query, key) => {
        const packet = Buffer.from(
          host.request({ method: 'query', query }),
          'base64',
        );
        published.set(key, packet);
        cache.publish(key, packet);
        return packet;
      },
      { maxFacts: 128, shared: new SharedTypeCacheReader(cache.descriptor) },
    );
    const services = client.parserServices(
      frame,
      projection,
      file.replaceAll('\\', '/'),
    ) as any;
    const other = host.request({
      method: 'project',
      file: resolve(directory, 'other.tsx'),
    });
    expect(() =>
      client.parserServices(
        frame,
        { ...other, tree: Buffer.from(other.tree, 'base64') },
        file.replaceAll('\\', '/'),
      ),
    ).toThrow('different source');
    const upstream = tsParser.parseForESLint(code, {
      filePath: file,
      project: config,
      tsconfigRootDir: directory,
      range: true,
      loc: true,
    });
    const collect = (ast: any, maps: any) => {
      const entries: any[] = [];
      const walk = (node: any, path: string) => {
        let native = maps.esTreeNodeToTSNodeMap.get(node);
        if (node.type === 'Identifier' && path.endsWith('.meta')) {
          // Go stores this keyword as a scalar; JS TypeScript creates a
          // virtual token. The synthetic identifier has no Go node handle.
          if (maps === services) {
            expect(native).toBeUndefined();
            expect(() => services.getTypeAtLocation(node)).toThrow(
              'no compiler origin',
            );
          } else
            expect([
              ts.SyntaxKind.NewKeyword,
              ts.SyntaxKind.ImportKeyword,
            ]).toContain(native.kind);
          native = undefined;
        }
        if (node.type === 'MetaProperty')
          expect(native.keywordToken).toBe(
            node.meta.name === 'new'
              ? ts.SyntaxKind.NewKeyword
              : ts.SyntaxKind.ImportKeyword,
          );
        const reverse = native && maps.tsNodeToESTreeNodeMap.get(native);
        entries.push([
          path,
          node.type,
          native ? ts.SyntaxKind[native.kind] : undefined,
          native?.getStart(),
          native?.end,
          reverse?.type,
          reverse?.range,
        ]);
        for (const key of VISITOR_KEYS[node.type] ?? []) {
          const value = node[key];
          if (Array.isArray(value))
            value.forEach(
              (child, index) =>
                child && walk(child, `${path}.${key}[${index}]`),
            );
          else if (value) walk(value, `${path}.${key}`);
        }
      };
      walk(ast, 'root');
      return entries;
    };
    expect(collect(frame.ast, services)).toEqual(
      collect(upstream.ast, upstream.services),
    );
    const body = frame.ast.body as any[];
    const variable = (name: string) =>
      body
        .flatMap((node) => node.declarations ?? [])
        .find((declaration) => declaration.id.name === name);
    const expression = (type: string) =>
      body.find(
        (node) =>
          node.type === 'ExpressionStatement' && node.expression.type === type,
      ).expression;
    const fn = body.find(
      (node) => node.type === 'ExportNamedDeclaration',
    ).declaration;
    const nativeFn = services.esTreeNodeToTSNodeMap.get(fn);
    for (let i = 0; i < fn.params.length; i++)
      expect(services.tsNodeToESTreeNodeMap.get(nativeFn.parameters[i])).toBe(
        fn.params[i],
      );
    const returnExpression = nativeFn.body.statements[0].expression;
    expect(services.tsNodeToESTreeNodeMap.get(returnExpression).type).toBe(
      'Identifier',
    );
    const outerChain = expression('ChainExpression');
    const innerTS = services.esTreeNodeToTSNodeMap.get(
      outerChain.expression.object,
    );
    expect(services.tsNodeToESTreeNodeMap.get(innerTS).type).toBe(
      'ChainExpression',
    );
    const sequence = services.esTreeNodeToTSNodeMap.get(
      expression('SequenceExpression'),
    );
    const detached = services.tsNodeToESTreeNodeMap.get(sequence.left);
    expect(detached.type).toBe('SequenceExpression');
    expect(detached.expressions.map((node: any) => node.name)).toEqual([
      'first',
      'second',
    ]);
    const options = services.program.getCompilerOptions();
    const referenceOptions = upstream.services.program!.getCompilerOptions();
    for (const key of [
      'strict',
      'strictNullChecks',
      'jsx',
      'moduleDetection',
      'newLine',
      'target',
      'module',
      'moduleResolution',
      'experimentalDecorators',
    ])
      expect(options[key], key).toBe(referenceOptions[key]);
    expect(services.experimentalDecorators).toBe(true);
    expect(
      services.program.getSourceFile(resolve(directory, 'missing.ts')),
    ).toBeUndefined();
    expect(services.program.getSourceFile(file)).toBe(nativeFn.getSourceFile());
    const declaration = variable('__name');
    const stringType = services.getTypeAtLocation(declaration.id);
    expect(stringType.value).toBe('\uD800');
    const symbol = services.getSymbolAtLocation(declaration.id);
    expect([symbol.name, symbol.escapedName]).toEqual(['__name', '___name']);
    expect(services.esTreeNodeToTSNodeMap.get(declaration.id).escapedText).toBe(
      '___name',
    );
    const booleanType = services.getTypeAtLocation(variable('truth').id);
    expect(booleanType.intrinsicName).toBe('true');
    expect(booleanType.isLiteral()).toBe(false);
    expect(booleanType.value).toBeUndefined();
    expect(
      services.getSymbolAtLocation(expression('Identifier')).declarations,
    ).toBeUndefined();
    const privateKey = body
      .find((node) => node.type === 'ClassDeclaration' && node.id.name === 'C')
      .body.body.find((node: any) => node.key.type === 'PrivateIdentifier').key;
    expect(services.esTreeNodeToTSNodeMap.get(privateKey).escapedText).toBe(
      '#value',
    );
    expect(services.getSymbolAtLocation(privateKey).name).toBe('#value');
    // A real worker realm reads a fact published by the original Go-backed
    // client. Its disposable byte view never becomes a plugin object.
    const query = {
      generation: descriptor.generation,
      checker: descriptor.checker,
      op: 'void',
    };
    const voidType = services.program.getTypeChecker().getVoidType();
    const key = typeQueryKey(query);
    const mailbox = new SharedArrayBuffer(8);
    const reader = new Worker(
      resolve(
        root,
        'packages/rslint-test-tools/tests/eslint-plugin/type-service-broker.mjs',
      ),
      {
        workerData: {
          readCache: cache.descriptor,
          key,
          expected: published.get(key),
          mailbox,
          readerModule,
        },
      },
    );
    const control = new Int32Array(mailbox);
    expect(Atomics.wait(control, 0, 0, 60_000)).not.toBe('timed-out');
    expect(Atomics.load(control, 1)).toBe(1);
    await reader.terminate();
    const saved = services.program.getCompilerOptions;
    client.close();
    expect(cache.readers).toBe(0);
    expect(() => voidType.flags).toThrow('closed');
    expect(() => saved()).toThrow('closed');
    expect(() => services.esTreeNodeToTSNodeMap.get(fn)).toThrow('closed');
    expect(() =>
      host.request({
        method: 'query',
        query: { ...query, generation: query.generation + 1 },
      }),
    ).toThrow('stale');
    host.request({ method: 'close' });
    expect(() => host.request({ method: 'query', query })).toThrow('closed');
    cache.close();
  } finally {
    await host.close();
    rmSync(directory, { recursive: true, force: true });
  }
});

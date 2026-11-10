import { beforeAll, afterAll, describe, expect, test } from 'rstack/test';
import { execFileSync } from 'node:child_process';
import { mkdtempSync, rmSync, writeFileSync, readFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { resolve, relative } from 'node:path';
import { Worker } from 'node:worker_threads';
import { pathToFileURL } from 'node:url';
import { Linter } from 'eslint';
import tsParser from '@typescript-eslint/parser';
import typedPlugin from '@typescript-eslint/eslint-plugin';
import sonarPlugin from 'eslint-plugin-sonarjs';
import effectorPlugin from 'eslint-plugin-effector';
import ts from 'typescript';
import {
  FileTypeServiceClient,
  TypeServiceClient,
  typeQueryKey,
  type TypeServiceDescriptor,
} from '../../../rslint/src/languages/js/type-service.js';
import {
  SharedTypeCacheOwner,
  MAX_SHARED_TYPE_PROBES,
  SharedTypeCacheReader,
} from '../../../rslint/src/languages/js/shared-type-cache.js';
import { lintFile } from '../../../rslint/src/eslint-plugin/linter/ecma-language-plugin.js';
import { decodeFrame } from '../../../rslint/src/languages/js/frame.js';
import {
  typeReplyPackets,
  TypePacketReader,
} from '../../../rslint/src/languages/js/type-service-codec.js';
import { VISITOR_KEYS } from '../../../rslint/src/languages/js/visitor-keys.js';

const root = resolve(import.meta.dirname, '../../../..');
const enums = {
  SyntaxKind: ts.SyntaxKind,
  TypeFlags: ts.TypeFlags,
  ObjectFlags: ts.ObjectFlags,
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
  source = (
    scope: { generation: number },
    file: string,
    metadataOnly = false,
  ) => {
    const reply = this.request({
      method: 'source',
      generation: scope.generation,
      file,
      metadataOnly,
    });
    return reply
      ? {
          fileName: reply.fileName,
          tree: reply.tree ? Buffer.from(reply.tree, 'base64') : undefined,
        }
      : undefined;
  };
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
    name: 'Promise aggregator arrays, tuples and iterators',
    plugin: typedPlugin,
    rule: 'await-thenable',
    code: 'declare const valid: Promise<number>[]; declare const invalid: number[]; declare const validTuple: readonly [Promise<number>, Promise<string>]; declare const invalidTuple: readonly [Promise<number>, number]; declare const validMap: Map<number, Promise<number>>; declare const invalidMap: Map<number, number>; Promise.all(valid); Promise.all(invalid); Promise.all(validTuple); Promise.all(invalidTuple); Promise.allSettled(validMap.values()); Promise.allSettled(invalidMap.values());',
    reports: 3,
    requiredOps: ['isArrayLike', 'numberIndexType', 'typeArguments'],
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
    name: 'already awaited returns and nonthenable fixes',
    plugin: typedPlugin,
    rule: 'return-await',
    options: ['always'],
    code: 'async function valid() { try { return await Promise.resolve(1); } finally { console.log("cleanup"); } } async function invalid() { return await 42; }',
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
  // See packages/eslint-plugin/tests/rules in typescript-eslint v8.65.0.
  {
    name: 'unsafe assignment through matching generic targets',
    plugin: typedPlugin,
    rule: 'no-unsafe-assignment',
    code: `
      declare const source: Map<string, Set<any>>;
      const unsafe: Map<string, Set<number>> = source;
      const safe: Map<string, Set<unknown>> = source;
      const differentTarget: ReadonlyMap<string, Set<number>> = source;
    `,
    reports: 1,
    requiredOps: ['typeArguments'],
  },
  {
    name: 'unsafe arguments after fixed and variadic tuple spreads',
    plugin: typedPlugin,
    rule: 'no-unsafe-argument',
    code: `
      declare function collect(first: number, second: string, third: number, ...rest: unknown[]): void;
      declare const fixed: readonly [number, string];
      declare const variadic: readonly [number, ...string[]];
      declare const unsafe: any;
      collect(...fixed, unsafe);
      collect(...variadic, unsafe);
    `,
    reports: 1,
    requiredOps: ['resolvedSignature', 'typeArguments'],
  },
  {
    name: 'async returns with unknown and concrete fulfillment types',
    plugin: typedPlugin,
    rule: 'no-unsafe-return',
    code: `
      declare const source: Promise<any>;
      async function safe(): Promise<unknown> { return source; }
      async function unsafe(): Promise<number> { return source; }
    `,
    reports: 1,
    requiredOps: ['awaited'],
  },
  {
    name: 'unsafe narrowing and safe widening of a fresh object literal',
    plugin: typedPlugin,
    rule: 'no-unsafe-type-assertion',
    code: `
      const safe = { name: 'leaf', count: 1 } as { name: string };
      declare const source: string | number;
      const unsafe = source as string;
    `,
    reports: 1,
    requiredOps: ['widened', 'isTypeAssignableTo'],
  },
  {
    name: 'redundant type guards, assertions and optional chaining',
    plugin: typedPlugin,
    rule: 'no-unnecessary-condition',
    options: [{ checkTypePredicates: true }],
    code: `
      declare function isString(value: unknown): value is string;
      declare function assert(value: unknown): asserts value;
      declare const text: string;
      declare const input: string | number;
      isString(text);
      isString(input);
      assert(true);
      declare const present: { [key: string]: string } | null;
      declare const optional: { [key: string]: string | undefined } | null;
      present?.item?.trim();
      optional?.item?.trim();
    `,
    reports: 3,
    requiredOps: [
      'typePredicate',
      'isTypeAssignableTo',
      'propertyType',
      'indexInfos',
    ],
  },
  {
    name: 'structurally redundant assertions respect readonly properties',
    plugin: typedPlugin,
    rule: 'no-unnecessary-type-assertion',
    code: `
      interface Source { value: string }
      interface Equivalent { value: string }
      interface ReadonlySource { readonly value: string }
      declare const source: Source;
      const redundant = source as Equivalent;
      const readonlyView = source as ReadonlySource;
    `,
    reports: 1,
    requiredOps: ['indexInfos', 'isTypeAssignableTo'],
  },
  {
    name: 'type parameter usage through signatures and derived types',
    plugin: typedPlugin,
    rule: 'no-unnecessary-type-parameters',
    code: `
      declare function mutable<T>(): [T];
      declare function immutable<T>(): readonly [T];
      declare function dictionary<T>(): { [key: string]: T };
      function echo<T>(this: T) { return this; }
      type Guard = <T>(value: unknown) => value is T;
      function label<T extends string>(value: T) { return \`label:\${value}\` as const; }
      function uppercase<T extends string>(value: T) { return value.toUpperCase() as Uppercase<T>; }
      function pick<T extends { value: string }, Key extends 'value'>(value: T, key: Key) {
        return value[key];
      }
      function compare<T, U>(value: T, target: U) { return 0 as T extends U ? 1 : 0; }
      declare function keys<T>(): keyof T;
    `,
    reports: 3,
    requiredOps: ['stringIndexType', 'typeOfSymbol', 'typePredicate'],
  },
  {
    name: 'default type arguments of a constructor-valued base class',
    plugin: typedPlugin,
    rule: 'no-unnecessary-type-arguments',
    code: `
      interface Base {}
      declare var Base: { new <T = string>(value: T): Base };
      class Redundant extends Base<string> {}
      class Explicit extends Base<number> {}
    `,
    reports: 1,
    requiredOps: ['typeOfSymbol'],
  },
  {
    name: 'bracket access through index signatures and named properties',
    plugin: typedPlugin,
    rule: 'dot-notation',
    options: [{ allowIndexSignaturePropertyAccess: true }],
    code: `
      declare const dictionary: { readonly [key: string]: number; known: number };
      dictionary['fromIndex'];
      dictionary['known'];
    `,
    reports: 1,
    requiredOps: ['indexInfos'],
  },
  {
    name: 'string conversion preserves the enum member exemption',
    plugin: typedPlugin,
    rule: 'no-unnecessary-type-conversion',
    code: `
      enum Status { Ready = 'ready' }
      Status.Ready.toString();
      declare const text: string;
      text.toString();
    `,
    reports: 1,
  },
  {
    name: 'namespace qualification depends on scope and symbol identity',
    plugin: typedPlugin,
    rule: 'no-unnecessary-qualifier',
    code: `
      namespace Outer {
        export type Value = number;
        export const direct: Outer.Value = 1;
        export namespace Inner {
          type Value = string;
          export const shadowed: Outer.Value = 1;
        }
      }
    `,
    reports: 1,
    requiredOps: ['symbolsInScope', 'exportSymbol'],
  },
  {
    name: 'named and star exports distinguish types from values across files',
    plugin: typedPlugin,
    rule: 'consistent-type-exports',
    code: `
      export { Item, count } from './dependency';
      export { Missing } from './dependency';
      export * from './types';
      export * from './dependency';
    `,
    dependencies: {
      'dependency.ts': 'export interface Item {} export const count = 1;',
      'types.ts': "export type * from './dependency';",
    },
    reports: 2,
    requiredOps: ['isUnknownSymbol', 'immediateAliasedSymbol', 'typeOfSymbol'],
  },
  {
    name: 'readonly parameters check index signatures and their value types',
    plugin: typedPlugin,
    rule: 'prefer-readonly-parameter-types',
    code: `
      declare function immutable(values: { readonly [key: string]: { readonly item: number } }): void;
      declare function mutableIndex(values: { [key: number]: number }): void;
      declare function mutableValue(values: { readonly [key: string]: { item: number } }): void;
    `,
    reports: 2,
    requiredOps: ['indexInfo', 'propertyType'],
  },
  {
    name: 'promise catch callback parameter types and suggestions',
    plugin: typedPlugin,
    rule: 'use-unknown-in-catch-callback-variable',
    code: `
      declare const promise: Promise<void>;
      promise.catch((error: unknown) => {});
      promise.catch(error => {});
    `,
    reports: 1,
    requiredOps: ['typeOfSymbol'],
  },
  {
    name: 'template interpolation inherits an allowlisted interface',
    plugin: typedPlugin,
    rule: 'restrict-template-expressions',
    options: [{ allow: ['Trusted'] }],
    code: `
      interface Trusted {}
      interface Derived extends Trusted {}
      interface Other {}
      declare const derived: Derived;
      declare const other: Other;
      \`\${derived}\`;
      \`\${other}\`;
    `,
    reports: 1,
    requiredOps: ['baseTypes'],
  },
  {
    name: 'collection size semantics',
    plugin: sonarPlugin,
    rule: 'no-collection-size-mischeck',
    code: 'declare const list: number[]; declare const set: Set<number>; declare const other: { length: number }; if (list.length < 0) {} if (set.size < 0) {} if (other.length < 0) {}',
    reports: 2,
  },
  {
    // SonarJS S3403 uses assignability in both directions for strict equality.
    name: 'SonarJS strict equality checks operand assignability',
    plugin: sonarPlugin,
    rule: 'different-types-comparison',
    code: `
      declare const count: number;
      declare const text: string;
      count === 1;
      count === text;
    `,
    reports: 1,
    requiredOps: ['isTypeAssignableTo'],
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
    ruleId: m.ruleId,
    messageId: m.messageId,
    message: m.message,
    line: m.line,
    column: m.column,
    endLine: m.endLine,
    endColumn: m.endColumn,
    fix: m.fix ? { range: m.fix.range, text: m.fix.text } : undefined,
    suggestions: m.suggestions?.map((s: any) => ({
      messageId: s.messageId,
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
  // The CLI and lint API sort worker diagnostics by source position.
  return [...result.diagnostics]
    .sort((a, b) => a.startPos - b.startPos)
    .map((d) => ({
      ruleId: d.ruleName,
      messageId: d.messageId,
      message: d.message,
      line: loc(d.startPos).line,
      column: loc(d.startPos).column,
      endLine: loc(d.endPos).line,
      endColumn: loc(d.endPos).column,
      fix: fix(d.fixes),
      suggestions: d.suggestions?.map((s) => ({
        messageId: s.messageId,
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
            const client = new FileTypeServiceClient(
              enums,
              descriptor,
              projection.binding,
              (query, key) => {
                operations.add(query.op);
                host.requests++;
                const packet = Buffer.from(
                  host.request({ method: 'query', query }),
                  'base64',
                );
                const reply = typeReplyPackets(packet);
                cache?.publish(key, reply.packet);
                for (const record of reply.records)
                  cache?.publish(
                    typeQueryKey({
                      ...query,
                      op: record.op,
                      args: { id: record.id },
                    }),
                    record.packet,
                  );
                return packet;
              },
              host.source,
              {
                shared: cache
                  ? new SharedTypeCacheReader(cache.descriptor)
                  : undefined,
                maxFactsBytes: 1 << 20,
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

test('generation clients reuse facts and lazy dependency identities across files', async () => {
  const directory = mkdtempSync(
    resolve(root, 'packages/rslint-test-tools/.type-services-'),
  );
  const config = resolve(directory, 'tsconfig.json');
  const code =
    'import { value } from "./dependency";\n' +
    Array.from({ length: 40 }, (_, index) => `const item${index}=value;`).join(
      '\n',
    );
  for (const name of ['first.ts', 'second.ts', 'third.ts'])
    writeFileSync(
      resolve(directory, name),
      name === 'third.ts' ? code + '\nawait value;' : code,
    );
  writeFileSync(
    resolve(directory, 'dependency.ts'),
    '/** 😀 */ export const value=1 as const;',
  );
  writeFileSync(
    config,
    JSON.stringify({
      compilerOptions: {
        strict: true,
        noLib: true,
        module: 'ESNext',
        moduleResolution: 'Bundler',
        types: [],
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
    let scalar = 0,
      batches = 0,
      sources = 0,
      metadataRequests = 0,
      failPrefetch = false;
    const owner = new TypeServiceClient(
      enums,
      descriptor,
      (query) => {
        scalar++;
        return Buffer.from(host.request({ method: 'query', query }), 'base64');
      },
      (scope, file, metadataOnly) => {
        if (metadataOnly) metadataRequests++;
        else sources++;
        return host.source(scope, file, metadataOnly);
      },
      {
        maxFactsBytes: 1 << 20,
        maxSyntaxBytes: 1 << 20,
        maxSyntaxFactsBytes: 1024,
        batchTransport: (queries) => {
          batches++;
          if (failPrefetch) throw new Error('speculative checker query failed');
          return host
            .request({ method: 'batch', queries })
            .map((packet: string) => Buffer.from(packet, 'base64'));
        },
      },
    );
    const open = (name: string) => {
      const file = resolve(directory, name),
        encoded = host.request({ method: 'project', file });
      const projection = {
        ...encoded,
        tree: Buffer.from(encoded.tree, 'base64'),
      };
      const frame = decodeFrame(Buffer.from(encoded.frame, 'base64'));
      const handle = owner.file(projection.binding, 'same rules and options');
      const services = handle.parserServices(
        frame,
        projection,
        file.replaceAll('\\', '/'),
      );
      return { file, frame, handle, services };
    };
    try {
      const first = open('first.ts');
      const declarations = (first.frame.ast.body as any[]).flatMap(
        (node) => node.declarations ?? [],
      );
      const firstType = first.services.getTypeAtLocation(declarations[0].id)!;
      expect(firstType.flags & ts.TypeFlags.NumberLiteral).toBeTruthy();
      for (const declaration of declarations)
        expect(first.services.getTypeAtLocation(declaration.id)).toBe(
          firstType,
        );
      const imported = ((first.frame.ast.body as any[])[0] as any).specifiers[0]
        .local;
      const alias = first.services.getSymbolAtLocation(imported)!;
      const symbol = owner.checker.getAliasedSymbol(alias)!;
      const declaration = symbol.declarations![0] as any;
      const source = declaration.getSourceFile();
      expect(source).toBe(
        owner.program.getSourceFile(resolve(directory, 'dependency.ts')),
      );
      expect(source.fileName.replaceAll('\\', '/')).toBe(
        resolve(directory, 'dependency.ts').replaceAll('\\', '/'),
      );
      expect(declaration.parent.kind).toBe(
        ts.SyntaxKind.VariableDeclarationList,
      );
      expect(declaration.name.text).toBe('value');
      expect(sources).toBe(0);
      // Resolve platform paths before counting the relative alias lookup.
      const firstSource = owner.program.getSourceFile(first.file),
        relativeFile = relative(descriptor.currentDirectory, first.file),
        beforeMetadata = metadataRequests;
      expect(owner.program.getSourceFile(relativeFile)).toBe(firstSource);
      expect(metadataRequests - beforeMetadata).toBe(1);
      expect(owner.program.getSourceFile(relativeFile)).toBe(firstSource);
      expect(metadataRequests - beforeMetadata).toBe(1);
      expect(sources).toBe(0);
      first.handle.close();
      expect(() => first.services.esTreeNodeToTSNodeMap.get(imported)).toThrow(
        'closed',
      );
      // File completion does not destroy the generation's semantic/native objects.
      expect(firstType.flags & ts.TypeFlags.NumberLiteral).toBeTruthy();
      const before = scalar,
        second = open('second.ts');
      expect(batches).toBeGreaterThan(0);
      for (const node of (second.frame.ast.body as any[]).flatMap(
        (node) => node.declarations ?? [],
      ))
        expect(second.services.getTypeAtLocation(node.id)).toBe(firstType);
      expect(scalar - before).toBe(0);
      expect(
        owner.checker.getAliasedSymbol(
          second.services.getSymbolAtLocation(
            ((second.frame.ast.body as any[])[0] as any).specifiers[0].local,
          )!,
        ),
      ).toBe(symbol);
      expect(symbol.declarations![0]).toBe(declaration);
      expect(declaration.getText()).toBe('value=1 as const');
      expect(sources).toBe(1);
      expect(symbol.declarations![0]).toBe(declaration);
      expect(declaration.getSourceFile()).toBe(source);
      expect('initializer' in declaration).toBe(true);
      expect(sources).toBe(1);
      second.handle.close();
      // The learned batch includes unconsumed nodes; a speculative failure must
      // not alter the demanded scalar result or diagnostics of this rule.
      failPrefetch = true;
      const thirdFile = resolve(directory, 'third.ts'),
        encoded = host.request({ method: 'project', file: thirdFile });
      const thirdProjection = {
        ...encoded,
        tree: Buffer.from(encoded.tree, 'base64'),
      };
      const thirdHandle = owner.file(
        thirdProjection.binding,
        'same rules and options',
      );
      const beforeDemand = scalar;
      const actual = lintFile(
        {
          filePath: thirdFile,
          frame: Buffer.from(encoded.frame, 'base64'),
          rules: { 'typed/await-thenable': { options: [] } },
          collectFixes: true,
          suggestionsMode: 'eager',
        },
        {
          rules: new Map([
            ['typed/await-thenable', typedPlugin.rules['await-thenable']],
          ]),
        } as any,
        (decoded) =>
          thirdHandle.parserServices(
            decoded,
            thirdProjection,
            thirdFile.replaceAll('\\', '/'),
          ),
      );
      const expected = new Linter().verify(
        code + '\nawait value;',
        [
          {
            files: ['**/*.ts'],
            languageOptions: {
              parser: tsParser,
              parserOptions: { project: config, tsconfigRootDir: directory },
            },
            plugins: { typed: typedPlugin as any },
            rules: { 'typed/await-thenable': 'error' },
          },
        ],
        { filename: thirdFile },
      );
      expect(expected).toHaveLength(1);
      expect(actual.ruleErrors).toBeUndefined();
      expect(actual.parseError).toBeUndefined();
      expect(normResult(actual, code + '\nawait value;')).toEqual(
        normMessages(expected),
      );
      expect(scalar).toBeGreaterThan(beforeDemand);
      thirdHandle.close();
      expect(owner.syntaxStats.summaryBytes).toBeLessThanOrEqual(1024);
      expect(owner.stats.factBytes).toBeLessThanOrEqual(1 << 20);
      expect(owner.syntaxStats.sourceBytes).toBeLessThanOrEqual(1 << 20);
      owner.close();
      expect(() => declaration.kind).toThrow('closed');
      expect(() => firstType.flags).toThrow('closed');
    } finally {
      owner.close();
    }
  } finally {
    await host.close();
    rmSync(directory, { recursive: true, force: true });
  }
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
declare const unique: unique symbol;
const uniqueKeys = {[unique]: true};
const truth = true as const;
const surrogateKeys = {"\\uD800":1 as const,"\\uD801":"two" as const,"\\uFFFD":true as const};
const spreadKeys = {...surrogateKeys};
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
await Promise.resolve(1);
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
    const makeClient = (maxFactsBytes = 1 << 16) =>
      new FileTypeServiceClient(
        enums,
        descriptor,
        projection.binding,
        (query, key) => {
          host.requests++;
          const packet = Buffer.from(
            host.request({ method: 'query', query }),
            'base64',
          );
          const reply = typeReplyPackets(packet);
          published.set(key, reply.packet);
          cache.publish(key, reply.packet);
          for (const record of reply.records) {
            const recordKey = typeQueryKey({
              ...query,
              op: record.op,
              args: { id: record.id },
            });
            published.set(recordKey, record.packet);
            cache.publish(recordKey, record.packet);
          }
          return packet;
        },
        host.source,
        {
          maxFactsBytes,
          shared: maxFactsBytes
            ? new SharedTypeCacheReader(cache.descriptor)
            : undefined,
        },
      );
    const client = makeClient();
    const beforeSource = client.program.getSourceFile(file);
    const beforeNodes = client.syntaxStats.nodes;
    const services = client.parserServices(
      frame,
      projection,
      file.replaceAll('\\', '/'),
    ) as any;
    expect(client.program.getSourceFile(file)).toBe(beforeSource);
    expect(client.syntaxStats.nodes).toBe(beforeNodes);
    expect(client.syntaxStats.aliases).toBe(0);
    expect(services.esTreeNodeToTSNodeMap.has(frame.ast)).toBe(true);
    expect(client.syntaxStats.nodes).toBe(beforeNodes);
    expect(() =>
      client.parserServices(frame, projection, file.replaceAll('\\', '/')),
    ).toThrow('already attached');

    const other = host.request({
      method: 'project',
      file: resolve(directory, 'other.tsx'),
    });
    const mismatch = makeClient();
    try {
      expect(() =>
        mismatch.parserServices(
          frame,
          { ...other, tree: Buffer.from(other.tree, 'base64') },
          file.replaceAll('\\', '/'),
        ),
      ).toThrow('different source');
    } finally {
      mismatch.close();
    }
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
    const nativeAwait = services.esTreeNodeToTSNodeMap.get(
      expression('AwaitExpression'),
    );
    expect(nativeAwait.getChildAt(1)).toBe(nativeAwait.expression);
    expect(nativeAwait.getChildAt(1).kind).toBe(ts.SyntaxKind.CallExpression);
    expect(() => nativeAwait.getChildAt(0)).toThrow('synthetic token');
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
    const canonical = expression('SequenceExpression');
    const sequence = services.esTreeNodeToTSNodeMap.get(canonical);
    const members = canonical.expressions;
    canonical.expressions = [members[2]];
    expect(services.tsNodeToESTreeNodeMap.has(sequence.left)).toBe(true);
    expect(client.syntaxStats.aliases).toBe(0);
    const detached = services.tsNodeToESTreeNodeMap.get(sequence.left);
    expect(client.syntaxStats.aliases).toBe(1);
    expect(services.tsNodeToESTreeNodeMap.get(sequence.left)).toBe(detached);
    expect(services.esTreeNodeToTSNodeMap.get(detached)).toBe(sequence.left);
    const upstreamSequence = (upstream.ast.body as any[]).find(
      (node) =>
        node.type === 'ExpressionStatement' &&
        node.expression.type === 'SequenceExpression',
    ).expression;
    const upstreamNative =
      upstream.services.esTreeNodeToTSNodeMap!.get(upstreamSequence);
    const upstreamAlias = upstream.services.tsNodeToESTreeNodeMap!.get(
      (upstreamNative as ts.BinaryExpression).left,
    );
    expect([detached.range, detached.loc]).toEqual([
      upstreamAlias.range,
      upstreamAlias.loc,
    ]);
    expect(detached.type).toBe('SequenceExpression');
    expect(detached.expressions.map((node: any) => node.name)).toEqual([
      'first',
      'second',
    ]);
    detached.expressions = [members[0]];
    expect(detached.expressions).toEqual([members[0]]);
    canonical.expressions = members;
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
    const factKey = typeQueryKey({
      ...projection.binding,
      op: 'typeRecord',
      args: { id: stringType.id },
    });
    const rawPacket = published.get(factKey)!;
    const ownedPacket = Uint8Array.from(rawPacket);
    rawPacket.fill(0);
    const beforeLocal = client.stats.localHits;
    expect(stringType.flags & ts.TypeFlags.StringLiteral).toBeTruthy();
    expect(client.stats.localHits).toBeGreaterThan(beforeLocal);
    rawPacket.set(ownedPacket);
    const lazy = new TypePacketReader(ownedPacket);
    expect(lazy.stats).toEqual({ strings: 0, lists: 0 });
    lazy.typeField('symbol');
    expect(lazy.stats).toEqual({ strings: 0, lists: 0 });
    expect(lazy.typeField('value')).toBe('\uD800');
    expect(lazy.typeField('value')).toBe('\uD800');
    expect(lazy.stats.strings).toBe(1);
    lazy.close();
    expect(lazy.retainedBytes).toBe(0);
    expect(() => lazy.typeField('symbol')).toThrow('closed');

    const symbol = services.getSymbolAtLocation(declaration.id);
    expect([symbol.name, symbol.escapedName]).toEqual(['__name', '___name']);
    expect(services.esTreeNodeToTSNodeMap.get(declaration.id).escapedText).toBe(
      '___name',
    );
    const uniqueType = services.getTypeAtLocation(variable('unique').id);
    const uniqueProperty = services
      .getTypeAtLocation(variable('uniqueKeys').id)
      .getProperties()[0];
    expect(uniqueType.flags & ts.TypeFlags.UniqueESSymbol).toBeTruthy();
    expect(uniqueType.escapedName).toBe(uniqueProperty.escapedName);
    expect(uniqueType.escapedName).toMatch(/^__@unique@\d+$/);
    const booleanType = services.getTypeAtLocation(variable('truth').id);
    expect(booleanType.intrinsicName).toBe('true');
    expect(booleanType.isLiteral()).toBe(false);
    expect(booleanType.value).toBeUndefined();
    const keyDeclaration = variable('surrogateKeys');
    const keyType = services.getTypeAtLocation(keyDeclaration.id);
    const checker = services.program.getTypeChecker();
    const keyLocation = services.esTreeNodeToTSNodeMap.get(keyDeclaration.id);
    expect(
      ['\uD800', '\uD801', '\uFFFD'].map((name) =>
        checker.typeToString(
          checker.getTypeOfSymbolAtLocation(
            checker.getPropertyOfType(keyType, name),
            keyLocation,
          ),
        ),
      ),
    ).toEqual(['1', '"two"', 'true']);

    const referenceChecker = upstream.services.program!.getTypeChecker();
    const referenceBody = upstream.ast.body as any[];
    const referenceVariable = (name: string) =>
      referenceBody
        .flatMap((node) => node.declarations ?? [])
        .find((entry) => entry.id.name === name);
    const referenceNodes = upstream.services.esTreeNodeToTSNodeMap!;
    const referenceFn = referenceNodes.get(
      referenceBody.find((node) => node.type === 'ExportNamedDeclaration')
        .declaration,
    ) as ts.FunctionDeclaration;
    const formatted = (
      typeChecker: ts.TypeChecker,
      declaration: ts.FunctionDeclaration,
      unique: ts.Type,
    ) => {
      const namespace = declaration
        .getSourceFile()
        .statements.filter(ts.isModuleDeclaration)
        .find((node) => node.name.text === 'N')!;
      const item = (namespace.body as ts.ModuleBlock).statements[1];
      const itemType = typeChecker.getTypeAtLocation(item);
      const rest = typeChecker.getTypeAtLocation(declaration.parameters[2]);
      const signature = typeChecker.getSignatureFromDeclaration(declaration)!;
      const none = ts.TypeFormatFlags.None;
      return [
        typeChecker.typeToString(unique),
        typeChecker.typeToString(unique, undefined, none),
        typeChecker.typeToString(rest, declaration, none),
        typeChecker.typeToString(
          rest,
          declaration,
          ts.TypeFormatFlags.WriteArrayAsGenericType,
        ),
        typeChecker.typeToString(itemType, item, none),
        typeChecker.typeToString(itemType, declaration, none),
        ...[ts.SignatureKind.Call, ts.SignatureKind.Construct].map((kind) =>
          typeChecker.signatureToString(
            signature,
            declaration,
            ts.TypeFormatFlags.WriteArrowStyleSignature,
            kind,
          ),
        ),
      ];
    };
    const expectedFormats = formatted(
      referenceChecker,
      referenceFn,
      referenceChecker.getTypeAtLocation(
        referenceNodes.get(referenceVariable('unique').id),
      ),
    );
    expect(new Set(expectedFormats).size).toBe(expectedFormats.length);
    expect(formatted(checker, nativeFn, uniqueType)).toEqual(expectedFormats);
    expect(() =>
      checker.typeToString(
        keyType,
        nativeFn,
        ts.TypeFormatFlags.MultilineObjectLiterals,
      ),
    ).toThrow('MultilineObjectLiterals');

    const referenceSpreadType = referenceChecker.getTypeAtLocation(
      referenceNodes.get(referenceVariable('spreadKeys').init),
    ) as ts.ObjectType;
    expect(
      referenceSpreadType.objectFlags & ts.ObjectFlags.ContainsSpread,
    ).toBe(ts.ObjectFlags.ContainsSpread);
    expect(
      services.getTypeAtLocation(variable('spreadKeys').init).objectFlags &
        ts.ObjectFlags.ContainsSpread,
    ).toBe(referenceSpreadType.objectFlags & ts.ObjectFlags.ContainsSpread);
    const truthLocation = services.esTreeNodeToTSNodeMap.get(
      variable('truth').id,
    );
    const referenceTruthLocation = referenceNodes.get(
      referenceVariable('truth').id,
    );
    for (const meaning of [
      ts.SymbolFlags.Value,
      ts.SymbolFlags.Type,
      ts.SymbolFlags.All,
    ]) {
      const actual = checker.resolveName(
        'truth',
        truthLocation,
        meaning,
        false,
      );
      const expected = referenceChecker.resolveName(
        'truth',
        referenceTruthLocation,
        meaning,
        false,
      );
      expect([actual?.name, actual?.flags]).toEqual([
        expected?.name,
        expected?.flags,
      ]);
    }
    const promise = checker.resolveName(
      'Promise',
      undefined,
      ts.SymbolFlags.All,
      false,
    );
    const referencePromise = referenceChecker.resolveName(
      'Promise',
      undefined,
      ts.SymbolFlags.All,
      false,
    );
    expect(referencePromise).toBeDefined();
    expect([promise?.name, promise?.flags]).toEqual([
      referencePromise?.name,
      referencePromise?.flags,
    ]);

    // A bulk scope result seeds a previously interned symbol without its fact packet.
    const uncached = makeClient(0);
    try {
      const checker = uncached.program.getTypeChecker();
      const source = uncached.program.getSourceFile(file)!;
      const meaning = ts.SymbolFlags.BlockScopedVariable;
      const truth = checker.resolveName('truth', source, meaning, false);
      expect(truth).toBeDefined();
      const symbols = checker.getSymbolsInScope(source, meaning);
      const before = host.requests;
      expect(symbols.find((symbol) => symbol.name === 'truth')).toBe(truth);
      expect(host.requests).toBe(before);
    } finally {
      uncached.close();
    }

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
      op: 'void' as const,
      args: {},
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
    const savedFlags = voidType.isUnion;
    expect(client.stats.factBytes).toBeLessThanOrEqual(1 << 16);
    client.close();
    expect(client.stats.factBytes).toBe(0);
    expect(client.syntaxStats.sourceBytes).toBe(0);
    expect(() => savedFlags()).toThrow('closed');
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
    expect(cache.retainedBytes).toBe(0);
    const replacement = host.request({ method: 'init', config });
    expect(replacement.generation).not.toBe(descriptor.generation);
    expect(() => host.source(projection.binding, file)).toThrow('stale');
  } finally {
    await host.close();
    rmSync(directory, { recursive: true, force: true });
  }
});

test('shared cache capacity and publication stay bounded', () => {
  const owner = new SharedTypeCacheOwner(1, 1 << 16, 8),
    reader = new SharedTypeCacheReader(owner.descriptor);
  const packet = new Uint8Array([1, 2, 3]);
  expect(owner.publish('too-large', new Uint8Array(1 << 17))).toBe(false);
  const unicode = new SharedTypeCacheOwner(2, 1 << 16, 16),
    unicodeReader = new SharedTypeCacheReader(unicode.descriptor);
  for (const [i, key] of ['\uD800', '\uD801', '\uFFFD'].entries())
    expect(unicode.publish(key, new Uint8Array([i]))).toBe(true);
  expect(
    ['\uD800', '\uD801', '\uFFFD'].map((key) => unicodeReader.get(key)![0]),
  ).toEqual([0, 1, 2]);
  unicodeReader.close();
  unicode.close();

  for (let i = 0; i < 8; i++)
    expect(owner.publish(`fact-${i}`, packet)).toBe(true);
  const used = owner.usedBytes,
    entries = owner.entries;
  expect(owner.publish('fact-0', packet)).toBe(true);
  expect([owner.usedBytes, owner.entries]).toEqual([used, entries]);
  const probes = owner.stats.probes;
  for (let i = 0; i < 100; i++)
    expect(owner.publish(`overflow-${i}`, packet)).toBe(false);
  expect(owner.stats.probes - probes).toBeLessThanOrEqual(
    100 * MAX_SHARED_TYPE_PROBES,
  );
  expect([...reader.get('fact-0')!]).toEqual([...packet]);
  reader.close();
  owner.close();
  expect([reader.retainedBytes, owner.retainedBytes]).toEqual([0, 0]);
});

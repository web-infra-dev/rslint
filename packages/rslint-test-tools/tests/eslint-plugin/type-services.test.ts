import { beforeAll, afterAll, describe, expect, test } from 'rstack/test';
import { execFileSync } from 'node:child_process';
import { mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { resolve, relative } from 'node:path';
import { Worker } from 'node:worker_threads';
import { Linter } from 'eslint';
import tsParser from '@typescript-eslint/parser';
import typedPlugin from '@typescript-eslint/eslint-plugin';
import sonarPlugin from 'eslint-plugin-sonarjs';
import effectorPlugin from 'eslint-plugin-effector';
import ts from 'typescript';
import { typeServiceCompilerData } from '../../../rslint/src/languages/js/type-service-compiler-data.js';
import {
  TypeServiceClient,
  type TypeServiceDescriptor,
} from '../../../rslint/src/languages/js/type-service.js';
import {
  SharedTypeCacheOwner,
  SharedTypeCacheReader,
} from '../../../rslint/src/languages/js/shared-type-cache.js';
import { lintFile } from '../../../rslint/src/eslint-plugin/linter/ecma-language-plugin.js';
import { decodeFrame } from '../../../rslint/src/languages/js/frame.js';
import type { ESTreeNode } from '../../../rslint/src/languages/js/types.js';
import { VISITOR_KEYS } from '../../../rslint/src/languages/js/visitor-keys.js';

const root = resolve(import.meta.dirname, '../../../..');
const enums = typeServiceCompilerData(ts);
let temp: string, binary: string;
beforeAll(() => {
  temp = mkdtempSync(resolve(tmpdir(), 'rslint-types-'));
  binary = resolve(temp, process.platform === 'win32' ? 'types.exe' : 'types');
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
  js?: boolean;
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
    name: 'promise-returning callbacks and inherited void methods',
    plugin: typedPlugin,
    rule: 'no-misused-promises',
    code: `
      declare function run(callback: () => void): void;
      run(async () => {}); run(() => {});
      interface Base { __run(): void }
      class Derived implements Base { async __run() {} }
    `,
    reports: 2,
  },
  {
    name: 'unbound JavaScript method with a JSDoc overload',
    plugin: typedPlugin,
    rule: 'unbound-method',
    js: true,
    code: `
      class Counter {
        value = 0;
        /** @overload
         * @param {number} step
         * @returns {number}
         */
        add(step) { return this.value + step; }
      }
      const counter = new Counter();
      counter.add(1);
      const detached = counter.add;
    `,
    reports: 1,
  },
  {
    name: 'returning an alias of polymorphic this',
    plugin: typedPlugin,
    rule: 'prefer-return-this-type',
    code: 'class Box { self(): Box { const self = this; return self; } fresh(): Box { return new Box(); } }',
    reports: 1,
  },
  {
    name: 'readonly properties inherited through mapped and const generic types',
    plugin: typedPlugin,
    rule: 'prefer-readonly-parameter-types',
    code: `
      type Copy<T> = { [K in keyof T]: T[K] };
      declare function capture<const T>(value: T): T;
      const inferred = capture({ x: 1 });
      function read(a: typeof inferred, b: Copy<{ readonly x: number }>, c: Copy<{ x: number }>) {}
    `,
    reports: 1,
  },
  {
    name: 'assertions that preserve or remove const generic readonlyness',
    plugin: typedPlugin,
    rule: 'no-unnecessary-type-assertion',
    code: `
      declare function capture<const T>(value: T): T;
      const inferred = capture({ x: 1 });
      const writable = inferred as { x: 1 }; writable.x = 1;
      const same = inferred as { readonly x: 1 };
      type Values<T extends unknown[]> = { [K in keyof T]: T[K] };
      function build<T extends unknown[]>() { return [] as Values<T>; }
    `,
    reports: 1,
  },
  {
    name: 'built-in argument types use fully qualified symbols',
    plugin: sonarPlugin,
    rule: 'argument-type',
    code: '"text".includes(123); "text".includes("t");',
    reports: 1,
  },
  {
    name: 'deprecated call suggestion uses UTF-16 spans',
    plugin: sonarPlugin,
    rule: 'deprecation',
    code: '/** @deprecated Use fresh instead. */\nfunction oldFn() {}\nconst emoji = "😀"; oldFn();',
    reports: 1,
  },
  {
    name: 'generated nodes distinguish function and array returns',
    plugin: sonarPlugin,
    rule: 'function-return-type',
    code: 'function value(b: boolean) { if (b) return [1]; return () => 1; }',
    reports: 1,
  },
  {
    name: 'ignored replace result distinguishes a callback from replacement text',
    plugin: sonarPlugin,
    rule: 'no-ignored-return',
    code: '"abc".replace("a", "d"); let n=0; "abc".replace("a", () => String(++n));',
    reports: 1,
  },
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
  },
  {
    name: 'type parameters in empty and materialized mapped types',
    plugin: typedPlugin,
    rule: 'no-unnecessary-type-parameters',
    code: `
      declare function keys<T>(): { [K in keyof T]: T[K] };
      declare function record<T>(): { [K in 'key']: T };
      declare function empty<T>(): { [K in never]: T };
      declare function single<T>(value: T): void;
    `,
    reports: 3,
  },
  {
    name: 'deprecated overloads and properties reached through inheritance',
    plugin: typedPlugin,
    rule: 'no-deprecated',
    code: `
      import { pick, Contract } from './dependency';
      import { Inherits, Own } from './legacy.js';
      class Implementation implements Contract { value = 1; }
      pick('retired');
      pick(42);
      new Implementation().value;
      new Inherits().read();
      new Own().read();
      void import.meta;
      function target() { return new.target; }
    `,
    dependencies: {
      'dependency.ts': `
        export declare function replacement(): void;
        /** @deprecated Use {@link replacement}. */
        export declare function pick(input: string): void;
        export declare function pick(input: number): void;
        export interface Contract {
          /** @deprecated */
          value: number;
        }
      `,
      'legacy.js': `
        class Base {
          /** @overload
           * @returns {number}
           */
          /** @deprecated Use fresh. */
          read() { return 1; }
        }
        export class Inherits extends Base { read() { return 2; } }
        export class Own extends Base {
          /** @overload
           * @returns {number}
           */
          read() { return 2; }
        }
      `,
    },
    reports: 3,
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
  },
  {
    name: 'collection size semantics',
    plugin: sonarPlugin,
    rule: 'no-collection-size-mischeck',
    code: 'declare const list: number[]; declare const set: Set<number>; declare const other: { length: number }; if (list.length < 0) {} if (set.size < 0) {} if (other.length < 0) {}',
    reports: 2,
  },
  {
    name: 'SonarJS recognizes JSDoc-annotated JavaScript constructors',
    plugin: sonarPlugin,
    rule: 'new-operator-misuse',
    options: [{ considerJSDoc: true }],
    js: true,
    code: `
      /** @constructor Creates a box. */
      function Box() { this.value = 1; }
      function compute() { return 1; }
      new Box();
      new compute();
    `,
    reports: 1,
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
  },
  {
    name: 'Effector contextual component signature',
    plugin: effectorPlugin,
    rule: 'mandatory-scope-binding',
    code: 'import type { FC } from "react"; import { createEvent } from "effector"; const clicked = createEvent(); const Button: FC = () => { clicked(); return null; }; declare function render(callback: () => null): void; render(() => null);',
    reports: 1,
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
  let host: Host;
  beforeAll(() => {
    host = new Host();
  });
  afterAll(() => host.close());

  for (const c of cases)
    test(c.name, () => {
      // Keep fixtures below the workspace so both compilers resolve the exact
      // installed npm declarations, without copying or faking library types.
      const directory = mkdtempSync(
        resolve(root, 'packages/rslint-test-tools/.type-services-'),
      );
      const extension = c.js ? 'js' : 'ts';
      const file = resolve(directory, `input.${extension}`),
        config = resolve(directory, 'tsconfig.json');
      writeFileSync(file, c.code);
      for (const [name, text] of Object.entries(c.dependencies ?? {}))
        writeFileSync(resolve(directory, name), text);
      writeFileSync(
        config,
        JSON.stringify({
          compilerOptions: {
            allowJs:
              c.js ||
              Object.keys(c.dependencies ?? {}).some((f) => f.endsWith('.js')),
            checkJs: c.js,
            strict: true,
            target: 'ES2022',
            module: 'ESNext',
            moduleResolution: 'Bundler',
            skipLibCheck: true,
          },
          include: [`*.${extension}`],
        }),
      );
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
              files: [`**/*.${extension}`],
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
        const client = new TypeServiceClient(
          enums,
          descriptor,
          (query) =>
            Buffer.from(host.request({ method: 'query', query }), 'base64'),
          host.source,
          { maxFactsBytes: 1 << 20 },
        );
        try {
          const handle = client.file(projection.binding);
          const result = lintFile(request, { rules } as any, (decoded) =>
            handle.parserServices(
              decoded,
              projection,
              projection.binding.sourceFileName,
            ),
          );
          expect(result.parseError).toBeUndefined();
          expect(result.ruleErrors).toBeUndefined();
          expect(normResult(result, c.code)).toEqual(normMessages(expected));
        } finally {
          client.close();
        }
      } finally {
        rmSync(directory, { recursive: true, force: true });
      }
    });
});

test('type and declaration identity survive closing one file', async () => {
  const directory = mkdtempSync(
    resolve(root, 'packages/rslint-test-tools/.type-services-'),
  );
  const config = resolve(directory, 'tsconfig.json');
  const code =
    'import { value } from "./dependency"; const a = value, b = value;';
  for (const name of ['first.ts', 'second.ts'])
    writeFileSync(resolve(directory, name), code);
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
    let syntaxReads = 0;
    const owner = new TypeServiceClient(
      enums,
      descriptor,
      (query) =>
        Buffer.from(host.request({ method: 'query', query }), 'base64'),
      (scope, file, metadataOnly) => {
        if (!metadataOnly) syntaxReads++;
        return host.source(scope, file, metadataOnly);
      },
      {
        maxFactsBytes: 1 << 20,
        maxSyntaxBytes: 0,
        batchTransport: (queries) =>
          host
            .request({ method: 'batch', queries })
            .map((packet: string) => Buffer.from(packet, 'base64')),
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
      const handle = owner.file(projection.binding);
      const services = handle.parserServices(frame, projection, file);
      if (services.program === null)
        throw new Error('expected project-backed parser services');
      return { file, frame, projection, handle, services };
    };
    try {
      const first = open('first.ts');
      const declarations = (first.frame.ast.body as any[])[1].declarations;
      const firstNode = first.services.esTreeNodeToTSNodeMap.get(
        declarations[0].id,
      ) as ts.Identifier;
      expect(firstNode.getText()).toBe('a');
      expect(syntaxReads).toBe(0);
      const firstType = first.services.getTypeAtLocation(declarations[0].id)!;
      expect(firstType.flags & ts.TypeFlags.NumberLiteral).toBeTruthy();
      const imported = ((first.frame.ast.body as any[])[0] as any).specifiers[0]
        .local;
      const alias = first.services.getSymbolAtLocation(imported)!;
      const symbol = owner.checker.getAliasedSymbol(alias)!;
      const declaration = symbol.declarations![0] as any;
      const source = declaration.getSourceFile();
      expect(source).toBe(
        owner.program.getSourceFile(resolve(directory, 'dependency.ts')),
      );
      expect(
        owner.program.getSourceFile(
          relative(descriptor.currentDirectory, first.file),
        ),
      ).toBe(owner.program.getSourceFile(first.file));
      first.handle.close();
      expect(() => first.services.esTreeNodeToTSNodeMap.get(imported)).toThrow(
        'closed',
      );
      const second = open('second.ts');
      const secondDeclarations = (second.frame.ast.body as any[])[1]
        .declarations as { id: ESTreeNode }[];
      owner.prefetch(
        secondDeclarations.map((declaration) => ({
          op: 'nodeRecord',
          args: {
            node: {
              file: second.projection.sourceFileName,
              index:
                second.projection.origins[
                  second.frame.nodes.indexOf(declaration.id)
                ],
            },
          },
        })),
      );
      for (const node of secondDeclarations) {
        const identifier = second.services.esTreeNodeToTSNodeMap.get(
          node.id,
        ) as ts.Identifier;
        expect(identifier.text).toBe(node.id.name);
        expect(second.services.getTypeAtLocation(node.id)).toBe(firstType);
      }
      expect(
        owner.checker.getAliasedSymbol(
          second.services.getSymbolAtLocation(
            ((second.frame.ast.body as any[])[0] as any).specifiers[0].local,
          )!,
        ),
      ).toBe(symbol);
      expect(declaration.getText()).toBe('value=1 as const');
      expect(declaration.getSourceFile()).toBe(source);
      const getProperties = firstType.getProperties.bind(firstType);
      getProperties();
      second.handle.close();
      owner.close();
      expect(getProperties).toThrow('closed');
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

function readMapping(host: Host, directory: string) {
  const file = resolve(directory, 'mapping.tsx'),
    config = resolve(directory, 'tsconfig.json');
  const longTypeName = '例'.repeat(110);
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
const multiline = { nested: { label: "中文" as const }, count: 1 };
function inferredScalar() { return '中文' as const; }
function inferredObject() { return { label: '中文' as const }; }
interface ${longTypeName} { value: number }
declare function multilineCall(
  { '\\é': escaped }: { '\\é': string },
  value: { label: "中文"; escaped: "\\é"; slash: "\\\\é" },
): { nested: { label: "中文" }; count: number };
/** Ordinary documentation. */
function docsTarget() {}
async function tokenContext() { await Promise.resolve(1); }
interface DocsBase {
  /** Base documentation linking to {@linkcode docsTarget | example}.
   * @param value Value to use.
   * @deprecated Replace this with {@link docsTarget}.
   */
  method(value: string): void;
}
class DocsChild implements DocsBase {
  /** @inheritDoc Additional detail. */
  method(value: string) {}
}
const docAssertion = /** @type {{ label: string }} */ ({ label: "temporary" });
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
  const descriptor = host.request({
    method: 'init',
    config,
  }) as TypeServiceDescriptor;
  const encoded = host.request({ method: 'project', file });
  const projection = {
    ...encoded,
    tree: Buffer.from(encoded.tree, 'base64'),
  };
  const sourceFileName: string = projection.binding.sourceFileName;
  const bytes = Buffer.from(encoded.frame, 'base64');
  const frame = decodeFrame(bytes);
  const client = new TypeServiceClient(
    enums,
    descriptor,
    (query) => Buffer.from(host.request({ method: 'query', query }), 'base64'),
    host.source,
    { maxFactsBytes: 1 << 16 },
  );
  const program = client.program as unknown as ts.Program;
  const beforeSource = program.getSourceFile(file);
  if (!beforeSource) throw new Error(`missing project source: ${file}`);
  const services = client
    .file(projection.binding)
    .parserServices(frame, projection, sourceFileName) as any;
  const upstream = tsParser.parseForESLint(code, {
    filePath: file,
    project: config,
    tsconfigRootDir: directory,
    range: true,
    loc: true,
  });
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
  return {
    client,
    file,
    bytes,
    directory,
    longTypeName,
    beforeSource,
    frame,
    services,
    upstream,
    body,
    variable,
    expression,
    fn,
    nativeFn,
    referenceChecker,
    referenceVariable,
    referenceNodes,
    referenceFn,
    checker: services.program.getTypeChecker(),
  };
}

describe('TypeScript node maps and type queries', () => {
  let directory: string, host: Host, mapping: ReturnType<typeof readMapping>;
  beforeAll(() => {
    directory = mkdtempSync(
      resolve(root, 'packages/rslint-test-tools/.type-services-'),
    );
    host = new Host();
    mapping = readMapping(host, directory);
  });
  afterAll(async () => {
    mapping?.client.close();
    await host.close();
    rmSync(directory, { recursive: true, force: true });
  });

  test('node maps, tokens and aliases match TypeScript', () => {
    const {
      beforeSource,
      frame,
      services,
      upstream,
      expression,
      fn,
      nativeFn,
    } = mapping;
    expect(services.esTreeNodeToTSNodeMap.has(frame.ast)).toBe(true);
    const collect = (ast: any, maps: any) => {
      const entries: any[] = [];
      const walk = (node: any, path: string) => {
        const native = maps.esTreeNodeToTSNodeMap.get(node);
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
    const nativeAwait = services.esTreeNodeToTSNodeMap.get(
      expression('AwaitExpression'),
    );
    expect(nativeAwait.getChildAt(1)).toBe(nativeAwait.expression);
    expect(nativeAwait.getChildAt(1).kind).toBe(ts.SyntaxKind.CallExpression);
    const awaitToken = nativeAwait.getChildAt(0);
    expect(awaitToken.kind).toBe(ts.SyntaxKind.AwaitKeyword);
    expect(awaitToken.parent).toBe(nativeAwait);
    expect(awaitToken.getText()).toBe('await');
    expect(nativeAwait.getFirstToken()).toBe(awaitToken);
    const call = nativeAwait.expression;
    const argumentsList = call.getChildAt(2);
    expect(argumentsList.kind).toBe(ts.SyntaxKind.SyntaxList);
    expect(argumentsList.parent).toBe(call);
    expect(argumentsList.getChildAt(0)).toBe(call.arguments[0]);
    expect(call.getChildCount()).toBe(4);
    expect(call.getChildren()).toBe(call.getChildren());
    expect(call.getLastToken()).toBe(call.getChildAt(3));
    expect(call.getLastToken().getSourceFile()).toBe(beforeSource);
    expect(services.tsNodeToESTreeNodeMap.has(argumentsList)).toBe(false);
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
    const detached = services.tsNodeToESTreeNodeMap.get(sequence.left);
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
  });

  test('parser services expose compiler options and unmapped type annotations', () => {
    const { directory, file, services, upstream, fn, nativeFn } = mapping;
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
    const annotation = fn.params[0].typeAnnotation;
    expect(services.esTreeNodeToTSNodeMap.has(annotation)).toBe(false);
    expect(services.getSymbolAtLocation(annotation)).toBeUndefined();
    expect(services.getTypeAtLocation(annotation).flags).toBe(ts.TypeFlags.Any);
  });

  test('lintFile returns a file error when parser services fail to initialize', () => {
    const { file, bytes } = mapping;
    const result = lintFile(
      {
        filePath: file,
        frame: bytes,
        rules: {},
        collectFixes: false,
        suggestionsMode: 'off',
      },
      { plugins: [], rules: new Map() },
      () => {
        throw new Error('stale file binding');
      },
    );
    expect(result.parseError).toBe('parser services: stale file binding');
  });

  test('literal types and symbol lookup match TypeScript', () => {
    const {
      body,
      expression,
      variable,
      services,
      checker,
      referenceChecker,
      referenceNodes,
      referenceVariable,
    } = mapping;
    const declaration = variable('__name');
    const stringType = services.getTypeAtLocation(declaration.id);
    expect(stringType.value).toBe('\uD800');
    expect(stringType.flags & ts.TypeFlags.StringLiteral).toBeTruthy();

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
  });

  test('generated type syntax and compiler node flags match TypeScript', () => {
    const {
      beforeSource,
      variable,
      services,
      checker,
      referenceChecker,
      referenceVariable,
      referenceNodes,
      referenceFn,
    } = mapping;
    const id = variable('surrogateKeys').id;
    const keyType = services.getTypeAtLocation(id);
    const keyLocation = services.esTreeNodeToTSNodeMap.get(id);
    // The node builder returns detached syntax. Its public kinds, child fields
    // and visitation must agree even when the compiler enums differ.
    const generatedShape = (n: ts.Node): unknown => {
      const children: unknown[] = [];
      n.forEachChild((child) => {
        children.push(generatedShape(child));
      });
      return [
        n.kind,
        // tsgo's node builder can reuse a source literal and its range.
        // For these leaves compare the literal's value as consumed by rules.
        ts.isStringLiteral(n) || ts.isNumericLiteral(n)
          ? n.text
          : [n.pos, n.end],
        ts.isIdentifier(n) ? n.escapedText : undefined,
        children,
      ];
    };
    const generated = checker.typeToTypeNode(
      keyType,
      keyLocation,
      ts.NodeBuilderFlags.InTypeAlias,
    )!;
    const referenceLocation = referenceNodes.get(
      referenceVariable('surrogateKeys').id,
    )!;
    expect(generatedShape(generated)).toEqual(
      generatedShape(
        referenceChecker.typeToTypeNode(
          referenceChecker.getTypeAtLocation(referenceLocation),
          referenceLocation,
          ts.NodeBuilderFlags.InTypeAlias,
        )!,
      ),
    );
    expect(generated.getSourceFile()).toBeUndefined();
    expect('referencedFiles' in generated).toBe(false);
    const contextualToken = (source: ts.SourceFile) => {
      const fn = source.statements.find(
        (node): node is ts.FunctionDeclaration =>
          ts.isFunctionDeclaration(node) && node.name?.text === 'tokenContext',
      )!;
      return (
        fn.body!.statements[0] as ts.ExpressionStatement
      ).expression.getFirstToken()!;
    };
    const contextFlags = contextualToken(referenceFn.getSourceFile()).flags;
    expect(contextFlags & ts.NodeFlags.ContextFlags).toBeTruthy();
    expect(contextualToken(beforeSource).flags).toBe(contextFlags);
    const flagsAndModifiers = (source: ts.SourceFile) => {
      const samples: number[] = [];
      const visit = (n: ts.Node) => {
        if (
          ts.isModuleDeclaration(n) ||
          ts.isMethodSignature(n) ||
          (ts.isPropertyAccessExpression(n) && n.questionDotToken) ||
          (ts.isFunctionDeclaration(n) && n.name?.text === 'docsTarget')
        )
          samples.push(n.flags);
        if (ts.isModuleDeclaration(n))
          samples.push(ts.getModifiers(n)?.length ?? 0);
        n.forEachChild(visit);
      };
      visit(source);
      return samples;
    };
    expect(flagsAndModifiers(beforeSource)).toEqual(
      flagsAndModifiers(referenceFn.getSourceFile()),
    );
  });

  test('type and signature formatting matches TypeScript', () => {
    const {
      beforeSource,
      variable,
      services,
      checker,
      nativeFn,
      referenceChecker,
      referenceNodes,
      referenceVariable,
      referenceFn,
      longTypeName,
    } = mapping;
    const uniqueType = services.getTypeAtLocation(variable('unique').id);
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
    expect(formatted(checker, nativeFn, uniqueType)).toEqual(expectedFormats);
    const multilineFormats = (
      typeChecker: ts.TypeChecker,
      sourceFile: ts.SourceFile,
      object: ts.Node,
    ) => {
      const declaration = sourceFile.statements.find(
        (node): node is ts.FunctionDeclaration =>
          ts.isFunctionDeclaration(node) && node.name?.text === 'multilineCall',
      )!;
      const type = typeChecker.getTypeAtLocation(object);
      const signature = typeChecker.getSignatureFromDeclaration(declaration)!;
      const longType = sourceFile.statements.find(
        (node): node is ts.InterfaceDeclaration =>
          ts.isInterfaceDeclaration(node) && node.name.text === longTypeName,
      )!;
      const multiline = ts.TypeFormatFlags.MultilineObjectLiterals;
      const flags = multiline | ts.TypeFormatFlags.NoTruncation;
      return [
        typeChecker.typeToString(typeChecker.getTypeAtLocation(longType.name)),
        typeChecker.typeToString(
          typeChecker.getTypeAtLocation(declaration.name!),
          declaration,
        ),
        typeChecker.signatureToString(signature, declaration),
        ...sourceFile.statements
          .filter(ts.isFunctionDeclaration)
          .filter((node) => node.name?.text.startsWith('inferred'))
          .map((node) =>
            typeChecker.signatureToString(
              typeChecker.getSignatureFromDeclaration(node)!,
              node,
            ),
          ),
        typeChecker.typeToString(type, declaration, flags),
        ...[ts.SignatureKind.Call, ts.SignatureKind.Construct].map((kind) =>
          typeChecker.signatureToString(
            signature,
            declaration,
            flags | ts.TypeFormatFlags.WriteArrowStyleSignature,
            kind,
          ),
        ),
        typeChecker.signatureToString(
          signature,
          declaration,
          flags | ts.TypeFormatFlags.UseSingleQuotesForStringLiteralType,
          ts.SignatureKind.Call,
        ),
      ];
    };
    expect(
      multilineFormats(
        checker,
        beforeSource,
        services.esTreeNodeToTSNodeMap.get(variable('multiline').id),
      ),
    ).toEqual(
      multilineFormats(
        referenceChecker,
        referenceFn.getSourceFile(),
        referenceNodes.get(referenceVariable('multiline').id),
      ),
    );
  });

  test('symbol and signature documentation matches TypeScript', () => {
    const {
      beforeSource,
      variable,
      services,
      checker,
      referenceChecker,
      referenceNodes,
      referenceVariable,
      referenceFn,
    } = mapping;
    const readDocumentation = (
      typeChecker: ts.TypeChecker,
      source: ts.SourceFile,
    ) => {
      const owner = source.statements.find(
        (node): node is ts.ClassDeclaration =>
          ts.isClassDeclaration(node) && node.name?.text === 'DocsChild',
      )!;
      const method = owner.members[0] as ts.MethodDeclaration;
      const symbol = typeChecker.getSymbolAtLocation(method.name)!;
      const signature = typeChecker.getSignatureFromDeclaration(method)!;
      const documentation = symbol.getDocumentationComment(typeChecker);
      const tags = symbol.getJsDocTags(typeChecker);
      expect(symbol.getDocumentationComment(typeChecker)).toBe(documentation);
      expect(symbol.getJsDocTags(typeChecker)).toBe(tags);
      expect(documentation.map((part) => part.kind)).toContain('linkName');
      return [
        documentation,
        tags,
        signature.getDocumentationComment(typeChecker),
        signature.getJsDocTags(),
      ];
    };
    expect(readDocumentation(checker, beforeSource)).toEqual(
      readDocumentation(referenceChecker, referenceFn.getSourceFile()),
    );
    const symbolTags = (typeChecker: ts.TypeChecker, node: ts.Node) =>
      typeChecker.getSymbolAtLocation(node)!.getJsDocTags(typeChecker);
    expect(
      symbolTags(
        checker,
        services.esTreeNodeToTSNodeMap.get(variable('docAssertion').id),
      ),
    ).toEqual(
      symbolTags(
        referenceChecker,
        referenceNodes.get(referenceVariable('docAssertion').id),
      ),
    );
  });
});

test('signature options preserve TypeScript query order across clients', async () => {
  const directory = mkdtempSync(resolve(temp, 'signature-options-'));
  const file = resolve(directory, 'input.ts');
  const config = resolve(directory, 'tsconfig.json');
  writeFileSync(
    file,
    `
    type Picked<T> = { [K in keyof T]: T[K] };
    declare function choose<T extends string>(value: T): { x: 1 };
    declare function choose<T extends string>(value: T, extra: T): { x: 2 };
    ({ x } = choose(false,));
    const fresh = choose(false,);
    declare let x: number;
    type Used = Picked<{ p: string }>;
  `,
  );
  const compilerOptions = { strict: true, noLib: true };
  writeFileSync(
    config,
    JSON.stringify({ compilerOptions, files: ['input.ts'] }),
  );
  const host = new Host();
  const assignment = (source: ts.SourceFile) =>
    (
      (source.statements[3] as ts.ExpressionStatement)
        .expression as ts.ParenthesizedExpression
    ).expression as ts.BinaryExpression;
  const exercise = (program: ts.Program, peer = program) => {
    const checker = program.getTypeChecker(),
      peerChecker = peer.getTypeChecker();
    const source = program.getSourceFile(file)!,
      peerSource = peer.getSourceFile(file)!;
    const call = assignment(source).right as ts.CallExpression,
      peerCall = assignment(peerSource).right as ts.CallExpression;
    const pattern = assignment(peerSource).left as ts.ObjectLiteralExpression;
    const propertyName = (
      pattern.properties[0] as ts.ShorthandPropertyAssignment
    ).name;
    const firstCall = (source.statements[4] as ts.VariableStatement)
      .declarationList.declarations[0].initializer as ts.CallExpression;
    const returned = (signature: ts.Signature | undefined, owner = checker) =>
      owner.typeToString(signature!.getReturnType());
    const read = () => [
      returned(peerChecker.getResolvedSignature(peerCall), peerChecker),
      peerChecker.typeToString(peerChecker.getTypeAtLocation(peerCall)),
      peerChecker.typeToString(peerChecker.getTypeOfAssignmentPattern(pattern)),
      peerChecker.typeToString(
        peerChecker.getTypeOfSymbol(
          peerChecker.getPropertySymbolOfDestructuringAssignment(propertyName)!,
        ),
      ),
    ];
    const withCandidates = (count: number) => {
      const candidates: ts.Signature[] = [];
      const chosen = checker.getResolvedSignature(call, candidates, count);
      return [
        returned(chosen),
        candidates.map((signature) => returned(signature)),
        candidates.indexOf(chosen!),
      ];
    };
    return [
      read(),
      returned(checker.getResolvedSignature(firstCall, undefined, 2)),
      returned(checker.getResolvedSignature(call, undefined, 2)),
      withCandidates(2),
      read(),
      returned(checker.getResolvedSignature(call, undefined, 2)),
      withCandidates(1),
      read(),
      withCandidates(2),
      read(),
    ];
  };
  const expected = exercise(ts.createProgram([file], compilerOptions));
  try {
    const descriptor: TypeServiceDescriptor = host.request({
      method: 'init',
      config,
    });
    const open = () =>
      new TypeServiceClient(
        enums,
        descriptor,
        (query) =>
          Buffer.from(host.request({ method: 'query', query }), 'base64'),
        host.source,
        { maxFactsBytes: 1 << 16 },
      );
    const owner = open(),
      peer = open();
    try {
      expect(
        exercise(
          owner.program as unknown as ts.Program,
          peer.program as unknown as ts.Program,
        ),
      ).toEqual(expected);
      const source = owner.program.getSourceFile(file) as ts.SourceFile;
      const checker = owner.checker;
      const call = assignment(source).right,
        signature = checker.getResolvedSignature(call)!;
      expect(() => checker.getResolvedSignature(call, [signature])).toThrow(
        'must be empty',
      );

      const mapped = checker.getTypeFromTypeNode(
        (source.statements[0] as ts.TypeAliasDeclaration).type,
      )!;
      expect('typeParameter' in mapped).toBe(true);
      expect('typeParameter' in checker.getTypeAtLocation(call)!).toBe(false);
      expect(mapped.typeParameter?.symbol?.name).toBe('K');
      expect(checker.typeToString(mapped.constraintType!)).toBe('keyof T');
      expect(mapped.templateType).toBeUndefined();
      const used = checker.getTypeFromTypeNode(
        (source.statements[6] as ts.TypeAliasDeclaration).type,
      )!;
      checker.getTypeOfPropertyOfType(used, 'p');
      expect(mapped.templateType?.indexType).toBe(mapped.typeParameter);
    } finally {
      owner.close();
      peer.close();
    }
  } finally {
    await host.close();
  }
});

test('public checker, Program and source reads match TypeScript', async () => {
  const dir = mkdtempSync(resolve(temp, 'public-'));
  const file = resolve(dir, 'input.ts'),
    config = resolve(dir, 'tsconfig.json');
  const code =
    [
      "export enum Values { N = 3, S = 'three', C = id(3) }",
      'export type Pair = readonly [first: string, second?: number];',
      'export function id<T extends { valueOf?<U,>(): number }>(input: T): T { return input; }',
      'export const result = id(42);',
      "export const incorrect: number = 'bad';",
      'export type Dictionary = { [name: string]: number };',
      "export function isString(input: unknown): input is string { return typeof input === 'string'; }",
      'export type Select<T> = T extends string ? true : false;',
      'export class Box { #value = 0; constructor(public label: string) {} read() { return this.#value; } [computedKey]() { return 1; } }',
      'export { id as alias };',
      'export const record = { n: 1 }; export const {n} = record;',
      "import type { External } from './external' with { 'resolution-mode': 'import' }; export type Used = External;",
      'export const grow = () => 1; grow.item = 1;',
      'declare const computedKey: string;',
      "export * from './external';",
      'declare const condition: boolean; condition;',
    ].join('\r\n') + '\r\n';
  writeFileSync(file, code);
  writeFileSync(resolve(dir, 'external.ts'), 'export type External = string;');
  writeFileSync(
    config,
    JSON.stringify({
      compilerOptions: { strict: true, target: 'esnext', module: 'esnext' },
      files: ['input.ts'],
    }),
  );
  const host = new Host();
  const descriptor = host.request({ method: 'init', config });
  const owner = new TypeServiceClient(
    enums,
    descriptor,
    (query) => Buffer.from(host.request({ method: 'query', query }), 'base64'),
    host.source,
    { maxFactsBytes: 1 << 18 },
  );
  try {
    const parsed = ts.getParsedCommandLineOfConfigFile(
      config,
      {},
      { ...ts.sys, onUnRecoverableConfigFileDiagnostic() {} },
    )!;
    const reference = ts.createProgram(parsed.fileNames, parsed.options);
    const exercise = (p: ts.Program) => {
      const c = p.getTypeChecker(),
        src = p.getSourceFile(file)!;
      const [enm, pair, fn, result, , dict, guard, select] = src.statements;
      const decl = fn as ts.FunctionDeclaration,
        sig = c.getSignatureFromDeclaration(decl)!;
      const generatedSignature = c.signatureToSignatureDeclaration(
        sig,
        ts.SyntaxKind.CallSignature,
        decl,
        0,
      )!;
      const generatedParameter = generatedSignature.parameters[0];
      const constraint = generatedSignature.typeParameters![0]
        .constraint as ts.TypeLiteralNode;
      const detached = c.typeToTypeNode(
        c.getTypeAtLocation(decl.name!),
        undefined,
        0,
      ) as ts.FunctionTypeNode;
      const detachedConstraint = detached.typeParameters![0]
        .constraint as ts.TypeLiteralNode;
      const param = decl.parameters[0],
        symbol = c.getSymbolAtLocation(param.name)!;
      const call = (result as ts.VariableStatement).declarationList
        .declarations[0].initializer as ts.CallExpression;
      const t = c.getTypeFromTypeNode(
        (pair as ts.TypeAliasDeclaration).type,
      ) as ts.TupleTypeReference;
      const tuple = t.target;
      const index = c.getIndexInfosOfType(
        c.getTypeFromTypeNode((dict as ts.TypeAliasDeclaration).type),
      )[0];
      const predicate = c.getTypePredicateOfSignature(
        c.getSignatureFromDeclaration(guard as ts.FunctionDeclaration)!,
      )!;
      const cond = c.getTypeFromTypeNode(
        (select as ts.TypeAliasDeclaration).type,
      ) as ts.ConditionalType;
      const format = (value: ts.Type) => c.typeToString(value);
      const name = (node: ts.Node | undefined) =>
        node && ts.SyntaxKind[node.kind];
      const primitive = [
        c.getAnyType(),
        c.getUnknownType(),
        c.getStringType(),
        c.getNumberType(),
        c.getBigIntType(),
        c.getBooleanType(),
        c.getTrueType(),
        c.getFalseType(),
        c.getESSymbolType(),
        c.getNonPrimitiveType(),
        c.getNeverType(),
        c.getNullType(),
        c.getUndefinedType(),
      ];
      const literal = c.getStringLiteralType('abc');
      const metadata = Symbol();
      let getterCalls = 0;
      Object.defineProperty(literal, metadata, {
        get() {
          getterCalls++;
        },
      });
      const wrapped = new Proxy(t, {});
      const computed = c.getTypeAtLocation(
        (enm as ts.EnumDeclaration).members[2].name,
      ) as ts.EnumType;
      const boolean = c.getBooleanType();
      const regularBefore = Reflect.get(boolean, 'regularType') === boolean;
      c.getTypeAtLocation(
        (src.statements.at(-1) as ts.ExpressionStatement).expression,
      );
      return {
        additional: (() => {
          const fnSymbol = c.getSymbolAtLocation(decl.name!)!;
          const box = src.statements[8] as ts.ClassDeclaration;
          const ctor = box.members[1] as ts.ConstructorDeclaration;
          const access = (
            (box.members[2] as ts.MethodDeclaration).body!
              .statements[0] as ts.ReturnStatement
          ).expression as ts.PropertyAccessExpression;
          const left = c.getTypeAtLocation(access.expression);
          const fileSymbol = c.getSymbolAtLocation(src)!;
          const factoryName = ts.factory.createIdentifier('input');
          const originalName = ts.setOriginalNode(
            ts.factory.createIdentifier('input'),
            param.name,
          );
          const exportSpec = (
            (src.statements[9] as ts.ExportDeclaration)
              .exportClause as ts.NamedExports
          ).elements[0];
          const dictionary = c.getSymbolAtLocation(
            (dict as ts.TypeAliasDeclaration).name,
          )!;
          const dictType = c.getDeclaredTypeOfSymbol(dictionary);
          const indexSymbol = dictType
            .getSymbol()!
            .members!.get('__index' as ts.__String)!;
          const computed = c.getSymbolAtLocation(box.members[3].name!)!;
          const imp = src.statements[12] as ts.ImportDeclaration;
          const grow = (src.statements[14] as ts.VariableStatement)
            .declarationList.declarations[0].initializer!;
          return {
            ambient: c.getAmbientModules().length,
            expando: c.getSymbolOfExpando(grow, false)?.name,
            entities: [
              name(
                c.symbolToExpression(fnSymbol, ts.SymbolFlags.Value, decl, 0),
              ),
              name(
                c.symbolToEntityName(fnSymbol, ts.SymbolFlags.Value, decl, 0),
              ),
            ],
            parameters: c
              .symbolToTypeParameterDeclarations(fnSymbol, decl, 0)
              ?.map((node) => node.name.text),
            aliasParameters: c
              .symbolToTypeParameterDeclarations(
                c.getSymbolAtLocation(
                  (select as ts.TypeAliasDeclaration).name,
                )!,
                decl,
                0,
              )
              ?.map((n) => n.name.text),
            emptyParameters: c
              .symbolToTypeParameterDeclarations(dictionary, decl, 0)
              ?.map((n) => n.name.text),
            private: c.getPrivateIdentifierPropertyOfType(
              left,
              '#value',
              access,
            )?.name,
            valid: c.isValidPropertyAccess(access, 'read'),
            parameterProperty: c
              .getSymbolsOfParameterPropertyDeclaration(
                ctor.parameters[0],
                'label',
              )
              .map((s) => s.name),
            local: c.getExportSpecifierLocalTargetSymbol(exportSpec)?.name,
            exp: c.tryGetMemberInModuleExports('alias', fileSymbol)?.name,
            // The native checker and TypeScript 6 enumerate symbol tables in different orders.
            rawExports: [...(fileSymbol.exports?.keys() ?? [])].sort(),
            exports: c
              .getExportsOfModule(fileSymbol)
              .map((symbol) => symbol.name)
              .sort(),
            indices: c
              .getIndexInfosOfIndexSymbol(indexSymbol)
              .map((i) => format(i.type)),
            computedIndices: [undefined, []].map((siblings) =>
              c
                .getIndexInfosOfIndexSymbol(computed, siblings)
                .map((info) => format(info.type)),
            ),
            generatedType: [
              c.getSymbolAtLocation(factoryName)?.name,
              format(c.getTypeAtLocation(originalName)),
              c.getTypeFromTypeNode(
                c.typeToTypeNode(c.getStringType(), undefined, undefined)!,
              ) === c.getAnyType(),
              format(c.getTypeAtLocation(generatedParameter.name)),
              c.getSymbolAtLocation(generatedParameter.name)?.name,
              ts.getParseTreeNode(generatedParameter.name) === param.name,
              constraint
                .getChildren(src)
                .map((child) => [child.kind, child.pos, child.end]),
              (constraint.members[0] as ts.MethodSignature).questionToken?.kind,
              (detachedConstraint.members[0] as ts.MethodSignature)
                .typeParameters?.hasTrailingComma,
              c.isOptionalParameter(generatedParameter),
              c.getSymbolAtLocation(
                ts.setOriginalNode(generatedParameter.name, decl.name),
              )?.name,
              c.getSymbolAtLocation(
                ts.setOriginalNode(generatedParameter.name, undefined),
              )?.name,
            ],
            modes: [
              imp.attributes?.elements[0].value.getText(src),
              p.getModeForUsageLocation(
                src,
                imp.moduleSpecifier as ts.StringLiteral,
              ),
              p.getModeForResolutionAtIndex(src, 0),
            ],
          };
        })(),
        shape: [
          [Object.hasOwn(literal, metadata), metadata in literal, getterCalls],
          wrapped.getProperties() === t.getProperties(),
          wrapped.typeArguments?.map(format),
          computed.regularType?.freshType === computed,
          [regularBefore, Reflect.get(boolean, 'regularType') === boolean],
          'types' in literal,
          'target' in c.getStringType(),
          'target' in literal,
          'thisType' in t,
          tuple.thisType && format(tuple.thisType),
          'types' in c.getBooleanType(),
          Reflect.get(c, 'getOptionalExtension'),
          Reflect.get(src, 'getOptionalExtension'),
          'root' in cond,
          'root' in literal,
          'elementFlags' in tuple,
          'elementFlags' in t,
          'texts' in literal,
        ],
        primitive: primitive.map(format),
        interned: c.getStringType() === c.getStringType(),
        literals: [
          literal.value,
          c.getNumberLiteralType(-2.5).value,
          c.getBigIntLiteralType({ negative: true, base10Value: '42' }).value,
        ],
        nullable: format(
          c.getNullableType(
            c.getStringType(),
            ts.TypeFlags.Null | ts.TypeFlags.Undefined,
          ),
        ),
        constants: (enm as ts.EnumDeclaration).members.map((member) =>
          c.getConstantValue(member),
        ),
        signatureArguments: c
          .getTypeArgumentsForResolvedSignature(c.getResolvedSignature(call)!)
          ?.map(format),
        parameters: c.getRootSymbols(symbol).map((s) => s.name),
        merged: c.getMergedSymbol(symbol) === symbol,
        flags: [
          Object.hasOwn(decl, 'jsDoc'),
          Object.hasOwn(src.endOfFileToken, 'jsDoc'),
          c.isArgumentsSymbol(symbol),
          c.isUndefinedSymbol(symbol),
          c.isOptionalParameter(param),
          c.isImplementationOfOverload(decl),
        ],
        builds: [
          name(generatedSignature),
          name(c.symbolToParameterDeclaration(symbol, decl, 0)),
          name(c.typeParameterToDeclaration(sig.typeParameters![0], decl, 0)),
          name(c.indexInfoToIndexSignatureDeclaration(index, decl, 0)),
        ],
        printed: [
          c.symbolToString(symbol, decl),
          c.typePredicateToString(predicate, guard),
        ],
        tuple: [
          tuple.minLength,
          tuple.fixedLength,
          tuple.hasRestElement,
          tuple.elementFlags,
          tuple.labeledElementDeclarations?.map(name),
        ],
        conditional: [
          name(cond.root.node),
          cond.root.isDistributive,
          format(cond.root.checkType),
          format(cond.root.extendsType),
        ],
        roots: p.getRootFileNames(),
        sourceByPath:
          p.getSourceFileByPath(
            (src as ts.SourceFile & { path: ts.Path }).path,
          ) === src,
        lines: [0, code.indexOf('\n'), code.length].map((pos) =>
          src.getLineEndOfPosition(pos),
        ),
        diagnostics: p
          .getSemanticDiagnostics(src)
          .map((d) => [d.code, d.start, d.length, d.file === src]),
        other: [
          p.getSyntacticDiagnostics(src).length,
          p.getOptionsDiagnostics().length,
          p.getConfigFileParsingDiagnostics().length,
          p.getProjectReferences(),
          p.getResolvedProjectReferences(),
        ],
        counts: [
          p.getNodeCount(),
          p.getIdentifierCount(),
          p.getSymbolCount(),
          p.getTypeCount(),
        ].map((n) => n > 0),
        relationKeys: Object.keys(p.getRelationCacheSizes()).sort(),
        cancel: c.runWithCancellationToken(
          {
            isCancellationRequested: () => false,
            throwIfCancellationRequested() {},
          },
          (inner) => inner === c,
        ),
      };
    };
    expect(exercise(owner.program as unknown as ts.Program)).toEqual(
      exercise(reference),
    );
    const checker = owner.program.getTypeChecker();
    const literal = checker.getStringLiteralType('abc');
    expect(Reflect.set(literal, 'value', 'changed')).toBe(false);
    expect(checker.typeToString(literal)).toBe('"abc"');
    expect(() => owner.program.emit()).toThrow('Program.emit');
    const src = owner.program.getSourceFile(file) as ts.SourceFile;
    const dictionary = checker.getTypeFromTypeNode(
      (src.statements[5] as ts.TypeAliasDeclaration).type,
    )!;
    const index = checker.getIndexInfoOfType(dictionary, ts.IndexKind.String)!;
    expect(Reflect.set(index, 'type', literal)).toBe(false);
    expect(() =>
      src.update('', { span: { start: 0, length: 0 }, newLength: 0 }),
    ).toThrow('SourceFile.update');
  } finally {
    owner.close();
    await host.close();
  }
});
test('declared members and source metadata follow TypeScript query order', async () => {
  const dir = mkdtempSync(resolve(temp, 'declared-'));
  const file = resolve(dir, 'input.ts'),
    config = resolve(dir, 'tsconfig.json');
  const names = ['input.ts', 'esm.mts', 'cjs.cts', 'empty.json'];
  writeFileSync(
    file,
    [
      '// 😀',
      '/// <reference no-default-lib="true" />',
      '/// <reference path="./esm.mts" />',
      '/// <reference path="./cjs.cts" preserve="true" />',
      '/// <reference types="a" />',
      '/// <reference types="b" resolution-mode="import" preserve="true" />',
      '/// <reference types="c" resolution-mode="require" preserve="false" />',
      '/// <reference lib="es2020" />',
      '/// <reference lib="es2015" preserve="true" />',
      'export interface Parent { inherited: string; }',
      'export interface Child extends Parent { own: number; (): string; new(): Child; [index: number]: boolean; }',
      'export interface Empty {}',
      'export class Box<T> { constructor(public value: T) {} }',
      'export type Instance = Box<string>;',
      `export type Pair = [${Array.from({ length: 12 }, (_, i) => 'slot' + i + ': number').join(',')}];`,
    ].join('\n'),
  );
  writeFileSync(
    resolve(dir, 'esm.mts'),
    "import './cjs.cjs'; declare global {} declare module './cjs.cjs' {}",
  );
  writeFileSync(resolve(dir, 'cjs.cts'), 'export {};');
  writeFileSync(resolve(dir, 'empty.json'), '{}');
  writeFileSync(resolve(dir, 'package.json'), '{"type":"module"}');
  const fields = [
    'declaredProperties',
    'declaredCallSignatures',
    'declaredConstructSignatures',
    'declaredIndexInfos',
  ] as const;
  const read = (program: ts.Program) => {
    const checker = program.getTypeChecker();
    const source = program.getSourceFile(file)!;
    const declared = (index: number) => {
      const declaration = source.statements[index];
      if (!ts.isDeclarationStatement(declaration) || !declaration.name)
        throw new Error('expected a named declaration');
      return checker.getDeclaredTypeOfSymbol(
        checker.getSymbolAtLocation(declaration.name)!,
      );
    };
    const inspect = (
      type: ts.Type &
        Partial<
          Pick<ts.InterfaceTypeWithDeclaredMembers, (typeof fields)[number]>
        >,
      includeAll = false,
    ) => {
      const presence = () => fields.map((key) => key in type);
      const before = presence();
      const props = checker.getPropertiesOfType(type);
      const calls = checker.getSignaturesOfType(type, ts.SignatureKind.Call);
      const constructs = checker.getSignaturesOfType(
        type,
        ts.SignatureKind.Construct,
      );
      const indices = checker.getIndexInfosOfType(type);
      return {
        before,
        after: presence(),
        all: includeAll ? props.map((s) => s.name) : undefined,
        own: type.declaredProperties?.map((s) => s.name),
        calls: type.declaredCallSignatures?.map((s) => [
          calls.includes(s),
          checker.signatureToString(s),
        ]),
        constructs: type.declaredConstructSignatures?.map((s) => [
          constructs.includes(s),
          checker.signatureToString(s),
        ]),
        indices: type.declaredIndexInfos?.map((info) => [
          indices.includes(info),
          checker.typeToString(info.keyType),
          checker.typeToString(info.type),
          info.declaration?.getSourceFile() === source,
        ]),
        identity: fields.map((key) => type[key] === type[key]),
        memberIdentity: type.declaredProperties?.every((prop) =>
          props.includes(prop),
        ),
      };
    };
    return {
      child: inspect(declared(1), true),
      empty: inspect(declared(2)),
      generic: inspect(declared(3)),
      instantiated: inspect(declared(4)),
      tuple: inspect((declared(5) as ts.TypeReference).target),
      modes: [0, 1].map((i) =>
        program.getModeForResolutionAtIndex(
          program.getSourceFile(resolve(dir, 'esm.mts'))!,
          i,
        ),
      ),
      sources: names.map((name) => {
        const src = program.getSourceFile(resolve(dir, name))!;
        return [
          src.languageVersion,
          src.hasNoDefaultLib,
          src.impliedNodeFormat,
          src.referencedFiles,
          src.typeReferenceDirectives,
          src.libReferenceDirectives,
          src.referencedFiles === src.referencedFiles,
        ];
      }),
    };
  };
  for (const compilerOptions of [
    { target: 'es2020', module: 'esnext', moduleResolution: 'bundler' },
    { module: 'nodenext' },
  ]) {
    const raw = {
      compilerOptions: { ...compilerOptions, resolveJsonModule: true },
      files: names,
    };
    writeFileSync(config, JSON.stringify(raw));
    const parsed = ts.parseJsonConfigFileContent(raw, ts.sys, dir);
    const expected = read(ts.createProgram(parsed.fileNames, parsed.options));
    expect(expected.tuple.own).toEqual([
      ...Array.from({ length: 12 }, (_, i) => String(i)),
      'length',
    ]);
    expect(expected.child.own).toEqual(['own']);
    expect(expected.child.all).toEqual(['own', 'inherited']);
    const host = new Host();
    const descriptor = host.request({ method: 'init', config });
    const client = new TypeServiceClient(
      enums,
      descriptor,
      (query) =>
        Buffer.from(host.request({ method: 'query', query }), 'base64'),
      host.source,
      { maxFactsBytes: 1 << 18 },
    );
    try {
      expect(read(client.program as unknown as ts.Program)).toStrictEqual(
        expected,
      );
    } finally {
      client.close();
      await host.close();
    }
  }
});

test('resolved project references retain config syntax, nesting and missing entries', async () => {
  const dir = mkdtempSync(resolve(temp, 'references-')),
    file = resolve(dir, 'input.ts'),
    config = resolve(dir, 'tsconfig.json');
  const child = resolve(dir, 'child.json'),
    nested = resolve(dir, 'nested.json');
  writeFileSync(file, 'import "./tsconfig.json"; export const n = 1;');
  writeFileSync(resolve(dir, 'nested.ts'), 'export const nested = 1;');
  writeFileSync(resolve(dir, 'child.ts'), 'export const child = 2;');
  const compilerOptions = {
    composite: true,
    resolveJsonModule: true,
    target: 'esnext',
    plugins: [{ name: 'language-plugin', settings: { mode: 'local' } }],
  };
  writeFileSync(
    resolve(dir, 'base.json'),
    '{"extends":"./nested.json","compilerOptions":{"unknownOption":true}}',
  );
  writeFileSync(
    nested,
    JSON.stringify({ compilerOptions, files: ['nested.ts'] }),
  );
  writeFileSync(
    child,
    JSON.stringify({
      compilerOptions,
      extends: './base.json',
      files: ['child.ts'],
      include: ['shallow/*.ts', 'deep/**/*.ts'],
      references: [{ path: './nested.json' }],
    }),
  );
  writeFileSync(
    config,
    JSON.stringify({
      extends: './base.json',
      compilerOptions: {
        ...compilerOptions,
        exactOptionalPropertyTypes: true,
        strictNullChecks: false,
      },
      files: ['input.ts'],
      references: [
        { path: './child.json' },
        { path: './nested.json' },
        { path: './missing.json' },
        { path: './child.json', circular: false },
      ],
    }),
  );
  const h = new Host(),
    descriptor = h.request({ method: 'init', config });
  const owner = new TypeServiceClient(
    enums,
    descriptor,
    (q) => Buffer.from(h.request({ method: 'query', query: q }), 'base64'),
    h.source,
    { maxFactsBytes: 1 << 16, maxSyntaxBytes: 0 },
  );
  try {
    const parsed = ts.getParsedCommandLineOfConfigFile(
      config,
      {},
      { ...ts.sys, onUnRecoverableConfigFileDiagnostic() {} },
    )!;
    const tsProgram = ts.createProgram({
      ...parsed,
      rootNames: parsed.fileNames,
      configFileParsingDiagnostics: parsed.errors,
    });
    const inspect = (p: ts.Program) => {
      const source = p.getOptionsDiagnostics().find((d) => d.code === 5052)!
        .file as ts.TsConfigSourceFile;
      const diagnosticSource = [
        Reflect.get(source, 'path'),
        source.extendedSourceFiles,
      ];
      const refs = p.getResolvedProjectReferences()!;
      const childBase = refs[0]!.commandLine.errors.find(
        (d) => d.code === 5023,
      )!.file as ts.TsConfigSourceFile;
      const options = p.getCompilerOptions();
      const configFile = options.configFile as ts.SourceFile;
      return {
        diagnosticSource,
        options: [
          options.plugins,
          configFile.fileName,
          configFile.text,
          configFile === p.getSourceFile(config),
          configFile.statements[0].getText(configFile),
          configFile.statements[0] === p.getSourceFile(config)!.statements[0],
          Object.keys(options).includes('configFile'),
        ],
        direct: p.getProjectReferences(),
        configs: refs.map(
          (ref) =>
            ref && [
              ref.sourceFile.fileName,
              ref.sourceFile.text,
              ref.sourceFile.flags,
              ref.sourceFile.statements[0].kind,
              ref.sourceFile.statements ===
                ref.sourceFile.forEachChild(
                  () => undefined,
                  (nodes) => nodes,
                ),
              ref.sourceFile.getSourceFile() === ref.sourceFile,
              ref.commandLine.fileNames,
              ref.commandLine.options.target,
              ref.commandLine.options.plugins,
              ref.commandLine.options.configFile === ref.sourceFile,
              ref.commandLine.errors
                .filter((d) => d.code === 5023)
                .map((d) => [
                  d.file?.fileName,
                  d.file?.text.slice(d.start!, d.start! + d.length!),
                ]),
              ref.commandLine.raw.compilerOptions.target,
              ref.commandLine.wildcardDirectories,
              (ref.sourceFile as ts.TsConfigSourceFile).extendedSourceFiles,
            ],
        ),
        nested: refs[0]!.references![0] === refs[1],
        sameReferences: refs === p.getResolvedProjectReferences(),
        distinctBaseFiles:
          childBase !==
          p.getConfigFileParsingDiagnostics().find((d) => d.code === 5023)!
            .file,
        baseAncestors: childBase.extendedSourceFiles,
        excluded: !p.getSourceFiles().some((f) => f.fileName === child),
      };
    };
    expect(inspect(owner.program as unknown as ts.Program)).toEqual(
      inspect(tsProgram),
    );
  } finally {
    owner.close();
    await h.close();
  }
});
test('a full shared cache preserves earlier results', () => {
  const owner = new SharedTypeCacheOwner(1, 1024, 2);
  const reader = new SharedTypeCacheReader(owner.descriptor);
  try {
    expect(owner.publish('\uD800', Uint8Array.of(1))).toBe(true);
    expect(owner.publish('\uD801', Uint8Array.of(2))).toBe(true);
    expect(owner.publish('overflow', Uint8Array.of(3))).toBe(false);
    expect(reader.get('\uD800')).toEqual(Uint8Array.of(1));
    expect(reader.get('\uD801')).toEqual(Uint8Array.of(2));
    expect(reader.get('overflow')).toBeUndefined();
  } finally {
    reader.close();
    owner.close();
  }
});

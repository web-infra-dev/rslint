import { beforeAll, afterAll, describe, expect, test } from 'rstack/test';
import { execFileSync } from 'node:child_process';
import { mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { resolve, relative, sep } from 'node:path';
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
  extension?: 'js' | 'tsx';
}
// Focused valid/invalid examples follow the original rule fixtures and exercise
// their actual npm implementations, including real type-utils dependencies.
const cases: Case[] = [
  {
    name: 'JavaScript scope tracks self-references without JSX',
    plugin: typedPlugin,
    rule: 'no-use-before-define',
    extension: 'js',
    code: 'export const item = item;',
    reports: 1,
  },
  {
    name: 'local declarations shadow project library types and values',
    plugin: typedPlugin,
    rule: 'no-shadow',
    options: [{ builtinGlobals: true, ignoreTypeValueShadow: false }],
    code: 'export function f() { type RequestInit = string; const URL = 1; return URL as unknown as RequestInit; }',
    reports: 2,
  },
  {
    name: 'JSX factories count as uses of their imports',
    plugin: typedPlugin,
    rule: 'no-unused-vars',
    extension: 'tsx',
    dependencies: { 'factory.ts': 'export const h = () => null, F = 0;' },
    code: 'import { h, F } from "./factory"; export const view = <><div /></>;',
    reports: 0,
  },
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
    extension: 'js',
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
    name: 'method declarations are readonly when requested',
    plugin: typedPlugin,
    rule: 'prefer-readonly-parameter-types',
    options: [{ treatMethodsAsReadonly: true }],
    code: `
      interface Methods { run(): void; optional?(): void; }
      class Runner { run() {} }
      const object = { run() {} };
      function read(a: Methods, b: Runner, c: typeof object, d: { run: () => void }) {}
    `,
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
    extension: 'js',
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
      const extension = c.extension ?? 'ts';
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
              extension === 'js' ||
              Object.keys(c.dependencies ?? {}).some((f) => f.endsWith('.js')),
            checkJs: extension === 'js',
            ...(extension === 'tsx' && {
              jsx: 'react',
              jsxFactory: 'h',
              jsxFragmentFactory: 'F',
            }),
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
      { maxFactsBytes: 1 << 20, maxSyntaxBytes: 0 },
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
      return { file, frame, handle, services };
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
    // Multiline, quote-style, and non-ASCII truncation formatting use tsgo's
    // native printer; verify semantic options and UTF-8 boundary safety rather
    // than TypeScript 6's ASCII-escape and single-line newline quirks.
    const multilineDecl = beforeSource.statements.find(
      (node): node is ts.FunctionDeclaration =>
        ts.isFunctionDeclaration(node) && node.name?.text === 'multilineCall',
    )!;
    const multilineSig = checker.getSignatureFromDeclaration(multilineDecl)!;
    const multilineObj = checker.getTypeAtLocation(
      services.esTreeNodeToTSNodeMap.get(variable('multiline').id),
    );
    const longType = beforeSource.statements.find(
      (node): node is ts.InterfaceDeclaration =>
        ts.isInterfaceDeclaration(node) && node.name.text === longTypeName,
    )!;
    const longFormatted = checker.typeToString(
      checker.getTypeAtLocation(longType.name),
    );
    expect(longFormatted).not.toContain('\uFFFD');
    expect(longFormatted.startsWith('例'.repeat(20))).toBe(true);

    const multilineFlags =
      ts.TypeFormatFlags.MultilineObjectLiterals |
      ts.TypeFormatFlags.NoTruncation;
    const multilineObjText = checker.typeToString(
      multilineObj,
      multilineDecl,
      multilineFlags,
    );
    expect(multilineObjText).toContain('\n');
    expect(multilineObjText).toContain('nested:');
    expect(multilineObjText).toContain('label: "中文"');
    expect(multilineObjText).toContain('count: number');

    const arrowCall = checker.signatureToString(
      multilineSig,
      multilineDecl,
      multilineFlags | ts.TypeFormatFlags.WriteArrowStyleSignature,
      ts.SignatureKind.Call,
    );
    const arrowConstruct = checker.signatureToString(
      multilineSig,
      multilineDecl,
      multilineFlags | ts.TypeFormatFlags.WriteArrowStyleSignature,
      ts.SignatureKind.Construct,
    );
    expect(arrowCall).toContain('=>');
    expect(arrowConstruct).toBe(`new ${arrowCall}`);

    const singleQuoteSig = checker.signatureToString(
      multilineSig,
      multilineDecl,
      multilineFlags | ts.TypeFormatFlags.UseSingleQuotesForStringLiteralType,
      ts.SignatureKind.Call,
    );
    expect(singleQuoteSig).toContain("label: '中文'");
    expect(singleQuoteSig).toContain("escaped: 'é'");
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
    // Regression: calls without a checker must not poison subsequent calls with a checker,
    // and cached arrays must be frozen against cross-rule mutation.
    const docsChild = beforeSource.statements.find(
      (node): node is ts.ClassDeclaration =>
        ts.isClassDeclaration(node) && node.name?.text === 'DocsChild',
    )!;
    const childMethod = docsChild.members[0] as ts.MethodDeclaration;
    const childSym = checker.getSymbolAtLocation(childMethod.name)!;
    const noCheckerDocs = childSym.getDocumentationComment(undefined);
    const withCheckerDocs = childSym.getDocumentationComment(checker);
    const noCheckerTags = childSym.getJsDocTags(undefined);
    const withCheckerTags = childSym.getJsDocTags(checker);
    expect(Object.isFrozen(noCheckerDocs)).toBe(true);
    expect(Object.isFrozen(withCheckerDocs)).toBe(true);
    expect(Object.isFrozen(noCheckerTags)).toBe(true);
    expect(Object.isFrozen(withCheckerTags)).toBe(true);
    expect(noCheckerDocs.map((p: any) => p.kind)).not.toContain('linkName');
    expect(withCheckerDocs.map((p: any) => p.kind)).toContain('linkName');
    expect(noCheckerTags.map((t: any) => t.name)).not.toContain('deprecated');
    expect(withCheckerTags.map((t: any) => t.name)).toContain('deprecated');
  });
});

test('public checker, Program, mapped types and declared index infos match TypeScript', async () => {
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
      'export type Picked<T> = { [K in keyof T]: T[K] };',
      'export type UsedPicked = Picked<{ p: string }>;',
      'export interface Parent { inherited: string; }',
      'export interface Child extends Parent { own: number; [index: number]: boolean; }',
      'export class Box { read() { return 1; } }',
      'export { id as alias };',
      'export const record = { n: 1 }; export const {n} = record;',
      "export * from './external';",
    ].join('\n') + '\n';
  writeFileSync(file, code);
  writeFileSync(resolve(dir, 'external.ts'), 'export type External = string;');
  writeFileSync(
    config,
    JSON.stringify({
      compilerOptions: {
        strict: true,
        target: 'esnext',
        module: 'esnext',
        lib: ['es2022'],
        types: [],
      },
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
    const statement = (source: ts.SourceFile, name: string) => {
      const found = source.statements.find((node) => {
        const declaration = ts.isVariableStatement(node)
          ? node.declarationList.declarations[0]
          : (node as ts.DeclarationStatement);
        const candidate = ts.getNameOfDeclaration(declaration);
        return (
          candidate && ts.isIdentifier(candidate) && candidate.text === name
        );
      });
      expect(found, name).toBeDefined();
      return found!;
    };
    const exercise = (p: ts.Program) => {
      const c = p.getTypeChecker(),
        src = p.getSourceFile(file)!;
      const enm = statement(src, 'Values');
      const pair = statement(src, 'Pair');
      const fn = statement(src, 'id') as ts.FunctionDeclaration;
      const dict = statement(src, 'Dictionary');
      const select = statement(src, 'Select');
      const picked = statement(src, 'Picked') as ts.TypeAliasDeclaration;
      const usedPicked = statement(
        src,
        'UsedPicked',
      ) as ts.TypeAliasDeclaration;
      const childIface = statement(src, 'Child') as ts.InterfaceDeclaration;
      const param = fn.parameters[0];
      const symbol = c.getSymbolAtLocation(param.name)!;
      const fileSymbol = c.getSymbolAtLocation(src)!;
      const t = c.getTypeFromTypeNode(
        (pair as ts.TypeAliasDeclaration).type,
      ) as ts.TupleTypeReference;
      const tuple = t.target;
      const cond = c.getTypeFromTypeNode(
        (select as ts.TypeAliasDeclaration).type,
      ) as ts.ConditionalType;
      const mapped = c.getTypeFromTypeNode(picked.type) as any;
      const usedMapped = c.getTypeFromTypeNode(usedPicked.type);
      const childType = c.getDeclaredTypeOfSymbol(
        c.getSymbolAtLocation(childIface.name)!,
      ) as ts.InterfaceTypeWithDeclaredMembers;
      const exportSpec = (
        (
          src.statements.find(
            (node) =>
              ts.isExportDeclaration(node) &&
              node.exportClause &&
              ts.isNamedExports(node.exportClause) &&
              node.exportClause.elements.some(
                (element) => element.name.text === 'alias',
              ),
          ) as ts.ExportDeclaration
        ).exportClause as ts.NamedExports
      ).elements[0];
      const format = (value: ts.Type) => c.typeToString(value);
      const name = (node: ts.Node | undefined) =>
        node && ts.SyntaxKind[node.kind];

      const beforeTemplate = mapped.templateType;
      (c as any).getTypeOfPropertyOfType(usedMapped, 'p');
      const afterTemplate = mapped.templateType;

      const beforeDeclaredIndex = childType.declaredIndexInfos;
      const allIndexInfos = c.getIndexInfosOfType(childType);
      const afterDeclaredIndex = childType.declaredIndexInfos;

      return {
        primitive: [
          c.getAnyType(),
          c.getUnknownType(),
          c.getStringType(),
          c.getNumberType(),
          c.getBigIntType(),
          c.getBooleanType(),
          c.getESSymbolType(),
          c.getNonPrimitiveType(),
          c.getNeverType(),
          c.getNullType(),
          c.getUndefinedType(),
          c.getVoidType(),
        ].map(format),
        interned: c.getStringType() === c.getStringType(),
        constants: (enm as ts.EnumDeclaration).members.map((member) =>
          c.getConstantValue(member),
        ),
        rootSymbols: c.getRootSymbols(symbol).map((s) => s.name),
        merged: c.getMergedSymbol(symbol) === symbol,
        symbolPredicates: [
          c.isArgumentsSymbol(symbol),
          c.isUndefinedSymbol(symbol),
        ],
        localExport: c.getExportSpecifierLocalTargetSymbol(exportSpec)?.name,
        moduleMember: c.tryGetMemberInModuleExports('alias', fileSymbol)?.name,
        rawExports: [...(fileSymbol.exports?.keys() ?? [])].sort(),
        exports: c
          .getExportsOfModule(fileSymbol)
          .map((s) => s.name)
          .sort(),
        symbolString: c.symbolToString(symbol, fn),
        shape: [
          'types' in c.getStringType(),
          'target' in c.getStringType(),
          'thisType' in t,
          tuple.thisType && format(tuple.thisType),
          'types' in c.getBooleanType(),
          'root' in cond,
          'root' in c.getStringType(),
          'elementFlags' in tuple,
          'elementFlags' in t,
          'typeParameter' in mapped,
          'typeParameter' in c.getStringType(),
        ],
        mapped: [
          mapped.typeParameter?.symbol?.name,
          format(mapped.constraintType),
          beforeTemplate === undefined,
          afterTemplate?.indexType === mapped.typeParameter,
        ],
        declaredIndex: [
          beforeDeclaredIndex === undefined,
          afterDeclaredIndex?.map((info) => [
            allIndexInfos.includes(info),
            format(info.keyType),
            format(info.type),
          ]),
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
        diagnostics: {
          semantic: p
            .getSemanticDiagnostics(src)
            .map((d) => [d.code, d.start, d.length, d.file === src]),
          syntactic: p.getSyntacticDiagnostics(src).length,
        },
        dictIndex: format(
          c.getIndexInfoOfType(
            c.getTypeFromTypeNode((dict as ts.TypeAliasDeclaration).type),
            ts.IndexKind.String,
          )!.type,
        ),
      };
    };
    const actual = owner.program as unknown as ts.Program;
    expect(exercise(actual)).toEqual(exercise(reference));

    const checker = owner.program.getTypeChecker();
    const src = owner.program.getSourceFile(file) as ts.SourceFile;
    const stringType = checker.getStringType();
    expect(Reflect.set(stringType, 'flags', 0)).toBe(false);
    expect(() => owner.program.emit()).toThrow('Program.emit');
    expect(() =>
      src.update('', { span: { start: 0, length: 0 }, newLength: 0 }),
    ).toThrow('SourceFile.update');
    const dictionary = checker.getTypeFromTypeNode(
      (statement(src, 'Dictionary') as ts.TypeAliasDeclaration).type,
    )!;
    const index = checker.getIndexInfoOfType(dictionary, ts.IndexKind.String)!;
    expect(Reflect.set(index, 'type', stringType)).toBe(false);
  } finally {
    owner.close();
    await host.close();
  }
});

test('decorator metadata flags: typed Program is authoritative, syntax-only honors overrides and preserves custom factories', async () => {
  const dir = mkdtempSync(resolve(temp, 'decorator-precedence-'));
  const serviceFile = resolve(dir, 'service.ts');
  const consumerFile = resolve(dir, 'consumer.ts');
  const withDecoratorsConfig = resolve(dir, 'tsconfig.decorators.json');
  const plainConfig = resolve(dir, 'tsconfig.plain.json');
  const code = `import { Service } from './service';
function Injectable(): ClassDecorator {
  return () => {};
}
@Injectable()
export class Consumer {
  constructor(private readonly service: Service) {}
}
`;
  writeFileSync(serviceFile, 'export class Service {}\n');
  writeFileSync(consumerFile, code);
  writeFileSync(
    withDecoratorsConfig,
    JSON.stringify({
      compilerOptions: {
        strict: true,
        target: 'ES2022',
        module: 'ESNext',
        experimentalDecorators: true,
        emitDecoratorMetadata: true,
      },
      include: ['service.ts', 'consumer.ts'],
    }),
  );
  writeFileSync(
    plainConfig,
    JSON.stringify({
      compilerOptions: {
        strict: true,
        target: 'ES2022',
        module: 'ESNext',
      },
      include: ['service.ts', 'consumer.ts'],
    }),
  );

  // Confirm upstream ESLint preserves the import even when conflicting flags
  // are supplied alongside a project tsconfig that enables decorator metadata.
  const eslint = new Linter({ cwd: dir });
  const upstreamEnabled = eslint.verify(
    code,
    [
      {
        files: ['**/*.ts'],
        languageOptions: {
          parser: tsParser,
          parserOptions: {
            project: withDecoratorsConfig,
            tsconfigRootDir: dir,
            emitDecoratorMetadata: false,
            experimentalDecorators: false,
          },
        },
        plugins: { '@typescript-eslint': typedPlugin as any },
        rules: { '@typescript-eslint/consistent-type-imports': 'error' },
      },
    ],
    { filename: consumerFile },
  );
  expect(upstreamEnabled).toEqual([]);

  const runMode = async (
    config: string,
    syntaxOnly: boolean,
    opts?: {
      requestParserOptions?: {
        experimentalDecorators?: boolean;
        emitDecoratorMetadata?: boolean;
      };
      fileParserOptions?: {
        experimentalDecorators?: boolean;
        emitDecoratorMetadata?: boolean;
      };
      customGetterFactory?: boolean;
    },
  ) => {
    const host = new Host();
    const descriptor = host.request({
      method: 'init',
      config,
      syntaxOnly,
    }) as TypeServiceDescriptor;
    const client = new TypeServiceClient(
      enums,
      descriptor,
      (query) =>
        Buffer.from(host.request({ method: 'query', query }), 'base64'),
      host.source,
      { maxFactsBytes: 1 << 20 },
    );
    try {
      const encoded = host.request({ method: 'project', file: consumerFile });
      const frame = Buffer.from(encoded.frame, 'base64');
      const projection = {
        ...encoded,
        tree: Buffer.from(encoded.tree, 'base64'),
      };
      const handle = client.file(projection.binding);
      const rules = new Map([
        [
          '@typescript-eslint/consistent-type-imports',
          typedPlugin.rules['consistent-type-imports'],
        ],
      ]);

      let seenService: Record<string, unknown> | undefined;
      let originalInstance: Record<string, unknown> | undefined;
      const probeRule = {
        meta: { type: 'problem', docs: {}, schema: [], messages: {} },
        create(context: any) {
          seenService = context.sourceCode.parserServices;
          return {};
        },
      };
      const allRules = opts?.customGetterFactory
        ? new Map([...rules, ['probe/check', probeRule as any]])
        : rules;
      const ruleConfigs: Record<string, { options: unknown[] }> = {
        '@typescript-eslint/consistent-type-imports': { options: [] },
        ...(opts?.customGetterFactory
          ? { 'probe/check': { options: [] } }
          : {}),
      };

      const result = lintFile(
        {
          filePath: consumerFile,
          frame,
          languageOptions: opts?.requestParserOptions
            ? { parserOptions: opts.requestParserOptions }
            : undefined,
          rules: ruleConfigs,
          collectFixes: true,
          suggestionsMode: 'eager',
        },
        { rules: allRules } as any,
        (decoded, parserOptions) => {
          const base = handle.parserServices(
            decoded,
            projection,
            projection.binding.sourceFileName,
            opts?.fileParserOptions ?? parserOptions,
          );
          if (!opts?.customGetterFactory) return base;
          // Custom class factory exposing node maps via prototype getters and
          // a prototype method reading a private field.
          class CustomSyntaxServices {
            readonly #tag = 'kept';
            readonly program = null;
            readonly emitDecoratorMetadata = false;
            readonly experimentalDecorators = false;
            readonly isolatedDeclarations = false;
            get esTreeNodeToTSNodeMap() {
              return base.esTreeNodeToTSNodeMap;
            }
            get tsNodeToESTreeNodeMap() {
              return base.tsNodeToESTreeNodeMap;
            }
            readTag() {
              return this.#tag;
            }
          }
          const instance = new CustomSyntaxServices();
          originalInstance = instance as unknown as Record<string, unknown>;
          return instance;
        },
      );
      expect(result.parseError).toBeUndefined();
      expect(result.ruleErrors).toBeUndefined();
      return { result, seenService, originalInstance };
    } finally {
      client.close();
      await host.close();
    }
  };

  // Typed project: tsconfig is authoritative; conflicting per-file flags (either
  // passed via handle.parserServices or via lintFile request) MUST NOT override it.
  const typedFileOverride = await runMode(withDecoratorsConfig, false, {
    fileParserOptions: {
      experimentalDecorators: false,
      emitDecoratorMetadata: false,
    },
  });
  expect(typedFileOverride.result.diagnostics).toEqual([]);

  const typedRequestOverride = await runMode(withDecoratorsConfig, false, {
    requestParserOptions: {
      experimentalDecorators: false,
      emitDecoratorMetadata: false,
    },
  });
  expect(typedRequestOverride.result.diagnostics).toEqual([]);

  // Syntax-only: tsconfig propagates decorator flags.
  const syntaxFromTsconfig = await runMode(withDecoratorsConfig, true);
  expect(syntaxFromTsconfig.result.diagnostics).toEqual([]);

  // Syntax-only: per-file parserOptions override works in both directions.
  const syntaxEnabledByOverride = await runMode(plainConfig, true, {
    requestParserOptions: {
      experimentalDecorators: true,
      emitDecoratorMetadata: true,
    },
  });
  expect(syntaxEnabledByOverride.result.diagnostics).toEqual([]);

  const syntaxDisabledByDefault = await runMode(plainConfig, true);
  expect(syntaxDisabledByDefault.result.diagnostics).toHaveLength(1);

  // Syntax-only custom class factory with prototype getters + private field method:
  // per-file override must preserve getters/methods and never mutate the original.
  const customWithOverride = await runMode(plainConfig, true, {
    requestParserOptions: {
      experimentalDecorators: true,
      emitDecoratorMetadata: true,
    },
    customGetterFactory: true,
  });
  expect(customWithOverride.result.diagnostics).toEqual([]);
  expect(customWithOverride.seenService?.emitDecoratorMetadata).toBe(true);
  expect(customWithOverride.originalInstance?.emitDecoratorMetadata).toBe(
    false,
  );
  expect((customWithOverride.seenService as any)?.readTag()).toBe('kept');
});

test('ts.getPreEmitDiagnostics and config diagnostics preserve tsconfig.json location', async () => {
  const dir = mkdtempSync(resolve(temp, 'pre-emit-config-'));
  const file = resolve(dir, 'input.ts');
  const config = resolve(dir, 'tsconfig.json');
  writeFileSync(file, `export const bad: number = 'oops';\n`);
  writeFileSync(
    config,
    JSON.stringify(
      {
        compilerOptions: {
          strict: true,
          target: 'ES2022',
          module: 'ESNext',
          skipLibCheck: true,
          inlineSources: true,
        },
        include: ['input.ts'],
      },
      null,
      2,
    ) + '\n',
  );

  const configSource = ts.readJsonConfigFile(config, ts.sys.readFile);
  const cmd = ts.parseJsonSourceFileConfigFileContent(
    configSource,
    ts.sys,
    dir,
    undefined,
    config,
  );
  const reference = ts.createProgram({
    rootNames: cmd.fileNames,
    options: cmd.options,
    configFileParsingDiagnostics: cmd.errors,
  });

  const host = new Host();
  const descriptor = host.request({
    method: 'init',
    config,
  }) as TypeServiceDescriptor;
  const owner = new TypeServiceClient(
    enums,
    descriptor,
    (query) => Buffer.from(host.request({ method: 'query', query }), 'base64'),
    host.source,
    { maxFactsBytes: 1 << 20 },
  );

  try {
    const actualProgram = owner.program as unknown as ts.Program;
    // tsconfig.json must remain excluded from Program source-file APIs.
    expect(actualProgram.getSourceFile(config)).toBeUndefined();
    expect(
      actualProgram.getSourceFiles().some((s) => s.fileName === config),
    ).toBe(false);

    const formatHost: ts.FormatDiagnosticsHost = {
      getCurrentDirectory: () => dir,
      getCanonicalFileName: (name) => name,
      getNewLine: () => '\n',
    };
    const summarize = (diags: readonly ts.Diagnostic[]) =>
      diags
        .map((d) => [d.code, d.start, d.length, d.file?.fileName])
        .sort((a, b) => Number(a[0]) - Number(b[0]));

    expect(summarize(actualProgram.getOptionsDiagnostics())).toEqual(
      summarize(reference.getOptionsDiagnostics()),
    );
    expect(summarize(ts.getPreEmitDiagnostics(actualProgram))).toEqual(
      summarize(ts.getPreEmitDiagnostics(reference)),
    );

    const actualTs5051 = actualProgram
      .getOptionsDiagnostics()
      .find((d) => d.code === 5051)!;
    const refTs5051 = reference
      .getOptionsDiagnostics()
      .find((d) => d.code === 5051)!;
    expect(actualTs5051.file?.fileName).toBe(refTs5051.file?.fileName);
    expect(ts.formatDiagnostic(actualTs5051, formatHost)).toBe(
      ts.formatDiagnostic(refTs5051, formatHost),
    );
    expect(ts.formatDiagnostic(actualTs5051, formatHost)).toContain(
      'tsconfig.json(',
    );
  } finally {
    owner.close();
    await host.close();
  }
});

test('getResolvedSignature forwards candidates and argumentCount without poisoning cached signatures', async () => {
  const dir = mkdtempSync(resolve(temp, 'overload-resolution-'));
  const file = resolve(dir, 'input.ts');
  const config = resolve(dir, 'tsconfig.json');
  writeFileSync(
    file,
    `export function pick(a: string): 1;
export function pick(a: number, b: boolean): 2;
export function pick(a: string | number, b?: boolean): 1 | 2 {
  return b ? 2 : 1;
}
export const incomplete = pick(true as unknown as never);
export const call1 = pick('ok');
export const call2 = pick(1, true);
`,
  );
  writeFileSync(
    config,
    JSON.stringify({
      compilerOptions: {
        strict: true,
        target: 'ES2022',
        module: 'ESNext',
        skipLibCheck: true,
      },
      include: ['input.ts'],
    }),
  );

  const configSource = ts.readJsonConfigFile(config, ts.sys.readFile);
  const cmd = ts.parseJsonSourceFileConfigFileContent(
    configSource,
    ts.sys,
    dir,
    undefined,
    config,
  );
  const reference = ts.createProgram({
    rootNames: cmd.fileNames,
    options: cmd.options,
    configFileParsingDiagnostics: cmd.errors,
  });

  const host = new Host();
  const descriptor = host.request({
    method: 'init',
    config,
  }) as TypeServiceDescriptor;
  const owner = new TypeServiceClient(
    enums,
    descriptor,
    (query) => Buffer.from(host.request({ method: 'query', query }), 'base64'),
    host.source,
    { maxFactsBytes: 1 << 20 },
  );

  try {
    const exercise = (p: ts.Program) => {
      const c = p.getTypeChecker();
      const src = p.getSourceFile(file)!;
      const callFor = (varName: string) => {
        const stmt = src.statements.find(
          (s): s is ts.VariableStatement =>
            ts.isVariableStatement(s) &&
            ts.isIdentifier(s.declarationList.declarations[0].name) &&
            s.declarationList.declarations[0].name.text === varName,
        )!;
        return stmt.declarationList.declarations[0]
          .initializer as ts.CallExpression;
      };
      const retOf = (sig: ts.Signature | undefined) =>
        sig ? c.typeToString(c.getReturnTypeOfSignature(sig)) : undefined;

      // Incomplete call: querying with (undefined, 2) after caching MUST NOT
      // overwrite the cached overload; passing a candidates array DOES resolve
      // with argumentCount = 2 and populates both candidates.
      const incomplete = callFor('incomplete');
      const first = retOf(c.getResolvedSignature(incomplete));
      const countWithoutCandidates = retOf(
        c.getResolvedSignature(incomplete, undefined, 2),
      );
      const after = retOf(c.getResolvedSignature(incomplete));
      const candidates: ts.Signature[] = [];
      const countWithCandidates = retOf(
        c.getResolvedSignature(incomplete, candidates, 2),
      );
      const afterCandidates = retOf(c.getResolvedSignature(incomplete));

      const call1Candidates: ts.Signature[] = [];
      const resolved1 = c.getResolvedSignature(
        callFor('call1'),
        call1Candidates,
      )!;
      const call2Candidates: ts.Signature[] = [];
      const resolved2 = c.getResolvedSignature(
        callFor('call2'),
        call2Candidates,
      )!;

      return {
        first,
        countWithoutCandidates,
        after,
        countWithCandidates,
        candidatesInfo: candidates.map(retOf),
        afterCandidates,
        resolved1: c.signatureToString(resolved1),
        call1Candidates: call1Candidates.map((s) => c.signatureToString(s)),
        resolved2: c.signatureToString(resolved2),
        call2Candidates: call2Candidates.map((s) => c.signatureToString(s)),
      };
    };

    const actual = exercise(owner.program as unknown as ts.Program);
    const expected = exercise(reference);
    expect(actual).toEqual(expected);
    expect([actual.first, actual.countWithoutCandidates, actual.after]).toEqual(
      ['1', '1', '1'],
    );
    expect(actual.candidatesInfo).toEqual(['1', '2']);
  } finally {
    owner.close();
    await host.close();
  }
});

test('core checker helpers and runWithCancellationToken match TypeScript', async () => {
  const dir = mkdtempSync(resolve(temp, 'checker-helpers-'));
  const file = resolve(dir, 'input.ts');
  const config = resolve(dir, 'tsconfig.json');
  writeFileSync(
    file,
    `export function pick(a: string): string;
export function pick(a: number, b: boolean): number;
export function pick(a: string | number, b?: boolean): string | number {
  return b ? a : a;
}
export class Holder {
  constructor(public readonly value: number, opt?: string) {
    void opt;
  }
}
export function isNum(x: unknown): x is number {
  return typeof x === 'number';
}
export function assertPresent(x: unknown): asserts x {
  if (!x) throw new Error('missing');
}
`,
  );
  writeFileSync(
    config,
    JSON.stringify({
      compilerOptions: {
        strict: true,
        target: 'ES2022',
        module: 'ESNext',
        skipLibCheck: true,
      },
      include: ['input.ts'],
    }),
  );

  const configSource = ts.readJsonConfigFile(config, ts.sys.readFile);
  const cmd = ts.parseJsonSourceFileConfigFileContent(
    configSource,
    ts.sys,
    dir,
    undefined,
    config,
  );
  const reference = ts.createProgram({
    rootNames: cmd.fileNames,
    options: cmd.options,
    configFileParsingDiagnostics: cmd.errors,
  });

  const host = new Host();
  const descriptor = host.request({
    method: 'init',
    config,
  }) as TypeServiceDescriptor;
  const owner = new TypeServiceClient(
    enums,
    descriptor,
    (query) => Buffer.from(host.request({ method: 'query', query }), 'base64'),
    host.source,
    { maxFactsBytes: 1 << 20 },
  );

  try {
    const exercise = (p: ts.Program) => {
      const c = p.getTypeChecker();
      const src = p.getSourceFile(file)!;
      const pickDecls = src.statements.filter(
        (s): s is ts.FunctionDeclaration =>
          ts.isFunctionDeclaration(s) && s.name?.text === 'pick',
      );
      const holder = src.statements.find(
        (s): s is ts.ClassDeclaration =>
          ts.isClassDeclaration(s) && s.name?.text === 'Holder',
      )!;
      const ctor = holder.members.find(ts.isConstructorDeclaration)!;
      const isNum = src.statements.find(
        (s): s is ts.FunctionDeclaration =>
          ts.isFunctionDeclaration(s) && s.name?.text === 'isNum',
      )!;
      const assertPresent = src.statements.find(
        (s): s is ts.FunctionDeclaration =>
          ts.isFunctionDeclaration(s) && s.name?.text === 'assertPresent',
      )!;
      const pred = c.getTypePredicateOfSignature(
        c.getSignatureFromDeclaration(isNum)!,
      )!;
      const assertsPred = c.getTypePredicateOfSignature(
        c.getSignatureFromDeclaration(assertPresent)!,
      )!;

      // Cancellation: pre-cancelled token throws during checker operations.
      let tokenPolls = 0;
      const cancelledToken: ts.CancellationToken = {
        isCancellationRequested: () => true,
        throwIfCancellationRequested() {
          tokenPolls++;
          throw new Error('cancelled by caller');
        },
      };
      const stringType = c.getStringType();
      let cancellationOutcome = '';
      try {
        cancellationOutcome = c.runWithCancellationToken(
          cancelledToken,
          (inner) => inner.typeToString(stringType),
        );
      } catch (err) {
        cancellationOutcome = String(err);
      }
      // After runWithCancellationToken exits, subsequent queries must not poll it.
      const pollsBeforeOutside = tokenPolls;
      const afterTokenFormat = c.typeToString(stringType);
      const pollsAfterOutside = tokenPolls;

      return {
        overloads: pickDecls.map((d) => c.isImplementationOfOverload(d)),
        optionalParams: ctor.parameters.map((param) =>
          c.isOptionalParameter(param),
        ),
        paramPropSymbols: c
          .getSymbolsOfParameterPropertyDeclaration(
            ctor.parameters[0] as ts.ParameterDeclaration,
            'value',
          )
          .map((s) => [s.name, s.flags]),
        literals: [
          c.typeToString(c.getStringLiteralType('hello')),
          c.typeToString(c.getNumberLiteralType(42)),
          c.typeToString(c.getNumberLiteralType(Number.NaN)),
          c.typeToString(c.getNumberLiteralType(Number.POSITIVE_INFINITY)),
          c.typeToString(
            c.getBigIntLiteralType({ negative: true, base10Value: '99' }),
          ),
          c.typeToString(
            c.getBigIntLiteralType({ negative: true, base10Value: '0' }),
          ),
          c.typeToString(c.getTrueType()),
          c.typeToString(c.getFalseType()),
          c.typeToString(
            c.getNullableType(c.getStringType(), ts.TypeFlags.Undefined),
          ),
        ],
        predicateString: c.typePredicateToString(pred, isNum),
        assertsPredicateString: c.typePredicateToString(
          assertsPred,
          assertPresent,
        ),
        withActiveToken: c.runWithCancellationToken(
          {
            isCancellationRequested: () => false,
            throwIfCancellationRequested() {},
          },
          (inner) => inner.typeToString(inner.getTrueType()),
        ),
        cancellationOutcome,
        polledAtLeastOnce: tokenPolls >= 1,
        restoredAfterExit: pollsAfterOutside === pollsBeforeOutside,
        afterTokenFormat,
      };
    };

    expect(exercise(owner.program as unknown as ts.Program)).toEqual(
      exercise(reference),
    );
  } finally {
    owner.close();
    await host.close();
  }
});

test('zero-byte fact cache uses single-pass reads without retaining pending records', async () => {
  const dir = mkdtempSync(resolve(temp, 'zero-cache-budget-'));
  const file = resolve(dir, 'input.ts');
  const config = resolve(dir, 'tsconfig.json');
  const propLines = Array.from(
    { length: 1000 },
    (_, i) => `  p${i}: number;`,
  ).join('\n');
  writeFileSync(
    file,
    `export interface Huge {\n${propLines}\n}\nexport declare const huge: Huge;\nexport function isNum(x: unknown): x is number {\n  return typeof x === 'number';\n}\n`,
  );
  writeFileSync(
    config,
    JSON.stringify({
      compilerOptions: {
        strict: true,
        target: 'ES2022',
        module: 'ESNext',
        skipLibCheck: true,
      },
      include: ['input.ts'],
    }),
  );

  const host = new Host();
  const descriptor = host.request({
    method: 'init',
    config,
  }) as TypeServiceDescriptor;

  let queries = 0;
  let recordQueries = 0;
  const zeroCacheClient = new TypeServiceClient(
    enums,
    descriptor,
    (query) => {
      queries++;
      if (
        query.op === 'typeRecord' ||
        query.op === 'symbolRecord' ||
        query.op === 'signatureRecord'
      ) {
        recordQueries++;
      }
      return Buffer.from(host.request({ method: 'query', query }), 'base64');
    },
    host.source,
    { maxFactsBytes: 0 },
  );

  try {
    const zc = zeroCacheClient.program.getTypeChecker();

    // Cold getStringType() consumes the bundled typeRecord during eager view
    // construction: 1 Go query total, 0 record queries, 0 retained bytes.
    queries = 0;
    recordQueries = 0;
    const zString = zc.getStringType();
    expect(zString.flags).toBe(ts.TypeFlags.String);
    expect(zString.intrinsicName).toBe('string');
    expect(queries).toBe(1);
    expect(recordQueries).toBe(0);

    // Reading all shallow fields of a lazy SymbolView / SignatureView with a
    // zero-byte cache takes at most 1 single-pass record query (not 5-6).
    const zSrc = zeroCacheClient.program.getSourceFile(file) as ts.SourceFile;
    const zIsNum = zSrc.statements.find(
      (s): s is ts.FunctionDeclaration =>
        ts.isFunctionDeclaration(s) && s.name?.text === 'isNum',
    )!;
    queries = 0;
    recordQueries = 0;
    const zSym = zc.getSymbolAtLocation(zIsNum.name!)!;
    void zSym.name;
    void zSym.escapedName;
    void zSym.flags;
    void zSym.declarations;
    void zSym.valueDeclaration;
    void zSym.links;
    expect(queries).toBe(2);
    expect(recordQueries).toBe(1);

    queries = 0;
    recordQueries = 0;
    const zSig = zc.getSignatureFromDeclaration(zIsNum)!;
    void zSig.parameters;
    void zSig.declaration;
    void zSig.typeParameters;
    void zSig.thisParameter;
    void zSig.minArgumentCount;
    expect(queries).toBe(2);
    expect(recordQueries).toBe(1);

    // Listing 1,000 properties with maxFactsBytes: 0 must NOT retain unconsumed
    // supplemental records in the client after the call returns.
    const hugeDecl = zSrc.statements.find(
      (s): s is ts.VariableStatement =>
        ts.isVariableStatement(s) &&
        ts.isIdentifier(s.declarationList.declarations[0].name) &&
        s.declarationList.declarations[0].name.text === 'huge',
    )!.declarationList.declarations[0];
    const hugeType = zc.getTypeAtLocation(hugeDecl)!;
    const props = zc.getPropertiesOfType(hugeType);
    expect(props).toHaveLength(1000);
    expect((zeroCacheClient as any).facts.size).toBe(0);
    expect((zeroCacheClient as any).factBytes).toBe(0);
    expect((zeroCacheClient as any).replyRecords).toBeUndefined();
  } finally {
    zeroCacheClient.close();
    await host.close();
  }
});

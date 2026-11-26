import { afterAll, beforeAll, describe, expect, test } from 'rstack/test';
import { mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { relative, resolve } from 'node:path';
import { Linter } from 'eslint';
import tsParser from '@typescript-eslint/parser';
import typedPlugin from '@typescript-eslint/eslint-plugin';
import sonarPlugin from 'eslint-plugin-sonarjs';
import effectorPlugin from 'eslint-plugin-effector';
import ts from 'typescript';
import {
  TypeServiceClient,
  type TypeServiceDescriptor,
} from '../../../../rslint/src/languages/js/type-service.js';
import { lintFile } from '../../../../rslint/src/eslint-plugin/linter/ecma-language-plugin.js';
import { decodeFrame } from '../../../../rslint/src/languages/js/frame.js';
import type { ESTreeNode } from '../../../../rslint/src/languages/js/types.js';
import { enums, Host, root, useTypeServiceHarness } from './harness.js';

const harness = useTypeServiceHarness();

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
    host = harness.createHost();
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
          const result = lintFile(
            request,
            { rules } as any,
            handle.parserServicesFactory(projection),
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
  const host = harness.createHost();
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

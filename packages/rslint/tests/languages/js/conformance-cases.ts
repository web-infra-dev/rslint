/** Shared syntax corpus for AST, token and plugin-fix conformance. */
export interface ConformanceCase {
  name: string;
  text: string;
  filePath: string;
  sourceType: 'module' | 'script';
}

export const CONFORMANCE_CASES: ConformanceCase[] = [];
function add(
  group: string,
  texts: readonly string[],
  extension = 'ts',
  sourceType: 'module' | 'script' = 'module',
) {
  for (const [index, text] of texts.entries()) {
    CONFORMANCE_CASES.push({
      name: `${group}/${index + 1}`,
      text,
      filePath: `case.${extension}`,
      sourceType,
    });
  }
}

add('expressions', [
  'a + b * c',
  '(a + b) * (c - d)',
  '(a, b, c)',
  'a ? (b, c) : d',
  'a && b || c',
  'a ?? (b || c)',
  'a ** b ** c',
  'a = b = c',
  'a ||= b; a &&= b; a ??= b;',
  'a <<= b; a >>= b; a >>>= b;',
  'a?.b',
  'a?.[b]',
  'a?.(b)',
  'a.b?.(c)',
  'a?.b(c)',
  '(a?.b).c',
  '(a?.b)(c)',
  '(a?.b)?.(c)',
  'a?.b!.c',
  'a?.b!()',
  'a!.b?.c',
  '(a!)?.b',
  '(a as T)?.b',
  '(a satisfies T)?.b',
  'new C',
  'new C()',
  'new (C)(x)',
  'new (factory<T>)(x)',
  'new C<A<B>>()',
  'import("mod")',
  'import("mod", {with: {type: "json"}})',
  'import.meta',
  'function f() { return new.target }',
  'void a; typeof b; delete c.d;',
  '++a; a++; --b; b--; !a; ~b; -c; +d;',
  'const a = { [key]: value, method() {}, get x() {}, set x(value) {} };',
  'const a = [,,value,...rest,];',
  'const a = (value as T)!;',
  'const a = <T>value;',
  'const a = <A<B>>value;',
  'const a = fn<T>;',
  'const a = fn<T>(value);',
  'const a = tag<T>`value${x}`;',
]);

const patterns = [
  '{key}',
  '{key = 1}',
  '{source: local = 1}',
  '{[key]: local = 1}',
  '{source: {nested} = fallback}',
  '{source: [nested = 1] = fallback}',
  '{key, ...rest}',
  '[value = 1]',
  '[,{source: local = 1},...rest]',
  '[first = (fallback), [nested = 1] = values]',
];
for (const pattern of patterns) {
  add(`binding/${pattern}`, [
    `const ${pattern} = input;`,
    `function f(${pattern}: T = fallback) {}`,
    `const f = (${pattern}: T = fallback) => value;`,
    `for (const ${pattern} of inputs) {}`,
    `(${pattern} = input);`,
  ]);
}

add('functions', [
  'function f() {}',
  'async function f() { await value; }',
  'function* f() { yield; yield value; yield* values; }',
  'async function* f() { yield await value; }',
  'const f = function named<T>(x: T): T { return x; };',
  'const f = <T,>(x: T): T => x;',
  'const f = async <T,>(x: T) => x;',
  'const f = (x) => ({x});',
  'const f = x => x;',
  'function f(...rest: T[]) {}',
  'function f(x?: T, y: T = value) {}',
  'function f(this: C, x: T) {}',
  'declare function f<T>(x: T): T;',
  'function f(x: T): T; function f(x: T) { return x; }',
  'function f(x: unknown): x is T { return true; }',
  'function f(x: unknown): asserts x is T {}',
  'const object = { async /* gap */ method /* gap */ <T> /* gap */ (x: T): T {}, *generator() {} };',
]);

const members = [
  'field;',
  'field = 1;',
  'field?: T;',
  'field!: T;',
  'readonly field: T;',
  'declare field: T;',
  'static field = 1;',
  'accessor field = 1;',
  'static accessor field: T = value;',
  '[key]: T = value;',
  '#field = 1;',
  'method() {}',
  'method /* gap */ <T> /* gap */ (x: T): T {}',
  'async method() {}',
  '*method() {}',
  'get field(): T {}',
  'set field(x: T) {}',
  'constructor() {}',
  'private /* name */ constructor() {}',
  'constructor(public readonly field: T = value) {}',
  'constructor(@dec public field: T) {}',
  'static constructor() {}',
  '"constructor"() {}',
  'static "constructor"() {}',
  'abstract method(x: T): T;',
  'abstract get field(): T;',
  'abstract set field(x: T);',
  'abstract accessor field: T;',
  'static { init(); }',
];
add(
  'class members',
  members.map((member) => `abstract class C { ${member} }`),
);
add('classes', [
  'class C extends A.B<T> implements I<T>, A.B<T> {}',
  'const C = class Named<T> extends Base<T> {};',
  '@dec class C { @dec field = 1; @dec method(@dec x: T) {} }',
  'export default class {}',
  'export default function() {}',
  '@dec export class C {}',
  'export @dec class C {}',
  'declare class C { constructor(x: T); method(x: T): T; }',
]);

const types = [
  'string',
  'number',
  'boolean',
  'null',
  'undefined',
  'void',
  'never',
  'unknown',
  'any',
  'object',
  'bigint',
  'symbol',
  'this',
  '"literal"',
  '1',
  '-1',
  '1n',
  '-1n',
  'true',
  'A<B<C>>',
  'A.B<T>',
  '(A)',
  '(A | B)[]',
  'A & B | C',
  'keyof T',
  'readonly T[]',
  'unique symbol',
  'typeof object.field',
  'typeof fn<T>',
  '`first\r\nlast`',
  'T[K]',
  'T extends infer U extends T ? U : never',
  '[T, U?]',
  '[...T[]]',
  '[first: T, second?: U, ...rest: V[]]',
  '(x: T) => U',
  '<T>(x: T) => U',
  'new<T>(x: T) => U',
  'abstract new(x: T) => U',
  '{ field?: T; readonly value: U; method<T>(x: T): T; }',
  '{ (x: T): U; new(x: T): U; [key: string]: T; }',
  '{ get value(): T; set value(x: T); }',
  '{ [K in keyof T]: T[K] }',
  '{ [K in keyof T]?: T[K] }',
  '{ +readonly [K in keyof T as K]-?: T[K] }',
  '`prefix-${T}-${U}`',
  '`first\r\n${T}\r\nlast`',
  'import("mod").Value<T>',
  'typeof import("mod").Value<T>',
  'import("mod", {with: {"resolution-mode": "import"}}).Value',
];
add(
  'types',
  types.map((type) => `type Result<T, U, V, K> = ${type};`),
);
add('interfaces', [
  'interface I<T> extends A.B<T>, C<D> { value: T; }',
  'interface I { method?: <T>(x: T) => T; method(x: T): T; }',
  'interface I { readonly [key: string]: T; new<T>(x: T): T; <T>(x: T): T; }',
]);

add('modules', [
  'import value from "mod";',
  'import * as ns from "mod";',
  'import {value, other as local, type Type} from "mod";',
  'import {"remote-name" as local} from "mod";',
  'import type {Type} from "mod";',
  'import value from "mod" with {type: "json"};',
  'import defer * as ns from "mod";',
  'import Local = NS.Value;',
  'import Local = require("mod");',
  'export {local as "remote-name"};',
  'export {value as local} from "mod";',
  'export type {Type};',
  'export * from "mod";',
  'export * as ns from "mod";',
  'export type * from "mod";',
  'export = value;',
  'export as namespace Library;',
  'export const value = 1;',
  'export declare const value: T;',
  'export declare function f<T>(x: T): T;',
  'export default abstract class C {}',
  'namespace A.B { export const value = 1; }',
  'declare namespace A.B {}',
  'declare module "mod" {}',
  'declare global { interface I {} }',
  'enum E { A, B = 1, C = "text" }',
  'const enum E {}',
  'declare enum E {}',
]);

add('statements', [
  'if (value) call(); else other();',
  'if (value) { call(); } else if (other) {}',
  'for (;;) {}',
  'for (let i = 0; i < 3; i++) {}',
  'for (key in object) {}',
  'for (const value of values) {}',
  'async function f() { for await (const value of values) {} }',
  'while (value) call();',
  'do {} while (value);',
  'label: for (;;) { continue label; break label; }',
  'try {} catch {}',
  'try {} catch ({source: local = 1}) {} finally {}',
  'switch (value) { case 1: call(); break; default: other(); }',
  'debugger; ;',
  'using value = getValue();',
  'await using value = getValue();',
]);

const jsxStart = CONFORMANCE_CASES.length;
add(
  'jsx',
  [
    '<A />',
    '<A.B />',
    '<ns:element ns:attribute="value" />',
    '<a-b data-value="value" />',
    '<this />',
    '<this.Component />',
    '<A enabled disabled={false} {...attrs} />',
    '<A value="\\" />',
    "<A value='\\' />",
    '<A value="first\nsecond" />',
    '<A value="first\r\nsecond" />',
    '<A value="&amp; &#x1F600;" />',
    '<A>text\r\nnext &amp;</A>',
    '<A>{value}</A>',
    '<A>{/* comment */}</A>',
    '<A>{...children}</A>',
    '<><A />text{value}</>',
    '<A<B<C>> value={left >= right} />',
  ].map((expression) => `const element = ${expression};`),
  'tsx',
);
for (const fixture of CONFORMANCE_CASES.slice(jsxStart)) {
  // Espree supports neither TS type arguments nor JSX spread children.
  // Both extensions are exercised through the TSX oracle above.
  if (
    fixture.text.includes('<A<B<C>>') ||
    fixture.text.includes('{...children}')
  )
    continue;
  CONFORMANCE_CASES.push({
    ...fixture,
    name: `JavaScript ${fixture.name}`,
    filePath: 'case.jsx',
  });
  CONFORMANCE_CASES.push({
    ...fixture,
    name: `JSX in js ${fixture.name}`,
    filePath: 'case.js',
  });
}

add(
  'literals',
  [
    "'use strict'; function f() { 'directive'; 'second'; call(); 'late'; }",
    "('use strict'); function f() { ('use strict'); }",
    'const value = 0xFF + 0b101 + 0o17 + 1_000 + .5e3 + 1n;',
    String.raw`const value = '\uD800\u0061\x62\n';`,
    'const value = `first\r\nlast`;',
    'const value = `first\r\n${x}\r\nlast`;',
    String.raw`const value = tag\`\unicode\`;`.replaceAll('\\`', '`'),
    'const value = /pattern/dgiu;',
    'const value = `${ /[/>]/g.test(text) }`;',
  ],
  'js',
);

add(
  'trivia',
  [
    '  const value = 1;  ',
    '// leading\nconst value = 1;\n// trailing',
    '/* leading */ const value = 1; /* trailing */',
    '// comment only',
    '/* comment only */',
    '',
    '  \n ',
    '\ufeffconst value = 1;',
    '#!/usr/bin/env node\nconst value = 1;',
    '/** @type {T} */ (value);',
    '/** @type {T} */ value;',
    '/** @param {T} value */ function f(value) {}',
  ],
  'js',
);

add(
  'plain JavaScript',
  [
    'const {source: local = 1} = input;',
    'const f = ({source: local = 1}) => local;',
    'class C { constructor() {} static constructor() {} } class D { "constructor"() {} }',
    '(a?.b)(); new C; const f = x => x;',
    'class C { #field = 1; static { init(); } get field() {} set field(value) {} }',
    'const value = {async method() {}, *generator() {}, get field() {}, set field(value) {}};',
  ],
  'js',
);
add(
  'script',
  ['with (object) { call(); }', 'var yield = 1, let = 2, static = 3;'],
  'js',
  'script',
);

const propertyNames = [
  'true',
  'false',
  'null',
  'this',
  'class',
  'if',
  'enum',
  'await',
  'yield',
  'let',
  'static',
  'get',
  'set',
  'constructor',
  'async',
  'type',
  'private',
  'public',
  'implements',
  'interface',
  'readonly',
];
for (const extension of ['js', 'ts']) {
  add(
    `keyword properties ${extension}`,
    propertyNames.flatMap((name) => [
      `const object = {${name}: value}; object.${name};`,
      `class C { ${name}() {} }`,
    ]),
    extension,
  );
}

for (const extension of ['jsx', 'tsx']) {
  add(
    `JSX boundaries ${extension}`,
    [
      '<ns:foo-bar ns:data-key="value" />',
      '<A value="&#xD800;&#xDC00;&#x110000;&#00000000000000000000097;" />',
      '<A>&#xD800;&#xDC00;&#x110000;&#00000000000000000000097;</A>',
      '<A value="&amp;&unknown;&amp;amp;& #12;&nested&amp;" />',
    ].map((expression) => `const element = ${expression};`),
    extension,
  );
}

add('ambient exports', [
  'export declare class C {}',
  'export declare enum E {A}',
  'export declare namespace N {}',
  'export function f(x: T): T;',
]);

add(
  'numeric boundaries',
  [
    '0',
    '-0',
    '1e400',
    '1e-400',
    '0xFFFFFFFFFFFFFFFFFFFFFFFF',
    '0b11111111111111111111111111111111111111111111111111111111111111111',
    '0o777777777777777777777777777',
    '9_007_199_254_740_993',
    '0x20000000000001n',
    '-0x20000000000001n',
  ].map((value) => `const number = ${value};`),
  'js',
);
add(
  'template boundaries',
  [
    '`first\rlast`',
    '`first\u2028last`',
    '`first\u2029last`',
    String.raw`tag\`\unicode\`;`.replaceAll('\\`', '`'),
    '`first\\\r\nlast`',
    '`first\\\r\n${value}last`',
  ].map((value) => `const template = ${value};`),
  'js',
);

add('schema coverage', [
  'throw value;',
  'class C extends Base { constructor() { super(); this.value = 1; } }',
  'type Builtin = intrinsic; abstract class C { abstract field: T; }',
]);

add('parenthesized generics', [
  'const value = (fn<T>)(argument);',
  'const value = (fn<T>)?.(argument);',
  'const value = (tag<T>)`text${argument}`;',
  'const value = new ((Ctor<T>))(argument);',
]);

for (const extension of ['jsx', 'tsx']) {
  add(
    `long entity prefix ${extension}`,
    [`const element = <A>${'&'.repeat(20000)}amp;</A>;`],
    extension,
  );
}

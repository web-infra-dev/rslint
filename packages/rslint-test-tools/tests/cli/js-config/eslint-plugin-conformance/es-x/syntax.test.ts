/**
 * Conformance: eslint-plugin-es-x (syntax) mounted in rslint via `plugins` must
 * report identically to ESLint v10. es-x rules are AST pattern matches over ES
 * version features / builtin APIs (no type info), so rslint reproduces ESLint
 * byte-for-byte. Representative triggers from the upstream test suite (v9.6.0).
 */
import { runConformanceSuite } from '../conformance.js';
import type { DiffCase } from '../harness.js';

const CASES: DiffCase[] = [
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-accessor-properties',
    code: '({ get a() {} })',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-accessor-properties',
    code: '({ set a(value) {} })',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-accessor-properties',
    code: 'class A { get a() {} }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-accessor-properties',
    code: 'class A { set a(value) {} }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-arbitrary-module-namespace-names',
    code: 'export * as "ns" from "mod"',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-arbitrary-module-namespace-names',
    code: 'export {foo as "bar"} from "mod"',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-arbitrary-module-namespace-names',
    code: 'export {"foo" as "bar"} from "mod"',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-arbitrary-module-namespace-names',
    code: 'import {"foo" as bar} from "mod"',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-arrow-functions', code: '() => 1' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-arrow-functions', code: '() => {}' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-arrow-functions',
    code: '() => this.data',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-arrow-functions', code: 'a => a' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-async-functions',
    code: 'async function f() {}',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-async-functions',
    code: 'const f = async function() {}',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-async-functions',
    code: 'const f = async () => {}',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-async-functions',
    code: '({ async method() {} })',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-async-iteration',
    code: 'async function* f() {}',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-async-iteration',
    code: 'const f = async function*() {}',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-async-iteration',
    code: '({ async* method() {} })',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-async-iteration',
    code: 'class A { async* method() {} }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-asyncdisposablestack',
    code: 'AsyncDisposableStack',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-asyncdisposablestack',
    code: 'function f() { AsyncDisposableStack }',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-bigint', code: '100n' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-bigint', code: '({ 100n: null })' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-bigint', code: '({ 100n() {} })' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-bigint', code: '({ get 100n() {} })' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-binary-numeric-literals',
    code: '0b01',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-binary-numeric-literals',
    code: '0B01',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-block-scoped-functions',
    code: '{ function f() {} }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-block-scoped-functions',
    code: 'if (a) { function f() {} }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-block-scoped-functions',
    code: 'if (a) ; else { function f() {} }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-block-scoped-functions',
    code: 'function wrap() { if (a) { function f() {} } }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-block-scoped-variables',
    code: 'const a = 1',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-block-scoped-variables',
    code: 'let a = 1',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-block-scoped-variables',
    code: '{ const a = 1 }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-block-scoped-variables',
    code: '{ let a = 1 }',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-classes', code: 'class A {}' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-classes', code: '(class {})' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-computed-properties',
    code: '({ [a]: 1 })',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-computed-properties',
    code: '({ [a]() {} })',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-computed-properties',
    code: '({ get [a]() {} })',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-computed-properties',
    code: '({ set [a](value) {} })',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-default-parameters',
    code: 'async function f(a = 0) {}',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-default-parameters',
    code: 'const f = async function(a = 0) {}',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-default-parameters',
    code: 'const f = async (a = 0) => {}',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-default-parameters',
    code: '({ async method(a = 0) {} })',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-destructuring',
    code: 'var [a, {b: []}, [c], [d] = ary1, ...e] = ary2',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-destructuring',
    code: 'let [a, {b: []}, [c], [d] = ary1, ...e] = ary2',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-destructuring',
    code: 'const [a, {b: []}, [c], [d] = ary1, ...e] = ary2',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-destructuring',
    code: '([a, {b: []}, [c], [d] = ary1, ...e] = ary2)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-disposablestack',
    code: 'DisposableStack',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-disposablestack',
    code: 'function f() { DisposableStack }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-dynamic-import',
    code: 'import(source)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-dynamic-import-options',
    code: 'const module = await import(source, options)',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-escape-unescape', code: 'escape' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-escape-unescape', code: 'unescape' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-escape-unescape', code: "escape('')" },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-escape-unescape',
    code: "unescape('')",
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-exponential-operators', code: 'a**b' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-exponential-operators',
    code: 'a**=b',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-export-ns-from',
    code: 'export * as ns from "mod"',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-float16array', code: 'Float16Array' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-float16array',
    code: 'function f() { Float16Array }',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-for-of-loops', code: 'for(a of b);' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-for-of-loops',
    code: 'for(var a of b);',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-for-of-loops',
    code: 'for(let a of b);',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-for-of-loops',
    code: 'for(const a of b);',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-function-prototype-bind',
    code: '(function fn(){}).bind(this)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-function-prototype-bind',
    code: '(()=>{}).bind(this)',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-global-this', code: 'globalThis' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-hashbang',
    code: "#!/usr/bin/env node\n            // in the Script Goal\n            'use strict';\n            console.log(1);\n            ",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-hashbang',
    code: '#!/usr/bin/env node\n            // in the Module Goal\n            export {};\n            console.log(1);\n            ',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-import-attributes',
    code: "import foo from 'foo' with { type: 'json' }",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-import-attributes',
    code: "export {foo} from 'foo' with { type: 'json' }",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-import-attributes',
    code: "export * from 'foo' with { type: 'json' }",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-import-attributes',
    code: "import('foo', { with: { type: 'json'} })",
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-import-meta', code: 'import.meta' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-keyword-properties',
    code: '({ if: 1 })',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-keyword-properties',
    code: '({ static: 2 })',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-keyword-properties', code: 'obj.if' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-keyword-properties',
    code: 'obj.class',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-labelled-function-declarations',
    code: 'label: function f() {}',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-legacy-object-prototype-accessor-methods',
    code: 'foo.__defineGetter__(prop, func)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-legacy-object-prototype-accessor-methods',
    code: 'foo.__defineSetter__(prop, val, func)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-legacy-object-prototype-accessor-methods',
    code: 'foo.__lookupGetter__(prop)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-legacy-object-prototype-accessor-methods',
    code: 'foo.__lookupSetter__(prop)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-logical-assignment-operators',
    code: 'x ||= y',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-logical-assignment-operators',
    code: 'x &&= y',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-logical-assignment-operators',
    code: 'x ??= y',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-logical-assignment-operators',
    code: 'a.b ||= c',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-malformed-template-literals',
    code: 'tag`\\unicode`',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-malformed-template-literals',
    code: 'tag`\\unicode${a}\\unicode${b}\\unicode${c}unicode`',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-modules', code: "import x from 'x'" },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-modules',
    code: "import * as x from 'x'",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-modules',
    code: "import x, {y, z} from 'x'",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-modules',
    code: "export {x} from 'x'",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-new-target',
    code: 'class A { constructor() { new.target } }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nullish-coalescing-operators',
    code: 'a??b',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nullish-coalescing-operators',
    code: ' /* ?? comment ?? */\n            a /* ?? comment ?? */\n            ?? /* ?? comment ?? */\n            b /* ?? comment ?? */',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nullish-coalescing-operators',
    code: 'a ?? b ?? c',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nullish-coalescing-operators',
    code: '(a ?? b) ?? c',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-octal-numeric-literals',
    code: '0o123',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-octal-numeric-literals',
    code: '0O123',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-optional-catch-binding',
    code: 'try {} catch {}',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-optional-chaining',
    code: 'var x = a?.b',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-optional-chaining',
    code: 'var x = a?.[b]',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-optional-chaining', code: 'foo?.()' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-optional-chaining',
    code: 'var x = ((a?.b)?.c)?.()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-private-in',
    code: 'class A { #x; f(obj) { return #x in obj } }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-private-in',
    code: 'class A { #x; f(obj) { return #x in obj.foo } }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-private-in',
    code: 'class A { #x; f(obj) { return #x in obj.#x } }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-property-shorthands',
    code: '({ a })',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-property-shorthands',
    code: '({ a() {} })',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-property-shorthands',
    code: '({ * a() {} })',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-property-shorthands',
    code: '({ [a]() {} })',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-proxy', code: 'Proxy' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-proxy',
    code: 'function f() { Proxy }',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-reflect', code: 'Reflect' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-reflect',
    code: 'function f() { Reflect }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-rest-parameters',
    code: 'function f(...a) {}',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-rest-parameters',
    code: '(function(...a) {})',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-rest-parameters',
    code: '(...a) => {}',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-rest-parameters',
    code: '({ f(...a) {} })',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-rest-spread-properties',
    code: '({...a})',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-rest-spread-properties',
    code: '({...a} = obj)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-rest-spread-properties',
    code: 'for ({...a} of x) {}',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-rest-spread-properties',
    code: 'function f({...a}) {}',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-shadow-catch-param',
    code: 'try {} catch (e) { var e }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-shadow-catch-param',
    code: 'try {} catch (err) { var err; var err; }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-shadow-catch-param',
    code: 'try {} catch (e) { if (foo) {var e} }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-shared-array-buffer',
    code: 'SharedArrayBuffer',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-shared-array-buffer',
    code: 'function f() { SharedArrayBuffer }',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-spread-elements', code: 'f(...a, b)' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-spread-elements', code: 'f(a, ...b)' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-spread-elements',
    code: 'new F(...a, b)',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-spread-elements', code: '[...a, b]' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-suppressederror',
    code: 'SuppressedError',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-suppressederror',
    code: 'function f() { SuppressedError }',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-temporal', code: 'Temporal' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-temporal',
    code: 'function f() { Temporal }',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-top-level-await', code: 'await expr' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-top-level-await',
    code: 'for await (a of b);',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-top-level-await',
    code: 'for await (var a of b);',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-top-level-await',
    code: 'for await (let a of b);',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-trailing-commas',
    code: 'var a = [1,]',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-trailing-commas',
    code: 'var obj = {a,}',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-trailing-commas',
    code: 'var obj = {a:1,}',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-trailing-dynamic-import-commas',
    code: 'import(source,)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-trailing-dynamic-import-commas',
    code: 'import(source, options,)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-trailing-function-commas',
    code: 'function f(a,) {}',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-trailing-function-commas',
    code: '(function(a,) {})',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-trailing-function-commas',
    code: '(a,) => {}',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-trailing-function-commas',
    code: '({ f(a,) {} })',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-using-declarations',
    code: 'using x = y',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-using-declarations',
    code: 'await using x = y',
  },
];

const CLEAN_CASES: DiffCase[] = [
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-accessor-properties',
    code: '({ get: function() {} })',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-accessor-properties',
    code: '({ set: function() {} })',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-arbitrary-module-namespace-names',
    code: 'export * from "mod"',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-arbitrary-module-namespace-names',
    code: 'export * as ns from "mod"',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-arrow-functions',
    code: 'function f() {}',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-arrow-functions',
    code: 'const f = function() {}',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-async-functions',
    code: 'function f() {}',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-async-functions',
    code: 'const f = function() {}',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-async-iteration',
    code: 'async function f() {}',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-async-iteration',
    code: 'const f = async function() {}',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-asyncdisposablestack', code: 'Array' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-asyncdisposablestack',
    code: 'Object',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-bigint', code: '100' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-binary-numeric-literals', code: '1' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-binary-numeric-literals',
    code: '1e10',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-block-scoped-functions',
    code: 'function f() {}',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-block-scoped-variables',
    code: 'var a = 1',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-block-scoped-variables',
    code: 'function f(a) {}',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-classes',
    code: 'function A() {} A.prototype.foo = function() {}',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-computed-properties',
    code: '({ foo: 1 })',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-computed-properties',
    code: '({ foo })',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-dataview-prototype-getfloat16-setfloat16',
    code: 'let foo = new DataView(); ',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-dataview-prototype-getfloat16-setfloat16',
    code: 'foo.getFloat32()',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-default-parameters',
    code: 'function f(a, ...rest) {}',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-default-parameters',
    code: 'const f = function(a, ...rest) {}',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-destructuring', code: '({})' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-destructuring', code: '({a: 1})' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-disposablestack', code: 'Array' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-disposablestack', code: 'Object' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-dynamic-import',
    code: "import a from 'a'",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-dynamic-import',
    code: 'obj.\nimport(source)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-dynamic-import-options',
    code: 'import(source)',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-escape-unescape', code: 'encodeURI' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-escape-unescape', code: 'decodeURI' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-exponential-operators', code: 'a*b' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-exponential-operators', code: 'a*=b' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-export-ns-from',
    code: 'export * from "mod"',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-export-ns-from',
    code: 'export default foo',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-float16array', code: 'Array' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-float16array', code: 'Object' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-for-of-loops', code: 'for(;;);' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-for-of-loops', code: 'for(a in b);' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-function-prototype-bind',
    code: 'foo.bind(this)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-function-prototype-bind',
    code: 'bind(this)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-global-this',
    code: 'window.globalThis',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-global-this', code: 'window' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-global-this', code: 'global' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-hashbang', code: '/* comment */' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-hashbang', code: '/** comment */' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-import-attributes',
    code: "import foo from 'foo'",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-import-attributes',
    code: "export {foo} from 'foo'",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-import-meta',
    code: "import * as Foo from 'foo'",
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-import-meta', code: "import('foo')" },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-initializers-in-for-in',
    code: 'for (var x in obj) {}',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-initializers-in-for-in',
    code: 'for (let x in obj) {}',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-keyword-properties',
    code: '({ a, b, c}.a)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-keyword-properties',
    code: '({ let: 1, of: 2}.let)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-labelled-function-declarations',
    code: 'function f() {}',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-labelled-function-declarations',
    code: 'label: { function f() {} }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-legacy-object-prototype-accessor-methods',
    code: 'Object.defineProperty(foo, prop, descriptor)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-legacy-object-prototype-accessor-methods',
    code: 'Object.getOwnPropertyDescriptor(foo, prop)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-logical-assignment-operators',
    code: 'x = x || y',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-logical-assignment-operators',
    code: 'x = x && y',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-malformed-template-literals',
    code: '`foo`',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-malformed-template-literals',
    code: 'tag`foo`',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-modules',
    code: 'module.exports = {}',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-modules', code: "require('x')" },
  { pkg: 'eslint-plugin-es-x', rule: 'no-new-target', code: 'new F()' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-new-target', code: 'target = 1' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nullish-coalescing-operators',
    code: 'a ? b : c',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-nullish-coalescing-operators',
    code: 'a && b',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-octal-numeric-literals', code: '123' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-octal-numeric-literals',
    code: '0123',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-optional-catch-binding',
    code: 'try {} catch (err) {}',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-optional-chaining',
    code: 'var x = a.b',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-optional-chaining',
    code: 'var x = a[b]',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-private-in',
    code: "class A { f(obj) { return '#x' in obj } }",
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-private-in',
    code: 'class A { f(obj) { return x in obj } }',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-property-shorthands', code: '({})' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-property-shorthands',
    code: '({ a: 1 })',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-proxy', code: 'Array' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-proxy', code: 'Object' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-reflect', code: 'Array' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-reflect', code: 'Object' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-rest-parameters', code: '[a, ...b]' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-rest-parameters',
    code: '[a, ...b] = array',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-rest-spread-properties',
    code: '[...a]',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-rest-spread-properties',
    code: '[...a] = array',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-shadow-catch-param',
    code: 'var e; try {} catch (e) {  }',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-shadow-catch-param',
    code: 'try {} catch (e) {}',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-shared-array-buffer', code: 'Array' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-shared-array-buffer', code: 'Object' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-spread-elements',
    code: '[a, ...b] = array',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-spread-elements',
    code: 'function f(a, ...b) {}',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-suppressederror', code: 'Array' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-suppressederror', code: 'Object' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-temporal', code: 'Array' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-temporal', code: 'Object' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-top-level-await',
    code: 'async function f() { await expr }',
  },
  { pkg: 'eslint-plugin-es-x', rule: 'no-top-level-await', code: 'expr;' },
  { pkg: 'eslint-plugin-es-x', rule: 'no-trailing-commas', code: 'var a = []' },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-trailing-commas',
    code: 'var a = [a]',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-trailing-dynamic-import-commas',
    code: 'import(source)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-trailing-dynamic-import-commas',
    code: 'import(source, options)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-trailing-function-commas',
    code: '[1,]',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-trailing-function-commas',
    code: '({a:1,})',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-using-declarations',
    code: 'let x = y',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-using-declarations',
    code: 'const x = y',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-weakmap-prototype-getorinsert',
    code: 'foo.(key, value)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-weakmap-prototype-getorinsert',
    code: '(key, value)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-weakmap-prototype-getorinsertcomputed',
    code: 'foo.(key, callbackFn)',
  },
  {
    pkg: 'eslint-plugin-es-x',
    rule: 'no-weakmap-prototype-getorinsertcomputed',
    code: '(key, callbackFn)',
  },
];

runConformanceSuite('eslint-plugin-es-x', CASES, CLEAN_CASES);

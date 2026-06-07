/**
 * Conformance: eslint-plugin-unicorn rules mounted in rslint via `plugins` must
 * report identically to ESLint v10. Shared assertion + excluded-category notes
 * live in ./conformance.ts.
 */
import { runConformanceSuite } from './conformance.js';
import type { DiffCase } from './harness.js';

/** 141 rules that report IDENTICALLY on a minimal trigger. */
const CASES: DiffCase[] = [
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'better-regex',
    code: 'const re = /[0-9]/;\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'catch-error-name',
    code: 'try {\n\tdoSomething();\n} catch (err) {\n\tconsole.log(err);\n}\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'consistent-assert',
    code: "import assert from 'node:assert';\nassert(foo);\n",
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'consistent-date-clone',
    code: 'const a = new Date(date.getTime());\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'consistent-destructuring',
    code: 'const {a} = foo;\nconsole.log(foo.b);\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'consistent-empty-array-spread',
    code: "const list = [...(foo ? [1, 2] : '')];\n",
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'consistent-existence-index-check',
    code: "const index = foo.indexOf('x');\nif (index < 0) {\n\tconsole.log(1);\n}\n",
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'consistent-function-scoping',
    code: 'function outer() {\n\tfunction inner() {\n\t\treturn 1;\n\t}\n\treturn inner();\n}\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'consistent-template-literal-escape',
    code: 'const a = `$\\{foo}`;\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'custom-error-definition',
    code: 'class fooError extends Error {}\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'empty-brace-spaces',
    code: 'class A { }\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'error-message',
    code: 'throw new Error();\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'escape-case',
    code: "const a = '\\xa9';\n",
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'explicit-length-check',
    code: 'const foo = [];\nif (foo.length) {\n\tconsole.log(1);\n}\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'import-style',
    code: "import {join} from 'node:path';\nconsole.log(join);\n",
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'isolated-functions',
    code: '// @isolated\nconst fn = () => {\n\treturn outsideVariable;\n};\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'new-for-builtins',
    code: 'const m = Map();\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'no-accessor-recursion',
    code: 'const obj = {\n\tget foo() {\n\t\treturn this.foo;\n\t},\n};\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'no-anonymous-default-export',
    code: 'export default function () {}\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'no-array-callback-reference',
    code: 'const result = array.map(callback);\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'no-array-for-each',
    code: 'array.forEach((x) => {\n\tconsole.log(x);\n});\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'no-array-method-this-argument',
    code: 'const result = array.map(function () {}, thisArgument);\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'no-array-reduce',
    code: 'const total = array.reduce((accumulator, element) => accumulator.concat(element));\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'no-array-reverse',
    code: 'const reversed = array.reverse();\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'no-array-sort',
    code: 'const sorted = array.sort();\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'no-await-expression-member',
    code: 'async function f() {\n\tconst a = (await foo).bar;\n\treturn a;\n}\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'no-await-in-promise-methods',
    code: 'async function f() { return Promise.all([await Promise.resolve(1)]); }',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'no-console-spaces',
    code: 'console.log("foo ", "bar");',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'no-empty-file',
    code: '// just a comment\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'no-for-loop',
    code: 'const arr = [1, 2, 3];\nfor (let i = 0; i < arr.length; i++) {\n  console.log(arr[i]);\n}',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'no-hex-escape',
    code: 'const x = "\\x41";',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'no-immediate-mutation',
    code: 'const arr = [];\narr.push(1);',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'no-instanceof-builtins',
    code: 'const x = foo instanceof Array;',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'no-invalid-fetch-options',
    code: 'fetch("/", { method: "GET", body: "x" });',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'no-invalid-remove-event-listener',
    code: 'el.removeEventListener("click", handler.bind(this));',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'no-keyword-prefix',
    code: 'const newFoo = 1;',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'no-lonely-if',
    code: 'if (a) {\n  if (b) {\n    foo();\n  }\n}',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'no-magic-array-flat-depth',
    code: 'const x = foo.flat(2);',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'no-named-default',
    code: 'import { default as foo } from "foo";',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'no-negated-condition',
    code: 'if (!a) {\n  foo();\n} else {\n  bar();\n}',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'no-negation-in-equality-check',
    code: 'const x = !a === b;',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'no-nested-ternary',
    code: 'const x = a ? b ? c : d : e;',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'no-new-array',
    code: 'const x = new Array(5);',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'no-new-buffer',
    code: 'const x = new Buffer(10);',
  },
  { pkg: 'eslint-plugin-unicorn', rule: 'no-null', code: 'const x = null;' },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'no-object-as-default-parameter',
    code: 'function foo(bar = { a: 1 }) {}',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'no-process-exit',
    code: 'process.exit(1);',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'no-single-promise-in-promise-methods',
    code: 'async function f() { return Promise.all([foo()]); }',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'no-thenable',
    code: 'export function then() {}',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'no-this-assignment',
    code: 'function f() {\n  const self = this;\n  return self;\n}',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'no-typeof-undefined',
    code: 'const foo = 1;\nconst x = typeof foo === "undefined";',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'no-unnecessary-array-flat-depth',
    code: 'const x = foo.flat(1);',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'no-unnecessary-array-splice-count',
    code: 'const arr = [1,2,3];\narr.splice(1, arr.length);\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'no-unnecessary-await',
    code: 'async function f() {\n  await 1;\n}\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'no-unnecessary-polyfills',
    code: "import 'array.prototype.flat';\n",
    options: [{ targets: 'node 18' }],
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'no-unnecessary-slice-end',
    code: 'const arr = [1,2,3];\narr.slice(1, arr.length);\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'no-unreadable-array-destructuring',
    code: 'const [, , , foo] = parts;\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'no-unreadable-iife',
    code: 'const foo = (bar => (bar ? bar.baz : baz))(getBar());\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'no-unused-properties',
    code: "const data = {\n  used: 'a',\n  unused: 'b',\n};\nconsole.log(data.used);\n",
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'no-useless-collection-argument',
    code: 'const set = new Set([]);\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'no-useless-error-capture-stack-trace',
    code: 'class A extends Error {\n  constructor() {\n    super();\n    Error.captureStackTrace(this, A);\n  }\n}\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'no-useless-fallback-in-spread',
    code: 'const foo = {...(bar || {})};\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'no-useless-iterator-to-array',
    code: 'const result = [...foo.values().toArray()];\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'no-useless-length-check',
    code: 'const foo = [];\nif (foo.length === 0 || foo.every(Boolean)) {\n}\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'no-useless-promise-resolve-reject',
    code: 'async function foo() {\n  return Promise.resolve(bar);\n}\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'no-useless-spread',
    code: 'const foo = [...[1, 2, 3]];\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'no-useless-switch-case',
    code: 'switch (foo) {\n  case 1:\n  default:\n    break;\n}\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'no-useless-undefined',
    code: 'function foo() {\n  return undefined;\n}\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'no-zero-fractions',
    code: 'const foo = 1.0;\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'number-literal-case',
    code: 'const foo = 0XFF;\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'numeric-separators-style',
    code: 'const foo = 1000000;\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'prefer-add-event-listener',
    code: 'foo.onclick = () => {};\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'prefer-array-find',
    code: 'const foo = array.filter(x => x.y)[0];\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'prefer-array-flat',
    code: 'const foo = [].concat(maybeArray);\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'prefer-array-flat-map',
    code: 'const foo = array.map(x => x).flat();\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'prefer-array-index-of',
    code: 'const index = foo.findIndex(x => x === bar);\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'prefer-array-some',
    code: 'const foo = array.find(x => x.y) ? 1 : 2;\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'prefer-at',
    code: 'const foo = array[array.length - 1];\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'prefer-bigint-literals',
    code: 'const foo = BigInt(1);\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'prefer-blob-reading-methods',
    code: 'const reader = new FileReader();\nreader.readAsArrayBuffer(blob);\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'prefer-class-fields',
    code: 'class A {\n  constructor() {\n    this.foo = 1;\n  }\n}\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'prefer-classlist-toggle',
    code: "if (foo) {\n  element.classList.add('bar');\n} else {\n  element.classList.remove('bar');\n}\n",
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'prefer-code-point',
    code: "const x = 'a'.charCodeAt(0);\n",
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'prefer-date-now',
    code: 'const x = new Date().getTime();\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'prefer-default-parameters',
    code: "function foo(bar) {\n  bar = bar || 'baz';\n  return bar;\n}\n",
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'prefer-dom-node-append',
    code: 'declare const foo: any;\ndeclare const bar: any;\nfoo.appendChild(bar);\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'prefer-dom-node-dataset',
    code: "declare const element: any;\nelement.setAttribute('data-unicorn', 'foo');\n",
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'prefer-dom-node-remove',
    code: 'declare const parentNode: any;\ndeclare const childNode: any;\nparentNode.removeChild(childNode);\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'prefer-dom-node-text-content',
    code: 'declare const foo: any;\nconst x = foo.innerText;\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'prefer-event-target',
    code: 'class Foo extends EventEmitter {}\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'prefer-export-from',
    code: "import {foo} from 'foo';\nexport {foo};\n",
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'prefer-global-this',
    code: 'const x = window.location;\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'prefer-import-meta-properties',
    code: "import {fileURLToPath} from 'node:url';\nconst filename = fileURLToPath(import.meta.url);\n",
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'prefer-includes',
    code: 'declare const arr: number[];\nconst x = arr.indexOf(1) !== -1;\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'prefer-json-parse-buffer',
    code: "import fs from 'node:fs';\nconst data = JSON.parse(fs.readFileSync('./foo.json', 'utf8'));\n",
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'prefer-keyboard-event-key',
    code: "declare const element: any;\nelement.addEventListener('keydown', (event: any) => {\n  if (event.keyCode === 27) {}\n});\n",
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'prefer-logical-operator-over-ternary',
    code: "declare const foo: any;\nconst x = foo ? foo : 'bar';\n",
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'prefer-math-min-max',
    code: 'declare const height: number;\nconst x = height > 10 ? 10 : height;\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'prefer-math-trunc',
    code: 'declare const x: number;\nconst y = ~~x;\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'prefer-modern-dom-apis',
    code: 'declare const oldChildNode: any;\ndeclare const newChildNode: any;\ndeclare const parentNode: any;\nparentNode.replaceChild(newChildNode, oldChildNode);\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'prefer-modern-math-apis',
    code: 'declare const x: number;\ndeclare const y: number;\nconst z = Math.sqrt(x * x + y * y);\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'prefer-module',
    code: "'use strict';\nconst x = 1;\n",
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'prefer-native-coercion-functions',
    code: 'const x = (v: unknown) => String(v);\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'prefer-negative-index',
    code: 'declare const foo: any[];\nconst x = foo.slice(foo.length - 1);\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'prefer-node-protocol',
    code: "import path from 'path';\nconst x = path.join('a', 'b');\n",
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'prefer-number-properties',
    code: 'const x = isNaN(1);\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'prefer-object-from-entries',
    code: 'declare const pairs: [string, number][];\nconst x = pairs.reduce((obj, [key, value]) => ({...obj, [key]: value}), {});\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'prefer-optional-catch-binding',
    code: 'try {\n  doSomething();\n} catch (error) {\n  doSomethingElse();\n}\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'prefer-prototype-methods',
    code: 'const x = [].slice.call(arguments);\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'prefer-query-selector',
    code: "declare const document: any;\nconst x = document.getElementById('foo');\n",
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'prefer-reflect-apply',
    code: 'declare function foo(...args: any[]): void;\nfoo.apply(null, [1, 2]);\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'prefer-regexp-test',
    code: 'declare const foo: string;\nif (/bar/.exec(foo)) {}\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'prefer-response-static-json',
    code: 'const r = new Response(JSON.stringify({a: 1}));\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'prefer-set-has',
    code: 'const list = [1, 2, 3];\nfunction check(v) {\n\treturn list.includes(v);\n}\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'prefer-set-size',
    code: 'const count = [...new Set([1, 2, 3])].length;\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'prefer-simple-condition-first',
    code: 'declare const a: boolean;\ndeclare const b: boolean;\ndeclare const c: boolean;\ndeclare const d: boolean;\nif ((a ? b : c) && d) {\n\tconsole.log(1);\n}\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'prefer-single-call',
    code: 'declare const foo: number[];\nfoo.push(1);\nfoo.push(2);\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'prefer-spread',
    code: 'declare const set: Set<number>;\nconst arr = Array.from(set);\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'prefer-string-raw',
    code: 'const path = "C:\\\\Users\\\\foo";\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'prefer-string-replace-all',
    code: 'declare const str: string;\nconst out = str.replace(/a/g, "b");\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'prefer-string-slice',
    code: 'declare const str: string;\nconst out = str.substr(1, 2);\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'prefer-string-starts-ends-with',
    code: 'declare const str: string;\nconst out = /^a/.test(str);\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'prefer-string-trim-start-end',
    code: 'declare const str: string;\nconst out = str.trimLeft();\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'prefer-structured-clone',
    code: 'declare const obj: object;\nconst clone = JSON.parse(JSON.stringify(obj));\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'prefer-switch',
    code: 'declare const x: number;\nif (x === 1) {\n\tconsole.log(1);\n} else if (x === 2) {\n\tconsole.log(2);\n} else if (x === 3) {\n\tconsole.log(3);\n} else {\n\tconsole.log(4);\n}\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'prefer-ternary',
    code: 'declare const cond: boolean;\nlet result;\nif (cond) {\n\tresult = 1;\n} else {\n\tresult = 2;\n}\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'prefer-top-level-await',
    code: 'declare const promise: Promise<number>;\npromise.then((value) => {\n\tconsole.log(value);\n});\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'prefer-type-error',
    code: 'declare const x: unknown;\nfunction f() {\n\tif (typeof x !== "string") {\n\t\tthrow new Error("x must be a string");\n\t}\n}\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'prevent-abbreviations',
    code: 'const e = 1;\nconsole.log(e);\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'relative-url-style',
    code: 'const url = new URL("./foo.js", import.meta.url);\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'require-array-join-separator',
    code: 'declare const arr: number[];\nconst out = arr.join();\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'require-module-attributes',
    code: 'import foo from "./foo.js" with {};\nconsole.log(foo);\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'require-module-specifiers',
    code: 'import {} from "./foo.js";\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'require-number-to-fixed-digits-argument',
    code: 'declare const n: number;\nconst out = n.toFixed();\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'require-post-message-target-origin',
    code: 'declare const data: unknown;\nwindow.postMessage(data);\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'string-content',
    code: 'const s = "foo";\nconsole.log(s);\n',
    options: [{ patterns: { foo: 'bar' } }],
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'switch-case-braces',
    code: 'declare const x: number;\nswitch (x) {\n\tcase 1:\n\t\tconsole.log(1);\n\t\tbreak;\n}\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'switch-case-break-position',
    code: 'declare const x: number;\nswitch (x) {\n\tcase 1: {\n\t\tconsole.log(1);\n\t}\n\tbreak;\n}\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'template-indent',
    code: 'declare function html(strings: TemplateStringsArray): string;\nconst markup = html`\n<div></div>\n`;\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'text-encoding-identifier-case',
    code: 'const encoding = "UTF-8";\nconsole.log(encoding);\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'throw-new-error',
    code: 'function f() {\n\tthrow Error("boom");\n}\n',
  },
];

const CLEAN_CASES: DiffCase[] = [
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'no-array-for-each',
    code: 'for (const x of array) { console.log(x); }\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'no-null',
    code: 'const x = undefined;\n',
  },
  {
    pkg: 'eslint-plugin-unicorn',
    rule: 'prefer-includes',
    code: 'const r = array.includes(1);\n',
  },
];

runConformanceSuite('eslint-plugin-unicorn', CASES, CLEAN_CASES);

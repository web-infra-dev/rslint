/**
 * Conformance: eslint-plugin-sonarjs rules mounted in rslint via `eslintPlugins` must
 * report identically to ESLint v10. Shared assertion + excluded-category notes
 * live in ./conformance.ts.
 */
import { runConformanceSuite } from './conformance.js';
import type { DiffCase } from './harness.js';

/** 169 rules that report IDENTICALLY on a minimal trigger. */
const CASES: DiffCase[] = [
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'anchor-precedence',
    code: 'const re = /^a|b|c$/;\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'arguments-order',
    code: 'function f(a, b) { return a - b; }\nconst a = 1;\nconst b = 2;\nf(b, a);\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'arguments-usage',
    code: 'function f() {\n  return arguments;\n}\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'array-constructor',
    code: 'const x = new Array(1, 2, 3);\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'arrow-function-convention',
    code: 'const f = (x) => x;\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'bitwise-operators',
    code: 'function f(a, b) {\n  if (a & b) {\n    return 1;\n  }\n  return 0;\n}\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'block-scoped-var',
    code: 'function f() {\n  {\n    var x = 1;\n  }\n  return x;\n}\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'bool-param-default',
    code: 'function f(a?: boolean) {\n  return a;\n}\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'call-argument-line',
    code: 'foo\n(bar);\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'chai-determinate-assertion',
    code: "import { expect } from 'chai';\nexpect(foo).to.not.throw(TypeError);\n",
  },
  { pkg: 'eslint-plugin-sonarjs', rule: 'class-name', code: 'class foo {}\n' },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'class-prototype',
    code: 'Foo.prototype.bar = function () {\n  return 1;\n};\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'code-eval',
    code: "const x = 'alert(1)';\neval(x);\n",
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'cognitive-complexity',
    code: 'function f(a, b) {\n  if (a) {\n    if (b) {\n      return 1;\n    }\n  }\n  return 0;\n}\n',
    options: [0],
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'comma-or-logical-or-case',
    code: 'switch (x) {\n  case 1, 2:\n    break;\n}\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'comment-regex',
    code: '// TODO fix this\nconst x = 1;\n',
    options: [{ regularExpression: 'TODO' }],
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'concise-regex',
    code: 'const re = /[0-9]/;\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'conditional-indentation',
    code: 'if (cond)\nfoo();\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'constructor-for-side-effects',
    code: 'new Foo();\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'cyclomatic-complexity',
    code: 'function f(a, b) {\n  if (a) {\n    return 1;\n  }\n  if (b) {\n    return 2;\n  }\n  return 0;\n}\n',
    options: [{ threshold: 1 }],
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'declarations-in-global-scope',
    code: 'var x = 1;\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'destructuring-assignment-syntax',
    code: 'const a = obj.a;\nconst b = obj.b;\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'disabled-timeout',
    code: "import 'chai';\ndescribe('s', function () {\n  this.timeout(2147483648);\n});\n",
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'dompurify-unsafe-config',
    code: "import DOMPurify from 'dompurify';\nDOMPurify.sanitize(input, { ADD_ATTR: ['onerror'] });\n",
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'duplicates-in-character-class',
    code: 'const re = /[aa]/;\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'elseif-without-else',
    code: 'if (a) { foo(); } else if (b) { bar(); }\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'empty-string-repetition',
    code: 'const re = /(?:)*/;\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'encryption-secure-mode',
    code: "import * as crypto from 'crypto';\nconst cipher = crypto.createCipheriv('aes-256-ecb', key, iv);\n",
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'expression-complexity',
    code: 'const x = a || b || c || d;\n',
    options: [{ max: 2 }],
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'fixme-tag',
    code: '// FIXME: do something\nconst x = 1;\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'for-in',
    code: 'for (const k in obj) {\n  doSomething(k);\n}\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'for-loop-increment-sign',
    code: 'for (let i = 0; i < 10; i--) {\n  use(i);\n}\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'function-inside-loop',
    code: 'var counter = 0;\nfor (let i = 0; i < 10; i++) {\n  const f = () => counter;\n  counter++;\n  f();\n}\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'function-name',
    code: 'function MyFunc() {\n  return 1;\n}\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'future-reserved-words',
    code: 'function f(interface) { return interface; }\nconst x = 1;\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'generator-without-yield',
    code: 'function* gen() {\n  const x = 1;\n  return x;\n}\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'inconsistent-function-call',
    code: 'function Foo() {}\nconst a = new Foo();\nconst b = Foo();\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'inverted-assertion-arguments',
    code: "it('t', () => {\n  assert.equal(42, actual);\n});\n",
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'label-position',
    code: 'myLabel: {\n  doSomething();\n}\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'max-lines-per-function',
    code: 'function big() {\n  const a = 1;\n  const b = 2;\n  return a + b;\n}\n',
    options: [{ maximum: 1 }],
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'max-switch-cases',
    code: 'switch (x) {\n  case 1: a(); break;\n  case 2: b(); break;\n  case 3: c(); break;\n}\n',
    options: [2],
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'max-union-size',
    code: 'function f(x: number | string | boolean | null): void {}\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'misplaced-loop-counter',
    code: 'for (let i = 0; i < 10; j++) {\n  use(i);\n}\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'nested-control-flow',
    code: 'if (a) {\n  if (b) {\n    if (c) {\n      if (d) {\n        doIt();\n      }\n    }\n  }\n}\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-all-duplicated-branches',
    code: 'if (cond) {\n  doSomething();\n} else {\n  doSomething();\n}\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-angular-bypass-sanitization',
    code: 'sanitizer.bypassSecurityTrustHtml(userInput);\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-built-in-override',
    code: 'function Object() {}\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-case-label-in-switch',
    code: 'switch (x) {\n  case 1:\n    myLabel: doSomething();\n    break;\n}\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-collapsible-if',
    code: 'if (a) {\n  if (b) {\n    doSomething();\n  }\n}\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-collection-size-mischeck',
    code: 'const arr = [1, 2, 3];\nif (arr.length < 0) {\n  doSomething();\n}\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-control-regex',
    code: 'const re = /\\x00/;\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-delete-var',
    code: 'var x = 1;\ndelete x;\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-duplicate-in-composite',
    code: 'type T = number | string | number;\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-duplicate-string',
    code: "const a = 'duplicate string value';\nconst b = 'duplicate string value';\nconst c = 'duplicate string value';\n",
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-duplicated-branches',
    code: 'if (a) {\n  doSomethingExtra();\n  doSomethingExtra();\n} else {\n  doSomethingExtra();\n  doSomethingExtra();\n}\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-element-overwrite',
    code: "const obj = {};\nobj[1] = 'a';\nobj[1] = 'b';\n",
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-empty-after-reluctant',
    code: 'const re = /a*?/;\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-empty-alternatives',
    code: 'const re = /a|/;\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-empty-character-class',
    code: 'const re = /[]/;\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-empty-collection',
    code: 'const arr = [];\nfor (const x of arr) {\n  console.log(x);\n}\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-empty-group',
    code: 'const re = /a()b/;\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-equals-in-for-termination',
    code: 'for (let i = 0; i == 10; i++) {\n  doSomething();\n}\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-exclusive-tests',
    code: "describe.only('suite', function () {\n  it('t', function () {});\n});\n",
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-extra-arguments',
    code: 'function f(a) {\n  return a;\n}\nf(1, 2, 3);\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-function-declaration-in-block',
    code: 'if (x) {\n  function f() {\n    return 1;\n  }\n  f();\n}\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-global-this',
    code: 'this.foo = 1;\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-globals-shadowing',
    code: 'function f(arguments) {\n  return arguments;\n}\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-gratuitous-expressions',
    code: 'function f(a) {\n  if (a) {\n    if (a) {\n      doSomething();\n    }\n  }\n}\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-hardcoded-ip',
    code: "const ip = '192.168.12.42';\n",
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-hardcoded-passwords',
    code: "const password = 'hardcodedSecret123';\n",
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-hardcoded-secrets',
    code: "const token = 'xK9mP2qR7sT4vW1yZ8bC3dF6gH5jL0nQ4eU';\n",
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-hook-setter-in-body',
    code: "import { useState } from 'react';\nfunction Component() {\n  const [count, setCount] = useState(0);\n  setCount(1);\n  return count;\n}\n",
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-identical-conditions',
    code: 'if (a) {\n  doSomething();\n} else if (a) {\n  doOther();\n}\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-identical-expressions',
    code: 'const x = a < a;\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-identical-functions',
    code: 'function f() {\n  doA();\n  doB();\n  return 1 + 2;\n}\nfunction g() {\n  doA();\n  doB();\n  return 1 + 2;\n}\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-ignored-exceptions',
    code: 'try {\n  doSomething();\n} catch (e) {}\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-implicit-dependencies',
    code: "import x from 'some-unlisted-package';\nconsole.log(x);\n",
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-implicit-global',
    code: 'x = 5;\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-incomplete-assertions',
    code: 'expect(foo);\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-internal-api-use',
    code: "const x = require('foo/node_modules/bar');\n",
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-intrusive-permissions',
    code: 'navigator.geolocation.getCurrentPosition(success);\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-invalid-regexp',
    code: "const re = new RegExp('[');\n",
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-inverted-boolean-check',
    code: 'if (!(a === b)) {\n  doSomething();\n}\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-ip-forward',
    code: "import httpProxy from 'http-proxy';\nhttpProxy.createProxyServer({ xfwd: true });\n",
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-labels',
    code: 'loop: for (let i = 0; i < 10; i++) {\n  break loop;\n}\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-literal-call',
    code: 'const x = 5();\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-mime-sniff',
    code: "import express from 'express';\nimport helmet from 'helmet';\nconst app = express();\napp.use(helmet({ noSniff: false }));\n",
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-mixed-content',
    code: "import express from 'express';\nimport helmet from 'helmet';\nconst app = express();\napp.use(helmet.contentSecurityPolicy({ directives: { defaultSrc: [\"'self'\"] } }));\n",
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-nested-assignment',
    code: 'let a;\ndoSomething(a = 5);\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-nested-conditional',
    code: 'const x = a ? b : c ? d : e;\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-nested-functions',
    code: 'function a() { function b() { function c() {} } }\n',
    options: [{ threshold: 2 }],
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-nested-incdec',
    code: 'let i = 0;\nconst x = - i++;\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-nested-switch',
    code: 'switch (a) {\n  case 1:\n    switch (b) {\n      case 2:\n        break;\n    }\n    break;\n}\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-nested-template-literals',
    code: 'const x = `a ${`b ${c} d`} e`;\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-os-command-from-path',
    code: "import { exec } from 'child_process';\nexec('mycommand');\n",
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-primitive-wrappers',
    code: 'const n = new Number(1);\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-redundant-boolean',
    code: 'const x = a == true;\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-redundant-jump',
    code: 'function f() {\n  doSomething();\n  return;\n}\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-redundant-parentheses',
    code: 'const x = ((1));\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-reference-error',
    code: 'foo();\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-referrer-policy',
    code: "import express from 'express';\nimport helmet from 'helmet';\nconst app = express();\napp.use(helmet({ referrerPolicy: false }));\n",
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-regex-spaces',
    code: 'const re = /a  b/;\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-require-or-define',
    code: "require('foo');\n",
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-same-argument-assert',
    code: "import { assert } from 'chai';\nassert.equal(foo, foo);\n",
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-same-line-conditional',
    code: 'if (a) {\n} if (b) {\n}\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-session-cookies-on-static-assets',
    code: "import express from 'express';\nimport session from 'express-session';\nconst app = express();\napp.use(session({ secret: 's' }));\napp.use(express.static('public'));\n",
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-small-switch',
    code: 'switch (a) {\n  case 1:\n    break;\n  default:\n    break;\n}\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-sonar-comments',
    code: 'const x = 1; // NOSONAR\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-undefined-assignment',
    code: 'let x = undefined;\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-unenclosed-multiline-block',
    code: 'if (cond)\n  doA();\n  doB();\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-unsafe-unzip',
    code: "import tar from 'tar';\ntar.x({ file: 'a.tar' });\n",
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-unthrown-error',
    code: "function f() {\n  new Error('boom');\n}\n",
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-unused-collection',
    code: 'function f() {\n  const arr = [];\n  arr.push(1);\n}\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-unused-function-argument',
    code: 'function f(a, b) {\n  return a;\n}\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-unused-vars',
    code: 'function f() {\n  const x = 1;\n}\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-use-of-empty-return-value',
    code: 'function f() {\n  doSomething();\n}\nconst x = f();\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-useless-catch',
    code: 'function f() {\n  try {\n    doSomething();\n  } catch (e) {\n    throw e;\n  }\n}\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-useless-increment',
    code: 'function f() {\n  let i = 0;\n  return i++;\n}\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-useless-react-setstate',
    code: "import { useState } from 'react';\nfunction f() {\n  const [count, setCount] = useState(0);\n  setCount(count);\n}\n",
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-variable-usage-before-declaration',
    code: 'function f() {\n  x = 1;\n  var x;\n}\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-vue-bypass-sanitization',
    code: "createElement('a', { attrs: { href: 'x' } });\n",
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-weak-cipher',
    code: "import crypto from 'crypto';\ncrypto.createCipheriv('des', 'key', 'iv');\n",
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-weak-keys',
    code: "import crypto from 'crypto';\ncrypto.generateKeyPairSync('rsa', { modulusLength: 1024 });\n",
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-wildcard-import',
    code: "import * as foo from 'foo';\n",
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'non-existent-operator',
    code: 'let x = 1;\nx =- 5;\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'os-command',
    code: "import { exec } from 'child_process';\nexec(cmd);\n",
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'prefer-default-last',
    code: 'switch (x) {\n  default:\n    foo();\n    break;\n  case 1:\n    bar();\n    break;\n}\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'prefer-immediate-return',
    code: 'function f() {\n  const x = 1 + 2;\n  return x;\n}\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'prefer-object-literal',
    code: 'const obj = {};\nobj.a = 1;\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'prefer-promise-shorthand',
    code: 'const p = new Promise(function (resolve) {\n  resolve(42);\n});\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'prefer-single-boolean-return',
    code: 'function f(x: number) {\n  if (x > 0) {\n    return true;\n  } else {\n    return false;\n  }\n}\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'prefer-type-guard',
    code: 'function isFoo(x: unknown) {\n  return (x as { type: string }).type !== undefined;\n}\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'prefer-while',
    code: 'for (; cond; ) {\n  doSomething();\n}\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'process-argv',
    code: 'const a = process.argv;\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'production-debug',
    code: "import errorhandler from 'errorhandler';\napp.use(errorhandler());\n",
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'pseudo-random',
    code: 'const r = Math.random();\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'public-static-readonly',
    code: 'class C {\n  static foo = 1;\n}\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'publicly-writable-directories',
    code: "const p = '/tmp/data';\n",
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'reduce-initial-value',
    code: 'const s = [1, 2, 3].reduce((a, b) => a + b);\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'redundant-type-aliases',
    code: 'type T = string;\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'regex-complexity',
    code: 'const re = /(a|b|c|d|e)(a|b|c|d|e)(a|b|c|d|e)(a|b|c|d|e)(a|b|c|d|e)(a|b|c|d|e)(a|b|c|d|e)+/;\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'regular-expr',
    code: "const re = new RegExp('(a+)+$');\n",
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'review-blockchain-mnemonic',
    code: "import { Wallet } from 'ethers';\nconst w = Wallet.fromMnemonic('test test test test test test test test test test test junk');\n",
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'session-regeneration',
    code: "import passport from 'passport';\npassport.authenticate('local')(req, res, function (err, user) {\n  res.redirect('/');\n});\n",
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'shorthand-property-grouping',
    code: 'const o = { a, x: 1, b };\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'single-char-in-character-classes',
    code: 'const re = /[a]/;\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'single-character-alternation',
    code: 'const re = /a|b|c/;\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'slow-regex',
    code: 'const re = /(a+)+$/;\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'sockets',
    code: "import net from 'net';\nconst s = net.createConnection({ port: 80 });\n",
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'sql-queries',
    code: "import mysql from 'mysql';\nconst con = mysql.createConnection({});\ncon.query(`SELECT * FROM t WHERE id = ${userId}`);\n",
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'standard-input',
    code: 'const s = process.stdin;\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'stateful-regex',
    code: 'const re = /foo/gy;\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'strict-transport-security',
    code: "import express from 'express';\nimport helmet from 'helmet';\nconst app = express();\napp.use(helmet.hsts({ maxAge: 100 }));\n",
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'test-check-exception',
    code: "it('throws', function (done) {\n  try {\n    foo();\n  } catch (e) {\n    done();\n  }\n});\n",
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'todo-tag',
    code: '// TODO fix this later\nconst x = 1;\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'too-many-break-or-continue-in-loop',
    code: 'for (let i = 0; i < 10; i++) {\n  if (i === 1) break;\n  if (i === 2) break;\n}\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'unicode-aware-regex',
    code: 'const re = /\\p{L}/;\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'unused-import',
    code: "import { foo } from './mod';\nconst x = 1;\n",
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'unverified-certificate',
    code: "import https from 'https';\nhttps.request({ rejectUnauthorized: false });\n",
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'unverified-hostname',
    code: "import https from 'https';\nhttps.request({ rejectUnauthorized: false });\n",
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'updated-const-var',
    code: 'const x = 1;\nx = 2;\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'updated-loop-counter',
    code: 'for (let i = 0; i < 10; i++) {\n  i = i + 2;\n}\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'use-type-alias',
    code: 'let a: string | number | boolean;\nlet b: string | number | boolean;\nlet c: string | number | boolean;\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'variable-name',
    code: 'const My_Bad_Name = 1;\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'void-use',
    code: 'const x = 1;\nvoid x;\n',
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'weak-ssl',
    code: "import https from 'https';\nhttps.request({ secureProtocol: 'TLSv1_method' });\n",
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'x-powered-by',
    code: "import express from 'express';\nconst app = express();\napp.get('/', (req, res) => res.send('hi'));\n",
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'xml-parser-xxe',
    code: "import libxml from 'libxmljs';\nlibxml.parseXmlString(xml, { noent: true });\n",
  },
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'xpath',
    code: "const r = document.evaluate('//a', document, null, 0, null);\n",
  },
];

const CLEAN_CASES: DiffCase[] = [
  {
    pkg: 'eslint-plugin-sonarjs',
    rule: 'no-identical-expressions',
    code: 'const d = a === b;\n',
  },
];

runConformanceSuite('eslint-plugin-sonarjs', CASES, CLEAN_CASES);

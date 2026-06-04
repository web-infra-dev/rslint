/**
 * Conformance: @stylistic/eslint-plugin rules mounted in rslint via `plugins` must
 * report identically to ESLint v10. Shared assertion + excluded-category notes
 * live in ./conformance.ts.
 */
import { runConformanceSuite } from './conformance.js';
import type { DiffCase } from './harness.js';

/** 87 rules that report IDENTICALLY on a minimal trigger. */
const CASES: DiffCase[] = [
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'array-bracket-newline',
    code: 'const a = [\n1, 2];\n',
    options: ['never'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'array-bracket-spacing',
    code: 'const a = [ 1, 2];\n',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'array-element-newline',
    code: 'const a = [1, 2];\n',
    options: ['always'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'arrow-parens',
    code: 'const f = x => x;\n',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'arrow-spacing',
    code: 'const f = ()=> 1;\n',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'block-spacing',
    code: 'function f() {return 1;}\n',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'brace-style',
    code: 'if (true) {\n  x();\n}\nelse {\n  y();\n}\n',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'comma-dangle',
    code: 'const a = [\n  1,\n  2,\n];\n',
    options: ['never'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'comma-spacing',
    code: 'const a = [1 ,2];\n',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'comma-style',
    code: 'const a = [1\n, 2];\n',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'computed-property-spacing',
    code: "const o = {}; o[ 'a'];\n",
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'curly-newline',
    code: 'if (true) {x();}\n',
    options: ['always'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'dot-location',
    code: 'const a = b.\nc;\n',
    options: ['property'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'exp-jsx-props-style',
    code: 'const x = <div\n  a="1" b="2"\n/>;\n',
    options: [],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'exp-list-style',
    code: 'const a = [ 1, 2 ];\n',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'function-call-argument-newline',
    code: 'foo(1, 2);\n',
    options: ['always'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'function-call-spacing',
    code: 'foo ();\n',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'function-paren-newline',
    code: 'function f(\n  a, b\n) {}\n',
    options: ['never'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'generator-star-spacing',
    code: 'function *gen() {}\n',
    options: ['after'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'implicit-arrow-linebreak',
    code: 'const f = () =>\n  1;\n',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'indent-binary-ops',
    code: 'const x = 1 +\n2;\n',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-closing-tag-location',
    code: 'export {};\nconst x = <div>\n  text\n  </div>;\n',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-curly-brace-presence',
    code: "export {};\nconst x = <div>{'hello'}</div>;\n",
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-curly-newline',
    code: 'export {};\nconst x = <div>{\nfoo\n}</div>;\n',
    options: ['never'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-first-prop-new-line',
    code: 'export {};\nconst x = <div foo={1}\n  bar={2}\n/>;\n',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-function-call-newline',
    code: 'export {};\ndeclare function f(x: unknown): void;\nf(<div>\n  a\n</div>);\n',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-indent-props',
    code: 'export {};\nconst x = <div\nfoo={1}\n/>;\n',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-max-props-per-line',
    code: 'export {};\nconst x = <div foo={1} bar={2} />;\n',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-newline',
    code: 'export {};\nconst x = <div>\n  <span>a</span>\n  <span>b</span>\n</div>;\n',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-one-expression-per-line',
    code: 'export {};\nconst x = <div><span>a</span><span>b</span></div>;\n',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-pascal-case',
    code: 'export {};\nconst x = <Foo_Bar />;\n',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-quotes',
    code: "export {};\nconst x = <div className='a' />;\n",
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-self-closing-comp',
    code: 'export {};\nconst x = <Foo></Foo>;\n',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-wrap-multilines',
    code: 'export {};\nfunction f() {\n  return <div>\n    a\n  </div>;\n}\n',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'key-spacing',
    code: 'const o = { a : 1 };\n',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'keyword-spacing',
    code: 'if (true) {}else {}\n',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'line-comment-position',
    code: 'const x = 1; // trailing\n',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'linebreak-style',
    code: 'const a = 1;\r\nconst b = 2;\r\n',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'lines-around-comment',
    code: 'const a = 1;\n// comment\nconst b = 2;\n',
    options: [{ beforeLineComment: true }],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'lines-between-class-members',
    code: 'class A {\n  foo() {}\n  bar() {}\n}\n',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'max-len',
    code: 'const aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa = 1;\n',
    options: [{ code: 80 }],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'max-statements-per-line',
    code: 'const a = 1; const b = 2;\n',
    options: [{ max: 1 }],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'member-delimiter-style',
    code: 'interface Foo {\n  a: number,\n  b: string,\n}\n',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'multiline-comment-style',
    code: '// a\n// b\nconst x = 1;\n',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'multiline-ternary',
    code: 'const a = cond ? 1 : 2;\n',
    options: ['always'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'new-parens',
    code: 'const x = new Foo;\n',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'newline-per-chained-call',
    code: 'foo.bar().baz().qux().quux();\n',
    options: [{ ignoreChainWithDepth: 2 }],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'no-confusing-arrow',
    code: 'const x = (a) => a ? 1 : 2;\n',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'no-extra-parens',
    code: 'const a = (1);\n',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'no-extra-semi',
    code: 'const a = 1;;\n',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'no-floating-decimal',
    code: 'const a = .5;\n',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'no-mixed-operators',
    code: 'const a = 1 + 2 % 3;\n',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'no-mixed-spaces-and-tabs',
    code: 'if (a) {\n \tx = 1;\n}\n',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'no-multi-spaces',
    code: 'const a =  1;\n',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'no-multiple-empty-lines',
    code: 'const a = 1;\n\n\n\nconst b = 2;\n',
    options: [{ max: 1 }],
  },
  { pkg: '@stylistic/eslint-plugin', rule: 'no-tabs', code: 'const a =\t1;\n' },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'no-trailing-spaces',
    code: 'const a = 1; \n',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'no-whitespace-before-property',
    code: 'foo .bar;\n',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'nonblock-statement-body-position',
    code: 'if (a)\n  foo();\n',
    options: ['beside'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'object-curly-newline',
    code: 'const a = {b: 1, c: 2};\n',
    options: ['always'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'object-curly-spacing',
    code: 'const a = {b: 1};\n',
    options: ['always'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'object-property-newline',
    code: 'const a = {b: 1, c: 2};\n',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'one-var-declaration-per-line',
    code: 'var a, b;\n',
    options: ['always'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'operator-linebreak',
    code: 'const a = 1 +\n  2;\n',
    options: ['before'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'padded-blocks',
    code: 'function foo() {\n  bar();\n}\n',
    options: ['always'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'padding-line-between-statements',
    code: 'function f() {\n  const a = 1;\n  return a;\n}\n',
    options: [{ blankLine: 'always', prev: 'const', next: 'return' }],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'quote-props',
    code: 'const obj = { foo: 1 };\n',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'quotes',
    code: "const s = 'hello';\n",
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'rest-spread-spacing',
    code: 'const arr = [1, 2, 3];\nconst copy = [... arr];\n',
  },
  { pkg: '@stylistic/eslint-plugin', rule: 'semi', code: 'const a = 1\n' },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'semi-spacing',
    code: 'const a = 1 ;\n',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'semi-style',
    code: 'const a = 1\n;const b = 2;\n',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'space-before-blocks',
    code: 'if (true){\n  const a = 1;\n}\n',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'space-before-function-paren',
    code: 'function foo() {\n  return 1;\n}\n',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'space-in-parens',
    code: 'const a = ( 1 );\n',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'space-infix-ops',
    code: 'const a = 1+2;\n',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'space-unary-ops',
    code: 'const a = typeof(1);\n',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'spaced-comment',
    code: '//comment\nconst a = 1;\n',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'switch-colon-spacing',
    code: 'switch (1) {\n  case 1 :\n    break;\n}\n',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'template-curly-spacing',
    code: 'const x = 1;\nconst s = `a${ x }b`;\n',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'template-tag-spacing',
    code: 'function tag(s: TemplateStringsArray) { return s; }\nconst x = tag `hello`;\n',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'type-annotation-spacing',
    code: 'const a:number = 1;\n',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'type-generic-spacing',
    code: 'type A<T= number> = T;\n',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'type-named-tuple-spacing',
    code: 'type T = [name:string];\n',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'wrap-iife',
    code: 'const x = function () { return 1; }();\n',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'wrap-regex',
    code: "const f = function () { return /foo/.test('foo'); };\n",
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'yield-star-spacing',
    code: 'function* gen() {\n  yield *[1, 2];\n}\n',
  },
];

const CLEAN_CASES: DiffCase[] = [
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'comma-spacing',
    code: 'const c = [1, 2];\n',
  },
  { pkg: '@stylistic/eslint-plugin', rule: 'quotes', code: 'const b = "x";\n' },
  { pkg: '@stylistic/eslint-plugin', rule: 'semi', code: 'const a = 1;\n' },
];

runConformanceSuite('@stylistic/eslint-plugin', CASES, CLEAN_CASES);

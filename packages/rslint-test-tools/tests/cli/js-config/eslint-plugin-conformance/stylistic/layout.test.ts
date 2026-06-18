/**
 * Conformance: @stylistic/eslint-plugin (layout) mounted in rslint via `plugins`
 * must report identically to ESLint v10. Representative triggers from the
 * upstream suite; each verified to reproduce ESLint v10 byte-for-byte.
 */
import { runConformanceSuite } from '../conformance.js';
import type { DiffCase } from '../harness.js';

const CASES: DiffCase[] = [
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'array-bracket-newline',
    code: 'var foo = [\n                [1,2]\n            ]',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'array-bracket-newline',
    code: 'var foo = [];',
    options: ['always'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'array-bracket-newline',
    code: 'var foo = [\n];',
    options: ['never'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'array-bracket-newline',
    code: 'var foo = [1,[\n2,3\n]\n];',
    options: [{ minItems: 2 }],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'array-element-newline',
    code: 'var foo = [\n1,[2,\n3]]',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'array-element-newline',
    code: 'var foo = [1, 2];',
    options: ['always'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'array-element-newline',
    code: 'var foo = [\n1,\n2\n];',
    options: ['never'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'curly-newline',
    code: '{void {foo}\n}',
    options: [],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'curly-newline',
    code: '{void {foo}}',
    options: ['always'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'curly-newline',
    code: '{\n}',
    options: ['never'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'function-call-argument-newline',
    code: 'fn(a, b)',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'function-call-argument-newline',
    code: 'fn(a, b, c)',
    options: ['always'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'function-call-argument-newline',
    code: 'fn(a,\n\tb,\n\tc)',
    options: ['never'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'function-call-argument-newline',
    code: 'fn(a,\n\tb, c)',
    options: ['consistent'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'function-paren-newline',
    code: 'function baz(foo,\n    bar\n) {}',
    options: ['multiline'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'function-paren-newline',
    code: 'function baz(foo, bar) {}',
    options: ['always'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'function-paren-newline',
    code: 'baz(\n    foo, bar);',
    options: ['multiline'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'function-paren-newline',
    code: 'function baz(foo, bar, qux) {}',
    options: [{ minItems: 3 }],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'function-paren-newline',
    code: 'function baz(foo,\n    bar\n) {}',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'function-paren-newline',
    code: 'function baz(\n    foo,\n    bar\n) {}',
    options: ['never'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'implicit-arrow-linebreak',
    code: '(foo) =>\n    bar();',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'implicit-arrow-linebreak',
    code: '() =>\n    bar =>\n        baz;',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'implicit-arrow-linebreak',
    code: '(foo) => bar();',
    options: ['below'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'implicit-arrow-linebreak',
    code: '(foo) => (bar);',
    options: ['below'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'indent-binary-ops',
    code: 'const a =\n  x +\n    y * z',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'indent-binary-ops',
    code: 'if (\n  aaaaaa >\nbbbbb\n) {}',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'indent-binary-ops',
    code: 'const a = 1\n+ 2\n    + 3;',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'indent-binary-ops',
    code: 'type Foo = A | B\n  | C | D\n    | E',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'indent',
    code: 'var a = b;\nif (a) {\nb();\n}',
    options: [2],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'indent',
    code: 'if (a){\n\tb=c;\n\t\tc=d;\ne=f;\n}',
    options: ['tab'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'indent',
    code: 'if (a){\n    b=c;\n      c=d;\n e=f;\n}',
    options: [4],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'linebreak-style',
    code: "var a = 'a';\r\n",
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'linebreak-style',
    code: "var a = 'a';\n",
    options: ['windows'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'linebreak-style',
    code: "var a = 'a'; ",
    options: ['unix'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'linebreak-style',
    code: '\r\n',
    options: ['unix'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'lines-around-comment',
    code: 'bar()\n/** block block block\n * block \n */\nvar a = 1;',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'lines-around-comment',
    code: 'baz()\n// A line comment with no empty line after\nvar a = 1;',
    options: [{ afterLineComment: true }],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'lines-around-comment',
    code: 'baz()\n// A line comment with no empty line after\nvar a = 1;',
    options: [{ beforeLineComment: true, afterLineComment: true }],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'lines-between-class-members',
    code: 'class foo{ bar(){}\nbaz(){}}',
    options: ['always'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'lines-between-class-members',
    code: 'class foo{ bar(){}\n\nbaz(){}}',
    options: ['never'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'lines-between-class-members',
    code: 'class foo{ bar(){\n}\nbaz(){}}',
    options: ['always', { exceptAfterSingleLine: true }],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'lines-between-class-members',
    code: 'class foo{ bar(){} // comment \nbaz(){}}',
    options: ['always'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'multiline-ternary',
    code: 'a ? b : c',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'multiline-ternary',
    code: 'a\n? b : c',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'multiline-ternary',
    code: 'a ? b\n: c',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'newline-per-chained-call',
    code: '_\n.chain({}).map(foo).filter(bar).value();',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'newline-per-chained-call',
    code: 'a().b().c().e.d()',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'newline-per-chained-call',
    code: 'var a = m1().m2\n.m3().m4().m5().m6().m7();',
    options: [{ ignoreChainWithDepth: 3 }],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'newline-per-chained-call',
    code: 'obj?.foo1()?.foo2()?.foo3()',
    options: [{ ignoreChainWithDepth: 1 }],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'nonblock-statement-body-position',
    code: 'if (foo)\n    bar();',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'nonblock-statement-body-position',
    code: 'do\n    bar();\nwhile (foo)',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'nonblock-statement-body-position',
    code: 'if (foo) bar();',
    options: ['below'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'nonblock-statement-body-position',
    code: 'while (foo)\n    bar();',
    options: ['below', { overrides: { while: 'beside' } }],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'object-curly-newline',
    code: 'var a = { a\n};',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'object-curly-newline',
    code: 'var a = {};',
    options: ['always'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'object-curly-newline',
    code: 'var b = {a: 1};',
    options: ['always'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'object-curly-newline',
    code: 'var a = {\n};',
    options: ['never'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'object-property-newline',
    code: "var obj = { k1: 'val1', k2: 'val2', k3: 'val3' };",
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'object-property-newline',
    code: "var obj = { k1: 'val1', k2: \n'val2', \nk3: 'val3' };",
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'object-property-newline',
    code: "var obj = {\nk1: 'val1', k2: 'val2'\n};",
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'one-var-declaration-per-line',
    code: 'var a, b;',
    options: ['always'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'one-var-declaration-per-line',
    code: 'const a = 0, b = 0;',
    options: ['always'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'one-var-declaration-per-line',
    code: 'var a, b, c = 0;',
    options: ['initializations'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'one-var-declaration-per-line',
    code: 'export let a, b;',
    options: ['always'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'operator-linebreak',
    code: '1\n+ 1',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'operator-linebreak',
    code: '1\n+\n1',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'operator-linebreak',
    code: '1 \n || 1',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'operator-linebreak',
    code: 'a\n += 1',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'padded-blocks',
    code: '{\n//comment\na();\n\n}',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'padded-blocks',
    code: '{\n\na();\n//comment\n}',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'padded-blocks',
    code: '{\na();\n\n}',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'padded-blocks',
    code: '{\n\na();\n}',
  },
];

const CLEAN_CASES: DiffCase[] = [
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'array-bracket-newline',
    code: 'var foo = [\n1,\n2\n];',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'array-bracket-newline',
    code: 'var foo = [\n1, 2\n];',
    options: ['always'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'array-element-newline',
    code: 'var foo = [1,\n2];',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'array-element-newline',
    code: 'var foo = [1,\n2,\n3];',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'curly-newline',
    code: '{\nvoid {foo}\n}',
    options: ['always'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'curly-newline',
    code: '{void {foo}}',
    options: ['never'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'function-call-argument-newline',
    code: 'fn(a, b)',
    options: ['never'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'function-call-argument-newline',
    code: 'fn(a,\n\tb,\n\tc)',
    options: ['consistent'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'function-paren-newline',
    code: 'function baz(foo, bar) {}',
    options: ['multiline'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'function-paren-newline',
    code: 'baz(\n    foo,\n    bar\n);',
    options: ['always'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'implicit-arrow-linebreak',
    code: '() => bar;',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'implicit-arrow-linebreak',
    code: '() =>\n    (bar);',
    options: ['below'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'indent-binary-ops',
    code: 'const a = 1\n  + 2\n  + 3;',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'indent-binary-ops',
    code: 'type Foo =\n  | A\n  | B',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'indent',
    code: "var x = [\n    'a',\n    'b',\n    'c'\n];",
    options: [4],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'indent',
    code: 'var x = 0 && 1;',
    options: [4],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'linebreak-style',
    code: "var a = 'a',\n b = 'b';\n\n function foo(params) {\n /* do stuff */ \n }\n",
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'linebreak-style',
    code: "var a = 'a',\r\n b = 'b';\r\n\r\n function foo(params) {\r\n /* do stuff */ \r\n }\r\n",
    options: ['windows'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'lines-around-comment',
    code: 'bar()\n\n/** block block block\n * block \n */\nvar a = 1;',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'lines-around-comment',
    code: 'foo()\n\n// line line line\nvar a = 1;',
    options: [{ beforeLineComment: true }],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'lines-between-class-members',
    code: 'class foo{ bar(){}\n\nbaz(){}}',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'lines-between-class-members',
    code: 'class foo{ bar(){}\nbaz(){}}',
    options: ['never'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'multiline-ternary',
    code: 'a\n? b\n: c',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'multiline-ternary',
    code: 'a ? b : c',
    options: ['never'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'newline-per-chained-call',
    code: '_\n.chain({})\n.map(foo)\n.filter(bar)\n.value();',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'newline-per-chained-call',
    code: 'a.b.c.d.e.f',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'nonblock-statement-body-position',
    code: 'if (foo) bar;',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'nonblock-statement-body-position',
    code: 'if (foo)\n    bar();',
    options: ['below'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'object-curly-newline',
    code: 'var a = { foo }',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'object-curly-newline',
    code: 'var d = {\n    a: 1,\n    b: 2\n};',
    options: ['always'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'object-property-newline',
    code: "var obj = {\nk1: 'val1',\nk2: 'val2',\nk3: 'val3',\nk4: 'val4'\n};",
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'object-property-newline',
    code: "var obj = { k1: 'val1', k2: 'val2', k3: 'val3' };",
    options: [{ allowAllPropertiesOnSameLine: true }],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'one-var-declaration-per-line',
    code: 'var a,\nb,\nc,\nd = 0;',
    options: ['always'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'one-var-declaration-per-line',
    code: 'var a, b, c,\nd = 0;',
    options: ['initializations'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'operator-linebreak',
    code: '1 +\n1',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'operator-linebreak',
    code: 'answer = everything ?\n  42 :\n  foo;',
    options: ['after'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'padded-blocks',
    code: '{\n\na();\n\n}',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'padded-blocks',
    code: 'switch (a) {\n\ncase 0: foo();\ncase 1: bar();\n\n}',
    options: ['always'],
  },
];

runConformanceSuite('@stylistic/eslint-plugin', CASES, CLEAN_CASES);

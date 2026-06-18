/**
 * Conformance: @stylistic/eslint-plugin (syntax) mounted in rslint via `plugins`
 * must report identically to ESLint v10. Representative triggers from the
 * upstream suite; each verified to reproduce ESLint v10 byte-for-byte.
 */
import { runConformanceSuite } from '../conformance.js';
import type { DiffCase } from '../harness.js';

const CASES: DiffCase[] = [
  { pkg: '@stylistic/eslint-plugin', rule: 'arrow-parens', code: 'a => {}' },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'arrow-parens',
    code: 'a.then(foo => {});',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'arrow-parens',
    code: '(a) => a',
    options: ['as-needed'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'brace-style',
    code: 'function foo() { return; }',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'brace-style',
    code: 'function foo() \n { \n return; }',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'brace-style',
    code: 'if (f) {\n  bar;\n}\nelse\n  baz;',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'comma-dangle',
    code: "var foo = { bar: 'baz', }",
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'comma-dangle',
    code: "foo({ bar: 'baz', qux: 'quux', });",
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'comma-dangle',
    code: "var foo = { bar: 'baz' }",
    options: ['always'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'comma-style',
    code: 'var foo = 1\n,bar = 2;',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'comma-style',
    code: 'var foo = 1\n,\nbar = 2;',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'comma-style',
    code: 'var foo = 1,\nbar = 2;',
    options: ['first'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'dot-location',
    code: 'obj\n.property',
    options: ['object'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'dot-location',
    code: 'obj.\nproperty',
    options: ['property'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'dot-location',
    code: '(obj).\nproperty',
    options: ['property'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'new-parens',
    code: 'var a = new Date;',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'new-parens',
    code: 'var a = (new Foo).bar;',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'new-parens',
    code: 'var a = new new Foo()',
    options: ['always'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'new-parens',
    code: 'var a = new Date();',
    options: ['never'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'no-confusing-arrow',
    code: 'a => 1 ? 2 : 3',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'no-confusing-arrow',
    code: 'a => 1 ? 2 : 3',
    options: [{ allowParens: false }],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'no-confusing-arrow',
    code: 'var x = (a, b) => 1 ? 2 : 3',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'no-confusing-arrow',
    code: 'var x = () => 1 ? 2 : 3',
    options: [{ onlyOneSimpleParam: false }],
  },
  { pkg: '@stylistic/eslint-plugin', rule: 'no-extra-parens', code: '(0)' },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'no-extra-parens',
    code: 'if((0));',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'no-extra-parens',
    code: 'throw(0)',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'no-extra-parens',
    code: 'switch(0){ case (1): break; }',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'no-extra-semi',
    code: 'var x = 5;;',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'no-extra-semi',
    code: 'function foo(){};',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'no-extra-semi',
    code: 'class A { a() {}; b() {} }',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'no-extra-semi',
    code: 'class Foo {\n  public foo: number = 0;;\n}',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'no-floating-decimal',
    code: 'var x = .5;',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'no-floating-decimal',
    code: 'var x = -.5;',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'no-floating-decimal',
    code: 'var x = 2.;',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'no-floating-decimal',
    code: 'var x = -2.;',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'no-mixed-operators',
    code: 'a && b || c',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'no-mixed-operators',
    code: 'a + b - c',
    options: [{ allowSamePrecedence: false }],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'no-mixed-operators',
    code: 'a || b ? c : d',
    options: [{ groups: [['&&', '||', '?:']] }],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'no-mixed-operators',
    code: 'a && b + c - d / e || f',
    options: [
      {
        groups: [
          ['&&', '||'],
          ['+', '-', '*', '/'],
        ],
      },
    ],
  },
  { pkg: '@stylistic/eslint-plugin', rule: 'quote-props', code: '({ a: 0 })' },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'quote-props',
    code: "({ 'a': 0 })",
    options: ['as-needed'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'quote-props',
    code: "({ '-a': 0, b: 0 })",
    options: ['consistent'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'quote-props',
    code: "({ a: 0, 'b': 0 })",
    options: ['consistent'],
  },
  { pkg: '@stylistic/eslint-plugin', rule: 'quotes', code: "var foo = 'bar';" },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'quotes',
    code: 'var foo = "bar";',
    options: ['single'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'quotes',
    code: 'var foo = `bar`;',
    options: ['single'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'quotes',
    code: "var foo = 'bar';",
    options: ['double'],
  },
  { pkg: '@stylistic/eslint-plugin', rule: 'semi-style', code: 'foo\n;bar' },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'semi-style',
    code: 'if(a)foo\n;bar',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'semi-style',
    code: 'var foo\n;bar',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'semi-style',
    code: 'for(a\n;b;c)d',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'semi',
    code: 'function foo() { return [] }',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'semi',
    code: 'while(true) { break }',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'wrap-iife',
    code: 'var a = function(){ }();',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'wrap-iife',
    code: '(function a(){ })();',
    options: ['outside'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'wrap-iife',
    code: '(function a(){ }());',
    options: ['inside'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'wrap-iife',
    code: 'if (function (){}()) {}',
    options: ['inside'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'wrap-regex',
    code: '/foo/.test(bar);',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'wrap-regex',
    code: '/foo/ig.test(bar);',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'wrap-regex',
    code: 'if(/foo/ig.test(bar));',
  },
];

const CLEAN_CASES: DiffCase[] = [
  { pkg: '@stylistic/eslint-plugin', rule: 'arrow-parens', code: '(a) => {}' },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'arrow-parens',
    code: 'a => a',
    options: ['as-needed'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'brace-style',
    code: 'if (foo) {\n}\nelse {\n}',
    options: ['stroustrup'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'brace-style',
    code: 'if (foo)\n{\n}\nelse\n{\n}',
    options: ['allman'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'comma-dangle',
    code: "var foo = { bar: 'baz' }",
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'comma-dangle',
    code: "var foo = { bar: 'baz', }",
    options: ['always'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'comma-style',
    code: 'var foo = 1, bar = 3;',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'comma-style',
    code: "var foo = ['apples'\n,'oranges'];",
    options: ['first'],
  },
  { pkg: '@stylistic/eslint-plugin', rule: 'dot-location', code: 'obj.prop' },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'dot-location',
    code: 'obj\n.prop',
    options: ['property'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'new-parens',
    code: 'var a = new Date();',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'new-parens',
    code: 'var a = new Date;',
    options: ['never'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'no-confusing-arrow',
    code: 'a => { return 1 ? 2 : 3; }',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'no-confusing-arrow',
    code: '(a, b) => 1 ? 2 : 3',
    options: [{ onlyOneSimpleParam: true }],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'no-extra-parens',
    code: 'a = (b, c)',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'no-extra-parens',
    code: '(a || b) && (c || d)',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'no-extra-semi',
    code: 'var x = 5;',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'no-extra-semi',
    code: 'class A { static { foo; } }',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'no-floating-decimal',
    code: 'var x = 2.5;',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'no-floating-decimal',
    code: 'var x = "2.5";',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'no-mixed-operators',
    code: 'a && b && c && d',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'no-mixed-operators',
    code: '(a || b) && c && d',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'quote-props',
    code: '({ "a": 0 })',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'quote-props',
    code: '({ a: 0, b(){} })',
    options: ['as-needed'],
  },
  { pkg: '@stylistic/eslint-plugin', rule: 'quotes', code: 'var foo = "bar";' },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'quotes',
    code: "var foo = 'bar';",
    options: ['single'],
  },
  { pkg: '@stylistic/eslint-plugin', rule: 'semi-style', code: 'foo;\nbar;' },
  { pkg: '@stylistic/eslint-plugin', rule: 'semi-style', code: 'for(a;b;c);' },
  { pkg: '@stylistic/eslint-plugin', rule: 'semi', code: 'var x = 5;' },
  { pkg: '@stylistic/eslint-plugin', rule: 'semi', code: 'foo();' },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'wrap-iife',
    code: '(function(){ }());',
    options: ['any'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'wrap-iife',
    code: '(function a(){ })();',
    options: ['inside'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'wrap-regex',
    code: '(/foo/).test(bar);',
  },
  { pkg: '@stylistic/eslint-plugin', rule: 'wrap-regex', code: 'a[/b/];' },
];

runConformanceSuite('@stylistic/eslint-plugin', CASES, CLEAN_CASES);

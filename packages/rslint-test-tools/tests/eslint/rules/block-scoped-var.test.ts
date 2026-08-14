import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('block-scoped-var', {
  valid: [
    // https://github.com/eslint/eslint/issues/2242
    'function f() { } f(); var exports = { f: f };',
    'var f = () => {}; f(); var exports = { f: f };',

    '!function f(){ f; }',
    'function f() { } f(); var exports = { f: f };',
    'function f() { var a, b; { a = true; } b = a; }',
    'var a; function f() { var b = a; }',
    'function f(a) { }',
    '!function(a) { };',
    '!function f(a) { };',
    'function f(a) { var b = a; }',
    '!function f(a) { var b = a; };',
    'function f() { var g = f; }',
    'function f() { } function g() { var f = g; }',
    'function f() { var hasOwnProperty; { hasOwnProperty; } }',
    'function f(){ a; b; var a, b; }',
    'function f(){ g(); function g(){} }',
    'if (true) { var a = 1; a; }',
    'var a; if (true) { a; }',
    'for (var i = 0; i < 10; i++) { i; }',
    'var i; for(i; i; i) { i; }',
    'function myFunc(foo) {  "use strict";  var { bar } = foo;  bar.hello();}',
    'function myFunc(foo) {  "use strict";  var [ bar ]  = foo;  bar.hello();}',
    'function myFunc(...foo) {  return foo;}',
    'var f = () => { var g = f; }',
    'class Foo {}\nexport default Foo;',
    'new Date',
    "var eslint = require('eslint');",
    'var fun = function({x}) {return x;};',
    'var fun = function([,x]) {return x;};',
    'function f(a) { return a.b; }',
    'var a = { "foo": 3 };',
    'var a = { foo: 3 };',
    'var a = { foo: 3, bar: 5 };',
    'var a = { set foo(a){}, get bar(){} };',
    'function f(a) { return arguments[0]; }',
    'function f() { }; var a = f;',
    'var a = f; function f() { };',
    'function f(){ for(var i; i; i) i; }',
    'function f(){ for(var a=0, b=1; a; b) a, b; }',
    'function f(){ for(var a in {}) a; }',
    'function f(){ switch(2) { case 1: var b = 2; b; break; default: b; break;} }',
    'a:;',
    'foo: while (true) { bar: for (var i = 0; i < 13; ++i) {if (i === 7) break foo; } }',
    'foo: while (true) { bar: for (var i = 0; i < 13; ++i) {if (i === 7) continue foo; } }',
    'const React = require("react/addons");const cx = React.addons.classSet;',
    'var v = 1;  function x() { return v; };',
    'import * as y from "./other.js"; y();',
    'import y from "./other.js"; y();',
    'import {x as y} from "./other.js"; y();',
    'var x; export {x};',
    'var x; export {x as v};',
    'export {x} from "./other.js";',
    'export {x as v} from "./other.js";',
    'class Test { myFunction() { return true; }}',
    'class Test { get flag() { return true; }}',
    'var Test = class { myFunction() { return true; }}',
    'var doStuff; let {x: y} = {x: 1}; doStuff(y);',
    'function foo({x: y}) { return y; }',

    // same coverage as no-undef
    '!function f(){}; f',
    'var f = function foo() { }; foo(); var exports = { f: foo };',
    'var f = () => { x; }',
    'function f(){ x; }',
    "var eslint = require('eslint');",
    'function f(a) { return a[b]; }',
    'function f() { return b.a; }',
    'var a = { foo: bar };',
    'var a = { foo: foo };',
    'var a = { bar: 7, foo: bar };',
    'var a = arguments;',
    'function x(){}; var a = arguments;',
    'function z(b){}; var a = b;',
    'function z(){var b;}; var a = b;',
    'function f(){ try{}catch(e){} e }',
    'a:b;',

    // https://github.com/eslint/eslint/issues/2253
    '/*global React*/ let {PropTypes, addons: {PureRenderMixin}} = React; let Test = React.createClass({mixins: [PureRenderMixin]});',
    '/*global prevState*/ const { virtualSize: prevVirtualSize = 0 } = prevState;',
    'const { dummy: { data, isLoading }, auth: { isLoggedIn } } = this.props;',

    // https://github.com/eslint/eslint/issues/2747
    'function a(n) { return n > 0 ? b(n - 1) : "a"; } function b(n) { return n > 0 ? a(n - 1) : "b"; }',

    // https://github.com/eslint/eslint/issues/2967
    '(function () { foo(); })(); function foo() {}',

    // class static blocks
    'class C { static { var foo; foo; } }',
    'class C { static { foo; var foo; } }',
    'class C { static { if (bar) { foo; } var foo; } }',
    'var foo; class C { static { foo; } } ',
    'class C { static { foo; } } var foo;',
    'var foo; class C { static {} [foo]; } ',
    'foo; class C { static {} } var foo; ',
  ],
  invalid: [
    {
      code: 'function f(){ x; { var x; } }',
      errors: [{ messageId: 'outOfScope', line: 1, column: 15 }],
    },
    {
      code: 'function f(){ { var x; } x; }',
      errors: [{ messageId: 'outOfScope', line: 1, column: 26 }],
    },
    {
      code: 'function f() { var a; { var b = 0; } a = b; }',
      errors: [{ messageId: 'outOfScope', line: 1, column: 42 }],
    },
    {
      code: 'function f() { try { var a = 0; } catch (e) { var b = a; } }',
      errors: [{ messageId: 'outOfScope', line: 1, column: 55 }],
    },
    {
      code: 'function a() { for(var b in {}) { var c = b; } c; }',
      errors: [{ messageId: 'outOfScope', line: 1, column: 48 }],
    },
    {
      code: 'function a() { for(var b of {}) { var c = b; } c; }',
      errors: [{ messageId: 'outOfScope', line: 1, column: 48 }],
    },
    {
      code: 'function f(){ switch(2) { case 1: var b = 2; b; break; default: b; break;} b; }',
      errors: [{ messageId: 'outOfScope', line: 1, column: 76 }],
    },
    {
      code: 'for (var a = 0;;) {} a;',
      errors: [{ messageId: 'outOfScope', line: 1, column: 22 }],
    },
    {
      code: 'for (var a in []) {} a;',
      errors: [{ messageId: 'outOfScope', line: 1, column: 22 }],
    },
    {
      code: 'for (var a of []) {} a;',
      errors: [{ messageId: 'outOfScope', line: 1, column: 22 }],
    },
    {
      code: '{ var a = 0; } a;',
      errors: [{ messageId: 'outOfScope', line: 1, column: 16 }],
    },
    {
      code: 'if (true) { var a; } a;',
      errors: [{ messageId: 'outOfScope', line: 1, column: 22 }],
    },
    {
      code: 'if (true) { var a = 1; } else { var a = 2; }',
      errors: [
        { messageId: 'outOfScope', line: 1, column: 17 },
        { messageId: 'outOfScope', line: 1, column: 37 },
      ],
    },
    {
      code: 'for (var i = 0;;) {} for(var i = 0;;) {}',
      errors: [
        { messageId: 'outOfScope', line: 1, column: 10 },
        { messageId: 'outOfScope', line: 1, column: 30 },
      ],
    },
    {
      code: 'class C { static { if (bar) { var foo; } foo; } }',
      errors: [{ messageId: 'outOfScope', line: 1, column: 42 }],
    },
    {
      code: '{ var foo,\n  bar; } bar;',
      errors: [{ messageId: 'outOfScope', line: 2, column: 10 }],
    },
    {
      code: '{ var { foo,\n  bar } = baz; } bar;',
      errors: [{ messageId: 'outOfScope', line: 2, column: 18 }],
    },
    {
      code: 'if (foo) { var a = 1; } else if (bar) { var a = 2; } else { var a = 3; }',
      errors: [
        { messageId: 'outOfScope', line: 1, column: 16 },
        { messageId: 'outOfScope', line: 1, column: 16 },
        { messageId: 'outOfScope', line: 1, column: 45 },
        { messageId: 'outOfScope', line: 1, column: 45 },
        { messageId: 'outOfScope', line: 1, column: 65 },
        { messageId: 'outOfScope', line: 1, column: 65 },
      ],
    },
  ],
});

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-redeclare', {
  valid: [
    'var a = 3; var b = function() { var a = 10; };',
    'var a = 3; a = 10;',
    'if (true) {\n    let b = 2;\n} else {    \nlet b = 3;\n}',
    'var a; class C { static { var a; } }',
    'class C { static { var a; } } var a; ',
    'function a(){} class C { static { var a; } }',
    'var a; class C { static { function a(){} } }',
    'class C { static { var a; } static { var a; } }',
    'class C { static { function a(){} } static { function a(){} } }',
    'class C { static { var a; { function a(){} } } }',
    'class C { static { function a(){}; { function a(){} } } }',
    'class C { static { var a; { let a; } } }',
    'class C { static { let a; { let a; } } }',
    'class C { static { { let a; } { let a; } } }',
    { code: 'var Object = 0;', options: { builtinGlobals: false } },
    { code: 'var top = 0;', options: { builtinGlobals: true } },
    {
      code: 'var top = 0;',
      options: { builtinGlobals: true },
      // SKIP: requires globalReturn plus configured-global source metadata.
      skip: true,
    },
    { code: 'var self = 1', options: { builtinGlobals: true } },
    {
      code: '/*globals Array */',
      options: { builtinGlobals: false },
    },
    {
      code: '/*globals a:off */',
      options: { builtinGlobals: true },
    },
    {
      code: '/*globals configured:off */ var configured = 1;',
      options: { builtinGlobals: true },
      languageOptions: { globals: { configured: 'readonly' } },
    },
    {
      code: '/*globals configured */',
      options: { builtinGlobals: true },
      languageOptions: { globals: { configured: 'off' } },
    },
  ],
  invalid: [
    {
      code: 'var a = 3;\nvar a = 10;',
      errors: [{ messageId: 'redeclared', line: 2, column: 5 }],
    },
    {
      code: 'switch(foo) { case a: var b = 3;\ncase b: var b = 4}',
      errors: [{ messageId: 'redeclared', line: 2, column: 13 }],
    },
    {
      code: 'var a = 3; var a = 10;',
      errors: [{ messageId: 'redeclared', line: 1, column: 16 }],
    },
    {
      code: 'var a = {}; var a = [];',
      errors: [{ messageId: 'redeclared', line: 1, column: 17 }],
    },
    {
      code: 'var a; function a() {}',
      errors: [{ messageId: 'redeclared', line: 1, column: 17 }],
    },
    {
      code: 'function a() {} function a() {}',
      errors: [{ messageId: 'redeclared', line: 1, column: 26 }],
    },
    {
      code: 'var a = function() { }; var a = function() { }',
      errors: [{ messageId: 'redeclared', line: 1, column: 29 }],
    },
    {
      code: 'var a = function() { }; var a = new Date();',
      errors: [{ messageId: 'redeclared', line: 1, column: 29 }],
    },
    {
      code: 'var a = 3; var a = 10; var a = 15;',
      errors: [
        { messageId: 'redeclared', line: 1, column: 16 },
        { messageId: 'redeclared', line: 1, column: 28 },
      ],
    },
    {
      code: 'var a; var a;',
      errors: [{ messageId: 'redeclared', line: 1, column: 12 }],
    },
    {
      code: 'export var a; var a;',
      errors: [{ messageId: 'redeclared', line: 1, column: 19 }],
    },
    {
      code: 'class C { static { var a; var a; } }',
      errors: [{ messageId: 'redeclared', line: 1, column: 31 }],
    },
    {
      code: 'class C { static { var a; { var a; } } }',
      errors: [{ messageId: 'redeclared', line: 1, column: 33 }],
    },
    {
      code: 'class C { static { { var a; } var a; } }',
      errors: [{ messageId: 'redeclared', line: 1, column: 35 }],
    },
    {
      code: 'class C { static { { var a; } { var a; } } }',
      errors: [{ messageId: 'redeclared', line: 1, column: 37 }],
    },
    {
      code: 'var Object = 0;',
      errors: [{ messageId: 'redeclaredAsBuiltin', line: 1, column: 5 }],
    },
    {
      code: 'var a;\nvar {a = 0, b: Object = 0} = {};',
      options: { builtinGlobals: true },
      errors: [
        { messageId: 'redeclared', line: 2, column: 6 },
        { messageId: 'redeclaredAsBuiltin', line: 2, column: 16 },
      ],
    },
    {
      code: 'var a;\nvar {a = 0, b: Object = 0} = {};',
      options: { builtinGlobals: false },
      errors: [{ messageId: 'redeclared', line: 2, column: 6 }],
    },
    {
      code: 'var globalThis = 0;',
      options: { builtinGlobals: true },
      errors: [{ messageId: 'redeclaredAsBuiltin', line: 1, column: 5 }],
    },
    {
      code: 'var a;\nvar {a = 0, b: globalThis = 0} = {};',
      options: { builtinGlobals: true },
      errors: [
        { messageId: 'redeclared', line: 2, column: 6 },
        { messageId: 'redeclaredAsBuiltin', line: 2, column: 16 },
      ],
    },
    {
      code: '/*global b:false*/ var b = 1;',
      options: { builtinGlobals: true },
      errors: [{ messageId: 'redeclaredBySyntax', line: 1, column: 10 }],
    },
    {
      code: '/*global b:true*/ var b = 1;',
      options: { builtinGlobals: true },
      errors: [{ messageId: 'redeclaredBySyntax', line: 1, column: 10 }],
    },
    {
      code: '/*globals Array */',
      options: { builtinGlobals: true },
      errors: [{ messageId: 'redeclaredAsBuiltin', line: 1, column: 11 }],
    },
    {
      code: '/*globals parseInt */',
      options: { builtinGlobals: true },
      errors: [{ messageId: 'redeclaredAsBuiltin', line: 1, column: 11 }],
    },
    {
      code: '/*globals foo, Array */',
      options: { builtinGlobals: true },
      errors: [{ messageId: 'redeclaredAsBuiltin', line: 1, column: 16 }],
    },
    {
      code: '/*globals a:readonly, Array:writable */',
      options: { builtinGlobals: true },
      errors: [{ messageId: 'redeclaredAsBuiltin', line: 1, column: 23 }],
    },
    {
      code: '/*globals\nArray */',
      options: { builtinGlobals: true },
      errors: [{ messageId: 'redeclaredAsBuiltin', line: 2, column: 1 }],
    },
    {
      code: '/*globals foo,\n    Array */',
      options: { builtinGlobals: true },
      errors: [{ messageId: 'redeclaredAsBuiltin', line: 2, column: 5 }],
    },
    {
      code: '/*globals a */ /*globals a */',
      errors: [{ messageId: 'redeclared', line: 1, column: 26 }],
    },
    {
      code: '/*globals a */',
      options: { builtinGlobals: true },
      errors: [{ messageId: 'redeclaredAsBuiltin', line: 1, column: 11 }],
      languageOptions: { globals: { a: 'readonly' } },
    },
    {
      code: 'var configured = 1;',
      options: { builtinGlobals: true },
      errors: [{ messageId: 'redeclaredAsBuiltin', line: 1, column: 5 }],
      languageOptions: { globals: { configured: 'readonly' } },
    },
    {
      code: '/*globals Object */ var Object = 0;',
      options: { builtinGlobals: true },
      errors: [{ messageId: 'redeclaredBySyntax', line: 1, column: 11 }],
      languageOptions: { globals: { Object: 'off' } },
    },
    {
      code: 'export {};\n/*globals Array */',
      options: { builtinGlobals: true },
      errors: [{ messageId: 'redeclaredAsBuiltin', line: 2, column: 11 }],
    },
    {
      code: 'function f() {\n  var a;\n  var a;\n}',
      errors: [{ messageId: 'redeclared', line: 3, column: 7 }],
    },
    {
      code: 'function f(a) { var a; }',
      errors: [{ messageId: 'redeclared', line: 1, column: 21 }],
    },
    {
      code: 'function f() { var a; if (test) { var a; } }',
      errors: [{ messageId: 'redeclared', line: 1, column: 39 }],
    },
    {
      code: 'for (var a, a;;);',
      errors: [{ messageId: 'redeclared', line: 1, column: 13 }],
    },
  ],
});

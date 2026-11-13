// Ported from ESLint v10.14.0 tests/lib/rules/no-redeclare.js.
// https://github.com/eslint/eslint/blob/v10.14.0/tests/lib/rules/no-redeclare.js
// Every upstream case keeps its code, options and language options. Messages,
// message IDs and complete ranges were captured from ESLint v10.14.0.
// The rule has no fixes or suggestions; the shared tester asserts their absence.
import { RuleTester } from '@rslint/rule-tester';
import globals from 'globals';

const ruleTester = new RuleTester({
  languageOptions: { sourceType: 'script' },
});

ruleTester.run('no-redeclare', {
  valid: [
    'var a = 3; var b = function() { var a = 10; };',
    'var a = 3; a = 10;',
    {
      code: 'if (true) {\n    let b = 2;\n} else {    \nlet b = 3;\n}',
      languageOptions: {
        ecmaVersion: 6,
      },
    },
    {
      code: 'var a; class C { static { var a; } }',
      languageOptions: {
        ecmaVersion: 2022,
      },
    },
    {
      code: 'class C { static { var a; } } var a; ',
      languageOptions: {
        ecmaVersion: 2022,
      },
    },
    {
      code: 'function a(){} class C { static { var a; } }',
      languageOptions: {
        ecmaVersion: 2022,
      },
    },
    {
      code: 'var a; class C { static { function a(){} } }',
      languageOptions: {
        ecmaVersion: 2022,
      },
    },
    {
      code: 'class C { static { var a; } static { var a; } }',
      languageOptions: {
        ecmaVersion: 2022,
      },
    },
    {
      code: 'class C { static { function a(){} } static { function a(){} } }',
      languageOptions: {
        ecmaVersion: 2022,
      },
    },
    {
      code: 'class C { static { var a; { function a(){} } } }',
      languageOptions: {
        ecmaVersion: 2022,
      },
    },
    {
      code: 'class C { static { function a(){}; { function a(){} } } }',
      languageOptions: {
        ecmaVersion: 2022,
      },
    },
    {
      code: 'class C { static { var a; { let a; } } }',
      languageOptions: {
        ecmaVersion: 2022,
      },
    },
    {
      code: 'class C { static { let a; { let a; } } }',
      languageOptions: {
        ecmaVersion: 2022,
      },
    },
    {
      code: 'class C { static { { let a; } { let a; } } }',
      languageOptions: {
        ecmaVersion: 2022,
      },
    },
    {
      code: 'var Object = 0;',
      options: [
        {
          builtinGlobals: false,
        },
      ],
    },
    {
      code: 'var Object = 0;',
      options: [
        {
          builtinGlobals: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'module',
      },
    },
    {
      code: 'var Object = 0;',
      options: [
        {
          builtinGlobals: true,
        },
      ],
      languageOptions: {
        parserOptions: {
          ecmaFeatures: {
            globalReturn: true,
          },
        },
      },
    },
    {
      code: 'var top = 0;',
      options: [
        {
          builtinGlobals: true,
        },
      ],
    },
    {
      code: 'var top = 0;',
      options: [
        {
          builtinGlobals: true,
        },
      ],
      languageOptions: {
        parserOptions: {
          ecmaFeatures: {
            globalReturn: true,
          },
        },
        globals: globals.browser,
      },
    },
    {
      code: 'var top = 0;',
      options: [
        {
          builtinGlobals: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'module',
        globals: globals.browser,
      },
    },
    {
      code: 'var self = 1',
      options: [
        {
          builtinGlobals: true,
        },
      ],
    },
    {
      code: 'var globalThis = foo',
      options: [
        {
          builtinGlobals: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 6,
      },
    },
    {
      code: 'var globalThis = foo',
      options: [
        {
          builtinGlobals: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
    },
    {
      code: '/*globals Array */',
      options: [
        {
          builtinGlobals: false,
        },
      ],
    },
    {
      code: '/*globals a */',
      options: [
        {
          builtinGlobals: false,
        },
      ],
      languageOptions: {
        globals: {
          a: 'readonly',
        },
      },
    },
    {
      code: '/*globals a */',
      options: [
        {
          builtinGlobals: false,
        },
      ],
      languageOptions: {
        globals: {
          a: 'writable',
        },
      },
    },
    {
      code: '/*globals a:off */',
      options: [
        {
          builtinGlobals: true,
        },
      ],
      languageOptions: {
        globals: {
          a: 'readonly',
        },
      },
    },
    {
      code: '/*globals a */',
      options: [
        {
          builtinGlobals: true,
        },
      ],
      languageOptions: {
        globals: {
          a: 'off',
        },
      },
    },
  ],
  invalid: [
    {
      code: 'var a = 3; var a = 10;',
      languageOptions: {
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'redeclared',
          message: "'a' is already defined.",
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: 'switch(foo) { case a: var b = 3;\ncase b: var b = 4}',
      errors: [
        {
          messageId: 'redeclared',
          message: "'b' is already defined.",
          line: 2,
          column: 13,
          endLine: 2,
          endColumn: 14,
        },
      ],
    },
    {
      code: 'var a = 3; var a = 10;',
      errors: [
        {
          messageId: 'redeclared',
          message: "'a' is already defined.",
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: 'var a = {}; var a = [];',
      errors: [
        {
          messageId: 'redeclared',
          message: "'a' is already defined.",
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'var a; function a() {}',
      errors: [
        {
          messageId: 'redeclared',
          message: "'a' is already defined.",
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'function a() {} function a() {}',
      errors: [
        {
          messageId: 'redeclared',
          message: "'a' is already defined.",
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: 'var a = function() { }; var a = function() { }',
      errors: [
        {
          messageId: 'redeclared',
          message: "'a' is already defined.",
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: 'var a = function() { }; var a = new Date();',
      errors: [
        {
          messageId: 'redeclared',
          message: "'a' is already defined.",
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: 'var a = 3; var a = 10; var a = 15;',
      errors: [
        {
          messageId: 'redeclared',
          message: "'a' is already defined.",
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 17,
        },
        {
          messageId: 'redeclared',
          message: "'a' is already defined.",
          line: 1,
          column: 28,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: 'var a; var a;',
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'module',
      },
      errors: [
        {
          messageId: 'redeclared',
          message: "'a' is already defined.",
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      code: 'export var a; var a;',
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'module',
      },
      errors: [
        {
          messageId: 'redeclared',
          message: "'a' is already defined.",
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: 'class C { static { var a; var a; } }',
      languageOptions: {
        ecmaVersion: 2022,
      },
      errors: [
        {
          messageId: 'redeclared',
          message: "'a' is already defined.",
          line: 1,
          column: 31,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      code: 'class C { static { var a; { var a; } } }',
      languageOptions: {
        ecmaVersion: 2022,
      },
      errors: [
        {
          messageId: 'redeclared',
          message: "'a' is already defined.",
          line: 1,
          column: 33,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      code: 'class C { static { { var a; } var a; } }',
      languageOptions: {
        ecmaVersion: 2022,
      },
      errors: [
        {
          messageId: 'redeclared',
          message: "'a' is already defined.",
          line: 1,
          column: 35,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },
    {
      code: 'class C { static { { var a; } { var a; } } }',
      languageOptions: {
        ecmaVersion: 2022,
      },
      errors: [
        {
          messageId: 'redeclared',
          message: "'a' is already defined.",
          line: 1,
          column: 37,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    {
      code: 'var Object = 0;',
      options: [
        {
          builtinGlobals: true,
        },
      ],
      errors: [
        {
          messageId: 'redeclaredAsBuiltin',
          message: "'Object' is already defined as a built-in global variable.",
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    {
      code: 'var top = 0;',
      options: [
        {
          builtinGlobals: true,
        },
      ],
      languageOptions: {
        globals: globals.browser,
      },
      errors: [
        {
          messageId: 'redeclaredAsBuiltin',
          message: "'top' is already defined as a built-in global variable.",
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: 'var a; var {a = 0, b: Object = 0} = {};',
      options: [
        {
          builtinGlobals: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'redeclared',
          message: "'a' is already defined.",
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 14,
        },
        {
          messageId: 'redeclaredAsBuiltin',
          message: "'Object' is already defined as a built-in global variable.",
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: 'var a; var {a = 0, b: Object = 0} = {};',
      options: [
        {
          builtinGlobals: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'module',
      },
      errors: [
        {
          messageId: 'redeclared',
          message: "'a' is already defined.",
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: 'var a; var {a = 0, b: Object = 0} = {};',
      options: [
        {
          builtinGlobals: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 6,
        parserOptions: {
          ecmaFeatures: {
            globalReturn: true,
          },
        },
      },
      errors: [
        {
          messageId: 'redeclared',
          message: "'a' is already defined.",
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: 'var a; var {a = 0, b: Object = 0} = {};',
      options: [
        {
          builtinGlobals: false,
        },
      ],
      languageOptions: {
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'redeclared',
          message: "'a' is already defined.",
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: 'var globalThis = 0;',
      options: [
        {
          builtinGlobals: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 2020,
      },
      errors: [
        {
          messageId: 'redeclaredAsBuiltin',
          message:
            "'globalThis' is already defined as a built-in global variable.",
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: 'var a; var {a = 0, b: globalThis = 0} = {};',
      options: [
        {
          builtinGlobals: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 2020,
      },
      errors: [
        {
          messageId: 'redeclared',
          message: "'a' is already defined.",
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 14,
        },
        {
          messageId: 'redeclaredAsBuiltin',
          message:
            "'globalThis' is already defined as a built-in global variable.",
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      code: '/*global b:false*/ var b = 1;',
      options: [
        {
          builtinGlobals: true,
        },
      ],
      errors: [
        {
          messageId: 'redeclaredBySyntax',
          message: "'b' is already defined by a variable declaration.",
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    {
      code: '/*global b:true*/ var b = 1;',
      options: [
        {
          builtinGlobals: true,
        },
      ],
      errors: [
        {
          messageId: 'redeclaredBySyntax',
          message: "'b' is already defined by a variable declaration.",
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    {
      code: 'function f() { var a; var a; }',
      errors: [
        {
          messageId: 'redeclared',
          message: "'a' is already defined.",
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: 'function f(a) { var a; }',
      errors: [
        {
          messageId: 'redeclared',
          message: "'a' is already defined.",
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: 'function f() { var a; if (test) { var a; } }',
      errors: [
        {
          messageId: 'redeclared',
          message: "'a' is already defined.",
          line: 1,
          column: 39,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      code: 'for (var a, a;;);',
      errors: [
        {
          messageId: 'redeclared',
          message: "'a' is already defined.",
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: 'var Object = 0;',
      errors: [
        {
          messageId: 'redeclaredAsBuiltin',
          message: "'Object' is already defined as a built-in global variable.",
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    {
      code: 'var top = 0;',
      languageOptions: {
        globals: globals.browser,
      },
      errors: [
        {
          messageId: 'redeclaredAsBuiltin',
          message: "'top' is already defined as a built-in global variable.",
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '/*globals Array */',
      options: [
        {
          builtinGlobals: true,
        },
      ],
      errors: [
        {
          messageId: 'redeclaredAsBuiltin',
          message: "'Array' is already defined as a built-in global variable.",
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: '/*globals parseInt */',
      options: [
        {
          builtinGlobals: true,
        },
      ],
      errors: [
        {
          messageId: 'redeclaredAsBuiltin',
          message:
            "'parseInt' is already defined as a built-in global variable.",
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '/*globals foo, Array */',
      options: [
        {
          builtinGlobals: true,
        },
      ],
      errors: [
        {
          messageId: 'redeclaredAsBuiltin',
          message: "'Array' is already defined as a built-in global variable.",
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '/* globals foo, Array, baz */',
      options: [
        {
          builtinGlobals: true,
        },
      ],
      errors: [
        {
          messageId: 'redeclaredAsBuiltin',
          message: "'Array' is already defined as a built-in global variable.",
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '/*global foo, Array, baz*/',
      options: [
        {
          builtinGlobals: true,
        },
      ],
      errors: [
        {
          messageId: 'redeclaredAsBuiltin',
          message: "'Array' is already defined as a built-in global variable.",
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '/*global array, Array*/',
      options: [
        {
          builtinGlobals: true,
        },
      ],
      errors: [
        {
          messageId: 'redeclaredAsBuiltin',
          message: "'Array' is already defined as a built-in global variable.",
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '/*globals a,Array*/',
      options: [
        {
          builtinGlobals: true,
        },
      ],
      errors: [
        {
          messageId: 'redeclaredAsBuiltin',
          message: "'Array' is already defined as a built-in global variable.",
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: '/*globals a:readonly, Array:writable */',
      options: [
        {
          builtinGlobals: true,
        },
      ],
      errors: [
        {
          messageId: 'redeclaredAsBuiltin',
          message: "'Array' is already defined as a built-in global variable.",
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: '\n/*globals Array */',
      options: [
        {
          builtinGlobals: true,
        },
      ],
      errors: [
        {
          messageId: 'redeclaredAsBuiltin',
          message: "'Array' is already defined as a built-in global variable.",
          line: 2,
          column: 11,
          endLine: 2,
          endColumn: 16,
        },
      ],
    },
    {
      code: '/*globals\nArray */',
      options: [
        {
          builtinGlobals: true,
        },
      ],
      errors: [
        {
          messageId: 'redeclaredAsBuiltin',
          message: "'Array' is already defined as a built-in global variable.",
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 6,
        },
      ],
    },
    {
      code: '\n/*globals\n\nArray*/',
      options: [
        {
          builtinGlobals: true,
        },
      ],
      errors: [
        {
          messageId: 'redeclaredAsBuiltin',
          message: "'Array' is already defined as a built-in global variable.",
          line: 4,
          column: 1,
          endLine: 4,
          endColumn: 6,
        },
      ],
    },
    {
      code: '/*globals foo,\n    Array */',
      options: [
        {
          builtinGlobals: true,
        },
      ],
      errors: [
        {
          messageId: 'redeclaredAsBuiltin',
          message: "'Array' is already defined as a built-in global variable.",
          line: 2,
          column: 5,
          endLine: 2,
          endColumn: 10,
        },
      ],
    },
    {
      code: '/*globals a */',
      options: [
        {
          builtinGlobals: true,
        },
      ],
      languageOptions: {
        globals: {
          a: 'readonly',
        },
      },
      errors: [
        {
          messageId: 'redeclaredAsBuiltin',
          message: "'a' is already defined as a built-in global variable.",
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    {
      code: '/*globals a */',
      options: [
        {
          builtinGlobals: true,
        },
      ],
      languageOptions: {
        globals: {
          a: 'writable',
        },
      },
      errors: [
        {
          messageId: 'redeclaredAsBuiltin',
          message: "'a' is already defined as a built-in global variable.",
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    {
      code: '/*globals a */ /*globals a */',
      errors: [
        {
          messageId: 'redeclared',
          message: "'a' is already defined.",
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: '/*globals a */ /*globals a */ var a = 0',
      options: [
        {
          builtinGlobals: true,
        },
      ],
      languageOptions: {
        globals: {
          a: 'writable',
        },
      },
      errors: [
        {
          messageId: 'redeclaredAsBuiltin',
          message: "'a' is already defined as a built-in global variable.",
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 12,
        },
        {
          messageId: 'redeclaredAsBuiltin',
          message: "'a' is already defined as a built-in global variable.",
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 27,
        },
        {
          messageId: 'redeclaredAsBuiltin',
          message: "'a' is already defined as a built-in global variable.",
          line: 1,
          column: 35,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },
    {
      code: '/*globals a */ var a;',
      options: [
        {
          builtinGlobals: true,
        },
      ],
      languageOptions: {
        globals: {
          a: 'off',
        },
      },
      errors: [
        {
          messageId: 'redeclaredBySyntax',
          message: "'a' is already defined by a variable declaration.",
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
  ],
});

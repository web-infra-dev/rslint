import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-var', {
  valid: [
    "const JOE = 'schmoe';",
    "let moo = 'car';",
    'const a = 1; let b = 2;',
    'for (let i = 0; i < 10; i++) {}',
    'for (const x of [1, 2]) {}',
    'declare global { var bar: string; }',
    'declare global {\n  var g1: string;\n  var g2: number;\n}',
  ],
  invalid: [
    // Script mode (global scope)
    {
      code: 'var foo = bar;',
      errors: [{ messageId: 'unexpectedVar' }],
    },
    {
      code: 'var foo = bar, toast = most;',
      errors: [{ messageId: 'unexpectedVar' }],
    },
    {
      code: 'var foo = bar; var baz = quux;',
      errors: [
        { messageId: 'unexpectedVar', line: 1, column: 1 },
        { messageId: 'unexpectedVar', line: 1, column: 16 },
      ],
    },
    {
      code: 'if (true) { var x = 1; }',
      errors: [{ messageId: 'unexpectedVar', line: 1, column: 13 }],
    },
    {
      code: 'for (var i = 0; i < 10; i++) {}',
      errors: [{ messageId: 'unexpectedVar', line: 1, column: 6 }],
    },
    {
      code: 'for (var x in obj) {}',
      errors: [{ messageId: 'unexpectedVar', line: 1, column: 6 }],
    },
    {
      code: 'for (var x of [1, 2]) {}',
      errors: [{ messageId: 'unexpectedVar', line: 1, column: 6 }],
    },
    {
      code: 'var { a, b } = { a: 1, b: 2 };',
      errors: [{ messageId: 'unexpectedVar' }],
    },
    {
      code: 'var [c, d] = [1, 2];',
      errors: [{ messageId: 'unexpectedVar' }],
    },
    // export var
    {
      code: 'export var exported = 1;',
      errors: [{ messageId: 'unexpectedVar', line: 1, column: 8 }],
    },
    // declare var
    {
      code: 'declare var declaredVar: number;',
      errors: [{ messageId: 'unexpectedVar', line: 1, column: 1 }],
    },
    // declare namespace
    {
      code: 'declare namespace NS { var nsVar: string; }',
      errors: [{ messageId: 'unexpectedVar', line: 1, column: 24 }],
    },
    // declare module
    {
      code: "declare module 'my-mod' { var modVar: string; }",
      errors: [{ messageId: 'unexpectedVar', line: 1, column: 27 }],
    },
    // var in function
    {
      code: 'function outer() { var nested = 1; }',
      errors: [{ messageId: 'unexpectedVar' }],
    },
    // switch case
    {
      code: 'switch (0) { case 0: var sw = 1; break; }',
      errors: [{ messageId: 'unexpectedVar' }],
    },
    // Module-mode tests
    {
      code: 'export {}; var foo = 1;',
      errors: [{ messageId: 'unexpectedVar' }],
    },
    {
      code: 'export {}; for (var i = 0; i < 10; i++) {}',
      errors: [{ messageId: 'unexpectedVar' }],
    },
    {
      code: 'export {}; for (var x in {}) {}',
      errors: [{ messageId: 'unexpectedVar' }],
    },
    {
      code: 'export {}; for (var x of [1, 2]) {}',
      errors: [{ messageId: 'unexpectedVar' }],
    },
    // TDZ: self-reference
    {
      code: 'export {}; function f() { var a = a; }',
      errors: [{ messageId: 'unexpectedVar' }],
    },
    // TDZ: destructuring default self-ref
    {
      code: 'export {}; function f() { var {a = a} = {}; }',
      errors: [{ messageId: 'unexpectedVar' }],
    },
    // TDZ: forward reference
    {
      code: 'export {}; function f() { var a = b, b = 1; }',
      errors: [{ messageId: 'unexpectedVar' }],
    },
    // TDZ: backward ref is safe (still reports, but fix is applied)
    {
      code: 'export {}; function f() { var {a, b = a} = {}; }',
      errors: [{ messageId: 'unexpectedVar' }],
    },
    // TDZ: function self-ref safe
    {
      code: 'export {}; var foo = function() { foo(); };',
      errors: [{ messageId: 'unexpectedVar' }],
    },
    // TDZ: IIFE not safe
    {
      code: 'export {}; var foo = (function() { foo(); })();',
      errors: [{ messageId: 'unexpectedVar' }],
    },
    // Redeclared
    {
      code: 'export {}; function f() { var x = 1; var x = 2; }',
      errors: [{ messageId: 'unexpectedVar' }, { messageId: 'unexpectedVar' }],
    },
    // Used from outside scope
    {
      code: 'export {}; function f() { if (true) { var x = 1; } console.log(x); }',
      errors: [{ messageId: 'unexpectedVar' }],
    },
    // Variable name is `let`
    {
      code: 'function f() { var let; }',
      errors: [{ messageId: 'unexpectedVar' }],
    },
    // Referenced before declaration
    {
      code: 'export {}; function f() { console.log(x); var x = 1; }',
      errors: [{ messageId: 'unexpectedVar' }],
    },
    // Closure in loop
    {
      code: 'export {}; function f() { for (var a of [1]) { setTimeout(() => console.log(a)); } }',
      errors: [{ messageId: 'unexpectedVar' }],
    },
    // Uninitialized in loop
    {
      code: "export {}; function f() { for (let i of [1]) { var c; console.log(c); c = 'hello'; } }",
      errors: [{ messageId: 'unexpectedVar' }],
    },
    // Statement position
    {
      code: 'export {}; function f() { if (true) var bar = 1; }',
      errors: [{ messageId: 'unexpectedVar' }],
    },
    // Partial fix: separate statements
    {
      code: 'export {}; var a = b; var b = 1;',
      errors: [{ messageId: 'unexpectedVar' }, { messageId: 'unexpectedVar' }],
    },
  ],
});

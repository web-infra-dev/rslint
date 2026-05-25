import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-global-assign', {
  valid: [
    // Lowercase identifier is not a builtin
    "string = 'hello world';",

    // Shadowed by var declaration
    "var String: any; String = 'test';",

    // Shadowed by let declaration
    'let Array: any; Array = 1;',

    // Shadowed by function parameter
    "function foo(String: any) { String = 'test'; }",

    // Shadowed by function declaration
    "function Object() {} Object = 'test';",

    // Exception option
    {
      code: 'Object = 0;',
      options: { exceptions: ['Object'] },
    },

    // Read-only usage (not a write reference)
    'var x = String(123);',

    // Property access (not an identifier assignment)
    'var x = Math.PI;',

    // Not a builtin name
    "foo = 'bar';",

    // Shadowed by class declaration
    'class Array {} Array = 1;',

    // Var shadows global inside function
    'function foo() { var Object: any; Object = 1; }',

    // Block-scoped shadow with let
    "{ let String: any; String = 'test'; }",

    // For-loop let variable shadows
    'for (let Object = 0; Object < 10; Object++) {}',

    // For-in let variable shadows
    'for (let String in {}) {}',

    // For-of let variable shadows
    'for (let Array of []) {}',

    // Enum declaration shadows (TS)
    'enum Promise { A, B } Promise = 1;',

    // Function expression name shadows inside body
    'const x = function Number() { Number = 1; };',

    // Arrow function parameter shadows
    'const f = (Object: any) => { Object = 1; };',

    // Catch variable shadows
    "try {} catch (String) { String = 'x'; }",

    // Import default shadows
    'import Array from "some-module"; Array = 1;',

    // Var hoisting shadows (assignment before declaration)
    'Number = 1; var Number: any;',

    // Shadowed by parameter in constructor
    'class C { constructor(Object: any) { Object = 1; } }',

    // Shadowed by parameter in setter
    'class C { set prop(Array: any) { Array = 1; } }',

    // Var hoists from for-loop to enclosing function scope
    'function f() { for (var Object = 0; Object < 1; Object++) {} Object = 99; }',

    // Var hoists from nested if block to function scope
    'function f() { Object = 1; if (true) { var Object: any; } }',

    // Class expression name shadows inside class body
    'const c = class Object { method() { Object = 1; } };',

    // Function expression name shadows in nested function within body
    'const x = function Object() { function inner() { Object = 1; } };',

    // Var hoists from switch case
    'function f() { switch(1) { case 1: var Object: any; } Object = 1; }',

    // Var hoists from try block
    'function f() { try { var Object: any; } catch(e) {} Object = 1; }',

    // Var hoists from labeled statement
    'function f() { label: { var Object: any; } Object = 1; }',

    // Var hoists from while body
    'function f() { while(false) { var Object: any; } Object = 1; }',

    // Var hoists from do-while body
    'function f() { do { var Object: any; } while(false); Object = 1; }',

    // Var hoists from for-in body
    'function f() { for (var x in {}) { var Object: any; } Object = 1; }',

    // Var hoists from deeply nested blocks
    'function f() { if (true) { for (;;) { { var Object: any; break; } } } Object = 1; }',

    // Namespace declaration shadows global
    'namespace Map { export const x = 1; } Map = 1;',

    // Const enum shadows global
    'const enum Set { A, B } Set = 1;',

    // Declare var shadows global
    'declare var WeakMap: any; WeakMap = 1;',

    // Declare function shadows global
    'declare function Symbol(): any; Symbol = 1;',

    // Declare class shadows global
    'declare class Promise {} Promise = 1;',

    // Let destructuring in for-of is a declaration (not global write)
    'for (let {Object} of [{}]) {}',

    // Type assertion write is NOT detected by ESLint scope analysis
    '((Object as any) as any) = 1;',

    // Satisfies expression write is NOT detected by ESLint scope analysis
    '(Object satisfies any) = 1;',

    // Object destructuring parameter shadows
    'function f({Object}: any) { Object = 1; }',

    // Array destructuring parameter shadows
    'function f([Object]: any) { Object = 1; }',

    // Renamed destructuring parameter shadows
    'function f({a: Object}: any) { Object = 1; }',

    // Destructuring parameter with default shadows
    'function f({Object = 0}: any) { Object = 1; }',

    // Nested destructuring parameter shadows
    'function f({a: {b: Object}}: any) { Object = 1; }',

    // Arrow with destructuring parameter shadows
    'const f = ({Object}: any) => { Object = 1; };',

    // Method with destructuring parameter shadows
    'class C { method({Object}: any) { Object = 1; } }',

    // Destructured catch variable shadows
    'try {} catch ({Object}: any) { Object = 1; }',

    // Var with object destructuring shadows
    'var {Object}: any = {}; Object = 1;',

    // Let with array destructuring shadows
    'let [Array]: any = []; Array = 1;',

    // Var with renamed destructuring shadows
    'var {a: Map}: any = {}; Map = 1;',

    // Var with nested destructuring shadows
    'var {a: {b: Set}}: any = {}; Set = 1;',

    // Hoisted var destructuring from nested block
    'function f() { if (true) { var {Object}: any = {}; } Object = 1; }',

    // Hoisted var array destructuring from nested block
    'function f() { if (true) { var [Array]: any = []; } Array = 1; }',

    // Hoisted var destructuring from switch
    'function f() { switch(1) { case 1: var {Object}: any = {}; } Object = 1; }',

    // Import equals declaration shadows
    "import Object = require('some-module'); Object = 1;",

    // Hoisted var destructuring in for-of
    'function f() { for (var {Object}: any of [{}]) {} Object = 1; }',

    // Multiple var declarations in one statement shadow both
    'var Object: any, Array: any; Object = 1; Array = 1;',

    // Closure accessing outer var
    'function f() { var Object: any; function inner() { Object = 1; } }',

    // Var after inner function still shadows (hoisting)
    'function f() { function inner() { Object = 1; } var Object: any; }',

    // Constructor parameter property shadows
    'class C { constructor(public Object: any) { Object = 1; } }',

    // Declare enum shadows
    'declare enum Promise { A, B } Promise = 1;',

    // Nested satisfies + as is not a real write
    '((Object satisfies any) as any) = 1;',

    // Nested as + satisfies is not a real write
    '((Object as any) satisfies any) = 1;',

    // Double satisfies is not a real write
    '((Object satisfies any) satisfies any) = 1;',

    // Non-null wrapped in satisfies is not a real write
    '((Object!) satisfies any) = 1;',

    // Satisfies wrapped in non-null is not a real write
    '((Object satisfies any)!) = 1;',

    // Type-only import shadows in ESLint scope
    'import type Object from "test"; Object = 1;',

    // Named import with alias shadows
    'import { foo as Object } from "test"; Object = 1;',

    // Namespace import shadows
    'import * as Object from "test"; Object = 1;',

    // Named import shadows
    'import { Object } from "test"; Object = 1;',

    // Overloaded function declaration shadows
    'function Object(x: string): string;\nfunction Object(x: number): number;\nfunction Object(x: any): any { return x; }\nObject = 1;',

    // Using declaration shadows in same scope
    'function f() { using Object = { [Symbol.dispose]() {} }; Object = 1; }',

    // Await using declaration shadows in same scope
    'async function f() { await using Object = { async [Symbol.asyncDispose]() {} }; Object = 1; }',

    // Var with computed property destructuring shadows
    'const k = "x"; var {[k]: Object}: any = {}; Object = 1;',
  ],
  invalid: [
    // Direct assignment to builtin
    {
      code: "String = 'hello world';",
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Postfix increment on builtin
    {
      code: 'String++;',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Assignment to Array
    {
      code: 'Array = 1;',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Assignment to Number
    {
      code: 'Number = 1;',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Compound assignment
    {
      code: 'Math += 1;',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Prefix decrement
    {
      code: '--Object;',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Assignment to undefined
    {
      code: 'undefined = 1;',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Assignment to NaN
    {
      code: 'NaN = 1;',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Assignment to Infinity
    {
      code: 'Infinity = 1;',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Multiple globals assigned
    {
      code: 'String = 1; Array = 2;',
      errors: [
        { messageId: 'globalShouldNotBeModified' },
        { messageId: 'globalShouldNotBeModified' },
      ],
    },

    // Destructuring assignment of global
    {
      code: '({Object} = {});',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Destructuring with default value
    {
      code: '({Object = 0, String = 0} = {});',
      errors: [
        { messageId: 'globalShouldNotBeModified' },
        { messageId: 'globalShouldNotBeModified' },
      ],
    },

    // Array destructuring
    {
      code: "[String] = ['x'];",
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Object destructuring with rename
    {
      code: '({a: Object} = {});',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Rest element in destructuring
    {
      code: '[...Array] = [1, 2];',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // For-in with global as target
    {
      code: 'for (Object in {}) {}',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // For-of with global as target
    {
      code: 'for (Array of []) {}',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Logical assignment
    {
      code: 'Object ??= {};',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Inner var does NOT shadow outer scope
    {
      code: 'function f() { Object = 1; function inner() { var Object: any; } }',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Interface does NOT shadow value binding (TS)
    {
      code: 'interface JSON { x: number } JSON = 1;',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Type alias does NOT shadow value binding (TS)
    {
      code: 'type RegExp = string; RegExp = 1;',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Assignment inside nested function
    {
      code: 'function outer() { Object = 1; }',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Assignment inside arrow function
    {
      code: "const f = () => { String = 'x'; };",
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Assignment inside class method
    {
      code: 'class C { method() { Object = 1; } }',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Deeply nested array destructuring
    {
      code: "[[[String]]] = [[['x']]];",
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Chained assignment
    {
      code: 'Object = Array = 1;',
      errors: [
        { messageId: 'globalShouldNotBeModified' },
        { messageId: 'globalShouldNotBeModified' },
      ],
    },

    // Function expression name only shadows inside, not outside
    {
      code: 'const x = function Number() {}; Number = 1;',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Object destructuring in for-of
    {
      code: 'for ({Object} of [{}]) {}',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Array destructuring in for-of
    {
      code: 'for ([Object] of [[1]]) {}',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Var in arrow function body does NOT hoist to outer
    {
      code: 'function f() { const fn = () => { var Object: any; }; Object = 1; }',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Var in class method does NOT hoist to outer
    {
      code: 'function f() { class C { m() { var Object: any; } } Object = 1; }',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Let does NOT shadow outside for-loop
    {
      code: 'function f() { for (let Object = 0;;) { break; } Object = 1; }',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Const does NOT shadow outside for-of
    {
      code: 'function f() { for (const Object of []) {} Object = 1; }',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Nested destructuring in for-of
    {
      code: 'for ({a: {b: Object}} of [{a: {b: 1}}]) {}',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Rest element in for-of destructuring
    {
      code: 'for ([...Object] of [[1,2]]) {}',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Destructuring with default in for-of
    {
      code: 'for ({Object = 0} of [{}]) {}',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Non-null assertion write
    {
      code: '(Object!) = 1;',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Nested parenthesized write
    {
      code: '(((((Object))))) = 1;',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Var in static block does NOT hoist to enclosing scope
    {
      code: 'class C { static { var Object: any; } } Object = 1;',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Assignment in class field initializer
    {
      code: 'class C { field = (Object = 1); }',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Assignment inside comma expression
    {
      code: '(0, Object = 1);',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Assignment in default parameter value
    {
      code: 'function f(x = (Object = 1)) {}',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Assignment in generator function
    {
      code: 'function* gen() { Object = 1; }',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Assignment in async function
    {
      code: 'async function af() { Object = 1; }',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // IIFE var does NOT hoist out
    {
      code: '(function() { var Object: any; })(); Object = 1;',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Assignment in while condition
    {
      code: 'while (Object = 1) { break; }',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Assignment in for-loop update
    {
      code: 'for (let i = 0; i < 1; Object++) {}',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Module augmentation interface does NOT shadow
    {
      code: 'declare module "test" {\n  interface Object { custom: string; }\n}\nObject = 1;',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Declare module var does NOT shadow outside
    {
      code: 'declare module "foo" {\n  var Object: any;\n}\nObject = 1;',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Var inside namespace does NOT hoist to file scope
    {
      code: 'namespace Foo {\n  var Object: any;\n}\nObject = 1;',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Declare global var does NOT shadow
    {
      code: 'declare global {\n  var Object: any;\n}\nObject = 1;',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Assignment in template expression
    {
      code: 'const s = `${Object = 1}`;',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Assignment in array literal
    {
      code: 'const a = [Object = 1];',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Using declaration in nested block does NOT hoist
    {
      code: 'function f() {\n  if (true) {\n    using Object = { [Symbol.dispose]() {} };\n  }\n  Object = 1;\n}',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Object rest in destructuring assignment
    {
      code: '({...Object} = {a: 1});',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Mixed object+array destructuring
    {
      code: '({a: [Object]} = {a: [1]});',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Mixed array+object destructuring
    {
      code: '([{a: Object}] = [{a: 1}]);',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Skip + rest in array destructuring
    {
      code: '[, ...Object] = [1, 2, 3];',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // For-await-of with global target
    {
      code: 'async function f() { for await (Object of []) {} }',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Export default with assignment
    {
      code: 'export default Object = 1;',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Computed property destructuring
    {
      code: 'const key = "x"; ({[key]: Object} = {x: 1});',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Getter body var does NOT hoist to outer function
    {
      code: 'function f() { const o = { get x() { var Object: any; return 1; } }; Object = 1; }',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Destructuring in for-in: object shorthand
    {
      code: 'for ({Object} in {}) {}',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Destructuring in for-in: array
    {
      code: 'for ([Object] in {}) {}',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Destructuring in for-in: rename
    {
      code: 'for ({a: Object} in {}) {}',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Destructuring in for-in: rest
    {
      code: 'for ([...Object] in {}) {}',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Object method shorthand body
    {
      code: 'const o = { m() { Object = 1; } };',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Static method body
    {
      code: 'class C { static m() { Object = 1; } }',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Optional catch binding (no catch parameter)
    {
      code: 'try {} catch { Object = 1; }',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Arrow in class field
    {
      code: 'class C { prop = () => { Object = 1; }; }',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Setter body var does NOT hoist outside
    {
      code: 'function f() { const o = { set x(v: any) { var Object: any; } }; Object = 1; }',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Await using in nested block does NOT hoist
    {
      code: 'async function f() {\n  if (true) {\n    await using Object = { async [Symbol.asyncDispose]() {} };\n  }\n  Object = 1;\n}',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // AggregateError is a read-only global
    {
      code: 'AggregateError = 1;',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // FinalizationRegistry is a read-only global
    {
      code: 'FinalizationRegistry = 1;',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Intl is a read-only global
    {
      code: 'Intl = 1;',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Object rest in for-in destructuring
    {
      code: 'for ({...Object} in {}) {}',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Default value in for-in destructuring
    {
      code: 'for ({Object = 0} in {}) {}',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Nested destructuring in for-in
    {
      code: 'for ({a: {b: Object}} in {}) {}',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Ternary with assignments to different globals
    {
      code: 'true ? (String = 1) : (Array = 2);',
      errors: [
        { messageId: 'globalShouldNotBeModified' },
        { messageId: 'globalShouldNotBeModified' },
      ],
    },

    // Inner catch param does NOT shadow outer scope
    {
      code: 'try { try {} catch (Object) {} Object = 1; } catch(e) {}',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Multiple globals in one array destructuring
    {
      code: '[Object, Array, ...String] = [1, 2, 3, 4];',
      errors: [
        { messageId: 'globalShouldNotBeModified' },
        { messageId: 'globalShouldNotBeModified' },
        { messageId: 'globalShouldNotBeModified' },
      ],
    },

    // Default value in destructuring triggers write to both globals
    {
      code: '({x: Object = Array = 1} = {});',
      errors: [
        { messageId: 'globalShouldNotBeModified' },
        { messageId: 'globalShouldNotBeModified' },
      ],
    },

    // Generator yield assignment
    {
      code: 'function* g() { Object = yield 1; }',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Async await assignment
    {
      code: 'async function f() { Object = await Promise.resolve(1); }',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Inner class does NOT shadow outer scope write
    {
      code: 'function f() { Object = 1; class Inner { m() { class Object {} } } }',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // Iterator is a read-only global
    {
      code: 'Iterator = 1;',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // SuppressedError is a read-only global
    {
      code: 'SuppressedError = 1;',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },

    // DisposableStack is a read-only global
    {
      code: 'DisposableStack = 1;',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },
  ],
});

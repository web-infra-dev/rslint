import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('symbol-description', {
  valid: [
    // From ESLint upstream
    'Symbol("Foo");',
    'var foo = "foo"; Symbol(foo);',
    'var Symbol = function () {}; Symbol();',
    'Symbol(); var Symbol = function () {};',
    'function bar() { var Symbol = function () {}; Symbol(); }',
    'function bar(Symbol) { Symbol(); }',

    // Argument shapes — any non-empty arg satisfies the rule
    'Symbol("");',
    'Symbol(`foo`);',
    'Symbol(`${x}`);',
    'Symbol(undefined);',
    'Symbol(null);',
    'Symbol(42);',
    'Symbol(cond ? "a" : "b");',
    'Symbol(getName());',
    'Symbol(...args);',
    'Symbol("a", "b");',

    // Not the global `Symbol` identifier being called
    'foo.Symbol();',
    'obj["Symbol"]();',
    'new Symbol();',
    'Symbol;',

    // Shadowing — various declaration forms
    'let Symbol = 1; Symbol();',
    'const Symbol = () => {}; Symbol();',
    'function Symbol() {} Symbol();',
    'class Symbol {} new Symbol();',
    'const f = (Symbol) => { Symbol(); };',
    'function f(...Symbol) { Symbol(); }',
    'function f({ Symbol }) { Symbol(); }',
    'var { Symbol } = obj; Symbol();',
    'var [Symbol] = arr; Symbol();',
    'try {} catch (Symbol) { Symbol(); }',
    'for (let Symbol of arr) { Symbol(); }',
    'for (var Symbol = 0;;) { Symbol(); }',
    'Symbol(); function Symbol() {}',

    // Outer shadow propagates into inner scopes
    'var Symbol = 1; function f() { Symbol(); }',
    'var Symbol = 1; const f = () => Symbol();',
    'var Symbol = 1; class C { m() { Symbol(); } }',

    // Import forms shadow the global Symbol
    'import { Symbol } from "x"; Symbol();',
    'import Symbol from "x"; Symbol();',
    'import * as Symbol from "x"; Symbol();',
    'import { Foo as Symbol } from "x"; Symbol();',

    // Tagged template is not a CallExpression
    'Symbol`foo`;',

    // Named expression self-reference shadows inside its own body
    'const f = function Symbol() { Symbol(); };',
    'const c = class Symbol { m() { Symbol(); } };',

    // TypeScript enum shadows the global value
    'enum Symbol { A } Symbol();',

    // Namespace / module with identifier name shadows
    'namespace Symbol {} Symbol();',
    'module Symbol {} Symbol();',
    'function f() { namespace Symbol {} Symbol(); }',

    // `declare` value declarations shadow like runtime ones
    'declare var Symbol: any; Symbol();',
    'declare function Symbol(): any; Symbol();',
    'declare const Symbol: any; Symbol();',

    // Class body computed key + class name — inner class name wins
    'class Symbol { [Symbol()]() {} }',
    'class X { static E = class Symbol { m() { Symbol(); } }; }',

    // Declaration merging: interface + const shadows
    'interface Symbol {} const Symbol = 1; Symbol();',

    // `import type` — ts-eslint scope manager treats as binding
    'import type { Symbol } from "x"; Symbol();',
    'import type Symbol from "x"; Symbol();',
    'import { type Symbol } from "x"; Symbol();',

    // `export declare` value forms shadow
    'export declare var Symbol: any; Symbol();',
    'export declare function Symbol(): any; Symbol();',
    'export declare namespace Symbol {} Symbol();',
    'declare namespace Symbol {} Symbol();',
    'declare class Symbol {} new Symbol(); Symbol();',

    // Inside a namespace block, local var/function shadow
    'namespace NS { var Symbol = 1; Symbol(); }',
    'namespace NS { function Symbol() {} Symbol(); }',

    // Nested-scope shadowing
    'for (const Symbol = foo; ;) { Symbol(); }',
    'class C { m() { function Symbol() {} Symbol(); } }',
    'const f = () => { class Symbol {} return Symbol(); };',
  ],
  invalid: [
    // From ESLint upstream
    { code: 'Symbol();', errors: [{ messageId: 'expected' }] },
    {
      code: 'Symbol(); Symbol = function () {};',
      errors: [{ messageId: 'expected' }],
    },

    // Position / nested-scope
    { code: 'var foo = Symbol();', errors: [{ messageId: 'expected' }] },
    {
      code: 'function f() { return Symbol(); }',
      errors: [{ messageId: 'expected' }],
    },
    { code: 'Symbol(\n);', errors: [{ messageId: 'expected' }] },

    // Parenthesized callee
    { code: '(Symbol)();', errors: [{ messageId: 'expected' }] },
    { code: '((Symbol))();', errors: [{ messageId: 'expected' }] },

    // Optional call
    { code: 'Symbol?.();', errors: [{ messageId: 'expected' }] },

    // Comments don't count as arguments
    {
      code: 'Symbol(/* no description */);',
      errors: [{ messageId: 'expected' }],
    },

    // Block-scoped shadow doesn't leak out
    {
      code: '{ let Symbol = 1; } Symbol();',
      errors: [{ messageId: 'expected' }],
    },
    {
      code: 'function f(Symbol) { Symbol(); } Symbol();',
      errors: [{ messageId: 'expected' }],
    },
    {
      code: '(function (Symbol) {})(1); Symbol();',
      errors: [{ messageId: 'expected' }],
    },

    // Call sites inside various containers, no shadow
    { code: 'const f = () => Symbol();', errors: [{ messageId: 'expected' }] },
    {
      code: 'class C { m() { Symbol(); } }',
      errors: [{ messageId: 'expected' }],
    },
    {
      code: 'class C { static { Symbol(); } }',
      errors: [{ messageId: 'expected' }],
    },

    // Multi-line
    { code: 'var s = Symbol(\n\n);', errors: [{ messageId: 'expected' }] },

    // Call embedded in various expression positions
    { code: 'throw Symbol();', errors: [{ messageId: 'expected' }] },
    { code: 'Symbol().toString();', errors: [{ messageId: 'expected' }] },
    { code: 'x || Symbol();', errors: [{ messageId: 'expected' }] },
    { code: '(a, Symbol());', errors: [{ messageId: 'expected' }] },

    // Class bodies: field initializer and computed method key
    { code: 'class C { x = Symbol(); }', errors: [{ messageId: 'expected' }] },
    {
      code: 'class C { [Symbol()]() {} }',
      errors: [{ messageId: 'expected' }],
    },
    // Object literal computed property key — different AST path than class
    {
      code: 'const obj = { [Symbol()]: 1 };',
      errors: [{ messageId: 'expected' }],
    },

    // Type-only declarations do NOT shadow the runtime value
    {
      code: 'type Symbol = string; Symbol();',
      errors: [{ messageId: 'expected' }],
    },
    {
      code: 'interface Symbol {} Symbol();',
      errors: [{ messageId: 'expected' }],
    },
    // Ambient module with string-literal name does not bind `Symbol`
    {
      code: 'declare module "x" {} Symbol();',
      errors: [{ messageId: 'expected' }],
    },
    // Type alias followed by runtime use — `type` is type-only
    {
      code: 'type Symbol = string; var s: Symbol = "x"; Symbol();',
      errors: [{ messageId: 'expected' }],
    },
    // Decorator position — the CallExpression is still detected
    {
      code: 'class C { @Symbol() method() {} }',
      errors: [{ messageId: 'expected' }],
    },

    // Class field named Symbol does NOT shadow the global
    {
      code: 'class C { accessor Symbol = 1; m() { Symbol(); } }',
      errors: [{ messageId: 'expected' }],
    },

    // Destructuring assignment is not a declaration
    {
      code: '({ Symbol } = obj); Symbol();',
      errors: [{ messageId: 'expected' }],
    },
    { code: '[Symbol] = arr; Symbol();', errors: [{ messageId: 'expected' }] },

    // declare global { var Symbol } doesn't bring a file-scope var
    {
      code: 'declare global { var Symbol: any; } Symbol();',
      errors: [{ messageId: 'expected' }],
    },

    // Type-position usage doesn't shadow the runtime value
    {
      code: 'const x: Symbol = null as any; Symbol();',
      errors: [{ messageId: 'expected' }],
    },
    {
      code: 'const x: { Symbol: any } = {} as any; Symbol();',
      errors: [{ messageId: 'expected' }],
    },

    // Each call evaluated independently
    { code: 'Symbol(); Symbol("b");', errors: [{ messageId: 'expected' }] },
    { code: 'Symbol("a"); Symbol();', errors: [{ messageId: 'expected' }] },
  ],
});

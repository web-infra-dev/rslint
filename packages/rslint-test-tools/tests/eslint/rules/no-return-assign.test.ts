import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-return-assign', {
  valid: [
    // ---- Upstream ESLint suite ----
    `module.exports = {'a': 1};`,
    `var result = a * b;`,
    `function x() { var result = a * b; return result; }`,
    `function x() { return (result = a * b); }`,
    {
      code: `function x() { var result = a * b; return result; }`,
      options: ['except-parens'] as any,
    },
    {
      code: `function x() { return (result = a * b); }`,
      options: ['except-parens'] as any,
    },
    {
      code: `function x() { var result = a * b; return result; }`,
      options: ['always'] as any,
    },
    {
      code: `function x() { return function y() { result = a * b }; }`,
      options: ['always'] as any,
    },
    {
      code: `() => { return (result = a * b); }`,
      options: ['except-parens'] as any,
    },
    {
      code: `() => (result = a * b)`,
      options: ['except-parens'] as any,
    },
    `const foo = (a,b,c) => ((a = b), c)`,
    `function foo(){
            return (a = b)
        }`,
    `function bar(){
            return function foo(){
                return (a = b) && c
            }
        }`,
    `const foo = (a) => (b) => (a = b)`,

    // ---- Non-assignment comparisons / declarations ----
    `function x() { return a == b; }`,
    `function x() { return a === b; }`,
    `function x() { var a = b; return a; }`,

    // ---- except-parens: nested parens directly wrapping the assignment ----
    `function x() { return ((a = b)); }`,
    `function x() { return (((a = b))); }`,
    `() => ((a = b))`,

    // ---- except-parens: parenthesised assignment inside a larger expression ----
    `function x() { return (a = b) && c; }`,
    `function x() { return c && (a = b); }`,
    `function x() { return (a = b) || c; }`,
    `function x() { return c || (a = b); }`,
    `function x() { return (a = b) ?? c; }`,
    `function x() { return (a = b) ? c : d; }`,
    `function x() { return c ? (a = b) : d; }`,
    `function x() { return c ? d : (a = b); }`,
    `function x() { return ((a = b), c); }`,
    `const foo = () => ((a = b), c)`,
    `function x() { return !(a = b); }`,
    `function x() { return typeof (a = b); }`,

    // ---- Sentinel blocks the walk-up ----
    `function x() { return function y() { a = b; }; }`,
    {
      code: `function x() { return function y() { a = b; }; }`,
      options: ['always'] as any,
    },
    `function x() { return class { m() { a = b; } }; }`,
    {
      code: `function x() { return class { m() { a = b; } }; }`,
      options: ['always'] as any,
    },
    `function x() { return () => { a = b; }; }`,
    {
      code: `function x() { return () => { a = b; }; }`,
      options: ['always'] as any,
    },
    {
      code: `function x() { return class { static { a = b; } }; }`,
      options: ['always'] as any,
    },

    // ---- Assignment outside any return/arrow-body ----
    `a = b;`,
    `function f() { a = b; }`,
    `if (x) { a = b; }`,
    `while (x) { a = b; }`,
    `for (let i = 0; i < 10; i++) { a = b; }`,
    `function f() { for (a = 0; a < 10; a++) {} return a; }`,
    `switch (x) { case 1: a = b; break; }`,
    `try { a = b; } catch (e) {}`,
    `class C { m() { a = b; } }`,
    `function f(a = (b = 1)) { return a; }`,

    // ---- TypeScript specific ----
    `function f(): number { return (a = b); }`,
    `function f(): number { return (a = b) as number; }`,
    `function f(): number { return (a = b) satisfies number; }`,
    `const f = (a: number): number => (a = 1)`,

    // ---- JSX coverage lives in Go tests (no_return_assign_test.go Tsx:true cases) ----
    // The shared eslint/rule-tester.ts hardcodes `src/virtual.ts`, and tsgo does
    // not parse JSX in .ts files. Extending that infra is out of scope here.
  ],
  invalid: [
    // ---- Upstream ESLint suite ----
    {
      code: `function x() { return result = a * b; };`,
      errors: [{ messageId: 'returnAssignment', line: 1, column: 16 }],
    },
    {
      code: `function x() { return (result) = (a * b); };`,
      errors: [{ messageId: 'returnAssignment', line: 1, column: 16 }],
    },
    {
      code: `function x() { return result = a * b; };`,
      options: ['except-parens'] as any,
      errors: [{ messageId: 'returnAssignment', line: 1, column: 16 }],
    },
    {
      code: `function x() { return (result) = (a * b); };`,
      options: ['except-parens'] as any,
      errors: [{ messageId: 'returnAssignment', line: 1, column: 16 }],
    },
    {
      code: `() => { return result = a * b; }`,
      errors: [{ messageId: 'returnAssignment', line: 1, column: 9 }],
    },
    {
      code: `() => result = a * b`,
      errors: [{ messageId: 'arrowAssignment', line: 1, column: 1 }],
    },
    {
      code: `function x() { return result = a * b; };`,
      options: ['always'] as any,
      errors: [{ messageId: 'returnAssignment', line: 1, column: 16 }],
    },
    {
      code: `function x() { return (result = a * b); };`,
      options: ['always'] as any,
      errors: [{ messageId: 'returnAssignment', line: 1, column: 16 }],
    },
    {
      code: `function x() { return result || (result = a * b); };`,
      options: ['always'] as any,
      errors: [{ messageId: 'returnAssignment', line: 1, column: 16 }],
    },
    {
      code: `function foo(){
                return a = b
            }`,
      errors: [{ messageId: 'returnAssignment', line: 2, column: 17 }],
    },
    {
      code: `function doSomething() {
                return foo = bar && foo > 0;
            }`,
      errors: [{ messageId: 'returnAssignment', line: 2, column: 17 }],
    },
    {
      code: `function doSomething() {
                return foo = function(){
                    return (bar = bar1)
                }
            }`,
      errors: [{ messageId: 'returnAssignment', line: 2, column: 17 }],
    },
    {
      code: `function doSomething() {
                return foo = () => a
            }`,
      errors: [{ messageId: 'returnAssignment', line: 2, column: 17 }],
    },
    {
      code: `function doSomething() {
                return () => a = () => b
            }`,
      errors: [{ messageId: 'arrowAssignment', line: 2, column: 24 }],
    },
    {
      code: `function foo(a){
                return function bar(b){
                    return a = b
                }
            }`,
      errors: [{ messageId: 'returnAssignment', line: 3, column: 21 }],
    },
    {
      code: `const foo = (a) => (b) => a = b`,
      errors: [{ messageId: 'arrowAssignment', line: 1, column: 20 }],
    },

    // ---- Assignment operator coverage (returnAssignment) ----
    {
      code: `function x() { return a -= b; }`,
      errors: [{ messageId: 'returnAssignment', line: 1, column: 16 }],
    },
    {
      code: `function x() { return a *= b; }`,
      errors: [{ messageId: 'returnAssignment', line: 1, column: 16 }],
    },
    {
      code: `function x() { return a /= b; }`,
      errors: [{ messageId: 'returnAssignment', line: 1, column: 16 }],
    },
    {
      code: `function x() { return a %= b; }`,
      errors: [{ messageId: 'returnAssignment', line: 1, column: 16 }],
    },
    {
      code: `function x() { return a **= b; }`,
      errors: [{ messageId: 'returnAssignment', line: 1, column: 16 }],
    },
    {
      code: `function x() { return a <<= b; }`,
      errors: [{ messageId: 'returnAssignment', line: 1, column: 16 }],
    },
    {
      code: `function x() { return a >>= b; }`,
      errors: [{ messageId: 'returnAssignment', line: 1, column: 16 }],
    },
    {
      code: `function x() { return a >>>= b; }`,
      errors: [{ messageId: 'returnAssignment', line: 1, column: 16 }],
    },
    {
      code: `function x() { return a &= b; }`,
      errors: [{ messageId: 'returnAssignment', line: 1, column: 16 }],
    },
    {
      code: `function x() { return a |= b; }`,
      errors: [{ messageId: 'returnAssignment', line: 1, column: 16 }],
    },
    {
      code: `function x() { return a ^= b; }`,
      errors: [{ messageId: 'returnAssignment', line: 1, column: 16 }],
    },
    {
      code: `function x() { return a &&= b; }`,
      errors: [{ messageId: 'returnAssignment', line: 1, column: 16 }],
    },
    {
      code: `function x() { return a ||= b; }`,
      errors: [{ messageId: 'returnAssignment', line: 1, column: 16 }],
    },
    {
      code: `function x() { return a ??= b; }`,
      errors: [{ messageId: 'returnAssignment', line: 1, column: 16 }],
    },

    // ---- Assignment operator coverage (arrowAssignment) ----
    {
      code: `() => a += b`,
      errors: [{ messageId: 'arrowAssignment', line: 1, column: 1 }],
    },
    {
      code: `() => a ??= b`,
      errors: [{ messageId: 'arrowAssignment', line: 1, column: 1 }],
    },
    {
      code: `(a, b) => a = b`,
      errors: [{ messageId: 'arrowAssignment', line: 1, column: 1 }],
    },

    // ---- Container coverage (returnAssignment) ----
    {
      code: `async function f() { return a = b; }`,
      errors: [{ messageId: 'returnAssignment', line: 1, column: 22 }],
    },
    {
      code: `function* g() { return a = b; }`,
      errors: [{ messageId: 'returnAssignment', line: 1, column: 17 }],
    },
    {
      code: `async function* ag() { return a = b; }`,
      errors: [{ messageId: 'returnAssignment', line: 1, column: 24 }],
    },
    {
      code: `class C { m() { return a = b; } }`,
      errors: [{ messageId: 'returnAssignment', line: 1, column: 17 }],
    },
    {
      code: `class C { get x() { return a = b; } }`,
      errors: [{ messageId: 'returnAssignment', line: 1, column: 21 }],
    },
    {
      code: `class C { set x(v) { return a = b; } }`,
      errors: [{ messageId: 'returnAssignment', line: 1, column: 22 }],
    },
    {
      code: `({ m() { return a = b; } })`,
      errors: [{ messageId: 'returnAssignment', line: 1, column: 10 }],
    },

    // ---- Container coverage (arrowAssignment) ----
    {
      code: `async () => a = b`,
      errors: [{ messageId: 'arrowAssignment', line: 1, column: 1 }],
    },

    // ---- Walk-up wrappers ----
    {
      code: `function x() { return a = b, c; }`,
      errors: [{ messageId: 'returnAssignment', line: 1, column: 16 }],
    },
    {
      code: `function x() { return a ? b = c : d; }`,
      errors: [{ messageId: 'returnAssignment', line: 1, column: 16 }],
    },
    {
      code: `function x() { return { x: a = b }; }`,
      errors: [{ messageId: 'returnAssignment', line: 1, column: 16 }],
    },
    {
      code: `function x() { return [a = b]; }`,
      errors: [{ messageId: 'returnAssignment', line: 1, column: 16 }],
    },
    {
      code: `function x() { return a = b && c; }`,
      errors: [{ messageId: 'returnAssignment', line: 1, column: 16 }],
    },
    {
      code: `function x() { return !(a = b); }`,
      options: ['always'] as any,
      errors: [{ messageId: 'returnAssignment', line: 1, column: 16 }],
    },
    {
      code: `function x() { return typeof (a = b); }`,
      options: ['always'] as any,
      errors: [{ messageId: 'returnAssignment', line: 1, column: 16 }],
    },

    // ---- always mode: parens don't exempt ----
    {
      code: `function x() { return ((a = b)); }`,
      options: ['always'] as any,
      errors: [{ messageId: 'returnAssignment', line: 1, column: 16 }],
    },
    {
      code: `() => ((a = b))`,
      options: ['always'] as any,
      errors: [{ messageId: 'arrowAssignment', line: 1, column: 1 }],
    },
    {
      code: `function x() { return c || (a = b); }`,
      options: ['always'] as any,
      errors: [{ messageId: 'returnAssignment', line: 1, column: 16 }],
    },
    {
      code: `function x() { return a + (b = c); }`,
      options: ['always'] as any,
      errors: [{ messageId: 'returnAssignment', line: 1, column: 16 }],
    },

    // ---- TypeScript syntax ----
    {
      code: `function f(): number { return a = b; }`,
      errors: [{ messageId: 'returnAssignment', line: 1, column: 24 }],
    },
    {
      code: `const f = (): number => a = b`,
      errors: [{ messageId: 'arrowAssignment', line: 1, column: 11 }],
    },
    {
      code: `function f(): number { return a = b as number; }`,
      errors: [{ messageId: 'returnAssignment', line: 1, column: 24 }],
    },

    // ---- Inner container reports; outer unaffected ----
    {
      code: `function f() { return (function g() { return a = b; })(); }`,
      errors: [{ messageId: 'returnAssignment', line: 1, column: 39 }],
    },
    {
      code: `function f() { return a = (b = c); }`,
      errors: [{ messageId: 'returnAssignment', line: 1, column: 16 }],
    },
    {
      code: `function f() { return a = (b = c); }`,
      options: ['always'] as any,
      errors: [
        { messageId: 'returnAssignment', line: 1, column: 16 },
        { messageId: 'returnAssignment', line: 1, column: 16 },
      ],
    },

    // ---- Multi-line ----
    {
      code: 'function x() {\n  return a\n    = b;\n}',
      errors: [{ messageId: 'returnAssignment', line: 2, column: 3 }],
    },
    {
      code: '() =>\n  a = b',
      errors: [{ messageId: 'arrowAssignment', line: 1, column: 1 }],
    },
  ],
});

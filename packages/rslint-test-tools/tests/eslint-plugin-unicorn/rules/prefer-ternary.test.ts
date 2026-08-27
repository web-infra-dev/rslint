import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const message = 'This `if` statement can be replaced by a ternary expression.';
const suggestionMessage = 'Use a ternary expression.';

ruleTester.run('prefer-ternary', null as never, {
  valid: [
    // Yield / await / throw are not mergeable.
    `if (test) { yield a; } else { yield b; }`,
    `if (test) { yield* a; } else { yield* b; }`,
    `if (test) { await a(); } else { await b(); }`,
    `if (test) { throw new Error('a'); } else { throw new TypeError('a'); }`,

    // Different left / different operator / ternary anywhere in the
    // shape — not mergeable.
    `if (test) { foo = a; } else { bar = b; }`,
    `if (test) { foo = a; } else { foo *= b; }`,
    `if (test) { foo().bar = a; } else { foo().bar = b; }`,
    `if (a ? b : c) { foo = a; } else { foo = b; }`,

    // Flat return statements / else-if chains / no consequent / calls
    // are not mergeable.
    `if (a) {b}`,
    `if (a) {} else {b}`,
    `if (a) {} else {}`,
    `if (test) { a(); } else { b(); }`,
    `function foo(){ if (a) return 1; else if (b) return 2; else if (c) return 3; else return 4; }`,

    // Variable-declaration gate is `let` only — `var` / `const` /
    // missing initializer / side-effectful initializer / variable
    // referenced in test or right side / non-adjacent / multiple
    // declarators / compound operator / destructuring / ternary init
    // / multi-statement body / `if` with `else` are all not gateable.
    `var x = a; if (test) { x = b; }`,
    `const x = a; if (test) { x = b; }`,
    `let x; if (test) { x = b; }`,
    `let x = foo(); if (test) { x = b; }`,
    `let x = new Foo(); if (test) { x = b; }`,
    `let x = a; if (x) { x = b; }`,
    `let x = a; if (test) { x = x + 1; }`,
    `let x = a; doSomething(); if (test) { x = b; }`,
    `let x = a, y = b; if (test) { x = c; }`,
    `let x = a; if (test) { x += b; }`,
    `let {a} = obj; if (test) { a = b; }`,
    `let x = condition ? a : b; if (test) { x = c; }`,
    `let x = a; if (test) { x = b; doSomething(); }`,
  ],
  invalid: [
    // Return statement → ternary.
    {
      code: `function unicorn() {
        if(test){
          return a;
        } else{
          return b;
        }
      }`,
      filename: 'file.ts',
      errors: [{ message }],
    },
    {
      code: `function unicorn() {
        if (test) return a; else return b;
      }`,
      filename: 'file.ts',
      errors: [{ message }],
    },
    // Assignment → ternary.
    {
      code: `function unicorn() {
        if(test){
          foo = a;
        } else{
          foo = b;
        }
      }`,
      filename: 'file.ts',
      errors: [{ message }],
    },
    {
      code: `function unicorn() {
        if(test){
          foo *= a;
        } else{
          foo *= b;
        }
      }`,
      filename: 'file.ts',
      errors: [{ message }],
    },
    // Bare `return;` is rewritten to `undefined`.
    {
      code: `function unicorn() {
        if(test){
          return;
        } else{
          return b;
        }
      }`,
      filename: 'file.ts',
      errors: [{ message }],
    },
    // `let x = a; if (test) { x = b; }` → suggestion `const x = test ? b : a;`
    {
      code: `let items = defaultData;
        if (data.length) {
          items = data;
        }`,
      filename: 'file.ts',
      errors: [
        {
          message,
          suggestions: [
            {
              messageId: 'prefer-ternary/suggestion',
              data: {},
              output: 'const items = data.length ? data : defaultData;',
            },
          ],
        },
      ],
    },
    // `let` with later writes keeps `let`.
    {
      code: `function foo() {
        let x = a;
        if (test) {
          x = b;
        }
        x = c;
      }`,
      filename: 'file.ts',
      errors: [
        {
          message,
          suggestions: [
            {
              messageId: 'prefer-ternary/suggestion',
              data: {},
              output: `function foo() {
        let x = test ? b : a;
        x = c;
      }`,
            },
          ],
        },
      ],
    },
    // Precedence: `a = b` in test position requires parens.
    {
      code: `if (a = b) {
        foo = 1;
      } else foo = 2;`,
      filename: 'file.ts',
      errors: [{ message }],
    },
    // `await` / `yield` / `yield*` in test position require parens.
    {
      code: `function* unicorn() {
        if (yield a) {
          foo = 1;
        } else foo = 2;
      }`,
      filename: 'file.ts',
      errors: [{ message }],
    },
    {
      code: `function* unicorn() {
        if (yield* a) {
          foo = 1;
        } else foo = 2;
      }`,
      filename: 'file.ts',
      errors: [{ message }],
    },
  ],
});

void suggestionMessage;

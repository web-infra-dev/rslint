import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('require-atomic-updates', {
  valid: [
    'let foo; async function x() { foo += bar; }',
    'let foo; async function x() { foo = foo + bar; }',
    'let foo; async function x() { foo = await bar + foo; }',
    'async function x() { let foo; foo += await bar; }',
    'let foo; async function x() { foo = (await result)(foo); }',
    'let foo; async function x() { foo = bar(await something, foo) }',
    'function* x() { let foo; foo += yield bar; }',
    'const foo = {}; async function x() { foo.bar = await baz; }',
    'const foo = []; async function x() { foo[x] += 1;  }',
    'let foo; function* x() { foo = bar + foo; }',
    'async function x() { let foo; bar(() => baz += 1); foo += await amount; }',
    // Update expressions are not tracked as assignments
    `let foo; function cb() { return foo; } async function f(x) { foo; await x; foo++; }`,
    // Await in LHS base doesn't register the member target
    `let foo; function cb() { return foo; } async function f(ptr, x) { foo; (await ptr).foo = x; }`,
    // Loop body awaits don't propagate to post-loop code
    `let foo; function cb() { return foo; }
     async function f(x) { foo; for (let i = 0; i < 10; i++) { await x; } foo = 1; }`,
    `let foo; function cb() { return foo; }
     async function f(x) { foo; while (x) { await x; } foo = 1; }`,
    `let foo; function cb() { return foo; }
     async function f(gen) { foo; for await (const it of gen) {} foo = 1; }`,
    // for-of / for-in / for-await-of bodies don't inherit pre-loop reads
    `let foo; function cb() { return foo; }
     async function f(gen) { foo; for await (const item of gen) { foo = item; } }`,
    `let foo; function cb() { return foo; }
     async function f(arr, hook) { foo; for (const c of arr) { await hook(); foo = 1; } }`,
    // for-of assignment form: bare-identifier target is pure write
    `let foo; function cb() { return foo; }
     async function f(gen, x) { for (foo of gen) { await x; foo = 1; } }`,
    // for-of assignment-form destructuring without pre-read
    `let foo; function cb() { return foo; }
     async function f(gen) { for ({foo} of gen) {} }`,
    // for-loop update runs with a fresh segment — simple `=` with await doesn't report
    `let foo; function cb() { return foo; } async function f(x) { foo; for(;; foo = await x) {} }`,
    // switch where every case exits via return: post-switch sees no outdated
    `let foo; function cb() { return foo; } async function f(x, n) { foo; switch (n) { case 1: await x; return; case 2: return; } foo = 1; }`,
    // RHS arrow/function silently skips check (ESLint quirk)
    `let foo; function cb() { return foo; } async function f(x) { foo; await x; foo = () => {}; }`,
    `let foo; function cb() { return foo; } async function f(x) { foo; await x; foo = function () {}; }`,
    // try-always-returns: catch entry = pre-try state
    `let foo; function cb() { return foo; } async function f(x) { foo; try { await x; return; } catch {} foo = 1; }`,
    `let foo; function cb() { return foo; } async function f(x) { try { foo; await x; return; } catch { foo = 1; } }`,
    // Namespace with a function export, function RHS — no report (function-RHS skip)
    `namespace NS { export function foo() {} } function cb() { return NS.foo; } async function f(x: any) { NS.foo; await x; NS.foo = () => {}; }`,
    // Catch binding shadows outer
    `let foo; function cb() { return foo; }
     async function f(x) { try { foo; await x; } catch (foo) { foo = 1; } }`,
    // Block-scoped let shadows outer for the block duration
    `let foo; function cb() { return foo; }
     async function f(x) { foo; await x; { let foo = 1; foo = 2; } }`,
    // for-loop let shadow
    `let foo; function cb() { return foo; }
     async function f(x) { foo; await x; for (let foo = 0; foo < 1; foo++) {} }`,
    // Labeled break before await skips the outdating path
    `let foo; function cb() { return foo; }
     async function f(x) { foo; outer: for (;;) { if (x) break outer; await x; } foo = 1; }`,
    // Property name in a closure does not cause outer variable to escape
    `async function x(list, api) {
      list.map((item) => item.foo);
      let foo = 1;
      if (list.length && foo !== 2) {
        await api.update({ foo });
        foo = 2;
      }
    }`,
    'let foo; async function x() { foo = condition ? foo : await bar; }',
    'async function x() { let foo; bar(() => { let foo; blah(foo); }); foo += await result; }',
    'let foo; async function x() { foo = foo + 1; await bar; }',
    'async function x() { foo += await bar; }',
    // allowProperties
    {
      code: `
        async function a(foo) {
          if (foo.bar) {
            foo.bar = await something;
          }
        }
      `,
      options: { allowProperties: true },
    },
    {
      code: `
        function* g(foo) {
          baz = foo.bar;
          yield something;
          foo.bar = 1;
        }
      `,
      options: { allowProperties: true },
    },
  ],

  invalid: [
    {
      code: 'let foo; async function x() { foo += await amount; }',
      errors: [{ messageId: 'nonAtomicUpdate' }],
    },
    {
      code: 'if (1); let foo; async function x() { foo += await amount; }',
      errors: [{ messageId: 'nonAtomicUpdate' }],
    },
    {
      code: 'let foo; async function x() { while (condition) { foo += await amount; } }',
      errors: [{ messageId: 'nonAtomicUpdate' }],
    },
    {
      code: 'let foo; async function x() { foo = foo + await amount; }',
      errors: [{ messageId: 'nonAtomicUpdate' }],
    },
    {
      code: 'let foo; async function x() { foo = foo + (bar ? baz : await amount); }',
      errors: [{ messageId: 'nonAtomicUpdate' }],
    },
    {
      code: 'let foo; async function x() { foo = foo + (bar ? await amount : baz); }',
      errors: [{ messageId: 'nonAtomicUpdate' }],
    },
    {
      code: 'let foo; async function x() { foo = condition ? foo + await amount : somethingElse; }',
      errors: [{ messageId: 'nonAtomicUpdate' }],
    },
    {
      code: 'let foo; async function x() { foo = (condition ? foo : await bar) + await bar; }',
      errors: [{ messageId: 'nonAtomicUpdate' }],
    },
    {
      code: 'let foo; async function x() { foo += bar + await amount; }',
      errors: [{ messageId: 'nonAtomicUpdate' }],
    },
    {
      code: 'async function x() { let foo; bar(() => foo); foo += await amount; }',
      errors: [{ messageId: 'nonAtomicUpdate' }],
    },
    {
      code: 'let foo; function* x() { foo += yield baz }',
      errors: [{ messageId: 'nonAtomicUpdate' }],
    },
    {
      code: 'let foo; async function x() { foo = bar(foo, await something) }',
      errors: [{ messageId: 'nonAtomicUpdate' }],
    },
    {
      code: 'const foo = {}; async function x() { foo.bar += await baz }',
      errors: [{ messageId: 'nonAtomicObjectUpdate' }],
    },
    {
      code: 'const foo = []; async function x() { foo[bar].baz += await result;  }',
      errors: [{ messageId: 'nonAtomicObjectUpdate' }],
    },
    {
      code: 'const foo = {}; class C { #bar; async wrap() { foo.#bar += await baz } }',
      errors: [{ messageId: 'nonAtomicObjectUpdate' }],
    },
    {
      code: 'let foo; async function* x() { foo = (yield foo) + await bar; }',
      errors: [{ messageId: 'nonAtomicUpdate' }],
    },
    {
      code: 'let foo; async function x() { foo = foo + await result(foo); }',
      errors: [{ messageId: 'nonAtomicUpdate' }],
    },
    {
      code: 'let foo; async function x() { foo = await result(foo, await somethingElse); }',
      errors: [{ messageId: 'nonAtomicUpdate' }],
    },
    {
      code: 'function* x() { let foo; yield async function y() { foo += await bar; } }',
      errors: [{ messageId: 'nonAtomicUpdate' }],
    },
    {
      code: 'let foo; async function* x() { foo = await foo + (yield bar); }',
      errors: [{ messageId: 'nonAtomicUpdate' }],
    },
    {
      code: 'let foo; async function x() { foo = bar + await foo; }',
      errors: [{ messageId: 'nonAtomicUpdate' }],
    },
    {
      code: "let foo = ''; async function x() { foo += await bar; }",
      errors: [{ messageId: 'nonAtomicUpdate' }],
    },
    {
      code: 'let foo = 0; async function x() { foo = (a ? b : foo) + await bar; if (baz); }',
      errors: [{ messageId: 'nonAtomicUpdate' }],
    },
    {
      code: 'let foo = 0; async function x() { foo = (a ? b ? c ? d ? foo : e : f : g : h) + await bar; if (baz); }',
      errors: [{ messageId: 'nonAtomicUpdate' }],
    },
    // Property access - default (no allowProperties option)
    {
      code: `
        async function a(foo) {
          if (foo.bar) {
            foo.bar = await something;
          }
        }
      `,
      errors: [{ messageId: 'nonAtomicObjectUpdate' }],
    },
    {
      code: `
        function* g(foo) {
          baz = foo.bar;
          yield something;
          foo.bar = 1;
        }
      `,
      errors: [{ messageId: 'nonAtomicObjectUpdate' }],
    },
    // allowProperties: true still reports variable issues
    {
      code: `
        let foo;
        async function a() {
          if (foo) {
            foo = await something;
          }
        }
      `,
      options: { allowProperties: true },
      errors: [{ messageId: 'nonAtomicUpdate' }],
    },
    {
      code: `
        let foo;
        function* g() {
          baz = foo;
          yield something;
          foo = 1;
        }
      `,
      options: { allowProperties: true },
      errors: [{ messageId: 'nonAtomicUpdate' }],
    },
    // Chained assignment `a = b = 1`
    {
      code: `let a, b; function cb() { return a + b; } async function f(x) { a; b; await x; a = b = 1; }`,
      errors: [
        { messageId: 'nonAtomicUpdate' },
        { messageId: 'nonAtomicUpdate' },
      ],
    },
    // Destructuring assignment: object pattern
    {
      code: `let foo; function cb() { return foo; } async function f(src, x) { foo; await x; ({foo} = src); }`,
      errors: [{ messageId: 'nonAtomicObjectUpdate' }],
    },
    // Destructuring assignment: array pattern
    {
      code: `let foo; function cb() { return foo; } async function f(src, x) { foo; await x; [foo] = src; }`,
      errors: [{ messageId: 'nonAtomicObjectUpdate' }],
    },
    // Destructuring assignment: aliased target
    {
      code: `let foo; function cb() { return foo; } async function f(src, x) { foo; await x; ({a: foo} = src); }`,
      errors: [{ messageId: 'nonAtomicObjectUpdate' }],
    },
    // Destructuring assignment: default with await
    {
      code: `let foo; function cb() { return foo; } async function f(src, x) { foo; ({foo = await x} = src); }`,
      errors: [{ messageId: 'nonAtomicObjectUpdate' }],
    },
    // Destructuring assignment: rest
    {
      code: `let foo; function cb() { return foo; } async function f(src, x) { foo; await x; ({...foo} = src); }`,
      errors: [{ messageId: 'nonAtomicObjectUpdate' }],
    },
    // for-of declaration-form destructure default with await
    {
      code: `let foo = 1; function cb() { return foo; } async function f(gen) { for (const { x = foo + await gen.y } of [{}]) { foo = x; } }`,
      errors: [{ messageId: 'nonAtomicUpdate' }],
    },
    // for-of assignment-form destructure default with await
    {
      code: `let foo = 1; function cb() { return foo; } async function f(gen) { for ({ x = foo + await gen.y } of [{}]) { foo = x; } }`,
      errors: [{ messageId: 'nonAtomicUpdate' }],
    },
    // for-of assignment form: member target base IS a read
    {
      code: `let foo = {}; function cb() { return foo; } async function f(gen, x) { for (foo.bar of gen) { await x; foo = 1; } }`,
      errors: [{ messageId: 'nonAtomicUpdate' }],
    },
    // Compound `foo += await x` in for-loop update still reports
    {
      code: `let foo; function cb() { return foo; } async function f(x) { for(;; foo += await x) {} }`,
      errors: [{ messageId: 'nonAtomicUpdate' }],
    },
    // Destructure default re-read does NOT clear outer outdated
    {
      code: `let foo, bar; function cb() { return foo + bar; } async function f(src, x) { foo; bar; const {a = await x, b = foo + bar} = src; foo = 1; }`,
      errors: [{ messageId: 'nonAtomicUpdate' }],
    },
    // Labeled break carries state to the labeled loop's post-loop point
    {
      code: `let foo; function cb() { return foo; } async function f(x) { outer: for (;;) { for (;;) { foo; await x; break outer; } } foo = 1; }`,
      errors: [{ messageId: 'nonAtomicUpdate' }],
    },
    // Outer foo still outdated after a shadowed inner block
    {
      code: `let foo; function cb() { return foo; } async function f(x) { foo; await x; { let foo = 1; foo = 2; } foo = 3; }`,
      errors: [{ messageId: 'nonAtomicUpdate' }],
    },
    // Catch-binding shadow: outer write after try/catch reports
    {
      code: `let foo; function cb() { return foo; } async function f(x) { try { foo; await x; } catch (foo) { foo = 1; } foo = 2; }`,
      errors: [{ messageId: 'nonAtomicUpdate' }],
    },
    // Optional catch (no binding): `foo` inside IS outer
    {
      code: `let foo; function cb() { return foo; } async function f(x) { try { foo; await x; } catch { foo = 1; } foo = 2; }`,
      errors: [
        { messageId: 'nonAtomicUpdate' },
        { messageId: 'nonAtomicUpdate' },
      ],
    },
    // TypeScript namespace at outer scope is tracked as declared
    {
      code: `namespace NS { export var foo = 1; } function cb() { return NS.foo; } async function f(x: any) { NS.foo; await x; NS.foo = 1; }`,
      errors: [{ messageId: 'nonAtomicObjectUpdate' }],
    },
    // Namespace with function export, value RHS — still reports
    {
      code: `namespace NS { export function foo() {} } function cb() { return NS.foo; } async function f(x: any) { NS.foo; await x; NS.foo = 1 as any; }`,
      errors: [{ messageId: 'nonAtomicObjectUpdate' }],
    },
    // Nested namespace member write
    {
      code: `namespace A { export namespace B { export var foo = 1; } } function cb() { return A.B.foo; } async function f(x: any) { A.B.foo; await x; A.B.foo = 2; }`,
      errors: [{ messageId: 'nonAtomicObjectUpdate' }],
    },
    // TS type wrappers on simple LHS report as nonAtomicObjectUpdate
    {
      code: `let foo: any; function cb() { return foo; } async function f(x: any) { foo; await x; (foo as any) = 1; foo = 2; }`,
      errors: [
        { messageId: 'nonAtomicObjectUpdate' },
        { messageId: 'nonAtomicUpdate' },
      ],
    },
    // Block-scoped function declaration shadow
    {
      code: `let foo; function cb() { return foo; } async function f(x) { foo; await x; { function foo() {} foo(); } foo = 1; }`,
      errors: [{ messageId: 'nonAtomicUpdate' }],
    },
    // for-loop let shadow: outer write after loop reports
    {
      code: `let foo; function cb() { return foo; } async function f(x) { foo; await x; for (let foo = 0; foo < 1; foo++) {} foo = 1; }`,
      errors: [{ messageId: 'nonAtomicUpdate' }],
    },
    // Element-access index on LHS of simple `=` is still read
    {
      code: `let x; let foo = {}; function cb() { return x; } async function f(bar) { foo[x] = await bar; x = 1; }`,
      errors: [{ messageId: 'nonAtomicUpdate' }],
    },
    // Non-null assertion on LHS breaks write-target chain
    {
      code: `let foo; function cb() { return foo; } async function f(bar) { foo!.bar = await bar; foo = 1; }`,
      errors: [{ messageId: 'nonAtomicUpdate' }],
    },
  ],
});

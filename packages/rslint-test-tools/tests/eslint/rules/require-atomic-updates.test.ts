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
  ],
});

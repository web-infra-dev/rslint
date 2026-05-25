import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('default-case-last', {
  valid: [
    // basic cases
    'switch (foo) {}',
    'switch (foo) { case 1: bar(); break; }',
    'switch (foo) { case 1: break; case 2: break; }',
    'switch (foo) { default: bar(); break; }',
    'switch (foo) { default: }',
    'switch (foo) { case 1: break; default: break; }',
    'switch (foo) { case 1: break; default: }',
    'switch (foo) { case 1: default: break; }',
    'switch (foo) { case 1: default: }',
    'switch (foo) { case 1: break; case 2: break; default: break; }',
    'switch (foo) { case 1: break; case 2: default: break; }',
    'switch (foo) { case 1: case 2: default: }',
    'switch (foo) { case 1: break; }',
    'switch (foo) { case 1: case 2: break; }',
    'switch (foo) { case 1: baz(); break; case 2: quux(); break; default: quuux(); break; }',

    // nested switch — both default last
    'switch (a) { case 1: switch (b) { case 2: break; default: break; } break; default: break; }',

    // switch inside default clause — both valid
    'switch (a) { case 1: break; default: switch (b) { case 2: break; default: break; } }',

    // triple-nested — all default last
    'switch (a) { case 1: switch (b) { case 2: switch (c) { case 3: break; default: break; } break; default: break; } break; default: break; }',

    // switch inside function/arrow/class method — default last
    'function f() { switch (a) { case 1: return 1; default: return 0; } }',
    'const f = () => { switch (a) { case 1: return 1; default: return 0; } }',
    'class C { m() { switch (a) { case 1: break; default: break; } } }',

    // switch inside control structures — default last
    'if (x) { switch (a) { case 1: break; default: break; } }',
    'for (let i = 0; i < 10; i++) { switch (a) { case 1: break; default: break; } }',
    'while (x) { switch (a) { case 1: break; default: break; } }',
    'try { switch (a) { case 1: break; default: break; } } catch (e) { switch (a) { case 1: break; default: break; } }',

    // switch inside IIFE and labeled statement — default last
    '(function() { switch (a) { case 1: break; default: break; } })()',
    'label: switch (a) { case 1: break; default: break; }',

    // many fall-through cases then default
    'switch (a) { case 1: case 2: case 3: console.log("x"); break; default: break; }',

    // do-while / for-in / for-of — default last
    'do { switch (a) { case 1: break; default: break; } } while (false)',
    'for (const k in obj) { switch (k) { case "a": break; default: break; } }',
    'for (const v of arr) { switch (v) { case 1: break; default: break; } }',

    // getter / setter / static block — default last
    'class G { get x() { switch (a) { case 1: return 1; default: return 0; } } }',
    'class S { set x(v: any) { switch (v) { case 1: break; default: break; } } }',
    'class SB { static { switch (a) { case 1: break; default: break; } } }',

    // generator / async function — default last
    'function* gen() { switch (a) { case 1: yield 1; break; default: yield 0; } }',
    'async function af() { switch (a) { case 1: break; default: break; } }',

    // finally block — default last
    'try {} finally { switch (a) { case 1: break; default: break; } }',

    // TypeScript namespace — default last
    'namespace NS { switch (a) { case 1: break; default: break; } }',
  ],
  invalid: [
    // basic cases
    {
      code: 'switch (foo) { default: bar(); break; case 1: baz(); break; }',
      errors: [{ messageId: 'notLast' }],
    },
    {
      code: 'switch (foo) { default: break; case 1: break; }',
      errors: [{ messageId: 'notLast' }],
    },
    {
      code: 'switch (foo) { default: case 1: break; }',
      errors: [{ messageId: 'notLast' }],
    },
    {
      code: 'switch (foo) { default: case 1: }',
      errors: [{ messageId: 'notLast' }],
    },
    {
      code: 'switch (foo) { case 1: break; default: break; case 2: break; }',
      errors: [{ messageId: 'notLast' }],
    },
    {
      code: 'switch (foo) { case 1: default: break; case 2: break; }',
      errors: [{ messageId: 'notLast' }],
    },
    {
      code: 'switch (foo) { case 1: default: case 2: break; }',
      errors: [{ messageId: 'notLast' }],
    },
    {
      code: 'switch (foo) { case 1: default: case 2: }',
      errors: [{ messageId: 'notLast' }],
    },
    {
      code: 'switch (foo) { default: break; case 1: break; case 2: break; }',
      errors: [{ messageId: 'notLast' }],
    },
    {
      code: 'switch (foo) { default: case 1: case 2: }',
      errors: [{ messageId: 'notLast' }],
    },

    // nested switch — outer default not last, inner valid (only outer reports)
    {
      code: 'switch (a) { default: switch (b) { case 2: break; default: break; } break; case 1: break; }',
      errors: [{ messageId: 'notLast' }],
    },

    // nested switch — outer valid, inner default not last (only inner reports)
    {
      code: 'switch (a) { case 1: switch (b) { default: break; case 2: break; } break; default: break; }',
      errors: [{ messageId: 'notLast' }],
    },

    // nested switch — both default not last (both report)
    {
      code: 'switch (a) { default: switch (b) { default: break; case 2: break; } break; case 1: break; }',
      errors: [{ messageId: 'notLast' }, { messageId: 'notLast' }],
    },

    // triple-nested — only innermost invalid
    {
      code: 'switch (a) { case 1: switch (b) { case 2: switch (c) { default: break; case 3: break; } break; default: break; } break; default: break; }',
      errors: [{ messageId: 'notLast' }],
    },

    // triple-nested — all three invalid
    {
      code: 'switch (a) { default: switch (b) { default: switch (c) { default: break; case 3: break; } break; case 2: break; } break; case 1: break; }',
      errors: [
        { messageId: 'notLast' },
        { messageId: 'notLast' },
        { messageId: 'notLast' },
      ],
    },

    // switch inside function — invalid
    {
      code: 'function f() { switch (a) { default: return 0; case 1: return 1; } }',
      errors: [{ messageId: 'notLast' }],
    },

    // switch inside arrow function — invalid
    {
      code: 'const f = () => { switch (a) { default: return 0; case 1: return 1; } }',
      errors: [{ messageId: 'notLast' }],
    },

    // switch inside class method — invalid
    {
      code: 'class C { m() { switch (a) { default: break; case 1: break; } } }',
      errors: [{ messageId: 'notLast' }],
    },

    // switch inside if — invalid
    {
      code: 'if (x) { switch (a) { default: break; case 1: break; } }',
      errors: [{ messageId: 'notLast' }],
    },

    // switch inside for loop — invalid
    {
      code: 'for (let i = 0; i < 10; i++) { switch (a) { default: break; case 1: break; } }',
      errors: [{ messageId: 'notLast' }],
    },

    // switch inside while — invalid
    {
      code: 'while (x) { switch (a) { default: break; case 1: break; } }',
      errors: [{ messageId: 'notLast' }],
    },

    // switch inside try/catch — both invalid
    {
      code: 'try { switch (a) { default: break; case 1: break; } } catch (e) { switch (a) { default: break; case 1: break; } }',
      errors: [{ messageId: 'notLast' }, { messageId: 'notLast' }],
    },

    // switch inside IIFE — invalid
    {
      code: '(function() { switch (a) { default: break; case 1: break; } })()',
      errors: [{ messageId: 'notLast' }],
    },

    // labeled switch — invalid
    {
      code: 'label: switch (a) { default: break; case 1: break; }',
      errors: [{ messageId: 'notLast' }],
    },

    // switch inside default clause of outer — inner invalid, outer valid
    {
      code: 'switch (a) { case 1: break; default: switch (b) { default: break; case 2: break; } }',
      errors: [{ messageId: 'notLast' }],
    },

    // multiple invalid switches in same block
    {
      code: 'switch (x) { default: break; case 1: break; } switch (y) { default: break; case 2: break; }',
      errors: [{ messageId: 'notLast' }, { messageId: 'notLast' }],
    },

    // default with throw not last
    {
      code: 'function f() { switch (a) { default: throw new Error("no"); case 1: return 1; } }',
      errors: [{ messageId: 'notLast' }],
    },

    // empty default not last
    {
      code: 'switch (a) { default: case 1: break; }',
      errors: [{ messageId: 'notLast' }],
    },

    // default with fall-through not last
    {
      code: 'switch (a) { case 1: default: case 2: case 3: break; }',
      errors: [{ messageId: 'notLast' }],
    },

    // deeply nested in complex structure
    {
      code: 'function f() { for (let i = 0; i < 10; i++) { if (i > 5) { try { switch (a) { default: break; case 1: break; } } catch (e) {} } } }',
      errors: [{ messageId: 'notLast' }],
    },

    // do-while — invalid
    {
      code: 'do { switch (a) { default: break; case 1: break; } } while (false)',
      errors: [{ messageId: 'notLast' }],
    },

    // for-in — invalid
    {
      code: 'for (const k in obj) { switch (k) { default: break; case "a": break; } }',
      errors: [{ messageId: 'notLast' }],
    },

    // for-of — invalid
    {
      code: 'for (const v of arr) { switch (v) { default: break; case 1: break; } }',
      errors: [{ messageId: 'notLast' }],
    },

    // getter — invalid
    {
      code: 'class G { get x() { switch (a) { default: return 1; case 1: return 0; } } }',
      errors: [{ messageId: 'notLast' }],
    },

    // setter — invalid
    {
      code: 'class S { set x(v: any) { switch (v) { default: break; case 1: break; } } }',
      errors: [{ messageId: 'notLast' }],
    },

    // static block — invalid
    {
      code: 'class SB { static { switch (a) { default: break; case 1: break; } } }',
      errors: [{ messageId: 'notLast' }],
    },

    // generator function — invalid
    {
      code: 'function* gen() { switch (a) { default: yield 0; break; case 1: yield 1; } }',
      errors: [{ messageId: 'notLast' }],
    },

    // async function — invalid
    {
      code: 'async function af() { switch (a) { default: break; case 1: break; } }',
      errors: [{ messageId: 'notLast' }],
    },

    // finally block — invalid
    {
      code: 'try {} finally { switch (a) { default: break; case 1: break; } }',
      errors: [{ messageId: 'notLast' }],
    },

    // TypeScript namespace — invalid
    {
      code: 'namespace NS { switch (a) { default: break; case 1: break; } }',
      errors: [{ messageId: 'notLast' }],
    },
  ],
});

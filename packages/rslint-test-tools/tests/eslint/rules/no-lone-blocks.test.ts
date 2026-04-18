import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-lone-blocks', {
  valid: [
    // Blocks that belong to a containing statement (not lone).
    'if (foo) { if (bar) { baz(); } }',
    'if (foo) { bar(); } else { baz(); }',
    'if (foo) { bar(); } else if (baz) { qux(); }',
    'do { bar(); } while (foo)',
    'while (foo) { bar(); }',
    'for (let i = 0; i < 10; i++) { bar(); }',
    'for (const x of xs) { bar(); }',
    'for (const x in obj) { bar(); }',
    'async function f() { for await (const x of xs) { bar(); } }',
    'try { foo(); } catch (e) { bar(); }',
    'try { foo(); } catch { bar(); }',
    'try { foo(); } finally { bar(); }',
    'function foo() { while (bar) { baz(); } }',
    'const f = () => { foo(); }',
    'const f = async () => { await foo(); }',
    'class C { method() { foo(); } }',
    'class C { constructor() { foo(); } }',
    'class C { get prop() { return 1; } }',
    'class C { set prop(v) { this._v = v; } }',
    'class C { static method() { foo(); } }',
    'function* gen() { yield 1; }',

    // Block-level bindings justify a lone block.
    '{ let x = 1; }',
    '{ const x = 1; }',
    '{ class Bar {} }',
    "'use strict'; { function bar() {} }",
    'export {}; { function bar() {} }',
    '{ let x; var y; }',
    '{ var x; let y; }',
    '{ let x; const y = 1; class Z {} }',

    // Nested lone block, each with its own binding.
    '{ {let y = 1;} let x = 1; }',
    '{ let x = 1; { let y = 2; } }',

    // Different scopes, same name is fine.
    'function foo() { { const x = 4 } const x = 3 }',

    // Switch: solo block per clause is allowed.
    `
switch (foo) {
    case bar: {
        baz;
    }
}
`,
    `
switch (foo) {
    case bar: {
        baz;
    }
    case qux: {
        boop;
    }
}
`,
    `
switch (foo) {
    case bar:
    {
        baz;
    }
}
`,
    `
switch (foo) {
    default: {
        baz;
    }
}
`,
    `
switch (foo) {
    case bar: {
        a;
    }
    default: {
        b;
    }
}
`,

    // Class static blocks.
    'class C { static {} }',
    'class C { static { foo; } }',
    'class C { static { foo; bar; } }',
    'class C { static { if (foo) { block; } } }',
    'class C { static { lbl: { block; } } }',
    'class C { static { { let block; } something; } }',
    'class C { static { something; { const block = 1; } } }',
    'class C { static { { function block(){} } something; } }',
    'class C { static { something; { class block {} } } }',

    // Labeled block at program scope.
    'lbl: { foo(); }',
    'lbl: { let x = 1; }',

    // TS namespace — a block inside a namespace body is not flagged (ESLint parity).
    'namespace N { { foo; } }',
    'namespace N { { let x = 1; } }',

    // `using` / `await using` declarations.
    `
{
    using x = makeDisposable();
}
`,
    `
async function f() {
    {
        await using x = makeDisposable();
    }
    bar();
}
`,
  ],
  invalid: [
    // Trivial program-scope lone blocks.
    {
      code: '{}',
      errors: [{ messageId: 'redundantBlock' }],
    },
    {
      code: '{var x = 1;}',
      errors: [{ messageId: 'redundantBlock' }],
    },
    {
      code: 'foo(); {} bar();',
      errors: [{ messageId: 'redundantBlock' }],
    },

    // Two sibling lone blocks at program scope.
    {
      code: '{} {}',
      errors: [
        { messageId: 'redundantBlock' },
        { messageId: 'redundantBlock' },
      ],
    },

    // Lone block nested inside a containing statement body.
    {
      code: 'if (foo) { bar(); {} baz(); }',
      errors: [{ messageId: 'redundantNestedBlock' }],
    },
    {
      code: 'function foo() { bar(); {} baz(); }',
      errors: [{ messageId: 'redundantNestedBlock' }],
    },
    {
      code: 'while (foo) { {} }',
      errors: [{ messageId: 'redundantNestedBlock' }],
    },
    {
      code: 'for (;;) { {} }',
      errors: [{ messageId: 'redundantNestedBlock' }],
    },
    {
      code: 'do { {} } while (foo)',
      errors: [{ messageId: 'redundantNestedBlock' }],
    },
    {
      code: 'try { {} } catch {}',
      errors: [{ messageId: 'redundantNestedBlock' }],
    },
    {
      code: 'try {} catch { {} }',
      errors: [{ messageId: 'redundantNestedBlock' }],
    },
    {
      code: 'try {} finally { {} }',
      errors: [{ messageId: 'redundantNestedBlock' }],
    },
    {
      code: 'class C { method() { {} } }',
      errors: [{ messageId: 'redundantNestedBlock' }],
    },
    {
      code: 'class C { constructor() { {} } }',
      errors: [{ messageId: 'redundantNestedBlock' }],
    },
    {
      code: 'const f = () => { {} };',
      errors: [{ messageId: 'redundantNestedBlock' }],
    },

    // Multi-line lone block (position captured by snapshot).
    {
      code: '{ \n{ } }',
      errors: [
        { messageId: 'redundantBlock' },
        { messageId: 'redundantNestedBlock' },
      ],
    },
    {
      code: '{\n    var x = 1;\n}',
      errors: [{ messageId: 'redundantBlock' }],
    },

    // Triple nesting.
    {
      code: '{ { {} } }',
      errors: [
        { messageId: 'redundantBlock' },
        { messageId: 'redundantNestedBlock' },
        { messageId: 'redundantNestedBlock' },
      ],
    },

    // Non-block-level binding does not justify the block.
    {
      code: '{ function bar() {} }',
      errors: [{ messageId: 'redundantBlock' }],
    },

    // Block containing only a non-binding is still redundant.
    {
      code: '{ lbl: foo(); }',
      errors: [{ messageId: 'redundantBlock' }],
    },
    {
      code: '{ type X = number; }',
      errors: [{ messageId: 'redundantBlock' }],
    },
    {
      code: '{ interface I {} }',
      errors: [{ messageId: 'redundantBlock' }],
    },
    {
      code: '{ declare var x: number; }',
      errors: [{ messageId: 'redundantBlock' }],
    },
    {
      code: '{ /* comment */ }',
      errors: [{ messageId: 'redundantBlock' }],
    },

    // Stack semantics: only bindings directly inside a block mark it.
    {
      code: '{ \n{var x = 1;}\n let y = 2; } {let z = 1;}',
      errors: [{ messageId: 'redundantNestedBlock' }],
    },
    {
      code: '{ \n{let x = 1;}\n var y = 2; } {let z = 1;}',
      errors: [{ messageId: 'redundantBlock' }],
    },
    {
      code: '{ \n{var x = 1;}\n var y = 2; }\n {var z = 1;}',
      errors: [
        { messageId: 'redundantBlock' },
        { messageId: 'redundantNestedBlock' },
        { messageId: 'redundantBlock' },
      ],
    },

    // Switch: block that is not the sole statement of a case/default clause.
    {
      code: `
switch (foo) {
    case 1:
        foo();
        {
            bar;
        }
}
`,
      errors: [{ messageId: 'redundantBlock' }],
    },
    {
      code: `
switch (foo) {
    case 1:
    {
        bar;
    }
    foo();
}
`,
      errors: [{ messageId: 'redundantBlock' }],
    },
    {
      code: `
switch (foo) {
    default:
        foo();
        {
            bar;
        }
}
`,
      errors: [{ messageId: 'redundantBlock' }],
    },
    {
      code: `
switch (foo) {
    default:
    {
        bar;
    }
    foo();
}
`,
      errors: [{ messageId: 'redundantBlock' }],
    },

    // Function body containing a single lone block (else-if branch fires).
    {
      code: `
function foo () {
    {
        const x = 4;
    }
}
`,
      errors: [{ messageId: 'redundantNestedBlock' }],
    },
    {
      code: `
function foo () {
    {
        var x = 4;
    }
}
`,
      errors: [{ messageId: 'redundantNestedBlock' }],
    },

    // Class static block cases.
    {
      code: `
class C {
    static {
        if (foo) {
            {
                let block;
            }
        }
    }
}
`,
      errors: [{ messageId: 'redundantNestedBlock' }],
    },
    {
      code: `
class C {
    static {
        if (foo) {
            {
                block;
            }
            something;
        }
    }
}
`,
      errors: [{ messageId: 'redundantNestedBlock' }],
    },
    {
      code: `
class C {
    static {
        {
            block;
        }
    }
}
`,
      errors: [{ messageId: 'redundantNestedBlock' }],
    },
    {
      code: `
class C {
    static {
        {
            let block;
        }
    }
}
`,
      errors: [{ messageId: 'redundantNestedBlock' }],
    },
    {
      code: `
class C {
    static {
        {
            const block = 1;
        }
    }
}
`,
      errors: [{ messageId: 'redundantNestedBlock' }],
    },
    {
      code: `
class C {
    static {
        {
            function block() {}
        }
    }
}
`,
      errors: [{ messageId: 'redundantNestedBlock' }],
    },
    {
      code: `
class C {
    static {
        {
            class block {}
        }
    }
}
`,
      errors: [{ messageId: 'redundantNestedBlock' }],
    },
    {
      code: `
class C {
    static {
        {
            var block;
        }
        something;
    }
}
`,
      errors: [{ messageId: 'redundantNestedBlock' }],
    },
    {
      code: `
class C {
    static {
        something;
        {
            var block;
        }
    }
}
`,
      errors: [{ messageId: 'redundantNestedBlock' }],
    },
    {
      code: `
class C {
    static {
        {
            block;
        }
        something;
    }
}
`,
      errors: [{ messageId: 'redundantNestedBlock' }],
    },
    {
      code: `
class C {
    static {
        something;
        {
            block;
        }
    }
}
`,
      errors: [{ messageId: 'redundantNestedBlock' }],
    },
  ],
});

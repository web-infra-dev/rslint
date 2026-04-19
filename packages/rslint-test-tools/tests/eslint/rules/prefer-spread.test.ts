import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('prefer-spread', {
  valid: [
    'foo.apply(obj, args);',
    'obj.foo.apply(null, args);',
    'obj.foo.apply(otherObj, args);',
    'a.b(x, y).c.foo.apply(a.b(x, z).c, args);',
    'a.b.foo.apply(a.b.c, args);',
    'foo.apply(undefined, [1, 2]);',
    'foo.apply(null, [1, 2]);',
    'obj.foo.apply(obj, [1, 2]);',
    'var apply; foo[apply](null, args);',
    'foo.apply();',
    'obj.foo.apply();',
    'obj.foo.apply(obj, ...args);',
    // `(a?.b).c` has extra parens around `a?.b`, `a?.b.c` does not — tokens differ
    '(a?.b).c.foo.apply(a?.b.c, args);',
    'a?.b.c.foo.apply((a?.b).c, args);',
    // Private identifier named `#apply` — the member access does not resolve
    // to the method name "apply"
    'class C { #apply; foo() { foo.#apply(undefined, args); } }',

    // ---- Real-world patterns beyond the ESLint suite ----
    'obj.foo.apply(this, args);',
    'class C { m(args: any) { this.foo.apply(that, args); } }',
    'class C extends B { m(args: any) { super.foo.apply(this, args); } }',
    '(obj as any).foo.apply(obj, args);',
    '[1, 2].concat.apply([1, 3], args);',
    // Hex vs decimal — ESLint equalTokens keeps them distinct
    '[0x1].concat.apply([1], args);',
    '[a,].concat.apply([a], args);',
    'outer(inner(x)).m.apply(outer(inner(y)).m, args);',
    // Non-static computed key
    'foo[getKey()](null, args);',
    // Cross-class negatives
    'foo["#apply"](null, args);',
    'foo.bind(null, args);',
    'foo.APPLY(null, args);',
  ],
  invalid: [
    {
      code: 'foo.apply(undefined, args);',
      errors: [{ messageId: 'preferSpread' }],
    },
    {
      code: 'foo.apply(void 0, args);',
      errors: [{ messageId: 'preferSpread' }],
    },
    {
      code: 'foo.apply(null, args);',
      errors: [{ messageId: 'preferSpread' }],
    },
    {
      code: 'obj.foo.apply(obj, args);',
      errors: [{ messageId: 'preferSpread' }],
    },
    {
      code: 'a.b.c.foo.apply(a.b.c, args);',
      errors: [{ messageId: 'preferSpread' }],
    },
    {
      code: 'a.b(x, y).c.foo.apply(a.b(x, y).c, args);',
      errors: [{ messageId: 'preferSpread' }],
    },
    {
      code: '[].concat.apply([ ], args);',
      errors: [{ messageId: 'preferSpread' }],
    },
    {
      code: '[].concat.apply([\n/*empty*/\n], args);',
      errors: [{ messageId: 'preferSpread' }],
    },
    // Optional chaining variants
    {
      code: 'foo.apply?.(undefined, args);',
      errors: [{ messageId: 'preferSpread' }],
    },
    {
      code: 'foo?.apply(undefined, args);',
      errors: [{ messageId: 'preferSpread' }],
    },
    {
      code: 'foo?.apply?.(undefined, args);',
      errors: [{ messageId: 'preferSpread' }],
    },
    {
      code: '(foo?.apply)(undefined, args);',
      errors: [{ messageId: 'preferSpread' }],
    },
    {
      code: '(foo?.apply)?.(undefined, args);',
      errors: [{ messageId: 'preferSpread' }],
    },
    {
      code: '(obj?.foo).apply(obj, args);',
      errors: [{ messageId: 'preferSpread' }],
    },
    {
      code: 'a?.b.c.foo.apply(a?.b.c, args);',
      errors: [{ messageId: 'preferSpread' }],
    },
    {
      code: '(a?.b.c).foo.apply(a?.b.c, args);',
      errors: [{ messageId: 'preferSpread' }],
    },
    {
      code: '(a?.b).c.foo.apply((a?.b).c, args);',
      errors: [{ messageId: 'preferSpread' }],
    },
    {
      code: 'class C { #foo; foo() { obj.#foo.apply(obj, args); } }',
      errors: [{ messageId: 'preferSpread' }],
    },

    // ---- Real-world patterns beyond the ESLint suite ----
    {
      code: 'class C { m(args: any) { this.foo.apply(this, args); } }',
      errors: [{ messageId: 'preferSpread' }],
    },
    {
      code: 'a.b().c.foo.apply(a.b().c, args);',
      errors: [{ messageId: 'preferSpread' }],
    },
    {
      code: 'getFn().apply(undefined, args);',
      errors: [{ messageId: 'preferSpread' }],
    },
    {
      code: 'foo["apply"](null, args);',
      errors: [{ messageId: 'preferSpread' }],
    },
    {
      code: 'obj["foo"].apply(obj, args);',
      errors: [{ messageId: 'preferSpread' }],
    },
    {
      code: 'foo[`apply`](null, args);',
      errors: [{ messageId: 'preferSpread' }],
    },
    {
      code: 'obj\n  .foo\n  .apply(obj, args);',
      errors: [{ messageId: 'preferSpread' }],
    },
    {
      code: 'obj /* x */ . foo . apply(obj, args);',
      errors: [{ messageId: 'preferSpread' }],
    },
    {
      code: 'foo.apply<any>(undefined, args);',
      errors: [{ messageId: 'preferSpread' }],
    },
    {
      code: '[1, 2].concat.apply([1, 2], args);',
      errors: [{ messageId: 'preferSpread' }],
    },
    {
      code: '[a,].concat.apply([a,], args);',
      errors: [{ messageId: 'preferSpread' }],
    },
    {
      code: '[].concat.apply([\n/* comment */\n], args);',
      errors: [{ messageId: 'preferSpread' }],
    },
    {
      code: '(new Foo()).bar.apply(new Foo(), args);',
      errors: [{ messageId: 'preferSpread' }],
    },
    {
      code: 'wrap(foo.apply(null, args));',
      errors: [{ messageId: 'preferSpread' }],
    },
    {
      code: 'outer(inner(x)).m.apply(outer(inner(x)), args);',
      errors: [{ messageId: 'preferSpread' }],
    },
  ],
});

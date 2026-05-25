import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-useless-call', {
  valid: [
    // `this` binding is different.
    'foo.apply(obj, 1, 2);',
    'obj.foo.apply(null, 1, 2);',
    'obj.foo.apply(otherObj, 1, 2);',
    'a.b(x, y).c.foo.apply(a.b(x, z).c, 1, 2);',
    'foo.apply(obj, [1, 2]);',
    'obj.foo.apply(null, [1, 2]);',
    'obj.foo.apply(otherObj, [1, 2]);',
    'a.b(x, y).c.foo.apply(a.b(x, z).c, [1, 2]);',
    'a.b.foo.apply(a.b.c, [1, 2]);',

    // ignores variadic.
    'foo.apply(null, args);',
    'obj.foo.apply(obj, args);',

    // ignores computed property.
    'var call; foo[call](null, 1, 2);',
    'var apply; foo[apply](null, [1, 2]);',

    // ignores incomplete things.
    'foo.call();',
    'obj.foo.call();',
    'foo.apply();',
    'obj.foo.apply();',

    // Optional chaining where receiver shape differs from thisArg.
    'obj?.foo.bar.call(obj.foo, 1, 2);',

    // TS-specific receiver shapes — token streams differ from thisArg.
    '(obj as any).foo.call(obj, 1, 2);',
    'obj!.foo.call(obj, 1, 2);',
    '(obj satisfies any).foo.call(obj, 1, 2);',
    'obj.foo.call<number>(other, 1, 2);',

    // Spread / non-object first argument.
    'foo.call(...args);',
    'foo.call(0, 1, 2);',
    'foo.call.call(other, 1, 2);',
  ],
  invalid: [
    // call.
    {
      code: 'foo.call(undefined, 1, 2);',
      errors: [{ messageId: 'unnecessaryCall' }],
    },
    {
      code: 'foo.call(void 0, 1, 2);',
      errors: [{ messageId: 'unnecessaryCall' }],
    },
    {
      code: 'foo.call(null, 1, 2);',
      errors: [{ messageId: 'unnecessaryCall' }],
    },
    {
      code: 'obj.foo.call(obj, 1, 2);',
      errors: [{ messageId: 'unnecessaryCall' }],
    },
    {
      code: 'a.b.c.foo.call(a.b.c, 1, 2);',
      errors: [{ messageId: 'unnecessaryCall' }],
    },
    {
      code: 'a.b(x, y).c.foo.call(a.b(x, y).c, 1, 2);',
      errors: [{ messageId: 'unnecessaryCall' }],
    },

    // apply.
    {
      code: 'foo.apply(undefined, [1, 2]);',
      errors: [{ messageId: 'unnecessaryCall' }],
    },
    {
      code: 'foo.apply(void 0, [1, 2]);',
      errors: [{ messageId: 'unnecessaryCall' }],
    },
    {
      code: 'foo.apply(null, [1, 2]);',
      errors: [{ messageId: 'unnecessaryCall' }],
    },
    {
      code: 'obj.foo.apply(obj, [1, 2]);',
      errors: [{ messageId: 'unnecessaryCall' }],
    },
    {
      code: 'a.b.c.foo.apply(a.b.c, [1, 2]);',
      errors: [{ messageId: 'unnecessaryCall' }],
    },
    {
      code: 'a.b(x, y).c.foo.apply(a.b(x, y).c, [1, 2]);',
      errors: [{ messageId: 'unnecessaryCall' }],
    },
    {
      code: '[].concat.apply([ ], [1, 2]);',
      errors: [{ messageId: 'unnecessaryCall' }],
    },
    {
      code: '[].concat.apply([\n/*empty*/\n], [1, 2]);',
      errors: [{ messageId: 'unnecessaryCall' }],
    },
    {
      code: 'abc.get("foo", 0).concat.apply(abc . get("foo",  0 ), [1, 2]);',
      errors: [{ messageId: 'unnecessaryCall' }],
    },

    // Optional chaining.
    {
      code: 'foo.call?.(undefined, 1, 2);',
      errors: [{ messageId: 'unnecessaryCall' }],
    },
    {
      code: 'foo?.call(undefined, 1, 2);',
      errors: [{ messageId: 'unnecessaryCall' }],
    },
    {
      code: '(foo?.call)(undefined, 1, 2);',
      errors: [{ messageId: 'unnecessaryCall' }],
    },
    {
      code: 'obj.foo.call?.(obj, 1, 2);',
      errors: [{ messageId: 'unnecessaryCall' }],
    },
    {
      code: 'obj?.foo.call(obj, 1, 2);',
      errors: [{ messageId: 'unnecessaryCall' }],
    },
    {
      code: '(obj?.foo).call(obj, 1, 2);',
      errors: [{ messageId: 'unnecessaryCall' }],
    },
    {
      code: '(obj?.foo.call)(obj, 1, 2);',
      errors: [{ messageId: 'unnecessaryCall' }],
    },
    {
      code: 'obj?.foo.bar.call(obj?.foo, 1, 2);',
      errors: [{ messageId: 'unnecessaryCall' }],
    },
    {
      code: '(obj?.foo).bar.call(obj?.foo, 1, 2);',
      errors: [{ messageId: 'unnecessaryCall' }],
    },
    {
      code: 'obj.foo?.bar.call(obj.foo, 1, 2);',
      errors: [{ messageId: 'unnecessaryCall' }],
    },

    // ElementAccessExpression as the applied function.
    {
      code: "obj['foo'].call(obj, 1, 2);",
      errors: [{ messageId: 'unnecessaryCall' }],
    },
    {
      code: "obj['foo'].apply(obj, [1, 2]);",
      errors: [{ messageId: 'unnecessaryCall' }],
    },

    // `this` keyword as receiver/thisArg.
    {
      code: 'class C { run() { this.foo.call(this, 1, 2); } }',
      errors: [{ messageId: 'unnecessaryCall' }],
    },
    {
      code: 'class C { run() { this.foo.apply(this, [1, 2]); } }',
      errors: [{ messageId: 'unnecessaryCall' }],
    },

    // Direct function/arrow call via .call/.apply.
    {
      code: '(() => 1).call(undefined);',
      errors: [{ messageId: 'unnecessaryCall' }],
    },
    {
      code: '(function () {}).apply(null, []);',
      errors: [{ messageId: 'unnecessaryCall' }],
    },

    // Extra edge cases.
    {
      code: 'foo.call((null), 1, 2);',
      errors: [{ messageId: 'unnecessaryCall' }],
    },
    {
      code: 'foo.apply((undefined), [1, 2]);',
      errors: [{ messageId: 'unnecessaryCall' }],
    },
    {
      code: '[].concat.apply([/*c*/], [1, 2]);',
      errors: [{ messageId: 'unnecessaryCall' }],
    },
    {
      code: '({}).valueOf.call({ }, );',
      errors: [{ messageId: 'unnecessaryCall' }],
    },
    {
      code: 'foo.apply(null, [...args]);',
      errors: [{ messageId: 'unnecessaryCall' }],
    },
  ],
});

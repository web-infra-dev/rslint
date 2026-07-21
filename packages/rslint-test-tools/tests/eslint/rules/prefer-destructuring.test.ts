import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('prefer-destructuring', {
  valid: [
    // Basic destructuring and uninitialized declarations
    'var [foo] = array;',
    'var { foo } = object;',
    'var foo;',
    // Renamed-property defaults and options
    {
      code: 'var foo = object.bar;',
      options: [{ VariableDeclarator: { object: true } }] as any,
    },
    { code: 'var foo = object.bar;', options: [{ object: true }] as any },
    {
      code: 'var foo = object.bar;',
      options: [
        { VariableDeclarator: { object: true } },
        { enforceForRenamedProperties: false },
      ] as any,
    },
    {
      code: 'var foo = object.bar;',
      options: [
        { object: true },
        { enforceForRenamedProperties: false },
      ] as any,
    },
    {
      code: "var foo = object['bar'];",
      options: [
        { VariableDeclarator: { object: true } },
        { enforceForRenamedProperties: false },
      ] as any,
    },
    {
      code: 'var foo = object[bar];',
      options: [
        { object: true },
        { enforceForRenamedProperties: false },
      ] as any,
    },
    {
      code: 'var { bar: foo } = object;',
      options: [
        { VariableDeclarator: { object: true } },
        { enforceForRenamedProperties: true },
      ] as any,
    },
    {
      code: 'var { bar: foo } = object;',
      options: [{ object: true }, { enforceForRenamedProperties: true }] as any,
    },
    {
      code: 'var { [bar]: foo } = object;',
      options: [
        { VariableDeclarator: { object: true } },
        { enforceForRenamedProperties: true },
      ] as any,
    },
    {
      code: 'var { [bar]: foo } = object;',
      options: [{ object: true }, { enforceForRenamedProperties: true }] as any,
    },
    // Per-kind enablement
    {
      code: 'var foo = array[0];',
      options: [{ VariableDeclarator: { array: false } }] as any,
    },
    { code: 'var foo = array[0];', options: [{ array: false }] as any },
    {
      code: 'var foo = object.foo;',
      options: [{ VariableDeclarator: { object: false } }] as any,
    },
    {
      code: "var foo = object['foo'];",
      options: [{ VariableDeclarator: { object: false } }] as any,
    },
    '({ foo } = object);',
    // Regression #8654
    {
      code: 'var foo = array[0];',
      options: [
        { VariableDeclarator: { array: false } },
        { enforceForRenamedProperties: true },
      ] as any,
    },
    {
      code: 'var foo = array[0];',
      options: [{ array: false }, { enforceForRenamedProperties: true }] as any,
    },
    // Assignment expressions and compound assignments
    '[foo] = array;',
    'foo += array[0]',
    { code: 'foo &&= array[0]', languageOptions: { ecmaVersion: 2021 } as any },
    'foo += bar.foo',
    { code: 'foo ||= bar.foo', languageOptions: { ecmaVersion: 2021 } as any },
    {
      code: "foo ??= bar['foo']",
      languageOptions: { ecmaVersion: 2021 } as any,
    },
    // Per-node-type assignment/declaration options
    {
      code: 'foo = object.foo;',
      options: [
        { AssignmentExpression: { object: false } },
        { enforceForRenamedProperties: true },
      ] as any,
    },
    {
      code: 'foo = object.foo;',
      options: [
        { AssignmentExpression: { object: false } },
        { enforceForRenamedProperties: false },
      ] as any,
    },
    {
      code: 'foo = array[0];',
      options: [
        { AssignmentExpression: { array: false } },
        { enforceForRenamedProperties: true },
      ] as any,
    },
    {
      code: 'foo = array[0];',
      options: [
        { AssignmentExpression: { array: false } },
        { enforceForRenamedProperties: false },
      ] as any,
    },
    {
      code: 'foo = array[0];',
      options: [
        {
          VariableDeclarator: { array: true },
          AssignmentExpression: { array: false },
        },
        { enforceForRenamedProperties: false },
      ] as any,
    },
    {
      code: 'var foo = array[0];',
      options: [
        {
          VariableDeclarator: { array: false },
          AssignmentExpression: { array: true },
        },
        { enforceForRenamedProperties: false },
      ] as any,
    },
    {
      code: 'foo = object.foo;',
      options: [
        {
          VariableDeclarator: { object: true },
          AssignmentExpression: { object: false },
        },
      ] as any,
    },
    {
      code: 'var foo = object.foo;',
      options: [
        {
          VariableDeclarator: { object: false },
          AssignmentExpression: { object: true },
        },
      ] as any,
    },
    // super, dynamic access, and already-destructured targets
    'class Foo extends Bar { static foo() {var foo = super.foo} }',
    'foo = bar[foo];',
    'var foo = bar[foo];',
    { code: 'var {foo: {bar}} = object;', options: [{ object: true }] as any },
    { code: 'var {bar} = object.foo;', options: [{ object: true }] as any },
    // Optional chaining
    'var foo = array?.[0];',
    'var foo = object?.foo;',
    // Private identifiers
    'class C { #x; foo() { const x = this.#x; } }',
    'class C { #x; foo() { x = this.#x; } }',
    'class C { #x; foo(a) { x = a.#x; } }',
    {
      code: 'class C { #x; foo() { const x = this.#x; } }',
      options: [
        { array: true, object: true },
        { enforceForRenamedProperties: true },
      ] as any,
    },
    {
      code: 'class C { #x; foo() { const y = this.#x; } }',
      options: [
        { array: true, object: true },
        { enforceForRenamedProperties: true },
      ] as any,
    },
    {
      code: 'class C { #x; foo() { x = this.#x; } }',
      options: [
        { array: true, object: true },
        { enforceForRenamedProperties: true },
      ] as any,
    },
    {
      code: 'class C { #x; foo() { y = this.#x; } }',
      options: [
        { array: true, object: true },
        { enforceForRenamedProperties: true },
      ] as any,
    },
    {
      code: 'class C { #x; foo(a) { x = a.#x; } }',
      options: [
        { array: true, object: true },
        { enforceForRenamedProperties: true },
      ] as any,
    },
    {
      code: 'class C { #x; foo(a) { y = a.#x; } }',
      options: [
        { array: true, object: true },
        { enforceForRenamedProperties: true },
      ] as any,
    },
    {
      code: 'class C { #x; foo() { x = this.a.#x; } }',
      options: [
        { array: true, object: true },
        { enforceForRenamedProperties: true },
      ] as any,
    },
    // Explicit resource management
    {
      code: 'using foo = array[0];',
      languageOptions: { sourceType: 'module', ecmaVersion: 2026 } as any,
    },
    {
      code: 'using foo = object.foo;',
      languageOptions: { sourceType: 'module', ecmaVersion: 2026 } as any,
    },
    {
      code: 'await using foo = array[0];',
      languageOptions: { sourceType: 'module', ecmaVersion: 2026 } as any,
    },
    {
      code: 'await using foo = object.foo;',
      languageOptions: { sourceType: 'module', ecmaVersion: 2026 } as any,
    },
  ],
  invalid: [
    // Array and basic object access
    {
      code: 'var foo = array[0];',
      errors: [{ messageId: 'preferDestructuring' }],
    },
    {
      code: 'foo = array[0];',
      errors: [{ messageId: 'preferDestructuring' }],
    },
    {
      code: 'var foo = object.foo;',
      errors: [{ messageId: 'preferDestructuring' }],
    },
    // Autofix receiver precedence
    {
      code: 'var foo = (a, b).foo;',
      errors: [{ messageId: 'preferDestructuring' }],
    },
    {
      code: 'var length = (() => {}).length;',
      errors: [{ messageId: 'preferDestructuring' }],
    },
    {
      code: 'var foo = (a = b).foo;',
      errors: [{ messageId: 'preferDestructuring' }],
    },
    {
      code: 'var foo = (a || b).foo;',
      errors: [{ messageId: 'preferDestructuring' }],
    },
    {
      code: 'var foo = (f()).foo;',
      errors: [{ messageId: 'preferDestructuring' }],
    },
    {
      code: 'var foo = object.bar.foo;',
      errors: [{ messageId: 'preferDestructuring' }],
    },
    // Renamed and computed properties
    {
      code: 'var foobar = object.bar;',
      options: [
        { VariableDeclarator: { object: true } },
        { enforceForRenamedProperties: true },
      ] as any,
      errors: [{ messageId: 'preferDestructuring' }],
    },
    {
      code: 'var foobar = object.bar;',
      options: [{ object: true }, { enforceForRenamedProperties: true }] as any,
      errors: [{ messageId: 'preferDestructuring' }],
    },
    {
      code: 'var foo = object[bar];',
      options: [
        { VariableDeclarator: { object: true } },
        { enforceForRenamedProperties: true },
      ] as any,
      errors: [{ messageId: 'preferDestructuring' }],
    },
    {
      code: 'var foo = object[bar];',
      options: [{ object: true }, { enforceForRenamedProperties: true }] as any,
      errors: [{ messageId: 'preferDestructuring' }],
    },
    {
      code: 'var foo = object[foo];',
      options: [{ object: true }, { enforceForRenamedProperties: true }] as any,
      errors: [{ messageId: 'preferDestructuring' }],
    },
    // Same-name string access and assignments
    {
      code: "var foo = object['foo'];",
      errors: [{ messageId: 'preferDestructuring' }],
    },
    {
      code: 'foo = object.foo;',
      errors: [{ messageId: 'preferDestructuring' }],
    },
    {
      code: "foo = object['foo'];",
      errors: [{ messageId: 'preferDestructuring' }],
    },
    // Per-kind and per-node-type options
    {
      code: 'var foo = array[0];',
      options: [
        { VariableDeclarator: { array: true } },
        { enforceForRenamedProperties: true },
      ] as any,
      errors: [{ messageId: 'preferDestructuring' }],
    },
    {
      code: 'foo = array[0];',
      options: [{ AssignmentExpression: { array: true } }] as any,
      errors: [{ messageId: 'preferDestructuring' }],
    },
    {
      code: 'var foo = array[0];',
      options: [
        {
          VariableDeclarator: { array: true },
          AssignmentExpression: { array: false },
        },
        { enforceForRenamedProperties: true },
      ] as any,
      errors: [{ messageId: 'preferDestructuring' }],
    },
    {
      code: 'var foo = array[0];',
      options: [
        {
          VariableDeclarator: { array: true },
          AssignmentExpression: { array: false },
        },
      ] as any,
      errors: [{ messageId: 'preferDestructuring' }],
    },
    {
      code: 'foo = array[0];',
      options: [
        {
          VariableDeclarator: { array: false },
          AssignmentExpression: { array: true },
        },
      ] as any,
      errors: [{ messageId: 'preferDestructuring' }],
    },
    {
      code: 'foo = object.foo;',
      options: [
        {
          VariableDeclarator: { array: true, object: false },
          AssignmentExpression: { object: true },
        },
      ] as any,
      errors: [{ messageId: 'preferDestructuring' }],
    },
    // Nested super access
    {
      code: 'class Foo extends Bar { static foo() {var bar = super.foo.bar} }',
      errors: [{ messageId: 'preferDestructuring' }],
    },
    // Comments
    {
      code: 'var /* comment */ foo = object.foo;',
      errors: [{ messageId: 'preferDestructuring' }],
    },
    {
      code: 'var a, /* comment */foo = object.foo;',
      errors: [{ messageId: 'preferDestructuring' }],
    },
    {
      code: 'var foo /* comment */ = object.foo;',
      errors: [{ messageId: 'preferDestructuring' }],
    },
    {
      code: 'var a, foo /* comment */ = object.foo;',
      errors: [{ messageId: 'preferDestructuring' }],
    },
    {
      code: 'var foo /* comment */ = object.foo, a;',
      errors: [{ messageId: 'preferDestructuring' }],
    },
    {
      code: 'var foo // comment\n = object.foo;',
      errors: [{ messageId: 'preferDestructuring' }],
    },
    {
      code: 'var foo = /* comment */ object.foo;',
      errors: [{ messageId: 'preferDestructuring' }],
    },
    {
      code: 'var foo = // comment\n object.foo;',
      errors: [{ messageId: 'preferDestructuring' }],
    },
    {
      code: 'var foo = (/* comment */ object).foo;',
      errors: [{ messageId: 'preferDestructuring' }],
    },
    {
      code: 'var foo = (object /* comment */).foo;',
      errors: [{ messageId: 'preferDestructuring' }],
    },
    {
      code: 'var foo = bar(/* comment */).foo;',
      errors: [{ messageId: 'preferDestructuring' }],
    },
    {
      code: 'var foo = bar/* comment */.baz.foo;',
      errors: [{ messageId: 'preferDestructuring' }],
    },
    {
      code: 'var foo = bar[// comment\nbaz].foo;',
      errors: [{ messageId: 'preferDestructuring' }],
    },
    {
      code: 'var foo // comment\n = bar(/* comment */).foo;',
      errors: [{ messageId: 'preferDestructuring' }],
    },
    {
      code: 'var foo = bar/* comment */.baz/* comment */.foo;',
      errors: [{ messageId: 'preferDestructuring' }],
    },
    {
      code: 'var foo = object// comment\n.foo;',
      errors: [{ messageId: 'preferDestructuring' }],
    },
    {
      code: 'var foo = object./* comment */foo;',
      errors: [{ messageId: 'preferDestructuring' }],
    },
    {
      code: 'var foo = (/* comment */ object.foo);',
      errors: [{ messageId: 'preferDestructuring' }],
    },
    {
      code: 'var foo = (object.foo /* comment */);',
      errors: [{ messageId: 'preferDestructuring' }],
    },
    {
      code: 'var foo = object.foo/* comment */;',
      errors: [{ messageId: 'preferDestructuring' }],
    },
    {
      code: 'var foo = object.foo// comment',
      errors: [{ messageId: 'preferDestructuring' }],
    },
    {
      code: 'var foo = object.foo/* comment */, a;',
      errors: [{ messageId: 'preferDestructuring' }],
    },
    {
      code: 'var foo = object.foo// comment\n, a;',
      errors: [{ messageId: 'preferDestructuring' }],
    },
    {
      code: 'var foo = object.foo, /* comment */ a;',
      errors: [{ messageId: 'preferDestructuring' }],
    },
  ],
});

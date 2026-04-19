import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-useless-computed-key', {
  valid: [
    // ---- Object literals ----
    "({ 'a': 0, b(){} })",
    '({ [x]: 0 });',
    '({ a: 0, [b](){} })',
    "({ ['__proto__']: [] })",

    // ---- Object destructuring ----
    "var { 'a': foo } = obj",
    'var { [a]: b } = obj;',
    'var { a } = obj;',
    'var { a: a } = obj;',
    'var { a: b } = obj;',

    // ---- Class members with enforceForClassMembers: true ----
    {
      code: 'class Foo { a() {} }',
      options: [{ enforceForClassMembers: true }] as any,
    },
    {
      code: "class Foo { 'a'() {} }",
      options: [{ enforceForClassMembers: true }] as any,
    },
    {
      code: 'class Foo { [x]() {} }',
      options: [{ enforceForClassMembers: true }] as any,
    },
    {
      code: "class Foo { ['constructor']() {} }",
      options: [{ enforceForClassMembers: true }] as any,
    },
    {
      code: "class Foo { static ['prototype']() {} }",
      options: [{ enforceForClassMembers: true }] as any,
    },
    {
      code: "(class { 'a'() {} })",
      options: [{ enforceForClassMembers: true }] as any,
    },
    {
      code: '(class { [x]() {} })',
      options: [{ enforceForClassMembers: true }] as any,
    },
    {
      code: "(class { ['constructor']() {} })",
      options: [{ enforceForClassMembers: true }] as any,
    },
    {
      code: "(class { static ['prototype']() {} })",
      options: [{ enforceForClassMembers: true }] as any,
    },

    // ---- Class members with default options ----
    "class Foo { 'x'() {} }",
    '(class { [x]() {} })',
    'class Foo { static constructor() {} }',
    'class Foo { prototype() {} }',

    // ---- Class members with enforceForClassMembers: false ----
    {
      code: "class Foo { ['x']() {} }",
      options: [{ enforceForClassMembers: false }] as any,
    },
    {
      code: "(class { ['x']() {} })",
      options: [{ enforceForClassMembers: false }] as any,
    },
    {
      code: "class Foo { static ['constructor']() {} }",
      options: [{ enforceForClassMembers: false }] as any,
    },
    {
      code: "class Foo { ['prototype']() {} }",
      options: [{ enforceForClassMembers: false }] as any,
    },

    // ---- Class fields ----
    {
      code: 'class Foo { a }',
      options: [{ enforceForClassMembers: true }] as any,
    },
    {
      code: "class Foo { ['constructor'] }",
      options: [{ enforceForClassMembers: true }] as any,
    },
    {
      code: "class Foo { static ['constructor'] }",
      options: [{ enforceForClassMembers: true }] as any,
    },
    {
      code: "class Foo { static ['prototype'] }",
      options: [{ enforceForClassMembers: true }] as any,
    },

    // BigInt literals are left alone — browsers throw on bigint property
    // names, so the rule deliberately doesn't touch them.
    '({ [99999999999999999n]: 0 })',

    // Template literals (even no-substitution) are not flagged.
    '({ [`x`]: 0 })',
    'class Foo { [`x`]() {} }',
    // Unary-negated / non-literal expressions are not literals.
    '({ [-1]: 0 })',
    '({ [void 0]: 0 })',
    '({ [/x/]: 0 })',
    "const k = 'x'; const o = { [k]: 1 }",
    'const x = { [Symbol()]: 1 }',

    // Value-position __proto__ in namespace — still carved out.
    "namespace N { export const v = { ['__proto__']: [] } }",

    // TS wrapping the inner expression makes it non-literal.
    "const x = { [('x' as const)]: 1 }",
    "const x = { [('x' satisfies string)]: 1 }",

    // Auto-accessor (TS 5 / ES stage-3) surfaces as AccessorProperty in
    // TSESTree, which the core rule doesn't listen for — match: no report.
    "class Foo { accessor ['x'] = 1 }",
    "class Foo { static accessor ['x'] = 1 }",
    "class Foo { accessor ['constructor'] = 1 }",
    "class Foo { static accessor ['constructor'] = 1 }",
    "class Foo { static accessor ['prototype'] = 1 }",

    // Abstract class members map to TSAbstract* nodes in TSESTree, which
    // the core rule doesn't listen for — match: no report.
    "abstract class Foo { abstract ['x'](): void }",
    "abstract class Foo { abstract ['x']: string }",
    "abstract class Foo { abstract get ['x'](): number }",
    "abstract class Foo { abstract set ['x'](v: number) }",
    "abstract class Foo { abstract readonly ['x']: number }",

    // TypeScript-only containers must not be flagged.
    "interface I { ['foo']: number }",
    "type T = { ['foo']: number }",
    "interface I { ['foo'](): void }",
    "type T = { readonly ['foo']: number }",
  ],
  invalid: [
    {
      code: "({ ['0']: 0 })",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: "var { ['0']: a } = obj",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: "({ ['0+1,234']: 0 })",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: '({ [0]: 0 })',
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: 'var { [0]: a } = obj',
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: "({ ['x']: 0 })",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: "var { ['x']: a } = obj",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: "var { ['__proto__']: a } = obj",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: "({ ['x']() {} })",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },

    // ---- Comments block auto-fix ----
    {
      code: "({ [/* this comment prevents a fix */ 'x']: 0 })",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: "({ ['x' /* this comment also prevents a fix */]: 0 })",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },

    // ---- Parenthesized literals ----
    {
      code: "({ [('x')]: 0 })",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: "var { [('x')]: a } = obj",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },

    // ---- Generator / async object methods ----
    {
      code: "({ *['x']() {} })",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: "({ async ['x']() {} })",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },

    // ---- Adjacency ----
    {
      code: '({ get[.2]() {} })',
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: '({ set[.2](value) {} })',
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: '({ async[.2]() {} })',
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: '({ [2]() {} })',
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: '({ get [2]() {} })',
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: '({ set [2](value) {} })',
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: '({ async [2]() {} })',
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: '({ get[2]() {} })',
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: '({ set[2](value) {} })',
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: '({ async[2]() {} })',
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: "({ get['foo']() {} })",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: '({ *[2]() {} })',
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: '({ async*[2]() {} })',
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },

    // ---- Object reserved-name keys ----
    {
      code: "({ ['constructor']: 1 })",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: "({ ['prototype']: 1 })",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },

    // ---- Class methods ----
    {
      code: "class Foo { ['0']() {} }",
      options: [{ enforceForClassMembers: true }] as any,
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: "class Foo { ['0+1,234']() {} }",
      options: [{}] as any,
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: "class Foo { ['x']() {} }",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: "class Foo { [/* this comment prevents a fix */ 'x']() {} }",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: "class Foo { ['x' /* this comment also prevents a fix */]() {} }",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: "class Foo { [('x')]() {} }",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: "class Foo { *['x']() {} }",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: "class Foo { async ['x']() {} }",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: 'class Foo { get[.2]() {} }',
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: 'class Foo { set[.2](value) {} }',
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: 'class Foo { async[.2]() {} }',
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: 'class Foo { [2]() {} }',
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: 'class Foo { get [2]() {} }',
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: 'class Foo { set [2](value) {} }',
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: 'class Foo { async [2]() {} }',
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: 'class Foo { get[2]() {} }',
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: 'class Foo { set[2](value) {} }',
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: 'class Foo { async[2]() {} }',
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: "class Foo { get['foo']() {} }",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: 'class Foo { *[2]() {} }',
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: 'class Foo { async*[2]() {} }',
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: "class Foo { static ['constructor']() {} }",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: "class Foo { ['prototype']() {} }",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: "(class { ['x']() {} })",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: "(class { ['__proto__']() {} })",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: "(class { static ['__proto__']() {} })",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: "(class { static ['constructor']() {} })",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: "(class { ['prototype']() {} })",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },

    // ---- Class fields ----
    {
      code: "class Foo { ['0'] }",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: "class Foo { ['0'] = 0 }",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: 'class Foo { static[0] }',
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: "class Foo { ['#foo'] }",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: "(class { ['__proto__'] })",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: "(class { static ['__proto__'] })",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: "(class { ['prototype'] })",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },

    // Multi-line cases — lock line/column & end-position via snapshot.
    {
      code: "({\n  ['x']: 0\n})",
      errors: [
        {
          messageId: 'unnecessarilyComputedProperty',
          line: 2,
          column: 3,
          endLine: 2,
          endColumn: 11,
        },
      ],
    },
    {
      code: "class Foo {\n  static ['x']() {\n    return 1;\n  }\n}",
      errors: [
        {
          messageId: 'unnecessarilyComputedProperty',
          line: 2,
          column: 3,
        },
      ],
    },

    // Numeric raw text preserved (hex, exponent).
    {
      code: '({ [0x1]: 0 })',
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: '({ [1e2]: 0 })',
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },

    // String-literal key variants.
    {
      code: 'const x = { ["x"]: 1 }',
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: 'const x = { [""]: 1 }',
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: 'const x = { ["\\u00e9"]: 1 }',
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: 'const x = { ["delete"]: 1 }',
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },

    // Numeric-literal raw preserved.
    {
      code: 'const x = { [0b10]: 1 }',
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: 'const x = { [0o7]: 1 }',
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: 'const x = { [1_000]: 1 }',
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: 'const x = { [1e-2]: 1 }',
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },

    // Multiple paren levels.
    {
      code: "const x = { [(('x'))]: 1 }",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },

    // Multi-modifier class members.
    {
      code: "class Foo { public static ['x']() {} }",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: "class Foo { static readonly ['x'] = 1 }",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: "class Foo { static async *['x']() {} }",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },

    // Expression contexts.
    {
      code: "const arrow = () => ({ ['x']: 1 })",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: "const cond = true ? { ['x']: 1 } : { a: 2 }",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: "function f() { return { ['x']: 1 } }",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },

    // Parameter destructuring with default.
    {
      code: "function f({ ['x']: a } = {} as any) { return a }",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: "const { ['x']: a, ...rest } = { x: 1, y: 2 }",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: "try {} catch ({ ['x']: e }) {}",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },

    // Multiple diagnostics in one object.
    {
      code: "const x = { ['x']: 1, ['y']: 2, a: 3, ['z']: 4 }",
      errors: [
        { messageId: 'unnecessarilyComputedProperty' },
        { messageId: 'unnecessarilyComputedProperty' },
        { messageId: 'unnecessarilyComputedProperty' },
      ],
    },

    // Multi-line computed bracket span.
    {
      code: "const x = {\n  [\n    'x'\n  ]: 1,\n}",
      errors: [
        { messageId: 'unnecessarilyComputedProperty', line: 2, column: 3 },
      ],
    },

    // Class implementing interface.
    {
      code: "interface I { x(): void } class C implements I { ['x']() {} }",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },

    // Class with generic.
    {
      code: "class Box<T> { ['x']: T | undefined }",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },

    // Namespace containing class.
    {
      code: "namespace N { export class C { ['x']() {} } }",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },

    // Nested for-of pattern.
    {
      code: "for (const { a: { ['x']: b } } of [{ a: { x: 1 } }]) { void b }",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },

    // Fix remains valid when key value is a context-sensitive keyword.
    {
      code: "const x = { async ['async']() {} }",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: "const x = { get ['get']() { return 1 } }",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: "const x = { set ['set'](v) {} }",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },

    // Non-static class method named 'prototype' reports (only
    // non-static 'constructor' is carved out).
    {
      code: "class Foo { ['prototype']() {} }",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },

    // Same literal key on get + set, both report.
    {
      code: "class Foo { get ['k']() { return 1 } set ['k'](v) {} }",
      errors: [
        { messageId: 'unnecessarilyComputedProperty' },
        { messageId: 'unnecessarilyComputedProperty' },
      ],
    },

    // Class getter/setter with computed literal key.
    {
      code: "class Foo { get ['x']() { return 1 } }",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: "class Foo { set ['x'](v) {} }",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },

    // Destructuring with default value.
    {
      code: "var { ['x']: a = 1 } = obj",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: "({ ['x']: a = 1 } = b)",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },

    // `override` modifier is a plain MethodDefinition in TSESTree.
    {
      code: "class Base { x() {} }; class Sub extends Base { override ['x']() {} }",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },

    // `declare class` members are plain MethodDefinition/PropertyDefinition.
    {
      code: "declare class Foo { ['x']: string }",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: "declare class Foo { ['x'](): void }",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },

    // Object literal wrapped in type-assertion / as / satisfies.
    {
      code: "const x = <const>{ ['x']: 1 }",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: "const x = { ['x']: 1 } as const",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: "const x = { ['x']: 1 } satisfies Record<string, number>",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },

    // Assignment destructuring must not get the __proto__ carve-out that
    // value-position object literals get.
    {
      code: "({ ['__proto__']: a } = b)",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: "({ y: { ['__proto__']: a } } = b)",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: "for ({ ['__proto__']: a } of arr) {}",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: "for ({ ['__proto__']: a } in arr) {}",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
    {
      code: "[{ ['__proto__']: a }] = [b]",
      errors: [{ messageId: 'unnecessarilyComputedProperty' }],
    },
  ],
});

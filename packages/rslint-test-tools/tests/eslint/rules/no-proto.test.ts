import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-proto', {
  valid: [
    // --- Object literal contexts (not member access) ---
    'var a = { __proto__: [] }',
    'var a = { __proto__ }',
    'var a = { ["__proto__"]: [] }',
    'var a = { __proto__() {} }',
    'var a = { get __proto__() { return 1; } }',
    'var a = { set __proto__(v) {} }',

    // --- __proto__ as declaration name ---
    'var __proto__ = 1;',
    'let __proto__ = 2;',
    'const __proto__ = 3;',
    'function __proto__() {}',
    'function foo(__proto__) {}',

    // --- Destructuring (binding pattern, not property access) ---
    'var { __proto__ } = obj;',
    'var { __proto__: proto } = obj;',
    'var { a: { __proto__ } } = obj;',
    'function foo({ __proto__ }) {}',
    // Array destructuring binding
    'var [__proto__] = arr;',
    // Rest element in destructuring
    'var { a, ...__proto__ } = obj;',
    'var [a, ...__proto__] = arr;',

    // --- Catch clause binding ---
    'try {} catch (__proto__) {}',

    // --- For-in / for-of declarations ---
    'for (var __proto__ in obj) {}',
    'for (var __proto__ of arr) {}',

    // --- Import / Export ---
    "import { __proto__ } from 'mod';",
    'var __proto__ = 1; export { __proto__ };',

    // --- TypeScript type-level constructs ---
    'interface I { __proto__: string }',
    'type T = { __proto__: string }',
    'declare class C { __proto__: string }',

    // --- Class member declarations ---
    'class C { __proto__ = 1 }',
    'class C { __proto__() {} }',
    'class C { get __proto__() { return 1; } }',
    'class C { static __proto__ = 1 }',

    // --- Enum member ---
    'enum E { __proto__ = 1 }',

    // --- __proto__ as type-level names ---
    'class __proto__ {}',
    'type __proto__ = string;',
    'namespace __proto__ {}',
    'function foo<__proto__>() {}',
    'declare function __proto__(): void;',
    'abstract class C { abstract __proto__(): void }',

    // --- Label ---
    '__proto__: for (;;) { break __proto__; }',

    // --- String / non-member-access usage ---
    'var s = "__proto__";',
    'var s = `__proto__`;',

    // --- Different property name ---
    'obj.prototype',
    'obj.__proto',
    'obj.proto__',
    'obj.__PROTO__',

    // --- Recommended alternatives ---
    'var a = Object.getPrototypeOf(obj);',
    'Object.setPrototypeOf(obj, b);',

    // --- Dynamic / non-static property access ---
    'var x = "__proto__"; obj[x];',
    "obj[`__${'proto'}__`]",
    'var k = "__proto__"; obj[k] = 1;',
  ],
  invalid: [
    // === Dot notation ===
    {
      code: 'var a = obj.__proto__;',
      errors: [{ messageId: 'unexpectedProto' }],
    },
    {
      code: 'obj.__proto__ = b;',
      errors: [{ messageId: 'unexpectedProto' }],
    },

    // === Bracket notation ===
    {
      code: 'var a = obj["__proto__"];',
      errors: [{ messageId: 'unexpectedProto' }],
    },
    {
      code: 'obj["__proto__"] = b;',
      errors: [{ messageId: 'unexpectedProto' }],
    },
    {
      code: "var a = obj['__proto__'];",
      errors: [{ messageId: 'unexpectedProto' }],
    },
    // Template literal (no substitution)
    {
      code: 'var a = obj[`__proto__`];',
      errors: [{ messageId: 'unexpectedProto' }],
    },

    // === Optional chaining ===
    {
      code: 'obj?.__proto__;',
      errors: [{ messageId: 'unexpectedProto' }],
    },
    {
      code: 'obj?.["__proto__"];',
      errors: [{ messageId: 'unexpectedProto' }],
    },

    // === Chained / nested access ===
    {
      code: 'var a = foo.bar.__proto__;',
      errors: [{ messageId: 'unexpectedProto' }],
    },
    {
      code: 'obj.__proto__.hasOwnProperty("foo");',
      errors: [{ messageId: 'unexpectedProto' }],
    },
    // Double __proto__ — two errors
    {
      code: 'obj.__proto__.__proto__;',
      errors: [
        { messageId: 'unexpectedProto' },
        { messageId: 'unexpectedProto' },
      ],
    },
    // Deeply chained
    {
      code: 'a.b.c.d.__proto__;',
      errors: [{ messageId: 'unexpectedProto' }],
    },
    // Optional chain continuing after __proto__
    {
      code: 'obj?.__proto__?.toString();',
      errors: [{ messageId: 'unexpectedProto' }],
    },

    // === Calling __proto__ as function ===
    {
      code: 'obj.__proto__();',
      errors: [{ messageId: 'unexpectedProto' }],
    },
    {
      code: 'obj["__proto__"]();',
      errors: [{ messageId: 'unexpectedProto' }],
    },

    // === this ===
    {
      code: 'var a = this.__proto__;',
      errors: [{ messageId: 'unexpectedProto' }],
    },
    {
      code: 'this["__proto__"];',
      errors: [{ messageId: 'unexpectedProto' }],
    },

    // === Parenthesized expression ===
    {
      code: '(obj).__proto__;',
      errors: [{ messageId: 'unexpectedProto' }],
    },
    {
      code: '(obj)["__proto__"];',
      errors: [{ messageId: 'unexpectedProto' }],
    },

    // === TypeScript expressions ===
    // as assertion
    {
      code: '(obj as any).__proto__;',
      errors: [{ messageId: 'unexpectedProto' }],
    },
    // Angle-bracket assertion
    {
      code: '(<any>obj).__proto__;',
      errors: [{ messageId: 'unexpectedProto' }],
    },
    // Non-null assertion
    {
      code: 'obj!.__proto__;',
      errors: [{ messageId: 'unexpectedProto' }],
    },
    // Satisfies expression
    {
      code: '(obj satisfies any).__proto__;',
      errors: [{ messageId: 'unexpectedProto' }],
    },

    // === As function argument ===
    {
      code: 'foo(obj.__proto__);',
      errors: [{ messageId: 'unexpectedProto' }],
    },
    {
      code: 'console.log(obj["__proto__"]);',
      errors: [{ messageId: 'unexpectedProto' }],
    },

    // === new / await / yield ===
    {
      code: 'new (obj.__proto__)();',
      errors: [{ messageId: 'unexpectedProto' }],
    },
    {
      code: 'async function f() { await obj.__proto__; }',
      errors: [{ messageId: 'unexpectedProto' }],
    },
    {
      code: 'function* g() { yield obj.__proto__; }',
      errors: [{ messageId: 'unexpectedProto' }],
    },

    // === Tagged template ===
    {
      code: 'obj.__proto__`tagged`;',
      errors: [{ messageId: 'unexpectedProto' }],
    },

    // === Update expressions (++/--) ===
    {
      code: '++obj.__proto__;',
      errors: [{ messageId: 'unexpectedProto' }],
    },
    {
      code: 'obj.__proto__++;',
      errors: [{ messageId: 'unexpectedProto' }],
    },

    // === In expressions ===
    // Template literal expression
    {
      code: 'var s = `${obj.__proto__}`;',
      errors: [{ messageId: 'unexpectedProto' }],
    },
    // Ternary
    {
      code: 'var a = x ? obj.__proto__ : null;',
      errors: [{ messageId: 'unexpectedProto' }],
    },
    // Logical OR / AND
    {
      code: 'var a = obj.__proto__ || default_;',
      errors: [{ messageId: 'unexpectedProto' }],
    },
    {
      code: 'var a = x && obj.__proto__;',
      errors: [{ messageId: 'unexpectedProto' }],
    },
    // Nullish coalescing
    {
      code: 'var a = obj.__proto__ ?? default_;',
      errors: [{ messageId: 'unexpectedProto' }],
    },
    // Comma operator
    {
      code: '(0, obj.__proto__);',
      errors: [{ messageId: 'unexpectedProto' }],
    },
    // in expression
    {
      code: '"x" in obj.__proto__;',
      errors: [{ messageId: 'unexpectedProto' }],
    },
    // instanceof
    {
      code: 'obj.__proto__ instanceof Object;',
      errors: [{ messageId: 'unexpectedProto' }],
    },

    // === Unary operators ===
    {
      code: 'typeof obj.__proto__;',
      errors: [{ messageId: 'unexpectedProto' }],
    },
    {
      code: 'void obj.__proto__;',
      errors: [{ messageId: 'unexpectedProto' }],
    },
    {
      code: 'delete obj.__proto__;',
      errors: [{ messageId: 'unexpectedProto' }],
    },

    // === Spread ===
    {
      code: 'var a = { ...obj.__proto__ };',
      errors: [{ messageId: 'unexpectedProto' }],
    },
    {
      code: 'var a = [...obj.__proto__];',
      errors: [{ messageId: 'unexpectedProto' }],
    },

    // === Assignment patterns ===
    // Compound assignment
    {
      code: 'obj.__proto__ += "";',
      errors: [{ messageId: 'unexpectedProto' }],
    },
    // Logical assignment operators
    {
      code: 'obj.__proto__ ||= x;',
      errors: [{ messageId: 'unexpectedProto' }],
    },
    {
      code: 'obj.__proto__ &&= x;',
      errors: [{ messageId: 'unexpectedProto' }],
    },
    {
      code: 'obj.__proto__ ??= x;',
      errors: [{ messageId: 'unexpectedProto' }],
    },
    // Destructuring default value
    {
      code: 'var { a = obj.__proto__ } = b;',
      errors: [{ messageId: 'unexpectedProto' }],
    },
    // Array destructuring assignment target
    {
      code: '[obj.__proto__] = [1];',
      errors: [{ messageId: 'unexpectedProto' }],
    },

    // === As value in object / array literal ===
    // Object property value
    {
      code: 'var a = { key: obj.__proto__ };',
      errors: [{ messageId: 'unexpectedProto' }],
    },
    // Array element
    {
      code: 'var a = [obj.__proto__];',
      errors: [{ messageId: 'unexpectedProto' }],
    },
    // Computed property key
    {
      code: 'var a = { [obj.__proto__]: 1 };',
      errors: [{ messageId: 'unexpectedProto' }],
    },

    // === for-in / for-of with member expression as target ===
    {
      code: 'for (obj.__proto__ in x) {}',
      errors: [{ messageId: 'unexpectedProto' }],
    },
    {
      code: 'for (obj.__proto__ of x) {}',
      errors: [{ messageId: 'unexpectedProto' }],
    },

    // === Function / class contexts ===
    {
      code: 'function f() { return obj.__proto__; }',
      errors: [{ messageId: 'unexpectedProto' }],
    },
    {
      code: 'var f = () => obj.__proto__;',
      errors: [{ messageId: 'unexpectedProto' }],
    },
    {
      code: 'async function f() { return obj.__proto__; }',
      errors: [{ messageId: 'unexpectedProto' }],
    },
    {
      code: 'class C { method() { return this.__proto__; } }',
      errors: [{ messageId: 'unexpectedProto' }],
    },
    {
      code: 'class C { constructor() { this.__proto__ = null; } }',
      errors: [{ messageId: 'unexpectedProto' }],
    },
    {
      code: 'class C { get p() { return obj.__proto__; } }',
      errors: [{ messageId: 'unexpectedProto' }],
    },
    {
      code: 'class C { static { obj.__proto__; } }',
      errors: [{ messageId: 'unexpectedProto' }],
    },
    // Class field initializer value
    {
      code: 'class C { x = obj.__proto__ }',
      errors: [{ messageId: 'unexpectedProto' }],
    },
    // IIFE
    {
      code: '(function() { return obj.__proto__; })();',
      errors: [{ messageId: 'unexpectedProto' }],
    },
    // Arrow returning object
    {
      code: 'var f = () => ({ a: obj.__proto__ });',
      errors: [{ messageId: 'unexpectedProto' }],
    },

    // === Namespace / module scope ===
    {
      code: 'namespace N { var a = obj.__proto__; }',
      errors: [{ messageId: 'unexpectedProto' }],
    },

    // === Export ===
    {
      code: 'export default obj.__proto__;',
      errors: [{ messageId: 'unexpectedProto' }],
    },

    // === Control flow ===
    {
      code: 'if (obj.__proto__) {}',
      errors: [{ messageId: 'unexpectedProto' }],
    },
    {
      code: 'for (var x in obj.__proto__) {}',
      errors: [{ messageId: 'unexpectedProto' }],
    },
    {
      code: 'for (var x of obj.__proto__) {}',
      errors: [{ messageId: 'unexpectedProto' }],
    },
    {
      code: 'while (obj.__proto__) {}',
      errors: [{ messageId: 'unexpectedProto' }],
    },
    {
      code: 'switch (obj.__proto__) { case 0: break; }',
      errors: [{ messageId: 'unexpectedProto' }],
    },
    {
      code: 'throw obj.__proto__;',
      errors: [{ messageId: 'unexpectedProto' }],
    },
    // do-while
    {
      code: 'do {} while (obj.__proto__);',
      errors: [{ messageId: 'unexpectedProto' }],
    },
    // try / catch / finally
    {
      code: 'try { obj.__proto__; } catch(e) {}',
      errors: [{ messageId: 'unexpectedProto' }],
    },
    {
      code: 'try {} catch(e) { obj.__proto__; }',
      errors: [{ messageId: 'unexpectedProto' }],
    },
    {
      code: 'try {} finally { obj.__proto__; }',
      errors: [{ messageId: 'unexpectedProto' }],
    },
    // For init
    {
      code: 'for (obj.__proto__ = 0;;) {}',
      errors: [{ messageId: 'unexpectedProto' }],
    },

    // === Multiple occurrences ===
    {
      code: 'a.__proto__;\nb.__proto__;',
      errors: [
        { messageId: 'unexpectedProto' },
        { messageId: 'unexpectedProto' },
      ],
    },
    // Mixed dot and bracket
    {
      code: 'obj.__proto__; obj["__proto__"];',
      errors: [
        { messageId: 'unexpectedProto' },
        { messageId: 'unexpectedProto' },
      ],
    },

    // === Multi-byte characters ===
    {
      code: '/* 🚀 */ obj.__proto__',
      errors: [{ messageId: 'unexpectedProto' }],
    },
  ],
});

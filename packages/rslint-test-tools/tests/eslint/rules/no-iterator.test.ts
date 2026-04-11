import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-iterator', {
  valid: [
    // Computed property with identifier (not a string literal)
    'var a = test[__iterator__];',
    // Variable declaration, not member access
    'var __iterator__ = null;',
    // Template literal missing trailing __
    'foo[`__iterator`] = null;',
    // Template literal with newline (not exact match)
    'foo[`__iterator__\n`] = null;',
    // Template literal with expression
    'foo[`__iterator__${x}`] = null;',
    // Object property key (not a member access)
    'var obj = { __iterator__: 1 };',
    // Object shorthand
    'var obj = { __iterator__ };',
    // Destructuring
    'const { __iterator__ } = obj;',
    'const { __iterator__: alias } = obj;',
    // Function declaration
    'function __iterator__() {}',
    // Class method declaration
    'class Foo { __iterator__() {} }',
    // TypeScript interface property
    'interface Foo { __iterator__: any; }',
    // TypeScript type alias
    'type Foo = { __iterator__: any };',
    // Parameter name
    'function foo(__iterator__: any) {}',
    // Enum member
    'enum Foo { __iterator__ }',
    // Private identifier (PrivateIdentifier, not Identifier)
    'class Foo { #__iterator__ = 1; m() { this.#__iterator__; } }',
    // TypeScript QualifiedName in type position (not PropertyAccessExpression)
    'declare namespace A { var __iterator__: number; } type X = typeof A.__iterator__;',
    // Import/export specifier
    "export { __iterator__ } from 'mod';",
    // Label
    '__iterator__: for (;;) { break __iterator__; }',
    // Computed with string concatenation (not a static literal)
    "obj['__iterator' + '__'];",
    // Computed with variable
    "const key = '__iterator__'; obj[key];",
    // Getter/setter declaration in object
    'var obj = { get __iterator__() { return 1; } };',
    // JSX attribute — only tested in Go (needs Tsx: true flag)

    // Unicode escape in identifier resolves to __iterator__ but
    // this is testing valid Go-side only since JS test tooling handles
    // identifiers differently
  ],
  invalid: [
    // ========================================
    // Basic access patterns
    // ========================================
    {
      code: 'var a = test.__iterator__;',
      errors: [{ messageId: 'noIterator' }],
    },
    {
      code: "var a = test['__iterator__'];",
      errors: [{ messageId: 'noIterator' }],
    },
    {
      code: 'var a = test["__iterator__"];',
      errors: [{ messageId: 'noIterator' }],
    },
    {
      code: 'var a = test[`__iterator__`];',
      errors: [{ messageId: 'noIterator' }],
    },

    // ========================================
    // Assignment targets
    // ========================================
    {
      code: 'Foo.prototype.__iterator__ = function() {};',
      errors: [{ messageId: 'noIterator' }],
    },
    {
      code: 'test[`__iterator__`] = function () {};',
      errors: [{ messageId: 'noIterator' }],
    },
    {
      code: 'obj.__iterator__ = 42;',
      errors: [{ messageId: 'noIterator' }],
    },

    // ========================================
    // Chained member access
    // ========================================
    {
      code: 'a.b.__iterator__;',
      errors: [{ messageId: 'noIterator' }],
    },
    {
      code: 'a.b.c.__iterator__;',
      errors: [{ messageId: 'noIterator' }],
    },
    // 2 errors: a.__iterator__ and a.__iterator__.__iterator__
    {
      code: 'a.__iterator__.__iterator__;',
      errors: [{ messageId: 'noIterator' }, { messageId: 'noIterator' }],
    },

    // ========================================
    // Optional chaining
    // ========================================
    {
      code: 'obj?.__iterator__;',
      errors: [{ messageId: 'noIterator' }],
    },
    {
      code: "obj?.['__iterator__'];",
      errors: [{ messageId: 'noIterator' }],
    },
    {
      code: 'a?.b?.__iterator__;',
      errors: [{ messageId: 'noIterator' }],
    },

    // ========================================
    // Parenthesized / TypeScript outer expressions
    // ========================================
    {
      code: '(obj).__iterator__;',
      errors: [{ messageId: 'noIterator' }],
    },
    {
      code: '((obj)).__iterator__;',
      errors: [{ messageId: 'noIterator' }],
    },
    {
      code: 'obj!.__iterator__;',
      errors: [{ messageId: 'noIterator' }],
    },
    {
      code: '(obj as any).__iterator__;',
      errors: [{ messageId: 'noIterator' }],
    },
    {
      code: '(<any>obj).__iterator__;',
      errors: [{ messageId: 'noIterator' }],
    },

    // ========================================
    // this / class contexts
    // ========================================
    {
      code: 'this.__iterator__;',
      errors: [{ messageId: 'noIterator' }],
    },
    {
      code: 'class A { m() { this.__iterator__; } }',
      errors: [{ messageId: 'noIterator' }],
    },
    {
      code: 'class A { x = this.__iterator__; }',
      errors: [{ messageId: 'noIterator' }],
    },
    {
      code: 'class A { static { this.__iterator__; } }',
      errors: [{ messageId: 'noIterator' }],
    },

    // ========================================
    // Nested in functions / arrows
    // ========================================
    {
      code: 'function foo() { obj.__iterator__; }',
      errors: [{ messageId: 'noIterator' }],
    },
    {
      code: 'const f = () => obj.__iterator__;',
      errors: [{ messageId: 'noIterator' }],
    },
    {
      code: 'const f = () => { return obj.__iterator__; };',
      errors: [{ messageId: 'noIterator' }],
    },

    // ========================================
    // In expressions
    // ========================================
    {
      code: 'var x = cond ? obj.__iterator__ : null;',
      errors: [{ messageId: 'noIterator' }],
    },
    {
      code: 'foo(obj.__iterator__);',
      errors: [{ messageId: 'noIterator' }],
    },
    {
      code: 'var x = `${obj.__iterator__}`;',
      errors: [{ messageId: 'noIterator' }],
    },

    // ========================================
    // Multiple violations in one file
    // ========================================
    {
      code: "a.__iterator__;\nb['__iterator__'];",
      errors: [{ messageId: 'noIterator' }, { messageId: 'noIterator' }],
    },

    // ========================================
    // In loops
    // ========================================
    {
      code: 'for (var x in obj.__iterator__) {}',
      errors: [{ messageId: 'noIterator' }],
    },
    {
      code: 'while (obj.__iterator__) {}',
      errors: [{ messageId: 'noIterator' }],
    },

    // ========================================
    // Unary / delete / typeof / void
    // ========================================
    {
      code: 'delete obj.__iterator__;',
      errors: [{ messageId: 'noIterator' }],
    },
    {
      code: 'typeof obj.__iterator__;',
      errors: [{ messageId: 'noIterator' }],
    },
    {
      code: 'void obj.__iterator__;',
      errors: [{ messageId: 'noIterator' }],
    },

    // ========================================
    // Call / new
    // ========================================
    {
      code: 'obj.__iterator__();',
      errors: [{ messageId: 'noIterator' }],
    },
    {
      code: 'new obj.__iterator__();',
      errors: [{ messageId: 'noIterator' }],
    },

    // ========================================
    // super
    // ========================================
    {
      code: 'class A extends B { m() { super.__iterator__; } }',
      errors: [{ messageId: 'noIterator' }],
    },
    {
      code: "class A extends B { m() { super['__iterator__']; } }",
      errors: [{ messageId: 'noIterator' }],
    },

    // ========================================
    // Spread / array / object value
    // ========================================
    {
      code: 'var a = [obj.__iterator__];',
      errors: [{ messageId: 'noIterator' }],
    },
    {
      code: 'var a = { x: obj.__iterator__ };',
      errors: [{ messageId: 'noIterator' }],
    },
    {
      code: 'var a = { ...obj.__iterator__ };',
      errors: [{ messageId: 'noIterator' }],
    },

    // ========================================
    // Logical / nullish / assignment operators
    // ========================================
    {
      code: 'obj.__iterator__ || fallback;',
      errors: [{ messageId: 'noIterator' }],
    },
    {
      code: 'obj.__iterator__ ?? fallback;',
      errors: [{ messageId: 'noIterator' }],
    },
    {
      code: 'obj.__iterator__ ??= fallback;',
      errors: [{ messageId: 'noIterator' }],
    },

    // ========================================
    // Async / generator
    // ========================================
    {
      code: 'async function f() { await obj.__iterator__; }',
      errors: [{ messageId: 'noIterator' }],
    },
    {
      code: 'function* g() { yield obj.__iterator__; }',
      errors: [{ messageId: 'noIterator' }],
    },

    // ========================================
    // Mixed bracket + dot chaining
    // ========================================
    {
      code: "obj['__iterator__'].__iterator__;",
      errors: [{ messageId: 'noIterator' }, { messageId: 'noIterator' }],
    },

    // ========================================
    // Escape sequences in string/template literals
    // ========================================
    {
      code: "obj['\\x5F\\x5Fiterator\\x5F\\x5F'];",
      errors: [{ messageId: 'noIterator' }],
    },
    {
      code: "obj['\\u005F\\u005Fiterator\\u005F\\u005F'];",
      errors: [{ messageId: 'noIterator' }],
    },
    {
      code: 'obj[`\\x5F\\x5Fiterator\\x5F\\x5F`];',
      errors: [{ messageId: 'noIterator' }],
    },
    {
      code: 'obj.\\u005F\\u005Fiterator\\u005F\\u005F;',
      errors: [{ messageId: 'noIterator' }],
    },

    // ========================================
    // Satisfies (TS 4.9+)
    // ========================================
    {
      code: '(obj satisfies any).__iterator__;',
      errors: [{ messageId: 'noIterator' }],
    },

    // ========================================
    // Multiline
    // ========================================
    {
      code: 'obj\n  .__iterator__;',
      errors: [{ messageId: 'noIterator' }],
    },
    {
      code: "obj\n  ['__iterator__'];",
      errors: [{ messageId: 'noIterator' }],
    },
  ],
});

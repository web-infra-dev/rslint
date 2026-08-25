import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();
const scriptLanguageOptions = { sourceType: 'script' as const };

ruleTester.run('no-invalid-this', {
  valid: [
    // ================================================================
    // Sloppy-mode script: rule never fires (default `this` binding is
    // the global object, not undefined)
    // ================================================================
    { code: 'console.log(this);', languageOptions: scriptLanguageOptions },
    {
      code: 'function foo() { console.log(this); }',
      languageOptions: scriptLanguageOptions,
    },

    // ================================================================
    // Constructors
    // ================================================================
    '"use strict"; function Foo() { this.a = 0; }',
    '"use strict"; var Foo = function Foo() { this.a = 0; };',
    '"use strict"; class A { constructor() { this.a = 0; } }',
    {
      code: '"use strict"; function Foo() { this.a = 0; }',
      options: { capIsConstructor: true },
    },

    // ================================================================
    // Methods
    // ================================================================
    '"use strict"; var obj = { foo: function() { this.a = 0; } };',
    '"use strict"; var obj = { foo() { this.a = 0; } };',
    '"use strict"; obj.foo = function() { this.a = 0; };',
    '"use strict"; class A { foo() { this.a = 0; } }',
    '"use strict"; class A { static foo() { this.a = 0; } }',
    '"use strict"; Object.defineProperty(obj, "foo", { value: function() { this.a = 0; } });',

    // ================================================================
    // Bind / call / apply / array-method thisArg
    // ================================================================
    '"use strict"; var foo = function() { this.a = 0; }.bind(obj);',
    '"use strict"; (function() { this.a = 0; }).call(obj);',
    '"use strict"; (function() { this.a = 0; }).apply(obj);',
    '"use strict"; Reflect.apply(function() { this.a = 0; }, obj, []);',
    '"use strict"; foo.forEach(function() { this.a = 0; }, thisArg);',
    '"use strict"; Array.from([], function() { this.a = 0; }, obj);',

    // ================================================================
    // @this tag
    // ================================================================
    '"use strict"; /** @this Foo */ function foo() { this.a = 0; }',
    '"use strict"; foo(/* @this Obj */ function() { this.a = 0; });',

    // ================================================================
    // Class fields / static blocks
    // ================================================================
    'class C { field = this; }',
    'class C { static field = this; }',
    'class C { foo = () => this; }',
    'class C { static { this.x; } }',
    'class C { accessor c = this.a; }',

    // ================================================================
    // Top-level `this` in a script is always valid
    // ================================================================
    { code: 'this.a = 0;', languageOptions: scriptLanguageOptions },

    // ================================================================
    // TypeScript: explicit `this` parameter
    // ================================================================
    '"use strict"; interface SomeType { prop: string; } function foo(this: SomeType) { this.prop; }',
    '"use strict"; function foo(this: prop) { this.propMethod(); }',
  ],

  invalid: [
    // ================================================================
    // Top-level `this` in an ES module is always invalid
    // ================================================================
    {
      code: 'export {}; console.log(this);',
      errors: [{ messageId: 'unexpectedThis' }],
    },

    // ================================================================
    // Just functions (strict mode)
    // ================================================================
    {
      code: '"use strict"; function foo() { console.log(this); }',
      errors: [{ messageId: 'unexpectedThis' }],
    },
    {
      code: '"use strict"; function foo() { console.log(this); }',
      options: { capIsConstructor: false },
      errors: [{ messageId: 'unexpectedThis' }],
    },
    {
      code: '"use strict"; function Foo() { console.log(this); }',
      options: { capIsConstructor: false },
      errors: [{ messageId: 'unexpectedThis' }],
    },
    {
      code: '"use strict"; (function() { console.log(this); })();',
      errors: [{ messageId: 'unexpectedThis' }],
    },

    // ================================================================
    // Nested plain function inside a method escapes the method's frame
    // ================================================================
    {
      code: '"use strict"; class A { foo() { return function() { console.log(this); }; } }',
      errors: [{ messageId: 'unexpectedThis' }],
    },

    // ================================================================
    // Bind/call/apply with a null/undefined receiver stay default-bound
    // ================================================================
    {
      code: '"use strict"; var foo = function() { console.log(this); }.bind(null);',
      errors: [{ messageId: 'unexpectedThis' }],
    },
    {
      code: '"use strict"; (function() { console.log(this); }).call(undefined);',
      errors: [{ messageId: 'unexpectedThis' }],
    },

    // ================================================================
    // Array-method callback with no thisArg
    // ================================================================
    {
      code: '"use strict"; foo.forEach(function() { console.log(this); });',
      errors: [{ messageId: 'unexpectedThis' }],
    },

    // ================================================================
    // @this tag absent
    // ================================================================
    {
      code: '"use strict"; /** @returns {void} */ function foo() { console.log(this); }',
      errors: [{ messageId: 'unexpectedThis' }],
    },

    // ================================================================
    // Class static block: nested plain function escapes the block's frame
    // ================================================================
    {
      code: 'class C { static { function foo() { this.x; } } }',
      errors: [{ messageId: 'unexpectedThis' }],
    },

    // ================================================================
    // TypeScript: no this-param, function used as a plain callback
    // ================================================================
    {
      code: '"use strict"; interface SomeType { prop: string; } function foo() { this.prop; }',
      errors: [{ messageId: 'unexpectedThis' }],
    },
  ],
});

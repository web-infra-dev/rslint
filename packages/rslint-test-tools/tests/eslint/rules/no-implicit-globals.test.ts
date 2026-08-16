import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-implicit-globals', {
  valid: [
    // Recommended way to create a global variable in the browser
    {
      code: 'window.foo = 1;',
      languageOptions: { globals: { window: 'readonly' } },
    },
    {
      code: 'window.foo = function() {};',
      languageOptions: { globals: { window: 'readonly' } },
    },

    // Another way to create a global variable, not this rule's concern
    'this.foo = 1;',

    // Test that the rule doesn't report global comments
    '/*global foo:readonly*/',
    '/*global foo:writable*/',

    // Doesn't report function expressions
    'typeof function() {}',
    '(function() {}) + (function foo() {})',

    // Recommended way to create local variables
    '(function() { var foo = 1; })();',
    '(function() { function foo() {} })();',

    // Test default option: const/let/class are not checked unless lexicalBindings is set
    'const foo = 1; let bar; class Baz {}',
    {
      code: 'const foo = 1; let bar; class Baz {}',
      options: { lexicalBindings: false },
    },

    // If the option is not set to true, even readonly-global redeclarations are allowed
    'const Array = 1; let Object; class Math {}',

    // Doesn't report class expressions
    { code: 'typeof class {}', options: { lexicalBindings: true } },

    // Recommended way to create local lexical bindings
    {
      code: '{ const foo = 1; let bar; class Baz {} }',
      options: { lexicalBindings: true },
    },

    // This rule doesn't report all undeclared variables, just leaks
    'foo',
    'foo + bar',
    'foo(bar)',
    'foo++',
    'foo += 1',

    // Leaks are not possible in strict mode
    "'use strict';foo = 1;",

    // This rule doesn't check the existence of the objects in property assignments
    'Foo.bar = 1;',
    'Utils.foo = 1;',

    // Not a leak: configured writable global
    { code: 'foo = 1;', languageOptions: { globals: { foo: 'writable' } } },

    // Redeclarations of writable global variables are allowed
    '/*global foo:writable*/ var foo = 1;',
    {
      code: 'function foo() {}',
      languageOptions: { globals: { foo: 'writable' } },
    },

    // This rule doesn't disallow assignments to properties of readonly globals
    'Array.from = 1;',
    "Object['assign'] = 1;",

    // This rule doesn't disallow updates of readonly globals
    '/*global foo:readonly*/ foo++;',
    '/*global foo:readonly*/ foo += 1;',

    // `/* exported name */` suppresses the declaration report for that name
    "/* exported foo */ var foo = 'foo';",
    '/* exported foo */ function foo() {}',
    {
      code: '/* exported a */ const a = 1;',
      options: { lexicalBindings: true },
    },
  ],
  invalid: [
    // `var` and function declarations
    {
      code: 'var foo = 1;',
      errors: [{ messageId: 'globalNonLexicalBinding' }],
    },
    {
      code: 'function foo() {}',
      errors: [{ messageId: 'globalNonLexicalBinding' }],
    },
    {
      code: 'var foo = 1, bar = 2;',
      errors: [
        { messageId: 'globalNonLexicalBinding' },
        { messageId: 'globalNonLexicalBinding' },
      ],
    },

    // `const`, `let` and class declarations (with lexicalBindings: true)
    {
      code: 'const a = 1;',
      options: { lexicalBindings: true },
      errors: [{ messageId: 'globalLexicalBinding' }],
    },
    {
      code: 'let a;',
      options: { lexicalBindings: true },
      errors: [{ messageId: 'globalLexicalBinding' }],
    },
    {
      code: 'class A {}',
      options: { lexicalBindings: true },
      errors: [{ messageId: 'globalLexicalBinding' }],
    },

    // Leaks
    {
      code: 'foo = 1',
      errors: [{ messageId: 'globalVariableLeak' }],
    },
    {
      code: 'for (foo in {});',
      errors: [{ messageId: 'globalVariableLeak' }],
    },
    {
      code: 'foo = 1, bar = 2;',
      errors: [
        { messageId: 'globalVariableLeak' },
        { messageId: 'globalVariableLeak' },
      ],
    },

    // Read-only globals
    {
      code: 'Array = 1',
      errors: [{ messageId: 'assignmentToReadonlyGlobal', column: 1 }],
    },
    {
      code: 'foo = 1;',
      languageOptions: { globals: { foo: 'readonly' } },
      errors: [{ messageId: 'assignmentToReadonlyGlobal' }],
    },
    {
      code: 'var Array = 1',
      errors: [{ messageId: 'redeclarationOfReadonlyGlobal' }],
    },
    {
      code: '/*global foo:readonly*/ var foo',
      errors: [{ messageId: 'redeclarationOfReadonlyGlobal' }],
    },
    {
      code: '/*global foo:readonly*/ function foo() {}',
      errors: [{ messageId: 'redeclarationOfReadonlyGlobal' }],
    },

    // `/* exported name */` doesn't suppress a mismatched name, nor leaks
    {
      code: "/* exported bar */ var foo = 'text';",
      errors: [{ messageId: 'globalNonLexicalBinding' }],
    },
    {
      code: '/* exported foo */ foo = 1',
      errors: [{ messageId: 'globalVariableLeak' }],
    },
  ],
});

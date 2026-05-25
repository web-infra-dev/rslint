import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-alert', {
  valid: [
    // ================================================================
    // Non-prohibited calls
    // ================================================================
    'a[o.k](1)',
    'foo.alert(foo)',
    'foo.confirm(foo)',
    'foo.prompt(foo)',
    'console.alert()',
    'window.scroll()',
    'window.focus()',

    // ================================================================
    // Shadowing: function declaration
    // ================================================================
    'function alert() {} alert();',
    'function confirm() {} confirm();',
    'function prompt() {} prompt();',

    // ================================================================
    // Shadowing: variable declaration (var / let / const)
    // ================================================================
    'var alert = function() {}; alert();',
    'let alert = 1; alert();',
    'const alert = () => {}; alert();',

    // ================================================================
    // Shadowing: destructuring
    // ================================================================
    'const { alert } = obj; alert();',
    'const { a: alert } = obj; alert();',
    'const [alert] = arr; alert();',

    // ================================================================
    // Shadowing: function parameter
    // ================================================================
    'function foo(alert) { alert(); }',
    'const foo = (alert) => { alert(); }',
    'function foo({ alert }) { alert(); }',
    'function foo([alert]) { alert(); }',

    // ================================================================
    // Shadowing: inner scope (variable stays in scope for nested calls)
    // ================================================================
    'function foo() { var alert = bar; alert(); }',
    'var alert = function() {}; function test() { alert(); }',
    'function foo() { var alert = function() {}; function test() { alert(); } }',

    // ================================================================
    // Shadowing: class declaration
    // ================================================================
    'class alert {} new alert();',

    // ================================================================
    // Shadowing: catch clause
    // ================================================================
    'try {} catch(alert) { alert(); }',

    // ================================================================
    // Shadowing: import (Go tests only — virtual file can't resolve modules)
    // ================================================================

    // ================================================================
    // Shadowing: for-loop / for-of / for-in
    // ================================================================
    'for (let alert = 0; alert < 10; alert++) {}',
    'for (const alert of arr) { alert(); }',
    'for (const alert in obj) { alert(); }',

    // ================================================================
    // Shadowing: enum
    // ================================================================
    'enum alert { A, B }',

    // ================================================================
    // Dynamic property access (no static name)
    // ================================================================
    'window[alert]();',
    'window[x + y]();',

    // ================================================================
    // this: inside function (not global scope)
    // ================================================================
    'function foo() { this.alert(); }',
    'const foo = function() { this.alert(); }',

    // ================================================================
    // this: inside arrow function (ESLint scope = "function", not "global")
    // ================================================================
    'const foo = () => this.alert();',
    'const foo = () => { this.alert(); }',
    'const f = () => { const g = () => { this.alert(); } }',

    // ================================================================
    // this: inside class (class body creates its own this binding)
    // ================================================================
    'class A { foo() { this.alert(); } }',
    'class A { constructor() { this.alert(); } }',
    'class A { get x() { return this.alert(); } }',
    'class A { set x(v) { this.alert(); } }',
    // class field initializer — this = instance
    'class A { x = this.alert(); }',
    // class static field initializer — this = class
    'class A { static x = this.alert(); }',
    // class static block — this = class
    'class A { static { this.alert(); } }',
    // nested: arrow inside class method
    'class A { foo() { const f = () => this.alert(); } }',

    // ================================================================
    // Shadowed window / globalThis
    // ================================================================
    'function foo() { var window = bar; window.alert(); }',
    'const window = {}; window.alert();',
    'let window = {}; window.alert();',
    'var globalThis = foo; globalThis.alert();',
    'function foo() { var globalThis = foo; globalThis.alert(); }',
    // import window from 'w'; window.alert(); — Go tests only (module resolution)

    // ================================================================
    // Callee is a chained member expression (not direct prohibited call)
    // ================================================================
    'alert.call(null)',
    'alert.apply(null, [])',
    'window.alert.call(null)',

    // ================================================================
    // Not a call (no CallExpression)
    // ================================================================
    'var x = alert;',
    'var x = window.alert;',
    'typeof alert;',
    'new alert()',
  ],
  invalid: [
    // ================================================================
    // Direct calls: all three prohibited names
    // ================================================================
    {
      code: 'alert(foo)',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'confirm(foo)',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'prompt(foo)',
      errors: [{ messageId: 'unexpected' }],
    },

    // ================================================================
    // window.* dot access
    // ================================================================
    {
      code: 'window.alert(foo)',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'window.confirm(foo)',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'window.prompt(foo)',
      errors: [{ messageId: 'unexpected' }],
    },

    // ================================================================
    // window['*'] bracket access
    // ================================================================
    {
      code: "window['alert'](foo)",
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "window['confirm'](foo)",
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "window['prompt'](foo)",
      errors: [{ messageId: 'unexpected' }],
    },

    // ================================================================
    // Template literal bracket access
    // ================================================================
    {
      code: 'window[`alert`](foo)',
      errors: [{ messageId: 'unexpected' }],
    },

    // ================================================================
    // alert shadowed locally but window.alert still flagged
    // ================================================================
    {
      code: 'function alert() {} window.alert(foo)',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'var alert = function() {};\nwindow.alert(foo)',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'function foo(alert) { window.alert(); }',
      errors: [{ messageId: 'unexpected' }],
    },

    // ================================================================
    // Direct call inside function/arrow/class (not shadowed)
    // ================================================================
    {
      code: 'function foo() { alert(); }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'const foo = () => alert()',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'const foo = () => { alert(); }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'class A { foo() { alert(); } }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'class A { constructor() { alert(); } }',
      errors: [{ messageId: 'unexpected' }],
    },

    // ================================================================
    // Deeply nested (3+ levels)
    // ================================================================
    {
      code: 'function a() { function b() { function c() { alert(); } } }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'const a = () => { const b = () => { const c = () => { alert(); }; }; }',
      errors: [{ messageId: 'unexpected' }],
    },

    // ================================================================
    // Shadowed inside function but not at call site
    // ================================================================
    {
      code: 'function foo() { var alert = function() {}; }\nalert();',
      errors: [{ messageId: 'unexpected' }],
    },

    // ================================================================
    // this.alert at global scope
    // ================================================================
    {
      code: 'this.alert(foo)',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "this['alert'](foo)",
      errors: [{ messageId: 'unexpected' }],
    },

    // ================================================================
    // window shadowed inside function but not at outer call site
    // ================================================================
    {
      code: 'function foo() { var window = bar; window.alert(); }\nwindow.alert();',
      errors: [{ messageId: 'unexpected' }],
    },

    // ================================================================
    // globalThis access
    // ================================================================
    {
      code: "globalThis['alert'](foo)",
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'globalThis.alert();',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'function foo() { var globalThis = bar; globalThis.alert(); }\nglobalThis.alert();',
      errors: [{ messageId: 'unexpected' }],
    },

    // ================================================================
    // Optional chaining
    // ================================================================
    {
      code: 'window?.alert(foo)',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: '(window?.alert)(foo)',
      errors: [{ messageId: 'unexpected' }],
    },

    // ================================================================
    // Parenthesized callee
    // ================================================================
    {
      code: '(alert)(foo)',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: '((alert))(foo)',
      errors: [{ messageId: 'unexpected' }],
    },

    // ================================================================
    // window/globalThis inside nested functions (not shadowed)
    // ================================================================
    {
      code: 'function foo() { window.alert(); }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'const foo = () => window.confirm()',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'class A { foo() { window.prompt(); } }',
      errors: [{ messageId: 'unexpected' }],
    },

    // ================================================================
    // Multiple errors in one file
    // ================================================================
    {
      code: 'alert();\nconfirm();\nprompt();',
      errors: [
        { messageId: 'unexpected' },
        { messageId: 'unexpected' },
        { messageId: 'unexpected' },
      ],
    },

    // ================================================================
    // IIFE
    // ================================================================
    {
      code: '(function() { alert(); })()',
      errors: [{ messageId: 'unexpected' }],
    },

    // ================================================================
    // Parenthesized object expression
    // ================================================================
    {
      code: '(window).alert(foo)',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: '(globalThis).alert(foo)',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: '(this).alert(foo)',
      errors: [{ messageId: 'unexpected' }],
    },

    // ================================================================
    // Non-null assertion on object
    // ================================================================
    {
      code: 'window!.alert(foo)',
      errors: [{ messageId: 'unexpected' }],
    },

    // ================================================================
    // Type assertion on object
    // ================================================================
    {
      code: '(window as any).alert(foo)',
      errors: [{ messageId: 'unexpected' }],
    },

    // ================================================================
    // Optional chaining on this/globalThis
    // ================================================================
    {
      code: 'this?.alert(foo)',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'globalThis?.confirm()',
      errors: [{ messageId: 'unexpected' }],
    },

    // ================================================================
    // Direct call in class field / static block (alert not shadowed)
    // ================================================================
    {
      code: 'class A { x = alert(); }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'class A { static { alert(); } }',
      errors: [{ messageId: 'unexpected' }],
    },

    // ================================================================
    // window.alert in class field / static block (window not shadowed)
    // ================================================================
    {
      code: 'class A { x = window.alert(); }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'class A { static { window.alert(); } }',
      errors: [{ messageId: 'unexpected' }],
    },

    // ================================================================
    // TypeScript outer expressions on callee
    // ================================================================
    {
      code: 'alert!(foo)',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: '(alert as Function)(foo)',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: '(window.alert!)(foo)',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: '(window.alert as Function)(foo)',
      errors: [{ messageId: 'unexpected' }],
    },
  ],
});

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-implied-eval', {
  valid: [
    // ---- Direct references without a call ----
    'setTimeout();',
    'setTimeout;',
    'setTimeout = foo;',
    'window.setTimeout;',
    'window.setTimeout = foo;',
    "window['setTimeout'];",
    "window['setTimeout'] = foo;",
    'global.setTimeout;',
    'global.setTimeout = foo;',
    "global['setTimeout'];",
    "global['setTimeout'] = foo;",
    "globalThis['setTimeout'] = foo;",

    // ---- Calls where first argument is not evaluated as a string ----
    'setTimeout(function() { x = 1; }, 100);',
    'setInterval(function() { x = 1; }, 100);',
    'execScript(function() { x = 1; }, 100);',
    'window.setTimeout(function() { x = 1; }, 100);',
    'window.setInterval(function() { x = 1; }, 100);',
    'window.execScript(function() { x = 1; }, 100);',
    'window.setTimeout(foo, 100);',
    'window.setInterval(foo, 100);',
    'window.execScript(foo, 100);',
    'global.setTimeout(function() { x = 1; }, 100);',
    'global.setInterval(function() { x = 1; }, 100);',
    'global.execScript(function() { x = 1; }, 100);',
    'global.setTimeout(foo, 100);',
    'global.setInterval(foo, 100);',
    'global.execScript(foo, 100);',
    'globalThis.setTimeout(foo, 100);',

    // ---- Non-global receiver ----
    "foo.setTimeout('hi')",

    // ---- Identifier / function-expression arguments are safe ----
    'setTimeout(foo, 10)',
    'setInterval(1, 10)',
    'execScript(2)',
    'setTimeout(function() {}, 10)',
    "foo.setInterval('hi')",
    'setInterval(foo, 10)',
    'setInterval(function() {}, 10)',
    "foo.execScript('hi')",
    'execScript(foo)',
    'execScript(function() {})',

    // ---- Binary `+` of non-strings ----
    'setTimeout(foo + bar, 10)',

    // ---- Only the first argument is checked ----
    "setTimeout(foobar, 'buzz')",
    "setTimeout(foobar, foo + 'bar')",

    // ---- Only the immediate subtree is inspected ----
    "setTimeout(function() { return 'foobar'; }, 10)",

    // ---- Prefix-match names that are not setTimeout/etc. ----
    "setTimeoutFooBar('Foo Bar')",

    // ---- Non-global intermediate receivers ----
    "foo.window.setTimeout('foo', 100);",
    "foo.global.setTimeout('foo', 100);",

    // ---- Shadowing ----
    "var window; window.setTimeout('foo', 100);",
    "var global; global.setTimeout('foo', 100);",
    "function foo(window) { window.setTimeout('foo', 100); }",
    "function foo(global) { global.setTimeout('foo', 100); }",

    // ---- Passed as argument, not called ----
    "foo('', window.setTimeout);",
    "foo('', global.setTimeout);",

    // ---- Shadowing by function declaration ----
    `
    function execScript(string) {
      console.log("This is not your grandparent's execScript().");
    }

    execScript('wibble');
    `,
    `
    function setTimeout(string) {
      console.log("This is not your grandparent's setTimeout().");
    }

    setTimeout('wibble');
    `,
    `
    function setInterval(string) {
      console.log("This is not your grandparent's setInterval().");
    }

    setInterval('wibble');
    `,

    // ---- Shadowing in a nested scope ----
    `
    function outer() {
      function setTimeout(string) {
        console.log("Shadowed setTimeout");
      }
      setTimeout('code');
    }
    `,
    `
    function outer() {
      function setInterval(string) {
        console.log("Shadowed setInterval");
      }
      setInterval('code');
    }
    `,
    `
    function outer() {
      function execScript(string) {
        console.log("Shadowed execScript");
      }
      execScript('code');
    }
    `,
    `
    {
      const setTimeout = function(string) {
        console.log("Block-scoped setTimeout");
      };
      setTimeout('code');
    }
    `,
    `
    {
      const setInterval = function(string) {
        console.log("Block-scoped setInterval");
      };
      setInterval('code');
    }
    `,

    // ---- Template literal bracket names that are not a static eval-like name ----
    "window[`SetTimeOut`]('foo', 100);",
    "global[`SetTimeOut`]('foo', 100);",
    'global[`setTimeout${foo}`]("foo", 100);',
    'globalThis[`setTimeout${foo}`]("foo", 100);',
    "self[`SetTimeOut`]('foo', 100);",
    'self[`setTimeout${foo}`]("foo", 100);',

    // ---- self as global candidate ----
    'self.setTimeout;',
    'self.setTimeout = foo;',
    "self['setTimeout'];",
    "self['setTimeout'] = foo;",
    'self.setTimeout(function() { x = 1; }, 100);',
    'self.setInterval(function() { x = 1; }, 100);',
    'self.execScript(function() { x = 1; }, 100);',
    'self.setTimeout(foo, 100);',
    "foo.self.setTimeout('foo', 100);",
    "var self; self.setTimeout('foo', 100);",
    "function foo(self) { self.setTimeout('foo', 100); }",
    "foo('', self.setTimeout);",
    `
    function outer() {
      function self() {
        console.log("Shadowed self");
      }
      self.setTimeout('code');
    }`,

    // ---- Cross-candidate chains: upstream only walks same-name chains ----
    "window.global.setTimeout('code');",
    "self.window.setTimeout('code');",
    "globalThis.window.setTimeout('code');",
    "window['global']['setTimeout']('code');",
    "self['globalThis']['setTimeout']('code');",
    // `this` is not a global candidate
    "this.setTimeout('code');",
    "this['setTimeout']('code');",

    // ---- TS outer expressions on callee / receiver: upstream's
    // isSpecificId / isSpecificMemberAccess reject non-ESTree nodes,
    // so these are NOT flagged.
    "setTimeout!('code')",
    "(setTimeout as Function)('code')",
    "(window.setTimeout!)('code')",
    "(window.setTimeout as Function)('code')",
    "window!.setTimeout('code')",
    "(window as any).setTimeout('code')",
    "(window as any).execScript('code')",

    // ---- let / var with writes: not resolvable → not flagged ----
    "let s = 'x'; s = 'y'; setTimeout(s);",
    "var s = 'x'; s = 'y'; setTimeout(s);",

    // ---- Conditional with unresolvable cond → not flagged ----
    "setTimeout(c ? 'x' : 'y');",

    // ---- Logical / nullish with non-string short-circuit result → not flagged ----
    "setTimeout(null && 'x');",
    "setTimeout(1 ?? 'b');",
    // Purely numeric binary produces a number, not a string.
    'setTimeout(1 + 2);',
    'const a = 1; const b = 2; setTimeout(a + b);',

    // ---- String() with unresolvable arg → not flagged ----
    'setTimeout(String(x));',
    'setTimeout(String(x + y));',
    "setTimeout(Number('x'));",
    "setTimeout(Boolean('x'));",
    "function String(){} setTimeout(String('x'));",

    // ---- String.raw with unresolvable sub → not flagged ----
    'setTimeout(String.raw`x${y}`);',

    // ---- typeof on unresolvable operand → not flagged ----
    'setTimeout(typeof x);',
    'setTimeout(void 0);',

    // ---- const of number / undefined / null → not a string ----
    'const n = 1; setTimeout(n);',
    'const u = undefined; setTimeout(u);',
    'const z = null; setTimeout(z);',

    // ---- Property access on unresolvable receiver → not flagged ----
    'setTimeout(o.x);',
    "let o = { x: 'y' }; o = null; setTimeout(o.x);",
    'setTimeout({ x: y }.x);',
    'const o = { x: { y: z } }; setTimeout(o.x.y);',
    "setTimeout({ [k]: 'y' }.x);",

    // ---- String methods that eslint-utils does NOT fold → not flagged ----
    "setTimeout('x'.repeat(2));",
    "setTimeout('x'.replace('a', 'b'));",
    "setTimeout('x'.replaceAll('a', 'b'));",
    "setTimeout('x'.split('a'));",
    "setTimeout('x'.charCodeAt(0));",
    "setTimeout('x'.indexOf('a'));",
    "setTimeout('x'.startsWith('a'));",
    "setTimeout('x'.toLocaleString());",

    // ---- Array with unresolvable element → array unresolvable ----
    'setTimeout([x].toString());',
    "setTimeout([x].join(','));",

    // ---- Method on unresolvable receiver → not flagged ----
    'setTimeout(foo().toString());',
    'setTimeout((typeof x).toUpperCase());',
    "setTimeout(Array.from('x'));",
    'setTimeout(tag`x`);',
  ],
  invalid: [
    // ---- Direct calls with string literal ----
    {
      code: 'setTimeout("x = 1;");',
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: 'setTimeout("x = 1;", 100);',
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: 'setInterval("x = 1;");',
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: 'execScript("x = 1;");',
      errors: [{ messageId: 'execScript' }],
    },

    // ---- Const resolution and String(...) constructor ----
    {
      code: "const s = 'x=1'; setTimeout(s, 100);",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "setTimeout(String('x=1'), 100);",
      errors: [{ messageId: 'impliedEval' }],
    },

    // ---- window.* member access ----
    {
      code: "window.setTimeout('foo')",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "window.setInterval('foo')",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "window.execScript('foo')",
      errors: [{ messageId: 'execScript' }],
    },
    {
      code: "window['setTimeout']('foo')",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "window['setInterval']('foo')",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: 'window[`setInterval`]("foo")',
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "window['execScript']('foo')",
      errors: [{ messageId: 'execScript' }],
    },
    {
      code: 'window[`execScript`]("foo")',
      errors: [{ messageId: 'execScript' }],
    },

    // ---- Chained window.window ----
    {
      code: "window.window['setInterval']('foo')",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "window.window['execScript']('foo')",
      errors: [{ messageId: 'execScript' }],
    },

    // ---- global.* member access ----
    {
      code: "global.setTimeout('foo')",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "global.setInterval('foo')",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "global.execScript('foo')",
      errors: [{ messageId: 'execScript' }],
    },
    {
      code: "global['setTimeout']('foo')",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "global['setInterval']('foo')",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: 'global[`setInterval`]("foo")',
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "global['execScript']('foo')",
      errors: [{ messageId: 'execScript' }],
    },
    {
      code: 'global[`execScript`]("foo")',
      errors: [{ messageId: 'execScript' }],
    },
    {
      code: "global.global['setInterval']('foo')",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "global.global['execScript']('foo')",
      errors: [{ messageId: 'execScript' }],
    },

    // ---- globalThis.* member access ----
    {
      code: "globalThis.setTimeout('foo')",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "globalThis.setInterval('foo')",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "globalThis.execScript('foo')",
      errors: [{ messageId: 'execScript' }],
    },

    // ---- Template literal arguments ----
    {
      code: 'setTimeout(`foo${bar}`)',
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: 'window.setTimeout(`foo${bar}`)',
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: 'window.window.setTimeout(`foo${bar}`)',
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: 'global.global.setTimeout(`foo${bar}`)',
      errors: [{ messageId: 'impliedEval' }],
    },

    // ---- String concatenation arguments ----
    {
      code: "setTimeout('foo' + bar)",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "setTimeout(foo + 'bar')",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: 'setTimeout(`foo` + bar)',
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "setTimeout(1 + ';' + 1)",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "window.setTimeout('foo' + bar)",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "window.setTimeout(foo + 'bar')",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: 'window.setTimeout(`foo` + bar)',
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "window.setTimeout(1 + ';' + 1)",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "window.window.setTimeout(1 + ';' + 1)",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "global.setTimeout('foo' + bar)",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "global.setTimeout(foo + 'bar')",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: 'global.setTimeout(`foo` + bar)',
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "global.setTimeout(1 + ';' + 1)",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "global.global.setTimeout(1 + ';' + 1)",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "globalThis.setTimeout('foo' + bar)",
      errors: [{ messageId: 'impliedEval' }],
    },

    // ---- Nested reporting ----
    {
      code:
        "setTimeout('foo' + (function() {\n" +
        '   setTimeout(helper);\n' +
        "   execScript('str');\n" +
        "   return 'bar';\n" +
        '})())',
      errors: [{ messageId: 'impliedEval' }, { messageId: 'execScript' }],
    },
    {
      code:
        "window.setTimeout('foo' + (function() {\n" +
        '   setTimeout(helper);\n' +
        "   window.execScript('str');\n" +
        "   return 'bar';\n" +
        '})())',
      errors: [{ messageId: 'impliedEval' }, { messageId: 'execScript' }],
    },
    {
      code:
        "global.setTimeout('foo' + (function() {\n" +
        '   setTimeout(helper);\n' +
        "   global.execScript('str');\n" +
        "   return 'bar';\n" +
        '})())',
      errors: [{ messageId: 'impliedEval' }, { messageId: 'execScript' }],
    },

    // ---- Optional chaining ----
    {
      code: "window?.setTimeout('code', 0)",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "(window?.setTimeout)('code', 0)",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "window?.execScript('code')",
      errors: [{ messageId: 'execScript' }],
    },
    {
      code: "(window?.execScript)('code')",
      errors: [{ messageId: 'execScript' }],
    },

    // ---- self.* member access ----
    {
      code: "self.setTimeout('foo')",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "self.setInterval('foo')",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "self.execScript('foo')",
      errors: [{ messageId: 'execScript' }],
    },
    {
      code: "self['setTimeout']('foo')",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "self['setInterval']('foo')",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: 'self[`setInterval`]("foo")',
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "self['execScript']('foo')",
      errors: [{ messageId: 'execScript' }],
    },
    {
      code: 'self[`execScript`]("foo")',
      errors: [{ messageId: 'execScript' }],
    },
    {
      code: "self.self['setInterval']('foo')",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "self.self['execScript']('foo')",
      errors: [{ messageId: 'execScript' }],
    },
    {
      code: 'self.setTimeout(`foo${bar}`)',
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: 'self.self.setTimeout(`foo${bar}`)',
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "self.setTimeout('foo' + bar)",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "self.setTimeout(foo + 'bar')",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: 'self.setTimeout(`foo` + bar)',
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "self.setTimeout(1 + ';' + 1)",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "self.self.setTimeout(1 + ';' + 1)",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "self?.setTimeout('code', 0)",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "(self?.setTimeout)('code', 0)",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "self?.execScript('code')",
      errors: [{ messageId: 'execScript' }],
    },
    {
      code: "(self?.execScript)('code')",
      errors: [{ messageId: 'execScript' }],
    },

    // ---- Parenthesized callee / receiver ----
    { code: "(setTimeout)('code')", errors: [{ messageId: 'impliedEval' }] },
    { code: "((setTimeout))('code')", errors: [{ messageId: 'impliedEval' }] },
    {
      code: "(window).setTimeout('code')",
      errors: [{ messageId: 'impliedEval' }],
    },

    // ---- Parenthesized first argument ----
    { code: "setTimeout(('code'))", errors: [{ messageId: 'impliedEval' }] },

    // ---- No-substitution template ----
    { code: 'setTimeout(`code`)', errors: [{ messageId: 'impliedEval' }] },
    {
      code: 'window.setTimeout(`code`)',
      errors: [{ messageId: 'impliedEval' }],
    },

    // ---- Optional call ----
    { code: "setTimeout?.('code')", errors: [{ messageId: 'impliedEval' }] },
    { code: "execScript?.('code')", errors: [{ messageId: 'execScript' }] },

    // ---- TS outer expressions on first argument ----
    {
      code: "setTimeout('code' as any)",
      errors: [{ messageId: 'impliedEval' }],
    },
    { code: "setTimeout('code'!)", errors: [{ messageId: 'impliedEval' }] },

    // ---- Class member context ----
    {
      code: "class A { x = setTimeout('code'); }",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "class A { static { window.setTimeout('code'); } }",
      errors: [{ messageId: 'impliedEval' }],
    },

    // ---- IIFE ----
    {
      code: "(function() { setTimeout('code'); })()",
      errors: [{ messageId: 'impliedEval' }],
    },

    // ---- Multiple errors in one file ----
    {
      code: "setTimeout('a');\nsetInterval('b');\nexecScript('c');",
      errors: [
        { messageId: 'impliedEval' },
        { messageId: 'impliedEval' },
        { messageId: 'execScript' },
      ],
    },

    // ---- Binary + with identifier resolution ----
    {
      code: "const a = 'x'; const b = 'y'; setTimeout(a + b);",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "const a = 1; const b = 'y'; setTimeout(a + b);",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "const a = 'x'; const b = 'y'; const c = a + b; setTimeout(c);",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "const a = 'x'; const b = a; setTimeout(b);",
      errors: [{ messageId: 'impliedEval' }],
    },

    // ---- let / var with no writes ----
    {
      code: "let s = 'x'; setTimeout(s);",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "var s = 'x'; setTimeout(s);",
      errors: [{ messageId: 'impliedEval' }],
    },

    // ---- Conditional ----
    {
      code: "setTimeout(true ? 'x' : 'y');",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "setTimeout(false ? 1 : 'y');",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "const flag = true; setTimeout(flag ? 'x' : 'y');",
      errors: [{ messageId: 'impliedEval' }],
    },

    // ---- Logical ||, &&, ?? ----
    { code: "setTimeout('a' || 'b');", errors: [{ messageId: 'impliedEval' }] },
    {
      code: "setTimeout('' || 'fallback');",
      errors: [{ messageId: 'impliedEval' }],
    },
    { code: "setTimeout('a' && 'b');", errors: [{ messageId: 'impliedEval' }] },
    {
      code: "setTimeout(null ?? 'x');",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "setTimeout(undefined ?? 'x');",
      errors: [{ messageId: 'impliedEval' }],
    },
    { code: "setTimeout('a' ?? 'b');", errors: [{ messageId: 'impliedEval' }] },

    // ---- typeof on resolvable operand ----
    { code: "setTimeout(typeof 'x');", errors: [{ messageId: 'impliedEval' }] },
    {
      code: 'const n = 1; setTimeout(typeof n);',
      errors: [{ messageId: 'impliedEval' }],
    },

    // ---- String() with any resolvable arg, or no args ----
    { code: 'setTimeout(String(5));', errors: [{ messageId: 'impliedEval' }] },
    {
      code: 'setTimeout(String(undefined));',
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: 'setTimeout(String(null));',
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: 'setTimeout(String(true));',
      errors: [{ messageId: 'impliedEval' }],
    },
    { code: 'setTimeout(String());', errors: [{ messageId: 'impliedEval' }] },

    // ---- String.raw tagged templates ----
    {
      code: 'setTimeout(String.raw`x`);',
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "const y = 'z'; setTimeout(String.raw`x${y}`);",
      errors: [{ messageId: 'impliedEval' }],
    },

    // ---- Deep nesting: conditional / logical inside binary + with resolvable operand ----
    {
      code: "const b = 'z'; setTimeout((true ? 'x' : 'y') + b);",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "const b = 'z'; setTimeout((true ? ('a' || 'c') : 'y') + b);",
      errors: [{ messageId: 'impliedEval' }],
    },

    // ---- Property access on const / let / var object literal ----
    {
      code: "const o = { x: 'y' }; setTimeout(o.x);",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "const o = { x: 'y' }; setTimeout(o['x']);",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "let o = { x: 'y' }; setTimeout(o.x);",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "const o = { a: { b: { c: 'y' } } }; setTimeout(o.a.b.c);",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "setTimeout({ ['x']: 'y' }.x);",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "const x = 'y'; setTimeout({ x }.x);",
      errors: [{ messageId: 'impliedEval' }],
    },

    // ---- String.prototype method whitelist ----
    {
      code: "setTimeout('x'.toString());",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "setTimeout('x'.toUpperCase());",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "setTimeout('x'.toLowerCase());",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "setTimeout(' x'.trim());",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "setTimeout(' x'.trimStart());",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "setTimeout('x '.trimEnd());",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "setTimeout('x'.concat('y'));",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "setTimeout('x'.slice(0));",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "setTimeout('xy'.substring(0, 1));",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "setTimeout('x'.padStart(3));",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "setTimeout('x'.padEnd(3));",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "setTimeout('x'.charAt(0));",
      errors: [{ messageId: 'impliedEval' }],
    },
    { code: "setTimeout('x'.at(0));", errors: [{ messageId: 'impliedEval' }] },
    {
      code: "setTimeout('x'.normalize());",
      errors: [{ messageId: 'impliedEval' }],
    },

    // ---- Method chains ----
    {
      code: "setTimeout('x'.toUpperCase().toLowerCase());",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "const s = 'x'; setTimeout(s.toUpperCase());",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "const o = { x: 'y' }; setTimeout(o.x.toUpperCase());",
      errors: [{ messageId: 'impliedEval' }],
    },

    // ---- Number.prototype method whitelist ----
    {
      code: 'setTimeout((1).toString());',
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: 'setTimeout((1).toFixed(2));',
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: 'setTimeout((1).toExponential());',
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: 'setTimeout((1).toPrecision(3));',
      errors: [{ messageId: 'impliedEval' }],
    },

    // ---- Array literal + method ----
    {
      code: 'setTimeout([1, 2].toString());',
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: 'setTimeout([].toString());',
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "setTimeout([1, 2].join(','));",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: 'setTimeout([1, 2].slice(0).toString());',
      errors: [{ messageId: 'impliedEval' }],
    },

    // ---- TS non-null on const string via member ----
    {
      code: "const o = { x: 'y' }; setTimeout(o.x!);",
      errors: [{ messageId: 'impliedEval' }],
    },

    // ---- Computed property key resolved through const ----
    {
      code: "const key = 'x'; setTimeout({ [key]: 'y' }.x);",
      errors: [{ messageId: 'impliedEval' }],
    },
    {
      code: "const key = 'x'; const o = { x: 'y' }; setTimeout(o[key]);",
      errors: [{ messageId: 'impliedEval' }],
    },
  ],
});

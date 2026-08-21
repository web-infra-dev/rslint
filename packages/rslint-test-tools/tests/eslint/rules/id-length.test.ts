import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('id-length', {
  valid: [
    'var xyz;',
    'var xy = 1;',
    'function xyz() {};',
    'function xyz(abc, de) {};',
    'var obj = { abc: 1, de: 2 };',
    "var obj = { 'a': 1, bc: 2 };",
    "var obj = {}; obj['a'] = 2;",
    'abc = d;',
    'try { blah(); } catch (err) { /* pass */ }',
    'var handler = function ($e) {};',
    'var _a = 2',
    'var _ad$$ = new $;',
    'var xyz = new ΣΣ();',
    'unrelatedExpressionThatNeedsToBeIgnored();',
    "var obj = { 'a': 1, bc: 2 }; obj.tk = obj.a;",
    "var query = location.query.q || '';",
    "var query = location.query.q ? location.query.q : ''",
    'let {a: foo} = bar;',
    'let foo = { [a]: 1 };',
    'let foo = { [a + b]: 1 };',
    { code: 'var x = Foo(42)', options: { min: 1 } },
    { code: 'var x = Foo(42)', options: { min: 0 } },
    { code: 'foo.$x = Foo(42)', options: { min: 1 } },
    { code: 'var lalala = Foo(42)', options: { max: 6 } },
    {
      code: 'for (var q, h=0; h < 10; h++) { console.log(h); q++; }',
      options: { exceptions: ['h', 'q'] },
    },
    '(num) => { num * num };',
    'function foo(num = 0) { }',
    'class MyClass { }',
    'class Foo { method() {} }',
    'function foo(...args) { }',
    'var { prop } = {};',
    'var { [a]: prop } = {};',
    { code: 'var { a: foo } = {};', options: { min: 3 } },
    { code: 'var { prop: foo } = {};', options: { max: 3 } },
    { code: 'var { longName: foo } = {};', options: { min: 3, max: 5 } },
    { code: 'var { foo: a } = {};', options: { exceptions: ['a'] } },
    'var { a: { b: { c: longName } } } = {};',
    {
      code: '({ a: obj.x.y.z } = {});',
      options: { properties: 'never' },
    },
    "import something from 'y';",
    'export var num = 0;',
    "import * as something from 'y';",
    "import { x } from 'y';",
    "import { x as x } from 'y';",
    "import { 'x' as x } from 'y';",
    "import { x as foo } from 'y';",
    { code: "import { longName } from 'y';", options: { max: 5 } },
    { code: "import { x as bar } from 'y';", options: { max: 5 } },
    '({ prop: obj.x.y.something } = {});',
    '({ prop: obj.longName } = {});',
    { code: 'var obj = { a: 1, bc: 2 };', options: { properties: 'never' } },
    { code: 'var obj = { [a]: 2 };', options: { properties: 'never' } },
    {
      code: 'var obj = {}; obj.a = 1; obj.bc = 2;',
      options: { properties: 'never' },
    },
    { code: '({ prop: obj.x } = {});', options: { properties: 'never' } },
    {
      code: 'var obj = { aaaaa: 1 };',
      options: { max: 4, properties: 'never' },
    },
    {
      code: 'var obj = {}; obj.aaaaa = 1;',
      options: { max: 4, properties: 'never' },
    },
    {
      code: '({ a: obj.x.y.z } = {});',
      options: { max: 4, properties: 'never' },
    },
    {
      code: '({ prop: obj.xxxxx } = {});',
      options: { max: 4, properties: 'never' },
    },
    'var arr = [i,j,f,b]',
    'function foo([arr]) {}',
    { code: 'var {x} = foo;', options: { properties: 'never' } },
    { code: 'var {x, y: {z}} = foo;', options: { properties: 'never' } },
    { code: 'let foo = { [a]: 1 };', options: { properties: 'always' } },
    {
      code: 'let foo = { [a + b]: 1 };',
      options: { properties: 'always' },
    },
    {
      code: 'function BEFORE_send() {};',
      options: { min: 3, max: 5, exceptionPatterns: ['^BEFORE_'] },
    },
    {
      code: 'function BEFORE_send() {};',
      options: { min: 3, max: 5, exceptionPatterns: ['^BEFORE_', 'send$'] },
    },
    {
      code: 'function BEFORE_send() {};',
      options: { min: 3, max: 5, exceptionPatterns: ['^BEFORE_', '^A', '^Z'] },
    },
    {
      code: 'function BEFORE_send() {};',
      options: { min: 3, max: 5, exceptionPatterns: ['^A', '^BEFORE_', '^Z'] },
    },
    {
      code: 'var x = 1 ;',
      options: { min: 3, max: 5, exceptionPatterns: ['[x-z]'] },
    },

    // ---- Class Fields ----
    'class Foo { #xyz() {} }',
    'class Foo { xyz = 1 }',
    'class Foo { #xyz = 1 }',
    { code: 'class Foo { #abc() {} }', options: { max: 3 } },
    { code: 'class Foo { abc = 1 }', options: { max: 3 } },
    { code: 'class Foo { #abc = 1 }', options: { max: 3 } },

    // ---- Identifier consisting of two code units ----
    { code: 'var 𠮟 = 2', options: { min: 1, max: 1 } },
    { code: 'var 葛󠄀 = 2', options: { min: 1, max: 1 } }, // 2 code points but only 1 grapheme
    { code: 'var a = { 𐌘: 1 };', options: { min: 1, max: 1 } },
    { code: '(𐌘) => { 𐌘 * 𐌘 };', options: { min: 1, max: 1 } },
    { code: 'class 𠮟 { }', options: { min: 1, max: 1 } },
    { code: 'class F { 𐌘() {} }', options: { min: 1, max: 1 } },
    { code: 'class F { #𐌘() {} }', options: { min: 1, max: 1 } },
    { code: 'class F { 𐌘 = 1 }', options: { min: 1, max: 1 } },
    { code: 'class F { #𐌘 = 1 }', options: { min: 1, max: 1 } },
    { code: 'function f(...𐌘) { }', options: { min: 1, max: 1 } },
    { code: 'function f([𐌘]) { }', options: { min: 1, max: 1 } },
    { code: 'var [ 𐌘 ] = a;', options: { min: 1, max: 1 } },
    { code: 'var { p: [𐌘]} = {};', options: { min: 1, max: 1 } },
    { code: 'function f({𐌘}) { }', options: { min: 1, max: 1 } },
    { code: 'var { 𐌘 } = {};', options: { min: 1, max: 1 } },
    { code: 'var { p: 𐌘} = {};', options: { min: 1, max: 1 } },
    { code: '({ prop: o.𐌘 } = {});', options: { min: 1, max: 1 } },

    // ---- Import attribute keys ----
    {
      code: "import foo from 'foo.json' with { type: 'json' }",
      options: { min: 1, max: 3, properties: 'always' },
    },
    {
      code: "export * from 'foo.json' with { type: 'json' }",
      options: { min: 1, max: 3, properties: 'always' },
    },
    {
      code: "export { default } from 'foo.json' with { type: 'json' }",
      options: { min: 1, max: 3, properties: 'always' },
    },
    {
      code: "import('foo.json', { with: { type: 'json' } })",
      options: { min: 1, max: 3, properties: 'always' },
    },
    {
      code: "import('foo.json', { 'with': { type: 'json' } })",
      options: { min: 1, max: 3, properties: 'always' },
    },
    {
      code: "import('foo.json', { with: { type } })",
      options: { min: 1, max: 3, properties: 'always' },
    },
  ],
  invalid: [
    { code: 'var x = 1;', errors: [{ messageId: 'tooShort' }] },
    { code: 'var x;', errors: [{ messageId: 'tooShort' }] },
    { code: 'obj.e = document.body;', errors: [{ messageId: 'tooShort' }] },
    { code: 'function x() {};', errors: [{ messageId: 'tooShort' }] },
    { code: 'function xyz(a) {};', errors: [{ messageId: 'tooShort' }] },
    {
      code: 'var obj = { a: 1, bc: 2 };',
      errors: [{ messageId: 'tooShort' }],
    },
    {
      code: 'try { blah(); } catch (e) { /* pass */ }',
      errors: [{ messageId: 'tooShort' }],
    },
    {
      code: 'var handler = function (e) {};',
      errors: [{ messageId: 'tooShort' }],
    },
    {
      code: 'for (var i=0; i < 10; i++) { console.log(i); }',
      errors: [{ messageId: 'tooShort' }],
    },
    {
      code: 'var j=0; while (j > -10) { console.log(--j); }',
      errors: [{ messageId: 'tooShort' }],
    },
    { code: 'var [i] = arr;', errors: [{ messageId: 'tooShort' }] },
    {
      code: 'var [,i,a] = arr;',
      errors: [{ messageId: 'tooShort' }, { messageId: 'tooShort' }],
    },
    { code: 'function foo([a]) {}', errors: [{ messageId: 'tooShort' }] },
    {
      code: "import x from 'module';",
      errors: [{ messageId: 'tooShort' }],
    },
    {
      code: "import { x as z } from 'module';",
      errors: [{ messageId: 'tooShort', column: 15 }],
    },
    {
      code: "import { foo as z } from 'module';",
      errors: [{ messageId: 'tooShort', column: 17 }],
    },
    {
      code: "import { 'foo' as z } from 'module';",
      errors: [{ messageId: 'tooShort', column: 19 }],
    },
    {
      code: "import * as x from 'module';",
      errors: [{ messageId: 'tooShort' }],
    },
    {
      code: "import longName from 'module';",
      options: { max: 5 },
      errors: [{ messageId: 'tooLong' }],
    },
    {
      code: "import * as longName from 'module';",
      options: { max: 5 },
      errors: [{ messageId: 'tooLong' }],
    },
    {
      code: "import { foo as longName } from 'module';",
      options: { max: 5 },
      errors: [{ messageId: 'tooLong', column: 17 }],
    },
    {
      code: 'var _$xt_$ = Foo(42)',
      options: { min: 2, max: 4 },
      errors: [{ messageId: 'tooLong' }],
    },
    {
      code: 'var _$x$_t$ = Foo(42)',
      options: { min: 2, max: 4 },
      errors: [{ messageId: 'tooLong' }],
    },
    {
      code: 'var toString;',
      options: { max: 5 },
      errors: [{ messageId: 'tooLong' }],
    },
    { code: '(a) => { a * a };', errors: [{ messageId: 'tooShort' }] },
    {
      code: 'function foo(x = 0) { }',
      errors: [{ messageId: 'tooShort' }],
    },
    { code: 'class x { }', errors: [{ messageId: 'tooShort' }] },
    { code: 'class Foo { x() {} }', errors: [{ messageId: 'tooShort' }] },
    {
      code: 'function foo(...x) { }',
      errors: [{ messageId: 'tooShort' }],
    },
    { code: 'function foo({x}) { }', errors: [{ messageId: 'tooShort' }] },
    {
      code: 'function foo({x: a}) { }',
      errors: [{ messageId: 'tooShort' }],
    },
    {
      code: 'function foo({x: a, longName}) { }',
      errors: [{ messageId: 'tooShort' }],
    },
    {
      code: 'function foo({ longName: a }) {}',
      options: { min: 3, max: 5 },
      errors: [{ messageId: 'tooShort' }],
    },
    {
      code: 'function foo({ prop: longName }) {};',
      options: { min: 3, max: 5 },
      errors: [{ messageId: 'tooLong' }],
    },
    {
      code: 'function foo({ a: b }) {};',
      options: { exceptions: ['a'] },
      errors: [{ messageId: 'tooShort', line: 1, column: 19 }],
    },
    {
      code: 'var hasOwnProperty;',
      options: { max: 10, exceptions: [] },
      errors: [{ messageId: 'tooLong', line: 1, column: 5 }],
    },
    {
      code: 'function foo({ a: { b: { c: d, e } } }) { }',
      errors: [
        { messageId: 'tooShort', line: 1, column: 29 },
        { messageId: 'tooShort', line: 1, column: 32 },
      ],
    },
    { code: 'var { x} = {};', errors: [{ messageId: 'tooShort' }] },
    { code: 'var { x: a} = {};', errors: [{ messageId: 'tooShort' }] },
    { code: 'var { a: a} = {};', errors: [{ messageId: 'tooShort' }] },
    { code: 'var { prop: a } = {};', errors: [{ messageId: 'tooShort' }] },
    {
      code: 'var { longName: a } = {};',
      options: { min: 3, max: 5 },
      errors: [{ messageId: 'tooShort' }],
    },
    {
      code: 'var { prop: [x] } = {};',
      errors: [{ messageId: 'tooShort' }],
    },
    {
      code: 'var { prop: [[x]] } = {};',
      errors: [{ messageId: 'tooShort' }],
    },
    {
      code: 'var { prop: longName } = {};',
      options: { min: 3, max: 5 },
      errors: [{ messageId: 'tooLong', line: 1, column: 13 }],
    },
    {
      code: 'var { x: a} = {};',
      options: { exceptions: ['x'] },
      errors: [{ messageId: 'tooShort', line: 1, column: 10 }],
    },
    {
      code: 'var { a: { b: { c: d } } } = {};',
      errors: [{ messageId: 'tooShort', line: 1, column: 20 }],
    },
    {
      code: 'var { a: { b: { c: d, e } } } = {};',
      errors: [
        { messageId: 'tooShort', line: 1, column: 20 },
        { messageId: 'tooShort', line: 1, column: 23 },
      ],
    },
    {
      code: 'var { a: { b: { c, e: longName } } } = {};',
      errors: [{ messageId: 'tooShort', line: 1, column: 17 }],
    },
    {
      code: 'var { a: { b: { c: d, e: longName } } } = {};',
      errors: [{ messageId: 'tooShort', line: 1, column: 20 }],
    },
    {
      code: 'var { a, b: { c: d, e: longName } } = {};',
      errors: [
        { messageId: 'tooShort', line: 1, column: 7 },
        { messageId: 'tooShort', line: 1, column: 18 },
      ],
    },
    {
      code: "import x from 'y';",
      errors: [{ messageId: 'tooShort' }],
    },
    {
      code: 'export var x = 0;',
      errors: [{ messageId: 'tooShort' }],
    },
    {
      code: '({ a: obj.x.y.z } = {});',
      errors: [{ messageId: 'tooShort', line: 1, column: 15 }],
    },
    {
      code: '({ prop: obj.x } = {});',
      errors: [{ messageId: 'tooShort', line: 1, column: 14 }],
    },
    {
      code: 'var x = 1;',
      options: { properties: 'never' },
      errors: [{ messageId: 'tooShort' }],
    },
    {
      code: 'var {prop: x} = foo;',
      options: { properties: 'never' },
      errors: [{ messageId: 'tooShort', line: 1, column: 12 }],
    },
    {
      code: 'var foo = {x: prop};',
      options: { properties: 'always' },
      errors: [{ messageId: 'tooShort' }],
    },
    {
      code: 'function BEFORE_send() {};',
      options: { min: 3, max: 5 },
      errors: [{ messageId: 'tooLong' }],
    },
    {
      code: 'function NOTMATCHED_send() {};',
      options: { min: 3, max: 5, exceptionPatterns: ['^BEFORE_'] },
      errors: [{ messageId: 'tooLong' }],
    },
    {
      code: 'function N() {};',
      options: { min: 3, max: 5, exceptionPatterns: ['^BEFORE_'] },
      errors: [{ messageId: 'tooShort' }],
    },

    // ---- Class Fields ----
    {
      code: 'class Foo { #x() {} }',
      errors: [{ messageId: 'tooShortPrivate' }],
    },
    { code: 'class Foo { x = 1 }', errors: [{ messageId: 'tooShort' }] },
    {
      code: 'class Foo { #x = 1 }',
      errors: [{ messageId: 'tooShortPrivate' }],
    },
    {
      code: 'class Foo { #abcdefg() {} }',
      options: { max: 3 },
      errors: [{ messageId: 'tooLongPrivate' }],
    },
    {
      code: 'class Foo { abcdefg = 1 }',
      options: { max: 3 },
      errors: [{ messageId: 'tooLong' }],
    },
    {
      code: 'class Foo { #abcdefg = 1 }',
      options: { max: 3 },
      errors: [{ messageId: 'tooLongPrivate' }],
    },

    // ---- Identifier consisting of two code units ----
    { code: 'var 𠮟 = 2', errors: [{ messageId: 'tooShort' }] },
    { code: 'var 葛󠄀 = 2', errors: [{ messageId: 'tooShort' }] }, // 2 code points but only 1 grapheme
    {
      code: 'var myObj = { 𐌘: 1 };',
      errors: [{ messageId: 'tooShort' }],
    },
    { code: '(𐌘) => { 𐌘 * 𐌘 };', errors: [{ messageId: 'tooShort' }] },
    { code: 'class 𠮟 { }', errors: [{ messageId: 'tooShort' }] },
    {
      code: 'class Foo { 𐌘() {} }',
      errors: [{ messageId: 'tooShort' }],
    },
    {
      code: 'class Foo1 { #𐌘() {} }',
      errors: [{ messageId: 'tooShortPrivate' }],
    },
    {
      code: 'class Foo2 { 𐌘 = 1 }',
      errors: [{ messageId: 'tooShort' }],
    },
    {
      code: 'class Foo3 { #𐌘 = 1 }',
      errors: [{ messageId: 'tooShortPrivate' }],
    },
    {
      code: 'function foo1(...𐌘) { }',
      errors: [{ messageId: 'tooShort' }],
    },
    {
      code: 'function foo([𐌘]) { }',
      errors: [{ messageId: 'tooShort' }],
    },
    { code: 'var [ 𐌘 ] = arr;', errors: [{ messageId: 'tooShort' }] },
    {
      code: 'var { prop: [𐌘]} = {};',
      errors: [{ messageId: 'tooShort' }],
    },
    {
      code: 'function foo({𐌘}) { }',
      errors: [{ messageId: 'tooShort' }],
    },
    { code: 'var { 𐌘 } = {};', errors: [{ messageId: 'tooShort' }] },
    {
      code: 'var { prop: 𐌘} = {};',
      errors: [{ messageId: 'tooShort' }],
    },
    {
      code: '({ prop: obj.𐌘 } = {});',
      errors: [{ messageId: 'tooShort' }],
    },
  ],
});

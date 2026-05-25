import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const OPENING = 'foo(function() {';
const CLOSING = '});';
const nestFunctions = (times: number) =>
  OPENING.repeat(times) + CLOSING.repeat(times);

ruleTester.run('max-nested-callbacks', {
  valid: [
    {
      code: 'foo(function() { bar(thing, function(data) {}); });',
      options: [3] as any,
    },
    {
      code: 'var foo = function() {}; bar(function(){ baz(function() { qux(foo); }) });',
      options: [2] as any,
    },
    {
      code: 'fn(function(){}, function(){}, function(){});',
      options: [2] as any,
    },
    {
      code: 'fn(() => {}, function(){}, function(){});',
      options: [2] as any,
    },
    nestFunctions(10),
    {
      code: 'foo(function() { bar(thing, function(data) {}); });',
      options: [{ max: 3 }] as any,
    },

    // Function-likes that are NOT direct CallExpression children — never push.
    {
      code: 'var a = function() { var b = function() { var c = function() {}; }; };',
      options: [0] as any,
    },
    {
      code: 'new Foo(function() { new Bar(function() {}); });',
      options: [0] as any,
    },
    'function a() { function b() { function c() {} } }',
    {
      code: 'class C { m() { class D { m() { class E { m() {} } } } } }',
      options: [0] as any,
    },
    {
      code: 'var o = { m() { return { m() { return { m() {} }; } }; } };',
      options: [0] as any,
    },

    // Paren-walked parent — IIFE callee and paren-wrapped arguments.
    { code: '(function() {})();', options: [1] as any },
    { code: '(((function() {})))();', options: [1] as any },
    { code: 'foo((function() {}));', options: [1] as any },

    // Optional chain calls push like normal calls.
    { code: 'foo?.(function() {});', options: [1] as any },
    { code: 'obj?.method(function() {});', options: [1] as any },

    // TS expression wrappers around the callback DON'T push.
    { code: 'foo(function() {} as any);', options: [0] as any },
    { code: 'foo(function() {} satisfies any);', options: [0] as any },

    // TS wrappers on the callee — function still in arguments, pushes.
    { code: '(foo as any)(function() {});', options: [1] as any },
    { code: 'foo!(function() {});', options: [1] as any },
    { code: 'foo<T>(function() {});', options: [1] as any },

    // async / generator / async generator and async arrow.
    { code: 'foo(async function() {});', options: [1] as any },
    { code: 'foo(function*() {});', options: [1] as any },
    { code: 'foo(async function*() {});', options: [1] as any },
    { code: 'foo(async () => {});', options: [1] as any },

    // Sibling callback after a class / object method — the method's exit pop
    // clears the outer frame, matching ESLint's exit-pop semantics. Without
    // the method-exit listener rslint would over-report here.
    {
      code: 'outer(function() { class C { m() {} } bar(function() {}); });',
      options: [1] as any,
    },
    {
      code: 'outer(function() { ({ m() {} }); bar(function() {}); });',
      options: [1] as any,
    },
    {
      code: 'outer(function() { class C { get x() {} } bar(function() {}); });',
      options: [1] as any,
    },
    {
      code: 'outer(function() { class C { constructor() {} } bar(function() {}); });',
      options: [1] as any,
    },

    // Container nesting — depth carries across control-flow constructs.
    {
      code: 'foo(function() { if (a) { bar(function() {}); } else { baz(function() {}); } });',
      options: [2] as any,
    },
    {
      code: 'foo(function() { for (const v of a) { bar(function() {}); } });',
      options: [2] as any,
    },
    {
      code: 'foo(function() { try { bar(function() {}); } catch (e) {} });',
      options: [2] as any,
    },
    {
      code: 'foo(async function() { await bar(function() {}); });',
      options: [2] as any,
    },

    // Concise-body arrow with call.
    { code: 'var f = () => bar(function() {});', options: [2] as any },

    // Real-world: jQuery / AJAX / iteration.
    {
      code: `$(function() {
  $.ajax({
    success: function(data) { handle(data); },
  });
});`,
      options: [2] as any,
    },
    // Real-world: Express middleware.
    {
      code: "app.get('/x', function(req, res, next) { next(); });",
      options: [1] as any,
    },
    // Real-world: Node fs.readFile.
    {
      code: "fs.readFile('a', function(err, data) { cb(null, data); });",
      options: [1] as any,
    },
  ],
  invalid: [
    {
      code: 'foo(function() { bar(thing, function(data) { baz(function() {}); }); });',
      options: [2] as any,
      errors: [{ messageId: 'exceed', line: 1, column: 50 }],
    },
    {
      code: 'foo(function() { bar(thing, (data) => { baz(function() {}); }); });',
      options: [2] as any,
      errors: [{ messageId: 'exceed', line: 1, column: 45 }],
    },
    {
      code: 'foo(() => { bar(thing, (data) => { baz( () => {}); }); });',
      options: [2] as any,
      errors: [{ messageId: 'exceed', line: 1, column: 41 }],
    },
    {
      code: 'foo(function() { if (isTrue) { bar(function(data) { baz(function() {}); }); } });',
      options: [2] as any,
      errors: [{ messageId: 'exceed', line: 1, column: 57 }],
    },
    {
      code: nestFunctions(11),
      errors: [{ messageId: 'exceed', line: 1, column: 165 }],
    },
    {
      code: nestFunctions(11),
      options: [{}] as any,
      errors: [{ messageId: 'exceed', line: 1, column: 165 }],
    },
    {
      code: 'foo(function() {})',
      options: [{ max: 0 }] as any,
      errors: [{ messageId: 'exceed', line: 1, column: 5 }],
    },
    {
      code: 'foo(function() { bar(thing, function(data) { baz(function() {}); }); });',
      options: [{ max: 2 }] as any,
      errors: [{ messageId: 'exceed', line: 1, column: 50 }],
    },
    // IIFE callee + nested callbacks.
    {
      code: '(function() { foo(function() { bar(function() {}); }); })();',
      options: [2] as any,
      errors: [{ messageId: 'exceed', line: 1, column: 36 }],
    },
    // Paren-wrapped arguments still trigger via paren-walk.
    {
      code: 'foo((function() { bar((function() { baz((function() {})); })); }));',
      options: [2] as any,
      errors: [{ messageId: 'exceed', line: 1, column: 42 }],
    },
    // Mixed FunctionExpression / ArrowFunction.
    {
      code: 'foo(() => bar(function() { baz(() => {}); }));',
      options: [2] as any,
      errors: [{ messageId: 'exceed', line: 1, column: 32 }],
    },
    // Optional chain calls in callback chains.
    {
      code: 'foo?.(function() { bar?.(function() { baz?.(function() {}); }); });',
      options: [2] as any,
      errors: [{ messageId: 'exceed', line: 1, column: 45 }],
    },
    // Legacy `maximum` key.
    {
      code: 'foo(function() { bar(function() { baz(function() {}); }); });',
      options: [{ maximum: 2 }] as any,
      errors: [{ messageId: 'exceed', line: 1, column: 39 }],
    },
    // `maximum` wins when both are present and `maximum` is truthy.
    {
      code: 'foo(function() { bar(function() { baz(function() {}); }); });',
      options: [{ maximum: 2, max: 5 }] as any,
      errors: [{ messageId: 'exceed', line: 1, column: 39 }],
    },
    // Real-world: Express deeply nested.
    {
      code: `app.get('/x', function(req, res) {
  db.connect(function(err, client) {
    client.query('SELECT', function(err, rows) {
      rows.forEach(function(row) {
        process(row);
      });
    });
  });
});`,
      options: [3] as any,
      errors: [{ messageId: 'exceed', line: 4, column: 20 }],
    },
    // Real-world: Node.js callback hell.
    {
      code: `fs.readFile('a', function(err, a) {
  fs.readFile('b', function(err, b) {
    fs.readFile('c', function(err, c) {
      fs.writeFile('out', a + b + c, function(err) {});
    });
  });
});`,
      options: [2] as any,
      errors: [
        { messageId: 'exceed', line: 3, column: 22 },
        { messageId: 'exceed', line: 4, column: 38 },
      ],
    },
    // Real-world: setTimeout pyramid.
    {
      code: `setTimeout(function() {
  setTimeout(function() {
    setTimeout(function() {
      setTimeout(function() {});
    }, 10);
  }, 10);
}, 10);`,
      options: [3] as any,
      errors: [{ messageId: 'exceed', line: 4, column: 18 }],
    },
    // Concise-body arrow chain.
    {
      code: 'foo(() => bar(() => baz(() => qux(() => {}))));',
      options: [3] as any,
      errors: [{ messageId: 'exceed', line: 1, column: 35 }],
    },
    // Optional chain combined with paren-wrapped function arg.
    {
      code: 'foo?.(function() { bar?.((function() { baz?.((function() {})); })); });',
      options: [2] as any,
      errors: [{ messageId: 'exceed', line: 1, column: 47 }],
    },
    // Mixed paren wraps in IIFE chain.
    {
      code: '(((function() { (((function() { (((function() {})))(); })))(); })))();',
      options: [2] as any,
      errors: [{ messageId: 'exceed', line: 1, column: 36 }],
    },
    // Object spread containing function — only the trailing direct arg pushes.
    {
      code: 'foo({...x, cb: function(){}}, function(){});',
      options: [0] as any,
      errors: [{ messageId: 'exceed', line: 1, column: 31 }],
    },
  ],
});

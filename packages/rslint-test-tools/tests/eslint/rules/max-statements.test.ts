import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('max-statements', {
  valid: [
    {
      code: 'function foo() { var bar = 1; function qux () { var noCount = 2; } return 3; }',
      options: [3] as any,
    },
    {
      code: 'function foo() { var bar = 1; if (true) { for (;;) { var qux = null; } } else { quxx(); } return 3; }',
      options: [6] as any,
    },
    {
      code: 'function foo() { var x = 5; function bar() { var y = 6; } bar(); z = 10; baz(); }',
      options: [5] as any,
    },
    'function foo() { var a; var b; var c; var x; var y; var z; bar(); baz(); qux(); quxx(); }',
    {
      code: '(function() { var bar = 1; return function () { return 42; }; })()',
      options: [1, { ignoreTopLevelFunctions: true }] as any,
    },
    {
      code: 'function foo() { var bar = 1; var baz = 2; }',
      options: [1, { ignoreTopLevelFunctions: true }] as any,
    },
    {
      code: "define(['foo', 'qux'], function(foo, qux) { var bar = 1; var baz = 2; })",
      options: [1, { ignoreTopLevelFunctions: true }] as any,
    },

    {
      code: 'var foo = { thing: function() { var bar = 1; var baz = 2; } }',
      options: [2] as any,
    },
    {
      code: 'var foo = { thing() { var bar = 1; var baz = 2; } }',
      options: [2] as any,
    },
    {
      code: "var foo = { ['thing']() { var bar = 1; var baz = 2; } }",
      options: [2] as any,
    },
    {
      code: 'var foo = { thing: () => { var bar = 1; var baz = 2; } }',
      options: [2] as any,
    },
    {
      code: 'var foo = { thing: function() { var bar = 1; var baz = 2; } }',
      options: [{ max: 2 }] as any,
    },

    {
      code: 'class C { static { one; two; three; { four; five; six; } } }',
      options: [2] as any,
    },
    {
      code: 'function foo() { class C { static { one; two; three; { four; five; six; } } } }',
      options: [2] as any,
    },
    {
      code: 'class C { static { one; two; three; function foo() { 1; 2; } four; five; six; } }',
      options: [2] as any,
    },
    {
      code: 'class C { static { { one; two; three; function foo() { 1; 2; } four; five; six; } } }',
      options: [2] as any,
    },
    {
      code: 'function top_level() { 1; /* 2 */ class C { static { one; two; three; { four; five; six; } } } 3;}',
      options: [2, { ignoreTopLevelFunctions: true }] as any,
    },
    {
      code: 'function top_level() { 1; 2; } class C { static { one; two; three; { four; five; six; } } }',
      options: [1, { ignoreTopLevelFunctions: true }] as any,
    },
    {
      code: 'class C { static { one; two; three; { four; five; six; } } } function top_level() { 1; 2; } ',
      options: [1, { ignoreTopLevelFunctions: true }] as any,
    },
    {
      code: 'function foo() { let one; let two = class { static { let three; let four; let five; if (six) { let seven; let eight; let nine; } } }; }',
      options: [2] as any,
    },
  ],
  invalid: [
    {
      code: 'function foo() { var bar = 1; var baz = 2; var qux = 3; }',
      options: [2] as any,
      errors: [{ messageId: 'exceed', line: 1, column: 1 }],
    },
    {
      code: 'var foo = () => { var bar = 1; var baz = 2; var qux = 3; };',
      options: [2] as any,
      errors: [{ messageId: 'exceed', line: 1, column: 14 }],
    },
    {
      code: 'var foo = function() { var bar = 1; var baz = 2; var qux = 3; };',
      options: [2] as any,
      errors: [{ messageId: 'exceed', line: 1, column: 11 }],
    },
    {
      code: 'function foo() { var bar = 1; if (true) { while (false) { var qux = null; } } return 3; }',
      options: [4] as any,
      errors: [{ messageId: 'exceed', line: 1, column: 1 }],
    },
    {
      code: 'function foo() { var bar = 1; if (true) { for (;;) { var qux = null; } } return 3; }',
      options: [4] as any,
      errors: [{ messageId: 'exceed', line: 1, column: 1 }],
    },
    {
      code: 'function foo() { var bar = 1; if (true) { for (;;) { var qux = null; } } else { quxx(); } return 3; }',
      options: [5] as any,
      errors: [{ messageId: 'exceed', line: 1, column: 1 }],
    },
    {
      code: 'function foo() { var x = 5; function bar() { var y = 6; } bar(); z = 10; baz(); }',
      options: [3] as any,
      errors: [{ messageId: 'exceed', line: 1, column: 1 }],
    },
    {
      code: 'function foo() { var x = 5; function bar() { var y = 6; } bar(); z = 10; baz(); }',
      options: [4] as any,
      errors: [{ messageId: 'exceed', line: 1, column: 1 }],
    },
    {
      code: ';(function() { var bar = 1; return function () { var z; return 42; }; })()',
      options: [1, { ignoreTopLevelFunctions: true }] as any,
      errors: [{ messageId: 'exceed', line: 1, column: 36 }],
    },
    {
      code: ';(function() { var bar = 1; var baz = 2; })(); (function() { var bar = 1; var baz = 2; })()',
      options: [1, { ignoreTopLevelFunctions: true }] as any,
      errors: [
        { messageId: 'exceed', line: 1, column: 3 },
        { messageId: 'exceed', line: 1, column: 49 },
      ],
    },
    {
      code: "define(['foo', 'qux'], function(foo, qux) { var bar = 1; var baz = 2; return function () { var z; return 42; }; })",
      options: [1, { ignoreTopLevelFunctions: true }] as any,
      errors: [{ messageId: 'exceed', line: 1, column: 78 }],
    },
    {
      code: 'function foo() { var a; var b; var c; var x; var y; var z; bar(); baz(); qux(); quxx(); foo(); }',
      errors: [{ messageId: 'exceed', line: 1, column: 1 }],
    },

    {
      code: 'var foo = { thing: function() { var bar = 1; var baz = 2; var baz2; } }',
      options: [2] as any,
      errors: [{ messageId: 'exceed', line: 1, column: 13 }],
    },
    {
      code: 'var foo = { thing() { var bar = 1; var baz = 2; var baz2; } }',
      options: [2] as any,
      errors: [{ messageId: 'exceed', line: 1, column: 13 }],
    },
    {
      code: 'var foo = { thing: () => { var bar = 1; var baz = 2; var baz2; } }',
      options: [2] as any,
      errors: [{ messageId: 'exceed', line: 1, column: 13 }],
    },
    {
      code: 'var foo = { thing: function() { var bar = 1; var baz = 2; var baz2; } }',
      options: [{ max: 2 }] as any,
      errors: [{ messageId: 'exceed', line: 1, column: 13 }],
    },
    {
      code: 'function foo() { 1; 2; 3; 4; 5; 6; 7; 8; 9; 10; 11; }',
      options: [{}] as any,
      errors: [{ messageId: 'exceed', line: 1, column: 1 }],
    },
    {
      code: 'function foo() { 1; }',
      options: [{ max: 0 }] as any,
      errors: [{ messageId: 'exceed', line: 1, column: 1 }],
    },
    {
      code: 'function foo() { foo_1; /* foo_ 2 */ class C { static { one; two; three; four; { five; six; seven; eight; } } } foo_3 }',
      options: [2] as any,
      errors: [{ messageId: 'exceed', line: 1, column: 1 }],
    },
    {
      code: 'class C { static { one; two; three; four; function not_top_level() { 1; 2; 3; } five; six; seven; eight; } }',
      options: [2, { ignoreTopLevelFunctions: true }] as any,
      errors: [{ messageId: 'exceed', line: 1, column: 43 }],
    },
    {
      code: 'class C { static { { one; two; three; four; function not_top_level() { 1; 2; 3; } five; six; seven; eight; } } }',
      options: [2, { ignoreTopLevelFunctions: true }] as any,
      errors: [{ messageId: 'exceed', line: 1, column: 45 }],
    },
    {
      code: 'class C { static { { one; two; three; four; } function not_top_level() { 1; 2; 3; } { five; six; seven; eight; } } }',
      options: [2, { ignoreTopLevelFunctions: true }] as any,
      errors: [{ messageId: 'exceed', line: 1, column: 47 }],
    },
  ],
});

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const unexpected = (count = 1) =>
  Array.from({ length: count }, () => ({ messageId: 'unexpected' }));

const invalid = (
  code: string,
  count = 1,
  options?: Record<string, unknown>,
) => ({
  code,
  errors: unexpected(count),
  ...(options ? { options } : {}),
});

ruleTester.run('no-else-return', {
  valid: [
    'function foo() { if (true) { if (false) { return x; } } else { return y; } }',
    'function foo() { if (true) { return x; } return y; }',
    'function foo() { if (true) { for (;;) { return x; } } else { return y; } }',
    'function foo() { var x = true; if (x) { return x; } else if (x === false) { return false; } }',
    'function foo() { if (true) notAReturn(); else return y; }',
    'function foo() {if (x) { notAReturn(); } else if (y) { return true; } else { notAReturn(); } }',
    'function foo() {if (x) { return true; } else if (y) { notAReturn() } else { notAReturn(); } }',
    'if (0) { if (0) {} else {} } else {}',
    `
            function foo() {
                if (foo)
                    if (bar) return;
                    else baz;
                else qux;
            }
        `,
    `
            function foo() {
                while (foo)
                    if (bar) return;
                    else baz;
            }
        `,
    {
      code: 'function foo19() { if (true) { return x; } else if (false) { return y; } }',
      options: { allowElseIf: true },
    },
    {
      code: 'function foo20() {if (x) { return true; } else if (y) { notAReturn() } else { notAReturn(); } }',
      options: { allowElseIf: true },
    },
    {
      code: 'function foo21() { var x = true; if (x) { return x; } else if (x === false) { return false; } }',
      options: { allowElseIf: true },
    },
  ],
  invalid: [
    invalid('function foo1() { if (true) { return x; } else { return y; } }'),
    invalid(
      'function foo2() { if (true) { var x = bar; return x; } else { var y = baz; return y; } }',
    ),
    invalid('function foo3() { if (true) return x; else return y; }'),
    invalid(
      'function foo4() { if (true) { if (false) return x; else return y; } else { return z; } }',
      2,
    ),
    invalid(
      'function foo5() { if (true) { if (false) { if (true) return x; else { w = y; } } else { w = x; } } else { return z; } }',
    ),
    invalid(
      'function foo6() { if (true) { if (false) { if (true) return x; else return y; } } else { return z; } }',
    ),
    invalid(
      'function foo7() { if (true) { if (false) { if (true) return x; else return y; } return w; } else { return z; } }',
      2,
    ),
    invalid(
      'function foo8() { if (true) { if (false) { if (true) return x; else return y; } else { w = x; } } else { return z; } }',
      2,
    ),
    invalid(
      'function foo9() {if (x) { return true; } else if (y) { return true; } else { notAReturn(); } }',
    ),
    invalid(
      'function foo9a() {if (x) { return true; } else if (y) { return true; } else { notAReturn(); } }',
      1,
      { allowElseIf: false },
    ),
    invalid(
      'function foo9b() {if (x) { return true; } if (y) { return true; } else { notAReturn(); } }',
      1,
      { allowElseIf: false },
    ),
    invalid('function foo10() { if (foo) return bar; else (foo).bar(); }'),
    invalid(
      'function foo11() { if (foo) return bar\nelse { [1, 2, 3].map(foo) } }',
    ),
    invalid(
      'function foo12() { if (foo) return bar\nelse { baz() }\n[1, 2, 3].map(foo) }',
    ),
    invalid(
      'function foo13() { if (foo) return bar;\nelse { [1, 2, 3].map(foo) } }',
    ),
    invalid(
      'function foo14() { if (foo) return bar\nelse { baz(); }\n[1, 2, 3].map(foo) }',
    ),
    invalid('function foo15() { if (foo) return bar; else { baz() } qaz() }'),
    invalid('function foo16() { if (foo) return bar\nelse { baz() } qaz() }'),
    invalid('function foo17() { if (foo) return bar\nelse { baz() }\nqaz() }'),
    invalid(
      'function foo18() { if (foo) return function() {}\nelse [1, 2, 3].map(bar) }',
    ),
    invalid(
      'function foo19() { if (true) { return x; } else if (false) { return y; } }',
      1,
      { allowElseIf: false },
    ),
    invalid(
      'function foo20() {if (x) { return true; } else if (y) { notAReturn() } else { notAReturn(); } }',
      1,
      { allowElseIf: false },
    ),
    invalid(
      'function foo21() { var x = true; if (x) { return x; } else if (x === false) { return false; } }',
      1,
      { allowElseIf: false },
    ),

    // https://github.com/eslint/eslint/issues/11069
    invalid(
      'function foo() { var a; if (bar) { return true; } else { var a; } }',
    ),
    invalid(
      'function foo() { if (bar) { var a; if (baz) { return true; } else { var a; } } }',
    ),
    invalid(
      'function foo() { let a; if (bar) { return true; } else { let a; } }',
    ),
    invalid(
      'class foo { bar() { let a; if (baz) { return true; } else { let a; } } }',
    ),
    invalid(
      'function foo() { if (bar) { let a; if (baz) { return true; } else { let a; } } }',
    ),
    invalid(
      'function foo() {let a; if (bar) { if (baz) { return true; } else { let a; } } }',
    ),
    invalid(
      'function foo() { const a = 1; if (bar) { return true; } else { let a; } }',
    ),
    invalid(
      'function foo() { if (bar) { const a = 1; if (baz) { return true; } else { let a; } } }',
    ),
    invalid(
      'function foo() { let a; if (bar) { return true; } else { const a = 1 } }',
    ),
    invalid(
      'function foo() { if (bar) { let a; if (baz) { return true; } else { const a = 1; } } }',
    ),
    invalid(
      'function foo() { class a {}; if (bar) { return true; } else { const a = 1; } }',
    ),
    invalid(
      'function foo() { if (bar) { class a {}; if (baz) { return true; } else { const a = 1; } } }',
    ),
    invalid(
      'function foo() { const a = 1; if (bar) { return true; } else { class a {} } }',
    ),
    invalid(
      'function foo() { if (bar) { const a = 1; if (baz) { return true; } else { class a {} } } }',
    ),
    invalid(
      'function foo() { var a; if (bar) { return true; } else { let a; } }',
    ),
    invalid(
      'function foo() { if (bar) { var a; return true; } else { let a; } }',
    ),
    invalid(
      'function foo() { if (bar) { return true; } else { let a; }  while (baz) { var a; } }',
    ),
    invalid('function foo(a) { if (bar) { return true; } else { let a; } }'),
    invalid(
      'function foo(a = 1) { if (bar) { return true; } else { let a; } }',
    ),
    invalid(
      'function foo(a, b = a) { if (bar) { return true; } else { let a; }  if (bar) { return true; } else { let b; }}',
      2,
    ),
    invalid(
      'function foo(...args) { if (bar) { return true; } else { let args; } }',
    ),
    invalid(
      'function foo() { try {} catch (a) { if (bar) { return true; } else { let a; } } }',
    ),
    invalid(
      'function foo() { try {} catch (a) { if (bar) { if (baz) { return true; } else { let a; } } } }',
    ),
    invalid(
      'function foo() { try {} catch ({bar, a = 1}) { if (baz) { return true; } else { let a; } } }',
    ),
    invalid(
      'function foo() { if (bar) { return true; } else { let arguments; } }',
    ),
    invalid(
      'function foo() { if (bar) { return true; } else { let arguments; } return arguments[0]; }',
    ),
    invalid(
      'function foo() { if (bar) { return true; } else { let arguments; } if (baz) { return arguments[0]; } }',
    ),
    invalid(
      'function foo() { if (bar) { if (baz) { return true; } else { let arguments; } } }',
    ),
    invalid('function foo() { if (bar) { return true; } else { let a; } a; }'),
    invalid(
      'function foo() { if (bar) { return true; } else { let a; } if (baz) { a; } }',
    ),
    invalid(
      'function foo() { if (bar) { if (baz) { return true; } else { let a; } } a; }',
    ),
    invalid(
      'function foo() { if (bar) { if (baz) { return true; } else { let a; } a; } }',
    ),
    invalid(
      'function foo() { if (bar) { if (baz) { return true; } else { let a; } if (quux) { a; } } }',
    ),
    invalid('function a() { if (foo) { return true; } else { let a; } a(); }'),
    invalid('function a() { if (a) { return true; } else { let a; } }'),
    invalid('function a() { if (foo) { return a; } else { let a; } }'),
    invalid(
      'function foo() { if (bar) { return true; } else { let a; } function baz() { a; } }',
    ),
    invalid(
      'function foo() { if (bar) { if (baz) { return true; } else { let a; } (() => a) } }',
    ),
    invalid(
      'function foo() { if (bar) { return true; } else { let a; } var a; }',
    ),
    invalid(
      'function foo() { if (bar) { if (baz) { return true; } else { let a; } var a; } }',
    ),
    invalid(
      'function foo() { if (bar) { if (baz) { return true; } else { let a; } var { a } = {}; } }',
    ),
    invalid(
      'function foo() { if (bar) { if (baz) { return true; } else { let a; } if (quux) { var a; } } }',
    ),
    invalid(
      'function foo() { if (bar) { if (baz) { return true; } else { let a; } } if (quux) { var a; } }',
    ),
    invalid(
      'function foo() { if (quux) { var a; } if (bar) { if (baz) { return true; } else { let a; } } }',
    ),
    invalid(
      'function foo() { if (bar) { return true; } else { let a; } function a(){} }',
    ),
    invalid(
      'function foo() { if (baz) { if (bar) { return true; } else { let a; } function a(){} } }',
    ),
    invalid(
      'function foo() { if (bar) { if (baz) { return true; } else { let a; } } if (quux) { function a(){}  } }',
    ),
    invalid(
      'function foo() { if (bar) { if (baz) { return true; } else { let a; } } function a(){} }',
    ),
    invalid(
      'function foo() { let a; if (bar) { return true; } else { function a(){} } }',
    ),
    invalid(
      'function foo() { var a; if (bar) { return true; } else { function a(){} } }',
    ),
    invalid(
      'function foo() { if (bar) { return true; } else function baz() {} };',
    ),
    invalid('if (foo) { return true; } else { let a; }'),
    invalid('let a; if (foo) { return true; } else { let a; }'),
  ],
});

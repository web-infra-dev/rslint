/**
 * Tests for the `no-extra-parens` rule.
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/no-extra-parens/no-extra-parens._js_.test.ts
 *   packages/eslint-plugin/rules/no-extra-parens/no-extra-parens._ts_.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, lang, parserOptions, valid, invalid })` ->
 *    `ruleTester.run('no-extra-parens', null as never, { valid, invalid })`.
 *  - The local `invalid(code, output, line?, config?)` helper is expanded to its
 *    final `{ code, output, options, errors: [{ messageId: 'unexpected', line? }] }`.
 *  - The `$` unindent template tag is evaluated to its real multi-line string;
 *    `[...].join('\n')` array-of-lines and plain backtick literals are evaluated
 *    to their final string; `Array.from({ length: N }, ...)` and `[...].map(...)`
 *    error/spread helpers are expanded to their final element arrays.
 *  - `parserOptions` (sourceType / ecmaVersion / ecmaFeatures.jsx) dropped — rslint
 *    resolves via tsconfig; the RuleTester routes a `.tsx` fixture when real JSX is
 *    present, else `.ts`.
 *  - The eslint-vitest-rule-tester `recursive` field (fix-pass count) is dropped:
 *    rslint always fixes to a stable point (multi-pass). Every `output` kept here
 *    already equals that fully-recursive result.
 *  - The rule has a single messageId `unexpected` (message "Unnecessary parentheses
 *    around expression.", no `{{data}}`); upstream errors only ever pin `messageId`
 *    plus (sometimes) `line`/`column`, all asserted.
 *  - No `suggestions` exist in the upstream cases.
 *  - Two invalid cases pin ONLY `code`+`output` (no `errors`): ported output-only
 *    (the tester asserts the fix plus >=1 diagnostic; positions are not invented).
 *
 * KNOWN GAPS (cases that surface a real rslint<->upstream gap) are NOT deleted or
 * altered: each is moved into the `KNOWN_GAPS` block at the bottom, annotated with
 * what upstream expects vs. what rslint produces.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-extra-parens', null as never, {
  valid: [
    // ==== from no-extra-parens._js_.test.ts ====
    "foo",
    "a = b, c = d",
    "a = b ? c : d",
    "a = (b, c)",
    "a || b ? c = d : e = f",
    "(a = b) ? (c, d) : (e, f)",
    "a && b || c && d",
    "(a ? b : c) || (d ? e : f)",
    "a | b && c | d",
    "(a || b) && (c || d)",
    "a ^ b | c ^ d",
    "(a && b) | (c && d)",
    "a & b ^ c & d",
    "(a | b) ^ (c | d)",
    "a == b & c != d",
    "(a ^ b) & (c ^ d)",
    "a < b === c in d",
    "(a & b) !== (c & d)",
    "a << b >= c >>> d",
    "(a == b) instanceof (c != d)",
    "a + b << c - d",
    "(a <= b) >> (c > d)",
    "a * b + c / d",
    "(a << b) - (c >> d)",
    "+a % !b",
    "(a + b) * (c - d)",
    "-void+delete~typeof!a",
    "!(a * b); typeof (a / b); +(a % b); delete (a * b); ~(a / b); void (a % b); -(a * b)",
    "a(b = c, (d, e))",
    "(++a)(b); (c++)(d);",
    "new (A())",
    "new (foo.Baz().foo)",
    "new (foo.baz.bar().foo.baz)",
    "new ({}.baz.bar.foo().baz)",
    "new (doSomething().baz.bar().foo)",
    "new ([][0].baz.foo().bar.foo)",
    "new (foo\n.baz\n.bar()\n.foo.baz)",
    "new A()()",
    "(new A)()",
    "(new (Foo || Bar))()",
    "(new new foo())()",
    "new (new A)()",
    "new (new a.b)()",
    "new (new new foo())(bar)",
    "(new foo).bar",
    "(new foo)[bar]",
    "(new foo).bar.baz",
    "(new foo.bar).baz",
    "(new foo).bar()",
    "(new foo.bar).baz()",
    "new (new foo).bar",
    "new (new foo.bar).baz",
    "(new new foo()).baz",
    "(2 + 3) ** 4",
    "2 ** (2 + 3)",
    "new (import(source))",
    "import((s,t))",
    "a, b, c",
    "a = b = c",
    "a ? b ? c : d : e",
    "a ? b : c ? d : e",
    "a || b || c",
    "a || (b || c)",
    "a && b && c",
    "a && (b && c)",
    "a | b | c",
    "a | (b | c)",
    "a ^ b ^ c",
    "a ^ (b ^ c)",
    "a & b & c",
    "a & (b & c)",
    "a == b == c",
    "a == (b == c)",
    "a < b < c",
    "a < (b < c)",
    "a << b << c",
    "a << (b << c)",
    "a + b + c",
    "a + (b + c)",
    "a * b * c",
    "a * (b * c)",
    "!!a; typeof +b; void -c; ~delete d;",
    "a(b)",
    "a(b)(c)",
    "a((b, c))",
    "new new A",
    "2 ** 3 ** 4",
    "(2 ** 3) ** 4",
    "if(a);",
    "with(a){}",
    "switch(a){ case 0: break; }",
    "function a(){ return b; }",
    "var a = () => { return b; }",
    "throw a;",
    "while(a);",
    "do; while(a);",
    "for(;;);",
    "for(a in b);",
    "for(a in b, c);",
    "for(a of b);",
    "for (a of (b, c));",
    "var a = (b, c);",
    "[]",
    "[a, b]",
    "!{a}",
    "!{a: 0, b: 1}",
    "!{[a]:0}",
    "!{[(a, b)]:0}",
    "!{a, ...b}",
    "const {a} = {}",
    "const {a:b} = {}",
    "const {a:b=1} = {}",
    "const {[a]:b} = {}",
    "const {[a]:b=1} = {}",
    "const {[(a, b)]:c} = {}",
    "const {a, ...b} = {}",
    "class foo {}",
    "class foo { constructor(){} a(){} get b(){} set b(bar){} get c(){} set d(baz){} static e(){} }",
    "class foo { [a](){} get [b](){} set [b](bar){} get [c](){} set [d](baz){} static [e](){} }",
    "class foo { [(a,b)](){} }",
    "class foo { a(){} [b](){} c(){} [(d,e)](){} }",
    "class foo { [(a,b)](){} c(){} [d](){} e(){} }",
    "const foo = class { constructor(){} a(){} get b(){} set b(bar){} get c(){} set d(baz){} static e(){} }",
    "class foo { x; }",
    "class foo { static x; }",
    "class foo { x = 1; }",
    "class foo { static x = 1; }",
    "class foo { #x; }",
    "class foo { static #x; }",
    "class foo { static #x = 1; }",
    "class foo { #x(){} get #y() {} set #y(value) {} static #z(){} static get #q() {} static set #q(value) {} }",
    "const foo  = class { #x(){} get #y() {} set #y(value) {} static #z(){} static get #q() {} static set #q(value) {} }",
    "class foo { [(x, y)]; }",
    "class foo { static [(x, y)]; }",
    "class foo { [(x, y)] = 1; }",
    "class foo { static [(x, y)] = 1; }",
    "class foo { x = (y, z); }",
    "class foo { static x = (y, z); }",
    "class foo { #x = (y, z); }",
    "class foo { static #x = (y, z); }",
    "class foo { [(1, 2)] = (3, 4) }",
    "const foo = class { [(1, 2)] = (3, 4) }",
    "({});",
    "(function(){});",
    "(function*(){});",
    "(class{});",
    "(0).a",
    "(123).a",
    "(5_000).a",
    "(5_000_00).a",
    "(function(){ }())",
    "({a: function(){}}.a());",
    "({a:0}.a ? b : c)",
    "var isA = (/^a$/).test('a');",
    "var regex = (/^a$/);",
    "function a(){ return (/^a$/); }",
    "function a(){ return (/^a$/).test('a'); }",
    "var isA = ((/^a$/)).test('a');",
    "var foo = (function() { return bar(); }())",
    "var o = { foo: (function() { return bar(); }()) };",
    "o.foo = (function(){ return bar(); }());",
    "(function(){ return bar(); }()), (function(){ return bar(); }())",
    "var foo = (function() { return bar(); })()",
    "var o = { foo: (function() { return bar(); })() };",
    "o.foo = (function(){ return bar(); })();",
    "(function(){ return bar(); })(), (function(){ return bar(); })()",
    "function foo() { return (function(){}()); }",
    "var foo = (function*() { if ((yield foo()) + 1) { return; } }())",
    "(() => 0)()",
    "(_ => 0)()",
    "_ => 0, _ => 1",
    "a = () => b = 0",
    "0 ? _ => 0 : _ => 0",
    "(_ => 0) || (_ => 0)",
    "x => ({foo: 1})",
    "1 + 2 ** 3",
    "1 - 2 ** 3",
    "2 ** -3",
    "(-2) ** 3",
    "(+2) ** 3",
    "+ (2 ** 3)",
    "a => ({b: c}[d])",
    "a => ({b: c}.d())",
    "a => ({b: c}.d.e)",
    {
      code: "(0)",
      options: ["functions"],
    },
    {
      code: "((0))",
      options: ["functions"],
    },
    {
      code: "a + (b * c)",
      options: ["functions"],
    },
    {
      code: "a + ((b * c))",
      options: ["functions"],
    },
    {
      code: "(a)(b)",
      options: ["functions"],
    },
    {
      code: "((a))(b)",
      options: ["functions"],
    },
    {
      code: "a, (b = c)",
      options: ["functions"],
    },
    {
      code: "a, ((b = c))",
      options: ["functions"],
    },
    {
      code: "for(a in (0));",
      options: ["functions"],
    },
    {
      code: "for(a in ((0)));",
      options: ["functions"],
    },
    {
      code: "var a = (b = c)",
      options: ["functions"],
    },
    {
      code: "var a = ((b = c))",
      options: ["functions"],
    },
    {
      code: "_ => (a = 0)",
      options: ["functions"],
    },
    {
      code: "_ => ((a = 0))",
      options: ["functions"],
    },
    {
      code: "while ((foo = bar())) {}",
      options: ["all",{"conditionalAssign":false}],
    },
    {
      code: "if ((foo = bar())) {}",
      options: ["all",{"conditionalAssign":false}],
    },
    {
      code: "do; while ((foo = bar()))",
      options: ["all",{"conditionalAssign":false}],
    },
    {
      code: "for (;(a = b););",
      options: ["all",{"conditionalAssign":false}],
    },
    {
      code: "var a = ((b = c)) ? foo : bar;",
      options: ["all",{"conditionalAssign":false}],
    },
    {
      code: "while (((foo = bar()))) {}",
      options: ["all",{"conditionalAssign":false}],
    },
    {
      code: "var a = (((b = c))) ? foo : bar;",
      options: ["all",{"conditionalAssign":false}],
    },
    {
      code: "(a && b) ? foo : bar",
      options: ["all",{"ternaryOperandBinaryExpressions":false}],
    },
    {
      code: "(a - b > a) ? foo : bar",
      options: ["all",{"ternaryOperandBinaryExpressions":false}],
    },
    {
      code: "foo ? (bar || baz) : qux",
      options: ["all",{"ternaryOperandBinaryExpressions":false}],
    },
    {
      code: "foo ? bar : (baz || qux)",
      options: ["all",{"ternaryOperandBinaryExpressions":false}],
    },
    {
      code: "(a, b) ? (c, d) : (e, f)",
      options: ["all",{"ternaryOperandBinaryExpressions":false}],
    },
    {
      code: "(a = b) ? c : d",
      options: ["all",{"ternaryOperandBinaryExpressions":false}],
    },
    {
      code: "a + (b * c)",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "(a * b) + c",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "(a * b) / c",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "a || (b && c)",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "a + ((b * c))",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "((a * b)) + c",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "((a * b)) / c",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "a || ((b && c))",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "function a(b) { return b || c; }",
      options: ["all",{"returnAssign":false}],
    },
    {
      code: "function a(b) { return; }",
      options: ["all",{"returnAssign":false}],
    },
    {
      code: "function a(b) { return (b = 1); }",
      options: ["all",{"returnAssign":false}],
    },
    {
      code: "function a(b) { return (b = c) || (b = d); }",
      options: ["all",{"returnAssign":false}],
    },
    {
      code: "function a(b) { return c ? (d = b) : (e = b); }",
      options: ["all",{"returnAssign":false}],
    },
    {
      code: "b => b || c;",
      options: ["all",{"returnAssign":false}],
    },
    {
      code: "b => (b = 1);",
      options: ["all",{"returnAssign":false}],
    },
    {
      code: "b => (b = c) || (b = d);",
      options: ["all",{"returnAssign":false}],
    },
    {
      code: "b => c ? (d = b) : (e = b);",
      options: ["all",{"returnAssign":false}],
    },
    {
      code: "b => { return b || c };",
      options: ["all",{"returnAssign":false}],
    },
    {
      code: "b => { return (b = 1) };",
      options: ["all",{"returnAssign":false}],
    },
    {
      code: "b => { return (b = c) || (b = d) };",
      options: ["all",{"returnAssign":false}],
    },
    {
      code: "b => { return c ? (d = b) : (e = b) };",
      options: ["all",{"returnAssign":false}],
    },
    {
      code: "function a(b) { return ((b = 1)); }",
      options: ["all",{"returnAssign":false}],
    },
    {
      code: "b => ((b = 1));",
      options: ["all",{"returnAssign":false}],
    },
    "(function(){}).foo(), 1, 2;",
    "(function(){}).foo++;",
    "(function(){}).foo() || bar;",
    "(function(){}).foo() + 1;",
    "(function(){}).foo() ? bar : baz;",
    "(function(){}).foo.bar();",
    "(function(){}.foo());",
    "(function(){}.foo.bar);",
    "(class{}).foo(), 1, 2;",
    "(class{}).foo++;",
    "(class{}).foo() || bar;",
    "(class{}).foo() + 1;",
    "(class{}).foo() ? bar : baz;",
    "(class{}).foo.bar();",
    "(class{}.foo());",
    "(class{}.foo.bar);",
    "function *a() { yield b; }",
    "function *a() { yield yield; }",
    "function *a() { yield b, c; }",
    "function *a() { yield (b, c); }",
    "function *a() { yield b + c; }",
    "function *a() { (yield b) + c; }",
    "function a() {\n  return (\n    a % b == 0\n  )\n}",
    "function a() {\n    return (\n        b\n    );\n}",
    "function a() {\n    return (\n        <JSX />\n    );\n}",
    "function a() {\n    return (\n        <></>\n    );\n}",
    "throw (\n    a\n);",
    "function *a() {\n    yield (\n        b\n    );\n}",
    "(a\n)++",
    "(a\n)--",
    "(a\n\n)++",
    "(a.b\n)--",
    "(a\n.b\n)++",
    "(a[\nb\n]\n)--",
    "(a[b]\n\n)++",
    "async function a() { await (a + b) }",
    "async function a() { await (a + await b) }",
    "async function a() { (await a)() }",
    "async function a() { new (await a) }",
    "async function a() { await (a ** b) }",
    "async function a() { (await a) ** b }",
    {
      code: "(foo instanceof bar) instanceof baz",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "(foo in bar) in baz",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "(foo + bar) + baz",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "(foo && bar) && baz",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "foo instanceof (bar instanceof baz)",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "foo in (bar in baz)",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "foo + (bar + baz)",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "foo && (bar && baz)",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "((foo instanceof bar)) instanceof baz",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "((foo in bar)) in baz",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    "(async function() {});",
    "(async function () { }());",
    {
      code: "const Component = (<div />)",
      options: ["all",{"ignoreJSX":"all"}],
    },
    {
      code: "const Component = (\n    <div\n        prop={true}\n    />\n)",
      options: ["all",{"ignoreJSX":"all"}],
    },
    {
      code: "const Component = ((<div />))",
      options: ["all",{"ignoreJSX":"all"}],
    },
    {
      code: "const Component = (<>\n  <p />\n</>);",
      options: ["all",{"ignoreJSX":"all"}],
    },
    {
      code: "const Component = ((<>\n  <p />\n</>));",
      options: ["all",{"ignoreJSX":"all"}],
    },
    {
      code: "const Component = (<div>\n  <p />\n</div>);",
      options: ["all",{"ignoreJSX":"all"}],
    },
    {
      code: "const Component = (\n  <div />\n);",
      options: ["all",{"ignoreJSX":"all"}],
    },
    {
      code: "const Component =\n  (<div />)",
      options: ["all",{"ignoreJSX":"all"}],
    },
    {
      code: "const Component = (<div />);",
      options: ["all",{"ignoreJSX":"single-line"}],
    },
    {
      code: "const Component = ((<div />));",
      options: ["all",{"ignoreJSX":"single-line"}],
    },
    {
      code: "const Component = (<div><p /></div>)",
      options: ["all",{"ignoreJSX":"single-line"}],
    },
    {
      code: "const Component = (\n  <div />\n);",
      options: ["all",{"ignoreJSX":"single-line"}],
    },
    {
      code: "const Component =\n(<div />)",
      options: ["all",{"ignoreJSX":"single-line"}],
    },
    {
      code: "const Component = (\n<div>\n  <p />\n</div>\n);",
      options: ["all",{"ignoreJSX":"multi-line"}],
    },
    {
      code: "const Component = ((\n<div>\n  <p />\n</div>\n));",
      options: ["all",{"ignoreJSX":"multi-line"}],
    },
    {
      code: "const Component = (<div>\n  <p />\n</div>);",
      options: ["all",{"ignoreJSX":"multi-line"}],
    },
    {
      code: "const Component =\n(<div>\n  <p />\n</div>);",
      options: ["all",{"ignoreJSX":"multi-line"}],
    },
    {
      code: "const Component = (<div\n  prop={true}\n/>)",
      options: ["all",{"ignoreJSX":"multi-line"}],
    },
    {
      code: "var a = b => 1 ? 2 : 3",
      options: ["all",{"enforceForArrowConditionals":false}],
    },
    {
      code: "var a = b => (1 ? 2 : 3)",
      options: ["all",{"enforceForArrowConditionals":false}],
    },
    {
      code: "var a = (b) => (1 ? 2 : 3)",
      options: ["all",{"enforceForArrowConditionals":false}],
    },
    {
      code: "var a = (b) => ((1 ? 2 : 3))",
      options: ["all",{"enforceForArrowConditionals":false}],
    },
    {
      code: "var a = b => (1 ? 2 : 3)",
      options: ["all",{"ignoredNodes":["ArrowFunctionExpression[body.type=ConditionalExpression]"]}],
    },
    {
      code: "var a = (b) => (1 ? 2 : 3)",
      options: ["all",{"ignoredNodes":["ArrowFunctionExpression[body.type=ConditionalExpression]"]}],
    },
    {
      code: "var a = (b) => ((1 ? 2 : 3))",
      options: ["all",{"ignoredNodes":["ArrowFunctionExpression[body.type=ConditionalExpression]"]}],
    },
    {
      code: "(a, b)",
      options: ["all",{"enforceForSequenceExpressions":false}],
    },
    {
      code: "((a, b))",
      options: ["all",{"enforceForSequenceExpressions":false}],
    },
    {
      code: "(foo(), bar());",
      options: ["all",{"enforceForSequenceExpressions":false}],
    },
    {
      code: "((foo(), bar()));",
      options: ["all",{"enforceForSequenceExpressions":false}],
    },
    {
      code: "if((a, b)){}",
      options: ["all",{"enforceForSequenceExpressions":false}],
    },
    {
      code: "if(((a, b))){}",
      options: ["all",{"enforceForSequenceExpressions":false}],
    },
    {
      code: "while ((val = foo(), val < 10));",
      options: ["all",{"enforceForSequenceExpressions":false}],
    },
    {
      code: "(new foo()).bar",
      options: ["all",{"enforceForNewInMemberExpressions":false}],
    },
    {
      code: "(new foo())[bar]",
      options: ["all",{"enforceForNewInMemberExpressions":false}],
    },
    {
      code: "(new foo()).bar()",
      options: ["all",{"enforceForNewInMemberExpressions":false}],
    },
    {
      code: "(new foo(bar)).baz",
      options: ["all",{"enforceForNewInMemberExpressions":false}],
    },
    {
      code: "(new foo.bar()).baz",
      options: ["all",{"enforceForNewInMemberExpressions":false}],
    },
    {
      code: "(new foo.bar()).baz()",
      options: ["all",{"enforceForNewInMemberExpressions":false}],
    },
    {
      code: "((new foo.bar())).baz()",
      options: ["all",{"enforceForNewInMemberExpressions":false}],
    },
    {
      code: "(new foo()).bar",
      options: ["all",{"ignoredNodes":["MemberExpression[object.type=NewExpression]"]}],
    },
    {
      code: "(new foo())[bar]",
      options: ["all",{"ignoredNodes":["MemberExpression[object.type=NewExpression]"]}],
    },
    {
      code: "(new foo()).bar()",
      options: ["all",{"ignoredNodes":["MemberExpression[object.type=NewExpression]"]}],
    },
    {
      code: "(new foo(bar)).baz",
      options: ["all",{"ignoredNodes":["MemberExpression[object.type=NewExpression]"]}],
    },
    {
      code: "(new foo.bar()).baz",
      options: ["all",{"ignoredNodes":["MemberExpression[object.type=NewExpression]"]}],
    },
    {
      code: "(new foo.bar()).baz()",
      options: ["all",{"ignoredNodes":["MemberExpression[object.type=NewExpression]"]}],
    },
    {
      code: "((new foo.bar())).baz()",
      options: ["all",{"ignoredNodes":["MemberExpression[object.type=NewExpression]"]}],
    },
    {
      code: "var foo = (function(){}).call()",
      options: ["all",{"enforceForFunctionPrototypeMethods":false}],
    },
    {
      code: "var foo = (function(){}).apply()",
      options: ["all",{"enforceForFunctionPrototypeMethods":false}],
    },
    {
      code: "var foo = (function(){}.call())",
      options: ["all",{"enforceForFunctionPrototypeMethods":false}],
    },
    {
      code: "var foo = (function(){}.apply())",
      options: ["all",{"enforceForFunctionPrototypeMethods":false}],
    },
    {
      code: "var foo = (function(){}).call(arg)",
      options: ["all",{"enforceForFunctionPrototypeMethods":false}],
    },
    {
      code: "var foo = (function(){}.apply(arg))",
      options: ["all",{"enforceForFunctionPrototypeMethods":false}],
    },
    {
      code: "var foo = (function(){}['call']())",
      options: ["all",{"enforceForFunctionPrototypeMethods":false}],
    },
    {
      code: "var foo = (function(){})[`apply`]()",
      options: ["all",{"enforceForFunctionPrototypeMethods":false}],
    },
    {
      code: "var foo = ((function(){})).call()",
      options: ["all",{"enforceForFunctionPrototypeMethods":false}],
    },
    {
      code: "var foo = ((function(){}).apply())",
      options: ["all",{"enforceForFunctionPrototypeMethods":false}],
    },
    {
      code: "var foo = ((function(){}.call()))",
      options: ["all",{"enforceForFunctionPrototypeMethods":false}],
    },
    {
      code: "var foo = ((((function(){})).apply()))",
      options: ["all",{"enforceForFunctionPrototypeMethods":false}],
    },
    {
      code: "foo((function(){}).call().bar)",
      options: ["all",{"enforceForFunctionPrototypeMethods":false}],
    },
    {
      code: "foo = (function(){}).call()()",
      options: ["all",{"enforceForFunctionPrototypeMethods":false}],
    },
    {
      code: "foo = (function(){}.call())()",
      options: ["all",{"enforceForFunctionPrototypeMethods":false}],
    },
    {
      code: "var foo = { bar: (function(){}.call()) }",
      options: ["all",{"enforceForFunctionPrototypeMethods":false}],
    },
    {
      code: "var foo = { [(function(){}.call())]: bar  }",
      options: ["all",{"enforceForFunctionPrototypeMethods":false}],
    },
    {
      code: "if((function(){}).call()){}",
      options: ["all",{"enforceForFunctionPrototypeMethods":false}],
    },
    {
      code: "while((function(){}.apply())){}",
      options: ["all",{"enforceForFunctionPrototypeMethods":false}],
    },
    "let a = [ ...b ]",
    "let a = { ...b }",
    { code: "let a = { ...b }" },
    "let a = [ ...(b, c) ]",
    "let a = { ...(b, c) }",
    { code: "let a = { ...(b, c) }" },
    "var [x = (1, foo)] = bar",
    "class A extends B {}",
    "const A = class extends B {}",
    "class A extends (B=C) {}",
    "const A = class extends (B=C) {}",
    "class A extends (++foo) {}",
    "() => ({ foo: 1 })",
    "() => ({ foo: 1 }).foo",
    "() => ({ foo: 1 }.foo().bar).baz.qux()",
    "() => ({ foo: 1 }.foo().bar + baz)",
    { code: "export default (a, b)" },
    { code: "export default (function(){}).foo" },
    { code: "export default (class{}).foo" },
    "({}).hasOwnProperty.call(foo, bar)",
    "({}) ? foo() : bar()",
    "({}) + foo",
    "(function(){}) + foo",
    {
      code: "((function(){}).foo.bar)();",
      options: ["functions"],
    },
    {
      code: "((function(){}).foo)();",
      options: ["functions"],
    },
    "for ((let[a]);;);",
    "for ((let)[a];;);",
    "for ((let[a] = 1);;);",
    "for ((let[a]) = 1;;);",
    "for ((let)[a] = 1;;);",
    "for ((let[a, b] = foo);;);",
    "for ((let[a].b = 1);;);",
    "for ((let[a].b) = 1;;);",
    "for ((let[a]).b = 1;;);",
    "for ((let)[a].b = 1;;);",
    "for ((let[a])();;);",
    "for ((let)[a]();;);",
    "for ((let[a]) + b;;);",
    "for ((let[foo]) in bar);",
    "for ((let)[foo] in bar);",
    "for ((let[foo].bar) in baz);",
    "for ((let[foo]).bar in baz);",
    "for ((let)[foo].bar in baz);",
    "for ((let) of foo);",
    "for ((let).foo of bar);",
    "for ((let.foo) of bar);",
    "for ((let[foo]) of bar);",
    "for ((let)[foo] of bar);",
    "for ((let.foo.bar) of baz);",
    "for ((let.foo).bar of baz);",
    "for ((let).foo.bar of baz);",
    "for ((let[foo].bar) of baz);",
    "for ((let[foo]).bar of baz);",
    "for ((let)[foo].bar of baz);",
    "for ((let)().foo of bar);",
    "for ((let()).foo of bar);",
    "for ((let().foo) of bar);",
    "for (let a = (b in c); ;);",
    "for (let a = (b && c in d); ;);",
    "for (let a = (b in c && d); ;);",
    "for (let a = (b => b in c); ;);",
    "for (let a = b => (b in c); ;);",
    "for (let a = (b in c in d); ;);",
    "for (let a = (b in c), d = (e in f); ;);",
    "for (let a = (b => c => b in c); ;);",
    "for (let a = (b && c && d in e); ;);",
    "for (let a = b && (c in d); ;);",
    "for (let a = (b in c) && (d in e); ;);",
    "for ((a in b); ;);",
    "for (a = (b in c); ;);",
    "for ((a in b && c in d && e in f); ;);",
    "for (let a = [] && (b in c); ;);",
    "for (let a = (b in [c]); ;);",
    "for (let a = b => (c in d); ;);",
    "for (let a = (b in c) ? d : e; ;);",
    "for (let a = (b in c ? d : e); ;);",
    "for (let a = b ? c : (d in e); ;);",
    "for (let a = (b in c), d = () => { for ((e in f);;); for ((g in h);;); }; ;); for((i in j); ;);",
    "for (let a = b; a; a); a; a;",
    "for (a; a; a); a; a;",
    "for (; a; a); a; a;",
    "for (let a = (b && c) === d; ;);",
    "new (a()).b.c;",
    "new (a().b).c;",
    "new (a().b.c);",
    "new (a().b().d);",
    "new a().b().d;",
    "new (a(b()).c)",
    "new (a.b()).c",
    { code: "var v = (a ?? b) || c" },
    { code: "var v = a ?? (b || c)" },
    { code: "var v = (a ?? b) && c" },
    { code: "var v = a ?? (b && c)" },
    { code: "var v = (a || b) ?? c" },
    { code: "var v = a || (b ?? c)" },
    { code: "var v = (a && b) ?? c" },
    { code: "var v = a && (b ?? c)" },
    { code: "var v = (obj?.aaa).bbb" },
    { code: "var v = (obj?.aaa)()" },
    { code: "var v = new (obj?.aaa)()" },
    { code: "var v = new (obj?.aaa)" },
    { code: "var v = (obj?.aaa)`template`" },
    { code: "var v = (obj?.()).bbb" },
    { code: "var v = (obj?.())()" },
    { code: "var v = new (obj?.())()" },
    { code: "var v = new (obj?.())" },
    { code: "var v = (obj?.())`template`" },
    { code: "(obj?.aaa).bbb = 0" },
    { code: "var foo = (function(){})?.()" },
    { code: "var foo = (function(){}?.())" },
    {
      code: "var foo = (function(){})?.call()",
      options: ["all",{"enforceForFunctionPrototypeMethods":false}],
    },
    {
      code: "var foo = (function(){}?.call())",
      options: ["all",{"enforceForFunctionPrototypeMethods":false}],
    },
    {
      code: "(0).toString();",
      options: ["functions"],
    },
    {
      code: "(Object.prototype.toString.call())",
      options: ["functions"],
    },
    {
      code: "({}.toString.call());",
      options: ["functions"],
    },
    {
      code: "(function(){} ? a() : b());",
      options: ["functions"],
    },
    {
      code: "(/^a$/).test(x);",
      options: ["functions"],
    },
    {
      code: "a = (b * c);",
      options: ["functions"],
    },
    {
      code: "(a * b) + c;",
      options: ["functions"],
    },
    {
      code: "typeof (a);",
      options: ["functions"],
    },
    "const span = /**@type {HTMLSpanElement}*/(event.currentTarget);",
    "if (/** @type {Compiler | MultiCompiler} */(options).hooks) console.log('good');",
    "validate(/** @type {Schema} */ (schema), options, {\n    name: \"Dev Server\",\n    baseDataPath: \"options\",\n});",
    "if (condition) {\n    /** @type {ServerOptions} */\n    (options.server.options).requestCert = false;\n}",
    "const x = /** @type {number} */ (value)",
    "if (/** @type {A} */ (/** @type {B} */ (/** @type {C} */ (/** @type {D} */ (expr))))) {}",
    {
      code: "const net = ipaddr.parseCIDR(/* any-string */ (cidr));",
      options: ["all",{"allowParensAfterCommentPattern":"any-string"}],
    },
    "(a ? b : c) ? d : e",
    {
      code: "a ? (b ? c : d) : e",
      options: ["all",{"nestedConditionalExpressions":false}],
    },
    {
      code: "a ? b : (c ? d : e)",
      options: ["all",{"nestedConditionalExpressions":false}],
    },
    "(a) = function () {};",
    "(a) = () => {};",
    "(a) = class {};",
    "(a) ??= function () {};",
    "(a) &&= class extends SuperClass {};",
    "(a) ||= async () => {}",
    {
      code: "((a)) = function () {};",
      options: ["functions"],
    },
    {
      code: "const x = [\n  ...a ? [1, 2, 3] : [],\n  ...(a ? [1, 2, 3] : []),\n]",
      options: ["all",{"allowNodesInSpreadElement":{"ConditionalExpression":true}}],
    },
    {
      code: "const x = [\n  ...a ? [1, 2, 3] : [],\n  ...(a ? [1, 2, 3] : []),\n]",
      options: ["all",{"ignoredNodes":["SpreadElement[argument.type=ConditionalExpression]"]}],
    },
    {
      code: "const x = [\n  ...b ?? c,\n  ...(b ?? c),\n]",
      options: ["all",{"allowNodesInSpreadElement":{"LogicalExpression":true}}],
    },
    {
      code: "const x = [\n  ...b ?? c,\n  ...(b ?? c),\n]",
      options: ["all",{"ignoredNodes":["SpreadElement[argument.type=LogicalExpression]"]}],
    },
    {
      code: "const fruits = {\n  ...isSummer && { watermelon: 30 },\n  ...(isSummer && { watermelon: 30 }),\n};",
      options: ["all",{"allowNodesInSpreadElement":{"LogicalExpression":true}}],
    },
    {
      code: "const fruits = {\n  ...isSummer && { watermelon: 30 },\n  ...(isSummer && { watermelon: 30 }),\n};",
      options: ["all",{"ignoredNodes":["SpreadElement[argument.type=LogicalExpression]"]}],
    },
    {
      code: "async function example() {\n  const promiseArray = Promise.resolve([1, 2, 3]);\n  console.log(...(await promiseArray));\n}",
      options: ["all",{"allowNodesInSpreadElement":{"AwaitExpression":true}}],
    },
    {
      code: "async function example() {\n  const promiseArray = Promise.resolve([1, 2, 3]);\n  console.log(...(await promiseArray));\n}",
      options: ["all",{"ignoredNodes":["SpreadElement[argument.type=AwaitExpression]"]}],
    },
    {
      code: "const x = [\n  ...a ? [1, 2, 3] : [],\n  ...(a ? [1, 2, 3] : []),\n]\n\nconst fruits = {\n  ...isSummer && { watermelon: 30 },\n  ...(isSummer && { watermelon: 30 }),\n};",
      options: ["all",{"allowNodesInSpreadElement":{"LogicalExpression":true,"ConditionalExpression":true}}],
    },
    {
      code: "const x = [\n  ...a ? [1, 2, 3] : [],\n  ...(a ? [1, 2, 3] : []),\n]\n\nconst fruits = {\n  ...isSummer && { watermelon: 30 },\n  ...(isSummer && { watermelon: 30 }),\n};",
      options: ["all",{"ignoredNodes":["SpreadElement"]}],
    },
    {
      code: "const conditionStatement = (\n  condition1 &&\n  condition2 &&\n  condition3\n);",
      options: ["all",{"ignoredNodes":["VariableDeclarator[init.type=\"LogicalExpression\"]"]}],
    },
    {
      code: "const joinedText = (\n  dataFromQuery\n    .filter((item) => item.isActive)\n    .map((item) => item.name)\n    .join(\"\")\n);",
      options: ["all",{"ignoredNodes":["VariableDeclarator[init]"]}],
    },

    // ==== from no-extra-parens._ts_.test.ts ====
    "async function f(arg: any) { await (arg as Promise<void>); }",
    "async function f(arg: any) { await (arg satisfies Promise<void>); }",
    "async function f(arg: any) { await (arg!); }",
    "async function f(arg: Promise<any>) { await arg; }",
    "(0).toString();",
    "(function(){}) ? a() : b();",
    "a<import('')>(1);",
    "new a<import('')>(1);",
    "a<A>(1);",
    { code: "(++(<A>a))(b); ((c as C)++)(d);" },
    { code: "const x = (1 as 1) | (1 as 1);" },
    { code: "const x = (<1>1) | (<1>1);" },
    { code: "const x = (1 as 1) | 2;" },
    { code: "const x = (1 as 1) + 2 + 2;" },
    { code: "const x = 1 + 1 + (2 as 2);" },
    { code: "const x = 1 | (2 as 2);" },
    { code: "const x = (<1>1) | 2;" },
    { code: "const x = 1 | (<2>2);" },
    { code: "t.true((me.get as SinonStub).calledWithExactly('/foo', other));" },
    { code: "t.true((<SinonStub>me.get).calledWithExactly('/foo', other));" },
    { code: "(requestInit.headers as Headers).get('Cookie');" },
    { code: "(<Headers> requestInit.headers).get('Cookie');" },
    { code: "class Foo {}" },
    { code: "class Foo extends (Bar as any) {}" },
    { code: "const foo = class extends (Bar as any) {}" },
    {
      code: "[a as b];",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "() => (1 as 1);",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "x = a as b;",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "const x = (1 as 1) | 2;",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "const x = 1 | (2 as 2);",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "const x = await (foo as Promise<void>);",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "const res2 = (fn as foo)();",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "(x as boolean) ? 1 : 0;",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "x ? (1 as 1) : 2;",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "x ? 1 : (2 as 2);",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "while (foo as boolean) {};",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "do {} while (foo as boolean);",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "for (let i of ([] as Foo)) {}",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "for (let i in ({} as Foo)) {}",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "for ((1 as 1);;) {}",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "for (;(1 as 1);) {}",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "for (;;(1 as 1)) {}",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "if (1 as 1) {}",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "const x = (1 as 1).toString();",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "new (1 as 1)();",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "const x = { ...(1 as 1), ...{} };",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "throw (1 as 1);",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "throw 1;",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "const x = !(1 as 1);",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "function *x() { yield (1 as 1); yield 1; }",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "switch (foo) { case 1: case (2 as 2): break; default: break; }",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "[<b>a];",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "() => (<1>1);",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "x = <b>a;",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "const x = (<1>1) | 2;",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "const x = 1 | (<2>2);",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "const x = await (<Promise<void>>foo);",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "const res2 = (<foo>fn)();",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "(<boolean>x) ? 1 : 0;",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "x ? (<1>1) : 2;",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "x ? 1 : (<2>2);",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "while (<boolean>foo) {};",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "do {} while (<boolean>foo);",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "for (let i of (<Foo>[])) {}",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "for (let i in (<Foo>{})) {}",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "for ((<1>1);;) {}",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "for (;(<1>1);) {}",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "for (;;(<1>1)) {}",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "if (<1>1) {}",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "const x = (<1>1).toString();",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "new (<1>1)();",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "const x = { ...(<1>1), ...{} };",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "throw (<1>1);",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "throw 1;",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "const x = !(<1>1);",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "function *x() { yield (<1>1); yield 1; }",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "switch (foo) { case 1: case (<2>2): break; default: break; }",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    { code: "declare const f: <T>(x: T) => any" },
    { code: "f<(number | string)[]>(['a', 1])" },
    { code: "f<(number)>(1)" },
    { code: "f<(number) | string>(1)" },
    { code: "const x = (1 satisfies number).toFixed();" },
    "type Foo = string & (number | 'bar')",
    "type Foo = (a extends string ? 'bar' : number)[]",
    {
      code: "type Foo = boolean | (Bar & Baz)",
      options: ["all",{"nestedBinaryExpressions":false}],
    },
    {
      code: "type TBar = (\n    First\n    & Second\n    & Third\n);",
      options: ["all",{"ignoredNodes":["TSTypeAliasDeclaration[typeAnnotation.type=\"TSIntersectionType\"]"]}],
    },
  ],

  invalid: [
    // ==== from no-extra-parens._js_.test.ts ====
    {
      code: "(0)",
      output: "0",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(  0  )",
      output: "  0  ",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "if((0));",
      output: "if(0);",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "if(( 0 ));",
      output: "if( 0 );",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "with((0)){}",
      output: "with(0){}",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "switch((0)){}",
      output: "switch(0){}",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "switch(0){ case (1): break; }",
      output: "switch(0){ case 1: break; }",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "for((0);;);",
      output: "for(0;;);",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "for(;(0););",
      output: "for(;0;);",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "for(;;(0));",
      output: "for(;;0);",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "throw(0)",
      output: "throw 0",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "while((0));",
      output: "while(0);",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "do; while((0))",
      output: "do; while(0)",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "for(a in (0));",
      output: "for(a in 0);",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "for(a of (0));",
      output: "for(a of 0);",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "const foo = {[(a)]:1}",
      output: "const foo = {[a]:1}",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "const foo = {[(a=b)]:1}",
      output: "const foo = {[a=b]:1}",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "const foo = {*[(Symbol.iterator)]() {}}",
      output: "const foo = {*[Symbol.iterator]() {}}",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "const foo = { get [(a)]() {}}",
      output: "const foo = { get [a]() {}}",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "const foo = {[(a+b)]:c, d}",
      output: "const foo = {[a+b]:c, d}",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "const foo = {a, [(b+c)]:d, e}",
      output: "const foo = {a, [b+c]:d, e}",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "const foo = {[(a+b)]:c, d:e}",
      output: "const foo = {[a+b]:c, d:e}",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "const foo = {a:b, [(c+d)]:e, f:g}",
      output: "const foo = {a:b, [c+d]:e, f:g}",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "const foo = {[(a+b)]:c, [d]:e}",
      output: "const foo = {[a+b]:c, [d]:e}",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "const foo = {[a]:b, [(c+d)]:e, [f]:g}",
      output: "const foo = {[a]:b, [c+d]:e, [f]:g}",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "const foo = {[(a+b)]:c, [(d,e)]:f}",
      output: "const foo = {[a+b]:c, [(d,e)]:f}",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "const foo = {[(a,b)]:c, [(d+e)]:f, [(g,h)]:e}",
      output: "const foo = {[(a,b)]:c, [d+e]:f, [(g,h)]:e}",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "const foo = {a, b:c, [(d+e)]:f, [(g,h)]:i, [j]:k}",
      output: "const foo = {a, b:c, [d+e]:f, [(g,h)]:i, [j]:k}",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "const foo = {[a+(b*c)]:d}",
      output: "const foo = {[a+b*c]:d}",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "const foo = {[(a, (b+c))]:d}",
      output: "const foo = {[(a, b+c)]:d}",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "const {[(a)]:b} = {}",
      output: "const {[a]:b} = {}",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "const {[(a=b)]:c=1} = {}",
      output: "const {[a=b]:c=1} = {}",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "const {[(a+b)]:c, d} = {}",
      output: "const {[a+b]:c, d} = {}",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "const {a, [(b+c)]:d, e} = {}",
      output: "const {a, [b+c]:d, e} = {}",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "const {[(a+b)]:c, d:e} = {}",
      output: "const {[a+b]:c, d:e} = {}",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "const {a:b, [(c+d)]:e, f:g} = {}",
      output: "const {a:b, [c+d]:e, f:g} = {}",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "const {[(a+b)]:c, [d]:e} = {}",
      output: "const {[a+b]:c, [d]:e} = {}",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "const {[a]:b, [(c+d)]:e, [f]:g} = {}",
      output: "const {[a]:b, [c+d]:e, [f]:g} = {}",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "const {[(a+b)]:c, [(d,e)]:f} = {}",
      output: "const {[a+b]:c, [(d,e)]:f} = {}",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "const {[(a,b)]:c, [(d+e)]:f, [(g,h)]:e} = {}",
      output: "const {[(a,b)]:c, [d+e]:f, [(g,h)]:e} = {}",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "const {a, b:c, [(d+e)]:f, [(g,h)]:i, [j]:k} = {}",
      output: "const {a, b:c, [d+e]:f, [(g,h)]:i, [j]:k} = {}",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "const {[a+(b*c)]:d} = {}",
      output: "const {[a+b*c]:d} = {}",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "const {[(a, (b+c))]:d} = {}",
      output: "const {[(a, b+c)]:d} = {}",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "class foo { [(a)](){} }",
      output: "class foo { [a](){} }",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "class foo {*[(Symbol.iterator)]() {}}",
      output: "class foo {*[Symbol.iterator]() {}}",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "class foo { get [(a)](){} }",
      output: "class foo { get [a](){} }",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "class foo { set [(a)](bar){} }",
      output: "class foo { set [a](bar){} }",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "class foo { static [(a)](bar){} }",
      output: "class foo { static [a](bar){} }",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "class foo { [(a=b)](){} }",
      output: "class foo { [a=b](){} }",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "class foo { constructor (){} [(a+b)](){} }",
      output: "class foo { constructor (){} [a+b](){} }",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "class foo { [(a+b)](){} constructor (){} }",
      output: "class foo { [a+b](){} constructor (){} }",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "class foo { [(a+b)](){} c(){} }",
      output: "class foo { [a+b](){} c(){} }",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "class foo { a(){} [(b+c)](){} d(){} }",
      output: "class foo { a(){} [b+c](){} d(){} }",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "class foo { [(a+b)](){} [c](){} }",
      output: "class foo { [a+b](){} [c](){} }",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "class foo { [a](){} [(b+c)](){} [d](){} }",
      output: "class foo { [a](){} [b+c](){} [d](){} }",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "class foo { [(a+b)](){} [(c,d)](){} }",
      output: "class foo { [a+b](){} [(c,d)](){} }",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "class foo { [(a,b)](){} [(c+d)](){} }",
      output: "class foo { [(a,b)](){} [c+d](){} }",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "class foo { [a+(b*c)](){} }",
      output: "class foo { [a+b*c](){} }",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "const foo = class { [(a)](){} }",
      output: "const foo = class { [a](){} }",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "class foo { [(x)]; }",
      output: "class foo { [x]; }",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "class foo { static [(x)]; }",
      output: "class foo { static [x]; }",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "class foo { [(x)] = 1; }",
      output: "class foo { [x] = 1; }",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "class foo { static [(x)] = 1; }",
      output: "class foo { static [x] = 1; }",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "const foo = class { [(x)]; }",
      output: "const foo = class { [x]; }",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "class foo { [(x = y)]; }",
      output: "class foo { [x = y]; }",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "class foo { [(x + y)]; }",
      output: "class foo { [x + y]; }",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "class foo { [(x ? y : z)]; }",
      output: "class foo { [x ? y : z]; }",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "class foo { [((x, y))]; }",
      output: "class foo { [(x, y)]; }",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "class foo { x = (y); }",
      output: "class foo { x = y; }",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "class foo { static x = (y); }",
      output: "class foo { static x = y; }",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "class foo { #x = (y); }",
      output: "class foo { #x = y; }",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "class foo { static #x = (y); }",
      output: "class foo { static #x = y; }",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "const foo = class { x = (y); }",
      output: "const foo = class { x = y; }",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "class foo { x = (() => {}); }",
      output: "class foo { x = () => {}; }",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "class foo { x = (y + z); }",
      output: "class foo { x = y + z; }",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "class foo { x = (y ? z : q); }",
      output: "class foo { x = y ? z : q; }",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "class foo { x = ((y, z)); }",
      output: "class foo { x = (y, z); }",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "var foo = (function*() { if ((yield foo())) { return; } }())",
      output: "var foo = (function*() { if (yield foo()) { return; } }())",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "f((0))",
      output: "f(0)",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "f(0, (1))",
      output: "f(0, 1)",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "!(0)",
      output: "!0",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "a[(1)]",
      output: "a[1]",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(a)(b)",
      output: "a(b)",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(async)",
      output: "async",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(a, b)",
      output: "a, b",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "var a = (b = c);",
      output: "var a = b = c;",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "function f(){ return (a); }",
      output: "function f(){ return a; }",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "[a, (b = c)]",
      output: "[a, b = c]",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "!{a: (b = c)}",
      output: "!{a: b = c}",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "typeof(0)",
      output: "typeof 0",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "typeof (0)",
      output: "typeof 0",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "typeof([])",
      output: "typeof[]",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "typeof ([])",
      output: "typeof []",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "typeof( 0)",
      output: "typeof 0",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "typeof(typeof 5)",
      output: "typeof typeof 5",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "typeof (typeof 5)",
      output: "typeof typeof 5",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "+(+foo)",
      output: "+ +foo",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "-(-foo)",
      output: "- -foo",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "+(-foo)",
      output: "+-foo",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "-(+foo)",
      output: "-+foo",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "-((bar+foo))",
      output: "-(bar+foo)",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "+((bar-foo))",
      output: "+(bar-foo)",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "++(foo)",
      output: "++foo",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "--(foo)",
      output: "--foo",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "++\n(foo)",
      output: "++\nfoo",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "--\n(foo)",
      output: "--\nfoo",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "++(\nfoo)",
      output: "++\nfoo",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "--(\nfoo)",
      output: "--\nfoo",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(foo)++",
      output: "foo++",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(foo)--",
      output: "foo--",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "((foo)\n)++",
      output: "(foo\n)++",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "((foo\n))--",
      output: "(foo\n)--",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "((foo\n)\n)++",
      output: "(foo\n\n)++",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(a\n.b)--",
      output: "a\n.b--",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(a.\nb)++",
      output: "a.\nb++",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(a\n[\nb\n])--",
      output: "a\n[\nb\n]--",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(a || b) ? c : d",
      output: "a || b ? c : d",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "a ? (b = c) : d",
      output: "a ? b = c : d",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "a ? b : (c = d)",
      output: "a ? b : c = d",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(c = d) ? (b) : c",
      output: "(c = d) ? b : c",
      options: ["all",{"conditionalAssign":false}],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(c = d) ? b : (c)",
      output: "(c = d) ? b : c",
      options: ["all",{"conditionalAssign":false}],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(a) ? foo : bar",
      output: "a ? foo : bar",
      options: ["all",{"ternaryOperandBinaryExpressions":false}],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(a()) ? foo : bar",
      output: "a() ? foo : bar",
      options: ["all",{"ternaryOperandBinaryExpressions":false}],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(a.b) ? foo : bar",
      output: "a.b ? foo : bar",
      options: ["all",{"ternaryOperandBinaryExpressions":false}],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(a || b) ? foo : (bar)",
      output: "(a || b) ? foo : bar",
      options: ["all",{"ternaryOperandBinaryExpressions":false}],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "f((a = b))",
      output: "f(a = b)",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "a, (b = c)",
      output: "a, b = c",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "a = (b * c)",
      output: "a = b * c",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "a + (b * c)",
      output: "a + b * c",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(a * b) + c",
      output: "a * b + c",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(a * b) / c",
      output: "a * b / c",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(2) ** 3 ** 4",
      output: "2 ** 3 ** 4",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "2 ** (3 ** 4)",
      output: "2 ** 3 ** 4",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(2 ** 3)",
      output: "2 ** 3",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(2 ** 3) + 1",
      output: "2 ** 3 + 1",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "1 - (2 ** 3)",
      output: "1 - 2 ** 3",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "-((2 ** 3))",
      output: "-(2 ** 3)",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "typeof ((a ** b));",
      output: "typeof (a ** b);",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "((-2)) ** 3",
      output: "(-2) ** 3",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "a = (b * c)",
      output: "a = b * c",
      options: ["all",{"nestedBinaryExpressions":false}],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(b * c)",
      output: "b * c",
      options: ["all",{"nestedBinaryExpressions":false}],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "a = (b = c)",
      output: "a = b = c",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(a).b",
      output: "a.b",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(0)[a]",
      output: "0[a]",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(0.0).a",
      output: "0.0.a",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(123.4).a",
      output: "123.4.a",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(0.0_0).a",
      output: "0.0_0.a",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(0xBEEF).a",
      output: "0xBEEF.a",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(0xBE_EF).a",
      output: "0xBE_EF.a",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(1e6).a",
      output: "1e6.a",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "a[(function() {})]",
      output: "a[function() {}]",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "new (function(){})",
      output: "new function(){}",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "new (\nfunction(){}\n)",
      output: "new \nfunction(){}\n",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "((function foo() {return 1;}))()",
      output: "(function foo() {return 1;})()",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "((function(){ return bar(); })())",
      output: "(function(){ return bar(); })()",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(foo()).bar",
      output: "foo().bar",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(foo.bar()).baz",
      output: "foo.bar().baz",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(foo\n.bar())\n.baz",
      output: "foo\n.bar()\n.baz",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(new foo()).bar",
      output: "new foo().bar",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(new foo())[bar]",
      output: "new foo()[bar]",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(new foo()).bar()",
      output: "new foo().bar()",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(new foo(bar)).baz",
      output: "new foo(bar).baz",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(new foo.bar()).baz",
      output: "new foo.bar().baz",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(new foo.bar()).baz()",
      output: "new foo.bar().baz()",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "new a[(b()).c]",
      output: "new a[b().c]",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(a)()",
      output: "a()",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(a.b)()",
      output: "a.b()",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(a())()",
      output: "a()()",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(a.b())()",
      output: "a.b()()",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(a().b)()",
      output: "a().b()",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(a().b.c)()",
      output: "a().b.c()",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "new (A)",
      output: "new A",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(new A())()",
      output: "new A()()",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(new A(1))()",
      output: "new A(1)()",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "((new A))()",
      output: "(new A)()",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "new (foo\n.baz\n.bar\n.foo.baz)",
      output: "new foo\n.baz\n.bar\n.foo.baz",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "new (foo.baz.bar.baz)",
      output: "new foo.baz.bar.baz",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "new ((a.b())).c",
      output: "new (a.b()).c",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "new ((a().b)).c",
      output: "new (a().b).c",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "new ((a().b().d))",
      output: "new (a().b().d)",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "new ((a())).b.d",
      output: "new (a()).b.d",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "new (a.b).d;",
      output: "new a.b.d;",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "new (new A())();",
      output: "new new A()();",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "new (new A());",
      output: "new new A();",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "new (new A);",
      output: "new new A;",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "new (new a.b);",
      output: "new new a.b;",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(a().b).d;",
      output: "a().b.d;",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(a.b()).d;",
      output: "a.b().d;",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(a.b).d;",
      output: "a.b.d;",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "0, (_ => 0)",
      output: "0, _ => 0",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "(_ => 0), 0",
      output: "_ => 0, 0",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "a = (_ => 0)",
      output: "a = _ => 0",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "_ => (a = 0)",
      output: "_ => a = 0",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "x => (({}))",
      output: "x => ({})",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "new (function(){})",
      output: "new function(){}",
      options: ["functions"],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "new (\nfunction(){}\n)",
      output: "new \nfunction(){}\n",
      options: ["functions"],
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "((function foo() {return 1;}))()",
      output: "(function foo() {return 1;})()",
      options: ["functions"],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "a[(function() {})]",
      output: "a[function() {}]",
      options: ["functions"],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "0, (_ => 0)",
      output: "0, _ => 0",
      options: ["functions"],
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "(_ => 0), 0",
      output: "_ => 0, 0",
      options: ["functions"],
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "a = (_ => 0)",
      output: "a = _ => 0",
      options: ["functions"],
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "var y = (function () {return 1;});",
      output: "var y = function () {return 1;};",
      options: ["functions"],
      errors: [{ messageId: "unexpected", column: 9 }],
    },
    {
      code: "function fn(){\n  return (a==b)\n}",
      output: "function fn(){\n  return a==b\n}",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "while ((foo = bar())) {}",
      output: "while (foo = bar()) {}",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "while ((foo = bar())) {}",
      output: "while (foo = bar()) {}",
      options: ["all",{"conditionalAssign":true}],
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "if ((foo = bar())) {}",
      output: "if (foo = bar()) {}",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "do; while ((foo = bar()))",
      output: "do; while (foo = bar())",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "for (;(a = b););",
      output: "for (;a = b;);",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "((function(){})).foo();",
      output: "(function(){}).foo();",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "((function(){}).foo());",
      output: "(function(){}).foo();",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "((function(){}).foo);",
      output: "(function(){}).foo;",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "0, (function(){}).foo();",
      output: "0, function(){}.foo();",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "void (function(){}).foo();",
      output: "void function(){}.foo();",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "++(function(){}).foo;",
      output: "++function(){}.foo;",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "bar || (function(){}).foo();",
      output: "bar || function(){}.foo();",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "1 + (function(){}).foo();",
      output: "1 + function(){}.foo();",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "bar ? (function(){}).foo() : baz;",
      output: "bar ? function(){}.foo() : baz;",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "bar ? baz : (function(){}).foo();",
      output: "bar ? baz : function(){}.foo();",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "bar((function(){}).foo(), 0);",
      output: "bar(function(){}.foo(), 0);",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "bar[(function(){}).foo()];",
      output: "bar[function(){}.foo()];",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "var bar = (function(){}).foo();",
      output: "var bar = function(){}.foo();",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "((class{})).foo();",
      output: "(class{}).foo();",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "((class{}).foo());",
      output: "(class{}).foo();",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "((class{}).foo);",
      output: "(class{}).foo;",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "0, (class{}).foo();",
      output: "0, class{}.foo();",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "void (class{}).foo();",
      output: "void class{}.foo();",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "++(class{}).foo;",
      output: "++class{}.foo;",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "bar || (class{}).foo();",
      output: "bar || class{}.foo();",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "1 + (class{}).foo();",
      output: "1 + class{}.foo();",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "bar ? (class{}).foo() : baz;",
      output: "bar ? class{}.foo() : baz;",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "bar ? baz : (class{}).foo();",
      output: "bar ? baz : class{}.foo();",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "bar((class{}).foo(), 0);",
      output: "bar(class{}.foo(), 0);",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "bar[(class{}).foo()];",
      output: "bar[class{}.foo()];",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "var bar = (class{}).foo();",
      output: "var bar = class{}.foo();",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "var foo = ((bar, baz));",
      output: "var foo = (bar, baz);",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "function *a() { yield (b); }",
      output: "function *a() { yield b; }",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "function *a() { (yield b), c; }",
      output: "function *a() { yield b, c; }",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "function *a() { yield ((b, c)); }",
      output: "function *a() { yield (b, c); }",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "function *a() { yield (b + c); }",
      output: "function *a() { yield b + c; }",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "function a() {\n    return (b);\n}",
      output: "function a() {\n    return b;\n}",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "function a() {\n    return\n    (b);\n}",
      output: "function a() {\n    return\n    b;\n}",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "function a() {\n    return ((\n       b\n    ));\n}",
      output: "function a() {\n    return (\n       b\n    );\n}",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "function a() {\n    return (<JSX />);\n}",
      output: "function a() {\n    return <JSX />;\n}",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "function a() {\n    return\n    (<JSX />);\n}",
      output: "function a() {\n    return\n    <JSX />;\n}",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "function a() {\n    return ((\n       <JSX />\n    ));\n}",
      output: "function a() {\n    return (\n       <JSX />\n    );\n}",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "function a() {\n    return ((\n       <></>\n    ));\n}",
      output: "function a() {\n    return (\n       <></>\n    );\n}",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "throw (a);",
      output: "throw a;",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "throw ((\n   a\n));",
      output: "throw (\n   a\n);",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "function *a() {\n    yield (b);\n}",
      output: "function *a() {\n    yield b;\n}",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "function *a() {\n    yield\n    (b);\n}",
      output: "function *a() {\n    yield\n    b;\n}",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "function *a() {\n    yield ((\n       b\n    ));\n}",
      output: "function *a() {\n    yield (\n       b\n    );\n}",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "function a(b) { return (b || c); }",
      output: "function a(b) { return b || c; }",
      options: ["all",{"returnAssign":false}],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "function a(b) { return ((b = c) || (d = e)); }",
      output: "function a(b) { return (b = c) || (d = e); }",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "function a(b) { return (b = 1); }",
      output: "function a(b) { return b = 1; }",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "function a(b) { return c ? (d = b) : (e = b); }",
      output: "function a(b) { return c ? d = b : e = b; }",
      errors: [{ messageId: "unexpected" }, { messageId: "unexpected" }],
    },
    {
      code: "b => (b || c);",
      output: "b => b || c;",
      options: ["all",{"returnAssign":false}],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "b => ((b = c) || (d = e));",
      output: "b => (b = c) || (d = e);",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "b => (b = 1);",
      output: "b => b = 1;",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "b => c ? (d = b) : (e = b);",
      output: "b => c ? d = b : e = b;",
      errors: [{ messageId: "unexpected" }, { messageId: "unexpected" }],
    },
    {
      code: "b => { return (b || c); }",
      output: "b => { return b || c; }",
      options: ["all",{"returnAssign":false}],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "b => { return ((b = c) || (d = e)) };",
      output: "b => { return (b = c) || (d = e) };",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "b => { return (b = 1) };",
      output: "b => { return b = 1 };",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "b => { return c ? (d = b) : (e = b); }",
      output: "b => { return c ? d = b : e = b; }",
      errors: [{ messageId: "unexpected" }, { messageId: "unexpected" }],
    },
    {
      code: "\n        ((a, b) => {\n          return (\n            a % b == 0\n          ) || (a % b == 1)\n        })()\n      ",
      output: "\n        ((a, b) => {\n          return (\n            a % b == 0\n          ) || a % b == 1\n        })()\n      ",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "\n        ((a, b) => {\n          return (\n            (a % b == 0)\n            || a % b == 1\n          )\n        })()\n      ",
      output: "\n        ((a, b) => {\n          return (\n            a % b == 0\n            || a % b == 1\n          )\n        })()\n      ",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "\n        ((a, b) => {\n          return (a % b == 0)\n            || (a % b == 1)\n        })()\n      ",
      output: "\n        ((a, b) => {\n          return a % b == 0\n            || a % b == 1\n        })()\n      ",
      errors: [{ messageId: "unexpected" }, { messageId: "unexpected" }],
    },
    {
      code: "\n        (a, b) => {\n          return (a % b == 0) || (a % b == 1)\n        }\n      ",
      output: "\n        (a, b) => {\n          return a % b == 0 || a % b == 1\n        }\n      ",
      errors: [{ messageId: "unexpected" }, { messageId: "unexpected" }],
    },
    {
      code: "async function a() { (await a) + (await b); }",
      output: "async function a() { await a + await b; }",
      errors: [{ messageId: "unexpected" }, { messageId: "unexpected" }],
    },
    {
      code: "async function a() { await (a); }",
      output: "async function a() { await a; }",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "async function a() { await (a()); }",
      output: "async function a() { await a(); }",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "async function a() { await (+a); }",
      output: "async function a() { await +a; }",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "async function a() { +(await a); }",
      output: "async function a() { +await a; }",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "async function a() { await ((a,b)); }",
      output: "async function a() { await (a,b); }",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "async function a() { a ** (await b); }",
      output: "async function a() { a ** await b; }",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(foo) instanceof bar",
      output: "foo instanceof bar",
      options: ["all",{"nestedBinaryExpressions":false}],
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "(foo) in bar",
      output: "foo in bar",
      options: ["all",{"nestedBinaryExpressions":false}],
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "(foo) + bar",
      output: "foo + bar",
      options: ["all",{"nestedBinaryExpressions":false}],
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "(foo) && bar",
      output: "foo && bar",
      options: ["all",{"nestedBinaryExpressions":false}],
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "foo instanceof (bar)",
      output: "foo instanceof bar",
      options: ["all",{"nestedBinaryExpressions":false}],
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "foo in (bar)",
      output: "foo in bar",
      options: ["all",{"nestedBinaryExpressions":false}],
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "foo + (bar)",
      output: "foo + bar",
      options: ["all",{"nestedBinaryExpressions":false}],
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "foo && (bar)",
      output: "foo && bar",
      options: ["all",{"nestedBinaryExpressions":false}],
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "const Component = (<div />);",
      output: "const Component = <div />;",
      options: ["all",{"ignoreJSX":"multi-line"}],
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "const Component = (\n  <div />\n);",
      output: "const Component = \n  <div />\n;",
      options: ["all",{"ignoreJSX":"multi-line"}],
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "const Component = (\n  <></>\n);",
      output: "const Component = \n  <></>\n;",
      options: ["all",{"ignoreJSX":"multi-line"}],
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "const Component = (\n<div>\n  <p />\n</div>\n);",
      output: "const Component = \n<div>\n  <p />\n</div>\n;",
      options: ["all",{"ignoreJSX":"single-line"}],
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "const Component = (<div>\n  <p />\n</div>);",
      output: "const Component = <div>\n  <p />\n</div>;",
      options: ["all",{"ignoreJSX":"single-line"}],
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "const Component = (<div\n  prop={true}\n/>)",
      output: "const Component = <div\n  prop={true}\n/>",
      options: ["all",{"ignoreJSX":"single-line"}],
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "const Component = (<div />);",
      output: "const Component = <div />;",
      options: ["all",{"ignoreJSX":"none"}],
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "const Component = (<div>\n<p />\n</div>)",
      output: "const Component = <div>\n<p />\n</div>",
      options: ["all",{"ignoreJSX":"none"}],
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "var a = (b) => (1 ? 2 : 3)",
      output: "var a = (b) => 1 ? 2 : 3",
      options: ["all",{"enforceForArrowConditionals":true}],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(a, b)",
      output: "a, b",
      options: ["all"],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(a, b)",
      output: "a, b",
      options: ["all",{}],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(a, b)",
      output: "a, b",
      options: ["all",{"enforceForSequenceExpressions":true}],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(foo(), bar());",
      output: "foo(), bar();",
      options: ["all",{"enforceForSequenceExpressions":true}],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "if((a, b)){}",
      output: "if(a, b){}",
      options: ["all",{"enforceForSequenceExpressions":true}],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "while ((val = foo(), val < 10));",
      output: "while (val = foo(), val < 10);",
      options: ["all",{"enforceForSequenceExpressions":true}],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(new foo()).bar",
      output: "new foo().bar",
      options: ["all"],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(new foo()).bar",
      output: "new foo().bar",
      options: ["all",{}],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(new foo()).bar",
      output: "new foo().bar",
      options: ["all",{"enforceForNewInMemberExpressions":true}],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(new foo())[bar]",
      output: "new foo()[bar]",
      options: ["all",{"enforceForNewInMemberExpressions":true}],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(new foo.bar()).baz",
      output: "new foo.bar().baz",
      options: ["all",{"enforceForNewInMemberExpressions":true}],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "var foo = (function(){}).call()",
      output: "var foo = function(){}.call()",
      options: ["all"],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "var foo = (function(){}.apply())",
      output: "var foo = function(){}.apply()",
      options: ["all"],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "var foo = (function(){}).apply()",
      output: "var foo = function(){}.apply()",
      options: ["all",{}],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "var foo = (function(){}.call())",
      output: "var foo = function(){}.call()",
      options: ["all",{}],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "var foo = (function(){}).call()",
      output: "var foo = function(){}.call()",
      options: ["all",{"enforceForFunctionPrototypeMethods":true}],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "var foo = (function(){}).apply()",
      output: "var foo = function(){}.apply()",
      options: ["all",{"enforceForFunctionPrototypeMethods":true}],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "var foo = (function(){}.call())",
      output: "var foo = function(){}.call()",
      options: ["all",{"enforceForFunctionPrototypeMethods":true}],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "var foo = (function(){}.apply())",
      output: "var foo = function(){}.apply()",
      options: ["all",{"enforceForFunctionPrototypeMethods":true}],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "var foo = (function(){}.call)()",
      output: "var foo = function(){}.call()",
      options: ["all",{"enforceForFunctionPrototypeMethods":false}],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "var foo = (function(){}.apply)()",
      output: "var foo = function(){}.apply()",
      options: ["all",{"enforceForFunctionPrototypeMethods":false}],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "var foo = (function(){}).call",
      output: "var foo = function(){}.call",
      options: ["all",{"enforceForFunctionPrototypeMethods":false}],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "var foo = (function(){}.call)",
      output: "var foo = function(){}.call",
      options: ["all",{"enforceForFunctionPrototypeMethods":false}],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "var foo = new (function(){}).call()",
      output: "var foo = new function(){}.call()",
      options: ["all",{"enforceForFunctionPrototypeMethods":false}],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "var foo = (new function(){}.call())",
      output: "var foo = new function(){}.call()",
      options: ["all",{"enforceForFunctionPrototypeMethods":false}],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "var foo = (function(){})[call]()",
      output: "var foo = function(){}[call]()",
      options: ["all",{"enforceForFunctionPrototypeMethods":false}],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "var foo = (function(){}[apply]())",
      output: "var foo = function(){}[apply]()",
      options: ["all",{"enforceForFunctionPrototypeMethods":false}],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "var foo = (function(){}).bar()",
      output: "var foo = function(){}.bar()",
      options: ["all",{"enforceForFunctionPrototypeMethods":false}],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "var foo = (function(){}.bar())",
      output: "var foo = function(){}.bar()",
      options: ["all",{"enforceForFunctionPrototypeMethods":false}],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "var foo = (function(){}).call.call()",
      output: "var foo = function(){}.call.call()",
      options: ["all",{"enforceForFunctionPrototypeMethods":false}],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "var foo = (function(){}.call.call())",
      output: "var foo = function(){}.call.call()",
      options: ["all",{"enforceForFunctionPrototypeMethods":false}],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "var foo = (call())",
      output: "var foo = call()",
      options: ["all",{"enforceForFunctionPrototypeMethods":false}],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "var foo = (apply())",
      output: "var foo = apply()",
      options: ["all",{"enforceForFunctionPrototypeMethods":false}],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "var foo = (bar).call()",
      output: "var foo = bar.call()",
      options: ["all",{"enforceForFunctionPrototypeMethods":false}],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "var foo = (bar.call())",
      output: "var foo = bar.call()",
      options: ["all",{"enforceForFunctionPrototypeMethods":false}],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "((() => {}).call())",
      output: "(() => {}).call()",
      options: ["all",{"enforceForFunctionPrototypeMethods":false}],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "var foo = function(){}.call((a.b))",
      output: "var foo = function(){}.call(a.b)",
      options: ["all",{"enforceForFunctionPrototypeMethods":false}],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "var foo = function(){}.call((a).b)",
      output: "var foo = function(){}.call(a.b)",
      options: ["all",{"enforceForFunctionPrototypeMethods":false}],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "var foo = function(){}[('call')]()",
      output: "var foo = function(){}['call']()",
      options: ["all",{"enforceForFunctionPrototypeMethods":false}],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "let a = [...(b)]",
      output: "let a = [...b]",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "let a = {...(b)}",
      output: "let a = {...b}",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "let a = {...(b)}",
      output: "let a = {...b}",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "let a = [...((b, c))]",
      output: "let a = [...(b, c)]",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "let a = {...((b, c))}",
      output: "let a = {...(b, c)}",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "let a = {...((b, c))}",
      output: "let a = {...(b, c)}",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "class A extends (B) {}",
      output: "class A extends B {}",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "const A = class extends (B) {}",
      output: "const A = class extends B {}",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "class A extends ((B=C)) {}",
      output: "class A extends (B=C) {}",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "const A = class extends ((B=C)) {}",
      output: "const A = class extends (B=C) {}",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "class A extends ((++foo)) {}",
      output: "class A extends (++foo) {}",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "export default ((a, b))",
      output: "export default (a, b)",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "export default (() => {})",
      output: "export default () => {}",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "export default ((a, b) => a + b)",
      output: "export default (a, b) => a + b",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "export default (a => a)",
      output: "export default a => a",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "export default (a = b)",
      output: "export default a = b",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "export default (a ? b : c)",
      output: "export default a ? b : c",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "export default (a)",
      output: "export default a",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "for (foo of(bar));",
      output: "for (foo of bar);",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "for ((foo) of bar);",
      output: "for (foo of bar);",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "for (foo of (baz = bar));",
      output: "for (foo of baz = bar);",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "function* f() { for (foo of (yield bar)); }",
      output: "function* f() { for (foo of yield bar); }",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "for (foo of ((bar, baz)));",
      output: "for (foo of (bar, baz));",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "for ((foo)in bar);",
      output: "for (foo in bar);",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "for ((foo['bar'])of baz);",
      output: "for (foo['bar']of baz);",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "() => (({ foo: 1 }).foo)",
      output: "() => ({ foo: 1 }).foo",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "(let).foo",
      output: "let.foo",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "for ((let);;);",
      output: "for (let;;);",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "for ((let);[];);",
      output: "for (let;[];);",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "for (((let[a]));;);",
      output: "for ((let[a]);;);",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "for (((let))[a];;);",
      output: "for ((let)[a];;);",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "for (((let[a])).b;;);",
      output: "for ((let[a]).b;;);",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "for (((let))[a].b;;);",
      output: "for ((let)[a].b;;);",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "for (((let)[a]).b;;);",
      output: "for ((let)[a].b;;);",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "for (((let[a]) = b);;);",
      output: "for ((let[a]) = b;;);",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "for (((let)[a]) = b;;);",
      output: "for ((let)[a] = b;;);",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "for (((let)[a] = b);;);",
      output: "for ((let)[a] = b;;);",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "for ((Let[a]);;);",
      output: "for (Let[a];;);",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "for ((lett)[a];;);",
      output: "for (lett[a];;);",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "for ((let) in foo);",
      output: "for (let in foo);",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "for (((let[a])) in foo);",
      output: "for ((let[a]) in foo);",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "for (((let))[a] in foo);",
      output: "for ((let)[a] in foo);",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "for (((let[a])).b in foo);",
      output: "for ((let[a]).b in foo);",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "for (((let))[a].b in foo);",
      output: "for ((let)[a].b in foo);",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "for (((let)[a]).b in foo);",
      output: "for ((let)[a].b in foo);",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "for (((let[a]).b) in foo);",
      output: "for ((let[a]).b in foo);",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "for ((Let[a]) in foo);",
      output: "for (Let[a] in foo);",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "for ((lett)[a] in foo);",
      output: "for (lett[a] in foo);",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "for (((let)) of foo);",
      output: "for ((let) of foo);",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "for (((let)).a of foo);",
      output: "for ((let).a of foo);",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "for (((let))[a] of foo);",
      output: "for ((let)[a] of foo);",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "for (((let).a) of foo);",
      output: "for ((let).a of foo);",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "for (((let[a]).b) of foo);",
      output: "for ((let[a]).b of foo);",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "for (((let).a).b of foo);",
      output: "for ((let).a.b of foo);",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "for (((let).a.b) of foo);",
      output: "for ((let).a.b of foo);",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "for (((let.a).b) of foo);",
      output: "for ((let.a).b of foo);",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "for (((let()).a) of foo);",
      output: "for ((let()).a of foo);",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "for ((Let) of foo);",
      output: "for (Let of foo);",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "for ((lett) of foo);",
      output: "for (lett of foo);",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "for (a in (b, c));",
      output: "for (a in b, c);",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "for (a of (b));",
      output: "for (a of b);",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(let)",
      output: "let",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "let s = `${(v)}`",
      output: "let s = `${v}`",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "let s = `${(a, b)}`",
      output: "let s = `${a, b}`",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "function foo(a = (b)) {}",
      output: "function foo(a = b) {}",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "const bar = (a = (b)) => a",
      output: "const bar = (a = b) => a",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "const [a = (b)] = []",
      output: "const [a = b] = []",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "const {a = (b)} = {}",
      output: "const {a = b} = {}",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(a) = b",
      output: "a = b",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(a.b) = c",
      output: "a.b = c",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(a) += b",
      output: "a += b",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(a.b) >>= c",
      output: "a.b >>= c",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "[(a) = b] = []",
      output: "[a = b] = []",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "[(a.b) = c] = []",
      output: "[a.b = c] = []",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "({ a: (b) = c } = {})",
      output: "({ a: b = c } = {})",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "({ a: (b.c) = d } = {})",
      output: "({ a: b.c = d } = {})",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "[(a)] = []",
      output: "[a] = []",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "[(a.b)] = []",
      output: "[a.b] = []",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "[,(a),,] = []",
      output: "[,a,,] = []",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "[...(a)] = []",
      output: "[...a] = []",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "[...(a.b)] = []",
      output: "[...a.b] = []",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "({ a: (b) } = {})",
      output: "({ a: b } = {})",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "({ a: (b.c) } = {})",
      output: "({ a: b.c } = {})",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "({ ...(a) } = {})",
      output: "({ ...a } = {})",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "({ ...(a.b) } = {})",
      output: "({ ...a.b } = {})",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "for ((a = (b in c)); ;);",
      output: "for ((a = b in c); ;);",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "for (let a = ((b in c) && (d in e)); ;);",
      output: "for (let a = (b in c && d in e); ;);",
      errors: [{ messageId: "unexpected" }, { messageId: "unexpected" }],
    },
    {
      code: "for (let a = ((b in c) in d); ;);",
      output: "for (let a = (b in c in d); ;);",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "for (let a = (b && (c in d)), e = (f in g); ;);",
      output: "for (let a = (b && c in d), e = (f in g); ;);",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "for (let a = (b + c), d = (e in f); ;);",
      output: "for (let a = b + c, d = (e in f); ;);",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "for (let [a = (b in c)] = []; ;);",
      output: "for (let [a = b in c] = []; ;);",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "for (let [a = b && (c in d)] = []; ;);",
      output: "for (let [a = b && c in d] = []; ;);",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "for (let a = { a: (b in c) }; ;);",
      output: "for (let a = { a: b in c }; ;);",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "for (let a = { a: b && (c in d) }; ;);",
      output: "for (let a = { a: b && c in d }; ;);",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "for (let { a = (b in c) } = {}; ;);",
      output: "for (let { a = b in c } = {}; ;);",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "for (let { a = b && (c in d) } = {}; ;);",
      output: "for (let { a = b && c in d } = {}; ;);",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "for (let { a: { b = c && (d in e) } } = {}; ;);",
      output: "for (let { a: { b = c && d in e } } = {}; ;);",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "for (let a = `${(a in b)}`; ;);",
      output: "for (let a = `${a in b}`; ;);",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "for (let a = `${a && (b in c)}`; ;);",
      output: "for (let a = `${a && b in c}`; ;);",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "for (let a = b((c in d)); ;);",
      output: "for (let a = b(c in d); ;);",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "for (let a = b(c, (d in e)); ;);",
      output: "for (let a = b(c, d in e); ;);",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "for (let a = b(c && (d in e)); ;);",
      output: "for (let a = b(c && d in e); ;);",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "for (let a = b(c, d && (e in f)); ;);",
      output: "for (let a = b(c, d && e in f); ;);",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "for (let a = new b((c in d)); ;);",
      output: "for (let a = new b(c in d); ;);",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "for (let a = new b(c, (d in e)); ;);",
      output: "for (let a = new b(c, d in e); ;);",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "for (let a = new b(c && (d in e)); ;);",
      output: "for (let a = new b(c && d in e); ;);",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "for (let a = new b(c, d && (e in f)); ;);",
      output: "for (let a = new b(c, d && e in f); ;);",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "for (let a = b[(c in d)]; ;);",
      output: "for (let a = b[c in d]; ;);",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "for (let a = b[c && (d in e)]; ;);",
      output: "for (let a = b[c && d in e]; ;);",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "for (let a = b ? (c in d) : e; ;);",
      output: "for (let a = b ? c in d : e; ;);",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "for (let a = b ? c && (d in e) : f; ;);",
      output: "for (let a = b ? c && d in e : f; ;);",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "for (a ? b && (c in d) : e; ;);",
      output: "for (a ? b && c in d : e; ;);",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "for (let a = ((b in c)); ;);",
      output: "for (let a = (b in c); ;);",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "for (((a in b)); ;);",
      output: "for ((a in b); ;);",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "for (((a && b in c && d)); ;);",
      output: "for ((a && b in c && d); ;);",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "for (let a = (!(b in c)); ;);",
      output: "for (let a = !(b in c); ;);",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "for (let a = (!(b && c in d)); ;);",
      output: "for (let a = !(b && c in d); ;);",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "for (let a = !((b in c) && (d in e)); ;);",
      output: "for (let a = !(b in c && d in e); ;);",
      errors: [{ messageId: "unexpected" }, { messageId: "unexpected" }],
    },
    {
      code: "for (let a = (x && (b in c)), d = () => { for ((e in f); ;); for ((g in h); ;); }; ;); for((i in j); ;);",
      output: "for (let a = (x && b in c), d = () => { for ((e in f); ;); for ((g in h); ;); }; ;); for((i in j); ;);",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "for (let a = (b in c), d = () => { for ((x && (e in f)); ;); for ((g in h); ;); }; ;); for((i in j); ;);",
      output: "for (let a = (b in c), d = () => { for ((x && e in f); ;); for ((g in h); ;); }; ;); for((i in j); ;);",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "for (let a = (b in c), d = () => { for ((e in f); ;); for ((x && (g in h)); ;); }; ;); for((i in j); ;);",
      output: "for (let a = (b in c), d = () => { for ((e in f); ;); for ((x && g in h); ;); }; ;); for((i in j); ;);",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "for (let a = (b in c), d = () => { for ((e in f); ;); for ((g in h); ;); }; ;); for((x && (i in j)); ;);",
      output: "for (let a = (b in c), d = () => { for ((e in f); ;); for ((g in h); ;); }; ;); for((x && i in j); ;);",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "for (let a = (x && (b in c)), d = () => { for ((e in f); ;); for ((y && (g in h)); ;); }; ;); for((i in j); ;);",
      output: "for (let a = (x && b in c), d = () => { for ((e in f); ;); for ((y && g in h); ;); }; ;); for((i in j); ;);",
      errors: [{ messageId: "unexpected" }, { messageId: "unexpected" }],
    },
    {
      code: "for (let a = (x && (b in c)), d = () => { for ((y && (e in f)); ;); for ((z && (g in h)); ;); }; ;); for((w && (i in j)); ;);",
      output: "for (let a = (x && b in c), d = () => { for ((y && e in f); ;); for ((z && g in h); ;); }; ;); for((w && i in j); ;);",
      errors: [{ messageId: "unexpected" }, { messageId: "unexpected" }, { messageId: "unexpected" }, { messageId: "unexpected" }],
    },
    {
      code: "for (let a = (b); a > (b); a = (b)) a = (b); a = (b);",
      output: "for (let a = b; a > b; a = b) a = b; a = b;",
      errors: [{ messageId: "unexpected" }, { messageId: "unexpected" }, { messageId: "unexpected" }, { messageId: "unexpected" }, { messageId: "unexpected" }],
    },
    {
      code: "for ((a = b); (a > b); (a = b)) (a = b); (a = b);",
      output: "for (a = b; a > b; a = b) a = b; a = b;",
      errors: [{ messageId: "unexpected" }, { messageId: "unexpected" }, { messageId: "unexpected" }, { messageId: "unexpected" }, { messageId: "unexpected" }],
    },
    {
      code: "for (let a = b; a > (b); a = (b)) a = (b); a = (b);",
      output: "for (let a = b; a > b; a = b) a = b; a = b;",
      errors: [{ messageId: "unexpected" }, { messageId: "unexpected" }, { messageId: "unexpected" }, { messageId: "unexpected" }],
    },
    {
      code: "for (let a = b; (a > b); (a = b)) (a = b); (a = b);",
      output: "for (let a = b; a > b; a = b) a = b; a = b;",
      errors: [{ messageId: "unexpected" }, { messageId: "unexpected" }, { messageId: "unexpected" }, { messageId: "unexpected" }],
    },
    {
      code: "for (; a > (b); a = (b)) a = (b); a = (b);",
      output: "for (; a > b; a = b) a = b; a = b;",
      errors: [{ messageId: "unexpected" }, { messageId: "unexpected" }, { messageId: "unexpected" }, { messageId: "unexpected" }],
    },
    {
      code: "for (; (a > b); (a = b)) (a = b); (a = b);",
      output: "for (; a > b; a = b) a = b; a = b;",
      errors: [{ messageId: "unexpected" }, { messageId: "unexpected" }, { messageId: "unexpected" }, { messageId: "unexpected" }],
    },
    {
      code: "for (let a = (b); a = (b in c); a = (b in c)) a = (b in c); a = (b in c);",
      output: "for (let a = b; a = b in c; a = b in c) a = b in c; a = b in c;",
      errors: [{ messageId: "unexpected" }, { messageId: "unexpected" }, { messageId: "unexpected" }, { messageId: "unexpected" }, { messageId: "unexpected" }],
    },
    {
      code: "for (let a = (b); (a in b); (a in b)) (a in b); (a in b);",
      output: "for (let a = b; a in b; a in b) a in b; a in b;",
      errors: [{ messageId: "unexpected" }, { messageId: "unexpected" }, { messageId: "unexpected" }, { messageId: "unexpected" }, { messageId: "unexpected" }],
    },
    {
      code: "for (let a = b; a = (b in c); a = (b in c)) a = (b in c); a = (b in c);",
      output: "for (let a = b; a = b in c; a = b in c) a = b in c; a = b in c;",
      errors: [{ messageId: "unexpected" }, { messageId: "unexpected" }, { messageId: "unexpected" }, { messageId: "unexpected" }],
    },
    {
      code: "for (let a = b; (a in b); (a in b)) (a in b); (a in b);",
      output: "for (let a = b; a in b; a in b) a in b; a in b;",
      errors: [{ messageId: "unexpected" }, { messageId: "unexpected" }, { messageId: "unexpected" }, { messageId: "unexpected" }],
    },
    {
      code: "for (; a = (b in c); a = (b in c)) a = (b in c); a = (b in c);",
      output: "for (; a = b in c; a = b in c) a = b in c; a = b in c;",
      errors: [{ messageId: "unexpected" }, { messageId: "unexpected" }, { messageId: "unexpected" }, { messageId: "unexpected" }],
    },
    {
      code: "for (; (a in b); (a in b)) (a in b); (a in b);",
      output: "for (; a in b; a in b) a in b; a in b;",
      errors: [{ messageId: "unexpected" }, { messageId: "unexpected" }, { messageId: "unexpected" }, { messageId: "unexpected" }],
    },
    {
      code: "for (let a = (b + c), d = () => { for ((e + f); ;); for ((g + h); ;); }; ;); for((i + j); ;);",
      output: "for (let a = b + c, d = () => { for (e + f; ;); for (g + h; ;); }; ;); for(i + j; ;);",
      errors: [{ messageId: "unexpected" }, { messageId: "unexpected" }, { messageId: "unexpected" }, { messageId: "unexpected" }],
    },
    {
      code: "import((source))",
      output: "import(source)",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "import((source = 'foo.js'))",
      output: "import(source = 'foo.js')",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "import(((s,t)))",
      output: "import((s,t))",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    {
      code: "[1, ((2, 3))];",
      output: "[1, (2, 3)];",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "const foo = () => ((bar, baz));",
      output: "const foo = () => (bar, baz);",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "foo = ((bar, baz));",
      output: "foo = (bar, baz);",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "foo + ((bar + baz));",
      output: "foo + (bar + baz);",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "foo * ((bar + baz));",
      output: "foo * (bar + baz);",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "((foo + bar)) * baz;",
      output: "(foo + bar) * baz;",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "new A(((foo, bar)))",
      output: "new A((foo, bar))",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "class A{ [((foo, bar))]() {} }",
      output: "class A{ [(foo, bar)]() {} }",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "new ((A, B))()",
      output: "new (A, B)()",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "((foo, bar)) ? bar : baz;",
      output: "(foo, bar) ? bar : baz;",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "((f ? o : o)) ? bar : baz;",
      output: "(f ? o : o) ? bar : baz;",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "((f = oo)) ? bar : baz;",
      output: "(f = oo) ? bar : baz;",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "foo ? ((bar, baz)) : baz;",
      output: "foo ? (bar, baz) : baz;",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "foo ? bar : ((bar, baz));",
      output: "foo ? bar : (bar, baz);",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "function foo(bar = ((baz1, baz2))) {}",
      output: "function foo(bar = (baz1, baz2)) {}",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "var foo = { bar: ((baz1, baz2)) };",
      output: "var foo = { bar: (baz1, baz2) };",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "var foo = { [((bar1, bar2))]: baz };",
      output: "var foo = { [(bar1, bar2)]: baz };",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "a+/**/(/**/b)",
      output: "a+/**//**/b",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "a+/**/(//\nb)",
      output: "a+/**///\nb",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "a in(/**/b)",
      output: "a in/**/b",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "a in(//\nb)",
      output: "a in//\nb",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "a+(/**/b)",
      output: "a+/**/b",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "a+/**/(b)",
      output: "a+/**/b",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "a+(//\nb)",
      output: "a+//\nb",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "a+//\n(b)",
      output: "a+//\nb",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "a+(/^b$/)",
      output: "a+/^b$/",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "a/(/**/b)",
      output: "a/ /**/b",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "a/(//\nb)",
      output: "a/ //\nb",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "a/(/^b$/)",
      output: "a/ /^b$/",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "var v = ((a ?? b)) || c",
      output: "var v = (a ?? b) || c",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "var v = a ?? ((b || c))",
      output: "var v = a ?? (b || c)",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "var v = ((a ?? b)) && c",
      output: "var v = (a ?? b) && c",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "var v = a ?? ((b && c))",
      output: "var v = a ?? (b && c)",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "var v = ((a || b)) ?? c",
      output: "var v = (a || b) ?? c",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "var v = a || ((b ?? c))",
      output: "var v = a || (b ?? c)",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "var v = ((a && b)) ?? c",
      output: "var v = (a && b) ?? c",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "var v = a && ((b ?? c))",
      output: "var v = a && (b ?? c)",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "var v = (a ?? b) ? b : c",
      output: "var v = a ?? b ? b : c",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "var v = (a | b) ?? c | d",
      output: "var v = a | b ?? c | d",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "var v = a | b ?? (c | d)",
      output: "var v = a | b ?? c | d",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "const span = /** {HTMLSpanElement}*/(event.currentTarget);",
      output: "const span = /** {HTMLSpanElement}*/event.currentTarget;",
      options: ["all",{"allowParensAfterCommentPattern":"invalid"}],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "if (/** {Compiler | MultiCompiler} */(options).hooks) console.log('good');",
      output: "if (/** {Compiler | MultiCompiler} */options.hooks) console.log('good');",
      options: ["all",{"allowParensAfterCommentPattern":"invalid"}],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "if (condition) {\n    /** @type {ServerOptions} */\n    /** extra comment */\n    (options.server.options).requestCert = false;\n}",
      output: "if (condition) {\n    /** @type {ServerOptions} */\n    /** extra comment */\n    options.server.options.requestCert = false;\n}",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "if (condition) {\n    /** @type {ServerOptions} */\n    ((options.server.options)).requestCert = false;\n}",
      output: "if (condition) {\n    /** @type {ServerOptions} */\n    (options.server.options).requestCert = false;\n}",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "if (condition) {\n    /** @type {ServerOptions} */\n    let foo = \"bar\";\n    (options.server.options).requestCert = false;\n}",
      output: "if (condition) {\n    /** @type {ServerOptions} */\n    let foo = \"bar\";\n    options.server.options.requestCert = false;\n}",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "var v = (obj?.aaa)?.aaa",
      output: "var v = obj?.aaa?.aaa",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "var v = (obj.aaa)?.aaa",
      output: "var v = obj.aaa?.aaa",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "var foo = (function(){})?.call()",
      output: "var foo = function(){}?.call()",
      options: ["all",{"enforceForFunctionPrototypeMethods":true}],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "var foo = (function(){}?.call())",
      output: "var foo = function(){}?.call()",
      options: ["all",{"enforceForFunctionPrototypeMethods":true}],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(Object.prototype.toString.call())",
      output: "Object.prototype.toString.call()",
      options: ["all"],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(a) = function foo() {};",
      output: "a = function foo() {};",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(a) = class Bar {};",
      output: "a = class Bar {};",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(a.b) = function () {};",
      output: "a.b = function () {};",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(newClass) = [(one)] = class { static * [Symbol.iterator]() { yield 1; } };",
      output: "newClass = [one] = class { static * [Symbol.iterator]() { yield 1; } };",
      errors: [{ messageId: "unexpected" }, { messageId: "unexpected" }],
    },
    {
      code: "((a)) = () => {};",
      output: "(a) = () => {};",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(a) = (function () {})();",
      output: "a = (function () {})();",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(a) **= function () {};",
      output: "a **= function () {};",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(a) *= function () {};",
      output: "a *= function () {};",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(a) /= function () {};",
      output: "a /= function () {};",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(a) %= function () {};",
      output: "a %= function () {};",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(a) += function () {};",
      output: "a += function () {};",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(a) -= function () {};",
      output: "a -= function () {};",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(a) <<= function () {};",
      output: "a <<= function () {};",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(a) >>= function () {};",
      output: "a >>= function () {};",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(a) >>>= function () {};",
      output: "a >>>= function () {};",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(a) &= function () {};",
      output: "a &= function () {};",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(a) ^= function () {};",
      output: "a ^= function () {};",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(a) |= function () {};",
      output: "a |= function () {};",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "('use strict');",
      output: null,
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "function f() { ('abc'); }",
      output: null,
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(function () { ('abc'); })();",
      output: null,
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "_ = () => { ('abc'); };",
      output: null,
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "'use strict';(\"foobar\");",
      output: null,
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "foo(); ('bar');",
      output: null,
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "foo = { bar() { ; (\"baz\"); } };",
      output: null,
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(12345);",
      output: "12345;",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(('foobar'));",
      output: "('foobar');",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "(`foobar`);",
      output: "`foobar`;",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "void ('foobar');",
      output: "void 'foobar';",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "_ = () => ('abc');",
      output: "_ = () => 'abc';",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "if (foo) ('bar');",
      output: "if (foo) 'bar';",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "const foo = () => ('bar');",
      output: "const foo = () => 'bar';",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "const x = /** @type {number} */ ((((value))))",
      output: "const x = /** @type {number} */ (value)",
    },
    {
      code: "if ((/** @type {A} */ (/** @type {B} */ (/** @type {C} */(expr))))) {}",
      output: "if (/** @type {A} */ (/** @type {B} */ (/** @type {C} */(expr)))) {}",
    },

    // ==== from no-extra-parens._ts_.test.ts ====
    {
      code: "a<import('')>((1));",
      output: "a<import('')>(1);",
      errors: [{ messageId: "unexpected", column: 15 }],
    },
    {
      code: "new a<import('')>((1));",
      output: "new a<import('')>(1);",
      errors: [{ messageId: "unexpected", column: 19 }],
    },
    {
      code: "a<(A)>((1));",
      output: "a<(A)>(1);",
      errors: [{ messageId: "unexpected", column: 8 }],
    },
    {
      code: "a<(A) | number>((1));",
      output: "a<(A) | number>(1);",
      errors: [{ messageId: "unexpected", column: 17 }],
    },
    {
      code: "async function f(arg: Promise<any>) { await (arg); }",
      output: "async function f(arg: Promise<any>) { await arg; }",
      errors: [{ messageId: "unexpected", column: 45 }],
    },
    {
      code: "async function f(arg: any) { await ((arg as Promise<void>)); }",
      output: "async function f(arg: any) { await (arg as Promise<void>); }",
      errors: [{ messageId: "unexpected", column: 37 }],
    },
    {
      code: "class Foo extends ((Bar as any)) {}",
      output: "class Foo extends (Bar as any) {}",
      errors: [{ messageId: "unexpected", column: 20 }],
    },
    {
      code: "const foo = class extends ((Bar as any)) {}",
      output: "const foo = class extends (Bar as any) {}",
      errors: [{ messageId: "unexpected", column: 28 }],
    },
    {
      code: "const x = (a as string)",
      output: "const x = a as string",
      options: ["all"],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "const x = a[(b as string)]",
      output: "const x = a[b as string]",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "const x = ({} satisfies X)",
      output: "const x = {} satisfies X",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "const x = (foo!)",
      output: "const x = foo!",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "const x = [(b as string)]",
      output: "const x = [b as string]",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "const x: (string) = \"\"",
      output: "const x: string = \"\"",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "function foo(x: (number)): (boolean) {}",
      output: "function foo(x: number): boolean {}",
      errors: [{ messageId: "unexpected" }, { messageId: "unexpected" }],
    },
    {
      code: "type Foo = ({\n  a: <T>(x: T) => any\n})",
      output: "type Foo = {\n  a: <T>(x: T) => any\n}",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "enum Foo {\n  A,\n  B = (\"x\"),\n}",
      output: "enum Foo {\n  A,\n  B = \"x\",\n}",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "type Foo = (string & number) | 'bar'",
      output: "type Foo = string & number | 'bar'",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "type Foo = ((string | number))[]",
      output: "type Foo = (string | number)[]",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "type Foo = ((import('x')))[]",
      output: "type Foo = import('x')[]",
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "type Foo = boolean | ((\n  & Bar\n  & Baz\n))",
      output: "type Foo = boolean | (\n  & Bar\n  & Baz\n)",
      options: ["all",{"nestedBinaryExpressions":false}],
      errors: [{ messageId: "unexpected" }],
    },
    {
      code: "type Foo = boolean | ((Bar & Baz))",
      output: "type Foo = boolean | Bar & Baz",
      errors: [{ messageId: "unexpected" }],
    },
  ],
});

// ---------------------------------------------------------------------------
// KNOWN GAPS: cases preserved verbatim but isolated because they surface a real
// rslint<->upstream difference (annotated per case). They are NOT run.
// ---------------------------------------------------------------------------
const KNOWN_GAPS = {
  valid: [
  // ---- js valid ----
    // GAP: ts-go parser gap: rslint flags the parens as unnecessary, but upstream keeps them. In sloppy-mode JS the parens are load-bearing — at statement start an unparenthesised `let[...]` is an ASI/`let`-declaration ambiguity (`let[a]=b` would parse as a destructuring `let` declaration; `(let)[foo]` keeps `let` an identifier). ts-go parses `let` differently, so it sees the parens as redundant -> a 1-diagnostic mismatch (upstream expects 0).
    "(let[a] = b);",
    // GAP: ts-go syntax error TS1489 (Decimals with leading zeros are not allowed). Upstream (sloppy-mode JS) parses 08 as a decimal and reports the extra parens; ts-go rejects the literal outright, so the fixture is unparseable.
    "(08).a",
    // GAP: ts-go syntax error TS1489 (Decimals with leading zeros are not allowed). Upstream (sloppy-mode JS via espree) parses the leading-zero number as a decimal and reports the extra parens; ts-go rejects the literal outright, so the fixture is unparseable.
    "(09).a",
    // GAP: ts-go syntax error TS1489 (Decimals with leading zeros are not allowed). Upstream (sloppy-mode JS via espree) parses the leading-zero number as a decimal and reports the extra parens; ts-go rejects the literal outright, so the fixture is unparseable.
    "(018).a",
    // GAP: ts-go syntax error TS1489 (Decimals with leading zeros are not allowed). Upstream (sloppy-mode JS via espree) parses the leading-zero number as a decimal and reports the extra parens; ts-go rejects the literal outright, so the fixture is unparseable.
    "(012934).a",
    // GAP: ts-go parser gap: rslint flags the parens as unnecessary, but upstream keeps them. In sloppy-mode JS the parens are load-bearing — at statement start an unparenthesised `let[...]` is an ASI/`let`-declaration ambiguity (`let[a]=b` would parse as a destructuring `let` declaration; `(let)[foo]` keeps `let` an identifier). ts-go parses `let` differently, so it sees the parens as redundant -> a 1-diagnostic mismatch (upstream expects 0).
    "(let)\nfoo",
    // GAP: ts-go parser gap: rslint flags the parens as unnecessary, but upstream keeps them. In sloppy-mode JS the parens are load-bearing — at statement start an unparenthesised `let[...]` is an ASI/`let`-declaration ambiguity (`let[a]=b` would parse as a destructuring `let` declaration; `(let)[foo]` keeps `let` an identifier). ts-go parses `let` differently, so it sees the parens as redundant -> a 1-diagnostic mismatch (upstream expects 0).
    "(let[foo]) = 1",
    // GAP: ts-go parser gap: rslint flags the parens as unnecessary, but upstream keeps them. In sloppy-mode JS the parens are load-bearing — at statement start an unparenthesised `let[...]` is an ASI/`let`-declaration ambiguity (`let[a]=b` would parse as a destructuring `let` declaration; `(let)[foo]` keeps `let` an identifier). ts-go parses `let` differently, so it sees the parens as redundant -> a 1-diagnostic mismatch (upstream expects 0).
    "(let)[foo]",
  ],
  invalid: [
  // ---- js invalid ----
    // GAP: ts-go syntax error TS1121 (Octal literals are not allowed; use 0o123). Upstream (sloppy-mode JS) parses 0123 as a legacy octal; ts-go rejects it, so the fixture is unparseable.
    {
      code: "(0123).a",
      output: "0123.a",
      errors: [{ messageId: "unexpected" }],
    },
    // GAP: ts-go syntax error TS1489 (Decimals with leading zeros are not allowed). Upstream (sloppy-mode JS via espree) parses the leading-zero number as a decimal and reports the extra parens; ts-go rejects the literal outright, so the fixture is unparseable.
    {
      code: "(08.1).a",
      output: "08.1.a",
      errors: [{ messageId: "unexpected" }],
    },
    // GAP: ts-go syntax error TS1489 (Decimals with leading zeros are not allowed). Upstream (sloppy-mode JS via espree) parses the leading-zero number as a decimal and reports the extra parens; ts-go rejects the literal outright, so the fixture is unparseable.
    {
      code: "(09.).a",
      output: "09..a",
      errors: [{ messageId: "unexpected" }],
    },
    // GAP: Multi-pass --fix divergence (diagnostics align, only `output` differs): upstream's eslint-vitest-rule-tester defaults to a single fix pass (`recursive: false`), so its `output` still contains an inner redundant paren that a second pass would remove. rslint always fixes to a stable point (multi-pass) and removes the inner paren too. Both report the same diagnostic; only the fixed source differs.
    {
      code: "var a = (b) => ((1 ? 2 : 3))",
      output: "var a = (b) => (1 ? 2 : 3)",
      options: ["all",{"enforceForArrowConditionals":true}],
      errors: [{ messageId: "unexpected" }],
    },
    // GAP: ts-go parser gap (TS1134 'Variable declaration expected'): the fixed form starts the for-init with `let` as an identifier (member/call/assignment). Upstream espree (sloppy JS) treats bare `let` as an identifier; ts-go treats `let` in a for-init as a declaration keyword, so the fixed code is unparseable. The unparseable fix also poisons rslint's multi-pass --fix for the whole batch.
    {
      code: "for ((let = 1);;);",
      output: "for (let = 1;;);",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    // GAP: ts-go parser gap (TS1134 'Variable declaration expected'): the fixed form starts the for-init with `let` as an identifier (member/call/assignment). Upstream espree (sloppy JS) treats bare `let` as an identifier; ts-go treats `let` in a for-init as a declaration keyword, so the fixed code is unparseable. The unparseable fix also poisons rslint's multi-pass --fix for the whole batch.
    {
      code: "for ((let) = 1;;);",
      output: "for (let = 1;;);",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    // GAP: ts-go parser gap (TS1134 'Variable declaration expected'): the fixed form starts the for-init with `let` as an identifier (member/call/assignment). Upstream espree (sloppy JS) treats bare `let` as an identifier; ts-go treats `let` in a for-init as a declaration keyword, so the fixed code is unparseable. The unparseable fix also poisons rslint's multi-pass --fix for the whole batch.
    {
      code: "for ((let = []);;);",
      output: "for (let = [];;);",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    // GAP: ts-go parser gap (TS1134 'Variable declaration expected'): the fixed form starts the for-init with `let` as an identifier (member/call/assignment). Upstream espree (sloppy JS) treats bare `let` as an identifier; ts-go treats `let` in a for-init as a declaration keyword, so the fixed code is unparseable. The unparseable fix also poisons rslint's multi-pass --fix for the whole batch.
    {
      code: "for ((let) = [];;);",
      output: "for (let = [];;);",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    // GAP: ts-go parser gap (TS1134 'Variable declaration expected'): the fixed form starts the for-init with `let` as an identifier (member/call/assignment). Upstream espree (sloppy JS) treats bare `let` as an identifier; ts-go treats `let` in a for-init as a declaration keyword, so the fixed code is unparseable. The unparseable fix also poisons rslint's multi-pass --fix for the whole batch.
    {
      code: "for ((let());;);",
      output: "for (let();;);",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    // GAP: ts-go parser gap (TS1134 'Variable declaration expected'): the fixed form starts the for-init with `let` as an identifier (member/call/assignment). Upstream espree (sloppy JS) treats bare `let` as an identifier; ts-go treats `let` in a for-init as a declaration keyword, so the fixed code is unparseable. The unparseable fix also poisons rslint's multi-pass --fix for the whole batch.
    {
      code: "for ((let([]));;);",
      output: "for (let([]);;);",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    // GAP: ts-go parser gap (TS1134 'Variable declaration expected'): the fixed form starts the for-init with `let` as an identifier (member/call/assignment). Upstream espree (sloppy JS) treats bare `let` as an identifier; ts-go treats `let` in a for-init as a declaration keyword, so the fixed code is unparseable. The unparseable fix also poisons rslint's multi-pass --fix for the whole batch.
    {
      code: "for ((let())[a];;);",
      output: "for (let()[a];;);",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    // GAP: ts-go parser gap (TS1134 'Variable declaration expected'): the fixed form starts the for-init with `let` as an identifier (member/call/assignment). Upstream espree (sloppy JS) treats bare `let` as an identifier; ts-go treats `let` in a for-init as a declaration keyword, so the fixed code is unparseable. The unparseable fix also poisons rslint's multi-pass --fix for the whole batch.
    {
      code: "for ((let`[]`);;);",
      output: "for (let`[]`;;);",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    // GAP: ts-go parser gap (TS1134 'Variable declaration expected'): the fixed form starts the for-init with `let` as an identifier (member/call/assignment). Upstream espree (sloppy JS) treats bare `let` as an identifier; ts-go treats `let` in a for-init as a declaration keyword, so the fixed code is unparseable. The unparseable fix also poisons rslint's multi-pass --fix for the whole batch.
    {
      code: "for ((let.a);;);",
      output: "for (let.a;;);",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    // GAP: ts-go parser gap (TS1134 'Variable declaration expected'): the fixed form starts the for-init with `let` as an identifier (member/call/assignment). Upstream espree (sloppy JS) treats bare `let` as an identifier; ts-go treats `let` in a for-init as a declaration keyword, so the fixed code is unparseable. The unparseable fix also poisons rslint's multi-pass --fix for the whole batch.
    {
      code: "for ((let).a;;);",
      output: "for (let.a;;);",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    // GAP: ts-go parser gap (TS1134 'Variable declaration expected'): the fixed form starts the for-init with `let` as an identifier (member/call/assignment). Upstream espree (sloppy JS) treats bare `let` as an identifier; ts-go treats `let` in a for-init as a declaration keyword, so the fixed code is unparseable. The unparseable fix also poisons rslint's multi-pass --fix for the whole batch.
    {
      code: "for ((let).a = 1;;);",
      output: "for (let.a = 1;;);",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    // GAP: ts-go parser gap (TS1134 'Variable declaration expected'): the fixed form starts the for-init with `let` as an identifier (member/call/assignment). Upstream espree (sloppy JS) treats bare `let` as an identifier; ts-go treats `let` in a for-init as a declaration keyword, so the fixed code is unparseable. The unparseable fix also poisons rslint's multi-pass --fix for the whole batch.
    {
      code: "for ((let).a[b];;);",
      output: "for (let.a[b];;);",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    // GAP: ts-go parser gap (TS1134 'Variable declaration expected'): the fixed form starts the for-init with `let` as an identifier (member/call/assignment). Upstream espree (sloppy JS) treats bare `let` as an identifier; ts-go treats `let` in a for-init as a declaration keyword, so the fixed code is unparseable. The unparseable fix also poisons rslint's multi-pass --fix for the whole batch.
    {
      code: "for ((let.a)[b];;);",
      output: "for (let.a[b];;);",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    // GAP: ts-go parser gap (TS1134 'Variable declaration expected'): the fixed form starts the for-init with `let` as an identifier (member/call/assignment). Upstream espree (sloppy JS) treats bare `let` as an identifier; ts-go treats `let` in a for-init as a declaration keyword, so the fixed code is unparseable. The unparseable fix also poisons rslint's multi-pass --fix for the whole batch.
    {
      code: "for ((let.a[b]);;);",
      output: "for (let.a[b];;);",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    // GAP: ts-go parser gap (TS1134 'Variable declaration expected'): the fixed form starts the for-init with `let` as an identifier (member/call/assignment). Upstream espree (sloppy JS) treats bare `let` as an identifier; ts-go treats `let` in a for-init as a declaration keyword, so the fixed code is unparseable. The unparseable fix also poisons rslint's multi-pass --fix for the whole batch.
    {
      code: "for ((let())[a] in foo);",
      output: "for (let()[a] in foo);",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    // GAP: ts-go parser gap (TS1134 'Variable declaration expected'): the fixed form starts the for-init with `let` as an identifier (member/call/assignment). Upstream espree (sloppy JS) treats bare `let` as an identifier; ts-go treats `let` in a for-init as a declaration keyword, so the fixed code is unparseable. The unparseable fix also poisons rslint's multi-pass --fix for the whole batch.
    {
      code: "for ((let.a) in foo);",
      output: "for (let.a in foo);",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    // GAP: ts-go parser gap (TS1134 'Variable declaration expected'): the fixed form starts the for-init with `let` as an identifier (member/call/assignment). Upstream espree (sloppy JS) treats bare `let` as an identifier; ts-go treats `let` in a for-init as a declaration keyword, so the fixed code is unparseable. The unparseable fix also poisons rslint's multi-pass --fix for the whole batch.
    {
      code: "for ((let).a in foo);",
      output: "for (let.a in foo);",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    // GAP: ts-go parser gap (TS1134 'Variable declaration expected'): the fixed form starts the for-init with `let` as an identifier (member/call/assignment). Upstream espree (sloppy JS) treats bare `let` as an identifier; ts-go treats `let` in a for-init as a declaration keyword, so the fixed code is unparseable. The unparseable fix also poisons rslint's multi-pass --fix for the whole batch.
    {
      code: "for ((let).a.b in foo);",
      output: "for (let.a.b in foo);",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    // GAP: ts-go parser gap (TS1134 'Variable declaration expected'): the fixed form starts the for-init with `let` as an identifier (member/call/assignment). Upstream espree (sloppy JS) treats bare `let` as an identifier; ts-go treats `let` in a for-init as a declaration keyword, so the fixed code is unparseable. The unparseable fix also poisons rslint's multi-pass --fix for the whole batch.
    {
      code: "for ((let).a[b] in foo);",
      output: "for (let.a[b] in foo);",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    // GAP: ts-go parser gap (TS1134 'Variable declaration expected'): the fixed form starts the for-init with `let` as an identifier (member/call/assignment). Upstream espree (sloppy JS) treats bare `let` as an identifier; ts-go treats `let` in a for-init as a declaration keyword, so the fixed code is unparseable. The unparseable fix also poisons rslint's multi-pass --fix for the whole batch.
    {
      code: "for ((let.a)[b] in foo);",
      output: "for (let.a[b] in foo);",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    // GAP: ts-go parser gap (TS1134 'Variable declaration expected'): the fixed form starts the for-init with `let` as an identifier (member/call/assignment). Upstream espree (sloppy JS) treats bare `let` as an identifier; ts-go treats `let` in a for-init as a declaration keyword, so the fixed code is unparseable. The unparseable fix also poisons rslint's multi-pass --fix for the whole batch.
    {
      code: "for ((let.a[b]) in foo);",
      output: "for (let.a[b] in foo);",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    // GAP: Multi-pass --fix divergence (diagnostics align, only `output` differs): upstream's eslint-vitest-rule-tester defaults to a single fix pass (`recursive: false`), so its `output` still contains an inner redundant paren that a second pass would remove. rslint always fixes to a stable point (multi-pass) and removes the inner paren too. Both report the same diagnostic; only the fixed source differs. (rslint also fully unwraps `((let))` -> `let`; ts-go accepts bare `let` here as an identifier expression statement.)
    {
      code: "((let))",
      output: "(let)",
      errors: [{ messageId: "unexpected", line: 1 }],
    },
    // GAP: ts-go parser gap (TS1005/TS1434): removing the parens lets an `in` operator sit unparenthesised inside the for-init, where ts-go parses it as a for-in head and rejects the surrounding array/arrow/function (the eslint #11706 shape). Upstream espree parses it as a plain `for(;;)` init. The unparseable fix also poisons rslint's multi-pass --fix for the whole batch.
    {
      code: "for (let a = [(b in c)]; ;);",
      output: "for (let a = [b in c]; ;);",
      errors: [{ messageId: "unexpected" }],
    },
    // GAP: ts-go parser gap (TS1005/TS1434): removing the parens lets an `in` operator sit unparenthesised inside the for-init, where ts-go parses it as a for-in head and rejects the surrounding array/arrow/function (the eslint #11706 shape). Upstream espree parses it as a plain `for(;;)` init. The unparseable fix also poisons rslint's multi-pass --fix for the whole batch.
    {
      code: "for (let a = [b, (c in d)]; ;);",
      output: "for (let a = [b, c in d]; ;);",
      errors: [{ messageId: "unexpected" }],
    },
    // GAP: ts-go parser gap (TS1005/TS1434): removing the parens lets an `in` operator sit unparenthesised inside the for-init, where ts-go parses it as a for-in head and rejects the surrounding array/arrow/function (the eslint #11706 shape). Upstream espree parses it as a plain `for(;;)` init. The unparseable fix also poisons rslint's multi-pass --fix for the whole batch.
    {
      code: "for (let a = ([b in c]); ;);",
      output: "for (let a = [b in c]; ;);",
      errors: [{ messageId: "unexpected" }],
    },
    // GAP: ts-go parser gap (TS1005/TS1434): removing the parens lets an `in` operator sit unparenthesised inside the for-init, where ts-go parses it as a for-in head and rejects the surrounding array/arrow/function (the eslint #11706 shape). Upstream espree parses it as a plain `for(;;)` init. The unparseable fix also poisons rslint's multi-pass --fix for the whole batch.
    {
      code: "for (let a = ([b, c in d]); ;);",
      output: "for (let a = [b, c in d]; ;);",
      errors: [{ messageId: "unexpected" }],
    },
    // GAP: ts-go parser gap (TS1005/TS1434): removing the parens lets an `in` operator sit unparenthesised inside the for-init, where ts-go parses it as a for-in head and rejects the surrounding array/arrow/function (the eslint #11706 shape). Upstream espree parses it as a plain `for(;;)` init. The unparseable fix also poisons rslint's multi-pass --fix for the whole batch.
    {
      code: "for ((a = [b in c]); ;);",
      output: "for (a = [b in c]; ;);",
      errors: [{ messageId: "unexpected" }],
    },
    // GAP: ts-go parser gap (TS1005/TS1434): removing the parens lets an `in` operator sit unparenthesised inside the for-init, where ts-go parses it as a for-in head and rejects the surrounding array/arrow/function (the eslint #11706 shape). Upstream espree parses it as a plain `for(;;)` init. The unparseable fix also poisons rslint's multi-pass --fix for the whole batch.
    {
      code: "for (let a = [b && (c in d)]; ;);",
      output: "for (let a = [b && c in d]; ;);",
      errors: [{ messageId: "unexpected" }],
    },
    // GAP: ts-go parser gap (TS1005/TS1434): removing the parens lets an `in` operator sit unparenthesised inside the for-init, where ts-go parses it as a for-in head and rejects the surrounding array/arrow/function (the eslint #11706 shape). Upstream espree parses it as a plain `for(;;)` init. The unparseable fix also poisons rslint's multi-pass --fix for the whole batch.
    {
      code: "for (let a = [(b && c in d)]; ;);",
      output: "for (let a = [b && c in d]; ;);",
      errors: [{ messageId: "unexpected" }],
    },
    // GAP: ts-go parser gap (TS1005/TS1434): removing the parens lets an `in` operator sit unparenthesised inside the for-init, where ts-go parses it as a for-in head and rejects the surrounding array/arrow/function (the eslint #11706 shape). Upstream espree parses it as a plain `for(;;)` init. The unparseable fix also poisons rslint's multi-pass --fix for the whole batch.
    {
      code: "for (let a = ([b && c in d]); ;);",
      output: "for (let a = [b && c in d]; ;);",
      errors: [{ messageId: "unexpected" }],
    },
    // GAP: ts-go parser gap (TS1005/TS1434): removing the parens lets an `in` operator sit unparenthesised inside the for-init, where ts-go parses it as a for-in head and rejects the surrounding array/arrow/function (the eslint #11706 shape). Upstream espree parses it as a plain `for(;;)` init. The unparseable fix also poisons rslint's multi-pass --fix for the whole batch.
    {
      code: "for ((a = [b && c in d]); ;);",
      output: "for (a = [b && c in d]; ;);",
      errors: [{ messageId: "unexpected" }],
    },
    // GAP: ts-go parser gap (TS1005/TS1434): removing the parens lets an `in` operator sit unparenthesised inside the for-init, where ts-go parses it as a for-in head and rejects the surrounding array/arrow/function (the eslint #11706 shape). Upstream espree parses it as a plain `for(;;)` init. The unparseable fix also poisons rslint's multi-pass --fix for the whole batch.
    {
      code: "for ([(a in b)]; ;);",
      output: "for ([a in b]; ;);",
      errors: [{ messageId: "unexpected" }],
    },
    // GAP: ts-go parser gap (TS1005/TS1434): removing the parens lets an `in` operator sit unparenthesised inside the for-init, where ts-go parses it as a for-in head and rejects the surrounding array/arrow/function (the eslint #11706 shape). Upstream espree parses it as a plain `for(;;)` init. The unparseable fix also poisons rslint's multi-pass --fix for the whole batch.
    {
      code: "for (([a in b]); ;);",
      output: "for ([a in b]; ;);",
      errors: [{ messageId: "unexpected" }],
    },
    // GAP: ts-go parser gap (TS1005/TS1434): removing the parens lets an `in` operator sit unparenthesised inside the for-init, where ts-go parses it as a for-in head and rejects the surrounding array/arrow/function (the eslint #11706 shape). Upstream espree parses it as a plain `for(;;)` init. The unparseable fix also poisons rslint's multi-pass --fix for the whole batch.
    {
      code: "for (let a = [(b in c)], d = (e in f); ;);",
      output: "for (let a = [b in c], d = (e in f); ;);",
      errors: [{ messageId: "unexpected" }],
    },
    // GAP: ts-go parser gap (TS1005/TS1434): removing the parens lets an `in` operator sit unparenthesised inside the for-init, where ts-go parses it as a for-in head and rejects the surrounding array/arrow/function (the eslint #11706 shape). Upstream espree parses it as a plain `for(;;)` init. The unparseable fix also poisons rslint's multi-pass --fix for the whole batch.
    {
      code: "for (let a = () => { (b in c) }; ;);",
      output: "for (let a = () => { b in c }; ;);",
      errors: [{ messageId: "unexpected" }],
    },
    // GAP: ts-go parser gap (TS1005/TS1434): removing the parens lets an `in` operator sit unparenthesised inside the for-init, where ts-go parses it as a for-in head and rejects the surrounding array/arrow/function (the eslint #11706 shape). Upstream espree parses it as a plain `for(;;)` init. The unparseable fix also poisons rslint's multi-pass --fix for the whole batch.
    {
      code: "for (let a = () => { a && (b in c) }; ;);",
      output: "for (let a = () => { a && b in c }; ;);",
      errors: [{ messageId: "unexpected" }],
    },
    // GAP: ts-go parser gap (TS1005/TS1434): removing the parens lets an `in` operator sit unparenthesised inside the for-init, where ts-go parses it as a for-in head and rejects the surrounding array/arrow/function (the eslint #11706 shape). Upstream espree parses it as a plain `for(;;)` init. The unparseable fix also poisons rslint's multi-pass --fix for the whole batch.
    {
      code: "for (let a = function () { (b in c) }; ;);",
      output: "for (let a = function () { b in c }; ;);",
      errors: [{ messageId: "unexpected" }],
    },
    // GAP: ts-go parser gap (TS1005/TS1434): removing the parens lets an `in` operator sit unparenthesised inside the for-init, where ts-go parses it as a for-in head and rejects the surrounding array/arrow/function (the eslint #11706 shape). Upstream espree parses it as a plain `for(;;)` init. The unparseable fix also poisons rslint's multi-pass --fix for the whole batch.
    {
      code: "for (let a = (b = (c in d)) => {}; ;);",
      output: "for (let a = (b = c in d) => {}; ;);",
      errors: [{ messageId: "unexpected" }],
    },
    // GAP: ts-go parser gap (TS1005/TS1434): removing the parens lets an `in` operator sit unparenthesised inside the for-init, where ts-go parses it as a for-in head and rejects the surrounding array/arrow/function (the eslint #11706 shape). Upstream espree parses it as a plain `for(;;)` init. The unparseable fix also poisons rslint's multi-pass --fix for the whole batch.
    {
      code: "for (let a = (b = c && (d in e)) => {}; ;);",
      output: "for (let a = (b = c && d in e) => {}; ;);",
      errors: [{ messageId: "unexpected" }],
    },
    // GAP: ts-go parser gap (TS1005/TS1434): removing the parens lets an `in` operator sit unparenthesised inside the for-init, where ts-go parses it as a for-in head and rejects the surrounding array/arrow/function (the eslint #11706 shape). Upstream espree parses it as a plain `for(;;)` init. The unparseable fix also poisons rslint's multi-pass --fix for the whole batch.
    {
      code: "for (let a = (b, c = d && (e in f)) => {}; ;);",
      output: "for (let a = (b, c = d && e in f) => {}; ;);",
      errors: [{ messageId: "unexpected" }],
    },
    // GAP: ts-go parser gap (TS1005/TS1434): removing the parens lets an `in` operator sit unparenthesised inside the for-init, where ts-go parses it as a for-in head and rejects the surrounding array/arrow/function (the eslint #11706 shape). Upstream espree parses it as a plain `for(;;)` init. The unparseable fix also poisons rslint's multi-pass --fix for the whole batch.
    {
      code: "for (let a = function (b = c && (d in e)) {}; ;);",
      output: "for (let a = function (b = c && d in e) {}; ;);",
      errors: [{ messageId: "unexpected" }],
    },
    // GAP: ts-go parser gap (TS1005/TS1434): removing the parens lets an `in` operator sit unparenthesised inside the for-init, where ts-go parses it as a for-in head and rejects the surrounding array/arrow/function (the eslint #11706 shape). Upstream espree parses it as a plain `for(;;)` init. The unparseable fix also poisons rslint's multi-pass --fix for the whole batch.
    {
      code: "for (let a = function (b, c = d && (e in f)) {}; ;);",
      output: "for (let a = function (b, c = d && e in f) {}; ;);",
      errors: [{ messageId: "unexpected" }],
    },
    // GAP: Multi-pass --fix divergence (diagnostics align, only `output` differs): upstream's eslint-vitest-rule-tester defaults to a single fix pass (`recursive: false`), so its `output` still contains an inner redundant paren that a second pass would remove. rslint always fixes to a stable point (multi-pass) and removes the inner paren too. Both report the same diagnostic; only the fixed source differs.
    {
      code: "((foo + bar)) + baz;",
      output: "(foo + bar) + baz;",
      errors: [{ messageId: "unexpected" }],
    },
  // ---- ts invalid ----
    // GAP: Multi-pass --fix divergence (diagnostics align, only `output` differs): upstream's eslint-vitest-rule-tester defaults to a single fix pass (`recursive: false`), so its `output` still contains an inner redundant paren that a second pass would remove. rslint always fixes to a stable point (multi-pass) and removes the inner paren too. Both report the same diagnostic; only the fixed source differs.
    {
      code: "class A{ accessor [((foo))] = 1 }",
      output: "class A{ accessor [(foo)] = 1 }",
      errors: [{ messageId: "unexpected" }],
    },
  ],
};
void KNOWN_GAPS;

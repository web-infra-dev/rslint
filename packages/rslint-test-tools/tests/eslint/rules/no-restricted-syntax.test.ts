import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-restricted-syntax', {
  valid: [
    // Upstream: trivial valid (no rule options)
    'doSomething();',

    // Upstream: string format — non-matching kinds
    { code: 'var foo = 42;', options: ['ConditionalExpression'] as any },
    {
      code: 'foo += 42;',
      options: ['VariableDeclaration', 'FunctionExpression'] as any,
    },
    { code: 'foo;', options: [`Identifier[name="bar"]`] as any },
    {
      code: '() => 5',
      options: ['ArrowFunctionExpression > BlockStatement'] as any,
    },
    {
      code: '({ foo: 1, bar: 2 })',
      options: ['Property > Literal.key'] as any,
    },
    { code: 'A: for (;;) break;', options: ['BreakStatement[label]'] as any },
    {
      code: 'function foo(bar, baz) {}',
      options: ['FunctionDeclaration[params.length>2]'] as any,
    },

    // Upstream: object format
    {
      code: 'var foo = 42;',
      options: [{ selector: 'ConditionalExpression' }] as any,
    },
    {
      code: '({ foo: 1, bar: 2 })',
      options: [{ selector: 'Property > Literal.key' }] as any,
    },
    {
      code: '({ foo: 1, bar: 2 })',
      options: [
        {
          selector: 'FunctionDeclaration[params.length>2]',
          message: 'custom error message.',
        },
      ] as any,
    },

    // Upstream: regex flag attribute
    {
      code: 'console.log(/a/);',
      options: ['Literal[regex.flags=/./]'] as any,
    },

    // Boundary: empty options array — a no-op rule
    { code: 'var a = 1;', options: [] as any },

    // Boundary: unknown ESTree type silently ignored
    { code: 'var a = 1;', options: ['NotARealNodeType'] as any },

    // Boundary: malformed selectors silently dropped
    { code: 'var a = 1;', options: ['['] as any },

    // BinaryExpression refinement — assignment / logical / comma should
    // NOT match a bare BinaryExpression selector.
    { code: 'a = b;', options: ['BinaryExpression'] as any },
    { code: 'a && b;', options: ['BinaryExpression'] as any },
    { code: '(a, b);', options: ['BinaryExpression'] as any },

    // ChainExpression / [optional=true] should not match plain access.
    { code: 'a.b.c;', options: ['ChainExpression'] as any },
    { code: 'a.b.c;', options: ['[optional=true]'] as any },
  ],
  invalid: [
    // Upstream: VariableDeclaration positive
    {
      code: 'var foo = 41;',
      options: ['VariableDeclaration'] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },
    {
      code: ';function lol(a) { return 42; }',
      options: ['EmptyStatement'] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },
    {
      code: 'try { voila(); } catch (e) { oops(); }',
      options: ['TryStatement', 'CallExpression', 'CatchClause'] as any,
      errors: [
        { messageId: 'restrictedSyntax' },
        { messageId: 'restrictedSyntax' },
        { messageId: 'restrictedSyntax' },
        { messageId: 'restrictedSyntax' },
      ],
    },
    {
      code: 'bar;',
      options: [`Identifier[name="bar"]`] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },
    {
      code: 'bar;',
      options: ['Identifier', `Identifier[name="bar"]`] as any,
      errors: [
        { messageId: 'restrictedSyntax' },
        { messageId: 'restrictedSyntax' },
      ],
    },
    {
      code: '() => {}',
      options: ['ArrowFunctionExpression > BlockStatement'] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },
    {
      code: `({ foo: 1, 'bar': 2 })`,
      options: ['Property > Literal.key'] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },
    {
      code: 'A: for (;;) break A;',
      options: ['BreakStatement[label]'] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },
    {
      code: 'function foo(bar, baz, qux) {}',
      options: ['FunctionDeclaration[params.length>2]'] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },

    // Upstream: object format
    {
      code: 'var foo = 41;',
      options: [{ selector: 'VariableDeclaration' }] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },
    {
      code: 'function foo(bar, baz, qux) {}',
      options: [{ selector: 'FunctionDeclaration[params.length>2]' }] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },
    {
      code: 'function foo(bar, baz, qux) {}',
      options: [
        {
          selector: 'FunctionDeclaration[params.length>2]',
          message: 'custom error message.',
        },
      ] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },

    // Upstream: regex / optional / using / await using / source path
    {
      code: 'console.log(/a/i);',
      options: ['Literal[regex.flags=/./]'] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },
    {
      code: 'var foo = foo?.bar?.();',
      options: ['ChainExpression'] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },
    {
      code: 'var foo = foo?.bar?.();',
      options: ['[optional=true]'] as any,
      errors: [
        { messageId: 'restrictedSyntax' },
        { messageId: 'restrictedSyntax' },
      ],
    },
    {
      code: '{ using x = foo(); }',
      options: [`VariableDeclaration[kind='using']`] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },
    {
      code: 'async function f() { await using x = foo(); }',
      options: [`VariableDeclaration[kind='await using']`] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },
    {
      code: `import values from 'some/path';`,
      options: [`ImportDeclaration[source.value=/^some\\/path$/]`] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },
    {
      code: 'foo + bar + baz',
      options: [
        `:is(Identifier[name='foo'], Identifier[name='bar'], Identifier[name='baz'])`,
      ] as any,
      errors: [
        { messageId: 'restrictedSyntax' },
        { messageId: 'restrictedSyntax' },
        { messageId: 'restrictedSyntax' },
      ],
    },

    // Upstream: :nth-child + a?.b — locks in esquery#110 behaviour
    {
      code: 'a?.b',
      options: [':nth-child(1)'] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },

    // Real-world: forbid `with`
    {
      code: 'with (x) { y; }',
      options: ['WithStatement'] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },

    // Real-world: forbid console.* via member access
    {
      code: `console.log('msg');`,
      options: [`CallExpression[callee.object.name='console']`] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },

    // Real-world: forbid setTimeout via callee.name
    {
      code: 'setTimeout(fn, 0);',
      options: [`CallExpression[callee.name='setTimeout']`] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },

    // Real-world: forbid for-in
    {
      code: 'for (const k in obj) {}',
      options: ['ForInStatement'] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },

    // Real-world: forbid default exports
    {
      code: 'export default 1;',
      options: ['ExportDefaultDeclaration'] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },

    // Real-world: ban async / generator functions
    {
      code: 'async function f() {}',
      options: ['FunctionDeclaration[async=true]'] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },
    {
      code: 'function* g() {}',
      options: ['FunctionDeclaration[generator=true]'] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },

    // Real-world: ban arrow with block body (require concise)
    {
      code: 'const f = () => { return 5; };',
      options: [`ArrowFunctionExpression[body.type='BlockStatement']`] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },

    // Real-world: forbid `var` (allow let/const)
    {
      code: 'var a = 1;',
      options: [`VariableDeclaration[kind='var']`] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },

    // Real-world: ban any `new Promise(...)`
    {
      code: 'new Promise(function (res, rej) { res(); });',
      options: [`NewExpression[callee.name='Promise']`] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },

    // Boundary: descendant combinator (whitespace)
    {
      code: 'function f() { return function g() {}; }',
      options: ['FunctionDeclaration FunctionExpression'] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },

    // Boundary: adjacent sibling combinator
    {
      code: 'f(1, 2);',
      options: ['Literal + Literal'] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },

    // Boundary: regex with case-insensitive flag
    {
      code: 'import foo from "Lodash";',
      options: ['ImportDeclaration[source.value=/lodash/i]'] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },

    // Boundary: unary vs update — ++x is UpdateExpression
    {
      code: '++x;',
      options: ['UpdateExpression'] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },
    {
      code: 'typeof x;',
      options: ['UnaryExpression'] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },

    // Boundary: AssignmentExpression on +=
    {
      code: 'a += b;',
      options: ['AssignmentExpression'] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },

    // Boundary: SequenceExpression on comma
    {
      code: 'var x = (a, b);',
      options: ['SequenceExpression'] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },

    // Boundary: Property covers all five tsgo kinds. Five matches.
    {
      code: '({ a, b: 1, get c() {}, set d(v) {}, e() {} });',
      options: ['Property'] as any,
      errors: [
        { messageId: 'restrictedSyntax' },
        { messageId: 'restrictedSyntax' },
        { messageId: 'restrictedSyntax' },
        { messageId: 'restrictedSyntax' },
        { messageId: 'restrictedSyntax' },
      ],
    },

    // Boundary: Mixed string + object option, two distinct messages
    {
      code: 'var a = 1; with (x) {}',
      options: [
        'WithStatement',
        {
          selector: 'VariableDeclaration',
          message: 'Use let or const, not var.',
        },
      ] as any,
      errors: [
        { messageId: 'restrictedSyntax' },
        { messageId: 'restrictedSyntax' },
      ],
    },

    // Boundary: VariableDeclarator (the inner declarator) — different from
    // the wrapping VariableDeclaration / VariableStatement.
    {
      code: 'var a = 1, b = 2;',
      options: ['VariableDeclarator'] as any,
      errors: [
        { messageId: 'restrictedSyntax' },
        { messageId: 'restrictedSyntax' },
      ],
    },

    // Real-world: see-through paren / non-null / as on the receiver
    {
      code: `(console).log('msg');`,
      options: [`MemberExpression[object.name='console']`] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },
    {
      code: `console!.log('msg');`,
      options: [`MemberExpression[object.name='console']`] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },
    {
      code: `(console as any).log('msg');`,
      options: [`MemberExpression[object.name='console']`] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },

    // Real-world: typeof / void / delete operator on dedicated kinds
    {
      code: `typeof x;`,
      options: [`UnaryExpression[operator='typeof']`] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },
    {
      code: `delete a.b;`,
      options: [`UnaryExpression[operator='delete']`] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },

    // Real-world: ban specific imported binding
    {
      code: `import { useEffect } from 'react';`,
      options: [`ImportSpecifier[imported.name='useEffect']`] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },

    // Real-world: ban `__proto__`, `hasOwnProperty`, `arguments.callee`
    {
      code: `obj.__proto__;`,
      options: [`MemberExpression[property.name='__proto__']`] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },
    {
      code: `obj.hasOwnProperty(k);`,
      options: [`CallExpression[callee.property.name='hasOwnProperty']`] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },
    {
      code: `arguments.callee;`,
      options: [
        `MemberExpression[object.name='arguments'][property.name='callee']`,
      ] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },

    // Real-world: ban `==`
    {
      code: `if (a == b) {}`,
      options: [`BinaryExpression[operator='==']`] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },

    // Real-world: regex on Identifier name
    {
      code: `var _hidden = 1;`,
      options: [`Identifier[name=/^_/]`] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },

    // Real-world: ban `eval`
    {
      code: `eval(userInput);`,
      options: [`CallExpression[callee.name='eval']`] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },

    // Real-world: ban `instanceof`
    {
      code: `x instanceof Foo;`,
      options: [`BinaryExpression[operator='instanceof']`] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },

    // Real-world: ban `for-in`
    {
      code: `for (var i in obj) {}`,
      options: ['ForInStatement'] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },

    // Real-world: ban labeled statements
    {
      code: `outer: for (;;) { break outer; }`,
      options: ['LabeledStatement'] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },

    // Real-world: ban Symbol.iterator access
    {
      code: `obj[Symbol.iterator];`,
      options: [
        `MemberExpression[property.object.name='Symbol'][property.property.name='iterator']`,
      ] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },

    // Real-world: ban specific call by name (mocha .skip / .only)
    {
      code: `it.skip('does nothing', () => {});`,
      options: [`CallExpression[callee.property.name='skip']`] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },

    // Real-world: ban setTimeout/setInterval via regex callee
    {
      code: `setTimeout(fn);`,
      options: [
        `CallExpression[callee.name=/^(setTimeout|setInterval)$/]`,
      ] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },

    // Real-world: ban arrow with no params (encourage at-least-one-param)
    {
      code: `const f = () => 1;`,
      options: [`ArrowFunctionExpression[params.length=0]`] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },

    // Real-world: ban export default of function/class — uses tsgo's
    // modifier-based representation (no ExportDefaultDeclaration wrapper).
    {
      code: `export default function f() {}`,
      options: ['ExportDefaultDeclaration > FunctionDeclaration'] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },

    // Real-world: ban computed object keys
    {
      code: `({ ['foo']: 1 });`,
      options: ['Property[computed=true]'] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },

    // Real-world: ban any private class fields
    {
      code: `class A { #x = 1; }`,
      options: ['PropertyDefinition > PrivateIdentifier.key'] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },

    // Real-world: ban catch with binding
    {
      code: `try { f(); } catch (e) { g(); }`,
      options: ['CatchClause[param]'] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },

    // Real-world: BigInt usage
    {
      code: `const n = 1n;`,
      options: ['Literal[value=/n$/]'] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },

    // Real-world: ban specific class name via attribute path
    {
      code: `class Component {}`,
      options: [`ClassDeclaration[id.name='Component']`] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },

    // Real-world: ban template literal with expressions
    {
      code: 'const s = `hi ${name}`;',
      options: ['TemplateLiteral'] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },

    // Real-world: complex `:is(:not())` — `var` declarations only
    {
      code: 'var a = 1;',
      options: [
        ":is(VariableDeclaration, FunctionDeclaration):not([kind='let'])",
      ] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },

    // Round 2 — bot review + self-audit fixes via the CLI/IPC path
    {
      code: 'function f() { return new.target; }',
      options: [`MetaProperty[meta.name='new']`] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },
    {
      code: 'const u = import.meta.url;',
      options: [`MetaProperty[property.name='meta']`] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },
    // Class method is MethodDefinition (not Property)
    {
      code: 'class C { foo() {} }',
      options: ['MethodDefinition'] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },
    // Object-literal method is Property (not MethodDefinition)
    {
      code: '({ foo() {} });',
      options: ['Property'] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },
    // for-in left.type → ESTree VariableDeclaration
    {
      code: 'for (const k in obj) {}',
      options: [`ForInStatement[left.type='VariableDeclaration']`] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },
    // for-of left.kind
    {
      code: 'for (let v of items) {}',
      options: [`ForOfStatement[left.kind='let']`] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },
    // ExportSpecifier local vs exported (tsgo flips the storage)
    {
      code: `export { foo as bar } from 'mod';`,
      options: [`ExportSpecifier[local.name='foo']`] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },
    {
      code: `export { foo as bar } from 'mod';`,
      options: [`ExportSpecifier[exported.name='bar']`] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },

    // Round 3 — ESTree↔tsgo divergence fixes via the CLI/IPC path.

    // MethodDefinition.kind variants
    {
      code: 'class C { foo() {} }',
      options: [`MethodDefinition[kind='method']`] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },
    {
      code: 'class C { constructor() {} }',
      options: [`MethodDefinition[kind='constructor']`] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },
    {
      code: 'class C { get x() { return 1; } }',
      options: [`MethodDefinition[kind='get']`] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },
    {
      code: 'class C { static foo() {} }',
      options: ['MethodDefinition[static=true]'] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },

    // Property.shorthand / Property.method
    {
      code: 'const a = 1; ({ a });',
      options: ['Property[shorthand=true]'] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },
    {
      code: '({ foo() {} });',
      options: ['Property[method=true]'] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },
    {
      code: '({ [x]: 1 });',
      options: ['Property[computed=true]'] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },

    // RestElement / AssignmentPattern on function parameters
    {
      code: 'function f(...args) {}',
      options: ['RestElement'] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },
    {
      code: 'function f(a = 1) {}',
      options: ['AssignmentPattern'] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },

    // ClassDeclaration.superClass
    {
      code: 'class C extends Component {}',
      options: [`ClassDeclaration[superClass.name='Component']`] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },

    // SwitchCase test/consequent
    {
      code: 'switch (x) { default: y; }',
      options: ['SwitchCase[test=null]'] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },
    {
      code: 'switch (x) { case 1: y; break; }',
      options: ['SwitchCase[consequent.length>0]'] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },

    // ForStatement.update
    {
      code: 'for (var i = 0; i < 10; i++) {}',
      options: ['ForStatement[update]'] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },

    // PropertyDefinition.value
    {
      code: 'class C { x = 1; }',
      options: ['PropertyDefinition[value]'] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },
    {
      code: 'class C { onClick = () => {}; }',
      options: [
        `PropertyDefinition[value.type='ArrowFunctionExpression']`,
      ] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },

    // TemplateLiteral.expressions length
    {
      code: 'const s = `a${b}c`;',
      options: ['TemplateLiteral[expressions.length>0]'] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },

    // ExpressionStatement.directive
    {
      code: '"use strict"; foo();',
      options: [`ExpressionStatement[directive='use strict']`] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },

    // Literal.bigint / Literal.regex
    {
      code: 'const n = 1n;',
      options: ['Literal[bigint]'] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },
    {
      code: 'const r = /abc/;',
      options: ['Literal[regex]'] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },

    // YieldExpression.delegate
    {
      code: 'function* g() { yield* x; }',
      options: ['YieldExpression[delegate=true]'] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },

    // CatchClause body & super invocation
    {
      code: 'try { f(); } catch (e) { g(); }',
      options: ['CatchClause > BlockStatement'] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },
    {
      code: 'class C extends B { constructor() { super(); } }',
      options: ['CallExpression > Super'] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },

    // StaticBlock
    {
      code: 'class C { static { initialize(); } }',
      options: ['StaticBlock'] as any,
      errors: [{ messageId: 'restrictedSyntax' }],
    },

    // JSXAttribute name covered in Go tests via Tsx:true; the TS
    // rule-tester always uses .ts so JSX syntax doesn't parse here.
  ],
});

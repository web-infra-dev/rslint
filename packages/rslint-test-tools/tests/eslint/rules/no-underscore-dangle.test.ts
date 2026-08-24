import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-underscore-dangle', {
  valid: [
    'var foo_bar = 1;',
    'function foo_bar() {}',
    'foo.bar.__proto__;',
    'console.log(__filename); console.log(__dirname);',
    "var _ = require('underscore');",
    'var a = b._;',
    'function foo(_bar) {}',
    'function foo(bar_) {}',
    '(function _foo() {})',
    { code: 'function foo(_bar) {}', options: [{}] as any },
    'function foo( _bar = 0) {}',
    'const foo = { onClick(_bar) { } }',
    'const foo = { onClick(_bar = 0) { } }',
    'const foo = (_bar) => {}',
    'const foo = (_bar = 0) => {}',
    'function foo( ..._bar) {}',
    'const foo = (..._bar) => {}',
    'const foo = { onClick(..._bar) { } }',
    'export default function() {}',
    { code: 'var _foo = 1', options: [{ allow: ['_foo'] }] as any },
    { code: 'var __proto__ = 1;', options: [{ allow: ['__proto__'] }] as any },
    { code: 'foo._bar;', options: [{ allow: ['_bar'] }] as any },
    { code: 'function _foo() {}', options: [{ allow: ['_foo'] }] as any },
    { code: 'this._bar;', options: [{ allowAfterThis: true }] as any },
    {
      code: 'class foo { constructor() { super._bar; } }',
      options: [{ allowAfterSuper: true }] as any,
    },
    'class foo { _onClick() { } }',
    'class foo { onClick_() { } }',
    'const o = { _onClick() { } }',
    'const o = { onClick_() { } }',
    {
      code: 'const o = { _onClick() { } }',
      options: [{ allow: ['_onClick'], enforceInMethodNames: true }] as any,
    },
    "const o = { _foo: 'bar' }",
    "const o = { foo_: 'bar' }",
    {
      code: 'this.constructor._bar',
      options: [{ allowAfterThisConstructor: true }] as any,
    },
    'const foo = { onClick(bar) { } }',
    'const foo = (bar) => {}',
    {
      code: 'function foo(_bar) {}',
      options: [{ allowFunctionParams: true }] as any,
    },
    {
      code: 'function foo( _bar = 0) {}',
      options: [{ allowFunctionParams: true }] as any,
    },
    {
      code: 'const foo = { onClick(_bar) { } }',
      options: [{ allowFunctionParams: true }] as any,
    },
    {
      code: 'const foo = (_bar) => {}',
      options: [{ allowFunctionParams: true }] as any,
    },
    {
      code: 'function foo(bar) {}',
      options: [{ allowFunctionParams: false }] as any,
    },
    {
      code: 'const foo = { onClick(bar) { } }',
      options: [{ allowFunctionParams: false }] as any,
    },
    {
      code: 'const foo = (bar) => {}',
      options: [{ allowFunctionParams: false }] as any,
    },
    {
      code: 'function foo(_bar) {}',
      options: [{ allowFunctionParams: false, allow: ['_bar'] }] as any,
    },
    {
      code: 'const foo = { onClick(_bar) { } }',
      options: [{ allowFunctionParams: false, allow: ['_bar'] }] as any,
    },
    {
      code: 'const foo = (_bar) => {}',
      options: [{ allowFunctionParams: false, allow: ['_bar'] }] as any,
    },
    {
      code: 'function foo([_bar]) {}',
      options: [{ allowFunctionParams: false }] as any,
    },
    {
      code: 'function foo([_bar] = []) {}',
      options: [{ allowFunctionParams: false }] as any,
    },
    {
      code: 'function foo( { _bar }) {}',
      options: [{ allowFunctionParams: false }] as any,
    },
    {
      code: 'function foo( { _bar = 0 } = {}) {}',
      options: [{ allowFunctionParams: false }] as any,
    },
    {
      code: 'function foo(...[_bar]) {}',
      options: [{ allowFunctionParams: false }] as any,
    },
    'const [_foo] = arr',
    { code: 'const [_foo] = arr', options: [{}] as any },
    {
      code: 'const [_foo] = arr',
      options: [{ allowInArrayDestructuring: true }] as any,
    },
    {
      code: 'const [foo, ...rest] = [1, 2, 3]',
      options: [{ allowInArrayDestructuring: false }] as any,
    },
    {
      code: 'const [foo, _bar] = [1, 2, 3]',
      options: [{ allowInArrayDestructuring: false, allow: ['_bar'] }] as any,
    },
    'const { _foo } = obj',
    { code: 'const { _foo } = obj', options: [{}] as any },
    {
      code: 'const { _foo } = obj',
      options: [{ allowInObjectDestructuring: true }] as any,
    },
    {
      code: 'const { foo, bar: _bar } = { foo: 1, bar: 2 }',
      options: [{ allowInObjectDestructuring: false, allow: ['_bar'] }] as any,
    },
    {
      code: 'const { foo, _bar } = { foo: 1, _bar: 2 }',
      options: [{ allowInObjectDestructuring: false, allow: ['_bar'] }] as any,
    },
    {
      code: 'const { foo, _bar: bar } = { foo: 1, _bar: 2 }',
      options: [{ allowInObjectDestructuring: false }] as any,
    },
    'class foo { _field; }',
    {
      code: 'class foo { _field; }',
      options: [{ enforceInClassFields: false }] as any,
    },
    'class foo { #_field; }',
    {
      code: 'class foo { #_field; }',
      options: [{ enforceInClassFields: false }] as any,
    },
    { code: 'class foo { _field; }', options: [{}] as any },
    "import foo from 'foo.json' with { _type: 'json' }",
    "export * from 'foo.json' with { _type: 'json' }",
    "export { default } from 'foo.json' with { _type: 'json' }",
    "import('foo.json', { _with: { _type: 'json' } })",
    "import('foo.json', { 'with': { _type: 'json' } })",
    "import('foo.json', { _with: { _type } })",
  ],
  invalid: [
    {
      code: 'var _foo = 1',
      errors: [{ messageId: 'unexpectedUnderscore', line: 1, column: 5 }],
    },
    {
      code: 'var foo_ = 1',
      errors: [{ messageId: 'unexpectedUnderscore', line: 1, column: 5 }],
    },
    {
      code: 'function _foo() {}',
      errors: [{ messageId: 'unexpectedUnderscore', line: 1, column: 1 }],
    },
    {
      code: 'function foo_() {}',
      errors: [{ messageId: 'unexpectedUnderscore', line: 1, column: 1 }],
    },
    {
      code: 'var __proto__ = 1;',
      errors: [{ messageId: 'unexpectedUnderscore', line: 1, column: 5 }],
    },
    {
      code: 'foo._bar;',
      errors: [{ messageId: 'unexpectedUnderscore', line: 1, column: 1 }],
    },
    {
      code: 'this._prop;',
      errors: [{ messageId: 'unexpectedUnderscore', line: 1, column: 1 }],
    },
    {
      code: 'class foo { constructor() { super._prop; } }',
      errors: [{ messageId: 'unexpectedUnderscore', line: 1, column: 29 }],
    },
    {
      code: 'class foo { constructor() { this._prop; } }',
      options: [{ allowAfterSuper: true }] as any,
      errors: [{ messageId: 'unexpectedUnderscore', line: 1, column: 29 }],
    },
    {
      code: 'class foo { _onClick() { } }',
      options: [{ enforceInMethodNames: true }] as any,
      errors: [{ messageId: 'unexpectedUnderscore', line: 1, column: 13 }],
    },
    {
      code: 'class foo { onClick_() { } }',
      options: [{ enforceInMethodNames: true }] as any,
      errors: [{ messageId: 'unexpectedUnderscore', line: 1, column: 13 }],
    },
    {
      code: 'const o = { _onClick() { } }',
      options: [{ enforceInMethodNames: true }] as any,
      errors: [{ messageId: 'unexpectedUnderscore', line: 1, column: 13 }],
    },
    {
      code: 'const o = { onClick_() { } }',
      options: [{ enforceInMethodNames: true }] as any,
      errors: [{ messageId: 'unexpectedUnderscore', line: 1, column: 13 }],
    },
    {
      code: 'this.constructor._bar',
      errors: [{ messageId: 'unexpectedUnderscore', line: 1, column: 1 }],
    },
    {
      code: 'function foo(_bar) {}',
      options: [{ allowFunctionParams: false }] as any,
      errors: [{ messageId: 'unexpectedUnderscore', line: 1, column: 14 }],
    },
    {
      code: '(function foo(_bar) {})',
      options: [{ allowFunctionParams: false }] as any,
      errors: [{ messageId: 'unexpectedUnderscore', line: 1, column: 15 }],
    },
    {
      code: 'function foo(bar, _foo) {}',
      options: [{ allowFunctionParams: false }] as any,
      errors: [{ messageId: 'unexpectedUnderscore', line: 1, column: 19 }],
    },
    {
      code: 'const foo = { onClick(_bar) { } }',
      options: [{ allowFunctionParams: false }] as any,
      errors: [{ messageId: 'unexpectedUnderscore', line: 1, column: 23 }],
    },
    {
      code: 'const foo = (_bar) => {}',
      options: [{ allowFunctionParams: false }] as any,
      errors: [{ messageId: 'unexpectedUnderscore', line: 1, column: 14 }],
    },
    {
      code: 'function foo(_bar = 0) {}',
      options: [{ allowFunctionParams: false }] as any,
      errors: [{ messageId: 'unexpectedUnderscore', line: 1, column: 14 }],
    },
    {
      code: 'const foo = { onClick(_bar = 0) { } }',
      options: [{ allowFunctionParams: false }] as any,
      errors: [{ messageId: 'unexpectedUnderscore', line: 1, column: 23 }],
    },
    {
      code: 'const foo = (_bar = 0) => {}',
      options: [{ allowFunctionParams: false }] as any,
      errors: [{ messageId: 'unexpectedUnderscore', line: 1, column: 14 }],
    },
    {
      code: 'function foo(..._bar) {}',
      options: [{ allowFunctionParams: false }] as any,
      errors: [{ messageId: 'unexpectedUnderscore', line: 1, column: 14 }],
    },
    {
      code: 'const foo = { onClick(..._bar) { } }',
      options: [{ allowFunctionParams: false }] as any,
      errors: [{ messageId: 'unexpectedUnderscore', line: 1, column: 23 }],
    },
    {
      code: 'const foo = (..._bar) => {}',
      options: [{ allowFunctionParams: false }] as any,
      errors: [{ messageId: 'unexpectedUnderscore', line: 1, column: 14 }],
    },
    {
      code: 'const [foo, _bar] = [1, 2]',
      options: [{ allowInArrayDestructuring: false }] as any,
      errors: [{ messageId: 'unexpectedUnderscore', line: 1, column: 7 }],
    },
    {
      code: 'const [_foo = 1] = arr',
      options: [{ allowInArrayDestructuring: false }] as any,
      errors: [{ messageId: 'unexpectedUnderscore', line: 1, column: 7 }],
    },
    {
      code: 'const [foo, ..._rest] = [1, 2, 3]',
      options: [{ allowInArrayDestructuring: false }] as any,
      errors: [{ messageId: 'unexpectedUnderscore', line: 1, column: 7 }],
    },
    {
      code: 'const [foo, [bar_, baz]] = [1, [2, 3]]',
      options: [{ allowInArrayDestructuring: false }] as any,
      errors: [{ messageId: 'unexpectedUnderscore', line: 1, column: 7 }],
    },
    {
      code: 'const { _foo, bar } = { _foo: 1, bar: 2 }',
      options: [{ allowInObjectDestructuring: false }] as any,
      errors: [{ messageId: 'unexpectedUnderscore', line: 1, column: 7 }],
    },
    {
      code: 'const { _foo = 1 } = obj',
      options: [{ allowInObjectDestructuring: false }] as any,
      errors: [{ messageId: 'unexpectedUnderscore', line: 1, column: 7 }],
    },
    {
      code: 'const { bar: _foo = 1 } = obj',
      options: [{ allowInObjectDestructuring: false }] as any,
      errors: [{ messageId: 'unexpectedUnderscore', line: 1, column: 7 }],
    },
    {
      code: 'const { foo: _foo, bar } = { foo: 1, bar: 2 }',
      options: [{ allowInObjectDestructuring: false }] as any,
      errors: [{ messageId: 'unexpectedUnderscore', line: 1, column: 7 }],
    },
    {
      code: 'const { foo, ..._rest} = { foo: 1, bar: 2, baz: 3 }',
      options: [{ allowInObjectDestructuring: false }] as any,
      errors: [{ messageId: 'unexpectedUnderscore', line: 1, column: 7 }],
    },
    {
      code: "const { foo: [_bar, { a: _a, b } ] } = { foo: [1, { a: 'a', b: 'b' }] }",
      options: [
        { allowInArrayDestructuring: false, allowInObjectDestructuring: false },
      ] as any,
      errors: [
        { messageId: 'unexpectedUnderscore', line: 1, column: 7 },
        { messageId: 'unexpectedUnderscore', line: 1, column: 7 },
      ],
    },
    {
      code: "const { foo: [_bar, { a: _a, b } ] } = { foo: [1, { a: 'a', b: 'b' }] }",
      options: [
        { allowInArrayDestructuring: true, allowInObjectDestructuring: false },
      ] as any,
      errors: [{ messageId: 'unexpectedUnderscore', line: 1, column: 7 }],
    },
    {
      code: "const [{ foo: [_bar, _, { bar: _baz }] }] = [{ foo: [1, 2, { bar: 'a' }] }]",
      options: [
        { allowInArrayDestructuring: false, allowInObjectDestructuring: false },
      ] as any,
      errors: [
        { messageId: 'unexpectedUnderscore', line: 1, column: 7 },
        { messageId: 'unexpectedUnderscore', line: 1, column: 7 },
      ],
    },
    {
      code: 'const { foo, bar: { baz, _qux } } = { foo: 1, bar: { baz: 3, _qux: 4 } }',
      options: [{ allowInObjectDestructuring: false }] as any,
      errors: [{ messageId: 'unexpectedUnderscore', line: 1, column: 7 }],
    },
    {
      code: 'class foo { #_bar() {} }',
      options: [{ enforceInMethodNames: true }] as any,
      errors: [{ messageId: 'unexpectedUnderscore', line: 1, column: 13 }],
    },
    {
      code: 'class foo { #bar_() {} }',
      options: [{ enforceInMethodNames: true }] as any,
      errors: [{ messageId: 'unexpectedUnderscore', line: 1, column: 13 }],
    },
    {
      code: 'class foo { _field; }',
      options: [{ enforceInClassFields: true }] as any,
      errors: [{ messageId: 'unexpectedUnderscore', line: 1, column: 13 }],
    },
    {
      code: 'class foo { #_field; }',
      options: [{ enforceInClassFields: true }] as any,
      errors: [{ messageId: 'unexpectedUnderscore', line: 1, column: 13 }],
    },
    {
      code: 'class foo { field_; }',
      options: [{ enforceInClassFields: true }] as any,
      errors: [{ messageId: 'unexpectedUnderscore', line: 1, column: 13 }],
    },
    {
      code: 'class foo { #field_; }',
      options: [{ enforceInClassFields: true }] as any,
      errors: [{ messageId: 'unexpectedUnderscore', line: 1, column: 13 }],
    },
  ],
});

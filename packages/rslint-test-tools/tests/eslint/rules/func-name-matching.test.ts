import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('func-name-matching', {
  valid: [
    'var foo;',
    'var foo = function foo() {};',
    { code: 'var foo = function foo() {};', options: ['always'] as any },
    { code: 'var foo = function bar() {};', options: ['never'] as any },
    'var foo = function() {}',
    { code: 'var foo = () => {}', languageOptions: { ecmaVersion: 6 } },
    'foo = function foo() {};',
    {
      code: 'foo &&= function foo() {};',
      languageOptions: { ecmaVersion: 2021 },
    },
    {
      code: "obj['foo'] ??= function foo() {};",
      languageOptions: { ecmaVersion: 2021 },
    },
    'obj.foo = function foo() {};',
    'obj.bar.foo = function foo() {};',
    "obj['foo'] = function foo() {};",
    'obj[foo] = function bar() {};',
    'var obj = {foo: function foo() {}};',
    "var obj = {'foo': function foo() {}};",
    {
      code: 'var obj = {[foo]: function bar() {}} ',
      languageOptions: { ecmaVersion: 6 },
    },
    'module.exports = function foo(name) {};',
    "module['exports'] = function foo(name) {};",
    {
      code: 'module.exports = function foo(name) {};',
      options: [{ includeCommonJSModuleExports: false }] as any,
      languageOptions: { ecmaVersion: 6 },
    },
    {
      code: "({['foo']: function foo() {}})",
      languageOptions: { ecmaVersion: 6 },
    },
    {
      code: '({[foo]: function bar() {}})',
      languageOptions: { ecmaVersion: 6 },
    },
    { code: '[a] = function foo() {}', languageOptions: { ecmaVersion: 6 } },
    {
      code: '({ value: function value() {} })',
      options: [{ considerPropertyDescriptor: true }] as any,
    },
    {
      code: "Object.defineProperty(foo, 'bar', { value: function bar() {} })",
      options: ['always', { considerPropertyDescriptor: true }] as any,
    },
    {
      code: 'Object.defineProperties(foo, { bar: { value: function bar() {} } })',
      options: ['always', { considerPropertyDescriptor: true }] as any,
    },
    {
      code: 'Object.create(proto, { bar: { value: function bar() {} } })',
      options: ['always', { considerPropertyDescriptor: true }] as any,
    },
    {
      code: "Reflect.defineProperty(foo, 'bar', { value: function bar() {} })",
      options: ['always', { considerPropertyDescriptor: true }] as any,
    },
    {
      code: 'class C { x = function x() {}; }',
      options: ['always'] as any,
      languageOptions: { ecmaVersion: 2022 },
    },
    {
      code: 'class C { #x = function x() {}; }',
      options: ['always'] as any,
      languageOptions: { ecmaVersion: 2022 },
    },
    {
      code: 'class C { #x; foo() { this.#x = function x() {}; } }',
      options: ['always'] as any,
      languageOptions: { ecmaVersion: 2022 },
    },
  ],
  invalid: [
    {
      code: 'let foo = function bar() {};',
      options: ['always'] as any,
      languageOptions: { ecmaVersion: 6 },
      errors: [{ messageId: 'matchVariable' }],
    },
    {
      code: 'foo = function bar() {};',
      languageOptions: { ecmaVersion: 6 },
      errors: [{ messageId: 'matchVariable' }],
    },
    {
      code: 'foo &&= function bar() {};',
      languageOptions: { ecmaVersion: 2021 },
      errors: [{ messageId: 'matchVariable' }],
    },
    {
      code: "obj['foo'] ??= function bar() {};",
      languageOptions: { ecmaVersion: 2021 },
      errors: [{ messageId: 'matchProperty' }],
    },
    {
      code: 'obj.foo = function bar() {};',
      languageOptions: { ecmaVersion: 6 },
      errors: [{ messageId: 'matchProperty' }],
    },
    {
      code: 'let obj = {foo: function bar() {}};',
      languageOptions: { ecmaVersion: 6 },
      errors: [{ messageId: 'matchProperty' }],
    },
    {
      code: "({['foo']: function bar() {}})",
      languageOptions: { ecmaVersion: 6 },
      errors: [{ messageId: 'matchProperty' }],
    },
    {
      code: 'module.exports = function foo(name) {};',
      options: [{ includeCommonJSModuleExports: true }] as any,
      languageOptions: { ecmaVersion: 6 },
      errors: [{ messageId: 'matchProperty' }],
    },
    {
      code: 'module.exports = function exports(name) {};',
      options: ['never', { includeCommonJSModuleExports: true }] as any,
      languageOptions: { ecmaVersion: 6 },
      errors: [{ messageId: 'notMatchProperty' }],
    },
    {
      code: 'var foo = function foo(name) {};',
      options: ['never'] as any,
      errors: [{ messageId: 'notMatchVariable' }],
    },
    {
      code: 'obj.foo = function foo(name) {};',
      options: ['never'] as any,
      errors: [{ messageId: 'notMatchProperty' }],
    },
    {
      code: "Object.defineProperty(foo, 'bar', { value: function baz() {} })",
      options: ['always', { considerPropertyDescriptor: true }] as any,
      errors: [{ messageId: 'matchProperty' }],
    },
    {
      code: 'Object.defineProperties(foo, { bar: { value: function baz() {} } })',
      options: ['always', { considerPropertyDescriptor: true }] as any,
      errors: [{ messageId: 'matchProperty' }],
    },
    {
      code: 'Object.create(proto, { bar: { value: function baz() {} } })',
      options: ['always', { considerPropertyDescriptor: true }] as any,
      errors: [{ messageId: 'matchProperty' }],
    },
    {
      code: 'var obj = { value: function foo(name) {} }',
      options: ['always', { considerPropertyDescriptor: true }] as any,
      errors: [{ messageId: 'matchProperty' }],
    },
    {
      code: "Reflect.defineProperty(foo, 'bar', { value: function baz() {} })",
      options: ['always', { considerPropertyDescriptor: true }] as any,
      errors: [{ messageId: 'matchProperty' }],
    },
    {
      code: '(obj?.aaa).foo = function bar() {};',
      languageOptions: { ecmaVersion: 2020 },
      errors: [{ messageId: 'matchProperty' }],
    },
    {
      code: "Object?.defineProperty(foo, 'bar', { value: function baz() {} })",
      options: ['always', { considerPropertyDescriptor: true }] as any,
      languageOptions: { ecmaVersion: 2020 },
      errors: [{ messageId: 'matchProperty' }],
    },
    {
      code: 'class C { x = function y() {}; }',
      options: ['always'] as any,
      languageOptions: { ecmaVersion: 2022 },
      errors: [{ messageId: 'matchProperty' }],
    },
    {
      code: 'class C { static x = function y() {}; }',
      options: ['always'] as any,
      languageOptions: { ecmaVersion: 2022 },
      errors: [{ messageId: 'matchProperty' }],
    },
  ],
});

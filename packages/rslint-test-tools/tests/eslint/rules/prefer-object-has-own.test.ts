import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('prefer-object-has-own', {
  valid: [
    'Object',
    'Object(obj, prop)',
    'Object.hasOwnProperty',
    'Object.hasOwnProperty(prop)',
    'hasOwnProperty(obj, prop)',
    'foo.hasOwnProperty(prop)',
    'foo.hasOwnProperty(obj, prop)',
    'Object.hasOwnProperty.call',
    'foo.Object.hasOwnProperty.call(obj, prop)',
    'foo.hasOwnProperty.call(obj, prop)',
    'foo.call(Object.prototype.hasOwnProperty, Object.prototype.hasOwnProperty.call)',
    'Object.foo.call(obj, prop)',
    'Object.hasOwnProperty.foo(obj, prop)',
    'Object.hasOwnProperty.call.foo(obj, prop)',
    'Object[hasOwnProperty].call(obj, prop)',
    'Object.hasOwnProperty[call](obj, prop)',
    'class C { #hasOwnProperty; foo() { Object.#hasOwnProperty.call(obj, prop) } }',
    'class C { #call; foo() { Object.hasOwnProperty.#call(obj, prop) } }',
    '(Object) => Object.hasOwnProperty.call(obj, prop)',
    'Object.prototype',
    'Object.prototype(obj, prop)',
    'Object.prototype.hasOwnProperty',
    'Object.prototype.hasOwnProperty(obj, prop)',
    'Object.prototype.hasOwnProperty.call',
    'foo.Object.prototype.hasOwnProperty.call(obj, prop)',
    'foo.prototype.hasOwnProperty.call(obj, prop)',
    'Object.foo.hasOwnProperty.call(obj, prop)',
    'Object.prototype.foo.call(obj, prop)',
    'Object.prototype.hasOwnProperty.foo(obj, prop)',
    'Object.prototype.hasOwnProperty.call.foo(obj, prop)',
    'Object.prototype.prototype.hasOwnProperty.call(a, b);',
    'Object.hasOwnProperty.prototype.hasOwnProperty.call(a, b);',
    'Object.prototype[hasOwnProperty].call(obj, prop)',
    'Object.prototype.hasOwnProperty[call](obj, prop)',
    'class C { #hasOwnProperty; foo() { Object.prototype.#hasOwnProperty.call(obj, prop) } }',
    'class C { #call; foo() { Object.prototype.hasOwnProperty.#call(obj, prop) } }',
    'Object[prototype].hasOwnProperty.call(obj, prop)',
    'class C { #prototype; foo() { Object.#prototype.hasOwnProperty.call(obj, prop) } }',
    '(Object) => Object.prototype.hasOwnProperty.call(obj, prop)',
    '({})',
    '({}(obj, prop))',
    '({}.hasOwnProperty)',
    '({}.hasOwnProperty(prop))',
    '({}.hasOwnProperty(obj, prop))',
    '({}.hasOwnProperty.call)',
    '({}).prototype.hasOwnProperty.call(a, b);',
    '({}.foo.call(obj, prop))',
    '({}.hasOwnProperty.foo(obj, prop))',
    '({}[hasOwnProperty].call(obj, prop))',
    '({}.hasOwnProperty[call](obj, prop))',
    '({}).hasOwnProperty[call](object, property)',
    '({})[hasOwnProperty].call(object, property)',
    'class C { #hasOwnProperty; foo() { ({}.#hasOwnProperty.call(obj, prop)) } }',
    'class C { #call; foo() { ({}.hasOwnProperty.#call(obj, prop)) } }',
    '({ foo }.hasOwnProperty.call(obj, prop))',
    '(Object) => ({}).hasOwnProperty.call(obj, prop)',
    '\n        let obj = {};\n        Object.hasOwn(obj,"");\n        ',
    'const hasProperty = Object.hasOwn(object, property);',
    '/* global Object: off */\n        ({}).hasOwnProperty.call(a, b);',
  ],
  invalid: [
    {
      code: "Object.hasOwnProperty.call(obj, 'foo')",
      errors: [{ messageId: 'useHasOwn' }],
    },
    {
      code: 'Object.hasOwnProperty.call(obj, property)',
      errors: [{ messageId: 'useHasOwn' }],
    },
    {
      code: "Object.prototype.hasOwnProperty.call(obj, 'foo')",
      errors: [{ messageId: 'useHasOwn' }],
    },
    {
      code: "({}).hasOwnProperty.call(obj, 'foo')",
      errors: [{ messageId: 'useHasOwn' }],
    },
    {
      code: 'Object/* comment */.prototype.hasOwnProperty.call(a, b);',
      errors: [{ messageId: 'useHasOwn' }],
    },
    {
      code: 'const hasProperty = Object.prototype.hasOwnProperty.call(object, property);',
      errors: [{ messageId: 'useHasOwn' }],
    },
    {
      code: 'const hasProperty = (( Object.prototype.hasOwnProperty.call(object, property) ));',
      errors: [{ messageId: 'useHasOwn' }],
    },
    {
      code: 'const hasProperty = (( Object.prototype.hasOwnProperty.call ))(object, property);',
      errors: [{ messageId: 'useHasOwn' }],
    },
    {
      code: 'const hasProperty = (( Object.prototype.hasOwnProperty )).call(object, property);',
      errors: [{ messageId: 'useHasOwn' }],
    },
    {
      code: 'const hasProperty = (( Object.prototype )).hasOwnProperty.call(object, property);',
      errors: [{ messageId: 'useHasOwn' }],
    },
    {
      code: 'const hasProperty = (( Object )).prototype.hasOwnProperty.call(object, property);',
      errors: [{ messageId: 'useHasOwn' }],
    },
    {
      code: 'const hasProperty = {}.hasOwnProperty.call(object, property);',
      errors: [{ messageId: 'useHasOwn' }],
    },
    {
      code: 'const hasProperty={}.hasOwnProperty.call(object, property);',
      errors: [{ messageId: 'useHasOwn' }],
    },
    {
      code: 'const hasProperty = (( {}.hasOwnProperty.call(object, property) ));',
      errors: [{ messageId: 'useHasOwn' }],
    },
    {
      code: 'const hasProperty = (( {}.hasOwnProperty.call ))(object, property);',
      errors: [{ messageId: 'useHasOwn' }],
    },
    {
      code: 'const hasProperty = (( {}.hasOwnProperty )).call(object, property);',
      errors: [{ messageId: 'useHasOwn' }],
    },
    {
      code: 'const hasProperty = (( {} )).hasOwnProperty.call(object, property);',
      errors: [{ messageId: 'useHasOwn' }],
    },
    {
      code: 'function foo(){return {}.hasOwnProperty.call(object, property)}',
      errors: [{ messageId: 'useHasOwn' }],
    },
    {
      code: 'function foo(){return{}.hasOwnProperty.call(object, property)}',
      errors: [{ messageId: 'useHasOwn' }],
    },
    {
      code: 'function foo(){return/*comment*/{}.hasOwnProperty.call(object, property)}',
      errors: [{ messageId: 'useHasOwn' }],
    },
    {
      code: 'async function foo(){return await{}.hasOwnProperty.call(object, property)}',
      errors: [{ messageId: 'useHasOwn' }],
    },
    {
      code: 'async function foo(){return await/*comment*/{}.hasOwnProperty.call(object, property)}',
      errors: [{ messageId: 'useHasOwn' }],
    },
    {
      code: 'for (const x of{}.hasOwnProperty.call(object, property).toString());',
      errors: [{ messageId: 'useHasOwn' }],
    },
    {
      code: 'for (const x of/*comment*/{}.hasOwnProperty.call(object, property).toString());',
      errors: [{ messageId: 'useHasOwn' }],
    },
    {
      code: 'for (const x in{}.hasOwnProperty.call(object, property).toString());',
      errors: [{ messageId: 'useHasOwn' }],
    },
    {
      code: 'for (const x in/*comment*/{}.hasOwnProperty.call(object, property).toString());',
      errors: [{ messageId: 'useHasOwn' }],
    },
    {
      code: 'function foo(){return({}.hasOwnProperty.call)(object, property)}',
      errors: [{ messageId: 'useHasOwn' }],
    },
    {
      code: "Object['prototype']['hasOwnProperty']['call'](object, property);",
      errors: [{ messageId: 'useHasOwn' }],
    },
    {
      code: 'Object[`prototype`][`hasOwnProperty`][`call`](object, property);',
      errors: [{ messageId: 'useHasOwn' }],
    },
    {
      code: "Object['hasOwnProperty']['call'](object, property);",
      errors: [{ messageId: 'useHasOwn' }],
    },
    {
      code: 'Object[`hasOwnProperty`][`call`](object, property);',
      errors: [{ messageId: 'useHasOwn' }],
    },
    {
      code: "({})['hasOwnProperty']['call'](object, property);",
      errors: [{ messageId: 'useHasOwn' }],
    },
    {
      code: '({})[`hasOwnProperty`][`call`](object, property);',
      errors: [{ messageId: 'useHasOwn' }],
    },
  ],
});

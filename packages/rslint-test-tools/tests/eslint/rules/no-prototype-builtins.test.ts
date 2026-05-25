import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-prototype-builtins', {
  valid: [
    "Object.prototype.hasOwnProperty.call(foo, 'bar')",
    "Object.prototype.isPrototypeOf.call(foo, 'bar')",
    "Object.prototype.propertyIsEnumerable.call(foo, 'bar')",
    "Object.prototype.hasOwnProperty.apply(foo, ['bar'])",
    "Object.prototype.isPrototypeOf.apply(foo, ['bar'])",
    "Object.prototype.propertyIsEnumerable.apply(foo, ['bar'])",
    'foo.hasOwnProperty',
    'foo.hasOwnProperty.bar()',
    'foo(hasOwnProperty)',
    "hasOwnProperty(foo, 'bar')",
    "isPrototypeOf(foo, 'bar')",
    "propertyIsEnumerable(foo, 'bar')",
    "({}.hasOwnProperty.call(foo, 'bar'))",
    "({}.isPrototypeOf.call(foo, 'bar'))",
    "({}.propertyIsEnumerable.call(foo, 'bar'))",
    "({}.hasOwnProperty.apply(foo, ['bar']))",
    "({}.isPrototypeOf.apply(foo, ['bar']))",
    "({}.propertyIsEnumerable.apply(foo, ['bar']))",
    "foo[hasOwnProperty]('bar')",
    "foo['HasOwnProperty']('bar')",
    "foo[`isPrototypeOff`]('bar')",
    "foo?.['propertyIsEnumerabl']('bar')",
    "foo[1]('bar')",
    "foo[null]('bar')",
    "class C { #hasOwnProperty; foo() { obj.#hasOwnProperty('bar'); } }",
    "foo['hasOwn' + 'Property']('bar')",
    "foo[`hasOwnProperty${''}`]('bar')",
  ],
  invalid: [
    {
      code: "foo.hasOwnProperty('bar')",
      errors: [
        {
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 19,
          messageId: 'prototypeBuildIn',
        },
      ],
    },
    {
      code: "foo.isPrototypeOf('bar')",
      errors: [
        {
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 18,
          messageId: 'prototypeBuildIn',
        },
      ],
    },
    {
      code: "foo.propertyIsEnumerable('bar')",
      errors: [
        {
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 25,
          messageId: 'prototypeBuildIn',
        },
      ],
    },
    {
      code: "foo.bar.hasOwnProperty('bar')",
      errors: [
        {
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 23,
          messageId: 'prototypeBuildIn',
        },
      ],
    },
    {
      code: "foo.bar.baz.isPrototypeOf('bar')",
      errors: [
        {
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 26,
          messageId: 'prototypeBuildIn',
        },
      ],
    },
    {
      code: "foo['hasOwnProperty']('bar')",
      errors: [
        {
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 21,
          messageId: 'prototypeBuildIn',
        },
      ],
    },
    {
      code: "foo[`isPrototypeOf`]('bar').baz",
      errors: [
        {
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 20,
          messageId: 'prototypeBuildIn',
        },
      ],
    },
    {
      code: `foo.bar["propertyIsEnumerable"]('baz')`,
      errors: [
        {
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 31,
          messageId: 'prototypeBuildIn',
        },
      ],
    },
    {
      code: "(function(Object) {return foo.hasOwnProperty('bar');})",
      errors: [{ messageId: 'prototypeBuildIn' }],
    },
    // Optional chaining
    {
      code: "foo?.hasOwnProperty('bar')",
      errors: [{ messageId: 'prototypeBuildIn' }],
    },
    {
      code: "foo?.bar.hasOwnProperty('baz')",
      errors: [{ messageId: 'prototypeBuildIn' }],
    },
    {
      code: "foo.hasOwnProperty?.('bar')",
      errors: [{ messageId: 'prototypeBuildIn' }],
    },
    {
      code: "foo?.hasOwnProperty('bar').baz",
      errors: [{ messageId: 'prototypeBuildIn' }],
    },
    {
      code: "foo.hasOwnProperty('bar')?.baz",
      errors: [{ messageId: 'prototypeBuildIn' }],
    },
    {
      code: "(a,b).hasOwnProperty('bar')",
      errors: [{ messageId: 'prototypeBuildIn' }],
    },
    {
      code: "(foo?.hasOwnProperty)('bar')",
      errors: [{ messageId: 'prototypeBuildIn' }],
    },
    {
      code: "(foo?.hasOwnProperty)?.('bar')",
      errors: [{ messageId: 'prototypeBuildIn' }],
    },
    {
      code: "foo?.['hasOwnProperty']('bar')",
      errors: [{ messageId: 'prototypeBuildIn' }],
    },
    {
      code: "(foo?.[`hasOwnProperty`])('bar')",
      errors: [{ messageId: 'prototypeBuildIn' }],
    },
  ],
});

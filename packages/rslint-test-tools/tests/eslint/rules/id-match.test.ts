import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('id-match', {
  valid: [
    {
      code: '__foo = "Matthieu"',
      options: ['^[a-z]+$', { onlyDeclarations: true }] as any,
    },
    {
      code: 'firstname = "Matthieu"',
      options: ['^[a-z]+$'] as any,
    },
    {
      code: 'first_name = "Matthieu"',
      options: ['[a-z]+'] as any,
    },
    {
      code: 'firstname = "Matthieu"',
      options: ['^f'] as any,
    },
    {
      code: 'last_Name = "Larcher"',
      options: ['^[a-z]+(_[A-Z][a-z]+)*$'] as any,
    },
    {
      code: 'param = "none"',
      options: ['^[a-z]+(_[A-Z][a-z])*$'] as any,
    },
    {
      code: 'function noUnder(){}',
      options: ['^[^_]+$'] as any,
    },
    {
      code: 'no_under()',
      options: ['^[^_]+$'] as any,
    },
    {
      code: 'foo.no_under2()',
      options: ['^[^_]+$'] as any,
    },
    {
      code: 'var foo = bar.no_under3;',
      options: ['^[^_]+$'] as any,
    },
    {
      code: 'var foo = bar.no_under4.something;',
      options: ['^[^_]+$'] as any,
    },
    {
      code: 'foo.no_under5.qux = bar.no_under6.something;',
      options: ['^[^_]+$'] as any,
    },
    {
      code: 'if (bar.no_under7) {}',
      options: ['^[^_]+$'] as any,
    },
    {
      code: 'var obj = { key: foo.no_under8 };',
      options: ['^[^_]+$'] as any,
    },
    {
      code: 'var arr = [foo.no_under9];',
      options: ['^[^_]+$'] as any,
    },
    {
      code: '[foo.no_under10]',
      options: ['^[^_]+$'] as any,
    },
    {
      code: 'var arr = [foo.no_under11.qux];',
      options: ['^[^_]+$'] as any,
    },
    {
      code: '[foo.no_under12.nesting]',
      options: ['^[^_]+$'] as any,
    },
    {
      code: 'if (foo.no_under13 === boom.no_under14) { [foo.no_under15] }',
      options: ['^[^_]+$'] as any,
    },
    {
      code: 'var myArray = new Array(); var myDate = new Date();',
      options: ['^[a-z$]+([A-Z][a-z]+)*$'] as any,
    },
    {
      code: 'var x = obj._foo;',
      options: ['^[^_]+$'] as any,
    },
    {
      code: 'var obj = {key: no_under}',
      options: ['^[^_]+$', { properties: true, onlyDeclarations: true }] as any,
    },
    {
      code: 'var {key_no_under: key} = {}',
      options: ['^[^_]+$', { properties: true }] as any,
    },
    {
      code: 'var { category_id } = query;',
      options: [
        '^[^_]+$',
        { properties: true, ignoreDestructuring: true },
      ] as any,
    },
    {
      code: 'var { category_id: category_id } = query;',
      options: [
        '^[^_]+$',
        { properties: true, ignoreDestructuring: true },
      ] as any,
    },
    {
      code: 'var { category_id = 1 } = query;',
      options: [
        '^[^_]+$',
        { properties: true, ignoreDestructuring: true },
      ] as any,
    },
    {
      code: 'var o = {key: 1}',
      options: ['^[^_]+$', { properties: true }] as any,
    },
    {
      code: 'var o = {no_under16: 1}',
      options: ['^[^_]+$', { properties: false }] as any,
    },
    {
      code: 'obj.no_under17 = 2;',
      options: ['^[^_]+$', { properties: false }] as any,
    },
    {
      code: 'var obj = {\n no_under18: 1 \n};\n obj.no_under19 = 2;',
      options: ['^[^_]+$', { properties: false }] as any,
    },
    {
      code: 'obj.no_under20 = function(){};',
      options: ['^[^_]+$', { properties: false }] as any,
    },
    {
      code: 'var x = obj._foo2;',
      options: ['^[^_]+$', { properties: false }] as any,
    },
    {
      code: '\n            const foo = Object.keys(bar);\n            const a = Array.from(b);\n            const bar = () => Array;\n            ',
      options: [
        '^\\$?[a-z]+([A-Z0-9][a-z0-9]+)*$',
        { properties: true },
      ] as any,
    },
    {
      code: '\n            const foo = {\n                foo_one: 1,\n                bar_one: 2,\n                fooBar: 3\n            };\n            ',
      options: ['^[^_]+$', { properties: false }] as any,
    },
    {
      code: '\n            const foo = {\n                foo_one: 1,\n                bar_one: 2,\n                fooBar: 3\n            };\n            ',
      options: ['^[^_]+$', { onlyDeclarations: true }] as any,
    },
    {
      code: '\n            const foo = {\n                foo_one: 1,\n                bar_one: 2,\n                fooBar: 3\n            };\n            ',
      options: [
        '^[^_]+$',
        { properties: false, onlyDeclarations: false },
      ] as any,
    },
    {
      code: '\n            const foo = {\n                [a]: 1,\n            };\n            ',
      options: ['^[^a]', { properties: true, onlyDeclarations: true }] as any,
    },
    {
      code: 'class x { foo() {} }',
      options: ['^[^_]+$'] as any,
    },
    {
      code: 'class x { #foo() {} }',
      options: ['^[^_]+$'] as any,
    },
    {
      code: 'class x { _foo = 1; }',
      options: ['^[^_]+$'] as any,
    },
    {
      code: 'class x { _foo = 1; }',
      options: ['^[^_]+$', { classFields: false }] as any,
    },
    {
      code: 'class x { #_foo = 1; }',
      options: ['^[^_]+$', { classFields: false }] as any,
    },
    {
      code: 'class x { #_foo = 1; }',
      options: ['^[^_]+$'] as any,
    },
    {
      code: 'import.meta',
      options: ['^$'] as any,
    },
    {
      code: 'function foo() { new.target; }',
      options: ['^foo$'] as any,
    },
    {
      code: "import foo from 'foo.json' with { type: 'json' }",
      options: ['^foo', { properties: true }] as any,
    },
    {
      code: "export * from 'foo.json' with { type: 'json' }",
      options: ['^foo', { properties: true }] as any,
    },
    {
      code: "export { default } from 'foo.json' with { type: 'json' }",
      options: ['^def', { properties: true }] as any,
    },
    {
      code: "import('foo.json', { with: { type: 'json' } })",
      options: ['^foo', { properties: true }] as any,
    },
    {
      code: "import('foo.json', { 'with': { type: 'json' } })",
      options: ['^foo', { properties: true }] as any,
    },
    {
      code: "import('foo.json', { with: { type } })",
      options: ['^foo', { properties: true }] as any,
    },
  ],
  invalid: [
    {
      code: 'var __foo = "Matthieu"',
      options: ['^[a-z]+$', { onlyDeclarations: true }] as any,
      errors: [{ messageId: 'notMatch' }],
    },
    {
      code: 'first_name = "Matthieu"',
      options: ['^[a-z]+$'] as any,
      errors: [{ messageId: 'notMatch' }],
    },
    {
      code: 'first_name = "Matthieu"',
      options: ['^z'] as any,
      errors: [{ messageId: 'notMatch' }],
    },
    {
      code: 'Last_Name = "Larcher"',
      options: ['^[a-z]+(_[A-Z][a-z])*$'] as any,
      errors: [{ messageId: 'notMatch' }],
    },
    {
      code: 'var obj = {key: no_under}',
      options: ['^[^_]+$', { properties: true }] as any,
      errors: [{ messageId: 'notMatch' }],
    },
    {
      code: 'function no_under21(){}',
      options: ['^[^_]+$'] as any,
      errors: [{ messageId: 'notMatch' }],
    },
    {
      code: 'obj.no_under22 = function(){};',
      options: ['^[^_]+$', { properties: true }] as any,
      errors: [{ messageId: 'notMatch' }],
    },
    {
      code: 'no_under23.foo = function(){};',
      options: ['^[^_]+$', { properties: true }] as any,
      errors: [{ messageId: 'notMatch' }],
    },
    {
      code: '[no_under24.baz]',
      options: ['^[^_]+$', { properties: true }] as any,
      errors: [{ messageId: 'notMatch' }],
    },
    {
      code: 'if (foo.bar_baz === boom.bam_pow) { [no_under25.baz] }',
      options: ['^[^_]+$', { properties: true }] as any,
      errors: [{ messageId: 'notMatch' }],
    },
    {
      code: 'foo.no_under26 = boom.bam_pow',
      options: ['^[^_]+$', { properties: true }] as any,
      errors: [{ messageId: 'notMatch' }],
    },
    {
      code: 'var foo = { no_under27: boom.bam_pow }',
      options: ['^[^_]+$', { properties: true }] as any,
      errors: [{ messageId: 'notMatch' }],
    },
    {
      code: 'foo.qux.no_under28 = { bar: boom.bam_pow }',
      options: ['^[^_]+$', { properties: true }] as any,
      errors: [{ messageId: 'notMatch' }],
    },
    {
      code: 'var o = {no_under29: 1}',
      options: ['^[^_]+$', { properties: true }] as any,
      errors: [{ messageId: 'notMatch' }],
    },
    {
      code: 'obj.no_under30 = 2;',
      options: ['^[^_]+$', { properties: true }] as any,
      errors: [{ messageId: 'notMatch' }],
    },
    {
      code: 'var { category_id: category_alias } = query;',
      options: ['^[^_]+$', { properties: true }] as any,
      errors: [{ messageId: 'notMatch' }],
    },
    {
      code: 'var { category_id: category_alias } = query;',
      options: [
        '^[^_]+$',
        { properties: true, ignoreDestructuring: true },
      ] as any,
      errors: [{ messageId: 'notMatch' }],
    },
    {
      code: 'var { category_id: categoryId, ...other_props } = query;',
      options: [
        '^[^_]+$',
        { properties: true, ignoreDestructuring: true },
      ] as any,
      errors: [{ messageId: 'notMatch' }],
    },
    {
      code: 'var { category_id } = query;',
      options: ['^[^_]+$', { properties: true }] as any,
      errors: [{ messageId: 'notMatch' }],
    },
    {
      code: 'var { category_id = 1 } = query;',
      options: ['^[^_]+$', { properties: true }] as any,
      errors: [{ messageId: 'notMatch' }],
    },
    {
      code: 'import no_camelcased from "external-module";',
      options: ['^[^_]+$', { properties: true }] as any,
      errors: [{ messageId: 'notMatch' }],
    },
    {
      code: 'import * as no_camelcased from "external-module";',
      options: ['^[^_]+$', { properties: true }] as any,
      errors: [{ messageId: 'notMatch' }],
    },
    {
      code: 'export * as no_camelcased from "external-module";',
      options: ['^[^_]+$'] as any,
      errors: [{ messageId: 'notMatch' }],
    },
    {
      code: 'import { no_camelcased } from "external-module";',
      options: ['^[^_]+$', { properties: true }] as any,
      errors: [{ messageId: 'notMatch' }],
    },
    {
      code: 'import { no_camelcased as no_camel_cased } from "external module";',
      options: ['^[^_]+$', { properties: true }] as any,
      errors: [{ messageId: 'notMatch' }],
    },
    {
      code: 'import { camelCased as no_camel_cased } from "external module";',
      options: ['^[^_]+$', { properties: true }] as any,
      errors: [{ messageId: 'notMatch' }],
    },
    {
      code: 'import { camelCased, no_camelcased } from "external-module";',
      options: ['^[^_]+$', { properties: true }] as any,
      errors: [{ messageId: 'notMatch' }],
    },
    {
      code: 'import { no_camelcased as camelCased, another_no_camelcased } from "external-module";',
      options: ['^[^_]+$', { properties: true }] as any,
      errors: [{ messageId: 'notMatch' }],
    },
    {
      code: 'import camelCased, { no_camelcased } from "external-module";',
      options: ['^[^_]+$', { properties: true }] as any,
      errors: [{ messageId: 'notMatch' }],
    },
    {
      code: 'import no_camelcased, { another_no_camelcased as camelCased } from "external-module";',
      options: ['^[^_]+$', { properties: true }] as any,
      errors: [{ messageId: 'notMatch' }],
    },
    {
      code: 'function foo({ no_camelcased }) {};',
      options: ['^[^_]+$', { properties: true }] as any,
      errors: [{ messageId: 'notMatch' }],
    },
    {
      code: "function foo({ no_camelcased = 'default value' }) {};",
      options: ['^[^_]+$', { properties: true }] as any,
      errors: [{ messageId: 'notMatch' }],
    },
    {
      code: 'const no_camelcased = 0; function foo({ camelcased_value = no_camelcased }) {}',
      options: ['^[^_]+$', { properties: true }] as any,
      errors: [{ messageId: 'notMatch' }, { messageId: 'notMatch' }],
    },
    {
      code: 'const { bar: no_camelcased } = foo;',
      options: ['^[^_]+$', { properties: true }] as any,
      errors: [{ messageId: 'notMatch' }],
    },
    {
      code: 'function foo({ value_1: my_default }) {}',
      options: ['^[^_]+$', { properties: true }] as any,
      errors: [{ messageId: 'notMatch' }],
    },
    {
      code: 'function foo({ isCamelcased: no_camelcased }) {};',
      options: ['^[^_]+$', { properties: true }] as any,
      errors: [{ messageId: 'notMatch' }],
    },
    {
      code: 'var { foo: bar_baz = 1 } = quz;',
      options: ['^[^_]+$', { properties: true }] as any,
      errors: [{ messageId: 'notMatch' }],
    },
    {
      code: 'const { no_camelcased = false } = bar;',
      options: ['^[^_]+$', { properties: true }] as any,
      errors: [{ messageId: 'notMatch' }],
    },
    {
      code: '\n            const foo_variable = 1;\n            class MyClass {\n            }\n            let a = new MyClass();\n            let b = {id: 1};\n            let c = Object.keys(b);\n            let d = Array.from(b);\n            let e = (Object) => Object.keys(obj, prop); // not global Object\n            let f = (Array) => Array.from(obj, prop); // not global Array\n            foo.Array = 5; // not global Array\n            ',
      options: [
        '^\\$?[a-z]+([A-Z0-9][a-z0-9]+)*$',
        { properties: true },
      ] as any,
      errors: [
        { messageId: 'notMatch' },
        { messageId: 'notMatch' },
        { messageId: 'notMatch' },
        { messageId: 'notMatch' },
        { messageId: 'notMatch' },
        { messageId: 'notMatch' },
        { messageId: 'notMatch' },
      ],
    },
    {
      code: 'class x { _foo() {} }',
      options: ['^[^_]+$'] as any,
      errors: [{ messageId: 'notMatch' }],
    },
    {
      code: 'class x { #_foo() {} }',
      options: ['^[^_]+$'] as any,
      errors: [{ messageId: 'notMatchPrivate' }],
    },
    {
      code: 'class x { _foo = 1; }',
      options: ['^[^_]+$', { classFields: true }] as any,
      errors: [{ messageId: 'notMatch' }],
    },
    {
      code: 'class x { #_foo = 1; }',
      options: ['^[^_]+$', { classFields: true }] as any,
      errors: [{ messageId: 'notMatchPrivate' }],
    },
    {
      code: '\n            const foo = {\n                foo_one: 1,\n                bar_one: 2,\n                fooBar: 3\n            };\n            ',
      options: ['^[^_]+$', { properties: true, onlyDeclarations: true }] as any,
      errors: [{ messageId: 'notMatch' }, { messageId: 'notMatch' }],
    },
    {
      code: '\n            const foo = {\n                foo_one: 1,\n                bar_one: 2,\n                fooBar: 3\n            };\n            ',
      options: [
        '^[^_]+$',
        { properties: true, onlyDeclarations: false },
      ] as any,
      errors: [{ messageId: 'notMatch' }, { messageId: 'notMatch' }],
    },
    {
      code: '\n            const foo = {\n                [a]: 1,\n            };\n            ',
      options: ['^[^a]', { properties: true, onlyDeclarations: false }] as any,
      errors: [{ messageId: 'notMatch' }],
    },
    {
      code: '\n            const foo = {\n                [a]: 1,\n            };\n            ',
      options: ['^[^a]', { properties: false, onlyDeclarations: false }] as any,
      errors: [{ messageId: 'notMatch' }],
    },
    {
      code: "import('foo.json', { with: { [type]: 'json' } })",
      options: ['^foo', { properties: true }] as any,
      errors: [{ messageId: 'notMatch' }],
    },
    {
      code: "import('foo.json', { with: { type: json } })",
      options: ['^foo', { properties: true }] as any,
      errors: [{ messageId: 'notMatch' }],
    },
  ],
});

import { RuleTester } from '@typescript-eslint/rule-tester';
import { getFixturesRootDir } from '../RuleTester.ts';

const rootPath = getFixturesRootDir();
const ruleTester = new RuleTester({
  // @ts-ignore
  languageOptions: {
    parserOptions: {
      project: './tsconfig.json',
      tsconfigRootDir: rootPath,
    },
  },
});

ruleTester.run('no-unused-vars', {
  valid: [
    // Variables that are used
    {
      code: 'const foo = 5; console.log(foo);',
    },
    {
      code: 'let foo = 5; foo = 10; console.log(foo);',
    },
    {
      code: 'const foo = 5; function bar() { return foo; }',
    },

    // Function declarations that are used
    {
      code: 'function foo() {} foo();',
    },
    {
      code: 'function foo() {} export { foo };',
    },

    // Parameters that are used
    {
      code: 'function foo(bar) { console.log(bar); }',
    },
    {
      code: '(function(foo) { console.log(foo); })',
    },
    {
      code: 'function foo(first, second) { console.log(second); }',
    },

    // Catch clause variables that are used
    {
      code: 'try {} catch (e) { console.log(e); }',
    },

    // Rest siblings
    {
      code: 'const { foo, ...rest } = { foo: 1, bar: 2 }; console.log(rest);',
      options: [{ ignoreRestSiblings: true }],
    },

    // vars: local
    {
      code: 'const foo = 1;',
      options: [{ vars: 'local' }],
    },

    // args: none
    {
      code: 'function foo(bar) {}',
      options: [{ args: 'none' }],
    },

    // args: after-used
    {
      code: 'function foo(first, second) { console.log(first); }',
      options: [{ args: 'after-used' }],
    },

    // caughtErrors: none
    {
      code: 'try {} catch (e) {}',
      options: [{ caughtErrors: 'none' }],
    },

    // Ignore patterns
    {
      code: 'const _foo = 1;',
      options: [{ varsIgnorePattern: '^_' }],
    },
    {
      code: 'function foo(_bar) {}',
      options: [{ argsIgnorePattern: '^_' }],
    },
    {
      code: 'try {} catch (_e) {}',
      options: [{ caughtErrorsIgnorePattern: '^_' }],
    },
    {
      code: 'const [_foo] = [1];',
      options: [{ destructuredArrayIgnorePattern: '^_' }],
    },

    // Class with static init block
    {
      code: 'class Foo { static {} } export { Foo };',
      options: [{ ignoreClassWithStaticInitBlock: true }],
    },

    // Type-only imports
    {
      code: 'import { type Foo } from "./foo"; const bar: Foo = {};',
    },
    {
      code: 'import type { Foo } from "./foo"; const bar: Foo = {};',
    },

    // Exports
    {
      code: 'export const foo = 1;',
    },
    {
      code: 'const foo = 1; export { foo };',
    },
    {
      code: 'export function foo() {}',
    },
    {
      code: 'export class Foo {}',
    },

    // Used in type positions
    {
      code: 'const foo = 1; type Bar = typeof foo;',
    },
    {
      code: 'function foo() {} type Bar = typeof foo;',
    },

    // Ambient declarations
    {
      code: 'declare const foo: number;',
      options: [{ vars: 'all' }],
    },
    {
      code: 'declare function foo(): void;',
      options: [{ vars: 'all' }],
    },
    {
      code: 'declare namespace foo {}',
      options: [{ vars: 'all' }],
    },
  ],

  invalid: [
    // Unused variables
    {
      code: 'const foo = 5;',
      errors: [
        {
          messageId: 'unusedVar',
          data: { varName: 'foo', action: 'defined', additional: '' },
        },
      ],
    },
    {
      code: 'let foo = 5;',
      errors: [
        {
          messageId: 'unusedVar',
          data: { varName: 'foo', action: 'defined', additional: '' },
        },
      ],
    },
    {
      code: 'let foo = 5; foo = 10;',
      errors: [
        {
          messageId: 'unusedVar',
          data: { varName: 'foo', action: 'assigned a value', additional: '' },
        },
      ],
    },

    // Unused function declarations
    {
      code: 'function foo() {}',
      errors: [
        {
          messageId: 'unusedVar',
          data: { varName: 'foo', action: 'defined', additional: '' },
        },
      ],
    },

    // Unused parameters
    {
      code: 'function foo(bar) {}',
      errors: [
        {
          messageId: 'unusedVar',
          data: { varName: 'bar', action: 'defined', additional: '' },
        },
      ],
    },
    {
      code: 'function foo(bar) {}',
      options: [{ args: 'all' }],
      errors: [
        {
          messageId: 'unusedVar',
          data: { varName: 'bar', action: 'defined', additional: '' },
        },
      ],
    },
    {
      code: 'function foo(first, second) { console.log(second); }',
      options: [{ args: 'after-used' }],
      errors: [
        {
          messageId: 'unusedVar',
          data: { varName: 'first', action: 'defined', additional: '' },
        },
      ],
    },

    // Unused catch clause variables
    {
      code: 'try {} catch (e) {}',
      errors: [
        {
          messageId: 'unusedVar',
          data: { varName: 'e', action: 'defined', additional: '' },
        },
      ],
    },
    {
      code: 'try {} catch (e) {}',
      options: [{ caughtErrors: 'all' }],
      errors: [
        {
          messageId: 'unusedVar',
          data: { varName: 'e', action: 'defined', additional: '' },
        },
      ],
    },

    // Rest siblings without ignoreRestSiblings
    {
      code: 'const { foo, ...rest } = { foo: 1, bar: 2 }; console.log(rest);',
      errors: [
        {
          messageId: 'unusedVar',
          data: { varName: 'foo', action: 'defined', additional: '' },
        },
      ],
    },

    // Class without static init block
    {
      code: 'class Foo {}',
      options: [{ ignoreClassWithStaticInitBlock: true }],
      errors: [
        {
          messageId: 'unusedVar',
          data: { varName: 'Foo', action: 'defined', additional: '' },
        },
      ],
    },

    // Type-only usage for non-imports
    {
      code: 'const foo = 1; type Bar = typeof foo;',
      options: [{ vars: 'all' }],
      errors: [
        {
          messageId: 'usedOnlyAsType',
          data: { varName: 'foo', action: 'defined', additional: '' },
        },
      ],
    },
    {
      code: 'function foo() {} type Bar = typeof foo;',
      options: [{ vars: 'all' }],
      errors: [
        {
          messageId: 'usedOnlyAsType',
          data: { varName: 'foo', action: 'defined', additional: '' },
        },
      ],
    },

    // Destructuring
    {
      code: 'const [foo] = [1];',
      errors: [
        {
          messageId: 'unusedVar',
          data: { varName: 'foo', action: 'defined', additional: '' },
        },
      ],
    },
    {
      code: 'const { foo } = { foo: 1 };',
      errors: [
        {
          messageId: 'unusedVar',
          data: { varName: 'foo', action: 'defined', additional: '' },
        },
      ],
    },
    {
      code: 'const { foo: bar } = { foo: 1 };',
      errors: [
        {
          messageId: 'unusedVar',
          data: { varName: 'bar', action: 'defined', additional: '' },
        },
      ],
    },

    // Pattern messages
    {
      code: 'const foo = 1;',
      options: [{ varsIgnorePattern: '^_' }],
      errors: [
        {
          messageId: 'unusedVar',
          data: {
            varName: 'foo',
            action: 'defined',
            additional: '. Allowed unused vars must match ^_',
          },
        },
      ],
    },
    {
      code: 'function foo(bar) {}',
      options: [{ argsIgnorePattern: '^_' }],
      errors: [
        {
          messageId: 'unusedVar',
          data: {
            varName: 'bar',
            action: 'defined',
            additional: '. Allowed unused args must match ^_',
          },
        },
      ],
    },
    {
      code: 'try {} catch (e) {}',
      options: [{ caughtErrorsIgnorePattern: '^_' }],
      errors: [
        {
          messageId: 'unusedVar',
          data: {
            varName: 'e',
            action: 'defined',
            additional: '. Allowed unused caught errors must match ^_',
          },
        },
      ],
    },
    {
      code: 'const [foo] = [1];',
      options: [{ destructuredArrayIgnorePattern: '^_' }],
      errors: [
        {
          messageId: 'unusedVar',
          data: {
            varName: 'foo',
            action: 'defined',
            additional:
              '. Allowed unused elements of array destructuring must match ^_',
          },
        },
      ],
    },

    // reportUsedIgnorePattern
    {
      code: 'const _foo = 1; console.log(_foo);',
      options: [{ varsIgnorePattern: '^_', reportUsedIgnorePattern: true }],
      errors: [
        {
          messageId: 'usedIgnoredVar',
          data: {
            varName: '_foo',
            additional: '. Used vars must not match ^_',
          },
        },
      ],
    },
    {
      code: 'function foo(_bar) { console.log(_bar); }',
      options: [{ argsIgnorePattern: '^_', reportUsedIgnorePattern: true }],
      errors: [
        {
          messageId: 'usedIgnoredVar',
          data: {
            varName: '_bar',
            additional: '. Used args must not match ^_',
          },
        },
      ],
    },
    {
      code: 'try {} catch (_e) { console.log(_e); }',
      options: [
        { caughtErrorsIgnorePattern: '^_', reportUsedIgnorePattern: true },
      ],
      errors: [
        {
          messageId: 'usedIgnoredVar',
          data: {
            varName: '_e',
            additional: '. Used caught errors must not match ^_',
          },
        },
      ],
    },
    {
      code: 'const [_foo] = [1]; console.log(_foo);',
      options: [
        { destructuredArrayIgnorePattern: '^_', reportUsedIgnorePattern: true },
      ],
      errors: [
        {
          messageId: 'usedIgnoredVar',
          data: {
            varName: '_foo',
            additional:
              '. Used elements of array destructuring must not match ^_',
          },
        },
      ],
    },
  ],
});

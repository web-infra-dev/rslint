import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-shadow-restricted-names', {
  valid: [
    'function foo(bar){ var baz; }',
    '!function foo(bar){ var baz; }',
    '!function(bar){ var baz; }',
    'try {} catch(e) {}',
    'export default function() {}',
    'try {} catch {}',
    'var undefined;',
    'var undefined; doSomething(undefined);',
    'var undefined; var undefined;',
    'let undefined',
    "import { undefined as undef } from 'foo';",
    {
      code: 'let globalThis;',
      options: [{ reportGlobalThis: false }] as any,
    },
    {
      code: 'class globalThis {}',
      options: [{ reportGlobalThis: false }] as any,
    },
    {
      code: "import { baz as globalThis } from 'foo';",
      options: [{ reportGlobalThis: false }] as any,
    },
    'globalThis.foo',
    'const foo = globalThis',
    'function foo() { return globalThis; }',
    "import { globalThis as foo } from 'bar'",
  ],
  invalid: [
    {
      code: 'function NaN(NaN) { var NaN; !function NaN(NaN) { try {} catch(NaN) {} }; }',
      errors: [
        { messageId: 'shadowingRestrictedName' },
        { messageId: 'shadowingRestrictedName' },
        { messageId: 'shadowingRestrictedName' },
        { messageId: 'shadowingRestrictedName' },
        { messageId: 'shadowingRestrictedName' },
        { messageId: 'shadowingRestrictedName' },
      ],
    },
    {
      code: 'function undefined(undefined) { !function undefined(undefined) { try {} catch(undefined) {} }; }',
      errors: [
        { messageId: 'shadowingRestrictedName' },
        { messageId: 'shadowingRestrictedName' },
        { messageId: 'shadowingRestrictedName' },
        { messageId: 'shadowingRestrictedName' },
        { messageId: 'shadowingRestrictedName' },
      ],
    },
    {
      code: 'function Infinity(Infinity) { var Infinity; !function Infinity(Infinity) { try {} catch(Infinity) {} }; }',
      errors: [
        { messageId: 'shadowingRestrictedName' },
        { messageId: 'shadowingRestrictedName' },
        { messageId: 'shadowingRestrictedName' },
        { messageId: 'shadowingRestrictedName' },
        { messageId: 'shadowingRestrictedName' },
        { messageId: 'shadowingRestrictedName' },
      ],
    },
    {
      code: 'function arguments(arguments) { var arguments; !function arguments(arguments) { try {} catch(arguments) {} }; }',
      errors: [
        { messageId: 'shadowingRestrictedName' },
        { messageId: 'shadowingRestrictedName' },
        { messageId: 'shadowingRestrictedName' },
        { messageId: 'shadowingRestrictedName' },
        { messageId: 'shadowingRestrictedName' },
        { messageId: 'shadowingRestrictedName' },
      ],
    },
    {
      code: 'function eval(eval) { var eval; !function eval(eval) { try {} catch(eval) {} }; }',
      errors: [
        { messageId: 'shadowingRestrictedName' },
        { messageId: 'shadowingRestrictedName' },
        { messageId: 'shadowingRestrictedName' },
        { messageId: 'shadowingRestrictedName' },
        { messageId: 'shadowingRestrictedName' },
        { messageId: 'shadowingRestrictedName' },
      ],
    },
    {
      code: 'var eval = (eval) => { var eval; !function eval(eval) { try {} catch(eval) {} }; }',
      errors: [
        { messageId: 'shadowingRestrictedName' },
        { messageId: 'shadowingRestrictedName' },
        { messageId: 'shadowingRestrictedName' },
        { messageId: 'shadowingRestrictedName' },
        { messageId: 'shadowingRestrictedName' },
        { messageId: 'shadowingRestrictedName' },
      ],
    },
    {
      code: 'var [undefined] = [1]',
      errors: [{ messageId: 'shadowingRestrictedName' }],
    },
    {
      code: 'var {undefined} = obj; var {a: undefined} = obj; var {a: {b: {undefined}}} = obj; var {a, ...undefined} = obj;',
      errors: [
        { messageId: 'shadowingRestrictedName' },
        { messageId: 'shadowingRestrictedName' },
        { messageId: 'shadowingRestrictedName' },
        { messageId: 'shadowingRestrictedName' },
      ],
    },
    {
      code: 'var undefined; undefined = 5;',
      errors: [{ messageId: 'shadowingRestrictedName' }],
    },
    {
      code: 'class undefined {}',
      errors: [{ messageId: 'shadowingRestrictedName' }],
    },
    {
      code: '(class undefined {})',
      errors: [{ messageId: 'shadowingRestrictedName' }],
    },
    {
      code: "import undefined from 'foo';",
      errors: [{ messageId: 'shadowingRestrictedName' }],
    },
    {
      code: "import { undefined } from 'foo';",
      errors: [{ messageId: 'shadowingRestrictedName' }],
    },
    {
      code: "import { baz as undefined } from 'foo';",
      errors: [{ messageId: 'shadowingRestrictedName' }],
    },
    {
      code: "import * as undefined from 'foo';",
      errors: [{ messageId: 'shadowingRestrictedName' }],
    },
    {
      code: 'function globalThis(globalThis) { var globalThis; !function globalThis(globalThis) { try {} catch(globalThis) {} }; }',
      errors: [
        { messageId: 'shadowingRestrictedName' },
        { messageId: 'shadowingRestrictedName' },
        { messageId: 'shadowingRestrictedName' },
        { messageId: 'shadowingRestrictedName' },
        { messageId: 'shadowingRestrictedName' },
        { messageId: 'shadowingRestrictedName' },
      ],
    },
    {
      code: 'const [globalThis] = [1]',
      errors: [{ messageId: 'shadowingRestrictedName' }],
    },
    {
      code: 'var {globalThis} = obj; var {a: globalThis} = obj; var {a: {b: {globalThis}}} = obj; var {a, ...globalThis} = obj;',
      errors: [
        { messageId: 'shadowingRestrictedName' },
        { messageId: 'shadowingRestrictedName' },
        { messageId: 'shadowingRestrictedName' },
        { messageId: 'shadowingRestrictedName' },
      ],
    },
    {
      code: 'let globalThis; globalThis = 5;',
      errors: [{ messageId: 'shadowingRestrictedName' }],
    },
    {
      code: 'class globalThis {}',
      errors: [{ messageId: 'shadowingRestrictedName' }],
    },
    {
      code: '(class globalThis {})',
      errors: [{ messageId: 'shadowingRestrictedName' }],
    },
    {
      code: "import globalThis from 'foo';",
      errors: [{ messageId: 'shadowingRestrictedName' }],
    },
    {
      code: "import { globalThis } from 'foo';",
      errors: [{ messageId: 'shadowingRestrictedName' }],
    },
    {
      code: "import { baz as globalThis } from 'foo';",
      errors: [{ messageId: 'shadowingRestrictedName' }],
    },
    {
      code: "import * as globalThis from 'foo';",
      errors: [{ messageId: 'shadowingRestrictedName' }],
    },
  ],
});

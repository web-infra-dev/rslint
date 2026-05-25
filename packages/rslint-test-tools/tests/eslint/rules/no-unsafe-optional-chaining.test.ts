import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-unsafe-optional-chaining', {
  valid: [
    'obj?.foo;',
    'obj?.foo();',
    '(obj?.foo ?? bar)();',
    '(obj?.foo || bar)();',
    // Spread in object literal is safe
    '({...obj?.foo});',
    // Sequence: chain is NOT last
    '(obj?.foo, bar)();',
    // for-in tolerates undefined
    'for (const x in obj?.foo) {}',
    // Non-destructuring binding defaults
    'const {x = obj?.foo} = obj;',
    'function f({x = obj?.foo}: any) {}',
    // Non-null assertion — developer asserts
    '(obj?.foo)!.bar;',
    '(obj?.foo)!();',
    'new (obj?.foo)!();',
    // && consumed by ||/?? fallback
    '((obj?.foo && bar) ?? fallback)();',
    '((obj?.foo && bar) || fallback).baz;',
  ],
  invalid: [
    // Basic contexts
    {
      code: '(obj?.foo)();',
      errors: [{ messageId: 'unsafeOptionalChain' }],
    },
    {
      code: '(obj?.foo).bar;',
      errors: [{ messageId: 'unsafeOptionalChain' }],
    },
    {
      code: 'new (obj?.foo)();',
      errors: [{ messageId: 'unsafeOptionalChain' }],
    },
    {
      code: '(obj?.foo)`text`;',
      errors: [{ messageId: 'unsafeOptionalChain' }],
    },
    {
      code: '[...obj?.foo];',
      errors: [{ messageId: 'unsafeOptionalChain' }],
    },
    {
      code: 'foo(...obj?.foo);',
      errors: [{ messageId: 'unsafeOptionalChain' }],
    },
    // Destructuring
    {
      code: 'const {foo} = obj?.bar;',
      errors: [{ messageId: 'unsafeOptionalChain' }],
    },
    {
      code: '({foo} = obj?.bar);',
      errors: [{ messageId: 'unsafeOptionalChain' }],
    },
    // Binding element (nested destructuring defaults)
    {
      code: 'const {x: {y} = obj?.foo} = obj;',
      errors: [{ messageId: 'unsafeOptionalChain' }],
    },
    {
      code: 'function f({x: {y} = obj?.foo}: any) {}',
      errors: [{ messageId: 'unsafeOptionalChain' }],
    },
    {
      code: 'const [, [a] = obj?.foo] = arr;',
      errors: [{ messageId: 'unsafeOptionalChain' }],
    },
    // Class extends
    {
      code: 'class Foo extends obj?.bar {}',
      errors: [{ messageId: 'unsafeOptionalChain' }],
    },
    {
      code: 'const C = class extends obj?.bar {};',
      errors: [{ messageId: 'unsafeOptionalChain' }],
    },
    // Ternary: both branches — 2 errors
    {
      code: '(cond ? obj?.foo : obj?.bar)();',
      errors: [
        { messageId: 'unsafeOptionalChain' },
        { messageId: 'unsafeOptionalChain' },
      ],
    },
    // && propagates both sides — 2 errors
    {
      code: '(obj?.foo && obj?.bar)();',
      errors: [
        { messageId: 'unsafeOptionalChain' },
        { messageId: 'unsafeOptionalChain' },
      ],
    },
    // && left only
    {
      code: '(obj?.foo && bar)();',
      errors: [{ messageId: 'unsafeOptionalChain' }],
    },
    // Sequence: chain IS last
    {
      code: '(a, b, obj?.foo)();',
      errors: [{ messageId: 'unsafeOptionalChain' }],
    },
    // Deep parentheses
    {
      code: '((((obj?.foo))))();',
      errors: [{ messageId: 'unsafeOptionalChain' }],
    },
    // Await
    {
      code: 'async function h() { (await obj?.foo)(); }',
      errors: [{ messageId: 'unsafeOptionalChain' }],
    },
    // Complex: in with ternary — 2 errors
    {
      code: '"key" in (cond ? obj?.a : obj?.b);',
      errors: [
        { messageId: 'unsafeOptionalChain' },
        { messageId: 'unsafeOptionalChain' },
      ],
    },
    // Complex: spread with conditional — 2 errors
    {
      code: '[...(cond ? obj?.a : obj?.b)];',
      errors: [
        { messageId: 'unsafeOptionalChain' },
        { messageId: 'unsafeOptionalChain' },
      ],
    },
  ],
});

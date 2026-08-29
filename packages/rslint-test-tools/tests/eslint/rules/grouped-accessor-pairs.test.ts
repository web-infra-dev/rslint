import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

// Layer 1 integration subset from ESLint v10.9.1. The complete upstream suite
// and rslint-specific augmentation live in the Go rule package.
ruleTester.run('grouped-accessor-pairs', {
  valid: [
    '({})',
    '({ get a(){}, set a(value){} })',
    '({ set a(value){}, get a(){} })',
    '({ get [a](){}, set [a](value){} })',
    'class A { get a(){} set a(value){} }',
    '(class { static set a(value){} static get a(){} })',
    'class A { get #a(){} middle(){} set "#a"(value){} }',
    {
      code: '({ get a(){}, set a(value){} })',
      options: ['getBeforeSet'] as any,
    },
    {
      code: 'interface I { get prop(): string; set prop(value: string); }',
      options: ['anyOrder', { enforceForTSTypes: true }] as any,
    },
    'interface I { get prop(): string; between: true; set prop(value: string); }',
  ],
  invalid: [
    {
      code: '({ get a(){}, middle: true, set a(value){} })',
      errors: [{ messageId: 'notGrouped' }],
    },
    {
      code: 'class A { get a(){} middle(){} set a(value){} }',
      errors: [{ messageId: 'notGrouped' }],
    },
    {
      code: 'class A { static get #a(){} middle(){} static set #a(value){} }',
      errors: [{ messageId: 'notGrouped' }],
    },
    {
      code: '({ set a(value){}, get a(){} })',
      options: ['getBeforeSet'] as any,
      errors: [{ messageId: 'invalidOrder' }],
    },
    {
      code: 'class A { static get a(){} static set a(value){} }',
      options: ['setBeforeGet'] as any,
      errors: [{ messageId: 'invalidOrder' }],
    },
    {
      code: 'interface I { get a(): string; between: true; set a(value: string); }',
      options: ['anyOrder', { enforceForTSTypes: true }] as any,
      errors: [{ messageId: 'notGrouped' }],
    },
    {
      code: 'type T = { get a(): string; set a(value: string); };',
      options: ['setBeforeGet', { enforceForTSTypes: true }] as any,
      errors: [{ messageId: 'invalidOrder' }],
    },
  ],
});

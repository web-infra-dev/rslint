import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-this-before-super', {
  valid: [
    'class A { constructor() { this.b = 0; } }',
    'class A extends null { constructor() { } }',
    'class A extends B { constructor() { super(); this.c = 0; } }',
    'class A extends B { constructor() { super(); super.c(); } }',
    'class A extends B { constructor() { super(); this.c(); } }',
    'class A { b() { this.c = 0; } }',
    'class A extends B { c() { this.d = 0; } }',
  ],
  invalid: [
    {
      code: 'class A extends B { constructor() { this.c = 0; } }',
      errors: [{ messageId: 'thisBeforeSuper' }],
    },
    {
      code: 'class A extends B { constructor() { this.c = 0; super(); } }',
      errors: [{ messageId: 'thisBeforeSuper' }],
    },
    {
      code: 'class A extends B { constructor() { super.c(); super(); } }',
      errors: [{ messageId: 'superBeforeSuper' }],
    },
  ],
});

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-fallthrough', {
  valid: [
    'switch(foo) { case 0: a(); break; case 1: b(); }',
    'switch(foo) { case 0: case 1: a(); break; }',
    'switch(foo) { case 0: a(); /* falls through */ case 1: b(); }',
    // "fallsthrough" (no space) matches regex
    'switch(foo) { case 0: a(); /* fallsthrough */ case 1: b(); }',
    // Try/catch where both terminate
    'switch(foo) { case 0: try { break; } catch(e) { break; } case 1: b(); }',
    // Nested switch with outer break
    'switch(foo) { case 0: switch(bar) { case 1: break; } break; case 2: b(); }',
  ],
  invalid: [
    {
      code: 'switch(foo) { case 0: a(); case 1: b(); }',
      errors: [{ messageId: 'case' }],
    },
    {
      code: 'switch(foo) { case 0: a(); default: b(); }',
      errors: [{ messageId: 'default' }],
    },
    // Nested switch: inner break does NOT prevent outer fallthrough
    {
      code: 'switch(foo) { case 0: switch(bar) { case 1: break; } case 2: b(); }',
      errors: [{ messageId: 'case' }],
    },
  ],
});

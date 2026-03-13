import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-ex-assign', {
  valid: [
    'try { } catch (e) { three = 2 + 1; }',
    'try { } catch ({e}) { this.something = 2; }',
    'function foo() { try { } catch (e) { return false; } }',
    'try { } catch (e) { let e = 10; e = 10; }',
    'try { } catch (e) { console.log({e}); }',
    'try { } catch (e) { foo({x: e}); }',
  ],
  invalid: [
    {
      code: 'try { } catch (e) { e = 10; }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'try { } catch (ex) { ex = 10; }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'try { } catch (ex) { [ex] = []; }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'try { } catch (ex) { ({x: ex = 0} = {}); }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'try { } catch ({message}) { message = 10; }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'try { } catch (e) { for (;;) { e = 10; } }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'try { } catch (e) { for (e in obj) {} }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'try { } catch (e) { e += 1; }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'try { } catch (e) { e++; }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'try { } catch (e) { --e; }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'try { } catch (e) { for (e of arr) {} }',
      errors: [{ messageId: 'unexpected' }],
    },
  ],
});

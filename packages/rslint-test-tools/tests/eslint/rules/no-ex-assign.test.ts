import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-ex-assign', {
  valid: [
    'try { } catch (e) { three = 2 + 1; }',
    'try { } catch ({e}) { this.something = 2; }',
    'function foo() { try { } catch (e) { return false; } }',
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
  ],
});
